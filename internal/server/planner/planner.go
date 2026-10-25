// Package planner computes every node's desired config from the mesh's
// desired state (spec S4). It is a pure function: no I/O, no clock, no
// storage. The server's reconciler feeds it the database contents and stores
// what it returns; golden tests pin its output.
//
// Topology rules:
//
//   - Servers form a full mesh: each server's AllowedIPs for another server is
//     that server's /32, no keepalive.
//   - Every spoke and phone peers with every server. On a server, a spoke's or
//     phone's AllowedIPs is its own /32 only.
//   - Traffic between two non-server nodes A and B relays through one server,
//     relay(A, B), and BOTH sides route the other's /32 to that server:
//     WireGuard drops a packet whose source is not in the AllowedIPs of the
//     peer it arrived from, so the relay must be the same in both directions
//     (amendment A3). relay(A, B) is the primary hub of the phone if either is
//     a phone (a phone's profile is static: its hub peer carries the whole
//     mesh CIDR), else the primary hub of the node with the lower id.
//   - A spoke with an exit adds 0.0.0.0/0 to that server's AllowedIPs.
//   - A server that relays gets ip_forward and the firewalld relay policy; an
//     exit-capable server gets ip_forward and the exit (masquerade) policy.
//   - A server without a known endpoint is not planned into anyone's peers
//     until it has one (amendment A17).
//   - Every server and spoke serves the whole mesh zone: one record per
//     active node plus the custom records. It forwards other names only in
//     exit mode or, on a server, for the phones that use it as their
//     resolver (amendment A12).
//   - A node whose containers are on the mesh has a /24 of it; every peer
//     routes that subnet wherever it routes the node's /32, and the node's
//     resolver forwards for it. In container mode the node's WireGuard is
//     in the agent container, which is its containers' gateway (policy
//     routing in its own namespace, no firewall section, never an exit); in
//     host mode the host routes the subnet on its bridge, which a server's
//     firewall forwards to (amendment A37).
//   - Every server gets the same compiled segmentation policy, the mesh
//     firewall, and its mesh zone lets the whole mesh through: the policy,
//     not the zone, decides what crosses. Spokes enforce nothing, since
//     every packet between two nodes crosses a server (amendment A43).
//   - Dual stack (M8): a node has an IPv4 mesh address, an IPv6 one, or
//     both, each allocated on its own. Two nodes route each other the
//     families both have, relays are chosen per family (a relay must have
//     it), and a node dials a server's endpoint in a family its ip_family
//     allows -- IPv4 first, the IPv6 endpoint as the fallback. An exit
//     carries ::/0 too when both ends have IPv6 and the exit has an IPv6
//     endpoint (public IPv6). A mesh without an IPv6 prefix plans exactly
//     as before.
package planner

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net/netip"
	"slices"
	"sort"
	"strings"

	"github.com/hafio/vpn-mesh-utility/internal/proto"
)

// Input is the desired state of the mesh.
type Input struct {
	// Slot is the mesh's slot (amendment A49), which every config carries.
	Slot     proto.Slot
	MeshCIDR netip.Prefix
	// MeshCIDR6 is the mesh's IPv6 prefix (M8); invalid without one.
	MeshCIDR6 netip.Prefix
	// DefaultHub is the hub for spokes without a usable primary hub; 0 means
	// the lowest-id server.
	DefaultHub int64
	Nodes      []Node
	// Upstream are the resolvers non-mesh queries are forwarded to (spec
	// S7.4), in the operator's order.
	Upstream []netip.Addr
	// Zone is the DNS zone every node serves; empty means
	// proto.DefaultZone (amendment A36).
	Zone string
	// Domains are the extra DNS domains every node resolves (amendment
	// A39).
	Domains []string
	// Records are the custom DNS records.
	Records []Record
	// Segmentation is the mesh firewall (amendment A43); empty, the
	// servers let through only the system rules.
	Segmentation Segmentation
}

// Record is a custom name in the mesh zone: an alias of a node (Node, which
// follows its mesh addresses) or a fixed IP, IP6 or both.
type Record struct {
	Name string
	Node int64
	IP   netip.Addr
	IP6  netip.Addr
}

