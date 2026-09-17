package main

import (
	"os"

	"github.com/spf13/cobra"

	"gnomon/internal/orchestrate"
)

var describeCmd = &cobra.Command{
	Use:     "describe",
	Short:   "Establish or update initial project knowledge (Initial Knowledge Establishment)",
	GroupID: groupProject,
	Args:    requireArgs("no arguments", cobra.NoArgs),
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := os.Getwd()
		if err != nil {
			return err
		}
		report, err := orchestrate.Describe(root, *describeAgentFlag, agentChooserForInvocation())
		return renderReport(report, err)
	},
}

var describeAgentFlag *string

func init() {
	describeAgentFlag = registerAgentFlag(describeCmd)
	rootCmd.AddCommand(describeCmd)
}
