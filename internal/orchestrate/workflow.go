package orchestrate

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gnomon/internal/adapter"
	"gnomon/internal/contract"
	"gnomon/internal/facts"
	"gnomon/internal/present"
	"gnomon/internal/project"
	"gnomon/internal/result"
)

// targetKind distinguishes how a command's optional target argument relates to the workflow it
// invokes — invocation-shape configuration fixed per command (the same category of fact a
// command's own argument arity already is: implement's <SPEC-id> is required, verify's [target]
// is optional and untyped, describe takes no argument at all), never a switch over a workflow's
// own runtime result. Result interpretation (classification, rendering) remains entirely
// Contract-driven regardless of which kind a given invocation used.
//
// Verification and Review deliberately take a generic, untyped target per their own Workflow
// Contract (specification_reference: none) — cli/WORKFLOW_CONTRACT.md is explicit that this is
// "not necessarily a Specification at all," so it is never checked for existence or lifecycle the
// way a real Specification identity is, and the Agent is never told it is one (adapter.Context.Target
// vs. SpecIdentity; see internal/adapter/process.go's buildPrompt).
type targetKind int

const (
	targetNone    targetKind = iota // no target argument (describe, bootstrap, finalize)
	targetSpec                      // a Specification identity — existence/lifecycle-checked (implement)
	targetGeneric                   // a free-form, untyped target — never checked (verify, review)
)

// Implement performs `gnomon implement <SPEC-id>`. Pre-start eligibility is checked first,
// deterministically, with no Agent provider resolved — and therefore no first-use "choose your
// Agent" prompt ever shown — if it fails. This ordering is what keeps an ineligible invocation
// (a Draft Specification, say) from provoking a provider prompt for a run that was always going
// to be refused. Only once eligible does ResolveAgent (agent.go) run, exactly once, to obtain the
// Adapter this run will use; orchestration never constructs a ClaudeAdapter or CodexAdapter
// itself.
func Implement(root, specID, agentOverride string, chooser AgentChooser) (*present.Report, error) {
	return implementWithHandoff(root, specID, ResolutionHandoff{}, agentOverride, chooser)
}

// implementWithHandoff is Implement's sibling for the interactive finding-resolution loop, which
// is the only caller that ever supplies a non-empty ResolutionHandoff (cmd/gnomon, via
// RunResolutionWorkflow). Implement's own eligibility logic — in particular the Approved-
// Specification gate — is unchanged and unduplicated: it lives here exactly once, and Implement
// is now a one-line delegator to this with a zero-value handoff, so every existing caller and
// test is unaffected.
func implementWithHandoff(root, specID string, handoff ResolutionHandoff, agentOverride string, chooser AgentChooser) (*present.Report, error) {
	l, wf, elig, err := prepareImplementation(root, specID)
	if err != nil {
		return nil, err
	}
	if !elig.Eligible {
		return &present.Report{
			Outcome: present.Blocked,
			Summary: "Implementation blocked",
			Target:  specID,
			Sections: []present.Section{
				{Label: "Unresolved", Body: elig.Reason},
			},
			Next: fmt.Sprintf("Define %s further, or `gnomon approve %s` if you judge it ready.", specID, specID),
		}, fmt.Errorf("ineligible: %s", elig.Reason)
	}

	workflowPath := filepath.Join(l.WorkflowsDir(), "implementation.md")
	return runResolutionAgent(root, l, wf, workflowPath, targetSpec, specID, handoff, agentOverride, chooser)
}

// Describe performs `gnomon describe`, wired directly to Initial Knowledge Establishment
// (workflows/initial-knowledge-establishment.md) — Core's own workflow identity and the sole
// authority for this command's responsibility, inputs, eligibility, and Result semantics;
// "describe" is only the Human-facing CLI verb (cli/COMMAND_SURFACE.md), never a separate
// workflow. It takes no target: the Contract declares specification_reference: none, and Core's
// text never names any Specification this workflow depends on.
//
// describe hands control to the Agent with no CLI-collected input beyond the ordinary invocation
// context: the workflow itself is responsible for inspecting available evidence and asking the
// Human directly, inside the same interactive session, when material project knowledge cannot be
// established from it — see Initial Knowledge Establishment's own Execution steps. The CLI does
// not attempt to classify a project as "new" or "existing" beforehand; that judgment belongs to
// the Agent, which can actually inspect the repository.
func Describe(root, agentOverride string, chooser AgentChooser) (*present.Report, error) {
	rep, err := runNoTargetWorkflow(root, "initial-knowledge-establishment.md", agentOverride, chooser)
	onboardingNext(rep, "gnomon bootstrap")
	return rep, err
}

