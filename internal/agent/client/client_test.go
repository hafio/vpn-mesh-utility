package client

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/hafio/vpn-mesh-utility/internal/proto"
	"github.com/hafio/vpn-mesh-utility/internal/testfix"
)

// verifying wraps a handler with the server's signature check for VPS1.
func verifying(t *testing.T, h http.HandlerFunc) http.HandlerFunc {
	pub, _ := testfix.IdentityKey(testfix.VPS1.Key)
	return func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		s, err := proto.ParseSigned(r.Header)
		if err == nil {
			err = s.Verify(pub, r.Method, r.URL.RequestURI(), body, testfix.Epoch)
		}
		if err != nil {
			t.Errorf("%s %s: %v", r.Method, r.URL.Path, err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if r.Header.Get("User-Agent") != "meshctl/v1.0.0" {
			t.Errorf("User-Agent %q", r.Header.Get("User-Agent"))
		}
		r.Body = io.NopCloser(strings.NewReader(string(body)))
		h(w, r)
	}
}

func newClient(t *testing.T, h http.Handler) *Client {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	_, priv := testfix.IdentityKey(testfix.VPS1.Key)
	c, err := New(Options{Server: srv.URL, NodeID: testfix.VPS1.ID, Key: priv, ConfigKey: testfix.Pinned(testfix.VPS1.ID).Key, Version: "v1.0.0",
		Now: func() time.Time { return testfix.Epoch }, Session: testSession()})
	if err != nil {
		t.Fatal(err)
	}
	return c
}

// testSession is vps-1's session in the tests.
func testSession() *Session {
	return &Session{ID: testfix.Session(testfix.VPS1.ID), Ticket: testfix.Ticket(testfix.VPS1.ID), Heartbeat: proto.Heartbeat}
}

// sign answers with v as the control plane signs it for vps-1 (amendment
// A50), bound to nonce ("" is the request's own).
func sign(w http.ResponseWriter, r *http.Request, kind, nonce string, v any) {
	b, _ := json.Marshal(v)
	b = append(b, '\n')
	if nonce == "" {
		nonce = r.Header.Get(proto.HeaderNonce)
	}
	w.Header().Set(proto.HeaderConfigSignature, proto.SignBody(testfix.ConfigKey(), kind, testfix.VPS1.ID, 0, nonce, b))
	_, _ = w.Write(b)
}

func TestConfigHonoursTheETag(t *testing.T) {
	cfg := testfix.ServerConfig()
	hash, _ := cfg.Hash()
	c := newClient(t, verifying(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("If-None-Match") == proto.ETag(hash) {
			w.WriteHeader(http.StatusNotModified)
			return
		}
		sign(w, r, proto.SignedConfig, "", cfg)
	}))
	got, signed, err := c.Config(context.Background(), "")
	if err != nil || got == nil || got.Node.Name != "vps-1" || testfix.Pinned(testfix.VPS1.ID).VerifyBody(proto.SignedConfig, signed) != nil {
		t.Fatalf("Config = %+v, %+v, %v", got, signed, err)
	}
	if got, _, err := c.Config(context.Background(), proto.ETag(hash)); got != nil || err != nil {
		t.Fatalf("Config with the current ETag = %+v, %v; want nil, nil", got, err)
	}
}

