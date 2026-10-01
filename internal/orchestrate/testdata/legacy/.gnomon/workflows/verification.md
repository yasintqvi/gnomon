---
identity: verification
specification_reference: none
requires_approved_specification: false
result:
  terminal_path: summary.aggregate
  classification:
    PASS: success
    FAIL: blocked
    UNVERIFIABLE: blocked
  schema:
    type: object
    required:
      - summary
    properties:
      evidence:
        type:
          - array
          - "null"
        items:
          type: object
          required:
            - obligation
            - result
          properties:
            obligation:
              type: string
            source:
              type:
                - string
                - "null"
            result:
              type: string
              enum:
                - PASS
                - FAIL
                - UNVERIFIABLE
            evidence:
              type:
                - string
                - "null"
            recommended_workflow:
              type:
                - string
                - "null"
              enum:
                - implementation
                - testing
                - knowledge-resolution
                - null
      summary:
        type: object
        required:
          - aggregate
        properties:
          obligations_evaluated:
            type:
              - string
              - "null"
          aggregate:
            type: string
            enum:
              - PASS
              - FAIL
              - UNVERIFIABLE
---

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

For an obligation traceable to a Specification with declared `Dependencies` (see [`SPECIFICATION_DEPENDENCIES.md`](../specifications/SPECIFICATION_DEPENDENCIES.md)), note each dependency's Target Presence and Target Lifecycle State; carry this into Step 4.

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

Mark an obligation UNVERIFIABLE, rather than PASS or FAIL, when it depends on a dependency whose Target Presence is `MISSING` or whose Target Lifecycle State is `DRAFT` — that target's behavior is not yet authoritative evidence. Evaluate the obligation normally when the dependency is `PRESENT` and `APPROVED`. Do not block the entire run over one affected obligation; evaluate every other applicable obligation normally.

### 5. Report Evidence

Record each obligation, authoritative source, result, and supporting evidence. Include artifact locations, commands, tests, runtime observations, or analysis results where useful, and distinguish observed evidence from contextual explanation.

For each obligation whose result is `FAIL` or `UNVERIFIABLE`, also set Recommended Workflow from the same fixed vocabulary Review uses (`implementation`, `testing`, `knowledge-resolution`).

The result alone never determines this. `FAIL` and `UNVERIFIABLE` are each compatible with more than one cause, and two obligations sharing the same result may legitimately warrant different recommendations — a `FAIL` may exist because the target itself does not do what is required, or because the mechanism that produced the failing evidence is itself broken or measuring the wrong thing; an `UNVERIFIABLE` may exist because the evidence needed simply was not gathered, or because the obligation's own authoritative basis is not yet settled. Determine the recommendation from the obligation's actual cause, reasoned from the evidence already gathered in Step 3, not from which of these two results was assigned:

* `implementation` — the evidence itself is trustworthy and shows the target does not do what the obligation requires.
* `testing` — the obligation's own verification mechanism is missing, inadequate, or could not be exercised, independent of whether the target itself is actually correct.
* `knowledge-resolution` — the obligation's own authoritative basis is missing, conflicting, or not yet Approved, so the obligation itself cannot yet be confirmed against anything settled — including when the cause does not cleanly fit either of the other two but still ultimately traces back to unsettled authoritative knowledge.

This is still an evidentiary determination of *why* the result is what it is, never a broader engineering assessment, a design suggestion, or a decision about what the obligation itself should require — Review, not Verification, evaluates engineering quality and risk. Leave Recommended Workflow unset for `PASS`, and unset for any `FAIL`/`UNVERIFIABLE` obligation whose cause genuinely fits none of the three.

## Outputs

**Verification Evidence** — one block per obligation:

```
Obligation:
Source:
Result: PASS | FAIL | UNVERIFIABLE
Evidence:
Recommended Workflow: implementation | testing | knowledge-resolution (FAIL/UNVERIFIABLE only)
```

**Verification Summary** — one block per run, mechanically derived from the per-obligation results above:

```
Verification Summary

Obligations evaluated:
Aggregate: PASS | FAIL | UNVERIFIABLE
```

`Aggregate` is `FAIL` if any obligation is `FAIL`; otherwise `UNVERIFIABLE` if any obligation is `UNVERIFIABLE`; otherwise `PASS`. It summarizes the per-obligation results above and is not itself a new judgment — Verification still does not classify defects, risks, knowledge gaps, or improvement opportunities, and `Aggregate` is not a lifecycle state or an approval decision; those belong to Review and Human Approval respectively.

## Rules

* Remain objective, reproducible, traceable, and within authorized scope.
* Remain read-only with respect to the target and authoritative project knowledge; non-destructive checks may execute.
* Support every result with objective evidence appropriate to the obligation.
* Do not introduce knowledge, reinterpret requirements, repair non-compliance, perform review, or recommend improvements.
* Treat existing tests as evidence, not automatic proof of complete compliance.
* Do not infer runtime or rendered compliance solely from source structure when the obligation depends materially on execution.
* Do not expand scope merely because additional issues are discovered.
* Keep evidence suitable for later Review.
* Determine Recommended Workflow from the obligation's actual cause and the evidence gathered for it — never mechanically from whether the result is `FAIL` or `UNVERIFIABLE` alone (for example, never assume `FAIL` always means `implementation` or `UNVERIFIABLE` always means `testing`), and never as a broader engineering recommendation, a design suggestion, or an assessment of quality or risk; that remains Review's responsibility, not Verification's.

## Failure Handling

### Missing Evidence

Identify what is missing and mark the affected obligation UNVERIFIABLE.

### Dependency Not Yet Available

A dependency's Target Presence is `MISSING`, or its Target Lifecycle State is `DRAFT`, for an obligation traceable to it. Mark only that obligation UNVERIFIABLE and continue evaluating every other applicable obligation normally.

### Missing or Conflicting Knowledge

Do not invent an expectation. Apply established ownership or precedence rules only when they resolve the issue objectively; otherwise report the undefined expectation or mark the affected obligation UNVERIFIABLE. Once a human decision resolves the gap, run [`workflows/knowledge-resolution.md`](knowledge-resolution.md) to update the affected authoritative knowledge before re-verifying the affected obligation.

### Execution Failure

Record the failure, try another objective mechanism when appropriate, and mark the obligation UNVERIFIABLE if sufficient evidence remains unavailable. Tool failure alone is not proof of non-compliance.

## Completion Criteria

Verification is complete when scope and governing knowledge are identified, every applicable traceable obligation has been evaluated using evidence appropriate to that obligation, obligations affected by a `MISSING` or `DRAFT` dependency are marked UNVERIFIABLE rather than blocking the run, unverifiable obligations are explicit, every `FAIL`/`UNVERIFIABLE` obligation carries a Recommended Workflow from the fixed vocabulary determined from its actual cause rather than its result type (or is left unset when none of the three genuinely fits), the reported `Aggregate` accurately reflects the per-obligation results, and no engineering conclusions or new project knowledge have been introduced.
