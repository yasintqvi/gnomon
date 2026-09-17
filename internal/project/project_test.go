package project

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMaterialize_FreshInit_CanonicalLayout(t *testing.T) {
	root := t.TempDir()
	if err := Materialize(root); err != nil {
		t.Fatal(err)
	}
	l := Layout{Root: root}

	expectFile := func(rel string) {
		t.Helper()
		if _, err := os.Stat(filepath.Join(l.GnomonRoot(), rel)); err != nil {
			t.Fatalf("expected %s to exist: %v", rel, err)
		}
	}
	expectFile(filepath.Join("workflows", "implementation.md"))
	expectFile(filepath.Join("specifications", "SPEC-000-use-case-name.md"))
	expectFile(filepath.Join("specifications", "SPECIFICATION_LIFECYCLE.md"))
	expectFile(filepath.Join("evaluations", "SPECIFICATION_READINESS_CRITERIA.md"))
	expectFile(filepath.Join("contracts", "use-case-execution.md"))
	expectFile(filepath.Join("decisions", "ADR-000-decision-title.md"))
	expectFile(filepath.Join("context", "PROJECT.md"))

	if info, err := os.Stat(l.ApprovalsDir()); err != nil || !info.IsDir() {
		t.Fatalf("expected approvals/ to exist as an empty directory")
	}
	entries, err := os.ReadDir(l.ApprovalsDir())
	if err != nil || len(entries) != 0 {
		t.Fatalf("expected approvals/ to start empty, got %v entries (err=%v)", len(entries), err)
	}

	v, err := l.ContractVersion()
	if err != nil {
		t.Fatal(err)
	}
	if v != SupportedContractVersion {
		t.Fatalf("expected CONTRACT_VERSION=%s, got %q", SupportedContractVersion, v)
	}
	if !l.IsInitialized() {
		t.Fatalf("expected project to be reported as initialized")
	}
}

func TestMaterialize_ReRun_NeverOverwritesExistingContent(t *testing.T) {
	root := t.TempDir()
	if err := Materialize(root); err != nil {
		t.Fatal(err)
	}
	l := Layout{Root: root}
	customPath := filepath.Join(l.WorkflowsDir(), "implementation.md")
	if err := os.WriteFile(customPath, []byte("customized content"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := Materialize(root); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(customPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "customized content" {
		t.Fatalf("re-running init must never overwrite existing content, got: %q", string(data))
	}
}

func TestMaterialize_FillsOnlyMissingParts(t *testing.T) {
	root := t.TempDir()
	if err := Materialize(root); err != nil {
		t.Fatal(err)
	}
	l := Layout{Root: root}
	victim := filepath.Join(l.WorkflowsDir(), "testing.md")
	if err := os.Remove(victim); err != nil {
		t.Fatal(err)
	}

	if err := Materialize(root); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(victim); err != nil {
		t.Fatalf("expected a manually removed bundled file to be restored: %v", err)
	}
}

func TestMaterialize_RefusesUnsupportedVersion(t *testing.T) {
	root := t.TempDir()
	if err := Materialize(root); err != nil {
		t.Fatal(err)
	}
	l := Layout{Root: root}
	if err := os.WriteFile(l.ContractVersionPath(), []byte("99"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Materialize(root); err == nil {
		t.Fatalf("expected Materialize to refuse an unsupported existing contract version")
	}
}

func TestLocate_WalksUpwardFromNestedDirectory(t *testing.T) {
	root := t.TempDir()
	if err := Materialize(root); err != nil {
		t.Fatal(err)
	}
	nested := filepath.Join(root, "a", "b", "c")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	l, err := Locate(nested)
	if err != nil {
		t.Fatal(err)
	}
	resolvedRoot, _ := filepath.EvalSymlinks(root)
	resolvedFound, _ := filepath.EvalSymlinks(l.Root)
	if resolvedFound != resolvedRoot {
		t.Fatalf("expected Locate to find %s, got %s", resolvedRoot, resolvedFound)
	}
}

func TestLocate_NotFound(t *testing.T) {
	if _, err := Locate(t.TempDir()); err == nil {
		t.Fatalf("expected an error when no .gnomon directory exists above start")
	}
}
