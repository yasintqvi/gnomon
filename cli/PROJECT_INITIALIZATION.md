# Project Initialization Model — CLI Step 3

## Purpose

Define exactly what `gnomon init` does, creates, validates, and leaves behind, and the precise, mechanically checkable definition of an "initialized" Gnomon project.

This is CLI-level design, not Core semantics. It does not alter what any workflow means or requires; it defines how the Bundled Default Core is delivered into a project and how the CLI recognizes that the delivery succeeded — a deterministic filesystem operation, nothing more.

---

## Scope

This document covers:

- The dedicated Gnomon root directory and the canonical initialized project layout.
- `gnomon init`'s exact responsibility boundary.
- The mechanically checkable definition of an initialized project.
- Materialization behavior for each Core artifact category.
- Re-initialization, root discovery, and validation behavior.
- The boundary between `gnomon init` and Initial Knowledge Establishment.

This document does not cover:

- Workflow Contract v1's field semantics — defined by CLI Step 1, restated below only as context, not redefined here.
- The Bundled Default Core / Project-local Core distribution model — defined by CLI Step 2, restated below only as context, not redefined here.
- `gnomon upgrade`, migration behavior, force/reset semantics, monorepo orchestration, Agent adapters, Agent provider selection, the workflow result protocol, CLI command syntax, or remote Core fetching — all later CLI Steps.
- The Approval Runtime itself — CLI Step 6 (`APPROVAL_RUNTIME.md`); this document only reserves the empty `approvals/` area init creates.

CLI Steps 1 and 2 currently exist only as prior design record; this document restates only what Step 3 depends on from them, and does not itself constitute their durable record.

---

## Preconditions Carried Forward (Steps 1–2, unchanged)

