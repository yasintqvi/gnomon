// Package approval implements CLI Step 6 — Approval Runtime: durable, append-only, immutable
// evidence, and the derived-state rule for whether a Specification's current content is
// Approved. It never persists a mutable "state: approved" field anywhere.
package approval

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"gnomon/internal/specs"
)

// Lifecycle is a Specification's derived lifecycle state.
type Lifecycle string

const (
	Draft    Lifecycle = "Draft"
	Approved Lifecycle = "Approved"
)

// Grant is one immutable approval record.
type Grant struct {
	SpecIdentity string `json:"spec_identity"`
	Fingerprint  string `json:"fingerprint"`
	Approver     string `json:"approver"`
	Timestamp    string `json:"timestamp"`
}

// Revocation is one immutable revocation record, referencing the grant it revokes.
type Revocation struct {
	Revokes   string `json:"revokes"`
	RevokedBy string `json:"revoked_by"`
	Timestamp string `json:"timestamp"`
}

// Fingerprint computes the whitespace-normalized content fingerprint of Specification body text.
// Only whitespace/formatting is normalized away; everything else is conservatively
// approval-relevant, per SPECIFICATION_LIFECYCLE.md's default-to-relevant-on-ambiguity rule.
func Fingerprint(content []byte) string {
	sum := sha256.Sum256([]byte(normalize(string(content))))
	return hex.EncodeToString(sum[:])
}

func normalize(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	lines := strings.Split(s, "\n")
	for i, l := range lines {
		lines[i] = strings.TrimRight(l, " \t")
	}
	out := make([]string, 0, len(lines))
	blank := false
	for _, l := range lines {
		if l == "" {
			if blank {
				continue
			}
			blank = true
		} else {
			blank = false
		}
		out = append(out, l)
	}
	return strings.Join(out, "\n")
}

func newRecordID() (string, error) {
	b := make([]byte, 6)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func specDir(approvalsDir, specID string) string {
	return filepath.Join(approvalsDir, specID)
}

// WriteGrant writes one new, immutable grant record, plus a sibling content snapshot
// (<id>.content.md) sharing the same record id — the entire revision-history mechanism: a
// revision is an approval grant with its content preserved alongside it. Neither file is ever
// edited or removed once written.
func WriteGrant(approvalsDir, specID, fingerprint, approver string, content []byte) error {
	dir := specDir(approvalsDir, specID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	id, err := newRecordID()
	if err != nil {
		return err
	}
	ts, err := nextGrantTimestamp(approvalsDir, specID)
	if err != nil {
		return err
	}
	g := Grant{
		SpecIdentity: specID,
		Fingerprint:  fingerprint,
		Approver:     approver,
		Timestamp:    ts.Format(time.RFC3339Nano),
	}
	data, err := json.MarshalIndent(g, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, id+".content.md"), content, 0o644); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, id+".grant.json"), data, 0o644)
}

// nextGrantTimestamp returns the current time, or the latest existing grant's timestamp plus one
// nanosecond if that would not be strictly later — coarse clock resolution (notably on Windows)
// can otherwise give two back-to-back grants an equal or out-of-order timestamp.
func nextGrantTimestamp(approvalsDir, specID string) (time.Time, error) {
	now := time.Now().UTC()
	grants, _, err := scanEvidence(approvalsDir, specID)
	if err != nil {
		return time.Time{}, err
	}
	var latest time.Time
	for _, g := range grants {
		if t := parseTimestamp(g.Timestamp); t.After(latest) {
			latest = t
		}
	}
	if !now.After(latest) {
		return latest.Add(time.Nanosecond), nil
	}
	return now, nil
}

// WriteRevocation writes one new, immutable revocation record referencing grantID.
func WriteRevocation(approvalsDir, specID, grantID, revokedBy string) error {
	dir := specDir(approvalsDir, specID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	id, err := newRecordID()
	if err != nil {
		return err
	}
	r := Revocation{
		Revokes:   grantID,
		RevokedBy: revokedBy,
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
	}
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, id+".revocation.json"), data, 0o644)
}

// parseTimestamp parses a Grant/Revocation record's stored timestamp. An empty or unparseable
// value returns the zero time, so ordering by it sorts such a record first rather than erroring.
func parseTimestamp(s string) time.Time {
	t, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		return time.Time{}
	}
	return t
}

