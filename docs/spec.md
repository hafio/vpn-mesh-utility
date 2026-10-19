# meshctl — Design Spec

**Status:** Draft v0.2 -- 12 Oct 2026 (v0.1 plus the Amendments in S17) **Scope:** A self-hosted, single-binary WireGuard mesh utility with a web UI. It builds and operates the private network only. It does not publish, proxy or deploy any application.

---

## 1. Problem

Three VPS nodes need a private network for Docker Swarm. A home NAS behind CGNAT, a Windows laptop, a Linux laptop and phones need to join that network, and the laptops need to pick which VPS their internet traffic exits from. Today this means hand-written `wg0.conf` files on every node, manual key exchange, manual firewall rules and no single view of what is connected.

Existing tools either add a control plane you do not own (Tailscale), bring their own reverse proxy that clashes with Dokploy's Traefik (Pangolin, NetBird), or are moving too fast to trust for the Swarm base. The hard part those tools solve, NAT traversal between two NAT'd peers, is not needed here: every server has a public IP, and every other device dials out.

## 2. Goals

1. **One binary, every role.** The same `meshctl` binary runs the server, every agent and the CLI, on Linux amd64/arm64 and Windows amd64.
2. **Zero hand-edited configs.** Add a node in the UI, run one join command on it, and it is in the mesh within 60 seconds.
3. **Server outage never breaks the mesh.** Agents keep their last applied config on disk. Only changes need the server.
4. **One view of the whole mesh.** Any node's UI shows every node's handshake age, traffic and applied config version.
5. **Exit switching in one click.** A laptop can switch its internet exit between servers, or turn it off, without reconnecting.

## 3. Non-goals

- **No app publishing or deployment.** meshctl does not configure Traefik, Dokploy, Docker, Swarm or Immich. It only exposes each node's mesh IP and `.mesh` hostname for other tools to use.
- **No NAT traversal between spokes.** Two NAT'd devices talk via a server. Hole punching and relays are a separate project.
- **No IPv6 inside the mesh in v1.** The schema reserves fields for it.
- **No SSO/OIDC in v1.** Built-in user + TOTP only. The auth layer is an interface so OIDC can be added later.
- **No per-user ACLs.** Access control is per node via AllowedIPs and rendered firewall rules.
- **No mobile agent.** Phones use the official WireGuard app via QR.
- **No automatic key rotation.** A manual "rotate key" action is P1.

## 4. Roles and topology

| Role | Examples | Public IP | Listens | Peers with | Can be exit |
| --- | --- | --- | --- | --- | --- |
| `server` | VPS nodes | Yes | UDP 51820 | All servers (full mesh), all spokes | Yes |
| `spoke` | NAS, laptops | No | No fixed port | All servers | No |
| `phone` | iOS/Android | No | No | All servers (static profile) | No |

**Rules the server applies when it computes peer lists:**

- Servers form a full mesh. Each server's AllowedIPs for another server is that server's `/32`.
- Every spoke peers directly with every server. Servers have public endpoints, so no relay is needed for spoke-to-server traffic.
- Spoke-to-spoke traffic relays through the spoke's **primary hub** (a server chosen per spoke, default: lowest ID). Only the primary hub carries the other spokes' `/32`s in the spoke's AllowedIPs, because WireGuard allows a given IP in only one peer's AllowedIPs.
- A spoke with an **exit** set gets `0.0.0.0/0` added to that server's AllowedIPs, and the agent installs policy routing (see §7.3).
- Spokes set `PersistentKeepalive = 25` toward every server. Servers set no keepalive toward each other.
- Every node's peer set, AllowedIPs, endpoint and keepalive are derived by the server. Agents never edit peer lists locally.

**Address plan (v1 default, configurable at setup):**

| Range | Use |
| --- | --- |
| `10.77.0.0/16` | Mesh |
| `10.77.0.1–254` | Servers |
| `10.77.1.0/24` | NAS and other always-on spokes |
| `10.77.2.0/24` | Laptops |
| `10.77.3.0/24` | Phones |

## 5. Architecture

```
                 Internet
                    │
      ┌─────────────┼─────────────────────┐
      │  VPS-1 (Dokploy host)             │
      │  Traefik ──► meshctl server :8080  │   UI (user+TOTP)  ◄── you, browser
      │              SQLite /data/mesh.db  │   Agent API       ◄── all agents, signed HTTPS
      │  meshctl agent ──► wg0 10.77.0.1   │
      └───────────────────────────────────┘
            ▲ SSE + heartbeat (HTTPS, outbound only)
   ┌────────┴────────┬──────────────┬───────────────┐
 VPS-2 agent     VPS-3 agent    NAS agent        Laptop agent
 wg0 10.77.0.2   wg0 10.77.0.3  (container,      (Windows service /
                                 host net)        Linux systemd)
                                 10.77.1.1        10.77.2.1
```

**Components inside the one binary:**

