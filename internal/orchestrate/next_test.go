package orchestrate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gnomon/internal/present"
	"gnomon/internal/project"
)

// commitEverything is defined once, in status_test.go, and reused here — both files share the
// orchestrate test package.

func TestNext_ZeroSpecifications_SuggestsDiscoveryOrCreate(t *testing.T) {
	root := t.TempDir()
	configureGitIdentity(t, root)
	if _, err := Init(root); err != nil {
		t.Fatal(err)
	}

	report, err := Next(root)
	if err != nil {
		t.Fatal(err)
	}
	if report.Outcome != present.Success {
		t.Fatalf("expected Success, got %v", report.Outcome)
	}
	rendered := present.Render(report, false)
	if !strings.Contains(rendered, "spec discover") || !strings.Contains(rendered, "spec create") {
		t.Fatalf("expected zero-Specification guidance to name both mechanisms that create one: %s", rendered)
	}
}

func TestNext_DraftSpecification_OffersDefineAndApprove_NotImplementOrTest(t *testing.T) {
	root := t.TempDir()
	configureGitIdentity(t, root)
	if _, err := Init(root); err != nil {
		t.Fatal(err)
	}
	if _, err := SpecCreate(root, "Password Reset"); err != nil {
		t.Fatal(err)
	}

	report, err := Next(root)
	if err != nil {
		t.Fatal(err)
	}
	rendered := present.Render(report, false)
	if !strings.Contains(rendered, "gnomon run specification-definition SPEC-001") {
		t.Fatalf("expected Draft guidance to offer spec define, matching cli/COMMAND_SURFACE.md's own worked example: %s", rendered)
	}
	if !strings.Contains(rendered, "gnomon approve SPEC-001") {
		t.Fatalf("expected Draft guidance to also offer approve, without presuming which applies: %s", rendered)
	}
	if strings.Contains(rendered, "gnomon run implementation SPEC-001") {
		t.Fatalf("did not expect implement to be offered for a Draft Specification: %s", rendered)
	}
	if strings.Contains(rendered, "gnomon run testing SPEC-001") {
		t.Fatalf("did not expect test to be offered for a Draft, Approved-gated Specification: %s", rendered)
	}
}

func TestNext_ApprovedSpecification_OffersImplementTestRevoke_NotApprove(t *testing.T) {
	root := setupApprovedSpec(t)

	report, err := Next(root)
	if err != nil {
		t.Fatal(err)
	}
	rendered := present.Render(report, false)
	for _, want := range []string{"gnomon run implementation SPEC-001", "gnomon run testing SPEC-001", "gnomon revoke SPEC-001", "gnomon run specification-definition SPEC-001"} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("expected Approved guidance to include %q: %s", want, rendered)
		}
	}
	if strings.Contains(rendered, "gnomon approve SPEC-001") {
		t.Fatalf("did not expect approve to be offered for an already-Approved Specification: %s", rendered)
	}
}

func TestNext_RevokedApproval_TreatedIdenticallyToOrdinaryDraft(t *testing.T) {
	root := setupApprovedSpec(t)
	if _, err := Revoke(root, "SPEC-001", nil); err != nil {
		t.Fatal(err)
	}

	report, err := Next(root)
	if err != nil {
		t.Fatal(err)
	}
	rendered := present.Render(report, false)
	if !strings.Contains(rendered, "gnomon run specification-definition SPEC-001") || !strings.Contains(rendered, "gnomon approve SPEC-001") {
		t.Fatalf("expected a revoked (now Draft) Specification to receive ordinary Draft guidance, with no distinct 'revoked' handling invented: %s", rendered)
	}
}

func TestNext_StaleApprovalEvidence_TreatedIdenticallyToOrdinaryDraft(t *testing.T) {
	root := setupApprovedSpec(t)
	l, err := project.Locate(root)
	if err != nil {
		t.Fatal(err)
	}
	specPath := filepath.Join(l.SpecificationsDir(), "SPEC-001-password-reset.md")
	data, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(specPath, append(data, []byte("\nContent changed after approval.\n")...), 0o644); err != nil {
		t.Fatal(err)
	}

	report, err := Next(root)
	if err != nil {
		t.Fatal(err)
	}
	rendered := present.Render(report, false)
	if !strings.Contains(rendered, "gnomon run specification-definition SPEC-001") || !strings.Contains(rendered, "gnomon approve SPEC-001") {
		t.Fatalf("expected stale approval evidence to derive ordinary Draft guidance, no distinct handling invented: %s", rendered)
	}
}

