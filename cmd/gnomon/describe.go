package main

import (
	"os"

	"github.com/spf13/cobra"

	"gnomon/internal/orchestrate"
)

var describeCmd = &cobra.Command{
	Use:     "describe",
	Short:   "Establish or update initial project knowledge (Initial Knowledge Establishment)",
	GroupID: groupProject,
	Long: `Establish or update initial project knowledge (Initial Knowledge Establishment).

Hands the project to the Agent, which inspects existing project knowledge and
repository evidence — source, docs, configuration, and whatever else carries
project intent — and reconciles it into recorded knowledge. When something
material can't be established from that evidence, the Agent asks you directly,
in the same session, rather than guessing. This is what lets the same command
work for an established project and a brand-new one: what's already evident is
used as-is, and only genuine gaps are ever asked about.`,
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
