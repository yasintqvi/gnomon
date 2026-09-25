package orchestrate

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"gnomon/internal/adapter"
	"gnomon/internal/agentconfig"
	"gnomon/internal/contract"
	"gnomon/internal/present"
	"gnomon/internal/project"
)

func configureGitIdentity(t *testing.T, root string) {
	t.Helper()
	run := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = root
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v failed: %v\n%s", args, err, out)
		}
	}
	run("init", "-q")
	run("config", "user.name", "Jane Doe")
	run("config", "user.email", "jane@example.com")
}

// sandboxAgentConfig points internal/agentconfig's persisted state at a fresh temp directory for
// the duration of one test, so tests never read or write the real developer's per-user config.
func sandboxAgentConfig(t *testing.T) {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))
}

// mustPrepareImplementation resolves the (Layout, Workflow) pair implementWithAdapter needs,
// asserting the given Specification is eligible — the precondition every implementWithAdapter
// test below already assumes, matching what Implement itself would have already confirmed by
// the point it calls ResolveAgent.
func mustPrepareImplementation(t *testing.T, root, specID string) (project.Layout, contract.Workflow) {
	t.Helper()
	l, wf, elig, err := prepareImplementation(root, specID)
	if err != nil {
		t.Fatalf("prepareImplementation: %v", err)
	}
	if !elig.Eligible {
		t.Fatalf("expected %s to be eligible, got: %s", specID, elig.Reason)
	}
	return l, wf
}

func TestFullSlice_InitCreateApproveImplement(t *testing.T) {
	root := t.TempDir()
	configureGitIdentity(t, root)

	if _, err := Init(root); err != nil {
		t.Fatalf("init: %v", err)
	}

	specReport, err := SpecCreate(root, "Password Reset")
	if err != nil {
		t.Fatalf("spec create: %v", err)
	}
	if specReport.Target != "SPEC-001" {
		t.Fatalf("unexpected spec create report target: %v", specReport.Target)
	}

	if _, err := Approve(root, "SPEC-001", nil); err != nil {
		t.Fatalf("approve: %v", err)
	}

	// scriptedAdapter learns the CLI-generated run_id/workflow identity in Prepare, then scripts
	// a matching valid result before Run executes — proving the round-trip end to end.
	scripted := &scriptedAdapter{FakeAdapter: &adapter.FakeAdapter{}}

	l, wf, elig, err := prepareImplementation(root, "SPEC-001")
	if err != nil {
		t.Fatalf("prepareImplementation: %v", err)
	}
	if !elig.Eligible {
		t.Fatalf("expected SPEC-001 to be eligible once approved, got: %s", elig.Reason)
	}

	report, err := implementWithAdapter(root, l, wf, "SPEC-001", scripted)
	if err != nil {
		t.Fatalf("implement: %v (report: %+v)", err, report)
	}

	if report.Outcome != present.Success {
		t.Fatalf("expected Success outcome, got %v", report.Outcome)
	}
	if report.Summary != "Implementation complete" {
		t.Fatalf("unexpected summary: %q", report.Summary)
	}
	rendered := present.Render(report, true)
	if !strings.Contains(rendered, "run_id:") {
		t.Fatalf("expected run_id to be available in verbose diagnostics: %s", rendered)
	}
	if !strings.Contains(rendered, "Delivered") || !strings.Contains(rendered, "a trivial change") {
		t.Fatalf("expected the Delivered section to be reported: %s", rendered)
	}

	// The transient result must be gone after consumption — no hidden state survives the run.
	runtimeDir := filepath.Join(root, ".gnomon-runtime")
	entries, err := os.ReadDir(runtimeDir)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".result.json") {
			t.Fatalf("expected no leftover transient result files, found %s", e.Name())
		}
	}
}

// TestSpecCreate_NextRecommendsSpecWorkspace proves the guidance shown right after creating a
// Draft names an actual command for every step, rather than pairing a real "gnomon approve"
// command with prose ("Define its content...") that named no command for the first step at all.
func TestSpecCreate_NextRecommendsSpecWorkspace(t *testing.T) {
	root := t.TempDir()
	configureGitIdentity(t, root)
	if _, err := Init(root); err != nil {
		t.Fatal(err)
	}

	report, err := SpecCreate(root, "Password Reset")
	if err != nil {
		t.Fatalf("spec create: %v", err)
	}
	if !strings.Contains(report.Next, "gnomon spec SPEC-001") {
		t.Fatalf("expected the Specification workspace recommended, got %q", report.Next)
	}
}

