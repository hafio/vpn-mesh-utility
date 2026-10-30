package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/hafio/vpn-mesh-utility/internal/agent"
	"github.com/hafio/vpn-mesh-utility/internal/agent/localapi"
	"github.com/hafio/vpn-mesh-utility/internal/agent/state"
	"github.com/hafio/vpn-mesh-utility/internal/proto"
	"github.com/hafio/vpn-mesh-utility/internal/server"
	"github.com/hafio/vpn-mesh-utility/internal/server/api"
	"github.com/hafio/vpn-mesh-utility/internal/server/export"
	"github.com/hafio/vpn-mesh-utility/internal/server/store"
	"github.com/hafio/vpn-mesh-utility/internal/testfix"
)

// shortDir is a temp dir short enough to hold a Unix socket on every runner.
func shortDir(t *testing.T) string {
	t.Helper()
	d, err := os.MkdirTemp("", "mcl")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(d) })
	return d
}

// startAdmin serves a real admin API on <dir>/admin.sock.
func startAdmin(t *testing.T) (string, *store.Store) {
	t.Helper()
	dir := shortDir(t)
	st, err := store.Open(context.Background(), filepath.Join(dir, server.DBFile), store.Options{Now: testfix.NewClock(testfix.Epoch).Now})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = st.Close() })
	hub := api.NewHub()
	d := api.Deps{DB: st, Hub: hub, Reconciler: api.NewReconciler(st, hub, nil), ServerURL: "https://mesh.example.com", Version: "v1.2.3"}
	ln, err := server.ListenAdmin(filepath.Join(dir, server.AdminSocket))
	if err != nil {
		t.Fatal(err)
	}
	srv := &http.Server{Handler: api.NewAdmin(d)}
	go func() { _ = srv.Serve(ln) }()
	t.Cleanup(func() { _ = srv.Close() })
	return dir, st
}

func TestNodeLifecycle(t *testing.T) {
	data, _ := startAdmin(t)
	code, out, errOut := runArgs("node", "add", "vps-1", "--role", "server", "--endpoint", "203.0.113.10:51820", "--data", data)
	if code != exitOK || !strings.Contains(out, "added vps-1 (server, mesh IP 10.77.0.1)") ||
		!strings.Contains(out, "curl -fsSL https://mesh.example.com/install.sh | sudo sh -s -- --mesh default --token ") ||
		!strings.Contains(out, "$env:MESHCTL_MESH = 'default'; irm https://mesh.example.com/install.ps1 | iex") {
		t.Fatalf("add: %d\n%s%s", code, out, errOut)
	}
	code, out, _ = runArgs("node", "add", "laptop", "--role", "spoke", "--data", data, "--json")
	var ch api.NodeChange
	if code != exitOK || json.Unmarshal([]byte(out), &ch) != nil || ch.Enrollment == nil || ch.Node.MeshIP != "10.77.2.1" {
		t.Fatalf("add --json: %d %s", code, out)
	}
	code, out, _ = runArgs("node", "list", "--data", data)
	if code != exitOK || !strings.Contains(out, "vps-1") || !strings.Contains(out, "not joined") || !strings.Contains(out, "203.0.113.10:51820") {
		t.Fatalf("list: %d\n%s", code, out)
	}
	if code, out, _ = runArgs("node", "show", "vps-1", "--data", data); code != exitOK || !strings.Contains(out, "10.77.0.1") {
		t.Fatalf("show: %d\n%s", code, out)
	}
	// Container mode (amendment A37).
	code, out, _ = runArgs("node", "add", "box", "--role", "spoke", "--mode", "container", "--data", data)
	if code != exitOK || !strings.Contains(out, "its containers: 10.77.128.0/24") || !strings.Contains(out, "meshctl node show box") {
		t.Fatalf("add --mode container: %d\n%s", code, out)
	}
	// Show gives the container network's values for docs/containers.md
	// (amendment A48): container mode with the agent's own address, host
	// mode on the agent's bridge.
	code, out, _ = runArgs("node", "show", "box", "--data", data)
	for _, want := range []string{"container: only its containers", "containers on the mesh", "meshctl-mesh\n", "10.77.128.128/25",
		"10.77.128.2, search mesh, ndots:1"} {
		if code != exitOK || !strings.Contains(out, want) {
			t.Errorf("show a container-mode node lacks %q: %d\n%s", want, code, out)
		}
	}
	if code, out, _ = runArgs("node", "set", "laptop", "--containers", "--data", data, "--json"); code != exitOK || !strings.Contains(out, `"container_subnet": "10.77.129.0/24"`) {
		t.Fatalf("set --containers: %d\n%s", code, out)
	}
	if code, out, _ = runArgs("node", "show", "laptop", "--data", data); code != exitOK || !strings.Contains(out, "the agent owns its bridge meshctl0") ||
		!strings.Contains(out, "10.77.129.1, search mesh") || strings.Contains(out, "  agent ") {
		t.Fatalf("show a host-mode node's containers: %d\n%s", code, out)
	}
	if code, out, _ = runArgs("node", "show", "vps-1", "--data", data); code != exitOK || strings.Contains(out, "containers on the mesh") {
		t.Fatalf("show a node without containers: %d\n%s", code, out)
	}
	if code, _, errOut = runArgs("node", "set", "box", "--containers=false", "--data", data); code != exitFailure || !strings.Contains(errOut, "always on the mesh") {
		t.Fatalf("set --containers=false in container mode: %d %s", code, errOut)
	}
	if code, _, errOut = runArgs("node", "add", "odd", "--role", "spoke", "--mode", "bridge", "--data", data); code != exitFailure || !strings.Contains(errOut, "mode") {
		t.Fatalf("add --mode bridge: %d %s", code, errOut)
	}
	if code, out, _ = runArgs("node", "set", "laptop", "--mtu", "1380", "--hub", "vps-1", "--data", data); code != exitOK || !strings.Contains(out, "updated laptop") {
		t.Fatalf("set: %d\n%s", code, out)
	}
	if code, out, _ = runArgs("node", "set", "laptop", "--hub", "default", "--data", data, "--json"); code != exitOK || !strings.Contains(out, `"name": "laptop"`) || strings.Contains(out, `"hub"`) {
		t.Fatalf("set --hub default must clear the hub: %d\n%s", code, out)
	}
	if code, out, _ = runArgs("node", "set", "laptop", "--mesh-ip", "10.77.2.9", "--data", data); code != exitOK ||
		!strings.Contains(out, "laptop moves from 10.77.2.1 to 10.77.2.9") {
		t.Fatalf("set --mesh-ip: %d\n%s", code, out)
	}
	if code, _, errOut = runArgs("node", "set", "laptop", "--data", data); code != exitUsage || !strings.Contains(errOut, "nothing to change") {
		t.Fatalf("set without changes: %d %s", code, errOut)
	}
	for _, verb := range []string{"disable", "enable", "token"} {
		if code, out, errOut = runArgs("node", verb, "laptop", "--data", data); code != exitOK {
			t.Fatalf("%s: %d %s%s", verb, code, out, errOut)
		}
	}
	if !strings.Contains(out, "join token") {
		t.Fatalf("token must print the enrollment:\n%s", out)
	}
	if code, out, _ = runArgs("node", "rm", "laptop", "--data", data); code != exitOK || !strings.Contains(out, "removed laptop") {
		t.Fatalf("rm: %d %s", code, out)
	}
	code, _, errOut = runArgs("node", "add", "x", "--role", "router", "--data", data)
	if code != exitFailure || !strings.Contains(errOut, "role") {
		t.Fatalf("server-side refusal: %d %s", code, errOut)
	}
	if code, _, errOut = runArgs("node", "add", "x", "--data", data); code != exitUsage || !strings.Contains(errOut, "--role is required") {
		t.Fatalf("missing role: %d %s", code, errOut)
	}
}

