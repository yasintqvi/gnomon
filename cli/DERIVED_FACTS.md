# Derived Facts — CLI Step 7

## Purpose

Identify the deterministic facts Gnomon CLI must derive from current repository and `.gnomon/` state, without storing duplicate or cached state.

This is CLI-level design, not Core semantics, and largely not new mechanism either: almost every fact documented here reuses a derivation already fully specified by an earlier Step. This document's job is to inventory the closed set of facts CLI v1 actually needs, name their sources and consumers, and fill the one placement Core itself explicitly deferred — where dependency-cycle detection runs.

---

## Scope

This document covers:

- The closed inventory of deterministic facts CLI v1 derives from repository truth.
- Each fact's source of truth, derivation, and consumer.
- The composite pre-start eligibility computation these facts combine into.
- The precise CLI-vs-Agent reasoning boundary.
- The one known v1 limitation in dependency-list extraction.

This document does not cover:

- How Draft/Approved is derived — CLI Step 6 (`APPROVAL_RUNTIME.md`) remains the sole authority; this document only consumes that result.
- Workflow Contract field semantics — CLI Step 1 (`WORKFLOW_CONTRACT.md`) remains the sole authority; this document only consumes those fields.
- `.gnomon/` materialization and the initialized-project definition — CLI Step 3 (`PROJECT_INITIALIZATION.md`) remains the sole authority; this document only consumes that check.
- Agent startup, Terminal Handoff, or the workflow result envelope — CLI Steps 4–5.
- Any dependency threshold, materiality judgment, readiness assessment, or other semantic conclusion — these remain entirely Agent/Core reasoning and are never computed here.

---

## Preconditions Carried Forward (Steps 1–6, unchanged)

- **Workflow Contract v1** (Step 1): `identity`, `specification_reference`, `requires_approved_specification`, and the Result Contract live in each workflow's own frontmatter.
- **Core Distribution / Project Initialization** (Steps 2–3): `.gnomon/` is the sole runtime authority; `CONTRACT_VERSION` records the global contract version.
- **Result Protocol** (Step 5): unaffected by this document — pre-start facts and post-run result validation remain separate concerns.
- **Approval Runtime** (Step 6): Draft/Approved is derived from current Specification content plus `.gnomon/approvals/` evidence, never stored.

Nothing in this document changes any of these. This document introduces no new persisted file, directory, cache, or state — every fact below is recomputed from data these earlier Steps already made durable.

---

## Fact Inventory

| Fact | Source of Truth | Deterministic Derivation | Consumer | Persisted? |
|---|---|---|---|---|
| **Specification existence** | `specifications/SPEC-NNN-*.md` | Filesystem check against Discovery's identity rule | Step 1 gates (Definition, Implementation, Testing); dependency facts below | No — recomputed on every query |
| **Specification Draft/Approved lifecycle** | Current Specification content + `.gnomon/approvals/SPEC-NNN/*` | Exactly Step 6's rule; not re-derived here | Step 1's `requires_approved_specification` gate; per-dependency lifecycle fact | No — Step 6's own guarantee |
| **Declared dependencies (raw list)** | The Specification's own `## Dependencies` section | Extraction from the existing narrow list convention (see Known Limitation) | Definition, Implementation, Testing, Verification, Review | No |
| **Per-dependency target presence** | Specification-existence fact, applied to each declared identity | Same mechanism as Specification existence | Same consumers as declared dependencies | No |
| **Per-dependency target lifecycle** | Draft/Approved fact, applied to each declared identity | Same mechanism as Specification lifecycle | Same consumers as declared dependencies | No |
| **Dependency-cycle existence** | Every Specification's own declared list, project-wide | Standard directed-graph cycle check over direct declared edges only | Specification Definition, before recording a new or changed dependency | No |
| **Next Specification identity** | `specifications/SPEC-NNN-*.md` directory listing | Discovery's already-defined highest-plus-one algorithm | Draft Creation (a deterministic CLI operation, not an Agent workflow) | No |
| **Workflow Contract metadata** | The selected workflow's own frontmatter | Direct read — loaded, not computed | CLI pre-start gating; Step 5 result validation | No |
| **Global contract-version compatibility** | `.gnomon/CONTRACT_VERSION` | Direct comparison against the CLI's own compiled-in supported range | Gate on trusting Workflow Contract metadata at all; every invocation | No |
| **Project-initialized structural state** | `.gnomon/` tree + `CONTRACT_VERSION` | Exactly Step 3's definition; not re-derived here | CLI startup, before any workflow invocation | No |
| **Pre-start eligibility (composite)** | Specification existence + lifecycle + Workflow Contract metadata + contract-version compatibility | Step 1's already-specified gate logic, applied to these facts | CLI orchestration, immediately before `run()` | No |
| **Working tree has uncommitted changes** | Git working-tree state | Any staged, unstaged, or untracked difference from `HEAD` — `true`/`false`, no further breakdown | CLI Step 8 next-action guidance, as a coarse signal that Git Finalization may be worth considering | No |

