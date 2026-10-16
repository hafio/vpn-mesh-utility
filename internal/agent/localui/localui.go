// Package localui is the agent's browser UI on 127.0.0.1:7443 (spec S7.6,
// amendment A7): this node's status and peers, the mesh as the server last
// showed it (readable while the server is down), the node's own settings --
// exit, keepalive, MTU, split DNS -- a tunnel restart, and the agent's log.
//
// Loopback TCP is open to every local user and, through DNS rebinding, to
// any web page, so the UI trusts neither: a browser signs in only with a
// one-time link from `meshctl ui`, which needs the agent socket (root or
// group meshctl); every request must name the UI's own host; every change
// must come from the UI's own page (Origin) with the session's CSRF token.
// The pages carry no script.
package localui

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"embed"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/netip"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/hafio/vpn-mesh-utility/internal/agent/localapi"
	"github.com/hafio/vpn-mesh-utility/internal/proto"
	"github.com/hafio/vpn-mesh-utility/internal/validate"
)

//go:embed templates/*.html static/app.css
var assets embed.FS

// Lifetimes and limits (amendment A7).
const (
	// TokenTTL is how long a sign-in link works, once.
	TokenTTL    = 2 * time.Minute
	sessionIdle = time.Hour
	sessionMax  = 12 * time.Hour
	maxTokens   = 8
	maxSessions = 8
	cookieName  = "meshctl_local"
	maxForm     = 16 << 10
)

// Funcs are what the UI calls in the agent.
type Funcs struct {
	Status  func(ctx context.Context) (proto.LocalStatus, error)
	Prefs   func(ctx context.Context, req proto.PrefsRequest) (proto.PrefsResponse, error)
	DNS     func(ctx context.Context, on bool) (proto.LocalStatus, error)
	Mesh    func(ctx context.Context) (proto.LocalMesh, error)
	Logs    localapi.Logs
	Restart func(ctx context.Context) error
}

// Config configures a UI.
type Config struct {
	// Addr is the loopback address the UI listens on (ParseAddr).
	Addr  netip.AddrPort
	Funcs Funcs
	Now   func() time.Time
	Log   *slog.Logger
	// Rand is the source of tokens; nil means crypto/rand.
	Rand io.Reader
}

// UI is the local UI. Sessions live in memory: an agent restart signs
// everyone out.
type UI struct {
	c     Config
	hosts []string
	pages *pages

	mu       sync.Mutex
	tokens   map[[32]byte]time.Time
	sessions map[[32]byte]*session
}

type session struct {
	csrf    string
	created time.Time
	seen    time.Time
}

// ParseAddr checks --ui-listen: a loopback IPv4 address and port, since
// the UI never listens on the network (amendment A7).
func ParseAddr(s string) (netip.AddrPort, error) {
	ap, err := validate.Endpoint("--ui-listen", s)
	if err != nil {
		return netip.AddrPort{}, err
	}
	if !ap.Addr().IsLoopback() {
		return netip.AddrPort{}, fmt.Errorf("--ui-listen: %s is not a loopback address; the local UI listens on 127.0.0.1 only (amendment A7)", s)
	}
	return ap, nil
}

// New returns the UI for c.
func New(c Config) (*UI, error) {
	if c.Now == nil {
		c.Now = time.Now
	}
	if c.Log == nil {
		c.Log = slog.New(slog.DiscardHandler)
	}
	if c.Rand == nil {
		c.Rand = rand.Reader
	}
	p, err := loadPages()
	if err != nil {
		return nil, err
	}
	port := c.Addr.Port()
	return &UI{
		c: c, pages: p,
		// The names a browser may use for the UI; any other Host header is
		// a rebinding page or a mistake.
		hosts:  []string{c.Addr.String(), fmt.Sprintf("localhost:%d", port)},
		tokens: map[[32]byte]time.Time{}, sessions: map[[32]byte]*session{},
	}, nil
}