func TestNext_MalformedApprovalEvidence_ExcludedNotFatal(t *testing.T) {
	root := setupApprovedSpec(t)
	l, err := project.Locate(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(l.ApprovalsDir(), "SPEC-001", "bad.grant.json"), []byte("not json"), 0o644); err != nil {
		t.Fatal(err)
	}

	report, err := Next(root)
	if err != nil {
		t.Fatalf("expected malformed evidence to be excluded, not fatal, to next: %v", err)
	}
	rendered := present.Render(report, false)
	if !strings.Contains(rendered, "gnomon run implementation SPEC-001") {
		t.Fatalf("expected the real, valid grant to still be honored: %s", rendered)
	}
}

func TestNext_MultipleDraftSpecifications_AllListedNoneChosen(t *testing.T) {
	root := t.TempDir()
	configureGitIdentity(t, root)
	if _, err := Init(root); err != nil {
		t.Fatal(err)
	}
	if _, err := SpecCreate(root, "Password Reset"); err != nil {
		t.Fatal(err)
	}
	if _, err := SpecCreate(root, "Email Verification"); err != nil {
		t.Fatal(err)
	}

	report, err := Next(root)
	if err != nil {
		t.Fatal(err)
	}
	rendered := present.Render(report, false)
	if !strings.Contains(rendered, "gnomon run specification-definition SPEC-001") || !strings.Contains(rendered, "gnomon run specification-definition SPEC-002") {
		t.Fatalf("expected guidance for both Draft Specifications, no single one chosen: %s", rendered)
	}
	if report.Target != "" {
		t.Fatalf("expected no single Specification to be singled out as report.Target, got %q", report.Target)
	}
	if !strings.Contains(report.Summary, "no priority implied") {
		t.Fatalf("expected the summary to state explicitly that no precedence is implied among multiple candidates: %s", report.Summary)
	}
}

func TestNext_MultipleApprovedSpecifications_AllListedNoneChosen(t *testing.T) {
	root := t.TempDir()
	configureGitIdentity(t, root)
	if _, err := Init(root); err != nil {
		t.Fatal(err)
	}
	if _, err := SpecCreate(root, "Password Reset"); err != nil {
		t.Fatal(err)
	}
	if _, err := SpecCreate(root, "Email Verification"); err != nil {
		t.Fatal(err)
	}
	if _, err := Approve(root, "SPEC-001", nil); err != nil {
		t.Fatal(err)
	}
	if _, err := Approve(root, "SPEC-002", nil); err != nil {
		t.Fatal(err)
	}

	report, err := Next(root)
	if err != nil {
		t.Fatal(err)
	}
	rendered := present.Render(report, false)
	if !strings.Contains(rendered, "gnomon run implementation SPEC-001") || !strings.Contains(rendered, "gnomon run implementation SPEC-002") {
		t.Fatalf("expected guidance for both Approved Specifications, no single one chosen: %s", rendered)
	}
}

func TestNext_MixedDraftAndApproved_BothRepresentedDistinctly(t *testing.T) {
	root := t.TempDir()
	configureGitIdentity(t, root)
	if _, err := Init(root); err != nil {
		t.Fatal(err)
	}
	if _, err := SpecCreate(root, "Password Reset"); err != nil {
		t.Fatal(err)
	}
	if _, err := SpecCreate(root, "Email Verification"); err != nil {
		t.Fatal(err)
	}
	if _, err := Approve(root, "SPEC-002", nil); err != nil {
		t.Fatal(err)
	}

	report, err := Next(root)
	if err != nil {
		t.Fatal(err)
	}
	rendered := present.Render(report, false)
	if !strings.Contains(rendered, "gnomon approve SPEC-001") {
		t.Fatalf("expected the Draft Specification's own guidance: %s", rendered)
	}
	if !strings.Contains(rendered, "gnomon run implementation SPEC-002") {
		t.Fatalf("expected the Approved Specification's own guidance: %s", rendered)
	}
}

