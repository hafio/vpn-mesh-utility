//go:build !windows

package fsutil

// transientRename is always false: a POSIX rename replaces a file that is
// open, so a failure is never a reader in the way.
func transientRename(error) bool { return false }
