# Approval Runtime — CLI Step 6

## Purpose

Define the smallest runtime/persistence mechanism that lets Gnomon CLI deterministically derive whether a Specification's current semantic content is Approved, per `SPECIFICATION_LIFECYCLE.md`.

This is CLI-level design, not Core semantics. Core already fixes what Draft/Approved mean, who may grant or revoke approval, and the properties approval evidence must have; this document only fills in what Core explicitly leaves deferred — how that evidence is captured, compared, and stored.

---

## Scope

This document covers:

- The physical, durable, append-only representation of approval evidence under `.gnomon/`.
- The exact content-fingerprint mechanism used to detect whether a Specification's approval-relevant content has changed.
- The derived Draft/Approved rule, expressed operationally.
- The minimal Human-attribution mechanism.
- The conceptual approval and revocation operations, and how Workflow Contract gates and the Result Protocol consume this runtime.
- Failure behavior, and the boundary of what remains inside Core's own, unmodified lifecycle semantics.

This document does not cover:

- Any redefinition of Draft, Approved, or approval-relevant content — those remain defined solely by `SPECIFICATION_LIFECYCLE.md`.
- Workflow Contract v1's eligibility fields or Result Contract — CLI Step 1 (`WORKFLOW_CONTRACT.md`).
- The Bundled Default Core / Project-local Core distribution model — CLI Step 2 (`CORE_DISTRIBUTION.md`).
- `.gnomon/` materialization and `gnomon init` behavior beyond the single, minimal layout addition this Step requires — CLI Step 3 (`PROJECT_INITIALIZATION.md`).
- Agent startup and Terminal Handoff — CLI Step 4 (`AGENT_ADAPTER.md`).
- The workflow result envelope and payload validation — CLI Step 5 (`RESULT_PROTOCOL.md`).
- Cryptographic signatures, authentication, accounts, or any identity-infrastructure beyond reusing what Git already provides.

---

## Preconditions Carried Forward (Core and Steps 1–5, unchanged)

- **Core lifecycle semantics** (`SPECIFICATION_LIFECYCLE.md`): exactly two states, always derived, never stored; approval binds to the exact approval-relevant content reviewed at grant time; only a Human may grant or revoke; evidence must be content-bound, separate from the Specification's own content, attributable, durable, and immutable/append-only; a revoked record can never revalidate the Specification, even on an exact later content match, until a new valid approval is granted.
- **Workflow Contract v1** (Step 1): workflows declare `requires_approved_specification`; the CLI checks this against a two-valued lifecycle answer before starting the Agent.
- **Core Distribution / Project Initialization** (Steps 2–3): `.gnomon/` is the sole runtime authority, materialized and committed by `gnomon init`.
- **Result Protocol** (Step 5): an Agent's reported workflow result is never treated as, or conflated with, a Human decision.

Nothing in this document changes any of these.

---

## Physical Evidence Layout

Approval evidence lives under a dedicated, durable, committed subtree of `.gnomon/`:

```text
.gnomon/
  approvals/
    SPEC-003/
      a1b2c3.grant.json
      d4e5f6.revocation.json
      g7h8i9.grant.json
```

Each grant or revocation is **one distinct, immutable file** — never a single mutable file rewritten in place, and never a shared append-only log. This is chosen deliberately over a per-Specification append-only log file: two branches each adding a different new file never conflict at the git level, by construction, whereas two branches each appending a line to the *same* file relies on git's merge heuristic correctly reconciling non-overlapping insertions — usually successful, but not an unconditional guarantee. One-file-per-event also gives the smallest possible corruption blast radius: a damaged file affects only itself, never any other grant or revocation for the same Specification.

The parent `approvals/` directory is created empty by `gnomon init`, alongside `decisions/` and `contracts/`. Each per-Specification subdirectory is created lazily, on that Specification's first approval — never speculatively for Specifications that have never been approved.

Filenames use a collision-resistant identifier generated independently per event — **never a shared counter or sequence file**, which would itself be a piece of shared mutable state two branches could contend over. Two independently-generated identifiers colliding is not a realistic concern at this scale.

---

## Grant Record

