package planner

import (
	"encoding/json"
	"flag"
	"net/netip"
	"os"
	"path/filepath"
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
//	          "printer.office" (a fixed IP) and "work" (laptop); custom rules
//	          9100/tcp from the mesh on every server and 5432/tcp from the nas
//	          on vps-1
//	degraded  full with the laptop disabled and vps-3 missing its endpoint
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
	in.Rules = []Rule{
		{ID: 1, Proto: "tcp", Port: 9100, Source: in.MeshCIDR, Note: "node exporter"},
		{ID: 2, Node: testfix.VPS1.ID, Proto: "tcp", Port: 5432, Source: netip.MustParsePrefix(testfix.NAS.MeshIP + "/32")},
	}
	vps2.CanExit = true
	nas, laptop, phone := node(testfix.NAS), node(testfix.Laptop), node(testfix.Phone)
	nas.PrimaryHub = vps1.ID
	laptop.PrimaryHub, laptop.ExitVia = vps2.ID, vps2.ID
	phone.PrimaryHub = vps1.ID
	if name == "degraded" {
		laptop.Disabled = true
		vps3.Endpoint = netip.AddrPort{}
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
	for _, name := range []string{"servers", "full", "degraded"} {
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
	for _, name := range []string{"servers", "full", "degraded"} {
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
// anything involving a server, through one relay for two non-servers.
func TestEveryPairCanTalk(t *testing.T) {
	for _, name := range []string{"full", "degraded"} {
		t.Run(name, func(t *testing.T) {
			in := topology(name)
			out := mustCompute(t, in)
			cfgs := map[string]proto.NodeConfig{}
			ips := map[string]netip.Addr{}
			for _, n := range in.Nodes {
				if c, ok := out[n.ID]; ok && (n.Role != proto.RoleServer || n.Endpoint.IsValid()) {
					cfgs[n.Name], ips[n.Name] = c, n.MeshIP
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
		{"listen port", func(in *Input) { in.Nodes[0].ListenPort = 0 }, "vps-1", "listen port"},
		{"bad name", func(in *Input) { in.Nodes[3].Name = "NAS box" }, "NAS box", "name"},
		{"unknown role", func(in *Input) { in.Nodes[3].Role = "router" }, "nas", "unknown role"},
		{"default hub", func(in *Input) { in.DefaultHub = testfix.NAS.ID }, "", "default hub"},
		{"record named like a node", func(in *Input) { in.Records[0].Name = "nas" }, "", "has the name of a node"},
		{"record defined twice", func(in *Input) { in.Records[1].Name = "files" }, "", "defined twice"},
		{"record name", func(in *Input) { in.Records[0].Name = "-files" }, "", "dns record"},
		{"record outside the mesh", func(in *Input) { in.Records[1].IP = netip.MustParseAddr("192.168.1.10") }, "", "outside the mesh"},
		{"record without an ip", func(in *Input) { in.Records[1].IP = netip.Addr{} }, "", "no IPv4 address"},
		{"record with both targets", func(in *Input) { in.Records[1].Node = testfix.NAS.ID }, "", "both a node and an IP"},
		{"record of a missing node", func(in *Input) { in.Records[0].Node = 99 }, "", "does not exist"},
		{"rule on a spoke", func(in *Input) { in.Rules[1].Node = testfix.NAS.ID }, "", "not a server"},
		{"rule protocol", func(in *Input) { in.Rules[0].Proto = "icmp" }, "", "tcp or udp"},
		{"rule port", func(in *Input) { in.Rules[0].Port = 0 }, "", "outside 1-65535"},
		{"rule source", func(in *Input) { in.Rules[0].Source = netip.MustParsePrefix("10.77.1.1/16") }, "", "IPv4 network"},
		{"rule note", func(in *Input) { in.Rules[0].Note = "a\nb" }, "", "control characters"},
		{"upstream in the mesh", func(in *Input) { in.Upstream = []netip.Addr{netip.MustParseAddr("10.77.0.1")} }, "", "would loop"},
		{"upstream not ipv4", func(in *Input) { in.Upstream = []netip.Addr{netip.MustParseAddr("2001:db8::1")} }, "", "not an IPv4"},
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

// Every server opens the resolver to the mesh; a custom rule lands on the
// servers it names, with where it came from.
func TestFirewallRules(t *testing.T) {
	out := mustCompute(t, topology("full"))
	has := func(c proto.NodeConfig, want proto.FWRule) bool {
		return slices.Contains(c.Firewall.Rules, want)
	}
	for _, s := range []testfix.Node{testfix.VPS1, testfix.VPS2, testfix.VPS3} {
		c := out[s.ID]
		for _, p := range []string{"udp", "tcp"} {
			if !has(c, proto.FWRule{Zone: proto.ZoneMesh, Proto: p, Port: 53, Source: testfix.MeshCIDR, Note: "dns"}) {
				t.Errorf("%s does not open 53/%s to the mesh", s.Name, p)
			}
		}
		if !has(c, proto.FWRule{Zone: proto.ZoneMesh, Proto: "tcp", Port: 9100, Source: testfix.MeshCIDR, Note: "custom: node exporter"}) {
			t.Errorf("%s misses the rule for every server", s.Name)
		}
		pg := proto.FWRule{Zone: proto.ZoneMesh, Proto: "tcp", Port: 5432, Source: testfix.NAS.MeshIP + "/32", Note: "custom rule 2"}
		if has(c, pg) != (s.ID == testfix.VPS1.ID) {
			t.Errorf("%s: the rule for vps-1 only: present=%v", s.Name, has(c, pg))
		}
	}
	// Two rules that differ only in their note keep one order.
	in := topology("full")
	in.Rules = append(in.Rules, Rule{ID: 3, Proto: "tcp", Port: 9100, Source: in.MeshCIDR, Note: "another exporter"})
	want := encode(t, mustCompute(t, in))
	in.Rules[0], in.Rules[2] = in.Rules[2], in.Rules[0]
	if got := encode(t, mustCompute(t, in)); string(got) != string(want) {
		t.Fatal("rule order depends on input order")
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