// Node is the planner's view of one node.
type Node struct {
	ID       int64
	Name     string
	Role     proto.Role
	MeshIP   netip.Addr
	WGPubKey string
	// Endpoint is a server's public IPv4:port; invalid (zero) when unknown.
	Endpoint   netip.AddrPort
	ListenPort int
	PrimaryHub int64
	CanExit    bool
	ExitVia    int64
	Keepalive  int
	MTU        int
	Enrolled   bool
	Disabled   bool
	// Mode is proto.ModeHost (or empty) or proto.ModeContainer; Containers
	// is the node's /24 for containers, invalid while they are not on the
	// mesh (amendment A37).
	Mode       string
	Containers netip.Prefix
	// Down marks a server the relay failover monitor judged down
	// (amendment A40): the pairs it relays go through another hub. It stays
	// a peer of everyone, so direct traffic resumes when it is back.
	Down bool
	// The IPv6 side (M8): MeshIP6 is invalid when the node has none;
	// Family is proto.FamilyDual (or empty), FamilyIPv4 or FamilyIPv6;
	// Endpoint6 a server's public [IPv6]:port; Containers6 its containers'
	// /112, invalid unless their IPv6 is on the mesh.
	MeshIP6     netip.Addr
	Family      string
	Endpoint6   netip.AddrPort
	Containers6 netip.Prefix
	// NoIPv4 and NoIPv6 say the node's agent reported that its host cannot
	// use that family (M8): the plan leaves it out for the node (usable).
	NoIPv4, NoIPv6 bool
}

// fam is a set of address families.
type fam uint8

const (
	f4 fam = 1 << iota
	f6
)

// active nodes are enrolled, enabled and have a key: the only ones that get
// a config or appear in anyone's peers.
func (n Node) active() bool { return n.Enrolled && !n.Disabled && n.WGPubKey != "" }

// usable is n as the mesh plans it (M8): without the family its agent
// reported its host cannot use -- that family's address, endpoint and
// containers -- so no peer routes it there, DNS does not name it there and
// nobody dials it there. ok is false when n cannot do without the family
// (none would be left, or it has a container network, which needs IPv4);
// it is then planned as configured.
func (n Node) usable() (u Node, ok bool) {
	switch {
	case !n.NoIPv4 && !n.NoIPv6:
		return n, true
	case n.NoIPv4 && (n.NoIPv6 || !n.has6() || n.Containers.IsValid()):
		return n, false
	case n.NoIPv6 && !n.has4():
		return n, false
	case n.NoIPv6:
		n.MeshIP6, n.Endpoint6, n.Containers6, n.Family = netip.Addr{}, netip.AddrPort{}, netip.Prefix{}, proto.FamilyIPv4
	default:
		n.MeshIP, n.Endpoint, n.Family = netip.Addr{}, netip.AddrPort{}, proto.FamilyIPv6
	}
	return n, true
}

// usable is in with every node as the mesh plans it (Node.usable).
func (in Input) usable() Input {
	in.Nodes = slices.Clone(in.Nodes) // the caller's slice stays as it was
	for i, n := range in.Nodes {
		if u, ok := n.usable(); ok {
			in.Nodes[i] = u
		}
	}
	return in
}

func (n Node) has4() bool { return n.MeshIP.Is4() }
func (n Node) has6() bool { return n.MeshIP6.IsValid() }

// fams is the families of n's mesh addresses.
func (n Node) fams() fam {
	var f fam
	if n.has4() {
		f |= f4
	}
	if n.has6() {
		f |= f6
	}
	return f
}

// dials is the families n may reach a server's endpoint in (its
// ip_family).
func (n Node) dials() fam {
	switch n.Family {
	case proto.FamilyIPv4:
		return f4
	case proto.FamilyIPv6:
		return f6
	}
	return f4 | f6
}

func (n Node) prefix() string  { return netip.PrefixFrom(n.MeshIP, 32).String() }
func (n Node) prefix6() string { return netip.PrefixFrom(n.MeshIP6, 128).String() }

// address and address6 are n's mesh addresses as an interface carries
// them, "" for a family it does not have.
func (n Node) address() string {
	if !n.has4() {
		return ""
	}
	return n.prefix()
}

func (n Node) address6() string {
	if !n.has6() {
		return ""
	}
	return n.prefix6()
}

// allowed is what a peer routes to n in the families f: its mesh
// addresses, and its containers' subnets when they are on the mesh.
func (n Node) allowed(f fam) []string {
	var out []string
	if f&f4 != 0 && n.has4() {
		out = append(out, n.prefix())
		if n.Containers.IsValid() {
			out = append(out, n.Containers.String())
		}
	}
	if f&f6 != 0 && n.has6() {
		out = append(out, n.prefix6())
		if n.Containers6.IsValid() {
			out = append(out, n.Containers6.String())
		}
	}
	return out
}

