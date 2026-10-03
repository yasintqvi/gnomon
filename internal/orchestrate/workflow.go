package orchestrate

import (
	"errors"
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
	"gnomon/internal/specs"
)

// targetKind is invocation-shape configuration fixed per command, never derived from a workflow's
// runtime result. targetGeneric (verify, review) is never existence/lifecycle-checked and is
// never described to the Agent as a Specification — cli/WORKFLOW_CONTRACT.md, specification_reference: none.
type targetKind int

const (
	targetNone    targetKind = iota // no target argument (describe, bootstrap, finalize)
	targetSpec                      // a Specification identity — existence/lifecycle-checked (implement)
	targetGeneric                   // a free-form, untyped target — never checked (verify, review)
)

// runRequest carries everything one Agent-invoking run needs. handoff's zero value means no
// handoff; a nil adapter means "resolve one via agentOverride/chooser"; a non-nil one is used as-is.
type runRequest struct {
	root          string
	l             project.Layout
	wf            contract.Workflow
	workflowPath  string
	kind          targetKind
	target        string
	handoff       ResolutionHandoff
	agentOverride string
	chooser       AgentChooser
	adapter       adapter.Adapter
	verification  VerificationOptions // Verification only
	scopeNote     string              // set by execute for Verification
}

// Implement performs `gnomon implement <SPEC-id>`. Eligibility is checked before any Agent
// provider is resolved, so an ineligible invocation never triggers the first-use provider prompt.
func Implement(root, specID, agentOverride string, chooser AgentChooser) (*present.Report, error) {
	return implementWithHandoff(root, specID, ResolutionHandoff{}, agentOverride, chooser)
}

// implementWithHandoff is Implement's sibling for the interactive finding-resolution loop
// (cmd/gnomon, via RunResolutionWorkflow), the only caller that supplies a non-empty handoff.
func implementWithHandoff(root, specID string, handoff ResolutionHandoff, agentOverride string, chooser AgentChooser) (*present.Report, error) {
	l, wf, workflowPath, elig, err := prepareImplementation(root, specID)
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
			// The reason varies (not approved, not found, ...), so point at the workspace rather
			// than guessing a specific next action.
			Next: specWorkspaceNext(l, specID, ""),
		}, fmt.Errorf("ineligible: %s", elig.Reason)
	}

	rep, _, err := execute(runRequest{root: root, l: l, wf: wf, workflowPath: workflowPath, kind: targetSpec, target: specID, handoff: handoff, agentOverride: agentOverride, chooser: chooser})
	return rep, err
}

// DescribeSuccessNext and BootstrapSuccessNext are the single source for each onboarding
// recommendation's text — referenced by Describe/Bootstrap, their tests, and cmd/gnomon's
// registered-command cross-check, so they can't silently drift from the real CLI command tree.
// Onboarding's next step is fixed by cli/PROJECT_INITIALIZATION.md, not derived like
// specWorkspaceNext's, so a plain constant is enough.
const (
	DescribeSuccessNext  = "gnomon bootstrap"
	BootstrapSuccessNext = "gnomon spec\n  Create a Specification directly, or have an Agent propose one (Discover)."
)

// Describe performs `gnomon describe`, wired to workflows/initial-knowledge-establishment.md.
// It takes no target, and the CLI does not pre-classify the project as new/existing — the Agent
// inspects the repository and asks the Human directly when knowledge can't be established from it.
func Describe(root, agentOverride string, chooser AgentChooser) (*present.Report, error) {
	rep, err := runNoTargetWorkflow(root, "initial-knowledge-establishment.md", agentOverride, chooser)
	onboardingNext(rep, DescribeSuccessNext)
	return rep, err
}

// Bootstrap performs `gnomon bootstrap`, wired to workflows/bootstrap.md — distinct from, and
// assuming the baseline of, Initial Knowledge Establishment.
func Bootstrap(root, agentOverride string, chooser AgentChooser) (*present.Report, error) {
	rep, err := runNoTargetWorkflow(root, "bootstrap.md", agentOverride, chooser)
	onboardingNext(rep, BootstrapSuccessNext)
	return rep, err
}

