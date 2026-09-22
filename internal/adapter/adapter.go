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

	// Target is a free-form, untyped invocation target — used only by workflows whose own
	// Contract declares specification_reference: none but which still accept an optional target
	// argument (Verification, Review), per cli/WORKFLOW_CONTRACT.md's own explicit statement that
	// such a target "is not necessarily a Specification at all." Mutually exclusive with
	// SpecIdentity for any one invocation: a command populates exactly one of the two, or
	// neither, never both.
	Target     string // "" when not applicable
	ResultPath string
	RunID      string

	// Handoff carries the finding that caused this run to be launched, when this invocation was
	// dispatched from Gnomon's interactive Verification/Review finding-resolution loop — nil for
	// every ordinary invocation. It is transient invocation context only: never persisted, never
	// part of any Result Contract, gone the moment this one invocation ends. It explains WHY this
	// run started; it is never a substitute for SpecIdentity/Target above, which remain this
	// workflow's own, completely independent, normal target and eligibility requirements.
	Handoff *ResolutionHandoff

	// PollResult is called periodically while the Agent is still running, to check whether a
	// valid terminal result now exists. It must be safe to call repeatedly and side-effect-free
	// except for its final, successful call. The Adapter never interprets what "valid" means —
	// that determination is Gnomon's Result Protocol; the caller injects it as a plain function
	// so this package never imports Result Protocol or Contract types, preserving the rule that
	// no Gnomon semantics live inside the Adapter. May be left nil, in which case the Adapter
	// simply waits for the process to exit on its own, with no early detection or termination.
	PollResult func() (ready bool, err error)
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

	// Terminated is true when Gnomon itself asked the process to stop — either because
	// PollResult confirmed a valid result, or because Cancel() was called — as opposed to the
	// process ending on its own (a Human ending the session directly, or a crash). This is what
	// lets a presentation layer avoid describing an intentional post-success termination as a
	// failure, without needing to inspect a raw exit code to guess at the reason.
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
