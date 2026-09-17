# Result Protocol — CLI Step 5

## Purpose

Define how the CLI obtains one deterministic, machine-readable workflow result from an Agent during a fully interactive Terminal Handoff run — without parsing the interactive terminal conversation, without introducing persisted workflow-execution state, and without duplicating any Core semantic rule.

The transient result artifact this document defines serves two roles at once: it is the workflow result transport (its original purpose), and it is also the out-of-band signal Gnomon uses to detect that the run is complete and end the Agent's session itself — the Agent is never asked to exit on its own, and is never run headless to force a deterministic exit.

This is CLI-level design, not Core semantics. It defines transport, timing, and structural validation; it does not define what any workflow's result means — that is Core's Outputs section and the Result Contract CLI Step 1 documents.

---

## Scope

This document covers:

- The logical result representation and the transient transport CLI v1 uses by default.
- Invocation identity (`run_id`) and how a stale or misrouted result is prevented from ever being consumed.
- Where the transient result artifact lives, its gitignore ownership, and its ephemeral lifecycle.
- The exact CLI validation sequence, including how it consumes the Result Contract CLI Step 1 defines, without hard-coding any workflow-specific schema.
- The separation between process/runtime status, protocol validity, and workflow result.
- The boundary excluding post-run repository-drift/eligibility re-checking from this Step.

This document does not cover:

- What any workflow's result means, which values are valid, or what its payload must contain — defined entirely by Core's own Outputs sections and projected, without adding meaning, into the Result Contract CLI Step 1 (`WORKFLOW_CONTRACT.md`) documents.
- Workflow Contract v1's eligibility fields or Result Contract structure themselves — defined by CLI Step 1.
- The Bundled Default Core / Project-local Core distribution model — CLI Step 2 (`CORE_DISTRIBUTION.md`).
- `.gnomon/` materialization and `gnomon init` — CLI Step 3 (`PROJECT_INITIALIZATION.md`).
- Agent startup, Terminal Handoff, and the Adapter's `prepare/run/cancel/status` interface — CLI Step 4 (`AGENT_ADAPTER.md`), whose reserved result-delivery instruction slot this document fills in.
- General CLI orchestration policy — including whether/how previously-known deterministic facts (target existence, approval state) are re-checked after a result arrives, and what happens if they no longer hold. That is a later CLI orchestration concern, deliberately not part of Result Protocol.

---

## Preconditions Carried Forward (Steps 1–4, unchanged)

- **Workflow Contract v1** (Step 1): every Agent workflow declares `identity`, `specification_reference`, `requires_approved_specification`, and a **Result Contract** (`result.terminal_path`, `result.schema`) in frontmatter.
- **Core Distribution** (Step 2): the project-local Core is the sole runtime authority during any run.
- **Project Initialization** (Step 3): `.gnomon/` is the materialized Core root, discovered by upward search.
- **Agent Adapter** (Step 4): Terminal Handoff is the default, and only, v1 interaction strategy — always fully interactive, never headless; the CLI never observes the Human↔Agent conversation; the Adapter exposes `prepare(context)`, `run()`, `cancel()`, `status()`, with a reserved context slot for a result-delivery instruction and an optional completion-check function `run()` polls while the Agent is still running.

Nothing in this document changes any of these.

---

## Logical Result Representation

The Agent returns a JSON structure with exactly three top-level fields:

```json
{
  "workflow": "implementation",
  "run_id": "…",
  "payload": { }
}
```

- **`workflow`** — the invoked workflow's own `identity` (Step 1), echoed by the Agent, letting the CLI detect misrouting with a plain string comparison rather than parsing a prose title.
- **`run_id`** — the unique identifier the CLI generated for this invocation (below), echoed by the Agent, letting the CLI detect a stale or path-collided result.
- **`payload`** — the workflow's own Outputs content, re-encoded as JSON per the Result Contract's `schema`, unmodified in meaning from Core's own text.

**There is no separate `status` field.** An earlier design duplicated the terminal value at the envelope level specifically because the CLI had no way to locate it inside an opaque payload. Now that each workflow's Result Contract declares `terminal_path`, that compensation is redundant: the CLI extracts the terminal value directly from `payload` at the declared path, and there is exactly one place that value exists — removing a place it could ever disagree with itself.

