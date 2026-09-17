package main

import "testing"

// TestSlice3Commands_AreRegisteredAtCorrectHierarchy proves gnomon spec discover, gnomon spec
// define, and gnomon test are real, registered Cobra commands at the exact hierarchy
// cli/COMMAND_SURFACE.md specifies: discover/define nested under the existing "spec" parent
// (never a conflicting top-level command), test kept top-level.
func TestSlice3Commands_AreRegisteredAtCorrectHierarchy(t *testing.T) {
	testFound := false
	for _, c := range rootCmd.Commands() {
		if c.Name() == "test" {
			testFound = true
		}
	}
	if !testFound {
		t.Fatalf("expected top-level %q to be registered", "test")
	}

	var specSubcommands map[string]bool
	for _, c := range rootCmd.Commands() {
		if c.Name() != "spec" {
			continue
		}
		specSubcommands = map[string]bool{}
		for _, sub := range c.Commands() {
			specSubcommands[sub.Name()] = true
		}
	}
	if specSubcommands == nil {
		t.Fatalf("expected the existing %q parent command to still be registered", "spec")
	}
	for _, want := range []string{"create", "discover", "define"} {
		if !specSubcommands[want] {
			t.Fatalf("expected `gnomon spec %s` to be registered under the existing spec command", want)
		}
	}

	// discover/define must never also exist as conflicting top-level commands.
	for _, c := range rootCmd.Commands() {
		if c.Name() == "discover" || c.Name() == "define" {
			t.Fatalf("did not expect %q to be registered as a top-level command — it belongs under spec", c.Name())
		}
	}
}

// TestSlice3Commands_HaveAgentFlag proves spec discover, spec define, and test all accept the
// shared --agent override, registered the same way every other Agent-invoking command's already
// does.
func TestSlice3Commands_HaveAgentFlag(t *testing.T) {
	if testCmd.Flags().Lookup("agent") == nil {
		t.Fatalf("expected top-level test to register the shared --agent flag")
	}
	if specDiscoverCmd.Flags().Lookup("agent") == nil {
		t.Fatalf("expected spec discover to register the shared --agent flag")
	}
	if specDefineCmd.Flags().Lookup("agent") == nil {
		t.Fatalf("expected spec define to register the shared --agent flag")
	}
}

// TestSlice3Commands_ArgumentShapesMatchDesign confirms each command's arity matches
// cli/COMMAND_SURFACE.md / the workflow's own Contract: `spec discover` takes no argument
// (specification_reference: none); `spec define <SPEC-id>` takes exactly one
// (specification_reference: required); `test [SPEC-id]` takes zero or one
// (specification_reference: optional).
func TestSlice3Commands_ArgumentShapesMatchDesign(t *testing.T) {
	if err := specDiscoverCmd.Args(specDiscoverCmd, nil); err != nil {
		t.Fatalf("spec discover: expected zero args to be accepted, got: %v", err)
	}
	if err := specDiscoverCmd.Args(specDiscoverCmd, []string{"unexpected"}); err == nil {
		t.Fatalf("spec discover: expected any argument to be refused")
	}

	if err := specDefineCmd.Args(specDefineCmd, []string{"SPEC-001"}); err != nil {
		t.Fatalf("spec define: expected exactly one arg to be accepted, got: %v", err)
	}
	if err := specDefineCmd.Args(specDefineCmd, nil); err == nil {
		t.Fatalf("spec define: expected zero args to be refused — a target Specification is required")
	}
	if err := specDefineCmd.Args(specDefineCmd, []string{"SPEC-001", "SPEC-002"}); err == nil {
		t.Fatalf("spec define: expected more than one arg to be refused")
	}

	if err := testCmd.Args(testCmd, nil); err != nil {
		t.Fatalf("test: expected zero args to be accepted (target is optional), got: %v", err)
	}
	if err := testCmd.Args(testCmd, []string{"SPEC-001"}); err != nil {
		t.Fatalf("test: expected exactly one arg to be accepted, got: %v", err)
	}
	if err := testCmd.Args(testCmd, []string{"SPEC-001", "SPEC-002"}); err == nil {
		t.Fatalf("test: expected more than one arg to be refused")
	}
}

// TestSlice3Commands_ScopeBoundary confirms none of the commands still explicitly out of scope
// for Slice 3 were registered as a side effect of this slice's work. "revoke",
// "validate"/"run", and "status"/"next" were correctly out of scope for Slice 3 specifically and
// are no longer listed here — later slices legitimately add them; see
// TestSlice4Commands_ScopeBoundary, TestSlice5Commands_ScopeBoundary, and
// TestSlice6Commands_ScopeBoundary.
func TestSlice3Commands_ScopeBoundary(t *testing.T) {
	outOfScope := map[string]bool{}
	for _, c := range rootCmd.Commands() {
		if outOfScope[c.Name()] {
			t.Fatalf("did not expect %q to be registered — out of scope for this slice", c.Name())
		}
	}
}
