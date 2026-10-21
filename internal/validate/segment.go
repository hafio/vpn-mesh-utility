package validate

import (
	"math/bits"
	"net/netip"
	"strconv"
	"strings"
)

// The grammar of segmentation (amendment A43): the addresses of objects and
// zones, the services of policies, and the kind:name references between
// them.

// AddrRange is an inclusive range of IPv4 addresses.
type AddrRange struct {
	First, Last netip.Addr
}

// String is the canonical form: one address, a network, or first-last.
func (r AddrRange) String() string {
	if r.First == r.Last {
		return r.First.String()
	}
	if p, ok := r.Prefix(); ok {
		return p.String()
	}
	return r.First.String() + "-" + r.Last.String()
}

// Prefix is the network the range is exactly, when it is one.
func (r AddrRange) Prefix() (netip.Prefix, bool) {
	lo, hi := U32(r.First), U32(r.Last)
	size := uint64(hi) - uint64(lo) + 1
	if size&(size-1) != 0 { // not a power of two
		return netip.Prefix{}, false
	}
	n := bits.TrailingZeros64(size)
	if lo&uint32((uint64(1)<<n)-1) != 0 { // not aligned to its size
		return netip.Prefix{}, false
	}
	return netip.PrefixFrom(r.First, 32-n), true
}

// U32 is an IPv4 address as a number, for range arithmetic.
func U32(a netip.Addr) uint32 {
	b := a.As4()
	return uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3])
}

// FromU32 is the address of a number.
func FromU32(n uint32) netip.Addr {
	return netip.AddrFrom4([4]byte{byte(n >> 24), byte(n >> 16), byte(n >> 8), byte(n)})
}

// AddrEntry validates one address of an object, a zone or a policy: an IPv4
// address, an IPv4 network without host bits, or a range first-last with
// first <= last.
func AddrEntry(field, s string) (AddrRange, error) {
	if first, last, ok := strings.Cut(s, "-"); ok {
		a, errA := netip.ParseAddr(first)
		b, errB := netip.ParseAddr(last)
		if errA != nil || errB != nil || !a.Is4() || !b.Is4() {
			return AddrRange{}, fail(field, s, "a range must be two IPv4 addresses such as 10.77.2.10-10.77.2.20")
		}
		if b.Less(a) {
			return AddrRange{}, fail(field, s, "a range must start at or below where it ends")
		}
		return AddrRange{First: a, Last: b}, nil
	}
	if strings.Contains(s, "/") {
		p, err := Prefix4(field, s)
		if err != nil {
			return AddrRange{}, err
		}
		return AddrRange{First: p.Addr(), Last: FromU32(U32(p.Addr()) | uint32(uint64(1)<<(32-p.Bits())-1))}, nil
	}
	a, err := netip.ParseAddr(s)
	if err != nil || !a.Is4() {
		return AddrRange{}, fail(field, s, "must be an IPv4 address, network or range such as 10.77.1.5, 10.77.1.0/24 or 10.77.1.5-10.77.1.20")
	}
	return AddrRange{First: a, Last: a}, nil
}

// Service protocols. ServiceICMP is ping: echo requests (the replies, and
// ICMP errors, belong to a flow the policy accepted).
const (
	ServiceTCP  = "tcp"
	ServiceUDP  = "udp"
	ServiceICMP = "icmp"
)

// Service is a protocol and an inclusive port range. A protocol with every
// port, and icmp, have no range (0, 0).
type Service struct {
	Proto       string
	First, Last int
}

// String is the canonical form: "tcp", "tcp/22", "udp/1000-2000", "icmp".
func (s Service) String() string {
	switch {
	case s.First == 0:
		return s.Proto
	case s.First == s.Last:
		return s.Proto + "/" + strconv.Itoa(s.First)
	}
	return s.Proto + "/" + strconv.Itoa(s.First) + "-" + strconv.Itoa(s.Last)
}

// Ports is the port part of String: "", "22" or "1000-2000".
func (s Service) Ports() string {
	_, p, _ := strings.Cut(s.String(), "/")
	return p
}

// ServiceEntry validates a service: tcp or udp, optionally with one port or
// a port range ("tcp", "tcp/22", "udp/1000-2000"), or icmp.
func ServiceEntry(field, s string) (Service, error) {
	proto, ports, hasPorts := strings.Cut(s, "/")
	if proto == ServiceICMP {
		if hasPorts {
			return Service{}, fail(field, s, "icmp (ping) has no ports")
		}
		return Service{Proto: ServiceICMP}, nil
	}
	if proto != ServiceTCP && proto != ServiceUDP {
		return Service{}, fail(field, s, "must be tcp, udp, tcp/PORT, udp/PORT, tcp/FIRST-LAST or icmp")
	}
	if !hasPorts {
		return Service{Proto: proto}, nil
	}
	first, last, isRange := strings.Cut(ports, "-")
	if !isRange {
		last = first
	}
	a, errA := strconv.Atoi(first)
	b, errB := strconv.Atoi(last)
	switch {
	case errA != nil || errB != nil || a < 1 || b > 65535 || strconv.Itoa(a) != first || strconv.Itoa(b) != last:
		return Service{}, fail(field, s, "ports must be numbers between 1 and 65535, such as tcp/22 or udp/1000-2000")
	case b < a:
		return Service{}, fail(field, s, "a port range must start at or below where it ends")
	}
	return Service{Proto: proto, First: a, Last: b}, nil
}

// Ref splits a reference "kind:name" whose kind is one of kinds and whose
// name is a Name. A string without a colon is not a reference: ok is false
// and the caller validates it as a literal.
func Ref(field, s string, kinds ...string) (kind, name string, ok bool, err error) {
	kind, name, ok = strings.Cut(s, ":")
	if !ok {
		return "", "", false, nil
	}
	if _, err := Enum(field, kind, kinds...); err != nil {
		return "", "", true, fail(field, s, "a reference is kind:name with kind one of "+strings.Join(kinds, ", "))
	}
	if _, err := Name(field, name); err != nil {
		return "", "", true, fail(field, s, "the name after "+kind+": must be 1-32 chars of a-z, 0-9 and '-'")
	}
	return kind, name, true, nil
}
