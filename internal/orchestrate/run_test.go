package orchestrate

import (
	"os"
	"testing"

	"gnomon/internal/adapter"
	"gnomon/internal/present"
	"gnomon/internal/project"
)

func writeFile(t *testing.T, path, content string) error {
	t.Helper()
	return os.WriteFile(path, []byte(content), 0o644)
}

// --- Run resolves the correct Contract by identity, and reuses the exact same engine as the
// dedicated commands ---

func TestRun_ResolvesCorrectWorkflowByIdentity(t *testing.T) {
	root := t.TempDir()
	configureGitIdentity(t, root)
	if _, err := Init(root); err != nil {
		t.Fatal(err)
	}
	l, err := project.Locate(root)
	if err != nil {
		t.Fatal(err)
	}
	wf, path, err := resolveWorkflowByIdentity(l, "verification")
	if err != nil {
		t.Fatal(err)
	}
	if wf.Identity != "verification" {
		t.Fatalf("expected identity verification, got %q", wf.Identity)
	}
	if wf.Result.TerminalPath != "summary.aggregate" {
		t.Fatalf("expected the nested Result Contract to be loaded correctly via identity resolution, got terminal_path=%q", wf.Result.TerminalPath)
	}
	if got, want := path, l.WorkflowsDir()+"/verification.md"; got != want {
		t.Fatalf("expected the real resolved path %q, got %q", want, got)
	}
}

func TestRun_UnknownIdentity_DeterministicRefusal(t *testing.T) {
	root := t.TempDir()
	configureGitIdentity(t, root)
	if _, err := Init(root); err != nil {
		t.Fatal(err)
	}

	report, err := Run(root, "no-such-workflow", "", "", nil)
	if err == nil {
		t.Fatalf("expected an unknown workflow identity to be refused")
	}
	if report != nil {
		t.Fatalf("expected no report for an unresolvable identity, got: %+v", report)
	}
}

func TestRun_DuplicateIdentity_DeterministicRefusal(t *testing.T) {
	root := t.TempDir()
	configureGitIdentity(t, root)
	if _, err := Init(root); err != nil {
		t.Fatal(err)
	}
	l, err := project.Locate(root)
	if err != nil {
		t.Fatal(err)
	}
	duplicate := "---\nidentity: bootstrap\nspecification_reference: none\nresult:\n" +
		"  terminal_path: outcome\n  classification:\n    OK: success\n  schema:\n    type: object\n" +
		"    required: [outcome]\n    properties:\n      outcome:\n        type: string\n        enum: [OK]\n" +
		"---\n# Duplicate\n"
	if err := writeFile(t, l.WorkflowsDir()+"/bootstrap-copy.md", duplicate); err != nil {
		t.Fatal(err)
	}

	report, err := Run(root, "bootstrap", "", "", nil)
	if err == nil {
		t.Fatalf("expected a duplicate identity to be refused rather than routed arbitrarily")
	}
	if report != nil {
		t.Fatalf("expected no report for a duplicate-identity refusal, got: %+v", report)
	}
}

// --- Contract eligibility is enforced before Agent resolution, exactly as every dedicated
// command's is, for required/optional/none specification_reference alike ---

func TestRun_RequiredReference_NonexistentTargetBlocksBeforeAgentResolution(t *testing.T) {
	sandboxAgentConfig(t)
	root := t.TempDir()
	configureGitIdentity(t, root)
	if _, err := Init(root); err != nil {
		t.Fatal(err)
	}

	chooserCalled := false
	chooser := func(providers []string) (string, error) {
		chooserCalled = true
		return "claude", nil
	}
	report, err := Run(root, "implementation", "SPEC-001", "", chooser)
	if err == nil {
		t.Fatalf("expected refusal: SPEC-001 does not exist")
	}
	if report == nil || report.Outcome != present.Blocked {
		t.Fatalf("expected a Blocked report, got: %+v", report)
	}
	if chooserCalled {
		t.Fatalf("the Agent provider must never be resolved when the pre-start eligibility check fails")
	}
}

