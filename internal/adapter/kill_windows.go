//go:build windows

package adapter

import (
	"fmt"
	"os/exec"
)

// forceKill terminates the whole process tree rooted at cmd's process. Necessary on Windows
// because npm-installed claude/codex are .cmd shims run by cmd.exe: killing only that shim leaves
// its own child (e.g. node) running. Falls back to killing just the shim if taskkill itself fails.
var forceKill = func(cmd *exec.Cmd) error {
	if cmd.Process == nil {
		return nil
	}
	tk := exec.Command("taskkill", "/T", "/F", "/PID", fmt.Sprintf("%d", cmd.Process.Pid))
	if err := tk.Run(); err == nil {
		return nil
	}
	return cmd.Process.Kill()
}
