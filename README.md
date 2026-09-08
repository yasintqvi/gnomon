<p align="center">
  <img src="assets/gnomon.png" alt="Gnomon logo" width="220" />
</p>

# AI Software Engineering System

A repository of Markdown templates for running AI-assisted software engineering as an explicit, document-driven process. It separates durable project knowledge, feature requirements, artifact behavior, architectural decisions, execution workflows, and evaluation criteria so an AI agent can load only the sources relevant to a task.

This repository is the system itself: it contains no application runtime, package dependencies, build step, or automated test suite.

## Principles

- AI assists; humans authorize scope and make product or architectural decisions.
- Approved documents are the source of truth.
- Each document has one responsibility.
- Workflows determine which knowledge must be loaded for a task.
- Verification records objective evidence; review applies engineering judgment.
- Prefer the smallest coherent change and preserve unrelated work.
- Add abstractions only when real usage demonstrates a need.

## How the Documents Fit Together

```text
User request
    |
    v
Choose an intent and workflow
    |
    v
Load applicable context, specifications, ADRs, and contracts
    |
    v
Execute the workflow within the authorized scope
    |
    v
Verify objectively ---> Review with engineering judgment
    |
    v
Optionally finalize the approved change in Git
```

Project knowledge is divided by responsibility:

| Area | Owns |
| --- | --- |
| `context/` | Project purpose, domain knowledge, architecture, technology choices, and conventions |
| `specifications/` | Use-case behavior, scope, flows, rules, and acceptance criteria |
| `contracts/` | Reusable behavioral boundaries for specific artifact types |
| `decisions/` | Significant architectural or technical decisions and their consequences |
| `workflows/` | Repeatable execution processes for common engineering intents |
| `evaluations/` | Reusable criteria for objective verification and judgment-based review |

Do not duplicate a rule across documents. For example, feature behavior belongs in a specification, cross-feature business rules belong in domain context, structural constraints belong in architecture, and artifact-specific behavior belongs in a contract.

## Repository Contents

```text
.
├── context/
│   ├── PROJECT.md
│   ├── DOMAIN.md
│   ├── ARCHITECTURE.md
│   ├── STACK.md
│   ├── CONVENTIONS.md
│   ├── PRODUCT_EXPERIENCE.md
│   ├── UI_FOUNDATION.md
│   └── INTERACTION_PATTERNS.md
├── specifications/
│   └── SPEC-001-use-case-name.md
├── contracts/
│   ├── behavior-verification.md
│   ├── input-qualification.md
│   ├── interactive-presentation.md
│   ├── persistent-structure-evolution.md
│   └── use-case-execution.md
├── decisions/
│   └── ADR-001-decision-title.md
├── workflows/
│   ├── bootstrap.md
│   ├── implementation.md
│   ├── testing.md
│   ├── verification.md
│   ├── review.md
│   └── git-finalization.md
├── evaluations/
│   ├── VERIFICATION_CRITERIA.md
│   └── ENGINEERING_REVIEW_QUESTIONS.md
├── LICENSE
└── README.md
```

The files under `context/`, along with the sample specification and ADR, are intentionally unfilled templates. Bracketed text such as `[Project name]` marks content to replace for a real project. The contracts, workflows, and evaluation assets are reusable baselines that may be adapted when a project's established rules require it. `PRODUCT_EXPERIENCE.md`, `UI_FOUNDATION.md`, and `INTERACTION_PATTERNS.md` are also reusable baselines rather than templates: unlike the other `context/` files, they are already fully authored and, together, form the Design Knowledge referenced by the workflows.

## Document Catalog

### Context

- [`PROJECT.md`](context/PROJECT.md) — project purpose, goals, users, scope, capabilities, constraints, and success criteria.
- [`DOMAIN.md`](context/DOMAIN.md) — implementation-independent business concepts, terminology, relationships, rules, and boundaries.
- [`ARCHITECTURE.md`](context/ARCHITECTURE.md) — components, dependencies, interactions, placement rules, structural patterns, and constraints.
- [`STACK.md`](context/STACK.md) — languages, runtimes, frameworks, libraries, infrastructure, tools, and version policy.
- [`CONVENTIONS.md`](context/CONVENTIONS.md) — naming, organization, formatting, documentation, testing, error-handling, and consistency rules.
- [`PRODUCT_EXPERIENCE.md`](context/PRODUCT_EXPERIENCE.md) — product information architecture, navigation, cross-feature journeys, and discoverability that make independently implemented features feel like one product.
- [`UI_FOUNDATION.md`](context/UI_FOUNDATION.md) — shared visual and presentation foundation: design tokens, color, typography, layout, responsive containment, and theming.
- [`INTERACTION_PATTERNS.md`](context/INTERACTION_PATTERNS.md) — shared interaction behavior such as feedback, confirmations, loading, forms, and collection interaction, so independently implemented features behave consistently.

