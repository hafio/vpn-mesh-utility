package api

import (
	"context"
	"encoding/json"
	"net/http"
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
	events, _, _, cancel := h.hub.Subscribe(testfix.VPS1.ID, store.FirstMesh)
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
	if joined.NodeID == testfix.VPS1.ID || joined.ConfigVersion != 1 {
		t.Fatalf("the lab's join = %+v", joined)
	}
	f.ID = joined.NodeID

	w, eb = serve(h.pub, signed(http.MethodGet, proto.PathConfig, nil, f, testfix.Epoch.Add(time.Second)))
	var cfg proto.NodeConfig
	if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &cfg) != nil || len(cfg.Peers) != 0 || cfg.Version != 1 {
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
