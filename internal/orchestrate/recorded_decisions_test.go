package orchestrate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gnomon/internal/adapter"
	"gnomon/internal/project"
)

const reservationsSpec = `# SPEC-002 — Reserve a tool

## Goal

Members can reserve a tool that is out.

## Decisions

- Can a member reserve an available item? — No, 409: borrow it instead.

## Acceptance Criteria

1. Reserving an available item is 409.

## Out of Scope

- Late fees. Noted by the user for that later work, not decided here: each member's first late return in a calendar year will be free.
`

const pickupSpec = `# SPEC-003 — Self-service pickup

## Goal

Members pick up reserved tools at the kiosk.

## Decisions

- Does the 3-tool limit apply to a pickup? — No.

## Acceptance Criteria

1. A pickup with three tools out is 201.

Deferred: whether the kiosk prints a receipt is left for later.

## Out of Scope

- Changes to POST /loans.
`

func specPath(t *testing.T, root, id string) string {
	t.Helper()
	l, _ := project.Locate(root)
	matches, _ := filepath.Glob(filepath.Join(l.SpecificationsDir(), id+"-*.md"))
	if len(matches) != 1 {
		t.Fatalf("expected one file for %s, got %v", id, matches)
	}
	return matches[0]
}

// projectWithRecordedDecisions: SPEC-001 approved (template), SPEC-002 approved with a deferred
// fee note, SPEC-003 approved later, then edited without re-approval.
func projectWithRecordedDecisions(t *testing.T) (string, project.Layout) {
	t.Helper()
	root := setupApprovedSpec(t)
	for _, c := range []struct{ title, body string }{{"Reserve a tool", reservationsSpec}, {"Self-service pickup", pickupSpec}} {
		res, err := specCreateDefined(root, c.title)
		if err != nil {
			t.Fatal(err)
		}
		_ = res
	}
	for _, c := range []struct{ id, body string }{{"SPEC-002", reservationsSpec}, {"SPEC-003", pickupSpec}} {
		if err := os.WriteFile(specPath(t, root, c.id), []byte(c.body), 0o644); err != nil {
			t.Fatal(err)
		}
		if _, err := Approve(root, c.id, nil, nil); err != nil {
			t.Fatal(err)
		}
	}
	edited := strings.Replace(pickupSpec, "— No.", "— Yes, it applies too.", 1)
	if err := os.WriteFile(specPath(t, root, "SPEC-003"), []byte(edited), 0o644); err != nil {
		t.Fatal(err)
	}
	l, _ := project.Locate(root)
	return root, l
}

func TestRecordedDecisions_ExtractsDecisionsAndDeferredNotesWithSourceAndState(t *testing.T) {
	_, l := projectWithRecordedDecisions(t)
	text, stats, err := buildRecordedDecisions(l, "SPEC-004")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"## SPEC-002 — Reserve a tool",
		"Source: .gnomon/specifications/SPEC-002-",
		"### Decisions\n\n- Can a member reserve an available item? — No, 409",
		"each member's first late return in a calendar year will be free",
		"### Deferred note\nDeferred: whether the kiosk prints a receipt is left for later.",
		"the later-approved one supersedes the earlier",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("expected %q in:\n%s", want, text)
		}
	}
	if strings.Contains(text, "## Acceptance Criteria") || strings.Contains(text, "Reserving an available item is 409") || strings.Contains(text, "## Goal") {
		t.Fatalf("only decision-bearing sections may be handed over, not whole Specifications:\n%s", text)
	}
	if stats.Bytes != len(text) || stats.Specifications < 2 {
		t.Fatalf("unexpected stats %+v", stats)
	}
}