// onboardingNext sets Next to onSuccess only on a Success Outcome.
func onboardingNext(rep *present.Report, onSuccess string) {
	if rep != nil && rep.Outcome == present.Success {
		rep.Next = onSuccess
	}
}

// Finalize performs `gnomon finalize`, wired to workflows/git-finalization.md. It takes no target
// and adds no authorization gate beyond ordinary eligibility: whether a commit/push/publish is
// authorized is decided entirely by the Agent following the workflow's own Rules, never by Gnomon.
func Finalize(root, agentOverride string, chooser AgentChooser) (*present.Report, error) {
	return runNoTargetWorkflow(root, "git-finalization.md", agentOverride, chooser)
}

// Verify performs `gnomon verify [target]`, wired to workflows/verification.md. target is
// generic/untyped (see targetKind).
func Verify(root, target, agentOverride string, chooser AgentChooser) (*present.Report, error) {
	return runGenericTargetWorkflow(root, "verification.md", target, agentOverride, chooser)
}

// Review performs `gnomon review [target]`, wired to workflows/review.md. Same generic/untyped
// target treatment as Verify.
func Review(root, target, agentOverride string, chooser AgentChooser) (*present.Report, error) {
	return runGenericTargetWorkflow(root, "review.md", target, agentOverride, chooser)
}

// SpecDiscover performs `gnomon spec discover`, wired to workflows/specification-discovery.md. It
// takes no target and creates or mutates nothing itself — accepting a proposed candidate is a
// separate, subsequent Human action through `gnomon spec create <title>` (SpecCreate, spec.go).
func SpecDiscover(root, agentOverride string, chooser AgentChooser) (*present.Report, error) {
	return runNoTargetWorkflow(root, "specification-discovery.md", agentOverride, chooser)
}

// SpecDefine performs `gnomon spec define <SPEC-id>`, wired to
// workflows/specification-definition.md. Its target must already exist (Definition never creates
// one) but is not required to be Approved. A `READY_FOR_APPROVAL` result never grants approval —
// that remains exclusively `gnomon approve`'s own act.
func SpecDefine(root, specID, agentOverride string, chooser AgentChooser) (*present.Report, error) {
	return runTargetedWorkflow(root, "specification-definition.md", runRequest{kind: targetSpec, target: specID, agentOverride: agentOverride, chooser: chooser})
}

// Test performs `gnomon test [SPEC-id]`, wired to workflows/testing.md. With no Specification it
// is always eligible (exploratory testing); with one, it must exist and be Approved.
func Test(root, specID, agentOverride string, chooser AgentChooser) (*present.Report, error) {
	return testWithHandoff(root, specID, ResolutionHandoff{}, agentOverride, chooser)
}

// testWithHandoff is Test's sibling for the interactive finding-resolution loop, the only caller
// that supplies a non-empty handoff.
func testWithHandoff(root, specID string, handoff ResolutionHandoff, agentOverride string, chooser AgentChooser) (*present.Report, error) {
	l, wf, workflowPath, err := prepareWorkflow(root, "testing.md")
	if err != nil {
		return nil, err
	}
	rep, _, err := runEligible(runRequest{root: root, l: l, wf: wf, workflowPath: workflowPath, kind: targetSpec, target: specID, handoff: handoff, agentOverride: agentOverride, chooser: chooser})
	return rep, err
}

// runNoTargetWorkflow and runGenericTargetWorkflow are convenience shapes over runTargetedWorkflow
// for the targetNone and targetGeneric cases (describe/bootstrap/finalize; verify/review).
func runNoTargetWorkflow(root, filename, agentOverride string, chooser AgentChooser) (*present.Report, error) {
	return runTargetedWorkflow(root, filename, runRequest{kind: targetNone, agentOverride: agentOverride, chooser: chooser})
}

