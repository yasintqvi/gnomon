package main

import "testing"

// TestUpdateCmd_IsRegisteredUnderTool proves `gnomon update` exists as a top-level command,
// grouped with Tool (not Advanced merely because it is new — see update.go).
func TestUpdateCmd_IsRegisteredUnderTool(t *testing.T) {
	found := false
	for _, c := range rootCmd.Commands() {
		if c.Name() != "update" {
			continue
		}
		found = true
		if c.GroupID != groupTool {
			t.Fatalf("expected update to be grouped under Tool, got GroupID %q", c.GroupID)
		}
	}
	if !found {
		t.Fatal("expected \"update\" to be registered as a top-level command")
	}
}

// TestUpdateCmd_RejectsArgs proves update takes no positional arguments.
func TestUpdateCmd_RejectsArgs(t *testing.T) {
	if err := updateCmd.Args(updateCmd, nil); err != nil {
		t.Fatalf("expected zero args to be accepted, got: %v", err)
	}
	if err := updateCmd.Args(updateCmd, []string{"unexpected"}); err == nil {
		t.Fatal("expected any argument to be refused")
	}
}

// TestUpdateCmd_HasNoAgentFlag proves update never touches an Agent — it is a deterministic,
// CLI-native command exactly like init/approve/revoke/agent, not a workflow invocation.
func TestUpdateCmd_HasNoAgentFlag(t *testing.T) {
	if updateCmd.Flags().Lookup("agent") != nil {
		t.Fatal("did not expect update to register the shared --agent flag")
	}
}
