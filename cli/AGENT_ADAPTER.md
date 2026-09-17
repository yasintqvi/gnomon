# Agent Adapter Architecture — CLI Step 4

## Purpose

Define how Gnomon CLI starts and regains control from an external coding Agent — Claude, Codex, or a future compatible Agent — without coupling Gnomon's orchestration, or any Core semantics, to a specific provider.

This is CLI-level design, not Core semantics. It defines execution mechanics and control ownership; it does not change what any workflow means, requires, or does, and it does not define how a terminal structured result is transported or parsed — that is CLI Step 5.

---

## Scope

This document covers:

- The final v1 execution lifecycle, from Human request to control returning to the CLI.
- Responsibility boundaries between the CLI, the Agent Adapter, the Agent process, the Human, and Core.
- Terminal Handoff as the default v1 interaction strategy, and why CLI-mediated interaction was rejected.
- The minimum provider-neutral invocation context the CLI prepares.
- The minimal conceptual Adapter interface.
- Session continuity, the Knowledge Resolution continuation case, cancellation, and failure behavior.
- The exact boundary this Step leaves for CLI Step 5 — the Result Protocol.

This document does not cover:

- Workflow Contract v1's field semantics — defined by CLI Step 1 (`WORKFLOW_CONTRACT.md`).
- The Bundled Default Core / Project-local Core distribution model — defined by CLI Step 2 (`CORE_DISTRIBUTION.md`).
- `.gnomon/` materialization and `gnomon init` behavior — defined by CLI Step 3 (`PROJECT_INITIALIZATION.md`).
- The Result Protocol's transport, schema, or parsing rules — CLI Step 5.
- Agent configuration storage, command hierarchy, interactive Gnomon UX, exact signal/timeout values, GUI/IDE interaction, multi-Agent routing, or any provider marketplace/plugin system — all later or explicitly deferred concerns.
- Approval-evidence persistence — defined by CLI Step 6 (`APPROVAL_RUNTIME.md`).

---

## Preconditions Carried Forward (Steps 1–3, unchanged)

- **Workflow Contract v1** (Step 1): every Agent workflow declares `identity`, `specification_reference` (`none | optional | required`), and `requires_approved_specification` in frontmatter, checked by the CLI before the Agent starts.
- **Core Distribution** (Step 2): the CLI ships a Bundled Default Core; after `gnomon init`, the project-local Core under `.gnomon/` is the **sole runtime authority**; the Bundled Default Core is never consulted during a workflow run.
- **Project Initialization** (Step 3): `gnomon init` is a deterministic filesystem operation only; it invokes no Agent and configures no Agent provider.

Nothing in this document changes any of these.

---

## Final Execution Model

```text
Human
  ↓
Gnomon CLI
  ↓
deterministic pre-run validation
  ↓
Agent Adapter ─── starts Agent, fully interactive ───┐
                                                       │
                                    External Agent ↔ Human
                                                       │
                         (Gnomon concurrently watches for
                          a valid terminal result — the
                          transient file, never the
                          conversation — while this runs)
                                                       │
                          valid result detected ───────┘
  ↓
Adapter terminates the Agent gracefully, waits for exit
  ↓
Gnomon CLI regains control
  ↓
structured result processing
```

During the Agent's run, Human interaction is directly between the Human and the Agent, in the Agent's own normal interactive UI. **Gnomon does not observe, classify, or mediate that conversation — it never parses terminal output.** What Gnomon *does* do concurrently is watch, out of band, for the Result Protocol's transient file to appear and validate; that file is both the workflow's result transport (Step 5) and the signal the CLI uses to know the run is complete. The CLI's own involvement — ending the session and resuming orchestration — begins only once that signal is confirmed, never based on anything read from the terminal.

---

## Responsibility Boundaries

### Gnomon CLI

