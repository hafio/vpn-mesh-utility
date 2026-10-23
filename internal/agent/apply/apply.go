// Package apply is the agent's apply engine: it turns a NodeConfig into host
// state in fixed steps (docs/agent-protocol.md "Apply"). Every step compares
// desired with observed state and changes only the difference, so applying
// the same config twice changes nothing (TestApplyIsIdempotent).
//
// Steps 1-8 are fatal: if one fails the config is not recorded as applied and
// the agent reports the error. Step 8 is a server's mesh firewall (amendment
// A43), installed before step 9 lets the mesh through firewalld's mesh zone.
// Steps 9-11 (firewall, the container network, DNS) are optional: a failure
// is reported as degraded and the tunnel stays up.
//
// Reassert is the drift check: it applies a config again after the steps
// that remember what they found on the host (Rechecker) forget it, so a
// change made underneath the agent is put back.
//
// A policy with schedules is compiled for the time of the apply; the
// Result says when a schedule next starts or ends, the time to apply again.
//
// A host that cannot use IPv6 (M8) applies a dual-stack config without its
// IPv6 side (forHost) and says so in Result.Families, which the agent
// reports; the server then plans the node without IPv6.
package apply

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/netip"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/hafio/vpn-mesh-utility/internal/agent/containers"
	"github.com/hafio/vpn-mesh-utility/internal/agent/dns"
	"github.com/hafio/vpn-mesh-utility/internal/agent/exitmode"
	"github.com/hafio/vpn-mesh-utility/internal/agent/policy"
	"github.com/hafio/vpn-mesh-utility/internal/agent/routes"
	"github.com/hafio/vpn-mesh-utility/internal/agent/wg"
	"github.com/hafio/vpn-mesh-utility/internal/proto"
)

// Step names, as reported in errors and status.
const (
	StepValidate   = "validate"
	StepInterface  = "interface"
	StepAddress    = "address"
	StepPeers      = "peers"
	StepRoutes     = "routes"
	StepExit       = "exit"
	StepSysctl     = "sysctl"
	StepPolicy     = "policy"
	StepFirewall   = "firewall"
	StepContainers = "containers"
	StepDNS        = "dns"
)

// Sysctls applies kernel parameters (internal/agent/sysctl).
type Sysctls interface {
	Apply(params []proto.Sysctl) (int, error)
}

// IPv6Host is a Sysctls that knows whether the host can use IPv6 (M8);
// without one, the host is taken to.
type IPv6Host interface {
	IPv6() bool
}

// MeshFirewall enforces a server's policy (internal/agent/policy): Sync
// makes the kernel hold the ruleset, or no table for nil.
type MeshFirewall interface {
	Sync(rs *policy.Ruleset) (int, error)
}

// Firewall applies a firewall section (internal/agent/firewall).
type Firewall interface {
	Apply(ctx context.Context, fw *proto.Firewall) (int, error)
}

// DNS serves the mesh zone and registers it with the host
// (internal/agent/dns). Apply returns the state (proto.DNSStates); Close
// stops it when the agent stops.
type DNS interface {
	Apply(ctx context.Context, w dns.Want) (string, int, error)
	Close(ctx context.Context, iface string)
}

// Rechecker is a step that remembers what it found on the host; Recheck
// makes its next Apply look again.
type Rechecker interface {
	Recheck()
}

// Deps are the engine's view of the host.
type Deps struct {
	Host   wg.Host
	Routes routes.Table
	// Exit is the host's policy routing; nil refuses a config that turns
	// exit mode on.
	Exit   exitmode.Policy
	Sysctl Sysctls
	// MeshFirewall is nil on a node that is not a server; a config with a
	// policy is then refused.
	MeshFirewall MeshFirewall
	// Firewall is nil when firewall management is off (--no-firewall, or a
	// node that is not a server).
	Firewall Firewall
	// DNS is nil when the agent runs no resolver.
	DNS DNS
	// Containers is the host of the node's container network (amendment
	// A37); nil refuses a config that has one, as degraded.
	Containers containers.Net
	// Iface is the mesh interface ("wg0").
	Iface string
	// Key is the node's WireGuard private key.
	Key wg.Key
	Log *slog.Logger
	// Now is the clock a policy's schedules are read on; nil is time.Now.
	Now func() time.Time
}

