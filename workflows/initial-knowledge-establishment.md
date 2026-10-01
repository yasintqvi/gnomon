---
identity: initial-knowledge-establishment
specification_reference: none
requires_approved_specification: false
result:
  terminal_path: outcome
  classification:
    READY_FOR_BOOTSTRAP: success
    BLOCKED: blocked
  schema:
    type: object
    required:
      - outcome
    properties:
      outcome:
        type: string
        enum:
          - READY_FOR_BOOTSTRAP
          - BLOCKED
      recorded_knowledge:
        type:
          - string
          - "null"
      non_blocking_deferred_items:
        type:
          - string
          - "null"
      remaining_unresolved:
        type:
          - string
          - "null"
---

# Describe the Project

Record the few project facts an agent needs that the repository itself doesn't show: what the project is for, who uses it, hard constraints, and project-wide decisions. This step is optional; the normal path is defining and implementing Specifications.

## Steps

1. **Inspect** the repository (README, code, configuration), the project knowledge files listed in your prompt, and any agent instruction files such as `AGENTS.md` or `CLAUDE.md`. Don't duplicate what those already say.
2. **Ask the user** only for facts you cannot determine and that matter for building features: purpose, users, hard constraints, project-wide decisions. If you recall project facts from your own memory or earlier conversations, confirm them with the user before recording them.
3. **Record** them in `.gnomon/context/PROJECT.md` (create it if absent; update the file that already holds a fact rather than adding a second copy). Keep it short — under about 100 lines — with sections such as Purpose, Users, Constraints, and Project-wide decisions (each as question — answer). Leave out what the code already shows (stack, layout, commands) unless it records a decision the code can't reveal.

## Result

- `outcome` — `READY_FOR_BOOTSTRAP` when the essential facts are recorded; `BLOCKED` if even the project's purpose cannot be established.
- `recorded_knowledge` — what you recorded, and where.
- `non_blocking_deferred_items` — facts left open that don't block work.
- `remaining_unresolved` — `BLOCKED` only: what is missing.
