package validate

import (
	"net/netip"
	"slices"
	"strconv"
	"strings"
)

// The grammar of segmentation (amendment A43): the addresses of objects and
// zones, the services of policies, and the kind:name references between
// them.

// AddrRange is an inclusive range of addresses of one family, IPv4 or IPv6
// (M8).
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
	if r.First.BitLen() != r.Last.BitLen() {
		return netip.Prefix{}, false
	}
	for bits := 0; bits <= r.First.BitLen(); bits++ {
		if p := netip.PrefixFrom(r.First, bits); p.Masked().Addr() == r.First && LastOf(p) == r.Last {
			return p, true
		}
	}
	return netip.Prefix{}, false
}

// PrefixRange is the range of addresses a network holds.
func PrefixRange(p netip.Prefix) AddrRange {
	return AddrRange{First: p.Masked().Addr(), Last: LastOf(p)}
}

// LastOf is the last address of a network: its host bits all set.
func LastOf(p netip.Prefix) netip.Addr {
	b := p.Masked().Addr().AsSlice()
	for i := p.Bits(); i < len(b)*8; i++ {
		b[i/8] |= 0x80 >> (i % 8)
	}
	a, _ := netip.AddrFromSlice(b)
	return a
}

// MergeRanges sorts rs and joins the ranges that overlap or touch; ranges
// of different families never meet (IPv4 sorts first).
func MergeRanges(rs []AddrRange) []AddrRange {
	if len(rs) == 0 {
		return nil
	}
	in := slices.Clone(rs)
	slices.SortFunc(in, func(a, b AddrRange) int { return a.First.Compare(b.First) })
	out := []AddrRange{in[0]}
	for _, r := range in[1:] {
		last := &out[len(out)-1]
		// Next of the family's last address is no address: nothing touches it.
		if next := last.Last.Next(); r.First.Compare(last.Last) <= 0 || (next.IsValid() && r.First == next) {
			if r.Last.Compare(last.Last) > 0 {
				last.Last = r.Last
			}
			continue
		}
		out = append(out, r)
	}
	return out
}

// RangesContain reports whether a is in one of rs.
func RangesContain(rs []AddrRange, a netip.Addr) bool {
	for _, r := range rs {
		if r.First.Compare(a) <= 0 && a.Compare(r.Last) <= 0 {
			return true
		}
	}
	return false
}

// AddrEntry validates one address of an object, a zone or a policy: an
// address, a network without host bits, or a range first-last with first
// <= last, of either family (IPv6, M8) -- a range within one.
func AddrEntry(field, s string) (AddrRange, error) {
	if first, last, ok := strings.Cut(s, "-"); ok {
		a, errA := IP(field, first)
		b, errB := IP(field, last)
		if errA != nil || errB != nil || a.Is4() != b.Is4() {
			return AddrRange{}, fail(field, s, "a range must be two addresses of one family such as 10.77.2.10-10.77.2.20 or fd10:77::2:10-fd10:77::2:20")
		}
		if b.Less(a) {
			return AddrRange{}, fail(field, s, "a range must start at or below where it ends")
		}
		return AddrRange{First: a, Last: b}, nil
	}
	if strings.Contains(s, "/") {
		p, err := Prefix(field, s)
		if err != nil {
			return AddrRange{}, err
		}
		return PrefixRange(p), nil
	}
	a, err := IP(field, s)
	if err != nil {
		return AddrRange{}, fail(field, s, "must be an address, network or range such as 10.77.1.5, 10.77.1.0/24, 10.77.1.5-10.77.1.20 or fd10:77::1:5")
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
	// An IPv6 literal has colons of its own (M8): with two or more and no
	// known kind before the first, it is a literal, valid or not.
	if !ok || (strings.Count(s, ":") >= 2 && !slices.Contains(kinds, kind)) {
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
