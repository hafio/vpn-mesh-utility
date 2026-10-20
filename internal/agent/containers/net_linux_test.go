//go:build linux

package containers

import (
	"net/netip"
	"os"
	"strings"
	"syscall"
	"testing"

	"github.com/hafio/vpn-mesh-utility/internal/agent/routes"
)

func TestKernelShapes(t *testing.T) {
	r := toRule(Rule{Priority: 1001, Iif: "wg0", Dst: netip.MustParsePrefix("10.77.128.1/32"), Table: BlackholeTable})
	if r.Protocol != routes.Protocol || r.Priority != 1001 || r.IifName != "wg0" || r.Table != BlackholeTable ||
		r.Src != nil || r.Dst == nil || r.Dst.String() != "10.77.128.1/32" || r.Family != syscall.AF_INET {
		t.Fatalf("rule %+v", r)
	}
	if r := toRule(Rule{Priority: 1005, Src: netip.MustParsePrefix("10.77.128.0/24"), Table: MainTable}); r.Src.String() != "10.77.128.0/24" || r.Dst != nil {
		t.Fatalf("rule %+v", r)
	}
	if h := blackhole(); h.Type != syscall.RTN_BLACKHOLE || h.Table != BlackholeTable || h.Dst.String() != "0.0.0.0/0" ||
		int(h.Protocol) != routes.Protocol {
		t.Fatalf("blackhole %+v", h)
	}
}

// Listing needs no privilege; changing anything needs CAP_NET_ADMIN, and the
// refusal says what was attempted. Skipped as root, where the calls would
// change the host (e2e covers that path).
func TestKernelNetWithoutPrivilege(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("as root this would change the host's policy routing")
	}
	n := NewNet()
	if _, err := n.Rules(); err != nil {
		t.Skipf("no netlink here: %v", err)
	}
	if _, err := n.Blackhole(); err != nil {
		t.Fatalf("Blackhole = %v", err)
	}
	if link, err := n.LinkHolding(netip.MustParsePrefix("127.0.0.0/8"), ""); err != nil || link != "lo" {
		t.Fatalf("LinkHolding(127/8) = %q, %v", link, err)
	}
	if link, err := n.LinkHolding(netip.MustParsePrefix("127.0.0.0/8"), "lo"); err != nil || link != "" {
		t.Fatalf("LinkHolding skipping lo = %q, %v", link, err)
	}
	if addrs, err := n.Addrs("lo"); err != nil || len(addrs) == 0 {
		t.Fatalf("Addrs(lo) = %v, %v", addrs, err)
	}
	if _, err := n.Addrs("mce-none0"); err == nil || !strings.Contains(err.Error(), "mce-none0") {
		t.Fatalf("Addrs of a missing interface = %v", err)
	}
	if err := n.AddAddr("mce-none0", netip.MustParsePrefix("10.77.128.2/16")); err == nil || !strings.Contains(err.Error(), "mce-none0") {
		t.Fatalf("AddAddr to a missing interface = %v", err)
	}
	if _, err := n.EnsureBridge("lo"); err == nil || !strings.Contains(err.Error(), "not a bridge") {
		t.Fatalf("EnsureBridge(lo) = %v", err)
	}
	r := Rule{Priority: PriorityGateway, Iif: "lo", Table: MainTable}
	for what, err := range map[string]error{
		"add a rule":           n.AddRule(r),
		"remove a rule":        n.DelRule(r),
		"add the blackhole":    n.SetBlackhole(true),
		"remove the blackhole": n.SetBlackhole(false),
		"add an address":       n.AddAddr("lo", netip.MustParsePrefix("127.0.0.9/8")),
	} {
		if err == nil {
			t.Errorf("%s succeeded without privilege", what)
		}
	}
	if _, err := n.EnsureBridge("mce-none0"); err == nil || !strings.Contains(err.Error(), "create") {
		t.Errorf("EnsureBridge without privilege = %v", err)
	}
}
