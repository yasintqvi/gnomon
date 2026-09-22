package main

import (
	"strings"
	"testing"

	"gnomon/internal/orchestrate"
	"gnomon/internal/present"
)

func verificationFindings() []orchestrate.EvaluationFinding {
	return []orchestrate.EvaluationFinding{
		{ID: "F-001", Classification: "FAIL", Headline: "Refund cancels the charge", Detail: "charge.status remained pending", RecommendedWorkflow: orchestrate.ResolveWithImplementation},
		{ID: "F-002", Classification: "UNVERIFIABLE", Headline: "Idempotent retries", Detail: "no test harness available", RecommendedWorkflow: orchestrate.ResolveWithTesting},
	}
}

func reviewFindings() []orchestrate.EvaluationFinding {
	return []orchestrate.EvaluationFinding{
		{ID: "F-001", Classification: "DEFECT", Headline: "Missing rollback", Detail: "see test output",
			ResolutionOwner: "Implementation", RecommendedWorkflow: orchestrate.ResolveWithImplementation},
		{ID: "F-002", Classification: "RISK", Headline: "Retried without idempotency key",
			RecommendedWorkflow: orchestrate.ResolveWithKnowledgeResolution, DecisionRequired: true},
		{ID: "F-003", Classification: "KNOWLEDGE GAP", Headline: "No stated refund policy",
			ResolutionOwner: "Domain", RecommendedWorkflow: orchestrate.ResolveWithKnowledgeResolution, DecisionRequired: true},
	}
}

// --- parseFindingSelection ---

func TestParseFindingSelection_ValidIDs_CaseInsensitive(t *testing.T) {
	findings := reviewFindings()
	selected, invalid := parseFindingSelection("f-001 F-003", findings)
	if len(invalid) != 0 {
		t.Fatalf("expected no invalid IDs, got %v", invalid)
	}
	if len(selected) != 2 || selected[0].ID != "F-001" || selected[1].ID != "F-003" {
		t.Fatalf("unexpected selection: %+v", selected)
	}
}

func TestParseFindingSelection_UnknownID_ReportedInvalid(t *testing.T) {
	findings := reviewFindings()
	selected, invalid := parseFindingSelection("F-001 F-999", findings)
	if len(selected) != 1 || selected[0].ID != "F-001" {
		t.Fatalf("expected the one valid ID still selected, got %+v", selected)
	}
	if len(invalid) != 1 || invalid[0] != "F-999" {
		t.Fatalf("expected F-999 reported invalid, got %v", invalid)
	}
}

func TestParseFindingSelection_DuplicateTokens_Collapsed(t *testing.T) {
	findings := reviewFindings()
	selected, invalid := parseFindingSelection("F-001 F-001 F-001", findings)
	if len(invalid) != 0 {
		t.Fatalf("expected no invalid IDs, got %v", invalid)
	}
	if len(selected) != 1 {
		t.Fatalf("expected duplicates collapsed to one selection, got %+v", selected)
	}
}

func TestParseFindingSelection_EmptyAnswer_NoSelectionNoInvalid(t *testing.T) {
	selected, invalid := parseFindingSelection("   ", reviewFindings())
	if selected != nil || invalid != nil {
		t.Fatalf("expected no selection and no invalid IDs for blank input, got selected=%v invalid=%v", selected, invalid)
	}
}

// --- renderFindingList ---

func TestRenderFindingList_Verification_HeaderNamesObligations(t *testing.T) {
	out := renderFindingList("verification", verificationFindings(), present.NewPalette(false))
	if !strings.Contains(out, "Verification found 2 unresolved obligations") {
		t.Fatalf("expected the Verification-specific header, got: %s", out)
	}
	if !strings.Contains(out, "F-001") || !strings.Contains(out, "FAIL") {
		t.Fatalf("expected the first finding's ID and classification, got: %s", out)
	}
	if !strings.Contains(out, "Refund cancels the charge") {
		t.Fatalf("expected the obligation headline, got: %s", out)
	}
}

func TestRenderFindingList_Verification_Singular(t *testing.T) {
	out := renderFindingList("verification", verificationFindings()[:1], present.NewPalette(false))
	if !strings.Contains(out, "Verification found 1 unresolved obligation") {
		t.Fatalf("expected singular 'obligation', got: %s", out)
	}
}

func TestRenderFindingList_Review_HeaderNamesFindings(t *testing.T) {
	out := renderFindingList("review", reviewFindings(), present.NewPalette(false))
	if !strings.Contains(out, "Review found 3 findings") {
		t.Fatalf("expected the Review-specific header, got: %s", out)
	}
	for _, want := range []string{"F-001", "DEFECT", "F-002", "RISK", "F-003", "KNOWLEDGE GAP"} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected %q in rendered output, got: %s", want, out)
		}
	}
}

