package adapter

import (
	"strings"
	"testing"
)

func TestPrepare_InvokesFullyInteractiveMode(t *testing.T) {
	a := &ClaudeAdapter{&processAdapter{executable: "claude"}}
	if err := a.Prepare(Context{WorkflowIdentity: "implementation", ResultPath: "/tmp/x", RunID: "r"}); err != nil {
		t.Fatal(err)
	}
	args := a.cmd.Args
	if len(args) != 2 {
		t.Fatalf("expected exactly [executable, prompt] with no mode flags, got: %v", args)
	}
	if a.cmd.Stdin == nil {
		t.Fatalf("expected stdin to be inherited for a fully interactive session")
	}
}

func TestNewClaudeAdapter_DefaultsToClaudeOnPath(t *testing.T) {
	t.Setenv("GNOMON_CLAUDE_EXECUTABLE", "")
	a := NewClaudeAdapter()
	if a.executable != "claude" {
		t.Fatalf("expected default executable %q, got %q", "claude", a.executable)
	}
}

func TestNewClaudeAdapter_RespectsExecutableOverride(t *testing.T) {
	t.Setenv("GNOMON_CLAUDE_EXECUTABLE", "/custom/claude")
	a := NewClaudeAdapter()
	if a.executable != "/custom/claude" {
		t.Fatalf("expected override executable %q, got %q", "/custom/claude", a.executable)
	}
}

func TestNewClaudeAdapter_UnaffectedByCodexOverride(t *testing.T) {
	t.Setenv("GNOMON_CLAUDE_EXECUTABLE", "")
	t.Setenv("GNOMON_CODEX_EXECUTABLE", "/custom/codex")
	a := NewClaudeAdapter()
	if a.executable != "claude" {
		t.Fatalf("Claude's executable resolution must be independent of Codex's override, got %q", a.executable)
	}
}

func TestBuildPrompt_NeverInstructsTheAgentToExit(t *testing.T) {
	ctx := Context{
		WorkflowIdentity: "implementation",
		WorkflowPath:     ".gnomon/workflows/implementation.md",
		SpecIdentity:     "SPEC-001",
		ResultPath:       "/tmp/proj/.gnomon-runtime/abc123.result.json",
		RunID:            "abc123",
	}
	prompt := buildPrompt(ctx)

	mustContain := []string{
		"publish exactly one valid JSON result",
		ctx.ResultPath,
		"Gnomon is watching for that file",
		"you do not need to exit or end the conversation yourself",
	}
	for _, s := range mustContain {
		if !strings.Contains(prompt, s) {
			t.Fatalf("expected prompt to contain %q, got:\n%s", s, prompt)
		}
	}
	mustNotContain := []string{"end this Claude Code session", "/exit", "non-interactive"}
	for _, s := range mustNotContain {
		if strings.Contains(prompt, s) {
			t.Fatalf("did not expect the prompt to contain %q — the Agent must never be told to end its own session, got:\n%s", s, prompt)
		}
	}
}

func TestBuildPrompt_OmitsSpecLineWhenNoneSupplied(t *testing.T) {
	ctx := Context{WorkflowIdentity: "bootstrap", WorkflowPath: "x", ResultPath: "y", RunID: "z"}
	prompt := buildPrompt(ctx)
	if strings.Contains(prompt, "governing Specification") {
		t.Fatalf("did not expect a Specification line when SpecIdentity is empty")
	}
	if strings.Contains(prompt, "The target for this run is") {
		t.Fatalf("did not expect a target line when Target is also empty")
	}
}

// TestBuildPrompt_UsesGenericTargetWordingNotSpecificationWording proves a free-form Target (as
// Verify/Review supply) is never described to the Agent as a Specification — Core's own text is
// explicit that such a target is "not necessarily a Specification at all."
func TestBuildPrompt_UsesGenericTargetWordingNotSpecificationWording(t *testing.T) {
	ctx := Context{WorkflowIdentity: "verification", WorkflowPath: "x", ResultPath: "y", RunID: "z", Target: "src/auth/"}
	prompt := buildPrompt(ctx)
	if !strings.Contains(prompt, "The target for this run is: src/auth/") {
		t.Fatalf("expected the generic target line, got:\n%s", prompt)
	}
	if strings.Contains(prompt, "governing Specification") {
		t.Fatalf("a free-form Target must never be described as a Specification, got:\n%s", prompt)
	}
}

