//go:build !windows

package orchestrate

import (
	"errors"
	"os"
	"syscall"
)

var errLockHeld = errors.New("lock held by another process")

// tryLockFile takes an exclusive flock on f without waiting. flock belongs to the open file, so a
// second open of the same path conflicts even within one process.
func tryLockFile(f *os.File) error {
	err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
	if errors.Is(err, syscall.EWOULDBLOCK) {
		return errLockHeld
	}
	return err
}

func unlockFile(f *os.File) {
	_ = syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
}
