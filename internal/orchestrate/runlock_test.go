package orchestrate

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"gnomon/internal/adapter"
	"gnomon/internal/present"
)

// runLockHelperEnv makes TestRunLockHelperProcess act as a separate gnomon process that holds the
// run lock of the project named by the variable until it is killed.
const runLockHelperEnv = "GNOMON_TEST_RUNLOCK_HELPER_ROOT"

func TestRunLockHelperProcess(t *testing.T) {
	root := os.Getenv(runLockHelperEnv)
	if root == "" {
		t.Skip("helper process for TestRunLock_OtherProcess…; runs only when started by it")
	}
	if _, err := acquireRunLock(root, "helper-run", "specification-definition", "SPEC-003"); err != nil {
		fmt.Println("ERROR", err)
		os.Exit(1)
	}
	fmt.Println("LOCKED")
	time.Sleep(2 * time.Minute) // killed by the parent long before this
	os.Exit(0)
}

// startLockHolder starts a separate process holding root's run lock and waits until it holds it.
func startLockHolder(t *testing.T, root string) *exec.Cmd {
	t.Helper()
	cmd := exec.Command(os.Args[0], "-test.run=^TestRunLockHelperProcess$")
	cmd.Env = append(os.Environ(), runLockHelperEnv+"="+root)
	out, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = cmd.Process.Kill(); _ = cmd.Wait() })

	locked := make(chan string, 1)
	go func() {
		sc := bufio.NewScanner(out)
		for sc.Scan() {
			if line := sc.Text(); line == "LOCKED" || strings.HasPrefix(line, "ERROR") {
				locked <- line
				return
			}
		}
		locked <- "helper exited without locking"
	}()
	select {
	case line := <-locked:
		if line != "LOCKED" {
			t.Fatalf("helper process: %s", line)
		}
	case <-time.After(20 * time.Second):
		t.Fatal("helper process did not take the lock in time")
	}
	return cmd
}

func successAdapter() *tamperingAdapter {
	return &tamperingAdapter{FakeAdapter: &adapter.FakeAdapter{}, Payload: implementationSuccessPayload()}
}

func assertRefusedNamingHolder(t *testing.T, rep *present.Report, err error, wantInBody ...string) {
	t.Helper()
	var active *RunActiveError
	if !errors.As(err, &active) {
		t.Fatalf("expected a *RunActiveError, got %v", err)
	}
	if rep == nil || rep.Outcome != present.Blocked {
		t.Fatalf("expected a Blocked report, got %+v", rep)
	}
	var body string
	for _, s := range rep.Sections {
		if s.Label == "Active Run" {
			body = s.Body
		}
	}
	for _, want := range wantInBody {
		if !strings.Contains(body, want) {
			t.Fatalf("expected the Active Run section to contain %q, got %q", want, body)
		}
	}
}

// A second run started while a first run's Agent is still working is refused before its own Agent
// is prepared, and names the first run. The first run then completes normally.
func TestRunLock_SecondRunDuringActiveRun_RefusedBeforeAgentStarts(t *testing.T) {
	root := setupApprovedSpec(t)
	l, wf := mustPrepareImplementation(t, root, "SPEC-001")

	second := successAdapter()
	var secondRep *present.Report
	var secondErr error
	first := successAdapter()
	first.Tamper = func() error {
		secondRep, secondErr = implementWithAdapter(root, l, wf, "SPEC-001", second)
		return nil
	}

	rep, err := implementWithAdapter(root, l, wf, "SPEC-001", first)
	if err != nil {
		t.Fatalf("first run: unexpected error: %v", err)
	}
	if rep.Outcome == present.Blocked || rep.Outcome == present.Failed {
		t.Fatalf("first run should complete normally, got %+v", rep)
	}

	assertRefusedNamingHolder(t, secondRep, secondErr, "implementation SPEC-001", first.Ctx.RunID)
	if second.Prepared {
		t.Fatal("the refused run must not prepare (launch) its Agent")
	}
	assertNoRunLock(t, root)
}

// The lock taken inside obtainOutcome (the path Discovery uses directly, without execute's early
// check) refuses too.
func TestRunLock_ObtainOutcomeRefusesWhileHeld(t *testing.T) {
	root := setupApprovedSpec(t)
	l, wf := mustPrepareImplementation(t, root, "SPEC-001")
	release, err := acquireRunLock(root, "run-A", "specification-definition", "SPEC-002")
	if err != nil {
		t.Fatal(err)
	}
	defer release()

	ad := successAdapter()
	_, _, failureRep, err := obtainOutcome(runRequest{root: root, l: l, wf: wf, kind: targetSpec, target: "SPEC-001", adapter: ad})
	assertRefusedNamingHolder(t, failureRep, err, "specification-definition SPEC-002", "run-A")
	if ad.Prepared {
		t.Fatal("the refused run must not prepare its Agent")
	}
}

