package orchestrate

import (
	"fmt"

	"gnomon/internal/present"
	"gnomon/internal/project"
	"gnomon/internal/specs"
)

// SpecCreate performs `gnomon spec create <title>` — deterministic Draft Creation, no Agent.
func SpecCreate(root, title string) (*present.Report, error) {
	if title == "" {
		return nil, fmt.Errorf("a title is required")
	}
	l, err := project.Locate(root)
	if err != nil {
		return nil, err
	}
	id, err := specs.NextIdentity(l.SpecificationsDir())
	if err != nil {
		return nil, err
	}
	path, err := specs.CreateDraft(l.SpecificationsDir(), id, title)
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
