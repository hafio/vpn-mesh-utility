package api

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"net/http"
	"net/netip"
	"strings"
	"testing"
	"time"

	"github.com/hafio/vpn-mesh-utility/internal/proto"
	"github.com/hafio/vpn-mesh-utility/internal/server/planner"
	"github.com/hafio/vpn-mesh-utility/internal/testfix"
)

func TestULAPrefix(t *testing.T) {
	fill := func(b []byte) (int, error) {
		for i := range b {
			b[i] = 0xab
		}
		return len(b), nil
	}
	if p, err := ULAPrefix(fill); err != nil || p.String() != "fdab:abab:abab::/64" {
		t.Errorf("ULAPrefix(0xab...) = %v, %v", p, err)
	}
	broken := func([]byte) (int, error) { return 0, errors.New("no entropy") }
	if _, err := ULAPrefix(broken); err == nil || !strings.Contains(err.Error(), "random IPv6 prefix") || !strings.Contains(err.Error(), "no entropy") {
		t.Errorf("ULAPrefix(broken) = %v", err)
	}
	p, err := ULAPrefix(rand.Read)
	if err != nil || p.Addr().As16()[0] != 0xfd || planner.CheckMeshCIDR6(p) != nil {
		t.Errorf("ULAPrefix(rand.Read) = %v, %v", p, err)
	}
}

func setIPv6(t *testing.T, h *harness, prefix string) (IPv6View, int, proto.ErrorBody) {
	t.Helper()
	var v IPv6View
	code, eb := adminCall(t, h, http.MethodPut, AdminIPv6, IPv6Request{Prefix: prefix}, &v)
	return v, code, eb
}

