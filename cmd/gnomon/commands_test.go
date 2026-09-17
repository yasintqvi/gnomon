package main

import "testing"

// TestSlice2Commands_AreRegistered proves gnomon describe/bootstrap/verify/review/finalize are
// real, registered Cobra commands, not merely orchestrate-package functions nothing wires up.
func TestSlice2Commands_AreRegistered(t *testing.T) {
	want := []string{"describe", "bootstrap", "verify", "review", "finalize"}
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

// TestSlice2Commands_HaveAgentFlag proves every Agent-invoking command in this slice accepts the
// shared --agent override registered the same way implement's already does.
func TestSlice2Commands_HaveAgentFlag(t *testing.T) {
	agentInvoking := map[string]bool{"describe": true, "bootstrap": true, "verify": true, "review": true, "finalize": true}
	for _, c := range rootCmd.Commands() {
		if !agentInvoking[c.Name()] {
			continue
		}
		if c.Flags().Lookup("agent") == nil {
			t.Fatalf("expected %q to register the shared --agent flag", c.Name())
		}
	}
}

// TestSlice2Commands_VerifyAndReviewAcceptOptionalTarget confirms verify/review's arity — an
// optional, single, generic target argument — matching cli/COMMAND_SURFACE.md's `verify [target]`
// / `review [target]`.
func TestSlice2Commands_VerifyAndReviewAcceptOptionalTarget(t *testing.T) {
	for _, name := range []string{"verify", "review"} {
		for _, c := range rootCmd.Commands() {
			if c.Name() != name {
				continue
			}
			if err := c.Args(c, nil); err != nil {
				t.Fatalf("%s: expected zero args to be accepted (target is optional), got: %v", name, err)
			}
			if err := c.Args(c, []string{"one"}); err != nil {
				t.Fatalf("%s: expected exactly one arg to be accepted, got: %v", name, err)
			}
			if err := c.Args(c, []string{"one", "two"}); err == nil {
				t.Fatalf("%s: expected more than one arg to be refused", name)
			}
		}
	}
}

// TestSlice2Commands_NoTargetCommandsRejectArgs confirms describe/bootstrap/finalize take no
// arguments at all, matching cli/COMMAND_SURFACE.md's bare `describe`/`bootstrap`/`finalize`.
func TestSlice2Commands_NoTargetCommandsRejectArgs(t *testing.T) {
	for _, name := range []string{"describe", "bootstrap", "finalize"} {
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

// TestSlice2Commands_ScopeBoundary confirms none of the commands out of scope for Slice 2 were
// accidentally registered as a side effect of that slice's work. "test", "spec
// discover"/"spec define", "revoke", "validate"/"run", and "status"/"next" were correctly out of
// scope for Slice 2 specifically and are no longer listed here — later slices legitimately add
// them; see TestSlice3Commands_ScopeBoundary, TestSlice4Commands_ScopeBoundary,
// TestSlice5Commands_ScopeBoundary, and TestSlice6Commands_ScopeBoundary.
func TestSlice2Commands_ScopeBoundary(t *testing.T) {
	outOfScope := map[string]bool{}
	for _, c := range rootCmd.Commands() {
		if outOfScope[c.Name()] {
			t.Fatalf("did not expect %q to be registered — out of scope for this slice", c.Name())
		}
	}
}
