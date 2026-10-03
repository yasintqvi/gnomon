package orchestrate

import (
	"strings"
	"testing"

	"gnomon/internal/present"
)

// Criteria shaped like study 2's final arm-C project: three approved Specifications state the
// loan's field set differently (exactly seven, then eight, then nine), and the later ones carry
// "compatibility notes" that extend the earlier criteria without changing their text.
var (
	sevenFields = map[string]interface{}{
		"obligation": "When staff POST /loans, the response is 201 with exactly the fields id, unit_id, member_id, loaned_at, due_at, status and returned_at.",
		"source":     "SPEC-003, criterion 1",
	}
	eightFields = map[string]interface{}{
		"obligation": "A new checkout returns the seven existing loan fields plus renewals: 0; listings expose those same eight fields.",
		"source":     "SPEC-005, criterion 1",
	}
	nineFields = map[string]interface{}{
		"obligation": "A loan response has the eight existing loan fields plus fee: 0; listings expose those same nine fields.",
		"source":     "SPEC-007, criterion 1",
	}
	unrelatedCriterion = map[string]interface{}{
		"obligation": "GET /version returns 200 with {\"version\": \"0.1.0\"}.",
		"source":     "SPEC-001, criterion 1",
		"result":     "PASS",
		"evidence":   "probe: 200 {\"version\": \"0.1.0\"}; test_version passes",
	}
)

func criterion(base map[string]interface{}, fields map[string]interface{}) map[string]interface{} {
	out := map[string]interface{}{}
	for k, v := range base {
		out[k] = v
	}
	for k, v := range fields {
		out[k] = v
	}
	return out
}

func verificationPayload(aggregate string, items ...map[string]interface{}) map[string]interface{} {
	evidence := make([]interface{}, len(items))
	for i, it := range items {
		evidence[i] = it
	}
	return map[string]interface{}{"evidence": evidence, "summary": map[string]interface{}{"aggregate": aggregate}}
}

// Regression (study 2): the field-count criteria contradict each other. Reported as CONFLICT, they
// are not passed, the overall result is a failure, and each is offered for knowledge resolution.
func TestVerification_ConflictingApprovedCriteria_NotPassed(t *testing.T) {
	probe := "probe: 201; field set is id, unit_id, member_id, loaned_at, due_at, status, returned_at, renewals, fee"
	rep, findings, err := runVerification(t, verificationPayload("FAIL",
		criterion(sevenFields, map[string]interface{}{"result": "CONFLICT", "evidence": probe,
			"conflicts_with": `SPEC-007, criterion 1: "those same nine fields"`, "recommended_workflow": "knowledge-resolution"}),
		criterion(eightFields, map[string]interface{}{"result": "CONFLICT", "evidence": probe,
			"conflicts_with": `SPEC-007, criterion 1: "nine fields"`, "recommended_workflow": "knowledge-resolution"}),
		criterion(nineFields, map[string]interface{}{"result": "CONFLICT", "evidence": probe,
			"conflicts_with": `SPEC-003, criterion 1: "exactly the fields" (seven)`, "recommended_workflow": "knowledge-resolution"}),
		unrelatedCriterion,
	))
	if err == nil || rep.Outcome != present.Blocked {
		t.Fatalf("conflicting approved criteria must not succeed, got %v (err=%v)", rep.Outcome, err)
	}
	if rep.Summary != "Verification: 1 passed, 0 failed, 3 in conflict, 0 unverified" {
		t.Fatalf("unexpected summary %q", rep.Summary)
	}
	criteria := sectionBody(rep, "Criteria")
	for _, want := range []string{
		"1. conflict   When staff POST /loans",
		`Conflicts with: SPEC-007, criterion 1: "those same nine fields"`,
		"4. passed     GET /version",
	} {
		if !strings.Contains(criteria, want) {
			t.Fatalf("expected %q in:\n%s", want, criteria)
		}
	}
	if sectionBody(rep, "Conflicts") == "" {
		t.Fatal("expected a Conflicts section explaining what to do")
	}
	if len(findings) != 3 {
		t.Fatalf("expected the three conflicting criteria as findings, got %+v", findings)
	}
	for _, f := range findings {
		if f.Classification != "CONFLICT" || f.RecommendedWorkflow != ResolveWithKnowledgeResolution {
			t.Fatalf("expected CONFLICT findings routed to knowledge resolution, got %+v", f)
		}
		if !strings.Contains(f.Detail, "Conflicts with: SPEC-") {
			t.Fatalf("expected the finding to name the conflicting statement, got %q", f.Detail)
		}
	}
}