// Result is one successful apply.
type Result struct {
	Version int64
	Hash    string
	// Changes counts what the steps changed; 0 means the host already
	// matched the config.
	Changes int
	// Degraded names the optional steps that failed; Errors says why.
	Degraded []string
	Errors   []string
	// DNS is the DNS step's state (proto.DNSStates); empty without one.
	DNS string
	// PolicyNext is when a schedule of the policy starts or ends, so the
	// config must be applied again; zero without one.
	PolicyNext time.Time
	// Families are the address families the host can use (M8), as the
	// status reports them.
	Families []string
}

// StepError is a fatal step's failure.
type StepError struct {
	Step string
	Err  error
}

func (e *StepError) Error() string { return fmt.Sprintf("apply step %s: %v", e.Step, e.Err) }
func (e *StepError) Unwrap() error { return e.Err }

// Endpoint failover (M8): how long a dialling peer may go without a
// handshake on the endpoint it is on before Failover tries its other one --
// staleAfter after the last handshake (a tunnel with keepalive renews it
// every two minutes), retryAfter after the endpoint was set without one.
const (
	staleAfter = 3 * time.Minute
	retryAfter = 45 * time.Second
)

// Engine applies configs one at a time.
type Engine struct {
	d  Deps
	mu sync.Mutex
	// dnsOff is the node's own `meshctl dns off`.
	dnsOff bool
	// dial is each dialling peer that has a fallback endpoint (M8), by key.
	dial map[wg.Key]*dialing
	// names6 are the mesh's containers' IPv6 addresses (the mesh view's
	// names, M8), sorted, which the containers' side answers for too.
	names6 []netip.Addr
}

// dialing is how a peer with two endpoints is reached: on which, and since
// when.
type dialing struct {
	name              string
	primary, fallback netip.AddrPort
	onFallback        bool
	since             time.Time
}

// New returns an engine.
func New(d Deps) *Engine {
	if d.Log == nil {
		d.Log = slog.New(slog.DiscardHandler)
	}
	if d.Now == nil {
		d.Now = time.Now
	}
	return &Engine{d: d, dial: map[wg.Key]*dialing{}}
}

