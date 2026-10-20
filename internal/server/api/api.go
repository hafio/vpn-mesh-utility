// Package api is the server's HTTP surface: the signed agent API (spec S6),
// the public health, install and download endpoints, and the admin API the
// `meshctl node` commands reach over a Unix socket. docs/agent-protocol.md
// lists every route; TestRoutesAreDocumented keeps the two in step.
//
// Every change to desired state goes through Reconciler.Apply, the single
// write path. Heartbeats are operational state and are written directly.
package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/netip"
	"time"

	"github.com/hafio/vpn-mesh-utility/internal/proto"
	"github.com/hafio/vpn-mesh-utility/internal/server/auth"
	"github.com/hafio/vpn-mesh-utility/internal/server/store"
)

// DB is the part of *store.Store the API needs. Tests wrap a real store to
// count calls, proving a rejected request never reached the database.
type DB interface {
	Tx(ctx context.Context, fn func(*store.Tx) error) error
}

// Deps are the API's dependencies.
type Deps struct {
	DB         DB
	Hub        *Hub
	Reconciler *Reconciler
	// Now is the clock for signatures, tokens and heartbeats.
	Now func() time.Time
	Log *slog.Logger
	// ServerURL is the validated public origin agents use (--url).
	ServerURL string
	// Version is the server build; install.sh installs the same version.
	Version string
	// DistDir holds the binaries and SHA256SUMS served under /dl/. Empty
	// turns /install.sh and /dl/ into 404s that say why.
	DistDir string
	// TrustedProxies are the reverse proxies whose X-Forwarded-For is
	// believed (--trusted-proxies).
	TrustedProxies []netip.Prefix
	// KeepAlive is the SSE comment interval; zero means 20 s.
	KeepAlive time.Duration
}

// Agent API timing.
const (
	defaultKeepAlive = 20 * time.Second
	// sseRetry is the reconnect hint sent at the top of every stream.
	sseRetry = 5 * time.Second
	// writeTimeout bounds every response write, including each SSE frame;
	// the http.Server has no WriteTimeout because streams are long-lived.
	writeTimeout = 30 * time.Second
	// Joins per client IP per minute. Tokens are 256-bit, so this guards
	// the database, not the token space.
	joinLimit = 10
)

type api struct {
	Deps
	replay *replayCache
	joins  *auth.RateLimiter
	nodes  *Nodes
}

func newAPI(d Deps) *api {
	if d.Now == nil {
		d.Now = time.Now
	}
	if d.Log == nil {
		d.Log = slog.New(slog.DiscardHandler)
	}
	if d.KeepAlive <= 0 {
		d.KeepAlive = defaultKeepAlive
	}
	return &api{Deps: d, replay: newReplayCache(), joins: auth.NewRateLimiter(joinLimit, time.Minute), nodes: NewNodes(d)}
}

// Route is one endpoint as docs/agent-protocol.md lists it.
type Route struct {
	Method string
	Path   string
	// Auth is "signature" (an enrolled node's signed request), "token" (a
	// one-time join token), "socket" (the admin socket's file mode) or "none".
	Auth string
}

type route struct {
	Route
	handler http.HandlerFunc
}

func (a *api) routes() []route {
	return []route{
		{Route{http.MethodPost, proto.PathJoin, "token"}, a.join},
		{Route{http.MethodGet, proto.PathConfig, "signature"}, a.config},
		{Route{http.MethodGet, proto.PathEvents, "signature"}, a.events},
		{Route{http.MethodPost, proto.PathStatus, "signature"}, a.status},
		{Route{http.MethodPost, proto.PathPrefs, "signature"}, a.prefs},
		{Route{http.MethodGet, proto.PathMesh, "signature"}, a.mesh},
		{Route{http.MethodPut, proto.PathNames, "signature"}, a.names},
		{Route{http.MethodGet, proto.PathHealth, "none"}, a.health},
		{Route{http.MethodGet, proto.PathInstall, "none"}, a.install},
		{Route{http.MethodGet, proto.PathDownload + "{file}", "none"}, a.download},
	}
}

// Routes lists the public endpoints served by New.
func Routes() []Route { return describe((&api{}).routes()) }

// AdminRoutes lists the endpoints served by NewAdmin.
func AdminRoutes() []Route { return describe((&api{}).adminRoutes()) }

func describe(rs []route) []Route {
	out := make([]Route, len(rs))
	for i, r := range rs {
		out[i] = r.Route
	}
	return out
}

// New returns the public handler: the agent API, health, install script and
// downloads. It is what Traefik routes the server's domain to.
func New(d Deps) http.Handler {
	a := newAPI(d)
	return a.handler(a.routes())
}

// NewAdmin returns the admin handler served on <data>/admin.sock.
func NewAdmin(d Deps) http.Handler {
	a := newAPI(d)
	return a.handler(a.adminRoutes())
}

