package state

import (
	"errors"
	"fmt"
	"net"
	"syscall"
)

func lockMesh(name string) (func(), error) {
	l, err := net.Listen("unix", "@"+name)
	switch {
	case errors.Is(err, syscall.EADDRINUSE):
		return nil, fmt.Errorf("%w (in this network namespace); stop it first", ErrMeshLocked)
	case err != nil:
		return nil, fmt.Errorf("take the mesh's lock: %w", err)
	}
	return func() { _ = l.Close() }, nil // released either way
}