// endpoints is where n reaches server s: the endpoint in a family n dials,
// IPv4 first, and the other family's as the fallback. primary is empty
// when s has no endpoint n may use.
func endpoints(n, s Node) (primary, fallback string) {
	var four, six string
	if n.dials()&f4 != 0 && s.Endpoint.IsValid() {
		four = s.Endpoint.String()
	}
	if n.dials()&f6 != 0 && s.Endpoint6.IsValid() {
		six = s.Endpoint6.String()
	}
	if four != "" {
		return four, six
	}
	return six, ""
}

func (n Node) mode() string {
	if n.Mode == "" {
		return proto.ModeHost
	}
	return n.Mode
}

func (n Node) inContainer() bool { return n.Mode == proto.ModeContainer }

// Compute returns the config of every active node, keyed by node id, plus the
// warnings Check found. It fails (and returns no configs) when Check finds an
// error: a half-valid plan is never applied.
func Compute(in Input) (map[int64]proto.NodeConfig, []Problem, error) {
	problems := Check(in)
	var errs []string
	for _, p := range problems {
		if !p.Warning {
			errs = append(errs, p.String())
		}
	}
	if len(errs) > 0 {
		return nil, problems, errors.New("the mesh plan has errors:\n  " + strings.Join(errs, "\n  "))
	}

	in = in.usable()
	p := newPlan(in)
	p.policy = CompilePolicy(in)
	out := map[int64]proto.NodeConfig{}
	for _, n := range p.nodes {
		if !n.active() {
			continue
		}
		var c proto.NodeConfig
		switch n.Role {
		case proto.RoleServer:
			c = p.server(n)
		case proto.RoleSpoke:
			c = p.spoke(n)
		case proto.RolePhone:
			c = p.phone(n)
		default:
			return nil, problems, fmt.Errorf("node %q has unknown role %q", n.Name, n.Role)
		}
		out[n.ID] = c
	}
	return out, problems, nil
}

// plan holds the derived topology of one Compute call.
type plan struct {
	in      Input
	nodes   []Node         // all nodes, by id
	byID    map[int64]Node // all nodes
	servers []Node         // active servers with an endpoint, by id
	others  []Node         // active spokes and phones, by id
	exits   []string       // names of planned exit-capable servers, sorted
	records []proto.Record // the mesh zone, by name
	// policy is what every server enforces.
	policy *proto.Policy
}

func newPlan(in Input) *plan {
	p := &plan{in: in, byID: map[int64]Node{}}
	p.nodes = append([]Node(nil), in.Nodes...)
	sort.Slice(p.nodes, func(i, j int) bool { return p.nodes[i].ID < p.nodes[j].ID })
	for _, n := range p.nodes {
		p.byID[n.ID] = n
		if !n.active() {
			continue
		}
		switch n.Role {
		case proto.RoleServer:
			if n.Endpoint.IsValid() || n.Endpoint6.IsValid() {
				p.servers = append(p.servers, n)
				if n.CanExit {
					p.exits = append(p.exits, n.Name)
				}
			}
		default:
			p.others = append(p.others, n)
		}
	}
	sort.Strings(p.exits)
	// A name resolves once its node is on the mesh; an alias of a node that
	// is not is left out until it is.
	for _, n := range p.nodes {
		if n.active() {
			p.records = append(p.records, record(n.Name, n.MeshIP, n.MeshIP6))
		}
	}
	for _, r := range in.Records {
		target, ok := p.byID[r.Node]
		switch {
		case r.Node == 0:
			p.records = append(p.records, record(r.Name, r.IP, r.IP6))
		case ok && target.active():
			p.records = append(p.records, record(r.Name, target.MeshIP, target.MeshIP6))
		}
	}
	sort.Slice(p.records, func(i, j int) bool { return p.records[i].Name < p.records[j].Name })
	return p
}

// record is a name with the addresses that are valid.
func record(name string, ip, ip6 netip.Addr) proto.Record {
	r := proto.Record{Name: name}
	if ip.IsValid() {
		r.IP = ip.String()
	}
	if ip6.IsValid() {
		r.IP6 = ip6.String()
	}
	return r
}

