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
//     resolver (amendment A12). Servers open 53 to the mesh and the custom
//     firewall rules in their mesh zone.
package planner

import (
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
	MeshCIDR netip.Prefix
	// DefaultHub is the hub for spokes without a usable primary hub; 0 means
	// the lowest-id server.
	DefaultHub int64
	Nodes      []Node
	// Upstream are the resolvers non-mesh queries are forwarded to (spec
	// S7.4), in the operator's order.
	Upstream []netip.Addr
	// Records are the custom DNS records, Rules the custom firewall
	// openings.
	Records []Record
	Rules   []Rule
}

// Record is a custom name in the mesh zone: an alias of a node (Node, which
// follows its mesh IP) or a fixed IP.
type Record struct {
	Name string
	Node int64
	IP   netip.Addr
}

// Rule opens a port in the mesh zone of one server (Node) or of every server
// (Node 0), from Source only.
type Rule struct {
	ID     int64
	Node   int64
	Proto  string
	Port   int
	Source netip.Prefix
	Note   string
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
}

// active nodes are enrolled, enabled and have a key: the only ones that get
// a config or appear in anyone's peers.
func (n Node) active() bool { return n.Enrolled && !n.Disabled && n.WGPubKey != "" }

func (n Node) prefix() string { return netip.PrefixFrom(n.MeshIP, 32).String() }

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

	p := newPlan(in)
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
			if n.Endpoint.IsValid() {
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
			p.records = append(p.records, proto.Record{Name: n.Name, IP: n.MeshIP.String()})
		}
	}
	for _, r := range in.Records {
		target, ok := p.byID[r.Node]
		switch {
		case r.Node == 0:
			p.records = append(p.records, proto.Record{Name: r.Name, IP: r.IP.String()})
		case ok && target.active():
			p.records = append(p.records, proto.Record{Name: r.Name, IP: target.MeshIP.String()})
		}
	}
	sort.Slice(p.records, func(i, j int) bool { return p.records[i].Name < p.records[j].Name })
	return p
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

// relay is the server that carries traffic between two non-server nodes. It
// is symmetric: relay(a, b) == relay(b, a).
func (p *plan) relay(a, b Node) (int64, bool) {
	switch {
	case a.Role == proto.RolePhone:
		return p.hub(a)
	case b.Role == proto.RolePhone:
		return p.hub(b)
	case a.ID < b.ID:
		return p.hub(a)
	default:
		return p.hub(b)
	}
}

// relays reports whether server id carries traffic for any pair.
func (p *plan) relays(id int64) bool {
	for i, a := range p.others {
		for _, b := range p.others[i+1:] {
			if r, ok := p.relay(a, b); ok && r == id {
				return true
			}
		}
	}
	return false
}

func (p *plan) server(n Node) proto.NodeConfig {
	port := n.ListenPort
	c := proto.NodeConfig{
		Node:      proto.Self{ID: n.ID, Name: n.Name, Role: n.Role},
		Interface: proto.Interface{Address: n.prefix(), ListenPort: &port, MTU: n.MTU},
		Peers:     []proto.Peer{},
	}
	for _, s := range p.servers {
		if s.ID == n.ID {
			continue
		}
		c.Peers = append(c.Peers, proto.Peer{
			Name: s.Name, PublicKey: s.WGPubKey, Endpoint: s.Endpoint.String(),
			AllowedIPs: []string{s.prefix()},
		})
	}
	// Spokes and phones are planned only into servers others can reach: an
	// endpointless server has no spokes pointing at it, so it lists none.
	if p.planned(n.ID) {
		for _, o := range p.others {
			c.Peers = append(c.Peers, proto.Peer{Name: o.Name, PublicKey: o.WGPubKey, AllowedIPs: []string{o.prefix()}})
		}
	}
	sortPeers(c.Peers)

	relay := p.relays(n.ID)
	fw := &proto.Firewall{
		Backend:     "firewalld",
		Rules:       []proto.FWRule{{Zone: proto.ZonePublic, Proto: "udp", Port: n.ListenPort, Note: "wireguard"}},
		Relay:       relay,
		ExitGateway: n.CanExit,
	}
	// Swarm ports only from the other servers (spec S7.5): a compromised spoke
	// reaches the mesh but never the Swarm control or data plane.
	for _, s := range p.servers {
		if s.ID == n.ID {
			continue
		}
		src := s.prefix()
		fw.Rules = append(fw.Rules,
			proto.FWRule{Zone: proto.ZoneMesh, Proto: "tcp", Port: 2377, Source: src, Note: "swarm manager: " + s.Name},
			proto.FWRule{Zone: proto.ZoneMesh, Proto: "tcp", Port: 7946, Source: src, Note: "swarm gossip: " + s.Name},
			proto.FWRule{Zone: proto.ZoneMesh, Proto: "udp", Port: 7946, Source: src, Note: "swarm gossip: " + s.Name},
			proto.FWRule{Zone: proto.ZoneMesh, Proto: "udp", Port: 4789, Source: src, Note: "swarm overlay: " + s.Name},
		)
	}
	// The resolver answers the whole mesh (spec S7.5); the custom rules are
	// the operator's.
	mesh := p.in.MeshCIDR.String()
	fw.Rules = append(fw.Rules,
		proto.FWRule{Zone: proto.ZoneMesh, Proto: "udp", Port: 53, Source: mesh, Note: "dns"},
		proto.FWRule{Zone: proto.ZoneMesh, Proto: "tcp", Port: 53, Source: mesh, Note: "dns"},
	)
	for _, r := range p.in.Rules {
		if r.Node == 0 || r.Node == n.ID {
			fw.Rules = append(fw.Rules, proto.FWRule{Zone: proto.ZoneMesh, Proto: r.Proto, Port: r.Port, Source: r.Source.String(),
				Note: ruleNote(r)})
		}
	}
	sortRules(fw.Rules)
	c.Firewall = fw
	if relay || n.CanExit {
		c.Sysctl = []proto.Sysctl{{Key: proto.SysctlIPForward, Value: "1"}}
	}
	c.DNS = p.dns(n, false)
	return c
}

