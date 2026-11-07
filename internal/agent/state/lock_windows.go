//go:build windows

package state

import (
	"errors"
	"os"

	"golang.org/x/sys/windows"
)

// lockOffset is where the locked byte lies: far past the pid the file
// holds, which another process must still be able to read.
const lockOffset = 1 << 30

// unwritable says why the agent cannot write its data dir (%s) and what to
// do: the service runs as LocalSystem, which the dir's ACL must allow.
const unwritable = "%s must give SYSTEM and Administrators full control; run the agent elevated or as its service"

// lockFile takes an exclusive lock on f, not waiting for it.
func lockFile(f *os.File) error {
	ol := windows.Overlapped{Offset: lockOffset}
	err := windows.LockFileEx(windows.Handle(f.Fd()), windows.LOCKFILE_EXCLUSIVE_LOCK|windows.LOCKFILE_FAIL_IMMEDIATELY, 0, 1, 0, &ol)
	if errors.Is(err, windows.ERROR_LOCK_VIOLATION) {
		return errBusy
	}
	return err
}

func unlockFile(f *os.File) {
	ol := windows.Overlapped{Offset: lockOffset}
	_ = windows.UnlockFileEx(windows.Handle(f.Fd()), 0, 1, 0, &ol)
}
