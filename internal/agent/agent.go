// Package agent is `meshctl agent`: join (spec S6.1) and run (S6.3).
//
// Run applies the cached config at boot without asking the server, then
// follows the server: an SSE stream pushes new versions, a 60 s poll covers a
// proxy that drops the stream (amendment A16), and a heartbeat every 30 s and
// after each apply reports status. Applies are serialised by the engine; the
// sync worker coalesces triggers, so a burst of events costs one fetch. On a
// server node every pushed change goes through safe apply (safeapply.go).
// Every 5 minutes the drift check applies the config on the interface again,
// after the steps that remember the host look again, so a change made
// underneath the agent (a firewall rule removed, wg0 moved to another zone,
// a route deleted) is put back. The mesh view for the local UI is fetched
// at start, every 5 minutes and when asked for, and cached in mesh.json so
// it stays readable while the server is down.
package agent

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"net"
	"net/http"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/hafio/vpn-mesh-utility/internal/agent/apply"
	"github.com/hafio/vpn-mesh-utility/internal/agent/client"
	"github.com/hafio/vpn-mesh-utility/internal/agent/dns"
	"github.com/hafio/vpn-mesh-utility/internal/agent/execrun"
	"github.com/hafio/vpn-mesh-utility/internal/agent/exitmode"
	"github.com/hafio/vpn-mesh-utility/internal/agent/firewall"
	"github.com/hafio/vpn-mesh-utility/internal/agent/localapi"
	"github.com/hafio/vpn-mesh-utility/internal/agent/localui"
	"github.com/hafio/vpn-mesh-utility/internal/agent/routes"
	"github.com/hafio/vpn-mesh-utility/internal/agent/state"
	"github.com/hafio/vpn-mesh-utility/internal/agent/sysctl"
	"github.com/hafio/vpn-mesh-utility/internal/agent/wg"
	"github.com/hafio/vpn-mesh-utility/internal/proto"
)

// Interface is the mesh interface on Linux (spec S7.1).
const Interface = "wg0"

// Timing defaults (spec S6.3, amendment A16).
const (
	defaultPoll       = 60 * time.Second
	defaultStatus     = 30 * time.Second
	defaultBackoffMin = time.Second
	defaultBackoffMax = 60 * time.Second
	// retryFailed is how long a config that failed to apply is left alone
	// before the same version is tried again (a newer one is tried at once).
	retryFailed = 5 * time.Minute
	// defaultSafeTick is how often safe apply's window is checked.
	defaultSafeTick = time.Second
	// defaultReassert is how often the drift check runs.
	defaultReassert = 5 * time.Minute
	// defaultMeshEvery is how often the cached mesh view is refreshed.
	defaultMeshEvery = 5 * time.Minute
	// meshFresh is how long a mesh view is shown before the server is
	// asked again; meshTimeout bounds asking, so a server that is down
	// costs the local UI a few seconds, not the client's 30.
	meshFresh   = 10 * time.Second
	meshTimeout = 5 * time.Second
)

// Deps configure Run. The host seams (Host, Routes, Exit, Sysctl, Firewall,
// DNS, Transport) default to the real thing; tests replace them.
type Deps struct {
	DataDir string
	// Socket is the local API socket.
	Socket string
	// UIListen is the local UI's loopback address (localui.ParseAddr);
	// empty runs no UI.
	UIListen string
	Version  string
	// NoFirewall leaves the host firewall alone (the operator manages it).
	NoFirewall bool
	// Backend is the WireGuard backend: auto (the default), kernel or
	// userspace. With Host set it only names what Host stands for in the
	// status report; empty means kernel.
	Backend string
	Log     *slog.Logger
	// Logs is the agent's recent log (a logbuf.Ring that Log also writes
	// to), served to `meshctl logs`; nil serves none.
	Logs localapi.Logs
	Now  func() time.Time

	Host      wg.Host
	Routes    routes.Table
	Exit      exitmode.Policy
	Sysctl    apply.Sysctls
	Firewall  apply.Firewall
	DNS       apply.DNS
	Transport http.RoundTripper

	PollEvery   time.Duration
	StatusEvery time.Duration
	BackoffMin  time.Duration
	BackoffMax  time.Duration
	// ConfirmWindow is safe apply's window, SafeTick how often it is
	// checked; zero means 60 s and 1 s.
	ConfirmWindow time.Duration
	SafeTick      time.Duration
	// ReassertEvery is the drift check's period; zero means 5 minutes.
	ReassertEvery time.Duration
	// MeshEvery is how often the mesh view is refreshed; zero means 5
	// minutes.
	MeshEvery time.Duration
}

