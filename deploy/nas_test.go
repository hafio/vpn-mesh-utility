package deploy

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"go.yaml.in/yaml/v3"
)

func nasParams(backend string) NASParams {
	return NASParams{Node: "nas", Server: "https://mesh.example.com", Version: "v1.2.3", Backend: backend}
}

// parseNAS parses a rendered file and returns its service: exactly the name
// and services keys, exactly one service.
func parseNAS(t *testing.T, b []byte) map[string]any {
	t.Helper()
	var doc map[string]any
	if err := yaml.Unmarshal(b, &doc); err != nil {
		t.Fatalf("the file does not parse: %v\n%s", err, b)
	}
	if !sameSet(keys(doc), []string{"name", "services"}) || doc["name"] != NASService {
		t.Fatalf("top-level keys %v, name %v", keys(doc), doc["name"])
	}
	svcs, _ := doc["services"].(map[string]any)
	svc, _ := svcs[NASService].(map[string]any)
	if len(svcs) != 1 || svc == nil {
		t.Fatalf("the file must define exactly the %s service:\n%s", NASService, b)
	}
	return svc
}

func TestNASComposeGolden(t *testing.T) {
	for _, b := range []string{NASUserspace, NASKernel} {
		got, err := RenderNASCompose(nasParams(b))
		if err != nil {
			t.Fatal(err)
		}
		path := filepath.Join("testdata", "compose.nas-"+b+".golden.yml")
		if *update {
			if err := os.WriteFile(path, got, 0o644); err != nil {
				t.Fatal(err)
			}
		}
		if !bytes.Equal(got, mustRead(t, path)) {
			t.Errorf("the %s compose file differs from %s; review the change, then ./scripts/dev.sh regen\n%s", b, path, got)
		}
	}
	// The defaults: userspace (amendment A19), DefaultNASData, and the same
	// image for 1.2.3 as for v1.2.3.
	def, err := RenderNASCompose(NASParams{Node: "nas", Server: "https://mesh.example.com", Version: "1.2.3"})
	if err != nil {
		t.Fatal(err)
	}
	if want, _ := RenderNASCompose(nasParams(NASUserspace)); !bytes.Equal(def, want) {
		t.Errorf("the defaults are not the userspace variant:\n%s", def)
	}
}

// The NAS agent's posture, exactly: the widenings docs/security.md names --
// uid 0, the host network, NET_ADMIN and NET_BIND_SERVICE, the tun device in
// the userspace variant only -- and nothing else.
func TestNASComposePosture(t *testing.T) {
	for _, b := range []string{NASUserspace, NASKernel} {
		out, err := RenderNASCompose(nasParams(b))
		if err != nil {
			t.Fatal(err)
		}
		svc := parseNAS(t, out)
		want := []string{"image", "command", "user", "privileged", "network_mode", "cap_drop", "cap_add", "security_opt",
			"read_only", "tmpfs", "volumes", "restart", "mem_limit", "pids_limit", "healthcheck"}
		exact := map[string]any{
			"image":        Image + ":1.2.3",
			"command":      []any{"agent", "run", "--data", "/data", "--wg-backend", b},
			"user":         "0:0",
			"privileged":   false,
			"network_mode": "host",
			"cap_drop":     []any{"ALL"},
			"cap_add":      []any{"NET_ADMIN", "NET_BIND_SERVICE"},
			"security_opt": []any{"no-new-privileges=true"},
			"read_only":    true,
			"volumes":      []any{DefaultNASData + ":/data"},
		}
		if b == NASUserspace {
			want = append(want, "devices")
			exact["devices"] = []any{"/dev/net/tun:/dev/net/tun"}
		} else if strings.Contains(string(out), "/dev/") {
			t.Errorf("the kernel variant mentions a device:\n%s", out)
		}
		if !sameSet(keys(svc), want) {
			t.Errorf("%s: service keys = %v, want exactly %v", b, keys(svc), want)
		}
		for k, v := range exact {
			if !reflect.DeepEqual(svc[k], v) {
				t.Errorf("%s: %s = %#v, want %#v", b, k, svc[k], v)
			}
		}
		for _, w := range []string{"pid:", "ipc:", "userns_mode", "privileged: true", "docker.sock", "seccomp", "apparmor",
			"SYS_ADMIN", "cgroup", "sysctls", "ports:"} {
			if strings.Contains(string(out), w) {
				t.Errorf("%s: the file contains the widening token %q", b, w)
			}
		}
	}
}

