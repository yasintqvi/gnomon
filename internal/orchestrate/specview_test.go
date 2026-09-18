package orchestrate

import (
	"os"
	"testing"

	"gnomon/internal/approval"
	"gnomon/internal/project"
)

func actionFor(actions []SpecAction, verb string) *SpecAction {
	for i := range actions {
		if actions[i].Verb == verb {
			return &actions[i]
		}
	}
	return nil
}

func TestSpecActions_Draft_DefineAndApproveAvailable_ImplementTestRevokeNot(t *testing.T) {
	root := t.TempDir()
	configureGitIdentity(t, root)
	if _, err := Init(root); err != nil {
		t.Fatal(err)
	}
	if _, err := SpecCreate(root, "Password Reset"); err != nil {
		t.Fatal(err)
	}
	l, err := project.Locate(root)
	if err != nil {
		t.Fatal(err)
	}

	actions, unreadable, err := SpecActions(l, "SPEC-001")
	if err != nil {
		t.Fatal(err)
	}
	if len(unreadable) != 0 {
		t.Fatalf("expected all built-in Contracts to load, got unreadable: %v", unreadable)
	}
	if a := actionFor(actions, "define"); a == nil || !a.Available {
		t.Fatalf("expected define available on a Draft Specification, got %+v", a)
	}
	if a := actionFor(actions, "approve"); a == nil || !a.Available {
		t.Fatalf("expected approve available on a Draft Specification, got %+v", a)
	}
	for _, verb := range []string{"implement", "test", "revoke"} {
		a := actionFor(actions, verb)
		if a == nil || a.Available {
			t.Fatalf("expected %s unavailable on a Draft Specification, got %+v", verb, a)
		}
		if a.Reason == "" {
			t.Fatalf("expected a reason for unavailable %s, got none", verb)
		}
	}
	if actionFor(actions, "verify") != nil || actionFor(actions, "review") != nil {
		t.Fatalf("expected verify/review to be absent — specification_reference: none means neither is Specification-scoped, got: %+v", actions)
	}
}

func TestSpecActions_Approved_ImplementTestRevokeAvailable_ApproveNot(t *testing.T) {
	root := setupApprovedSpec(t)
	l, err := project.Locate(root)
	if err != nil {
		t.Fatal(err)
	}

	actions, _, err := SpecActions(l, "SPEC-001")
	if err != nil {
		t.Fatal(err)
	}
	for _, verb := range []string{"implement", "test", "revoke"} {
		if a := actionFor(actions, verb); a == nil || !a.Available {
			t.Fatalf("expected %s available on an Approved Specification, got %+v", verb, actionFor(actions, verb))
		}
	}
	if a := actionFor(actions, "approve"); a == nil || a.Available {
		t.Fatalf("expected approve unavailable on an already-Approved Specification, got %+v", a)
	}
	if a := actionFor(actions, "define"); a == nil || !a.Available {
		t.Fatalf("expected define to remain available regardless of lifecycle, got %+v", a)
	}
}

func TestListSpecs_EmptyProject(t *testing.T) {
	root := t.TempDir()
	configureGitIdentity(t, root)
	if _, err := Init(root); err != nil {
		t.Fatal(err)
	}
	l, err := project.Locate(root)
	if err != nil {
		t.Fatal(err)
	}
	summaries, err := ListSpecs(l)
	if err != nil {
		t.Fatal(err)
	}
	if len(summaries) != 0 {
		t.Fatalf("expected no Specifications, got %v", summaries)
	}
}

func TestListSpecs_ReturnsTitleAndLifecycle(t *testing.T) {
	root := setupApprovedSpec(t)
	l, err := project.Locate(root)
	if err != nil {
		t.Fatal(err)
	}
	summaries, err := ListSpecs(l)
	if err != nil {
		t.Fatal(err)
	}
	if len(summaries) != 1 {
		t.Fatalf("expected 1 Specification, got %d", len(summaries))
	}
	if summaries[0].ID != "SPEC-001" || summaries[0].Lifecycle != approval.Approved {
		t.Fatalf("unexpected summary: %+v", summaries[0])
	}
	if summaries[0].Title == "" {
		t.Fatalf("expected a non-empty title")
	}
}

func TestSpecDetailFor_NonexistentSpec_Errors(t *testing.T) {
	root := t.TempDir()
	configureGitIdentity(t, root)
	if _, err := Init(root); err != nil {
		t.Fatal(err)
	}
	l, err := project.Locate(root)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := SpecDetailFor(l, "SPEC-999"); err == nil {
		t.Fatalf("expected an error for a nonexistent Specification")
	}
}

func TestSpecDetailFor_Approved_ReportsActiveGrant(t *testing.T) {
	root := setupApprovedSpec(t)
	l, err := project.Locate(root)
	if err != nil {
		t.Fatal(err)
	}
	detail, err := SpecDetailFor(l, "SPEC-001")
	if err != nil {
		t.Fatal(err)
	}
	if detail.Lifecycle != approval.Approved {
		t.Fatalf("expected Approved, got %s", detail.Lifecycle)
	}
	if detail.ActiveGrant == nil {
		t.Fatalf("expected an active grant to be reported")
	}
	if detail.ActiveGrant.Fingerprint != detail.Fingerprint {
		t.Fatalf("expected the active grant's fingerprint to match current content")
	}
	if len(detail.Revisions) != 1 {
		t.Fatalf("expected 1 revision, got %d", len(detail.Revisions))
	}
}

func TestSpecDetailFor_StaleApproval_NoActiveGrant_PreviousRevisionPreserved(t *testing.T) {
	root := setupApprovedSpec(t)
	l, err := project.Locate(root)
	if err != nil {
		t.Fatal(err)
	}
	specPath := l.SpecificationsDir() + "/SPEC-001-password-reset.md"
	data, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := writeFile(t, specPath, string(data)+"\nExtra line.\n"); err != nil {
		t.Fatal(err)
	}

	detail, err := SpecDetailFor(l, "SPEC-001")
	if err != nil {
		t.Fatal(err)
	}
	if detail.Lifecycle != approval.Draft {
		t.Fatalf("expected Draft after content changed, got %s", detail.Lifecycle)
	}
	if detail.ActiveGrant != nil {
		t.Fatalf("expected no active grant once content changed, got %+v", detail.ActiveGrant)
	}
	if len(detail.Revisions) != 1 || detail.Revisions[0].Content == "" {
		t.Fatalf("expected the previous approved revision's content preserved, got %+v", detail.Revisions)
	}
}