func runGenericTargetWorkflow(root, filename, target, agentOverride string, chooser AgentChooser) (*present.Report, error) {
	return runTargetedWorkflow(root, filename, runRequest{kind: targetGeneric, target: target, agentOverride: agentOverride, chooser: chooser})
}

// runTargetedWorkflow loads the named workflow's Contract, then delegates to runEligible. req
// carries only kind/target/agentOverride/chooser; root/l/wf/workflowPath are filled in here once
// the named file is loaded.
func runTargetedWorkflow(root, filename string, req runRequest) (*present.Report, error) {
	l, wf, workflowPath, err := prepareWorkflow(root, filename)
	if err != nil {
		return nil, err
	}
	req.root, req.l, req.wf, req.workflowPath = root, l, wf, workflowPath
	rep, _, err := runEligible(req)
	return rep, err
}

// runEligible checks pre-start eligibility for an already-loaded workflow and, only if eligible,
// calls execute. Its ineligibility Report is deliberately generic — unlike Implement's own
// tailored message — since elig.Reason (e.g. "SPEC-003 is not Approved") already says what to do.
func runEligible(req runRequest) (*present.Report, *result.Outcome, error) {
	elig, err := facts.Eligible(req.l, req.wf, req.target)
	if err != nil {
		return nil, nil, err
	}
	if !elig.Eligible {
		return &present.Report{
			Outcome: present.Blocked,
			Summary: fmt.Sprintf("%s blocked", humanizeIdentity(req.wf.Identity)),
			Target:  req.target,
			Sections: []present.Section{
				{Label: "Unresolved", Body: elig.Reason},
			},
		}, nil, fmt.Errorf("ineligible: %s", elig.Reason)
	}

	return execute(req)
}

// Run performs `gnomon run <workflow-identity> [target]` — reaching any workflow by its own
// declared Contract identity, built-in or custom (cli/COMMAND_SURFACE.md, "run Rule"). Unlike a
// dedicated command, it resolves the target filename by identity (resolveWorkflowByIdentity) and
// derives target treatment from the Contract (targetKindFor) rather than a per-workflow table.
func Run(root, workflowIdentity, target, agentOverride string, chooser AgentChooser) (*present.Report, error) {
	rep, _, err := RunEvaluation(root, workflowIdentity, target, agentOverride, chooser)
	return rep, err
}

// RunEvaluation is Run's sibling, additionally returning Verification/Review's own actionable
// findings (extracted via ActionableFindings) for the interactive resolution loop. findings is
// nil for every other identity, and for a clean PASS — never treated as an error either way.
func RunEvaluation(root, workflowIdentity, target, agentOverride string, chooser AgentChooser) (*present.Report, []EvaluationFinding, error) {
	return RunEvaluationWithOptions(root, workflowIdentity, target, agentOverride, chooser, VerificationOptions{})
}

// RunEvaluationWithOptions is RunEvaluation with Verification's scope options (--full, --since),
// which are refused for any other workflow.
func RunEvaluationWithOptions(root, workflowIdentity, target, agentOverride string, chooser AgentChooser, opts VerificationOptions) (*present.Report, []EvaluationFinding, error) {
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

	if (opts.Full || opts.Since != "") && wf.Identity != verificationIdentity {
		return nil, nil, fmt.Errorf("--full and --since apply only to verification, not %s", wf.Identity)
	}
	if opts.Full && opts.Since != "" {
		return nil, nil, fmt.Errorf("--full checks everything; it cannot be combined with --since")
	}
	if opts.Since != "" && target != "" {
		return nil, nil, fmt.Errorf("--since limits verification to a change; it cannot be combined with a target")
	}

	rep, outcome, err := runEligible(runRequest{root: root, l: l, wf: wf, workflowPath: workflowPath, kind: targetKindFor(wf), target: target, agentOverride: agentOverride, chooser: chooser, verification: opts})
	if outcome == nil {
		return rep, nil, err
	}
	return rep, ActionableFindings(wf.Identity, outcome.Payload), err
}

