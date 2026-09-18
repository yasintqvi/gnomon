package main

import (
	"strings"
	"testing"

	"gnomon/internal/approval"
	"gnomon/internal/orchestrate"
	"gnomon/internal/present"
)

func TestTruncate_ShortStringUnchanged(t *testing.T) {
	if got := truncate("Password Reset", 30); got != "Password Reset" {
		t.Fatalf("unexpected: %q", got)
	}
}

func TestTruncate_LongStringEllipsized(t *testing.T) {
	got := truncate("A very long Specification title that exceeds the column width", 20)
	if !strings.HasSuffix(got, "…") {
		t.Fatalf("expected an ellipsis marker, got %q", got)
	}
	if !strings.HasPrefix(got, "A very long Specifi") {
		t.Fatalf("expected the first 19 characters preserved, got %q", got)
	}
}

func TestCapitalize(t *testing.T) {
	cases := map[string]string{
		"define":  "Define",
		"approve": "Approve",
		"":        "",
	}
	for in, want := range cases {
		if got := capitalize(in); got != want {
			t.Errorf("capitalize(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestShort_FingerprintTruncatedTo12(t *testing.T) {
	if got := short("abcdefghijklmnopqrstuvwxyz"); got != "abcdefghijkl" {
		t.Fatalf("unexpected: %q", got)
	}
	if got := short("short"); got != "short" {
		t.Fatalf("expected a fingerprint shorter than 12 to pass through unchanged, got %q", got)
	}
}

// --- gnomon spec: registration, args shape, non-interactive behavior ---

func TestSpecCmd_AcceptsZeroOrOneArg(t *testing.T) {
	if err := specCmd.Args(specCmd, nil); err != nil {
		t.Fatalf("expected zero args (the browser) to be accepted, got: %v", err)
	}
	if err := specCmd.Args(specCmd, []string{"SPEC-001"}); err != nil {
		t.Fatalf("expected one arg (a direct workspace) to be accepted, got: %v", err)
	}
	if err := specCmd.Args(specCmd, []string{"SPEC-001", "extra"}); err == nil {
		t.Fatalf("expected more than one arg to be refused")
	}
}

func TestSpecCmd_HasNoAgentFlag(t *testing.T) {
	// gnomon spec itself never resolves an Agent directly — its contextual actions (runSpecAction)
	// always resolve through the persisted default or first-use chooser, never a per-invocation
	// flag, since there is no per-invocation CLI flag to read inside an interactive session.
	if specCmd.Flags().Lookup("agent") != nil {
		t.Fatalf("did not expect gnomon spec itself to register --agent")
	}
}

func TestSpecCmd_OnlyCreateRemainsAsASubcommand(t *testing.T) {
	// create has no Workflow Contract for `run` to invoke, so it keeps its own non-interactive
	// entry point. discover/define are Contract-driven Agent workflows fully representable by
	// `gnomon run <identity> [target]` and must not also exist as dedicated subcommands.
	names := map[string]bool{}
	for _, c := range specCmd.Commands() {
		names[c.Name()] = true
	}
	if !names["create"] {
		t.Fatalf("expected gnomon spec create to remain registered for non-interactive use")
	}
	for _, unwanted := range []string{"discover", "define"} {
		if names[unwanted] {
			t.Fatalf("did not expect gnomon spec %s to remain registered; use gnomon run instead", unwanted)
		}
	}
}

// --- renderSpecDetail / lifecycleStyle: SPEC workspace presentation ---

func draftDetail() *orchestrate.SpecDetail {
	return &orchestrate.SpecDetail{
		ID:        "SPEC-001",
		Title:     "Password Reset",
		Lifecycle: approval.Draft,
		Actions: []orchestrate.SpecAction{
			{Verb: "define", Command: "gnomon run specification-definition SPEC-001", Available: true},
			{Verb: "approve", Command: "gnomon approve SPEC-001", Available: true},
			{Verb: "implement", Available: false, Reason: "SPEC-001 is not Approved (currently Draft)"},
		},
	}
}

func approvedDetailWithRevisions() *orchestrate.SpecDetail {
	return &orchestrate.SpecDetail{
		ID:        "SPEC-001",
		Title:     "Password Reset",
		Lifecycle: approval.Approved,
		ActiveGrant: &approval.Revision{
			GrantID:     "g2",
			Fingerprint: "abc123",
			Approver:    "Jane Doe",
			ApprovedAt:  "2026-01-01T00:00:00Z",
		},
		Revisions: []approval.Revision{
			{GrantID: "g1", Fingerprint: "old111111111", Approver: "Jane Doe", ApprovedAt: "2025-12-01T00:00:00Z"},
			{GrantID: "g1r", Fingerprint: "rev222222222", Approver: "Jane Doe", ApprovedAt: "2025-12-15T00:00:00Z", Revoked: true},
			{GrantID: "g2", Fingerprint: "abc123", Approver: "Jane Doe", ApprovedAt: "2026-01-01T00:00:00Z"},
		},
		Actions: []orchestrate.SpecAction{
			{Verb: "implement", Available: true},
		},
	}
}

func TestRenderSpecDetail_Plain_ShowsIdentityLifecycleAndUnavailableReasons(t *testing.T) {
	out := renderSpecDetail(draftDetail(), present.NewPalette(false))
	if !strings.Contains(out, "SPEC-001 — Password Reset") {
		t.Fatalf("expected the identity/title heading, got: %s", out)
	}
	if !strings.Contains(out, "Draft") {
		t.Fatalf("expected the current lifecycle state shown, got: %s", out)
	}
	if !strings.Contains(out, "implement (SPEC-001 is not Approved (currently Draft))") {
		t.Fatalf("expected the unavailable action's own reason shown, got: %s", out)
	}
	if strings.Contains(out, "\033[") {
		t.Fatalf("expected no ANSI when color is disabled, got: %q", out)
	}
}

func TestRenderSpecDetail_Color_LifecycleAndRevisionStatusDistinguished(t *testing.T) {
	p := present.NewPalette(true)
	out := renderSpecDetail(approvedDetailWithRevisions(), p)
	if !strings.Contains(out, p.Good("Approved")) {
		t.Fatalf("expected Approved styled as the positive/complete state, got: %s", out)
	}
	if !strings.Contains(out, p.Good("current")) {
		t.Fatalf("expected the active revision marked 'current' in the same positive style, got: %s", out)
	}
	if !strings.Contains(out, p.Bad("revoked")) {
		t.Fatalf("expected a revoked revision marked in the failure style, got: %s", out)
	}
	if stripAnsiForTest(out) != renderSpecDetail(approvedDetailWithRevisions(), present.NewPalette(false)) {
		t.Fatalf("expected stripping ANSI from the colored render to reproduce the plain render exactly")
	}
}

func TestRenderSpecDetail_Draft_LifecycleStyledAsWarnNotFailure(t *testing.T) {
	p := present.NewPalette(true)
	out := renderSpecDetail(draftDetail(), p)
	if !strings.Contains(out, p.Warn("Draft")) {
		t.Fatalf("expected Draft styled as 'needs attention', not a failure color, got: %s", out)
	}
	if strings.Contains(out, p.Bad("Draft")) {
		t.Fatalf("did not expect Draft styled as a failure: %s", out)
	}
}

func TestRenderSpecDetail_UnreadableContract_ShownAsWarning(t *testing.T) {
	d := draftDetail()
	d.Unreadable = []string{"testing.md"}
	out := renderSpecDetail(d, present.NewPalette(false))
	if !strings.Contains(out, "Warning: testing.md could not be loaded") {
		t.Fatalf("expected the unreadable Contract warning, got: %s", out)
	}
}

func TestLifecycleStyle_ApprovedGood_DraftWarn(t *testing.T) {
	p := present.NewPalette(true)
	if got := lifecycleStyle(p, approval.Approved); got != p.Good("Approved") {
		t.Fatalf("expected Approved to use Good styling, got %q", got)
	}
	if got := lifecycleStyle(p, approval.Draft); got != p.Warn("Draft") {
		t.Fatalf("expected Draft to use Warn styling, got %q", got)
	}
}

func stripAnsiForTest(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == '\033' {
			for i < len(s) && s[i] != 'm' {
				i++
			}
			continue
		}
		b.WriteByte(s[i])
	}
	return b.String()
}

func TestSpecCmd_NonInteractive_RefusesRatherThanHangs(t *testing.T) {
	// go test's own stdin is never a real terminal, so RunE's own stdinIsInteractive() check must
	// refuse immediately — before ever touching a project directory — rather than entering
	// runSpecBrowser/runSpecWorkspace, which would block forever reading keystrokes that will
	// never arrive.
	if err := specCmd.RunE(specCmd, nil); err == nil {
		t.Fatalf("expected a non-interactive invocation to be refused")
	}
}
