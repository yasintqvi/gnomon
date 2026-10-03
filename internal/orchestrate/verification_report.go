package orchestrate

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gnomon/internal/present"
	"gnomon/internal/result"
)

// verificationIdentity is the workflow whose results Gnomon checks for evidence before reporting.
const verificationIdentity = "verification"

// noEvidenceNote replaces the evidence of a criterion reported as passing with nothing to show.
const noEvidenceNote = "No evidence was reported, so this criterion is unverified."

// missingTestNote prefixes the evidence of a criterion whose cited test cannot be found.
const missingTestNote = "The cited test %s was not found in the project, so this criterion is unverified."

// citedTestExists checks a cited test reference ("path", "path::name" or "path::Class::name")
// against the project: the file must exist under projectRoot, and the last name, if given, must
// appear in it. It proves the test is real, not that it asserts the criterion.
func citedTestExists(projectRoot, cited string) bool {
	parts := strings.Split(cited, "::")
	file := strings.TrimSpace(parts[0])
	if file == "" || filepath.IsAbs(file) || strings.HasPrefix(filepath.Clean(file), "..") {
		return false
	}
	data, err := os.ReadFile(filepath.Join(projectRoot, filepath.FromSlash(file)))
	if err != nil {
		return false
	}
	if len(parts) == 1 {
		return true
	}
	name := strings.TrimSpace(parts[len(parts)-1])
	if i := strings.IndexAny(name, "[( "); i >= 0 {
		name = name[:i] // parametrized ids, call parentheses
	}
	return name != "" && strings.Contains(string(data), name)
}

// evidenceDisclaimer accompanies every verification report.
const evidenceDisclaimer = "The Agent gathered and reported this evidence. It shows what was checked and how, not independent proof that the product works."

// conflictResolution is the only next step offered for a CONFLICT: the approved texts have to be
// reconciled before either can be met.
const conflictResolution = "knowledge-resolution"

// enforceEvidence makes a Verification result unable to claim more than its evidence shows: a
// criterion reported PASS without evidence, or resting on a cited test (test) that does not exist
// in the project, becomes UNVERIFIABLE; a criterion that names a
// conflicting approved statement (conflicts_with) is CONFLICT whatever the Agent concluded, and
// the overall result is lowered (never raised) to the worst criterion's — a CONFLICT counts as
// FAIL there, and PASS overall also requires at least one criterion. It edits outcome in place,
// so the classification, report, and finding extraction all see it.
func enforceEvidence(outcome *result.Outcome, projectRoot string) {
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
		if strings.TrimSpace(payloadString(item, "conflicts_with")) != "" {
			res = "CONFLICT"
			item["result"] = res
		}
		if res == "CONFLICT" {
			item["recommended_workflow"] = conflictResolution
		}
		if res == "PASS" {
			if cited := strings.TrimSpace(payloadString(item, "test")); cited != "" && !citedTestExists(projectRoot, cited) {
				res = "UNVERIFIABLE"
				item["result"] = res
				item["evidence"] = joinNonEmpty(fmt.Sprintf(missingTestNote, cited), payloadString(item, "evidence"))
			}
		}
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

// worseResult orders FAIL > UNVERIFIABLE > PASS for the overall result. A criterion's CONFLICT
// counts as FAIL: the overall result keeps the contract's three values.
func worseResult(a, b string) string {
	if b == "CONFLICT" {
		b = "FAIL"
	}
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
	case "CONFLICT":
		return "conflict"
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
		if test := strings.TrimSpace(payloadString(item, "test")); test != "" {
			lines = append(lines, fmt.Sprintf("   %-10s Test: %s", "", test))
		}
		lines = append(lines, fmt.Sprintf("   %-10s Evidence: %s", "", indentContinuation(evidence)))
		if word == "conflict" {
			with := strings.TrimSpace(payloadString(item, "conflicts_with"))
			if with == "" {
				with = "not named by the Agent"
			}
			lines = append(lines, fmt.Sprintf("   %-10s Conflicts with: %s", "", indentContinuation(with)))
		}
	}

	if len(lines) == 0 {
		rep.Summary = "Verification: nothing was verified"
		rep.AddSection("Criteria", "No criteria were reported, so nothing is verified.")
	} else {
		rep.Summary = fmt.Sprintf("Verification: %d passed, %d failed, %d unverified", counts["passed"], counts["failed"], counts["unverified"])
		if counts["conflict"] > 0 {
			rep.Summary = fmt.Sprintf("Verification: %d passed, %d failed, %d in conflict, %d unverified", counts["passed"], counts["failed"], counts["conflict"], counts["unverified"])
		}
		rep.AddSection("Criteria", strings.Join(lines, "\n"))
		if counts["conflict"] > 0 {
			rep.AddSection("Conflicts", "Approved Specifications contradict each other about the criteria marked conflict, so they cannot pass as written. "+
				"Reconcile the approved texts (Define on the affected Specifications, then approve), and verify again.")
		}
	}
	if summary, _ := payload["summary"].(map[string]interface{}); summary != nil {
		if scope := payloadString(summary, "obligations_evaluated"); scope != "" {
			rep.AddSection("Scope", scope)
		}
	}
	rep.AddSection("Note", evidenceDisclaimer)
}
