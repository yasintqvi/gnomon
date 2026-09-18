package orchestrate

import (
	"fmt"
	"strings"

	"gnomon/internal/gitutil"
	"gnomon/internal/present"
	"gnomon/internal/project"
)

// Next performs `gnomon next` — cross-session, coarse, repository-derived guidance only, per
// cli/COMMAND_SURFACE.md. It never claims to remember a specific past Agent result or any notion
// of workflow progress, because none is ever stored, and it never picks a single "the next
// action": every existing Specification is listed, each annotated with its own currently valid
// actions (via SpecActions, the same derivation the Specification workspace uses), in identity
// order — a complete, honest enumeration, never a selection among candidates.
//
// Next never invokes, gates on, or reasons about an Agent: it never accepts --agent, never calls
// ResolveAgent, and never touches the Result Protocol. It is read-only, exactly like Status and
// Validate.
func Next(root string) (*present.Report, error) {
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
			Summary: "Project is not in a state Next can trust",
			Sections: []present.Section{
				{Label: "Contract Version", Body: reason},
			},
			Next: "Run `gnomon validate` for a full structural diagnosis.",
		}, fmt.Errorf("%s", reason)
	}

	summaries, err := ListSpecs(l)
	if err != nil {
		return nil, err
	}

	rep := &present.Report{Outcome: present.Success}
	var unreadableAll map[string]bool

	if len(summaries) == 0 {
		rep.Summary = "No Specifications exist yet"
		rep.AddSection("Specifications", "None exist yet. `gnomon spec discover` or `gnomon spec create \"<title>\"` can create one.")
	} else {
		rep.Summary = fmt.Sprintf("Valid actions for %d Specification(s), by identity — no priority implied", len(summaries))
		unreadableAll = map[string]bool{}
		for _, s := range summaries {
			actions, unreadable, err := SpecActions(l, s.ID)
			if err != nil {
				return nil, err
			}
			for _, f := range unreadable {
				unreadableAll[f] = true
			}

			var valid []string
			for _, a := range actions {
				if a.Available {
					valid = append(valid, fmt.Sprintf("`%s`", a.Command))
				}
			}

			body := fmt.Sprintf("%s.", s.Lifecycle)
			if len(valid) > 0 {
				body = fmt.Sprintf("%s. Valid: %s.", s.Lifecycle, strings.Join(valid, ", "))
			}
			rep.AddSection(s.ID, body)
		}
	}

	if len(unreadableAll) > 0 {
		var files []string
		for f := range unreadableAll {
			files = append(files, f)
		}
		rep.AddSection("Unreadable Workflow Contracts", fmt.Sprintf(
			"%s could not be loaded, so any action they would gate is omitted above. Run `gnomon validate` for details.",
			strings.Join(files, ", "),
		))
	}

	gitStatus, gitErr := gitutil.Inspect(root)
	switch {
	case gitErr != nil:
		// Git is a capability Git-related guidance uses, not a prerequisite for Specification
		// guidance — an unexpected Git failure degrades this section, it does not discard the
		// Specification guidance already computed above.
		rep.Outcome = present.Failed
		rep.AddSection("Working Tree", fmt.Sprintf("Git state could not be inspected: %s. Specification guidance above is unaffected.", gitErr))
		rep.Next = "Investigate the Git error above, then retry."
		return rep, gitErr
	case !gitStatus.Available:
		rep.AddSection("Working Tree", "Not a Git repository — Git Finalization is not applicable here.")
	case gitStatus.Changed:
		rep.AddSection("Working Tree", "Uncommitted changes are present — `gnomon finalize` may be worth considering when you judge the work ready.")
	}

	return rep, nil
}
