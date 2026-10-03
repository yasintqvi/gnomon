package orchestrate

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gnomon/internal/approval"
	"gnomon/internal/facts"
	"gnomon/internal/present"
	"gnomon/internal/project"
	"gnomon/internal/result"
	"gnomon/internal/specs"
)

// fileSnapshot maps a file's absolute path to its exact bytes, taken before or after a
// change-sensitive Agent run. The run guard compares two snapshots of the same root to detect any
// modification the Agent made outside its own responsibility — used for both the approvals
// directory and (Rule A) the governing Specification's file.
type fileSnapshot map[string][]byte

// snapshotTree reads every regular file under root recursively. A root that does not exist (or
// no longer exists) snapshots as empty — nothing to protect yet, never an error.
func snapshotTree(root string) (fileSnapshot, error) {
	snap := fileSnapshot{}
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		if info.IsDir() {
			return nil
		}
		data, rerr := os.ReadFile(path)
		if rerr != nil {
			return rerr
		}
		snap[path] = data
		return nil
	})
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	return snap, nil
}

// fileDiff is every path added, changed, or removed between two snapshots of the same root —
// absolute paths, sorted, so a report over them reads deterministically.
type fileDiff struct {
	Added, Changed, Removed []string
}

func (d fileDiff) empty() bool {
	return len(d.Added) == 0 && len(d.Changed) == 0 && len(d.Removed) == 0
}

// diffSnapshots compares two snapshots of the same root.
func diffSnapshots(before, after fileSnapshot) fileDiff {
	var d fileDiff
	for path, data := range after {
		old, existed := before[path]
		switch {
		case !existed:
			d.Added = append(d.Added, path)
		case !bytes.Equal(old, data):
			d.Changed = append(d.Changed, path)
		}
	}
	for path := range before {
		if _, stillThere := after[path]; !stillThere {
			d.Removed = append(d.Removed, path)
		}
	}
	sort.Strings(d.Added)
	sort.Strings(d.Changed)
	sort.Strings(d.Removed)
	return d
}

// restoreSnapshot undoes diff against before: added files deleted, changed files rewritten,
// removed files recreated. This is not a violation of evidence's append-only rule — it undoes
// bytes that were never valid evidence. Continues past a single failure; err is the first error.
func restoreSnapshot(before fileSnapshot, diff fileDiff) (failed []string, err error) {
	note := func(path string, e error) {
		if e == nil {
			return
		}
		failed = append(failed, path)
		if err == nil {
			err = e
		}
	}
	for _, path := range diff.Added {
		note(path, os.Remove(path))
	}
	for _, path := range diff.Changed {
		note(path, os.WriteFile(path, before[path], 0o644))
	}
	for _, path := range diff.Removed {
		if e := os.MkdirAll(filepath.Dir(path), 0o755); e != nil {
			note(path, e)
			continue
		}
		note(path, os.WriteFile(path, before[path], 0o644))
	}
	return failed, err
}

// approvalsResult is what the approvals-directory check (protecting .gnomon/approvals/) found
// and did, for buildRejectionReport.
type approvalsResult struct {
	Dir           string
	Diff          fileDiff
	RestoreFailed []string
	RestoreErr    error
}

func (a approvalsResult) violated() bool { return !a.Diff.empty() }

// specResult is what Rule A (the governing Specification must not change) found and did, if it
// applied to this run at all — Violation is nil when Rule A does not apply, or applied and found
// nothing wrong.
type specResult struct {
	Before        *specSnapshot
	Violation     *specViolation
	Preserved     []string // absolute paths under rejected/<run-id>/ the Agent's version was saved to
	RestoreFailed []string
	RestoreErr    error
}

func (s specResult) violated() bool { return s.Violation != nil }