// What the study-2 run actually did: it saw the contradiction, relied on a compatibility note,
// and reported PASS overall. A criterion that names a conflicting approved statement can never
// pass, whatever the Agent concluded — and Gnomon lowers the overall result.
func TestVerification_PassDespiteNamedConflict_BecomesConflict(t *testing.T) {
	rep, findings, err := runVerification(t, verificationPayload("PASS",
		criterion(sevenFields, map[string]interface{}{"result": "PASS",
			"evidence":       "probe: field set is the seven plus renewals (SPEC-005) and fee (SPEC-007)",
			"conflicts_with": "SPEC-007, criterion 1 (nine fields); SPEC-007's compatibility note extends this criterion"}),
		unrelatedCriterion,
	))
	if err == nil || rep.Outcome != present.Blocked {
		t.Fatalf("a compatibility note must not turn contradictory approved criteria into a pass, got %v (err=%v)", rep.Outcome, err)
	}
	if rep.Summary != "Verification: 1 passed, 0 failed, 1 in conflict, 0 unverified" {
		t.Fatalf("unexpected summary %q", rep.Summary)
	}
	if len(findings) != 1 || findings[0].Classification != "CONFLICT" || findings[0].RecommendedWorkflow != ResolveWithKnowledgeResolution {
		t.Fatalf("expected one CONFLICT finding for knowledge resolution, got %+v", findings)
	}
}

// An Agent that reports CONFLICT but an overall PASS, or recommends implementation for it, is
// corrected: the overall result is a failure and the only next step is knowledge resolution.
func TestVerification_ConflictLowersAggregateAndForcesKnowledgeResolution(t *testing.T) {
	rep, findings, err := runVerification(t, verificationPayload("PASS",
		criterion(eightFields, map[string]interface{}{"result": "CONFLICT", "evidence": "probe: nine fields",
			"conflicts_with": "SPEC-007, criterion 1", "recommended_workflow": "implementation"}),
	))
	if err == nil || rep.Outcome != present.Blocked {
		t.Fatalf("expected the overall result lowered to a failure, got %v (err=%v)", rep.Outcome, err)
	}
	if len(findings) != 1 || findings[0].RecommendedWorkflow != ResolveWithKnowledgeResolution {
		t.Fatalf("a conflict is resolved by reconciling the texts, not by implementation; got %+v", findings)
	}
}

// A CONFLICT whose conflicting statement was not named is still not passed, and says so.
func TestVerification_ConflictWithoutNamedStatement_StillNotPassed(t *testing.T) {
	rep, _, err := runVerification(t, verificationPayload("FAIL",
		criterion(eightFields, map[string]interface{}{"result": "CONFLICT", "evidence": "probe: nine fields"}),
	))
	if err == nil || rep.Outcome != present.Blocked {
		t.Fatalf("expected a failure, got %v (err=%v)", rep.Outcome, err)
	}
	if !strings.Contains(sectionBody(rep, "Criteria"), "Conflicts with: not named by the Agent") {
		t.Fatalf("expected the report to say the conflicting statement was not named:\n%s", sectionBody(rep, "Criteria"))
	}
}

// Ordinary verification is not rejected: a later Specification that adds behavior an earlier
// criterion does not address (no "exactly"), with conflicts_with null or empty, still passes,
// even when the evidence mentions the later Specification. The report format is unchanged.
func TestVerification_NonConflictingChange_StillPasses(t *testing.T) {
	includesFields := map[string]interface{}{
		"obligation": "When staff POST /loans, the response is 201 and includes id, unit_id, member_id, loaned_at, due_at, status and returned_at.",
		"source":     "SPEC-003, criterion 1",
		"result":     "PASS",
		"evidence":   "probe: all seven present; the response also has renewals, added by SPEC-005",
	}
	renewals := criterion(eightFields, map[string]interface{}{"result": "PASS", "evidence": "test_renewals passes", "conflicts_with": nil})
	emptyConflict := criterion(unrelatedCriterion, map[string]interface{}{"conflicts_with": "  "})

	rep, findings, err := runVerification(t, verificationPayload("PASS", includesFields, renewals, emptyConflict))
	if err != nil || rep.Outcome != present.Success {
		t.Fatalf("non-conflicting criteria must pass, got %v (err=%v)", rep.Outcome, err)
	}
	if rep.Summary != "Verification: 3 passed, 0 failed, 0 unverified" {
		t.Fatalf("unexpected summary %q", rep.Summary)
	}
	if sectionBody(rep, "Conflicts") != "" || strings.Contains(sectionBody(rep, "Criteria"), "Conflicts with") {
		t.Fatal("a report without conflicts must not mention them")
	}
	if findings != nil {
		t.Fatalf("expected no findings, got %+v", findings)
	}
}

// A plain FAIL elsewhere still reads as before when there are no conflicts.
func TestVerification_FailWithoutConflict_Unchanged(t *testing.T) {
	rep, findings, err := runVerification(t, verificationPayload("FAIL",
		criterion(unrelatedCriterion, map[string]interface{}{"result": "FAIL", "evidence": "probe: 500", "recommended_workflow": "implementation"}),
	))
	if err == nil || rep.Summary != "Verification: 0 passed, 1 failed, 0 unverified" {
		t.Fatalf("unexpected %q (err=%v)", rep.Summary, err)
	}
	if len(findings) != 1 || findings[0].Classification != "FAIL" || findings[0].RecommendedWorkflow != ResolveWithImplementation {
		t.Fatalf("expected an unchanged FAIL finding, got %+v", findings)
	}
}
