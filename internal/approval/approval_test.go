package approval

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"
)

func readGrantIDs(approvalsDir, specID string) ([]string, error) {
	entries, err := os.ReadDir(specDir(approvalsDir, specID))
	if err != nil {
		return nil, err
	}
	var ids []string
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".grant.json") {
			ids = append(ids, strings.TrimSuffix(e.Name(), ".grant.json"))
		}
	}
	return ids, nil
}

func writeFileHelper(dir, name, content string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644)
}

func TestFingerprint_IgnoresWhitespaceOnlyChanges(t *testing.T) {
	a := "# Title\n\nSome content.  \nMore content.\n"
	b := "# Title\n\n\n\nSome content.\nMore content.\n\n" // trailing spaces removed, extra blank lines
	if Fingerprint([]byte(a)) != Fingerprint([]byte(b)) {
		t.Fatalf("expected whitespace-only difference to produce the same fingerprint")
	}
}

func TestFingerprint_ChangesOnRealContentChange(t *testing.T) {
	a := "# Title\n\nSome content.\n"
	b := "# Title\n\nDifferent content.\n"
	if Fingerprint([]byte(a)) == Fingerprint([]byte(b)) {
		t.Fatalf("expected a real content change to produce a different fingerprint")
	}
}

func TestDerive_NoEvidence_Draft(t *testing.T) {
	dir := t.TempDir()
	state, err := Derive(dir, "SPEC-001", "abc")
	if err != nil {
		t.Fatal(err)
	}
	if state != Draft {
		t.Fatalf("expected Draft, got %s", state)
	}
}

func TestDerive_MatchingGrant_Approved(t *testing.T) {
	dir := t.TempDir()
	fp := "the-fingerprint"
	if err := WriteGrant(dir, "SPEC-001", fp, "Jane <jane@example.com>", []byte("content")); err != nil {
		t.Fatal(err)
	}
	state, err := Derive(dir, "SPEC-001", fp)
	if err != nil {
		t.Fatal(err)
	}
	if state != Approved {
		t.Fatalf("expected Approved, got %s", state)
	}
}

func TestDerive_NonMatchingFingerprint_Draft(t *testing.T) {
	dir := t.TempDir()
	if err := WriteGrant(dir, "SPEC-001", "old-fingerprint", "Jane <jane@example.com>", []byte("content")); err != nil {
		t.Fatal(err)
	}
	state, err := Derive(dir, "SPEC-001", "new-fingerprint")
	if err != nil {
		t.Fatal(err)
	}
	if state != Draft {
		t.Fatalf("expected Draft after content changed, got %s", state)
	}
}

func TestDerive_RevokedGrant_Draft(t *testing.T) {
	dir := t.TempDir()
	fp := "the-fingerprint"
	if err := WriteGrant(dir, "SPEC-001", fp, "Jane <jane@example.com>", []byte("content")); err != nil {
		t.Fatal(err)
	}
	entries, err := readGrantIDs(dir, "SPEC-001")
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected exactly one grant id, got %d", len(entries))
	}
	if err := WriteRevocation(dir, "SPEC-001", entries[0], "Jane <jane@example.com>"); err != nil {
		t.Fatal(err)
	}
	state, err := Derive(dir, "SPEC-001", fp)
	if err != nil {
		t.Fatal(err)
	}
	if state != Draft {
		t.Fatalf("expected Draft after revocation even though content still matches, got %s", state)
	}
}

func TestDerive_ReapprovalAfterRevocation_Approved(t *testing.T) {
	dir := t.TempDir()
	fp := "the-fingerprint"
	if err := WriteGrant(dir, "SPEC-001", fp, "Jane <jane@example.com>", []byte("content")); err != nil {
		t.Fatal(err)
	}
	ids, _ := readGrantIDs(dir, "SPEC-001")
	if err := WriteRevocation(dir, "SPEC-001", ids[0], "Jane <jane@example.com>"); err != nil {
		t.Fatal(err)
	}
	// A genuinely new grant, even for identical content, must revalidate the Specification.
	if err := WriteGrant(dir, "SPEC-001", fp, "Jane <jane@example.com>", []byte("content")); err != nil {
		t.Fatal(err)
	}
	state, err := Derive(dir, "SPEC-001", fp)
	if err != nil {
		t.Fatal(err)
	}
	if state != Approved {
		t.Fatalf("expected Approved after a fresh re-grant, got %s", state)
	}
}

// --- ActiveGrant: "the latest approval decision counts" ---

func TestActiveGrant_NoGrants_NotApprovedNoID(t *testing.T) {
	dir := t.TempDir()
	id, approved, err := ActiveGrant(dir, "SPEC-001", "fp")
	if err != nil {
		t.Fatal(err)
	}
	if id != "" || approved {
		t.Fatalf("expected no active grant, got id=%q approved=%v", id, approved)
	}
}

func TestActiveGrant_OlderContentRestoredAfterNewerApproval_Draft(t *testing.T) {
	dir := t.TempDir()
	if err := WriteGrant(dir, "SPEC-001", "fp-v1", "Jane <jane@example.com>", []byte("v1")); err != nil {
		t.Fatal(err)
	}
	if err := WriteGrant(dir, "SPEC-001", "fp-v2", "Jane <jane@example.com>", []byte("v2")); err != nil {
		t.Fatal(err)
	}
	state, err := Derive(dir, "SPEC-001", "fp-v1")
	if err != nil {
		t.Fatal(err)
	}
	if state != Draft {
		t.Fatalf("expected Draft when content reverts to an older, non-latest grant, got %s", state)
	}
}

