# Artifact Contract — Component

## Purpose

Define the contract for user-interface components that present state and emit user intent without owning application workflows.

---

## Artifact Type

**Name**

UI Component

**Description**

A reusable user-interface artifact with an explicit visual responsibility, typed inputs, emitted user intents, and accessible interaction states.

---

## Scope

This contract applies to:

* Reusable or feature-local user-interface components.
* Components that display application state or collect user interaction.

This contract does not define:

* Business requirements
* Feature-specific behavior
* Architectural decisions
* Project-wide conventions
* Technology selection

---

## Responsibility

The artifact is responsible for:

* Representing one coherent presentation responsibility from declared input and presentation state.
* Emitting explicit user intents and presenting loading, success, empty, and error states.
* Preserving accessibility, keyboard operation, and responsive layout.

The artifact must not take responsibility for:

* Enforcing server-authoritative authorization or business invariants.
* Calling provider SDKs or owning multi-step application orchestration.

---

## Inputs

| Input | Description | Required |
| --- | --- | --- |
| Typed input | Data and configuration required to render | As declared |
| Presentation or interaction state | Current interaction and validation state | As required |
| Composition or content extension points | Explicit extension points for caller-provided content | No |

---

## Outputs

| Output | Description |
| --- | --- |
| Rendered interface | Accessible UI representing the current state |
| Typed events | User intent emitted to the owning page or parent |
| Local interaction state | Ephemeral presentation state that does not replace server state |

---

## Behavior

### Render Known State

**Condition**

The component receives valid typed input.

**Expected Behavior**

It renders deterministic content and exposes only interactions valid for the supplied presentation state.

### Pending or Failed Interaction

**Condition**

An owned form submission is pending or returns validation/application errors.

**Expected Behavior**

The component prevents accidental duplicate actions where required, preserves useful user input, associates errors with controls, and offers a clear recovery path.

---

### State Ownership

The component owns only transient presentation state.

Application state, business state, and workflow state remain owned by the application.

---

## Constraints

The artifact must:

* Use strongly typed inputs, outputs, and interaction data where supported by the project's technology stack.
* Provide semantic labels, focus behavior, keyboard access, and non-color-only status cues.
* Make disabled and pending states explicit.
* Keep user-facing text in the project's localization resources.

The artifact must not:

* Infer authorization from hidden controls or treat client validation as authoritative.
* Duplicate domain rules already enforced by the server.
* Hide network or workflow side effects inside generic presentational components.
* Display raw provider, stack trace, or internal storage errors.

---

## Dependencies

### Allowed Dependencies

* Presentation framework abstractions approved by the project
* Typed presentation interfaces
* Shared UI primitives
* Localization mechanism
* Feature-local child components
* Presentation utilities

### Restricted Dependencies

* Domain models and Provider-specific implementations
* Cross-feature global mutable state without an approved application need

---

## Error Handling

### Validation Error

**Condition**

The server returns field or form-level validation errors.

**Expected Result**

Show localization-ready messages near the relevant controls, preserve correct input, and move focus or announce the error accessibly.

### Unexpected Failure

**Condition**

An operation fails without a recoverable field-level error.

**Expected Result**

Present a safe general message and an appropriate retry or exit path without exposing internal details.

---

## Integration

### Owning Screen

**Participant**

Owning interface

**Interaction**

The owning interface supplies presentation data,
handles navigation or application interaction,
and reacts to emitted user intent.

**Constraints**

* Server responses remain authoritative; local optimistic state must reconcile with them.

---

## Verification

* Component tests cover rendering, emitted intent, pending state, validation errors, and disabled behavior.
* Accessibility checks cover labels, keyboard navigation, focus, and status announcements.
* Public inputs and outputs are validated according to the capabilities of the selected project Stack.
* Tests do not depend on implementation-private DOM structure when accessible roles or labels are available.

---

## Examples

### Valid Example

```text
A form component receives declared input,
renders validation state,
emits user intent,
prevents duplicate interaction while pending,
and never executes the application use case directly.
```

### Invalid Example

```text
A generic button component
calls an external service directly,
performs authorization,
updates application state,
and decides business outcomes.
```

---

## Related Architecture

* Application Boundary
* Application Orchestration
* Module Boundaries


---

## Related Conventions

* File Structure
* Naming
* Formatting
* Testing
* Localization rules

---

## Notes

Owning interface artifacts may coordinate feature-specific interaction.
Reusable child components should remain narrowly presentation-focused.
