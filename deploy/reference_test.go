package deploy

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// composeDoc is docs/compose.md: every compose and stack file meshctl uses,
// rendered with example values, to read without a server at hand.
// TestComposeDoc keeps it equal to the renderers (./scripts/dev.sh regen).
var composeDoc = filepath.Join("..", "docs", "compose.md")

// The example mesh: the default CIDR, its zone, a release and an origin.
const (
	exampleServer  = "https://mesh.example.com"
	exampleVersion = "v1.0.0"
	exampleMesh    = "10.77.0.0/16"
)

var exampleResolvers = []string{"10.77.0.1", "10.77.0.2", "10.77.0.3"}

// exampleContainer is a container-mode node at the /24 that starts at
// subnet, the first ones the server hands out.
func exampleContainer(node, subnet string) ContainerParams {
	base := strings.TrimSuffix(subnet, ".0/24")
	return ContainerParams{Node: node, Server: exampleServer, Version: exampleVersion, ListenPort: 51820, Network: exampleMesh,
		Subnet: subnet, Range: base + ".128/25", Bridge: base + ".1", Gateway: base + ".2", Zone: "mesh"}
}

func renderComposeDoc(t *testing.T) string {
	t.Helper()
	var b strings.Builder
	w := func(format string, args ...any) { fmt.Fprintf(&b, format, args...) }
	yamlBlock := func(body []byte) { w("```yaml\n%s```\n", strings.TrimRight(string(body), "\n")+"\n") }
	file := func(name string) []byte {
		body, err := os.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		return body
	}
	nas := func(backend string) []byte {
		body, err := RenderNASCompose(NASParams{Node: "nas", Server: exampleServer, Version: exampleVersion, Backend: backend})
		if err != nil {
			t.Fatal(err)
		}
		return body
	}
	container := func(p ContainerParams) []byte {
		body, err := RenderContainerCompose(p)
		if err != nil {
			t.Fatal(err)
		}
		return body
	}

	w(`# Compose and stack files

Every compose and stack file meshctl uses, in one place. The server's are the files in
[deploy/](../deploy); an agent's are what the server renders for a node, shown here for example
nodes of the mesh %s (zone `+"`mesh`"+`, server %s, release %s). A node's own file has its
name, its container subnet and the server's release in place of these: get it with
`+"`meshctl node compose <name>`"+` (with `+"`--runtime docker|podman|swarm`"+` in container mode) or from
the node's page in the web UI. Generated from the renderers by `+"`./scripts/dev.sh regen`"+`
(`+"`TestComposeDoc`"+`); do not edit.

| Node | Runs as | File |
| --- | --- | --- |
| the server, Docker or Dokploy compose app | a container on Traefik's network | [Server: Docker Compose](#server-docker-compose) |
| the server, Docker Swarm or Dokploy stack | one Swarm task, pinned to a labelled node | [Server: Docker Swarm stack](#server-docker-swarm-stack) |
| a Linux host (VPS, laptop) | the meshctl binary under systemd | no compose: `+"`curl -fsSL %s/install.sh \\| sudo sh -s -- --token <token>`"+` |
| a NAS or any spoke on the host network | a container with the host's network | [Client: host network](#client-host-network-nas) |
| a host whose containers only are on the mesh | the agent's container, their gateway | [Container mode](#container-mode) |
| a host-mode node's containers (Swarm services on the VPSes) | the host's own agent | [A host-mode node's containers](#a-host-mode-nodes-containers) |

## Server: Docker Compose

`+"`deploy/compose.server.yml`"+`: a Dokploy compose app, or `+"`docker compose up -d`"+` with the variables in
its header set.

`, exampleMesh, exampleServer, exampleVersion, exampleServer)
	yamlBlock(file("compose.server.yml"))
	w(`
## Server: Docker Swarm stack

` + "`deploy/stack.server.yml`" + `: ` + "`docker stack deploy`" + ` or a Dokploy stack app, one task on the node labelled
` + "`meshctl.server=true`" + `.

`)
	yamlBlock(file("stack.server.yml"))
	w(`
## Client: host network (NAS)

A spoke whose agent runs in a container on the host's network (` + "`meshctl node compose nas`" + `): the
whole host is on the mesh. Userspace WireGuard (the default; needs /dev/net/tun):

`)
	yamlBlock(nas(NASUserspace))
	w(`
The kernel's WireGuard, where the host has the module (` + "`--wg-backend kernel`" + `):

`)
	yamlBlock(nas(NASKernel))
	w(`
## Container mode

A node added with ` + "`--mode container`" + `: only the containers on its ` + "`meshctl-mesh`" + ` network are on the
mesh, and the agent's container is their gateway. Each file's header has the commands that create
that network once, the join, and how a service attaches.

### Container mode: a client on Docker

`)
	yamlBlock(container(exampleContainer("box", "10.77.128.0/24")))
	hub := exampleContainer("hub", "10.77.129.0/24")
	hub.Hub, hub.PublishPort = true, 51820
	w(`
### Container mode: a server on Docker

A server publishes its WireGuard port on the host, which its endpoint names (` + "`node add hub --role server --mode container --endpoint <ip>:51820`" + `):

`)
	yamlBlock(container(hub))
	podman := exampleContainer("pbox", "10.77.130.0/24")
	podman.Runtime = RuntimePodman
	w(`
### Container mode: Podman

Rootful or rootless Podman (` + "`--runtime podman`" + `); a rootless one can only be a client:

`)
	yamlBlock(container(podman))
	swarm := exampleContainer("sbox", "10.77.131.0/24")
	swarm.Runtime, swarm.Resolvers = RuntimeSwarm, exampleResolvers
	w(`
### Container mode: a Docker Swarm node

The agent itself is a plain compose container on each node (a Swarm service cannot have
/dev/net/tun or the capabilities it needs); Swarm services attach to a swarm-scoped network
built from each node's config-only one, and resolve through the servers (` + "`--runtime swarm`" + `):

`)
	yamlBlock(container(swarm))
	w(`
## A host-mode node's containers

A host-mode node with ` + "`--containers`" + ` keeps the host on the mesh and gives its containers mesh
addresses too: the host's own agent routes them on the bridge ` + "`meshctl0`" + `. Create the network once
on the host, then attach services (` + "`meshctl node network <name> --runtime ...`" + `). For a server at
10.77.132.0/24:

`)
	for _, rt := range []string{RuntimeDocker, RuntimeSwarm} {
		setup, err := MeshNetworkSetup(NetworkParams{Runtime: rt, Host: true, Network: exampleMesh, Range: "10.77.132.128/25",
			Bridge: "10.77.132.1", Gateway: "10.77.132.1", Zone: "mesh", Resolvers: exampleResolvers})
		if err != nil {
			t.Fatal(err)
		}
		w("%s:\n\n```sh\n%s\n```\n\n```yaml\n%s\n```\n\n", map[string]string{RuntimeDocker: "Docker", RuntimeSwarm: "Swarm services"}[rt],
			strings.Join(setup.Commands, "\n"), strings.Join(setup.Service, "\n"))
	}
	return strings.TrimRight(b.String(), "\n") + "\n"
}

