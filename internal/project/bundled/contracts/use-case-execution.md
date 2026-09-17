# Artifact Contract — Use-Case Execution

## Purpose

Define the contract for the responsibility that executes one declared use case and coordinates its required boundaries and effects.

---

## Artifact Type

**Name**

Use-Case Execution

**Description**

A responsibility, represented by one artifact or an architecture-defined collaboration, that receives a declared request, executes one use case, and exposes its outcome.

This contract does not require a particular artifact name, type, layer, or implementation shape.

---

## Scope

This contract applies to:

* Execution of a declared use case that may read state, change state, or produce other observable effects.
* Synchronous or asynchronous coordination across boundaries required by that use case.

This contract does not define business requirements, feature behavior, architectural representation, project conventions, or technology selection.

---

## Responsibility

The artifact is responsible for:

* Executing one coherent use case defined by authoritative project knowledge.
* Coordinating applicable validation, authorization, business rules, and state preconditions through their approved owning boundaries.
* Coordinating required dependencies and effects while preserving specified consistency guarantees.
* Exposing the declared success or failure outcome.

The artifact must not define missing behavior, reimplement responsibilities owned by collaborating boundaries, combine unrelated use cases, couple execution to undeclared entry or provider details, or bypass project boundaries.

---

## Inputs

| Input                 | Description                                                                | Required    |
| --------------------- | -------------------------------------------------------------------------- | ----------- |
| Declared request      | Data or signal required to initiate the use case                           | Yes         |
| Execution context     | Identity, authorization, tenancy, correlation, or other applicable context | As required |
| Required capabilities | Declared collaborators or boundaries needed for execution                  | As required |

---

## Outputs

| Output             | Description                                                               |
| ------------------ | ------------------------------------------------------------------------- |
| Declared result    | Successful outcome defined by authoritative project knowledge             |
| Declared failure   | Stable expected rejection or failure without protected-detail leakage     |
| Observable effects | State changes, emitted signals, or deferred work required by the use case |

---

## Behavior

### Successful Execution

**Condition**

The request, context, state, and required capabilities satisfy approved preconditions.

**Expected Behavior**

Execution produces the declared result and effects while preserving applicable consistency, ordering, and boundary guarantees.

### Repeated Execution

**Condition**

The same logical request is retried, repeated, or delivered more than once.

**Expected Behavior**

Execution follows approved duplicate-handling rules. When idempotency is required, repetition produces no duplicate authoritative state or downstream effects.

---

## Constraints

The artifact must:

* Represent one coherent execution responsibility using declared inputs, outcomes, dependencies, and boundaries.
* Preserve required atomicity, consistency, ordering, and durability.
* Coordinate effects according to approved consistency and ordering rules.

The artifact must not:

* Depend on undeclared entry or presentation representations.
* Leak provider-specific results beyond their owning boundary.
* Persist transient boundary data without authority.
* Bypass protected dependencies or approved project boundaries.

---

## Dependencies

### Allowed Dependencies

* Dependencies permitted by approved Architecture, Stack, Conventions, and applicable boundaries.

### Restricted Dependencies

* Dependencies that bypass approved boundaries.
* Provider implementations where an owned abstraction boundary is required.
* Protected internals of another module or subsystem.
* Dependencies prohibited by Architecture, Stack, Conventions, or applicable ADRs.

---

## Error Handling

### Expected Rejection

**Condition**

Validation, authorization, conflict, or state preconditions reject execution.

**Expected Result**

Expose a declared failure without forbidden or partial effects or protected-information leakage.

### Dependency Failure

**Condition**

A required collaborator fails, times out, or becomes unavailable.

**Expected Result**

Preserve applicable consistency guarantees, prevent untracked partial completion, translate boundary failures where required, and leave retry or recovery behavior explicit.

---

## Integration

### Execution Boundary

**Participant**

The architecture-defined initiator.

**Interaction**

The participant supplies the request and context, triggers execution through an approved boundary, and handles the declared outcome.

**Constraints**

* Integration preserves declared inputs, outcomes, and safeguards.
* Entry and presentation concerns remain at their owning boundaries.
* Business rules are not inconsistently duplicated.

---

## Verification

* Evidence demonstrates successful execution and declared outcomes.
* Evidence demonstrates expected rejection and dependency-failure behavior where applicable.
* Evidence demonstrates repeated-execution behavior where applicable.
* Evidence demonstrates required and forbidden effects and applicable consistency guarantees.
* Preconditions are enforced through their approved owning boundaries.
* Outcomes expose no undeclared entry, presentation, or provider details.
* Applicable project boundaries are preserved.

---

## Examples

### Valid Example

```
A declared request enters through an approved boundary.

Applicable preconditions are enforced by their owners.

Required effects are coordinated under approved consistency rules.

The declared outcome is exposed without integration leakage.
```

### Invalid Example

```
Execution invents business decisions, bypasses protected boundaries,

allows untracked partial effects, and leaks provider-specific results.
```

---

## Related Architecture

* Use-Case or Operation Boundaries
* Consistency Boundaries
* Module or Subsystem Boundaries
* External Integration Boundaries

---

## Related Conventions

* Input and Outcome Representation
* Error Handling
* Idempotency and Retry
* Dependency Boundaries
* Testing

---

## Notes

Specifications and Domain knowledge define behavior.

ARCHITECTURE.md, STACK.md, CONVENTIONS.md, applicable ADRs, and other authoritative knowledge determine representation, placement, invocation, and implementation.
