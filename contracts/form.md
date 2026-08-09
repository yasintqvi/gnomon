# Artifact Contract — Form

## Purpose

Define the contract for input-validation artifacts that authorize, normalize, and validate incoming application commands before they reach the application layer.

---

## Artifact Type

**Name**

Input Validation

**Description**

A delivery-boundary artifact that converts untrusted input into validated data suitable for constructing the application's declared input model.

---

## Scope

This contract applies to:

* Input authorization.
* Structural validation.
* Safe normalization of incoming data.
* Delivery-boundary validation concerns.

This contract does not define:

* Business requirements
* Feature-specific behavior
* Architectural decisions
* Project-wide conventions
* Technology selection

---

## Responsibility

The artifact is responsible for:

* Rejecting malformed, missing, unsupported, or unauthorized input.
* Producing validated scalar and structured data with stable field names.
* Returning localization-ready validation failures through the delivery boundary.

The artifact must not take responsibility for:

* Persisting application state.
* Executing application use cases.
* Scheduling background work.
* Verifying facts that require external side effects or business workflow execution.

---

## Inputs

| Input                 | Description                                                   | Required    |
| --------------------- | ------------------------------------------------------------- | ----------- |
| Incoming request data | Untrusted values received through the delivery boundary       | Yes         |
| Authorization context | Current execution identity when authorization applies         | As required |
| Resource context      | Referenced resources required for authorization or validation | As required |

---

## Outputs

| Output                  | Description                                                             |
| ----------------------- | ----------------------------------------------------------------------- |
| Validated input         | Authorized, structurally valid data ready for application-layer mapping |
| Validation failures     | Stable field-level validation failures                                  |
| Authorization rejection | Safe rejection without protected-resource disclosure                    |

---

## Behavior

### Valid Input

**Condition**

Authorization succeeds and all declared structural rules pass.

**Expected Behavior**

The artifact exposes only validated values for mapping into the application's declared input model.

### Invalid or Unauthorized Input

**Condition**

Authorization fails or one or more validation rules fail.

**Expected Behavior**

Processing stops before the application use case executes, returning stable validation or authorization failures without side effects.

---

## Constraints

The artifact must:

* Treat all client-supplied data as untrusted.
* Keep accepted values, formats, units, and required fields explicit.
* Use the project's authorization capabilities rather than duplicating authorization logic.
* Keep machine-readable error identifiers separate from localized display text.

The artifact must not:

* Trust client-provided metadata as authoritative facts.
* Persist state.
* Invoke external capabilities.
* Schedule background work.
* Execute application use cases.
* Contain transactional business workflow.

---

## Dependencies

### Allowed Dependencies

* Project-approved validation facilities
* Project-approved authorization facilities
* Enumerations
* Value parsers
* Read-only lookup rules appropriate to the delivery boundary

### Restricted Dependencies

* Application use-case artifacts
* Provider-specific implementations
* Infrastructure services that modify application state

---

## Error Handling

### Malformed Input

**Condition**

Input cannot be normalized safely or violates a declared structural constraint.

**Expected Result**

Return field-addressable validation failures and execute no application use case.

### Unauthorized Resource

**Condition**

The actor is not permitted to access or modify the referenced resource.

**Expected Result**

Return the project's safe authorization response without exposing protected-resource details.

---

## Integration

### Delivery Boundary

**Participant**

Application entry point

**Interaction**

The entry point receives validated input and maps it explicitly into the application's declared input model.

**Constraints**

* Unvalidated values must never bypass the validation artifact.

---

## Verification

* Tests cover required, malformed, boundary, unsupported, and unauthorized inputs.
* Invalid requests never execute the application use case.
* Only validated values can be mapped into the application's declared input model.
* Validation failures are stable and localization-ready.

---

## Examples

### Valid Example

```text
Validate uploaded file metadata, declared category, size,
duration, and referenced resource identifiers.

Normalize accepted values.

Map the validated result into the application's declared input model.
```

### Invalid Example

```text
Accept unvalidated client metadata,
persist application state,
invoke external services,
schedule background work,
and execute the application use case directly from the validation artifact.
```

---

## Related Architecture

* Delivery Boundary
* Authorization Boundary
* Application Boundary

---

## Related Conventions

* Action and Data Objects
* Error Handling
* Naming
* Localization

---

## Notes

Input validation complements, but never replaces, application-level business rules or authoritative verification.

The application layer remains the authoritative owner of business behavior.
