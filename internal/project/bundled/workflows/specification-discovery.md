---
identity: specification-discovery
specification_reference: none
requires_approved_specification: false
result:
  terminal_path: outcome
  classification:
    CANDIDATE_PROPOSED: success
    NO_CANDIDATE_IDENTIFIED: success
    BLOCKED: blocked
  schema:
    type: object
    required:
      - outcome
    properties:
      outcome:
        type: string
        enum:
          - CANDIDATE_PROPOSED
          - NO_CANDIDATE_IDENTIFIED
          - BLOCKED
      candidate_title:
        type:
          - string
          - "null"
      candidate_identity:
        type:
          - string
          - "null"
      rationale:
        type:
          - string
          - "null"
      derived_from:
        type:
          - string
          - "null"
      non_blocking_caveat:
        type:
          - string
          - "null"
      remaining_knowledge_gap:
        type:
          - string
          - "null"
---

# Specification Discovery

Propose one next Specification worth writing, based on recorded project knowledge and the existing Specifications. Don't create it — Gnomon creates it only if the user accepts.

## Steps

1. **Read** the project knowledge files listed in your prompt, the existing Specifications in `.gnomon/specifications/`, and the repository.
2. **Find one capability** the recorded knowledge calls for that no existing Specification covers, small enough to define and implement on its own. Don't invent product behavior the knowledge doesn't support.

## Result

- `outcome` — `CANDIDATE_PROPOSED`; `NO_CANDIDATE_IDENTIFIED` when none can be found without inventing behavior (this says nothing about whether the set is complete); `BLOCKED` when recorded knowledge is too thin to judge.
- `candidate_title` — a short title (`CANDIDATE_PROPOSED` only).
- `candidate_identity` — a short lowercase phrase Gnomon uses for the filename, for example `password-reset` (`CANDIDATE_PROPOSED` only).
- `rationale` — why this one, and why now.
- `derived_from` — the knowledge it is based on, for example `PROJECT.md, Purpose`.
- `non_blocking_caveat` — optional: anything the user should know.
- `remaining_knowledge_gap` — `BLOCKED` only: what knowledge is missing.
