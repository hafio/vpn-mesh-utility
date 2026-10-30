package deploy

import (
	"errors"
	"io"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"

	"go.yaml.in/yaml/v3"
)

// containersDoc is docs/containers.md: the container samples meshctl ships
// instead of generating deployments (amendment A48).
var containersDoc = filepath.Join("..", "docs", "containers.md")

// image is the published image's name; a sample service or container with
// it is meshctl's, any other (an app on the mesh) is the operator's.
const image = "ghcr.io/hafio/vpn-mesh-utility:"

// What a node's sample may add to a locked-down container, and nothing
// more; the control plane may add nothing (docs/security.md).
var (
	sampleCaps    = []string{"NET_ADMIN", "NET_BIND_SERVICE"}
	sampleDevices = []string{"/dev/net/tun:/dev/net/tun"}
	sampleSysctls = []string{"net.ipv4.ip_forward", "net.ipv4.conf.all.proxy_arp", "net.ipv6.conf.all.disable_ipv6",
		"net.ipv6.conf.all.forwarding", "net.ipv6.conf.all.proxy_ndp"}
	sampleSockets = []string{"/var/run/docker.sock", "/run/podman/podman.sock"}
	composeKeys   = []string{"image", "environment", "user", "privileged", "network_mode", "cap_drop", "cap_add", "security_opt",
		"read_only", "tmpfs", "devices", "sysctls", "volumes", "ports", "networks", "restart", "mem_limit", "pids_limit"}
)

// machineID is the host's machine id, which every node's sample mounts
// read-only: it pins the node to its machine (amendment A50).
const machineID = "/etc/machine-id:/etc/machine-id:ro"

// tokenDir is where the global Swarm service reads each host's token.
const tokenDir = "/etc/meshctl:/run/meshctl:ro"

// stackKeys are the keys a node's Swarm service may use: a stack has no
// security_opt, devices, network_mode, restart or compose limits.
var stackKeys = []string{"image", "environment", "user", "cap_drop", "cap_add", "read_only", "sysctls", "volumes", "ports",
	"networks", "deploy"}

// block is a fenced code block of the doc: its language and its text.
type block struct{ lang, text string }

func docBlocks(t *testing.T) []block {
	t.Helper()
	b, err := os.ReadFile(containersDoc)
	if err != nil {
		t.Fatal(err)
	}
	var (
		out []block
		cur *block
	)
	for _, line := range strings.Split(string(b), "\n") {
		switch {
		case cur == nil && strings.HasPrefix(line, "```"):
			cur = &block{lang: strings.TrimPrefix(line, "```")}
		case cur != nil && line == "```":
			out = append(out, *cur)
			cur = nil
		case cur != nil:
			cur.text += line + "\n"
		}
	}
	if cur != nil {
		t.Fatal("an unterminated code block")
	}
	return out
}

// Every sample in docs/containers.md parses and keeps the posture the doc
// states: compose services and Kubernetes containers of meshctl's image,
// and every `podman run` line. The counts keep a sample from dropping out
// of the check unnoticed.
func TestContainerSamples(t *testing.T) {
	var compose, stack, kube, run int
	for i, b := range docBlocks(t) {
		switch b.lang {
		case "yaml":
			dec := yaml.NewDecoder(strings.NewReader(b.text))
			for {
				var doc map[string]any
				err := dec.Decode(&doc)
				if errors.Is(err, io.EOF) {
					break
				}
				if err != nil {
					t.Fatalf("block %d does not parse: %v\n%s", i, err, b.text)
				}
				compose += checkComposeSample(t, doc)
				stack += checkStackSample(t, doc)
				kube += checkKubeSample(t, doc)
			}
		case "sh":
			for _, cmd := range shellCommands(b.text) {
				if len(cmd) > 2 && cmd[1] == "run" && (cmd[0] == "podman" || cmd[0] == "docker") {
					checkRunSample(t, cmd)
					run++
				}
			}
		case "":
		default:
			t.Errorf("block %d is %q: the doc's samples are yaml or sh", i, b.lang)
		}
	}
	if compose != 3 || stack != 3 || kube != 2 || run != 3 {
		t.Errorf("checked %d compose services, %d Swarm services, %d Kubernetes containers, %d podman runs; want 3, 3, 2, 3",
			compose, stack, kube, run)
	}
}

