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
- **`timestamp`** — UTC RFC 3339 with fractional seconds (nanosecond precision). This is what determines grant order — oldest to newest — and therefore which grant is the most recent decision (**L**) the Derived Lifecycle Rule below picks: a new grant's timestamp is kept strictly later than every existing grant for that Specification, advanced by one nanosecond over the latest existing one when the clock alone would not guarantee that (coarse clock resolution, notably on Windows). Legacy records written at second precision remain valid; two of them landing in the same second sort deterministically (by grant id) but not necessarily chronologically.

## Revocation Record

```json
{
  "revokes": "a1b2c3",
  "revoked_by": "Name <email>",
  "timestamp": "…"
}
```

`revokes` references the specific grant's own filename stem — the same identifier that already uniquely names that grant, requiring no separate cross-reference scheme. A revocation record is **never** a modification of the grant it revokes; the grant file is never touched.

## Evidence Validity

Every grant and revocation record is checked against the same rules, by the same one function per record type, whether the check is running to derive Draft/Approved or to report `gnomon validate`'s own findings — a record is never treated as valid by one and invalid by the other.

**Error** — the record is invalid and excluded from derivation entirely:
- malformed JSON
- a grant missing/empty `spec_identity`, `fingerprint`, or `approver`; or a `spec_identity` that does not match the directory it is stored in
- a revocation missing/empty `revokes` or `revoked_by`
- `timestamp` missing, or not parseable as RFC 3339 (with or without fractional seconds)
- a file that is neither `*.grant.json`, `*.revocation.json`, nor `*.content.md`

**Warning** — reported, but never affects derivation:
- a revocation referencing a grant id that does not exist in the same directory
- a content snapshot with no matching grant (can happen if a grant write failed after its snapshot was written)
- more than one revocation for the same grant
- an evidence directory for a Specification identity that has no Specification file

A grant with no content snapshot is not reported at all — records written before snapshots existed are legitimate.

`gnomon validate` exits non-zero when any error exists anywhere in the project, including approval evidence; warnings alone still exit zero but are always printed, grouped separately and after any errors.

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

**The latest approval decision counts.** Every valid grant record belonging to a Specification's identity is ordered by when it was granted (its timestamp, then its own record id as a tie-break for two grants written in the same instant — the identical ordering the CLI's own revision history already displays). The most recent one, call it **L**, is the only grant that can currently make the Specification Approved — regardless of whether some older grant happens to match current content.

A Specification is **Approved** if and only if:

- L exists;
- L has not been revoked by any revocation record referencing it;
- L matches the current normalized-content fingerprint.

Otherwise, the Specification derives **Draft** — including when L simply no longer matches current content, and including when L has been revoked, however many older grants might still happen to match. Revoking any grant other than L never changes this answer; only a new grant (which becomes the new L) or revoking L itself does.

**No mutable `state: approved` field, or any equivalent cached value, is ever persisted anywhere** — not on the Specification, not in a separate file, not anywhere. Draft and Approved are never written down; they are the return value of this rule, recomputed fresh every time it is asked, from whatever content and evidence files currently, actually exist.

---

## Human Identity

The default attribution source is the repository's already-configured Git identity — `user.name` and `user.email`, formatted as `"Name <email>"`, matching Git's own commit-author convention. This is **attribution, not authentication**: nothing verifies the claim is truthful, exactly matching Core's own bar ("attributable to an explicit Human action," never cryptographic proof). No new identity, account, or credential system is introduced.

If Git identity cannot be resolved, the CLI prompts the Human for a one-time identity at the moment of the approval or revocation action. This fallback value is used only for that single record — **it is never persisted as new Gnomon configuration**. If no identity can be obtained either way, the CLI **refuses the action and writes nothing** — a record with no resolvable approver is not valid evidence under Core's own attribution requirement, and it is better to refuse at write time than to persist a record already known to be invalid.

---

## Approval Operation

