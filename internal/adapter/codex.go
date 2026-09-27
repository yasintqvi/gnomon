package adapter

import "os"

// codexExecutableEnvVar overrides Codex CLI's executable path/name when it isn't on PATH as
// "codex", or a different binary is wanted — provider-specific, since Claude has its own,
// independent override (claude.go).
const codexExecutableEnvVar = "GNOMON_CODEX_EXECUTABLE"

// CodexAdapter's invocation shape was confirmed against a real, installed Codex CLI (`codex
// --help`, codex-cli 0.154.0): bare executable plus one positional prompt argument, no `exec`/
// `review` subcommand and no sandbox/approval-bypass flags — Codex's own native interactive
// approval UI stays fully intact and visible to the Human. Mechanics otherwise shared via processAdapter.
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
