package orchestrate

import (
	"fmt"
	"strings"

	"gnomon/internal/adapter"
	"gnomon/internal/present"
)

// Workflows Gnomon's interactive finding-resolution loop can dispatch — the closed vocabulary
// recommended_workflow is validated against, both by the Result Contract's enum and again here, so
// arbitrary Agent-authored text can never reach cmd/gnomon as something to execute.
const (
	ResolveWithImplementation      = "implementation"
	ResolveWithTesting             = "testing"
	ResolveWithKnowledgeResolution = "knowledge-resolution"
)

// dispatchableResolutions is its own map, not a literal comparison, so dispatch safety never
// depends solely on the enum matching this set by construction.
var dispatchableResolutions = map[string]bool{
	ResolveWithImplementation:      true,
	ResolveWithTesting:             true,
	ResolveWithKnowledgeResolution: true,
}

// IsDispatchableResolution reports whether approach names a workflow Gnomon can launch — checked
// again here independently of the Result Contract's own enum, so dispatch safety never depends
// solely on schema validation upstream.
func IsDispatchableResolution(approach string) bool {
	return dispatchableResolutions[approach]
}

// EvaluationFinding is one actionable item surfaced by a Verification or Review run, for the
// interactive resolution loop only (cmd/gnomon) — never written anywhere, never part of any Result
// Contract. Verification's payload has no identity field, so ID is synthesized here (report-local
// order); Review's own stable per-report finding_id is carried through as-is.
type EvaluationFinding struct {
	ID             string
	Classification string // "FAIL" | "CONFLICT" | "UNVERIFIABLE" (Verification), or "DEFECT" | "RISK" | "KNOWLEDGE GAP" (Review)
	Headline       string // Verification: the obligation text. Review: the finding summary.
	Detail         string // Supporting evidence — and, for Review, reasoning/impact — exactly as the Agent reported it, never reworded.

	// RecommendedWorkflow is the Evaluation Agent's own recommendation, re-checked against
	// IsDispatchableResolution — displayed and offered for confirmation, never executed automatically.
	RecommendedWorkflow string

	// ResolutionOwner is Review only ("Domain", "Architecture", ...); always "" for Verification.
	ResolutionOwner string

	// DecisionRequired says whether resolving this finding needs a Human decision, independent of
	// which workflow handles it. Review only; always false for Verification.
	DecisionRequired bool
}

// ActionableFindings extracts EvaluationFinding from a Verification or Review run's validated
// payload, so cmd/gnomon never needs to know that payload shape directly. Any other workflow
// identity returns nil, never an error — callers simply skip the interactive resolution offer.
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

// actionableVerificationFindings keeps only non-PASS obligations. IDs are synthetic (F-001, ...),
// assigned in evidence-array order — never persisted or treated as a durable identity.
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
			Detail:              verificationFindingDetail(item),
			RecommendedWorkflow: safeResolution(payloadString(item, "recommended_workflow")),
		})
	}
	return findings
}

// verificationFindingDetail is a finding's evidence, plus the conflicting statement for a CONFLICT.
func verificationFindingDetail(item map[string]interface{}) string {
	detail := payloadString(item, "evidence")
	if with := strings.TrimSpace(payloadString(item, "conflicts_with")); with != "" {
		detail = joinNonEmpty(detail, "Conflicts with: "+with)
	}
	return detail
}

// actionableReviewFindings reads the "findings" array as-is: review.md's Step 5 ("Produce
// Actionable Findings") already guarantees only DEFECT/RISK/KNOWLEDGE GAP appear there.
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

// safeResolution passes through only a recognized, dispatchable recommendation, dropping anything
// else to "" — extraction never trusts the Result Contract's enum alone.
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

// ResolutionHandoff is the transient context passed to a resolution Agent invocation — never
// persisted, gone the moment the invocation ends. It explains why the run started; it never
// replaces the workflow's own normal target/eligibility requirements.
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

// RunResolutionWorkflow launches one of the three dispatchable resolution workflows carrying
// handoff, on top of — never instead of — that workflow's completely normal target/eligibility
// handling (Implement's Approved-Specification gate is unmodified). specID is required for
// "implementation", optional for "testing", unused for "knowledge-resolution".
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
		// Unreachable given IsDispatchableResolution above; kept explicit, not a silent fallthrough.
		return nil, fmt.Errorf("%q is not a resolution workflow Gnomon can dispatch", approach)
	}
}

// knowledgeResolutionWithHandoff mirrors Run(root, "knowledge-resolution", "", ...) but threads
// handoff through, and resolves the file directly since the identity is already known.
func knowledgeResolutionWithHandoff(root string, handoff ResolutionHandoff, agentOverride string, chooser AgentChooser) (*present.Report, error) {
	l, wf, workflowPath, err := prepareWorkflow(root, "knowledge-resolution.md")
	if err != nil {
		return nil, err
	}
	rep, _, err := runEligible(runRequest{root: root, l: l, wf: wf, workflowPath: workflowPath, kind: targetKindFor(wf), handoff: handoff, agentOverride: agentOverride, chooser: chooser})
	return rep, err
}
