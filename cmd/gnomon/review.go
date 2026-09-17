package main

import (
	"os"

	"github.com/spf13/cobra"

	"gnomon/internal/orchestrate"
)

var reviewCmd = &cobra.Command{
	Use:     "review [target]",
	Short:   "Apply engineering judgment to an artifact within an authorized scope (Review)",
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
		report, err := orchestrate.Review(root, target, *reviewAgentFlag, agentChooserForInvocation())
		return renderReport(report, err)
	},
}

var reviewAgentFlag *string

func init() {
	reviewAgentFlag = registerAgentFlag(reviewCmd)
	rootCmd.AddCommand(reviewCmd)
}
