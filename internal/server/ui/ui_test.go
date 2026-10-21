package ui

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"flag"
	"image/png"
	"io/fs"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"

	"github.com/hafio/vpn-mesh-utility/deploy"
	"github.com/hafio/vpn-mesh-utility/internal/proto"
	"github.com/hafio/vpn-mesh-utility/internal/server/api"
	"github.com/hafio/vpn-mesh-utility/internal/server/auth"
	"github.com/hafio/vpn-mesh-utility/internal/server/store"
	"github.com/hafio/vpn-mesh-utility/internal/server/versions"
	"github.com/hafio/vpn-mesh-utility/internal/testfix"
)

var update = flag.Bool("update", false, "rewrite testdata goldens")

const (
	origin   = "https://mesh.example.com"
	password = "correct horse battery"
)

type rig struct {
	t     *testing.T
	h     http.Handler
	st    *store.Store
	svc   *auth.Service
	rec   *api.Reconciler
	clock *testfix.Clock
	// secret is the admin's authenticator secret once signIn has enrolled it.
	secret string
	// broken fails every mesh transaction; sign-in keeps the real store.
	broken atomic.Bool
	logs   bytes.Buffer
}

// errDisk is what a broken store returns; it must reach the log, never a page.
var errDisk = errors.New("disk I/O error: /data/mesh.db")

// faultyDB is the store as the UI and api.Nodes see it.
type faultyDB struct {
	st     *store.Store
	broken *atomic.Bool
}

func (f faultyDB) Tx(ctx context.Context, fn func(*store.Tx) error) error {
	if f.broken.Load() {
		return errDisk
	}
	return f.st.Tx(ctx, fn)
}

func newRig(t *testing.T) *rig {
	t.Helper()
	clock := testfix.NewClock(testfix.Epoch)
	st, err := store.Open(context.Background(), filepath.Join(t.TempDir(), "mesh.db"), store.Options{Now: clock.Now})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = st.Close() })
	sealer, err := auth.NewSealer(bytes.Repeat([]byte{7}, 32))
	if err != nil {
		t.Fatal(err)
	}
	svc := auth.New(auth.Config{DB: st, Sealer: sealer, Now: clock.Now, Params: auth.Params{Time: 1, MemoryKiB: 64, Threads: 1}})
	if _, err := svc.AddUser(context.Background(), "cli", "admin", password); err != nil {
		t.Fatal(err)
	}
	r := &rig{t: t, st: st, svc: svc, clock: clock}
	db := faultyDB{st: st, broken: &r.broken}
	hub := api.NewHub()
	r.rec = api.NewReconciler(db, hub, nil)
	d := api.Deps{DB: db, Hub: hub, Reconciler: r.rec, Now: clock.Now, ServerURL: origin, Version: "v1.0.0"}
	r.h, err = New(Deps{DB: db, Nodes: api.NewNodes(d), Reconciler: r.rec, Auth: svc, ServerURL: origin, Version: "v1.0.0", Now: clock.Now,
		Log: slog.New(slog.NewTextHandler(&r.logs, nil))})
	if err != nil {
		t.Fatal(err)
	}
	return r
}

// do sends a request with an optional session cookie; a non-nil form makes
// it a same-origin form POST.
func (r *rig) do(method, path, cookie string, form url.Values) *httptest.ResponseRecorder {
	r.t.Helper()
	var req *http.Request
	if form != nil {
		req = httptest.NewRequest(method, path, strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("Origin", origin)
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	if cookie != "" {
		req.AddCookie(&http.Cookie{Name: auth.CookieName, Value: cookie})
	}
	w := httptest.NewRecorder()
	r.h.ServeHTTP(w, req)
	return w
}

func cookieOf(t *testing.T, w *httptest.ResponseRecorder) string {
	t.Helper()
	for _, c := range w.Result().Cookies() {
		if c.Name == auth.CookieName && c.MaxAge >= 0 {
			return c.Value
		}
	}
	t.Fatalf("no session cookie in %v", w.Header())
	return ""
}

func code(t *testing.T, secret string, at time.Time) string {
	t.Helper()
	c, err := totp.GenerateCodeCustom(secret, at, totp.ValidateOpts{Period: 30, Digits: otp.DigitsSix, Algorithm: otp.AlgorithmSHA1})
	if err != nil {
		t.Fatal(err)
	}
	return c
}

// signIn walks the first sign-in (password, then enrolment) through the UI
// and returns the full session's cookie and CSRF token.
func (r *rig) signIn() (string, string) {
	r.t.Helper()
	w := r.do(http.MethodPost, "/login", "", url.Values{"username": {"admin"}, "password": {password}})
	if w.Code != http.StatusSeeOther || w.Header().Get("Location") != "/login/enroll" {
		r.t.Fatalf("POST /login = %d %s\n%s", w.Code, w.Header().Get("Location"), w.Body)
	}
	cookie := cookieOf(r.t, w)
	if w := r.do(http.MethodGet, "/login/enroll", cookie, nil); w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "/login/enroll.png") {
		r.t.Fatalf("GET /login/enroll = %d\n%s", w.Code, w.Body)
	}
	key, err := r.svc.BeginEnrollment(context.Background(), cookie)
	if err != nil {
		r.t.Fatal(err)
	}
	s, err := r.svc.Current(context.Background(), cookie)
	if err != nil {
		r.t.Fatal(err)
	}
	w = r.do(http.MethodPost, "/login/enroll", cookie, url.Values{"csrf": {s.CSRF}, "code": {code(r.t, key.Secret(), r.clock.Now())}})
	if w.Code != http.StatusSeeOther || w.Header().Get("Location") != "/" {
		r.t.Fatalf("POST /login/enroll = %d %s\n%s", w.Code, w.Header().Get("Location"), w.Body)
	}
	r.secret = key.Secret()
	return cookie, s.CSRF
}

func TestEveryResponseCarriesTheSecurityHeaders(t *testing.T) {
	r := newRig(t)
	cookie, _ := r.signIn()
	for _, p := range []string{"/login", "/", "/nodes", "/versions", "/audit", "/settings", "/users", "/account", "/static/app.css", "/nope"} {
		w := r.do(http.MethodGet, p, cookie, nil)
		h := w.Header()
		// same-origin keeps the browser's Origin on the UI's own posts
		// (no-referrer would make it "null", and sign-in impossible).
		if h.Get("Content-Security-Policy") != CSP || h.Get("X-Frame-Options") != "DENY" || h.Get("X-Content-Type-Options") != "nosniff" ||
			h.Get("Referrer-Policy") != "same-origin" {
			t.Errorf("%s: headers %v", p, h)
		}
	}
}

func TestSignInFlow(t *testing.T) {
	r := newRig(t)
	if w := r.do(http.MethodGet, "/login", "", nil); w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `action="/login"`) {
		t.Fatalf("the sign-in page = %d", w.Code)
	}
	cookie, csrf := r.signIn()
	if w := r.do(http.MethodGet, "/", cookie, nil); w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "Sign out") {
		t.Fatalf("dashboard = %d", w.Code)
	}
	for _, p := range []string{"/login", "/login/totp"} {
		if w := r.do(http.MethodGet, p, cookie, nil); w.Code != http.StatusSeeOther || w.Header().Get("Location") != "/" {
			t.Fatalf("%s for a signed-in user = %d %s", p, w.Code, w.Header().Get("Location"))
		}
	}
	if w := r.do(http.MethodPost, "/logout", cookie, url.Values{"csrf": {csrf}}); w.Code != http.StatusSeeOther {
		t.Fatalf("logout = %d", w.Code)
	}
	if w := r.do(http.MethodGet, "/", cookie, nil); w.Code != http.StatusSeeOther || w.Header().Get("Location") != "/login" {
		t.Fatalf("after logout = %d", w.Code)
	}

	// The second sign-in asks for a code.
	r.clock.Advance(time.Minute)
	w := r.do(http.MethodPost, "/login", "", url.Values{"username": {"admin"}, "password": {password}})
	if w.Header().Get("Location") != "/login/totp" {
		t.Fatalf("second sign-in goes to %q", w.Header().Get("Location"))
	}
	pending := cookieOf(t, w)
	s, _ := r.svc.Current(context.Background(), pending)
	if w := r.do(http.MethodGet, "/login/totp", pending, nil); w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `name="code"`) {
		t.Fatalf("the code page = %d", w.Code)
	}
	// An enrolled account cannot enrol again, and gets no QR code.
	if w := r.do(http.MethodGet, "/login/enroll", pending, nil); w.Code != http.StatusBadRequest || !strings.Contains(w.Body.String(), "already has an authenticator") {
		t.Fatalf("enrolment for an enrolled account = %d\n%s", w.Code, w.Body)
	}
	if w := r.do(http.MethodGet, "/login/enroll.png", pending, nil); w.Code != http.StatusNotFound {
		t.Fatalf("a QR code for an enrolled account = %d", w.Code)
	}
	if w := r.do(http.MethodPost, "/login/totp", pending, url.Values{"csrf": {s.CSRF}, "code": {"000000"}}); w.Code != http.StatusBadRequest ||
		!strings.Contains(w.Body.String(), "the code is wrong") {
		t.Fatalf("wrong code = %d\n%s", w.Code, w.Body)
	}
	if w := r.do(http.MethodPost, "/login/totp", pending, url.Values{"code": {"000000"}}); w.Code != http.StatusForbidden {
		t.Fatalf("a code without the CSRF token = %d", w.Code)
	}
	if w := r.do(http.MethodGet, "/", pending, nil); w.Code != http.StatusSeeOther {
		t.Fatalf("a pending session reached the dashboard: %d", w.Code)
	}
	w = r.do(http.MethodPost, "/login/totp", pending, url.Values{"csrf": {s.CSRF}, "code": {code(t, r.secret, r.clock.Now())}})
	if w.Code != http.StatusSeeOther || w.Header().Get("Location") != "/" {
		t.Fatalf("the right code = %d %s\n%s", w.Code, w.Header().Get("Location"), w.Body)
	}
	if w := r.do(http.MethodGet, "/", pending, nil); w.Code != http.StatusOK {
		t.Fatalf("the dashboard after the code = %d", w.Code)
	}
}

