// Package agentfake holds in-memory fakes of the agent's host interfaces for
// tests: a WireGuard host, a routing table, policy routing, sysctls, the
// firewall and DNS steps. Each applies changes the way the host would and
// counts every mutation, so a test can assert that applying the same config
// twice changes nothing. Production code never imports it
// (TestModuleBoundaries).
package agentfake

import (
	"context"
	"errors"
	"fmt"
	"net/netip"
	"reflect"
	"slices"
	"strings"
	"sync"

	"github.com/hafio/vpn-mesh-utility/internal/agent/dns"
	"github.com/hafio/vpn-mesh-utility/internal/agent/exitmode"
	"github.com/hafio/vpn-mesh-utility/internal/agent/wg"
	"github.com/hafio/vpn-mesh-utility/internal/proto"
)

type link struct {
	ours  bool
	mtu   int
	up    bool
	addrs []netip.Prefix
	dev   wg.Device
}

// Host is a fake wg.Host.
type Host struct {
	mu    sync.Mutex
	links map[string]*link
	// Calls lists the mutating calls in order ("CreateLink wg0").
	Calls []string
	// Stats is what Device reports as peer statistics.
	Stats []wg.PeerStats
	// Fail makes the named method fail.
	Fail string
}

// NewHost returns a host with no interfaces.
func NewHost() *Host { return &Host{links: map[string]*link{}} }

// AddForeignLink creates an interface the agent did not create.
func (h *Host) AddForeignLink(name string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.links[name] = &link{mtu: 1420}
}

// RemoveLink deletes an interface, as a reboot or `ip link del` would.
func (h *Host) RemoveLink(name string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.links, name)
}

// SetFail makes the named method fail from now on ("" for none), safely
// while an agent is running on the host.
func (h *Host) SetFail(method string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.Fail = method
}

// Mutations is how many mutating calls were made.
func (h *Host) Mutations() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return len(h.Calls)
}

func (h *Host) call(method, name string) (*link, error) {
	if h.Fail == method {
		return nil, fmt.Errorf("fake %s failed", method)
	}
	l := h.links[name]
	if l == nil && method != "Link" && method != "CreateLink" {
		return nil, fmt.Errorf("%s: no interface %s", method, name)
	}
	return l, nil
}

func (h *Host) Link(name string) (wg.Link, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	l, err := h.call("Link", name)
	if err != nil || l == nil {
		return wg.Link{}, err
	}
	return wg.Link{Exists: true, Ours: l.ours, MTU: l.mtu, Up: l.up, Addrs: slices.Clone(l.addrs)}, nil
}

func (h *Host) CreateLink(name string) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, err := h.call("CreateLink", name); err != nil {
		return err
	}
	if h.links[name] != nil {
		return errors.New("file exists")
	}
	h.links[name] = &link{ours: true, mtu: 1420}
	h.Calls = append(h.Calls, "CreateLink "+name)
	return nil
}

func (h *Host) SetMTU(name string, mtu int) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	l, err := h.call("SetMTU", name)
	if err != nil {
		return err
	}
	l.mtu = mtu
	h.Calls = append(h.Calls, fmt.Sprintf("SetMTU %s %d", name, mtu))
	return nil
}

func (h *Host) SetUp(name string) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	l, err := h.call("SetUp", name)
	if err != nil {
		return err
	}
	l.up = true
	h.Calls = append(h.Calls, "SetUp "+name)
	return nil
}

func (h *Host) AddAddr(name string, p netip.Prefix) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	l, err := h.call("AddAddr", name)
	if err != nil {
		return err
	}
	l.addrs = append(l.addrs, p)
	h.Calls = append(h.Calls, fmt.Sprintf("AddAddr %s %s", name, p))
	return nil
}

func (h *Host) DelAddr(name string, p netip.Prefix) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	l, err := h.call("DelAddr", name)
	if err != nil {
		return err
	}
	l.addrs = slices.DeleteFunc(l.addrs, func(a netip.Prefix) bool { return a == p })
	h.Calls = append(h.Calls, fmt.Sprintf("DelAddr %s %s", name, p))
	return nil
}

