package orchestrate

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"gnomon/internal/approval"
	"gnomon/internal/facts"
	"gnomon/internal/present"
	"gnomon/internal/project"
)

// --- Core revocation behavior ---

func TestRevoke_ApprovedBecomesNonApproved(t *testing.T) {
	root := t.TempDir()
	configureGitIdentity(t, root)
	if _, err := Init(root); err != nil {
		t.Fatal(err)
	}
	if _, err := SpecCreate(root, "Password Reset"); err != nil {
		t.Fatal(err)
	}
	if _, err := Approve(root, "SPEC-001", nil); err != nil {
		t.Fatal(err)
	}

	l, err := project.Locate(root)
	if err != nil {
		t.Fatal(err)
	}
	before, err := facts.Lifecycle(l, "SPEC-001")
	if err != nil {
		t.Fatal(err)
	}
	if before != approval.Approved {
		t.Fatalf("expected SPEC-001 to be Approved before revocation, got %v", before)
	}

	report, err := Revoke(root, "SPEC-001", nil)
	if err != nil {
		t.Fatalf("revoke: %v (report: %+v)", err, report)
	}
	if report.Outcome != present.Success {
		t.Fatalf("expected Success, got %v", report.Outcome)
	}

	after, err := facts.Lifecycle(l, "SPEC-001")
	if err != nil {
		t.Fatal(err)
	}
	if after != approval.Draft {
		t.Fatalf("expected SPEC-001 to derive Draft after revocation, got %v", after)
	}
}

// TestRevoke_NextRecommendsSpecWorkspace proves Revoke points at the Specification workspace
// rather than any dedicated command — there is no dedicated "define" command to name, and no
// single workflow is uniquely unlocked by returning to Draft the way Approved uniquely unlocks
// Implementation.
func TestRevoke_NextRecommendsSpecWorkspace(t *testing.T) {
	root := t.TempDir()
	configureGitIdentity(t, root)
	if _, err := Init(root); err != nil {
		t.Fatal(err)
	}
	if _, err := SpecCreate(root, "Password Reset"); err != nil {
		t.Fatal(err)
	}
	if _, err := Approve(root, "SPEC-001", nil); err != nil {
		t.Fatal(err)
	}

	report, err := Revoke(root, "SPEC-001", nil)
	if err != nil {
		t.Fatalf("revoke: %v", err)
	}
	if !strings.Contains(report.Next, "gnomon spec SPEC-001") {
		t.Fatalf("expected the Specification workspace recommended, got %q", report.Next)
	}
}

func TestRevoke_SpecificationContentUnmodified(t *testing.T) {
	root := t.TempDir()
	configureGitIdentity(t, root)
	if _, err := Init(root); err != nil {
		t.Fatal(err)
	}
	if _, err := SpecCreate(root, "Password Reset"); err != nil {
		t.Fatal(err)
	}
	if _, err := Approve(root, "SPEC-001", nil); err != nil {
		t.Fatal(err)
	}

	l, err := project.Locate(root)
	if err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(filepath.Join(l.SpecificationsDir(), "SPEC-001-password-reset.md"))
	if err != nil {
		t.Fatal(err)
	}

	if _, err := Revoke(root, "SPEC-001", nil); err != nil {
		t.Fatal(err)
	}

	after, err := os.ReadFile(filepath.Join(l.SpecificationsDir(), "SPEC-001-password-reset.md"))
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != string(after) {
		t.Fatalf("expected revocation to never modify the Specification's own content")
	}
}

func TestRevoke_UnrelatedGrantsUntouched(t *testing.T) {
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
	if _, err := Approve(root, "SPEC-001", nil); err != nil {
		t.Fatal(err)
	}
	if _, err := Approve(root, "SPEC-002", nil); err != nil {
		t.Fatal(err)
	}

	l, err := project.Locate(root)
	if err != nil {
		t.Fatal(err)
	}
	spec002Before, err := os.ReadDir(filepath.Join(l.ApprovalsDir(), "SPEC-002"))
	if err != nil {
		t.Fatal(err)
	}

	if _, err := Revoke(root, "SPEC-001", nil); err != nil {
		t.Fatal(err)
	}

	spec002After, err := os.ReadDir(filepath.Join(l.ApprovalsDir(), "SPEC-002"))
	if err != nil {
		t.Fatal(err)
	}
	if len(spec002Before) != len(spec002After) {
		t.Fatalf("expected SPEC-002's own approval evidence to be untouched by revoking SPEC-001: before=%d files, after=%d files", len(spec002Before), len(spec002After))
	}
	for i := range spec002Before {
		if spec002Before[i].Name() != spec002After[i].Name() {
			t.Fatalf("expected SPEC-002's approval evidence filenames to be unchanged, got a difference")
		}
	}

	state, err := facts.Lifecycle(l, "SPEC-002")
	if err != nil {
		t.Fatal(err)
	}
	if state != approval.Approved {
		t.Fatalf("expected SPEC-002 to remain Approved after revoking a different Specification's approval, got %v", state)
	}
}

