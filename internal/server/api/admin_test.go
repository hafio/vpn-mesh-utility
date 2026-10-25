package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/hafio/vpn-mesh-utility/internal/proto"
	"github.com/hafio/vpn-mesh-utility/internal/server/store"
	"github.com/hafio/vpn-mesh-utility/internal/testfix"
)

func adminCall(t *testing.T, h *harness, method, path string, in any, out any) (int, proto.ErrorBody) {
	t.Helper()
	var body []byte
	if in != nil {
		var err error
		if body, err = json.Marshal(in); err != nil {
			t.Fatal(err)
		}
	}
	w, eb := serve(h.admin, httptest.NewRequest(method, path, bytes.NewReader(body)))
	if w.Code == 200 && out != nil {
		if err := json.Unmarshal(w.Body.Bytes(), out); err != nil {
			t.Fatalf("decode %s: %v", w.Body.String(), err)
		}
	}
	return w.Code, eb
}

func TestAdminAddIssuesAToken(t *testing.T) {
	h := newHarness(t)
	h.seedServers(t)
	var ch NodeChange
	code, eb := adminCall(t, h, http.MethodPost, AdminNodes, NodeAddRequest{Name: "nas", Role: "spoke", AlwaysOn: true, Hub: "vps-2"}, &ch)
	if code != 200 {
		t.Fatalf("%d %s", code, eb.Message)
	}
	if ch.Node.MeshIP != "10.77.1.1" || ch.Node.Hub != "vps-2" || ch.Node.Enrolled || ch.Version != 0 {
		t.Fatalf("node %+v version %d (an unjoined node changes no config)", ch.Node, ch.Version)
	}
	en := ch.Enrollment
	if en == nil || len(en.Token) != 43 || !strings.Contains(en.Install, testURL+"/install.sh") || !strings.Contains(en.Join, "--token "+en.Token) {
		t.Fatalf("enrollment %+v", en)
	}
	var audit []store.AuditEntry
	if err := h.st.MeshTx(context.Background(), store.FirstMesh, func(tx *store.MeshTx) error {
		var err error
		audit, err = tx.AuditLog(1, 0)
		return err
	}); err != nil {
		t.Fatal(err)
	}
	if len(audit) != 1 || audit[0].Action != "node.add" || audit[0].Actor != "cli" || strings.Contains(audit[0].Detail, en.Token) {
		t.Fatalf("audit %+v: must record the add without the token", audit)
	}
	// The next server takes the next free address.
	if code, _ := adminCall(t, h, http.MethodPost, AdminNodes, NodeAddRequest{Name: "vps-4", Role: "server"}, &ch); code != 200 || ch.Node.MeshIP != "10.77.0.4" {
		t.Fatalf("second add: %d %+v", code, ch.Node)
	}
}

func TestAdminAddRefusesBadInput(t *testing.T) {
	h := newHarness(t)
	h.seedServers(t)
	cases := map[string]NodeAddRequest{
		"bad name":            {Name: "VPS_4", Role: "server"},
		"shell in name":       {Name: "a;b", Role: "server"},
		"unknown role":        {Name: "x", Role: "router"},
		"ip outside the mesh": {Name: "x", Role: "server", IP: "192.168.1.1"},
		"ip taken":            {Name: "x", Role: "server", IP: "10.77.0.1"},
		"endpoint with host":  {Name: "x", Role: "server", Endpoint: "vps.example:51820"},
		"spoke endpoint":      {Name: "x", Role: "spoke", Endpoint: "203.0.113.50:51820"},
		"mtu too small":       {Name: "x", Role: "spoke", MTU: 576},
		"unknown hub":         {Name: "x", Role: "spoke", Hub: "vps-9"},
		"duplicate":           {Name: "vps-1", Role: "server"},
	}
	for name, req := range cases {
		code, eb := adminCall(t, h, http.MethodPost, AdminNodes, req, nil)
		if code < 400 || code >= 500 || eb.Message == "" {
			t.Errorf("%s: %d %q", name, code, eb.Message)
		}
	}
	var l NodeList
	if code, _ := adminCall(t, h, http.MethodGet, AdminNodes, nil, &l); code != 200 || len(l.Nodes) != 3 {
		t.Fatalf("list after refusals: %d nodes", len(l.Nodes))
	}
}

