package planner

import (
	"cmp"
	"fmt"
	"maps"
	"net/netip"
	"slices"
	"strconv"
	"strings"

	"github.com/hafio/vpn-mesh-utility/internal/proto"
	"github.com/hafio/vpn-mesh-utility/internal/validate"
)

// Segmentation (amendment A43) as the planner sees it: the mesh firewall's
// document, its references unresolved. Every packet between two mesh
// addresses crosses a server (spokes and phones peer only with servers), so
// the servers alone enforce it: each gets the same compiled proto.Policy.
type Segmentation struct {
	// Groups are node names by group; Addresses and Services the entries
	// of each object.
	Groups    map[string][]string
	Addresses map[string][]string
	Services  map[string][]string
	Zones     []SegZone
	// Policies are in order.
	Policies []SegPolicy
}

// SegZone is a zone: its members, and what traffic within it gets when no
// policy matched it first (proto.PolicyAllow or PolicyDeny).
type SegZone struct {
	Name    string
	Intra   string
	Members []string
}

// SegPolicy is one policy; an empty From, To or Services is any.
type SegPolicy struct {
	Name     string
	Disabled bool
	From     []string
	To       []string
	Services []string
	Action   string
}

// The reference kinds (as in store) and the roles a role: reference names.
const (
	refNode       = "node"
	refContainers = "containers"
	refGroup      = "group"
	refRole       = "role"
	refAddr       = "addr"
	refZone       = "zone"
	refService    = "svc"
)

var (
	memberKinds = []string{refNode, refContainers, refGroup, refRole, refAddr}
	policyKinds = []string{refZone, refNode, refContainers, refGroup, refRole, refAddr}
)

// systemPrefix names the rules every policy starts with, which no policy
// changes: ping across the mesh, DNS to the servers, Swarm between them.
const systemPrefix = "system: "

// swarmPorts are the ports Swarm uses between its nodes (spec S7.5).
var swarmPorts = []proto.Service{{Proto: validate.ServiceTCP, Ports: []string{"2377", "7946"}},
	{Proto: validate.ServiceUDP, Ports: []string{"4789", "7946"}}}

// span is an inclusive range of addresses as numbers; spans is a sorted,
// merged set of them.
type span struct{ lo, hi uint32 }
type spans []span

func spanOf(r validate.AddrRange) span { return span{validate.U32(r.First), validate.U32(r.Last)} }

// merged sorts and merges s, joining ranges that touch.
func (s spans) merged() spans {
	if len(s) == 0 {
		return nil
	}
	in := slices.Clone(s)
	slices.SortFunc(in, func(a, b span) int { return cmp.Compare(a.lo, b.lo) })
	out := spans{in[0]}
	for _, x := range in[1:] {
		last := &out[len(out)-1]
		if uint64(x.lo) <= uint64(last.hi)+1 {
			last.hi = max(last.hi, x.hi)
			continue
		}
		out = append(out, x)
	}
	return out
}

// strings is the set in validate.AddrEntry's canonical form.
func (s spans) strings() []string {
	var out []string
	for _, x := range s {
		out = append(out, validate.AddrRange{First: validate.FromU32(x.lo), Last: validate.FromU32(x.hi)}.String())
	}
	return out
}

func (s spans) overlaps(o spans) bool {
	for _, a := range s {
		for _, b := range o {
			if a.lo <= b.hi && b.lo <= a.hi {
				return true
			}
		}
	}
	return false
}

// covers reports whether every address of o is in s (both merged).
func (s spans) covers(o spans) bool {
	for _, b := range o {
		ok := false
		for _, a := range s {
			if a.lo <= b.lo && b.hi <= a.hi {
				ok = true
				break
			}
		}
		if !ok {
			return false
		}
	}
	return true
}

