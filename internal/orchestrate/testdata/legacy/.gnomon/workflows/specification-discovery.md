---
identity: specification-discovery
specification_reference: none
requires_approved_specification: false
result:
  terminal_path: outcome
  classification:
    CANDIDATE_PROPOSED: success
    NO_CANDIDATE_IDENTIFIED: success
    BLOCKED: blocked
  schema:
    type: object
    required:
      - outcome
    properties:
      outcome:
        type: string
        enum:
          - CANDIDATE_PROPOSED
          - NO_CANDIDATE_IDENTIFIED
          - BLOCKED
      candidate_title:
        type:
          - string
          - "null"
      candidate_identity:
        type:
          - string
          - "null"
      rationale:
        type:
          - string
          - "null"
      derived_from:
        type:
          - string
          - "null"
      non_blocking_caveat:
        type:
          - string
          - "null"
      remaining_knowledge_gap:
        type:
          - string
          - "null"
---

# Specification Discovery Workflow

## Purpose

Help the Human identify, one at a time, a use case that appears to deserve its own Specification next, given the project knowledge that currently exists — and, whether a candidate comes from that reasoning or the Human already knows what they want, create it deterministically as a new Draft Specification with a stable, collision-free identity.

Specification Discovery does not define a Specification's behavior, does not approve a Specification, does not modify an existing Specification, and does not record or enforce dependency relationships between Specifications.

## When to Use

Use this workflow's Discovery reasoning (Steps 1–7) when the Human wants help identifying what Specification to create next. Use its Draft Creation mechanism directly, without Discovery's reasoning, when the Human already knows the title they want — Discovery is optional; Draft Creation is not gated behind it.

Do not use Discovery to define a Specification's behavior — that is [`workflows/specification-definition.md`](specification-definition.md). Do not use it to grant, revoke, or simulate approval — that is [`SPECIFICATION_LIFECYCLE.md`](../specifications/SPECIFICATION_LIFECYCLE.md). Do not use it to record or enforce dependencies between Specifications — that is deliberately out of scope here.

## Inputs

- Current project knowledge — particularly `PROJECT.md`'s Key Capabilities and Scope, and `DOMAIN.md` — describing what the project should do
- The actual content of every existing Specification, Draft or Approved
- The Specification template
- A title, when Draft Creation is invoked directly without Discovery's reasoning

## Execution

### 1. Inspect Project Knowledge

Read `PROJECT.md` and `DOMAIN.md` (and any other applicable `context/` knowledge) for capabilities, goals, or business concepts describing what the project should do.

### 2. Inspect Existing Specifications

Read every existing Specification's actual content — its Use Case actor and goal — regardless of lifecycle state or who created it. Treat every existing Specification, Draft or Approved alike, as already-represented use-case territory.

### 3. Identify a Candidate

Identify one capability or business concept not yet represented by any existing Specification's Use Case. A valid candidate has one plausible actor and one plausible goal — a bounded use case, not a theme. If what's uncovered is broader than one actor+goal pair, narrow it to the smallest piece that stands alone; never propose the broader theme itself.

### 4. Check for Duplicates and Theme Shape

Compare the candidate's actor and goal against every Specification inspected in Step 2. Discard it if it substantially overlaps an existing one. Discard it if it cannot be reduced to a single bounded actor+goal pair — propose the narrower piece instead, or continue searching.

### 5. Determine Outcome

- If a valid, non-duplicate, bounded candidate was identified, proceed to Step 6.
- If no such candidate can be identified from current project knowledge without inventing what it should be, report `NO_CANDIDATE_IDENTIFIED`.
- If current project knowledge is itself insufficient to identify any candidate — for example `PROJECT.md`'s Key Capabilities and Scope are unfilled and the Human has not otherwise clarified what the project should do — report `BLOCKED` rather than inventing material product behavior to manufacture a candidate.

### 6. Present the Candidate

Report exactly one candidate: its title, a short explanation, the rationale for why it deserves an independent Specification now — which may mention other Specifications by name in prose, but must never record or enforce a formal dependency — and the authoritative knowledge it was derived from. Never define its flows, rules, preconditions, or acceptance criteria; that is Specification Definition's responsibility, not this one's.

Alongside the title, also provide a concise semantic identity for the candidate: a few lowercase words capturing its core action and subject (for example, title "Create a Project", identity "create project"). The title is for the Human to read; the identity is what Gnomon mechanically normalizes into the Specification's filename. Do not construct the filename, SPEC number, path, or kebab-case form yourself — that normalization is Gnomon's responsibility, not yours. Keep the identity a genuine semantic summary, not merely the title with minor words stripped.

### 7. Report Result

Produce the structured result described in Outputs. Create nothing yet — creation happens only after the Human decides.

## Outputs

```
Specification Discovery Result

Outcome: CANDIDATE_PROPOSED | NO_CANDIDATE_IDENTIFIED | BLOCKED
Candidate title:
Candidate identity:
Rationale:
Derived from:
Non-blocking caveat:
Remaining knowledge gap:
```

- **Outcome** — `CANDIDATE_PROPOSED` when Step 6 presents a candidate; `NO_CANDIDATE_IDENTIFIED` when none can be identified from current project knowledge without inventing one — a statement about current knowledge only, never a claim that the Specification set is complete; `BLOCKED` when current project knowledge is itself insufficient to identify any candidate without inventing material product behavior.
- **Candidate title** — present only when `CANDIDATE_PROPOSED`.
- **Candidate identity** — present only when `CANDIDATE_PROPOSED`: the concise semantic identity described in Step 6, separate from the title, for Gnomon's own mechanical filename normalization.
- **Rationale** — present only when `CANDIDATE_PROPOSED`: the short explanation and why this candidate deserves an independent Specification now.
- **Derived from** — present only when `CANDIDATE_PROPOSED`: the specific authoritative knowledge the proposal is grounded in (for example, "`PROJECT.md` Key Capabilities").
- **Non-blocking caveat** — optional, any Outcome: something the Human should know that does not block the result.
- **Remaining knowledge gap** — present only when `BLOCKED`: what is missing from current project knowledge that prevents identifying a candidate.