func (d *Deps) defaults() {
	if d.Log == nil {
		d.Log = slog.New(slog.DiscardHandler)
	}
	if d.Backend == "" && d.Host == nil {
		d.Backend = wg.BackendAuto
	}
	if d.Now == nil {
		d.Now = time.Now
	}
	if d.PollEvery <= 0 {
		d.PollEvery = defaultPoll
	}
	if d.StatusEvery <= 0 {
		d.StatusEvery = defaultStatus
	}
	if d.BackoffMin <= 0 {
		d.BackoffMin = defaultBackoffMin
	}
	if d.BackoffMax <= 0 {
		d.BackoffMax = defaultBackoffMax
	}
	if d.ConfirmWindow <= 0 {
		d.ConfirmWindow = defaultConfirmWindow
	}
	if d.SafeTick <= 0 {
		d.SafeTick = defaultSafeTick
	}
	if d.ReassertEvery <= 0 {
		d.ReassertEvery = defaultReassert
	}
	if d.MeshEvery <= 0 {
		d.MeshEvery = defaultMeshEvery
	}
}

// runner is one running agent.
type runner struct {
	d      Deps
	dir    *state.Dir
	enr    state.Enrollment
	engine *apply.Engine
	cl     *client.Client
	log    *slog.Logger
	// backend is the WireGuard backend in use, as reported.
	backend string

	syncKick   chan struct{}
	statusKick chan struct{}

	// cfgMu orders the config transitions -- apply, confirm, rollback -- and
	// the files that record them.
	cfgMu sync.Mutex
	// safe is safe apply's state, on server nodes only.
	safe *safeApply

	mu          sync.Mutex
	applied     *proto.NodeConfig
	appliedHash string
	appliedAt   time.Time
	lastContact time.Time
	degraded    []string
	errs        []string
	dnsState    string
	failedHash  string
	failedAt    time.Time
	// rejectedHash is a config safe apply rolled back; it is not applied
	// again until the server's config differs.
	rejectedHash string
	// rollback is reported until a newer config is confirmed.
	rollback *proto.Rollback

	// meshMu guards the mesh view and serialises asking for it;
	// meshStale is set by a new config, which means the mesh changed.
	meshMu    sync.Mutex
	mesh      proto.LocalMesh
	meshErr   string
	meshTried time.Time
	meshStale atomic.Bool
}

