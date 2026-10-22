// Package policy enforces a server's mesh firewall (amendment A43): it
// compiles the config's proto.Policy into a table of the agent's own,
// `table inet meshctl`, and keeps the kernel holding exactly that table.
// The table filters both families (M8): every rule that matches addresses
// or ICMP is of one, and a policy rule becomes one per family it names.
//
// The compile is pure (Compile, Ruleset.Text, pinned by a golden); the
// kernel is behind Kernel -- nftables and conntrack over netlink on Linux.
// No operator string reaches the kernel: the table carries addresses,
// ports and verdicts only, and one comment, the ruleset's fingerprint, by
// which Sync sees whether the table in the kernel is still the one it
// installed.
//
// Every packet between two mesh addresses that crosses the server is
// mesh traffic: forwarded (it relays between nodes, or to its containers),
// received (to the server, whatever the address it is sent to) or sent.
// Each kind has its chain of the policy's rules, in order: a forwarded
// packet is matched on both addresses, a received one on its source (the
// server is its destination), a sent one on its destination. Replies and
// ICMP errors of an accepted flow pass on connection tracking; traffic
// within the node -- loopback, its own containers -- and traffic with an
// end outside the mesh is not the mesh firewall's.
package policy

import (
	"cmp"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"net/netip"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/hafio/vpn-mesh-utility/internal/proto"
	"github.com/hafio/vpn-mesh-utility/internal/validate"
)

// TableName is the agent's table, in the inet family (both IPv4 and IPv6).
const TableName = "meshctl"

// The base chains' hooks.
const (
	HookForward = "forward"
	HookInput   = "input"
	HookOutput  = "output"
)

// Verdicts.
const (
	Accept = "accept"
	Drop   = "drop"
	Jump   = "jump"
)

// Connection states a rule matches.
const (
	StateEstablishedRelated = "established,related"
	StateInvalid            = "invalid"
)

// Ruleset is the whole table.
type Ruleset struct {
	// Mesh and Mesh6 are the mesh's prefixes (Mesh6 invalid without IPv6):
	// whose flows a new policy judges again.
	Mesh, Mesh6 netip.Prefix
	Chains      []Chain
	// Next is when a scheduled rule starts or ends, so the policy must be
	// compiled again; zero when none has a schedule. Not part of Text.
	Next time.Time
}

// Chain is a base chain (Hook set; it accepts what no rule decides) or one
// only jumped to.
type Chain struct {
	Name  string
	Hook  string
	Rules []Rule
}

// Range is an inclusive range of ports.
type Range struct{ First, Last uint32 }

// Address families of a rule: one of them, or Both for a rule that matches
// neither addresses nor ICMP.
const (
	Both = 0
	IPv4 = 4
	IPv6 = 6
)

// ICMPv6 is the protocol of IPv6's ping, as nft names it.
const ICMPv6 = "ipv6-icmp"

// Rule matches a packet when every field that is set does.
type Rule struct {
	// IIF and OIF are the interfaces it comes in and goes out on.
	IIF, OIF string
	// State is a connection state, StateEstablishedRelated or
	// StateInvalid.
	State string
	// Family is IPv4 or IPv6, or Both.
	Family int
	// Src and Dst are merged, sorted ranges of Family; nil is any.
	Src, Dst []validate.AddrRange
	// Proto is tcp, udp, icmp or ICMPv6; Ports its destination ports (tcp
	// and udp; nil is every port); Echo an echo request (ping).
	Proto string
	Ports []Range
	Echo  bool
	// Verdict is Accept, Drop or Jump to Target.
	Verdict string
	Target  string
	// Why says where the rule comes from, in Text only.
	Why string
}

// Local is what the agent knows of its own node.
type Local struct {
	// Self and Self6 are the node's mesh IPs, either of them invalid when
	// it has none (M8): a packet to the node is received, one from it sent.
	Self, Self6 netip.Addr
	// Containers and Containers6 are the node's own container subnets,
	// invalid when it has none: traffic within the node is not the mesh
	// firewall's.
	Containers, Containers6 netip.Prefix
}

// merged sorts port ranges and joins those that overlap or touch.
func merged(rs []Range) []Range {
	if len(rs) == 0 {
		return nil
	}
	in := slices.Clone(rs)
	slices.SortFunc(in, func(a, b Range) int { return cmp.Compare(a.First, b.First) })
	out := []Range{in[0]}
	for _, r := range in[1:] {
		last := &out[len(out)-1]
		if uint64(r.First) <= uint64(last.Last)+1 {
			last.Last = max(last.Last, r.Last)
			continue
		}
		out = append(out, r)
	}
	return out
}

