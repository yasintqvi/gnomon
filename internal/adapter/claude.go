package adapter

import "os"

// claudeExecutableEnvVar overrides Claude Code's executable path/name when it isn't on PATH as
// "claude", or a different binary is wanted — provider-specific, since Codex has its own,
// independent override (codex.go).
const claudeExecutableEnvVar = "GNOMON_CLAUDE_EXECUTABLE"

// ClaudeAdapter is Terminal Handoff to a real, fully interactive Claude Code subprocess
// (cli/AGENT_ADAPTER.md). Mechanics are shared via processAdapter; this type exists only to be a
// distinct, resolvable provider identity (resolve.go).
type ClaudeAdapter struct {
	*processAdapter
}

// NewClaudeAdapter constructs the Claude Code Adapter: "claude" resolved from PATH by default,
// overridable via GNOMON_CLAUDE_EXECUTABLE.
func NewClaudeAdapter() *ClaudeAdapter {
	exe := os.Getenv(claudeExecutableEnvVar)
	if exe == "" {
		exe = "claude"
	}
	return &ClaudeAdapter{&processAdapter{executable: exe}}
}