// planned reports whether id is a server that appears in peer lists.
func (p *plan) planned(id int64) bool {
	for _, s := range p.servers {
		if s.ID == id {
			return true
		}
	}
	return false
}

// hub is the server a spoke or phone relays through: its primary hub when
// that server is planned, else the default hub, else the lowest-id server.
// ok is false when there is no planned server at all.
func (p *plan) hub(n Node) (int64, bool) {
	if len(p.servers) == 0 {
		return 0, false
	}
	if n.PrimaryHub != 0 && p.planned(n.PrimaryHub) {
		return n.PrimaryHub, true
	}
	if p.in.DefaultHub != 0 && p.planned(p.in.DefaultHub) {
		return p.in.DefaultHub, true
	}
	return p.servers[0].ID, true
}

// relay is the server that carries traffic of family f between two
// non-server nodes. It is symmetric: relay(a, b, f) == relay(b, a, f). A
// pair with a phone relays through the phone's hub, which its profile
// names. Any other pair relays through the lower-id node's hub; when the
// failover monitor judged that one down (amendment A40), or it lacks f or
// one of the two cannot dial it (M8), through the other node's hub, else
// the default hub, else the lowest-id server that qualifies -- each chosen
// from the pair and the mesh alone, so both ends agree.
func (p *plan) relay(a, b Node, f fam) (int64, bool) {
	fits := func(id int64) bool {
		s := p.byID[id]
		pa, _ := endpoints(a, s)
		pb, _ := endpoints(b, s)
		return s.fams()&f != 0 && pa != "" && pb != ""
	}
	switch {
	case a.Role == proto.RolePhone:
		h, ok := p.hub(a)
		return h, ok && fits(h)
	case b.Role == proto.RolePhone:
		h, ok := p.hub(b)
		return h, ok && fits(h)
	}
	if b.ID < a.ID {
		a, b = b, a
	}
	up := func(id int64) bool { return fits(id) && !p.byID[id].Down }
	first, ok := p.hub(a)
	if !ok {
		return 0, false
	}
	if up(first) {
		return first, true
	}
	if other, _ := p.hub(b); up(other) {
		return other, true
	}
	if d := p.in.DefaultHub; d != 0 && p.planned(d) && up(d) {
		return d, true
	}
	for _, s := range p.servers {
		if up(s.ID) {
			return s.ID, true
		}
	}
	// Every server that would do is down: none is better.
	if fits(first) {
		return first, true
	}
	for _, s := range p.servers {
		if fits(s.ID) {
			return s.ID, true
		}
	}
	return 0, false
}

// relays reports whether server id carries traffic of a family in f for
// any pair.
func (p *plan) relays(id int64, f fam) bool {
	for i, a := range p.others {
		for _, b := range p.others[i+1:] {
			for _, x := range []fam{f4, f6} {
				if f&x == 0 || a.fams()&b.fams()&x == 0 {
					continue
				}
				if r, ok := p.relay(a, b, x); ok && r == id {
					return true
				}
			}
		}
	}
	return false
}

