package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"gnomon/internal/present"
)

// version is Gnomon's build version, reported by `gnomon --version`. "dev" is the value for an
// ordinary local build; a release build overrides it at compile time with
// -ldflags "-X main.version=vX.Y.Z". Step 14 owns tagging and cutting a release; this is only the
// injection point it uses.
var version = "dev"

var verbose bool

// Command groups, in `gnomon --help` order — a presentation grouping only (cli/COMMAND_SURFACE.md),
// never a change to any command's syntax or placement.
const (
	groupProject       = "project"
	groupSpecification = "specification"
	groupGuidance      = "guidance"
	groupTool          = "tool"
	groupAdvanced      = "advanced"
)

var rootCmd = &cobra.Command{
	Use:     "gnomon",
	Version: version,
	Short:   "AI Software Engineering System",
	Long: `Gnomon — AI Software Engineering System

Gnomon runs engineering work as an explicit, document-driven process: Specifications
record approved behavior, Workflows carry it out through an Agent, and deterministic
gates — never an Agent's own judgment — decide what is currently valid to do.

  gnomon status   what state is this project currently in?
  gnomon next     what actions are currently valid, given that state?
  gnomon --help   the full command surface`,
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "show low-level diagnostics (exit codes, run IDs, paths)")

	rootCmd.AddGroup(
		&cobra.Group{ID: groupProject, Title: "Project:"},
		&cobra.Group{ID: groupSpecification, Title: "Specification:"},
		&cobra.Group{ID: groupGuidance, Title: "Guidance / Inspection:"},
		&cobra.Group{ID: groupTool, Title: "Tool:"},
		&cobra.Group{ID: groupAdvanced, Title: "Advanced:"},
	)
	// Shell completion is a Cobra default, not a designed v1 feature — kept functional but hidden
	// from the help listing.
	rootCmd.CompletionOptions.HiddenDefaultCmd = true
}

// Execute runs the CLI, exiting non-zero on any command error. Only errors with no Report reach
// here — a command with a Report renders and exits directly (renderReport), never a second line.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprint(os.Stderr, present.RenderErrorColor(err, colorEnabled()))
		os.Exit(1)
	}
}

// renderReport prints r, then exits with a failing status if err is non-nil — bypassing Cobra's
// own error path so the Report is the only thing the Human sees.
func renderReport(r *present.Report, err error) error {
	if r == nil {
		return err
	}
	fmt.Print(present.RenderWithOptions(r, present.RenderOptions{Verbose: verbose, Color: colorEnabled()}))
	if err != nil {
		os.Exit(1)
	}
	return nil
}

// colorEnabled decides whether ANSI color is appropriate: never when NO_COLOR is set
// (https://no-color.org), and only when stdout is a real, interactive terminal.
func colorEnabled() bool {
	if _, set := os.LookupEnv("NO_COLOR"); set {
		return false
	}
	return term.IsTerminal(int(os.Stdout.Fd()))
}
