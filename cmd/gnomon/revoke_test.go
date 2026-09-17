package main

import "testing"

// TestRevoke_IsRegisteredAtTopLevel proves gnomon revoke is a real, registered Cobra command,
// kept top-level (not nested under spec) — matching cli/COMMAND_SURFACE.md's `gnomon revoke
// <SPEC-id>` and its stated rationale for keeping approve/revoke visually equal to the workflow
// verbs.
func TestRevoke_IsRegisteredAtTopLevel(t *testing.T) {
	found := false
	for _, c := range rootCmd.Commands() {
		if c.Name() == "revoke" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected top-level %q to be registered", "revoke")
	}

	for _, c := range rootCmd.Commands() {
		if c.Name() != "spec" {
			continue
		}
		for _, sub := range c.Commands() {
			if sub.Name() == "revoke" {
				t.Fatalf("did not expect revoke to be nested under spec — cli/COMMAND_SURFACE.md keeps it top-level")
			}
		}
	}
}

// TestRevoke_TakesExactlyOneArgument confirms the arity matches `gnomon revoke <SPEC-id>`.
func TestRevoke_TakesExactlyOneArgument(t *testing.T) {
	if err := revokeCmd.Args(revokeCmd, []string{"SPEC-001"}); err != nil {
		t.Fatalf("expected exactly one arg to be accepted, got: %v", err)
	}
	if err := revokeCmd.Args(revokeCmd, nil); err == nil {
		t.Fatalf("expected zero args to be refused — a target Specification is required")
	}
	if err := revokeCmd.Args(revokeCmd, []string{"SPEC-001", "SPEC-002"}); err == nil {
		t.Fatalf("expected more than one arg to be refused")
	}
}

// TestRevoke_HasNoAgentFlag proves revoke is CLI-native and Agent-free: unlike every
// Agent-invoking command (implement, describe, bootstrap, verify, review, finalize, spec
// discover, spec define, test), it never registers the shared --agent override, because it never
// resolves or invokes an Agent at all — cli/COMMAND_SURFACE.md classifies it "CLI-native,
// Human-owned", the same category as approve, which also has no --agent flag.
func TestRevoke_HasNoAgentFlag(t *testing.T) {
	if revokeCmd.Flags().Lookup("agent") != nil {
		t.Fatalf("did not expect revoke to register --agent — it is CLI-native and never invokes an Agent")
	}
}

// TestApprove_AlsoHasNoAgentFlag confirms the established precedent revoke is following: approve
// (Slice 0-era, unmodified) has never had this flag either.
func TestApprove_AlsoHasNoAgentFlag(t *testing.T) {
	if approveCmd.Flags().Lookup("agent") != nil {
		t.Fatalf("did not expect approve to register --agent")
	}
}

// TestSlice4Commands_ScopeBoundary confirms none of the commands still explicitly out of scope
// for Slice 4 were registered as a side effect of this slice's work. "validate"/"run" and
// "status"/"next" were correctly out of scope for Slice 4 specifically and are no longer listed
// here — later slices legitimately add them; see TestSlice5Commands_ScopeBoundary and
// TestSlice6Commands_ScopeBoundary.
func TestSlice4Commands_ScopeBoundary(t *testing.T) {
	outOfScope := map[string]bool{}
	for _, c := range rootCmd.Commands() {
		if outOfScope[c.Name()] {
			t.Fatalf("did not expect %q to be registered — out of scope for this slice", c.Name())
		}
	}
}
