package agent

import (
	"context"
	"crypto/ed25519"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/netip"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/hafio/vpn-mesh-utility/internal/agent/agentfake"
	"github.com/hafio/vpn-mesh-utility/internal/agent/apply"
	"github.com/hafio/vpn-mesh-utility/internal/agent/localapi"
	"github.com/hafio/vpn-mesh-utility/internal/agent/logbuf"
	"github.com/hafio/vpn-mesh-utility/internal/agent/names"
	"github.com/hafio/vpn-mesh-utility/internal/agent/policy"
	"github.com/hafio/vpn-mesh-utility/internal/agent/state"
	"github.com/hafio/vpn-mesh-utility/internal/agent/wg"
	"github.com/hafio/vpn-mesh-utility/internal/proto"
	"github.com/hafio/vpn-mesh-utility/internal/testfix"
)

// fakeServer speaks the agent API for one node and checks every signature
// against that node's identity key. The agent tests cannot import the real
// server (TestModuleBoundaries), so this is the contract seen from the agent.
type fakeServer struct {
	t   *testing.T
	pub ed25519.PublicKey
	// refuse, when set, is the error every signed call is answered with.
	refuse *proto.ErrorBody

	mu      sync.Mutex
	cfg     proto.NodeConfig
	reports []proto.StatusReport
	push    chan proto.ConfigEvent
	// statusDown makes status answer 503, as a server the new config cut
	// off would look to safe apply.
	statusDown bool
	// prefs answers POST prefs; nil answers 404.
	prefs func(proto.PrefsRequest) (proto.PrefsResponse, *proto.ErrorBody)
	// mesh answers GET mesh, nil 503; meshAsked counts the calls.
	mesh      *proto.MeshView
	meshAsked int
	// nudge sends a mesh event; names records every PUT names, and
	// namesRefuse answers the next one with an error.
	nudge       chan struct{}
	names       []proto.NamesRequest
	namesRefuse *proto.ErrorBody
}

func newFakeServer(t *testing.T, pub ed25519.PublicKey, cfg proto.NodeConfig, refuse *proto.ErrorBody) (*fakeServer, *httptest.Server) {
	f := &fakeServer{t: t, pub: pub, cfg: cfg, refuse: refuse, push: make(chan proto.ConfigEvent, 4), nudge: make(chan struct{}, 1)}
	mux := http.NewServeMux()
	mux.HandleFunc("GET "+proto.PathConfig, f.config)
	mux.HandleFunc("GET "+proto.PathEvents, f.events)
	mux.HandleFunc("POST "+proto.PathStatus, f.status)
	mux.HandleFunc("POST "+proto.PathPrefs, f.prefsHandler)
	mux.HandleFunc("GET "+proto.PathMesh, f.meshHandler)
	mux.HandleFunc("PUT "+proto.PathNames, f.namesHandler)
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return f, srv
}

func (f *fakeServer) verify(w http.ResponseWriter, r *http.Request) ([]byte, bool) {
	body, _ := io.ReadAll(r.Body)
	s, err := proto.ParseSigned(r.Header)
	if err == nil {
		err = s.Verify(f.pub, r.Method, r.URL.RequestURI(), body, time.Now())
	}
	if err != nil {
		f.t.Errorf("%s %s: %v", r.Method, r.URL.Path, err)
		w.WriteHeader(http.StatusUnauthorized)
		return nil, false
	}
	if f.refuse != nil {
		status := http.StatusUnauthorized
		if f.refuse.Code == proto.CodeNodeDisabled {
			status = http.StatusForbidden
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(f.refuse)
		return nil, false
	}
	return body, true
}

func (f *fakeServer) set(cfg proto.NodeConfig) proto.ConfigEvent {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.cfg = cfg
	h, _ := cfg.Hash()
	return proto.ConfigEvent{Version: cfg.Version, Hash: h}
}

func (f *fakeServer) config(w http.ResponseWriter, r *http.Request) {
	if _, ok := f.verify(w, r); !ok {
		return
	}
	f.mu.Lock()
	cfg := f.cfg
	f.mu.Unlock()
	h, _ := cfg.Hash()
	w.Header().Set("ETag", proto.ETag(h))
	if r.Header.Get("If-None-Match") == proto.ETag(h) {
		w.WriteHeader(http.StatusNotModified)
		return
	}
	_ = json.NewEncoder(w).Encode(cfg)
}

func (f *fakeServer) events(w http.ResponseWriter, r *http.Request) {
	if _, ok := f.verify(w, r); !ok {
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	rc := http.NewResponseController(w)
	_ = proto.WriteRetry(w, time.Second)
	_ = rc.Flush()
	for {
		select {
		case <-r.Context().Done():
			return
		case ev := <-f.push:
			b, _ := json.Marshal(ev)
			if proto.WriteEvent(w, proto.Event{Name: proto.EventConfig, Data: b}) != nil || rc.Flush() != nil {
				return
			}
		case <-f.nudge:
			if proto.WriteEvent(w, proto.Event{Name: proto.EventMesh, Data: []byte("{}")}) != nil || rc.Flush() != nil {
				return
			}
		}
	}
}

func (f *fakeServer) namesHandler(w http.ResponseWriter, r *http.Request) {
	body, ok := f.verify(w, r)
	if !ok {
		return
	}
	var req proto.NamesRequest
	if err := json.Unmarshal(body, &req); err != nil || req.Validate() != nil {
		f.t.Errorf("names %s: %v", body, err)
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if eb := f.namesRefuse; eb != nil {
		f.namesRefuse = nil
		w.WriteHeader(http.StatusConflict)
		_ = json.NewEncoder(w).Encode(eb)
		return
	}
	f.names = append(f.names, req)
	resp := proto.NamesResponse{Stored: len(req.Names)}
	for _, n := range req.Names {
		if n.Name == "vps-1" {
			resp.Dropped = append(resp.Dropped, proto.DroppedName{Name: n.Name, IP: n.IP, Reason: "node vps-1 has this name"})
		}
	}
	_ = json.NewEncoder(w).Encode(resp)
}

func (f *fakeServer) namesPut() []proto.NamesRequest {
	f.mu.Lock()
	defer f.mu.Unlock()
	return slices.Clone(f.names)
}

func (f *fakeServer) status(w http.ResponseWriter, r *http.Request) {
	body, ok := f.verify(w, r)
	if !ok {
		return
	}
	var rep proto.StatusReport
	if err := json.Unmarshal(body, &rep); err != nil || rep.Validate() != nil {
		f.t.Errorf("status report %s: %v", body, err)
	}
	f.mu.Lock()
	if f.statusDown {
		f.mu.Unlock()
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(proto.ErrorBody{Code: proto.CodeInternal, Message: "down"})
		return
	}
	f.reports = append(f.reports, rep)
	h, _ := f.cfg.Hash()
	resp := proto.StatusResponse{ConfigVersion: f.cfg.Version, ConfigHash: h}
	f.mu.Unlock()
	_ = json.NewEncoder(w).Encode(resp)
}

func (f *fakeServer) prefsHandler(w http.ResponseWriter, r *http.Request) {
	body, ok := f.verify(w, r)
	if !ok {
		return
	}
	var req proto.PrefsRequest
	if err := json.Unmarshal(body, &req); err != nil || f.prefs == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	resp, refuse := f.prefs(req)
	if refuse != nil {
		w.WriteHeader(http.StatusConflict)
		_ = json.NewEncoder(w).Encode(refuse)
		return
	}
	_ = json.NewEncoder(w).Encode(resp)
}

func (f *fakeServer) meshHandler(w http.ResponseWriter, r *http.Request) {
	if _, ok := f.verify(w, r); !ok {
		return
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.meshAsked++
	if f.mesh == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(proto.ErrorBody{Code: proto.CodeInternal, Message: "the server is down"})
		return
	}
	_ = json.NewEncoder(w).Encode(f.mesh)
}

func (f *fakeServer) setMesh(v *proto.MeshView) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.mesh = v
}

func (f *fakeServer) meshCalls() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.meshAsked
}

func (f *fakeServer) setStatusDown(down bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.statusDown = down
}

func (f *fakeServer) lastReport() (proto.StatusReport, bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.reports) == 0 {
		return proto.StatusReport{}, false
	}
	return f.reports[len(f.reports)-1], true
}

