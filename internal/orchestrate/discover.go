package orchestrate

import (
	"fmt"

	"gnomon/internal/facts"
	"gnomon/internal/present"
	"gnomon/internal/project"
	"gnomon/internal/specs"
)

// DiscoveryCandidate is Specification Discovery's CANDIDATE_PROPOSED result, presented to the
// Human for accept/cancel. Identity (candidate_identity) drives filename normalization; Title is
// shown to the Human only. Gnomon never derives one from the other.
type DiscoveryCandidate struct {
	Title             string
	Identity          string
	Rationale         string
	DerivedFrom       string
	NonBlockingCaveat string
}

// DiscoverCandidate runs Specification Discovery and returns a structured candidate when the
// result is CANDIDATE_PROPOSED — the Agent never creates anything itself; the caller decides, then
// AcceptDiscoveryCandidate performs the actual creation. For every other outcome, candidate is nil
// and the ordinary rendered Report is returned via classifyAndRender instead.
//
// A CANDIDATE_PROPOSED result missing candidate_identity is refused as a protocol violation rather
// than falling back to deriving one from the title — that would cross the Agent/CLI normalization
// boundary this field exists to enforce.
func DiscoverCandidate(root, agentOverride string, chooser AgentChooser) (*DiscoveryCandidate, *present.Report, error) {
	l, wf, workflowPath, err := prepareWorkflow(root, "specification-discovery.md")
	if err != nil {
		return nil, nil, err
	}
	if elig, err := facts.Eligible(l, wf, ""); err != nil {
		return nil, nil, err
	} else if !elig.Eligible {
		return nil, &present.Report{
			Outcome:  present.Blocked,
			Summary:  "Specification Discovery blocked",
			Sections: []present.Section{{Label: "Unresolved", Body: elig.Reason}},
		}, fmt.Errorf("ineligible: %s", elig.Reason)
	}
	ad, err := ResolveAgent(agentOverride, chooser)
	if err != nil {
		return nil, &present.Report{
			Outcome: present.Failed,
			Summary: "Could not determine which Agent to use",
			Sections: []present.Section{
				{Label: "Reason", Body: err.Error()},
			},
			Next: "Pass --agent claude|codex, or run `gnomon agent set-default <claude|codex>`.",
		}, err
	}

	// Rule B, inlined: Discovery has no governing Specification of its own (kind is always
	// targetNone), so only the "some OTHER Specification lost approval" check ever applies here —
	// same mechanism execute uses on every other run, via the same shared helpers.
	approvedBefore, snapErr := approvedIdentities(l)
	if snapErr != nil {
		return nil, nil, snapErr
	}

	outcome, detail, failureRep, err := obtainOutcome(runRequest{root: root, l: l, wf: wf, workflowPath: workflowPath, kind: targetNone, adapter: ad})
	if failureRep != nil {
		return nil, discoveryLostApprovalsReport(l, approvedBefore, failureRep), err
	}

	if outcome.Terminal != "CANDIDATE_PROPOSED" {
		rep, renderErr := classifyAndRender(wf, "", detail, outcome)
		return nil, discoveryLostApprovalsReport(l, approvedBefore, rep), renderErr
	}

	title, _ := outcome.Payload["candidate_title"].(string)
	identity, _ := outcome.Payload["candidate_identity"].(string)
	if title == "" || identity == "" {
		rep := &present.Report{
			Outcome: present.Failed,
			Summary: "Discovery result is missing required fields",
			Sections: []present.Section{
				{Label: "Reason", Body: "CANDIDATE_PROPOSED requires both candidate_title and a non-empty candidate_identity; the Agent never determines a filename directly."},
			},
			Detail: detail,
		}
		return nil, discoveryLostApprovalsReport(l, approvedBefore, rep), fmt.Errorf("discovery result missing candidate_title or candidate_identity")
	}
	rationale, _ := outcome.Payload["rationale"].(string)
	derivedFrom, _ := outcome.Payload["derived_from"].(string)
	caveat, _ := outcome.Payload["non_blocking_caveat"].(string)

	// CANDIDATE_PROPOSED returns a DiscoveryCandidate here, not a Report, so there is nothing to
	// attach a lost-approval warning to at this point — the caller renders its own presentation
	// from the candidate fields. Closing this one gap would mean changing this return shape.
	return &DiscoveryCandidate{
		Title:             title,
		Identity:          identity,
		Rationale:         rationale,
		DerivedFrom:       derivedFrom,
		NonBlockingCaveat: caveat,
	}, nil, nil
}

// discoveryLostApprovalsReport applies Rule B's "other Specification lost approval" warning to
// rep, the same way execute's own deferred check does — Discovery never has a governing
// Specification of its own, so only that side of the check ever applies here. A failure computing
// it is silently ignored, exactly as in execute: this is a best-effort warning annotation, never a
// reason to change the run's own outcome.
func discoveryLostApprovalsReport(l project.Layout, approvedBefore map[string]bool, rep *present.Report) *present.Report {
	if rep == nil {
		return rep
	}
	lost, err := lostApprovals(l, approvedBefore, "")
	if err != nil || len(lost) == 0 {
		return rep
	}
	rep.Sections = append(rep.Sections, lostApprovalSection(lost))
	return rep
}

// AcceptDiscoveryCandidate performs deterministic Draft Creation for an accepted candidate — the
// Human's explicit authorization, never the Agent's own action. Reuses createDraftSpec (the same
// operation gnomon spec create uses); the candidate's identity (Slugified) drives the filename.
func AcceptDiscoveryCandidate(root string, candidate *DiscoveryCandidate) (*present.Report, error) {
	if candidate == nil || candidate.Title == "" || candidate.Identity == "" {
		return nil, fmt.Errorf("a candidate with both a title and identity is required")
	}
	return createDraftSpec(root, candidate.Title, specs.Slugify(candidate.Identity), false)
}