func TestActiveGrant_LatestRevoked_ContentRevertedToOlder_Draft(t *testing.T) {
	dir := t.TempDir()
	if err := WriteGrant(dir, "SPEC-001", "fp-v1", "Jane <jane@example.com>", []byte("v1")); err != nil {
		t.Fatal(err)
	}
	if err := WriteGrant(dir, "SPEC-001", "fp-v2", "Jane <jane@example.com>", []byte("v2")); err != nil {
		t.Fatal(err)
	}
	revs, err := Revisions(dir, "SPEC-001")
	if err != nil {
		t.Fatal(err)
	}
	latest := revs[len(revs)-1]
	if latest.Fingerprint != "fp-v2" {
		t.Fatalf("expected v2's grant to be latest, got %+v", latest)
	}
	if err := WriteRevocation(dir, "SPEC-001", latest.GrantID, "Jane <jane@example.com>"); err != nil {
		t.Fatal(err)
	}
	state, err := Derive(dir, "SPEC-001", "fp-v1")
	if err != nil {
		t.Fatal(err)
	}
	if state != Draft {
		t.Fatalf("expected Draft: revoking the latest grant must never resurrect an older one, got %s", state)
	}
}

func TestActiveGrant_LatestRevoked_ContentStillMatchesRevokedGrant_Draft(t *testing.T) {
	dir := t.TempDir()
	if err := WriteGrant(dir, "SPEC-001", "fp-v1", "Jane <jane@example.com>", []byte("v1")); err != nil {
		t.Fatal(err)
	}
	if err := WriteGrant(dir, "SPEC-001", "fp-v2", "Jane <jane@example.com>", []byte("v2")); err != nil {
		t.Fatal(err)
	}
	revs, err := Revisions(dir, "SPEC-001")
	if err != nil {
		t.Fatal(err)
	}
	latest := revs[len(revs)-1]
	if err := WriteRevocation(dir, "SPEC-001", latest.GrantID, "Jane <jane@example.com>"); err != nil {
		t.Fatal(err)
	}
	state, err := Derive(dir, "SPEC-001", "fp-v2")
	if err != nil {
		t.Fatal(err)
	}
	if state != Draft {
		t.Fatalf("expected Draft: L is revoked, even though content still matches it, got %s", state)
	}
}

func TestActiveGrant_ContentEditedThenRestoredExactly_Approved(t *testing.T) {
	dir := t.TempDir()
	if err := WriteGrant(dir, "SPEC-001", "fp-v1", "Jane <jane@example.com>", []byte("v1")); err != nil {
		t.Fatal(err)
	}
	state, err := Derive(dir, "SPEC-001", "fp-edited")
	if err != nil {
		t.Fatal(err)
	}
	if state != Draft {
		t.Fatalf("expected Draft while content is edited away from L, got %s", state)
	}
	state, err = Derive(dir, "SPEC-001", "fp-v1")
	if err != nil {
		t.Fatal(err)
	}
	if state != Approved {
		t.Fatalf("expected Approved once content is restored to exactly L's content, got %s", state)
	}
}

func TestActiveGrant_ApproveAgainAfterRevertingToOlderContent_NewGrantBecomesLatest(t *testing.T) {
	dir := t.TempDir()
	if err := WriteGrant(dir, "SPEC-001", "fp-v1", "Jane <jane@example.com>", []byte("v1")); err != nil {
		t.Fatal(err)
	}
	if err := WriteGrant(dir, "SPEC-001", "fp-v2", "Jane <jane@example.com>", []byte("v2")); err != nil {
		t.Fatal(err)
	}
	// Content reverted to v1; a fresh approval for it must become the new latest grant.
	if err := WriteGrant(dir, "SPEC-001", "fp-v1", "Jane <jane@example.com>", []byte("v1 again")); err != nil {
		t.Fatal(err)
	}
	state, err := Derive(dir, "SPEC-001", "fp-v1")
	if err != nil {
		t.Fatal(err)
	}
	if state != Approved {
		t.Fatalf("expected Approved after re-approving the reverted content, got %s", state)
	}
	id, approved, err := ActiveGrant(dir, "SPEC-001", "fp-v1")
	if err != nil {
		t.Fatal(err)
	}
	if !approved {
		t.Fatalf("expected ActiveGrant to agree with Derive")
	}
	revs, err := Revisions(dir, "SPEC-001")
	if err != nil {
		t.Fatal(err)
	}
	if id != revs[len(revs)-1].GrantID {
		t.Fatalf("expected the newest grant to be L, got %q, latest revision is %q", id, revs[len(revs)-1].GrantID)
	}
}

func TestActiveGrant_LegacyDuplicates_LatestIsL_RevokingIt_Draft(t *testing.T) {
	dir := t.TempDir()
	fp := "the-fingerprint"
	if err := WriteGrant(dir, "SPEC-001", fp, "Jane <jane@example.com>", []byte("content")); err != nil {
		t.Fatal(err)
	}
	if err := WriteGrant(dir, "SPEC-001", fp, "Jane <jane@example.com>", []byte("content")); err != nil {
		t.Fatal(err)
	}
	state, err := Derive(dir, "SPEC-001", fp)
	if err != nil {
		t.Fatal(err)
	}
	if state != Approved {
		t.Fatalf("expected Approved with two unrevoked duplicate grants, got %s", state)
	}
	id, _, err := ActiveGrant(dir, "SPEC-001", fp)
	if err != nil {
		t.Fatal(err)
	}
	revs, err := Revisions(dir, "SPEC-001")
	if err != nil {
		t.Fatal(err)
	}
	if id != revs[len(revs)-1].GrantID {
		t.Fatalf("expected L to be the latest of the two duplicates, got %q, latest revision is %q", id, revs[len(revs)-1].GrantID)
	}

	if err := WriteRevocation(dir, "SPEC-001", id, "Jane <jane@example.com>"); err != nil {
		t.Fatal(err)
	}
	state, err = Derive(dir, "SPEC-001", fp)
	if err != nil {
		t.Fatal(err)
	}
	if state != Draft {
		t.Fatalf("expected Draft after revoking L, even though an older duplicate for the same content remains unrevoked, got %s", state)
	}
}

