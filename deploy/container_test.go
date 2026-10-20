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

// containerParams is a node in container mode at the first container /24 of
// the default mesh: a hub publishing 51999 when hub is set.
func containerParams(hub bool) ContainerParams {
	p := ContainerParams{Node: "box", Server: "https://mesh.example.com", Version: "v1.2.3", ListenPort: 51820,
		Network: "10.77.0.0/16", Subnet: "10.77.128.0/24", Range: "10.77.128.128/25", Bridge: "10.77.128.1",
		Gateway: "10.77.128.2", Zone: "mesh"}
	if hub {
		p.Node, p.Hub, p.PublishPort = "hub", true, 51999
	}
	return p
}

// The golden variants: a hub on Docker with the defaults, a Podman client
// on userspace WireGuard with a host directory, a Swarm client on the
// kernel module without names, its services resolving through the servers.
var containerVariants = map[string]func() ContainerParams{
	"hub-docker": func() ContainerParams { return containerParams(true) },
	"podman": func() ContainerParams {
		p := containerParams(false)
		p.Runtime, p.Backend, p.DataDir = RuntimePodman, ContainerUserspace, "/srv/meshctl"
		return p
	},
	"swarm": func() ContainerParams {
		p := containerParams(false)
		p.Runtime, p.Backend, p.NoNames = RuntimeSwarm, ContainerKernel, true
		p.Resolvers = []string{"10.77.0.1", "10.77.0.2"}
		return p
	},
}

// parseContainer parses a rendered file and returns its document and the
// one service: exactly name, services, networks (and volumes).
func parseContainer(t *testing.T, b []byte) (map[string]any, map[string]any) {
	t.Helper()
	var doc map[string]any
	if err := yaml.Unmarshal(b, &doc); err != nil {
		t.Fatalf("the file does not parse: %v\n%s", err, b)
	}
	top := []string{"name", "services", "networks"}
	if _, ok := doc["volumes"]; ok {
		top = append(top, "volumes")
	}
	if !sameSet(keys(doc), top) || doc["name"] != ContainerService {
		t.Fatalf("top-level keys %v, name %v", keys(doc), doc["name"])
	}
	svcs, _ := doc["services"].(map[string]any)
	svc, _ := svcs[ContainerService].(map[string]any)
	if len(svcs) != 1 || svc == nil {
		t.Fatalf("the file must define exactly the %s service:\n%s", ContainerService, b)
	}
	return doc, svc
}

func TestContainerComposeGolden(t *testing.T) {
	for name, params := range containerVariants {
		got, err := RenderContainerCompose(params())
		if err != nil {
			t.Fatal(err)
		}
		path := filepath.Join("testdata", "compose.container-"+name+".golden.yml")
		if *update {
			if err := os.WriteFile(path, got, 0o644); err != nil {
				t.Fatal(err)
			}
		}
		if !bytes.Equal(got, mustRead(t, path)) {
			t.Errorf("the %s compose file differs from %s; review the change, then ./scripts/dev.sh regen\n%s", name, path, got)
		}
	}
	// The defaults: Docker, auto, a named volume.
	p := containerParams(false)
	def, err := RenderContainerCompose(p)
	if err != nil {
		t.Fatal(err)
	}
	p.Runtime, p.Backend = RuntimeDocker, ContainerAuto
	if want, _ := RenderContainerCompose(p); !bytes.Equal(def, want) {
		t.Errorf("the defaults are not docker and auto:\n%s", def)
	}
	if !strings.Contains(string(def), ContainerVolume+":/data") {
		t.Errorf("the default data is not the named volume:\n%s", def)
	}
}

