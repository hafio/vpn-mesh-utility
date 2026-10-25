//go:build !windows

package state

import (
	"errors"
	"os"
	"syscall"
)

// lockFile takes an exclusive flock on f, not waiting for it.
func lockFile(f *os.File) error {
	err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
	if errors.Is(err, syscall.EWOULDBLOCK) {
		return errBusy
	}
	return err
}

func unlockFile(f *os.File) { _ = syscall.Flock(int(f.Fd()), syscall.LOCK_UN) }