func (p *plan) server(n Node) proto.NodeConfig {
	port := n.ListenPort
	c := proto.NodeConfig{
		Node:      proto.Self{ID: n.ID, Name: n.Name, Role: n.Role, Slot: p.in.Slot},
		Interface: proto.Interface{Address: n.address(), Address6: n.address6(), ListenPort: &port, MTU: n.MTU},
		Peers:     []proto.Peer{},
	}
	for _, s := range p.servers {
		common := n.fams() & s.fams()
		if s.ID == n.ID || common == 0 {
			continue
		}
		// Either end may open the tunnel: one that cannot dial the other
		// lists it without an endpoint and waits for its handshake.
		primary, fallback := endpoints(n, s)
		if back, _ := endpoints(s, n); primary == "" && back == "" {
			continue
		}
		c.Peers = append(c.Peers, proto.Peer{
			Name: s.Name, PublicKey: s.WGPubKey, Endpoint: primary, Fallback: fallback,
			AllowedIPs: sortPrefixes(s.allowed(common)),
		})
	}
	// Spokes and phones are planned only into servers others can reach: an
	// endpointless server has no spokes pointing at it, so it lists none --
	// nor one that cannot dial it, or shares no family with it (M8).
	if p.planned(n.ID) {
		for _, o := range p.others {
			common := n.fams() & o.fams()
			if dial, _ := endpoints(o, n); common == 0 || dial == "" {
				continue
			}
			c.Peers = append(c.Peers, proto.Peer{Name: o.Name, PublicKey: o.WGPubKey, AllowedIPs: sortPrefixes(o.allowed(common))})
		}
	}
	sortPeers(c.Peers)

	relay4, relay6 := p.relays(n.ID, f4), p.relays(n.ID, f6)
	relay := relay4 || relay6
	exit6 := n.CanExit && n.Endpoint6.IsValid() && p.in.MeshCIDR6.IsValid() && n.has6()
	c.Sysctl = p.sysctl(n, relay4 || n.CanExit, relay6 || exit6, false)
	c.DNS = p.dns(n, false)
	c.Containers = p.containers(n)
	c.Policy = p.policy
	if n.inContainer() {
		c.Containers.Relay = relay
		return c
	}
	fw := &proto.Firewall{
		Backend:     proto.FirewallFirewalld,
		Rules:       []proto.FWRule{{Zone: proto.ZonePublic, Proto: "udp", Port: n.ListenPort, Note: "wireguard"}},
		Relay:       relay,
		ExitGateway: n.CanExit,
	}
	// firewalld's policy masquerade is IPv4 only (trial 8a, F4): an exit
	// with public IPv6 masquerades the mesh's IPv6 prefix by a rich rule.
	if exit6 {
		fw.Exit6 = p.in.MeshCIDR6.String()
	}
	// The containers' bridge forwards to and from wg0 like a relay.
	if n.Containers.IsValid() {
		fw.Bridge, fw.Relay = p.in.Slot.Bridge(), true
	}
	// The policy decides what crosses the mesh zone: the zone lets the
	// whole mesh through to it, and the policy's table drops what it does
	// not allow (amendment A43). Swarm, DNS and ping are its system rules.
	fw.Rules = append(fw.Rules, proto.FWRule{Zone: proto.ZoneMesh, Proto: proto.ProtoAll, Source: p.in.MeshCIDR.String(),
		Note: "segmentation: the policy decides"})
	if p.in.MeshCIDR6.IsValid() && n.has6() {
		// The policy filters IPv6 too (M8): the zone lets the mesh's IPv6 in.
		fw.Rules = append(fw.Rules, proto.FWRule{Zone: proto.ZoneMesh, Proto: proto.ProtoAll, Source: p.in.MeshCIDR6.String(),
			Note: "segmentation: the policy decides"})
	}
	sortRules(fw.Rules)
	c.Firewall = fw
	return c
}

func (p *plan) containers(n Node) *proto.Containers {
	return ContainerNetwork(p.in.MeshCIDR, p.in.MeshCIDR6, n)
}

// ServerIPs are the mesh IPs of the servers in peer lists, by id: the
// resolvers a Swarm service on the mesh uses, whose tasks run on several
// nodes.
func ServerIPs(in Input) []string {
	var out []string
	for _, s := range newPlan(in).servers {
		out = append(out, s.MeshIP.String())
	}
	return out
}

// ContainerNetwork is n's container network in mesh (amendment A37), nil
// when its containers are not on the mesh: .1 of its subnet is the
// runtime's bridge, .2 the agent in container mode, the upper half the
// runtime's range, and the network on the link the whole mesh. Its IPv6
// side, when the containers' IPv6 is on the mesh (M8), is the same shape
// in its /112 of mesh6: ::1 the bridge, ::2 the agent, the upper half the
// range, mesh6 on the link.
func ContainerNetwork(mesh, mesh6 netip.Prefix, n Node) *proto.Containers {
	if !n.Containers.IsValid() {
		return nil
	}
	a := n.Containers.Addr().As4()
	at := func(host byte) netip.Addr {
		a[3] = host
		return netip.AddrFrom4(a)
	}
	k := &proto.Containers{Mode: n.mode(), Subnet: n.Containers.String(), Network: mesh.String(),
		Range: netip.PrefixFrom(at(128), 25).String(), Bridge: at(1).String()}
	k.Gateway = k.Bridge
	if n.inContainer() {
		k.Gateway = at(2).String()
	}
	if n.Containers6.IsValid() && mesh6.IsValid() {
		b := n.Containers6.Addr().As16()
		at6 := func(host uint16) netip.Addr {
			binary.BigEndian.PutUint16(b[14:16], host)
			return netip.AddrFrom16(b)
		}
		k.Subnet6, k.Network6 = n.Containers6.String(), mesh6.String()
		k.Range6, k.Bridge6 = netip.PrefixFrom(at6(0x8000), 113).String(), at6(1).String()
		k.Gateway6 = k.Bridge6
		if n.inContainer() {
			k.Gateway6 = at6(2).String()
		}
	}
	return k
}

