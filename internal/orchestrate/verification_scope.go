package orchestrate

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"gnomon/internal/gitutil"
	"gnomon/internal/present"
	"gnomon/internal/project"
	"gnomon/internal/result"
)

// VerificationOptions selects how much a Verification run checks.
//
//   - Standard (the default): with no target, only what changed since the last passing
//     verification (or since Since); with a target, that target. A passing test may serve as
//     evidence for the criterion it actually asserts.
//   - Full: every criterion of the target (all Approved Specifications without one), each checked
//     independently; a passing test alone is not enough evidence.
//
// Conflict checks always cover every Approved Specification, whatever the scope.
type VerificationOptions struct {
	Full  bool
	Since string // a Git ref to measure the change from, instead of the last passing verification
}

// verificationStateName is the record of the last passing verification that covered every change,
// in the project's transient directory: local to this working copy, never committed.
const verificationStateName = "verification-state.json"

// verificationState is what the next standard run measures "changed" against: the commit verified,
// plus the content of every file that differed from that commit at the time (a verification of
// uncommitted work), so those files count as changed again only if they change again.
type verificationState struct {
	Commit      string            `json:"commit"`
	Uncommitted map[string]string `json:"uncommitted,omitempty"` // path -> content hash ("" = deleted)
	VerifiedAt  string            `json:"verified_at"`
}

// verificationScope is what a run will check, decided before any Agent starts.
type verificationScope struct {
	note        string   // the scope text handed to the Agent
	recordState bool     // a PASS covers every change, so it becomes the new baseline
	base        string   // the commit changes are measured from, "" when none
	files       []string // changed files, for recording the baseline
}

func verificationStatePath(projectRoot string) (string, error) {
	dir, err := result.TransientDir(projectRoot)
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, verificationStateName), nil
}