// Run runs the agent until ctx ends.
func Run(ctx context.Context, d Deps) error {
	d.defaults()
	dir, err := state.Open(d.DataDir)
	if err != nil {
		return err
	}
	enr, err := dir.Enrollment()
	if errors.Is(err, state.ErrNotJoined) {
		return fmt.Errorf("%w: run `meshctl agent join --server <url> --token <token>` first (data dir %s)", err, d.DataDir)
	}
	if err != nil {
		return err
	}
	keys, err := dir.LoadKeys()
	if err != nil {
		return err
	}
	host, backend := d.Host, d.Backend
	if host == nil {
		if host, backend, err = wg.Open(d.Backend, d.Log); err != nil {
			return err
		}
		defer host.Close()
	} else if backend != wg.BackendUserspace {
		backend = wg.BackendKernel
	}
	if d.Routes == nil {
		d.Routes = routes.NewTable()
	}
	if d.Exit == nil {
		d.Exit = exitmode.NewPolicy()
	}
	if d.Sysctl == nil {
		d.Sysctl = sysctl.New()
	}
	if d.Firewall == nil && !d.NoFirewall && enr.Role == proto.RoleServer {
		d.Firewall = &firewall.Firewall{Run: execrun.Exec{}, Dir: dir.Path, Iface: Interface}
	}
	if d.DNS == nil {
		d.DNS = dns.New(dns.NewRegistrar(), d.Log)
	}
	local, err := dir.Local()
	if err != nil {
		d.Log.Warn("ignoring the node's own settings; DNS stays on", "err", err)
	}
	cl, err := client.New(client.Options{Server: enr.Server, NodeID: enr.NodeID, Key: keys.Identity, Version: d.Version, Now: d.Now, Transport: d.Transport})
	if err != nil {
		return err
	}
	r := &runner{
		d: d, dir: dir, enr: enr, cl: cl, backend: backend,
		log:      d.Log.With("node", enr.Name),
		syncKick: make(chan struct{}, 1), statusKick: make(chan struct{}, 1),
		engine: apply.New(apply.Deps{Host: host, Routes: d.Routes, Exit: d.Exit, Sysctl: d.Sysctl, Firewall: d.Firewall,
			DNS: d.DNS, Iface: Interface, Key: wg.Key(keys.WG), Log: d.Log}),
	}
	r.engine.SetDNSOff(local.DNSOff)
	if r.mesh, err = dir.Mesh(); err != nil {
		r.log.Warn("ignoring the cached mesh view", "err", err)
	}
	// The resolver stops with the agent, after every loop has.
	defer func() {
		cctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		r.engine.Close(cctx)
	}()
	if enr.Role == proto.RoleServer {
		r.safe = &safeApply{window: d.ConfirmWindow}
	}

	// Boot from cache: the mesh comes up even when the server is down. A
	// config safe apply was still waiting on was never confirmed, so the
	// confirmed one is what comes back.
	cfg, err := dir.Config()
	if err != nil {
		r.log.Warn("ignoring the cached config", "err", err)
	}
	r.recoverPending(cfg)
	if cfg != nil {
		r.apply(ctx, *cfg, "cache")
	}

	// The socket first: it is the one listener the agent cannot run
	// without, and nothing has started yet when it fails.
	var api net.Listener
	if d.Socket != "" {
		if api, err = localapi.Listen(d.Socket); err != nil {
			return err
		}
	}
	var loops sync.WaitGroup
	uiToken, err := r.startUI(ctx, &loops)
	if err != nil {
		if api != nil {
			_ = api.Close() // failing anyway
		}
		return err
	}
	if api != nil {
		loops.Add(1)
		go func() {
			defer loops.Done()
			f := localapi.Funcs{Status: r.localStatus, Prefs: r.prefs, DNS: r.setDNS, Mesh: r.meshView, Logs: d.Logs, UIToken: uiToken}
			if err := localapi.Serve(ctx, api, localapi.Handler(f)); err != nil {
				r.log.Error("local API stopped", "err", err)
			}
		}()
	}
	r.log.Info("meshctl agent running", "server", enr.Server, "node_id", enr.NodeID, "version", d.Version, "wg_backend", backend)

	for _, loop := range []func(context.Context){r.syncLoop, r.eventsLoop, r.pollLoop, r.statusLoop, r.safeLoop, r.reassertLoop, r.meshLoop} {
		loops.Add(1)
		go func() {
			defer loops.Done()
			loop(ctx)
		}()
	}
	r.trigger()
	<-ctx.Done()
	loops.Wait()
	return nil
}

