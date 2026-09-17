---
identity: testing
specification_reference: optional
requires_approved_specification: true
result:
  terminal_path: outcome
  classification:
    TESTING_COMPLETE: success
    BLOCKED: blocked
  schema:
    type: object
    required:
      - outcome
    properties:
      outcome:
        type: string
        enum:
          - TESTING_COMPLETE
          - BLOCKED
      evidence:
        type:
          - string
          - "null"
      coverage:
        type:
          - string
          - "null"
      remaining_unresolved:
        type:
          - string
          - "null"
---

# Testing Workflow

## Purpose

Design, implement, execute, and evaluate tests that provide reproducible evidence for approved behavior without defining or changing that behavior.

## When to Use

Use this workflow to verify a Specification or explicitly scoped behavior, validate a change or refactor, reproduce a defect, or close an authorized coverage gap. Use Bootstrap when test infrastructure is absent, Implementation when production changes are primary, and Review when no test work is requested.

When the behavior under test is governed by a Specification, that Specification must be `Approved`, per [`SPECIFICATION_LIFECYCLE.md`](../specifications/SPECIFICATION_LIFECYCLE.md). Do not treat `Draft` content as an authoritative test contract, and do not write tests against it under any authorization — run [`workflows/specification-definition.md`](specification-definition.md) and obtain Human Approval first.

## Inputs

* Target behavior, acceptance criteria, defect, or authorized test scope
* Existing implementation, tests, configuration, fixtures, and environment
* Relevant Specifications and Domain knowledge — a governing Specification must be `Approved`
* Each governing Specification's `Dependencies` (see [`SPECIFICATION_DEPENDENCIES.md`](../specifications/SPECIFICATION_DEPENDENCIES.md)) and their current Target Presence and Target Lifecycle State
* ARCHITECTURE.md, STACK.md, and CONVENTIONS.md, as applicable
* Applicable ADRs and Artifact Contracts
* Applicable Design Knowledge when user-facing behavior is under test:

  * PRODUCT_EXPERIENCE.md
  * UI_FOUNDATION.md
  * INTERACTION_PATTERNS.md
* Applicable verification criteria
* Explicit test-level, environment, data, or execution constraints

Expected outcomes come from their authoritative sources. Tests do not create product, business, architecture, design, or policy decisions.

## Execution

### 1. Establish Scope

Identify target behavior, acceptance criteria, applicable rules, success and failure cases, exclusions, assumptions, and blocking ambiguities.

When the target behavior is governed by a Specification, confirm it is `Approved` before designing tests against it. For each Specification the target behavior depends on (per its `Dependencies` section), check Target Presence and Target Lifecycle State: do not treat a `MISSING` or `DRAFT` target's behavior as an authoritative test contract — exclude the affected tests from this run rather than testing against unconfirmed behavior.

### 2. Inspect

Inspect implementation and test infrastructure. Identify existing coverage, affected boundaries, dependencies, reusable fixtures and helpers, testability constraints, regression risks, and the smallest coherent testing scope.

### 3. Design

Choose the minimum sufficient tests and the lowest reliable test level capable of actually exercising the behavior being tested.

Cover applicable success, failure, boundary, authorization, validation, state-transition, persistence, integration, retry, concurrency, regression, and Design Knowledge obligations according to actual behavior and risk.

When behavior materially depends on runtime rendering, interaction, integration, or execution, use a test level that exercises that behavior. Do not substitute source-level assertions or lower-level tests that cannot reliably demonstrate the required outcome.

### 4. Prepare and Implement

Prepare deterministic, isolated data and test support, then create or update focused tests that verify observable behavior and follow existing conventions. Reuse valid infrastructure and preserve unrelated tests. An execution-only task may skip implementation when sufficient tests exist.

Do not alter production behavior solely for test convenience unless explicitly authorized and consistent with project knowledge.

### 5. Execute

Run the narrowest relevant test first, then expand to related tests, broader regression, builds, static analysis, integration, or runtime-level checks when required by the behavior, impact, or results.

Record commands, passes, failures, skips, logs, and environment limitations.

### 6. Evaluate and Report