// Bootstrap performs `gnomon bootstrap`, wired directly to workflows/bootstrap.md — a distinct
// Core workflow from Initial Knowledge Establishment, never merged with it: Bootstrap's own
// prose assumes Initial Knowledge Establishment's baseline already exists, and neither one's
// responsibility is folded into the other here.
func Bootstrap(root, agentOverride string, chooser AgentChooser) (*present.Report, error) {
	rep, err := runNoTargetWorkflow(root, "bootstrap.md", agentOverride, chooser)
	onboardingNext(rep, `gnomon spec create "<title>", or gnomon spec discover to have an Agent propose one`)
	return rep, err
}

// onboardingNext sets Next to onSuccess only on a Success Outcome. Describe and Bootstrap's
// Result Contracts each declare exactly one success terminal value, so Success here is never
// ambiguous about which one occurred.
func onboardingNext(rep *present.Report, onSuccess string) {
	if rep != nil && rep.Outcome == present.Success {
		rep.Next = onSuccess
	}
}

// Finalize performs `gnomon finalize`, wired directly to workflows/git-finalization.md. It takes
// no target and enforces no CLI-level authorization gate beyond the ordinary eligibility check:
// Git Finalization's one hard rule (work governed by a Draft Specification is ineligible to
// finalize) has no deterministic mapping from an arbitrary change to the Specification(s) that
// govern it — cli/WORKFLOW_CONTRACT.md already documents this as entirely Agent-side reasoning,
// correctly reflected by the workflow's own specification_reference: none, not a gap this command
// needs to fill by inventing state. Gnomon itself never commits, pushes, or publishes anything —
// every one of those actions, and whether it is authorized, is decided entirely by the Agent
// following the workflow's own Rules; the CLI's role here is exactly what it is for every other
// Agent-invoking command: launch, watch for a valid result, terminate, report.
func Finalize(root, agentOverride string, chooser AgentChooser) (*present.Report, error) {
	return runNoTargetWorkflow(root, "git-finalization.md", agentOverride, chooser)
}

// Verify performs `gnomon verify [target]`, wired directly to workflows/verification.md. target
// is deliberately generic/untyped per the workflow's own Contract (specification_reference:
// none) — never checked for existence or lifecycle, and never described to the Agent as a
// Specification (see targetKind above).
func Verify(root, target, agentOverride string, chooser AgentChooser) (*present.Report, error) {
	return runGenericTargetWorkflow(root, "verification.md", target, agentOverride, chooser)
}

// Review performs `gnomon review [target]`, wired directly to workflows/review.md. Same
// generic/untyped target treatment as Verify, for the same Core-stated reason.
func Review(root, target, agentOverride string, chooser AgentChooser) (*present.Report, error) {
	return runGenericTargetWorkflow(root, "review.md", target, agentOverride, chooser)
}

// SpecDiscover performs `gnomon spec discover`, wired directly to
// workflows/specification-discovery.md — Core's own workflow identity and the sole authority for
// this command's responsibility, prerequisites, and Result semantics; "spec discover" is only the
// Human-facing CLI verb, never a second discovery model. It takes no target
// (specification_reference: none) and creates or mutates nothing itself: the workflow's own text
// is explicit that Discovery "does not modify an existing Specification" and "Create[s] nothing
// yet — creation happens only after the Human decides." Accepting a proposed candidate is a
// separate, subsequent Human action through the already-existing, unchanged `gnomon spec create
// <title>` (SpecCreate, spec.go) — the same deterministic Draft Creation mechanism Discovery's
// own "Draft Creation" section describes, not a second implementation of it.
func SpecDiscover(root, agentOverride string, chooser AgentChooser) (*present.Report, error) {
	return runNoTargetWorkflow(root, "specification-discovery.md", agentOverride, chooser)
}

// SpecDefine performs `gnomon spec define <SPEC-id>`, wired directly to
// workflows/specification-definition.md. Its target is a real Specification identity
// (specification_reference: required) that must already exist — Definition never creates one;
// facts.Eligible's existing existence-only check (requires_approved_specification: false, so
// lifecycle is never gated here) refuses cleanly otherwise, before any Agent is resolved. The
// Agent edits the existing Specification's own content directly, via the same ordinary
// filesystem access every other Agent-invoking command already relies on — no CLI mediation, no
// second content-validation step. A `READY_FOR_APPROVAL` result is exactly that: a workflow
// result, rendered like any other by the generic engine — it never grants, implies, or performs
// approval, which remains exclusively `gnomon approve`'s own, entirely separate act.
func SpecDefine(root, specID, agentOverride string, chooser AgentChooser) (*present.Report, error) {
	return runTargetedWorkflow(root, "specification-definition.md", targetSpec, specID, agentOverride, chooser)
}

