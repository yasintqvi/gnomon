package main

import (
	"os"

	"github.com/spf13/cobra"

	"gnomon/internal/orchestrate"
)

var finalizeCmd = &cobra.Command{
	Use:     "finalize",
	Short:   "Hand off approved, validated work to Git (Git Finalization)",
	GroupID: groupEngineering,
	Args:    requireArgs("no arguments", cobra.NoArgs),
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := os.Getwd()
		if err != nil {
			return err
		}
		report, err := orchestrate.Finalize(root, *finalizeAgentFlag, agentChooserForInvocation())
		return renderReport(report, err)
	},
}

var finalizeAgentFlag *string

func init() {
	finalizeAgentFlag = registerAgentFlag(finalizeCmd)
	rootCmd.AddCommand(finalizeCmd)
}
