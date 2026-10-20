//go:build linux

package containers

import (
	"errors"
	"fmt"
	"net"
	"net/netip"
	"syscall"

	"github.com/vishvananda/netlink"

	"github.com/hafio/vpn-mesh-utility/internal/agent/netconv"
	"github.com/hafio/vpn-mesh-utility/internal/agent/routes"
)

type netlinkNet struct{}

// NewNet returns the kernel's view of the host.
func NewNet() Net { return netlinkNet{} }

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

func prefixOf(n *net.IPNet) netip.Prefix {
	if n == nil {
		return netip.Prefix{}
	}
	p, _ := netconv.Prefix(*n)
	return p
}

func (netlinkNet) Rules() ([]Rule, error) {
	rs, err := dump(func() ([]netlink.Rule, error) { return netlink.RuleList(netlink.FAMILY_V4) })
	if err != nil {
		return nil, fmt.Errorf("list policy rules: %w", err)
	}
	var out []Rule
	for _, r := range rs {
		if r.Protocol != routes.Protocol || !Owned(r.Priority) {
			continue
		}
		out = append(out, Rule{Priority: r.Priority, Iif: r.IifName, Src: prefixOf(r.Src), Dst: prefixOf(r.Dst), Table: r.Table})
	}
	return out, nil
}

func toRule(r Rule) *netlink.Rule {
	nr := netlink.NewRule()
	nr.Family, nr.Priority, nr.Table, nr.IifName, nr.Protocol = netlink.FAMILY_V4, r.Priority, r.Table, r.Iif, routes.Protocol
	if r.Src.IsValid() {
		n := netconv.IPNet(r.Src)
		nr.Src = &n
	}
	if r.Dst.IsValid() {
		n := netconv.IPNet(r.Dst)
		nr.Dst = &n
	}
	return nr
}

func (netlinkNet) AddRule(r Rule) error {
	if err := netlink.RuleAdd(toRule(r)); err != nil {
		return fmt.Errorf("add %s: %w", r, err)
	}
	return nil
}

func (netlinkNet) DelRule(r Rule) error {
	if err := netlink.RuleDel(toRule(r)); err != nil {
		return fmt.Errorf("remove %s: %w", r, err)
	}
	return nil
}

func blackhole() *netlink.Route {
	return &netlink.Route{Table: BlackholeTable, Protocol: routes.Protocol, Family: netlink.FAMILY_V4, Type: syscall.RTN_BLACKHOLE,
		Dst: &net.IPNet{IP: net.IPv4zero.To4(), Mask: net.CIDRMask(0, 32)}}
}

func (netlinkNet) Blackhole() (bool, error) {
	filter := &netlink.Route{Table: BlackholeTable, Protocol: routes.Protocol}
	rs, err := dump(func() ([]netlink.Route, error) {
		return netlink.RouteListFiltered(netlink.FAMILY_V4, filter, netlink.RT_FILTER_TABLE|netlink.RT_FILTER_PROTOCOL)
	})
	if err != nil {
		return false, fmt.Errorf("list routes in table %d: %w", BlackholeTable, err)
	}
	for _, r := range rs {
		if r.Type == syscall.RTN_BLACKHOLE {
			return true, nil
		}
	}
	return false, nil
}

func (netlinkNet) SetBlackhole(on bool) error {
	if on {
		if err := netlink.RouteReplace(blackhole()); err != nil {
			return fmt.Errorf("add the blackhole default to table %d: %w", BlackholeTable, err)
		}
		return nil
	}
	if err := netlink.RouteDel(blackhole()); err != nil {
		return fmt.Errorf("remove the blackhole default from table %d: %w", BlackholeTable, err)
	}
	return nil
}

func (netlinkNet) LinkHolding(p netip.Prefix, skip string) (string, error) {
	addrs, err := dump(func() ([]netlink.Addr, error) { return netlink.AddrList(nil, netlink.FAMILY_V4) })
	if err != nil {
		return "", fmt.Errorf("list addresses: %w", err)
	}
	for _, a := range addrs {
		if a.IPNet == nil || !p.Contains(prefixOf(a.IPNet).Addr()) {
			continue
		}
		l, err := netlink.LinkByIndex(a.LinkIndex)
		if err != nil {
			return "", fmt.Errorf("look up the interface of %s: %w", a.IPNet, err)
		}
		if name := l.Attrs().Name; name != skip {
			return name, nil
		}
	}
	return "", nil
}

func (netlinkNet) Addrs(link string) ([]netip.Prefix, error) {
	l, err := netlink.LinkByName(link)
	if err != nil {
		return nil, fmt.Errorf("look up interface %s: %w", link, err)
	}
	addrs, err := dump(func() ([]netlink.Addr, error) { return netlink.AddrList(l, netlink.FAMILY_V4) })
	if err != nil {
		return nil, fmt.Errorf("list addresses of %s: %w", link, err)
	}
	var out []netip.Prefix
	for _, a := range addrs {
		if p := prefixOf(a.IPNet); p.IsValid() {
			out = append(out, p)
		}
	}
	return out, nil
}

func (netlinkNet) AddAddr(link string, a netip.Prefix) error {
	l, err := netlink.LinkByName(link)
	if err != nil {
		return fmt.Errorf("look up interface %s: %w", link, err)
	}
	n := netconv.IPNet(a)
	return netlink.AddrAdd(l, &netlink.Addr{IPNet: &n})
}

func (netlinkNet) EnsureBridge(name string) (bool, error) {
	created := false
	l, err := netlink.LinkByName(name)
	var notFound netlink.LinkNotFoundError
	switch {
	case errors.As(err, &notFound):
		br := &netlink.Bridge{LinkAttrs: netlink.LinkAttrs{Name: name}}
		if err := netlink.LinkAdd(br); err != nil {
			return false, fmt.Errorf("create: %w", err)
		}
		if l, err = netlink.LinkByName(name); err != nil {
			return true, fmt.Errorf("look it up after creating it: %w", err)
		}
		created = true
	case err != nil:
		return false, fmt.Errorf("look it up: %w", err)
	case l.Type() != "bridge":
		return false, fmt.Errorf("%s exists but is a %s, not a bridge; rename or remove it", name, l.Type())
	}
	if l.Attrs().Flags&net.FlagUp == 0 {
		if err := netlink.LinkSetUp(l); err != nil {
			return created, fmt.Errorf("bring it up: %w", err)
		}
	}
	return created, nil
}
