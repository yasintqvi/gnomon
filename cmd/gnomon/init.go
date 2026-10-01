package main

import (
	"os"

	"github.com/spf13/cobra"

	"gnomon/internal/orchestrate"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a Gnomon project in the current directory",
	Long: `Initialize a Gnomon project in the current directory.

Creates a minimal .gnomon/ directory: specifications/, approvals/, and CONTRACT_VERSION.
Workflows and Specification templates are read from the Gnomon binary, so they stay
current when Gnomon is updated. Re-running init never overwrites existing files.

--full also creates the optional templates older versions created by default: project
context and design documents, contracts, evaluation criteria, and an ADR template.
Unfilled templates are never handed to the Agent as project knowledge.`,
	GroupID: groupProject,
	Args:    requireArgs("no arguments", cobra.NoArgs),
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := os.Getwd()
		if err != nil {
			return err
		}
		initialize := orchestrate.Init
		if initFull {
			initialize = orchestrate.InitFull
		}
		report, err := initialize(root)
		return renderReport(report, err)
	},
}

var initFull bool

func init() {
	initCmd.Flags().BoolVar(&initFull, "full", false, "also create the optional knowledge, design, contract, and ADR templates")
	rootCmd.AddCommand(initCmd)
}