// targetKindFor derives how Run should treat a workflow's target purely from its own Contract's
// specification_reference — never a hardcoded per-workflow-name table.
func targetKindFor(wf contract.Workflow) targetKind {
	if wf.SpecificationReference == contract.SpecReferenceNone {
		return targetGeneric
	}
	return targetSpec
}

// resolveWorkflowByIdentity finds the workflow in effect (bundled, or the project's own) whose
// declared Contract identity matches. A file that fails to load is silently excluded
// (cli/WORKFLOW_CONTRACT.md). More than one file declaring the same identity is refused outright.
func resolveWorkflowByIdentity(l project.Layout, identity string) (contract.Workflow, string, error) {
	all, _, err := l.Workflows()
	if err != nil {
		return contract.Workflow{}, "", err
	}

	var matches []project.WorkflowFile
	var matched contract.Workflow
	for _, f := range all {
		wf, err := contract.Parse(f.Data, f.Name)
		if err != nil {
			continue // invalid Contract metadata — excluded from routing
		}
		if wf.Identity == identity {
			matches = append(matches, f)
			matched = wf
		}
	}
	switch len(matches) {
	case 0:
		return contract.Workflow{}, "", fmt.Errorf("no workflow with identity %q (bundled or in %s)", identity, l.WorkflowsDir())
	case 1:
		path, err := readableWorkflowPath(l, matches[0])
		if err != nil {
			return contract.Workflow{}, "", err
		}
		return matched, path, nil
	default:
		return contract.Workflow{}, "", fmt.Errorf("multiple workflow files declare identity %q — refusing to route to either until resolved", identity)
	}
}

// loadWorkflowContract parses the Contract of the workflow in effect under filename, without
// writing anything.
func loadWorkflowContract(l project.Layout, filename string) (contract.Workflow, error) {
	f, err := l.Workflow(filename)
	if err != nil {
		return contract.Workflow{}, err
	}
	return contract.Parse(f.Data, filename)
}

// readableWorkflowPath returns a path the Agent can read the workflow from: the project's own file
// when customized, otherwise the current bundled version written to the gitignored runtime
// directory (refreshed on every run, so it can never go stale).
func readableWorkflowPath(l project.Layout, f project.WorkflowFile) (string, error) {
	if f.Customized() {
		return f.ProjectPath, nil
	}
	dir, err := result.TransientDir(l.Root)
	if err != nil {
		return "", err
	}
	dir = filepath.Join(dir, "workflows")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	path := filepath.Join(dir, f.Name)
	return path, os.WriteFile(path, f.Data, 0o644)
}

// prepareImplementation is a thin, Implementation-specific convenience over prepareWorkflow and
// facts.Eligible.
func prepareImplementation(root, specID string) (project.Layout, contract.Workflow, string, facts.Eligibility, error) {
	l, wf, path, err := prepareWorkflow(root, "implementation.md")
	if err != nil {
		return project.Layout{}, contract.Workflow{}, "", facts.Eligibility{}, err
	}
	elig, err := facts.Eligible(l, wf, specID)
	if err != nil {
		return project.Layout{}, contract.Workflow{}, "", facts.Eligibility{}, err
	}
	return l, wf, path, elig, nil
}

// prepareWorkflow resolves project location, contract-version compatibility, and the named
// workflow in effect (the project's customization, or the bundled version), returning its Contract
// and a path the Agent can read it from. Shared by every caller.
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

	f, err := l.Workflow(filename)
	if err != nil {
		return project.Layout{}, contract.Workflow{}, "", err
	}
	label := filename
	if f.Customized() {
		label = f.ProjectPath
	}
	wf, err := contract.Parse(f.Data, label)
	if err != nil {
		return project.Layout{}, contract.Workflow{}, "", err
	}
	workflowPath, err := readableWorkflowPath(l, f)
	if err != nil {
		return project.Layout{}, contract.Workflow{}, "", err
	}
	return l, wf, workflowPath, nil
}

