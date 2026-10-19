//go:build windows

package exitmode

// NewPolicy is nil on Windows: it has no policy routing, so the WireGuard
// host does exit mode itself (wg.ExitRouter) with split default routes and
// a socket bound to the physical interface.
func NewPolicy() Policy { return nil }
