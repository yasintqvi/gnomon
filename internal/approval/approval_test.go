package approval

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
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

// --- ActiveGrantID (Step 12 Slice 4) ---

func TestActiveGrantID_NoEvidence_NotFound(t *testing.T) {
	dir := t.TempDir()
	id, ok, err := ActiveGrantID(dir, "SPEC-001", "abc")
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatalf("expected no active grant, got id %q", id)
	}
}

func TestActiveGrantID_MatchingGrant_ReturnsItsID(t *testing.T) {
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
	id, ok, err := ActiveGrantID(dir, "SPEC-001", fp)
	if err != nil {
		t.Fatal(err)
	}
	if !ok || id != want[0] {
		t.Fatalf("expected active grant id %q, got %q (ok=%v)", want[0], id, ok)
	}
}

func TestActiveGrantID_NonMatchingFingerprint_NotFound(t *testing.T) {
	dir := t.TempDir()
	if err := WriteGrant(dir, "SPEC-001", "old-fingerprint", "Jane <jane@example.com>", []byte("content")); err != nil {
		t.Fatal(err)
	}
	_, ok, err := ActiveGrantID(dir, "SPEC-001", "new-fingerprint")
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatalf("expected no active grant for a stale fingerprint")
	}
}

func TestActiveGrantID_RevokedGrant_NotFound(t *testing.T) {
	dir := t.TempDir()
	fp := "the-fingerprint"
	if err := WriteGrant(dir, "SPEC-001", fp, "Jane <jane@example.com>", []byte("content")); err != nil {
		t.Fatal(err)
	}
	ids, _ := readGrantIDs(dir, "SPEC-001")
	if err := WriteRevocation(dir, "SPEC-001", ids[0], "Jane <jane@example.com>"); err != nil {
		t.Fatal(err)
	}
	_, ok, err := ActiveGrantID(dir, "SPEC-001", fp)
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatalf("expected no active grant once the only matching grant is revoked")
	}
}

func TestActiveGrantID_AgreesWithDerive(t *testing.T) {
	// The two functions must always agree: an active grant id exists if and only if Derive
	// reports Approved — proving they share one derivation, never two that could drift apart.
	dir := t.TempDir()
	fp := "the-fingerprint"
	if err := WriteGrant(dir, "SPEC-001", fp, "Jane <jane@example.com>", []byte("content")); err != nil {
		t.Fatal(err)
	}
	state, err := Derive(dir, "SPEC-001", fp)
	if err != nil {
		t.Fatal(err)
	}
	_, ok, err := ActiveGrantID(dir, "SPEC-001", fp)
	if err != nil {
		t.Fatal(err)
	}
	if (state == Approved) != ok {
		t.Fatalf("expected ActiveGrantID's ok (%v) to agree with Derive's Approved (%v)", ok, state == Approved)
	}
}

// --- ValidateEvidence (Step 12 Slice 5) ---

func TestValidateEvidence_NoApprovalsDir_NoIssues(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "does-not-exist")
	issues, err := ValidateEvidence(dir)
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
	issues, err := ValidateEvidence(dir)
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
	issues, err := ValidateEvidence(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(issues) != 1 || issues[0].Filename != "bad.grant.json" {
		t.Fatalf("expected exactly one issue for the malformed file, got: %+v", issues)
	}
}

func TestValidateEvidence_MissingRequiredField_Surfaced(t *testing.T) {
	dir := t.TempDir()
	// A grant with an empty approver — structurally parseable JSON, but missing required
	// attribution per cli/APPROVAL_RUNTIME.md's Grant Record.
	if err := writeFileHelper(specDir(dir, "SPEC-001"), "x.grant.json",
		`{"spec_identity":"SPEC-001","fingerprint":"fp","approver":"","timestamp":"t"}`); err != nil {
		t.Fatal(err)
	}
	issues, err := ValidateEvidence(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(issues) != 1 {
		t.Fatalf("expected exactly one issue for the missing field, got: %+v", issues)
	}
}

func TestValidateEvidence_UnrecognizedFilename_Surfaced(t *testing.T) {
	dir := t.TempDir()
	if err := writeFileHelper(specDir(dir, "SPEC-001"), "stray.txt", "hello"); err != nil {
		t.Fatal(err)
	}
	issues, err := ValidateEvidence(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(issues) != 1 || issues[0].Filename != "stray.txt" {
		t.Fatalf("expected exactly one issue for the stray file, got: %+v", issues)
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
	if _, err := ValidateEvidence(dir); err != nil {
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
	issues, err := ValidateEvidence(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(issues) != 0 {
		t.Fatalf("expected the .content.md snapshot to be a recognized evidence file, got issues: %+v", issues)
	}
}
