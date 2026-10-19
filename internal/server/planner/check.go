package planner

import (
	"fmt"
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
		if !n.MeshIP.Is4() {
			errf(n.Name, "has no IPv4 mesh IP")
		} else {
			if in.MeshCIDR.IsValid() && !in.MeshCIDR.Contains(n.MeshIP) {
				errf(n.Name, "mesh IP %s is outside the mesh %s", n.MeshIP, in.MeshCIDR)
			}
			if other, dup := ips[n.MeshIP.String()]; dup {
				errf(n.Name, "mesh IP %s is also assigned to %q", n.MeshIP, other)
			}
			ips[n.MeshIP.String()] = n.Name
		}
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
	}

	isServer := func(id int64) bool {
		s, ok := byID[id]
		return ok && s.Role == proto.RoleServer
	}
	if in.DefaultHub != 0 && !isServer(in.DefaultHub) {
		errf("", "the default hub (node id %d) is not a server", in.DefaultHub)
	}
	checkDNS(in, names, byID, errf)
	for _, r := range in.Rules {
		what := fmt.Sprintf("firewall rule %s/%d from %s", r.Proto, r.Port, r.Source)
		if r.Node != 0 && !isServer(r.Node) {
			errf("", "%s is for node id %d, which is not a server; only servers have a firewall meshctl manages", what, r.Node)
		}
		if r.Proto != "tcp" && r.Proto != "udp" {
			errf("", "%s: protocol must be tcp or udp", what)
		}
		if r.Port < 1 || r.Port > 65535 {
			errf("", "%s: port is outside 1-65535", what)
		}
		if !r.Source.IsValid() || !r.Source.Addr().Is4() || r.Source.Masked() != r.Source {
			errf("", "%s: the source must be an IPv4 network such as 10.77.0.0/16", what)
		}
		if _, err := validate.Text("note", r.Note, 120); err != nil {
			errf("", "%s: %v", what, err)
		}
	}

	p := newPlan(in)
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
			if !n.Endpoint.IsValid() {
				warnf(n.Name, "has no public endpoint, so no other node peers with it yet; set it with `meshctl node set %s --endpoint IP:PORT`", n.Name)
			}
		default:
			hub, ok := p.hub(n)
			switch {
			case !ok:
				warnf(n.Name, "there is no reachable server to peer with yet")
			case n.PrimaryHub != 0 && hub != n.PrimaryHub:
				warnf(n.Name, "primary hub %q is not reachable; relaying through %q until it is back", byID[n.PrimaryHub].Name, byID[hub].Name)
			}
		}
	}
	return out
}

// checkDNS checks the zone, the custom records and the upstream
// resolvers. Node names are records already, so a custom record may not
// take one; record names are given without the zone.
func checkDNS(in Input, nodeNames map[string]bool, byID map[int64]Node, errf func(node, format string, args ...any)) {
	zone := in.ZoneName()
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
		case r.Node != 0 && r.IP.IsValid():
			errf("", "DNS record %q points at both a node and an IP; keep one", r.Name)
		case r.Node != 0 && !known:
			errf("", "DNS record %q points at node id %d, which does not exist", r.Name, r.Node)
		case r.Node != 0:
		case !r.IP.Is4():
			errf("", "DNS record %q has no IPv4 address", r.Name)
		case in.MeshCIDR.IsValid() && !in.MeshCIDR.Contains(r.IP):
			errf("", "DNS record %q points at %s, outside the mesh %s; records name mesh addresses only", r.Name, r.IP, in.MeshCIDR)
		}
	}
	for _, u := range in.Upstream {
		switch {
		case !u.Is4():
			errf("", "upstream DNS %s is not an IPv4 address", u)
		case in.MeshCIDR.IsValid() && in.MeshCIDR.Contains(u):
			errf("", "upstream DNS %s is inside the mesh %s; a mesh resolver forwarding to another would loop", u, in.MeshCIDR)
		}
	}
}

func sortedByID(ns []Node) []Node {
	out := append([]Node(nil), ns...)
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}
