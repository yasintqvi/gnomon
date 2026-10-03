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

This is the non-interactive way to invoke a workflow: scripting, CI, or a manual
one-off. For Specification work, the everyday interface is "gnomon spec" /
"gnomon spec <SPEC-id>" — an interactive browser and workspace that dispatches to
the exact same eligibility and execution engine "run" uses, so nothing behaves
differently between the two.

There is no dedicated command for implementation, testing, verification, review,
Specification Definition, or Specification Discovery — "run" is how each is
invoked outside the interactive workspace, by its own declared identity:

  gnomon run implementation SPEC-014
  gnomon run testing SPEC-014
  gnomon run verification              (what changed since the last passing verification)
  gnomon run verification --full       (every Approved Specification, checked independently)
  gnomon run verification src/auth/
  gnomon run review src/auth/
  gnomon run specification-definition SPEC-014
  gnomon run specification-discovery

describe/bootstrap keep their own dedicated commands (they are not Specification-
scoped), and a project's own custom workflows are reachable here the same way any
built-in one is.`,
	Example: `  gnomon run implementation SPEC-014
  gnomon run verification src/auth/
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
		opts := orchestrate.VerificationOptions{Full: *runFullFlag, Since: *runSinceFlag}
		report, findings, err := orchestrate.RunEvaluationWithOptions(root, args[0], target, *runAgentFlag, agentChooserForInvocation(), opts)
		// Offered only with a real interactive terminal — non-interactive invocations always fall
		// through to plain renderReport, never prompting or hanging.
		if len(findings) > 0 && stdinIsInteractive() {
			return runEvaluationResolution(root, args[0], target, report, findings, err)
		}
		return renderReport(report, err)
	},
}

var (
	runAgentFlag *string
	runFullFlag  *bool
	runSinceFlag *string
)

func init() {
	runAgentFlag = registerAgentFlag(runCmd)
	runFullFlag = runCmd.Flags().Bool("full", false, "verification only: check every criterion of every Approved Specification (or of the target) independently")
	runSinceFlag = runCmd.Flags().String("since", "", "verification only: check the changes since this Git ref instead of since the last passing verification")
	rootCmd.AddCommand(runCmd)
}