func TestAdminSetDisableEnableRemove(t *testing.T) {
	h := newHarness(t)
	h.seedServers(t)
	v0 := h.latestVersion(t)

	var ch NodeChange
	if code, eb := adminCall(t, h, http.MethodPatch, AdminNodes+"/vps-3", NodeSetRequest{MTU: ptr(1380), CanExit: ptr(true)}, &ch); code != 200 {
		t.Fatalf("set: %d %s", code, eb.Message)
	}
	if ch.Node.MTU != 1380 || !ch.Node.CanExit || ch.Version <= v0 {
		t.Fatalf("after set %+v v%d", ch.Node, ch.Version)
	}
	if code, _ := adminCall(t, h, http.MethodPatch, AdminNodes+"/vps-3", NodeSetRequest{MTU: ptr(9000)}, nil); code != 400 {
		t.Fatalf("bad mtu: %d", code)
	}

	_, _, done, cancel := h.hub.Subscribe(testfix.VPS3.ID, store.FirstMesh)
	defer cancel()
	if code, _ := adminCall(t, h, http.MethodPost, AdminNodes+"/vps-3/disable", nil, &ch); code != 200 || !ch.Node.Disabled {
		t.Fatalf("disable: %d %+v", code, ch.Node)
	}
	select {
	case <-done:
	default:
		t.Fatal("disable must end the node's stream")
	}
	var d NodeDetail
	if code, _ := adminCall(t, h, http.MethodGet, AdminNodes+"/vps-1", nil, &d); code != 200 || len(d.Config.Peers) != 1 {
		t.Fatalf("vps-1 must drop the disabled peer: %+v", d.Config)
	}
	if code, _ := adminCall(t, h, http.MethodPost, AdminNodes+"/vps-3/enable", nil, &ch); code != 200 || ch.Node.Disabled {
		t.Fatalf("enable: %d %+v", code, ch.Node)
	}
	if code, _ := adminCall(t, h, http.MethodDelete, AdminNodes+"/vps-3", nil, &ch); code != 200 || ch.Node.Name != "vps-3" {
		t.Fatalf("rm: %d %+v", code, ch.Node)
	}
	if code, eb := adminCall(t, h, http.MethodGet, AdminNodes+"/vps-3", nil, nil); code != 404 || !strings.Contains(eb.Message, "node list") {
		t.Fatalf("show after rm: %d %q", code, eb.Message)
	}
}

func TestAdminPlanErrorsChangeNothing(t *testing.T) {
	h := newHarness(t)
	h.seedServers(t)
	if code, _ := adminCall(t, h, http.MethodPost, AdminNodes, NodeAddRequest{Name: "laptop", Role: "spoke"}, nil); code != 200 {
		t.Fatal("add laptop")
	}
	v := h.latestVersion(t)
	// An exit through a server that is not exit-capable is a plan error.
	code, eb := adminCall(t, h, http.MethodPatch, AdminNodes+"/laptop", NodeSetRequest{Exit: ptr("vps-1")}, nil)
	if code != 409 || !strings.Contains(eb.Message, "mesh plan would have errors") {
		t.Fatalf("%d %q", code, eb.Message)
	}
	if h.node(t, "laptop").ExitVia != 0 || h.latestVersion(t) != v {
		t.Fatal("a refused change was written")
	}
}

func TestAdminTokenAndShow(t *testing.T) {
	h := newHarness(t)
	h.seedServers(t)
	var ch NodeChange
	if code, _ := adminCall(t, h, http.MethodPost, AdminNodes+"/vps-2/token", nil, &ch); code != 200 || ch.Enrollment == nil {
		t.Fatalf("token: %d %+v", code, ch)
	}
	var d NodeDetail
	if code, _ := adminCall(t, h, http.MethodGet, AdminNodes+"/vps-2", nil, &d); code != 200 || d.Config == nil || d.Node.Name != "vps-2" {
		t.Fatalf("show: %d %+v", code, d)
	}
	if code, _ := adminCall(t, h, http.MethodGet, AdminNodes+"/Bad_Name", nil, nil); code != 400 {
		t.Fatalf("show with a bad name: %d", code)
	}
}