// A phone gets no join command; `node phone` prints its profile alone on
// stdout (so it can be redirected to a file) and the node then shows that
// profile as current.
func TestNodePhone(t *testing.T) {
	data, st := startAdmin(t)
	if code, out, errOut := runArgs("node", "add", "vps-1", "--role", "server", "--endpoint", testfix.VPS1.Endpoint, "--data", data); code != exitOK {
		t.Fatalf("add vps-1: %d %s%s", code, out, errOut)
	}
	// vps-1 joins.
	if err := st.MeshTx(context.Background(), store.FirstMesh, func(tx *store.MeshTx) error {
		n, err := tx.NodeByName("vps-1")
		if err != nil {
			return err
		}
		n.Enrolled, n.WGPubKey, n.IdentityPubKey = true, testfix.WGKey(1), testfix.IdentityKeyB64(1)
		return tx.UpdateNode(n)
	}); err != nil {
		t.Fatal(err)
	}
	code, out, _ := runArgs("node", "add", "phone", "--role", "phone", "--data", data)
	if code != exitOK || strings.Contains(out, "join token") || !strings.Contains(out, "meshctl node phone phone") {
		t.Fatalf("add a phone: %d\n%s", code, out)
	}
	code, out, errOut := runArgs("node", "phone", "phone", "--data", data)
	if code != exitOK || !strings.HasPrefix(out, "[Interface]\nPrivateKey = ") || !strings.Contains(out, "DNS = "+testfix.VPS1.MeshIP+"\n") ||
		!strings.Contains(errOut, "shown once") || strings.Contains(errOut, "PrivateKey") {
		t.Fatalf("phone: %d\n%s---\n%s", code, out, errOut)
	}
	var p api.PhoneProfile
	if code, out, _ = runArgs("node", "phone", "phone", "--json", "--data", data); code != exitOK || json.Unmarshal([]byte(out), &p) != nil ||
		!strings.HasPrefix(p.Profile, "[Interface]") {
		t.Fatalf("phone --json: %d %s", code, out)
	}
	if code, out, _ = runArgs("node", "list", "--data", data); code != exitOK || !strings.Contains(out, "profile current") {
		t.Fatalf("list: %d\n%s", code, out)
	}
	if code, out, _ = runArgs("node", "show", "phone", "--data", data); code != exitOK || !strings.Contains(out, "last handshake") {
		t.Fatalf("show: %d\n%s", code, out)
	}
	if code, _, errOut = runArgs("node", "phone", "vps-1", "--data", data); code != exitFailure || !strings.Contains(errOut, "only phones") {
		t.Fatalf("phone for a server: %d %s", code, errOut)
	}
	if code, _, errOut = runArgs("node", "phone", "phone", "--exit", "vps-1", "--data", data); code != exitFailure || !strings.Contains(errOut, "can be an exit") {
		t.Fatalf("via a server that is no exit: %d %s", code, errOut)
	}
}

