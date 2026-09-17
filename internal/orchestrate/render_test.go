package orchestrate

import (
	"path/filepath"
	"strings"
	"testing"

	"gnomon/internal/contract"
	"gnomon/internal/present"
)

// TestRenderPayloadSections_GenericAcrossNestedNonImplementationWorkflow loads the real,
// migrated verification.md Contract — a payload shape (an array of per-obligation objects plus a
// nested "summary" object, terminal value at "summary.aggregate") structurally nothing like
// Implementation's flat fields — and proves the generic renderer handles it correctly with no Go
// code anywhere naming "obligation", "evidence", "summary", or "aggregate" specially.
func TestRenderPayloadSections_GenericAcrossNestedNonImplementationWorkflow(t *testing.T) {
	path := filepath.Join("..", "..", "workflows", "verification.md")
	wf, err := contract.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if wf.Result.TerminalPath != "summary.aggregate" {
		t.Fatalf("expected verification.md's real terminal_path to be summary.aggregate, got %q", wf.Result.TerminalPath)
	}

	payload := map[string]interface{}{
		"evidence": []interface{}{
			map[string]interface{}{
				"obligation": "All tests pass",
				"source":     "go test ./...",
				"result":     "FAIL",
				"evidence":   "TestFoo failed",
			},
		},
		"summary": map[string]interface{}{
			"obligations_evaluated": "3 of 3",
			"aggregate":             "FAIL",
		},
	}

	if err := wf.Result.ValidatePayload(payload); err != nil {
		t.Fatalf("expected this nested payload to satisfy verification.md's real schema: %v", err)
	}

	// A workflow's own negative conclusion (FAIL) must classify as "blocked", never as a process
	// failure — proving "Verification FAIL may still be a successfully completed workflow
	// execution" holds at the classification layer, not just in prose, and that classification
	// correctly resolves through a nested terminal value.
	terminal, err := wf.Result.TerminalValue(payload)
	if err != nil {
		t.Fatalf("expected the terminal value to extract through the nested path: %v", err)
	}
	class, ok := wf.Result.ClassificationFor(terminal)
	if !ok || class != "blocked" {
		t.Fatalf("expected FAIL to classify as blocked, got %q, ok=%v", class, ok)
	}

	sections := renderPayloadSections(wf.Result, payload)

	var foundObligationsEvaluated, foundEvidenceBlock bool
	for _, s := range sections {
		if s.Label == "Aggregate" || s.Label == "Summary" {
			t.Fatalf("the terminal leaf must never render as its own section, and its container ('summary') must never render as one opaque block — got label %q in: %+v", s.Label, sections)
		}
		if s.Label == "Obligations Evaluated" && s.Body == "3 of 3" {
			foundObligationsEvaluated = true
		}
		if s.Label == "Evidence" && strings.Contains(s.Body, "TestFoo failed") {
			foundEvidenceBlock = true
		}
	}
	if !foundObligationsEvaluated {
		t.Fatalf("expected the terminal value's sibling field (summary.obligations_evaluated) to still render, got: %+v", sections)
	}
	if !foundEvidenceBlock {
		t.Fatalf("expected the top-level array-of-objects 'evidence' field to render its nested content, got: %+v", sections)
	}

	rendered := present.Render(&present.Report{Outcome: present.Blocked, Summary: humanizeTerminalValue(terminal), Sections: sections}, false)
	if !strings.Contains(rendered, "! Fail") {
		t.Fatalf("expected the generic terminal-value humanizer to read naturally: %s", rendered)
	}
	if strings.Contains(rendered, "Aggregate") {
		t.Fatalf("expected no leaked 'Aggregate' section in final rendered output: %s", rendered)
	}
}

