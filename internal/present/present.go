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
// for success, yellow for a Blocked/Cancelled outcome that is not a failure, red for an actual
// failure or unclassified error, cyan for headings/structure/next-action text. Applying color is
// purely a rendering-time decision (RenderWithOptions/RenderErrorColor, both gated on a caller-
// supplied bool) — it never changes Outcome, Summary, Section, or any other field a caller reads.
const (
	ansiReset  = "\033[0m"
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
// semantic vocabulary when set: the outcome symbol/summary in the outcome's own color (green
// success, yellow blocked/cancelled, red failed), section labels and the next-action arrow in
// cyan. Structure and content are identical to Render at Color: false.
func RenderWithOptions(r *Report, opts RenderOptions) string {
	var b strings.Builder

	color := ""
	if opts.Color {
		color = r.Outcome.color()
	}

	b.WriteString(colorize(color, r.Outcome.symbol()))
	b.WriteString(" ")
	summary := r.Summary
	if r.Target != "" && !strings.Contains(r.Summary, r.Target) {
		summary += " — " + r.Target
	}
	b.WriteString(colorize(color, summary))
	b.WriteString("\n")

	headingColor := ""
	if opts.Color {
		headingColor = ansiCyan
	}

	for _, s := range r.Sections {
		if s.Body == "" {
			continue
		}
		b.WriteString("\n")
		b.WriteString(colorize(headingColor, s.Label))
		b.WriteString("\n")
		b.WriteString(indent(s.Body))
		b.WriteString("\n")
	}

	if r.Next != "" {
		b.WriteString("\n")
		b.WriteString(colorize(headingColor, "→"))
		b.WriteString(" ")
		b.WriteString(r.Next)
		b.WriteString("\n")
	}

	if opts.Verbose && len(r.Detail) > 0 {
		b.WriteString("\n")
		b.WriteString(colorize(headingColor, "Diagnostics:"))
		b.WriteString("\n")
		for _, d := range r.Detail {
			b.WriteString("  ")
			b.WriteString(d)
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
	return colorize(ansiRed, "✗") + " " + err.Error() + "\n"
}

func indent(body string) string {
	lines := strings.Split(strings.TrimRight(body, "\n"), "\n")
	for i, l := range lines {
		lines[i] = "  " + l
	}
	return strings.Join(lines, "\n")
}
