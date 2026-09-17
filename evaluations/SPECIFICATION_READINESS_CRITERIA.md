# Specification Readiness Criteria

## Purpose

Define reusable criteria used by Specification Definition to determine whether a Specification's approval-relevant content is sufficiently and coherently defined to report `READY_FOR_APPROVAL`.

These criteria evaluate behavioral sufficiency, not template completeness. A criterion is satisfied when the Specification's actual behavioral needs are met, not when every template section is populated.

Purpose and Scope, Use Case Ownership, Main Flow, Inputs and Outputs, Acceptance Criteria, Domain and Knowledge Consistency, and Material Gaps apply to every Specification and may never resolve NOT APPLICABLE. Preconditions, Alternative and Failure Flows, and Business Rules and Constraints apply only when material to this specific use case and may legitimately resolve NOT APPLICABLE when they are not.

These criteria do not grant approval and do not determine SPEC lifecycle state; lifecycle state is governed solely by `SPECIFICATION_LIFECYCLE.md`.

---

# Purpose and Scope

- The use case's purpose is clear.
- In-scope and out-of-scope behavior are distinguishable.

---

# Use Case Ownership

- The primary actor is identified.
- The actor's goal is clear.

---

# Main Flow

- The primary behavior the use case must achieve is defined, without prescribing implementation details.

---

# Preconditions

- Preconditions material to correct behavior are stated.

---

# Alternative and Failure Flows

- Alternative or failure flows material to correct behavior are covered, unless the behavior is already owned by an existing Contract or Convention.

---

# Business Rules and Constraints

- Business rules and constraints this use case's correctness depends on are established — in this Specification, or by reference to Domain knowledge — not left implicit.

---

# Inputs and Outputs

- Business-meaningful inputs and outputs are identified, at business grain, not implementation grain.

---

# Acceptance Criteria

- Acceptance criteria cover the main flow's success outcome and its materially likely failure outcomes.

---

# Domain and Knowledge Consistency

- Domain concepts or rules this use case depends on are established and correctly referenced, not duplicated inline.
- The Specification's content does not contradict Domain, Architecture, Stack, Conventions, Design Knowledge, or applicable ADRs.

---

# Material Gaps

- No material decision — one Implementation, Testing, or Verification would otherwise have to invent — remains unresolved.

---

# Readiness Criterion Result

Every criterion must produce one of the following results:

- SATISFIED
- MATERIAL GAP
- NOT APPLICABLE

`READY_FOR_APPROVAL` requires every applicable criterion to be SATISFIED and none to be a MATERIAL GAP. A criterion resolved NOT APPLICABLE never blocks readiness. Supporting reasoning for each MATERIAL GAP or NOT APPLICABLE result is mandatory.
