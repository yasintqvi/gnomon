package orchestrate

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"gnomon/internal/approval"
	"gnomon/internal/project"
	"gnomon/internal/result"
	"gnomon/internal/specs"
)

// definitionIdentity is the workflow that receives the decisions recorded in other Specifications.
const definitionIdentity = "specification-definition"

// recordedSectionPattern matches the headings whose sections hold decisions and deferred notes, in
// either template: "## Decisions", "## Business Rules", "### Out of Scope", "## Related Decisions".
var recordedSectionPattern = regexp.MustCompile(`(?i)^(#{2,4})\s+(decisions|business rules|out of scope|related decisions)\s*$`)

// deferredLinePattern matches a "Deferred:" note written anywhere in a Specification.
var deferredLinePattern = regexp.MustCompile(`(?i)^\s*[-*]?\s*(\*\*)?deferred(\*\*)?\s*:`)

var headingPattern = regexp.MustCompile(`^(#{1,6})\s`)

// recordedSource is one other Specification's agreed (or proposed) decision text.
type recordedSource struct {
	id, title, path string
	status          string // how far this text is agreed
	approvedAt      string // "" for a Draft that was never approved
	sections        []string
}

// recordedDecisionsStats measures what was handed over.
type recordedDecisionsStats struct {
	Specifications int
	Bytes          int
}

// buildRecordedDecisions extracts, from every Specification except exclude, the sections that hold
// decisions and deferred notes, from the latest approved revision where there is one, labelled
// with source and approval state and ordered from earliest to latest approval (Drafts last).
// It returns "" when no other Specification records anything.
func buildRecordedDecisions(l project.Layout, exclude string) (string, recordedDecisionsStats, error) {
	ids, err := specs.List(l.SpecificationsDir())
	if err != nil {
		return "", recordedDecisionsStats{}, err
	}
	var sources []recordedSource
	for _, id := range ids {
		if string(id) == exclude {
			continue
		}
		ok, path, err := specs.Exists(l.SpecificationsDir(), id)
		if err != nil || !ok {
			continue
		}
		current, err := os.ReadFile(path)
		if err != nil {
			return "", recordedDecisionsStats{}, err
		}
		src := recordedSource{id: string(id), title: specs.Title(id, current), path: relSlash(l.Root, path)}
		text := string(current)

		latest := latestApprovedRevision(l.ApprovalsDir(), string(id))
		switch {
		case latest == nil:
			src.status = "Draft, never approved: proposals, not agreed decisions"
		case latest.Fingerprint == approval.Fingerprint(current):
			src.status = "Approved " + latest.ApprovedAt
			src.approvedAt = latest.ApprovedAt
		default:
			src.status = "Approved " + latest.ApprovedAt + "; shown as approved — the file has changed since, and those edits are not agreed"
			src.approvedAt = latest.ApprovedAt
			text = latest.Content
		}

		src.sections = extractRecordedSections(text)
		if len(src.sections) > 0 {
			sources = append(sources, src)
		}
	}
	if len(sources) == 0 {
		return "", recordedDecisionsStats{}, nil
	}

	sort.SliceStable(sources, func(i, j int) bool {
		a, b := sources[i].approvedAt, sources[j].approvedAt
		if (a == "") != (b == "") {
			return a != "" // approved before drafts
		}
		return a < b
	})

	var out strings.Builder
	out.WriteString("# Decisions and deferred notes recorded in other Specifications\n\n")
	out.WriteString("Extracted by Gnomon for this run from every other Specification, oldest approval first. Read all of it before asking the user anything.\n\n")
	out.WriteString("- Statements from Approved Specifications are decisions the user has agreed to. Where two disagree on the same point, the later-approved one supersedes the earlier.\n")
	out.WriteString("- An Out of Scope or deferred note (\"for later\", \"not decided here\") is something the user already said about future work. If this request is that work, the note applies unless a later-approved statement replaced it: record it as the user's earlier decision, citing its source, instead of deciding it again. Ask only if it is unclear whether or how it applies.\n")
	out.WriteString("- A note about something this request does not touch does not apply to it.\n")
	out.WriteString("- Draft statements are proposals, not agreed.\n")
	for _, s := range sources {
		fmt.Fprintf(&out, "\n## %s — %s\n\nSource: %s — %s\n", s.id, s.title, s.path, s.status)
		for _, sec := range s.sections {
			out.WriteString("\n" + sec + "\n")
		}
	}
	text := out.String()
	return text, recordedDecisionsStats{Specifications: len(sources), Bytes: len(text)}, nil
}

// latestApprovedRevision is the newest non-revoked grant whose content snapshot exists, or nil.
func latestApprovedRevision(approvalsDir, id string) *approval.Revision {
	revs, err := approval.Revisions(approvalsDir, id)
	if err != nil {
		return nil
	}
	for i := len(revs) - 1; i >= 0; i-- {
		if !revs[i].Revoked && revs[i].ContentKnown {
			return &revs[i]
		}
	}
	return nil
}

// extractRecordedSections returns each decision-bearing section (heading included, down to the
// next heading of the same or a higher level) and each "Deferred:" line found outside them.
func extractRecordedSections(text string) []string {
	lines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	var sections []string
	inSection := false
	sectionLevel := 0
	var cur []string
	flush := func() {
		if body := strings.TrimSpace(strings.Join(cur, "\n")); body != "" && strings.Contains(body, "\n") {
			sections = append(sections, body)
		}
		cur = nil
	}
	for _, line := range lines {
		if m := headingPattern.FindStringSubmatch(line); m != nil {
			level := len(m[1])
			if inSection && level <= sectionLevel {
				flush()
				inSection = false
			}
			if rm := recordedSectionPattern.FindStringSubmatch(strings.TrimSpace(line)); rm != nil && !inSection {
				inSection = true
				sectionLevel = len(rm[1])
				cur = []string{"### " + strings.TrimSpace(strings.TrimLeft(strings.TrimSpace(line), "#"))}
				continue
			}
		}
		if inSection {
			cur = append(cur, line)
		} else if deferredLinePattern.MatchString(line) {
			sections = append(sections, "### Deferred note\n"+strings.TrimSpace(line))
		}
	}
	if inSection {
		flush()
	}
	return sections
}

// writeRecordedDecisions writes the extract for one Define run into the transient directory and
// returns its project-relative path, or "" when there is nothing to hand over.
func writeRecordedDecisions(root string, l project.Layout, runID, exclude string) (string, recordedDecisionsStats, error) {
	text, stats, err := buildRecordedDecisions(l, exclude)
	if err != nil || text == "" {
		return "", stats, err
	}
	dir, err := result.TransientDir(root)
	if err != nil {
		return "", stats, err
	}
	path := filepath.Join(dir, runID+".recorded-decisions.md")
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		return "", stats, err
	}
	return relSlash(root, path), stats, nil
}
