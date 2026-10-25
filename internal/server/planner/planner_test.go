package planner

import (
	"encoding/json"
	"flag"
	"net/netip"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/hafio/vpn-mesh-utility/internal/proto"
	"github.com/hafio/vpn-mesh-utility/internal/testfix"
)

var update = flag.Bool("update", false, "rewrite testdata/*.golden.json from the current planner (review the diff)")

// node maps a reference-topology node to planner input: enrolled, enabled,
// default MTU and keepalive.
func node(f testfix.Node) Node {
	n := Node{
		ID: f.ID, Name: f.Name, Role: f.Role, MeshIP: netip.MustParseAddr(f.MeshIP),
		WGPubKey: testfix.WGKey(f.Key), ListenPort: proto.DefaultListenPort,
		Keepalive: proto.DefaultKeepalive, MTU: proto.DefaultMTU, Enrolled: true,
	}
	if f.Endpoint != "" {
		n.Endpoint = netip.MustParseAddrPort(f.Endpoint)
	}
	return n
}

// topology builds the named fixture, every one with the default upstream
// resolvers:
//
//	servers   vps-1..3 only
//	full      + vps-2 exit-capable, nas (hub vps-1), laptop (hub vps-2, exit
//	          via vps-2), phone (hub vps-1); custom records "files" (nas),
//	          "printer.office" (a fixed IP) and "work" (laptop); the mesh
//	          firewall of meshFirewall
//	degraded  full with the laptop disabled and vps-3 missing its endpoint
//	containers full with the containers of the nas and vps-2 on the mesh
//	          (host mode, 10.77.128.0/24, 10.77.131.0/24), vps-3 and the
//	          laptop in container mode (10.77.129.0/24, 10.77.130.0/24; the
//	          laptop without its exit)
//	dual      full on a dual-stack mesh (M8, testfix.MeshCIDR6): every node
//	          but the phone (ip_family ipv4) has its IPv6 address, vps-2 an
//	          IPv6 endpoint (so the laptop's exit carries ::/0), the nas's
//	          containers IPv6 too, and a fixed IPv6 record "printer6"
func topology(name string) Input {
	in := Input{MeshCIDR: netip.MustParsePrefix(testfix.MeshCIDR),
		Upstream: []netip.Addr{netip.MustParseAddr("1.1.1.1"), netip.MustParseAddr("9.9.9.9")}}
	vps1, vps2, vps3 := node(testfix.VPS1), node(testfix.VPS2), node(testfix.VPS3)
	if name == "servers" {
		in.Nodes = []Node{vps1, vps2, vps3}
		return in
	}
	in.Records = []Record{
		{Name: "files", Node: testfix.NAS.ID},
		{Name: "printer.office", IP: netip.MustParseAddr("10.77.1.200")},
		{Name: "work", Node: testfix.Laptop.ID},
	}
	in.Segmentation = meshFirewall()
	vps2.CanExit = true
	nas, laptop, phone := node(testfix.NAS), node(testfix.Laptop), node(testfix.Phone)
	nas.PrimaryHub = vps1.ID
	laptop.PrimaryHub, laptop.ExitVia = vps2.ID, vps2.ID
	phone.PrimaryHub = vps1.ID
	switch name {
	case "degraded":
		laptop.Disabled = true
		vps3.Endpoint = netip.AddrPort{}
	case "containers":
		nas.Containers = netip.MustParsePrefix(testfix.ContainerSubnet)
		vps2.Containers = netip.MustParsePrefix("10.77.131.0/24")
		vps3.Mode, vps3.Containers = proto.ModeContainer, netip.MustParsePrefix("10.77.129.0/24")
		laptop.Mode, laptop.Containers, laptop.ExitVia = proto.ModeContainer, netip.MustParsePrefix("10.77.130.0/24"), 0
	case "dual":
		in.MeshCIDR6 = netip.MustParsePrefix(testfix.MeshCIDR6)
		for _, x := range []struct {
			n *Node
			f testfix.Node
		}{{&vps1, testfix.VPS1}, {&vps2, testfix.VPS2}, {&vps3, testfix.VPS3}, {&nas, testfix.NAS}, {&laptop, testfix.Laptop}} {
			x.n.MeshIP6 = netip.MustParseAddr(x.f.MeshIP6)
		}
		vps2.Endpoint6 = netip.MustParseAddrPort(testfix.VPS2.Endpoint6)
		phone.Family = proto.FamilyIPv4
		nas.Containers, nas.Containers6 = netip.MustParsePrefix(testfix.ContainerSubnet), netip.MustParsePrefix(testfix.ContainerSubnet6)
		in.Records = append(in.Records, Record{Name: "printer6", IP6: netip.MustParseAddr("fd10:77::1:200")})
	}
	in.Nodes = []Node{vps1, vps2, vps3, nas, laptop, phone}
	return in
}

func mustCompute(t *testing.T, in Input) map[int64]proto.NodeConfig {
	t.Helper()
	out, _, err := Compute(in)
	if err != nil {
		t.Fatalf("Compute: %v", err)
	}
	return out
}

func encode(t *testing.T, v any) []byte {
	t.Helper()
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	return append(b, '\n')
}

// The goldens were derived by hand from the rules in the package comment;
// regenerate only after deciding the new output is what you wanted
// (./scripts/dev.sh regen).
func TestComputeGolden(t *testing.T) {
	for _, name := range []string{"servers", "full", "degraded", "dual"} {
		t.Run(name, func(t *testing.T) {
			got := encode(t, mustCompute(t, topology(name)))
			path := filepath.Join("testdata", name+".golden.json")
			if *update {
				if err := os.WriteFile(path, got, 0o644); err != nil {
					t.Fatal(err)
				}
				return
			}
			want, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != string(want) {
				t.Fatalf("%s differs from the golden (run ./scripts/dev.sh regen only if the change is intended):\n%s", path, firstDiff(string(want), string(got)))
			}
		})
	}
}

// firstDiff shows the first differing line with a little context.
func firstDiff(want, got string) string {
	w, g := strings.Split(want, "\n"), strings.Split(got, "\n")
	for i := 0; i < len(w) || i < len(g); i++ {
		var wl, gl string
		if i < len(w) {
			wl = w[i]
		}
		if i < len(g) {
			gl = g[i]
		}
		if wl != gl {
			return "line " + strconv.Itoa(i+1) + ":\n  want: " + wl + "\n  got:  " + gl
		}
	}
	return "(no line differs; trailing bytes?)"
}