// Apply runs the steps for cfg.
func (e *Engine) Apply(ctx context.Context, cfg proto.NodeConfig) (Result, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	res := Result{Version: cfg.Version}
	fail := func(step string, err error) (Result, error) {
		return Result{}, &StepError{Step: step, Err: err}
	}

	// 1. Nothing on the host is touched until the whole config validates.
	if err := cfg.Validate(); err != nil {
		return fail(StepValidate, err)
	}
	hash, err := cfg.Hash()
	if err != nil {
		return fail(StepValidate, err)
	}
	res.Hash = hash
	ipv6 := true
	if h, ok := e.d.Sysctl.(IPv6Host); ok {
		ipv6 = h.IPv6()
	}
	res.Families = []string{proto.FamilyIPv4}
	if ipv6 {
		res.Families = append(res.Families, proto.FamilyIPv6)
	}
	// What the server sent is what the hash and the report are of; the
	// steps apply what this host can carry.
	if cfg, err = forHost(cfg, ipv6); err != nil {
		return fail(StepValidate, err)
	}
	var addrs []netip.Prefix
	for _, s := range []string{cfg.Interface.Address, cfg.Interface.Address6} {
		if s == "" {
			continue // a node has one or both (validated)
		}
		p, err := netip.ParsePrefix(s)
		if err != nil {
			return fail(StepValidate, err)
		}
		addrs = append(addrs, p)
	}
	if len(addrs) == 0 {
		return fail(StepValidate, errors.New("the config gives the interface no address"))
	}
	// The node's own address for the resolver and the mesh firewall: its
	// IPv4 one, the IPv6 one on an IPv6-only node; self6 is the IPv6 one
	// beside an IPv4 one, where the resolver listens too.
	self, self6 := addrs[0].Addr(), netip.Addr{}
	if len(addrs) > 1 {
		self6 = addrs[1].Addr()
	}
	dev, err := wg.Desired(cfg, e.d.Key)
	if err != nil {
		return fail(StepValidate, err)
	}
	e.keepEndpoints(cfg, &dev)
	want, err := routes.Desired(cfg)
	if err != nil {
		return fail(StepValidate, err)
	}
	var rules *policy.Ruleset
	if cfg.Policy != nil {
		var local policy.Local // the node's own addresses, of both families (M8)
		for _, a := range addrs {
			if a.Addr().Is4() {
				local.Self = a.Addr()
			} else {
				local.Self6 = a.Addr()
			}
		}
		if cfg.Containers != nil {
			local.Containers, _ = netip.ParsePrefix(cfg.Containers.Subnet) // validated above
			local.Containers6, _ = netip.ParsePrefix(cfg.Containers.Subnet6)
		}
		if rules, err = policy.Compile(*cfg.Policy, local, e.d.Now()); err != nil {
			return fail(StepValidate, err)
		}
		res.PolicyNext = rules.Next
	}

	// 2-5. Interface, address, WireGuard, routes.
	steps := []struct {
		name string
		run  func() (int, error)
	}{
		{StepInterface, func() (int, error) { return wg.SyncLink(e.d.Host, e.d.Iface, cfg.Interface.MTU) }},
		{StepAddress, func() (int, error) { return wg.SyncAddrs(e.d.Host, e.d.Iface, addrs) }},
		{StepPeers, func() (int, error) { return wg.SyncDevice(e.d.Host, e.d.Iface, dev) }},
		{StepRoutes, func() (int, error) { return routes.Sync(e.d.Routes, e.d.Iface, want) }},
	}
	for _, s := range steps {
		n, err := s.run()
		res.Changes += n
		if err != nil {
			return fail(s.name, err)
		}
	}

	// 6. Exit mode: the policy rules follow the config, so exit off (every
	// server's case) removes rules a previous config left.
	// Where the platform has no policy routing (Windows), the host does it.
	// The exit carries IPv6 when its peer has ::/0 (M8; never on a host
	// without IPv6, forHost took it away).
	exitOn := cfg.Exit != nil && cfg.Exit.Via != ""
	via6 := exitOn && slices.ContainsFunc(cfg.Peers, func(p proto.Peer) bool {
		return p.Name == cfg.Exit.Via && slices.Contains(p.AllowedIPs, "::/0")
	})
	switch router, ok := e.d.Host.(wg.ExitRouter); {
	case e.d.Exit != nil:
		n, err := exitmode.Sync(e.d.Exit, e.d.Iface, exitOn, via6)
		res.Changes += n
		switch {
		case errors.Is(err, exitmode.ErrIPv6Unguarded):
			// IPv4 is in place; only the host's IPv6 is not (amendment A45).
			res.Degraded = append(res.Degraded, StepExit)
			res.Errors = append(res.Errors, err.Error())
			e.d.Log.Warn("exit mode carries IPv4 only", "err", err)
		case err != nil:
			return fail(StepExit, err)
		}
	case ok:
		n, err := router.SetExit(e.d.Iface, exitOn)
		res.Changes += n
		if err != nil {
			return fail(StepExit, err)
		}
	case exitOn:
		return fail(StepExit, fmt.Errorf("exit via %q was requested, but this agent has no policy routing; turn it off with `meshctl exit off`", cfg.Exit.Via))
	}

	// 7. The planner asks for sysctls only when the node relays or exits,
	// so a failure here would break the mesh: fatal.
	if len(cfg.Sysctl) > 0 {
		n, err := e.d.Sysctl.Apply(cfg.Sysctl)
		res.Changes += n
		if err != nil {
			return fail(StepSysctl, err)
		}
	}

	// 8. The mesh firewall: fatal, so a server never runs a config whose
	// policy it does not enforce (safe apply puts the previous one back).
	if rules != nil {
		if e.d.MeshFirewall == nil {
			return fail(StepPolicy, errors.New("this node enforces no mesh firewall, but its config has a policy (only servers do)"))
		}
		n, err := e.d.MeshFirewall.Sync(rules)
		res.Changes += n
		if err != nil {
			return fail(StepPolicy, err)
		}
	}
	// 9. Firewall: optional.
	firewallFailed := false
	if e.d.Firewall != nil {
		n, err := e.d.Firewall.Apply(ctx, cfg.Firewall)
		res.Changes += n
		if err != nil {
			firewallFailed = true
			res.Degraded = append(res.Degraded, StepFirewall)
			res.Errors = append(res.Errors, err.Error())
			e.d.Log.Warn("firewall step failed; the tunnel stays up", "err", err)
		}
	}
	// A config without a policy takes the table away -- after firewalld has
	// closed the mesh zone again, and not when it could not (the firewall is
	// degraded already): a stale table filtering is safer than an open zone
	// with none, and the next apply tries again.
	if rules == nil && e.d.MeshFirewall != nil && !firewallFailed {
		n, err := e.d.MeshFirewall.Sync(nil)
		res.Changes += n
		if err != nil {
			res.Degraded = append(res.Degraded, StepPolicy)
			res.Errors = append(res.Errors, err.Error())
			e.d.Log.Warn("mesh firewall not removed", "err", err)
		}
	}
	// 10. The container network: optional, and before DNS, whose resolver
	// answers the containers on their gateway address. Without IPv6 policy
	// routing its IPv4 side is up all the same.
	var also netip.Addr
	switch {
	case e.d.Containers != nil:
		n, err := containers.Sync(e.d.Containers, e.d.Sysctl, cfg.Containers, e.d.Iface, proxied(cfg, addrs, e.names6))
		res.Changes += n
		if err != nil {
			res.Degraded = append(res.Degraded, StepContainers)
			res.Errors = append(res.Errors, err.Error())
			e.d.Log.Warn("containers step failed; the tunnel stays up", "err", err)
		}
		if cfg.Containers != nil && (err == nil || errors.Is(err, containers.ErrIPv6Unrouted)) {
			also, _ = netip.ParseAddr(cfg.Containers.Gateway) // validated in step 1
		}
	case cfg.Containers != nil:
		res.Degraded = append(res.Degraded, StepContainers)
		res.Errors = append(res.Errors, "containers: this agent manages no container network; the mesh reaches the node, not its containers")
	}
	// 11. DNS: optional. The resolver binds the address step 3 put on the
	// interface, and the containers' gateway.
	if e.d.DNS != nil {
		state, n, err := e.d.DNS.Apply(ctx, dns.Want{Iface: e.d.Iface, IP: self, IP6: self6, Also: also, Config: cfg.DNS, Exit: exitOn,
			Off: e.dnsOff})
		res.Changes += n
		res.DNS = state
		if err != nil {
			res.Degraded = append(res.Degraded, StepDNS)
			res.Errors = append(res.Errors, err.Error())
			e.d.Log.Warn("dns step failed; the tunnel stays up", "err", err)
		}
	}
	return res, nil
}