// implementWithAdapter is a thin, Implementation-specific convenience over execute, used directly
// by tests against a fake Adapter.
func implementWithAdapter(root string, l project.Layout, wf contract.Workflow, specID string, ad adapter.Adapter) (*present.Report, error) {
	f, err := l.Workflow("implementation.md")
	if err != nil {
		return nil, err
	}
	workflowPath, err := readableWorkflowPath(l, f)
	if err != nil {
		return nil, err
	}
	rep, _, err := execute(runRequest{root: root, l: l, wf: wf, workflowPath: workflowPath, kind: targetSpec, target: specID, adapter: ad})
	return rep, err
}

// execute is the one run path every Agent-invoking command shares: resolve the Agent (unless
// req.adapter is already given), invoke, validate the result, classify, render. It never checks
// eligibility itself — that's runEligible's job — so callers that already checked it their own way
// (Implement's tailored message) call this directly.
//
// It also carries Rule B: any Specification that derived Approved before the run and no longer
// does afterward is reported as a warning on the Report — never restored, never a reason to
// reject. Wrapped here rather than in obtainOutcome because it needs to annotate the final Report.
func execute(req runRequest) (rep *present.Report, outcome *result.Outcome, err error) {
	// Early check so a refused run never asks which Agent to use; obtainOutcome's own acquisition
	// is the one that counts.
	if active := probeRunLock(req.root); active != nil {
		return runActiveReport(req.target, active), nil, active
	}

	var scope *verificationScope
	if req.wf.Identity == verificationIdentity {
		var nothingToDo *present.Report
		scope, nothingToDo, err = decideVerificationScope(req.l, req.target, req.verification)
		if err != nil {
			return nil, nil, err
		}
		if nothingToDo != nil {
			nothingToDo.Target = req.target
			return nothingToDo, nil, nil
		}
		req.scopeNote = scope.note
	}

	ad := req.adapter
	if ad == nil {
		ad, err = ResolveAgent(req.agentOverride, req.chooser)
		if err != nil {
			return &present.Report{
				Outcome: present.Failed,
				Summary: "Could not determine which Agent to use",
				Target:  req.target,
				Sections: []present.Section{
					{Label: "Reason", Body: err.Error()},
				},
				Next: "Pass --agent claude|codex, or run `gnomon agent set-default <claude|codex>`.",
			}, nil, err
		}
		req.adapter = ad
	}

	approvedBefore, snapErr := approvedIdentities(req.l)
	if snapErr != nil {
		return nil, nil, snapErr
	}
	defer func() {
		var active *RunActiveError
		if rep == nil || errors.As(err, &active) {
			return
		}
		lost, lostErr := lostApprovals(req.l, approvedBefore, governingSpecIdentity(req.wf, req.kind, req.target))
		if lostErr != nil || len(lost) == 0 {
			return
		}
		rep.Sections = append(rep.Sections, lostApprovalSection(lost))
	}()

	var detail []string
	var failureRep *present.Report
	outcome, detail, failureRep, err = obtainOutcome(req)
	if failureRep != nil {
		rep, outcome = failureRep, nil
		return
	}
	if req.wf.Identity == verificationIdentity {
		enforceEvidence(outcome, req.l.Root)
	}
	rep, err = classifyAndRender(req.wf, req.target, detail, outcome)
	if scope != nil && scope.recordState && err == nil && outcome != nil && outcome.Terminal == "PASS" {
		if recErr := recordVerificationState(req.l.Root, scope); recErr != nil && rep != nil {
			rep.AddSection("Baseline Not Recorded", "This passing verification could not be recorded as the starting point for the next one ("+recErr.Error()+"); the next standard run will check the same changes again.")
		}
	}
	return
}

