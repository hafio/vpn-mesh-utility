package planner

import (
	"fmt"
	"net/netip"
	"sort"
	"strings"

	"github.com/hafio/vpn-mesh-utility/internal/proto"
	"github.com/hafio/vpn-mesh-utility/internal/validate"
)

// Problem is one finding about the desired state. Errors block a change;
// warnings are shown to the operator and the plan still applies.
type Problem struct {
	Node    string
	Message string
	Warning bool
}

func (p Problem) String() string {
	level := "error"
	if p.Warning {
		level = "warning"
	}
	if p.Node == "" {
		return level + ": " + p.Message
	}
	return fmt.Sprintf("%s: node %q: %s", level, p.Node, p.Message)
}

// Check validates the desired state as a whole: the constraints that span
// nodes (uniqueness, references, roles) and the per-node values a computed
// config would carry.
func Check(in Input) []Problem {
	var out []Problem
	errf := func(node, format string, args ...any) {
		out = append(out, Problem{Node: node, Message: fmt.Sprintf(format, args...)})
	}
	warnf := func(node, format string, args ...any) {
		out = append(out, Problem{Node: node, Message: fmt.Sprintf(format, args...), Warning: true})
	}

	if err := CheckMeshCIDR(in.MeshCIDR); err != nil {
		errf("", "%v", err)
	}
	if in.MeshCIDR6.IsValid() {
		if err := CheckMeshCIDR6(in.MeshCIDR6); err != nil {
			errf("", "%v", err)
		}
	}
	byID := map[int64]Node{}
	names := map[string]bool{}
	ips := map[string]string{}
	keys := map[string]string{}
	for _, n := range in.Nodes {
		byID[n.ID] = n
		if _, err := validate.Name("name", n.Name); err != nil {
			errf(n.Name, "%v", err)
		}
		if names[n.Name] {
			errf(n.Name, "the name is used by two nodes")
		}
		names[n.Name] = true
		switch n.Role {
		case proto.RoleServer, proto.RoleSpoke, proto.RolePhone:
		default:
			errf(n.Name, "unknown role %q (want server, spoke or phone)", n.Role)
		}
		switch {
		case !n.MeshIP.IsValid() && !n.MeshIP6.IsValid():
			errf(n.Name, "has no mesh address, neither IPv4 nor IPv6")
		case n.MeshIP.IsValid() && !n.MeshIP.Is4():
			errf(n.Name, "mesh IP %s is not an IPv4 address", n.MeshIP)
		case n.MeshIP.IsValid():
			if in.MeshCIDR.IsValid() && !in.MeshCIDR.Contains(n.MeshIP) {
				errf(n.Name, "mesh IP %s is outside the mesh %s", n.MeshIP, in.MeshCIDR)
			}
			if other, dup := ips[n.MeshIP.String()]; dup {
				errf(n.Name, "mesh IP %s is also assigned to %q", n.MeshIP, other)
			}
			ips[n.MeshIP.String()] = n.Name
		}
		checkNode6(in, n, ips, errf)
		checkUsable(n, warnf)
		if n.WGPubKey != "" {
			if other, dup := keys[n.WGPubKey]; dup {
				errf(n.Name, "has the same WireGuard key as %q", other)
			}
			keys[n.WGPubKey] = n.Name
		}
		if n.MTU < proto.MinMTU || n.MTU > proto.MaxMTU {
			errf(n.Name, "MTU %d is outside %d-%d", n.MTU, proto.MinMTU, proto.MaxMTU)
		}
		if n.Keepalive < 0 || n.Keepalive > 3600 {
			errf(n.Name, "keepalive %d is outside 0-3600 seconds", n.Keepalive)
		}
		if n.Role == proto.RoleServer && (n.ListenPort < 1 || n.ListenPort > 65535) {
			errf(n.Name, "listen port %d is outside 1-65535", n.ListenPort)
		}
		if n.CanExit && n.Role != proto.RoleServer {
			errf(n.Name, "only servers can be an exit")
		}
		switch n.Mode {
		case "", proto.ModeHost, proto.ModeContainer:
		default:
			errf(n.Name, "unknown mode %q (want host or container)", n.Mode)
		}
		switch {
		case n.Role == proto.RolePhone && (n.inContainer() || n.Containers.IsValid()):
			errf(n.Name, "a phone has no agent, so it has no container network")
		case n.inContainer() && !n.Containers.IsValid():
			errf(n.Name, "is in container mode but has no container subnet")
		case n.inContainer() && n.CanExit:
			errf(n.Name, "a container-mode server cannot be an exit: its host is not on the mesh; revoke can-exit, or run it in host mode")
		}
	}
	checkContainers(in, errf)
	checkContainers6(in, errf)

	isServer := func(id int64) bool {
		s, ok := byID[id]
		return ok && s.Role == proto.RoleServer
	}
	if in.DefaultHub != 0 && !isServer(in.DefaultHub) {
		errf("", "the default hub (node id %d) is not a server", in.DefaultHub)
	}
	checkDNS(in, names, byID, errf)

	// What follows judges the mesh as it is planned: without the families
	// the hosts cannot use.
	in = in.usable()
	p := newPlan(in)
	checkSegmentation(in, p.servers, errf, warnf)
	for _, n := range sortedByID(in.Nodes) {
		if n.PrimaryHub != 0 && !isServer(n.PrimaryHub) {
			errf(n.Name, "primary hub (node id %d) is not a server", n.PrimaryHub)
		}
		if n.ExitVia != 0 {
			exit, ok := byID[n.ExitVia]
			switch {
			case n.Role == proto.RoleServer:
				errf(n.Name, "servers do not use an exit")
			case n.Role == proto.RolePhone:
				errf(n.Name, "phones choose their exit with the via-exit QR profile, not an exit setting")
			case n.inContainer():
				errf(n.Name, "a container-mode node cannot use an exit: only its containers are on the mesh; clear the exit")
			case !ok || exit.Role != proto.RoleServer:
				errf(n.Name, "exit (node id %d) is not a server", n.ExitVia)
			case !exit.CanExit:
				errf(n.Name, "exit %q is not exit-capable; enable can-exit on it first", exit.Name)
			case n.active() && !p.planned(exit.ID):
				warnf(n.Name, "exit %q is disabled, not enrolled or has no endpoint; exit is off until it is back", exit.Name)
			}
		}
		if !n.active() {
			continue
		}
		switch n.Role {
		case proto.RoleServer:
			if !n.Endpoint.IsValid() && !n.Endpoint6.IsValid() {
				warnf(n.Name, "has no public endpoint, so no other node peers with it yet; set it with `meshctl node set %s --endpoint IP:PORT`", n.Name)
			}
			if n.Down {
				warnf(n.Name, "is down (no heartbeat and no handshake): the pairs it relays go through other hubs until it is back")
			}
		default:
			hub, ok := p.hub(n)
			switch {
			case !ok:
				warnf(n.Name, "there is no reachable server to peer with yet")
			case n.PrimaryHub != 0 && hub != n.PrimaryHub:
				warnf(n.Name, "primary hub %q is not reachable; relaying through %q until it is back", byID[n.PrimaryHub].Name, byID[hub].Name)
			}
			if ok && !p.reachesAServer(n) {
				warnf(n.Name, "can reach none of the servers: none shares an address family with it, or none has an endpoint its ip_family (%s) lets it dial",
					n.familyName())
			}
		}
	}
	return out
}