// startUI starts the local UI (amendment A7) and returns what mints its
// sign-in links. A UI that cannot listen leaves the agent running without
// it -- the tunnel matters more -- and `meshctl ui` says why.
func (r *runner) startUI(ctx context.Context, loops *sync.WaitGroup) (func(context.Context) (proto.UIToken, error), error) {
	down := "the local UI is off on this node (`meshctl agent run --ui-listen off`)"
	refuse := func(context.Context) (proto.UIToken, error) {
		return proto.UIToken{}, &localapi.Refusal{Status: http.StatusConflict, Code: proto.CodeConflict, Message: down}
	}
	if r.d.UIListen == "" {
		return refuse, nil
	}
	addr, err := localui.ParseAddr(r.d.UIListen)
	if err != nil {
		return nil, err
	}
	ui, err := localui.New(localui.Config{Addr: addr, Now: r.d.Now, Log: r.log, Funcs: localui.Funcs{
		Status: r.localStatus, Prefs: r.prefs, DNS: r.setDNS, Mesh: r.meshView, Logs: r.d.Logs, Restart: r.restart}})
	if err != nil {
		return nil, err
	}
	ln, err := localui.Listen(addr)
	if err != nil {
		r.log.Error("local UI not started; the agent runs without it", "err", err)
		down = "the local UI is not running: " + err.Error()
		return refuse, nil
	}
	loops.Add(1)
	go func() {
		defer loops.Done()
		if err := ui.Serve(ctx, ln); err != nil {
			r.log.Error("local UI stopped", "err", err)
		}
	}()
	r.log.Info("local UI listening; `meshctl ui` prints a sign-in link", "addr", addr.String())
	return ui.Token, nil
}

func kick(ch chan struct{}) {
	select {
	case ch <- struct{}{}:
	default:
	}
}

// trigger asks the sync worker to fetch the config.
func (r *runner) trigger() { kick(r.syncKick) }

func (r *runner) etag() string {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.appliedHash == "" {
		return ""
	}
	return proto.ETag(r.appliedHash)
}

func (r *runner) contact() {
	r.mu.Lock()
	r.lastContact = r.d.Now()
	r.mu.Unlock()
}

// syncLoop fetches and applies the config whenever it is triggered.
func (r *runner) syncLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-r.syncKick:
		}
		cfg, err := r.cl.Config(ctx, r.etag())
		if err != nil {
			r.serverError("fetch config", err)
			continue
		}
		r.contact()
		if cfg == nil {
			continue // 304: already applied
		}
		r.apply(ctx, *cfg, "server")
	}
}

// apply runs the engine and records the outcome. A config that failed is not
// retried until a newer one arrives or retryFailed passes, so a bad config
// does not loop; one safe apply rolled back is not retried at all until the
// server's config differs (amendment A15).
func (r *runner) apply(ctx context.Context, cfg proto.NodeConfig, source string) {
	r.cfgMu.Lock()
	defer r.cfgMu.Unlock()
	hash, err := cfg.Hash()
	if err != nil {
		r.log.Error("cannot hash config", "err", err)
		return
	}
	r.mu.Lock()
	failed := hash == r.failedHash && r.d.Now().Sub(r.failedAt) < retryFailed
	rejected := hash == r.rejectedHash
	r.mu.Unlock()
	if failed || rejected {
		r.log.Debug("not applying a config that failed or was rolled back", "version", cfg.Version, "rolled_back", rejected)
		return
	}
	res, err := r.engine.Apply(ctx, cfg)
	r.mu.Lock()
	if err != nil {
		r.failedHash, r.failedAt = hash, r.d.Now()
		r.errs = []string{err.Error()}
		r.mu.Unlock()
		r.log.Error("config not applied; the previous one stays in place", "version", cfg.Version, "source", source, "err", err)
		kick(r.statusKick)
		return
	}
	prev, prevHash := r.applied, r.appliedHash
	c := cfg
	r.applied, r.appliedHash, r.appliedAt = &c, res.Hash, r.d.Now()
	r.failedHash, r.degraded, r.errs, r.dnsState = "", res.Degraded, res.Errors, res.DNS
	if source == "server" {
		r.rejectedHash = "" // the server moved on from the rejected config
		r.meshStale.Store(true)
	}
	pending := false
	if r.safe != nil && source == "server" {
		switch {
		case r.safe.pending != nil && hash == r.safe.confirmedHash():
			r.safe.disarm() // the server went back to the confirmed config
		case prev != nil && hash != prevHash:
			r.safe.arm(cfg, hash, *prev, r.d.Now())
			pending = true
		}
	}
	r.mu.Unlock()
	switch {
	case pending:
		// config.json stays the confirmed config until the server answers.
		if err := r.dir.SavePending(cfg); err != nil {
			r.log.Error("config applied but not recorded as pending; a restart before it is confirmed keeps it", "err", err)
		}
	case source != "cache":
		if err := r.dir.SaveConfig(cfg); err != nil {
			r.log.Error("config applied but not cached; a reboot without the server would miss it", "err", err)
		}
		if err := r.dir.ClearPending(); err != nil {
			r.log.Warn("could not remove the pending config", "err", err)
		}
	}
	r.log.Info("config applied", "version", cfg.Version, "source", source, "changes", res.Changes, "degraded", res.Degraded,
		"awaiting_confirmation", pending)
	kick(r.statusKick)
}