func TestSetIPv6(t *testing.T) {
	h := newHarness(t)
	h.seed(t, storeNode(testfix.VPS1, true), storeNode(testfix.VPS2, true), storeNode(testfix.VPS3, true),
		storeNode(testfix.NAS, true), storeNode(testfix.Laptop, true), storeNode(testfix.Phone, true))
	if code, eb := adminCall(t, h, http.MethodPatch, AdminNodes+"/laptop", NodeSetRequest{Family: ptr(proto.FamilyIPv4)}, nil); code != 200 {
		t.Fatalf("laptop to ipv4 before any prefix: %d %s", code, eb.Message)
	}
	var v IPv6View
	if code, _ := adminCall(t, h, http.MethodGet, AdminIPv6, nil, &v); code != 200 || v.Prefix != "" || v.Addressed != 0 {
		t.Fatalf("GET without a prefix: %d %+v", code, v)
	}

	// Bad prefixes are refused before the store is touched.
	before := h.db.n.Load()
	for _, bad := range []string{"fd10:77::1/64", "fd10::/40", "fd10:77::/96", "10.77.0.0/16", "fd10:77::/64;x", "fd10:77::/64\nx"} {
		if _, code, eb := setIPv6(t, h, bad); code != 400 || eb.Message == "" {
			t.Errorf("prefix %q: %d %q, want 400", bad, code, eb.Message)
		}
	}
	if h.db.n.Load() != before || h.audits(t, "ipv6.set") != 0 {
		t.Fatal("a refused prefix reached the store")
	}

	v, code, eb := setIPv6(t, h, testfix.MeshCIDR6)
	if code != 200 || v.Prefix != testfix.MeshCIDR6 || v.Addressed != 5 || v.Version == 0 {
		t.Fatalf("set %s: %d %+v %s", testfix.MeshCIDR6, code, v, eb.Message)
	}
	for _, f := range []testfix.Node{testfix.VPS1, testfix.VPS2, testfix.VPS3, testfix.NAS, testfix.Phone} {
		if got := h.node(t, f.Name).MeshIP6; got != f.MeshIP6 {
			t.Errorf("%s: IPv6 mesh address %q, want %q (its role's block, as its IPv4)", f.Name, got, f.MeshIP6)
		}
	}
	if got := h.node(t, "laptop").MeshIP6; got != "" {
		t.Errorf("the ipv4 laptop got %q", got)
	}
	if cfg := h.config(t, "vps-1"); cfg.Interface.Address6 != testfix.VPS1.MeshIP6+"/128" {
		t.Errorf("vps-1's interface: %+v", cfg.Interface)
	}

	if v, code, _ := setIPv6(t, h, IPv6Auto); code != 200 || v.Prefix != testfix.MeshCIDR6 || v.Version != 0 {
		t.Errorf("auto with a prefix keeps it: %d %+v", code, v)
	}
	if _, code, eb := setIPv6(t, h, "fd20::/64"); code != 409 || !strings.Contains(eb.Message, "give IPv6 up first") {
		t.Errorf("another prefix over addresses: %d %q", code, eb.Message)
	}
	if h.node(t, "vps-1").MeshIP6 != testfix.VPS1.MeshIP6 {
		t.Fatal("a refused swap changed an address")
	}

	// IPv6-only nodes have nothing left without IPv6.
	var ch NodeChange
	if code, eb := adminCall(t, h, http.MethodPost, AdminNodes, NodeAddRequest{Name: "pad", Role: "spoke", Family: proto.FamilyIPv6}, &ch); code != 200 ||
		ch.Node.MeshIP != "" || ch.Node.MeshIP6 != "fd10:77::2:1" {
		t.Fatalf("an ipv6 spoke: %d %+v %s", code, ch.Node, eb.Message)
	}
	if _, code, eb := setIPv6(t, h, IPv6Off); code != 409 || !strings.Contains(eb.Message, "pad is IPv6 only") {
		t.Errorf("off with an ipv6-only node: %d %q", code, eb.Message)
	}
	if code, eb := adminCall(t, h, http.MethodPatch, AdminNodes+"/pad", NodeSetRequest{Family: ptr(proto.FamilyDual)}, &ch); code != 200 || ch.Node.MeshIP != "10.77.2.2" {
		t.Fatalf("pad to dual: %d %+v %s", code, ch.Node, eb.Message)
	}
	v, code, _ = setIPv6(t, h, "")
	if code != 200 || v.Prefix != "" || v.Addressed != 0 || v.Version == 0 {
		t.Fatalf("off: %d %+v", code, v)
	}
	for _, name := range []string{"vps-1", "nas", "pad"} {
		if n := h.node(t, name); n.MeshIP6 != "" || n.MeshIP == "" {
			t.Errorf("%s after off: %q / %q", name, n.MeshIP, n.MeshIP6)
		}
	}

	// auto on a mesh without one draws a unique local /64.
	v, code, _ = setIPv6(t, h, IPv6Auto)
	if code != 200 || !strings.HasPrefix(v.Prefix, "fd") || !strings.HasSuffix(v.Prefix, "::/64") || v.Addressed != 6 {
		t.Errorf("auto: %d %+v", code, v)
	}
	if n := h.audits(t, "ipv6.set"); n != 4 {
		t.Errorf("%d ipv6.set audits, want 4 (set, auto, off, auto)", n)
	}
}

