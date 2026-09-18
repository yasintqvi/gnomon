package orchestrate

import (
	"fmt"

	"gnomon/internal/present"
	"gnomon/internal/project"
)

// Init performs `gnomon init` — deterministic filesystem materialization only, no Agent.
func Init(root string) (*present.Report, error) {
	if err := project.Materialize(root); err != nil {
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
		Next:    "gnomon describe\n  Establish or reconcile project knowledge before starting work.",
		Detail:  []string{fmt.Sprintf("root: %s", root), fmt.Sprintf("contract version: %s", v)},
	}, nil
}