Determine whether the executed tests provide sufficient evidence for the approved behavior. A passing suite is insufficient when the required behavior was not actually exercised.

Report:

* scope and tests created or modified;
* commands and results;
* verified, failed, and unverified behavior;
* confirmed acceptance criteria;
* regression impact;
* classified failures and infrastructure limitations;
* checks not run, remaining risks, and next action.

## Rules

* Verify approved observable behavior; do not invent expected outcomes.
* Prefer deterministic, isolated, reproducible tests and existing valid infrastructure.
* Use the lowest test level that can reliably demonstrate the required behavior.
* Do not infer runtime, rendered, interactive, or integration behavior from tests that do not actually exercise it.
* Mock external boundaries only when the test still exercises its intended behavior.
* Do not weaken, delete, skip, or bypass a valid test without explicit justification and authorization.
* Distinguish implementation, test, fixture/data, environment, and knowledge failures.
* Separate unrelated pre-existing failures from failures caused by the current scope.
* Preserve user-owned changes and remain within authorized scope.
* Never claim evidence beyond what was actually executed.
* Do not modify production behavior unless implementation changes are explicitly authorized.
* Never write or run tests that treat a `Draft` Specification's behavior as authoritative; no authorization permits this exception.
* Exclude only the tests that materially depend on a `MISSING` or `DRAFT` dependency target; continue unaffected testing.

## Outputs

```
Testing Result

Outcome: TESTING_COMPLETE | BLOCKED
Evidence:
Coverage:
Remaining unresolved (BLOCKED only):
```

- **Outcome** — `TESTING_COMPLETE` only when the authorized scope has reproducible test evidence and no valid check was weakened; `BLOCKED` when a governing Specification is not `Approved`, a dependency's current state excludes affected tests, or a material gap remains unresolved.
- **Evidence** — informational: execution evidence, classified failures, and coverage-gap analysis.
- **Coverage** — informational: new or updated tests and required test support.
- **Remaining unresolved** — populated only when `BLOCKED`: the unapproved Specification, the excluded tests and the dependency causing exclusion, or the material decision still open.

## Failure Handling

### Governing Specification Not Approved

Exclude the affected tests and report `BLOCKED`. Direct the Specification to [`workflows/specification-definition.md`](specification-definition.md) or to Human Approval as applicable. Do not design or run tests against its `Draft` content under any authorization.

### Dependency Not Yet Available

A dependency's Target Presence is `MISSING`, or its Target Lifecycle State is `DRAFT`, and the affected tests would treat that target's behavior as authoritative. Exclude only those tests, continue unaffected testing, and report the exclusion. Rerun the affected tests once the dependency becomes `Approved`.

### Missing or Conflicting Knowledge

Identify the issue and its owner, apply established ownership or precedence rules, and do not encode an invented interpretation. Continue unaffected work; stop affected test design only when the decision materially changes the expected result. Once a human decision resolves the gap, run [`workflows/knowledge-resolution.md`](knowledge-resolution.md) to update the affected authoritative knowledge before resuming affected test design.

### Infrastructure Failure

Confirm the failure is distinct from target behavior, diagnose and correct in-scope test infrastructure, rerun the smallest relevant command, and report unresolved blockers separately from implementation defects.

### Test Failure

Reproduce the failure, confirm the expectation against approved knowledge, classify its cause, correct in-scope test or fixture defects, rerun the affected test, and report implementation defects rather than changing expectations.

### Flaky Test

Confirm nondeterminism, investigate relevant causes, and stabilize within scope. Do not hide flakiness with retries unless retry behavior is itself under test; report unresolved flakiness as a limitation.

### Existing Suite Failure

Establish pre-existence where possible, keep it separate from current failures, avoid unauthorized unrelated edits, and report its effect on confidence.

## Completion Criteria

Testing is complete when any governing Specification is confirmed `Approved`; the authorized scope and applicable acceptance criteria have reproducible test evidence; relevant behavior has been exercised at a sufficient and reliable test level; required tests ran; failures are accurately classified; no valid check was weakened; every dependency excluding affected tests has been identified rather than silently bypassed; and unexecuted checks, unverified conditions, risks, and limitations are explicit.
