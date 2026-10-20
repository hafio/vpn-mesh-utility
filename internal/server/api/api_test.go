package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/hafio/vpn-mesh-utility/internal/proto"
	"github.com/hafio/vpn-mesh-utility/internal/server/store"
	"github.com/hafio/vpn-mesh-utility/internal/testfix"
)

// countingDB wraps the store and counts transactions: a request the API
// rejects early must leave the count at zero. When fail is set, that
// transaction (counting from 1) fails the way a broken disk would.
type countingDB struct {
	st   *store.Store
	n    atomic.Int64
	fail atomic.Int64
}

// errDisk is what a broken store returns; it must reach the log, never an
// answer.
var errDisk = errors.New("disk I/O error: /data/mesh.db")

func (c *countingDB) Tx(ctx context.Context, fn func(*store.Tx) error) error {
	if c.n.Add(1) == c.fail.Load() {
		return errDisk
	}
	return c.st.Tx(ctx, fn)
}

type harness struct {
	st    *store.Store
	db    *countingDB
	hub   *Hub
	rec   *Reconciler
	clock *testfix.Clock
	pub   http.Handler
	admin http.Handler
}

const testURL = "https://mesh.example.com"

func newHarness(t *testing.T) *harness {
	t.Helper()
	clock := testfix.NewClock(testfix.Epoch)
	st, err := store.Open(context.Background(), filepath.Join(t.TempDir(), "mesh.db"), store.Options{Now: clock.Now})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = st.Close() })
	db := &countingDB{st: st}
	hub := NewHub()
	rec := NewReconciler(db, hub, nil)
	d := Deps{DB: db, Hub: hub, Reconciler: rec, Now: clock.Now, ServerURL: testURL, Version: "v1.0.0", KeepAlive: 20 * time.Millisecond}
	return &harness{st: st, db: db, hub: hub, rec: rec, clock: clock, pub: New(d), admin: NewAdmin(d)}
}

func storeNode(f testfix.Node, enrolled bool) store.Node {
	n := store.Node{Name: f.Name, Role: string(f.Role), MeshIP: f.MeshIP, Endpoint: f.Endpoint, ListenPort: 51820,
		Keepalive: 25, MTU: 1420}
	if enrolled {
		n.WGPubKey, n.IdentityPubKey, n.Enrolled = testfix.WGKey(f.Key), testfix.IdentityKeyB64(f.Key), true
	}
	return n
}

