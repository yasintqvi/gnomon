# Architecture

## Purpose

Define the structural design of the software system.

This document describes how the system is organized, how its major responsibilities are separated, how components interact, and which architectural constraints must be respected.

It provides the structural context required to design and implement features consistently.

---

## Scope

This document covers:

* Architectural structure
* System boundaries
* Major components and layers
* Component responsibilities
* Dependencies and interaction rules
* Architectural constraints
* Structural patterns
* Placement rules for implementation artifacts

This document does not cover:

* Business requirements
* Feature-specific behavior
* Detailed implementation steps
* Coding style
* Framework-specific syntax
* Individual feature specifications
* Artifact-specific behavioral contracts

---

## Architectural Overview

[Describe the overall architectural structure of the system.]

```text
[Component / Layer]
        ↓
[Component / Layer]
        ↓
[Component / Layer]
```

---

## Components

### [Component or Layer Name]

**Purpose**

[Explain why this component or layer exists.]

**Responsibilities**

* [Architectural responsibility]
* [Architectural responsibility]

**Dependencies**

* [Allowed dependency]
* [Allowed dependency]

**Constraints**

* [Architectural constraint]
* [Architectural constraint]

---

### [Component or Layer Name]

**Purpose**

[Explain the purpose of this component or layer.]

**Responsibilities**

* [Architectural responsibility]

**Dependencies**

* [Allowed dependency]

**Constraints**

* [Architectural constraint]

---

## Dependency Rules

Define which components or layers may depend on each other.

| Source        | Allowed Dependency | Notes        |
| ------------- | ------------------ | ------------ |
| [Component A] | [Component B]      | [Constraint] |
| [Component B] | [Component C]      | [Constraint] |

Dependencies not explicitly allowed by the architecture should not be introduced without an architectural decision.

---

## Interaction Rules

Describe important interaction patterns between architectural components.

### [Interaction]

**Participants**

* [Component]
* [Component]

**Flow**

```text
[Component A]
      ↓
[Component B]
      ↓
[Component C]
```

**Rules**

* [Interaction rule]
* [Interaction constraint]

---

## Artifact Placement

Define where different categories of implementation artifacts belong within the architecture.

| Artifact Type   | Architectural Location | Responsibility   |
| --------------- | ---------------------- | ---------------- |
| [Artifact Type] | [Location]             | [Responsibility] |
| [Artifact Type] | [Location]             | [Responsibility] |

Artifact placement must follow the architecture rather than being defined independently by individual specifications.

---

## Architectural Constraints

* [Architectural constraint]
* [Architectural constraint]
* [Architectural constraint]

These constraints apply across the system unless explicitly overridden by an approved architectural decision.

---

## Architectural Patterns

Document architectural patterns that are intentionally used by the system.

### [Pattern Name]

**Purpose**

[Explain why the pattern is used.]

**Application**

[Describe where and how the pattern applies.]

**Constraints**

* [Constraint]
* [Constraint]

---

## Boundaries

### In Scope

* [Architectural area]
* [Architectural component]
* [Architectural concern]

### Out of Scope

* [Excluded architectural area]
* [Excluded concern]

---

## Evolution Rules

Describe how architectural changes should be handled.

* Significant architectural changes require an ADR.
* New architectural dependencies must follow the dependency rules.
* Architectural constraints must not be changed implicitly during feature implementation.

---

## Notes

[Optional architectural clarification that does not belong to another section.]
