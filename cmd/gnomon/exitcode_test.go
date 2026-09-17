package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// exitCodeTestBinary is built once, into a temp directory, by TestMain below — subprocess exit
// codes are the one behavior that genuinely cannot be observed by calling package functions
// directly (os.Exit terminates the calling process), so this is real, black-box verification of
// the documented cli/COMMAND_SURFACE.md "Process Exit Codes" contract: 0 = Success, non-zero =
// anything else (Blocked, Cancelled, Failed, or invalid usage).
var exitCodeTestBinary string

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "gnomon-exitcode-test")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)

	exitCodeTestBinary = filepath.Join(dir, "gnomon")
	build := exec.Command("go", "build", "-o", exitCodeTestBinary, ".")
	if out, err := build.CombinedOutput(); err != nil {
		panic("building the exit-code test binary: " + err.Error() + "\n" + string(out))
	}

	os.Exit(m.Run())
}

func runGnomon(t *testing.T, dir string, env []string, args ...string) int {
	t.Helper()
	cmd := exec.Command(exitCodeTestBinary, args...)
	cmd.Dir = dir
	cmd.Env = env
	_ = cmd.Run()
	if cmd.ProcessState == nil {
		t.Fatalf("gnomon %v never produced a process state", args)
	}
	return cmd.ProcessState.ExitCode()
}

func gitInitializedRepo(t *testing.T) (dir string, env []string) {
	t.Helper()
	dir = t.TempDir()
	home := t.TempDir()
	env = append(os.Environ(), "HOME="+home, "XDG_CONFIG_HOME="+filepath.Join(home, ".config"))

	run := func(args ...string) {
		c := exec.Command("git", args...)
		c.Dir = dir
		c.Env = env
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	run("init", "-q")
	run("config", "user.name", "Jane Doe")
	run("config", "user.email", "jane@example.com")
	return dir, env
}

func TestExitCode_Success_IsZero(t *testing.T) {
	dir, env := gitInitializedRepo(t)
	if got := runGnomon(t, dir, env, "init"); got != 0 {
		t.Fatalf("expected a Success result (gnomon init on a fresh repo) to exit 0, got %d", got)
	}
}

func TestExitCode_Blocked_IsNonZero(t *testing.T) {
	dir, env := gitInitializedRepo(t)
	if got := runGnomon(t, dir, env, "init"); got != 0 {
		t.Fatalf("setup: init exit %d", got)
	}
	if got := runGnomon(t, dir, env, "spec", "create", "Password Reset"); got != 0 {
		t.Fatalf("setup: spec create exit %d", got)
	}
	// SPEC-001 remains Draft — implement must report Blocked, not Success.
	got := runGnomon(t, dir, env, "implement", "SPEC-001")
	if got == 0 {
		t.Fatalf("expected a Blocked result (implement on a Draft Specification) to exit non-zero, got 0")
	}
}

func TestExitCode_InvalidUsage_IsNonZero(t *testing.T) {
	dir, env := gitInitializedRepo(t)
	if got := runGnomon(t, dir, env, "implement"); got == 0 {
		t.Fatalf("expected a missing required argument to exit non-zero, got 0")
	}
	if got := runGnomon(t, dir, env, "not-a-real-command"); got == 0 {
		t.Fatalf("expected an unknown command to exit non-zero, got 0")
	}
}

func TestExitCode_MissingProject_IsNonZero(t *testing.T) {
	dir, env := gitInitializedRepo(t)
	if got := runGnomon(t, dir, env, "status"); got == 0 {
		t.Fatalf("expected status with no .gnomon project to exit non-zero, got 0")
	}
}

// TestExitCode_TwoTiersOnly documents, in one place, exactly what the process exit code does and
// does not distinguish — guarding against a future change silently introducing a richer taxonomy
// (which cli/COMMAND_SURFACE.md's "Process Exit Codes" section explicitly defers, not decides).
func TestExitCode_TwoTiersOnly(t *testing.T) {
	dir, env := gitInitializedRepo(t)
	if got := runGnomon(t, dir, env, "init"); got != 0 {
		t.Fatalf("setup: init exit %d", got)
	}
	if got := runGnomon(t, dir, env, "spec", "create", "Password Reset"); got != 0 {
		t.Fatalf("setup: spec create exit %d", got)
	}

	blockedCode := runGnomon(t, dir, env, "implement", "SPEC-001")
	invalidUsageCode := runGnomon(t, dir, env, "implement")
	unknownCmdCode := runGnomon(t, dir, env, "bogus")

	if blockedCode == 0 || invalidUsageCode == 0 || unknownCmdCode == 0 {
		t.Fatalf("expected every non-Success case to be non-zero: blocked=%d invalidUsage=%d unknownCmd=%d", blockedCode, invalidUsageCode, unknownCmdCode)
	}
	// v1 makes no promise these are equal, nor that they differ — only that each is non-zero.
	// This test exists to be revisited, not extended, if a richer taxonomy is ever adopted.
}
