# Artifact Contract — Interactive Presentation

## Purpose

Define the contract for presenting information and available interactions from declared state without owning authoritative business behavior.

---

## Artifact Type

**Name**

Interactive Presentation

**Description**

A responsibility, represented in any architecture-approved form, that communicates state and exposes user interaction through an accessible presentation boundary.

This contract does not require a particular artifact name, composition model, runtime location, or implementation technology.

---

## Scope

This contract applies to interactive presentation of approved information, state, and user intent.

This contract does not define business requirements, feature behavior, architectural representation, project conventions, design decisions, or technology selection.

---

## Responsibility

The artifact is responsible for:

* Presenting declared information and state for one bounded user-facing concern.
* Exposing only declared interactions.
* Communicating applicable presentation and interaction states required by authoritative project knowledge.
* Preserving applicable obligations defined by Design Knowledge.

The artifact must not:

* Make authoritative authorization or business decisions.
* Own unrelated workflows.
* Directly depend on undeclared providers or boundaries.
* Replace authoritative state with local presentation state.
* Define missing presentation or interaction behavior.

---

## Inputs

| Input                    | Description                                                                                     | Required    |
| ------------------------ | ----------------------------------------------------------------------------------------------- | ----------- |
| Presentation data        | Approved information required for presentation                                                  | As declared |
| Presentation state       | Current interaction, validation, availability, progress, or other applicable presentation state | As required |
| Interaction availability | Declared availability or constraints affecting exposed interactions                             | As required |

---

## Outputs

| Output               | Description                                                        |
| -------------------- | ------------------------------------------------------------------ |
| Presented experience | Accessible representation of declared information and state        |
| User intent          | Explicit interaction communicated through the owning boundary      |
| Ephemeral state      | Local presentation state that does not replace authoritative state |

---

## Behavior

### Present Known State

**Condition**

Declared inputs represent a supported state.

**Expected Behavior**

Present accurate information and expose only interactions permitted by supplied state and authoritative project knowledge.

### Pending or Failed Interaction

**Condition**

An interaction is pending, rejected, or fails.

**Expected Behavior**

Prevent unintended duplicate intent where required, preserve relevant context where applicable, communicate safe feedback accessibly, and expose approved recovery behavior where applicable.

### State Ownership

**Condition**

The presentation maintains local or temporary state.

**Expected Behavior**

Presentation may own ephemeral interaction state only. Business, workflow, and other authoritative state remain with their architecture-defined owners.

When temporary presentation state anticipates an authoritative outcome, it must be reconciled with the resulting authoritative state.

---

## Constraints

The artifact must:

* Use declared data, state, intent, and boundaries.
* Follow applicable Design Knowledge and other authoritative presentation requirements.
* Preserve approved state ownership and interaction boundaries.

The artifact must not:

* Infer authorization from presentation or visibility.
* Treat local validation as authoritative unless explicitly owned by the presentation boundary.
* Duplicate authoritative business rules.
* Hide consequential effects inside generic presentation behavior.
* Expose protected implementation or failure details.
* Introduce presentation or interaction decisions absent from authoritative project knowledge.

---

## Dependencies

### Allowed Dependencies

* Presentation dependencies permitted by applicable Architecture, Stack, Conventions, Design Knowledge, and other authoritative project knowledge.

### Restricted Dependencies

* Protected authoritative business-state internals.
* Provider implementations where an approved boundary exists.
* Undeclared cross-context mutable state.
* Dependencies prohibited by Architecture, Stack, Conventions, or applicable ADRs.

---

## Error Handling

### Validation Failure

**Condition**

The owning boundary returns validation feedback.

**Expected Result**

Communicate approved safe feedback in the relevant context, preserve applicable user state where required, and follow applicable Design Knowledge.

### Unexpected Failure

**Condition**

An interaction fails without a more specific approved outcome.

**Expected Result**

Present a safe outcome without protected-detail leakage and expose approved recovery behavior where applicable.

---

## Integration

### Owning Interaction Boundary

**Participant**

The architecture-defined owner of surrounding interaction and authoritative state.

**Interaction**

The participant supplies declared inputs, receives user intent, coordinates applicable navigation or execution, and provides resulting authoritative state.

**Constraints**

* Integration preserves state ownership and declared interaction boundaries.
* Presentation must not assume responsibilities owned by the integrating boundary.

---

## Verification

* Evidence demonstrates accurate presentation of supported states and declared information.
* Evidence demonstrates that exposed interactions correspond to approved availability and produce declared user intent.
* Evidence demonstrates pending, rejection, failure, and recovery behavior where applicable.
* Evidence demonstrates applicable state-ownership and reconciliation behavior.
* Evidence demonstrates applicable accessibility and Design Knowledge obligations using evidence appropriate to their observable behavior.
* Evidence demonstrates that presentation does not make authoritative business or authorization decisions.
* Evidence demonstrates that protected implementation, provider, and failure details do not cross the presentation boundary.

---

## Examples

### Valid Example

```
Declared state is presented through an approved presentation boundary.

Available interaction communicates explicit user intent.

Temporary presentation state remains non-authoritative.

Resulting authoritative state returns through its owning boundary.

Applicable presentation and interaction behavior follows Design Knowledge.
```

### Invalid Example

```
Presentation calls an undeclared provider directly,
decides authorization and business outcomes,
treats hidden controls as access control,
and replaces authoritative state with local state.
```

---

## Related Architecture

* Presentation Boundaries
* Interaction and State Ownership
* Module or Product Context Boundaries

---

## Related Conventions

* Accessibility
* Localization and Formatting
* Interaction and State
* Testing

---

## Notes

Design Knowledge defines approved experience, presentation, and interaction behavior.

Architecture, Stack, Conventions, ADRs, and other authoritative project knowledge determine how this responsibility is represented, composed, integrated, and implemented.