// Derive computes the current lifecycle state for a Specification given its current fingerprint —
// recomputed fresh from whatever evidence files currently exist, never cached.
func Derive(approvalsDir, specID, currentFingerprint string) (Lifecycle, error) {
	_, approved, err := ActiveGrant(approvalsDir, specID, currentFingerprint)
	if err != nil {
		return Draft, err
	}
	if approved {
		return Approved, nil
	}
	return Draft, nil
}

// ActiveGrant is the single source of truth for whether a Specification is Approved: "the latest
// approval decision counts." L is the latest valid grant, by the same ordering Revisions displays
// (sortedGrantIDs) — chosen regardless of its own revocation status, so revoking any grant other
// than L never changes this answer. approved is true only when L exists, is not revoked, and its
// fingerprint matches currentFingerprint. Every caller deciding Approved/Draft, or picking "the
// active grant," goes through this function — directly, or via Derive.
func ActiveGrant(approvalsDir, specID, currentFingerprint string) (grantID string, approved bool, err error) {
	grants, revocations, err := scanEvidence(approvalsDir, specID)
	if err != nil {
		return "", false, err
	}
	ids := sortedGrantIDs(grants)
	if len(ids) == 0 {
		return "", false, nil
	}
	latest := ids[len(ids)-1]
	approved = revocations[latest] == nil && grants[latest].Fingerprint == currentFingerprint
	return latest, approved, nil
}

// ActiveGrantIDs returns the ids of every unrevoked grant record matching currentFingerprint —
// used by Revoke to also clear legacy duplicate grants for the same content. Never a decision
// about Approved/Draft; see ActiveGrant for that.
func ActiveGrantIDs(approvalsDir, specID, currentFingerprint string) ([]string, error) {
	return activeGrants(approvalsDir, specID, currentFingerprint)
}

func activeGrants(approvalsDir, specID, currentFingerprint string) ([]string, error) {
	grants, revocations, err := scanEvidence(approvalsDir, specID)
	if err != nil {
		return nil, err
	}
	var ids []string
	for id, g := range grants {
		if revocations[id] != nil {
			continue
		}
		if g.Fingerprint == currentFingerprint {
			ids = append(ids, id)
		}
	}
	sort.Strings(ids)
	return ids, nil
}

// sortedGrantIDs orders grant ids the same way Revisions displays them: parsed timestamp, then
// grant id as a tie-break for legacy same-second grants. The one ordering "the latest grant" (L)
// is decided by anywhere in this package — Revisions and ActiveGrant both use it, never a second
// comparator.
func sortedGrantIDs(grants map[string]Grant) []string {
	ids := make([]string, 0, len(grants))
	for id := range grants {
		ids = append(ids, id)
	}
	sort.SliceStable(ids, func(i, j int) bool {
		ti, tj := parseTimestamp(grants[ids[i]].Timestamp), parseTimestamp(grants[ids[j]].Timestamp)
		if !ti.Equal(tj) {
			return ti.Before(tj)
		}
		return ids[i] < ids[j]
	})
	return ids
}

// IssueLevel distinguishes a problem that excludes a record from derivation (Error) from one
// that is only ever reported (Warning).
type IssueLevel string

const (
	LevelError   IssueLevel = "error"
	LevelWarning IssueLevel = "warning"
)

// Problem is one thing wrong with a single evidence file, found by checkGrant or checkRevocation
// — the one place either record type's validity is decided, so derivation and reporting never
// disagree about a record.
type Problem struct {
	Level  IssueLevel
	Reason string
}

func hasError(problems []Problem) bool {
	for _, p := range problems {
		if p.Level == LevelError {
			return true
		}
	}
	return false
}

// checkGrant parses and validates one grant record; specID is checked against the record's own
// spec_identity. Always returns the parsed (possibly invalid) Grant — callers check hasError.
func checkGrant(specID string, data []byte) (Grant, []Problem) {
	var g Grant
	if err := json.Unmarshal(data, &g); err != nil {
		return g, []Problem{{LevelError, fmt.Sprintf("malformed JSON: %v", err)}}
	}
	var problems []Problem
	if g.SpecIdentity == "" || g.Fingerprint == "" || g.Approver == "" {
		problems = append(problems, Problem{LevelError, "missing required field (spec_identity, fingerprint, or approver)"})
	}
	if g.SpecIdentity != "" && g.SpecIdentity != specID {
		problems = append(problems, Problem{LevelError, fmt.Sprintf("spec_identity %q does not match its directory %q", g.SpecIdentity, specID)})
	}
	if p := checkTimestamp(g.Timestamp); p != nil {
		problems = append(problems, *p)
	}
	return g, problems
}