| Subcommand | Runs where | Responsibility |
| --- | --- | --- |
| `meshctl server` | One VPS, as a container behind Traefik | Source of truth, UI, agent API, config computation, join tokens, DNS zone, audit log |
| `meshctl agent` | Every node (including the server VPS) | Enroll, pull config, apply WireGuard + routes + firewall + DNS, report status, serve local UI |
| `meshctl exit / status / dns / logs` | Any node | Thin CLI over the local agent's socket |
| `meshctl user / export / import` | Server | Admin tasks |

**Why the server VPS also runs an agent:** the server process only computes config. The agent on the same host applies it, exactly like every other node. This keeps one code path.

## 6. Control plane

### 6.1 Enrollment

1. Admin creates a node in the UI: name, role, optional fixed mesh IP, exit capability, primary hub.
2. UI shows a one-time join token (single use, expires in 1 hour) and the exact join command.
3. On the node: `meshctl agent join --server https://mesh.example.com --token <token>`
4. The agent generates two keys locally and stores them in its data dir:
   - Ed25519 identity key, used to sign every API request.
   - WireGuard private key (Curve25519). Neither private key ever leaves the node.
5. The agent POSTs both public keys, hostname, OS, arch, detected WireGuard support (kernel/userspace) and, for servers, the public endpoint it detected.
6. The server validates the token, stores the node, assigns the mesh IP, and returns `node_id` and `config_version`.
7. The agent installs itself as a service (`systemd` unit or Windows service) and starts `meshctl agent run`.

### 6.2 Request authentication

All agent requests carry:

```
X-Mesh-Node:      <node_id>
X-Mesh-Timestamp: <unix seconds>
X-Mesh-Signature: base64(ed25519_sign(method + "\n" + path + "\n" + timestamp + "\n" + sha256(body)))
```

The server rejects timestamps outside ±60 s and caches recent signatures to block replay. Agent endpoints never use cookies, so the browser session and agent auth cannot be confused.

### 6.3 Config sync

- `GET /api/v1/agent/events` — long-lived SSE stream. The server emits `config` events with the new version number whenever this node's computed config changes.
- `GET /api/v1/agent/config` — full desired state for this node:

```json
{
  "version": 42,
  "interface": { "address": "10.77.2.1/32", "listen_port": null, "mtu": 1420, "fwmark": 51820 },
  "peers": [
    { "name": "vps-1", "public_key": "...", "endpoint": "203.0.113.10:51820",
      "allowed_ips": ["10.77.0.1/32", "10.77.1.0/24", "10.77.2.0/24", "0.0.0.0/0"],
      "keepalive": 25 }
  ],
  "exit": { "via": "vps-1" },
  "dns": { "zone": "mesh", "records": [ { "name": "nas", "ip": "10.77.1.1" } ],
           "upstream": ["1.1.1.1", "9.9.9.9"], "split": true },
  "firewall": { "backend": "firewalld", "rules": [ ... ] }
}
```

- Agents apply idempotently and record `applied_version`.
- `POST /api/v1/agent/status` every 30 s (and immediately after apply): applied version, per-peer last handshake, rx/tx bytes, current endpoint seen, errors, agent version.
- On disconnect the agent retries with backoff (1 s → 60 s). The last config is cached at `<data>/config.json` and re-applied on boot without contacting the server.

### 6.4 Safe apply

For `server` nodes only: after applying a new config, the agent must receive a successful status response within 60 s. If not, it reverts to the previous config and reports `rollback`. This prevents a bad change from locking you out of a VPS.

## 7. Agent

### 7.1 WireGuard backend

- Linux: kernel WireGuard via `wgctrl-go` (netlink). Detected by creating a test interface.
- Fallback (TerraMaster without the module, Windows): userspace `wireguard-go` compiled into the binary. Linux needs `/dev/net/tun`; Windows uses the bundled `wintun.dll`.
- Interface name: `wg0` on Linux, `meshctl` on Windows.
- Config changes go through `wgctrl` `ConfigureDevice` with `ReplacePeers=false` plus explicit removals, so existing handshakes survive.

### 7.2 Routes

- Each peer's AllowedIPs are installed as routes on the interface (what `wg-quick` does), excluding `0.0.0.0/0`.
- MTU default 1420. Per-node override in the UI (TerraMaster on PPPoE may need 1380).

### 7.3 Exit switching (spokes)

- **On:** add `0.0.0.0/0` to the exit server's AllowedIPs, set interface fwmark 51820, add `ip rule add not fwmark 51820 table 51820`, `ip rule add table main suppress_prefixlength 0`, `ip route add default dev wg0 table 51820`. Windows: add `0.0.0.0/1` and `128.0.0.0/1` via the tunnel and a host route to the endpoint via the physical gateway.
- **Switch:** move `0.0.0.0/0` from one peer's AllowedIPs to another. Routes stay. No reconnect.
- **Off:** remove `0.0.0.0/0` and the policy rules. Normal internet returns. meshctl is not a kill switch.
- In exit mode the agent's DNS resolver forwards upstream queries through the tunnel, so DNS does not leak.
- Exit servers get `net.ipv4.ip_forward=1` and firewalld masquerade on their public zone, rendered by the server.

### 7.4 DNS