// Many simultaneous attempts on one project: exactly one gets the lock.
func TestRunLock_SimultaneousAcquire_ExactlyOneWins(t *testing.T) {
	root := setupApprovedSpec(t)
	const n = 8
	var wg sync.WaitGroup
	start := make(chan struct{})
	releases := make(chan func(), n)
	refused := make(chan error, n)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-start
			release, err := acquireRunLock(root, fmt.Sprintf("run-%d", i), "implementation", "SPEC-001")
			if err != nil {
				refused <- err
				return
			}
			releases <- release
		}(i)
	}
	close(start)
	wg.Wait()
	close(releases)
	close(refused)

	if len(releases) != 1 {
		t.Fatalf("expected exactly one holder, got %d", len(releases))
	}
	for err := range refused {
		var active *RunActiveError
		if !errors.As(err, &active) {
			t.Fatalf("expected refusals to be *RunActiveError, got %v", err)
		}
	}
	for release := range releases {
		release()
	}
	assertNoRunLock(t, root)
}

// A run in another gnomon process blocks this project; when that process is killed — no cleanup
// code runs, its description stays in run.lock — the project is usable again at once.
func TestRunLock_OtherProcessHoldsLock_KilledProcessDoesNotBlock(t *testing.T) {
	root := setupApprovedSpec(t)
	l, wf := mustPrepareImplementation(t, root, "SPEC-001")
	holder := startLockHolder(t, root)

	ad := successAdapter()
	rep, err := implementWithAdapter(root, l, wf, "SPEC-001", ad)
	assertRefusedNamingHolder(t, rep, err, "specification-definition SPEC-003", "helper-run",
		fmt.Sprintf("process %d", holder.Process.Pid))
	if ad.Prepared {
		t.Fatal("the refused run must not prepare its Agent")
	}
	if _, err := Approve(root, "SPEC-001", nil, nil); err == nil || !strings.Contains(err.Error(), "helper-run") {
		t.Fatalf("expected approve to be refused naming the active run, got %v", err)
	}
	if _, err := Status(root); err != nil {
		t.Fatalf("status must keep working while a run is active: %v", err)
	}

	if err := holder.Process.Kill(); err != nil {
		t.Fatal(err)
	}
	_ = holder.Wait()

	path, err := runLockPath(root)
	if err != nil {
		t.Fatal(err)
	}
	if stale := readRunLock(path); stale == nil || stale.RunID != "helper-run" {
		t.Fatalf("expected the killed run's description to be left behind, got %+v", stale)
	}

	rep, err = implementWithAdapter(root, l, wf, "SPEC-001", successAdapter())
	if err != nil {
		t.Fatalf("expected the run to proceed after the holder was killed, got %v (%+v)", err, rep)
	}
	assertNoRunLock(t, root)
}

// Leftover run.lock content with no live holder (for example from a crash) never blocks.
func TestRunLock_StaleFileWithoutHolder_DoesNotBlock(t *testing.T) {
	root := setupApprovedSpec(t)
	path, err := runLockPath(root)
	if err != nil {
		t.Fatal(err)
	}
	stale := `{"run_id": "crashed-run", "workflow_identity": "implementation", "pid": 999999, "started_at": "2020-01-01T00:00:00Z"}`
	if err := os.WriteFile(path, []byte(stale), 0o644); err != nil {
		t.Fatal(err)
	}
	l, wf := mustPrepareImplementation(t, root, "SPEC-001")
	if _, err := implementWithAdapter(root, l, wf, "SPEC-001", successAdapter()); err != nil {
		t.Fatalf("expected a stale lock file to be ignored, got %v", err)
	}
	if _, err := Approve(root, "SPEC-001", nil, nil); err != nil && strings.Contains(err.Error(), "crashed-run") {
		t.Fatalf("approve must not be blocked by a stale lock file: %v", err)
	}
	assertNoRunLock(t, root)
}

func TestRunLock_DifferentProjectsAreIndependent(t *testing.T) {
	rootA := setupApprovedSpec(t)
	rootB := setupApprovedSpec(t)
	release, err := acquireRunLock(rootA, "run-A", "implementation", "SPEC-001")
	if err != nil {
		t.Fatal(err)
	}
	defer release()

	l, wf := mustPrepareImplementation(t, rootB, "SPEC-001")
	if _, err := implementWithAdapter(rootB, l, wf, "SPEC-001", successAdapter()); err != nil {
		t.Fatalf("a run in another project must not be blocked: %v", err)
	}
}

// Runs started from a subdirectory belong to the same project and share its lock.
func TestRunLock_SubdirectoryOfSameProjectConflicts(t *testing.T) {
	root := setupApprovedSpec(t)
	sub := filepath.Join(root, "src", "pkg")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	release, err := acquireRunLock(root, "run-A", "implementation", "SPEC-001")
	if err != nil {
		t.Fatal(err)
	}
	defer release()

	_, err = acquireRunLock(sub, "run-B", "verification", "src/")
	var active *RunActiveError
	if !errors.As(err, &active) || active.Holder == nil || active.Holder.RunID != "run-A" {
		t.Fatalf("expected the subdirectory run to be refused by run-A, got %v", err)
	}
}

func TestRunLock_StatusWorksWhileRunActive(t *testing.T) {
	root := setupApprovedSpec(t)
	release, err := acquireRunLock(root, "run-A", "implementation", "SPEC-001")
	if err != nil {
		t.Fatal(err)
	}
	defer release()
	if _, err := Status(root); err != nil {
		t.Fatalf("status must work while a run is active: %v", err)
	}
	if active := probeRunLock(root); active == nil {
		t.Fatal("status must not release or disturb the active run's lock")
	}
}
