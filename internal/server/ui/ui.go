// Package ui is the server's web UI (spec S8.3): server-rendered Go
// templates and a vendored htmx, embedded in the binary, with no build step
// and no CDN. It signs users in through internal/server/auth and changes the
// mesh only through api.Nodes and the reconciler, so a change made here is
// versioned, audited and pushed exactly like one made with `meshctl node`.
// Each mesh's pages live under /m/<mesh>/ (amendment A49); the users, the
// account and the meshes themselves are the control plane's.
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
	"slices"
	"strings"
	"time"

	"github.com/hafio/vpn-mesh-utility/internal/server/api"
	"github.com/hafio/vpn-mesh-utility/internal/server/auth"
	"github.com/hafio/vpn-mesh-utility/internal/server/store"
	"github.com/hafio/vpn-mesh-utility/internal/validate"
)

// Deps are the UI's dependencies.
type Deps struct {
	DB    api.DB
	Nodes *api.Nodes
	Auth  *auth.Service
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
	// mesh is the mesh a page of one mesh works on (inMesh), and meshes
	// every mesh, for the switcher; both are zero on the control plane's
	// own pages.
	mesh   store.Mesh
	meshes []store.Mesh
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
	mux.HandleFunc("GET /setup", u.setupPage)
	mux.HandleFunc("POST /setup", u.sameOrigin(u.setup))

	// Everything else needs a full session: the control plane's pages...
	for pattern, h := range map[string]handler{
		"POST /logout":                  u.logout,
		"GET /{$}":                      u.home,
		"GET /m/{mesh}":                 u.meshHome,
		"GET /meshes":                   u.adminOnly(u.meshesPage),
		"POST /meshes":                  u.adminOnly(u.addMesh),
		"POST /meshes/{name}":           u.adminOnly(u.renameMesh),
		"POST /meshes/{name}/remove":    u.adminOnly(u.removeMesh),
		"GET /users":                    u.adminOnly(u.usersPage),
		"POST /users":                   u.adminOnly(u.addUser),
		"POST /users/{name}/disable":    u.adminOnly(u.disableUser),
		"POST /users/{name}/enable":     u.adminOnly(u.enableUser),
		"POST /users/{name}/reset-totp": u.adminOnly(u.resetUser),
		"POST /users/{name}/admin":      u.adminOnly(u.setAdmin),
		"POST /users/{name}/grant":      u.adminOnly(u.grantMesh),
		"POST /users/{name}/revoke":     u.adminOnly(u.revokeMesh),
		"GET /account":                  u.accountPage,
		"POST /account/password":        u.changePassword,
	} {
		mux.HandleFunc(pattern, u.signedIn(pattern, h))
	}
	// ...and each mesh's, under /m/<mesh>.
	for pattern, h := range map[string]meshHandler{
		"GET /{$}":                                    (*ui).dashboard,
		"GET /nodes":                                  (*ui).nodesPage,
		"POST /nodes":                                 (*ui).addNode,
		"GET /nodes/{name}":                           (*ui).nodePage,
		"POST /nodes/{name}":                          (*ui).editNode,
		"POST /nodes/{name}/disable":                  (*ui).disableNode,
		"POST /nodes/{name}/enable":                   (*ui).enableNode,
		"POST /nodes/{name}/token":                    (*ui).tokenNode,
		"POST /nodes/{name}/phone":                    (*ui).phoneProfile,
		"GET /nodes/{name}/remove":                    (*ui).removePage,
		"POST /nodes/{name}/remove":                   (*ui).removeNode,
		"GET /dns":                                    (*ui).dnsPage,
		"POST /dns":                                   (*ui).addRecord,
		"POST /dns/{name}/remove":                     (*ui).removeRecord,
		"GET /firewall":                               (*ui).firewallPage,
		"POST /firewall/policies":                     (*ui).savePolicy,
		"GET /firewall/policies/{name}":               (*ui).policyPage,
		"POST /firewall/policies/{name}":              (*ui).savePolicy,
		"POST /firewall/policies/{name}/move":         (*ui).movePolicy,
		"POST /firewall/policies/{name}/remove":       (*ui).removePolicy,
		"POST /firewall/timezone":                     (*ui).saveTimeZone,
		"GET /firewall/zones":                         (*ui).zonesPage,
		"POST /firewall/zones":                        (*ui).saveZone,
		"GET /firewall/zones/{name}":                  (*ui).zonePage,
		"POST /firewall/zones/{name}":                 (*ui).saveZone,
		"POST /firewall/zones/{name}/remove":          (*ui).removeZone,
		"GET /firewall/objects":                       (*ui).objectsPage,
		"POST /firewall/objects/{kind}":               (*ui).saveObject,
		"POST /firewall/objects/{kind}/{name}/remove": (*ui).removeObject,
		"GET /firewall/reach":                         (*ui).reachPage,
		"GET /versions":                               (*ui).versionsPage,
		"GET /versions/{n}":                           (*ui).versionPage,
		"POST /versions/{n}/rollback":                 (*ui).rollback,
		"GET /audit":                                  (*ui).auditPage,
		"GET /settings":                               (*ui).settingsPage,
		"POST /settings":                              (*ui).saveSettings,
		"POST /settings/ipv6":                         (*ui).saveIPv6,
		"GET /export":                                 (*ui).exportYAML,
		"GET /fragments/nodes":                        (*ui).nodesFragment,
	} {
		method, path, _ := strings.Cut(pattern, " ")
		full := method + " /m/{mesh}" + path
		mux.HandleFunc(full, u.signedIn(full, u.inMesh(h)))
	}
	return headers(mux), nil
}