// Container networks over the admin API (amendment A37): a container-mode
// node always gets the lowest free container /24, a host-mode one on
// request; the refusals name what to do; show gives the values a runtime
// creates the network with (amendment A48).
func TestAdminContainerMode(t *testing.T) {
	h := newHarness(t)
	h.seedServers(t)
	add := func(req NodeAddRequest) (NodeChange, int, proto.ErrorBody) {
		var ch NodeChange
		code, eb := adminCall(t, h, http.MethodPost, AdminNodes, req, &ch)
		return ch, code, eb
	}
	ch, code, eb := add(NodeAddRequest{Name: "box", Role: "spoke", Mode: proto.ModeContainer})
	if code != 200 || ch.Node.Mode != proto.ModeContainer || ch.Node.ContainerSubnet != testfix.ContainerSubnet {
		t.Fatalf("container-mode spoke: %d %+v %+v", code, ch.Node, eb)
	}
	ch, code, _ = add(NodeAddRequest{Name: "hub4", Role: "server", Mode: proto.ModeContainer, Endpoint: "198.51.100.7:51999"})
	if code != 200 || ch.Node.ContainerSubnet != "10.77.129.0/24" {
		t.Fatalf("container-mode server: %d %+v", code, ch.Node)
	}
	ch, code, _ = add(NodeAddRequest{Name: "pc", Role: "spoke", Containers: true})
	if code != 200 || ch.Node.Mode != proto.ModeHost || ch.Node.ContainerSubnet != "10.77.130.0/24" {
		t.Fatalf("host-mode spoke with containers: %d %+v", code, ch.Node)
	}
	for _, c := range []struct {
		req  NodeAddRequest
		want string
	}{
		{NodeAddRequest{Name: "hub5", Role: "server", Mode: proto.ModeContainer}, "needs an endpoint"},
		{NodeAddRequest{Name: "hub5", Role: "server", Mode: proto.ModeContainer, Endpoint: "198.51.100.8:51820", CanExit: true}, "cannot be an exit"},
		{NodeAddRequest{Name: "ph", Role: "phone", Mode: proto.ModeContainer}, "no agent"},
		{NodeAddRequest{Name: "ph", Role: "phone", Containers: true}, "no agent"},
		{NodeAddRequest{Name: "odd", Role: "spoke", Mode: "bridge"}, "mode"},
	} {
		if _, code, eb := add(c.req); code != 400 || !strings.Contains(eb.Message, c.want) {
			t.Errorf("%+v = %d %q, want 400 mentioning %q", c.req, code, eb.Message, c.want)
		}
	}
	var d NodeDetail
	if code, _ := adminCall(t, h, http.MethodGet, AdminNodes+"/box", nil, &d); code != 200 || d.Node.ContainerSubnet != testfix.ContainerSubnet {
		t.Fatalf("show box: %d %+v", code, d.Node)
	}

	set := func(name string, req NodeSetRequest) (NodeChange, int, proto.ErrorBody) {
		var ch NodeChange
		code, eb := adminCall(t, h, http.MethodPatch, AdminNodes+"/"+name, req, &ch)
		return ch, code, eb
	}
	if _, code, eb := set("box", NodeSetRequest{Containers: ptr(false)}); code != 400 || !strings.Contains(eb.Message, "always on the mesh") {
		t.Errorf("container mode off: %d %+v", code, eb)
	}
	if ch, code, _ := set("pc", NodeSetRequest{Containers: ptr(false)}); code != 200 || ch.Node.ContainerSubnet != "" {
		t.Errorf("containers off: %d %+v", code, ch.Node)
	}
	if ch, code, _ := set("pc", NodeSetRequest{Containers: ptr(true)}); code != 200 || ch.Node.ContainerSubnet != "10.77.130.0/24" {
		t.Errorf("containers on again: %d %+v", code, ch.Node)
	}
	if ch, code, _ := set("pc", NodeSetRequest{Containers: ptr(true)}); code != 200 || ch.Node.ContainerSubnet != "10.77.130.0/24" {
		t.Errorf("containers on twice keeps the subnet: %d %+v", code, ch.Node)
	}

	// The values docs/containers.md's samples take: container mode with the
	// agent at .2 and every server node as a Swarm service's resolver (hub4
	// has not joined, so it is in no peer list), host mode on the agent's
	// own bridge, a container-mode server's published port from its
	// endpoint.
	show := func(name string) *ContainerNetwork {
		var d NodeDetail
		if code, eb := adminCall(t, h, http.MethodGet, AdminNodes+"/"+name, nil, &d); code != 200 {
			t.Fatalf("show %s: %d %+v", name, code, eb)
		}
		return d.Containers
	}
	servers := []string{"10.77.0.1", "10.77.0.2", "10.77.0.3"}
	for name, want := range map[string]*ContainerNetwork{
		"box": {Name: proto.MeshNetwork, BridgeName: proto.ContainerBridge, Mode: proto.ModeContainer, Subnet: testfix.ContainerSubnet,
			Network: testfix.MeshCIDR, Range: "10.77.128.128/25", Bridge: "10.77.128.1", Gateway: "10.77.128.2", Zone: "mesh", Resolvers: servers},
		"pc": {Name: proto.MeshNetwork, BridgeName: proto.ContainerBridge, Mode: proto.ModeHost, Subnet: "10.77.130.0/24",
			Network: testfix.MeshCIDR, Range: "10.77.130.128/25", Bridge: "10.77.130.1", Gateway: "10.77.130.1", Zone: "mesh", Resolvers: servers},
		"hub4": {Name: proto.MeshNetwork, BridgeName: proto.ContainerBridge, Mode: proto.ModeContainer, Subnet: "10.77.129.0/24",
			Network: testfix.MeshCIDR, Range: "10.77.129.128/25", Bridge: "10.77.129.1", Gateway: "10.77.129.2", Zone: "mesh", Resolvers: servers,
			PublishPort: 51999, ListenPort: 51820},
		"vps-1": nil,
	} {
		if got := show(name); !reflect.DeepEqual(got, want) {
			t.Errorf("%s's container network = %+v, want %+v", name, got, want)
		}
	}
	if _, code, _ := set("hub4", NodeSetRequest{Endpoint: ptr("")}); code != 200 {
		t.Fatalf("clear hub4's endpoint: %d", code)
	}
	if got := show("hub4"); got == nil || got.PublishPort != 0 || got.ListenPort != 51820 {
		t.Errorf("a container-mode server without an endpoint publishes no known port: %+v", got)
	}
	if _, code, eb := add(NodeAddRequest{Name: "ph2", Role: "phone"}); code != 200 {
		t.Fatalf("add a phone: %d %+v", code, eb)
	}
	if _, code, eb := set("ph2", NodeSetRequest{Containers: ptr(true)}); code != 400 || !strings.Contains(eb.Message, "phone") {
		t.Errorf("containers on a phone = %d %+v", code, eb)
	}
}