```json
{
  "spec_identity": "SPEC-003",
  "fingerprint": "…",
  "approver": "Name <email>",
  "timestamp": "…"
}
```

- **`spec_identity`** — the Specification's stable identity (Discovery's identity rule), independent of filename.
- **`fingerprint`** — the normalized-content fingerprint (below) of the exact content reviewed at grant time.
- **`approver`** — the resolved Human attribution (below).
- **`timestamp`** — informational/audit context; no derivation rule below depends on it, but Core's own "durable" property implies evidence should remain meaningfully inspectable later, and this costs nothing to include.

## Revocation Record

```json
{
  "revokes": "a1b2c3",
  "revoked_by": "Name <email>",
  "timestamp": "…"
}
```

`revokes` references the specific grant's own filename stem — the same identifier that already uniquely names that grant, requiring no separate cross-reference scheme. A revocation record is **never** a modification of the grant it revokes; the grant file is never touched.

---

## Fingerprint Model

The fingerprint is computed by:

1. Normalizing line endings to one canonical form.
2. Trimming trailing whitespace from each line.
3. Collapsing runs of blank lines to a single canonical blank-line representation.
4. Hashing the resulting normalized text.

No Markdown semantic AST, no section-by-section extraction, and no semantic-diff engine are built. **Any content change not clearly covered by this representation-only normalization is conservatively treated as approval-relevant** — this directly follows `SPECIFICATION_LIFECYCLE.md`'s own explicit instruction to default ambiguous cases toward approval-relevant rather than silently preserving Approved status. Core's own text names only whitespace/formatting as unambiguously non-semantic; it does not give a complete boundary for every borderline case (for example, a reworded sentence or a typo fix in explanatory prose) — this is a real, acknowledged gap in Core, and this mechanism resolves it the one way Core's own text sanctions: erring toward requiring re-approval rather than risking a false Approved. The exact hash algorithm is an implementation detail, not fixed here.

---

## Derived Lifecycle Rule

A Specification is **Approved** if and only if at least one grant record:

- belongs to that Specification's identity;
- matches the current normalized-content fingerprint;
- has valid Human attribution;
- has not been revoked by any revocation record referencing it.

Otherwise, the Specification derives **Draft**.

**No mutable `state: approved` field, or any equivalent cached value, is ever persisted anywhere** — not on the Specification, not in a separate file, not anywhere. Draft and Approved are never written down; they are the return value of this rule, recomputed fresh every time it is asked, from whatever content and evidence files currently, actually exist.

---

## Human Identity

The default attribution source is the repository's already-configured Git identity — `user.name` and `user.email`, formatted as `"Name <email>"`, matching Git's own commit-author convention. This is **attribution, not authentication**: nothing verifies the claim is truthful, exactly matching Core's own bar ("attributable to an explicit Human action," never cryptographic proof). No new identity, account, or credential system is introduced.

If Git identity cannot be resolved, the CLI prompts the Human for a one-time identity at the moment of the approval or revocation action. This fallback value is used only for that single record — **it is never persisted as new Gnomon configuration**. If no identity can be obtained either way, the CLI **refuses the action and writes nothing** — a record with no resolvable approver is not valid evidence under Core's own attribution requirement, and it is better to refuse at write time than to persist a record already known to be invalid.

---

## Approval Operation

1. Resolve the target Specification's identity.
2. Compute the fingerprint of its current content.
3. Confirm this is an explicit Human action in the current session — never inferred from an Agent's reported result.
4. Resolve Human attribution (above); refuse if none can be obtained.
5. Write one new grant record.
6. Re-derive lifecycle state (now Approved, by construction) and report it.

Re-approval — after any edit, semantic or otherwise — always creates a **new** grant record. Prior grant and revocation records are never modified or deleted.

## Revocation Operation