// Test performs `gnomon test [SPEC-id]`, wired directly to workflows/testing.md. Its Contract
// declares specification_reference: optional with requires_approved_specification: true: when no
// Specification is supplied, facts.Eligible's existing optional-reference branch is always
// eligible (ad-hoc/exploratory testing, per the workflow's own When to Use); when one is
// supplied, it must already exist and be Approved — the same hard gate Implementation uses,
// applied identically here because the Contract, not command-specific Go code, declares it.
// Testing is deliberately kept distinct from Verification: they are separate Contract files
// invoked by separate commands, and nothing here invokes one from the other automatically.
func Test(root, specID, agentOverride string, chooser AgentChooser) (*present.Report, error) {
	return testWithHandoff(root, specID, ResolutionHandoff{}, agentOverride, chooser)
}

// testWithHandoff is Test's sibling for the interactive finding-resolution loop — the only
// caller that ever supplies a non-empty ResolutionHandoff. Test's own eligibility (facts.Eligible
// applied to testing.md's own specification_reference: optional / requires_approved_specification:
// true) is unchanged: this reproduces runTargetedWorkflow's own prepare-then-eligibility-check
// sequence exactly, just ending in runResolutionAgent instead of runWorkflow so a handoff can be
// carried when one is supplied; Test is now a one-line delegator with a zero-value handoff, so
// every existing caller and test is unaffected.
func testWithHandoff(root, specID string, handoff ResolutionHandoff, agentOverride string, chooser AgentChooser) (*present.Report, error) {
	l, wf, workflowPath, err := prepareWorkflow(root, "testing.md")
	if err != nil {
		return nil, err
	}
	rep, _, err := runEligibleWorkflowFull(root, l, wf, workflowPath, targetSpec, specID, handoff, agentOverride, chooser)
	return rep, err
}

// runNoTargetWorkflow and runGenericTargetWorkflow are two convenience shapes over
// runTargetedWorkflow for the targetNone and targetGeneric cases (describe, bootstrap, finalize;
// verify, review). SpecDefine and Test — whose target is a real Specification identity
// (targetSpec), governed by facts.Eligible exactly as Implement's is — call runTargetedWorkflow
// directly instead, with no separate convenience wrapper needed.
func runNoTargetWorkflow(root, filename, agentOverride string, chooser AgentChooser) (*present.Report, error) {
	return runTargetedWorkflow(root, filename, targetNone, "", agentOverride, chooser)
}

func runGenericTargetWorkflow(root, filename, target, agentOverride string, chooser AgentChooser) (*present.Report, error) {
	return runTargetedWorkflow(root, filename, targetGeneric, target, agentOverride, chooser)
}

// runTargetedWorkflow loads the named workflow's Contract, checks its deterministic pre-start
// eligibility via the one shared facts.Eligible (never inferring anything the Contract doesn't
// itself state), and — only if eligible — hands off to the same runWorkflow Implement uses. It is
// shared by every command in this slice, including SpecDefine and Test, whose target genuinely is
// a Specification identity. The ineligibility Report here is deliberately generic (no
// Specification-flavored "gnomon approve"/"gnomon spec create" suggestion, unlike Implement's own
// tailored message above) — a single deterministic Contract-driven message serves every caller
// uniformly, rather than special-casing wording per workflow; elig.Reason itself (e.g. "SPEC-003
// does not exist", "SPEC-003 is not Approved") already states plainly what to do next.
func runTargetedWorkflow(root, filename string, kind targetKind, target, agentOverride string, chooser AgentChooser) (*present.Report, error) {
	l, wf, workflowPath, err := prepareWorkflow(root, filename)
	if err != nil {
		return nil, err
	}
	return runEligibleWorkflow(root, l, wf, workflowPath, kind, target, agentOverride, chooser)
}

// runEligibleWorkflow checks the ordinary deterministic pre-start eligibility for an
// already-loaded workflow, and — only if eligible — hands off to runWorkflow. Shared by
// runTargetedWorkflow (every dedicated command's filename-based dispatch) and Run
// (identity-based dispatch, below), so both reduce to exactly the same eligibility-then-run
// sequence — the concrete proof that a dedicated command and `gnomon run <identity>` addressing
// the same workflow converge on identical execution semantics, not two parallel
// implementations. The ineligibility Report here is deliberately generic (no
// Specification-flavored "gnomon approve"/"gnomon spec create" suggestion, unlike Implement's own
// tailored message) — a single deterministic Contract-driven message serves every caller
// uniformly; elig.Reason itself (e.g. "SPEC-003 does not exist", "SPEC-003 is not Approved")
// already states plainly what to do next.
func runEligibleWorkflow(root string, l project.Layout, wf contract.Workflow, workflowPath string, kind targetKind, target, agentOverride string, chooser AgentChooser) (*present.Report, error) {
	rep, _, err := runEligibleWorkflowWithOutcome(root, l, wf, workflowPath, kind, target, agentOverride, chooser)
	return rep, err
}

