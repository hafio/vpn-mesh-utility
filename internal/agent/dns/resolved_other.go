//go:build !linux && !windows

package dns

import "github.com/hafio/vpn-mesh-utility/internal/proto"

// NewRegistrar returns nil: this platform has no host resolver to register
// with, so the resolver serves the mesh and the state is "unavailable".
func NewRegistrar(proto.Slot) Registrar { return nil }