func TestNext_UncommittedChanges_SuggestsFinalizeAsCoarseSignal(t *testing.T) {
	root := t.TempDir()
	configureGitIdentity(t, root)
	if _, err := Init(root); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "untracked.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	report, err := Next(root)
	if err != nil {
		t.Fatal(err)
	}
	rendered := present.Render(report, false)
	if !strings.Contains(rendered, "gnomon finalize") {
		t.Fatalf("expected uncommitted changes to surface a coarse Git Finalization signal, per cli/DERIVED_FACTS.md: %s", rendered)
	}
}

func TestNext_CleanWorkingTree_NoFinalizeSuggestion(t *testing.T) {
	root := t.TempDir()
	configureGitIdentity(t, root)
	if _, err := Init(root); err != nil {
		t.Fatal(err)
	}
	commitEverything(t, root)

	report, err := Next(root)
	if err != nil {
		t.Fatal(err)
	}
	rendered := present.Render(report, false)
	if strings.Contains(rendered, "gnomon finalize") {
		t.Fatalf("did not expect a finalize suggestion when the working tree is clean: %s", rendered)
	}
}

// TestNext_NotAGitRepository_SpecGuidanceStillProduced_GitFinalizationNotApplicable is the
// dogfooding regression: an initialized Gnomon project that is not (yet) a Git repository must
// still receive full Specification guidance from `gnomon next` — Git is a capability Git-related
// guidance uses, not a prerequisite for Specification lifecycle inspection.
func TestNext_NotAGitRepository_SpecGuidanceStillProduced_GitFinalizationNotApplicable(t *testing.T) {
	root := t.TempDir()
	if _, err := Init(root); err != nil {
		t.Fatal(err)
	}
	if _, err := SpecCreate(root, "Password Reset"); err != nil {
		t.Fatal(err)
	}

	report, err := Next(root)
	if err != nil {
		t.Fatalf("expected a non-Git project to still produce Next guidance, got: %v", err)
	}
	if report.Outcome != present.Success {
		t.Fatalf("expected Success, got %v", report.Outcome)
	}
	rendered := present.Render(report, false)
	if !strings.Contains(rendered, "gnomon run specification-definition SPEC-001") || !strings.Contains(rendered, "gnomon approve SPEC-001") {
		t.Fatalf("expected ordinary Draft Specification guidance despite no Git repository: %s", rendered)
	}
	if !strings.Contains(rendered, "Not a Git repository") || !strings.Contains(rendered, "Git Finalization is not applicable") {
		t.Fatalf("expected an explicit, concise statement that Git Finalization is not applicable: %s", rendered)
	}
}

// TestNext_UnexpectedGitFailure_PreservesSpecGuidance_NotMisclassifiedAsNoRepo proves a genuinely
// unexpected Git failure is reported distinctly from "not a repository" and does not discard the
// Specification guidance already derived.
func TestNext_UnexpectedGitFailure_PreservesSpecGuidance_NotMisclassifiedAsNoRepo(t *testing.T) {
	root := t.TempDir()
	configureGitIdentity(t, root)
	if _, err := Init(root); err != nil {
		t.Fatal(err)
	}
	if _, err := SpecCreate(root, "Password Reset"); err != nil {
		t.Fatal(err)
	}
	// Corrupt the index so `git rev-parse --git-dir` still succeeds (it is still a Git
	// repository) but `git status` fails for a reason that is NOT "not a repository".
	if err := os.WriteFile(filepath.Join(root, ".git", "index"), []byte("garbage not an index"), 0o644); err != nil {
		t.Fatal(err)
	}

	report, err := Next(root)
	if err == nil {
		t.Fatalf("expected an unexpected Git failure to be reported as an error")
	}
	if report == nil {
		t.Fatalf("expected a Report to still be returned despite the Git failure")
	}
	rendered := present.Render(report, false)
	if !strings.Contains(rendered, "gnomon run specification-definition SPEC-001") {
		t.Fatalf("expected Specification guidance to survive an unexpected Git failure: %s", rendered)
	}
	if strings.Contains(rendered, "Not a Git repository") {
		t.Fatalf("did not expect an unexpected Git failure to be misclassified as 'not a Git repository': %s", rendered)
	}
	if strings.Contains(err.Error(), "exit status 128") && !strings.Contains(err.Error(), ":") {
		t.Fatalf("expected a real diagnostic, not a bare exit code: %v", err)
	}
}

