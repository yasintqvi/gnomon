# Workflow Contract v1 — CLI Step 1

## Purpose

Define the minimal, structured, machine-readable metadata the CLI needs from an Agent workflow file to decide, deterministically, whether it may start the Agent — without duplicating, overriding, or reinterpreting any semantic rule Core already owns.

It also defines the machine-readable **Result Contract** each workflow owns, letting the CLI deterministically validate an Agent's terminal result without hard-coding workflow-specific schemas in CLI source and without heuristically parsing Markdown prose.

This is CLI-level design, not Core semantics. It adds a structured surface to each workflow file; it does not change what any workflow means, requires, or does.

---

## Scope

This document covers:

- The exact fields Workflow Contract v1 requires, and their semantics.
- The representation of those fields (YAML frontmatter embedded in each workflow's own Markdown file).
- The source-of-truth relationship between structured frontmatter and workflow prose.
- The finalized field mapping for every current Agent workflow.
- The machine-readable Result Contract each workflow owns: the terminal-result path and payload schema the CLI uses to validate an Agent's result.
- The minimal structural validation model, for both the eligibility fields and the Result Contract.
- The single global contract/schema version and where it is recorded.

This document does not cover:

- Any Core semantic rule (lifecycle, dependency, readiness, verification, review) — those remain defined solely by their own Core documents and are never restated here as machine-checkable logic beyond what a workflow's own fields already express.
- The Bundled Default Core / Project-local Core distribution model — defined by CLI Step 2 (`CORE_DISTRIBUTION.md`).
- `.gnomon/` materialization and initialization behavior — defined by CLI Step 3 (`PROJECT_INITIALIZATION.md`).
- Agent adapters, provider configuration, the Result Protocol's transport, envelope, and runtime consumption flow — defined by CLI Step 5 (`RESULT_PROTOCOL.md`), which also defines how the Result Contract documented here is used at runtime.
- An eligibility expression language, dependency-state metadata, next-action metadata, or runtime execution — all deliberately absent from this contract.

---

## Contract Fields

Workflow Contract v1 consists of three scalar eligibility fields, described below, plus one structured Result Contract block, described in its own section. No further field, expression, or nested structure beyond these is part of v1.

### `identity`

**Type:** string.

A stable, declared workflow identity, independent of the file's name or path. Used by the CLI for routing, and to detect whether a recognized built-in workflow's frontmatter has drifted from its bundled default (a signal for reclassification, not itself designed here).

### `specification_reference`

**Type:** enum — `none | optional | required`.

Whether invoking this workflow may or must supply a reference to a governing Specification:

- `none` — no Specification reference exists for this workflow. Any target it operates on (if any) is not a Specification the CLI validates.
- `optional` — a Specification reference may be supplied. When one is, `requires_approved_specification` applies to it. When none is, no Specification-related check applies at all.
- `required` — a Specification reference must be supplied, and the CLI must confirm it currently exists before the Agent may start.

### `requires_approved_specification`

**Type:** boolean.

When `true`, any governing Specification actually supplied for this invocation must currently be in the derived `Approved` lifecycle state, per `SPECIFICATION_LIFECYCLE.md`, before the Agent may start. For `specification_reference: optional`, this gate applies only at invocations where a Specification was actually supplied — it is never checked, and never blocks anything, when none was given. This field is meaningless, and always `false`, when `specification_reference` is `none`.

No generic operand-type taxonomy, no invocation-mode concept, no dependency-state field, no eligibility expression, no Agent-side reasoning requirement, no runtime/execution state, no next-action metadata, and no Agent-provider configuration are part of this contract. Each of these was evaluated against the actual current workflows and found either already handled by the three fields above, or correctly belonging to Agent-side reasoning inside the workflow itself, never to a CLI-checked pre-execution fact.

---

## Result Contract

Each workflow's frontmatter additionally carries a machine-readable **Result Contract** — the structural definition the CLI uses to validate an Agent's terminal result, without hard-coding workflow-specific schemas in CLI source and without heuristically parsing Markdown prose.

This is a machine-readable *projection* of what each workflow's own Outputs section already states in prose — it introduces no new terminal value, no new required content, and no new semantics anywhere. It only re-expresses an existing truth in a form the CLI can check deterministically, exactly as `specification_reference` and `requires_approved_specification` already re-express eligibility rules Core's prose independently states, rather than inventing new ones.

### `result.terminal_path`

**Type:** string (a dotted path into the Agent's result payload).

Identifies exactly where the workflow's own terminal value lives within the payload. For eight of the ten current workflows this is a flat top-level field (for example `outcome`); for Verification and Review — whose Outputs section defines a separate summary block, distinct from and mechanically derived from their per-item evidence/findings — this is a nested path (`summary.aggregate`). This mirrors a real structural distinction Core's own text draws, not an incidental formatting choice: both workflows literally name "Verification Summary"/"Review Summary" as their own reported section, and explicitly describe `Aggregate` as *derived from* the per-item results above it — a different relationship than the other eight workflows' `Outcome`, which is always a direct, primary field. The CLI extracts the terminal value from exactly this declared location; no separate, duplicated copy of it is required anywhere else (see `RESULT_PROTOCOL.md` for how this is used at runtime).

A nested terminal_path is fully supported, generically, by the CLI: dotted-path extraction (`TerminalValue`) already walked arbitrary nesting depth from the start; generic rendering (below) suppresses only the one true terminal leaf at whatever depth it lives, while still rendering every sibling field around it, at every level — never by hardcoding "summary" or "aggregate" as known names. An earlier revision of this migration flattened Verification and Review's terminal_path to avoid extending the renderer for this case; that flattening was reverted; the renderer was extended instead, and Verification/Review's terminal_path is nested exactly as Core's own document structure calls for.

### `result.classification`

**Type:** a map from every value `result.schema`'s terminal_path property permits to exactly one of two strings: `success` or `blocked`.

This is CLI/runtime handling of an already-valid terminal result — never a reinterpretation of the workflow's own conclusion. A workflow reporting a valid negative conclusion (Verification's `FAIL`, Review's `DEFECT`) is a **successfully completed run** that classifies as `blocked` for presentation purposes, not a process or Agent failure; `failed` is never a classification value; that outcome is reserved entirely for actual CLI/process/protocol failure (an Agent crash, an invalid or missing result), computed outside this map and never confused with it.

Classification is **exhaustive by construction**: every value the terminal_path property's schema `enum` permits must have exactly one entry here, every entry must reference an actual enum value, and no third classification value beyond `success`/`blocked` is recognized. A workflow with a missing, incomplete (an enum value with no entry), unknown (an entry referencing a value the schema doesn't permit), or inconsistent (a value that isn't `success`/`blocked`) classification fails Contract validation at load time — never guessed or defaulted to `success` at runtime.

### `result.schema`

**Type:** a JSON Schema document, restricted in practice to its basic vocabulary — `type`, `enum`, `required`, `properties`, `items`.

No conditional (`if`/`then`), pattern, format, or cross-field validation is used. Conditional field presence (for example, a field populated only when the terminal value is `BLOCKED`) is expressed by making that field optional/nullable in the schema, never by a conditional-requiredness rule — enforcing *when* a field must be populated remains Core prose's responsibility for the Agent to follow, not a structural rule the CLI enforces. `additionalProperties` is left permissive by default, extending this contract's existing forward-compatibility principle for unknown fields to the Result Contract as well.

Standard JSON Schema is used deliberately rather than a bespoke format: every current workflow's Outputs section requires only objects, arrays of repeating records (Verification's obligations, Review's findings), enums, required fields, and nested objects — expressiveness that maps directly onto JSON Schema's basic vocabulary — and using the existing standard gives any future CLI implementation off-the-shelf validation tooling rather than a hand-maintained, Gnomon-specific validator.

A schema's declared `properties` order is meaningful, not incidental, at every nesting level: generic rendering (CLI Step 12) presents a payload's fields in exactly the order their own author declared them — including a nested object's own sibling fields, rendered in their own declared order once the object containing the terminal value is descended into — so that no CLI code ever needs to hardcode a specific field name to control presentation.

### Illustrative shape (syntax not final)

```yaml
---
identity: implementation
specification_reference: required
requires_approved_specification: true
result:
  terminal_path: outcome
  classification:
    IMPLEMENTATION_COMPLETE: success
    BLOCKED: blocked
  schema:
    type: object
    required: [outcome]
    properties:
      outcome:
        type: string
        enum: [IMPLEMENTATION_COMPLETE, BLOCKED]
      delivered: { type: [string, "null"] }
      verification_evidence: { type: [string, "null"] }
      remaining_unresolved: { type: [string, "null"] }
---
```

```yaml
# Verification's real, migrated shape — a nested terminal_path, mirroring its Outputs section's
# own separate "Verification Summary" block, and a repeating record array of raw per-item evidence
result:
  terminal_path: summary.aggregate
  classification:
    PASS: success
    FAIL: blocked
    UNVERIFIABLE: blocked
  schema:
    type: object
    required: [summary]
    properties:
      evidence:
        type: [array, "null"]
        items:
          type: object
          required: [obligation, result]
          properties:
            obligation: { type: string }
            source: { type: [string, "null"] }
            result: { type: string, enum: [PASS, FAIL, UNVERIFIABLE] }
            evidence: { type: [string, "null"] }
      summary:
        type: object
        required: [aggregate]
        properties:
          obligations_evaluated: { type: [string, "null"] }
          aggregate:
            type: string
            enum: [PASS, FAIL, UNVERIFIABLE]
```

Exact key names (`result`, `terminal_path`, `classification`, `schema`) illustrate the architectural relationship; the field names inside a given workflow's own `schema.properties` are not finalized syntax beyond what that workflow's own Outputs section already states.

### Terminal Result Path Per Workflow

Derived directly from each workflow's own current Outputs section, not reinterpreted, and confirmed against the real migrated frontmatter (CLI Step 12 Slice 1):

| Workflow | `terminal_path` | Terminal enum (from Core's own text) | Classification |
|---|---|---|---|
| Initial Knowledge Establishment | `outcome` | `READY_FOR_BOOTSTRAP \| BLOCKED` | success \| blocked |
| Bootstrap | `outcome` | `BOOTSTRAP_COMPLETE \| BLOCKED` | success \| blocked |
| Specification Discovery | `outcome` | `CANDIDATE_PROPOSED \| NO_CANDIDATE_IDENTIFIED \| BLOCKED` | success \| success \| blocked |
| Specification Definition | `outcome` | `READY_FOR_APPROVAL \| BLOCKED` | success \| blocked |
| Knowledge Resolution | `outcome` | `RESOLVED \| BLOCKED` | success \| blocked |
| Implementation | `outcome` | `IMPLEMENTATION_COMPLETE \| BLOCKED` | success \| blocked |
| Testing | `outcome` | `TESTING_COMPLETE \| BLOCKED` | success \| blocked |
| Verification | `summary.aggregate` | `PASS \| FAIL \| UNVERIFIABLE` | success \| blocked \| blocked |
| Review | `summary.aggregate` | `DEFECT \| RISK \| KNOWLEDGE GAP \| PASS` | blocked \| blocked \| blocked \| success |
| Git Finalization | `outcome` | `COMMIT_PREPARED \| PUBLISHED \| BLOCKED` | success \| success \| blocked |

Verification and Review are the only two workflows whose terminal value is nested under a summary block rather than sitting at the payload's top level — a direct, unmodified consequence of how their Outputs sections are already structured (see `result.terminal_path` above), not a new distinction introduced here. `NO_CANDIDATE_IDENTIFIED` and `COMMIT_PREPARED` both classify as `success` despite not being the "fullest" possible outcome (a proposed candidate, a published change) — both are legitimate, non-blocked terminal states their own Outputs prose describes as ordinary, not partial failures.

---

## Representation

The authoritative machine-readable representation is **YAML frontmatter embedded directly in each workflow's own Markdown file** — never a separate contract file, and never a central manifest. This applies identically to the eligibility fields and the Result Contract; both live in the one block shown in full under Result Contract above.

```yaml
---
identity: implementation
specification_reference: required
requires_approved_specification: true
result: { … see Result Contract above … }
---
# Implementation Workflow
...
```

The same file serves both consumers:

- The **CLI** reads the structured frontmatter to decide whether to start the Agent.
- The **Agent** reads the full Markdown — frontmatter and prose alike — as its instructions.

### Source-of-truth rule

Structured frontmatter is authoritative for every CLI-consumed pre-execution fact. Workflow prose (Purpose, When to Use, Inputs, Rules, Failure Handling, and the Outputs section itself) remains authoritative for Agent reasoning and execution guidance. Neither may independently redefine a contradictory version of the same machine-consumed fact — prose may explain or gloss a declared field (for example, restating in plain language that a governing Specification must be Approved), but it must never assert that fact as if it were its own separate source, since doing so would create two truths that could drift apart. This mirrors the same discipline already proven by the existing per-workflow `Outcome:` result block used by every workflow today.

This applies equally to the Result Contract: `result.terminal_path`, `result.classification`, and `result.schema` are authoritative for what the CLI structurally validates in an Agent's payload and how it presents a valid result; the Outputs prose remains authoritative for what the Agent must actually determine and report, and for *when* an optional field should be populated. Neither may state a different set of terminal values or required content than the other — a contradiction here is a defect to be caught by Human review when either is edited (see Result Contract above), not something CLI attempts to resolve at runtime.

---

## Workflow Mapping

The following mapping was derived directly from each workflow's current Core text (re-verified against the current files before this document was written) and is recorded here as the standalone Step 1 record, not reopened or reinterpreted:

| Workflow | `identity` | `specification_reference` | `requires_approved_specification` |
|---|---|---|---|
| Initial Knowledge Establishment | `initial-knowledge-establishment` | `none` | `false` |
| Bootstrap | `bootstrap` | `none` | `false` |
| Specification Discovery | `specification-discovery` | `none` | `false` |
| Specification Definition | `specification-definition` | `required` | `false` |
| Knowledge Resolution | `knowledge-resolution` | `none` | `false` |
| Implementation | `implementation` | `required` | `true` |
| Testing | `testing` | `optional` | `true` |
| Verification | `verification` | `none` | `false` |
| Review | `review` | `none` | `false` |
| Git Finalization | `git-finalization` | `none` | `false` |

Notes on the non-obvious rows, restated from the finalized analysis, not re-derived:

- **Specification Definition** requires a governing Specification to exist (`specification_reference: required`) but never gates on its lifecycle state — Definition operates on Draft content by design, and editing Approved content simply reverts it to Draft under `SPECIFICATION_LIFECYCLE.md`'s own deterministic rule, so no separate CLI-level approval gate is needed or correct here.
- **Testing** is the sole workflow with a conditional Specification reference: SPEC-governed testing requires the target to exist and be Approved; ad-hoc testing (defect reproduction, coverage-gap closure, rerun) requires no reference and is gated on nothing. This single field, plus its own definition of "optional," fully expresses both modes — no separate invocation-mode concept was introduced for this.
- **Verification** and **Review** operate on a deliberately untyped target that is not necessarily a Specification at all, and neither ever gates on a Specification's lifecycle state — `specification_reference: none` correctly reflects that no CLI-checked Specification concept applies to either, independent of what their target happens to be in a given invocation.
- **Git Finalization** has a real, Core-stated hard rule (work governed by a Draft Specification is ineligible for finalization), but no deterministic mapping exists from an arbitrary change/diff to the Specification(s) that govern it. This rule therefore remains entirely Agent-side reasoning (Steps 1–2 of the workflow) and is correctly represented here as `specification_reference: none` at the CLI level — the absence of a CLI-checkable field does not weaken the rule itself, which Core still states as absolute.

**Manual Draft Specification creation is explicitly not an Agent workflow** and therefore receives no entry in this table and no frontmatter contract. It is a deterministic CLI operation (input: a title; output: a new Draft Specification at the correctly computed next identity), designed and documented separately from the Agent-workflow contract model.

---

## Validation

Validation is structural and deterministic only. It never performs, duplicates, or approximates any workflow's own semantic reasoning (materiality, readiness, dependency evaluation, and so on).

| Condition | Classification |
|---|---|
| Missing `identity` | Workflow-level incompatibility — this file cannot be routed to; other workflows are unaffected. |
| Duplicate `identity` across two files | Workflow-level incompatibility, scoped to both files — the CLI refuses to route to either until resolved. |
| Invalid `specification_reference` value (outside `none \| optional \| required`) | Workflow-level incompatibility. |
| `requires_approved_specification: true` while `specification_reference: none` | Workflow-level incompatibility — an internally contradictory declaration, since there is no reference to gate. |
| Malformed frontmatter (unparseable YAML) | Workflow-level incompatibility — the file cannot be read structurally at all. |
| Unknown/extra fields | Safe to ignore — forward-compatible by design, never rejected. |
| Missing contract metadata entirely (no frontmatter block) | Safe to ignore for discovery purposes; if a Human explicitly names the file for CLI-orchestrated invocation, a clear workflow-level incompatibility is reported instead. |
| Built-in `identity` present but frontmatter differs from its bundled default | Warning only — reclassify from recognized built-in to project-customized workflow; the file still runs. |
| Unsupported global contract version | Fatal Core incompatibility — see Global Contract Version below; the CLI refuses to proceed with any workflow rather than guess at an unfamiliar shape. |
| Missing `result.terminal_path` or `result.schema` | Workflow-level incompatibility — this workflow's results can never be structurally validated until it is added. |
| `result.schema` is not valid JSON Schema | Workflow-level incompatibility. |
| Missing `result.classification` | Workflow-level incompatibility — no terminal value could ever be classified without guessing. |
| A value the terminal_path property's `enum` permits has no `result.classification` entry (incomplete) | Workflow-level incompatibility. |
| `result.classification` references a value the schema's `enum` does not permit (unknown) | Workflow-level incompatibility. |
| A `result.classification` entry's value is neither `success` nor `blocked` (inconsistent) | Workflow-level incompatibility. |

Validating an actual Agent-produced payload against a workflow's `result.schema` at runtime — including what happens on a schema violation, a missing terminal value, or a malformed payload — is a per-invocation event, not a static frontmatter-validation concern, and is defined by CLI Step 5 (`RESULT_PROTOCOL.md`), not here.

---

## Global Contract Version

Workflow Contract v1 uses **one global contract/schema version**, not a version per workflow. No workflow's individual metadata shape has been found to need independent versioning from any other's — every field discovered across all ten workflows is already the same shared shape.

The current contract version is **`1`**.

This versions the CLI/Core machine contract itself — the shape and meaning of `identity`, `specification_reference`, `requires_approved_specification`, and the Result Contract (`result.terminal_path`, `result.classification`, `result.schema`) — and nothing else. It is not the version of any individual Specification's content, and not a version of any individual workflow's prose or behavior; those are governed entirely by their own Core semantics (lifecycle, readiness, and so on), independent of this number.

Per the finalized Step 3 model, the project-local record of this version lives at:

```
.gnomon/CONTRACT_VERSION
```

No second versioning mechanism is introduced. See `PROJECT_INITIALIZATION.md` for how this file is materialized, validated, and used to detect incompatibility.

---

## Deferred to Future Work

The following are intentionally not defined here, and nothing in this document depends on them:

- Agent adapter architecture and Agent provider configuration.
- The Result Protocol's transport, envelope, and runtime consumption flow — CLI Step 5 (`RESULT_PROTOCOL.md`), which consumes the Result Contract documented here but is not itself defined here.
- Any eligibility expression language, boolean combinator, or generic rules engine — every case audited across all ten current workflows reduced to the three eligibility fields above, plus the Result Contract, with no further exception found.
- Dependency-state metadata — confirmed, across every workflow that has a dependency concept, to be a partial-work condition evaluated by the Agent during execution, never a CLI start gate.
- Next-action metadata — already expressed in prose today (for example, Review's `Recommended Next Workflow` field) and not duplicated here.
- Runtime/execution state, persistence, or a workflow state machine.
- Custom-workflow discovery mechanics (though custom workflows are already confirmed able to use this exact contract with no CLI code changes).

---

## Notes

This document persists the Step 1 decisions already finalized in prior design analysis, including the Result Contract addition resolved during the CLI Step 5 challenge process — a machine-readable projection of each workflow's existing Outputs prose, not a new semantic rule. It does not reopen, reinterpret, or expand that analysis — the field set, the workflow mapping, the Result Contract, and the validation model above are a direct record of those conclusions, checked against the current workflow files for terminology consistency only.

Physically adding this frontmatter — the three eligibility fields and the Result Contract alike — to the ten actual `workflows/*.md` files remains a separate, explicit, not-yet-performed migration task. As of this document, none of the ten workflow files carry any frontmatter at all; this document is the design record of what that frontmatter will contain once that migration is separately authorized.
