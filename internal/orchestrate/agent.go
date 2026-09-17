package orchestrate

import (
	"fmt"
	"strings"

	"gnomon/internal/adapter"
	"gnomon/internal/agentconfig"
	"gnomon/internal/present"
)

// AgentChooser interactively asks the Human to choose one of the given Agent providers, returning
// their choice. Wired to a real terminal prompt only at the CLI edge (cmd/gnomon), exactly like
// approval.go's PromptFunc for Human-identity fallback — nil means interaction is unavailable,
// never "prompt silently" or "prompt with no visible output."
type AgentChooser func(providers []string) (string, error)

// ResolveAgent determines which Agent provider a workflow invocation should use and constructs
// its Adapter. This is the one function every workflow-invoking command goes through — none of
// them ever constructs a ClaudeAdapter or CodexAdapter directly; only adapter.Resolve does that.
//
// Precedence: an explicit per-invocation override (the CLI's --agent flag) always wins and is
// never persisted. Otherwise the persisted user-level default is used, if one exists. If none
// exists yet and chooser is non-nil (the CLI determined stdin is interactive), the Human is asked
// once and the choice is persisted as the new default — but only after it validates, so an
// invalid answer is never written. If none exists and chooser is nil, this refuses rather than
// guessing: a non-interactive invocation must never silently pick a provider on the Human's
// behalf.
func ResolveAgent(override string, chooser AgentChooser) (adapter.Adapter, error) {
	if override != "" {
		return adapter.Resolve(override)
	}

	cfg, err := agentconfig.Load()
	if err != nil {
		return nil, err
	}
	if cfg.DefaultProvider != "" {
		return adapter.Resolve(cfg.DefaultProvider)
	}

	if chooser == nil {
		return nil, fmt.Errorf(
			"no default Agent provider is configured; pass --agent (%s), or run `gnomon agent set-default <provider>`",
			strings.Join(adapter.Providers(), "|"),
		)
	}

	choice, err := chooser(adapter.Providers())
	if err != nil {
		return nil, err
	}
	ad, err := adapter.Resolve(choice)
	if err != nil {
		return nil, err // an invalid choice is never persisted
	}
	if err := agentconfig.Save(agentconfig.Config{DefaultProvider: choice}); err != nil {
		return nil, err
	}
	return ad, nil
}

// AgentStatus performs `gnomon agent` — reports the current persisted default Agent provider (or
// its absence) and every available provider. This requires no Gnomon project: Agent provider
// identity is a Human/installation-level concern, never a project-level one, so it works from any
// directory.
func AgentStatus() (*present.Report, error) {
	cfg, err := agentconfig.Load()
	if err != nil {
		return nil, err
	}
	providers := adapter.Providers()

	summary := "No default Agent provider is configured"
	next := "Run `gnomon agent set-default <provider>`, or pass --agent on any Agent-invoking command."
	if cfg.DefaultProvider != "" {
		summary = fmt.Sprintf("Default Agent provider: %s", cfg.DefaultProvider)
		next = ""
	}

	return &present.Report{
		Outcome: present.Success,
		Summary: summary,
		Sections: []present.Section{
			{Label: "Available", Body: strings.Join(providers, "\n")},
		},
		Next: next,
	}, nil
}

// AgentSetDefault performs `gnomon agent set-default <provider>` — the only deterministic,
// explicitly Human-invoked way to change the persisted default. A per-invocation --agent override
// never reaches this function and never changes what is persisted here.
func AgentSetDefault(provider string) (*present.Report, error) {
	if _, err := adapter.Resolve(provider); err != nil {
		return nil, err
	}
	if err := agentconfig.Save(agentconfig.Config{DefaultProvider: provider}); err != nil {
		return nil, err
	}
	return &present.Report{
		Outcome: present.Success,
		Summary: fmt.Sprintf("Default Agent provider set to %s", provider),
	}, nil
}