// The container-mode agent's posture, exactly (docs/security.md): uid 0,
// NET_ADMIN and NET_BIND_SERVICE in its own namespace, the two sysctls, the
// tun device unless the kernel backend, one published UDP port on a hub,
// the egress and mesh networks, the runtime's socket unless --no-names --
// and nothing else: no host network, no other capability.
func TestContainerComposePosture(t *testing.T) {
	for name, params := range containerVariants {
		p := params()
		out, err := RenderContainerCompose(p)
		if err != nil {
			t.Fatal(err)
		}
		doc, svc := parseContainer(t, out)
		want := []string{"image", "command", "user", "privileged", "cap_drop", "cap_add", "security_opt", "read_only", "tmpfs",
			"sysctls", "volumes", "networks", "restart", "mem_limit", "pids_limit", "healthcheck"}
		exact := map[string]any{
			"image":        Image + ":1.2.3",
			"command":      []any{"agent", "run", "--data", "/data", "--wg-backend", p.Backend},
			"user":         "0:0",
			"privileged":   false,
			"cap_drop":     []any{"ALL"},
			"cap_add":      []any{"NET_ADMIN", "NET_BIND_SERVICE"},
			"security_opt": []any{"no-new-privileges=true"},
			"read_only":    true,
			"sysctls":      map[string]any{"net.ipv4.ip_forward": "1", "net.ipv4.conf.all.proxy_arp": "1"},
			"networks":     []any{"egress", MeshNetwork},
		}
		if p.Backend == "" {
			exact["command"] = []any{"agent", "run", "--data", "/data", "--wg-backend", ContainerAuto}
		}
		data := ContainerVolume + ":/data"
		if p.DataDir != "" {
			data = p.DataDir + ":/data"
		}
		if p.NoNames {
			exact["command"] = append(exact["command"].([]any), "--runtime-socket", "off")
			exact["volumes"] = []any{data}
			if strings.Contains(string(out), ".sock") {
				t.Errorf("%s: --no-names, yet the file mounts a socket:\n%s", name, out)
			}
		} else {
			exact["volumes"] = []any{data, hostSocket(p.Runtime) + ":/var/run/docker.sock:ro"}
		}
		if p.Backend != ContainerKernel {
			want = append(want, "devices")
			exact["devices"] = []any{"/dev/net/tun:/dev/net/tun"}
		} else if strings.Contains(string(out), "/dev/") {
			t.Errorf("%s: the kernel variant mentions a device:\n%s", name, out)
		}
		if p.Hub {
			want = append(want, "ports")
			exact["ports"] = []any{"51999:51820/udp"}
		}
		if !sameSet(keys(svc), want) {
			t.Errorf("%s: service keys = %v, want exactly %v", name, keys(svc), want)
		}
		for k, v := range exact {
			if !reflect.DeepEqual(svc[k], v) {
				t.Errorf("%s: %s = %#v, want %#v", name, k, svc[k], v)
			}
		}
		nets := map[string]any{"egress": map[string]any{}, MeshNetwork: map[string]any{"external": true}}
		if !reflect.DeepEqual(doc["networks"], nets) {
			t.Errorf("%s: networks = %#v, want %#v", name, doc["networks"], nets)
		}
		for _, w := range []string{"network_mode", "pid:", "ipc:", "userns_mode", "privileged: true", "label", "containerd.sock",
			"seccomp", "apparmor", "SYS_ADMIN", "NET_RAW", "cgroup", "ipv4_address", "/proc"} {
			if strings.Contains(string(out), w) {
				t.Errorf("%s: the file contains the widening token %q", name, w)
			}
		}
	}
}

// The header creates the mesh network the way the 6a trial proved per
// runtime, and shows how a service attaches.
func TestContainerComposeHeader(t *testing.T) {
	for name, want := range map[string][]string{
		"hub-docker": {"docker network create --internal --subnet 10.77.0.0/16 --ip-range 10.77.128.128/25 --gateway 10.77.128.1",
			"-o com.docker.network.bridge.name=meshctl0 meshctl-mesh", "the host publishes 51999/udp", "docker compose up -d",
			"<service>.mesh (every node's, round robin) and <service>.<node>.mesh", "`meshctl node compose --no-names` leaves it out"},
		"podman": {"podman network create --internal --disable-dns --subnet 10.77.0.0/16", "--interface-name meshctl0 meshctl-mesh",
			"podman compose run --rm", "wireguard-go", "systemctl enable --now podman.socket", "/run/user/<uid>/podman/podman.sock"},
		"swarm": {"docker network create --config-only --subnet 10.77.0.0/16 --ip-range 10.77.128.128/25 --gateway 10.77.128.1 meshctl-mesh-config",
			"--scope swarm --attachable --config-from meshctl-mesh-config meshctl-mesh", "the host's module", "Names: none",
			"dns: [10.77.0.1, 10.77.0.2]", "its dns is the servers"},
	} {
		out, err := RenderContainerCompose(containerVariants[name]())
		if err != nil {
			t.Fatal(err)
		}
		if name != "swarm" {
			want = append(want, "dns: [10.77.128.2]")
		}
		for _, w := range append(want, "dns_search: [mesh]", `dns_opt: ["ndots:1"]`,
			"--server https://mesh.example.com --name ") {
			if !strings.Contains(string(out), w) {
				t.Errorf("%s: the header lacks %q:\n%s", name, w, out)
			}
		}
	}
}

