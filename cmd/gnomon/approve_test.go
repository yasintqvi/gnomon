package main

import (
	"io"
	"os"
	"strings"
	"testing"
)

// TestApproveConfirm_NonInteractive_PrintsWarningAndProceeds is the only branch exercisable
// without a real terminal: go test's own stdin is never a TTY, so approveConfirm must always take
// the non-interactive path here — print the warning, then proceed without asking.
func TestApproveConfirm_NonInteractive_PrintsWarningAndProceeds(t *testing.T) {
	if stdinIsInteractive() {
		t.Skip("stdin is unexpectedly a real terminal in this test environment")
	}

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	orig := os.Stderr
	os.Stderr = w
	defer func() { os.Stderr = orig }()

	proceed, err := approveConfirm("SPEC-001 still contains 3 template placeholders")
	w.Close()
	os.Stderr = orig

	if err != nil {
		t.Fatalf("approveConfirm: %v", err)
	}
	if !proceed {
		t.Fatalf("expected non-interactive approveConfirm to proceed without asking")
	}

	out, _ := io.ReadAll(r)
	if !strings.Contains(string(out), "template placeholders") {
		t.Fatalf("expected the warning printed to stderr, got %q", out)
	}
}
