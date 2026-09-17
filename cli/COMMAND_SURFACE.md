# Command Surface — CLI Step 9

## Purpose

Define the smallest, clearest set of CLI commands through which a Human can perform the User Journey (Step 8) without needing to understand Gnomon's internal architecture.

This is CLI-level design, not Core semantics, and not implementation. It names commands and maps each to the mechanism Steps 1–8 already fully specify; it defines no argument-parsing behavior, no internal package structure, and no exact output formatting.

---

## Scope

This document covers:

- The finalized public command hierarchy.
- Which mechanism each command invokes, and where that mechanism is authoritatively defined.
- The distinct responsibilities of `status`, `validate`, and `next`.
- Naming and discoverability principles.
- Conceptual gating behavior when a command's deterministic pre-start eligibility check fails.

This document does not cover:

- Any CLI framework, argument parser, command handler, or internal package structure — CLI Step 10 and later.
- Provider selection or Agent configuration — CLI Step 4's own deferred scope.
- Exact terminal styling, output formatting, or error-message wording.
- Any change to User Journey (Step 8), Derived Facts (Step 7), Approval Runtime (Step 6), Result Protocol (Step 5), Agent Adapter (Step 4), or Workflow Contract (Step 1) — all referenced here, none redefined.

---

## Preconditions Carried Forward (Steps 1–8, unchanged)

- **Workflow Contract v1** (Step 1): every workflow's `specification_reference`/`requires_approved_specification` define its deterministic pre-start gate; custom workflows already work with this same contract, no CLI code changes required.
- **Approval Runtime** (Step 6): only explicit Human action creates or revokes approval; an Agent result never constitutes approval.
- **Derived Facts** (Step 7): every fact a command consults is recomputed fresh from repository truth, never cached.
- **User Journey** (Step 8): no canonical mandatory pipeline; valid actions are mechanical, recommended actions are non-authoritative and split into same-session (high-confidence) and cross-session (coarse) modes.

Nothing in this document changes any of these. No command introduces new persisted state.

---

## Final Public Command Hierarchy

```text
gnomon agent
gnomon agent set-default <claude|codex>

gnomon init
gnomon describe
gnomon bootstrap

gnomon spec discover
gnomon spec create <title>
gnomon spec define <SPEC-id>

gnomon approve <SPEC-id>
gnomon revoke <SPEC-id>

gnomon implement <SPEC-id>
gnomon test [SPEC-id]
gnomon verify [target]
gnomon review [target]
gnomon finalize

gnomon status
gnomon validate
gnomon next

gnomon run <workflow-identity> [target]
```

---

## Command Mapping