func TestRevoke_GrantFileItselfUntouched(t *testing.T) {
	root := t.TempDir()
	configureGitIdentity(t, root)
	if _, err := Init(root); err != nil {
		t.Fatal(err)
	}
	if _, err := SpecCreate(root, "Password Reset"); err != nil {
		t.Fatal(err)
	}
	if _, err := Approve(root, "SPEC-001", nil); err != nil {
		t.Fatal(err)
	}

	l, err := project.Locate(root)
	if err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(l.ApprovalsDir(), "SPEC-001")
	entriesBefore, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	var grantName string
	for _, e := range entriesBefore {
		if strings.HasSuffix(e.Name(), ".grant.json") {
			grantName = e.Name()
		}
	}
	grantContentBefore, err := os.ReadFile(filepath.Join(dir, grantName))
	if err != nil {
		t.Fatal(err)
	}

	if _, err := Revoke(root, "SPEC-001", nil); err != nil {
		t.Fatal(err)
	}

	grantContentAfter, err := os.ReadFile(filepath.Join(dir, grantName))
	if err != nil {
		t.Fatalf("expected the original grant file to still exist, unmodified: %v", err)
	}
	if string(grantContentBefore) != string(grantContentAfter) {
		t.Fatalf("expected the grant file's own content to be byte-identical after revocation — revocation must never rewrite the grant it revokes")
	}

	entriesAfter, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entriesAfter) != len(entriesBefore)+1 {
		t.Fatalf("expected exactly one new file (the revocation record) to appear, got %d before, %d after", len(entriesBefore), len(entriesAfter))
	}
}

// --- Deterministic error/idempotency cases ---

func TestRevoke_NonexistentSpec(t *testing.T) {
	root := t.TempDir()
	configureGitIdentity(t, root)
	if _, err := Init(root); err != nil {
		t.Fatal(err)
	}

	report, err := Revoke(root, "SPEC-999", nil)
	if err == nil {
		t.Fatalf("expected an error for a nonexistent Specification")
	}
	if report != nil {
		t.Fatalf("expected no report for a nonexistent Specification, got: %+v", report)
	}
}

func TestRevoke_DraftNeverApproved_NoActiveApprovalToRevoke(t *testing.T) {
	root := t.TempDir()
	configureGitIdentity(t, root)
	if _, err := Init(root); err != nil {
		t.Fatal(err)
	}
	if _, err := SpecCreate(root, "Password Reset"); err != nil {
		t.Fatal(err)
	}

	report, err := Revoke(root, "SPEC-001", nil)
	if err == nil {
		t.Fatalf("expected an error: SPEC-001 was never approved, nothing to revoke")
	}
	if report == nil || report.Outcome != present.Blocked {
		t.Fatalf("expected a Blocked report, got: %+v", report)
	}
}

func TestRevoke_StaleFingerprint_NoActiveApprovalToRevoke(t *testing.T) {
	// The only grant no longer matches current content (a manual edit after approval) — the
	// Specification already effectively derives Draft, so there is no "grant currently causing
	// Approved" to revoke, per cli/APPROVAL_RUNTIME.md's Revocation Operation step 1.
	root := t.TempDir()
	configureGitIdentity(t, root)
	if _, err := Init(root); err != nil {
		t.Fatal(err)
	}
	if _, err := SpecCreate(root, "Password Reset"); err != nil {
		t.Fatal(err)
	}
	if _, err := Approve(root, "SPEC-001", nil); err != nil {
		t.Fatal(err)
	}

	l, err := project.Locate(root)
	if err != nil {
		t.Fatal(err)
	}
	specPath := filepath.Join(l.SpecificationsDir(), "SPEC-001-password-reset.md")
	content, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(specPath, append(content, []byte("\nA new sentence changing meaning.\n")...), 0o644); err != nil {
		t.Fatal(err)
	}

	state, err := facts.Lifecycle(l, "SPEC-001")
	if err != nil {
		t.Fatal(err)
	}
	if state != approval.Draft {
		t.Fatalf("expected the manual edit to already derive Draft, got %v", state)
	}

	report, err := Revoke(root, "SPEC-001", nil)
	if err == nil {
		t.Fatalf("expected an error: no grant currently matches this Specification's content")
	}
	if report == nil || report.Outcome != present.Blocked {
		t.Fatalf("expected a Blocked report, got: %+v", report)
	}
}

