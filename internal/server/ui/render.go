package ui

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"path"
	"strings"

	"github.com/hafio/vpn-mesh-utility/internal/server/api"
	"github.com/hafio/vpn-mesh-utility/internal/server/auth"
	"github.com/hafio/vpn-mesh-utility/internal/server/store"
	"github.com/hafio/vpn-mesh-utility/internal/validate"
)

//go:embed templates/*.html
var templateFS embed.FS

//go:embed static/app.css static/htmx.min.js
var staticFS embed.FS

// HTMXSHA256 pins the vendored htmx 2.0.11 (npm htmx.org, dist/htmx.min.js,
// whose tarball matched the registry's sha512 integrity when vendored).
// TestHTMXIsPinned fails if the file changes.
const HTMXSHA256 = "d6fdc75f204e6bdefa99b69bf1e6d4ac69b8a364f77929f45c13476b4000f717"

// page is what every template receives.
type page struct {
	Title string
	// Nav is the active section ("" on the sign-in pages, which have no nav).
	Nav     string
	User    string
	CSRF    string
	Flash   string
	Error   string
	Version string
	Data    any
	// Mesh is the mesh a page of one mesh shows and Base the path of its
	// pages ("/m/<mesh>"), which every link to them starts with; Meshes are
	// every mesh, for the switcher. All are empty on the control plane's
	// own pages.
	Mesh   string
	Base   string
	Meshes []string
	// Admin shows the control plane's own pages in the navigation.
	Admin bool
}

// renderer holds one parsed template set per page: layout.html, the shared
// partials (*.partial.html), and the page's own file, which defines
// "content".
type renderer struct {
	pages map[string]*template.Template
}

func newRenderer() (*renderer, error) {
	base, err := template.New("layout.html").ParseFS(templateFS, "templates/layout.html", "templates/*.partial.html")
	if err != nil {
		return nil, fmt.Errorf("parse layout: %w", err)
	}
	files, err := fs.Glob(templateFS, "templates/*.html")
	if err != nil {
		return nil, fmt.Errorf("list templates: %w", err)
	}
	r := &renderer{pages: map[string]*template.Template{}}
	for _, f := range files {
		name := strings.TrimSuffix(path.Base(f), ".html")
		if name == "layout" || strings.HasSuffix(name, ".partial") {
			continue
		}
		t, err := base.Clone()
		if err != nil {
			return nil, err
		}
		if _, err := t.ParseFS(templateFS, f); err != nil {
			return nil, fmt.Errorf("parse %s: %w", f, err)
		}
		r.pages[name] = t
	}
	return r, nil
}

// render executes a page into a buffer first, so a template error never
// sends half a page.
func (r *renderer) render(name string, p page) ([]byte, error) {
	t, ok := r.pages[name]
	if !ok {
		return nil, fmt.Errorf("no page template %q", name)
	}
	var b bytes.Buffer
	if err := t.ExecuteTemplate(&b, "layout.html", p); err != nil {
		return nil, fmt.Errorf("render %s: %w", name, err)
	}
	return b.Bytes(), nil
}

// fragment executes one named block of a page's template set.
func (r *renderer) fragment(name, block string, data any) ([]byte, error) {
	t, ok := r.pages[name]
	if !ok {
		return nil, fmt.Errorf("no page template %q", name)
	}
	var b bytes.Buffer
	if err := t.ExecuteTemplate(&b, block, data); err != nil {
		return nil, fmt.Errorf("render %s/%s: %w", name, block, err)
	}
	return b.Bytes(), nil
}

func writeHTML(w http.ResponseWriter, status int, body []byte) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_, _ = w.Write(body) // the browser went away; nothing to do
}

// show renders a page for a signed-in user.
func (u *ui) show(w http.ResponseWriter, r *http.Request, s auth.Session, status int, name, nav, title string, p page) {
	p.Title, p.Nav, p.User, p.CSRF, p.Version = title, nav, s.Username, s.CSRF, u.Version
	p.Mesh, p.Base, p.Admin = u.mesh.Name, u.base(), s.Admin
	for _, m := range u.meshes {
		p.Meshes = append(p.Meshes, m.Name)
	}
	body, err := u.pages.render(name, p)
	if err != nil {
		u.Log.Error("render page", "page", name, "err", err)
		http.Error(w, "internal error; the server log has the details", http.StatusInternalServerError)
		return
	}
	writeHTML(w, status, body)
}

// refuse renders a short error page, without the signed-in navigation.
func (u *ui) refuse(w http.ResponseWriter, r *http.Request, status int, msg string) {
	u.show(w, r, auth.Session{}, status, "error", "", http.StatusText(status), page{Error: msg})
}

// fail renders err the way an operator should see it.
func (u *ui) fail(w http.ResponseWriter, r *http.Request, s auth.Session, err error) {
	status, msg := message(err)
	if status >= http.StatusInternalServerError {
		u.Log.Error("ui request failed", "method", r.Method, "path", r.URL.Path, "err", err)
	}
	u.show(w, r, s, status, "error", "", http.StatusText(status), page{Error: msg})
}

// message maps an error to a status and a message safe to show.
func message(err error) (int, string) {
	var (
		le *auth.LockoutError
		ve *validate.Error
	)
	switch {
	case errors.As(err, &le):
		return http.StatusTooManyRequests, le.Error()
	case errors.As(err, &ve):
		return http.StatusBadRequest, ve.Error()
	case errors.Is(err, auth.ErrBadLogin), errors.Is(err, auth.ErrNoSession):
		return http.StatusUnauthorized, err.Error()
	case errors.Is(err, auth.ErrBadCode), errors.Is(err, store.ErrTOTPReplay), errors.Is(err, auth.ErrWrongPassword),
		errors.Is(err, auth.ErrNoEnrollment), errors.Is(err, auth.ErrHasTOTP):
		return http.StatusBadRequest, err.Error()
	case errors.Is(err, auth.ErrUserExists), errors.Is(err, auth.ErrLastAdmin):
		return http.StatusConflict, err.Error()
	case errors.Is(err, auth.ErrNoUser), errors.Is(err, auth.ErrNoMesh):
		return http.StatusNotFound, err.Error()
	}
	return api.ErrorMessage(err)
}

func staticHandler() http.Handler {
	sub, err := fs.Sub(staticFS, "static")
	if err != nil {
		panic(err) // the embed pattern above guarantees the directory
	}
	files := http.StripPrefix("/static/", http.FileServer(http.FS(sub)))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/") {
			http.NotFound(w, r) // no directory listings
			return
		}
		w.Header().Set("Cache-Control", "no-cache")
		files.ServeHTTP(w, r)
	})
}
