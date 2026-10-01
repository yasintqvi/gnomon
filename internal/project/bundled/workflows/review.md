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

# Review

Look for problems in the target that its acceptance criteria and tests would not catch: defects, risks, and behavior nobody has decided. Don't change anything.

## Steps

1. **Read** the target, the Specifications that govern it, the project knowledge files listed in your prompt that apply, and the related tests. Reuse any verification evidence you are given.
2. **Look for what could cause wrong behavior, data loss, security exposure, or costly rework.** Report only findings you can point to in the code or text; skip style preferences.
3. **Classify each finding**:
   - `DEFECT` — the target contradicts its Specification, recorded knowledge, or plain correctness.
   - `RISK` — not wrong today, but likely to cause harm (for example missing failure handling, unsafe concurrency, an easy-to-break assumption).
   - `KNOWLEDGE GAP` — the right behavior isn't decided anywhere; set `decision_required` when the user must decide.
4. **Recommend a next step** for each: `implementation`, `testing`, `knowledge-resolution`, or `null` — chosen from what caused the finding, not from its class alone.

## Result

- `findings` — each with `finding_id` (`F-001`, `F-002`, …), `classification`, `summary`, `evidence` (file and line, or quoted text), `engineering_reasoning`, `impact`, `resolution_owner` (who must act: implementation, the Specification, or a knowledge area), `decision_required`, `recommended_workflow`.
- `summary.findings` — a one-line overview.
- `summary.aggregate` — `DEFECT` if any finding is a defect; otherwise `RISK` if any is a risk; otherwise `KNOWLEDGE GAP` if any is a gap; otherwise `PASS`.