---

## Transport

**Default v1 transport is an invocation-specific transient file**, written once by the Agent via its own ordinary filesystem access — never via stdout/stderr, and never observed live. This is compatible with Terminal Handoff by construction: the write happens through the same filesystem access the Agent already uses to modify the repository, entirely independent of whatever the Human sees rendered.

The **authoritative level is the logical structure above, not the literal file**. A future SDK-backed Adapter may instead hand the CLI the equivalent in-memory structure directly, without ever touching a file, and remain fully compliant — provider neutrality is real at this level, not merely rhetorical, because the Adapter/CLI boundary is defined at the logical shape, and a literal file is CLI v1's only currently-real realization of it.

---

## Invocation Identity

The CLI generates a unique `run_id` before starting the Adapter, and derives the transient file's destination path from it. This is purely a runtime transport concern — never Core semantics, never persisted beyond the run, never stored anywhere durable.

The identifier is checked in two places for two different reasons: the **destination path** being invocation-unique is what prevents two runs from ever colliding at the filesystem level; the **payload-echoed `run_id`** is a cheap, already-available cross-check against a distinct, believable failure mode — a path-generation or cleanup bug causing the same path to be reused across invocations — which path-uniqueness alone cannot self-detect, since it assumes its own correctness. Both checks are retained; neither is redundant with the other.

---

## Transient Artifact Location

The result file lives in a **project-local directory, sibling to `.gnomon/`, never nested inside it** — created lazily by the CLI's pre-run orchestration immediately before the first `run()`, not by `gnomon init`. Nesting it inside `.gnomon/` would force Step 3's currently complete, closed canonical layout to carry a permanent exception for a runtime-only subtree; a sibling directory needs none, leaving Step 3 entirely unmodified.

The directory carries its **own, self-contained gitignore rule**, touching no other `.gitignore` anywhere in the project — not the project root's, not one inside `.gnomon/`. This satisfies Step 3's "no unnecessary `.gitignore` mutation" principle by not mutating any existing file at all.

*(Step 3's own canonical-layout documentation gains a small, additive cross-reference to this sibling directory when this Step is next revisited — not a change to what `.gnomon/` means or what `gnomon init` does.)*

---

## Persistence Policy

**Fully ephemeral.** The CLI deletes the result file immediately after a successful read. A failed or malformed read may leave the file in place briefly for diagnostics, swept unconditionally at the *next* invocation's own startup. Nothing accumulates into a log or history. This is deliberately distinct from Human Approval evidence — Core's own reserved, durable, historical-fact mechanism — and an ordinary workflow result must never become a second, competing persisted-state system.

---

## Agent Publication Instruction

The Step 4 reserved context slot tells the Agent: the exact destination path; the `run_id` to echo; its own `identity` to echo as `workflow`; and to serialize its own workflow's Outputs content into `payload`, conforming to that workflow's own Result Contract (`result.schema`) as declared in the same file it is already reading for its instructions. The CLI never copies the schema into the prompt separately — the Agent reads it from the one authoritative location, exactly as it reads every other instruction for the workflow it is following. The Agent publishes exactly once, before exiting.

---

## Consumption Timing

The CLI checks for the result **while the Agent may still be running**, not only after `run()` returns — this is what lets Gnomon detect completion and end the session itself, without requiring the Agent to exit on its own or running it headless. The check is a periodic, out-of-band poll of the transient file's path, at a short, fixed interval, entirely separate from the interactive terminal stream, which is never read or parsed. Every poll before the real one is expected to fail — file absent, or present but not yet valid — and is treated identically to "not ready yet," never as an error; only the poll that finally succeeds matters. The same validation logic (below) governs every attempt, whether made mid-run or, as a final fallback, once the process has exited on its own with the poll never having succeeded.

This supersedes an earlier version of this document, which concluded no filesystem-watching mechanism was needed and consumption happened only post-run — that conclusion assumed the Agent's own process would exit on its own once finished, which proved unreliable in practice for an interactive session (see `AGENT_ADAPTER.md`'s Notes).

---

## Validation Flow

