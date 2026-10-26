package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/hafio/vpn-mesh-utility/internal/proto"
	"github.com/hafio/vpn-mesh-utility/internal/server/store"
	"github.com/hafio/vpn-mesh-utility/internal/testfix"
)

// A container's address is on the node's /24 or, with the containers'
// IPv6 (M8), its /112; neither the network's own addresses nor another
// family's lookalike count.
func TestContainerAddr(t *testing.T) {
	n := store.Node{Name: "nas", ContainerSubnet: testfix.ContainerSubnet, ContainerSubnet6: testfix.ContainerSubnet6}
	for ip, want := range map[string]bool{
		"10.77.128.0": false, "10.77.128.1": false, "10.77.128.2": false, "10.77.128.255": false,
		"10.77.128.3": true, "10.77.128.129": true, "10.77.128.254": true, "10.77.129.130": false, "10.77.0.5": false,
		"fd10:77::8000:0": false, "fd10:77::8000:1": false, "fd10:77::8000:2": false, "fd10:77::8000:3": true,
		"fd10:77::8000:ffff": true, "fd10:77::8001:3": false, "::ffff:10.77.128.129": false,
	} {
		if got := containerAddr(n, netip.MustParseAddr(ip)); got != want {
			t.Errorf("containerAddr(%s) = %v, want %v", ip, got, want)
		}
	}
	if containerAddr(store.Node{}, netip.MustParseAddr("10.77.128.129")) || containerAddr(n, netip.Addr{}) {
		t.Error("a node without a container network has no container addresses")
	}
	v4 := store.Node{Name: "nas", ContainerSubnet: testfix.ContainerSubnet}
	if containerAddr(v4, netip.MustParseAddr("fd10:77::8000:3")) {
		t.Error("a node whose containers have no IPv6 has no IPv6 container addresses")
	}
	for ip, want := range map[string]string{
		"10.77.129.5":     "10.77.129.5 is not a container address of the node's network 10.77.128.0/24",
		"fd10:77::8000:3": "fd10:77::8000:3 is IPv6, and the node's containers have no IPv6 on the mesh (`meshctl node set nas --containers6`)",
	} {
		if got := notContainerAddr(v4, ip); got != want {
			t.Errorf("notContainerAddr(%s) = %q, want %q", ip, got, want)
		}
	}
	if got := notContainerAddr(n, "fd10:77::8001:3"); !strings.Contains(got, "network fd10:77::8000:0/112") {
		t.Errorf("notContainerAddr(IPv6 off the /112) = %q", got)
	}
}

