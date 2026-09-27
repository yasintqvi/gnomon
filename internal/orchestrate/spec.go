package orchestrate

import (
	"fmt"

	"gnomon/internal/present"
	"gnomon/internal/project"
	"gnomon/internal/specs"
)

// SpecCreate performs `gnomon spec create <title>` — deterministic Draft Creation, no Agent. The
// filename slug is mechanically derived from the title itself.
func SpecCreate(root, title string) (*present.Report, error) {
	if title == "" {
		return nil, fmt.Errorf("a title is required")
	}
	return createDraftSpec(root, title, specs.Slugify(title))
}

// createDraftSpec is Draft Creation's one implementation, shared by SpecCreate and
// AcceptDiscoveryCandidate.
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
		Next:    specWorkspaceNext(l, string(id), ""),
		Detail:  []string{fmt.Sprintf("path: %s", path)},
	}, nil
}
