package adapter

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/term"
)

// fakeTerminalIO is the injectable terminalIO double every Part B test uses instead of a real TTY.
// It never distinguishes fd values by number (a swapped os.Stdout in a test has a different fd
// than the real one) — the only two call sites are Prepare's stdin check and finish's stdout
// check, so "is this the stdin fd" is enough to tell them apart.
type fakeTerminalIO struct {
	stdinIsTerminal  bool
	stdoutIsTerminal bool
	getStateErr      error
	restoreErr       error

	isTerminalFds  []int
	getStateCalled int
	restoreCalled  int
}

func (f *fakeTerminalIO) IsTerminal(fd int) bool {
	f.isTerminalFds = append(f.isTerminalFds, fd)
	if fd == int(os.Stdin.Fd()) {
		return f.stdinIsTerminal
	}
	return f.stdoutIsTerminal
}

func (f *fakeTerminalIO) GetState(fd int) (*term.State, error) {
	f.getStateCalled++
	if f.getStateErr != nil {
		return nil, f.getStateErr
	}
	return &term.State{}, nil
}

func (f *fakeTerminalIO) Restore(fd int, s *term.State) error {
	f.restoreCalled++
	return f.restoreErr
}

// prepareTestAdapter runs the real Prepare (so termState/promptPath are set exactly as production
// would) against executable "true", then swaps in a shell script for the actual test process —
// Prepare always builds exec.Command(executable, oneArg), which can't express an inline shell
// script directly.
func prepareTestAdapter(t *testing.T, fake *fakeTerminalIO, script string) *processAdapter {
	t.Helper()
	root := t.TempDir()
	a := &processAdapter{executable: "true", term: fake}
	ctx := Context{ProjectRoot: root, WorkflowIdentity: "implementation", ResultPath: filepath.Join(root, "result.json"), RunID: "r1"}
	if err := a.Prepare(ctx); err != nil {
		t.Fatal(err)
	}
	a.cmd = exec.Command("sh", "-c", script)
	a.cmd.Dir = root
	return a
}

// captureStdout swaps os.Stdout for the duration of fn and returns whatever was written to it —
// used so the reset-sequence write doesn't spill into the test runner's own terminal.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	orig := os.Stdout
	os.Stdout = w
	fn()
	w.Close()
	os.Stdout = orig
	out, _ := io.ReadAll(r)
	return string(out)
}

func TestProcessAdapter_NaturalExit_RestoresTerminalAndCleansUpPromptFile(t *testing.T) {
	fake := &fakeTerminalIO{stdinIsTerminal: true, stdoutIsTerminal: true}
	a := prepareTestAdapter(t, fake, "true")
	promptPath := a.promptPath

	out := captureStdout(t, func() {
		if err := a.Run(); err != nil {
			t.Fatal(err)
		}
	})

	if fake.getStateCalled != 1 || fake.restoreCalled != 1 {
		t.Fatalf("expected exactly one save and one restore, got GetState=%d Restore=%d", fake.getStateCalled, fake.restoreCalled)
	}
	if out != terminalResetSequence {
		t.Fatalf("expected the reset sequence written to stdout, got %q", out)
	}
	if _, err := os.Stat(promptPath); !os.IsNotExist(err) {
		t.Fatalf("expected the prompt file to be removed after a natural exit")
	}
}

func TestProcessAdapter_GracefulTermination_RestoresTerminal(t *testing.T) {
	fake := &fakeTerminalIO{stdinIsTerminal: true, stdoutIsTerminal: true}
	a := prepareTestAdapter(t, fake, "sleep 10")
	a.ctx.PollResult = func() (bool, error) { return true, nil } // ready on the very first tick

	captureStdout(t, func() {
		_ = a.Run()
	})

	if !a.status.Terminated {
		t.Fatalf("expected the run to report Terminated for a poll-triggered termination")
	}
	if fake.restoreCalled != 1 {
		t.Fatalf("expected exactly one restore after graceful termination, got %d", fake.restoreCalled)
	}
	if _, err := os.Stat(a.promptPath); !os.IsNotExist(err) {
		t.Fatalf("expected the prompt file to be removed after graceful termination")
	}
}

