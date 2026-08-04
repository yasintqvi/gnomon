# Testing Workflow

## Purpose

Define the execution process for the **Verify Behavior Through Testing** intent.

This workflow describes how the system should analyze test requirements, design appropriate tests, implement or execute them, and report evidence about the correctness of approved behavior.

It verifies implementation against Specifications, Domain rules, Architecture, Artifact Contracts, and project testing conventions without changing approved behavior.

---

## Intent

Verify that an existing or newly implemented change behaves correctly, satisfies its acceptance criteria, preserves relevant constraints, and does not introduce unintended regressions.

---

## When to Use

Use this workflow when:

* An approved Specification or explicitly scoped behavior must be verified.
* Automated tests must be created, updated, or executed.
* Existing behavior must be validated after implementation or refactoring.
* A defect must be reproduced and verified through testing.
* Test coverage gaps must be identified for an approved scope.

Do not use this workflow when:

* The requested behavior is still undefined or requires a product decision.
* The primary task is feature implementation rather than verification.
* The task is limited to code review without executing or designing tests.
* The project baseline does not yet support testing and requires Bootstrap first.

Testing may be performed independently or as part of the Verification stage of another Workflow.

---

## Inputs

### User Input

The target behavior, Specification, implementation scope, defect description, verification objective, and any explicit test boundaries.

Examples may include:

* A target Specification
* A changed module or artifact
* A reported defect
* A required test level
* A restricted test scope
* A required environment or dataset

### Project Context

The current repository state, existing implementation, existing test suite, test configuration, fixtures, environments, and user-owned changes.

### Required Knowledge

* Relevant Specification
* Relevant Domain rules
* `ARCHITECTURE.md`
* `CONVENTIONS.md`
* `STACK.md`
* Applicable ADRs
* Applicable Artifact Contracts
* `REVIEW_CHECKLIST.md`
* Existing tests and testing infrastructure

The Specification owns feature behavior and acceptance criteria.

Domain knowledge owns cross-feature business rules and invariants.

Artifact Contracts own artifact-specific verification requirements.

Project testing conventions own test naming, organization, and structure.

---

## Execution

### 1. Understand

**Purpose**

Establish exactly what behavior must be verified and what evidence is required.

**Required Knowledge**

* Relevant Specification
* Relevant Domain rules
* User-provided verification scope
* Existing defect or change description

**Expected Result**

A clear verification scope containing:

* Target behavior
* Acceptance criteria
* Applicable business rules
* Relevant success and failure cases
* Explicit exclusions
* Known assumptions
* Blocking ambiguities

---

### 2. Inspect

**Purpose**

Inspect the implementation and existing test suite before designing or changing tests.

**Required Knowledge**

* Current implementation
* Existing tests
* Testing configuration
* Applicable Artifact Contracts
* Project Architecture and Conventions

**Expected Result**

A testing impact analysis identifying:

* Existing relevant tests
* Missing or outdated coverage
* Affected boundaries and dependencies
* Available fixtures, factories, helpers, and test utilities
* Testability constraints
* Regression risks
* The smallest coherent verification scope

---

### 3. Design

**Purpose**

Design the minimum sufficient set of tests required to verify the approved behavior and relevant risks.

**Required Knowledge**

* Verification scope
* Relevant Specification and acceptance criteria
* Domain rules
* Applicable Artifact Contracts
* Existing testing patterns

**Expected Result**

A test design defining, where applicable:

* Primary success scenarios
* Relevant failure scenarios
* Boundary conditions
* Authorization and ownership cases
* Validation cases
* State transitions
* Persistence behavior
* Integration boundaries
* Retry, failure, or concurrency behavior
* Regression coverage
* Required test level

The design must select the lowest test level that provides reliable evidence while adding broader tests where cross-boundary behavior requires them.

---

### 4. Prepare

**Purpose**

Prepare the test environment, data, dependencies, and supporting test artifacts required for reliable execution.

**Required Knowledge**

