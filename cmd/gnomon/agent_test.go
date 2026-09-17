package main

import "testing"

func TestProviderLabel_KnownProviders(t *testing.T) {
	cases := map[string]string{
		"claude": "Claude Code",
		"codex":  "Codex",
	}
	for id, want := range cases {
		if got := providerLabel(id); got != want {
			t.Fatalf("providerLabel(%q) = %q, want %q", id, got, want)
		}
	}
}

func TestProviderLabel_UnknownFallsBackToTheIdentityItself(t *testing.T) {
	if got := providerLabel("some-future-provider"); got != "some-future-provider" {
		t.Fatalf("expected an unmapped identity to fall back to itself, got %q", got)
	}
}

// TestInteractiveAgentChooser_ReturnsCanonicalValuesNotLabels confirms the items built for the
// select menu carry the canonical provider identity as their value — the Human only ever sees
// the label, but what the menu returns (and what ResolveAgent/agentconfig then persist) is
// always "claude" or "codex", never "Claude Code" or "Codex".
func TestInteractiveAgentChooser_BuildsCanonicalValuedItems(t *testing.T) {
	providers := []string{"claude", "codex"}
	items := make([]selectItem, len(providers))
	for i, p := range providers {
		items[i] = selectItem{value: p, label: providerLabel(p)}
	}
	if items[0].value != "claude" || items[0].label != "Claude Code" {
		t.Fatalf("unexpected first item: %+v", items[0])
	}
	if items[1].value != "codex" || items[1].label != "Codex" {
		t.Fatalf("unexpected second item: %+v", items[1])
	}
}