// parseSpans reads canonical entries back; one that does not parse is left
// out (configs are validated before anything reads them).
func parseSpans(in []string) spans {
	var out spans
	for _, e := range in {
		if r, err := validate.AddrEntry("entry", e); err == nil {
			out = append(out, spanOf(r))
		}
	}
	return out.merged()
}

// segResolver resolves references against one Input. report receives every
// problem it meets; Compute's second pass discards them (Check has them).
type segResolver struct {
	in     Input
	seg    *Segmentation
	nodes  map[string]Node
	zones  map[string]SegZone
	memo   map[string]spans
	report func(warning bool, format string, args ...any)
}

func newSegResolver(in Input, report func(bool, string, ...any)) *segResolver {
	r := &segResolver{in: in, seg: &in.Segmentation, nodes: map[string]Node{}, zones: map[string]SegZone{},
		memo: map[string]spans{}, report: report}
	for _, n := range in.Nodes {
		r.nodes[n.Name] = n
	}
	for _, z := range in.Segmentation.Zones {
		r.zones[z.Name] = z
	}
	return r
}

// literal is an address entry, clipped to the mesh: the mesh firewall only
// sees mesh addresses, so 0.0.0.0/0 (a custom rule's source, say) is the
// whole mesh, and an entry wholly outside it matches nothing.
func (r *segResolver) literal(where, s string) (span, bool) {
	a, err := validate.AddrEntry(where, s)
	if err != nil {
		r.report(false, "%v", err)
		return span{}, false
	}
	x, ok := r.clip(spanOf(a))
	if !ok {
		r.report(true, "%s: %s is outside the mesh %s, so it matches no address", where, s, r.in.MeshCIDR)
	}
	return x, ok
}

// clip is the part of x inside the mesh; ok is false when there is none.
func (r *segResolver) clip(x span) (span, bool) {
	m := r.in.MeshCIDR
	if !m.IsValid() {
		return x, true
	}
	mesh, _ := validate.AddrEntry("mesh", m.String()) // a valid prefix
	ms := spanOf(mesh)
	lo, hi := max(x.lo, ms.lo), min(x.hi, ms.hi)
	return span{lo, hi}, lo <= hi
}

func (r *segResolver) node(where, name string) (Node, bool) {
	n, ok := r.nodes[name]
	if !ok {
		r.report(false, "%s: there is no node %q; add it, or take the reference out first", where, name)
	}
	return n, ok
}

func host(a netip.Addr) spans {
	if !a.Is4() {
		return nil
	}
	n := validate.U32(a)
	return spans{{n, n}}
}

// refs resolves a list (a zone's members, a policy's From or To) to one
// merged set.
func (r *segResolver) refs(where string, in []string, kinds []string) spans {
	var out spans
	for i, s := range in {
		out = append(out, r.ref(fmt.Sprintf("%s[%d] %q", where, i, s), s, kinds)...)
	}
	return out.merged()
}

func (r *segResolver) ref(where, s string, kinds []string) spans {
	kind, name, isRef, err := validate.Ref(where, s, kinds...)
	switch {
	case err != nil:
		r.report(false, "%v", err)
		return nil
	case !isRef:
		if x, ok := r.literal(where, s); ok {
			return spans{x}
		}
		return nil
	}
	switch kind {
	case refNode:
		n, _ := r.node(where, name)
		return host(n.MeshIP)
	case refContainers:
		n, ok := r.node(where, name)
		if !ok {
			return nil
		}
		if !n.Containers.IsValid() {
			r.report(true, "%s: %q has no container network, so this matches no address", where, name)
			return nil
		}
		k, _ := validate.AddrEntry(where, n.Containers.String()) // a /24, checked with the node
		return spans{spanOf(k)}
	case refGroup:
		members, ok := r.seg.Groups[name]
		if !ok {
			r.report(false, "%s: there is no group %q; a node joins one with `meshctl node set NAME --groups %s`", where, name, name)
			return nil
		}
		var out spans
		for _, m := range members {
			if n, ok := r.nodes[m]; ok {
				out = append(out, host(n.MeshIP)...)
			}
		}
		return out
	case refRole:
		var out spans
		switch proto.Role(name) {
		case proto.RoleServer, proto.RoleSpoke, proto.RolePhone:
		default:
			r.report(false, "%s: the roles are server, spoke and phone", where)
			return nil
		}
		for _, n := range r.in.Nodes {
			if n.Role == proto.Role(name) {
				out = append(out, host(n.MeshIP)...)
			}
		}
		return out
	case refAddr:
		entries, ok := r.seg.Addresses[name]
		if !ok {
			r.report(false, "%s: there is no address object %q", where, name)
			return nil
		}
		var out spans
		for _, x := range parseSpans(entries) { // reported once, with the object
			if c, ok := r.clip(x); ok {
				out = append(out, c)
			}
		}
		return out
	case refZone:
		if _, ok := r.zones[name]; !ok {
			r.report(false, "%s: there is no zone %q", where, name)
			return nil
		}
		return r.zone(name)
	}
	return nil
}

