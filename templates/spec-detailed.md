# SPEC-[ID] — [Use Case Name]

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

One observable behavior each: the decided behaviors and the boundaries a plausible implementation could get wrong, usually about ten. Not the interface the request already fixes, and not a repeat of another criterion.

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

## Decisions

Questions whose answer changes what gets built, each once, with the answer the user gave. A decision already recorded elsewhere is referred to by its source, not repeated.

* [Question] — [Answer]

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

List other Specifications this one depends on, by identity only, and only when this use case's behavior would be incomplete, incorrect, or unverifiable without them.

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
