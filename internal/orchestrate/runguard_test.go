package orchestrate

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gnomon/internal/adapter"
	"gnomon/internal/approval"
	"gnomon/internal/facts"
	"gnomon/internal/present"
	"gnomon/internal/project"
)

// tamperingAdapter runs an arbitrary file-system mutation during Run, then (unless SkipResult)
// writes a valid scripted result exactly like payloadAdapter — proving the run guard rejects the
// result despite it being otherwise structurally valid.
type tamperingAdapter struct {
	*adapter.FakeAdapter
	Payload    map[string]interface{}
	Tamper     func() error
	SkipResult bool
}

func (s *tamperingAdapter) Run() error {
	if s.Tamper != nil {
		if err := s.Tamper(); err != nil {
			return err
		}
	}
	if s.SkipResult {
		return s.FakeAdapter.RunErr
	}
	payload, _ := json.Marshal(s.Payload)
	env, _ := json.Marshal(map[string]interface{}{
		"workflow": s.FakeAdapter.Ctx.WorkflowIdentity,
		"run_id":   s.FakeAdapter.Ctx.RunID,
		"payload":  json.RawMessage(payload),
	})
	s.FakeAdapter.ResultContent = env
	return s.FakeAdapter.Run()
}

func implementationSuccessPayload() map[string]interface{} {
	return map[string]interface{}{"outcome": "IMPLEMENTATION_COMPLETE", "delivered": "a trivial change"}
}

func writeForgedGrant(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	forged := approval.Grant{SpecIdentity: "SPEC-002", Fingerprint: "forged-fp", Approver: "Agent", Timestamp: "2024-01-01T00:00:00Z"}
	data, err := json.Marshal(forged)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
}

// --- 1. Agent adds a grant file ---

func TestRunGuard_AgentAddsGrantFile_RunRejectedFileRemovedSpecStillDraft(t *testing.T) {
	root := setupApprovedSpec(t) // SPEC-001 Approved
	if _, err := specCreateDefined(root, "Email Verification"); err != nil {
		t.Fatal(err) // SPEC-002, Draft
	}
	l, wf := mustPrepareImplementation(t, root, "SPEC-001")

	forgedPath := filepath.Join(l.ApprovalsDir(), "SPEC-002", "forged01.grant.json")
	ad := &tamperingAdapter{
		FakeAdapter: &adapter.FakeAdapter{},
		Payload:     implementationSuccessPayload(),
		Tamper:      func() error { writeForgedGrant(t, forgedPath); return nil },
	}

	rep, err := implementWithAdapter(root, l, wf, "SPEC-001", ad)
	if err == nil {
		t.Fatalf("expected the run to be rejected")
	}
	if rep == nil || rep.Outcome != present.Failed {
		t.Fatalf("expected a Failed report, got: %+v", rep)
	}
	rendered := present.Render(rep, false)
	if !strings.Contains(rendered, "modified approval evidence") {
		t.Fatalf("unexpected report: %s", rendered)
	}
	if !strings.Contains(rendered, "added: SPEC-002/forged01.grant.json") {
		t.Fatalf("expected the offending file listed as added, got: %s", rendered)
	}
	if !strings.Contains(rendered, "restored") {
		t.Fatalf("expected the report to state the approvals directory was restored, got: %s", rendered)
	}

	if _, statErr := os.Stat(forgedPath); !os.IsNotExist(statErr) {
		t.Fatalf("expected the forged grant file removed, stat err: %v", statErr)
	}

	l2, err := project.Locate(root)
	if err != nil {
		t.Fatal(err)
	}
	state, err := facts.Lifecycle(l2, "SPEC-002")
	if err != nil {
		t.Fatal(err)
	}
	if state != approval.Draft {
		t.Fatalf("expected SPEC-002 to remain Draft, got %v", state)
	}
}

// --- 2. Agent modifies an existing grant ---