1. Locate the grant record currently causing the Specification to derive Approved, if any.
2. Confirm explicit Human action.
3. Resolve Human attribution; refuse if none can be obtained.
4. Write one new revocation record referencing that specific grant.
5. Re-derive lifecycle state (now Draft, by construction — the derivation rule's own conditions, not a separate write) and report it.

---

## Interaction with Workflow Contract Gates

A workflow declaring `requires_approved_specification: true` calls one small, generic "derive this Specification's current lifecycle state" function and receives back Draft or Approved. **The gating code never needs to know evidence is stored as one-file-per-event under `.gnomon/approvals/`, never needs the fingerprint algorithm, and never needs to know revocation records exist** — the same clean boundary discipline already established for Result Contract validation in Step 5.

## Interaction with Agent Outcomes / Result Protocol

An Agent's reported `READY_FOR_APPROVAL` — a workflow result, transported and structurally validated exactly per Step 5 — **never constitutes, implies, or triggers approval**, under any circumstance. Approval is created only by the explicit operation above, only on explicit Human action, entirely independent of what any Agent run reported. The CLI may use a `READY_FOR_APPROVAL` result to recommend that the Human consider approving next; the actual grant remains a wholly separate act.

## Manual Specification Editing

Because lifecycle state is never cached, a Human editing a Specification outside any Gnomon workflow requires no invalidation step of any kind: the next time anything needs this Specification's state, the CLI recomputes the fingerprint of whatever content currently exists on disk and checks it against existing evidence, exactly as it always does. A representation-only edit leaves the fingerprint unchanged and Approved survives automatically; any other edit produces a fresh fingerprint that no existing grant matches, and Draft is derived — with nothing to invalidate, because nothing was ever cached.

## Git and Branch Behavior

Evidence is committed like any other durable Gnomon artifact. Because every event is its own new file, two branches independently approving, revoking, or leaving a Specification's approval history untouched merge without conflict at the file level in the ordinary case. After a merge, lifecycle state is recomputed fresh against whatever content and evidence files the merge actually produced — never carried forward from either branch — so a branch that approved old content and a branch that changed the Specification correctly derive Draft once merged, with no special merge-conflict handling required for the lifecycle state itself.

---

## Failure Behavior

| Case | Resolution |
|---|---|
| No evidence exists for this Specification | Draft — the ordinary case for any never-approved Specification |
| Fingerprint mismatch | Draft — the ordinary case for changed content |
| The only matching grant has been revoked | Draft |
| Malformed evidence file | Excluded from consideration; diagnostic warning surfaced; derivation proceeds using whatever other records remain readable |
| Grant record missing attribution | Treated as invalid; excluded from derivation |
| Specification deleted | Its evidence remains (append-only, never deleted), inert, surfaced only if explicitly inspected |
| Multiple valid historical grants for one Specification | Normal and expected — one per historical approval cycle |
| Manual evidence fabrication or editing | Treated exactly as manually editing a Specification's own content — inside the same existing repository trust boundary Gnomon already relies on everywhere else; no tamper-detection mechanism is introduced |

**No case introduces a third lifecycle state.** Every case resolves to Draft, Approved, or a diagnostic warning accompanying one of those two.

---

## Trust Model

Approval evidence is stored in a repository the Human already has full write access to. Manually fabricating or editing an evidence record is treated identically to manually editing a Specification's own content to claim something false — Core's trust model already assumes a repository owner will not maliciously lie to themselves, and Step 4's own security-boundary conclusion (no second sandbox) applies identically here. Git's own commit history over the evidence files provides a free, already-present audit trail for the ordinary, good-faith case. Cryptographic signing is deliberately not adopted for v1 — disproportionate to any requirement actually evidenced — though the one-file-per-event shape leaves room for an additive, optional signature field later without requiring any redesign.

---

## Deferred to Future Work

- Cryptographic or signed approval evidence.
- Any GUI-specific presentation of approval history.
- A finer, more precise operational definition of "representation-only" beyond whitespace/formatting — the gap named under Fingerprint Model above, deliberately left as Core's own conservative default rather than resolved unilaterally here.
- Exact record-identifier generation scheme and exact hash algorithm — implementation details, not fixed by this document.

---

## Notes

This document persists the Step 6 architecture finalized through analysis and challenge in prior design discussion. It does not reopen Core's lifecycle semantics or Steps 1–5. The only change made elsewhere as a consequence of this Step is the smallest possible addition to `PROJECT_INITIALIZATION.md`'s canonical layout — an empty `approvals/` area created by `gnomon init` — never a change to init's own semantics, and never the physical creation of any approval evidence at init time.
