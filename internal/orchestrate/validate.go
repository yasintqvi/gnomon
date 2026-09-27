package orchestrate

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gnomon/internal/approval"
	"gnomon/internal/contract"
	"gnomon/internal/present"
	"gnomon/internal/project"
)

// Validate performs `gnomon validate` — a deterministic, CLI-native, Agent-free structural audit
// of .gnomon/ (cli/COMMAND_SURFACE.md's scope list: structure, CONTRACT_VERSION, workflow/Result
// Contract validity, duplicate identities, approval evidence). Dependency declarations/cycles are
// out of scope. Validate only inspects and reports — never repairs, rewrites, or deletes anything.
func Validate(root string) (*present.Report, error) {
	l, err := project.Locate(root)
	if err != nil {
		return nil, err
	}

	var sections []present.Section
	hasError := false

	// 1. .gnomon/ structural integrity.
	if missing := l.MissingSections(); len(missing) > 0 {
		sections = append(sections, present.Section{
			Label: "Project Structure",
			Body:  fmt.Sprintf("missing canonical .gnomon/ subdirectories: %s", strings.Join(missing, ", ")),
		})
		hasError = true
	}

	// 2. CONTRACT_VERSION compatibility — reported as a finding, not a hard refusal (unlike
	// prepareWorkflow's real execution path), so a Human sees every finding in one pass.
	v, err := l.ContractVersion()
	if err != nil {
		return nil, err
	}
	switch {
	case v == "":
		sections = append(sections, present.Section{Label: "Contract Version", Body: "no CONTRACT_VERSION recorded"})
		hasError = true
	case v != project.SupportedContractVersion:
		sections = append(sections, present.Section{
			Label: "Contract Version",
			Body:  fmt.Sprintf("project contract version %q is not supported by this CLI (supports %q)", v, project.SupportedContractVersion),
		})
		hasError = true
	}

	// 3–6. Workflow Contract validity, Result Contract validity, and duplicate identities.
	if wfIssues, err := validateWorkflowContracts(l); err != nil {
		return nil, err
	} else if len(wfIssues) > 0 {
		var lines []string
		for _, i := range wfIssues {
			lines = append(lines, fmt.Sprintf("%s: %s", i.subject, i.problem))
		}
		sections = append(sections, present.Section{Label: "Workflow Contracts", Body: strings.Join(lines, "\n")})
		hasError = true
	}

	// 7. Approval evidence: errors exclude a record from derivation and fail validate; warnings are
	// reported but don't (cli/APPROVAL_RUNTIME.md).
	evIssues, err := approval.ValidateEvidence(l.ApprovalsDir(), l.SpecificationsDir())
	if err != nil {
		return nil, err
	}
	var errLines, warnLines []string
	for _, i := range evIssues {
		subject := i.SpecID
		if i.Filename != "" {
			subject = i.SpecID + "/" + i.Filename
		}
		line := fmt.Sprintf("%s: %s", subject, i.Problem)
		if i.Level == approval.LevelError {
			errLines = append(errLines, line)
		} else {
			warnLines = append(warnLines, line)
		}
	}
	if len(errLines) > 0 {
		sections = append(sections, present.Section{Label: "Approval Evidence Errors", Body: strings.Join(errLines, "\n")})
		hasError = true
	}
	if len(warnLines) > 0 {
		sections = append(sections, present.Section{Label: "Approval Evidence Warnings", Body: strings.Join(warnLines, "\n")})
	}

	if !hasError {
		summary := "Project structure and Contracts are valid"
		if len(sections) > 0 {
			summary = fmt.Sprintf("Project structure and Contracts are valid, with %d warning(s)", len(warnLines))
		}
		return &present.Report{Outcome: present.Success, Summary: summary, Sections: sections}, nil
	}
	return &present.Report{
		Outcome:  present.Failed,
		Summary:  fmt.Sprintf("Validation found problems in %d area(s)", len(sections)),
		Sections: sections,
	}, fmt.Errorf("validation found problems in %d area(s)", len(sections))
}

// workflowContractIssue is one deterministic problem found while validating a single workflow
// file's frontmatter, or a duplicate-identity conflict found across the whole set.
type workflowContractIssue struct {
	subject string // filename, or an identity for a duplicate-identity finding
	problem string
}

// validateWorkflowContracts loads every workflow file via contract.Load (the same parser real
// invocations use, so validate can't accept what a real run would reject) and reports load
// failures plus any identity declared by more than one file (cli/WORKFLOW_CONTRACT.md).
func validateWorkflowContracts(l project.Layout) ([]workflowContractIssue, error) {
	entries, err := os.ReadDir(l.WorkflowsDir())
	if err != nil {
		return nil, err
	}

	var issues []workflowContractIssue
	byIdentity := map[string][]string{}

	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".md" {
			continue
		}
		wf, err := contract.Load(filepath.Join(l.WorkflowsDir(), e.Name()))
		if err != nil {
			issues = append(issues, workflowContractIssue{subject: e.Name(), problem: err.Error()})
			continue
		}
		byIdentity[wf.Identity] = append(byIdentity[wf.Identity], e.Name())
	}

	for identity, files := range byIdentity {
		if len(files) > 1 {
			issues = append(issues, workflowContractIssue{
				subject: identity,
				problem: fmt.Sprintf("duplicate identity declared by: %s", strings.Join(files, ", ")),
			})
		}
	}

	return issues, nil
}
