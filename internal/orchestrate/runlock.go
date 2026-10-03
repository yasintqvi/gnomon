package orchestrate

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gnomon/internal/present"
	"gnomon/internal/project"
	"gnomon/internal/result"
)

// --- Run lock: one Agent run at a time per project ---
//
// The lock is an operating-system lock on run.lock in the project's transient directory (flock on
// Unix, LockFileEx on Windows), never the file's mere existence. The OS releases it when the
// holding process ends for any reason — normal return, failure, Ctrl+C, or a kill — so a crashed
// run can never leave the project permanently locked. The file itself stays in place; its content
// only describes the current holder so a refused run can name it, and is cleared on release.

const runLockName = "run.lock"

// runLock describes the run holding the lock. Informational only: whether a run is active is
// decided by the OS lock alone, so stale content left by a killed process is simply overwritten.
type runLock struct {
	RunID            string `json:"run_id"`
	WorkflowIdentity string `json:"workflow_identity"`
	Target           string `json:"target,omitempty"`
	PID              int    `json:"pid"`
	Host             string `json:"host,omitempty"`
	StartedAt        string `json:"started_at"`
}

// RunActiveError is returned when another Agent run already holds the project's lock. Holder is
// nil when the other run has taken the lock but not yet written its description.
type RunActiveError struct {
	Path   string
	Holder *runLock
}

func (e *RunActiveError) Error() string {
	return "another Gnomon run is active in this project: " + e.describeHolder()
}

func (e *RunActiveError) describeHolder() string {
	if e.Holder == nil {
		return "a run is starting (its details are not written yet)"
	}
	h := e.Holder
	what := h.WorkflowIdentity
	if h.Target != "" {
		what += " " + h.Target
	}
	where := fmt.Sprintf("process %d", h.PID)
	if h.Host != "" {
		where += " on " + h.Host
	}
	return fmt.Sprintf("%s (run %s, %s, started %s)", what, h.RunID, where, h.StartedAt)
}

// runLockPath returns the lock file for the project containing root. Keyed on the located
// project root, not the working directory, so runs started from different subdirectories of the
// same project still exclude each other.
func runLockPath(root string) (string, error) {
	projectRoot := root
	if l, err := project.Locate(root); err == nil {
		projectRoot = l.Root
	}
	dir, err := result.TransientDir(projectRoot)
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, runLockName), nil
}

// acquireRunLock takes the project's run lock for the duration of one Agent run, or returns a
// *RunActiveError naming the run that holds it. Never waits. The caller must call release once
// the run is over; if the process dies first, the OS releases the lock instead.
func acquireRunLock(root, runID, workflowIdentity, target string) (release func(), err error) {
	f, err := lockProject(root)
	if err != nil {
		return nil, err
	}
	host, _ := os.Hostname()
	data, err := json.MarshalIndent(runLock{
		RunID:            runID,
		WorkflowIdentity: workflowIdentity,
		Target:           target,
		PID:              os.Getpid(),
		Host:             host,
		StartedAt:        time.Now().UTC().Format(time.RFC3339),
	}, "", "  ")
	if err == nil {
		err = writeLockContent(f, data)
	}
	if err != nil {
		unlockFile(f)
		f.Close()
		return nil, err
	}

	return func() {
		_ = writeLockContent(f, nil)
		unlockFile(f)
		f.Close()
	}, nil
}

// lockProject opens root's project lock file and locks it, returning the open, locked file.
func lockProject(root string) (*os.File, error) {
	path, err := runLockPath(root)
	if err != nil {
		return nil, err
	}
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0o644)
	if err != nil {
		return nil, err
	}
	if err := tryLockFile(f); err != nil {
		f.Close()
		if errors.Is(err, errLockHeld) {
			return nil, &RunActiveError{Path: path, Holder: readRunLock(path)}
		}
		return nil, fmt.Errorf("could not lock %s: %w", path, err)
	}
	return f, nil
}

func writeLockContent(f *os.File, data []byte) error {
	if err := f.Truncate(0); err != nil {
		return err
	}
	if _, err := f.WriteAt(data, 0); err != nil {
		return err
	}
	return f.Sync()
}

// readRunLock reads the holder's description, nil if it is missing, empty, or unreadable.
func readRunLock(path string) *runLock {
	data, err := os.ReadFile(path)
	if err != nil || len(strings.TrimSpace(string(data))) == 0 {
		return nil
	}
	var h runLock
	if json.Unmarshal(data, &h) != nil || h.RunID == "" {
		return nil
	}
	return &h
}

// probeRunLock reports the run holding root's project lock, nil if none. It takes and releases
// the lock at once, writing nothing; an error other than "held" counts as not held, leaving the
// real acquisition to report it.
func probeRunLock(root string) *RunActiveError {
	f, err := lockProject(root)
	if err != nil {
		var active *RunActiveError
		if errors.As(err, &active) {
			return active
		}
		return nil
	}
	unlockFile(f)
	f.Close()
	return nil
}

// refuseIfRunActive refuses when an Agent workflow run is in progress for root's project —
// approve/revoke call this first so a run in progress never races with an approval change
// mid-check.
func refuseIfRunActive(root string) error {
	if active := probeRunLock(root); active != nil {
		return fmt.Errorf("%s; approval changes are not allowed until it ends", active.Error())
	}
	return nil
}

// runActiveReport is shown instead of starting an Agent when another run holds the lock.
func runActiveReport(target string, e *RunActiveError) *present.Report {
	return &present.Report{
		Outcome: present.Blocked,
		Summary: "Another Gnomon run is active in this project",
		Target:  target,
		Sections: []present.Section{
			{Label: "Active Run", Body: e.describeHolder()},
			{Label: "Result", Body: "No Agent was started. Nothing in the project was changed."},
		},
		Next: "Wait for the active run to finish, or end it in its own terminal, then retry. " +
			"A run whose process has ended — even by a crash or kill — no longer blocks the project.",
	}
}
