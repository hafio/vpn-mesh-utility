package api

import (
	"context"
	"net/http"
	"slices"
	"strings"
	"testing"

	"github.com/hafio/vpn-mesh-utility/internal/proto"
	"github.com/hafio/vpn-mesh-utility/internal/server/store"
	"github.com/hafio/vpn-mesh-utility/internal/testfix"
)

func (h *harness) nodes() *Nodes {
	return NewNodes(Deps{DB: h.db, Hub: h.hub, Reconciler: h.rec, Now: h.clock.Now, ServerURL: testURL, Version: "v1.0.0"})
}

// meshHarness is the three servers, the enrolled NAS and a laptop that has
// not joined yet.
func meshHarness(t *testing.T) *harness {
	h := newHarness(t)
	h.seed(t, storeNode(testfix.VPS1, true), storeNode(testfix.VPS2, true), storeNode(testfix.VPS3, true),
		storeNode(testfix.NAS, true), storeNode(testfix.Laptop, false))
	return h
}

// config is what the server computes for a node now.
func (h *harness) config(t *testing.T, name string) proto.NodeConfig {
	t.Helper()
	// Looked up first: the store has one connection, so a Tx inside a Tx
	// waits forever.
	id := h.node(t, name).ID
	var cfg *proto.NodeConfig
	if err := h.st.MeshTx(context.Background(), store.FirstMesh, func(tx *store.MeshTx) error {
		var err error
		_, cfg, err = configFor(tx, id)
		return err
	}); err != nil || cfg == nil {
		t.Fatalf("config of %s: %v, %v", name, cfg, err)
	}
	return *cfg
}

func refused(t *testing.T, err error, status int, msg string) {
	t.Helper()
	got, text := ErrorMessage(err)
	if err == nil || got != status || !strings.Contains(text, msg) {
		t.Fatalf("err = %v (%d %q), want %d mentioning %q", err, got, text, status, msg)
	}
}

func TestCustomDNSRecords(t *testing.T) {
	h := meshHarness(t)
	s, ctx, by := h.nodes(), context.Background(), Actor{Name: "user:ops", UserID: 1}

	ch, err := s.AddRecord(ctx, by, RecordRequest{Name: "files", Node: "nas"})
	if err != nil || ch.Version == 0 {
		t.Fatalf("alias of the nas = %+v, %v", ch, err)
	}
	if ch, err = s.AddRecord(ctx, by, RecordRequest{Name: "printer", IP: "10.77.1.200"}); err != nil || ch.Version == 0 {
		t.Fatalf("fixed record = %+v, %v", ch, err)
	}
	// An alias of a node that is not on the mesh changes no config yet.
	if ch, err = s.AddRecord(ctx, by, RecordRequest{Name: "work", Node: "laptop"}); err != nil || ch.Version != 0 {
		t.Fatalf("alias of the unjoined laptop = %+v, %v", ch, err)
	}
	for _, n := range []string{"vps-1", "nas"} {
		var got []string
		for _, r := range h.config(t, n).DNS.Records {
			got = append(got, r.Name+"="+r.IP)
		}
		if want := "files=10.77.1.1 nas=10.77.1.1 printer=10.77.1.200 vps-1=10.77.0.1 vps-2=10.77.0.2 vps-3=10.77.0.3"; strings.Join(got, " ") != want {
			t.Errorf("%s serves %v, want %s", n, got, want)
		}
	}
	v, err := s.DNS(ctx)
	if err != nil {
		t.Fatal(err)
	}
	var rows []string
	for _, e := range v.Records {
		rows = append(rows, strings.Join([]string{e.Name, e.IP, e.Node, map[bool]string{true: "custom"}[e.Custom],
			map[bool]string{true: "live"}[e.Live]}, ","))
	}
	want := []string{"laptop,10.77.2.1,laptop,,", "nas,10.77.1.1,nas,,live", "vps-1,10.77.0.1,vps-1,,live", "vps-2,10.77.0.2,vps-2,,live",
		"vps-3,10.77.0.3,vps-3,,live", "files,10.77.1.1,nas,custom,live", "printer,10.77.1.200,,custom,live", "work,10.77.2.1,laptop,custom,"}
	if !slices.Equal(rows, want) {
		t.Errorf("DNS page rows:\n%s\nwant:\n%s", strings.Join(rows, "\n"), strings.Join(want, "\n"))
	}
	if strings.Join(v.Upstream, ",") != "1.1.1.1,9.9.9.9" || len(v.Nodes) != 5 {
		t.Errorf("upstream %v, nodes %v", v.Upstream, v.Nodes)
	}

	// Refused before the database: no transaction runs.
	h.db.n.Store(0)
	for _, c := range []struct {
		req  RecordRequest
		want string
	}{
		{RecordRequest{Name: "x", Node: "nas", IP: "10.77.1.9"}, "not both or neither"},
		{RecordRequest{Name: "x"}, "not both or neither"},
		{RecordRequest{Name: "a;b", IP: "10.77.1.9"}, "name"},
		{RecordRequest{Name: "x", IP: "10.77.1.300"}, "ip"},
		{RecordRequest{Name: "x", Node: "NAS box"}, "node"},
	} {
		_, err := s.AddRecord(ctx, by, c.req)
		refused(t, err, http.StatusBadRequest, c.want)
	}
	if n := h.db.n.Load(); n != 0 {
		t.Fatalf("a refused request ran %d transactions", n)
	}
	// Refused by the store or the plan: nothing is written.
	before := h.latestVersion(t)
	_, err = s.AddRecord(ctx, by, RecordRequest{Name: "files", IP: "10.77.1.9"})
	refused(t, err, http.StatusConflict, "exists already")
	_, err = s.AddRecord(ctx, by, RecordRequest{Name: "nas", IP: "10.77.1.9"})
	refused(t, err, http.StatusConflict, "has the name of a node")
	_, err = s.AddRecord(ctx, by, RecordRequest{Name: "router", IP: "192.168.1.1"})
	refused(t, err, http.StatusConflict, "outside the mesh")
	// Names are given without the zone.
	_, err = s.AddRecord(ctx, by, RecordRequest{Name: "photos.mesh", IP: "10.77.1.60"})
	refused(t, err, http.StatusConflict, "ends with the zone")
	_, err = s.AddRecord(ctx, by, RecordRequest{Name: "x", Node: "ghost"})
	refused(t, err, http.StatusNotFound, "ghost")
	if h.latestVersion(t) != before || h.audits(t, "dns.add") != 3 {
		t.Fatalf("a refused change was written: version %d -> %d, %d audit rows", before, h.latestVersion(t), h.audits(t, "dns.add"))
	}

	if ch, err = s.RemoveRecord(ctx, by, "files"); err != nil || ch.Version == 0 {
		t.Fatalf("remove = %+v, %v", ch, err)
	}
	_, err = s.RemoveRecord(ctx, by, "nas")
	refused(t, err, http.StatusNotFound, "node names cannot be removed")
	_, err = s.RemoveRecord(ctx, by, "a b")
	refused(t, err, http.StatusBadRequest, "name")
	if h.audits(t, "dns.rm") != 1 {
		t.Fatal("the removal was not audited")
	}
}
