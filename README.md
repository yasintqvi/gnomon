<p align="center">
  <img src="assets/gnomon.png" alt="Gnomon logo" width="190" />
</p>

<h1 align="center">Gnomon</h1>
<p align="center"><em>AI Software Engineering System</em></p>

## Why Gnomon?

AI coding agents can read your code, implement features, run tests, and make changes quickly. But access to a repository doesn't tell an Agent everything about how the project is supposed to work.

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
```

*(No package manager, curl installer, or prebuilt binary exists yet — that's not a documentation gap, it genuinely isn't built.)*

In an existing project — any existing project, Gnomon doesn't care what's already there — run:

```
$ gnomon init
✓ Gnomon project initialized

→ gnomon describe
  Establish or reconcile project knowledge before starting work.
```

That materializes a `.gnomon/` directory: a template Specification, the Draft/Approved lifecycle rules, the ten built-in workflows (Implementation, Testing, Verification, Review, and so on), and a version marker. Nothing else in your repository is touched. If you never run another `gnomon` command, this did nothing to your codebase.

`gnomon describe` (Initial Knowledge Establishment) reads whatever intent you give it *together with* whatever's already in the repository — existing code, config, docs — and reconciles the two into approved `PROJECT.md`/`DOMAIN.md`/`ARCHITECTURE.md`/`STACK.md`/`CONVENTIONS.md` content; it does not assume the project is empty, so it's equally useful for a brand-new project and for a first-time adoption of an existing one. `gnomon bootstrap` then establishes the verified baseline from that knowledge. This is recommended, not enforced — nothing downstream requires either to have run, so going straight to `spec create` is a legitimate choice if this project's knowledge is already accurate. The walkthrough below does exactly that, since that's the more common case for a project with established conventions; if you're starting from nothing, run `gnomon describe` then `gnomon bootstrap` first and each will point you to the next step.

## Choosing an Agent

Gnomon drives either **Claude Code** or **Codex** — whichever you already use, or both. The first time a command needs one and none is configured, you'll see an arrow-key prompt:

```
? Choose your default Agent:

  ❯ Claude Code
    Codex

↑/↓ navigate • Enter select
```

Your choice is saved as the default. You can override it per-command with `--agent claude` / `--agent codex`, or change the default anytime with `gnomon agent set-default <claude|codex>`. If nothing is configured and the command isn't running in an interactive terminal (a script, CI), it refuses with a clear message instead of guessing.

---

## The everyday interface: `gnomon spec`

You don't need to memorize a command sequence. In a real interactive terminal, `gnomon spec` opens a browser of every Specification — searchable, with each one's lifecycle visible at a glance:

```
? Specifications (1)
  Filter: (type to filter)

 ❯ + Create new Specification
   + Discover next Specification
   SPEC-001   Password Reset                 Draft

↑/↓ navigate • type to filter • Enter select • Esc cancel
```

Selecting a Specification — or running `gnomon spec SPEC-001` directly — opens its workspace: current state, why any action is unavailable, and only the actions that are actually valid right now as an interactive menu.

```
SPEC-001 — Password Reset
  Lifecycle:   Draft
  Fingerprint: f094efce22b1
  Not currently available: implement (SPEC-001 is not Approved (currently Draft)); test (...); revoke (...)

? SPEC-001 — choose an action

  ❯ Define
    Approve
    Verify
    Review
    Back
```

Choosing Approve runs the same deterministic operation `gnomon approve` always has — it's still the one Human-exclusive act in the system, still requires deliberately selecting it. Nothing here auto-approves or chains one action into the next; after Approve completes, the workspace redraws with freshly re-derived state, offering Implement/Test/Revoke now that they're actually valid.

Discovery works the same way from the browser: it proposes one candidate, shows you its rationale, and creates nothing until you explicitly choose "Yes, create it" — declining leaves nothing behind.

Non-interactive contexts (scripts, CI, a piped command) get a clear refusal instead of hanging waiting for keystrokes that will never come, pointing at `gnomon run` — see [The advanced path](#the-advanced-path-gnomon-run) below. The rest of this section walks through what each contextual action actually does, step by step, since that's the clearest way to explain *why* each one exists; day to day, `gnomon spec` does all of it, interactively, without needing any of the commands named along the way.

---

## Walking through it: adding subscription cancellation

Say you're adding cancellation to an existing SaaS app. Here's what actually happens inside the workspace, step by step, with real output.

### 1. Start the Specification

```
$ gnomon spec create "Subscription Cancellation"
✓ Created Draft Specification — SPEC-001

