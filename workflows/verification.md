# Verification Workflow

## Purpose

Objectively verify a target artifact against applicable approved project knowledge and reusable verification criteria, producing reproducible evidence without modifying the target or applying engineering judgment.

## When to Use

Use this workflow when objective compliance evidence is required before review, approval, or completion. Do not use it to implement fixes, assess broader engineering quality, recommend improvements, or decide undefined requirements.

## Inputs

* Verification target, authorized scope, and explicit constraints
* Repository, related artifacts, and available tests or analysis results
* Relevant Specifications and Domain knowledge
* PROJECT.md, ARCHITECTURE.md, STACK.md, and CONVENTIONS.md, as applicable
* Applicable ADRs and Artifact Contracts
* Applicable Design Knowledge for user-facing targets:

  * PRODUCT_EXPERIENCE.md
  * UI_FOUNDATION.md
  * INTERACTION_PATTERNS.md
* Applicable verification criteria

## Execution

### 1. Establish Scope

Identify the target, authorized scope, governing knowledge, applicable contracts and criteria, and available evidence sources. Exclude knowledge and criteria outside scope.

### 2. Derive Obligations

Create a bounded set of objectively verifiable obligations traceable to approved sources: acceptance criteria, project or Domain requirements, Architecture, Stack, Conventions, ADRs, Contracts, applicable Design Knowledge, or reusable verification criteria.

Do not create an obligation from an assumption or undefined expectation.

### 3. Gather Evidence

Collect observable and reproducible evidence sufficient to evaluate each obligation.

Use the evidence mechanism required to actually prove the obligation. Source inspection alone is insufficient when compliance materially depends on runtime, rendered, interactive, persisted, or otherwise executable behavior.

Evidence may include artifacts, configuration, documentation, tests, static analysis, builds, runtime behavior, or persisted state.

Assumptions are not evidence, and the existence of implementation intended to satisfy an obligation is not itself proof that the obligation is satisfied.

### 4. Evaluate

Assign exactly one result to every applicable obligation:

* PASS — evidence demonstrates compliance.
* FAIL — evidence demonstrates non-compliance.
* UNVERIFIABLE — available evidence cannot determine compliance.

Do not convert absent or insufficient evidence into PASS or FAIL.

### 5. Report Evidence

Record each obligation, authoritative source, result, and supporting evidence. Include artifact locations, commands, tests, runtime observations, or analysis results where useful, and distinguish observed evidence from contextual explanation.

## Output

**Verification Evidence**

```
Obligation:
Source:
Result: PASS | FAIL | UNVERIFIABLE
Evidence:
```

Verification does not classify defects, risks, knowledge gaps, or improvement opportunities; those belong to Review.

## Rules

* Remain objective, reproducible, traceable, and within authorized scope.
* Remain read-only with respect to the target and authoritative project knowledge; non-destructive checks may execute.
* Support every result with objective evidence appropriate to the obligation.
* Do not introduce knowledge, reinterpret requirements, repair non-compliance, perform review, or recommend improvements.
* Treat existing tests as evidence, not automatic proof of complete compliance.
* Do not infer runtime or rendered compliance solely from source structure when the obligation depends materially on execution.
* Do not expand scope merely because additional issues are discovered.
* Keep evidence suitable for later Review.

## Failure Handling

### Missing Evidence

Identify what is missing and mark the affected obligation UNVERIFIABLE.

### Missing or Conflicting Knowledge

Do not invent an expectation. Apply established ownership or precedence rules only when they resolve the issue objectively; otherwise report the undefined expectation or mark the affected obligation UNVERIFIABLE. Once a human decision resolves the gap, run [`workflows/knowledge-resolution.md`](knowledge-resolution.md) to update the affected authoritative knowledge before re-verifying the affected obligation.

### Execution Failure

Record the failure, try another objective mechanism when appropriate, and mark the obligation UNVERIFIABLE if sufficient evidence remains unavailable. Tool failure alone is not proof of non-compliance.

## Completion Criteria

Verification is complete when scope and governing knowledge are identified, every applicable traceable obligation has been evaluated using evidence appropriate to that obligation, unverifiable obligations are explicit, and no engineering conclusions or new project knowledge have been introduced.