1. Resolve the target Specification's identity.
2. Compute the fingerprint of its current content.
3. If L (per the Derived Lifecycle Rule) already matches this exact fingerprint and has not been revoked, the Specification is already Approved for this content: write nothing further and report that state. Content matching only an older, non-latest grant is not exempted here — it falls through to step 7 below.
4. If the Specification's template is readable, check its content for remaining template placeholders (whether Define was ever reached is not persisted, so this is the only available signal). If any remain, warn the Human, listing up to 5 and suggesting Define; with a real interactive terminal, ask for confirmation before continuing (default No — declining writes nothing and reports cancellation). Without one, print the warning and proceed — approval is never blocked by this check. A missing or unreadable template skips it silently.
5. Confirm this is an explicit Human action in the current session — never inferred from an Agent's reported result.
6. Resolve Human attribution (above); refuse if none can be obtained.
7. Write one new grant record.
8. Re-derive lifecycle state (now Approved, by construction — the new grant is now L) and report it.

Re-approval after any edit, semantic or otherwise, always creates a **new** grant record. Re-approval of content that already matches L creates nothing new (step 3) — this is what keeps step 7 from ever producing a second active grant for the same content. Content matching only an older, non-latest grant is *not* a no-op: approving it writes a fresh grant, which becomes the new L — this is how a Human returns to an earlier version. Prior grant and revocation records are never modified or deleted.

## Revocation Operation