- Each agent runs a resolver (`miekg/dns`) bound to its own mesh IP, UDP+TCP 53, serving the `mesh.` zone from the synced record list.
- Automatic records: one per node (`vps-1.mesh`, `nas.mesh`). Custom records: any name → mesh IP, managed in the server UI.
- Non-mesh queries are forwarded to the configured upstream only when `split` is false (phones, exit mode). Otherwise the resolver answers `REFUSED`.
- Split-DNS registration, so nothing else on the host changes:
  - Linux: `resolvectl dns wg0 <mesh-ip>` and `resolvectl domain wg0 ~mesh` via the systemd-resolved D-Bus API. If resolved is absent, DNS is reported as "unavailable" in the UI; `/etc/hosts` and `/etc/resolv.conf` are never written.
  - Windows: NRPT rule for `.mesh` → agent IP, removed on stop.
  - Containers on the host do not see split DNS. The UI states this on the DNS page.

### 7.5 Firewall rendering

- `server` nodes (AlmaLinux): firewalld via `firewall-cmd` (D-Bus in a later version). The agent owns a zone named `mesh` bound to `wg0`:
  - Swarm ports 2377/tcp, 7946/tcp+udp, 4789/udp allowed only from server `/32`s.
  - 53/udp+tcp and the local UI port allowed from the mesh range.
  - Default policy for the zone: drop.
  - Public zone: 51820/udp, plus masquerade if the node is an exit.
- `spoke` nodes: no system firewall changes. The agent binds every listener to the mesh IP or loopback only. The UI shows a reminder for the NAS to allow its app ports from `10.77.0.0/16` in the TOS firewall.
- Rules are rendered from the same data as peers, so removing a node removes its rules on the next apply.

### 7.6 Local UI

- Binds `127.0.0.1:7443` only. Optional: also on the mesh IP, protected by the server's session (P1).
- Shows the full mesh dashboard (data fetched from the server and cached, so it works read-only while the server is down).
- Local edits only: exit selection, keepalive, DNS on/off, MTU override request, view logs, restart tunnel.
- The CLI (`meshctl exit vps-2`, `meshctl status`) talks to the same local API over a Unix socket / named pipe.

### 7.7 Platform packaging

| Platform | Runs as | Install |
| --- | --- | --- |
| AlmaLinux VPS | systemd unit `meshctl-agent.service`, root | `install.sh` downloads the binary, runs `join`, enables the unit |
| TerraMaster NAS | Docker container, `network_mode: host`, `cap_add: NET_ADMIN`, `/dev/net/tun`, volume `/data` | Compose file generated by the server UI |
| Linux laptop | systemd unit | `install.sh` |
| Windows laptop | Windows service via `x/sys/windows/svc`, bundled `wintun.dll` | `meshctl agent install --server … --token …` (elevated prompt) |

## 8. Server

### 8.1 Storage

SQLite (`modernc.org/sqlite`, no cgo) at `/data/mesh.db`. WAL mode. Nightly `meshctl export` to YAML into `/data/exports/` for commit to git.

### 8.2 Data model

| Table | Key fields |
| --- | --- |
| `nodes` | id, name, role, os, arch, wg_backend, mesh_ip, wg_pubkey, identity_pubkey, endpoint, listen_port, primary_hub_id, can_exit, exit_via, keepalive, mtu, applied_version, last_seen, created_at, disabled |
| `join_tokens` | token_hash, node_id, expires_at, used_at |
| `dns_records` | id, name, ip, node_id (null for custom), created_at |
| `firewall_rules` | id, node_id (null = all servers), proto, port, source_cidr, note |
| `peer_status` | node_id, peer_id, last_handshake, rx_bytes, tx_bytes, endpoint, updated_at |
| `config_versions` | version, created_at, user_id, summary, snapshot_json |
| `users` | id, username, password_hash (argon2id), totp_secret (encrypted), created_at |
| `sessions` | id, user_id, expires_at, csrf |
| `audit_log` | id, at, user_id, action, target, detail |
| `settings` | key, value (mesh_cidr, server_url, upstream_dns, default_hub, default_mtu) |

Every write that changes any node's computed config creates a `config_versions` row. The UI can diff and roll back to any version, which simply becomes a new version.

### 8.3 Web UI

Server-rendered Go templates + HTMX, embedded with `go:embed`. No build step, no CDN.

| Page | Contents |
| --- | --- |
| **Dashboard** | All nodes: role, mesh IP, handshake age to each peer, rx/tx, applied vs current version, agent version, exit in use. Red/amber/green per node. |
| **Nodes** | Add (role, name, IP, hub, exit capability), disable, remove, rotate key (P1), show join command, generate NAS compose file, generate phone QR (mesh-only and via-exit profiles). |
| **Node detail** | Everything the server computed for this node, read-only, plus "edit" for keepalive, MTU, hub, exit. |
| **DNS** | Automatic and custom records, upstream servers. |
| **Firewall** | Rendered rules per server, with the source for each rule. |
| **Versions** | Timeline of config versions, diff, roll back. |
| **Settings** | Mesh CIDR, server URL, users and TOTP enrollment, export/import. |
| **Audit** | Who changed what, when. |

