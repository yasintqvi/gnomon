package adapter

import "testing"

func TestProviders_ListsClaudeAndCodex(t *testing.T) {
	got := Providers()
	want := []string{"claude", "codex"}
	if len(got) != len(want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("expected %v, got %v", want, got)
		}
	}
}

func TestProviders_ReturnsAnIndependentCopy(t *testing.T) {
	got := Providers()
	got[0] = "mutated"
	fresh := Providers()
	if fresh[0] == "mutated" {
		t.Fatalf("expected Providers() to return an independent copy on each call")
	}
}

func TestResolve_Claude(t *testing.T) {
	ad, err := Resolve("claude")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := ad.(*ClaudeAdapter); !ok {
		t.Fatalf("expected *ClaudeAdapter, got %T", ad)
	}
}

func TestResolve_Codex(t *testing.T) {
	ad, err := Resolve("codex")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := ad.(*CodexAdapter); !ok {
		t.Fatalf("expected *CodexAdapter, got %T", ad)
	}
}

func TestResolve_UnknownProviderRefuses(t *testing.T) {
	if _, err := Resolve("gpt5-agent"); err == nil {
		t.Fatalf("expected an unrecognized provider name to be refused")
	}
}

// TestResolve_BothProvidersSatisfyAdapter confirms adapter.Adapter itself needed no change: both
// concrete types satisfy it exactly as ClaudeAdapter alone did before Codex existed.
func TestResolve_BothProvidersSatisfyAdapter(t *testing.T) {
	for _, name := range Providers() {
		ad, err := Resolve(name)
		if err != nil {
			t.Fatalf("Resolve(%q): %v", name, err)
		}
		var _ Adapter = ad
	}
}