// runEligibleWorkflowWithOutcome is runEligibleWorkflow's sibling for the one caller that also
// needs the raw, validated Outcome once execution actually reaches obtainWorkflowOutcome —
// RunEvaluation, which extracts Verification/Review's own findings from it for the interactive
// resolution loop (cmd/gnomon). Kept as an addition, not a signature change to
// runEligibleWorkflow, so every existing caller and test is unaffected; runEligibleWorkflow is
// now a one-line delegator to this, so the eligibility logic itself still exists exactly once.
func runEligibleWorkflowWithOutcome(root string, l project.Layout, wf contract.Workflow, workflowPath string, kind targetKind, target, agentOverride string, chooser AgentChooser) (*present.Report, *result.Outcome, error) {
	return runEligibleWorkflowFull(root, l, wf, workflowPath, kind, target, ResolutionHandoff{}, agentOverride, chooser)
}

// runEligibleWorkflowFull is runEligibleWorkflowWithOutcome's sibling for callers that also
// supply a ResolutionHandoff (implementWithHandoff, testWithHandoff, knowledgeResolutionWithHandoff
// — the interactive finding-resolution loop's own dispatch path). The eligibility logic itself
// exists exactly once here; runEligibleWorkflowWithOutcome is now a one-line delegator with a
// zero-value handoff, so every existing caller and test is unaffected.
func runEligibleWorkflowFull(root string, l project.Layout, wf contract.Workflow, workflowPath string, kind targetKind, target string, handoff ResolutionHandoff, agentOverride string, chooser AgentChooser) (*present.Report, *result.Outcome, error) {
	elig, err := facts.Eligible(l, wf, target)
	if err != nil {
		return nil, nil, err
	}
	if !elig.Eligible {
		return &present.Report{
			Outcome: present.Blocked,
			Summary: fmt.Sprintf("%s blocked", humanizeIdentity(wf.Identity)),
			Target:  target,
			Sections: []present.Section{
				{Label: "Unresolved", Body: elig.Reason},
			},
		}, nil, fmt.Errorf("ineligible: %s", elig.Reason)
	}

	return runWorkflowFull(root, l, wf, workflowPath, kind, target, handoff, agentOverride, chooser)
}

// Run performs `gnomon run <workflow-identity> [target]` — reaching any workflow by its own
// declared Contract identity, built-in or custom, per cli/COMMAND_SURFACE.md's "run Rule":
// Workflow Contract v1 does not distinguish built-in from custom at execution time, so no
// artificial restriction is introduced here. Unlike every dedicated command, Run does not know
// its target workflow's filename in advance — resolveWorkflowByIdentity finds it by scanning
// every workflow file's own declared identity, exactly the routing cli/WORKFLOW_CONTRACT.md's
// `identity` field exists for ("independent of the file's name or path"). targetKindFor derives
// how to treat the supplied target purely from the resolved workflow's own
// specification_reference — never a hardcoded per-workflow-name table — so a custom workflow
// receives exactly the same target treatment a built-in one with the same
// specification_reference value would. From there this reduces to the identical
// runEligibleWorkflow → runWorkflow → runWorkflowWithAdapter chain every dedicated command uses.
func Run(root, workflowIdentity, target, agentOverride string, chooser AgentChooser) (*present.Report, error) {
	rep, _, err := RunEvaluation(root, workflowIdentity, target, agentOverride, chooser)
	return rep, err
}

// RunEvaluation is Run's sibling for the one caller (cmd/gnomon's `gnomon run`) that also needs
// Verification/Review's own actionable findings, extracted from the run's validated payload via
// ActionableFindings, so it can offer the interactive resolution loop without this package
// leaking Result Contract payload shape into cmd/gnomon. findings is always nil for every
// workflow identity other than "verification"/"review", and for those two whenever the result
// itself has nothing actionable (a clean PASS) — never treated as an error either way. Run is now
// a one-line delegator to this, so every other identity's behavior through `gnomon run` is
// byte-for-byte unchanged.
func RunEvaluation(root, workflowIdentity, target, agentOverride string, chooser AgentChooser) (*present.Report, []EvaluationFinding, error) {
	l, err := project.Locate(root)
	if err != nil {
		return nil, nil, err
	}
	v, err := l.ContractVersion()
	if err != nil {
		return nil, nil, err
	}
	if v != project.SupportedContractVersion {
		return nil, nil, fmt.Errorf("project contract version %q is not supported by this CLI (supports %q)", v, project.SupportedContractVersion)
	}

	wf, workflowPath, err := resolveWorkflowByIdentity(l, workflowIdentity)
	if err != nil {
		return nil, nil, err
	}

	rep, outcome, err := runEligibleWorkflowWithOutcome(root, l, wf, workflowPath, targetKindFor(wf), target, agentOverride, chooser)
	if outcome == nil {
		return rep, nil, err
	}
	return rep, ActionableFindings(wf.Identity, outcome.Payload), err
}