The sequence below runs on every poll while the Agent is still running, and once more after it exits if no poll ever succeeded — the same logic either way, the only difference being that a failure is silently retried mid-run and only reported once the process has actually ended with nothing valid found:

```text
CLI reads the invocation-specific result file (mid-run poll, or post-exit)
   ↓
parse JSON                                → fail: not ready yet (mid-run) / protocol failure (final)
   ↓
workflow == invoked identity              → fail: not ready yet (mid-run) / protocol failure (final) — identity mismatch
run_id == CLI-generated id                → fail: not ready yet (mid-run) / protocol failure (final) — stale/mismatched run
   ↓
load the already-read invoked workflow's own Result Contract
(result.terminal_path + result.schema, from its frontmatter — Step 1)
   ↓
validate payload against result.schema   → fail: protocol failure (result contract violation, with the specific field/value reported)
   ↓
extract terminal value at terminal_path  → this is the reported workflow result
   ↓
report: protocol success; process status (Step 4); workflow result — three separate facts
```

**No workflow-specific schema is hard-coded anywhere in CLI source.** The CLI's logic is always "read the Result Contract declared by *this* workflow's own file, validate against it" — never a per-workflow branch. The schema is loaded from the same file already read once before starting the Agent for Step 1's eligibility check; no new file access or new trust boundary is introduced.

---

## Malformed/Missing Result Behavior

These classifications apply to the **final** state — reached once the Agent process has exited with no valid result ever having appeared, or when a single post-run check is the only check performed at all (no mid-run polling). A transient invalid read *during* polling, while the Agent is still running, is never one of these; it is simply "not ready yet" (see Consumption Timing). Every case below is a **protocol failure**, reported distinctly from any Core `Outcome` — never mapped onto `BLOCKED` or any other workflow-defined value:

- No result file produced, empty file, unparseable JSON.
- `workflow` does not match the invoked identity.
- `run_id` does not match the current invocation.
- `payload` fails validation against the workflow's own `result.schema` (unknown terminal value, missing required field, malformed array/object).

A result that is structurally valid is never rejected merely because the process's own exit code was non-zero, and a clean exit is never treated as implying a successful workflow result — process status and protocol status are independent facts, checked and reported separately (see below).

---

## Process Status, Protocol Status, and Workflow Result

Three independent facts, never conflated:

- **Process/runtime status** — from the Adapter's `status()` (Step 4): clean exit, crash, cancellation, forced termination.
- **Protocol status** — whether a structurally valid, correctly-identified, contract-conformant result arrived at all (this document).
- **Workflow result** — the Agent's own faithfully-extracted terminal value and payload, semantically owned entirely by Core.

A clean process exit does not imply protocol success. Protocol success does not imply the workflow succeeded (`BLOCKED` is a perfectly valid protocol success). A non-zero exit does not invalidate an otherwise contract-valid workflow result — all three are reported together, distinctly, never merged.

---

## Crash and Interruption Behavior

No partial result ever counts — anything short of a fully valid, contract-conformant result fails validation outright, with no separate "partial credit" tier. On any crash or missing result, the CLI reports a protocol failure; the repository remains the source of truth exactly as Step 4 established, and the next invocation of any relevant workflow simply re-inspects current repository state fresh — never resuming from any assumed in-progress status, since none is ever recorded.

---

## Atomicity

**Still no write-temp-then-rename, no completion marker, no file locking** — but the reasoning has changed along with Consumption Timing above. A genuine concurrent read/write window now exists: the CLI may poll the destination path at the exact moment the Agent is partway through writing it. This is handled by tolerance, not by a locking or atomic-rename mechanism: a poll that finds the file absent, empty, or unparseable is indistinguishable from — and treated identically to — "not ready yet," and is simply retried on the next tick. Nothing about a transient parse failure during polling is ever reported as an error; only a final, still-invalid state — reached once the Agent process has exited with no valid result ever having appeared — is a genuine malformed/missing-result failure (below). A provider/Adapter whose Agent process exits while a detached subprocess is still writing on its behalf remains a Step 4 Adapter-compatibility concern, not something this protocol compensates for.

---

## Concurrency

