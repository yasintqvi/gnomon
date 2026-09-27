package orchestrate

import (
	"bytes"
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
	"gnomon/internal/result"
)

// appendWhitespaceOnly inserts trailing spaces right before the first newline — normalize()'s own
// TrimRight undoes this exactly, so the fingerprint is unchanged even though the bytes differ.
func appendWhitespaceOnly(t *testing.T, content []byte) []byte {
	t.Helper()
	idx := bytes.IndexByte(content, '\n')
	if idx < 0 {
		idx = len(content)
	}
	out := append([]byte{}, content[:idx]...)
	out = append(out, []byte("   ")...)
	out = append(out, content[idx:]...)
	return out
}

func appendSemanticEdit(content []byte, note string) []byte {
	return append(append([]byte{}, content...), []byte("\n"+note+"\n")...)
}

func anyFileUnder(t *testing.T, root string) bool {
	t.Helper()
	found := false
	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err == nil && info != nil && !info.IsDir() {
			found = true
		}
		return nil
	})
	return found
}

// --- Rule A: the governing Specification must not change ---

func TestSpecGuard_ImplementationEditsGoverningSpec_RejectedRestoredPreserved(t *testing.T) {
	root := setupApprovedSpec(t) // SPEC-001 Approved
	l, wf := mustPrepareImplementation(t, root, "SPEC-001")

	specPath := filepath.Join(l.SpecificationsDir(), "SPEC-001-password-reset.md")
	original, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatal(err)
	}

	ad := &tamperingAdapter{
		FakeAdapter: &adapter.FakeAdapter{},
		Payload:     implementationSuccessPayload(),
		Tamper: func() error {
			return os.WriteFile(specPath, appendSemanticEdit(original, "Edited by the Agent."), 0o644)
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
	if !strings.Contains(rendered, "changed the approved Specification SPEC-001") {
		t.Fatalf("unexpected summary: %s", rendered)
	}

	after, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != string(original) {
		t.Fatalf("expected the Specification's original bytes restored, got %q", after)
	}

	transientDir, err := result.TransientDir(root)
	if err != nil {
		t.Fatal(err)
	}
	if !anyFileUnder(t, filepath.Join(transientDir, "rejected")) {
		t.Fatalf("expected the Agent's version preserved under rejected/<run-id>/")
	}

	l2, err := project.Locate(root)
	if err != nil {
		t.Fatal(err)
	}
	state, err := facts.Lifecycle(l2, "SPEC-001")
	if err != nil {
		t.Fatal(err)
	}
	if state != approval.Approved {
		t.Fatalf("expected SPEC-001 to still derive Approved after restoration, got %v", state)
	}
}

func TestSpecGuard_WhitespaceOnlyEdit_AcceptedFileUntouchedByGuard(t *testing.T) {
	root := setupApprovedSpec(t)
	l, wf := mustPrepareImplementation(t, root, "SPEC-001")
	specPath := filepath.Join(l.SpecificationsDir(), "SPEC-001-password-reset.md")
	original, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatal(err)
	}
	edited := appendWhitespaceOnly(t, original)

	ad := &tamperingAdapter{
		FakeAdapter: &adapter.FakeAdapter{},
		Payload:     implementationSuccessPayload(),
		Tamper:      func() error { return os.WriteFile(specPath, edited, 0o644) },
	}

	rep, err := implementWithAdapter(root, l, wf, "SPEC-001", ad)
	if err != nil {
		t.Fatalf("expected acceptance for a whitespace-only edit, got err: %v (report: %+v)", err, rep)
	}
	if rep.Outcome != present.Success {
		t.Fatalf("expected Success, got %v", rep.Outcome)
	}

	after, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(after, edited) {
		t.Fatalf("expected the whitespace edit left exactly as the Agent wrote it (guard must not touch it)")
	}
}