// A mesh too small for container networks refuses them where they are
// asked for, with what to do.
func TestContainerNetworksNeedRoom(t *testing.T) {
	h := newHarness(t)
	h.seedServers(t)
	if err := h.st.MeshTx(context.Background(), store.FirstMesh, func(tx *store.MeshTx) error { return tx.SetSetting(store.SettingMeshCIDR, "10.77.0.0/22") }); err != nil {
		t.Fatal(err)
	}
	for _, req := range []NodeAddRequest{{Name: "box", Role: "spoke", Mode: proto.ModeContainer}, {Name: "pc", Role: "spoke", Containers: true}} {
		if code, eb := adminCall(t, h, http.MethodPost, AdminNodes, req, nil); code != 409 || !strings.Contains(eb.Message, "/21 or larger") {
			t.Errorf("%+v = %d %+v", req, code, eb)
		}
	}
	if code, eb := adminCall(t, h, http.MethodPost, AdminNodes, NodeAddRequest{Name: "pc", Role: "spoke"}, nil); code != 200 {
		t.Fatalf("add pc: %d %+v", code, eb)
	}
	if code, eb := adminCall(t, h, http.MethodPatch, AdminNodes+"/pc", NodeSetRequest{Containers: ptr(true)}, nil); code != 409 ||
		!strings.Contains(eb.Message, "/21 or larger") {
		t.Errorf("set --containers in a /22 = %d %+v", code, eb)
	}
}

