package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"gnomon/internal/orchestrate"
)

var approveCmd = &cobra.Command{
	Use:     "approve <SPEC-id>",
	Short:   "Grant Human approval to a Specification's current content",
	GroupID: groupSpecification,
	Args:    requireArgs("<SPEC-id>", cobra.ExactArgs(1)),
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := os.Getwd()
		if err != nil {
			return err
		}
		report, err := orchestrate.Approve(root, args[0], stdinPrompt)
		return renderReport(report, err)
	},
}

// stdinPrompt is the real, interactive one-time attribution fallback (Step 6) — wired only here,
// at the CLI edge, so internal/orchestrate stays testable without real stdin.
func stdinPrompt(message string) (string, error) {
	fmt.Fprint(os.Stderr, message)
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		return scanner.Text(), nil
	}
	return "", scanner.Err()
}

func init() {
	rootCmd.AddCommand(approveCmd)
}
