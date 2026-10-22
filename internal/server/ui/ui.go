// Package ui is the server's web UI (spec S8.3): server-rendered Go
// templates and a vendored htmx, embedded in the binary, with no build step
// and no CDN. It signs users in through internal/server/auth and changes the
// mesh only through api.Nodes and the reconciler, so a change made here is
// versioned, audited and pushed exactly like one made with `meshctl node`.
//
// Every response carries a strict CSP (no inline script or style). Every
// request that changes something must come from this origin (Origin header)
// and carry the session's CSRF token.
package ui

import (
	"errors"
	"log/slog"
	"net/http"
	"net/netip"
	"net/url"
	"strings"
	"time"

	"github.com/hafio/vpn-mesh-utility/internal/server/api"
	"github.com/hafio/vpn-mesh-utility/internal/server/auth"
)

// Deps are the UI's dependencies.
type Deps struct {
	DB         api.DB
	Nodes      *api.Nodes
	Reconciler *api.Reconciler
	Auth       *auth.Service
	// ServerURL is the public origin; POSTs from any other origin are
	// refused.
	ServerURL      string
	Version        string
	TrustedProxies []netip.Prefix
	Now            func() time.Time
	Log            *slog.Logger
}

// CSP allows only this origin's scripts, styles and images, and no framing.
const CSP = "default-src 'none'; script-src 'self'; style-src 'self'; img-src 'self'; connect-src 'self'; " +
	"form-action 'self'; frame-ancestors 'none'; base-uri 'none'"

type ui struct {
	Deps
	pages *renderer
}

// New returns the UI handler. It serves every path the agent API does not.
func New(d Deps) (http.Handler, error) {
	if d.Now == nil {
		d.Now = time.Now
	}
	if d.Log == nil {
		d.Log = slog.New(slog.DiscardHandler)
	}
	pages, err := newRenderer()
	if err != nil {
		return nil, err
	}
	u := &ui{Deps: d, pages: pages}
	mux := http.NewServeMux()
	mux.Handle("GET /static/", staticHandler())

	// Sign-in: no full session yet.
	mux.HandleFunc("GET /login", u.loginPage)
	mux.HandleFunc("POST /login", u.sameOrigin(u.login))
	mux.HandleFunc("GET /login/totp", u.pending(u.totpPage))
	mux.HandleFunc("POST /login/totp", u.sameOrigin(u.pending(u.totp)))
	mux.HandleFunc("GET /login/enroll", u.pending(u.enrollPage))
	mux.HandleFunc("GET /login/enroll.png", u.pending(u.enrollQR))
	mux.HandleFunc("POST /login/enroll", u.sameOrigin(u.pending(u.enroll)))

	// Everything else needs a full session.
	for pattern, h := range map[string]handler{
		"POST /logout":                                u.logout,
		"GET /{$}":                                    u.dashboard,
		"GET /nodes":                                  u.nodesPage,
		"POST /nodes":                                 u.addNode,
		"GET /nodes/{name}":                           u.nodePage,
		"POST /nodes/{name}":                          u.editNode,
		"POST /nodes/{name}/disable":                  u.disableNode,
		"POST /nodes/{name}/enable":                   u.enableNode,
		"POST /nodes/{name}/token":                    u.tokenNode,
		"GET /nodes/{name}/compose":                   u.composeYAML,
		"POST /nodes/{name}/phone":                    u.phoneProfile,
		"GET /nodes/{name}/remove":                    u.removePage,
		"POST /nodes/{name}/remove":                   u.removeNode,
		"GET /dns":                                    u.dnsPage,
		"POST /dns":                                   u.addRecord,
		"POST /dns/{name}/remove":                     u.removeRecord,
		"GET /firewall":                               u.firewallPage,
		"POST /firewall/policies":                     u.savePolicy,
		"GET /firewall/policies/{name}":               u.policyPage,
		"POST /firewall/policies/{name}":              u.savePolicy,
		"POST /firewall/policies/{name}/move":         u.movePolicy,
		"POST /firewall/policies/{name}/remove":       u.removePolicy,
		"POST /firewall/timezone":                     u.saveTimeZone,
		"GET /firewall/zones":                         u.zonesPage,
		"POST /firewall/zones":                        u.saveZone,
		"GET /firewall/zones/{name}":                  u.zonePage,
		"POST /firewall/zones/{name}":                 u.saveZone,
		"POST /firewall/zones/{name}/remove":          u.removeZone,
		"GET /firewall/objects":                       u.objectsPage,
		"POST /firewall/objects/{kind}":               u.saveObject,
		"POST /firewall/objects/{kind}/{name}/remove": u.removeObject,
		"GET /firewall/reach":                         u.reachPage,
		"GET /versions":                               u.versionsPage,
		"GET /versions/{n}":                           u.versionPage,
		"POST /versions/{n}/rollback":                 u.rollback,
		"GET /audit":                                  u.auditPage,
		"GET /settings":                               u.settingsPage,
		"POST /settings":                              u.saveSettings,
		"GET /users":                                  u.usersPage,
		"POST /users":                                 u.addUser,
		"POST /users/{name}/disable":                  u.disableUser,
		"POST /users/{name}/enable":                   u.enableUser,
		"POST /users/{name}/reset-totp":               u.resetUser,
		"GET /account":                                u.accountPage,
		"POST /account/password":                      u.changePassword,
		"GET /export":                                 u.exportYAML,
		"GET /fragments/nodes":                        u.nodesFragment,
	} {
		mux.HandleFunc(pattern, u.signedIn(pattern, h))
	}
	return headers(mux), nil
}

