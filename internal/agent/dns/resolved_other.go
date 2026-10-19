//go:build !linux && !windows

package dns

// NewRegistrar returns nil: this platform has no host resolver to register
// with, so the resolver serves the mesh and the state is "unavailable".
func NewRegistrar() Registrar { return nil }
