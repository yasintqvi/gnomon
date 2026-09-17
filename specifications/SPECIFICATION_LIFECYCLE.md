# Specification Lifecycle

## Purpose

Define the lifecycle and approval semantics that govern every Specification: the states a Specification's content can be in, who has authority over those states, what approval applies to, and how approval relates to later changes in content.

This document defines a semantic contract, not a runtime mechanism. It does not implement, compute, store, or enforce anything itself; it defines what any future implementation — CLI, Agent adapter, or persistence mechanism — must satisfy.

---

## Scope

This document covers:

- The Specification lifecycle states and what each one means.
- Who may change lifecycle state, and who may not.
- What content approval applies to, and how to treat ambiguity.
- The required properties of approval evidence, described conceptually.
- How restoration of prior content and human revocation affect lifecycle state.
- The boundary between this lifecycle and other, similarly named concerns.

This document does not cover:

- Specification completeness or readiness for approval — a Specification Definition workflow's responsibility.
- How Specifications are discovered or identified — a Specification Discovery concern. How one Specification depends on another — defined in [`SPECIFICATION_DEPENDENCIES.md`](SPECIFICATION_DEPENDENCIES.md), never established by Discovery itself.
- How approval-relevant content is compared, computed, or normalized, or how approval evidence is captured or persisted — a later runtime/tooling design.
- Implementation, verification, review, or other workflow-execution or result state.

---

## Lifecycle States

A Specification has exactly two lifecycle states:

- **Draft** — the Specification's current approval-relevant content has no valid, matching human approval.
- **Approved** — the Specification's current approval-relevant content exactly matches valid, non-revoked approval evidence.

A newly created Specification starts as Draft.

Draft does not mean incomplete, undefined, or low-quality. A Draft may be empty, partially defined, fully and carefully authored by a human, or fully worked through by an Agent — Draft states only the absence of a currently valid human approval for the content as it now stands. Determining whether a Draft is complete or ready for approval is the responsibility of the Specification Definition workflow, not this lifecycle. A workflow result such as `READY_FOR_APPROVAL` is a workflow outcome, not a lifecycle state, and must never be treated as, stored as, or presented as equivalent to Approved.

Approved is not a permanent status and not an event that, once granted, holds independent of content. It is a standing relationship between content and evidence: a Specification is Approved only for as long as its current approval-relevant content matches valid approval evidence. It is not something a Specification "becomes" once and keeps regardless of later edits.

---

## Authority

- **Human** — the only party who may grant or revoke approval. Granting approval affirms, after review, that the current approval-relevant content is correct. Revoking approval invalidates a previously valid approval without requiring any change to content.
- **Agent** — may author, edit, or restructure a Specification's content, and may determine and report whether it appears ready for approval. An Agent must never grant, revoke, or simulate approval, and must never present a readiness assessment as equivalent to approval.
- **Workflow** — may consume an Approved Specification's content and may report a Specification's current lifecycle state, but has no authority to change that state. Assessing readiness ([`workflows/specification-definition.md`](../workflows/specification-definition.md)) is a workflow responsibility; granting approval is not, and never becomes one.
- **CLI or other tooling** (not designed here) — may present lifecycle state and offer an approval or revocation action to a human, but must not itself decide, infer, or default to approval on the human's behalf.

This applies Gnomon's existing principle — AI assists while humans authorize scope and decisions — as a specific, structural rule for Specification approval rather than a general expectation.

---

## What Approval Applies To

Approval is not granted to a Specification's identity in the abstract, and it does not apply "going forward" regardless of later edits. Approval applies to the exact approval-relevant content the human reviewed at the moment they granted it.

**Approval-relevant content** is a Specification's authored semantic content — everything that defines the business requirements, behavior, rules, and constraints it states. All authored semantic content is approval-relevant.

**Representation-only differences are not approval-relevant** — for example, non-semantic formatting or whitespace changes that do not alter what the content states.

When it is ambiguous whether a change is representation-only or semantic, treat it as approval-relevant. A false "still Approved" on content a human never actually reviewed is a materially worse failure than an unnecessary request to re-review; this lifecycle must never resolve ambiguity toward silently preserving Approved status.

This document does not enumerate every Specification section's exact treatment, and does not specify how approval-relevant content is compared, computed, or normalized — that is a deferred, later design and implementation concern. What is fixed here is the principle any such mechanism must satisfy: it must distinguish semantic content from pure representation, and it must default to treating ambiguous cases as approval-relevant.

