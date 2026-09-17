---
identity: initial-knowledge-establishment
specification_reference: none
requires_approved_specification: false
result:
  terminal_path: outcome
  classification:
    READY_FOR_BOOTSTRAP: success
    BLOCKED: blocked
  schema:
    type: object
    required:
      - outcome
    properties:
      outcome:
        type: string
        enum:
          - READY_FOR_BOOTSTRAP
          - BLOCKED
      recorded_knowledge:
        type:
          - string
          - "null"
      non_blocking_deferred_items:
        type:
          - string
          - "null"
      remaining_unresolved:
        type:
          - string
          - "null"
---

# Initial Knowledge Establishment Workflow

## Purpose

Convert a new project's initial intent, together with any existing repository or project information, into the minimum approved project-level knowledge required for Bootstrap to proceed without inventing project-wide decisions.

Initial Knowledge Establishment does not implement features, define use-case behavior, scaffold the application, or perform Bootstrap itself.

## When to Use

Use this workflow before Bootstrap, when `context/` does not yet hold sufficient approved project-level knowledge — for example a fresh copy of Gnomon's unfilled templates, or a partially started project whose `PROJECT.md`, `ARCHITECTURE.md`, `STACK.md`, or `CONVENTIONS.md` content is incomplete or contradicts the stated intent.

Do not use it for a knowledge gap raised while an existing workflow is already running — that is [`workflows/knowledge-resolution.md`](knowledge-resolution.md). Do not use it to discover or author Specifications, and do not use it to perform Bootstrap itself.

## Inputs

- The user's initial project intent
- The existing repository as found: any partially filled `context/` documents, prior code, configuration, or documentation
- Gnomon's ownership model for `PROJECT.md`, `DOMAIN.md`, `ARCHITECTURE.md`, `STACK.md`, `CONVENTIONS.md`, and — when Bootstrap's authorized scope requires it — `PRODUCT_EXPERIENCE.md`, `UI_FOUNDATION.md`, and `INTERACTION_PATTERNS.md`
- Bootstrap's own Execution steps and Rules (`workflows/bootstrap.md`), as the reference point for what is material
- Any ADRs already related to the project

## Execution

### 1. Inspect

Inspect the repository before asking anything. Read existing `context/` documents, distinguishing genuinely approved project content from unfilled Gnomon template placeholders (bracketed text such as `[Project name]`), and review any prior code, configuration, or documentation that carries project intent. Do not assume the project is empty.

### 2. Reconcile Intent With Existing Knowledge

Compare the stated initial intent with what Inspect found. Identify what is already established and consistent, what conflicts, and what is entirely absent.

### 3. Determine Bootstrap's Authorized Scope

Establish, at the level Bootstrap itself operates at, what this project's Bootstrap run will actually Scaffold, Configure, Lock Dependencies, and Validate — for example whether a UI foundation is in scope, which runtime and package manager apply, and the top-level component or module shape. This scope is the fixed reference point for every materiality decision in Step 4; do not expand it to a hypothetical future or out-of-scope subsystem (`workflows/bootstrap.md`'s own rule against blocking on decisions needed only by future or out-of-scope subsystems applies here too).

### 4. Classify Unknowns

Sort open questions into:

- **Material unknowns** — knowledge Bootstrap's Scaffold, Configure, Lock Dependencies, or Validate responsibilities cannot proceed through, for the scope determined in Step 3, without inventing a decision or making an unsafe assumption (`workflows/bootstrap.md`'s own rules against inventing project-level decisions and treating assumptions as requirements apply here too).
- **Non-blocking unknowns** — everything else: knowledge a later workflow (Bootstrap, Implementation, or a future Specification) can safely resolve when it actually becomes material.

Template completeness never determines materiality — a section may stay unfilled indefinitely if nothing in the scope determined in Step 3 depends on it.

Treat a candidate domain concept or business rule like any other candidate: classify it as material only when an architecture, stack, or scope decision actually depends on knowing it. `DOMAIN.md` is not a mandatory prerequisite for Bootstrap — leave it untouched when nothing else depends on it yet.

Apply the same test to Design Knowledge (`PRODUCT_EXPERIENCE.md`, `UI_FOUNDATION.md`, `INTERACTION_PATTERNS.md`). The project having a frontend or UI does not by itself make Design Knowledge material — these documents are already fully authored, reusable baselines and apply as-is by default. They become material only when the scope determined in Step 3 requires a specific design decision the existing baseline does not already cover, and Bootstrap would otherwise have to invent it.

### 5. Ask

Present only the material unknowns to the human, as a bounded, specific set of questions. Do not ask about non-blocking unknowns, and do not ask a question whose only purpose is to complete a template section Bootstrap does not need filled.

### 6. Record Decisions in Their Owner

For each material unknown the human resolves, update the document that owns that knowledge — `PROJECT.md`, `DOMAIN.md`, `ARCHITECTURE.md`, `STACK.md`, `CONVENTIONS.md`, or, when Design Knowledge is material per Step 4, `PRODUCT_EXPERIENCE.md`, `UI_FOUNDATION.md`, or `INTERACTION_PATTERNS.md` — at the depth Bootstrap actually needs, not full template completion. Do not duplicate the same statement across documents. Never invent a project-wide product, domain, architecture, stack, convention, or design decision that was not approved.

### 7. Create an ADR Only Where Independently Justified