// forHost is cfg as this host can carry it (M8). A host that can use IPv6
// carries all of it; one that cannot gets it without the IPv6 side -- the
// interface's IPv6 address, every IPv6 route and endpoint (a server's IPv4
// fallback endpoint takes the place of its IPv6 one), the IPv6 sysctls,
// the exit's IPv6 masquerade and the containers' IPv6 -- unless the node is
// IPv6 only, which such a host cannot run at all. cfg itself is left as it
// was.
func forHost(cfg proto.NodeConfig, ipv6 bool) (proto.NodeConfig, error) {
	switch {
	case ipv6:
		return cfg, nil
	case cfg.Interface.Address == "":
		return cfg, errors.New("this host cannot use IPv6 (net.ipv6.conf.all or .default.disable_ipv6 is 1, or the kernel has no IPv6), " +
			"and the node is IPv6 only: turn IPv6 on, or give the node IPv4 (`meshctl node set <name> --family dual`)")
	}
	v6 := func(s string) bool {
		ap, err := netip.ParseAddrPort(s)
		return err == nil && ap.Addr().Is6()
	}
	cfg.Interface.Address6 = ""
	peers := make([]proto.Peer, 0, len(cfg.Peers))
	for _, p := range cfg.Peers {
		p.AllowedIPs = slices.DeleteFunc(slices.Clone(p.AllowedIPs), func(s string) bool {
			pf, err := netip.ParsePrefix(s)
			return err == nil && pf.Addr().Is6()
		})
		if v6(p.Endpoint) {
			p.Endpoint = p.Fallback // the IPv4 one, or none
		}
		p.Fallback = "" // the other family's: IPv6 now, or taken
		peers = append(peers, p)
	}
	cfg.Peers = peers
	cfg.Sysctl = slices.DeleteFunc(slices.Clone(cfg.Sysctl), func(s proto.Sysctl) bool { return strings.HasPrefix(s.Key, "net.ipv6.") })
	if cfg.Firewall != nil {
		fw := *cfg.Firewall
		fw.Exit6 = ""
		cfg.Firewall = &fw
	}
	if cfg.Containers != nil {
		k := *cfg.Containers
		k.Subnet6, k.Network6, k.Range6, k.Bridge6, k.Gateway6 = "", "", "", "", ""
		cfg.Containers = &k
	}
	return cfg, nil
}