// TestBuildPrompt_SpecIdentityTakesPrecedenceOverTarget confirms the two fields are mutually
// exclusive in practice — SpecIdentity, when set, is what gets described, matching how callers
// only ever populate one of the two.
func TestBuildPrompt_SpecIdentityTakesPrecedenceOverTarget(t *testing.T) {
	ctx := Context{WorkflowIdentity: "implementation", WorkflowPath: "x", ResultPath: "y", RunID: "z", SpecIdentity: "SPEC-001", Target: "should not appear"}
	prompt := buildPrompt(ctx)
	if !strings.Contains(prompt, "governing Specification is: SPEC-001") {
		t.Fatalf("expected the Specification line, got:\n%s", prompt)
	}
	if strings.Contains(prompt, "should not appear") {
		t.Fatalf("did not expect Target to also render when SpecIdentity is set, got:\n%s", prompt)
	}
}

// TestBuildPrompt_HandoffCarriesTheFindingThatCausedThisRun proves a resolution Agent invocation
// actually receives the finding that caused it to be launched — the mechanism the interactive
// finding-resolution loop relies on so a fresh Agent process never loses that context.
func TestBuildPrompt_HandoffCarriesTheFindingThatCausedThisRun(t *testing.T) {
	ctx := Context{
		WorkflowIdentity: "implementation", WorkflowPath: "x", ResultPath: "y", RunID: "z",
		SpecIdentity: "SPEC-014",
		Handoff: &ResolutionHandoff{
			OriginWorkflow: "review", OriginTarget: "src/billing/",
			FindingID: "F-001", Classification: "DEFECT",
			Summary: "Refund does not release inventory reservation.", Evidence: "Reservation remains active after refund completion.",
		},
	}
	prompt := buildPrompt(ctx)
	for _, want := range []string{
		"earlier review of src/billing/",
		"Finding: F-001 — DEFECT",
		"Refund does not release inventory reservation.",
		"Reservation remains active after refund completion.",
		"following the implementation workflow's own process",
		"governing Specification is: SPEC-014",
	} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("expected prompt to contain %q, got:\n%s", want, prompt)
		}
	}
}

func TestBuildPrompt_NilHandoff_NoHandoffText(t *testing.T) {
	ctx := Context{WorkflowIdentity: "implementation", WorkflowPath: "x", ResultPath: "y", RunID: "z", SpecIdentity: "SPEC-014"}
	prompt := buildPrompt(ctx)
	if strings.Contains(prompt, "launched to resolve a finding") {
		t.Fatalf("did not expect handoff framing when Handoff is nil, got:\n%s", prompt)
	}
}

// TestBuildPrompt_HandoffTellsTheAgentToAskTheHumanDirectly proves the resolution Agent is told,
// in-prompt, to ask the Human directly for any material decision it needs — the mechanism Model B
// relies on instead of the CLI pre-collecting a Human decision before dispatch.
func TestBuildPrompt_HandoffTellsTheAgentToAskTheHumanDirectly(t *testing.T) {
	ctx := Context{
		WorkflowIdentity: "knowledge-resolution", WorkflowPath: "x", ResultPath: "y", RunID: "z",
		Handoff: &ResolutionHandoff{
			OriginWorkflow: "review", OriginTarget: "src/billing/",
			FindingID: "F-002", Classification: "RISK",
			Summary: "Retried without idempotency key.",
		},
	}
	prompt := buildPrompt(ctx)
	if !strings.Contains(prompt, "ask the Human directly in this session") {
		t.Fatalf("expected the prompt to direct the Agent to ask the Human directly, got:\n%s", prompt)
	}
	if strings.Contains(prompt, "Human Decision:") {
		t.Fatalf("did not expect a CLI-collected Human Decision line — the Agent asks in-session, got:\n%s", prompt)
	}
}
