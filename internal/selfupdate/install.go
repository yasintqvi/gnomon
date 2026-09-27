package selfupdate

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// ExecutablePath resolves the file that is actually running right now — the one Gnomon replaces,
// never a fixed install location — so the update always targets whatever the Human actually
// invoked. EvalSymlinks ensures this is the real binary, not merely a symlink to it.
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

// commitRename is the actual rename that makes the new executable live — Unix's single rename,
// and Windows' second (committing) rename in replaceOnWindows — as one overridable seam, so a
// test can inject a failure at that specific step without depending on real filesystem
// permissions (which root bypasses, and which Windows does not enforce the same way for
// directories).
var commitRename = os.Rename

// replaceExecutable installs newContent as targetPath, never destroying the existing file if any
// step fails first. Writes to a temp file in targetPath's own directory (never the system temp
// dir, which may be a different filesystem and would turn the commit into a copy instead of a
// rename). The temp file is removed on every path that doesn't end in a successful rename.
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

	// Unix: renaming onto a running executable is safe and atomic — it keeps executing off its own
	// now-unlinked inode, and the new file takes over the name for every subsequent launch.
	if err := commitRename(tmpPath, targetPath); err != nil {
		return fmt.Errorf("cannot replace %s: %w", targetPath, err)
	}
	committed = true
	return nil
}

// replaceOnWindows implements the standard pattern for replacing a running Windows executable:
// Windows won't let the running image be overwritten in place, but does allow it to be renamed
// aside, after which the new binary takes its name. The renamed-aside original is best-effort
// cleaned up; if that fails because it's in use, it's left behind, not treated as a failure.
func replaceOnWindows(targetPath, tmpPath string, committed *bool) error {
	oldPath := targetPath + ".old"
	os.Remove(oldPath) // best-effort cleanup of a previous update's leftover, if any

	if err := os.Rename(targetPath, oldPath); err != nil {
		return fmt.Errorf("cannot replace %s: %w", targetPath, err)
	}
	if err := commitRename(tmpPath, targetPath); err != nil {
		// Restore the original so the installation is left exactly as it was, not half-updated.
		os.Rename(oldPath, targetPath)
		return fmt.Errorf("cannot replace %s: %w", targetPath, err)
	}
	*committed = true
	os.Remove(oldPath) // best-effort; may still be locked while this process is running
	return nil
}