// governingSpecIdentity returns the identity Rule A protects, or "" if it doesn't apply: the
// Contract must require an Approved Specification, and one must actually be supplied.
func governingSpecIdentity(wf contract.Workflow, kind targetKind, target string) string {
	if wf.RequiresApprovedSpecification && kind == targetSpec && target != "" {
		return target
	}
	return ""
}

// obtainOutcome runs the Agent (req.adapter must already be set) and returns its raw,
// Contract-validated but not yet classified/rendered result. A non-nil failureRep means the run
// never produced a valid result at all; callers must return it as-is. Discovery's interactive
// accept flow calls this directly, skipping execute's classify/render and Rule B.
//
// This wraps the run guard: approval evidence, and (Rule A) the governing Specification's file,
// are snapshotted before the run and compared after; either difference rejects the run outright
// (runguard.go). The project's run lock (runlock.go) is taken first: while it is held, a second
// Agent run in the same project and gnomon approve/revoke are refused.
func obtainOutcome(req runRequest) (outcome *result.Outcome, detail []string, failureRep *present.Report, err error) {
	runID, err := result.NewRunID()
	if err != nil {
		return nil, nil, nil, err
	}
	unlock, err := acquireRunLock(req.root, runID, req.wf.Identity, req.target)
	if err != nil {
		var active *RunActiveError
		if errors.As(err, &active) {
			return nil, nil, runActiveReport(req.target, active), err
		}
		return nil, nil, nil, err
	}
	defer unlock()

	approvalsDir := req.l.ApprovalsDir()
	before, err := snapshotTree(approvalsDir)
	if err != nil {
		return nil, nil, nil, err
	}

	var specBefore *specSnapshot
	if identity := governingSpecIdentity(req.wf, req.kind, req.target); identity != "" {
		specBefore, err = snapshotGoverningSpec(req.l.SpecificationsDir(), identity)
		if err != nil {
			return nil, nil, nil, err
		}
	}

	defer func() {
		// Runs on every path, including an otherwise-successful result, since that result is
		// still rejected if the Agent touched approval evidence or the governing Specification.
		after, snapErr := snapshotTree(approvalsDir)
		if snapErr != nil {
			outcome, detail, failureRep, err = nil, nil, approvalCheckFailedReport(snapErr), snapErr
			return
		}
		aResult := approvalsResult{Dir: approvalsDir, Diff: diffSnapshots(before, after)}

		var sResult specResult
		if specBefore != nil {
			viol, checkErr := checkGoverningSpec(req.l.SpecificationsDir(), specBefore)
			if checkErr != nil {
				outcome, detail, failureRep, err = nil, nil, specCheckFailedReport(specBefore.Identity, checkErr), checkErr
				return
			}
			sResult = specResult{Before: specBefore, Violation: viol}
		}

		if !aResult.violated() && !sResult.violated() {
			return
		}

		if aResult.violated() {
			aResult.RestoreFailed, aResult.RestoreErr = restoreSnapshot(before, aResult.Diff)
		}
		if sResult.violated() {
			sResult.Preserved, sResult.RestoreFailed, sResult.RestoreErr = preserveAndRestoreGoverningSpec(req.root, runID, specBefore)
		}

		failureRep = buildRejectionReport(aResult, sResult)
		outcome, detail = nil, nil
		err = fmt.Errorf("workflow run rejected: the agent tampered with approval evidence or the governing specification")
	}()

	return obtainOutcomeUnguarded(req, runID)
}