func TestRun_OptionalReference_NoTargetIsAlwaysEligible(t *testing.T) {
	sandboxAgentConfig(t)
	root := t.TempDir()
	configureGitIdentity(t, root)
	if _, err := Init(root); err != nil {
		t.Fatal(err)
	}

	// An invalid --agent override proves eligibility passed and Agent resolution was actually
	// reached — without launching any real process, since resolution fails before Prepare/Run.
	report, err := Run(root, "testing", "", "not-a-real-provider", nil)
	if err == nil {
		t.Fatalf("expected the invalid --agent override to be refused")
	}
	if report == nil || report.Outcome != present.Failed {
		t.Fatalf("expected a Failed (Agent-resolution) report, not an eligibility Blocked one, got: %+v", report)
	}
}

func TestRun_NoneReference_AlwaysEligible(t *testing.T) {
	sandboxAgentConfig(t)
	root := t.TempDir()
	configureGitIdentity(t, root)
	if _, err := Init(root); err != nil {
		t.Fatal(err)
	}

	report, err := Run(root, "verification", "src/auth/", "not-a-real-provider", nil)
	if err == nil {
		t.Fatalf("expected the invalid --agent override to be refused")
	}
	if report == nil || report.Outcome != present.Failed {
		t.Fatalf("expected a Failed (Agent-resolution) report, not an eligibility Blocked one, got: %+v", report)
	}
}

// --- Target semantics: run must never treat a generic/untyped target as a Specification, or
// vice versa, purely from the resolved workflow's own specification_reference ---

func TestTargetKindFor_DerivedFromSpecificationReferenceAlone(t *testing.T) {
	root := t.TempDir()
	configureGitIdentity(t, root)
	if _, err := Init(root); err != nil {
		t.Fatal(err)
	}
	l, err := project.Locate(root)
	if err != nil {
		t.Fatal(err)
	}

	cases := map[string]targetKind{
		"implementation":           targetSpec,    // required
		"testing":                  targetSpec,    // optional
		"specification-definition": targetSpec,    // required
		"verification":             targetGeneric, // none
		"review":                   targetGeneric, // none
		"bootstrap":                targetGeneric, // none, no dedicated target at all
	}
	for identity, want := range cases {
		wf, _, err := resolveWorkflowByIdentity(l, identity)
		if err != nil {
			t.Fatalf("resolving %s: %v", identity, err)
		}
		if got := targetKindFor(wf); got != want {
			t.Errorf("targetKindFor(%s) = %v, want %v", identity, got, want)
		}
	}
}

// --- Dedicated commands and Run converge on the same generic execution engine and the same
// Result Contract handling for the same workflow ---

func TestRun_ConvergesWithVerify_NestedTerminalPathAndBlockedClassification(t *testing.T) {
	root := t.TempDir()
	configureGitIdentity(t, root)
	if _, err := Init(root); err != nil {
		t.Fatal(err)
	}
	l, wf, workflowPath, err := prepareWorkflow(root, "verification.md")
	if err != nil {
		t.Fatal(err)
	}

	scripted := &payloadAdapter{
		FakeAdapter: &adapter.FakeAdapter{},
		Payload: map[string]interface{}{
			"summary": map[string]interface{}{"aggregate": "FAIL"},
		},
	}

	// Simulates exactly what Run does once it has resolved (l, wf, workflowPath) by identity: the
	// same runEligibleWorkflow → runWorkflow → runWorkflowWithAdapter chain Verify itself uses.
	kind := targetKindFor(wf)
	if kind != targetGeneric {
		t.Fatalf("expected targetGeneric for specification_reference: none, got %v", kind)
	}
	report, err := runWorkflowWithAdapter(root, l, wf, workflowPath, kind, "src/", scripted)
	if err == nil {
		t.Fatalf("expected a BLOCKED workflow conclusion to be reported as a non-nil error")
	}
	if report.Outcome != present.Blocked {
		t.Fatalf("expected Blocked (a valid negative workflow conclusion), got %v", report.Outcome)
	}
	if report.Outcome == present.Failed {
		t.Fatalf("a Contract-classified blocked result must never become a process/protocol failure merely because it was reached via the same engine Run also uses")
	}
}
