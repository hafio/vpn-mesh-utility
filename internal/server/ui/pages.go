package ui

import (
	"bytes"
	"cmp"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/hafio/vpn-mesh-utility/internal/proto"
	"github.com/hafio/vpn-mesh-utility/internal/server/api"
	"github.com/hafio/vpn-mesh-utility/internal/server/auth"
	"github.com/hafio/vpn-mesh-utility/internal/server/export"
	"github.com/hafio/vpn-mesh-utility/internal/server/planner"
	"github.com/hafio/vpn-mesh-utility/internal/server/store"
	"github.com/hafio/vpn-mesh-utility/internal/server/versions"
	"github.com/hafio/vpn-mesh-utility/internal/validate"
)

// Display helpers. Times are computed here, not in templates, so a page
// renders the same way for a given clock.

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

// stamp is a time as the pages show it: in the server's time zone (TZ; the
// clock's location), with the zone named.
func (u *ui) stamp(unix int64) string {
	return time.Unix(unix, 0).In(u.Now().Location()).Format("2006-01-02 15:04:05 MST")
}

type peerAge struct {
	Name string
	Age  string
}

type nodeRow struct {
	Name, Role, MeshIP, Endpoint string
	Class, State                 string
	Versions, Agent, Seen, Exit  string
	Peers                        []peerAge
}

type dashboardData struct {
	// Base is the mesh's pages' path, for the table's links.
	Base     string
	Rows     []nodeRow
	Warnings []string
}

func (u *ui) dashboardData(r *http.Request) (dashboardData, error) {
	list, err := u.Nodes.List(r.Context())
	if err != nil {
		return dashboardData{}, err
	}
	peers := map[int64][]store.PeerStat{}
	err = u.DB.Tx(r.Context(), func(tx *store.Tx) error {
		for _, n := range list.Nodes {
			ps, err := tx.PeerStats(n.ID)
			if err != nil {
				return err
			}
			peers[n.ID] = ps
		}
		return nil
	})
	if err != nil {
		return dashboardData{}, err
	}
	names := map[int64]string{}
	for _, n := range list.Nodes {
		names[n.ID] = n.Name
	}
	now := u.Now()
	d := dashboardData{Base: u.base(), Warnings: list.Warnings}
	for _, n := range list.Nodes {
		row := nodeRow{Name: n.Name, Role: n.Role, MeshIP: cmp.Or(n.MeshIP, n.MeshIP6), Endpoint: cmp.Or(n.Endpoint, n.Endpoint6), Exit: n.Exit,
			Versions: fmt.Sprintf("v%d / v%d", n.AppliedVersion, n.ConfigVersion), Agent: n.AgentVersion, Seen: ago(n.Seen(), now)}
		row.Class, row.State = api.Health(n, now)
		for _, p := range peers[n.ID] {
			row.Peers = append(row.Peers, peerAge{Name: names[p.PeerID], Age: ago(p.LastHandshake, now)})
		}
		sort.Slice(row.Peers, func(i, j int) bool { return row.Peers[i].Name < row.Peers[j].Name })
		d.Rows = append(d.Rows, row)
	}
	return d, nil
}

func (u *ui) dashboard(w http.ResponseWriter, r *http.Request, s auth.Session) {
	d, err := u.dashboardData(r)
	if err != nil {
		u.fail(w, r, s, err)
		return
	}
	u.show(w, r, s, http.StatusOK, "dashboard", "dashboard", "Dashboard", page{Data: d})
}

// nodesFragment is the dashboard table alone, which the page refreshes every
// 10 s.
func (u *ui) nodesFragment(w http.ResponseWriter, r *http.Request, s auth.Session) {
	d, err := u.dashboardData(r)
	if err != nil {
		u.fail(w, r, s, err)
		return
	}
	body, err := u.pages.fragment("dashboard", "nodes-table", d)
	if err != nil {
		u.fail(w, r, s, err)
		return
	}
	writeHTML(w, http.StatusOK, body)
}

// nodeForm is the add form's values, kept on a refusal so nothing is retyped.
type nodeForm struct {
	Name, Role, IP, Endpoint string
	ListenPort, Hub          string
	Keepalive, MTU, Mode     string
	Groups                   string
	CanExit, AlwaysOn        bool
	Containers               bool
	// The IPv6 side (M8).
	Family, IP6, Endpoint6 string
	Containers6            bool
}

type nodesData struct {
	Rows    []nodeRow
	Servers []string
	Form    nodeForm
	// IPv6 says the mesh has an IPv6 prefix (M8).
	IPv6 bool
}