func readVerificationState(projectRoot string) *verificationState {
	path, err := verificationStatePath(projectRoot)
	if err != nil {
		return nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var s verificationState
	if json.Unmarshal(data, &s) != nil || s.Commit == "" {
		return nil
	}
	return &s
}

func contentHash(projectRoot, rel string) string {
	data, err := os.ReadFile(filepath.Join(projectRoot, filepath.FromSlash(rel)))
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

// decideVerificationScope works out what a Verification run checks. A non-nil report means there
// is nothing to check and no Agent should start.
func decideVerificationScope(l project.Layout, target string, opts VerificationOptions) (*verificationScope, *present.Report, error) {
	conflictLine := "Check conflicts against every Approved Specification, also those outside this scope."

	if opts.Full {
		what := target
		if what == "" {
			what = "every Approved Specification"
		}
		return &verificationScope{
			note: fmt.Sprintf("Verification mode: full independent check of %s.\n"+
				"Check every criterion directly (run it, or for what cannot be run, the exact code that implements it). A passing test alone is not enough evidence in this mode.\n%s", what, conflictLine),
			recordState: target == "",
		}, nil, nil
	}

	if target != "" {
		return &verificationScope{
			note: fmt.Sprintf("Verification mode: standard check of %s.\n%s", target, conflictLine),
		}, nil, nil
	}

	if st, err := gitutil.Inspect(l.Root); err != nil || !st.Available {
		return &verificationScope{
			note: "Verification mode: standard check of every Approved Specification (this project is not in a Git repository, so changes cannot be measured).\n" + conflictLine,
		}, nil, nil
	}

	head, err := gitutil.Head(l.Root)
	if err != nil {
		if opts.Since != "" {
			return nil, nil, err
		}
		return &verificationScope{
			note: "Verification mode: standard check of every Approved Specification (this repository has no commit to measure changes from).\n" + conflictLine,
		}, nil, nil
	}

	var base, baseLabel string
	var state *verificationState
	switch {
	case opts.Since != "":
		base, err = gitutil.ResolveCommit(l.Root, opts.Since)
		if err != nil {
			return nil, nil, fmt.Errorf("--since %s: %w", opts.Since, err)
		}
		baseLabel = fmt.Sprintf("%s (%s)", opts.Since, short(base))
	default:
		state = readVerificationState(l.Root)
		if state != nil {
			if _, err := gitutil.ResolveCommit(l.Root, state.Commit); err != nil {
				state = nil // the recorded commit is gone (history rewritten): start over
			}
		}
		if state == nil {
			return &verificationScope{
				note:        "Verification mode: standard check of every Approved Specification (no earlier passing verification is recorded here, so there is no change to limit it to).\n" + conflictLine,
				recordState: true,
				base:        head,
			}, nil, nil
		}
		base = state.Commit
		baseLabel = fmt.Sprintf("the last passing verification (%s, %s)", short(base), state.VerifiedAt)
	}

	changed, err := gitutil.ChangedSince(l.Root, base)
	if err != nil {
		return nil, nil, err
	}
	var files []string
	for _, f := range changed {
		if state != nil {
			if h, ok := state.Uncommitted[f]; ok && h == contentHash(l.Root, f) {
				continue // verified as it is now
			}
		}
		files = append(files, f)
	}

	if len(files) == 0 {
		rep := &present.Report{
			Outcome: present.Success,
			Summary: "Verification: nothing changed since " + baseLabel,
			Sections: []present.Section{
				{Label: "Result", Body: "No Agent was started and nothing was checked."},
			},
			Next: "Run `gnomon run verification --full` for a full independent check of every Approved Specification.",
		}
		if opts.Since != "" {
			rep.Outcome = present.Blocked
			rep.Summary = "Verification: nothing changed since " + baseLabel + ", so nothing was verified"
		}
		return nil, rep, nil
	}

	changedSpecs, otherFiles := splitChangedFiles(l, files)
	var b strings.Builder
	fmt.Fprintf(&b, "Verification mode: standard check of the changes since %s.\n", baseLabel)
	if len(changedSpecs) > 0 {
		fmt.Fprintf(&b, "Specifications whose text or approval changed: %s.\n", strings.Join(changedSpecs, ", "))
	} else {
		b.WriteString("No Specification's text or approval changed.\n")
	}
	if len(otherFiles) > 0 {
		fmt.Fprintf(&b, "Other changed files (%d): %s\n", len(otherFiles), listFiles(otherFiles, 200))
	}
	b.WriteString("Verify the criteria of the changed Specifications and of every Approved Specification whose behavior the changed files implement or test; say in summary.obligations_evaluated which ones you chose and why. Do not re-check Specifications the change cannot affect.\n")
	b.WriteString(conflictLine)

	return &verificationScope{note: b.String(), recordState: true, base: base, files: files}, nil, nil
}

// splitChangedFiles separates the Specifications whose file or approval evidence changed from the
// other changed files.
func splitChangedFiles(l project.Layout, files []string) (changedSpecs, others []string) {
	specsRel := relSlash(l.Root, l.SpecificationsDir()) + "/"
	approvalsRel := relSlash(l.Root, l.ApprovalsDir()) + "/"
	seen := map[string]bool{}
	for _, f := range files {
		var id string
		switch {
		case strings.HasPrefix(f, specsRel):
			id = specIdentityOf(strings.TrimPrefix(f, specsRel))
		case strings.HasPrefix(f, approvalsRel):
			id = strings.SplitN(strings.TrimPrefix(f, approvalsRel), "/", 2)[0]
		default:
			others = append(others, f)
			continue
		}
		if id != "" && !seen[id] {
			seen[id] = true
			changedSpecs = append(changedSpecs, id)
		}
	}
	sort.Strings(changedSpecs)
	return changedSpecs, others
}

var specFilePattern = regexp.MustCompile(`^(SPEC-\d+)(-.*)?\.md$`)

// specIdentityOf returns "SPEC-007" for "SPEC-007-late-fees.md", "" for anything else.
func specIdentityOf(name string) string {
	if m := specFilePattern.FindStringSubmatch(name); m != nil {
		return m[1]
	}
	return ""
}

func relSlash(root, path string) string {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return filepath.ToSlash(path)
	}
	return filepath.ToSlash(rel)
}

func listFiles(files []string, max int) string {
	if len(files) <= max {
		return strings.Join(files, ", ")
	}
	return strings.Join(files[:max], ", ") + fmt.Sprintf(", and %d more", len(files)-max)
}

func short(commit string) string {
	if len(commit) > 12 {
		return commit[:12]
	}
	return commit
}

// recordVerificationState makes a passing run that covered every change the new baseline: the
// commit at the start of the run, plus the content of each changed file not yet committed.
func recordVerificationState(projectRoot string, scope *verificationScope) error {
	head, err := gitutil.Head(projectRoot)
	if err != nil {
		return err
	}
	dirty, err := gitutil.ChangedSince(projectRoot, head)
	if err != nil {
		return err
	}
	s := verificationState{Commit: head, Uncommitted: map[string]string{}, VerifiedAt: time.Now().UTC().Format(time.RFC3339)}
	for _, f := range dirty {
		s.Uncommitted[f] = contentHash(projectRoot, f)
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	path, err := verificationStatePath(projectRoot)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
