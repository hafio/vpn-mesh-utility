package ui

import (
	"net/http"
	"strings"

	"github.com/hafio/vpn-mesh-utility/internal/server/api"
	"github.com/hafio/vpn-mesh-utility/internal/server/auth"
)

// The DNS page (spec S8.3): the mesh zone, and the custom records.

// recordForm is the add form's values, kept on a refusal.
type recordForm struct {
	Name, Node, IP, IP6 string
}

type dnsData struct {
	View api.DNSView
	Form recordForm
}

func (u *ui) dnsPage(w http.ResponseWriter, r *http.Request, s auth.Session) {
	u.renderDNS(w, r, s, http.StatusOK, page{}, recordForm{})
}

func (u *ui) renderDNS(w http.ResponseWriter, r *http.Request, s auth.Session, status int, p page, f recordForm) {
	v, err := u.Nodes.DNS(r.Context())
	if err != nil {
		u.fail(w, r, s, err)
		return
	}
	p.Data = dnsData{View: v, Form: f}
	u.show(w, r, s, status, "dns", "dns", "DNS", p)
}

func (u *ui) addRecord(w http.ResponseWriter, r *http.Request, s auth.Session) {
	f := recordForm{Name: strings.TrimSpace(r.PostFormValue("name")), Node: r.PostFormValue("node"), IP: strings.TrimSpace(r.PostFormValue("ip")),
		IP6: strings.TrimSpace(r.PostFormValue("ip6"))}
	ch, err := u.Nodes.AddRecord(r.Context(), actor(s), api.RecordRequest{Name: f.Name, Node: f.Node, IP: f.IP, IP6: f.IP6})
	if err != nil {
		status, msg := u.refusal(r, err)
		u.renderDNS(w, r, s, status, page{Error: msg}, f)
		return
	}
	u.renderDNS(w, r, s, http.StatusOK, page{Flash: changed("Record "+f.Name+" added", api.NodeChange{Version: ch.Version})}, recordForm{})
}

func (u *ui) removeRecord(w http.ResponseWriter, r *http.Request, s auth.Session) {
	name := r.PathValue("name")
	ch, err := u.Nodes.RemoveRecord(r.Context(), actor(s), name)
	if err != nil {
		status, msg := u.refusal(r, err)
		u.renderDNS(w, r, s, status, page{Error: msg}, recordForm{})
		return
	}
	u.renderDNS(w, r, s, http.StatusOK, page{Flash: changed("Record "+name+" removed", api.NodeChange{Version: ch.Version})}, recordForm{})
}

// refusal is how a failed change is shown on its page. A failure that is
// not the operator's is logged with its cause, which the page never shows.
func (u *ui) refusal(r *http.Request, err error) (int, string) {
	status, msg := message(err)
	if status >= http.StatusInternalServerError {
		u.Log.Error("ui change failed", "method", r.Method, "path", r.URL.Path, "err", err)
	}
	return status, msg
}