func TestActiveGrant_LatestRecordInvalid_IgnoredInFavorOfLatestValidGrant(t *testing.T) {
	dir := t.TempDir()
	if err := WriteGrant(dir, "SPEC-001", "fp-good", "Jane <jane@example.com>", []byte("good content")); err != nil {
		t.Fatal(err)
	}
	// A later, but invalid (unparseable timestamp), record must never participate in derivation.
	writeRawGrant(t, dir, "SPEC-001", "zzzzzz", Grant{SpecIdentity: "SPEC-001", Fingerprint: "fp-bad", Approver: "Jane <jane@example.com>", Timestamp: "not-a-timestamp"})

	state, err := Derive(dir, "SPEC-001", "fp-good")
	if err != nil {
		t.Fatal(err)
	}
	if state != Approved {
		t.Fatalf("expected Approved from the latest VALID grant, invalid record ignored, got %s", state)
	}
	id, approved, err := ActiveGrant(dir, "SPEC-001", "fp-good")
	if err != nil {
		t.Fatal(err)
	}
	if !approved || id == "zzzzzz" {
		t.Fatalf("expected L to be the valid grant, not the invalid record, got id=%q approved=%v", id, approved)
	}
}

func TestActiveGrant_LegacySameSecondGrants_DeterministicAcrossRepeatedCalls(t *testing.T) {
	dir := t.TempDir()
	ts := time.Now().UTC().Truncate(time.Second).Format(time.RFC3339)
	writeRawGrant(t, dir, "SPEC-001", "bbbbbb", Grant{SpecIdentity: "SPEC-001", Fingerprint: "fp-b", Approver: "Jane <jane@example.com>", Timestamp: ts})
	writeRawGrant(t, dir, "SPEC-001", "aaaaaa", Grant{SpecIdentity: "SPEC-001", Fingerprint: "fp-a", Approver: "Jane <jane@example.com>", Timestamp: ts})

	for i := 0; i < 5; i++ {
		id, approved, err := ActiveGrant(dir, "SPEC-001", "fp-b")
		if err != nil {
			t.Fatal(err)
		}
		if id != "bbbbbb" {
			t.Fatalf("call %d: expected L to deterministically be %q (grant id tie-break), got %q", i, "bbbbbb", id)
		}
		if !approved {
			t.Fatalf("call %d: expected fp-b (L's own content) to be Approved", i)
		}
	}
}

// --- ActiveGrantIDs (Step 12 Slice 4) ---

func TestActiveGrantIDs_NoEvidence_NotFound(t *testing.T) {
	dir := t.TempDir()
	ids, err := ActiveGrantIDs(dir, "SPEC-001", "abc")
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 0 {
		t.Fatalf("expected no active grants, got %v", ids)
	}
}

func TestActiveGrantIDs_MatchingGrant_ReturnsItsID(t *testing.T) {
	dir := t.TempDir()
	fp := "the-fingerprint"
	if err := WriteGrant(dir, "SPEC-001", fp, "Jane <jane@example.com>", []byte("content")); err != nil {
		t.Fatal(err)
	}
	want, err := readGrantIDs(dir, "SPEC-001")
	if err != nil {
		t.Fatal(err)
	}
	if len(want) != 1 {
		t.Fatalf("expected exactly one grant id, got %d", len(want))
	}
	ids, err := ActiveGrantIDs(dir, "SPEC-001", fp)
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 1 || ids[0] != want[0] {
		t.Fatalf("expected active grant id %q, got %v", want[0], ids)
	}
}

func TestActiveGrantIDs_NonMatchingFingerprint_NotFound(t *testing.T) {
	dir := t.TempDir()
	if err := WriteGrant(dir, "SPEC-001", "old-fingerprint", "Jane <jane@example.com>", []byte("content")); err != nil {
		t.Fatal(err)
	}
	ids, err := ActiveGrantIDs(dir, "SPEC-001", "new-fingerprint")
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 0 {
		t.Fatalf("expected no active grants for a stale fingerprint, got %v", ids)
	}
}

func TestActiveGrantIDs_RevokedGrant_NotFound(t *testing.T) {
	dir := t.TempDir()
	fp := "the-fingerprint"
	if err := WriteGrant(dir, "SPEC-001", fp, "Jane <jane@example.com>", []byte("content")); err != nil {
		t.Fatal(err)
	}
	ids, _ := readGrantIDs(dir, "SPEC-001")
	if err := WriteRevocation(dir, "SPEC-001", ids[0], "Jane <jane@example.com>"); err != nil {
		t.Fatal(err)
	}
	active, err := ActiveGrantIDs(dir, "SPEC-001", fp)
	if err != nil {
		t.Fatal(err)
	}
	if len(active) != 0 {
		t.Fatalf("expected no active grants once the only matching grant is revoked, got %v", active)
	}
}

func TestActiveGrantIDs_AgreesWithDerive(t *testing.T) {
	// With a single grant, it is trivially both L and the only ActiveGrantIDs candidate, so the
	// two must agree here. With more than one grant they can differ by design — ActiveGrantIDs is
	// a fingerprint match, not an approval decision; see ActiveGrant for that.
	dir := t.TempDir()
	fp := "the-fingerprint"
	if err := WriteGrant(dir, "SPEC-001", fp, "Jane <jane@example.com>", []byte("content")); err != nil {
		t.Fatal(err)
	}
	state, err := Derive(dir, "SPEC-001", fp)
	if err != nil {
		t.Fatal(err)
	}
	ids, err := ActiveGrantIDs(dir, "SPEC-001", fp)
	if err != nil {
		t.Fatal(err)
	}
	if (state == Approved) != (len(ids) > 0) {
		t.Fatalf("expected ActiveGrantIDs (%v) to agree with Derive's Approved (%v)", ids, state == Approved)
	}
}

