// Package fsutil holds the file-system helpers both the server and the agent
// need. It is stdlib-only and imports nothing else from this module.
package fsutil

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

// renameBudget is how long a rename over a file someone else has open is
// retried. On Windows every os.Open shares neither delete nor rename, so the
// rename fails until that reader closes the file -- a moment for a reader,
// longer for a virus scanner.
const renameBudget = 2 * time.Second

// WriteAtomic replaces dir/name with data: a temp file in the same directory,
// fsync, rename over the target, fsync of the directory. A crash leaves the
// old file or the new one, never half of either.
func WriteAtomic(dir, name string, data []byte, perm fs.FileMode) error {
	target := filepath.Join(dir, name)
	f, err := os.CreateTemp(dir, "."+name+".tmp-*")
	if err != nil {
		return fmt.Errorf("write %s: %w", target, err)
	}
	tmp := f.Name()
	done := false
	defer func() {
		if !done {
			_ = os.Remove(tmp) // best effort; the write error is what matters
		}
	}()
	_, werr := f.Write(data)
	if werr == nil {
		werr = f.Chmod(perm)
	}
	if werr == nil {
		werr = f.Sync()
	}
	if cerr := f.Close(); werr == nil {
		werr = cerr
	}
	if werr != nil {
		return fmt.Errorf("write %s: %w", target, werr)
	}
	if err := renameRetrying(tmp, target, renameBudget, transientRename); err != nil {
		return fmt.Errorf("write %s: %w", target, err)
	}
	done = true
	return syncDir(dir)
}

// renameRetrying renames tmp over target, trying again while transient
// says the failure is a reader in the way, until budget has passed.
func renameRetrying(tmp, target string, budget time.Duration, transient func(error) bool) error {
	deadline := time.Now().Add(budget)
	for wait := time.Millisecond; ; wait = min(2*wait, 100*time.Millisecond) {
		err := os.Rename(tmp, target)
		if err == nil || !transient(err) || time.Now().Add(wait).After(deadline) {
			return err
		}
		time.Sleep(wait)
	}
}

// syncDir makes the rename durable. Windows cannot fsync a directory; NTFS
// journals the rename itself.
func syncDir(dir string) error {
	if runtime.GOOS == "windows" {
		return nil
	}
	d, err := os.Open(dir)
	if err != nil {
		return fmt.Errorf("sync %s: %w", dir, err)
	}
	defer d.Close()
	if err := d.Sync(); err != nil {
		return fmt.Errorf("sync %s: %w", dir, err)
	}
	return nil
}
