<p align="center">
  <img src="assets/gnomon.png" alt="Gnomon logo" width="190" />
</p>

<h1 align="center">Gnomon</h1>
<p align="center"><em>AI Software Engineering System</em></p>

<<<<<<< HEAD
## Why Gnomon?

AI coding agents can read your code, implement features, run tests, and make changes quickly. But access to a repository doesn't tell an Agent everything about how the project is supposed to work.
=======
A document-driven system for running AI-assisted software engineering as an explicit process, plus a small Go CLI that runs it. The documents — Gnomon **Core** — separate durable project knowledge, feature requirements, artifact behavior, architectural decisions, execution workflows, and evaluation criteria, so an AI agent loads only what a given task actually needs. The **CLI** (`gnomon`, in [`cmd/gnomon`](cmd/gnomon)) turns Core's rules into an installable tool: it deterministically gates what an Agent is allowed to start, launches Claude Code or Codex to do the work, and validates what comes back — never deciding engineering judgment calls itself.

Core (the Markdown documents this repository is built from) has no runtime of its own and can be read, copied, or followed by hand. The CLI is the recommended way to run it in a real project — see [Gnomon CLI](#gnomon-cli) below.
>>>>>>> afb60ff (docs: introduce the CLI in README and add process exit-code release-gate tests)

Important requirements, previous decisions, unresolved questions, and engineering constraints can live outside the code — often in conversations that disappear from the next session. Without an explicit engineering process, an Agent may have to reconstruct that context or make reasonable assumptions where a Human decision was actually needed.

**Gnomon gives AI coding agents an engineering system to work within.**

It keeps important project knowledge and Specifications with the project, separates Human authorization from Agent execution, and gives engineering workflows explicit rules and structured results.

It doesn't make the Agent smarter. It makes the engineering process around the Agent explicit and durable.

The best way to understand the difference is to use it. The walkthrough below takes a real feature from definition to implementation and verification.

---

## Install and initialize

Right now there's exactly one supported way to get `gnomon`: build it from a clone.

```bash
git clone <this-repository-url>
cd gnomon
go build -o gnomon ./cmd/gnomon
./gnomon --version
````

*(No package manager, curl installer, or prebuilt binary exists yet — that's not a documentation gap, it genuinely isn't built.)*

In an existing project — any existing project, Gnomon doesn't care what's already there — run:

```
$ gnomon init
✓ Gnomon project initialized

→ gnomon spec create "<title>"
```

That materializes a `.gnomon/` directory: a template Specification, the Draft/Approved lifecycle rules, the ten built-in workflows (Implementation, Testing, Verification, Review, and so on), and a version marker. Nothing else in your repository is touched. If you never run another `gnomon` command, this did nothing to your codebase.

## Choosing an Agent

Gnomon drives either **Claude Code** or **Codex** — whichever you already use, or both. The first time a command needs one and none is configured, you'll see an arrow-key prompt:

```
? Choose your default Agent:

  ❯ Claude Code
    Codex

<<<<<<< HEAD
↑/↓ navigate • Enter select
```
=======
## Gnomon CLI

The CLI (`gnomon`) is the recommended way to run the lifecycle above in a real project. It never makes an engineering judgment call itself — it deterministically decides what is *allowed* to start (a Specification's Draft/Approved state, a workflow's own declared prerequisites), launches an Agent provider (Claude Code or Codex) to do the actual work, and validates the structured result that comes back. Everything about *what* to build, define, or change remains the Agent's and the Human's.

### Install / build

There is currently one supported way to get `gnomon`: build it from a local clone with Go 1.24 or later.

```bash
git clone <this-repository-url>
cd gnomon
go build -o gnomon ./cmd/gnomon
./gnomon --version
```

`go install ./cmd/gnomon` also works from inside the clone, installing to your `$GOPATH/bin`. There is no published package-manager, curl-installer, or prebuilt-binary distribution for v1 — only a local build from source.

### First use

```bash
gnomon init                        # materialize .gnomon/ in the current directory
gnomon status                      # confirm the project is initialized
gnomon agent set-default claude    # or: codex — see Agent providers, below
gnomon describe                    # capture initial project knowledge (new projects)
gnomon bootstrap                   # establish the verified project baseline
gnomon spec create "Password Reset"   # or: gnomon spec discover, to have an Agent propose one
gnomon spec define SPEC-001        # bring the Draft to a reviewable, approval-ready state
gnomon approve SPEC-001            # the one Human-exclusive authorization step
gnomon implement SPEC-001          # now eligible, because SPEC-001 is Approved
gnomon test SPEC-001
gnomon verify
gnomon review                      # only when the work actually calls for judgment-based review
gnomon next                        # what's currently valid to do, at any point
gnomon finalize                    # once approved and authorized, hand off to Git
```

`describe`/`bootstrap` matter most for a new project; an existing one may only ever need `spec create`/`spec define`/`approve`/`implement` for each change. Nothing about this sequence is enforced by the CLI as a required pipeline — `gnomon status` and `gnomon next` always reflect current repository truth, never a remembered position in a script, so it is always safe to stop, resume later, or run these in a different order as the work actually calls for it.

### Agent providers

Every command that invokes an Agent (`describe`, `bootstrap`, `spec discover`, `spec define`, `implement`, `test`, `verify`, `review`, `finalize`, `run`) needs to know whether to launch Claude Code or Codex. Resolution order: an explicit `--agent claude|codex` flag for that one invocation, otherwise the persisted default (`gnomon agent set-default <claude|codex>`), otherwise — only in an interactive terminal — a one-time arrow-key prompt whose choice is then saved as the new default. A non-interactive invocation (a script, CI) with no default configured refuses clearly rather than guessing. `gnomon agent` shows the current default and every available provider.

### Main commands

| Group | Commands |
| --- | --- |
| Project | `init`, `describe`, `bootstrap` |
| Specification | `spec discover`, `spec create <title>`, `spec define <SPEC-id>`, `approve <SPEC-id>`, `revoke <SPEC-id>` |
| Engineering | `implement <SPEC-id>`, `test [SPEC-id]`, `verify [target]`, `review [target]`, `finalize` |
| Guidance / Inspection | `status`, `validate`, `next` |
| Advanced | `run <workflow-identity> [target]`, `agent`, `agent set-default <claude\|codex>` |

`status` answers "what state is this project in" (initialization, every Specification's Draft/Approved lifecycle, uncommitted changes); `validate` answers "is the Gnomon structure itself valid" (Contract/evidence integrity, suited to CI); `next` lists every currently valid action across the project, deterministically, without picking or ranking one.

`run` is the generic, Contract-driven path — for a project's own custom workflows, or the rare manual invocation of a built-in workflow by identity. Ordinary work uses the dedicated command (`implement`, `test`, `verify`, `review`, …) instead; both converge on the exact same execution engine.

Every command's full syntax, flags, and description are available via `gnomon --help` and `gnomon <command> --help` — this README stays introductory rather than duplicating that reference.

### Specifications and approval, practically

A Specification (`.gnomon/specifications/SPEC-NNN-*.md`) is `Draft` or `Approved` — never a third state, and never something the CLI remembers separately from the evidence itself. `gnomon approve SPEC-001` records a fingerprint of the Specification's exact current content; any subsequent edit changes that fingerprint, so the Specification is `Draft` again until re-approved — approval is bound to content, not to a name. `gnomon implement`/`gnomon test` (when targeting a Specification) both require `Approved`; nothing bypasses this by construction. `gnomon revoke SPEC-001` withdraws an active approval explicitly; it is a Human-exclusive action, never triggered automatically by an Agent's result.

Project knowledge is divided by responsibility:
>>>>>>> afb60ff (docs: introduce the CLI in README and add process exit-code release-gate tests)

Your choice is saved as the default. You can override it per-command with `--agent claude` / `--agent codex`, or change the default anytime with `gnomon agent set-default <claude|codex>`. If nothing is configured and the command isn't running in an interactive terminal (a script, CI), it refuses with a clear message instead of guessing.

---

## Walking through it: adding subscription cancellation

Say you're adding cancellation to an existing SaaS app. Here's what actually happens, command by command, with real output.

### 1. Start the Specification

```
$ gnomon spec create "Subscription Cancellation"
✓ Created Draft Specification — SPEC-001

→ Define its content, then `gnomon approve SPEC-001` once you judge it ready.
```

This is entirely deterministic — no Agent runs. Gnomon assigns the next identity, copies the Specification template into `.gnomon/specifications/SPEC-001-subscription-cancellation.md`, and stops. It's an empty template with section headers (Use Case, Scope, Business Rules, Acceptance Criteria, Dependencies, …) — nothing has been written yet, and Gnomon doesn't pretend otherwise.

**First thing worth noticing:** this file exists, but it isn't real requirements yet. Watch what happens if you try to jump straight to building it anyway:

```
$ gnomon implement SPEC-001
! Implementation blocked — SPEC-001

Unresolved
  SPEC-001 is not Approved (currently Draft)

→ Define SPEC-001 further, or `gnomon approve SPEC-001` if you judge it ready.
```

This is the first real difference from "just pointing an agent at the repo." No Agent was started for that call — none. The CLI checked the Specification's lifecycle state against the filesystem and refused before spending a single token. If you'd been running an AI coding agent directly with no Gnomon in the way, there is nothing stopping it from happily implementing cancellation against a spec that's still an empty template — it would just do its best, and its best means guessing.

### 2. Define it — and watch it ask instead of guess

```
$ gnomon spec define SPEC-001
```

This launches your Agent with the Specification Definition workflow's instructions loaded. It reads the current (empty) content, reads your domain/architecture knowledge, and starts filling in what it can determine on its own — the actor, the general shape of the flow. But it hits real, material questions no amount of code-reading answers: *does access end immediately, or at the end of the current billing period? does the customer get a refund for unused time?* These are exactly the kind of business decisions the workflow's own rules forbid it from inventing — so instead of picking a plausible answer and moving on, it asks you, right there in the same terminal session, and writes your answer into the Specification.

That's the second real difference. It isn't that the agent is incapable of guessing — it's that "guess" isn't an available move for a material business decision under this workflow. If a gap belongs somewhere else entirely — say, how this interacts with your existing billing-provider integration in `ARCHITECTURE.md` — it doesn't quietly decide that either; it stops that part of the work, tells you what decision is needed and who owns it, and once you answer, a separate step (Knowledge Resolution) applies that decision to the right document rather than burying it inside this one Specification.

Once every material gap has an answer, the Agent reports back and the workflow ends. The Specification file now has real content — not because you wrote it, and not because the Agent invented it, but because you answered exactly the questions that couldn't be skipped.

### 3. Approve it — explicitly, by you

```
$ gnomon approve SPEC-001
✓ SPEC-001 is now Approved

→ gnomon implement SPEC-001
```

This is the one step in the entire system that's exclusively yours. No workflow result, no Agent output, no "looks good to me" from the model can do this — `approve` is a separate CLI command a Human runs on purpose. Under the hood it records a fingerprint of the Specification's exact current content. That matters more than it sounds like it should:

```
$ gnomon status
✓ Initialized, contract v1, 1 Specification(s)

Specifications
  SPEC-001: Approved

Working Tree
  uncommitted changes present
```

Now edit the Specification — fix a typo, clarify a sentence, anything:

```
$ gnomon status
✓ Initialized, contract v1, 1 Specification(s)

Specifications
  SPEC-001: Draft

Working Tree
  uncommitted changes present
```

Back to Draft. Automatically. Nobody had to remember to revoke it — the approval was bound to *that exact content*, and the content changed. There's no stale "approved" checkbox anywhere to forget about.

### 4. Build it — now that it's actually eligible

```
$ gnomon approve SPEC-001
$ gnomon implement SPEC-001
```

*(The second command runs the same way `spec define` did: your Agent is launched with the Implementation workflow's instructions and the approved Specification as its scope. It writes the code, then reports back.)*

Here's the part that doesn't show up when you just watch the terminal: Gnomon doesn't accept "I'm done" as the result. The Implementation workflow declares, in its own file, exactly what a finished result has to look like — structured fields, not prose. The CLI parses whatever comes back and checks it against that shape before it will tell you it succeeded. If the Agent's process exits cleanly but never produces a valid result, that is reported as a failure, not a success — a clean exit does not mean the workflow succeeded, and a nonzero one doesn't automatically mean it failed, because Gnomon itself gracefully ends the session the moment it sees a valid result, and that intentional shutdown is never shown to you as a crash.

### 5. Test and verify — and notice the difference between them

```
$ gnomon test SPEC-001
$ gnomon verify
```

Testing designs and runs the actual tests. Verification is a distinct workflow, deliberately: it doesn't get to say "everything looks fine." Its Result Contract requires a `PASS`, `FAIL`, or `UNVERIFIABLE` **per obligation**, plus an aggregate the CLI itself computes from those — not from anything the Agent narrates. If even one obligation comes back `FAIL` or `UNVERIFIABLE`, the aggregate is `FAIL`/blocked, full stop; there's no vague "mostly working" state to hide in.

### 6. Review — only if this actually calls for it

```
$ gnomon review
```

Review is separate from Verification on purpose: Verification asks "is this objectively true," Review asks "is this actually a good idea" — risk, quality, things Verification's checklist can't see. Not every change needs it. A subscription-cancellation flow touching billing and access control probably does; a config typo fix doesn't. Gnomon doesn't force it into the sequence — you run it when the stakes call for it, and its findings land in a different vocabulary (`DEFECT`, `RISK`, `KNOWLEDGE GAP`) from Verification's `FAIL`, because they're different kinds of claims.

### 7. Check in without re-reading the whole conversation

Close your terminal. Come back tomorrow, or hand this to a teammate. Nobody needs the original chat:

```
$ gnomon next
✓ Valid actions for 1 Specification(s), by identity — no priority implied

SPEC-001
  Approved. Valid: `gnomon spec define SPEC-001`, `gnomon implement SPEC-001`, `gnomon test SPEC-001`, `gnomon revoke SPEC-001`.

Working Tree
  Uncommitted changes are present — `gnomon finalize` may be worth considering when you judge the work ready.
```

`next` never claims to remember what you did last session — it doesn't have that information, on purpose (more on that below). It's telling you what's *currently* true and *currently* valid, computed fresh from the filesystem, the same answer no matter who asks or which Agent they're using.

### 8. Hand it off to Git — as its own authorized step

```
$ gnomon finalize
```

This is also its own workflow, not an automatic last step of Implementation. And it has one hard rule worth knowing: any part of the change still governed by a Draft Specification is categorically excluded from what gets finalized — not "discouraged," excluded, with no override. "Implement, then quietly commit and push" is not a thing that can happen by accident here.

---

## What just happened, underneath

If you followed the walkthrough, you already felt the important parts. Here's the vocabulary for them.

**Specifications have exactly two states: Draft and Approved.** Never a third, never stored as a flag you could forget to flip. Approved is *derived*, every time, from a fingerprint of the Specification's current content plus an approval record — edit the content, the fingerprint changes, the derivation changes. Nothing is cached.

**Approval is the one Human-exclusive act in the system.** `gnomon approve` / `gnomon revoke` are the only things that change it, and no Agent result can trigger either. It's the boundary between "behavior being defined" and "behavior authorized to drive consequential work" — Implementation, Testing, and Git Finalization all require it.

**Workflows are Agent-driven but their results are machine-checked.** Every built-in workflow (`workflows/*.md`) declares, in its own frontmatter, exactly what a finished result must contain and how each possible outcome classifies (success or blocked). Gnomon validates the Agent's structured output against that before deciding what to tell you — it's not parsing English and hoping.

**A material knowledge gap gets asked, not guessed.** Specification Definition is explicit about this: a gap is material only if leaving it unresolved would force later work to invent something, and materiality is judged against the Specification's own scope — not "did every template field get filled in." A gap owned by the Specification itself gets resolved right there, in conversation. A gap owned elsewhere (architecture, domain rules, conventions) gets routed through a separate step, Knowledge Resolution, so the decision lands in the document that actually owns it instead of getting buried.

**Verification and Review ask different questions.** Verification: is this objectively true, obligation by obligation (`PASS`/`FAIL`/`UNVERIFIABLE`). Review: given what's known, is this actually sound engineering (`PASS`/`DEFECT`/`RISK`/`KNOWLEDGE GAP`). Conflating them is exactly how "looks fine to me" quietly replaces an actual check.

**The CLI decides what's *allowed*; it never decides what's *right*.** `gnomon status`, `validate`, and `next` are read-only and Agent-free — they compute facts from the filesystem (is this Approved? does the Contract parse? is the working tree clean?) and never render an engineering judgment. Everything about what to actually build remains the Agent's and yours.

---

## Where Gnomon doesn't help — said plainly

This only works if you know where the edges are.

* **It doesn't make the Agent more correct.** Verification still runs through the same kind of Agent that wrote the code. Gnomon forces the *claim* into a checkable shape (`PASS`/`FAIL` per obligation instead of a paragraph) — that makes it easier for you to spot-check, but it can't detect a confidently wrong answer on its own.
* **It is not unattended automation.** Every Agent-invoking command hands the Agent a real interactive terminal session — you're present for it, including its own permission prompts, the same as running an AI coding agent directly. There's no headless "kick it off and walk away" mode.
* **It remembers nothing about what ran before.** `status` and `next` never claim a workflow "already completed" or "already passed" — that would require persisted execution history, and none exists, on purpose (it's how a stale record can never lie to you). If you need to know whether verification passed last Tuesday, that's still in your terminal scrollback, same as without Gnomon.
* **Approval is a discipline, not a lock.** Enforcement is the CLI refusing to proceed — there's no server, no permissions system. Someone determined to bypass it could hand-edit the evidence files directly. It stops accidents, not a determined adversary.
* **It adds real ceremony.** Draft → Definition → Approval → Implementation → Verification is genuine overhead for a one-line fix, and nothing tells you when to skip it — you just edit the file yourself when it's not worth the process; Gnomon only ever touches its own `.gnomon/` directory, never the rest of your repo.
* **Dependencies between Specifications are tracked, not judged.** Gnomon can tell an Agent whether a declared dependency exists and what its lifecycle is. Whether a missing or Draft dependency actually blocks the task at hand is left entirely to that workflow's own judgment, same as any other engineering call.

---

## Human / Agent / CLI — who decides what

|                    | Decides                                                                                                                                                    |
| ------------------ | ---------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Human**          | What to build, whether a Specification is ready to authorize, engineering judgment calls Review surfaces, whether to publish Git work                      |
| **Agent**          | How to build it, what a Specification's content should say (subject to asking about material gaps), what the Verification/Review evidence actually shows   |
| **CLI** (`gnomon`) | Whether a request is currently *allowed* to start, whether a returned result matches its declared shape — never whether the underlying decision was *good* |

---

## Command reference

Full syntax, flags, and per-command help are always available from the binary itself — that's the authoritative reference, not this table:

```
gnomon --help
gnomon <command> --help
```

| Group                 | Commands                                                                                                 |         |
| --------------------- | -------------------------------------------------------------------------------------------------------- | ------- |
| Project               | `init`, `describe`, `bootstrap`                                                                          |         |
| Specification         | `spec discover`, `spec create <title>`, `spec define <SPEC-id>`, `approve <SPEC-id>`, `revoke <SPEC-id>` |         |
| Engineering           | `implement <SPEC-id>`, `test [SPEC-id]`, `verify [target]`, `review [target]`, `finalize`                |         |
| Guidance / Inspection | `status`, `validate`, `next`                                                                             |         |
| Advanced              | `run <workflow-identity> [target]`, `agent`, `agent set-default <claude                                  | codex>` |

A few worth knowing about before you need them:

* **`describe` / `bootstrap`** matter for a brand-new project with no established `PROJECT.md`/`ARCHITECTURE.md`/etc. yet — `describe` captures that initial knowledge, `bootstrap` establishes the verified baseline from it. An existing project (like the walkthrough above) may never need either.
* **`validate`** is a different question from `status`: not "what state is my project in" but "is the Gnomon structure itself intact" — Contract well-formedness, evidence file shape, duplicate identities. Clean pass/fail, meant for CI.
* **`run <identity> [target]`** is the generic escape hatch — it invokes any workflow by its own declared identity, including a project's own custom workflows, through the exact same engine the dedicated commands use. Ordinary work should use the dedicated command; `run` exists for the cases that don't have one.

---

## Repository layout

```text
.
├── context/
│   ├── PROJECT.md, DOMAIN.md, ARCHITECTURE.md, STACK.md, CONVENTIONS.md
│   └── PRODUCT_EXPERIENCE.md, UI_FOUNDATION.md, INTERACTION_PATTERNS.md   (Design Knowledge)
├── specifications/
│   ├── SPEC-000-use-case-name.md            (template)
│   ├── SPECIFICATION_LIFECYCLE.md
│   └── SPECIFICATION_DEPENDENCIES.md
├── contracts/            # reusable behavioral invariants per artifact category
├── decisions/
│   └── ADR-000-decision-title.md            (template)
├── workflows/             # the 10 built-in workflows Gnomon drives an Agent through
├── evaluations/
│   ├── VERIFICATION_CRITERIA.md
│   ├── ENGINEERING_REVIEW_QUESTIONS.md
│   └── SPECIFICATION_READINESS_CRITERIA.md
<<<<<<< HEAD
├── cmd/gnomon/            # the CLI's entry point (Cobra wiring only)
├── internal/              # the CLI's implementation
├── cli/                   # the CLI's own design documents
=======
├── cmd/gnomon/          # the gnomon CLI's entry point (Cobra command wiring only)
├── internal/            # the CLI's implementation — see cli/IMPLEMENTATION_ARCHITECTURE.md
├── cli/                 # the CLI's own design documents (Steps 1-13)
>>>>>>> afb60ff (docs: introduce the CLI in README and add process exit-code release-gate tests)
├── go.mod, go.sum
├── LICENSE
└── README.md
```

Everything under `context/`, plus `SPEC-000-use-case-name.md` and `ADR-000-decision-title.md`, are intentionally unfilled templates — `gnomon spec create`/Discovery copy the Specification template for you; the ADR template you copy by hand. Bracketed text like `[Project name]` marks what to replace. `contracts/`, `workflows/`, and `evaluations/` are reusable baselines you adapt, not templates you fill in from scratch.

Each document owns one kind of knowledge, and Gnomon expects you not to duplicate a rule across them: feature behavior belongs in a Specification, cross-feature business rules in `DOMAIN.md`, structural constraints in `ARCHITECTURE.md`, artifact-specific invariants in a contract.

### The ten workflows

| Workflow                                                                             | Produces                                                                                              |
| ------------------------------------------------------------------------------------ | ----------------------------------------------------------------------------------------------------- |
| [`initial-knowledge-establishment.md`](workflows/initial-knowledge-establishment.md) | `READY_FOR_BOOTSTRAP` or `BLOCKED` — minimum approved project knowledge for a new project             |
| [`bootstrap.md`](workflows/bootstrap.md)                                             | `BOOTSTRAP_COMPLETE` or `BLOCKED` — the verified project baseline                                     |
| [`specification-discovery.md`](workflows/specification-discovery.md)                 | `CANDIDATE_PROPOSED`, `NO_CANDIDATE_IDENTIFIED`, or `BLOCKED` — proposes the next Draft Specification |
| [`specification-definition.md`](workflows/specification-definition.md)               | `READY_FOR_APPROVAL` or `BLOCKED` — brings a Draft to approval-ready                                  |
| [`implementation.md`](workflows/implementation.md)                                   | `IMPLEMENTATION_COMPLETE` or `BLOCKED` — requires an `Approved` Specification                         |
| [`testing.md`](workflows/testing.md)                                                 | `TESTING_COMPLETE` or `BLOCKED` — requires an `Approved` Specification, when one is given             |
| [`verification.md`](workflows/verification.md)                                       | Per-obligation `PASS`/`FAIL`/`UNVERIFIABLE`, plus a computed aggregate                                |
| [`review.md`](workflows/review.md)                                                   | `PASS`/`DEFECT`/`RISK`/`KNOWLEDGE GAP` findings, plus a computed aggregate                            |
| [`knowledge-resolution.md`](workflows/knowledge-resolution.md)                       | `RESOLVED` or `BLOCKED` — applies a Human decision to its owning document                             |
| [`git-finalization.md`](workflows/git-finalization.md)                               | `COMMIT_PREPARED`, `PUBLISHED`, or `BLOCKED` — excludes Draft-governed work entirely                  |

---

## Writing prompts for Gnomon

The workflow you invoke already defines which knowledge to load and which steps to follow — your prompt doesn't need to re-derive any of that. What it should carry: **intent** (implement, verify, review, fix, finalize), **scope** (which Specification or area), **the outcome you want**, **real constraints** ("must not touch X"), and **explicit authorization** for anything destructive or for a Git push.

What it normally shouldn't try to do: reproduce the workflow's own steps, restate conventions or contracts already owned elsewhere, pre-fill a Specification before Definition has asked what's actually needed, or invent a requirement that isn't approved anywhere — a real gap should be reported, not guessed around, same as inside the workflows themselves.

| Situation                                         | Example prompt                                                                                                                                             |
| ------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Discovering what to specify next                  | "What's the next Specification worth creating for this project? Propose one and tell me why."                                                              |
| Creating a Specification you already have in mind | "Create a new Specification titled 'Bulk Invite'. Don't define it yet."                                                                                    |
| Bringing a Draft to readiness                     | "Define SPEC-014 (bulk invite) so it's ready for approval. Ask me about anything material it's still missing."                                             |
| Implementing an approved feature                  | "Implement SPEC-014 (bulk invite). Stay within its acceptance criteria and flag anything it leaves undefined."                                             |
| Fixing a bug                                      | "Users report a duplicate charge when retrying a failed checkout. Reproduce and fix it within the checkout use case; don't touch unrelated payment logic." |
| Verifying an implementation                       | "Run verification on the checkout implementation against SPEC-009. Report PASS/FAIL/UNVERIFIABLE per obligation."                                          |
| Reviewing engineering quality                     | "Review the checkout implementation for risk and completeness. Verification evidence already exists — reuse it, don't repeat it."                          |
| Incomplete or ambiguous knowledge                 | "SPEC-009 doesn't define behavior for a canceled-then-retried order. Don't guess — report what's missing and who owns the decision."                       |
| Resolving a reported gap                          | "For the canceled-then-retried gap in SPEC-009: retried orders must reuse the original idempotency key."                                                   |
| Finalizing completed work                         | "The checkout fix is verified and reviewed. Commit it on a new branch. Do not push."                                                                       |

---

## Extending Gnomon

Add a document only when it has a distinct, demonstrated responsibility — not preemptively:

* A Specification for a new bounded use case.
* An ADR for a significant decision.
* A contract when an artifact category has a reusable behavioral invariant not owned elsewhere.
* A new or revised workflow when a recurring engineering intent needs a stable procedure.
* Evaluation criteria when a reusable verification obligation or review question is missing.

<<<<<<< HEAD
Keep new documents focused and link them to their authoritative sources, rather than turning a preference into a rule everyone has to follow.
=======
### Evaluation assets

- [`VERIFICATION_CRITERIA.md`](evaluations/VERIFICATION_CRITERIA.md) — reusable checks spanning specifications, domain, architecture, contracts, stack, conventions, documentation, tests, static analysis, and formatting.
- [`ENGINEERING_REVIEW_QUESTIONS.md`](evaluations/ENGINEERING_REVIEW_QUESTIONS.md) — judgment-based questions covering authorization, authentication, data integrity, state, failures, security, asynchronous or AI work, maintainability, and completeness.
- [`SPECIFICATION_READINESS_CRITERIA.md`](evaluations/SPECIFICATION_READINESS_CRITERIA.md) — reusable criteria used by Specification Definition to determine `READY_FOR_APPROVAL`, based on behavioral sufficiency rather than template completeness.

## Getting Started

The steps below describe the underlying process in terms of the workflow documents directly — useful for understanding what actually happens, or for following it by hand. Most users should instead run this same sequence through the [Gnomon CLI](#gnomon-cli) (`gnomon init`, `gnomon describe`, `gnomon bootstrap`, …), which deterministically gates each step and launches the Agent for you.

1. Copy this repository, or copy its document directories into the target repository.
2. Run [`workflows/initial-knowledge-establishment.md`](workflows/initial-knowledge-establishment.md) to turn the initial intent and anything already in the repository into the minimum approved `PROJECT.md`, `DOMAIN.md` (only when material), `ARCHITECTURE.md`, `STACK.md`, and `CONVENTIONS.md` knowledge Bootstrap needs. Ask the human only about unknowns material to that baseline, and stop once its accuracy is confirmed rather than completing every template section.
3. Run [`workflows/bootstrap.md`](workflows/bootstrap.md) to establish the minimum verified project baseline from that approved knowledge.
4. Create each Specification either by running [`workflows/specification-discovery.md`](workflows/specification-discovery.md) to propose one candidate at a time, or directly by title when you already know what you want — both use its Specification Identity rule and produce a `Draft`. Keep it as-is until it's needed.
5. Run [`workflows/specification-definition.md`](workflows/specification-definition.md) to bring a Draft Specification to `READY_FOR_APPROVAL`, then have a human approve it per [`SPECIFICATION_LIFECYCLE.md`](specifications/SPECIFICATION_LIFECYCLE.md) before relying on it.
6. Copy and rename the ADR template whenever a significant decision needs a durable record.
7. Select the workflow that matches the user's intent. Follow its required-knowledge section rather than loading every document automatically.
8. Apply only the contracts relevant to the artifacts being created or changed.
9. Use verification criteria to produce objective evidence, then use review questions when engineering judgment is required.
10. Use Git finalization only after the change is complete, and treat pushing or opening a pull request as a separate authorization gate.

There is nothing to install or execute in this repository. Its Markdown files can be used directly by people, coding agents, or repository-level agent instructions.

## Usage Example

For an approved feature request:

1. Create a specification from `specifications/SPEC-001-use-case-name.md`.
2. Select `workflows/implementation.md`.
3. Load the relevant context, specification, ADRs, and artifact contracts named by the workflow and the change.
4. Inspect the existing code and patterns before designing the smallest coherent implementation.
5. Implement and gather reproducible verification evidence.
6. Run `workflows/verification.md` for objective compliance.
7. Run `workflows/review.md` if risks, quality, or incomplete knowledge require engineering evaluation.
8. Resolve findings through their identified owner and repeat the applicable workflow.
9. When authorized, use `workflows/git-finalization.md` to prepare the Git handoff.

## Prompt Guide

Gnomon's workflows already define which knowledge to load, which steps to follow, and how to handle missing information. A prompt should express engineering intent, scope, and constraints — not restate that procedure.

A good prompt normally includes:

- **Intent** — what kind of change this is (implement, fix, verify, review, finalize, ...).
- **Target/scope** — the specification, use case, artifact, or area affected.
- **Requested change or outcome** — what should be true when the work is done.
- **Known constraints** — anything that limits the change (must not touch X, must preserve Y).
- **Explicit authorization** — required for destructive actions and for any Git push or pull request.

A prompt normally does **not** need to:

- Manually orchestrate which internal knowledge the agent loads — the selected workflow's Inputs section already determines that. Referencing the Specification, ADR, use case, or artifact that defines the task's scope (for example, "Implement SPEC-014") is expected, not something to avoid.
- Reproduce the workflow's steps — `workflows/` already owns the procedure.
- Repeat contract or convention rules — `contracts/`, `context/CONVENTIONS.md`, and the evaluation criteria already own them.
- Invent missing requirements — an approved [`Specification`](specifications/SPEC-001-use-case-name.md) or [`ADR`](decisions/ADR-001-decision-title.md) does, and a gap should be reported rather than guessed.
- Name which workflow resolves a reported knowledge gap — once a gap has been reported, stating the decision is enough; the agent recognizes the context, applies [`workflows/knowledge-resolution.md`](workflows/knowledge-resolution.md), and resumes the paused work.
- Pre-answer every `context/` template section before starting a new project — [`workflows/initial-knowledge-establishment.md`](workflows/initial-knowledge-establishment.md) identifies only what is material to reach Bootstrap and asks about that.
- Pre-fill every section of a Specification before asking for it to be defined — [`workflows/specification-definition.md`](workflows/specification-definition.md) identifies only what this use case's behavior actually needs, and reports `READY_FOR_APPROVAL` without requiring template completeness.
- Plan out the project's whole Specification set in advance — [`workflows/specification-discovery.md`](workflows/specification-discovery.md) proposes one candidate at a time from current project knowledge; ask again for the next one once you've decided on the current proposal.

### Examples by situation

| Situation | Example prompt |
| --- | --- |
| Starting a brand-new project | "I want to build a workplace space-reservation system." |
| Discovering what to specify next | "What's the next Specification worth creating for this project? Propose one and tell me why." |
| Creating a Specification you already have in mind | "Create a new Specification titled 'Bulk Invite'. Don't define it yet." |
| Bringing a Draft Specification to readiness | "Define SPEC-014 (bulk invite) so it's ready for approval. Ask me about anything material it's still missing." |
| Implementing an approved feature | "Implement SPEC-014 (bulk invite). Stay within its acceptance criteria and flag anything it leaves undefined." |
| Changing existing behavior | "Update the export use case to also support CSV per SPEC-009's revised acceptance criteria. Preserve the existing JSON export behavior." |
| Fixing a bug | "Users report a duplicate charge when retrying a failed checkout. Reproduce and fix it within the checkout use case; don't touch unrelated payment logic." |
| Verifying an implementation | "Run verification on the checkout implementation against SPEC-009 and the use-case-execution contract. Report PASS/FAIL/UNVERIFIABLE per obligation." |
| Reviewing engineering quality | "Review the checkout implementation for risk and completeness. Verification evidence already exists — reuse it rather than repeating it." |
| Incomplete or ambiguous knowledge | "SPEC-009 doesn't define behavior for a canceled-then-retried order. Don't guess — report what's missing and who owns the decision." |
| Resolving a reported knowledge gap | "For the canceled-then-retried gap in SPEC-009: retried orders must reuse the original idempotency key." |
| Finalizing completed work in Git | "The checkout fix is verified and reviewed. Commit it on a new branch. Do not push." |

## Extending the System

Add a document only when it has a distinct, demonstrated responsibility:

- Add a specification for a new bounded use case.
- Add an ADR for a significant decision.
- Add a contract when an artifact category has reusable behavioral invariants not owned elsewhere.
- Add or revise a workflow when a recurring engineering intent needs a stable execution process.
- Add evaluation criteria when a reusable verification obligation or review question is missing.

Keep new documents focused, link them to their authoritative sources, and avoid turning examples or preferences into project requirements.
>>>>>>> afb60ff (docs: introduce the CLI in README and add process exit-code release-gate tests)

## License

Released under the [MIT License](LICENSE).