func TestNext_UnsupportedContractVersion_RefusesAndPointsAtValidate(t *testing.T) {
	root := t.TempDir()
	configureGitIdentity(t, root)
	if _, err := Init(root); err != nil {
		t.Fatal(err)
	}
	l, err := project.Locate(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(l.ContractVersionPath(), []byte("999\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	report, err := Next(root)
	if err == nil {
		t.Fatalf("expected an unsupported contract version to be refused")
	}
	if report == nil || report.Outcome != present.Failed {
		t.Fatalf("expected a Failed report, got: %+v", report)
	}
	if !strings.Contains(report.Next, "gnomon validate") {
		t.Fatalf("expected next to point at validate for the structural diagnosis, got Next=%q", report.Next)
	}
}

func TestNext_NeverInvokesAnAgent(t *testing.T) {
	sandboxAgentConfig(t)
	root := setupApprovedSpec(t)
	if _, err := Next(root); err != nil {
		t.Fatalf("expected Next to succeed without ever needing an Agent provider: %v", err)
	}
}

func TestNext_NeverMutatesProjectFiles(t *testing.T) {
	root := setupApprovedSpec(t)
	l, err := project.Locate(root)
	if err != nil {
		t.Fatal(err)
	}

	before := map[string]string{}
	_ = filepath.Walk(l.GnomonRoot(), func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		data, _ := os.ReadFile(path)
		before[path] = string(data)
		return nil
	})

	if _, err := Next(root); err != nil {
		t.Fatal(err)
	}

	for path, want := range before {
		got, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("expected %s to still exist after Next: %v", path, err)
		}
		if string(got) != want {
			t.Fatalf("expected Next to never modify %s, but its content changed", path)
		}
	}
}

// TestNext_NeverExecutesAWorkflow proves next only ever names a command as guidance text — it
// never calls Run, runWorkflow, or anything that would create a transient result directory, even
// when a Specification is fully eligible for every gated workflow.
func TestNext_NeverExecutesAWorkflow(t *testing.T) {
	root := setupApprovedSpec(t)

	if _, err := Next(root); err != nil {
		t.Fatal(err)
	}

	runtimeDir := filepath.Join(root, ".gnomon-runtime")
	entries, err := os.ReadDir(runtimeDir)
	if err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".result.json") {
			t.Fatalf("expected next to never execute a workflow or leave a transient result behind, found %s", e.Name())
		}
	}
}

// TestNext_NeverMentionsDescribeOrBootstrap guards against next's scope growing to include
// always-valid workflows, which carry no state-dependent information.
func TestNext_NeverMentionsDescribeOrBootstrap(t *testing.T) {
	root := setupApprovedSpec(t)

	report, err := Next(root)
	if err != nil {
		t.Fatal(err)
	}
	rendered := present.Render(report, false)
	if strings.Contains(rendered, "describe") || strings.Contains(rendered, "bootstrap") {
		t.Fatalf("expected next to never mention describe/bootstrap, got: %s", rendered)
	}
}

func TestNext_DiffersFromRun_NeverAcceptsAWorkflowIdentityArgument(t *testing.T) {
	// This is a structural/behavioral distinction, not a syntax one (covered in cmd/gnomon):
	// Next's own signature takes only a project root — it has no way to be pointed at a specific
	// workflow identity or target the way Run does.
	root := setupApprovedSpec(t)
	report, err := Next(root)
	if err != nil {
		t.Fatal(err)
	}
	if report.Target != "" {
		t.Fatalf("Next must never resolve or report a single Target the way Run's underlying engine does, got %q", report.Target)
	}
}
