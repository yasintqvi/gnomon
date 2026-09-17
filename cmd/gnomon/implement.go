package main

import (
	"os"

	"github.com/spf13/cobra"

	"gnomon/internal/orchestrate"
)

var implementCmd = &cobra.Command{
	Use:     "implement <SPEC-id>",
	Short:   "Implement an Approved Specification",
	GroupID: groupEngineering,
	Args:    requireArgs("<SPEC-id>", cobra.ExactArgs(1)),
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := os.Getwd()
		if err != nil {
			return err
		}
		report, err := orchestrate.Implement(root, args[0], *implementAgentFlag, agentChooserForInvocation())
		return renderReport(report, err)
	},
}

var implementAgentFlag *string

func init() {
	implementAgentFlag = registerAgentFlag(implementCmd)
	rootCmd.AddCommand(implementCmd)
}
