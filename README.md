# meshctl

A self-hosted, single-binary WireGuard mesh for a small fleet: three VPS servers that run
Docker Swarm over the mesh, a NAS behind CGNAT, laptops that pick which server their internet
traffic exits from, and phones that join by QR code. One binary is the server (source of truth,
web UI, agent API), the agent on every node, and the CLI.

meshctl builds and operates the private network only. It never configures Traefik, Dokploy,
Docker, Swarm or any application; it exposes each node's mesh IP and `<name>.mesh` hostname
for other tools to use.

**Status: pre-release.** Delivery follows the milestones in [docs/spec.md](docs/spec.md) S15.
M1 (server via CLI, agent join/run, kernel WireGuard, the server full mesh with its firewalld
zone) and M2 (the web UI with password + TOTP sign-in, versions and rollback, audit, users,
import/export) are implemented. M3 has Linux spokes, relaying, `meshctl exit`, safe apply, the
userspace WireGuard backend and the NAS container. M4 so far has the
`.mesh` resolver on every node with split DNS through systemd-resolved (`meshctl dns`), custom
DNS records in the web UI, the agent's drift check, and phones, with a profile
for the WireGuard app as a QR code (mesh only or via an exit), each node's cached view of the
mesh and `meshctl logs`, and each node's local UI (`meshctl ui`). Windows (M5) is written: the
agent installs as a Windows service with its local API on a named pipe, runs WireGuard on the
bundled wintun driver, and has split DNS (NRPT) and exit mode. Container mode (M6, in progress)
puts only a host's containers on the mesh: each node gets a container /24, and the agent runs in
a container as its containers' gateway on Docker, Podman or Swarm (`meshctl node add --mode
container`), and a host-mode node can put its containers on the mesh too (`--containers`); `node
show` gives the network's values. Every node resolves those containers by their compose
service, container name, aliases or hostname (`web.mesh` round robin across nodes, `web.nas.mesh`
one node's), which the agent reads from the runtime's socket, and extra DNS domains of your own (Settings) resolve the containers' `hostname.domainname`
across the mesh. When a hub goes down, the server moves the pairs it relays to another hub and
back once it returns (relay failover). The mesh firewall (M7, in progress) segments the mesh:
groups (`meshctl node add --groups`), zones, address and service objects and ordered policies,
enforced by every server in an nftables table of its own, edited on the web UI's Firewall page or
with `meshctl firewall`; a mesh starts with one `allow-all` policy. A policy may be scheduled (days
and a window, in the server's time zone or one you set), and every server switches at the edge by
its own clock. Dual stack (M8, in progress) runs the mesh on IPv4 and IPv6 by default: a unique
local IPv6 prefix of its own (`meshctl ipv6`), each node IPv4, IPv6 or both (`--family`), servers'
IPv6 endpoints with failover between the families, IPv6 through an exit that has it, the mesh
firewall over both, the containers' IPv6 when asked (`--containers6`), and a host without IPv6
carried on IPv4. Containers run from their environment: the image's `meshctl start` runs the
control plane with `MESHCTL_MODE=control` and a node's agent otherwise, and a node's container
joins on its first start with `MESHCTL_TOKEN` (exported, in a `.env` file, or a file) and again
only when given a new token. meshctl writes no container deployment: [docs/containers.md](docs/containers.md)
has compose, Podman and Kubernetes samples. One control plane holds several meshes (M9, in
progress): `meshctl mesh add|list|show|set|rm`, each with its own network, DNS zone, IPv6 prefix,
firewall and history, none sharing an address or a domain with another, `--mesh` on every
admin command, each mesh's web UI pages under `/m/<mesh>/` with a Meshes page, admins and
users granted meshes, and a host in two meshes running one agent per mesh, each on its slot's
interface, port, tables and firewall names. Configs, mesh views and refusals are signed by the
control plane and checked by every agent; every join makes fresh keys, and a new token cuts a
node off at once. A node runs once, on its own machine: a second copy is refused (it takes down
what it brought up), one on another machine wipes itself, and a restart carries on at once. The
live tests and the v1.0.0 release follow.

## Quick start

```
# on the Docker host running the control plane (deploy/compose.control.yml, or
# deploy/stack.control.yml on Swarm; docs/containers.md has samples for every container)
MESHCTL_VERSION=1.0.0 MESHCTL_URL=https://mesh.example.com docker compose up -d
docker exec <container> /app/meshctl node add vps-1 --role server
# on vps-1, as root: the command node add printed
curl -fsSL https://mesh.example.com/install.sh | sudo sh -s -- --token <token>
```

[docs/operations.md](docs/operations.md) walks through it, including Swarm over the mesh.

## How it fits together

- `meshctl control` (the control plane) runs in a container on one VPS behind your reverse proxy. It stores the mesh
  in SQLite, computes every node's peers, AllowedIPs and the servers' mesh firewall, and pushes changes to
  agents over server-sent events.
- `meshctl agent` runs on every node, including the server VPS. It generates its keys locally
  (private keys never leave the node), applies the config it is given idempotently, caches it
  on disk so the mesh survives a server outage, and reports status every 30 s.
- Every agent request is signed with the node's Ed25519 identity key; joining uses a one-time
  token that expires after an hour.

## Building

Everything goes through the dev scripts; see [docs/developer.md](docs/developer.md).

```
./scripts/dev.sh all scan     # build vet test + vulnerability scan (what CI runs)
./scripts/dev.sh dist         # cross-compile linux/amd64, linux/arm64, windows/amd64
```

## Documentation

| Doc | Contents |
| --- | --- |
| [docs/spec.md](docs/spec.md) | Design spec, with the Amendments table (S17) |
| [docs/operations.md](docs/operations.md) | Running the server, enrolling nodes, backups, troubleshooting |
| [docs/commands.md](docs/commands.md) | Every command and flag (generated) |
| [docs/agent-protocol.md](docs/agent-protocol.md) | The agent API: routes, signing, errors, SSE, apply steps |
| [docs/security.md](docs/security.md) | Keys, tokens, the exec gate, container and systemd posture |
| [docs/developer.md](docs/developer.md) | Toolchain, dev-script tasks, goldens, releases, extending the code |
| [docs/test.md](docs/test.md) | Catalogue of tests, fixtures and seams |