// sysctl is the kernel settings n needs, by key: forwarding of a family
// when it relays it, exits it or routes containers in it; proxy ARP and
// proxy NDP for a container-mode gateway (the compose file sets them
// there, so the agent only checks them); src_valid_mark in exit mode. A
// host-mode bridge's own settings are the agent's containers step, which
// runs once the bridge exists.
func (p *plan) sysctl(n Node, forward, forward6, exitMode bool) []proto.Sysctl {
	on := map[string]bool{}
	if (forward || n.Containers.IsValid()) && n.has4() {
		on[proto.SysctlIPForward] = true
	}
	if (forward6 || n.Containers6.IsValid()) && n.has6() {
		on[proto.SysctlIPv6Forward] = true
	}
	if n.Containers.IsValid() && n.inContainer() {
		on[proto.SysctlProxyARP] = true
	}
	if n.Containers6.IsValid() && n.inContainer() {
		on[proto.SysctlProxyNDP] = true
	}
	if exitMode {
		on[proto.SysctlSrcValidMark] = true
	}
	var out []proto.Sysctl
	for _, k := range proto.SysctlKeys(p.in.Slot) {
		if on[k] {
			out = append(out, proto.Sysctl{Key: k, Value: "1"})
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Key < out[j].Key })
	return out
}

// ZoneName is the zone the mesh serves: Zone, or the default.
func (in Input) ZoneName() string {
	if in.Zone == "" {
		return proto.DefaultZone
	}
	return in.Zone
}

// dns is the resolver config of a server or spoke: the whole zone, and
// forwarding of other names in exit mode (all of the node's lookups go
// through the tunnel), for the containers of every container network (a
// Swarm service's tasks on several nodes share one resolver list, so any
// node's resolver serves them), and, on a server, for the phones whose hub
// it is (their profile names it as their resolver).
func (p *plan) dns(n Node, exitOn bool) *proto.DNS {
	d := &proto.DNS{Zone: p.in.ZoneName(), Records: slices.Clone(p.records), Split: !exitOn}
	if len(p.in.Domains) > 0 {
		d.Domains = slices.Sorted(slices.Values(p.in.Domains))
	}
	for _, u := range p.in.Upstream {
		d.Upstream = append(d.Upstream, u.String())
	}
	var forward []string
	for _, o := range p.nodes {
		if o.active() && o.Containers.IsValid() {
			forward = append(forward, o.Containers.String())
		}
		if o.active() && o.Containers6.IsValid() {
			forward = append(forward, o.Containers6.String())
		}
	}
	if n.Role == proto.RoleServer {
		for _, o := range p.others {
			if o.Role != proto.RolePhone {
				continue
			}
			if hub, ok := p.hub(o); ok && hub == n.ID {
				forward = append(forward, o.allowed(f4|f6)...)
			}
		}
	}
	if len(forward) > 0 {
		d.ForwardFor = sortPrefixes(forward)
	}
	return d
}