func TestComputeIsDeterministic(t *testing.T) {
	in := topology("full")
	want := encode(t, mustCompute(t, in))
	perms := [][]int{{5, 4, 3, 2, 1, 0}, {2, 0, 4, 1, 5, 3}, {3, 5, 1, 0, 2, 4}}
	for _, perm := range perms {
		shuffled := in
		shuffled.Nodes = make([]Node, len(perm))
		for i, j := range perm {
			shuffled.Nodes[i] = in.Nodes[j]
		}
		if got := encode(t, mustCompute(t, shuffled)); string(got) != string(want) {
			t.Fatalf("output depends on input order %v", perm)
		}
	}
}

// Every config the planner emits must pass the agent's own validation: no
// prefix on two peers, sane values, exit pointing at a peer with 0.0.0.0/0.
func TestComputedConfigsPassAgentValidation(t *testing.T) {
	for _, name := range []string{"servers", "full", "degraded", "containers", "dual"} {
		for id, c := range mustCompute(t, topology(name)) {
			if err := c.Validate(); err != nil {
				t.Errorf("%s: node %d: %v", name, id, err)
			}
		}
	}
}

func TestServerPeersForSpokesAreExactly32(t *testing.T) {
	in := topology("full")
	out := mustCompute(t, in)
	role := map[string]proto.Role{}
	ips := map[string]string{}
	for _, n := range in.Nodes {
		role[n.Name], ips[n.Name] = n.Role, n.prefix()
	}
	for _, c := range out {
		if c.Node.Role != proto.RoleServer {
			continue
		}
		for _, p := range c.Peers {
			if role[p.Name] == proto.RoleServer {
				continue
			}
			if len(p.AllowedIPs) != 1 || p.AllowedIPs[0] != ips[p.Name] {
				t.Errorf("%s: peer %s AllowedIPs = %v, want exactly [%s] (spec S9: a compromised spoke owns only its /32)", c.Node.Name, p.Name, p.AllowedIPs, ips[p.Name])
			}
			if p.Keepalive != 0 || p.Endpoint != "" {
				t.Errorf("%s: spoke peer %s must have no endpoint or keepalive: %+v", c.Node.Name, p.Name, p)
			}
		}
	}
}

// lookup is WireGuard's cryptokey routing: the peer whose AllowedIPs holds the
// longest prefix containing ip. Used both for "which peer do I send to" and
// "which peer may this source arrive from".
func lookup(c proto.NodeConfig, ip netip.Addr) string {
	best, peer := -1, ""
	for _, p := range c.Peers {
		for _, s := range p.AllowedIPs {
			pf := netip.MustParsePrefix(s)
			if pf.Contains(ip) && pf.Bits() > best {
				best, peer = pf.Bits(), p.Name
			}
		}
	}
	return peer
}

// deliver simulates one packet from -> to through the computed configs and
// returns why it would be dropped, or "".
func deliver(cfgs map[string]proto.NodeConfig, ips map[string]netip.Addr, from, to string) string {
	hop := lookup(cfgs[from], ips[to])
	switch {
	case hop == "":
		return from + " has no route to " + to
	case hop == to:
		if src := lookup(cfgs[to], ips[from]); src != from {
			return to + " drops a direct packet from " + from + " (source maps to peer " + src + ")"
		}
		return ""
	}
	if src := lookup(cfgs[hop], ips[from]); src != from {
		return "relay " + hop + " drops the packet from " + from
	}
	if next := lookup(cfgs[hop], ips[to]); next != to {
		return "relay " + hop + " cannot forward to " + to
	}
	if src := lookup(cfgs[to], ips[from]); src != hop {
		return to + " drops the packet " + from + " relayed via " + hop + " (source maps to peer " + src + ")"
	}
	return ""
}

// TestEveryPairCanTalk is the property behind amendment A3: between any two
// planned nodes, packets get through in both directions -- direct for
// anything involving a server, through one relay for two non-servers -- in
// each family both have (M8).
func TestEveryPairCanTalk(t *testing.T) {
	for _, name := range []string{"full", "degraded", "containers", "dual", "dual/6"} {
		t.Run(name, func(t *testing.T) {
			topo, six := strings.CutSuffix(name, "/6")
			in := topology(topo)
			out := mustCompute(t, in)
			cfgs := map[string]proto.NodeConfig{}
			ips := map[string]netip.Addr{}
			for _, n := range in.Nodes {
				ip := n.MeshIP
				if six {
					ip = n.MeshIP6
				}
				if c, ok := out[n.ID]; ok && ip.IsValid() && (n.Role != proto.RoleServer || n.Endpoint.IsValid() || n.Endpoint6.IsValid()) {
					cfgs[n.Name], ips[n.Name] = c, ip
				}
			}
			for a := range cfgs {
				for b := range cfgs {
					if a == b || (cfgs[a].Node.Role == proto.RolePhone && cfgs[b].Node.Role == proto.RolePhone) {
						continue
					}
					if why := deliver(cfgs, ips, a, b); why != "" {
						t.Errorf("%s -> %s: %s", a, b, why)
					}
				}
			}
		})
	}
}

// A Swarm service resolves through the servers in peer lists, by id; a
// server without an endpoint is in none.
func TestServerIPs(t *testing.T) {
	if got := ServerIPs(topology("full")); !slices.Equal(got, []string{testfix.VPS1.MeshIP, testfix.VPS2.MeshIP, testfix.VPS3.MeshIP}) {
		t.Errorf("full = %v", got)
	}
	if got := ServerIPs(topology("degraded")); !slices.Equal(got, []string{testfix.VPS1.MeshIP, testfix.VPS2.MeshIP}) {
		t.Errorf("degraded (vps-3 without an endpoint) = %v", got)
	}
	if got := ServerIPs(Input{MeshCIDR: netip.MustParsePrefix(testfix.MeshCIDR)}); got != nil {
		t.Errorf("no servers = %v", got)
	}
}

// withDown marks the server id down (amendment A40).
func withDown(in Input, id int64) Input {
	in.Nodes = append([]Node(nil), in.Nodes...)
	for i := range in.Nodes {
		if in.Nodes[i].ID == id {
			in.Nodes[i].Down = true
		}
	}
	return in
}

