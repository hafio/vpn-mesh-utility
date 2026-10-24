// Package containers keeps a node's container network (amendment A37) on
// the host.
//
// In container mode the agent is its containers' gateway, in its own
// network namespace: it holds the gateway address (.2) on the interface the
// runtime attached to the mesh network -- the runtime gave it another one,
// or none it could choose, on a swarm-scoped network -- and isolates by
// policy routing (checkpoint 6a's trial proved it on Docker, Podman and
// Swarm):
//
//	1000: iif lo lookup main                      the agent's own traffic
//	1001: iif wg0 to <bridge>/32 lookup 51821     the mesh never reaches the host
//	1002: iif wg0 to <subnet> lookup main         the mesh reaches the containers
//	1003: iif wg0 to <mesh> lookup main           a hub relays (Relay only)
//	1004: iif <mesh side> from <bridge>/32 lookup 51821  the host never reaches the mesh
//	1005: iif <mesh side> from <subnet> to <mesh> lookup main  containers reach the mesh
//	1006: lookup 51821                            anything else forwarded: dropped
//	table 51821: blackhole default
//
// Local delivery (the local table, priority 0) comes first, so the
// resolver and WireGuard keep answering; the host's ARP for a mesh address
// meets rule 1004, so the gateway does not even answer it.
//
// In host mode the host is on the mesh and routes the subnet itself: the
// agent owns the bridge meshctl0 the runtime attaches to (it makes it when
// it is missing, holds the bridge address, and turns on the bridge's
// forwarding and proxy ARP and wg0's forwarding), and four rules keep the
// subnet to the mesh -- a runtime that lets anyone who routes to the host
// reach its containers (Docker's nat-unprotected) must not open them to the
// host's LAN:
//
//	5200: iif wg0 to <subnet> lookup main
//	5201: iif meshctl0 lookup main
//	5202: iif lo to <subnet> lookup main
//	5203: to <subnet> lookup 51821
//
// With the containers' IPv6 on the mesh (M8, opt-in) the same rules hold
// for IPv6 (ip -6 rule, a blackhole default of its own in table 51821),
// with the node's /112 and the mesh's IPv6 prefix; in host mode the bridge
// also gets
//
//	ipv6 5201: iif meshctl0 to <mesh6> lookup main
//	ipv6 5204: iif meshctl0 unreachable
//
// in place of the IPv4 5201: a runtime that makes the mesh bridge its
// containers' IPv6 default route (Docker) sends them their IPv6 internet
// there, and an unreachable at once lets them fall back to IPv4 instead of
// waiting. A container sees the whole mesh prefix on its link, and IPv6
// has no proxy ARP: the agent keeps one proxy neighbour entry per far mesh
// address on the mesh side (Proxies), so the gateway answers neighbour
// discovery for them. A kernel without IPv6 policy routing cannot isolate
// IPv6: the containers then get IPv4 only -- no IPv6 address, rule or
// entry of the agent's, and a blackhole route for the /112 in the main
// table (Dropped), so whatever the runtime set up carries nothing between
// them and the mesh -- and Sync says so (ErrIPv6Unrouted).
//
// Every rule and route here carries protocol 201 and a priority of its
// own, so Sync lists, and removes, only these.
package containers

import (
	"errors"
	"fmt"
	"net/netip"
	"slices"

	"github.com/hafio/vpn-mesh-utility/internal/proto"
)

// Tables a rule looks up.
const (
	MainTable = 254
	// BlackholeTable holds only a blackhole default route.
	BlackholeTable = 51821
)

// The priorities of the gateway's rules (container mode) and of the guard
// (host mode); each set is consecutive from its base.
const (
	PriorityGateway = 1000
	PriorityGuard   = 5200
	rulesPerSet     = 7
)

// Owned reports whether a rule at priority is one of this package's.
func Owned(priority int) bool {
	return (priority >= PriorityGateway && priority < PriorityGateway+rulesPerSet) ||
		(priority >= PriorityGuard && priority < PriorityGuard+rulesPerSet)
}

