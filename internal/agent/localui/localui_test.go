package localui

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"flag"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/hafio/vpn-mesh-utility/internal/agent/localapi"
	"github.com/hafio/vpn-mesh-utility/internal/agent/logbuf"
	"github.com/hafio/vpn-mesh-utility/internal/proto"
	"github.com/hafio/vpn-mesh-utility/internal/testfix"
)

var update = flag.Bool("update", false, "rewrite testdata goldens")

// fakeAgent answers the UI as a laptop's agent would. refuse, when set, is
// what Prefs, DNS and Restart answer.
type fakeAgent struct {
	mu       sync.Mutex
	st       proto.LocalStatus
	mesh     proto.LocalMesh
	prefs    []proto.PrefsRequest
	dns      []bool
	restarts int
	statuses int
	refuse   error
}

func (a *fakeAgent) status(context.Context) (proto.LocalStatus, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.statuses++
	return a.st, nil
}

func (a *fakeAgent) setPrefs(_ context.Context, req proto.PrefsRequest) (proto.PrefsResponse, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.prefs = append(a.prefs, req)
	if a.refuse != nil {
		return proto.PrefsResponse{}, a.refuse
	}
	return proto.PrefsResponse{Version: 12, ConfigVersion: 12, ConfigHash: strings.Repeat("c", 64)}, nil
}

func (a *fakeAgent) setDNS(_ context.Context, on bool) (proto.LocalStatus, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.dns = append(a.dns, on)
	if a.refuse != nil {
		return proto.LocalStatus{}, a.refuse
	}
	a.st.DNS = proto.DNSOff
	if on {
		a.st.DNS = proto.DNSOn
	}
	return a.st, nil
}

func (a *fakeAgent) meshView(context.Context) (proto.LocalMesh, error) { return a.mesh, nil }

func (a *fakeAgent) restart(context.Context) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.restarts++
	return a.refuse
}

// laptop is the status of the reference laptop, exit via vps-2.
func laptop() proto.LocalStatus {
	now := testfix.Epoch.Unix()
	return proto.LocalStatus{
		Node: proto.Self{ID: testfix.Laptop.ID, Name: testfix.Laptop.Name, Role: proto.RoleSpoke}, MeshIP: testfix.Laptop.MeshIP,
		Server: "https://mesh.example.com", AgentVersion: "v1.0.0", WGBackend: "kernel", AppliedVersion: 7,
		AppliedHash: strings.Repeat("a", 64), AppliedAt: now - 90, LastContact: now - 5, Exit: "vps-2",
		ExitChoices: []string{"vps-2", "vps-3"}, DNS: proto.DNSOn, MTU: 1420, Keepalive: 25,
		Peers: []proto.PeerStatus{
			{Name: "vps-1", LastHandshake: now - 40, RxBytes: 2048, TxBytes: 4096, Endpoint: "203.0.113.10:51820"},
			{Name: "vps-2", LastHandshake: now - 15, RxBytes: 3 << 20, TxBytes: 512, Endpoint: "203.0.113.11:51820"},
		},
	}
}

// counter is a deterministic token source that never repeats (sha256 of
// a counter), so tokens differ and pages render the same every run.
type counter struct{ n uint64 }

func (c *counter) Read(p []byte) (int, error) {
	for i := 0; i < len(p); {
		c.n++
		h := sha256.Sum256(binary.BigEndian.AppendUint64(nil, c.n))
		i += copy(p[i:], h[:])
	}
	return len(p), nil
}