func TestComposeDoc(t *testing.T) {
	got := renderComposeDoc(t)
	if *update {
		if err := os.WriteFile(composeDoc, []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(composeDoc)
	if err != nil {
		t.Fatal(err)
	}
	if got != string(want) {
		gl, wl := strings.Split(got, "\n"), strings.Split(string(want), "\n")
		for i := 0; i < len(gl) || i < len(wl); i++ {
			var g, w string
			if i < len(gl) {
				g = gl[i]
			}
			if i < len(wl) {
				w = wl[i]
			}
			if g != w {
				t.Fatalf("%s is out of date at line %d (run ./scripts/dev.sh regen):\n got: %q\nwant: %q", filepath.Base(composeDoc), i+1, g, w)
			}
		}
	}
}

// The Swarm stack states the posture a stack can: non-root, no
// capabilities, read-only root, one task pinned to a labelled node and
// stopped before the next starts; nothing that widens it.
func TestServerStackPosture(t *testing.T) {
	doc, svc := service(t, "stack.server.yml")
	if got, want := keys(svc), []string{"image", "command", "environment", "user", "cap_drop", "read_only", "volumes", "networks",
		"stop_grace_period", "healthcheck", "deploy"}; !sameSet(got, want) {
		t.Errorf("stack service keys = %v, want exactly %v", got, want)
	}
	for k, v := range map[string]any{"user": "65532:65532", "cap_drop": []any{"ALL"}, "read_only": true,
		"environment": []any{"TZ=${TZ:-UTC}"}} {
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
	if dep["mode"] != "replicated" || dep["replicas"] != 1 || !reflect.DeepEqual(place["constraints"], []any{"node.labels.meshctl.server == true"}) ||
		update["order"] != "stop-first" {
		t.Errorf("stack deploy = %#v", dep)
	}
	nets, _ := doc["networks"].(map[string]any)
	if dok, _ := nets["dokploy-network"].(map[string]any); dok["external"] != true {
		t.Error("dokploy-network must be external (Dokploy owns it)")
	}
	// The same command as the compose file: only how it is run differs.
	_, compose := service(t, "compose.server.yml")
	if !reflect.DeepEqual(svc["command"], compose["command"]) || !reflect.DeepEqual(svc["healthcheck"], compose["healthcheck"]) {
		t.Errorf("the stack runs %#v, the compose file %#v", svc["command"], compose["command"])
	}
}
