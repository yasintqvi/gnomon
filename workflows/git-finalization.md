# Git Finalization Workflow

## Purpose

Finalize a completed engineering work unit into a safe and traceable Git history and, when permitted, publish it through a pull request.

## Input

- Active work target
- Current repository state
- Current verification/review outcome
- Git policy

## Preconditions

The workflow must confirm:

- the repository is a Git repository;
- there are no unresolved merge conflicts;
- the active work scope is identifiable;
- required verification has passed or accepted limitations are recorded;
- the current repository state does not contain ambiguous changes that cannot be safely attributed to the active work.

## Workflow

### 1. Inspect Repository

Inspect:

- current branch;
- default branch;
- working tree;
- staged changes;
- untracked files;
- remotes;
- commits relevant to the active work.

### 2. Determine Change Scope

Classify each changed file as:

- In Scope
- Supporting
- Unrelated
- Uncertain

Do not automatically stage Unrelated or Uncertain changes.

### 3. Prepare Branch

If currently on the default or protected branch, create a dedicated branch derived from the active engineering intent.

### 4. Prepare Commit Plan

Group changes by engineering intent.

Prefer multiple coherent commits over one oversized commit when the work contains independently meaningful changes.

Commit messages must describe intent rather than file operations.

### 5. Stage and Validate

Stage only the files belonging to the current commit.

Inspect the staged diff before committing.

Never use broad staging when unrelated changes are present.

### 6. Commit

Create local commits only when:

- staged scope is coherent;
- no unrelated changes are included;
- known required checks are satisfied.

### 7. Publication Gate

Apply the configured Git mode.

For managed-with-approval mode:

- local commits may be created automatically;
- stop before the first push;
- present branch, commits, and PR plan;
- require explicit user approval.

### 8. Push

Push only the prepared feature branch.

Never push directly to a protected default branch.

Never force-push unless explicitly authorized.

### 9. Pull Request

Create or update one PR for the coherent work unit.

The PR must include:

- purpose;
- implemented specification or decision;
- important changes;
- verification evidence;
- resolved findings;
- known limitations or deferred work.

### 10. Report

Report:

- branch name;
- commits created;
- files intentionally left untouched;
- push status;
- PR status;
- remaining Git concerns.

## Prohibited Behavior

The workflow must not:

- blindly run `git add .`;
- include unrelated changes;
- commit unresolved merge conflicts;
- publish known failing work as complete;
- rewrite published history without explicit authorization;
- merge its own PR.