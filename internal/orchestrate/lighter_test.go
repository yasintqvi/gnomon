package orchestrate

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gnomon/internal/adapter"
	"gnomon/internal/approval"
	"gnomon/internal/present"
	"gnomon/internal/project"
	"gnomon/internal/specs"
)

// legacyProject copies testdata/legacy — a .gnomon/ directory exactly as an older `gnomon init`
// left it (every workflow, the old SPEC-000 template, unfilled context/contract/evaluation
// templates) — into a fresh Git-configured project root.
func legacyProject(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	configureGitIdentity(t, root)
	src := filepath.Join("testdata", "legacy")
	err := filepath.WalkDir(src, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, p)
		dest := filepath.Join(root, rel)
		if d.IsDir() {
			return os.MkdirAll(dest, 0o755)
		}
		data, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		return os.WriteFile(dest, data, 0o644)
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, ".gnomon", "approvals"), 0o755); err != nil {
		t.Fatal(err)
	}
	return root
}

func TestLegacyProject_StaleWorkflowCopiesIgnored_BundledUsed(t *testing.T) {
	root := legacyProject(t)

	_, _, path, err := prepareWorkflow(root, "implementation.md")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(path, ".gnomon-runtime") {
		t.Fatalf("expected the bundled workflow to be used instead of the stale project copy, got %s", path)
	}
	data, _ := os.ReadFile(path)
	old, _ := os.ReadFile(filepath.Join(root, ".gnomon", "workflows", "implementation.md"))
	if string(data) == string(old) {
		t.Fatalf("expected the current bundled workflow, got the old project copy")
	}

	rep, err := Validate(root)
	if err != nil {
		t.Fatalf("a legacy project must still validate: %v (%+v)", err, rep)
	}
	body := sectionBody(rep, "Workflows")
	if !strings.Contains(body, "10 unmodified workflow copies") {
		t.Fatalf("expected validate to report the ignored copies, got:\n%s", body)
	}
}

func TestLegacyProject_UnfilledTemplatesAreNotKnowledge_EditedOnesAre(t *testing.T) {
	root := legacyProject(t)
	l, _ := project.Locate(root)

	if k, err := l.KnowledgeFiles(); err != nil || len(k) != 0 {
		t.Fatalf("expected no knowledge from unfilled templates, got %v (err=%v)", k, err)
	}
	domain := filepath.Join(l.GnomonRoot(), "context", "DOMAIN.md")
	data, _ := os.ReadFile(domain)
	if err := os.WriteFile(domain, append(data, []byte("\nA Center has exactly one active representative.\n")...), 0o644); err != nil {
		t.Fatal(err)
	}
	k, err := l.KnowledgeFiles()
	if err != nil || len(k) != 1 || k[0] != ".gnomon/context/DOMAIN.md" {
		t.Fatalf("expected only the edited file, got %v (err=%v)", k, err)
	}

	// The Agent is told exactly that list.
	ad := &adapter.FakeAdapter{}
	_, _ = implementLegacySpec(t, root, ad) // the fake writes no result; only the prepared context matters
	if len(ad.Ctx.Knowledge) != 1 || ad.Ctx.Knowledge[0] != ".gnomon/context/DOMAIN.md" {
		t.Fatalf("expected the prompt context to list only the edited knowledge file, got %v", ad.Ctx.Knowledge)
	}
	if !strings.HasPrefix(ad.Ctx.SpecPath, ".gnomon/specifications/SPEC-001-") {
		t.Fatalf("expected the governing Specification's path, got %q", ad.Ctx.SpecPath)
	}
}

// implementLegacySpec creates, defines, and approves SPEC-001 from the legacy project's own old
// template, then runs Implementation with ad.
func implementLegacySpec(t *testing.T, root string, ad adapter.Adapter) (*present.Report, error) {
	t.Helper()
	if _, err := specCreateDefined(root, "Register a Center"); err != nil {
		t.Fatal(err)
	}
	if _, err := Approve(root, "SPEC-001", nil, nil); err != nil {
		t.Fatalf("approve: %v", err)
	}
	l, wf := mustPrepareImplementation(t, root, "SPEC-001")
	return implementWithAdapter(root, l, wf, "SPEC-001", ad)
}