// buildRejectionReport builds the one combined Failed report for whichever of the two checks
// found a problem — both, if both did, per the "report both in one rejected report" requirement.
func buildRejectionReport(a approvalsResult, s specResult) *present.Report {
	var summaryParts []string
	var sections []present.Section
	restorationFailed := false

	if a.violated() {
		summaryParts = append(summaryParts, "modified approval evidence")
		sections = append(sections, present.Section{Label: "Approval Evidence", Body: describeDiff(a.Dir, a.Diff)})
		if a.RestoreErr != nil {
			restorationFailed = true
			sections = append(sections, present.Section{
				Label: "Approval Evidence Not Restored",
				Body: fmt.Sprintf("Could not restore:\n%s\n\nInspect .gnomon/approvals/ directly before trusting `gnomon status`.",
					strings.Join(relPaths(a.Dir, a.RestoreFailed), "\n")),
			})
		} else {
			sections = append(sections, present.Section{
				Label: "Approval Evidence Result",
				Body:  "The approvals directory was restored to its state before this run.",
			})
		}
	}

	if s.violated() {
		verb := specViolationVerb(s.Violation.Kind)
		summaryParts = append(summaryParts, fmt.Sprintf("%s the approved Specification %s", verb, s.Before.Identity))

		body := fmt.Sprintf("%s was %s during this run.", s.Before.Identity, verb)
		if s.Violation.Kind == "renamed" && s.Violation.NewPath != "" {
			body = fmt.Sprintf("%s was renamed during this run (the Agent created %s under the same identity).",
				s.Before.Identity, filepath.Base(s.Violation.NewPath))
		}
		if len(s.Preserved) > 0 {
			body += "\n\nThe Agent's version was saved for inspection at:\n" + strings.Join(s.Preserved, "\n")
		} else {
			body += "\n\nThe Agent left no file behind under this identity to preserve."
		}
		sections = append(sections, present.Section{Label: "Governing Specification", Body: body})
		sections = append(sections, present.Section{
			Label: "Code Changes",
			Body:  "Any other changes the Agent made elsewhere in the repository were NOT reverted — the working tree may contain work done against a changed requirement.",
		})

		if s.RestoreErr != nil {
			restorationFailed = true
			sections = append(sections, present.Section{
				Label: "Specification Not Restored",
				Body: fmt.Sprintf("Could not restore:\n%s\n\nInspect .gnomon/specifications/ directly before trusting `gnomon status`.",
					strings.Join(s.RestoreFailed, "\n")),
			})
		} else {
			sections = append(sections, present.Section{Label: "Governing Specification Result", Body: fmt.Sprintf("%s was restored to its state before this run.", s.Before.Identity)})
			sections = append(sections, present.Section{
				Label: "Next",
				Body: fmt.Sprintf("If this change is wanted, revise %s through Define and re-approve it — this run's proposal was not accepted.",
					s.Before.Identity),
			})
		}
	}

	summary := "Workflow run rejected: the Agent " + strings.Join(summaryParts, " and ")
	if restorationFailed {
		summary += ", and restoring it failed"
	}
	return &present.Report{Outcome: present.Failed, Summary: summary, Sections: sections}
}

func specViolationVerb(kind string) string {
	switch kind {
	case "renamed":
		return "renamed"
	case "deleted":
		return "deleted"
	default:
		return "changed"
	}
}

func describeDiff(dir string, diff fileDiff) string {
	var lines []string
	for _, p := range relPaths(dir, diff.Added) {
		lines = append(lines, "added: "+p)
	}
	for _, p := range relPaths(dir, diff.Changed) {
		lines = append(lines, "changed: "+p)
	}
	for _, p := range relPaths(dir, diff.Removed) {
		lines = append(lines, "removed: "+p)
	}
	return strings.Join(lines, "\n")
}

func relPaths(dir string, paths []string) []string {
	out := make([]string, len(paths))
	for i, p := range paths {
		if r, err := filepath.Rel(dir, p); err == nil {
			out[i] = r
		} else {
			out[i] = p
		}
	}
	return out
}

// approvalCheckFailedReport is used only when the guard cannot even take its own "after"
// snapshot (e.g. the approvals directory became unreadable) — trust in whatever outcome the run
// itself produced is not warranted either, so this is reported as a Failed run, not silently
// passed through.
func approvalCheckFailedReport(err error) *present.Report {
	return &present.Report{
		Outcome: present.Failed,
		Summary: "Workflow run rejected: approval evidence could not be verified after the run",
		Sections: []present.Section{
			{Label: "Reason", Body: err.Error()},
			{Label: "Result", Body: "Inspect .gnomon/approvals/ directly before trusting `gnomon status`."},
		},
	}
}

// specCheckFailedReport mirrors approvalCheckFailedReport for Rule A's own "after" check.
func specCheckFailedReport(identity string, err error) *present.Report {
	return &present.Report{
		Outcome: present.Failed,
		Summary: fmt.Sprintf("Workflow run rejected: %s could not be verified after the run", identity),
		Sections: []present.Section{
			{Label: "Reason", Body: err.Error()},
			{Label: "Result", Body: fmt.Sprintf("Inspect %s directly before trusting `gnomon status`.", identity)},
		},
	}
}

// --- Rule A: the governing Specification must not change during a run that requires it Approved ---

// specSnapshot is the governing Specification's state captured before a run requires it
// Approved — path, exact bytes, and approval fingerprint.
type specSnapshot struct {
	Identity string
	Path     string
	Content  []byte
}

func (s specSnapshot) fingerprint() string { return approval.Fingerprint(s.Content) }

// snapshotGoverningSpec captures identity's current file. A Specification that does not
// currently resolve returns (nil, nil) — eligibility already required it to exist and be
// Approved before the run started, so this is not expected, but is not an error here either.
func snapshotGoverningSpec(specsDir, identity string) (*specSnapshot, error) {
	ok, path, err := specs.Exists(specsDir, specs.Identity(identity))
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, nil
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return &specSnapshot{Identity: identity, Path: path, Content: content}, nil
}

// specViolation describes how the governing Specification changed, per checkGoverningSpec.
type specViolation struct {
	Kind    string // "changed", "renamed", or "deleted"
	NewPath string // the file now claiming the identity; "" for "deleted"
}

