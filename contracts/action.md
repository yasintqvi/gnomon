# Artifact Contract — Action

## Purpose

Define the contract for application Actions that execute one explicit state-changing use case and coordinate application boundaries.

This contract defines the canonical characteristics of an Action independently of any project, framework, or implementation technology.

---

## Artifact Type

**Name**

Application Action

**Description**

An application-layer artifact that executes one explicit state-changing use case, coordinates the required application boundaries, and produces a declared outcome.

---

## Scope

This contract applies to:

* State-changing application use cases.
* Application orchestration invoked through an authorized application entry point.

This contract does not define:

* Business requirements
* Feature-specific behavior
* Architectural decisions
* Project-wide conventions
* Technology selection

---

## Responsibility

The artifact is responsible for:

* Orchestrating one explicit application use case.
* Coordinating authorization checks where applicable.
* Ensuring business invariants are enforced by their owning boundaries.
* Coordinating persistence, transactions, application capabilities, events, and background work.
* Returning a declared successful result or an explicit application failure.

The artifact must not take responsibility for:

* Parsing transport-specific requests.
* Producing presentation-layer responses.
* Implementing provider-specific behavior.
* Coordinating multiple unrelated use cases.
* Owning presentation concerns.

---

## Inputs

| Input                    | Description                                                      | Required                   |
| ------------------------ | ---------------------------------------------------------------- | -------------------------- |
| Declared input           | Validated input required by the use case                         | Yes                        |
| Authorization context    | Identity or execution context under which the Action executes    | When authorization applies |
| Application capabilities | Declared application-owned capabilities required by the use case | As required                |

---

## Outputs

| Output              | Description                                                                                       |
| ------------------- | ------------------------------------------------------------------------------------------------- |
| Declared result     | Explicit successful outcome returned to the caller                                                |
| Application failure | Stable application-level failure representing an expected rejected outcome                        |
| Side effects        | Persisted state, emitted events, or scheduled background work explicitly required by the use case |

---

## Behavior

### Successful Execution

**Condition**

Input, authorization, current state, and required dependencies satisfy the owning Specification.

**Expected Behavior**

The Action executes the accepted use case, coordinates the required application boundaries, performs the defined side effects, preserves consistency where required, and returns its declared result.

### Repeated Execution

**Condition**

The same logical command is retried or submitted more than once.

**Expected Behavior**

The Action follows the idempotency requirements defined by the owning Specification and must not produce unintended duplicate authoritative effects.

---

## Constraints

The artifact must:

* Have one public entry method.
* Represent one explicit application use case.
* Use explicit declared inputs and outputs, with strong typing where supported by the project Stack.
* Define explicit transaction boundaries where consistency requires them.
* Depend only on application-owned interfaces for external capabilities.
* Schedule background work only after the state it depends on has been durably committed.

The artifact must not:

* Accept transport-specific request objects.
* Produce presentation-layer output.
* Return provider-specific objects.
* Persist transient transport data.
* Reach into another module's internal implementation or bypass declared interfaces.

---

## Dependencies

### Allowed Dependencies

* Domain entities
* Value objects
* Policies
* Repositories
* Application-owned interfaces
* Data objects
* Transactions
* Events
* Background work abstractions

### Restricted Dependencies

* Presentation-layer artifacts
* Provider-specific implementations
* Another module's internal implementation

---

## Error Handling

### Expected Rejection

**Condition**

Authorization, validation, conflict, or state preconditions reject the command.

**Expected Result**

The Action returns or throws a stable application-level failure without partial state changes or protected-data leakage.

### Dependency Failure

**Condition**

A required persistence operation or external capability fails.

**Expected Result**

The Action preserves transactional consistency, translates boundary failures where appropriate, and leaves retry behavior explicit.

---

## Integration

### Application Boundary

**Participant**

Authorized application entry point

**Interaction**

The participant maps incoming input into the Action's declared input, invokes the Action, and handles its declared result.

**Constraints**

* Business rules must not be duplicated at the delivery boundary.
* Delivery-specific concerns remain outside the Action.

---

## Verification

* Tests cover successful execution, expected rejection, authorization, dependency failure, and idempotency where applicable.
* Dependencies can be substituted with test doubles through application-owned interfaces.
* The Action contains no presentation-specific behavior.
* The Action contains no provider-specific behavior.
* Declared inputs and outputs are validated according to the capabilities of the selected project Stack.

---

## Examples

### Valid Example

```text
ConfirmFileUploadAction

Receives validated upload input.

Coordinates authorization.

Verifies current state.

Registers one authoritative file record.

Schedules downstream processing after commit.

Returns a declared successful result.
```

### Invalid Example

```text
A delivery endpoint validates input,
performs authorization,
updates multiple records,
calls an external provider directly,
schedules background work,
and constructs the response without an application Action.
```

---

## Related Architecture

* Application Boundary
* Application Orchestration
* Module Boundaries

---

## Related Conventions

* Action and Data Objects
* Background Work and Idempotency
* Ports and Adapters
* Error Handling

---

## Notes

An Action implements an application use case defined elsewhere.

Business behavior remains defined by the relevant Specification and Domain knowledge.

Technology-specific implementation details remain defined by the project Stack.

An Action coordinates behavior; it does not become the authoritative owner of business rules, presentation behavior, or transport concerns.
