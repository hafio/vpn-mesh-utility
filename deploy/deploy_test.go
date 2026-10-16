package deploy

import (
	"bufio"
	"bytes"
	"flag"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"go.yaml.in/yaml/v3"
)

var update = flag.Bool("update", false, "rewrite testdata goldens")

// The unit's [Service] section, exactly. Adding, removing or widening any
// setting fails here first; docs/security.md explains each one. There is no
// User= (the agent is root, bounded by the capability set) and no
// ProtectKernelTunables (it writes net.ipv4.ip_forward).
var wantService = map[string]string{
	"Type":                     "simple",
	"ExecStart":                "/usr/local/bin/meshctl agent run --data /var/lib/meshctl",
	"Restart":                  "always",
	"RestartSec":               "5",
	"StateDirectory":           "meshctl",
	"StateDirectoryMode":       "0700",
	"RuntimeDirectory":         "meshctl wireguard",
	"RuntimeDirectoryMode":     "0755",
	"RuntimeDirectoryPreserve": "yes",
	"UMask":                    "0077",
	"CapabilityBoundingSet":    "CAP_NET_ADMIN CAP_NET_BIND_SERVICE",
	"AmbientCapabilities":      "CAP_NET_ADMIN CAP_NET_BIND_SERVICE",
	"NoNewPrivileges":          "yes",
	"ProtectSystem":            "strict",
	"ReadWritePaths":           "/var/lib/meshctl",
	"ProtectHome":              "yes",
	"PrivateTmp":               "yes",
	"ProtectKernelModules":     "yes",
	"ProtectKernelLogs":        "yes",
	"ProtectControlGroups":     "yes",
	"ProtectClock":             "yes",
	"ProtectHostname":          "yes",
	"RestrictNamespaces":       "yes",
	"RestrictRealtime":         "yes",
	"RestrictSUIDSGID":         "yes",
	"LockPersonality":          "yes",
	"MemoryDenyWriteExecute":   "yes",
	"SystemCallArchitectures":  "native",
	"DevicePolicy":             "closed",
	"DeviceAllow":              "/dev/net/tun rw",
	"RestrictAddressFamilies":  "AF_UNIX AF_INET AF_INET6 AF_NETLINK",
}

func TestUnitPosture(t *testing.T) {
	got := map[string]string{}
	section := ""
	sc := bufio.NewScanner(bytes.NewReader(SystemdUnit()))
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		switch {
		case line == "" || strings.HasPrefix(line, "#"):
		case strings.HasPrefix(line, "["):
			section = line
		case section == "[Service]":
			k, v, ok := strings.Cut(line, "=")
			if !ok {
				t.Fatalf("malformed line %q", line)
			}
			if _, dup := got[k]; dup {
				t.Fatalf("%s is set twice", k)
			}
			got[k] = v
		}
	}
	if !reflect.DeepEqual(got, wantService) {
		for k, v := range got {
			if wantService[k] != v {
				t.Errorf("%s=%s, want %q", k, v, wantService[k])
			}
		}
		for k := range wantService {
			if _, ok := got[k]; !ok {
				t.Errorf("%s is missing", k)
			}
		}
	}
	if !strings.Contains(string(SystemdUnit()), "WantedBy=multi-user.target") {
		t.Error("the unit must be enabled for multi-user.target")
	}
}

func testParams() InstallParams {
	return InstallParams{Origin: "https://mesh.example.com", Version: "v1.2.3",
		SHA256Amd64: strings.Repeat("a", 64), SHA256Arm64: strings.Repeat("b", 64)}
}

func TestInstallGolden(t *testing.T) {
	got, err := RenderInstall(testParams())
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join("testdata", "install.golden.sh")
	if *update {
		if err := os.WriteFile(path, got, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("install.sh differs from %s; review the template change, then ./scripts/dev.sh regen", path)
	}
}

// The rendered script differs from the template only at the markers, each of
// which appears exactly once, inside single quotes.
func TestInstallChangesOnlyTheMarkers(t *testing.T) {
	p := testParams()
	got, err := RenderInstall(p)
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range []string{markerOrigin, markerVersion, markerSHA256Amd64, markerSHA256Arm64} {
		if n := strings.Count(installTemplate, "'"+m+"'"); n != 1 || strings.Count(installTemplate, m) != 1 {
			t.Errorf("marker %s must appear exactly once, quoted", m)
		}
	}
	back := strings.NewReplacer("'"+p.Origin+"'", "'"+markerOrigin+"'", "'"+p.Version+"'", "'"+markerVersion+"'",
		"'"+p.SHA256Amd64+"'", "'"+markerSHA256Amd64+"'", "'"+p.SHA256Arm64+"'", "'"+markerSHA256Arm64+"'").Replace(string(got))
	if back != installTemplate {
		t.Fatal("rendering changed more than the markers")
	}
}

func TestInstallParses(t *testing.T) {
	sh, err := exec.LookPath("sh")
	if err != nil {
		t.Skip("no sh on this runner")
	}
	got, err := RenderInstall(testParams())
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(sh, "-n")
	cmd.Stdin = bytes.NewReader(got)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("sh -n: %v\n%s", err, out)
	}
}