Owns orchestration: locating the Gnomon project (`.gnomon/`, per Step 3's root discovery); reading the project-local Core; resolving the selected workflow; validating its Workflow Contract metadata (Step 1); enforcing the deterministic pre-start gates that metadata implies (target existence, approval where required); constructing the provider-neutral invocation context (below), including how the Adapter should detect completion; selecting an Agent Adapter; detecting workflow completion by validating the Result Protocol's transient file while the Agent may still be running; terminating the Agent once that result is confirmed valid; receiving process/runtime completion information from the Adapter once the run ends; consuming the structured terminal result (Step 5); and continuing orchestration — presenting the next recommended action — afterward. Provider mechanics never belong to this layer.

### Agent Adapter

Owns provider-specific execution mechanics only: preparing the provider-specific invocation from the CLI's provider-neutral context; launching the Agent with the resolved project root as its working directory; attaching or inheriting terminal interaction where the provider has a native interactive mode; keeping the Agent process/session alive; concurrently invoking a CLI-supplied completion check while the process runs, entirely opaque to the Adapter itself (see Adapter Abstraction below); gracefully terminating the process once that check reports success, and waiting for the real exit; supporting cancellation; and reporting process/runtime completion status once the run ends. The Adapter contains no Gnomon workflow semantics — it never interprets an `Outcome` value, a lifecycle state, a dependency fact, or the completion check's own logic, which is injected as an opaque function, not implemented here.

### Agent

Owns reading the selected project-local workflow file and whatever other project-local `.gnomon/` knowledge and project artifacts it needs; inspecting and modifying the project where the workflow's own rules permit; reasoning according to Core; asking the Human questions directly and receiving decisions directly, in its own native interactive form; continuing that reasoning within the same run, including transitioning between Core workflows when Core's own text calls for it (see Knowledge Resolution below); and producing the terminal structured result Core's own Outputs section for that workflow requires.

### Human

Interacts directly with the active Agent while terminal control is handed off — answering questions, granting authorizations, responding to the Agent's own native prompts — with no Gnomon layer standing between them.

### Core

Remains entirely provider-neutral. No workflow document names or depends on a specific Agent product; every workflow describes reasoning and Human interaction in terms of "the Agent" and "the Human" only.

---

## Terminal Handoff

Terminal Handoff is the default, and only, v1 interaction strategy — the Human's normal interactive Agent UI remains available for the run's entire duration, never replaced by a non-interactive or headless mode:

```text
Gnomon
   ↓
Agent ↔ Human   (Gnomon concurrently watches the Result Protocol's file, not this conversation)
   ↓
Gnomon
```

Gnomon does not attempt to classify live Agent output into progress, questions, answers, tool calls, permission prompts, or reasoning events, and never parses stdout/stderr to determine anything. Provider-native interactive UX — permission prompts, tool confirmations, progress rendering, interactive selectors, provider-native commands, provider-native cancellation UX — remains entirely provider-owned and is never modeled, proxied, or recreated by Gnomon. Gnomon does not become a second Agent UI or a second permission system. Watching the Result Protocol's transient file is not an exception to this — it is a separate, out-of-band signal (Step 5), never terminal content.

---

## Working Directory and Core Access

The Agent runs with its working directory set to the consuming project's resolved root — the same root `.gnomon/` root discovery (Step 3) locates — regardless of the Human's own current directory at invocation time. From there the Agent has direct, ordinary filesystem access to the project's implementation, the project-local `.gnomon/` Core, the selected workflow file, project knowledge, and Specifications and other relevant artifacts. The project-local Core remains the sole runtime authority exactly as Step 2 defines; the Bundled Default Core is never injected or duplicated into the Agent's context.

---

## Invocation Context

The CLI prepares a minimum, provider-neutral invocation context, containing only what is actually needed to start the workflow, as references rather than duplicated content:

- The resolved project root.
- The Gnomon root (`.gnomon/`).
- The selected workflow's `identity` (Step 1) — the Adapter resolves the actual file from this plus the known Gnomon root.
- The supplied Specification identity, when `specification_reference` isn't `none` and one was given — the bare identity, never the Specification's content.
- Explicit asserted session context, when required (see Knowledge Resolution below).
- The Human's initiating instruction, where applicable.
- A slot for whatever result-delivery instruction CLI Step 5 defines the Agent must follow to produce its terminal result — its exact content is not decided here.
- An opaque completion-check function the Adapter invokes periodically while the Agent runs, letting it detect and act on a valid result without ever knowing what "valid" means (see Adapter Abstraction below). May be omitted, in which case the Adapter simply waits for the process to exit on its own.

---

## Adapter Abstraction

The smallest conceptual v1 interface, expressed as operations, not a language-specific API:

### `prepare(context)`

Resolves provider-specific invocation mechanics from the provider-neutral context above — the one place provider-specific translation happens.

### `run()`

Starts the Agent and hands over or inherits terminal interaction for the run's duration. If a completion-check function was supplied, concurrently invokes it at a short, fixed interval while the process runs; the instant it reports a valid result, `run()` gracefully terminates the process and waits for the real exit before returning. If the process ends on its own first — a Human ending the session directly, or a crash — that natural exit is reported as-is, with nothing further attempted. If no completion-check function was supplied, `run()` simply blocks until the process exits, exactly as originally specified. Either way, `run()` remains a single blocking call from the caller's perspective.

### `cancel()`

Requests termination of the active Agent run.

### `status()`

Exposes provider/runtime completion information after termination — clean exit, non-zero exit/runtime failure, Human cancellation, or forced termination.

**`send` and `receive` conversational operations are not part of this interface.** They existed only to support mediating the Human↔Agent conversation, and Terminal Handoff requires no such mediation.

---

## Session Continuity

Within one Agent workflow run, the same Agent process/session is preserved for its entire duration:

```text
Implementation starts
→ Agent works
→ Agent asks Human
→ Human answers
→ Agent continues
→ Agent finishes
```

Gnomon does not terminate and recreate the Agent around each Human interaction — this follows automatically from Terminal Handoff, since `run()` simply remains blocking while the Agent and Human converse directly. No persisted cross-run Agent session state is introduced. A separate workflow invocation may start a fresh Agent session unless a future requirement proves otherwise.

---

## Knowledge Resolution

Two conceptually valid cases, neither requiring adapter complexity beyond what is already defined:

### Same-run continuation

An Agent may enter Knowledge Resolution reasoning during the same active session — exactly as Specification Definition's own text describes ("...run [`workflows/knowledge-resolution.md`](../workflows/knowledge-resolution.md)... before resuming") — and later return to the originating workflow's reasoning. The Adapter does not need to observe this transition; from its perspective the run simply continues, since it was never observing conversation content in the first place.

### Separate invocation

A future orchestration path may instead end the run and launch Knowledge Resolution as a separate Agent invocation, supplying the reported gap and decision as explicit asserted session context (per Invocation Context above). This, too, requires no live conversation mediation — only the reported result and current repository state.

Knowledge Resolution's own semantics are unchanged by either case.

---

## Process/Runtime Completion Information

The CLI must be able to distinguish, from the Adapter's `status()`: the process exited cleanly; the process failed or crashed; the process was cancelled by the Human; the process was forcibly terminated; and whether a terminal structured result was, or was not, delivered at all.

`status()` separately reports whether *Gnomon itself* asked the process to stop — a valid result confirmed while the Agent was still running (Cancellation and Signals, above), or an explicit `cancel()` — as opposed to the process ending on its own. This is what lets a presentation layer describe a Gnomon-initiated termination that followed a confirmed valid result as the success it is, rather than surfacing its underlying signal/exit code as if the Agent had failed; the raw exit code remains available for diagnostics regardless.

**This is provider/runtime status, never a Core workflow `Outcome`.** A clean process exit does not itself mean the workflow's Outcome was, for example, `IMPLEMENTATION_COMPLETE` — that determination belongs entirely to the Result Protocol (Step 5) and the terminal result it defines.

---

## Boundary With Step 5

Step 4 defines how the Agent is started, how it interacts with the Human, and how control returns to the CLI. **Step 5 defines how the Agent returns a deterministic structured workflow result that the CLI can validate after the run.**

Step 4 imposes exactly one requirement on Step 5: the Result Protocol must work **without requiring the CLI to parse or observe the interactive terminal conversation**, at any point — Step 5's mechanism may now be consulted *while* the Agent is still running (that is precisely how `run()` detects completion), but only ever by examining its own transient artifact, never the terminal stream. It must also make "a valid result was produced" and "no result was produced" unambiguously distinguishable. Step 4 does not choose the result's transport or schema.

---

## Cancellation and Signals

The CLI must be able to regain control from an active Agent process at any time — either because the Human cancels directly, or because `run()`'s own completion check reports a valid result. Both paths terminate gracefully first (a plain termination signal to the single Agent process, giving it a chance to restore terminal state and clean up its own children exactly as it would on any other exit path) and escalate to a forced kill only after a bounded grace period with no response. No exact timeout value is fixed here — that remains an implementation detail — but the two-step shape (graceful, then forced) is part of this Step's model, not left open.

Terminal signals a Human generates directly (for example Ctrl+C) should normally reach the foreground Agent naturally, through ordinary process/terminal behavior, without any Gnomon involvement.

**PTY allocation and process-group management were both considered and deliberately not adopted.** Interactive access works today by directly inheriting the CLI's own stdio into the Agent process — no PTY is needed for that, since Gnomon never intercepts or rewrites the stream. Moving the Agent into its own process group (to reach any subprocess it spawns when terminating) was considered for cleanup completeness, but rejected: doing so would move the Agent out of the controlling terminal's foreground process group, risking Ctrl+C and similar terminal-generated signals no longer reaching it correctly for the run's duration — a real regression to the interactive experience this Step exists to preserve, for a speculative benefit. A single, direct signal to the one Agent process is smaller and carries no such risk; an Agent's own responsibility for its child processes on any termination path (a Human's Ctrl+C included) is unchanged by how Gnomon asks it to stop.

---

## Provider Neutrality and Compatibility

The outer Adapter contract — `prepare → run → completion`, with cancellation and status support — remains provider-neutral. **Terminal Handoff is the default strategy for an interactive CLI Agent, not the universal provider protocol.** A future provider's Adapter may internally use a native CLI process, a PTY, an SDK, an API, batch execution, or another provider-specific mechanism to satisfy the same outer lifecycle. Core never depends on, or names, any provider or its capabilities.

---

## Batch/Non-Interactive Compatibility

The outer abstraction does not require a real terminal. For a provider with no native interactive UX, `run()` may execute without any terminal handoff at all and simply block until completion — a strict subset of the same lifecycle, not a second architecture. One common `prepare/run/cancel/status` abstraction covers both interactive and batch Adapters.

---

## First v1 Adapter

CLI v1 begins with one real Adapter, built and dogfooded against **Claude Code** — the tool this Gnomon design process has itself been conducted through — while the provider-neutral Adapter boundary above is preserved from day one, so a second Adapter is a pure addition, never a refactor. No fallback Agents, routing, automatic model selection, or multi-Agent collaboration are designed here.

A second Adapter, **Codex CLI**, was subsequently added exactly as this section anticipated: a pure addition sharing the same `prepare/run/cancel/status` lifecycle and the same Terminal Handoff mechanism, with `adapter.Adapter`, `Context`, and `Status` all unchanged. Both are resolved by name through one function (`adapter.Resolve`); orchestration never constructs either concrete type directly. Which provider a given invocation uses — an explicit per-invocation override, or a persisted per-user default — is a CLI Step 9/10 concern, not part of this Step's own boundary.

---

## Security Boundary

Agent repository permissions remain governed entirely by the Agent's own provider/runtime environment. Provider-native permission prompts remain fully visible to, and answered directly by, the Human. **Gnomon does not introduce a second sandbox or permission model in CLI v1**, and must not silently weaken, obscure, or bypass any provider-native permission behavior by attempting to proxy or recreate it.

---

## Failure Behavior

Without introducing any persisted workflow-execution state:

- **Agent executable unavailable** or **authentication/runtime failure** — `prepare()`/`run()` fails cleanly before anything touches the repository; the CLI reports the failure.
- **Agent crash**, **exit without a terminal result**, or **provider/runtime failure** — surfaced via `status()`; the repository is left exactly as the Agent left it.
- **Human cancellation** or **forced termination** — surfaced via `status()` as such, distinct from an ordinary failure.

These are invocation/runtime failures, never new Core workflow `Outcome` values. The repository remains the durable source of truth: a later invocation re-inspects current repository state and proceeds — exactly Core's own existing derived-state discipline, applied here to process-level failure as well as content-level failure — rather than resuming from any Gnomon-maintained `RUNNING`/`INTERRUPTED` status.

---

## Why CLI-Mediated Interaction Was Rejected for v1

Not because it was judged "too complex" in the abstract, but for a specific architectural reason: **Core requires the Agent to interact directly with the Human — it never requires the CLI to observe or mediate that interaction.** Every workflow's own text assigns asking and waiting to the Agent, and answering to the Human; nothing in any workflow's Outputs or Completion Criteria depends on the CLI having witnessed the exchange. Mediation would have introduced provider-specific event translation, question detection, message routing, input plumbing, and a duplicated, Gnomon-authored version of each provider's own interactive UX — with a genuine risk of interfering with a provider's own permission flows — without satisfying any additional requirement Core actually states. This may be reconsidered for a future GUI/IDE environment, which has no terminal to hand off and would need an internally-mediated Adapter strategy of its own — but that would be a new Adapter implementation choice behind the same outer lifecycle, not a change to Core semantics.

---

## Deferred to Future Work

The following are intentionally not defined here, and nothing in this document depends on them:

- The Result Protocol's transport and schema (CLI Step 5).
- Approval-evidence persistence (CLI Step 6).
- Agent configuration storage beyond provider identity and per-provider executable path (model, credentials, richer per-provider settings) — provider identity and executable path are now covered by `internal/agentconfig` and the provider-specific executable-override variables (CLI Step 9/10).
- CLI command hierarchy and interactive Gnomon UX.
- Exact signal/timeout/escalation values.
- A GUI/IDE interaction model, and any internally-mediated Adapter strategy it would require.
- Multi-Agent routing, fallback Agents, and automatic model selection.
- Cross-run session persistence and background Agents.
- Any provider marketplace or plugin system.

---

## Notes

This document persists the Step 4 architecture already finalized through analysis and challenge in prior design discussion — including the explicit reconsideration of an earlier CLI-mediated recommendation — and does not reopen that decision. It does not restate or redesign CLI Steps 1–3, referenced above only as context.

A later revision finalized how `run()` actually regains control: not by requiring the Agent to end its own session (unreliable in practice — an in-conversation instruction cannot make an interactive session's own harness terminate the process), and not by running the Agent headless (which broke interactive Human↔Agent use and, separately, blocked the Agent's own tool use behind unanswerable permission prompts). Instead, `run()` watches the Result Protocol's transient file out of band while the Agent stays fully interactive, and terminates the process itself once a valid result is confirmed. This is recorded above as the Adapter Abstraction's `run()` behavior, the Cancellation and Signals section's PTY/process-group decision, and the Final Execution Model — not as a new Step, since it fits entirely within the `prepare/run/cancel/status` shape already specified.