// reachesAServer reports whether n has a tunnel to some planned server:
// one with an endpoint n dials and a family both have.
func (p *plan) reachesAServer(n Node) bool {
	for _, s := range p.servers {
		if primary, _ := endpoints(n, s); primary != "" && n.fams()&s.fams() != 0 {
			return true
		}
	}
	return false
}

// checkUsable says what the families n's host cannot use do to its plan
// (M8, Node.usable).
func checkUsable(n Node, warnf func(node, format string, args ...any)) {
	lost := "IPv6"
	if n.NoIPv4 {
		lost = "IPv4"
	}
	switch u, ok := n.usable(); {
	case !ok:
		warnf(n.Name, "its agent reports that its host cannot use %s, which it cannot do without (it has no other family, or is a server or has a container network, which need IPv4); it is planned as configured", lost)
	case u != n:
		warnf(n.Name, "its agent reports that its host cannot use %s; the mesh leaves %s out for it until the host can", lost, lost)
	}
}

func (n Node) familyName() string {
	if n.Family == "" {
		return proto.FamilyDual
	}
	return n.Family
}

// checkNode6 checks n's IPv6 side (M8): its family setting, its IPv6 mesh
// address in the mesh's prefix and unique, its IPv6 endpoint.
func checkNode6(in Input, n Node, ips map[string]string, errf func(node, format string, args ...any)) {
	switch n.Family {
	case "", proto.FamilyDual, proto.FamilyIPv4, proto.FamilyIPv6:
	default:
		errf(n.Name, "unknown ip_family %q (want dual, ipv4 or ipv6)", n.Family)
	}
	switch {
	case n.Family == proto.FamilyIPv4 && n.MeshIP6.IsValid():
		errf(n.Name, "is IPv4-only (ip_family ipv4) but has the IPv6 mesh address %s; clear it or make it dual", n.MeshIP6)
	case n.Family == proto.FamilyIPv6 && n.MeshIP.IsValid():
		errf(n.Name, "is IPv6-only (ip_family ipv6) but has the IPv4 mesh IP %s; clear it or make it dual", n.MeshIP)
	}
	// The mesh firewall filters IPv4; a server without an IPv4 address
	// would enforce nothing (IPv6 policy is M8's 8d).
	if n.Role == proto.RoleServer && !n.MeshIP.IsValid() {
		errf(n.Name, "a server needs an IPv4 mesh address: the mesh firewall it enforces is IPv4")
	}
	if n.Endpoint6.IsValid() {
		switch {
		case n.Role != proto.RoleServer:
			errf(n.Name, "only servers have an endpoint; clear the IPv6 endpoint %s", n.Endpoint6)
		case !n.Endpoint6.Addr().Is6() || n.Endpoint6.Addr().Is4In6() || n.Endpoint6.Port() == 0:
			errf(n.Name, "IPv6 endpoint %s must be [IPv6]:port", n.Endpoint6)
		}
	}
	if !n.MeshIP6.IsValid() {
		return
	}
	switch {
	case !n.MeshIP6.Is6() || n.MeshIP6.Is4In6():
		errf(n.Name, "IPv6 mesh address %s is not an IPv6 address", n.MeshIP6)
		return
	case !in.MeshCIDR6.IsValid():
		errf(n.Name, "has the IPv6 mesh address %s, but the mesh has no IPv6 prefix; set one first", n.MeshIP6)
		return
	case !in.MeshCIDR6.Contains(n.MeshIP6):
		errf(n.Name, "IPv6 mesh address %s is outside the mesh's %s", n.MeshIP6, in.MeshCIDR6)
	case containers6(in.MeshCIDR6).Contains(n.MeshIP6):
		errf(n.Name, "IPv6 mesh address %s is in %s, where the container networks are", n.MeshIP6, containers6(in.MeshCIDR6))
	}
	if other, dup := ips[n.MeshIP6.String()]; dup {
		errf(n.Name, "IPv6 mesh address %s is also assigned to %q", n.MeshIP6, other)
	}
	ips[n.MeshIP6.String()] = n.Name
}