// Rule is one policy rule, of IPv4 or of IPv6: an input interface (""
// any), a source and a destination (invalid: any), and the table it looks
// up -- or, Unreachable, none: the packet is refused at once.
type Rule struct {
	Priority    int
	IPv6        bool
	Iif         string
	Src, Dst    netip.Prefix
	Table       int
	Unreachable bool
}

func (r Rule) String() string {
	s := fmt.Sprintf("rule %d:", r.Priority)
	if r.IPv6 {
		s = "ipv6 " + s
	}
	if r.Iif != "" {
		s += " iif " + r.Iif
	}
	if r.Src.IsValid() {
		s += " from " + r.Src.String()
	}
	if r.Dst.IsValid() {
		s += " to " + r.Dst.String()
	}
	if r.Unreachable {
		return s + " unreachable"
	}
	return s + fmt.Sprintf(" lookup %d", r.Table)
}

// Net is the host as this package sees it: its policy rules, the
// blackhole table, the interfaces' addresses and proxy neighbour entries,
// the bridge, and the main table's drops.
type Net interface {
	// Rules returns the rules of both families with protocol 201 at an
	// Owned priority (IPv4 ones only without IPv6 policy routing).
	Rules() ([]Rule, error)
	AddRule(r Rule) error
	DelRule(r Rule) error
	// IPv6Rules reports whether the kernel has IPv6 policy routing.
	IPv6Rules() (bool, error)
	// Blackhole reports whether BlackholeTable holds the family's
	// blackhole default.
	Blackhole(ipv6 bool) (bool, error)
	SetBlackhole(ipv6, on bool) error
	// LinkHolding is the interface, other than skip, that holds an
	// address inside p; "" when none does.
	LinkHolding(p netip.Prefix, skip string) (string, error)
	// Addrs is the interface's addresses of both families, link-local ones
	// aside.
	Addrs(link string) ([]netip.Prefix, error)
	// AddAddr adds an address (an IPv6 one without duplicate address
	// detection: it is the mesh's own).
	AddAddr(link string, a netip.Prefix) error
	// EnsureBridge makes the bridge exist and be up, and reports whether
	// it had to create it.
	EnsureBridge(name string) (bool, error)
	// Proxies is the interface's IPv6 proxy neighbour entries: the
	// addresses it answers neighbour discovery for, not being them.
	Proxies(link string) ([]netip.Addr, error)
	AddProxy(link string, a netip.Addr) error
	DelProxy(link string, a netip.Addr) error
	// Dropped is the main table's IPv6 blackhole routes of ours: prefixes
	// the host drops, whatever routes them there.
	Dropped() ([]netip.Prefix, error)
	SetDropped(p netip.Prefix, on bool) error
}

// Sysctls writes kernel parameters (internal/agent/sysctl).
type Sysctls interface {
	Apply(params []proto.Sysctl) (int, error)
}

// network is a container network's parsed addresses; the IPv6 side is
// invalid when the containers have no IPv6.
type network struct {
	subnet, mesh, subnet6, mesh6       netip.Prefix
	bridge, gateway, bridge6, gateway6 netip.Addr
}

func parse(k proto.Containers) (network, error) {
	var (
		n    network
		errs []error
	)
	prefix := func(s string) netip.Prefix {
		p, err := netip.ParsePrefix(s)
		if err != nil {
			errs = append(errs, err)
		}
		return p
	}
	addr := func(s string) netip.Addr {
		a, err := netip.ParseAddr(s)
		if err != nil {
			errs = append(errs, err)
		}
		return a
	}
	n.subnet, n.mesh, n.bridge, n.gateway = prefix(k.Subnet), prefix(k.Network), addr(k.Bridge), addr(k.Gateway)
	if k.Subnet6 != "" {
		n.subnet6, n.mesh6, n.bridge6, n.gateway6 = prefix(k.Subnet6), prefix(k.Network6), addr(k.Bridge6), addr(k.Gateway6)
	}
	if err := errors.Join(errs...); err != nil {
		return network{}, fmt.Errorf("containers: %w", err)
	}
	return n, nil
}