// targetKindFor derives how Run should treat a workflow's target purely from its own Contract:
// specification_reference: none means no CLI-checked Specification concept applies to this
// workflow — cli/WORKFLOW_CONTRACT.md's own words — so any supplied target is treated as
// free-form/untyped, never existence/lifecycle-checked and never described to the Agent as a
// Specification; required or optional means the target genuinely is a Specification identity,
// gated by facts.Eligible exactly as implement's is. This is the one piece of Contract metadata
// the existing design already ties directly to this distinction; Run never hardcodes a
// per-workflow-name table to reconstruct it, and this produces byte-identical treatment to every
// existing dedicated command for every input that command could ever have received.
func targetKindFor(wf contract.Workflow) targetKind {
	if wf.SpecificationReference == contract.SpecReferenceNone {
		return targetGeneric
	}
	return targetSpec
}

// resolveWorkflowByIdentity scans every workflow file in the project's .gnomon/workflows/
// directory for the one whose own declared Contract identity matches — Run's own resolution
// mechanism, since, unlike every dedicated command, it does not know its target filename in
// advance. A file that fails to load (malformed or absent Contract metadata) is excluded from
// the scan, exactly as cli/WORKFLOW_CONTRACT.md's own validation table treats a file with no
// readable contract metadata as "safe to ignore for discovery purposes" — such a file cannot
// match any identity in the first place. More than one file declaring the same identity is
// refused outright, per that same table's "Duplicate identity across two files... the CLI
// refuses to route to either until resolved."
func resolveWorkflowByIdentity(l project.Layout, identity string) (contract.Workflow, string, error) {
	entries, err := os.ReadDir(l.WorkflowsDir())
	if err != nil {
		return contract.Workflow{}, "", err
	}

	type match struct {
		wf   contract.Workflow
		path string
	}
	var matches []match
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".md" {
			continue
		}
		path := filepath.Join(l.WorkflowsDir(), e.Name())
		wf, err := contract.Load(path)
		if err != nil {
			continue // unreadable or invalid Contract metadata — excluded from routing
		}
		if wf.Identity == identity {
			matches = append(matches, match{wf, path})
		}
	}
	switch len(matches) {
	case 0:
		return contract.Workflow{}, "", fmt.Errorf("no workflow with identity %q found in %s", identity, l.WorkflowsDir())
	case 1:
		return matches[0].wf, matches[0].path, nil
	default:
		return contract.Workflow{}, "", fmt.Errorf("multiple workflow files declare identity %q — refusing to route to either until resolved", identity)
	}
}

// prepareImplementation is a thin, Implementation-specific convenience over prepareWorkflow and
// facts.Eligible — kept exactly as Slice 0/1 left it, both in signature and behavior, so every
// test written against that shape continues to work unchanged.
func prepareImplementation(root, specID string) (project.Layout, contract.Workflow, facts.Eligibility, error) {
	l, wf, _, err := prepareWorkflow(root, "implementation.md")
	if err != nil {
		return project.Layout{}, contract.Workflow{}, facts.Eligibility{}, err
	}
	elig, err := facts.Eligible(l, wf, specID)
	if err != nil {
		return project.Layout{}, contract.Workflow{}, facts.Eligibility{}, err
	}
	return l, wf, elig, nil
}

// prepareWorkflow resolves everything every workflow-invoking command needs before deciding
// whether to touch an Agent provider or check eligibility at all: project location and
// contract-version compatibility, then the named workflow's own Contract — plus the exact path
// that Contract was loaded from, since cli/WORKFLOW_CONTRACT.md's own `identity` field is
// "independent of the file's name or path": nothing downstream may assume a workflow's path can
// be reconstructed from its identity (true for all 10 built-in files today, not guaranteed for a
// custom one Run might resolve by identity alone). Shared by every workflow-invoking command.
func prepareWorkflow(root, filename string) (project.Layout, contract.Workflow, string, error) {
	l, err := project.Locate(root)
	if err != nil {
		return project.Layout{}, contract.Workflow{}, "", err
	}

	v, err := l.ContractVersion()
	if err != nil {
		return project.Layout{}, contract.Workflow{}, "", err
	}
	if v != project.SupportedContractVersion {
		return project.Layout{}, contract.Workflow{}, "", fmt.Errorf("project contract version %q is not supported by this CLI (supports %q)", v, project.SupportedContractVersion)
	}

	workflowPath := filepath.Join(l.WorkflowsDir(), filename)
	wf, err := contract.Load(workflowPath)
	if err != nil {
		return project.Layout{}, contract.Workflow{}, "", err
	}
	return l, wf, workflowPath, nil
}

