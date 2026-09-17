# User Journey — CLI Step 8

## Purpose

Define how a Human actually uses Gnomon from initialization through completing real work, and how the CLI determines and recommends valid next actions without storing workflow progress.

This is CLI-level design, not Core semantics. It does not define a pipeline Core doesn't state; it traces the journeys Core's own workflow text already implies, and defines how next-action guidance is computed fresh from repository truth rather than remembered.

---

## Scope

This document covers:

- The new-project, existing-project, and feature/change journeys, derived from Core's actual ordering and gating constraints.
- Which workflows are mandatory-given-a-path, contextual, optional, or repeatable — and the corrected classification of Git Finalization specifically.
- The valid-action / recommended-action distinction, including the two recommendation modes.
- The recovery model: why Agent crashes, CLI restarts, manual edits, Git checkouts, and fresh clones all recover identically.
- The Human/Agent/CLI responsibility boundary as it applies across a full journey, not just a single invocation.

This document does not cover:

- Command names or CLI surface syntax — CLI Step 9.
- Any change to Core workflow semantics, Workflow Contract (Step 1), Approval Runtime (Step 6), Result Protocol (Step 5), or Agent Adapter (Step 4) — all referenced here, none redefined.
- New derived facts beyond the single, small addition this Step requires — see the update to `DERIVED_FACTS.md` (Step 7).

---

## Preconditions Carried Forward (Steps 1–7, unchanged)

- **Workflow Contract v1** (Step 1): each workflow's `specification_reference`/`requires_approved_specification` define its deterministic pre-start gate.
- **Approval Runtime** (Step 6): lifecycle state is always derived, never stored.
- **Result Protocol** (Step 5): a workflow's terminal result is transported once, consumed once, then deleted — never persisted.
- **Derived Facts** (Step 7): every fact this Step relies on is recomputed fresh from current repository and `.gnomon/` truth, never cached.

Nothing in this document changes any of these. This document introduces no new persisted state.

---

## No Single Canonical Pipeline

Core defines only the ordering and gating constraints its own text actually states — nothing more, and nothing less:

- A Specification must exist before Specification Definition can act on it (`specification_reference: required`).
- A Specification must be Approved before Implementation, or SPEC-governed Testing, may act on it (`requires_approved_specification: true`) — a hard gate, no exception.
- Beyond these, Core places no fixed relative order on Knowledge Resolution, Testing, Verification, or Review — each is context-triggered, appearing wherever its own When-to-Use text calls for it, zero or more times, in whatever order the actual work requires.

**Repository truth, not persisted workflow progress, determines what is currently valid.** No `current_step`, workflow cursor, or lifecycle/state machine exists anywhere in this model.

---

## New-Project Journey

```text
gnomon init
     ↓
Initial Knowledge Establishment
     ↓
Bootstrap
     ↓
first governed work (Feature/Change journey)
```

`gnomon init` is deterministic and Agent-free (Step 3). Initial Knowledge Establishment is the recommended next action — not enforced — and is the one workflow that resolves its own material gaps entirely within its own internal loop (Steps 3–9 of its own text), never handing off to Knowledge Resolution; Knowledge Resolution's own When to Use names Bootstrap, Definition, Implementation, Testing, Verification, and Review as the workflows it resumes, deliberately excluding IKE. Once IKE reports `READY_FOR_BOOTSTRAP`, Bootstrap establishes the verified baseline — and may itself route through Knowledge Resolution if a gap surfaces that IKE didn't catch. Once `BOOTSTRAP_COMPLETE`, the project has a runnable baseline and the Human enters the Feature/Change journey.

---

## Existing-Project Journey

Mechanically identical to the new-project journey at the `init` and workflow-invocation level — Step 3 already established that `init` never inspects or branches on codebase content, and this Step confirms that holds at the journey level too. **The distinction between a new and an existing project lives entirely inside Initial Knowledge Establishment's own reasoning**, not in `init` or in any CLI-level branching: IKE's own Step 1 (Inspect) finds prior code, configuration, and documentation carrying project intent and reconciles it against what the Human states. From IKE's `READY_FOR_BOOTSTRAP` onward, the two journeys converge completely into the same path toward Bootstrap and governed work.

---

## Feature/Change Journey

Two entry points converge on the same artifact:

