package orchestrate

import (
	"fmt"
	"strings"

	"gnomon/internal/facts"
	"gnomon/internal/gitutil"
	"gnomon/internal/present"
	"gnomon/internal/project"
	"gnomon/internal/specs"
)

// Status performs `gnomon status` — a deterministic snapshot of project/work state: init and
// CONTRACT_VERSION, each Specification's derived Draft/Approved state, and working-tree changes.
// Status answers a question about work; validate answers one about tooling/installation integrity
// (cli/COMMAND_SURFACE.md) — Status never audits Contracts itself, it points at `validate` instead.
// Every fact is recomputed fresh (cli/DERIVED_FACTS.md); nothing here mutates anything.
func Status(root string) (*present.Report, error) {
	l, err := project.Locate(root)
	if err != nil {
		return nil, err
	}

	v, err := l.ContractVersion()
	if err != nil {
		return nil, err
	}
	if v == "" || v != project.SupportedContractVersion {
		reason := fmt.Sprintf("project contract version %q is not supported by this CLI (supports %q)", v, project.SupportedContractVersion)
		if v == "" {
			reason = "no CONTRACT_VERSION recorded"
		}
		return &present.Report{
			Outcome: present.Failed,
			Summary: "Project is not in a state Status can trust",
			Sections: []present.Section{
				{Label: "Contract Version", Body: reason},
			},
			Next: "Run `gnomon validate` for a full structural diagnosis.",
		}, fmt.Errorf("%s", reason)
	}

	ids, err := specs.List(l.SpecificationsDir())
	if err != nil {
		return nil, err
	}

	var specLines []string
	for _, id := range ids {
		lifecycle, err := facts.Lifecycle(l, string(id))
		if err != nil {
			return nil, err
		}
		specLines = append(specLines, fmt.Sprintf("%s: %s", id, lifecycle))
	}
	specBody := "No Specifications exist yet."
	if len(specLines) > 0 {
		specBody = strings.Join(specLines, "\n")
	}

	rep := &present.Report{
		Outcome: present.Success,
		Summary: fmt.Sprintf("Initialized, contract v%s, %d Specification(s)", v, len(ids)),
	}
	rep.AddSection("Specifications", specBody)

	gitStatus, gitErr := gitutil.Inspect(root)
	switch {
	case gitErr != nil:
		// A Git failure degrades this section only — it doesn't discard state reported above.
		rep.Outcome = present.Failed
		rep.AddSection("Working Tree", fmt.Sprintf("Git state could not be inspected: %s. Specification state above is unaffected.", gitErr))
		rep.Next = "Investigate the Git error above, then retry."
		return rep, gitErr
	case !gitStatus.Available:
		rep.AddSection("Working Tree", "not a Git repository — Git Finalization is not applicable here")
	case gitStatus.Changed:
		rep.AddSection("Working Tree", "uncommitted changes present")
	default:
		rep.AddSection("Working Tree", "clean — no uncommitted changes")
	}

	return rep, nil
}