func checkComposeSample(t *testing.T, doc map[string]any) int {
	t.Helper()
	svcs, _ := doc["services"].(map[string]any)
	n := 0
	for name, v := range svcs {
		svc, _ := v.(map[string]any)
		if img, _ := svc["image"].(string); !strings.HasPrefix(img, image) || svc["deploy"] != nil {
			continue
		}
		n++
		env, _ := svc["environment"].(map[string]any)
		control := env["MESHCTL_MODE"] == "control"
		for k := range svc {
			if !slices.Contains(composeKeys, k) {
				t.Errorf("%s: %s is not a key the samples may use", name, k)
			}
		}
		want := map[string]any{"privileged": false, "cap_drop": []any{"ALL"}, "security_opt": []any{"no-new-privileges=true"}, "read_only": true}
		for k, v := range want {
			if !reflect.DeepEqual(svc[k], v) {
				t.Errorf("%s: %s = %#v, want %#v", name, k, svc[k], v)
			}
		}
		if user, _ := svc["user"].(string); control && user != "65532:65532" || !control && user != "0:0" {
			t.Errorf("%s: user %q (the control plane is 65532, a node 0 for its capabilities)", name, user)
		}
		if mode, ok := svc["network_mode"]; ok && (control || mode != "host") {
			t.Errorf("%s: network_mode %v", name, mode)
		}
		for _, c := range list(svc["cap_add"]) {
			if control || !slices.Contains(sampleCaps, c) {
				t.Errorf("%s: cap_add %s", name, c)
			}
		}
		for _, d := range list(svc["devices"]) {
			if control || !slices.Contains(sampleDevices, d) {
				t.Errorf("%s: device %s", name, d)
			}
		}
		sysctls, _ := svc["sysctls"].(map[string]any)
		for k := range sysctls {
			if control || !slices.Contains(sampleSysctls, k) {
				t.Errorf("%s: sysctl %s", name, k)
			}
		}
		for _, v := range list(svc["volumes"]) {
			if !sampleVolume(v, !control) {
				t.Errorf("%s: volume %s", name, v)
			}
		}
		if !control && !slices.Contains(list(svc["volumes"]), machineID) {
			t.Errorf("%s: no %s volume; a node needs the host's machine id", name, machineID)
		}
		for _, p := range list(svc["ports"]) {
			if control && !strings.HasPrefix(p, "127.0.0.1:") || !control && !strings.HasSuffix(p, "/udp") {
				t.Errorf("%s: port %s (the control plane on loopback behind a proxy, a node its WireGuard port)", name, p)
			}
		}
		for k, v := range env {
			if s, _ := v.(string); k == "MESHCTL_TOKEN" && s != "${MESHCTL_TOKEN:-}" {
				t.Errorf("%s: MESHCTL_TOKEN = %q: a sample takes it from the shell or a .env file, never a value", name, s)
			}
		}
	}
	return n
}

