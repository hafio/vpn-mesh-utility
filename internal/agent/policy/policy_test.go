package policy_test

import (
	"errors"
	"flag"
	"net/netip"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hafio/vpn-mesh-utility/internal/agent/agentfake"
	"github.com/hafio/vpn-mesh-utility/internal/agent/policy"
	"github.com/hafio/vpn-mesh-utility/internal/proto"
	"github.com/hafio/vpn-mesh-utility/internal/testfix"
	"github.com/hafio/vpn-mesh-utility/internal/validate"
)

var update = flag.Bool("update", false, "rewrite testdata/ruleset.golden.txt (review the diff)")

// vps1 is the reference server without containers; vps2 hosts the
// containers of testfix.ContainerSubnet.
var (
	vps1 = policy.Local{Self: netip.MustParseAddr(testfix.VPS1.MeshIP)}
	vps2 = policy.Local{Self: netip.MustParseAddr(testfix.VPS2.MeshIP), Containers: netip.MustParsePrefix(testfix.ContainerSubnet)}
)

func compile(t *testing.T, p *proto.Policy, l policy.Local) *policy.Ruleset {
	t.Helper()
	rs, err := policy.Compile(*p, l)
	if err != nil {
		t.Fatal(err)
	}
	return rs
}

// The table for vps-1 and for vps-2 with its containers, in nft's spelling.
func TestCompileGolden(t *testing.T) {
	got := compile(t, testfix.Policy(), vps1).Text() + "\n" + compile(t, testfix.Policy(), vps2).Text()
	path := filepath.Join("testdata", "ruleset.golden.txt")
	if *update {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if got != string(want) {
		t.Fatalf("%s differs (run ./scripts/dev.sh regen only if the change is intended):\n%s", path, got)
	}
}

// packet is one new packet, as the kernel sees it at a hook.
type packet struct {
	src, dst uint32
	proto    string
	port     int
	iif, oif string
}

func (p packet) matches(r policy.Rule) bool {
	in := func(rs []policy.Range, v uint32) bool {
		if rs == nil {
			return true
		}
		for _, x := range rs {
			if x.First <= v && v <= x.Last {
				return true
			}
		}
		return false
	}
	switch {
	case r.IIF != "" && r.IIF != p.iif, r.OIF != "" && r.OIF != p.oif:
		return false
	case r.State != "": // a new packet is neither established nor invalid
		return false
	case !in(r.Src, p.src) || !in(r.Dst, p.dst):
		return false
	case r.Proto != "" && r.Proto != p.proto:
		return false
	case r.Ports != nil && !in(r.Ports, uint32(p.port)):
		return false
	}
	return !r.Echo || p.proto == validate.ServiceICMP // our icmp packets are echo requests
}

// walk is the verdict of chain for p: true accepts. A chain that runs out
// returns to its caller; a base chain then accepts.
func walk(rs *policy.Ruleset, name string, p packet) (bool, bool) {
	for _, c := range rs.Chains {
		if c.Name != name {
			continue
		}
		for _, r := range c.Rules {
			if !p.matches(r) {
				continue
			}
			switch r.Verdict {
			case policy.Accept:
				return true, true
			case policy.Drop:
				return false, true
			case policy.Jump:
				if v, ok := walk(rs, r.Target, p); ok {
					return v, true
				}
			}
		}
	}
	return false, false
}

func admit(rs *policy.Ruleset, hook string, p packet) bool {
	for _, c := range rs.Chains {
		if c.Hook == hook {
			if v, ok := walk(rs, c.Name, p); ok {
				return v
			}
			return true // the base chain's policy
		}
	}
	panic("no chain for " + hook)
}

// The table decides every packet as the policy does (proto.Matcher, which
// the server's reachability check uses): forwarded, received and sent, for
// every pair of addresses, protocol and port -- except traffic within the
// node, which is not the mesh firewall's.
func TestCompileAgreesWithDecide(t *testing.T) {
	pol := testfix.Policy()
	m, err := pol.Matcher()
	if err != nil {
		t.Fatal(err)
	}
	pool := []string{testfix.VPS1.MeshIP, testfix.VPS2.MeshIP, testfix.VPS3.MeshIP, testfix.NAS.MeshIP, testfix.Laptop.MeshIP,
		testfix.Phone.MeshIP, "10.77.128.5", "10.77.128.200", "10.77.200.1", "192.168.1.5", "1.1.1.1"}
	checked := 0
	for _, l := range []policy.Local{vps1, vps2} {
		rs := compile(t, pol, l)
		for _, s := range pool {
			for _, d := range pool {
				src, dst := netip.MustParseAddr(s), netip.MustParseAddr(d)
				if src == dst {
					continue
				}
				for _, pr := range []string{"tcp", "udp", "icmp"} {
					for _, port := range []int{22, 53, 445, 2377, 4789, 9000} {
						p := packet{src: validate.U32(src), dst: validate.U32(dst), proto: pr, port: port, iif: "wg0", oif: "wg0"}
						hook := policy.HookForward
						switch {
						case dst == l.Self:
							hook = policy.HookInput
						case src == l.Self:
							hook = policy.HookOutput
						}
						want, _ := m.Decide(src, dst, pr, port)
						inNode := func(a netip.Addr) bool {
							return a == l.Self || (l.Containers.IsValid() && l.Containers.Contains(a))
						}
						if inNode(src) && inNode(dst) {
							want = true // within the node
						}
						if got := admit(rs, hook, p); got != want {
							t.Fatalf("self %s: %s -> %s %s/%d at %s: table %v, policy %v\n%s", l.Self, s, d, pr, port, hook, got, want, rs.Text())
						}
						checked++
					}
				}
			}
		}
	}
	if checked < 1000 {
		t.Fatalf("only %d packets checked", checked)
	}
	// Loopback is the node itself.
	rs := compile(t, pol, vps1)
	if !admit(rs, policy.HookInput, packet{src: validate.U32(vps1.Self), dst: validate.U32(vps1.Self), proto: "tcp", port: 22, iif: "lo"}) {
		t.Fatal("loopback was filtered")
	}
}

func TestCompileRefuses(t *testing.T) {
	for name, c := range map[string]struct {
		edit func(*proto.Policy)
		self string
		want string
	}{
		"mesh":             {func(p *proto.Policy) { p.Mesh = "10.77.0.1/16" }, testfix.VPS1.MeshIP, "policy.mesh"},
		"self outside":     {func(*proto.Policy) {}, "192.168.1.5", "outside the mesh"},
		"address":          {func(p *proto.Policy) { p.Rules[3].From = []string{"10.77.2.1;drop"} }, testfix.VPS1.MeshIP, "policy.rules[3].from[0]"},
		"service":          {func(p *proto.Policy) { p.Rules[3].Services[0].Proto = "sctp" }, testfix.VPS1.MeshIP, "not a service"},
		"port":             {func(p *proto.Policy) { p.Rules[3].Services[0].Ports = []string{"0"} }, testfix.VPS1.MeshIP, "between 1 and 65535"},
		"action":           {func(p *proto.Policy) { p.Rules[3].Action = "accept" }, testfix.VPS1.MeshIP, "neither allow nor deny"},
		"icmp with a port": {func(p *proto.Policy) { p.Rules[0].Services[0].Ports = []string{"8"} }, testfix.VPS1.MeshIP, "not a service"},
	} {
		p := testfix.Policy()
		c.edit(p)
		if _, err := policy.Compile(*p, policy.Local{Self: netip.MustParseAddr(c.self)}); err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("%s: Compile = %v, want an error about %q", name, err, c.want)
		}
	}
}

