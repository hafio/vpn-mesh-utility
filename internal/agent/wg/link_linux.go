//go:build linux

package wg

import (
	"errors"
	"fmt"
	"net"
	"net/netip"

	"github.com/vishvananda/netlink"

	"github.com/hafio/vpn-mesh-utility/internal/agent/netconv"
)

// links is the netlink half of both Linux backends: whether the interface
// exists and is ours, its MTU, state and IPv4 addresses. kind is the link
// type the backend creates: "wireguard" in the kernel, "tuntap" for
// wireguard-go.
type links struct{ kind string }

func lookup(name string) (netlink.Link, error) {
	l, err := netlink.LinkByName(name)
	if err != nil {
		return nil, fmt.Errorf("look up interface %s: %w", name, err)
	}
	return l, nil
}

func (n links) Link(name string) (Link, error) {
	l, err := netlink.LinkByName(name)
	var missing netlink.LinkNotFoundError
	if errors.As(err, &missing) {
		return Link{}, nil
	}
	if err != nil {
		return Link{}, fmt.Errorf("look up interface %s: %w", name, err)
	}
	a := l.Attrs()
	out := Link{Exists: true, Ours: a.Alias == Alias && l.Type() == n.kind, MTU: a.MTU, Up: a.Flags&net.FlagUp != 0}
	var addrs []netlink.Addr
	for try := 0; try < 3; try++ {
		// A dump interrupted by a concurrent change is retried.
		if addrs, err = netlink.AddrList(l, netlink.FAMILY_V4); err == nil {
			break
		}
	}
	if err != nil {
		return Link{}, fmt.Errorf("list addresses of %s: %w", name, err)
	}
	for _, ad := range addrs {
		if ad.IPNet == nil {
			continue
		}
		if p, ok := netconv.Prefix(*ad.IPNet); ok {
			out.Addrs = append(out.Addrs, p)
		}
	}
	return out, nil
}

func (links) SetMTU(name string, mtu int) error {
	l, err := lookup(name)
	if err != nil {
		return err
	}
	if err := netlink.LinkSetMTU(l, mtu); err != nil {
		return fmt.Errorf("set MTU %d on %s: %w", mtu, name, err)
	}
	return nil
}

func (links) SetUp(name string) error {
	l, err := lookup(name)
	if err != nil {
		return err
	}
	if err := netlink.LinkSetUp(l); err != nil {
		return fmt.Errorf("bring %s up: %w", name, err)
	}
	return nil
}

func (links) AddAddr(name string, p netip.Prefix) error {
	l, err := lookup(name)
	if err != nil {
		return err
	}
	n := netconv.IPNet(p)
	if err := netlink.AddrAdd(l, &netlink.Addr{IPNet: &n}); err != nil {
		return fmt.Errorf("add address %s to %s: %w", p, name, err)
	}
	return nil
}

func (links) DelAddr(name string, p netip.Prefix) error {
	l, err := lookup(name)
	if err != nil {
		return err
	}
	n := netconv.IPNet(p)
	if err := netlink.AddrDel(l, &netlink.Addr{IPNet: &n}); err != nil {
		return fmt.Errorf("remove address %s from %s: %w", p, name, err)
	}
	return nil
}