### 8.4 Authentication

- Username + password (argon2id) + TOTP (RFC 6238), enrolled by QR on first login.
- Session cookie: `HttpOnly`, `Secure`, `SameSite=Strict`, 12 h idle expiry. CSRF token on every form.
- Rate limit: 5 failed logins per 15 min per IP, then 15 min lockout.
- Runs behind Dokploy's Traefik for TLS. The existing GeoIP allowlist can sit in front of the UI path; the agent path `/api/v1/agent/*` must stay reachable from every node's country.
- `meshctl user add` / `meshctl user reset-totp` from the server container for recovery.

## 9. Security model

| Asset | Protection |
| --- | --- |
| Node private keys | Generated on the node, stored `0600` in the data dir, never transmitted. |
| Agent identity | Ed25519 signed requests with timestamp and replay cache. |
| Join tokens | 32 random bytes, stored hashed, single use, 1 h expiry. |
| Server compromise | Attacker can add a rogue peer or change routes, but cannot read node keys or decrypt traffic. Mitigations: TOTP, audit log, versions with rollback, safe-apply on servers, optional restriction of the UI to the mesh after bootstrap. |
| NAS compromise | Its AllowedIPs on every server are its own `/32`; firewall drops its traffic to Swarm ports. |
| Lost laptop | Disable the node in the UI; every server drops its peer on the next apply (seconds). |
| Local UI | Loopback only; OS user boundary is the auth boundary. |

## 10. Requirements

### P0 — must ship in v1

- [ ] Single Go binary; builds for `linux/amd64`, `linux/arm64`, `windows/amd64`.
- [ ] `server`: SQLite, UI with user + TOTP, node CRUD, join tokens, config computation per §4, SSE push, heartbeat intake.
- [ ] `agent`: join, run as service, kernel + userspace WireGuard, idempotent apply, cached config on disk, status reporting.
- [ ] Server full mesh suitable for Swarm: `/32` AllowedIPs, no keepalive, MTU 1420, firewalld zone with Swarm ports restricted to servers.
- [ ] Spokes: NAS container and laptops join, reach all servers, reach each other via primary hub.
- [ ] Exit switching on Linux and Windows: on / switch / off without reconnect, DNS via tunnel in exit mode.
- [ ] DNS: `.mesh` zone, automatic + custom records, split-DNS via systemd-resolved and NRPT, never writes `/etc/hosts`.
- [ ] Phone QR: mesh-only and via-exit profiles.
- [ ] Add/remove/disable nodes from the server UI; removal propagates to every node.
- [ ] Local UI on every node: full read-only mesh view, local-only edits.
- [ ] Config versions with diff and rollback; audit log.
- [ ] Safe-apply rollback on server nodes.

**Acceptance (sample):**

- Given a new VPS, when I paste the join command, then it appears green on the dashboard within 60 s and `wg show` on an existing server lists it with a recent handshake.
- Given a laptop with exit off, when I click "Exit via vps-2", then `curl ifconfig.me` returns vps-2's IP within 5 s and no existing SSH session over the mesh drops.
- Given the server container is stopped, when any node reboots, then it rejoins the mesh from its cached config and `nas.mesh` still resolves on the laptop.
- Given a node is disabled in the UI, when 30 s pass, then no server has it in `wg show` and its mesh IP is unreachable.

### P1 — fast follow

- [ ] Rotate WireGuard key per node from the UI (overlap old and new key for 60 s).
- [ ] Local UI reachable on the mesh IP with server session.
- [ ] firewalld via D-Bus instead of `firewall-cmd` exec.
- [ ] Prometheus `/metrics` on the server.
- [ ] Email/webhook alert when a server node misses 3 heartbeats.
- [ ] Multiple exit groups ("EU", "Asia") with latency display in the exit picker.

### P2 — design for, do not build

- IPv6 inside the mesh (`mesh_ip6` column reserved).
- OIDC login (auth is behind an interface).
- NAT traversal between spokes (peer status already records observed endpoints).
- Multiple meshes per server (`mesh_id` column reserved, single value in v1).

## 11. CLI reference

```
meshctl server [--data /data] [--listen :8080] [--url https://mesh.example.com]
meshctl node add <name> --role server|spoke|phone [--ip] [--endpoint] [--hub] [--can-exit] [--always-on]
meshctl node list | show <name> | set <name> [flags] | disable <name> | enable <name> | rm <name> | token <name>
meshctl user add <name> | reset-totp <name>
meshctl export [--out mesh.yaml] | import mesh.yaml

meshctl agent join --server <url> --token <token> [--name <name>] [--endpoint IP:PORT] [--no-service]
meshctl agent run | install | uninstall
meshctl status            # peers, handshake age, exit, applied version
meshctl exit <node>|off   # spokes only
meshctl dns on|off
meshctl logs [-f]
meshctl ui                # one-time login URL for the local UI
meshctl healthcheck server|agent
meshctl version
```

