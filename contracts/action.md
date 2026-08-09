# Artifact Contract — Action

## Purpose

Define the contract for application Actions that execute one explicit use case and coordinate domain and infrastructure boundaries.

---

## Artifact Type

**Name**

Application Action

**Description**

An application-layer artifact that executes one explicit use case, coordinates application boundaries, and returns a declared result.

---

## Scope

This contract applies to:

* State-changing application use cases.
* Synchronous or asynchronous application orchestration invoked through an authorized application entry point.

This contract does not define:

* Business requirements
* Feature-specific behavior
* Architectural decisions
* Project-wide conventions
* Technology selection

---

## Responsibility

The artifact is responsible for:

* Orchestrating one named use case.
* Enforcing authorization and applicable business invariants at the appropriate application boundary.
* Coordinating persistence, transactions, application ports, events, and background work.
* Returning a declared successful result or an explicit application failure.

The artifact must not take responsibility for:

* Parsing transport-specific requests or formatting transport-specific responses.
* Implementing provider-specific behavior.
* Coordinating multiple unrelated use cases.
* Owning presentation behavior.

---

## Inputs

| Input                    | Description                                                      | Required                   |
| ------------------------ | ---------------------------------------------------------------- | -------------------------- |
| Typed input              | Validated use-case input                                         | Yes                        |
| Authorization context    | Identity or execution context under which the Action runs        | When authorization applies |
| Application capabilities | Declared application-owned capabilities required by the use case | As required                |

---

## Outputs

| Output              | Description                                                                                       |
| ------------------- | ------------------------------------------------------------------------------------------------- |
| Typed result        | Explicit successful outcome returned to the caller                                                |
| Application failure | Stable application-level failure representing an expected rejected outcome                        |
| Side effects        | Persisted state, emitted events, or scheduled background work explicitly required by the use case |

---

## Behavior

### Successful Execution

**Condition**

Input, authorization, current state, and required dependencies satisfy the owning specification.

**Expected Behavior**

The Action executes the use case exactly once, commits mutually dependent state atomically where required, performs the defined side effects, and returns its declared result.

### Repeated Execution

**Condition**

The same logical command is retried or submitted more than once.

**Expected Behavior**

The Action follows the specification's idempotency rules and must not create duplicate authoritative state or duplicate downstream processing.

---

## Constraints

The artifact must:

* Have one public entry method.
* Represent one explicit application use case.
* Use strongly typed inputs and outputs.
* Define explicit transaction boundaries where consistency requires them.
* Depend only on application-owned interfaces for external capabilities.
* Schedule background work only after the state it depends on is durably committed.

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
* Provider SDKs
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

* Business rules must not be duplicated outside the Action.
* Delivery-specific concerns remain outside the Action.

---

## Verification

* Tests cover successful execution, expected rejection, authorization, dependency failure, and idempotency where applicable.
* Dependencies can be substituted with test doubles through application-owned interfaces.
* The Action contains no presentation-specific behavior.
* The Action contains no provider-specific behavior.
* Public inputs and outputs are type-checked where supported by the project's technology stack.

---

## Examples

### Valid Example

```text
ConfirmFileUploadAction

Receives validated upload input.

Verifies authorization and current state.

Registers one authoritative file record atomically.

Schedules downstream processing after commit.

Returns a declared successful result.
```

### Invalid Example

```text
A delivery endpoint validates input, performs authorization,
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
* Jobs and Idempotency
* Ports and Adapters
* Error Handling

---

## Notes

An Action implements a use case defined elsewhere.

Business behavior remains defined by the relevant Specification and Domain knowledge.

Technology-specific implementation details remain defined by the project Stack.