// Relay failover (amendment A40): with any one server down, every pair
// still gets through, no pair of spokes relays through the down server, and
// the choice is the same from both ends. A phone's pairs keep the phone's
// hub: its profile names that one.
func TestRelayFailover(t *testing.T) {
	for _, name := range []string{"full", "containers"} {
		for _, down := range []testfix.Node{testfix.VPS1, testfix.VPS2, testfix.VPS3} {
			t.Run(name+"/"+down.Name, func(t *testing.T) {
				in := withDown(topology(name), down.ID)
				out := mustCompute(t, in)
				cfgs := map[string]proto.NodeConfig{}
				ips := map[string]netip.Addr{}
				for _, n := range in.Nodes {
					cfgs[n.Name], ips[n.Name] = out[n.ID], n.MeshIP
				}
				for a := range cfgs {
					for b := range cfgs {
						ra, rb := cfgs[a].Node.Role, cfgs[b].Node.Role
						if a == b || (ra == proto.RolePhone && rb == proto.RolePhone) {
							continue
						}
						if why := deliver(cfgs, ips, a, b); why != "" {
							t.Errorf("%s -> %s: %s", a, b, why)
						}
						if hop := lookup(cfgs[a], ips[b]); ra == proto.RoleSpoke && rb == proto.RoleSpoke && hop == down.Name {
							t.Errorf("%s -> %s relays through the down %s", a, b, down.Name)
						}
					}
				}
				p := newPlan(in)
				for _, x := range p.others {
					for _, y := range p.others {
						rx, _ := p.relay(x, y, f4)
						ry, _ := p.relay(y, x, f4)
						if rx != ry {
							t.Errorf("relay(%s, %s) = %d but relay(%s, %s) = %d", x.Name, y.Name, rx, y.Name, x.Name, ry)
						}
					}
				}
			})
		}
	}
	// Without a phone on it, a down hub relays nothing, so it stops
	// forwarding; the warning names it; when every server is down nothing
	// moves.
	in := topology("full")
	in.Nodes = in.Nodes[:5]
	if !mustCompute(t, in)[testfix.VPS1.ID].Firewall.Relay {
		t.Fatal("vps-1 relays nas <-> laptop while it is up")
	}
	in = withDown(in, testfix.VPS1.ID)
	out, problems, err := Compute(in)
	if err != nil || out[testfix.VPS1.ID].Firewall.Relay || !out[testfix.VPS2.ID].Firewall.Relay {
		t.Fatalf("vps-1 down: relay vps-1 %v, vps-2 %v (%v)", out[testfix.VPS1.ID].Firewall.Relay, out[testfix.VPS2.ID].Firewall.Relay, err)
	}
	if !hasProblem(problems, "vps-1", "is down", true) {
		t.Errorf("no warning for the down hub: %v", problems)
	}
	all := withDown(withDown(withDown(topology("full"), testfix.VPS1.ID), testfix.VPS2.ID), testfix.VPS3.ID)
	p := newPlan(all)
	if r, ok := p.relay(p.byID[testfix.NAS.ID], p.byID[testfix.Laptop.ID], f4); !ok || r != testfix.VPS1.ID {
		t.Errorf("every server down: relay = %d, %v; want the usual vps-1", r, ok)
	}
	// The default hub comes before the lowest-id server once both ends'
	// hubs are down.
	two := withDown(withDown(topology("full"), testfix.VPS1.ID), testfix.VPS2.ID)
	two.DefaultHub = testfix.VPS3.ID
	p = newPlan(two)
	if r, _ := p.relay(p.byID[testfix.NAS.ID], p.byID[testfix.Laptop.ID], f4); r != testfix.VPS3.ID {
		t.Errorf("both hubs down: relay = %d, want the default hub vps-3", r)
	}
}

// deliverAddr is deliver with addresses of the nodes' own choosing: a
// container's address behind a node is routed wherever the node is.
func deliverAddr(cfgs map[string]proto.NodeConfig, from string, src netip.Addr, to string, dst netip.Addr) string {
	hop := lookup(cfgs[from], dst)
	switch {
	case hop == "":
		return from + " has no route to " + dst.String()
	case hop == to:
		if back := lookup(cfgs[to], src); back != from {
			return to + " drops a direct packet from " + src.String() + " (source maps to peer " + back + ")"
		}
		return ""
	}
	if back := lookup(cfgs[hop], src); back != from {
		return "relay " + hop + " drops the packet from " + src.String()
	}
	if next := lookup(cfgs[hop], dst); next != to {
		return "relay " + hop + " cannot forward to " + dst.String()
	}
	if back := lookup(cfgs[to], src); back != hop {
		return to + " drops the packet from " + src.String() + " relayed via " + hop + " (source maps to peer " + back + ")"
	}
	return ""
}

// A container behind any node reaches every other node, and is reached
// from it, in both directions -- the property amendment A37 adds to A3.
func TestContainersTalkToEveryNode(t *testing.T) {
	in := topology("containers")
	out := mustCompute(t, in)
	cfgs := map[string]proto.NodeConfig{}
	ips := map[string]netip.Addr{}
	behind := map[string]netip.Addr{} // one container address per node with containers
	for _, n := range in.Nodes {
		cfgs[n.Name], ips[n.Name] = out[n.ID], n.MeshIP
		if n.Containers.IsValid() {
			a := n.Containers.Addr().As4()
			a[3] = 130
			behind[n.Name] = netip.AddrFrom4(a)
		}
	}
	for host, c := range behind {
		for other := range cfgs {
			if other == host {
				continue
			}
			if why := deliverAddr(cfgs, other, ips[other], host, c); why != "" {
				t.Errorf("%s -> container %s behind %s: %s", other, c, host, why)
			}
			if why := deliverAddr(cfgs, host, c, other, ips[other]); why != "" {
				t.Errorf("container %s behind %s -> %s: %s", c, host, other, why)
			}
		}
	}
}

// What a container network adds to each config (amendment A37): every peer
// routes the subnet with the node's /32 and nothing more, the node's
// resolver forwards for it, and the container network itself; a
// container-mode node has no firewall section, no exit and no custom rules,
// a host-mode server's firewall forwards to its bridge.
// A mesh in slot 3 (amendment A49): every config carries the slot, and a
// host-mode server's container bridge is the slot's; nothing else changes.
func TestSlotInEveryConfig(t *testing.T) {
	in := topology("containers")
	base := mustCompute(t, in)
	in.Slot = 3
	out := mustCompute(t, in)
	hash := func(c proto.NodeConfig) string {
		h, err := c.Hash()
		if err != nil {
			t.Fatal(err)
		}
		return h
	}
	for id, c := range out {
		if c.Node.Slot != 3 {
			t.Errorf("%s's config has slot %d", c.Node.Name, c.Node.Slot)
		}
		want := base[id]
		want.Node.Slot = 3
		if want.Firewall != nil && want.Firewall.Bridge != "" {
			fw := *want.Firewall
			fw.Bridge = "meshctl3"
			want.Firewall = &fw
		}
		if hash(c) != hash(want) {
			t.Errorf("%s's slot 3 config differs from slot 0's in more than the slot and the bridge", c.Node.Name)
		}
	}
}

