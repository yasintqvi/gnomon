<p align="center">
  <img src="assets/gnomon.png" alt="Gnomon logo" width="190" />
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

Gnomon works with an external AI coding agent. Use `gnomon agent` to inspect or configure the Agent provider Gnomon should use.

## Quick start

Run Gnomon from the root of your project.

### 1. Start with the project

Initialize Gnomon:

```text
$ gnomon init

✓ Gnomon project initialized

→ gnomon describe
  Establish or reconcile project knowledge before starting work.
```

Gnomon creates a `.gnomon/` directory inside the repository. This is where project knowledge, Specifications, decisions, workflows, and the other engineering artifacts used by Gnomon live.

Next, let Gnomon establish what the project already knows:

```text
$ gnomon describe
```

The agent inspects the repository and any project knowledge already recorded — source, docs, configuration, whatever's there — and establishes or reconciles knowledge such as the project's purpose, domain, architecture, stack, and conventions from it. If something material genuinely can't be determined that way, it asks you directly, in the same session, rather than guessing; on a brand-new project with little to inspect yet, that's how it learns what you're building.

Gnomon doesn't require you to explain an existing project from scratch every time a new session begins. That knowledge stays with the repository and can evolve with it.

Once the project knowledge is established, Gnomon points you toward bootstrap:

```text
→ gnomon bootstrap
  Establish a verified runnable baseline for the project.
```

Bootstrap gives future work a known engineering baseline. These onboarding steps are recommended rather than permanent gates — once the project is established, normal work revolves around Specifications.

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

Choose **Create Specification** and enter:

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
SPEC-001 — Mark a task complete

  Lifecycle: Draft
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

When Verification or Review finishes with something to act on, an interactive terminal offers to resolve it right there: it shows you what was found, its own recommendation for how to address it when it has one, and lets you confirm that, choose a different approach, or skip it. Once you're done, Gnomon re-runs the same evaluation fresh, so the result you see is always a real re-check, never the fix's own say-so.

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

- **Project setup:** [Initialization](cli/PROJECT_INITIALIZATION.md) · [Bootstrap](workflows/bootstrap.md)
- **Defining features:** [Specification Definition](workflows/specification-definition.md) · [Specification Readiness Criteria](evaluations/SPECIFICATION_READINESS_CRITERIA.md)
- **Specification lifecycle:** [Lifecycle](specifications/SPECIFICATION_LIFECYCLE.md) · [Dependencies](specifications/SPECIFICATION_DEPENDENCIES.md)
- **Evaluating work:** [Testing](workflows/testing.md) · [Verification](workflows/verification.md) · [Review](workflows/review.md)
- **Workflows and customization:** [Core Distribution](cli/CORE_DISTRIBUTION.md) · [Workflow Contracts](cli/WORKFLOW_CONTRACT.md) · [Result Protocol](cli/RESULT_PROTOCOL.md)

[Releases](https://github.com/yasintqvi/gnomon/releases) · [Report an issue](https://github.com/yasintqvi/gnomon/issues)

## License

[MIT](LICENSE)