// proxied is what the containers' side of the node answers neighbour
// discovery for (M8): the node's own IPv6 mesh address, every single IPv6
// address it routes over the tunnel and the mesh's containers' (names6),
// sorted.
func proxied(cfg proto.NodeConfig, addrs []netip.Prefix, names6 []netip.Addr) []netip.Addr {
	out := slices.Clone(names6)
	for _, a := range addrs {
		if a.Addr().Is6() {
			out = append(out, a.Addr())
		}
	}
	for _, p := range cfg.Peers {
		for _, s := range p.AllowedIPs {
			if pf, err := netip.ParsePrefix(s); err == nil && pf.Addr().Is6() && pf.IsSingleIP() {
				out = append(out, pf.Addr())
			}
		}
	}
	slices.SortFunc(out, netip.Addr.Compare)
	return slices.Compact(out)
}

// keepEndpoints keeps each dialling peer on the endpoint Failover chose:
// the device dials a peer's fallback while Failover has it there, until a
// config gives the peer other endpoints. Only a peer the node keeps alive
// (a spoke's servers) is followed: an idle server-to-server tunnel has no
// handshakes to judge it by. The caller holds mu.
func (e *Engine) keepEndpoints(cfg proto.NodeConfig, dev *wg.Device) {
	now, seen := e.d.Now(), map[wg.Key]bool{}
	for _, p := range cfg.Peers {
		k, err := wg.ParseKey(p.PublicKey)
		primary, perr := netip.ParseAddrPort(p.Endpoint)
		fallback, ferr := netip.ParseAddrPort(p.Fallback)
		if err != nil || perr != nil || ferr != nil || p.Keepalive == 0 {
			continue
		}
		seen[k] = true
		d := e.dial[k]
		if d == nil || d.primary != primary || d.fallback != fallback {
			d = &dialing{name: p.Name, primary: primary, fallback: fallback, since: now}
			e.dial[k] = d
		}
		if !d.onFallback {
			continue
		}
		for i := range dev.Peers {
			if dev.Peers[i].PublicKey == k {
				dev.Peers[i].Endpoint = fallback
			}
		}
	}
	for k := range e.dial {
		if !seen[k] {
			delete(e.dial, k)
		}
	}
}

// Failover moves each dialling peer whose endpoint has gone without a
// handshake for too long (staleAfter, retryAfter) to its other endpoint
// (M8): a spoke whose IPv4 path to a server is broken tries the server's
// IPv6 endpoint, and back. It returns what it moved, "<peer> to
// <endpoint>", sorted.
func (e *Engine) Failover() ([]string, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if len(e.dial) == 0 {
		return nil, nil
	}
	_, stats, err := e.d.Host.Device(e.d.Iface)
	if err != nil {
		return nil, err
	}
	last := map[wg.Key]time.Time{}
	for _, s := range stats {
		last[s.PublicKey] = s.LastHandshake
	}
	now := e.d.Now()
	var (
		c     wg.Change
		moved []string
		which []*dialing
	)
	for k, d := range e.dial {
		switch h := last[k]; {
		case h.After(d.since) && now.Sub(h) < staleAfter:
			continue // it answers on this endpoint
		case !h.After(d.since) && now.Sub(d.since) < retryAfter:
			continue // not long enough on this endpoint to judge it
		}
		ep := d.fallback
		if d.onFallback {
			ep = d.primary
		}
		c.Peers = append(c.Peers, wg.PeerChange{PublicKey: k, Endpoint: &ep})
		moved, which = append(moved, d.name+" to "+ep.String()), append(which, d)
	}
	if len(moved) == 0 {
		return nil, nil
	}
	slices.Sort(moved)
	if err := e.d.Host.Configure(e.d.Iface, c); err != nil {
		return nil, fmt.Errorf("dial %s: %w", strings.Join(moved, ", "), err)
	}
	for _, d := range which {
		d.onFallback, d.since = !d.onFallback, now
	}
	return moved, nil
}

