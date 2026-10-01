package orchestrate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gnomon/internal/approval"
	"gnomon/internal/facts"
	"gnomon/internal/present"
	"gnomon/internal/project"
)

// --- "The latest approval decision counts" — orchestrate-level behavior ---

func TestApprove_ContentMatchesOlderNonLatestGrant_WritesNewGrant_Approved(t *testing.T) {
	root := t.TempDir()
	configureGitIdentity(t, root)
	if _, err := Init(root); err != nil {
		t.Fatal(err)
	}
	if _, err := specCreateDefined(root, "Password Reset"); err != nil {
		t.Fatal(err)
	}
	l, err := project.Locate(root)
	if err != nil {
		t.Fatal(err)
	}
	specPath := filepath.Join(l.SpecificationsDir(), "SPEC-001-password-reset.md")

	v1, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Approve(root, "SPEC-001", nil, nil); err != nil {
		t.Fatalf("approve v1: %v", err)
	}

	v2 := append(append([]byte{}, v1...), []byte("\nv2 sentence.\n")...)
	if err := os.WriteFile(specPath, v2, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Approve(root, "SPEC-001", nil, nil); err != nil {
		t.Fatalf("approve v2: %v", err)
	}

	// Content reverts to exactly v1 — an older, non-latest grant. Approving it must write a new
	// grant, never a no-op, even though a grant for this exact content already exists.
	if err := os.WriteFile(specPath, v1, 0o644); err != nil {
		t.Fatal(err)
	}
	report, err := Approve(root, "SPEC-001", nil, nil)
	if err != nil {
		t.Fatalf("re-approve v1: %v", err)
	}
	if report.Outcome != present.Success {
		t.Fatalf("expected Success, got %v", report.Outcome)
	}
	if strings.Contains(report.Summary, "already Approved") {
		t.Fatalf("expected a real new approval, not the already-Approved no-op, got %q", report.Summary)
	}

	revs, err := approval.Revisions(l.ApprovalsDir(), "SPEC-001")
	if err != nil {
		t.Fatal(err)
	}
	if len(revs) != 3 {
		t.Fatalf("expected three grants (v1, v2, v1-again), got %d: %+v", len(revs), revs)
	}
	id, approved, err := approval.ActiveGrant(l.ApprovalsDir(), "SPEC-001", approval.Fingerprint(v1))
	if err != nil {
		t.Fatal(err)
	}
	if !approved || id != revs[len(revs)-1].GrantID {
		t.Fatalf("expected the newest grant to be L and Approved, got id=%q approved=%v latest=%q", id, approved, revs[len(revs)-1].GrantID)
	}
}

func TestApprove_ContentMatchesLatestGrant_NoOp_NothingWritten(t *testing.T) {
	root := t.TempDir()
	configureGitIdentity(t, root)
	if _, err := Init(root); err != nil {
		t.Fatal(err)
	}
	if _, err := specCreateDefined(root, "Password Reset"); err != nil {
		t.Fatal(err)
	}
	l, err := project.Locate(root)
	if err != nil {
		t.Fatal(err)
	}
	specPath := filepath.Join(l.SpecificationsDir(), "SPEC-001-password-reset.md")

	v1, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Approve(root, "SPEC-001", nil, nil); err != nil {
		t.Fatal(err)
	}
	v2 := append(append([]byte{}, v1...), []byte("\nv2 sentence.\n")...)
	if err := os.WriteFile(specPath, v2, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Approve(root, "SPEC-001", nil, nil); err != nil {
		t.Fatal(err)
	}

	revsBefore, err := approval.Revisions(l.ApprovalsDir(), "SPEC-001")
	if err != nil {
		t.Fatal(err)
	}

	// Content already equals L (v2) — approving again is the no-op, nothing new written.
	report, err := Approve(root, "SPEC-001", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(report.Summary, "already Approved") {
		t.Fatalf("expected the already-Approved no-op message, got %q", report.Summary)
	}

	revsAfter, err := approval.Revisions(l.ApprovalsDir(), "SPEC-001")
	if err != nil {
		t.Fatal(err)
	}
	if len(revsAfter) != len(revsBefore) {
		t.Fatalf("expected no new grant written, had %d revisions, now %d", len(revsBefore), len(revsAfter))
	}
}

func TestRevoke_OlderMatchingGrantNeverBecomesApprovedAfter(t *testing.T) {
	root := t.TempDir()
	configureGitIdentity(t, root)
	if _, err := Init(root); err != nil {
		t.Fatal(err)
	}
	if _, err := specCreateDefined(root, "Password Reset"); err != nil {
		t.Fatal(err)
	}
	l, err := project.Locate(root)
	if err != nil {
		t.Fatal(err)
	}
	specPath := filepath.Join(l.SpecificationsDir(), "SPEC-001-password-reset.md")

	v1, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Approve(root, "SPEC-001", nil, nil); err != nil {
		t.Fatal(err)
	}
	v2 := append(append([]byte{}, v1...), []byte("\nv2 sentence.\n")...)
	if err := os.WriteFile(specPath, v2, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Approve(root, "SPEC-001", nil, nil); err != nil {
		t.Fatal(err)
	}

	if _, err := Revoke(root, "SPEC-001", nil); err != nil {
		t.Fatalf("revoke: %v", err)
	}

	// Restore content to v1 — an older grant that still exists, unrevoked. It must never make the
	// Specification Approved again; revoking L is final until a fresh approval.
	if err := os.WriteFile(specPath, v1, 0o644); err != nil {
		t.Fatal(err)
	}
	state, err := facts.Lifecycle(l, "SPEC-001")
	if err != nil {
		t.Fatal(err)
	}
	if state != approval.Draft {
		t.Fatalf("expected Draft: an older grant must never become active again, got %v", state)
	}
}

func TestSpecDetailFor_RevisionStatuses_ActiveRevokedSuperseded(t *testing.T) {
	root := t.TempDir()
	configureGitIdentity(t, root)
	if _, err := Init(root); err != nil {
		t.Fatal(err)
	}
	if _, err := specCreateDefined(root, "Password Reset"); err != nil {
		t.Fatal(err)
	}
	l, err := project.Locate(root)
	if err != nil {
		t.Fatal(err)
	}
	specPath := filepath.Join(l.SpecificationsDir(), "SPEC-001-password-reset.md")

	v1, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Approve(root, "SPEC-001", nil, nil); err != nil {
		t.Fatal(err)
	}
	v2 := append(append([]byte{}, v1...), []byte("\nv2 sentence.\n")...)
	if err := os.WriteFile(specPath, v2, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Approve(root, "SPEC-001", nil, nil); err != nil {
		t.Fatal(err)
	}

	revs, err := approval.Revisions(l.ApprovalsDir(), "SPEC-001")
	if err != nil {
		t.Fatal(err)
	}
	if len(revs) != 2 {
		t.Fatalf("expected 2 revisions, got %d", len(revs))
	}
	v1GrantID, v2GrantID := revs[0].GrantID, revs[1].GrantID

	// Directly revoke the older (non-L) grant — orchestrate.Revoke only ever acts on the current
	// active grant, so this bypasses it to set up a "revoked but not L" revision.
	if err := approval.WriteRevocation(l.ApprovalsDir(), "SPEC-001", v1GrantID, "Jane <jane@example.com>"); err != nil {
		t.Fatal(err)
	}

	// Move content past L (v2) too, so L itself becomes superseded rather than active.
	v3 := append(append([]byte{}, v2...), []byte("\nv3 sentence.\n")...)
	if err := os.WriteFile(specPath, v3, 0o644); err != nil {
		t.Fatal(err)
	}

	detail, err := SpecDetailFor(l, "SPEC-001")
	if err != nil {
		t.Fatal(err)
	}
	if detail.Lifecycle != approval.Draft {
		t.Fatalf("expected Draft, got %s", detail.Lifecycle)
	}
	if detail.ActiveGrant != nil {
		t.Fatalf("expected no active grant (superseded, not active), got %+v", detail.ActiveGrant)
	}
	byID := map[string]approval.Revision{}
	for _, r := range detail.Revisions {
		byID[r.GrantID] = r
	}
	if !byID[v1GrantID].Revoked {
		t.Fatalf("expected the v1 grant marked revoked, got %+v", byID[v1GrantID])
	}
	if byID[v2GrantID].Revoked {
		t.Fatalf("expected L (v2) not revoked — it is superseded by content, not revoked, got %+v", byID[v2GrantID])
	}
}