func TestLegacyProject_NewSpecUsesShortTemplate_OldTemplateStillRecognized(t *testing.T) {
	root := legacyProject(t)
	l, _ := project.Locate(root)

	// A Specification created long ago from the old template, never filled in.
	oldTemplate, _ := os.ReadFile(filepath.Join(l.SpecificationsDir(), "SPEC-000-use-case-name.md"))
	if _, err := specs.CreateDraft(l.SpecificationsDir(), "SPEC-001", "Old Untouched", "old-untouched", oldTemplate); err != nil {
		t.Fatal(err)
	}
	if _, err := Approve(root, "SPEC-001", nil, nil); err == nil {
		t.Fatalf("expected an untouched old-template Specification to be refused")
	}

	// The project's SPEC-000 is an unmodified shipped copy, so new Specifications are short.
	if _, err := SpecCreate(root, "New Feature"); err != nil {
		t.Fatal(err)
	}
	created, err := specs.ReadContent(l.SpecificationsDir(), "SPEC-002")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(created), "## Decisions") || strings.Contains(string(created), "## Alternative Flows") {
		t.Fatalf("expected the short template, got:\n%s", created)
	}
}

// TestLegacyProject_ApprovedEmptySpec_FlaggedNotRevoked mirrors a real project where an untouched
// template was approved before the empty-template check existed: it stays Approved (evidence is
// never rewritten), and status says what it is.
func TestLegacyProject_ApprovedEmptySpec_FlaggedNotRevoked(t *testing.T) {
	root := legacyProject(t)
	l, _ := project.Locate(root)
	oldTemplate, _ := os.ReadFile(filepath.Join(l.SpecificationsDir(), "SPEC-000-use-case-name.md"))
	path, err := specs.CreateDraft(l.SpecificationsDir(), "SPEC-004", "Submit an Activity Report", "submit-an-activity-report", oldTemplate)
	if err != nil {
		t.Fatal(err)
	}
	content, _ := os.ReadFile(path)
	if err := approval.WriteGrant(l.ApprovalsDir(), "SPEC-004", approval.Fingerprint(content), "someone", content); err != nil {
		t.Fatal(err)
	}

	rep, err := Status(root)
	if err != nil {
		t.Fatal(err)
	}
	if want := "SPEC-004: Approved — approved, but contains only the unfilled template"; !strings.Contains(sectionBody(rep, "Specifications"), want) {
		t.Fatalf("expected %q, got:\n%s", want, sectionBody(rep, "Specifications"))
	}
}

