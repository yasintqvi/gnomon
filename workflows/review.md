# Review Workflow

## Purpose

Define the execution process for the Review intent.

Review evaluates an artifact using reusable engineering questions together with project knowledge and available verification evidence.

The objective is to identify engineering findings that cannot be established through rule-based verification alone.

---

## Intent

Evaluate whether the current implementation and evaluation sufficiently address relevant engineering scenarios.

---

## Inputs

### User Input

- Review target
- Review scope
- Constraints

### Project Context

- Repository
- Related artifacts

### Required Knowledge

- Relevant Specifications
- DOMAIN
- ARCHITECTURE
- STACK
- ADRs
- Contracts

### Evaluation Assets

ENGINEERING_REVIEW_QUESTIONS.md

### Optional Inputs

Verification Evidence

If verification evidence exists, it should be consumed.

If unavailable, the review should explicitly report reduced confidence.

---

## Execution

### 1. Understand

Determine:

- Review target
- Scope
- Applicable project knowledge

---

### 2. Load Questions

Load applicable engineering review questions.

Ignore questions outside the authorized scope.

---

### 3. Evaluate

For every applicable question:

Evaluate using:

- Artifact
- Project Knowledge
- Verification Evidence (when available)

Determine whether the current evaluation sufficiently addresses the engineering scenario.

---

### 4. Produce Findings

Possible outcomes:

- PASS
- DEFECT
- RISK
- KNOWLEDGE GAP
- NOT APPLICABLE

Supporting reasoning is mandatory.

---

### 5. Determine Resolution

#### Purpose

Determine the appropriate resolution path for every approved finding.

Review does not resolve findings.

Instead, it identifies the owning project artifact responsible for resolving each finding.

#### Required Knowledge

* Classified Findings
* Applicable Project Knowledge

#### Expected Result

Every finding identifies:

* Finding Classification
* Resolution Owner
* Whether an explicit project decision is required
* Recommended next workflow

Possible Resolution Owners include:

* Implementation
* Specification
* Architecture
* ADR
* Contract
* Workflow
* Evaluation
* Documentation

If resolving a finding requires a new product, business, or architectural decision, the finding must be marked as **Decision Required**.

The Review workflow must not invent or approve such decisions.

---

### Primary Output

An evidence-based engineering review report.

### Supporting Outputs

* Classified Findings
* Resolution Owner for every finding
* Decision Required indicators
* Risk Assessment
* Compliance Summary
* Remaining Uncertainties

---

### Finding Identity

Every actionable finding must receive a unique identifier within the review report.

Identifiers must use the following format:

`F-001`, `F-002`, `F-003`, ...

The identifier must remain stable when the same review result is referenced for approval or resolution.

Related observations that share the same underlying issue should not receive separate identifiers unless they require independent resolution.

---

## Rules

- Review must not redefine project knowledge.
- Review must distinguish facts from engineering reasoning.
- Review must report uncertainty explicitly.
- Review may recommend creating new project knowledge when required.
- Review must not silently assume undefined behavior.
Review identifies findings.
- Review must not propose implementation solutions.
- When a finding requires a project decision, the outcome shall be reported as a Knowledge Gap.
- Review must not repeat verification evidence unless required to support a finding.
* Review must not resolve findings directly.
* Review must not modify project knowledge.
* Review must not modify implementation.
* Review must identify the appropriate Resolution Owner for every finding.
* Review must explicitly indicate whether resolving a finding requires a new project decision.
* Review recommendations are not approved actions and require explicit user approval before resolution.

---

## Completion Criteria

* Review completes when every applicable engineering question has been evaluated.
* Every finding has an identified Resolution Owner.
* Findings requiring explicit project decisions have been identified.