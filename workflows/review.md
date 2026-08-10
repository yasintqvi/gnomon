# Review Workflow

## Purpose

Define the execution process for the **Review** intent.

Review evaluates an artifact using reusable engineering questions together with approved project knowledge and available verification evidence.

The objective is to identify engineering defects, risks, knowledge gaps, and other concerns that require engineering reasoning and cannot be established through objective verification alone.

Review evaluates existing work.

It does not redefine project knowledge, modify implementation, or approve resolutions.

---

## Intent

Evaluate whether the current implementation and available evaluation sufficiently address relevant engineering scenarios within the authorized review scope.

---

## When to Use

Use this workflow when:

* An implementation or other project artifact requires engineering evaluation.
* Objective verification alone is insufficient to evaluate engineering quality or risk.
* Potential defects, risks, or missing project knowledge must be identified.
* Verification Evidence requires engineering interpretation.

Do not use this workflow when:

* The task is limited to objective compliance verification.
* The task is to implement or modify the target.
* The task is to define new product, business, architectural, or project-policy knowledge.
* A previously identified finding is being resolved.

---

## Inputs

### User Input

* Review target
* Review scope
* Explicit constraints

### Project Context

* Repository
* Existing implementation
* Related artifacts
* Existing project patterns

### Required Knowledge

As applicable to the review target:

* Relevant Specifications
* Relevant Domain knowledge
* `PROJECT.md`
* `ARCHITECTURE.md`
* `STACK.md`
* `CONVENTIONS.md`
* Applicable ADRs
* Applicable Artifact Contracts

### Evaluation Assets

* Applicable engineering review questions

### Optional Inputs

* Verification Evidence

When Verification Evidence exists, Review should consume it rather than independently repeating equivalent verification work.

When relevant Verification Evidence is unavailable, Review must explicitly report the resulting reduction in confidence where it materially affects a conclusion.

---

## Execution

### 1. Understand

#### Purpose

Determine the review target, authorized scope, governing project knowledge, and available evidence.

#### Expected Result

Identify:

* Review target
* Authorized review scope
* Applicable project knowledge
* Applicable Artifact Contracts
* Applicable engineering review questions
* Available Verification Evidence
* Relevant existing implementation and project patterns

Questions and knowledge outside the authorized scope must not affect the review result.

---

### 2. Derive Review Questions

#### Purpose

Determine which engineering scenarios require evaluation for the target.

#### Sources

Review questions may originate from:

* Reusable engineering review questions
* Relevant Specifications
* Project-level requirements
* Domain knowledge
* Architectural boundaries
* Stack constraints
* Project conventions
* Applicable ADRs
* Applicable Artifact Contracts
* Existing implementation relationships
* Available Verification Evidence

#### Expected Result

A bounded set of engineering questions relevant to the authorized review scope.

Review must not invent new requirements while deriving questions.

Questions derived from project knowledge evaluate the engineering consequences of existing approved knowledge; they do not create new project rules.

---

### 3. Evaluate

#### Purpose

Evaluate every applicable engineering question using available evidence and engineering reasoning.

Evaluation may use:

* Review target
* Existing implementation
* Approved project knowledge
* Existing project patterns
* Verification Evidence
* Relevant technical evidence

Review must distinguish:

* Observed facts
* Existing project requirements
* Engineering reasoning
* Uncertainty

Review must not treat an engineering preference as an approved project requirement.

---

### 4. Produce Findings

Each applicable review question receives one of the following outcomes:

* `PASS`
* `DEFECT`
* `RISK`
* `KNOWLEDGE GAP`
* `NOT APPLICABLE`

#### PASS

Available evidence and engineering analysis identify no material concern for the evaluated scenario.

#### DEFECT

The artifact demonstrably fails to satisfy approved behavior, project knowledge, or a required engineering invariant.

#### RISK

The current artifact may satisfy known requirements but exposes a credible engineering failure mode, maintainability concern, operational concern, security concern, usability concern, or other material engineering weakness.

A Risk must not be based solely on stylistic preference.

#### KNOWLEDGE GAP

The scenario cannot be resolved safely because required product, business, architectural, project-policy, contract, or other authoritative knowledge is missing or materially ambiguous.

#### NOT APPLICABLE

The engineering scenario does not apply to the review target or authorized scope.

Supporting reasoning is mandatory for every non-trivial outcome.

---

### 5. Identify Actionable Findings

Only `DEFECT`, `RISK`, and `KNOWLEDGE GAP` outcomes produce actionable findings.

Every actionable finding must identify:

* Finding ID
* Classification
* Summary
* Evidence
* Engineering reasoning
* Impact
* Resolution Owner
* Decision Required status
* Recommended next workflow

Review may explain the nature of the required resolution but must not design or implement the solution.

---

### 6. Determine Resolution

#### Purpose

Determine the appropriate resolution path for each actionable finding.

Review does not resolve findings.