// checkRevocation parses and validates one revocation record.
func checkRevocation(data []byte) (Revocation, []Problem) {
	var r Revocation
	if err := json.Unmarshal(data, &r); err != nil {
		return r, []Problem{{LevelError, fmt.Sprintf("malformed JSON: %v", err)}}
	}
	var problems []Problem
	if r.Revokes == "" || r.RevokedBy == "" {
		problems = append(problems, Problem{LevelError, "missing required field (revokes or revoked_by)"})
	}
	if p := checkTimestamp(r.Timestamp); p != nil {
		problems = append(problems, *p)
	}
	return r, problems
}

// checkTimestamp reports the one error-level problem a record's timestamp field can have:
// missing, or not parseable as RFC 3339 (time.RFC3339Nano also accepts second precision).
func checkTimestamp(ts string) *Problem {
	if ts == "" {
		return &Problem{LevelError, "timestamp is missing"}
	}
	if _, err := time.Parse(time.RFC3339Nano, ts); err != nil {
		return &Problem{LevelError, fmt.Sprintf("timestamp %q is not a valid RFC 3339 timestamp", ts)}
	}
	return nil
}

// scanEvidence reads every grant/revocation file for one Specification once, so activeGrants and
// Revisions derive from identical parsing. A record with any error-level problem is excluded here,
// exactly as ValidateEvidence reports it invalid. revocations is keyed by the grant id it revokes.
func scanEvidence(approvalsDir, specID string) (map[string]Grant, map[string]*Revocation, error) {
	dir := specDir(approvalsDir, specID)
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return nil, nil, nil
	}
	if err != nil {
		return nil, nil, err
	}

	grants := map[string]Grant{}
	revocations := map[string]*Revocation{}

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		data, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			continue
		}
		switch {
		case strings.HasSuffix(name, ".grant.json"):
			g, problems := checkGrant(specID, data)
			if hasError(problems) {
				continue
			}
			grants[strings.TrimSuffix(name, ".grant.json")] = g
		case strings.HasSuffix(name, ".revocation.json"):
			r, problems := checkRevocation(data)
			if hasError(problems) {
				continue
			}
			revocations[r.Revokes] = &r
		}
	}
	return grants, revocations, nil
}

// Revision is one durably-evidenced approval grant, together with the content it was granted
// against. Content is empty (ContentKnown false) only when a snapshot was never captured or
// couldn't be read — represented honestly as unavailable, never reconstructed.
type Revision struct {
	GrantID      string
	Fingerprint  string
	Approver     string
	ApprovedAt   string
	Content      string
	ContentKnown bool
	Revoked      bool
	RevokedBy    string
	RevokedAt    string
}

// Revisions returns every revision this Specification has ever had a grant for, oldest first —
// derived entirely from existing approval evidence, never a separately persisted history. This
// reflects only Human approval acts, never execution history.
func Revisions(approvalsDir, specID string) ([]Revision, error) {
	grants, revocations, err := scanEvidence(approvalsDir, specID)
	if err != nil {
		return nil, err
	}

	ids := sortedGrantIDs(grants)
	revisions := make([]Revision, 0, len(ids))
	for _, id := range ids {
		g := grants[id]
		rev := Revision{
			GrantID:     id,
			Fingerprint: g.Fingerprint,
			Approver:    g.Approver,
			ApprovedAt:  g.Timestamp,
		}
		if content, err := os.ReadFile(filepath.Join(specDir(approvalsDir, specID), id+".content.md")); err == nil {
			rev.Content = string(content)
			rev.ContentKnown = true
		}
		if r := revocations[id]; r != nil {
			rev.Revoked = true
			rev.RevokedBy = r.RevokedBy
			rev.RevokedAt = r.Timestamp
		}
		revisions = append(revisions, rev)
	}
	return revisions, nil
}

