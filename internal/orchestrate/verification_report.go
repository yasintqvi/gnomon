package orchestrate

import (
	"fmt"
	"strings"

	"gnomon/internal/present"
	"gnomon/internal/result"
)

// verificationIdentity is the workflow whose results Gnomon checks for evidence before reporting.
const verificationIdentity = "verification"

// noEvidenceNote replaces the evidence of a criterion reported as passing with nothing to show.
const noEvidenceNote = "No evidence was reported, so this criterion is unverified."

// evidenceDisclaimer accompanies every verification report.
const evidenceDisclaimer = "The Agent gathered and reported this evidence. It shows what was checked and how, not independent proof that the product works."

// enforceEvidence makes a Verification result unable to claim more than its evidence shows: a
// criterion reported PASS without evidence becomes UNVERIFIABLE, and the overall result is lowered
// (never raised) to the worst criterion's — PASS overall also requires at least one criterion.
// It edits outcome in place, so the classification, report, and finding extraction all see it.
func enforceEvidence(outcome *result.Outcome) {
	items, _ := outcome.Payload["evidence"].([]interface{})
	worst := "PASS"
	if len(items) == 0 {
		worst = "UNVERIFIABLE"
	}
	for _, raw := range items {
		item, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}
		res := payloadString(item, "result")
		if res == "PASS" && strings.TrimSpace(payloadString(item, "evidence")) == "" {
			res = "UNVERIFIABLE"
			item["result"] = res
			item["evidence"] = noEvidenceNote
		}
		worst = worseResult(worst, res)
	}

	summary, _ := outcome.Payload["summary"].(map[string]interface{})
	if summary == nil {
		return
	}
	aggregate := worseResult(payloadString(summary, "aggregate"), worst)
	summary["aggregate"] = aggregate
	outcome.Terminal = aggregate
}

// worseResult orders FAIL > UNVERIFIABLE > PASS.
func worseResult(a, b string) string {
	rank := map[string]int{"PASS": 0, "UNVERIFIABLE": 1, "FAIL": 2}
	if rank[b] > rank[a] {
		return b
	}
	return a
}

// resultWord is the plain word a report uses for a verification result.
func resultWord(res string) string {
	switch res {
	case "PASS":
		return "passed"
	case "FAIL":
		return "failed"
	default:
		return "unverified"
	}
}

// renderVerification builds a Verification report: one entry per criterion with its result and
// evidence, a one-line tally, and the evidence disclaimer.
func renderVerification(rep *present.Report, payload map[string]interface{}) {
	items, _ := payload["evidence"].([]interface{})
	counts := map[string]int{}
	var lines []string
	for i, raw := range items {
		item, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}
		word := resultWord(payloadString(item, "result"))
		counts[word]++
		lines = append(lines, fmt.Sprintf("%d. %-10s %s", i+1, word, payloadString(item, "obligation")))
		if src := payloadString(item, "source"); src != "" {
			lines = append(lines, fmt.Sprintf("   %-10s Source: %s", "", src))
		}
		evidence := strings.TrimSpace(payloadString(item, "evidence"))
		if evidence == "" {
			evidence = "none reported"
		}
		lines = append(lines, fmt.Sprintf("   %-10s Evidence: %s", "", indentContinuation(evidence)))
	}

	if len(lines) == 0 {
		rep.Summary = "Verification: nothing was verified"
		rep.AddSection("Criteria", "No criteria were reported, so nothing is verified.")
	} else {
		rep.Summary = fmt.Sprintf("Verification: %d passed, %d failed, %d unverified", counts["passed"], counts["failed"], counts["unverified"])
		rep.AddSection("Criteria", strings.Join(lines, "\n"))
	}
	if summary, _ := payload["summary"].(map[string]interface{}); summary != nil {
		if scope := payloadString(summary, "obligations_evaluated"); scope != "" {
			rep.AddSection("Scope", scope)
		}
	}
	rep.AddSection("Note", evidenceDisclaimer)
}
