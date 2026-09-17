package adapter

import (
	"fmt"
	"strings"
)

// providerNames lists every currently supported Agent provider, in the order shown to the Human
// when choosing or listing (gnomon agent, and the first-use chooser prompt).
var providerNames = []string{"claude", "codex"}

// Providers returns the currently supported Agent provider names. The returned slice is an
// independent copy — callers may not mutate providerNames through it.
func Providers() []string {
	out := make([]string, len(providerNames))
	copy(out, providerNames)
	return out
}

// Resolve constructs the concrete Adapter for a named provider. This is the one place any caller
// is allowed to learn a concrete Adapter type exists — orchestration calls this and never
// constructs ClaudeAdapter or CodexAdapter directly.
func Resolve(provider string) (Adapter, error) {
	switch provider {
	case "claude":
		return NewClaudeAdapter(), nil
	case "codex":
		return NewCodexAdapter(), nil
	default:
		return nil, fmt.Errorf("unknown Agent provider %q (valid: %s)", provider, strings.Join(providerNames, ", "))
	}
}
