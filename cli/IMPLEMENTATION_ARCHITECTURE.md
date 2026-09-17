# Implementation Architecture — CLI Step 10

## Purpose

Derive the smallest maintainable implementation architecture for Gnomon CLI v1 from the already-finalized design (Steps 1–9), before writing production code.

This is CLI-level design, not Core semantics, and not code. It names the technology, package structure, dependency direction, and testing strategy that faithfully implements Steps 1–9; it defines no Go code, no module, no dependencies, and no new runtime state.

---

## Scope

This document covers:

- The technology choice (Go, Cobra, standard library) and why it fits this workload specifically.
- The package structure, each package's responsibility, and allowed dependency direction.
- The one genuinely required interface boundary and why every other candidate boundary was rejected.
- The v1 configuration model (none beyond a single environment-variable override).
- The JSON Schema validation approach for Result Contract payloads.
- Representative command flows, traced through the proposed structure.
- The testing strategy and the infrastructure explicitly rejected as premature.

This document does not cover:

- Any change to Core semantics or to Steps 1–9's own conclusions — all referenced here, none redefined.
- Actual Go code, module initialization, or dependency selection beyond naming the class of library needed.
- Exact library names, environment-variable names, or Go module paths — deliberately left as implementation details.

---

## Preconditions Carried Forward (Steps 1–9, unchanged)

- **Workflow Contract v1 / Result Contract** (Step 1): every workflow's frontmatter carries eligibility fields plus a Result Contract (`terminal_path`, `schema`), loaded generically — never hardcoded per workflow.
- **Core Distribution / Project Initialization** (Steps 2–3): `.gnomon/` is the sole runtime authority, materialized by `gnomon init`.
- **Agent Adapter** (Step 4): Terminal Handoff is the default strategy; `prepare/run/cancel/status` is the entire Adapter surface; no Gnomon semantics belong inside it.
- **Result Protocol** (Step 5): a transient, invocation-specific JSON envelope (`workflow`, `run_id`, `payload`), consumed once, then deleted.
- **Approval Runtime** (Step 6): lifecycle is always derived, never stored; evidence is durable, append-only, under `.gnomon/approvals/`.
- **Derived Facts** (Step 7): every fact recomputed fresh from repository truth, including the working-tree-changes fact.
- **User Journey** (Step 8): no persisted progress; valid vs. recommended actions, same-session vs. cross-session guidance.
- **Command Surface** (Step 9): the finalized command hierarchy, `describe`/`validate` included, `run` as the documented-secondary escape hatch.

Nothing in this document changes any of these.

---

## Technology

**Go**, with **Cobra** for command routing, standard library wherever practical, and mature third-party Go libraries for YAML/frontmatter parsing and JSON Schema validation rather than implementing either standard ourselves. Exact library choices remain implementation details, deferred past this document.

### Why Go over Rust and TypeScript