func TestSpecGuard_AgentRenamesSpec_OriginalRestoredRenamedRemoved(t *testing.T) {
	root := setupApprovedSpec(t)
	l, wf := mustPrepareImplementation(t, root, "SPEC-001")
	specPath := filepath.Join(l.SpecificationsDir(), "SPEC-001-password-reset.md")
	renamedPath := filepath.Join(l.SpecificationsDir(), "SPEC-001-renamed-by-agent.md")
	original, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatal(err)
	}

	ad := &tamperingAdapter{
		FakeAdapter: &adapter.FakeAdapter{},
		Payload:     implementationSuccessPayload(),
		Tamper:      func() error { return os.Rename(specPath, renamedPath) },
	}

	rep, err := implementWithAdapter(root, l, wf, "SPEC-001", ad)
	if err == nil {
		t.Fatalf("expected the run to be rejected")
	}
	if rep == nil || rep.Outcome != present.Failed {
		t.Fatalf("expected a Failed report, got: %+v", rep)
	}

	after, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatalf("expected the original path restored: %v", err)
	}
	if string(after) != string(original) {
		t.Fatalf("expected original bytes restored")
	}
	if _, statErr := os.Stat(renamedPath); !os.IsNotExist(statErr) {
		t.Fatalf("expected the renamed file removed, stat err: %v", statErr)
	}

	matches, err := specFilesMatchingIdentity(l.SpecificationsDir(), "SPEC-001")
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) != 1 || matches[0] != specPath {
		t.Fatalf("expected the identity to resolve to exactly one file (%s), got %v", specPath, matches)
	}
}

func TestSpecGuard_AgentDeletesSpec_RestoredRejected(t *testing.T) {
	root := setupApprovedSpec(t)
	l, wf := mustPrepareImplementation(t, root, "SPEC-001")
	specPath := filepath.Join(l.SpecificationsDir(), "SPEC-001-password-reset.md")
	original, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatal(err)
	}

	ad := &tamperingAdapter{
		FakeAdapter: &adapter.FakeAdapter{},
		Payload:     implementationSuccessPayload(),
		Tamper:      func() error { return os.Remove(specPath) },
	}

	rep, err := implementWithAdapter(root, l, wf, "SPEC-001", ad)
	if err == nil {
		t.Fatalf("expected the run to be rejected")
	}
	if rep == nil || rep.Outcome != present.Failed {
		t.Fatalf("expected a Failed report, got: %+v", rep)
	}

	after, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatalf("expected the Specification restored: %v", err)
	}
	if string(after) != string(original) {
		t.Fatalf("expected original bytes restored")
	}
}

func TestSpecGuard_TestingWithGoverningSpec_EditRejected(t *testing.T) {
	root := setupApprovedSpec(t)
	l, wf, workflowPath, err := prepareWorkflow(root, "testing.md")
	if err != nil {
		t.Fatal(err)
	}
	specPath := filepath.Join(l.SpecificationsDir(), "SPEC-001-password-reset.md")
	original, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatal(err)
	}

	ad := &tamperingAdapter{
		FakeAdapter: &adapter.FakeAdapter{},
		Payload:     map[string]interface{}{"outcome": "TESTING_COMPLETE", "evidence": "go test ./... passed"},
		Tamper: func() error {
			return os.WriteFile(specPath, appendSemanticEdit(original, "Edited."), 0o644)
		},
	}

	rep, _, err := execute(runRequest{root: root, l: l, wf: wf, workflowPath: workflowPath, kind: targetSpec, target: "SPEC-001", adapter: ad})
	if err == nil {
		t.Fatalf("expected the run to be rejected")
	}
	if rep == nil || rep.Outcome != present.Failed {
		t.Fatalf("expected a Failed report, got: %+v", rep)
	}
}

func TestSpecGuard_TestingWithoutSpec_RuleADoesNotApply(t *testing.T) {
	root := setupApprovedSpec(t)
	l, wf, workflowPath, err := prepareWorkflow(root, "testing.md")
	if err != nil {
		t.Fatal(err)
	}
	specPath := filepath.Join(l.SpecificationsDir(), "SPEC-001-password-reset.md")

	ad := &tamperingAdapter{
		FakeAdapter: &adapter.FakeAdapter{},
		Payload:     map[string]interface{}{"outcome": "TESTING_COMPLETE", "evidence": "go test ./... passed"},
		Tamper: func() error {
			content, err := os.ReadFile(specPath)
			if err != nil {
				return err
			}
			return os.WriteFile(specPath, appendSemanticEdit(content, "Edited without a governing target."), 0o644)
		},
	}

	rep, _, err := execute(runRequest{root: root, l: l, wf: wf, workflowPath: workflowPath, kind: targetSpec, target: "", adapter: ad})
	if err != nil {
		t.Fatalf("expected acceptance (Rule A does not apply with no supplied Spec), got err: %v (report: %+v)", err, rep)
	}
	if rep.Outcome != present.Success {
		t.Fatalf("expected Success, got %v", rep.Outcome)
	}
}

