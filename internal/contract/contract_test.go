package contract

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_RealMigratedImplementationWorkflow(t *testing.T) {
	// Ties directly into the actual Core migration this task performed — not a fixture.
	path := filepath.Join("..", "..", "workflows", "implementation.md")
	w, err := Load(path)
	if err != nil {
		t.Fatalf("expected the real workflows/implementation.md to load cleanly: %v", err)
	}
	if w.Identity != "implementation" {
		t.Fatalf("expected identity=implementation, got %q", w.Identity)
	}
	if w.SpecificationReference != SpecReferenceRequired {
		t.Fatalf("expected specification_reference=required, got %q", w.SpecificationReference)
	}
	if !w.RequiresApprovedSpecification {
		t.Fatalf("expected requires_approved_specification=true")
	}
	if w.Result.TerminalPath != "outcome" {
		t.Fatalf("expected result.terminal_path=outcome, got %q", w.Result.TerminalPath)
	}
	if err := w.Result.ValidatePayload(map[string]interface{}{"outcome": "IMPLEMENTATION_COMPLETE"}); err != nil {
		t.Fatalf("expected a valid enum value to pass schema validation: %v", err)
	}
	if err := w.Result.ValidatePayload(map[string]interface{}{"outcome": "BANANA"}); err == nil {
		t.Fatalf("expected an unknown Outcome value to fail schema validation")
	}
	if err := w.Result.ValidatePayload(map[string]interface{}{}); err == nil {
		t.Fatalf("expected a payload missing the required outcome field to fail schema validation")
	}
}