* Test design
* Existing test infrastructure
* `STACK.md`
* `CONVENTIONS.md`
* Applicable test-related contracts

**Expected Result**

A deterministic and isolated test setup containing only what is required for the approved verification scope.

Preparation may include:

* Test data
* Fixtures
* Factories
* Fakes
* Stubs
* Mocks
* Test doubles
* Environment configuration
* Required test helpers

Test preparation must not alter production behavior merely to make testing easier unless the change is explicitly authorized and architecturally valid.

---

### 5. Implement

**Purpose**

Create or update the tests required by the approved test design.

**Required Knowledge**

* Approved test design
* Project testing conventions
* Existing test patterns
* Applicable Artifact Contracts

**Expected Result**

Focused automated tests that:

* Verify observable behavior rather than implementation trivia
* Follow project conventions
* Remain deterministic
* Clearly communicate the behavior under test
* Avoid unnecessary duplication
* Preserve unrelated existing tests
* Fail for the intended reason before a required implementation fix, when applicable

This stage may be skipped when the authorized task is execution-only and sufficient tests already exist.

---

### 6. Execute

**Purpose**

Run the smallest relevant set of tests and expand execution only when results or risk justify it.

**Required Knowledge**

* Target tests
* Test commands and configuration
* Relevant build and runtime requirements

**Expected Result**

Recorded test results showing:

* Commands executed
* Tests passed
* Tests failed
* Tests skipped
* Environment limitations
* Reproducibility of failures
* Relevant logs or diagnostic evidence

Execution should normally progress from narrow to broad:

1. Targeted test
2. Related test group or module
3. Broader regression suite when justified
4. Build, static analysis, or integration checks when applicable

---

### 7. Evaluate

**Purpose**

Determine whether the test results provide sufficient evidence that the approved behavior is correct.

**Required Knowledge**

* Test results
* Relevant Specification
* Acceptance criteria
* Domain rules
* Applicable Artifact Contracts
* Review Checklist

**Expected Result**

An evidence-based evaluation identifying:

* Verified behavior
* Failed behavior
* Unverified behavior
* Regression impact
* Test infrastructure failures
* Product or Specification ambiguity
* Remaining risks and limitations

A passing test suite is not sufficient when the tests do not cover the required behavior.

---

### 8. Report

**Purpose**

Present the verification result clearly and distinguish implementation defects from test defects, environment failures, and unresolved requirements.

**Required Knowledge**

* Evaluation result
* Executed test evidence
* Remaining limitations

**Expected Result**

A concise testing report containing:

* Scope tested
* Tests created or modified
* Commands executed
* Results
* Confirmed acceptance criteria
* Failures and their classification
* Unverified conditions
* Remaining risks
* Recommended next action

---

## Rules

* Inspect existing tests and implementation before creating new tests.
* Verify approved behavior rather than inventing new requirements.
* Derive expected outcomes from the relevant Specification and Domain.
* Follow applicable Artifact Contracts and testing conventions.
* Prefer deterministic, isolated, and reproducible tests.
* Test observable behavior rather than private implementation details unless a lower-level contract explicitly requires it.
* Reuse existing test infrastructure when it remains valid.
* Avoid excessive mocking that removes the behavior being verified.
* Mock or fake external boundaries only when doing so preserves the purpose of the test.
* Do not modify production behavior solely to obtain a passing test.
* Do not weaken, delete, skip, or bypass a valid existing test without explicit justification and authorization.
* Do not interpret unrelated pre-existing failures as failures caused by the current change.
* Distinguish test failures, implementation defects, environment failures, and requirement ambiguity.
* Run the smallest relevant checks first and expand verification according to demonstrated risk.
* Preserve unrelated user-owned changes.
* Report commands that were not run and explain why.
* Never claim verification without evidence.

The workflow must not introduce business rules that are not defined by the relevant Specification or Domain.

The workflow must not introduce architectural rules that are not defined by the Architecture or an approved ADR.

---

## Outputs