---

## Approval Evidence

Lifecycle state is derived, not stored. A Specification's own content must never carry an editable status field — such as a `Status: Approved` marker — that a normal content edit could set or change. Whether a Specification is Draft or Approved is computed by comparing its current approval-relevant content against approval evidence recorded elsewhere.

Approval evidence must have the following properties, independent of how it is eventually persisted:

- **Content-bound** — bound to the exact approval-relevant content it was granted for, such that a later comparison can determine whether current content still matches it.
- **Conceptually separate from authored Specification content** — never stored inside the Specification as editable prose, and never itself a second editable source of truth for what the Specification states.
- **Attributable to an explicit Human action** — produced only by a human's deliberate act of granting approval, never by an Agent's or workflow's ordinary content-authoring operations.
- **Bound to a stable Specification identity** — unambiguously associated with the Specification it concerns, independent of a mutable file path or filename. That identity is assigned by [`workflows/specification-discovery.md`](../workflows/specification-discovery.md)'s Specification Identity rule; this document only requires that evidence bind to it.
- **Durable** — persists independent of any single session, conversation, or tool invocation.
- **Immutable and append-only** — once recorded, an evidence record is never edited in place. A new approval action produces a new record; it never rewrites a prior one. Revocation invalidates an existing record's applicability; it does not delete or alter the historical record itself.

How evidence is captured, computed, or stored — hashing, snapshotting, a state file, a database, or any other mechanism — is not decided here and is explicitly deferred to later work.

---

## Derived State Rules

- If a Specification's current approval-relevant content does not match any valid, non-revoked approval evidence, its effective lifecycle state is **Draft**.
- If a Specification's current approval-relevant content exactly matches valid, non-revoked approval evidence, its effective lifecycle state is **Approved**.
- Editing approval-relevant content away from previously approved content makes the Specification Draft. This is deterministic and never depends on a judgment call about whether a change was "important enough."
- If content is later restored to exactly what previously valid, non-revoked approval evidence describes, the Specification is Approved again under that evidence — no new human action is required, because the human already reviewed and approved that exact content.
- Human revocation invalidates a specific approval evidence record's applicability going forward. It does not delete or alter that record's historical existence, and it does not itself change Specification content. A revoked approval can never make a Specification Approved again — even if content later matches the revoked record exactly — until a new, valid approval is granted.

---

## Boundaries — What Does Not Belong to This Lifecycle

- Changes to Domain knowledge, ADRs, or any other referenced project knowledge never silently change a Specification's lifecycle state, even when they affect what the Specification means in context. A Specification whose dependencies have shifted underneath it, while its own approval-relevant content is unchanged, remains Approved under this lifecycle. Whether it is still *correct* given that drift is a consistency and review concern — for Review's `KNOWLEDGE GAP` classification or Knowledge Resolution's consistency check to surface as a finding — never an automatic lifecycle transition.
- Implementation, verification, review, completion, blocked, failed, and any other workflow-execution or result state are not lifecycle states and must never be represented as, or conflated with, Draft or Approved. A Specification can be Approved and not yet implemented, or Approved and fully implemented; this lifecycle says nothing about either.
- Readiness for approval (a Specification Definition workflow result) is not a lifecycle state and must never substitute for Approved.

---

## Deferred to Future Work

The following are intentionally not defined here. They are runtime, tooling, or later-step design concerns that must satisfy the contract above, not redefine it:

- How approval-relevant content is compared, normalized, hashed, or snapshotted.
- How approval evidence is persisted, and where.
- How a Specification's stable identity is established and maintained — defined in [`workflows/specification-discovery.md`](../workflows/specification-discovery.md)'s Specification Identity rule; this document only requires that evidence bind to it.
- How a human grants, revokes, or reviews approval through any CLI, UI, or Agent interaction.
- When or how lifecycle state is recomputed — on read, on save, via a watcher, or otherwise.
- Specification completeness or readiness assessment (Specification Definition).
- Whether any workflow must treat a dependency as blocking, and dependency ordering/execution behavior — the meaning of a dependency itself is defined in [`SPECIFICATION_DEPENDENCIES.md`](SPECIFICATION_DEPENDENCIES.md); this document only requires that its content-change rule apply to a Specification's `Dependencies` section like any other authored content.
- Workflow execution and result state.

---

## Notes

[Optional clarification that does not belong to another section.]
