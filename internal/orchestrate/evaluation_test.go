package orchestrate

import (
	"reflect"
	"testing"

	"gnomon/internal/adapter"
	"gnomon/internal/present"
)

// --- ActionableFindings: Verification ---

func TestActionableFindings_Verification_FiltersOutPass_AssignsReportLocalIDs(t *testing.T) {
	payload := map[string]interface{}{
		"evidence": []interface{}{
			map[string]interface{}{"obligation": "Refund cancels the charge", "result": "FAIL", "evidence": "charge.status remained pending", "recommended_workflow": "implementation"},
			map[string]interface{}{"obligation": "Refund amount matches", "result": "PASS", "evidence": "matched"},
			map[string]interface{}{"obligation": "Idempotent retries", "result": "UNVERIFIABLE", "evidence": "no test harness", "recommended_workflow": "testing"},
		},
		"summary": map[string]interface{}{"aggregate": "FAIL"},
	}
	got := ActionableFindings("verification", payload)
	want := []EvaluationFinding{
		{ID: "F-001", Classification: "FAIL", Headline: "Refund cancels the charge", Detail: "charge.status remained pending", RecommendedWorkflow: "implementation"},
		{ID: "F-002", Classification: "UNVERIFIABLE", Headline: "Idempotent retries", Detail: "no test harness", RecommendedWorkflow: "testing"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

// TestActionableFindings_Verification_SameResult_DifferentWorkflows proves the extraction layer
// carries through whatever the Agent recommended per-obligation rather than deriving anything
// from the shared "FAIL" result — two obligations sharing a result may legitimately warrant
// different workflows, decided by the Agent from cause, never by CLI/orchestration code.
func TestActionableFindings_Verification_SameResult_DifferentWorkflows(t *testing.T) {
	payload := map[string]interface{}{
		"evidence": []interface{}{
			map[string]interface{}{"obligation": "A", "result": "FAIL", "recommended_workflow": "implementation"},
			map[string]interface{}{"obligation": "B", "result": "FAIL", "recommended_workflow": "testing"},
			map[string]interface{}{"obligation": "C", "result": "FAIL", "recommended_workflow": "knowledge-resolution"},
		},
	}
	got := ActionableFindings("verification", payload)
	if len(got) != 3 {
		t.Fatalf("expected 3 findings, got %d: %+v", len(got), got)
	}
	want := []string{"implementation", "testing", "knowledge-resolution"}
	for i, w := range want {
		if got[i].Classification != "FAIL" {
			t.Fatalf("expected every finding to share classification FAIL, got %+v", got[i])
		}
		if got[i].RecommendedWorkflow != w {
			t.Fatalf("expected finding %d to carry recommendation %q despite the shared FAIL result, got %+v", i, w, got[i])
		}
	}
}

// TestActionableFindings_Verification_UnrecognizedRecommendation_DroppedNotTrusted proves an
// out-of-vocabulary recommended_workflow value (which the Result Contract's own enum should
// already reject at Result Protocol validation time) is defensively dropped here too, rather than
// ever reaching EvaluationFinding as something that looks dispatchable.
func TestActionableFindings_Verification_UnrecognizedRecommendation_DroppedNotTrusted(t *testing.T) {
	payload := map[string]interface{}{
		"evidence": []interface{}{
			map[string]interface{}{"obligation": "X", "result": "FAIL", "recommended_workflow": "rm -rf /"},
		},
	}
	got := ActionableFindings("verification", payload)
	if len(got) != 1 || got[0].RecommendedWorkflow != "" {
		t.Fatalf("expected an unrecognized recommendation dropped to \"\", got %+v", got)
	}
}

// TestActionableFindings_Verification_HumanDecisionValue_Rejected proves the removed
// "human-decision" value is no longer trusted even if an Agent (old prompt caching, a bug, or a
// malicious payload) still emits it — it is treated exactly like any other unrecognized string.
func TestActionableFindings_Verification_HumanDecisionValue_Rejected(t *testing.T) {
	payload := map[string]interface{}{
		"evidence": []interface{}{
			map[string]interface{}{"obligation": "X", "result": "FAIL", "recommended_workflow": "human-decision"},
		},
	}
	got := ActionableFindings("verification", payload)
	if len(got) != 1 || got[0].RecommendedWorkflow != "" {
		t.Fatalf("expected \"human-decision\" rejected as not dispatchable, got %+v", got)
	}
}

func TestActionableFindings_Verification_AllPass_ReturnsNil(t *testing.T) {
	payload := map[string]interface{}{
		"evidence": []interface{}{
			map[string]interface{}{"obligation": "A", "result": "PASS"},
			map[string]interface{}{"obligation": "B", "result": "PASS"},
		},
		"summary": map[string]interface{}{"aggregate": "PASS"},
	}
	if got := ActionableFindings("verification", payload); got != nil {
		t.Fatalf("expected no actionable findings for an all-PASS result, got %+v", got)
	}
}

func TestActionableFindings_Verification_NullEvidence_ReturnsNil(t *testing.T) {
	payload := map[string]interface{}{"evidence": nil, "summary": map[string]interface{}{"aggregate": "PASS"}}
	if got := ActionableFindings("verification", payload); got != nil {
		t.Fatalf("expected nil for null evidence, got %+v", got)
	}
}

// --- ActionableFindings: Review ---

func TestActionableFindings_Review_PassesThroughRealIDsAndFields(t *testing.T) {
	payload := map[string]interface{}{
		"findings": []interface{}{
			map[string]interface{}{
				"finding_id": "F-001", "classification": "DEFECT", "summary": "Missing rollback",
				"evidence": "see test output", "engineering_reasoning": "no compensating action",
				"impact": "data inconsistency", "resolution_owner": "Implementation",
				"decision_required": false, "recommended_workflow": "implementation",
			},
			map[string]interface{}{
				"finding_id": "F-002", "classification": "KNOWLEDGE GAP", "summary": "No refund policy",
				"resolution_owner": "Domain", "decision_required": true,
				"recommended_workflow": "knowledge-resolution",
			},
			map[string]interface{}{
				"finding_id": "F-003", "classification": "RISK", "summary": "Retried without idempotency key",
				"recommended_workflow": "knowledge-resolution", "decision_required": true,
			},
		},
		"summary": map[string]interface{}{"aggregate": "DEFECT"},
	}
	got := ActionableFindings("review", payload)
	if len(got) != 3 {
		t.Fatalf("expected 3 findings, got %d: %+v", len(got), got)
	}
	if got[0].ID != "F-001" || got[0].Classification != "DEFECT" || got[0].Headline != "Missing rollback" {
		t.Fatalf("unexpected first finding: %+v", got[0])
	}
	if got[0].Detail != "see test output\n\nno compensating action\n\nImpact: data inconsistency" {
		t.Fatalf("expected evidence+reasoning+impact joined, got %q", got[0].Detail)
	}
	if got[0].ResolutionOwner != "Implementation" || got[0].RecommendedWorkflow != "implementation" {
		t.Fatalf("expected resolution owner/recommendation carried through, got %+v", got[0])
	}
	if got[1].ID != "F-002" || got[1].Classification != "KNOWLEDGE GAP" || !got[1].DecisionRequired {
		t.Fatalf("unexpected second finding: %+v", got[1])
	}
	// F-003 is a RISK recommending knowledge-resolution with decision_required — proves
	// classification (RISK) does not determine the workflow, and decision_required is preserved
	// independently of it rather than being folded into a separate routing value.
	if got[2].Classification != "RISK" || got[2].RecommendedWorkflow != "knowledge-resolution" || !got[2].DecisionRequired {
		t.Fatalf("expected RISK + knowledge-resolution + decision_required all independently preserved, got %+v", got[2])
	}
}

// TestActionableFindings_Review_SameClassification_DifferentWorkflows proves classification alone
// never determines the recommended workflow — two DEFECT findings may legitimately recommend
// different workflows depending on their actual cause, and extraction must not collapse that.
func TestActionableFindings_Review_SameClassification_DifferentWorkflows(t *testing.T) {
	payload := map[string]interface{}{
		"findings": []interface{}{
			map[string]interface{}{"finding_id": "F-001", "classification": "DEFECT", "recommended_workflow": "implementation"},
			map[string]interface{}{"finding_id": "F-002", "classification": "DEFECT", "recommended_workflow": "knowledge-resolution"},
		},
	}
	got := ActionableFindings("review", payload)
	if len(got) != 2 || got[0].RecommendedWorkflow == got[1].RecommendedWorkflow {
		t.Fatalf("expected two DEFECT findings to carry independently different recommendations, got %+v", got)
	}
}

// TestActionableFindings_Review_UnrecognizedRecommendation_DroppedNotTrusted mirrors the
// Verification case: defense in depth beyond the Result Contract's own enum.
func TestActionableFindings_Review_UnrecognizedRecommendation_DroppedNotTrusted(t *testing.T) {
	payload := map[string]interface{}{
		"findings": []interface{}{
			map[string]interface{}{"finding_id": "F-001", "classification": "RISK", "recommended_workflow": "just fix it somehow"},
		},
	}
	got := ActionableFindings("review", payload)
	if len(got) != 1 || got[0].RecommendedWorkflow != "" {
		t.Fatalf("expected an unrecognized recommendation dropped to \"\", got %+v", got)
	}
}

// TestActionableFindings_Review_HumanDecisionValue_Rejected proves "human-decision" — the removed
// routing value — is never treated as dispatchable even if it still appears in a payload.
func TestActionableFindings_Review_HumanDecisionValue_Rejected(t *testing.T) {
	payload := map[string]interface{}{
		"findings": []interface{}{
			map[string]interface{}{"finding_id": "F-001", "classification": "RISK", "recommended_workflow": "human-decision"},
		},
	}
	got := ActionableFindings("review", payload)
	if len(got) != 1 || got[0].RecommendedWorkflow != "" {
		t.Fatalf("expected \"human-decision\" rejected as not dispatchable, got %+v", got)
	}
}

// --- IsDispatchableResolution ---

func TestIsDispatchableResolution(t *testing.T) {
	for _, approach := range []string{"implementation", "testing", "knowledge-resolution"} {
		if !IsDispatchableResolution(approach) {
			t.Errorf("expected %q to be dispatchable", approach)
		}
	}
	for _, approach := range []string{"human-decision", "", "delete everything", "Implementation"} {
		if IsDispatchableResolution(approach) {
			t.Errorf("expected %q to NOT be dispatchable", approach)
		}
	}
}

func TestActionableFindings_Review_NullFindings_ReturnsNil(t *testing.T) {
	payload := map[string]interface{}{"findings": nil, "summary": map[string]interface{}{"aggregate": "PASS"}}
	if got := ActionableFindings("review", payload); got != nil {
		t.Fatalf("expected nil for null findings (a PASS result), got %+v", got)
	}
}

func TestActionableFindings_UnknownIdentity_ReturnsNil(t *testing.T) {
	if got := ActionableFindings("bootstrap", map[string]interface{}{"outcome": "BOOTSTRAP_COMPLETE"}); got != nil {
		t.Fatalf("expected nil for a non-evaluation workflow, got %+v", got)
	}
}

// --- RunEvaluation: integration with the real Contract-driven pipeline ---

func TestRunEvaluation_VerificationFail_ReturnsFindingsAlongsideReport(t *testing.T) {
	root := setupVerificationTarget(t)
	sandboxAgentConfig(t)
	scripted := &payloadAdapter{
		FakeAdapter: &adapter.FakeAdapter{},
		Payload: map[string]interface{}{
			"evidence": []interface{}{
				map[string]interface{}{"obligation": "Refund cancels the charge", "result": "FAIL", "evidence": "still pending"},
			},
			"summary": map[string]interface{}{"aggregate": "FAIL"},
		},
	}
	l, wf, workflowPath, err := prepareWorkflow(root, "verification.md")
	if err != nil {
		t.Fatal(err)
	}
	rep, outcome, err := runWorkflowWithAdapterAndOutcome(root, l, wf, workflowPath, targetGeneric, "src/billing/", scripted)
	if err == nil {
		t.Fatalf("expected FAIL to classify as blocked (non-nil error)")
	}
	if rep == nil {
		t.Fatalf("expected a Report to still be produced for a Blocked classification")
	}
	findings := ActionableFindings(wf.Identity, outcome.Payload)
	if len(findings) != 1 || findings[0].ID != "F-001" || findings[0].Classification != "FAIL" {
		t.Fatalf("expected exactly one FAIL finding with synthetic ID F-001, got %+v", findings)
	}
}

func TestRunEvaluation_ReviewPass_NoFindings(t *testing.T) {
	root := setupVerificationTarget(t)
	scripted := &payloadAdapter{
		FakeAdapter: &adapter.FakeAdapter{},
		Payload: map[string]interface{}{
			"findings": nil,
			"summary":  map[string]interface{}{"aggregate": "PASS"},
		},
	}
	l, wf, workflowPath, err := prepareWorkflow(root, "review.md")
	if err != nil {
		t.Fatal(err)
	}
	rep, outcome, err := runWorkflowWithAdapterAndOutcome(root, l, wf, workflowPath, targetGeneric, "src/billing/", scripted)
	if err != nil {
		t.Fatalf("expected PASS to succeed, got %v (report: %+v)", err, rep)
	}
	if findings := ActionableFindings(wf.Identity, outcome.Payload); findings != nil {
		t.Fatalf("expected no findings for a PASS review, got %+v", findings)
	}
}

func TestRun_StillReturnsOnlyTwoValues_UnaffectedByRunEvaluation(t *testing.T) {
	// Regression guard: Run's own exported signature and behavior must be byte-for-byte
	// unchanged now that it delegates to RunEvaluation internally.
	root := setupVerificationTarget(t)
	sandboxAgentConfig(t)
	report, err := Run(root, "verification", "src/billing/", "not-a-real-provider", nil)
	if err == nil {
		t.Fatalf("expected an invalid --agent override to be refused")
	}
	if report == nil {
		t.Fatalf("expected a Report even on Agent-resolution failure")
	}
}

// --- Resolution handoff: reaches adapter.Context, never bypasses eligibility ---

func TestObtainWorkflowOutcomeFull_HandoffReachesAdapterContext(t *testing.T) {
	root := setupVerificationTarget(t)
	scripted := &payloadAdapter{
		FakeAdapter: &adapter.FakeAdapter{},
		Payload:     map[string]interface{}{"outcome": "IMPLEMENTATION_COMPLETE"},
	}
	l, wf, workflowPath, err := prepareWorkflow(root, "implementation.md")
	if err != nil {
		t.Fatal(err)
	}
	handoff := ResolutionHandoff{
		OriginWorkflow: "review", OriginTarget: "src/billing/",
		FindingID: "F-001", Classification: "DEFECT",
		Summary:  "Refund does not release inventory reservation.",
		Evidence: "Reservation remains active after refund completion.",
	}
	if _, _, _, err := obtainWorkflowOutcomeFull(root, l, wf, workflowPath, targetSpec, "SPEC-001", handoff, scripted); err != nil {
		t.Fatal(err)
	}
	got := scripted.FakeAdapter.Ctx.Handoff
	if got == nil {
		t.Fatalf("expected Context.Handoff to be populated")
	}
	if got.OriginWorkflow != "review" || got.FindingID != "F-001" || got.Summary != handoff.Summary || got.Evidence != handoff.Evidence {
		t.Fatalf("unexpected handoff reaching the Agent's own Context: %+v", got)
	}
	// The handoff never replaces the workflow's own normal target/eligibility handling.
	if scripted.FakeAdapter.Ctx.SpecIdentity != "SPEC-001" {
		t.Fatalf("expected SpecIdentity to remain the real Implementation target, got %q", scripted.FakeAdapter.Ctx.SpecIdentity)
	}
}

// TestObtainWorkflowOutcomeFull_KnowledgeResolutionHandoff_CarriesFindingOnly proves a
// knowledge-resolution dispatch's handoff carries only the Finding itself — no CLI-collected
// Human decision field exists any more; the launched Agent asks the Human directly, in its own
// session, for whatever decision it needs.
func TestObtainWorkflowOutcomeFull_KnowledgeResolutionHandoff_CarriesFindingOnly(t *testing.T) {
	root := setupVerificationTarget(t)
	scripted := &payloadAdapter{
		FakeAdapter: &adapter.FakeAdapter{},
		Payload:     map[string]interface{}{"outcome": "RESOLVED"},
	}
	l, wf, workflowPath, err := prepareWorkflow(root, "knowledge-resolution.md")
	if err != nil {
		t.Fatal(err)
	}
	handoff := ResolutionHandoff{
		OriginWorkflow: "review", OriginTarget: "src/billing/",
		FindingID: "F-002", Classification: "RISK",
		Summary: "Retried without idempotency key.",
	}
	if _, _, _, err := obtainWorkflowOutcomeFull(root, l, wf, workflowPath, targetGeneric, "", handoff, scripted); err != nil {
		t.Fatal(err)
	}
	got := scripted.FakeAdapter.Ctx.Handoff
	if got == nil || got.FindingID != "F-002" || got.Summary != handoff.Summary {
		t.Fatalf("expected the Finding to reach the Agent's own Context, got %+v", got)
	}
}

func TestObtainWorkflowOutcome_ZeroValueHandoff_NoHandoffOnContext(t *testing.T) {
	root := setupVerificationTarget(t)
	scripted := &payloadAdapter{
		FakeAdapter: &adapter.FakeAdapter{},
		Payload:     map[string]interface{}{"outcome": "IMPLEMENTATION_COMPLETE"},
	}
	l, wf, workflowPath, err := prepareWorkflow(root, "implementation.md")
	if err != nil {
		t.Fatal(err)
	}
	if _, _, _, err := obtainWorkflowOutcome(root, l, wf, workflowPath, targetSpec, "SPEC-001", scripted); err != nil {
		t.Fatal(err)
	}
	if scripted.FakeAdapter.Ctx.Handoff != nil {
		t.Fatalf("expected no Handoff for an ordinary (non-resolution) invocation, got %+v", scripted.FakeAdapter.Ctx.Handoff)
	}
}

// TestRunResolutionWorkflow_ImplementationRequiresApprovedSpecification proves the handoff never
// bypasses Implement's own Approved-Specification gate — this fails at eligibility, before
// ResolveAgent (and therefore any real Agent) is ever reached.
func TestRunResolutionWorkflow_ImplementationRequiresApprovedSpecification(t *testing.T) {
	root := setupVerificationTarget(t)
	handoff := ResolutionHandoff{OriginWorkflow: "review", OriginTarget: "src/billing/", FindingID: "F-001", Classification: "DEFECT"}

	rep, err := RunResolutionWorkflow(root, ResolveWithImplementation, "SPEC-999", handoff, "", nil)
	if err == nil {
		t.Fatalf("expected a nonexistent Specification to be refused")
	}
	if rep == nil || rep.Outcome != present.Blocked {
		t.Fatalf("expected a Blocked report, got %+v", rep)
	}
}

func TestRunResolutionWorkflow_UnrecognizedApproach_Refused(t *testing.T) {
	rep, err := RunResolutionWorkflow(t.TempDir(), "delete-everything", "", ResolutionHandoff{}, "", nil)
	if err == nil {
		t.Fatalf("expected an unrecognized resolution approach to be refused")
	}
	if rep != nil {
		t.Fatalf("expected no report for a refused, never-attempted dispatch, got %+v", rep)
	}
}

// TestRunResolutionWorkflow_HumanDecisionValue_NotDispatchable proves the removed "human-decision"
// value is refused as a dispatch target exactly like any other unrecognized string — it no longer
// exists as a named constant, but nothing should ever again trust it as an approach.
func TestRunResolutionWorkflow_HumanDecisionValue_NotDispatchable(t *testing.T) {
	rep, err := RunResolutionWorkflow(t.TempDir(), "human-decision", "", ResolutionHandoff{}, "", nil)
	if err == nil {
		t.Fatalf("expected human-decision to be refused as a dispatch target")
	}
	if rep != nil {
		t.Fatalf("expected no report, got %+v", rep)
	}
}

// setupVerificationTarget returns a plain initialized project root — Verification/Review declare
// specification_reference: none, so no Specification setup is needed for eligibility.
func setupVerificationTarget(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	configureGitIdentity(t, root)
	if _, err := Init(root); err != nil {
		t.Fatal(err)
	}
	return root
}