### Primary Output

Evidence-based verification of the target behavior.

### Supporting Outputs

* New or updated automated tests
* Test data, fixtures, factories, or test helpers when required
* Reproduction tests for confirmed defects
* Test execution evidence
* Coverage gap analysis
* Classified failure report
* Remaining risk and limitation report

---

## Failure Handling

### Missing Information

When expected behavior or test scope is unclear:

1. Identify the missing information.
2. Determine which project document owns it.
3. Decide whether reliable testing can continue without it.
4. Stop only the affected verification work when it is blocking.
5. Continue unaffected testing where possible.
6. Request clarification only when the missing decision materially changes expected behavior.

Do not invent expected outcomes.

---

### Conflicting Information

When project knowledge defines conflicting expectations:

1. Identify the conflicting sources explicitly.
2. Apply the project's knowledge ownership rules.
3. Do not encode one interpretation silently in tests.
4. Stop the affected test design when the conflict cannot be resolved within the Workflow's authority.
5. Continue unaffected verification where possible.

---

### Test Infrastructure Failure

When tests cannot run because of environment or infrastructure problems:

1. Confirm that the failure is unrelated to the behavior under test.
2. Diagnose the test environment or configuration.
3. Correct testing-related issues within the authorized scope.
4. Re-run the smallest relevant command.
5. Report unresolved infrastructure blockers separately from product failures.

Do not classify an infrastructure failure as an implementation defect.

---

### Test Failure

When a test fails:

1. Reproduce the failure.
2. Confirm that the test expectation is supported by approved project knowledge.
3. Determine whether the cause is:

   * An implementation defect
   * A test defect
   * A fixture or data defect
   * An environment failure
   * A requirement ambiguity
4. Correct test-caused defects within scope.
5. Report implementation defects rather than silently changing expected behavior.
6. Re-run the affected test after correction.

---

### Flaky Test

When a test produces inconsistent results:

1. Re-run it enough to confirm nondeterminism.
2. Identify shared state, timing, concurrency, network, or environment causes.
3. Stabilize the test within scope.
4. Do not hide flakiness through retries unless retries are themselves part of the behavior being tested.
5. Report unresolved flakiness as a verification limitation.

---

### Existing Suite Failure

When unrelated tests already fail:

1. Confirm that the failure existed independently of the current scope when possible.
2. Separate pre-existing failures from newly introduced failures.
3. Do not modify unrelated tests without authorization.
4. Report their effect on confidence in the final verification result.

---

## Completion Criteria

The workflow is complete when:

* The authorized verification scope has been tested.
* Applicable acceptance criteria have corresponding evidence.
* Relevant success, failure, and boundary scenarios have been covered.
* Required tests have been created or updated.
* Relevant tests have been executed.
* Test results have been evaluated against approved project knowledge.
* Failures have been classified accurately.
* Newly introduced defects within scope have been corrected or reported.
* No valid checks have been weakened or bypassed.
* Unexecuted checks and unverified conditions have been reported.
* Remaining risks and limitations are explicit.
* The final testing report is supported by reproducible evidence.

---

## Workflow Constraints

* Testing verifies approved behavior; it does not define new behavior.
* Testing work must remain within the authorized scope.
* The workflow may create or modify test artifacts but must not change production behavior unless implementation changes are explicitly authorized.
* Prefer the smallest sufficient test set that provides reliable evidence.
* Expand to broader regression testing according to actual impact and risk.
* Avoid speculative tests for hypothetical future requirements.
* Do not require every behavior to be tested at every test level.
* Preserve the distinction between unit, integration, feature, contract, system, and end-to-end testing where the project defines those levels.
* Verification confidence must reflect what was actually executed, not what was merely designed.

---

## Notes

Testing may expose implementation defects, missing Specifications, weak Artifact Contracts, architectural ambiguity, or limitations in the testing environment.

Such findings should be reported to the document or Workflow that owns the unresolved issue rather than being silently resolved inside the test suite.