// newDir is a data dir with keys, as `agent join` creates them.
func newDir(t *testing.T) (*state.Dir, ed25519.PublicKey) {
	t.Helper()
	dir, err := state.Open(filepath.Join(shortTemp(t), "data"))
	if err != nil {
		t.Fatal(err)
	}
	keys, err := dir.LoadOrCreateKeys()
	if err != nil {
		t.Fatal(err)
	}
	return dir, keys.Identity.Public().(ed25519.PublicKey)
}

// enroll records the enrollment `agent join` would have written for vps-1.
func enroll(t *testing.T, dir *state.Dir, server string) { enrollAs(t, dir, server, testfix.VPS1) }

func enrollAs(t *testing.T, dir *state.Dir, server string, n testfix.Node) {
	t.Helper()
	if err := dir.SaveEnrollment(state.Enrollment{Server: server, NodeID: n.ID, Name: n.Name, Role: n.Role, MeshIP: n.MeshIP}); err != nil {
		t.Fatal(err)
	}
}

// joinedTo is a data dir enrolled with a fake server serving cfg.
func joinedTo(t *testing.T, cfg proto.NodeConfig) (*state.Dir, *fakeServer) {
	t.Helper()
	dir, pub := newDir(t)
	f, srv := newFakeServer(t, pub, cfg, nil)
	enroll(t, dir, srv.URL)
	return dir, f
}

// syncBuffer is a log sink the agent's goroutines may share.
type syncBuffer struct {
	mu sync.Mutex
	b  strings.Builder
}

func (s *syncBuffer) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.b.Write(p)
}

func (s *syncBuffer) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.b.String()
}

func shortTemp(t *testing.T) string {
	t.Helper()
	d, err := os.MkdirTemp("", "mag")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(d) })
	return d
}

func eventually(t *testing.T, what string, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %s", what)
}

type rig struct {
	host   *agentfake.Host
	policy *agentfake.Policy
	dns    *agentfake.DNS
	fw     *agentfake.Firewall
	kernel *agentfake.Kernel
	deps   Deps
	done   chan error
	cancel context.CancelFunc
}

func start(t *testing.T, dir *state.Dir, socket string) *rig {
	t.Helper()
	return startWith(t, dir, socket, nil)
}

func startLogged(t *testing.T, dir *state.Dir, socket string, log *slog.Logger) *rig {
	t.Helper()
	return startWith(t, dir, socket, func(d *Deps) { d.Log = log })
}

// startWith runs the agent on fakes; edit adjusts its Deps first.
func startWith(t *testing.T, dir *state.Dir, socket string, edit func(*Deps)) *rig {
	t.Helper()
	r := &rig{host: agentfake.NewHost(), policy: &agentfake.Policy{}, dns: &agentfake.DNS{}, fw: &agentfake.Firewall{}, kernel: &agentfake.Kernel{},
		done: make(chan error, 1)}
	r.deps = Deps{DataDir: dir.Path, Socket: socket, Version: "v1.0.0", Host: r.host, Routes: agentfake.NewRoutes(),
		Exit: r.policy, Sysctl: &agentfake.Sysctl{}, MeshFirewall: &policy.Enforcer{K: r.kernel}, Firewall: r.fw, DNS: r.dns,
		Containers: agentfake.NewNet(), PollEvery: 50 * time.Millisecond, StatusEvery: 50 * time.Millisecond,
		BackoffMin: 10 * time.Millisecond, BackoffMax: 50 * time.Millisecond,
		RuntimeSocket: RuntimeOff} // never the host's container runtime
	if edit != nil {
		edit(&r.deps)
	}
	ctx, cancel := context.WithCancel(context.Background())
	r.cancel = cancel
	go func() { r.done <- Run(ctx, r.deps) }()
	t.Cleanup(func() {
		cancel()
		select {
		case <-r.done:
		case <-time.After(10 * time.Second):
			t.Error("Run did not stop")
		}
	})
	return r
}

func (r *rig) peers() int {
	d, ok := r.host.DeviceOf(Interface)
	if !ok {
		return -1
	}
	return len(d.Peers)
}

func TestRunFollowsTheServer(t *testing.T) {
	cfg := testfix.ServerConfig()
	dir, f := joinedTo(t, cfg)
	sock := filepath.Join(dir.Path, "agent.sock")
	r := start(t, dir, sock)

	want, _ := cfg.Hash()
	eventually(t, "the first config and its report", func() bool {
		rep, ok := f.lastReport()
		return r.peers() == 2 && ok && rep.AppliedHash == want && rep.AppliedVersion == 3
	})

	next := testfix.ServerConfig()
	next.Version = 4
	next.Peers = next.Peers[:1]
	f.push <- f.set(next)
	eventually(t, "the pushed change", func() bool { return r.peers() == 1 })
	eventually(t, "the cache to hold v4", func() bool {
		c, err := dir.Config()
		return err == nil && c != nil && c.Version == 4
	})

	st, err := localapi.NewClient(sock).Status(context.Background())
	if err != nil || st.Node.Name != "vps-1" || st.AppliedVersion != 4 || st.LastContact == 0 {
		t.Fatalf("local status %+v, %v", st, err)
	}
}

// The agent of a mesh in slot 2 (amendment A49) applies its mesh's config
// on the slot's interface, wg2, and its local UI's default is the slot's
// port.
func TestRunNamesItsSlot(t *testing.T) {
	cfg := testfix.ServerConfig()
	cfg.Node.Slot = 2
	dir, pub := newDir(t)
	_, srv := newFakeServer(t, pub, cfg, nil)
	n := testfix.VPS1
	if err := dir.SaveEnrollment(state.Enrollment{Server: srv.URL, NodeID: n.ID, Name: n.Name, Role: n.Role, MeshIP: n.MeshIP,
		Mesh: "lab", Slot: 2}); err != nil {
		t.Fatal(err)
	}
	r := start(t, dir, "")
	eventually(t, "the config on wg2", func() bool {
		d, ok := r.host.DeviceOf("wg2")
		return ok && len(d.Peers) == len(cfg.Peers)
	})
	if _, ok := r.host.DeviceOf(Interface); ok {
		t.Fatal("slot 2's agent made wg0")
	}
	if got := uiListen(UIAuto, 2); got != "127.0.0.1:7445" {
		t.Errorf("slot 2's UI = %s", got)
	}
	if got := uiListen("127.0.0.1:9000", 2); got != "127.0.0.1:9000" {
		t.Errorf("a UI address given = %s", got)
	}
}

// The poll covers a proxy that swallows the stream (amendment A16).
func TestRunPollsWhenTheStreamIsSilent(t *testing.T) {
	dir, f := joinedTo(t, testfix.ServerConfig())
	r := start(t, dir, "")
	eventually(t, "the first config", func() bool { return r.peers() == 2 })
	next := testfix.ServerConfig()
	next.Version = 5
	next.Peers = next.Peers[1:]
	f.set(next) // no event
	eventually(t, "the polled change", func() bool { return r.peers() == 1 })
}

func TestRunBootsFromTheCacheWithoutTheServer(t *testing.T) {
	dir, _ := newDir(t)
	enroll(t, dir, "http://127.0.0.1:1") // nothing listens there
	if err := dir.SaveConfig(testfix.ServerConfig()); err != nil {
		t.Fatal(err)
	}
	r := start(t, dir, "")
	eventually(t, "the cached config on the interface", func() bool { return r.peers() == 2 })
}