func TestContainerNetworks(t *testing.T) {
	in := topology("containers")
	out := mustCompute(t, in)
	nasNet, vps3Net := testfix.ContainerSubnet, "10.77.129.0/24"
	peer := func(c proto.NodeConfig, name string) proto.Peer {
		for _, p := range c.Peers {
			if p.Name == name {
				return p
			}
		}
		t.Fatalf("%s has no peer %s", c.Node.Name, name)
		return proto.Peer{}
	}
	vps1, vps2, vps3 := out[testfix.VPS1.ID], out[testfix.VPS2.ID], out[testfix.VPS3.ID]
	nas, laptop, phone := out[testfix.NAS.ID], out[testfix.Laptop.ID], out[testfix.Phone.ID]
	for _, c := range []struct {
		where string
		got   []string
		want  []string
	}{
		{"vps-1's peer nas", peer(vps1, "nas").AllowedIPs, []string{testfix.NAS.MeshIP + "/32", nasNet}},
		{"vps-1's peer vps-3", peer(vps1, "vps-3").AllowedIPs, []string{testfix.VPS3.MeshIP + "/32", vps3Net}},
		{"nas's peer vps-3", peer(nas, "vps-3").AllowedIPs, []string{testfix.VPS3.MeshIP + "/32", vps3Net}},
		{"nas's peer vps-1 (relays the laptop and the phone)", peer(nas, "vps-1").AllowedIPs,
			[]string{testfix.VPS1.MeshIP + "/32", testfix.Laptop.MeshIP + "/32", testfix.Phone.MeshIP + "/32", "10.77.130.0/24"}},
		{"the phone's peer vps-3", peer(phone, "vps-3").AllowedIPs, []string{testfix.VPS3.MeshIP + "/32", vps3Net}},
		{"nas's peer vps-2", peer(nas, "vps-2").AllowedIPs, []string{testfix.VPS2.MeshIP + "/32", "10.77.131.0/24"}},
		{"the phone's hub vps-1", peer(phone, "vps-1").AllowedIPs, []string{testfix.MeshCIDR}},
		// Every resolver forwards for every container network: a Swarm
		// service's tasks share one resolver list.
		{"nas's forward_for", nas.DNS.ForwardFor, []string{nasNet, vps3Net, "10.77.130.0/24", "10.77.131.0/24"}},
		{"vps-3's forward_for", vps3.DNS.ForwardFor, []string{nasNet, vps3Net, "10.77.130.0/24", "10.77.131.0/24"}},
		{"vps-1's forward_for", vps1.DNS.ForwardFor, []string{testfix.Phone.MeshIP + "/32", nasNet, vps3Net, "10.77.130.0/24", "10.77.131.0/24"}},
	} {
		if !slices.Equal(c.got, c.want) {
			t.Errorf("%s = %v, want %v", c.where, c.got, c.want)
		}
	}
	if k := nas.Containers; k == nil || *k != *testfix.Containers(proto.ModeHost) {
		t.Errorf("nas's container network = %+v, want %+v", k, testfix.Containers(proto.ModeHost))
	}
	if k := vps3.Containers; k == nil || k.Mode != proto.ModeContainer || k.Gateway != "10.77.129.2" || k.Bridge != "10.77.129.1" ||
		k.Range != "10.77.129.128/25" || k.Network != testfix.MeshCIDR {
		t.Errorf("vps-3's container network = %+v", k)
	}
	if vps1.Containers != nil {
		t.Errorf("vps-1 has no containers on the mesh, got %+v", vps1.Containers)
	}
	sysctls := func(c proto.NodeConfig) []string {
		var out []string
		for _, s := range c.Sysctl {
			out = append(out, s.Key+"="+s.Value)
		}
		return out
	}
	// The bridge's own sysctls are the agent's, once the bridge exists.
	if got, want := sysctls(nas), []string{proto.SysctlIPForward + "=1"}; !slices.Equal(got, want) {
		t.Errorf("nas's sysctls = %v, want %v", got, want)
	}
	if fw := vps2.Firewall; fw == nil || fw.Bridge != proto.ContainerBridge || !fw.Relay {
		t.Errorf("vps-2's firewall = %+v, want the bridge forwarded to", fw)
	}
	if fw := vps1.Firewall; fw == nil || fw.Bridge != "" {
		t.Errorf("vps-1 has no containers on the mesh, so no bridge: %+v", fw)
	}
	for _, c := range []proto.NodeConfig{vps3, laptop} {
		if got, want := sysctls(c), []string{proto.SysctlProxyARP + "=1", proto.SysctlIPForward + "=1"}; !slices.Equal(got, want) {
			t.Errorf("%s's sysctls = %v, want %v", c.Node.Name, got, want)
		}
		if c.Firewall != nil {
			t.Errorf("%s is in container mode, whose isolation is the agent's routing, but has a firewall: %+v", c.Node.Name, c.Firewall)
		}
		if c.Exit != nil {
			t.Errorf("%s is in container mode but has an exit section: %+v", c.Node.Name, c.Exit)
		}
	}
	if nas.Firewall != nil {
		t.Errorf("a host-mode spoke has no firewall meshctl manages, got %+v", nas.Firewall)
	}
	// A container-mode hub that relays forwards between its peers.
	if k := vps3.Containers; k == nil || k.Relay {
		t.Errorf("vps-3 relays for no one, yet its container network says %+v", k)
	}
	in.Nodes = slices.Clone(in.Nodes)
	in.Nodes[3].PrimaryHub = testfix.VPS3.ID
	if k := mustCompute(t, in)[testfix.VPS3.ID].Containers; k == nil || !k.Relay {
		t.Errorf("vps-3 relays for the nas, yet its container network says %+v", k)
	}
}

// Spokes whose primary hubs differ are the case the spec's original
// per-spoke-hub rule got wrong; pin it explicitly.
func TestSpokesWithDifferentHubsShareOneRelay(t *testing.T) {
	out := mustCompute(t, topology("full"))
	nas, laptop := out[testfix.NAS.ID], out[testfix.Laptop.ID]
	if hop := lookup(nas, netip.MustParseAddr(testfix.Laptop.MeshIP)); hop != "vps-1" {
		t.Errorf("nas -> laptop goes via %q, want vps-1 (hub of the lower id)", hop)
	}
	if hop := lookup(laptop, netip.MustParseAddr(testfix.NAS.MeshIP)); hop != "vps-1" {
		t.Errorf("laptop -> nas goes via %q, want vps-1 even though the laptop's hub is vps-2", hop)
	}
}