// SetMeshNames takes the mesh view's container names (M8): the containers'
// side of the node answers neighbour discovery for their IPv6 addresses
// too, from the next apply on. It reports whether those addresses changed.
func (e *Engine) SetMeshNames(ns []proto.MeshName) bool {
	var addrs []netip.Addr
	for _, n := range ns {
		if a, err := netip.ParseAddr(n.IP); err == nil && a.Is6() {
			addrs = append(addrs, a)
		}
	}
	slices.SortFunc(addrs, netip.Addr.Compare)
	addrs = slices.Compact(addrs)
	e.mu.Lock()
	defer e.mu.Unlock()
	if slices.Equal(addrs, e.names6) {
		return false
	}
	e.names6 = addrs
	return true
}

// SetDNSOff records the node's own `meshctl dns on|off`; the next apply
// acts on it.
func (e *Engine) SetDNSOff(off bool) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.dnsOff = off
}

// Reassert applies cfg again after the steps that remember the host forget
// what they saw, so whatever changed underneath the agent since the last
// apply is put back. Changes counts what it put back.
func (e *Engine) Reassert(ctx context.Context, cfg proto.NodeConfig) (Result, error) {
	e.mu.Lock()
	for _, step := range []any{e.d.Firewall, e.d.DNS} {
		if r, ok := step.(Rechecker); ok {
			r.Recheck()
		}
	}
	e.mu.Unlock()
	return e.Apply(ctx, cfg)
}

// Restart drops every peer's session, then applies cfg as Reassert does:
// each peer is added back with no state, so every tunnel handshakes anew,
// and the firewall and DNS steps look at the host again (the local UI's
// "restart tunnel"). The interface, its address and routes stay up.
func (e *Engine) Restart(ctx context.Context, cfg proto.NodeConfig) (Result, error) {
	e.mu.Lock()
	dropped, err := e.dropPeers()
	e.mu.Unlock()
	if err != nil {
		return Result{}, &StepError{Step: StepPeers, Err: fmt.Errorf("drop the peers to restart the tunnel: %w", err)}
	}
	res, err := e.Reassert(ctx, cfg)
	res.Changes += dropped
	return res, err
}

// dropPeers removes every peer of an interface that is ours. The caller
// holds mu.
func (e *Engine) dropPeers() (int, error) {
	l, err := e.d.Host.Link(e.d.Iface)
	if err != nil || !l.Exists || !l.Ours {
		return 0, err // Reassert creates it, or refuses a foreign one
	}
	dev, _, err := e.d.Host.Device(e.d.Iface)
	if err != nil {
		return 0, err
	}
	var c wg.Change
	for _, p := range dev.Peers {
		c.Peers = append(c.Peers, wg.PeerChange{PublicKey: p.PublicKey, Remove: true})
	}
	if len(c.Peers) == 0 {
		return 0, nil
	}
	return len(c.Peers), e.d.Host.Configure(e.d.Iface, c)
}

// Close stops what the steps run beside the interface (the resolver), when
// the agent stops.
func (e *Engine) Close(ctx context.Context) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.d.DNS != nil {
		e.d.DNS.Close(ctx, e.d.Iface)
	}
}

// Stats returns the runtime state of the peers, named after cfg's peers.
// Peers the config does not know are omitted.
func (e *Engine) Stats(cfg proto.NodeConfig) ([]proto.PeerStatus, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	_, stats, err := e.d.Host.Device(e.d.Iface)
	if err != nil {
		return nil, err
	}
	names := map[wg.Key]string{}
	for _, p := range cfg.Peers {
		if k, err := wg.ParseKey(p.PublicKey); err == nil {
			names[k] = p.Name
		}
	}
	out := []proto.PeerStatus{}
	for _, s := range stats {
		name, ok := names[s.PublicKey]
		if !ok {
			continue
		}
		ps := proto.PeerStatus{Name: name, RxBytes: s.RxBytes, TxBytes: s.TxBytes}
		if !s.LastHandshake.IsZero() && s.LastHandshake.After(time.Unix(0, 0)) {
			ps.LastHandshake = s.LastHandshake.Unix()
		}
		if s.Endpoint.IsValid() {
			ps.Endpoint = s.Endpoint.String()
		}
		out = append(out, ps)
	}
	return out, nil
}

// IsStepError reports whether err is a fatal step failure and which step.
func IsStepError(err error) (string, bool) {
	var se *StepError
	if errors.As(err, &se) {
		return se.Step, true
	}
	return "", false
}
