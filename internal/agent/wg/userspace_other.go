//go:build !linux && !windows

package wg

import (
	"fmt"
	"log/slog"
	"runtime"
)

// NewUserspace runs on Linux and Windows only.
func NewUserspace(*slog.Logger) (Host, error) {
	return nil, fmt.Errorf("userspace WireGuard does not run on %s", runtime.GOOS)
}
