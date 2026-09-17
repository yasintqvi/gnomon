package main

import (
	"os"

	"github.com/spf13/cobra"

	"gnomon/internal/orchestrate"
)

var nextCmd = &cobra.Command{
	Use:     "next",
	Short:   "Show valid next actions from current project state",
	GroupID: groupGuidance,
	Long: `Show valid next actions from current project state.

Answers: what actions are currently valid? For every Specification, the commands
its current Draft/Approved state makes valid — never a single "recommended" pick,
and never a claim about what previously ran, since Gnomon persists no execution
history. When several Specifications are eligible for the same action, all are
listed; none is presented as more important than another.

next is not an Agent, a planner, or a recommendation engine — only a deterministic
report of what current repository state already makes valid.`,
	Args: requireArgs("no arguments", cobra.NoArgs),
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := os.Getwd()
		if err != nil {
			return err
		}
		report, err := orchestrate.Next(root)
		return renderReport(report, err)
	},
}

func init() {
	rootCmd.AddCommand(nextCmd)
}