// checkStackSample holds a node's Swarm services (amendment A53) to the
// posture docs/containers.md states: kernel WireGuard, the two
// capabilities, a read-only root, pinned by a label, UDP published in host
// mode, and nothing a stack cannot state.
func checkStackSample(t *testing.T, doc map[string]any) int {
	t.Helper()
	svcs, _ := doc["services"].(map[string]any)
	n := 0
	for name, v := range svcs {
		svc, _ := v.(map[string]any)
		if img, _ := svc["image"].(string); !strings.HasPrefix(img, image) || svc["deploy"] == nil {
			continue
		}
		n++
		for k := range svc {
			if !slices.Contains(stackKeys, k) {
				t.Errorf("%s: %s is not a key a Swarm sample may use", name, k)
			}
		}
		if !reflect.DeepEqual(svc["cap_drop"], []any{"ALL"}) || svc["read_only"] != true || svc["user"] != "0:0" {
			t.Errorf("%s: cap_drop %v, read_only %v, user %v; want [ALL], true, 0:0", name, svc["cap_drop"], svc["read_only"], svc["user"])
		}
		for _, c := range list(svc["cap_add"]) {
			if !slices.Contains(sampleCaps, c) {
				t.Errorf("%s: cap_add %s", name, c)
			}
		}
		sysctls, _ := svc["sysctls"].(map[string]any)
		for k := range sysctls {
			if !slices.Contains(sampleSysctls, k) {
				t.Errorf("%s: sysctl %s", name, k)
			}
		}
		env, _ := svc["environment"].(map[string]any)
		if env["MESHCTL_WG_BACKEND"] != "kernel" {
			t.Errorf("%s: MESHCTL_WG_BACKEND %v; Swarm cannot map /dev/net/tun, so the agent needs kernel WireGuard", name, env["MESHCTL_WG_BACKEND"])
		}
		if tok, ok := env["MESHCTL_TOKEN"]; ok && tok != "${MESHCTL_TOKEN:-}" {
			t.Errorf("%s: MESHCTL_TOKEN = %v: a sample takes it from the shell, never a value", name, tok)
		}
		vols, _ := svc["volumes"].([]any)
		id := false
		for _, x := range vols {
			switch m := x.(type) {
			case string:
				id = id || m == machineID
				if !sampleVolume(m, true) && m != tokenDir {
					t.Errorf("%s: volume %s", name, m)
				}
			case map[string]any:
				if m["type"] != "tmpfs" || m["target"] != "/tmp" {
					t.Errorf("%s: volume %v", name, m)
				}
			}
		}
		if !id {
			t.Errorf("%s: no %s volume; a node needs the host's machine id", name, machineID)
		}
		ports, _ := svc["ports"].([]any)
		for _, x := range ports {
			if p, _ := x.(map[string]any); p["protocol"] != "udp" || p["mode"] != "host" {
				t.Errorf("%s: port %v; a node publishes its WireGuard port, UDP, in host mode", name, x)
			}
		}
		for _, nw := range list(svc["networks"]) {
			switch {
			case nw == "hostnet":
				if h := dig(doc, "networks", "hostnet"); h["external"] != true || h["name"] != "host" {
					t.Errorf("%s: hostnet is %v, not the external network host", name, h)
				}
			case nw != "egress" && !strings.HasPrefix(nw, "meshctl-mesh"):
				t.Errorf("%s: network %s", name, nw)
			}
		}
		if c := list(dig(svc, "deploy", "placement")["constraints"]); len(c) != 1 || !strings.HasPrefix(c[0], "node.labels.meshctl.") {
			t.Errorf("%s: placement %v; a node's agent is pinned to its host by its label", name, c)
		}
	}
	return n
}

// sampleVolume is a volume a sample may mount: data at /data (a named
// volume, or a host directory for a node), a node's read-only runtime
// socket, or the host's machine id.
func sampleVolume(v string, node bool) bool {
	src, dst, _ := strings.Cut(v, ":")
	switch {
	case node && v == machineID:
		return true
	case dst == "/data":
		// The samples run on Linux hosts: their paths are POSIX on every OS
		// the test runs on.
		return !strings.HasPrefix(src, "/") && !strings.HasPrefix(src, ".") || node && path.IsAbs(src)
	case node && dst == "/var/run/docker.sock:ro":
		return slices.Contains(sampleSockets, src)
	}
	return false
}

func checkKubeSample(t *testing.T, doc map[string]any) int {
	t.Helper()
	if doc["kind"] != "Deployment" {
		return 0
	}
	spec := dig(doc, "spec", "template", "spec")
	for _, k := range []string{"hostPID", "hostIPC"} {
		if spec[k] != nil {
			t.Errorf("a Deployment sets %s", k)
		}
	}
	podUser := dig(spec, "securityContext")["runAsUser"]
	n := 0
	cs, _ := spec["containers"].([]any)
	for _, v := range cs {
		c, _ := v.(map[string]any)
		if img, _ := c["image"].(string); !strings.HasPrefix(img, image) {
			continue
		}
		n++
		sc := dig(c, "securityContext")
		for k, want := range map[string]any{"privileged": false, "allowPrivilegeEscalation": false, "readOnlyRootFilesystem": true} {
			if sc[k] != want {
				t.Errorf("container %v: securityContext.%s = %v, want %v", c["name"], k, sc[k], want)
			}
		}
		caps := dig(sc, "capabilities")
		if !reflect.DeepEqual(caps["drop"], []any{"ALL"}) {
			t.Errorf("container %v drops %v, not ALL", c["name"], caps["drop"])
		}
		control := false
		env, _ := c["env"].([]any)
		for _, e := range env {
			ev, _ := e.(map[string]any)
			switch ev["name"] {
			case "MESHCTL_MODE":
				control = ev["value"] == "control"
			case "MESHCTL_TOKEN":
				t.Errorf("container %v has MESHCTL_TOKEN in its manifest: use MESHCTL_TOKEN_FILE and a secret", c["name"])
			}
		}
		for _, a := range list(caps["add"]) {
			if control || !slices.Contains(sampleCaps, a) {
				t.Errorf("container %v adds %s", c["name"], a)
			}
		}
		if !control && !mountsMachineID(c) {
			t.Errorf("container %v does not mount the host's /etc/machine-id read-only", c["name"])
		}
		user := sc["runAsUser"]
		if user == nil {
			user = podUser
		}
		if control && user != 65532 || !control && user != 0 || control && spec["hostNetwork"] != nil {
			t.Errorf("container %v: runAsUser %v, hostNetwork %v", c["name"], user, spec["hostNetwork"])
		}
	}
	vols, _ := spec["volumes"].([]any)
	for _, v := range vols {
		vol, _ := v.(map[string]any)
		if hp := dig(vol, "hostPath"); len(hp) > 0 && hp["path"] != "/var/lib/meshctl" && hp["path"] != "/etc/machine-id" {
			t.Errorf("volume %v mounts the host's %v; only a node's data directory and its machine id", vol["name"], hp["path"])
		}
	}
	return n
}

