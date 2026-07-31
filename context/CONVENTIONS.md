# Conventions

## Purpose

Define the project-wide conventions used to maintain consistency across the codebase.

This document establishes common implementation conventions for naming, organization, formatting, structure, and recurring development patterns.

Conventions should improve consistency without defining business behavior or architectural responsibilities.

---

## Scope

This document covers:

* Naming conventions
* File and directory conventions
* Code organization
* Formatting conventions
* Common implementation patterns
* Documentation conventions
* Testing conventions
* General consistency rules

This document does not cover:

* Business rules
* Domain concepts
* Architectural structure
* Technology selection
* Feature requirements
* Artifact-specific behavioral contracts
* Architectural decisions

---

## Naming

### Classes

[Define class naming conventions.]

Example:

```text
[Example]
```

### Methods

[Define method naming conventions.]

Example:

```text
[Example]
```

### Variables

[Define variable naming conventions.]

Example:

```text
[Example]
```

### Files

[Define file naming conventions.]

Example:

```text
[Example]
```

### Directories

[Define directory naming conventions.]

Example:

```text
[Example]
```

---

## Code Organization

Define conventions for organizing code within files and directories.

### File Structure

[Describe the preferred ordering and organization of code inside files.]

### Directory Structure

[Describe recurring directory organization conventions that are not architectural decisions.]

---

## Implementation Patterns

Document recurring implementation patterns used for consistency.

### [Pattern Name]

**Purpose**

[Explain the consistency problem this pattern solves.]

**Convention**

[Describe how the pattern should normally be applied.]

**Example**

```text
[Example]
```

---

## Formatting

Define formatting conventions that are not already enforced automatically by project tooling.

* [Formatting rule]
* [Formatting rule]
* [Formatting rule]

---

## Documentation

Define conventions for documenting code and project artifacts.

* [Documentation rule]
* [Documentation rule]

---

## Testing Conventions

Define conventions for organizing and naming tests.

### Test Naming

[Define test naming convention.]

### Test Organization

[Define test organization convention.]

### Test Structure

[Define recurring test structure conventions.]

---

## Error Handling

Define general consistency conventions for handling errors.

* [Convention]
* [Convention]

Business error behavior belongs to the relevant domain or specification.

Artifact-specific error behavior belongs to the relevant Artifact Contract.

---

## Comments

Define when comments should be used.

* [Comment convention]
* [Comment convention]

Comments should explain intent or non-obvious reasoning rather than restating code.

---

## Exceptions

Define when a convention may be intentionally violated.

An exception should:

* Have a clear justification.
* Not contradict architectural constraints.
* Not violate business rules.
* Be documented when the deviation could affect future development.

---

## Enforcement

Describe how conventions are enforced.

Examples:

* Automated formatting
* Static analysis
* Linters
* Code review
* Automated tests

---

## Evolution Rules

* New conventions should be added only when they solve a recurring consistency problem.
* Existing conventions should not be changed implicitly during feature implementation.
* Changes that affect architecture belong in `ARCHITECTURE.md` or an ADR.
* Changes that affect business behavior belong in the relevant domain or specification.
* Artifact-specific behavioral rules belong in Artifact Contracts.

---

## Notes

[Optional clarification that does not belong to another section.]
