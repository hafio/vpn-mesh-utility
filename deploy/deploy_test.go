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

func TestControlComposePosture(t *testing.T) {
	if !bytes.Equal(ComposeControl(), mustRead(t, "compose.control.yml")) {
		t.Fatal("the embedded compose file is stale")
	}
	doc, svc := service(t, "compose.control.yml")
	checkPosture(t, "compose.control.yml", svc, []string{"image", "environment", "user", "privileged", "cap_drop", "security_opt",
		"read_only", "tmpfs", "volumes", "expose", "networks", "restart", "mem_limit", "pids_limit", "healthcheck"})
	if !reflect.DeepEqual(svc["environment"], controlEnv()) {
		t.Errorf("the control plane's environment = %#v, want exactly %#v", svc["environment"], controlEnv())
	}
	if !reflect.DeepEqual(svc["expose"], []any{"${MESHCTL_PORT:-8080}"}) {
		t.Errorf("expose = %#v: the control plane's port only", svc["expose"])
	}
	if img, _ := svc["image"].(string); !strings.HasPrefix(img, "ghcr.io/hafio/vpn-mesh-utility:${MESHCTL_VERSION:?") {
		t.Errorf("image %q must be the pinned release, required from the environment", img)
	}
	nets, _ := doc["networks"].(map[string]any)
	dok, _ := nets["dokploy-network"].(map[string]any)
	if dok["external"] != true {
		t.Error("dokploy-network must be external (Dokploy owns it)")
	}
}

// controlEnv is the control plane's environment in its compose and stack
// files (amendments A47, A48): `meshctl start` in control mode, every value
// from the app's environment, the origin required.
func controlEnv() map[string]any {
	return map[string]any{
		"MESHCTL_MODE":            "control",
		"MESHCTL_PORT":            "${MESHCTL_PORT:-8080}",
		"MESHCTL_URL":             "${MESHCTL_URL:?set MESHCTL_URL to the public https origin}",
		"MESHCTL_TRUSTED_PROXIES": "${MESHCTL_TRUSTED_PROXIES:-}",
		"MESHCTL_HUB_DOWN_AFTER":  "${MESHCTL_HUB_DOWN_AFTER:-}",
		"MESHCTL_LOG_LEVEL":       "${MESHCTL_LOG_LEVEL:-}",
		"TZ":                      "${TZ:-UTC}",
	}
}

// startHealth is the image's own health check, which follows MESHCTL_MODE.
func startHealth() []any { return []any{"CMD", "/app/meshctl", "healthcheck", "start"} }

func TestDevComposePosture(t *testing.T) {
	_, svc := service(t, "compose.dev.yml")
	checkPosture(t, "compose.dev.yml", svc, []string{"image", "environment", "user", "privileged", "cap_drop", "security_opt",
		"read_only", "tmpfs", "volumes", "ports", "restart", "mem_limit", "pids_limit", "healthcheck"})
	if env := map[string]any{"MESHCTL_MODE": "control", "MESHCTL_URL": "http://127.0.0.1:8080"}; !reflect.DeepEqual(svc["environment"], env) {
		t.Errorf("the dev control plane's environment = %#v, want %#v", svc["environment"], env)
	}
	if hc, _ := svc["healthcheck"].(map[string]any); !reflect.DeepEqual(hc["test"], startHealth()) {
		t.Errorf("healthcheck = %#v", svc["healthcheck"])
	}
	if !reflect.DeepEqual(svc["ports"], []any{"127.0.0.1:8080:8080"}) {
		t.Errorf("the dev stack publishes on loopback only: %#v", svc["ports"])
	}
}

// The Swarm stack states the posture a stack can: non-root, no
// capabilities, read-only root, one task pinned to a labelled node and
// stopped before the next starts; nothing that widens it.
func TestControlStackPosture(t *testing.T) {
	doc, svc := service(t, "stack.control.yml")
	if got, want := keys(svc), []string{"image", "environment", "user", "cap_drop", "read_only", "volumes", "networks",
		"stop_grace_period", "healthcheck", "deploy"}; !sameSet(got, want) {
		t.Errorf("stack service keys = %v, want exactly %v", got, want)
	}
	for k, v := range map[string]any{"user": "65532:65532", "cap_drop": []any{"ALL"}, "read_only": true,
		"environment": controlEnv()} {
		if !reflect.DeepEqual(svc[k], v) {
			t.Errorf("stack: %s = %#v, want %#v", k, svc[k], v)
		}
	}
	if img, _ := svc["image"].(string); !strings.HasPrefix(img, "ghcr.io/hafio/vpn-mesh-utility:${MESHCTL_VERSION:?") {
		t.Errorf("image %q must be the pinned release, required from the environment", img)
	}
	vols, _ := svc["volumes"].([]any)
	if len(vols) != 2 || vols[0] != "meshctl-data:/data" {
		t.Errorf("stack volumes = %#v: the named data volume and a tmpfs only", vols)
	}
	if tmp, _ := vols[1].(map[string]any); tmp["type"] != "tmpfs" || tmp["target"] != "/tmp" {
		t.Errorf("stack tmpfs = %#v", vols[1])
	}
	dep, _ := svc["deploy"].(map[string]any)
	place, _ := dep["placement"].(map[string]any)
	update, _ := dep["update_config"].(map[string]any)
	if dep["mode"] != "replicated" || dep["replicas"] != 1 || !reflect.DeepEqual(place["constraints"], []any{"node.labels.meshctl.control == true"}) ||
		update["order"] != "stop-first" {
		t.Errorf("stack deploy = %#v", dep)
	}
	nets, _ := doc["networks"].(map[string]any)
	if dok, _ := nets["dokploy-network"].(map[string]any); dok["external"] != true {
		t.Error("dokploy-network must be external (Dokploy owns it)")
	}
	// The same environment and health check as the compose file: only how
	// it is run differs.
	_, compose := service(t, "compose.control.yml")
	if !reflect.DeepEqual(svc["environment"], compose["environment"]) || !reflect.DeepEqual(svc["healthcheck"], compose["healthcheck"]) {
		t.Errorf("the stack runs with %#v, the compose file with %#v", svc["environment"], compose["environment"])
	}
	if hc, _ := svc["healthcheck"].(map[string]any); !reflect.DeepEqual(hc["test"], startHealth()) {
		t.Errorf("healthcheck = %#v", svc["healthcheck"])
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