func host(a netip.Addr) netip.Prefix { return netip.PrefixFrom(a, a.BitLen()) }

// Gateway is the container-mode gateway's rules for k: wg the tunnel,
// meshIf the interface on the mesh network.
func Gateway(k proto.Containers, wg, meshIf string) ([]Rule, error) {
	n, err := parse(k)
	if err != nil {
		return nil, err
	}
	rules := gateway(false, n.subnet, n.mesh, n.bridge, wg, meshIf, k.Relay)
	if n.subnet6.IsValid() {
		rules = append(rules, gateway(true, n.subnet6, n.mesh6, n.bridge6, wg, meshIf, k.Relay)...)
	}
	return rules, nil
}

// gateway is one family's gateway rules.
func gateway(v6 bool, subnet, mesh netip.Prefix, bridge netip.Addr, wg, meshIf string, relay bool) []Rule {
	p := PriorityGateway
	rules := []Rule{
		{Priority: p, IPv6: v6, Iif: "lo", Table: MainTable},
		{Priority: p + 1, IPv6: v6, Iif: wg, Dst: host(bridge), Table: BlackholeTable},
		{Priority: p + 2, IPv6: v6, Iif: wg, Dst: subnet, Table: MainTable},
	}
	if relay {
		rules = append(rules, Rule{Priority: p + 3, IPv6: v6, Iif: wg, Dst: mesh, Table: MainTable})
	}
	return append(rules,
		Rule{Priority: p + 4, IPv6: v6, Iif: meshIf, Src: host(bridge), Table: BlackholeTable},
		Rule{Priority: p + 5, IPv6: v6, Iif: meshIf, Src: subnet, Dst: mesh, Table: MainTable},
		Rule{Priority: p + 6, IPv6: v6, Table: BlackholeTable},
	)
}

// Guard is a host-mode node's rules for k: its subnet is reached from wg,
// the bridge and the host itself, and from nowhere else; IPv6 from the
// bridge goes to the mesh or is refused.
func Guard(k proto.Containers, wg string) ([]Rule, error) {
	n, err := parse(k)
	if err != nil {
		return nil, err
	}
	p, br := PriorityGuard, proto.ContainerBridge
	rules := []Rule{
		{Priority: p, Iif: wg, Dst: n.subnet, Table: MainTable},
		{Priority: p + 1, Iif: br, Table: MainTable},
		{Priority: p + 2, Iif: "lo", Dst: n.subnet, Table: MainTable},
		{Priority: p + 3, Dst: n.subnet, Table: BlackholeTable},
	}
	if n.subnet6.IsValid() {
		rules = append(rules,
			Rule{Priority: p, IPv6: true, Iif: wg, Dst: n.subnet6, Table: MainTable},
			Rule{Priority: p + 1, IPv6: true, Iif: br, Dst: n.mesh6, Table: MainTable},
			Rule{Priority: p + 2, IPv6: true, Iif: "lo", Dst: n.subnet6, Table: MainTable},
			Rule{Priority: p + 3, IPv6: true, Dst: n.subnet6, Table: BlackholeTable},
			Rule{Priority: p + 4, IPv6: true, Iif: br, Unreachable: true},
		)
	}
	return rules, nil
}

// ErrIPv6Unrouted is why a container network carries IPv4 only although
// its config has IPv6: the kernel has no IPv6 policy routing, which
// isolates the containers' IPv6 (M8).
var ErrIPv6Unrouted = errors.New("containers: the kernel has no IPv6 policy routing (CONFIG_IPV6_MULTIPLE_TABLES), " +
	"which isolates the containers' IPv6, so their network carries IPv4 only; " +
	"take their IPv6 off the mesh (`meshctl node set <name> --containers6=false`)")

