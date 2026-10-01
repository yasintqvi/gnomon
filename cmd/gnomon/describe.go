package main

import (
	"os"

	"github.com/spf13/cobra"

	"gnomon/internal/orchestrate"
)

var describeCmd = &cobra.Command{
	Use:     "describe",
	Short:   "Optionally record project facts the code doesn't show",
	GroupID: groupProject,
	Long: `Optionally record project facts the code doesn't show.

Hands the project to the Agent, which inspects the repository and any recorded
knowledge, asks you only for what it can't determine and that matters for building
features — purpose, users, hard constraints, project-wide decisions — and records it
briefly in .gnomon/context/PROJECT.md. It is not required: the normal path starts
with "gnomon spec".`,
	Args: requireArgs("no arguments", cobra.NoArgs),
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := os.Getwd()
		if err != nil {
			return err
		}
		report, err := orchestrate.Describe(root, *describeAgentFlag, agentChooserForInvocation())
		return renderReport(report, err)
	},
}

var describeAgentFlag *string

func init() {
	describeAgentFlag = registerAgentFlag(describeCmd)
	rootCmd.AddCommand(describeCmd)
}