func TestAdminClientOverTheSocket(t *testing.T) {
	h := newHarness(t)
	h.seedServers(t)
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

	c := NewAdminClient(sock)
	ctx := context.Background()
	if l, err := c.List(ctx); err != nil || len(l.Nodes) != 3 {
		t.Fatalf("List = %d, %v", len(l.Nodes), err)
	}
	ch, err := c.Add(ctx, NodeAddRequest{Name: "laptop", Role: "spoke"})
	if err != nil || ch.Enrollment == nil {
		t.Fatalf("Add = %+v, %v", ch, err)
	}
	if _, err := c.Add(ctx, NodeAddRequest{Name: "laptop", Role: "spoke"}); err == nil {
		t.Fatal("duplicate Add succeeded")
	} else if ae, ok := err.(*AdminError); !ok || ae.Code != proto.CodeConflict {
		t.Fatalf("duplicate Add = %#v", err)
	}
	if d, err := c.Show(ctx, "laptop"); err != nil || d.Node.Name != "laptop" || d.Containers != nil {
		t.Fatalf("Show = %+v, %v", d, err)
	}
	if _, err := c.Add(ctx, NodeAddRequest{Name: "phone", Role: "phone"}); err != nil {
		t.Fatal(err)
	}
	if p, err := c.Phone(ctx, "phone", ""); err != nil || !strings.HasPrefix(p.Profile, "[Interface]\nPrivateKey = ") {
		t.Fatalf("Phone = %+v, %v", p, err)
	}
	if v, err := c.SetIPv6(ctx, IPv6Request{Prefix: testfix.MeshCIDR6}); err != nil || v.Prefix != testfix.MeshCIDR6 || v.Addressed != 5 {
		t.Fatalf("SetIPv6 = %+v, %v", v, err)
	}
	if v, err := c.IPv6(ctx); err != nil || v.Prefix != testfix.MeshCIDR6 {
		t.Fatalf("IPv6 = %+v, %v", v, err)
	}
	if _, err := c.SetIPv6(ctx, IPv6Request{Prefix: "fd10::/40"}); err == nil {
		t.Fatal("SetIPv6 of a /40 succeeded")
	} else if ae, ok := err.(*AdminError); !ok || ae.Code != proto.CodeBadRequest {
		t.Fatalf("SetIPv6(/40) = %#v", err)
	}
	for _, step := range []func() (NodeChange, error){
		func() (NodeChange, error) { return c.Set(ctx, "laptop", NodeSetRequest{Keepalive: ptr(15)}) },
		func() (NodeChange, error) { return c.Token(ctx, "laptop") },
		func() (NodeChange, error) { return c.Disable(ctx, "laptop") },
		func() (NodeChange, error) { return c.Enable(ctx, "laptop") },
		func() (NodeChange, error) { return c.Remove(ctx, "laptop") },
	} {
		if _, err := step(); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := c.Show(ctx, "laptop"); err == nil {
		t.Fatal("Show after Remove succeeded")
	}
	if _, err := NewAdminClient(filepath.Join(dir, "none.sock")).List(ctx); err == nil || !strings.Contains(err.Error(), "meshctl control") {
		t.Fatalf("no server: %v", err)
	}
}

// Every admin write answers each failure with its own code, and none of them
// changes anything.
func TestAdminErrorCodes(t *testing.T) {
	h := newHarness(t)
	h.seedServers(t)
	h.seed(t, storeNode(testfix.Phone, false))
	if code, _ := adminCall(t, h, http.MethodPost, AdminNodes+"/vps-3/disable", nil, nil); code != 200 {
		t.Fatal("disable vps-3")
	}
	v := h.latestVersion(t)
	bad := "Bad Name"
	cases := []struct {
		name   string
		method string
		path   string
		body   any
		status int
	}{
		{"set unknown node", http.MethodPatch, AdminNodes + "/ghost", NodeSetRequest{MTU: ptr(1400)}, 404},
		{"disable unknown node", http.MethodPost, AdminNodes + "/ghost/disable", nil, 404},
		{"enable unknown node", http.MethodPost, AdminNodes + "/ghost/enable", nil, 404},
		{"rm unknown node", http.MethodDelete, AdminNodes + "/ghost", nil, 404},
		{"token unknown node", http.MethodPost, AdminNodes + "/ghost/token", nil, 404},
		{"token for a phone", http.MethodPost, AdminNodes + "/phone/token", nil, 400},
		{"token for a disabled node", http.MethodPost, AdminNodes + "/vps-3/token", nil, 409},
		{"set listen port 0", http.MethodPatch, AdminNodes + "/vps-1", NodeSetRequest{ListenPort: ptr(0)}, 400},
		{"set negative keepalive", http.MethodPatch, AdminNodes + "/vps-1", NodeSetRequest{Keepalive: ptr(-1)}, 400},
		{"set bad hub name", http.MethodPatch, AdminNodes + "/vps-1", NodeSetRequest{Hub: &bad}, 400},
		{"set bad exit name", http.MethodPatch, AdminNodes + "/vps-1", NodeSetRequest{Exit: &bad}, 400},
		{"set endpoint with a host", http.MethodPatch, AdminNodes + "/vps-1", NodeSetRequest{Endpoint: ptr("vps.example:51820")}, 400},
		{"set unknown hub", http.MethodPatch, AdminNodes + "/phone", NodeSetRequest{Hub: ptr("ghost")}, 404},
		{"set unknown exit", http.MethodPatch, AdminNodes + "/phone", NodeSetRequest{Exit: ptr("ghost")}, 404},
		{"set unknown field", http.MethodPatch, AdminNodes + "/vps-1", map[string]any{"colour": "red"}, 400},
		{"add can_exit spoke", http.MethodPost, AdminNodes, NodeAddRequest{Name: "x", Role: "spoke", CanExit: true}, 400},
		{"add keepalive out of range", http.MethodPost, AdminNodes, NodeAddRequest{Name: "x", Role: "spoke", Keepalive: 5000}, 400},
		{"add listen port out of range", http.MethodPost, AdminNodes, NodeAddRequest{Name: "x", Role: "server", ListenPort: 70000}, 400},
		{"add bad hub name", http.MethodPost, AdminNodes, NodeAddRequest{Name: "x", Role: "spoke", Hub: "Bad Name"}, 400},
		{"add bad ip", http.MethodPost, AdminNodes, NodeAddRequest{Name: "x", Role: "server", IP: "10.77.0.300"}, 400},
		{"add not json", http.MethodPost, AdminNodes, "not an object", 400},
		{"profile for an unknown node", http.MethodPost, AdminNodes + "/ghost/phone", PhoneRequest{}, 404},
		{"profile for a server", http.MethodPost, AdminNodes + "/vps-1/phone", PhoneRequest{}, 409},
		{"profile with a bad exit name", http.MethodPost, AdminNodes + "/phone/phone", PhoneRequest{Exit: bad}, 400},
		{"profile via an unknown exit", http.MethodPost, AdminNodes + "/phone/phone", PhoneRequest{Exit: "ghost"}, 404},
		{"profile via a server that is no exit", http.MethodPost, AdminNodes + "/phone/phone", PhoneRequest{Exit: "vps-1"}, 409},
		{"profile with an unknown field", http.MethodPost, AdminNodes + "/phone/phone", map[string]any{"colour": "red"}, 400},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			code, eb := adminCall(t, h, c.method, c.path, c.body, nil)
			if code != c.status || eb.Message == "" {
				t.Fatalf("%d %q, want %d with a message", code, eb.Message, c.status)
			}
		})
	}
	var l NodeList
	if code, _ := adminCall(t, h, http.MethodGet, AdminNodes, nil, &l); code != 200 || len(l.Nodes) != 4 || h.latestVersion(t) != v {
		t.Fatalf("refused writes changed state: %d nodes, version %d (was %d)", len(l.Nodes), h.latestVersion(t), v)
	}
}

