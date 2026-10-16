//go:build !linux

package wg

import (
	"fmt"
	"log/slog"
	"runtime"
)

// NewUserspace is Linux-only until Windows gets wintun in M5.
func NewUserspace(*slog.Logger) (Host, error) {
	return nil, fmt.Errorf("userspace WireGuard on %s arrives in milestone M5", runtime.GOOS)
}
