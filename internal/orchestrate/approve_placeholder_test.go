package orchestrate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gnomon/internal/present"
	"gnomon/internal/project"
)

// TestApprove_Placeholders_InteractiveNo_WritesNothingAndCancels covers both a "n" answer and an
// empty answer (the same default-No behavior a real confirm callback gives for either).
func TestApprove_Placeholders_InteractiveNo_WritesNothingAndCancels(t *testing.T) {
	root := t.TempDir()
	configureGitIdentity(t, root)
	if _, err := Init(root); err != nil {
		t.Fatal(err)
	}
	if _, err := specCreateDefined(root, "Password Reset"); err != nil {
		t.Fatal(err)
	}

	confirmCalled := false
	confirm := func(warning string) (bool, error) {
		confirmCalled = true
		return false, nil
	}
	promptCalled := false
	prompt := func(message string) (string, error) {
		promptCalled = true
		return "", nil
	}

	report, err := Approve(root, "SPEC-001", prompt, confirm)
	if err != nil {
		t.Fatalf("approve: %v", err)
	}
	if !confirmCalled {
		t.Fatalf("expected the placeholder confirmation to be asked")
	}
	if promptCalled {
		t.Fatalf("expected identity to never be resolved when approval is cancelled")
	}
	if report.Outcome != present.Cancelled {
		t.Fatalf("expected Cancelled, got %v", report.Outcome)
	}

	l, err := project.Locate(root)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(l.ApprovalsDir()); err == nil {
		entries, _ := os.ReadDir(filepath.Join(l.ApprovalsDir(), "SPEC-001"))
		if len(entries) > 0 {
			t.Fatalf("expected no approval evidence written, found %v", entries)
		}
	}
}

func TestApprove_Placeholders_InteractiveYes_ApprovesNormally(t *testing.T) {
	root := t.TempDir()
	configureGitIdentity(t, root)
	if _, err := Init(root); err != nil {
		t.Fatal(err)
	}
	if _, err := specCreateDefined(root, "Password Reset"); err != nil {
		t.Fatal(err)
	}

	confirm := func(warning string) (bool, error) { return true, nil }

	report, err := Approve(root, "SPEC-001", nil, confirm)
	if err != nil {
		t.Fatalf("approve: %v", err)
	}
	if report.Outcome != present.Success {
		t.Fatalf("expected Success, got %v", report.Outcome)
	}
	if !strings.Contains(report.Summary, "Approved") {
		t.Fatalf("expected the Specification to be Approved, got %q", report.Summary)
	}
}

func TestApprove_Placeholders_NonInteractive_ApprovesAndReportsWarning(t *testing.T) {
	root := t.TempDir()
	configureGitIdentity(t, root)
	if _, err := Init(root); err != nil {
		t.Fatal(err)
	}
	if _, err := specCreateDefined(root, "Password Reset"); err != nil {
		t.Fatal(err)
	}

	var warningSeen string
	// Mirrors the real non-interactive confirm: always "print" (here, just capture) the warning,
	// then proceed without asking.
	confirm := func(warning string) (bool, error) {
		warningSeen = warning
		return true, nil
	}

	report, err := Approve(root, "SPEC-001", nil, confirm)
	if err != nil {
		t.Fatalf("approve: %v", err)
	}
	if report.Outcome != present.Success {
		t.Fatalf("expected Success, got %v", report.Outcome)
	}
	if warningSeen == "" {
		t.Fatalf("expected the placeholder warning to be produced")
	}
	if !strings.Contains(warningSeen, "template placeholder") {
		t.Fatalf("expected the warning to mention template placeholders, got %q", warningSeen)
	}
}