func TestRunKeepsTheOldConfigWhenANewOneFails(t *testing.T) {
	dir, f := joinedTo(t, testfix.ServerConfig())
	r := start(t, dir, "")
	eventually(t, "the first config", func() bool { return r.peers() == 2 })

	bad := testfix.ServerConfig()
	bad.Version = 6
	bad.Interface.MTU = 9000 // refused by validation, before the host is touched
	f.push <- f.set(bad)
	eventually(t, "the failure in a report", func() bool {
		rep, ok := f.lastReport()
		return ok && len(rep.Errors) > 0 && strings.Contains(rep.Errors[0], "validate") && rep.AppliedVersion == 3
	})
	if r.peers() != 2 {
		t.Fatal("a refused config changed the interface")
	}
	if c, _ := dir.Config(); c == nil || c.Version != 3 {
		t.Fatal("a refused config replaced the cache")
	}
}

// `meshctl exit` reaches the server through the agent (amendment A10); the
// agent then fetches the new config at once and turns exit mode on.
func TestExitThroughTheLocalAPI(t *testing.T) {
	dir, pub := newDir(t)
	f, srv := newFakeServer(t, pub, testfix.SpokeConfig(""), nil)
	f.prefs = func(req proto.PrefsRequest) (proto.PrefsResponse, *proto.ErrorBody) {
		if *req.Exit == "vps-3" {
			return proto.PrefsResponse{}, &proto.ErrorBody{Code: proto.CodeConflict, Message: `exit "vps-3" is not exit-capable`}
		}
		cfg := testfix.SpokeConfig(*req.Exit)
		cfg.Version = 9
		ev := f.set(cfg) // no event: the agent must fetch on its own
		return proto.PrefsResponse{Version: 9, ConfigVersion: 9, ConfigHash: ev.Hash}, nil
	}
	enrollAs(t, dir, srv.URL, testfix.Laptop)
	sock := filepath.Join(dir.Path, "agent.sock")
	r := startWith(t, dir, sock, func(d *Deps) { d.PollEvery = time.Hour })
	eventually(t, "the first config", func() bool { return r.peers() == 3 })

	c := localapi.NewClient(sock)
	via := "vps-2"
	resp, err := c.Prefs(context.Background(), proto.PrefsRequest{Exit: &via})
	if err != nil || resp.Version != 9 {
		t.Fatalf("Prefs = %+v, %v", resp, err)
	}
	eventually(t, "exit mode via vps-2", func() bool {
		st, err := c.Status(context.Background())
		rules, _ := r.policy.Rules()
		return err == nil && st.AppliedHash == resp.ConfigHash && st.Exit == "vps-2" && len(st.ExitChoices) == 2 && len(rules) == 4
	})
	via = "vps-3"
	var rf *localapi.Refusal
	if _, err := c.Prefs(context.Background(), proto.PrefsRequest{Exit: &via}); !errors.As(err, &rf) || rf.Status != http.StatusConflict ||
		!strings.Contains(rf.Message, "not exit-capable") {
		t.Fatalf("a refused exit = %v", err)
	}
}

