# Git Finalization Workflow

## Purpose

Prepare completed, reviewed work for Git handoff while preserving repository integrity and keeping publication separately authorized.

This workflow packages work; it does not implement, review, verify, merge, or broaden the change.

## Inputs and Preconditions

- Completed authorized change and verification status
- Repository state and branch policy
- Requested commit, push, or pull-request scope
- Explicit publication authorization when applicable

Before proceeding, confirm the repository is valid, the intended change is identifiable, unresolved conflicts are absent, and known failures or incomplete work are disclosed. Stop if unrelated changes cannot be safely separated.

## Execution

### 1. Inspect

Inspect status, diff, current branch, upstream and remote configuration, and recent history. Identify staged, unstaged, untracked, generated, sensitive, and unrelated artifacts.

### 2. Define Change Scope

Classify changed artifacts as in-scope, supporting, unrelated, or uncertain. Stage only the first two categories; exclude sensitive files and preserve user-owned work. Never assume the whole working tree belongs to the task.

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

## Rules

- Do not include unrelated, sensitive, temporary, or environment-specific files.
- Do not use broad staging when unrelated or uncertain changes exist.
- Do not discard or overwrite user-owned work.
- Do not commit unresolved conflicts or present known failing work as complete.
- Do not weaken or bypass hooks and checks without explicit authorization.
- Do not push, force-push, open or update a pull request, or otherwise publish without authorization.
- Do not rewrite published history or merge a pull request unless explicitly authorized by a separate scope.
- Never use finalization to conceal incomplete implementation, verification gaps, or known failures.

## Completion Criteria

Finalization is complete at the authorized boundary: either a validated local commit is prepared, or the authorized publication and pull-request steps are complete. The report must make that boundary and every remaining action explicit.