// seed inserts nodes (ids follow insertion order, matching testfix ids when
// the reference nodes are given in order), reconciles, and resets the
// transaction count.
func (h *harness) seed(t *testing.T, nodes ...store.Node) {
	t.Helper()
	err := h.st.Tx(context.Background(), func(tx *store.Tx) error {
		for i := range nodes {
			if err := tx.InsertNode(&nodes[i]); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := h.rec.Apply(context.Background(), Change{Actor: "test"}, nil); err != nil {
		t.Fatal(err)
	}
	h.db.n.Store(0)
}

func (h *harness) seedServers(t *testing.T) {
	t.Helper()
	h.seed(t, storeNode(testfix.VPS1, true), storeNode(testfix.VPS2, true), storeNode(testfix.VPS3, true))
}

func (h *harness) node(t *testing.T, name string) store.Node {
	t.Helper()
	var n store.Node
	if err := h.st.Tx(context.Background(), func(tx *store.Tx) error {
		var err error
		n, err = tx.NodeByName(name)
		return err
	}); err != nil {
		t.Fatal(err)
	}
	return n
}

func (h *harness) latestVersion(t *testing.T) int64 {
	t.Helper()
	var v int64
	if err := h.st.Tx(context.Background(), func(tx *store.Tx) error {
		var err error
		v, err = tx.LatestVersion()
		return err
	}); err != nil {
		t.Fatal(err)
	}
	return v
}

// signed builds a request signed by node f at time at.
func signed(method, target string, body []byte, f testfix.Node, at time.Time) *http.Request {
	r := httptest.NewRequest(method, target, bytes.NewReader(body))
	_, priv := testfix.IdentityKey(f.Key)
	proto.Sign(priv, f.ID, method, r.URL.RequestURI(), at, body).Apply(r.Header)
	return r
}

func serve(h http.Handler, r *http.Request) (*httptest.ResponseRecorder, proto.ErrorBody) {
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	var eb proto.ErrorBody
	if w.Code >= 400 {
		_ = json.Unmarshal(w.Body.Bytes(), &eb) // a test failure below shows the raw body
	}
	return w, eb
}

func TestSignedRequestRejections(t *testing.T) {
	h := newHarness(t)
	disabled := storeNode(testfix.NAS, true)
	disabled.Disabled = true
	pending := storeNode(testfix.Laptop, false)
	h.seed(t, storeNode(testfix.VPS1, true), storeNode(testfix.VPS2, true), storeNode(testfix.VPS3, true), disabled, pending)
	now := testfix.Epoch

	cases := []struct {
		name   string
		req    func() *http.Request
		status int
		code   string
		// noDB: the rejection happens before any database access.
		noDB bool
	}{
		{"unsigned", func() *http.Request { return httptest.NewRequest(http.MethodGet, proto.PathConfig, nil) },
			401, proto.CodeUnauthorized, true},
		{"malformed nonce", func() *http.Request {
			r := signed(http.MethodGet, proto.PathConfig, nil, testfix.VPS1, now)
			r.Header.Set(proto.HeaderNonce, "x;rm -rf")
			return r
		}, 401, proto.CodeUnauthorized, true},
		{"clock 61s behind", func() *http.Request {
			return signed(http.MethodGet, proto.PathConfig, nil, testfix.VPS1, now.Add(-61*time.Second))
		}, 401, proto.CodeClockSkew, true},
		{"body over the cap", func() *http.Request {
			return signed(http.MethodPost, proto.PathStatus, bytes.Repeat([]byte("x"), proto.MaxBodyBytes+1), testfix.VPS1, now)
		}, 413, proto.CodeTooLarge, true},
		{"unknown node", func() *http.Request {
			f := testfix.VPS1
			f.ID = 99
			return signed(http.MethodGet, proto.PathConfig, nil, f, now)
		}, 401, proto.CodeUnauthorized, false},
		{"wrong key", func() *http.Request {
			f := testfix.VPS1
			f.Key = 9
			return signed(http.MethodGet, proto.PathConfig, nil, f, now)
		}, 401, proto.CodeUnauthorized, false},
		{"query tampered after signing", func() *http.Request {
			r := signed(http.MethodGet, proto.PathConfig, nil, testfix.VPS1, now)
			r.URL.RawQuery = "x=1"
			r.RequestURI = proto.PathConfig + "?x=1"
			return r
		}, 401, proto.CodeUnauthorized, false},
		{"disabled node", func() *http.Request { return signed(http.MethodGet, proto.PathConfig, nil, testfix.NAS, now) },
			403, proto.CodeNodeDisabled, false},
		{"not joined", func() *http.Request { return signed(http.MethodGet, proto.PathConfig, nil, testfix.Laptop, now) },
			401, proto.CodeUnauthorized, false},
	}
	before := h.latestVersion(t)
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			h.db.n.Store(0)
			w, eb := serve(h.pub, c.req())
			if w.Code != c.status || eb.Code != c.code {
				t.Fatalf("status %d code %q (%s), want %d %q", w.Code, eb.Code, eb.Message, c.status, c.code)
			}
			if c.noDB && h.db.n.Load() != 0 {
				t.Fatalf("the database was touched %d times for a request refused up front", h.db.n.Load())
			}
			if c.code == proto.CodeClockSkew && eb.ServerTime != now.Unix() {
				t.Fatalf("a skew error must carry the server time; got %d", eb.ServerTime)
			}
		})
	}
	if h.latestVersion(t) != before || h.hub.Subscribers() != 0 {
		t.Fatal("a rejected request changed state or opened a stream")
	}
}

func TestReplayIsRefusedBeforeTheDatabase(t *testing.T) {
	h := newHarness(t)
	h.seedServers(t)
	r := signed(http.MethodGet, proto.PathConfig, nil, testfix.VPS1, testfix.Epoch)
	if w, eb := serve(h.pub, r); w.Code != 200 {
		t.Fatalf("first request: %d %s", w.Code, eb.Message)
	}
	again := httptest.NewRequest(http.MethodGet, proto.PathConfig, nil)
	again.Header = r.Header.Clone()
	h.db.n.Store(0)
	w, eb := serve(h.pub, again)
	if w.Code != 401 || eb.Code != proto.CodeReplay || h.db.n.Load() != 0 {
		t.Fatalf("replay: %d %q, db %d", w.Code, eb.Code, h.db.n.Load())
	}
}