func TestDualStackNodes(t *testing.T) {
	h := newHarness(t)
	h.seedServers(t)
	add := func(req NodeAddRequest) (NodeInfo, int, string) {
		t.Helper()
		var ch NodeChange
		code, eb := adminCall(t, h, http.MethodPost, AdminNodes, req, &ch)
		return ch.Node, code, eb.Message
	}
	set := func(name string, req NodeSetRequest) (NodeChange, int, string) {
		t.Helper()
		var ch NodeChange
		code, eb := adminCall(t, h, http.MethodPatch, AdminNodes+"/"+name, req, &ch)
		return ch, code, eb.Message
	}

	// Without a prefix, nothing gets IPv6.
	for name, req := range map[string]NodeAddRequest{
		"fixed ip6": {Name: "x", Role: "spoke", IP6: "fd10:77::2:9"},
		"ipv6 only": {Name: "x", Role: "spoke", Family: proto.FamilyIPv6},
	} {
		if _, code, msg := add(req); code != 409 || !strings.Contains(msg, "no IPv6 prefix yet") {
			t.Errorf("%s without a prefix: %d %q", name, code, msg)
		}
	}
	if n, code, _ := add(NodeAddRequest{Name: "early", Role: "spoke"}); code != 200 || n.MeshIP6 != "" || n.Family != proto.FamilyDual {
		t.Errorf("a dual spoke without a prefix: %d %+v", code, n)
	}
	if _, code, _ := setIPv6(t, h, testfix.MeshCIDR6); code != 200 {
		t.Fatal("set the prefix")
	}

	// Refused before the store.
	before := h.db.n.Load()
	for name, req := range map[string]NodeAddRequest{
		"unknown family":     {Name: "x", Role: "spoke", Family: "ipv5"},
		"ipv4 with ip6":      {Name: "x", Role: "spoke", Family: proto.FamilyIPv4, IP6: "fd10:77::2:9"},
		"ipv6 with ip":       {Name: "x", Role: "spoke", Family: proto.FamilyIPv6, IP: "10.77.2.9"},
		"ip6 not IPv6":       {Name: "x", Role: "spoke", IP6: "10.77.2.9"},
		"ip6 with a zone":    {Name: "x", Role: "spoke", IP6: "fe80::1%eth0"},
		"spoke endpoint6":    {Name: "x", Role: "spoke", Endpoint6: "[2001:db8::20]:51820"},
		"endpoint6 is IPv4":  {Name: "x", Role: "server", Endpoint6: "203.0.113.20:51820"},
		"endpoint6 injected": {Name: "x", Role: "server", Endpoint6: "[2001:db8::20]:51820\nrole: phone"},
		"containers6 alone":  {Name: "x", Role: "spoke", Containers6: true},
		"containers6 ipv4":   {Name: "x", Role: "spoke", Family: proto.FamilyIPv4, Containers: true, Containers6: true},
		"containers6 ipv6":   {Name: "x", Role: "spoke", Family: proto.FamilyIPv6, Containers: true, Containers6: true},
	} {
		if _, code, msg := add(req); code != 400 || msg == "" {
			t.Errorf("%s: %d %q, want 400", name, code, msg)
		}
	}
	if h.db.n.Load() != before {
		t.Fatal("a refused add reached the store")
	}
	for name, req := range map[string]NodeAddRequest{
		"ip6 taken":         {Name: "x", Role: "spoke", IP6: testfix.VPS1.MeshIP6},
		"ip6 outside":       {Name: "x", Role: "spoke", IP6: "fd20::5"},
		"ip6 in containers": {Name: "x", Role: "spoke", IP6: "fd10:77::8000:5"},
		"ip6 the prefix":    {Name: "x", Role: "spoke", IP6: "fd10:77::"},
	} {
		if _, code, msg := add(req); code != 400 || !strings.Contains(msg, "IPv6 mesh address") {
			t.Errorf("%s: %d %q", name, code, msg)
		}
	}
	if _, code, msg := add(NodeAddRequest{Name: "v6vps", Role: "server", Family: proto.FamilyIPv6}); code != 409 || !strings.Contains(msg, "needs an IPv4 mesh address") {
		t.Errorf("an IPv6-only server: %d %q", code, msg)
	}

	if n, code, msg := add(NodeAddRequest{Name: "nas", Role: "spoke", AlwaysOn: true}); code != 200 || n.MeshIP != "10.77.1.1" || n.MeshIP6 != "fd10:77::1:1" {
		t.Errorf("dual nas: %d %+v %s", code, n, msg)
	}
	if n, code, _ := add(NodeAddRequest{Name: "lap4", Role: "spoke", Family: proto.FamilyIPv4}); code != 200 || n.MeshIP != "10.77.2.2" || n.MeshIP6 != "" {
		t.Errorf("ipv4 laptop: %d %+v", code, n)
	}
	// early took fd10:77::2:1 when the prefix came.
	if n, code, _ := add(NodeAddRequest{Name: "lap6", Role: "spoke", Family: proto.FamilyIPv6}); code != 200 || n.MeshIP != "" || n.MeshIP6 != "fd10:77::2:2" ||
		n.Family != proto.FamilyIPv6 {
		t.Errorf("ipv6 laptop: %d %+v", code, n)
	}
	if n, code, _ := add(NodeAddRequest{Name: "fixed", Role: "spoke", IP6: "FD10:77::2:0099"}); code != 200 || n.MeshIP6 != "fd10:77::2:99" {
		t.Errorf("fixed ip6: %d %+v", code, n)
	}
	if n, code, _ := add(NodeAddRequest{Name: "vps-4", Role: "server", Endpoint6: "[2001:DB8::20]:51820"}); code != 200 || n.Endpoint6 != "[2001:db8::20]:51820" {
		t.Errorf("server with endpoint6: %d %+v", code, n)
	}
	if n, code, _ := add(NodeAddRequest{Name: "box", Role: "spoke", Containers: true, Containers6: true}); code != 200 ||
		n.ContainerSubnet6 != testfix.ContainerSubnet6 || n.ContainerSubnet == "" {
		t.Errorf("containers6: %d %+v", code, n)
	}

	// Family changes allocate what a family adds and drop what it takes.
	// box took fd10:77::2:3; fixed sits at ::2:99.
	if ch, code, msg := set("lap4", NodeSetRequest{Family: ptr(proto.FamilyDual)}); code != 200 || ch.Node.MeshIP6 != "fd10:77::2:4" || ch.Node.MeshIP != "10.77.2.2" {
		t.Errorf("lap4 to dual: %d %+v %s", code, ch.Node, msg)
	}
	if ch, code, _ := set("lap6", NodeSetRequest{Family: ptr(proto.FamilyDual)}); code != 200 || ch.Node.MeshIP != "10.77.2.5" || ch.Node.MeshIP6 != "fd10:77::2:2" {
		t.Errorf("lap6 to dual: %d %+v", code, ch.Node)
	}
	if ch, code, _ := set("box", NodeSetRequest{Family: ptr(proto.FamilyIPv4)}); code != 200 || ch.Node.MeshIP6 != "" || ch.Node.ContainerSubnet6 != "" ||
		ch.Node.ContainerSubnet == "" {
		t.Errorf("box to ipv4: %d %+v", code, ch.Node)
	}
	if _, code, msg := set("box", NodeSetRequest{Family: ptr(proto.FamilyIPv6)}); code != 409 || !strings.Contains(msg, "container network") {
		t.Errorf("box to ipv6 with containers: %d %q", code, msg)
	}
	if ch, code, _ := set("nas", NodeSetRequest{Family: ptr(proto.FamilyIPv6)}); code != 200 || ch.Node.MeshIP != "" || ch.Node.MeshIP6 != "fd10:77::1:1" {
		t.Errorf("nas to ipv6: %d %+v", code, ch.Node)
	}
	if _, code, msg := set("nas", NodeSetRequest{Family: ptr("both")}); code != 400 || !strings.Contains(msg, "ip_family") {
		t.Errorf("unknown family: %d %q", code, msg)
	}

	// Moving an IPv6 address.
	ch, code, msg := set("lap4", NodeSetRequest{MeshIP6: ptr("fd10:77::2:50")})
	if code != 200 || ch.Node.MeshIP6 != "fd10:77::2:50" || !strings.Contains(strings.Join(ch.Notes, "\n"), "lap4 moves from fd10:77::2:4 to fd10:77::2:50") {
		t.Errorf("move lap4: %d %+v %s", code, ch, msg)
	}
	if ch, code, _ := set("lap4", NodeSetRequest{MeshIP6: ptr("fd10:77::2:50")}); code != 200 || ch.Version != 0 || len(ch.Notes) != 0 {
		t.Errorf("the same address: %d %+v", code, ch)
	}
	if _, code, msg := set("lap4", NodeSetRequest{MeshIP6: ptr(testfix.VPS1.MeshIP6)}); code != 400 || !strings.Contains(msg, "already assigned") {
		t.Errorf("onto vps-1's: %d %q", code, msg)
	}
	if _, code, msg := set("box", NodeSetRequest{MeshIP6: ptr("fd10:77::2:60")}); code != 400 || !strings.Contains(msg, "IPv4 only") {
		t.Errorf("an ipv4 node's IPv6: %d %q", code, msg)
	}
	if _, code, msg := set("lap4", NodeSetRequest{MeshIP6: ptr("fd10:77::2:60;rm")}); code != 400 || !strings.Contains(msg, "mesh_ip6") {
		t.Errorf("a bad IPv6: %d %q", code, msg)
	}

	// A server's IPv6 endpoint, and the containers' IPv6.
	if ch, code, _ := set("vps-1", NodeSetRequest{Endpoint6: ptr("[2001:db8::10]:51820")}); code != 200 || ch.Node.Endpoint6 != "[2001:db8::10]:51820" {
		t.Errorf("vps-1 endpoint6: %d %+v", code, ch.Node)
	}
	if ch, code, _ := set("vps-1", NodeSetRequest{Endpoint6: ptr("")}); code != 200 || ch.Node.Endpoint6 != "" {
		t.Errorf("clear vps-1 endpoint6: %d %+v", code, ch.Node)
	}
	if _, code, msg := set("vps-1", NodeSetRequest{Endpoint6: ptr("203.0.113.10:51820")}); code != 400 || !strings.Contains(msg, "[IPv6]:port") {
		t.Errorf("an IPv4 endpoint6: %d %q", code, msg)
	}
	if _, code, msg := set("lap4", NodeSetRequest{Containers6: ptr(true)}); code != 400 || !strings.Contains(msg, "no container network") {
		t.Errorf("containers6 without containers: %d %q", code, msg)
	}
	if _, code, msg := set("box", NodeSetRequest{Containers6: ptr(true)}); code != 400 || !strings.Contains(msg, "no IPv6 mesh address") {
		t.Errorf("containers6 on an ipv4 node: %d %q", code, msg)
	}
	if _, code, _ := set("box", NodeSetRequest{Family: ptr(proto.FamilyDual)}); code != 200 {
		t.Fatal("box back to dual")
	}
	if ch, code, _ := set("box", NodeSetRequest{Containers6: ptr(true)}); code != 200 || ch.Node.ContainerSubnet6 != testfix.ContainerSubnet6 {
		t.Errorf("containers6 on: %d %+v", code, ch.Node)
	}
	if ch, code, _ := set("box", NodeSetRequest{Containers: ptr(false)}); code != 200 || ch.Node.ContainerSubnet6 != "" {
		t.Errorf("containers off takes their IPv6 too: %d %+v", code, ch.Node)
	}

	var d NodeDetail
	if code, _ := adminCall(t, h, http.MethodGet, AdminNodes+"/nas", nil, &d); code != 200 || d.Node.Family != proto.FamilyIPv6 || d.Node.MeshIP6 != "fd10:77::1:1" {
		t.Errorf("show nas: %d %+v", code, d.Node)
	}
}

