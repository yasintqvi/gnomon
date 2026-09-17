package main

import (
	"os"

	"github.com/spf13/cobra"

	"gnomon/internal/orchestrate"
)

var initCmd = &cobra.Command{
	Use:     "init",
	Short:   "Initialize a Gnomon project in the current directory",
	GroupID: groupProject,
	Args:    requireArgs("no arguments", cobra.NoArgs),
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := os.Getwd()
		if err != nil {
			return err
		}
		report, err := orchestrate.Init(root)
		return renderReport(report, err)
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