A human answer does not by itself create an ADR. Update the authoritative owner identified in Step 6 whenever its knowledge changes; in addition, create an ADR only when establishing that knowledge also involved a significant architectural or technical decision whose rationale, alternatives, and consequences are worth preserving. The ADR supplements the authoritative document — it never substitutes for updating it.

### 8. Check Sufficiency

Repeat Steps 3–7 until no Bootstrap-blocking material decision remains for the scope determined in Step 3. Stop as soon as that is true; do not continue refining documents toward full template completion once sufficiency is reached.

### 9. Check Consistency

Re-read every document touched in Step 6, and any ADR created in Step 7, for contradiction or duplicated knowledge introduced by the recorded decisions.

### 10. Confirm Record Accuracy

Once Step 8 (Check Sufficiency) is satisfied and Step 9 (Check Consistency) is clean, present the human with a single summary: what was recorded, and in which documents, kept clearly separate from the remaining non-blocking unknowns. This confirms only that the written record accurately reflects the decisions already made in Step 6 — it is not a re-request for approval of any individual decision, and it happens exactly once, not per loop iteration. State plainly that this confirmation covers only the recorded knowledge — deferred or non-blocking matters remain open and are not decided by it.

### 11. Complete

Initial Knowledge Establishment is complete, and Bootstrap becomes the applicable next workflow, only once the human has confirmed the record's accuracy as presented in Step 10.

## Outputs

```
Initial Knowledge Establishment Result

Outcome: READY_FOR_BOOTSTRAP | BLOCKED
Recorded knowledge:
Non-blocking deferred items:
Remaining unresolved (BLOCKED only):
```

- **Outcome** — `READY_FOR_BOOTSTRAP` only when no Bootstrap-blocking material decision remains, within the scope determined in Step 3, and the human has confirmed the recorded knowledge's accuracy (Step 10); `BLOCKED` otherwise.
- **Recorded knowledge** — informational: the `context/` documents updated this run — including Design Knowledge documents when material — limited to those whose owned knowledge is now material and recorded, and any ADR created under Step 7, linked to the document(s) it explains.
- **Non-blocking deferred items** — informational: the explicit list of remaining non-blocking unknowns, kept separate from the recorded knowledge.
- **Remaining unresolved** — populated only when `BLOCKED`: the material decision(s) still open.

`READY_FOR_BOOTSTRAP` reflects the human's confirmation that the recorded knowledge accurately reflects the decisions made — never approval of the non-blocking unknowns, and never re-approval of any individual decision.

## Rules

- Never invent a project-wide product, domain, architecture, stack, convention, or design decision.
- Judge materiality only against the scope determined in Step 3 and Bootstrap's own Scaffold, Configure, Lock Dependencies, and Validate responsibilities (`workflows/bootstrap.md`) — never against template completeness.
- Ask the human only about material unknowns; do not force exhaustive upfront design or complete a template merely to make it look finished.
- One unresolved material decision stops only the work that depends on it; continue resolving unrelated material unknowns in the meantime.
- Update only the document that owns the knowledge that changed; do not duplicate a statement across documents.
- A human answer does not automatically create an ADR; create one only when a significant architectural or technical decision's rationale, alternatives, and consequences should be preserved, in addition to — never instead of — updating the authoritative owner.
- Do not touch `DOMAIN.md` unless a business concept or rule is actually material to a Bootstrap-relevant decision.
- Do not treat Design Knowledge as material merely because the project has a frontend or UI; `PRODUCT_EXPERIENCE.md`, `UI_FOUNDATION.md`, and `INTERACTION_PATTERNS.md` apply as-is unless the scope determined in Step 3 requires a decision they do not already cover.
- Do not discover or create Specifications, define use-case behavior, implement features, scaffold the application, install dependencies, or otherwise perform Bootstrap.
- Do not attempt to resolve every possible future project decision.
- The final confirmation covers only the accuracy of the recorded knowledge; never treat it as approving deferred or non-blocking matters, and never use it to re-request approval of a decision already made.

## Failure Handling

### Insufficient Initial Intent

The stated intent is too thin to identify even the material unknowns. Ask a small set of clarifying questions about purpose, primary users, and core capability rather than inventing them; do not advance toward approval until a coherent, if minimal, purpose is established.

### Contradictory Existing Knowledge

Existing repository content conflicts with itself or with the stated intent. Identify the conflicting sources and ask the human to resolve the conflict; do not silently prefer one source over another.

### Human Defers a Material Decision

Continue resolving unrelated material unknowns; stop only the work that depends on the deferred one, and do not proceed to Step 10 while any Bootstrap-blocking material decision remains open. If the human explicitly narrows Bootstrap's authorized scope (Step 3) so the decision is no longer material to it, reclassify it as non-blocking and record that narrowing alongside the confirmation.

## Completion Criteria

Initial Knowledge Establishment is complete when the repository and existing knowledge have been inspected; Bootstrap's authorized scope has been determined; material unknowns — including Design Knowledge when applicable — have been distinguished from non-blocking ones against that scope, not against template completeness; no Bootstrap-blocking material decision remains unresolved; every resolved decision has been recorded in its correct authoritative owner; any independently justified ADR has been created and linked; recorded knowledge is internally consistent; remaining non-blocking unknowns are explicit and clearly separated from the recorded knowledge; and the human has confirmed, once, that the recorded knowledge accurately reflects the decisions made. At that point Bootstrap is the applicable next workflow.