// zone is a zone's addresses, resolved once.
func (r *segResolver) zone(name string) spans {
	if s, ok := r.memo[name]; ok {
		return s
	}
	z := r.zones[name]
	s := r.refs(fmt.Sprintf("zone %q: members", name), z.Members, memberKinds)
	r.memo[name] = s
	return s
}

// services resolves a policy's services to proto's form: per protocol,
// every port when any entry has none, else the merged ranges.
func (r *segResolver) services(where string, in []string) []proto.Service {
	var all []validate.Service
	for i, s := range in {
		f := fmt.Sprintf("%s[%d] %q", where, i, s)
		_, name, isRef, err := validate.Ref(f, s, refService)
		switch {
		case err != nil:
			r.report(false, "%v", err)
		case isRef:
			entries, ok := r.seg.Services[name]
			if !ok {
				r.report(false, "%s: there is no service object %q", f, name)
			}
			for _, e := range entries {
				if svc, err := validate.ServiceEntry(f, e); err == nil { // checked with the object
					all = append(all, svc)
				}
			}
		default:
			svc, err := validate.ServiceEntry(f, s)
			if err != nil {
				r.report(false, "%v", err)
				continue
			}
			all = append(all, svc)
		}
	}
	return mergeServices(all)
}

func mergeServices(in []validate.Service) []proto.Service {
	ports := map[string]spans{}
	every := map[string]bool{}
	for _, s := range in {
		if s.First == 0 {
			every[s.Proto] = true
			continue
		}
		ports[s.Proto] = append(ports[s.Proto], span{uint32(s.First), uint32(s.Last)})
	}
	protos := maps.Clone(every)
	for p := range ports {
		protos[p] = true
	}
	var out []proto.Service
	for _, p := range slices.Sorted(maps.Keys(protos)) {
		svc := proto.Service{Proto: p}
		if !every[p] {
			for _, x := range ports[p].merged() {
				if x.lo == x.hi {
					svc.Ports = append(svc.Ports, strconv.FormatUint(uint64(x.lo), 10))
				} else {
					svc.Ports = append(svc.Ports, fmt.Sprintf("%d-%d", x.lo, x.hi))
				}
			}
		}
		out = append(out, svc)
	}
	return out
}