// checkDNS checks the zone, the extra domains, the custom records and the
// upstream resolvers. Node names are records already, so a custom record
// may not take one; record names are given without the zone.
func checkDNS(in Input, nodeNames map[string]bool, byID map[int64]Node, errf func(node, format string, args ...any)) {
	zone := in.ZoneName()
	if _, err := validate.Domains("extra DNS domains", zone, in.Domains); err != nil {
		errf("", "%v", err)
	}
	if _, err := validate.Zone("dns zone", zone); err != nil {
		errf("", "%v", err)
	}
	seen := map[string]bool{}
	for _, r := range in.Records {
		if _, err := validate.DNSName("dns record", r.Name); err != nil {
			errf("", "%v", err)
			continue
		}
		if r.Name == zone || strings.HasSuffix(r.Name, "."+zone) {
			errf("", "DNS record %q ends with the zone %q; give the name without it", r.Name, zone)
		}
		if nodeNames[r.Name] {
			errf("", "DNS record %q has the name of a node, which resolves to that node already", r.Name)
		}
		if seen[r.Name] {
			errf("", "DNS record %q is defined twice", r.Name)
		}
		seen[r.Name] = true
		_, known := byID[r.Node]
		switch {
		case r.Node != 0 && (r.IP.IsValid() || r.IP6.IsValid()):
			errf("", "DNS record %q points at both a node and an IP; keep one", r.Name)
		case r.Node != 0 && !known:
			errf("", "DNS record %q points at node id %d, which does not exist", r.Name, r.Node)
		case r.Node != 0:
		case !r.IP.IsValid() && !r.IP6.IsValid():
			errf("", "DNS record %q has no address", r.Name)
		}
		if r.Node != 0 {
			continue
		}
		switch {
		case r.IP.IsValid() && !r.IP.Is4():
			errf("", "DNS record %q has %s as its IPv4 address", r.Name, r.IP)
		case r.IP.IsValid() && in.MeshCIDR.IsValid() && !in.MeshCIDR.Contains(r.IP):
			errf("", "DNS record %q points at %s, outside the mesh %s; records name mesh addresses only", r.Name, r.IP, in.MeshCIDR)
		}
		switch {
		case r.IP6.IsValid() && (!r.IP6.Is6() || r.IP6.Is4In6()):
			errf("", "DNS record %q has %s as its IPv6 address", r.Name, r.IP6)
		case r.IP6.IsValid() && !in.MeshCIDR6.IsValid():
			errf("", "DNS record %q has the IPv6 address %s, but the mesh has no IPv6 prefix; remove the record or set a prefix first", r.Name, r.IP6)
		case r.IP6.IsValid() && !in.MeshCIDR6.Contains(r.IP6):
			errf("", "DNS record %q points at %s, outside the mesh's IPv6 prefix %s; records name mesh addresses only", r.Name, r.IP6, in.MeshCIDR6)
		}
	}
	for _, u := range in.Upstream {
		switch {
		case !u.Is4() && (!u.Is6() || u.Is4In6()):
			errf("", "upstream DNS %s is not an IPv4 or IPv6 address", u)
		case in.MeshCIDR.IsValid() && in.MeshCIDR.Contains(u), in.MeshCIDR6.IsValid() && in.MeshCIDR6.Contains(u):
			errf("", "upstream DNS %s is inside the mesh; a mesh resolver forwarding to another would loop", u)
		}
	}
}