// recoverPending handles a config safe apply was still waiting on when the
// agent stopped: it was never confirmed, so the node boots from the
// confirmed config, rejects the pending one until the server's differs, and
// reports the rollback.
func (r *runner) recoverPending(confirmed *proto.NodeConfig) {
	p, err := r.dir.PendingConfig()
	if err != nil {
		r.log.Warn("ignoring the pending config", "err", err)
	}
	if p == nil {
		if err != nil {
			_ = r.dir.ClearPending() // damaged; nothing to recover from it
		}
		return
	}
	hash, err := p.Hash()
	if err != nil {
		r.log.Warn("ignoring the pending config", "err", err)
		_ = r.dir.ClearPending() // it cannot be compared with anything
		return
	}
	to := int64(0)
	if confirmed != nil {
		to = confirmed.Version
	}
	r.mu.Lock()
	r.rejectedHash = hash
	r.rollback = &proto.Rollback{From: p.Version, To: to,
		Reason: fmt.Sprintf("the agent stopped before the server confirmed v%d, so the node is back on v%d", p.Version, to)}
	r.mu.Unlock()
	if err := r.dir.ClearPending(); err != nil {
		r.log.Warn("could not remove the pending config", "err", err)
	}
	r.log.Warn("safe apply: rolled back a config that was never confirmed", "from", p.Version, "to", to)
}

// confirm finishes safe apply once a report sent after the apply is
// answered: the new config becomes the cached one.
func (r *runner) confirm(sentAt time.Time) {
	if r.safe == nil {
		return
	}
	r.cfgMu.Lock()
	defer r.cfgMu.Unlock()
	r.mu.Lock()
	p, ok := r.safe.confirm(sentAt)
	if ok {
		r.rollback = nil
	}
	r.mu.Unlock()
	if !ok {
		return
	}
	if err := r.dir.SaveConfig(p.cfg); err != nil {
		r.log.Error("config confirmed but not cached; a reboot without the server would miss it", "err", err)
	}
	if err := r.dir.ClearPending(); err != nil {
		r.log.Warn("could not remove the pending config", "err", err)
	}
	r.log.Info("safe apply: the server confirmed the config", "version", p.cfg.Version)
}

// safeLoop rolls back a config the server did not confirm in time.
func (r *runner) safeLoop(ctx context.Context) {
	if r.safe == nil {
		return
	}
	t := time.NewTicker(r.d.SafeTick)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
		}
		r.cfgMu.Lock()
		r.mu.Lock()
		p, ok := r.safe.expired(r.d.Now())
		r.mu.Unlock()
		if ok {
			r.rollBack(ctx, p)
		}
		r.cfgMu.Unlock()
	}
}

// rollBack puts the last confirmed config back. The caller holds cfgMu.
func (r *runner) rollBack(ctx context.Context, p pendingApply) {
	reason := fmt.Sprintf("no answer from the server within %s of applying v%d", r.d.ConfirmWindow, p.cfg.Version)
	res, err := r.engine.Apply(ctx, p.prev)
	r.mu.Lock()
	r.rejectedHash = p.hash
	r.rollback = &proto.Rollback{From: p.cfg.Version, To: p.prev.Version, Reason: reason}
	if err != nil {
		r.errs = []string{fmt.Sprintf("safe apply: rolling back to v%d failed: %v", p.prev.Version, err)}
	} else {
		prev := p.prev
		r.applied, r.appliedHash, r.appliedAt = &prev, res.Hash, r.d.Now()
		r.degraded, r.errs, r.dnsState = res.Degraded, res.Errors, res.DNS
	}
	r.mu.Unlock()
	if cerr := r.dir.ClearPending(); cerr != nil {
		r.log.Warn("could not remove the pending config", "err", cerr)
	}
	if err != nil {
		r.log.Error("safe apply: the rollback failed too; check the interface by hand", "from", p.cfg.Version, "to", p.prev.Version, "err", err)
	} else {
		r.log.Error("safe apply: rolled back", "from", p.cfg.Version, "to", p.prev.Version, "reason", reason)
	}
	kick(r.statusKick)
}