`node` commands talk to the running server over `<data>/admin.sock`; `user` and `export` open
the database directly so they work with the server stopped; `import` requires the server to be
stopped (Amendments A9).

## 12. Deployment

**Server (Dokploy compose app):**

```yaml
services:
  meshctl:
    image: <registry>/meshctl:latest
    command: ["server", "--data", "/data"]
    volumes: ["meshctl-data:/data"]
    ports: ["127.0.0.1:8080:8080"]   # Traefik routes mesh.example.com → :8080
    restart: unless-stopped
volumes:
  meshctl-data: {}
```

Dokploy owns the domain, TLS and Traefik routing. meshctl needs nothing else.

**NAS (generated by the UI):**

```yaml
services:
  meshctl-agent:
    image: <registry>/meshctl:latest
    command: ["agent", "run"]
    network_mode: host
    cap_add: [NET_ADMIN]
    devices: ["/dev/net/tun:/dev/net/tun"]
    volumes: ["/volume1/docker/meshctl:/data"]
    restart: unless-stopped
```

First run: `docker compose run --rm meshctl-agent agent join --server … --token …`.

**VPS and Linux laptop:** `curl -fsSL https://mesh.example.com/install.sh | sudo sh -s -- --token <token>` installs the binary to `/usr/local/bin`, writes the systemd unit, joins, starts.

**Windows:** download `meshctl.exe`, run `meshctl agent install --server … --token …` from an elevated prompt.

## 13. Build

- Go 1.27 (`go 1.27`, `toolchain go1.27.2` in go.mod), `CGO_ENABLED=0` for every shipped binary.
- Key modules: `golang.zx2c4.com/wireguard/wgctrl`, `golang.zx2c4.com/wireguard` (userspace;
  its `tun` package brings in `golang.zx2c4.com/wintun` on Windows),
  `golang.zx2c4.com/wireguard/windows/tunnel/winipcfg`, `github.com/vishvananda/netlink`,
  `github.com/miekg/dns`, `modernc.org/sqlite`, `github.com/pquerna/otp`,
  `golang.org/x/crypto/argon2`, `github.com/godbus/dbus/v5`, `golang.org/x/sys/windows/svc`,
  `github.com/Microsoft/go-winio`, `github.com/boombuler/barcode` (A23, A30), `go.yaml.in/yaml/v3` (A20).
- Repo layout:

```
cmd/meshctl/          main, subcommand wiring
internal/proto/       request signing, config schema, SSE (shared by server and agent)
internal/validate/    the input-validation gate (shared)
internal/server/      api, ui (handlers, templates and static assets, go:embed), store, planner, auth
internal/agent/       wg (kernel, userspace), routes, exitmode, dns, firewall, localui, platform/{linux,windows}
internal/testfix/     shared test fixtures
internal/archtest/    module-boundary test
deploy/               compose examples, systemd unit, install.sh template
Dockerfile            multi-stage image build (repo root: the release workflow expects it there)
scripts/              dev.sh / dev.ps1 -- the only place that knows how to build, test and scan
```

- `./scripts/dev.sh build` (or `dev.ps1`) compiles `dist/meshctl-<os>-<arch>[.exe]` with the
  version stamped via `-ldflags "-X main.version=..."`; `dist` cross-compiles all three targets.
- CI is GitHub Actions, triggered by a `v*` tag only (`.github/workflows/tag.yml`): gates on
  ubuntu and windows, a multi-arch image on GHCR (`ghcr.io/hafio/vpn-mesh-utility`), the
  binaries, then the release. Workflows call dev-script task names only. See
  [developer.md](developer.md).

## 14. Swarm notes (informational, not managed by meshctl)

meshctl does not run Docker commands. For reference, Swarm should bind to the mesh:

```
docker swarm init --advertise-addr 10.77.0.1 --listen-addr 10.77.0.1 --data-path-addr 10.77.0.1
docker network create -d overlay --opt com.docker.network.driver.mtu=1370 <name>
```

The rendered firewall rules already restrict Swarm ports to server peers.

## 15. Milestones

| Milestone | Delivers |
| --- | --- |
| **M1 Core** | server (no UI, CLI only) + agent join/run/apply, kernel WireGuard, server full mesh, cached config. Swarm can run on it. |
| **M2 UI + auth** | Dashboard, nodes, versions, user + TOTP, audit. |
| **M3 Spokes + exit** | NAS container (userspace fallback), Linux laptop, exit on/switch/off, safe-apply. |
| **M4 DNS + firewall + phones** | Resolver, split-DNS, firewalld zone rendering, QR profiles, local UI. |
| **M5 Windows** | Service, wintun, NRPT, exit routing. Release v1.0. |

## 16. Open questions

| Question | Blocking? | Owner |
| --- | --- | --- |
| Does the TerraMaster kernel ship the WireGuard module? (`ls /sys/module/wireguard`) Decides whether userspace fallback is needed in M3. | No | Owner |
| Keep 2 hubs per spoke or peer spokes to all servers (current plan: all servers)? | No | Owner |
| Should the server UI be restricted to the mesh after bootstrap, with the public path serving only `/api/v1/agent/*`? | No | Owner |
| Exit-mode DNS: forward to upstream via tunnel (plan) or to the exit server's resolver? | No | Owner |