// A host-mode node's mesh network is on the agent's own bridge: Docker takes
// it as it is and filters nothing the guard does not, Podman attaches to it
// unmanaged; Swarm carries the same options in its config-only network.
func TestHostModeNetworkSetup(t *testing.T) {
	p := NetworkParams{Host: true, Network: "10.77.0.0/16", Range: "10.77.128.128/25", Bridge: "10.77.128.1", Gateway: "10.77.128.1", Zone: "mesh"}
	for runtime, want := range map[string][]string{
		RuntimeDocker: {"docker network create --subnet 10.77.0.0/16", "com.docker.network.bridge.inhibit_ipv4=true",
			"enable_ip_masquerade=false", "gateway_mode_ipv4=nat-unprotected", "    meshctl-mesh"},
		RuntimePodman: {"podman network create --disable-dns", "-o mode=unmanaged -o no_default_route=true meshctl-mesh"},
		RuntimeSwarm:  {"--config-only", "inhibit_ipv4=true", "meshctl-mesh-config", "--scope swarm"},
	} {
		p.Runtime = runtime
		s, err := MeshNetworkSetup(p)
		if err != nil {
			t.Fatal(err)
		}
		text := strings.Join(s.Commands, "\n")
		for _, w := range want {
			if !strings.Contains(text, w) {
				t.Errorf("%s: the commands lack %q:\n%s", runtime, w, text)
			}
		}
		if strings.Contains(text, "--internal") {
			t.Errorf("%s: a host-mode network must not be internal (podman turns its forwarding off):\n%s", runtime, text)
		}
		if !strings.Contains(strings.Join(s.Service, "\n"), "dns: [10.77.128.1]") {
			t.Errorf("%s: the service example = %v", runtime, s.Service)
		}
		// With the servers known, a Swarm service resolves through them;
		// anything else through its node.
		q := p
		q.Resolvers = []string{"10.77.0.1", "10.77.0.2", "10.77.0.3"}
		want := "dns: [10.77.128.1]"
		if runtime == RuntimeSwarm {
			want = "dns: [10.77.0.1, 10.77.0.2, 10.77.0.3]"
		}
		if s, err := MeshNetworkSetup(q); err != nil || !strings.Contains(strings.Join(s.Service, "\n"), want) {
			t.Errorf("%s with the servers: %v, %v", runtime, s.Service, err)
		}
	}
	p.Runtime = ""
	if s, err := MeshNetworkSetup(p); err != nil || !strings.HasPrefix(s.Commands[0], "  docker network create") {
		t.Errorf("the default runtime = %v, %v", s, err)
	}
	for name, edit := range map[string]func(*NetworkParams){
		"runtime": func(p *NetworkParams) { p.Runtime = "lxc" },
		"network": func(p *NetworkParams) { p.Network = "10.77.0.0/16 $(id)" },
		"range":   func(p *NetworkParams) { p.Range = "all" },
		"bridge":  func(p *NetworkParams) { p.Bridge = "10.77.128.1;reboot" },
		"gateway": func(p *NetworkParams) { p.Gateway = "gw" },
		"zone":    func(p *NetworkParams) { p.Zone = "corp.local" },
		"dns":     func(p *NetworkParams) { p.Resolvers = []string{"10.77.0.1]\n      privileged: true"} },
	} {
		q := p
		edit(&q)
		if _, err := MeshNetworkSetup(q); err == nil {
			t.Errorf("%s: accepted", name)
		}
	}
}