// mountsMachineID reports whether a Kubernetes container mounts
// /etc/machine-id read-only.
func mountsMachineID(c map[string]any) bool {
	ms, _ := c["volumeMounts"].([]any)
	for _, m := range ms {
		vm, _ := m.(map[string]any)
		if vm["mountPath"] == "/etc/machine-id" && vm["readOnly"] == true {
			return true
		}
	}
	return false
}

// checkRunSample holds a `podman run` line to the compose samples' posture.
func checkRunSample(t *testing.T, cmd []string) {
	t.Helper()
	line := strings.Join(cmd, " ")
	control := slices.Contains(cmd, "MESHCTL_MODE=control")
	for _, want := range []string{"--cap-drop=ALL", "--security-opt=no-new-privileges", "--read-only"} {
		if !slices.Contains(cmd, want) {
			t.Errorf("%s: no %s", line, want)
		}
	}
	user := "--user=0:0"
	if control {
		user = "--user=65532:65532"
	}
	if !slices.Contains(cmd, user) {
		t.Errorf("%s: no %s", line, user)
	}
	if i := slices.Index(cmd, machineID); !control && (i < 1 || cmd[i-1] != "-v") {
		t.Errorf("%s: no -v %s; a node needs the host's machine id", line, machineID)
	}
	for i, a := range cmd {
		flag, val, _ := strings.Cut(a, "=")
		switch flag {
		case "--privileged", "--pid", "--ipc", "--userns":
			t.Errorf("%s: %s", line, a)
		case "--cap-add":
			if control || !slices.Contains(sampleCaps, val) {
				t.Errorf("%s: %s", line, a)
			}
		case "--device":
			if control || val != "/dev/net/tun" {
				t.Errorf("%s: %s", line, a)
			}
		case "--sysctl":
			k, _, _ := strings.Cut(val, "=")
			if control || !slices.Contains(sampleSysctls, k) {
				t.Errorf("%s: %s", line, a)
			}
		case "--network":
			if val == "host" && control {
				t.Errorf("%s: %s", line, a)
			}
		case "-v":
			if i+1 >= len(cmd) || !sampleVolume(cmd[i+1], !control) {
				t.Errorf("%s: a volume the samples may not mount", line)
			}
		case "-e":
			if i+1 < len(cmd) && strings.HasPrefix(cmd[i+1], "MESHCTL_TOKEN=") {
				t.Errorf("%s: the token on the command line; -e MESHCTL_TOKEN takes it from the shell", line)
			}
		}
	}
}

// shellCommands splits a sh block into commands, joining continued lines.
func shellCommands(text string) [][]string {
	var out [][]string
	var cur []string
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		cont := strings.HasSuffix(line, "\\")
		cur = append(cur, strings.Fields(strings.TrimSuffix(line, "\\"))...)
		if !cont && len(cur) > 0 {
			out = append(out, cur)
			cur = nil
		}
	}
	return out
}

func list(v any) []string {
	var out []string
	vs, _ := v.([]any)
	for _, x := range vs {
		s, _ := x.(string)
		out = append(out, s)
	}
	return out
}

// dig is the map at keys under m; nil when one is missing.
func dig(m map[string]any, keys ...string) map[string]any {
	for _, k := range keys {
		m, _ = m[k].(map[string]any)
	}
	return m
}