func TestExitRouting(t *testing.T) {
	out := mustCompute(t, topology("full"))
	internet := netip.MustParseAddr("8.8.8.8")
	if hop := lookup(out[testfix.Laptop.ID], internet); hop != "vps-2" {
		t.Errorf("laptop internet traffic goes via %q, want vps-2", hop)
	}
	if hop := lookup(out[testfix.NAS.ID], internet); hop != "" {
		t.Errorf("nas has no exit but internet traffic goes via %q", hop)
	}
	if out[testfix.Laptop.ID].Exit == nil || out[testfix.Laptop.ID].Exit.Via != "vps-2" {
		t.Errorf("laptop exit = %+v, want via vps-2", out[testfix.Laptop.ID].Exit)
	}
}

func TestRelayAndExitServersForward(t *testing.T) {
	out := mustCompute(t, topology("full"))
	for _, c := range []struct {
		id          int64
		relay, exit bool
	}{{testfix.VPS1.ID, true, false}, {testfix.VPS2.ID, false, true}, {testfix.VPS3.ID, false, false}} {
		cfg := out[c.id]
		if cfg.Firewall.Relay != c.relay || cfg.Firewall.ExitGateway != c.exit {
			t.Errorf("%s: relay=%v exit=%v, want relay=%v exit=%v", cfg.Node.Name, cfg.Firewall.Relay, cfg.Firewall.ExitGateway, c.relay, c.exit)
		}
		forwards := len(cfg.Sysctl) == 1 && cfg.Sysctl[0] == proto.Sysctl{Key: proto.SysctlIPForward, Value: "1"}
		if forwards != (c.relay || c.exit) {
			t.Errorf("%s: ip_forward=%v, want %v", cfg.Node.Name, forwards, c.relay || c.exit)
		}
	}
}

func TestUnreachableHubFallsBack(t *testing.T) {
	in := topology("full")
	for i := range in.Nodes {
		switch in.Nodes[i].Name {
		case "vps-3":
			in.Nodes[i].Disabled = true
		case "nas":
			in.Nodes[i].PrimaryHub = testfix.VPS3.ID
		}
	}
	out, problems, err := Compute(in)
	if err != nil {
		t.Fatal(err)
	}
	if hop := lookup(out[testfix.NAS.ID], netip.MustParseAddr(testfix.Laptop.MeshIP)); hop != "vps-1" {
		t.Errorf("nas falls back via %q, want the lowest-id server vps-1", hop)
	}
	if !hasProblem(problems, "nas", "not reachable", true) {
		t.Errorf("want a warning about the unreachable hub, got %v", problems)
	}
}

func hasProblem(ps []Problem, node, substr string, warning bool) bool {
	for _, p := range ps {
		if p.Node == node && p.Warning == warning && strings.Contains(p.Message, substr) {
			return true
		}
	}
	return false
}

func TestEndpointlessServerWarns(t *testing.T) {
	_, problems, err := Compute(topology("degraded"))
	if err != nil {
		t.Fatal(err)
	}
	if !hasProblem(problems, "vps-3", "--endpoint", true) {
		t.Fatalf("want an actionable endpoint warning for vps-3, got %v", problems)
	}
}

func TestCheckRejects(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*Input)
		node   string
		want   string
	}{
		{"duplicate name", func(in *Input) { in.Nodes[1].Name = "vps-1" }, "vps-1", "two nodes"},
		{"duplicate ip", func(in *Input) { in.Nodes[3].MeshIP = in.Nodes[0].MeshIP }, "nas", "also assigned"},
		{"duplicate key", func(in *Input) { in.Nodes[3].WGPubKey = in.Nodes[0].WGPubKey }, "nas", "same WireGuard key"},
		{"ip outside mesh", func(in *Input) { in.Nodes[3].MeshIP = netip.MustParseAddr("192.168.1.5") }, "nas", "outside the mesh"},
		{"mesh too big", func(in *Input) { in.MeshCIDR = netip.MustParsePrefix("10.0.0.0/8") }, "", "/16 to /22"},
		{"hub not a server", func(in *Input) { in.Nodes[3].PrimaryHub = testfix.Laptop.ID }, "nas", "not a server"},
		{"server with exit", func(in *Input) { in.Nodes[0].ExitVia = testfix.VPS2.ID }, "vps-1", "do not use an exit"},
		{"exit not capable", func(in *Input) { in.Nodes[4].ExitVia = testfix.VPS3.ID }, "laptop", "not exit-capable"},
		{"phone with exit", func(in *Input) { in.Nodes[5].ExitVia = testfix.VPS2.ID }, "phone", "via-exit QR profile"},
		{"spoke can exit", func(in *Input) { in.Nodes[3].CanExit = true }, "nas", "only servers"},
		{"mtu", func(in *Input) { in.Nodes[3].MTU = 9000 }, "nas", "MTU"},
		{"keepalive", func(in *Input) { in.Nodes[3].Keepalive = -1 }, "nas", "keepalive"},
		{"extra domain in the zone", func(in *Input) { in.Domains = []string{"lab.mesh"} }, "", "the mesh's zone"},
		{"extra domains overlap", func(in *Input) { in.Domains = []string{"lan", "x.lan"} }, "", "inside lan"},
		{"listen port", func(in *Input) { in.Nodes[0].ListenPort = 0 }, "vps-1", "listen port"},
		{"bad name", func(in *Input) { in.Nodes[3].Name = "NAS box" }, "NAS box", "name"},
		{"unknown role", func(in *Input) { in.Nodes[3].Role = "router" }, "nas", "unknown role"},
		{"default hub", func(in *Input) { in.DefaultHub = testfix.NAS.ID }, "", "default hub"},
		{"record named like a node", func(in *Input) { in.Records[0].Name = "nas" }, "", "has the name of a node"},
		{"record defined twice", func(in *Input) { in.Records[1].Name = "files" }, "", "defined twice"},
		{"record name", func(in *Input) { in.Records[0].Name = "-files" }, "", "dns record"},
		{"record outside the mesh", func(in *Input) { in.Records[1].IP = netip.MustParseAddr("192.168.1.10") }, "", "outside the mesh"},
		{"record without an ip", func(in *Input) { in.Records[1].IP = netip.Addr{} }, "", "has no address"},
		{"record with both targets", func(in *Input) { in.Records[1].Node = testfix.NAS.ID }, "", "both a node and an IP"},
		{"record of a missing node", func(in *Input) { in.Records[0].Node = 99 }, "", "does not exist"},
		{"upstream in the mesh", func(in *Input) { in.Upstream = []netip.Addr{netip.MustParseAddr("10.77.0.1")} }, "", "would loop"},
		{"upstream mapped", func(in *Input) { in.Upstream = []netip.Addr{netip.MustParseAddr("::ffff:1.1.1.1")} }, "", "not an IPv4 or IPv6"},
		{"zone", func(in *Input) { in.Zone = "office.local" }, "", "dns zone"},
		{"record with the zone", func(in *Input) { in.Records[0].Name = "files.mesh" }, "", "ends with the zone"},
		{"unknown mode", func(in *Input) { in.Nodes[3].Mode = "bridge" }, "nas", "unknown mode"},
		{"container mode without a subnet", func(in *Input) { in.Nodes[3].Mode = proto.ModeContainer }, "nas", "no container subnet"},
		{"phone with containers", func(in *Input) { in.Nodes[5].Containers = netip.MustParsePrefix("10.77.131.0/24") }, "phone", "no agent"},
		{"container-mode exit server", func(in *Input) {
			in.Nodes[1].Mode, in.Nodes[1].Containers = proto.ModeContainer, netip.MustParsePrefix("10.77.131.0/24")
		}, "vps-2", "cannot be an exit"},
		{"container-mode spoke with an exit", func(in *Input) {
			in.Nodes[4].Mode, in.Nodes[4].Containers = proto.ModeContainer, netip.MustParsePrefix("10.77.131.0/24")
		}, "laptop", "cannot use an exit"},
		{"container subnet not a /24", func(in *Input) { in.Nodes[3].Containers = netip.MustParsePrefix("10.77.128.0/23") }, "nas", "IPv4 /24"},
		{"container subnet outside the mesh", func(in *Input) { in.Nodes[3].Containers = netip.MustParsePrefix("10.88.0.0/24") }, "nas", "outside the mesh"},
		{"container subnet twice", func(in *Input) {
			in.Nodes[3].Containers = netip.MustParsePrefix(testfix.ContainerSubnet)
			in.Nodes[4].Containers = netip.MustParsePrefix(testfix.ContainerSubnet)
		}, "laptop", "is also \"nas\"'s"},
		{"node inside a container subnet", func(in *Input) { in.Nodes[3].Containers = netip.MustParsePrefix("10.77.2.0/24") }, "laptop", "inside \"nas\"'s container subnet"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			in := topology("full")
			in.Nodes = append([]Node(nil), in.Nodes...)
			c.mutate(&in)
			out, problems, err := Compute(in)
			if err == nil || out != nil {
				t.Fatalf("Compute accepted an invalid plan (err=%v, %d configs)", err, len(out))
			}
			if !hasProblem(problems, c.node, c.want, false) {
				t.Fatalf("want an error for %q mentioning %q, got %v", c.node, c.want, problems)
			}
		})
	}
}

