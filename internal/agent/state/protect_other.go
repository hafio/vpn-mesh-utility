//go:build !windows

package state

// protect is a no-op where the directory's mode (0700, set by Open) is the
// guard.
func protect(string) error { return nil }
