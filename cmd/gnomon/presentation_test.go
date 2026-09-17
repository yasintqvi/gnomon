package main

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// captureOutput runs fn with rootCmd's output redirected to a buffer and returns what was
// written — the same mechanism Cobra itself recommends for exercising --help/--version without a
// real terminal.
func captureOutput(t *testing.T, args []string) string {
	t.Helper()
	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs(args)
	defer func() {
		rootCmd.SetOut(nil)
		rootCmd.SetErr(nil)
		rootCmd.SetArgs(nil)
		// Cobra's auto-added --help/--version bool flags persist their last-Set value across
		// Execute() calls on the same *Command tree (pflag never zeroes a flag merely because a
		// later Parse omits it) — reset both, on every command in the tree, so one test's --help
		// or --version can never leak into the next test's plain invocation.
		resetHelpAndVersionFlags(rootCmd)
	}()
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("rootCmd.Execute(%v): %v", args, err)
	}
	return buf.String()
}

func resetHelpAndVersionFlags(cmd *cobra.Command) {
	if f := cmd.Flags().Lookup("help"); f != nil {
		_ = f.Value.Set("false")
	}
	if f := cmd.Flags().Lookup("version"); f != nil {
		_ = f.Value.Set("false")
	}
	for _, sub := range cmd.Commands() {
		resetHelpAndVersionFlags(sub)
	}
}

// --- Root experience ---

func TestRoot_BareInvocation_IdentifiesGnomon(t *testing.T) {
	out := captureOutput(t, []string{})
	if !strings.Contains(out, "Gnomon") || !strings.Contains(out, "AI Software Engineering System") {
		t.Fatalf("expected a bare invocation to identify the product, got: %s", out)
	}
	if !strings.Contains(out, "gnomon status") || !strings.Contains(out, "gnomon next") {
		t.Fatalf("expected a bare invocation to point toward status/next as the entry points, got: %s", out)
	}
}

func TestRoot_HelpListsFullCommandSurface(t *testing.T) {
	out := captureOutput(t, []string{"--help"})
	for _, name := range []string{
		"init", "describe", "bootstrap", "spec", "approve", "revoke",
		"implement", "test", "verify", "review", "finalize",
		"status", "validate", "next", "run", "agent",
	} {
		if !strings.Contains(out, name) {
			t.Errorf("expected --help to list %q somewhere in the command surface", name)
		}
	}
}

func TestRoot_HelpText_UsesIntentNotImplementation(t *testing.T) {
	out := captureOutput(t, []string{"--help"})
	for _, jargon := range []string{"orchestrate.", "adapter.Adapter", "Result Protocol", "internal/"} {
		if strings.Contains(out, jargon) {
			t.Errorf("did not expect internal implementation jargon %q in help output: %s", jargon, out)
		}
	}
}

// --- Command grouping (item 3) ---

func TestRoot_CommandGrouping_MatchesDesignedCategories(t *testing.T) {
	want := map[string]string{
		"init": groupProject, "describe": groupProject, "bootstrap": groupProject,
		"spec": groupSpecification, "approve": groupSpecification, "revoke": groupSpecification,
		"implement": groupEngineering, "test": groupEngineering, "verify": groupEngineering,
		"review": groupEngineering, "finalize": groupEngineering,
		"status": groupGuidance, "validate": groupGuidance, "next": groupGuidance,
		"run": groupAdvanced, "agent": groupAdvanced,
	}
	got := map[string]string{}
	for _, c := range rootCmd.Commands() {
		got[c.Name()] = c.GroupID
	}
	for name, group := range want {
		if got[name] != group {
			t.Errorf("expected %q grouped under %q, got %q", name, group, got[name])
		}
	}
}

// TestRoot_ApproveRevoke_RemainTopLevel guards against the one explicit "do not move" instruction:
// approve/revoke are grouped for presentation, but never nested under spec.
func TestRoot_ApproveRevoke_RemainTopLevel(t *testing.T) {
	for _, c := range rootCmd.Commands() {
		if c.Name() != "spec" {
			continue
		}
		for _, sub := range c.Commands() {
			if sub.Name() == "approve" || sub.Name() == "revoke" {
				t.Fatalf("approve/revoke must remain top-level, never nested under spec")
			}
		}
	}
}

func TestRoot_CompletionCommand_HiddenButFunctional(t *testing.T) {
	if !rootCmd.CompletionOptions.HiddenDefaultCmd {
		t.Fatalf("expected the default completion command to be hidden from help — it is not a designed part of v1")
	}
}

// --- Version (item 9) ---

func TestVersion_FlagExists_NoShorthandCollisionWithVerbose(t *testing.T) {
	out := captureOutput(t, []string{"--version"})
	if !strings.Contains(out, "gnomon") {
		t.Fatalf("expected --version output to name gnomon, got: %q", out)
	}
	if !strings.Contains(out, version) {
		t.Fatalf("expected --version output to include the build version %q, got: %q", version, out)
	}
	// -v is already --verbose; --version must never silently steal it.
	f := rootCmd.PersistentFlags().ShorthandLookup("v")
	if f == nil || f.Name != "verbose" {
		t.Fatalf("expected -v to remain bound to --verbose, not --version")
	}
}