// Run stops at once, saying what to do, when its data dir or its socket is
// unusable; nothing on the host is touched.
func TestRunRefusesAnUnusableDataDir(t *testing.T) {
	run := func(d Deps) error {
		t.Helper()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return Run(ctx, d)
	}
	if err := run(Deps{DataDir: "data?x=1", Host: agentfake.NewHost()}); err == nil || !strings.Contains(err.Error(), "--data") {
		t.Errorf("a data dir with '?': %v", err)
	}

	damaged, _ := newDir(t)
	if err := os.WriteFile(filepath.Join(damaged.Path, state.EnrollmentFile), []byte("{"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := run(Deps{DataDir: damaged.Path, Host: agentfake.NewHost()}); err == nil || !strings.Contains(err.Error(), "re-join") {
		t.Errorf("a damaged enrollment: %v", err)
	}

	keyless, err := state.Open(filepath.Join(shortTemp(t), "data"))
	if err != nil {
		t.Fatal(err)
	}
	enroll(t, keyless, "http://127.0.0.1:1")
	if err := run(Deps{DataDir: keyless.Path, Host: agentfake.NewHost()}); !errors.Is(err, state.ErrNotJoined) || !strings.Contains(err.Error(), "agent join") {
		t.Errorf("an enrollment without keys: %v", err)
	}

	// Without a Host the backend flag is checked before anything is opened.
	dir, _ := newDir(t)
	enroll(t, dir, "http://127.0.0.1:1")
	if err := run(Deps{DataDir: dir.Path, Backend: "kernel;reboot"}); err == nil || !strings.Contains(err.Error(), "--wg-backend") {
		t.Errorf("an unknown backend: %v", err)
	}

	// The real routes, policy, sysctl and firewall seams are built, but the
	// socket fails before any of them is used.
	h := agentfake.NewHost()
	if err := run(Deps{DataDir: dir.Path, Host: h, Socket: strings.Repeat("x", 120)}); err == nil || !strings.Contains(err.Error(), "too long") {
		t.Errorf("an over-long socket path: %v", err)
	}
	if h.Mutations() != 0 {
		t.Errorf("a refused start changed the host: %v", h.Calls)
	}

	// With a cached config, and a pending one another agent's safe apply may
	// be waiting on, the socket still fails first: no step is asked anything
	// and the pending config stays. The two hash apart, as safe apply only
	// writes a pending config that does; with equal hashes a cache boot that
	// came first would skip the cached one as rolled back and pass here too.
	// Every seam is a fake, so such a boot would touch nothing real either.
	cfg, pending := withDNS(version(3)), withDNS(version(4))
	if err := dir.SaveConfig(cfg); err != nil {
		t.Fatal(err)
	}
	if err := dir.SavePending(pending); err != nil {
		t.Fatal(err)
	}
	h = agentfake.NewHost()
	rt, fw, resolver := agentfake.NewRoutes(), &agentfake.Firewall{}, &agentfake.DNS{}
	if err := run(Deps{DataDir: dir.Path, Host: h, Routes: rt, Exit: &agentfake.Policy{}, Sysctl: &agentfake.Sysctl{}, Firewall: fw, DNS: resolver, Socket: strings.Repeat("x", 120)}); err == nil || !strings.Contains(err.Error(), "too long") {
		t.Errorf("an over-long socket path with a cached config: %v", err)
	}
	if _, asked := resolver.Last(); h.Mutations() != 0 || rt.Mutations() != 0 || len(fw.Applied) != 0 || asked {
		t.Errorf("a refused start applied the cached config: host %v, routes %v, %d firewall applies, DNS asked %v", h.Calls, rt.Calls, len(fw.Applied), asked)
	}
	if !pendingExists(dir) {
		t.Error("a refused start removed the pending config")
	}
}

// A damaged cached or pending config is ignored, the pending one removed,
// and the node follows the server as if neither were there.
func TestRunIgnoresDamagedConfigFiles(t *testing.T) {
	dir, f := joinedTo(t, testfix.ServerConfig())
	for _, name := range []string{state.ConfigFile, state.PendingConfigFile} {
		if err := os.WriteFile(filepath.Join(dir.Path, name), []byte("{"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	var logs syncBuffer
	r := startLogged(t, dir, "", slog.New(slog.NewTextHandler(&logs, nil)))
	eventually(t, "the server's config, cached", func() bool {
		c, _ := dir.Config()
		return r.peers() == 2 && c != nil && c.Version == 3 && !pendingExists(dir)
	})
	for _, want := range []string{"ignoring the cached config", "ignoring the pending config"} {
		if !strings.Contains(logs.String(), want) {
			t.Errorf("the log lacks %q", want)
		}
	}
	if rep, ok := f.lastReport(); ok && rep.Rollback != nil {
		t.Fatalf("a damaged pending file was reported as a rollback: %+v", rep.Rollback)
	}
}

func TestRunReportsUnreadablePeerStatus(t *testing.T) {
	dir, f := joinedTo(t, testfix.ServerConfig())
	r := start(t, dir, "")
	eventually(t, "the first config", func() bool { return r.peers() == 2 })
	r.host.SetFail("Device")
	eventually(t, "the failed read in a report", func() bool {
		rep, ok := f.lastReport()
		return ok && len(rep.Errors) > 0 && strings.Contains(rep.Errors[len(rep.Errors)-1], "read peer status")
	})
}

// `meshctl exit` while the server is down says so; the agent does not
// change anything on its own.
func TestPrefsNeedTheServer(t *testing.T) {
	dir, _ := newDir(t)
	enrollAs(t, dir, "http://127.0.0.1:1", testfix.Laptop) // nothing listens there
	if err := dir.SaveConfig(testfix.SpokeConfig("")); err != nil {
		t.Fatal(err)
	}
	sock := filepath.Join(dir.Path, "agent.sock")
	r := start(t, dir, sock)
	eventually(t, "the cached config", func() bool { return r.peers() == 3 })
	via := "vps-2"
	var rf *localapi.Refusal
	_, err := localapi.NewClient(sock).Prefs(context.Background(), proto.PrefsRequest{Exit: &via})
	if !errors.As(err, &rf) || rf.Status != http.StatusBadGateway || !strings.Contains(rf.Message, proto.PathPrefs) {
		t.Fatalf("Prefs without the server = %v", err)
	}
	if rules, _ := r.policy.Rules(); len(rules) != 0 {
		t.Fatalf("exit mode turned on without the server: %v", rules)
	}
}

// A report fits the server's limits however many errors piled up.
func TestReportKeepsToTheServersLimits(t *testing.T) {
	r := &runner{d: Deps{Version: "v1.0.0"}, backend: wg.BackendKernel, errs: make([]string, 40)}
	for i := range r.errs {
		r.errs[i] = fmt.Sprintf("error %d", i)
	}
	rep := r.report()
	if len(rep.Errors) != 32 || rep.Errors[31] != "error 31" {
		t.Fatalf("report kept %d errors", len(rep.Errors))
	}
	if err := rep.Validate(); err != nil {
		t.Fatalf("the report must pass the server's validation: %v", err)
	}
}

func TestDefaultSocket(t *testing.T) {
	if got := DefaultSocket("linux", "/var/lib/meshctl"); got != "/run/meshctl/agent.sock" {
		t.Fatalf("standard data dir: %s", got)
	}
	if got, want := DefaultSocket("linux", filepath.Join("data", "mesh")), filepath.Join("data", "mesh", "agent.sock"); got != want {
		t.Fatalf("custom data dir: %s, want %s", got, want)
	}
	if got := DefaultSocket("windows", `C:\ProgramData\meshctl`); got != `\\.\pipe\meshctl-agent` {
		t.Fatalf("Windows: %s", got)
	}
}

// The report names the backend in use; a fake host stands for the kernel
// unless the test says otherwise.
func TestReportNamesTheBackend(t *testing.T) {
	for b, want := range map[string]string{"": wg.BackendKernel, wg.BackendUserspace: wg.BackendUserspace} {
		dir, f := joinedTo(t, testfix.ServerConfig())
		startWith(t, dir, "", func(d *Deps) { d.Backend = b })
		eventually(t, "a report from the "+want+" backend", func() bool {
			rep, ok := f.lastReport()
			return ok && rep.WGBackend == want
		})
	}
}

func TestRunNeedsAnEnrollment(t *testing.T) {
	err := Run(context.Background(), Deps{DataDir: filepath.Join(t.TempDir(), "data"), Host: agentfake.NewHost()})
	if !errors.Is(err, state.ErrNotJoined) || !strings.Contains(err.Error(), "agent join") {
		t.Fatalf("Run = %v", err)
	}
}

type fakeService struct {
	ok        bool
	installed bool
}

func (s *fakeService) Available() (bool, string) { return s.ok, "no systemd here" }
func (s *fakeService) Install(context.Context) error {
	s.installed = true
	return nil
}

func TestJoin(t *testing.T) {
	var got proto.JoinRequest
	nodeID, mesh, slot := int64(7), "", 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != proto.PathJoin || r.Header.Get(proto.HeaderSignature) != "" {
			t.Errorf("join must be an unsigned POST to %s", proto.PathJoin)
		}
		b, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(b, &got)
		_ = json.NewEncoder(w).Encode(proto.JoinResponse{NodeID: nodeID, Name: "vps-1", Role: proto.RoleServer, MeshIP: "10.77.0.1",
			Mesh: mesh, MeshUUID: "8c2b6f3e-0d1a-4c55-9e7a-1f2d3c4b5a69", Slot: slot})
	}))
	defer srv.Close()
	data := filepath.Join(t.TempDir(), "data")
	svc := &fakeService{ok: true}
	var out strings.Builder
	opts := JoinOptions{DataDir: data, Server: srv.URL, Token: testfix.Token, Version: "v1.0.0", Service: svc, Out: &out,
		Hostname: func() (string, error) { return "My_VPS.Example.COM", nil }, Backend: func() string { return "kernel" },
		PublicIPv4: func() (netip.Addr, bool) { return netip.MustParseAddr("198.51.100.7"), true },
		PublicIPv6: func() (netip.Addr, bool) { return netip.MustParseAddr("2001:db8::7"), true }, Endpoint6: "[2001:db8::70]:51820"}
	if _, err := Join(context.Background(), opts); err != nil {
		t.Fatal(err)
	}
	dir, _ := state.Open(data)
	keys, err := dir.LoadKeys()
	if err != nil {
		t.Fatal(err)
	}
	wgPub, _ := keys.WGPublic()
	if got.Token != testfix.Token || got.WGPublicKey != wgPub || got.IdentityPublicKey != keys.IdentityPublic() ||
		got.Hostname != "my-vps.example.com" || got.PublicIP != "198.51.100.7" || got.WGBackend != "kernel" ||
		got.PublicIP6 != "2001:db8::7" || got.Endpoint6 != "[2001:db8::70]:51820" {
		t.Fatalf("join request %+v", got)
	}
	if e, err := dir.Enrollment(); err != nil || e.NodeID != 7 || e.Server != srv.URL {
		t.Fatalf("enrollment %+v, %v", e, err)
	}
	if !svc.installed || !strings.Contains(out.String(), "installed and started") || !strings.Contains(out.String(), "(server, mesh IP 10.77.0.1, node id 7)") {
		t.Fatalf("service not installed: %s", out.String())
	}

	// Re-joining as another node drops the cached config.
	if err := dir.SaveConfig(testfix.ServerConfig()); err != nil {
		t.Fatal(err)
	}
	nodeID = 8
	opts.Service, opts.NoService = &fakeService{ok: false}, false
	out.Reset()
	if _, err := Join(context.Background(), opts); err != nil {
		t.Fatal(err)
	}
	if c, _ := dir.Config(); c != nil {
		t.Fatal("the old node's cached config survived a re-join as another node")
	}
	if !strings.Contains(out.String(), "no systemd here") {
		t.Fatalf("the reason for skipping the service must be printed: %s", out.String())
	}
	if again, _ := dir.LoadKeys(); !again.Identity.Equal(keys.Identity) {
		t.Fatal("re-join replaced the identity key")
	}

	// A mesh in slot 1 (amendment A49) is recorded, and its names said; a
	// slot this agent does not know is refused and nothing recorded.
	mesh, slot = "lab", 1
	out.Reset()
	if _, err := Join(context.Background(), opts); err != nil {
		t.Fatal(err)
	}
	if e, err := dir.Enrollment(); err != nil || e.Mesh != "lab" || e.Slot != 1 || e.MeshUUID == "" {
		t.Fatalf("enrollment %+v, %v", e, err)
	}
	if !strings.Contains(out.String(), "mesh lab, slot 1: interface wg1, local UI port 7444") {
		t.Fatalf("join output: %s", out.String())
	}
	slot = proto.MaxSlot + 1
	if _, err := Join(context.Background(), opts); err == nil || !strings.Contains(err.Error(), "slot 64") {
		t.Fatalf("join into slot 64 = %v", err)
	}
	if e, err := dir.Enrollment(); err != nil || e.Slot != 1 {
		t.Fatalf("enrollment after the refusal %+v, %v", e, err)
	}
}

func TestJoinValidatesBeforeCallingTheServer(t *testing.T) {
	called := false
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { called = true }))
	defer srv.Close()
	base := JoinOptions{DataDir: filepath.Join(t.TempDir(), "data"), Server: srv.URL, Token: testfix.Token, Version: "v1",
		NoService: true, Backend: func() string { return "kernel" }, PublicIPv4: func() (netip.Addr, bool) { return netip.Addr{}, false },
		PublicIPv6: func() (netip.Addr, bool) { return netip.Addr{}, false }}
	for name, edit := range map[string]func(*JoinOptions){
		"http to a public server": func(o *JoinOptions) { o.Server = "http://mesh.example.com" },
		"short token":             func(o *JoinOptions) { o.Token = "abc" },
		"bad name":                func(o *JoinOptions) { o.Name = "VPS 1" },
		"endpoint with a host":    func(o *JoinOptions) { o.Endpoint = "vps.example:51820" },
		"endpoint6 is IPv4":       func(o *JoinOptions) { o.Endpoint6 = "203.0.113.10:51820" },
	} {
		o := base
		edit(&o)
		if _, err := Join(context.Background(), o); err == nil {
			t.Errorf("%s: Join succeeded", name)
		}
	}
	if called {
		t.Fatal("an invalid join reached the server")
	}
}

