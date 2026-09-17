package adapter

import "testing"

// TestCodexPrepare_InvokesFullyInteractiveMode confirms CodexAdapter launches exactly
// [executable, prompt] with no subcommand and no sandbox/approval flags, with stdio inherited —
// the real behavior of a currently-installed Codex CLI (codex-cli 0.154.0), confirmed via
// `codex --help` before this was written, not assumed from Claude's own shape. See codex.go's own
// doc comment for the exact inspected usage line this mirrors.
func TestCodexPrepare_InvokesFullyInteractiveMode(t *testing.T) {
	a := NewCodexAdapter()
	if err := a.Prepare(Context{WorkflowIdentity: "implementation", ResultPath: "/tmp/x", RunID: "r"}); err != nil {
		t.Fatal(err)
	}
	args := a.cmd.Args
	if len(args) != 2 {
		t.Fatalf("expected exactly [executable, prompt] with no subcommand/sandbox/approval flags, got: %v", args)
	}
	if a.cmd.Args[0] != a.executable {
		t.Fatalf("expected argv[0] to be the resolved executable, got %q", a.cmd.Args[0])
	}
	if a.cmd.Stdin == nil {
		t.Fatalf("expected stdin to be inherited for a fully interactive session")
	}
}

func TestNewCodexAdapter_DefaultsToCodexOnPath(t *testing.T) {
	t.Setenv("GNOMON_CODEX_EXECUTABLE", "")
	a := NewCodexAdapter()
	if a.executable != "codex" {
		t.Fatalf("expected default executable %q, got %q", "codex", a.executable)
	}
}

func TestNewCodexAdapter_RespectsExecutableOverride(t *testing.T) {
	t.Setenv("GNOMON_CODEX_EXECUTABLE", "/custom/codex")
	a := NewCodexAdapter()
	if a.executable != "/custom/codex" {
		t.Fatalf("expected override executable %q, got %q", "/custom/codex", a.executable)
	}
}

func TestNewCodexAdapter_UnaffectedByClaudeOverride(t *testing.T) {
	t.Setenv("GNOMON_CODEX_EXECUTABLE", "")
	t.Setenv("GNOMON_CLAUDE_EXECUTABLE", "/custom/claude")
	a := NewCodexAdapter()
	if a.executable != "codex" {
		t.Fatalf("Codex's executable resolution must be independent of Claude's override, got %q", a.executable)
	}
}

// TestCodexAdapter_UsesSameProviderNeutralPromptAsClaudeAdapter proves both Adapters receive
// byte-identical instructions for an identical Context — buildPrompt never names a provider, so
// neither Adapter needs (or gets) its own variant.
func TestCodexAdapter_UsesSameProviderNeutralPromptAsClaudeAdapter(t *testing.T) {
	ctx := Context{WorkflowIdentity: "implementation", WorkflowPath: "x", ResultPath: "y", RunID: "z", SpecIdentity: "SPEC-001"}

	codex := NewCodexAdapter()
	if err := codex.Prepare(ctx); err != nil {
		t.Fatal(err)
	}
	claude := NewClaudeAdapter()
	if err := claude.Prepare(ctx); err != nil {
		t.Fatal(err)
	}
	if codex.cmd.Args[1] != claude.cmd.Args[1] {
		t.Fatalf("expected an identical provider-neutral prompt for both Adapters given an identical Context")
	}
}