// Listen opens the UI's socket.
func Listen(addr netip.AddrPort) (net.Listener, error) {
	ln, err := net.Listen("tcp", addr.String())
	if err != nil {
		return nil, fmt.Errorf("the local UI cannot listen on %s: %w (another program has the port? choose another with --ui-listen, or turn the UI off with --ui-listen off)", addr, err)
	}
	return ln, nil
}

// Serve serves the UI on ln until ctx ends.
func (u *UI) Serve(ctx context.Context, ln net.Listener) error {
	srv := &http.Server{Handler: u.Handler(), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 30 * time.Second,
		WriteTimeout: 30 * time.Second, IdleTimeout: 2 * time.Minute, BaseContext: func(net.Listener) context.Context { return ctx }}
	go func() {
		<-ctx.Done()
		sctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = srv.Shutdown(sctx) // ending anyway
	}()
	if err := srv.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

// Token mints a one-time sign-in link, good for TokenTTL.
func (u *UI) Token(context.Context) (proto.UIToken, error) {
	tok, err := u.random()
	if err != nil {
		return proto.UIToken{}, err
	}
	now := u.c.Now()
	u.mu.Lock()
	u.expire(now)
	for len(u.tokens) >= maxTokens {
		delete(u.tokens, oldest(u.tokens, func(at time.Time) time.Time { return at }))
	}
	u.tokens[sum(tok)] = now.Add(TokenTTL)
	u.mu.Unlock()
	return proto.UIToken{URL: "http://" + u.c.Addr.String() + "/login?t=" + tok, ExpiresAt: now.Add(TokenTTL).Unix()}, nil
}

func (u *UI) random() (string, error) {
	b := make([]byte, 32)
	if _, err := io.ReadFull(u.c.Rand, b); err != nil {
		return "", fmt.Errorf("make a local UI token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func sum(s string) [32]byte { return sha256.Sum256([]byte(s)) }

// oldest is the key whose time is earliest.
func oldest[V any](m map[[32]byte]V, at func(V) time.Time) [32]byte {
	var key [32]byte
	var first time.Time
	for k, v := range m {
		if t := at(v); first.IsZero() || t.Before(first) {
			key, first = k, t
		}
	}
	return key
}

// expire drops spent tokens and sessions. The caller holds mu.
func (u *UI) expire(now time.Time) {
	for k, at := range u.tokens {
		if !now.Before(at) {
			delete(u.tokens, k)
		}
	}
	for k, s := range u.sessions {
		if now.Sub(s.seen) > sessionIdle || now.Sub(s.created) > sessionMax {
			delete(u.sessions, k)
		}
	}
}

// Handler serves the UI. Every answer forbids framing, sniffing, caching
// and referrers; a request for any other host is refused before routing.
func (u *UI) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /login", u.login)
	mux.HandleFunc("GET /static/app.css", u.css)
	mux.HandleFunc("GET /{$}", u.signedIn(u.dashboard))
	mux.HandleFunc("GET /logs", u.signedIn(u.logs))
	mux.HandleFunc("POST /exit", u.change(u.exit))
	mux.HandleFunc("POST /settings", u.change(u.settings))
	mux.HandleFunc("POST /dns", u.change(u.dns))
	mux.HandleFunc("POST /restart", u.change(u.restart))
	mux.HandleFunc("POST /logout", u.change(u.logout))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("Content-Security-Policy", "default-src 'self'; frame-ancestors 'none'; form-action 'self'; base-uri 'none'")
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "no-referrer")
		h.Set("Cache-Control", "no-store")
		if !slices.Contains(u.hosts, r.Host) {
			http.Error(w, "This is meshctl's local UI; open it at http://"+u.c.Addr.String()+"/", http.StatusMisdirectedRequest)
			return
		}
		mux.ServeHTTP(w, r)
	})
}

