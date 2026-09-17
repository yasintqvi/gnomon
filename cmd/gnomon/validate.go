package main

import (
	"os"

	"github.com/spf13/cobra"

	"gnomon/internal/orchestrate"
)

var validateCmd = &cobra.Command{
	Use:     "validate",
	Short:   "Check deterministic Gnomon project invariants",
	GroupID: groupGuidance,
	Long: `Check deterministic Gnomon project invariants.

Answers: is the Gnomon structure itself valid? .gnomon/ structural integrity,
CONTRACT_VERSION support, every workflow's Contract, duplicate workflow identities,
and approval evidence structure — clean pass/fail, suited to scripts and CI.

validate is not status: it never reports which Specifications exist or their
lifecycle. Use "gnomon status" for that.`,
	Args: requireArgs("no arguments", cobra.NoArgs),
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := os.Getwd()
		if err != nil {
			return err
		}
		report, err := orchestrate.Validate(root)
		return renderReport(report, err)
	},
}

func init() {
	rootCmd.AddCommand(validateCmd)
}