func writeFixture(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestValidate_MissingIdentity(t *testing.T) {
	dir := t.TempDir()
	path := writeFixture(t, dir, "wf.md", "---\nspecification_reference: none\nresult:\n  terminal_path: outcome\n  schema:\n    type: object\n---\n# Workflow\n")
	if _, err := Load(path); err == nil {
		t.Fatalf("expected an error for missing identity")
	}
}

func TestValidate_InvalidSpecificationReference(t *testing.T) {
	dir := t.TempDir()
	path := writeFixture(t, dir, "wf.md", "---\nidentity: x\nspecification_reference: sometimes\nresult:\n  terminal_path: outcome\n  schema:\n    type: object\n---\n# Workflow\n")
	if _, err := Load(path); err == nil {
		t.Fatalf("expected an error for an invalid specification_reference value")
	}
}

func TestValidate_ContradictoryApprovalRequirement(t *testing.T) {
	dir := t.TempDir()
	path := writeFixture(t, dir, "wf.md", "---\nidentity: x\nspecification_reference: none\nrequires_approved_specification: true\nresult:\n  terminal_path: outcome\n  schema:\n    type: object\n---\n# Workflow\n")
	if _, err := Load(path); err == nil {
		t.Fatalf("expected an error when requires_approved_specification is true while specification_reference is none")
	}
}

func TestValidate_MalformedFrontmatter(t *testing.T) {
	dir := t.TempDir()
	path := writeFixture(t, dir, "wf.md", "---\nidentity: [unclosed\n---\n# Workflow\n")
	if _, err := Load(path); err == nil {
		t.Fatalf("expected an error for malformed YAML")
	}
}

func TestValidate_MissingFrontmatterBlock(t *testing.T) {
	dir := t.TempDir()
	path := writeFixture(t, dir, "wf.md", "# Workflow\n\nNo frontmatter here.\n")
	if _, err := Load(path); err == nil {
		t.Fatalf("expected an error when no frontmatter block exists")
	}
}

// --- Terminal classification exhaustiveness (Step 12 Slice 1) ---

func TestValidate_MissingClassification(t *testing.T) {
	dir := t.TempDir()
	path := writeFixture(t, dir, "wf.md",
		"---\nidentity: x\nspecification_reference: none\nresult:\n  terminal_path: outcome\n"+
			"  schema:\n    type: object\n    required: [outcome]\n    properties:\n      outcome:\n"+
			"        type: string\n        enum: [A, B]\n---\n# Workflow\n")
	if _, err := Load(path); err == nil {
		t.Fatalf("expected an error when result.classification is entirely missing")
	}
}

func TestValidate_IncompleteClassification(t *testing.T) {
	dir := t.TempDir()
	path := writeFixture(t, dir, "wf.md",
		"---\nidentity: x\nspecification_reference: none\nresult:\n  terminal_path: outcome\n"+
			"  classification:\n    A: success\n"+
			"  schema:\n    type: object\n    required: [outcome]\n    properties:\n      outcome:\n"+
			"        type: string\n        enum: [A, B]\n---\n# Workflow\n")
	if _, err := Load(path); err == nil {
		t.Fatalf("expected an error when a schema-permitted terminal value (B) has no classification entry")
	}
}

func TestValidate_UnknownClassificationTarget(t *testing.T) {
	dir := t.TempDir()
	path := writeFixture(t, dir, "wf.md",
		"---\nidentity: x\nspecification_reference: none\nresult:\n  terminal_path: outcome\n"+
			"  classification:\n    A: success\n    B: blocked\n    C: blocked\n"+
			"  schema:\n    type: object\n    required: [outcome]\n    properties:\n      outcome:\n"+
			"        type: string\n        enum: [A, B]\n---\n# Workflow\n")
	if _, err := Load(path); err == nil {
		t.Fatalf("expected an error when classification references a value (C) the schema does not permit")
	}
}

func TestValidate_InconsistentClassificationValue(t *testing.T) {
	dir := t.TempDir()
	path := writeFixture(t, dir, "wf.md",
		"---\nidentity: x\nspecification_reference: none\nresult:\n  terminal_path: outcome\n"+
			"  classification:\n    A: success\n    B: maybe\n"+
			"  schema:\n    type: object\n    required: [outcome]\n    properties:\n      outcome:\n"+
			"        type: string\n        enum: [A, B]\n---\n# Workflow\n")
	if _, err := Load(path); err == nil {
		t.Fatalf("expected an error when a classification value is neither success nor blocked")
	}
}

func TestValidate_ExhaustiveConsistentClassification_Accepted(t *testing.T) {
	dir := t.TempDir()
	path := writeFixture(t, dir, "wf.md",
		"---\nidentity: x\nspecification_reference: none\nresult:\n  terminal_path: outcome\n"+
			"  classification:\n    A: success\n    B: blocked\n"+
			"  schema:\n    type: object\n    required: [outcome]\n    properties:\n      outcome:\n"+
			"        type: string\n        enum: [A, B]\n---\n# Workflow\n")
	w, err := Load(path)
	if err != nil {
		t.Fatalf("expected a complete, consistent classification to be accepted: %v", err)
	}
	if class, ok := w.Result.ClassificationFor("A"); !ok || class != "success" {
		t.Fatalf("expected ClassificationFor(A) = success, got %q, ok=%v", class, ok)
	}
	if class, ok := w.Result.ClassificationFor("B"); !ok || class != "blocked" {
		t.Fatalf("expected ClassificationFor(B) = blocked, got %q, ok=%v", class, ok)
	}
	if _, ok := w.Result.ClassificationFor("C"); ok {
		t.Fatalf("expected ClassificationFor to report unknown for a value never declared")
	}
}

// TestLoad_CapturesSchemaPropertyOrder proves PropertyOrder reflects declared document order —
// "zeta" before "alpha" here — never alphabetical or map-iteration order, since generic rendering
// depends on this to present a payload's fields the way its own author actually declared them.
func TestLoad_CapturesSchemaPropertyOrder(t *testing.T) {
	dir := t.TempDir()
	path := writeFixture(t, dir, "wf.md",
		"---\nidentity: x\nspecification_reference: none\nresult:\n  terminal_path: outcome\n"+
			"  classification:\n    A: success\n"+
			"  schema:\n    type: object\n    required: [outcome]\n    properties:\n"+
			"      outcome:\n        type: string\n        enum: [A]\n"+
			"      zeta:\n        type: [string, \"null\"]\n"+
			"      alpha:\n        type: [string, \"null\"]\n---\n# Workflow\n")
	w, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"outcome", "zeta", "alpha"}
	got := w.Result.PropertyOrder[""]
	if len(got) != len(want) {
		t.Fatalf("expected property order %v, got %v", want, got)
	}
	for i, k := range want {
		if got[i] != k {
			t.Fatalf("expected property order %v, got %v", want, got)
		}
	}
}

