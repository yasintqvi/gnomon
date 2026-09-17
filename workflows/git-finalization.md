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

# Git Finalization Workflow

## Purpose

Prepare completed, reviewed work for Git handoff while preserving repository integrity and keeping publication separately authorized.

This workflow packages work; it does not implement, review, verify, merge, or broaden the change.

## Inputs and Preconditions

- Completed authorized change and verification status
- Repository state and branch policy
- Requested commit, push, or pull-request scope
- Explicit publication authorization when applicable
- The lifecycle state, per [`SPECIFICATION_LIFECYCLE.md`](../specifications/SPECIFICATION_LIFECYCLE.md), of any Specification governing the disclosed change

Before proceeding, confirm the repository is valid, the intended change is identifiable, unresolved conflicts are absent, and known failures or incomplete work are disclosed. Stop if unrelated changes cannot be safely separated. If any part of the disclosed change is governed by a Specification that is `Draft`, that part is not eligible for finalization — see Failure Handling. There is no authorization that permits finalizing it anyway.

## Execution

### 1. Inspect

Inspect status, diff, current branch, upstream and remote configuration, and recent history. Identify staged, unstaged, untracked, generated, sensitive, and unrelated artifacts.

### 2. Define Change Scope

Classify changed artifacts as in-scope, supporting, unrelated, or uncertain. Stage only the first two categories; exclude sensitive files and preserve user-owned work. Never assume the whole working tree belongs to the task.

Exclude any in-scope or supporting artifact whose governing Specification is `Draft` — treat it as ineligible for this run regardless of how the rest of the change is classified.

### 3. Prepare Branch

Use the current branch when appropriate or create an authorized, policy-compliant branch. If currently on a default or protected branch, create a dedicated branch. Do not rewrite published history, switch away from unprotected work, or alter unrelated branch state without authorization.

### 4. Plan Commits

Choose the smallest coherent commit set. Keep inseparable code, tests, configuration, migrations, and documentation together; split independent concerns only when that improves traceability without breaking intermediate validity. Derive clear messages from actual changes and repository conventions.

### 5. Stage and Validate

Stage explicit intended paths, inspect the staged diff, and confirm it contains no secrets, temporary artifacts, accidental generated output, unrelated changes, or unresolved conflicts. Re-run relevant checks if staging or finalization changed artifacts.

### 6. Commit

Commit only the validated staged scope. Report any local hooks or checks that fail; do not bypass them without explicit authorization.

### 7. Publication Gate

Treat push and pull-request creation as external publication. Proceed only when explicitly requested or already authorized; otherwise stop after local commit preparation and report the next command or action.

### 8. Push and Pull Request

When authorized, push only the intended feature branch—never directly to a protected default branch—and do not force-push unless explicitly authorized and safe. Create or update one pull request for the coherent work unit, covering its purpose, governing specification or decision, important changes, verification evidence, resolved findings, and known limitations. Do not merge it.

### 9. Report

Report the branch, commit hashes and messages, included scope, files intentionally left untouched, verification status, publication state, pull-request reference when created, and remaining user actions or blockers.

## Outputs

```
Git Finalization Result

Outcome: COMMIT_PREPARED | PUBLISHED | BLOCKED
Included:
Excluded:
Remaining unresolved (BLOCKED only):
```

- **Outcome** — `COMMIT_PREPARED` when a validated local commit is prepared but publication was not authorized or requested; `PUBLISHED` when the authorized publication and pull-request steps are also complete; `BLOCKED` when the intended change cannot be identified, unresolved conflicts remain, or the entire requested scope is Draft-governed and therefore excluded.
- **Included** — informational: branch, commit hashes and messages, and publication state when applicable.
- **Excluded** — informational: files intentionally left untouched, including any excluded because their governing Specification is `Draft`.
- **Remaining unresolved** — populated only when `BLOCKED`: the unidentifiable change, the unresolved conflict, or the Draft-governed exclusion that left nothing eligible to finalize.

## Rules

- Do not include unrelated, sensitive, temporary, or environment-specific files.
- Do not use broad staging when unrelated or uncertain changes exist.
- Do not discard or overwrite user-owned work.
- Do not commit unresolved conflicts or present known failing work as complete.
- Do not weaken or bypass hooks and checks without explicit authorization.
- Do not push, force-push, open or update a pull request, or otherwise publish without authorization.
- Do not rewrite published history or merge a pull request unless explicitly authorized by a separate scope.
- Never use finalization to conceal incomplete implementation, verification gaps, or known failures.
- Never finalize work governed by a `Draft` Specification; no authorization or urgency permits this exception.

## Failure Handling

### Draft-Governed Work

Part or all of the disclosed change is governed by a Specification that is `Draft`. Exclude that part entirely — do not stage, commit, or publish it under any authorization. If nothing eligible remains, report `BLOCKED` and direct the governing Specification to [`workflows/specification-definition.md`](specification-definition.md) or to Human Approval as applicable. If an eligible, unrelated portion remains, finalize only that portion.

## Completion Criteria

Finalization is complete at the authorized boundary: either a validated local commit is prepared, or the authorized publication and pull-request steps are complete — and no part of what was finalized is governed by a `Draft` Specification. The report must make that boundary, every exclusion, and every remaining action explicit, with the reported Outcome accurately reflecting which boundary was actually reached.
