package main

import (
	"strings"
	"testing"
)

func TestTruncate_ShortStringUnchanged(t *testing.T) {
	if got := truncate("Password Reset", 30); got != "Password Reset" {
		t.Fatalf("unexpected: %q", got)
	}
}

func TestTruncate_LongStringEllipsized(t *testing.T) {
	got := truncate("A very long Specification title that exceeds the column width", 20)
	if !strings.HasSuffix(got, "…") {
		t.Fatalf("expected an ellipsis marker, got %q", got)
	}
	if !strings.HasPrefix(got, "A very long Specifi") {
		t.Fatalf("expected the first 19 characters preserved, got %q", got)
	}
}

func TestCapitalize(t *testing.T) {
	cases := map[string]string{
		"define":  "Define",
		"approve": "Approve",
		"":        "",
	}
	for in, want := range cases {
		if got := capitalize(in); got != want {
			t.Errorf("capitalize(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestShort_FingerprintTruncatedTo12(t *testing.T) {
	if got := short("abcdefghijklmnopqrstuvwxyz"); got != "abcdefghijkl" {
		t.Fatalf("unexpected: %q", got)
	}
	if got := short("short"); got != "short" {
		t.Fatalf("expected a fingerprint shorter than 12 to pass through unchanged, got %q", got)
	}
}

// --- gnomon spec: registration, args shape, non-interactive behavior ---

func TestSpecCmd_AcceptsZeroOrOneArg(t *testing.T) {
	if err := specCmd.Args(specCmd, nil); err != nil {
		t.Fatalf("expected zero args (the browser) to be accepted, got: %v", err)
	}
	if err := specCmd.Args(specCmd, []string{"SPEC-001"}); err != nil {
		t.Fatalf("expected one arg (a direct workspace) to be accepted, got: %v", err)
	}
	if err := specCmd.Args(specCmd, []string{"SPEC-001", "extra"}); err == nil {
		t.Fatalf("expected more than one arg to be refused")
	}
}

func TestSpecCmd_HasNoAgentFlag(t *testing.T) {
	// gnomon spec itself never resolves an Agent directly — its contextual actions (runSpecAction)
	// always resolve through the persisted default or first-use chooser, never a per-invocation
	// flag, since there is no per-invocation CLI flag to read inside an interactive session.
	if specCmd.Flags().Lookup("agent") != nil {
		t.Fatalf("did not expect gnomon spec itself to register --agent")
	}
}

func TestSpecCmd_OnlyCreateRemainsAsASubcommand(t *testing.T) {
	// create has no Workflow Contract for `run` to invoke, so it keeps its own non-interactive
	// entry point. discover/define are Contract-driven Agent workflows fully representable by
	// `gnomon run <identity> [target]` and must not also exist as dedicated subcommands.
	names := map[string]bool{}
	for _, c := range specCmd.Commands() {
		names[c.Name()] = true
	}
	if !names["create"] {
		t.Fatalf("expected gnomon spec create to remain registered for non-interactive use")
	}
	for _, unwanted := range []string{"discover", "define"} {
		if names[unwanted] {
			t.Fatalf("did not expect gnomon spec %s to remain registered; use gnomon run instead", unwanted)
		}
	}
}

func TestSpecCmd_NonInteractive_RefusesRatherThanHangs(t *testing.T) {
	// go test's own stdin is never a real terminal, so RunE's own stdinIsInteractive() check must
	// refuse immediately — before ever touching a project directory — rather than entering
	// runSpecBrowser/runSpecWorkspace, which would block forever reading keystrokes that will
	// never arrive.
	if err := specCmd.RunE(specCmd, nil); err == nil {
		t.Fatalf("expected a non-interactive invocation to be refused")
	}
}