func TestNodeCommandsExplainAMissingServer(t *testing.T) {
	code, _, errOut := runArgs("node", "list", "--data", shortDir(t))
	if code != exitFailure || !strings.Contains(errOut, "meshctl control") {
		t.Fatalf("%d %s", code, errOut)
	}
}

func TestExportCommand(t *testing.T) {
	data, _ := startAdmin(t)
	if code, _, errOut := runArgs("node", "add", "vps-1", "--role", "server", "--data", data); code != exitOK {
		t.Fatal(errOut)
	}
	code, out, errOut := runArgs("export", "--data", data)
	if code != exitOK || !strings.Contains(out, "format: meshctl-export/1") || !strings.Contains(out, "name: vps-1") {
		t.Fatalf("export: %d\n%s%s", code, out, errOut)
	}
	file := filepath.Join(t.TempDir(), "mesh.yaml")
	if err := os.WriteFile(file, []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}
	if code, _, errOut = runArgs("export", "--data", data, "--out", file); code != exitOK {
		t.Fatalf("export --out: %d %s", code, errOut)
	}
	b, err := os.ReadFile(file)
	if err != nil || !strings.Contains(string(b), "vps-1") {
		t.Fatalf("export file %q, %v", b, err)
	}
	if st, err := os.Stat(file); err == nil && runtime.GOOS != "windows" && st.Mode().Perm() != 0o600 {
		t.Fatalf("export file mode %v, want 0600", st.Mode().Perm())
	}
	if code, _, _ = runArgs("export", "--data", shortDir(t)); code != exitFailure {
		t.Fatalf("export without a database: %d", code)
	}
}

func TestControlFlagErrors(t *testing.T) {
	data := shortDir(t)
	if code, _, errOut := runArgs("control", "--data", data); code != exitUsage || !strings.Contains(errOut, "--url") {
		t.Fatalf("no url: %d %s", code, errOut)
	}
	if code, _, errOut := runArgs("control", "--data", data, "--url", "https://mesh.example.com", "--trusted-proxies", "10.0.0.1/8"); code != exitUsage || !strings.Contains(errOut, "CIDR") {
		t.Fatalf("bad proxies: %d %s", code, errOut)
	}
	if code, _, errOut := runArgs("control", "--data", data, "--url", "https://mesh.example.com", "--log-level", "loud"); code != exitUsage {
		t.Fatalf("bad log level: %d %s", code, errOut)
	}
	for _, bad := range []string{"500ms", "2h", "-1m", "soon"} {
		if code, _, errOut := runArgs("control", "--data", data, "--url", "https://mesh.example.com", "--hub-down-after", bad); code != exitUsage ||
			!strings.Contains(errOut, "hub-down-after") {
			t.Errorf("--hub-down-after %s: %d %s", bad, code, errOut)
		}
	}
}

func TestAgentJoinTakesTheTokenFromTheEnvironment(t *testing.T) {
	var got proto.JoinRequest
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(b, &got)
		_ = json.NewEncoder(w).Encode(proto.JoinResponse{NodeID: 7, Name: "vps-1", Role: proto.RoleServer, MeshIP: "10.77.0.1",
			ConfigKey: testfix.ConfigKeyB64()})
	}))
	defer srv.Close()

	if code, _, errOut := runArgs("agent", "join", "--server", srv.URL, "--data", t.TempDir()); code != exitUsage || !strings.Contains(errOut, "MESHCTL_TOKEN") {
		t.Fatalf("no token: %d %s", code, errOut)
	}
	e, out, errb := testEnv(map[string]string{"MESHCTL_TOKEN": testfix.Token})
	code := run(context.Background(), e, []string{"agent", "join", "--server", srv.URL, "--data", t.TempDir(),
		"--endpoint6", "[2001:db8::10]:51820"})
	if code != exitOK || got.Token != testfix.Token || !strings.Contains(out.String(), "joined") || got.Endpoint6 != "[2001:db8::10]:51820" {
		t.Fatalf("join: %d %q\n%s%s", code, got.Token, out, errb)
	}
	if strings.Contains(out.String()+errb.String(), testfix.Token) {
		t.Fatal("the token was echoed")
	}
}