// TestImplement_IneligibleNeverResolvesAgent proves the ordering Implement's own doc comment
// promises: for a Draft Specification, the pre-start eligibility check must refuse before the
// Agent provider is ever resolved or prompted for — an ineligible invocation must never trigger
// the first-use "choose your Agent" prompt for a run that was always going to be blocked.
func TestImplement_IneligibleNeverResolvesAgent(t *testing.T) {
	sandboxAgentConfig(t)
	root := t.TempDir()
	configureGitIdentity(t, root)
	if _, err := Init(root); err != nil {
		t.Fatal(err)
	}
	if _, err := SpecCreate(root, "Password Reset"); err != nil {
		t.Fatal(err)
	}
	// SPEC-001 remains Draft — never approved.

	chooserCalled := false
	chooser := func(providers []string) (string, error) {
		chooserCalled = true
		return "claude", nil
	}

	report, err := Implement(root, "SPEC-001", "", chooser)
	if err == nil {
		t.Fatalf("expected implement to refuse for a Draft Specification")
	}
	if report == nil || report.Outcome != present.Blocked {
		t.Fatalf("expected a Blocked report, got: %+v", report)
	}
	if chooserCalled {
		t.Fatalf("the Agent provider must never be resolved or prompted for when the pre-start eligibility check fails")
	}
	// Regression guard: this used to assume the block was always "not yet Approved" and
	// recommend `gnomon approve` directly — wrong whenever elig.Reason is something else, and
	// paired with prose ("Define %s further") that named no actual command at all. The
	// Specification workspace is correct regardless of which reason actually applies.
	if !strings.Contains(report.Next, "gnomon spec SPEC-001") {
		t.Fatalf("expected the Specification workspace recommended, got %q", report.Next)
	}
}

// scriptedAdapter writes a valid result envelope matching whatever run_id/workflow Prepare
// actually received, proving the CLI-generated identity round-trips correctly end to end.
type scriptedAdapter struct {
	*adapter.FakeAdapter
}

func (s *scriptedAdapter) Prepare(ctx adapter.Context) error {
	if err := s.FakeAdapter.Prepare(ctx); err != nil {
		return err
	}
	payload, _ := json.Marshal(map[string]interface{}{
		"outcome":   "IMPLEMENTATION_COMPLETE",
		"delivered": "a trivial change",
	})
	env, _ := json.Marshal(map[string]interface{}{
		"workflow": ctx.WorkflowIdentity,
		"run_id":   ctx.RunID,
		"payload":  json.RawMessage(payload),
	})
	s.FakeAdapter.ResultContent = env
	return nil
}

// TestApprove_NextRecommendsSpecWorkspaceAndImplementation is the direct regression guard for the
// stale "gnomon implement <SPEC-id>" recommendation this command used to print — that command no
// longer exists (it is reachable only through the Specification workspace or `gnomon run
// implementation <SPEC-id>`); this proves the real, current guidance takes its place.
func TestApprove_NextRecommendsSpecWorkspaceAndImplementation(t *testing.T) {
	root := t.TempDir()
	configureGitIdentity(t, root)
	if _, err := Init(root); err != nil {
		t.Fatal(err)
	}
	if _, err := SpecCreate(root, "Password Reset"); err != nil {
		t.Fatal(err)
	}

	report, err := Approve(root, "SPEC-001", nil)
	if err != nil {
		t.Fatalf("approve: %v", err)
	}
	if !strings.Contains(report.Next, "gnomon spec SPEC-001") {
		t.Fatalf("expected the Specification workspace recommended, got %q", report.Next)
	}
	if !strings.Contains(report.Next, "gnomon run implementation SPEC-001") {
		t.Fatalf("expected the advanced gnomon run form recommended, got %q", report.Next)
	}
	if strings.Contains(report.Next, "gnomon implement ") {
		t.Fatalf("did not expect the no-longer-registered \"gnomon implement\" command, got %q", report.Next)
	}
}

