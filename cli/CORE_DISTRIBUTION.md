# Core Distribution / Consumption — CLI Step 2

## Purpose

Define where Gnomon Core lives relative to the CLI and a user's project, how the CLI and the Agent come to consume the exact same authoritative Core during a workflow run, and how that arrangement supports customization, portability, and future upgrades without a runtime merge engine.

This is CLI-level design, not Core semantics. It defines delivery and authority, not what any workflow means.

---

## Scope

This document covers:

- The Bundled Default Core and its role.
- The Project-local Core and its status as sole runtime authority.
- The categories of artifacts involved, and which are Gnomon-owned versus project-owned.
- The reasons this model was chosen over the alternatives considered.
- How version compatibility is determined, and against which copy.
- What remains inside the CLI as mechanism, never as project Markdown.

This document does not cover:

- Workflow Contract v1's field semantics — defined by CLI Step 1 (`WORKFLOW_CONTRACT.md`).
- The exact materialized layout, `gnomon init`'s behavior, or the initialized-project definition — defined by CLI Step 3 (`PROJECT_INITIALIZATION.md`), referenced here rather than duplicated.
- `gnomon upgrade`, migration/merge mechanics, Agent adapters, Agent provider configuration, or remote Core fetching — later CLI Steps.
- Approval-evidence persistence — defined by CLI Step 6 (`APPROVAL_RUNTIME.md`).

---

## The Model

```text
Gnomon CLI installation
├── CLI mechanisms
└── Bundled Default Core
          │
          │ gnomon init
          ▼
User Project
└── .gnomon/
      └── Project-local Core
             │
             └── sole runtime authority
```

---

## Bundled Default Core

The CLI ships with a complete **Bundled Default Core** — every workflow, semantic-definition document, evaluation criterion, template, and default Artifact Contract Gnomon defines, suitable for materialization into a project.

Its role is limited to: providing the default assets `gnomon init` materializes, and providing the reference used by future compatibility and upgrade mechanisms (not designed here) to detect drift and source updated content.

**It is not a second runtime Core.** Once a project has been initialized, the Bundled Default Core is never consulted during a workflow run — only during init, and later during upgrade/validation operations, which are categorically separate from executing a workflow.

---

## Project-local Core

After initialization, the materialized Core under the project's `.gnomon/` directory is the **sole runtime authority**.

During workflow execution:

- The CLI reads the project-local workflow metadata (frontmatter, per `WORKFLOW_CONTRACT.md`).
- The Agent reads the project-local workflow Markdown — the same file, frontmatter and prose together — as its instructions, along with whatever other project-local Core or project-owned documents it needs to inspect.

Both therefore consume the same authoritative artifact, because there is exactly one live copy of it during any given run. **The CLI must never validate one Core while the Agent executes another** — this is not achieved by an added detection mechanism; it follows directly from there being only one file to open in the first place. The Bundled Default Core is dormant during execution and cannot become a second source of truth by construction, not merely by convention.

---

## Project Ownership and Customization

