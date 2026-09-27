package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// runGnomonOutput is runGnomon's sibling returning combined output too, for clearer failure
// messages in these end-to-end scenarios.
func runGnomonOutput(t *testing.T, dir string, env []string, args ...string) (int, string) {
	t.Helper()
	cmd := exec.Command(exitCodeTestBinary, args...)
	cmd.Dir = dir
	cmd.Env = env
	out, _ := cmd.CombinedOutput()
	if cmd.ProcessState == nil {
		t.Fatalf("gnomon %v never produced a process state", args)
	}
	return cmd.ProcessState.ExitCode(), string(out)
}

// specPathFor finds SPEC-001's file under a real, gnomon-init'd project — the filename includes a
// slug derived from the title, so this reads the directory rather than hardcoding it.
func specPathFor(t *testing.T, dir string) string {
	t.Helper()
	entries, err := os.ReadDir(filepath.Join(dir, ".gnomon", "specifications"))
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if len(e.Name()) >= 8 && e.Name()[:8] == "SPEC-001" {
			return filepath.Join(dir, ".gnomon", "specifications", e.Name())
		}
	}
	t.Fatalf("SPEC-001 file not found in %v", entries)
	return ""
}

// TestE2E_ContentRevertedToOlderApprovedVersion_Draft reproduces problem scenario 1 from the
// task: v1 approved, v2 approved, content reverted to exactly v1 — must be Draft, not
// re-Approved through the older grant.
func TestE2E_ContentRevertedToOlderApprovedVersion_Draft(t *testing.T) {
	dir, env := gitInitializedRepo(t)
	if got := runGnomon(t, dir, env, "init"); got != 0 {
		t.Fatalf("init exit %d", got)
	}
	if got := runGnomon(t, dir, env, "spec", "create", "Mark a task complete"); got != 0 {
		t.Fatalf("spec create exit %d", got)
	}
	specPath := specPathFor(t, dir)

	v1, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatal(err)
	}
	if got, out := runGnomonOutput(t, dir, env, "approve", "SPEC-001"); got != 0 {
		t.Fatalf("approve v1 exit %d: %s", got, out)
	}

	v2 := append(append([]byte{}, v1...), []byte("\nA v2 sentence.\n")...)
	if err := os.WriteFile(specPath, v2, 0o644); err != nil {
		t.Fatal(err)
	}
	if got, out := runGnomonOutput(t, dir, env, "approve", "SPEC-001"); got != 0 {
		t.Fatalf("approve v2 exit %d: %s", got, out)
	}

	// Revert to exactly v1 — an older, non-latest grant.
	if err := os.WriteFile(specPath, v1, 0o644); err != nil {
		t.Fatal(err)
	}

	// Draft has nothing to revoke — revoke must refuse (non-zero), proving the reverted content
	// is not Approved through v1's own, now-superseded grant.
	got, out := runGnomonOutput(t, dir, env, "revoke", "SPEC-001")
	if got == 0 {
		t.Fatalf("expected Draft (nothing to revoke) after reverting to an older approved version, got exit 0: %s", out)
	}
}

// TestE2E_RevokingLatestNeverResurrectsOlderApproval reproduces problem scenario 2 from the task:
// v1 approved, v2 approved, v2's grant revoked, content set back to v1 — must be Draft, not
// re-Approved through v1's own still-unrevoked grant.
func TestE2E_RevokingLatestNeverResurrectsOlderApproval(t *testing.T) {
	dir, env := gitInitializedRepo(t)
	if got := runGnomon(t, dir, env, "init"); got != 0 {
		t.Fatalf("init exit %d", got)
	}
	if got := runGnomon(t, dir, env, "spec", "create", "Mark a task complete"); got != 0 {
		t.Fatalf("spec create exit %d", got)
	}
	specPath := specPathFor(t, dir)

	v1, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatal(err)
	}
	if got, out := runGnomonOutput(t, dir, env, "approve", "SPEC-001"); got != 0 {
		t.Fatalf("approve v1 exit %d: %s", got, out)
	}

	v2 := append(append([]byte{}, v1...), []byte("\nA v2 sentence.\n")...)
	if err := os.WriteFile(specPath, v2, 0o644); err != nil {
		t.Fatal(err)
	}
	if got, out := runGnomonOutput(t, dir, env, "approve", "SPEC-001"); got != 0 {
		t.Fatalf("approve v2 exit %d: %s", got, out)
	}

	// Revoke v2 (the latest, active grant).
	if got, out := runGnomonOutput(t, dir, env, "revoke", "SPEC-001"); got != 0 {
		t.Fatalf("revoke v2 exit %d: %s", got, out)
	}

	// Set content back to v1 — its own grant still exists, unrevoked, but must never become
	// active again once the latest grant has been revoked.
	if err := os.WriteFile(specPath, v1, 0o644); err != nil {
		t.Fatal(err)
	}

	got, out := runGnomonOutput(t, dir, env, "revoke", "SPEC-001")
	if got == 0 {
		t.Fatalf("expected Draft (nothing to revoke): revoking the latest grant must not resurrect v1's, got exit 0: %s", out)
	}
}