func TestIPv6DNSRecords(t *testing.T) {
	h := meshHarness(t)
	s, ctx, by := h.nodes(), context.Background(), Actor{Name: "user:ops", UserID: 1}
	if v, err := s.DNS(ctx); err != nil || v.IPv6 {
		t.Fatalf("DNS before a prefix: IPv6 %v, %v", v.IPv6, err)
	}
	if _, err := s.SetIPv6(ctx, by, IPv6Request{Prefix: testfix.MeshCIDR6}); err != nil {
		t.Fatal(err)
	}
	if ch, err := s.AddRecord(ctx, by, RecordRequest{Name: "printer6", IP6: "FD10:77::1:200"}); err != nil || ch.Version == 0 {
		t.Fatalf("an IPv6-only record = %+v, %v", ch, err)
	}
	if _, err := s.AddRecord(ctx, by, RecordRequest{Name: "both", IP: "10.77.1.201", IP6: "fd10:77::1:201"}); err != nil {
		t.Fatalf("a dual record: %v", err)
	}
	h.db.n.Store(0)
	for _, c := range []struct {
		req  RecordRequest
		want string
	}{
		{RecordRequest{Name: "x", Node: "nas", IP6: "fd10:77::1:9"}, "not both or neither"},
		{RecordRequest{Name: "x", IP6: "10.77.1.9"}, "ip6"},
		{RecordRequest{Name: "x", IP: "fd10:77::1:9"}, "ip"},
		{RecordRequest{Name: "x", IP6: "fd10:77::1:9\nfd10:77::1:8"}, "ip6"},
	} {
		_, err := s.AddRecord(ctx, by, c.req)
		refused(t, err, http.StatusBadRequest, c.want)
	}
	if n := h.db.n.Load(); n != 0 {
		t.Fatalf("a refused request ran %d transactions", n)
	}
	_, err := s.AddRecord(ctx, by, RecordRequest{Name: "far", IP6: "fd20::5"})
	refused(t, err, http.StatusConflict, "outside")

	v, err := s.DNS(ctx)
	if err != nil || !v.IPv6 {
		t.Fatalf("DNS: IPv6 %v, %v", v.IPv6, err)
	}
	got := map[string]string{}
	for _, e := range v.Records {
		got[e.Name] = e.IP + "|" + e.IP6
	}
	for name, want := range map[string]string{"nas": "10.77.1.1|fd10:77::1:1", "printer6": "|fd10:77::1:200", "both": "10.77.1.201|fd10:77::1:201",
		"laptop": "10.77.2.1|fd10:77::2:1"} {
		if got[name] != want {
			t.Errorf("DNS page %s = %q, want %q", name, got[name], want)
		}
	}
	served := map[string]string{}
	for _, r := range h.config(t, "vps-1").DNS.Records {
		served[r.Name] = r.IP + "|" + r.IP6
	}
	if served["printer6"] != "|fd10:77::1:200" || served["nas"] != "10.77.1.1|fd10:77::1:1" {
		t.Errorf("vps-1 serves %v", served)
	}
}

