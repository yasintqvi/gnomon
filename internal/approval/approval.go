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
	"strings"
	"time"
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

// WriteGrant writes one new, immutable grant record. It never edits or removes any prior record.
func WriteGrant(approvalsDir, specID, fingerprint, approver string) error {
	dir := specDir(approvalsDir, specID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	id, err := newRecordID()
	if err != nil {
		return err
	}
	g := Grant{
		SpecIdentity: specID,
		Fingerprint:  fingerprint,
		Approver:     approver,
		Timestamp:    time.Now().UTC().Format(time.RFC3339),
	}
	data, err := json.MarshalIndent(g, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, id+".grant.json"), data, 0o644)
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
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, id+".revocation.json"), data, 0o644)
}

// Derive computes the current lifecycle state for a Specification given its current fingerprint —
// recomputed fresh from whatever evidence files currently exist, never cached.
func Derive(approvalsDir, specID, currentFingerprint string) (Lifecycle, error) {
	_, ok, err := activeGrant(approvalsDir, specID, currentFingerprint)
	if err != nil {
		return Draft, err
	}
	if ok {
		return Approved, nil
	}
	return Draft, nil
}

// ActiveGrantID returns the id of the grant record currently causing the Specification to derive
// Approved for the given fingerprint, if any — exactly cli/APPROVAL_RUNTIME.md's Revocation
// Operation step 1: "Locate the grant record currently causing the Specification to derive
// Approved, if any." This is the same grant Derive's own rule already finds; it is exposed
// separately only because Revocation needs the grant's own identifier (to write a Revocation
// record referencing it), not merely the Draft/Approved answer Derive returns.
func ActiveGrantID(approvalsDir, specID, currentFingerprint string) (string, bool, error) {
	id, ok, err := activeGrant(approvalsDir, specID, currentFingerprint)
	if err != nil {
		return "", false, err
	}
	return id, ok, nil
}

// activeGrant is the one shared derivation both Derive and ActiveGrantID read from, so the
// "which grant, if any, currently causes Approved" question is answered by exactly one piece of
// logic — never reimplemented or allowed to drift between the two callers.
func activeGrant(approvalsDir, specID, currentFingerprint string) (string, bool, error) {
	dir := specDir(approvalsDir, specID)
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}

	grants := map[string]Grant{}
	revoked := map[string]bool{}

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		data, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			continue // unreadable evidence is excluded, never treated as valid
		}
		switch {
		case strings.HasSuffix(name, ".grant.json"):
			var g Grant
			if err := json.Unmarshal(data, &g); err != nil || g.Approver == "" || g.Fingerprint == "" {
				continue // malformed or unattributed — excluded from consideration
			}
			grants[strings.TrimSuffix(name, ".grant.json")] = g
		case strings.HasSuffix(name, ".revocation.json"):
			var r Revocation
			if err := json.Unmarshal(data, &r); err != nil || r.Revokes == "" {
				continue
			}
			revoked[r.Revokes] = true
		}
	}

	for id, g := range grants {
		if revoked[id] {
			continue
		}
		if g.Fingerprint == currentFingerprint {
			return id, true, nil
		}
	}
	return "", false, nil
}

// EvidenceIssue is one structural problem found in a single grant or revocation file — never a
// judgment about whether a Specification is correctly Draft/Approved, only about whether the
// evidence file itself has the shape cli/APPROVAL_RUNTIME.md's Grant/Revocation Record sections
// require.
type EvidenceIssue struct {
	SpecID   string
	Filename string
	Problem  string
}

// ValidateEvidence scans every approval evidence file under approvalsDir and reports any that
// fail Step 6's own record-shape requirements — unparseable JSON, a required field missing or
// empty (spec_identity/fingerprint/approver for a grant; revokes/revoked_by for a revocation), or
// a filename that is neither a *.grant.json nor a *.revocation.json. It never repairs, deletes,
// or rewrites anything — purely a report. This is exactly the same class of file Derive/
// activeGrant already silently exclude from lifecycle derivation (never treated as valid
// evidence there either); ValidateEvidence is what finally surfaces cli/APPROVAL_RUNTIME.md's own
// "diagnostic warning surfaced" requirement for malformed evidence, which nothing implemented
// before gnomon validate.
func ValidateEvidence(approvalsDir string) ([]EvidenceIssue, error) {
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
			issues = append(issues, EvidenceIssue{SpecID: specID, Problem: err.Error()})
			continue
		}
		for _, f := range files {
			name := f.Name()
			data, err := os.ReadFile(filepath.Join(dir, name))
			if err != nil {
				issues = append(issues, EvidenceIssue{SpecID: specID, Filename: name, Problem: err.Error()})
				continue
			}
			switch {
			case strings.HasSuffix(name, ".grant.json"):
				var g Grant
				if err := json.Unmarshal(data, &g); err != nil {
					issues = append(issues, EvidenceIssue{SpecID: specID, Filename: name, Problem: fmt.Sprintf("malformed JSON: %v", err)})
					continue
				}
				if g.SpecIdentity == "" || g.Fingerprint == "" || g.Approver == "" {
					issues = append(issues, EvidenceIssue{SpecID: specID, Filename: name, Problem: "missing required field (spec_identity, fingerprint, or approver)"})
				}
			case strings.HasSuffix(name, ".revocation.json"):
				var r Revocation
				if err := json.Unmarshal(data, &r); err != nil {
					issues = append(issues, EvidenceIssue{SpecID: specID, Filename: name, Problem: fmt.Sprintf("malformed JSON: %v", err)})
					continue
				}
				if r.Revokes == "" || r.RevokedBy == "" {
					issues = append(issues, EvidenceIssue{SpecID: specID, Filename: name, Problem: "missing required field (revokes or revoked_by)"})
				}
			default:
				issues = append(issues, EvidenceIssue{SpecID: specID, Filename: name, Problem: "unrecognized evidence filename (expected *.grant.json or *.revocation.json)"})
			}
		}
	}
	return issues, nil
}