func joinBody(t *testing.T, token string, f testfix.Node, edit func(*proto.JoinRequest)) []byte {
	t.Helper()
	req := proto.JoinRequest{Token: token, Hostname: f.Name + ".example.com", OS: "linux", Arch: "amd64", WGBackend: "kernel",
		WGPublicKey: testfix.WGKey(f.Key), IdentityPublicKey: testfix.IdentityKeyB64(f.Key), AgentVersion: "v1.0.0"}
	if edit != nil {
		edit(&req)
	}
	b, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func (h *harness) issue(t *testing.T, name string) string {
	t.Helper()
	var tok string
	if err := h.st.Tx(context.Background(), func(tx *store.Tx) error {
		n, err := tx.NodeByName(name)
		if err != nil {
			return err
		}
		tok, _, err = tx.IssueToken(n.ID, store.TokenTTL)
		return err
	}); err != nil {
		t.Fatal(err)
	}
	h.db.n.Store(0)
	return tok
}

func postJoin(h *harness, body []byte) (*httptest.ResponseRecorder, proto.ErrorBody) {
	return serve(h.pub, httptest.NewRequest(http.MethodPost, proto.PathJoin, bytes.NewReader(body)))
}

func TestJoinEnrollsAndConsumesTheToken(t *testing.T) {
	h := newHarness(t)
	h.seed(t, storeNode(testfix.VPS1, true), storeNode(testfix.VPS2, false))
	tok := h.issue(t, "vps-2")
	events, _, _, cancel := h.hub.Subscribe(testfix.VPS1.ID)
	defer cancel()

	w, eb := postJoin(h, joinBody(t, tok, testfix.VPS2, func(r *proto.JoinRequest) { r.PublicIP = "198.51.100.7" }))
	if w.Code != 200 {
		t.Fatalf("join: %d %s", w.Code, eb.Message)
	}
	var resp proto.JoinResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	n := h.node(t, "vps-2")
	if resp.NodeID != n.ID || resp.Role != proto.RoleServer || resp.MeshIP != testfix.VPS2.MeshIP || resp.ConfigVersion == 0 {
		t.Fatalf("response %+v", resp)
	}
	if !n.Enrolled || n.WGPubKey != testfix.WGKey(testfix.VPS2.Key) || n.Hostname != "vps-2.example.com" {
		t.Fatalf("node after join %+v", n)
	}
	if n.Endpoint != "198.51.100.7:51820" {
		t.Fatalf("endpoint = %q, want the detected public IP with the listen port", n.Endpoint)
	}
	select {
	case ev := <-events:
		if ev.Version != resp.ConfigVersion {
			t.Fatalf("vps-1 event %+v, want version %d", ev, resp.ConfigVersion)
		}
	default:
		t.Fatal("vps-1 was not told about its new peer")
	}

	w, eb = postJoin(h, joinBody(t, tok, testfix.VPS2, nil))
	if w.Code != 401 || eb.Code != proto.CodeTokenInvalid || !strings.Contains(eb.Message, "already used") {
		t.Fatalf("reused token: %d %q %q", w.Code, eb.Code, eb.Message)
	}
}

// Bad input is refused before the database or the hub sees it, and the error
// never echoes the token.
func TestJoinRefusesHostileInputUpFront(t *testing.T) {
	h := newHarness(t)
	h.seed(t, storeNode(testfix.VPS1, false))
	tok := h.issue(t, "vps-1")
	cases := map[string][]byte{
		"shell in hostname":  joinBody(t, tok, testfix.VPS1, func(r *proto.JoinRequest) { r.Hostname = "a;reboot" }),
		"newline in name":    joinBody(t, tok, testfix.VPS1, func(r *proto.JoinRequest) { r.Name = "vps-1\nx" }),
		"quote in version":   joinBody(t, tok, testfix.VPS1, func(r *proto.JoinRequest) { r.AgentVersion = `v1"` }),
		"short key":          joinBody(t, tok, testfix.VPS1, func(r *proto.JoinRequest) { r.WGPublicKey = "abc" }),
		"endpoint with host": joinBody(t, tok, testfix.VPS1, func(r *proto.JoinRequest) { r.Endpoint = "evil.example:51820" }),
		"token with slash":   joinBody(t, tok[:42]+"/", testfix.VPS1, nil),
		"unknown field":      []byte(`{"token":"` + tok + `","admin":true}`),
		"not json":           []byte("token=" + tok),
	}
	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			h.db.n.Store(0)
			w, eb := postJoin(h, body)
			if w.Code != 400 || eb.Code != proto.CodeBadRequest {
				t.Fatalf("%d %q %q", w.Code, eb.Code, eb.Message)
			}
			if h.db.n.Load() != 0 {
				t.Fatal("the database was touched")
			}
			if strings.Contains(w.Body.String(), tok) {
				t.Fatalf("the error echoes the token: %s", w.Body.String())
			}
		})
	}
	// The token survived every refusal.
	if w, eb := postJoin(h, joinBody(t, tok, testfix.VPS1, nil)); w.Code != 200 {
		t.Fatalf("join after refusals: %d %s", w.Code, eb.Message)
	}
}

// A join refused after the token was consumed rolls back, so the token
// still works for the right node.
func TestRefusedJoinKeepsTheToken(t *testing.T) {
	h := newHarness(t)
	disabled := storeNode(testfix.VPS2, false)
	disabled.Disabled = true
	h.seed(t, storeNode(testfix.VPS1, true), disabled, storeNode(testfix.VPS3, false))
	if w, eb := postJoin(h, joinBody(t, h.issue(t, "vps-2"), testfix.VPS2, nil)); w.Code != 403 || eb.Code != proto.CodeNodeDisabled {
		t.Fatalf("disabled node: %d %q", w.Code, eb.Code)
	}
	tok := h.issue(t, "vps-3")
	w, eb := postJoin(h, joinBody(t, tok, testfix.VPS3, func(r *proto.JoinRequest) { r.Name = "vps-9" }))
	if w.Code != 409 || !strings.Contains(eb.Message, `belongs to node "vps-3"`) {
		t.Fatalf("name mismatch: %d %q", w.Code, eb.Message)
	}
	w, eb = postJoin(h, joinBody(t, tok, testfix.VPS1, nil)) // vps-1's keys: already registered
	if w.Code != 409 || !strings.Contains(eb.Message, "data dir copied") {
		t.Fatalf("duplicate key: %d %q", w.Code, eb.Message)
	}
	if w, eb := postJoin(h, joinBody(t, tok, testfix.VPS3, nil)); w.Code != 200 {
		t.Fatalf("the token must survive refused joins: %d %s", w.Code, eb.Message)
	}
}