func (h *Host) Device(name string) (wg.Device, []wg.PeerStats, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	l, err := h.call("Device", name)
	if err != nil {
		return wg.Device{}, nil, err
	}
	d := l.dev
	d.Peers = nil
	for _, p := range l.dev.Peers {
		p.AllowedIPs = slices.Clone(p.AllowedIPs)
		d.Peers = append(d.Peers, p)
	}
	return d, slices.Clone(h.Stats), nil
}

// Configure applies a change with the kernel's semantics: ReplacePeers off,
// per-peer replace or append of AllowedIPs, explicit removal.
func (h *Host) Configure(name string, c wg.Change) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	l, err := h.call("Configure", name)
	if err != nil {
		return err
	}
	d := &l.dev
	if c.PrivateKey != nil {
		d.PrivateKey = *c.PrivateKey
	}
	if c.ListenPort != nil {
		d.ListenPort = *c.ListenPort
	}
	if c.FwMark != nil {
		d.FwMark = *c.FwMark
	}
	for _, pc := range c.Peers {
		i := slices.IndexFunc(d.Peers, func(p wg.Peer) bool { return p.PublicKey == pc.PublicKey })
		if pc.Remove {
			if i >= 0 {
				d.Peers = slices.Delete(d.Peers, i, i+1)
			}
			continue
		}
		if i < 0 {
			d.Peers = append(d.Peers, wg.Peer{PublicKey: pc.PublicKey})
			i = len(d.Peers) - 1
		}
		p := &d.Peers[i]
		if pc.Endpoint != nil {
			p.Endpoint = *pc.Endpoint
		}
		if pc.Keepalive != nil {
			p.Keepalive = *pc.Keepalive
		}
		if pc.ReplaceAllowedIPs {
			p.AllowedIPs = nil
		}
		p.AllowedIPs = append(p.AllowedIPs, pc.AllowedIPs...)
		wg.SortPrefixes(p.AllowedIPs)
	}
	h.Calls = append(h.Calls, fmt.Sprintf("Configure %s (%d changes)", name, c.Count()))
	return nil
}

func (h *Host) Close() error { return nil }

// DeviceOf returns the configured device of an interface, for assertions.
func (h *Host) DeviceOf(name string) (wg.Device, bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	l := h.links[name]
	if l == nil {
		return wg.Device{}, false
	}
	return l.dev, true
}

// Routes is a fake routes.Table.
type Routes struct {
	mu     sync.Mutex
	routes map[string][]netip.Prefix
	// Calls lists the mutating calls in order.
	Calls []string
	Fail  error
}

// NewRoutes returns an empty table.
func NewRoutes() *Routes { return &Routes{routes: map[string][]netip.Prefix{}} }

func (r *Routes) List(link string) ([]netip.Prefix, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.Fail != nil {
		return nil, r.Fail
	}
	return slices.Clone(r.routes[link]), nil
}

func (r *Routes) Add(link string, p netip.Prefix) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.routes[link] = append(r.routes[link], p)
	r.Calls = append(r.Calls, "add "+p.String())
	return nil
}

func (r *Routes) Del(link string, p netip.Prefix) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.routes[link] = slices.DeleteFunc(r.routes[link], func(x netip.Prefix) bool { return x == p })
	r.Calls = append(r.Calls, "del "+p.String())
	return nil
}

// Mutations is how many routes were added or removed.
func (r *Routes) Mutations() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.Calls)
}

// Sysctl is a fake sysctl writer.
type Sysctl struct {
	mu     sync.Mutex
	Values map[string]string
	Writes int
}

func (s *Sysctl) Apply(params []proto.Sysctl) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.Values == nil {
		s.Values = map[string]string{}
	}
	n := 0
	for _, p := range params {
		if s.Values[p.Key] != p.Value {
			s.Values[p.Key] = p.Value
			n++
		}
	}
	s.Writes += n
	return n, nil
}

// Firewall is a fake firewall step.
type Firewall struct {
	mu      sync.Mutex
	Applied []*proto.Firewall
	Err     error
	// Rechecks counts the drift check's Recheck calls.
	Rechecks int
}

func (f *Firewall) Apply(_ context.Context, fw *proto.Firewall) (int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.Applied = append(f.Applied, fw)
	return 0, f.Err
}

func (f *Firewall) Recheck() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.Rechecks++
}

// Rechecked is Rechecks, safely while an agent is running.
func (f *Firewall) Rechecked() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.Rechecks
}

