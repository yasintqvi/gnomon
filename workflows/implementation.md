---
identity: implementation
specification_reference: required
requires_approved_specification: true
result:
  terminal_path: outcome
  classification:
    IMPLEMENTATION_COMPLETE: success
    BLOCKED: blocked
  schema:
    type: object
    required:
      - outcome
    properties:
      outcome:
        type: string
        enum:
          - IMPLEMENTATION_COMPLETE
          - BLOCKED
      delivered:
        type: string
      verification_evidence:
        type: string
      remaining_unresolved:
        type:
          - string
          - "null"
---

# Implementation Workflow

## Purpose

Implement an approved change as the smallest coherent, maintainable, and verifiable modification consistent with authoritative project knowledge.

Implementation translates approved knowledge into project artifacts. It does not define product behavior, business rules, architecture, design, project policy, or technology choices.

## When to Use

Use this workflow to create, modify, refactor, or integrate implementation artifacts for approved behavior. Use Bootstrap first when the repository lacks the required baseline, and use Testing or Verification when the task is evaluation-only.

The Specification governing this work must be `Approved`, per [`SPECIFICATION_LIFECYCLE.md`](../specifications/SPECIFICATION_LIFECYCLE.md). If it is `Draft`, do not implement its behavior — run [`workflows/specification-definition.md`](specification-definition.md) and obtain Human Approval first. There is no authorization that permits implementing against a `Draft` Specification; this is a hard precondition, not a risk to be accepted.

## Inputs

- Authorized change and scope
- Relevant Specifications and Domain knowledge — the governing Specification must be `Approved`
- Each governing Specification's `Dependencies` (see [`SPECIFICATION_DEPENDENCIES.md`](../specifications/SPECIFICATION_DEPENDENCIES.md)) and their current Target Presence and Target Lifecycle State
- Existing repository, implementation, tests, and user-owned changes
- `PROJECT.md`, `ARCHITECTURE.md`, `STACK.md`, and `CONVENTIONS.md`, as applicable
- Applicable ADRs and Artifact Contracts
- For user-facing work, applicable Design Knowledge:
  - `PRODUCT_EXPERIENCE.md` for product context, discoverability, hierarchy, and journeys
  - `UI_FOUNDATION.md` for visual and presentation rules
  - `INTERACTION_PATTERNS.md` for shared interaction behavior
- Explicit constraints and required verification

Load only knowledge relevant to the authorized change. Specifications own feature behavior; Design Knowledge governs how user-facing behavior fits and appears without authorizing new capabilities.

## Execution

### 1. Understand

Establish the intended behavior, acceptance criteria, scope, constraints, applicable knowledge, assumptions, and unresolved decisions. Stop only affected work when a missing decision materially changes the result; do not invent it.

Confirm the governing Specification is `Approved`. If it is not, stop before implementing its behavior — see Failure Handling. For each Specification the work depends on (per its `Dependencies` section), check Target Presence and Target Lifecycle State: proceed on portions relying on a `MISSING` or `DRAFT` target only if they do not materially depend on that target's guarantee; stop only the portion that does.

### 2. Analyze

Inspect the current implementation before designing changes. Identify affected artifacts and boundaries, existing patterns to reuse, applicable contracts, dependencies, tests, integration points, migration or compatibility concerns, and unrelated changes to preserve.

For user-facing work, also identify the governing product context, experience integration, presentation foundation, and interaction patterns from Design Knowledge.

### 3. Design

Define the smallest coherent implementation plan, including:

- artifacts to create or modify;
- responsibilities and dependency direction;
- data flow, state transitions, validation, errors, and integration boundaries;
- test and verification strategy;
- for user-facing work, the applicable experience, visual, responsive, accessibility, localization, and interaction obligations derived from Design Knowledge.

Reuse established architecture, patterns, and shared primitives. Do not introduce speculative abstractions or new project-wide decisions.

### 4. Implement

