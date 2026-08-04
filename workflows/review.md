# Review Workflow

## Purpose

Define the execution process for the **Review Artifact** intent.

This workflow describes how the system evaluates an existing artifact against approved project knowledge and produces actionable findings without introducing new product, business, or architectural decisions.

The workflow may review any project artifact, including documentation, specifications, architecture, implementation, tests, configurations, or other approved project assets.

---

## Intent

Evaluate an existing artifact for correctness, consistency, completeness, and compliance with approved project knowledge.

---

## When to Use

Use this workflow when:

* An artifact must be reviewed before approval or implementation.
* An implementation must be evaluated after completion.
* Documentation, Specifications, Architecture, ADRs, or Contracts require validation.
* A pull request or change set requires review.
* Project consistency must be verified.

Do not use this workflow when:

* The requested work is implementation rather than evaluation.
* The requested behavior has not yet been defined.
* The task requires creating new project knowledge.
* The objective is automated verification through testing rather than analytical review.

---

## Inputs

### User Input

The artifact to review, review scope, explicit review objectives, and any user-defined constraints.

### Project Context

The current project state, related artifacts, repository contents, and approved project knowledge.

### Required Knowledge

* Relevant Specification
* Relevant Domain knowledge
* `ARCHITECTURE.md`
* `STACK.md`
* `CONVENTIONS.md`
* Applicable ADRs
* Applicable Artifact Contracts
* `REVIEW_CHECKLIST.md`
* Existing project artifacts related to the review target

---

## Execution

### 1. Understand

#### Purpose

Understand the artifact being reviewed, its intended purpose, ownership, and review boundaries.

#### Required Knowledge

* Target artifact
* User-defined scope
* Relevant project knowledge

#### Expected Result

A clearly defined review scope containing:

* Target artifact
* Intended purpose
* Applicable project knowledge
* Explicit exclusions
* Review boundaries

---

### 2. Inspect

#### Purpose

Inspect the artifact and identify its observable behavior, structure, dependencies, assumptions, and responsibilities.

#### Required Knowledge

* Target artifact
* Related project artifacts
* Applicable contracts

#### Expected Result

An inspection summary identifying:

* Existing behavior
* Responsibilities
* Dependencies
* Assumptions
* Missing information
* Relevant relationships

---

### 3. Evaluate

#### Purpose

Evaluate the artifact against approved project knowledge.

#### Required Knowledge

* Relevant Specification
* Domain
* Architecture
* ADRs
* Conventions
* Artifact Contracts
* Review Checklist

#### Expected Result

An evidence-based evaluation identifying:

* Compliance
* Violations
* Inconsistencies
* Missing requirements
* Unnecessary complexity
* Maintainability concerns
* Potential risks

---

### 4. Classify

#### Purpose

Classify every finding according to its nature and impact.

#### Required Knowledge

* Evaluation results
* Review Checklist

#### Expected Result

Every finding is classified using appropriate categories, such as:

* Defect
* Requirement mismatch
* Contract violation
* Architectural violation
* Convention violation
* Documentation issue
* Improvement opportunity
* Risk
* Recommendation

Each finding should include an estimated severity and sufficient evidence.

---

### 5. Report

#### Purpose

Produce an actionable review report.

#### Required Knowledge

* Classified findings

#### Expected Result

A structured review report containing:

* Review scope
* Reviewed artifacts
* Positive observations
* Actionable findings
* Severity
* Supporting evidence
* Remaining uncertainties
* Recommended next actions

---

## Rules

* Review existing artifacts without redefining project knowledge.
* Base every finding on evidence from approved project knowledge.
* Distinguish facts from assumptions.
* Do not invent missing requirements.
* Do not introduce new business rules.
* Do not introduce new architectural decisions.
* Respect the ownership of every project document.
* Separate confirmed defects from improvement suggestions.
* Report uncertainty explicitly.
* Preserve existing user-owned work.
* Produce actionable findings rather than subjective opinions.
* Explain why each finding matters.

The workflow must not introduce business rules that are not defined by the relevant Specification or Domain.

The workflow must not introduce architectural rules that are not defined by the Architecture or an approved ADR.

---

## Outputs

### Primary Output

An evidence-based review report.

### Supporting Outputs

* Classified findings
* Improvement recommendations
* Risk assessment
* Compliance summary
* Remaining uncertainties

---

## Failure Handling

### Missing Information

When required knowledge is unavailable:

1. Identify the missing information.
2. Determine which project document owns it.
3. Continue reviewing unaffected areas where possible.
4. Report review limitations explicitly.
5. Request clarification only when it materially affects the review.

Do not invent missing project knowledge.

---

### Conflicting Information

When project knowledge conflicts:

1. Identify the conflicting sources.
2. Apply the project's knowledge ownership rules.
3. Report the conflict explicitly.
4. Avoid selecting one interpretation without authority.
5. Continue reviewing unaffected areas when possible.

---

### Insufficient Evidence

When available evidence is insufficient:

1. Report the limitation.
2. Avoid speculative conclusions.
3. Mark the finding as inconclusive.
4. Recommend additional evidence when appropriate.

---

## Completion Criteria

The workflow is complete when:

* The authorized review scope has been evaluated.
* Applicable project knowledge has been considered.
* Findings are evidence-based.
* Findings are classified.
* Actionable recommendations have been produced.
* Remaining uncertainties have been reported.
* No new project knowledge has been introduced.

---

## Workflow Constraints

* Review evaluates existing artifacts; it does not create new ones.
* Review must remain within the authorized scope.
* Recommendations must not be presented as approved decisions.
* Review should prioritize correctness, consistency, maintainability, and compliance.
* Review should minimize subjective judgment.

---

## Notes

Review may reveal missing Specifications, weak Contracts, architectural inconsistencies, implementation defects, documentation gaps, or testing deficiencies.

Such findings should be reported to the document or workflow that owns the unresolved issue rather than being resolved implicitly during review.
