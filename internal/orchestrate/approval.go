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
		Next:    fmt.Sprintf("gnomon implement %s", specID),
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
		Detail:  []string{fmt.Sprintf("revoked by: %s", identity)},
	}, nil
}
