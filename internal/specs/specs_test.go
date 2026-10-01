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
	if _, err := CreateDraft(dir, "SPEC-002", "Second", Slugify("Second"), []byte(templateFixture)); err != nil {
		t.Fatal(err)
	}
	if _, err := CreateDraft(dir, "SPEC-001", "First", Slugify("First"), []byte(templateFixture)); err != nil {
		t.Fatal(err)
	}
	if _, err := CreateDraft(dir, "SPEC-003", "Third", Slugify("Third"), []byte(templateFixture)); err != nil {
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
	path, err := CreateDraft(dir, "SPEC-004", "Password Reset", Slugify("Password Reset"), []byte(templateFixture))
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
	if _, err := CreateDraft(dir, "SPEC-004", "Password Reset", Slugify("Password Reset"), []byte(templateFixture)); err != nil {
		t.Fatal(err)
	}
	if _, err := CreateDraft(dir, "SPEC-004", "Password Reset", Slugify("Password Reset"), []byte(templateFixture)); err == nil {
		t.Fatalf("expected an error when the destination file already exists")
	}
}

func TestCreateDraft_RefusesEmptySlug(t *testing.T) {
	dir := setupSpecsDir(t)
	if _, err := CreateDraft(dir, "SPEC-004", "Password Reset", "", []byte(templateFixture)); err == nil {
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

func realTemplate(t *testing.T) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "..", "templates", "spec-detailed.md"))
	if err != nil {
		t.Fatalf("reading real repository template: %v", err)
	}
	return data
}

func TestRemainingPlaceholders_FreshDraftFromRealTemplate(t *testing.T) {
	template := realTemplate(t)
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "SPEC-000-use-case-name.md"), template, 0o644); err != nil {
		t.Fatal(err)
	}
	path, err := CreateDraft(dir, "SPEC-001", "Mark a task complete", Slugify("Mark a task complete"), template)
	if err != nil {
		t.Fatal(err)
	}
	spec, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	remaining := RemainingPlaceholders(template, spec)
	if len(remaining) == 0 {
		t.Fatalf("expected an untouched draft to still report template placeholders")
	}
	for _, substituted := range []string{"[ID]", "[Use Case Name]", "[Use case name]"} {
		for _, tok := range remaining {
			if tok == substituted {
				t.Fatalf("expected substituted token %q not to be reported, got %v", substituted, remaining)
			}
		}
	}
	found := map[string]bool{}
	for _, tok := range remaining {
		found[tok] = true
	}
	for _, want := range []string{"[Business rule]", "[Primary actor]", "[SPEC-ID]"} {
		if !found[want] {
			t.Fatalf("expected %q among remaining placeholders, got %v", want, remaining)
		}
	}
}

func TestRemainingPlaceholders_MarkdownLinksNeverReported(t *testing.T) {
	template := []byte("See [`SPECIFICATION_LIFECYCLE.md`](../specifications/SPECIFICATION_LIFECYCLE.md) for details.\n[Business rule]\n")
	spec := []byte("See [`SPECIFICATION_LIFECYCLE.md`](../specifications/SPECIFICATION_LIFECYCLE.md) for details.\n[Business rule]\n")
	remaining := RemainingPlaceholders(template, spec)
	if len(remaining) != 1 || remaining[0] != "[Business rule]" {
		t.Fatalf("expected only [Business rule], got %v", remaining)
	}
}

func TestRemainingPlaceholders_IgnoresInlineCodeAndFencedBlocks(t *testing.T) {
	template := []byte("Use `[ID]` as the identity.\n\n```\nExample: [Not a placeholder]\n```\n\n[Business rule]\n")
	spec := []byte("Use `[ID]` as the identity.\n\n```\nExample: [Not a placeholder]\n```\n\n[Business rule]\n")
	remaining := RemainingPlaceholders(template, spec)
	if len(remaining) != 1 || remaining[0] != "[Business rule]" {
		t.Fatalf("expected only [Business rule], got %v", remaining)
	}
}

func TestRemainingPlaceholders_FullyDefinedSpec_NoPlaceholders(t *testing.T) {
	template := realTemplate(t)
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "SPEC-000-use-case-name.md"), template, 0o644); err != nil {
		t.Fatal(err)
	}
	path, err := CreateDraft(dir, "SPEC-001", "Mark a task complete", Slugify("Mark a task complete"), template)
	if err != nil {
		t.Fatal(err)
	}
	spec, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	// Replace every remaining bracketed placeholder with concrete content, simulating Define.
	defined := bracketToken.ReplaceAll(spec, []byte("defined"))

	remaining := RemainingPlaceholders(template, defined)
	if len(remaining) != 0 {
		t.Fatalf("expected no placeholders in a fully defined specification, got %v", remaining)
	}
}