func TestMeshAddresses(t *testing.T) {
	for want, r := range map[string]proto.JoinResponse{
		"IP 10.77.0.1":                {MeshIP: "10.77.0.1"},
		"IP 10.77.0.1 and fd10:77::1": {MeshIP: "10.77.0.1", MeshIP6: "fd10:77::1"},
		"IPv6 fd10:77::2:1":           {MeshIP6: "fd10:77::2:1"},
	} {
		if got := meshAddresses(r); got != want {
			t.Errorf("meshAddresses(%+v) = %q, want %q", r, got, want)
		}
	}
}

func TestCleanHostname(t *testing.T) {
	for in, want := range map[string]string{
		"vps-1":              "vps-1",
		"My_VPS.Example.COM": "my-vps.example.com",
		"nas.":               "nas",
		"":                   "unknown",
		"-.-":                "unknown",
		"bad name\n":         "bad-name",
	} {
		if got := cleanHostname(in); got != want {
			t.Errorf("cleanHostname(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestOneLine(t *testing.T) {
	if got := oneLine("a\nb\x07c\u0085d\xff"); got != "a b c d?" {
		t.Fatalf("oneLine = %q", got)
	}
	// 1024 bytes would end inside a two-byte rune; the rune goes.
	long := oneLine(strings.Repeat("e", 1001) + strings.Repeat("é", 100))
	if len(long) != 1023 || !utf8.ValidString(long) {
		t.Fatalf("oneLine kept %d bytes", len(long))
	}
	rep := proto.StatusReport{AgentVersion: "v1", WGBackend: "kernel", Errors: []string{long, oneLine("x\ty")}}
	if err := rep.Validate(); err != nil {
		t.Fatalf("a sanitised error must pass the server's validation: %v", err)
	}
}

// A refusal the operator must act on is logged with what to do, and the
// cached config stays on the interface.
func TestRunExplainsRefusals(t *testing.T) {
	for _, c := range []struct {
		refuse proto.ErrorBody
		want   string
	}{
		{proto.ErrorBody{Code: proto.CodeNodeDisabled, Message: "node vps-1 is disabled"}, "an operator disabled this node"},
		{proto.ErrorBody{Code: proto.CodeClockSkew, Message: "node clock is 2m0s behind", ServerTime: testfix.Epoch.Unix()}, "sync it with NTP"},
	} {
		t.Run(c.refuse.Code, func(t *testing.T) {
			dir, pub := newDir(t)
			_, srv := newFakeServer(t, pub, testfix.ServerConfig(), &c.refuse)
			enroll(t, dir, srv.URL)
			if err := dir.SaveConfig(testfix.ServerConfig()); err != nil {
				t.Fatal(err)
			}
			var logs syncBuffer
			r := startLogged(t, dir, "", slog.New(slog.NewTextHandler(&logs, nil)))
			eventually(t, "the refusal in the log", func() bool { return strings.Contains(logs.String(), c.want) })
			if r.peers() != 2 {
				t.Fatal("the cached config must stay applied")
			}
		})
	}
}

func TestIsPublic(t *testing.T) {
	for s, want := range map[string]bool{
		"203.0.113.10": true, "198.51.100.7": true, "10.0.0.1": false, "172.16.5.4": false, "192.168.1.1": false,
		"100.64.1.1": false, "127.0.0.1": false, "169.254.1.1": false, "::1": false, "2001:db8::1": false,
	} {
		if got := isPublic(netip.MustParseAddr(s)); got != want {
			t.Errorf("isPublic(%s) = %v", s, got)
		}
	}
	if ip, ok := PublicIPv4(); ok && !isPublic(ip) {
		t.Fatalf("PublicIPv4 returned a non-public address %s", ip)
	}
	for s, want := range map[string]bool{
		"2001:db8::10": true, "2a01:4f8::1": true, "fd10:77::1": false, "fe80::1": false, "::1": false, "::ffff:203.0.113.10": false,
		"ff02::1": false, "203.0.113.10": false,
	} {
		if got := isPublic6(netip.MustParseAddr(s)); got != want {
			t.Errorf("isPublic6(%s) = %v", s, got)
		}
	}
	if ip, ok := PublicIPv6(); ok && !isPublic6(ip) {
		t.Fatalf("PublicIPv6 returned a non-global address %s", ip)
	}
}

// withDNS adds a DNS section to a config.
func withDNS(cfg proto.NodeConfig) proto.NodeConfig {
	cfg.DNS = &proto.DNS{Zone: proto.DefaultZone, Split: true, Records: []proto.Record{{Name: "vps-1", IP: testfix.VPS1.MeshIP}}}
	return cfg
}

// `meshctl dns off` takes effect at once, is reported, survives a restart
// and never reaches the server; `dns on` undoes it. The resolver stops with
// the agent.
func TestDNSOnAndOffStayOnTheNode(t *testing.T) {
	dir, f := joinedTo(t, withDNS(testfix.ServerConfig()))
	sock := filepath.Join(dir.Path, "agent.sock")
	r := start(t, dir, sock)
	eventually(t, "DNS on in a report", func() bool {
		rep, ok := f.lastReport()
		return ok && rep.DNS == proto.DNSOn
	})
	w, _ := r.dns.Last()
	if w.IP.String() != testfix.VPS1.MeshIP || w.Config == nil || w.Off {
		t.Fatalf("the DNS step was asked %+v", w)
	}
	c := localapi.NewClient(sock)
	st, err := c.DNS(context.Background(), false)
	if err != nil || st.DNS != proto.DNSOff {
		t.Fatalf("dns off = %+v, %v", st, err)
	}
	if w, _ := r.dns.Last(); !w.Off {
		t.Fatal("dns off did not reach the DNS step")
	}
	eventually(t, "DNS off in a report", func() bool {
		rep, ok := f.lastReport()
		return ok && rep.DNS == proto.DNSOff
	})
	if l, err := dir.Local(); err != nil || !l.DNSOff {
		t.Fatalf("local.json = %+v, %v", l, err)
	}
	r.cancel()
	eventually(t, "the resolver to stop with the agent", r.dns.IsClosed)

	again := start(t, dir, filepath.Join(dir.Path, "again.sock"))
	eventually(t, "a restarted agent to keep DNS off", func() bool {
		w, ok := again.dns.Last()
		return ok && w.Off && w.Config != nil
	})
	st, err = localapi.NewClient(filepath.Join(dir.Path, "again.sock")).DNS(context.Background(), true)
	if err != nil || st.DNS != proto.DNSOn {
		t.Fatalf("dns on = %+v, %v", st, err)
	}
}

// A DNS step that fails leaves the tunnel up and the node degraded.
func TestDNSFailureIsReported(t *testing.T) {
	dir, f := joinedTo(t, withDNS(testfix.ServerConfig()))
	r := startWith(t, dir, "", func(d *Deps) {
		d.DNS.(*agentfake.DNS).SetErr(errors.New("dns: answer on 10.77.0.1:53: address already in use"))
	})
	eventually(t, "the degraded report", func() bool {
		rep, ok := f.lastReport()
		return ok && slices.Contains(rep.Degraded, "dns") && rep.AppliedVersion != 0 && strings.Contains(strings.Join(rep.Errors, " "), "already in use")
	})
	if r.peers() != 2 {
		t.Fatal("the tunnel is not up")
	}
}

// The drift check applies the config again, after the firewall looked again:
// an interface deleted underneath the agent comes back without a new config.
func TestDriftCheckPutsBackWhatChanged(t *testing.T) {
	dir, _ := joinedTo(t, testfix.ServerConfig())
	r := startWith(t, dir, "", func(d *Deps) { d.ReassertEvery, d.PollEvery = 20*time.Millisecond, time.Hour })
	eventually(t, "the first config", func() bool { return r.peers() == 2 })
	r.host.RemoveLink(Interface)
	eventually(t, "wg0 back with its peers", func() bool { return r.peers() == 2 })
	eventually(t, "the firewall to look again", func() bool { return r.fw.Rechecked() > 0 })
}

// A scheduled rule of the mesh firewall comes in when its window opens,
// without a new config: the agent applies again at the edge. The agent's
// clock is moved (by under 30 s, inside the signatures' skew) so the
// window opens 3 s after the start.
func TestScheduleAppliesAtTheEdge(t *testing.T) {
	now := time.Now().UTC()
	edge := now.Add(3 * time.Second).Round(time.Minute)
	off := edge.Add(-3 * time.Second).Sub(now)
	cfg := testfix.ServerConfig()
	cfg.Policy = testfix.Policy()
	cfg.Policy.TZ = "UTC"
	cfg.Policy.Rules[3].Schedule = &proto.Schedule{Start: edge.Format("15:04"), End: edge.Add(time.Hour).Format("15:04")}
	dir, _ := joinedTo(t, cfg)
	r := startWith(t, dir, "", func(d *Deps) { d.Now, d.PollEvery = func() time.Time { return time.Now().Add(off) }, time.Hour })
	scheduled := func() (bool, bool) {
		table, _ := r.kernel.Holds()
		return table != nil, table != nil && strings.Contains(table.Text(), "policy: laptop-smb")
	}
	eventually(t, "the first table", func() bool { in, _ := scheduled(); return in })
	if _, has := scheduled(); has {
		t.Fatal("the scheduled rule is in before its window")
	}
	eventually(t, "the scheduled rule when its window opens", func() bool { _, has := scheduled(); return has })
}

// The setting cannot be saved: nothing changes, and the CLI hears why.
func TestDNSSettingThatCannotBeSaved(t *testing.T) {
	dir, _ := joinedTo(t, withDNS(testfix.ServerConfig()))
	sock := filepath.Join(dir.Path, "agent.sock")
	r := start(t, dir, sock)
	eventually(t, "the first config", func() bool { return r.peers() == 2 })
	if err := os.MkdirAll(filepath.Join(dir.Path, state.LocalFile, "x"), 0o700); err != nil {
		t.Fatal(err)
	}
	var rf *localapi.Refusal
	if _, err := localapi.NewClient(sock).DNS(context.Background(), false); !errors.As(err, &rf) || rf.Status != http.StatusInternalServerError ||
		!strings.Contains(rf.Message, state.LocalFile) {
		t.Fatalf("dns off = %v", err)
	}
	if w, _ := r.dns.Last(); w.Off {
		t.Fatal("an unsaved setting was applied")
	}
}

// The mesh view is fetched at start and cached; a view under meshFresh old
// is shown without asking again, unless a new config arrived; a restarted
// agent shows the cached one, and why, while the server does not answer
// (spec S7.6).
func TestMeshViewIsCached(t *testing.T) {
	dir, f := joinedTo(t, testfix.ServerConfig())
	view := testfix.MeshView()
	f.setMesh(&view)
	sock := filepath.Join(dir.Path, "agent.sock")
	r := start(t, dir, sock)
	eventually(t, "the first config, and the mesh view in mesh.json", func() bool {
		st, err := localapi.NewClient(sock).Status(context.Background())
		m, merr := dir.Mesh()
		return err == nil && st.AppliedVersion == 3 && merr == nil && len(m.View.Nodes) == len(view.Nodes)
	})
	eventually(t, "the view fetched at start and after the first config", func() bool { return f.meshCalls() >= 2 })
	m, err := localapi.NewClient(sock).Mesh(context.Background())
	if err != nil || m.Error != "" || m.FetchedAt == 0 || len(m.View.Nodes) != len(view.Nodes) || m.View.Nodes[4] != view.Nodes[4] {
		t.Fatalf("mesh = %+v, %v", m, err)
	}
	asked := f.meshCalls()
	if _, err := localapi.NewClient(sock).Mesh(context.Background()); err != nil || f.meshCalls() != asked {
		t.Fatalf("a view under %s old: the server was asked again (%d -> %d, %v)", meshFresh, asked, f.meshCalls(), err)
	}
	// A new config means the mesh changed: the view is fetched again at
	// once, and a look right after is answered from that fetch. (The
	// version alone is not part of the hash, so a peer goes.)
	next := testfix.ServerConfig()
	next.Version = 4
	next.Peers = next.Peers[:1]
	f.push <- f.set(next)
	eventually(t, "v4 applied", func() bool {
		st, err := localapi.NewClient(sock).Status(context.Background())
		return err == nil && st.AppliedVersion == 4
	})
	eventually(t, "the view fetched after the new config", func() bool { return f.meshCalls() == asked+1 })
	if _, err := localapi.NewClient(sock).Mesh(context.Background()); err != nil || f.meshCalls() != asked+1 {
		t.Fatalf("after a new config the server was asked %d times in all, want %d (%v)", f.meshCalls(), asked+1, err)
	}

	last, err := dir.Mesh()
	if err != nil {
		t.Fatal(err)
	}
	r.cancel()
	f.setMesh(nil)
	again := filepath.Join(dir.Path, "again.sock")
	start(t, dir, again)
	var cached proto.LocalMesh
	eventually(t, "the cached view while the server is down", func() bool {
		cached, err = localapi.NewClient(again).Mesh(context.Background())
		return err == nil && cached.Error != ""
	})
	if cached.FetchedAt != last.FetchedAt || len(cached.View.Nodes) != len(view.Nodes) || !strings.Contains(cached.Error, "the server is down") {
		t.Fatalf("cached mesh = %+v", cached)
	}
}

// A view that cannot be cached is still shown, and the log says the node
// will not have it while the server is down.
func TestMeshViewThatCannotBeCached(t *testing.T) {
	dir, f := joinedTo(t, testfix.ServerConfig())
	view := testfix.MeshView()
	f.setMesh(&view)
	if err := os.MkdirAll(filepath.Join(dir.Path, state.MeshFile, "x"), 0o700); err != nil {
		t.Fatal(err)
	}
	logs := &syncBuffer{}
	sock := filepath.Join(dir.Path, "agent.sock")
	startLogged(t, dir, sock, slog.New(slog.NewTextHandler(logs, nil)))
	eventually(t, "the warning", func() bool { return strings.Contains(logs.String(), "mesh view not cached") })
	if !strings.Contains(logs.String(), "ignoring the cached mesh view") {
		t.Fatalf("an unreadable mesh.json at start is not reported:\n%s", logs.String())
	}
	if m, err := localapi.NewClient(sock).Mesh(context.Background()); err != nil || len(m.View.Nodes) != len(view.Nodes) {
		t.Fatalf("mesh = %+v, %v", m, err)
	}
}

// The view is refreshed every MeshEvery, so mesh.json is recent when the
// server goes down.
func TestMeshViewIsRefreshed(t *testing.T) {
	dir, f := joinedTo(t, testfix.ServerConfig())
	view := testfix.MeshView()
	f.setMesh(&view)
	startWith(t, dir, "", func(d *Deps) { d.MeshEvery = 20 * time.Millisecond })
	eventually(t, "three fetches", func() bool { return f.meshCalls() >= 3 })
}

// `meshctl logs` reads the ring the agent's log also writes to.
func TestLogsAreServed(t *testing.T) {
	dir, _ := joinedTo(t, testfix.ServerConfig())
	ring := logbuf.New(logbuf.DefaultLines)
	sock := filepath.Join(dir.Path, "agent.sock")
	startWith(t, dir, sock, func(d *Deps) { d.Log, d.Logs = slog.New(slog.NewTextHandler(ring, nil)), ring })
	var out strings.Builder
	eventually(t, "the start line in `meshctl logs`", func() bool {
		out.Reset()
		return localapi.NewClient(sock).Logs(context.Background(), false, &out) == nil && strings.Contains(out.String(), "meshctl agent running")
	})
}

// freePort is a loopback address nothing listens on, for the local UI.
func freePort(t *testing.T) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := ln.Addr().String()
	_ = ln.Close()
	return addr
}

// The local UI through the agent: `meshctl ui`'s link signs a browser in,
// the dashboard shows the node, and "restart the tunnel" drops and re-adds
// every peer (amendment A7).
func TestLocalUIThroughTheAgent(t *testing.T) {
	dir, _ := joinedTo(t, testfix.ServerConfig())
	sock, addr := filepath.Join(dir.Path, "agent.sock"), freePort(t)
	r := startWith(t, dir, sock, func(d *Deps) { d.UIListen = addr })
	eventually(t, "the first config", func() bool { return r.peers() == 2 })
	st, err := localapi.NewClient(sock).Status(context.Background())
	if err != nil || st.MTU != proto.DefaultMTU || st.Keepalive != 0 {
		t.Fatalf("status = %+v, %v; want the config's MTU and no keepalive on a server", st, err)
	}
	tok, err := localapi.NewClient(sock).UIToken(context.Background())
	if err != nil || !strings.HasPrefix(tok.URL, "http://"+addr+"/login?t=") {
		t.Fatalf("UIToken = %+v, %v", tok, err)
	}
	jar, _ := cookiejar.New(nil)
	browser := &http.Client{Jar: jar, Timeout: 5 * time.Second}
	resp, err := browser.Get(tok.URL)
	if err != nil {
		t.Fatal(err)
	}
	page, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	m := regexp.MustCompile(`name="csrf" value="([^"]+)"`).FindSubmatch(page)
	if resp.StatusCode != http.StatusOK || !strings.Contains(string(page), "vps-1") || m == nil {
		t.Fatalf("dashboard = %d\n%s", resp.StatusCode, page)
	}

	before := r.host.Mutations()
	req, _ := http.NewRequest(http.MethodPost, "http://"+addr+"/restart", strings.NewReader(url.Values{"csrf": {string(m[1])}}.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Origin", "http://"+addr)
	resp, err = browser.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	page, _ = io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusOK || !strings.Contains(string(page), "Tunnel restarted") || r.peers() != 2 || r.host.Mutations() != before+2 {
		t.Fatalf("restart = %d, %d peers, %d host changes (want 2: the peers dropped, then added back)\n%s",
			resp.StatusCode, r.peers(), r.host.Mutations()-before, page)
	}
}

// Without a UI, or with its port taken, the agent runs and `meshctl ui`
// says why there is no link; a UI address off loopback stops Run before
// anything starts.
func TestLocalUIOffOrUnavailable(t *testing.T) {
	dir, _ := joinedTo(t, testfix.ServerConfig())
	sock := filepath.Join(dir.Path, "agent.sock")
	start(t, dir, sock)
	var rf *localapi.Refusal
	eventually(t, "the refusal of a link", func() bool {
		_, err := localapi.NewClient(sock).UIToken(context.Background())
		return errors.As(err, &rf) && rf.Status == http.StatusConflict && strings.Contains(rf.Message, "--ui-listen off")
	})

	taken, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer taken.Close()
	dir2, _ := joinedTo(t, testfix.ServerConfig())
	sock2, logs := filepath.Join(dir2.Path, "agent.sock"), &syncBuffer{}
	startWith(t, dir2, sock2, func(d *Deps) { d.UIListen, d.Log = taken.Addr().String(), slog.New(slog.NewTextHandler(logs, nil)) })
	eventually(t, "a running agent without its UI", func() bool {
		_, err := localapi.NewClient(sock2).UIToken(context.Background())
		return errors.As(err, &rf) && strings.Contains(rf.Message, "the local UI is not running") && strings.Contains(logs.String(), "local UI not started")
	})

	dir3, _ := joinedTo(t, testfix.ServerConfig())
	host := agentfake.NewHost()
	err = Run(context.Background(), Deps{DataDir: dir3.Path, UIListen: "0.0.0.0:7443", Host: host, Routes: agentfake.NewRoutes(), Exit: &agentfake.Policy{},
		Sysctl: &agentfake.Sysctl{}, Firewall: &agentfake.Firewall{}, DNS: &agentfake.DNS{}})
	if err == nil || !strings.Contains(err.Error(), "loopback") || host.Mutations() != 0 {
		t.Fatalf("Run with the UI off loopback = %v, %d host calls", err, host.Mutations())
	}
}

// Restarting before any config is applied is refused before the engine is
// asked.
func TestRestartNeedsAConfig(t *testing.T) {
	var rf *localapi.Refusal
	if err := (&runner{}).restart(context.Background()); !errors.As(err, &rf) || rf.Status != http.StatusConflict ||
		!strings.Contains(rf.Message, "no config is applied yet") {
		t.Fatalf("restart before any config = %v", err)
	}
}

// fakeRuntime is a container runtime (names.Source): List answers the
// containers, Watch passes on every change sent to it.
type fakeRuntime struct {
	mu       sync.Mutex
	cs       []names.Container
	err      error
	watchErr error
	lists    int
	changes  chan struct{}
}

func newFakeRuntime(cs ...names.Container) *fakeRuntime {
	return &fakeRuntime{cs: cs, changes: make(chan struct{}, 4)}
}

func (f *fakeRuntime) List(context.Context) ([]names.Container, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.lists++
	return slices.Clone(f.cs), f.err
}

func (f *fakeRuntime) Watch(ctx context.Context, changed func()) error {
	f.mu.Lock()
	err := f.watchErr
	f.mu.Unlock()
	if err != nil {
		return err
	}
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-f.changes:
			changed()
		}
	}
}

func (f *fakeRuntime) set(cs []names.Container, err error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.cs, f.err = cs, err
}

func (f *fakeRuntime) listed() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.lists
}

func webContainer(ip, service string) names.Container {
	return names.Container{ID: strings.Repeat("a", 64), Name: "/shop-" + service, IP: netip.MustParseAddr(ip),
		Labels: map[string]string{names.LabelComposeService: service}}
}

// Container names (amendment A38): the agent reports what the runtime
// calls the containers on its network once the config gives it one, again
// when they change and not when they do not, sends a refused report again,
// and logs what the server does not publish.
func TestContainerNamesAreReported(t *testing.T) {
	dir, pub := newDir(t)
	cfg := testfix.ContainerSpokeConfig()
	f, srv := newFakeServer(t, pub, cfg, nil)
	enrollAs(t, dir, srv.URL, testfix.Laptop)
	rt := newFakeRuntime(webContainer("10.77.128.129", "web"))
	logs := &syncBuffer{}
	startWith(t, dir, "", func(d *Deps) {
		d.Runtime, d.NamesSettle, d.Log = rt, 10*time.Millisecond, slog.New(slog.NewTextHandler(logs, nil))
	})
	web := proto.ContainerName{Name: "web", Kind: proto.NameService, IP: "10.77.128.129"}
	eventually(t, "the first report", func() bool {
		put := f.namesPut()
		return len(put) == 1 && len(put[0].Names) == 2 && slices.Contains(put[0].Names, web)
	})

	// The same containers again: listed, not sent.
	before := rt.listed()
	rt.changes <- struct{}{}
	eventually(t, "a listing after the change", func() bool { return rt.listed() > before })
	if put := f.namesPut(); len(put) != 1 {
		t.Fatalf("an unchanged listing was sent: %v", put)
	}

	// A refused report is sent again with the next change.
	f.mu.Lock()
	f.namesRefuse = &proto.ErrorBody{Code: proto.CodeConflict, Message: "no container network"}
	f.mu.Unlock()
	rt.set([]names.Container{webContainer("10.77.128.129", "web"), webContainer("10.77.128.130", "vps-1")}, nil)
	rt.changes <- struct{}{}
	eventually(t, "the refusal logged", func() bool { return strings.Contains(logs.String(), "report the containers' names failed") })
	rt.changes <- struct{}{}
	eventually(t, "the report sent again", func() bool { return len(f.namesPut()) == 2 })
	if put := f.namesPut(); len(put[1].Names) != 4 {
		t.Fatalf("second report %v", put[1])
	}
	eventually(t, "the dropped name logged", func() bool {
		return strings.Contains(logs.String(), "container name not published") && strings.Contains(logs.String(), "node vps-1 has this name")
	})

	// A runtime that fails is logged once, however often it fails.
	rt.set(nil, errors.New("permission denied"))
	rt.changes <- struct{}{}
	eventually(t, "the failure logged", func() bool { return strings.Contains(logs.String(), "permission denied") })
	n := rt.listed()
	rt.changes <- struct{}{}
	eventually(t, "a second failed listing", func() bool { return rt.listed() > n })
	if c := strings.Count(logs.String(), "containers' names not read"); c != 1 {
		t.Fatalf("the same failure was logged %d times", c)
	}
}

// Without a container network nothing is listed or sent; a runtime whose
// events fail is logged.
func TestContainerNamesNeedAContainerNetwork(t *testing.T) {
	dir, f := joinedTo(t, testfix.ServerConfig())
	rt := newFakeRuntime(webContainer("10.77.128.129", "web"))
	rt.watchErr = errors.New("connection refused")
	logs := &syncBuffer{}
	sock := filepath.Join(dir.Path, "agent.sock")
	startWith(t, dir, sock, func(d *Deps) {
		d.Runtime, d.NamesSettle, d.Log = rt, 10*time.Millisecond, slog.New(slog.NewTextHandler(logs, nil))
	})
	eventually(t, "the config applied, and the events failure logged", func() bool {
		st, err := localapi.NewClient(sock).Status(context.Background())
		return err == nil && st.AppliedVersion == 3 && strings.Contains(logs.String(), "container runtime events stopped")
	})
	time.Sleep(50 * time.Millisecond)
	if rt.listed() != 0 || len(f.namesPut()) != 0 {
		t.Fatalf("a node without a container network listed %d times and sent %v", rt.listed(), f.namesPut())
	}
}

// The mesh view's names reach the resolver: from mesh.json at start, from
// the server, and again as soon as the server says they changed.
func TestMeshNamesReachTheResolver(t *testing.T) {
	dir, f := joinedTo(t, testfix.ServerConfig())
	cached := testfix.MeshView()
	cached.Names = testfix.MeshNames()[:1]
	if err := dir.SaveMesh(proto.LocalMesh{View: cached, FetchedAt: 1}); err != nil {
		t.Fatal(err)
	}
	r := startWith(t, dir, "", func(d *Deps) { d.MeshEvery = time.Hour })
	eventually(t, "the cached names, or the server's", func() bool { return len(r.dns.Names()) > 0 })
	view := testfix.MeshView()
	view.Names = testfix.MeshNames()
	f.setMesh(&view)
	f.nudge <- struct{}{}
	eventually(t, "the server's names after the nudge", func() bool { return slices.Equal(r.dns.Names(), view.Names) })
}

// The mesh view's containers' IPv6 addresses reach the containers' side
// (M8): when the server says the names changed, the agent applies its
// config again and the bridge answers neighbour discovery for them too.
func TestMeshNamesReachTheNeighbourEntries(t *testing.T) {
	cfg := testfix.DualSpokeConfig("")
	cfg.Exit = nil
	cfg.Containers = testfix.Containers6(proto.ModeHost)
	cfg.Sysctl = []proto.Sysctl{{Key: proto.SysctlIPForward, Value: "1"}, {Key: proto.SysctlIPv6Forward, Value: "1"}}
	dir, pub := newDir(t)
	f, srv := newFakeServer(t, pub, cfg, nil)
	enrollAs(t, dir, srv.URL, testfix.Laptop)
	nt := agentfake.NewNet()
	startWith(t, dir, "", func(d *Deps) { d.Containers, d.MeshEvery = nt, time.Hour })
	answers := func(a string) func() bool {
		return func() bool { return slices.Contains(nt.ProxiesOf(proto.ContainerBridge), netip.MustParseAddr(a)) }
	}
	eventually(t, "the bridge to answer for vps-1", answers(testfix.VPS1.MeshIP6))
	view := testfix.MeshView()
	view.Names = []proto.MeshName{{Name: "db", Node: testfix.NAS.Name, IP: "fd10:77::8001:8002"}}
	f.setMesh(&view)
	f.nudge <- struct{}{}
	eventually(t, "the bridge to answer for the NAS's container", answers("fd10:77::8001:8002"))
}

// proxyLoop applies the config on the interface again at each kick and
// logs what came of it: a pass that changed the host, and one that failed
// (the drift check retries it).
func TestProxyLoopAppliesAgain(t *testing.T) {
	k, err := wg.ParseKey(testfix.WGKey(5))
	if err != nil {
		t.Fatal(err)
	}
	var log syncBuffer
	cfg := testfix.SpokeConfig("")
	r := &runner{log: slog.New(slog.NewTextHandler(&log, nil)), proxyKick: make(chan struct{}, 1), applied: &cfg,
		engine: apply.New(apply.Deps{Host: agentfake.NewHost(), Routes: agentfake.NewRoutes(), Exit: &agentfake.Policy{},
			Sysctl: &agentfake.Sysctl{}, Iface: Interface, Key: k})}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer close(done)
		r.proxyLoop(ctx)
	}()
	defer func() {
		cancel()
		<-done
	}()
	kick(r.proxyKick)
	eventually(t, "a pass that changed the host", func() bool { return strings.Contains(log.String(), "follows the mesh's containers") })
	bad := cfg
	bad.Interface.Address = ""
	r.mu.Lock()
	r.applied = &bad
	r.mu.Unlock()
	kick(r.proxyKick)
	eventually(t, "a failed pass", func() bool { return strings.Contains(log.String(), "could not follow the mesh's containers") })
}

