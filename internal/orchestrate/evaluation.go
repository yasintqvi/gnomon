package orchestrate

import (
	"fmt"

	"gnomon/internal/adapter"
	"gnomon/internal/present"
)

// Workflows Gnomon's interactive finding-resolution loop can dispatch. This is the one closed
// vocabulary recommended_workflow is validated against, both by the Result Contract's own enum
// (schema level) and again here (code level) before anything is ever offered as an action —
// arbitrary Agent-authored text can never reach cmd/gnomon as something to execute. Every value
// here names a real, dispatchable workflow — there is no non-workflow placeholder value.
const (
	ResolveWithImplementation      = "implementation"
	ResolveWithTesting             = "testing"
	ResolveWithKnowledgeResolution = "knowledge-resolution"
)

// dispatchableResolutions are the recommended_workflow values that correspond to a workflow
// Gnomon can launch — currently all of them, since the vocabulary no longer contains a
// non-workflow value. Kept as its own map (rather than collapsing IsDispatchableResolution to a
// literal comparison) so dispatch safety never depends solely on the enum matching this set by
// construction.
var dispatchableResolutions = map[string]bool{
	ResolveWithImplementation:      true,
	ResolveWithTesting:             true,
	ResolveWithKnowledgeResolution: true,
}

// IsDispatchableResolution reports whether approach names a resolution workflow Gnomon actually
// knows how to launch — the one gate every resolution dispatch passes through, regardless of
// whether approach came from the Evaluation Agent's own recommendation or a Human's explicit
// choice from cmd/gnomon's fixed fallback menu. An Agent-authored value already passed the
// Result Contract's own enum at Result Protocol validation time, but this is checked again here,
// independently, so nothing about safe dispatch ever depends solely on schema validation having
// run correctly upstream.
func IsDispatchableResolution(approach string) bool {
	return dispatchableResolutions[approach]
}

// EvaluationFinding is one actionable item surfaced by a single Verification or Review run, for
// the interactive resolution loop only (cmd/gnomon). It is derived fresh from that run's own
// Result Protocol payload, exists only for the duration of the current CLI invocation, and is
// never written anywhere, never a durable identity, and never part of any Result Contract —
// Verification's and Review's schemas are completely unchanged by this type. Verification's own
// payload has no identity field of its own, so ID is assigned here, purely for this one
// interaction, in report-local order; Review's is its own already-stable-per-report finding_id
// (`F-001`, ...), carried through as-is.
type EvaluationFinding struct {
	ID             string
	Classification string // "FAIL" | "UNVERIFIABLE" (Verification), or "DEFECT" | "RISK" | "KNOWLEDGE GAP" (Review)
	Headline       string // Verification: the obligation text. Review: the finding summary.
	Detail         string // Supporting evidence — and, for Review, reasoning/impact — exactly as the Agent reported it, never reworded.

	// RecommendedWorkflow is the Evaluation Agent's own recommendation — one of the ResolveWith*
	// constants above, or "" when the Agent gave none. Read from the Result Contract's own
	// recommended_workflow field (already schema-validated against a closed enum) and re-checked
	// against IsDispatchableResolution before ever being offered as a default action; this field
	// is displayed and offered for confirmation, never executed automatically.
	RecommendedWorkflow string

	// ResolutionOwner is Review only; always "" for Verification, which has no equivalent field.
	// Free-form ("Domain", "Architecture", ...) and purely informational — distinct from
	// RecommendedWorkflow, which is Gnomon's own closed, dispatchable vocabulary.
	ResolutionOwner string

	// DecisionRequired signals that resolving this finding requires a material Human decision —
	// independent of RecommendedWorkflow: it says whether a Human decision is involved, not which
	// workflow handles it. Review only; always false for Verification.
	DecisionRequired bool
}

// ActionableFindings extracts EvaluationFinding from a Verification or Review run's own validated
// payload — the one place any Gnomon code reads "evidence" vs. "findings" or any other Result
// Contract payload field name for this purpose, so cmd/gnomon never needs to know that shape
// directly. Any other workflow identity returns nil: it has no equivalent concept, and that is
// never treated as an error — callers simply skip the interactive resolution offer.
func ActionableFindings(workflowIdentity string, payload map[string]interface{}) []EvaluationFinding {
	switch workflowIdentity {
	case "verification":
		return actionableVerificationFindings(payload)
	case "review":
		return actionableReviewFindings(payload)
	default:
		return nil
	}
}

// actionableVerificationFindings reads the "evidence" array and keeps only obligations whose
// result is not PASS — verification.md's own Outputs section defines exactly these three result
// values, and PASS is deliberately never offered as something to resolve. IDs are synthetic
// (F-001, F-002, ...), assigned in the order this run's own evidence array declares them; a
// different run over the same target may assign the same ID to a different obligation, which is
// exactly why these are never persisted or treated as a durable identity.
func actionableVerificationFindings(payload map[string]interface{}) []EvaluationFinding {
	items, _ := payload["evidence"].([]interface{})
	var findings []EvaluationFinding
	n := 0
	for _, raw := range items {
		item, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}
		result := payloadString(item, "result")
		if result == "" || result == "PASS" {
			continue
		}
		n++
		findings = append(findings, EvaluationFinding{
			ID:                  fmt.Sprintf("F-%03d", n),
			Classification:      result,
			Headline:            payloadString(item, "obligation"),
			Detail:              payloadString(item, "evidence"),
			RecommendedWorkflow: safeResolution(payloadString(item, "recommended_workflow")),
		})
	}
	return findings
}