func TestNASComposeRefusesHostileValues(t *testing.T) {
	cases := map[string]func(*NASParams){
		"newline in the data dir":   func(p *NASParams) { p.DataDir = "/volume1/x\n    privileged: true" },
		"interpolation in data dir": func(p *NASParams) { p.DataDir = "/volume1/${HOME}" },
		"document break in data":    func(p *NASParams) { p.DataDir = "/volume1/x\n---\nservices: {}" },
		"relative data dir":         func(p *NASParams) { p.DataDir = "volume1/meshctl" },
		"dot segment in data dir":   func(p *NASParams) { p.DataDir = "/volume1/../etc" },
		"key in the node name":      func(p *NASParams) { p.Node = "nas: x" },
		"newline in the node name":  func(p *NASParams) { p.Node = "nas\nprivileged: true" },
		"server with a path":        func(p *NASParams) { p.Server = "https://mesh.example.com/x" },
		"non-canonical server":      func(p *NASParams) { p.Server = "https://Mesh.Example.com" },
		"a development build":       func(p *NASParams) { p.Version = "dev" },
		"substitution in version":   func(p *NASParams) { p.Version = "v1.2.3-$(id)" },
		"unknown backend":           func(p *NASParams) { p.Backend = "wireguard-go" },
	}
	for name, edit := range cases {
		p := nasParams(NASUserspace)
		edit(&p)
		if out, err := RenderNASCompose(p); err == nil || out != nil {
			t.Errorf("%s: rendered", name)
		}
	}
	p := nasParams(NASUserspace)
	p.Version = "dev"
	if _, err := RenderNASCompose(p); err == nil || !strings.Contains(err.Error(), "development build") {
		t.Errorf("a development build: %v", err)
	}
}

// Fed straight to the renderer, past the validators, a hostile value is
// still only a value: the keys, the posture and the comment block hold.
func TestNASRendererHoldsItsStructure(t *testing.T) {
	for _, h := range []string{"x\n    privileged: true", "${HOME}", "x\n---\nservices: {}", "x: y", "a #b", `'"`, "\x00\x1b[2J"} {
		out, err := renderNAS(NASParams{Node: h, Server: h, Version: h, Backend: h, DataDir: h})
		if err != nil {
			t.Fatalf("%q: %v", h, err)
		}
		svc := parseNAS(t, out)
		want := []string{"image", "command", "user", "privileged", "network_mode", "cap_drop", "cap_add", "security_opt",
			"read_only", "tmpfs", "volumes", "restart", "mem_limit", "pids_limit", "healthcheck"}
		if !sameSet(keys(svc), want) {
			t.Errorf("%q moved the keys: %v", h, keys(svc))
		}
		if svc["privileged"] != false || svc["user"] != "0:0" || !reflect.DeepEqual(svc["cap_add"], []any{"NET_ADMIN", "NET_BIND_SERVICE"}) {
			t.Errorf("%q moved the posture: %v", h, svc)
		}
		if vols, _ := svc["volumes"].([]any); len(vols) != 1 || vols[0] != strings.ReplaceAll(h, "$", "$$")+":/data" {
			t.Errorf("%q: volumes = %#v (a '$' must be doubled)", h, svc["volumes"])
		}
		lines := strings.Split(string(out), "\n")
		i := 0
		for strings.HasPrefix(lines[i], "#") {
			i++
		}
		if lines[i] != "name: "+NASService {
			t.Errorf("%q broke the comment block at line %d: %q", h, i+1, lines[i])
		}
	}
}