// ruleNote says where a custom rule comes from, for the firewall page.
func ruleNote(r Rule) string {
	if r.Note == "" {
		return fmt.Sprintf("custom rule %d", r.ID)
	}
	return "custom: " + r.Note
}

// dns is the resolver config of a server or spoke: the whole zone, and
// forwarding of other names in exit mode (all of the node's lookups go
// through the tunnel) or, on a server, for the phones whose hub it is (their
// profile names it as their resolver).
func (p *plan) dns(n Node, exitOn bool) *proto.DNS {
	d := &proto.DNS{Zone: proto.Zone, Records: slices.Clone(p.records), Split: !exitOn}
	for _, u := range p.in.Upstream {
		d.Upstream = append(d.Upstream, u.String())
	}
	if n.Role == proto.RoleServer {
		var phones []string
		for _, o := range p.others {
			if o.Role != proto.RolePhone {
				continue
			}
			if hub, ok := p.hub(o); ok && hub == n.ID {
				phones = append(phones, o.prefix())
			}
		}
		d.ForwardFor = sortPrefixes(phones)
	}
	return d
}

func (p *plan) spoke(n Node) proto.NodeConfig {
	c := proto.NodeConfig{
		Node:      proto.Self{ID: n.ID, Name: n.Name, Role: n.Role},
		Interface: proto.Interface{Address: n.prefix(), MTU: n.MTU, FwMark: proto.FwMark},
		Peers:     []proto.Peer{},
	}
	exitOn := false
	for _, s := range p.servers {
		allowed := []string{s.prefix()}
		for _, o := range p.others {
			if o.ID == n.ID {
				continue
			}
			if r, ok := p.relay(n, o); ok && r == s.ID {
				allowed = append(allowed, o.prefix())
			}
		}
		if n.ExitVia == s.ID && s.CanExit {
			allowed = append(allowed, "0.0.0.0/0")
			exitOn = true
		}
		c.Peers = append(c.Peers, proto.Peer{
			Name: s.Name, PublicKey: s.WGPubKey, Endpoint: s.Endpoint.String(),
			AllowedIPs: sortPrefixes(allowed), Keepalive: n.Keepalive,
		})
	}
	sortPeers(c.Peers)
	if len(p.exits) > 0 || exitOn {
		c.Exit = &proto.Exit{Choices: append([]string(nil), p.exits...)}
		if exitOn {
			c.Exit.Via = p.byID[n.ExitVia].Name
		}
	}
	if exitOn {
		// wg-quick's companion setting for fwmark policy routing.
		c.Sysctl = []proto.Sysctl{{Key: proto.SysctlSrcValidMark, Value: "1"}}
	}
	c.DNS = p.dns(n, exitOn)
	return c
}

// phone builds the static profile a phone imports by QR. Its hub peer carries
// the whole mesh CIDR (a static profile cannot learn about nodes added
// later); longest-prefix match still sends traffic for another server
// straight to that server's /32 peer. The via-exit variant is derived from
// this at QR time by adding 0.0.0.0/0 to the chosen exit's peer.
func (p *plan) phone(n Node) proto.NodeConfig {
	c := proto.NodeConfig{
		Node:      proto.Self{ID: n.ID, Name: n.Name, Role: n.Role},
		Interface: proto.Interface{Address: n.prefix(), MTU: n.MTU},
		Peers:     []proto.Peer{},
	}
	hub, _ := p.hub(n)
	for _, s := range p.servers {
		allowed := []string{s.prefix()}
		if s.ID == hub {
			allowed = []string{p.in.MeshCIDR.String()}
		}
		c.Peers = append(c.Peers, proto.Peer{
			Name: s.Name, PublicKey: s.WGPubKey, Endpoint: s.Endpoint.String(),
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
