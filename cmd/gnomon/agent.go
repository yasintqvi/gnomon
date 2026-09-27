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

// registerAgentFlag adds the shared --agent override, valid for that single invocation only —
// never modifies the persisted default. Deterministic commands never touch an Adapter, so they
// never call this.
func registerAgentFlag(cmd *cobra.Command) *string {
	var val string
	cmd.Flags().StringVar(&val, "agent", "", "override the Agent provider for this invocation only (claude|codex); never changes the stored default")
	return &val
}

// stdinIsInteractive reports whether stdin is a real terminal — the CLI edge's own decision;
// orchestrate never performs it. Uses term.IsTerminal, not a bare os.ModeCharDevice check: /dev/null
// is also a character device, so that simpler check would misclassify `< /dev/null` as interactive.
func stdinIsInteractive() bool {
	return term.IsTerminal(int(os.Stdin.Fd()))
}

// agentChooserForInvocation returns a real interactive chooser only when stdin is a terminal, nil
// otherwise — so a non-interactive invocation with no stored default fails clearly instead of
// hanging on stdin.
func agentChooserForInvocation() orchestrate.AgentChooser {
	if !stdinIsInteractive() {
		return nil
	}
	return interactiveAgentChooser
}

// providerLabels maps each canonical provider identity to its Human-facing select-menu label —
// presentation-only; an identity with no entry here falls back to itself.
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

// interactiveAgentChooser is the real AgentChooser: an arrow-key select menu. The canonical value
// is returned, never the display label.
func interactiveAgentChooser(providers []string) (string, error) {
	items := make([]selectItem, len(providers))
	for i, p := range providers {
		items[i] = selectItem{value: p, label: providerLabel(p)}
	}
	return runSelectMenu("Choose your default Agent:", items)
}
