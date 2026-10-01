---
identity: specification-definition
specification_reference: required
requires_approved_specification: false
result:
  terminal_path: outcome
  classification:
    READY_FOR_APPROVAL: success
    BLOCKED: blocked
  schema:
    type: object
    required:
      - outcome
    properties:
      outcome:
        type: string
        enum:
          - READY_FOR_APPROVAL
          - BLOCKED
      target:
        type:
          - string
          - "null"
      remaining_unresolved:
        type:
          - string
          - "null"
      non_blocking_deferred_items:
        type:
          - string
          - "null"
      consistency_check:
        type:
          - string
          - "null"
      resolved_this_run:
        type:
          - string
          - "null"
---

# Specification Definition Workflow

## Purpose

Bring an existing Draft Specification from its current state — empty, partially written, manually authored, or previously worked on by an Agent — to the point where Gnomon can report `READY_FOR_APPROVAL`, without inventing material decisions and without approving it.

Specification Definition does not grant or simulate Human approval, does not discover Specifications or determine dependencies between them, and does not perform Bootstrap, Implementation, Testing, Verification, or Review.

## When to Use

Use this workflow to define, complete, or resolve gaps in a Draft Specification before it can be approved.

Do not use it to grant, revoke, or simulate approval — that is governed solely by [`SPECIFICATION_LIFECYCLE.md`](../specifications/SPECIFICATION_LIFECYCLE.md). Do not use it to discover which Specifications exist or how they depend on one another — that is Specification Discovery. If defining this Specification reveals that a different, already-Approved Specification may need to change, report that as a finding; do not modify the other Specification here.

## Inputs

- The Draft Specification's current content, exactly as it exists
- Relevant Domain knowledge
- Applicable `ARCHITECTURE.md`, `STACK.md`, and `CONVENTIONS.md`
- Applicable Design Knowledge for user-facing use cases: `PRODUCT_EXPERIENCE.md`, `UI_FOUNDATION.md`, `INTERACTION_PATTERNS.md`
- Applicable ADRs and Artifact Contracts — including to recognize behavior a Contract or Convention already owns, so it need not be restated in the Specification
- [`evaluations/SPECIFICATION_READINESS_CRITERIA.md`](../evaluations/SPECIFICATION_READINESS_CRITERIA.md)

## Execution

### 1. Inspect

Read the Specification exactly as it currently exists. Distinguish authored content from unfilled template placeholders. Do not assume the Specification is empty, and do not assume existing content — manually authored or previously produced by an Agent — is wrong; treat it as potentially valid until shown otherwise.

### 2. Determine Scope

Establish the use case's purpose, actor, and boundaries from its own Purpose, Use Case, and Scope content, clarifying with the human only where genuinely unclear. This scope is the fixed reference point for every materiality decision in Step 4; do not expand it beyond what this Specification's own Scope defines.

### 3. Read Relevant Authoritative Knowledge

Read the Domain, Architecture, Stack, Conventions, Design Knowledge (when user-facing), Contract, and ADR content applicable to this use case, to understand what already governs it before deciding what else is needed.

### 4. Identify and Classify Gaps

Evaluate the Specification against [`evaluations/SPECIFICATION_READINESS_CRITERIA.md`](../evaluations/SPECIFICATION_READINESS_CRITERIA.md). For each criterion, determine whether it is satisfied, represents a material gap, or is not applicable to this use case's scope. A gap is material only when leaving it unresolved would force Implementation, Testing, or Verification to invent a decision. Template completeness never determines materiality.

For each material gap, determine its owner:

- **Specification-owned** — specific to this use case's own flow, criteria, or preconditions. A dependency on another Specification (see [`SPECIFICATION_DEPENDENCIES.md`](../specifications/SPECIFICATION_DEPENDENCIES.md)) is Specification-owned content and may be established here, in the Specification's own `Dependencies` section, when its governing test is met — never merely because another Specification seemed useful to build first.
- **Owned elsewhere** — Domain, Architecture, Stack, Conventions, Design Knowledge, or a Contract.

For each declared dependency, also check its Target Presence. A target that is `MISSING` is itself a material gap for the portion of this Specification that relies on it — do not invent the missing Specification's behavior to work around it. A `DRAFT` or `APPROVED` target is not a gap; Definition proceeds normally either way.

### 5. Resolve Specification-Owned Gaps

Ask the human only about material Specification-owned gaps. Never invent a decision. Record each resolved decision directly into the Specification's own content, at the depth this use case actually needs — not full template completion.

Before recording a new or changed dependency, check whether it would create a cycle: would tracing the target's own declared dependencies, and theirs, lead back to this Specification? Do not record a dependency that would create one. If a declared dependency's target is `MISSING`, or a proposed dependency would create a cycle, stop only the portion of this Specification that depends on it and ask the human how to proceed — for example correcting or removing the dependency here, or creating the missing Specification separately, outside this workflow's scope.

### 6. Route Externally-Owned Gaps Through Knowledge Resolution

For each material gap owned elsewhere, stop only the work that depends on it, present the gap and the decision it requires to the human, and — once the human decides — run [`workflows/knowledge-resolution.md`](knowledge-resolution.md) to apply the decision to its correct authoritative owner, with an ADR created only where independently justified, before resuming. Never resolve an externally-owned gap directly inside the Specification.

