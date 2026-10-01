package orchestrate

import (
	"fmt"
	"strings"

	"gnomon/internal/approval"
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
		line := fmt.Sprintf("%s: %s", id, lifecycle)
		if note := lifecycleNote(l, string(id), lifecycle); note != "" {
			line += " — " + note
		}
		specLines = append(specLines, line)
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

// lifecycleNote explains a Specification's derived state using only existing approval evidence:
// whether a Draft was never approved, had its approval revoked, or changed after approval (with
// how many lines differ from the approved text), and whether its content is still only the
// unfilled template. "" when there is nothing to add.
func lifecycleNote(l project.Layout, id string, lifecycle approval.Lifecycle) string {
	content, err := specs.ReadContent(l.SpecificationsDir(), specs.Identity(id))
	if err != nil {
		return ""
	}
	empty := specs.EffectivelyEmpty(l.SpecTemplateCandidates(), content, specs.Identity(id))

	if lifecycle == approval.Approved {
		if empty {
			return "approved, but contains only the unfilled template"
		}
		return ""
	}

	var note string
	revisions, err := approval.Revisions(l.ApprovalsDir(), id)
	switch {
	case err != nil || len(revisions) == 0:
		note = "never approved"
	case revisions[len(revisions)-1].Revoked:
		note = "approval revoked"
	default:
		latest := revisions[len(revisions)-1]
		note = "changed since approval"
		if date := approvalDate(latest.ApprovedAt); date != "" {
			note += " on " + date
		}
		if latest.ContentKnown {
			note += fmt.Sprintf(" (%d line(s) changed)", changedLines(latest.Content, string(content)))
		}
	}
	if empty {
		note += "; unfilled template"
	}
	return note
}

// approvalDate returns the YYYY-MM-DD part of a grant timestamp, or "" if it has none.
func approvalDate(ts string) string {
	if len(ts) >= 10 {
		return ts[:10]
	}
	return ""
}

// changedLines approximates how many non-blank lines changed between the approved and current
// text — the larger of lines removed and lines added — as a size hint, not a diff.
func changedLines(approved, current string) int {
	count := func(s string) map[string]int {
		m := map[string]int{}
		for _, line := range strings.Split(s, "\n") {
			if t := strings.TrimSpace(line); t != "" {
				m[t]++
			}
		}
		return m
	}
	a, c := count(approved), count(current)
	removed, added := 0, 0
	for line, k := range a {
		if d := k - c[line]; d > 0 {
			removed += d
		}
	}
	for line, k := range c {
		if d := k - a[line]; d > 0 {
			added += d
		}
	}
	return max(removed, added)
}
