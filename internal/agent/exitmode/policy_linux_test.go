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
	rules, rts := Desired(0, "wg0", true, true, false)
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
	// Another slot's (amendment A49): its priorities and table, the same mark.
	rules, rts = Desired(2, "wg2", true, false, false)
	if tunnel := toRule(rules[1]); tunnel.Priority != PriorityTunnel-20 || tunnel.Table != proto.RouteTable+4 || tunnel.Mark != proto.FwMark {
		t.Fatalf("slot 2's tunnel rule %+v", tunnel)
	}
	if r, err := toRoute(Route{Family: IPv6, Table: rts[0].Table}); err != nil || r.Table != proto.RouteTable+4 {
		t.Fatalf("slot 2's route %+v, %v", r, err)
	}
}

// Exit mode lists every slot's exit rules and no other: the container
// network's carry protocol 201 too, at other priorities, and must survive
// exit off.
func TestOwnsOnlyItsPriorities(t *testing.T) {
	for _, c := range []struct {
		priority, protocol int
		want               bool
	}{
		{PrioritySuppress, routes.Protocol, true},
		{PriorityTunnel, routes.Protocol, true},
		{PriorityTunnel - 10, routes.Protocol, true},
		{PrioritySuppress - 630, routes.Protocol, true},
		{PriorityTunnel - 2, routes.Protocol, false},
		{PriorityTunnel - 640, routes.Protocol, false},
		{32766, routes.Protocol, false},
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
	if _, err := p.Routes(proto.RouteTable); err != nil {
		t.Fatalf("Routes = %v", err)
	}
	want, routesWanted := Desired(0, "lo", true, p.IPv6() == V6Policy, false)
	for _, err := range []error{p.AddRule(want[1]), p.DelRule(want[1]), p.AddRoute(routesWanted[0]), p.DelRoute(routesWanted[0])} {
		if err == nil || (!strings.Contains(err.Error(), "table 51820") && !strings.Contains(err.Error(), "lookup 51820")) {
			t.Errorf("an unprivileged change = %v", err)
		}
	}
}