### 7. Re-evaluate

Repeat Steps 4–6 until no material gap remains for this Specification's scope, or until progress is genuinely blocked on a decision the human has not yet provided.

### 8. Check Consistency

Re-read the Specification's content against the authoritative knowledge read in Step 3 for contradiction or staleness introduced by this run's changes.

### 9. Evaluate Completion Criteria

Evaluate every applicable criterion in [`evaluations/SPECIFICATION_READINESS_CRITERIA.md`](../evaluations/SPECIFICATION_READINESS_CRITERIA.md). Reach `READY_FOR_APPROVAL` only when every applicable criterion is satisfied and Step 8's consistency check is clean.

### 10. Report Result

Produce the structured result described in Outputs. Do not request, imply, or simulate Human approval; readiness is a workflow result, not a lifecycle state, and this workflow has no authority to change lifecycle state.

## Outputs

```
Specification Definition Result

Target:
Outcome: READY_FOR_APPROVAL | BLOCKED
Remaining unresolved:
Non-blocking deferred items:
Consistency check: CLEAN | ISSUES_FOUND
Resolved this run (optional):
```

- **Target** — the Specification this result concerns.
- **Outcome** — `READY_FOR_APPROVAL` only when every applicable Completion Criterion is satisfied and Step 8 is clean; `BLOCKED` otherwise.
- **Remaining unresolved** — populated when `BLOCKED`: each material gap, its owner, and whether a human decision is pending.
- **Non-blocking deferred items** — criteria resolved NOT APPLICABLE, and any other knowingly deferred non-material unknown.
- **Consistency check** — `CLEAN`, or `ISSUES_FOUND` with the issues listed.
- **Resolved this run** — optional reporting detail; consumers must be able to interpret the result correctly without it.

## Rules

- Never invent a material product, business, domain, architectural, technical, or policy decision; ask the human instead.
- Judge materiality only against this Specification's own scope (Step 2) and [`evaluations/SPECIFICATION_READINESS_CRITERIA.md`](../evaluations/SPECIFICATION_READINESS_CRITERIA.md) — never against template completeness.
- Preserve existing valid content; do not rewrite manually authored or previously Agent-authored material without cause.
- Resolve only Specification-owned decisions directly; route every externally-owned material gap through Knowledge Resolution rather than resolving it inline.
- Keep definition at business/behavioral level; do not force implementation-specific technical detail into a Specification.
- If defining this Specification reveals that a different, already-Approved Specification may need to change, report that as a finding; never modify the other Specification, and never introduce dependency-ordering behavior.
- `READY_FOR_APPROVAL` is a workflow result only. It never changes SPEC lifecycle state and never constitutes or implies Human approval.
- Never add a Human confirmation step to this workflow; approval remains governed solely by `SPECIFICATION_LIFECYCLE.md`.
- Do not discover Specifications, determine dependencies between them, or perform Bootstrap, Implementation, Testing, Verification, or Review.
- Treat a `MISSING` dependency target as a material gap; never invent the missing Specification's behavior to work around it.
- Never record a dependency that would create a cycle; surface it as a material issue requiring human resolution instead.

## Failure Handling

### Insufficient Initial Content

The Specification's Purpose, Use Case, and Scope are too thin to determine what else is material. Ask the human a small set of clarifying questions about purpose, actor, and goal rather than inventing them, and report `BLOCKED` rather than guessing.

### Contradictory Existing Content

Existing Specification content conflicts with itself or with authoritative knowledge read in Step 3. Identify the conflict and ask the human to resolve it; do not silently rewrite the content and do not silently trust either side.

### Human Defers a Material Decision

Continue resolving unrelated material gaps; stop only the work that depends on the deferred one. Report `BLOCKED` rather than `READY_FOR_APPROVAL` while any material gap remains open.

### Dependency Target Missing

A declared dependency's Target Presence is `MISSING`. Stop only the portion of this Specification that relies on it, ask the human how to proceed, continue unaffected work, and report `BLOCKED` if this prevents `READY_FOR_APPROVAL`. Do not invent the missing Specification's behavior to work around it.

### Dependency Would Create a Cycle

Recording or changing a dependency would create a cycle with dependencies already declared elsewhere. Do not record it. Surface the cycle as a material issue requiring human resolution, stop only the affected portion, continue unaffected work, and report `BLOCKED` if this prevents `READY_FOR_APPROVAL`.

## Completion Criteria

Specification Definition is complete when the Specification's existing content has been inspected without assuming it was empty; relevant authoritative knowledge has been read; every applicable criterion in [`evaluations/SPECIFICATION_READINESS_CRITERIA.md`](../evaluations/SPECIFICATION_READINESS_CRITERIA.md) has been evaluated against this Specification's own scope, not template completeness; every declared dependency's Target Presence has been checked, with any `MISSING` target treated as a material gap; no declared or proposed dependency creates a cycle; every Specification-owned material gap has been resolved directly and every externally-owned material gap has been routed through Knowledge Resolution and applied to its correct owner; no material gap remains, or the workflow has correctly reported `BLOCKED` because one does; the consistency check is clean; and the reported Outcome accurately reflects whether `READY_FOR_APPROVAL`'s criteria were actually met — never implying Human approval, which remains governed solely by `SPECIFICATION_LIFECYCLE.md`.