type syncBuffer struct {
	mu sync.Mutex
	b  bytes.Buffer
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

type rig struct {
	t     *testing.T
	ui    *UI
	h     http.Handler
	clock *testfix.Clock
	agent *fakeAgent
	ring  *logbuf.Ring
	log   *syncBuffer
}

const origin = "http://127.0.0.1:7443"

func newRig(t *testing.T) *rig {
	t.Helper()
	r := &rig{t: t, clock: testfix.NewClock(testfix.Epoch), ring: logbuf.New(10), log: &syncBuffer{},
		agent: &fakeAgent{st: laptop(), mesh: proto.LocalMesh{View: testfix.MeshView(), FetchedAt: testfix.Epoch.Unix() - 30}}}
	ui, err := New(Config{Addr: netip.MustParseAddrPort("127.0.0.1:7443"), Now: r.clock.Now, Rand: &counter{},
		Log: slog.New(slog.NewTextHandler(r.log, nil)), Funcs: Funcs{Status: r.agent.status, Prefs: r.agent.setPrefs,
			DNS: r.agent.setDNS, Mesh: r.agent.meshView, Logs: r.ring, Restart: r.agent.restart}})
	if err != nil {
		t.Fatal(err)
	}
	r.ui, r.h = ui, ui.Handler()
	return r
}

// do sends a request the way a browser on this node would; a form is
// posted with the UI's Origin.
func (r *rig) do(method, path string, cookie *http.Cookie, form url.Values, edit func(*http.Request)) *httptest.ResponseRecorder {
	r.t.Helper()
	var body io.Reader
	if form != nil {
		body = strings.NewReader(form.Encode())
	}
	req := httptest.NewRequest(method, origin+path, body)
	if form != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("Origin", origin)
	}
	if cookie != nil {
		req.AddCookie(cookie)
	}
	if edit != nil {
		edit(req)
	}
	w := httptest.NewRecorder()
	r.h.ServeHTTP(w, req)
	return w
}

var reCSRF = regexp.MustCompile(`name="csrf" value="([^"]+)"`)

// signIn mints a link and follows it, returning the session cookie and the
// page's CSRF token.
func (r *rig) signIn() (*http.Cookie, string) {
	r.t.Helper()
	tok, err := r.ui.Token(context.Background())
	if err != nil {
		r.t.Fatal(err)
	}
	link, err := url.Parse(tok.URL)
	if err != nil {
		r.t.Fatal(err)
	}
	w := r.do(http.MethodGet, link.RequestURI(), nil, nil, nil)
	cookies := w.Result().Cookies()
	if w.Code != http.StatusSeeOther || w.Header().Get("Location") != "/" || len(cookies) != 1 {
		r.t.Fatalf("login = %d %v %v", w.Code, w.Header(), cookies)
	}
	page := r.do(http.MethodGet, "/", cookies[0], nil, nil)
	m := reCSRF.FindStringSubmatch(page.Body.String())
	if page.Code != http.StatusOK || m == nil {
		r.t.Fatalf("dashboard after login = %d\n%s", page.Code, page.Body)
	}
	return cookies[0], m[1]
}

func TestParseAddr(t *testing.T) {
	for _, s := range []string{"127.0.0.1:7443", "127.0.0.2:9000"} {
		if _, err := ParseAddr(s); err != nil {
			t.Errorf("ParseAddr(%q) = %v", s, err)
		}
	}
	for _, s := range []string{"0.0.0.0:7443", "10.77.0.1:7443", "localhost:7443", ":7443", "127.0.0.1:0", "[::1]:7443", "127.0.0.1:7443;id"} {
		if _, err := ParseAddr(s); err == nil || !strings.Contains(err.Error(), "--ui-listen") {
			t.Errorf("ParseAddr(%q) = %v, want a refusal naming --ui-listen", s, err)
		}
	}
}