func (u *ui) nodesView(r *http.Request, form nodeForm) (nodesData, error) {
	d, err := u.dashboardData(r)
	if err != nil {
		return nodesData{}, err
	}
	out := nodesData{Rows: d.Rows, Form: form}
	if out.IPv6, err = u.meshHasIPv6(r); err != nil {
		return nodesData{}, err
	}
	for _, row := range d.Rows {
		if row.Role == string(proto.RoleServer) {
			out.Servers = append(out.Servers, row.Name)
		}
	}
	return out, nil
}

func (u *ui) nodesPage(w http.ResponseWriter, r *http.Request, s auth.Session) {
	d, err := u.nodesView(r, nodeForm{Role: string(proto.RoleServer)})
	if err != nil {
		u.fail(w, r, s, err)
		return
	}
	u.show(w, r, s, http.StatusOK, "nodes", "nodes", "Nodes", page{Data: d})
}

// meshHasIPv6 reports whether the mesh has an IPv6 prefix (M8): the forms
// offer IPv6 addresses only then.
func (u *ui) meshHasIPv6(r *http.Request) (bool, error) {
	v, err := u.Nodes.IPv6(r.Context())
	return v.Prefix != "", err
}

// atoi parses an optional number field; "" is 0 (the default).
func atoi(field, v string) (int, error) {
	if v = strings.TrimSpace(v); v == "" {
		return 0, nil
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0, &validate.Error{Field: field, Value: v, Reason: "must be a number"}
	}
	return n, nil
}

func (u *ui) addNode(w http.ResponseWriter, r *http.Request, s auth.Session) {
	f := nodeForm{
		Name: strings.TrimSpace(r.PostFormValue("name")), Role: r.PostFormValue("role"), IP: strings.TrimSpace(r.PostFormValue("ip")),
		Endpoint: strings.TrimSpace(r.PostFormValue("endpoint")), ListenPort: r.PostFormValue("listen_port"), Hub: r.PostFormValue("hub"),
		Keepalive: r.PostFormValue("keepalive"), MTU: r.PostFormValue("mtu"), Mode: r.PostFormValue("mode"),
		CanExit: r.PostFormValue("can_exit") != "", AlwaysOn: r.PostFormValue("always_on") != "",
		Containers: r.PostFormValue("containers") != "", Groups: strings.TrimSpace(r.PostFormValue("groups")),
		Family: r.PostFormValue("ip_family"), IP6: strings.TrimSpace(r.PostFormValue("ip6")),
		Endpoint6: strings.TrimSpace(r.PostFormValue("endpoint6")), Containers6: r.PostFormValue("containers6") != "",
	}
	req := api.NodeAddRequest{Name: f.Name, Role: f.Role, IP: f.IP, Endpoint: f.Endpoint, Hub: f.Hub, CanExit: f.CanExit, AlwaysOn: f.AlwaysOn,
		Mode: f.Mode, Containers: f.Containers, Groups: listOf(f.Groups), Family: f.Family, IP6: f.IP6, Endpoint6: f.Endpoint6,
		Containers6: f.Containers6}
	var err error
	if req.ListenPort, err = atoi("listen_port", f.ListenPort); err == nil {
		if req.Keepalive, err = atoi("keepalive", f.Keepalive); err == nil {
			req.MTU, err = atoi("mtu", f.MTU)
		}
	}
	var ch api.NodeChange
	if err == nil {
		ch, err = u.Nodes.Add(r.Context(), actor(s), req)
	}
	if err != nil {
		status, msg := message(err)
		d, verr := u.nodesView(r, f)
		if verr != nil {
			u.fail(w, r, s, verr)
			return
		}
		u.show(w, r, s, status, "nodes", "nodes", "Nodes", page{Error: msg, Data: d})
		return
	}
	if ch.Enrollment == nil {
		// A phone: its page makes its profile.
		u.renderNodeNamed(w, r, s, ch.Node.Name, http.StatusOK, page{Flash: "Phone " + ch.Node.Name + " added; make its profile below"})
		return
	}
	u.showEnrollment(w, r, s, ch, "Node "+ch.Node.Name+" added")
}

type enrollmentData struct {
	Node     string
	Token    string
	Expires  string
	Install  string
	Join     string
	Warnings []string
}

// showEnrollment shows a join token once; the server keeps only its hash.
func (u *ui) showEnrollment(w http.ResponseWriter, r *http.Request, s auth.Session, ch api.NodeChange, flash string) {
	d := enrollmentData{Node: ch.Node.Name, Warnings: ch.Warnings}
	if e := ch.Enrollment; e != nil {
		d.Token, d.Expires, d.Install, d.Join = e.Token, u.stamp(e.ExpiresAt), e.Install, e.Join
	}
	u.show(w, r, s, http.StatusOK, "enrollment", "nodes", "Join "+ch.Node.Name, page{Flash: flash, Data: d})
}

