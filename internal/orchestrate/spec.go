package orchestrate

import (
	"fmt"

	"gnomon/internal/present"
	"gnomon/internal/project"
	"gnomon/internal/specs"
)

// SpecCreate performs `gnomon spec create <title>` — deterministic Draft Creation, no Agent. The
// filename slug is mechanically derived from the title itself, since a direct Human-supplied
// title has no separate semantic identity (unlike a Discovery candidate — see
// createDraftSpec/AcceptDiscoveryCandidate).
func SpecCreate(root, title string) (*present.Report, error) {
	if title == "" {
		return nil, fmt.Errorf("a title is required")
	}
	return createDraftSpec(root, title, specs.Slugify(title))
}

// createDraftSpec is Draft Creation's one deterministic implementation, shared by SpecCreate
// (slug mechanically derived from the title) and AcceptDiscoveryCandidate (slug derived from the
// candidate's own semantic identity field) — never duplicated between the two entry points.
func createDraftSpec(root, title, slug string) (*present.Report, error) {
	l, err := project.Locate(root)
	if err != nil {
		return nil, err
	}
	id, err := specs.NextIdentity(l.SpecificationsDir())
	if err != nil {
		return nil, err
	}
	path, err := specs.CreateDraft(l.SpecificationsDir(), id, title, slug)
	if err != nil {
		return nil, err
	}
	return &present.Report{
		Outcome: present.Success,
		Summary: "Created Draft Specification",
		Target:  string(id),
		Next:    fmt.Sprintf("Define its content, then `gnomon approve %s` once you judge it ready.", id),
		Detail:  []string{fmt.Sprintf("path: %s", path)},
	}, nil
}
