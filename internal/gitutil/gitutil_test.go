package gitutil

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func newTempRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	run := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v failed: %v\n%s", args, err, out)
		}
	}
	run("init", "-q")
	return dir
}

func TestIdentity_ResolvesFromRealGitConfig(t *testing.T) {
	dir := newTempRepo(t)
	cmd := exec.Command("git", "config", "user.name", "Jane Doe")
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}
	cmd = exec.Command("git", "config", "user.email", "jane@example.com")
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}

	got, err := Identity(dir)
	if err != nil {
		t.Fatal(err)
	}
	if got != "Jane Doe <jane@example.com>" {
		t.Fatalf("unexpected identity: %q", got)
	}
}

func TestInspect_CleanRepo_AvailableNoChanges(t *testing.T) {
	dir := newTempRepo(t)
	if err := os.WriteFile(filepath.Join(dir, "committed.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	run := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v failed: %v\n%s", args, err, out)
		}
	}
	run("config", "user.name", "Jane Doe")
	run("config", "user.email", "jane@example.com")
	run("add", "committed.txt")
	run("commit", "-q", "-m", "initial")

	status, err := Inspect(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !status.Available {
		t.Fatalf("expected a real repository to be reported Available")
	}
	if status.Changed {
		t.Fatalf("expected a clean, fully-committed repo to report no changes")
	}
}

func TestInspect_UntrackedFile_Changed(t *testing.T) {
	dir := newTempRepo(t)
	if err := os.WriteFile(filepath.Join(dir, "untracked.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	status, err := Inspect(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !status.Available {
		t.Fatalf("expected a real repository to be reported Available")
	}
	if !status.Changed {
		t.Fatalf("expected an untracked file to be reported as a change")
	}
}

// TestInspect_NotAGitRepository_AvailableFalse_NoError proves "not a Git repository" is a
// legitimate, non-error state — Git is a capability Git-related guidance uses, not a prerequisite
// for inspecting a directory at all.
func TestInspect_NotAGitRepository_AvailableFalse_NoError(t *testing.T) {
	dir := t.TempDir()

	status, err := Inspect(dir)
	if err != nil {
		t.Fatalf("expected no error for a directory that simply isn't a Git repository, got: %v", err)
	}
	if status.Available {
		t.Fatalf("expected Available=false for a non-Git directory, got %+v", status)
	}
}

// TestInspect_UnexpectedFailure_ReturnsErrorWithDiagnostic proves a genuinely unexpected Git
// failure is never misclassified as "not a repository", and that its diagnostic is preserved
// rather than reduced to a bare exit code.
func TestInspect_UnexpectedFailure_ReturnsErrorWithDiagnostic(t *testing.T) {
	dir := newTempRepo(t)
	if err := os.WriteFile(filepath.Join(dir, "committed.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	run := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v failed: %v\n%s", args, err, out)
		}
	}
	run("config", "user.name", "Jane Doe")
	run("config", "user.email", "jane@example.com")
	run("add", "committed.txt")
	run("commit", "-q", "-m", "initial")
	// Corrupt the index so `git rev-parse --git-dir` still succeeds (it is still a Git
	// repository) but `git status` fails for a reason that is NOT "not a repository".
	if err := os.WriteFile(filepath.Join(dir, ".git", "index"), []byte("garbage not an index"), 0o644); err != nil {
		t.Fatal(err)
	}

	status, err := Inspect(dir)
	if err == nil {
		t.Fatalf("expected a corrupted repository to surface an unexpected error, got status=%+v", status)
	}
	if strings.Contains(err.Error(), "not a git repository") {
		t.Fatalf("did not expect an unexpected failure to be misclassified as 'not a repository': %v", err)
	}
	if err.Error() == "running git status: exit status 128" {
		t.Fatalf("expected a real diagnostic beyond a bare exit code, got: %v", err)
	}
}

func TestIdentity_NoConfig_ReturnsErrNoIdentity(t *testing.T) {
	dir := newTempRepo(t)
	// Isolate from any ambient global/system git config that might otherwise supply an identity.
	t.Setenv("HOME", t.TempDir())
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	if _, err := Identity(dir); err != ErrNoIdentity {
		t.Fatalf("expected ErrNoIdentity, got %v", err)
	}
}