// implementWithAdapter is a thin, Implementation-specific convenience over
// runWorkflowWithAdapter, kept exactly as Slice 0/1 left it — both in signature and behavior — so
// every test exercising it directly against a fake Adapter continues to work unchanged. Its own
// workflowPath is safe to fix as "implementation.md" here precisely because this function is
// itself hardcoded to Implementation, not a generic dispatcher.
func implementWithAdapter(root string, l project.Layout, wf contract.Workflow, specID string, ad adapter.Adapter) (*present.Report, error) {
	workflowPath := filepath.Join(l.WorkflowsDir(), "implementation.md")
	return runWorkflowWithAdapter(root, l, wf, workflowPath, targetSpec, specID, ad)
}

// runWorkflow resolves the Agent provider — the one function every Agent-invoking command in
// this slice, plus Implement, goes through; orchestration never constructs a ClaudeAdapter or
// CodexAdapter directly — and, once resolved, hands off to runWorkflowWithAdapter for the shared
// prepare/run/consume/classify/render lifecycle every one of these commands shares.
func runWorkflow(root string, l project.Layout, wf contract.Workflow, workflowPath string, kind targetKind, target, agentOverride string, chooser AgentChooser) (*present.Report, error) {
	rep, _, err := runWorkflowWithOutcome(root, l, wf, workflowPath, kind, target, agentOverride, chooser)
	return rep, err
}

// runWorkflowWithOutcome is runWorkflow's sibling for RunEvaluation, the one caller that also
// needs the raw Outcome once an Agent is actually resolved and run. runWorkflow is now a
// one-line delegator to this, so the Agent-resolution logic itself still exists exactly once.
func runWorkflowWithOutcome(root string, l project.Layout, wf contract.Workflow, workflowPath string, kind targetKind, target, agentOverride string, chooser AgentChooser) (*present.Report, *result.Outcome, error) {
	return runWorkflowFull(root, l, wf, workflowPath, kind, target, ResolutionHandoff{}, agentOverride, chooser)
}

// runWorkflowFull is runWorkflowWithOutcome's sibling for callers that also supply a
// ResolutionHandoff. The Agent-resolution logic itself exists exactly once here;
// runWorkflowWithOutcome is now a one-line delegator with a zero-value handoff, so every existing
// caller and test is unaffected.
func runWorkflowFull(root string, l project.Layout, wf contract.Workflow, workflowPath string, kind targetKind, target string, handoff ResolutionHandoff, agentOverride string, chooser AgentChooser) (*present.Report, *result.Outcome, error) {
	ad, err := ResolveAgent(agentOverride, chooser)
	if err != nil {
		return &present.Report{
			Outcome: present.Failed,
			Summary: "Could not determine which Agent to use",
			Target:  target,
			Sections: []present.Section{
				{Label: "Reason", Body: err.Error()},
			},
			Next: "Pass --agent claude|codex, or run `gnomon agent set-default <claude|codex>`.",
		}, nil, err
	}

	return runWorkflowWithAdapterFull(root, l, wf, workflowPath, kind, target, handoff, ad)
}

// runResolutionAgent is runWorkflowFull's 2-return-value convenience for resolution dispatch
// (implementWithHandoff, testWithHandoff), which never needs the raw Outcome the way
// RunEvaluation does — a resolution workflow's own result is rendered and shown, never mined for
// findings.
func runResolutionAgent(root string, l project.Layout, wf contract.Workflow, workflowPath string, kind targetKind, target string, handoff ResolutionHandoff, agentOverride string, chooser AgentChooser) (*present.Report, error) {
	rep, _, err := runWorkflowFull(root, l, wf, workflowPath, kind, target, handoff, agentOverride, chooser)
	return rep, err
}

// runWorkflowWithAdapter is the generic execution lifecycle every Agent-invoking command in this
// slice shares — load Contract → (eligibility, checked by the caller) → resolve Agent → invoke →
// receive the transient Result Protocol payload → validate against the Result Contract → extract
// the terminal value → classify using the Contract → render generically. Kept separate from
// runWorkflow, and unexported, so this lifecycle remains directly testable against a fake
// Adapter, without a real provider resolution in the way — the same reason implementWithAdapter
// was split out in Slice 1. Nothing here ever switches on a workflow's identity or its result
// vocabulary; every workflow-specific fact comes from the already-loaded Contract (wf.Result),
// and workflowPath is always the caller's own already-resolved path — never reconstructed from
// wf.Identity here.
func runWorkflowWithAdapter(root string, l project.Layout, wf contract.Workflow, workflowPath string, kind targetKind, target string, ad adapter.Adapter) (*present.Report, error) {
	rep, _, err := runWorkflowWithAdapterAndOutcome(root, l, wf, workflowPath, kind, target, ad)
	return rep, err
}

