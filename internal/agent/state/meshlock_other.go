//go:build !linux && !windows

package state

// lockMesh takes no lock where there is neither (the data-dir lock and the
// control plane still apply).
func lockMesh(string) (func(), error) { return func() {}, nil }
