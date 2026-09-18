package orchestrate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gnomon/internal/adapter"
	"gnomon/internal/agentconfig"
	"gnomon/internal/present"
	"gnomon/internal/project"
)

// --- CLI -> Core workflow mapping (each command loads the correct, real workflow file) ---

func TestDescribe_MapsToInitialKnowledgeEstablishment(t *testing.T) {
	root := t.TempDir()
	configureGitIdentity(t, root)
	if _, err := Init(root); err != nil {
		t.Fatal(err)
	}

	l, wf, workflowPath, err := prepareWorkflow(root, "initial-knowledge-establishment.md")
	if err != nil {
		t.Fatal(err)
	}
	if wf.Identity != "initial-knowledge-establishment" {
		t.Fatalf("expected describe to map to identity initial-knowledge-establishment, got %q", wf.Identity)
	}

	scripted := &payloadAdapter{
		FakeAdapter: &adapter.FakeAdapter{},
		Payload:     map[string]interface{}{"outcome": "READY_FOR_BOOTSTRAP", "recorded_knowledge": "PROJECT.md updated"},
	}
	report, err := runWorkflowWithAdapter(root, l, wf, workflowPath, targetNone, "", scripted)
	if err != nil {
		t.Fatalf("describe: %v (report: %+v)", err, report)
	}
	if report.Outcome != present.Success {
		t.Fatalf("expected Success, got %v", report.Outcome)
	}
	if report.Summary != "Ready for bootstrap" {
		t.Fatalf("unexpected summary: %q", report.Summary)
	}
}

// --- Onboarding guidance (Init -> describe -> bootstrap -> Specification work) ---

func TestOnboardingNext_SetsNextOnlyOnSuccess(t *testing.T) {
	success := &present.Report{Outcome: present.Success}
	onboardingNext(success, "gnomon bootstrap")
	if success.Next != "gnomon bootstrap" {
		t.Fatalf("expected Next to be set on a Success report, got %q", success.Next)
	}

	for _, outcome := range []present.Outcome{present.Blocked, present.Cancelled, present.Failed} {
		rep := &present.Report{Outcome: outcome}
		onboardingNext(rep, "gnomon bootstrap")
		if rep.Next != "" {
			t.Fatalf("expected Next to remain empty for Outcome %v, got %q", outcome, rep.Next)
		}
	}

	// A nil report (the shape prepareWorkflow's own early-return errors produce) must never panic.
	onboardingNext(nil, "gnomon bootstrap")
}

func TestInit_RecommendsDescribeAsFirstStep(t *testing.T) {
	root := t.TempDir()
	configureGitIdentity(t, root)

	report, err := Init(root)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(report.Next, "gnomon describe") {
		t.Fatalf("expected Init's Next to recommend describe, got %q", report.Next)
	}
	if strings.Contains(report.Next, "spec create") {
		t.Fatalf("did not expect Init's Next to list spec create as an alternative, got %q", report.Next)
	}
}

func TestDescribe_FailedAgentResolution_NeverGetsSuccessNext(t *testing.T) {
	sandboxAgentConfig(t)
	root := t.TempDir()
	configureGitIdentity(t, root)
	if _, err := Init(root); err != nil {
		t.Fatal(err)
	}

	report, err := Describe(root, "not-a-real-provider", nil)
	if err == nil {
		t.Fatalf("expected an invalid --agent override to be refused")
	}
	if report == nil || report.Outcome != present.Failed {
		t.Fatalf("expected a Failed report, got: %+v", report)
	}
	if report.Next == "gnomon bootstrap" {
		t.Fatalf("did not expect the success-only 'gnomon bootstrap' suggestion on a Failed result")
	}
}

func TestBootstrap_FailedAgentResolution_NeverGetsSuccessNext(t *testing.T) {
	sandboxAgentConfig(t)
	root := t.TempDir()
	configureGitIdentity(t, root)
	if _, err := Init(root); err != nil {
		t.Fatal(err)
	}

	report, err := Bootstrap(root, "not-a-real-provider", nil)
	if err == nil {
		t.Fatalf("expected an invalid --agent override to be refused")
	}
	if report == nil || report.Outcome != present.Failed {
		t.Fatalf("expected a Failed report, got: %+v", report)
	}
	if strings.Contains(report.Next, "spec create") {
		t.Fatalf("did not expect the success-only spec-create suggestion on a Failed result, got %q", report.Next)
	}
}

