package state

import "errors"

// ErrMeshLocked means another agent of the mesh runs where this one would
// (amendment A50): in the same network namespace on Linux, on the same
// host on Windows -- a copy of this data dir in another directory, or
// another node of the mesh, which would fight over its interface.
var ErrMeshLocked = errors.New("another agent of this mesh already runs here")

// LockMesh takes the mesh's lock, named after key (its uuid), for this
// process: an abstract unix socket on Linux, which the kernel scopes to the
// network namespace and drops when the process dies; a global mutex on
// Windows. unlock releases it earlier.
func LockMesh(key string) (unlock func(), err error) { return lockMesh("meshctl-" + key) }
