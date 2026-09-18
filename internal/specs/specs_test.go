package specs

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const templateFixture = `# SPEC-[ID] — [Use Case Name]

## Use Case

**Name**

[Use case name]

## Dependencies

None.
`

func setupSpecsDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "SPEC-000-use-case-name.md"), []byte(templateFixture), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestList_EmptyDir_ExcludesTemplate(t *testing.T) {
	dir := setupSpecsDir(t)
	ids, err := List(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 0 {
		t.Fatalf("expected no Specifications (SPEC-000 template excluded), got %v", ids)
	}
}

func TestList_ReturnsAllIdentitiesInOrder(t *testing.T) {
	dir := setupSpecsDir(t)
	if _, err := CreateDraft(dir, "SPEC-002", "Second", Slugify("Second")); err != nil {
		t.Fatal(err)
	}
	if _, err := CreateDraft(dir, "SPEC-001", "First", Slugify("First")); err != nil {
		t.Fatal(err)
	}
	if _, err := CreateDraft(dir, "SPEC-003", "Third", Slugify("Third")); err != nil {
		t.Fatal(err)
	}

	ids, err := List(dir)
	if err != nil {
		t.Fatal(err)
	}
	want := []Identity{"SPEC-001", "SPEC-002", "SPEC-003"}
	if len(ids) != len(want) {
		t.Fatalf("expected %v, got %v", want, ids)
	}
	for i := range want {
		if ids[i] != want[i] {
			t.Fatalf("expected %v, got %v", want, ids)
		}
	}
}

func TestNextIdentity_EmptyDir(t *testing.T) {
	dir := setupSpecsDir(t)
	id, err := NextIdentity(dir)
	if err != nil {
		t.Fatal(err)
	}
	if id != "SPEC-001" {
		t.Fatalf("expected SPEC-001, got %s", id)
	}
}

func TestNextIdentity_SkipsTemplateAndFindsMax(t *testing.T) {
	dir := setupSpecsDir(t)
	for _, name := range []string{"SPEC-001-a.md", "SPEC-004-b.md", "SPEC-002-c.md"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	id, err := NextIdentity(dir)
	if err != nil {
		t.Fatal(err)
	}
	if id != "SPEC-005" {
		t.Fatalf("expected SPEC-005, got %s", id)
	}
}

func TestExists(t *testing.T) {
	dir := setupSpecsDir(t)
	if err := os.WriteFile(filepath.Join(dir, "SPEC-003-password-reset.md"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	ok, path, err := Exists(dir, "SPEC-003")
	if err != nil {
		t.Fatal(err)
	}
	if !ok || !strings.HasSuffix(path, "SPEC-003-password-reset.md") {
		t.Fatalf("expected SPEC-003 to exist, got ok=%v path=%s", ok, path)
	}
	ok, _, err = Exists(dir, "SPEC-999")
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatalf("expected SPEC-999 not to exist")
	}
}

func TestCreateDraft_SubstitutesOnlyIdentityAndTitle(t *testing.T) {
	dir := setupSpecsDir(t)
	path, err := CreateDraft(dir, "SPEC-004", "Password Reset", Slugify("Password Reset"))
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(path) != "SPEC-004-password-reset.md" {
		t.Fatalf("unexpected filename: %s", filepath.Base(path))
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if !strings.Contains(content, "# SPEC-004 — Password Reset") {
		t.Fatalf("expected substituted title heading, got:\n%s", content)
	}
	if strings.Contains(content, "[ID]") || strings.Contains(content, "[Use Case Name]") {
		t.Fatalf("expected identity/title placeholders to be fully substituted, got:\n%s", content)
	}
	if !strings.Contains(content, "None.") {
		t.Fatalf("expected untouched Dependencies section to survive unmodified")
	}
}

func TestCreateDraft_RefusesToOverwriteExisting(t *testing.T) {
	dir := setupSpecsDir(t)
	if _, err := CreateDraft(dir, "SPEC-004", "Password Reset", Slugify("Password Reset")); err != nil {
		t.Fatal(err)
	}
	if _, err := CreateDraft(dir, "SPEC-004", "Password Reset", Slugify("Password Reset")); err == nil {
		t.Fatalf("expected an error when the destination file already exists")
	}
}

func TestCreateDraft_RefusesEmptySlug(t *testing.T) {
	dir := setupSpecsDir(t)
	if _, err := CreateDraft(dir, "SPEC-004", "Password Reset", ""); err == nil {
		t.Fatalf("expected an empty slug to be refused")
	}
}

func TestSlugify_MechanicalOnly_NoStopWordRemoval(t *testing.T) {
	// Slugify never applies English-specific stop-word/NLP heuristics — "a" survives verbatim.
	if got := Slugify("Create a Project"); got != "create-a-project" {
		t.Fatalf("expected purely mechanical normalization, got %q", got)
	}
}

func TestTitle_ExtractsFromHeading(t *testing.T) {
	content := "# SPEC-001 — Password Reset\n\n## Use Case\n"
	if got := Title("SPEC-001", []byte(content)); got != "Password Reset" {
		t.Fatalf("expected extracted title, got %q", got)
	}
}

func TestTitle_FallsBackToIdentity_WhenHeadingMissingOrMalformed(t *testing.T) {
	if got := Title("SPEC-002", []byte("no heading here\n")); got != "SPEC-002" {
		t.Fatalf("expected fallback to identity, got %q", got)
	}
	if got := Title("SPEC-003", []byte("")); got != "SPEC-003" {
		t.Fatalf("expected fallback to identity for empty content, got %q", got)
	}
}