// Sync makes the host match k -- nil takes the rules and the blackhole
// routes away; addresses, proxy entries and the bridge stay, since a
// runtime may still use them, and so do the IPv6 ones when k has no IPv6
// -- and returns how many changes it made. wg is the tunnel; proxied are
// the IPv6 mesh addresses the mesh side answers neighbour discovery for
// (the node's own and those it routes over wg), of which those of the
// mesh's prefix outside the node's /112 are kept. Without IPv6 policy
// routing the IPv4 side is set up, the IPv6 side left out and the /112
// dropped, and the error is ErrIPv6Unrouted.
func Sync(nt Net, sys Sysctls, k *proto.Containers, wg string, proxied []netip.Addr) (int, error) {
	if k == nil {
		return syncRules(nt, nil)
	}
	n, err := parse(*k)
	if err != nil {
		return 0, err
	}
	changes := 0
	var (
		want        []Rule
		link        string // the mesh side
		addr6       netip.Addr
		sysctls     []proto.Sysctl
		errUnrouted error
	)
	switch k.Mode {
	case proto.ModeContainer:
		if link, err = nt.LinkHolding(n.subnet, wg); err != nil {
			return 0, fmt.Errorf("containers: %w", err)
		}
		if link == "" {
			return 0, fmt.Errorf("containers: no interface holds an address of %s, so the agent is not on the %s network: "+
				"create it with the node's values (`meshctl node show <name>`; docs/containers.md), then recreate the agent's container",
				n.subnet, proto.MeshNetwork)
		}
		c, err := ensureAddr(nt, link, netip.PrefixFrom(n.gateway, n.mesh.Bits()))
		changes += c
		if err != nil {
			return changes, err
		}
		if want, err = Gateway(*k, wg, link); err != nil {
			return changes, err
		}
		addr6 = n.gateway6
	case proto.ModeHost:
		link = proto.ContainerBridge
		created, err := nt.EnsureBridge(link)
		if err != nil {
			return 0, fmt.Errorf("containers: bridge %s: %w", link, err)
		}
		if created {
			changes++
		}
		c, err := ensureAddr(nt, link, netip.PrefixFrom(n.bridge, n.mesh.Bits()))
		changes += c
		if err != nil {
			return changes, err
		}
		sysctls = []proto.Sysctl{{Key: proto.SysctlBridgeForwarding, Value: "1"}, {Key: proto.SysctlBridgeProxyARP, Value: "1"},
			{Key: proto.SysctlWGForwarding, Value: "1"}}
		if want, err = Guard(*k, wg); err != nil {
			return changes, err
		}
		addr6 = n.bridge6
	default:
		return 0, fmt.Errorf("containers: unknown mode %q", k.Mode)
	}
	var (
		far  []netip.Addr
		drop []netip.Prefix
	)
	if n.subnet6.IsValid() {
		routed, err := nt.IPv6Rules()
		if err != nil {
			return changes, fmt.Errorf("containers: %w", err)
		}
		if routed {
			c, err := ensureAddr(nt, link, netip.PrefixFrom(addr6, n.mesh6.Bits()))
			changes += c
			if err != nil {
				return changes, err
			}
			if k.Mode == proto.ModeHost {
				sysctls = append(sysctls, proto.Sysctl{Key: proto.SysctlBridgeProxyNDP, Value: "1"})
			}
			for _, a := range proxied {
				if n.mesh6.Contains(a) && !n.subnet6.Contains(a) {
					far = append(far, a)
				}
			}
		} else {
			want = slices.DeleteFunc(want, func(r Rule) bool { return r.IPv6 })
			drop = []netip.Prefix{n.subnet6}
			errUnrouted = ErrIPv6Unrouted
		}
	}
	if len(sysctls) > 0 {
		c, err := sys.Apply(sysctls)
		changes += c
		if err != nil {
			return changes, fmt.Errorf("containers: %w", err)
		}
	}
	c, err := syncRules(nt, want)
	changes += c
	if err != nil {
		return changes, err
	}
	if n.subnet6.IsValid() {
		c, err = syncDropped(nt, drop)
		changes += c
		if err != nil {
			return changes, err
		}
		c, err = syncProxies(nt, link, far)
		changes += c
		if err != nil {
			return changes, err
		}
	}
	return changes, errUnrouted
}