func TestVersion_DefaultsToDev_NeverHardcodedRelease(t *testing.T) {
	if version == "" {
		t.Fatalf("expected a non-empty default version")
	}
	if version == "v1.0.0" || version == "1.0.0" {
		t.Fatalf("did not expect the version to be hardcoded to a release value during Step 13")
	}
}

// --- Args error presentation (item 7) ---

func TestArgsErrors_NameWhatIsExpected_NotRawCobraCount(t *testing.T) {
	err := implementCmd.Args(implementCmd, nil)
	if err == nil {
		t.Fatalf("expected an error for a missing required argument")
	}
	if !strings.Contains(err.Error(), "<SPEC-id>") {
		t.Fatalf("expected the error to name the missing argument, got: %q", err.Error())
	}
	if strings.Contains(err.Error(), "arg(s)") {
		t.Fatalf("did not expect Cobra's raw 'arg(s)' phrasing to leak through: %q", err.Error())
	}
}

func TestArgsErrors_StillEnforceTheSameArity(t *testing.T) {
	// requireArgs must never change accept/reject behavior, only the message.
	if err := implementCmd.Args(implementCmd, []string{"SPEC-001"}); err != nil {
		t.Fatalf("expected exactly one arg to still be accepted: %v", err)
	}
	if err := implementCmd.Args(implementCmd, []string{"SPEC-001", "extra"}); err == nil {
		t.Fatalf("expected more than one arg to still be refused")
	}
	if err := runCmd.Args(runCmd, []string{"identity"}); err != nil {
		t.Fatalf("expected run to still accept one arg: %v", err)
	}
	if err := runCmd.Args(runCmd, []string{"identity", "target"}); err != nil {
		t.Fatalf("expected run to still accept two args: %v", err)
	}
	if err := runCmd.Args(runCmd, nil); err == nil {
		t.Fatalf("expected run to still refuse zero args")
	}
}

// --- status/validate/next wording distinctions (item 11) ---

func TestGuidanceCommands_HelpText_ExplainsDistinctFromEachOther(t *testing.T) {
	statusOut := captureOutput(t, []string{"status", "--help"})
	if !strings.Contains(statusOut, "validate") {
		t.Errorf("expected status --help to point at validate for the distinction: %s", statusOut)
	}
	validateOut := captureOutput(t, []string{"validate", "--help"})
	if !strings.Contains(validateOut, "status") {
		t.Errorf("expected validate --help to point at status for the distinction: %s", validateOut)
	}
	nextOut := captureOutput(t, []string{"next", "--help"})
	lower := strings.ToLower(nextOut)
	if strings.Contains(lower, "ai planner") || strings.Contains(lower, "prioritiz") {
		t.Errorf("did not expect next --help to sound like a planning/prioritization engine: %s", nextOut)
	}
	if !strings.Contains(nextOut, "deterministic") {
		t.Errorf("expected next --help to state it is deterministic: %s", nextOut)
	}
	if !strings.Contains(nextOut, "never a single") {
		t.Errorf("expected next --help to explicitly disclaim picking a single recommended action: %s", nextOut)
	}
}

// --- run discoverability (item 12) ---

func TestRun_HelpText_MarksItselfAdvanced(t *testing.T) {
	if !strings.Contains(strings.ToLower(runCmd.Short), "advanced") {
		t.Fatalf("expected run's Short description to mark it as the advanced/generic path, got: %q", runCmd.Short)
	}
	out := captureOutput(t, []string{"run", "--help"})
	if !strings.Contains(out, "implement") {
		t.Fatalf("expected run's help to name the dedicated commands ordinary users should prefer: %s", out)
	}
	if runCmd.GroupID != groupAdvanced {
		t.Fatalf("expected run grouped under Advanced, got %q", runCmd.GroupID)
	}
}

// --- Color / TTY / NO_COLOR (items 5-6) ---

func TestColorEnabled_RespectsNoColor(t *testing.T) {
	old, had := os.LookupEnv("NO_COLOR")
	t.Setenv("NO_COLOR", "1")
	if colorEnabled() {
		t.Fatalf("expected NO_COLOR to force color off regardless of TTY state")
	}
	if had {
		t.Setenv("NO_COLOR", old)
	}
}

func TestColorEnabled_FalseForNonTTYTestProcess(t *testing.T) {
	// go test's own stdout is never a real interactive terminal.
	if colorEnabled() {
		t.Fatalf("expected color to be disabled for a non-TTY process")
	}
}

// --- Agent selector regression (item 10) ---

func TestAgentChooser_NonInteractive_NeverPanicsOrHangs(t *testing.T) {
	// With stdin not a terminal (the normal `go test` condition), the chooser must be nil, never
	// a real interactive menu — this is the exact defect fixed in this slice (stdinIsInteractive
	// previously misclassified non-TTY character devices like /dev/null as interactive).
	if agentChooserForInvocation() != nil {
		t.Fatalf("expected a nil chooser when stdin is not a real terminal")
	}
}

func TestMoveSelection_StillWorks_SelectorUnchanged(t *testing.T) {
	// Regression guard: Step 13 must not have altered Slice 0's selector navigation logic.
	if got := moveSelection(0, -1, 2); got != 1 {
		t.Fatalf("expected wraparound up from 0 over 2 items to land on 1, got %d", got)
	}
	if got := moveSelection(1, 1, 2); got != 0 {
		t.Fatalf("expected wraparound down from 1 over 2 items to land on 0, got %d", got)
	}
}
