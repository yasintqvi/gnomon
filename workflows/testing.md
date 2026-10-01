---
identity: testing
specification_reference: optional
requires_approved_specification: true
result:
  terminal_path: outcome
  classification:
    TESTING_COMPLETE: success
    BLOCKED: blocked
  schema:
    type: object
    required:
      - outcome
    properties:
      outcome:
        type: string
        enum:
          - TESTING_COMPLETE
          - BLOCKED
      evidence:
        type:
          - string
          - "null"
      coverage:
        type:
          - string
          - "null"
      remaining_unresolved:
        type:
          - string
          - "null"
---

# Testing

Write or run tests: for the acceptance criteria of the Specification named in your prompt, or, with no Specification, for the area the user asks about. Gnomon has already confirmed a named Specification is Approved.

## Steps

1. **Read** the Specification (if any), the relevant code, and its existing tests.
2. **Find the gaps**: which criteria (or requested behaviors) have no test that would fail if the behavior broke.
3. **Write the smallest tests** that check behavior through the project's existing test setup. Ask the user if the expected behavior is not decided anywhere.
4. **Run** them and the related existing tests.

## Result

- `outcome` — `TESTING_COMPLETE` when tests were written or run and their results are reported, including failures; `BLOCKED` when tests cannot be run or a needed decision is missing.
- `evidence` — tests added or run, with commands and results.
- `coverage` — which criteria are covered by a test and which are not.
- `remaining_unresolved` — `BLOCKED` only: what prevents testing.

## Rules

- Don't change product code to make a test pass; report the failure.
- Never weaken an assertion to get a passing result.