func TestContainerComposeRefusesHostileValues(t *testing.T) {
	cases := map[string]func(*ContainerParams){
		"newline in the data dir":  func(p *ContainerParams) { p.DataDir = "/srv/x\n    privileged: true" },
		"interpolation in data":    func(p *ContainerParams) { p.DataDir = "/srv/${HOME}" },
		"relative data dir":        func(p *ContainerParams) { p.DataDir = "srv/meshctl" },
		"key in the node name":     func(p *ContainerParams) { p.Node = "box: x" },
		"server with a path":       func(p *ContainerParams) { p.Server = "https://mesh.example.com/x" },
		"non-canonical server":     func(p *ContainerParams) { p.Server = "https://Mesh.Example.com" },
		"a development build":      func(p *ContainerParams) { p.Version = "dev" },
		"unknown backend":          func(p *ContainerParams) { p.Backend = "boringtun" },
		"unknown runtime":          func(p *ContainerParams) { p.Runtime = "lxc" },
		"listen port":              func(p *ContainerParams) { p.ListenPort = 0 },
		"publish port on a hub":    func(p *ContainerParams) { p.Hub, p.PublishPort = true, 70000 },
		"network with a command":   func(p *ContainerParams) { p.Network = "10.77.0.0/16 $(id)" },
		"subnet garbage":           func(p *ContainerParams) { p.Subnet = "10.77.128.0/24\n- x" },
		"range garbage":            func(p *ContainerParams) { p.Range = "all" },
		"bridge garbage":           func(p *ContainerParams) { p.Bridge = "10.77.128.1;reboot" },
		"gateway garbage":          func(p *ContainerParams) { p.Gateway = "gateway" },
		"zone with a quote":        func(p *ContainerParams) { p.Zone = `mesh"]` },
		"zone ending in .local":    func(p *ContainerParams) { p.Zone = "corp.local" },
		"newline in the node name": func(p *ContainerParams) { p.Node = "box\nprivileged: true" },
		"resolver garbage":         func(p *ContainerParams) { p.Resolvers = []string{"10.77.0.1, 1.1.1.1"} },
	}
	for name, edit := range cases {
		p := containerParams(false)
		edit(&p)
		if out, err := RenderContainerCompose(p); err == nil || out != nil {
			t.Errorf("%s: rendered", name)
		}
	}
}

// Fed straight to the renderer, past the validators, a hostile value is
// still only a value: the keys, the posture and the comment block hold.
func TestContainerRendererHoldsItsStructure(t *testing.T) {
	for _, h := range []string{"x\n    privileged: true", "${HOME}", "x\n---\nservices: {}", "x: y", "a #b", `'"`, "\x00\x1b[2J"} {
		p := ContainerParams{Node: h, Server: h, Version: h, Backend: h, Runtime: h, DataDir: h, Network: h, Subnet: h, Range: h,
			Bridge: h, Gateway: h, Zone: h, ListenPort: 51820, PublishPort: 51999, Hub: true}
		out, err := renderContainer(p)
		if err != nil {
			t.Fatalf("%q: %v", h, err)
		}
		doc, svc := parseContainer(t, out)
		if _, ok := doc["volumes"]; ok {
			t.Errorf("%q: a host directory was given, yet the file declares a named volume", h)
		}
		if svc["privileged"] != false || svc["user"] != "0:0" || !reflect.DeepEqual(svc["cap_add"], []any{"NET_ADMIN", "NET_BIND_SERVICE"}) {
			t.Errorf("%q moved the posture: %v", h, svc)
		}
		if vols, _ := svc["volumes"].([]any); len(vols) != 2 || vols[0] != strings.ReplaceAll(h, "$", "$$")+":/data" ||
			vols[1] != RuntimeSocket+":"+RuntimeSocket+":ro" {
			t.Errorf("%q: volumes = %#v (a '$' must be doubled)", h, svc["volumes"])
		}
		lines := strings.Split(string(out), "\n")
		i := 0
		for strings.HasPrefix(lines[i], "#") {
			i++
		}
		if lines[i] != "name: "+ContainerService {
			t.Errorf("%q broke the comment block at line %d: %q", h, i+1, lines[i])
		}
	}
}
