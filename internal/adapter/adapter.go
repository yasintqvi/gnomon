// Package adapter defines the provider-neutral Agent lifecycle (CLI Step 4) — the only interface
// boundary this architecture requires. It contains no Gnomon semantics: it never interprets an
// Outcome, a lifecycle state, or a dependency fact.
package adapter

// Context is the minimum provider-neutral invocation context the CLI prepares before starting
// an Agent — references only, never duplicated file content.
type Context struct {
	ProjectRoot      string
	GnomonRoot       string
	WorkflowPath     string
	WorkflowIdentity string
	SpecIdentity     string // "" when not applicable
	SpecPath         string // the governing Specification's file, project-relative; "" when not applicable

	// Knowledge lists project-written knowledge files (project-relative), excluding unmodified
	// Gnomon templates — references only, never content.
	Knowledge []string

	// Target is a free-form, untyped invocation target (Verification, Review) — never checked as
	// a Specification (cli/WORKFLOW_CONTRACT.md). Mutually exclusive with SpecIdentity.
	Target string // "" when not applicable

	// ScopeNote is extra, workflow-specific text about what this run covers (Verification's mode
	// and change set) — passed through to the Agent verbatim, never interpreted here.
	ScopeNote string

	// UserRequest is what the user asked for on the command line for this run, verbatim; "" when
	// they gave nothing.
	UserRequest string

	// ReadFirst lists project-relative files the Agent must read before starting, each with why —
	// prepared by the CLI for this run (e.g. decisions extracted from other Specifications).
	ReadFirst []ReadFirstFile

	ResultPath string
	RunID      string

	// Handoff carries the finding that caused this run to be launched — nil for every ordinary
	// invocation, never persisted, never a substitute for SpecIdentity/Target's own requirements.
	Handoff *ResolutionHandoff

	// PollResult checks periodically whether a valid terminal result now exists; safe to call
	// repeatedly. Injected as a plain function so this package never imports Result Protocol
	// types. Nil means the Adapter simply waits for the process to exit on its own.
	PollResult func() (ready bool, err error)
}

// ReadFirstFile is one file the Agent must read before starting, and why.
type ReadFirstFile struct {
	Path    string
	Purpose string
}

// ResolutionHandoff is the finding a resolution Agent invocation was launched to address —
// carried as plain reference data, never interpreted by this package. See Context.Handoff.
type ResolutionHandoff struct {
	OriginWorkflow string // "verification" | "review"
	OriginTarget   string // the target the originating evaluation ran against
	FindingID      string // report-local; only meaningful for this one handoff
	Classification string
	Summary        string
	Evidence       string
}

// Status is the provider/runtime completion information available once a run has ended —
// deliberately never a Core workflow Outcome.
type Status struct {
	Started  bool
	Exited   bool
	ExitCode int
	Err      error

	// Terminated is true when Gnomon itself asked the process to stop (PollResult confirmed a
	// result, or Cancel), so a presentation layer never describes it as an Agent failure.
	Terminated bool
}

// Adapter is the small lifecycle every Agent integration implements: prepare, run (Terminal
// Handoff, blocking), cancel, and report completion status.
type Adapter interface {
	Prepare(ctx Context) error
	Run() error
	Cancel() error
	Status() Status
}