// TestSpecWorkspaceNext_NeverHighlightsAnUnavailableAction is the direct proof that
// specWorkspaceNext derives its "Advanced:" line from SpecActions' own real eligibility
// derivation — never from the caller's hint alone. Asking it to highlight "implement" for a
// Draft Specification (Implementation is never eligible until Approved) must not show it, even
// though the caller asked for exactly that verb: the caller only names a candidate worth
// checking, this function is the one place that actually verifies it against authoritative
// state before ever naming it as a recommendation.
func TestSpecWorkspaceNext_NeverHighlightsAnUnavailableAction(t *testing.T) {
	root := t.TempDir()
	configureGitIdentity(t, root)
	if _, err := Init(root); err != nil {
		t.Fatal(err)
	}
	if _, err := SpecCreate(root, "Password Reset"); err != nil {
		t.Fatal(err)
	}
	l, err := project.Locate(root)
	if err != nil {
		t.Fatal(err)
	}

	next := specWorkspaceNext(l, "SPEC-001", "implement")
	if strings.Contains(next, "Advanced:") {
		t.Fatalf("expected no Advanced line for a Draft Specification (Implementation is not yet eligible), got %q", next)
	}
	if !strings.Contains(next, "gnomon spec SPEC-001") {
		t.Fatalf("expected the workspace pointer still shown, got %q", next)
	}
}

// TestSpecWorkspaceNext_HighlightsAvailableAction proves the converse: once the highlighted verb
// genuinely is Available, per the same SpecActions/facts.Eligible derivation `gnomon next` and
// the workspace both use, the Advanced line names its real, currently valid command — sourced
// from SpecAction.Command itself, never reconstructed from a hardcoded identity string here.
func TestSpecWorkspaceNext_HighlightsAvailableAction(t *testing.T) {
	root := t.TempDir()
	configureGitIdentity(t, root)
	if _, err := Init(root); err != nil {
		t.Fatal(err)
	}
	if _, err := SpecCreate(root, "Password Reset"); err != nil {
		t.Fatal(err)
	}
	if _, err := Approve(root, "SPEC-001", nil); err != nil {
		t.Fatal(err)
	}
	l, err := project.Locate(root)
	if err != nil {
		t.Fatal(err)
	}

	next := specWorkspaceNext(l, "SPEC-001", "implement")
	if !strings.Contains(next, "Advanced:\n  gnomon run implementation SPEC-001") {
		t.Fatalf("expected the real gnomon run implementation command, got %q", next)
	}
}

func TestApprove_RefusesWhenNoIdentityIsAvailable(t *testing.T) {
	root := t.TempDir()
	// A real, unconfigured temp repo — no git identity, and no prompt supplied.
	cmd := exec.Command("git", "init", "-q")
	cmd.Dir = root
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git init failed: %v\n%s", err, out)
	}
	t.Setenv("HOME", t.TempDir())
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	if _, err := Init(root); err != nil {
		t.Fatal(err)
	}
	if _, err := SpecCreate(root, "Password Reset"); err != nil {
		t.Fatal(err)
	}
	if _, err := Approve(root, "SPEC-001", nil); err == nil {
		t.Fatalf("expected approve to refuse when no Human identity can be resolved and no prompt is available")
	}
}

// setupApprovedSpec produces a project with SPEC-001 Approved and ready for Implement, so each
// presentation-focused test below can start directly from an eligible run.
func setupApprovedSpec(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	configureGitIdentity(t, root)
	if _, err := Init(root); err != nil {
		t.Fatalf("init: %v", err)
	}
	if _, err := SpecCreate(root, "Password Reset"); err != nil {
		t.Fatalf("spec create: %v", err)
	}
	if _, err := Approve(root, "SPEC-001", nil); err != nil {
		t.Fatalf("approve: %v", err)
	}
	return root
}

// payloadAdapter writes a result envelope built from whatever payload the test supplies, with the
// CLI-generated workflow identity/run_id filled in during Prepare — like scriptedAdapter, but
// with a caller-controlled payload so BLOCKED and empty-optional-field cases can be exercised too.
type payloadAdapter struct {
	*adapter.FakeAdapter
	Payload map[string]interface{}
}