func TestJoinIsRateLimitedPerClient(t *testing.T) {
	h := newHarness(t)
	for i := 0; i < joinLimit; i++ {
		if w, _ := postJoin(h, []byte("{}")); w.Code != 400 {
			t.Fatalf("attempt %d: %d", i, w.Code)
		}
	}
	w, eb := postJoin(h, []byte("{}"))
	if w.Code != 429 || eb.Code != proto.CodeRateLimited || w.Header().Get("Retry-After") == "" {
		t.Fatalf("over the limit: %d %q", w.Code, eb.Code)
	}
	h.clock.Advance(time.Minute)
	if w, _ := postJoin(h, []byte("{}")); w.Code != 400 {
		t.Fatalf("after the window: %d", w.Code)
	}
}

func TestPickEndpoint(t *testing.T) {
	cases := []struct {
		name               string
		endpoint, publicIP string
		current, observed  string
		want               string
	}{
		{"explicit wins", "203.0.113.10:4000", "198.51.100.7", "192.0.2.1:51820", "198.51.100.8", "203.0.113.10:4000"},
		{"detected next", "", "198.51.100.7", "192.0.2.1:51820", "198.51.100.8", "198.51.100.7:51821"},
		{"operator's next", "", "10.0.0.5", "192.0.2.1:51820", "198.51.100.8", "192.0.2.1:51820"},
		{"observed last", "", "", "", "198.51.100.8", "198.51.100.8:51821"},
		{"private observed is ignored", "", "", "", "10.1.2.3", ""},
	}
	for _, c := range cases {
		obs, _ := netip.ParseAddr(c.observed) // "" stays the zero Addr
		got := pickEndpoint(proto.JoinRequest{Endpoint: c.endpoint, PublicIP: c.publicIP}, c.current, obs, 51821)
		if got != c.want {
			t.Errorf("%s: got %q, want %q", c.name, got, c.want)
		}
	}
}

func TestConfigAnswers304ForTheAppliedHash(t *testing.T) {
	h := newHarness(t)
	h.seedServers(t)
	w, eb := serve(h.pub, signed(http.MethodGet, proto.PathConfig, nil, testfix.VPS1, testfix.Epoch))
	if w.Code != 200 {
		t.Fatalf("%d %s", w.Code, eb.Message)
	}
	var cfg proto.NodeConfig
	if err := json.Unmarshal(w.Body.Bytes(), &cfg); err != nil {
		t.Fatal(err)
	}
	hash, _ := cfg.Hash()
	if cfg.Node.Name != "vps-1" || len(cfg.Peers) != 2 || w.Header().Get("ETag") != proto.ETag(hash) || cfg.Version != h.node(t, "vps-1").ConfigVersion {
		t.Fatalf("config %+v etag %q", cfg, w.Header().Get("ETag"))
	}
	if hash != h.node(t, "vps-1").ConfigHash {
		t.Fatal("the served config does not match the reconciled hash")
	}
	r := signed(http.MethodGet, proto.PathConfig, nil, testfix.VPS1, testfix.Epoch)
	r.Header.Set("If-None-Match", proto.ETag(hash))
	if w, _ := serve(h.pub, r); w.Code != http.StatusNotModified || w.Body.Len() != 0 {
		t.Fatalf("If-None-Match: %d %q", w.Code, w.Body.String())
	}
}

