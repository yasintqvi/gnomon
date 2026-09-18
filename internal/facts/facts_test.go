package facts

import (
	"os"
	"path/filepath"
	"testing"

	"gnomon/internal/approval"
	"gnomon/internal/contract"
	"gnomon/internal/project"
)

func setupLayout(t *testing.T) project.Layout {
	t.Helper()
	root := t.TempDir()
	if err := project.Materialize(root); err != nil {
		t.Fatal(err)
	}
	return project.Layout{Root: root}
}

func writeSpec(t *testing.T, l project.Layout, filename, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(l.SpecificationsDir(), filename), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestEligible_NoneReference_AlwaysEligible(t *testing.T) {
	l := setupLayout(t)
	wf := contract.Workflow{SpecificationReference: contract.SpecReferenceNone}
	e, err := Eligible(l, wf, "")
	if err != nil {
		t.Fatal(err)
	}
	if !e.Eligible {
		t.Fatalf("expected a specification_reference=none workflow to always be eligible")
	}
}

func TestEligible_RequiredReference_MissingArgument(t *testing.T) {
	l := setupLayout(t)
	wf := contract.Workflow{SpecificationReference: contract.SpecReferenceRequired}
	e, err := Eligible(l, wf, "")
	if err != nil {
		t.Fatal(err)
	}
	if e.Eligible {
		t.Fatalf("expected ineligibility when a required Specification reference is missing")
	}
}

func TestEligible_RequiredReference_NonexistentSpec(t *testing.T) {
	l := setupLayout(t)
	wf := contract.Workflow{SpecificationReference: contract.SpecReferenceRequired}
	e, err := Eligible(l, wf, "SPEC-999")
	if err != nil {
		t.Fatal(err)
	}
	if e.Eligible {
		t.Fatalf("expected ineligibility for a nonexistent Specification")
	}
}

func TestEligible_RequiresApproval_DraftSpec_Ineligible(t *testing.T) {
	l := setupLayout(t)
	writeSpec(t, l, "SPEC-004-x.md", "# SPEC-004 — X\n\nsome content\n")
	wf := contract.Workflow{SpecificationReference: contract.SpecReferenceRequired, RequiresApprovedSpecification: true}
	e, err := Eligible(l, wf, "SPEC-004")
	if err != nil {
		t.Fatal(err)
	}
	if e.Eligible {
		t.Fatalf("expected a Draft Specification to be ineligible for a workflow requiring Approved")
	}
}

func TestEligible_RequiresApproval_ApprovedSpec_Eligible(t *testing.T) {
	l := setupLayout(t)
	content := "# SPEC-004 — X\n\nsome content\n"
	writeSpec(t, l, "SPEC-004-x.md", content)
	fp := approval.Fingerprint([]byte(content))
	if err := approval.WriteGrant(l.ApprovalsDir(), "SPEC-004", fp, "Jane <jane@example.com>", []byte("content")); err != nil {
		t.Fatal(err)
	}
	wf := contract.Workflow{SpecificationReference: contract.SpecReferenceRequired, RequiresApprovedSpecification: true}
	e, err := Eligible(l, wf, "SPEC-004")
	if err != nil {
		t.Fatal(err)
	}
	if !e.Eligible {
		t.Fatalf("expected an Approved Specification to be eligible, got reason: %s", e.Reason)
	}
}

func TestEligible_ManualEditAfterApproval_RevertsToDraft(t *testing.T) {
	l := setupLayout(t)
	original := "# SPEC-004 — X\n\nsome content\n"
	writeSpec(t, l, "SPEC-004-x.md", original)
	fp := approval.Fingerprint([]byte(original))
	if err := approval.WriteGrant(l.ApprovalsDir(), "SPEC-004", fp, "Jane <jane@example.com>", []byte("content")); err != nil {
		t.Fatal(err)
	}

	// A Human edits the approved content manually, outside any Gnomon workflow.
	writeSpec(t, l, "SPEC-004-x.md", "# SPEC-004 — X\n\nsome DIFFERENT content\n")

	wf := contract.Workflow{SpecificationReference: contract.SpecReferenceRequired, RequiresApprovedSpecification: true}
	e, err := Eligible(l, wf, "SPEC-004")
	if err != nil {
		t.Fatal(err)
	}
	if e.Eligible {
		t.Fatalf("expected the manual edit to silently revert lifecycle to Draft with no explicit action")
	}
}
