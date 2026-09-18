package present

import (
	"errors"
	"strings"
	"testing"
)

func TestRender_PlainByDefault_NoAnsi(t *testing.T) {
	r := &Report{Outcome: Success, Summary: "done", Sections: []Section{{Label: "Delivered", Body: "x"}}, Next: "gnomon status"}
	out := Render(r, false)
	if strings.Contains(out, "\033[") {
		t.Fatalf("Render (2-arg, plain) must never emit ANSI escape codes: %q", out)
	}
}

func TestRenderWithOptions_ColorFalse_MatchesRenderExactly(t *testing.T) {
	r := &Report{Outcome: Blocked, Summary: "blocked", Target: "SPEC-001", Sections: []Section{{Label: "Unresolved", Body: "reason"}}, Next: "do X"}
	plain := Render(r, true)
	withOpts := RenderWithOptions(r, RenderOptions{Verbose: true, Color: false})
	if plain != withOpts {
		t.Fatalf("expected RenderWithOptions(Color:false) to byte-for-byte match Render:\nRender:  %q\nOptions: %q", plain, withOpts)
	}
}

func TestRenderWithOptions_ColorTrue_WrapsOutcomeAndHeadings(t *testing.T) {
	cases := []struct {
		name    string
		outcome Outcome
		code    string
	}{
		{"success", Success, ansiGreen},
		{"blocked", Blocked, ansiYellow},
		{"cancelled", Cancelled, ansiYellow},
		{"failed", Failed, ansiRed},
	}
	for _, c := range cases {
		r := &Report{Outcome: c.outcome, Summary: "summary text", Sections: []Section{{Label: "Heading", Body: "body"}}, Next: "next text"}
		out := RenderWithOptions(r, RenderOptions{Color: true})
		if !strings.Contains(out, c.code+"✓") && !strings.Contains(out, c.code+"!") && !strings.Contains(out, c.code+"✗") {
			t.Errorf("%s: expected the outcome symbol colored with %q, got: %q", c.name, c.code, out)
		}
		if !strings.Contains(out, ansiCyan+"Heading"+ansiReset) {
			t.Errorf("%s: expected the section heading colored cyan, got: %q", c.name, out)
		}
		if !strings.Contains(out, ansiReset) {
			t.Errorf("%s: expected color codes to be reset, got: %q", c.name, out)
		}
	}
}

func TestRenderWithOptions_ColorTrue_ContentUnchanged(t *testing.T) {
	// Color must be purely cosmetic: every word of content present in the plain render must still
	// be present, unmodified, in the colored render — only ANSI wrapping differs.
	r := &Report{Outcome: Success, Summary: "Implementation complete", Target: "SPEC-001", Sections: []Section{{Label: "Delivered", Body: "a trivial change"}}, Next: "gnomon test SPEC-001"}
	plain := Render(r, false)
	colored := RenderWithOptions(r, RenderOptions{Color: true})
	stripped := stripAnsi(colored)
	if stripped != plain {
		t.Fatalf("expected stripping ANSI from the colored render to reproduce the plain render exactly:\nplain:   %q\nstripped: %q", plain, stripped)
	}
}

func TestRenderError_Plain_NoAnsi(t *testing.T) {
	out := RenderError(errors.New("boom"))
	if strings.Contains(out, "\033[") {
		t.Fatalf("RenderError must never emit ANSI: %q", out)
	}
	if out != "✗ boom\n" {
		t.Fatalf("unexpected error rendering: %q", out)
	}
}

func TestRenderErrorColor_FalseMatchesPlain(t *testing.T) {
	err := errors.New("boom")
	if RenderErrorColor(err, false) != RenderError(err) {
		t.Fatalf("expected RenderErrorColor(err, false) to match RenderError")
	}
}

func TestRenderErrorColor_TrueWrapsRed(t *testing.T) {
	out := RenderErrorColor(errors.New("boom"), true)
	if !strings.Contains(out, ansiRed+"✗"+ansiReset) {
		t.Fatalf("expected the error symbol colored red, got: %q", out)
	}
	if stripAnsi(out) != RenderError(errors.New("boom")) {
		t.Fatalf("expected stripped colored error output to match the plain rendering")
	}
}

