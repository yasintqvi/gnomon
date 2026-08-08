# Verification Workflow

## Purpose

Define the execution process for the Verification intent.

Verification determines whether an artifact objectively complies with approved project knowledge.

The workflow produces evidence only.

It does not perform engineering reasoning, identify improvement opportunities, or introduce new project knowledge.

---

## Intent

Verify an artifact against approved project knowledge using reusable verification criteria.

---

## Inputs

### User Input

- Verification target
- Verification scope
- Constraints

### Project Context

- Repository
- Related artifacts

### Required Knowledge

- Relevant Specifications
- DOMAIN
- ARCHITECTURE
- STACK
- CONVENTIONS
- ADRs
- Contracts

### Evaluation Assets

VERIFICATION_CRITERIA.md

---

## Execution

### 1. Understand

Determine:

- Verification target
- Verification scope
- Applicable project knowledge

---

### 2. Load Criteria

Load applicable verification criteria.

Ignore criteria outside the authorized scope.

---

### 3. Verify

Verify every applicable criterion using objective evidence.

Evidence may include:

- Implementation
- Documentation
- Tests
- Static analysis
- Build output

No engineering judgment is allowed.

---

### 4. Produce Evidence

For every criterion produce:

- PASS
- FAIL
- UNVERIFIABLE

Supporting evidence is mandatory.

---

## Outputs

Primary Output

Verification Evidence

---

## Rules

- Verification must be objective.
- Verification must be reproducible.
- Verification must not introduce project knowledge.
- Verification must not perform engineering reasoning.
- Verification must distinguish evidence from assumptions.

---

## Completion Criteria

Verification completes when every applicable criterion has been evaluated.