CLI v1 does not support concurrent workflow execution within one project — no background-Agent concept exists (Step 4). Invocation-unique result paths and the echoed `run_id` remain valuable independent of concurrency, specifically to prevent a *stale, already-finished* prior run's result from ever being mistaken for the current one.

---

## Repository Drift and Post-Run Eligibility

**Out of scope for this document, deliberately.** Whether the CLI re-checks previously-known deterministic facts (target existence, approval state) after a valid result arrives, and what consequence any drift has, belongs to general CLI orchestration — not to Result Protocol. This document's only commitment on the point: the Agent's reported result is never silently rewritten, downgraded, or discarded by this protocol to make it agree with current repository facts, and it is never merged with any later drift finding. A later orchestration layer may report drift as separate, clearly-labeled information alongside the result — never inside or in place of it.

---

## Diagnostics

The smallest useful set, all already available from mechanisms already defined: process/runtime status (Step 4's `status()`); the checked result-path location; the specific field or value that failed contract validation, when applicable; any provider/runtime error `status()` already surfaces. No persisted full Agent transcript — there is nothing to persist, since Terminal Handoff never captured the conversation in the first place.

---

## Stress-Test Examples

**Implementation** — `terminal_path: outcome`. `{"payload": {"outcome": "BANANA"}}` fails validation (not in the workflow's declared enum). `{"payload": {}}` fails validation (`outcome` missing, a required field).

**Specification Definition** — `terminal_path: outcome`, with `remaining_unresolved` validated as an array of `{gap, owner, decision_pending}` objects — array item boundaries are syntactic, never inferred from prose.

**Verification** — `terminal_path: summary.aggregate`; `obligations` validated as an array of `{obligation, source, result: enum[PASS,FAIL,UNVERIFIABLE], evidence}` objects.

**Review** — `terminal_path: summary.aggregate`; `findings` validated as an array of `{finding_id, classification: enum[...], decision_required, ...}` objects.

Every one of the ten current workflows is representable by this model with no exception.

---

## Structural vs. Semantic Validation

The CLI can say "this result conforms to the workflow's Result Contract" — valid enum, required fields present, correct structure. The CLI cannot and does not say "the Agent was correct to conclude `IMPLEMENTATION_COMPLETE`" — that remains, entirely and exclusively, Agent reasoning governed by Core. Structural validation catches nonsense (`BANANA`, an empty payload); it never adjudicates whether the Agent's semantic conclusion was actually right.

---

## Deferred to Future Work

- General CLI orchestration policy for post-run deterministic re-checking and its consequences.
- Approval-evidence persistence — defined by CLI Step 6 (`APPROVAL_RUNTIME.md`).
- Exact frontmatter key names for the Result Contract (`WORKFLOW_CONTRACT.md`'s illustration is not final syntax).
- Any narrow, best-effort automated check between a workflow's prose-stated terminal enum and its declared schema — a possible future linting nicety, not a v1 requirement; beyond the single-line terminal-field declaration, prose/schema consistency remains a Human review responsibility.
- CLI command hierarchy, interactive Gnomon UX, and everything Step 4 already deferred.

---

## Notes

This document persists the Step 5 architecture finalized through analysis and challenge in prior design discussion, including the resolution that a machine-readable Result Contract — documented in `WORKFLOW_CONTRACT.md`, not here — is necessary for deterministic payload validation, and the consequent removal of the envelope's redundant `status` field. It does not reopen Steps 1–4. Physical application of the Result Contract to the ten actual `workflows/*.md` files is a separate, not-yet-performed migration task, exactly as noted in `WORKFLOW_CONTRACT.md`.

A later revision corrected Consumption Timing and Atomicity, which had concluded no filesystem-watching mechanism was needed. In practice, an interactive Agent session does not reliably end its own process on completion, and running the Agent headless to force a deterministic exit broke both interactive Human↔Agent use and, separately, the Agent's own ability to use tools without a Human present to answer permission prompts. The transient result file already described here turned out to be exactly what was needed to solve this without either cost: Gnomon polls it while the Agent stays fully interactive, and terminates the session itself once a valid result appears. The envelope, schema, transport location, and validation logic are all unchanged — only when they are consulted changed, from "once, after exit" to "repeatedly while running, tolerating not-yet-ready reads, with one final check after exit as a fallback."
