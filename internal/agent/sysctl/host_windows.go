//go:build windows

package sysctl

// New returns the host's writer: Windows has no /proc/sys.
func New() *Writer { return &Writer{NoProc: true} }