// TestRenderFindingList_DoesNotShowRecommendationOrEvidence proves the top-level list stays a
// scannable summary — full evidence and the recommendation live in renderFindingDetail, shown
// per finding once the Human is actually deciding on it, not in the overview.
func TestRenderFindingList_DoesNotShowRecommendationOrEvidence(t *testing.T) {
	out := renderFindingList("review", reviewFindings(), present.NewPalette(false))
	if strings.Contains(out, "Recommended resolution") {
		t.Fatalf("did not expect the recommendation in the overview list, got: %s", out)
	}
	if strings.Contains(out, "see test output") {
		t.Fatalf("did not expect full evidence in the overview list, got: %s", out)
	}
}

func TestRenderFindingList_NoColor_NoAnsi(t *testing.T) {
	out := renderFindingList("review", reviewFindings(), present.NewPalette(false))
	if strings.Contains(out, "\033[") {
		t.Fatalf("expected no ANSI codes with color disabled, got: %q", out)
	}
}

func TestRenderFindingList_Color_ClassificationStyled(t *testing.T) {
	p := present.NewPalette(true)
	out := renderFindingList("review", reviewFindings(), p)
	if !strings.Contains(out, p.Bad("DEFECT")) {
		t.Fatalf("expected DEFECT styled as a failure, got: %s", out)
	}
	if !strings.Contains(out, p.Warn("RISK")) || !strings.Contains(out, p.Warn("KNOWLEDGE GAP")) {
		t.Fatalf("expected RISK/KNOWLEDGE GAP styled as needing attention, got: %s", out)
	}
}

// --- renderFindingDetail ---

func TestRenderFindingDetail_ShowsRecommendationAsReference(t *testing.T) {
	out := renderFindingDetail(reviewFindings()[0], present.NewPalette(false))
	if !strings.Contains(out, "see test output") {
		t.Fatalf("expected full evidence shown, got: %s", out)
	}
	if !strings.Contains(out, "Recommended workflow:") || !strings.Contains(out, "Implementation") {
		t.Fatalf("expected the recommendation shown as reference, got: %s", out)
	}
}

// TestRenderFindingDetail_DecisionRequired_ShownWhenPresent proves "Decision required" is
// displayed as informational context when the Evaluation result carried it — independent of, and
// alongside, the recommended workflow — never used to alter which workflow is offered.
func TestRenderFindingDetail_DecisionRequired_ShownWhenPresent(t *testing.T) {
	out := renderFindingDetail(reviewFindings()[1], present.NewPalette(false))
	if !strings.Contains(out, "Recommended workflow:") || !strings.Contains(out, "Knowledge Resolution") {
		t.Fatalf("expected Knowledge Resolution shown as the recommended workflow, got: %s", out)
	}
	if !strings.Contains(out, "Decision required:") || !strings.Contains(out, "Yes") {
		t.Fatalf("expected Decision required shown, got: %s", out)
	}
}

// TestRenderFindingDetail_DecisionNotRequired_NoDecisionLine proves the line is omitted, not shown
// as "No", when the Evaluation result did not flag a decision as required.
func TestRenderFindingDetail_DecisionNotRequired_NoDecisionLine(t *testing.T) {
	out := renderFindingDetail(reviewFindings()[0], present.NewPalette(false))
	if strings.Contains(out, "Decision required") {
		t.Fatalf("did not expect a Decision required line when DecisionRequired is false, got: %s", out)
	}
}

func TestRenderFindingDetail_NoRecommendation_NoRecommendationLine(t *testing.T) {
	f := orchestrate.EvaluationFinding{ID: "F-001", Classification: "RISK", Headline: "x"}
	out := renderFindingDetail(f, present.NewPalette(false))
	if strings.Contains(out, "Recommended workflow") {
		t.Fatalf("did not expect a recommendation line when none was given, got: %s", out)
	}
}

// --- resolutionLabel ---

func TestResolutionLabel_KnownApproaches(t *testing.T) {
	cases := map[string]string{
		orchestrate.ResolveWithImplementation:      "Implementation",
		orchestrate.ResolveWithTesting:             "Testing",
		orchestrate.ResolveWithKnowledgeResolution: "Knowledge Resolution",
	}
	for approach, want := range cases {
		if got := resolutionLabel(approach); got != want {
			t.Errorf("resolutionLabel(%q) = %q, want %q", approach, got, want)
		}
	}
}

// --- findingMenuItems ---

