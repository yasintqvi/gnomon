package orchestrate

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"gnomon/internal/present"
	"gnomon/internal/project"
)

// commitEverything stages and commits all current content, so a freshly-initialized project (whose
// materialized .gnomon/ tree is otherwise untracked) can genuinely be exercised as a clean working
// tree — matching what "clean" means to Git, not merely "nothing edited since Init".
func commitEverything(t *testing.T, root string) {
	t.Helper()
	run := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = root
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v failed: %v\n%s", args, err, out)
		}
	}
	run("add", "-A")
	run("commit", "-q", "-m", "snapshot")
}

func TestStatus_CleanInitializedProject_ZeroSpecifications(t *testing.T) {
	root := t.TempDir()
	configureGitIdentity(t, root)
	if _, err := Init(root); err != nil {
		t.Fatal(err)
	}
	commitEverything(t, root)

	report, err := Status(root)
	if err != nil {
		t.Fatalf("expected a clean initialized project to report cleanly: %v", err)
	}
	if report.Outcome != present.Success {
		t.Fatalf("expected Success, got %v", report.Outcome)
	}
	rendered := present.Render(report, false)
	if !strings.Contains(rendered, "No Specifications exist yet") {
		t.Fatalf("expected zero Specifications to be reported plainly: %s", rendered)
	}
	if !strings.Contains(rendered, "clean") {
		t.Fatalf("expected a clean working tree to be reported: %s", rendered)
	}
}

func TestStatus_DraftSpecification(t *testing.T) {
	root := t.TempDir()
	configureGitIdentity(t, root)
	if _, err := Init(root); err != nil {
		t.Fatal(err)
	}
	if _, err := SpecCreate(root, "Password Reset"); err != nil {
		t.Fatal(err)
	}

	report, err := Status(root)
	if err != nil {
		t.Fatal(err)
	}
	rendered := present.Render(report, false)
	if !strings.Contains(rendered, "SPEC-001: Draft") {
		t.Fatalf("expected SPEC-001's Draft lifecycle to be reported: %s", rendered)
	}
}

func TestStatus_ApprovedSpecification(t *testing.T) {
	root := setupApprovedSpec(t)

	report, err := Status(root)
	if err != nil {
		t.Fatal(err)
	}
	rendered := present.Render(report, false)
	if !strings.Contains(rendered, "SPEC-001: Approved") {
		t.Fatalf("expected SPEC-001's Approved lifecycle to be reported: %s", rendered)
	}
}

func TestStatus_RevokedApproval_ReflectedAsDraft(t *testing.T) {
	root := setupApprovedSpec(t)
	if _, err := Revoke(root, "SPEC-001", nil); err != nil {
		t.Fatal(err)
	}

	report, err := Status(root)
	if err != nil {
		t.Fatal(err)
	}
	rendered := present.Render(report, false)
	if !strings.Contains(rendered, "SPEC-001: Draft") {
		t.Fatalf("expected a revoked Specification to derive back to Draft, exactly as facts.Lifecycle already guarantees: %s", rendered)
	}
}

func TestStatus_StaleApproval_ContentChangedAfterGrant_ReflectedAsDraft(t *testing.T) {
	root := setupApprovedSpec(t)

	l, err := project.Locate(root)
	if err != nil {
		t.Fatal(err)
	}
	specPath := filepath.Join(l.SpecificationsDir(), "SPEC-001-password-reset.md")
	data, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(specPath, append(data, []byte("\nExtra content invalidating the fingerprint.\n")...), 0o644); err != nil {
		t.Fatal(err)
	}

	report, err := Status(root)
	if err != nil {
		t.Fatal(err)
	}
	rendered := present.Render(report, false)
	if !strings.Contains(rendered, "SPEC-001: Draft") {
		t.Fatalf("expected stale approval evidence (fingerprint mismatch) to derive Draft: %s", rendered)
	}
}

