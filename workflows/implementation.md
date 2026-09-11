# Implementation Workflow

## Purpose

Implement an approved change as the smallest coherent, maintainable, and verifiable modification consistent with authoritative project knowledge.

Implementation translates approved knowledge into project artifacts. It does not define product behavior, business rules, architecture, design, project policy, or technology choices.

## When to Use

Use this workflow to create, modify, refactor, or integrate implementation artifacts for approved behavior. Use Bootstrap first when the repository lacks the required baseline, and use Testing or Verification when the task is evaluation-only.

## Inputs

- Authorized change and scope
- Relevant Specifications and Domain knowledge
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

- Verified implementation of the approved change
- Updated or supporting artifacts and documentation
- Required automated tests
- Reproducible verification evidence
- Remaining assumptions, risks, limitations, or unresolved decisions

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

## Failure Handling

### Missing or Conflicting Knowledge

Identify the issue and its owning document, apply established ownership or precedence rules, continue unaffected work, and stop only the affected portion when the issue remains material. Do not silently choose or invent behavior, policy, architecture, design, or technology. Once a human decision resolves the gap, run [`workflows/knowledge-resolution.md`](knowledge-resolution.md) to update the affected authoritative knowledge before resuming the affected portion.

### Repository Conflict

Identify conflicting artifacts, preserve user-owned changes, prefer compatible modification, and obtain explicit authorization before destructive action.

### Verification Failure

Diagnose and correct in-scope implementation defects, rerun relevant verification, and report unresolved failures honestly.

## Completion Criteria

Implementation is complete when the approved scope and acceptance criteria are satisfied; applicable artifacts, tests, and documentation are updated; relevant project and Design Knowledge is respected; required verification has run; unrelated behavior is preserved; and remaining limitations are reported.
