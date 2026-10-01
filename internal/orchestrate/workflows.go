package orchestrate

import (
	"fmt"
	"strings"

	"gnomon/internal/contract"
	"gnomon/internal/present"
	"gnomon/internal/project"
)

// Workflows performs `gnomon workflows`: every workflow in effect and where it comes from.
func Workflows(root string) (*present.Report, error) {
	l, err := project.Locate(root)
	if err != nil {
		return nil, err
	}
	all, stale, err := l.Workflows()
	if err != nil {
		return nil, err
	}
	var lines []string
	for _, f := range all {
		identity := strings.TrimSuffix(f.Name, ".md")
		if wf, err := contract.Parse(f.Data, f.Name); err == nil {
			identity = wf.Identity
		}
		source := "bundled"
		if f.Customized() {
			source = "customized: " + relToRoot(l, f.ProjectPath)
		}
		lines = append(lines, fmt.Sprintf("%-34s %s", identity, source))
	}
	rep := &present.Report{
		Outcome:  present.Success,
		Summary:  fmt.Sprintf("%d workflow(s) in effect", len(all)),
		Sections: []present.Section{{Label: "Workflows", Body: strings.Join(lines, "\n")}},
		Next:     "gnomon workflows customize <file>\n  Copy a bundled workflow into .gnomon/workflows/ to change it for this project.",
	}
	if len(stale) > 0 {
		var paths []string
		for _, p := range stale {
			paths = append(paths, relToRoot(l, p))
		}
		rep.AddSection("Ignored", "Unmodified copies from an older Gnomon (the current bundled versions are used; these can be deleted):\n"+strings.Join(paths, "\n"))
	}
	return rep, nil
}

// CustomizeWorkflow performs `gnomon workflows customize <file>`.
func CustomizeWorkflow(root, name string) (*present.Report, error) {
	l, err := project.Locate(root)
	if err != nil {
		return nil, err
	}
	if !strings.HasSuffix(name, ".md") {
		name += ".md"
	}
	dest, err := l.CustomizeWorkflow(name)
	if err != nil {
		return nil, err
	}
	return &present.Report{
		Outcome: present.Success,
		Summary: fmt.Sprintf("Copied %s for customization", name),
		Sections: []present.Section{{
			Label: "Note",
			Body:  "Gnomon uses this file instead of the bundled version once you change it. While it is identical to a shipped version, the bundled version keeps being used.",
		}},
		Detail: []string{"path: " + relToRoot(l, dest)},
	}, nil
}
