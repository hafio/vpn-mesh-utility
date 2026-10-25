package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/hafio/vpn-mesh-utility/internal/server/api"
	"github.com/hafio/vpn-mesh-utility/internal/server/export"
	"github.com/hafio/vpn-mesh-utility/internal/testfix"
)

// The mesh commands add, list, show, change and remove meshes, and --mesh
// picks the mesh of every other admin command and of export (amendment
// A49).
func TestMeshCommands(t *testing.T) {
	data, _ := startAdmin(t)
	run := func(want int, args ...string) string {
		t.Helper()
		code, out, errOut := runArgs(append(args, "--data", data)...)
		if code != want {
			t.Fatalf("meshctl %s: %d, want %d\n%s%s", strings.Join(args, " "), code, want, out, errOut)
		}
		return out + errOut
	}
	if out := run(exitOK, "mesh", "add", "lab"); !strings.Contains(out, "added mesh lab in slot 1: 10.78.0.0/16, fd") ||
		!strings.Contains(out, "zone lab") || !strings.Contains(out, "node add --mesh lab") {
		t.Fatalf("mesh add lab:\n%s", out)
	}
	if out := run(exitOK, "mesh", "add", "home", "--mesh-cidr", "10.90.0.0/20", "--dns-zone", "home.arpa", "--ipv6", "off"); !strings.Contains(out,
		"slot 2: 10.90.0.0/20, no IPv6, zone home.arpa") {
		t.Fatalf("mesh add home:\n%s", out)
	}
	if out := run(exitFailure, "mesh", "add", "lab"); !strings.Contains(out, "exists already") {
		t.Fatalf("a second lab:\n%s", out)
	}
	if out := run(exitUsage, "mesh", "add"); !strings.Contains(out, "<name>") {
		t.Fatalf("mesh add without a name:\n%s", out)
	}
	out := run(exitOK, "mesh", "list")
	for _, want := range []string{"NAME", "default  0     0", "lab      1     0      10.78.0.0/16", "home     2     0      10.90.0.0/20  -"} {
		if !strings.Contains(out, want) {
			t.Fatalf("mesh list lacks %q:\n%s", want, out)
		}
	}
	var l api.MeshList
	if err := json.Unmarshal([]byte(run(exitOK, "mesh", "list", "--json")), &l); err != nil || len(l.Meshes) != 3 {
		t.Fatalf("mesh list --json: %+v, %v", l, err)
	}

	// With three meshes a node command names one.
	if out := run(exitFailure, "node", "list"); !strings.Contains(out, "3 meshes (default, lab, home): name one with --mesh") {
		t.Fatalf("node list without --mesh:\n%s", out)
	}
	run(exitOK, "node", "add", "nas", "--role", "spoke", "--mesh", "lab")
	if out := run(exitOK, "node", "list", "--mesh", "lab"); !strings.Contains(out, "nas") {
		t.Fatalf("the lab's nodes:\n%s", out)
	}
	if out := run(exitFailure, "node", "list", "--mesh", "none"); !strings.Contains(out, `no mesh "none"`) {
		t.Fatalf("an unknown mesh:\n%s", out)
	}
	if out := run(exitOK, "export", "--mesh", "lab"); !strings.Contains(out, "name: nas") || !strings.Contains(out, "10.78.0.0/16") {
		t.Fatalf("export --mesh lab:\n%s", out)
	}
	if out := run(exitFailure, "export"); !strings.Contains(out, "--mesh") {
		t.Fatalf("export without --mesh:\n%s", out)
	}
	if out := run(exitUsage, "export", "--mesh", "Lab!"); !strings.Contains(out, "--mesh") {
		t.Fatalf("export --mesh Lab!:\n%s", out)
	}

	out = run(exitOK, "mesh", "show", "lab")
	for _, want := range []string{"uuid", "slot           1", "nodes          1", "ipv4 network   10.78.0.0/16", "dns zone       lab",
		"extra domains  none", "upstream dns   1.1.1.1, 9.9.9.9", "default mtu    1420", "default hub    none"} {
		if !strings.Contains(out, want) {
			t.Fatalf("mesh show lacks %q:\n%s", want, out)
		}
	}
	var m api.MeshInfo
	if err := json.Unmarshal([]byte(run(exitOK, "mesh", "show", "lab", "--json")), &m); err != nil || m.Slot != 1 || m.Nodes != 1 {
		t.Fatalf("mesh show --json: %+v, %v", m, err)
	}
	if out := run(exitOK, "mesh", "set", "home", "--name", "house", "--dns-zone", "house", "--mesh-cidr", "10.91.0.0/20",
		"--dns-extra-domains", "files.home", "--upstream-dns", "9.9.9.9", "--default-mtu", "1380", "--default-hub", ""); !strings.Contains(out,
		"updated mesh house; no node's config changed") {
		t.Fatalf("mesh set:\n%s", out)
	}
	out = run(exitOK, "mesh", "show", "house")
	for _, want := range []string{"10.91.0.0/20", "dns zone       house", "files.home", "upstream dns   9.9.9.9", "1380"} {
		if !strings.Contains(out, want) {
			t.Fatalf("mesh show after set lacks %q:\n%s", want, out)
		}
	}
	var u api.MeshUpdate
	if err := json.Unmarshal([]byte(run(exitOK, "mesh", "set", "house", "--default-mtu", "1400", "--json")), &u); err != nil || u.Mesh.DefaultMTU != 1400 {
		t.Fatalf("mesh set --json: %+v, %v", u, err)
	}
	if out := run(exitUsage, "mesh", "set", "house"); !strings.Contains(out, "nothing to change") {
		t.Fatalf("mesh set without a change:\n%s", out)
	}
	if out := run(exitFailure, "mesh", "set", "house", "--dns-zone", "lab"); !strings.Contains(out, "mesh lab's lab") {
		t.Fatalf("a zone another mesh has:\n%s", out)
	}

	if out := run(exitFailure, "mesh", "rm", "lab"); !strings.Contains(out, "still has 1 node(s) (nas)") {
		t.Fatalf("mesh rm with a node:\n%s", out)
	}
	run(exitOK, "node", "rm", "nas", "--mesh", "lab")
	if out := run(exitOK, "mesh", "rm", "lab"); !strings.Contains(out, "removed mesh lab") {
		t.Fatalf("mesh rm:\n%s", out)
	}
	if out := run(exitOK, "mesh", "rm", "house", "--json"); !strings.Contains(out, `"removed": "house"`) {
		t.Fatalf("mesh rm --json:\n%s", out)
	}
	if out := run(exitOK, "node", "list"); !strings.Contains(out, "NAME") {
		t.Fatalf("node list with one mesh again:\n%s", out)
	}
}