func TestStatus_LifecycleNotes_NeverApproved_Changed_Revoked(t *testing.T) {
	root := t.TempDir()
	configureGitIdentity(t, root)
	if _, err := Init(root); err != nil {
		t.Fatal(err)
	}
	if _, err := specCreateDefined(root, "Mark a task complete"); err != nil {
		t.Fatal(err)
	}
	l, _ := project.Locate(root)
	_, path, _ := specs.Exists(l.SpecificationsDir(), "SPEC-001")

	note := func() string {
		t.Helper()
		rep, err := Status(root)
		if err != nil {
			t.Fatal(err)
		}
		return sectionBody(rep, "Specifications")
	}

	if got := note(); !strings.Contains(got, "SPEC-001: Draft — never approved") {
		t.Fatalf("got:\n%s", got)
	}
	if _, err := Approve(root, "SPEC-001", nil, nil); err != nil {
		t.Fatal(err)
	}
	if got := note(); strings.Contains(got, "—") {
		t.Fatalf("an Approved, defined Specification needs no note, got:\n%s", got)
	}

	approved, _ := os.ReadFile(path)
	edited := strings.Replace(string(approved), "works as described.", "works as described, including undo.", 1)
	if err := os.WriteFile(path, []byte(edited), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := note(); !strings.Contains(got, "SPEC-001: Draft — changed since approval on ") || !strings.Contains(got, "(1 line(s) changed)") {
		t.Fatalf("expected the change since approval to be reported, got:\n%s", got)
	}

	// Restoring the approved text restores Approved under the same grant (unchanged behavior).
	if err := os.WriteFile(path, approved, 0o644); err != nil {
		t.Fatal(err)
	}
	if got := note(); !strings.Contains(got, "SPEC-001: Approved") {
		t.Fatalf("expected Approved after restoring the approved text, got:\n%s", got)
	}

	if _, err := Revoke(root, "SPEC-001", nil); err != nil {
		t.Fatal(err)
	}
	if got := note(); !strings.Contains(got, "SPEC-001: Draft — approval revoked") {
		t.Fatalf("got:\n%s", got)
	}
}

func TestCustomizedWorkflow_UsedOnlyOnceChanged(t *testing.T) {
	root := t.TempDir()
	configureGitIdentity(t, root)
	if _, err := Init(root); err != nil {
		t.Fatal(err)
	}

	if _, err := CustomizeWorkflow(root, "review"); err != nil {
		t.Fatal(err)
	}
	l, _ := project.Locate(root)
	projectPath := filepath.Join(l.WorkflowsDir(), "review.md")
	if _, _, path, err := prepareWorkflow(root, "review.md"); err != nil || path == projectPath {
		t.Fatalf("an unchanged copy must not override the bundled workflow (path=%s, err=%v)", path, err)
	}

	data, _ := os.ReadFile(projectPath)
	if err := os.WriteFile(projectPath, append(data, []byte("\n## Team rule\n\nAlso check the audit log.\n")...), 0o644); err != nil {
		t.Fatal(err)
	}
	_, wf, path, err := prepareWorkflow(root, "review.md")
	if err != nil {
		t.Fatal(err)
	}
	if path != projectPath || wf.Identity != "review" {
		t.Fatalf("expected the customized project workflow to be used, got %s (%s)", path, wf.Identity)
	}
	rep, err := Workflows(root)
	if err != nil {
		t.Fatal(err)
	}
	if body := sectionBody(rep, "Workflows"); !strings.Contains(body, "customized: .gnomon/workflows/review.md") {
		t.Fatalf("expected the listing to show the customization, got:\n%s", body)
	}
	if _, err := CustomizeWorkflow(root, "review.md"); err == nil {
		t.Fatalf("customize must never overwrite an existing project file")
	}
}

func TestProjectDefinedWorkflow_RunnableByIdentity(t *testing.T) {
	root := t.TempDir()
	configureGitIdentity(t, root)
	if _, err := Init(root); err != nil {
		t.Fatal(err)
	}
	l, _ := project.Locate(root)
	if err := os.MkdirAll(l.WorkflowsDir(), 0o755); err != nil {
		t.Fatal(err)
	}
	custom := "---\nidentity: release-notes\nspecification_reference: none\nrequires_approved_specification: false\nresult:\n  terminal_path: outcome\n  classification:\n    DONE: success\n  schema:\n    type: object\n    required: [outcome]\n    properties:\n      outcome:\n        type: string\n        enum: [DONE]\n---\n# Release notes\n"
	if err := os.WriteFile(filepath.Join(l.WorkflowsDir(), "release-notes.md"), []byte(custom), 0o644); err != nil {
		t.Fatal(err)
	}
	wf, path, err := resolveWorkflowByIdentity(l, "release-notes")
	if err != nil || wf.Identity != "release-notes" || path != filepath.Join(l.WorkflowsDir(), "release-notes.md") {
		t.Fatalf("expected the project workflow to resolve (path=%s, err=%v)", path, err)
	}
	if rep, err := Validate(root); err != nil {
		t.Fatalf("expected validate to accept the project workflow: %v (%+v)", err, rep)
	}
}

func runVerification(t *testing.T, payload map[string]interface{}) (*present.Report, []EvaluationFinding, error) {
	t.Helper()
	root := t.TempDir()
	configureGitIdentity(t, root)
	if _, err := Init(root); err != nil {
		t.Fatal(err)
	}
	l, wf, workflowPath, err := prepareWorkflow(root, "verification.md")
	if err != nil {
		t.Fatal(err)
	}
	rep, outcome, err := execute(runRequest{root: root, l: l, wf: wf, workflowPath: workflowPath, kind: targetGeneric, target: "SPEC-001", adapter: &payloadAdapter{FakeAdapter: &adapter.FakeAdapter{}, Payload: payload}})
	var findings []EvaluationFinding
	if outcome != nil {
		findings = ActionableFindings(wf.Identity, outcome.Payload)
	}
	return rep, findings, err
}

func TestVerification_PassWithoutEvidence_ReportedUnverified(t *testing.T) {
	rep, findings, err := runVerification(t, map[string]interface{}{
		"evidence": []interface{}{
			map[string]interface{}{"obligation": "An open task can be marked complete", "source": "SPEC-001, criterion 1", "result": "PASS", "evidence": "TaskTest::testComplete passed"},
			map[string]interface{}{"obligation": "A completed task can be reopened", "source": "SPEC-001, criterion 2", "result": "PASS"},
		},
		"summary": map[string]interface{}{"aggregate": "PASS"},
	})
	if err == nil || rep.Outcome != present.Blocked {
		t.Fatalf("a PASS claim with a criterion lacking evidence must not succeed, got %v (err=%v)", rep.Outcome, err)
	}
	if rep.Summary != "Verification: 1 passed, 0 failed, 1 unverified" {
		t.Fatalf("unexpected summary %q", rep.Summary)
	}
	criteria := sectionBody(rep, "Criteria")
	for _, want := range []string{
		"1. passed     An open task can be marked complete",
		"Evidence: TaskTest::testComplete passed",
		"2. unverified A completed task can be reopened",
		"Evidence: " + noEvidenceNote,
		"Source: SPEC-001, criterion 2",
	} {
		if !strings.Contains(criteria, want) {
			t.Fatalf("expected %q in:\n%s", want, criteria)
		}
	}
	if sectionBody(rep, "Note") != evidenceDisclaimer {
		t.Fatalf("expected the evidence disclaimer on every verification report")
	}
	if len(findings) != 1 || findings[0].Classification != "UNVERIFIABLE" {
		t.Fatalf("expected the downgraded criterion to be offered as a finding, got %+v", findings)
	}
}

func TestVerification_NoCriteria_NeverSuccess(t *testing.T) {
	rep, _, err := runVerification(t, map[string]interface{}{"summary": map[string]interface{}{"aggregate": "PASS"}})
	if err == nil || rep.Outcome == present.Success {
		t.Fatalf("a verification with no criteria must never report success, got %v", rep.Outcome)
	}
	if rep.Summary != "Verification: nothing was verified" {
		t.Fatalf("unexpected summary %q", rep.Summary)
	}
}

func TestVerification_AgentFailNeverRaised_AllEvidencedPassSucceeds(t *testing.T) {
	item := map[string]interface{}{"obligation": "c1", "result": "PASS", "evidence": "ran it"}
	rep, _, err := runVerification(t, map[string]interface{}{"evidence": []interface{}{item}, "summary": map[string]interface{}{"aggregate": "FAIL"}})
	if err == nil || rep.Outcome != present.Blocked {
		t.Fatalf("Gnomon may lower a result, never raise it; got %v", rep.Outcome)
	}

	item2 := map[string]interface{}{"obligation": "c1", "result": "PASS", "evidence": "ran it"}
	rep, _, err = runVerification(t, map[string]interface{}{"evidence": []interface{}{item2}, "summary": map[string]interface{}{"aggregate": "PASS"}})
	if err != nil || rep.Outcome != present.Success || rep.Summary != "Verification: 1 passed, 0 failed, 0 unverified" {
		t.Fatalf("expected success for fully evidenced criteria, got %v %q (err=%v)", rep.Outcome, rep.Summary, err)
	}
}

func TestPromptContext_NewProject_NoKnowledgeNoTemplates(t *testing.T) {
	root := setupApprovedSpec(t)
	l, wf := mustPrepareImplementation(t, root, "SPEC-001")
	ad := &adapter.FakeAdapter{}
	_, _ = implementWithAdapter(root, l, wf, "SPEC-001", ad)
	if len(ad.Ctx.Knowledge) != 0 {
		t.Fatalf("a fresh project has no knowledge files to hand over, got %v", ad.Ctx.Knowledge)
	}
	if !strings.HasPrefix(ad.Ctx.WorkflowPath, filepath.Join(root, ".gnomon-runtime", "workflows")) {
		t.Fatalf("expected the bundled workflow path, got %s", ad.Ctx.WorkflowPath)
	}
}

func sectionBody(rep *present.Report, label string) string {
	if rep == nil {
		return ""
	}
	for _, s := range rep.Sections {
		if s.Label == label {
			return s.Body
		}
	}
	return ""
}
