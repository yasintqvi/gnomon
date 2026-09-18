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

// Status performs `gnomon status` — a deterministic, CLI-native, Agent-free snapshot of current
// project/work state, per cli/COMMAND_SURFACE.md's own enumerated scope: whether the project is
// initialized; global CONTRACT_VERSION compatibility; the list of Specifications and each one's
// derived Draft/Approved state; whether the working tree has uncommitted changes; and concise
// diagnostics when something foundational is broken.
//
// Status deliberately does not become validate: it never enumerates workflow frontmatter, never
// audits Result Contracts, and never performs a structural integrity sweep — cli/COMMAND_SURFACE.md
// draws that line explicitly ("status answers a question about work, validate answers a question
// about tooling/installation integrity"). When something foundational is broken (an unsupported
// or missing CONTRACT_VERSION), Status refuses cleanly and points at `validate` rather than
// attempting its own structural audit.
//
// Status only inspects and reports — it never mutates any authoritative artifact, and it
// introduces no new persisted state of its own; every fact below is recomputed fresh, exactly as
// cli/DERIVED_FACTS.md requires.
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
		// Git is a capability Git-related guidance uses, not a prerequisite for Specification
		// state — an unexpected Git failure degrades this section, it does not discard the
		// Specification state already reported above.
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