func TestEventsStreamPushesChanges(t *testing.T) {
	h := newHarness(t)
	h.seedServers(t)
	srv := httptest.NewServer(h.pub)
	defer srv.Close()

	open := func() (*http.Response, *proto.Reader) {
		t.Helper()
		r, err := http.NewRequest(http.MethodGet, srv.URL+proto.PathEvents, nil)
		if err != nil {
			t.Fatal(err)
		}
		_, priv := testfix.IdentityKey(testfix.VPS1.Key)
		proto.Sign(priv, testfix.VPS1.ID, http.MethodGet, r.URL.RequestURI(), testfix.Epoch, nil).Apply(r.Header)
		resp, err := srv.Client().Do(r)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != 200 || resp.Header.Get("Content-Type") != "text/event-stream" || resp.Header.Get("X-Accel-Buffering") != "no" {
			t.Fatalf("stream: %d %v", resp.StatusCode, resp.Header)
		}
		return resp, proto.NewReader(resp.Body)
	}
	next := func(rd *proto.Reader) proto.ConfigEvent {
		t.Helper()
		ev, err := rd.Next()
		if err != nil || ev.Name != proto.EventConfig {
			t.Fatalf("Next = %+v, %v", ev, err)
		}
		var ce proto.ConfigEvent
		if err := json.Unmarshal(ev.Data, &ce); err != nil {
			t.Fatal(err)
		}
		return ce
	}

	resp, rd := open()
	defer resp.Body.Close()
	first := next(rd)
	if first.Hash != h.node(t, "vps-1").ConfigHash {
		t.Fatalf("the stream must open with the current version: %+v", first)
	}
	// vps-2 moves: vps-1's peer entry for it changes.
	body, _ := json.Marshal(NodeSetRequest{Endpoint: ptr("203.0.113.99:51820")})
	if w, eb := serve(h.admin, httptest.NewRequest(http.MethodPatch, AdminNodes+"/vps-2", bytes.NewReader(body))); w.Code != 200 {
		t.Fatalf("set: %d %s", w.Code, eb.Message)
	}
	got := next(rd)
	if got.Version <= first.Version || got.Hash == first.Hash {
		t.Fatalf("after the change: %+v (first %+v)", got, first)
	}

	// A second stream for the same node replaces the first.
	resp2, _ := open()
	defer resp2.Body.Close()
	if _, err := rd.Next(); !errors.Is(err, io.EOF) && !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("old stream after replacement: %v, want it closed", err)
	}
}

func ptr[T any](v T) *T { return &v }

func TestStatusRecordsTheHeartbeat(t *testing.T) {
	h := newHarness(t)
	h.seedServers(t)
	vps1 := h.node(t, "vps-1")
	rep := proto.StatusReport{AppliedVersion: vps1.ConfigVersion, AppliedHash: vps1.ConfigHash, AgentVersion: "v1.0.0", WGBackend: "kernel",
		Peers:    []proto.PeerStatus{{Name: "vps-2", LastHandshake: 100, RxBytes: 1, TxBytes: 2}, {Name: "gone", RxBytes: 9}},
		Degraded: []string{"firewall"}, Errors: []string{"firewall: firewalld is not running"},
		Rollback: &proto.Rollback{From: 3, To: 2, Reason: "no confirmation"}, DNS: proto.DNSUnavailable}
	body, _ := json.Marshal(rep)
	w, eb := serve(h.pub, signed(http.MethodPost, proto.PathStatus, body, testfix.VPS1, testfix.Epoch))
	if w.Code != 200 {
		t.Fatalf("%d %s", w.Code, eb.Message)
	}
	var resp proto.StatusResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil || resp.ConfigHash != vps1.ConfigHash || resp.ServerTime != testfix.Epoch.Unix() {
		t.Fatalf("response %+v, %v", resp, err)
	}
	n := h.node(t, "vps-1")
	if n.AppliedHash != vps1.ConfigHash || n.LastSeen != testfix.Epoch.Unix() || !strings.Contains(n.StatusJSON, "firewall") ||
		n.ObservedIP != "192.0.2.1" {
		t.Fatalf("node after heartbeat %+v", n)
	}
	var peers []store.PeerStat
	var audit []store.AuditEntry
	if err := h.st.Tx(context.Background(), func(tx *store.Tx) error {
		var err error
		if peers, err = tx.PeerStats(n.ID); err != nil {
			return err
		}
		audit, err = tx.AuditLog(10, 0)
		return err
	}); err != nil {
		t.Fatal(err)
	}
	if len(peers) != 1 || peers[0].RxBytes != 1 {
		t.Fatalf("peer rows %+v (an unknown peer name must be skipped)", peers)
	}
	if info, err := h.nodes().Show(context.Background(), "vps-1"); err != nil || info.Node.DNS != proto.DNSUnavailable ||
		len(info.Node.Degraded) != 1 {
		t.Fatalf("node show after the heartbeat = %+v, %v", info.Node, err)
	}
	if len(audit) == 0 || audit[0].Action != "node.rollback" {
		t.Fatalf("a rollback must be audited: %+v", audit)
	}

	// The agent repeats the rollback in every heartbeat until a newer config
	// is confirmed; it is audited once, and again only when it changes.
	for _, rb := range []*proto.Rollback{rep.Rollback, rep.Rollback, {From: 4, To: 2, Reason: "no confirmation"}} {
		rep.Rollback = rb
		body, _ := json.Marshal(rep)
		if w, eb := serve(h.pub, signed(http.MethodPost, proto.PathStatus, body, testfix.VPS1, testfix.Epoch)); w.Code != 200 {
			t.Fatalf("%d %s", w.Code, eb.Message)
		}
	}
	if n := h.audits(t, "node.rollback"); n != 2 {
		t.Fatalf("%d rollback audit rows, want 2 (one per distinct rollback)", n)
	}
}

