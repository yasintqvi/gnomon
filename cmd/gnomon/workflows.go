package main

import (
	"os"

	"github.com/spf13/cobra"

	"gnomon/internal/orchestrate"
)

var workflowsCmd = &cobra.Command{
	Use:     "workflows",
	Short:   "List the workflows in effect and where each comes from",
	GroupID: groupAdvanced,
	Long: `List the workflows in effect and where each comes from.

Workflows are read from the Gnomon binary, so they stay current when Gnomon is updated.
A file in .gnomon/workflows/ replaces the bundled workflow of the same name once it
differs from every version Gnomon has shipped; an unmodified copy left by an older
"gnomon init" is ignored. Additional project workflows in that directory are run with
"gnomon run <identity>".`,
	Args: requireArgs("no arguments", cobra.NoArgs),
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := os.Getwd()
		if err != nil {
			return err
		}
		report, err := orchestrate.Workflows(root)
		return renderReport(report, err)
	},
}

var workflowsCustomizeCmd = &cobra.Command{
	Use:   "customize <file>",
	Short: "Copy a bundled workflow into .gnomon/workflows/ to change it for this project",
	Args:  requireArgs("<file>", cobra.ExactArgs(1)),
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := os.Getwd()
		if err != nil {
			return err
		}
		report, err := orchestrate.CustomizeWorkflow(root, args[0])
		return renderReport(report, err)
	},
}

func init() {
	workflowsCmd.AddCommand(workflowsCustomizeCmd)
	rootCmd.AddCommand(workflowsCmd)
}