func TestRemainingPlaceholders_UserWrittenBracketsNotInTemplate_NotReported(t *testing.T) {
	template := []byte("[Business rule]\n")
	spec := []byte("[Business rule]\n\nSome [optional] note the Human wrote themselves.\n")
	remaining := RemainingPlaceholders(template, spec)
	if len(remaining) != 1 || remaining[0] != "[Business rule]" {
		t.Fatalf("expected only [Business rule], got %v", remaining)
	}
}

func TestRemainingPlaceholders_CustomTemplate_DetectsItsOwnPlaceholders(t *testing.T) {
	template := []byte("# SPEC-[ID] — [Use Case Name]\n\n[Custom Field]\n\n[Another Field]\n")
	spec := []byte("# SPEC-001 — Custom Thing\n\n[Custom Field]\n\n[Another Field]\n")
	remaining := RemainingPlaceholders(template, spec)
	if len(remaining) != 2 || remaining[0] != "[Custom Field]" || remaining[1] != "[Another Field]" {
		t.Fatalf("expected both custom placeholders in order, got %v", remaining)
	}
}

func shortTemplate(t *testing.T) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "..", "templates", "spec.md"))
	if err != nil {
		t.Fatalf("reading short template: %v", err)
	}
	return data
}

func TestEffectivelyEmpty_FreshDraftFromEitherTemplate(t *testing.T) {
	for name, template := range map[string][]byte{"short": shortTemplate(t), "detailed": realTemplate(t)} {
		dir := t.TempDir()
		path, err := CreateDraft(dir, "SPEC-007", "Mark a task complete", Slugify("Mark a task complete"), template)
		if err != nil {
			t.Fatal(err)
		}
		spec, _ := os.ReadFile(path)
		if !EffectivelyEmpty([][]byte{template}, spec, "SPEC-007") {
			t.Fatalf("%s: a fresh draft must be effectively empty, authored lines: %v", name, AuthoredLines(template, spec, "SPEC-007"))
		}
		// Reformatting alone (extra blank lines, trailing spaces) authors nothing.
		reformatted := strings.ReplaceAll(string(spec), "\n", "  \n\n")
		if !EffectivelyEmpty([][]byte{template}, []byte(reformatted), "SPEC-007") {
			t.Fatalf("%s: whitespace-only changes must not count as content", name)
		}
	}
}

func TestEffectivelyEmpty_OneRealCriterionIsContent(t *testing.T) {
	template := shortTemplate(t)
	dir := t.TempDir()
	path, err := CreateDraft(dir, "SPEC-001", "Mark a task complete", Slugify("Mark a task complete"), template)
	if err != nil {
		t.Fatal(err)
	}
	spec, _ := os.ReadFile(path)
	edited := strings.Replace(string(spec), "1. [Given a situation, when something happens, then this is the result.]", "1. An open task can be marked complete.", 1)
	if EffectivelyEmpty([][]byte{template}, []byte(edited), "SPEC-001") {
		t.Fatalf("a Specification with a real acceptance criterion is not empty")
	}
	if got := AuthoredLines(template, []byte(edited), "SPEC-001"); len(got) != 1 || got[0] != "1. An open task can be marked complete." {
		t.Fatalf("expected exactly the authored criterion, got %v", got)
	}
}

func TestEffectivelyEmpty_ChecksEveryCandidateTemplate(t *testing.T) {
	// A Specification created from an older project template is still recognized as empty when
	// that template is among the candidates, even though it doesn't match the bundled ones.
	legacy := []byte(templateFixture)
	dir := t.TempDir()
	path, err := CreateDraft(dir, "SPEC-004", "Submit a Report", Slugify("Submit a Report"), legacy)
	if err != nil {
		t.Fatal(err)
	}
	spec, _ := os.ReadFile(path)
	if EffectivelyEmpty([][]byte{shortTemplate(t)}, spec, "SPEC-004") {
		t.Fatalf("against an unrelated template, the legacy draft's text counts as authored")
	}
	if !EffectivelyEmpty([][]byte{shortTemplate(t), legacy}, spec, "SPEC-004") {
		t.Fatalf("expected the legacy draft to be recognized as empty against its own template")
	}
}

func TestClosestTemplate_PicksTheOriginatingTemplate(t *testing.T) {
	short, detailed := shortTemplate(t), realTemplate(t)
	dir := t.TempDir()
	path, err := CreateDraft(dir, "SPEC-001", "Thing", Slugify("Thing"), detailed)
	if err != nil {
		t.Fatal(err)
	}
	spec, _ := os.ReadFile(path)
	if got := ClosestTemplate([][]byte{short, detailed}, spec, "SPEC-001"); string(got) != string(detailed) {
		t.Fatalf("expected the detailed template to be identified as the origin")
	}
}