func TestBadSignInsAreRefusedAndThrottled(t *testing.T) {
	r := newRig(t)
	for i := 0; i < 5; i++ {
		w := r.do(http.MethodPost, "/login", "", url.Values{"username": {"admin"}, "password": {"wrong horse battery"}})
		if w.Code != http.StatusUnauthorized || !strings.Contains(w.Body.String(), "wrong username or password") ||
			!strings.Contains(w.Body.String(), `value="admin"`) {
			t.Fatalf("attempt %d = %d", i+1, w.Code)
		}
	}
	if w := r.do(http.MethodPost, "/login", "", url.Values{"username": {"admin"}, "password": {password}}); w.Code != http.StatusTooManyRequests {
		t.Fatalf("after five failures = %d", w.Code)
	}
}

func TestPagesNeedAFullSession(t *testing.T) {
	r := newRig(t)
	if w := r.do(http.MethodGet, "/nodes", "", nil); w.Code != http.StatusSeeOther || w.Header().Get("Location") != "/login" {
		t.Fatalf("GET without a session = %d", w.Code)
	}
	if w := r.do(http.MethodPost, "/nodes", "", url.Values{"name": {"x"}}); w.Code != http.StatusUnauthorized {
		t.Fatalf("POST without a session = %d", w.Code)
	}
	req := httptest.NewRequest(http.MethodGet, "/fragments/nodes", nil)
	req.Header.Set("HX-Request", "true")
	w := httptest.NewRecorder()
	r.h.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized || w.Header().Get("HX-Redirect") != "/login" {
		t.Fatalf("an htmx refresh without a session = %d %v", w.Code, w.Header())
	}
	for _, p := range []string{"/login/totp", "/login/enroll", "/login/enroll.png"} {
		if w := r.do(http.MethodGet, p, "", nil); w.Code != http.StatusSeeOther {
			t.Errorf("%s without a session = %d", p, w.Code)
		}
	}
}