type nodeData struct {
	Node     api.NodeInfo
	Class    string
	State    string
	Seen     string
	Config   string
	Peers    []peerRow
	Warnings []string
	Servers  []string
	Exits    []string
	IsServer bool
	IsSpoke  bool
	IsPhone  bool
	// InContainer is a container-mode node (amendment A37).
	InContainer bool
	// IPv6 says the mesh has an IPv6 prefix (M8).
	IPv6 bool
	// Containers is the node's container network, the values its host's
	// runtime creates it with (amendment A48); nil without one.
	Containers *api.ContainerNetwork
	// DNS explains the node's split DNS as it last reported it.
	DNS string
	// Profile says, for a phone, whether the profile it holds is current.
	Profile string
}

type peerRow struct {
	Name, Endpoint, Handshake, Rx, Tx string
}

func (u *ui) nodeView(r *http.Request, name string) (nodeData, error) {
	d, err := u.Nodes.Show(r.Context(), name)
	if err != nil {
		return nodeData{}, err
	}
	list, err := u.Nodes.List(r.Context())
	if err != nil {
		return nodeData{}, err
	}
	now := u.Now()
	out := nodeData{Node: d.Node, Seen: ago(d.Node.Seen(), now), Warnings: d.Warnings, IsServer: d.Node.Role == string(proto.RoleServer),
		IsSpoke: d.Node.Role == string(proto.RoleSpoke), IsPhone: d.Node.Role == string(proto.RolePhone),
		InContainer: d.Node.Mode == proto.ModeContainer, Containers: d.Containers}
	out.Class, out.State = api.Health(d.Node, now)
	out.DNS = dnsState(d.Node.DNS)
	if out.IPv6, err = u.meshHasIPv6(r); err != nil {
		return nodeData{}, err
	}
	if out.IsPhone {
		out.Profile = phoneState(d.Node)
	}
	if d.Config != nil {
		b, err := json.MarshalIndent(d.Config, "", "  ")
		if err != nil {
			return nodeData{}, err
		}
		out.Config = string(b)
	}
	for _, p := range d.Peers {
		out.Peers = append(out.Peers, peerRow{Name: p.Name, Endpoint: p.Endpoint, Handshake: ago(p.LastHandshake, now),
			Rx: bytesize(p.RxBytes), Tx: bytesize(p.TxBytes)})
	}
	for _, n := range list.Nodes {
		if n.Role == string(proto.RoleServer) && n.Name != d.Node.Name {
			out.Servers = append(out.Servers, n.Name)
			if n.CanExit {
				out.Exits = append(out.Exits, n.Name)
			}
		}
	}
	return out, nil
}

// dnsState says what a node's reported split DNS means for it.
func dnsState(s string) string {
	switch s {
	case proto.DNSOn:
		return "on: the host resolves .mesh names"
	case proto.DNSOff:
		return "off: turned off on the node (`meshctl dns on` turns it back on)"
	case proto.DNSUnavailable:
		return "unavailable: the node answers .mesh on its mesh IP, but its host has no systemd-resolved to send its own lookups there"
	}
	return ""
}

func (u *ui) nodePage(w http.ResponseWriter, r *http.Request, s auth.Session) {
	u.renderNode(w, r, s, http.StatusOK, page{})
}

func (u *ui) renderNode(w http.ResponseWriter, r *http.Request, s auth.Session, status int, p page) {
	u.renderNodeNamed(w, r, s, r.PathValue("name"), status, p)
}

func (u *ui) renderNodeNamed(w http.ResponseWriter, r *http.Request, s auth.Session, name string, status int, p page) {
	d, err := u.nodeView(r, name)
	if err != nil {
		u.fail(w, r, s, err)
		return
	}
	p.Data = d
	u.show(w, r, s, status, "node", "nodes", d.Node.Name, p)
}

