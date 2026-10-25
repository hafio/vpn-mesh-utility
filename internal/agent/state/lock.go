package state

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// ErrLocked means another process uses the data dir: an agent runs on it,
// or a join is under way. Nothing was touched.
var ErrLocked = errors.New("another meshctl process uses this data dir")

// Lock takes the data dir's lock for this process (amendment A50): `agent
// run` holds it for its whole life and `agent join` while it writes keys,
// so a second agent, or a join under a running one, is refused before it
// touches anything. The kernel releases the lock when the process dies, so
// a crash never leaves it behind; unlock releases it earlier.
func (d *Dir) Lock() (unlock func(), err error) {
	f, err := os.OpenFile(d.file(LockFile), os.O_RDWR|os.O_CREATE, 0o600)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", d.file(LockFile), err)
	}
	if err := lockFile(f); err != nil {
		holder := ""
		if b, rerr := os.ReadFile(d.file(LockFile)); rerr == nil {
			if pid, perr := strconv.Atoi(strings.TrimSpace(string(b))); perr == nil {
				holder = fmt.Sprintf(" (pid %d)", pid)
			}
		}
		_ = f.Close() // not ours
		if errors.Is(err, errBusy) {
			return nil, fmt.Errorf("%w%s: %s; stop it first", ErrLocked, holder, d.Path)
		}
		return nil, fmt.Errorf("lock %s: %w", d.file(LockFile), err)
	}
	// The holder's pid, for the next one's message; best effort.
	if err := f.Truncate(0); err == nil {
		_, _ = f.WriteAt([]byte(strconv.Itoa(os.Getpid())+"\n"), 0)
	}
	return func() {
		unlockFile(f)
		_ = f.Close() // released either way
	}, nil
}

// errBusy is lockFile's answer when another process holds the lock.
var errBusy = errors.New("held by another process")