// The agreed text is what counts: a Specification edited after approval is shown as approved,
// with its unapproved edits left out and said so.
func TestRecordedDecisions_ChangedSinceApproval_ShowsApprovedText(t *testing.T) {
	_, l := projectWithRecordedDecisions(t)
	text, _, _ := buildRecordedDecisions(l, "")
	section := text[strings.Index(text, "## SPEC-003"):]
	if !strings.Contains(section, "the file has changed since, and those edits are not agreed") {
		t.Fatalf("expected SPEC-003 marked as changed since approval:\n%s", section)
	}
	if !strings.Contains(section, "Does the 3-tool limit apply to a pickup? — No.") || strings.Contains(section, "Yes, it applies too") {
		t.Fatalf("expected the approved answer, not the unapproved edit:\n%s", section)
	}
}

// Oldest approval first, so a later statement visibly comes after the one it may supersede; a
// never-approved Draft comes last, marked as not agreed; the Specification being defined is left out.
func TestRecordedDecisions_OrderAndDraftsAndExclusion(t *testing.T) {
	root, l := projectWithRecordedDecisions(t)
	if _, err := specCreateDefined(root, "Late fees"); err != nil { // SPEC-004, Draft
		t.Fatal(err)
	}
	draft := "# SPEC-004 — Late fees\n\n## Decisions\n\n- Is there a cap? — 10 euros.\n"
	if err := os.WriteFile(specPath(t, root, "SPEC-004"), []byte(draft), 0o644); err != nil {
		t.Fatal(err)
	}

	text, _, _ := buildRecordedDecisions(l, "")
	i2, i3, i4 := strings.Index(text, "## SPEC-002"), strings.Index(text, "## SPEC-003"), strings.Index(text, "## SPEC-004")
	if !(i2 >= 0 && i2 < i3 && i3 < i4) {
		t.Fatalf("expected approval order SPEC-002, SPEC-003, then the Draft SPEC-004:\n%s", text)
	}
	if !strings.Contains(text[i4:], "Draft, never approved: proposals, not agreed decisions") {
		t.Fatalf("expected the Draft marked as not agreed:\n%s", text[i4:])
	}

	text, _, _ = buildRecordedDecisions(l, "SPEC-004")
	if strings.Contains(text, "## SPEC-004") {
		t.Fatal("the Specification being defined must not be handed over as another one")
	}
}

// A Define run gets the extract as a file to read first; other workflows do not.
func TestDefine_ReceivesRecordedDecisionsFile(t *testing.T) {
	root, _ := projectWithRecordedDecisions(t)
	if _, err := specCreateDefined(root, "Late fees"); err != nil { // SPEC-004
		t.Fatal(err)
	}
	l, wf, workflowPath, err := prepareWorkflow(root, "specification-definition.md")
	if err != nil {
		t.Fatal(err)
	}
	ad := &adapter.FakeAdapter{}
	_, _, _ = execute(runRequest{root: root, l: l, wf: wf, workflowPath: workflowPath, kind: targetSpec, target: "SPEC-004", adapter: ad})
	if len(ad.Ctx.ReadFirst) != 1 || !strings.HasSuffix(ad.Ctx.ReadFirst[0].Path, ".recorded-decisions.md") {
		t.Fatalf("expected one recorded-decisions file to read first, got %+v", ad.Ctx.ReadFirst)
	}
	data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(ad.Ctx.ReadFirst[0].Path)))
	if err != nil || !strings.Contains(string(data), "first late return in a calendar year will be free") {
		t.Fatalf("expected the deferred fee note in the file: %v\n%s", err, data)
	}

	impl := &adapter.FakeAdapter{}
	lImpl, wfImpl := mustPrepareImplementation(t, root, "SPEC-002")
	_, _ = implementWithAdapter(root, lImpl, wfImpl, "SPEC-002", impl)
	if len(impl.Ctx.ReadFirst) != 0 {
		t.Fatalf("only Define receives the recorded decisions, got %+v", impl.Ctx.ReadFirst)
	}
}

func TestRecordedDecisions_NothingRecorded_NoFile(t *testing.T) {
	root := setupApprovedSpec(t) // only SPEC-001, from the template
	l, _ := project.Locate(root)
	path, _, err := writeRecordedDecisions(root, l, "r1", "SPEC-001")
	if err != nil || path != "" {
		t.Fatalf("with no other Specification there is nothing to hand over, got %q %v", path, err)
	}
}