// A change needs the same origin and the session's CSRF token; a refused
// change writes nothing.
func TestChangesNeedOriginAndCSRF(t *testing.T) {
	r := newRig(t)
	cookie, csrf := r.signIn()
	add := url.Values{"name": {"vps-1"}, "role": {"server"}}
	if w := r.do(http.MethodPost, "/nodes", cookie, add); w.Code != http.StatusForbidden {
		t.Fatalf("without the CSRF token = %d", w.Code)
	}
	add.Set("csrf", "wrong")
	if w := r.do(http.MethodPost, "/nodes", cookie, add); w.Code != http.StatusForbidden {
		t.Fatalf("with a wrong CSRF token = %d", w.Code)
	}
	add.Set("csrf", csrf)
	req := httptest.NewRequest(http.MethodPost, "/nodes", strings.NewReader(add.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Origin", "https://evil.example")
	req.AddCookie(&http.Cookie{Name: auth.CookieName, Value: cookie})
	w := httptest.NewRecorder()
	r.h.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("from another origin = %d", w.Code)
	}
	if w := r.do(http.MethodPost, "/login", "", nil); w.Code != http.StatusForbidden {
		t.Fatalf("a sign-in POST without Origin = %d", w.Code)
	}
	if list := r.nodes(); len(list) != 0 {
		t.Fatalf("refused requests added %d nodes", len(list))
	}
}

func (r *rig) nodes() []store.Node {
	r.t.Helper()
	var out []store.Node
	if err := r.st.Tx(context.Background(), func(tx *store.Tx) error {
		var err error
		out, err = tx.Nodes()
		return err
	}); err != nil {
		r.t.Fatal(err)
	}
	return out
}

func TestNodeLifecycle(t *testing.T) {
	r := newRig(t)
	cookie, csrf := r.signIn()
	post := func(path string, form url.Values) *httptest.ResponseRecorder {
		t.Helper()
		form.Set("csrf", csrf)
		return r.do(http.MethodPost, path, cookie, form)
	}

	w := post("/nodes", url.Values{"name": {"vps-1"}, "role": {"server"}, "endpoint": {"203.0.113.10:51820"}, "can_exit": {"on"}})
	body := w.Body.String()
	if w.Code != http.StatusOK || !strings.Contains(body, "shown once") || !strings.Contains(body, "install.sh | sudo sh -s -- --token ") {
		t.Fatalf("add = %d\n%s", w.Code, body)
	}
	if w := post("/nodes", url.Values{"name": {"Bad Name"}, "role": {"server"}}); w.Code != http.StatusBadRequest || !strings.Contains(w.Body.String(), `value="Bad Name"`) {
		t.Fatalf("a bad name = %d (and the form must keep what was typed)", w.Code)
	}
	if w := post("/nodes", url.Values{"name": {"x"}, "role": {"server"}, "mtu": {"lots"}}); w.Code != http.StatusBadRequest {
		t.Fatalf("a non-numeric MTU = %d", w.Code)
	}
	if w := r.do(http.MethodGet, "/nodes/vps-1", cookie, nil); w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "203.0.113.10:51820") {
		t.Fatalf("node page = %d", w.Code)
	}
	if w := post("/nodes/vps-1", url.Values{"endpoint": {"203.0.113.10:51820"}, "listen_port": {"51820"}, "mtu": {"1380"}}); w.Code != http.StatusOK ||
		!strings.Contains(w.Body.String(), "Saved") {
		t.Fatalf("edit = %d\n%s", w.Code, w.Body)
	}
	n := r.nodes()[0]
	if n.MTU != 1380 || n.CanExit {
		t.Fatalf("after edit %+v (an unticked box clears can_exit)", n)
	}
	for _, step := range []string{"disable", "enable"} {
		if w := post("/nodes/vps-1/"+step, url.Values{}); w.Code != http.StatusOK {
			t.Fatalf("%s = %d", step, w.Code)
		}
	}
	if w := post("/nodes/vps-1/token", url.Values{}); w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "New join token") {
		t.Fatalf("token = %d", w.Code)
	}
	if w := r.do(http.MethodGet, "/nodes/vps-1/remove", cookie, nil); w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "Remove vps-1") {
		t.Fatalf("remove page = %d", w.Code)
	}
	if w := post("/nodes/vps-1/remove", url.Values{}); w.Code != http.StatusSeeOther || len(r.nodes()) != 0 {
		t.Fatalf("remove = %d, %d nodes left", w.Code, len(r.nodes()))
	}
	if w := r.do(http.MethodGet, "/nodes/vps-1", cookie, nil); w.Code != http.StatusNotFound {
		t.Fatalf("a removed node's page = %d", w.Code)
	}
	var audit []store.AuditEntry
	if err := r.st.Tx(context.Background(), func(tx *store.Tx) error {
		var err error
		audit, err = tx.AuditLog(1, 0)
		return err
	}); err != nil {
		t.Fatal(err)
	}
	if len(audit) != 1 || audit[0].Action != "node.rm" || audit[0].Actor != "user:admin" || audit[0].UserID == 0 {
		t.Fatalf("audit %+v", audit)
	}
	if w := r.do(http.MethodGet, "/audit", cookie, nil); w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "node.rm") {
		t.Fatalf("audit page = %d", w.Code)
	}
}

// A spoke's page offers its container compose file, which downloads with the
// choices from the form; a server has none, and a bad choice comes back on
// the page.
func TestSpokeComposeDownload(t *testing.T) {
	r := newRig(t)
	r.seedServers()
	cookie, csrf := r.signIn()
	form := url.Values{"name": {"nas"}, "role": {"spoke"}, "csrf": {csrf}}
	if w := r.do(http.MethodPost, "/nodes", cookie, form); w.Code != http.StatusOK {
		t.Fatalf("add spoke = %d\n%s", w.Code, w.Body)
	}
	body := r.do(http.MethodGet, "/nodes/nas", cookie, nil).Body.String()
	if !strings.Contains(body, `action="/nodes/nas/compose"`) || !strings.Contains(body, `value="`+deploy.DefaultNASData+`"`) {
		t.Fatalf("the spoke page must offer its compose file:\n%s", body)
	}
	if body := r.do(http.MethodGet, "/nodes/vps-1", cookie, nil).Body.String(); strings.Contains(body, "/compose") {
		t.Fatalf("a server's page offers a compose file:\n%s", body)
	}
	w := r.do(http.MethodGet, "/nodes/nas/compose?backend=kernel&data=/srv/meshctl", cookie, nil)
	if w.Code != http.StatusOK || w.Header().Get("Content-Type") != "application/yaml" ||
		!strings.Contains(w.Header().Get("Content-Disposition"), `filename="compose.yml"`) ||
		!strings.Contains(w.Body.String(), "/srv/meshctl:/data") || !strings.Contains(w.Body.String(), "--wg-backend") {
		t.Fatalf("download = %d %v\n%s", w.Code, w.Header(), w.Body)
	}
	if w := r.do(http.MethodGet, "/nodes/vps-1/compose", cookie, nil); w.Code != http.StatusConflict || !strings.Contains(w.Body.String(), "a server in host mode") {
		t.Fatalf("a server's compose file = %d\n%s", w.Code, w.Body)
	}
	if w := r.do(http.MethodGet, "/nodes/nas/compose?data=/srv/x%0Aprivileged:%20true", cookie, nil); w.Code != http.StatusBadRequest ||
		!strings.Contains(w.Body.String(), "absolute path") {
		t.Fatalf("a hostile directory = %d\n%s", w.Code, w.Body)
	}
	if w := r.do(http.MethodGet, "/nodes/nas/compose", "", nil); w.Code == http.StatusOK {
		t.Fatal("the compose file downloaded without a session")
	}
}