- **Workflow Contract v1** (Step 1): every Agent workflow declares `identity`, `specification_reference` (`none | optional | required`), and `requires_approved_specification` (boolean, meaningful only when the reference isn't `none`) in frontmatter. A single global `CONTRACT_VERSION` value identifies the Core/contract compatibility version.
- **Core Distribution** (Step 2): the CLI ships a Bundled Default Core. `gnomon init` materializes it into the project. The resulting **project-local Core is the sole runtime authority**. The Bundled Default Core is never a competing runtime authority during a workflow run — it is consulted only during init and future upgrade/validation.

Nothing in this document changes either conclusion.

---

## The Gnomon Root

`gnomon init` materializes Core into a dedicated directory, **`.gnomon/`**, at the project's root.

Why this choice, evaluated against the required properties:

- **Clearly Gnomon-owned, minimal collision risk** — a leading dot follows the established convention for tool-recognized project roots (`.git/`, `.github/`) that few projects will coincidentally already have, unlike a bare `gnomon/` name.
- **Suitable for committing to Git, and for humans** — a dot-prefix does not mean hidden or unimportant: `.github/` is the direct precedent for a dot-directory holding actively human-edited, load-bearing content (workflows, templates) that renders normally in a Git host's file browser. Gnomon's project knowledge is exactly this kind of content.
- **Contains both project-local Core and project-owned knowledge/artifacts** — everything Core-derived and everything project-owned lives under this one root; see the canonical layout below.
- **Works with upward search from a nested working directory** — the CLI locates the root by walking upward from the current working directory looking for a directory literally named `.gnomon`, exactly as Git locates `.git`. No new discovery mechanism is introduced.
- **Does not confuse the upstream Gnomon source repository with a consuming project** — Gnomon's own source repository places `workflows/`, `context/`, etc. directly at its own root, with no self-referential root marker, because it does not need one. A consuming project's materialized layout is a related but distinct thing, and must never be assumed identical to the source repository's own layout.
- **No unnecessary nesting** — every category sits directly under `.gnomon/`, at the same depth Gnomon's own source repository already uses for each one.

### Canonical Initialized Project Layout

```
.gnomon/
  CONTRACT_VERSION
  workflows/
    initial-knowledge-establishment.md
    bootstrap.md
    specification-discovery.md
    specification-definition.md
    knowledge-resolution.md
    implementation.md
    testing.md
    verification.md
    review.md
    git-finalization.md
  specifications/
    SPECIFICATION_LIFECYCLE.md
    SPECIFICATION_DEPENDENCIES.md
    SPEC-000-use-case-name.md
  evaluations/
    SPECIFICATION_READINESS_CRITERIA.md
    ENGINEERING_REVIEW_QUESTIONS.md
    VERIFICATION_CRITERIA.md
  contracts/
    behavior-verification.md
    input-qualification.md
    interactive-presentation.md
    persistent-structure-evolution.md
    use-case-execution.md
  decisions/
    ADR-000-decision-title.md
  approvals/
  context/
    PROJECT.md
    DOMAIN.md
    ARCHITECTURE.md
    STACK.md
    CONVENTIONS.md
    PRODUCT_EXPERIENCE.md
    UI_FOUNDATION.md
    INTERACTION_PATTERNS.md
```

This uses only artifact categories that actually exist in Gnomon's current Core — no invented category has been added.

---

## Materialization: What Comes From the Bundled Default Core

- **`workflows/*.md`** — copied verbatim (frontmatter + prose). See [`../workflows/`](../workflows/) for the current set.
- **`specifications/SPECIFICATION_LIFECYCLE.md`, `SPECIFICATION_DEPENDENCIES.md`** — copied verbatim; see [`../specifications/SPECIFICATION_LIFECYCLE.md`](../specifications/SPECIFICATION_LIFECYCLE.md) and [`../specifications/SPECIFICATION_DEPENDENCIES.md`](../specifications/SPECIFICATION_DEPENDENCIES.md).
- **`specifications/SPEC-000-use-case-name.md`** — copied as the template; no actual `SPEC-NNN-*.md` is created by init.
- **`evaluations/*.md`** — copied verbatim (readiness, review-question, and verification criteria).
- **`contracts/*.md`** — copied verbatim, as already-authored, reusable Artifact Contract defaults, on the same footing as Design Knowledge below: usable as-is, project-customizable, no frontmatter/gating concept applies to them.
- **`decisions/ADR-000-decision-title.md`** — copied as the template only; no actual `ADR-NNN-*.md` is created by init.
- **`context/PRODUCT_EXPERIENCE.md`, `UI_FOUNDATION.md`, `INTERACTION_PATTERNS.md`** — copied with real, usable default content, consistent with Initial Knowledge Establishment's own description of these as already-authored, reusable baselines that apply as-is by default.
- **`context/PROJECT.md`, `DOMAIN.md`, `ARCHITECTURE.md`, `STACK.md`, `CONVENTIONS.md`** — copied as **unfilled template placeholders only**. Init never invents project-specific content here; these become project-owned the instant they exist, ready for Initial Knowledge Establishment to fill.
- **`CONTRACT_VERSION`** — written by the CLI from its own bundled version, last (see Initialization-Completion Detection).

## What Is Created as Project-Owned Scaffolding, Not Copied Content

- `specifications/` beyond the template, `decisions/` beyond the template, and `approvals/` — directories exist (the first two holding only their template; `approvals/` created fully empty) but contain no actual instances.
- Actual `SPEC-NNN-*.md` files, actual `ADR-NNN-*.md` files, filled-in `context/PROJECT.md`/`DOMAIN.md`/`ARCHITECTURE.md`/`STACK.md`/`CONVENTIONS.md` content, and approval evidence are all created later, exclusively by their owning workflows or mechanisms (Specification Discovery, Knowledge Resolution, Initial Knowledge Establishment, and the Approval Runtime — CLI Step 6, `APPROVAL_RUNTIME.md` — respectively) — never by init. No Specification is approved at init time; `approvals/` starts, and typically remains for some time, completely empty.

---

## `gnomon init` Responsibility

**Init owns:**

- Locating or establishing the Gnomon root (`.gnomon/`).
- Validating the Bundled Default Core and the destination before writing anything.
- Materializing the artifacts listed above.
- Writing `CONTRACT_VERSION`.

**Init does not own, under any circumstance:**

- Inspecting or interpreting the project's existing source code, configuration, or documentation. Init's behavior is identical for an empty project and an existing codebase — that differentiation belongs entirely to Initial Knowledge Establishment's own Inspect step, never to init.
- Capturing Human project intent, or writing any such intent into `context/` itself. Any intent a Human states in the same CLI session belongs to Initial Knowledge Establishment as its Input, never to init as stored state.
- Populating semantic project knowledge of any kind.
- Invoking an Agent, in any workflow, for any reason.
- Running Initial Knowledge Establishment or Bootstrap.
- Configuring an Agent provider, or storing any Agent-related credential or preference.
- Running `git init`, making a commit, or otherwise managing Git beyond leaving files ready to be committed. Git initialization is a separate, consequential decision that must never be a silent side effect of `gnomon init`.

---

## Definition of an Initialized Project

A project is initialized when, and only when, the following structural facts are all true, checked by direct inspection of the filesystem:

1. A `.gnomon/` root exists at a locatable path.
2. Every path the Bundled Default Core specifies as materialized content exists at its expected location under that root.
3. `.gnomon/CONTRACT_VERSION` exists, is well-formed, and holds a value the running CLI recognizes as supported.

This definition is **independent of whether Initial Knowledge Establishment, Bootstrap, or any Agent workflow has ever run.** It is a fixed structural fact about Gnomon's own scaffolding, not a claim about the project's knowledge maturity — conflating the two would make "initialized" a moving target rather than something the CLI can check in one pass.

---

## Initialization-Completion Detection

`CONTRACT_VERSION`'s only semantic meaning is, and remains, the Core/contract compatibility version — it is never repurposed to also mean "initialization is complete." No separate marker file, initialization database, state machine, history log, or timestamp is introduced.

"Initialized," "partially/interrupted," and "structurally damaged" are all **derived**, computed fresh each time by the same structural inspection used in the Definition above:

- **Fully initialized** — every required path is present, and `CONTRACT_VERSION` is present, well-formed, and supported.
- **Partially initialized / interrupted** — the root exists, but one or more required paths (`CONTRACT_VERSION` included) are simply **absent**.
- **Structurally damaged** — a required path exists but its content is invalid (e.g., `CONTRACT_VERSION` exists but does not hold a parseable version; a workflow file exists but its frontmatter is malformed). This reuses Workflow Contract v1's own already-defined validation classifications (Step 1) rather than inventing a new vocabulary for init specifically.

During materialization, the CLI still writes `CONTRACT_VERSION` **last**. This is an operational sequencing detail of the write process, not a second responsibility assigned to the file: an uninterrupted init naturally produces this file last regardless of what it means, so its absence is simply the most diagnostic instance of "a required path is absent" — not a deliberately bolted-on completion flag. No marker beyond this was found to be justified; deriving these three states from plain structural inspection is sufficient on its own.

---

## Re-initialization Behavior

| State found | Behavior |
|---|---|
| Never initialized | Full init. |
| Correctly initialized, matches expected | **No-op** — report already initialized; write nothing. |
| Partially present (some required paths missing) | **Fill only the missing paths** — never touch a path that already exists. |
| Required paths manually deleted | Same as partial — filling an absent path is always safe, since nothing existing is overwritten. |
| `CONTRACT_VERSION` unsupported (too new for this CLI) | **Refuse clearly** — write nothing; report the incompatibility. |
| Custom workflow files present | Untouched in every case above — outside the Bundled Default Core's file set entirely, and outside init's authority to alter. |

No broad repair framework is introduced. Every case above reduces to one rule: **write only what is absent; never overwrite what already exists; refuse on version incompatibility.**

---

## Existing Codebases

Init's behavior does not branch on codebase content in any way — the same checks, the same writes, the same omissions occur whether the target is an empty project or an established codebase. Understanding what already exists in the project is Initial Knowledge Establishment's and Bootstrap's responsibility, exercised only after init has finished, never init's own.

---

## Git

Init does not require the project to already be a Git repository, does not run `git init`, does not create a commit, and does not modify `.gitignore`. Every materialized path is left ready for the Human to review and commit through their own ordinary Git workflow. This is unrelated to, and must not be confused with, Git Finalization's own semantics.

---

## Agent Configuration

None. Init requires no Agent provider to be selected or configured to succeed, and stores no Agent-related preference or credential. A project must remain structurally valid and clonable on a machine where no Agent has ever been configured.

---

## Init → Initial Knowledge Establishment Boundary

Init establishes structure. Initial Knowledge Establishment establishes project knowledge. These are strictly sequential and categorically separate: init's success and completion, as defined above, never depend on Initial Knowledge Establishment running at all. The CLI may recommend Initial Knowledge Establishment as the natural next action once init succeeds — a product convenience — but this is a UX recommendation layered on top of, not a part of, init itself.

---

## Root Discovery

The CLI locates the Gnomon root by walking upward from the current working directory, at each level checking for a directory literally named `.gnomon`, stopping at the first match — the same algorithm Git uses to locate `.git`. No new discovery mechanism, registry, or configuration is introduced. Multiple independent Gnomon projects within a single monorepo are not addressed by this mechanism and are explicitly out of scope for this Step (see Deferred).

---

## Validation

- **Bundled Default Core** — validated at CLI build/release time (frontmatter well-formed, workflow identities unique, all required assets present) as a packaging correctness check, and re-validated cheaply at every init as defense against a corrupted install. A malformed bundle causes init to refuse and report an internal error rather than partially materialize a known-broken Core.
- **Materialized Project Core** — checked against the Definition above (path presence, `CONTRACT_VERSION` presence and validity).
- **Workflow Contract v1** — reuses Step 1's own validation classifications (missing identity, invalid `specification_reference` value, malformed frontmatter, unknown fields safely ignored, and so on) without redefining them here.
- **Global contract compatibility** — a direct comparison of the project's `CONTRACT_VERSION` against the CLI's own supported range: matches or is older-but-supported → proceed; newer than supported → refuse.

Init performs none of the semantic workflow reasoning any Core document already owns (materiality, readiness, dependency evaluation, and so on) — validation here is limited strictly to structure.

---

## Deferred to Future Work

The following are intentionally not defined here, and nothing in this document depends on them:

- `gnomon upgrade` and any migration mechanism.
- Force/reset semantics for a deliberate hard-reset to the bundled default.
- Monorepo orchestration (multiple independent Gnomon projects in one repository).
- Agent adapter architecture and Agent provider selection (CLI Step 4 and beyond).
- The workflow result protocol.
- CLI command hierarchy and interactive terminal UX.
- Remote Core fetching.
- Runtime workflow execution mechanics.

---

## Notes

CLI Steps 1 and 2 are restated above only as far as Step 3 depends on them, and remain otherwise recorded only in prior design discussion pending their own dedicated documentation.
