package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

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
		report, err := orchestrate.Approve(root, args[0], stdinPrompt, approveConfirm)
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

// approveConfirm always prints the warning, then, only with a real interactive terminal, asks
// before proceeding — defaulting to No. Non-interactive invocations always proceed after printing.
func approveConfirm(warning string) (bool, error) {
	fmt.Fprintln(os.Stderr, warning)
	if !stdinIsInteractive() {
		return true, nil
	}
	answer, err := stdinPrompt("Approve anyway? [y/N]: ")
	if err != nil {
		return false, err
	}
	answer = strings.ToLower(strings.TrimSpace(answer))
	return answer == "y" || answer == "yes", nil
}

func init() {
	rootCmd.AddCommand(approveCmd)
}