→ Define its content, then `gnomon approve SPEC-001` once you judge it ready.
```

This is entirely deterministic — no Agent runs. Gnomon assigns the next identity, copies the Specification template into `.gnomon/specifications/SPEC-001-subscription-cancellation.md`, and stops. It's an empty template with section headers (Use Case, Scope, Business Rules, Acceptance Criteria, Dependencies, …) — nothing has been written yet, and Gnomon doesn't pretend otherwise. (Discovery, from the browser, reaches this same point after you accept a proposed candidate — see [Discovery](#the-everyday-interface-gnomon-spec) above.)

**First thing worth noticing:** this file exists, but it isn't real requirements yet. Open its workspace (`gnomon spec SPEC-001`) and Implement is listed, but not selectable:

```
SPEC-001 — Subscription Cancellation
  Lifecycle: Draft
  Not currently available: implement (SPEC-001 is not Approved (currently Draft)); test (...); revoke (...)
```

This is the first real difference from "just pointing an agent at the repo." No Agent is ever started for an unavailable action — the CLI checks the Specification's lifecycle state against the filesystem before anything can run. If you'd been running an AI coding agent directly with no Gnomon in the way, there is nothing stopping it from happily implementing cancellation against a spec that's still an empty template — it would just do its best, and its best means guessing.

### 2. Define it — and watch it ask instead of guess

Selecting **Define** from the workspace menu launches your Agent with the Specification Definition workflow's instructions loaded. It reads the current (empty) content, reads your domain/architecture knowledge, and starts filling in what it can determine on its own — the actor, the general shape of the flow. But it hits real, material questions no amount of code-reading answers: *does access end immediately, or at the end of the current billing period? does the customer get a refund for unused time?* These are exactly the kind of business decisions the workflow's own rules forbid it from inventing — so instead of picking a plausible answer and moving on, it asks you, right there in the same terminal session, and writes your answer into the Specification.

That's the second real difference. It isn't that the agent is incapable of guessing — it's that "guess" isn't an available move for a material business decision under this workflow. If a gap belongs somewhere else entirely — say, how this interacts with your existing billing-provider integration in `ARCHITECTURE.md` — it doesn't quietly decide that either; it stops that part of the work, tells you what decision is needed and who owns it, and once you answer, a separate step (Knowledge Resolution) applies that decision to the right document rather than burying it inside this one Specification.

Once every material gap has an answer, the Agent reports back and the workflow ends. The Specification file now has real content — not because you wrote it, and not because the Agent invented it, but because you answered exactly the questions that couldn't be skipped.

### 3. Approve it — explicitly, by you

Selecting **Approve** is the one step in the entire system that's exclusively yours. No workflow result, no Agent output, no "looks good to me" from the model can do this. Under the hood it records a fingerprint of the Specification's exact current content. That matters more than it sounds like it should:

```
$ gnomon status
✓ Initialized, contract v1, 1 Specification(s)

Specifications
  SPEC-001: Approved
```

Now edit the Specification — fix a typo, clarify a sentence, anything:

```
$ gnomon status
✓ Initialized, contract v1, 1 Specification(s)

Specifications
  SPEC-001: Draft
```

Back to Draft. Automatically. Nobody had to remember to revoke it — the approval was bound to *that exact content*, and the content changed. There's no stale "approved" checkbox anywhere to forget about. Its previous approved text, who approved it, and when, remain visible in the workspace's revision list rather than disappearing.

### 4. Build it — now that it's actually eligible

Re-approve, and **Implement** appears in the menu, now that the Specification is Approved again. Selecting it launches your Agent with the Implementation workflow's instructions and the approved Specification as its scope; it writes the code, then reports back.

Here's the part that doesn't show up when you just watch the terminal: Gnomon doesn't accept "I'm done" as the result. The Implementation workflow declares, in its own file, exactly what a finished result has to look like — structured fields, not prose. The CLI parses whatever comes back and checks it against that shape before it will tell you it succeeded. If the Agent's process exits cleanly but never produces a valid result, that is reported as a failure, not a success — a clean exit does not mean the workflow succeeded, and a nonzero one doesn't automatically mean it failed, because Gnomon itself gracefully ends the session the moment it sees a valid result, and that intentional shutdown is never shown to you as a crash.

### 5. Test it

Selecting **Test** designs and runs the actual tests against the same Approved Specification — also listed in the workspace only once Implementation's precondition (Approved) is met.

### 6. Verify and review — deliberately not Specification-scoped

Verification and Review are not contextual actions inside a Specification's workspace, on purpose: their own Contracts declare `specification_reference: none` — Core's own statement that their target isn't necessarily a Specification at all, and nothing ties either one to *this* Specification's lifecycle the way Implementation and Testing are tied to it. Run them against whatever the actual change touches, through [the advanced path](#the-advanced-path-gnomon-run):

```
gnomon run verification src/billing/
gnomon run review src/billing/
```

Verification's Result Contract requires a `PASS`, `FAIL`, or `UNVERIFIABLE` **per obligation**, plus an aggregate the CLI itself computes from those — not from anything the Agent narrates; if even one obligation comes back `FAIL` or `UNVERIFIABLE`, the aggregate is blocked, full stop. Review asks a different question — "is this actually a good idea," risk and quality Verification's checklist can't see — and isn't needed for every change; a subscription-cancellation flow touching billing and access control probably warrants it, a config typo fix doesn't. Its findings land in a different vocabulary (`DEFECT`, `RISK`, `KNOWLEDGE GAP`) from Verification's `FAIL`, because they're different kinds of claims.

### 7. Check in without re-reading the whole conversation

Close your terminal. Come back tomorrow, or hand this to a teammate. Nobody needs the original chat:

```
$ gnomon next
✓ Valid actions for 1 Specification(s), by identity — no priority implied