// contains reports whether a is in rs; nil is any.
func contains(rs []validate.AddrRange, a netip.Addr) bool {
	return rs == nil || validate.RangesContain(rs, a)
}

func addrs(field string, in []string) ([]validate.AddrRange, error) {
	var out []validate.AddrRange
	for i, s := range in {
		r, err := validate.AddrEntry(fmt.Sprintf("%s[%d]", field, i), s)
		if err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return validate.MergeRanges(out), nil
}

// ofFamily is the part of a rule side in family f: any stays any, and ok
// is false when the side names addresses but none of f.
func ofFamily(rs []validate.AddrRange, f int) ([]validate.AddrRange, bool) {
	if rs == nil {
		return nil, true
	}
	var out []validate.AddrRange
	for _, r := range rs {
		if r.First.Is4() == (f == IPv4) {
			out = append(out, r)
		}
	}
	return out, out != nil
}

// family is one address family the table filters, with the node's own
// addresses in it.
type family struct {
	n    int
	mesh netip.Prefix
	self netip.Addr
	own  netip.Prefix
}

// service is one protocol of a rule's services.
type service struct {
	proto string
	ports []Range
	echo  bool
}

func services(field string, in []proto.Service) ([]service, error) {
	if len(in) == 0 {
		return []service{{}}, nil // any traffic
	}
	var out []service
	for i, s := range in {
		f := fmt.Sprintf("%s[%d]", field, i)
		switch {
		case s.Proto == validate.ServiceICMP && len(s.Ports) == 0:
			out = append(out, service{proto: validate.ServiceICMP, echo: true})
		case (s.Proto == validate.ServiceTCP || s.Proto == validate.ServiceUDP) && len(s.Ports) == 0:
			out = append(out, service{proto: s.Proto})
		case s.Proto == validate.ServiceTCP || s.Proto == validate.ServiceUDP:
			svc := service{proto: s.Proto}
			for _, p := range s.Ports {
				e, err := validate.ServiceEntry(f, s.Proto+"/"+p)
				if err != nil {
					return nil, err
				}
				svc.ports = append(svc.ports, Range{uint32(e.First), uint32(e.Last)})
			}
			svc.ports = merged(svc.ports)
			out = append(out, svc)
		default:
			return nil, fmt.Errorf("%s: protocol %q with ports %v is not a service the mesh firewall knows", f, s.Proto, s.Ports)
		}
	}
	return out, nil
}

// Chain names.
const (
	chainForward     = "forward"
	chainInput       = "input"
	chainOutput      = "output"
	chainMeshForward = "mesh_forward"
	chainMeshInput   = "mesh_input"
	chainMeshOutput  = "mesh_output"
)

// Compile turns a server's policy into its table at the time at: the rules
// whose schedule does not hold then are left out. It refuses a policy that
// does not parse, as the config's validation did already.
func Compile(p proto.Policy, l Local, at time.Time) (*Ruleset, error) {
	m, err := validate.Prefix4("policy.mesh", p.Mesh)
	if err != nil {
		return nil, err
	}
	fams := []family{{n: IPv4, mesh: m, self: l.Self, own: l.Containers}}
	if p.Mesh6 != "" {
		m6, err := validate.Prefix6("policy.mesh6", p.Mesh6)
		if err != nil {
			return nil, err
		}
		fams = append(fams, family{n: IPv6, mesh: m6, self: l.Self6, own: l.Containers6})
	}
	if !l.Self.IsValid() && !l.Self6.IsValid() {
		return nil, errors.New("mesh firewall: this node has no mesh address")
	}
	for _, f := range fams {
		if f.self.IsValid() && (f.self.Is4() != (f.n == IPv4) || !f.mesh.Contains(f.self)) {
			return nil, fmt.Errorf("mesh firewall: this node's address %s is outside the mesh %s", f.self, f.mesh)
		}
	}
	if l.Self6.IsValid() && len(fams) == 1 {
		return nil, fmt.Errorf("mesh firewall: this node has the IPv6 address %s, but the policy has no IPv6 prefix", l.Self6)
	}
	// Only a policy with schedules needs its clock read; proto.Matcher
	// parses those.
	active, next := make([]bool, len(p.Rules)), time.Time{}
	for i := range active {
		active[i] = true
	}
	if slices.ContainsFunc(p.Rules, func(r proto.PolicyRule) bool { return r.Schedule != nil }) {
		matcher, err := p.Matcher()
		if err != nil {
			return nil, err
		}
		active, next = matcher.Active(at), matcher.Next(at)
	}
	established := Rule{State: StateEstablishedRelated, Verdict: Accept, Why: "replies and errors of accepted flows"}
	fwd := Chain{Name: chainForward, Hook: HookForward, Rules: []Rule{established}}
	in := Chain{Name: chainInput, Hook: HookInput, Rules: []Rule{{IIF: "lo", Verdict: Accept, Why: "the node itself"}, established}}
	out := Chain{Name: chainOutput, Hook: HookOutput, Rules: []Rule{{OIF: "lo", Verdict: Accept, Why: "the node itself"}, established}}
	for _, f := range fams {
		mesh := []validate.AddrRange{validate.PrefixRange(f.mesh)}
		if f.own.IsValid() {
			own := []validate.AddrRange{validate.PrefixRange(f.own)}
			fwd.Rules = append(fwd.Rules, Rule{Family: f.n, Src: own, Dst: own, Verdict: Accept, Why: "within the node's containers"})
			in.Rules = append(in.Rules, Rule{Family: f.n, Src: own, Verdict: Accept, Why: "the node's containers"})
			out.Rules = append(out.Rules, Rule{Family: f.n, Dst: own, Verdict: Accept, Why: "the node's containers"})
		}
		fwd.Rules = append(fwd.Rules, Rule{Family: f.n, Src: mesh, Dst: mesh, Verdict: Jump, Target: chainMeshForward, Why: "mesh traffic"})
		in.Rules = append(in.Rules, Rule{Family: f.n, Src: mesh, Verdict: Jump, Target: chainMeshInput, Why: "mesh traffic"})
		out.Rules = append(out.Rules, Rule{Family: f.n, Dst: mesh, Verdict: Jump, Target: chainMeshOutput, Why: "mesh traffic"})
	}

	invalid := Rule{State: StateInvalid, Verdict: Drop, Why: "not part of any flow"}
	mf := Chain{Name: chainMeshForward, Rules: []Rule{invalid}}
	mi := Chain{Name: chainMeshInput, Rules: []Rule{invalid}}
	mo := Chain{Name: chainMeshOutput, Rules: []Rule{invalid}}
	for i, r := range p.Rules {
		if !active[i] {
			continue
		}
		field := fmt.Sprintf("policy.rules[%d]", i)
		from, err := addrs(field+".from", r.From)
		if err != nil {
			return nil, err
		}
		to, err := addrs(field+".to", r.To)
		if err != nil {
			return nil, err
		}
		svcs, err := services(field+".services", r.Services)
		if err != nil {
			return nil, err
		}
		verdict := Accept
		switch r.Action {
		case proto.PolicyAllow:
		case proto.PolicyDeny:
			verdict = Drop
		default:
			return nil, fmt.Errorf("%s.action: %q is neither allow nor deny", field, r.Action)
		}
		for _, f := range fams {
			fromF, okFrom := ofFamily(from, f.n)
			toF, okTo := ofFamily(to, f.n)
			if !okFrom || !okTo {
				continue // the rule names no address of this family
			}
			for _, s := range svcs {
				base := Rule{Family: f.n, Proto: s.proto, Ports: s.ports, Echo: s.echo, Verdict: verdict, Why: r.Name}
				if s.proto == validate.ServiceICMP && f.n == IPv6 {
					base.Proto = ICMPv6
				}
				fr := base
				fr.Src, fr.Dst = fromF, toF
				mf.Rules = append(mf.Rules, fr)
				if f.self.IsValid() && contains(toF, f.self) {
					received := base
					received.Src = fromF
					mi.Rules = append(mi.Rules, received)
				}
				if f.self.IsValid() && contains(fromF, f.self) {
					sent := base
					sent.Dst = toF
					mo.Rules = append(mo.Rules, sent)
				}
			}
		}
	}
	rest := Rule{Verdict: Drop, Why: "no rule matched"}
	mf.Rules, mi.Rules, mo.Rules = append(mf.Rules, rest), append(mi.Rules, rest), append(mo.Rules, rest)
	rs := &Ruleset{Mesh: m, Chains: []Chain{fwd, in, out, mf, mi, mo}, Next: next}
	if len(fams) > 1 {
		rs.Mesh6 = fams[1].mesh
	}
	return rs, nil
}

// Count is the number of rules.
func (rs *Ruleset) Count() int {
	n := 0
	for _, c := range rs.Chains {
		n += len(c.Rules)
	}
	return n
}

func addrText(rs []validate.AddrRange) string {
	var parts []string
	for _, r := range rs {
		parts = append(parts, r.String())
	}
	return "{ " + strings.Join(parts, ", ") + " }"
}

func portText(rs []Range) string {
	var parts []string
	for _, r := range rs {
		if r.First == r.Last {
			parts = append(parts, strconv.FormatUint(uint64(r.First), 10))
			continue
		}
		parts = append(parts, fmt.Sprintf("%d-%d", r.First, r.Last))
	}
	return "{ " + strings.Join(parts, ", ") + " }"
}

// String is the rule in nft's own spelling.
func (r Rule) String() string {
	var parts []string
	add := func(s ...string) { parts = append(parts, s...) }
	if r.IIF != "" {
		add("iifname", strconv.Quote(r.IIF))
	}
	if r.OIF != "" {
		add("oifname", strconv.Quote(r.OIF))
	}
	if r.State != "" {
		add("ct state", r.State)
	}
	ip := "ip"
	switch r.Family {
	case IPv4:
		add("meta nfproto ipv4")
	case IPv6:
		add("meta nfproto ipv6")
		ip = "ip6"
	}
	if r.Src != nil {
		add(ip+" saddr", addrText(r.Src))
	}
	if r.Dst != nil {
		add(ip+" daddr", addrText(r.Dst))
	}
	if r.Proto != "" {
		add("meta l4proto", r.Proto)
	}
	if r.Ports != nil {
		add("th dport", portText(r.Ports))
	}
	switch {
	case r.Echo && r.Proto == ICMPv6:
		add("icmpv6 type echo-request")
	case r.Echo:
		add("icmp type echo-request")
	}
	add(r.Verdict)
	if r.Verdict == Jump {
		add(r.Target)
	}
	return strings.Join(parts, " ")
}

// Text is the table in nft's own spelling, each rule with where it comes
// from: what the golden pins and `nft list table inet meshctl` resembles.
func (rs *Ruleset) Text() string {
	var b strings.Builder
	fmt.Fprintf(&b, "table inet %s\n", TableName)
	for _, c := range rs.Chains {
		if c.Hook != "" {
			fmt.Fprintf(&b, "chain %s hook %s priority filter policy accept\n", c.Name, c.Hook)
		} else {
			fmt.Fprintf(&b, "chain %s\n", c.Name)
		}
		for _, r := range c.Rules {
			fmt.Fprintf(&b, "  %s  # %s\n", r, r.Why)
		}
	}
	return b.String()
}

// Fingerprint names the ruleset: the start of the sha256 of its Text.
func (rs *Ruleset) Fingerprint() string {
	sum := sha256.Sum256([]byte(rs.Text()))
	return hex.EncodeToString(sum[:8])
}

// Kernel holds the table: nftables and conntrack over netlink on Linux
// (NewKernel).
type Kernel interface {
	// Installed is the fingerprint the table in the kernel carries, "" when
	// there is no table, and something no fingerprint equals when it has
	// none of ours.
	Installed() (string, error)
	// Replace makes rs the whole table, marked with fingerprint, in one
	// transaction: the kernel holds the old table or the new one.
	Replace(rs *Ruleset, fingerprint string) error
	// Remove deletes the table.
	Remove() error
	// Forget deletes the connection tracking of the flows between two
	// addresses of mesh, so their next packet is judged again.
	Forget(mesh netip.Prefix) (int, error)
}

// Enforcer keeps the kernel holding the ruleset it is given.
type Enforcer struct {
	K   Kernel
	Log *slog.Logger
}

// ErrNoKernel is the refusal of a platform without nftables.
var ErrNoKernel = errors.New("the mesh firewall needs Linux nftables, and servers run on Linux")

// Sync makes the kernel hold rs -- or no table, for nil -- and reports
// what it changed. A table already holding rs is left alone; any other is
// replaced whole, and the mesh flows the old one let through are judged
// again by the new one.
func (e *Enforcer) Sync(rs *Ruleset) (int, error) {
	have, err := e.K.Installed()
	if err != nil {
		return 0, fmt.Errorf("mesh firewall: read the table %s: %w", TableName, err)
	}
	if rs == nil {
		if have == "" {
			return 0, nil
		}
		if err := e.K.Remove(); err != nil {
			return 0, fmt.Errorf("mesh firewall: remove the table %s: %w", TableName, err)
		}
		return 1, nil
	}
	want := rs.Fingerprint()
	if have == want {
		return 0, nil
	}
	if err := e.K.Replace(rs, want); err != nil {
		return 0, fmt.Errorf("mesh firewall: install the table %s (%d rules): %w", TableName, rs.Count(), err)
	}
	if have != "" {
		log := e.Log
		if log == nil {
			log = slog.New(slog.DiscardHandler)
		}
		for _, mesh := range []netip.Prefix{rs.Mesh, rs.Mesh6} {
			if !mesh.IsValid() {
				continue
			}
			n, err := e.K.Forget(mesh)
			switch {
			case err != nil:
				log.Warn("mesh firewall: open mesh connections could not be re-checked; they keep going until they end", "mesh", mesh, "err", err)
			case n > 0:
				log.Info("mesh firewall: open mesh connections are judged again by the new policy", "mesh", mesh, "flows", n)
			}
		}
	}
	return 1, nil
}
