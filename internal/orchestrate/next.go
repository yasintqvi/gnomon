package orchestrate

import (
	"fmt"
	"path/filepath"
	"strings"

	"gnomon/internal/approval"
	"gnomon/internal/contract"
	"gnomon/internal/facts"
	"gnomon/internal/gitutil"
	"gnomon/internal/present"
	"gnomon/internal/project"
	"gnomon/internal/specs"
)

// Next performs `gnomon next` — cross-session, coarse, repository-derived guidance only, per
// cli/COMMAND_SURFACE.md: "a list of currently valid actions (Step 1 gates + Step 7 facts),
// annotated with generic suggestions. It never claims to remember a specific past Agent result or
// any notion of workflow progress, because none is ever stored." Same-session, high-confidence
// guidance (the Next field a workflow-invoking command's own Report already carries) is not this
// command's job at all.
//
// Next never picks a single "the next action". Gnomon persists no execution history, so it has no
// durable basis for choosing among multiple simultaneously-valid actions or multiple candidate
// Specifications — cli/USER_JOURNEY.md's Valid vs. Recommended Actions distinction and the
// project's own design constraints for this command require surfacing every valid action,
// unranked, rather than inventing precedence. Concretely: every existing Specification is listed,
// each annotated with its own currently valid actions, in identity order (a complete, honest
// enumeration — never a selection among candidates).
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

	// The three built-in, Specification-governed workflows whose Contracts determine per-
	// Specification valid actions. A Contract that fails to load is excluded from consideration
	// (never crashes the whole command) — the same "unreadable Contract is excluded, not fatal"
	// treatment resolveWorkflowByIdentity already applies when scanning by identity.
	defineWf, defineErr := contract.Load(filepath.Join(l.WorkflowsDir(), "specification-definition.md"))
	implWf, implErr := contract.Load(filepath.Join(l.WorkflowsDir(), "implementation.md"))
	testWf, testErr := contract.Load(filepath.Join(l.WorkflowsDir(), "testing.md"))

	var unreadable []string
	if defineErr != nil {
		unreadable = append(unreadable, "specification-definition.md")
	}
	if implErr != nil {
		unreadable = append(unreadable, "implementation.md")
	}
	if testErr != nil {
		unreadable = append(unreadable, "testing.md")
	}

	ids, err := specs.List(l.SpecificationsDir())
	if err != nil {
		return nil, err
	}

	rep := &present.Report{Outcome: present.Success}

	if len(ids) == 0 {
		rep.Summary = "No Specifications exist yet"
		rep.AddSection("Specifications", "None exist yet. `gnomon spec discover` or `gnomon spec create \"<title>\"` can create one.")
	} else {
		rep.Summary = fmt.Sprintf("Valid actions for %d Specification(s), by identity — no priority implied", len(ids))
		for _, id := range ids {
			lifecycle, err := facts.Lifecycle(l, string(id))
			if err != nil {
				return nil, err
			}

			var actions []string
			if defineErr == nil {
				if elig, err := facts.Eligible(l, defineWf, string(id)); err != nil {
					return nil, err
				} else if elig.Eligible {
					actions = append(actions, fmt.Sprintf("`gnomon spec define %s`", id))
				}
			}
			if lifecycle == approval.Draft {
				actions = append(actions, fmt.Sprintf("`gnomon approve %s`", id))
			}
			if implErr == nil {
				if elig, err := facts.Eligible(l, implWf, string(id)); err != nil {
					return nil, err
				} else if elig.Eligible {
					actions = append(actions, fmt.Sprintf("`gnomon implement %s`", id))
				}
			}
			if testErr == nil {
				if elig, err := facts.Eligible(l, testWf, string(id)); err != nil {
					return nil, err
				} else if elig.Eligible {
					actions = append(actions, fmt.Sprintf("`gnomon test %s`", id))
				}
			}
			if lifecycle == approval.Approved {
				actions = append(actions, fmt.Sprintf("`gnomon revoke %s`", id))
			}

			body := fmt.Sprintf("%s.", lifecycle)
			if len(actions) > 0 {
				body = fmt.Sprintf("%s. Valid: %s.", lifecycle, strings.Join(actions, ", "))
			}
			rep.AddSection(string(id), body)
		}
	}

	if len(unreadable) > 0 {
		rep.AddSection("Unreadable Workflow Contracts", fmt.Sprintf(
			"%s could not be loaded, so any action they would gate is omitted above. Run `gnomon validate` for details.",
			strings.Join(unreadable, ", "),
		))
	}

	changed, err := gitutil.HasChanges(root)
	if err != nil {
		return nil, err
	}
	if changed {
		rep.AddSection("Working Tree", "Uncommitted changes are present — `gnomon finalize` may be worth considering when you judge the work ready.")
	}

	return rep, nil
}
