package orchestrate

import (
	"fmt"

	"gnomon/internal/facts"
	"gnomon/internal/present"
	"gnomon/internal/specs"
)

// DiscoveryCandidate is Specification Discovery's CANDIDATE_PROPOSED result, extracted from the
// raw payload before generic rendering — the shape the interactive accept/cancel flow presents to
// the Human. Identity is Discovery's own semantic-identity field (candidate_identity): a concise
// summary the Agent provides specifically for Gnomon's mechanical filename normalization,
// distinct from Title, which is only ever shown to the Human. Gnomon never derives one from the
// other.
type DiscoveryCandidate struct {
	Title             string
	Identity          string
	Rationale         string
	DerivedFrom       string
	NonBlockingCaveat string
}

// DiscoverCandidate runs Specification Discovery and returns a structured candidate when the
// result is CANDIDATE_PROPOSED, for the interactive accept/cancel flow to present — the Agent
// never creates anything itself; this function's caller decides, then AcceptDiscoveryCandidate
// performs the actual, deterministic creation. For every other outcome (NO_CANDIDATE_IDENTIFIED,
// BLOCKED, or a process/protocol failure), candidate is nil and the ordinary rendered Report is
// returned instead, via the same classifyAndRender path every other workflow-invoking command
// uses — never a duplicated rendering.
//
// A CANDIDATE_PROPOSED result missing a non-empty candidate_identity is refused as a protocol
// violation (present.Failed) rather than silently falling back to deriving one from the title:
// that would cross the Agent/CLI normalization boundary this field exists to enforce.
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

	outcome, detail, failureRep, err := obtainWorkflowOutcome(root, l, wf, workflowPath, targetNone, "", ad)
	if failureRep != nil {
		return nil, failureRep, err
	}

	if outcome.Terminal != "CANDIDATE_PROPOSED" {
		rep, renderErr := classifyAndRender(wf, "", detail, outcome)
		return nil, rep, renderErr
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
		return nil, rep, fmt.Errorf("discovery result missing candidate_title or candidate_identity")
	}
	rationale, _ := outcome.Payload["rationale"].(string)
	derivedFrom, _ := outcome.Payload["derived_from"].(string)
	caveat, _ := outcome.Payload["non_blocking_caveat"].(string)

	return &DiscoveryCandidate{
		Title:             title,
		Identity:          identity,
		Rationale:         rationale,
		DerivedFrom:       derivedFrom,
		NonBlockingCaveat: caveat,
	}, nil, nil
}

// AcceptDiscoveryCandidate performs deterministic Draft Creation for an accepted candidate — the
// Human's explicit authorization, never the Agent's own action. It reuses createDraftSpec, the
// exact same Core operation gnomon spec create uses, with the candidate's semantic identity
// (mechanically Slugified, never NLP-processed) determining the filename and the candidate's
// title determining only the displayed heading.
func AcceptDiscoveryCandidate(root string, candidate *DiscoveryCandidate) (*present.Report, error) {
	if candidate == nil || candidate.Title == "" || candidate.Identity == "" {
		return nil, fmt.Errorf("a candidate with both a title and identity is required")
	}
	return createDraftSpec(root, candidate.Title, specs.Slugify(candidate.Identity))
}
