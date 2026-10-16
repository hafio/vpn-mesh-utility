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
	c, err := New(Options{Server: srv.URL, NodeID: testfix.VPS1.ID, Key: priv, Version: "v1.0.0", Now: func() time.Time { return testfix.Epoch }})
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestConfigHonoursTheETag(t *testing.T) {
	cfg := testfix.ServerConfig()
	hash, _ := cfg.Hash()
	c := newClient(t, verifying(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("If-None-Match") == proto.ETag(hash) {
			w.WriteHeader(http.StatusNotModified)
			return
		}
		_ = json.NewEncoder(w).Encode(cfg)
	}))
	got, err := c.Config(context.Background(), "")
	if err != nil || got == nil || got.Node.Name != "vps-1" {
		t.Fatalf("Config = %+v, %v", got, err)
	}
	if got, err := c.Config(context.Background(), proto.ETag(hash)); got != nil || err != nil {
		t.Fatalf("Config with the current ETag = %+v, %v; want nil, nil", got, err)
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

// The mesh view is signed for, and one that does not validate never
// reaches the cache or the page.
func TestMeshIsCheckedBeforeUse(t *testing.T) {
	view := testfix.MeshView()
	c := newClient(t, verifying(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != proto.PathMesh {
			t.Errorf("%s %s", r.Method, r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(view)
	}))
	got, err := c.Mesh(context.Background())
	if err != nil || len(got.Nodes) != len(view.Nodes) || got.Nodes[4] != view.Nodes[4] {
		t.Fatalf("Mesh = %+v, %v", got, err)
	}
	view.Nodes[0].State = "<script>\n"
	if got, err := c.Mesh(context.Background()); err == nil || !strings.Contains(err.Error(), "nodes[0].state") || len(got.Nodes) != 0 {
		t.Fatalf("an invalid view = %+v, %v", got, err)
	}
	refused := newClient(t, verifying(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_ = json.NewEncoder(w).Encode(proto.ErrorBody{Code: proto.CodeNodeDisabled, Message: "node is disabled"})
	}))
	var ce *Error
	if _, err := refused.Mesh(context.Background()); !errors.As(err, &ce) || ce.Code != proto.CodeNodeDisabled {
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
	_, err := c.Config(context.Background(), "")
	var ce *Error
	if !errors.As(err, &ce) || ce.Code != proto.CodeClockSkew || ce.ServerTime != 42 || ce.Status != 401 {
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
	if _, err := page.Config(context.Background(), ""); !errors.As(err, &ce) || !strings.Contains(ce.Message, strings.Repeat("x", 120)+"...") ||
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
	c, err := New(Options{Server: down.URL, NodeID: testfix.VPS1.ID, Key: priv, Version: "v1.0.0"})
	if err != nil {
		t.Fatal(err)
	}
	if err := c.Events(context.Background(), func(proto.ConfigEvent) {}); err == nil || !strings.Contains(err.Error(), "open event stream") {
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
	if err := junk.Events(context.Background(), func(proto.ConfigEvent) { called = true }); err == nil || !strings.Contains(err.Error(), "malformed config event") || called {
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
		_ = proto.WriteEvent(w, proto.Event{Name: proto.EventConfig, Data: []byte(`{"version":5,"hash":"h5"}`)})
	}))
	var got []proto.ConfigEvent
	err := c.Events(context.Background(), func(ev proto.ConfigEvent) { got = append(got, ev) })
	if err == nil || len(got) != 2 || got[1].Version != 5 {
		t.Fatalf("Events = %v, got %+v", err, got)
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
	if err := c.Events(context.Background(), func(proto.ConfigEvent) {}); err == nil || !strings.Contains(err.Error(), "proxy") {
		t.Fatalf("Events = %v", err)
	}
	disabled := newClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_ = json.NewEncoder(w).Encode(proto.ErrorBody{Code: proto.CodeNodeDisabled, Message: "disabled"})
	}))
	var ce *Error
	if err := disabled.Events(context.Background(), func(proto.ConfigEvent) {}); !errors.As(err, &ce) || ce.Code != proto.CodeNodeDisabled {
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