func (s *payloadAdapter) Prepare(ctx adapter.Context) error {
	if err := s.FakeAdapter.Prepare(ctx); err != nil {
		return err
	}
	payload, _ := json.Marshal(s.Payload)
	env, _ := json.Marshal(map[string]interface{}{
		"workflow": ctx.WorkflowIdentity,
		"run_id":   ctx.RunID,
		"payload":  json.RawMessage(payload),
	})
	s.FakeAdapter.ResultContent = env
	return nil
}

func TestImplement_GnomonTerminationAfterValidResult_IsNotPresentedAsFailure(t *testing.T) {
	root := setupApprovedSpec(t)
	l, wf := mustPrepareImplementation(t, root, "SPEC-001")

	// A Gnomon-initiated SIGTERM: the Agent's own exit code reflects the signal (143), but
	// Terminated=true records that Gnomon asked for this, not that the Agent failed.
	scripted := &payloadAdapter{
		FakeAdapter: &adapter.FakeAdapter{ExitCode: 143, Terminated: true},
		Payload:     map[string]interface{}{"outcome": "IMPLEMENTATION_COMPLETE", "delivered": "a trivial change"},
	}

	report, err := implementWithAdapter(root, l, wf, "SPEC-001", scripted)
	if err != nil {
		t.Fatalf("implement: %v (report: %+v)", err, report)
	}
	if report.Outcome != present.Success {
		t.Fatalf("expected Success outcome despite the Gnomon-initiated exit signal, got %v", report.Outcome)
	}
	rendered := present.Render(report, false)
	if strings.Contains(strings.ToLower(rendered), "fail") {
		t.Fatalf("a Gnomon-initiated termination after a valid result must never read as a failure: %s", rendered)
	}
	verboseRendered := present.Render(report, true)
	if !strings.Contains(verboseRendered, "ended automatically once Gnomon confirmed the result") {
		t.Fatalf("expected verbose diagnostics to explain the intentional termination: %s", verboseRendered)
	}
	if !strings.Contains(verboseRendered, "143") {
		t.Fatalf("expected the real exit code to remain available in verbose diagnostics: %s", verboseRendered)
	}
}

func TestImplement_Blocked(t *testing.T) {
	root := setupApprovedSpec(t)
	l, wf := mustPrepareImplementation(t, root, "SPEC-001")

	scripted := &payloadAdapter{
		FakeAdapter: &adapter.FakeAdapter{},
		Payload: map[string]interface{}{
			"outcome":              "BLOCKED",
			"remaining_unresolved": "needs a design decision from the Human",
		},
	}

	report, err := implementWithAdapter(root, l, wf, "SPEC-001", scripted)
	if err == nil {
		t.Fatalf("expected a BLOCKED workflow result to be reported as a non-nil error")
	}
	if report.Outcome != present.Blocked {
		t.Fatalf("expected Blocked outcome, got %v", report.Outcome)
	}
	rendered := present.Render(report, false)
	if !strings.Contains(rendered, "needs a design decision from the Human") {
		t.Fatalf("expected the unresolved reason to be reported: %s", rendered)
	}
}

func TestImplement_ProcessFailure(t *testing.T) {
	root := setupApprovedSpec(t)
	l, wf := mustPrepareImplementation(t, root, "SPEC-001")

	// No result is ever written, and the process itself reports a genuine, non-Gnomon-initiated
	// non-zero exit.
	failing := &adapter.FakeAdapter{ExitCode: 2}

	report, err := implementWithAdapter(root, l, wf, "SPEC-001", failing)
	if err == nil {
		t.Fatalf("expected a process failure with no result to be reported as an error")
	}
	if report.Outcome != present.Failed {
		t.Fatalf("expected Failed outcome, got %v", report.Outcome)
	}
	rendered := present.Render(report, false)
	if strings.Contains(rendered, "<nil>") {
		t.Fatalf("must never print a raw Go nil value: %s", rendered)
	}
}

