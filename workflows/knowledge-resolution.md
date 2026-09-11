# Knowledge Resolution Workflow

## Purpose

Apply a human decision that resolves a material knowledge gap reported by another workflow: update every authoritative knowledge owner whose content the decision actually changes, preserve the rationale behind any significant decision made along the way, and return control to the workflow that was paused.

Knowledge Resolution does not originate product or architectural direction. It acts only after a human has resolved a reported gap, and it never treats a human answer as automatically mapping to a single document or automatically requiring an ADR.

## When to Use

Use this workflow when Bootstrap, Implementation, Testing, Verification, or Review has stopped affected work on a material missing or conflicting knowledge issue and a human has since provided the decision. Do not use it to invent a decision, and do not use it for knowledge changes that were not preceded by a reported gap and a human decision — those follow ordinary authoring of the owning document. Do not use it to establish a new project's initial baseline from scratch — that is [`workflows/initial-knowledge-establishment.md`](initial-knowledge-establishment.md).

## Inputs

- The reported gap: the affected artifact, the workflow and task that stopped, and why the issue was material
- The human decision resolving the gap
- The authoritative knowledge under `context/`, `specifications/`, `contracts/`, and `decisions/`
- Applicable Artifact Contracts and ADRs already related to the affected area

## Execution

### 1. Confirm the Decision

Restate the decision and the question it answers. If it does not cover the reported gap, or leaves part of it open, stop and return to the human rather than filling the remainder by inference.

### 2. Determine the Primary Knowledge Effect

Determine the authoritative owner of the knowledge the decision directly resolves, following Gnomon's existing ownership rules — for example, a Specification for undefined use-case behavior, or Domain for an undefined business concept — and update it precisely as approved, within its existing structure and scope.

Do not force a document choice merely to satisfy a single-owner assumption. If ownership of the primary effect is itself materially ambiguous, report the ambiguity rather than inventing an owner.

### 3. Determine Secondary Knowledge Consequences

Independently check whether the decision also changes knowledge owned elsewhere, using existing ownership boundaries:

- **Domain** — update when the decision establishes or changes a business rule that applies across more than the one use case that raised the gap.
- **Architecture** — update when the decision changes structural boundaries, component responsibilities, dependencies, or architectural constraints.
- **Stack or Conventions** — update when the decision changes a technology choice, tool, or project-wide convention owned by that document.
- **Contract** — update or create a contract when the decision establishes reusable, artifact-category behavior rather than behavior specific to one use case.

A decision may have zero, one, or several secondary consequences. Update only the documents whose owned knowledge actually changed; do not touch a document because it is plausibly related.

### 4. Determine Additional Engineering Decisions

Evaluate whether implementing the approved decision exposes a further engineering decision that was not part of the original gap — for example, implementing an approved behavior change reveals a missing architectural dependency. Exposing such a decision is not the same as having it approved: the original human decision never implicitly authorizes it.

Check whether existing authoritative knowledge already determines it. If it does, follow that knowledge; do not treat it as a gap. If it does not, treat it as its own material knowledge gap — stop only the affected work, present the new gap and the decision it requires to the human, and wait for an explicit decision before proceeding. Once decided, apply this workflow independently to that decision, starting from Step 1.

### 5. Create ADRs Where Independently Justified

Create an ADR only when resolving or implementing the decision required a significant architectural or technical decision whose rationale, alternatives, consequences, or trade-offs should be preserved. An ADR never substitutes for updating the authoritative document a change concerns: the document records what the system now requires; the ADR preserves why the significant choice was made. Link the ADR to the authoritative knowledge it explains, but reference the ADR back from an affected document only when an existing Gnomon convention calls for it (for example, a Specification's Related Decisions section) or when the reference is materially useful to understanding that document — do not modify a document solely to add a backlink.

### 6. Update Authoritative Knowledge

Apply every update identified in steps 2–5 to its owning document, preserving each document's existing structure, scope, and unrelated content.

### 7. Check Consistency

Re-read the affected sections of every document touched or referenced in steps 2–5, and any Specification, Contract, or ADR that references them, for contradiction or statements the change now makes stale. Resolve contradictions within the scope of the decision; report any that require a further decision rather than resolving them silently.

### 8. Resume

Return to the workflow and task that reported the gap, with the decision and updated knowledge available as Inputs. Resume only the work that was stopped; do not expand scope beyond the original task.

## Outputs

- Authoritative knowledge updated, limited to documents whose owned content the decision changed
- Any new or updated ADR, with rationale, alternatives, and consequences
- A short record distinguishing the primary effect, secondary consequences, and additional engineering decisions
- Confirmation that affected knowledge is free of contradiction or staleness introduced by the change
- The resumed workflow and task

## Rules

- Never resolve a gap without an explicit human decision, including an additional engineering decision exposed while implementing an already-approved one — that original decision is never implicit approval of it.
- Never force a single-document or single-owner choice for the primary effect; report materially ambiguous ownership instead of inventing an owner.
- Never assume a human decision always requires an ADR.
- An ADR never substitutes for updating the authoritative document it concerns, and a document needs no backlink to an ADR beyond what existing convention or genuine usefulness requires.
- Update only documents whose owned knowledge changed; do not duplicate the same statement across documents.
- Do not expand a decision beyond what the human approved.
- Keep each updated document internally consistent with the rest of its own content.

## Failure Handling

### Decision Does Not Resolve the Gap

Report the remaining ambiguity and return to the human rather than approximating an interpretation.

### New Gap Surfaces During Resolution

Treat it as a new material knowledge gap: stop only the affected part of resolution, identify and present it, and wait for its own decision before continuing.

### Consistency Check Finds Unrelated Contradictions

Report contradictions outside the current decision's scope rather than resolving them silently, and recommend the workflow that owns them.

## Completion Criteria

Knowledge Resolution is complete when the human decision has been confirmed; the primary and secondary effects have been determined; any additional engineering decision exposed by implementation has either been resolved by existing knowledge or independently decided by the human, never inferred from the original approval; every authoritative document whose owned knowledge changed has been updated; any independently justified ADR has been created and linked; affected knowledge has been checked for contradiction or staleness; and the originating workflow and task have resumed.
