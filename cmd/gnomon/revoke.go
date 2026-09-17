package main

import (
	"os"

	"github.com/spf13/cobra"

	"gnomon/internal/orchestrate"
)

var revokeCmd = &cobra.Command{
	Use:     "revoke <SPEC-id>",
	Short:   "Revoke Human approval from a Specification's current content",
	GroupID: groupSpecification,
	Args:    requireArgs("<SPEC-id>", cobra.ExactArgs(1)),
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := os.Getwd()
		if err != nil {
			return err
		}
		report, err := orchestrate.Revoke(root, args[0], stdinPrompt)
		return renderReport(report, err)
	},
}

func init() {
	rootCmd.AddCommand(revokeCmd)
}