func TestRevoke_Repeated_SameDeterministicBlockedResult(t *testing.T) {
	root := t.TempDir()
	configureGitIdentity(t, root)
	if _, err := Init(root); err != nil {
		t.Fatal(err)
	}
	if _, err := SpecCreate(root, "Password Reset"); err != nil {
		t.Fatal(err)
	}
	if _, err := Approve(root, "SPEC-001", nil); err != nil {
		t.Fatal(err)
	}

	if _, err := Revoke(root, "SPEC-001", nil); err != nil {
		t.Fatalf("first revoke: %v", err)
	}

	l, err := project.Locate(root)
	if err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(l.ApprovalsDir(), "SPEC-001")
	entriesAfterFirst, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}

	report, err := Revoke(root, "SPEC-001", nil)
	if err == nil {
		t.Fatalf("expected the second revoke to refuse deterministically: nothing currently active to revoke")
	}
	if report == nil || report.Outcome != present.Blocked {
		t.Fatalf("expected a Blocked report on repeated revocation, got: %+v", report)
	}

	entriesAfterSecond, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entriesAfterFirst) != len(entriesAfterSecond) {
		t.Fatalf("expected repeated revocation to write nothing new: %d files after first, %d after second", len(entriesAfterFirst), len(entriesAfterSecond))
	}
}

func TestRevoke_MalformedGrantIsExcludedNotSilentlyRepaired(t *testing.T) {
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
	dir := filepath.Join(l.ApprovalsDir(), "SPEC-001")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "deadbeef.grant.json"), []byte("not valid json"), 0o644); err != nil {
		t.Fatal(err)
	}

	// A malformed grant is excluded from consideration entirely (existing, inherited Derive/
	// ActiveGrantID behavior) — never repaired, reinterpreted, or treated as a usable grant.
	report, err := Revoke(root, "SPEC-001", nil)
	if err == nil {
		t.Fatalf("expected refusal: the only evidence present is malformed, so there is nothing valid to revoke")
	}
	if report == nil || report.Outcome != present.Blocked {
		t.Fatalf("expected a Blocked report, got: %+v", report)
	}

	// The malformed file itself must still be untouched — not deleted, not rewritten.
	data, err := os.ReadFile(filepath.Join(dir, "deadbeef.grant.json"))
	if err != nil {
		t.Fatalf("expected the malformed file to still exist untouched: %v", err)
	}
	if string(data) != "not valid json" {
		t.Fatalf("expected the malformed file's content to be untouched, got: %q", string(data))
	}
}

func TestRevoke_RefusesWhenNoIdentityIsAvailable(t *testing.T) {
	root := t.TempDir()
	configureGitIdentity(t, root) // a real, configured identity, so Approve below can succeed

	if _, err := Init(root); err != nil {
		t.Fatal(err)
	}
	if _, err := SpecCreate(root, "Password Reset"); err != nil {
		t.Fatal(err)
	}
	if _, err := Approve(root, "SPEC-001", nil); err != nil {
		t.Fatalf("approve: %v", err)
	}

	// Now remove this repo's own git identity, and sandbox away any global fallback, so the
	// revoke call below has no identity to resolve and no prompt supplied — mirroring
	// TestApprove_RefusesWhenNoIdentityIsAvailable exactly, for the revocation path.
	unsetGitIdentity(t, root)
	t.Setenv("HOME", t.TempDir())
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	report, err := Revoke(root, "SPEC-001", nil)
	if err == nil {
		t.Fatalf("expected revoke to refuse when no Human identity can be resolved and no prompt is available")
	}
	if report != nil {
		t.Fatalf("expected no report when refusing for lack of identity, got: %+v", report)
	}
}

func unsetGitIdentity(t *testing.T, root string) {
	t.Helper()
	for _, key := range []string{"user.name", "user.email"} {
		cmd := exec.Command("git", "config", "--unset", key)
		cmd.Dir = root
		_ = cmd.Run() // best-effort; a missing key is not an error here
	}
}
