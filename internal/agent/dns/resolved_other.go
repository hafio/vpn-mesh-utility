//go:build !linux

package dns

// NewRegistrar returns nil: Windows registers through NRPT, which arrives
// in milestone M5. Until then the resolver serves the mesh and the state is
// "unavailable".
func NewRegistrar() Registrar { return nil }
