//go:build !linux

package wg

import "errors"

// NewKernel is Linux-only; Windows runs the userspace backend on wintun.
func NewKernel() (Host, error) {
	return nil, errors.New("kernel WireGuard is Linux-only; this platform uses the userspace backend (--wg-backend userspace)")
}

// DetectBackend reports the only backend this platform can have.
func DetectBackend() string { return BackendUserspace }