- **Discovery-assisted** — Specification Discovery proposes one candidate; the Human chooses Create/Skip/Tell me more/Stop; Create triggers Draft Creation.
- **Direct** — the Human already knows the title; Draft Creation runs immediately, with no Discovery step required (Step 1's own explicit carve-out: Draft Creation is a deterministic CLI operation, never gated behind Discovery).

From a new Draft Specification (identity + title only):

```text
Draft Specification
     ↓
Specification Definition  (iterates; Knowledge Resolution as needed)
     ↓
READY_FOR_APPROVAL  →  Human Approval (Step 6, a separate, explicit act)
     ↓
Approved  →  Implementation eligible
     ↓
Implementation  (Knowledge Resolution as needed; may partially block on a Draft/Missing dependency)
     ↓
Testing (optional/contextual) · Verification (optional/repeatable) · Review (optional/repeatable)
     ↓
Git Finalization  (contextual, whenever the Human judges the work ready)
```

- **Specification Definition** iterates its own Steps 4–7 re-evaluation loop until `READY_FOR_APPROVAL` or `BLOCKED`. Specification-owned gaps are resolved inline; gaps owned elsewhere route through Knowledge Resolution.
- **Human Approval** is recommended right after `READY_FOR_APPROVAL`, but Core never makes that signal a technical precondition of the Approval operation itself — a Human may approve at their own judgment at any time.
- **Implementation** becomes eligible only once Approved (the hard gate). It may route through Knowledge Resolution, and may complete unaffected work while a dependency-affected portion stays blocked, rerunning only that portion once the dependency resolves.
- **Testing** is optional and contextual — invoked when Implementation couldn't run tests itself, for standalone defect reproduction, coverage-gap closure, or a deeper/independent pass — never a default stage every Implementation run requires.
- **Verification** and **Review** are each optional and freely repeatable, with no limit Core states — naturally run when objective evidence or engineering judgment is needed, and naturally re-run after addressing a finding, though Core states this as sound practice rather than a rule. Review's `Recommended Next Workflow` may route back through Knowledge Resolution.
- **Git Finalization** — see the corrected classification below.

---

## Workflow Characterization

| Workflow | Character |
|---|---|
| Initial Knowledge Establishment | Conditionally needed, depending on whether sufficient project knowledge already exists |
| Bootstrap | Contextual and repeatable — recurs as baseline needs change |
| Specification Discovery | Optional (Draft Creation bypasses it) and repeatable |
| Specification Definition | Iterative and repeatable until `READY_FOR_APPROVAL` |
| Knowledge Resolution | Interrupt- and context-driven; appears wherever a workflow reports a material gap and a Human has decided |
| Implementation | Required only when Approved behavior actually needs to be realized — not a fixed-cadence stage |
| Testing | Optional and repeatable |
| Verification | Optional and repeatable |
| Review | Optional and repeatable |
| Git Finalization | The Core-defined workflow for git handoff — invoked contextually whenever the Human chooses to finalize; **not a mandatory stage every change passes through, and not on any fixed cadence** |
| Draft Creation / Approval / Revocation | Deterministic, CLI-native operations — not Agent workflows, invocable whenever needed |

### Git Finalization, precisely

Git Finalization is **not** a mandatory pipeline stage — Core never states that every change must pass through it, and nothing enforces a fixed cadence. It **is** the Core-defined, authoritative mechanism for git handoff of Gnomon-governed work: the one workflow that carries the Draft-governance safeguard ("no authorization... permits finalizing it anyway"), with no Core-defined alternative path to that same safeguard. It **is** invoked contextually, entirely at the Human's discretion, whenever they judge work ready to hand off. In short: mandatory *if* the Human chooses to hand work off through Gnomon's governed system, never mandatory in occurrence or timing.

---

## Valid vs. Recommended Actions

### Valid action

An action the CLI can mechanically confirm is not blocked by any deterministic pre-start gate — Step 1's Workflow Contract combined with Step 7's Derived Facts, evaluated fresh for a specific candidate (workflow + optional Specification). Narrow, purely mechanical, and correct in the moment it's checked. Initial Knowledge Establishment, Bootstrap, Discovery, Verification, and Review are essentially always valid (`specification_reference: none`); Definition is valid whenever its target Specification exists; Implementation and SPEC-governed Testing are valid only when that target is Approved; Git Finalization is always valid to attempt (its one hard rule is Agent-side, checked after starting, never a pre-start gate).

### Recommended action

A non-authoritative suggestion — always a subset of valid actions, never enforced, never a substitute for Human judgment. Computed in exactly two modes, and the CLI must never blur them:

- **Same-session, high-confidence** — immediately after an Agent run, while that run's own just-consumed result is still in hand (before Step 5 deletes the transient file), the CLI can recommend precisely: "Definition just reported `READY_FOR_APPROVAL` — consider approving," or "Implementation reported `BLOCKED` on a knowledge gap — resolve it, then Knowledge Resolution can apply your decision." This is a transient UX courtesy layered on an already-ephemeral result, never a new persisted fact.
- **Cross-session, coarse** — at any other time (a fresh invocation, a restart, a clone, days later), the CLI has no record of what any past run concluded and must not pretend otherwise. It offers only structurally-derived, non-presumptuous guidance: "This Specification is Draft — you can run Definition on it, or approve it directly if you judge it sufficiently defined." **It never claims to know whether a past Definition run already reached `READY_FOR_APPROVAL`**, because that signal was never asked to be durable and Step 5 deliberately made it ephemeral.

No `current_step`, workflow cursor, persisted progress, or workflow state machine is introduced by either mode.

---

## Recovery Model

Agent crash, CLI restart, manual edits, `git checkout`, and a fresh clone all recover **identically**, by the same mechanism: nothing is ever assumed to be in progress anywhere, so there is nothing to resume from. The next invocation simply re-inspects current repository truth and recomputes every Derived Fact fresh — exactly the discipline every workflow's own "inspect before writing" language already assumes, now confirmed to hold at the whole-journey level, not just within a single invocation.

---

## Human / Agent / CLI Boundaries

- **Human** owns granting and revoking Approval, every material decision any workflow surfaces, and the discretionary choice of whether and when to run any optional workflow (Testing, Verification, Review, Git Finalization).
- **Agent** owns semantic reasoning and the execution of every workflow — including all engineering judgment about readiness, correctness, materiality, and classification.
- **CLI** owns deterministic validation, Derived Facts, pre-start gating, and next-action guidance — computing what is *valid* and offering what is *recommended*, never deciding what is *correct*.

---

## Deferred to Future Work

- Command names and CLI surface syntax — CLI Step 9.
- Any richer, persisted notion of journey progress — deliberately rejected here, not merely postponed; nothing in this analysis found it necessary, and Steps 5–7's ephemeral/derived-state discipline argues directly against it.

---

## Notes

This document persists the Step 8 architecture finalized through prior design analysis and its subsequent correction — specifically, Git Finalization's classification was revised from an earlier, unjustified "mandatory" characterization to the precise, Core-grounded description recorded above. It does not reopen Core or Steps 1–7.
