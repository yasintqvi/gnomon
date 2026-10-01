package project

import (
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"
)

// TestStockFingerprints_CoverCurrentBundle fails when a shipped file changes without regenerating
// stock_fingerprints.go (`go run ./tools/stockgen`). Without it, a project holding an unmodified
// copy of the current version would be treated as customized after the next change.
func TestStockFingerprints_CoverCurrentBundle(t *testing.T) {
	err := fs.WalkDir(bundled, bundledRoot, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		data, err := fs.ReadFile(bundled, p)
		if err != nil {
			return err
		}
		if !IsStock(path.Base(p), data) {
			t.Errorf("%s is not in stock_fingerprints.go; run `go run ./tools/stockgen`", p)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestIsStock_IgnoresLineEndingsAndTrailingWhitespace(t *testing.T) {
	data, err := fs.ReadFile(bundled, path.Join(bundledRoot, "workflows", "implementation.md"))
	if err != nil {
		t.Fatal(err)
	}
	crlf := strings.ReplaceAll(string(data), "\n", "  \r\n")
	if !IsStock("implementation.md", []byte(crlf)) {
		t.Fatalf("a Windows checkout of a shipped file must still count as unmodified")
	}
	if IsStock("implementation.md", append(data, []byte("\nOne project-specific rule.\n")...)) {
		t.Fatalf("an edited file must count as a customization")
	}
	if IsStock("my-workflow.md", data) {
		t.Fatalf("stock status is per shipped filename")
	}
}

func TestSpecTemplate_ShortByDefault_DetailedOnRequest_CustomizedHonored(t *testing.T) {
	root := t.TempDir()
	if err := Materialize(root, false); err != nil {
		t.Fatal(err)
	}
	l := Layout{Root: root}
	short, _ := l.SpecTemplate(false)
	detailed, _ := l.SpecTemplate(true)
	if !strings.Contains(string(short), "## Decisions") || strings.Count(string(short), "\n") > 30 {
		t.Fatalf("expected the short default template, got:\n%s", short)
	}
	if !strings.Contains(string(detailed), "## Alternative Flows") {
		t.Fatalf("expected the detailed template")
	}

	// An unmodified legacy SPEC-000 copy doesn't override the short default...
	legacy := filepath.Join(l.SpecificationsDir(), legacySpecTemplate)
	if err := os.WriteFile(legacy, detailedHistorical(t), 0o644); err != nil {
		t.Fatal(err)
	}
	if got, _ := l.SpecTemplate(false); string(got) != string(short) {
		t.Fatalf("an unmodified legacy template copy must not replace the short default")
	}
	// ...but an edited one is the project's own template.
	if err := os.WriteFile(legacy, []byte("# SPEC-[ID] — [Use Case Name]\n\n[Our field]\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got, _ := l.SpecTemplate(false); !strings.Contains(string(got), "[Our field]") {
		t.Fatalf("expected the customized project template")
	}
}

// detailedHistorical returns a SPEC-000 version an older init shipped — any entry in the
// fingerprint set qualifies, so the fixture is the legacy test project's copy.
func detailedHistorical(t *testing.T) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "orchestrate", "testdata", "legacy", ".gnomon", "specifications", legacySpecTemplate))
	if err != nil {
		t.Fatal(err)
	}
	if !IsStock(legacySpecTemplate, data) {
		t.Fatalf("fixture should be a shipped version")
	}
	return data
}