// editNode applies the edit form: the fields a node of its role has.
func (u *ui) editNode(w http.ResponseWriter, r *http.Request, s auth.Session) {
	name := r.PathValue("name")
	cur, err := u.Nodes.Show(r.Context(), name)
	if err != nil {
		u.fail(w, r, s, err)
		return
	}
	var req api.NodeSetRequest
	mtu, err := atoi("mtu", r.PostFormValue("mtu"))
	req.MTU = &mtu
	// A new mesh IP only when the form changes it (amendment A41).
	if ip := strings.TrimSpace(r.PostFormValue("mesh_ip")); ip != "" && ip != cur.Node.MeshIP {
		req.MeshIP = &ip
	}
	// The IPv6 fields are on the form when the mesh has an IPv6 prefix (M8).
	if r.PostFormValue("ipv6_form") != "" {
		family := r.PostFormValue("ip_family")
		if family != "" && family != cur.Node.Family {
			req.Family = &family
		}
		if ip6 := strings.TrimSpace(r.PostFormValue("mesh_ip6")); ip6 != "" && ip6 != cur.Node.MeshIP6 && family != proto.FamilyIPv4 {
			req.MeshIP6 = &ip6
		}
		// Going IPv4 only takes the containers' IPv6 too, ticked or not.
		if cur.Node.Role != string(proto.RolePhone) {
			containers6 := r.PostFormValue("containers6") != "" && family != proto.FamilyIPv4
			req.Containers6 = &containers6
		}
	}
	if cur.Node.Role == string(proto.RoleServer) {
		endpoint := strings.TrimSpace(r.PostFormValue("endpoint"))
		endpoint6 := strings.TrimSpace(r.PostFormValue("endpoint6"))
		canExit := r.PostFormValue("can_exit") != ""
		req.Endpoint, req.Endpoint6, req.CanExit = &endpoint, &endpoint6, &canExit
		if err == nil {
			var port int
			port, err = atoi("listen_port", r.PostFormValue("listen_port"))
			req.ListenPort = &port
		}
	} else {
		hub := r.PostFormValue("hub")
		req.Hub = &hub
		// A phone's exit is its profile's, not a setting.
		if cur.Node.Role != string(proto.RolePhone) {
			exit := r.PostFormValue("exit")
			req.Exit = &exit
		}
		if err == nil {
			var ka int
			ka, err = atoi("keepalive", r.PostFormValue("keepalive"))
			req.Keepalive = &ka
		}
	}
	// A container-mode node's containers are always on the mesh.
	if cur.Node.Role != string(proto.RolePhone) && cur.Node.Mode != proto.ModeContainer {
		containers := r.PostFormValue("containers") != ""
		req.Containers = &containers
	}
	if _, ok := r.PostForm["groups"]; ok {
		groups := listOf(r.PostFormValue("groups"))
		req.Groups = &groups
	}
	var ch api.NodeChange
	if err == nil {
		ch, err = u.Nodes.Set(r.Context(), actor(s), name, req)
	}
	if err != nil {
		status, msg := message(err)
		u.renderNode(w, r, s, status, page{Error: msg})
		return
	}
	flash := changed("Saved", ch)
	if len(ch.Notes) > 0 {
		flash += ". " + strings.Join(ch.Notes, ". ")
	}
	u.renderNode(w, r, s, http.StatusOK, page{Flash: flash})
}

// changed describes what a write did to the computed configs.
func changed(what string, ch api.NodeChange) string {
	if ch.Version == 0 {
		return what + "; no node's config changed"
	}
	return fmt.Sprintf("%s; config v%d is on its way to the agents", what, ch.Version)
}

func (u *ui) disableNode(w http.ResponseWriter, r *http.Request, s auth.Session) {
	u.setDisabled(w, r, s, true)
}

func (u *ui) enableNode(w http.ResponseWriter, r *http.Request, s auth.Session) {
	u.setDisabled(w, r, s, false)
}

func (u *ui) setDisabled(w http.ResponseWriter, r *http.Request, s auth.Session, disabled bool) {
	ch, err := u.Nodes.SetDisabled(r.Context(), actor(s), r.PathValue("name"), disabled)
	if err != nil {
		status, msg := message(err)
		u.renderNode(w, r, s, status, page{Error: msg})
		return
	}
	what := "Enabled"
	if disabled {
		what = "Disabled"
	}
	u.renderNode(w, r, s, http.StatusOK, page{Flash: changed(what, ch)})
}

// tokenData is the confirmation of a token for a node that is probably
// running (amendment A50).
type tokenData struct {
	Node api.NodeInfo
	Ago  time.Duration
}

func (u *ui) tokenNode(w http.ResponseWriter, r *http.Request, s auth.Session) {
	if r.PostFormValue("confirm") != "yes" {
		d, err := u.Nodes.Show(r.Context(), r.PathValue("name"))
		if err != nil {
			u.fail(w, r, s, err)
			return
		}
		// A token cuts a running node off: ask first.
		if ago := u.Now().Sub(time.Unix(d.Node.LastSeen, 0)); d.Node.Enrolled && d.Node.LastSeen != 0 && ago < api.RecentlySeen {
			u.show(w, r, s, http.StatusOK, "token", "nodes", "New join token for "+d.Node.Name,
				page{Data: tokenData{Node: d.Node, Ago: ago.Round(time.Second)}})
			return
		}
	}
	ch, err := u.Nodes.Token(r.Context(), actor(s), r.PathValue("name"))
	if err != nil {
		status, msg := message(err)
		u.renderNode(w, r, s, status, page{Error: msg})
		return
	}
	flash := "New join token for " + ch.Node.Name + "; unused older tokens no longer work"
	for _, n := range ch.Notes {
		flash += ". " + n
	}
	u.showEnrollment(w, r, s, ch, flash)
}

