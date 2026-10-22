package planner

import (
	"reflect"
	"testing"
	"time"

	"github.com/hafio/vpn-mesh-utility/internal/proto"
	"github.com/hafio/vpn-mesh-utility/internal/testfix"
)

// meshFirewall is the mesh firewall of topology("full"): the custom rule of
// before as migration 0006 makes it (9100/tcp from anywhere to every
// server), the laptop allowed SMB into the home zone (the nas and the
// printer), the phones denied it, allow-all last; and a policy left
// disabled.
func meshFirewall() Segmentation {
	return Segmentation{
		Groups:    map[string][]string{"laptops": {"laptop"}, "storage": {"nas"}},
		Addresses: map[string][]string{"printer": {"10.77.1.200"}},
		Services:  map[string][]string{"smb": {"tcp/445", "tcp/139"}},
		Zones:     []SegZone{{Name: "home", Intra: proto.PolicyAllow, Members: []string{"group:storage", "addr:printer"}}},
		Policies: []SegPolicy{
			{Name: "firewall-1", From: []string{"0.0.0.0/0"}, To: []string{"role:server"}, Services: []string{"tcp/9100"}, Action: proto.PolicyAllow},
			{Name: "laptop-smb", From: []string{"group:laptops"}, To: []string{"zone:home"}, Services: []string{"svc:smb"}, Action: proto.PolicyAllow},
			{Name: "phones-out", From: []string{"role:phone"}, To: []string{"zone:home"}, Action: proto.PolicyDeny},
			{Name: "later", Disabled: true, From: []string{"node:nas"}, Action: proto.PolicyDeny},
			{Name: "allow-all", Action: proto.PolicyAllow},
		},
		TimeZone: "UTC",
	}
}

// The compiled policy, derived by hand from compile's rules: system rules,
// the enabled policies in order with their references resolved (anywhere
// clipped to the mesh, the services merged), each allowing zone.
func TestPolicyCompiles(t *testing.T) {
	servers := []string{"10.77.0.1-10.77.0.3"}
	home := []string{"10.77.1.1", "10.77.1.200"}
	svc := func(name string, ports ...string) proto.Service { return proto.Service{Proto: name, Ports: ports} }
	want := &proto.Policy{Mesh: testfix.MeshCIDR, Rules: []proto.PolicyRule{
		{Name: "system: ping", Services: []proto.Service{svc("icmp")}, Action: "allow"},
		{Name: "system: dns", To: servers, Services: []proto.Service{svc("tcp", "53"), svc("udp", "53")}, Action: "allow"},
		{Name: "system: swarm", From: servers, To: servers, Services: []proto.Service{svc("tcp", "2377", "7946"), svc("udp", "4789", "7946")}, Action: "allow"},
		{Name: "policy: firewall-1", From: []string{testfix.MeshCIDR}, To: servers, Services: []proto.Service{svc("tcp", "9100")}, Action: "allow"},
		{Name: "policy: laptop-smb", From: []string{testfix.Laptop.MeshIP}, To: home, Services: []proto.Service{svc("tcp", "139", "445")}, Action: "allow"},
		{Name: "policy: phones-out", From: []string{testfix.Phone.MeshIP}, To: home, Action: "deny"},
		{Name: "policy: allow-all", Action: "allow"},
		{Name: "zone: home", From: home, To: home, Action: "allow"},
	}}
	got := CompilePolicy(topology("full"))
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("CompilePolicy =\n%s\nwant\n%s", encode(t, got), encode(t, want))
	}
	if _, err := got.Matcher(); err != nil {
		t.Fatalf("the compiled policy does not pass the agent's check: %v", err)
	}
	// Without any segmentation the servers keep the system rules only.
	bare := CompilePolicy(topology("servers"))
	if len(bare.Rules) != 3 || bare.Rules[2].Name != "system: swarm" {
		t.Fatalf("a mesh without a firewall compiles to %+v", bare.Rules)
	}
	if !hasProblem(Check(topology("servers")), "", "no policy or zone allows any traffic", true) {
		t.Fatal("no warning for a mesh that denies everything")
	}
}