// A container-mode node from the add form (amendment A37): its page names
// the mode and the subnet, offers the compose file per runtime and leaves
// out what container mode cannot do (an exit); a host-mode node's edit
// form puts its containers on the mesh and takes them off.
func TestContainerModePages(t *testing.T) {
	r := newRig(t)
	r.seedServers()
	cookie, csrf := r.signIn()
	form := url.Values{"name": {"box"}, "role": {"spoke"}, "mode": {"container"}, "csrf": {csrf}}
	if w := r.do(http.MethodPost, "/nodes", cookie, form); w.Code != http.StatusOK {
		t.Fatalf("add a container-mode spoke = %d\n%s", w.Code, w.Body)
	}
	body := r.do(http.MethodGet, "/nodes/box", cookie, nil).Body.String()
	for _, want := range []string{"container: only its containers are on the mesh", testfix.ContainerSubnet, `name="runtime"`,
		`<option value="swarm">`, `action="/nodes/box/compose"`, `name="names" value="off"`} {
		if !strings.Contains(body, want) {
			t.Errorf("the container-mode page lacks %q", want)
		}
	}
	if strings.Contains(body, `name="exit"`) || strings.Contains(body, `name="containers"`) {
		t.Errorf("the container-mode page offers an exit or a containers switch:\n%s", body)
	}
	w := r.do(http.MethodGet, "/nodes/box/compose?runtime=podman&backend=auto&data=", cookie, nil)
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "podman network create --internal --disable-dns") ||
		!strings.Contains(w.Body.String(), "/run/podman/podman.sock:/var/run/docker.sock:ro") {
		t.Fatalf("podman download = %d\n%s", w.Code, w.Body)
	}
	if w := r.do(http.MethodGet, "/nodes/box/compose?names=off", cookie, nil); w.Code != http.StatusOK ||
		strings.Contains(w.Body.String(), ".sock") || !strings.Contains(w.Body.String(), "--runtime-socket") {
		t.Fatalf("a download without names = %d\n%s", w.Code, w.Body)
	}
	if w := r.do(http.MethodGet, "/nodes/box/compose?runtime=lxc", cookie, nil); w.Code != http.StatusBadRequest {
		t.Fatalf("an unknown runtime = %d", w.Code)
	}
	// The DNS page lists what box's runtime reported (amendment A38): how a
	// published name resolves, and why one is not.
	if body := r.do(http.MethodGet, "/dns", cookie, nil).Body.String(); !strings.Contains(body, "No container has reported a name yet") {
		t.Fatalf("the DNS page without names:\n%s", body)
	}
	if err := r.st.Tx(context.Background(), func(tx *store.Tx) error {
		box, err := tx.NodeByName("box")
		if err != nil {
			return err
		}
		_, err = tx.SetContainerNames(box.ID, []store.ContainerName{{Name: "web", IP: "10.77.128.129", Kind: proto.NameService},
			{Name: "vps-1", IP: "10.77.128.129", Kind: proto.NameAlias}})
		return err
	}); err != nil {
		t.Fatal(err)
	}
	body = r.do(http.MethodGet, "/dns", cookie, nil).Body.String()
	for _, want := range []string{"<code>web.mesh</code>, <code>web.box.mesh</code>", "not published: node vps-1 has this name"} {
		if !strings.Contains(body, want) {
			t.Errorf("the DNS page lacks %q:\n%s", want, body)
		}
	}

	form = url.Values{"name": {"pc"}, "role": {"spoke"}, "containers": {"on"}, "csrf": {csrf}}
	if w := r.do(http.MethodPost, "/nodes", cookie, form); w.Code != http.StatusOK {
		t.Fatalf("add a host-mode spoke with containers = %d\n%s", w.Code, w.Body)
	}
	if body := r.do(http.MethodGet, "/nodes/pc", cookie, nil).Body.String(); !strings.Contains(body, "10.77.129.0/24") ||
		!strings.Contains(body, `name="containers" checked`) {
		t.Fatalf("the host-mode page must show its containers:\n%s", body)
	}
	// A host-mode node with containers shows its mesh network's commands.
	if body := r.do(http.MethodGet, "/nodes/pc", cookie, nil).Body.String(); !strings.Contains(body, "Containers on the mesh") ||
		!strings.Contains(body, "mode=unmanaged") || !strings.Contains(body, "inhibit_ipv4=true") || !strings.Contains(body, "dns: [10.77.129.1]") ||
		!strings.Contains(body, "A Swarm service") || !strings.Contains(body, "dns: [10.77.0.1, 10.77.0.2]") {
		t.Fatalf("the host-mode page must show its network's commands:\n%s", body)
	}
	edit := url.Values{"mtu": {"1420"}, "keepalive": {"25"}, "csrf": {csrf}} // containers unticked
	if w := r.do(http.MethodPost, "/nodes/pc", cookie, edit); w.Code != http.StatusOK || strings.Contains(w.Body.String(), "10.77.129.0/24") {
		t.Fatalf("untick containers = %d\n%s", w.Code, w.Body)
	}
	form = url.Values{"name": {"ph"}, "role": {"phone"}, "mode": {"container"}, "csrf": {csrf}}
	if w := r.do(http.MethodPost, "/nodes", cookie, form); w.Code != http.StatusBadRequest || !strings.Contains(w.Body.String(), "no agent") {
		t.Fatalf("a container-mode phone = %d\n%s", w.Code, w.Body)
	}
}

// The DNS page lists the zone and manages the custom records; the firewall
// page shows each server's rules with their source and manages the custom
// rules. A refusal comes back on the page with the form as it was.
func TestDNSPage(t *testing.T) {
	r := newRig(t)
	r.seedServers()
	cookie, csrf := r.signIn()
	post := func(path string, form url.Values) *httptest.ResponseRecorder {
		t.Helper()
		form.Set("csrf", csrf)
		return r.do(http.MethodPost, path, cookie, form)
	}
	body := r.do(http.MethodGet, "/dns", cookie, nil).Body.String()
	for _, want := range []string{"<td>vps-1.mesh</td><td>10.77.0.1</td>", "automatic", "1.1.1.1, 9.9.9.9", "any other container uses its own resolver"} {
		if !strings.Contains(body, want) {
			t.Fatalf("the DNS page lacks %q:\n%s", want, body)
		}
	}
	if w := post("/dns", url.Values{"name": {"files"}, "node": {"vps-2"}}); w.Code != http.StatusOK ||
		!strings.Contains(w.Body.String(), "Record files added; config v") || !strings.Contains(w.Body.String(), `action="/dns/files/remove"`) {
		t.Fatalf("add record = %d\n%s", w.Code, w.Body)
	}
	if w := post("/dns", url.Values{"name": {"vps-1"}, "ip": {"10.77.1.5"}}); w.Code != http.StatusConflict ||
		!strings.Contains(w.Body.String(), "has the name of a node") || !strings.Contains(w.Body.String(), `value="10.77.1.5"`) {
		t.Fatalf("a record named like a node = %d\n%s", w.Code, w.Body)
	}
	if w := r.do(http.MethodPost, "/dns", cookie, url.Values{"name": {"x"}, "ip": {"10.77.1.5"}}); w.Code != http.StatusForbidden {
		t.Fatalf("a record without the CSRF token = %d", w.Code)
	}
	if w := post("/dns/files/remove", url.Values{}); w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "Record files removed") {
		t.Fatalf("remove record = %d\n%s", w.Code, w.Body)
	}

}