// TestRenderObjectSections_SuppressesOnlyTheTerminalLeaf is a synthetic, workflow-independent
// proof that path-aware suppression works generically at an arbitrary nesting depth and with
// arbitrary field names — not merely for "summary.aggregate" specifically. It builds its own
// throwaway ResultContract rather than loading any real workflow file.
func TestRenderObjectSections_SuppressesOnlyTheTerminalLeaf(t *testing.T) {
	rc := contract.ResultContract{
		TerminalPath: "box.status",
		PropertyOrder: map[string][]string{
			"":    {"label", "box"},
			"box": {"note", "status"},
		},
	}
	payload := map[string]interface{}{
		"label": "outer value",
		"box": map[string]interface{}{
			"note":   "inner sibling value",
			"status": "DONE",
		},
	}

	sections := renderObjectSections(rc, "", strings.Split(rc.TerminalPath, "."), payload)

	var labels []string
	for _, s := range sections {
		labels = append(labels, s.Label)
	}

	wantPresent := map[string]string{"Label": "outer value", "Note": "inner sibling value"}
	for label, body := range wantPresent {
		found := false
		for _, s := range sections {
			if s.Label == label && s.Body == body {
				found = true
			}
		}
		if !found {
			t.Fatalf("expected section %q=%q among %+v", label, body, sections)
		}
	}
	for _, s := range sections {
		if s.Label == "Status" || s.Label == "Box" {
			t.Fatalf("did not expect a section for the terminal leaf itself or its bare container, got: %+v", sections)
		}
	}
}

// TestRenderObjectSections_UnrelatedNestedObjectRendersNormally proves a nested object that is
// NOT on the terminal path's chain renders through the ordinary path (a single labeled block),
// unaffected by path-aware suppression — confirming that mechanism only ever engages for the
// object(s) actually on the terminal_path's own chain.
func TestRenderObjectSections_UnrelatedNestedObjectRendersNormally(t *testing.T) {
	rc := contract.ResultContract{
		TerminalPath:  "outcome",
		PropertyOrder: map[string][]string{"": {"outcome", "metadata"}},
	}
	payload := map[string]interface{}{
		"outcome":  "DONE",
		"metadata": map[string]interface{}{"author": "agent", "run": "1"},
	}
	sections := renderObjectSections(rc, "", strings.Split(rc.TerminalPath, "."), payload)
	if len(sections) != 1 || sections[0].Label != "Metadata" {
		t.Fatalf("expected exactly one 'Metadata' section for an object unrelated to the terminal path, got: %+v", sections)
	}
}

// TestRenderPayloadSections_RespectsDeclaredSchemaOrder proves section order follows the Result
// Contract's own declared property order, not alphabetical or Go map-iteration order — using the
// real implementation.md Contract, whose declared order is outcome, delivered,
// verification_evidence, remaining_unresolved (alphabetical would put remaining_unresolved before
// verification_evidence, which is not what this checks for).
func TestRenderPayloadSections_RespectsDeclaredSchemaOrder(t *testing.T) {
	path := filepath.Join("..", "..", "workflows", "implementation.md")
	wf, err := contract.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	payload := map[string]interface{}{
		"outcome":               "IMPLEMENTATION_COMPLETE",
		"delivered":             "a change",
		"verification_evidence": "go test passed",
		"remaining_unresolved":  "one open question",
	}
	sections := renderPayloadSections(wf.Result, payload)
	var labels []string
	for _, s := range sections {
		labels = append(labels, s.Label)
	}
	want := []string{"Delivered", "Verification Evidence", "Remaining Unresolved"}
	if len(labels) != len(want) {
		t.Fatalf("expected labels %v, got %v", want, labels)
	}
	for i, l := range want {
		if labels[i] != l {
			t.Fatalf("expected declared-order labels %v, got %v", want, labels)
		}
	}
}

// TestHumanizeTerminalValue_CoversRealVocabularyAcrossWorkflows proves the one generic
// SCREAMING_SNAKE_CASE -> sentence transform reads naturally for real terminal values drawn from
// several different workflows' own vocabularies, including the one with a literal space.
func TestHumanizeTerminalValue_CoversRealVocabularyAcrossWorkflows(t *testing.T) {
	cases := map[string]string{
		"IMPLEMENTATION_COMPLETE": "Implementation complete",
		"BLOCKED":                 "Blocked",
		"BOOTSTRAP_COMPLETE":      "Bootstrap complete",
		"READY_FOR_APPROVAL":      "Ready for approval",
		"COMMIT_PREPARED":         "Commit prepared",
		"PASS":                    "Pass",
		"KNOWLEDGE GAP":           "Knowledge gap",
	}
	for in, want := range cases {
		if got := humanizeTerminalValue(in); got != want {
			t.Errorf("humanizeTerminalValue(%q) = %q, want %q", in, got, want)
		}
	}
}
