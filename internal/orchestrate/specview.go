package orchestrate

import (
	"fmt"
	"path/filepath"

	"gnomon/internal/approval"
	"gnomon/internal/contract"
	"gnomon/internal/facts"
	"gnomon/internal/project"
	"gnomon/internal/specs"
)

// SpecAction is one deterministically-derived lifecycle action for a Specification — available
// or not, with a reason in the latter case. This is the one place action availability is
// computed; every caller (gnomon next, the Specification workspace) reads from it rather than
// re-deriving lifecycle rules of its own.
//
// Only define/approve/implement/test/revoke ever appear here. Verification and Review are
// deliberately excluded: their Contracts declare specification_reference: none — Core's own
// statement that their target "is not necessarily a Specification at all" — so unlike
// Definition/Implementation/Testing, no Contract establishes any relationship between them and a
// particular Specification's lifecycle. Listing them here would assert a Specification-scoped
// meaning the authoritative model does not define; they remain reachable, generically, only via
// `gnomon run verification|review <target>`.
type SpecAction struct {
	Verb      string // "define", "approve", "implement", "test", "revoke"
	Command   string
	Available bool
	Reason    string
}

// SpecActions returns every lifecycle action's current availability for one existing
// Specification, composed entirely from facts.Eligible/facts.Lifecycle against the real
// Contract files — never a hardcoded rule of its own. define/implement/test come from the
// Specification-governed workflow Contracts, by their own declared identity (never a hardcoded
// "spec define" style guess); approve/revoke come from facts.Lifecycle's own Draft/Approved
// derivation (the same operations Approve/Revoke themselves gate on).
//
// Command names gnomon run <identity> <id> for define/implement/test: these three have no
// dedicated top-level command (cli/COMMAND_SURFACE.md's public surface centers on gnomon spec /
// gnomon spec <SPEC-id>, which call the same underlying Core operations directly; run is the
// documented non-interactive path for the same operation). approve/revoke keep their own
// dedicated command names, since they are deterministic, Human-exclusive CLI-native operations
// with no Workflow Contract at all — run has nothing to invoke for them.
//
// The second return value lists any of the three Contract files that failed to load — the
// actions they would have gated are simply absent from the first return value, matching
// resolveWorkflowByIdentity's own "unreadable Contract is excluded, not fatal" treatment.
func SpecActions(l project.Layout, id string) ([]SpecAction, []string, error) {
	lifecycle, err := facts.Lifecycle(l, id)
	if err != nil {
		return nil, nil, err
	}

	defineWf, defineErr := contract.Load(filepath.Join(l.WorkflowsDir(), "specification-definition.md"))
	implWf, implErr := contract.Load(filepath.Join(l.WorkflowsDir(), "implementation.md"))
	testWf, testErr := contract.Load(filepath.Join(l.WorkflowsDir(), "testing.md"))

	var unreadable []string
	var actions []SpecAction

	addFromContract := func(verb, filename string, wf contract.Workflow, loadErr error) error {
		if loadErr != nil {
			unreadable = append(unreadable, filename)
			return nil
		}
		elig, err := facts.Eligible(l, wf, id)
		if err != nil {
			return err
		}
		actions = append(actions, SpecAction{
			Verb:      verb,
			Command:   fmt.Sprintf("gnomon run %s %s", wf.Identity, id),
			Available: elig.Eligible,
			Reason:    elig.Reason,
		})
		return nil
	}

	if err := addFromContract("define", "specification-definition.md", defineWf, defineErr); err != nil {
		return nil, nil, err
	}

	approveAction := SpecAction{Verb: "approve", Command: fmt.Sprintf("gnomon approve %s", id)}
	revokeAction := SpecAction{Verb: "revoke", Command: fmt.Sprintf("gnomon revoke %s", id)}
	if lifecycle == approval.Draft {
		approveAction.Available = true
		revokeAction.Reason = fmt.Sprintf("%s is not Approved (currently Draft)", id)
	} else {
		revokeAction.Available = true
		approveAction.Reason = fmt.Sprintf("%s is already Approved for its current content", id)
	}
	actions = append(actions, approveAction)

	if err := addFromContract("implement", "implementation.md", implWf, implErr); err != nil {
		return nil, nil, err
	}
	if err := addFromContract("test", "testing.md", testWf, testErr); err != nil {
		return nil, nil, err
	}
	actions = append(actions, revokeAction)

	return actions, unreadable, nil
}

