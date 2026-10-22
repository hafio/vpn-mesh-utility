package validate

import "net/netip"

// The IPv6 side of the address grammar (M8, dual stack). An IPv6 value is
// a plain global or unique-local address: no zone (fe80::1%eth0), no
// IPv4-mapped form (::ffff:10.77.0.1), which would be an IPv4 address in
// disguise.

func plain6(a netip.Addr) bool { return a.Is6() && !a.Is4In6() && a.Zone() == "" }

// IPv6 validates a bare IPv6 address.
func IPv6(field, s string) (netip.Addr, error) {
	a, err := netip.ParseAddr(s)
	if err != nil || !plain6(a) {
		return netip.Addr{}, fail(field, s, "must be an IPv6 address such as fd00:10:77::1")
	}
	return a, nil
}

// Prefix6 validates an IPv6 CIDR without host bits.
func Prefix6(field, s string) (netip.Prefix, error) {
	p, err := netip.ParsePrefix(s)
	if err != nil || !plain6(p.Addr()) {
		return netip.Prefix{}, fail(field, s, "must be an IPv6 CIDR such as fd00:10:77::/64")
	}
	if p.Masked() != p {
		return netip.Prefix{}, fail(field, s, "has host bits set; did you mean "+p.Masked().String()+"?")
	}
	return p, nil
}

// IP validates an address of either family.
func IP(field, s string) (netip.Addr, error) {
	if a, err := IPv4(field, s); err == nil {
		return a, nil
	}
	if a, err := IPv6(field, s); err == nil {
		return a, nil
	}
	return netip.Addr{}, fail(field, s, "must be an IPv4 or IPv6 address such as 10.77.0.1 or fd00:10:77::1")
}

// Prefix validates a CIDR of either family, without host bits.
func Prefix(field, s string) (netip.Prefix, error) {
	if p, err := netip.ParsePrefix(s); err == nil && p.Addr().Is4() {
		return Prefix4(field, s)
	}
	if p, err := netip.ParsePrefix(s); err == nil && plain6(p.Addr()) {
		return Prefix6(field, s)
	}
	return netip.Prefix{}, fail(field, s, "must be an IPv4 or IPv6 CIDR such as 10.77.0.0/16 or fd00:10:77::/64")
}

// Endpoint6 validates an IPv6 WireGuard endpoint, [address]:port.
func Endpoint6(field, s string) (netip.AddrPort, error) {
	ap, err := netip.ParseAddrPort(s)
	if err != nil || !plain6(ap.Addr()) || ap.Port() == 0 {
		return netip.AddrPort{}, fail(field, s, "must be [IPv6]:port such as [2001:db8::10]:51820")
	}
	return ap, nil
}

// PeerEndpoint validates a peer's endpoint in a config: IPv4:port or
// [IPv6]:port.
func PeerEndpoint(field, s string) (netip.AddrPort, error) {
	if ap, err := Endpoint(field, s); err == nil {
		return ap, nil
	}
	if ap, err := Endpoint6(field, s); err == nil {
		return ap, nil
	}
	return netip.AddrPort{}, fail(field, s, "must be IPv4:port or [IPv6]:port such as 203.0.113.10:51820 or [2001:db8::10]:51820")
}
