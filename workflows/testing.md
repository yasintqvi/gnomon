# Testing Workflow

## Purpose

Design, implement, execute, and evaluate tests that provide reproducible evidence for approved behavior without defining or changing that behavior.

## When to Use

Use this workflow to verify a Specification or explicitly scoped behavior, validate a change or refactor, reproduce a defect, or close an authorized coverage gap. Use Bootstrap when test infrastructure is absent, Implementation when production changes are primary, and Review when no test work is requested.

## Inputs

* Target behavior, acceptance criteria, defect, or authorized test scope
* Existing implementation, tests, configuration, fixtures, and environment
* Relevant Specifications and Domain knowledge
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

## Outputs

* Evidence-based testing of the target behavior
* New or updated tests and required test support
* Execution evidence and coverage-gap analysis
* Classified failures, unverified conditions, risks, and limitations

## Failure Handling

### Missing or Conflicting Knowledge

Identify the issue and its owner, apply established ownership or precedence rules, and do not encode an invented interpretation. Continue unaffected work; stop affected test design only when the decision materially changes the expected result.

### Infrastructure Failure

Confirm the failure is distinct from target behavior, diagnose and correct in-scope test infrastructure, rerun the smallest relevant command, and report unresolved blockers separately from implementation defects.

### Test Failure

Reproduce the failure, confirm the expectation against approved knowledge, classify its cause, correct in-scope test or fixture defects, rerun the affected test, and report implementation defects rather than changing expectations.

### Flaky Test

Confirm nondeterminism, investigate relevant causes, and stabilize within scope. Do not hide flakiness with retries unless retry behavior is itself under test; report unresolved flakiness as a limitation.

### Existing Suite Failure

Establish pre-existence where possible, keep it separate from current failures, avoid unauthorized unrelated edits, and report its effect on confidence.

## Completion Criteria

Testing is complete when the authorized scope and applicable acceptance criteria have reproducible test evidence; relevant behavior has been exercised at a sufficient and reliable test level; required tests ran; failures are accurately classified; no valid check was weakened; and unexecuted checks, unverified conditions, risks, and limitations are explicit.
