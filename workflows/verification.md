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
                - CONFLICT
                - UNVERIFIABLE
            evidence:
              type:
                - string
                - "null"
            conflicts_with:
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

# Verification

Check, criterion by criterion, whether the target does what was approved, and show the evidence for each result. Don't change anything.

## Steps

1. **Pick the criteria.** If the target is a Specification (for example `SPEC-003`), use its acceptance criteria, one entry per criterion, in order. For a path or area, use the acceptance criteria of the Specifications that govern it, and say which. With no target, ask the user what to verify.
2. **Gather evidence** for each criterion that someone else could re-check: a test or command you ran and its result, or the exact code location that implements it. Prefer running things over reading code.
3. **Check each criterion for conflicts** with the other Approved Specifications (`gnomon status` lists each one's state). A conflict is another Approved Specification's acceptance criterion or recorded decision that requires something about the same behavior that cannot be true together with this criterion — for example "exactly seven fields" here and a ninth field required there. A note in a later Specification saying it extends, overrides or is compatible with this criterion does not resolve the conflict while this criterion's own approved text still says otherwise. Not a conflict: another Specification adding behavior this criterion does not address, or an Out of Scope or deferred note, which requires nothing.
4. **Decide each result.** `CONFLICT` when step 3 found a conflict, whatever the code does — meeting one of two contradictory requirements is not a pass. Otherwise `PASS` only when your evidence shows the criterion is met, `FAIL` when it shows it is not, and `UNVERIFIABLE` when you could not get evidence either way. Gnomon reports any `PASS` without evidence as unverified, and any criterion with a conflict as `CONFLICT`.
5. **Recommend a next step** for each non-passing criterion: `implementation` (behavior missing or wrong), `testing` (a test is needed to tell), `knowledge-resolution` (the criterion is ambiguous, contradicts recorded knowledge, or is in conflict — always for `CONFLICT`), or `null`.

## Result

- `evidence` — one item per criterion: `obligation` (the criterion text), `source` (for example `SPEC-003, criterion 2`), `result`, `evidence` (what you ran or looked at and what it showed), `conflicts_with` (for `CONFLICT`: the conflicting statement's source and wording, for example `SPEC-007, criterion 1: "nine fields"`; otherwise `null`), `recommended_workflow`.
- `summary.obligations_evaluated` — what was in scope.
- `summary.aggregate` — `FAIL` if any criterion failed or is in conflict; otherwise `UNVERIFIABLE` if any is unverifiable or none was evaluated; otherwise `PASS`.

## Rules

- Report what the evidence shows, not engineering opinions — that is Review's job.
- Don't modify code, tests, Specifications, or knowledge.
