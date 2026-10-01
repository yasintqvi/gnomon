# Specification Dependencies

## Purpose

Define what it means for one Specification to depend on another, where that relationship is recorded, and how it interacts with the Specification lifecycle — as the shared semantic contract every Specification and workflow must follow.

This document defines semantics only. It does not implement, compute, persist, or enforce anything itself; it defines what any future workflow, gating rule, or tooling must satisfy.

---

## Scope

This document covers:

- What qualifies as a real dependency between Specifications, as distinct from ordering convenience.
- Where a dependency is recorded and who may establish or remove one.
- How a dependency interacts with `SPECIFICATION_LIFECYCLE.md`'s Draft/Approved model.
- The smallest mechanically derivable facts about a dependency's target.
- The rules against dependency cycles and authored transitive dependencies.

This document does not cover:

- Whether any workflow must treat a dependency as blocking, or what condition (existence, Draft, Approved) it requires before proceeding — a Step 6 / workflow-execution concern.
- How dependency facts are persisted, cached, or recomputed at runtime.
- How staleness — whether a Specification is still substantively correct given a changed dependency — is detected or compared; that determination belongs to Review and Knowledge Resolution.
- Cycle-detection algorithms, CLI rendering, or any Agent-adapter behavior.
- Specification Discovery's ordering reasoning, which never becomes a dependency by itself (see Boundaries).

---

## What a Dependency Means

`SPEC-B depends on SPEC-A` means B's own approval-relevant content — its Preconditions, Business Rules, Main Flow, Alternative Flows, or Acceptance Criteria — relies on a capability, rule, guarantee, or outcome that A owns.

The governing test: if A did not exist, or its behavior were materially different, would B become incomplete, incorrect, or unverifiable?

- If yes, it is a real dependency.
- If no — A is merely convenient or natural to build first — it is not a formal dependency. It may remain informal ordering reasoning, exactly as Specification Discovery already produces, and must never be recorded as a dependency on that basis alone.

Shared Domain terminology alone never creates a dependency. Two Specifications both referencing the same concept in `DOMAIN.md` are each correctly using Domain knowledge that already belongs to `DOMAIN.md` — this is Domain ownership working as intended, not a relationship between the two Specifications.

---

## Ownership

A dependency is authored, approval-relevant semantic content of the **dependent** Specification only — the one whose behavior relies on another. It is never recorded in the target Specification, and never stored anywhere else.

- Do not maintain a central dependency registry separate from the Specifications themselves.
- Do not record a reciprocal or back-reference declaration in the target ("depended on by B") — "what depends on A" is always a derived query over every other Specification's own declarations, never stored content.
- Do not duplicate a dependency in more than one place. A Specification's `Dependencies` section is the sole authoritative record of its own dependencies.

Because a dependency is ordinary authored content, establishing or removing one requires no new authority model: an Agent may author or propose it — most naturally while defining the Specification — but it becomes authoritative only once a Human approves the Specification's content, through the existing mechanism `SPECIFICATION_LIFECYCLE.md` already defines. Nothing here creates a separate approval concept for dependencies.

---

## Lifecycle Interaction

Adding, removing, or changing a dependency declaration is a semantic edit to the dependent Specification (B) and is governed entirely by `SPECIFICATION_LIFECYCLE.md`'s existing rules — no new rule is introduced. If B was Approved, changing its `Dependencies` section invalidates that approval exactly as changing any other section would, regardless of whether the edit was made by an Agent or manually by a Human outside any Gnomon workflow. No separate manual-edit mechanism exists or is needed for this section.

The reverse is not true. If the dependency's target (A):

- becomes Draft;
- has its approval-relevant content changed;
- is deleted;
- becomes otherwise missing;

**B's lifecycle state must not change merely because of that external event.** B remains Approved for as long as its own approval-relevant content — including its own unchanged `Dependencies` section — still matches valid, non-revoked approval evidence. This follows directly from `SPECIFICATION_LIFECYCLE.md:100`'s existing rule that external knowledge changes never silently change a Specification's lifecycle state.