// A config or mesh view is used only as the control plane signed it for
// this request and this node (amendment A50); a definitive refusal ends
// the run only when signed for this request too.
func TestAnswersMustBeSigned(t *testing.T) {
	cfg := testfix.ServerConfig()
	for name, h := range map[string]http.HandlerFunc{
		"unsigned": func(w http.ResponseWriter, r *http.Request) { _ = json.NewEncoder(w).Encode(cfg) },
		"another request's": func(w http.ResponseWriter, r *http.Request) {
			sign(w, r, proto.SignedConfig, testfix.Nonce, cfg)
		},
		"changed on the way": func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set(proto.HeaderConfigSignature, proto.SignBody(testfix.ConfigKey(), proto.SignedConfig, testfix.VPS1.ID, 0,
				r.Header.Get(proto.HeaderNonce), []byte("{}\n")))
			_ = json.NewEncoder(w).Encode(cfg)
		},
		"a mesh view as a config": func(w http.ResponseWriter, r *http.Request) { sign(w, r, proto.SignedMesh, "", cfg) },
	} {
		c := newClient(t, verifying(t, h))
		if got, _, err := c.Config(context.Background(), ""); got != nil || err == nil || !strings.Contains(err.Error(), "not signed by") {
			t.Errorf("%s config = %+v, %v", name, got, err)
		}
	}
	view := testfix.MeshView()
	unsignedView := newClient(t, verifying(t, func(w http.ResponseWriter, r *http.Request) { _ = json.NewEncoder(w).Encode(view) }))
	if _, _, err := unsignedView.Mesh(context.Background()); err == nil || !strings.Contains(err.Error(), "not signed by") {
		t.Errorf("an unsigned mesh view = %v", err)
	}
	// Signed, but not JSON: decoded only after the signature, and refused.
	junk := newClient(t, verifying(t, func(w http.ResponseWriter, r *http.Request) {
		kind := proto.SignedConfig
		if r.URL.Path == proto.PathMesh {
			kind = proto.SignedMesh
		}
		w.Header().Set(proto.HeaderConfigSignature, proto.SignBody(testfix.ConfigKey(), kind, testfix.VPS1.ID, 0, r.Header.Get(proto.HeaderNonce), []byte("<html>")))
		_, _ = w.Write([]byte("<html>"))
	}))
	if _, _, err := junk.Config(context.Background(), ""); err == nil || !strings.Contains(err.Error(), "decode answer") {
		t.Errorf("a signed config that is not JSON = %v", err)
	}
	if _, _, err := junk.Mesh(context.Background()); err == nil || !strings.Contains(err.Error(), "decode answer") {
		t.Errorf("a signed mesh view that is not JSON = %v", err)
	}

	for name, c := range map[string]struct {
		sign       func(r *http.Request, at int64) string
		definitive bool
	}{
		"signed": {func(r *http.Request, at int64) string {
			return proto.SignRefusal(testfix.ConfigKey(), testfix.VPS1.ID, proto.CodeReplaced, r.Header.Get(proto.HeaderNonce), at)
		}, true},
		"unsigned": {func(*http.Request, int64) string { return "" }, false},
		"for another request": {func(_ *http.Request, at int64) string {
			return proto.SignRefusal(testfix.ConfigKey(), testfix.VPS1.ID, proto.CodeReplaced, testfix.Nonce, at)
		}, false},
		"for another node": {func(r *http.Request, at int64) string {
			return proto.SignRefusal(testfix.ConfigKey(), testfix.VPS2.ID, proto.CodeReplaced, r.Header.Get(proto.HeaderNonce), at)
		}, false},
	} {
		refusing := newClient(t, verifying(t, func(w http.ResponseWriter, r *http.Request) {
			if sig := c.sign(r, 42); sig != "" {
				w.Header().Set(proto.HeaderRefusalSignature, sig)
			}
			w.WriteHeader(http.StatusForbidden)
			_ = json.NewEncoder(w).Encode(proto.ErrorBody{Code: proto.CodeReplaced, Message: "replaced", ServerTime: 42})
		}))
		var ce *Error
		_, err := refusing.Status(context.Background(), proto.StatusReport{})
		if !errors.As(err, &ce) || ce.Code != proto.CodeReplaced || ce.Definitive != c.definitive {
			t.Errorf("a %s refusal = %#v, want definitive %v", name, err, c.definitive)
		}
	}
}

func TestStatusRoundTrip(t *testing.T) {
	c := newClient(t, verifying(t, func(w http.ResponseWriter, r *http.Request) {
		var rep proto.StatusReport
		if err := json.NewDecoder(r.Body).Decode(&rep); err != nil || rep.AppliedVersion != 3 || r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("report %+v, %v", rep, err)
		}
		_ = json.NewEncoder(w).Encode(proto.StatusResponse{ConfigVersion: 4, ServerTime: 9})
	}))
	resp, err := c.Status(context.Background(), proto.StatusReport{AppliedVersion: 3, AgentVersion: "v1.0.0", WGBackend: "kernel"})
	if err != nil || resp.ConfigVersion != 4 {
		t.Fatalf("Status = %+v, %v", resp, err)
	}
}

func TestPrefsRoundTrip(t *testing.T) {
	c := newClient(t, verifying(t, func(w http.ResponseWriter, r *http.Request) {
		var req proto.PrefsRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || r.URL.Path != proto.PathPrefs || req.Exit == nil || *req.Exit != "" {
			t.Errorf("prefs %s %+v, %v", r.URL.Path, req, err)
		}
		_ = json.NewEncoder(w).Encode(proto.PrefsResponse{Version: 9, ConfigVersion: 9, ConfigHash: "h"})
	}))
	off := ""
	resp, err := c.Prefs(context.Background(), proto.PrefsRequest{Exit: &off})
	if err != nil || resp.Version != 9 || resp.ConfigHash != "h" {
		t.Fatalf("Prefs = %+v, %v", resp, err)
	}
}