// login trades a one-time token for a session cookie, then sends the
// browser on to the dashboard, so the token leaves the address bar.
func (u *UI) login(w http.ResponseWriter, r *http.Request) {
	tok, err := validate.LocalToken("t", r.URL.Query().Get("t"))
	if err != nil || !u.consume(tok) {
		u.c.Log.Warn("local UI: a sign-in link was refused (used, expired or mistyped)", "from", r.RemoteAddr)
		u.render(w, http.StatusUnauthorized, "signin.html", page{Title: "Sign in",
			Error: "This sign-in link does not work: it was used already, it expired after " + TokenTTL.String() + ", or it was mistyped."})
		return
	}
	id, err := u.random()
	var csrf string
	if err == nil {
		csrf, err = u.random()
	}
	if err != nil {
		u.c.Log.Error("local UI: no session", "err", err)
		http.Error(w, "The agent could not make a session; its log says why.", http.StatusInternalServerError)
		return
	}
	now := u.c.Now()
	u.mu.Lock()
	u.expire(now)
	for len(u.sessions) >= maxSessions {
		delete(u.sessions, oldest(u.sessions, func(s *session) time.Time { return s.seen }))
	}
	u.sessions[sum(id)] = &session{csrf: csrf, created: now, seen: now}
	u.mu.Unlock()
	// Not Secure: the UI is plain HTTP on loopback.
	http.SetCookie(w, &http.Cookie{Name: cookieName, Value: id, Path: "/", HttpOnly: true, SameSite: http.SameSiteStrictMode})
	u.c.Log.Info("local UI: signed in", "from", r.RemoteAddr)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// consume spends a sign-in token: true once, while it is fresh.
func (u *UI) consume(tok string) bool {
	k := sum(tok)
	u.mu.Lock()
	defer u.mu.Unlock()
	u.expire(u.c.Now())
	_, ok := u.tokens[k]
	delete(u.tokens, k)
	return ok
}

// session is the request's session, nil when it has none.
func (u *UI) session(r *http.Request) *session {
	c, err := r.Cookie(cookieName)
	if err != nil {
		return nil
	}
	if _, err := validate.LocalToken("cookie", c.Value); err != nil {
		return nil
	}
	now := u.c.Now()
	u.mu.Lock()
	defer u.mu.Unlock()
	u.expire(now)
	s := u.sessions[sum(c.Value)]
	if s != nil {
		s.seen = now
	}
	return s
}

type handler func(w http.ResponseWriter, r *http.Request, s *session)

// signedIn serves h to a signed-in browser, the sign-in page to anyone else.
func (u *UI) signedIn(h handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s := u.session(r)
		if s == nil {
			u.render(w, http.StatusUnauthorized, "signin.html", page{Title: "Sign in"})
			return
		}
		h(w, r, s)
	}
}

// change serves h to a change from the UI's own page: a session, the UI's
// Origin and the session's CSRF token, or nothing is called.
func (u *UI) change(h handler) http.HandlerFunc {
	return u.signedIn(func(w http.ResponseWriter, r *http.Request, s *session) {
		if o := r.Header.Get("Origin"); !slices.Contains(u.hosts, hostOf(o)) {
			u.c.Log.Warn("local UI: refused a change from another page", "origin", validOrNot(o), "path", r.URL.Path)
			http.Error(w, "This change did not come from the local UI's own page; open it at http://"+u.c.Addr.String()+"/", http.StatusForbidden)
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, maxForm)
		if err := r.ParseForm(); err != nil || subtle.ConstantTimeCompare([]byte(r.PostFormValue("csrf")), []byte(s.csrf)) != 1 {
			http.Error(w, "This form is out of date; reload the page and try again.", http.StatusForbidden)
			return
		}
		h(w, r, s)
	})
}

// hostOf is the host of an http:// origin; anything else is "".
func hostOf(origin string) string {
	host, ok := strings.CutPrefix(origin, "http://")
	if !ok {
		return ""
	}
	return host
}

// validOrNot is an Origin fit for the log: it came from a browser, so it is
// logged only when it is plain text.
func validOrNot(o string) string {
	if _, err := validate.Text("origin", o, 200); err != nil {
		return "(not shown: not plain text)"
	}
	return o
}

func (u *UI) css(w http.ResponseWriter, _ *http.Request) {
	b, err := assets.ReadFile("static/app.css")
	if err != nil {
		http.Error(w, "missing stylesheet", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/css; charset=utf-8")
	_, _ = w.Write(b) // the browser went away
}
