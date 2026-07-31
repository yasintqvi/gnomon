# Review Checklist

## Purpose

Define the project-wide verification criteria used to review generated or manually implemented artifacts.

This checklist ensures that outputs remain aligned with the project specification, architecture, conventions, stack, artifact contracts, and quality expectations.

---

## Scope

This checklist applies to:

* Generated code
* Manually written code
* Tests
* Documentation
* Configuration
* Refactoring changes

This checklist does not define:

* Business requirements
* Architectural decisions
* Feature-specific acceptance criteria
* Artifact-specific behavior

Those requirements must be obtained from their corresponding sources.

---

## Specification Compliance

* [ ] The implementation satisfies the relevant Specification.
* [ ] All applicable acceptance criteria are covered.
* [ ] The implementation remains within the defined scope.
* [ ] No unapproved business behavior has been introduced.
* [ ] Business constraints are preserved.

---

## Domain Compliance

* [ ] Domain terminology is used consistently.
* [ ] Domain rules and invariants are preserved.
* [ ] Domain concepts are not redefined inside implementation code.
* [ ] No implementation detail has been treated as a business rule.

---

## Architecture Compliance

* [ ] Responsibilities are placed in the correct architectural location.
* [ ] Dependency rules are respected.
* [ ] Architectural boundaries are preserved.
* [ ] No new architectural pattern has been introduced without approval.
* [ ] Relevant ADRs have been followed.

---

## Convention Compliance

* [ ] Naming follows project conventions.
* [ ] File and directory organization follows project conventions.
* [ ] Code organization and formatting are consistent.
* [ ] Error handling follows project conventions.
* [ ] Documentation follows project conventions.

---

## Stack Compliance

* [ ] Only approved technologies and libraries are used.
* [ ] Required version constraints are respected.
* [ ] Deprecated or unsupported technologies have not been introduced.
* [ ] Framework and runtime capabilities are used correctly.

---

## Artifact Contract Compliance

For every generated or modified artifact:

* [ ] The artifact follows its applicable contract.
* [ ] The artifact performs only its defined responsibilities.
* [ ] The artifact does not take responsibility owned by another artifact type.
* [ ] Inputs and outputs comply with the contract.
* [ ] Dependency restrictions are respected.
* [ ] Required error behavior is implemented.

---

## Code Quality

* [ ] The implementation is readable and understandable.
* [ ] Responsibilities are clearly separated.
* [ ] Unnecessary abstractions have not been introduced.
* [ ] Duplicate logic has been avoided where appropriate.
* [ ] Complex behavior is justified.
* [ ] Comments explain intent rather than restating code.
* [ ] Dead or unrelated code has not been introduced.

---

## Testing

* [ ] Required automated tests exist.
* [ ] Tests cover the primary success flow.
* [ ] Tests cover relevant failure flows.
* [ ] Tests verify business rules and acceptance criteria.
* [ ] Tests are deterministic.
* [ ] Tests follow project conventions.
* [ ] Existing tests remain passing.

---

## Security

* [ ] User input is handled safely.
* [ ] Authorization and ownership rules are respected.
* [ ] Sensitive information is not exposed.
* [ ] Secrets or credentials are not committed.
* [ ] Security-relevant failures are handled appropriately.

---

## Data Integrity

* [ ] Database constraints support important invariants where necessary.
* [ ] Persistence behavior is atomic where required.
* [ ] Concurrent operations cannot silently violate critical rules.
* [ ] Data ownership and relationships remain valid.
* [ ] Destructive changes are intentional and reviewed.

---

## Error Handling

* [ ] Expected errors are handled explicitly.
* [ ] Business errors are distinguishable from unexpected failures.
* [ ] Errors are translated at the appropriate boundary.
* [ ] Unexpected exceptions are not silently swallowed.
* [ ] User-facing errors do not expose internal details.

---

## Documentation

* [ ] Relevant documentation has been updated.
* [ ] New architectural decisions are recorded in ADRs when required.
* [ ] New recurring conventions are documented when justified.
* [ ] Artifact Contracts are updated when their behavior changes.
* [ ] Documentation does not duplicate another source of truth.

---

## Scope Control

* [ ] The change contains only work required by the request.
* [ ] Unrelated refactoring has not been included.
* [ ] Hypothetical future requirements have not been implemented.
* [ ] New abstractions solve a demonstrated problem.
* [ ] Deferred work remains deferred.

---

## Final Verification

* [ ] The requested output has been produced.
* [ ] The output is usable in its intended environment.
* [ ] Syntax, build, or static analysis checks pass.
* [ ] Automated tests pass.
* [ ] No unresolved ambiguity has been hidden by assumptions.
* [ ] The result is ready for human review.