// compile is the policy every server gets: the system rules, the enabled
// policies in order, then an allow within each zone whose Intra is allow;
// what matches none is dropped. A policy whose From, To or Services
// resolves to nothing matches nothing and is left out (Check warns), never
// widened to any.
func (r *segResolver) compile(servers []Node) *proto.Policy {
	pol := &proto.Policy{Mesh: r.in.MeshCIDR.String(), Rules: []proto.PolicyRule{
		{Name: systemPrefix + "ping", Services: []proto.Service{{Proto: validate.ServiceICMP}}, Action: proto.PolicyAllow},
	}}
	var srv spans
	for _, s := range servers {
		srv = append(srv, host(s.MeshIP)...)
	}
	if srv = srv.merged(); len(srv) > 0 {
		dns := []proto.Service{{Proto: validate.ServiceTCP, Ports: []string{"53"}}, {Proto: validate.ServiceUDP, Ports: []string{"53"}}}
		pol.Rules = append(pol.Rules,
			proto.PolicyRule{Name: systemPrefix + "dns", To: srv.strings(), Services: dns, Action: proto.PolicyAllow},
			proto.PolicyRule{Name: systemPrefix + "swarm", From: srv.strings(), To: srv.strings(), Services: swarmPorts, Action: proto.PolicyAllow})
	}
	for _, p := range r.seg.Policies {
		if p.Disabled {
			continue
		}
		where := fmt.Sprintf("policy %q", p.Name)
		rule := proto.PolicyRule{Name: "policy: " + p.Name, Action: p.Action}
		empty := ""
		if len(p.From) > 0 {
			if rule.From = r.refs(where+": from", p.From, policyKinds).strings(); rule.From == nil {
				empty = "from"
			}
		}
		if len(p.To) > 0 {
			if rule.To = r.refs(where+": to", p.To, policyKinds).strings(); rule.To == nil && empty == "" {
				empty = "to"
			}
		}
		if len(p.Services) > 0 {
			if rule.Services = r.services(where+": services", p.Services); rule.Services == nil && empty == "" {
				empty = "services"
			}
		}
		if empty != "" {
			r.report(true, "%s matches no traffic (its %s resolves to nothing), so it is left out", where, empty)
			continue
		}
		pol.Rules = append(pol.Rules, rule)
	}
	for _, z := range slices.SortedFunc(slices.Values(r.seg.Zones), func(a, b SegZone) int { return cmp.Compare(a.Name, b.Name) }) {
		if z.Intra != proto.PolicyAllow {
			continue
		}
		if s := r.zone(z.Name).strings(); s != nil {
			pol.Rules = append(pol.Rules, proto.PolicyRule{Name: "zone: " + z.Name, From: s, To: s, Action: proto.PolicyAllow})
		}
	}
	return pol
}

// checkSegmentation reports what is wrong with the segmentation document as
// the mesh stands: references to nothing, addresses outside the mesh, and
// what the operator should know about the servers' policy.
func checkSegmentation(in Input, servers []Node, errf, warnf func(node, format string, args ...any)) {
	seg := in.Segmentation
	r := newSegResolver(in, func(warning bool, format string, args ...any) {
		if warning {
			warnf("", format, args...)
			return
		}
		errf("", format, args...)
	})
	for _, kind := range []struct {
		what    string
		objects map[string][]string
		check   func(where, e string)
	}{
		{"address object", seg.Addresses, func(where, e string) { r.literal(where, e) }},
		{"service object", seg.Services, func(where, e string) {
			if _, err := validate.ServiceEntry(where, e); err != nil {
				errf("", "%v", err)
			}
		}},
	} {
		for _, name := range slices.Sorted(maps.Keys(kind.objects)) {
			for i, e := range kind.objects[name] {
				kind.check(fmt.Sprintf("%s %q: entries[%d]", kind.what, name, i), e)
			}
		}
	}
	for _, g := range slices.Sorted(maps.Keys(seg.Groups)) {
		for _, n := range seg.Groups[g] {
			if _, ok := r.nodes[n]; !ok {
				errf("", "group %q: there is no node %q", g, n)
			}
		}
	}
	for _, z := range seg.Zones {
		if _, err := validate.Enum(fmt.Sprintf("zone %q: intra", z.Name), z.Intra, proto.PolicyAllow, proto.PolicyDeny); err != nil {
			errf("", "%v", err)
		}
		s := r.zone(z.Name)
		if len(s) == 0 {
			warnf("", "zone %q has no addresses: policies that name it match nothing from or to it", z.Name)
		}
	}
	for _, p := range seg.Policies {
		if _, err := validate.Enum(fmt.Sprintf("policy %q: action", p.Name), p.Action, proto.PolicyAllow, proto.PolicyDeny); err != nil {
			errf("", "%v", err)
		}
	}
	// Disabled policies' references must hold too: enabling one later must
	// not be the change that fails.
	for _, p := range seg.Policies {
		if p.Disabled {
			where := fmt.Sprintf("policy %q (disabled)", p.Name)
			r.refs(where+": from", p.From, policyKinds)
			r.refs(where+": to", p.To, policyKinds)
			r.services(where+": services", p.Services)
		}
	}
	pol := r.compile(servers)
	if len(pol.Rules) > proto.MaxPolicyRules {
		errf("", "the policies compile to %d rules, more than the %d a server takes; merge some", len(pol.Rules), proto.MaxPolicyRules)
	}
	allows := false
	for _, rule := range pol.Rules {
		if n := max(len(rule.From), len(rule.To)); n > proto.MaxPolicyEntries {
			errf("", "%s has %d address ranges, more than the %d a rule takes; use wider networks", rule.Name, n, proto.MaxPolicyEntries)
		}
		allows = allows || (rule.Action == proto.PolicyAllow && !strings.HasPrefix(rule.Name, systemPrefix))
	}
	if !allows {
		warnf("", "no policy or zone allows any traffic: between nodes only ping, DNS to the servers and Swarm between them cross the mesh")
	}
}