func TestInstallRefusesHostileValues(t *testing.T) {
	cases := map[string]func(*InstallParams){
		"quote in origin":         func(p *InstallParams) { p.Origin = "https://mesh.example.com'; id; '" },
		"origin with a path":      func(p *InstallParams) { p.Origin = "https://mesh.example.com/x" },
		"non-canonical origin":    func(p *InstallParams) { p.Origin = "https://Mesh.Example.com" },
		"plain http":              func(p *InstallParams) { p.Origin = "http://mesh.example.com" },
		"substitution in version": func(p *InstallParams) { p.Version = "v1$(id)" },
		"newline in version":      func(p *InstallParams) { p.Version = "v1\nid" },
		"short checksum":          func(p *InstallParams) { p.SHA256Amd64 = "abc" },
		"quote in checksum":       func(p *InstallParams) { p.SHA256Arm64 = strings.Repeat("a", 63) + "'" },
	}
	for name, edit := range cases {
		p := testParams()
		edit(&p)
		if out, err := RenderInstall(p); err == nil || out != nil {
			t.Errorf("%s: rendered", name)
		}
	}
}

// Every token that would widen a container's privileges.
var widening = []string{"cap_add", "devices", "network_mode", "pid:", "ipc:", "userns_mode", "privileged: true",
	"docker.sock", "seccomp", "apparmor", "SYS_ADMIN", "/dev/", "host:"}

func service(t *testing.T, file string) (map[string]any, map[string]any) {
	t.Helper()
	b, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	for _, w := range widening {
		if strings.Contains(string(b), w) {
			t.Errorf("%s contains the widening token %q", file, w)
		}
	}
	var doc map[string]any
	if err := yaml.Unmarshal(b, &doc); err != nil {
		t.Fatal(err)
	}
	svcs, _ := doc["services"].(map[string]any)
	svc, _ := svcs["meshctl"].(map[string]any)
	if len(svcs) != 1 || svc == nil {
		t.Fatalf("%s must define exactly the meshctl service", file)
	}
	return doc, svc
}

func keys(m map[string]any) []string {
	var out []string
	for k := range m {
		out = append(out, k)
	}
	return out
}

func checkPosture(t *testing.T, file string, svc map[string]any, want []string) {
	t.Helper()
	got := keys(svc)
	if !sameSet(got, want) {
		t.Errorf("%s service keys = %v, want exactly %v", file, got, want)
	}
	exact := map[string]any{
		"user":         "65532:65532",
		"privileged":   false,
		"cap_drop":     []any{"ALL"},
		"security_opt": []any{"no-new-privileges=true"},
		"read_only":    true,
	}
	for k, v := range exact {
		if !reflect.DeepEqual(svc[k], v) {
			t.Errorf("%s: %s = %#v, want %#v", file, k, svc[k], v)
		}
	}
	vols, _ := svc["volumes"].([]any)
	for _, v := range vols {
		if s, _ := v.(string); strings.HasPrefix(s, "/") || strings.HasPrefix(s, ".") {
			t.Errorf("%s mounts a host path %q; only named volumes are allowed", file, s)
		}
	}
}

func sameSet(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	seen := map[string]bool{}
	for _, s := range a {
		seen[s] = true
	}
	for _, s := range b {
		if !seen[s] {
			return false
		}
	}
	return true
}

func TestServerComposePosture(t *testing.T) {
	if !bytes.Equal(ComposeServer(), mustRead(t, "compose.server.yml")) {
		t.Fatal("the embedded compose file is stale")
	}
	doc, svc := service(t, "compose.server.yml")
	checkPosture(t, "compose.server.yml", svc, []string{"image", "command", "user", "privileged", "cap_drop", "security_opt",
		"read_only", "tmpfs", "volumes", "expose", "networks", "restart", "mem_limit", "pids_limit", "healthcheck"})
	if img, _ := svc["image"].(string); !strings.HasPrefix(img, "ghcr.io/hafio/vpn-mesh-utility:${MESHCTL_VERSION:?") {
		t.Errorf("image %q must be the pinned release, required from the environment", img)
	}
	nets, _ := doc["networks"].(map[string]any)
	dok, _ := nets["dokploy-network"].(map[string]any)
	if dok["external"] != true {
		t.Error("dokploy-network must be external (Dokploy owns it)")
	}
}

func TestDevComposePosture(t *testing.T) {
	_, svc := service(t, "compose.dev.yml")
	checkPosture(t, "compose.dev.yml", svc, []string{"image", "command", "user", "privileged", "cap_drop", "security_opt",
		"read_only", "tmpfs", "volumes", "ports", "restart", "mem_limit", "pids_limit", "healthcheck"})
	if !reflect.DeepEqual(svc["ports"], []any{"127.0.0.1:8080:8080"}) {
		t.Errorf("the dev stack publishes on loopback only: %#v", svc["ports"])
	}
}

func mustRead(t *testing.T, name string) []byte {
	t.Helper()
	b, err := os.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	return b
}