func TestPickEndpoint6(t *testing.T) {
	cases := []struct {
		name               string
		endpoint6, public6 string
		current, observed  string
		want               string
	}{
		{"explicit wins", "[2001:db8::10]:4000", "2001:db8::7", "[2001:db8::1]:51820", "2001:db8::8", "[2001:db8::10]:4000"},
		{"detected next", "", "2001:db8::7", "[2001:db8::1]:51820", "2001:db8::8", "[2001:db8::7]:51821"},
		{"a unique local address is not public", "", "fd10::7", "[2001:db8::1]:51820", "2001:db8::8", "[2001:db8::1]:51820"},
		{"observed last", "", "", "", "2001:db8::8", "[2001:db8::8]:51821"},
		{"an IPv4 observed address is no IPv6 endpoint", "", "", "", "198.51.100.8", ""},
		{"a link-local observed address is ignored", "", "", "", "fe80::1", ""},
	}
	for _, c := range cases {
		obs, _ := netip.ParseAddr(c.observed) // "" stays the zero Addr
		got := pickEndpoint6(proto.JoinRequest{Endpoint6: c.endpoint6, PublicIP6: c.public6}, c.current, obs, 51821)
		if got != c.want {
			t.Errorf("%s: got %q, want %q", c.name, got, c.want)
		}
	}
}

