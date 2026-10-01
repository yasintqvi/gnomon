---
identity: knowledge-resolution
specification_reference: none
requires_approved_specification: false
result:
  terminal_path: outcome
  classification:
    RESOLVED: success
    BLOCKED: blocked
  schema:
    type: object
    required:
      - outcome
    properties:
      outcome:
        type: string
        enum:
          - RESOLVED
          - BLOCKED
      updated_knowledge:
        type:
          - string
          - "null"
      adr_created:
        type:
          - string
          - "null"
      resumed:
        type:
          - string
          - "null"
      remaining_unresolved:
        type:
          - string
          - "null"
---

# Knowledge Resolution

Record a decision the user has made, in the one place it belongs, so later work uses it. Usually entered from another workflow that found a gap.

## Where a decision belongs

- A decision about one Specification's behavior belongs in that Specification's Decisions section. If that Specification is Approved, don't edit it: report that it needs revising and re-approval.
- A project-wide decision belongs in the knowledge file that already covers the topic, or in `.gnomon/context/PROJECT.md`. Use `.gnomon/decisions/` only if the project already keeps decision records there.
- Record each decision once. Anywhere else that needs it should refer to it, not restate it.

## Steps

1. **Identify** the gap and the decision. If the user hasn't decided yet, ask them in this session; never record your own guess as their decision.
2. **Write** the decision where it belongs.
3. **Check** other knowledge files for statements that now contradict it, and correct them. Report contradictions in Approved Specifications instead of editing them.

## Result

- `outcome` — `RESOLVED` when the decision is recorded; `BLOCKED` when it can't be (no decision yet, or it requires changing an Approved Specification).
- `updated_knowledge` — what changed, and where.
- `adr_created` — path of a decision record you wrote, or `null`.
- `resumed` — what work can now continue.
- `remaining_unresolved` — `BLOCKED` only: what is still needed.