// actionableReviewFindings reads the "findings" array as-is: review.md's own Step 5 ("Produce
// Actionable Findings") already guarantees only DEFECT, RISK, and KNOWLEDGE GAP ever appear
// there — PASS never creates a finding entry at all — so no further filtering happens here.
func actionableReviewFindings(payload map[string]interface{}) []EvaluationFinding {
	items, _ := payload["findings"].([]interface{})
	var findings []EvaluationFinding
	for _, raw := range items {
		item, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}
		id := payloadString(item, "finding_id")
		if id == "" {
			continue
		}
		detail := payloadString(item, "evidence")
		if reasoning := payloadString(item, "engineering_reasoning"); reasoning != "" {
			detail = joinNonEmpty(detail, reasoning)
		}
		if impact := payloadString(item, "impact"); impact != "" {
			detail = joinNonEmpty(detail, "Impact: "+impact)
		}
		decisionRequired, _ := item["decision_required"].(bool)
		findings = append(findings, EvaluationFinding{
			ID:                  id,
			Classification:      payloadString(item, "classification"),
			Headline:            payloadString(item, "summary"),
			Detail:              detail,
			ResolutionOwner:     payloadString(item, "resolution_owner"),
			RecommendedWorkflow: safeResolution(payloadString(item, "recommended_workflow")),
			DecisionRequired:    decisionRequired,
		})
	}
	return findings
}

// safeResolution passes through only a recognized, dispatchable recommendation and silently
// drops anything else to "". The Result Contract's own enum should already guarantee this, but
// extraction never trusts that alone.
func safeResolution(raw string) string {
	if dispatchableResolutions[raw] {
		return raw
	}
	return ""
}

func payloadString(m map[string]interface{}, key string) string {
	s, _ := m[key].(string)
	return s
}

func joinNonEmpty(a, b string) string {
	if a == "" {
		return b
	}
	return a + "\n\n" + b
}

// ResolutionHandoff is the transient context passed to a resolution Agent invocation launched
// from the interactive finding-resolution loop — never persisted, never part of any Result
// Contract, gone the moment that one invocation ends. It explains WHY the run started; it never
// replaces the resolution workflow's own normal target/eligibility requirements, which the
// caller supplies and validates completely independently (see RunResolutionWorkflow's specID).
type ResolutionHandoff struct {
	OriginWorkflow string // "verification" | "review"
	OriginTarget   string // the target the originating evaluation ran against
	FindingID      string // report-local; only meaningful for this one handoff
	Classification string
	Summary        string
	Evidence       string
}

// adapterHandoff converts to the adapter package's own wire type, so cmd/gnomon (this package's
// public API) never needs to import internal/adapter directly.
func (h ResolutionHandoff) adapterHandoff() *adapter.ResolutionHandoff {
	if h.FindingID == "" {
		return nil
	}
	return &adapter.ResolutionHandoff{
		OriginWorkflow: h.OriginWorkflow,
		OriginTarget:   h.OriginTarget,
		FindingID:      h.FindingID,
		Classification: h.Classification,
		Summary:        h.Summary,
		Evidence:       h.Evidence,
	}
}

// RunResolutionWorkflow launches one of the three dispatchable resolution workflows
// (implementation, testing, knowledge-resolution) carrying handoff as transient invocation
// context, on top of — never instead of — that workflow's completely normal target and
// eligibility handling: Implement's own Approved-Specification gate, in particular, is entirely
// unmodified and is what actually enforces that requirement here, not this function. specID is
// required for "implementation", optional for "testing", and unused for "knowledge-resolution".
// approach is re-validated with IsDispatchableResolution before anything runs; any unrecognized
// value is refused rather than silently doing nothing.
func RunResolutionWorkflow(root, approach, specID string, handoff ResolutionHandoff, agentOverride string, chooser AgentChooser) (*present.Report, error) {
	if !IsDispatchableResolution(approach) {
		return nil, fmt.Errorf("%q is not a resolution workflow Gnomon can dispatch", approach)
	}
	switch approach {
	case ResolveWithImplementation:
		return implementWithHandoff(root, specID, handoff, agentOverride, chooser)
	case ResolveWithTesting:
		return testWithHandoff(root, specID, handoff, agentOverride, chooser)
	case ResolveWithKnowledgeResolution:
		return knowledgeResolutionWithHandoff(root, handoff, agentOverride, chooser)
	default:
		// Unreachable given IsDispatchableResolution above; kept explicit rather than a silent
		// fallthrough.
		return nil, fmt.Errorf("%q is not a resolution workflow Gnomon can dispatch", approach)
	}
}

// knowledgeResolutionWithHandoff mirrors Run(root, "knowledge-resolution", "", ...) — the same
// eligibility (always none, per its own specification_reference: none) and execution path — but
// threading handoff through, and resolving the file directly since the identity is already known
// rather than scanning by identity the way the generic gnomon run does.
func knowledgeResolutionWithHandoff(root string, handoff ResolutionHandoff, agentOverride string, chooser AgentChooser) (*present.Report, error) {
	l, wf, workflowPath, err := prepareWorkflow(root, "knowledge-resolution.md")
	if err != nil {
		return nil, err
	}
	rep, _, err := runEligibleWorkflowFull(root, l, wf, workflowPath, targetKindFor(wf), "", handoff, agentOverride, chooser)
	return rep, err
}
