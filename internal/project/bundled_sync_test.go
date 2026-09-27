package project

import (
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"testing"
)

// bundledDirNames returns the top-level directory names under the bundled tree — derived from the
// embedded FS itself, never hard-coded, so a newly added bundled category (context, contracts,
// decisions, evaluations, specifications, workflows, ...) is covered automatically.
func bundledDirNames(t *testing.T) []string {
	t.Helper()
	entries, err := fs.ReadDir(bundled, bundledRoot)
	if err != nil {
		t.Fatal(err)
	}
	var dirs []string
	for _, e := range entries {
		if e.IsDir() {
			dirs = append(dirs, e.Name())
		}
	}
	return dirs
}

// TestBundledFiles_MatchCanonicalCore proves every embedded bundled file, across every bundled
// category, is byte-for-byte identical to the real, canonical file at the repo root it is meant to
// mirror. The two exist as separate files specifically because //go:embed patterns cannot
// reference paths outside their own package directory (Core's own directories live outside
// internal/project/), and nothing else guards against the two silently drifting apart as either is
// edited — this test is that guard, reading the bundle through the real embed.FS a compiled binary
// would actually ship, not a second copy of the comparison logic.
func TestBundledFiles_MatchCanonicalCore(t *testing.T) {
	dirs := bundledDirNames(t)
	checked := 0
	for _, dir := range dirs {
		coreDir := filepath.Join("..", "..", dir)
		entries, err := os.ReadDir(coreDir)
		if err != nil {
			t.Fatalf("reading canonical directory %s: %v", coreDir, err)
		}
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			checked++

			corePath := filepath.Join(coreDir, e.Name())
			coreContent, err := os.ReadFile(corePath)
			if err != nil {
				t.Fatalf("reading canonical %s: %v", corePath, err)
			}

			bundledPath := path.Join(bundledRoot, dir, e.Name())
			bundledContent, err := fs.ReadFile(bundled, bundledPath)
			if err != nil {
				t.Fatalf("%s/%s exists at the repo root but has no bundled counterpart embedded at %s: %v", dir, e.Name(), bundledPath, err)
			}

			if string(coreContent) != string(bundledContent) {
				t.Fatalf("%s/%s has drifted from its bundled copy at internal/project/%s — the two must stay byte-for-byte identical", dir, e.Name(), bundledPath)
			}
		}
	}
	if checked == 0 {
		t.Fatalf("expected to check at least one file across bundled categories %v — did the bundled tree change unexpectedly?", dirs)
	}
}

// TestBundledFiles_NoOrphanedBundledFiles proves the reverse direction too, across every bundled
// category: every bundled file corresponds to a real canonical file at the repo root, so a file
// removed from the repo root can never silently keep shipping a stale bundled copy to newly
// initialized projects.
func TestBundledFiles_NoOrphanedBundledFiles(t *testing.T) {
	for _, dir := range bundledDirNames(t) {
		bundledEntries, err := fs.ReadDir(bundled, path.Join(bundledRoot, dir))
		if err != nil {
			t.Fatalf("reading bundled directory %s: %v", dir, err)
		}
		for _, e := range bundledEntries {
			if e.IsDir() {
				continue
			}
			corePath := filepath.Join("..", "..", dir, e.Name())
			if _, err := os.Stat(corePath); err != nil {
				t.Fatalf("bundled/%s/%s has no canonical counterpart at %s", dir, e.Name(), corePath)
			}
		}
	}
}