// A link from `meshctl ui` signs a browser in once, within TokenTTL; the
// session cookie is HttpOnly and SameSite=Strict; neither the link nor the
// cookie is ever logged; signing out ends the session.
func TestSignIn(t *testing.T) {
	r := newRig(t)
	if w := r.do(http.MethodGet, "/", nil, nil, nil); w.Code != http.StatusUnauthorized || !strings.Contains(w.Body.String(), "meshctl ui") ||
		r.agent.statuses != 0 {
		t.Fatalf("no session = %d (status asked %d times)\n%s", w.Code, r.agent.statuses, w.Body)
	}

	tok, err := r.ui.Token(context.Background())
	if err != nil || !regexp.MustCompile(`^http://127\.0\.0\.1:7443/login\?t=[A-Za-z0-9_-]{43}$`).MatchString(tok.URL) ||
		tok.ExpiresAt != testfix.Epoch.Add(TokenTTL).Unix() || TokenTTL != 2*time.Minute {
		t.Fatalf("Token = %+v, %v (the sign-in page says 2 minutes)", tok, err)
	}
	link, _ := url.Parse(tok.URL)
	w := r.do(http.MethodGet, link.RequestURI(), nil, nil, nil)
	c := w.Result().Cookies()
	if w.Code != http.StatusSeeOther || len(c) != 1 || !c[0].HttpOnly || c[0].SameSite != http.SameSiteStrictMode || c[0].Path != "/" || c[0].Secure {
		t.Fatalf("login = %d %+v", w.Code, c)
	}
	if again := r.do(http.MethodGet, link.RequestURI(), nil, nil, nil); again.Code != http.StatusUnauthorized ||
		!strings.Contains(again.Body.String(), "used already") || len(again.Result().Cookies()) != 0 {
		t.Fatalf("the link a second time = %d\n%s", again.Code, again.Body)
	}
	late, _ := r.ui.Token(context.Background())
	r.clock.Advance(TokenTTL)
	link2, _ := url.Parse(late.URL)
	if w := r.do(http.MethodGet, link2.RequestURI(), nil, nil, nil); w.Code != http.StatusUnauthorized {
		t.Fatalf("an expired link = %d", w.Code)
	}
	for _, bad := range []string{"/login", "/login?t=x;id", "/login?t=" + strings.Repeat("A", 43)} {
		if w := r.do(http.MethodGet, bad, nil, nil, nil); w.Code != http.StatusUnauthorized {
			t.Errorf("%s = %d", bad, w.Code)
		}
	}
	if strings.Contains(r.log.String(), link.Query().Get("t")) || strings.Contains(r.log.String(), c[0].Value) {
		t.Fatalf("a token or cookie reached the log:\n%s", r.log)
	}

	cookie, csrf := r.signIn()
	w = r.do(http.MethodPost, "/logout", cookie, url.Values{"csrf": {csrf}}, nil)
	if out := w.Result().Cookies(); w.Code != http.StatusOK || len(out) != 1 || out[0].MaxAge >= 0 || !strings.Contains(w.Body.String(), "Signed out") {
		t.Fatalf("logout = %d %+v", w.Code, out)
	}
	if w := r.do(http.MethodGet, "/", cookie, nil, nil); w.Code != http.StatusUnauthorized {
		t.Fatalf("after logout = %d", w.Code)
	}
}

// A session ends after an idle hour or 12 hours in all; at most maxTokens
// links and maxSessions sessions are kept, the oldest dropped first.
func TestSessionsAndLinksAreBounded(t *testing.T) {
	r := newRig(t)
	cookie, _ := r.signIn()
	r.clock.Advance(sessionIdle + time.Second)
	if w := r.do(http.MethodGet, "/", cookie, nil, nil); w.Code != http.StatusUnauthorized {
		t.Fatalf("after an idle hour = %d", w.Code)
	}

	cookie, _ = r.signIn()
	for range 14 {
		r.clock.Advance(50 * time.Minute)
		if w := r.do(http.MethodGet, "/", cookie, nil, nil); w.Code != http.StatusOK {
			t.Fatalf("a session in use = %d", w.Code)
		}
	}
	r.clock.Advance(30 * time.Minute)
	if w := r.do(http.MethodGet, "/", cookie, nil, nil); w.Code != http.StatusUnauthorized {
		t.Fatalf("after 12 hours in use = %d", w.Code)
	}

	first, _ := r.ui.Token(context.Background())
	for range maxTokens {
		r.clock.Advance(time.Second)
		if _, err := r.ui.Token(context.Background()); err != nil {
			t.Fatal(err)
		}
	}
	link, _ := url.Parse(first.URL)
	if w := r.do(http.MethodGet, link.RequestURI(), nil, nil, nil); w.Code != http.StatusUnauthorized {
		t.Fatalf("the oldest of %d links = %d, want it dropped", maxTokens+1, w.Code)
	}

	oldest, _ := r.signIn()
	for range maxSessions {
		r.clock.Advance(time.Second)
		r.signIn()
	}
	if w := r.do(http.MethodGet, "/", oldest, nil, nil); w.Code != http.StatusUnauthorized {
		t.Fatalf("the oldest of %d sessions = %d, want it dropped", maxSessions+1, w.Code)
	}
}