// `agent run` on a node that never joined says what to run first, and a bad
// log level is a usage error; neither touches the host.
func TestAgentRunNeedsAJoin(t *testing.T) {
	data := filepath.Join(shortDir(t), "data")
	sock := filepath.Join(shortDir(t), "agent.sock")
	if code, _, errOut := runArgs("agent", "run", "--data", data, "--socket", sock); code != exitFailure || !strings.Contains(errOut, "meshctl agent join") {
		t.Fatalf("agent run before a join = %d %q", code, errOut)
	}
	if code, _, errOut := runArgs("agent", "run", "--data", data, "--log-level", "loud"); code != exitUsage || !strings.Contains(errOut, "--log-level") {
		t.Fatalf("agent run --log-level loud = %d %q", code, errOut)
	}
	// The machine id file (amendment A50) is a path, checked before use.
	for _, args := range [][]string{{"agent", "run", "--data", data}, {"agent", "join", "--server", "https://mesh.example.com", "--token", testfix.Token,
		"--data", data}} {
		if code, _, errOut := runArgs(append(args, "--machine-id-file", "/etc/machine-id?x")...); code != exitUsage || !strings.Contains(errOut, "--machine-id-file") {
			t.Errorf("%v --machine-id-file with '?' = %d %q", args[:2], code, errOut)
		}
	}
}

// `agent reset` needs a data dir with a node in it; it refuses before it
// touches the host. (Its teardown is covered with fakes in the agent
// package: here it would take apart this host's real tunnel.)
func TestAgentReset(t *testing.T) {
	data := filepath.Join(shortDir(t), "data")
	if code, _, errOut := runArgs("agent", "reset", "--data", data); code != exitFailure || !strings.Contains(errOut, "nothing to reset") {
		t.Fatalf("agent reset of a data dir without a node = %d %q", code, errOut)
	}
	if code, _, _ := runArgs("agent", "reset", "--data", data, "now"); code != exitUsage {
		t.Fatalf("agent reset now = %d", code)
	}
}

// `agent install` and `agent uninstall` are gone (amendment A51): the
// install scripts set up the service.
func TestNoInstallCommands(t *testing.T) {
	for _, cmd := range []string{"install", "uninstall"} {
		if code, _, _ := runArgs("agent", cmd); code != exitUsage {
			t.Errorf("agent %s = %d", cmd, code)
		}
	}
}

// Every command that reaches the local agent refuses a bad --mesh before it
// dials anything.
func TestLocalCommandsCheckTheMesh(t *testing.T) {
	for _, args := range [][]string{
		{"status"}, {"exit"}, {"dns"}, {"logs"}, {"ui"}, {"healthcheck", "agent"}, {"agent", "run"},
	} {
		code, _, stderr := runArgs(append(args, "--mesh", "lab;reboot")...)
		if code != exitUsage || !strings.Contains(stderr, "--mesh") {
			t.Errorf("%v --mesh lab;reboot = %d, %q", args, code, stderr)
		}
	}
}

