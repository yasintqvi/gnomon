package orchestrate

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"gnomon/internal/adapter"
	"gnomon/internal/present"
	"gnomon/internal/project"
)

func gitRun(t *testing.T, root string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = root
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

func writeProjectFile(t *testing.T, root, rel, content string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// committedProject is an initialized project with one Approved Specification, some code and a
// test file, all committed.
func committedProject(t *testing.T) string {
	t.Helper()
	root := setupApprovedSpec(t)
	writeProjectFile(t, root, "app/tasks.py", "def complete(task):\n    task.done = True\n")
	writeProjectFile(t, root, "tests/test_tasks.py", "def test_complete_marks_task_done():\n    assert True\n")
	gitRun(t, root, "add", "-A")
	gitRun(t, root, "commit", "-q", "-m", "baseline")
	return root
}

var passingCriterion = map[string]interface{}{
	"obligation": "An open task can be marked complete", "source": "SPEC-001, criterion 1",
	"result": "PASS", "evidence": "tests/test_tasks.py::test_complete_marks_task_done asserts done is true; it passed",
}

// verifyWith runs Verification through execute with a scripted result and returns the report and
// the adapter, so a test can see whether an Agent was started and what scope it was given.
func verifyWith(t *testing.T, root, target string, opts VerificationOptions, payload map[string]interface{}) (*present.Report, *payloadAdapter, error) {
	t.Helper()
	l, wf, workflowPath, err := prepareWorkflow(root, "verification.md")
	if err != nil {
		t.Fatal(err)
	}
	ad := &payloadAdapter{FakeAdapter: &adapter.FakeAdapter{}, Payload: payload}
	rep, _, err := execute(runRequest{root: root, l: l, wf: wf, workflowPath: workflowPath, kind: targetGeneric, target: target, adapter: ad, verification: opts})
	return rep, ad, err
}

func passPayload() map[string]interface{} {
	return verificationPayload("PASS", criterion(passingCriterion, nil))
}

func TestVerificationScope_FirstRunChecksEverything_ThenNothingChanged_NoAgent(t *testing.T) {
	root := committedProject(t)

	rep, ad, err := verifyWith(t, root, "", VerificationOptions{}, passPayload())
	if err != nil || rep.Outcome != present.Success {
		t.Fatalf("first run: %v %v", rep, err)
	}
	if !strings.Contains(ad.Ctx.ScopeNote, "every Approved Specification (no earlier passing verification") {
		t.Fatalf("first run should check everything, scope note: %q", ad.Ctx.ScopeNote)
	}

	rep, ad, err = verifyWith(t, root, "", VerificationOptions{}, passPayload())
	if err != nil || rep.Outcome != present.Success {
		t.Fatalf("unchanged project: %v %v", rep, err)
	}
	if ad.Prepared {
		t.Fatal("an unchanged project must not start an Agent")
	}
	if !strings.HasPrefix(rep.Summary, "Verification: nothing changed since the last passing verification") {
		t.Fatalf("unexpected summary %q", rep.Summary)
	}
}

func TestVerificationScope_ChangeAffectingTwoSpecifications_ListsChangedSpecsAndFiles(t *testing.T) {
	root := committedProject(t)
	if _, err := specCreateDefined(root, "Task Reminders"); err != nil { // SPEC-002
		t.Fatal(err)
	}
	if _, err := Approve(root, "SPEC-002", nil, nil); err != nil {
		t.Fatal(err)
	}
	gitRun(t, root, "add", "-A")
	gitRun(t, root, "commit", "-q", "-m", "second spec")
	if _, _, err := verifyWith(t, root, "", VerificationOptions{}, passPayload()); err != nil {
		t.Fatal(err)
	}

	// The change: SPEC-001's text and approval, code, and an uncommitted new file.
	l, _ := project.Locate(root)
	matches, _ := filepath.Glob(filepath.Join(l.SpecificationsDir(), "SPEC-001-*.md"))
	data, _ := os.ReadFile(matches[0])
	if err := os.WriteFile(matches[0], append(data, []byte("\nA completed task can be reopened.\n")...), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Approve(root, "SPEC-001", nil, nil); err != nil {
		t.Fatal(err)
	}
	writeProjectFile(t, root, "app/tasks.py", "def complete(task):\n    task.done = True\n\ndef reopen(task):\n    task.done = False\n")
	gitRun(t, root, "add", "-A")
	gitRun(t, root, "commit", "-q", "-m", "reopen tasks")
	writeProjectFile(t, root, "app/reminders.py", "def remind(task):\n    pass\n")

	_, ad, err := verifyWith(t, root, "", VerificationOptions{}, passPayload())
	if err != nil {
		t.Fatal(err)
	}
	note := ad.Ctx.ScopeNote
	for _, want := range []string{
		"standard check of the changes since the last passing verification",
		"Specifications whose text or approval changed: SPEC-001.",
		"app/reminders.py", "app/tasks.py",
		"every Approved Specification whose behavior the changed files implement or test",
		"Check conflicts against every Approved Specification, also those outside this scope.",
	} {
		if !strings.Contains(note, want) {
			t.Fatalf("expected %q in scope note:\n%s", want, note)
		}
	}
	if strings.Contains(note, "SPEC-002") {
		t.Fatalf("SPEC-002 did not change and must not be named as changed:\n%s", note)
	}
}

// A passing run of uncommitted work makes that work the baseline: the same files count as changed
// again only when their content changes.
func TestVerificationScope_VerifiedUncommittedWork_NotRecheckedUntilItChanges(t *testing.T) {
	root := committedProject(t)
	if _, _, err := verifyWith(t, root, "", VerificationOptions{}, passPayload()); err != nil {
		t.Fatal(err)
	}
	writeProjectFile(t, root, "app/tasks.py", "def complete(task):\n    task.done = True  # edited\n")
	if _, ad, err := verifyWith(t, root, "", VerificationOptions{}, passPayload()); err != nil || !ad.Prepared {
		t.Fatalf("an uncommitted change must be verified: prepared=%v err=%v", ad.Prepared, err)
	}
	if _, ad, _ := verifyWith(t, root, "", VerificationOptions{}, passPayload()); ad.Prepared {
		t.Fatal("verified uncommitted work that did not change again must not be re-checked")
	}
	writeProjectFile(t, root, "app/tasks.py", "def complete(task):\n    task.done = 1\n")
	if _, ad, _ := verifyWith(t, root, "", VerificationOptions{}, passPayload()); !ad.Prepared {
		t.Fatal("a further change to the same file must be verified")
	}
}

// Only a passing run moves the baseline: after a failure, the same change is checked again.
func TestVerificationScope_FailureDoesNotMoveBaseline(t *testing.T) {
	root := committedProject(t)
	if _, _, err := verifyWith(t, root, "", VerificationOptions{}, passPayload()); err != nil {
		t.Fatal(err)
	}
	writeProjectFile(t, root, "app/tasks.py", "def complete(task):\n    pass\n")
	failing := verificationPayload("FAIL", criterion(passingCriterion, map[string]interface{}{"result": "FAIL", "evidence": "done stays false"}))
	if _, _, err := verifyWith(t, root, "", VerificationOptions{}, failing); err == nil {
		t.Fatal("expected a failing verification")
	}
	if _, ad, _ := verifyWith(t, root, "", VerificationOptions{}, passPayload()); !ad.Prepared {
		t.Fatal("after a failure the same change must be verified again")
	}
}

func TestVerificationScope_SinceRef(t *testing.T) {
	root := committedProject(t)
	base := gitRun(t, root, "rev-parse", "HEAD")

	rep, ad, err := verifyWith(t, root, "", VerificationOptions{Since: base}, passPayload())
	if ad.Prepared || err != nil || rep.Outcome != present.Blocked || !strings.Contains(rep.Summary, "nothing was verified") {
		t.Fatalf("no change since the given ref: nothing to verify, never a pass; got %+v err=%v", rep, err)
	}

	writeProjectFile(t, root, "app/tasks.py", "def complete(task):\n    task.done = True\n    task.at = 1\n")
	_, ad, err = verifyWith(t, root, "", VerificationOptions{Since: base}, passPayload())
	if err != nil || !strings.Contains(ad.Ctx.ScopeNote, "changes since "+base) || !strings.Contains(ad.Ctx.ScopeNote, "app/tasks.py") {
		t.Fatalf("expected the change since %s in scope, got %q (err=%v)", base, ad.Ctx.ScopeNote, err)
	}

	if _, _, err := verifyWith(t, root, "", VerificationOptions{Since: "no-such-ref"}, passPayload()); err == nil {
		t.Fatal("an unknown --since ref must be an error")
	}
}

func TestVerificationScope_FullAndTargetModes(t *testing.T) {
	root := committedProject(t)
	if _, _, err := verifyWith(t, root, "", VerificationOptions{}, passPayload()); err != nil {
		t.Fatal(err)
	}

	_, ad, _ := verifyWith(t, root, "", VerificationOptions{Full: true}, passPayload())
	if !ad.Prepared || !strings.Contains(ad.Ctx.ScopeNote, "full independent check of every Approved Specification") ||
		!strings.Contains(ad.Ctx.ScopeNote, "A passing test alone is not enough evidence") {
		t.Fatalf("--full must check everything independently even when nothing changed: %q", ad.Ctx.ScopeNote)
	}

	_, ad, _ = verifyWith(t, root, "SPEC-001", VerificationOptions{}, passPayload())
	if !ad.Prepared || !strings.Contains(ad.Ctx.ScopeNote, "standard check of SPEC-001") {
		t.Fatalf("an explicit target is always checked: %q", ad.Ctx.ScopeNote)
	}
}

func TestVerificationOptions_RefusedForOtherWorkflowsAndBadCombinations(t *testing.T) {
	root := committedProject(t)
	for _, c := range []struct {
		identity, target string
		opts             VerificationOptions
	}{
		{"implementation", "SPEC-001", VerificationOptions{Full: true}},
		{"verification", "", VerificationOptions{Full: true, Since: "HEAD"}},
		{"verification", "SPEC-001", VerificationOptions{Since: "HEAD"}},
	} {
		if _, _, err := RunEvaluationWithOptions(root, c.identity, c.target, "", nil, c.opts); err == nil {
			t.Fatalf("expected %s %q %+v to be refused", c.identity, c.target, c.opts)
		}
	}
}

func TestVerificationEvidence_CitedTestMustExist(t *testing.T) {
	root := committedProject(t)
	cases := []struct {
		test   string
		passes bool
	}{
		{"tests/test_tasks.py::test_complete_marks_task_done", true},
		{"tests/test_tasks.py::test_complete_marks_task_done[case-1]", true},
		{"tests/test_tasks.py", true},
		{"tests/test_tasks.py::test_that_does_not_exist", false},
		{"tests/missing.py::test_complete_marks_task_done", false},
		{"../outside.py::test_x", false},
	}
	for _, c := range cases {
		payload := verificationPayload("PASS", criterion(passingCriterion, map[string]interface{}{"test": c.test}))
		rep, _, err := verifyWith(t, root, "SPEC-001", VerificationOptions{}, payload)
		passed := err == nil && rep.Outcome == present.Success
		if passed != c.passes {
			t.Fatalf("test %q: passed=%v, want %v (%q)", c.test, passed, c.passes, rep.Summary)
		}
		if !c.passes && !strings.Contains(sectionBody(rep, "Criteria"), "was not found in the project") {
			t.Fatalf("test %q: expected the report to say the cited test was not found:\n%s", c.test, sectionBody(rep, "Criteria"))
		}
	}
}