func TestSpecGuard_SpecificationDefinitionEditsDraftTarget_AcceptedAsToday(t *testing.T) {
	root := t.TempDir()
	configureGitIdentity(t, root)
	if _, err := Init(root); err != nil {
		t.Fatal(err)
	}
	if _, err := SpecCreate(root, "Password Reset"); err != nil {
		t.Fatal(err)
	}

	l, wf, workflowPath, err := prepareWorkflow(root, "specification-definition.md")
	if err != nil {
		t.Fatal(err)
	}
	specPath := filepath.Join(l.SpecificationsDir(), "SPEC-001-password-reset.md")

	ad := &tamperingAdapter{
		FakeAdapter: &adapter.FakeAdapter{},
		Payload:     map[string]interface{}{"outcome": "READY_FOR_APPROVAL", "target": "SPEC-001", "consistency_check": "CLEAN"},
		Tamper: func() error {
			content, err := os.ReadFile(specPath)
			if err != nil {
				return err
			}
			return os.WriteFile(specPath, appendSemanticEdit(content, "- A newly defined requirement."), 0o644)
		},
	}

	rep, _, err := execute(runRequest{root: root, l: l, wf: wf, workflowPath: workflowPath, kind: targetSpec, target: "SPEC-001", adapter: ad})
	if err != nil {
		t.Fatalf("expected acceptance (requires_approved_specification is false for this workflow), got err: %v (report: %+v)", err, rep)
	}
	if rep.Outcome != present.Success {
		t.Fatalf("expected Success, got %v", rep.Outcome)
	}
}

func TestSpecGuard_AgentFailsWithoutResult_StillRestoredAndRejected(t *testing.T) {
	root := setupApprovedSpec(t)
	l, wf := mustPrepareImplementation(t, root, "SPEC-001")
	specPath := filepath.Join(l.SpecificationsDir(), "SPEC-001-password-reset.md")
	original, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatal(err)
	}

	ad := &tamperingAdapter{
		FakeAdapter: &adapter.FakeAdapter{RunErr: fmt.Errorf("agent crashed")},
		SkipResult:  true,
		Tamper: func() error {
			return os.WriteFile(specPath, appendSemanticEdit(original, "Edited."), 0o644)
		},
	}

	rep, err := implementWithAdapter(root, l, wf, "SPEC-001", ad)
	if err == nil {
		t.Fatalf("expected the run to be rejected")
	}
	if rep == nil || rep.Outcome != present.Failed {
		t.Fatalf("expected a Failed report, got: %+v", rep)
	}

	after, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != string(original) {
		t.Fatalf("expected original bytes restored")
	}
}

func TestSpecGuard_BothApprovalsAndSpecTampered_OneReportBothRestored(t *testing.T) {
	root := setupApprovedSpec(t)
	if _, err := SpecCreate(root, "Email Verification"); err != nil {
		t.Fatal(err) // SPEC-002, Draft
	}
	l, wf := mustPrepareImplementation(t, root, "SPEC-001")

	specPath := filepath.Join(l.SpecificationsDir(), "SPEC-001-password-reset.md")
	originalSpec, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatal(err)
	}
	forgedGrantPath := filepath.Join(l.ApprovalsDir(), "SPEC-002", "forged.grant.json")

	ad := &tamperingAdapter{
		FakeAdapter: &adapter.FakeAdapter{},
		Payload:     implementationSuccessPayload(),
		Tamper: func() error {
			if err := os.WriteFile(specPath, appendSemanticEdit(originalSpec, "Edited."), 0o644); err != nil {
				return err
			}
			writeForgedGrant(t, forgedGrantPath)
			return nil
		},
	}

	rep, err := implementWithAdapter(root, l, wf, "SPEC-001", ad)
	if err == nil {
		t.Fatalf("expected the run to be rejected")
	}
	if rep == nil || rep.Outcome != present.Failed {
		t.Fatalf("expected a single Failed report, got: %+v", rep)
	}
	rendered := present.Render(rep, false)
	if !strings.Contains(rendered, "modified approval evidence") {
		t.Fatalf("expected the approvals violation mentioned, got: %s", rendered)
	}
	if !strings.Contains(rendered, "changed the approved Specification SPEC-001") {
		t.Fatalf("expected the Specification violation mentioned in the same report, got: %s", rendered)
	}

	afterSpec, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(afterSpec) != string(originalSpec) {
		t.Fatalf("expected the Specification restored")
	}
	if _, statErr := os.Stat(forgedGrantPath); !os.IsNotExist(statErr) {
		t.Fatalf("expected the forged grant removed")
	}
}