// runWorkflowWithAdapterAndOutcome is runWorkflowWithAdapter's sibling, additionally returning
// the raw validated Outcome — needed only by RunEvaluation (via runWorkflowWithOutcome), which
// reads Verification/Review's own findings from its payload. runWorkflowWithAdapter is now a
// one-line delegator to this, so the classify/render lifecycle itself still exists exactly once,
// and every existing test calling runWorkflowWithAdapter directly against a fake Adapter is
// unaffected.
func runWorkflowWithAdapterAndOutcome(root string, l project.Layout, wf contract.Workflow, workflowPath string, kind targetKind, target string, ad adapter.Adapter) (*present.Report, *result.Outcome, error) {
	return runWorkflowWithAdapterFull(root, l, wf, workflowPath, kind, target, ResolutionHandoff{}, ad)
}

// runWorkflowWithAdapterFull is runWorkflowWithAdapterAndOutcome's sibling for callers that also
// supply a ResolutionHandoff — the classify/render lifecycle exists exactly once here;
// runWorkflowWithAdapterAndOutcome is now a one-line delegator with a zero-value handoff, so
// every existing test calling it (or runWorkflowWithAdapter) directly against a fake Adapter is
// unaffected.
func runWorkflowWithAdapterFull(root string, l project.Layout, wf contract.Workflow, workflowPath string, kind targetKind, target string, handoff ResolutionHandoff, ad adapter.Adapter) (*present.Report, *result.Outcome, error) {
	outcome, detail, failureRep, err := obtainWorkflowOutcomeFull(root, l, wf, workflowPath, kind, target, handoff, ad)
	if failureRep != nil {
		return failureRep, nil, err
	}
	rep, err := classifyAndRender(wf, target, detail, outcome)
	return rep, outcome, err
}

// obtainWorkflowOutcome runs the Agent and returns its raw, Contract-validated but not yet
// classified/rendered result — the part of runWorkflowWithAdapter's sequence every caller shares
// (launch, poll, consume). A non-nil failureRep means the run never produced a valid result at
// all (process failure, cancellation, or protocol error): callers must return it as-is rather
// than attempting to classify anything. This split exists so a caller needing the raw payload
// before generic rendering (Discovery's interactive accept flow) can reuse this exact sequence
// instead of duplicating it.
func obtainWorkflowOutcome(root string, l project.Layout, wf contract.Workflow, workflowPath string, kind targetKind, target string, ad adapter.Adapter) (*result.Outcome, []string, *present.Report, error) {
	return obtainWorkflowOutcomeFull(root, l, wf, workflowPath, kind, target, ResolutionHandoff{}, ad)
}

// obtainWorkflowOutcomeFull is obtainWorkflowOutcome's sibling for resolution dispatch — the one
// place adapter.Context is actually constructed, and the only place a non-empty
// ResolutionHandoff ever becomes Context.Handoff. obtainWorkflowOutcome is now a one-line
// delegator with a zero-value handoff (adapterHandoff() returns nil for it), so every existing
// caller and test is unaffected.
func obtainWorkflowOutcomeFull(root string, l project.Layout, wf contract.Workflow, workflowPath string, kind targetKind, target string, handoff ResolutionHandoff, ad adapter.Adapter) (*result.Outcome, []string, *present.Report, error) {
	runID, err := result.NewRunID()
	if err != nil {
		return nil, nil, nil, err
	}
	destPath, err := result.Destination(root, runID)
	if err != nil {
		return nil, nil, nil, err
	}

	// Captured by the polling closure the instant a valid result is detected, while the Agent may
	// still be running — this is what lets Gnomon terminate the session itself rather than
	// requiring the Agent to end it. result.Consume both validates and deletes the transient file
	// on success, so it is safe to call repeatedly from the closure: every call before the real
	// one simply fails (not yet written, still being written, or not yet matching) and is treated
	// as "not ready", never as an error worth reporting.
	var polled *result.Outcome
	ctx := adapter.Context{
		ProjectRoot:      root,
		GnomonRoot:       l.GnomonRoot(),
		WorkflowPath:     workflowPath,
		WorkflowIdentity: wf.Identity,
		ResultPath:       destPath,
		RunID:            runID,
		PollResult: func() (bool, error) {
			outcome, err := result.Consume(destPath, wf.Identity, runID, wf.Result)
			if err != nil {
				return false, nil
			}
			polled = &outcome
			return true, nil
		},
	}
	switch kind {
	case targetSpec:
		ctx.SpecIdentity = target
	case targetGeneric:
		ctx.Target = target
	}
	ctx.Handoff = handoff.adapterHandoff()

	if err := ad.Prepare(ctx); err != nil {
		return nil, nil, nil, err
	}
	_ = ad.Run()
	status := ad.Status()

	detail := []string{
		fmt.Sprintf("process: %s", processLabel(status)),
		fmt.Sprintf("run_id: %s", runID),
	}

	outcome := polled
	var consumeErr error
	if outcome == nil {
		// The polling loop never caught a valid result before the process ended — one final,
		// authoritative check in case one appeared in the narrow window right at the end, exactly
		// as a run with no polling at all would have done.
		o, cErr := result.Consume(destPath, wf.Identity, runID, wf.Result)
		if cErr != nil {
			consumeErr = cErr
		} else {
			outcome = &o
		}
	}

	if consumeErr != nil {
		label := humanizeIdentity(wf.Identity)
		rep := &present.Report{Target: target, Detail: detail}
		switch {
		case status.Terminated:
			rep.Outcome = present.Cancelled
			rep.Summary = fmt.Sprintf("%s cancelled before completion", label)
		case status.Err != nil || (status.ExitCode != 0 && !status.Terminated):
			rep.Outcome = present.Failed
			rep.Summary = fmt.Sprintf("%s failed", label)
			rep.AddSection("Reason", "The Agent "+processLabel(status)+".")
		default:
			rep.Outcome = present.Failed
			rep.Summary = fmt.Sprintf("%s could not complete", label)
			rep.AddSection("Reason", humanizeProtocolError(consumeErr)+".")
		}
		return nil, nil, rep, consumeErr
	}

	return outcome, detail, nil, nil
}

