// Package routes keeps the routes on the mesh interface equal to the peers'
// AllowedIPs, as wg-quick does (spec S7.2), minus 0.0.0.0/0: exit mode (M3)
// routes the default by policy, never in the main table.
//
// Routes the agent installs carry protocol 201 (amendment A18), so Sync only
// ever lists, and therefore only ever removes, the agent's own routes.
package routes

import (
	"fmt"
	"net/netip"
	"slices"

	"github.com/hafio/vpn-mesh-utility/internal/proto"
)

// Protocol tags the agent's routes ("proto 201" in `ip route`).
const Protocol = 201

// Table is the routing table as the agent sees it: its own routes on one
// interface.
type Table interface {
	// List returns the agent's routes (protocol 201) on link in the main table.
	List(link string) ([]netip.Prefix, error)
	Add(link string, p netip.Prefix) error
	Del(link string, p netip.Prefix) error
}

var defaultRoute = netip.MustParsePrefix("0.0.0.0/0")

// Desired is every peer's AllowedIPs except the default route, sorted and
// without duplicates.
func Desired(cfg proto.NodeConfig) ([]netip.Prefix, error) {
	var out []netip.Prefix
	for _, p := range cfg.Peers {
		for _, s := range p.AllowedIPs {
			pf, err := netip.ParsePrefix(s)
			if err != nil {
				return nil, fmt.Errorf("peer %q: allowed ip %q: %w", p.Name, s, err)
			}
			if pf != defaultRoute {
				out = append(out, pf)
			}
		}
	}
	slices.SortFunc(out, compare)
	return slices.Compact(out), nil
}

func compare(a, b netip.Prefix) int {
	if c := a.Addr().Compare(b.Addr()); c != 0 {
		return c
	}
	return a.Bits() - b.Bits()
}

// Sync makes the agent's routes on link exactly desired, adding before it
// removes so traffic keeps a route throughout. It returns how many changes
// it made.
func Sync(t Table, link string, desired []netip.Prefix) (int, error) {
	have, err := t.List(link)
	if err != nil {
		return 0, err
	}
	n := 0
	for _, p := range desired {
		if !slices.Contains(have, p) {
			if err := t.Add(link, p); err != nil {
				return n, err
			}
			n++
		}
	}
	for _, p := range have {
		if !slices.Contains(desired, p) {
			if err := t.Del(link, p); err != nil {
				return n, err
			}
			n++
		}
	}
	return n, nil
}
