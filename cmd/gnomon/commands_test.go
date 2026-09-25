package main

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"gnomon/internal/orchestrate"
)

// TestProjectCommands_AreRegistered proves gnomon describe/bootstrap remain real, registered
// Cobra commands — project setup is out of scope for the Specification-centered CLI redesign and
// is untouched by it.
func TestProjectCommands_AreRegistered(t *testing.T) {
	want := []string{"describe", "bootstrap"}
	got := map[string]bool{}
	for _, c := range rootCmd.Commands() {
		got[c.Name()] = true
	}
	for _, name := range want {
		if !got[name] {
			t.Fatalf("expected %q to be registered as a top-level command", name)
		}
	}
}

// recommendedCommandIsRegistered walks the real, live Cobra command tree (rootCmd, and one level
// of subcommands) and reports whether the command next's first line leads with actually exists —
// deep enough to catch the class of bug this guards against (a recommendation naming a command
// that was removed, or misspelling a subcommand like "spec create") without needing to parse or
// simulate a full invocation. next may be a bare command ("gnomon bootstrap") or a multi-word one
// ("gnomon spec create \"<title>\""); only the words up to and including the first quoted or
// placeholder-looking argument are treated as the command path.
//
// This only validates the one command the recommendation leads with — it does not parse every
// command mention that might appear later in a longer sentence (deliberately: doing that without
// a real argument parser would be exactly the "complicated command parsing" this guard should
// avoid). Every current recommendation in this codebase names exactly one command, in its own
// first line, for this reason; TestOnboardingRecommendations_DoNotNameKnownRemovedCommands
// (internal/orchestrate) is the complementary denylist guard against specific known-stale phrases
// anywhere in a recommendation's text, including ones this function's leading-command check alone
// would not reach.
func recommendedCommandIsRegistered(t *testing.T, next string) bool {
	t.Helper()
	firstLine := strings.SplitN(next, "\n", 2)[0]
	fields := strings.Fields(firstLine)
	if len(fields) < 2 || fields[0] != "gnomon" {
		t.Fatalf("expected a recommendation starting with \"gnomon <command>\", got %q", firstLine)
	}

	cmd, ok := findCommand(rootCmd, fields[1])
	if !ok {
		return false
	}
	// A second bare word (no leading "<", "\"", or "-") names a registered subcommand too, e.g.
	// "gnomon spec create". Anything else (a placeholder like <title>, or nothing) means the
	// recommendation targets cmd itself.
	if len(fields) >= 3 {
		second := fields[2]
		if second[0] != '<' && second[0] != '"' && second[0] != '-' {
			_, ok = findCommand(cmd, second)
			return ok
		}
	}
	return true
}

func findCommand(parent *cobra.Command, name string) (*cobra.Command, bool) {
	for _, c := range parent.Commands() {
		if c.Name() == name {
			return c, true
		}
	}
	return nil, false
}

// TestDescribeNext_RecommendsARegisteredCommand and TestBootstrapNext_RecommendsARegisteredCommand
// are the genuine regression guard for the onboarding recommendations orchestrate.DescribeSuccessNext/
// orchestrate.BootstrapSuccessNext contain: unlike a check against the constant's own text (which
// cannot tell a valid command from a plausible-looking removed one — that was exactly how
// "gnomon spec discover" went undetected), this walks the real, currently registered rootCmd tree.
func TestDescribeNext_RecommendsARegisteredCommand(t *testing.T) {
	if !recommendedCommandIsRegistered(t, orchestrate.DescribeSuccessNext) {
		t.Fatalf("orchestrate.DescribeSuccessNext (%q) does not name a currently registered command", orchestrate.DescribeSuccessNext)
	}
}

func TestBootstrapNext_RecommendsARegisteredCommand(t *testing.T) {
	if !recommendedCommandIsRegistered(t, orchestrate.BootstrapSuccessNext) {
		t.Fatalf("orchestrate.BootstrapSuccessNext (%q) does not name a currently registered command", orchestrate.BootstrapSuccessNext)
	}
}

// TestProjectCommands_HaveAgentFlag confirms describe/bootstrap still accept --agent.
func TestProjectCommands_HaveAgentFlag(t *testing.T) {
	for _, name := range []string{"describe", "bootstrap"} {
		for _, c := range rootCmd.Commands() {
			if c.Name() != name {
				continue
			}
			if c.Flags().Lookup("agent") == nil {
				t.Fatalf("expected %q to register the shared --agent flag", name)
			}
		}
	}
}

// TestProjectCommands_RejectArgs confirms describe/bootstrap still take no arguments.
func TestProjectCommands_RejectArgs(t *testing.T) {
	for _, name := range []string{"describe", "bootstrap"} {
		for _, c := range rootCmd.Commands() {
			if c.Name() != name {
				continue
			}
			if err := c.Args(c, nil); err != nil {
				t.Fatalf("%s: expected zero args to be accepted, got: %v", name, err)
			}
			if err := c.Args(c, []string{"unexpected"}); err == nil {
				t.Fatalf("%s: expected any argument to be refused", name)
			}
		}
	}
}

// TestDescribeCmd_NonInteractive_NeverHangs proves gnomon describe never blocks on stdin —
// describeCmd.RunE returns promptly (here: refused for lack of a Gnomon project at the test
// process's own working directory) in the normal non-interactive shape (go test's own stdin,
// CI, scripted use), mirroring TestSpecCmd_NonInteractive_RefusesRatherThanHangs's same proof
// for spec.
func TestDescribeCmd_NonInteractive_NeverHangs(t *testing.T) {
	if err := describeCmd.RunE(describeCmd, nil); err == nil {
		t.Fatalf("expected a non-interactive invocation with no project here to be refused")
	}
}

// TestWorkflowCommands_NoLongerRegisteredAsTopLevel proves implement/test/verify/review/finalize
// (and spec discover/spec define) have no dedicated top-level command: each is a Contract-driven
// Agent workflow with no Human-exclusive or non-Contract-representable operation of its own, so
// each is reachable only through gnomon run <identity> [target] (advanced/non-interactive) or as
// a contextual action inside gnomon spec <SPEC-id> (everyday/interactive) — never through both a
// dedicated command and run at once.
func TestWorkflowCommands_NoLongerRegisteredAsTopLevel(t *testing.T) {
	removed := map[string]bool{
		"implement": true, "test": true, "verify": true, "review": true, "finalize": true,
	}
	for _, c := range rootCmd.Commands() {
		if removed[c.Name()] {
			t.Fatalf("did not expect %q to remain a dedicated top-level command; it is reachable via gnomon run", c.Name())
		}
	}
	for _, sub := range specCmd.Commands() {
		if sub.Name() == "discover" || sub.Name() == "define" {
			t.Fatalf("did not expect gnomon spec %s to remain registered; it is reachable via gnomon run", sub.Name())
		}
	}
}

// TestWorkflowCommands_ReachableViaRun confirms the built-in identity each removed command used
// to wrap is still a valid gnomon run target — the capability was relocated, not lost.
func TestWorkflowCommands_ReachableViaRun(t *testing.T) {
	for _, identity := range []string{
		"implementation", "testing", "verification", "review", "git-finalization",
		"specification-discovery", "specification-definition",
	} {
		if err := runCmd.Args(runCmd, []string{identity}); err != nil {
			t.Fatalf("expected gnomon run %s to be a syntactically valid invocation, got: %v", identity, err)
		}
	}
}
