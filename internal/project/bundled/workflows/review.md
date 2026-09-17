---
identity: review
specification_reference: none
requires_approved_specification: false
result:
  terminal_path: summary.aggregate
  classification:
    PASS: success
    DEFECT: blocked
    RISK: blocked
    "KNOWLEDGE GAP": blocked
  schema:
    type: object
    required:
      - summary
    properties:
      findings:
        type:
          - array
          - "null"
        items:
          type: object
          required:
            - finding_id
            - classification
          properties:
            finding_id:
              type: string
            classification:
              type: string
              enum:
                - DEFECT
                - RISK
                - "KNOWLEDGE GAP"
            summary:
              type:
                - string
                - "null"
            evidence:
              type:
                - string
                - "null"
            engineering_reasoning:
              type:
                - string
                - "null"
            impact:
              type:
                - string
                - "null"
            resolution_owner:
              type:
                - string
                - "null"
            decision_required:
              type:
                - boolean
                - "null"
            recommended_next_workflow:
              type:
                - string
                - "null"
      summary:
        type: object
        required:
          - aggregate
        properties:
          findings:
            type:
              - string
              - "null"
          aggregate:
            type: string
            enum:
              - DEFECT
              - RISK
              - "KNOWLEDGE GAP"
              - PASS
---

# Review Workflow

## Purpose

Evaluate an existing artifact within an authorized scope using approved project knowledge, reusable engineering questions, available verification evidence, and engineering judgment.

Review identifies defects, risks, and knowledge gaps. It does not modify implementation or project knowledge, approve decisions, or resolve findings.

## When to Use

Use Review when objective verification alone cannot assess engineering quality, risk, or uncertainty. Use Verification for objective compliance, Implementation to change the target, and the owning knowledge process to make new decisions.

## Inputs

- Review target, authorized scope, and explicit constraints
- Repository, related artifacts, implementation, and established patterns
- Relevant Specifications and Domain knowledge
- `PROJECT.md`, `ARCHITECTURE.md`, `STACK.md`, and `CONVENTIONS.md`, as applicable
- Applicable ADRs and Artifact Contracts
- Applicable Design Knowledge for user-facing targets:
  - `PRODUCT_EXPERIENCE.md`
  - `UI_FOUNDATION.md`
  - `INTERACTION_PATTERNS.md`
- Applicable engineering review questions
- Available Verification Evidence

Reuse existing Verification Evidence instead of repeating equivalent checks. When relevant evidence is absent, report any material reduction in confidence; absence alone is not a defect.

## Execution

### 1. Establish Scope

Identify the target, scope, governing knowledge, applicable contracts and review questions, available evidence, and relevant implementation relationships. Exclude questions and knowledge outside scope.

### 2. Derive Questions

Create a bounded question set from reusable review questions, approved project and Design Knowledge, contracts, implementation relationships, and Verification Evidence. Questions may examine engineering consequences of approved knowledge but must not create requirements.

When the target is governed by a Specification with declared `Dependencies` (see [`SPECIFICATION_DEPENDENCIES.md`](../specifications/SPECIFICATION_DEPENDENCIES.md)), include each dependency's Target Presence and Target Lifecycle State among the facts available to Step 3.

### 3. Evaluate

Evaluate each applicable question using technical evidence and engineering reasoning. Explicitly distinguish observed facts, approved requirements, reasoning, preferences, and uncertainty.

### 4. Classify Outcomes

Assign one outcome to each applicable question:

- `PASS` — no material concern is identified.
- `DEFECT` — the artifact demonstrably violates approved behavior, project knowledge, or a required engineering invariant.
- `RISK` — a credible material failure mode or engineering weakness exists despite possible compliance with known requirements; style preference alone is insufficient.
- `KNOWLEDGE GAP` — missing, ambiguous, or conflicting authoritative knowledge prevents safe resolution.
- `NOT APPLICABLE` — the question does not apply.

Provide supporting reasoning for every non-trivial outcome.

A dependency whose Target Presence is `MISSING` is ordinarily a `KNOWLEDGE GAP` (a dangling reference) or `RISK`, as the evidence warrants; a dependency whose Target Lifecycle State is `DRAFT` is a `RISK` or `KNOWLEDGE GAP` only when the target's unresolved state is materially relevant to the question at hand — not by default. A dependency that is `PRESENT` and `APPROVED` raises no dependency-state concern by itself. Use the existing outcome vocabulary above; do not introduce a new classification for dependency concerns.

### 5. Produce Actionable Findings

Only `DEFECT`, `RISK`, and `KNOWLEDGE GAP` create actionable findings. Consolidate observations with one underlying issue unless they require independent resolution.

Each finding receives a stable report-local ID (`F-001`, `F-002`, ...), evidence, reasoning, impact, Resolution Owner, Decision Required status, and recommended next workflow.

### 6. Determine Resolution Path

Identify the artifact or implementation area that owns resolution, such as Implementation, Specification, Domain, Project, Architecture, ADR, Stack, Conventions, Design Knowledge, Contract, Workflow, Evaluation, or Documentation.

Set `Decision Required: Yes` when resolution requires new or changed authoritative knowledge. Review may describe the required resolution category but must not design or approve the solution.

## Outputs

**Engineering Review Report** — one block per finding:

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

**Review Summary** — one block per run, mechanically derived from the findings above:

```text
Review Summary

Findings: [count by classification]
Aggregate: DEFECT | RISK | KNOWLEDGE GAP | PASS
```

`Aggregate` is `DEFECT` if any finding is `DEFECT`; otherwise `RISK` if any finding is `RISK`; otherwise `KNOWLEDGE GAP` if any finding is `KNOWLEDGE GAP`; otherwise `PASS`. It summarizes the findings above and is not itself a new judgment or a lifecycle state.

Also report scope, passes where useful, remaining uncertainty, and the effect of missing evidence.

## Rules

- Remain within scope and trace conclusions to the artifact, approved knowledge, and evidence.
- Do not redefine knowledge, modify implementation, resolve findings, approve decisions, or provide detailed implementation solutions.
- Do not treat preference as requirement or lack of verification as defect.
- Report material uncertainty and never silently assume undefined behavior.
- Classify a material missing-decision issue as `KNOWLEDGE GAP` and identify its owner.
- Recommendations require authorization before resolution.

## Failure Handling

### Missing Verification Evidence

Continue where evidence is sufficient, identify confidence limits, and recommend Verification when objective evidence could resolve them.

### Missing or Conflicting Knowledge

Identify the owning or conflicting sources and apply established ownership or precedence rules. If the issue remains material, produce a `KNOWLEDGE GAP`, name the Resolution Owner, mark whether a decision is required, and set Recommended Next Workflow to [`workflows/knowledge-resolution.md`](knowledge-resolution.md) when a human decision is expected to resolve it.

### Insufficient Technical Evidence

Identify missing evidence, report uncertainty, avoid fabricated conclusions, and recommend Verification when appropriate.

## Completion Criteria

Review is complete when all applicable questions have been evaluated; Verification Evidence has been reused where relevant; actionable findings have stable IDs, owners, decision status, and next workflows; dependency-state concerns have been classified using the existing outcome vocabulary rather than a new one; the reported `Aggregate` accurately reflects the findings; material uncertainty is reported; and neither implementation nor project knowledge has been changed.

Verification asks whether objective evidence proves compliance. Review asks whether the artifact exposes a material defect, risk, or knowledge gap. Resolution begins only after authorization and remains a separate responsibility.