func (u *ui) removePage(w http.ResponseWriter, r *http.Request, s auth.Session) {
	d, err := u.Nodes.Show(r.Context(), r.PathValue("name"))
	if err != nil {
		u.fail(w, r, s, err)
		return
	}
	u.show(w, r, s, http.StatusOK, "remove", "nodes", "Remove "+d.Node.Name, page{Data: d.Node})
}

func (u *ui) removeNode(w http.ResponseWriter, r *http.Request, s auth.Session) {
	if _, err := u.Nodes.Remove(r.Context(), actor(s), r.PathValue("name")); err != nil {
		u.fail(w, r, s, err)
		return
	}
	http.Redirect(w, r, u.base()+"/nodes", http.StatusSeeOther)
}

type versionRow struct {
	Number         int64
	When, Actor    string
	Summary        string
	RolledBackFrom int64
}

type versionsData struct {
	Rows  []versionRow
	Older int64
}

const versionsPerPage = 50

// before reads the ?before= pager; anything but a positive number is "newest".
func before(r *http.Request) int64 {
	n, err := strconv.ParseInt(r.URL.Query().Get("before"), 10, 64)
	if err != nil || n < 0 {
		return 0
	}
	return n
}

func (u *ui) versionsPage(w http.ResponseWriter, r *http.Request, s auth.Session) {
	var vs []store.Version
	err := u.Nodes.Tx(r.Context(), func(tx *store.MeshTx) error {
		var err error
		vs, err = tx.Versions(versionsPerPage, before(r))
		return err
	})
	if err != nil {
		u.fail(w, r, s, err)
		return
	}
	d := versionsData{}
	for _, v := range vs {
		d.Rows = append(d.Rows, versionRow{Number: v.Version, When: u.stamp(v.CreatedAt), Actor: v.Actor, Summary: v.Summary,
			RolledBackFrom: v.RolledBackFrom})
	}
	if len(vs) == versionsPerPage {
		d.Older = vs[len(vs)-1].Version
	}
	u.show(w, r, s, http.StatusOK, "versions", "versions", "Versions", page{Data: d})
}

type versionData struct {
	Version versionRow
	// Change is what this version changed; Rollback what rolling back to
	// it would change now.
	Change, Rollback versions.Diff
	Current          bool
}

func (u *ui) versionView(r *http.Request, n int64) (versionData, error) {
	var d versionData
	err := u.Nodes.Tx(r.Context(), func(tx *store.MeshTx) error {
		v, err := tx.VersionByNumber(n)
		if err != nil {
			return err
		}
		snap, err := store.DecodeSnapshot(v.Snapshot)
		if err != nil {
			return err
		}
		prev := store.Snapshot{Format: store.SnapshotFormat, Settings: map[string]string{}}
		if p, err := tx.VersionByNumber(n - 1); err == nil {
			if prev, err = store.DecodeSnapshot(p.Snapshot); err != nil {
				return err
			}
		}
		cur, err := tx.Snapshot()
		if err != nil {
			return err
		}
		d.Version = versionRow{Number: v.Version, When: u.stamp(v.CreatedAt), Actor: v.Actor, Summary: v.Summary, RolledBackFrom: v.RolledBackFrom}
		d.Change, d.Rollback = versions.Compare(prev, snap), versions.Compare(cur, snap)
		d.Current = d.Rollback.Empty()
		return nil
	})
	return d, err
}

func versionNumber(r *http.Request) (int64, error) {
	n, err := strconv.ParseInt(r.PathValue("n"), 10, 64)
	if err != nil || n <= 0 {
		return 0, &validate.Error{Field: "version", Value: r.PathValue("n"), Reason: "must be a positive number"}
	}
	return n, nil
}

func (u *ui) versionPage(w http.ResponseWriter, r *http.Request, s auth.Session) {
	u.renderVersion(w, r, s, http.StatusOK, page{})
}

func (u *ui) renderVersion(w http.ResponseWriter, r *http.Request, s auth.Session, status int, p page) {
	n, err := versionNumber(r)
	if err != nil {
		u.fail(w, r, s, err)
		return
	}
	d, err := u.versionView(r, n)
	if err != nil {
		u.fail(w, r, s, err)
		return
	}
	p.Data = d
	u.show(w, r, s, status, "version", "versions", fmt.Sprintf("Config v%d", n), p)
}