// handler is a page handler with its session.
type handler func(w http.ResponseWriter, r *http.Request, s auth.Session)

// headers sets the security headers of every UI response.
func headers(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("Content-Security-Policy", CSP)
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		// same-origin, not no-referrer: under no-referrer browsers send
		// "Origin: null" on the UI's own form posts, which the origin
		// check then refuses. Nothing is sent to another site either way.
		h.Set("Referrer-Policy", "same-origin")
		h.Set("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	})
}

// sameOrigin refuses a POST that a page on another origin could have sent.
// Browsers send Origin on every POST; its absence is refused too.
func (u *ui) sameOrigin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !sameOrigin(r.Header.Get("Origin"), u.ServerURL) {
			u.refuse(w, r, http.StatusForbidden, "this request did not come from "+u.ServerURL+"; open the web UI at that address")
			return
		}
		next(w, r)
	}
}

func sameOrigin(origin, want string) bool {
	o, err := url.Parse(origin)
	return err == nil && origin != "" && strings.EqualFold(o.Scheme+"://"+o.Host, want)
}

// pending runs a sign-in step for a session that passed the password.
func (u *ui) pending(next handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s, err := u.Auth.Current(r.Context(), auth.CookieValue(r))
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		if s.TOTPOK {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		if r.Method != http.MethodGet && !s.CheckCSRF(r) {
			u.refuse(w, r, http.StatusForbidden, "the form expired; reload the page and try again")
			return
		}
		next(w, r, s)
	}
}

// signedIn requires a full session, and for a change the same origin and the
// session's CSRF token. Requests the page makes on its own (the dashboard's
// refresh) do not extend the session.
func (u *ui) signedIn(pattern string, next handler) http.HandlerFunc {
	change := !strings.HasPrefix(pattern, "GET ")
	background := strings.HasPrefix(pattern, "GET /fragments/")
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			s   auth.Session
			err error
		)
		if background {
			s, err = u.Auth.Peek(r.Context(), auth.CookieValue(r))
		} else {
			s, err = u.Auth.Authorize(r.Context(), auth.CookieValue(r))
		}
		switch {
		case errors.Is(err, auth.ErrNoSession):
			if r.Header.Get("HX-Request") != "" {
				w.Header().Set("HX-Redirect", "/login")
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			if change {
				u.refuse(w, r, http.StatusUnauthorized, "your session ended; sign in again")
				return
			}
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		case err != nil:
			u.fail(w, r, s, err)
			return
		}
		if change {
			if !sameOrigin(r.Header.Get("Origin"), u.ServerURL) {
				u.refuse(w, r, http.StatusForbidden, "this request did not come from "+u.ServerURL+"; open the web UI at that address")
				return
			}
			if !s.CheckCSRF(r) {
				u.refuse(w, r, http.StatusForbidden, "the form expired; reload the page and try again")
				return
			}
		}
		next(w, r, s)
	}
}

// actor is who a UI change is audited as.
func actor(s auth.Session) api.Actor { return api.Actor{Name: "user:" + s.Username, UserID: s.UserID} }

func (u *ui) clientIP(r *http.Request) string { return auth.ClientIP(r, u.TrustedProxies).String() }