// import --mesh replaces a named mesh, creating it when the control plane
// has none of that name; a failed import takes the new mesh away again.
func TestImportIntoAMesh(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("import takes the Linux-only data dir lock")
	}
	write := func(cidr, zone string, id int64, name, ip string, key byte) string {
		t.Helper()
		doc := export.Document{Format: export.Format, Settings: map[string]string{"mesh_cidr": cidr, "dns_zone": zone},
			Nodes: []export.Node{{ID: id, Name: name, Role: "server", MeshIP: ip, Endpoint: testfix.VPS1.Endpoint, ListenPort: 51820,
				Keepalive: 25, MTU: 1420, Enrolled: true, WGPublicKey: testfix.WGKey(key), IdentityPublicKey: testfix.IdentityKeyB64(key)}}}
		var b bytes.Buffer
		if err := export.Write(&b, doc); err != nil {
			t.Fatal(err)
		}
		file := filepath.Join(t.TempDir(), "mesh.yaml")
		if err := os.WriteFile(file, b.Bytes(), 0o600); err != nil {
			t.Fatal(err)
		}
		return file
	}
	data := shortDir(t)
	// The first mesh gets a vps-1 with id 1 (into the sole mesh).
	first := write("10.77.0.0/16", "mesh", 1, "vps-1", testfix.VPS1.MeshIP, 1)
	if code, out, errOut := runArgs("import", first, "--data", data); code != exitOK {
		t.Fatalf("import: %d %s%s", code, out, errOut)
	}
	// id 1 is the first mesh's: the lab is made, then goes again.
	clash := write("10.78.0.0/16", "lab", 1, "lab-1", "10.78.0.1", 2)
	if code, _, errOut := runArgs("import", clash, "--data", data, "--mesh", "lab"); code != exitFailure || !strings.Contains(errOut, "lab-1") {
		t.Fatalf("an import onto another mesh's id: %d %s", code, errOut)
	}
	if code, _, errOut := runArgs("export", "--data", data, "--mesh", "lab"); code != exitFailure || !strings.Contains(errOut, `no mesh "lab"`) {
		t.Fatalf("the failed import's mesh: %d %s", code, errOut)
	}
	lab := write("10.78.0.0/16", "lab", 9, "lab-1", "10.78.0.1", 2)
	if code, out, errOut := runArgs("import", lab, "--data", data, "--mesh", "lab"); code != exitOK || !strings.Contains(out, "imported 1 nodes") {
		t.Fatalf("import --mesh lab: %d %s%s", code, out, errOut)
	}
	if code, out, errOut := runArgs("export", "--data", data, "--mesh", "lab"); code != exitOK || !strings.Contains(out, "name: lab-1") ||
		!strings.Contains(out, "dns_zone: lab") {
		t.Fatalf("export --mesh lab: %d %s%s", code, out, errOut)
	}
	if code, out, _ := runArgs("export", "--data", data, "--mesh", "default"); code != exitOK || !strings.Contains(out, "name: vps-1") {
		t.Fatalf("export --mesh default: %d %s", code, out)
	}
	if code, _, errOut := runArgs("import", lab, "--data", data, "--mesh", "Lab!"); code != exitUsage || !strings.Contains(errOut, "--mesh") {
		t.Fatalf("import --mesh Lab!: %d %s", code, errOut)
	}
}
