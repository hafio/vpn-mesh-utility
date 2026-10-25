package exitmode_test

import (
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/hafio/vpn-mesh-utility/internal/agent/agentfake"
	"github.com/hafio/vpn-mesh-utility/internal/agent/exitmode"
	"github.com/hafio/vpn-mesh-utility/internal/proto"
)

func TestDesiredIsWgQuicksExitRouting(t *testing.T) {
	rules, routes := exitmode.Desired(0, "wg0", true, true, false)
	var got []string
	for _, r := range rules {
		got = append(got, r.String())
	}
	for _, r := range routes {
		got = append(got, r.String())
	}
	want := []string{
		"IPv4 rule 32764: lookup 254 suppress_prefixlength 0",
		"IPv4 rule 32765: lookup 51820 not fwmark 0xca6c",
		"IPv6 rule 32764: lookup 254 suppress_prefixlength 0",
		"IPv6 rule 32765: lookup 51820 not fwmark 0xca6c",
		"IPv4 default dev wg0 table 51820",
		"IPv6 blackhole default table 51820",
	}
	if !slices.Equal(got, want) {
		t.Fatalf("Desired =\n  %s\nwant\n  %s", strings.Join(got, "\n  "), strings.Join(want, "\n  "))
	}
	if r, rt := exitmode.Desired(0, "wg0", true, false, false); len(r) != 2 || len(rt) != 1 {
		t.Fatalf("without IPv6: %v %v", r, rt)
	}
	if r, rt := exitmode.Desired(0, "wg0", false, true, false); r != nil || rt != nil {
		t.Fatalf("exit off wants %v %v", r, rt)
	}
}

func TestSyncTurnsExitModeOnAndOff(t *testing.T) {
	p := &agentfake.Policy{}
	if n, err := exitmode.Sync(p, 0, "wg0", true, false); err != nil || n != 6 {
		t.Fatalf("on = %d, %v", n, err)
	}
	// Routes first: the table is ready before any traffic is sent to it.
	if !strings.HasPrefix(p.Calls[0], "add IPv4 default dev wg0") || !strings.HasPrefix(p.Calls[1], "add IPv6 blackhole") {
		t.Fatalf("calls %q", p.Calls)
	}
	if n, err := exitmode.Sync(p, 0, "wg0", true, false); err != nil || n != 0 {
		t.Fatalf("a second Sync changed %d (%v)", n, err)
	}
	p.Calls = nil
	if n, err := exitmode.Sync(p, 0, "wg0", false, false); err != nil || n != 6 {
		t.Fatalf("off = %d, %v", n, err)
	}
	// Rules first: nothing is sent to a table that is being emptied.
	if !strings.HasPrefix(p.Calls[0], "del IPv4 rule") || !strings.HasPrefix(p.Calls[5], "del IPv6 blackhole") {
		t.Fatalf("calls %q", p.Calls)
	}
	rules, _ := p.Rules()
	routes, _ := p.Routes(proto.RouteTable)
	if len(rules)+len(routes) != 0 {
		t.Fatalf("left behind %v %v", rules, routes)
	}
}

// A kernel with IPv6 but no IPv6 policy routing (amendment A45): exit off
// has nothing to do; exit on puts IPv4 in place and says that IPv6 is not
// held back.
func TestSyncWithoutIPv6PolicyRouting(t *testing.T) {
	p := &agentfake.Policy{V6Unrouted: true}
	if n, err := exitmode.Sync(p, 0, "wg0", false, false); n != 0 || err != nil {
		t.Fatalf("exit off = %d, %v", n, err)
	}
	n, err := exitmode.Sync(p, 0, "wg0", true, false)
	rules, _ := p.Rules()
	routes, _ := p.Routes(proto.RouteTable)
	wantRules, wantRoutes := exitmode.Desired(0, "wg0", true, false, false)
	if !errors.Is(err, exitmode.ErrIPv6Unguarded) || n != 3 || !slices.Equal(rules, wantRules) || !slices.Equal(routes, wantRoutes) {
		t.Fatalf("exit on = %d, %v; rules %v, routes %v", n, err, rules, routes)
	}
	if !strings.Contains(err.Error(), "disable_ipv6") {
		t.Fatalf("the error does not say what to do: %v", err)
	}
}

