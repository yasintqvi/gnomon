package main

import (
	"os"

	"github.com/spf13/cobra"

	"gnomon/internal/orchestrate"
)

var verifyCmd = &cobra.Command{
	Use:     "verify [target]",
	Short:   "Objectively check verifiable obligations against evidence (Verification)",
	GroupID: groupEngineering,
	Args:    requireArgs("at most one argument: [target]", cobra.MaximumNArgs(1)),
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := os.Getwd()
		if err != nil {
			return err
		}
		var target string
		if len(args) == 1 {
			target = args[0]
		}
		report, err := orchestrate.Verify(root, target, *verifyAgentFlag, agentChooserForInvocation())
		return renderReport(report, err)
	},
}

var verifyAgentFlag *string

func init() {
	verifyAgentFlag = registerAgentFlag(verifyCmd)
	rootCmd.AddCommand(verifyCmd)
}