// CompilePolicy is the policy every server gets. in is assumed to pass
// Check.
func CompilePolicy(in Input) *proto.Policy {
	return newSegResolver(in, func(bool, string, ...any) {}).compile(newPlan(in).servers)
}

// ZoneAddresses is each zone's addresses, canonical and merged.
func ZoneAddresses(in Input) map[string][]string {
	out := map[string][]string{}
	r := newSegResolver(in, func(bool, string, ...any) {})
	for _, z := range in.Segmentation.Zones {
		out[z.Name] = r.zone(z.Name).strings()
	}
	return out
}

// Overlaps reports whether a rule side (nil is any) shares an address with
// addrs.
func Overlaps(side, addrs []string) bool {
	if side == nil {
		return len(addrs) > 0
	}
	return parseSpans(side).overlaps(parseSpans(addrs))
}

// Verdicts of Summarize.
const (
	ReachAllow   = "allow"
	ReachDeny    = "deny"
	ReachPartial = "partial"
)

// Summarize is how pol treats traffic from addresses from to addresses to:
// allowed in full, denied in full, or partly (some services, or some
// addresses), with the indexes of the rules that decide it. It walks the
// rules in order until one matches every address pair and every service;
// without such a rule the default drop takes the rest. Two rules that
// together cover the pair still read as partial: the reachability check
// gives the exact answer for one flow.
func Summarize(pol *proto.Policy, from, to []string) (string, []int) {
	a, b := parseSpans(from), parseSpans(to)
	if len(a) == 0 || len(b) == 0 {
		return ReachDeny, nil
	}
	var (
		rules   []int
		actions = map[string]bool{}
		full    bool
	)
	for i, rule := range pol.Rules {
		rf, rt := parseSpans(rule.From), parseSpans(rule.To)
		if (rule.From != nil && !rf.overlaps(a)) || (rule.To != nil && !rt.overlaps(b)) {
			continue
		}
		rules = append(rules, i)
		actions[rule.Action] = true
		if (rule.From == nil || rf.covers(a)) && (rule.To == nil || rt.covers(b)) && rule.Services == nil {
			full = true
			break
		}
	}
	if !full {
		actions[proto.PolicyDeny] = true // the default drop takes the rest
	}
	switch {
	case len(actions) > 1:
		return ReachPartial, rules
	case actions[proto.PolicyAllow]:
		return ReachAllow, rules
	}
	return ReachDeny, rules
}