func TestAdminSetClearsReferences(t *testing.T) {
	h := newHarness(t)
	h.seedServers(t)
	var ch NodeChange
	if code, eb := adminCall(t, h, http.MethodPatch, AdminNodes+"/vps-3", NodeSetRequest{Endpoint: ptr("")}, &ch); code != 200 {
		t.Fatalf("clear endpoint: %d %s", code, eb.Message)
	}
	if ch.Node.Endpoint != "" || len(ch.Warnings) == 0 {
		t.Fatalf("an endpointless server must warn: %+v %v", ch.Node, ch.Warnings)
	}
	if code, _ := adminCall(t, h, http.MethodPatch, AdminNodes+"/vps-3", NodeSetRequest{Endpoint: ptr("203.0.113.12:51820"), Keepalive: ptr(0)}, &ch); code != 200 || ch.Node.Endpoint != "203.0.113.12:51820" {
		t.Fatalf("set endpoint back: %d %+v", code, ch.Node)
	}
}

// A node moves to another mesh IP (amendment A41): every peer and its DNS
// name follow in a new version, and the notes say what does not -- a rule
// naming the old address, a server's Swarm, a phone's profile. The same
// address changes nothing; a taken, outside, broadcast or malformed one is
// refused, and one inside a container subnet is the plan's to refuse.
func TestMoveMeshIP(t *testing.T) {
	h := newHarness(t)
	nas := storeNode(testfix.NAS, true)
	nas.ContainerSubnet = testfix.ContainerSubnet
	h.seed(t, storeNode(testfix.VPS1, true), storeNode(testfix.VPS2, true), storeNode(testfix.VPS3, true), nas,
		storeNode(testfix.Laptop, true), storeNode(testfix.Phone, true))
	if err := h.st.MeshTx(context.Background(), store.FirstMesh, func(tx *store.MeshTx) error {
		return tx.SetSegmentation(store.LegacySegmentation([]store.FirewallRule{
			{ID: 1, NodeID: testfix.VPS1.ID, Proto: "tcp", Port: 5432, SourceCIDR: testfix.Laptop.MeshIP + "/32"}}, map[int64]string{testfix.VPS1.ID: "vps-1"}))
	}); err != nil {
		t.Fatal(err)
	}
	move := func(name, ip string) (NodeChange, int, proto.ErrorBody) {
		t.Helper()
		var ch NodeChange
		code, eb := adminCall(t, h, http.MethodPatch, AdminNodes+"/"+name, NodeSetRequest{MeshIP: ptr(ip)}, &ch)
		return ch, code, eb
	}
	before := h.latestVersion(t)
	ch, code, eb := move("laptop", "10.77.2.9")
	if code != 200 || ch.Node.MeshIP != "10.77.2.9" || ch.Version <= before {
		t.Fatalf("move the laptop: %d %+v %+v", code, ch, eb)
	}
	notes := strings.Join(ch.Notes, "\n")
	if !strings.Contains(notes, "laptop moves from 10.77.2.1 to 10.77.2.9") || !strings.Contains(notes, `policy "firewall-1" names 10.77.2.1/32 by address`) {
		t.Errorf("notes:\n%s", notes)
	}
	vps1 := h.config(t, "vps-1")
	var peer, record string
	for _, p := range vps1.Peers {
		if p.Name == "laptop" {
			peer = strings.Join(p.AllowedIPs, ",")
		}
	}
	for _, r := range vps1.DNS.Records {
		if r.Name == "laptop" {
			record = r.IP
		}
	}
	if peer != "10.77.2.9/32" || record != "10.77.2.9" {
		t.Errorf("vps-1 after the move: peer %q, record %q", peer, record)
	}
	if ch, code, _ := move("laptop", "10.77.2.9"); code != 200 || ch.Version != 0 || len(ch.Notes) != 0 {
		t.Errorf("the same address: %d %+v", code, ch)
	}
	for ip, want := range map[string]string{
		testfix.VPS1.MeshIP: "already assigned",
		"192.168.1.5":       "outside the mesh",
		"10.77.2.255":       "broadcast",
		"laptop":            "mesh_ip",
	} {
		if _, code, eb := move("laptop", ip); code != 400 || !strings.Contains(eb.Message, want) {
			t.Errorf("move to %q: %d %q, want 400 with %q", ip, code, eb.Message, want)
		}
	}
	if _, code, eb := move("laptop", "10.77.128.5"); code != 409 || !strings.Contains(eb.Message, "container subnet") {
		t.Errorf("into the NAS's container subnet: %d %q", code, eb.Message)
	}
	if h.node(t, "laptop").MeshIP != "10.77.2.9" {
		t.Fatal("a refused move changed the address")
	}
	if ch, _, _ := move("vps-3", "10.77.0.9"); !strings.Contains(strings.Join(ch.Notes, "\n"), "--advertise-addr 10.77.0.9") {
		t.Errorf("a server's notes: %v", ch.Notes)
	}
	if ch, _, _ := move("phone", "10.77.3.9"); !strings.Contains(strings.Join(ch.Notes, "\n"), "meshctl node phone phone") {
		t.Errorf("a phone's notes: %v", ch.Notes)
	}
}
