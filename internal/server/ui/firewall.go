package ui

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/hafio/vpn-mesh-utility/internal/server/api"
	"github.com/hafio/vpn-mesh-utility/internal/server/auth"
)

// The firewall page (spec S8.3): every server's rendered rules with where
// each comes from, and the custom rules.

// ruleForm is the add form's values, kept on a refusal.
type ruleForm struct {
	Node, Proto, Port, Source, Note string
}

type firewallData struct {
	View    api.FirewallView
	Servers []string
	Form    ruleForm
}

func (u *ui) firewallPage(w http.ResponseWriter, r *http.Request, s auth.Session) {
	u.renderFirewall(w, r, s, http.StatusOK, page{}, ruleForm{Proto: "tcp"})
}

func (u *ui) renderFirewall(w http.ResponseWriter, r *http.Request, s auth.Session, status int, p page, f ruleForm) {
	v, err := u.Nodes.Firewall(r.Context())
	if err != nil {
		u.fail(w, r, s, err)
		return
	}
	d := firewallData{View: v, Form: f}
	for _, sf := range v.Servers {
		d.Servers = append(d.Servers, sf.Name)
	}
	p.Data = d
	u.show(w, r, s, status, "firewall", "firewall", "Firewall", p)
}

func (u *ui) addRule(w http.ResponseWriter, r *http.Request, s auth.Session) {
	f := ruleForm{Node: r.PostFormValue("node"), Proto: r.PostFormValue("proto"), Port: strings.TrimSpace(r.PostFormValue("port")),
		Source: strings.TrimSpace(r.PostFormValue("source")), Note: strings.TrimSpace(r.PostFormValue("note"))}
	port, err := atoi("port", f.Port)
	var ch api.MeshChange
	if err == nil {
		ch, err = u.Nodes.AddRule(r.Context(), actor(s), api.RuleRequest{Node: f.Node, Proto: f.Proto, Port: port, Source: f.Source, Note: f.Note})
	}
	if err != nil {
		status, msg := u.refusal(r, err)
		u.renderFirewall(w, r, s, status, page{Error: msg}, f)
		return
	}
	what := fmt.Sprintf("Rule %s/%d from %s added", f.Proto, port, f.Source)
	u.renderFirewall(w, r, s, http.StatusOK, page{Flash: changed(what, api.NodeChange{Version: ch.Version})}, ruleForm{Proto: "tcp"})
}

func (u *ui) removeRule(w http.ResponseWriter, r *http.Request, s auth.Session) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		u.renderFirewall(w, r, s, http.StatusNotFound, page{Error: "there is no such custom rule; the list below has them all"}, ruleForm{Proto: "tcp"})
		return
	}
	ch, err := u.Nodes.RemoveRule(r.Context(), actor(s), id)
	if err != nil {
		status, msg := u.refusal(r, err)
		u.renderFirewall(w, r, s, status, page{Error: msg}, ruleForm{Proto: "tcp"})
		return
	}
	u.renderFirewall(w, r, s, http.StatusOK, page{Flash: changed(fmt.Sprintf("Rule %d removed", id), api.NodeChange{Version: ch.Version})},
		ruleForm{Proto: "tcp"})
}
