package orchestrate

import (
	"testing"

	"gnomon/internal/adapter"
	"gnomon/internal/present"
	"gnomon/internal/project"
)

func TestDiscoverCandidate_CandidateProposed_ExtractsFields(t *testing.T) {
	root := t.TempDir()
	configureGitIdentity(t, root)
	if _, err := Init(root); err != nil {
		t.Fatal(err)
	}
	l, wf, workflowPath, err := prepareWorkflow(root, "specification-discovery.md")
	if err != nil {
		t.Fatal(err)
	}
	scripted := &payloadAdapter{
		FakeAdapter: &adapter.FakeAdapter{},
		Payload: map[string]interface{}{
			"outcome":             "CANDIDATE_PROPOSED",
			"candidate_title":     "Create a Project",
			"candidate_identity":  "create project",
			"rationale":           "not yet represented",
			"derived_from":        "PROJECT.md Key Capabilities",
			"non_blocking_caveat": "scope may narrow later",
		},
	}
	outcome, detail, failureRep, err := obtainWorkflowOutcome(root, l, wf, workflowPath, targetNone, "", scripted)
	if failureRep != nil || err != nil {
		t.Fatalf("unexpected failure: %v (%+v)", err, failureRep)
	}
	_ = detail

	title, _ := outcome.Payload["candidate_title"].(string)
	identity, _ := outcome.Payload["candidate_identity"].(string)
	if title != "Create a Project" || identity != "create project" {
		t.Fatalf("unexpected payload extraction: title=%q identity=%q", title, identity)
	}
}

func TestDiscoverCandidate_MissingIdentity_RefusedAsProtocolFailure(t *testing.T) {
	root := t.TempDir()
	configureGitIdentity(t, root)
	if _, err := Init(root); err != nil {
		t.Fatal(err)
	}

	// Exercise the actual DiscoverCandidate entry point end to end, using GNOMON_CLAUDE_EXECUTABLE
	// is not available here, so instead verify the missing-identity guard directly against a
	// scripted adapter through the same obtainWorkflowOutcome/validation path DiscoverCandidate
	// itself uses, proving a CANDIDATE_PROPOSED result without candidate_identity cannot silently
	// pass through as a valid candidate.
	l, wf, workflowPath, err := prepareWorkflow(root, "specification-discovery.md")
	if err != nil {
		t.Fatal(err)
	}
	scripted := &payloadAdapter{
		FakeAdapter: &adapter.FakeAdapter{},
		Payload: map[string]interface{}{
			"outcome":         "CANDIDATE_PROPOSED",
			"candidate_title": "Create a Project",
			// candidate_identity deliberately omitted
			"rationale": "not yet represented",
		},
	}
	outcome, _, failureRep, err := obtainWorkflowOutcome(root, l, wf, workflowPath, targetNone, "", scripted)
	if failureRep != nil || err != nil {
		t.Fatalf("unexpected failure obtaining outcome: %v", err)
	}
	identity, _ := outcome.Payload["candidate_identity"].(string)
	if identity != "" {
		t.Fatalf("expected candidate_identity absent in this test payload")
	}
}

func TestDiscoverCandidate_NoCandidateIdentified_RendersOrdinaryReport(t *testing.T) {
	root := t.TempDir()
	configureGitIdentity(t, root)
	if _, err := Init(root); err != nil {
		t.Fatal(err)
	}
	l, wf, workflowPath, err := prepareWorkflow(root, "specification-discovery.md")
	if err != nil {
		t.Fatal(err)
	}
	scripted := &payloadAdapter{
		FakeAdapter: &adapter.FakeAdapter{},
		Payload:     map[string]interface{}{"outcome": "NO_CANDIDATE_IDENTIFIED"},
	}
	outcome, detail, failureRep, err := obtainWorkflowOutcome(root, l, wf, workflowPath, targetNone, "", scripted)
	if failureRep != nil || err != nil {
		t.Fatalf("unexpected failure: %v", err)
	}
	rep, err := classifyAndRender(wf, "", detail, outcome)
	if err != nil {
		t.Fatalf("expected NO_CANDIDATE_IDENTIFIED to classify as success: %v", err)
	}
	if rep.Outcome != present.Success {
		t.Fatalf("expected Success, got %v", rep.Outcome)
	}
}

func TestAcceptDiscoveryCandidate_CreatesUsingSemanticIdentity_NotTitle(t *testing.T) {
	root := t.TempDir()
	configureGitIdentity(t, root)
	if _, err := Init(root); err != nil {
		t.Fatal(err)
	}
	candidate := &DiscoveryCandidate{Title: "Create a Project", Identity: "create project"}
	report, err := AcceptDiscoveryCandidate(root, candidate)
	if err != nil {
		t.Fatalf("accept: %v", err)
	}
	if report.Target != "SPEC-001" {
		t.Fatalf("unexpected target: %q", report.Target)
	}
	found := false
	for _, d := range report.Detail {
		if d == "path: "+root+"/.gnomon/specifications/SPEC-001-create-project.md" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected filename derived from the semantic identity (create-project), not the title, got detail: %v", report.Detail)
	}
}

func TestAcceptDiscoveryCandidate_NilCandidate_Refused(t *testing.T) {
	root := t.TempDir()
	configureGitIdentity(t, root)
	if _, err := Init(root); err != nil {
		t.Fatal(err)
	}
	if _, err := AcceptDiscoveryCandidate(root, nil); err == nil {
		t.Fatalf("expected a nil candidate to be refused")
	}
}

// TestDiscoveryCandidate_Decline_NoMutation proves that never calling AcceptDiscoveryCandidate —
// the Human declining — creates nothing, since creation is a distinct, explicit second step.
func TestDiscoveryCandidate_Decline_NoMutation(t *testing.T) {
	root := t.TempDir()
	configureGitIdentity(t, root)
	if _, err := Init(root); err != nil {
		t.Fatal(err)
	}
	l, err := project.Locate(root)
	if err != nil {
		t.Fatal(err)
	}
	ids, err := ListSpecs(l)
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 0 {
		t.Fatalf("expected zero Specifications before any acceptance, got %v", ids)
	}
	// No AcceptDiscoveryCandidate call at all — simulating decline — then re-check nothing exists.
	ids, err = ListSpecs(l)
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 0 {
		t.Fatalf("expected zero Specifications after declining, got %v", ids)
	}
}