| Command | Invokes | Character |
|---|---|---|
| `agent` | Shows the persisted default Agent provider and available providers | CLI-native, no project required |
| `agent set-default <claude\|codex>` | Explicitly changes the persisted default Agent provider | CLI-native, Human-owned, no project required |
| `init` | Deterministic filesystem setup (Step 3) | CLI-native |
| `describe` | Initial Knowledge Establishment | Agent workflow |
| `bootstrap` | Bootstrap | Agent workflow |
| `spec discover` | Specification Discovery (Agent reasoning mode) | Agent workflow |
| `spec create <title>` | Draft Creation (Step 1's deterministic carve-out) | CLI-native, no Agent |
| `spec define <SPEC-id>` | Specification Definition | Agent workflow |
| `approve <SPEC-id>` | The Approval operation (Step 6) | CLI-native, Human-owned |
| `revoke <SPEC-id>` | The Revocation operation (Step 6) | CLI-native, Human-owned |
| `implement <SPEC-id>` | Implementation | Agent workflow |
| `test [SPEC-id]` | Testing (target optional, per its conditional Contract) | Agent workflow |
| `verify [target]` | Verification (target generic/untyped, per its Contract) | Agent workflow |
| `review [target]` | Review (target generic/untyped, per its Contract) | Agent workflow |
| `finalize` | Git Finalization | Agent workflow |
| `status` | See below | CLI-native |
| `validate` | See below | CLI-native |
| `next` | See below | CLI-native |
| `run <workflow-identity> [target]` | Any workflow, by declared identity | See "The `run` Rule" |

`describe` — this name was chosen deliberately over Core's internal name ("Initial Knowledge Establishment"): a bare verb needs to be self-evident to a new Human without requiring knowledge of Core's own vocabulary, and "describe my project" directly matches what the workflow actually does — capture intent, reconcile it against what exists, record it.

`approve`/`revoke` are kept as **top-level** verbs, not nested under `spec`, specifically to give them visual prominence equal to the workflow verbs — reinforcing, not diminishing, their status as the one Human-exclusive action set in the whole system. Neither is ever triggered automatically by an Agent result; a `READY_FOR_APPROVAL` outcome may, at most, produce a same-session suggestion to run `approve`, never execute it.

`agent` is the first top-level noun with both a direct action (`gnomon agent` alone shows current state) and a subcommand (`agent set-default`) — unlike `spec`, which is purely a parent with no bare behavior of its own; this is a deliberate, minimal exception, not a silent departure from that pattern. `agent` and `agent set-default` are also the first commands that require no Gnomon project at all: which Agent provider to use is a Human/installation-level concern, not a project-level one, so both work from any directory, with or without `.gnomon/` present.

Every Agent-invoking command (`describe`, `bootstrap`, `spec discover`, `spec define`, `implement`, `test`, `verify`, `review`, `finalize`, `run`) additionally accepts a shared `--agent claude|codex` flag, overriding the resolved provider for that one invocation only — it never modifies the persisted default `agent set-default` controls. Deterministic commands never accept it, since they never touch an Agent at all.

---

## The `run` Rule

`run` is **technically capable of invoking any workflow**, built-in or custom — Workflow Contract v1 does not distinguish built-in from custom at execution time, only at which verb conventionally routes to it, so no artificial restriction is introduced.

But `run` is **documented and taught only as the path for workflows without a dedicated command**: custom, project-defined workflows, and the rare manual/cold invocation of Knowledge Resolution (its ordinary path remains an invisible, same-session continuation the Agent enters on its own reasoning, never a Human-issued command). Built-in workflow verbs (`implement`, `test`, `verify`, `review`, `finalize`, `bootstrap`, `describe`, `spec define`) remain the sole primary documented path for their respective workflows; `run` is never presented as an alternative to them in help output or documentation, even though it would technically work. This resolves duplicate-path confusion and unnecessary identity-string exposure through documentation scoping, not through an architectural restriction the underlying mechanism doesn't need.

---

## `status`

Answers: **"What is true about my current Gnomon project?"**

At minimum, drawn directly from the Derived Facts already defined in `DERIVED_FACTS.md`:

- Whether the project is initialized.
- Global `CONTRACT_VERSION` compatibility.
- The list of Specifications and each one's derived Draft/Approved state.
- Whether the working tree has uncommitted changes.
- Concise, relevant diagnostics — for example, refusing cleanly and pointing at `validate` if something foundational is broken (an unsupported contract version), without attempting a deep structural audit itself.

`status` is a snapshot of project/work state, not a validation tool. It does not enumerate every workflow's frontmatter, does not audit Result Contracts, and does not become a dashboard beyond what a Human needs to see "where my work currently stands."

## `validate`

Answers: **"Is the Gnomon/Core/Contract structure itself valid?"**

Conceptually audits:

- `.gnomon/` structural integrity.
- Whether `CONTRACT_VERSION` is supported.
- Every workflow's frontmatter validity (Step 1's own classification table).
- Duplicate workflow identities.
- Workflow Contract validity generally.
- Result Contract validity (Step 5's schema-declaration requirements).
- Approval evidence structural validity (Step 6's record shape).

`validate` is distinct from `status` by design, not overlapping: `status` answers a question about *work*, `validate` answers a question about *tooling/installation integrity*, and it is built for clean pass/fail semantics suited to CI and deliberate troubleshooting — a mixed-purpose `status` output would be a poor fit for a script needing to distinguish "3 Specifications are Draft, that's normal" from "a Contract is malformed." Exact output format is not defined here.

## `next`

Answers: **"What could/should I do next?"**

Preserves Step 8's two-mode distinction exactly:

- **Same-session, high-confidence guidance** is not `next`'s job — it is trailing output automatically printed after whatever command just produced a result (for example, `spec define` reporting `READY_FOR_APPROVAL` ends with a suggestion to run `approve`).
- **`gnomon next` itself provides only cross-session, coarse, repository-derived guidance** — a list of currently valid actions (Step 1 gates + Step 7 facts), annotated with generic suggestions. It never claims to remember a specific past Agent result or any notion of workflow progress, because none is ever stored.

---

## Naming and UX Principles

- Domain-oriented commands are the primary UX; a Human should be able to predict most command names from having read the User Journey narrative alone, without learning Core's internal workflow identities.
- `run` is intentionally secondary, per the rule above.
- Knowledge Resolution has no dedicated public verb, reflecting its actual, established shape as a same-session continuation rather than a routinely Human-issued command.
- The command surface itself must never imply a mandatory pipeline: Testing, Verification, Review, and Git Finalization are commands like any other, invoked whenever the Human judges them useful — nothing in their naming, grouping, or presentation suggests they must run in sequence or run at all for a given change.

---

## Gating Behavior

Every command that targets a workflow resolves its target and passes through the same deterministic pre-start eligibility check (Step 1's gate, evaluated against Step 7's facts) **before** any Agent is started. For example:

```text
gnomon implement SPEC-003
```

against a Draft `SPEC-003` refuses before the Agent starts, and reports: which gate failed (the Specification is not Approved); the current fact (its derived lifecycle state); and valid next actions, presented with the same cross-session honesty `next` uses — offering both `spec define` and `approve` as possibilities without presuming which applies, rather than guessing. This is the same guidance logic `next` uses, triggered by a blocked command instead of an explicit request. Exact wording and terminal styling are not designed here.

---

## Process Exit Codes

Gnomon's process exit code follows the ordinary Unix convention, and only that convention:

- `0` — the requested command completed successfully (`Success`).
- non-zero — the requested command did not complete successfully, for any reason: a `Blocked` workflow/gate result, a `Cancelled` run, a genuine `Failed` process/protocol error, or an invalid invocation (bad arguments, unknown command).

This is v1's entire exit-code contract: two tiers, not a taxonomy. It is documented here because it already matches the CLI's actual behavior (`cmd/gnomon/root.go`'s `renderReport`/`Execute` — any non-nil error exits `1`), not because a richer scheme was designed and is being recorded. No authoritative Step 1–13 document defines a process-level distinction between a `Blocked` workflow result and a genuine `Failed` one — Steps 5 and 8 are explicit that this distinction matters for *presentation* (a Human must never read `Blocked` as a crash) and for *Result Protocol classification*, but neither states it must also produce a different OS exit code. A script that needs to tell "the work isn't done yet" apart from "something is actually broken" should parse the rendered output (or, in a future machine-readable mode, structured output — not designed in v1) rather than the exit code alone.

A richer exit-code taxonomy (for example, a distinct code for `Blocked` versus `Failed` versus invalid usage) is a legitimate future enhancement, but is a new public CLI contract — once published, scripts and CI pipelines could reasonably depend on it, so it is not something to introduce silently during a release gate. It is deferred below, not decided here.

---

## Deferred to Future Work

- CLI framework, argument parser, command handlers, and internal package structure — CLI Step 10 and later.
- Agent provider selection and configuration surface — Step 4's own deferred scope.
- Short aliases for any command — a later UX-polish decision, not this Step's architecture.
- Exact `validate`/`status`/error output formatting.
- A richer, multi-value process exit-code taxonomy distinguishing `Blocked`/`Cancelled`/`Failed`/invalid-usage at the OS process level, beyond the two-tier `0`/non-zero convention documented above.

---

## Notes

This document persists the Step 9 architecture finalized through prior design analysis and its subsequent refinement — specifically, `establish` was renamed to `describe` for self-evidence, and `validate` was added as a distinct command from `status` after determining the two answer genuinely different questions. It does not reopen Core or Steps 1–8.