func TestImplement_InvalidResult(t *testing.T) {
	root := setupApprovedSpec(t)
	l, wf := mustPrepareImplementation(t, root, "SPEC-001")

	malformed := &adapter.FakeAdapter{ResultContent: []byte("{not valid json")}

	report, err := implementWithAdapter(root, l, wf, "SPEC-001", malformed)
	if err == nil {
		t.Fatalf("expected a malformed result to be reported as an error")
	}
	if report.Outcome != present.Failed {
		t.Fatalf("expected Failed outcome for a malformed result, got %v", report.Outcome)
	}
}

func TestImplement_EmptyOptionalFieldsRenderNaturally(t *testing.T) {
	root := setupApprovedSpec(t)
	l, wf := mustPrepareImplementation(t, root, "SPEC-001")

	// Only the required "outcome" field is present — delivered, verification_evidence, and
	// remaining_unresolved are all absent, as a real minimal-but-valid result could be.
	scripted := &payloadAdapter{
		FakeAdapter: &adapter.FakeAdapter{},
		Payload:     map[string]interface{}{"outcome": "IMPLEMENTATION_COMPLETE"},
	}

	report, err := implementWithAdapter(root, l, wf, "SPEC-001", scripted)
	if err != nil {
		t.Fatalf("implement: %v (report: %+v)", err, report)
	}
	rendered := present.Render(report, true)
	if strings.Contains(rendered, "<nil>") {
		t.Fatalf("must never print a raw Go nil value for an absent optional field: %s", rendered)
	}
	// The generic renderer knows nothing about "remaining_unresolved" or "delivered" as field
	// names — an absent optional field of any kind is simply omitted entirely, with no per-field
	// placeholder text (that would require hardcoding the concept of "the unresolved field"),
	// exactly matching how a genuinely generic Contract-driven renderer must behave.
	if strings.Contains(rendered, "Unresolved") {
		t.Fatalf("an absent optional field must be omitted entirely, not rendered with any placeholder: %s", rendered)
	}
	if strings.Contains(rendered, "Delivered") {
		t.Fatalf("an absent Delivered field must be omitted entirely, not rendered empty: %s", rendered)
	}
	if len(report.Sections) != 0 {
		t.Fatalf("expected zero sections when every optional field is absent, got: %+v", report.Sections)
	}
}

// --- Agent provider resolution (agent.go) ---

