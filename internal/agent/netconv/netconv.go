// Package netconv converts between net/netip, which the agent's own code
// uses, and the net.IPNet and net.UDPAddr types of the netlink and wgctrl
// libraries.
package netconv

import (
	"net"
	"net/netip"
)

// IPNet converts a prefix.
func IPNet(p netip.Prefix) net.IPNet {
	return net.IPNet{IP: p.Addr().AsSlice(), Mask: net.CIDRMask(p.Bits(), p.Addr().BitLen())}
}

// Prefix converts an IPNet, unmapping IPv4-in-IPv6 forms. ok is false for a
// non-canonical mask or a malformed address.
func Prefix(n net.IPNet) (netip.Prefix, bool) {
	ip, ok := netip.AddrFromSlice(n.IP)
	ones, bits := n.Mask.Size()
	if !ok || bits == 0 {
		return netip.Prefix{}, false
	}
	switch {
	case bits == 32:
		ip = ip.Unmap()
	case bits == 128 && ip.Is4In6():
		ip, ones = ip.Unmap(), ones-96
	}
	return netip.PrefixFrom(ip, ones), true
}

// AddrPort converts a UDP address; the zero value for nil.
func AddrPort(a *net.UDPAddr) netip.AddrPort {
	if a == nil {
		return netip.AddrPort{}
	}
	ap := a.AddrPort()
	return netip.AddrPortFrom(ap.Addr().Unmap(), ap.Port())
}