func TestRunGuard_AgentModifiesExistingGrant_OriginalBytesRestored(t *testing.T) {
	root := setupApprovedSpec(t)
	l, wf := mustPrepareImplementation(t, root, "SPEC-001")

	grantPath := findEvidenceFile(t, l, "SPEC-001", ".grant.json")
	original, err := os.ReadFile(grantPath)
	if err != nil {
		t.Fatal(err)
	}

	ad := &tamperingAdapter{
		FakeAdapter: &adapter.FakeAdapter{},
		Payload:     implementationSuccessPayload(),
		Tamper: func() error {
			return os.WriteFile(grantPath, append(append([]byte{}, original...), []byte("\ntampered")...), 0o644)
		},
	}

	rep, err := implementWithAdapter(root, l, wf, "SPEC-001", ad)
	if err == nil {
		t.Fatalf("expected the run to be rejected")
	}
	if rep == nil || rep.Outcome != present.Failed {
		t.Fatalf("expected a Failed report, got: %+v", rep)
	}

	after, err := os.ReadFile(grantPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != string(original) {
		t.Fatalf("expected the grant file's original bytes restored, got %q", after)
	}
}

// --- 3. Agent deletes a revocation ---

func TestRunGuard_AgentDeletesRevocation_FileRestoredSpecRemainsDraft(t *testing.T) {
	root := setupApprovedSpec(t) // SPEC-001 Approved
	if _, err := specCreateDefined(root, "Email Verification"); err != nil {
		t.Fatal(err)
	}
	if _, err := Approve(root, "SPEC-002", nil, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := Revoke(root, "SPEC-002", nil); err != nil {
		t.Fatal(err)
	}

	l, wf := mustPrepareImplementation(t, root, "SPEC-001")
	revocationPath := findEvidenceFile(t, l, "SPEC-002", ".revocation.json")
	original, err := os.ReadFile(revocationPath)
	if err != nil {
		t.Fatal(err)
	}

	ad := &tamperingAdapter{
		FakeAdapter: &adapter.FakeAdapter{},
		Payload:     implementationSuccessPayload(),
		Tamper:      func() error { return os.Remove(revocationPath) },
	}

	rep, err := implementWithAdapter(root, l, wf, "SPEC-001", ad)
	if err == nil {
		t.Fatalf("expected the run to be rejected")
	}
	if rep == nil || rep.Outcome != present.Failed {
		t.Fatalf("expected a Failed report, got: %+v", rep)
	}

	after, err := os.ReadFile(revocationPath)
	if err != nil {
		t.Fatalf("expected the revocation file restored: %v", err)
	}
	if string(after) != string(original) {
		t.Fatalf("expected byte-identical restoration, got %q", after)
	}

	l2, err := project.Locate(root)
	if err != nil {
		t.Fatal(err)
	}
	state, err := facts.Lifecycle(l2, "SPEC-002")
	if err != nil {
		t.Fatal(err)
	}
	if state != approval.Draft {
		t.Fatalf("expected SPEC-002 to remain Draft (revocation still in effect), got %v", state)
	}
}

// --- 4. Agent deletes the whole approvals directory ---

func TestRunGuard_AgentDeletesApprovalsDirectory_Restored(t *testing.T) {
	root := setupApprovedSpec(t)
	l, wf := mustPrepareImplementation(t, root, "SPEC-001")

	before, err := snapshotTree(l.ApprovalsDir())
	if err != nil {
		t.Fatal(err)
	}
	if len(before) == 0 {
		t.Fatalf("expected existing approval evidence to snapshot")
	}

	ad := &tamperingAdapter{
		FakeAdapter: &adapter.FakeAdapter{},
		Payload:     implementationSuccessPayload(),
		Tamper:      func() error { return os.RemoveAll(l.ApprovalsDir()) },
	}

	rep, err := implementWithAdapter(root, l, wf, "SPEC-001", ad)
	if err == nil {
		t.Fatalf("expected the run to be rejected")
	}
	if rep == nil || rep.Outcome != present.Failed {
		t.Fatalf("expected a Failed report, got: %+v", rep)
	}

	after, err := snapshotTree(l.ApprovalsDir())
	if err != nil {
		t.Fatal(err)
	}
	if len(after) != len(before) {
		t.Fatalf("expected every approval file restored, got %d of %d", len(after), len(before))
	}
	for path, data := range before {
		if string(after[path]) != string(data) {
			t.Fatalf("expected %s restored byte-identical", path)
		}
	}
}

// --- 5. Agent touches nothing: unaffected ---

func TestRunGuard_AgentTouchesNothing_ReportUnaffected(t *testing.T) {
	root := setupApprovedSpec(t)
	l, wf := mustPrepareImplementation(t, root, "SPEC-001")
	ad := &tamperingAdapter{FakeAdapter: &adapter.FakeAdapter{}, Payload: implementationSuccessPayload()}

	rep, err := implementWithAdapter(root, l, wf, "SPEC-001", ad)
	if err != nil {
		t.Fatalf("expected success, got err: %v (report: %+v)", err, rep)
	}
	if rep.Outcome != present.Success {
		t.Fatalf("expected Success, got %v", rep.Outcome)
	}
	if rep.Summary != "Implementation complete" {
		t.Fatalf("unexpected summary: %q", rep.Summary)
	}
}

// --- 6. Agent fails without a result, but still tampered ---

func TestRunGuard_AgentFailsWithoutResult_StillRestoredAndRejected(t *testing.T) {
	root := setupApprovedSpec(t)
	if _, err := specCreateDefined(root, "Email Verification"); err != nil {
		t.Fatal(err)
	}
	l, wf := mustPrepareImplementation(t, root, "SPEC-001")

	forgedPath := filepath.Join(l.ApprovalsDir(), "SPEC-002", "forged02.grant.json")
	ad := &tamperingAdapter{
		FakeAdapter: &adapter.FakeAdapter{RunErr: fmt.Errorf("agent crashed")},
		SkipResult:  true,
		Tamper:      func() error { writeForgedGrant(t, forgedPath); return nil },
	}

	rep, err := implementWithAdapter(root, l, wf, "SPEC-001", ad)
	if err == nil {
		t.Fatalf("expected the run to be rejected")
	}
	if rep == nil || rep.Outcome != present.Failed {
		t.Fatalf("expected the guard's Failed report to take priority over the plain process failure, got: %+v", rep)
	}
	if !strings.Contains(rep.Summary, "modified approval evidence") {
		t.Fatalf("expected the guard's own summary to take priority, got %q", rep.Summary)
	}
	if _, statErr := os.Stat(forgedPath); !os.IsNotExist(statErr) {
		t.Fatalf("expected the forged file removed")
	}
}

// --- 7. Run lock ---

func TestRefuseIfRunActive_LockPresent_ApproveAndRevokeRefuseWriteNothing(t *testing.T) {
	root := setupApprovedSpec(t) // SPEC-001 Approved
	l, err := project.Locate(root)
	if err != nil {
		t.Fatal(err)
	}

	unlock, err := acquireRunLock(root, "test-run-id", "implementation", "SPEC-001")
	if err != nil {
		t.Fatal(err)
	}
	defer unlock()

	dir := filepath.Join(l.ApprovalsDir(), "SPEC-001")
	before, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := Approve(root, "SPEC-001", nil, nil); err == nil {
		t.Fatalf("expected Approve to refuse while a run is active")
	}
	if _, err := Revoke(root, "SPEC-001", nil); err == nil {
		t.Fatalf("expected Revoke to refuse while a run is active")
	}

	after, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(before) != len(after) {
		t.Fatalf("expected no new evidence files written while refusing, before=%d after=%d", len(before), len(after))
	}
}

func TestRunLock_AbsentAfterNormalRun(t *testing.T) {
	root := setupApprovedSpec(t)
	l, wf := mustPrepareImplementation(t, root, "SPEC-001")
	ad := &tamperingAdapter{FakeAdapter: &adapter.FakeAdapter{}, Payload: implementationSuccessPayload()}
	if _, err := implementWithAdapter(root, l, wf, "SPEC-001", ad); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertNoRunLock(t, root)
}

func TestRunLock_AbsentAfterFailedRun(t *testing.T) {
	root := setupApprovedSpec(t)
	l, wf := mustPrepareImplementation(t, root, "SPEC-001")
	ad := &tamperingAdapter{FakeAdapter: &adapter.FakeAdapter{RunErr: fmt.Errorf("boom")}, SkipResult: true}
	if _, err := implementWithAdapter(root, l, wf, "SPEC-001", ad); err == nil {
		t.Fatalf("expected an error")
	}
	assertNoRunLock(t, root)
}

func TestRunLock_AbsentAfterRejectedRun(t *testing.T) {
	root := setupApprovedSpec(t)
	if _, err := specCreateDefined(root, "Email Verification"); err != nil {
		t.Fatal(err)
	}
	l, wf := mustPrepareImplementation(t, root, "SPEC-001")
	forgedPath := filepath.Join(l.ApprovalsDir(), "SPEC-002", "forged03.grant.json")
	ad := &tamperingAdapter{
		FakeAdapter: &adapter.FakeAdapter{},
		Payload:     implementationSuccessPayload(),
		Tamper:      func() error { writeForgedGrant(t, forgedPath); return nil },
	}
	if _, err := implementWithAdapter(root, l, wf, "SPEC-001", ad); err == nil {
		t.Fatalf("expected rejection")
	}
	assertNoRunLock(t, root)
}

// assertNoRunLock checks that no run holds the project lock any more and that the holder
// description was cleared. The lock file itself is kept on purpose (runlock.go).
func assertNoRunLock(t *testing.T, root string) {
	t.Helper()
	if active := probeRunLock(root); active != nil {
		t.Fatalf("expected the run lock to be released, still held: %v", active)
	}
	path, err := runLockPath(root)
	if err != nil {
		t.Fatal(err)
	}
	if data, readErr := os.ReadFile(path); readErr == nil && len(data) != 0 {
		t.Fatalf("expected the holder description cleared, got %q", data)
	}
}

// --- 8. Restoration failure ---

func TestRunGuard_RestorationFailure_ReportListsWhatCouldNotBeRestored(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("running as root bypasses permission checks")
	}
	root := setupApprovedSpec(t)
	if _, err := specCreateDefined(root, "Email Verification"); err != nil {
		t.Fatal(err)
	}
	l, wf := mustPrepareImplementation(t, root, "SPEC-001")

	dir := filepath.Join(l.ApprovalsDir(), "SPEC-002")
	forgedPath := filepath.Join(dir, "forged04.grant.json")
	t.Cleanup(func() { os.Chmod(dir, 0o755) }) // let t.TempDir() clean up afterward

	ad := &tamperingAdapter{
		FakeAdapter: &adapter.FakeAdapter{},
		Payload:     implementationSuccessPayload(),
		Tamper: func() error {
			if err := os.MkdirAll(dir, 0o755); err != nil {
				return err
			}
			if err := os.WriteFile(forgedPath, []byte("{}"), 0o644); err != nil {
				return err
			}
			return os.Chmod(dir, 0o555) // now un-writable: restoring (removing) forgedPath will fail
		},
	}

	rep, err := implementWithAdapter(root, l, wf, "SPEC-001", ad)
	if err == nil {
		t.Fatalf("expected the run to be rejected")
	}
	if rep == nil || rep.Outcome != present.Failed {
		t.Fatalf("expected a Failed report, got: %+v", rep)
	}
	rendered := present.Render(rep, false)
	if !strings.Contains(rendered, "restoring it failed") {
		t.Fatalf("expected the summary to say restoring failed, got: %s", rendered)
	}
	if !strings.Contains(rendered, "forged04.grant.json") {
		t.Fatalf("expected the unrestored file listed, got: %s", rendered)
	}
	if !strings.Contains(rendered, "gnomon status") {
		t.Fatalf("expected the report to warn about trusting gnomon status, got: %s", rendered)
	}
}

// findEvidenceFile returns the one file under l.ApprovalsDir()/specID matching suffix.
func findEvidenceFile(t *testing.T, l project.Layout, specID, suffix string) string {
	t.Helper()
	dir := filepath.Join(l.ApprovalsDir(), specID)
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), suffix) {
			return filepath.Join(dir, e.Name())
		}
	}
	t.Fatalf("expected a %s file under %s", suffix, dir)
	return ""
}
