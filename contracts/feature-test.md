# Artifact Contract — Feature Test

## Purpose

Define the contract for executable feature-level tests that verify observable application behavior across supported boundaries.

This contract defines the canonical characteristics of a Feature Test independently of any project, framework, or implementation technology.

---

## Artifact Type

**Name**

Feature Test

**Description**

An automated test that verifies an observable application behavior through a supported application boundary while controlling external and nondeterministic dependencies.

---

## Scope

This contract applies to:

* Application behavior spanning multiple collaborating artifacts.
* Use-case behavior exercised through a supported application boundary.
* Authorization, validation, persistence, events, background work, and other observable effects where applicable.
* Acceptance criteria requiring verification across more than one isolated unit.

This contract does not define:

* Business requirements
* Feature-specific behavior
* Architectural decisions
* Project-wide conventions
* Technology selection

---

## Responsibility

The artifact is responsible for:

* Verifying externally observable outcomes defined by Specifications and applicable Artifact Contracts.
* Detecting integration failures between collaborating application boundaries and artifacts.
* Verifying successful, rejected, boundary, and repeated-execution scenarios where applicable.
* Verifying required side effects and the absence of forbidden side effects.
* Providing deterministic evidence that the tested behavior satisfies its declared contract.

The artifact must not take responsibility for:

* Defining business behavior that is absent from authoritative project knowledge.
* Re-testing behavior owned entirely by the selected technology stack or external providers.
* Verifying private implementation details that are not part of the observable contract.
* Replacing focused lower-level tests for isolated deterministic logic.

---

## Inputs

| Input                           | Description                                                                           | Required    |
| ------------------------------- | ------------------------------------------------------------------------------------- | ----------- |
| Test scenario                   | Initial state, actor or execution context, operation, and expected observable outcome | Yes         |
| Controlled test data            | Minimal deterministic state required by the scenario                                  | Yes         |
| Controlled boundary substitutes | Deterministic substitutes for external or nondeterministic dependencies               | As required |

---

## Outputs

| Output              | Description                                                                    |
| ------------------- | ------------------------------------------------------------------------------ |
| Verification result | Deterministic pass or failure representing whether the declared behavior holds |
| Diagnostic evidence | Failure context sufficient to identify the violated observable behavior        |

---

## Behavior

### Observable Success

**Condition**

A valid operation is performed through a supported application boundary under valid preconditions.

**Expected Behavior**

The test verifies the declared result, authoritative state changes, required side effects, and relevant observable interactions without depending on private implementation steps.

### Observable Rejection

**Condition**

Authorization, validation, lifecycle state, conflict, or another declared precondition rejects the operation.

**Expected Behavior**

The test verifies the declared rejection outcome, absence of unauthorized or invalid authoritative state changes, and absence of forbidden downstream effects.

### Repeated Execution

**Condition**

The owning Specification defines behavior for retries, repeated delivery, duplicate submission, or another repeated logical operation.

**Expected Behavior**

The test verifies the declared repeat-execution semantics, including idempotency or duplicate-effect prevention where required.

---

## Test State Ownership

A Feature Test owns only the controlled state required to establish and observe its scenario.

Authoritative application behavior must be exercised through supported application boundaries rather than reproduced inside the test.

The test must not duplicate production business logic in fixtures, helpers, assertions, or test-only implementations.

---

## Constraints

The artifact must:

* Use a clear scenario structure such as Arrange–Act–Assert or an equivalent form.
* Keep each test independent, deterministic, and free from execution-order dependencies.
* Establish only the state necessary for the behavior being verified.
* Assert observable outcomes rather than private implementation steps.
* Assert critical forbidden side effects when their absence is part of the expected behavior.
* Control external, slow, nondeterministic, time-dependent, or environment-dependent capabilities where required for deterministic execution.
* Exercise realistic application boundaries appropriate to the behavior under test.
* Preserve the distinction between test setup, execution, and verification.

The artifact must not:

* Depend on live external services or public networks.
* Require production credentials or protected production data.
* Depend on uncontrolled wall-clock timing, randomness, shared mutable state, arbitrary delays, or test execution order.
* Use retries to hide nondeterministic or flaky behavior.
* Assert private methods, incidental internal calls, or implementation-specific structure unless explicitly defined as part of the contract.
* Reimplement the behavior under test inside test helpers or fixtures.

---

## Dependencies

### Allowed Dependencies

* Project-approved testing capabilities
* Controlled fixtures and test-data builders
* Deterministic fakes, stubs, or test adapters
* Public application boundaries
* Observable persistence, event, and background-work interfaces where applicable

### Restricted Dependencies

* Live external services
* Production credentials
* Protected production data
* Private implementation details
* Undocumented internal state
* Uncontrolled nondeterministic dependencies

---

## Error Handling

### Nondeterministic Test

**Condition**

The test depends on uncontrolled time, randomness, concurrency, external output, shared state, or environment behavior.

**Expected Result**

The nondeterministic dependency is replaced or controlled through an appropriate test boundary.

Retries must not be introduced merely to hide flaky behavior.

### Unverifiable Requirement

**Condition**

The expected behavior cannot be derived from an authoritative Specification, applicable Artifact Contract, or other approved project knowledge.

**Expected Result**

The test must not invent the missing behavior.

The knowledge gap must be identified and resolved at the appropriate authoritative source before the behavior is encoded as a normative test expectation.

---

## Integration

### Specification Verification

**Participant**

Owning Specification and applicable Artifact Contracts

**Interaction**

Acceptance criteria and artifact obligations are translated into executable observable scenarios.

**Constraints**

* Every applicable acceptance criterion must be covered by an appropriate verification layer or explicitly assigned elsewhere.
* Tests must preserve the terminology and behavioral intent of the authoritative project knowledge.
* Feature Tests must not silently introduce requirements absent from authoritative project knowledge.

### Application Boundary

**Participant**

Supported application entry boundary

**Interaction**

The test establishes controlled preconditions, performs the declared operation through the supported boundary, and observes the resulting application behavior.

**Constraints**

* The test must not bypass the behavior it is intended to verify merely to simplify setup or assertions.

---

## Verification

A valid Feature Test must satisfy the following:

* It fails when its required observable behavior is deliberately violated.
* It passes deterministically when the declared behavior is satisfied.
* It can execute independently of other Feature Tests.
* It does not depend on execution order.
* External and nondeterministic dependencies are controlled.
* No production credentials or protected production data are required.
* Failure output identifies the violated observable behavior with useful diagnostic context.
* Assertions remain focused on declared behavior rather than incidental implementation details.
* Test expectations are traceable to authoritative project knowledge.

---

## Examples

### Valid Example

```text
Given an authorized actor and valid application state,
when the same state-changing command is submitted twice,
then the authoritative outcome matches the declared repeat-execution rule
and no unintended duplicate downstream effect is produced.
```

### Invalid Example

```text
Invoke a private implementation method,
assert that specific internal methods were called,
reproduce business logic inside the test,
and never verify the observable application outcome.
```

---

## Related Architecture

* Application Boundaries
* Module Boundaries
* External System Boundaries

---

## Related Conventions

* Testing
* Error Handling
* Documentation

---

## Notes

Feature Tests verify observable behavior across collaborating application artifacts.

Focused lower-level tests remain appropriate for isolated deterministic rules and implementation units.

Feature Tests do not define requirements. Their expectations must remain traceable to authoritative project knowledge.

Technology-specific testing tools, runners, assertion libraries, and test-environment mechanisms remain defined by the project Stack.