// SpecSummary is one Specification's browser-list row: identity, display title, and lifecycle —
// the minimum needed to list and filter Specifications without opening each one's full workspace.
type SpecSummary struct {
	ID        string
	Title     string
	Lifecycle approval.Lifecycle
}

// ListSpecs returns every existing Specification's browser-row summary, in identity order —
// the same complete, unranked enumeration facts.Lifecycle-consuming callers throughout this
// package already use (Next, Status); never a "current Specification" selection.
func ListSpecs(l project.Layout) ([]SpecSummary, error) {
	ids, err := specs.List(l.SpecificationsDir())
	if err != nil {
		return nil, err
	}
	summaries := make([]SpecSummary, 0, len(ids))
	for _, id := range ids {
		content, err := specs.ReadContent(l.SpecificationsDir(), id)
		if err != nil {
			return nil, err
		}
		lifecycle, err := facts.Lifecycle(l, string(id))
		if err != nil {
			return nil, err
		}
		summaries = append(summaries, SpecSummary{
			ID:        string(id),
			Title:     specs.Title(id, content),
			Lifecycle: lifecycle,
		})
	}
	return summaries, nil
}

// SpecDetail is the full derived state the Specification workspace displays for one
// Specification: current identity/title/lifecycle/fingerprint, the active grant when Approved,
// every action's current availability, and the durable revision history. Everything here is
// derived fresh from authoritative artifacts on each call — nothing is cached or remembered
// across invocations.
type SpecDetail struct {
	ID          string
	Title       string
	Path        string
	Content     string
	Fingerprint string
	Lifecycle   approval.Lifecycle
	ActiveGrant *approval.Revision // the revision currently causing Approved, if any
	Actions     []SpecAction
	Unreadable  []string
	Revisions   []approval.Revision // oldest first
}

// ListSpecsForRoot locates the project at root and returns ListSpecs's result — the entry point
// CLI commands (which know only a filesystem root, not an already-located Layout) call.
func ListSpecsForRoot(root string) ([]SpecSummary, error) {
	l, err := project.Locate(root)
	if err != nil {
		return nil, err
	}
	return ListSpecs(l)
}

// SpecDetailForRoot locates the project at root and returns SpecDetailFor's result — the entry
// point CLI commands call.
func SpecDetailForRoot(root, id string) (*SpecDetail, error) {
	l, err := project.Locate(root)
	if err != nil {
		return nil, err
	}
	return SpecDetailFor(l, id)
}

// SpecDetailFor derives the full workspace view for one existing Specification.
func SpecDetailFor(l project.Layout, id string) (*SpecDetail, error) {
	ok, path, err := specs.Exists(l.SpecificationsDir(), specs.Identity(id))
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("%s does not exist", id)
	}
	content, err := specs.ReadContent(l.SpecificationsDir(), specs.Identity(id))
	if err != nil {
		return nil, err
	}
	fingerprint := approval.Fingerprint(content)
	lifecycle, err := facts.Lifecycle(l, id)
	if err != nil {
		return nil, err
	}
	actions, unreadable, err := SpecActions(l, id)
	if err != nil {
		return nil, err
	}
	revisions, err := approval.Revisions(l.ApprovalsDir(), id)
	if err != nil {
		return nil, err
	}

	detail := &SpecDetail{
		ID:          id,
		Title:       specs.Title(specs.Identity(id), content),
		Path:        path,
		Content:     string(content),
		Fingerprint: fingerprint,
		Lifecycle:   lifecycle,
		Actions:     actions,
		Unreadable:  unreadable,
		Revisions:   revisions,
	}
	if lifecycle == approval.Approved {
		// revisions is oldest-first; the latest non-revoked match is the one actually causing
		// Approved (matching approval.activeGrant's own rule, applied here for display).
		for i := len(revisions) - 1; i >= 0; i-- {
			if revisions[i].Fingerprint == fingerprint && !revisions[i].Revoked {
				r := revisions[i]
				detail.ActiveGrant = &r
				break
			}
		}
	}
	return detail, nil
}