// TestActiveGrantIDs_MultipleUnrevokedGrantsSameFingerprint_AllReturned is the direct regression
// guard for the duplicate-grant bug: two grants written for the same content (the old,
// unguarded-against double-approve) must both come back, sorted by id, so a caller revoking
// "the active grant" can actually clear all of them instead of leaving one behind.
func TestActiveGrantIDs_MultipleUnrevokedGrantsSameFingerprint_AllReturned(t *testing.T) {
	dir := t.TempDir()
	fp := "the-fingerprint"
	if err := WriteGrant(dir, "SPEC-001", fp, "Jane <jane@example.com>", []byte("content")); err != nil {
		t.Fatal(err)
	}
	if err := WriteGrant(dir, "SPEC-001", fp, "Jane <jane@example.com>", []byte("content")); err != nil {
		t.Fatal(err)
	}
	want, err := readGrantIDs(dir, "SPEC-001")
	if err != nil {
		t.Fatal(err)
	}
	if len(want) != 2 {
		t.Fatalf("expected exactly two grant files, got %d", len(want))
	}
	sort.Strings(want)

	got, err := ActiveGrantIDs(dir, "SPEC-001", fp)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected both grant ids sorted %v, got %v", want, got)
	}
}

// TestActiveGrantIDs_MixOfRevokedAndOtherFingerprint_OnlyGenuineMatchesReturned proves a revoked
// duplicate and a grant for different content are both excluded, alongside a genuine match.
func TestActiveGrantIDs_MixOfRevokedAndOtherFingerprint_OnlyGenuineMatchesReturned(t *testing.T) {
	dir := t.TempDir()
	fp := "the-fingerprint"
	if err := WriteGrant(dir, "SPEC-001", fp, "Jane <jane@example.com>", []byte("content")); err != nil {
		t.Fatal(err)
	}
	if err := WriteGrant(dir, "SPEC-001", fp, "Jane <jane@example.com>", []byte("content")); err != nil {
		t.Fatal(err)
	}
	if err := WriteGrant(dir, "SPEC-001", "other-fingerprint", "Jane <jane@example.com>", []byte("other content")); err != nil {
		t.Fatal(err)
	}
	all, err := readGrantIDs(dir, "SPEC-001")
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 3 {
		t.Fatalf("expected three grant files, got %d", len(all))
	}
	// Revoke one of the two matching-fingerprint grants, leaving one still active.
	before, err := ActiveGrantIDs(dir, "SPEC-001", fp)
	if err != nil {
		t.Fatal(err)
	}
	if len(before) != 2 {
		t.Fatalf("expected two active grants before revocation, got %v", before)
	}
	if err := WriteRevocation(dir, "SPEC-001", before[0], "Jane <jane@example.com>"); err != nil {
		t.Fatal(err)
	}

	got, err := ActiveGrantIDs(dir, "SPEC-001", fp)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0] != before[1] {
		t.Fatalf("expected only the still-unrevoked matching grant %q, got %v", before[1], got)
	}
}

// --- ValidateEvidence (Step 12 Slice 5) ---