// The IPv6 side refuses what an agent could not apply (M8).
func TestCheckRejectsIPv6(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*Input)
		node   string
		want   string
	}{
		{"prefix too big", func(in *Input) { in.MeshCIDR6 = netip.MustParsePrefix("fd10::/32") }, "", "/48 to /64"},
		{"no prefix", func(in *Input) { in.MeshCIDR6 = netip.Prefix{} }, "nas", "the mesh has no IPv6 prefix"},
		{"address outside", func(in *Input) { in.Nodes[3].MeshIP6 = netip.MustParseAddr("fd10:78::1") }, "nas", "outside the mesh's"},
		{"address twice", func(in *Input) { in.Nodes[3].MeshIP6 = in.Nodes[0].MeshIP6 }, "nas", "also assigned"},
		{"address in the container part", func(in *Input) { in.Nodes[4].MeshIP6 = netip.MustParseAddr("fd10:77::9000:1") }, "laptop", "container networks"},
		{"ipv4 node with IPv6", func(in *Input) { in.Nodes[3].Family = proto.FamilyIPv4 }, "nas", "IPv4-only"},
		{"ipv6 node with IPv4", func(in *Input) { in.Nodes[3].Family = proto.FamilyIPv6 }, "nas", "IPv6-only"},
		{"unknown family", func(in *Input) { in.Nodes[3].Family = "both" }, "nas", "unknown ip_family"},
		{"no address at all", func(in *Input) { in.Nodes[4].MeshIP, in.Nodes[4].MeshIP6 = netip.Addr{}, netip.Addr{} }, "laptop", "no mesh address"},
		{"endpoint6 on a spoke", func(in *Input) { in.Nodes[3].Endpoint6 = netip.MustParseAddrPort("[2001:db8::9]:51820") }, "nas", "only servers"},
		{"container subnet6 size", func(in *Input) { in.Nodes[3].Containers6 = netip.MustParsePrefix("fd10:77::8000:0/111") }, "nas", "IPv6 /112"},
		{"container subnet6 outside", func(in *Input) { in.Nodes[3].Containers6 = netip.MustParsePrefix("fd10:77::5:0/112") }, "nas", "outside the container part"},
		{"container subnet6 alone", func(in *Input) { in.Nodes[4].Containers6 = netip.MustParsePrefix("fd10:77::8001:0/112") }, "laptop", "no IPv4 container network"},
		{"container subnet6 twice", func(in *Input) {
			in.Nodes[1].Containers, in.Nodes[1].Containers6 = netip.MustParsePrefix("10.77.131.0/24"), in.Nodes[3].Containers6
		}, "nas", "is also \"vps-2\"'s"},
		{"record6 outside", func(in *Input) { in.Records[3].IP6 = netip.MustParseAddr("fd99::1") }, "", "outside the mesh's IPv6 prefix"},
		{"record6 without a prefix", func(in *Input) {
			in.MeshCIDR6, in.Records[3].IP6 = netip.Prefix{}, netip.MustParseAddr("fd10:77::1:9")
		}, "", "but the mesh has no IPv6 prefix"},
		{"upstream in the IPv6 mesh", func(in *Input) { in.Upstream = []netip.Addr{netip.MustParseAddr("fd10:77::1")} }, "", "would loop"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			in := topology("dual")
			in.Nodes = append([]Node(nil), in.Nodes...)
			c.mutate(&in)
			out, problems, err := Compute(in)
			if err == nil || out != nil {
				t.Fatalf("Compute accepted an invalid plan (err=%v, %d configs)", err, len(out))
			}
			if !hasProblem(problems, c.node, c.want, false) {
				t.Fatalf("want an error for %q mentioning %q, got %v", c.node, c.want, problems)
			}
		})
	}
}

