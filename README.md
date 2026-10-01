<p align="center">
  <img src="assets/gnomon-logo.png" alt="Gnomon logo" width="190" />
</p>

<h1 align="center">Gnomon</h1>
<p align="center"><em>AI Software Engineering System</em></p>

AI coding agents are good at working with code. The harder part is keeping them aligned with the engineering intent behind it: what a feature should actually do, what the project already knows, which decisions are still open, and what "done" means for a particular change.

Gnomon gives that work a durable structure inside your repository. You describe what you want, Gnomon helps the agent turn it into a concrete Specification, unresolved decisions are surfaced instead of silently assumed, and implementation proceeds against what you actually approved.

The result is a workflow that can start with a simple request and carry its engineering intent through implementation, testing, and evaluation — without depending on one conversation or one agent session.

## A small example

Suppose you're building a task manager and want to let users **mark a task complete**.

You start with the intent:

> Let users mark a task complete.

That sounds clear enough to implement, but there's already a product decision hiding inside it:

```text
Agent

The requested behavior leaves one decision unresolved:

Can a completed task be reopened, or is completion final?
```

You decide:

> A completed task can be reopened.

Now the behavior is concrete enough to record as a **Specification**:

> **Mark a task complete**
>
> - An open task can be marked complete.
> - A completed task can be reopened.
> - The task keeps its title and description in either state.

That Specification becomes the shared reference for the change. You review it, approve it when it reflects what you want, and the agent implements against it.

The rest of this README follows the same feature.

## Install

