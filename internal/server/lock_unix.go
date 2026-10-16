//go:build !windows

package server

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

// Lock takes the data dir's exclusive lock (<data>/server.lock). The server
// holds it while it runs and `meshctl import` takes it, so an import can
// never write under a running server (amendment A25). The lock is released
// when unlock is called or the process exits.
func Lock(dataDir string) (unlock func(), err error) {
	path := filepath.Join(dataDir, LockFile)
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		_ = f.Close() // the lock error is the one worth reporting
		if errors.Is(err, syscall.EWOULDBLOCK) {
			return nil, fmt.Errorf("%w (%s)", ErrLocked, path)
		}
		return nil, fmt.Errorf("lock %s: %w", path, err)
	}
	return func() {
		_ = syscall.Flock(int(f.Fd()), syscall.LOCK_UN) // closing releases it anyway
		_ = f.Close()
	}, nil
}
