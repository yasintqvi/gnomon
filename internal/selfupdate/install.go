package selfupdate

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// ExecutablePath resolves the file that is actually running right now — the one Gnomon replaces,
// never a fixed install location such as /usr/local/bin or $GOPATH/bin, so the update always
// targets whatever the Human actually invoked. os.Executable already resolves through /proc/self/exe
// on Linux (and the equivalent on other platforms), and EvalSymlinks normalizes the remaining
// cases, so this is the real underlying binary, not merely a symlink to it.
func ExecutablePath() (string, error) {
	p, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("locating the running executable: %w", err)
	}
	resolved, err := filepath.EvalSymlinks(p)
	if err != nil {
		// A resolution failure here is unusual but not fatal to the caller's other options —
		// fall back to the unresolved path rather than refusing outright.
		return p, nil
	}
	return resolved, nil
}

// replaceExecutable installs newContent as targetPath, never destroying the existing file if any
// step fails first.
//
// It writes newContent to a temporary file in targetPath's own directory — never the system temp
// directory, which may be a different filesystem/volume and would turn the final replacement into
// a copy instead of a rename — so the commit step below is a single filesystem-level rename
// rather than a multi-step copy that could be left half-done. The temporary file is removed on
// every path that does not end in a successful rename, including a panic-free early return.
func replaceExecutable(targetPath string, newContent []byte) (err error) {
	dir := filepath.Dir(targetPath)
	tmp, err := os.CreateTemp(dir, ".gnomon-update-*")
	if err != nil {
		return fmt.Errorf("cannot write to %s: %w", dir, err)
	}
	tmpPath := tmp.Name()
	committed := false
	defer func() {
		if !committed {
			os.Remove(tmpPath)
		}
	}()

	if _, err := tmp.Write(newContent); err != nil {
		tmp.Close()
		return fmt.Errorf("writing the update: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("writing the update: %w", err)
	}

	mode := os.FileMode(0o755)
	if info, statErr := os.Stat(targetPath); statErr == nil {
		mode = info.Mode().Perm()
	}
	if err := os.Chmod(tmpPath, mode); err != nil {
		return fmt.Errorf("setting permissions on the update: %w", err)
	}

	if runtime.GOOS == "windows" {
		return replaceOnWindows(targetPath, tmpPath, &committed)
	}

	// Unix: renaming onto a running executable is safe and atomic — the already-running process
	// keeps executing off its own now-unlinked inode, and the new file takes over the name for
	// every subsequent launch.
	if err := os.Rename(tmpPath, targetPath); err != nil {
		return fmt.Errorf("cannot replace %s: %w", targetPath, err)
	}
	committed = true
	return nil
}

// replaceOnWindows implements the standard, widely-used pattern for replacing a running Windows
// executable: Windows will not let the running image be overwritten or deleted in place, but it
// does allow the running file to be renamed aside, after which the new binary can take its name
// for the next launch. The renamed-aside original is best-effort cleaned up immediately; if that
// still fails because it is in use, it is left behind rather than treated as an update failure —
// nothing about the update itself is incomplete at that point.
func replaceOnWindows(targetPath, tmpPath string, committed *bool) error {
	oldPath := targetPath + ".old"
	os.Remove(oldPath) // best-effort cleanup of a previous update's leftover, if any

	if err := os.Rename(targetPath, oldPath); err != nil {
		return fmt.Errorf("cannot replace %s: %w", targetPath, err)
	}
	if err := os.Rename(tmpPath, targetPath); err != nil {
		// Restore the original so the installation is left exactly as it was, not half-updated.
		os.Rename(oldPath, targetPath)
		return fmt.Errorf("cannot replace %s: %w", targetPath, err)
	}
	*committed = true
	os.Remove(oldPath) // best-effort; may still be locked while this process is running
	return nil
}