// Only the UI's own names are served: a DNS-rebinding page arrives with
// its own Host and gets nothing, not even the sign-in.
func TestOnlyTheUIsOwnHostIsServed(t *testing.T) {
	r := newRig(t)
	cookie, _ := r.signIn()
	before := r.agent.statuses
	for _, host := range []string{"evil.example:7443", "127.0.0.1:8080", "10.77.2.1:7443", ""} {
		w := r.do(http.MethodGet, "/", cookie, nil, func(req *http.Request) { req.Host = host })
		if w.Code != http.StatusMisdirectedRequest || w.Header().Get("Content-Security-Policy") == "" {
			t.Errorf("Host %q = %d", host, w.Code)
		}
	}
	tok, _ := r.ui.Token(context.Background())
	link, _ := url.Parse(tok.URL)
	if w := r.do(http.MethodGet, link.RequestURI(), nil, nil, func(req *http.Request) { req.Host = "evil.example:7443" }); w.Code != http.StatusMisdirectedRequest {
		t.Fatalf("login through a foreign host = %d", w.Code)
	}
	if r.agent.statuses != before {
		t.Fatal("a foreign host reached the agent")
	}
	if w := r.do(http.MethodGet, link.RequestURI(), nil, nil, func(req *http.Request) { req.Host = "localhost:7443" }); w.Code != http.StatusSeeOther {
		t.Fatalf("login at localhost = %d (the link must still work after the refused try)", w.Code)
	}
}

// A change needs the session, the UI's Origin and the session's CSRF
// token; without any of them nothing reaches the agent.
func TestChangesNeedTheUIsOriginAndCSRF(t *testing.T) {
	r := newRig(t)
	cookie, csrf := r.signIn()
	other, _ := r.signIn()
	for name, c := range map[string]struct {
		cookie *http.Cookie
		form   url.Values
		edit   func(*http.Request)
		code   int
	}{
		"no session":           {nil, url.Values{"csrf": {csrf}}, nil, http.StatusUnauthorized},
		"no Origin":            {cookie, url.Values{"csrf": {csrf}}, func(req *http.Request) { req.Header.Del("Origin") }, http.StatusForbidden},
		"another page":         {cookie, url.Values{"csrf": {csrf}}, func(req *http.Request) { req.Header.Set("Origin", "http://evil.example") }, http.StatusForbidden},
		"an opaque origin":     {cookie, url.Values{"csrf": {csrf}}, func(req *http.Request) { req.Header.Set("Origin", "null") }, http.StatusForbidden},
		"https origin":         {cookie, url.Values{"csrf": {csrf}}, func(req *http.Request) { req.Header.Set("Origin", "https://127.0.0.1:7443") }, http.StatusForbidden},
		"no CSRF token":        {cookie, url.Values{}, nil, http.StatusForbidden},
		"another session's":    {other, url.Values{"csrf": {csrf}}, nil, http.StatusForbidden},
		"a form over the size": {cookie, url.Values{"csrf": {csrf}, "x": {strings.Repeat("x", maxForm)}}, nil, http.StatusForbidden},
	} {
		t.Run(name, func(t *testing.T) {
			if w := r.do(http.MethodPost, "/restart", c.cookie, c.form, c.edit); w.Code != c.code {
				t.Fatalf("= %d, want %d\n%s", w.Code, c.code, w.Body)
			}
		})
	}
	if r.agent.restarts != 0 || !strings.Contains(r.log.String(), "refused a change from another page") {
		t.Fatalf("%d restarts reached the agent; log:\n%s", r.agent.restarts, r.log)
	}
	if w := r.do(http.MethodPost, "/restart", cookie, url.Values{"csrf": {csrf}}, nil); w.Code != http.StatusOK ||
		!strings.Contains(w.Body.String(), "Tunnel restarted") || r.agent.restarts != 1 {
		t.Fatalf("restart = %d\n%s", w.Code, w.Body)
	}
	if w := r.do(http.MethodPost, "/restart", cookie, url.Values{"csrf": {csrf}}, func(req *http.Request) { req.Header.Set("Origin", "http://localhost:7443") }); w.Code != http.StatusOK {
		t.Fatalf("restart from localhost = %d", w.Code)
	}
}