// A server's join brings its IPv6 endpoint unless it is IPv4 only (M8).
func TestJoinTakesAnIPv6Endpoint(t *testing.T) {
	h := newHarness(t)
	four := storeNode(testfix.VPS3, false)
	four.Family = proto.FamilyIPv4
	h.seed(t, storeNode(testfix.VPS1, true), storeNode(testfix.VPS2, false), four)
	if _, err := h.nodes().SetIPv6(context.Background(), cli(), IPv6Request{Prefix: testfix.MeshCIDR6}); err != nil {
		t.Fatal(err)
	}
	dual := func(r *proto.JoinRequest) { r.PublicIP, r.PublicIP6 = "198.51.100.7", "2001:db8::7" }
	w, eb := postJoin(h, joinBody(t, h.issue(t, "vps-2"), testfix.VPS2, dual))
	var resp proto.JoinResponse
	if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &resp) != nil || resp.MeshIP6 != testfix.VPS2.MeshIP6 {
		t.Fatalf("join vps-2: %d %s %+v", w.Code, eb.Message, resp)
	}
	if n := h.node(t, "vps-2"); n.Endpoint != "198.51.100.7:51820" || n.Endpoint6 != "[2001:db8::7]:51820" {
		t.Errorf("vps-2's endpoints: %q, %q", n.Endpoint, n.Endpoint6)
	}
	only4 := func(r *proto.JoinRequest) { r.Endpoint6 = "[2001:db8::30]:51820" }
	if w, eb := postJoin(h, joinBody(t, h.issue(t, "vps-3"), testfix.VPS3, only4)); w.Code != 200 {
		t.Fatalf("join vps-3: %d %s", w.Code, eb.Message)
	}
	if n := h.node(t, "vps-3"); n.Endpoint6 != "" || n.Endpoint != testfix.VPS3.Endpoint {
		t.Errorf("the IPv4-only vps-3's endpoints: %q, %q", n.Endpoint, n.Endpoint6)
	}
}