Download the latest build for macOS, Linux, or Windows from [**Releases**](https://github.com/yasintqvi/gnomon/releases), extract it, and add `gnomon` (`gnomon.exe` on Windows) to your `PATH`.

Then verify the installation:

```sh
gnomon --version
```

Later, `gnomon update` downloads and installs the latest stable release in place, without needing to repeat the steps above.

Gnomon works with an external AI coding agent. Use `gnomon agent` to inspect or configure the Agent provider Gnomon should use.

## Quick start

Run Gnomon from the root of your project.

### 1. Start with the project

Initialize Gnomon:

```text
$ gnomon init

✓ Gnomon project initialized

→ gnomon spec
  Create a Specification for the change you want, then Define it.
```

This creates a small `.gnomon/` directory: `specifications/`, `approvals/`, and `CONTRACT_VERSION`. Nothing else is copied in. Workflows and Specification templates come from the Gnomon binary, so they stay current when you run `gnomon update`.

The normal path starts right away with a Specification. Two optional steps exist for when they help:

- `gnomon describe` records the few project facts an agent can't learn from the code — purpose, users, hard constraints, project-wide decisions — in a short `.gnomon/context/PROJECT.md`.
- `gnomon bootstrap` makes sure the project builds and its tests run, so later work can be checked.

`gnomon init --full` additionally creates the longer knowledge, design, contract, and ADR templates earlier versions created by default. Templates you haven't filled in are never handed to the agent as project knowledge.

### 2. Start a feature

Open the Specification workspace:

```text
$ gnomon spec
```

You'll get an interactive browser for the project's Specifications:

```text
Specifications

  › + Create Specification
    ✦ Discover next Specification

    No Specifications yet.

↑/↓ navigate • Enter select • Esc back
```

Choose **Create Specification** (or run `gnomon spec create "<title>"`) and enter:

```text
Mark a task complete
```

Gnomon creates the Specification as a Draft and opens its workspace:

```text
SPEC-001 — Mark a task complete

  Lifecycle: Draft

? SPEC-001 — choose an action

  ❯ Define
    Approve
    Back
```

The new Specification uses a short template:

```markdown
# SPEC-001 — Mark a task complete

## Goal
## Decisions          (each question with the answer you gave)
## Acceptance Criteria
## Out of Scope
```

`gnomon spec create --detailed` uses the full use-case template (actors, flows, business rules, inputs and outputs, dependencies) when a change needs it.

At this point you have an identity for the feature, but not an approved requirement.

Choose **Define**.

Describe what you want in normal terms:

> Let users mark a task complete.

The agent works from both your request and the knowledge already recorded for the project. When something material is missing, it should surface the decision rather than quietly choose one.

For our example:

```text
Agent

The requested behavior leaves one decision unresolved:

Can a completed task be reopened, or is completion final?
```

You answer:

> A completed task can be reopened. Its title and description should remain unchanged.

The agent can now finish defining the Specification around that decision.

Instead of relying on the conversation as the lasting source of truth, the result is recorded in the project:

```text
SPEC-001 — Mark a task complete

- An open task can be marked complete.
- A completed task can be reopened.
- The task keeps its title and description in either state.
```

### 3. Approve what should be built

Review the Specification.

If it reflects what you actually want:

```text
$ gnomon approve SPEC-001

✓ SPEC-001 is now Approved
```

The workspace now reflects that state:

```text
SPEC-001 — Mark a task complete

  Lifecycle: Approved

? SPEC-001 — choose an action

  ❯ Define
    Implement
    Test
    Revoke
    Back
```

Choose **Implement** and the agent works against the approved Specification while following the knowledge and conventions already established for the project.

Approval doesn't automatically start implementation, and defining a Specification doesn't automatically approve it. Each transition remains an explicit choice.

A Specification that still contains only the unfilled template can't be approved — there would be nothing to approve. One that is partly filled in (some placeholders left) warns you and, in an interactive terminal, asks you to confirm; that remains your call.

### 4. Keep working from the Specification

You don't need to remember where a previous session stopped.

Run:

```text
$ gnomon spec
```

and reopen the Specification:

```text
Specifications

    SPEC-001  Mark a task complete  Approved
```

The workspace derives what actions are currently valid from the project itself.

For a broader view:

```text
$ gnomon status

✓ Initialized, contract v1, 1 Specification(s)

Specifications
  SPEC-001: Approved

Working Tree
  uncommitted changes present
```

And when you want to see what the project currently allows you to do next:

```sh
gnomon next
```

Gnomon derives that guidance from the current Specifications, approvals, evidence, and repository state rather than relying on hidden workflow history.

## When requirements change

Specifications aren't frozen descriptions of features.

Suppose the requirement changes later:

> Completing a task should now be permanent. A completed task cannot be reopened.

That is still the same feature, but it is no longer the same requirement you approved.

Gnomon recognizes that the approved content has changed:

```text
$ gnomon status

Specifications
  SPEC-001: Draft — changed since approval on 2026-09-24 (1 line(s) changed)
```

The revised Specification needs your approval again before work that requires an approved Specification can continue.

The earlier approved revision remains part of the record, so the project retains both what was previously agreed and what changed.

This makes approval about **the actual requirements**, not merely the name of a feature.

## When the answer isn't known yet

Sometimes the right next feature isn't obvious.

From the Specification browser:

```text
Specifications

  › + Create Specification
    ✦ Discover next Specification
```

**Discover next Specification** lets the agent examine the knowledge already established for the project and propose a candidate.

The proposal is shown before a Specification is created. You can accept it or walk away without changing the project.

The same principle applies while defining features: when required product knowledge is missing, the goal isn't for the agent to invent a plausible answer. Gnomon makes the gap visible so the decision can be made deliberately and recorded where future work can find it.

## Your Agent

Gnomon doesn't replace an AI coding agent and isn't built around a particular provider.

Use:

```sh
gnomon agent
```

to inspect the available Agent providers and configure the one Gnomon should use.

When an Agent is needed and no default has been configured yet, Gnomon lets you choose one interactively. That choice becomes your default and can be changed later through the same Agent interface.

This keeps the engineering workflow independent from whichever Agent provider is doing the work.

## Beyond implementation

A Specification defines what should be built, but implementation isn't always the end of the engineering process.

Gnomon separates different kinds of evaluation because they answer different questions.

**Testing** exercises the implementation where an additional testing pass is useful.

**Verification** asks whether the available evidence demonstrates that the required behavior was actually satisfied.

**Review** looks at the engineering quality, risks, and consequences of the change.

They aren't a mandatory checklist that every change must pass through in exactly the same order. Use the evaluation appropriate to the work.

For advanced workflow execution, Gnomon exposes its workflows directly through:

```sh
gnomon run <workflow> [target]
```

For example:

```sh
gnomon run verification src/tasks/
gnomon run review src/tasks/
```

The normal Specification workspace stays focused on the actions relevant to the Specification itself; the generic `run` interface is there when you need direct access to a workflow.

Verification reports each acceptance criterion with its evidence and one of three results:

```text
! Verification: 1 passed, 0 failed, 1 unverified

Criteria
  1. passed     An open task can be marked complete
                Source: SPEC-001, criterion 1
                Evidence: tests/TaskTest.php::test_complete passed
  2. unverified A completed task can be reopened
                Source: SPEC-001, criterion 2
                Evidence: No evidence was reported, so this criterion is unverified.

Note
  The Agent gathered and reported this evidence. It shows what was checked and how,
  not independent proof that the product works.
```

A criterion reported as passing without evidence is shown as unverified, and a verification with no criteria never reports success.

When Verification or Review finds something to act on, it doesn't just report it and stop — the same interactive session offers to resolve it. Each finding may come with a recommended workflow, reasoned from what actually caused it rather than mechanically from its classification or result: two findings that fail for the same reason can still warrant different fixes, and two that fail differently can end up recommending the same one. You stay in control of what happens with it:

```text
Verification / Review
  ↓
Findings, each with a recommended workflow when one applies
  ↓
You confirm it, choose a different workflow, or skip it
  ↓
Gnomon checks eligibility and launches that workflow
  ↓
The Agent resolves it — asking you anything it needs, in the same session
  ↓
A fresh Verification / Review reports the current state
```

For example, a review of the finished feature might turn up something worth fixing:

```text
$ gnomon run review src/tasks/

! Defect — src/tasks/

Review found 1 finding

F-001  DEFECT
  Reopening a task doesn't clear its completed_at timestamp

Recommended workflow:
  Implementation

? What would you like to do?

  ❯ Continue with Implementation
    Choose another workflow
    Skip
```

Confirming launches Implementation with this finding as context, the same way it would if you ran it directly — it still requires an Approved Specification. From there, Gnomon's own part is done: the Agent does the actual resolution work, including asking you directly if it needs a decision along the way. Once it finishes, Gnomon reruns Review itself, fresh, rather than taking the fix's own word for it — nothing declares a finding resolved except that new result.

Skipping a finding doesn't accept, waive, or resolve it — it just means nothing is launched for it this time. Findings aren't tracked between runs; each Verification or Review reports the project's current state, not a backlog.

## Workflows and upgrades

Workflows are read from the Gnomon binary. To change one for a project:

```sh
gnomon workflows                          # what's in effect, and where it comes from
gnomon workflows customize review.md      # copy it into .gnomon/workflows/ and edit it there
```

An edited file in `.gnomon/workflows/` replaces the bundled workflow of the same name; a new file there adds a workflow you run with `gnomon run <identity>`.

**Existing projects need no migration.** Workflow files an older `gnomon init` copied into `.gnomon/workflows/` are recognized as unmodified copies and ignored in favor of the current versions (`gnomon validate` lists them; you can delete them). Copies you edited stay in effect. Existing Specifications, approvals, and edited knowledge files keep working; unfilled templates stop being handed to the agent.

## Working with agents outside Gnomon

Product decisions and approved behavior live in `.gnomon/`: each decision in the Specification it governs, project-wide ones in `.gnomon/context/`. Gnomon tells the agent that these override its own memory and other instruction files. An agent's private memory (for example Claude Code's auto memory or Codex memories) is per machine and invisible to other agents and teammates, so record decisions in `.gnomon/`, not there.

