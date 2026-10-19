//go:build !windows

package sysctl

// New returns the host's writer.
func New() *Writer { return &Writer{Root: "/proc/sys"} }