// reassert applies the config on the interface again after the steps that
// remember the host look again: the drift check, and `meshctl dns on|off`.
func (r *runner) reassert(ctx context.Context) (apply.Result, error) {
	return r.reapply(ctx, r.engine.Reassert)
}

// reapply runs an engine pass over the config on the interface and records
// the outcome; nothing applied yet is nothing to do.
func (r *runner) reapply(ctx context.Context, run func(context.Context, proto.NodeConfig) (apply.Result, error)) (apply.Result, error) {
	r.cfgMu.Lock()
	defer r.cfgMu.Unlock()
	r.mu.Lock()
	cfg := r.applied
	r.mu.Unlock()
	if cfg == nil {
		return apply.Result{}, nil
	}
	res, err := run(ctx, *cfg)
	r.mu.Lock()
	if err != nil {
		r.errs = []string{err.Error()}
	} else {
		r.degraded, r.errs, r.dnsState = res.Degraded, res.Errors, res.DNS
	}
	r.mu.Unlock()
	if err != nil || res.Changes > 0 {
		kick(r.statusKick)
	}
	return res, err
}

// restart is the local UI's "restart tunnel": every peer handshakes anew
// and the drift check's look at the host runs at once.
func (r *runner) restart(ctx context.Context) error {
	r.mu.Lock()
	none := r.applied == nil
	r.mu.Unlock()
	if none {
		return &localapi.Refusal{Status: http.StatusConflict, Code: proto.CodeConflict,
			Message: "no config is applied yet, so there is no tunnel to restart; `meshctl status` shows why"}
	}
	res, err := r.reapply(ctx, r.engine.Restart)
	if err != nil {
		r.log.Error("tunnel restart failed; the drift check tries again", "err", err)
		return err
	}
	r.log.Info("tunnel restarted from the local UI: every peer handshakes anew", "changes", res.Changes)
	return nil
}

// reassertLoop is the drift check.
func (r *runner) reassertLoop(ctx context.Context) {
	t := time.NewTicker(r.d.ReassertEvery)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
		}
		res, err := r.reassert(ctx)
		switch {
		case ctx.Err() != nil:
			return
		case err != nil:
			r.log.Error("drift check: the config could not be put back; the next check tries again", "err", err)
		case res.Changes > 0:
			r.log.Warn("drift check: put back what changed on the host since the last apply", "changes", res.Changes, "degraded", res.Degraded)
		}
	}
}

// setDNS turns the node's split DNS on or off (`meshctl dns on|off`). The
// choice is the node's own (amendment A10): kept in local.json, applied at
// once to the config on the interface, never sent to the server.
func (r *runner) setDNS(ctx context.Context, on bool) (proto.LocalStatus, error) {
	if err := r.dir.SaveLocal(state.Local{DNSOff: !on}); err != nil {
		return proto.LocalStatus{}, &localapi.Refusal{Status: http.StatusInternalServerError, Code: proto.CodeInternal, Message: err.Error()}
	}
	r.engine.SetDNSOff(!on)
	if _, err := r.reassert(ctx); err != nil {
		return proto.LocalStatus{}, &localapi.Refusal{Status: http.StatusInternalServerError, Code: proto.CodeInternal,
			Message: "the setting is saved, but applying it failed: " + err.Error()}
	}
	st, err := r.localStatus(ctx)
	r.log.Info("split DNS changed on this node", "on", on, "dns", st.DNS)
	return st, err
}