// specsWithFiles creates a specifications directory containing an empty file for each identity —
// so a test whose focus is some other rule does not incidentally also trigger the "no
// Specification file exists for this identity" warning.
func specsWithFiles(t *testing.T, ids ...string) string {
	t.Helper()
	dir := t.TempDir()
	for _, id := range ids {
		if err := os.WriteFile(filepath.Join(dir, id+".md"), []byte("# "+id), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func TestValidateEvidence_NoApprovalsDir_NoIssues(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "does-not-exist")
	issues, err := ValidateEvidence(dir, specsWithFiles(t))
	if err != nil {
		t.Fatal(err)
	}
	if len(issues) != 0 {
		t.Fatalf("expected no issues for an absent approvals directory, got: %+v", issues)
	}
}

func TestValidateEvidence_ValidGrantAndRevocation_NoIssues(t *testing.T) {
	dir := t.TempDir()
	if err := WriteGrant(dir, "SPEC-001", "fp", "Jane <jane@example.com>", []byte("content")); err != nil {
		t.Fatal(err)
	}
	ids, err := readGrantIDs(dir, "SPEC-001")
	if err != nil {
		t.Fatal(err)
	}
	if err := WriteRevocation(dir, "SPEC-001", ids[0], "Jane <jane@example.com>"); err != nil {
		t.Fatal(err)
	}
	issues, err := ValidateEvidence(dir, specsWithFiles(t, "SPEC-001"))
	if err != nil {
		t.Fatal(err)
	}
	if len(issues) != 0 {
		t.Fatalf("expected no issues for valid evidence, got: %+v", issues)
	}
}

func TestValidateEvidence_MalformedJSON_Surfaced(t *testing.T) {
	dir := t.TempDir()
	if err := writeFileHelper(specDir(dir, "SPEC-001"), "bad.grant.json", "not json"); err != nil {
		t.Fatal(err)
	}
	issues, err := ValidateEvidence(dir, specsWithFiles(t, "SPEC-001"))
	if err != nil {
		t.Fatal(err)
	}
	if len(issues) != 1 || issues[0].Filename != "bad.grant.json" || issues[0].Level != LevelError {
		t.Fatalf("expected exactly one error issue for the malformed file, got: %+v", issues)
	}
}

func TestValidateEvidence_MissingRequiredField_Surfaced(t *testing.T) {
	dir := t.TempDir()
	// A grant with an empty approver — structurally parseable JSON, but missing required
	// attribution per cli/APPROVAL_RUNTIME.md's Grant Record. Timestamp is a real one, since this
	// test isolates the missing-field rule from the separate timestamp-validity rule.
	if err := writeFileHelper(specDir(dir, "SPEC-001"), "x.grant.json",
		`{"spec_identity":"SPEC-001","fingerprint":"fp","approver":"","timestamp":"2024-01-01T00:00:00Z"}`); err != nil {
		t.Fatal(err)
	}
	issues, err := ValidateEvidence(dir, specsWithFiles(t, "SPEC-001"))
	if err != nil {
		t.Fatal(err)
	}
	if len(issues) != 1 || issues[0].Level != LevelError {
		t.Fatalf("expected exactly one error issue for the missing field, got: %+v", issues)
	}
}

func TestValidateEvidence_UnrecognizedFilename_Surfaced(t *testing.T) {
	dir := t.TempDir()
	if err := writeFileHelper(specDir(dir, "SPEC-001"), "stray.txt", "hello"); err != nil {
		t.Fatal(err)
	}
	issues, err := ValidateEvidence(dir, specsWithFiles(t, "SPEC-001"))
	if err != nil {
		t.Fatal(err)
	}
	if len(issues) != 1 || issues[0].Filename != "stray.txt" || issues[0].Level != LevelError {
		t.Fatalf("expected exactly one error issue for the stray file, got: %+v", issues)
	}
}

func TestValidateEvidence_NeverModifiesFiles(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(specDir(dir, "SPEC-001"), "bad.grant.json")
	if err := writeFileHelper(specDir(dir, "SPEC-001"), "bad.grant.json", "not json"); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ValidateEvidence(dir, specsWithFiles(t)); err != nil {
		t.Fatal(err)
	}
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != string(after) {
		t.Fatalf("expected ValidateEvidence to never modify a malformed file, only report it")
	}
}

// --- One test per validity rule: what ValidateEvidence reports, and whether Derive agrees ---

// findIssue returns the first issue matching filename, if any.
func findIssue(issues []EvidenceIssue, filename string) (EvidenceIssue, bool) {
	for _, i := range issues {
		if i.Filename == filename {
			return i, true
		}
	}
	return EvidenceIssue{}, false
}

func TestRule_TimestampNotRFC3339_ErrorAndExcludedFromDerivation(t *testing.T) {
	dir := t.TempDir()
	writeRawGrant(t, dir, "SPEC-001", "badts", Grant{
		SpecIdentity: "SPEC-001", Fingerprint: "fp", Approver: "Jane <jane@example.com>", Timestamp: "x",
	})

	issues, err := ValidateEvidence(dir, specsWithFiles(t, "SPEC-001"))
	if err != nil {
		t.Fatal(err)
	}
	issue, found := findIssue(issues, "badts.grant.json")
	if !found || issue.Level != LevelError {
		t.Fatalf("expected an error issue for the invalid timestamp, got: %+v", issues)
	}

	state, err := Derive(dir, "SPEC-001", "fp")
	if err != nil {
		t.Fatal(err)
	}
	if state != Draft {
		t.Fatalf("expected the Specification to derive Draft (its only matching grant is invalid), got %v", state)
	}
}

func TestRule_SpecIdentityMismatch_ErrorAndExcludedFromDerivation(t *testing.T) {
	dir := t.TempDir()
	// Stored under SPEC-001/ but claims to belong to SPEC-002.
	writeRawGrant(t, dir, "SPEC-001", "mismatched", Grant{
		SpecIdentity: "SPEC-002", Fingerprint: "fp", Approver: "Jane <jane@example.com>", Timestamp: "2024-01-01T00:00:00Z",
	})

	issues, err := ValidateEvidence(dir, specsWithFiles(t, "SPEC-001", "SPEC-002"))
	if err != nil {
		t.Fatal(err)
	}
	issue, found := findIssue(issues, "mismatched.grant.json")
	if !found || issue.Level != LevelError || issue.SpecID != "SPEC-001" {
		t.Fatalf("expected an error issue under SPEC-001 for the identity mismatch, got: %+v", issues)
	}

	state, err := Derive(dir, "SPEC-001", "fp")
	if err != nil {
		t.Fatal(err)
	}
	if state != Draft {
		t.Fatalf("expected SPEC-001 to derive Draft (the only grant in its directory does not belong to it), got %v", state)
	}
}

func TestRule_RevocationOfUnknownGrant_WarningAndDerivationUnaffected(t *testing.T) {
	dir := t.TempDir()
	if err := WriteGrant(dir, "SPEC-001", "fp", "Jane <jane@example.com>", []byte("content")); err != nil {
		t.Fatal(err)
	}
	if err := WriteRevocation(dir, "SPEC-001", "ghost-grant-id", "Jane <jane@example.com>"); err != nil {
		t.Fatal(err)
	}

	issues, err := ValidateEvidence(dir, specsWithFiles(t, "SPEC-001"))
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, i := range issues {
		if i.Level == LevelWarning && strings.Contains(i.Problem, "ghost-grant-id") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected a warning about the revocation's unknown grant id, got: %+v", issues)
	}

	// The real grant is untouched by the unrelated, unresolvable revocation.
	state, err := Derive(dir, "SPEC-001", "fp")
	if err != nil {
		t.Fatal(err)
	}
	if state != Approved {
		t.Fatalf("expected the real grant to still derive Approved, got %v", state)
	}
}

func TestRule_OrphanContentSnapshot_Warning(t *testing.T) {
	dir := t.TempDir()
	if err := writeFileHelper(specDir(dir, "SPEC-001"), "orphan.content.md", "orphaned text"); err != nil {
		t.Fatal(err)
	}

	issues, err := ValidateEvidence(dir, specsWithFiles(t, "SPEC-001"))
	if err != nil {
		t.Fatal(err)
	}
	issue, found := findIssue(issues, "orphan.content.md")
	if !found || issue.Level != LevelWarning {
		t.Fatalf("expected a warning for the orphaned content snapshot, got: %+v", issues)
	}
}

func TestRule_NoSpecificationFile_Warning(t *testing.T) {
	dir := t.TempDir()
	if err := WriteGrant(dir, "SPEC-001", "fp", "Jane <jane@example.com>", []byte("content")); err != nil {
		t.Fatal(err)
	}

	// specsWithFiles(t) with no ids: an empty specifications directory, so SPEC-001 has no file.
	issues, err := ValidateEvidence(dir, specsWithFiles(t))
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, i := range issues {
		if i.SpecID == "SPEC-001" && i.Filename == "" && i.Level == LevelWarning {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected a directory-level warning for the missing Specification file, got: %+v", issues)
	}
}

func TestRule_DuplicateRevocations_WarningAndGrantStillRevoked(t *testing.T) {
	dir := t.TempDir()
	if err := WriteGrant(dir, "SPEC-001", "fp", "Jane <jane@example.com>", []byte("content")); err != nil {
		t.Fatal(err)
	}
	ids, err := readGrantIDs(dir, "SPEC-001")
	if err != nil {
		t.Fatal(err)
	}
	if err := WriteRevocation(dir, "SPEC-001", ids[0], "Jane <jane@example.com>"); err != nil {
		t.Fatal(err)
	}
	if err := WriteRevocation(dir, "SPEC-001", ids[0], "Bob <bob@example.com>"); err != nil {
		t.Fatal(err)
	}

	issues, err := ValidateEvidence(dir, specsWithFiles(t, "SPEC-001"))
	if err != nil {
		t.Fatal(err)
	}
	warnings := 0
	for _, i := range issues {
		if i.Level == LevelWarning && strings.Contains(i.Problem, "2 revocations") {
			warnings++
		}
	}
	if warnings != 2 {
		t.Fatalf("expected both duplicate revocation files flagged, got %d warnings in: %+v", warnings, issues)
	}

	state, err := Derive(dir, "SPEC-001", "fp")
	if err != nil {
		t.Fatal(err)
	}
	if state != Draft {
		t.Fatalf("expected the grant to still be revoked (Draft) despite the duplicate revocations, got %v", state)
	}
}

func TestRule_LegacyGrantNoContentSnapshot_NoProblems(t *testing.T) {
	dir := t.TempDir()
	writeRawGrant(t, dir, "SPEC-001", "legacy1", Grant{
		SpecIdentity: "SPEC-001", Fingerprint: "fp", Approver: "Jane <jane@example.com>",
		Timestamp: time.Now().UTC().Truncate(time.Second).Format(time.RFC3339), // second precision, legacy
	})
	// Deliberately no legacy1.content.md — records written before content snapshots existed.

	issues, err := ValidateEvidence(dir, specsWithFiles(t, "SPEC-001"))
	if err != nil {
		t.Fatal(err)
	}
	if len(issues) != 0 {
		t.Fatalf("expected a grant missing only its content snapshot to raise no problems, got: %+v", issues)
	}

	state, err := Derive(dir, "SPEC-001", "fp")
	if err != nil {
		t.Fatal(err)
	}
	if state != Approved {
		t.Fatalf("expected the legacy grant to still be valid for derivation, got %v", state)
	}
}

func TestRule_FullyValidProject_NoProblems(t *testing.T) {
	root := t.TempDir()
	if err := WriteGrant(root, "SPEC-001", "fp1", "Jane <jane@example.com>", []byte("content one")); err != nil {
		t.Fatal(err)
	}
	if err := WriteGrant(root, "SPEC-002", "fp2", "Jane <jane@example.com>", []byte("content two")); err != nil {
		t.Fatal(err)
	}
	ids, err := readGrantIDs(root, "SPEC-002")
	if err != nil {
		t.Fatal(err)
	}
	if err := WriteRevocation(root, "SPEC-002", ids[0], "Jane <jane@example.com>"); err != nil {
		t.Fatal(err)
	}

	issues, err := ValidateEvidence(root, specsWithFiles(t, "SPEC-001", "SPEC-002"))
	if err != nil {
		t.Fatal(err)
	}
	if len(issues) != 0 {
		t.Fatalf("expected a fully valid project to raise no problems, got: %+v", issues)
	}
}

func TestDerive_MalformedRecordIsIgnored(t *testing.T) {
	dir := t.TempDir()
	specDirPath := specDir(dir, "SPEC-001")
	if err := writeFileHelper(specDirPath, "bad.grant.json", "not json"); err != nil {
		t.Fatal(err)
	}
	state, err := Derive(dir, "SPEC-001", "anything")
	if err != nil {
		t.Fatal(err)
	}
	if state != Draft {
		t.Fatalf("expected malformed evidence to be ignored, deriving Draft, got %s", state)
	}
}

// --- Revisions ---

func TestRevisions_EmptyWhenNeverApproved(t *testing.T) {
	dir := t.TempDir()
	revs, err := Revisions(dir, "SPEC-001")
	if err != nil {
		t.Fatal(err)
	}
	if len(revs) != 0 {
		t.Fatalf("expected no revisions, got %v", revs)
	}
}

func TestRevisions_OneGrant_ContentPreserved(t *testing.T) {
	dir := t.TempDir()
	if err := WriteGrant(dir, "SPEC-001", "fp1", "Jane <jane@example.com>", []byte("first content")); err != nil {
		t.Fatal(err)
	}
	revs, err := Revisions(dir, "SPEC-001")
	if err != nil {
		t.Fatal(err)
	}
	if len(revs) != 1 {
		t.Fatalf("expected 1 revision, got %d", len(revs))
	}
	r := revs[0]
	if r.Fingerprint != "fp1" || r.Approver != "Jane <jane@example.com>" {
		t.Fatalf("unexpected revision: %+v", r)
	}
	if !r.ContentKnown || r.Content != "first content" {
		t.Fatalf("expected content snapshot preserved, got %+v", r)
	}
	if r.Revoked {
		t.Fatalf("expected an unrevoked grant to report Revoked=false")
	}
}

func TestRevisions_SupersededApproval_PreviousRevisionStillInspectable(t *testing.T) {
	dir := t.TempDir()
	if err := WriteGrant(dir, "SPEC-001", "fp1", "Jane <jane@example.com>", []byte("original approved text")); err != nil {
		t.Fatal(err)
	}
	// A later edit changes content/fingerprint and produces a second grant — the fingerprint
	// derivation rule already makes the Specification Draft again until re-approved; Revisions
	// must still show the first, now-superseded revision's content unchanged.
	if err := WriteGrant(dir, "SPEC-001", "fp2", "Jane <jane@example.com>", []byte("revised approved text")); err != nil {
		t.Fatal(err)
	}
	revs, err := Revisions(dir, "SPEC-001")
	if err != nil {
		t.Fatal(err)
	}
	if len(revs) != 2 {
		t.Fatalf("expected 2 revisions, got %d", len(revs))
	}
	if revs[0].Content != "original approved text" {
		t.Fatalf("expected the first revision's original content preserved, got %q", revs[0].Content)
	}
	if revs[1].Content != "revised approved text" {
		t.Fatalf("expected the second revision's content preserved, got %q", revs[1].Content)
	}
}

// --- Timestamp precision, monotonicity, and sort determinism ---

func TestWriteGrant_BackToBack_StrictlyIncreasingTimestamps(t *testing.T) {
	dir := t.TempDir()
	if err := WriteGrant(dir, "SPEC-001", "fp1", "Jane <jane@example.com>", []byte("v1")); err != nil {
		t.Fatal(err)
	}
	if err := WriteGrant(dir, "SPEC-001", "fp2", "Jane <jane@example.com>", []byte("v2")); err != nil {
		t.Fatal(err)
	}
	revs, err := Revisions(dir, "SPEC-001")
	if err != nil {
		t.Fatal(err)
	}
	if len(revs) != 2 {
		t.Fatalf("expected 2 revisions, got %d", len(revs))
	}
	t1, err1 := time.Parse(time.RFC3339Nano, revs[0].ApprovedAt)
	t2, err2 := time.Parse(time.RFC3339Nano, revs[1].ApprovedAt)
	if err1 != nil || err2 != nil {
		t.Fatalf("expected both timestamps to parse, got %v / %v", err1, err2)
	}
	if !t2.After(t1) {
		t.Fatalf("expected the second grant's timestamp (%s) strictly after the first's (%s)", t2, t1)
	}
}

// TestWriteGrant_LatestExistingTimestampInFuture_UsesLatestPlusOneNanosecond covers the coarse-
// clock case: an existing grant's timestamp is already later than "now" (simulated here by
// hand-writing one, since a real clock won't naturally do this in a fast test run).
func TestWriteGrant_LatestExistingTimestampInFuture_UsesLatestPlusOneNanosecond(t *testing.T) {
	dir := t.TempDir()
	future := time.Now().UTC().Add(24 * time.Hour)
	writeRawGrant(t, dir, "SPEC-001", "future1", Grant{
		SpecIdentity: "SPEC-001", Fingerprint: "fp-old", Approver: "Jane <jane@example.com>",
		Timestamp: future.Format(time.RFC3339Nano),
	})

	if err := WriteGrant(dir, "SPEC-001", "fp-new", "Jane <jane@example.com>", []byte("new content")); err != nil {
		t.Fatal(err)
	}

	grants, _, err := scanEvidence(dir, "SPEC-001")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := grants["future1"]; !ok {
		t.Fatalf("expected the hand-written grant to still be readable")
	}
	var newTS string
	for id, gg := range grants {
		if id != "future1" {
			newTS = gg.Timestamp
		}
	}
	got, err := time.Parse(time.RFC3339Nano, newTS)
	if err != nil {
		t.Fatalf("expected the new grant's timestamp to parse: %v", err)
	}
	want := future.Add(time.Nanosecond)
	if !got.Equal(want) {
		t.Fatalf("expected latest+1ns (%s), got %s", want, got)
	}
}

// TestRevisions_LegacySameSecondTimestamps_OrderedByGrantIDDeterministically proves two legacy
// grants with identical (second-precision) timestamps sort by grant id, the same way every time.
func TestRevisions_LegacySameSecondTimestamps_OrderedByGrantIDDeterministically(t *testing.T) {
	dir := t.TempDir()
	ts := time.Now().UTC().Truncate(time.Second).Format(time.RFC3339)
	writeRawGrant(t, dir, "SPEC-001", "bbbbbb", Grant{SpecIdentity: "SPEC-001", Fingerprint: "fp-b", Approver: "Jane <jane@example.com>", Timestamp: ts})
	writeRawGrant(t, dir, "SPEC-001", "aaaaaa", Grant{SpecIdentity: "SPEC-001", Fingerprint: "fp-a", Approver: "Jane <jane@example.com>", Timestamp: ts})

	for i := 0; i < 5; i++ {
		revs, err := Revisions(dir, "SPEC-001")
		if err != nil {
			t.Fatal(err)
		}
		if len(revs) != 2 || revs[0].GrantID != "aaaaaa" || revs[1].GrantID != "bbbbbb" {
			t.Fatalf("call %d: expected [aaaaaa, bbbbbb] for equal timestamps, got %+v", i, revs)
		}
	}
}

// TestRevisions_UnparseableTimestamp_ExcludedEntirely proves a grant with an unparseable
// timestamp is invalid (checkGrant), not merely sorted specially: excluded from Revisions and
// from Derive's own Approved computation, and reported as an error by ValidateEvidence — the
// behavior changed from an earlier phase (before checkGrant validated the timestamp field at
// all), where such a grant was treated as valid and sorted first via the zero-time fallback.
func TestRevisions_UnparseableTimestamp_ExcludedEntirely(t *testing.T) {
	dir := t.TempDir()
	writeRawGrant(t, dir, "SPEC-001", "badts01", Grant{SpecIdentity: "SPEC-001", Fingerprint: "fp-bad", Approver: "Jane <jane@example.com>", Timestamp: "not-a-timestamp"})
	if err := WriteGrant(dir, "SPEC-001", "fp-good", "Jane <jane@example.com>", []byte("good content")); err != nil {
		t.Fatal(err)
	}

	revs, err := Revisions(dir, "SPEC-001")
	if err != nil {
		t.Fatal(err)
	}
	if len(revs) != 1 {
		t.Fatalf("expected only the valid grant as a revision, got %+v", revs)
	}
	for _, r := range revs {
		if r.GrantID == "badts01" {
			t.Fatalf("expected the unparseable-timestamp grant excluded, got it present: %+v", r)
		}
	}

	state, err := Derive(dir, "SPEC-001", "fp-bad")
	if err != nil {
		t.Fatal(err)
	}
	if state != Draft {
		t.Fatalf("expected the unparseable-timestamp grant to never make its Specification Approved, got %v", state)
	}

	issues, err := ValidateEvidence(dir, specsWithFiles(t, "SPEC-001"))
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, i := range issues {
		if i.Filename == "badts01.grant.json" && i.Level == LevelError {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected ValidateEvidence to report the unparseable timestamp as an error, got: %+v", issues)
	}
}

// TestRevisions_MixedLegacyAndNanosecondTimestamps_SortChronologically proves a legacy
// second-precision record and a nanosecond-precision record still sort correctly against each
// other by actual time, not by string comparison.
func TestRevisions_MixedLegacyAndNanosecondTimestamps_SortChronologically(t *testing.T) {
	dir := t.TempDir()
	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	writeRawGrant(t, dir, "SPEC-001", "legacy1", Grant{SpecIdentity: "SPEC-001", Fingerprint: "fp-legacy", Approver: "Jane <jane@example.com>", Timestamp: base.Format(time.RFC3339)})
	writeRawGrant(t, dir, "SPEC-001", "newer1", Grant{SpecIdentity: "SPEC-001", Fingerprint: "fp-newer", Approver: "Jane <jane@example.com>", Timestamp: base.Add(time.Hour).Format(time.RFC3339Nano)})

	revs, err := Revisions(dir, "SPEC-001")
	if err != nil {
		t.Fatal(err)
	}
	if len(revs) != 2 || revs[0].GrantID != "legacy1" || revs[1].GrantID != "newer1" {
		t.Fatalf("expected chronological order [legacy1, newer1], got %+v", revs)
	}
}

// writeRawGrant hand-writes a grant file bypassing WriteGrant, so a test can construct evidence
// shapes (a specific id, a specific raw timestamp) WriteGrant itself would never produce.
func writeRawGrant(t *testing.T, approvalsDir, specID, id string, g Grant) {
	t.Helper()
	dir := specDir(approvalsDir, specID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(g)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, id+".grant.json"), data, 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestRevisions_RevokedGrant_MarkedRevokedWithAttribution(t *testing.T) {
	dir := t.TempDir()
	if err := WriteGrant(dir, "SPEC-001", "fp1", "Jane <jane@example.com>", []byte("content")); err != nil {
		t.Fatal(err)
	}
	revs, err := Revisions(dir, "SPEC-001")
	if err != nil {
		t.Fatal(err)
	}
	if err := WriteRevocation(dir, "SPEC-001", revs[0].GrantID, "Bob <bob@example.com>"); err != nil {
		t.Fatal(err)
	}
	revs, err = Revisions(dir, "SPEC-001")
	if err != nil {
		t.Fatal(err)
	}
	if !revs[0].Revoked || revs[0].RevokedBy != "Bob <bob@example.com>" {
		t.Fatalf("expected revocation attribution to be reflected, got %+v", revs[0])
	}
}

func TestRevisions_ContentUnknown_WhenSnapshotMissing(t *testing.T) {
	// Simulates approval evidence written before the content-snapshot feature existed — the
	// grant.json exists, but no matching .content.md does. Revisions must represent this
	// honestly (ContentKnown: false) rather than fabricating or guessing content.
	dir := t.TempDir()
	specDirPath := specDir(dir, "SPEC-001")
	if err := writeFileHelper(specDirPath, "abc123.grant.json", `{"spec_identity":"SPEC-001","fingerprint":"fp1","approver":"Jane","timestamp":"2020-01-01T00:00:00Z"}`); err != nil {
		t.Fatal(err)
	}
	revs, err := Revisions(dir, "SPEC-001")
	if err != nil {
		t.Fatal(err)
	}
	if len(revs) != 1 {
		t.Fatalf("expected 1 revision, got %d", len(revs))
	}
	if revs[0].ContentKnown {
		t.Fatalf("expected ContentKnown=false when no snapshot exists, got %+v", revs[0])
	}
}

func TestValidateEvidence_ContentSnapshotIsRecognized(t *testing.T) {
	dir := t.TempDir()
	if err := WriteGrant(dir, "SPEC-001", "fp1", "Jane <jane@example.com>", []byte("content")); err != nil {
		t.Fatal(err)
	}
	issues, err := ValidateEvidence(dir, specsWithFiles(t, "SPEC-001"))
	if err != nil {
		t.Fatal(err)
	}
	if len(issues) != 0 {
		t.Fatalf("expected the .content.md snapshot to be a recognized evidence file, got issues: %+v", issues)
	}
}