// audits counts the audit rows with action.
func (h *harness) audits(t *testing.T, action string) int {
	t.Helper()
	var audit []store.AuditEntry
	if err := h.st.Tx(context.Background(), func(tx *store.Tx) error {
		var err error
		audit, err = tx.AuditLog(1000, 0)
		return err
	}); err != nil {
		t.Fatal(err)
	}
	n := 0
	for _, e := range audit {
		if e.Action == action {
			n++
		}
	}
	return n
}

// A node changes its own exit, keepalive and MTU (amendment A10) through the
// same write path as an operator: versioned, audited as the node.
func TestPrefsChangeTheNodesOwnSettings(t *testing.T) {
	h := newHarness(t)
	vps2 := storeNode(testfix.VPS2, true)
	vps2.CanExit = true
	h.seed(t, storeNode(testfix.VPS1, true), vps2, storeNode(testfix.VPS3, true), storeNode(testfix.NAS, true))
	post := func(f testfix.Node, body string) (*httptest.ResponseRecorder, proto.ErrorBody, proto.PrefsResponse) {
		t.Helper()
		w, eb := serve(h.pub, signed(http.MethodPost, proto.PathPrefs, []byte(body), f, testfix.Epoch))
		var resp proto.PrefsResponse
		if w.Code == 200 {
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				t.Fatal(err)
			}
		}
		return w, eb, resp
	}

	w, eb, resp := post(testfix.NAS, `{"exit":"vps-2","keepalive":15}`)
	nas := h.node(t, "nas")
	if w.Code != 200 || resp.Version == 0 || resp.ConfigHash != nas.ConfigHash || resp.ConfigVersion != nas.ConfigVersion {
		t.Fatalf("set exit = %d %s, %+v (node %+v)", w.Code, eb.Message, resp, nas)
	}
	if nas.ExitVia != h.node(t, "vps-2").ID || nas.Keepalive != 15 {
		t.Fatalf("after prefs %+v", nas)
	}
	var audit []store.AuditEntry
	if err := h.st.Tx(context.Background(), func(tx *store.Tx) error {
		var err error
		audit, err = tx.AuditLog(1, 0)
		return err
	}); err != nil {
		t.Fatal(err)
	}
	if len(audit) != 1 || audit[0].Action != "node.set" || audit[0].Actor != "node:4" || audit[0].Target != "nas" {
		t.Fatalf("audit %+v", audit)
	}
	if _, _, resp := post(testfix.NAS, `{"exit":"vps-2"}`); resp.Version != 0 || resp.ConfigHash != nas.ConfigHash {
		t.Fatalf("an unchanged exit made version %d", resp.Version)
	}
	if w, eb, _ := post(testfix.VPS1, `{"mtu":1380}`); w.Code != 200 || h.node(t, "vps-1").MTU != 1380 {
		t.Fatalf("a server's MTU = %d %s", w.Code, eb.Message)
	}

	for _, c := range []struct {
		name   string
		from   testfix.Node
		body   string
		status int
		code   string
	}{
		{"a server choosing an exit", testfix.VPS1, `{"exit":"vps-2"}`, 400, proto.CodeBadRequest},
		{"a server's keepalive", testfix.VPS1, `{"keepalive":10}`, 400, proto.CodeBadRequest},
		{"an exit that cannot exit", testfix.NAS, `{"exit":"vps-3"}`, 409, proto.CodeConflict},
		{"an unknown exit", testfix.NAS, `{"exit":"ghost"}`, 404, proto.CodeNotFound},
		{"an exit with a metacharacter", testfix.NAS, `{"exit":"vps-2;reboot"}`, 400, proto.CodeBadRequest},
		{"nothing to change", testfix.NAS, `{}`, 400, proto.CodeBadRequest},
		{"a field prefs does not take", testfix.NAS, `{"can_exit":true}`, 400, proto.CodeBadRequest},
	} {
		if w, eb, _ := post(c.from, c.body); w.Code != c.status || eb.Code != c.code {
			t.Errorf("%s = %d %s %q, want %d %s", c.name, w.Code, eb.Code, eb.Message, c.status, c.code)
		}
	}
	if n := h.node(t, "nas"); n.ExitVia != nas.ExitVia || n.Keepalive != 15 {
		t.Fatalf("a refused prefs request changed the node: %+v", n)
	}

	if w, eb, resp := post(testfix.NAS, `{"exit":""}`); w.Code != 200 || resp.Version == 0 || h.node(t, "nas").ExitVia != 0 {
		t.Fatalf("exit off = %d %s %+v", w.Code, eb.Message, resp)
	}
}

