package orchestrate

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gnomon/internal/project"
)

func TestTitleFromRequest(t *testing.T) {
	for in, want := range map[string]string{
		"Late fees":                                  "Late fees",
		"Add late fees. 1 euro per day late.":        "Add late fees",
		"Late fees: a tool returned late gets a fee": "Late fees",
		"Can members reserve tools? They want to.":   "Can members reserve tools",
		"Members picking up a reserved tool at the self-service kiosk should always get it even with three tools out": "Members picking up a reserved tool at the self-service kiosk",
		"First line\nsecond line": "First line",
	} {
		if got := titleFromRequest(in); got != want {
			t.Errorf("titleFromRequest(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestSavedRequest_RoundTripsMultiline(t *testing.T) {
	path := filepath.Join(t.TempDir(), "SPEC-001-x.md")
	if err := os.WriteFile(path, []byte("# SPEC-001 — X\n\n## Goal\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	request := "Add late fees.\nRequired interface: loan gains \"fee\"."
	if err := saveRequestInDraft(path, request); err != nil {
		t.Fatal(err)
	}
	content, _ := os.ReadFile(path)
	if got := savedRequest(string(content)); got != request {
		t.Fatalf("saved request %q, want %q\n%s", got, request, content)
	}
	if !strings.HasPrefix(string(content), "# SPEC-001 — X\n\n> Request: Add late fees.\n> Required interface") {
		t.Fatalf("expected the request directly below the title:\n%s", content)
	}
}

// No Draft is created while another run is active: the refusal comes first.
func TestDefine_NewRequestWhileRunActive_CreatesNothing(t *testing.T) {
	root := setupApprovedSpec(t)
	release, err := acquireRunLock(root, "run-A", "implementation", "SPEC-001")
	if err != nil {
		t.Fatal(err)
	}
	defer release()

	rep, err := Define(root, []string{"Late fees"}, DefineOptions{}, "", nil)
	var active *RunActiveError
	if !errors.As(err, &active) || rep == nil || !strings.Contains(sectionBody(rep, "Active Run"), "run-A") {
		t.Fatalf("expected a refusal naming the active run, got %v %+v", err, rep)
	}
	l, _ := project.Locate(root)
	matches, _ := filepath.Glob(filepath.Join(l.SpecificationsDir(), "SPEC-002-*.md"))
	if len(matches) != 0 {
		t.Fatalf("no Draft may be created while a run is active: %v", matches)
	}
}

func TestDefine_NoArgumentsRefused(t *testing.T) {
	root := setupApprovedSpec(t)
	if _, err := Define(root, []string{"  "}, DefineOptions{}, "", nil); err == nil {
		t.Fatal("an empty request must be refused")
	}
}