Three distinct things must never be conflated:

1. **B's own approval validity** — governed purely by B's own content match. Unaffected by A's state.
2. **Dependency satisfaction** — the derived facts below. Changes freely, independent of B's content or approval.
3. **Consistency or staleness** — whether B is still substantively *correct* given A's current state. A judgment call for Review's `KNOWLEDGE GAP` classification or Knowledge Resolution's consistency check to surface as a finding — never an automatic lifecycle transition.

---

## Derived Dependency Facts

For a given dependency, exactly two facts are mechanically derivable from current repository state, computed fresh each time and never stored:

- **Target Presence**: `PRESENT | MISSING` — whether a Specification with the referenced identity currently exists.
- **Target Lifecycle State**: `DRAFT | APPROVED`, when present — the target's current lifecycle state, read directly from `SPECIFICATION_LIFECYCLE.md`'s existing derivation; this document does not redefine or duplicate that derivation.

No universal `SATISFIED` / `UNSATISFIED` verdict is defined here. Whether a given Target Presence and Target Lifecycle State combination is acceptable is workflow-relative — Specification Definition, Human Approval, Implementation, Testing, Verification, and Review may each reasonably require different conditions from the same two facts. Defining or enforcing any such threshold is a Step 6 concern.

---

## Cycles and Transitivity

A dependency cycle (for example `SPEC-A depends on SPEC-B`, and `SPEC-B depends on SPEC-A`, directly or through intermediate Specifications) is invalid. A dependency means the target must be meaningfully established for the dependent to be meaningful; a cycle would require both to be established before either is meaningful, which has no valid resolution order. A dependency declaration must not create a cycle with the dependencies already declared elsewhere.

Only direct dependencies are authored. If `SPEC-A` depends on `SPEC-B` and `SPEC-B` depends on `SPEC-C`, `SPEC-A` does not also author a dependency on `SPEC-C`. Transitive relationships are always derived by following direct declarations at query time; they are never separately authored or duplicated into any Specification's content.

This document defines the no-cycle and direct-only rules as semantic requirements. It does not specify a cycle-detection algorithm or where such a check runs at runtime — that is later implementation work.

---

## Boundaries — What Does Not Belong to This Model

- **Specification Discovery's ordering reasoning never becomes a dependency automatically.** Discovery may mention another Specification by name in its ephemeral, prose rationale for why a candidate seems useful now — this is explicitly permitted by [`workflows/specification-discovery.md`](../workflows/specification-discovery.md) and remains informal. Accepting a Discovery proposal and creating a Draft never creates a dependency; Draft Creation writes only identity and title. A dependency becomes formal only when it is written into the dependent Specification's own `Dependencies` section, during Specification Definition or direct manual authoring.
- **Workflow gating is not defined here.** Whether Implementation, Testing, Verification, or Review must stop, warn, or proceed because of a dependency's current facts is a workflow-execution concern for Step 6, not a semantic question this document answers.
- **Staleness detection and comparison mechanisms are not defined here.** Determining that a target's content has materially changed since a dependency was declared — and deciding whether that matters — is deferred, consistent with how `SPECIFICATION_LIFECYCLE.md` defers the analogous content-comparison mechanism for approval evidence.

---

## Deferred to Future Work

The following are intentionally not defined here. They are workflow-execution, tooling, or later-step concerns that must satisfy the contract above, not redefine it:

- Whether any workflow requires a dependency's target to be Approved, Draft-or-better, or merely Present before proceeding — specified in each workflow's own Rules and Failure Handling (`implementation.md`, `testing.md`, `verification.md`, `review.md`), not here.
- How dependency facts are computed, cached, or recomputed at runtime.
- How staleness is detected or compared.
- Cycle-detection algorithm implementation.
- CLI rendering of dependency facts.
- Any Agent-adapter or orchestration behavior.

---

## Notes

[Optional clarification that does not belong to another section.]
