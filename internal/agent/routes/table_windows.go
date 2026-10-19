//go:build windows

package routes

import (
	"fmt"
	"net"
	"net/netip"
	"slices"

	"golang.org/x/sys/windows"
	"golang.zx2c4.com/wireguard/windows/tunnel/winipcfg"
)

// winTable is the routing table on Windows: the static routes on the mesh
// adapter. Every one is the agent's -- so is the adapter -- so where Linux
// tags its routes with Protocol, Windows needs only to leave alone the
// routes it makes itself for the adapter's address (protocol "local"), and
// the split default routes, which are exit mode's (ExitPrefixes).
type winTable struct{}

// NewTable returns the host's routing table.
func NewTable() Table { return winTable{} }

func luidOf(link string) (winipcfg.LUID, error) {
	ifc, err := net.InterfaceByName(link)
	if err != nil {
		return 0, fmt.Errorf("look up adapter %s: %w", link, err)
	}
	luid, err := winipcfg.LUIDFromIndex(uint32(ifc.Index))
	if err != nil {
		return 0, fmt.Errorf("look up adapter %s: %w", link, err)
	}
	return luid, nil
}

func (winTable) List(link string) ([]netip.Prefix, error) {
	luid, err := luidOf(link)
	if err != nil {
		return nil, err
	}
	rows, err := winipcfg.GetIPForwardTable2(windows.AF_INET)
	if err != nil {
		return nil, fmt.Errorf("list routes: %w", err)
	}
	var out []netip.Prefix
	for i := range rows {
		r := &rows[i]
		if p := r.DestinationPrefix.Prefix(); r.InterfaceLUID == luid && r.Protocol == winipcfg.RouteProtocolNetMgmt && !slices.Contains(ExitPrefixes, p) {
			out = append(out, p)
		}
	}
	return out, nil
}

// Add puts p on the adapter as an on-link route; the adapter's own metric
// orders it against the host's other routes.
func (winTable) Add(link string, p netip.Prefix) error {
	luid, err := luidOf(link)
	if err == nil {
		err = luid.AddRoute(p, netip.IPv4Unspecified(), 0)
	}
	if err != nil {
		return fmt.Errorf("add route %s via %s: %w", p, link, err)
	}
	return nil
}

func (winTable) Del(link string, p netip.Prefix) error {
	luid, err := luidOf(link)
	if err == nil {
		err = luid.DeleteRoute(p, netip.IPv4Unspecified())
	}
	if err != nil {
		return fmt.Errorf("remove route %s via %s: %w", p, link, err)
	}
	return nil
}