If you also work with agents directly, you can add this optional pointer to `AGENTS.md` (Codex reads it; Claude Code reads it when there is no `CLAUDE.md`, or via `@AGENTS.md` from one). Gnomon doesn't add it for you: in a small trial, Claude Code and Codex found and applied a decision in `.gnomon/specifications/` with or without it.

```markdown
<!-- BEGIN:gnomon -->
Product decisions and approved behavior for this project live in `.gnomon/specifications/`
(and project knowledge in `.gnomon/context/`, if present). Read the relevant ones before
designing or changing behavior; they override other instructions and your own memory.
<!-- END:gnomon -->
```

## The model behind it

The normal workflow is intentionally small:

```text
Intent
  ↓
Project Knowledge
  ↓
Specification
  ↓
Human Approval
  ↓
Implementation
  ↓
Testing / Verification / Review when needed
```

The important part is what survives between those steps.

Project knowledge has an explicit home. Feature behavior lives in Specifications. Material decisions aren't silently replaced by Agent assumptions. Approval is tied to the requirements that were actually reviewed. Agent work follows defined workflows, and evaluation has explicit evidence and outcomes.

All of this stays with the repository rather than depending on a particular conversation.

Gnomon doesn't try to make every engineering decision for you. It gives the Agent enough structure to work from what the project already knows — and makes it clear when the project doesn't know enough yet.

## Learn more

The README intentionally covers the normal path rather than every part of the system.

For command syntax, use:

```sh
gnomon --help
gnomon <command> --help
```

For the deeper model:

- **Project setup:** [Initialization](cli/PROJECT_INITIALIZATION.md) · [Bootstrap](workflows/bootstrap.md) · [Specification templates](templates/)
- **Defining features:** [Specification Definition](workflows/specification-definition.md) · [Specification Readiness Criteria](evaluations/SPECIFICATION_READINESS_CRITERIA.md)
- **Specification lifecycle:** [Lifecycle](specifications/SPECIFICATION_LIFECYCLE.md) · [Dependencies](specifications/SPECIFICATION_DEPENDENCIES.md)
- **Evaluating work:** [Testing](workflows/testing.md) · [Verification](workflows/verification.md) · [Review](workflows/review.md)
- **Workflows and customization:** [Core Distribution](cli/CORE_DISTRIBUTION.md) · [Workflow Contracts](cli/WORKFLOW_CONTRACT.md) · [Result Protocol](cli/RESULT_PROTOCOL.md)

[Releases](https://github.com/yasintqvi/gnomon/releases) · [Report an issue](https://github.com/yasintqvi/gnomon/issues)

## License

[MIT](LICENSE)