func TestDescribe_SuccessfulResult_RecommendsBootstrap(t *testing.T) {
	root := t.TempDir()
	configureGitIdentity(t, root)
	if _, err := Init(root); err != nil {
		t.Fatal(err)
	}
	l, wf, workflowPath, err := prepareWorkflow(root, "initial-knowledge-establishment.md")
	if err != nil {
		t.Fatal(err)
	}
	scripted := &payloadAdapter{
		FakeAdapter: &adapter.FakeAdapter{},
		Payload:     map[string]interface{}{"outcome": "READY_FOR_BOOTSTRAP", "recorded_knowledge": "PROJECT.md updated"},
	}
	report, err := runWorkflowWithAdapter(root, l, wf, workflowPath, targetNone, "", scripted)
	if err != nil {
		t.Fatalf("describe: %v (report: %+v)", err, report)
	}
	onboardingNext(report, "gnomon bootstrap")
	if report.Next != "gnomon bootstrap" {
		t.Fatalf("expected a successful describe result to recommend gnomon bootstrap, got %q", report.Next)
	}
}

func TestBootstrap_SuccessfulResult_RecommendsSpecWork(t *testing.T) {
	root := t.TempDir()
	configureGitIdentity(t, root)
	if _, err := Init(root); err != nil {
		t.Fatal(err)
	}
	l, wf, workflowPath, err := prepareWorkflow(root, "bootstrap.md")
	if err != nil {
		t.Fatal(err)
	}
	scripted := &payloadAdapter{
		FakeAdapter: &adapter.FakeAdapter{},
		Payload:     map[string]interface{}{"outcome": "BOOTSTRAP_COMPLETE", "baseline_established": "scaffolded"},
	}
	report, err := runWorkflowWithAdapter(root, l, wf, workflowPath, targetNone, "", scripted)
	if err != nil {
		t.Fatalf("bootstrap: %v (report: %+v)", err, report)
	}
	onboardingNext(report, `gnomon spec create "<title>", or gnomon spec discover to have an Agent propose one`)
	if !strings.Contains(report.Next, "gnomon spec create") || !strings.Contains(report.Next, "gnomon spec discover") {
		t.Fatalf("expected a successful bootstrap result to recommend both ways to start Specification work, got %q", report.Next)
	}
}

func TestBootstrap_MapsToBootstrapWorkflow(t *testing.T) {
	root := t.TempDir()
	configureGitIdentity(t, root)
	if _, err := Init(root); err != nil {
		t.Fatal(err)
	}

	l, wf, workflowPath, err := prepareWorkflow(root, "bootstrap.md")
	if err != nil {
		t.Fatal(err)
	}
	if wf.Identity != "bootstrap" {
		t.Fatalf("expected bootstrap to map to identity bootstrap, got %q", wf.Identity)
	}

	scripted := &payloadAdapter{
		FakeAdapter: &adapter.FakeAdapter{},
		Payload:     map[string]interface{}{"outcome": "BOOTSTRAP_COMPLETE"},
	}
	report, err := runWorkflowWithAdapter(root, l, wf, workflowPath, targetNone, "", scripted)
	if err != nil {
		t.Fatalf("bootstrap: %v (report: %+v)", err, report)
	}
	if report.Outcome != present.Success {
		t.Fatalf("expected Success, got %v", report.Outcome)
	}
	if report.Summary != "Bootstrap complete" {
		t.Fatalf("unexpected summary: %q", report.Summary)
	}
}

func TestFinalize_MapsToGitFinalizationWorkflow(t *testing.T) {
	root := t.TempDir()
	configureGitIdentity(t, root)
	if _, err := Init(root); err != nil {
		t.Fatal(err)
	}

	l, wf, workflowPath, err := prepareWorkflow(root, "git-finalization.md")
	if err != nil {
		t.Fatal(err)
	}
	if wf.Identity != "git-finalization" {
		t.Fatalf("expected finalize to map to identity git-finalization, got %q", wf.Identity)
	}

	scripted := &payloadAdapter{
		FakeAdapter: &adapter.FakeAdapter{},
		Payload:     map[string]interface{}{"outcome": "COMMIT_PREPARED", "included": "1 file"},
	}
	report, err := runWorkflowWithAdapter(root, l, wf, workflowPath, targetNone, "", scripted)
	if err != nil {
		t.Fatalf("finalize: %v (report: %+v)", err, report)
	}
	if report.Outcome != present.Success {
		t.Fatalf("expected Success, got %v", report.Outcome)
	}
	if report.Summary != "Commit prepared" {
		t.Fatalf("unexpected summary: %q", report.Summary)
	}
}

