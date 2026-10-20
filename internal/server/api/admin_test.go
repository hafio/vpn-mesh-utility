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
	"strings"
	"testing"

	"github.com/hafio/vpn-mesh-utility/deploy"
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
	if err := h.st.Tx(context.Background(), func(tx *store.Tx) error {
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

	_, _, done, cancel := h.hub.Subscribe(testfix.VPS3.ID)
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

// A spoke's compose file carries the server's own release and origin;
// servers, unknown nodes and bad choices are refused before anything is
// rendered.
func TestAdminCompose(t *testing.T) {
	h := newHarness(t)
	h.seedServers(t)
	if code, eb := adminCall(t, h, http.MethodPost, AdminNodes, NodeAddRequest{Name: "nas", Role: "spoke"}, nil); code != 200 {
		t.Fatalf("add: %d %+v", code, eb)
	}
	var f ComposeFile
	code, _ := adminCall(t, h, http.MethodGet, AdminNodes+"/nas/compose", nil, &f)
	if code != 200 || f.Backend != deploy.NASUserspace || f.DataDir != deploy.DefaultNASData ||
		!strings.Contains(f.YAML, "image: "+deploy.Image+":1.0.0") || !strings.Contains(f.YAML, "--server "+testURL+" --name nas") {
		t.Fatalf("compose: %d %+v", code, f)
	}
	code, _ = adminCall(t, h, http.MethodGet, AdminNodes+"/nas/compose?backend=kernel&data=/srv/meshctl", nil, &f)
	if code != 200 || f.Backend != deploy.NASKernel || !strings.Contains(f.YAML, "/srv/meshctl:/data") || strings.Contains(f.YAML, "devices:") {
		t.Fatalf("kernel compose: %d %+v", code, f)
	}
	for _, c := range []struct {
		path string
		want int
	}{
		{"/vps-1/compose", 409},
		{"/none/compose", 404},
		{"/Bad_Name/compose", 400},
		{"/nas/compose?backend=wireguard", 400},
		{"/nas/compose?data=/srv/%24HOME", 400},
		{"/nas/compose?data=relative", 400},
	} {
		if code, eb := adminCall(t, h, http.MethodGet, AdminNodes+c.path, nil, nil); code != c.want || eb.Message == "" {
			t.Errorf("%s = %d %+v, want %d", c.path, code, eb, c.want)
		}
	}
	dev := NewNodes(Deps{DB: h.db, ServerURL: testURL, Version: "dev"})
	if _, err := dev.Compose(context.Background(), "nas", ComposeOptions{}); err == nil {
		t.Fatal("a development build rendered a compose file")
	} else if status, msg := ErrorMessage(err); status != 400 || !strings.Contains(msg, "development build") {
		t.Fatalf("a development build: %d %s", status, msg)
	}
}

// Container networks over the admin API (amendment A37): a container-mode
// node always gets the lowest free container /24, a host-mode one on
// request; the refusals name what to do; the compose file follows the
// node's mode and the runtime.
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

	var f ComposeFile
	if code, eb := adminCall(t, h, http.MethodGet, AdminNodes+"/box/compose", nil, &f); code != 200 || f.Runtime != deploy.RuntimeDocker ||
		f.Backend != deploy.ContainerAuto || !strings.Contains(f.YAML, "docker network create --internal") ||
		!strings.Contains(f.YAML, "dns: [10.77.128.2]") || strings.Contains(f.YAML, "ports:") || strings.Contains(f.YAML, "network_mode") ||
		!strings.Contains(f.YAML, "docker.sock") {
		t.Fatalf("box compose: %d %+v\n%s", code, eb, f.YAML)
	}
	if code, _ := adminCall(t, h, http.MethodGet, AdminNodes+"/box/compose?names=off", nil, &f); code != 200 || strings.Contains(f.YAML, ".sock") {
		t.Errorf("compose without names: %d\n%s", code, f.YAML)
	}
	for runtime, want := range map[string]string{"podman": "podman network create --internal --disable-dns", "swarm": "--scope swarm"} {
		if code, _ := adminCall(t, h, http.MethodGet, AdminNodes+"/box/compose?runtime="+runtime, nil, &f); code != 200 || !strings.Contains(f.YAML, want) {
			t.Errorf("%s compose: %d\n%s", runtime, code, f.YAML)
		}
	}
	// A Swarm service resolves through the servers in peer lists (hub4 has
	// not joined).
	if code, _ := adminCall(t, h, http.MethodGet, AdminNodes+"/box/compose?runtime=swarm", nil, &f); code != 200 ||
		!strings.Contains(f.YAML, "dns: [10.77.0.1, 10.77.0.2, 10.77.0.3]") {
		t.Errorf("the swarm service's resolvers: %d\n%s", code, f.YAML)
	}
	if code, _ := adminCall(t, h, http.MethodGet, AdminNodes+"/hub4/compose", nil, &f); code != 200 || !strings.Contains(f.YAML, "51999:51820/udp") {
		t.Errorf("hub compose publishes the endpoint's port: %d\n%s", code, f.YAML)
	}
	// A host-mode spoke (pc) gets the NAS file, which runs with any runtime.
	for path, want := range map[string]int{
		"/box/compose?runtime=lxc":    400,
		"/box/compose?backend=wg":     400,
		"/pc/compose?runtime=podman":  400,
		"/box/compose?data=/srv/$x":   400,
		"/box/compose?data=relative":  400,
		"/hub4/compose?runtime=swarm": 200,
	} {
		if code, eb := adminCall(t, h, http.MethodGet, AdminNodes+path, nil, nil); code != want {
			t.Errorf("%s = %d %+v, want %d", path, code, eb, want)
		}
	}
	if _, code, _ := set("hub4", NodeSetRequest{Endpoint: ptr("")}); code != 200 {
		t.Fatalf("clear hub4's endpoint: %d", code)
	}
	if code, eb := adminCall(t, h, http.MethodGet, AdminNodes+"/hub4/compose", nil, nil); code != 409 || !strings.Contains(eb.Message, "no endpoint") {
		t.Errorf("a hub without an endpoint: %d %+v", code, eb)
	}
	// The mesh network's commands for its host: container mode as in the
	// compose header, host mode on the agent's own bridge.
	var ni NetworkInfo
	if code, eb := adminCall(t, h, http.MethodGet, AdminNodes+"/pc/network?runtime=podman", nil, &ni); code != 200 || ni.Mode != proto.ModeHost ||
		ni.Subnet != "10.77.130.0/24" || !strings.Contains(strings.Join(ni.Commands, " "), "-o mode=unmanaged") ||
		!strings.Contains(strings.Join(ni.Service, " "), "dns: [10.77.130.1]") {
		t.Fatalf("pc network: %d %+v %+v", code, eb, ni)
	}
	if code, _ := adminCall(t, h, http.MethodGet, AdminNodes+"/box/network", nil, &ni); code != 200 || ni.Runtime != deploy.RuntimeDocker ||
		!strings.Contains(strings.Join(ni.Commands, " "), "docker network create --internal") {
		t.Fatalf("box network: %d %+v", code, ni)
	}
	for path, want := range map[string]int{"/vps-1/network": 409, "/pc/network?runtime=lxc": 400, "/none/network": 404} {
		if code, eb := adminCall(t, h, http.MethodGet, AdminNodes+path, nil, nil); code != want || eb.Message == "" {
			t.Errorf("%s = %d %+v, want %d", path, code, eb, want)
		}
	}
	if code, _ := adminCall(t, h, http.MethodGet, AdminNodes+"/Bad_Name/network", nil, nil); code != 400 {
		t.Errorf("a bad name's network = %d", code)
	}
	if _, code, eb := add(NodeAddRequest{Name: "ph2", Role: "phone"}); code != 200 {
		t.Fatalf("add a phone: %d %+v", code, eb)
	}
	if _, code, eb := set("ph2", NodeSetRequest{Containers: ptr(true)}); code != 400 || !strings.Contains(eb.Message, "phone") {
		t.Errorf("containers on a phone = %d %+v", code, eb)
	}
	// deploy repeats these names rather than import proto.
	if deploy.MeshNetwork != proto.MeshNetwork || deploy.MeshBridge != proto.ContainerBridge {
		t.Errorf("deploy names %s/%s, proto %s/%s", deploy.MeshNetwork, deploy.MeshBridge, proto.MeshNetwork, proto.ContainerBridge)
	}
}

// A mesh too small for container networks refuses them where they are
// asked for, with what to do.
func TestContainerNetworksNeedRoom(t *testing.T) {
	h := newHarness(t)
	h.seedServers(t)
	if err := h.st.Tx(context.Background(), func(tx *store.Tx) error { return tx.SetSetting(store.SettingMeshCIDR, "10.77.0.0/22") }); err != nil {
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
	if f, err := c.Compose(ctx, "laptop", ComposeOptions{}); err != nil || f.Backend != "userspace" {
		t.Fatalf("Compose = %+v, %v", f, err)
	}
	if f, err := c.Compose(ctx, "laptop", ComposeOptions{Backend: "kernel", DataDir: "/srv/meshctl"}); err != nil ||
		f.Backend != "kernel" || f.DataDir != "/srv/meshctl" {
		t.Fatalf("Compose(kernel) = %+v, %v", f, err)
	}
	if _, err := c.Add(ctx, NodeAddRequest{Name: "phone", Role: "phone"}); err != nil {
		t.Fatal(err)
	}
	if p, err := c.Phone(ctx, "phone", ""); err != nil || !strings.HasPrefix(p.Profile, "[Interface]\nPrivateKey = ") {
		t.Fatalf("Phone = %+v, %v", p, err)
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
	if _, err := NewAdminClient(filepath.Join(dir, "none.sock")).List(ctx); err == nil || !strings.Contains(err.Error(), "meshctl server") {
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