// handler is a page handler with its session.
type handler func(w http.ResponseWriter, r *http.Request, s auth.Session)

// meshHandler is a page of one mesh. It runs on a copy of the UI whose Nodes,
// links and switcher are that mesh's (amendment A49).
type meshHandler func(u *ui, w http.ResponseWriter, r *http.Request, s auth.Session)

// inMesh runs h for the mesh the path names. One the control plane does not
// have, or the user may not manage, is the same 404, so a user learns no
// mesh they were not granted.
func (u *ui) inMesh(h meshHandler) handler {
	return func(w http.ResponseWriter, r *http.Request, s auth.Session) {
		meshes, err := u.meshList(r, s)
		if err != nil {
			u.fail(w, r, s, err)
			return
		}
		name := r.PathValue("mesh")
		i := slices.IndexFunc(meshes, func(m store.Mesh) bool { return m.Name == name })
		if i < 0 {
			u.show(w, r, s, http.StatusNotFound, "error", "", "Not Found", page{Error: "you manage no such mesh; the dashboard opens one you do"})
			return
		}
		c := *u
		c.Nodes, c.mesh, c.meshes = u.Nodes.In(meshes[i].ID), meshes[i], meshes
		h(&c, w, r, s)
	}
}

// adminOnly runs h for an admin; anyone else is told who may (amendment A49).
func (u *ui) adminOnly(h handler) handler {
	return func(w http.ResponseWriter, r *http.Request, s auth.Session) {
		if !s.Admin {
			u.show(w, r, s, http.StatusForbidden, "error", "", "Forbidden", page{Error: "only an admin manages the meshes and the users; ask one"})
			return
		}
		h(w, r, s)
	}
}

// meshList is every mesh the session's user may manage, by id.
func (u *ui) meshList(r *http.Request, s auth.Session) ([]store.Mesh, error) {
	var all []store.Mesh
	err := u.DB.Tx(r.Context(), func(tx *store.Tx) error {
		var err error
		all, err = tx.Meshes()
		return err
	})
	return slices.DeleteFunc(all, func(m store.Mesh) bool { return !s.CanManage(m.ID) }), err
}

// base is the path of the current mesh's pages ("/m/<mesh>"), "" on the
// control plane's own.
func (u *ui) base() string {
	if u.mesh.Name == "" {
		return ""
	}
	return "/m/" + u.mesh.Name
}

// home opens the first mesh the user manages; an admin with none goes to
// the Meshes page, anyone else is told to ask one.
func (u *ui) home(w http.ResponseWriter, r *http.Request, s auth.Session) {
	meshes, err := u.meshList(r, s)
	if err != nil {
		u.fail(w, r, s, err)
		return
	}
	switch {
	case len(meshes) == 0 && s.Admin:
		http.Redirect(w, r, "/meshes", http.StatusSeeOther)
		return
	case len(meshes) == 0:
		u.show(w, r, s, http.StatusForbidden, "error", "", "No mesh yet", page{Error: "no mesh is granted to you yet; ask an admin to grant you one"})
		return
	}
	http.Redirect(w, r, "/m/"+meshes[0].Name+"/", http.StatusSeeOther)
}

// meshHome sends /m/<mesh> to its dashboard, /m/<mesh>/.
func (u *ui) meshHome(w http.ResponseWriter, r *http.Request, s auth.Session) {
	name, err := validate.Name("mesh", r.PathValue("mesh"))
	if err != nil {
		u.fail(w, r, s, err)
		return
	}
	http.Redirect(w, r, "/m/"+name+"/", http.StatusSeeOther)
}

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
	background := strings.HasPrefix(pattern, "GET ") && strings.Contains(pattern, "/fragments/")
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
