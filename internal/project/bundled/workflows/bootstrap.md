---
identity: bootstrap
specification_reference: none
requires_approved_specification: false
result:
  terminal_path: outcome
  classification:
    BOOTSTRAP_COMPLETE: success
    BLOCKED: blocked
  schema:
    type: object
    required:
      - outcome
    properties:
      outcome:
        type: string
        enum:
          - BOOTSTRAP_COMPLETE
          - BLOCKED
      baseline_established:
        type:
          - string
          - "null"
      validation_evidence:
        type:
          - string
          - "null"
      remaining_unresolved:
        type:
          - string
          - "null"
---

# Bootstrap

Make sure the project can be built and its tests run from a clean checkout, so later implementation can be checked.

## Steps

1. **Inspect** the repository and the project knowledge files listed in your prompt.
2. **If a working baseline exists** (it builds, and at least one test runs), confirm it by running it and report.
3. **Otherwise establish the smallest baseline** consistent with recorded decisions. Ask the user about any technology choice that isn't already decided in `.gnomon/` or by the existing code; don't pick one silently.
4. **Run** the build and tests and record the results.

## Result

- `outcome` — `BOOTSTRAP_COMPLETE` when the project builds and its tests run; otherwise `BLOCKED`.
- `baseline_established` — what exists now (build, test command, anything you added).
- `validation_evidence` — the commands you ran and their results.
- `remaining_unresolved` — `BLOCKED` only: what prevents a working baseline.

## Rules

- Preserve existing work; ask before deleting or replacing anything you didn't create.