func TestVerify_MapsToVerificationWorkflow_NestedResultContract(t *testing.T) {
	root := t.TempDir()
	configureGitIdentity(t, root)
	if _, err := Init(root); err != nil {
		t.Fatal(err)
	}

	l, wf, workflowPath, err := prepareWorkflow(root, "verification.md")
	if err != nil {
		t.Fatal(err)
	}
	if wf.Identity != "verification" {
		t.Fatalf("expected verify to map to identity verification, got %q", wf.Identity)
	}
	if wf.Result.TerminalPath != "summary.aggregate" {
		t.Fatalf("expected the nested Result Contract established in Slice 1 to be preserved, got terminal_path=%q", wf.Result.TerminalPath)
	}

	scripted := &payloadAdapter{
		FakeAdapter: &adapter.FakeAdapter{},
		Payload: map[string]interface{}{
			"evidence": []interface{}{
				map[string]interface{}{"obligation": "All tests pass", "result": "PASS"},
			},
			"summary": map[string]interface{}{"aggregate": "PASS", "obligations_evaluated": "1 of 1"},
		},
	}
	report, err := runWorkflowWithAdapter(root, l, wf, workflowPath, targetGeneric, "src/auth/", scripted)
	if err != nil {
		t.Fatalf("verify: %v (report: %+v)", err, report)
	}
	if report.Outcome != present.Success {
		t.Fatalf("expected Success, got %v", report.Outcome)
	}
	if report.Target != "src/auth/" {
		t.Fatalf("expected the free-form target to be reported, got %q", report.Target)
	}
}

func TestReview_MapsToReviewWorkflow_NestedResultContract(t *testing.T) {
	root := t.TempDir()
	configureGitIdentity(t, root)
	if _, err := Init(root); err != nil {
		t.Fatal(err)
	}

	l, wf, workflowPath, err := prepareWorkflow(root, "review.md")
	if err != nil {
		t.Fatal(err)
	}
	if wf.Identity != "review" {
		t.Fatalf("expected review to map to identity review, got %q", wf.Identity)
	}
	if wf.Result.TerminalPath != "summary.aggregate" {
		t.Fatalf("expected the nested Result Contract established in Slice 1 to be preserved, got terminal_path=%q", wf.Result.TerminalPath)
	}

	scripted := &payloadAdapter{
		FakeAdapter: &adapter.FakeAdapter{},
		Payload: map[string]interface{}{
			"findings": []interface{}{},
			"summary":  map[string]interface{}{"aggregate": "PASS", "findings": "0 findings"},
		},
	}
	report, err := runWorkflowWithAdapter(root, l, wf, workflowPath, targetGeneric, "", scripted)
	if err != nil {
		t.Fatalf("review: %v (report: %+v)", err, report)
	}
	if report.Outcome != present.Success {
		t.Fatalf("expected Success, got %v", report.Outcome)
	}
}

// --- Negative/blocked workflow conclusions are never process/protocol failure ---

func TestVerify_FAILIsBlockedNeverProcessFailure(t *testing.T) {
	root := t.TempDir()
	configureGitIdentity(t, root)
	if _, err := Init(root); err != nil {
		t.Fatal(err)
	}
	l, wf, workflowPath, err := prepareWorkflow(root, "verification.md")
	if err != nil {
		t.Fatal(err)
	}

	scripted := &payloadAdapter{
		FakeAdapter: &adapter.FakeAdapter{},
		Payload: map[string]interface{}{
			"evidence": []interface{}{
				map[string]interface{}{"obligation": "All tests pass", "result": "FAIL", "evidence": "TestFoo failed"},
			},
			"summary": map[string]interface{}{"aggregate": "FAIL", "obligations_evaluated": "1 of 1"},
		},
	}
	report, err := runWorkflowWithAdapter(root, l, wf, workflowPath, targetGeneric, "src/", scripted)
	if err == nil {
		t.Fatalf("expected a BLOCKED workflow conclusion to be reported as a non-nil error")
	}
	if report.Outcome != present.Blocked {
		t.Fatalf("expected Verification FAIL to classify as Blocked, got %v", report.Outcome)
	}
	if report.Outcome == present.Failed {
		t.Fatalf("Verification FAIL must never be reported as a process/protocol Failed outcome")
	}
}

func TestVerify_UNVERIFIABLEIsBlockedNeverProcessFailure(t *testing.T) {
	root := t.TempDir()
	configureGitIdentity(t, root)
	if _, err := Init(root); err != nil {
		t.Fatal(err)
	}
	l, wf, workflowPath, err := prepareWorkflow(root, "verification.md")
	if err != nil {
		t.Fatal(err)
	}

	scripted := &payloadAdapter{
		FakeAdapter: &adapter.FakeAdapter{},
		Payload: map[string]interface{}{
			"summary": map[string]interface{}{"aggregate": "UNVERIFIABLE"},
		},
	}
	report, err := runWorkflowWithAdapter(root, l, wf, workflowPath, targetGeneric, "", scripted)
	if err == nil {
		t.Fatalf("expected a BLOCKED workflow conclusion to be reported as a non-nil error")
	}
	if report.Outcome != present.Blocked {
		t.Fatalf("expected Verification UNVERIFIABLE to classify as Blocked, got %v", report.Outcome)
	}
}