## 17. Amendments

Deliberate deviations from draft v0.1, decided 12 Oct 2026 before implementation. Where a
section above has been edited in place, the row says so; otherwise this table overrides the
text above.

| # | Amendment | Why |
| --- | --- | --- |
| A1 | Build and CI use mirrored `scripts/dev.sh` / `dev.ps1` and tag-triggered GitHub Actions instead of a Makefile; Go 1.27 with an exact toolchain pin; the Dockerfile sits at the repo root. S13 edited. | House conventions; the release workflow expects a root Dockerfile. |
| A2 | Add `vishvananda/netlink` (links, addresses, routes, policy rules), `Microsoft/go-winio` (named pipe) and `wireguard/windows/tunnel/winipcfg`; wintun comes in through wireguard-go's `tun` package. S13 edited. | wgctrl configures devices but cannot create links or routes; the v0.1 wintun import path does not exist. |
| A3 | Spoke-to-spoke traffic uses one relay per pair: relay(A, B) is the primary hub of the phone if either is a phone, else the primary hub of the lower-id node, and both sides route the other's /32 to it. A phone's hub peer carries the whole mesh CIDR. Phone-to-phone works only through the same hub. Replaces the S4 rule "only the primary hub carries the other spokes' /32s". | WireGuard drops a packet whose source is not in the AllowedIPs of the peer it arrived from, so per-spoke hubs drop traffic between spokes with different hubs. |
| A4 | Every relaying server gets `net.ipv4.ip_forward=1` and a firewalld policy for mesh-to-mesh forwarding; exit-capable servers also get the mesh-to-public policy with masquerade. Extends S7.3/S7.5. | The mesh zone's DROP target blocks forwarding, and v0.1 only enabled forwarding on exits. |
| A5 | The minimal firewalld rendering (public 51820/udp; mesh zone with the Swarm ports from server /32s) moves into M1. Extends S15. | On AlmaLinux neither WireGuard nor Swarm works over wg0 without it, so "Swarm can run on it" needs it. |
| A6 | Phone keypairs are generated by the server when the QR is shown; the QR is shown once and only the public key is stored. Exception to S9 "never transmitted". | Phones have no agent. |
| A7 | The CLI reaches the agent over a Unix socket (0660, group `meshctl`) or a named pipe (Administrators + SYSTEM). The browser UI on 127.0.0.1:7443 requires a one-time token from `meshctl ui`, a Host allowlist, an Origin check and CSRF tokens. Replaces S9 "OS user boundary is the auth boundary" for the loopback UI. | Loopback TCP is reachable by every local user and by DNS rebinding. |
| A8 | The signature covers the path with its raw query, a per-request random nonce (`X-Mesh-Nonce`) and hex(sha256(body)); the replay cache is in memory, 120 s, size-capped and refuses rather than evicts when full; a skew error returns the server time. Refines S6.2. | Closes a query-tampering gap. Ed25519 signatures are deterministic, so without the nonce two identical requests in one second would look like a replay. |
| A9 | Admin CLI `meshctl node ...` over `<data>/admin.sock` (0600); `meshctl healthcheck` and `meshctl ui`; `user` and `export` work on the database directly; `import` is offline-only. S11 edited. | M1 has no UI, and one writer must bump versions and push SSE. |
| A10 | Exit, keepalive and MTU edits from a node go to the server (signed `POST /api/v1/agent/prefs`); `dns on/off` stays local. | The server owns AllowedIPs; consistent with goal 3 (only changes need the server). |
| A11 | Windows exit mode binds the outer UDP socket to the default physical interface with a route monitor (fallback: host routes to every server endpoint); exit mode adds an IPv6 blackhole on both OSes. Refines S7.3. | v0.1 routed only one endpoint around the tunnel, and IPv6 would leak. |
| A12 | Exit-mode DNS: systemd-resolved `~.` plus default-route on wg0 (split mode registers only `~mesh`); NRPT `.` on Windows. The resolver forwards non-mesh queries in exit mode or for phone sources (`forward_for`), otherwise answers REFUSED. Refines S7.4. | Prevents leaks; phones use a server's resolver. |
| A13 | Server compose uses `expose` and the external `dokploy-network` instead of a 127.0.0.1 port; both compose files emit their full posture. The NAS agent runs as uid 0 with `cap_drop: [ALL]` and only NET_ADMIN and NET_BIND_SERVICE added; `/dev/net/tun` only in the userspace variant. Replaces the S12 examples. | Traefik's container cannot reach a host loopback port; DNS on :53 needs NET_BIND_SERVICE; capabilities do not survive no-new-privileges for a non-root user. |
| A14 | The image bakes in the binaries for all three targets; the server serves `/install.sh`, `/dl/*` and `SHA256SUMS`, and install.sh verifies the checksum. The token is taken from `MESHCTL_TOKEN` or `--token`. Refines S12. | Self-hosted install with integrity checking. |
| A15 | Safe apply arms only on server-pushed changes, never on boot-from-cache; a rejected version is not retried until a newer one arrives. Refines S6.4. | Avoids rollback loops after a reboot. |
| A16 | A config version row is written only when some node's computed config changes; agents also poll `/config` every 60 s as an SSE fallback. Refines S6.3/S8.2. | No history churn; covers proxies that drop SSE. |
| A17 | A server's endpoint comes from `--endpoint` at join, else the first public IPv4 on its interfaces (sent as `public_ip`) with its listen port, else the endpoint set at `node add`, else the public address the server saw; a server with no endpoint is not planned into other nodes' peers until it has one. Refines S6.1. | v0.1 step 5 is underspecified. |
| A18 | The agent refuses to adopt an existing interface it did not create (link alias `meshctl`), and tags its routes with rtproto 201 so cleanup touches only its own. | Hosts that already run wg-quick. |
| A19 | The NAS compose generator has kernel and userspace variants, defaulting to userspace until S16's first question is answered; the other S16 questions keep v0.1's defaults. | S16. |
| A20 | `go.yaml.in/yaml/v3` (the maintained continuation, same API) replaces `gopkg.in/yaml.v3`, which is archived. Amends S13. | Pinned and patched dependencies. |
| A21 | The agent's single exec site is `internal/agent/execrun`, allowlisting `firewall-cmd` and `systemctl` (service install at join); `agent run --no-firewall` leaves the host firewall to the operator. Sysctls are written to `/proc/sys` and re-asserted at every boot from the cached config; nothing is written to `/etc/sysctl.d`. | One audited place that runs programs; e2e and hand-managed firewalls; the unit stays read-only outside the data dir. |
| A22 | In M1 the relay and exit forwarding policies (A4) are reported as a degraded firewall step rather than applied; they ship with spokes in M3. `meshctl node set` clears a reference with the keywords `--endpoint none`, `--hub default`, `--exit off`. | M1 is servers only; an unimplemented policy must be visible, not silently missing. |
| A23 | TOTP enrolment QR codes come from `pquerna/otp`'s own image support (`boombuler/barcode`); whether the phone QR (M4) reuses it instead of adding `skip2/go-qrcode` is decided in M4. | One QR library where one suffices. |
| A24 | Users are never exported (the `--with-users` idea is dropped); they, and `secret.key`, are restored from the volume backup. `meshctl user passwd` joins `user add` and `user reset-totp` as the recovery commands. | An export is for git; password hashes and sealed TOTP secrets do not belong there. |
| A25 | The server holds an exclusive flock on `<data>/server.lock` while it runs; `meshctl import` takes the same lock, so it runs only with the server stopped, and validates the whole file before restoring it through the reconciler (node ids kept, one audited version). | A restore must never write under a running server, and a hand-edited file must not half-apply. |
| A26 | The web UI lives in `internal/server/ui` with its templates and static files embedded there (no top-level `ui/`), served on the same listener at `/` beside the agent paths; htmx 2.0.11 is vendored and pinned by sha256; the dashboard's background refresh does not extend a session. S13 edited. | One package per surface; an open dashboard must not defeat the idle timeout. |
| A27 | Exit mode installs wg-quick's rules at priorities 32764 (`lookup main suppress_prefixlength 0`) and 32765 (`not fwmark 51820 lookup 51820`) for IPv4 and IPv6, with `default dev wg0` and an IPv6 blackhole default in table 51820; every rule and route carries protocol 201. The A4 forwarding is two firewalld policies, `meshctl-relay` (mesh to mesh) and `meshctl-exit` (mesh to public, masquerade), replacing A22's degraded notice. `meshctl exit` reaches the server through the local agent (`POST /v1/prefs`, relayed as the A10 request), and the server audits a reported rollback once per distinct rollback. | wg-quick's order lets the main table's specific routes win; tagged rules are never confused with a host's own; firewalld policies are the supported way to forward between and out of zones; a rollback repeated in every heartbeat must not flood the audit log. |
| A28 | `agent run --wg-backend auto\|kernel\|userspace` (auto: the kernel when the module is there). Userspace is wireguard-go inside the agent on a tun `wg0` with the `meshctl` alias, configured through WireGuard's cross-platform text protocol; the interface lives with the agent process, and its UAPI socket (`/var/run/wireguard/wg0.sock`, so `wg show` works) is best effort, which the unit allows with `RuntimeDirectory=meshctl wireguard`. The server renders a spoke's compose file (`GET /nodes/{name}/compose` in the UI, `meshctl node compose`) with its own release as the image tag (no file from a development build), `--data /data` on every command and a one-time `docker compose run ... agent join` in its header; `deploy` may import yaml.v3 for it. Refines S7.1, S12. | The plan's userspace fallback needs a selector and an owner for its socket; a compose file must pin an image that exists and must not carry the join token. |
| A29 | Every server and spoke runs the resolver on its mesh IP. The zone is one record per enrolled, enabled node plus custom records: an alias follows its node, a fixed IP must be inside the mesh, and neither may take a node's name. A name without the asked type is NODATA, an unknown one NXDOMAIN, both with an SOA; other zones are REFUSED except in exit mode or from a server's phones (`forward_for`). Nodes report `dns` as on, off or unavailable: there is no systemd-resolved or system bus, or the agent runs in a network namespace other than pid 1's, from which it never registers. resolved is driven per link without auto-starting it, with DNSSEC, LLMNR and mDNS off on wg0; `meshctl dns off` is kept in `<data>/local.json`. Servers open 53 to the mesh CIDR; custom firewall rules open one mesh-zone port on one server or all. Every 5 minutes the agent re-applies its config after asking firewalld again (the interface's zone at runtime, each recorded rule): the drift check. Refines S7.4, S7.5. | An empty non-terminal must not answer NXDOMAIN; a VPS without resolved still serves the others; a namespaced agent would configure a host interface by index; NetworkManager moves interfaces between zones at runtime. |
| A30 | A phone is added with `node add --role phone` and gets no join token. Each profile (`meshctl node phone`, or the phone's page) has a new key pair, the private key in that answer only (A6), shown as text and as a QR code drawn inline as SVG with `boombuler/barcode` (A23: no `skip2/go-qrcode`). The profile's DNS server is the hub's mesh IP; the via-exit profile adds `0.0.0.0/0` and `::/0` to an exit-capable server's peer, so IPv6 cannot leak around it. The server records the config a profile was made from, so a phone shows `profile out of date` when its config changes, and a phone's state is the latest handshake its servers report. Amends S8.3, S13. | A phone has no agent to fetch a new config or report status; the CSP allows no `data:` images, and the key must not outlive the answer. |
| A31 | `GET /api/v1/agent/mesh` (signed) is every node as the server's dashboard shows it -- name, role, mesh IP, hub, exit, colour and label, last seen, versions -- without keys, endpoints, hostnames, observed addresses or error text. The agent validates it and caches it in `mesh.json` (at start, every 5 minutes, and when asked once the last try is 10 s old or a new config arrived), and serves it on the local socket as `GET /v1/mesh`, with the reason when the server does not answer. The agent keeps its last 2000 log lines in memory for `meshctl logs [-f]` (`GET /v1/logs`). Amends S6, S7.6. | The local UI must show the mesh while the server is down without handing every node the others' keys, addresses or failures; a node's log must be readable where the journal is not (the NAS container). |
| A32 | The local UI (A7) listens on `agent run --ui-listen` (default `127.0.0.1:7443`, loopback only; `off` turns it off; a port in use leaves the agent running without it). `meshctl ui` asks the agent socket (`POST /v1/ui-token`) for a link that works once, within 2 minutes; the session cookie is `HttpOnly` and `SameSite=Strict`, ends after an idle hour or 12 hours, and lives in memory. Any other Host is a 421; a change needs the UI's Origin and the session's CSRF token; the pages carry no script. "Restart tunnel" drops and re-adds every peer, then re-applies as the drift check does; it lives in the UI only, with no CLI. The residual risk of cookies shared across loopback ports is documented in security.md. Amends S7.6, S11. | The agent socket's group is the boundary the CLI already has, so the browser borrows it instead of having accounts; a loopback page must hold against other local users and rebinding pages; a restart that keeps the interface up fixes a stuck session without dropping routes. |
| A33 | On Windows the agent is the service `meshctl-agent` (LocalSystem, automatic start, restarted on failure, including an exit with an error), installed by `meshctl agent install` -- which copies meshctl to `%ProgramFiles%\meshctl` and joins -- or by `agent join`, and removed by `agent uninstall [--purge]`; the binary stays. The local API is the pipe `\\.\pipe\meshctl-agent` (go-winio v0.6.4) with A7's DACL; the data dir gets a protected, inherited ACL for SYSTEM and Administrators (plus an unelevated creator); the service logs to `<data>\agent.log` (10 MiB, one rotated copy). On Linux, `agent install`/`uninstall` point to install.sh and systemctl. The unit gains `ProtectKernelLogs=yes` (`systemd-analyze security`: 4.3 "OK" before it). Amends S7.7, S11. | Windows services have no stderr anyone reads and no file modes; the pipe and ACL give the boundary Linux gets from the socket's group and mode 0700. |
| A34 | On Windows the agent runs wireguard-go on a wintun adapter `wg0` with a fixed GUID (Windows keeps one network profile for it), setting its MTU, addresses and routes through the IP helper API (`golang.zx2c4.com/wireguard/windows` v1.1.1, `tunnel/winipcfg`); every static route on the adapter is the agent's, so no protocol tag is needed. `wintun.dll` 0.14.1 is fetched at build time by `internal/tools/fetchwintun` (zip and DLL pinned by SHA-256), embedded with build tag `wintun`, never committed, and written beside `meshctl.exe` at start when the file there does not match the pin. Builds without the tag (vet, test) need no network and refuse to start WireGuard on Windows. Amends S7.7, S13. | A blob in git has no provenance; a pinned fetch does. A fixed GUID avoids "Network 2, 3, ..." on every restart. |
