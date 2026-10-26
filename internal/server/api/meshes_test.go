package api

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/hafio/vpn-mesh-utility/internal/proto"
	"github.com/hafio/vpn-mesh-utility/internal/server/store"
	"github.com/hafio/vpn-mesh-utility/internal/testfix"
)

// addLab creates a second mesh, "lab" in slot 1, on 10.78.0.0/16 with
// zone lab.
func (h *harness) addLab(t *testing.T) store.Mesh {
	t.Helper()
	m := store.Mesh{Name: "lab", Slot: 1}
	err := h.st.Tx(context.Background(), func(tx *store.Tx) error {
		if err := tx.InsertMesh(&m); err != nil {
			return err
		}
		for k, v := range map[string]string{store.SettingMeshCIDR: "10.78.0.0/16", store.SettingDNSZone: "lab",
			store.SettingDefaultMTU: "1420", store.SettingUpstreamDNS: "1.1.1.1", store.SettingDefaultHub: ""} {
			if err := tx.In(m.ID).SetSetting(k, v); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return m
}

// A request that names no mesh means the only one; with two it must name
// one, and with none there is nothing to manage (amendment A49).
func TestSoleMesh(t *testing.T) {
	h := newHarness(t)
	h.seedServers(t)
	if l, err := h.nodes().List(context.Background()); err != nil || len(l.Nodes) != 3 {
		t.Fatalf("the sole mesh's nodes = %+v, %v", l, err)
	}
	lab := h.addLab(t)
	_, err := h.nodes().List(context.Background())
	refused(t, err, http.StatusBadRequest, "2 meshes (default, lab): name one with --mesh")
	if code, eb := adminCall(t, h, http.MethodGet, AdminNodes, nil, nil); code != 400 || !strings.Contains(eb.Message, "--mesh") {
		t.Errorf("the admin API with two meshes = %d %+v", code, eb)
	}
	if l, err := h.nodes().In(lab.ID).List(context.Background()); err != nil || len(l.Nodes) != 0 {
		t.Errorf("the lab's nodes = %+v, %v", l, err)
	}
	if _, err := h.nodes().Add(context.Background(), cli(), NodeAddRequest{Name: "nas", Role: "spoke"}); err == nil {
		t.Error("a write that names no mesh went through with two meshes")
	}

	empty := newHarness(t)
	if err := empty.st.Tx(context.Background(), func(tx *store.Tx) error { return tx.DeleteMesh(store.FirstMesh) }); err != nil {
		t.Fatal(err)
	}
	_, err = empty.nodes().List(context.Background())
	refused(t, err, http.StatusNotFound, "no mesh: create one with `meshctl mesh add <name>`")
	if _, err := empty.rec.Apply(context.Background(), Change{Actor: "test"}, nil); err == nil || !strings.Contains(err.Error(), "names no mesh") {
		t.Errorf("a change without a mesh = %v", err)
	}
}

// A node of another mesh joins, gets a config and a mesh view of its own
// mesh only, and its changes are versioned, audited and pushed there.
func TestMeshesAreApartOnTheAgentAPI(t *testing.T) {
	h := meshHarness(t)
	lab := h.addLab(t)
	labNodes := h.nodes().In(lab.ID)
	before := h.latestVersion(t)
	events, _, _, _, cancel := h.hub.Subscribe(testfix.VPS1.ID, store.FirstMesh)
	defer cancel()

	// The same name as a node of the first mesh, the laptop's (unused) keys.
	ch, err := labNodes.Add(context.Background(), cli(), NodeAddRequest{Name: "vps-1", Role: "server", Endpoint: "198.51.100.20:51820"})
	if err != nil || ch.Enrollment == nil {
		t.Fatalf("add the lab's vps-1: %+v, %v", ch, err)
	}
	if _, err := labNodes.Add(context.Background(), cli(), NodeAddRequest{Name: "nas", Role: "spoke"}); err != nil {
		t.Fatal(err)
	}
	f := testfix.Node{Name: "vps-1", Role: proto.RoleServer, Key: testfix.Laptop.Key}
	w, eb := postJoin(h, joinBody(t, ch.Enrollment.Token, f, nil))
	if w.Code != 200 {
		t.Fatalf("join the lab: %d %+v", w.Code, eb)
	}
	var joined proto.JoinResponse
	if err := json.Unmarshal(w.Body.Bytes(), &joined); err != nil {
		t.Fatal(err)
	}
	if joined.NodeID == testfix.VPS1.ID || joined.ConfigVersion != 1 || joined.Mesh != "lab" || joined.MeshUUID != lab.UUID || joined.Slot != 1 {
		t.Fatalf("the lab's join = %+v", joined)
	}
	f.ID = joined.NodeID

	w, eb = serve(h.pub, signed(http.MethodGet, proto.PathConfig, nil, f, testfix.Epoch.Add(time.Second)))
	var cfg proto.NodeConfig
	if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &cfg) != nil || len(cfg.Peers) != 0 || cfg.Version != 1 || cfg.Node.Slot != 1 {
		t.Fatalf("the lab's config: %d %+v %+v", w.Code, eb, cfg)
	}
	w, eb = serve(h.pub, signed(http.MethodGet, proto.PathMesh, nil, f, testfix.Epoch.Add(2*time.Second)))
	var view proto.MeshView
	if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &view) != nil || len(view.Nodes) != 2 || view.Version != 1 {
		t.Fatalf("the lab's mesh view: %d %+v %+v", w.Code, eb, view)
	}

	if v := h.latestVersion(t); v != before {
		t.Errorf("the lab's join made the first mesh's v%d (was v%d)", v, before)
	}
	for _, p := range h.config(t, "vps-1").Peers {
		if p.PublicKey == testfix.WGKey(testfix.Laptop.Key) {
			t.Errorf("the first mesh's vps-1 peers the lab's vps-1: %+v", p)
		}
	}
	select {
	case ev := <-events:
		t.Errorf("the first mesh's vps-1 got the lab's event %+v", ev)
	default:
	}
	var audit []store.AuditEntry
	if err := h.st.MeshTx(context.Background(), lab.ID, func(tx *store.MeshTx) error {
		var err error
		audit, err = tx.AuditLog(10, 0)
		return err
	}); err != nil {
		t.Fatal(err)
	}
	if len(audit) == 0 || audit[0].Action != "node.join" || audit[0].MeshID != lab.ID {
		t.Errorf("the lab's audit = %+v", audit)
	}
	if n := h.audits(t, "node.join"); n != 0 {
		t.Errorf("the first mesh shows %d of the lab's joins", n)
	}

	// The slot's names (amendment A49): the server's default port, and a
	// container network's runtime network and bridge.
	if d, err := labNodes.Show(context.Background(), "vps-1"); err != nil || d.Node.ListenPort != 51821 {
		t.Fatalf("the lab's vps-1: %+v, %v", d.Node, err)
	}
	if _, err := labNodes.Add(context.Background(), cli(), NodeAddRequest{Name: "box", Role: "spoke", Containers: true}); err != nil {
		t.Fatal(err)
	}
	if d, err := labNodes.Show(context.Background(), "box"); err != nil || d.Containers == nil ||
		d.Containers.Name != "meshctl-mesh1" || d.Containers.BridgeName != "meshctl1" {
		t.Fatalf("the lab's box: %+v, %v", d.Containers, err)
	}
}

// The service's Tx, Apply and Rollback (the web UI's way in) work in its
// mesh; naming none with two meshes is refused.
func TestNodesServiceInAMesh(t *testing.T) {
	h := meshHarness(t)
	ctx := context.Background()
	lab := h.addLab(t)
	s := h.nodes().In(lab.ID)
	setMTU := func(tx *store.MeshTx, _ *Change) error { return tx.SetSetting(store.SettingDefaultMTU, "1380") }
	if _, err := s.Apply(ctx, Change{Actor: "test", Action: "settings.set"}, setMTU); err != nil {
		t.Fatal(err)
	}
	var labMTU, firstMTU string
	if err := s.Tx(ctx, func(tx *store.MeshTx) error {
		if tx.MeshID() != lab.ID {
			t.Errorf("the lab's service reads mesh %d", tx.MeshID())
		}
		var err error
		labMTU, err = tx.Setting(store.SettingDefaultMTU)
		return err
	}); err != nil {
		t.Fatal(err)
	}
	if err := h.st.MeshTx(ctx, store.FirstMesh, func(tx *store.MeshTx) error {
		var err error
		firstMTU, err = tx.Setting(store.SettingDefaultMTU)
		return err
	}); err != nil {
		t.Fatal(err)
	}
	if labMTU != "1380" || firstMTU != "1420" {
		t.Errorf("default MTU: the lab %q, the first mesh %q", labMTU, firstMTU)
	}
	_, err := s.Rollback(ctx, Change{Actor: "test"}, 999)
	refused(t, err, http.StatusNotFound, "v999")
	_, err = h.nodes().Rollback(ctx, Change{Actor: "test"}, 1)
	refused(t, err, http.StatusBadRequest, "--mesh")
	_, err = h.nodes().Apply(ctx, Change{Actor: "test"}, setMTU)
	refused(t, err, http.StatusBadRequest, "--mesh")
	err = h.nodes().Tx(ctx, func(*store.MeshTx) error { return nil })
	refused(t, err, http.StatusBadRequest, "--mesh")
}

// Meshes are added with their defaults or the values given, listed, shown,
// renamed and changed, and removed once they have no nodes (amendment A49).
func TestMeshAdmin(t *testing.T) {
	h := newHarness(t)
	h.seedServers(t)
	ctx := context.Background()
	s := h.nodes()

	// The next free /16, the name as the zone, a unique local /64, the next
	// slot and the first mesh's other settings.
	lab, err := s.AddMesh(ctx, cli(), MeshAddRequest{Name: "lab"})
	m := lab.Mesh
	if err != nil || m.Slot != 1 || m.MeshCIDR != "10.78.0.0/16" || m.Zone != "lab" || !strings.HasPrefix(m.MeshCIDR6, "fd") ||
		m.DefaultMTU != 1420 || strings.Join(m.Upstream, ",") != "1.1.1.1,9.9.9.9" || m.Nodes != 0 || m.UUID == "" || lab.Version != 0 {
		t.Fatalf("AddMesh(lab) = %+v, %v", lab, err)
	}
	home, err := s.AddMesh(ctx, cli(), MeshAddRequest{Name: "home", MeshCIDR: "10.90.0.0/20", Zone: "home.arpa", IPv6: IPv6Off})
	if m := home.Mesh; err != nil || m.Slot != 2 || m.MeshCIDR != "10.90.0.0/20" || m.Zone != "home.arpa" || m.MeshCIDR6 != "" {
		t.Fatalf("AddMesh(home) = %+v, %v", home, err)
	}
	if seg, err := s.Named("home").Segment(ctx); err != nil || len(seg.Config.Policies) != 1 || seg.Config.Policies[0].Name != "allow-all" {
		t.Fatalf("a new mesh's firewall = %+v, %v", seg.Config, err)
	}
	if v6, err := s.Named("home").IPv6(ctx); err != nil || v6.Prefix != "" {
		t.Fatalf("IPv6 off = %+v, %v", v6, err)
	}
	var audit []store.AuditEntry
	if err := h.st.Tx(ctx, func(tx *store.Tx) error {
		lab, err := tx.MeshByName("lab")
		if err != nil {
			return err
		}
		audit, err = tx.In(lab.ID).AuditLog(10, 0)
		return err
	}); err != nil || len(audit) == 0 || audit[0].Action != "mesh.add" || audit[0].Target != "lab" {
		t.Fatalf("the lab's audit = %+v, %v", audit, err)
	}

	// Refused before the database, or as a conflict with nothing written.
	h.db.n.Store(0)
	for _, c := range []struct {
		req  MeshAddRequest
		want string
	}{
		{MeshAddRequest{Name: "Lab!"}, "name"},
		{MeshAddRequest{Name: "local"}, "dns_zone"},
		{MeshAddRequest{Name: "big", MeshCIDR: "10.0.0.0/8"}, "mesh_cidr"},
		{MeshAddRequest{Name: "wide", IPv6: "fd00::/32"}, "/48"},
		{MeshAddRequest{Name: "v6", IPv6: "10.0.0.0/16"}, "ipv6"},
	} {
		_, err := s.AddMesh(ctx, cli(), c.req)
		refused(t, err, http.StatusBadRequest, c.want)
	}
	if n := h.db.n.Load(); n != 0 {
		t.Fatalf("a refused mesh ran %d transactions", n)
	}
	for _, c := range []struct {
		req  MeshAddRequest
		want string
	}{
		{MeshAddRequest{Name: "lab"}, `mesh "lab" exists already`},
		{MeshAddRequest{Name: "half", MeshCIDR: "10.77.128.0/17"}, "IPv4 network 10.77.128.0/17 overlaps mesh default's 10.77.0.0/16"},
		{MeshAddRequest{Name: "twin", Zone: "mesh"}, "DNS domain mesh is mesh default's mesh"},
		{MeshAddRequest{Name: "inner", Zone: "x.lab"}, "DNS domain x.lab is mesh lab's lab"},
		{MeshAddRequest{Name: "same6", IPv6: m.MeshCIDR6}, "IPv6 prefix"},
	} {
		_, err := s.AddMesh(ctx, cli(), c.req)
		refused(t, err, http.StatusConflict, c.want)
	}
	if l, err := s.Meshes(ctx); err != nil || len(l.Meshes) != 3 || l.Meshes[0].Name != "default" || l.Meshes[0].Nodes != 3 ||
		l.Meshes[1].Name != "lab" || l.Meshes[2].Name != "home" {
		t.Fatalf("Meshes = %+v, %v", l, err)
	}

	// A rename keeps the uuid and the slot; settings go through the
	// reconciler, so a new zone is a config version for the servers.
	up, err := s.SetMesh(ctx, cli(), "home", MeshSetRequest{Name: ptr("house"), SettingsRequest: SettingsRequest{Zone: ptr("house")}})
	if err != nil || up.Mesh.Name != "house" || up.Mesh.Slot != 2 || up.Mesh.UUID != home.Mesh.UUID || up.Mesh.Zone != "house" {
		t.Fatalf("SetMesh(home) = %+v, %v", up, err)
	}
	if _, err := s.ShowMesh(ctx, "home"); err == nil {
		t.Fatal("the old name still shows")
	}
	before := h.latestVersion(t)
	up, err = s.SetMesh(ctx, cli(), "default", MeshSetRequest{SettingsRequest: SettingsRequest{Zone: ptr("corp"), DefaultHub: ptr("vps-2")}})
	if err != nil || up.Version != before+1 || up.Mesh.Zone != "corp" || up.Mesh.DefaultHub != "vps-2" {
		t.Fatalf("SetMesh(default) = %+v, %v", up, err)
	}
	for _, c := range []struct {
		name, mesh string
		req        MeshSetRequest
		status     int
		want       string
	}{
		{"nothing", "lab", MeshSetRequest{}, http.StatusBadRequest, "nothing to change"},
		{"bad mesh", "Lab!", MeshSetRequest{Name: ptr("x")}, http.StatusBadRequest, "mesh"},
		{"bad name", "lab", MeshSetRequest{Name: ptr("X")}, http.StatusBadRequest, "name"},
		{"no mesh", "none", MeshSetRequest{Name: ptr("x")}, http.StatusNotFound, `no mesh "none"`},
		{"taken name", "lab", MeshSetRequest{Name: ptr("house")}, http.StatusConflict, `mesh "house" exists already`},
		{"taken zone", "lab", MeshSetRequest{SettingsRequest: SettingsRequest{Zone: ptr("corp")}}, http.StatusConflict, "mesh default's corp"},
		{"domain in a zone", "lab", MeshSetRequest{SettingsRequest: SettingsRequest{Domains: &[]string{"a.house"}}}, http.StatusConflict,
			"DNS domain a.house is mesh house's house"},
		{"cidr with nodes", "default", MeshSetRequest{SettingsRequest: SettingsRequest{MeshCIDR: ptr("10.79.0.0/16")}}, http.StatusBadRequest,
			"no nodes"},
		{"overlapping cidr", "lab", MeshSetRequest{SettingsRequest: SettingsRequest{MeshCIDR: ptr("10.90.0.0/16")}}, http.StatusConflict,
			"overlaps mesh house's 10.90.0.0/20"},
		{"bad mtu", "lab", MeshSetRequest{SettingsRequest: SettingsRequest{DefaultMTU: ptr(100)}}, http.StatusBadRequest, "default_mtu"},
		{"no upstream", "lab", MeshSetRequest{SettingsRequest: SettingsRequest{Upstream: &[]string{}}}, http.StatusBadRequest, "one to four"},
		{"bad upstream", "lab", MeshSetRequest{SettingsRequest: SettingsRequest{Upstream: &[]string{"dns.example"}}}, http.StatusBadRequest,
			"upstream_dns"},
		{"bad zone", "lab", MeshSetRequest{SettingsRequest: SettingsRequest{Zone: ptr("a..b")}}, http.StatusBadRequest, "dns_zone"},
		{"server's domain", "lab", MeshSetRequest{SettingsRequest: SettingsRequest{Domains: &[]string{"example.com"}}}, http.StatusBadRequest,
			"server's address"},
		{"bad hub", "lab", MeshSetRequest{SettingsRequest: SettingsRequest{DefaultHub: ptr("Hub!")}}, http.StatusBadRequest, "default_hub"},
		{"hub of another mesh", "lab", MeshSetRequest{SettingsRequest: SettingsRequest{DefaultHub: ptr("vps-1")}}, http.StatusBadRequest,
			"server of this mesh"},
	} {
		_, err := s.SetMesh(ctx, cli(), c.mesh, c.req)
		if got, text := ErrorMessage(err); err == nil || got != c.status || !strings.Contains(text, c.want) {
			t.Errorf("%s: %v (%d), want %d mentioning %q", c.name, err, got, c.status, c.want)
		}
	}

	// A mesh with nodes stays; without them it goes, as the control
	// plane's audit entry.
	if _, err := s.Named("lab").Add(ctx, cli(), NodeAddRequest{Name: "nas", Role: "spoke"}); err != nil {
		t.Fatal(err)
	}
	refused(t, s.RemoveMesh(ctx, cli(), "lab"), http.StatusConflict, "still has 1 node(s) (nas)")
	refused(t, s.RemoveMesh(ctx, cli(), "Lab!"), http.StatusBadRequest, "mesh")
	refused(t, s.RemoveMesh(ctx, cli(), "none"), http.StatusNotFound, "none")
	if _, err := s.Named("lab").Remove(ctx, cli(), "nas"); err != nil {
		t.Fatal(err)
	}
	if err := s.RemoveMesh(ctx, cli(), "lab"); err != nil {
		t.Fatal(err)
	}
	if l, err := s.Meshes(ctx); err != nil || len(l.Meshes) != 2 || h.audits(t, "mesh.rm") != 1 {
		t.Fatalf("after removing the lab: %+v, %v, %d mesh.rm", l, err, h.audits(t, "mesh.rm"))
	}
}

// A mesh moved onto another's addresses or names by any write -- the IPv6
// prefix here -- is refused by the reconciler, and nothing changes.
func TestMeshesStayApart(t *testing.T) {
	h := meshHarness(t)
	ctx := context.Background()
	if _, err := h.nodes().SetIPv6(ctx, cli(), IPv6Request{Prefix: testfix.MeshCIDR6}); err != nil {
		t.Fatal(err)
	}
	lab := h.addLab(t)
	before := h.latestVersion(t)
	_, err := h.nodes().In(lab.ID).SetIPv6(ctx, cli(), IPv6Request{Prefix: testfix.MeshCIDR6})
	refused(t, err, http.StatusConflict, "IPv6 prefix "+testfix.MeshCIDR6+" overlaps mesh default's")
	if v6, err := h.nodes().In(lab.ID).IPv6(ctx); err != nil || v6.Prefix != "" || h.latestVersion(t) != before {
		t.Fatalf("after the refusal: %+v, %v", v6, err)
	}
}

// Every admin route but the meshes' own takes ?mesh=; a name that is not a
// mesh name is refused before the database, an unknown one says how to list
// them, and the admin client names its mesh on each request.
func TestAdminRoutesNameTheirMesh(t *testing.T) {
	h := newHarness(t)
	h.seedServers(t)
	h.addLab(t)
	var l NodeList
	if code, eb := adminCall(t, h, http.MethodGet, AdminNodes+"?mesh=lab", nil, &l); code != 200 || len(l.Nodes) != 0 {
		t.Fatalf("the lab's nodes = %d %+v %+v", code, eb, l)
	}
	if code, eb := adminCall(t, h, http.MethodGet, AdminNodes+"?mesh=default", nil, &l); code != 200 || len(l.Nodes) != 3 {
		t.Fatalf("the first mesh's nodes = %d %+v %+v", code, eb, l)
	}
	if code, eb := adminCall(t, h, http.MethodGet, AdminNodes+"?mesh=none", nil, nil); code != 404 || !strings.Contains(eb.Message, "meshctl mesh list") {
		t.Errorf("an unknown mesh = %d %+v", code, eb)
	}
	h.db.n.Store(0)
	if code, eb := adminCall(t, h, http.MethodGet, AdminNodes+"?mesh=Lab%21", nil, nil); code != 400 || !strings.Contains(eb.Message, "mesh") {
		t.Errorf("a bad mesh name = %d %+v", code, eb)
	}
	if n := h.db.n.Load(); n != 0 {
		t.Fatalf("a bad mesh name ran %d transactions", n)
	}
	var seg SegmentView
	if code, eb := adminCall(t, h, http.MethodGet, AdminSegmentation+"?mesh=lab", nil, &seg); code != 200 || len(seg.Nodes) != 0 {
		t.Errorf("the lab's firewall = %d %+v", code, eb)
	}

	dir, err := os.MkdirTemp("", "mad")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	sock := filepath.Join(dir, "admin.sock")
	srv := httptest.NewUnstartedServer(h.admin)
	ln, err := net.Listen("unix", sock)
	if err != nil {
		t.Fatal(err)
	}
	srv.Listener = ln
	srv.Start()
	defer srv.Close()
	c, ctx := NewAdminClient(sock), context.Background()
	if _, err := c.List(ctx); err == nil || !strings.Contains(err.Error(), "--mesh") {
		t.Fatalf("List with two meshes = %v", err)
	}
	if ch, err := c.In("lab").Add(ctx, NodeAddRequest{Name: "nas", Role: "spoke"}); err != nil || !strings.HasPrefix(ch.Node.MeshIP, "10.78.") {
		t.Fatalf("Add in the lab = %+v, %v", ch, err)
	}
	if r, err := c.In("lab").Reach(ctx, ReachRequest{From: "nas", To: "10.78.0.9", Service: "tcp/22"}); err != nil || !strings.HasPrefix(r.From, "10.78.") {
		t.Fatalf("Reach in the lab = %+v, %v", r, err)
	}
	if u, err := c.AddMesh(ctx, MeshAddRequest{Name: "home"}); err != nil || u.Mesh.Slot != 2 {
		t.Fatalf("AddMesh = %+v, %v", u, err)
	}
	if l, err := c.Meshes(ctx); err != nil || len(l.Meshes) != 3 || l.Meshes[1].Nodes != 1 {
		t.Fatalf("Meshes = %+v, %v", l, err)
	}
	if u, err := c.SetMesh(ctx, "home", MeshSetRequest{Name: ptr("house")}); err != nil || u.Mesh.Name != "house" {
		t.Fatalf("SetMesh = %+v, %v", u, err)
	}
	if m, err := c.ShowMesh(ctx, "house"); err != nil || m.Slot != 2 {
		t.Fatalf("ShowMesh = %+v, %v", m, err)
	}
	if err := c.RemoveMesh(ctx, "house"); err != nil {
		t.Fatal(err)
	}
	if err := c.RemoveMesh(ctx, "lab"); err == nil {
		t.Fatal("a mesh with a node was removed")
	}
}

// SetSettings (the web UI's Settings page) changes its mesh's settings as a
// config version, refuses a bad value before the database, and names no mesh
// it cannot tell.
func TestSetSettings(t *testing.T) {
	h := meshHarness(t)
	ctx := context.Background()
	before := h.latestVersion(t)
	if ch, err := h.nodes().SetSettings(ctx, Actor{Name: "user:ops", UserID: 1}, SettingsRequest{Zone: ptr("corp")}); err != nil || ch.Version != before+1 {
		t.Fatalf("a new zone = %+v, %v", ch, err)
	}
	h.db.n.Store(0)
	_, err := h.nodes().SetSettings(ctx, cli(), SettingsRequest{DefaultMTU: ptr(100)})
	refused(t, err, http.StatusBadRequest, "default_mtu")
	if n := h.db.n.Load(); n != 0 {
		t.Fatalf("a refused setting ran %d transactions", n)
	}
	h.addLab(t)
	_, err = h.nodes().SetSettings(ctx, cli(), SettingsRequest{DefaultMTU: ptr(1380)})
	refused(t, err, http.StatusBadRequest, "--mesh")
}
