//go:build !linux

package wg

import "errors"

// NewKernel is Linux-only. Windows gets the userspace backend in M5.
func NewKernel() (Host, error) {
	return nil, errors.New("kernel WireGuard is Linux-only; this platform uses the userspace backend (milestone M5 for Windows)")
}

// DetectBackend reports the only backend this platform can have.
func DetectBackend() string { return BackendUserspace }
