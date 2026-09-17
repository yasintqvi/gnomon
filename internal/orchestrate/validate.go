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

// Validate performs `gnomon validate` — a deterministic, CLI-native, Agent-free structural audit,
// per cli/COMMAND_SURFACE.md's own enumerated scope: .gnomon/ structural integrity;
// CONTRACT_VERSION compatibility; every workflow's frontmatter validity; duplicate workflow
// identities; Workflow Contract validity generally; Result Contract validity; and approval
// evidence structural validity. Nothing outside that enumerated list is checked here —
// dependency declarations, dependency target existence, and dependency-cycle detection are
// deliberately not attempted: cli/COMMAND_SURFACE.md's own scope list for validate never names
// them, and no other authoritative document assigns their runtime placement to validate either
// (see the accompanying report's documentation-gap discussion).
//
// Validate only inspects and reports — it never repairs, rewrites, or deletes any authoritative
// artifact (Specification, approval evidence, dependency declaration, workflow, or project
// knowledge), and it introduces no new persisted state of its own.
func Validate(root string) (*present.Report, error) {
	l, err := project.Locate(root)
	if err != nil {
		return nil, err
	}

	var sections []present.Section

	// 1. .gnomon/ structural integrity.
	if missing := l.MissingSections(); len(missing) > 0 {
		sections = append(sections, present.Section{
			Label: "Project Structure",
			Body:  fmt.Sprintf("missing canonical .gnomon/ subdirectories: %s", strings.Join(missing, ", ")),
		})
	}

	// 2. CONTRACT_VERSION compatibility. Reported as a finding here, not a hard refusal — unlike
	// the real workflow-execution path (prepareWorkflow), which correctly refuses to trust an
	// unfamiliar Contract shape enough to launch a real Agent, validate's whole purpose is
	// diagnostic: a Human troubleshooting an unsupported version still benefits from seeing every
	// other structural finding in the same pass, not being stopped after the first one.
	v, err := l.ContractVersion()
	if err != nil {
		return nil, err
	}
	switch {
	case v == "":
		sections = append(sections, present.Section{Label: "Contract Version", Body: "no CONTRACT_VERSION recorded"})
	case v != project.SupportedContractVersion:
		sections = append(sections, present.Section{
			Label: "Contract Version",
			Body:  fmt.Sprintf("project contract version %q is not supported by this CLI (supports %q)", v, project.SupportedContractVersion),
		})
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
	}

	// 7. Approval evidence structural validity.
	evIssues, err := approval.ValidateEvidence(l.ApprovalsDir())
	if err != nil {
		return nil, err
	}
	if len(evIssues) > 0 {
		var lines []string
		for _, i := range evIssues {
			subject := i.SpecID
			if i.Filename != "" {
				subject = i.SpecID + "/" + i.Filename
			}
			lines = append(lines, fmt.Sprintf("%s: %s", subject, i.Problem))
		}
		sections = append(sections, present.Section{Label: "Approval Evidence", Body: strings.Join(lines, "\n")})
	}

	if len(sections) == 0 {
		return &present.Report{
			Outcome: present.Success,
			Summary: "Project structure and Contracts are valid",
		}, nil
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

// validateWorkflowContracts loads every workflow file's Contract (exactly contract.Load — the
// same parser and the same Validate() every real workflow invocation already depends on, so
// validate can never accept something a real invocation would reject, or vice versa) and reports
// any that fail to load, plus any identity declared by more than one file — per
// cli/WORKFLOW_CONTRACT.md's own validation table ("Duplicate identity across two files... the
// CLI refuses to route to either until resolved").
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
