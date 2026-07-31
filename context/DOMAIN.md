# Domain

## Purpose

Define the business domain of the system.

This document establishes the shared business language, core business concepts, relationships, rules, and domain boundaries required to understand the system.

It must remain independent of implementation details, frameworks, libraries, infrastructure, and technology choices.

---

## Scope

This document covers:

* Business concepts
* Business terminology
* Relationships between domain concepts
* Domain rules
* Domain constraints
* Domain boundaries

This document does not cover:

* Application architecture
* Implementation details
* Framework-specific behavior
* Database design
* API design
* UI behavior
* Technology choices
* Coding conventions

---

## Concepts

### [Concept Name]

**Definition**

[Define the business concept clearly and concisely.]

**Purpose**

[Explain why this concept exists in the business domain.]

**Responsibilities**

* [Business responsibility]
* [Business responsibility]

**Attributes**

* `[attribute]` — [Business meaning]
* `[attribute]` — [Business meaning]

**Rules**

* [Business rule governing this concept]
* [Business constraint governing this concept]

---

### [Concept Name]

**Definition**

[Define the business concept.]

**Purpose**

[Explain its business purpose.]

**Responsibilities**

* [Business responsibility]

**Attributes**

* `[attribute]` — [Business meaning]

**Rules**

* [Business rule]

---

## Relationships

Describe meaningful relationships between domain concepts.

### [Concept A] → [Concept B]

**Relationship**

[Describe the business relationship.]

**Cardinality**

[Describe the business cardinality when relevant.]

**Rules**

* [Relationship-specific business rule]

---

## Domain Rules

Define business rules that apply across multiple concepts or represent important domain invariants.

* [Domain rule]
* [Domain rule]
* [Domain rule]

Each rule must describe business behavior or a business constraint, not an implementation mechanism.

---

## Domain Constraints

Define limitations imposed by the business domain.

* [Constraint]
* [Constraint]

---

## Domain Boundaries

### In Scope

* [Business area]
* [Business concept]
* [Business capability]

### Out of Scope

* [Excluded business area]
* [Excluded business concept]
* [Excluded business capability]

---

## Terminology

| Term   | Definition            |
| ------ | --------------------- |
| [Term] | [Business definition] |
| [Term] | [Business definition] |

Use this section for terms that require a precise shared definition across the project.

---

## Notes

[Optional domain-level clarification that does not belong to another section.]
