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
	// Blocked — more is needed before this can proceed. Never a system failure.
	Blocked
	// Cancelled — the Human (or Gnomon, on their behalf) ended the run intentionally.
	Cancelled
	// Failed — something did not work as intended: a crash, a missing/invalid result, or any
	// other genuine error.
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

// ansi is this package's restrained semantic color vocabulary: green success, yellow
// needs-attention, red failure, cyan headings, bold for compact values, dim for secondary text.
// Applying color is purely a rendering-time decision, gated on a caller-supplied bool — it never
// changes Outcome, Summary, Section, or any other field a caller reads.
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

// Palette exposes this package's semantic color vocabulary for the rare caller that builds its own
// Human-facing text outside a Report (the interactive Specification workspace). Every method is a
// no-op when constructed with color disabled.
type Palette struct {
	color bool
}

// NewPalette returns a Palette that applies color only when color is true.
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

// Code styles a compact, machine-oriented value (a command, identifier, or path) bold — never
// syntax highlighting, every value gets the same treatment.
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

// styleCodeSpans bolds every `backtick-delimited` span in s, leaving the backticks and everything
// else untouched — reusing Markdown's inline-code convention as the marker for "this is a command,
// path, or identifier". A no-op when color is "".
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

// Section is one labeled, omittable block of content (e.g. "Delivered" or "Unresolved"). A Section
// with an empty Body is skipped entirely by Render.
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
// zero value ("plain, non-verbose").
type RenderOptions struct {
	// Verbose includes low-level diagnostics (exit codes, run IDs, paths).
	Verbose bool
	// Color applies this package's semantic ANSI vocabulary — purely presentational, content is
	// byte-identical either way. Callers decide this at the CLI edge; present never inspects the
	// environment itself.
	Color bool
}

// Render produces the CLI's plain-text presentation of r — no color, ever. A thin wrapper over
// RenderWithOptions with Color left false.
func Render(r *Report, verbose bool) string {
	return RenderWithOptions(r, RenderOptions{Verbose: verbose})
}

// RenderWithOptions produces the CLI's presentation of r, applying opts.Color's semantic
// vocabulary when set: the outcome line in its own color, section labels and the next-action arrow
// as cyan headings, and any `backtick-delimited` value bolded in place. Structure and content are
// identical to Render at Color: false — color never changes what a line says.
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

// RenderError is the fallback presentation for a genuine error with no Report. Always plain; see
// RenderErrorColor for the color-aware equivalent.
func RenderError(err error) string {
	return "✗ " + err.Error() + "\n"
}

// RenderErrorColor is RenderError's content, additionally colored red when color is true.
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