// What is published, and why the rest is not: a node, a custom record, the
// zone and the server's own address keep their names; a name in the zone
// is one label; a node whose network moved, or that is disabled, has none.
func TestPublishNames(t *testing.T) {
	nas := store.Node{ID: testfix.NAS.ID, Name: testfix.NAS.Name, ContainerSubnet: testfix.ContainerSubnet, ContainerSubnet6: testfix.ContainerSubnet6}
	laptop := store.Node{ID: testfix.Laptop.ID, Name: testfix.Laptop.Name, ContainerSubnet: "10.77.129.0/24"}
	old := store.Node{ID: 9, Name: "old", ContainerSubnet: "10.77.130.0/24", Disabled: true}
	vps1 := store.Node{ID: testfix.VPS1.ID, Name: testfix.VPS1.Name}
	nodes := []store.Node{vps1, nas, laptop, old}
	records := []store.DNSRecord{{Name: "printer", IP: "192.168.1.9"}}
	at := func(n store.Node, name, kind, ip string) store.ContainerName {
		return store.ContainerName{NodeID: n.ID, Name: name, Kind: kind, IP: ip}
	}
	set := PublishNames("mesh", []string{"example.lan"}, "api.mesh", nodes, records, []store.ContainerName{
		at(nas, "web", proto.NameService, "10.77.128.130"),
		at(nas, "web", proto.NameService, "fd10:77::8000:8002"), // M8: published beside its IPv4
		at(laptop, "cache", proto.NameService, "fd10:77::8001:8002"),
		at(nas, "web", proto.NameService, "10.77.128.129"),
		at(nas, "web", proto.NameHostname, "10.77.128.129"), // the same name twice: published once
		at(nas, "web.mesh", proto.NameHostname, "10.77.128.131"),
		at(laptop, "web", proto.NameService, "10.77.129.130"),
		at(nas, "vps-1", proto.NameAlias, "10.77.128.132"),
		at(nas, "printer", proto.NameAlias, "10.77.128.132"),
		at(nas, "mesh", proto.NameAlias, "10.77.128.132"),
		at(nas, "api", proto.NameAlias, "10.77.128.132"),
		at(nas, "web.example.org", proto.NameHostname, "10.77.128.132"),
		at(nas, "web.example.lan", proto.NameHostname, "10.77.128.130"), // an extra domain: as it is
		at(nas, "example.lan", proto.NameAlias, "10.77.128.132"),
		at(nas, "a.b.mesh", proto.NameHostname, "10.77.128.132"),
		at(laptop, "db", proto.NameService, "10.77.128.133"),
		at(old, "db", proto.NameService, "10.77.130.129"),
		{NodeID: 99, Name: "ghost", Kind: proto.NameService, IP: "10.77.131.129"},
	})
	want := []proto.MeshName{
		{Name: "web", Node: "laptop", IP: "10.77.129.130"},
		{Name: "web", Node: "nas", IP: "10.77.128.129"},
		{Name: "web", Node: "nas", IP: "10.77.128.130"},
		{Name: "web", Node: "nas", IP: "10.77.128.131"},
		{Name: "web", Node: "nas", IP: "fd10:77::8000:8002"},
		{Name: "web.example.lan", Node: "nas", IP: "10.77.128.130"},
	}
	if !slices.Equal(set.Published, want) {
		t.Fatalf("Published =\n%v\nwant\n%v", set.Published, want)
	}
	dropped := map[string]string{}
	for _, e := range set.Entries {
		if e.Dropped != "" {
			dropped[e.Node+"/"+e.Name] = e.Dropped
		}
	}
	for key, reason := range map[string]string{
		"nas/vps-1":           "node vps-1 has this name",
		"nas/printer":         "a DNS record has this name",
		"nas/mesh":            "is the zone itself",
		"nas/api":             "is the server's address",
		"nas/web.example.org": "is neither a single name in the zone mesh (<name> or <name>.mesh) nor in an extra domain (Settings)",
		"nas/a.b.mesh":        "is neither a single name in the zone mesh (<name> or <name>.mesh) nor in an extra domain (Settings)",
		"nas/example.lan":     "is an extra domain itself",
		"laptop/db":           "the node's container network no longer holds 10.77.128.133",
		"laptop/cache":        "the node's container network no longer holds fd10:77::8001:8002",
		"old/db":              "the node is disabled",
	} {
		if dropped[key] != reason {
			t.Errorf("%s dropped = %q, want %q", key, dropped[key], reason)
		}
		delete(dropped, key)
	}
	if len(dropped) != 0 {
		t.Errorf("dropped for no reason the test knows: %v", dropped)
	}
	if len(set.Entries) != 17 || set.Entries[0].Name != "a.b.mesh" {
		t.Errorf("Entries (every reported name of a known node, by name) = %v", set.Entries)
	}
	resolves := map[string]string{}
	for _, e := range set.Entries {
		resolves[e.Node+"/"+e.Name+"/"+e.IP] = strings.Join(e.Resolves, " ")
	}
	for key, want := range map[string]string{
		"nas/web.mesh/10.77.128.131":        "web.mesh web.nas.mesh",
		"laptop/web/10.77.129.130":          "web.mesh web.laptop.mesh",
		"nas/web.example.lan/10.77.128.130": "web.example.lan",
		"nas/vps-1/10.77.128.132":           "",
	} {
		if resolves[key] != want {
			t.Errorf("%s resolves as %q, want %q", key, resolves[key], want)
		}
	}
}

// An extra domain may not hold the server's own address.
func TestCheckServerDomains(t *testing.T) {
	if err := CheckServerDomains("https://mesh.hafio.services", []string{"example.lan", "home.arpa"}); err != nil {
		t.Fatal(err)
	}
	for _, d := range []string{"hafio.services", "mesh.hafio.services"} {
		if err := CheckServerDomains("https://Mesh.Hafio.Services:8443", []string{"example.lan", d}); err == nil ||
			!strings.Contains(err.Error(), "dns_extra_domains") || !strings.Contains(err.Error(), "mesh.hafio.services") {
			t.Errorf("%s = %v", d, err)
		}
	}
	if err := CheckServerDomains("https://192.0.2.1:8443", []string{"lan"}); err != nil {
		t.Errorf("a server known by its address = %v", err)
	}
	if err := CheckServerDomains("https://[::1", []string{"lan"}); err != nil {
		t.Errorf("not a URL (nothing to protect) = %v", err)
	}
}

