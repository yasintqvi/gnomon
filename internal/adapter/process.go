package adapter

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"golang.org/x/term"

	"gnomon/internal/result"
)

// terminalIO is the small terminal state seam Prepare/Run/Cancel use to save and restore the
// controlling terminal's mode around the Agent process — separated from the real golang.org/x/term
// calls only so tests never need a real TTY.
type terminalIO interface {
	IsTerminal(fd int) bool
	GetState(fd int) (*term.State, error)
	Restore(fd int, state *term.State) error
}

type realTerminalIO struct{}

func (realTerminalIO) IsTerminal(fd int) bool               { return term.IsTerminal(fd) }
func (realTerminalIO) GetState(fd int) (*term.State, error) { return term.GetState(fd) }
func (realTerminalIO) Restore(fd int, s *term.State) error  { return term.Restore(fd, s) }

// terminalResetSequence shows the cursor and leaves the alternate screen — the minimal cleanup for
// a full-screen Agent UI (Claude Code, Codex) that was force-killed before it could restore the
// terminal itself. Nothing else is written.
const terminalResetSequence = "\x1b[?25h\x1b[?1049l"

// promptArgPattern is the safe character set for the single-line Agent argument (Part C): letters,
// digits, spaces, and a few path-safe punctuation characters. Anything else falls back to the
// absolute prompt path rather than risk a shell or Windows .cmd shim misinterpreting it.
var promptArgPattern = regexp.MustCompile(`^[A-Za-z0-9 ._/-]+$`)

// processAdapter is the Terminal Handoff mechanism every Adapter shares: launch one interactive
// subprocess with inherited stdio, and let Gnomon poll for a valid Result Protocol file rather than
// requiring the Agent to end its own session (run.go). Only the executable differs per provider.
type processAdapter struct {
	executable string
	ctx        Context
	cmd        *exec.Cmd
	status     Status

	term       terminalIO  // nil defaults to realTerminalIO{} (termIO()) — injectable for tests
	termState  *term.State // nil when stdin isn't a terminal, or saving it failed
	promptPath string      // removed by finish() on every exit path
	finishOnce sync.Once
}

func (a *processAdapter) termIO() terminalIO {
	if a.term != nil {
		return a.term
	}
	return realTerminalIO{}
}

// promptInstruction builds the single, short argument to pass the Agent instead of the full
// prompt text (Part C): a project-root-relative pointer at promptPath, forward-slashed. Falling
// back to the absolute path when the relative one has unexpected characters.
func promptInstruction(projectRoot, promptPath string) string {
	path := promptPath
	if rel, err := filepath.Rel(projectRoot, promptPath); err == nil {
		path = filepath.ToSlash(rel)
	}
	arg := fmt.Sprintf("Read and follow the instructions in %s", path)
	if promptArgPattern.MatchString(arg) {
		return arg
	}
	// Unsafe characters in the relative path — fall back to the absolute path rather than risk a
	// shell or Windows .cmd shim misinterpreting it. Never fails the run; only a debug note.
	fmt.Fprintf(os.Stderr, "gnomon: debug: prompt argument path had unexpected characters (%q); using the absolute path instead\n", path)
	return fmt.Sprintf("Read and follow the instructions in %s", promptPath)
}

func (a *processAdapter) Prepare(ctx Context) error {
	a.ctx = ctx

	dir, err := result.TransientDir(ctx.ProjectRoot)
	if err != nil {
		return err
	}
	promptPath := filepath.Join(dir, ctx.RunID+".prompt.md")
	if err := os.WriteFile(promptPath, []byte(buildPrompt(ctx)), 0o644); err != nil {
		return err
	}
	a.promptPath = promptPath

	a.cmd = exec.Command(a.executable, promptInstruction(ctx.ProjectRoot, promptPath))
	a.cmd.Dir = ctx.ProjectRoot
	a.cmd.Stdin = os.Stdin
	a.cmd.Stdout = os.Stdout
	a.cmd.Stderr = os.Stderr

	if a.termIO().IsTerminal(int(os.Stdin.Fd())) {
		if state, err := a.termIO().GetState(int(os.Stdin.Fd())); err == nil {
			a.termState = state
		}
	}
	return nil
}

