// Package policy enforces a server's mesh firewall (amendment A43): it
// compiles the config's proto.Policy into a table of the agent's own,
// `table ip meshctl`, and keeps the kernel holding exactly that table.
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

// TableName is the agent's table, in the ip family.
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
	// Mesh is the mesh CIDR: whose flows a new policy judges again.
	Mesh   netip.Prefix
	Chains []Chain
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

// Range is an inclusive range of addresses (as numbers) or of ports.
type Range struct{ First, Last uint32 }

// Rule matches a packet when every field that is set does.
type Rule struct {
	// IIF and OIF are the interfaces it comes in and goes out on.
	IIF, OIF string
	// State is a connection state, StateEstablishedRelated or
	// StateInvalid.
	State string
	// Src and Dst are merged, sorted ranges; nil is any.
	Src, Dst []Range
	// Proto is tcp, udp or icmp; Ports its destination ports (tcp and udp;
	// nil is every port); Echo an ICMP echo request.
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
	// Self is the node's mesh IP: a packet to the node is received, one
	// from it sent.
	Self netip.Addr
	// Containers is the node's own container subnet, invalid when it has
	// none: traffic within the node is not the mesh firewall's.
	Containers netip.Prefix
}

func u32(a netip.Addr) uint32 { return validate.U32(a) }

func prefixRange(p netip.Prefix) Range {
	r, _ := validate.AddrEntry("prefix", p.Masked().String()) // a valid IPv4 prefix
	return Range{u32(r.First), u32(r.Last)}
}

// merged sorts rs and joins ranges that overlap or touch.
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

func contains(rs []Range, v uint32) bool {
	if rs == nil {
		return true // any
	}
	for _, r := range rs {
		if r.First <= v && v <= r.Last {
			return true
		}
	}
	return false
}

func addrs(field string, in []string) ([]Range, error) {
	var out []Range
	for i, s := range in {
		r, err := validate.AddrEntry(fmt.Sprintf("%s[%d]", field, i), s)
		if err != nil {
			return nil, err
		}
		out = append(out, Range{u32(r.First), u32(r.Last)})
	}
	return merged(out), nil
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
	if !l.Self.Is4() || !m.Contains(l.Self) {
		return nil, fmt.Errorf("mesh firewall: this node's address %s is outside the mesh %s", l.Self, m)
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
	mesh, self := []Range{prefixRange(m)}, u32(l.Self)
	var own []Range
	if l.Containers.IsValid() {
		own = []Range{prefixRange(l.Containers)}
	}
	established := Rule{State: StateEstablishedRelated, Verdict: Accept, Why: "replies and errors of accepted flows"}
	fwd := Chain{Name: chainForward, Hook: HookForward, Rules: []Rule{established}}
	in := Chain{Name: chainInput, Hook: HookInput, Rules: []Rule{{IIF: "lo", Verdict: Accept, Why: "the node itself"}, established}}
	out := Chain{Name: chainOutput, Hook: HookOutput, Rules: []Rule{{OIF: "lo", Verdict: Accept, Why: "the node itself"}, established}}
	if own != nil {
		fwd.Rules = append(fwd.Rules, Rule{Src: own, Dst: own, Verdict: Accept, Why: "within the node's containers"})
		in.Rules = append(in.Rules, Rule{Src: own, Verdict: Accept, Why: "the node's containers"})
		out.Rules = append(out.Rules, Rule{Dst: own, Verdict: Accept, Why: "the node's containers"})
	}
	fwd.Rules = append(fwd.Rules, Rule{Src: mesh, Dst: mesh, Verdict: Jump, Target: chainMeshForward, Why: "mesh traffic"})
	in.Rules = append(in.Rules, Rule{Src: mesh, Verdict: Jump, Target: chainMeshInput, Why: "mesh traffic"})
	out.Rules = append(out.Rules, Rule{Dst: mesh, Verdict: Jump, Target: chainMeshOutput, Why: "mesh traffic"})

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
		for _, s := range svcs {
			base := Rule{Proto: s.proto, Ports: s.ports, Echo: s.echo, Verdict: verdict, Why: r.Name}
			f := base
			f.Src, f.Dst = from, to
			mf.Rules = append(mf.Rules, f)
			if contains(to, self) {
				received := base
				received.Src = from
				mi.Rules = append(mi.Rules, received)
			}
			if contains(from, self) {
				sent := base
				sent.Dst = to
				mo.Rules = append(mo.Rules, sent)
			}
		}
	}
	rest := Rule{Verdict: Drop, Why: "no rule matched"}
	mf.Rules, mi.Rules, mo.Rules = append(mf.Rules, rest), append(mi.Rules, rest), append(mo.Rules, rest)
	return &Ruleset{Mesh: m, Chains: []Chain{fwd, in, out, mf, mi, mo}, Next: next}, nil
}

// Count is the number of rules.
func (rs *Ruleset) Count() int {
	n := 0
	for _, c := range rs.Chains {
		n += len(c.Rules)
	}
	return n
}

func addrText(rs []Range) string {
	var parts []string
	for _, r := range rs {
		parts = append(parts, validate.AddrRange{First: validate.FromU32(r.First), Last: validate.FromU32(r.Last)}.String())
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
	if r.Src != nil {
		add("ip saddr", addrText(r.Src))
	}
	if r.Dst != nil {
		add("ip daddr", addrText(r.Dst))
	}
	if r.Proto != "" {
		add("meta l4proto", r.Proto)
	}
	if r.Ports != nil {
		add("th dport", portText(r.Ports))
	}
	if r.Echo {
		add("icmp type echo-request")
	}
	add(r.Verdict)
	if r.Verdict == Jump {
		add(r.Target)
	}
	return strings.Join(parts, " ")
}

// Text is the table in nft's own spelling, each rule with where it comes
// from: what the golden pins and `nft list table ip meshctl` resembles.
func (rs *Ruleset) Text() string {
	var b strings.Builder
	fmt.Fprintf(&b, "table ip %s\n", TableName)
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
		n, err := e.K.Forget(rs.Mesh)
		switch {
		case err != nil:
			log.Warn("mesh firewall: open mesh connections could not be re-checked; they keep going until they end", "err", err)
		case n > 0:
			log.Info("mesh firewall: open mesh connections are judged again by the new policy", "flows", n)
		}
	}
	return 1, nil
}