func TestSyncReplacesAStaleRuleOfItsOwn(t *testing.T) {
	p := &agentfake.Policy{NoIPv6: true}
	stale := exitmode.Rule{Family: exitmode.IPv4, Priority: 100, Table: 51820, SuppressPrefixlen: -1, NotMark: 51820}
	p.Install(stale)
	if n, err := exitmode.Sync(p, 0, "wg0", true, false); err != nil || n != 4 {
		t.Fatalf("Sync = %d, %v (one route, two rules, one removal)", n, err)
	}
	if rules, _ := p.Rules(); slices.Contains(rules, stale) || len(rules) != 2 {
		t.Fatalf("rules %v", rules)
	}
}

func TestSyncStopsAtTheFirstFailure(t *testing.T) {
	p := &agentfake.Policy{Fail: errors.New("operation not permitted")}
	if n, err := exitmode.Sync(p, 0, "wg0", true, false); err == nil || n != 0 || !strings.Contains(err.Error(), "not permitted") {
		t.Fatalf("Sync = %d, %v", n, err)
	}
	// Exit off on a host with nothing installed touches nothing: this is
	// every server's apply.
	if n, err := exitmode.Sync(p, 0, "wg0", false, false); err != nil || n != 0 {
		t.Fatalf("off with nothing installed = %d, %v", n, err)
	}
}

// Each failing call stops Sync where it is and is returned; the changes
// made before it are counted.
func TestSyncReportsEachFailure(t *testing.T) {
	boom := errors.New("operation not permitted")
	tunnel := exitmode.Route{Family: exitmode.IPv4, Table: proto.RouteTable, Link: "wg0"}
	suppress := exitmode.Rule{Family: exitmode.IPv4, Priority: exitmode.PrioritySuppress, Table: exitmode.MainTable}
	for _, c := range []struct {
		failOn string
		on     bool
		rule   *exitmode.Rule
		route  *exitmode.Route
		n      int
	}{
		{failOn: "Rules", on: true},
		{failOn: "Routes", on: true},
		{failOn: "AddRule", on: true, n: 1},
		{failOn: "DelRule", rule: &suppress},
		{failOn: "DelRoute", route: &tunnel},
	} {
		p := &agentfake.Policy{NoIPv6: true, Fail: boom, FailOn: c.failOn}
		if c.rule != nil {
			p.Install(*c.rule)
		}
		if c.route != nil {
			p.InstallRoute(*c.route)
		}
		if n, err := exitmode.Sync(p, 0, "wg0", c.on, false); !errors.Is(err, boom) || n != c.n {
			t.Errorf("%s failing: Sync = %d, %v; want %d changes and the error", c.failOn, n, err, c.n)
		}
	}
}

// An exit that carries IPv6 (M8) gets IPv6 through the tunnel instead of
// the blackhole; switching to one that does not replaces the route, and no
// moment passes without one.
func TestSyncCarriesIPv6ThroughTheExit(t *testing.T) {
	p := &agentfake.Policy{}
	if n, err := exitmode.Sync(p, 0, "wg0", true, true); err != nil || n != 6 {
		t.Fatalf("on with IPv6 = %d, %v", n, err)
	}
	six := exitmode.Route{Family: exitmode.IPv6, Table: proto.RouteTable, Link: "wg0"}
	if routes, _ := p.Routes(proto.RouteTable); !slices.Contains(routes, six) || len(routes) != 2 {
		t.Fatalf("routes %v", routes)
	}
	p.Calls = nil
	if n, err := exitmode.Sync(p, 0, "wg0", true, false); err != nil || n != 1 || !slices.Equal(p.Calls, []string{"add IPv6 blackhole default table 51820"}) {
		t.Fatalf("to an exit without IPv6 = %d, %v, calls %q", n, err, p.Calls)
	}
	if routes, _ := p.Routes(proto.RouteTable); slices.Contains(routes, six) || len(routes) != 2 {
		t.Fatalf("routes %v", routes)
	}
	if n, err := exitmode.Sync(p, 0, "wg0", true, true); err != nil || n != 1 {
		t.Fatalf("back = %d, %v", n, err)
	}
	if n, err := exitmode.Sync(p, 0, "wg0", false, true); err != nil || n != 6 {
		t.Fatalf("off = %d, %v", n, err)
	}
}