// The mesh view shows any joined node every node as the dashboard does, and
// nothing one node has no business knowing about another: no key, endpoint,
// hostname, observed address or error text.
func TestMeshViewCarriesNoSecrets(t *testing.T) {
	h := newHarness(t)
	nas := storeNode(testfix.NAS, true)
	nas.Hostname = "nas-private.lan"
	h.seed(t, storeNode(testfix.VPS1, true), storeNode(testfix.VPS2, true), storeNode(testfix.VPS3, true), nas, storeNode(testfix.Phone, true))
	vps1 := h.node(t, "vps-1")
	rep := proto.StatusReport{AppliedVersion: vps1.ConfigVersion, AppliedHash: vps1.ConfigHash, AgentVersion: "v1.0.0", WGBackend: "kernel",
		Peers:  []proto.PeerStatus{{Name: "phone", LastHandshake: testfix.Epoch.Unix() - 60, Endpoint: "198.51.100.9:4500"}},
		Errors: []string{"firewall: no rule on host vps-1-private"},
	}
	body, _ := json.Marshal(rep)
	if w, eb := serve(h.pub, signed(http.MethodPost, proto.PathStatus, body, testfix.VPS1, testfix.Epoch)); w.Code != 200 {
		t.Fatalf("status: %d %s", w.Code, eb.Message)
	}

	h.db.n.Store(0)
	if w, eb := serve(h.pub, httptest.NewRequest(http.MethodGet, proto.PathMesh, nil)); w.Code != 401 || eb.Code != proto.CodeUnauthorized || h.db.n.Load() != 0 {
		t.Fatalf("unsigned: %d %q, db %d", w.Code, eb.Code, h.db.n.Load())
	}
	w, eb := serve(h.pub, signed(http.MethodGet, proto.PathMesh, nil, testfix.NAS, testfix.Epoch.Add(time.Second)))
	if w.Code != 200 {
		t.Fatalf("mesh: %d %s", w.Code, eb.Message)
	}
	secrets := []string{testfix.VPS1.Endpoint, "nas-private", "vps-1-private", "198.51.100.9", "192.0.2.1"}
	for k := byte(1); k <= 6; k++ {
		secrets = append(secrets, testfix.WGKey(k), testfix.IdentityKeyB64(k))
	}
	for _, s := range secrets {
		if strings.Contains(w.Body.String(), s) {
			t.Fatalf("the mesh view carries %q:\n%s", s, w.Body)
		}
	}
	var v proto.MeshView
	if err := json.Unmarshal(w.Body.Bytes(), &v); err != nil || v.Validate() != nil || v.Version != h.latestVersion(t) ||
		v.ServerTime != testfix.Epoch.Unix() || len(v.Nodes) != 5 {
		t.Fatalf("mesh view %+v, %v", v, err)
	}
	got := map[string]proto.MeshNode{}
	for _, n := range v.Nodes {
		got[n.Name] = n
	}
	if n := got["vps-1"]; n.Health != proto.HealthOK || n.State != "in sync" || n.LastSeen != testfix.Epoch.Unix() || n.MeshIP != testfix.VPS1.MeshIP {
		t.Fatalf("vps-1 after its heartbeat: %+v", n)
	}
	if n := got["nas"]; n.Health != proto.HealthBad || n.Role != proto.RoleSpoke || n.ConfigVersion == 0 {
		t.Fatalf("nas, never heard from: %+v", n)
	}
	if n := got["phone"]; n.Health != proto.HealthWarn || n.LastSeen != testfix.Epoch.Unix()-60 {
		t.Fatalf("phone, seen by vps-1 a minute ago: %+v", n)
	}
}

func TestStatusRefusesAnInvalidReport(t *testing.T) {
	h := newHarness(t)
	h.seedServers(t)
	for name, body := range map[string]string{
		"negative counter": `{"agent_version":"v1","wg_backend":"kernel","peers":[{"name":"vps-2","rx_bytes":-1,"tx_bytes":0}]}`,
		"control char":     `{"agent_version":"v1","wg_backend":"kernel","peers":[],"errors":["a\u0007b"]}`,
		"unknown backend":  `{"agent_version":"v1","wg_backend":"magic","peers":[]}`,
	} {
		w, eb := serve(h.pub, signed(http.MethodPost, proto.PathStatus, []byte(body), testfix.VPS1, testfix.Epoch))
		if w.Code != 400 || eb.Code != proto.CodeBadRequest {
			t.Errorf("%s: %d %q", name, w.Code, eb.Code)
		}
	}
	if h.node(t, "vps-1").LastSeen != 0 {
		t.Fatal("a refused heartbeat was recorded")
	}
}

func TestHealth(t *testing.T) {
	h := newHarness(t)
	w, _ := serve(h.pub, httptest.NewRequest(http.MethodGet, proto.PathHealth, nil))
	if w.Code != 200 || w.Body.String() != "ok\n" || w.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Fatalf("%d %q %v", w.Code, w.Body.String(), w.Header())
	}
	if w, _ := serve(h.pub, httptest.NewRequest(http.MethodPost, proto.PathHealth, nil)); w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("POST /healthz = %d", w.Code)
	}
}