func (a *api) handler(rs []route) http.Handler {
	mux := http.NewServeMux()
	for _, r := range rs {
		mux.HandleFunc(r.Method+" "+r.Path, r.handler)
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// The http.Server has no WriteTimeout (SSE streams are long-lived),
		// so every response gets one here; events() extends it per frame.
		_ = http.NewResponseController(w).SetWriteDeadline(time.Now().Add(writeTimeout)) // unsupported writers keep the server's timeouts
		h := w.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("Cache-Control", "no-store")
		h.Set("Referrer-Policy", "no-referrer")
		mux.ServeHTTP(w, r)
	})
}

// apiError is a failure with its HTTP status and machine-readable code.
// Handlers return it, even from inside a transaction, and fail renders it.
// msg is sent to the client, so it must never carry a secret.
type apiError struct {
	status int
	code   string
	msg    string
}

func (e *apiError) Error() string { return e.msg }

func errorf(status int, code, format string, args ...any) *apiError {
	return &apiError{status: status, code: code, msg: fmt.Sprintf(format, args...)}
}

// toAPIError maps the store's and the reconciler's errors to responses.
// Anything unrecognised is a 500 whose detail stays in the server log.
func toAPIError(err error) *apiError {
	var ae *apiError
	var pe *PlanError
	switch {
	case errors.As(err, &ae):
		return ae
	case errors.As(err, &pe):
		return errorf(http.StatusConflict, proto.CodeConflict, "%v", pe)
	case errors.Is(err, store.ErrTokenUnknown), errors.Is(err, store.ErrTokenUsed), errors.Is(err, store.ErrTokenExpired):
		return errorf(http.StatusUnauthorized, proto.CodeTokenInvalid, "%v; issue a new one with `meshctl node token <name>`", err)
	case errors.Is(err, store.ErrNotFound):
		return errorf(http.StatusNotFound, proto.CodeNotFound, "%v", err)
	case errors.Is(err, store.ErrConflict):
		return errorf(http.StatusConflict, proto.CodeConflict, "%v", err)
	default:
		return errorf(http.StatusInternalServerError, proto.CodeInternal, "internal error; the server log has the details")
	}
}

func (a *api) fail(w http.ResponseWriter, r *http.Request, err error) {
	ae := toAPIError(err)
	body := proto.ErrorBody{Code: ae.code, Message: ae.msg}
	if ae.code == proto.CodeClockSkew {
		body.ServerTime = a.Now().Unix()
	}
	if ae.status >= http.StatusInternalServerError {
		a.Log.Error("request failed", "method", r.Method, "path", r.URL.Path, "err", err)
	} else {
		a.Log.Info("request refused", "method", r.Method, "path", r.URL.Path, "status", ae.status, "code", ae.code, "reason", ae.msg)
	}
	if ae.status == http.StatusTooManyRequests {
		w.Header().Set("Retry-After", "60")
	}
	writeJSON(w, ae.status, body)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	b, err := json.Marshal(v)
	if err != nil {
		status, b = http.StatusInternalServerError, []byte(`{"code":"internal","message":"encode response"}`)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(append(b, '\n')) // the client went away; nothing to do
}

// readBody reads the request body, capped at proto.MaxBodyBytes and bounded
// in time. The deadline is cleared afterwards: the http.Server sets no
// ReadTimeout because one would also cancel long-lived SSE requests.
func readBody(w http.ResponseWriter, r *http.Request) ([]byte, error) {
	rc := http.NewResponseController(w)
	_ = rc.SetReadDeadline(time.Now().Add(writeTimeout)) // unsupported writers keep the server's timeouts
	defer func() { _ = rc.SetReadDeadline(time.Time{}) }()
	b, err := io.ReadAll(http.MaxBytesReader(w, r.Body, proto.MaxBodyBytes))
	var tooBig *http.MaxBytesError
	switch {
	case errors.As(err, &tooBig):
		return nil, errorf(http.StatusRequestEntityTooLarge, proto.CodeTooLarge, "request body exceeds %d bytes", proto.MaxBodyBytes)
	case err != nil:
		return nil, errorf(http.StatusBadRequest, proto.CodeBadRequest, "read request body: %v", err)
	}
	return b, nil
}

// decodeJSON parses exactly one JSON value with no unknown fields: a field
// this server does not know is a client bug, not something to ignore.
func decodeJSON(body []byte, v any) error {
	dec := json.NewDecoder(bytes.NewReader(body))
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		return errorf(http.StatusBadRequest, proto.CodeBadRequest, "request body is not valid for this endpoint: %v", err)
	}
	if dec.More() {
		return errorf(http.StatusBadRequest, proto.CodeBadRequest, "request body has data after the JSON value")
	}
	return nil
}

// invalid wraps a validation failure (already safe to show) as a 400.
func invalid(err error) error {
	return errorf(http.StatusBadRequest, proto.CodeBadRequest, "%v", err)
}