// rollback restores a version as a new one. A rollback that adds or removes
// nodes needs the confirmation box ticked (spec S8.3).
func (u *ui) rollback(w http.ResponseWriter, r *http.Request, s auth.Session) {
	n, err := versionNumber(r)
	if err != nil {
		u.fail(w, r, s, err)
		return
	}
	d, err := u.versionView(r, n)
	if err != nil {
		u.fail(w, r, s, err)
		return
	}
	if d.Rollback.ChangesMembership() && r.PostFormValue("confirm") != "yes" {
		u.renderVersion(w, r, s, http.StatusConflict, page{Error: "this rollback adds or removes nodes; tick the box to confirm it"})
		return
	}
	res, err := u.Nodes.Rollback(r.Context(), api.Change{Actor: actor(s).Name, UserID: s.UserID}, n)
	if err != nil {
		status, msg := message(err)
		u.renderVersion(w, r, s, status, page{Error: msg})
		return
	}
	if res.Version == 0 {
		u.renderVersion(w, r, s, http.StatusOK, page{Flash: "the mesh already matches this version; nothing changed"})
		return
	}
	http.Redirect(w, r, fmt.Sprintf("%s/versions/%d", u.base(), res.Version), http.StatusSeeOther)
}

type auditRow struct {
	When, Actor, Action, Target, Detail string
}

type auditData struct {
	Rows  []auditRow
	Older int64
}

const auditPerPage = 100

func (u *ui) auditPage(w http.ResponseWriter, r *http.Request, s auth.Session) {
	// An admin sees the control plane's own entries (users, meshes) too.
	var es []store.AuditEntry
	err := u.Nodes.Tx(r.Context(), func(tx *store.MeshTx) error {
		var err error
		if s.Admin {
			es, err = tx.AuditLog(auditPerPage, before(r))
		} else {
			es, err = tx.OwnAuditLog(auditPerPage, before(r))
		}
		return err
	})
	if err != nil {
		u.fail(w, r, s, err)
		return
	}
	d := auditData{}
	for _, e := range es {
		d.Rows = append(d.Rows, auditRow{When: u.stamp(e.At), Actor: e.Actor, Action: e.Action, Target: e.Target, Detail: e.Detail})
	}
	if len(es) == auditPerPage {
		d.Older = es[len(es)-1].ID
	}
	u.show(w, r, s, http.StatusOK, "audit", "audit", "Audit log", page{Data: d})
}

type settingsData struct {
	MeshCIDR    string
	CIDRLocked  bool
	DefaultMTU  string
	DefaultHub  string
	UpstreamDNS string
	DNSZone     string
	// ExtraDomains are the extra DNS domains, as the form shows them.
	ExtraDomains string
	Servers      []string
	// IPv6Prefix is the mesh's IPv6 prefix, "" for none, and IPv6Addressed
	// how many nodes have an IPv6 mesh address (M8).
	IPv6Prefix    string
	IPv6Addressed int
}

func (u *ui) settingsView(r *http.Request) (settingsData, error) {
	var d settingsData
	err := u.Nodes.Tx(r.Context(), func(tx *store.MeshTx) error {
		all, err := tx.Settings()
		if err != nil {
			return err
		}
		nodes, err := tx.Nodes()
		if err != nil {
			return err
		}
		d.MeshCIDR, d.DefaultMTU, d.UpstreamDNS = all[store.SettingMeshCIDR], all[store.SettingDefaultMTU], all[store.SettingUpstreamDNS]
		d.DNSZone = planner.Input{Zone: all[store.SettingDNSZone]}.ZoneName()
		d.ExtraDomains = strings.Join(store.SplitDomains(all[store.SettingDNSExtraDomains]), ", ")
		d.CIDRLocked = len(nodes) > 0
		for _, n := range nodes {
			if n.Role == string(proto.RoleServer) {
				d.Servers = append(d.Servers, n.Name)
				if strconv.FormatInt(n.ID, 10) == all[store.SettingDefaultHub] {
					d.DefaultHub = n.Name
				}
			}
		}
		return nil
	})
	return d, err
}

func (u *ui) settingsPage(w http.ResponseWriter, r *http.Request, s auth.Session) {
	u.renderSettings(w, r, s, http.StatusOK, page{})
}

func (u *ui) renderSettings(w http.ResponseWriter, r *http.Request, s auth.Session, status int, p page) {
	d, err := u.settingsView(r)
	if err != nil {
		u.fail(w, r, s, err)
		return
	}
	v6, err := u.Nodes.IPv6(r.Context())
	if err != nil {
		u.fail(w, r, s, err)
		return
	}
	d.IPv6Prefix, d.IPv6Addressed = v6.Prefix, v6.Addressed
	p.Data = d
	u.show(w, r, s, status, "settings", "settings", "Settings", p)
}

