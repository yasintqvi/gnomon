// Package present is Gnomon's small, reusable Human-facing presentation layer. It defines a
// renderer-agnostic Report shape any command can return and any UI (this package's own CLI
// Render, or a future GUI) can consume — the data shape itself carries no terminal styling.
package present

import "strings"

// Outcome is the small, closed visual vocabulary every command result maps onto.
type Outcome int

const (
	// Success — the operation completed as intended.
	Success Outcome = iota
	// Blocked — a workflow, or a deterministic gate, correctly reports more is needed before
	// this can proceed. Never a system failure.
	Blocked
	// Cancelled — the Human (or Gnomon, on their behalf) ended the run intentionally before
	// completion.
	Cancelled
	// Failed — something did not work as intended: a process crash, a missing/invalid result,
	// or any other genuine error.
	Failed
)

func (o Outcome) symbol() string {
	switch o {
	case Success:
		return "✓"
	case Blocked, Cancelled:
		return "!"
	case Failed:
		return "✗"
	default:
		return "?"
	}
}

// ansi is the one place this package's restrained semantic color vocabulary is defined: green
// for success, yellow for a Blocked/Cancelled outcome (or anything else that needs attention
// without being a failure), red for an actual failure or unclassified error, cyan for
// headings/structure/next-action text, bold for compact machine-oriented values (commands,
// identifiers, paths — never full syntax highlighting, just one consistent emphasis), dim for
// secondary/contextual text that should visually recede rather than compete with the content
// around it. Applying color is purely a rendering-time decision (RenderWithOptions/
// RenderErrorColor/Palette, all gated on a caller-supplied bool) — it never changes Outcome,
// Summary, Section, or any other field a caller reads.
const (
	ansiReset  = "\033[0m"
	ansiBold   = "\033[1m"
	ansiDim    = "\033[2m"
	ansiGreen  = "\033[32m"
	ansiYellow = "\033[33m"
	ansiRed    = "\033[31m"
	ansiCyan   = "\033[36m"
)

func (o Outcome) color() string {
	switch o {
	case Success:
		return ansiGreen
	case Blocked, Cancelled:
		return ansiYellow
	case Failed:
		return ansiRed
	default:
		return ""
	}
}

func colorize(color, s string) string {
	if color == "" || s == "" {
		return s
	}
	return color + s + ansiReset
}

// Palette is this package's small, closed semantic color vocabulary, exposed for the rare caller
// that builds its own Human-facing text outside a Report (the interactive Specification
// workspace) — so every surface draws from exactly one restrained vocabulary, decided once at the
// CLI edge, rather than each scattering its own ANSI escape codes. Every method is a no-op when
// the Palette was constructed with color disabled, and always returns s unstyled when s is empty.
type Palette struct {
	color bool
}

// NewPalette returns a Palette that applies color only when color is true — the same bool every
// other Color-gated entry point in this package already takes.
func NewPalette(color bool) Palette {
	return Palette{color: color}
}

// Heading styles a major label — a section heading, a screen title — the same treatment Report
// section labels receive.
func (p Palette) Heading(s string) string { return p.wrap(ansiBold+ansiCyan, s) }

// Good styles a positive, complete, or currently-active state (e.g. "Approved", a passing
// result) — the same green Success itself uses.
func (p Palette) Good(s string) string { return p.wrap(ansiGreen, s) }

// Warn styles a state that needs attention without being a failure (e.g. "Draft", a knowledge
// gap, a non-fatal warning) — the same yellow Blocked/Cancelled itself use.
func (p Palette) Warn(s string) string { return p.wrap(ansiYellow, s) }

// Bad styles an actual failure — the same red Failed itself uses.
func (p Palette) Bad(s string) string { return p.wrap(ansiRed, s) }

// Code styles a compact, machine-oriented value — a command, an identifier, a path — with a
// single consistent emphasis (bold, not a color), so it reads as "a value to copy or type"
// without competing with this package's semantic-outcome colors. Never syntax highlighting: every
// such value gets the same treatment regardless of what kind of value it is.
func (p Palette) Code(s string) string { return p.wrap(ansiBold, s) }

// Muted styles secondary or contextual text — metadata, an unavailable action's reason — so it
// visually recedes behind the primary content around it.
func (p Palette) Muted(s string) string { return p.wrap(ansiDim, s) }

func (p Palette) wrap(color, s string) string {
	if !p.color {
		return s
	}
	return colorize(color, s)
}

// styleCodeSpans bolds every `backtick-delimited` span in s and leaves everything else, including
// the backticks themselves, untouched — reusing the Markdown-style inline-code convention already
// used throughout Gnomon's own guidance text (and naturally present in any Agent narrative that
// happens to use it) as the one deliberate, unambiguous marker for "this is a command, path, or
// identifier", rather than inspecting content or attempting real syntax highlighting. A no-op
// when color is "".
func styleCodeSpans(s, color string) string {
	if color == "" || !strings.Contains(s, "`") {
		return s
	}
	var b strings.Builder
	for {
		start := strings.IndexByte(s, '`')
		if start == -1 {
			b.WriteString(s)
			break
		}
		end := strings.IndexByte(s[start+1:], '`')
		if end == -1 {
			b.WriteString(s)
			break
		}
		end += start + 1
		b.WriteString(s[:start])
		b.WriteString(color)
		b.WriteString(s[start : end+1])
		b.WriteString(ansiReset)
		s = s[end+1:]
	}
	return b.String()
}

