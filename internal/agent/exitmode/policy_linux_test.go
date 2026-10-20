//go:build linux

package exitmode

import (
	"os"
	"strings"
	"syscall"
	"testing"

	"github.com/vishvananda/netlink"

	"github.com/hafio/vpn-mesh-utility/internal/agent/routes"
	"github.com/hafio/vpn-mesh-utility/internal/proto"
)

func TestKernelShapes(t *testing.T) {
	rules, rts := Desired("wg0", true, true)
	tunnel := toRule(rules[1])
	if tunnel.Protocol != routes.Protocol || !tunnel.Invert || tunnel.Mark != proto.FwMark || tunnel.Table != proto.RouteTable ||
		tunnel.Priority != PriorityTunnel || tunnel.SuppressPrefixlen != -1 {
		t.Fatalf("tunnel rule %+v", tunnel)
	}
	suppress := toRule(rules[0])
	if suppress.Invert || suppress.Mark != 0 || suppress.Table != MainTable || suppress.SuppressPrefixlen != 0 {
		t.Fatalf("suppress rule %+v", suppress)
	}
	if v6 := toRule(rules[2]); v6.Family != syscall.AF_INET6 {
		t.Fatalf("IPv6 rule family %d", v6.Family)
	}
	hole, err := toRoute(rts[1])
	if err != nil || hole.Type != syscall.RTN_BLACKHOLE || hole.Dst.String() != "::/0" || hole.Table != proto.RouteTable ||
		int(hole.Protocol) != routes.Protocol {
		t.Fatalf("blackhole %+v, %v", hole, err)
	}
	if _, err := toRoute(Route{Family: IPv4, Link: "mce-none0"}); err == nil || !strings.Contains(err.Error(), "mce-none0") {
		t.Fatalf("a route on a missing interface = %v", err)
	}
}

// Exit mode lists only its own rules: the container network's carry
// protocol 201 too, at other priorities, and must survive exit off.
func TestOwnsOnlyItsPriorities(t *testing.T) {
	for _, c := range []struct {
		priority, protocol int
		want               bool
	}{
		{PrioritySuppress, routes.Protocol, true},
		{PriorityTunnel, routes.Protocol, true},
		{1006, routes.Protocol, false},
		{5203, routes.Protocol, false},
		{PriorityTunnel, 4, false},
	} {
		r := netlink.NewRule()
		r.Priority, r.Protocol = c.priority, uint8(c.protocol)
		if got := owned(*r); got != c.want {
			t.Errorf("priority %d protocol %d: owned = %v, want %v", c.priority, c.protocol, got, c.want)
		}
	}
}

// Listing needs no privilege; changing anything needs CAP_NET_ADMIN, and the
// refusal says what was attempted. Skipped as root, where the calls would
// change the host (e2e covers that path).
func TestKernelPolicyWithoutPrivilege(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("as root this would change the host's policy routing")
	}
	p := NewPolicy()
	if _, err := p.Rules(); err != nil {
		t.Skipf("no netlink here: %v", err)
	}
	if _, err := p.Routes(); err != nil {
		t.Fatalf("Routes = %v", err)
	}
	want, routesWanted := Desired("lo", true, p.IPv6())
	for _, err := range []error{p.AddRule(want[1]), p.DelRule(want[1]), p.AddRoute(routesWanted[0]), p.DelRoute(routesWanted[0])} {
		if err == nil || (!strings.Contains(err.Error(), "table 51820") && !strings.Contains(err.Error(), "lookup 51820")) {
			t.Errorf("an unprivileged change = %v", err)
		}
	}
}