// meshView answers GET /v1/mesh (spec S7.6): the server's mesh view, asked
// for again once the last try is meshFresh old or a new config arrived;
// when the server does not answer, the one held (from mesh.json after a
// restart) and why.
func (r *runner) meshView(ctx context.Context) (proto.LocalMesh, error) {
	r.meshMu.Lock()
	defer r.meshMu.Unlock()
	if r.meshStale.Swap(false) || r.d.Now().Sub(r.meshTried) >= meshFresh {
		r.refreshMesh(ctx)
	}
	m := r.mesh
	m.Error = r.meshErr
	return m, nil
}

// refreshMesh asks the server for the mesh view and caches it. The caller
// holds meshMu.
func (r *runner) refreshMesh(ctx context.Context) {
	r.meshTried = r.d.Now()
	fctx, cancel := context.WithTimeout(ctx, meshTimeout)
	defer cancel()
	v, err := r.cl.Mesh(fctx)
	if err != nil {
		r.serverError("fetch the mesh view", err)
		r.meshErr = "the server did not answer: " + oneLine(err.Error())
		return
	}
	r.contact()
	r.mesh, r.meshErr = proto.LocalMesh{View: v, FetchedAt: r.meshTried.Unix()}, ""
	if err := r.dir.SaveMesh(r.mesh); err != nil {
		r.log.Warn("mesh view not cached; the local UI cannot show it while the server is down", "err", err)
	}
}

// meshLoop fetches the mesh view at start and every MeshEvery, so mesh.json
// is recent when the server goes down.
func (r *runner) meshLoop(ctx context.Context) {
	t := time.NewTicker(r.d.MeshEvery)
	defer t.Stop()
	for {
		r.meshMu.Lock()
		r.refreshMesh(ctx)
		r.meshMu.Unlock()
		select {
		case <-ctx.Done():
			return
		case <-t.C:
		}
	}
}

// serverError logs a failed server call at a level that matches its cause.
func (r *runner) serverError(what string, err error) {
	var ce *client.Error
	switch {
	case errors.Is(err, context.Canceled):
	case errors.As(err, &ce) && ce.Code == proto.CodeClockSkew:
		r.log.Error(what+": this node's clock is off; sync it with NTP", "server_time", time.Unix(ce.ServerTime, 0).UTC(), "err", err)
	case errors.As(err, &ce) && ce.Code == proto.CodeNodeDisabled:
		r.log.Error(what+": an operator disabled this node; it keeps its last config and retries", "err", err)
	default:
		r.log.Warn(what+" failed; will retry", "err", err)
	}
}

// eventsLoop keeps the SSE stream open, reconnecting with jittered backoff
// from BackoffMin to BackoffMax. A stream that stayed up resets the backoff.
func (r *runner) eventsLoop(ctx context.Context) {
	backoff := r.d.BackoffMin
	for ctx.Err() == nil {
		opened := time.Now()
		err := r.cl.Events(ctx, func(ev proto.ConfigEvent) {
			r.contact()
			r.mu.Lock()
			stale := ev.Hash != r.appliedHash && ev.Hash != r.rejectedHash
			r.mu.Unlock()
			if stale {
				r.trigger()
			}
		})
		if ctx.Err() != nil {
			return
		}
		r.serverError("event stream", err)
		if time.Since(opened) > 2*r.d.BackoffMax {
			backoff = r.d.BackoffMin
		}
		wait := backoff/2 + rand.N(backoff/2+1)
		select {
		case <-ctx.Done():
			return
		case <-time.After(wait):
		}
		backoff = min(2*backoff, r.d.BackoffMax)
		r.trigger() // the stream was down: something may have changed
	}
}

func (r *runner) pollLoop(ctx context.Context) {
	t := time.NewTicker(r.d.PollEvery)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			r.trigger()
		}
	}
}

func (r *runner) statusLoop(ctx context.Context) {
	t := time.NewTicker(r.d.StatusEvery)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
		case <-r.statusKick:
		}
		sentAt := r.d.Now()
		rep := r.report()
		resp, err := r.cl.Status(ctx, rep)
		if err != nil {
			r.serverError("status report", err)
			continue
		}
		r.contact()
		r.confirm(sentAt)
		r.mu.Lock()
		missed := resp.ConfigHash != "" && resp.ConfigHash != r.appliedHash && resp.ConfigHash != r.failedHash &&
			resp.ConfigHash != r.rejectedHash
		r.mu.Unlock()
		if missed {
			r.trigger()
		}
	}
}

