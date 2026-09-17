package project

import (
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"testing"
)

// TestBundledWorkflows_MatchCanonicalCore proves the embedded bundle gnomon init actually
// materializes into a new project is byte-for-byte identical to the canonical Core workflow files
// it is meant to mirror. The two exist as separate files specifically because //go:embed patterns
// cannot reference paths outside their own package directory (Core's workflows/ lives outside
// internal/project/), and nothing else guards against the two silently drifting apart as either
// is edited — this test is that guard, reading the bundle through the real embed.FS a compiled
// binary would actually ship, not a second copy of the comparison logic.
func TestBundledWorkflows_MatchCanonicalCore(t *testing.T) {
	coreDir := filepath.Join("..", "..", "workflows")
	entries, err := os.ReadDir(coreDir)
	if err != nil {
		t.Fatal(err)
	}

	checked := 0
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".md" {
			continue
		}
		checked++

		corePath := filepath.Join(coreDir, e.Name())
		coreContent, err := os.ReadFile(corePath)
		if err != nil {
			t.Fatalf("reading canonical %s: %v", corePath, err)
		}

		bundledPath := path.Join(bundledRoot, "workflows", e.Name())
		bundledContent, err := fs.ReadFile(bundled, bundledPath)
		if err != nil {
			t.Fatalf("workflows/%s exists in Core but has no bundled counterpart embedded at %s: %v", e.Name(), bundledPath, err)
		}

		if string(coreContent) != string(bundledContent) {
			t.Fatalf("workflows/%s has drifted from its bundled copy at internal/project/%s — the two must stay byte-for-byte identical", e.Name(), bundledPath)
		}
	}
	if checked != 10 {
		t.Fatalf("expected to check all 10 built-in workflow files, checked %d — did workflows/ change without this test being updated?", checked)
	}
}

// TestBundledWorkflows_NoOrphanedBundledFiles proves the reverse direction too: every bundled
// workflow file corresponds to a real canonical Core file, so a workflow removed from Core can
// never silently keep shipping a stale bundled copy to newly-initialized projects.
func TestBundledWorkflows_NoOrphanedBundledFiles(t *testing.T) {
	bundledEntries, err := fs.ReadDir(bundled, path.Join(bundledRoot, "workflows"))
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range bundledEntries {
		if e.IsDir() {
			continue
		}
		corePath := filepath.Join("..", "..", "workflows", e.Name())
		if _, err := os.Stat(corePath); err != nil {
			t.Fatalf("bundled/workflows/%s has no canonical Core counterpart at %s", e.Name(), corePath)
		}
	}
}