// --- Rule B: other Approved Specifications losing approval is reported, never restored ---

func TestRuleB_KnowledgeResolutionEditsApprovedSpec_AcceptedWithWarning(t *testing.T) {
	root := t.TempDir()
	configureGitIdentity(t, root)
	if _, err := Init(root); err != nil {
		t.Fatal(err)
	}
	if _, err := SpecCreate(root, "Spec One"); err != nil {
		t.Fatal(err) // SPEC-001
	}
	if _, err := SpecCreate(root, "Spec Two"); err != nil {
		t.Fatal(err) // SPEC-002
	}
	if _, err := Approve(root, "SPEC-002", nil, nil); err != nil {
		t.Fatal(err)
	}

	l, wf, workflowPath, err := prepareWorkflow(root, "knowledge-resolution.md")
	if err != nil {
		t.Fatal(err)
	}
	specPath := filepath.Join(l.SpecificationsDir(), "SPEC-002-spec-two.md")

	ad := &tamperingAdapter{
		FakeAdapter: &adapter.FakeAdapter{},
		Payload:     map[string]interface{}{"outcome": "RESOLVED", "updated_knowledge": "SPEC-002 updated per Human decision"},
		Tamper: func() error {
			content, err := os.ReadFile(specPath)
			if err != nil {
				return err
			}
			return os.WriteFile(specPath, appendSemanticEdit(content, "Updated per Human decision."), 0o644)
		},
	}

	rep, _, err := execute(runRequest{root: root, l: l, wf: wf, workflowPath: workflowPath, kind: targetNone, target: "", adapter: ad})
	if err != nil {
		t.Fatalf("expected acceptance, got err: %v (report: %+v)", err, rep)
	}
	if rep.Outcome != present.Success {
		t.Fatalf("expected Success, got %v", rep.Outcome)
	}
	rendered := present.Render(rep, false)
	if !strings.Contains(rendered, "Other Approvals Lost") || !strings.Contains(rendered, "SPEC-002") {
		t.Fatalf("expected a warning that SPEC-002 lost approval, got: %s", rendered)
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
		t.Fatalf("expected SPEC-002 to actually derive Draft (Rule B never restores), got %v", state)
	}
}

func TestRuleB_ImplementationEditsUnrelatedApprovedSpec_AcceptedWithWarning(t *testing.T) {
	root := setupApprovedSpec(t) // SPEC-001 Approved
	if _, err := SpecCreate(root, "Email Verification"); err != nil {
		t.Fatal(err) // SPEC-002
	}
	if _, err := Approve(root, "SPEC-002", nil, nil); err != nil {
		t.Fatal(err)
	}

	l, wf := mustPrepareImplementation(t, root, "SPEC-001")
	spec2Path := filepath.Join(l.SpecificationsDir(), "SPEC-002-email-verification.md")

	ad := &tamperingAdapter{
		FakeAdapter: &adapter.FakeAdapter{},
		Payload:     implementationSuccessPayload(),
		Tamper: func() error {
			content, err := os.ReadFile(spec2Path)
			if err != nil {
				return err
			}
			return os.WriteFile(spec2Path, appendSemanticEdit(content, "Edited incidentally."), 0o644)
		},
	}

	rep, err := implementWithAdapter(root, l, wf, "SPEC-001", ad)
	if err != nil {
		t.Fatalf("expected acceptance (Rule A only governs SPEC-001), got err: %v (report: %+v)", err, rep)
	}
	if rep.Outcome != present.Success {
		t.Fatalf("expected Success, got %v", rep.Outcome)
	}
	rendered := present.Render(rep, false)
	if !strings.Contains(rendered, "SPEC-002") {
		t.Fatalf("expected a warning listing SPEC-002, got: %s", rendered)
	}
}

func TestRuleB_NothingChanges_NoWarning(t *testing.T) {
	root := setupApprovedSpec(t)
	l, wf := mustPrepareImplementation(t, root, "SPEC-001")
	ad := &tamperingAdapter{FakeAdapter: &adapter.FakeAdapter{}, Payload: implementationSuccessPayload()}

	rep, err := implementWithAdapter(root, l, wf, "SPEC-001", ad)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	rendered := present.Render(rep, false)
	if strings.Contains(rendered, "Other Approvals Lost") {
		t.Fatalf("did not expect a Rule B warning when nothing changed, got: %s", rendered)
	}
}