// Merged addresses and ports: overlapping and touching ranges become one,
// a range that is a network reads as one, every port wins over ports.
func TestPolicyMerges(t *testing.T) {
	in := topology("full")
	in.Segmentation.Policies = []SegPolicy{{Name: "p", From: []string{"10.77.4.0-10.77.4.127", "10.77.4.128/25", "10.77.4.10"},
		To: []string{"10.77.5.1-10.77.5.10", "10.77.5.5-10.77.5.20"}, Services: []string{"udp/10-20", "udp/15-30", "tcp/80", "tcp", "icmp"},
		Action: "allow"}}
	r := CompilePolicy(in).Rules[3]
	if !reflect.DeepEqual(r.From, []string{"10.77.4.0/24"}) || !reflect.DeepEqual(r.To, []string{"10.77.5.1-10.77.5.20"}) {
		t.Errorf("addresses: from %v, to %v", r.From, r.To)
	}
	want := []proto.Service{{Proto: "icmp"}, {Proto: "tcp"}, {Proto: "udp", Ports: []string{"10-30"}}}
	if !reflect.DeepEqual(r.Services, want) {
		t.Errorf("services = %+v, want %+v", r.Services, want)
	}
}

func TestCheckSegmentation(t *testing.T) {
	mon := &proto.Schedule{Days: []string{"mon"}}
	for _, c := range []struct {
		name string
		edit func(*Segmentation)
		want string
	}{
		{"zone member of no node", func(s *Segmentation) { s.Zones[0].Members = []string{"node:ghost"} }, `there is no node "ghost"`},
		{"group of nobody", func(s *Segmentation) { s.Policies[1].From = []string{"group:ghost"} }, `there is no group "ghost"`},
		{"zone of nothing", func(s *Segmentation) { s.Policies[1].To = []string{"zone:ghost"} }, `there is no zone "ghost"`},
		{"address object", func(s *Segmentation) { s.Zones[0].Members = []string{"addr:ghost"} }, `there is no address object "ghost"`},
		{"service object", func(s *Segmentation) { s.Policies[1].Services = []string{"svc:ghost"} }, `there is no service object "ghost"`},
		{"role", func(s *Segmentation) { s.Policies[2].From = []string{"role:admin"} }, "the roles are"},
		{"group lists a stranger", func(s *Segmentation) { s.Groups["laptops"] = []string{"ghost"} }, `group "laptops": there is no node "ghost"`},
		{"literal", func(s *Segmentation) { s.Zones[0].Members = []string{"10.77.1.5;reboot"} }, "members[0]"},
		{"service", func(s *Segmentation) { s.Policies[0].Services = []string{"tcp/0"} }, "between 1 and 65535"},
		{"object entry", func(s *Segmentation) { s.Services["smb"] = []string{"tcp/smb"} }, `service object "smb": entries[0]`},
		{"zone within", func(s *Segmentation) { s.Zones[0].Intra = "maybe" }, `zone "home": intra`},
		{"policy action", func(s *Segmentation) { s.Policies[0].Action = "accept" }, `policy "firewall-1": action`},
		{"a disabled policy's reference", func(s *Segmentation) { s.Policies[3].From = []string{"node:ghost"} }, `policy "later" (disabled)`},
		{"a zone inside a zone", func(s *Segmentation) { s.Zones[0].Members = []string{"zone:home"} }, "kind one of"},
		{"a schedule", func(s *Segmentation) { s.Policies[1].Schedule = &proto.Schedule{Start: "9:00", End: "18:00"} }, `policy "laptop-smb": schedule.start`},
		{"a schedule without a time zone", func(s *Segmentation) { s.TimeZone, s.Policies[1].Schedule = "", mon }, "the policies' time zone"},
		{"a disabled policy's schedule", func(s *Segmentation) { s.Policies[3].Schedule = &proto.Schedule{} }, `policy "later" (disabled): schedule`},
	} {
		in := topology("full")
		in.Segmentation = meshFirewall()
		c.edit(&in.Segmentation)
		if !hasProblem(Check(in), "", c.want, false) {
			t.Errorf("%s: Check = %v, want an error about %q", c.name, Check(in), c.want)
		}
		if _, _, err := Compute(in); err == nil {
			t.Errorf("%s: Compute planned a mesh with a broken firewall", c.name)
		}
	}
}

func TestSegmentationWarnings(t *testing.T) {
	in := topology("full")
	in.Segmentation.Groups["empty"] = nil
	in.Segmentation.Addresses["lan"] = []string{"192.168.0.0/24"}
	in.Segmentation.Zones = append(in.Segmentation.Zones, SegZone{Name: "boxes", Intra: "allow", Members: []string{"containers:nas", "addr:lan"}})
	in.Segmentation.Policies = append([]SegPolicy{{Name: "nobody", From: []string{"group:empty"}, Action: "deny"}}, in.Segmentation.Policies...)
	ps := Check(in)
	for _, want := range []string{
		`"nas" has no container network, so this matches no address`,
		"192.168.0.0/24 is outside the mesh",
		`zone "boxes" has no addresses`,
		`policy "nobody" matches no traffic (its from resolves to nothing), so it is left out`,
	} {
		if !hasProblem(ps, "", want, true) {
			t.Errorf("no warning %q in %v", want, ps)
		}
	}
	for _, r := range CompilePolicy(in).Rules {
		if r.Name == "policy: nobody" || r.Name == "zone: boxes" {
			t.Errorf("%s matches nothing but was compiled: %+v", r.Name, r)
		}
	}
	if _, _, err := Compute(in); err != nil {
		t.Fatalf("warnings must not block the plan: %v", err)
	}
}

