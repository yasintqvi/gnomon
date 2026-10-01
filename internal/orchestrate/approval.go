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

// ConfirmFunc presents a non-blocking warning to the Human and reports whether to proceed anyway.
// The command layer owns how (and whether) this actually prompts, so orchestration itself never
// touches a terminal. A nil ConfirmFunc skips the check silently, same as a missing template.
type ConfirmFunc func(warning string) (bool, error)

// placeholderWarning renders the Human-facing warning for a Specification that still contains
// template placeholders: how many, up to 5 of them, and a nudge toward Define.
func placeholderWarning(specID string, remaining []string) string {
	n := len(remaining)
	shown := remaining
	more := 0
	if n > 5 {
		shown = remaining[:5]
		more = n - 5
	}
	var b strings.Builder
	fmt.Fprintf(&b, "%s still contains %d template placeholder", specID, n)
	if n != 1 {
		b.WriteString("s")
	}
	b.WriteString(":\n")
	for _, tok := range shown {
		fmt.Fprintf(&b, "  %s\n", tok)
	}
	if more > 0 {
		fmt.Fprintf(&b, "  and %d more\n", more)
	}
	b.WriteString("Consider running Define before approving.")
	return b.String()
}

// specWorkspaceNext builds the standard next-step guidance after an operation that changed a
// Specification's lifecycle state: a pointer to the workspace (cli/COMMAND_SURFACE.md), plus, when
// highlightVerb is non-empty and SpecActions (re-derived fresh, never assumed) actually reports it
// Available, the equivalent `gnomon run` invocation. Never hardcodes a lifecycle assumption of its
// own — recommendation code must not become a second source of lifecycle truth.
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
func Approve(root, specID string, prompt PromptFunc, confirm ConfirmFunc) (*present.Report, error) {
	if err := refuseIfRunActive(root); err != nil {
		return nil, err
	}

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

	// Already Approved for this exact content under "the latest decision counts": refuse before
	// resolving identity, so a repeated approve never prompts and never writes a redundant grant.
	// Content matching an older, non-latest grant instead falls through and writes a new one —
	// that is how the Human returns to an earlier version.
	_, approved, err := approval.ActiveGrant(l.ApprovalsDir(), specID, fingerprint)
	if err != nil {
		return nil, err
	}
	if approved {
		return &present.Report{
			Outcome: present.Success,
			Summary: fmt.Sprintf("%s is already Approved", specID),
			Target:  specID,
			Next:    specWorkspaceNext(l, specID, "implement"),
		}, nil
	}

	// A Specification with nothing written beyond its template would approve no behavior at all —
	// refused outright, interactive or not. Merely unfinished content (some placeholders left) is
	// the Human's call: warned about, never blocked.
	templates := l.SpecTemplateCandidates()
	if specs.EffectivelyEmpty(templates, content, specs.Identity(specID)) {
		return &present.Report{
			Outcome: present.Blocked,
			Summary: fmt.Sprintf("%s has nothing to approve yet", specID),
			Target:  specID,
			Sections: []present.Section{{
				Label: "Unresolved",
				Body:  "It contains only the unfilled template. Write its goal and acceptance criteria, or run Define, then approve.",
			}},
			Next: specWorkspaceNext(l, specID, ""),
		}, fmt.Errorf("%s contains only the unfilled template", specID)
	}
	if confirm != nil {
		if template := specs.ClosestTemplate(templates, content, specs.Identity(specID)); template != nil {
			if remaining := specs.RemainingPlaceholders(template, content); len(remaining) > 0 {
				proceed, err := confirm(placeholderWarning(specID, remaining))
				if err != nil {
					return nil, err
				}
				if !proceed {
					return &present.Report{
						Outcome: present.Cancelled,
						Summary: fmt.Sprintf("Approval of %s was cancelled", specID),
						Target:  specID,
					}, nil
				}
			}
		}
	}

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

// Revoke performs `gnomon revoke <SPEC-id>` — a deterministic, Human-owned, Agent-free operation
// (cli/APPROVAL_RUNTIME.md, Revocation Operation). Writes one new revocation record per grant
// currently causing Approved (grant files themselves are never touched). Refuses if nothing is
// currently Approved, rather than writing a revocation that references nothing active.
func Revoke(root, specID string, prompt PromptFunc) (*present.Report, error) {
	if err := refuseIfRunActive(root); err != nil {
		return nil, err
	}

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

	// Gated on ActiveGrant (the same rule Derive uses), not a plain fingerprint match: content
	// matching only an older, non-latest grant is already Draft and has nothing to revoke.
	_, approved, err := approval.ActiveGrant(l.ApprovalsDir(), specID, fingerprint)
	if err != nil {
		return nil, err
	}
	if !approved {
		return &present.Report{
			Outcome: present.Blocked,
			Summary: fmt.Sprintf("%s has no active approval to revoke", specID),
			Target:  specID,
		}, fmt.Errorf("%s has no active approval to revoke", specID)
	}

	// The active grant (L) is always included here, since it is unrevoked and matches
	// fingerprint by construction; any legacy duplicate grant for the same content is revoked
	// alongside it, tidying evidence without changing which one L actually is.
	grantIDs, err := approval.ActiveGrantIDs(l.ApprovalsDir(), specID, fingerprint)
	if err != nil {
		return nil, err
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

	var revoked []string
	for _, grantID := range grantIDs {
		if err := approval.WriteRevocation(l.ApprovalsDir(), specID, grantID, identity); err != nil {
			sections := []present.Section{{Label: "Reason", Body: err.Error()}}
			if len(revoked) > 0 {
				sections = append(sections, present.Section{Label: "Revoked", Body: strings.Join(revoked, ", ")})
			}
			return &present.Report{
				Outcome:  present.Failed,
				Summary:  fmt.Sprintf("%s: revocation did not complete", specID),
				Target:   specID,
				Sections: sections,
			}, err
		}
		revoked = append(revoked, grantID)
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