func (h *harness) putNames(t *testing.T, f testfix.Node, at time.Time, names ...proto.ContainerName) (*httptest.ResponseRecorder, proto.ErrorBody) {
	t.Helper()
	body, err := json.Marshal(proto.NamesRequest{Names: names})
	if err != nil {
		t.Fatal(err)
	}
	return serve(h.pub, signed(http.MethodPut, proto.PathNames, body, f, at))
}

func (h *harness) storedNames(t *testing.T) []store.ContainerName {
	t.Helper()
	var out []store.ContainerName
	if err := h.st.MeshTx(context.Background(), store.FirstMesh, func(tx *store.MeshTx) error {
		var err error
		out, err = tx.ContainerNames()
		return err
	}); err != nil {
		t.Fatal(err)
	}
	return out
}

func containerNAS() store.Node {
	n := storeNode(testfix.NAS, true)
	n.Mode, n.ContainerSubnet = proto.ModeContainer, testfix.ContainerSubnet
	return n
}

// A report replaces the node's names, drops an address that is not one of
// its containers', says what is not published, nudges every stream when
// something changed, and shows in the mesh view and on the DNS page.
func TestNamesEndpoint(t *testing.T) {
	h := newHarness(t)
	h.seed(t, storeNode(testfix.VPS1, true), storeNode(testfix.VPS2, true), storeNode(testfix.VPS3, true), containerNAS())
	_, mesh, _, _, cancel := h.hub.Subscribe(testfix.VPS1.ID, store.FirstMesh)
	defer cancel()
	web := proto.ContainerName{Name: "web", Kind: proto.NameService, IP: "10.77.128.129"}
	names := []proto.ContainerName{web, {Name: "vps-2", Kind: proto.NameAlias, IP: "10.77.128.129"},
		{Name: "db", Kind: proto.NameService, IP: "10.77.129.5"}}
	w, eb := h.putNames(t, testfix.NAS, testfix.Epoch, names...)
	if w.Code != 200 {
		t.Fatalf("names: %d %s", w.Code, eb.Message)
	}
	var resp proto.NamesResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Stored != 2 || len(resp.Dropped) != 2 || resp.Dropped[0].Name != "db" || resp.Dropped[1].Reason != "node vps-2 has this name" {
		t.Fatalf("response %+v", resp)
	}
	if !closed(mesh) {
		t.Fatal("a change to the names must nudge every stream")
	}
	if got := h.storedNames(t); len(got) != 2 {
		t.Fatalf("stored %v", got)
	}
	// The same report again changes nothing and nudges nobody.
	if w, eb := h.putNames(t, testfix.NAS, testfix.Epoch.Add(time.Second), names...); w.Code != 200 {
		t.Fatalf("names again: %d %s", w.Code, eb.Message)
	}
	if closed(mesh) {
		t.Fatal("an unchanged report must not nudge")
	}

	w, eb = serve(h.pub, signed(http.MethodGet, proto.PathMesh, nil, testfix.VPS1, testfix.Epoch.Add(2*time.Second)))
	if w.Code != 200 {
		t.Fatalf("mesh: %d %s", w.Code, eb.Message)
	}
	var v proto.MeshView
	if err := json.Unmarshal(w.Body.Bytes(), &v); err != nil || v.Validate() != nil {
		t.Fatalf("mesh view %s: %v", w.Body, err)
	}
	if want := []proto.MeshName{{Name: "web", Node: "nas", IP: "10.77.128.129"}}; !slices.Equal(v.Names, want) {
		t.Fatalf("mesh view names = %v, want %v", v.Names, want)
	}
	dv, err := NewNodes(Deps{DB: h.db, ServerURL: testURL}).DNS(context.Background())
	if err != nil || len(dv.Names) != 2 || dv.Names[0].Name != "vps-2" || dv.Names[0].Dropped == "" || dv.Names[1].Dropped != "" {
		t.Fatalf("DNS view names = %+v, %v", dv.Names, err)
	}

	// An empty report clears them.
	if w, eb := h.putNames(t, testfix.NAS, testfix.Epoch.Add(3*time.Second)); w.Code != 200 || len(h.storedNames(t)) != 0 {
		t.Fatalf("clearing: %d %s", w.Code, eb.Message)
	}
}

