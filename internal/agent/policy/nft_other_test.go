//go:build !linux

package policy_test

import (
	"errors"
	"net/netip"
	"testing"

	"github.com/hafio/vpn-mesh-utility/internal/agent/policy"
	"github.com/hafio/vpn-mesh-utility/internal/testfix"
)

// Off Linux the kernel holds no table and refuses one.
func TestNoKernelOffLinux(t *testing.T) {
	k := policy.NewKernel()
	if mark, err := k.Installed(); mark != "" || err != nil {
		t.Fatalf("Installed = %q, %v", mark, err)
	}
	if err := k.Replace(&policy.Ruleset{}, "x"); !errors.Is(err, policy.ErrNoKernel) {
		t.Fatalf("Replace = %v", err)
	}
	if n, err := k.Forget(netip.MustParsePrefix(testfix.MeshCIDR)); n != 0 || err != nil || k.Remove() != nil {
		t.Fatalf("Forget = %d, %v", n, err)
	}
}