These three are collectively referred to as Design Knowledge by the workflows below.

### Requirements and decisions

- [`SPEC-001-use-case-name.md`](specifications/SPEC-001-use-case-name.md) — template for one bounded use case, including flows and acceptance criteria. Copy it for additional specifications and assign each one a unique ID.
- [`ADR-001-decision-title.md`](decisions/ADR-001-decision-title.md) — template for an architectural decision record, including alternatives, consequences, constraints, and impact. Copy it for additional decisions and assign each one a unique ID.

### Artifact contracts

- [`use-case-execution.md`](contracts/use-case-execution.md) — execution and effect coordination for one declared use case.
- [`interactive-presentation.md`](contracts/interactive-presentation.md) — accessible presentation of state and user intent.
- [`input-qualification.md`](contracts/input-qualification.md) — validation and safe normalization of untrusted input.
- [`behavior-verification.md`](contracts/behavior-verification.md) — executable verification of observable approved behavior.
- [`persistent-structure-evolution.md`](contracts/persistent-structure-evolution.md) — safe and reproducible evolution of durable data structures.

Contracts define invariant behavior and verification expectations for an artifact category. A specification still owns feature-specific behavior; architecture owns placement and dependency boundaries.

### Workflows

| Intent | Workflow | Result |
| --- | --- | --- |
| Establish a missing project baseline | [`bootstrap.md`](workflows/bootstrap.md) | Minimal verified structure and configuration required for later work |
| Build or change approved behavior | [`implementation.md`](workflows/implementation.md) | Coherent implementation plus verification evidence |
| Design, add, or execute behavioral tests | [`testing.md`](workflows/testing.md) | Test implementation and reproducible test results |
| Check objective compliance | [`verification.md`](workflows/verification.md) | Per-obligation `PASS`, `FAIL`, or `UNVERIFIABLE` evidence |
| Evaluate engineering quality and risk | [`review.md`](workflows/review.md) | `PASS`, `DEFECT`, `RISK`, `KNOWLEDGE GAP`, or `NOT APPLICABLE` findings |
| Prepare and optionally publish Git work | [`git-finalization.md`](workflows/git-finalization.md) | Scoped branch/commit preparation and, only when authorized, push or pull request |

Verification and review are deliberately separate. Verification asks whether traceable obligations are objectively satisfied. Review consumes available evidence and applies engineering reasoning; it identifies findings and their owner but does not silently resolve missing product or architectural decisions.

### Evaluation assets

- [`VERIFICATION_CRITERIA.md`](evaluations/VERIFICATION_CRITERIA.md) — reusable checks spanning specifications, domain, architecture, contracts, stack, conventions, documentation, tests, static analysis, and formatting.
- [`ENGINEERING_REVIEW_QUESTIONS.md`](evaluations/ENGINEERING_REVIEW_QUESTIONS.md) — judgment-based questions covering authorization, authentication, data integrity, state, failures, security, asynchronous or AI work, maintainability, and completeness.

## Getting Started

1. Copy this repository, or copy its document directories into the target repository.
2. Replace the placeholders in `context/`, starting with `PROJECT.md` and `DOMAIN.md`, then document the architecture, stack, and conventions that actually exist.
3. Copy and rename the specification template for each approved use case. Keep requirements and acceptance criteria explicit.
4. Copy and rename the ADR template whenever a significant decision needs a durable record.
5. Select the workflow that matches the user's intent. Follow its required-knowledge section rather than loading every document automatically.
6. Apply only the contracts relevant to the artifacts being created or changed.
7. Use verification criteria to produce objective evidence, then use review questions when engineering judgment is required.
8. Use Git finalization only after the change is complete, and treat pushing or opening a pull request as a separate authorization gate.

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

## Extending the System

Add a document only when it has a distinct, demonstrated responsibility:

- Add a specification for a new bounded use case.
- Add an ADR for a significant decision.
- Add a contract when an artifact category has reusable behavioral invariants not owned elsewhere.
- Add or revise a workflow when a recurring engineering intent needs a stable execution process.
- Add evaluation criteria when a reusable verification obligation or review question is missing.

Keep new documents focused, link them to their authoritative sources, and avoid turning examples or preferences into project requirements.

## License

Released under the [MIT License](LICENSE).
