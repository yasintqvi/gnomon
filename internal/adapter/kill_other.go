//go:build !windows

package adapter

import "os/exec"

// forceKill just kills the one process — see run.go's documented decision against process-group
// management on Unix: no process tree to worry about here the way a Windows .cmd shim has one.
var forceKill = func(cmd *exec.Cmd) error {
	if cmd.Process == nil {
		return nil
	}
	return cmd.Process.Kill()
}