// The families an agent reports its host can use reach the plan through the
// reconciler, once per change (M8).
func TestReportedHostFamilies(t *testing.T) {
	h := meshHarness(t)
	if _, err := h.nodes().SetIPv6(context.Background(), cli(), IPv6Request{Prefix: testfix.MeshCIDR6}); err != nil {
		t.Fatal(err)
	}
	calls := 0
	report := func(families ...string) (int, proto.StatusResponse) {
		t.Helper()
		calls++
		body, _ := json.Marshal(proto.StatusReport{AgentVersion: "v1.0.0", WGBackend: "kernel", Families: families})
		at := testfix.Epoch.Add(time.Duration(calls) * time.Second) // a fresh signature per call
		w, _ := serve(h.pub, signed(http.MethodPost, proto.PathStatus, body, testfix.NAS, at))
		var resp proto.StatusResponse
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		return w.Code, resp
	}
	routes := func() string { // what vps-1 routes to the nas
		t.Helper()
		for _, p := range h.config(t, "vps-1").Peers {
			if p.Name == "nas" {
				return strings.Join(p.AllowedIPs, ",")
			}
		}
		return ""
	}
	if got := routes(); !strings.Contains(got, testfix.NAS.MeshIP6) {
		t.Fatalf("vps-1 routes %s to the dual nas", got)
	}
	before := h.latestVersion(t)
	code, resp := report(proto.FamilyIPv4)
	if code != 200 || resp.ConfigVersion <= before || h.node(t, "nas").HostFamilies != "ipv4" || h.audits(t, "node.families") != 1 {
		t.Fatalf("an IPv4-only host: %d %+v, %q", code, resp, h.node(t, "nas").HostFamilies)
	}
	if got := routes(); strings.Contains(got, ":") {
		t.Errorf("vps-1 still routes IPv6 to the nas: %s", got)
	}
	var d NodeDetail
	if code, _ := adminCall(t, h, http.MethodGet, AdminNodes+"/nas", nil, &d); code != 200 || strings.Join(d.Node.HostFamilies, ",") != "ipv4" ||
		!strings.Contains(strings.Join(d.Warnings, "\n"), "cannot use IPv6") {
		t.Errorf("show nas: %d %+v %v", code, d.Node.HostFamilies, d.Warnings)
	}
	v := h.latestVersion(t)
	if code, _ := report(proto.FamilyIPv4); code != 200 || h.latestVersion(t) != v || h.audits(t, "node.families") != 1 {
		t.Errorf("the same report again changed something")
	}
	if code, _ := report(proto.FamilyIPv6, proto.FamilyIPv4); code != 200 || h.node(t, "nas").HostFamilies != "ipv4,ipv6" ||
		!strings.Contains(routes(), testfix.NAS.MeshIP6) {
		t.Errorf("both again: %d %q %s", code, h.node(t, "nas").HostFamilies, routes())
	}
	if code, _ := report(); code != 200 || h.node(t, "nas").HostFamilies != "" || h.audits(t, "node.families") != 3 {
		t.Errorf("not said: %d %q", code, h.node(t, "nas").HostFamilies)
	}
	if code, _ := report("ipx"); code != 400 || h.node(t, "nas").HostFamilies != "" || h.audits(t, "node.families") != 3 {
		t.Errorf("an unknown family: %d", code)
	}
}
