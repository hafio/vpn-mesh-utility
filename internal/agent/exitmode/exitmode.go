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
// exit mode never leaks IPv6 around the tunnel (amendment A11).
//
// Switching exits only moves 0.0.0.0/0 between peers; the rules stay, so
// nothing reconnects. Every rule and route here carries protocol 201, like
// the agent's mesh routes, so Sync lists, and therefore removes, only its own
// (amendment A27).
package exitmode

import (
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

// Rule priorities: what wg-quick picks on a host with only the default
// rules (main is 32766), so the suppress rule is looked at first.
const (
	PrioritySuppress = 32764
	PriorityTunnel   = 32765
)

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

// Route is a default route in proto.RouteTable.
type Route struct {
	Family Family
	// Link is the interface of a unicast route; "" is a blackhole.
	Link string
}

func (r Route) String() string {
	if r.Link == "" {
		return fmt.Sprintf("%s blackhole default table %d", r.Family, proto.RouteTable)
	}
	return fmt.Sprintf("%s default dev %s table %d", r.Family, r.Link, proto.RouteTable)
}

// Policy is the host's policy routing as the agent sees it: its own rules
// and its own routes in proto.RouteTable.
type Policy interface {
	Rules() ([]Rule, error)
	AddRule(r Rule) error
	DelRule(r Rule) error
	Routes() ([]Route, error)
	AddRoute(r Route) error
	DelRoute(r Route) error
	// IPv6 reports whether the host has IPv6 at all; without it nothing
	// can leak and there is nothing to add.
	IPv6() bool
}

// Desired is what exit mode through link needs; nothing when on is false.
func Desired(link string, on, ipv6 bool) ([]Rule, []Route) {
	if !on {
		return nil, nil
	}
	families := []Family{IPv4}
	if ipv6 {
		families = append(families, IPv6)
	}
	var rules []Rule
	for _, f := range families {
		rules = append(rules,
			Rule{Family: f, Priority: PrioritySuppress, Table: MainTable, SuppressPrefixlen: 0},
			Rule{Family: f, Priority: PriorityTunnel, Table: proto.RouteTable, SuppressPrefixlen: -1, NotMark: proto.FwMark})
	}
	routes := []Route{{Family: IPv4, Link: link}}
	if ipv6 {
		routes = append(routes, Route{Family: IPv6})
	}
	return rules, routes
}

// Sync makes the agent's rules and table routes exactly what exit mode needs.
// Routes are added before rules, so the table is ready when traffic is sent
// to it, and rules are removed before routes. It returns how many changes it
// made.
func Sync(p Policy, link string, on bool) (int, error) {
	wantRules, wantRoutes := Desired(link, on, p.IPv6())
	haveRules, err := p.Rules()
	if err != nil {
		return 0, err
	}
	haveRoutes, err := p.Routes()
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
		if !slices.Contains(wantRoutes, r) {
			if err := p.DelRoute(r); err != nil {
				return n, err
			}
			n++
		}
	}
	return n, nil
}
