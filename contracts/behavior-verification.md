# Artifact Contract — Behavioral Verification

## Purpose

Define the contract for executable verification of observable behavior relevant to an approved scenario.

---

## Artifact Type

**Name**

Behavioral Verification

**Description**

An executable verification responsibility that establishes appropriate preconditions, exercises approved behavior, and evaluates observable outcomes.

This contract does not require a particular test category, level, artifact name, execution model, or implementation technology.

---

## Scope

This contract applies when approved behavior or acceptance criteria require executable evidence of observable behavior.

It does not define requirements, expected behavior, architectural representation, testing strategy, project conventions, or technology selection.

---

## Responsibility

The artifact is responsible for:

* Verifying one bounded behavior or coherent scenario traceable to authoritative project knowledge.
* Exercising the behavior through boundaries appropriate to the declared scenario.
* Evaluating observable outcomes and applicable required or forbidden effects.
* Producing reliable, reproducible, and diagnostic evidence.

The artifact must not:

* Invent expected behavior.
* Reproduce production behavior in verification logic.
* Assert incidental implementation details absent from approved obligations.
* Duplicate verification already owned at a more appropriate level without a justified purpose.
* Impose a testing strategy or verification level absent from authoritative project knowledge.

---

## Inputs

| Input              | Description                                                                   | Required    |
| ------------------ | ----------------------------------------------------------------------------- | ----------- |
| Scenario           | Preconditions, behavior, and expected observable outcome                      | Yes         |
| Controlled state   | State required to establish and observe the scenario reliably                 | As required |
| Execution controls | Controls required to make execution and observation reliable and reproducible | As required |

---

## Outputs

| Output              | Description                                                          |
| ------------------- | -------------------------------------------------------------------- |
| Verification result | Observable pass or failure for the declared behavior                 |
| Diagnostic evidence | Evidence identifying the satisfied or violated observable obligation |

---

## Behavior

### Observable Success

**Condition**

The declared behavior occurs under applicable approved preconditions.

**Expected Behavior**

Verify the declared outcome, applicable authoritative state changes, required effects, and relevant observable interactions without depending on incidental implementation steps.

### Observable Rejection

**Condition**

An approved precondition rejects the behavior.

**Expected Behavior**

Verify the declared rejection and absence of unauthorized, invalid, or otherwise forbidden effects.

### Repeated Execution

**Condition**

Authoritative project knowledge defines retry, duplicate, concurrent, or repeated-execution behavior.

**Expected Behavior**

Verify the declared repetition semantics, including idempotency, duplicate-effect prevention, or other applicable guarantees.

---

## Test State Ownership

The artifact owns only the state and controls required to establish and observe its declared scenario.

It must exercise the approved behavior rather than reproduce that behavior in setup, helpers, assertions, substitutes, or other verification support.

---

## Constraints

The artifact must:

* Be bounded, focused, reproducible, and traceable to approved behavior.
* Keep scenario establishment, behavior execution, and outcome evaluation distinguishable.
* Assert observable obligations rather than incidental implementation structure.
* Control or account for nondeterminism sufficiently to produce reliable evidence.
* Use an execution level capable of actually exercising the behavior being verified.

The artifact must not:

* Depend on uncontrolled state or execution ordering.
* Use retries merely to hide unexplained flakiness.
* Use production credentials or protected production data unless explicitly authorized by applicable project knowledge.
* Bypass the behavior or boundary being verified.
* Encode expectations that cannot be traced to authoritative project knowledge.

---

## Dependencies

### Allowed Dependencies

* Dependencies and verification capabilities permitted by applicable Architecture, Stack, Conventions, testing strategy, and other authoritative project knowledge.
* Controlled or real execution dependencies appropriate to the selected verification level.

### Restricted Dependencies

* Uncontrolled dependencies that make evidence unreliable.
* Protected production capabilities or data without explicit authorization.
* Undocumented internals and incidental implementation details not required by the declared obligation.
* Dependencies prohibited by Architecture, Stack, Conventions, or applicable ADRs.

---

## Error Handling

### Unreliable Verification

**Condition**

Results materially depend on uncontrolled time, randomness, concurrency, external behavior, shared state, environment conditions, or other nondeterminism.

**Expected Result**

Control, isolate, or explicitly account for the source of nondeterminism using an approved mechanism. Do not hide unreliable behavior through unexplained retries.

### Undefined Expected Behavior

**Condition**

The expected outcome cannot be traced to authoritative project knowledge.

**Expected Result**

Do not encode the expectation. Report that the expected behavior is undefined and defer resolution to the owning knowledge process.

---

## Integration

### Requirement Verification

**Participant**

Owning Specifications, project knowledge, and applicable Artifact Contracts.

**Interaction**

Translate approved observable obligations into executable scenarios at the verification level selected by the applicable testing strategy.

**Constraints**

* Verification introduces no silent requirements.
* The selected scenario and verification level must remain traceable to approved obligations.

### Execution Boundary

**Participant**

The approved boundary appropriate to the behavior.

**Interaction**

Establish required preconditions, exercise the behavior, and observe its outcomes without bypassing the behavior under verification.

---

## Verification

* Evidence demonstrates that the artifact exercises the declared behavior through an appropriate boundary.
* Assertions distinguish declared compliant and non-compliant observable outcomes.
* Evidence demonstrates reliable and reproducible execution within applicable constraints.
* Applicable nondeterministic dependencies are controlled or explicitly accounted for.
* Failure evidence identifies the violated observable obligation with sufficient diagnostic context.
* Assertions remain traceable to authoritative behavior rather than incidental implementation details.
* Verification support does not reproduce the production behavior being verified.

---

## Examples

### Valid Example

```
Given approved preconditions,
the declared behavior is exercised through an appropriate boundary.

Observable outcomes and required effects are evaluated against
authoritative project knowledge.

Necessary execution dependencies are controlled or accounted for
sufficiently to produce reliable evidence.
```

### Invalid Example

```
Verification reproduces production logic,
depends on uncontrolled behavior,
asserts private implementation details,
and never meaningfully observes the declared outcome.
```

---

## Related Architecture

* Execution and Observation Boundaries
* Module or Subsystem Boundaries
* External System Boundaries

---

## Related Conventions

* Testing
* Error Handling
* Diagnostics
* Documentation

---

## Notes

Authoritative project knowledge defines the expected behavior.

The project testing strategy determines the appropriate verification level and environment.

Architecture, Stack, Conventions, ADRs, and other authoritative project knowledge determine the representation, tools, dependencies, and execution mechanics used to produce the evidence.
