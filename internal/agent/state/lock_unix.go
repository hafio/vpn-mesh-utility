//go:build !windows

package state

import (
	"errors"
	"os"
	"syscall"
)

// unwritable says why the agent cannot write its data dir (%s) and what to
// do: it runs as root without CAP_DAC_OVERRIDE (the unit's and the
// samples' capability set), so root must own the dir or be in its group.
const unwritable = "the agent runs as root without CAP_DAC_OVERRIDE, so %s must belong to user or group 0; " +
	"in a container, a volume an older image made is the control plane's (uid 65532): " +
	"chown it to 0:0, or remove it if the node never joined (docs/containers.md, troubleshooting)"

// lockFile takes an exclusive flock on f, not waiting for it.
func lockFile(f *os.File) error {
	err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
	if errors.Is(err, syscall.EWOULDBLOCK) {
		return errBusy
	}
	return err
}

func unlockFile(f *os.File) { _ = syscall.Flock(int(f.Fd()), syscall.LOCK_UN) }