// saveSettings hands the form to the mesh's settings (api.SetSettings),
// which checks every field and applies them through the reconciler: a new
// default hub changes spokes' configs. The mesh CIDR can change only while
// the mesh has no nodes; a zone left out stays, and so do the extra domains
// (an empty field clears them).
func (u *ui) saveSettings(w http.ResponseWriter, r *http.Request, s auth.Session) {
	mtu, err := atoi("default_mtu", r.PostFormValue("default_mtu"))
	if err != nil {
		status, msg := message(err)
		u.renderSettings(w, r, s, status, page{Error: msg})
		return
	}
	var upstream []string
	for _, f := range strings.Split(r.PostFormValue("upstream_dns"), ",") {
		if f = strings.TrimSpace(f); f != "" {
			upstream = append(upstream, f)
		}
	}
	hub := r.PostFormValue("default_hub")
	req := api.SettingsRequest{DefaultMTU: &mtu, Upstream: &upstream, DefaultHub: &hub}
	if cidr := strings.TrimSpace(r.PostFormValue("mesh_cidr")); cidr != "" {
		req.MeshCIDR = &cidr
	}
	if zone := strings.TrimSpace(r.PostFormValue("dns_zone")); zone != "" {
		req.Zone = &zone
	}
	if _, ok := r.PostForm["dns_extra_domains"]; ok {
		fields := strings.FieldsFunc(r.PostFormValue("dns_extra_domains"), func(c rune) bool { return c == ',' || unicode.IsSpace(c) })
		req.Domains = &fields
	}
	res, err := u.Nodes.SetSettings(r.Context(), actor(s), req)
	if err != nil {
		status, msg := message(err)
		u.renderSettings(w, r, s, status, page{Error: msg})
		return
	}
	u.renderSettings(w, r, s, http.StatusOK, page{Flash: changed("Settings saved", api.NodeChange{Version: res.Version})})
}

// saveIPv6 sets, generates or gives up the mesh's IPv6 prefix (POST
// /settings/ipv6, M8); setting it gives every dual node an IPv6 address.
func (u *ui) saveIPv6(w http.ResponseWriter, r *http.Request, s auth.Session) {
	v, err := u.Nodes.SetIPv6(r.Context(), actor(s), api.IPv6Request{Prefix: r.PostFormValue("ipv6_prefix")})
	if err != nil {
		status, msg := message(err)
		u.renderSettings(w, r, s, status, page{Error: msg})
		return
	}
	what := "IPv6 given up"
	if v.Prefix != "" {
		what = fmt.Sprintf("IPv6 prefix %s; %d node(s) have an IPv6 mesh address", v.Prefix, v.Addressed)
	}
	u.renderSettings(w, r, s, http.StatusOK, page{Flash: changed(what, api.NodeChange{Version: v.Version})})
}

type userRow struct {
	Name     string
	TOTP     bool
	Disabled bool
	Created  string
	Self     bool
	// Admin manages everything; anyone else Meshes, the meshes granted to
	// them (amendment A49). Grantable are the meshes they do not have yet.
	Admin     bool
	Meshes    []string
	Grantable []string
}

// usersData is the Users page: every user, and the meshes the add form
// offers.
type usersData struct {
	Rows   []userRow
	Meshes []string
}

func (u *ui) usersPage(w http.ResponseWriter, r *http.Request, s auth.Session) {
	u.renderUsers(w, r, s, http.StatusOK, page{})
}

func (u *ui) renderUsers(w http.ResponseWriter, r *http.Request, s auth.Session, status int, p page) {
	users, err := u.Auth.Users(r.Context())
	if err != nil {
		u.fail(w, r, s, err)
		return
	}
	grants, err := u.Auth.Grants(r.Context())
	if err != nil {
		u.fail(w, r, s, err)
		return
	}
	meshes, err := u.meshList(r, s)
	if err != nil {
		u.fail(w, r, s, err)
		return
	}
	var rows []userRow
	for _, x := range users {
		row := userRow{Name: x.Username, TOTP: x.TOTPEnrolled, Disabled: x.Disabled, Created: u.stamp(x.CreatedAt),
			Self: x.ID == s.UserID, Admin: x.Admin, Meshes: grants[x.ID]}
		for _, m := range meshes {
			if !slices.Contains(row.Meshes, m.Name) {
				row.Grantable = append(row.Grantable, m.Name)
			}
		}
		rows = append(rows, row)
	}
	d := usersData{Rows: rows}
	for _, m := range meshes {
		d.Meshes = append(d.Meshes, m.Name)
	}
	p.Data = d
	u.show(w, r, s, status, "users", "users", "Users", p)
}

func (u *ui) addUser(w http.ResponseWriter, r *http.Request, s auth.Session) {
	name, pw := strings.TrimSpace(r.PostFormValue("username")), r.PostFormValue("password")
	var (
		x   store.User
		err error
	)
	a := auth.Access{Admin: r.PostFormValue("admin") != "", Mesh: r.PostFormValue("mesh")}
	if pw != r.PostFormValue("password2") {
		err = &validate.Error{Field: "password", Reason: "the two passwords differ"}
	} else {
		x, err = u.Auth.AddUser(r.Context(), actor(s).Name, name, pw, a)
	}
	if err != nil {
		status, msg := message(err)
		u.renderUsers(w, r, s, status, page{Error: msg})
		return
	}
	role := "; grant them meshes below"
	switch {
	case x.Admin:
		role = " as an admin"
	case a.Mesh != "":
		role = ", who manages mesh " + a.Mesh
	}
	u.renderUsers(w, r, s, http.StatusOK, page{Flash: "Added " + name + role + "; they set up an authenticator at their first sign-in"})
}

