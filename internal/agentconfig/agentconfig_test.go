package agentconfig

import (
	"os"
	"path/filepath"
	"testing"
)

// sandboxHome points os.UserConfigDir() at a fresh temp directory for the duration of one test,
// on every OS this is likely to run on: darwin consults only $HOME, Linux consults
// $XDG_CONFIG_HOME first, and setting both keeps the test portable without needing to know which
// rule the current OS uses.
func sandboxHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))
	return home
}

func TestLoad_MissingFileIsUnconfiguredNotError(t *testing.T) {
	sandboxHome(t)
	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected a missing config file to be a normal, unconfigured state, got error: %v", err)
	}
	if cfg.DefaultProvider != "" {
		t.Fatalf("expected an empty default provider, got %q", cfg.DefaultProvider)
	}
}

func TestSaveLoad_RoundTrip(t *testing.T) {
	sandboxHome(t)
	if err := Save(Config{DefaultProvider: "codex"}); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.DefaultProvider != "codex" {
		t.Fatalf("expected round-tripped provider %q, got %q", "codex", cfg.DefaultProvider)
	}
}

// TestSave_WritesToTheDocumentedOSStandardPath confirms the actual on-disk location matches what
// callers and documentation both state: <os.UserConfigDir()>/gnomon/config.json.
func TestSave_WritesToTheDocumentedOSStandardPath(t *testing.T) {
	sandboxHome(t)
	if err := Save(Config{DefaultProvider: "claude"}); err != nil {
		t.Fatal(err)
	}
	dir, err := os.UserConfigDir()
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(dir, "gnomon", "config.json")
	if _, err := os.Stat(want); err != nil {
		t.Fatalf("expected config at %s, got: %v", want, err)
	}
}

// TestSave_LeavesNoStrayTempFiles confirms a successful atomic write cleans up after itself —
// the temp file used to achieve atomicity must never remain alongside the real config.
func TestSave_LeavesNoStrayTempFiles(t *testing.T) {
	sandboxHome(t)
	if err := Save(Config{DefaultProvider: "claude"}); err != nil {
		t.Fatal(err)
	}
	p, err := path()
	if err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(filepath.Dir(p))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name() != "config.json" {
		t.Fatalf("expected exactly one file, config.json — an atomic write must leave no temp artifacts, got: %v", entries)
	}
}

// TestSave_FailureLeavesExistingConfigUntouched demonstrates atomicity directly: forcing the
// write half of Save to fail (an unwritable directory) must never corrupt or replace a config
// that was already durably saved — the write-to-temp-then-rename shape guarantees the real file
// is only ever touched by the final, successful rename.
func TestSave_FailureLeavesExistingConfigUntouched(t *testing.T) {
	sandboxHome(t)
	if err := Save(Config{DefaultProvider: "claude"}); err != nil {
		t.Fatal(err)
	}
	p, err := path()
	if err != nil {
		t.Fatal(err)
	}
	dir := filepath.Dir(p)

	if err := os.Chmod(dir, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(dir, 0o755) })

	if err := Save(Config{DefaultProvider: "codex"}); err == nil {
		t.Fatalf("expected Save to fail when its directory is not writable")
	}

	os.Chmod(dir, 0o755)
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.DefaultProvider != "claude" {
		t.Fatalf("expected the failed write to leave the prior config untouched, got %q", cfg.DefaultProvider)
	}
}

func TestSave_OverwritesPreviousValue(t *testing.T) {
	sandboxHome(t)
	if err := Save(Config{DefaultProvider: "claude"}); err != nil {
		t.Fatal(err)
	}
	if err := Save(Config{DefaultProvider: "codex"}); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.DefaultProvider != "codex" {
		t.Fatalf("expected the second Save to fully replace the first, got %q", cfg.DefaultProvider)
	}
}
