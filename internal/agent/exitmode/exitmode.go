// Package exitmode sends a spoke's internet traffic through its exit server
// (spec S7.3) the way wg-quick does. The planner puts 0.0.0.0/0 on the exit
// server's peer; this package adds the policy routing that uses it:
//
//	32764: from all lookup main suppress_prefixlength 0
//	32765: not from all fwmark 0xca6c lookup 51820
//	table 51820: default dev wg0
//
// The first rule keeps every specific route of the main table (the LAN, the
// mesh /32s, the way to the server endpoints) and ignores only its default
// route; the second sends what is left to table 51820, unless it carries the
// interface's fwmark -- WireGuard's own UDP, which must take the normal path.
// IPv6 gets the same two rules and a blackhole default in table 51820, so
// exit mode never leaks IPv6 around the tunnel (amendment A11) -- or, when
// the exit carries IPv6 (M8: ::/0 on its peer), a default dev wg0 there
// too, so IPv6 goes through the exit as IPv4 does. A kernel
// with IPv6 but without IPv6 policy routing (some NAS kernels) has no rules
// to manage: exit mode then carries IPv4 only and says so (amendment A45).
//
// Switching exits only moves 0.0.0.0/0 between peers; the rules stay, so
// nothing reconnects. Every rule and route here carries protocol 201, like
// the agent's mesh routes, so Sync lists, and therefore removes, only its own
// (amendment A27).
//
// The agent of a mesh in slot s (amendment A49) uses priorities 10*s lower
// and table 51820+2s, on its own interface; the fwmark is every slot's, so
// the other meshes' tunnels never go through an exit. One mesh at a time
// carries a host's internet: Sync turns exit mode on only while no other
// slot's rules are there, and says so (ErrExitTaken).
package exitmode

import (
	"errors"
	"fmt"
	"slices"

	"github.com/hafio/vpn-mesh-utility/internal/proto"
)

// Family is an address family.
type Family int

const (
	IPv4 Family = 4
	IPv6 Family = 6
)

func (f Family) String() string { return fmt.Sprintf("IPv%d", int(f)) }

// Rule priorities of slot 0: what wg-quick picks on a host with only the
// default rules (main is 32766), so the suppress rule is looked at first.
const (
	PrioritySuppress = 32764
	PriorityTunnel   = 32765
)

// Priorities are slot s's two rule priorities, RuleShift below slot 0's.
func Priorities(s proto.Slot) (suppress, tunnel int) {
	return PrioritySuppress - s.RuleShift(), PriorityTunnel - s.RuleShift()
}

// SlotOf is the slot whose exit-mode rule sits at priority, if any slot's
// does.
func SlotOf(priority int) (proto.Slot, bool) {
	d := PriorityTunnel - priority // 10s for a tunnel rule, 10s+1 for a suppress rule
	s := proto.Slot(d / 10)
	if d < 0 || d%10 > 1 || !s.Valid() {
		return 0, false
	}
	return s, true
}

// MainTable is the kernel's main routing table.
const MainTable = 254

// Rule is one policy rule.
type Rule struct {
	Family   Family
	Priority int
	// Table is the table the rule looks up.
	Table int
	// SuppressPrefixlen is 0 on the suppress rule and -1 on every other.
	SuppressPrefixlen int
	// NotMark matches packets that do not carry this fwmark; 0 for none.
	NotMark uint32
}

func (r Rule) String() string {
	s := fmt.Sprintf("%s rule %d: lookup %d", r.Family, r.Priority, r.Table)
	if r.SuppressPrefixlen >= 0 {
		s += fmt.Sprintf(" suppress_prefixlength %d", r.SuppressPrefixlen)
	}
	if r.NotMark != 0 {
		s += fmt.Sprintf(" not fwmark %#x", r.NotMark)
	}
	return s
}

// Route is a default route in a slot's table (proto.Slot.RouteTable).
type Route struct {
	Family Family
	Table  int
	// Link is the interface of a unicast route; "" is a blackhole.
	Link string
}

func (r Route) String() string {
	if r.Link == "" {
		return fmt.Sprintf("%s blackhole default table %d", r.Family, r.Table)
	}
	return fmt.Sprintf("%s default dev %s table %d", r.Family, r.Link, r.Table)
}

// Policy is the host's policy routing as the agents see it: every slot's
// exit-mode rules (protocol 201 at a priority SlotOf knows) and the
// protocol 201 default routes of a table.
type Policy interface {
	Rules() ([]Rule, error)
	AddRule(r Rule) error
	DelRule(r Rule) error
	Routes(table int) ([]Route, error)
	// AddRoute replaces the table's default route of r's family: there is
	// one per family.
	AddRoute(r Route) error
	DelRoute(r Route) error
	// IPv6 is what the host has of IPv6.
	IPv6() V6Support
}

// V6Support is what exit mode can do about a host's IPv6.
type V6Support int

