package orchestrate

import (
	"os"
	"testing"

	"gnomon/internal/adapter"
	"gnomon/internal/approval"
	"gnomon/internal/facts"
	"gnomon/internal/present"
)

// --- CLI -> Core workflow mapping ---

func TestSpecDiscover_MapsToSpecificationDiscovery(t *testing.T) {
	root := t.TempDir()
	configureGitIdentity(t, root)
	if _, err := Init(root); err != nil {
		t.Fatal(err)
	}

	l, wf, workflowPath, err := prepareWorkflow(root, "specification-discovery.md")
	if err != nil {
		t.Fatal(err)
	}
	if wf.Identity != "specification-discovery" {
		t.Fatalf("expected spec discover to map to identity specification-discovery, got %q", wf.Identity)
	}

	scripted := &payloadAdapter{
		FakeAdapter: &adapter.FakeAdapter{},
		Payload: map[string]interface{}{
			"outcome":         "CANDIDATE_PROPOSED",
			"candidate_title": "Password Reset",
			"rationale":       "Not yet represented by any existing Specification.",
			"derived_from":    "PROJECT.md Key Capabilities",
		},
	}
	report, err := runWorkflowWithAdapter(root, l, wf, workflowPath, targetNone, "", scripted)
	if err != nil {
		t.Fatalf("spec discover: %v (report: %+v)", err, report)
	}
	if report.Outcome != present.Success {
		t.Fatalf("expected Success, got %v", report.Outcome)
	}
	if report.Summary != "Candidate proposed" {
		t.Fatalf("unexpected summary: %q", report.Summary)
	}
}

func TestSpecDefine_MapsToSpecificationDefinition(t *testing.T) {
	root := t.TempDir()
	configureGitIdentity(t, root)
	if _, err := Init(root); err != nil {
		t.Fatal(err)
	}
	if _, err := SpecCreate(root, "Password Reset"); err != nil {
		t.Fatal(err)
	}

	l, wf, workflowPath, err := prepareWorkflow(root, "specification-definition.md")
	if err != nil {
		t.Fatal(err)
	}
	if wf.Identity != "specification-definition" {
		t.Fatalf("expected spec define to map to identity specification-definition, got %q", wf.Identity)
	}

	scripted := &payloadAdapter{
		FakeAdapter: &adapter.FakeAdapter{},
		Payload: map[string]interface{}{
			"outcome":           "READY_FOR_APPROVAL",
			"target":            "SPEC-001",
			"consistency_check": "CLEAN",
		},
	}
	report, err := runWorkflowWithAdapter(root, l, wf, workflowPath, targetSpec, "SPEC-001", scripted)
	if err != nil {
		t.Fatalf("spec define: %v (report: %+v)", err, report)
	}
	if report.Outcome != present.Success {
		t.Fatalf("expected Success, got %v", report.Outcome)
	}
	if report.Summary != "Ready for approval" {
		t.Fatalf("unexpected summary: %q", report.Summary)
	}
}

func TestTest_MapsToTestingWorkflow(t *testing.T) {
	root := t.TempDir()
	configureGitIdentity(t, root)
	if _, err := Init(root); err != nil {
		t.Fatal(err)
	}

	l, wf, workflowPath, err := prepareWorkflow(root, "testing.md")
	if err != nil {
		t.Fatal(err)
	}
	if wf.Identity != "testing" {
		t.Fatalf("expected test to map to identity testing, got %q", wf.Identity)
	}
	if wf.Identity == "verification" {
		t.Fatalf("Testing must remain distinct from Verification")
	}

	scripted := &payloadAdapter{
		FakeAdapter: &adapter.FakeAdapter{},
		Payload:     map[string]interface{}{"outcome": "TESTING_COMPLETE", "evidence": "go test ./... passed"},
	}
	report, err := runWorkflowWithAdapter(root, l, wf, workflowPath, targetSpec, "", scripted)
	if err != nil {
		t.Fatalf("test: %v (report: %+v)", err, report)
	}
	if report.Outcome != present.Success {
		t.Fatalf("expected Success, got %v", report.Outcome)
	}
	if report.Summary != "Testing complete" {
		t.Fatalf("unexpected summary: %q", report.Summary)
	}
}

// --- Eligibility gated before Agent resolution, using real, deterministic cases (unlike
// Slice 2, which needed a synthetic Contract — spec define/test genuinely have real gates) ---

func TestSpecDefine_NonexistentTargetBlocksBeforeAgentResolution(t *testing.T) {
	sandboxAgentConfig(t)
	root := t.TempDir()
	configureGitIdentity(t, root)
	if _, err := Init(root); err != nil {
		t.Fatal(err)
	}
	// SPEC-001 is never created.

	chooserCalled := false
	chooser := func(providers []string) (string, error) {
		chooserCalled = true
		return "claude", nil
	}
	report, err := SpecDefine(root, "SPEC-001", "", chooser)
	if err == nil {
		t.Fatalf("expected refusal for a nonexistent target Specification")
	}
	if report == nil || report.Outcome != present.Blocked {
		t.Fatalf("expected a Blocked report, got: %+v", report)
	}
	if chooserCalled {
		t.Fatalf("the Agent provider must never be resolved when the pre-start eligibility check fails")
	}
}

