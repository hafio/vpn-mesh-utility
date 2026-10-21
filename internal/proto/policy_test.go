package proto_test

import (
	"net/netip"
	"strings"
	"testing"

	"github.com/hafio/vpn-mesh-utility/internal/proto"
	"github.com/hafio/vpn-mesh-utility/internal/testfix"
)

// First match wins, the rest is dropped, and traffic with an end outside
// the mesh is not the policy's.
func TestPolicyDecide(t *testing.T) {
	m, err := testfix.Policy().Matcher()
	if err != nil {
		t.Fatal(err)
	}
	a := netip.MustParseAddr
	for _, c := range []struct {
		name     string
		src, dst string
		proto    string
		port     int
		allow    bool
		rule     int
	}{
		{"ping anywhere", testfix.Laptop.MeshIP, testfix.VPS1.MeshIP, "icmp", 0, true, 0},
		{"dns to a server", testfix.Laptop.MeshIP, testfix.VPS2.MeshIP, "udp", 53, true, 1},
		{"dns to a spoke is no system rule", testfix.Laptop.MeshIP, testfix.NAS.MeshIP, "udp", 53, true, 5},
		{"swarm between servers", testfix.VPS1.MeshIP, testfix.VPS3.MeshIP, "tcp", 2377, true, 2},
		{"swarm from a spoke", testfix.NAS.MeshIP, testfix.VPS3.MeshIP, "tcp", 2377, false, proto.RuleDefault},
		{"smb before the deny", testfix.Laptop.MeshIP, testfix.NAS.MeshIP, "tcp", 445, true, 3},
		{"the laptop denied ssh to a server", testfix.Laptop.MeshIP, testfix.VPS1.MeshIP, "tcp", 22, false, 4},
		{"within the zone", testfix.NAS.MeshIP, testfix.Laptop.MeshIP, "udp", 9000, true, 5},
		{"nothing matches", testfix.Phone.MeshIP, testfix.NAS.MeshIP, "tcp", 80, false, proto.RuleDefault},
		{"internet through an exit", testfix.Laptop.MeshIP, "1.1.1.1", "tcp", 443, true, proto.RuleNotMesh},
		{"from outside the mesh", "192.168.1.5", testfix.NAS.MeshIP, "tcp", 22, true, proto.RuleNotMesh},
	} {
		allow, rule := m.Decide(a(c.src), a(c.dst), c.proto, c.port)
		if allow != c.allow || rule != c.rule {
			t.Errorf("%s: Decide = %v, rule %d; want %v, rule %d", c.name, allow, rule, c.allow, c.rule)
		}
	}
}

func TestPolicyMatcherRefusesBadValues(t *testing.T) {
	for name, c := range map[string]struct {
		edit func(*proto.Policy)
		want string
	}{
		"mesh":             {func(p *proto.Policy) { p.Mesh = "10.77.0.1/16" }, "policy.mesh"},
		"action":           {func(p *proto.Policy) { p.Rules[3].Action = "accept" }, "policy.rules[3].action"},
		"name newline":     {func(p *proto.Policy) { p.Rules[3].Name = "x\ny" }, "policy.rules[3].name"},
		"address":          {func(p *proto.Policy) { p.Rules[3].From = []string{"10.77.2.1;drop"} }, "policy.rules[3].from[0]"},
		"outside the mesh": {func(p *proto.Policy) { p.Rules[3].To = []string{"192.168.0.0/24"} }, "outside the mesh"},
		"protocol":         {func(p *proto.Policy) { p.Rules[3].Services[0].Proto = "sctp" }, "policy.rules[3].services[0]"},
		"port":             {func(p *proto.Policy) { p.Rules[3].Services[0].Ports = []string{"445-1"} }, "start at or below"},
		"icmp port":        {func(p *proto.Policy) { p.Rules[0].Services[0].Ports = []string{"8"} }, "icmp"},
		"too many rules": {func(p *proto.Policy) {
			for len(p.Rules) <= proto.MaxPolicyRules {
				p.Rules = append(p.Rules, p.Rules[0])
			}
		}, "at most 512"},
		"too many entries": {func(p *proto.Policy) {
			for len(p.Rules[3].From) <= proto.MaxPolicyEntries {
				p.Rules[3].From = append(p.Rules[3].From, testfix.Laptop.MeshIP)
			}
		}, "at most 1024"},
	} {
		p := testfix.Policy()
		c.edit(p)
		if _, err := p.Matcher(); err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("%s: Matcher = %v, want an error about %q", name, err, c.want)
		}
	}
}

func TestPolicyRuleString(t *testing.T) {
	p := testfix.Policy()
	for i, want := range map[int]string{
		0: "allow any -> any icmp (system: ping)",
		3: "allow 10.77.2.1 -> 10.77.1.1 tcp/445 (policy: laptop-smb)",
		4: "deny 10.77.2.1 -> 10.77.0.1-10.77.0.3 any (policy: laptop-servers)",
	} {
		if got := p.Rules[i].String(); got != want {
			t.Errorf("rule %d = %q, want %q", i, got, want)
		}
	}
}
