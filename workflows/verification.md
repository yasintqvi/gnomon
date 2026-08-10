# Verification Workflow

## Purpose

Define the execution process for the **Verification** intent.

Verification determines whether a target artifact objectively complies with applicable approved project knowledge and verification criteria.

The workflow produces reproducible evidence only.

It does not perform engineering review, recommend improvements, make design decisions, or introduce new project knowledge.

---

## Intent

Verify a target artifact against applicable approved project knowledge using reusable verification criteria and objective evidence.

---

## When to Use

Use this workflow when:

* An implementation or other project artifact requires objective compliance verification.
* Evidence is required before review, approval, or completion.
* An approved change must be checked against its governing project knowledge.

Do not use this workflow when:

* Engineering quality or improvement opportunities must be evaluated.
* New requirements or architectural decisions must be made.
* The task is to implement or modify the target.
* The expected behavior is not defined by approved project knowledge.

---

## Inputs

### User Input

* Verification target
* Verification scope
* Explicit constraints

### Project Context

* Repository
* Existing implementation
* Related artifacts
* Available tests, build output, and analysis results

### Required Knowledge

As applicable to the verification target:

* Relevant Specifications
* Relevant Domain knowledge
* `PROJECT.md`
* `ARCHITECTURE.md`
* `STACK.md`
* `CONVENTIONS.md`
* Applicable ADRs
* Applicable Artifact Contracts

### Evaluation Assets

* Applicable verification criteria

---

## Execution

### 1. Understand

#### Purpose

Determine exactly what must be verified and which approved knowledge governs the target.

#### Expected Result

Identify:

* Verification target
* Authorized verification scope
* Applicable project knowledge
* Applicable Artifact Contracts
* Applicable verification criteria
* Available evidence sources

Criteria and knowledge outside the authorized scope must not affect the result.

---

### 2. Derive Verification Obligations

#### Purpose

Determine the objectively verifiable obligations that apply to the target.

#### Sources

Verification obligations may originate from:

* Acceptance criteria in the relevant Specification
* Project-level requirements
* Architectural constraints
* Stack constraints
* Project conventions
* Applicable ADRs
* Applicable Artifact Contracts
* Reusable verification criteria

#### Expected Result

A bounded set of verification obligations where each obligation is traceable to an approved source.

Verification must not create an obligation that cannot be traced to approved project knowledge or applicable verification criteria.

---

### 3. Gather Evidence

#### Purpose

Collect objective evidence sufficient to evaluate each applicable obligation.

Evidence may include:

* Implementation artifacts
* Configuration
* Documentation
* Automated tests
* Static analysis
* Build output
* Runtime behavior
* Persisted state
* Other reproducible technical evidence appropriate to the obligation

Evidence must be derived from observable artifacts or reproducible execution.

Assumptions are not evidence.

---

### 4. Verify

#### Purpose

Evaluate every applicable verification obligation against the gathered evidence.

Each obligation receives exactly one result:

* `PASS`
* `FAIL`
* `UNVERIFIABLE`

#### PASS

The available objective evidence demonstrates that the obligation is satisfied.

#### FAIL

The available objective evidence demonstrates that the obligation is not satisfied.

#### UNVERIFIABLE

The available evidence is insufficient to determine compliance objectively.

Lack of evidence must not be converted into either `PASS` or `FAIL`.

---

### 5. Produce Evidence

For every evaluated obligation, record:

* Obligation
* Source
* Result
* Supporting evidence

Where useful, evidence should include:

* Artifact path
* Relevant location
* Executed verification command or mechanism
* Relevant test or analysis result

The output must distinguish observed evidence from any contextual explanation.

---

## Outputs

### Primary Output

**Verification Evidence**

A reproducible record of every applicable verification obligation and its result.

### Result Model

```text
Obligation:
Source:
Result: PASS | FAIL | UNVERIFIABLE
Evidence:
```

Verification does not classify findings as defects, risks, knowledge gaps, or improvement opportunities.

Those classifications belong to Engineering Review.

---

## Rules

* Verification must be objective.
* Verification must be reproducible.
* Verification must remain within the authorized scope.
* Every verification obligation must be traceable to an approved source.
* Every result must be supported by objective evidence.
* Verification must distinguish evidence from assumptions.
* Absence of evidence must result in `UNVERIFIABLE` when compliance cannot be determined.
* Verification must not introduce project knowledge.
* Verification must not modify the target artifact.
* Verification must not perform engineering review.
* Verification must not recommend implementation improvements.
* Verification must not weaken or reinterpret an approved requirement to obtain a passing result.
* Existing tests are evidence, not automatically proof of complete compliance.
* A passing test suite does not replace verification of other applicable obligations.

---

## Failure Handling

### Missing Evidence

When an obligation cannot be evaluated from available evidence:

1. Identify the missing evidence.
2. Record the obligation as `UNVERIFIABLE`.
3. Do not infer compliance.

---

### Missing Project Knowledge

When verification requires behavior that is not defined by approved project knowledge:

1. Do not invent the expected behavior.
2. Do not create a new verification obligation.
3. Report that the expected behavior is not defined when relevant to the requested scope.

Resolution of missing project knowledge occurs outside Verification.

---

### Conflicting Project Knowledge

When applicable approved sources conflict:

1. Identify the conflicting sources.
2. Apply established project knowledge ownership or precedence rules when those rules resolve the conflict objectively.
3. If the expected obligation remains ambiguous, mark it `UNVERIFIABLE`.
4. Do not choose an interpretation through engineering judgment.

---

### Verification Execution Failure

When a verification mechanism cannot execute successfully:

1. Record the execution failure as evidence.
2. Determine whether another objective mechanism can verify the same obligation.
3. If no sufficient evidence can be obtained, mark the obligation `UNVERIFIABLE`.
4. Do not treat verification-tool failure itself as proof that the implementation violates the requirement.

---

## Completion Criteria

Verification completes when:

* The verification target and scope are defined.
* Applicable project knowledge has been identified.
* Applicable verification obligations have been derived.
* Every applicable obligation has been evaluated.
* Every result is supported by objective evidence.
* Unverifiable obligations are explicitly reported.
* No engineering conclusions or new project knowledge have been introduced.

---

## Workflow Constraints

* Verification is read-only with respect to the target implementation and authoritative project knowledge.
* Verification evaluates compliance; it does not repair non-compliance.
* Verification may execute tests, builds, analysis, or other non-destructive verification mechanisms.
* Verification must not expand its scope merely because additional issues are discovered.
* Verification Evidence must remain suitable for consumption by a later Engineering Review.

---

## Notes

Verification answers:

> Does objective evidence demonstrate that this artifact complies with its applicable approved obligations?

It does not answer:

> Is this implementation well designed?

> What should be improved?

> Is an undefined requirement missing?

Those questions belong to Engineering Review or the project-knowledge workflow that owns the relevant decision.
