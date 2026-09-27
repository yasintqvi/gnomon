package orchestrate

import (
	"fmt"
	"os"
	"path/filepath"

	"gnomon/internal/approval"
	"gnomon/internal/contract"
	"gnomon/internal/facts"
	"gnomon/internal/project"
	"gnomon/internal/specs"
)

// SpecAction is one deterministically-derived lifecycle action for a Specification — available or
// not, with a reason if not. The one place action availability is computed; every caller (gnomon
// next, the workspace) reads from it rather than re-deriving lifecycle rules of its own.
//
// Only define/approve/implement/test/revoke appear here. Verification and Review are excluded:
// their Contracts declare specification_reference: none, so no Contract ties them to a
// Specification's lifecycle — they remain reachable only via `gnomon run verification|review <target>`.
type SpecAction struct {
	Verb      string // "define", "approve", "implement", "test", "revoke"
	Command   string
	Available bool
	Reason    string
}

// SpecActions returns every lifecycle action's current availability, composed entirely from
// facts.Eligible/facts.Lifecycle against the real Contract files — never a hardcoded rule of its
// own. define/implement/test use `gnomon run <identity> <id>` (no dedicated top-level command);
// approve/revoke keep their own command names, since they are Human-exclusive CLI-native
// operations with no Workflow Contract at all.
//
// The second return value lists any Contract file that failed to load — its action is simply
// absent from the first return value, not a fatal error.
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

// SpecSummary is one Specification's browser-list row: identity, display title, and lifecycle.
type SpecSummary struct {
	ID        string
	Title     string
	Lifecycle approval.Lifecycle
}

// ListSpecs returns every existing Specification's browser-row summary, in identity order — never
// a "current Specification" selection.
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

// SpecDetail is the full derived state the Specification workspace displays. Everything here is
// derived fresh from authoritative artifacts on each call — nothing is cached.
type SpecDetail struct {
	ID           string
	Title        string
	Path         string
	Content      string
	Fingerprint  string
	Lifecycle    approval.Lifecycle
	ActiveGrant  *approval.Revision // the revision currently causing Approved, if any
	Actions      []SpecAction
	Unreadable   []string
	Revisions    []approval.Revision // oldest first
	Placeholders []string            // remaining template placeholders (see specs.RemainingPlaceholders); nil if none or the template is missing/unreadable
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
	if template, tmplErr := os.ReadFile(filepath.Join(l.SpecificationsDir(), "SPEC-000-use-case-name.md")); tmplErr == nil {
		detail.Placeholders = specs.RemainingPlaceholders(template, content)
	}
	if lifecycle == approval.Approved {
		// revisions is oldest-first, from the same ordering approval.ActiveGrant uses — so its
		// last element is always L whenever Lifecycle above says Approved.
		r := revisions[len(revisions)-1]
		detail.ActiveGrant = &r
	}
	return detail, nil
}
