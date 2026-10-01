package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestE2E_NormalPath_MinimalInitShortSpecApproveEditDraft drives the real binary along the normal
// path: a minimal init, a short Specification, refusal to approve it untouched, approval once it
// has content, and the return to Draft (with the reason shown) after an edit.
func TestE2E_NormalPath_MinimalInitShortSpecApproveEditDraft(t *testing.T) {
	dir, env := gitInitializedRepo(t)
	code, out := runGnomonOutput(t, dir, env, "init")
	if code != 0 || !strings.Contains(out, "gnomon spec") {
		t.Fatalf("init exit %d, expected the next step to be defining a change:\n%s", code, out)
	}
	entries, _ := os.ReadDir(filepath.Join(dir, ".gnomon"))
	var names []string
	for _, e := range entries {
		names = append(names, e.Name())
	}
	if strings.Join(names, ",") != "CONTRACT_VERSION,approvals,specifications" {
		t.Fatalf("expected a minimal .gnomon/, got %v", names)
	}

	if code, out := runGnomonOutput(t, dir, env, "spec", "create", "Mark a task complete"); code != 0 {
		t.Fatalf("spec create exit %d: %s", code, out)
	}
	specPath := specPathFor(t, dir)
	created, _ := os.ReadFile(specPath)
	for _, section := range []string{"## Goal", "## Decisions", "## Acceptance Criteria", "## Out of Scope"} {
		if !strings.Contains(string(created), section) {
			t.Fatalf("expected the short template's %q section, got:\n%s", section, created)
		}
	}

	if code, out := runGnomonOutput(t, dir, env, "approve", "SPEC-001"); code == 0 || !strings.Contains(out, "nothing to approve yet") {
		t.Fatalf("expected an untouched template to be refused, got exit %d:\n%s", code, out)
	}

	defined := strings.NewReplacer(
		"[Who needs what, and why — one or two sentences.]", "Users mark finished tasks complete.",
		"- [Question] — [Answer]", "- Can a completed task be reopened? — Yes.",
		"1. [Given a situation, when something happens, then this is the result.]", "1. An open task can be marked complete.\n2. A completed task can be reopened.",
		"- [Something this change deliberately does not do]", "- Bulk completion.",
	).Replace(string(created))
	if err := os.WriteFile(specPath, []byte(defined), 0o644); err != nil {
		t.Fatal(err)
	}
	if code, out := runGnomonOutput(t, dir, env, "approve", "SPEC-001"); code != 0 || !strings.Contains(out, "SPEC-001 is now Approved") {
		t.Fatalf("expected approval, got exit %d:\n%s", code, out)
	}

	if err := os.WriteFile(specPath, []byte(strings.Replace(defined, "— Yes.", "— No, completion is final.", 1)), 0o644); err != nil {
		t.Fatal(err)
	}
	code, out = runGnomonOutput(t, dir, env, "status")
	if code != 0 || !strings.Contains(out, "SPEC-001: Draft — changed since approval on") {
		t.Fatalf("expected status to show the change since approval, got exit %d:\n%s", code, out)
	}

	code, out = runGnomonOutput(t, dir, env, "validate")
	if code != 0 {
		t.Fatalf("expected a minimal project to validate, got exit %d:\n%s", code, out)
	}
}

func TestE2E_DetailedTemplateAndFullInitRemainAvailable(t *testing.T) {
	dir, env := gitInitializedRepo(t)
	if code, out := runGnomonOutput(t, dir, env, "init", "--full"); code != 0 {
		t.Fatalf("init --full exit %d: %s", code, out)
	}
	if _, err := os.Stat(filepath.Join(dir, ".gnomon", "context", "UI_FOUNDATION.md")); err != nil {
		t.Fatalf("expected --full to create the optional templates: %v", err)
	}
	if code, out := runGnomonOutput(t, dir, env, "spec", "create", "--detailed", "Mark a task complete"); code != 0 {
		t.Fatalf("spec create --detailed exit %d: %s", code, out)
	}
	created, _ := os.ReadFile(specPathFor(t, dir))
	if !strings.Contains(string(created), "## Alternative Flows") {
		t.Fatalf("expected the detailed template, got:\n%s", created)
	}
	code, out := runGnomonOutput(t, dir, env, "workflows")
	if code != 0 || !strings.Contains(out, "implementation") || !strings.Contains(out, "bundled") {
		t.Fatalf("expected the workflow listing, got exit %d:\n%s", code, out)
	}
}