// TestLoad_CapturesNestedSchemaPropertyOrder proves PropertyOrder is captured at every object
// nesting level, keyed by dotted path — not only the schema's own top level — using a synthetic
// two-level object shape, independent of any real workflow's actual field names.
func TestLoad_CapturesNestedSchemaPropertyOrder(t *testing.T) {
	dir := t.TempDir()
	path := writeFixture(t, dir, "wf.md",
		"---\nidentity: x\nspecification_reference: none\nresult:\n  terminal_path: summary.status\n"+
			"  classification:\n    OK: success\n"+
			"  schema:\n    type: object\n    required: [summary]\n    properties:\n"+
			"      items:\n        type: [array, \"null\"]\n"+
			"      summary:\n        type: object\n        required: [status]\n        properties:\n"+
			"          note:\n            type: [string, \"null\"]\n"+
			"          status:\n            type: string\n            enum: [OK]\n---\n# Workflow\n")
	w, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := w.Result.PropertyOrder[""], []string{"items", "summary"}; !equalStrings(got, want) {
		t.Fatalf("expected top-level order %v, got %v", want, got)
	}
	if got, want := w.Result.PropertyOrder["summary"], []string{"note", "status"}; !equalStrings(got, want) {
		t.Fatalf("expected summary-level order %v, got %v", want, got)
	}
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// --- All 10 built-in workflow Contracts (Step 12 Slice 1) ---

// TestAllBuiltInWorkflows_LoadWithValidContracts loads every real, migrated workflow file and
// checks it against cli/WORKFLOW_CONTRACT.md's own finalized mapping table — proving all 10 load
// successfully, carry the exact declared Workflow Contract metadata, and have exhaustive terminal
// classification (implicitly, since Load would already have failed otherwise; explicitly, since
// this also confirms no extra/undeclared classification entries snuck in).
func TestAllBuiltInWorkflows_LoadWithValidContracts(t *testing.T) {
	cases := []struct {
		file             string
		identity         string
		specRef          SpecReference
		requiresApproved bool
		terminalPath     string
		terminalValues   []string
	}{
		{"implementation.md", "implementation", SpecReferenceRequired, true, "outcome",
			[]string{"IMPLEMENTATION_COMPLETE", "BLOCKED"}},
		{"bootstrap.md", "bootstrap", SpecReferenceNone, false, "outcome",
			[]string{"BOOTSTRAP_COMPLETE", "BLOCKED"}},
		{"git-finalization.md", "git-finalization", SpecReferenceNone, false, "outcome",
			[]string{"COMMIT_PREPARED", "PUBLISHED", "BLOCKED"}},
		{"initial-knowledge-establishment.md", "initial-knowledge-establishment", SpecReferenceNone, false, "outcome",
			[]string{"READY_FOR_BOOTSTRAP", "BLOCKED"}},
		{"knowledge-resolution.md", "knowledge-resolution", SpecReferenceNone, false, "outcome",
			[]string{"RESOLVED", "BLOCKED"}},
		{"specification-definition.md", "specification-definition", SpecReferenceRequired, false, "outcome",
			[]string{"READY_FOR_APPROVAL", "BLOCKED"}},
		{"specification-discovery.md", "specification-discovery", SpecReferenceNone, false, "outcome",
			[]string{"CANDIDATE_PROPOSED", "NO_CANDIDATE_IDENTIFIED", "BLOCKED"}},
		{"testing.md", "testing", SpecReferenceOptional, true, "outcome",
			[]string{"TESTING_COMPLETE", "BLOCKED"}},
		{"review.md", "review", SpecReferenceNone, false, "summary.aggregate",
			[]string{"DEFECT", "RISK", "KNOWLEDGE GAP", "PASS"}},
		{"verification.md", "verification", SpecReferenceNone, false, "summary.aggregate",
			[]string{"PASS", "FAIL", "UNVERIFIABLE"}},
	}

	if len(cases) != 10 {
		t.Fatalf("expected exactly 10 built-in workflows, got %d test cases", len(cases))
	}

	for _, c := range cases {
		t.Run(c.file, func(t *testing.T) {
			path := filepath.Join("..", "..", "workflows", c.file)
			w, err := Load(path)
			if err != nil {
				t.Fatalf("expected %s to load successfully with a valid Contract: %v", path, err)
			}
			if w.Identity != c.identity {
				t.Errorf("identity: got %q, want %q", w.Identity, c.identity)
			}
			if w.SpecificationReference != c.specRef {
				t.Errorf("specification_reference: got %q, want %q", w.SpecificationReference, c.specRef)
			}
			if w.RequiresApprovedSpecification != c.requiresApproved {
				t.Errorf("requires_approved_specification: got %v, want %v", w.RequiresApprovedSpecification, c.requiresApproved)
			}
			if w.Result.TerminalPath != c.terminalPath {
				t.Errorf("result.terminal_path: got %q, want %q", w.Result.TerminalPath, c.terminalPath)
			}
			for _, val := range c.terminalValues {
				class, ok := w.Result.ClassificationFor(val)
				if !ok {
					t.Errorf("expected an explicit classification for terminal value %q", val)
					continue
				}
				if class != "success" && class != "blocked" {
					t.Errorf("terminal value %q classifies as %q, want success or blocked", val, class)
				}
			}
			if len(w.Result.Classification) != len(c.terminalValues) {
				t.Errorf("expected exactly %d classified terminal values, got %d — classification must be exhaustive with no extra entries",
					len(c.terminalValues), len(w.Result.Classification))
			}
		})
	}
}

// TestClassification_ReviewKnowledgeGapWithSpace confirms the one genuinely unusual terminal
// value in Core's own vocabulary — Review's Aggregate literally reads "KNOWLEDGE GAP" with a
// space, not underscore-separated like every other workflow's tokens — round-trips correctly
// through schema validation and classification exactly as authored in Core's own prose, without
// this migration silently renaming it to "KNOWLEDGE_GAP" to fit a convention Core itself doesn't
// use here.
func TestClassification_ReviewKnowledgeGapWithSpace(t *testing.T) {
	path := filepath.Join("..", "..", "workflows", "review.md")
	wf, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	payload := map[string]interface{}{
		"findings": []interface{}{
			map[string]interface{}{
				"finding_id":     "F1",
				"classification": "KNOWLEDGE GAP",
				"summary":        "Unclear ownership of the retry policy.",
			},
		},
		"summary": map[string]interface{}{
			"aggregate": "KNOWLEDGE GAP",
		},
	}
	if err := wf.Result.ValidatePayload(payload); err != nil {
		t.Fatalf("expected the real review.md schema to accept the literal 'KNOWLEDGE GAP' value Core's own prose uses: %v", err)
	}
	class, ok := wf.Result.ClassificationFor("KNOWLEDGE GAP")
	if !ok || class != "blocked" {
		t.Fatalf("expected 'KNOWLEDGE GAP' to classify as blocked, got %q, ok=%v", class, ok)
	}
}

// TestClassification_ResolvesThroughRealNestedTerminalPath proves terminalEnum/ClassificationFor
// correctly walk a genuinely nested terminal_path (review.md's real "summary.aggregate") end to
// end: extraction, schema validation, and classification all agree on the same nested value.
func TestClassification_ResolvesThroughRealNestedTerminalPath(t *testing.T) {
	path := filepath.Join("..", "..", "workflows", "verification.md")
	wf, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if wf.Result.TerminalPath != "summary.aggregate" {
		t.Fatalf("expected verification.md's terminal_path to be summary.aggregate, got %q", wf.Result.TerminalPath)
	}
	payload := map[string]interface{}{
		"summary": map[string]interface{}{
			"aggregate":             "FAIL",
			"obligations_evaluated": "3 of 3",
		},
	}
	if err := wf.Result.ValidatePayload(payload); err != nil {
		t.Fatalf("expected this nested payload to satisfy the real schema: %v", err)
	}
	terminal, err := wf.Result.TerminalValue(payload)
	if err != nil {
		t.Fatalf("expected the terminal value to be extractable through the nested path: %v", err)
	}
	if terminal != "FAIL" {
		t.Fatalf("expected extracted terminal value %q, got %q", "FAIL", terminal)
	}
	class, ok := wf.Result.ClassificationFor(terminal)
	if !ok || class != "blocked" {
		t.Fatalf("expected the extracted nested terminal value to classify as blocked, got %q, ok=%v", class, ok)
	}
}