// A spoke's edit form sets its hub, exit and keepalive; its page offers the
// servers as hubs and the can_exit ones as exits. Refused edits change nothing.
func TestEditSpoke(t *testing.T) {
	r := newRig(t)
	r.seedServers()
	cookie, csrf := r.signIn()
	post := func(path string, form url.Values) *httptest.ResponseRecorder {
		t.Helper()
		form.Set("csrf", csrf)
		return r.do(http.MethodPost, path, cookie, form)
	}
	if w := r.do(http.MethodGet, "/nodes/vps-1", cookie, nil); w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "listen_port") {
		t.Fatalf("an enrolled server's page shows its computed config: %d\n%s", w.Code, w.Body)
	}
	if w := post("/nodes/vps-2", url.Values{"endpoint": {testfix.VPS2.Endpoint}, "listen_port": {"51820"}, "mtu": {"1420"}, "can_exit": {"on"}}); w.Code != http.StatusOK {
		t.Fatalf("make vps-2 an exit = %d\n%s", w.Code, w.Body)
	}
	if w := post("/nodes", url.Values{"name": {"laptop"}, "role": {"spoke"}}); w.Code != http.StatusOK {
		t.Fatalf("add spoke = %d\n%s", w.Code, w.Body)
	}
	body := r.do(http.MethodGet, "/nodes/laptop", cookie, nil).Body.String()
	if !strings.Contains(body, `<option value="vps-1"`) || strings.Count(body, `<option value="vps-2"`) != 2 {
		t.Fatalf("the spoke page must offer both servers as hubs and vps-2 as an exit:\n%s", body)
	}
	if w := post("/nodes/laptop", url.Values{"hub": {"vps-1"}, "exit": {"vps-2"}, "keepalive": {"15"}, "mtu": {"1380"}}); w.Code != http.StatusOK ||
		!strings.Contains(w.Body.String(), "Saved") {
		t.Fatalf("edit spoke = %d\n%s", w.Code, w.Body)
	}
	byName := map[string]store.Node{}
	for _, n := range r.nodes() {
		byName[n.Name] = n
	}
	l := byName["laptop"]
	if l.PrimaryHubID != byName["vps-1"].ID || l.ExitVia != byName["vps-2"].ID || l.Keepalive != 15 || l.MTU != 1380 {
		t.Fatalf("after edit %+v", l)
	}
	for name, c := range map[string]struct {
		form url.Values
		want int
	}{
		"keepalive not a number": {url.Values{"keepalive": {"often"}, "mtu": {"1380"}}, http.StatusBadRequest},
		"keepalive too long":     {url.Values{"keepalive": {"9999"}, "mtu": {"1380"}}, http.StatusBadRequest},
		"hub not a node":         {url.Values{"hub": {"ghost"}, "keepalive": {"15"}, "mtu": {"1380"}}, http.StatusNotFound},
	} {
		if w := post("/nodes/laptop", c.form); w.Code != c.want || !strings.Contains(w.Body.String(), "<h1>laptop</h1>") {
			t.Errorf("%s = %d, want %d on the node page\n%s", name, w.Code, c.want, w.Body)
		}
	}
	if n := r.nodes(); len(n) != 3 || n[2].Keepalive != 15 || n[2].PrimaryHubID != byName["vps-1"].ID {
		t.Fatalf("a refused edit changed the spoke: %+v", n)
	}
	// A new mesh IP from the form (amendment A41), and what it does not reach.
	if w := post("/nodes/laptop", url.Values{"mesh_ip": {"10.77.2.9"}, "hub": {"vps-1"}, "keepalive": {"15"}, "mtu": {"1380"}}); w.Code != http.StatusOK ||
		!strings.Contains(w.Body.String(), "laptop moves from 10.77.2.1 to 10.77.2.9") || !strings.Contains(w.Body.String(), `name="mesh_ip" value="10.77.2.9"`) {
		t.Fatalf("a new mesh IP = %d\n%s", w.Code, w.Body)
	}
	if w := post("/nodes/laptop", url.Values{"mesh_ip": {"10.77.0.1"}, "hub": {"vps-1"}, "keepalive": {"15"}, "mtu": {"1380"}}); w.Code != http.StatusBadRequest ||
		!strings.Contains(w.Body.String(), "already assigned") {
		t.Fatalf("a taken mesh IP = %d\n%s", w.Code, w.Body)
	}
	if w := post("/nodes/laptop/disable", url.Values{}); w.Code != http.StatusOK {
		t.Fatalf("disable = %d", w.Code)
	}
	if w := post("/nodes/laptop/token", url.Values{}); w.Code != http.StatusConflict || !strings.Contains(w.Body.String(), "enable it first") {
		t.Fatalf("a token for a disabled node = %d\n%s", w.Code, w.Body)
	}
	if w := post("/nodes/ghost/disable", url.Values{}); w.Code != http.StatusNotFound {
		t.Fatalf("disabling an unknown node = %d", w.Code)
	}
}

// A failing store gives a page that names no detail; the cause goes to the log.
func TestStoreFailuresAreLoggedNotShown(t *testing.T) {
	r := newRig(t)
	r.seedServers()
	cookie, csrf := r.signIn()
	r.broken.Store(true)
	for _, c := range []struct{ method, path string }{
		{http.MethodGet, "/"}, {http.MethodGet, "/fragments/nodes"}, {http.MethodGet, "/nodes"}, {http.MethodGet, "/nodes/vps-1"},
		{http.MethodGet, "/nodes/vps-1/remove"}, {http.MethodGet, "/versions"}, {http.MethodGet, "/versions/1"},
		{http.MethodGet, "/audit"}, {http.MethodGet, "/settings"}, {http.MethodGet, "/export"},
		{http.MethodPost, "/nodes/vps-1"}, {http.MethodPost, "/nodes/vps-1/remove"}, {http.MethodPost, "/versions/1/rollback"},
		{http.MethodGet, "/dns"}, {http.MethodGet, "/firewall"}, {http.MethodPost, "/dns/files/remove"},
		{http.MethodGet, "/firewall/zones"}, {http.MethodGet, "/firewall/objects"}, {http.MethodGet, "/firewall/reach"},
		{http.MethodPost, "/firewall/policies/allow-all/remove"}, {http.MethodPost, "/firewall/zones"},
		{http.MethodPost, "/nodes/vps-1/phone"},
	} {
		var form url.Values
		if c.method == http.MethodPost {
			form = url.Values{"csrf": {csrf}}
		}
		r.logs.Reset()
		w := r.do(c.method, c.path, cookie, form)
		if w.Code != http.StatusInternalServerError || !strings.Contains(w.Body.String(), "the server log has the details") ||
			strings.Contains(w.Body.String(), "mesh.db") {
			t.Errorf("%s %s = %d\n%s", c.method, c.path, w.Code, w.Body)
		}
		if !strings.Contains(r.logs.String(), "mesh.db") {
			t.Errorf("%s %s: the cause was not logged: %q", c.method, c.path, r.logs.String())
		}
	}
	r.broken.Store(false)
	if len(r.nodes()) != 2 {
		t.Fatal("a failed remove removed a node")
	}
}

// seedServers inserts two enrolled servers and reconciles: version 1.
func (r *rig) seedServers() {
	r.t.Helper()
	err := r.st.Tx(context.Background(), func(tx *store.Tx) error {
		for _, f := range []testfix.Node{testfix.VPS1, testfix.VPS2} {
			n := store.Node{Name: f.Name, Role: "server", MeshIP: f.MeshIP, Endpoint: f.Endpoint, ListenPort: 51820, Keepalive: 25,
				MTU: 1420, Enrolled: true, WGPubKey: testfix.WGKey(f.Key), IdentityPubKey: testfix.IdentityKeyB64(f.Key)}
			if err := tx.InsertNode(&n); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		r.t.Fatal(err)
	}
	if _, err := r.rec.Apply(context.Background(), api.Change{Actor: "test"}, nil); err != nil {
		r.t.Fatal(err)
	}
}

func TestRollbackThatAddsNodesNeedsConfirmation(t *testing.T) {
	r := newRig(t)
	r.seedServers()
	cookie, csrf := r.signIn()
	if w := r.do(http.MethodPost, "/nodes/vps-2/remove", cookie, url.Values{"csrf": {csrf}}); w.Code != http.StatusSeeOther {
		t.Fatalf("remove = %d", w.Code)
	}
	w := r.do(http.MethodGet, "/versions/1", cookie, nil)
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "node vps-2 added") || !strings.Contains(w.Body.String(), `name="confirm"`) {
		t.Fatalf("version page = %d\n%s", w.Code, w.Body)
	}
	if w := r.do(http.MethodPost, "/versions/1/rollback", cookie, url.Values{"csrf": {csrf}}); w.Code != http.StatusConflict || len(r.nodes()) != 1 {
		t.Fatalf("an unconfirmed rollback = %d, %d nodes", w.Code, len(r.nodes()))
	}
	w = r.do(http.MethodPost, "/versions/1/rollback", cookie, url.Values{"csrf": {csrf}, "confirm": {"yes"}})
	if w.Code != http.StatusSeeOther || w.Header().Get("Location") != "/versions/3" || len(r.nodes()) != 2 {
		t.Fatalf("a confirmed rollback = %d %s, %d nodes", w.Code, w.Header().Get("Location"), len(r.nodes()))
	}
	if w := r.do(http.MethodGet, "/versions", cookie, nil); w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "restores v1") {
		t.Fatalf("versions page = %d\n%s", w.Code, w.Body)
	}
	if w := r.do(http.MethodGet, "/versions/abc", cookie, nil); w.Code != http.StatusBadRequest {
		t.Fatalf("a bad version number = %d", w.Code)
	}
	if w := r.do(http.MethodGet, "/versions/99", cookie, nil); w.Code != http.StatusNotFound {
		t.Fatalf("an unknown version = %d", w.Code)
	}
	w = r.do(http.MethodPost, "/versions/3/rollback", cookie, url.Values{"csrf": {csrf}})
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "already matches") {
		t.Fatalf("a rollback to the current version = %d\n%s", w.Code, w.Body)
	}
}

