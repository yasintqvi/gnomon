package adapter

import (
	"os"
	"os/exec"
	"syscall"
	"time"
)

// terminationSignal asks the Agent process to end gracefully, giving it a chance to restore
// terminal state and clean up its own children on its own.
//
// Deliberately no process-group management (no Setpgid, no signaling a negative pid): that would
// move the Agent into a different process group than the terminal's foreground one, risking Ctrl+C
// no longer reaching it during the run.
var terminationSignal os.Signal = syscall.SIGTERM

// sendTerminationSignal sends terminationSignal to p — a package-level seam so tests can simulate
// a platform where the signal itself fails (e.g. Windows, where SIGTERM is not supported) without
// needing an actual unsupported platform to run on.
var sendTerminationSignal = func(p *os.Process) error {
	return p.Signal(terminationSignal)
}

// pollInterval is how often Gnomon checks whether a valid terminal result now exists while the
// Agent process is still running.
const pollInterval = 500 * time.Millisecond

// gracefulTerminationTimeout is how long Gnomon waits, after asking the Agent process to
// terminate, before escalating to a forced kill.
const gracefulTerminationTimeout = 5 * time.Second

// runWithPolling starts cmd and, if pollResult is non-nil, concurrently polls it at interval. As
// soon as it reports ready, the process is terminated gracefully; if it exits on its own first,
// that natural exit is reported as-is. A pollResult error or ready=false is always "not ready yet"
// — an incomplete or mismatched result is retried next tick, never mistaken for failure.
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

// terminateGracefully asks the process to end, waits up to timeout, and only then force-kills. If
// the signal itself can't be sent (e.g. SIGTERM unsupported on this platform), it force-kills
// immediately instead of waiting out a certain timeout — decided from the returned error, never
// runtime.GOOS.
func terminateGracefully(cmd *exec.Cmd, exitCh <-chan error, timeout time.Duration) error {
	if cmd.Process == nil {
		return <-exitCh
	}
	if err := sendTerminationSignal(cmd.Process); err != nil {
		_ = forceKill(cmd)
		return <-exitCh
	}
	select {
	case err := <-exitCh:
		return err
	case <-time.After(timeout):
		_ = forceKill(cmd)
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