func TestProcessAdapter_ForcedKill_RestoresTerminal(t *testing.T) {
	// Reusing Part A's own seam: a signal that cannot be sent forces an immediate kill, without
	// waiting out the real 5s graceful timeout.
	origSignal := sendTerminationSignal
	sendTerminationSignal = func(*os.Process) error { return errSignalUnsupported }
	defer func() { sendTerminationSignal = origSignal }()

	fake := &fakeTerminalIO{stdinIsTerminal: true, stdoutIsTerminal: true}
	a := prepareTestAdapter(t, fake, "sleep 10")
	a.ctx.PollResult = func() (bool, error) { return true, nil }

	captureStdout(t, func() {
		_ = a.Run()
	})

	if fake.restoreCalled != 1 {
		t.Fatalf("expected exactly one restore after a forced kill, got %d", fake.restoreCalled)
	}
	if _, err := os.Stat(a.promptPath); !os.IsNotExist(err) {
		t.Fatalf("expected the prompt file to be removed after a forced kill")
	}
}

func TestProcessAdapter_Cancel_RestoresTerminal(t *testing.T) {
	fake := &fakeTerminalIO{stdinIsTerminal: true, stdoutIsTerminal: true}
	a := prepareTestAdapter(t, fake, "sleep 10")
	if err := a.cmd.Start(); err != nil {
		t.Fatal(err)
	}
	a.status.Started = true

	captureStdout(t, func() {
		_ = a.Cancel()
	})
	a.cmd.Wait() // reap, so the test process doesn't leave a zombie behind

	if fake.restoreCalled != 1 {
		t.Fatalf("expected exactly one restore after Cancel, got %d", fake.restoreCalled)
	}
	if _, err := os.Stat(a.promptPath); !os.IsNotExist(err) {
		t.Fatalf("expected the prompt file to be removed after Cancel")
	}
}

func TestProcessAdapter_NonTerminalStdin_NoSaveOrRestoreCalls(t *testing.T) {
	fake := &fakeTerminalIO{stdinIsTerminal: false, stdoutIsTerminal: false}
	a := prepareTestAdapter(t, fake, "true")

	captureStdout(t, func() {
		if err := a.Run(); err != nil {
			t.Fatal(err)
		}
	})

	if fake.getStateCalled != 0 || fake.restoreCalled != 0 {
		t.Fatalf("expected no GetState/Restore calls for non-terminal stdin, got GetState=%d Restore=%d", fake.getStateCalled, fake.restoreCalled)
	}
}

func TestProcessAdapter_RestoreError_RunStillSucceeds(t *testing.T) {
	fake := &fakeTerminalIO{stdinIsTerminal: true, stdoutIsTerminal: true, restoreErr: errRestoreFailed}
	a := prepareTestAdapter(t, fake, "true")

	captureStdout(t, func() {
		if err := a.Run(); err != nil {
			t.Fatalf("expected a restore failure to never fail the run, got %v", err)
		}
	})
}

var (
	errSignalUnsupported = fmt.Errorf("signal not supported on this platform")
	errRestoreFailed     = fmt.Errorf("failed to restore terminal state")
)

// --- Part C: prompt-file handoff ---

func TestPromptInstruction_SafeRelativePath(t *testing.T) {
	root := filepath.FromSlash("/tmp/proj")
	promptPath := filepath.Join(root, ".gnomon-runtime", "abc123.prompt.md")
	got := promptInstruction(root, promptPath)
	want := "Read and follow the instructions in .gnomon-runtime/abc123.prompt.md"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
	if !promptArgPattern.MatchString(got) {
		t.Fatalf("expected the argument to match the safe character set, got %q", got)
	}
}

func TestPromptInstruction_UnsafeCharacters_FallsBackToAbsolutePath(t *testing.T) {
	root := filepath.FromSlash("/tmp/proj")
	promptPath := filepath.Join(root, "weird#dir", "abc123.prompt.md")

	r, w, _ := os.Pipe()
	origStderr := os.Stderr
	os.Stderr = w
	got := promptInstruction(root, promptPath)
	w.Close()
	os.Stderr = origStderr
	stderr, _ := io.ReadAll(r)

	want := "Read and follow the instructions in " + promptPath
	if got != want {
		t.Fatalf("expected the fallback to use the absolute path verbatim, got %q, want %q", got, want)
	}
	if !strings.Contains(string(stderr), "debug") {
		t.Fatalf("expected a debug note on the fallback, got stderr: %q", stderr)
	}
}