func TestNamesRoundTrip(t *testing.T) {
	c := newClient(t, verifying(t, func(w http.ResponseWriter, r *http.Request) {
		var req proto.NamesRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || r.Method != http.MethodPut || r.URL.Path != proto.PathNames ||
			len(req.Names) != 1 || req.Names[0].Name != "web" {
			t.Errorf("names %s %s %+v, %v", r.Method, r.URL.Path, req, err)
		}
		_ = json.NewEncoder(w).Encode(proto.NamesResponse{Stored: 1})
	}))
	resp, err := c.Names(context.Background(), proto.NamesRequest{Names: []proto.ContainerName{{Name: "web", Kind: proto.NameService, IP: "10.77.128.129"}}})
	if err != nil || resp.Stored != 1 {
		t.Fatalf("Names = %+v, %v", resp, err)
	}
	refused := newClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
		_ = json.NewEncoder(w).Encode(proto.ErrorBody{Code: proto.CodeConflict, Message: "no container network"})
	}))
	var ce *Error
	if _, err := refused.Names(context.Background(), proto.NamesRequest{}); !errors.As(err, &ce) || ce.Code != proto.CodeConflict {
		t.Fatalf("a refusal = %v", err)
	}
}

// The mesh view is signed for, and one that does not validate never
// reaches the cache or the page.
func TestMeshIsCheckedBeforeUse(t *testing.T) {
	view := testfix.MeshView()
	c := newClient(t, verifying(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != proto.PathMesh {
			t.Errorf("%s %s", r.Method, r.URL.Path)
		}
		sign(w, r, proto.SignedMesh, "", view)
	}))
	got, signed, err := c.Mesh(context.Background())
	if err != nil || len(got.Nodes) != len(view.Nodes) || got.Nodes[4] != view.Nodes[4] || len(signed.Body) == 0 {
		t.Fatalf("Mesh = %+v, %v", got, err)
	}
	view.Nodes[0].State = "<script>\n"
	if got, _, err := c.Mesh(context.Background()); err == nil || !strings.Contains(err.Error(), "nodes[0].state") || len(got.Nodes) != 0 {
		t.Fatalf("an invalid view = %+v, %v", got, err)
	}
	refused := newClient(t, verifying(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_ = json.NewEncoder(w).Encode(proto.ErrorBody{Code: proto.CodeNodeDisabled, Message: "node is disabled"})
	}))
	var ce *Error
	if _, _, err := refused.Mesh(context.Background()); !errors.As(err, &ce) || ce.Code != proto.CodeNodeDisabled {
		t.Fatalf("a refusal = %v", err)
	}
}

func TestErrorsCarryTheCode(t *testing.T) {
	c := newClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == proto.PathConfig {
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(proto.ErrorBody{Code: proto.CodeClockSkew, Message: "node clock is 2m0s behind", ServerTime: 42})
			return
		}
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte("<html>502 Bad Gateway</html>"))
	}))
	_, _, err := c.Config(context.Background(), "")
	var ce *Error
	if !errors.As(err, &ce) || ce.Code != proto.CodeClockSkew || ce.ServerTime != 42 || ce.Status != 401 || ce.Definitive {
		t.Fatalf("Config = %#v", err)
	}
	_, err = c.Status(context.Background(), proto.StatusReport{})
	if !errors.As(err, &ce) || ce.Status != 502 || !strings.Contains(ce.Message, "502 Bad Gateway") {
		t.Fatalf("a proxy error page = %#v", err)
	}
	if _, err := c.Prefs(context.Background(), proto.PrefsRequest{}); !errors.As(err, &ce) || ce.Status != 502 {
		t.Fatalf("Prefs through a broken proxy = %#v", err)
	}
	if !strings.Contains(err.Error(), "(502 internal)") {
		t.Fatalf("Error() = %q, want the status and code", err.Error())
	}

	// A long page is clipped, so a log line stays a line.
	page := newClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte(strings.Repeat("x", 500)))
	}))
	if _, _, err := page.Config(context.Background(), ""); !errors.As(err, &ce) || !strings.Contains(ce.Message, strings.Repeat("x", 120)+"...") ||
		strings.Contains(ce.Message, strings.Repeat("x", 121)) {
		t.Fatalf("a long error page = %v", err)
	}
}