It identifies which project artifact or implementation area owns the missing or incorrect knowledge.

#### Possible Resolution Owners

Depending on the finding, ownership may belong to:

* Implementation
* Specification
* Domain
* Project
* Architecture
* ADR
* Stack
* Conventions
* Contract
* Workflow
* Evaluation
* Documentation

#### Decision Required

A finding must be marked **Decision Required** when resolution requires new or changed authoritative project knowledge rather than correction of implementation against existing knowledge.

Review must not invent, approve, or silently select that decision.

#### Expected Result

Every actionable finding identifies:

* Finding Classification
* Resolution Owner
* Decision Required: Yes or No
* Recommended next workflow

---

## Finding Identity

Every actionable finding must receive a unique identifier within the review report.

Identifiers use:

`F-001`, `F-002`, `F-003`, ...

The identifier must remain stable when the same finding is referenced for approval or resolution.

Related observations that share the same underlying issue should remain one finding unless they require independent resolution.

---

## Outputs

### Primary Output

**Engineering Review Report**

An evidence-based evaluation of the applicable engineering scenarios.

### Supporting Outputs

As applicable:

* Classified Findings
* Resolution Owner for every actionable finding
* Decision Required indicators
* Risk Assessment
* Compliance Summary
* Remaining Uncertainties

### Finding Model

```text
Finding ID:
Classification: DEFECT | RISK | KNOWLEDGE GAP
Summary:
Evidence:
Engineering Reasoning:
Impact:
Resolution Owner:
Decision Required: Yes | No
Recommended Next Workflow:
```

---

## Rules

* Review must remain within the authorized scope.
* Review must not redefine project knowledge.
* Review must distinguish observed facts from engineering reasoning.
* Review must distinguish approved requirements from engineering preferences.
* Review must report material uncertainty explicitly.
* Review must not silently assume undefined behavior.
* Review must not modify implementation.
* Review must not modify project knowledge.
* Review must not resolve findings directly.
* Review must not approve project decisions.
* Review must not propose detailed implementation solutions.
* Review may identify that new or changed project knowledge is required.
* When a finding requires a project decision, it must be classified as `KNOWLEDGE GAP`.
* Every actionable finding must identify its Resolution Owner.
* Every actionable finding must identify whether a project decision is required.
* Review must not repeat Verification Evidence unless required to support engineering reasoning or a finding.
* Existing Verification Evidence should be reused where applicable.
* Lack of Verification Evidence must not automatically produce a defect.
* Engineering preferences alone must not produce defects.
* Review recommendations are not approved actions and require explicit authorization before resolution.

---

## Failure Handling

### Missing Verification Evidence

When relevant Verification Evidence is unavailable:

1. Continue review where sufficient evidence exists.
2. Identify conclusions whose confidence is materially reduced.
3. Do not invent verification results.
4. Do not classify absence of verification alone as a defect.

---

### Missing Project Knowledge

When engineering evaluation requires authoritative knowledge that is missing or materially ambiguous:

1. Identify the missing knowledge.
2. Do not invent the expected rule.
3. Produce a `KNOWLEDGE GAP` when the gap materially affects the reviewed scenario.
4. Identify the appropriate Resolution Owner.
5. Mark `Decision Required` when resolution requires an explicit project decision.

---

### Conflicting Project Knowledge

When applicable approved sources conflict:

1. Identify the conflicting sources.
2. Apply established project knowledge ownership or precedence rules when available.
3. If the conflict remains material, produce a `KNOWLEDGE GAP`.
4. Do not silently choose one interpretation.

---

### Insufficient Technical Evidence

When a question cannot be evaluated responsibly from available technical evidence:

1. Identify the missing evidence.
2. Report the resulting uncertainty.
3. Do not fabricate a conclusion.
4. Recommend Verification when objective evidence can resolve the uncertainty.

---

## Completion Criteria

Review completes when:

* The review target and authorized scope are defined.
* Applicable project knowledge has been identified.
* Applicable engineering questions have been evaluated.
* Available Verification Evidence has been consumed where relevant.
* Every actionable finding has a stable identifier.
* Every actionable finding has a Resolution Owner.
* Every actionable finding identifies whether a project decision is required.
* Material uncertainty has been reported.
* No implementation or project knowledge has been modified.

---

## Workflow Constraints

* Review evaluates; it does not implement.
* Review reasons about engineering quality; it does not create requirements.
* Review may consume Verification Evidence but must not silently replace Verification.
* Review must remain traceable to the reviewed artifact, approved project knowledge, and available evidence.
* Detailed resolution work occurs only after the relevant finding has been explicitly authorized.

---

## Notes

Verification asks:

> Does objective evidence demonstrate compliance with approved obligations?

Review asks:

> Given the approved project knowledge and available evidence, does this implementation expose a material engineering defect, risk, or unresolved knowledge gap?

Resolution asks:

> What approved change should be made to address that finding?

These responsibilities must remain separate.
