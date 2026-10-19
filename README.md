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
userspace WireGuard backend and the NAS container (`meshctl node compose`). M4 so far has the
`.mesh` resolver on every node with split DNS through systemd-resolved (`meshctl dns`), custom
DNS records and firewall rules in the web UI, the agent's drift check, and phones, with a profile
for the WireGuard app as a QR code (mesh only or via an exit), each node's cached view of the
mesh and `meshctl logs`, and each node's local UI (`meshctl ui`). Windows (M5) is under way:
the agent installs as a Windows service with its local API on a named pipe and runs WireGuard on
the bundled wintun driver; exit mode and split DNS there are next.

## Quick start

```
# on the Docker host running the server (deploy/compose.server.yml)
docker exec <container> /app/meshctl node add vps-1 --role server
# on vps-1, as root: the command node add printed
curl -fsSL https://mesh.example.com/install.sh | sudo sh -s -- --token <token>
```

[docs/operations.md](docs/operations.md) walks through it, including Swarm over the mesh.

## How it fits together

- `meshctl server` runs in a container on one VPS behind your reverse proxy. It stores the mesh
  in SQLite, computes every node's peers, AllowedIPs and firewall rules, and pushes changes to
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