// Run hands the terminal to the Agent for the run's duration while watching for a valid result via
// Context.PollResult; if the Human ends the session directly instead, that natural exit is
// reported as-is. The terminal is restored and the prompt file removed on every exit path (finish).
func (a *processAdapter) Run() error {
	if a.cmd == nil {
		return fmt.Errorf("adapter not prepared")
	}
	defer a.finish()
	a.status = runWithPolling(a.cmd, pollInterval, a.ctx.PollResult)
	return a.status.Err
}

func (a *processAdapter) Cancel() error {
	if a.cmd == nil || a.cmd.Process == nil {
		return nil
	}
	a.status.Terminated = true
	defer a.finish()
	return forceKill(a.cmd)
}

// finish restores the terminal (a full-screen Agent UI may have left it in raw mode / hidden
// cursor) and removes the prompt file, on every exit path — Run and Cancel. Guarded so a
// concurrent Run/Cancel pair only does this once. Every failure here is ignored.
func (a *processAdapter) finish() {
	a.finishOnce.Do(func() {
		if a.termState != nil {
			_ = a.termIO().Restore(int(os.Stdin.Fd()), a.termState)
		}
		if a.termIO().IsTerminal(int(os.Stdout.Fd())) {
			_, _ = os.Stdout.WriteString(terminalResetSequence)
		}
		if a.promptPath != "" {
			_ = os.Remove(a.promptPath)
		}
	})
}

func (a *processAdapter) Status() Status { return a.status }

// buildPrompt is the reserved Step 4/5 instruction: where/how to publish the result, pointing back
// at the workflow's own frontmatter rather than restating its schema. Never instructs the Agent to
// exit, and never names a provider — Gnomon ends the session itself once the result validates.
func buildPrompt(ctx Context) string {
	var context strings.Builder
	switch {
	case ctx.SpecIdentity != "":
		fmt.Fprintf(&context, "\nThe governing Specification is: %s\n", ctx.SpecIdentity)
	case ctx.Target != "":
		fmt.Fprintf(&context, "\nThe target for this run is: %s\n", ctx.Target)
	}
	// A Handoff is orthogonal to SpecIdentity/Target (can accompany either or neither), so it's
	// appended rather than folded into the same case — it explains why the run started, never a
	// substitute for SpecIdentity/Target's own requirements.
	if ctx.Handoff != nil {
		fmt.Fprintf(&context, "\nThis run was launched to resolve a finding from an earlier %s of %s. Address it while following this workflow's own normal process — the finding explains why this run started; it does not replace this workflow's own target or requirements. If resolving it requires a Human decision or any other material input, ask the Human directly in this session:\n\nFinding: %s — %s\nSummary: %s\nEvidence: %s\n",
			ctx.Handoff.OriginWorkflow, ctx.Handoff.OriginTarget, ctx.Handoff.FindingID, ctx.Handoff.Classification, ctx.Handoff.Summary, ctx.Handoff.Evidence)
		fmt.Fprintf(&context, "\nObjective: address this finding while following the %s workflow's own process.\n", ctx.WorkflowIdentity)
	}
	return fmt.Sprintf(`You are executing the Gnomon workflow %q for this project.

Read and follow the workflow instructions at: %s
%s
Once you have completed the workflow, publish exactly one valid JSON result to this path:
%s

The JSON object must have exactly this envelope shape:
{"workflow": %q, "run_id": %q, "payload": { ... }}

The "payload" must conform to the Result Contract declared in this workflow file's own frontmatter (the "result" block at the top of %s) — follow that schema exactly; do not invent fields it does not declare. Write the file exactly once.

Gnomon is watching for that file and will end this session automatically once it appears and validates — you do not need to exit or end the conversation yourself.`,
		ctx.WorkflowIdentity, ctx.WorkflowPath, context.String(), ctx.ResultPath, ctx.WorkflowIdentity, ctx.RunID, ctx.WorkflowPath)
}