// With the containers' IPv6 (M8) a report's IPv6 addresses on the node's
// /112 are kept and published beside the IPv4 ones; one off it is
// dropped, saying why.
func TestNamesEndpointWithIPv6(t *testing.T) {
	h := newHarness(t)
	if err := h.st.MeshTx(context.Background(), store.FirstMesh, func(tx *store.MeshTx) error { return tx.SetSetting(store.SettingMeshCIDR6, testfix.MeshCIDR6) }); err != nil {
		t.Fatal(err)
	}
	nas := containerNAS()
	nas.MeshIP6, nas.ContainerSubnet6 = testfix.NAS.MeshIP6, testfix.ContainerSubnet6
	h.seed(t, storeNode(testfix.VPS1, true), storeNode(testfix.VPS2, true), storeNode(testfix.VPS3, true), nas)
	w, eb := h.putNames(t, testfix.NAS, testfix.Epoch, proto.ContainerName{Name: "web", Kind: proto.NameService, IP: "10.77.128.129"},
		proto.ContainerName{Name: "web", Kind: proto.NameService, IP: "fd10:77::8000:8001"},
		proto.ContainerName{Name: "db", Kind: proto.NameService, IP: "fd10:77::8001:5"})
	if w.Code != 200 {
		t.Fatalf("names: %d %s", w.Code, eb.Message)
	}
	var resp proto.NamesResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil || resp.Stored != 2 || len(resp.Dropped) != 1 ||
		!strings.Contains(resp.Dropped[0].Reason, "network "+testfix.ContainerSubnet6) {
		t.Fatalf("response %+v, %v", resp, err)
	}
	w, eb = serve(h.pub, signed(http.MethodGet, proto.PathMesh, nil, testfix.VPS1, testfix.Epoch.Add(time.Second)))
	var v proto.MeshView
	if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &v) != nil || v.Validate() != nil {
		t.Fatalf("mesh: %d %s %s", w.Code, eb.Message, w.Body)
	}
	if want := []proto.MeshName{{Name: "web", Node: "nas", IP: "10.77.128.129"}, {Name: "web", Node: "nas", IP: "fd10:77::8000:8001"}}; !slices.Equal(v.Names, want) {
		t.Fatalf("mesh view names = %v, want %v", v.Names, want)
	}
}

// Refused reports store nothing: a node without a container network, a
// malformed name, an unsigned request (which never reaches the database).
func TestNamesEndpointRefuses(t *testing.T) {
	h := newHarness(t)
	h.seed(t, storeNode(testfix.VPS1, true), storeNode(testfix.VPS2, true), storeNode(testfix.VPS3, true), containerNAS(),
		storeNode(testfix.Laptop, true))
	web := proto.ContainerName{Name: "web", Kind: proto.NameService, IP: "10.77.128.129"}
	if w, eb := h.putNames(t, testfix.Laptop, testfix.Epoch, web); w.Code != 409 || eb.Code != proto.CodeConflict ||
		!bytes.Contains(w.Body.Bytes(), []byte("--containers")) {
		t.Fatalf("a node without a container network: %d %+v", w.Code, eb)
	}
	bad := web
	bad.Name = "web;id"
	if w, eb := h.putNames(t, testfix.NAS, testfix.Epoch, bad); w.Code != 400 || eb.Code != proto.CodeBadRequest {
		t.Fatalf("a malformed name: %d %+v", w.Code, eb)
	}
	body, _ := json.Marshal(proto.NamesRequest{Names: []proto.ContainerName{web}})
	h.db.n.Store(0)
	if w, eb := serve(h.pub, httptest.NewRequest(http.MethodPut, proto.PathNames, bytes.NewReader(body))); w.Code != 401 ||
		eb.Code != proto.CodeUnauthorized || h.db.n.Load() != 0 {
		t.Fatalf("unsigned: %d %q, db %d", w.Code, eb.Code, h.db.n.Load())
	}
	if got := h.storedNames(t); len(got) != 0 {
		t.Fatalf("a refused report stored %v", got)
	}
}
