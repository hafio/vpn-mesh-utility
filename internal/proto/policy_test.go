package proto_test

import (
	"net/netip"
	"reflect"
	"strings"
	"testing"
	"time"

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
		allow, rule := m.Decide(a(c.src), a(c.dst), c.proto, c.port, time.Time{})
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
		"schedule":         {func(p *proto.Policy) { p.TZ, p.Rules[3].Schedule = "UTC", &proto.Schedule{Days: []string{"monday"}} }, "policy.rules[3].schedule.days[0]"},
		"no time zone":     {func(p *proto.Policy) { p.Rules[3].Schedule = &proto.Schedule{Days: []string{"mon"}} }, "no time zone (tz)"},
		"time zone":        {func(p *proto.Policy) { p.TZ = "Local" }, "policy.tz"},
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

// officeHours is testfix.Policy with the laptop's SMB to the NAS allowed
// on weekdays from 09:00 to 18:00 in Singapore; outside it, the zone's
// rule decides.
func officeHours() *proto.Policy {
	p := testfix.Policy()
	p.TZ = "Asia/Singapore"
	p.Rules[3].Schedule = &proto.Schedule{Days: []string{"mon", "tue", "wed", "thu", "fri"}, Start: "09:00", End: "18:00"}
	return p
}

// sgt is a time in Singapore (UTC+8, no daylight saving); 2026-10-19 is a
// Monday.
func sgt(day, hour, minute int) time.Time {
	return time.Date(2026, 10, day, hour, minute, 0, 0, time.FixedZone("SGT", 8*3600))
}

// A scheduled rule decides only inside its window, read in the policy's
// zone whatever the zone of the time asked about; Next is the next start
// or end.
func TestPolicySchedules(t *testing.T) {
	m, err := officeHours().Matcher()
	if err != nil {
		t.Fatal(err)
	}
	laptop, nas := netip.MustParseAddr(testfix.Laptop.MeshIP), netip.MustParseAddr(testfix.NAS.MeshIP)
	for _, c := range []struct {
		name string
		at   time.Time
		rule int
		next time.Time
	}{
		{"monday morning", sgt(19, 10, 0), 3, sgt(19, 18, 0)},
		{"the same moment in UTC", sgt(19, 10, 0).UTC(), 3, sgt(19, 18, 0)},
		{"monday evening", sgt(19, 20, 0), 5, sgt(20, 9, 0)},
		{"the window opens", sgt(20, 9, 0), 3, sgt(20, 18, 0)},
		{"friday at closing", sgt(23, 18, 0), 5, sgt(24, 9, 0)},
		{"saturday", sgt(24, 10, 0), 5, sgt(24, 18, 0)},
	} {
		if _, rule := m.Decide(laptop, nas, "tcp", 445, c.at); rule != c.rule {
			t.Errorf("%s: rule %d, want %d", c.name, rule, c.rule)
		}
		want := []bool{true, true, true, c.rule == 3, true, true}
		if got := m.Active(c.at); !reflect.DeepEqual(got, want) {
			t.Errorf("%s: Active = %v, want %v", c.name, got, want)
		}
		if got := m.Next(c.at); !got.Equal(c.next) {
			t.Errorf("%s: Next = %v, want %v", c.name, got, c.next)
		}
	}
	plain, err := testfix.Policy().Matcher()
	if err != nil {
		t.Fatal(err)
	}
	if next := plain.Next(sgt(19, 10, 0)); !next.IsZero() {
		t.Errorf("Next without schedules = %v", next)
	}
}

// Across a daylight-saving change Next still moves forward, by at most a
// day: every edge is a time of day.
func TestPolicyNextAcrossDaylightSaving(t *testing.T) {
	p := testfix.Policy()
	p.TZ = "Europe/Berlin"
	p.Rules[3].Schedule = &proto.Schedule{Start: "02:30", End: "03:15"} // 2026-03-29 skips 02:00-03:00
	m, err := p.Matcher()
	if err != nil {
		t.Fatal(err)
	}
	berlin, _ := time.LoadLocation("Europe/Berlin")
	for at := time.Date(2026, 3, 28, 0, 0, 0, 0, berlin); at.Before(time.Date(2026, 3, 31, 0, 0, 0, 0, berlin)); at = at.Add(10 * time.Minute) {
		next := m.Next(at)
		if !next.After(at) || next.Sub(at) > 25*time.Hour {
			t.Fatalf("Next(%v) = %v", at, next)
		}
	}
}

func TestPolicyRuleString(t *testing.T) {
	p := officeHours()
	for i, want := range map[int]string{
		0: "allow any -> any icmp (system: ping)",
		3: "allow 10.77.2.1 -> 10.77.1.1 tcp/445 mon-fri 09:00-18:00 (policy: laptop-smb)",
		4: "deny 10.77.2.1 -> 10.77.0.1-10.77.0.3 any (policy: laptop-servers)",
	} {
		if got := p.Rules[i].String(); got != want {
			t.Errorf("rule %d = %q, want %q", i, got, want)
		}
	}
}