const (
	// V6None is a host without IPv6: nothing can leak, nothing to add.
	V6None V6Support = iota
	// V6Unrouted is IPv6 without IPv6 policy routing (a kernel built
	// without IPV6_MULTIPLE_TABLES): there are no rules to list or add,
	// and exit mode can neither send IPv6 through the tunnel nor hold it
	// back.
	V6Unrouted
	// V6Policy is IPv6 with policy routing: exit mode blackholes it.
	V6Policy
)

// ErrIPv6Unguarded is exit mode on a V6Unrouted host: IPv4 goes through the
// exit, IPv6 is neither sent through it nor held back. The apply reports it
// as degraded; the rest of exit mode is in place.
var ErrIPv6Unguarded = errors.New("exit mode carries IPv4 only: this kernel has IPv6 but no IPv6 policy routing, so IPv6 traffic " +
	"bypasses the exit; turn IPv6 off on this host (sysctl net.ipv6.conf.all.disable_ipv6=1) to keep it from leaking")

// ErrExitTaken is exit mode asked for while the agent of another mesh on
// the same host already sends the host's internet through its exit: this
// mesh's exit stays off, and the apply reports it as degraded.
var ErrExitTaken = errors.New("exit mode is off: the agent of another mesh on this host already sends its internet through that mesh's exit")

// Desired is what slot s's exit mode through link needs; nothing when on is
// false. ipv6 is a host with IPv6 policy routing, via6 an exit that
// carries IPv6.
func Desired(s proto.Slot, link string, on, ipv6, via6 bool) ([]Rule, []Route) {
	if !on {
		return nil, nil
	}
	families := []Family{IPv4}
	if ipv6 {
		families = append(families, IPv6)
	}
	suppress, tunnel := Priorities(s)
	var rules []Rule
	for _, f := range families {
		rules = append(rules,
			Rule{Family: f, Priority: suppress, Table: MainTable, SuppressPrefixlen: 0},
			Rule{Family: f, Priority: tunnel, Table: s.RouteTable(), SuppressPrefixlen: -1, NotMark: proto.FwMark})
	}
	routes := []Route{{Family: IPv4, Table: s.RouteTable(), Link: link}}
	switch {
	case ipv6 && via6:
		routes = append(routes, Route{Family: IPv6, Table: s.RouteTable(), Link: link})
	case ipv6:
		routes = append(routes, Route{Family: IPv6, Table: s.RouteTable()})
	}
	return rules, routes
}

// Sync makes slot s's rules and table routes exactly what exit mode needs.
// Routes are added before rules, so the table is ready when traffic is sent
// to it, and rules are removed before routes. It returns how many changes it
// made, and ErrIPv6Unguarded once IPv4 is in place on a host whose IPv6 it
// cannot hold back. via6 is an exit that carries IPv6 (M8).
//
// Another slot's rules are never touched. While they are there, exit mode
// stays off (and the error wraps ErrExitTaken, naming that slot's
// interface): the first mesh to turn it on keeps it, and of two that did at
// once the lower slot does.
func Sync(p Policy, s proto.Slot, link string, on, via6 bool) (int, error) {
	v6 := p.IPv6()
	all, err := p.Rules()
	if err != nil {
		return 0, err
	}
	var haveRules []Rule
	holder := proto.Slot(-1) // the lowest other slot with exit rules
	for _, r := range all {
		switch rs, _ := SlotOf(r.Priority); {
		case rs == s:
			haveRules = append(haveRules, r)
		case holder < 0 || rs < holder:
			holder = rs
		}
	}
	var taken error
	if on && holder >= 0 && (len(haveRules) == 0 || holder < s) {
		on = false
		taken = fmt.Errorf("%w (its interface %s); turn that exit off first (`meshctl exit off` for that mesh)", ErrExitTaken, holder.Interface())
	}
	wantRules, wantRoutes := Desired(s, link, on, v6 == V6Policy, via6)
	haveRoutes, err := p.Routes(s.RouteTable())
	if err != nil {
		return 0, err
	}
	n := 0
	for _, r := range wantRoutes {
		if !slices.Contains(haveRoutes, r) {
			if err := p.AddRoute(r); err != nil {
				return n, err
			}
			n++
		}
	}
	for _, r := range wantRules {
		if !slices.Contains(haveRules, r) {
			if err := p.AddRule(r); err != nil {
				return n, err
			}
			n++
		}
	}
	for _, r := range haveRules {
		if !slices.Contains(wantRules, r) {
			if err := p.DelRule(r); err != nil {
				return n, err
			}
			n++
		}
	}
	for _, r := range haveRoutes {
		replaced := slices.ContainsFunc(wantRoutes, func(w Route) bool { return w.Family == r.Family })
		if !slices.Contains(wantRoutes, r) && !replaced {
			if err := p.DelRoute(r); err != nil {
				return n, err
			}
			n++
		}
	}
	if on && v6 == V6Unrouted {
		return n, ErrIPv6Unguarded
	}
	return n, taken
}