// report builds the heartbeat.
func (r *runner) report() proto.StatusReport {
	r.mu.Lock()
	rep := proto.StatusReport{AgentVersion: r.d.Version, WGBackend: r.backend, Peers: []proto.PeerStatus{},
		Errors: slices.Clone(r.errs), Degraded: slices.Clone(r.degraded), DNS: r.dnsState}
	if r.rollback != nil {
		rb := *r.rollback
		rep.Rollback = &rb
	}
	applied := r.applied
	if applied != nil {
		rep.AppliedVersion, rep.AppliedHash = applied.Version, r.appliedHash
	}
	r.mu.Unlock()
	if applied != nil {
		peers, err := r.engine.Stats(*applied)
		if err != nil {
			rep.Errors = append(rep.Errors, "read peer status: "+err.Error())
		} else {
			rep.Peers = peers
		}
	}
	if len(rep.Errors) > 32 {
		rep.Errors = rep.Errors[:32]
	}
	for i, e := range rep.Errors {
		rep.Errors[i] = oneLine(e)
	}
	return rep
}

// oneLine fits an error into a status report, which the server validates as
// text: valid UTF-8, no control characters, at most 1 KiB.
func oneLine(s string) string {
	s = strings.Map(func(c rune) rune {
		if unicode.IsControl(c) {
			return ' '
		}
		return c
	}, strings.ToValidUTF8(s, "?"))
	if len(s) > 1024 {
		s = s[:1024]
		for !utf8.ValidString(s) {
			s = s[:len(s)-1]
		}
	}
	return s
}

func (r *runner) localStatus(context.Context) (proto.LocalStatus, error) {
	rep := r.report()
	r.mu.Lock()
	defer r.mu.Unlock()
	st := proto.LocalStatus{
		Node:   proto.Self{ID: r.enr.NodeID, Name: r.enr.Name, Role: r.enr.Role},
		MeshIP: r.enr.MeshIP, Server: r.enr.Server, AgentVersion: r.d.Version, WGBackend: rep.WGBackend,
		AppliedVersion: rep.AppliedVersion, AppliedHash: rep.AppliedHash, Peers: rep.Peers,
		Errors: rep.Errors, Degraded: rep.Degraded, Rollback: rep.Rollback, DNS: rep.DNS,
	}
	if !r.appliedAt.IsZero() {
		st.AppliedAt = r.appliedAt.Unix()
	}
	if !r.lastContact.IsZero() {
		st.LastContact = r.lastContact.Unix()
	}
	if r.applied != nil && r.applied.Exit != nil {
		st.Exit, st.ExitChoices = r.applied.Exit.Via, slices.Clone(r.applied.Exit.Choices)
	}
	if r.applied != nil {
		st.MTU = r.applied.Interface.MTU
		if r.enr.Role == proto.RoleSpoke && len(r.applied.Peers) > 0 {
			st.Keepalive = r.applied.Peers[0].Keepalive
		}
	}
	return st, nil
}

// prefs relays a local change of this node's settings (`meshctl exit`) to the
// server as a signed request (amendment A10), then fetches the config that
// carries it at once rather than waiting for the event.
func (r *runner) prefs(ctx context.Context, req proto.PrefsRequest) (proto.PrefsResponse, error) {
	resp, err := r.cl.Prefs(ctx, req)
	if err != nil {
		var ce *client.Error
		if errors.As(err, &ce) {
			return proto.PrefsResponse{}, &localapi.Refusal{Status: ce.Status, Code: ce.Code, Message: ce.Message}
		}
		return proto.PrefsResponse{}, err
	}
	r.contact()
	r.trigger()
	return resp, nil
}

// DefaultSocket is the local API socket for a data dir: the runtime dir for
// the standard data dir, else inside the data dir.
func DefaultSocket(dataDir string) string {
	if dataDir == "/var/lib/meshctl" {
		return "/run/meshctl/agent.sock"
	}
	return filepath.Join(dataDir, "agent.sock")
}
