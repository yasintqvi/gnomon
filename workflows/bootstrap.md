# Bootstrap Workflow

## Purpose

Establish the minimum verified project baseline required for approved development without implementing feature behavior or inventing project decisions.

## When to Use

Use Bootstrap when a repository's code, configuration, or foundational structure is new, incomplete, inconsistent with its approved baseline, or missing required runtime, build, test, or analysis setup. Do not use it for feature implementation, speculative infrastructure, or changes already supported by a valid baseline.

A new codebase is not by itself sufficient to start: Bootstrap assumes `PROJECT.md`, `ARCHITECTURE.md`, `STACK.md`, and `CONVENTIONS.md` already hold approved initial project knowledge, not unfilled templates. When that knowledge does not yet exist, run [`workflows/initial-knowledge-establishment.md`](initial-knowledge-establishment.md) first rather than starting Bootstrap.

## Inputs

- Bootstrap objective, authorized scope, and explicit constraints
- Existing repository, configuration, environment, and user-owned changes
- `PROJECT.md`, `ARCHITECTURE.md`, `STACK.md`, and `CONVENTIONS.md`, approved and sufficient for a baseline, as applicable
- Relevant ADRs
- Applicable Design Knowledge when bootstrapping user-interface foundations:
  - `PRODUCT_EXPERIENCE.md` for required product shell or structural experience context
  - `UI_FOUNDATION.md` for shared presentation foundations
  - `INTERACTION_PATTERNS.md` for shared interaction infrastructure
- Required verification mechanisms and environment constraints

Design Knowledge governs only applicable UI foundations; Bootstrap must not infer feature behavior or create missing design decisions.

## Execution

### 1. Inspect

Inventory the repository before changing it: files, configuration, runtime and dependency state, build and test tooling, existing conventions, partial setup, conflicts, and unrelated work to preserve.

### 2. Verify the Baseline

Compare the observed state with authoritative knowledge. Identify what is valid, missing, inconsistent, obsolete, or blocked, and define the smallest required bootstrap scope. Do not treat assumptions as requirements.

If `context/` holds no approved project-level knowledge beyond unfilled templates, stop before scaffolding and hand off to [`workflows/initial-knowledge-establishment.md`](initial-knowledge-establishment.md) rather than defining a bootstrap scope from invented assumptions.

### 3. Scaffold

Create only the directories and foundational artifacts required by the approved Architecture, Stack, Conventions, and—when applicable—Design Knowledge. Reuse valid existing structure and avoid feature-specific code or speculative modules.

### 4. Configure

Configure only approved runtime, dependency, build, test, static-analysis, formatting, localization, asset, and UI-foundation mechanisms required for the baseline. Use versions and tools defined by Stack or other owning knowledge; do not silently select missing project-wide choices.

### 5. Lock Dependencies

Generate or update dependency manifests and lock state through the approved package manager when dependencies are in scope. Preserve reproducibility, avoid unrelated upgrades, and do not hand-edit generated lock data unless the tool and project explicitly support it.

### 6. Validate

Run available installation, build, test, analysis, formatting, configuration, and minimal startup checks relevant to the baseline. Record commands and results, distinguish environment limitations from project failures, and report checks not run.

## Outputs

- Minimal verified project baseline
- Required structure, configuration, manifests, and lock state
- Foundational test or analysis setup where approved
- Validation evidence
- Explicit blockers, assumptions, and remaining setup work

## Rules

- Remain within authorized bootstrap scope and preserve user-owned changes.
- Do not treat optional practices or commonly paired tools as requirements unless approved knowledge makes them necessary.
- Do not let decisions needed only by future or out-of-scope subsystems block the authorized baseline.
- Prefer existing valid configuration over replacement.
- Do not implement product features, business rules, speculative infrastructure, or future abstractions.
- Do not invent architecture, design, conventions, technology, version, or project-policy decisions.
- Do not scaffold, configure, or lock dependencies from assumed or invented project-level knowledge; confirm approved initial project knowledge exists first.
- Keep generated artifacts consistent and reproducible.
- Require explicit authorization for destructive replacement or removal.
- Do not bypass or weaken validation to report success.

## Failure Handling

### Missing or Conflicting Knowledge

Identify the missing or conflicting decision and its owner, apply established ownership or precedence rules, continue unaffected setup, and stop only the affected work when the issue remains material. Do not select a project-wide default silently. Once a human decision resolves the gap, run [`workflows/knowledge-resolution.md`](knowledge-resolution.md) to update the affected authoritative knowledge before resuming the affected setup.

### Environment Blocker

Record the command and failure, distinguish external restrictions from repository defects, try safe in-scope alternatives, and report the exact remaining prerequisite.

### Repository Conflict

Preserve existing and user-owned work, prefer additive or compatible changes, and obtain authorization before destructive replacement.

### Validation Failure

Diagnose and correct in-scope setup defects, rerun the narrowest failed check, and report unresolved failures without weakening the check.

## Completion Criteria

Bootstrap is complete when the approved baseline is runnable, required services, tools, and foundational structures are configured consistently with applicable project and Design Knowledge, dependency state is reproducible where relevant, validation evidence is recorded, unrelated work is preserved, and blocking or deferred limitations are explicit.
