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
            test:
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

1. **Pick the criteria.** The prompt states this run's scope. For a standard check of a change: the criteria of the Specifications it names as changed, plus those of every Approved Specification whose behavior the changed files implement or test — say which in `summary.obligations_evaluated`, and why; leave out Specifications the change cannot affect. For a target (for example `SPEC-003`, or a path): its criteria, or those of the Specifications that govern the path. For a full check, or when there is no change to limit it to: every Approved Specification's criteria (or the target's). One entry per criterion, in order.
2. **Gather evidence** for each criterion that someone else could re-check.
   - *Standard check:* run the project's test suite once. A passing test is enough evidence for a criterion only when you have read it and it asserts that criterion's own observable outcome — not a neighbouring case, a weaker condition, or a value different from the criterion's. Then name it in `test` (`path::name`) and say in `evidence` what it asserts. Otherwise check the criterion directly (run it; for what cannot be run, the exact code that implements it), or mark it `UNVERIFIABLE`. If a test fails, include the criterion it asserts, even outside the scope.
   - *Full check:* check every criterion directly. A passing test alone is not enough evidence.
3. **Check each criterion for conflicts** with the other Approved Specifications (`gnomon status` lists each one's state). A conflict is another Approved Specification's acceptance criterion or recorded decision that requires something about the same behavior that cannot be true together with this criterion — for example "exactly seven fields" here and a ninth field required there. A note in a later Specification saying it extends, overrides or is compatible with this criterion does not resolve the conflict while this criterion's own approved text still says otherwise. Not a conflict: another Specification adding behavior this criterion does not address, or an Out of Scope or deferred note, which requires nothing.
4. **Decide each result.** `CONFLICT` when step 3 found a conflict, whatever the code does — meeting one of two contradictory requirements is not a pass. Otherwise `PASS` only when your evidence shows the criterion is met, `FAIL` when it shows it is not, and `UNVERIFIABLE` when you could not get evidence either way. Gnomon reports any `PASS` without evidence as unverified, and any criterion with a conflict as `CONFLICT`.
5. **Recommend a next step** for each non-passing criterion: `implementation` (behavior missing or wrong), `testing` (a test is needed to tell), `knowledge-resolution` (the criterion is ambiguous, contradicts recorded knowledge, or is in conflict — always for `CONFLICT`), or `null`.

## Result

- `evidence` — one item per criterion: `obligation` (the criterion text), `source` (for example `SPEC-003, criterion 2`), `result`, `evidence` (what you ran or looked at and what it showed), `conflicts_with` (for `CONFLICT`: the conflicting statement's source and wording, for example `SPEC-007, criterion 1: "nine fields"`; otherwise `null`), `test` (the passing test used as evidence, `path::name`, otherwise `null`; Gnomon reports a criterion whose cited test does not exist as unverified), `recommended_workflow`.
- `summary.obligations_evaluated` — what was in scope.
- `summary.aggregate` — `FAIL` if any criterion failed or is in conflict; otherwise `UNVERIFIABLE` if any is unverifiable or none was evaluated; otherwise `PASS`.

## Rules

- Report what the evidence shows, not engineering opinions — that is Review's job.
- Don't modify code, tests, Specifications, or knowledge.