func TestSettings(t *testing.T) {
	r := newRig(t)
	cookie, csrf := r.signIn()
	form := func(kv ...string) url.Values {
		v := url.Values{"csrf": {csrf}, "default_mtu": {"1420"}, "upstream_dns": {"1.1.1.1, 9.9.9.9"}}
		for i := 0; i < len(kv); i += 2 {
			v.Set(kv[i], kv[i+1])
		}
		return v
	}
	for name, f := range map[string]url.Values{
		"mtu out of range":  form("default_mtu", "100"),
		"upstream hostname": form("upstream_dns", "dns.example"),
		"hub not a server":  form("default_hub", "ghost"),
		"cidr with host":    form("mesh_cidr", "10.88.0.1/16"),
		"cidr too large":    form("mesh_cidr", "10.0.0.0/8"),
		"no upstream":       form("upstream_dns", " , "),
		"zone":              form("dns_zone", "office.local"),
		"extra domain":      form("dns_extra_domains", "lan, 10.in-addr.arpa"),
		"server's domain":   form("dns_extra_domains", "example.com"),
	} {
		if w := r.do(http.MethodPost, "/settings", cookie, f); w.Code != http.StatusBadRequest {
			t.Errorf("%s = %d\n%s", name, w.Code, w.Body)
		}
	}
	// The CIDR may change while the mesh is empty.
	for _, cidr := range []string{"10.88.0.0/16", "10.77.0.0/16"} {
		if w := r.do(http.MethodPost, "/settings", cookie, form("mesh_cidr", cidr, "default_mtu", "1400")); w.Code != http.StatusOK {
			t.Fatalf("save %s = %d\n%s", cidr, w.Code, w.Body)
		}
	}
	r.seedServers()
	w := r.do(http.MethodGet, "/settings", cookie, nil)
	if !strings.Contains(w.Body.String(), `value="10.77.0.0/16" readonly`) || !strings.Contains(w.Body.String(), `value="1400"`) {
		t.Fatalf("settings page after save:\n%s", w.Body)
	}
	if w := r.do(http.MethodPost, "/settings", cookie, form("mesh_cidr", "10.99.0.0/16")); w.Code != http.StatusBadRequest {
		t.Fatalf("changing the CIDR of a populated mesh = %d", w.Code)
	}
	if w := r.do(http.MethodPost, "/settings", cookie, form("default_hub", "vps-2")); w.Code != http.StatusOK {
		t.Fatalf("default hub = %d", w.Code)
	}
	// The zone renames every node's name; the DNS page shows the new one.
	if !strings.Contains(r.do(http.MethodGet, "/settings", cookie, nil).Body.String(), `name="dns_zone" value="mesh"`) {
		t.Fatal("the settings page does not show the default zone")
	}
	if w := r.do(http.MethodPost, "/settings", cookie, form("dns_zone", "home.arpa")); w.Code != http.StatusOK {
		t.Fatalf("dns zone = %d\n%s", w.Code, w.Body)
	}
	if body := r.do(http.MethodGet, "/dns", cookie, nil).Body.String(); !strings.Contains(body, "vps-1.home.arpa") {
		t.Fatalf("the DNS page after the zone change:\n%s", body)
	}
	// Extra domains (amendment A39): saved sorted, shown on both pages; one
	// around the zone is the plan's to refuse; an empty field clears them,
	// and a form without the field keeps them.
	if w := r.do(http.MethodPost, "/settings", cookie, form("dns_extra_domains", "home.arpa.lab, Example.Lan")); w.Code != http.StatusOK ||
		!strings.Contains(w.Body.String(), `value="example.lan, home.arpa.lab"`) {
		t.Fatalf("extra domains = %d\n%s", w.Code, w.Body)
	}
	if body := r.do(http.MethodGet, "/dns", cookie, nil).Body.String(); !strings.Contains(body, "<code>example.lan</code>, <code>home.arpa.lab</code>") {
		t.Fatalf("the DNS page lacks the extra domains:\n%s", body)
	}
	if w := r.do(http.MethodPost, "/settings", cookie, form("dns_extra_domains", "arpa.home")); w.Code != http.StatusOK {
		t.Fatalf("another set of domains = %d\n%s", w.Code, w.Body)
	}
	if w := r.do(http.MethodPost, "/settings", cookie, form("dns_zone", "home.arpa", "dns_extra_domains", "lab.home.arpa")); w.Code != http.StatusConflict {
		t.Fatalf("a domain inside the zone = %d\n%s", w.Code, w.Body)
	}
	if w := r.do(http.MethodPost, "/settings", cookie, form()); w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `value="arpa.home"`) {
		t.Fatalf("a form without the field = %d\n%s", w.Code, w.Body)
	}
	if w := r.do(http.MethodPost, "/settings", cookie, form("dns_extra_domains", "")); w.Code != http.StatusOK ||
		!strings.Contains(w.Body.String(), `name="dns_extra_domains" value=""`) {
		t.Fatalf("clearing them = %d\n%s", w.Code, w.Body)
	}
}