func TestProcessAdapter_PromptFileContentEqualsBuildPrompt(t *testing.T) {
	root := t.TempDir()
	ctx := Context{ProjectRoot: root, WorkflowIdentity: "implementation", WorkflowPath: "x", SpecIdentity: "SPEC-001", ResultPath: filepath.Join(root, "result.json"), RunID: "r1"}

	a := &processAdapter{executable: "true", term: &fakeTerminalIO{}}
	if err := a.Prepare(ctx); err != nil {
		t.Fatal(err)
	}

	got, err := os.ReadFile(a.promptPath)
	if err != nil {
		t.Fatal(err)
	}
	want := buildPrompt(ctx)
	if string(got) != want {
		t.Fatalf("expected the prompt file's content to equal buildPrompt's output exactly")
	}
	if filepath.Base(a.promptPath) != "r1.prompt.md" {
		t.Fatalf("expected the prompt file named <run-id>.prompt.md, got %q", filepath.Base(a.promptPath))
	}
}

func TestProcessAdapter_ArgumentIsSingleLineAndSafe(t *testing.T) {
	root := t.TempDir()
	ctx := Context{ProjectRoot: root, WorkflowIdentity: "implementation", WorkflowPath: "x", ResultPath: filepath.Join(root, "result.json"), RunID: "r1"}

	a := &processAdapter{executable: "true", term: &fakeTerminalIO{}}
	if err := a.Prepare(ctx); err != nil {
		t.Fatal(err)
	}

	arg := a.cmd.Args[1]
	if strings.Contains(arg, "\n") {
		t.Fatalf("expected a single-line argument, got %q", arg)
	}
	if !promptArgPattern.MatchString(arg) {
		t.Fatalf("expected the argument to match the safe character set, got %q", arg)
	}
}

// TestProcessAdapter_HandoffSpecialCharactersOnlyInFile proves that finding text with newlines,
// quotes, &, |, and % — exactly what a resolution run's handoff can legitimately contain — never
// reaches the command-line argument, only the prompt file.
func TestProcessAdapter_HandoffSpecialCharactersOnlyInFile(t *testing.T) {
	root := t.TempDir()
	nasty := "line one\nline two \"quoted\" & piped | percent%done"
	ctx := Context{
		ProjectRoot: root, WorkflowIdentity: "implementation", WorkflowPath: "x", SpecIdentity: "SPEC-001",
		ResultPath: filepath.Join(root, "result.json"), RunID: "r1",
		Handoff: &ResolutionHandoff{OriginWorkflow: "review", OriginTarget: "src/", FindingID: "F-1", Classification: "DEFECT", Summary: nasty, Evidence: nasty},
	}

	a := &processAdapter{executable: "true", term: &fakeTerminalIO{}}
	if err := a.Prepare(ctx); err != nil {
		t.Fatal(err)
	}

	arg := a.cmd.Args[1]
	for _, bad := range []string{"\n", "\"", "&", "|", "%"} {
		if strings.Contains(arg, bad) {
			t.Fatalf("did not expect %q in the command-line argument, got %q", bad, arg)
		}
	}
	content, err := os.ReadFile(a.promptPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), nasty) {
		t.Fatalf("expected the handoff text to appear in the prompt file")
	}
}

func TestProcessAdapter_PromptFileRemoved_AfterFailure(t *testing.T) {
	fake := &fakeTerminalIO{}
	a := prepareTestAdapter(t, fake, "exit 1")
	promptPath := a.promptPath

	captureStdout(t, func() {
		if err := a.Run(); err == nil && a.status.ExitCode == 0 {
			t.Fatalf("expected a non-zero exit")
		}
	})

	if _, err := os.Stat(promptPath); !os.IsNotExist(err) {
		t.Fatalf("expected the prompt file to be removed after a failing run")
	}
}

func TestBuildPrompt_ScopeNotePassedThroughVerbatim(t *testing.T) {
	ctx := Context{WorkflowIdentity: "verification", WorkflowPath: "w.md", ResultPath: "r.json", RunID: "r1",
		ScopeNote: "Verification mode: standard check of the changes since abc.\nOther changed files (1): app/x.py"}
	p := buildPrompt(ctx)
	if !strings.Contains(p, "Scope for this run:\n"+ctx.ScopeNote+"\n") {
		t.Fatalf("expected the scope note verbatim in the prompt:\n%s", p)
	}
	if strings.Contains(buildPrompt(Context{WorkflowIdentity: "implementation"}), "Scope for this run") {
		t.Fatal("no scope section when there is no scope note")
	}
}
