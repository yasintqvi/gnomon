package adapter

import "os"

// codexExecutableEnvVar overrides Codex CLI's executable path/name when it isn't on PATH as
// "codex", or a different binary is wanted — provider-specific, since Claude has its own,
// independent override (claude.go).
const codexExecutableEnvVar = "GNOMON_CODEX_EXECUTABLE"

// CodexAdapter is the second v1 Adapter implementation. Its invocation shape was confirmed
// directly against a real, installed Codex CLI (`codex --help`, codex-cli 0.154.0) before being
// written, not assumed from Claude's own shape:
//
//	Usage: codex [OPTIONS] [PROMPT]
//	  [PROMPT]  Optional user prompt to start the session
//	  "If no subcommand is specified, options will be forwarded to the interactive CLI."
//
// This is the same "bare executable plus one positional prompt argument" shape Claude Code uses
// — no `exec`/`review` subcommand (Codex's own non-interactive modes, which would break the
// required interactive UX exactly as Claude's now-rejected `-p` mode would have), and no
// sandbox/approval-bypass flags (`--sandbox`, `--ask-for-approval`, `--approve-for-me`,
// `--dangerously-bypass-approvals-and-sandbox`) — leaving Codex's own native interactive
// approval UI fully intact and visible to the Human, exactly as Claude's own permission prompts
// remain untouched. All mechanics are otherwise shared via processAdapter.
type CodexAdapter struct {
	*processAdapter
}

// NewCodexAdapter constructs the Codex Adapter: "codex" resolved from PATH by default,
// overridable via GNOMON_CODEX_EXECUTABLE.
func NewCodexAdapter() *CodexAdapter {
	exe := os.Getenv(codexExecutableEnvVar)
	if exe == "" {
		exe = "codex"
	}
	return &CodexAdapter{&processAdapter{executable: exe}}
}
