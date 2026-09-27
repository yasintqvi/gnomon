package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

// requireArgs wraps a Cobra positional-argument validator, replacing its default error text
// ("accepts 1 arg(s), received 0") with one naming what the command actually expects.
func requireArgs(usage string, validate cobra.PositionalArgs) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if err := validate(cmd, args); err != nil {
			return fmt.Errorf("%s: expected %s", cmd.CommandPath(), usage)
		}
		return nil
	}
}