// Each mesh's agent has its data dir under the agent root (amendment A51):
// --data wins, then --mesh, then the one mesh that has joined; none, or
// several, says what to do. join names the mesh's dir, default for none,
// and sends the mesh only when it chose the dir.
func TestMeshFlags(t *testing.T) {
	e, _, _ := testEnv(nil)
	root := agentRoot(e)
	parse := func(args ...string) (string, error) {
		fs := flag.NewFlagSet("t", flag.ContinueOnError)
		data := meshFlags(fs, e)
		if err := fs.Parse(args); err != nil {
			return "", err
		}
		return data()
	}
	if dir, err := parse("--mesh", "lab"); err != nil || dir != meshDir(e, root, "lab") {
		t.Fatalf("--mesh lab = %q, %v", dir, err)
	}
	if dir, err := parse("--mesh", "lab", "--data", "/srv/x"); err != nil || dir != "/srv/x" {
		t.Fatalf("--data wins: %q, %v", dir, err)
	}
	if _, err := parse("--mesh", "Lab;x"); err == nil || !strings.Contains(err.Error(), "--mesh") {
		t.Fatalf("a bad mesh name = %v", err)
	}

	tmp := t.TempDir()
	if _, err := onlyMesh(e, tmp); err == nil || !strings.Contains(err.Error(), "agent join --mesh") {
		t.Fatalf("no mesh = %v", err)
	}
	for _, m := range []string{"lab", "default", "empty"} {
		if err := os.MkdirAll(filepath.Join(tmp, m), 0o700); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(tmp, "lab", state.EnrollmentFile), []byte("{}"), 0o600); err != nil {
		t.Fatal(err)
	}
	if dir, err := onlyMesh(e, tmp); err != nil || dir != meshDir(e, tmp, "lab") {
		t.Fatalf("the one mesh = %q, %v", dir, err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "default", state.EnrollmentFile), []byte("{}"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := onlyMesh(e, tmp); err == nil || !strings.Contains(err.Error(), "default, lab") {
		t.Fatalf("several meshes = %v", err)
	}

	join := func(args ...string) (agent.JoinOptions, error) {
		fs := flag.NewFlagSet("t", flag.ContinueOnError)
		opts := joinFlags(fs, e)
		if err := fs.Parse(append([]string{"--server", "https://mesh.example.com", "--token", testfix.Token}, args...)); err != nil {
			return agent.JoinOptions{}, err
		}
		return opts()
	}
	for _, c := range []struct {
		args       []string
		dir, mesh  string
		wantErrMsg string
	}{
		{nil, meshDir(e, root, "default"), "default", ""},
		{[]string{"--mesh", "lab"}, meshDir(e, root, "lab"), "lab", ""},
		{[]string{"--data", "/srv/x"}, "/srv/x", "", ""},
		{[]string{"--mesh", "-x"}, "", "", "--mesh"},
	} {
		o, err := join(c.args...)
		switch {
		case c.wantErrMsg != "" && (err == nil || !strings.Contains(err.Error(), c.wantErrMsg)):
			t.Errorf("join %v = %v", c.args, err)
		case c.wantErrMsg == "" && (err != nil || o.DataDir != c.dir || o.Mesh != c.mesh):
			t.Errorf("join %v = %q in %q, %v", c.args, o.Mesh, o.DataDir, err)
		}
	}
	if w := meshDir(&env{goos: "windows"}, `C:\ProgramData\meshctl`, "lab"); w != `C:\ProgramData\meshctl\lab` {
		t.Fatal(w)
	}
}

// A mesh whose data dir lives elsewhere (install.sh --data, amendment A54) is
// found through its link under the agent root; a link to a directory with no
// enrollment is no mesh.
func TestOnlyMeshFollowsALink(t *testing.T) {
	e, _, _ := testEnv(nil)
	root, elsewhere := t.TempDir(), t.TempDir()
	if err := os.WriteFile(filepath.Join(elsewhere, state.EnrollmentFile), []byte("{}"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(elsewhere, filepath.Join(root, "nas")); err != nil {
		t.Skipf("no symbolic links here: %v", err)
	}
	if err := os.Symlink(t.TempDir(), filepath.Join(root, "empty")); err != nil {
		t.Fatal(err)
	}
	if dir, err := onlyMesh(e, root); err != nil || dir != meshDir(e, root, "nas") {
		t.Fatalf("the linked mesh = %q, %v", dir, err)
	}
}

// The agent logs to stderr, or as the Windows service to agent.log in its
// data dir; a data dir that cannot be opened is the error.
func TestAgentLog(t *testing.T) {
	var stderr strings.Builder
	out, done, err := agentLog(&stderr, "unused", false)
	if err != nil || out != &stderr {
		t.Fatalf("agentLog outside a service = %v, %v", out, err)
	}
	done()
	data := filepath.Join(shortDir(t), "data")
	out, done, err = agentLog(&stderr, data, true)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := io.WriteString(out, "level=INFO msg=\"meshctl agent running\"\n"); err != nil {
		t.Fatal(err)
	}
	done()
	if b, err := os.ReadFile(filepath.Join(data, "agent.log")); err != nil || !strings.Contains(string(b), "agent running") || stderr.Len() != 0 {
		t.Fatalf("agent.log = %q, %v; stderr %q", b, err, stderr.String())
	}
	if _, _, err := agentLog(&stderr, filepath.Join(shortDir(t), "data?x"), true); err == nil || !strings.Contains(err.Error(), "--data") {
		t.Fatalf("agentLog with an unusable data dir = %v", err)
	}
	blocked := filepath.Join(shortDir(t), "blocked")
	if err := os.MkdirAll(filepath.Join(blocked, "agent.log", "x"), 0o700); err != nil {
		t.Fatal(err)
	}
	if _, _, err := agentLog(&stderr, blocked, true); err == nil || !strings.Contains(err.Error(), "open log file") {
		t.Fatalf("agentLog over a directory = %v", err)
	}
}

func TestHealthcheckControl(t *testing.T) {
	status := http.StatusOK
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != proto.PathHealth {
			http.NotFound(w, r)
			return
		}
		w.WriteHeader(status)
	}))
	defer srv.Close()
	addr := strings.TrimPrefix(srv.URL, "http://")
	if code, _, errOut := runArgs("healthcheck", "control", "--addr", addr); code != exitOK {
		t.Fatalf("healthy: %d %s", code, errOut)
	}
	status = http.StatusServiceUnavailable
	if code, _, _ := runArgs("healthcheck", "control", "--addr", addr); code != exitFailure {
		t.Fatalf("unhealthy: %d", code)
	}
	if code, _, _ := runArgs("healthcheck", "control", "--addr", "no-port"); code != exitUsage {
		t.Fatalf("bad addr: %d", code)
	}
}

func TestStatusAndAgentHealthcheck(t *testing.T) {
	sock := filepath.Join(shortDir(t), "agent.sock")
	ln, err := localapi.Listen(sock)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	st := proto.LocalStatus{
		Node:           proto.Self{ID: 1, Name: "vps-1", Role: proto.RoleServer},
		MeshIP:         "10.77.0.1",
		AgentVersion:   "v1.0.0",
		WGBackend:      "kernel",
		AppliedVersion: 4,
		AppliedHash:    strings.Repeat("a", 64),
		AppliedAt:      testfix.Epoch.Unix() - 30,
		Peers: []proto.PeerStatus{
			{Name: "vps-2", Endpoint: "203.0.113.11:51820", LastHandshake: testfix.Epoch.Unix() - 5, RxBytes: 2048, TxBytes: 10},
		},
		Degraded: []string{"firewall"},
		Rollback: &proto.Rollback{From: 5, To: 4, Reason: "no answer from the server within 1m0s of applying v5"},
	}
	go func() {
		_ = localapi.Serve(ctx, ln, localapi.Handler(localapi.Funcs{Status: func(context.Context) (proto.LocalStatus, error) { return st, nil }}))
	}()

	code, out, errOut := runArgs("status", "--socket", sock)
	for _, want := range []string{"vps-1 (server)", "config    v4, applied 30s ago", "degraded  firewall", "vps-2", "5s ago", "2.0 KiB",
		"rollback  v5 -> v4: no answer"} {
		if !strings.Contains(out, want) {
			t.Errorf("status lacks %q:\n%s%s", want, out, errOut)
		}
	}
	if code != exitOK {
		t.Fatalf("status: %d", code)
	}
	if code, out, _ = runArgs("status", "--socket", sock, "--json"); code != exitOK || !strings.Contains(out, `"applied_version": 4`) {
		t.Fatalf("status --json: %d %s", code, out)
	}
	if code, _, _ = runArgs("healthcheck", "agent", "--socket", sock); code != exitOK {
		t.Fatalf("healthcheck agent: %d", code)
	}
	if code, _, _ = runArgs("healthcheck", "agent", "--socket", filepath.Join(shortDir(t), "none.sock")); code != exitFailure {
		t.Fatalf("healthcheck agent without an agent: %d", code)
	}
}

// A laptop that has applied nothing yet, with an exit chosen, an error to
// show, and later a peer whose clock runs ahead.
func TestPrintStatusOfALaptop(t *testing.T) {
	st := proto.LocalStatus{Node: proto.Self{Name: "laptop", Role: proto.RoleSpoke}, MeshIP: "10.77.2.1", Exit: "vps-2",
		Errors: []string{"apply step exit: operation not permitted"}, DNS: proto.DNSOn}
	var b strings.Builder
	printStatus(&b, st, testfix.Epoch)
	for _, want := range []string{"config    none applied yet", "exit      via vps-2", "error     apply step exit",
		"dns       on: this host sends .mesh lookups to the resolver on 10.77.2.1"} {
		if !strings.Contains(b.String(), want) {
			t.Errorf("status lacks %q:\n%s", want, b.String())
		}
	}
	if strings.Contains(b.String(), "PEER") {
		t.Errorf("a peer table without peers:\n%s", b.String())
	}

	st.Peers = []proto.PeerStatus{{Name: "vps-1", LastHandshake: testfix.Epoch.Unix() + 5}}
	b.Reset()
	printStatus(&b, st, testfix.Epoch)
	lines := strings.Split(strings.TrimSpace(b.String()), "\n")
	if f := strings.Fields(lines[len(lines)-1]); len(f) < 4 || f[0] != "vps-1" || f[1] != "-" || f[2]+" "+f[3] != "0s ago" {
		t.Fatalf("peer row %q: no endpoint is -, a handshake in the future is 0s ago", lines[len(lines)-1])
	}

	if code, _, _ := runArgs("status", "extra"); code != exitUsage {
		t.Errorf("status with an argument = %d", code)
	}
	if code, _, errOut := runArgs("status", "--socket", filepath.Join(shortDir(t), "none.sock")); code != exitFailure ||
		!strings.Contains(errOut, "reach the agent at") {
		t.Errorf("status without an agent = %d %q", code, errOut)
	}
}

func TestBytesize(t *testing.T) {
	for n, want := range map[int64]string{0: "0 B", 1023: "1023 B", 1024: "1.0 KiB", 5 << 20: "5.0 MiB"} {
		if got := bytesize(n); got != want {
			t.Errorf("bytesize(%d) = %q, want %q", n, got, want)
		}
	}
}

func TestNodeShowAnEnrolledServer(t *testing.T) {
	const rollback = `{"rollback":{"from":5,"to":4,"reason":"no answer from the server"}}`
	data, st := startAdmin(t)
	err := st.MeshTx(context.Background(), store.FirstMesh, func(tx *store.MeshTx) error {
		ids := map[string]int64{}
		for _, f := range []testfix.Node{testfix.VPS1, testfix.VPS2} {
			n := store.Node{Name: f.Name, Role: "server", MeshIP: f.MeshIP, Endpoint: f.Endpoint, ListenPort: 51820, Keepalive: 25,
				MTU: 1420, Enrolled: true, WGPubKey: testfix.WGKey(f.Key), IdentityPubKey: testfix.IdentityKeyB64(f.Key)}
			if err := tx.InsertNode(&n); err != nil {
				return err
			}
			ids[f.Name] = n.ID
		}
		return tx.RecordStatus(ids["vps-1"], store.StatusUpdate{AgentVersion: "v1.0.0", WGBackend: "kernel", StatusJSON: rollback,
			Peers: []store.PeerStat{{PeerID: ids["vps-2"], LastHandshake: testfix.Epoch.Unix() - 3, RxBytes: 4096}}})
	})
	if err != nil {
		t.Fatal(err)
	}
	code, out, errOut := runArgs("node", "show", "vps-1", "--data", data)
	for _, want := range []string{"listen port", "51820", "can exit", "kernel", "peers as last reported:", "vps-2", "3s ago", "4.0 KiB",
		"computed config:", `"name": "vps-1"`, "v5 -> v4: no answer from the server", "rolled back"} {
		if !strings.Contains(out, want) {
			t.Errorf("node show lacks %q:\n%s%s", want, out, errOut)
		}
	}
	if code != exitOK {
		t.Fatalf("node show: %d", code)
	}
	if code, out, _ = runArgs("node", "list", "--data", data); code != exitOK || !strings.Contains(out, "joined, no heartbeat") {
		t.Fatalf("node list: %d\n%s", code, out)
	}
}

// passwords answers password prompts from a list, as a person at a terminal
// would.
func passwords(answers ...string) func(string) (string, error) {
	return func(string) (string, error) {
		if len(answers) == 0 {
			return "", errors.New("no more answers")
		}
		a := answers[0]
		answers = answers[1:]
		return a, nil
	}
}

func runWith(e *env, args ...string) (int, string, string) {
	out, errb := e.stdout.(*bytes.Buffer), e.stderr.(*bytes.Buffer)
	out.Reset()
	errb.Reset()
	code := run(context.Background(), e, args)
	return code, out.String(), errb.String()
}

func TestUserCommands(t *testing.T) {
	data := shortDir(t)
	e, _, _ := testEnv(nil)
	if code, _, errOut := runWith(e, "user", "add", "admin", "--data", data); code != exitFailure || !strings.Contains(errOut, "terminal") {
		t.Fatalf("without a terminal: %d %s", code, errOut)
	}
	e.password = passwords("correct horse battery", "correct horse batterz")
	if code, _, errOut := runWith(e, "user", "add", "admin", "--data", data); code != exitFailure || !strings.Contains(errOut, "differ") {
		t.Fatalf("mismatched passwords: %d %s", code, errOut)
	}
	e.password = passwords("short")
	if code, _, errOut := runWith(e, "user", "add", "admin", "--data", data); code != exitFailure || !strings.Contains(errOut, "at least 12") {
		t.Fatalf("short password: %d %s", code, errOut)
	}
	e.password = passwords("correct horse battery", "correct horse battery")
	if code, out, errOut := runWith(e, "user", "add", "admin", "--data", data); code != exitOK || !strings.Contains(out, "added user admin") {
		t.Fatalf("add: %d %s%s", code, out, errOut)
	}
	e.password = passwords("correct horse battery", "correct horse battery")
	if code, _, errOut := runWith(e, "user", "add", "admin", "--data", data); code != exitFailure || !strings.Contains(errOut, "already exists") {
		t.Fatalf("duplicate: %d %s", code, errOut)
	}
	e.password = passwords("another horse battery", "another horse battery")
	if code, out, _ := runWith(e, "user", "passwd", "admin", "--data", data); code != exitOK || !strings.Contains(out, "ended their sessions") {
		t.Fatalf("passwd: %d %s", code, out)
	}
	if code, out, _ := runWith(e, "user", "reset-totp", "admin", "--data", data); code != exitOK || !strings.Contains(out, "removed the authenticator") {
		t.Fatalf("reset-totp: %d %s", code, out)
	}
	if code, _, errOut := runWith(e, "user", "reset-totp", "ghost", "--data", data); code != exitFailure || !strings.Contains(errOut, "does not exist") {
		t.Fatalf("unknown user: %d %s", code, errOut)
	}
	if code, out, errOut := runWith(e, "user", "add", "admin", "--data", data); strings.Contains(out+errOut, "horse") {
		t.Fatalf("a password was echoed: %d %s%s", code, out, errOut)
	}
}

// user add --admin|--mesh, grant, revoke and list (amendment A49): the first
// user is an admin (never one for a single mesh), the last admin stays one,
// and grant and revoke take a mesh or the admin role, not both.
func TestUserAccessCommands(t *testing.T) {
	data := shortDir(t)
	e, _, _ := testEnv(nil)
	run := func(want int, args ...string) string {
		t.Helper()
		e.password = passwords("correct horse battery", "correct horse battery")
		code, out, errOut := runWith(e, append(args, "--data", data)...)
		if code != want {
			t.Fatalf("meshctl %s = %d, want %d\n%s%s", strings.Join(args, " "), code, want, out, errOut)
		}
		return out + errOut
	}
	if out := run(exitOK, "user", "add", "root"); !strings.Contains(out, "added user root as an admin") {
		t.Fatalf("the first user:\n%s", out)
	}
	if out := run(exitOK, "user", "add", "ops"); !strings.Contains(out, "meshctl user grant ops --mesh <mesh>") {
		t.Fatalf("a second user:\n%s", out)
	}
	if out := run(exitOK, "user", "add", "boss", "--admin"); !strings.Contains(out, "added user boss as an admin") {
		t.Fatalf("an admin:\n%s", out)
	}
	if out := run(exitOK, "user", "grant", "ops", "--mesh", "default"); !strings.Contains(out, "ops manages mesh default") {
		t.Fatalf("grant:\n%s", out)
	}
	for _, c := range []struct {
		want int
		args []string
		says string
	}{
		{exitFailure, []string{"user", "grant", "ops", "--mesh", "none"}, "no such mesh"},
		{exitFailure, []string{"user", "grant", "ghost", "--mesh", "default"}, "does not exist"},
		{exitUsage, []string{"user", "grant", "ops"}, "--mesh <mesh> or --admin"},
		{exitUsage, []string{"user", "revoke", "ops", "--mesh", "default", "--admin"}, "--mesh <mesh> or --admin"},
	} {
		if out := run(c.want, c.args...); !strings.Contains(out, c.says) {
			t.Errorf("meshctl %s:\n%s", strings.Join(c.args, " "), out)
		}
	}
	out := run(exitOK, "user", "list")
	for _, want := range []string{"NAME  MANAGES", "boss  everything (admin)", "ops   default", "root  everything (admin)", "at next sign-in"} {
		if !strings.Contains(out, want) {
			t.Fatalf("user list lacks %q:\n%s", want, out)
		}
	}
	if out := run(exitOK, "user", "revoke", "ops", "--mesh", "default"); !strings.Contains(out, "ops no longer manages mesh default") {
		t.Fatalf("revoke:\n%s", out)
	}
	if out := run(exitOK, "user", "revoke", "root", "--admin"); !strings.Contains(out, "root is no longer an admin") {
		t.Fatalf("revoke --admin:\n%s", out)
	}
	if out := run(exitFailure, "user", "revoke", "boss", "--admin"); !strings.Contains(out, "at least one admin") {
		t.Fatalf("the last admin:\n%s", out)
	}
	if out := run(exitOK, "user", "grant", "ops", "--admin"); !strings.Contains(out, "ops is now an admin") {
		t.Fatalf("grant --admin:\n%s", out)
	}
	if out := run(exitOK, "user", "list"); !strings.Contains(out, "ops   everything (admin)") || !strings.Contains(out, "root  -") {
		t.Fatalf("user list after the changes:\n%s", out)
	}
	if out := run(exitUsage, "user", "add", "x", "--admin", "--mesh", "default"); !strings.Contains(out, "not both") {
		t.Fatalf("an admin for one mesh:\n%s", out)
	}

	if out := run(exitOK, "user", "add", "lab", "--mesh", "default"); !strings.Contains(out, "added user lab, who manages mesh default") {
		t.Fatalf("a user for one mesh:\n%s", out)
	}
	if out := run(exitOK, "user", "list"); !strings.Contains(out, "lab   default") {
		t.Fatalf("user list:\n%s", out)
	}

	// The first user is an admin, never one for a single mesh.
	data = shortDir(t)
	if out := run(exitFailure, "user", "add", "lab", "--mesh", "default"); !strings.Contains(out, "first user becomes one") {
		t.Fatalf("a first user for one mesh:\n%s", out)
	}
	if out := run(exitOK, "user", "list"); strings.Contains(out, "lab") {
		t.Fatalf("user list after the refusal:\n%s", out)
	}
}

func TestImportCommand(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("import takes the Linux-only data dir lock")
	}
	doc := export.Document{Format: export.Format, Settings: map[string]string{}, Nodes: []export.Node{{
		ID: 1, Name: "vps-1", Role: "server", MeshIP: testfix.VPS1.MeshIP, Endpoint: testfix.VPS1.Endpoint, ListenPort: 51820,
		Keepalive: 25, MTU: 1420, Enrolled: true, WGPublicKey: testfix.WGKey(1), IdentityPublicKey: testfix.IdentityKeyB64(1)}}}
	var b bytes.Buffer
	if err := export.Write(&b, doc); err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(t.TempDir(), "mesh.yaml")
	if err := os.WriteFile(file, b.Bytes(), 0o600); err != nil {
		t.Fatal(err)
	}
	data := shortDir(t)
	code, out, errOut := runArgs("import", file, "--data", data)
	if code != exitOK || !strings.Contains(out, "imported 1 nodes") || !strings.Contains(out, "config v1;") {
		t.Fatalf("import: %d %s%s", code, out, errOut)
	}
	if code, out, _ = runArgs("export", "--data", data); code != exitOK || !strings.Contains(out, "name: vps-1") {
		t.Fatalf("export after import: %d %s", code, out)
	}
	if err := os.WriteFile(file, []byte("format: meshctl-export/1\nusers: []\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if code, _, errOut = runArgs("import", file, "--data", data); code != exitFailure || !strings.Contains(errOut, "users") {
		t.Fatalf("a file with unknown fields: %d %s", code, errOut)
	}
	if code, _, _ = runArgs("import", "--data", data); code != exitUsage {
		t.Fatalf("no file: %d", code)
	}
}

// A join token's expiry is shown in the local time zone.
func TestPrintEnrollmentInTheLocalZone(t *testing.T) {
	var b strings.Builder
	printEnrollment(&b, time.FixedZone("SGT", 8*3600), &api.Enrollment{Token: testfix.Token, ExpiresAt: testfix.Epoch.Unix()})
	if !strings.Contains(b.String(), "valid until 2026-10-12T16:00:00+08:00") {
		t.Fatalf("printEnrollment =\n%s", b.String())
	}
	b.Reset()
	if printEnrollment(&b, time.UTC, nil); b.Len() != 0 {
		t.Fatalf("no enrollment printed %q", b.String())
	}
}