// The dual-stack plan in detail (M8): endpoints by family with the
// fallback, ::/0 only through an exit with public IPv6, an IPv6-only
// laptop on the one server it can dial, the exit's IPv6 masquerade.
// A family a node's agent reports its host cannot use is planned away from
// that node, and never refuses the plan (M8).
func TestReportedHostFamilies(t *testing.T) {
	in := topology("dual")
	in.Nodes = append([]Node(nil), in.Nodes...)
	in.Nodes[3].NoIPv6 = true // nas
	in.Nodes[4].NoIPv4 = true // laptop
	in.Nodes[0].NoIPv4 = true // vps-1, a server, goes IPv6 only
	out, problems, err := Compute(in)
	if err != nil {
		t.Fatal(err)
	}
	for node, want := range map[string]string{"nas": "cannot use IPv6; the mesh leaves IPv6 out", "laptop": "cannot use IPv4; the mesh leaves IPv4 out",
		"vps-1": "cannot use IPv4; the mesh leaves IPv4 out"} {
		if !hasProblem(problems, node, want, true) {
			t.Errorf("no warning %q for %s: %v", want, node, problems)
		}
	}
	if !in.Nodes[3].MeshIP6.IsValid() || !in.Nodes[4].MeshIP.IsValid() {
		t.Fatal("Compute changed the caller's nodes")
	}
	nas, laptop, vps1 := out[testfix.NAS.ID], out[testfix.Laptop.ID], out[testfix.VPS1.ID]
	if nas.Interface.Address6 != "" || nas.Interface.Address != testfix.NAS.MeshIP+"/32" || nas.Containers == nil || nas.Containers.Subnet6 != "" {
		t.Errorf("nas without IPv6: %+v, containers %+v", nas.Interface, nas.Containers)
	}
	if laptop.Interface.Address != "" || laptop.Interface.Address6 != testfix.Laptop.MeshIP6+"/128" {
		t.Errorf("laptop without IPv4: %+v", laptop.Interface)
	}
	if vps1.Interface.Address != "" || vps1.Interface.Address6 != testfix.VPS1.MeshIP6+"/128" {
		t.Errorf("vps-1 without IPv4: %+v", vps1.Interface)
	}
	for id, c := range out {
		for _, p := range c.Peers {
			for _, a := range p.AllowedIPs {
				if a == testfix.NAS.MeshIP6+"/128" || a == testfix.ContainerSubnet6 || a == testfix.Laptop.MeshIP+"/32" || a == testfix.VPS1.MeshIP+"/32" {
					t.Errorf("node %d routes %s to %s", id, a, p.Name)
				}
			}
		}
	}
	for _, r := range out[testfix.VPS2.ID].DNS.Records {
		if (r.Name == "nas" && (r.IP6 != "" || r.IP == "")) || (r.Name == "laptop" && (r.IP != "" || r.IP6 == "")) {
			t.Errorf("record %+v", r)
		}
	}

	// A node that cannot do without the family is planned as configured.
	in = topology("dual")
	in.Nodes = append([]Node(nil), in.Nodes...)
	// The IPv4-only phone has nothing to leave out; an IPv6-only laptop and
	// the nas, which has a container network, cannot do without theirs.
	in.Nodes[5].NoIPv6 = true
	in.Nodes[4].MeshIP, in.Nodes[4].Family, in.Nodes[4].NoIPv6 = netip.Addr{}, proto.FamilyIPv6, true
	in.Nodes[3].NoIPv4 = true
	out, problems, err = Compute(in)
	if err != nil {
		t.Fatal(err)
	}
	if !hasProblem(problems, "laptop", "planned as configured", true) || !hasProblem(problems, "nas", "planned as configured", true) ||
		hasProblem(problems, "phone", "cannot use", true) {
		t.Errorf("problems: %v", problems)
	}
	if out[testfix.Laptop.ID].Interface.Address6 == "" || out[testfix.NAS.ID].Interface.Address == "" {
		t.Error("a node that cannot do without the family lost it")
	}
}

func TestDualStack(t *testing.T) {
	in := topology("dual")
	out := mustCompute(t, in)
	peer := func(c proto.NodeConfig, name string) proto.Peer {
		t.Helper()
		for _, p := range c.Peers {
			if p.Name == name {
				return p
			}
		}
		t.Fatalf("%s has no peer %s", c.Node.Name, name)
		return proto.Peer{}
	}
	laptop := out[testfix.Laptop.ID]
	vps2 := peer(laptop, "vps-2")
	if laptop.Interface.Address6 != testfix.Laptop.MeshIP6+"/128" || vps2.Endpoint != testfix.VPS2.Endpoint ||
		vps2.Fallback != testfix.VPS2.Endpoint6 || !slices.Contains(vps2.AllowedIPs, "::/0") || !slices.Contains(vps2.AllowedIPs, "0.0.0.0/0") {
		t.Fatalf("the laptop: %+v; vps-2 %+v", laptop.Interface, vps2)
	}
	if p := peer(laptop, "vps-1"); p.Fallback != "" || !slices.Contains(p.AllowedIPs, testfix.NAS.MeshIP6+"/128") ||
		!slices.Contains(p.AllowedIPs, testfix.ContainerSubnet6) {
		t.Fatalf("the laptop's relay for the nas, both families: %+v", p)
	}
	server := out[testfix.VPS2.ID]
	if server.Firewall.Exit6 != testfix.MeshCIDR6 || !slices.Contains(server.Sysctl, proto.Sysctl{Key: proto.SysctlIPv6Forward, Value: "1"}) {
		t.Fatalf("vps-2, an exit with IPv6: %+v, %v", server.Firewall, server.Sysctl)
	}
	if p := peer(server, "laptop"); !slices.Equal(p.AllowedIPs, []string{testfix.Laptop.MeshIP + "/32", testfix.Laptop.MeshIP6 + "/128"}) {
		t.Fatalf("a server routes a spoke's two addresses only: %v", p.AllowedIPs)
	}
	phone := out[testfix.Phone.ID]
	if phone.Interface.Address6 != "" || slices.Contains(peer(phone, "vps-1").AllowedIPs, testfix.MeshCIDR6) {
		t.Fatalf("an IPv4-only phone got IPv6: %+v", phone)
	}
	for _, r := range out[testfix.VPS1.ID].DNS.Records {
		if r.Name == "laptop" && (r.IP != testfix.Laptop.MeshIP || r.IP6 != testfix.Laptop.MeshIP6) ||
			r.Name == "phone" && r.IP6 != "" || r.Name == "printer6" && (r.IP != "" || r.IP6 != "fd10:77::1:200") {
			t.Errorf("record %+v", r)
		}
	}

	// IPv6 only: the laptop dials vps-2's IPv6 endpoint, the one there is,
	// and relays through it; the others are not its peers.
	in.Nodes = append([]Node(nil), in.Nodes...)
	in.Nodes[4].Family, in.Nodes[4].MeshIP = proto.FamilyIPv6, netip.Addr{}
	out, _, err := Compute(in)
	if err != nil {
		t.Fatal(err)
	}
	laptop = out[testfix.Laptop.ID]
	if len(laptop.Peers) != 1 || laptop.Peers[0].Endpoint != testfix.VPS2.Endpoint6 || laptop.Peers[0].Fallback != "" ||
		!slices.Contains(laptop.Peers[0].AllowedIPs, testfix.NAS.MeshIP6+"/128") || slices.Contains(laptop.Peers[0].AllowedIPs, "0.0.0.0/0") {
		t.Fatalf("an IPv6-only laptop: %+v", laptop.Peers)
	}
	if p := peer(out[testfix.NAS.ID], "vps-2"); !slices.Contains(p.AllowedIPs, testfix.Laptop.MeshIP6+"/128") {
		t.Fatalf("the nas must relay to the laptop where the laptop can be: %v", p.AllowedIPs)
	}
	for _, c := range out {
		if err := c.Validate(); err != nil {
			t.Errorf("%s: %v", c.Node.Name, err)
		}
	}
	// Without any server it can dial, it is warned.
	in.Nodes[1].Endpoint6 = netip.AddrPort{}
	if _, problems, _ := Compute(in); !hasProblem(problems, "laptop", "can reach none of the servers", true) {
		t.Fatalf("no warning: %v", problems)
	}
}