// DNS is a fake DNS step. It records what each apply asked for, counts a
// change when that differs from the last one, and answers the state a host
// with systemd-resolved would (or, with Unavailable, one without).
type DNS struct {
	mu          sync.Mutex
	wants       []dns.Want
	Err         error
	Unavailable bool
	Closed      bool
}

func (d *DNS) Apply(_ context.Context, w dns.Want) (string, int, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	n := 0
	if len(d.wants) == 0 || !reflect.DeepEqual(d.wants[len(d.wants)-1], w) {
		n = 1
	}
	d.wants = append(d.wants, w)
	switch {
	case d.Err != nil:
		return "", n, d.Err
	case w.Config == nil:
		return "", n, nil
	case w.Off:
		return proto.DNSOff, n, nil
	case d.Unavailable:
		return proto.DNSUnavailable, n, nil
	}
	return proto.DNSOn, n, nil
}

func (d *DNS) Close(context.Context, string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.Closed = true
}

// Last is what the last apply asked for.
func (d *DNS) Last() (dns.Want, bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if len(d.wants) == 0 {
		return dns.Want{}, false
	}
	return d.wants[len(d.wants)-1], true
}

// SetErr makes every apply fail from now on (nil for none), safely while an
// agent is running.
func (d *DNS) SetErr(err error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.Err = err
}

// IsClosed reports whether Close was called.
func (d *DNS) IsClosed() bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.Closed
}

// Policy is a fake exitmode.Policy: the agent's rules and table routes.
type Policy struct {
	mu     sync.Mutex
	rules  []exitmode.Rule
	routes []exitmode.Route
	// NoIPv6 is a host without IPv6.
	NoIPv6 bool
	// Calls lists the mutating calls in order ("add IPv4 rule 32764: ...").
	Calls []string
	// Fail makes every mutating call fail, or with FailOn only the named
	// method, which may also be Rules or Routes.
	Fail   error
	FailOn string
}

// failure is the error method returns, if any.
func (p *Policy) failure(method string) error {
	if p.FailOn == method || p.FailOn == "" && method != "Rules" && method != "Routes" {
		return p.Fail
	}
	return nil
}

func (p *Policy) Rules() ([]exitmode.Rule, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if err := p.failure("Rules"); err != nil {
		return nil, err
	}
	return slices.Clone(p.rules), nil
}

func (p *Policy) Routes() ([]exitmode.Route, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if err := p.failure("Routes"); err != nil {
		return nil, err
	}
	return slices.Clone(p.routes), nil
}

func (p *Policy) IPv6() bool { return !p.NoIPv6 }

// entry is a rule or a route.
type entry interface {
	comparable
	String() string
}

// mutate records a call, refusing it when Fail is set or when an add finds
// what it adds already there (the kernel answers EEXIST).
func mutate[T entry](p *Policy, method string, list *[]T, x T) error {
	op, show := strings.ToLower(method[:3]), x.String()
	p.mu.Lock()
	defer p.mu.Unlock()
	if err := p.failure(method); err != nil {
		return err
	}
	switch i := slices.Index(*list, x); {
	case op == "add" && i >= 0:
		return fmt.Errorf("add %s: file exists", show)
	case op == "add":
		*list = append(*list, x)
	case i < 0:
		return fmt.Errorf("remove %s: no such entry", show)
	default:
		*list = slices.Delete(*list, i, i+1)
	}
	p.Calls = append(p.Calls, op+" "+show)
	return nil
}

func (p *Policy) AddRule(r exitmode.Rule) error   { return mutate(p, "AddRule", &p.rules, r) }
func (p *Policy) DelRule(r exitmode.Rule) error   { return mutate(p, "DelRule", &p.rules, r) }
func (p *Policy) AddRoute(r exitmode.Route) error { return mutate(p, "AddRoute", &p.routes, r) }
func (p *Policy) DelRoute(r exitmode.Route) error { return mutate(p, "DelRoute", &p.routes, r) }

// Install puts a rule in place as if something else had added it.
func (p *Policy) Install(r exitmode.Rule) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.rules = append(p.rules, r)
}

// InstallRoute puts a table route in place as if something else had added
// it.
func (p *Policy) InstallRoute(r exitmode.Route) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.routes = append(p.routes, r)
}
