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
package apply

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/netip"
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
}

// StepError is a fatal step's failure.
type StepError struct {
	Step string
	Err  error
}

func (e *StepError) Error() string { return fmt.Sprintf("apply step %s: %v", e.Step, e.Err) }
func (e *StepError) Unwrap() error { return e.Err }

// Engine applies configs one at a time.
type Engine struct {
	d  Deps
	mu sync.Mutex
	// dnsOff is the node's own `meshctl dns off`.
	dnsOff bool
}

// New returns an engine.
func New(d Deps) *Engine {
	if d.Log == nil {
		d.Log = slog.New(slog.DiscardHandler)
	}
	if d.Now == nil {
		d.Now = time.Now
	}
	return &Engine{d: d}
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
	addr, err := netip.ParsePrefix(cfg.Interface.Address)
	if err != nil {
		return fail(StepValidate, err)
	}
	dev, err := wg.Desired(cfg, e.d.Key)
	if err != nil {
		return fail(StepValidate, err)
	}
	want, err := routes.Desired(cfg)
	if err != nil {
		return fail(StepValidate, err)
	}
	var rules *policy.Ruleset
	if cfg.Policy != nil {
		local := policy.Local{Self: addr.Addr()}
		if cfg.Containers != nil {
			local.Containers, _ = netip.ParsePrefix(cfg.Containers.Subnet) // validated above
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
		{StepAddress, func() (int, error) { return wg.SyncAddr(e.d.Host, e.d.Iface, addr) }},
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
	exitOn := cfg.Exit != nil && cfg.Exit.Via != ""
	switch router, ok := e.d.Host.(wg.ExitRouter); {
	case e.d.Exit != nil:
		n, err := exitmode.Sync(e.d.Exit, e.d.Iface, exitOn)
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
	// answers the containers on their gateway address.
	var also netip.Addr
	switch {
	case e.d.Containers != nil:
		n, err := containers.Sync(e.d.Containers, e.d.Sysctl, cfg.Containers, e.d.Iface)
		res.Changes += n
		if err != nil {
			res.Degraded = append(res.Degraded, StepContainers)
			res.Errors = append(res.Errors, err.Error())
			e.d.Log.Warn("containers step failed; the tunnel stays up", "err", err)
		} else if cfg.Containers != nil {
			also, _ = netip.ParseAddr(cfg.Containers.Gateway) // validated in step 1
		}
	case cfg.Containers != nil:
		res.Degraded = append(res.Degraded, StepContainers)
		res.Errors = append(res.Errors, "containers: this agent manages no container network; the mesh reaches the node, not its containers")
	}
	// 11. DNS: optional. The resolver binds the address step 3 put on the
	// interface, and the containers' gateway.
	if e.d.DNS != nil {
		state, n, err := e.d.DNS.Apply(ctx, dns.Want{Iface: e.d.Iface, IP: addr.Addr(), Also: also, Config: cfg.DNS, Exit: exitOn,
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
