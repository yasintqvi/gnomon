package main

import "testing"

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
