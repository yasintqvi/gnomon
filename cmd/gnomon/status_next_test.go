package main

import "testing"

// TestStatus_IsRegisteredWithNoArgs proves gnomon status is a real, registered Cobra command,
// matching cli/COMMAND_SURFACE.md's bare `gnomon status` (no arguments).
func TestStatus_IsRegisteredWithNoArgs(t *testing.T) {
	found := false
	for _, c := range rootCmd.Commands() {
		if c.Name() == "status" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected top-level %q to be registered", "status")
	}
	if err := statusCmd.Args(statusCmd, nil); err != nil {
		t.Fatalf("expected zero args to be accepted, got: %v", err)
	}
	if err := statusCmd.Args(statusCmd, []string{"unexpected"}); err == nil {
		t.Fatalf("expected any argument to be refused")
	}
}

// TestStatus_HasNoAgentFlag proves status is Agent-free: cli/COMMAND_SURFACE.md classifies it
// CLI-native, and "deterministic commands never accept [--agent], since they never touch an Agent
// at all."
func TestStatus_HasNoAgentFlag(t *testing.T) {
	if statusCmd.Flags().Lookup("agent") != nil {
		t.Fatalf("did not expect status to register --agent — it never invokes an Agent")
	}
}

// TestNext_IsRegisteredWithNoArgs proves gnomon next is a real, registered Cobra command,
// matching cli/COMMAND_SURFACE.md's bare `gnomon next` (no arguments) — next never takes a
// Specification or workflow argument; its whole point is repository-wide, cross-session guidance.
func TestNext_IsRegisteredWithNoArgs(t *testing.T) {
	found := false
	for _, c := range rootCmd.Commands() {
		if c.Name() == "next" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected top-level %q to be registered", "next")
	}
	if err := nextCmd.Args(nextCmd, nil); err != nil {
		t.Fatalf("expected zero args to be accepted, got: %v", err)
	}
	if err := nextCmd.Args(nextCmd, []string{"unexpected"}); err == nil {
		t.Fatalf("expected any argument to be refused")
	}
}

// TestNext_HasNoAgentFlag proves next is Agent-free, the same CLI-native category as status and
// validate.
func TestNext_HasNoAgentFlag(t *testing.T) {
	if nextCmd.Flags().Lookup("agent") != nil {
		t.Fatalf("did not expect next to register --agent — it never invokes an Agent")
	}
}

// TestSlice6Commands_DistinctFromValidateAndRun confirms status/next remain their own commands,
// never aliases or wrappers of validate/run.
func TestSlice6Commands_DistinctFromValidateAndRun(t *testing.T) {
	if statusCmd == validateCmd || nextCmd == validateCmd {
		t.Fatalf("status/next must be distinct commands from validate, not the same registration")
	}
	if statusCmd == runCmd || nextCmd == runCmd {
		t.Fatalf("status/next must be distinct commands from run, not the same registration")
	}
}

// TestCommandSurface_FinalShape is the current authoritative statement of the full public
// command surface: the everyday interactive interface, CLI-native Human-exclusive/deterministic
// operations, project setup, guidance/inspection, and the advanced run/agent path. Contract-driven
// Agent workflows with no non-run-representable operation of their own (implement, test, verify,
// review, finalize, spec discover, spec define) are deliberately absent — see
// TestWorkflowCommands_NoLongerRegisteredAsTopLevel.
func TestCommandSurface_FinalShape(t *testing.T) {
	want := []string{
		"init", "describe", "bootstrap", "spec", "approve", "revoke",
		"status", "validate", "next", "run", "agent",
	}
	have := map[string]bool{}
	for _, c := range rootCmd.Commands() {
		have[c.Name()] = true
	}
	for _, name := range want {
		if !have[name] {
			t.Fatalf("expected top-level %q to be registered", name)
		}
	}
}