// classifyAndRender is the generic classify-then-render half of runWorkflowWithAdapter's
// sequence, driven entirely by the workflow's own already-loaded Contract — never a hardcoded
// switch over a specific workflow's terminal vocabulary. wf.Validate() (run inside contract.Load,
// always before a workflow can be invoked at all) already guarantees every value the schema's
// enum permits has an explicit "success"/"blocked" classification, so a schema-conformant payload
// reaching here with an unclassified terminal value is expected to be unreachable — handled
// defensively below as a failure, never silently treated as success. This is also exactly what
// keeps a workflow's own valid negative conclusion (Verification FAIL, Review DEFECT/RISK/
// KNOWLEDGE GAP) from ever being reported as a process/protocol failure: those values reach here
// as a normally-classified "blocked" outcome, never through obtainWorkflowOutcome's failure path.
func classifyAndRender(wf contract.Workflow, target string, detail []string, outcome *result.Outcome) (*present.Report, error) {
	rep := &present.Report{Target: target, Detail: detail}

	class, known := wf.Result.ClassificationFor(outcome.Terminal)
	if !known {
		rep.Outcome = present.Failed
		rep.Summary = fmt.Sprintf("Unrecognized workflow result: %s", outcome.Terminal)
		rep.AddSection("Reason", "This terminal value has no declared classification in the workflow's own Result Contract.")
		return rep, fmt.Errorf("unclassified terminal value: %s", outcome.Terminal)
	}
	switch class {
	case "success":
		rep.Outcome = present.Success
	case "blocked":
		rep.Outcome = present.Blocked
	}
	rep.Summary = humanizeTerminalValue(outcome.Terminal)
	rep.Sections = renderPayloadSections(wf.Result, outcome.Payload)

	if rep.Outcome != present.Success {
		return rep, fmt.Errorf("workflow reported %s", outcome.Terminal)
	}
	return rep, nil
}

// processLabel is the single, reusable classification of what happened to the Agent process,
// used both in verbose diagnostics and in failure-path summaries — so the same fact is always
// described the same way, and a Gnomon-initiated termination after success is never worded as if
// it were an Agent failure.
func processLabel(status adapter.Status) string {
	switch {
	case status.Err != nil:
		return fmt.Sprintf("failed to run (%v)", status.Err)
	case status.Terminated:
		if status.ExitCode != 0 {
			return fmt.Sprintf("ended automatically once Gnomon confirmed the result (exit code %d)", status.ExitCode)
		}
		return "ended automatically once Gnomon confirmed the result"
	case status.ExitCode != 0:
		return fmt.Sprintf("exited unexpectedly (exit code %d)", status.ExitCode)
	default:
		return "exited normally"
	}
}

// humanizeProtocolError strips the Result Protocol's own internal "protocol failure: " label,
// leaving the already-plain-English remainder for a Human to read directly.
func humanizeProtocolError(err error) string {
	return strings.TrimPrefix(err.Error(), "protocol failure: ")
}

// humanizeIdentity turns a workflow's own hyphenated identity into a natural sentence-leading
// label — e.g. "git-finalization" -> "Git finalization", "implementation" -> "Implementation" —
// the same mechanical, no-hardcoded-name transform humanizeTerminalValue (render.go) already
// applies to terminal values, applied here to the one other place a workflow's own name needs to
// appear in Human-facing text (process-outcome summaries), without ever naming a specific
// workflow in Go code.
func humanizeIdentity(identity string) string {
	words := strings.Split(identity, "-")
	for i, w := range words {
		if i == 0 && w != "" {
			words[i] = strings.ToUpper(w[:1]) + w[1:]
		}
	}
	return strings.Join(words, " ")
}
