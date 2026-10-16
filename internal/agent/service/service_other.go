//go:build !windows

package service

import (
	"context"
	"errors"
)

// IsWindowsService reports whether the service manager started this
// process: never, off Windows.
func IsWindowsService() (bool, error) { return false, nil }

// RunWindowsService is Windows-only.
func RunWindowsService(func(context.Context) error) error {
	return errors.New("not a Windows service")
}