// A server that is down, or answers with something that is not its JSON,
// is an error that names the call.
func TestUnreachableAndMalformedAnswers(t *testing.T) {
	down := httptest.NewServer(http.NotFoundHandler())
	down.Close()
	if _, err := Join(context.Background(), Options{Server: down.URL, Version: "v1.0.0"}, proto.JoinRequest{}); err == nil ||
		!strings.Contains(err.Error(), "POST "+proto.PathJoin) {
		t.Errorf("Join with the server down = %v", err)
	}
	_, priv := testfix.IdentityKey(testfix.VPS1.Key)
	c, err := New(Options{Server: down.URL, NodeID: testfix.VPS1.ID, Key: priv, Version: "v1.0.0", Session: testSession()})
	if err != nil {
		t.Fatal(err)
	}
	if err := c.Events(context.Background(), func(proto.ConfigEvent) {}, func() {}); err == nil || !strings.Contains(err.Error(), "open event stream") {
		t.Errorf("Events with the server down = %v", err)
	}

	junk := newClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == proto.PathEvents {
			w.Header().Set("Content-Type", "text/event-stream")
			_ = proto.WriteEvent(w, proto.Event{Name: proto.EventConfig, Data: []byte("not json")})
			return
		}
		_, _ = w.Write([]byte("<html>ok</html>"))
	}))
	if _, err := junk.Status(context.Background(), proto.StatusReport{}); err == nil || !strings.Contains(err.Error(), "decode answer") {
		t.Errorf("a 200 that is not JSON = %v", err)
	}
	called := false
	if err := junk.Events(context.Background(), func(proto.ConfigEvent) { called = true }, func() {}); err == nil || !strings.Contains(err.Error(), "malformed config event") || called {
		t.Errorf("a malformed event = %v (delivered %v)", err, called)
	}
}

func TestEventsDeliversConfigEvents(t *testing.T) {
	c := newClient(t, verifying(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_ = proto.WriteRetry(w, time.Second)
		_ = proto.WriteComment(w, "ka")
		_ = proto.WriteEvent(w, proto.Event{Name: "other", Data: []byte("{}")})
		_ = proto.WriteEvent(w, proto.Event{Name: proto.EventConfig, Data: []byte(`{"version":4,"hash":"h4"}`)})
		_ = proto.WriteEvent(w, proto.Event{Name: proto.EventMesh, Data: []byte("{}")})
		_ = proto.WriteEvent(w, proto.Event{Name: proto.EventConfig, Data: []byte(`{"version":5,"hash":"h5"}`)})
	}))
	var got []proto.ConfigEvent
	nudges := 0
	err := c.Events(context.Background(), func(ev proto.ConfigEvent) { got = append(got, ev) }, func() { nudges++ })
	if err == nil || len(got) != 2 || got[1].Version != 5 || nudges != 1 {
		t.Fatalf("Events = %v, got %+v and %d mesh events", err, got, nudges)
	}
}

func TestEventsRefusesAnythingButAStream(t *testing.T) {
	c := newClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Has("x") {
			return
		}
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte("<html>login</html>"))
	}))
	if err := c.Events(context.Background(), func(proto.ConfigEvent) {}, func() {}); err == nil || !strings.Contains(err.Error(), "proxy") {
		t.Fatalf("Events = %v", err)
	}
	disabled := newClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_ = json.NewEncoder(w).Encode(proto.ErrorBody{Code: proto.CodeNodeDisabled, Message: "disabled"})
	}))
	var ce *Error
	if err := disabled.Events(context.Background(), func(proto.ConfigEvent) {}, func() {}); !errors.As(err, &ce) || ce.Code != proto.CodeNodeDisabled {
		t.Fatalf("Events on a disabled node = %v", err)
	}
}

func TestJoinIsUnsigned(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get(proto.HeaderSignature) != "" || r.Method != http.MethodPost || r.URL.Path != proto.PathJoin {
			t.Errorf("join request %s %s %v", r.Method, r.URL.Path, r.Header)
		}
		_ = json.NewEncoder(w).Encode(proto.JoinResponse{NodeID: 3, Name: "vps-3"})
	}))
	defer srv.Close()
	resp, err := Join(context.Background(), Options{Server: srv.URL, Version: "v1.0.0"}, proto.JoinRequest{Token: testfix.Token})
	if err != nil || resp.NodeID != 3 {
		t.Fatalf("Join = %+v, %v", resp, err)
	}
	if _, err := Join(context.Background(), Options{Server: "http://mesh.example.com"}, proto.JoinRequest{}); err == nil {
		t.Fatal("plain http to a public host was accepted")
	}
}