// TestFindingMenuItems_DispatchableRecommendation_OffersContinueWith proves a real, validated
// recommendation is offered as the primary "Continue with <X>" action.
func TestFindingMenuItems_DispatchableRecommendation_OffersContinueWith(t *testing.T) {
	f := orchestrate.EvaluationFinding{ID: "F-001", Classification: "DEFECT", RecommendedWorkflow: orchestrate.ResolveWithImplementation}
	items := findingMenuItems(f)
	if len(items) != 3 {
		t.Fatalf("expected 3 items, got %d: %+v", len(items), items)
	}
	if items[0].value != "recommended" || items[0].label != "Continue with Implementation" {
		t.Fatalf("expected the first item to confirm the recommendation, got %+v", items[0])
	}
	if items[1].value != "choose" || items[2].value != "skip" {
		t.Fatalf("expected Choose another workflow then Skip, got %+v", items[1:])
	}
}

// TestFindingMenuItems_KnowledgeResolutionRecommendation_OffersContinueWith proves a RISK
// recommending knowledge-resolution — the case that used to route through the removed
// "human-decision"/"Make decision" special case — is offered exactly like any other dispatchable
// recommendation. There is no CLI-side semantic decision step; the launched Agent handles that.
func TestFindingMenuItems_KnowledgeResolutionRecommendation_OffersContinueWith(t *testing.T) {
	f := orchestrate.EvaluationFinding{ID: "F-002", Classification: "RISK", RecommendedWorkflow: orchestrate.ResolveWithKnowledgeResolution, DecisionRequired: true}
	items := findingMenuItems(f)
	if len(items) != 3 {
		t.Fatalf("expected 3 items, got %d: %+v", len(items), items)
	}
	if items[0].value != "recommended" || items[0].label != "Continue with Knowledge Resolution" {
		t.Fatalf("expected the first item to confirm the recommendation, got %+v", items[0])
	}
	for _, it := range items {
		if it.value == "make-decision" || strings.Contains(it.label, "Make decision") {
			t.Fatalf("did not expect a 'Make decision' item — that CLI-side flow was removed, got %+v", items)
		}
	}
}

// TestFindingMenuItems_NoRecommendation_OnlyChooseAndSkip proves a finding with no recommendation
// at all offers exactly the fallback path, no dead-first-item.
func TestFindingMenuItems_NoRecommendation_OnlyChooseAndSkip(t *testing.T) {
	f := orchestrate.EvaluationFinding{ID: "F-003", Classification: "RISK"}
	items := findingMenuItems(f)
	if len(items) != 2 || items[0].value != "choose" || items[1].value != "skip" {
		t.Fatalf("expected exactly [choose, skip], got %+v", items)
	}
}

// TestFindingMenuItems_HumanDecisionValue_NotOfferedAsDispatchable proves that even if a
// recommendation somehow still carries the removed "human-decision" string, it is never offered
// as something to dispatch (IsDispatchableResolution rejects it, matching the Contract's own
// closed enum, which no longer contains this value at all).
func TestFindingMenuItems_HumanDecisionValue_NotOfferedAsDispatchable(t *testing.T) {
	f := orchestrate.EvaluationFinding{ID: "F-004", Classification: "RISK", RecommendedWorkflow: "human-decision"}
	items := findingMenuItems(f)
	if len(items) != 2 || items[0].value != "choose" || items[1].value != "skip" {
		t.Fatalf("expected exactly [choose, skip] for an undispatchable recommendation, got %+v", items)
	}
}

// --- classificationStyle / findingListHeader / pluralize ---

func TestClassificationStyle_DefectAndFail_UseBad(t *testing.T) {
	p := present.NewPalette(true)
	if classificationStyle(p, "DEFECT") != p.Bad("DEFECT") {
		t.Fatalf("expected DEFECT to use Bad styling")
	}
	if classificationStyle(p, "FAIL") != p.Bad("FAIL") {
		t.Fatalf("expected FAIL to use Bad styling")
	}
}

func TestClassificationStyle_RiskUnverifiableKnowledgeGap_UseWarn(t *testing.T) {
	p := present.NewPalette(true)
	for _, c := range []string{"RISK", "UNVERIFIABLE", "KNOWLEDGE GAP"} {
		if classificationStyle(p, c) != p.Warn(c) {
			t.Fatalf("expected %s to use Warn styling", c)
		}
	}
}

func TestPluralize(t *testing.T) {
	if pluralize(1, "obligation", "obligations") != "obligation" {
		t.Fatalf("expected singular for 1")
	}
	if pluralize(0, "obligation", "obligations") != "obligations" {
		t.Fatalf("expected plural for 0")
	}
	if pluralize(2, "obligation", "obligations") != "obligations" {
		t.Fatalf("expected plural for 2")
	}
}
