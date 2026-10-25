//go:build linux

package containers

import (
	"errors"
	"fmt"
	"net"
	"net/netip"
	"syscall"

	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"

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

func family(ipv6 bool) int {
	if ipv6 {
		return netlink.FAMILY_V6
	}
	return netlink.FAMILY_V4
}

// Rules lists both families' rules. The library does not report a rule's
// action, so a rule of ours with no table is the unreachable one (the only
// kind without).
func (netlinkNet) Rules() ([]Rule, error) {
	var out []Rule
	for _, v6 := range []bool{false, true} {
		rs, err := dump(func() ([]netlink.Rule, error) { return netlink.RuleList(family(v6)) })
		switch {
		case v6 && errors.Is(err, syscall.EAFNOSUPPORT):
			continue // no IPv6 policy routing: none of ours either
		case err != nil:
			return nil, fmt.Errorf("list policy rules: %w", err)
		}
		for _, r := range rs {
			if r.Protocol != routes.Protocol || !Owned(r.Priority) {
				continue
			}
			out = append(out, Rule{Priority: r.Priority, IPv6: v6, Iif: r.IifName, Src: prefixOf(r.Src), Dst: prefixOf(r.Dst),
				Table: r.Table, Unreachable: r.Table == 0})
		}
	}
	return out, nil
}

func (netlinkNet) IPv6Rules() (bool, error) {
	_, err := netlink.RuleList(netlink.FAMILY_V6)
	switch {
	case errors.Is(err, syscall.EAFNOSUPPORT):
		return false, nil
	case err != nil && !errors.Is(err, netlink.ErrDumpInterrupted):
		return false, fmt.Errorf("list IPv6 policy rules: %w", err)
	}
	return true, nil
}

func toRule(r Rule) *netlink.Rule {
	nr := netlink.NewRule()
	nr.Family, nr.Priority, nr.Table, nr.IifName, nr.Protocol = family(r.IPv6), r.Priority, r.Table, r.Iif, routes.Protocol
	if r.Unreachable {
		nr.Table, nr.Type = 0, unix.RTN_UNREACHABLE
	}
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

func blackhole(table int, ipv6 bool) *netlink.Route {
	dst := &net.IPNet{IP: net.IPv4zero.To4(), Mask: net.CIDRMask(0, 32)}
	if ipv6 {
		dst = &net.IPNet{IP: net.IPv6zero, Mask: net.CIDRMask(0, 128)}
	}
	return &netlink.Route{Table: table, Protocol: routes.Protocol, Family: family(ipv6), Type: syscall.RTN_BLACKHOLE, Dst: dst}
}

func (netlinkNet) Blackhole(table int, ipv6 bool) (bool, error) {
	filter := &netlink.Route{Table: table, Protocol: routes.Protocol}
	rs, err := dump(func() ([]netlink.Route, error) {
		return netlink.RouteListFiltered(family(ipv6), filter, netlink.RT_FILTER_TABLE|netlink.RT_FILTER_PROTOCOL)
	})
	if err != nil {
		return false, fmt.Errorf("list routes in table %d: %w", table, err)
	}
	for _, r := range rs {
		if r.Type == syscall.RTN_BLACKHOLE {
			return true, nil
		}
	}
	return false, nil
}

func (netlinkNet) SetBlackhole(table int, ipv6, on bool) error {
	what := "IPv4"
	if ipv6 {
		what = "IPv6"
	}
	if on {
		if err := netlink.RouteReplace(blackhole(table, ipv6)); err != nil {
			return fmt.Errorf("add the %s blackhole default to table %d: %w", what, table, err)
		}
		return nil
	}
	if err := netlink.RouteDel(blackhole(table, ipv6)); err != nil {
		return fmt.Errorf("remove the %s blackhole default from table %d: %w", what, table, err)
	}
	return nil
}

func dropRoute(p netip.Prefix) *netlink.Route {
	dst := netconv.IPNet(p)
	return &netlink.Route{Table: MainTable, Protocol: routes.Protocol, Family: family(p.Addr().Is6()), Type: syscall.RTN_BLACKHOLE, Dst: &dst}
}

// Dropped lists the main table's IPv6 blackhole routes of ours; the
// routes step's own (on wg0) are unicast.
func (netlinkNet) Dropped() ([]netip.Prefix, error) {
	filter := &netlink.Route{Table: MainTable, Protocol: routes.Protocol}
	rs, err := dump(func() ([]netlink.Route, error) {
		return netlink.RouteListFiltered(netlink.FAMILY_V6, filter, netlink.RT_FILTER_TABLE|netlink.RT_FILTER_PROTOCOL)
	})
	if err != nil {
		return nil, fmt.Errorf("list the main table's IPv6 routes: %w", err)
	}
	var out []netip.Prefix
	for _, r := range rs {
		if p := prefixOf(r.Dst); r.Type == syscall.RTN_BLACKHOLE && p.IsValid() {
			out = append(out, p)
		}
	}
	return out, nil
}

func (netlinkNet) SetDropped(p netip.Prefix, on bool) error {
	if on {
		return netlink.RouteReplace(dropRoute(p))
	}
	return netlink.RouteDel(dropRoute(p))
}

func (netlinkNet) LinkHolding(p netip.Prefix, skip string) (string, error) {
	addrs, err := dump(func() ([]netlink.Addr, error) { return netlink.AddrList(nil, family(p.Addr().Is6())) })
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
	addrs, err := dump(func() ([]netlink.Addr, error) { return netlink.AddrList(l, netlink.FAMILY_ALL) })
	if err != nil {
		return nil, fmt.Errorf("list addresses of %s: %w", link, err)
	}
	var out []netip.Prefix
	for _, a := range addrs {
		if p := prefixOf(a.IPNet); p.IsValid() && !p.Addr().IsLinkLocalUnicast() {
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
	addr := &netlink.Addr{IPNet: &n}
	if a.Addr().Is6() {
		addr.Flags = unix.IFA_F_NODAD
	}
	return netlink.AddrAdd(l, addr)
}

func proxy(index int, a netip.Addr) *netlink.Neigh {
	return &netlink.Neigh{LinkIndex: index, Family: netlink.FAMILY_V6, State: netlink.NUD_PERMANENT, Flags: netlink.NTF_PROXY,
		IP: net.IP(a.AsSlice())}
}

func (netlinkNet) Proxies(link string) ([]netip.Addr, error) {
	l, err := netlink.LinkByName(link)
	if err != nil {
		return nil, fmt.Errorf("look up interface %s: %w", link, err)
	}
	index := l.Attrs().Index
	ns, err := dump(func() ([]netlink.Neigh, error) { return netlink.NeighProxyList(index, netlink.FAMILY_V6) })
	if err != nil {
		return nil, fmt.Errorf("list the proxy entries of %s: %w", link, err)
	}
	var out []netip.Addr
	for _, n := range ns {
		if a, ok := netip.AddrFromSlice(n.IP); ok {
			out = append(out, a.Unmap())
		}
	}
	return out, nil
}

func (netlinkNet) AddProxy(link string, a netip.Addr) error {
	l, err := netlink.LinkByName(link)
	if err != nil {
		return fmt.Errorf("look up interface %s: %w", link, err)
	}
	return netlink.NeighAdd(proxy(l.Attrs().Index, a))
}

func (netlinkNet) DelProxy(link string, a netip.Addr) error {
	l, err := netlink.LinkByName(link)
	if err != nil {
		return fmt.Errorf("look up interface %s: %w", link, err)
	}
	return netlink.NeighDel(proxy(l.Attrs().Index, a))
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
