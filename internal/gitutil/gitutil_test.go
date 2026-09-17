package gitutil

import (
	"os"
	"os/exec"
	"path/filepath"
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

func TestHasChanges_CleanRepo_False(t *testing.T) {
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

	changed, err := HasChanges(dir)
	if err != nil {
		t.Fatal(err)
	}
	if changed {
		t.Fatalf("expected a clean, fully-committed repo to report no changes")
	}
}

func TestHasChanges_UntrackedFile_True(t *testing.T) {
	dir := newTempRepo(t)
	if err := os.WriteFile(filepath.Join(dir, "untracked.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	changed, err := HasChanges(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !changed {
		t.Fatalf("expected an untracked file to be reported as a change")
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
