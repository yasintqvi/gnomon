package project

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMaterialize_FreshInit_MinimalLayout(t *testing.T) {
	root := t.TempDir()
	if err := Materialize(root, false); err != nil {
		t.Fatal(err)
	}
	l := Layout{Root: root}

	for _, dir := range []string{l.ApprovalsDir(), l.SpecificationsDir()} {
		entries, err := os.ReadDir(dir)
		if err != nil || len(entries) != 0 {
			t.Fatalf("expected %s to exist and start empty, got %d entries (err=%v)", dir, len(entries), err)
		}
	}
	// No workflows, knowledge templates, contracts, or design documents by default.
	for _, absent := range []string{"workflows", "context", "contracts", "evaluations", "decisions"} {
		if _, err := os.Stat(filepath.Join(l.GnomonRoot(), absent)); !os.IsNotExist(err) {
			t.Fatalf("a plain init must not create .gnomon/%s (err=%v)", absent, err)
		}
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
	if missing := l.MissingSections(); len(missing) != 0 {
		t.Fatalf("a fresh minimal project must pass the structural check, missing: %v", missing)
	}
	if k, err := l.KnowledgeFiles(); err != nil || len(k) != 0 {
		t.Fatalf("a fresh project has no project knowledge, got %v (err=%v)", k, err)
	}
}

func TestMaterialize_Full_CopiesOptionalTemplatesButNoWorkflows(t *testing.T) {
	root := t.TempDir()
	if err := Materialize(root, true); err != nil {
		t.Fatal(err)
	}
	l := Layout{Root: root}
	for _, rel := range []string{
		filepath.Join("context", "PROJECT.md"),
		filepath.Join("context", "UI_FOUNDATION.md"),
		filepath.Join("contracts", "use-case-execution.md"),
		filepath.Join("decisions", "ADR-000-decision-title.md"),
		filepath.Join("evaluations", "SPECIFICATION_READINESS_CRITERIA.md"),
		filepath.Join("specifications", "SPECIFICATION_LIFECYCLE.md"),
	} {
		if _, err := os.Stat(filepath.Join(l.GnomonRoot(), rel)); err != nil {
			t.Fatalf("expected --full to create %s: %v", rel, err)
		}
	}
	if _, err := os.Stat(l.WorkflowsDir()); !os.IsNotExist(err) {
		t.Fatalf("workflows are read from the binary; --full must not copy them")
	}
	// Unfilled templates are never presented as project knowledge.
	if k, err := l.KnowledgeFiles(); err != nil || len(k) != 0 {
		t.Fatalf("unmodified templates must not count as knowledge, got %v (err=%v)", k, err)
	}
}

func TestMaterialize_ReRun_NeverOverwritesExistingContent(t *testing.T) {
	root := t.TempDir()
	if err := Materialize(root, true); err != nil {
		t.Fatal(err)
	}
	l := Layout{Root: root}
	customPath := filepath.Join(l.GnomonRoot(), "context", "PROJECT.md")
	if err := os.WriteFile(customPath, []byte("customized content"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := Materialize(root, true); err != nil {
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

func TestMaterialize_RefusesUnsupportedVersion(t *testing.T) {
	root := t.TempDir()
	if err := Materialize(root, false); err != nil {
		t.Fatal(err)
	}
	l := Layout{Root: root}
	if err := os.WriteFile(l.ContractVersionPath(), []byte("99"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Materialize(root, false); err == nil {
		t.Fatalf("expected Materialize to refuse an unsupported existing contract version")
	}
}

func TestLocate_WalksUpwardFromNestedDirectory(t *testing.T) {
	root := t.TempDir()
	if err := Materialize(root, false); err != nil {
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
