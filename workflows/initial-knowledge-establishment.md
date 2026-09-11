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
- Gnomon's ownership model for `PROJECT.md`, `DOMAIN.md`, `ARCHITECTURE.md`, `STACK.md`, and `CONVENTIONS.md`
- Any ADRs already related to the project

## Execution

### 1. Inspect

Inspect the repository before asking anything. Read existing `context/` documents, distinguishing genuinely approved project content from unfilled Gnomon template placeholders (bracketed text such as `[Project name]`), and review any prior code, configuration, or documentation that carries project intent. Do not assume the project is empty.

### 2. Reconcile Intent With Existing Knowledge

Compare the stated initial intent with what Inspect found. Identify what is already established and consistent, what conflicts, and what is entirely absent.

### 3. Classify Unknowns

Sort open questions into:

- **Material unknowns** — knowledge Bootstrap cannot proceed without, because leaving it undefined would force Bootstrap to invent a project-wide product, domain, architecture, stack, or convention decision.
- **Non-blocking unknowns** — knowledge a later workflow (Bootstrap, Implementation, or a future Specification) can safely resolve when it actually becomes material.

Treat a candidate domain concept or business rule like any other candidate: classify it as material only when an architecture, stack, or scope decision actually depends on knowing it. `DOMAIN.md` is not a mandatory prerequisite for Bootstrap — leave it untouched when nothing else depends on it yet.

### 4. Ask

Present only the material unknowns to the human, as a bounded, specific set of questions. Do not ask about non-blocking unknowns, and do not ask a question whose only purpose is to complete a template section Bootstrap does not need filled.

### 5. Record Decisions in Their Owner

For each material unknown the human resolves, update the document that owns that knowledge — `PROJECT.md`, `DOMAIN.md`, `ARCHITECTURE.md`, `STACK.md`, or `CONVENTIONS.md` — at the depth Bootstrap actually needs, not full template completion. Do not duplicate the same statement across documents. Never invent a project-wide product, domain, architecture, stack, or convention decision that was not approved.

### 6. Create an ADR Only Where Independently Justified

A human answer does not by itself create an ADR. Update the authoritative owner identified in Step 5 whenever its knowledge changes; in addition, create an ADR only when establishing that knowledge also involved a significant architectural or technical decision whose rationale, alternatives, and consequences are worth preserving. The ADR supplements the authoritative document — it never substitutes for updating it.

### 7. Check Sufficiency

Repeat Steps 3–6 until Bootstrap could proceed without inventing any project-wide decision. Stop as soon as that is true; do not continue refining documents toward full template completion once sufficiency is reached.

### 8. Check Consistency

Re-read every document touched in Step 5, and any ADR created in Step 6, for contradiction or duplicated knowledge introduced by the recorded decisions.

### 9. Present for Approval

Present the resulting initial project knowledge to the human: what was recorded, and in which documents, kept clearly separate from a list of the remaining non-blocking unknowns. State plainly that approval covers only the recorded knowledge — deferred or non-blocking matters remain open and are not decided by this approval.

### 10. Complete

Initial Knowledge Establishment is complete, and Bootstrap becomes the applicable next workflow, only once the human has explicitly approved the recorded knowledge as presented in Step 9.

## Outputs

- Updated `context/` documents, limited to those whose owned knowledge is now material and approved
- Any ADR created under Step 6, linked to the document(s) it explains
- An explicit list of remaining non-blocking unknowns, kept separate from the approved knowledge
- The human's explicit approval of the recorded knowledge — not of the non-blocking unknowns
- Confirmation that Bootstrap can proceed without inventing a project-wide decision

## Rules

- Never invent a project-wide product, domain, architecture, stack, or convention decision.
- Ask the human only about material unknowns; do not force exhaustive upfront design or complete a template merely to make it look finished.
- Update only the document that owns the knowledge that changed; do not duplicate a statement across documents.
- A human answer does not automatically create an ADR; create one only when a significant architectural or technical decision's rationale, alternatives, and consequences should be preserved, in addition to — never instead of — updating the authoritative owner.
- Do not touch `DOMAIN.md` unless a business concept or rule is actually material to a Bootstrap-relevant decision.
- Do not discover or create Specifications, define use-case behavior, implement features, scaffold the application, install dependencies, or otherwise perform Bootstrap.
- Do not attempt to resolve every possible future project decision.
- Human approval covers only the knowledge actually recorded; never treat it as approving deferred or non-blocking matters.

## Failure Handling

### Insufficient Initial Intent

The stated intent is too thin to identify even the material unknowns. Ask a small set of clarifying questions about purpose, primary users, and core capability rather than inventing them; do not advance toward approval until a coherent, if minimal, purpose is established.

### Contradictory Existing Knowledge

Existing repository content conflicts with itself or with the stated intent. Identify the conflicting sources and ask the human to resolve the conflict; do not silently prefer one source over another.

### Human Defers a Material Decision

If the human declines to resolve an unknown this workflow classified as material, do not proceed to Step 9. Either wait for the decision, or, if the human explicitly narrows Bootstrap's scope so the decision is no longer material to it, reclassify it as non-blocking and record that narrowing alongside the approval.

## Completion Criteria

Initial Knowledge Establishment is complete when the repository and existing knowledge have been inspected; material unknowns have been distinguished from non-blocking ones; every material unknown has been resolved by the human and recorded in its correct authoritative owner; any independently justified ADR has been created and linked; recorded knowledge is internally consistent; remaining non-blocking unknowns are explicit and clearly separated from the approved baseline; and the human has explicitly approved the recorded knowledge, with that approval understood to cover only what was recorded, not the deferred unknowns. At that point Bootstrap is the applicable next workflow.