// TestSemanticDistinctions_SurviveColor proves the exact distinctions Step 12 established —
// Blocked vs Failed, and their symbols — remain intact regardless of color, since a caller (or a
// test) can always fall back to the Outcome enum itself, never string-sniffing rendered text.
func TestSemanticDistinctions_SurviveColor(t *testing.T) {
	blocked := &Report{Outcome: Blocked, Summary: "Verification FAIL"}
	failed := &Report{Outcome: Failed, Summary: "Agent process crashed"}
	if blocked.Outcome == failed.Outcome {
		t.Fatalf("Blocked and Failed must remain distinct Outcome values")
	}
	for _, opts := range []RenderOptions{{Color: false}, {Color: true}} {
		b := RenderWithOptions(blocked, opts)
		f := RenderWithOptions(failed, opts)
		if !strings.Contains(b, "!") {
			t.Errorf("Color:%v: expected Blocked's '!' symbol to survive: %q", opts.Color, b)
		}
		if !strings.Contains(f, "✗") {
			t.Errorf("Color:%v: expected Failed's '✗' symbol to survive: %q", opts.Color, f)
		}
	}
}

// TestRenderWithOptions_ColorTrue_BacktickSpansBolded proves `backtick-delimited` values within a
// Section body and Next text are emphasized when color is on, with the backticks themselves and
// every other character left untouched — the one device this package uses for distinguishing
// commands/paths/identifiers, short of real syntax highlighting.
func TestRenderWithOptions_ColorTrue_BacktickSpansBolded(t *testing.T) {
	r := &Report{
		Outcome:  Success,
		Summary:  "done",
		Sections: []Section{{Label: "Working Tree", Body: "Uncommitted changes — `gnomon finalize` may help."}},
		Next:     "Run `gnomon validate` for details.",
	}
	out := RenderWithOptions(r, RenderOptions{Color: true})
	if !strings.Contains(out, ansiBold+"`gnomon finalize`"+ansiReset) {
		t.Fatalf("expected the backtick span in the section body bolded, got: %q", out)
	}
	if !strings.Contains(out, ansiBold+"`gnomon validate`"+ansiReset) {
		t.Fatalf("expected the backtick span in Next bolded, got: %q", out)
	}
	if stripAnsi(out) != Render(r, false) {
		t.Fatalf("expected stripping ANSI to reproduce the plain render exactly")
	}
}

func TestRenderWithOptions_ColorFalse_BackticksLeftAsIs(t *testing.T) {
	r := &Report{Outcome: Success, Summary: "done", Sections: []Section{{Label: "X", Body: "run `gnomon status`"}}}
	out := Render(r, false)
	if !strings.Contains(out, "`gnomon status`") {
		t.Fatalf("expected backticks preserved verbatim in plain output, got: %q", out)
	}
	if strings.Contains(out, "\033[") {
		t.Fatalf("did not expect any ANSI in plain output, got: %q", out)
	}
}

// --- Palette ---

func TestPalette_ColorDisabled_AllMethodsPassThrough(t *testing.T) {
	p := NewPalette(false)
	for name, got := range map[string]string{
		"Heading": p.Heading("x"),
		"Good":    p.Good("x"),
		"Warn":    p.Warn("x"),
		"Bad":     p.Bad("x"),
		"Code":    p.Code("x"),
		"Muted":   p.Muted("x"),
	} {
		if got != "x" {
			t.Errorf("%s: expected color-disabled Palette to pass text through unchanged, got %q", name, got)
		}
	}
}

func TestPalette_ColorEnabled_DistinctCodesPerMethod_AlwaysReset(t *testing.T) {
	p := NewPalette(true)
	cases := map[string]struct{ got, want string }{
		"Heading": {p.Heading("x"), ansiBold + ansiCyan + "x" + ansiReset},
		"Good":    {p.Good("x"), ansiGreen + "x" + ansiReset},
		"Warn":    {p.Warn("x"), ansiYellow + "x" + ansiReset},
		"Bad":     {p.Bad("x"), ansiRed + "x" + ansiReset},
		"Code":    {p.Code("x"), ansiBold + "x" + ansiReset},
		"Muted":   {p.Muted("x"), ansiDim + "x" + ansiReset},
	}
	for name, c := range cases {
		if c.got != c.want {
			t.Errorf("%s: got %q, want %q", name, c.got, c.want)
		}
	}
}

func TestPalette_EmptyString_NeverWrapped(t *testing.T) {
	p := NewPalette(true)
	if got := p.Good(""); got != "" {
		t.Fatalf("expected an empty string to remain empty even with color enabled, got %q", got)
	}
}

func stripAnsi(s string) string {
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