func TestStatus_MultipleSpecifications_AllListedNoCurrentSpecChosen(t *testing.T) {
	root := t.TempDir()
	configureGitIdentity(t, root)
	if _, err := Init(root); err != nil {
		t.Fatal(err)
	}
	if _, err := SpecCreate(root, "Password Reset"); err != nil {
		t.Fatal(err)
	}
	if _, err := SpecCreate(root, "Email Verification"); err != nil {
		t.Fatal(err)
	}
	if _, err := Approve(root, "SPEC-002", nil); err != nil {
		t.Fatal(err)
	}

	report, err := Status(root)
	if err != nil {
		t.Fatal(err)
	}
	rendered := present.Render(report, false)
	if !strings.Contains(rendered, "SPEC-001: Draft") {
		t.Fatalf("expected SPEC-001 (Draft) to be listed: %s", rendered)
	}
	if !strings.Contains(rendered, "SPEC-002: Approved") {
		t.Fatalf("expected SPEC-002 (Approved) to be listed: %s", rendered)
	}
	if report.Target != "" {
		t.Fatalf("expected no single Specification to be singled out as report.Target — status never invents a current Specification, got %q", report.Target)
	}
}

func TestStatus_UncommittedChanges_Reported(t *testing.T) {
	root := t.TempDir()
	configureGitIdentity(t, root)
	if _, err := Init(root); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "untracked.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	report, err := Status(root)
	if err != nil {
		t.Fatal(err)
	}
	rendered := present.Render(report, false)
	if !strings.Contains(rendered, "uncommitted changes present") {
		t.Fatalf("expected an untracked file to be reported as an uncommitted change: %s", rendered)
	}
}

// TestStatus_NotAGitRepository_SpecStateStillReported_GitFinalizationNotApplicable is status's
// side of the same dogfooding regression covered in next_test.go: an initialized project that is
// not a Git repository must still report Specification state.
func TestStatus_NotAGitRepository_SpecStateStillReported_GitFinalizationNotApplicable(t *testing.T) {
	root := t.TempDir()
	if _, err := Init(root); err != nil {
		t.Fatal(err)
	}
	if _, err := SpecCreate(root, "Password Reset"); err != nil {
		t.Fatal(err)
	}

	report, err := Status(root)
	if err != nil {
		t.Fatalf("expected a non-Git project to still produce Status, got: %v", err)
	}
	if report.Outcome != present.Success {
		t.Fatalf("expected Success, got %v", report.Outcome)
	}
	rendered := present.Render(report, false)
	if !strings.Contains(rendered, "SPEC-001: Draft") {
		t.Fatalf("expected Specification state despite no Git repository: %s", rendered)
	}
	if !strings.Contains(rendered, "not a Git repository") || !strings.Contains(rendered, "Git Finalization is not applicable") {
		t.Fatalf("expected an explicit, concise statement that Git Finalization is not applicable: %s", rendered)
	}
}

// TestStatus_UnexpectedGitFailure_PreservesSpecState_NotMisclassifiedAsNoRepo mirrors
// TestNext_UnexpectedGitFailure_PreservesSpecGuidance_NotMisclassifiedAsNoRepo for status.
func TestStatus_UnexpectedGitFailure_PreservesSpecState_NotMisclassifiedAsNoRepo(t *testing.T) {
	root := t.TempDir()
	configureGitIdentity(t, root)
	if _, err := Init(root); err != nil {
		t.Fatal(err)
	}
	if _, err := SpecCreate(root, "Password Reset"); err != nil {
		t.Fatal(err)
	}
	// Corrupt the index so `git rev-parse --git-dir` still succeeds (it is still a Git
	// repository) but `git status` fails for a reason that is NOT "not a repository".
	if err := os.WriteFile(filepath.Join(root, ".git", "index"), []byte("garbage not an index"), 0o644); err != nil {
		t.Fatal(err)
	}

	report, err := Status(root)
	if err == nil {
		t.Fatalf("expected an unexpected Git failure to be reported as an error")
	}
	if report == nil {
		t.Fatalf("expected a Report to still be returned despite the Git failure")
	}
	rendered := present.Render(report, false)
	if !strings.Contains(rendered, "SPEC-001: Draft") {
		t.Fatalf("expected Specification state to survive an unexpected Git failure: %s", rendered)
	}
	if strings.Contains(rendered, "not a Git repository") {
		t.Fatalf("did not expect an unexpected Git failure to be misclassified as 'not a Git repository': %s", rendered)
	}
}

