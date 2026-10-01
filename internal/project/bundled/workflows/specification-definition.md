---
identity: specification-definition
specification_reference: required
requires_approved_specification: false
result:
  terminal_path: outcome
  classification:
    READY_FOR_APPROVAL: success
    BLOCKED: blocked
  schema:
    type: object
    required:
      - outcome
    properties:
      outcome:
        type: string
        enum:
          - READY_FOR_APPROVAL
          - BLOCKED
      target:
        type:
          - string
          - "null"
      remaining_unresolved:
        type:
          - string
          - "null"
      non_blocking_deferred_items:
        type:
          - string
          - "null"
      consistency_check:
        type:
          - string
          - "null"
      resolved_this_run:
        type:
          - string
          - "null"
---

# Specification Definition

Bring a Draft Specification to text the user can approve: its goal, the decisions that change what gets built (each answered by the user), acceptance criteria someone can check, and what is out of scope. You never approve it — only the user can, with `gnomon approve`.

## Steps

1. **Read** the Specification as it is now, the project knowledge files listed in your prompt that bear on it, and the code it would touch.
2. **Find the open decisions**: questions where reasonable answers lead to different behavior — who may do it, edge and failure cases, limits, what happens to existing data or state. Skip anything already decided in `.gnomon/` or by the existing code, and anything that doesn't change behavior.
3. **Ask the user** each open decision directly in this session, a few at a time, with your recommended answer when you have one. Never record your own guess as the user's answer. A decision the user defers stays open.
4. **Write the Specification** in its existing structure (whichever template it came from). Record each decision once, in its Decisions section (add one if missing), as the question and the user's answer. Do not copy decisions into other files. If a project-wide decision is already recorded in a knowledge file, refer to it instead of restating it.
5. **Write acceptance criteria**: each one observable behavior that can be checked, together covering the decisions. Remove template placeholders and sections that don't apply. Keep the whole Specification short enough to review in a few minutes.
6. **Check consistency** with the knowledge files and other Specifications. Report contradictions instead of silently choosing. If another Specification seems to need changing, report it; don't edit it.

## Result

- `outcome` — `READY_FOR_APPROVAL` when no open decision blocks the behavior and every criterion is checkable; otherwise `BLOCKED`.
- `target` — the Specification and its path.
- `resolved_this_run` — what you wrote, and each decision recorded with the user's answer.
- `remaining_unresolved` — `BLOCKED` only: the open decisions or contradictions that block approval.
- `non_blocking_deferred_items` — anything left open that doesn't block approval.
- `consistency_check` — `CLEAN`, or the contradictions found.