// Section is one labeled, omittable block of content (for example "Delivered" or "Unresolved").
// A Section with an empty Body is skipped entirely by Render — callers should supply a natural
// equivalent (e.g. "No unresolved issues.") only where that is itself meaningful, never a
// placeholder for its own sake.
type Section struct {
	Label string
	Body  string
}

// Report is a structured, renderer-agnostic description of what a command produced.
type Report struct {
	Outcome  Outcome
	Summary  string    // one line, human-readable, no Go/process terminology
	Target   string    // e.g. "SPEC-001"; "" if not applicable
	Sections []Section // ordered; empty-Body sections are skipped
	Next     string    // a suggested next action; "" if none
	Detail   []string  // low-level diagnostic lines — never shown unless verbose
}

// AddSection appends a section only if body is non-empty, keeping omission the caller's default
// rather than something every call site has to remember to check itself.
func (r *Report) AddSection(label, body string) {
	if body == "" {
		return
	}
	r.Sections = append(r.Sections, Section{Label: label, Body: body})
}

// RenderOptions controls how RenderWithOptions presents a Report. Both fields default to their
// zero value ("plain, non-verbose") so a caller that only needs one behaves exactly as Render
// itself already did.
type RenderOptions struct {
	// Verbose includes low-level diagnostics (exit codes, run IDs, paths) — the same switch
	// Render's own verbose parameter has always been.
	Verbose bool
	// Color applies this package's small semantic ANSI vocabulary. It is purely presentational:
	// Outcome/Summary/Section/Next content is byte-identical whether or not this is set: only
	// which escape codes surround it changes. Callers decide this at the CLI edge (a real,
	// interactive TTY, with NO_COLOR unset) — present itself never inspects the environment.
	Color bool
}

// Render produces the CLI's plain-text presentation of r — no color, ever. Every existing caller
// (including every test asserting on exact rendered text) keeps this exact, deterministic,
// ANSI-free behavior; it is a thin wrapper over RenderWithOptions with Color left false.
func Render(r *Report, verbose bool) string {
	return RenderWithOptions(r, RenderOptions{Verbose: verbose})
}

// RenderWithOptions produces the CLI's presentation of r, applying opts.Color's restrained
// semantic vocabulary when set: the outcome symbol/summary bold in the outcome's own color (green
// success, yellow blocked/cancelled, red failed) as the report's one title line; section labels
// and the next-action arrow bold cyan as headings, visually distinct from the plain body text
// beneath them; and, within body/Next/Detail text, any `backtick-delimited` command, path, or
// identifier bolded in place — the one restrained device this package uses for "compact
// machine-oriented value", reusing Gnomon's own existing inline-code convention rather than
// attempting real syntax highlighting. Structure and content are identical to Render at
// Color: false — color never changes what a line says, only how it is emphasized.
func RenderWithOptions(r *Report, opts RenderOptions) string {
	var b strings.Builder

	titleColor := ""
	codeColor := ""
	if opts.Color {
		titleColor = ansiBold + r.Outcome.color()
		codeColor = ansiBold
	}

	b.WriteString(colorize(titleColor, r.Outcome.symbol()))
	b.WriteString(" ")
	summary := r.Summary
	if r.Target != "" && !strings.Contains(r.Summary, r.Target) {
		summary += " — " + r.Target
	}
	b.WriteString(colorize(titleColor, summary))
	b.WriteString("\n")

	headingColor := ""
	if opts.Color {
		headingColor = ansiBold + ansiCyan
	}

	for _, s := range r.Sections {
		if s.Body == "" {
			continue
		}
		b.WriteString("\n")
		b.WriteString(colorize(headingColor, s.Label))
		b.WriteString("\n")
		b.WriteString(indent(styleCodeSpans(s.Body, codeColor)))
		b.WriteString("\n")
	}

	if r.Next != "" {
		b.WriteString("\n")
		b.WriteString(colorize(headingColor, "→"))
		b.WriteString(" ")
		b.WriteString(styleCodeSpans(r.Next, codeColor))
		b.WriteString("\n")
	}

	if opts.Verbose && len(r.Detail) > 0 {
		mutedColor := ""
		if opts.Color {
			mutedColor = ansiDim
		}
		b.WriteString("\n")
		b.WriteString(colorize(headingColor, "Diagnostics:"))
		b.WriteString("\n")
		for _, d := range r.Detail {
			b.WriteString("  ")
			b.WriteString(colorize(mutedColor, d))
			b.WriteString("\n")
		}
	}

	return b.String()
}

// RenderError is the fallback presentation for a genuine, unclassified error with no Report —
// kept in the same small visual vocabulary rather than a raw Go error line. Always plain, exactly
// like Render; see RenderErrorColor for the color-aware equivalent.
func RenderError(err error) string {
	return "✗ " + err.Error() + "\n"
}

// RenderErrorColor is RenderError's content, additionally colored red when color is true — the
// one other place this package's vocabulary applies, since a Cobra-level error (an unknown
// command, a rejected flag) never produces a Report at all.
func RenderErrorColor(err error, color bool) string {
	if !color {
		return RenderError(err)
	}
	return colorize(ansiBold+ansiRed, "✗") + " " + err.Error() + "\n"
}

func indent(body string) string {
	lines := strings.Split(strings.TrimRight(body, "\n"), "\n")
	for i, l := range lines {
		lines[i] = "  " + l
	}
	return strings.Join(lines, "\n")
}
