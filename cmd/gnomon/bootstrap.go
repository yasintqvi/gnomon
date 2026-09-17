package main

import (
	"os"

	"github.com/spf13/cobra"

	"gnomon/internal/orchestrate"
)

var bootstrapCmd = &cobra.Command{
	Use:     "bootstrap",
	Short:   "Establish the minimum verified project baseline (Bootstrap)",
	GroupID: groupProject,
	Args:    requireArgs("no arguments", cobra.NoArgs),
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := os.Getwd()
		if err != nil {
			return err
		}
		report, err := orchestrate.Bootstrap(root, *bootstrapAgentFlag, agentChooserForInvocation())
		return renderReport(report, err)
	},
}

var bootstrapAgentFlag *string

func init() {
	bootstrapAgentFlag = registerAgentFlag(bootstrapCmd)
	rootCmd.AddCommand(bootstrapCmd)
}