func TestStatus_UnsupportedContractVersion_RefusesAndPointsAtValidate(t *testing.T) {
	root := t.TempDir()
	configureGitIdentity(t, root)
	if _, err := Init(root); err != nil {
		t.Fatal(err)
	}
	l, err := project.Locate(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(l.ContractVersionPath(), []byte("999\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	report, err := Status(root)
	if err == nil {
		t.Fatalf("expected an unsupported contract version to be refused")
	}
	if report == nil || report.Outcome != present.Failed {
		t.Fatalf("expected a Failed report, got: %+v", report)
	}
	if !strings.Contains(report.Next, "gnomon validate") {
		t.Fatalf("expected status to point at validate for the structural diagnosis, got Next=%q", report.Next)
	}
}

func TestStatus_MalformedApprovalEvidence_StillDerivesCleanly(t *testing.T) {
	// Malformed evidence is excluded from lifecycle derivation (facts.Lifecycle's own guarantee,
	// unchanged by this slice) — status must not crash or misreport merely because it exists;
	// diagnosing it is validate's job, not status's.
	root := setupApprovedSpec(t)
	l, err := project.Locate(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(l.ApprovalsDir(), "SPEC-001", "bad.grant.json"), []byte("not json"), 0o644); err != nil {
		t.Fatal(err)
	}

	report, err := Status(root)
	if err != nil {
		t.Fatalf("expected malformed evidence to be excluded, not fatal, to status: %v", err)
	}
	rendered := present.Render(report, false)
	if !strings.Contains(rendered, "SPEC-001: Approved") {
		t.Fatalf("expected the real, valid grant to still be honored: %s", rendered)
	}
}

func TestStatus_NeverInvokesAnAgent(t *testing.T) {
	sandboxAgentConfig(t)
	root := setupApprovedSpec(t)
	if _, err := Status(root); err != nil {
		t.Fatalf("expected Status to succeed without ever needing an Agent provider: %v", err)
	}
}

func TestStatus_NeverMutatesProjectFiles(t *testing.T) {
	root := setupApprovedSpec(t)
	l, err := project.Locate(root)
	if err != nil {
		t.Fatal(err)
	}

	before := map[string]string{}
	_ = filepath.Walk(l.GnomonRoot(), func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		data, _ := os.ReadFile(path)
		before[path] = string(data)
		return nil
	})

	if _, err := Status(root); err != nil {
		t.Fatal(err)
	}

	for path, want := range before {
		got, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("expected %s to still exist after Status: %v", path, err)
		}
		if string(got) != want {
			t.Fatalf("expected Status to never modify %s, but its content changed", path)
		}
	}
}

func TestStatus_DiffersFromValidate_DoesNotAuditContracts(t *testing.T) {
	// A duplicate workflow identity is a validate-scope finding; status must not surface it —
	// status answers a question about work, not tooling/installation integrity.
	root := t.TempDir()
	configureGitIdentity(t, root)
	if _, err := Init(root); err != nil {
		t.Fatal(err)
	}
	l, err := project.Locate(root)
	if err != nil {
		t.Fatal(err)
	}
	duplicate := "---\nidentity: bootstrap\nspecification_reference: none\nresult:\n" +
		"  terminal_path: outcome\n  classification:\n    OK: success\n  schema:\n    type: object\n" +
		"    required: [outcome]\n    properties:\n      outcome:\n        type: string\n        enum: [OK]\n" +
		"---\n# Duplicate\n"
	if err := os.WriteFile(filepath.Join(l.WorkflowsDir(), "bootstrap-copy.md"), []byte(duplicate), 0o644); err != nil {
		t.Fatal(err)
	}

	statusReport, statusErr := Status(root)
	if statusErr != nil {
		t.Fatalf("expected status to succeed despite a validate-scope problem it does not audit: %v", statusErr)
	}
	if statusReport.Outcome != present.Success {
		t.Fatalf("expected status to still succeed: %+v", statusReport)
	}

	validateReport, validateErr := Validate(root)
	if validateErr == nil {
		t.Fatalf("expected validate to actually catch the duplicate identity status correctly ignores")
	}
	if validateReport.Outcome != present.Failed {
		t.Fatalf("expected validate to fail on the duplicate identity, got %+v", validateReport)
	}
}
