# Artifact Contract — Input Qualification

## Purpose

Define the contract for converting candidate or unqualified input into a declared acceptable input outcome before consequential processing.

---

## Artifact Type

**Name**

Input Qualification

**Description**

A responsibility, represented at boundaries chosen by project Architecture, that validates and safely normalizes input for subsequent processing.

Where authoritative project knowledge assigns authorization-related admissibility checks to the same boundary, the responsibility applies those checks without assuming ownership of authorization itself.

This contract does not require a particular artifact name, boundary type, data representation, or implementation technology.

---

## Scope

This contract applies to structural validation, safe normalization, admissibility checks, and other input-qualification rules assigned to the qualifying boundary.

Authorization-related checks apply only when authoritative project knowledge assigns them to this boundary.

This contract does not define business requirements, business invariants, feature behavior, authorization policy, architectural representation, project conventions, or technology selection.

---

## Responsibility

The artifact is responsible for:

* Treating input crossing an applicable trust boundary as untrusted.
* Evaluating candidate input against approved qualification rules assigned to this boundary.
* Rejecting malformed, missing, unsupported, or otherwise inadmissible input.
* Safely normalizing accepted values where required.
* Producing only qualified values and safe, stable qualification failures.

The artifact must not:

* Persist authoritative state.
* Execute the owning use case.
* Initiate unrelated consequential effects.
* Invent validation, admissibility, or authorization rules.
* Reimplement business invariants owned by another boundary.
* Assume ownership of authorization merely because authorization-related admissibility is evaluated here.

---

## Inputs

| Input                 | Description                                                                            | Required    |
| --------------------- | -------------------------------------------------------------------------------------- | ----------- |
| Candidate input       | Values or signals requiring qualification                                              | Yes         |
| Evaluation context    | Applicable identity, resource, tenancy, locale, trust, or other contextual information | As required |
| Qualification context | Approved contextual information required to apply qualification rules                  | As required |

---

## Outputs

| Output                | Description                                                                                         |
| --------------------- | --------------------------------------------------------------------------------------------------- |
| Qualified input       | Accepted values in the declared representation                                                      |
| Qualification failure | Stable and safely exposable rejection of input that does not satisfy applicable qualification rules |

---

## Behavior

### Accepted Input

**Condition**

All applicable qualification rules assigned to this boundary succeed.

**Expected Behavior**

Expose only declared, safely qualified values for subsequent processing.

### Rejected Input

**Condition**

One or more applicable qualification rules fail or input cannot be normalized safely.

**Expected Behavior**

Stop the affected input from crossing the qualifying boundary and expose a safe declared failure without forbidden consequential effects.

### Authorization-Related Admissibility

**Condition**

Authoritative project knowledge assigns an authorization-related admissibility check to this boundary.

**Expected Behavior**

Apply the approved check through its owning authorization capability and expose the approved safe outcome without assuming broader authorization ownership.

---

## Constraints

The artifact must:

* Apply only explicit, approved qualification and normalization rules assigned to its boundary.
* Preserve the distinction between input qualification and authoritative business invariants.
* Use owning capabilities when qualification depends on responsibilities outside this boundary.
* Keep failure identity and semantics independent from localized presentation where user-facing localization applies.

The artifact must not:

* Trust claimed metadata as authoritative when an approved source of truth is required.
* Mutate authoritative state.
* Execute use cases or transactional business workflows.
* Invoke unrelated consequential effects.
* Duplicate rules owned by other authoritative boundaries.
* Introduce qualification or authorization behavior absent from approved project knowledge.

---

## Dependencies

### Allowed Dependencies

* Dependencies permitted by applicable Architecture, Stack, Conventions, and authoritative qualification rules.

### Restricted Dependencies

* Consequential execution responsibilities not required for qualification.
* State-changing capabilities unrelated to qualification.
* Provider implementations where an approved boundary exists.
* Protected internals of another module or subsystem.
* Dependencies prohibited by Architecture, Stack, Conventions, or applicable ADRs.

---

## Error Handling

### Malformed or Inadmissible Input

**Condition**

Input cannot be safely normalized or does not satisfy an applicable qualification rule.

**Expected Result**

Expose a safe, stable qualification failure and prevent the affected input from crossing the qualifying boundary.

### Authorization-Related Rejection

**Condition**

An authorization-related admissibility check assigned to this boundary rejects the input or context.

**Expected Result**

Expose the approved safe authorization outcome without revealing protected existence or details and without assuming ownership of broader authorization behavior.

---

## Integration

### Consuming Boundary

**Participant**

The architecture-defined consumer of qualified input.

**Interaction**

The consumer receives the qualified outcome through an approved boundary and does not bypass required qualification with unqualified values.

**Constraints**

* Qualification complements rather than replaces authoritative business invariants.
* Responsibilities owned by other boundaries remain with their approved owners.
* Integration must preserve the distinction between qualification and consequential execution.

---

## Verification

* Evidence demonstrates acceptance and safe normalization of valid input where applicable.
* Evidence demonstrates rejection of malformed, missing, unsupported, boundary-invalid, or otherwise inadmissible input where applicable.
* Evidence demonstrates applicable authorization-related admissibility behavior where assigned to this boundary.
* Evidence demonstrates that rejected input does not cross the qualifying boundary or produce forbidden consequential effects.
* Evidence demonstrates that only qualified declared values reach the consuming boundary.
* Evidence demonstrates safe and stable failure semantics.
* Evidence demonstrates that business invariants and responsibilities owned by other boundaries are not reimplemented as input qualification.

---

## Examples

### Valid Example

```
Candidate values are evaluated against approved qualification rules.

Accepted values are safely normalized into the declared representation.

Rejected values produce safe failures before crossing the qualifying boundary.

Business invariants and authorization ownership remain with their approved boundaries.
```

### Invalid Example

```
Claimed metadata is trusted without authority,
business rules are duplicated as input validation,
rejected values reach consequential execution,
and the qualifying responsibility mutates authoritative state.
```

---

## Related Architecture

* Trust and Input Boundaries
* Authorization Boundaries
* Execution Boundaries

---

## Related Conventions

* Input Representation
* Validation and Error Handling
* Localization
* Testing

---

## Notes

Specifications and Domain knowledge define valid behavior and business invariants.

Architecture, Stack, Conventions, ADRs, and other authoritative project knowledge determine where qualification occurs, which rules belong to that boundary, how qualified input is represented, and where authorization responsibilities are enforced.

Input Qualification determines whether candidate input may cross its qualifying boundary; it does not determine what the owning use case does with that input.
