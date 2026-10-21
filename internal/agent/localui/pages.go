package localui

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/hafio/vpn-mesh-utility/internal/agent/localapi"
	"github.com/hafio/vpn-mesh-utility/internal/proto"
	"github.com/hafio/vpn-mesh-utility/internal/validate"
)

// page is what every template gets.
type page struct {
	Title string
	// CSRF is the session's token; empty on the sign-in page.
	CSRF  string
	Flash string
	Error string
	// Refresh reloads the dashboard every 10 s.
	Refresh bool
	Data    any
}

type pages struct{ t *template.Template }

func loadPages() (*pages, error) {
	t, err := template.New("").ParseFS(assets, "templates/*.html")
	if err != nil {
		return nil, fmt.Errorf("local UI templates: %w", err)
	}
	return &pages{t: t}, nil
}

// render writes a page, or a plain error when it does not render.
func (u *UI) render(w http.ResponseWriter, status int, name string, p page) {
	var b bytes.Buffer
	if err := u.pages.t.ExecuteTemplate(&b, name, p); err != nil {
		u.c.Log.Error("local UI: a page did not render", "page", name, "err", err)
		http.Error(w, "The page did not render; the agent's log says why.", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_, _ = b.WriteTo(w) // the browser went away
}

// Display helpers, as `meshctl status` has them (cmd/meshctl/status.go):
// the agent cannot import the CLI.

func ago(unix int64, now time.Time) string {
	if unix == 0 {
		return "never"
	}
	d := now.Sub(time.Unix(unix, 0)).Round(time.Second)
	if d < 0 {
		d = 0
	}
	return d.String() + " ago"
}

func bytesize(n int64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	div, exp := int64(unit), 0
	for m := n / unit; m >= unit; m /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(n)/float64(div), "KMGTPE"[exp])
}

type peerRow struct{ Name, Endpoint, Handshake, Rx, Tx string }

type meshRow struct{ Name, Role, Via, MeshIP, Class, State, Seen, Versions string }

type dashboardData struct {
	Status          proto.LocalStatus
	Contact, Config string
	Peers           []peerRow
	IsSpoke         bool
	// DNSOn offers to turn split DNS off; otherwise on.
	DNSOn     bool
	Mesh      []meshRow
	MeshAsOf  string
	MeshError string
}

func (u *UI) dashboard(w http.ResponseWriter, r *http.Request, s *session) {
	u.showDashboard(w, r, s, http.StatusOK, "", "")
}

// showDashboard renders this node and the mesh, with what a change did.
func (u *UI) showDashboard(w http.ResponseWriter, r *http.Request, s *session, status int, flash, problem string) {
	ctx, now := r.Context(), u.c.Now()
	st, err := u.c.Funcs.Status(ctx)
	if err != nil {
		problem = strings.TrimSpace(problem + " The agent's status: " + err.Error())
	}
	d := dashboardData{Status: st, Contact: ago(st.LastContact, now), IsSpoke: st.Node.Role == proto.RoleSpoke, DNSOn: st.DNS != proto.DNSOff,
		Config: fmt.Sprintf("v%d, applied %s", st.AppliedVersion, ago(st.AppliedAt, now))}
	if st.AppliedVersion == 0 && st.AppliedHash == "" {
		d.Config = "none applied yet"
	}
	for _, p := range st.Peers {
		ep := p.Endpoint
		if ep == "" {
			ep = "-"
		}
		d.Peers = append(d.Peers, peerRow{Name: p.Name, Endpoint: ep, Handshake: ago(p.LastHandshake, now), Rx: bytesize(p.RxBytes), Tx: bytesize(p.TxBytes)})
	}
	m, err := u.c.Funcs.Mesh(ctx)
	d.MeshError = m.Error
	if err != nil {
		d.MeshError = err.Error()
	}
	d.MeshAsOf = "never fetched yet"
	if m.FetchedAt != 0 {
		d.MeshAsOf = "fetched " + ago(m.FetchedAt, now)
	}
	for _, n := range m.View.Nodes {
		var via []string
		if n.Hub != "" {
			via = append(via, "hub "+n.Hub)
		}
		if n.Exit != "" {
			via = append(via, "exit "+n.Exit)
		}
		d.Mesh = append(d.Mesh, meshRow{Name: n.Name, Role: string(n.Role), Via: strings.Join(via, ", "), MeshIP: n.MeshIP,
			Class: n.Health, State: n.State, Seen: ago(n.LastSeen, now), Versions: fmt.Sprintf("v%d / v%d", n.AppliedVersion, n.ConfigVersion)})
	}
	u.render(w, status, "dashboard.html", page{Title: st.Node.Name, CSRF: s.csrf, Flash: flash, Error: problem, Refresh: true, Data: d})
}

// refused shows why a change did not happen: the agent's or the server's
// refusal with its status, anything else (the server unreachable) as 502.
func (u *UI) refused(w http.ResponseWriter, r *http.Request, s *session, err error) {
	status, msg := http.StatusBadGateway, "The server could not be asked: "+err.Error()
	var rf *localapi.Refusal
	if errors.As(err, &rf) {
		status, msg = rf.Status, rf.Message
	}
	u.showDashboard(w, r, s, status, "", msg)
}

// prefsFlash says what the server made of a change of this node's settings.
func prefsFlash(what string, resp proto.PrefsResponse) string {
	if resp.Version == 0 {
		return what + ": nothing changed."
	}
	return fmt.Sprintf("%s: the server made config v%d; this node applies it in a moment.", what, resp.Version)
}

func (u *UI) exit(w http.ResponseWriter, r *http.Request, s *session) {
	via := r.PostFormValue("exit")
	what := "Exit via " + via
	if via == "off" {
		via, what = "", "Exit off"
	}
	if via != "" {
		if _, err := validate.Name("exit", via); err != nil {
			u.showDashboard(w, r, s, http.StatusBadRequest, "", err.Error())
			return
		}
	}
	resp, err := u.c.Funcs.Prefs(r.Context(), proto.PrefsRequest{Exit: &via})
	if err != nil {
		u.refused(w, r, s, err)
		return
	}
	u.c.Log.Info("local UI: exit changed", "exit", r.PostFormValue("exit"), "version", resp.Version)
	u.showDashboard(w, r, s, http.StatusOK, prefsFlash(what, resp), "")
}

func (u *UI) settings(w http.ResponseWriter, r *http.Request, s *session) {
	var req proto.PrefsRequest
	for _, f := range []struct {
		name string
		dst  **int
	}{{"keepalive", &req.Keepalive}, {"mtu", &req.MTU}} {
		v := strings.TrimSpace(r.PostFormValue(f.name))
		if v == "" {
			continue
		}
		n, err := strconv.Atoi(v)
		if err != nil {
			u.showDashboard(w, r, s, http.StatusBadRequest, "", f.name+": must be a whole number")
			return
		}
		*f.dst = &n
	}
	if err := req.Validate(); err != nil {
		u.showDashboard(w, r, s, http.StatusBadRequest, "", err.Error())
		return
	}
	resp, err := u.c.Funcs.Prefs(r.Context(), req)
	if err != nil {
		u.refused(w, r, s, err)
		return
	}
	u.c.Log.Info("local UI: settings changed", "version", resp.Version)
	u.showDashboard(w, r, s, http.StatusOK, prefsFlash("Settings", resp), "")
}

func (u *UI) dns(w http.ResponseWriter, r *http.Request, s *session) {
	v, err := validate.Enum("dns", r.PostFormValue("dns"), "on", "off")
	if err != nil {
		u.showDashboard(w, r, s, http.StatusBadRequest, "", err.Error())
		return
	}
	st, err := u.c.Funcs.DNS(r.Context(), v == "on")
	if err != nil {
		u.refused(w, r, s, err)
		return
	}
	u.c.Log.Info("local UI: split DNS changed", "dns", st.DNS)
	u.showDashboard(w, r, s, http.StatusOK, "Split DNS "+v+" on this node (now "+st.DNS+").", "")
}

func (u *UI) restart(w http.ResponseWriter, r *http.Request, s *session) {
	if err := u.c.Funcs.Restart(r.Context()); err != nil {
		u.refused(w, r, s, err)
		return
	}
	u.showDashboard(w, r, s, http.StatusOK, "Tunnel restarted: every peer handshakes anew.", "")
}

func (u *UI) logout(w http.ResponseWriter, r *http.Request, _ *session) {
	if c, err := r.Cookie(cookieName); err == nil {
		u.mu.Lock()
		delete(u.sessions, sum(c.Value))
		u.mu.Unlock()
	}
	http.SetCookie(w, &http.Cookie{Name: cookieName, Value: "", Path: "/", MaxAge: -1, HttpOnly: true, SameSite: http.SameSiteStrictMode})
	u.render(w, http.StatusOK, "signin.html", page{Title: "Sign in", Flash: "Signed out."})
}

type logsData struct {
	Lines []string
}

func (u *UI) logs(w http.ResponseWriter, _ *http.Request, s *session) {
	var d logsData
	if u.c.Funcs.Logs != nil {
		d.Lines = u.c.Funcs.Logs.Lines()
	}
	u.render(w, http.StatusOK, "logs.html", page{Title: "Log", CSRF: s.csrf, Data: d})
}