SPEC-001
  Approved. Valid: `gnomon run specification-definition SPEC-001`, `gnomon run implementation SPEC-001`, `gnomon run testing SPEC-001`, `gnomon revoke SPEC-001`.

Working Tree
  Uncommitted changes are present — `gnomon run git-finalization` may be worth considering when you judge the work ready.
```

`next` never claims to remember what you did last session — it doesn't have that information, on purpose (more on that below). It's telling you what's *currently* true and *currently* valid, computed fresh from the filesystem, the same answer no matter who asks or which Agent they're using. (It names the `run` form here because that's what's actually runnable outside the workspace; inside `gnomon spec SPEC-001` the same actions appear as plain menu entries.)

### 8. Hand it off to Git — as its own authorized step

```
gnomon run git-finalization
```

Finalization is its own workflow too, not an automatic last step of Implementation, and — like Verify/Review — not Specification-scoped, so it isn't a workspace action either. It has one hard rule worth knowing: any part of the change still governed by a Draft Specification is categorically excluded from what gets finalized — not "discouraged," excluded, with no override. "Implement, then quietly commit and push" is not a thing that can happen by accident here.

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

| Group | Commands |
|---|---|
| Everyday (interactive) | `spec`, `spec <SPEC-id>` |
| Project | `init`, `describe`, `bootstrap` |
| Specification (CLI-native) | `spec create <title>`, `approve <SPEC-id>`, `revoke <SPEC-id>` |
| Guidance / Inspection | `status`, `validate`, `next` |
| Advanced | `run <workflow-identity> [target]`, `agent`, `agent set-default <claude/codex>` |

A few worth knowing about before you need them:

* **`spec` / `spec <SPEC-id>`** are the two everyday commands — see "The everyday interface" above. There is no separate command for Define, Implement, Test, Verify, or Review to memorize; each is a contextual action from the workspace, or, outside it, a `gnomon run` invocation — see below.
* **`spec create`, `approve`, `revoke`** are the three Specification-related operations that keep their own dedicated command even outside the workspace: none of them is a Workflow Contract `gnomon run` could invoke — Draft Creation and Approval/Revocation are deterministic, CLI-native operations Core itself carves out from the Agent workflow model entirely.
* **`describe` / `bootstrap`** are the recommended first step after `init` — for a brand-new project and for first-time adoption of an existing one alike, since `describe`'s own first step reads whatever's already in the repository rather than assuming it's empty. Recommended, never enforced: if your project's knowledge is already accurate, going straight to Specification work is equally legitimate.
* **`validate`** is a different question from `status`: not "what state is my project in" but "is the Gnomon structure itself intact" — Contract well-formedness, evidence file shape, duplicate identities. Clean pass/fail, meant for CI.

### The advanced path: `gnomon run`

Every Contract-driven Agent workflow — Implementation, Testing, Verification, Review, Specification Definition, Specification Discovery, Git Finalization — is reachable non-interactively by its own declared identity, for scripts, CI, or a manual one-off:

```
gnomon run implementation SPEC-014
gnomon run testing SPEC-014
gnomon run verification src/auth/
gnomon run review src/auth/
gnomon run specification-definition SPEC-014
gnomon run specification-discovery
gnomon run git-finalization
```

This is the exact same eligibility and execution engine the workspace's contextual actions use — nothing behaves differently between them. It's also how a project's own custom, Contract-driven workflows are invoked, the same way as any built-in one. Ordinary interactive work should use `gnomon spec` / `gnomon spec <SPEC-id>` instead; `run` exists for where an interactive browser can't run, not as a second way to do the same thing day to day.

### Specification revisions

Every time a Specification is approved, Gnomon preserves the exact content that was approved — not just the fingerprint. Editing an Approved Specification makes it Draft again (as always), but the previous approved text, who approved it, and when, all remain inspectable in its workspace rather than disappearing. This is approval history, not execution history: Gnomon still never remembers which workflow last ran or whether a past Verification passed — only durable, Human-authorized evidence is kept.

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
├── cmd/gnomon/            # the CLI's entry point (Cobra wiring only)
├── internal/              # the CLI's implementation
├── cli/                   # the CLI's own design documents
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

Keep new documents focused and link them to their authoritative sources, rather than turning a preference into a rule everyone has to follow.

## License

Released under the [MIT License](LICENSE).