// A scheduled policy keeps its schedule, its days in the week's order, and
// the policy carries the time zone; outside the window the summary skips
// it.
func TestPolicySchedules(t *testing.T) {
	in := topology("full")
	in.Segmentation.TimeZone = "Asia/Singapore"
	in.Segmentation.Policies[1].Schedule = &proto.Schedule{Days: []string{"fri", "mon"}, Start: "09:00", End: "18:00"}
	if _, _, err := Compute(in); err != nil {
		t.Fatal(err)
	}
	pol := CompilePolicy(in)
	want := &proto.Schedule{Days: []string{"mon", "fri"}, Start: "09:00", End: "18:00"}
	if pol.TZ != "Asia/Singapore" || !reflect.DeepEqual(pol.Rules[4].Schedule, want) || pol.Rules[5].Schedule != nil {
		t.Fatalf("compiled %+v, time zone %q", pol.Rules[4], pol.TZ)
	}
	m, err := pol.Matcher()
	if err != nil {
		t.Fatal(err)
	}
	tuesday := m.Active(time.Date(2026, 10, 20, 10, 0, 0, 0, time.UTC))
	v, rules := Summarize(pol, tuesday, []string{testfix.Laptop.MeshIP}, ZoneAddresses(in)["home"])
	if v != ReachAllow || !reflect.DeepEqual(rules, []int{0, 6}) {
		t.Fatalf("the laptop to home on a Tuesday = %s %v", v, rules)
	}
	if bare := CompilePolicy(topology("full")); bare.TZ != "" {
		t.Fatalf("a policy without schedules has the time zone %q", bare.TZ)
	}
}

func TestSummarize(t *testing.T) {
	in := topology("full")
	pol := CompilePolicy(in)
	zones := ZoneAddresses(in)
	if !reflect.DeepEqual(zones["home"], []string{"10.77.1.1", "10.77.1.200"}) {
		t.Fatalf("ZoneAddresses = %v", zones)
	}
	for _, c := range []struct {
		name     string
		from, to []string
		verdict  string
		rules    []int
	}{
		{"within the home zone", zones["home"], zones["home"], ReachAllow, []int{0, 6}},
		{"the laptop to home", []string{testfix.Laptop.MeshIP}, zones["home"], ReachAllow, []int{0, 4, 6}},
		{"the phone to home", []string{testfix.Phone.MeshIP}, zones["home"], ReachPartial, []int{0, 5}},
		{"nothing", nil, zones["home"], ReachDeny, nil},
	} {
		v, rules := Summarize(pol, nil, c.from, c.to)
		if v != c.verdict || !reflect.DeepEqual(rules, c.rules) {
			t.Errorf("%s: Summarize = %s %v, want %s %v", c.name, v, rules, c.verdict, c.rules)
		}
	}
	// Without allow-all, home reaches a server only for ping, DNS and 9100.
	in.Segmentation.Policies = in.Segmentation.Policies[:4]
	pol = CompilePolicy(in)
	if v, _ := Summarize(pol, nil, zones["home"], []string{testfix.VPS1.MeshIP}); v != ReachPartial {
		t.Errorf("home to vps-1 = %s, want partial", v)
	}
	deny := &proto.Policy{Mesh: testfix.MeshCIDR, Rules: []proto.PolicyRule{{Name: "x", From: zones["home"], Action: "deny"}}}
	if v, rules := Summarize(deny, nil, zones["home"], []string{testfix.VPS1.MeshIP}); v != ReachDeny || !reflect.DeepEqual(rules, []int{0}) {
		t.Errorf("a full deny = %s %v", v, rules)
	}
}

func TestOverlaps(t *testing.T) {
	home := []string{"10.77.1.1", "10.77.1.200"}
	for _, c := range []struct {
		side []string
		want bool
	}{
		{nil, true}, // any
		{[]string{"10.77.1.0/24"}, true},
		{[]string{"10.77.1.2-10.77.1.199"}, false},
		{[]string{"10.77.2.1"}, false},
	} {
		if got := Overlaps(c.side, home); got != c.want {
			t.Errorf("Overlaps(%v) = %v", c.side, got)
		}
	}
	if Overlaps(nil, nil) {
		t.Error("any overlaps an empty zone")
	}
}