func (p *plan) spoke(n Node) proto.NodeConfig {
	c := proto.NodeConfig{
		Node:      proto.Self{ID: n.ID, Name: n.Name, Role: n.Role, Slot: p.in.Slot},
		Interface: proto.Interface{Address: n.address(), Address6: n.address6(), MTU: n.MTU, FwMark: proto.FwMark},
		Peers:     []proto.Peer{},
	}
	exitOn := false
	for _, s := range p.servers {
		primary, fallback := endpoints(n, s)
		common := n.fams() & s.fams()
		if primary == "" || common == 0 {
			continue // no tunnel to it, or nothing to route through one
		}
		allowed := s.allowed(common)
		for _, o := range p.others {
			if o.ID == n.ID {
				continue
			}
			for _, f := range []fam{f4, f6} {
				if common&o.fams()&f == 0 {
					continue
				}
				if r, ok := p.relay(n, o, f); ok && r == s.ID {
					allowed = append(allowed, o.allowed(f)...)
				}
			}
		}
		// Only containers are on the mesh in container mode: the host's own
		// traffic has no tunnel to take, so there is no exit (Check). IPv6
		// goes through the exit when both ends have it and the exit has
		// public IPv6 (an IPv6 endpoint); otherwise the agent holds it back.
		if n.ExitVia == s.ID && s.CanExit && !n.inContainer() {
			if common&f4 != 0 {
				allowed = append(allowed, "0.0.0.0/0")
				exitOn = true
			}
			if common&f6 != 0 && s.Endpoint6.IsValid() && p.in.MeshCIDR6.IsValid() {
				allowed = append(allowed, "::/0")
				exitOn = true
			}
		}
		c.Peers = append(c.Peers, proto.Peer{
			Name: s.Name, PublicKey: s.WGPubKey, Endpoint: primary, Fallback: fallback,
			AllowedIPs: sortPrefixes(allowed), Keepalive: n.Keepalive,
		})
	}
	sortPeers(c.Peers)
	if (len(p.exits) > 0 || exitOn) && !n.inContainer() {
		c.Exit = &proto.Exit{Choices: append([]string(nil), p.exits...)}
		if exitOn {
			c.Exit.Via = p.byID[n.ExitVia].Name
		}
	}
	// src_valid_mark is wg-quick's companion setting for fwmark policy
	// routing in exit mode.
	c.Sysctl = p.sysctl(n, false, false, exitOn)
	c.DNS = p.dns(n, exitOn)
	c.Containers = p.containers(n)
	return c
}

// phone builds the static profile a phone imports by QR. Its hub peer carries
// the whole mesh CIDR (a static profile cannot learn about nodes added
// later); longest-prefix match still sends traffic for another server
// straight to that server's /32 peer. The via-exit variant is derived from
// this at QR time by adding 0.0.0.0/0 to the chosen exit's peer.
func (p *plan) phone(n Node) proto.NodeConfig {
	c := proto.NodeConfig{
		Node:      proto.Self{ID: n.ID, Name: n.Name, Role: n.Role, Slot: p.in.Slot},
		Interface: proto.Interface{Address: n.address(), Address6: n.address6(), MTU: n.MTU},
		Peers:     []proto.Peer{},
	}
	hub, _ := p.hub(n)
	for _, s := range p.servers {
		// A profile has one endpoint per peer: no fallback.
		primary, _ := endpoints(n, s)
		common := n.fams() & s.fams()
		if primary == "" || common == 0 {
			continue
		}
		allowed := sortPrefixes(s.allowed(common))
		if s.ID == hub {
			var whole []string
			if common&f4 != 0 {
				whole = append(whole, p.in.MeshCIDR.String())
			}
			if common&f6 != 0 {
				whole = append(whole, p.in.MeshCIDR6.String())
			}
			allowed = sortPrefixes(whole)
		}
		c.Peers = append(c.Peers, proto.Peer{
			Name: s.Name, PublicKey: s.WGPubKey, Endpoint: primary,
			AllowedIPs: allowed, Keepalive: n.Keepalive,
		})
	}
	sortPeers(c.Peers)
	if len(p.exits) > 0 {
		c.Exit = &proto.Exit{Choices: append([]string(nil), p.exits...)}
	}
	return c
}

func sortPeers(ps []proto.Peer) {
	sort.Slice(ps, func(i, j int) bool { return ps[i].Name < ps[j].Name })
}

// sortPrefixes orders by address then length, with the default route last.
func sortPrefixes(in []string) []string {
	ps := make([]netip.Prefix, 0, len(in))
	for _, s := range in {
		ps = append(ps, netip.MustParsePrefix(s))
	}
	sort.Slice(ps, func(i, j int) bool {
		di, dj := ps[i].Bits() == 0, ps[j].Bits() == 0
		if di != dj {
			return dj
		}
		if c := ps[i].Addr().Compare(ps[j].Addr()); c != 0 {
			return c < 0
		}
		return ps[i].Bits() < ps[j].Bits()
	})
	out := make([]string, len(ps))
	for i, p := range ps {
		out[i] = p.String()
	}
	return out
}

func sortRules(rs []proto.FWRule) {
	sort.Slice(rs, func(i, j int) bool {
		a, b := rs[i], rs[j]
		if a.Zone != b.Zone {
			return a.Zone < b.Zone
		}
		if a.Port != b.Port {
			return a.Port < b.Port
		}
		if a.Proto != b.Proto {
			return a.Proto < b.Proto
		}
		if a.Source != b.Source {
			return a.Source < b.Source
		}
		return a.Note < b.Note
	})
}