1. Confirm L (per the Derived Lifecycle Rule) currently makes the Specification Approved; refuse otherwise — there is nothing active to revoke, however many older grants might still happen to match current content.
2. Locate every unrevoked grant record matching the current fingerprint — L itself, plus any legacy duplicate grant for the same content (evidence written before the Approval Operation's step 3 existed to prevent duplicates). Revoking every one keeps evidence tidy without changing which grant is L.
3. Confirm explicit Human action.
4. Resolve Human attribution once; refuse if none can be obtained.
5. Write one new revocation record for each grant located in step 2 — each revocation record still references exactly one grant.
6. Re-derive lifecycle state (now Draft, by construction — the derivation rule's own conditions, not a separate write) and report it.

---

## Interaction with Workflow Contract Gates

A workflow declaring `requires_approved_specification: true` calls one small, generic "derive this Specification's current lifecycle state" function and receives back Draft or Approved. **The gating code never needs to know evidence is stored as one-file-per-event under `.gnomon/approvals/`, never needs the fingerprint algorithm, and never needs to know revocation records exist** — the same clean boundary discipline already established for Result Contract validation in Step 5.

## Interaction with Agent Outcomes / Result Protocol

An Agent's reported `READY_FOR_APPROVAL` — a workflow result, transported and structurally validated exactly per Step 5 — **never constitutes, implies, or triggers approval**, under any circumstance. Approval is created only by the explicit operation above, only on explicit Human action, entirely independent of what any Agent run reported. The CLI may use a `READY_FOR_APPROVAL` result to recommend that the Human consider approving next; the actual grant remains a wholly separate act.

## Protection During Agent Runs

An Agent process has ordinary filesystem write access for the duration of a run and could otherwise write, edit, or delete files under `.gnomon/approvals/` directly, bypassing the Approval/Revocation Operations above entirely. The CLI treats this exactly like any other attempt to grant or revoke approval outside those two explicit Human operations: not permitted.

Before starting the Agent, the CLI snapshots every file under `.gnomon/approvals/`; after the run ends — regardless of whether it succeeded, failed, or was cancelled — it compares the directory again. Any difference restores the directory to its pre-run state and rejects the run outright, reporting it Failed, even if the Agent's own reported result was otherwise valid: a run is never accepted on the strength of evidence the Agent itself supplied.

While a run is active, `gnomon approve`/`gnomon revoke` refuse — a Human approving or revoking in another terminal during the run would otherwise be undone by the restore step above.

### The Governing Specification Must Not Change (Rule A)

An Agent could otherwise achieve the same effect indirectly: not by touching `.gnomon/approvals/` at all, but by editing the governing Specification's own content mid-run to match whatever was actually built, so the Human's original approval is quietly reattached to different requirements. A workflow whose Contract declares `requires_approved_specification: true`, run with a Specification actually supplied, is protected against exactly this:

Before the Agent starts, the CLI records that Specification's file path, exact bytes, and fingerprint. After the run ends — on every exit path, exactly like the approvals check above — it resolves the same identity again. A whitespace-only edit (same fingerprint) is not a violation and is left untouched; the Specification no longer resolving, resolving to a different file (a rename), or resolving to a different fingerprint all are. On a violation, the Agent's own version (and any other file now claiming the same identity) is preserved under the project's transient run directory for Human inspection, the original file is restored at its original path, and the run is rejected — Failed, never accepted, even if the Agent's reported result was otherwise valid. Code changes elsewhere in the repository are **not** reverted; the report says so explicitly, since the working tree may still contain work done against a requirement that briefly read differently. If both this and the approvals check above fire in the same run, both are reported together and both are restored.

This does not apply to a workflow whose Contract does not require an Approved Specification (Specification Definition may freely edit its own, still-Draft, target) or to Knowledge Resolution (`specification_reference: none`), which exists specifically to apply a Human decision to authoritative knowledge, Specifications included.

### Other Specifications Losing Approval Is Reported, Never Restored (Rule B)

Every run, regardless of workflow, is also checked for a side effect Rule A does not cover: some *other* Specification, not the one this run governs, derived Approved beforehand and no longer does afterward (its content changed, or it was removed). This is reported as a warning on the run's own result — never restored, never a reason to reject — so that Knowledge Resolution's legitimate work (updating an Approved Specification as part of applying a Human decision) keeps working exactly as intended, while the loss of approval is never silently invisible. A Specification Rule A already governs and reported in full is never also listed here.

Because lifecycle state is never cached, a Human editing a Specification outside any Gnomon workflow requires no invalidation step of any kind: the next time anything needs this Specification's state, the CLI recomputes the fingerprint of whatever content currently exists on disk and checks it against L, exactly as it always does. A representation-only edit leaves the fingerprint unchanged and Approved survives automatically; any other edit produces a fresh fingerprint that no longer matches L — an older grant might still coincidentally match, but only L decides — and Draft is derived, with nothing to invalidate, because nothing was ever cached.

## Git and Branch Behavior

Evidence is committed like any other durable Gnomon artifact. Because every event is its own new file, two branches independently approving, revoking, or leaving a Specification's approval history untouched merge without conflict at the file level in the ordinary case. After a merge, lifecycle state is recomputed fresh against whatever content and evidence files the merge actually produced — never carried forward from either branch — so a branch that approved old content and a branch that changed the Specification correctly derive Draft once merged, with no special merge-conflict handling required for the lifecycle state itself.

When both branches approved — each its own, different content — the merge leaves both grants in evidence, and the Derived Lifecycle Rule picks whichever is L by timestamp, regardless of which branch it came from. The Specification may therefore need to be approved again after the merge, even if the final merged content happens to match one of the two branches' own already-approved version.

---

## Failure Behavior

| Case | Resolution |
|---|---|
| No evidence exists for this Specification | Draft — the ordinary case for any never-approved Specification |
| Fingerprint mismatch against L | Draft — the ordinary case for changed content |
| L has been revoked | Draft, even if content still matches L exactly |
| Content matches an older, non-latest grant, while a newer L exists that doesn't match | Draft — the older grant never becomes active again; approve again to make it L |
| Malformed evidence file | Excluded from consideration; diagnostic warning surfaced; derivation proceeds using whatever other records remain readable |
| Grant record missing attribution | Treated as invalid; excluded from derivation |
| Specification deleted | Its evidence remains (append-only, never deleted), inert, surfaced only if explicitly inspected |
| Multiple valid historical grants for one Specification | Normal and expected — one per historical approval cycle |
| More than one unrevoked grant matches the same content (legacy duplicates) | Approved only if one of them is L; Revocation still clears every matching grant, not just L, to keep evidence tidy |
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
