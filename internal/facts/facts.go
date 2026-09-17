// Package facts composes Derived Facts (CLI Step 7): Specification existence, derived lifecycle,
// and a workflow's own Contract fields, into pre-start eligibility. It never invents a universal
// dependency SATISFIED/UNSATISFIED state and never performs Agent-side reasoning.
package facts

import (
	"fmt"

	"gnomon/internal/approval"
	"gnomon/internal/contract"
	"gnomon/internal/project"
	"gnomon/internal/specs"
)

// Eligibility is the composite pre-start fact a Workflow Contract gate resolves to.
type Eligibility struct {
	Eligible bool
	Reason   string
}

// Lifecycle re-derives a Specification's current Draft/Approved state — always recomputed fresh
// from current content plus current evidence, per Step 6, never cached here.
func Lifecycle(l project.Layout, specID string) (approval.Lifecycle, error) {
	content, err := specs.ReadContent(l.SpecificationsDir(), specs.Identity(specID))
	if err != nil {
		return approval.Draft, err
	}
	return approval.Derive(l.ApprovalsDir(), specID, approval.Fingerprint(content))
}

// Eligible derives whether a workflow may start for the given (optional) supplied Specification,
// per its own declared specification_reference/requires_approved_specification — the exact rule
// cli/WORKFLOW_CONTRACT.md defines, applied generically for any workflow's Contract.
func Eligible(l project.Layout, wf contract.Workflow, suppliedSpec string) (Eligibility, error) {
	switch wf.SpecificationReference {
	case contract.SpecReferenceNone:
		return Eligibility{Eligible: true}, nil
	case contract.SpecReferenceRequired:
		if suppliedSpec == "" {
			return Eligibility{Eligible: false, Reason: "a Specification reference is required"}, nil
		}
	case contract.SpecReferenceOptional:
		if suppliedSpec == "" {
			return Eligibility{Eligible: true}, nil
		}
	}

	ok, _, err := specs.Exists(l.SpecificationsDir(), specs.Identity(suppliedSpec))
	if err != nil {
		return Eligibility{}, err
	}
	if !ok {
		return Eligibility{Eligible: false, Reason: fmt.Sprintf("%s does not exist", suppliedSpec)}, nil
	}

	if !wf.RequiresApprovedSpecification {
		return Eligibility{Eligible: true}, nil
	}

	state, err := Lifecycle(l, suppliedSpec)
	if err != nil {
		return Eligibility{}, err
	}
	if state != approval.Approved {
		return Eligibility{Eligible: false, Reason: fmt.Sprintf("%s is not Approved (currently %s)", suppliedSpec, state)}, nil
	}
	return Eligibility{Eligible: true}, nil
}
