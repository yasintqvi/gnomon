# Artifact Contract — Migration

## Purpose

Define the contract for database migration artifacts that evolve the application's persistent schema safely, explicitly, and reproducibly.

---

## Artifact Type

**Name**

Migration

**Description**

A version-controlled schema change that evolves the application's persistent data model.

---

## Scope

This contract applies to:

* Creating or modifying persistent schema structures.
* Schema evolution required by application features.
* Schema compatibility across application versions.

This contract does not define:

* Business requirements
* Feature-specific behavior
* Architectural decisions
* Project-wide conventions
* Technology selection

---

## Responsibility

The artifact is responsible for:

* Applying one coherent schema change deterministically.
* Preserving structural integrity through explicit constraints.
* Maintaining compatibility with the application's deployment strategy.
* Providing a safe rollback whenever the change is technically reversible.

The artifact must not take responsibility for:

* Executing business workflows.
* Executing application use cases.
* Performing hidden destructive data transformations.
* Coordinating application behavior outside schema evolution.

---

## Inputs

| Input                  | Description                                                   | Required |
| ---------------------- | ------------------------------------------------------------- | -------- |
| Existing schema state  | Schema produced by previously applied migrations              | Yes      |
| Approved data model    | Required entities, attributes, relationships, and constraints | Yes      |
| Deployment constraints | Compatibility or rollout limitations affecting the migration  | No       |

---

## Outputs

| Output             | Description                                   |
| ------------------ | --------------------------------------------- |
| Updated schema     | Persistent schema after successful migration  |
| Rollback operation | Reverse transition when rollback is supported |

---

## Behavior

### Apply

**Condition**

The migration executes against the expected preceding schema.

**Expected Behavior**

The migration applies the complete schema change or fails without leaving the schema in an inconsistent state.

### Roll Back

**Condition**

The migration is reversed in a supported environment.

**Expected Behavior**

Only structures introduced by that migration are removed. Irreversible changes must be explicitly documented and handled through an approved deployment strategy.

---

## Constraints

The artifact must:

* Declare schema structures explicitly.
* Define constraints, relationships, defaults, and indexes explicitly where applicable.
* Consider existing production data, compatibility, and deployment safety.
* Use consistent temporal data representations across the project.

The artifact must not:

* Depend on application-layer artifacts.
* Depend on presentation-layer artifacts.
* Invoke external systems.
* Execute application behavior.
* Silently delete or rewrite material data without an approved migration strategy.
* Store large binary objects inside the relational schema unless explicitly required by the approved data model.

---

## Dependencies

### Allowed Dependencies

* Project-approved schema definition facilities
* Database-native schema capabilities required for structural definition

### Restricted Dependencies

* Application-layer artifacts
* Presentation-layer artifacts
* External services
* Infrastructure capabilities unrelated to schema evolution

---

## Error Handling

### Invalid Schema State

**Condition**

The expected prior schema or required database capability is unavailable.

**Expected Result**

The migration fails explicitly with actionable diagnostic information. No partial schema change is considered successful.

---

## Integration

### Persistence Layer

**Participant**

Persistent storage layer

**Interaction**

Application persistence artifacts rely on the schema only after the migration has been successfully applied.

**Constraints**

* Deployment order must preserve compatibility between the application and the persistent schema.

---

## Verification

* The migration applies successfully from the expected preceding schema.
* Rollback succeeds whenever declared reversible.
* Declared constraints enforce the intended structural invariants.
* A clean environment can execute the complete migration sequence successfully.

---

## Examples

### Valid Example

```text
Create a new persistent entity with explicit identifiers,
relationships, constraints, indexes, timestamps,
and any required uniqueness rules.
```

### Invalid Example

```text
Create generic fields with implicit semantics,
omit required constraints,
modify application data silently,
and rely on application logic alone to preserve integrity.
```

---

## Related Architecture

* Persistence Layer
* Data Model
* Deployment Strategy

---

## Related Conventions

* Naming
* Error Handling
* Documentation

---

## Notes

A Migration defines how the persistent schema evolves.

The owning Specification and approved data model define what the schema must represent.

Technology-specific migration mechanisms remain defined by the project Stack.