Apply the approved plan within scope. Preserve unrelated behavior and user-owned changes, follow applicable contracts and authoritative knowledge, and keep artifacts internally consistent.

User-facing implementation must consume the applicable Design Knowledge and approved Stack-defined tools rather than redefining detailed UI/UX rules locally.

### 5. Verify

Gather reproducible evidence that acceptance criteria and other applicable obligations are satisfied. Run the narrowest relevant tests and checks first, expanding according to impact and risk. Verify applicable Architecture, Stack, Conventions, Contracts, and Design Knowledge obligations, and report limitations or checks not run.

## Outputs

```
Implementation Result

Outcome: IMPLEMENTATION_COMPLETE | BLOCKED
Delivered:
Verification evidence:
Remaining unresolved (BLOCKED only):
```

- **Outcome** — `IMPLEMENTATION_COMPLETE` only when the approved scope and acceptance criteria are satisfied and required verification has run; `BLOCKED` when the governing Specification is not `Approved`, a dependency's current state blocks the affected portion, or a material gap remains unresolved.
- **Delivered** — informational: the artifacts, documentation, and tests updated or created.
- **Verification evidence** — informational: the reproducible evidence gathered in Step 5.
- **Remaining unresolved** — populated only when `BLOCKED`: the unapproved Specification, the specific dependency and portion it blocks, or the material decision still open, plus any remaining assumptions, risks, or limitations.

## Rules

- Remain within authorized scope and preserve unrelated work.
- Prefer modification and reuse over unnecessary replacement or variation.
- Never redefine or contradict authoritative project knowledge.
- Do not introduce business rules absent from Specifications or Domain knowledge.
- Do not introduce architectural rules absent from Architecture or an approved ADR.
- Do not introduce design decisions absent from applicable Design Knowledge.
- Do not introduce project-wide policy or technology choices absent from their owning sources.
- Require explicit authorization for destructive operations.
- Do not weaken verification to obtain a passing result.
- Never implement behavior governed by a `Draft` Specification; no authorization or urgency permits this exception.
- Stop only the portion of work that materially depends on a `MISSING` or `DRAFT` dependency target; continue unaffected work.

## Failure Handling

### Governing Specification Not Approved

Stop the affected work and report `BLOCKED`. Direct the Specification to [`workflows/specification-definition.md`](specification-definition.md) if it is not yet `READY_FOR_APPROVAL`, or to Human Approval if it already is. Do not proceed against its `Draft` content under any authorization.

### Dependency Not Yet Available

A dependency's Target Presence is `MISSING`, or its Target Lifecycle State is `DRAFT`, and the affected portion of work materially relies on it. Stop only that portion, continue unaffected work, and report the blocked portion. Once the dependency becomes `Approved` (or is otherwise resolved), rerun the affected portion.

### Missing or Conflicting Knowledge

Identify the issue and its owning document, apply established ownership or precedence rules, continue unaffected work, and stop only the affected portion when the issue remains material. Do not silently choose or invent behavior, policy, architecture, design, or technology. Once a human decision resolves the gap, run [`workflows/knowledge-resolution.md`](knowledge-resolution.md) to update the affected authoritative knowledge before resuming the affected portion.

### Repository Conflict

Identify conflicting artifacts, preserve user-owned changes, prefer compatible modification, and obtain explicit authorization before destructive action.

### Verification Failure

Diagnose and correct in-scope implementation defects, rerun relevant verification, and report unresolved failures honestly.

## Completion Criteria

Implementation is complete when the governing Specification is confirmed `Approved`; the approved scope and acceptance criteria are satisfied; applicable artifacts, tests, and documentation are updated; relevant project and Design Knowledge is respected; required verification has run; unrelated behavior is preserved; every dependency blocking an affected portion has been identified rather than silently bypassed; and the reported Outcome accurately reflects whether `IMPLEMENTATION_COMPLETE`'s conditions were actually met.
