package orchestrate

import (
	"fmt"
	"strings"

	"gnomon/internal/approval"
	"gnomon/internal/facts"
	"gnomon/internal/gitutil"
	"gnomon/internal/present"
	"gnomon/internal/project"
	"gnomon/internal/specs"
)

// PromptFunc requests one-time Human attribution when no Git identity is configured — injected
// so orchestration stays testable without real stdin, per Step 6's fallback rule.
type PromptFunc func(message string) (string, error)

// specWorkspaceNext builds the standard next-step guidance shown after an operation that changed
// a Specification's lifecycle state: the Human-facing workspace (cli/COMMAND_SURFACE.md's
// everyday entry point for Specification work), which shows every action the Specification's new
// state actually makes available — never a dedicated per-workflow command, since none exists for
// implement/test/define anymore (cli/COMMAND_SURFACE.md's "run Rule"). This is the one place that
// guidance is built, so every caller renders it identically.
//
// When highlightVerb is non-empty, the equivalent explicit `gnomon run <identity> <SPEC-id>`
// invocation is also shown — but only when SpecActions, re-derived fresh here from the same
// facts.Eligible every other caller (gnomon next, the workspace) already uses, currently reports
// that verb Available. This deliberately never hardcodes a lifecycle assumption of its own (e.g.
// "Approved unlocks Implementation") — the caller names which action would be worth highlighting
// if it turns out to be available, and this function asks the authoritative source rather than
// assuming; if SpecActions disagrees, or fails to load, or the verb isn't found at all, only the
// workspace pointer is shown. That refusal-to-guess is deliberate, not a bug: recommendation code
// must never become a second, independent source of lifecycle truth.
func specWorkspaceNext(l project.Layout, specID, highlightVerb string) string {
	next := fmt.Sprintf("gnomon spec %s\n  View the Specification and available actions.", specID)
	if highlightVerb == "" {
		return next
	}
	actions, _, err := SpecActions(l, specID)
	if err != nil {
		return next
	}
	for _, a := range actions {
		if a.Verb == highlightVerb && a.Available {
			return next + fmt.Sprintf("\n\nAdvanced:\n  %s", a.Command)
		}
	}
	return next
}

// Approve performs `gnomon approve <SPEC-id>` — a deterministic, Human-owned operation. An
// Agent's own reported result never grants approval; only this explicit action does.
func Approve(root, specID string, prompt PromptFunc) (*present.Report, error) {
	l, err := project.Locate(root)
	if err != nil {
		return nil, err
	}

	ok, _, err := specs.Exists(l.SpecificationsDir(), specs.Identity(specID))
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("%s does not exist", specID)
	}

	content, err := specs.ReadContent(l.SpecificationsDir(), specs.Identity(specID))
	if err != nil {
		return nil, err
	}
	fingerprint := approval.Fingerprint(content)

	identity, err := gitutil.Identity(root)
	if err != nil {
		if prompt == nil {
			return nil, fmt.Errorf("no Human identity available for attribution; refusing to write unattributed approval evidence")
		}
		answer, promptErr := prompt("No Git identity is configured. Enter your name for this approval: ")
		if promptErr != nil || strings.TrimSpace(answer) == "" {
			return nil, fmt.Errorf("no Human identity available for attribution; refusing to write unattributed approval evidence")
		}
		identity = strings.TrimSpace(answer)
	}

	if err := approval.WriteGrant(l.ApprovalsDir(), specID, fingerprint, identity, content); err != nil {
		return nil, err
	}

	state, err := facts.Lifecycle(l, specID)
	if err != nil {
		return nil, err
	}

	return &present.Report{
		Outcome: present.Success,
		Summary: fmt.Sprintf("%s is now %s", specID, state),
		Target:  specID,
		Next:    specWorkspaceNext(l, specID, "implement"),
		Detail:  []string{fmt.Sprintf("approved by: %s", identity)},
	}, nil
}

// Revoke performs `gnomon revoke <SPEC-id>` — a deterministic, Human-owned operation, CLI-native
// and Agent-free, per cli/APPROVAL_RUNTIME.md's Revocation Operation: locate the grant record
// currently causing this Specification to derive Approved, if any; confirm explicit Human action
// and attribution; write one new revocation record referencing that specific grant (the grant
// file itself is never touched); re-derive lifecycle state and report it. If no grant is
// currently causing Approved — the Specification is already Draft, was never approved, its only
// grant no longer matches current content, or it has already been revoked — there is nothing to
// revoke, per the Operation's own "if any" and Core's "a revoked record can never revalidate the
// Specification" rule; this refuses rather than writing a Revocation record that references
// nothing currently active.
func Revoke(root, specID string, prompt PromptFunc) (*present.Report, error) {
	l, err := project.Locate(root)
	if err != nil {
		return nil, err
	}

	ok, _, err := specs.Exists(l.SpecificationsDir(), specs.Identity(specID))
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("%s does not exist", specID)
	}

	content, err := specs.ReadContent(l.SpecificationsDir(), specs.Identity(specID))
	if err != nil {
		return nil, err
	}
	fingerprint := approval.Fingerprint(content)

	grantID, active, err := approval.ActiveGrantID(l.ApprovalsDir(), specID, fingerprint)
	if err != nil {
		return nil, err
	}
	if !active {
		return &present.Report{
			Outcome: present.Blocked,
			Summary: fmt.Sprintf("%s has no active approval to revoke", specID),
			Target:  specID,
		}, fmt.Errorf("%s has no active approval to revoke", specID)
	}

	identity, err := gitutil.Identity(root)
	if err != nil {
		if prompt == nil {
			return nil, fmt.Errorf("no Human identity available for attribution; refusing to write unattributed revocation evidence")
		}
		answer, promptErr := prompt("No Git identity is configured. Enter your name for this revocation: ")
		if promptErr != nil || strings.TrimSpace(answer) == "" {
			return nil, fmt.Errorf("no Human identity available for attribution; refusing to write unattributed revocation evidence")
		}
		identity = strings.TrimSpace(answer)
	}

	if err := approval.WriteRevocation(l.ApprovalsDir(), specID, grantID, identity); err != nil {
		return nil, err
	}

	state, err := facts.Lifecycle(l, specID)
	if err != nil {
		return nil, err
	}

	return &present.Report{
		Outcome: present.Success,
		Summary: fmt.Sprintf("%s is now %s", specID, state),
		Target:  specID,
		// No single action is more worth highlighting than another after returning to Draft —
		// the workspace itself already shows whatever SpecActions currently reports valid.
		Next:   specWorkspaceNext(l, specID, ""),
		Detail: []string{fmt.Sprintf("revoked by: %s", identity)},
	}, nil
}