// checkContainers checks the container subnets: each an IPv4 /24 of the
// mesh, no two the same, none holding a node's mesh IP (amendment A37).
func checkContainers(in Input, errf func(node, format string, args ...any)) {
	nodes := sortedByID(in.Nodes)
	owner := map[netip.Prefix]string{}
	var subnets []netip.Prefix
	for _, n := range nodes {
		k := n.Containers
		if !k.IsValid() {
			continue
		}
		switch {
		case !k.Addr().Is4() || k.Bits() != 24 || k.Masked() != k:
			errf(n.Name, "container subnet %s must be an IPv4 /24", k)
			continue
		case in.MeshCIDR.IsValid() && (in.MeshCIDR.Bits() > 24 || !in.MeshCIDR.Contains(k.Addr())):
			errf(n.Name, "container subnet %s is outside the mesh %s", k, in.MeshCIDR)
		}
		if other, dup := owner[k]; dup {
			errf(n.Name, "container subnet %s is also %q's", k, other)
			continue
		}
		owner[k] = n.Name
		subnets = append(subnets, k)
	}
	for _, n := range nodes {
		for _, k := range subnets {
			if n.MeshIP.IsValid() && k.Contains(n.MeshIP) {
				errf(n.Name, "mesh IP %s is inside %q's container subnet %s", n.MeshIP, owner[k], k)
			}
		}
	}
}

// checkContainers6 checks the containers' IPv6 subnets (M8): each a /112
// of the mesh's container blocks, no two the same, only beside an IPv4
// container network and on a node with an IPv6 address.
func checkContainers6(in Input, errf func(node, format string, args ...any)) {
	owner := map[netip.Prefix]string{}
	for _, n := range sortedByID(in.Nodes) {
		k := n.Containers6
		if !k.IsValid() {
			continue
		}
		switch {
		case !k.Addr().Is6() || k.Bits() != 112 || k.Masked() != k:
			errf(n.Name, "container IPv6 subnet %s must be an IPv6 /112", k)
			continue
		case !in.MeshCIDR6.IsValid() || !containers6(in.MeshCIDR6).Contains(k.Addr()):
			errf(n.Name, "container IPv6 subnet %s is outside the container part of the mesh's IPv6 prefix", k)
		case !n.Containers.IsValid():
			errf(n.Name, "has a container IPv6 subnet but no IPv4 container network: the containers' IPv6 comes beside their IPv4")
		case !n.has6():
			errf(n.Name, "has a container IPv6 subnet but no IPv6 mesh address of its own")
		}
		if other, dup := owner[k]; dup {
			errf(n.Name, "container IPv6 subnet %s is also %q's", k, other)
		}
		owner[k] = n.Name
	}
}

func sortedByID(ns []Node) []Node {
	out := append([]Node(nil), ns...)
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}
