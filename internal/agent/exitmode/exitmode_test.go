package exitmode_test

import (
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/hafio/vpn-mesh-utility/internal/agent/agentfake"
	"github.com/hafio/vpn-mesh-utility/internal/agent/exitmode"
)

func TestDesiredIsWgQuicksExitRouting(t *testing.T) {
	rules, routes := exitmode.Desired("wg0", true, true, false)
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
	if r, rt := exitmode.Desired("wg0", true, false, false); len(r) != 2 || len(rt) != 1 {
		t.Fatalf("without IPv6: %v %v", r, rt)
	}
	if r, rt := exitmode.Desired("wg0", false, true, false); r != nil || rt != nil {
		t.Fatalf("exit off wants %v %v", r, rt)
	}
}

func TestSyncTurnsExitModeOnAndOff(t *testing.T) {
	p := &agentfake.Policy{}
	if n, err := exitmode.Sync(p, "wg0", true, false); err != nil || n != 6 {
		t.Fatalf("on = %d, %v", n, err)
	}
	// Routes first: the table is ready before any traffic is sent to it.
	if !strings.HasPrefix(p.Calls[0], "add IPv4 default dev wg0") || !strings.HasPrefix(p.Calls[1], "add IPv6 blackhole") {
		t.Fatalf("calls %q", p.Calls)
	}
	if n, err := exitmode.Sync(p, "wg0", true, false); err != nil || n != 0 {
		t.Fatalf("a second Sync changed %d (%v)", n, err)
	}
	p.Calls = nil
	if n, err := exitmode.Sync(p, "wg0", false, false); err != nil || n != 6 {
		t.Fatalf("off = %d, %v", n, err)
	}
	// Rules first: nothing is sent to a table that is being emptied.
	if !strings.HasPrefix(p.Calls[0], "del IPv4 rule") || !strings.HasPrefix(p.Calls[5], "del IPv6 blackhole") {
		t.Fatalf("calls %q", p.Calls)
	}
	rules, _ := p.Rules()
	routes, _ := p.Routes()
	if len(rules)+len(routes) != 0 {
		t.Fatalf("left behind %v %v", rules, routes)
	}
}

// A kernel with IPv6 but no IPv6 policy routing (amendment A45): exit off
// has nothing to do; exit on puts IPv4 in place and says that IPv6 is not
// held back.
func TestSyncWithoutIPv6PolicyRouting(t *testing.T) {
	p := &agentfake.Policy{V6Unrouted: true}
	if n, err := exitmode.Sync(p, "wg0", false, false); n != 0 || err != nil {
		t.Fatalf("exit off = %d, %v", n, err)
	}
	n, err := exitmode.Sync(p, "wg0", true, false)
	rules, _ := p.Rules()
	routes, _ := p.Routes()
	wantRules, wantRoutes := exitmode.Desired("wg0", true, false, false)
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
	if n, err := exitmode.Sync(p, "wg0", true, false); err != nil || n != 4 {
		t.Fatalf("Sync = %d, %v (one route, two rules, one removal)", n, err)
	}
	if rules, _ := p.Rules(); slices.Contains(rules, stale) || len(rules) != 2 {
		t.Fatalf("rules %v", rules)
	}
}

func TestSyncStopsAtTheFirstFailure(t *testing.T) {
	p := &agentfake.Policy{Fail: errors.New("operation not permitted")}
	if n, err := exitmode.Sync(p, "wg0", true, false); err == nil || n != 0 || !strings.Contains(err.Error(), "not permitted") {
		t.Fatalf("Sync = %d, %v", n, err)
	}
	// Exit off on a host with nothing installed touches nothing: this is
	// every server's apply.
	if n, err := exitmode.Sync(p, "wg0", false, false); err != nil || n != 0 {
		t.Fatalf("off with nothing installed = %d, %v", n, err)
	}
}

// Each failing call stops Sync where it is and is returned; the changes
// made before it are counted.
func TestSyncReportsEachFailure(t *testing.T) {
	boom := errors.New("operation not permitted")
	tunnel := exitmode.Route{Family: exitmode.IPv4, Link: "wg0"}
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
		if n, err := exitmode.Sync(p, "wg0", c.on, false); !errors.Is(err, boom) || n != c.n {
			t.Errorf("%s failing: Sync = %d, %v; want %d changes and the error", c.failOn, n, err, c.n)
		}
	}
}

// An exit that carries IPv6 (M8) gets IPv6 through the tunnel instead of
// the blackhole; switching to one that does not replaces the route, and no
// moment passes without one.
func TestSyncCarriesIPv6ThroughTheExit(t *testing.T) {
	p := &agentfake.Policy{}
	if n, err := exitmode.Sync(p, "wg0", true, true); err != nil || n != 6 {
		t.Fatalf("on with IPv6 = %d, %v", n, err)
	}
	six := exitmode.Route{Family: exitmode.IPv6, Link: "wg0"}
	if routes, _ := p.Routes(); !slices.Contains(routes, six) || len(routes) != 2 {
		t.Fatalf("routes %v", routes)
	}
	p.Calls = nil
	if n, err := exitmode.Sync(p, "wg0", true, false); err != nil || n != 1 || !slices.Equal(p.Calls, []string{"add IPv6 blackhole default table 51820"}) {
		t.Fatalf("to an exit without IPv6 = %d, %v, calls %q", n, err, p.Calls)
	}
	if routes, _ := p.Routes(); slices.Contains(routes, six) || len(routes) != 2 {
		t.Fatalf("routes %v", routes)
	}
	if n, err := exitmode.Sync(p, "wg0", true, true); err != nil || n != 1 {
		t.Fatalf("back = %d, %v", n, err)
	}
	if n, err := exitmode.Sync(p, "wg0", false, true); err != nil || n != 6 {
		t.Fatalf("off = %d, %v", n, err)
	}
}
