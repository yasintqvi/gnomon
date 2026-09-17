# SPEC-[ID] — [Use Case Name]

## Purpose

Define the business requirements and expected behavior of the `[Use Case Name]` use case.

This specification describes what the system must achieve without prescribing implementation details.

This document's lifecycle and approval state (Draft or Approved) are governed by [`SPECIFICATION_LIFECYCLE.md`](SPECIFICATION_LIFECYCLE.md) and are never recorded as a field inside this document.

A Draft Specification is brought to `READY_FOR_APPROVAL` through [`workflows/specification-definition.md`](../workflows/specification-definition.md), which does not require every section below to be filled — only what this use case's behavior actually needs.

The `[ID]` above is assigned deterministically by [`workflows/specification-discovery.md`](../workflows/specification-discovery.md)'s Specification Identity rule, whether this file was created through Discovery or directly by a Human — never chosen freely and never `000`, which is reserved for this template.

---

## Use Case

**Name**

[Use case name]

**Description**

[Describe the business capability provided by this use case.]

**Actor**

[Primary actor]

**Goal**

[Describe what the actor wants to achieve.]

---

## Scope

### In Scope

* [Business behavior]
* [Business behavior]

### Out of Scope

* [Excluded behavior]
* [Excluded behavior]

---

## Preconditions

The following conditions must be satisfied before the use case can begin.

* [Precondition]
* [Precondition]

---

## Main Flow

1. [Business step]
2. [Business step]
3. [Business step]
4. [Expected result]

---

## Alternative Flows

### [Alternative Flow]

**Condition**

[Describe when this flow applies.]

**Flow**

1. [Business step]
2. [Business step]

**Result**

[Expected business result.]

---

## Business Rules

* [Business rule]
* [Business rule]
* [Business rule]

---

## Constraints

* [Business constraint]
* [Business constraint]

---

## Inputs

| Input   | Description        | Required |
| ------- | ------------------ | -------- |
| [Input] | [Business meaning] | Yes/No   |

Only describe inputs relevant to the business use case.

Do not define implementation-specific request structures.

---

## Outputs

### Success

[Describe the expected business outcome.]

### Failure

[Describe possible business-level failure outcomes.]

---

## Acceptance Criteria

### [Criterion]

**Given**

[Initial business condition]

**When**

[Business action]

**Then**

[Expected business result]

---

### [Criterion]

**Given**

[Initial business condition]

**When**

[Business action]

**Then**

[Expected business result]

---

## Domain References

List relevant domain concepts or rules used by this specification.

* [Domain concept]
* [Domain rule]

---

## Related Decisions

List relevant ADRs when an existing architectural decision affects the interpretation or implementation of this use case.

* [ADR]

---

## Dependencies

List other Specifications this one depends on, by identity only — never by title or prose. A dependency belongs here only when this use case's own behavior would become incomplete, incorrect, or unverifiable without the listed Specification; see [`SPECIFICATION_DEPENDENCIES.md`](SPECIFICATION_DEPENDENCIES.md) for the governing test. Do not list a Specification here merely because it seemed useful to build first.

* [SPEC-ID]
* [SPEC-ID]

When there are none:

```
## Dependencies

None.
```

---

## Notes

[Optional clarification that belongs specifically to this use case.]
