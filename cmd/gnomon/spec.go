package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"gnomon/internal/orchestrate"
)

var specCmd = &cobra.Command{
	Use:     "spec [SPEC-id]",
	Short:   "Browse and manage Specifications",
	GroupID: groupSpecification,
	Long: `Browse and manage Specifications.

"gnomon spec" opens an interactive browser of every Specification — its lifecycle,
approval state, and the actions currently valid for it. "gnomon spec SPEC-001" opens
that one Specification directly. Both require a real interactive terminal.

Every lifecycle action — Define, Approve, Revoke, Implement, Test — is available as a
contextual action from there; only the ones currently valid for that Specification's
state can actually be run, and the rest are shown with why not. This is the normal way
to work with Specifications; there is no separate set of commands to memorize.

"gnomon spec create <title>" below remains available directly for scripting,
automation, and CI, where an interactive browser cannot run — it is the one
Specification-related operation with no Workflow Contract for "gnomon run" to invoke.
Every other lifecycle action has a non-interactive equivalent through
"gnomon run <workflow-identity> <SPEC-id>" and, for approval itself,
"gnomon approve"/"gnomon revoke".`,
	Example: `  gnomon spec
  gnomon spec SPEC-001`,
	Args: requireArgs("at most one argument: [SPEC-id]", cobra.MaximumNArgs(1)),
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := os.Getwd()
		if err != nil {
			return err
		}
		if !stdinIsInteractive() {
			return fmt.Errorf("gnomon spec requires an interactive terminal; for automation, use `gnomon spec create <title>`, `gnomon approve <SPEC-id>`, `gnomon revoke <SPEC-id>`, or `gnomon run <workflow-identity> [target]`")
		}
		if len(args) == 1 {
			return runSpecWorkspace(root, args[0])
		}
		return runSpecBrowser(root)
	},
}

var specCreateCmd = &cobra.Command{
	Use:   "create <title>",
	Short: "Create a new Draft Specification from the template",
	Long: `Create a new Draft Specification, without running the Agent.

To start a feature and define it in one step, use gnomon define "<request>" instead.

The default template is short: Goal, Decisions (each question with the user's answer),
Acceptance Criteria, and Out of Scope. --detailed uses the full use-case template
(actors, flows, business rules, inputs/outputs, dependencies) instead.`,
	Args: requireArgs("<title>", cobra.ExactArgs(1)),
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := os.Getwd()
		if err != nil {
			return err
		}
		report, err := orchestrate.SpecCreateFrom(root, args[0], specCreateDetailed)
		return renderReport(report, err)
	},
}

var specCreateDetailed bool

func init() {
	specCreateCmd.Flags().BoolVar(&specCreateDetailed, "detailed", false, "use the detailed use-case template instead of the short one")
	specCmd.AddCommand(specCreateCmd)
	rootCmd.AddCommand(specCmd)
}