// EvidenceIssue is one problem found in a grant/revocation file or an evidence directory as a
// whole — never a judgment about Draft/Approved, only about the evidence itself. Level Error means
// scanEvidence excludes the record from derivation; Level Warning means it's reported only.
type EvidenceIssue struct {
	SpecID   string
	Filename string // "" for a directory-level issue (e.g. no matching Specification)
	Level    IssueLevel
	Problem  string
}

// ValidateEvidence scans every approval evidence file under approvalsDir and reports every problem
// checkGrant/checkRevocation find, plus directory-wide warnings: a revocation referencing a
// nonexistent grant id, an orphaned content snapshot, duplicate revocations, and an evidence
// directory with no matching Specification file (specificationsDir). Never repairs or rewrites
// anything.
func ValidateEvidence(approvalsDir, specificationsDir string) ([]EvidenceIssue, error) {
	entries, err := os.ReadDir(approvalsDir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var issues []EvidenceIssue
	for _, specEntry := range entries {
		if !specEntry.IsDir() {
			continue
		}
		specID := specEntry.Name()
		dir := filepath.Join(approvalsDir, specID)
		files, err := os.ReadDir(dir)
		if err != nil {
			issues = append(issues, EvidenceIssue{SpecID: specID, Level: LevelError, Problem: err.Error()})
			continue
		}

		grantIDsPresent := map[string]bool{}
		for _, f := range files {
			if strings.HasSuffix(f.Name(), ".grant.json") {
				grantIDsPresent[strings.TrimSuffix(f.Name(), ".grant.json")] = true
			}
		}

		revocationFilesFor := map[string][]string{} // valid revocations' grant id -> their own filenames
		for _, f := range files {
			name := f.Name()
			data, err := os.ReadFile(filepath.Join(dir, name))
			if err != nil {
				issues = append(issues, EvidenceIssue{SpecID: specID, Filename: name, Level: LevelError, Problem: err.Error()})
				continue
			}
			switch {
			case strings.HasSuffix(name, ".grant.json"):
				_, problems := checkGrant(specID, data)
				for _, p := range problems {
					issues = append(issues, EvidenceIssue{SpecID: specID, Filename: name, Level: p.Level, Problem: p.Reason})
				}
			case strings.HasSuffix(name, ".revocation.json"):
				r, problems := checkRevocation(data)
				for _, p := range problems {
					issues = append(issues, EvidenceIssue{SpecID: specID, Filename: name, Level: p.Level, Problem: p.Reason})
				}
				if hasError(problems) {
					continue
				}
				if !grantIDsPresent[r.Revokes] {
					issues = append(issues, EvidenceIssue{SpecID: specID, Filename: name, Level: LevelWarning,
						Problem: fmt.Sprintf("revokes grant id %q, which does not exist in this directory", r.Revokes)})
				}
				revocationFilesFor[r.Revokes] = append(revocationFilesFor[r.Revokes], name)
			case strings.HasSuffix(name, ".content.md"):
				if id := strings.TrimSuffix(name, ".content.md"); !grantIDsPresent[id] {
					issues = append(issues, EvidenceIssue{SpecID: specID, Filename: name, Level: LevelWarning,
						Problem: "content snapshot has no matching grant"})
				}
			default:
				issues = append(issues, EvidenceIssue{SpecID: specID, Filename: name, Level: LevelError,
					Problem: "unrecognized evidence filename (expected *.grant.json, *.revocation.json, or *.content.md)"})
			}
		}

		var duplicated []string
		for grantID, revNames := range revocationFilesFor {
			if len(revNames) > 1 {
				duplicated = append(duplicated, grantID)
			}
		}
		sort.Strings(duplicated)
		for _, grantID := range duplicated {
			names := revocationFilesFor[grantID]
			sort.Strings(names)
			for _, name := range names {
				issues = append(issues, EvidenceIssue{SpecID: specID, Filename: name, Level: LevelWarning,
					Problem: fmt.Sprintf("grant id %q has %d revocations", grantID, len(names))})
			}
		}

		// A missing specifications directory means no Specification exists at all — the same
		// warning as any other identity with no matching file, not a hard error.
		ok, _, existsErr := specs.Exists(specificationsDir, specs.Identity(specID))
		if existsErr != nil && !os.IsNotExist(existsErr) {
			return nil, existsErr
		}
		if !ok {
			issues = append(issues, EvidenceIssue{SpecID: specID, Level: LevelWarning, Problem: "no Specification file exists for this identity"})
		}
	}
	return issues, nil
}
