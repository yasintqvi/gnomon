package selfupdate

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestReplaceExecutable_Success_ReplacesContentAndPreservesMode(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "gnomon")
	if err := os.WriteFile(target, []byte("old binary"), 0o744); err != nil {
		t.Fatalf("seeding target: %v", err)
	}

	if err := replaceExecutable(target, []byte("new binary")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("reading replaced file: %v", err)
	}
	if string(got) != "new binary" {
		t.Fatalf("got %q, want %q", got, "new binary")
	}

	info, err := os.Stat(target)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if runtime.GOOS != "windows" && info.Mode().Perm() != 0o744 {
		t.Fatalf("expected the original file's permissions (0744) preserved, got %o", info.Mode().Perm())
	}

	// No leftover temp file should remain in the target's directory.
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("reading dir: %v", err)
	}
	if len(entries) != 1 || entries[0].Name() != "gnomon" {
		t.Fatalf("expected only the replaced executable to remain, got %+v", entries)
	}
}

// TestReplaceExecutable_UnwritableDirectory_LeavesOriginalIntact is the direct proof for
// requirement 14/15: when the target's directory cannot be written to (a permissions problem,
// never elevated automatically), the existing executable is left exactly as it was.
func TestReplaceExecutable_UnwritableDirectory_LeavesOriginalIntact(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("directory write-permission semantics differ on windows; covered on unix runners")
	}
	if os.Geteuid() == 0 {
		t.Skip("running as root bypasses permission checks")
	}

	dir := t.TempDir()
	target := filepath.Join(dir, "gnomon")
	if err := os.WriteFile(target, []byte("old binary"), 0o755); err != nil {
		t.Fatalf("seeding target: %v", err)
	}
	if err := os.Chmod(dir, 0o555); err != nil {
		t.Fatalf("making dir read-only: %v", err)
	}
	t.Cleanup(func() { os.Chmod(dir, 0o755) }) // let t.TempDir() clean up afterward

	err := replaceExecutable(target, []byte("new binary"))
	if err == nil {
		t.Fatal("expected replaceExecutable to fail against an unwritable directory")
	}

	got, readErr := os.ReadFile(target)
	if readErr != nil {
		t.Fatalf("reading target after failed update: %v", readErr)
	}
	if string(got) != "old binary" {
		t.Fatalf("expected the original executable untouched, got %q", got)
	}
}

func TestExecutablePath_ReturnsAnExistingFile(t *testing.T) {
	p, err := ExecutablePath()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := os.Stat(p); err != nil {
		t.Fatalf("ExecutablePath returned %q, which does not exist: %v", p, err)
	}
}

// --- Windows replacement logic, exercised directly so it has real coverage even though this
// suite normally runs on unix: replaceOnWindows performs plain, OS-agnostic rename calls, so its
// logic (not the OS's own running-executable semantics, which cannot be verified off Windows) is
// fully testable here.

func TestReplaceOnWindows_Success(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "gnomon.exe")
	tmp := filepath.Join(dir, ".gnomon-update-tmp")
	if err := os.WriteFile(target, []byte("old"), 0o755); err != nil {
		t.Fatalf("seeding target: %v", err)
	}
	if err := os.WriteFile(tmp, []byte("new"), 0o755); err != nil {
		t.Fatalf("seeding tmp: %v", err)
	}

	var committed bool
	if err := replaceOnWindows(target, tmp, &committed); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !committed {
		t.Fatal("expected committed to be set true on success")
	}
	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("reading replaced target: %v", err)
	}
	if string(got) != "new" {
		t.Fatalf("got %q, want %q", got, "new")
	}
}

// TestReplaceOnWindows_SecondRenameFails_RestoresOriginal proves that if installing the new file
// fails after the original has already been renamed aside, the original is put back rather than
// left missing.
func TestReplaceOnWindows_SecondRenameFails_RestoresOriginal(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "gnomon.exe")
	missingTmp := filepath.Join(dir, "does-not-exist")
	if err := os.WriteFile(target, []byte("old"), 0o755); err != nil {
		t.Fatalf("seeding target: %v", err)
	}

	var committed bool
	err := replaceOnWindows(target, missingTmp, &committed)
	if err == nil {
		t.Fatal("expected an error when the new file cannot be moved into place")
	}
	if committed {
		t.Fatal("expected committed to remain false on failure")
	}
	got, readErr := os.ReadFile(target)
	if readErr != nil {
		t.Fatalf("expected the original to be restored at target, but it is missing: %v", readErr)
	}
	if string(got) != "old" {
		t.Fatalf("expected the restored original content, got %q", got)
	}
}
