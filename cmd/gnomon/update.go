package main

import (
	"os"
	"runtime"

	"github.com/spf13/cobra"

	"gnomon/internal/selfupdate"
)

var updateCmd = &cobra.Command{
	Use:     "update",
	Short:   "Update Gnomon to the latest stable release",
	GroupID: groupTool,
	Long: `Update Gnomon to the latest stable release.

Downloads the latest stable release from Gnomon's official GitHub Releases,
verifies it against the release's own published checksum, and replaces the
currently running executable with it. Prereleases and drafts are never
installed.

This requires no Gnomon project and works from any directory — updating the
CLI itself is a Human/installation-level action, never a project-level one,
and never happens automatically as a side effect of any other command.

If the executable's directory cannot be written to, nothing is changed;
gnomon never requests elevated privileges on your behalf.`,
	Args: requireArgs("no arguments", cobra.NoArgs),
	RunE: func(cmd *cobra.Command, args []string) error {
		execPath, err := selfupdate.ExecutablePath()
		if err != nil {
			return err
		}
		report, err := selfupdate.Run(os.Stdout, selfupdate.Options{
			CurrentVersion: version,
			GOOS:           runtime.GOOS,
			GOARCH:         runtime.GOARCH,
			ExecutablePath: execPath,
			Fetcher:        selfupdate.NewGitHubFetcher(),
		})
		return renderReport(report, err)
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
