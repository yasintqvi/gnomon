package main

import (
	"os"

	"github.com/spf13/cobra"

	"gnomon/internal/orchestrate"
)

var runCmd = &cobra.Command{
	Use:     "run <workflow-identity> [target]",
	Short:   "Run any workflow by its declared Contract identity (advanced)",
	GroupID: groupAdvanced,
	Long: `Run any workflow by its declared Contract identity — built-in or custom.

Most engineering work has a dedicated command (implement, test, verify, review,
bootstrap, describe, spec define) and should use it. run exists for workflows
without one: a project's own custom, Contract-driven workflows, and the rare
manual invocation of a built-in workflow by name. It uses exactly the same
eligibility and execution engine every dedicated command already does.`,
	Example: `  gnomon run verification src/auth/
  gnomon run my-custom-workflow SPEC-014`,
	Args: requireArgs("<workflow-identity> [target]", cobra.RangeArgs(1, 2)),
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := os.Getwd()
		if err != nil {
			return err
		}
		var target string
		if len(args) == 2 {
			target = args[1]
		}
		report, err := orchestrate.Run(root, args[0], target, *runAgentFlag, agentChooserForInvocation())
		return renderReport(report, err)
	},
}

var runAgentFlag *string

func init() {
	runAgentFlag = registerAgentFlag(runCmd)
	rootCmd.AddCommand(runCmd)
}
