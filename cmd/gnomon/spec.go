package main

import (
	"os"

	"github.com/spf13/cobra"

	"gnomon/internal/orchestrate"
)

var specCmd = &cobra.Command{
	Use:     "spec",
	Short:   "Create, discover, and define Specifications",
	GroupID: groupSpecification,
	Long: `Create, discover, and define Specifications.

A Specification is a Draft until a Human explicitly approves it (gnomon approve) —
Implementation and Testing both require Approved. "spec discover" and "spec create"
are two ways to get a new Draft; "spec define" brings an existing Draft to a
reviewable, approval-ready state.`,
	Example: `  gnomon spec create "Password Reset"
  gnomon spec discover
  gnomon spec define SPEC-001`,
}

var specCreateCmd = &cobra.Command{
	Use:   "create <title>",
	Short: "Create a new Draft Specification from the template",
	Args:  requireArgs("<title>", cobra.ExactArgs(1)),
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := os.Getwd()
		if err != nil {
			return err
		}
		report, err := orchestrate.SpecCreate(root, args[0])
		return renderReport(report, err)
	},
}

var specDiscoverCmd = &cobra.Command{
	Use:   "discover",
	Short: "Propose the next Specification candidate from current project knowledge (Specification Discovery)",
	Args:  requireArgs("no arguments", cobra.NoArgs),
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := os.Getwd()
		if err != nil {
			return err
		}
		report, err := orchestrate.SpecDiscover(root, *specDiscoverAgentFlag, agentChooserForInvocation())
		return renderReport(report, err)
	},
}

var specDefineCmd = &cobra.Command{
	Use:   "define <SPEC-id>",
	Short: "Bring a Draft Specification to READY_FOR_APPROVAL (Specification Definition)",
	Args:  requireArgs("<SPEC-id>", cobra.ExactArgs(1)),
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := os.Getwd()
		if err != nil {
			return err
		}
		report, err := orchestrate.SpecDefine(root, args[0], *specDefineAgentFlag, agentChooserForInvocation())
		return renderReport(report, err)
	},
}

var specDiscoverAgentFlag *string
var specDefineAgentFlag *string

func init() {
	specDiscoverAgentFlag = registerAgentFlag(specDiscoverCmd)
	specDefineAgentFlag = registerAgentFlag(specDefineCmd)

	specCmd.AddCommand(specCreateCmd)
	specCmd.AddCommand(specDiscoverCmd)
	specCmd.AddCommand(specDefineCmd)
	rootCmd.AddCommand(specCmd)
}