- **Gnomon semantic/kernel material** — workflow frontmatter, the lifecycle and dependency semantic documents, evaluation/readiness criteria. Materialized into the project, but protected: a Human may still edit it, at the cost of that workflow being reclassified away from "recognized built-in" (per `WORKFLOW_CONTRACT.md`'s validation model).
- **Agent-facing workflow guidance** — the prose half of the same workflow files. Materialized into the project, freely customizable in place, with no separate override file or merge step: editing the prose directly is the entire customization mechanism.
- **Project-owned knowledge** — `PROJECT.md`, `DOMAIN.md`, `ARCHITECTURE.md`, `STACK.md`, `CONVENTIONS.md`, Design Knowledge, ADRs. Always project-local; never shipped as project-specific content, only as unfilled placeholders or reusable defaults (see `PROJECT_INITIALIZATION.md`).
- **Specifications, decisions, project artifacts** — always project-local, created only by their owning workflows, never part of the Bundled Default Core's own content beyond their templates.
- **Templates/default assets** — the SPEC template, the ADR template, and the default Artifact Contracts. Gnomon-owned defaults, materialized into the project, and — like workflow prose — fully open to project customization, with no protected/frontmatter concept applying to them at all.
- **CLI runtime mechanisms** — parsing, lifecycle derivation, dependency calculation, and all other code. Never materialized; see What Remains Inside the CLI.

Project-local content is Git-visible and portable with the repository — cloning the project recovers its entire Gnomon constitution, no external fetch required. Legitimate project customization requires no runtime merge or override engine in v1: because frontmatter and prose occupy clearly separated regions of the same materialized file, "customize" simply means "edit the file," and validation (not merging) is what keeps that safe. Upgrade merging itself is not designed in this document.

---

## Distribution Principles

This model was chosen, over bundling Core inside the CLI with no local copy, for the following reasons, each tested against a real requirement:

- The Agent can naturally inspect project-local files — including following the relative links Core documents already use between each other — using its normal filesystem access, with no CLI-assembled context injection required.
- CLI and Agent share exactly one runtime source of truth, by construction, not by an added synchronization mechanism.
- Projects remain reproducible and portable: cloning the repository alone recovers the full Gnomon constitution.
- Initialization works offline, because the Bundled Default Core ships with the CLI and is never fetched over a network.
- Project-specific guidance lives with the project, versioned and reviewable through the project's own ordinary Git history.
- Runtime execution never depends on fetching a remote Core.
- The upstream Gnomon source-repository layout (Core content at its own root, with no self-referential marker) is distinct from, and must not be confused with, a consuming project's materialized layout under `.gnomon/`.

A bundled-only model with no local materialization was considered and rejected specifically because it fails the first principle above for any filesystem-scoped Agent — the dominant integration shape — and because it eliminates project customization as a byproduct of eliminating any local copy to customize.

---

## Version Compatibility

Three distinct version facts participate in compatibility:

- **CLI-supported contract version** — the range of Workflow Contract versions the running CLI understands; compiled into the CLI itself.
- **Bundled Default Core contract version** — the version of the reference Core the CLI would install or upgrade to; consulted only during init/upgrade.
- **Project-local `CONTRACT_VERSION`** — the version actually in force for this project, read from `.gnomon/CONTRACT_VERSION`.

**At runtime, compatibility is always determined against the project-local `CONTRACT_VERSION`, because the project-local Core is authoritative.** The Bundled Default Core's own version matters only when installing or upgrading, never as an alternate compatibility source during a workflow run. Migration behavior for a version mismatch is explicitly not designed here.

---

## What Remains Inside the CLI

The following stay CLI-internal mechanism, now and for any future Step, and are never moved into project Markdown merely for the sake of portability:

- Frontmatter parsing and structural validation.
- Lifecycle derivation (comparing current content against approval evidence).
- Dependency-fact calculation (Target Presence, Target Lifecycle State, cycle detection).
- Agent adapters and process management.
- Git mechanisms.

These are mechanisms that *implement* what Core's semantic documents define; they are not themselves semantic content, and moving them into a Markdown file would not make the project more portable — it would only make the mechanism harder to trust and harder to change safely.

---

## What Is Materialized

The categories of Core and project-facing artifacts materialized into `.gnomon/` — the exact layout, what is copied verbatim, what is copied as an unfilled placeholder, what is created empty, and `gnomon init`'s full behavior — are defined in [`PROJECT_INITIALIZATION.md`](PROJECT_INITIALIZATION.md) and are not duplicated here.

---

## Deferred to Future Work

The following are intentionally not defined here, and nothing in this document depends on them:

- `gnomon upgrade` and any migration or merge mechanism.
- Agent adapter architecture and Agent provider configuration.
- Approval-evidence persistence — defined by CLI Step 6 (`APPROVAL_RUNTIME.md`).
- Remote Core fetching.
- Monorepo orchestration.

---

## Notes

This document persists the Step 2 decisions already finalized in prior design analysis. It does not reopen or redesign that analysis; it is the standalone durable record of it, checked against `PROJECT_INITIALIZATION.md` and the current Core documents for terminology consistency only.