func TestUsersAndAccount(t *testing.T) {
	r := newRig(t)
	cookie, csrf := r.signIn()
	post := func(path string, form url.Values) *httptest.ResponseRecorder {
		form.Set("csrf", csrf)
		return r.do(http.MethodPost, path, cookie, form)
	}
	if w := post("/users", url.Values{"username": {"ops"}, "password": {"another horse battery"}, "password2": {"another horse batterz"}}); w.Code != http.StatusBadRequest {
		t.Fatalf("mismatched passwords = %d", w.Code)
	}
	if w := post("/users", url.Values{"username": {"ops"}, "password": {"another horse battery"}, "password2": {"another horse battery"}}); w.Code != http.StatusOK ||
		!strings.Contains(w.Body.String(), "Added ops") {
		t.Fatalf("add user = %d", w.Code)
	}
	if w := post("/users", url.Values{"username": {"ops"}, "password": {"another horse battery"}, "password2": {"another horse battery"}}); w.Code != http.StatusConflict {
		t.Fatalf("duplicate user = %d", w.Code)
	}
	if w := post("/users/admin/disable", url.Values{}); w.Code != http.StatusBadRequest {
		t.Fatalf("disabling yourself = %d", w.Code)
	}
	for _, step := range []string{"disable", "enable", "reset-totp"} {
		if w := post("/users/ops/"+step, url.Values{}); w.Code != http.StatusOK {
			t.Fatalf("%s ops = %d", step, w.Code)
		}
	}
	if w := post("/users/ghost/reset-totp", url.Values{}); w.Code != http.StatusNotFound {
		t.Fatalf("an unknown user = %d", w.Code)
	}
	if w := post("/account/password", url.Values{"current": {"wrong horse battery"}, "new": {"a new horse battery"}, "new2": {"a new horse battery"}}); w.Code != http.StatusBadRequest {
		t.Fatalf("wrong current password = %d", w.Code)
	}
	if w := post("/account/password", url.Values{"current": {password}, "new": {"a new horse battery"}, "new2": {"a new horse batterz"}}); w.Code != http.StatusBadRequest ||
		!strings.Contains(w.Body.String(), "differ") {
		t.Fatalf("mismatched new passwords = %d", w.Code)
	}
	if w := post("/account/password", url.Values{"current": {password}, "new": {"a new horse battery"}, "new2": {"a new horse battery"}}); w.Code != http.StatusSeeOther {
		t.Fatalf("change password = %d\n%s", w.Code, w.Body)
	}
	if w := r.do(http.MethodGet, "/", cookie, nil); w.Code != http.StatusSeeOther {
		t.Fatal("a password change must end the session")
	}
}

func TestDashboardFragmentAndExport(t *testing.T) {
	r := newRig(t)
	r.seedServers()
	cookie, _ := r.signIn()
	w := r.do(http.MethodGet, "/fragments/nodes", cookie, nil)
	body := w.Body.String()
	if w.Code != http.StatusOK || !strings.HasPrefix(body, "\n<div id=\"nodes\" hx-get=\"/fragments/nodes\"") || strings.Contains(body, "<html") {
		t.Fatalf("fragment = %d\n%s", w.Code, body)
	}
	if !strings.Contains(body, "no heartbeat yet") {
		t.Fatalf("the dashboard must show nodes without a heartbeat in red:\n%s", body)
	}
	w = r.do(http.MethodGet, "/export", cookie, nil)
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "name: vps-1") ||
		w.Header().Get("Content-Disposition") != `attachment; filename="mesh-2026-10-12.yaml"` {
		t.Fatalf("export = %d %v", w.Code, w.Header())
	}
}

func TestEnrolmentQRCodeIsAPNG(t *testing.T) {
	r := newRig(t)
	w := r.do(http.MethodPost, "/login", "", url.Values{"username": {"admin"}, "password": {password}})
	cookie := cookieOf(t, w)
	w = r.do(http.MethodGet, "/login/enroll.png", cookie, nil)
	if w.Code != http.StatusOK || w.Header().Get("Content-Type") != "image/png" {
		t.Fatalf("QR = %d %v", w.Code, w.Header())
	}
	if _, err := png.Decode(w.Body); err != nil {
		t.Fatalf("not a PNG: %v", err)
	}
}

// A wrong enrolment code shows the same secret again, so the app already set
// up keeps working.
func TestWrongEnrolmentCodeKeepsTheSecret(t *testing.T) {
	r := newRig(t)
	cookie := cookieOf(t, r.do(http.MethodPost, "/login", "", url.Values{"username": {"admin"}, "password": {password}}))
	key, err := r.svc.BeginEnrollment(context.Background(), cookie)
	if err != nil {
		t.Fatal(err)
	}
	s, _ := r.svc.Current(context.Background(), cookie)
	w := r.do(http.MethodPost, "/login/enroll", cookie, url.Values{"csrf": {s.CSRF}, "code": {"000000"}})
	if w.Code != http.StatusBadRequest || !strings.Contains(w.Body.String(), group(key.Secret(), 4)) {
		t.Fatalf("a wrong enrolment code = %d\n%s", w.Code, w.Body)
	}
	if w := r.do(http.MethodGet, "/", cookie, nil); w.Code != http.StatusSeeOther {
		t.Fatalf("a failed enrolment reached the dashboard: %d", w.Code)
	}
}

func TestHTMXIsPinned(t *testing.T) {
	b, err := fs.ReadFile(staticFS, "static/htmx.min.js")
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(b)
	if got := hex.EncodeToString(sum[:]); got != HTMXSHA256 {
		t.Fatalf("static/htmx.min.js has sha256 %s, want %s; vendor a new release deliberately and update HTMXSHA256", got, HTMXSHA256)
	}
	r := newRig(t)
	if w := r.do(http.MethodGet, "/static/htmx.min.js", "", nil); w.Code != http.StatusOK || !bytes.Equal(w.Body.Bytes(), b) {
		t.Fatalf("GET /static/htmx.min.js = %d", w.Code)
	}
	for _, p := range []string{"/static/", "/static/ui.go", "/static/../ui.go"} {
		if w := r.do(http.MethodGet, p, "", nil); w.Code == http.StatusOK {
			t.Errorf("%s was served", p)
		}
	}
}

// The CSP forbids inline script and style, so no template may carry them.
func TestTemplatesHaveNoInlineScriptOrStyle(t *testing.T) {
	files, err := fs.Glob(templateFS, "templates/*.html")
	if err != nil {
		t.Fatal(err)
	}
	script := regexp.MustCompile(`(?i)<script\b[^>]*>(</script>)?`)
	inline := regexp.MustCompile(`(?i)\sstyle=|<style\b|\son[a-z]+=|javascript:`)
	for _, f := range files {
		b, err := fs.ReadFile(templateFS, f)
		if err != nil {
			t.Fatal(err)
		}
		for _, m := range script.FindAll(b, -1) {
			if !bytes.Contains(m, []byte(" src=")) || !bytes.HasSuffix(m, []byte("</script>")) {
				t.Errorf("%s has an inline script: %q", f, m)
			}
		}
		if m := inline.Find(b); m != nil {
			t.Errorf("%s contains %q", f, m)
		}
	}
}