- **Distribution**: Go and Rust both produce trivial single, statically-linked binaries; Node/TypeScript would require bundling a runtime or a Node install, a real distribution weakness for a tool meant to be simply installed and run — independent of Claude Code itself happening to be Node-based, since Gnomon only ever spawns it as an opaque subprocess, never links its runtime.
- **Terminal Handoff** (Step 4's central mechanism): Go's `os/exec`, with `Stdin`/`Stdout`/`Stderr` inherited directly, realizes Terminal Handoff almost for free.
- **Workload fit**: Gnomon CLI reads files, makes deterministic decisions, and waits on a subprocess — it is not memory-safety- or performance-critical in the way Rust's ownership model is built to reward. That rigor is a real cost here (slower iteration, steeper onboarding) without a corresponding benefit for this specific workload.
- **Testability**: Go's interface-based mocking maps directly onto the one interface this architecture actually needs (`adapter.Adapter`), with no framework required.
- **Contributor accessibility**: Go's shallow learning curve and broad familiarity matter concretely given Steps 1/9's own extensibility goals — custom workflows and future adapters are meant to be approachable to contribute.
- **Startup overhead**: Go's near-instant startup suits a CLI whose deterministic commands (`status`, `validate`, `next`) should feel immediate; Node's VM warmup is a real, if modest, cost Go avoids.

JSON Schema tooling maturity favors Node's ecosystem somewhat, but Step 1/5 deliberately restricted the Result Contract vocabulary to `type`/`enum`/`required`/`properties`/`items` — a small subset any mature Go library already handles reliably, narrowing what would otherwise be Node's strongest advantage to the point of not being decisive.

---

## Package Structure

```text
cmd/
  gnomon/

internal/
  project/
  contract/
  specs/
  approval/
  facts/
  adapter/
  agentconfig/
  result/
  orchestrate/
  gitutil/
```

### Responsibilities and Dependency Direction

| Package | Responsibility | May depend on | Must not know about |
|---|---|---|---|
| `cmd/gnomon` | Cobra command definitions, argument parsing, output rendering — thin UI wiring only | `orchestrate` | Every domain package directly, `adapter`, `contract` internals |
| `orchestrate` | Application-level sequencing for every Step 9 command; the only package that knows all domain packages together | `contract`, `facts`, `adapter`, `agentconfig`, `result` | Cobra, any CLI-framework type |
| `project` | Project-root discovery, `.gnomon/` layout knowledge, `init` materialization | standard library only | `specs`, `approval`, `adapter`, `contract` internals |
| `contract` | Load and validate a workflow's frontmatter — both eligibility fields and Result Contract, from one parse | `project` | `facts`, `adapter`, `approval` |
| `specs` | Specification existence, dependency-list extraction, next-identity allocation | `project` | `adapter`, `contract` |
| `approval` | Fingerprinting, evidence read/write, Draft/Approved derivation | `project`, `specs`, `gitutil` | `adapter`, `contract`, `result` |
| `facts` | Composes all Derived Facts, including pre-start eligibility | `project`, `specs`, `approval`, `contract`, `gitutil` | `adapter` |
| `adapter` | The Adapter interface (`prepare`/`run`/`cancel`/`status`) and every concrete Agent provider implementation, resolved by name (`Resolve`) | standard library process/terminal primitives only | `facts`, `contract`, `approval`, `specs`, `agentconfig` — no internal Gnomon package |
| `agentconfig` | Persists and loads the Human's chosen default Agent provider — one field, at a per-user, OS-standard location, entirely independent of any project | standard library only | Everything else |
| `result` | Transient Result Protocol: envelope handling, transient-directory lifecycle, payload validation against a loaded Result Contract | `project`, `contract` | `adapter`, `approval` |
| `gitutil` | Thin wrapper shelling out to the `git` binary for the few operations actually needed (working-tree changes, identity) | standard library only | Everything else |

No cycles exist in this graph. `adapter` sits at the bottom with zero dependency on any other internal package — the structural mechanism that enforces "no Gnomon semantics in the Adapter," not merely a stated convention.

---

## `orchestrate` — Internal Organization

`orchestrate` remains **one package**, never split into per-command packages or per-domain "use case" sub-packages — both would be Clean Architecture/DDD ceremony this architecture deliberately avoids, and neither is warranted: every command shares one cohesive responsibility (sequence domain packages to fulfill a Step 9 command), exercised many times, not several unrelated responsibilities. Multiple files within one package is ordinary, idiomatic Go, not a workaround:

```text
orchestrate/
  init.go       — gnomon init
  spec.go       — spec discover / spec create / spec define
  approval.go   — approve / revoke
  agent.go      — Agent provider resolution (ResolveAgent), gnomon agent, gnomon agent set-default
  workflow.go   — the shared "load contract → check eligibility → resolve Agent → run adapter → consume result" sequence, reused by implement/test/verify/review/finalize
  render.go     — generic Result Contract payload rendering, driven entirely by a loaded Contract's own classification and declared schema property order, never a hardcoded field name
  validate.go   — gnomon validate
  status.go     — gnomon status
  next.go       — gnomon next
```

`validate.go`/`status.go`/`next.go` were built as three separate files rather than the single `guidance.go` this section originally sketched: each landed in its own CLI Step (9 through 12, Slices 5 and 6), at different times, and each is a self-contained, independently testable responsibility — ordinary Go file-per-command-family organization, the same granularity `init.go`/`spec.go`/`approval.go`/`agent.go` already use, not a departure from "`orchestrate` remains one package."

`workflow.go`'s shared sequence exists specifically to prevent the five workflow-invoking commands from duplicating the same load/gate/run/consume logic, which is the actual mechanism preventing this package from becoming unwieldy as the command surface fills in — not additional package boundaries. `render.go` exists for the analogous reason on the output side: every workflow-invoking command needs the same generic "classify the terminal value, render the remaining fields" logic, driven by `contract.ResultContract` alone, never by a per-workflow Go switch.

---

## `contract` — One Package, Two Types

Workflow eligibility metadata and the Result Contract are **not** split into separate packages, despite serving different purposes at different times (gating before an Agent starts, payload validation after it exits). The reason: both live in the same frontmatter block of the same file, read by the same parse, at the same moment — the actual test for package cohesion is shared data lifecycle, not which downstream package consumes which field. Splitting would force either duplicated file-parsing across two packages, or one package awkwardly importing the other's intermediate struct for what is really one loading operation.

The distinction the two purposes deserve is expressed as **two clearly-named types from one loader**, not two packages — a workflow's loaded contract carries a distinct eligibility structure (identity, `specification_reference`, `requires_approved_specification`) and a distinct Result Contract structure (`terminal_path`, `schema`); `facts` reads the former, `result` reads the latter, from the same already-loaded value.

---

## Interface Boundaries

**`adapter.Adapter` is the only interface this architecture requires.** It exists because Step 4 explicitly designed for multiple future Agent providers behind exactly this shape (`prepare`/`run`/`cancel`/`status`), a genuine, evidenced need for substitutability.

No other interface is introduced:

- **Filesystem**: no custom abstraction — Go's standard library (`os`, `io/fs`) is already sufficient and testable (`fstest.MapFS`, `t.TempDir()`); wrapping it further would be abstraction without a real consumer.
- **Git**: `gitutil` is **concrete**, built directly on `os/exec` calling the `git` binary — no interface. It is tested by running against real, temporary Git repositories (`git init` in a `t.TempDir()`), not by substituting a fake. Nothing upstream needs `gitutil` faked either: `facts` and `approval`'s decision logic are unit-tested by injecting already-resolved values directly, never by exercising `gitutil` through a mock.

---

## Configuration

**No project-level configuration file exists in v1.** Claude Code and Codex CLI are both first-class providers, resolved by name through `adapter.Resolve` — the final model:

- **Provider selection** is a persisted, per-user (never per-project) default, plus a per-invocation override:
  - `--agent claude|codex`, available on every Agent-invoking command, overrides the provider for that single invocation only and never changes the stored default.
  - Otherwise, the persisted user-level default is used.
  - If no default is configured yet and stdin is interactive, the Human is asked once to choose, and that choice is persisted as the new default.
  - If no default is configured and interaction is unavailable (non-interactive stdin), the CLI refuses clearly rather than guessing.
  - `gnomon agent` shows the current default and available providers; `gnomon agent set-default <provider>` explicitly changes the persisted default. Neither requires a Gnomon project — provider identity is a Human/installation-level concern, not a project-level one.
  - No `GNOMON_AGENT_PROVIDER` or equivalent environment variable exists — persisted default plus `--agent` is the entire selection surface.
- **Provider-specific executable overrides**: `GNOMON_CLAUDE_EXECUTABLE` and `GNOMON_CODEX_EXECUTABLE`, each read only by its own Adapter's constructor, defaulting to `claude`/`codex` resolved from `PATH`. These are independent, single-purpose overrides — not a general configuration surface.
- **Persistence mechanism**: exactly one small, per-user file, `<os.UserConfigDir()>/gnomon/config.json`, holding exactly one field (`default_provider`), written atomically (temp file plus rename) by `internal/agentconfig`, standard library only. This is not a general preferences system, and not project-owned configuration — it lives outside any project, is never committed, and holds nothing beyond the one field named above. Introducing it satisfies, rather than contradicts, this Step's own earlier reservation: a config file becomes justified once a genuine second provider requiring selection, or a real persistent-preference need, actually materializes — both now have.
- No `.gnomonrc`, `gnomon.toml`, or any project-level provider configuration exists in v1, or is implied by the user-level file above.

---

## JSON Schema Validation

Result Contract payload validation (Step 5) uses a **mature, existing Go JSON Schema library**, never a hand-rolled validator. Only the restricted vocabulary Steps 1 and 5 already fixed needs support — `type`, `enum`, `required`, `properties`, `items` — no draft negotiation, `$ref` resolution, or format validators are needed, so even a minimal library is sufficient. Exact library selection is an implementation detail, deferred past this document.

---

## Representative Flows

**`gnomon init`**: `cmd/gnomon` → `orchestrate.Init` → `project.Locate` (detect existing/partial state) → `project.Materialize` (copy bundle, write `CONTRACT_VERSION` last) → report. No `facts`, `adapter`, or `result` involved.

**`gnomon spec create "Password Reset"`**: `cmd/gnomon` → `orchestrate.SpecCreate` → `specs.NextIdentity` → `specs.CreateDraft` → report. No eligibility check — Draft Creation is never gated.

**`gnomon approve SPEC-004`**: `cmd/gnomon` → `orchestrate.Approve` → `specs.Exists` → `approval.Fingerprint` → `approval.ResolveIdentity` (Git identity, fallback, or refuse per Step 6) → `approval.WriteGrant` → `facts.Lifecycle` (re-derive, confirm) → report.

**`gnomon implement SPEC-004`**: `cmd/gnomon` → `orchestrate.Implement` (via `workflow.go`'s shared sequence) → `contract.Load("implementation")` → `facts.Eligible` (existence + lifecycle + contract-version) → if ineligible, report the blocking fact and valid next actions, **neither `adapter` nor Agent provider resolution is ever touched**; if eligible, `orchestrate.ResolveAgent` (`--agent` override → persisted `agentconfig` default → interactive first-use chooser, CLI-edge-only → refuse if none of those apply) → `adapter.Resolve` (never a concrete type constructed directly) → build the invocation context → `adapter.Prepare` → `adapter.Run` (Terminal Handoff, blocking) → `adapter.Status` → `result.Consume` (validate envelope + payload against `contract`'s already-loaded Result Contract) → report process status, protocol status, and workflow result as three separate facts (Step 5) → `orchestrate` prints same-session guidance if applicable.

**`gnomon agent`**: `cmd/gnomon` → `orchestrate.AgentStatus` → `agentconfig.Load` → `adapter.Providers` → report. No project required; no `adapter.Resolve`, no `contract`, no `facts`.

**`gnomon agent set-default codex`**: `cmd/gnomon` → `orchestrate.AgentSetDefault` → `adapter.Resolve` (validate the name only) → `agentconfig.Save` (atomic) → report. No project required.

**`gnomon status`**: `cmd/gnomon` → `orchestrate.Status` → `project.Locate` + contract-version check → `specs.List` + `facts.Lifecycle` per Specification → `gitutil.HasChanges` → report. No `adapter`, no `result`.

**`gnomon validate`**: `cmd/gnomon` → `orchestrate.Validate` → `project.Layout.MissingSections` → `contract.Load` per workflow file (Step 1's classification table, including Result Contract schema well-formedness, plus duplicate-identity detection) → `approval.ValidateEvidence` (Step 6's failure-case table) → report pass/fail with itemized issues.

**`gnomon next`**: `cmd/gnomon` → `orchestrate.Next` → contract-version check → `specs.List`, then per Specification `facts.Lifecycle` + `facts.Eligible` against the loaded `specification-definition`/`implementation`/`testing` Contracts → `gitutil.HasChanges` for a coarse Git Finalization signal → report every currently valid action, unranked. No `adapter`, no `result`, and no single action is ever chosen among multiple simultaneously-valid candidates.

---

## Testing Architecture

- **Pure unit tests**: `approval` fingerprinting (given raw strings), `specs.NextIdentity` (given a filename list), `contract` parsing/validation (given YAML bytes), `facts` composition (given already-resolved sub-facts) — all decision logic written as pure functions specifically so it is testable without filesystem access.
- **Filesystem integration tests**: `project`, `specs`, `approval`'s evidence I/O, `result`'s transient I/O — against `t.TempDir()`, real filesystem, no mocking.
- **Git integration tests**: `gitutil`, against real temporary Git repositories, not mocked.
- **Agent-adapter tests**: `orchestrate` tested against a fake `adapter.Adapter` returning canned results.
- **Narrow tests for the real Claude Code adapter**: a smaller, separate layer, not part of the main unit-test suite.
- **Full real-Agent end-to-end tests**: deliberately deferred past this Step.

---

## Explicitly Rejected Infrastructure

- **Database** — every fact is a file read or a cheap recomputation; no query/join/transaction need exists.
- **Event bus** — one call path per command; no pub/sub, no independent subscribers.
- **State machine** — foreclosed by Core's own derived-state philosophy across every Step; nothing here has states to transition between.
- **DI framework** — the call graph is shallow enough for manual wiring in `main`; a framework would add indirection with no offsetting benefit.
- **Repository pattern everywhere** — only `adapter` genuinely needs a substitutable interface; forcing one onto `specs`/`approval`/`contract`, each with exactly one real implementation ever planned, is ceremony without benefit.
- **Plugin framework** — custom workflows already work through the Contract mechanism alone (Step 1), with zero CLI code changes.
- **Background daemon** — every command is a short-lived, synchronous invocation.
- **Persistent workflow sessions** — would directly contradict Step 5/8's deliberately ephemeral result and guidance model.
- **Broad configuration system** — provider-specific executable overrides, a persisted per-user default, and a per-invocation flag are the entire v1 configuration surface; no config file holds more than the one field named in Configuration above, and none is project-owned.

---

## Deferred to Future Work

- Exact Go module layout conventions, module path, and file-per-command boundaries beyond the grouping shown.
- Exact YAML frontmatter and JSON Schema library selections.
- Any further provider-specific settings beyond executable path (model selection, credentials) — not needed by either current Adapter.
- A richer, general per-user or per-project preferences system — the one-field provider default in Configuration above remains deliberately narrow.
- Everything already deferred by Steps 1–9.

---

## Notes

This document persists the Step 10 architecture finalized through prior design analysis and its subsequent refinement — specifically, `orchestrate`'s internal multi-file organization, `contract`'s one-package/two-type resolution, `gitutil`'s concrete (non-interface) design, and the no-config-file v1 model were each resolved during that refinement, not left open. It does not reopen Core or Steps 1–9, and it contains no Go code, module initialization, or dependency selection.