// Sync installs a ruleset once, leaves the same one alone, replaces a
// different one (re-judging the open mesh flows), removes the table for
// nil, and says which kernel call failed.
func TestEnforcer(t *testing.T) {
	k := &agentfake.Kernel{Forgotten: 3}
	e := &policy.Enforcer{K: k}
	a := compile(t, testfix.Policy(), vps1)
	other := testfix.Policy()
	other.Rules = other.Rules[:4]
	b := compile(t, other, vps1)
	step := func(rs *policy.Ruleset, changes, replaces, forgets int) {
		t.Helper()
		n, err := e.Sync(rs)
		if err != nil || n != changes || k.Replaces != replaces || k.Forgets != forgets {
			t.Fatalf("Sync = %d, %v; %d replaces, %d forgets; want %d, %d, %d", n, err, k.Replaces, k.Forgets, changes, replaces, forgets)
		}
	}
	step(a, 1, 1, 0) // a first table re-checks no flow: there was no policy before
	if table, mark := k.Holds(); table != a || mark != a.Fingerprint() {
		t.Fatalf("the kernel holds %v marked %q", table, mark)
	}
	step(a, 0, 1, 0)
	step(b, 1, 2, 1)
	k.Mark = "unmarked" // someone else's table under our name
	step(b, 1, 3, 2)
	k.ForgetErr = errors.New("conntrack: operation not permitted")
	step(a, 1, 4, 3) // the flows keep going; the table is in
	step(nil, 1, 4, 3)
	if table, mark := k.Holds(); table != nil || mark != "" {
		t.Fatal("the table was not removed")
	}
	step(nil, 0, 4, 3)

	for name, c := range map[string]struct {
		fail func()
		rs   *policy.Ruleset
		want string
	}{
		"read":    {func() { k.ReadErr = errors.New("permission denied") }, a, "read the table meshctl"},
		"replace": {func() { k.ReplaceErr = errors.New("no buffer space") }, a, "install the table meshctl"},
		"remove":  {func() { k.Mark, k.RemoveErr = "x", errors.New("busy") }, nil, "remove the table meshctl"},
	} {
		k.ReadErr, k.ReplaceErr, k.RemoveErr = nil, nil, nil
		c.fail()
		if _, err := e.Sync(c.rs); err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("%s: Sync = %v, want %q", name, err, c.want)
		}
	}
	if !strings.Contains(policy.ErrNoKernel.Error(), "Linux nftables") {
		t.Error("ErrNoKernel does not say what is missing")
	}
}