// SlotOf finds the slot of any slot's exit rule (amendment A49), and of no
// other priority.
func TestSlotOf(t *testing.T) {
	for s := proto.Slot(0); s <= proto.MaxSlot; s++ {
		suppress, tunnel := exitmode.Priorities(s)
		if a, ok := exitmode.SlotOf(suppress); !ok || a != s {
			t.Fatalf("SlotOf(%d) = %d, %v; want %d", suppress, a, ok, s)
		}
		if b, ok := exitmode.SlotOf(tunnel); !ok || b != s {
			t.Fatalf("SlotOf(%d) = %d, %v; want %d", tunnel, b, ok, s)
		}
	}
	for _, p := range []int{32766, 32763, 32762, 32133, 32125, 1000, 5200} {
		if s, ok := exitmode.SlotOf(p); ok {
			t.Errorf("SlotOf(%d) = %d", p, s)
		}
	}
}

// Two meshes' agents on one host (amendment A49): each has its own rules
// and table; the first to turn exit mode on keeps it, the other stays off
// and names the interface that holds it; of two that hold rules at once,
// the lower slot keeps them; exit off never touches another slot's rules.
func TestOneMeshCarriesTheExit(t *testing.T) {
	p := &agentfake.Policy{NoIPv6: true}
	if n, err := exitmode.Sync(p, 1, "wg1", true, false); err != nil || n != 3 {
		t.Fatalf("slot 1 on = %d, %v", n, err)
	}
	want := "add IPv4 default dev wg1 table 51822,add IPv4 rule 32754: lookup 254 suppress_prefixlength 0," +
		"add IPv4 rule 32755: lookup 51822 not fwmark 0xca6c"
	if got := strings.Join(p.Calls, ","); got != want {
		t.Fatalf("slot 1's calls %s", got)
	}
	p.Calls = nil
	n, err := exitmode.Sync(p, 0, "wg0", true, false)
	if !errors.Is(err, exitmode.ErrExitTaken) || !strings.Contains(err.Error(), "wg1") || n != 0 || len(p.Calls) != 0 {
		t.Fatalf("slot 0 on while slot 1 holds the exit = %d, %v, calls %q", n, err, p.Calls)
	}
	if n, err := exitmode.Sync(p, 0, "wg0", false, false); err != nil || n != 0 {
		t.Fatalf("slot 0 off = %d, %v", n, err)
	}
	if rules, _ := p.Rules(); len(rules) != 2 {
		t.Fatalf("slot 1's rules after slot 0's syncs: %v", rules)
	}

	// Both turned it on at once: slot 0 keeps its rules, slot 1 yields.
	rules, _ := exitmode.Desired(0, "wg0", true, false, false)
	for _, r := range rules {
		p.Install(r)
	}
	if n, err := exitmode.Sync(p, 0, "wg0", true, false); err != nil || n != 1 {
		t.Fatalf("slot 0 holding rules too = %d, %v", n, err)
	}
	if n, err := exitmode.Sync(p, 1, "wg1", true, false); !errors.Is(err, exitmode.ErrExitTaken) || !strings.Contains(err.Error(), "wg0") || n != 3 {
		t.Fatalf("slot 1 yielding = %d, %v", n, err)
	}
	if left, _ := p.Routes(proto.Slot(1).RouteTable()); len(left) != 0 {
		t.Fatalf("slot 1 left %v", left)
	}
	if got, _ := p.Rules(); !slices.Equal(got, rules) {
		t.Fatalf("rules %v, want slot 0's %v", got, rules)
	}
}

// Holder: the lowest other slot keeps the exit, unless own holds its own
// too and is lower; nobody else, nothing kept.
func TestHolder(t *testing.T) {
	for _, c := range []struct {
		own     proto.Slot
		holding bool
		others  []proto.Slot
		want    proto.Slot
		ok      bool
	}{
		{0, false, nil, 0, false},
		{0, false, []proto.Slot{0, 0}, 0, false},
		{0, false, []proto.Slot{3, 1, 1}, 1, true},
		{0, true, []proto.Slot{3, 1}, 0, false},
		{2, true, []proto.Slot{3, 1}, 1, true},
		{2, false, []proto.Slot{3}, 3, true},
	} {
		if got, ok := exitmode.Holder(c.own, c.holding, c.others); got != c.want || ok != c.ok {
			t.Errorf("Holder(%d, %v, %v) = %d, %v; want %d, %v", c.own, c.holding, c.others, got, ok, c.want, c.ok)
		}
	}
	if err := exitmode.Taken(4); !errors.Is(err, exitmode.ErrExitTaken) || !strings.Contains(err.Error(), "wg4") {
		t.Fatalf("Taken = %v", err)
	}
}