// setAdmin makes a user an admin (admin=on) or takes it away; your own role
// stays, so an admin cannot lock themself out.
func (u *ui) setAdmin(w http.ResponseWriter, r *http.Request, s auth.Session) {
	on := r.PostFormValue("admin") == "on"
	done := map[bool]string{true: "Made an admin:", false: "No longer an admin:"}[on]
	u.userChange(w, r, s, done, func(name string) error {
		if name == s.Username {
			return &validate.Error{Field: "user", Reason: "you cannot change your own role; another admin can"}
		}
		return u.Auth.SetAdmin(r.Context(), actor(s).Name, name, on)
	})
}

func (u *ui) grantMesh(w http.ResponseWriter, r *http.Request, s auth.Session) {
	mesh := r.PostFormValue("mesh")
	u.userChange(w, r, s, "Granted mesh "+mesh+" to", func(name string) error { return u.Auth.Grant(r.Context(), actor(s).Name, name, mesh) })
}

func (u *ui) revokeMesh(w http.ResponseWriter, r *http.Request, s auth.Session) {
	mesh := r.PostFormValue("mesh")
	u.userChange(w, r, s, "Took mesh "+mesh+" away from", func(name string) error { return u.Auth.Revoke(r.Context(), actor(s).Name, name, mesh) })
}

func (u *ui) disableUser(w http.ResponseWriter, r *http.Request, s auth.Session) {
	u.userChange(w, r, s, "Disabled", func(name string) error {
		if name == s.Username {
			return &validate.Error{Field: "user", Reason: "you cannot disable your own account"}
		}
		return u.Auth.SetDisabled(r.Context(), actor(s).Name, name, true)
	})
}

func (u *ui) enableUser(w http.ResponseWriter, r *http.Request, s auth.Session) {
	u.userChange(w, r, s, "Enabled", func(name string) error { return u.Auth.SetDisabled(r.Context(), actor(s).Name, name, false) })
}

func (u *ui) resetUser(w http.ResponseWriter, r *http.Request, s auth.Session) {
	u.userChange(w, r, s, "Reset the authenticator of", func(name string) error {
		if name == s.Username {
			return &validate.Error{Field: "user", Reason: "reset your own authenticator with `meshctl user reset-totp` on the server"}
		}
		return u.Auth.ResetTOTP(r.Context(), actor(s).Name, name)
	})
}

func (u *ui) userChange(w http.ResponseWriter, r *http.Request, s auth.Session, done string, change func(string) error) {
	name := r.PathValue("name")
	if err := change(name); err != nil {
		status, msg := message(err)
		u.renderUsers(w, r, s, status, page{Error: msg})
		return
	}
	u.renderUsers(w, r, s, http.StatusOK, page{Flash: done + " " + name})
}

func (u *ui) accountPage(w http.ResponseWriter, r *http.Request, s auth.Session) {
	u.show(w, r, s, http.StatusOK, "account", "account", "Your account", page{})
}

// changePassword ends every session of the user, this one included.
func (u *ui) changePassword(w http.ResponseWriter, r *http.Request, s auth.Session) {
	next := r.PostFormValue("new")
	var err error
	if next != r.PostFormValue("new2") {
		err = &validate.Error{Field: "password", Reason: "the two new passwords differ"}
	} else {
		err = u.Auth.ChangePassword(r.Context(), s, r.PostFormValue("current"), next)
	}
	if err != nil {
		status, msg := message(err)
		u.show(w, r, s, status, "account", "account", "Your account", page{Error: msg})
		return
	}
	auth.ClearCookie(w)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// exportYAML downloads the desired state, as `meshctl export` writes it.
func (u *ui) exportYAML(w http.ResponseWriter, r *http.Request, s auth.Session) {
	var snap store.Snapshot
	if err := u.Nodes.Tx(r.Context(), func(tx *store.MeshTx) error {
		var err error
		snap, err = tx.Snapshot()
		return err
	}); err != nil {
		u.fail(w, r, s, err)
		return
	}
	var b bytes.Buffer
	if err := export.Write(&b, export.FromSnapshot(snap)); err != nil {
		u.fail(w, r, s, err)
		return
	}
	w.Header().Set("Content-Type", "application/yaml")
	w.Header().Set("Content-Disposition", `attachment; filename="`+u.mesh.Name+"-"+u.Now().Format("2006-01-02")+`.yaml"`)
	_, _ = w.Write(b.Bytes()) // the browser went away; nothing to do
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