Every fact is **recomputed fresh from current repository and `.gnomon/` truth on each query** — none is cached, stored, or carried forward between invocations. No new directory, file, database, or state mechanism is introduced by this document.

The working-tree fact is deliberately a plain boolean, not a staged/unstaged/untracked breakdown: both of its consumers need presence, not composition — `gnomon status` reports it as part of its project-state snapshot (`COMMAND_SURFACE.md`), and CLI Step 8's `next` uses it for a coarse "Git Finalization may be worth considering" suggestion. Git Finalization's own Step 1 (Inspect) independently re-derives the full breakdown itself when actually invoked; CLI must never infer semantic readiness to finalize from this fact — that judgment remains entirely the Agent's.

---

## Dependency Facts, Precisely Bounded

Core mechanically defines exactly two derivable facts per declared dependency edge — target presence and target lifecycle state (`SPECIFICATION_DEPENDENCIES.md`'s own "Derived Dependency Facts") — and explicitly refuses to define a third, universal verdict: *"No universal `SATISFIED` / `UNSATISFIED` verdict is defined here... Whether a given Target Presence and Target Lifecycle State combination is acceptable is workflow-relative."*

**CLI computes only the two raw facts and never a threshold.** Whether a `MISSING` or `DRAFT` dependency actually blocks the work at hand is already fully resolved, per workflow, inside each of the ten workflow files' own Rules and Failure Handling (Implementation's "stop only the portion... that materially depends on it," Testing's per-test exclusion, Verification's `UNVERIFIABLE` marking, Review's `RISK`/`KNOWLEDGE GAP` classification) — this remains entirely Agent reasoning, applying that specific workflow's own Core-stated rule to the raw facts CLI supplies. CLI must never invent, compute, or store a universal `SATISFIED`/`UNSATISFIED` state.

**Dependency-cycle existence** is the one dependency question Core states as a hard, non-workflow-relative rule, and is therefore appropriately computed directly by CLI: a standard cycle check over a directed graph built from every Specification's own declared, **direct-only** dependencies — matching `SPECIFICATION_DEPENDENCIES.md`'s own rule that transitive relationships are always derived by following direct declarations, never separately authored. Core states the rule but explicitly defers "where such a check runs at runtime" as later implementation work; this document is where that deferred placement lands, filling a gap Core itself named rather than introducing new semantics.

---

## Conceptual Relationship

```text
Repository truth (.gnomon/ + specifications/)
        ↓
Global contract-version compatibility
        ↓
Workflow Contract metadata (per selected workflow)
        +
Specification existence / lifecycle (for any supplied Specification)
        ↓
Pre-start eligibility
```

Contract-version compatibility sits ahead of Workflow Contract metadata, because it is the precondition for trusting that the frontmatter parse means what the CLI expects — an unsupported version stops the chain before that metadata is interpreted at all, per Step 1's own "refuse rather than guess" principle.

---

## CLI-vs-Agent Reasoning Boundary

**Deterministic, CLI-owned:** every fact in the inventory above — existence, lifecycle, declared-dependency list, per-dependency presence and lifecycle, cycle existence, next identity, Contract metadata, contract-version compatibility, initialized-project state, and pre-start eligibility.

**Agent-owned, never CLI:** whether a `MISSING`/`DRAFT` dependency should actually block a specific portion of current work; whether a knowledge gap is material; whether Specification Definition's readiness criteria are satisfied; whether an implementation correctly satisfies its acceptance criteria; whether Verification evidence sufficiently proves an obligation; whether a Review finding is a `DEFECT` versus a `RISK`; and, generally, the entire semantic content of every workflow's `Outcome`, `Aggregate`, or findings. CLI computes and gates on facts; it never renders a judgment.

---

## Known v1 Limitation

A Specification's `## Dependencies` section is currently extracted from the existing narrow Markdown list convention (a bulleted `* SPEC-NNN` list, per the established template) rather than from a dedicated machine-readable contract. This is narrow and regular enough to extract reliably via simple, single-line pattern matching — closer to a safe, closed-vocabulary declaration than to general free-form prose parsing — but it is not as formally guaranteed as a schema-validated structure, since nothing today mechanically enforces that exact shape. This is recorded as a known, currently-acceptable limitation, not evidenced as causing actual problems. **Core is not redesigned and no new dependency schema is introduced to address it in this task.**

---

## Deferred to Future Work

- A machine-readable Dependencies contract, if the known limitation above is ever shown to cause real unreliability.
- Any general-purpose eligibility or rules language — no case examined here or in any prior Step required one.
- Everything already deferred by Steps 1–6, unaffected by this document.

---

## Notes

This document persists the Step 7 architecture finalized through prior design analysis. It introduces no new mechanism beyond naming dependency-cycle detection's deterministic placement — a gap `SPECIFICATION_DEPENDENCIES.md` itself explicitly left open — and consolidating facts already fully specified by Steps 1, 3, 5, and 6 into one inventory. It does not reopen or redesign any earlier Step, and it does not modify Core.