// syncDropped makes the agent's main-table blackhole routes exactly want.
func syncDropped(nt Net, want []netip.Prefix) (int, error) {
	have, err := nt.Dropped()
	if err != nil {
		return 0, fmt.Errorf("containers: %w", err)
	}
	n := 0
	for _, p := range want {
		if !slices.Contains(have, p) {
			if err := nt.SetDropped(p, true); err != nil {
				return n, fmt.Errorf("containers: drop %s: %w", p, err)
			}
			n++
		}
	}
	for _, p := range have {
		if !slices.Contains(want, p) {
			if err := nt.SetDropped(p, false); err != nil {
				return n, fmt.Errorf("containers: stop dropping %s: %w", p, err)
			}
			n++
		}
	}
	return n, nil
}

// syncProxies makes the link's proxy neighbour entries exactly want.
func syncProxies(nt Net, link string, want []netip.Addr) (int, error) {
	have, err := nt.Proxies(link)
	if err != nil {
		return 0, fmt.Errorf("containers: proxy entries of %s: %w", link, err)
	}
	n := 0
	for _, a := range want {
		if !slices.Contains(have, a) {
			if err := nt.AddProxy(link, a); err != nil {
				return n, fmt.Errorf("containers: answer neighbour discovery for %s on %s: %w", a, link, err)
			}
			have = append(have, a) // once, if want repeats it
			n++
		}
	}
	for _, a := range have {
		if !slices.Contains(want, a) {
			if err := nt.DelProxy(link, a); err != nil {
				return n, fmt.Errorf("containers: stop answering neighbour discovery for %s on %s: %w", a, link, err)
			}
			n++
		}
	}
	return n, nil
}

func ensureAddr(nt Net, link string, a netip.Prefix) (int, error) {
	have, err := nt.Addrs(link)
	if err != nil {
		return 0, fmt.Errorf("containers: addresses of %s: %w", link, err)
	}
	if slices.Contains(have, a) {
		return 0, nil
	}
	if err := nt.AddAddr(link, a); err != nil {
		return 0, fmt.Errorf("containers: add %s to %s: %w", a, link, err)
	}
	return 1, nil
}

// syncRules makes the agent's rules exactly want: a family's blackhole
// route before any rule that looks it up, and rules removed before the
// route goes. IPv6's route is looked at only while it has rules, so a host
// without IPv6 policy routing is never asked about it.
func syncRules(nt Net, want []Rule) (int, error) {
	have, err := nt.Rules()
	if err != nil {
		return 0, fmt.Errorf("containers: %w", err)
	}
	of := func(rs []Rule, v6 bool) bool {
		return slices.ContainsFunc(rs, func(r Rule) bool { return r.IPv6 == v6 })
	}
	hole := map[bool]bool{}
	for _, v6 := range []bool{false, true} {
		if v6 && !of(have, true) && !of(want, true) {
			continue
		}
		if hole[v6], err = nt.Blackhole(v6); err != nil {
			return 0, fmt.Errorf("containers: %w", err)
		}
	}
	n := 0
	for _, v6 := range []bool{false, true} {
		if of(want, v6) && !hole[v6] {
			if err := nt.SetBlackhole(v6, true); err != nil {
				return n, fmt.Errorf("containers: %w", err)
			}
			n++
		}
	}
	for _, r := range want {
		if !slices.Contains(have, r) {
			if err := nt.AddRule(r); err != nil {
				return n, fmt.Errorf("containers: %w", err)
			}
			n++
		}
	}
	for _, r := range have {
		if !slices.Contains(want, r) {
			if err := nt.DelRule(r); err != nil {
				return n, fmt.Errorf("containers: %w", err)
			}
			n++
		}
	}
	for _, v6 := range []bool{false, true} {
		if !of(want, v6) && hole[v6] {
			if err := nt.SetBlackhole(v6, false); err != nil {
				return n, fmt.Errorf("containers: %w", err)
			}
			n++
		}
	}
	return n, nil
}
