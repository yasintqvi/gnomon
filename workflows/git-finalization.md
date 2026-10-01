---
identity: git-finalization
specification_reference: none
requires_approved_specification: false
result:
  terminal_path: outcome
  classification:
    COMMIT_PREPARED: success
    PUBLISHED: success
    BLOCKED: blocked
  schema:
    type: object
    required:
      - outcome
    properties:
      outcome:
        type: string
        enum:
          - COMMIT_PREPARED
          - PUBLISHED
          - BLOCKED
      included:
        type:
          - string
          - "null"
      excluded:
        type:
          - string
          - "null"
      remaining_unresolved:
        type:
          - string
          - "null"
---

# Git Finalization

Commit the finished work cleanly. Push or open a pull request only when the user has explicitly asked for it.

## Steps

1. **Inspect** status, diff, branch, and remotes.
2. **Choose what to include**: only files that belong to this work. Exclude secrets, generated or temporary files, and unrelated changes. Exclude changes that implement a Draft Specification, whatever else is authorized.
3. **Branch** if you are on a default or protected branch.
4. **Commit** the staged scope with messages that follow the repository's conventions. Don't bypass hooks.
5. **Publish** (push, pull request) only if explicitly authorized. Never force-push or merge.

## Result

- `outcome` — `COMMIT_PREPARED` (committed locally), `PUBLISHED` (pushed or pull request opened), or `BLOCKED`.
- `included` — branch, commits, and files.
- `excluded` — files deliberately left out, and why.
- `remaining_unresolved` — `BLOCKED` only: what prevents finalizing.