func TestTest_DraftSpecificationBlocksBeforeAgentResolution(t *testing.T) {
	sandboxAgentConfig(t)
	root := t.TempDir()
	configureGitIdentity(t, root)
	if _, err := Init(root); err != nil {
		t.Fatal(err)
	}
	if _, err := SpecCreate(root, "Password Reset"); err != nil {
		t.Fatal(err)
	}
	// SPEC-001 remains Draft — never approved.

	chooserCalled := false
	chooser := func(providers []string) (string, error) {
		chooserCalled = true
		return "claude", nil
	}
	report, err := Test(root, "SPEC-001", "", chooser)
	if err == nil {
		t.Fatalf("expected refusal: SPEC-governed Testing requires an Approved Specification")
	}
	if report == nil || report.Outcome != present.Blocked {
		t.Fatalf("expected a Blocked report, got: %+v", report)
	}
	if chooserCalled {
		t.Fatalf("the Agent provider must never be resolved when the pre-start eligibility check fails")
	}
}

func TestTest_NoTargetIsAlwaysEligible(t *testing.T) {
	sandboxAgentConfig(t)
	root := t.TempDir()
	configureGitIdentity(t, root)
	if _, err := Init(root); err != nil {
		t.Fatal(err)
	}

	// An invalid --agent override proves eligibility passed and Agent resolution was actually
	// reached (ad-hoc, unSpecified Testing per specification_reference: optional) — without
	// launching any real Agent process, since resolution fails before Prepare/Run are ever called.
	report, err := Test(root, "", "not-a-real-provider", nil)
	if err == nil {
		t.Fatalf("expected the invalid --agent override to be refused")
	}
	if report == nil || report.Outcome != present.Failed {
		t.Fatalf("expected a Failed (Agent-resolution) report, not an eligibility Blocked one, got: %+v", report)
	}
}

// --- Definition never silently approves ---

func TestSpecDefine_ReadyForApprovalNeverGrantsApproval(t *testing.T) {
	root := t.TempDir()
	configureGitIdentity(t, root)
	if _, err := Init(root); err != nil {
		t.Fatal(err)
	}
	if _, err := SpecCreate(root, "Password Reset"); err != nil {
		t.Fatal(err)
	}

	l, wf, workflowPath, err := prepareWorkflow(root, "specification-definition.md")
	if err != nil {
		t.Fatal(err)
	}
	scripted := &payloadAdapter{
		FakeAdapter: &adapter.FakeAdapter{},
		Payload:     map[string]interface{}{"outcome": "READY_FOR_APPROVAL", "target": "SPEC-001"},
	}
	report, err := runWorkflowWithAdapter(root, l, wf, workflowPath, targetSpec, "SPEC-001", scripted)
	if err != nil {
		t.Fatalf("spec define: %v (report: %+v)", err, report)
	}
	if report.Outcome != present.Success {
		t.Fatalf("expected Success, got %v", report.Outcome)
	}

	state, err := facts.Lifecycle(l, "SPEC-001")
	if err != nil {
		t.Fatal(err)
	}
	if state != approval.Draft {
		t.Fatalf("a READY_FOR_APPROVAL workflow result must never itself grant approval — expected Draft, got %v", state)
	}
}

// --- Discovery never mutates authoritative Specifications ---

func TestSpecDiscover_CandidateProposedNeverCreatesOrMutatesASpecification(t *testing.T) {
	root := t.TempDir()
	configureGitIdentity(t, root)
	if _, err := Init(root); err != nil {
		t.Fatal(err)
	}

	l, wf, workflowPath, err := prepareWorkflow(root, "specification-discovery.md")
	if err != nil {
		t.Fatal(err)
	}

	before, err := os.ReadDir(l.SpecificationsDir())
	if err != nil {
		t.Fatal(err)
	}

	scripted := &payloadAdapter{
		FakeAdapter: &adapter.FakeAdapter{},
		Payload: map[string]interface{}{
			"outcome":         "CANDIDATE_PROPOSED",
			"candidate_title": "Password Reset",
			"rationale":       "Not yet represented.",
			"derived_from":    "PROJECT.md",
		},
	}
	report, err := runWorkflowWithAdapter(root, l, wf, workflowPath, targetNone, "", scripted)
	if err != nil {
		t.Fatalf("spec discover: %v (report: %+v)", err, report)
	}
	if report.Outcome != present.Success {
		t.Fatalf("expected Success, got %v", report.Outcome)
	}

	after, err := os.ReadDir(l.SpecificationsDir())
	if err != nil {
		t.Fatal(err)
	}
	if len(before) != len(after) {
		t.Fatalf("expected specifications/ to be untouched by a CANDIDATE_PROPOSED result: before=%d files, after=%d files", len(before), len(after))
	}
	for i := range before {
		if before[i].Name() != after[i].Name() {
			t.Fatalf("expected specifications/ contents to be unchanged, got a difference: %q vs %q", before[i].Name(), after[i].Name())
		}
	}
}

// --- Valid BLOCKED results remain workflow results, never process/protocol failure ---

func TestTest_BlockedIsNotProcessFailure(t *testing.T) {
	root := t.TempDir()
	configureGitIdentity(t, root)
	if _, err := Init(root); err != nil {
		t.Fatal(err)
	}

	l, wf, workflowPath, err := prepareWorkflow(root, "testing.md")
	if err != nil {
		t.Fatal(err)
	}
	scripted := &payloadAdapter{
		FakeAdapter: &adapter.FakeAdapter{},
		Payload:     map[string]interface{}{"outcome": "BLOCKED", "remaining_unresolved": "a governing Specification is not Approved"},
	}
	report, err := runWorkflowWithAdapter(root, l, wf, workflowPath, targetSpec, "", scripted)
	if err == nil {
		t.Fatalf("expected a BLOCKED workflow conclusion to be reported as a non-nil error")
	}
	if report.Outcome != present.Blocked {
		t.Fatalf("expected Blocked, got %v", report.Outcome)
	}
	if report.Outcome == present.Failed {
		t.Fatalf("Testing BLOCKED must never be reported as a process/protocol Failed outcome")
	}
}
