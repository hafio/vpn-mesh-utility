package proto

import (
	"errors"
	"fmt"
	"net/netip"
	"strings"
	"time"

	"github.com/hafio/vpn-mesh-utility/internal/validate"
)

// Policy is the mesh firewall every server enforces (amendment A43). Every packet between two mesh addresses that crosses the
// server -- relayed, to it, or from it -- is checked against Rules in order:
// the first rule that matches decides, and a packet no rule matches is
// dropped. Replies and ICMP errors of a flow a rule allowed always pass.
// Traffic with an end outside Mesh (an exit's internet traffic) is not mesh
// traffic and is never filtered here. A rule with a Schedule takes part only
// while its window holds, on the wall clock of TZ.
type Policy struct {
	// Mesh is the mesh CIDR.
	Mesh  string       `json:"mesh"`
	Rules []PolicyRule `json:"rules"`
	// TZ is the IANA time zone the schedules are read in; set when a rule
	// has one.
	TZ string `json:"tz,omitempty"`
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
	// Schedule limits the rule to times of the week; nil is always.
	Schedule *Schedule `json:"schedule,omitempty"`
}

// Schedule is when a rule applies (validate.Schedule): on Days (every day
// when empty), from Start to End, "HH:MM", an End before Start ending the
// next day; without Start and End, the whole of each day. A window belongs
// to the day it starts: fri 22:00-06:00 holds until Saturday 06:00.
type Schedule struct {
	Days  []string `json:"days,omitempty"`
	Start string   `json:"start,omitempty"`
	End   string   `json:"end,omitempty"`
}

// window parses s.
func (s Schedule) window(field string) (validate.Window, error) {
	return validate.Schedule(field, s.Days, s.Start, s.End)
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
	loc   *time.Location // where the schedules are read; nil without TZ
	rules []matchRule
}

type matchRule struct {
	from, to []validate.AddrRange // nil is any
	services []validate.Service   // nil is any
	when     *validate.Window     // nil is always
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
	if p.TZ != "" {
		m.loc, err = validate.TimeZone("policy.tz", p.TZ)
		errs = append(errs, err)
	}
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
		if r.Schedule != nil {
			w, err := r.Schedule.window(field + ".schedule")
			if err == nil && p.TZ == "" {
				err = fmt.Errorf("%s.schedule: the policy has no time zone (tz) to read it in", field)
			}
			errs = append(errs, err)
			mr.when = &w
		}
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

// Decide is the verdict for one packet sent at at: the index of the first
// rule that applies then and matches it, or RuleDefault (dropped), or
// RuleNotMesh (not mesh traffic, so not filtered). proto is
// validate.ServiceTCP, ServiceUDP or ServiceICMP; port is ignored for icmp.
func (m *Matcher) Decide(src, dst netip.Addr, proto string, port int, at time.Time) (allow bool, rule int) {
	if !m.mesh.Contains(src) || !m.mesh.Contains(dst) {
		return true, RuleNotMesh
	}
	for i, r := range m.rules {
		if m.applies(r, at) && inRanges(r.from, src) && inRanges(r.to, dst) && serves(r.services, proto, port) {
			return r.allow, i
		}
	}
	return false, RuleDefault
}

func (m *Matcher) applies(r matchRule, at time.Time) bool {
	return r.when == nil || r.when.Holds(at.In(m.loc))
}

// Active says which rules apply at at: every rule without a schedule, and
// those whose window holds on the wall clock of the policy's zone.
func (m *Matcher) Active(at time.Time) []bool {
	out := make([]bool, len(m.rules))
	for i, r := range m.rules {
		out[i] = m.applies(r, at)
	}
	return out
}

// Next is the first time after at when a schedule may start or end, so the
// rules that apply must be worked out again; the zero time when no rule has
// a schedule. Every edge is a time of day, so one comes within a day or
// two, daylight-saving days included.
func (m *Matcher) Next(at time.Time) time.Time {
	var edges [24 * 60]bool
	scheduled := false
	for _, r := range m.rules {
		if r.when == nil {
			continue
		}
		scheduled = true
		for _, e := range r.when.Edges() {
			edges[e] = true
		}
	}
	if !scheduled {
		return time.Time{}
	}
	y, mo, d := at.In(m.loc).Date()
	for day := 0; day <= 7; day++ {
		for e, ok := range edges {
			if !ok {
				continue
			}
			if t := time.Date(y, mo, d+day, e/60, e%60, 0, 0, m.loc); t.After(at) {
				return t
			}
		}
	}
	return time.Time{} // unreachable: every edge recurs daily
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
	when := ""
	if r.Schedule != nil {
		if w, err := r.Schedule.window("schedule"); err == nil {
			when = " " + w.String()
		}
	}
	return fmt.Sprintf("%s %s -> %s %s%s (%s)", r.Action, side(r.From), side(r.To), svc, when, r.Name)
}
