package adapter

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"testing"
	"time"

	"gnomon/internal/contract"
	"gnomon/internal/result"
)

const testPollInterval = 100 * time.Millisecond

func testResultContract() contract.ResultContract {
	return contract.ResultContract{
		TerminalPath: "outcome",
		Schema: map[string]interface{}{
			"type":     "object",
			"required": []interface{}{"outcome"},
			"properties": map[string]interface{}{
				"outcome": map[string]interface{}{
					"type": "string",
					"enum": []interface{}{"IMPLEMENTATION_COMPLETE", "BLOCKED"},
				},
			},
		},
	}
}

func writeEnvelope(t *testing.T, path, workflow, runID string, payload map[string]interface{}) {
	t.Helper()
	payloadBytes, _ := json.Marshal(payload)
	data, err := json.Marshal(map[string]interface{}{
		"workflow": workflow,
		"run_id":   runID,
		"payload":  json.RawMessage(payloadBytes),
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
}

// pollFromRealConsume mirrors exactly what orchestrate.Implement wires up in production: a
// closure over the real, unmodified result.Consume, never a stand-in.
func pollFromRealConsume(path, workflow, runID string, rc contract.ResultContract, captured **result.Outcome) func() (bool, error) {
	return func() (bool, error) {
		outcome, err := result.Consume(path, workflow, runID, rc)
		if err != nil {
			return false, nil
		}
		*captured = &outcome
		return true, nil
	}
}

func TestRunWithPolling_ValidResultAppearingMidRun_TerminatesEarlyAndIsConsumed(t *testing.T) {
	dest := t.TempDir() + "/result.json"
	rc := testResultContract()
	var captured *result.Outcome

	go func() {
		time.Sleep(3 * testPollInterval) // let a couple of polls happen against a missing file first
		writeEnvelope(t, dest, "implementation", "run-1", map[string]interface{}{
			"outcome":   "IMPLEMENTATION_COMPLETE",
			"delivered": "a trivial change",
		})
	}()

	cmd := exec.Command("sleep", "10")
	start := time.Now()
	status := runWithPolling(cmd, testPollInterval, pollFromRealConsume(dest, "implementation", "run-1", rc, &captured))
	elapsed := time.Since(start)

	if !status.Started || !status.Exited {
		t.Fatalf("expected Started and Exited to both be true, got %+v", status)
	}
	if elapsed > 5*time.Second {
		t.Fatalf("expected early termination well before the 10s sleep would have finished naturally, took %v", elapsed)
	}
	if captured == nil {
		t.Fatalf("expected the polling closure to have captured a valid outcome")
	}
	if captured.Terminal != "IMPLEMENTATION_COMPLETE" {
		t.Fatalf("unexpected terminal value: %s", captured.Terminal)
	}
	if _, err := os.Stat(dest); !os.IsNotExist(err) {
		t.Fatalf("expected the transient result file to be deleted after consumption")
	}
}

func TestRunWithPolling_InvalidResultNeverTerminatesEarly(t *testing.T) {
	dest := t.TempDir() + "/result.json"
	if err := os.WriteFile(dest, []byte("not valid json"), 0o644); err != nil {
		t.Fatal(err)
	}
	rc := testResultContract()
	var captured *result.Outcome

	cmd := exec.Command("sleep", "1")
	start := time.Now()
	status := runWithPolling(cmd, testPollInterval, pollFromRealConsume(dest, "implementation", "run-1", rc, &captured))
	elapsed := time.Since(start)

	if captured != nil {
		t.Fatalf("expected no outcome to be captured from malformed content")
	}
	if status.ExitCode != 0 || status.Err != nil {
		t.Fatalf("expected the process to run to its own natural, clean completion, got %+v", status)
	}
	if elapsed < 900*time.Millisecond {
		t.Fatalf("expected the process to run its full natural duration (~1s), took only %v — suggests it was terminated early", elapsed)
	}
	// The malformed file must be left in place for diagnostics, per Result Protocol.
	if _, err := os.Stat(dest); err != nil {
		t.Fatalf("expected the malformed result file to remain on disk: %v", err)
	}
}

func TestRunWithPolling_WrongRunIDNeverTerminatesEarly(t *testing.T) {
	dest := t.TempDir() + "/result.json"
	rc := testResultContract()
	var captured *result.Outcome

	// A well-formed, valid envelope — but for a different run entirely.
	writeEnvelope(t, dest, "implementation", "some-other-run", map[string]interface{}{"outcome": "IMPLEMENTATION_COMPLETE"})

	cmd := exec.Command("sleep", "1")
	start := time.Now()
	status := runWithPolling(cmd, testPollInterval, pollFromRealConsume(dest, "implementation", "run-1", rc, &captured))
	elapsed := time.Since(start)

	if captured != nil {
		t.Fatalf("expected no outcome to be captured for a mismatched run_id")
	}
	if elapsed < 900*time.Millisecond {
		t.Fatalf("expected the process to run its full natural duration (~1s), took only %v", elapsed)
	}
	if status.ExitCode != 0 {
		t.Fatalf("expected a clean natural exit, got exit code %d", status.ExitCode)
	}
}

func TestRunWithPolling_WrongWorkflowNeverTerminatesEarly(t *testing.T) {
	dest := t.TempDir() + "/result.json"
	rc := testResultContract()
	var captured *result.Outcome

	writeEnvelope(t, dest, "verification", "run-1", map[string]interface{}{"outcome": "IMPLEMENTATION_COMPLETE"})

	cmd := exec.Command("sleep", "1")
	status := runWithPolling(cmd, testPollInterval, pollFromRealConsume(dest, "implementation", "run-1", rc, &captured))

	if captured != nil {
		t.Fatalf("expected no outcome to be captured for a mismatched workflow identity")
	}
	if status.ExitCode != 0 {
		t.Fatalf("expected a clean natural exit, got exit code %d", status.ExitCode)
	}
}

// startForTermination starts a real process and returns it alongside the same kind of exit
// channel runWithPolling itself feeds terminateGracefully — letting these tests call
// terminateGracefully directly, deterministically, without a real Agent binary.
func startForTermination(t *testing.T, name string, args ...string) (*exec.Cmd, chan error) {
	t.Helper()
	cmd := exec.Command(name, args...)
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	exitCh := make(chan error, 1)
	go func() { exitCh <- cmd.Wait() }()
	return cmd, exitCh
}

func TestTerminateGracefully_SignalFails_ForcesImmediately(t *testing.T) {
	cmd, exitCh := startForTermination(t, "sleep", "10")

	origSignal := sendTerminationSignal
	sendTerminationSignal = func(*os.Process) error { return fmt.Errorf("signal not supported on this platform") }
	defer func() { sendTerminationSignal = origSignal }()

	start := time.Now()
	if err := terminateGracefully(cmd, exitCh, 5*time.Second); err == nil {
		t.Fatalf("expected an error from the forced kill's exit status")
	}
	elapsed := time.Since(start)

	if elapsed > 2*time.Second {
		t.Fatalf("expected immediate forced termination when the signal cannot be sent, took %v (well under the 5s timeout)", elapsed)
	}
}

func TestTerminateGracefully_SignalSucceeds_ProcessExits_NoForcedKill(t *testing.T) {
	cmd, exitCh := startForTermination(t, "sleep", "10")

	killCalled := false
	origKill := forceKill
	forceKill = func(c *exec.Cmd) error { killCalled = true; return origKill(c) }
	defer func() { forceKill = origKill }()

	// "sleep" honors SIGTERM's default action (exit) — the real, un-injected signal is used here.
	if err := terminateGracefully(cmd, exitCh, 5*time.Second); err == nil {
		t.Fatalf("expected sleep's SIGTERM-terminated exit to be reported as an error")
	}
	if killCalled {
		t.Fatalf("expected no forced kill when the process honors the graceful signal")
	}
}

func TestTerminateGracefully_SignalIgnored_ForcedKillAfterTimeout(t *testing.T) {
	cmd, exitCh := startForTermination(t, "sh", "-c", "trap '' TERM; sleep 10")
	time.Sleep(100 * time.Millisecond) // let the trap actually install before signaling

	killCalled := false
	origKill := forceKill
	forceKill = func(c *exec.Cmd) error { killCalled = true; return origKill(c) }
	defer func() { forceKill = origKill }()

	timeout := 300 * time.Millisecond
	start := time.Now()
	terminateGracefully(cmd, exitCh, timeout)
	elapsed := time.Since(start)

	if !killCalled {
		t.Fatalf("expected a forced kill once the process ignored the signal and the timeout elapsed")
	}
	if elapsed < timeout {
		t.Fatalf("expected to wait out the timeout before forcing, took only %v", elapsed)
	}
	if elapsed > 5*time.Second {
		t.Fatalf("expected the short test timeout to be honored, not the production default, took %v", elapsed)
	}
}

func TestRunWithPolling_NilPollResult_WaitsForNaturalExit(t *testing.T) {
	cmd := exec.Command("sleep", "1")
	start := time.Now()
	status := runWithPolling(cmd, testPollInterval, nil)
	elapsed := time.Since(start)

	if elapsed < 900*time.Millisecond {
		t.Fatalf("expected the process to run its full natural duration with no poll function, took %v", elapsed)
	}
	if !status.Started || !status.Exited || status.ExitCode != 0 {
		t.Fatalf("unexpected status: %+v", status)
	}
}