func TestResolveAgent_ExplicitOverrideWinsAndNeverPersists(t *testing.T) {
	sandboxAgentConfig(t)

	ad, err := ResolveAgent("codex", nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := ad.(*adapter.CodexAdapter); !ok {
		t.Fatalf("expected *adapter.CodexAdapter, got %T", ad)
	}

	cfg, err := agentconfig.Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.DefaultProvider != "" {
		t.Fatalf("an explicit --agent override must never persist a new default, got %q", cfg.DefaultProvider)
	}
}

func TestResolveAgent_UsesStoredDefaultWithoutPrompting(t *testing.T) {
	sandboxAgentConfig(t)
	if err := agentconfig.Save(agentconfig.Config{DefaultProvider: "codex"}); err != nil {
		t.Fatal(err)
	}

	ad, err := ResolveAgent("", func(providers []string) (string, error) {
		t.Fatalf("chooser must not be called when a default is already stored")
		return "", nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := ad.(*adapter.CodexAdapter); !ok {
		t.Fatalf("expected the stored default (codex) to be used, got %T", ad)
	}
}

func TestResolveAgent_FirstUseChooserPersistsChoice(t *testing.T) {
	sandboxAgentConfig(t)

	ad, err := ResolveAgent("", func(providers []string) (string, error) {
		return "codex", nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := ad.(*adapter.CodexAdapter); !ok {
		t.Fatalf("expected the chosen provider (codex) to be resolved, got %T", ad)
	}

	cfg, err := agentconfig.Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.DefaultProvider != "codex" {
		t.Fatalf("expected the first-use choice to be persisted as the new default, got %q", cfg.DefaultProvider)
	}
}

func TestResolveAgent_NonInteractiveWithNoDefaultFailsClearly(t *testing.T) {
	sandboxAgentConfig(t)

	if _, err := ResolveAgent("", nil); err == nil {
		t.Fatalf("expected a clear refusal when no default is configured and no chooser is available")
	}

	cfg, err := agentconfig.Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.DefaultProvider != "" {
		t.Fatalf("a failed resolution must never persist anything, got %q", cfg.DefaultProvider)
	}
}

func TestResolveAgent_InvalidOverrideRefuses(t *testing.T) {
	sandboxAgentConfig(t)
	if _, err := ResolveAgent("chatgpt-5000", nil); err == nil {
		t.Fatalf("expected an unrecognized provider override to be refused")
	}
}

func TestResolveAgent_InvalidChooserChoiceNeverPersists(t *testing.T) {
	sandboxAgentConfig(t)

	_, err := ResolveAgent("", func(providers []string) (string, error) {
		return "not-a-real-provider", nil
	})
	if err == nil {
		t.Fatalf("expected an invalid chooser response to be refused")
	}

	cfg, loadErr := agentconfig.Load()
	if loadErr != nil {
		t.Fatal(loadErr)
	}
	if cfg.DefaultProvider != "" {
		t.Fatalf("an invalid choice must never be persisted, got %q", cfg.DefaultProvider)
	}
}

func TestResolveAgent_ClaudeResolution(t *testing.T) {
	sandboxAgentConfig(t)
	ad, err := ResolveAgent("claude", nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := ad.(*adapter.ClaudeAdapter); !ok {
		t.Fatalf("expected *adapter.ClaudeAdapter, got %T", ad)
	}
}

func TestResolveAgent_CodexResolution(t *testing.T) {
	sandboxAgentConfig(t)
	ad, err := ResolveAgent("codex", nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := ad.(*adapter.CodexAdapter); !ok {
		t.Fatalf("expected *adapter.CodexAdapter, got %T", ad)
	}
}

// --- gnomon agent / gnomon agent set-default ---

func TestAgentStatus_ReportsUnconfiguredState(t *testing.T) {
	sandboxAgentConfig(t)

	report, err := AgentStatus()
	if err != nil {
		t.Fatal(err)
	}
	if report.Outcome != present.Success {
		t.Fatalf("expected AgentStatus to succeed even when unconfigured")
	}
	rendered := present.Render(report, false)
	if !strings.Contains(rendered, "No default") {
		t.Fatalf("expected the unconfigured state to be stated plainly: %s", rendered)
	}
	if !strings.Contains(rendered, "claude") || !strings.Contains(rendered, "codex") {
		t.Fatalf("expected both available providers listed: %s", rendered)
	}
}

func TestAgentStatus_ReportsConfiguredDefault(t *testing.T) {
	sandboxAgentConfig(t)
	if err := agentconfig.Save(agentconfig.Config{DefaultProvider: "codex"}); err != nil {
		t.Fatal(err)
	}

	report, err := AgentStatus()
	if err != nil {
		t.Fatal(err)
	}
	rendered := present.Render(report, false)
	if !strings.Contains(rendered, "codex") {
		t.Fatalf("expected the configured default to be reported: %s", rendered)
	}
	if strings.Contains(rendered, "No default") {
		t.Fatalf("must not claim unconfigured once a default is stored: %s", rendered)
	}
}

func TestAgentSetDefault_PersistsAndReports(t *testing.T) {
	sandboxAgentConfig(t)

	report, err := AgentSetDefault("codex")
	if err != nil {
		t.Fatal(err)
	}
	if report.Outcome != present.Success {
		t.Fatalf("expected success, got %v", report.Outcome)
	}

	cfg, err := agentconfig.Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.DefaultProvider != "codex" {
		t.Fatalf("expected set-default to persist codex, got %q", cfg.DefaultProvider)
	}

	status, err := AgentStatus()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(present.Render(status, false), "codex") {
		t.Fatalf("expected gnomon agent to reflect the newly set default")
	}
}

func TestAgentSetDefault_RefusesInvalidProvider(t *testing.T) {
	sandboxAgentConfig(t)

	if _, err := AgentSetDefault("bogus"); err == nil {
		t.Fatalf("expected set-default to refuse an unrecognized provider")
	}

	cfg, err := agentconfig.Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.DefaultProvider != "" {
		t.Fatalf("an invalid set-default must never persist, got %q", cfg.DefaultProvider)
	}
}
