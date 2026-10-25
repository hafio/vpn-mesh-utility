//go:build !linux

package policy

import (
	"net/netip"

	"github.com/hafio/vpn-mesh-utility/internal/proto"
)

// NewKernel is a kernel that holds no table: servers, which enforce the
// mesh firewall, run on Linux.
func NewKernel(proto.Slot) Kernel { return noKernel{} }

type noKernel struct{}

func (noKernel) Installed() (string, error)       { return "", nil }
func (noKernel) Replace(*Ruleset, string) error   { return ErrNoKernel }
func (noKernel) Remove() error                    { return nil }
func (noKernel) Forget(netip.Prefix) (int, error) { return 0, nil }