// checkGoverningSpec re-resolves before.Identity and reports whether anything actually changed —
// nil if not. A whitespace-only edit (same fingerprint) is deliberately not a violation.
func checkGoverningSpec(specsDir string, before *specSnapshot) (*specViolation, error) {
	ok, path, err := specs.Exists(specsDir, specs.Identity(before.Identity))
	if err != nil {
		return nil, err
	}
	if !ok {
		return &specViolation{Kind: "deleted"}, nil
	}
	if path != before.Path {
		return &specViolation{Kind: "renamed", NewPath: path}, nil
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if approval.Fingerprint(content) != before.fingerprint() {
		return &specViolation{Kind: "changed", NewPath: path}, nil
	}
	return nil, nil
}

// specFilesMatchingIdentity lists every file in specsDir currently matching identity (same rule as
// specs.Exists), sorted. More than one match is the ambiguous state Rule A's restore step resolves.
func specFilesMatchingIdentity(specsDir, identity string) ([]string, error) {
	entries, err := os.ReadDir(specsDir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	prefix := identity + "-"
	var matches []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if e.Name() == identity+".md" || strings.HasPrefix(e.Name(), prefix) {
			matches = append(matches, filepath.Join(specsDir, e.Name()))
		}
	}
	sort.Strings(matches)
	return matches, nil
}

// preserveAndRestoreGoverningSpec saves every file currently claiming before.Identity into
// result.TransientDir(root)/rejected/<runID>/ for Human inspection, restores before's own file, and
// removes every other file claiming the identity. Continues past a single failure.
func preserveAndRestoreGoverningSpec(root, runID string, before *specSnapshot) (preserved, failed []string, err error) {
	specsDir := filepath.Dir(before.Path)
	matches, mErr := specFilesMatchingIdentity(specsDir, before.Identity)
	if mErr != nil {
		return nil, nil, mErr
	}

	transientDir, tErr := result.TransientDir(root)
	if tErr != nil {
		return nil, nil, tErr
	}
	rejectedDir := filepath.Join(transientDir, "rejected", runID)
	if mkErr := os.MkdirAll(rejectedDir, 0o755); mkErr != nil {
		return nil, nil, mkErr
	}

	note := func(path string, e error) {
		if e == nil {
			return
		}
		failed = append(failed, path)
		if err == nil {
			err = e
		}
	}

	for _, path := range matches {
		data, rerr := os.ReadFile(path)
		if rerr != nil {
			note(path, rerr)
			continue
		}
		dest := filepath.Join(rejectedDir, filepath.Base(path))
		if werr := os.WriteFile(dest, data, 0o644); werr != nil {
			note(path, werr)
			continue
		}
		preserved = append(preserved, dest)
	}

	note(before.Path, os.WriteFile(before.Path, before.Content, 0o644))

	for _, path := range matches {
		if path == before.Path {
			continue
		}
		note(path, os.Remove(path))
	}

	return preserved, failed, err
}

// --- Rule B: report (never restore or reject) when a run removes approval from a Specification
// other than the one Rule A already governs ---

// approvedIdentities lists every Specification identity that currently derives Approved.
func approvedIdentities(l project.Layout) (map[string]bool, error) {
	ids, err := specs.List(l.SpecificationsDir())
	if err != nil {
		return nil, err
	}
	approved := map[string]bool{}
	for _, id := range ids {
		state, err := facts.Lifecycle(l, string(id))
		if err != nil {
			return nil, err
		}
		if state == approval.Approved {
			approved[string(id)] = true
		}
	}
	return approved, nil
}

// stillApproved reports whether identity currently derives Approved — false, never an error, if
// it no longer exists at all.
func stillApproved(l project.Layout, identity string) (bool, error) {
	ok, _, err := specs.Exists(l.SpecificationsDir(), specs.Identity(identity))
	if err != nil {
		return false, err
	}
	if !ok {
		return false, nil
	}
	state, err := facts.Lifecycle(l, identity)
	if err != nil {
		return false, err
	}
	return state == approval.Approved, nil
}

// lostApprovals returns, sorted, every identity in before that no longer derives Approved.
// excludeIdentity (Rule A's own governing Specification) is skipped so Rule B never repeats it.
func lostApprovals(l project.Layout, before map[string]bool, excludeIdentity string) ([]string, error) {
	var lost []string
	for id := range before {
		if id == excludeIdentity {
			continue
		}
		ok, err := stillApproved(l, id)
		if err != nil {
			return nil, err
		}
		if !ok {
			lost = append(lost, id)
		}
	}
	sort.Strings(lost)
	return lost, nil
}

// lostApprovalSection builds Rule B's warning Section — always additive to whatever report the
// run itself produced, never a reason to restore or reject anything.
func lostApprovalSection(lost []string) present.Section {
	var lines []string
	for _, id := range lost {
		lines = append(lines, fmt.Sprintf("%s (content changed or removed during this run)", id))
	}
	return present.Section{Label: "Other Approvals Lost", Body: strings.Join(lines, "\n")}
}