## Human Decision

Presenting a candidate is not itself a decision. The Human alone decides what happens next:

- **Create** — proceed to Draft Creation for the presented candidate.
- **Skip** — do not create this candidate now. Skipping has no effect on Core beyond this interaction; it is not recorded anywhere, and a later Discovery run may propose the same candidate again if it is still not represented by an existing Specification.
- **Tell me more** — continue explaining the same candidate already presented. This requires no new workflow step or result value, only further conversation grounded in the same rationale.
- **Stop Discovery** — end this interaction. This is not a claim that the project's Specification set is complete; it only means no further candidate is proposed in this run.

## Draft Creation

Draft Creation applies identically whether it follows an accepted Discovery proposal or a Human directly requesting a new Specification with a known title. Either way, no Agent reasoning is required or permitted to determine the identity or generate behavioral content.

Given a title — the accepted candidate's title, or a title the Human supplies directly:

1. Determine the next Specification identity (see Specification Identity).
2. Copy the Specification template, substituting only the identity and the given title. When creation follows an accepted Discovery proposal, Gnomon derives the Specification's filename from the candidate's own semantic identity (Step 6), never from the title text itself.
3. The new Specification is `Draft`. This follows directly from [`SPECIFICATION_LIFECYCLE.md`](../specifications/SPECIFICATION_LIFECYCLE.md) — a new Specification is effectively Draft; nothing here changes or needs to change that.

Draft Creation never writes flows, rules, preconditions, or acceptance criteria. That content, if any, is [`workflows/specification-definition.md`](specification-definition.md)'s responsibility, invoked separately and only when the Human chooses to define the new Draft.

## Specification Identity

A Specification's identity is the numeric ID in its filename (`SPEC-NNN-...`):

- `SPEC-000` is permanently reserved for the template and must never be assigned to an actual Specification.
- The next identity is the highest numeric ID currently present among `specifications/SPEC-NNN-*.md` files, plus one — starting at `001` when no such file exists yet.
- Identities are zero-padded to at least three digits.
- This rule depends only on the current state of the `specifications/` directory — never on history, Agent judgment, or the candidate's title. Applied identically by Discovery-driven and manual creation, it always produces the same next identity for the same repository state.

This rule does not reuse an identity as long as a higher-numbered Specification file still exists. If the highest-numbered Specification file is ever deleted, its identity becomes available to be assigned again, because this rule has no memory beyond current repository state. Deleting an Approved Specification's file is already a destructive action outside this workflow's scope; this rule does not independently guard against the identity collision that could follow.

## Rules

- Never define a candidate's flows, rules, preconditions, or acceptance criteria; propose identity only — title, explanation, rationale, and derivation.
- Provide the candidate's semantic identity as a genuine, concise summary, never the title merely stripped of minor words; never construct the filename, SPEC number, or path yourself.
- Never invent material product behavior to manufacture a candidate; report `BLOCKED` instead when current project knowledge is insufficient.
- Treat every existing Specification — Draft or Approved, Discovery-created or manually created — as already-represented use-case territory.
- Propose exactly one candidate per run; never a batch, queue, ranking, or project-wide plan.
- Reasoning about why a candidate is useful now may be expressed in prose and may mention other Specifications by name; never record, persist, or enforce a formal dependency relationship.
- Never modify an existing Specification's content, regardless of what a candidate search reveals about it; report a finding instead.
- Never grant, revoke, or simulate Human approval, and never perform Specification Definition.
- `NO_CANDIDATE_IDENTIFIED` is a statement about current project knowledge only; never represent or imply that the project's Specification set is permanently complete.
- Skip has no persisted effect; do not record it anywhere Core would later consult.
- Stopping Discovery ends the current interaction only; never represent it as, or treat it as evidence of, completeness.
- Draft Creation never requires Agent reasoning to determine identity or generate content; the identity rule and template copy apply identically whether invoked from an accepted proposal or directly from a Human-supplied title.

## Failure Handling

### Insufficient Project Knowledge

Current project knowledge does not describe enough about what the project should do to identify even one candidate without inventing it. Report `BLOCKED` with the specific gap, rather than guessing at product behavior.

### Ambiguous Theme

What appears uncovered resolves to more than one plausible actor+goal pair rather than a single bounded use case. Narrow to the smallest piece that stands alone and propose that; if no single piece can be isolated without a judgment call that amounts to inventing a product decision, treat this the same as insufficient knowledge.

## Completion Criteria

Specification Discovery is complete when current project knowledge and every existing Specification's actual content have been inspected; at most one candidate has been identified and checked against every existing Specification for duplication and theme shape; the reported Outcome accurately reflects `CANDIDATE_PROPOSED`, `NO_CANDIDATE_IDENTIFIED`, or `BLOCKED` as actually determined; no candidate's behavior has been defined, no existing Specification has been modified, and no dependency relationship has been recorded; and — if the Human accepted a proposal or supplied a title directly — Draft Creation has produced a new Specification with a correctly assigned identity, in its lifecycle-default `Draft` state, containing only its identity and title.