func TestJoinRefusesAPhone(t *testing.T) {
	h := newHarness(t)
	h.seed(t, storeNode(testfix.VPS1, true), storeNode(testfix.Phone, false))
	tok := h.issue(t, "phone")
	for i := 0; i < 2; i++ {
		// The second attempt gets the same answer: the refusal did not
		// consume the token.
		w, eb := postJoin(h, joinBody(t, tok, testfix.Phone, nil))
		if w.Code != 400 || !strings.Contains(eb.Message, "QR code") {
			t.Fatalf("attempt %d: %d %q", i+1, w.Code, eb.Message)
		}
	}
}

func TestHealthFailsWithoutTheDatabase(t *testing.T) {
	h := newHarness(t)
	if err := h.st.Close(); err != nil {
		t.Fatal(err)
	}
	w, eb := serve(h.pub, httptest.NewRequest(http.MethodGet, proto.PathHealth, nil))
	if w.Code != 500 || eb.Code != proto.CodeInternal || strings.Contains(eb.Message, "sql") {
		t.Fatalf("%d %q %q: a failed health check must not leak the cause", w.Code, eb.Code, eb.Message)
	}
}

// A store failure at any step of an agent call answers 500 without the
// cause, which stays in the server log.
func TestStoreFailuresAreInternalErrors(t *testing.T) {
	h := newHarness(t)
	h.seed(t, storeNode(testfix.VPS1, true), storeNode(testfix.VPS2, true), storeNode(testfix.VPS3, true), storeNode(testfix.NAS, true),
		storeNode(testfix.Laptop, false))
	report := `{"agent_version":"v1","wg_backend":"kernel","peers":[]}`
	for i, c := range []struct {
		name, method, path, body string
		tx                       int64
	}{
		{"config: authenticate", http.MethodGet, proto.PathConfig, "", 1},
		{"config: read", http.MethodGet, proto.PathConfig, "", 2},
		{"events: authenticate", http.MethodGet, proto.PathEvents, "", 1},
		{"events: read", http.MethodGet, proto.PathEvents, "", 2},
		{"status: record", http.MethodPost, proto.PathStatus, report, 2},
		{"prefs: authenticate", http.MethodPost, proto.PathPrefs, `{"mtu":1380}`, 1},
		{"prefs: read back", http.MethodPost, proto.PathPrefs, `{"mtu":1380}`, 3},
		{"mesh: read", http.MethodGet, proto.PathMesh, "", 2},
	} {
		h.db.n.Store(0)
		h.db.fail.Store(c.tx)
		at := testfix.Epoch.Add(time.Duration(i) * time.Second) // a fresh signature per call
		w, eb := serve(h.pub, signed(c.method, c.path, []byte(c.body), testfix.NAS, at))
		if w.Code != 500 || eb.Code != proto.CodeInternal || strings.Contains(eb.Message, "disk") {
			t.Errorf("%s: %d %s %q", c.name, w.Code, eb.Code, eb.Message)
		}
	}

	// A join that committed but cannot read the node back.
	h.db.fail.Store(0)
	tok := h.issue(t, "laptop")
	h.db.n.Store(0)
	h.db.fail.Store(2)
	if w, eb := postJoin(h, joinBody(t, tok, testfix.Laptop, nil)); w.Code != 500 || strings.Contains(eb.Message, "disk") {
		t.Fatalf("join: %d %q", w.Code, eb.Message)
	}
}

// brokenStream is an SSE client that goes away after ok sends: every flush
// after those fails.
type brokenStream struct {
	*httptest.ResponseRecorder
	ok int
}

func (b *brokenStream) FlushError() error {
	if b.ok == 0 {
		return errors.New("write: broken pipe")
	}
	b.ok--
	b.ResponseRecorder.Flush()
	return nil
}

// The stream ends at the first send that fails -- the retry hint, the
// current version or a keepalive -- rather than writing on into a dead
// connection.
func TestEventsEndAtAFailedSend(t *testing.T) {
	h := newHarness(t)
	h.seedServers(t)
	for ok, c := range []struct{ last, never string }{
		{"retry:", "event: config"},
		{"event: config", ": ka"},
		{": ka", ""},
	} {
		w := &brokenStream{ResponseRecorder: httptest.NewRecorder(), ok: ok}
		at := testfix.Epoch.Add(time.Duration(ok) * time.Second)
		done := make(chan struct{})
		go func() {
			defer close(done)
			h.pub.ServeHTTP(w, signed(http.MethodGet, proto.PathEvents, nil, testfix.VPS1, at))
		}()
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatalf("after %d good sends the stream did not end at a failed one", ok)
		}
		body := w.Body.String()
		if !strings.Contains(body, c.last) || (c.never != "" && strings.Contains(body, c.never)) {
			t.Errorf("after %d good sends the stream sent %q", ok, body)
		}
	}
}
