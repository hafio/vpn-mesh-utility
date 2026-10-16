//go:build linux

package routes

import (
	"fmt"
	"net/netip"

	"github.com/vishvananda/netlink"

	"github.com/hafio/vpn-mesh-utility/internal/agent/netconv"
)

type netlinkTable struct{}

// NewTable returns the kernel's routing table.
func NewTable() Table { return netlinkTable{} }

func index(link string) (int, error) {
	l, err := netlink.LinkByName(link)
	if err != nil {
		return 0, fmt.Errorf("look up interface %s: %w", link, err)
	}
	return l.Attrs().Index, nil
}

func (netlinkTable) List(link string) ([]netip.Prefix, error) {
	idx, err := index(link)
	if err != nil {
		return nil, err
	}
	filter := &netlink.Route{LinkIndex: idx, Protocol: Protocol}
	var rs []netlink.Route
	for try := 0; try < 3; try++ {
		// A dump interrupted by a concurrent change is retried.
		if rs, err = netlink.RouteListFiltered(netlink.FAMILY_V4, filter, netlink.RT_FILTER_OIF|netlink.RT_FILTER_PROTOCOL); err == nil {
			break
		}
	}
	if err != nil {
		return nil, fmt.Errorf("list routes on %s: %w", link, err)
	}
	var out []netip.Prefix
	for _, r := range rs {
		if r.Dst == nil {
			continue
		}
		if p, ok := netconv.Prefix(*r.Dst); ok {
			out = append(out, p)
		}
	}
	return out, nil
}

func route(idx int, p netip.Prefix) *netlink.Route {
	dst := netconv.IPNet(p)
	return &netlink.Route{LinkIndex: idx, Dst: &dst, Scope: netlink.SCOPE_LINK, Protocol: Protocol}
}

func (netlinkTable) Add(link string, p netip.Prefix) error {
	idx, err := index(link)
	if err != nil {
		return err
	}
	if err := netlink.RouteReplace(route(idx, p)); err != nil {
		return fmt.Errorf("add route %s dev %s: %w", p, link, err)
	}
	return nil
}

func (netlinkTable) Del(link string, p netip.Prefix) error {
	idx, err := index(link)
	if err != nil {
		return err
	}
	if err := netlink.RouteDel(route(idx, p)); err != nil {
		return fmt.Errorf("remove route %s dev %s: %w", p, link, err)
	}
	return nil
}
