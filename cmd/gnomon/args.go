package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

// requireArgs wraps an existing Cobra positional-argument validator, replacing its default error
// text with one that names exactly what the command expects — never changing which argument
// counts are accepted or rejected, only how a rejection reads. Cobra's own default messages
// ("accepts 1 arg(s), received 0") are accurate but say nothing about what the missing argument
// actually is.
func requireArgs(usage string, validate cobra.PositionalArgs) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if err := validate(cmd, args); err != nil {
			return fmt.Errorf("%s: expected %s", cmd.CommandPath(), usage)
		}
		return nil
	}
}
