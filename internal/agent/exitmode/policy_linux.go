//go:build linux

package exitmode

import (
	"errors"
	"fmt"
	"net"
	"os"
	"syscall"

	"github.com/vishvananda/netlink"

	"github.com/hafio/vpn-mesh-utility/internal/agent/routes"
	"github.com/hafio/vpn-mesh-utility/internal/proto"
)

type netlinkPolicy struct{}

// NewPolicy returns the kernel's policy routing.
func NewPolicy() Policy { return netlinkPolicy{} }

func family(f Family) int {
	if f == IPv6 {
		return netlink.FAMILY_V6
	}
	return netlink.FAMILY_V4
}

// IPv6 is V6None without the IPv6 stack, V6Unrouted when the kernel
// refuses IPv6 rule dumps (EAFNOSUPPORT: no IPV6_MULTIPLE_TABLES), else
// V6Policy.
func (netlinkPolicy) IPv6() V6Support {
	if _, err := os.Stat("/proc/net/if_inet6"); err != nil {
		return V6None
	}
	if _, err := netlink.RuleList(netlink.FAMILY_V6); errors.Is(err, syscall.EAFNOSUPPORT) {
		return V6Unrouted
	}
	return V6Policy
}

// dump retries a netlink dump that a concurrent change interrupted.
func dump[T any](list func() ([]T, error)) ([]T, error) {
	var (
		out []T
		err error
	)
	for try := 0; try < 3; try++ {
		if out, err = list(); !errors.Is(err, netlink.ErrDumpInterrupted) {
			break
		}
	}
	return out, err
}

func (p netlinkPolicy) families() []Family {
	if p.IPv6() == V6Policy {
		return []Family{IPv4, IPv6}
	}
	return []Family{IPv4}
}

func (p netlinkPolicy) Rules() ([]Rule, error) {
	var out []Rule
	for _, f := range p.families() {
		rs, err := dump(func() ([]netlink.Rule, error) { return netlink.RuleList(family(f)) })
		if err != nil {
			return nil, fmt.Errorf("list %s policy rules: %w", f, err)
		}
		for _, r := range rs {
			if !owned(r) {
				continue
			}
			x := Rule{Family: f, Priority: r.Priority, Table: r.Table, SuppressPrefixlen: r.SuppressPrefixlen}
			if r.Invert {
				x.NotMark = r.Mark
			}
			out = append(out, x)
		}
	}
	return out, nil
}

// owned reports whether r is exit mode's: protocol 201 at one of its two
// priorities. The container network's rules carry protocol 201 too
// (internal/agent/containers), at priorities of their own.
func owned(r netlink.Rule) bool {
	return r.Protocol == routes.Protocol && (r.Priority == PrioritySuppress || r.Priority == PriorityTunnel)
}

func toRule(r Rule) *netlink.Rule {
	nr := netlink.NewRule()
	nr.Family = family(r.Family)
	nr.Priority = r.Priority
	nr.Table = r.Table
	nr.SuppressPrefixlen = r.SuppressPrefixlen
	nr.Protocol = routes.Protocol
	if r.NotMark != 0 {
		nr.Mark, nr.Invert = r.NotMark, true
	}
	return nr
}

func (netlinkPolicy) AddRule(r Rule) error {
	if err := netlink.RuleAdd(toRule(r)); err != nil {
		return fmt.Errorf("add %s: %w", r, err)
	}
	return nil
}

func (netlinkPolicy) DelRule(r Rule) error {
	if err := netlink.RuleDel(toRule(r)); err != nil {
		return fmt.Errorf("remove %s: %w", r, err)
	}
	return nil
}

func (p netlinkPolicy) Routes() ([]Route, error) {
	var out []Route
	for _, f := range p.families() {
		filter := &netlink.Route{Table: proto.RouteTable, Protocol: routes.Protocol}
		rs, err := dump(func() ([]netlink.Route, error) {
			return netlink.RouteListFiltered(family(f), filter, netlink.RT_FILTER_TABLE|netlink.RT_FILTER_PROTOCOL)
		})
		if err != nil {
			return nil, fmt.Errorf("list %s routes in table %d: %w", f, proto.RouteTable, err)
		}
		for _, r := range rs {
			if r.Dst != nil {
				if ones, _ := r.Dst.Mask.Size(); ones != 0 {
					continue // only default routes are the agent's here
				}
			}
			x := Route{Family: f}
			if r.Type != syscall.RTN_BLACKHOLE {
				l, err := netlink.LinkByIndex(r.LinkIndex)
				if err != nil {
					return nil, fmt.Errorf("look up the interface of a route in table %d: %w", proto.RouteTable, err)
				}
				x.Link = l.Attrs().Name
			}
			out = append(out, x)
		}
	}
	return out, nil
}

func toRoute(r Route) (*netlink.Route, error) {
	nr := &netlink.Route{Table: proto.RouteTable, Protocol: routes.Protocol, Family: family(r.Family)}
	if r.Family == IPv6 {
		nr.Dst = &net.IPNet{IP: net.IPv6zero, Mask: net.CIDRMask(0, 128)}
	} else {
		nr.Dst = &net.IPNet{IP: net.IPv4zero.To4(), Mask: net.CIDRMask(0, 32)}
	}
	if r.Link == "" {
		nr.Type = syscall.RTN_BLACKHOLE
		return nr, nil
	}
	l, err := netlink.LinkByName(r.Link)
	if err != nil {
		return nil, fmt.Errorf("look up interface %s: %w", r.Link, err)
	}
	nr.LinkIndex, nr.Scope = l.Attrs().Index, netlink.SCOPE_LINK
	return nr, nil
}

func (netlinkPolicy) AddRoute(r Route) error {
	nr, err := toRoute(r)
	if err != nil {
		return err
	}
	if err := netlink.RouteReplace(nr); err != nil {
		return fmt.Errorf("add %s: %w", r, err)
	}
	return nil
}

func (netlinkPolicy) DelRoute(r Route) error {
	nr, err := toRoute(r)
	if err != nil {
		return err
	}
	if err := netlink.RouteDel(nr); err != nil {
		return fmt.Errorf("remove %s: %w", r, err)
	}
	return nil
}
