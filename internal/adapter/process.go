package adapter

import (
	"fmt"
	"os"
	"os/exec"
)

// processAdapter is the Terminal Handoff mechanism every current Adapter implementation shares:
// launch one interactive subprocess, inherit stdio so the Human's normal interactive UI remains
// available for the run's duration, and let Gnomon poll for a valid Result Protocol file rather
// than requiring the Agent to end its own session (run.go). None of this differs between
// providers — the only thing that does is which executable to start, supplied by each concrete
// Adapter's own constructor (claude.go, codex.go).
type processAdapter struct {
	executable string
	ctx        Context
	cmd        *exec.Cmd
	status     Status
}

func (a *processAdapter) Prepare(ctx Context) error {
	a.ctx = ctx
	a.cmd = exec.Command(a.executable, buildPrompt(ctx))
	a.cmd.Dir = ctx.ProjectRoot
	a.cmd.Stdin = os.Stdin
	a.cmd.Stdout = os.Stdout
	a.cmd.Stderr = os.Stderr
	return nil
}

// Run hands the terminal to the Agent for the run's duration — Gnomon does not observe or mediate
// the conversation — while concurrently watching for a valid terminal result via
// Context.PollResult. As soon as one appears, the process is terminated gracefully and control
// returns; if the Human ends the session directly instead, that natural exit is reported as-is.
func (a *processAdapter) Run() error {
	if a.cmd == nil {
		return fmt.Errorf("adapter not prepared")
	}
	a.status = runWithPolling(a.cmd, pollInterval, a.ctx.PollResult)
	return a.status.Err
}

func (a *processAdapter) Cancel() error {
	if a.cmd == nil || a.cmd.Process == nil {
		return nil
	}
	a.status.Terminated = true
	return a.cmd.Process.Kill()
}

func (a *processAdapter) Status() Status { return a.status }

// buildPrompt is the reserved Step 4/5 instruction: where/how to publish the result, pointing
// back at the workflow's own already-authoritative frontmatter rather than restating its schema.
// It deliberately does not instruct the Agent to exit, and deliberately never names a provider —
// Gnomon ends the session itself once the result is detected and validated, and this exact text
// is used unchanged regardless of which Agent is executing it.
func buildPrompt(ctx Context) string {
	context := ""
	switch {
	case ctx.SpecIdentity != "":
		context = fmt.Sprintf("\nThe governing Specification is: %s\n", ctx.SpecIdentity)
	case ctx.Target != "":
		context = fmt.Sprintf("\nThe target for this run is: %s\n", ctx.Target)
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
		ctx.WorkflowIdentity, ctx.WorkflowPath, context, ctx.ResultPath, ctx.WorkflowIdentity, ctx.RunID, ctx.WorkflowPath)
}
