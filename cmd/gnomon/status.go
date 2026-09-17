package main

import (
	"os"

	"github.com/spf13/cobra"

	"gnomon/internal/orchestrate"
)

var statusCmd = &cobra.Command{
	Use:     "status",
	Short:   "Show the current derived project state",
	GroupID: groupGuidance,
	Long: `Show the current derived project state.

Answers: what is true about my project right now? Initialization and contract
compatibility, every Specification and its Draft/Approved state, and whether the
working tree has uncommitted changes — a snapshot, recomputed fresh every time,
never cached.

status is not validate: it never audits Contract/Result-Contract structural
integrity. Use "gnomon validate" for that.`,
	Args: requireArgs("no arguments", cobra.NoArgs),
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := os.Getwd()
		if err != nil {
			return err
		}
		report, err := orchestrate.Status(root)
		return renderReport(report, err)
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
