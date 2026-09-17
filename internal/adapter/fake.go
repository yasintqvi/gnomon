package adapter

import "os"

// FakeAdapter is a test-only Adapter — it writes a scripted result file (if any) and reports a
// scripted status, letting orchestration be tested fully without a real Agent, per
// cli/IMPLEMENTATION_ARCHITECTURE.md's testing strategy.
type FakeAdapter struct {
	Ctx           Context
	ResultContent []byte // written to Ctx.ResultPath during Run, if non-nil
	ExitCode      int
	RunErr        error
	Prepared      bool
	Terminated    bool // settable directly, so tests can exercise the presentation layer's handling of it
}

func (f *FakeAdapter) Prepare(ctx Context) error {
	f.Ctx = ctx
	f.Prepared = true
	return nil
}

func (f *FakeAdapter) Run() error {
	if f.ResultContent != nil {
		if err := os.WriteFile(f.Ctx.ResultPath, f.ResultContent, 0o644); err != nil {
			return err
		}
	}
	return f.RunErr
}

func (f *FakeAdapter) Cancel() error { return nil }

func (f *FakeAdapter) Status() Status {
	return Status{Started: true, Exited: true, ExitCode: f.ExitCode, Err: f.RunErr, Terminated: f.Terminated}
}