func golden(t *testing.T, name string, got []byte) {
	t.Helper()
	path := filepath.Join("testdata", name+".golden.html")
	if *update {
		if err := os.WriteFile(path, got, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read golden: %v (run ./scripts/dev.sh regen to create it, then review it)", err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("%s differs from %s; review the change, then ./scripts/dev.sh regen\ngot:\n%s", name, path, got)
	}
}

// TestPageGoldens pins the rendered HTML of the main pages for fixed data.
func TestPageGoldens(t *testing.T) {
	pages, err := newRenderer()
	if err != nil {
		t.Fatal(err)
	}
	base := page{User: "admin", CSRF: "csrf-token", Version: "v1.0.0"}
	rows := []nodeRow{
		{Name: "vps-1", Role: "server", MeshIP: "10.77.0.1", Endpoint: "203.0.113.10:51820", Class: "ok", State: "in sync",
			Versions: "v4 / v4", Agent: "v1.0.0", Seen: "12s ago", Peers: []peerAge{{Name: "vps-2", Age: "40s ago"}}},
		{Name: "nas", Role: "spoke", MeshIP: "10.77.1.1", Class: "off", State: "not joined", Versions: "v0 / v0", Seen: "never"},
	}
	cases := map[string]page{
		"login": {Title: "Sign in", Version: "v1.0.0", Error: "wrong username or password", Data: loginData{Username: "admin"}},
		"dashboard": func() page {
			p := base
			p.Title, p.Nav = "Dashboard", "dashboard"
			p.Data = dashboardData{Rows: rows, Warnings: []string{`warning: node "vps-3": no endpoint`}}
			return p
		}(),
		"node": func() page {
			p := base
			p.Title, p.Nav, p.Flash = "vps-1", "nodes", "Saved; config v5 is on its way to the agents"
			p.Data = nodeData{
				Node:  api.NodeInfo{Name: "vps-1", Role: "server", MeshIP: "10.77.0.1", Endpoint: "203.0.113.10:51820", ListenPort: 51820, MTU: 1420, Enrolled: true, AgentVersion: "v1.0.0", WGBackend: "kernel", Hostname: "vps-1.example.com", OS: "linux", Arch: "amd64", AppliedVersion: 4, ConfigVersion: 5},
				Class: "warn", State: "applying v5", Seen: "12s ago", Config: "{\n  \"version\": 5\n}",
				Peers: []peerRow{{Name: "vps-2", Endpoint: "203.0.113.11:51820", Handshake: "40s ago", Rx: "1.0 KiB", Tx: "2.0 KiB"}}, IsServer: true,
				DNS: dnsState(proto.DNSUnavailable),
			}
			return p
		}(),
		"firewall": func() page {
			p := base
			p.Title, p.Nav, p.Flash = "Firewall", "firewall", "Policy smb saved; config v6 is on its way to the agents"
			rules := []proto.PolicyRule{{Name: "system: ping", Services: []proto.Service{{Proto: "icmp"}}, Action: "allow"},
				{Name: "policy: smb", From: []string{"10.77.2.1"}, To: []string{"10.77.1.1"}, Services: []proto.Service{{Proto: "tcp", Ports: []string{"445"}}},
					Action: "allow"},
				{Name: "policy: allow-all", Action: "allow"}}
			p.Data = fwPage{Tab: "policies", Tabs: fwTabs, Warnings: []string{`warning: zone "lab" has no addresses`}, View: policiesView{
				Policies: []policyRow{
					{Pos: 1, Up: 0, Down: 2, Name: "smb", Action: "allow", From: "group:laptops", To: "node:nas", Services: "tcp/445", First: true},
					{Pos: 2, Up: 1, Down: 3, Name: "allow-all", Action: "allow", From: "any", To: "any", Services: "any", Description: "Every mesh address", Last: true},
				},
				Rules: ruleRows(rules),
				Form:  newPolicyForm(2),
			}}
			return p
		}(),
		"version": func() page {
			p := base
			p.Title, p.Nav = "Config v3", "versions"
			p.Data = versionData{
				Version:  versionRow{Number: 3, When: "2026-10-12 08:00:00 UTC", Actor: "user:admin", Summary: "node.rm vps-2"},
				Change:   versions.Diff{RemovedNodes: []string{"vps-2"}},
				Rollback: versions.Diff{AddedNodes: []string{"vps-2"}, Settings: []versions.FieldChange{{Field: "default_mtu", From: "1400", To: "1420"}}},
			}
			return p
		}(),
	}
	for name, p := range cases {
		t.Run(name, func(t *testing.T) {
			got, err := pages.render(name, p)
			if err != nil {
				t.Fatal(err)
			}
			golden(t, name, got)
		})
	}
}

func TestEveryPageRenders(t *testing.T) {
	pages, err := newRenderer()
	if err != nil {
		t.Fatal(err)
	}
	data := map[string]any{
		"login": loginData{}, "totp": nil, "enroll": enrollData{Account: "admin", Secret: "ABCD EFGH"}, "dashboard": dashboardData{},
		"nodes": nodesData{}, "enrollment": enrollmentData{Node: "vps-1", Token: "t"},
		"node": nodeData{Node: api.NodeInfo{Rollback: &proto.Rollback{From: 5, To: 4, Reason: "no answer from the server"}}}, "remove": api.NodeInfo{Name: "x"},
		"versions": versionsData{}, "version": versionData{Current: true}, "audit": auditData{}, "settings": settingsData{},
		"users": []userRow{{Name: "ops"}, {Name: "admin", Self: true}}, "account": nil, "error": nil,
		"dns": dnsData{}, "firewall": fwPage{Tabs: fwTabs, View: policiesView{Policies: []policyRow{{Pos: 1, Name: "p", First: true}}}},
		"policy": fwPage{View: policyView{}}, "zones": fwPage{View: zonesView{Zones: []zoneRow{{Name: "z"}}}},
		"zone": fwPage{View: zoneView{Inbound: []ruleRow{{N: 1, Text: "x", System: true}}}}, "objects": fwPage{View: objectsView{}},
		"reach": fwPage{View: reachView{Result: &api.ReachResult{Index: -1}, Zones: []string{"z"}, Rows: []matrixRow{{Zone: "z",
			Cells: []matrixCell{{Class: "ok", Verdict: "allow"}}}}}},
		"phone": phoneData{Node: "phone", Exit: "vps-2", QR: qrCode{View: "-4 -4 29 29", Path: "M0 0h7v1h-7z"}, Profile: "[Interface]", Warnings: []string{"w"}},
	}
	for name := range pages.pages {
		d, ok := data[name]
		if !ok {
			t.Errorf("page %q has no render test", name)
			continue
		}
		if _, err := pages.render(name, page{Title: name, User: "admin", CSRF: "c", Data: d}); err != nil {
			t.Errorf("render %s: %v", name, err)
		}
	}
	if _, err := pages.render("missing", page{}); err == nil {
		t.Fatal("rendering an unknown page succeeded")
	}
}

func TestHelpers(t *testing.T) {
	now := testfix.Epoch
	seen := func(d time.Duration) int64 { return now.Add(-d).Unix() }
	if group("ABCDEFGHIJ", 4) != "ABCD EFGH IJ" || bytesize(1536) != "1.5 KiB" || bytesize(512) != "512 B" || bytesize(3<<20) != "3.0 MiB" {
		t.Fatal("helpers")
	}
	// A clock that ran ahead of the node's report shows 0s, never a negative age.
	if ago(0, now) != "never" || ago(seen(90*time.Second), now) != "1m30s ago" || ago(now.Add(time.Minute).Unix(), now) != "0s ago" {
		t.Fatalf("ago: %q %q", ago(seen(90*time.Second), now), ago(now.Add(time.Minute).Unix(), now))
	}
	if !sameOrigin("https://MESH.example.com", origin) || sameOrigin("", origin) || sameOrigin("https://mesh.example.com.evil", origin) {
		t.Fatal("sameOrigin")
	}
	// Times are shown in the server's time zone (TZ, the clock's location).
	sgt := time.FixedZone("SGT", 8*3600)
	for zone, want := range map[*time.Location]string{time.UTC: "2026-10-12 08:00:00 UTC", sgt: "2026-10-12 16:00:00 SGT"} {
		u := &ui{Deps: Deps{Now: func() time.Time { return now.In(zone) }}}
		if got := u.stamp(now.Unix()); got != want {
			t.Errorf("stamp in %s = %q, want %q", zone, got, want)
		}
	}
}
