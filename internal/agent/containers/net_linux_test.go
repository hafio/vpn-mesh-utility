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
	if h := blackhole(false); h.Type != syscall.RTN_BLACKHOLE || h.Table != BlackholeTable || h.Dst.String() != "0.0.0.0/0" ||
		int(h.Protocol) != routes.Protocol || h.Family != syscall.AF_INET {
		t.Fatalf("blackhole %+v", h)
	}
	// IPv6 (M8): the family, the unreachable rule, the IPv6 blackhole and
	// a proxy neighbour entry.
	r = toRule(Rule{Priority: 1002, IPv6: true, Iif: "wg0", Dst: netip.MustParsePrefix("fd10:77::8000:0/112"), Table: MainTable})
	if r.Family != syscall.AF_INET6 || r.Dst.String() != "fd10:77::8000:0/112" || r.Type != 0 {
		t.Fatalf("IPv6 rule %+v", r)
	}
	if r := toRule(Rule{Priority: 5204, IPv6: true, Iif: "meshctl0", Unreachable: true}); r.Type != syscall.RTN_UNREACHABLE || r.Table != 0 {
		t.Fatalf("unreachable rule %+v", r)
	}
	if h := blackhole(true); h.Dst.String() != "::/0" || h.Family != syscall.AF_INET6 || h.Type != syscall.RTN_BLACKHOLE {
		t.Fatalf("IPv6 blackhole %+v", h)
	}
	if p := proxy(7, netip.MustParseAddr("fd10:77::2:1")); p.LinkIndex != 7 || p.Family != syscall.AF_INET6 ||
		p.Flags&0x08 == 0 || p.IP.String() != "fd10:77::2:1" {
		t.Fatalf("proxy entry %+v", p)
	}
	if d := dropRoute(netip.MustParsePrefix("fd10:77::8000:0/112")); d.Table != MainTable || d.Type != syscall.RTN_BLACKHOLE ||
		d.Family != syscall.AF_INET6 || d.Dst.String() != "fd10:77::8000:0/112" || int(d.Protocol) != routes.Protocol || d.LinkIndex != 0 {
		t.Fatalf("drop route %+v", d)
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
	for _, v6 := range []bool{false, true} {
		if _, err := n.Blackhole(v6); err != nil {
			t.Fatalf("Blackhole(%v) = %v", v6, err)
		}
	}
	if _, err := n.IPv6Rules(); err != nil {
		t.Fatalf("IPv6Rules = %v", err)
	}
	if _, err := n.Dropped(); err != nil {
		t.Fatalf("Dropped = %v", err)
	}
	if _, err := n.Proxies("lo"); err != nil {
		t.Fatalf("Proxies(lo) = %v", err)
	}
	if _, err := n.Proxies("mce-none0"); err == nil || !strings.Contains(err.Error(), "mce-none0") {
		t.Fatalf("Proxies of a missing interface = %v", err)
	}
	for what, err := range map[string]error{
		"add a proxy entry to":    n.AddProxy("mce-none0", netip.MustParseAddr("fd10:77::1")),
		"remove a proxy entry of": n.DelProxy("mce-none0", netip.MustParseAddr("fd10:77::1")),
	} {
		if err == nil || !strings.Contains(err.Error(), "mce-none0") {
			t.Errorf("%s a missing interface = %v", what, err)
		}
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
		"add a rule":                n.AddRule(r),
		"remove a rule":             n.DelRule(r),
		"add an IPv6 rule":          n.AddRule(Rule{Priority: PriorityGuard + 4, IPv6: true, Iif: "lo", Unreachable: true}),
		"add the blackhole":         n.SetBlackhole(false, true),
		"remove the blackhole":      n.SetBlackhole(false, false),
		"add the IPv6 blackhole":    n.SetBlackhole(true, true),
		"remove the IPv6 blackhole": n.SetBlackhole(true, false),
		"add an address":            n.AddAddr("lo", netip.MustParsePrefix("127.0.0.9/8")),
		"add an IPv6 address":       n.AddAddr("lo", netip.MustParsePrefix("fd10:77::9/128")),
		"add a proxy entry":         n.AddProxy("lo", netip.MustParseAddr("fd10:77::1")),
		"drop a prefix":             n.SetDropped(netip.MustParsePrefix("fd10:77::8000:0/112"), true),
	} {
		if err == nil {
			t.Errorf("%s succeeded without privilege", what)
		}
	}
	if _, err := n.EnsureBridge("mce-none0"); err == nil || !strings.Contains(err.Error(), "create") {
		t.Errorf("EnsureBridge without privilege = %v", err)
	}
}
