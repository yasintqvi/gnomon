# Implementation Workflow

## Purpose

Define the execution process for the **Implement Approved Change** intent.

This workflow describes how the system transforms approved project knowledge into a verified implementation while preserving architectural boundaries, business rules, and project conventions.

Implementation produces or modifies project artifacts but never defines new product, business, or architectural knowledge.

---

## Intent

Implement an approved Specification or explicitly authorized change using approved project knowledge.

---

## When to Use

Use this workflow when:

* An approved Specification exists.
* A feature, enhancement, defect fix, or authorized change must be implemented.
* Existing implementation artifacts require modification.
* Supporting implementation artifacts must be created.

Do not use this workflow when:

* Product behavior is still undefined.
* Architecture is still being designed.
* The task is limited to environment setup.
* The task is limited to review or testing.
* The requested change requires a new business or architectural decision.

---

## Inputs

### User Input

The approved implementation request, target Specification, implementation scope, and explicit constraints.

### Project Context

The current repository, existing implementation, project structure, related artifacts, and user-owned changes.

### Required Knowledge

* Relevant Specification
* Relevant Domain knowledge
* `ARCHITECTURE.md`
* `STACK.md`
* `CONVENTIONS.md`
* Applicable ADRs
* Applicable Artifact Contracts
* Existing implementation

---

## Execution

### 1. Understand

#### Purpose

Understand the approved behavior, implementation scope, constraints, and expected outcome.

#### Required Knowledge

* Relevant Specification
* Relevant Domain knowledge

#### Expected Result

A complete understanding of:

* Requested behavior
* Acceptance criteria
* Scope
* Constraints
* Assumptions
* Explicit exclusions

---

### 2. Analyze

#### Purpose

Analyze the existing implementation and determine the smallest coherent implementation strategy.

#### Required Knowledge

* Existing implementation
* Architecture
* Applicable ADRs
* Applicable Artifact Contracts

#### Expected Result

An implementation impact analysis identifying:

* Existing reusable implementation
* Required modifications
* Dependencies
* Integration points
* Risks
* Affected artifacts
* Required verification scope

---

### 3. Design

#### Purpose

Design how the approved behavior maps onto the existing implementation without introducing new project knowledge.

#### Required Knowledge

* Relevant Specification
* Architecture
* Applicable Artifact Contracts
* Existing implementation

#### Expected Result

An implementation plan defining:

* Responsibilities
* Collaboration between artifacts
* Data flow
* Error handling
* Integration boundaries
* Required implementation artifacts

---

### 4. Implement

#### Purpose

Create or modify the required implementation artifacts.

#### Required Knowledge

* Approved implementation plan
* Project conventions
* Existing project patterns
* Applicable Artifact Contracts

#### Expected Result

Implementation artifacts that:

* Satisfy the approved Specification
* Respect project Architecture
* Follow project Conventions
* Preserve unrelated behavior
* Integrate correctly with existing artifacts

Implementation should reuse existing patterns whenever appropriate rather than introducing unnecessary variation.

---

### 5. Verify

#### Purpose

Verify that the implementation satisfies the approved behavior and remains consistent with project knowledge.

#### Required Knowledge

* Relevant Specification
* Applicable Artifact Contracts
* Review Checklist
* Relevant testing requirements

#### Expected Result

Evidence that:

* Acceptance criteria are satisfied.
* Required implementation artifacts were produced.
* Applicable verification has completed successfully.
* Architecture and Contracts have been respected.
* Remaining limitations are explicitly reported.

---

## Rules

* Understand before implementing.
* Analyze existing implementation before introducing new artifacts.
* Reuse existing project patterns whenever appropriate.
* Work only within the approved implementation scope.
* Preserve unrelated user-owned changes.
* Produce implementation consistent with approved project knowledge.
* Follow applicable Artifact Contracts.
* Respect project Architecture, Stack, and Conventions.
* Avoid speculative implementation.
* Avoid unnecessary abstraction.
* Prefer the smallest coherent implementation.
* Keep generated artifacts internally consistent.
* Produce verifiable implementation.
* Never redefine product, business, or architectural knowledge during implementation.

The workflow must not introduce business rules that are not defined by the relevant Specification or Domain.

The workflow must not introduce architectural rules that are not defined by the Architecture or an approved ADR.

---

## Outputs

### Primary Output

A verified implementation of the approved change.

### Supporting Outputs

* Updated implementation artifacts
* Supporting implementation artifacts
* Required documentation updates
* Required automated tests when applicable
* Verification evidence

---

## Failure Handling

### Missing Information

When implementation depends on missing project knowledge:

1. Identify the missing information.
2. Determine which project document owns it.
3. Continue unaffected implementation where possible.
4. Stop only the affected implementation work.
5. Request clarification only when progress materially depends on the missing decision.

Do not invent missing behavior.

---

### Conflicting Information

When approved project knowledge conflicts:

1. Identify the conflicting sources.
2. Apply the project's knowledge ownership rules.
3. Do not silently choose one interpretation.
4. Stop only the affected implementation work.
5. Continue unaffected implementation where possible.

---

### Implementation Conflict

When the requested implementation conflicts with the existing repository:

1. Identify the conflicting artifacts.
2. Preserve user-owned changes.
3. Prefer modification over replacement when appropriate.
4. Report destructive changes before performing them.
5. Require explicit authorization for destructive operations.

---

### Verification Failure

When implementation verification fails:

1. Diagnose the failure.
2. Correct implementation defects within scope.
3. Re-run the relevant verification.
4. Report unresolved failures honestly.
5. Do not weaken verification merely to obtain a successful result.

---

## Completion Criteria

The workflow is complete when:

* The approved Specification has been implemented.
* Acceptance criteria have been satisfied.
* Applicable implementation artifacts have been updated.
* Required verification has completed.
* Architecture and Contracts remain respected.
* Existing unrelated behavior has been preserved.
* Remaining assumptions, risks, and limitations have been reported.

---

## Workflow Constraints

* Implementation must remain within the approved scope.
* Implementation must not redefine project knowledge.
* Prefer modification over unnecessary replacement.
* Prefer minimal, maintainable, and verifiable implementation.
* Avoid speculative implementation for future requirements.

---

## Notes

Implementation transforms approved project knowledge into executable project artifacts.

Questions about product behavior, architecture, or project policy discovered during implementation should be reported to the document that owns that knowledge rather than being resolved implicitly during coding.
