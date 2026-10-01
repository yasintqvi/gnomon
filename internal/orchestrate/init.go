package orchestrate

import (
	"fmt"

	"gnomon/internal/present"
	"gnomon/internal/project"
)

// InitNext is the guidance after `gnomon init`: the normal path is defining a change, not
// documenting the project first.
const InitNext = "gnomon spec\n  Create a Specification for the change you want, then Define it.\n\nOptional:\n  gnomon describe   record project context the code doesn't show"

// Init performs `gnomon init` — deterministic, minimal filesystem setup, no Agent.
func Init(root string) (*present.Report, error) {
	return initProject(root, false)
}

// InitFull performs `gnomon init --full`: Init plus the optional reference and knowledge templates
// older versions created by default.
func InitFull(root string) (*present.Report, error) {
	return initProject(root, true)
}

func initProject(root string, full bool) (*present.Report, error) {
	if err := project.Materialize(root, full); err != nil {
		return nil, err
	}
	l := project.Layout{Root: root}
	v, err := l.ContractVersion()
	if err != nil {
		return nil, err
	}
	return &present.Report{
		Outcome: present.Success,
		Summary: "Gnomon project initialized",
		Next:    InitNext,
		Detail:  []string{fmt.Sprintf("root: %s", root), fmt.Sprintf("contract version: %s", v)},
	}, nil
}
