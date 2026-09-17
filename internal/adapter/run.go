package adapter

import (
	"os"
	"os/exec"
	"syscall"
	"time"
)

// terminationSignal is sent to ask the Agent process to end gracefully. SIGTERM (rather than a
// forced kill) is what gives the process a chance to restore terminal state and clean up its own
// child processes on its own, exactly as it would on any other exit path.
//
// Deliberately no process-group management (no Setpgid, no signaling a negative pid) is used
// here: doing so would move the Agent into a different process group than the one the controlling
// terminal treats as foreground, which risks Ctrl+C and similar terminal-generated signals no
// longer reaching it correctly during the run — a real regression to the interactive experience
// this model exists to preserve. A single, direct signal to the one child process is smaller,
// avoids that risk, and matches how the process would be asked to stop by any other path (a
// Human's own Ctrl+C, or a normal exit) — if the Agent needs to clean up its own child
// processes (e.g. a tool-use subprocess), that is its own responsibility on any termination path,
// not something Gnomon reaches around it to do.
var terminationSignal os.Signal = syscall.SIGTERM

// pollInterval is how often Gnomon checks whether a valid terminal result now exists while the
// Agent process is still running.
const pollInterval = 500 * time.Millisecond

// gracefulTerminationTimeout is how long Gnomon waits, after asking the Agent process to
// terminate, before escalating to a forced kill.
const gracefulTerminationTimeout = 5 * time.Second

// runWithPolling starts cmd and, if pollResult is non-nil, concurrently polls it at interval
// while the process is still running. As soon as pollResult reports the result is ready, the
// process is terminated gracefully and awaited; if the process exits on its own first (normal
// completion, crash, or a Human ending it directly), that natural exit is reported as-is. A
// pollResult error, or ready=false, is always treated identically to "not ready yet" — this is
// what lets an incomplete or mismatched result (still being written, or belonging to a different
// workflow/run) be retried on the next tick rather than mistaken for failure.
//
// This is exported at package scope (not tied to ClaudeAdapter) specifically so it can be tested
// deterministically against a small, controllable real process, without depending on the actual
// Claude Code binary.
func runWithPolling(cmd *exec.Cmd, interval time.Duration, pollResult func() (bool, error)) Status {
	var status Status
	if err := cmd.Start(); err != nil {
		status.Err = err
		return status
	}
	status.Started = true

	exitCh := make(chan error, 1)
	go func() { exitCh <- cmd.Wait() }()

	if pollResult == nil {
		recordExit(&status, <-exitCh)
		return status
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case err := <-exitCh:
			// The process ended on its own — nothing left to terminate. Whatever result exists
			// (or doesn't) is Gnomon's to consume afterward, exactly as if polling had never run.
			recordExit(&status, err)
			return status
		case <-ticker.C:
			ready, err := pollResult()
			if err != nil || !ready {
				continue
			}
			status.Terminated = true
			recordExit(&status, terminateGracefully(cmd, exitCh, gracefulTerminationTimeout))
			return status
		}
	}
}

// terminateGracefully asks the process to end, waits up to timeout for it to actually exit, and
// only escalates to a forced kill if it does not. A graceful signal first (rather than jumping
// straight to a forced kill) is what gives the Agent process a chance to restore terminal state
// on its own, exactly as it would on any other exit path.
func terminateGracefully(cmd *exec.Cmd, exitCh <-chan error, timeout time.Duration) error {
	if cmd.Process == nil {
		return <-exitCh
	}
	_ = cmd.Process.Signal(terminationSignal)
	select {
	case err := <-exitCh:
		return err
	case <-time.After(timeout):
		_ = cmd.Process.Kill()
		return <-exitCh
	}
}

func recordExit(status *Status, err error) {
	status.Exited = true
	if err == nil {
		return
	}
	if exitErr, ok := err.(*exec.ExitError); ok {
		status.ExitCode = exitErr.ExitCode()
	} else {
		status.Err = err
	}
}
