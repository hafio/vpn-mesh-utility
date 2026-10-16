package wg

import (
	"log/slog"

	"github.com/hafio/vpn-mesh-utility/internal/validate"
)

// The WireGuard backends (spec S7.1): the kernel module where the host has
// it, else wireguard-go inside the agent behind a tun interface (a NAS
// without the module, a container).
const (
	BackendAuto      = "auto"
	BackendKernel    = "kernel"
	BackendUserspace = "userspace"
)

// Open opens the backend asked for (`agent run --wg-backend`) and says which
// one it opened; auto picks the kernel when it is available (DetectBackend).
func Open(backend string, log *slog.Logger) (Host, string, error) {
	backend, err := validate.Enum("--wg-backend", backend, BackendAuto, BackendKernel, BackendUserspace)
	if err != nil {
		return nil, "", err
	}
	if backend == BackendAuto {
		backend = DetectBackend()
	}
	var h Host
	if backend == BackendKernel {
		h, err = NewKernel()
	} else {
		h, err = NewUserspace(log)
	}
	if err != nil {
		return nil, "", err
	}
	return h, backend, nil
}