// The runtime socket: off reads none, a path is used as given, and auto
// takes one of the usual places or none.
func TestOpenRuntime(t *testing.T) {
	quiet := slog.New(slog.DiscardHandler)
	off := &runner{d: Deps{RuntimeSocket: RuntimeOff}, log: quiet}
	if off.openRuntime(); off.d.Runtime != nil {
		t.Fatal("off opened a runtime")
	}
	given := &runner{d: Deps{RuntimeSocket: "/run/user/1000/podman/podman.sock"}, log: quiet}
	if given.openRuntime(); given.d.Runtime == nil || given.d.Runtime.(*names.Runtime).Socket() != "/run/user/1000/podman/podman.sock" {
		t.Fatalf("a given socket = %v", given.d.Runtime)
	}
	auto := &runner{d: Deps{RuntimeSocket: RuntimeAuto}, log: quiet}
	auto.openRuntime()
	if auto.d.Runtime != nil && !slices.Contains(names.Sockets(), auto.d.Runtime.(*names.Runtime).Socket()) {
		t.Fatalf("auto picked %v", auto.d.Runtime)
	}
}

// The heartbeat says which address families the host can use (M8), from
// the apply; the server plans the node by it.
func TestReportsTheHostFamilies(t *testing.T) {
	for _, c := range []struct {
		noIPv6 bool
		want   string
	}{{false, "ipv4,ipv6"}, {true, "ipv4"}} {
		dir, f := joinedTo(t, testfix.ServerConfig())
		startWith(t, dir, filepath.Join(dir.Path, "agent.sock"), func(d *Deps) { d.Sysctl = &agentfake.Sysctl{NoIPv6: c.noIPv6} })
		eventually(t, "the families "+c.want+" in a report", func() bool {
			rep, ok := f.lastReport()
			return ok && strings.Join(rep.Families, ",") == c.want
		})
	}
}
