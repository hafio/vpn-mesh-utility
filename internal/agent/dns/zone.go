package dns

import (
	"fmt"
	"net"
	"net/netip"
	"slices"
	"strconv"
	"strings"
	"sync/atomic"

	mdns "github.com/miekg/dns"

	"github.com/hafio/vpn-mesh-utility/internal/proto"
	"github.com/hafio/vpn-mesh-utility/internal/validate"
)

// ttl is short so a name follows its node's mesh IP within a minute; a
// container's name (amendment A38) follows its containers within seconds.
const (
	ttl     = 60
	nameTTL = 10
)

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
	// domains are the extra domains (amendment A39), fully qualified.
	domains []string
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
	for _, dom := range d.Domains {
		z.domains = append(z.domains, strings.ToLower(dom)+".")
	}
	return z, nil
}

// inDomain reports whether name (fully qualified, lower case) is one of the
// extra domains or under one.
func (z *zone) inDomain(name string) bool {
	for _, d := range z.domains {
		if name == d || strings.HasSuffix(name, "."+d) {
			return true
		}
	}
	return false
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

// answer resolves req, with the containers' names t (nil: none). forward
// means it is for another zone and src may have it forwarded; resp is nil
// then. A name of the config wins over a container's.
func (z *zone) answer(req *mdns.Msg, src netip.Addr, t *nameTable) (resp *mdns.Msg, forward bool) {
	resp = new(mdns.Msg).SetReply(req)
	if len(req.Question) != 1 {
		resp.Rcode = mdns.RcodeFormatError
		return resp, false
	}
	q := req.Question[0]
	name := strings.ToLower(q.Name)
	if z.inDomain(name) {
		return z.answerDomain(req, resp, name, t)
	}
	if name != z.apex && !strings.HasSuffix(name, "."+z.apex) {
		if z.mayForward(src) {
			return nil, true
		}
		resp.Rcode = mdns.RcodeRefused
		return resp, false
	}
	resp.Authoritative = true
	ip, ok := z.names[name]
	var addrs []netip.Addr
	if !ip.IsValid() {
		addrs = t.lookup(strings.TrimSuffix(name, "."+z.apex))
		ok = ok || len(addrs) > 0
	}
	switch {
	case q.Qclass != mdns.ClassINET && q.Qclass != mdns.ClassANY:
		resp.Authoritative, resp.Rcode = false, mdns.RcodeRefused
	case !ok:
		resp.Rcode, resp.Ns = mdns.RcodeNameError, []mdns.RR{z.soa()}
	case len(addrs) > 0 && (q.Qtype == mdns.TypeA || q.Qtype == mdns.TypeANY):
		for _, a := range addrs {
			resp.Answer = append(resp.Answer, &mdns.A{Hdr: mdns.RR_Header{Name: q.Name, Rrtype: mdns.TypeA, Class: mdns.ClassINET, Ttl: nameTTL},
				A: net.IP(a.AsSlice())})
		}
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

// answerDomain answers a name in an extra domain (amendment A39): a
// container's name there from t, anything else from the upstream resolvers,
// whoever asks -- the host sends the whole domain here, so a miss must still
// resolve, and only these domains are forwarded so for everyone.
func (z *zone) answerDomain(req, resp *mdns.Msg, name string, t *nameTable) (*mdns.Msg, bool) {
	q := req.Question[0]
	addrs := t.lookupFQDN(strings.TrimSuffix(name, "."))
	switch {
	case len(addrs) == 0:
		return nil, true
	case q.Qclass != mdns.ClassINET && q.Qclass != mdns.ClassANY:
		resp.Rcode = mdns.RcodeRefused
	case q.Qtype == mdns.TypeA || q.Qtype == mdns.TypeANY:
		for _, a := range addrs {
			resp.Answer = append(resp.Answer, &mdns.A{Hdr: mdns.RR_Header{Name: q.Name, Rrtype: mdns.TypeA, Class: mdns.ClassINET, Ttl: nameTTL},
				A: net.IP(a.AsSlice())})
		}
	}
	// Another type of a container's name: no record, and no authority over
	// the domain to say more.
	return resp, false
}

// nameTable is the containers' published names (amendment A38). In the
// zone they are relative to it: "web" holds every node's addresses,
// "web.nas" the NAS's. A dotted name is one in an extra domain (amendment
// A39), held as it is. It never changes once built; SetNames swaps in a new
// one.
type nameTable struct {
	addrs map[string][]netip.Addr
	fqdns map[string][]netip.Addr
	// next rotates the answers: round robin across the addresses.
	next atomic.Uint32
}

func newNameTable(names []proto.MeshName) *nameTable {
	t := &nameTable{addrs: map[string][]netip.Addr{}, fqdns: map[string][]netip.Addr{}}
	add := func(m map[string][]netip.Addr, key string, ip netip.Addr) {
		if !slices.Contains(m[key], ip) {
			m[key] = append(m[key], ip)
		}
	}
	for _, n := range names {
		ip, err := netip.ParseAddr(n.IP)
		if err != nil || !ip.Is4() {
			continue // validated by the client
		}
		name := strings.ToLower(n.Name)
		if strings.Contains(name, ".") {
			add(t.fqdns, name, ip)
			continue
		}
		add(t.addrs, name, ip)
		add(t.addrs, name+"."+strings.ToLower(n.Node), ip)
	}
	for _, m := range []map[string][]netip.Addr{t.addrs, t.fqdns} {
		for _, a := range m {
			slices.SortFunc(a, netip.Addr.Compare)
		}
	}
	return t
}

// lookup returns the addresses of rel, a name relative to the zone,
// rotated by one on every call; nil when the table is nil or has no such
// name.
func (t *nameTable) lookup(rel string) []netip.Addr {
	if t == nil {
		return nil
	}
	return t.rotate(t.addrs[rel])
}

// lookupFQDN is lookup for a name in an extra domain.
func (t *nameTable) lookupFQDN(name string) []netip.Addr {
	if t == nil {
		return nil
	}
	return t.rotate(t.fqdns[name])
}

func (t *nameTable) rotate(a []netip.Addr) []netip.Addr {
	if len(a) < 2 {
		return a
	}
	i := int(t.next.Add(1) % uint32(len(a)))
	return append(slices.Clone(a[i:]), a[:i]...)
}