// The mesh zone and who may forward through a node's resolver (spec S7.4,
// amendment A12).
func TestDNS(t *testing.T) {
	full, degraded := mustCompute(t, topology("full")), mustCompute(t, topology("degraded"))
	names := func(d *proto.DNS) string {
		var out []string
		for _, r := range d.Records {
			out = append(out, r.Name+"="+r.IP)
		}
		return strings.Join(out, " ")
	}
	want := "files=10.77.1.1 laptop=10.77.2.1 nas=10.77.1.1 phone=10.77.3.1 printer.office=10.77.1.200 vps-1=10.77.0.1 " +
		"vps-2=10.77.0.2 vps-3=10.77.0.3 work=10.77.2.1"
	for id, c := range full {
		if c.Node.Role == proto.RolePhone {
			if c.DNS != nil {
				t.Errorf("a phone has no agent, so no resolver: %+v", c.DNS)
			}
			continue
		}
		if got := names(c.DNS); got != want {
			t.Errorf("node %d serves %s, want every active node and custom record: %s", id, got, want)
		}
		if strings.Join(c.DNS.Upstream, ",") != "1.1.1.1,9.9.9.9" {
			t.Errorf("node %d upstream = %v, want the operator's order", id, c.DNS.Upstream)
		}
	}
	// Every config names the default zone; another zone set for the mesh
	// reaches every node.
	for _, c := range full {
		if c.DNS != nil && c.DNS.Zone != proto.DefaultZone {
			t.Errorf("zone %q, want the default %q", c.DNS.Zone, proto.DefaultZone)
		}
	}
	in := topology("full")
	in.Zone = "home.arpa"
	for _, c := range mustCompute(t, in) {
		if c.DNS != nil && c.DNS.Zone != "home.arpa" {
			t.Errorf("zone %q with the mesh set to home.arpa", c.DNS.Zone)
		}
	}
	// The extra domains (amendment A39) reach every resolver, sorted; with
	// none the section has none, so it hashes as before.
	for _, c := range full {
		if c.DNS != nil && c.DNS.Domains != nil {
			t.Errorf("domains %v without any set", c.DNS.Domains)
		}
	}
	in = topology("full")
	in.Domains = []string{"home.arpa", "example.lan"}
	for _, c := range mustCompute(t, in) {
		if c.DNS != nil && strings.Join(c.DNS.Domains, ",") != "example.lan,home.arpa" {
			t.Errorf("%s: domains %v", c.Node.Name, c.DNS.Domains)
		}
	}
	// A disabled node's name, and an alias of it, resolve no more.
	if got := names(degraded[testfix.NAS.ID].DNS); strings.Contains(got, "laptop") || strings.Contains(got, "work") {
		t.Errorf("degraded zone still names the disabled laptop: %s", got)
	}
	// Only exit mode forwards a spoke's other lookups; only the phone's hub
	// forwards for it.
	if !full[testfix.NAS.ID].DNS.Split || full[testfix.Laptop.ID].DNS.Split {
		t.Errorf("split: nas %v, laptop (exit via vps-2) %v", full[testfix.NAS.ID].DNS.Split, full[testfix.Laptop.ID].DNS.Split)
	}
	for _, s := range []testfix.Node{testfix.VPS1, testfix.VPS2, testfix.VPS3} {
		d := full[s.ID].DNS
		fwd := ""
		if s.ID == testfix.VPS1.ID {
			fwd = testfix.Phone.MeshIP + "/32"
		}
		if !d.Split || strings.Join(d.ForwardFor, ",") != fwd {
			t.Errorf("%s: split %v, forward_for %v, want %q", s.Name, d.Split, d.ForwardFor, fwd)
		}
	}
}

// Every server's mesh zone lets the mesh through and carries the one
// compiled policy (amendment A43); no other node gets a policy.
func TestFirewallRules(t *testing.T) {
	out := mustCompute(t, topology("full"))
	want := []proto.FWRule{
		{Zone: proto.ZoneMesh, Proto: proto.ProtoAll, Source: testfix.MeshCIDR, Note: "segmentation: the policy decides"},
		{Zone: proto.ZonePublic, Proto: "udp", Port: proto.DefaultListenPort, Note: "wireguard"},
	}
	pol := CompilePolicy(topology("full"))
	for _, n := range topology("full").Nodes {
		c := out[n.ID]
		if n.Role != proto.RoleServer {
			if c.Policy != nil {
				t.Errorf("%s (a %s) has a policy", n.Name, n.Role)
			}
			continue
		}
		if !slices.Equal(c.Firewall.Rules, want) {
			t.Errorf("%s's firewall rules = %+v, want %+v", n.Name, c.Firewall.Rules, want)
		}
		if !reflect.DeepEqual(c.Policy, pol) {
			t.Errorf("%s's policy is not the mesh's", n.Name)
		}
	}
	// A container-mode server has no firewalld, but enforces the policy.
	k := mustCompute(t, topology("containers"))[testfix.VPS3.ID]
	if k.Firewall != nil || !reflect.DeepEqual(k.Policy, CompilePolicy(topology("containers"))) {
		t.Errorf("container-mode vps-3: firewall %+v, policy %v", k.Firewall, k.Policy != nil)
	}
}

func TestProblemString(t *testing.T) {
	if s := (Problem{Node: "nas", Message: "x"}).String(); s != `error: node "nas": x` {
		t.Errorf("String = %q", s)
	}
	if s := (Problem{Message: "y", Warning: true}).String(); s != "warning: y" {
		t.Errorf("String = %q", s)
	}
}
