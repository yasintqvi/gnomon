package main

import "testing"

// TestValidate_IsRegisteredTopLevelWithNoArgs proves gnomon validate is a real, registered Cobra
// command, matching cli/COMMAND_SURFACE.md's bare `gnomon validate` (no arguments).
func TestValidate_IsRegisteredTopLevelWithNoArgs(t *testing.T) {
	found := false
	for _, c := range rootCmd.Commands() {
		if c.Name() == "validate" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected top-level %q to be registered", "validate")
	}
	if err := validateCmd.Args(validateCmd, nil); err != nil {
		t.Fatalf("expected zero args to be accepted, got: %v", err)
	}
	if err := validateCmd.Args(validateCmd, []string{"unexpected"}); err == nil {
		t.Fatalf("expected any argument to be refused")
	}
}

// TestValidate_HasNoAgentFlag proves validate is Agent-free — the same CLI-native category as
// approve/revoke, never registering the shared --agent override.
func TestValidate_HasNoAgentFlag(t *testing.T) {
	if validateCmd.Flags().Lookup("agent") != nil {
		t.Fatalf("did not expect validate to register --agent — it never invokes an Agent")
	}
}

// TestRun_IsRegisteredWithExactSyntax proves gnomon run is registered with the exact syntax
// cli/COMMAND_SURFACE.md specifies: `run <workflow-identity> [target]` — one required argument,
// one optional.
func TestRun_IsRegisteredWithExactSyntax(t *testing.T) {
	found := false
	for _, c := range rootCmd.Commands() {
		if c.Name() == "run" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected top-level %q to be registered", "run")
	}
	if err := runCmd.Args(runCmd, nil); err == nil {
		t.Fatalf("expected zero args to be refused — a workflow-identity is required")
	}
	if err := runCmd.Args(runCmd, []string{"implementation"}); err != nil {
		t.Fatalf("expected exactly one arg (identity only) to be accepted, got: %v", err)
	}
	if err := runCmd.Args(runCmd, []string{"implementation", "SPEC-001"}); err != nil {
		t.Fatalf("expected two args (identity + target) to be accepted, got: %v", err)
	}
	if err := runCmd.Args(runCmd, []string{"implementation", "SPEC-001", "extra"}); err == nil {
		t.Fatalf("expected more than two args to be refused")
	}
}

// TestRun_HasAgentFlag proves run accepts the shared --agent override, registered the same way
// every other Agent-invoking command's already does.
func TestRun_HasAgentFlag(t *testing.T) {
	if runCmd.Flags().Lookup("agent") == nil {
		t.Fatalf("expected run to register the shared --agent flag")
	}
}

// TestSlice5Commands_ScopeBoundary confirms none of the commands still explicitly out of scope
// for Slice 5 were registered as a side effect of this slice's work. "status"/"next" were
// correctly out of scope for Slice 5 specifically and are no longer listed here — Slice 6
// legitimately adds them; see TestSlice6Commands_ScopeBoundary.
func TestSlice5Commands_ScopeBoundary(t *testing.T) {
	outOfScope := map[string]bool{}
	for _, c := range rootCmd.Commands() {
		if outOfScope[c.Name()] {
			t.Fatalf("did not expect %q to be registered — out of scope for this slice", c.Name())
		}
	}
}