func TestApprove_NoPlaceholders_NoWarningNoPrompt(t *testing.T) {
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
	specPath := filepath.Join(l.SpecificationsDir(), "SPEC-001-password-reset.md")
	if err := os.WriteFile(specPath, []byte("# SPEC-001 — Password Reset\n\nFully defined. No brackets here at all.\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	confirmCalled := false
	confirm := func(warning string) (bool, error) {
		confirmCalled = true
		return true, nil
	}

	report, err := Approve(root, "SPEC-001", nil, confirm)
	if err != nil {
		t.Fatalf("approve: %v", err)
	}
	if confirmCalled {
		t.Fatalf("expected the placeholder confirmation to never be asked when there are no placeholders")
	}
	if report.Outcome != present.Success {
		t.Fatalf("expected Success, got %v", report.Outcome)
	}
}

func TestApprove_UnfilledTemplate_RefusedEvenNonInteractive(t *testing.T) {
	for _, detailed := range []bool{false, true} {
		root := t.TempDir()
		configureGitIdentity(t, root)
		if _, err := Init(root); err != nil {
			t.Fatal(err)
		}
		if _, err := SpecCreateFrom(root, "Password Reset", detailed); err != nil {
			t.Fatal(err)
		}

		// nil confirm is the non-interactive path, which used to approve after a warning.
		report, err := Approve(root, "SPEC-001", nil, nil)
		if err == nil {
			t.Fatalf("detailed=%v: expected an unfilled template to be refused", detailed)
		}
		if report == nil || report.Outcome != present.Blocked {
			t.Fatalf("detailed=%v: expected Blocked, got %+v", detailed, report)
		}
		l, _ := project.Locate(root)
		if entries, _ := os.ReadDir(filepath.Join(l.ApprovalsDir(), "SPEC-001")); len(entries) > 0 {
			t.Fatalf("detailed=%v: expected no approval evidence, found %v", detailed, entries)
		}
	}
}

// TestApprove_CustomizedLegacyTemplate_IsUsedAndRecognized covers a project whose own
// .gnomon/specifications/SPEC-000-use-case-name.md was edited: new Specifications come from it,
// an untouched one is refused, and its own placeholders drive the warning.
func TestApprove_CustomizedLegacyTemplate_IsUsedAndRecognized(t *testing.T) {
	root := t.TempDir()
	configureGitIdentity(t, root)
	if _, err := Init(root); err != nil {
		t.Fatal(err)
	}
	l, err := project.Locate(root)
	if err != nil {
		t.Fatal(err)
	}
	custom := "# SPEC-[ID] — [Use Case Name]\n\n## Story\n\n[Team-specific story field]\n\n## Checks\n\n[Team-specific check]\n"
	if err := os.WriteFile(filepath.Join(l.SpecificationsDir(), "SPEC-000-use-case-name.md"), []byte(custom), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := SpecCreate(root, "Password Reset"); err != nil {
		t.Fatal(err)
	}
	specPath := filepath.Join(l.SpecificationsDir(), "SPEC-001-password-reset.md")
	created, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(created), "[Team-specific story field]") {
		t.Fatalf("expected the project's customized template to be used, got:\n%s", created)
	}
	if _, err := Approve(root, "SPEC-001", nil, nil); err == nil {
		t.Fatalf("expected the untouched draft to be refused")
	}

	edited := strings.Replace(string(created), "[Team-specific story field]", "A user resets a forgotten password by email.", 1)
	if err := os.WriteFile(specPath, []byte(edited), 0o644); err != nil {
		t.Fatal(err)
	}
	var warning string
	report, err := Approve(root, "SPEC-001", nil, func(w string) (bool, error) { warning = w; return true, nil })
	if err != nil || report.Outcome != present.Success {
		t.Fatalf("expected approval to proceed after confirmation, got %+v (err=%v)", report, err)
	}
	if !strings.Contains(warning, "[Team-specific check]") || strings.Contains(warning, "[Team-specific story field]") {
		t.Fatalf("expected the warning to list exactly the remaining custom placeholder, got:\n%s", warning)
	}
}

func TestApprove_AlreadyApproved_NoPlaceholderPrompt(t *testing.T) {
	root := t.TempDir()
	configureGitIdentity(t, root)
	if _, err := Init(root); err != nil {
		t.Fatal(err)
	}
	if _, err := specCreateDefined(root, "Password Reset"); err != nil {
		t.Fatal(err)
	}
	if _, err := Approve(root, "SPEC-001", nil, func(string) (bool, error) { return true, nil }); err != nil {
		t.Fatalf("first approve: %v", err)
	}

	confirmCalled := false
	confirm := func(warning string) (bool, error) {
		confirmCalled = true
		return true, nil
	}

	report, err := Approve(root, "SPEC-001", nil, confirm)
	if err != nil {
		t.Fatalf("second approve: %v", err)
	}
	if confirmCalled {
		t.Fatalf("expected no placeholder prompt for an already-Approved re-approval")
	}
	if !strings.Contains(report.Summary, "already Approved") {
		t.Fatalf("expected the already-Approved no-op message, got %q", report.Summary)
	}
}
