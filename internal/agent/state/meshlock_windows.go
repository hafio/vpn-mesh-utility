package state

import (
	"errors"
	"fmt"

	"golang.org/x/sys/windows"
)

func lockMesh(name string) (func(), error) {
	p, err := windows.UTF16PtrFromString(`Global\` + name)
	if err != nil {
		return nil, fmt.Errorf("take the mesh's lock: %w", err)
	}
	h, err := windows.CreateMutex(nil, false, p)
	switch {
	case errors.Is(err, windows.ERROR_ALREADY_EXISTS):
		_ = windows.CloseHandle(h) // not ours
		return nil, fmt.Errorf("%w (on this host); stop it first", ErrMeshLocked)
	case err != nil:
		return nil, fmt.Errorf("take the mesh's lock: %w", err)
	}
	return func() { _ = windows.CloseHandle(h) }, nil // released either way
}
