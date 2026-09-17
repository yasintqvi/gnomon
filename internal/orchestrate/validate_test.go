package orchestrate

import (
	"os"
	"path/filepath"
	"testing"

	"gnomon/internal/present"
	"gnomon/internal/project"
)

func TestValidate_CleanProject_Success(t *testing.T) {
	root := t.TempDir()
	configureGitIdentity(t, root)
	if _, err := Init(root); err != nil {
		t.Fatal(err)
	}

	report, err := Validate(root)
	if err != nil {
		t.Fatalf("expected a freshly-initialized project to validate cleanly: %v (report: %+v)", err, report)
	}
	if report.Outcome != present.Success {
		t.Fatalf("expected Success, got %v", report.Outcome)
	}
}

func TestValidate_NeverInvokesAnAgent(t *testing.T) {
	// No --agent, no chooser, no agentconfig exist anywhere in Validate's own signature — this is
	// a structural guarantee, confirmed here simply by the fact that a call with no default Agent
	// provider configured, and no real Agent installed, still succeeds (nothing was ever resolved
	// or invoked).
	sandboxAgentConfig(t)
	root := t.TempDir()
	configureGitIdentity(t, root)
	if _, err := Init(root); err != nil {
		t.Fatal(err)
	}
	if _, err := Validate(root); err != nil {
		t.Fatalf("expected Validate to succeed without ever needing an Agent provider: %v", err)
	}
}

func TestValidate_MissingCanonicalDirectory_Detected(t *testing.T) {
	root := t.TempDir()
	configureGitIdentity(t, root)
	if _, err := Init(root); err != nil {
		t.Fatal(err)
	}
	l, err := project.Locate(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.RemoveAll(filepath.Join(l.GnomonRoot(), "contracts")); err != nil {
		t.Fatal(err)
	}

	report, err := Validate(root)
	if err == nil {
		t.Fatalf("expected a missing canonical directory to be detected")
	}
	if report.Outcome != present.Failed {
		t.Fatalf("expected Failed, got %v", report.Outcome)
	}
}

func TestValidate_DuplicateWorkflowIdentity_Detected(t *testing.T) {
	root := t.TempDir()
	configureGitIdentity(t, root)
	if _, err := Init(root); err != nil {
		t.Fatal(err)
	}
	l, err := project.Locate(root)
	if err != nil {
		t.Fatal(err)
	}
	// A second file declaring the exact same identity as the real bootstrap.md.
	duplicate := "---\nidentity: bootstrap\nspecification_reference: none\nresult:\n" +
		"  terminal_path: outcome\n  classification:\n    OK: success\n  schema:\n    type: object\n" +
		"    required: [outcome]\n    properties:\n      outcome:\n        type: string\n        enum: [OK]\n" +
		"---\n# Duplicate\n"
	if err := os.WriteFile(filepath.Join(l.WorkflowsDir(), "bootstrap-copy.md"), []byte(duplicate), 0o644); err != nil {
		t.Fatal(err)
	}

	report, err := Validate(root)
	if err == nil {
		t.Fatalf("expected a duplicate workflow identity to be detected")
	}
	if report.Outcome != present.Failed {
		t.Fatalf("expected Failed, got %v", report.Outcome)
	}
}

func TestValidate_MalformedWorkflowFrontmatter_Detected(t *testing.T) {
	root := t.TempDir()
	configureGitIdentity(t, root)
	if _, err := Init(root); err != nil {
		t.Fatal(err)
	}
	l, err := project.Locate(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(l.WorkflowsDir(), "broken.md"), []byte("---\nidentity: [unclosed\n---\n# Broken\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	report, err := Validate(root)
	if err == nil {
		t.Fatalf("expected a malformed workflow frontmatter to be detected")
	}
	if report.Outcome != present.Failed {
		t.Fatalf("expected Failed, got %v", report.Outcome)
	}
}

func TestValidate_MalformedApprovalEvidence_Detected(t *testing.T) {
	root := t.TempDir()
	configureGitIdentity(t, root)
	if _, err := Init(root); err != nil {
		t.Fatal(err)
	}
	l, err := project.Locate(root)
	if err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(l.ApprovalsDir(), "SPEC-001")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "bad.grant.json"), []byte("not json"), 0o644); err != nil {
		t.Fatal(err)
	}

	report, err := Validate(root)
	if err == nil {
		t.Fatalf("expected malformed approval evidence to be detected")
	}
	if report.Outcome != present.Failed {
		t.Fatalf("expected Failed, got %v", report.Outcome)
	}
}

func TestValidate_NeverMutatesProjectFiles(t *testing.T) {
	root := t.TempDir()
	configureGitIdentity(t, root)
	if _, err := Init(root); err != nil {
		t.Fatal(err)
	}
	if _, err := SpecCreate(root, "Password Reset"); err != nil {
		t.Fatal(err)
	}
	if _, err := Approve(root, "SPEC-001", nil); err != nil {
		t.Fatal(err)
	}

	l, err := project.Locate(root)
	if err != nil {
		t.Fatal(err)
	}

	// Snapshot every file's content under .gnomon/ before validating.
	before := map[string]string{}
	_ = filepath.Walk(l.GnomonRoot(), func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		data, _ := os.ReadFile(path)
		before[path] = string(data)
		return nil
	})

	if _, err := Validate(root); err != nil {
		t.Fatalf("expected a valid project to validate cleanly: %v", err)
	}

	for path, want := range before {
		got, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("expected %s to still exist after Validate: %v", path, err)
		}
		if string(got) != want {
			t.Fatalf("expected Validate to never modify %s, but its content changed", path)
		}
	}
}