func golden(t *testing.T, name, got string) {
	t.Helper()
	path := filepath.Join("testdata", name)
	if *update {
		if err := os.MkdirAll("testdata", 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read golden: %v (run ./scripts/dev.sh regen to create it)", err)
	}
	if got != string(want) {
		t.Fatalf("page differs from %s:\ngot:\n%s\nwant:\n%s", path, got, want)
	}
}

// The dashboard: this node, its peers, its settings forms and the mesh as
// the server last showed it, with why it is not fresh; every answer forbids
// framing, caching and script from anywhere.
func TestDashboard(t *testing.T) {
	r := newRig(t)
	r.agent.mesh.Error = "the server did not answer: dial tcp 203.0.113.10:443: connection refused"
	cookie, _ := r.signIn()
	w := r.do(http.MethodGet, "/", cookie, nil, nil)
	h := w.Header()
	if w.Code != http.StatusOK || h.Get("Content-Security-Policy") != "default-src 'self'; frame-ancestors 'none'; form-action 'self'; base-uri 'none'" ||
		h.Get("X-Frame-Options") != "DENY" || h.Get("Cache-Control") != "no-store" || h.Get("Content-Type") != "text/html; charset=utf-8" ||
		h.Get("Referrer-Policy") != "same-origin" {
		t.Fatalf("dashboard = %d %v", w.Code, h)
	}
	golden(t, "dashboard.golden.html", w.Body.String())
	if strings.Contains(w.Body.String(), "<script") {
		t.Fatal("the dashboard carries script")
	}

	// A server node: no exit or keepalive, and an agent with nothing applied.
	r.agent.st = proto.LocalStatus{Node: proto.Self{Name: "vps-1", Role: proto.RoleServer}, MeshIP: testfix.VPS1.MeshIP}
	r.agent.mesh = proto.LocalMesh{Error: "the server did not answer: connection refused"}
	body := r.do(http.MethodGet, "/", cookie, nil, nil).Body.String()
	for _, want := range []string{"none applied yet", "never fetched yet", "connection refused", `name="mtu"`} {
		if !strings.Contains(body, want) {
			t.Errorf("a fresh server's dashboard has no %q", want)
		}
	}
	for _, absent := range []string{`action="/exit"`, `name="keepalive"`, `action="/dns"`, "<h2>Peers</h2>"} {
		if strings.Contains(body, absent) {
			t.Errorf("a fresh server's dashboard has %q", absent)
		}
	}

	if w := r.do(http.MethodGet, "/static/app.css", nil, nil, nil); w.Code != http.StatusOK || !strings.HasPrefix(w.Header().Get("Content-Type"), "text/css") {
		t.Fatalf("app.css = %d %v", w.Code, w.Header())
	}
}

// Each form reaches the agent with exactly what it says, and what the
// agent or the server answers is shown; a bad value never leaves the page.
func TestForms(t *testing.T) {
	r := newRig(t)
	cookie, csrf := r.signIn()
	post := func(path string, form url.Values) (int, string) {
		t.Helper()
		form.Set("csrf", csrf)
		w := r.do(http.MethodPost, path, cookie, form, nil)
		return w.Code, w.Body.String()
	}
	if code, body := post("/exit", url.Values{"exit": {"vps-3"}}); code != http.StatusOK || !strings.Contains(body, "Exit via vps-3: the server made config v12") {
		t.Fatalf("exit vps-3 = %d\n%s", code, body)
	}
	if code, body := post("/exit", url.Values{"exit": {"off"}}); code != http.StatusOK || !strings.Contains(body, "Exit off") {
		t.Fatalf("exit off = %d\n%s", code, body)
	}
	if got := r.agent.prefs; len(got) != 2 || *got[0].Exit != "vps-3" || *got[1].Exit != "" || got[0].MTU != nil {
		t.Fatalf("prefs sent %+v", got)
	}
	if code, body := post("/settings", url.Values{"keepalive": {"15"}, "mtu": {" 1380 "}}); code != http.StatusOK || !strings.Contains(body, "Settings: the server made config v12") {
		t.Fatalf("settings = %d\n%s", code, body)
	}
	if got := r.agent.prefs[2]; *got.Keepalive != 15 || *got.MTU != 1380 || got.Exit != nil {
		t.Fatalf("settings sent %+v", got)
	}
	for _, c := range []struct {
		path string
		form url.Values
		want string
	}{
		{"/exit", url.Values{"exit": {"vps-2;id"}}, "exit"},
		{"/settings", url.Values{"mtu": {"big"}}, "mtu: must be a whole number"},
		{"/settings", url.Values{"mtu": {"9000"}}, "mtu"},
		{"/settings", url.Values{"mtu": {""}}, "nothing to change"},
		{"/dns", url.Values{"dns": {"maybe"}}, "dns"},
	} {
		if code, body := post(c.path, c.form); code != http.StatusBadRequest || !strings.Contains(body, c.want) {
			t.Errorf("%s %v = %d, want 400 naming %q", c.path, c.form, code, c.want)
		}
	}
	if len(r.agent.prefs) != 3 || len(r.agent.dns) != 0 {
		t.Fatalf("a refused value reached the agent: %+v %v", r.agent.prefs, r.agent.dns)
	}
	if code, body := post("/dns", url.Values{"dns": {"off"}}); code != http.StatusOK || !strings.Contains(body, "Split DNS off on this node (now off)") ||
		!strings.Contains(body, "Turn split DNS on") {
		t.Fatalf("dns off = %d\n%s", code, body)
	}

	r.agent.refuse = &localapi.Refusal{Status: http.StatusConflict, Code: proto.CodeConflict, Message: `exit "vps-1" is not exit-capable`}
	if code, body := post("/exit", url.Values{"exit": {"vps-1"}}); code != http.StatusConflict || !strings.Contains(body, "is not exit-capable") {
		t.Fatalf("a refused exit = %d\n%s", code, body)
	}
	r.agent.refuse = errors.New("POST /api/v1/agent/prefs: connection refused")
	if code, body := post("/settings", url.Values{"mtu": {"1400"}}); code != http.StatusBadGateway || !strings.Contains(body, "could not be asked") {
		t.Fatalf("the server unreachable = %d\n%s", code, body)
	}
	r.agent.refuse = &localapi.Refusal{Status: http.StatusConflict, Code: proto.CodeConflict, Message: "no config is applied yet"}
	if code, body := post("/restart", url.Values{}); code != http.StatusConflict || !strings.Contains(body, "no config is applied yet") {
		t.Fatalf("a refused restart = %d\n%s", code, body)
	}
	if code, _ := post("/dns", url.Values{"dns": {"on"}}); code != http.StatusConflict {
		t.Fatalf("a refused dns = %d", code)
	}
}

// The log page shows the ring as text: a line that looks like markup stays
// text.
func TestLogsPage(t *testing.T) {
	r := newRig(t)
	_, _ = io.WriteString(r.ring, "level=INFO msg=\"meshctl agent running\"\nlevel=WARN msg=\"<script>alert(1)</script>\"\n")
	cookie, _ := r.signIn()
	w := r.do(http.MethodGet, "/logs", cookie, nil, nil)
	body := w.Body.String()
	if w.Code != http.StatusOK || !strings.Contains(body, "2 lines") || !strings.Contains(body, "meshctl agent running") ||
		strings.Contains(body, "<script>") || !strings.Contains(body, "&lt;script&gt;") {
		t.Fatalf("logs = %d\n%s", w.Code, body)
	}
}

// The pages carry no script, no inline style and no event handler: the
// CSP would block them, and the UI must work without.
func TestTemplatesHaveNoScript(t *testing.T) {
	names, err := filepath.Glob("templates/*.html")
	if err != nil || len(names) == 0 {
		t.Fatalf("templates: %v %v", names, err)
	}
	bad := regexp.MustCompile(`(?i)<script|<style|\sstyle=|\son[a-z]+=|javascript:`)
	for _, n := range names {
		b, err := os.ReadFile(n)
		if err != nil {
			t.Fatal(err)
		}
		if m := bad.Find(b); m != nil {
			t.Errorf("%s has %q", n, m)
		}
	}
}

// Serve answers on a real socket and stops with its context; a port in
// use is an error that says what to do.
func TestServeAndListen(t *testing.T) {
	ln, err := Listen(netip.MustParseAddrPort("127.0.0.1:0"))
	if err != nil {
		t.Fatal(err)
	}
	addr := netip.MustParseAddrPort(ln.Addr().String())
	if _, err := Listen(addr); err == nil || !strings.Contains(err.Error(), "--ui-listen off") {
		t.Fatalf("Listen on a port in use = %v", err)
	}
	ui, err := New(Config{Addr: addr, Funcs: Funcs{Status: (&fakeAgent{}).status}})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- ui.Serve(ctx, ln) }()
	resp, err := http.Get("http://" + addr.String() + "/")
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("GET / = %d", resp.StatusCode)
	}
	cancel()
	if err := <-done; err != nil {
		t.Fatalf("Serve = %v", err)
	}
	closed, _ := net.Listen("tcp", "127.0.0.1:0")
	_ = closed.Close()
	ctx, cancel = context.WithCancel(context.Background())
	defer cancel()
	if err := ui.Serve(ctx, closed); err == nil {
		t.Fatal("Serve on a closed listener succeeded")
	}
}
