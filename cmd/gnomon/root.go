package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"gnomon/internal/present"
)

// version is Gnomon's build version, reported by `gnomon --version`. "dev" is the value for an
// ordinary local build; a real release build overrides it at compile time with
// -ldflags "-X main.version=vX.Y.Z" — the smallest conventional Go mechanism for this, requiring
// no new dependency and no version string duplicated anywhere else in the source tree. Step 14
// owns tagging and actually cutting a release; this is only the injection point it will use.
var version = "dev"

var verbose bool

// Command groups, in the order they should appear in `gnomon --help` — a presentation grouping
// only (cli/COMMAND_SURFACE.md's own categorization), never a change to any command's syntax,
// hierarchy, or top-level/nested placement. approve and revoke remain top-level, grouped under
// Specification, exactly as cli/COMMAND_SURFACE.md requires.
const (
	groupProject       = "project"
	groupSpecification = "specification"
	groupEngineering   = "engineering"
	groupGuidance      = "guidance"
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
		&cobra.Group{ID: groupEngineering, Title: "Engineering:"},
		&cobra.Group{ID: groupGuidance, Title: "Guidance / Inspection:"},
		&cobra.Group{ID: groupAdvanced, Title: "Advanced:"},
	)
	// Shell completion is a Cobra default, not a designed part of the v1 command surface — kept
	// functional (it still works if invoked directly) but out of the visible help listing, so it
	// never reads as an advertised v1 feature.
	rootCmd.CompletionOptions.HiddenDefaultCmd = true
}

// Execute runs the CLI, exiting non-zero on any command error. Only errors with no Report reach
// here — a command that already produced a Report renders it and exits directly, since the
// Report already explains the outcome and a second raw error line would be redundant.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprint(os.Stderr, present.RenderErrorColor(err, colorEnabled()))
		os.Exit(1)
	}
}

// renderReport prints r in Gnomon's standard presentation, then, if err is non-nil, exits with a
// failing status directly — bypassing Cobra's own error path so the Report is the only thing the
// Human sees, rather than the Report followed by a second, lower-level "error: ..." line.
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

// colorEnabled decides, once per invocation, whether ANSI color is appropriate: never when
// NO_COLOR is set (https://no-color.org, a convention Gnomon respects rather than reinvents), and
// only when stdout is a real, interactive terminal — never for a redirected file, a pipe, or CI
// log capture. term.IsTerminal (already a dependency, already used by select.go) correctly
// distinguishes a real TTY from any other character device, unlike a bare os.ModeCharDevice
// check.
func colorEnabled() bool {
	if _, set := os.LookupEnv("NO_COLOR"); set {
		return false
	}
	return term.IsTerminal(int(os.Stdout.Fd()))
}
