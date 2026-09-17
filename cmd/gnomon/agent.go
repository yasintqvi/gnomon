package main

import (
	"os"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"gnomon/internal/orchestrate"
)

var agentCmd = &cobra.Command{
	Use:     "agent",
	Short:   "Show or change the default Agent provider",
	GroupID: groupAdvanced,
	Args:    requireArgs("no arguments", cobra.NoArgs),
	RunE: func(cmd *cobra.Command, args []string) error {
		report, err := orchestrate.AgentStatus()
		return renderReport(report, err)
	},
}

var agentSetDefaultCmd = &cobra.Command{
	Use:   "set-default <claude|codex>",
	Short: "Change the persistent default Agent provider",
	Args:  requireArgs("<claude|codex>", cobra.ExactArgs(1)),
	RunE: func(cmd *cobra.Command, args []string) error {
		report, err := orchestrate.AgentSetDefault(args[0])
		return renderReport(report, err)
	},
}

func init() {
	agentCmd.AddCommand(agentSetDefaultCmd)
	rootCmd.AddCommand(agentCmd)
}

// registerAgentFlag adds the shared --agent override to an Agent-invoking command, returning the
// bound value. It overrides everything for that single invocation only and never modifies the
// persisted default. Deterministic commands (init, spec create, approve, revoke, agent, agent
// set-default) must never call this — they never touch an Adapter at all.
func registerAgentFlag(cmd *cobra.Command) *string {
	var val string
	cmd.Flags().StringVar(&val, "agent", "", "override the Agent provider for this invocation only (claude|codex); never changes the stored default")
	return &val
}

// stdinIsInteractive reports whether stdin is a real terminal. This is the CLI edge's own
// decision; orchestrate never performs it, so a chooser it receives is either real or nil, never
// something it has to reason about further.
//
// This uses term.IsTerminal rather than a bare os.ModeCharDevice check on os.Stdin.Stat(): /dev/
// null is also a character device, so the simpler check misclassified `< /dev/null` (a common
// non-interactive shape, including in CI) as interactive — the raw-mode select menu would then
// be entered against a stdin that cannot supply it, failing with a confusing low-level "operation
// not supported by device" error instead of ResolveAgent's own clear, deterministic refusal
// message. term.IsTerminal (already a dependency, already used by select.go for the same reason)
// does not have this false positive.
func stdinIsInteractive() bool {
	return term.IsTerminal(int(os.Stdin.Fd()))
}

// agentChooserForInvocation returns a real interactive chooser only when stdin is a terminal,
// and nil otherwise — keeping the interactivity decision at the CLI edge, never inside
// orchestrate. A nil chooser tells orchestrate.ResolveAgent that interaction is unavailable, so a
// non-interactive invocation with no stored default fails clearly instead of hanging on stdin.
func agentChooserForInvocation() orchestrate.AgentChooser {
	if !stdinIsInteractive() {
		return nil
	}
	return interactiveAgentChooser
}

// providerLabels maps each canonical provider identity (what --agent, gnomon agent set-default,
// and agentconfig all use) to the Human-facing label shown in the select menu. This mapping is
// presentation-only and local to the CLI edge — adapter.Providers() remains the sole authority on
// which identities are valid; an identity with no entry here just falls back to itself.
var providerLabels = map[string]string{
	"claude": "Claude Code",
	"codex":  "Codex",
}

func providerLabel(id string) string {
	if l, ok := providerLabels[id]; ok {
		return l
	}
	return id
}

// interactiveAgentChooser is the real AgentChooser wired in at the CLI edge: an arrow-key select
// menu, not a typed prompt. The Human never types "claude" or "codex" — they navigate with ↑/↓
// and press Enter, and the item's own canonical value (never its display label) is what's
// returned, so it resolves and persists exactly as any other chooser response always has.
func interactiveAgentChooser(providers []string) (string, error) {
	items := make([]selectItem, len(providers))
	for i, p := range providers {
		items[i] = selectItem{value: p, label: providerLabel(p)}
	}
	return runSelectMenu("Choose your default Agent:", items)
}
