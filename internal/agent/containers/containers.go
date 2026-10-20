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

// Rule is one IPv4 policy rule: an input interface ("" any), a source and
// a destination (invalid: any), and the table it looks up.
type Rule struct {
	Priority int
	Iif      string
	Src, Dst netip.Prefix
	Table    int
}

func (r Rule) String() string {
	s := fmt.Sprintf("rule %d:", r.Priority)
	if r.Iif != "" {
		s += " iif " + r.Iif
	}
	if r.Src.IsValid() {
		s += " from " + r.Src.String()
	}
	if r.Dst.IsValid() {
		s += " to " + r.Dst.String()
	}
	return s + fmt.Sprintf(" lookup %d", r.Table)
}

// Net is the host as this package sees it: its policy rules, the
// blackhole table, the interfaces' addresses and the bridge.
type Net interface {
	// Rules returns the IPv4 rules with protocol 201 at an Owned priority.
	Rules() ([]Rule, error)
	AddRule(r Rule) error
	DelRule(r Rule) error
	// Blackhole reports whether BlackholeTable holds its blackhole default.
	Blackhole() (bool, error)
	SetBlackhole(on bool) error
	// LinkHolding is the interface, other than skip, that holds an IPv4
	// address inside p; "" when none does.
	LinkHolding(p netip.Prefix, skip string) (string, error)
	Addrs(link string) ([]netip.Prefix, error)
	AddAddr(link string, a netip.Prefix) error
	// EnsureBridge makes the bridge exist and be up, and reports whether
	// it had to create it.
	EnsureBridge(name string) (bool, error)
}

// Sysctls writes kernel parameters (internal/agent/sysctl).
type Sysctls interface {
	Apply(params []proto.Sysctl) (int, error)
}

// network is a container network's parsed addresses.
type network struct {
	subnet, mesh    netip.Prefix
	bridge, gateway netip.Addr
}

func parse(k proto.Containers) (network, error) {
	var (
		n    network
		errs []error
		err  error
	)
	keep := func(e error) {
		if e != nil {
			errs = append(errs, e)
		}
	}
	n.subnet, err = netip.ParsePrefix(k.Subnet)
	keep(err)
	n.mesh, err = netip.ParsePrefix(k.Network)
	keep(err)
	n.bridge, err = netip.ParseAddr(k.Bridge)
	keep(err)
	n.gateway, err = netip.ParseAddr(k.Gateway)
	keep(err)
	if err := errors.Join(errs...); err != nil {
		return network{}, fmt.Errorf("containers: %w", err)
	}
	return n, nil
}

func host(a netip.Addr) netip.Prefix { return netip.PrefixFrom(a, 32) }

// Gateway is the container-mode gateway's rules for k: wg the tunnel,
// meshIf the interface on the mesh network.
func Gateway(k proto.Containers, wg, meshIf string) ([]Rule, error) {
	n, err := parse(k)
	if err != nil {
		return nil, err
	}
	p := PriorityGateway
	rules := []Rule{
		{Priority: p, Iif: "lo", Table: MainTable},
		{Priority: p + 1, Iif: wg, Dst: host(n.bridge), Table: BlackholeTable},
		{Priority: p + 2, Iif: wg, Dst: n.subnet, Table: MainTable},
	}
	if k.Relay {
		rules = append(rules, Rule{Priority: p + 3, Iif: wg, Dst: n.mesh, Table: MainTable})
	}
	return append(rules,
		Rule{Priority: p + 4, Iif: meshIf, Src: host(n.bridge), Table: BlackholeTable},
		Rule{Priority: p + 5, Iif: meshIf, Src: n.subnet, Dst: n.mesh, Table: MainTable},
		Rule{Priority: p + 6, Table: BlackholeTable},
	), nil
}

// Guard is a host-mode node's rules for k: its subnet is reached from wg,
// the bridge and the host itself, and from nowhere else.
func Guard(k proto.Containers, wg string) ([]Rule, error) {
	n, err := parse(k)
	if err != nil {
		return nil, err
	}
	p := PriorityGuard
	return []Rule{
		{Priority: p, Iif: wg, Dst: n.subnet, Table: MainTable},
		{Priority: p + 1, Iif: proto.ContainerBridge, Table: MainTable},
		{Priority: p + 2, Iif: "lo", Dst: n.subnet, Table: MainTable},
		{Priority: p + 3, Dst: n.subnet, Table: BlackholeTable},
	}, nil
}

// Sync makes the host match k -- nil takes the rules and the blackhole
// route away; addresses and the bridge stay, since a runtime may still use
// them -- and returns how many changes it made. wg is the tunnel.
func Sync(nt Net, sys Sysctls, k *proto.Containers, wg string) (int, error) {
	if k == nil {
		return syncRules(nt, nil)
	}
	n, err := parse(*k)
	if err != nil {
		return 0, err
	}
	changes := 0
	var want []Rule
	switch k.Mode {
	case proto.ModeContainer:
		meshIf, err := nt.LinkHolding(n.subnet, wg)
		if err != nil {
			return 0, fmt.Errorf("containers: %w", err)
		}
		if meshIf == "" {
			return 0, fmt.Errorf("containers: no interface holds an address of %s, so the agent is not on the %s network: "+
				"create it with the commands in the compose file's header (`meshctl node compose`), then recreate the agent's container",
				n.subnet, proto.MeshNetwork)
		}
		c, err := ensureAddr(nt, meshIf, netip.PrefixFrom(n.gateway, n.mesh.Bits()))
		changes += c
		if err != nil {
			return changes, err
		}
		if want, err = Gateway(*k, wg, meshIf); err != nil {
			return changes, err
		}
	case proto.ModeHost:
		created, err := nt.EnsureBridge(proto.ContainerBridge)
		if err != nil {
			return 0, fmt.Errorf("containers: bridge %s: %w", proto.ContainerBridge, err)
		}
		if created {
			changes++
		}
		c, err := ensureAddr(nt, proto.ContainerBridge, netip.PrefixFrom(n.bridge, n.mesh.Bits()))
		changes += c
		if err != nil {
			return changes, err
		}
		params := []proto.Sysctl{{Key: proto.SysctlBridgeForwarding, Value: "1"}, {Key: proto.SysctlBridgeProxyARP, Value: "1"},
			{Key: proto.SysctlWGForwarding, Value: "1"}}
		c, err = sys.Apply(params)
		changes += c
		if err != nil {
			return changes, fmt.Errorf("containers: %w", err)
		}
		if want, err = Guard(*k, wg); err != nil {
			return changes, err
		}
	default:
		return 0, fmt.Errorf("containers: unknown mode %q", k.Mode)
	}
	c, err := syncRules(nt, want)
	return changes + c, err
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

// syncRules makes the agent's rules exactly want: the blackhole route
// before any rule that looks it up, and rules removed before the route
// goes.
func syncRules(nt Net, want []Rule) (int, error) {
	have, err := nt.Rules()
	if err != nil {
		return 0, fmt.Errorf("containers: %w", err)
	}
	hole, err := nt.Blackhole()
	if err != nil {
		return 0, fmt.Errorf("containers: %w", err)
	}
	n := 0
	if len(want) > 0 && !hole {
		if err := nt.SetBlackhole(true); err != nil {
			return n, fmt.Errorf("containers: %w", err)
		}
		n++
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
	if len(want) == 0 && hole {
		if err := nt.SetBlackhole(false); err != nil {
			return n, fmt.Errorf("containers: %w", err)
		}
		n++
	}
	return n, nil
}
