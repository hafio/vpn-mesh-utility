package proto

import (
	"errors"
	"fmt"
	"net/netip"
	"strings"

	"github.com/hafio/vpn-mesh-utility/internal/validate"
)

// Policy is the mesh firewall every server enforces (amendment A43). Every packet between two mesh addresses that crosses the
// server -- relayed, to it, or from it -- is checked against Rules in order:
// the first rule that matches decides, and a packet no rule matches is
// dropped. Replies and ICMP errors of a flow a rule allowed always pass.
// Traffic with an end outside Mesh (an exit's internet traffic) is not mesh
// traffic and is never filtered here.
type Policy struct {
	// Mesh is the mesh CIDR.
	Mesh  string       `json:"mesh"`
	Rules []PolicyRule `json:"rules"`
}

// Policy actions.
const (
	PolicyAllow = "allow"
	PolicyDeny  = "deny"
)

// PolicyRule matches packets from an address in From to one in To for one of
// Services. An empty From, To or Services is any; the planner never emits an
// empty list for a reference that resolved to nothing (it drops the rule).
type PolicyRule struct {
	// Name says where the rule comes from ("system: ping", "policy: web",
	// "zone: office"). Display only; it never reaches the kernel.
	Name string `json:"name"`
	// From and To are addresses, networks and ranges in validate.AddrEntry's
	// canonical form, sorted and merged.
	From     []string  `json:"from,omitempty"`
	To       []string  `json:"to,omitempty"`
	Services []Service `json:"services,omitempty"`
	Action   string    `json:"action"`
}

// Service is a protocol (validate.ServiceTCP, ServiceUDP or ServiceICMP)
// and its ports, "22" or "1000-2000", sorted and merged; none is every port
// (icmp never has any).
type Service struct {
	Proto string   `json:"proto"`
	Ports []string `json:"ports,omitempty"`
}

// Verdicts of Matcher.Decide that are not a rule's index.
const (
	// RuleDefault is the drop at the end: no rule matched.
	RuleDefault = -1
	// RuleNotMesh is traffic with an end outside the mesh: not filtered.
	RuleNotMesh = -2
)

// Limits on a policy, so a config cannot make the agent build an unbounded
// ruleset.
const (
	MaxPolicyRules   = 512
	MaxPolicyEntries = 1024
)

// Matcher is a parsed Policy.
type Matcher struct {
	mesh  netip.Prefix
	rules []matchRule
}

type matchRule struct {
	from, to []validate.AddrRange // nil is any
	services []validate.Service   // nil is any
	allow    bool
}

// Matcher parses the policy, refusing any value outside the grammar.
func (p Policy) Matcher() (*Matcher, error) {
	var errs []error
	mesh, err := validate.Prefix4("policy.mesh", p.Mesh)
	if err != nil {
		errs = append(errs, err)
	}
	if len(p.Rules) > MaxPolicyRules {
		errs = append(errs, fmt.Errorf("policy.rules: %d rules, at most %d", len(p.Rules), MaxPolicyRules))
	}
	m := &Matcher{mesh: mesh}
	for i, r := range p.Rules {
		field := fmt.Sprintf("policy.rules[%d]", i)
		if _, err := validate.Text(field+".name", r.Name, 200); err != nil {
			errs = append(errs, err)
		}
		mr := matchRule{allow: r.Action == PolicyAllow}
		if _, err := validate.Enum(field+".action", r.Action, PolicyAllow, PolicyDeny); err != nil {
			errs = append(errs, err)
		}
		var ferr, terr error
		mr.from, ferr = parseAddrs(field+".from", r.From, mesh)
		mr.to, terr = parseAddrs(field+".to", r.To, mesh)
		errs = append(errs, ferr, terr)
		for j, s := range r.Services {
			sf := fmt.Sprintf("%s.services[%d]", field, j)
			if s.Proto == validate.ServiceICMP || len(s.Ports) == 0 {
				svc, err := validate.ServiceEntry(sf, s.Proto)
				if err == nil && len(s.Ports) > 0 {
					err = fmt.Errorf("%s: icmp has no ports", sf)
				}
				errs = append(errs, err)
				mr.services = append(mr.services, svc)
				continue
			}
			for _, port := range s.Ports {
				svc, err := validate.ServiceEntry(sf, s.Proto+"/"+port)
				errs = append(errs, err)
				mr.services = append(mr.services, svc)
			}
		}
		m.rules = append(m.rules, mr)
	}
	if err := errors.Join(errs...); err != nil {
		return nil, err
	}
	return m, nil
}

func parseAddrs(field string, in []string, mesh netip.Prefix) ([]validate.AddrRange, error) {
	if len(in) > MaxPolicyEntries {
		return nil, fmt.Errorf("%s: %d entries, at most %d", field, len(in), MaxPolicyEntries)
	}
	var (
		out  []validate.AddrRange
		errs []error
	)
	for i, s := range in {
		r, err := validate.AddrEntry(fmt.Sprintf("%s[%d]", field, i), s)
		if err == nil && mesh.IsValid() && (!mesh.Contains(r.First) || !mesh.Contains(r.Last)) {
			err = fmt.Errorf("%s[%d]: %s is outside the mesh %s", field, i, s, mesh)
		}
		errs = append(errs, err)
		out = append(out, r)
	}
	return out, errors.Join(errs...)
}

// Decide is the verdict for one packet: the index of the first rule that
// matches it, or RuleDefault (dropped), or RuleNotMesh (not mesh traffic,
// so not filtered). proto is validate.ServiceTCP, ServiceUDP or ServiceICMP;
// port is ignored for icmp.
func (m *Matcher) Decide(src, dst netip.Addr, proto string, port int) (allow bool, rule int) {
	if !m.mesh.Contains(src) || !m.mesh.Contains(dst) {
		return true, RuleNotMesh
	}
	for i, r := range m.rules {
		if inRanges(r.from, src) && inRanges(r.to, dst) && serves(r.services, proto, port) {
			return r.allow, i
		}
	}
	return false, RuleDefault
}

func inRanges(rs []validate.AddrRange, a netip.Addr) bool {
	if rs == nil {
		return true
	}
	for _, r := range rs {
		if !a.Less(r.First) && !r.Last.Less(a) {
			return true
		}
	}
	return false
}

func serves(ss []validate.Service, proto string, port int) bool {
	if ss == nil {
		return true
	}
	for _, s := range ss {
		if s.Proto != proto {
			continue
		}
		if proto == validate.ServiceICMP || s.First == 0 || (port >= s.First && port <= s.Last) {
			return true
		}
	}
	return false
}

// String is a rule in one line, for the agent's status and the web UI.
func (r PolicyRule) String() string {
	side := func(as []string) string {
		if len(as) == 0 {
			return "any"
		}
		return strings.Join(as, ",")
	}
	svc := "any"
	if len(r.Services) > 0 {
		var parts []string
		for _, s := range r.Services {
			if len(s.Ports) == 0 {
				parts = append(parts, s.Proto)
				continue
			}
			parts = append(parts, s.Proto+"/"+strings.Join(s.Ports, ","))
		}
		svc = strings.Join(parts, " ")
	}
	return fmt.Sprintf("%s %s -> %s %s (%s)", r.Action, side(r.From), side(r.To), svc, r.Name)
}
