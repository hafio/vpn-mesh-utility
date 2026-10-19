package dns

import (
	"fmt"
	"net"
	"net/netip"
	"strconv"
	"strings"

	mdns "github.com/miekg/dns"

	"github.com/hafio/vpn-mesh-utility/internal/proto"
	"github.com/hafio/vpn-mesh-utility/internal/validate"
)

// ttl is short so a name follows its node's mesh IP within a minute.
const ttl = 60

// zone answers queries from one config's DNS section. It never changes once
// built; Apply swaps in a new one.
type zone struct {
	// apex is the zone's own name, fully qualified ("mesh.").
	apex string
	// names holds every name in the zone, lower case and fully qualified.
	// An empty non-terminal (office.mesh. above printer.office.mesh.) has
	// the zero address: it exists, but has no record.
	names      map[string]netip.Addr
	upstream   []string
	split      bool
	forwardFor []netip.Prefix
}

func newZone(d *proto.DNS, upstreamPort int) (*zone, error) {
	if _, err := validate.Zone("dns: zone", d.Zone); err != nil {
		return nil, err
	}
	apex := d.Zone + "."
	z := &zone{apex: apex, names: map[string]netip.Addr{apex: {}}, split: d.Split}
	for _, r := range d.Records {
		ip, err := netip.ParseAddr(r.IP)
		if err != nil || !ip.Is4() {
			return nil, fmt.Errorf("dns: record %q: %q is not an IPv4 address", r.Name, r.IP)
		}
		name := strings.ToLower(r.Name) + "." + apex
		z.names[name] = ip
		for p := parent(name); p != apex; p = parent(p) {
			if _, ok := z.names[p]; !ok {
				z.names[p] = netip.Addr{}
			}
		}
	}
	for _, u := range d.Upstream {
		ip, err := netip.ParseAddr(u)
		if err != nil {
			return nil, fmt.Errorf("dns: upstream %q is not an IP address", u)
		}
		z.upstream = append(z.upstream, net.JoinHostPort(ip.String(), strconv.Itoa(upstreamPort)))
	}
	for _, f := range d.ForwardFor {
		p, err := netip.ParsePrefix(f)
		if err != nil {
			return nil, fmt.Errorf("dns: forward_for %q is not a network", f)
		}
		z.forwardFor = append(z.forwardFor, p)
	}
	return z, nil
}

// parent drops the first label of a fully qualified name.
func parent(name string) string { return name[strings.IndexByte(name, '.')+1:] }

func (z *zone) soa() mdns.RR {
	return &mdns.SOA{Hdr: mdns.RR_Header{Name: z.apex, Rrtype: mdns.TypeSOA, Class: mdns.ClassINET, Ttl: ttl},
		Ns: z.apex, Mbox: "hostmaster." + z.apex, Serial: 1, Refresh: 3600, Retry: 600, Expire: 86400, Minttl: ttl}
}

// mayForward reports whether a query from src for another zone is
// forwarded: in exit mode (not split), or from a phone that uses this node
// as its resolver.
func (z *zone) mayForward(src netip.Addr) bool {
	if !z.split {
		return true
	}
	for _, p := range z.forwardFor {
		if p.Contains(src) {
			return true
		}
	}
	return false
}

// answer resolves req. forward means it is for another zone and src may
// have it forwarded; resp is nil then.
func (z *zone) answer(req *mdns.Msg, src netip.Addr) (resp *mdns.Msg, forward bool) {
	resp = new(mdns.Msg).SetReply(req)
	if len(req.Question) != 1 {
		resp.Rcode = mdns.RcodeFormatError
		return resp, false
	}
	q := req.Question[0]
	name := strings.ToLower(q.Name)
	if name != z.apex && !strings.HasSuffix(name, "."+z.apex) {
		if z.mayForward(src) {
			return nil, true
		}
		resp.Rcode = mdns.RcodeRefused
		return resp, false
	}
	resp.Authoritative = true
	ip, ok := z.names[name]
	switch {
	case q.Qclass != mdns.ClassINET && q.Qclass != mdns.ClassANY:
		resp.Authoritative, resp.Rcode = false, mdns.RcodeRefused
	case !ok:
		resp.Rcode, resp.Ns = mdns.RcodeNameError, []mdns.RR{z.soa()}
	case ip.IsValid() && (q.Qtype == mdns.TypeA || q.Qtype == mdns.TypeANY):
		// The name as asked: resolvers that randomise its case check it.
		resp.Answer = []mdns.RR{&mdns.A{Hdr: mdns.RR_Header{Name: q.Name, Rrtype: mdns.TypeA, Class: mdns.ClassINET, Ttl: ttl},
			A: net.IP(ip.AsSlice())}}
	case name == z.apex && q.Qtype == mdns.TypeSOA:
		resp.Answer = []mdns.RR{z.soa()}
	default:
		// The name exists without a record of this type (AAAA, say).
		resp.Ns = []mdns.RR{z.soa()}
	}
	return resp, false
}
