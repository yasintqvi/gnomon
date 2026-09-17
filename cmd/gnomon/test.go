package main

import (
	"os"

	"github.com/spf13/cobra"

	"gnomon/internal/orchestrate"
)

var testCmd = &cobra.Command{
	Use:     "test [SPEC-id]",
	Short:   "Design, implement, execute, and evaluate tests (Testing)",
	GroupID: groupEngineering,
	Args:    requireArgs("at most one argument: [SPEC-id]", cobra.MaximumNArgs(1)),
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := os.Getwd()
		if err != nil {
			return err
		}
		var specID string
		if len(args) == 1 {
			specID = args[0]
		}
		report, err := orchestrate.Test(root, specID, *testAgentFlag, agentChooserForInvocation())
		return renderReport(report, err)
	},
}

var testAgentFlag *string

func init() {
	testAgentFlag = registerAgentFlag(testCmd)
	rootCmd.AddCommand(testCmd)
}
