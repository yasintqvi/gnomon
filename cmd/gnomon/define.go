package main

import (
	"os"

	"github.com/spf13/cobra"

	"gnomon/internal/orchestrate"
)

var defineCmd = &cobra.Command{
	Use:     `define <"request" | SPEC-id> [change]`,
	Short:   "Start a new feature from a request, or define an existing Specification",
	GroupID: groupSpecification,
	Long: `Define a Specification with the Agent.

With a request in your own words, it starts a new feature: Gnomon creates the Draft
Specification (showing its identity), saves your request in it, and runs Define. Running
the same request again continues that Draft instead of creating another one.

With a Specification identity, it defines that existing Specification; any further text
is handed to the Agent as the change you want.

A Define that stops before the Specification is ready leaves it as a Draft; continue it
with "gnomon define <SPEC-id>". "gnomon spec create <title>" still creates an empty Draft
without running the Agent.`,
	Example: `  gnomon define "Members can reserve a tool that is out"
  gnomon define "Late fees: 1 euro per day late" --title "Late fees"
  gnomon define SPEC-004
  gnomon define SPEC-004 "members may now see retired items"`,
	Args: requireArgs(`<"request" | SPEC-id> [change]`, cobra.MinimumNArgs(1)),
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := os.Getwd()
		if err != nil {
			return err
		}
		report, err := orchestrate.Define(root, args, orchestrate.DefineOptions{Title: defineTitle, Detailed: defineDetailed},
			*defineAgentFlag, agentChooserForInvocation())
		return renderReport(report, err)
	},
}

var (
	defineTitle     string
	defineDetailed  bool
	defineAgentFlag *string
)

func init() {
	defineCmd.Flags().StringVar(&defineTitle, "title", "", "title for a new Specification (default: from the request)")
	defineCmd.Flags().BoolVar(&defineDetailed, "detailed", false, "use the detailed use-case template for a new Specification")
	defineAgentFlag = registerAgentFlag(defineCmd)
	rootCmd.AddCommand(defineCmd)
}