func TestRulesetText(t *testing.T) {
	rs := compile(t, testfix.Policy(), vps2)
	if rs.Count() == 0 || len(rs.Fingerprint()) != 16 || rs.Fingerprint() == compile(t, testfix.Policy(), vps1).Fingerprint() {
		t.Fatalf("count %d, fingerprint %q", rs.Count(), rs.Fingerprint())
	}
	r := policy.Rule{IIF: "lo", OIF: "wg0", State: policy.StateInvalid, Src: []policy.Range{{1, 1}}, Proto: "udp",
		Ports: []policy.Range{{53, 53}, {1000, 2000}}, Verdict: policy.Jump, Target: "x"}
	want := `iifname "lo" oifname "wg0" ct state invalid ip saddr { 0.0.0.1 } meta l4proto udp th dport { 53, 1000-2000 } jump x`
	if got := r.String(); got != want {
		t.Fatalf("String =\n%s\nwant\n%s", got, want)
	}
}

// A rule's addresses and ports are merged even when the config's are not
// (overlapping, touching, inside one another), and tcp without ports is
// every port.
func TestCompileMerges(t *testing.T) {
	p := &proto.Policy{Mesh: testfix.MeshCIDR, Rules: []proto.PolicyRule{{Name: "merge",
		From:     []string{"10.77.2.3", "10.77.2.1-10.77.2.5", "10.77.2.6"},
		Services: []proto.Service{{Proto: "tcp"}, {Proto: "udp", Ports: []string{"53", "50-60", "61"}}}, Action: proto.PolicyAllow}}}
	text := compile(t, p, vps1).Text()
	for _, want := range []string{
		"  ip saddr { 10.77.2.1-10.77.2.6 } meta l4proto tcp accept  # merge\n",
		"  ip saddr { 10.77.2.1-10.77.2.6 } meta l4proto udp th dport { 50-61 } accept  # merge\n",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("the table lacks %q:\n%s", want, text)
		}
	}
}