func TestReview_DefectRiskKnowledgeGapAreBlockedNeverProcessFailure(t *testing.T) {
	root := t.TempDir()
	configureGitIdentity(t, root)
	if _, err := Init(root); err != nil {
		t.Fatal(err)
	}
	l, wf, workflowPath, err := prepareWorkflow(root, "review.md")
	if err != nil {
		t.Fatal(err)
	}

	for _, aggregate := range []string{"DEFECT", "RISK", "KNOWLEDGE GAP"} {
		t.Run(aggregate, func(t *testing.T) {
			scripted := &payloadAdapter{
				FakeAdapter: &adapter.FakeAdapter{},
				Payload: map[string]interface{}{
					"findings": []interface{}{
						map[string]interface{}{"finding_id": "F1", "classification": aggregate},
					},
					"summary": map[string]interface{}{"aggregate": aggregate},
				},
			}
			report, err := runWorkflowWithAdapter(root, l, wf, workflowPath, targetGeneric, "", scripted)
			if err == nil {
				t.Fatalf("expected a BLOCKED workflow conclusion (%s) to be reported as a non-nil error", aggregate)
			}
			if report.Outcome != present.Blocked {
				t.Fatalf("expected Review %s to classify as Blocked, got %v", aggregate, report.Outcome)
			}
			if report.Outcome == present.Failed {
				t.Fatalf("Review %s must never be reported as a process/protocol Failed outcome", aggregate)
			}
		})
	}
}

// --- Eligibility gating is generic, and checked before any Agent is resolved ---

// TestRunTargetedWorkflow_IneligibleNeverResolvesAgent proves the mechanism shared by describe,
// bootstrap, verify, review, and finalize (runTargetedWorkflow) refuses before ever resolving an
// Agent provider for any workflow whose Contract declares a real pre-start gate — using a
// synthetic Contract, since none of the five real, built-in workflows this slice wires up
// currently declare one (all five are specification_reference: none, so they are always eligible
// today; the mechanism itself remains generic and gate-checked regardless).
func TestRunTargetedWorkflow_IneligibleNeverResolvesAgent(t *testing.T) {
	root := t.TempDir()
	configureGitIdentity(t, root)
	if _, err := Init(root); err != nil {
		t.Fatal(err)
	}
	l := project.Layout{Root: root}
	synthetic := "---\nidentity: synthetic-gated\nspecification_reference: required\nresult:\n" +
		"  terminal_path: outcome\n  classification:\n    OK: success\n  schema:\n    type: object\n" +
		"    required: [outcome]\n    properties:\n      outcome:\n        type: string\n        enum: [OK]\n" +
		"---\n# Synthetic\n"
	if err := os.WriteFile(filepath.Join(l.WorkflowsDir(), "synthetic-gated.md"), []byte(synthetic), 0o644); err != nil {
		t.Fatal(err)
	}

	chooserCalled := false
	chooser := func(providers []string) (string, error) {
		chooserCalled = true
		return "claude", nil
	}
	report, err := runNoTargetWorkflow(root, "synthetic-gated.md", "", chooser)
	if err == nil {
		t.Fatalf("expected refusal: no Specification supplied for a required-reference workflow")
	}
	if report == nil || report.Outcome != present.Blocked {
		t.Fatalf("expected a Blocked report, got: %+v", report)
	}
	if chooserCalled {
		t.Fatalf("the Agent provider must never be resolved when the pre-start eligibility check fails")
	}
}

// --- Agent override threading (each new command, not just Implement) ---

// TestBootstrap_AgentOverrideIsUsedAndNeverPersisted proves a newly-wired command threads its
// --agent override through to ResolveAgent (an invalid value is refused immediately, before any
// real Agent process is ever touched, matching how Slice 0's own tests prove this for Implement)
// and never persists it as the new saved default.
func TestBootstrap_AgentOverrideIsUsedAndNeverPersisted(t *testing.T) {
	sandboxAgentConfig(t)
	root := t.TempDir()
	configureGitIdentity(t, root)
	if _, err := Init(root); err != nil {
		t.Fatal(err)
	}

	report, err := Bootstrap(root, "not-a-real-provider", func(providers []string) (string, error) {
		t.Fatalf("chooser must not be called when --agent is explicit")
		return "", nil
	})
	if err == nil {
		t.Fatalf("expected an invalid --agent override to be refused")
	}
	if report == nil || report.Outcome != present.Failed {
		t.Fatalf("expected a Failed report for an unresolvable Agent provider, got: %+v", report)
	}

	cfg, loadErr := agentconfig.Load()
	if loadErr != nil {
		t.Fatal(loadErr)
	}
	if cfg.DefaultProvider != "" {
		t.Fatalf("an --agent override must never persist a default, got %q", cfg.DefaultProvider)
	}
}