// obtainOutcomeUnguarded is obtainOutcome's unwrapped body, split out so the run guard wraps it via
// a plain call. runID is generated by the caller so the run lock and the Result Protocol
// destination agree on the same run.
func obtainOutcomeUnguarded(req runRequest, runID string) (*result.Outcome, []string, *present.Report, error) {
	destPath, err := result.Destination(req.root, runID)
	if err != nil {
		return nil, nil, nil, err
	}

	// Captured by the polling closure the instant a valid result is detected, so Gnomon can end the
	// session itself. Safe to call repeatedly: a call before the real result exists just fails, and
	// is treated as "not ready" rather than an error.
	var polled *result.Outcome
	ctx := adapter.Context{
		ProjectRoot:      req.root,
		GnomonRoot:       req.l.GnomonRoot(),
		WorkflowPath:     req.workflowPath,
		WorkflowIdentity: req.wf.Identity,
		ResultPath:       destPath,
		RunID:            runID,
		PollResult: func() (bool, error) {
			outcome, err := result.Consume(destPath, req.wf.Identity, runID, req.wf.Result)
			if err != nil {
				return false, nil
			}
			polled = &outcome
			return true, nil
		},
	}
	switch req.kind {
	case targetSpec:
		ctx.SpecIdentity = req.target
		if ok, path, err := specs.Exists(req.l.SpecificationsDir(), specs.Identity(req.target)); err == nil && ok {
			ctx.SpecPath = relToRoot(req.l, path)
		}
	case targetGeneric:
		ctx.Target = req.target
	}
	ctx.ScopeNote = req.scopeNote
	if req.wf.Identity == definitionIdentity {
		recorded, _, err := writeRecordedDecisions(req.root, req.l, runID, req.target)
		if err != nil {
			return nil, nil, nil, err
		}
		if recorded != "" {
			ctx.ReadFirst = append(ctx.ReadFirst, adapter.ReadFirstFile{Path: recorded,
				Purpose: "decisions and deferred notes already recorded in the other Specifications, with their source and approval state"})
		}
	}
	if knowledge, err := req.l.KnowledgeFiles(); err == nil {
		ctx.Knowledge = knowledge
	}
	ctx.Handoff = req.handoff.adapterHandoff()

	if err := req.adapter.Prepare(ctx); err != nil {
		return nil, nil, nil, err
	}
	_ = req.adapter.Run()
	status := req.adapter.Status()

	detail := []string{
		fmt.Sprintf("process: %s", processLabel(status)),
		fmt.Sprintf("run_id: %s", runID),
	}

	outcome := polled
	var consumeErr error
	if outcome == nil {
		// Polling never caught a valid result — one final check for the narrow window right at the end.
		o, cErr := result.Consume(destPath, req.wf.Identity, runID, req.wf.Result)
		if cErr != nil {
			consumeErr = cErr
		} else {
			outcome = &o
		}
	}

	if consumeErr != nil {
		label := humanizeIdentity(req.wf.Identity)
		rep := &present.Report{Target: req.target, Detail: detail}
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

// classifyAndRender classifies and renders a run's outcome, driven entirely by the workflow's own
// Contract — never a hardcoded switch over a specific workflow's terminal vocabulary. An
// unclassified terminal value should be unreachable (wf.Validate() guarantees the schema's enum is
// fully classified) but is handled defensively as a failure rather than silently treated as
// success. A workflow's own valid negative conclusion (Verification FAIL, Review DEFECT/RISK/
// KNOWLEDGE GAP) is classified "blocked" here, never routed through obtainOutcome's failure path.
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
	if wf.Identity == verificationIdentity {
		renderVerification(rep, outcome.Payload)
	} else {
		rep.Summary = humanizeTerminalValue(outcome.Terminal)
		rep.Sections = renderPayloadSections(wf.Result, outcome.Payload)
	}

	if rep.Outcome != present.Success {
		return rep, fmt.Errorf("workflow reported %s", outcome.Terminal)
	}
	return rep, nil
}

// processLabel describes what happened to the Agent process, so a Gnomon-initiated termination
// after success is never worded as if it were an Agent failure.
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

// humanizeIdentity turns a hyphenated workflow identity into a sentence-leading label, e.g.
// "git-finalization" -> "Git finalization".
func humanizeIdentity(identity string) string {
	words := strings.Split(identity, "-")
	for i, w := range words {
		if i == 0 && w != "" {
			words[i] = strings.ToUpper(w[:1]) + w[1:]
		}
	}
	return strings.Join(words, " ")
}
