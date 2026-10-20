# Security

What protects what, and every place an artifact is wider than the default-deny posture, with
the reason. The threat model is spec S9; the amendments that changed it are in spec S17.

## Keys and tokens

| Asset | Protection |
| --- | --- |
| Node private keys | generated on the node by `agent join` (`identity.key` Ed25519, `wg.key` X25519), mode 0600 in a 0700 data dir; the agent refuses to start if either is readable by others. Only public keys are sent. |
| Phone private keys | the one exception (amendment A6): a phone has no agent, so the server generates its X25519 key pair when an operator asks for its profile, keeps only the public key, and puts the private key in that one answer (the profile page, `meshctl node phone`). It is never stored, logged, audited or exported, and no error carries it (`TestRenderRefuses`, `TestPhonePage`); the page is `Cache-Control: no-store` and draws the QR code inline, so no second request has to fetch it. A new profile is a new key, so a profile seen by someone else is revoked by making another; disabling the phone drops it from every server. |
| Agent identity | every request after join is signed: method, path with query, timestamp, a per-request nonce and the body hash (docs/agent-protocol.md). +-60 s skew; signatures are remembered for 120 s against replay. |
| Join tokens | 32 random bytes, stored as SHA-256, single use, valid one hour; issuing a new one revokes the node's unused ones. Never logged, audited or exported; masked in every error. |
| Admin API | a Unix socket in the data dir, mode 0600: only the server's own user (in the container, `docker exec`) can reach it. Every write is audited with actor `cli`. |
| Local agent API | a Unix socket, 0660 with group `meshctl` (else 0600); status only. Never a network listener. |

## Web UI sign-in (internal/server/auth)

The web UI (internal/server/ui) and the `meshctl user` commands share these accounts.

| Concern | Protection |
| --- | --- |
| Passwords | argon2id, t=3, m=64 MiB, p=4, 16-byte random salt, stored as a PHC string. A sign-in with a hash made under other parameters re-hashes it. At least 12 characters; read by the CLI from a terminal only, never from a pipe, a file or the command line. Never logged, audited or echoed in an error. |
| Unknown users | verified against a dummy hash, so the answer takes as long as a wrong password; unknown user, wrong password and disabled account all get the same error. |
| Second factor | TOTP (RFC 6238, 30 s, 6 digits, +-1 step), enrolled with an authenticator app at the first sign-in. Each accepted time step is recorded with a conditional update, so a code works once, even for two requests racing with it. |
| TOTP secrets at rest | AES-256-GCM with `<data>/secret.key` (32 random bytes, created at first start, mode 0600, refused when readable by others), bound to the user id. A database copy without the key does not reveal them. Back the key up with the database; losing it means `meshctl user reset-totp` for everyone. |
| Sessions | a 256-bit cookie `__Host-meshctl`: HttpOnly, Secure, SameSite=Strict, Path=/, host-only. The server stores only its sha256. 12 h idle, 7 days absolute; a session that passed the password but not the second factor lives 5 min and authorizes nothing else. A password change, a TOTP reset or disabling the user ends every session of that user. |
| CSRF | a per-session token, required (header `X-CSRF-Token` or form field `csrf`) on every request that changes something, compared in constant time. |
| Guessing | 5 failures from one address within 15 min lock it out for 15 min; 20 failures for one username within an hour lock the username, from any address. Addresses come from `X-Forwarded-For` only when the peer is a `--trusted-proxies` network. |
| Accounts | created and recovered with `meshctl user add / passwd / reset-totp` on the server host (they open the database directly, so they work with the server stopped). Every account change is audited, without secrets. Users are never exported. |

## The web UI's pages (internal/server/ui)

| Concern | Protection |
| --- | --- |
| Script injection | `html/template` escapes every value by context. A strict CSP (`default-src 'none'; script-src 'self'; style-src 'self'; img-src 'self'; connect-src 'self'; form-action 'self'; frame-ancestors 'none'; base-uri 'none'`) allows no inline script or style; `TestTemplatesHaveNoInlineScriptOrStyle` keeps the templates that way. htmx runs with `allowEval` and `allowScriptTags` off. |
| Third-party code | one file, htmx 2.0.11, vendored (checked against the npm registry's sha512 when vendored) and pinned by `HTMXSHA256` / `TestHTMXIsPinned`. No CDN, no fonts, no other assets. |
| Cross-site requests | every request that changes something must carry `Origin: <--url>` and the session's CSRF token; the session cookie is SameSite=Strict as well. Sign-in POSTs need the same origin. |
| Framing, sniffing, referrers | `X-Frame-Options: DENY`, `frame-ancestors 'none'`, `nosniff`, `Referrer-Policy: same-origin` (no referrer to other sites; `no-referrer` would make browsers send `Origin: null` on the UI's own posts, failing the origin check), `Cache-Control: no-store` on pages. |
| Idle sessions | only what the user does extends a session; the dashboard's background refresh does not. |
| Writes | every change goes through the same audited, versioned path as `meshctl node`, with actor `user:<name>`; join tokens and phone profiles are shown once and never stored in the clear. |

## Input and commands

Every value from outside the process passes `internal/validate` before it is stored, rendered
into an artifact or handed to a command; validators accept a positive grammar. The agent runs
programs only through `internal/agent/execrun`: an allowlist (`firewall-cmd`, `systemctl`), no
shell, and every argv element re-checked against `[A-Za-z0-9 ._/:=,"-]`. `TestModuleBoundaries`
fails if any other package imports `os/exec`. firewall-cmd arguments are built only from
validated zone, protocol, port and prefix values; `TestHostileRulesNeverReachFirewallCmd` proves a
hostile rule never runs a command. install.sh substitutes four markers, each validated
(`TestInstallRefusesHostileValues`) and placed inside single quotes.

## Server container

`deploy/compose.server.yml` (and the dev stack) state their posture explicitly, pinned by
`TestServerComposePosture` / `TestDevComposePosture`, which also forbid every widening token
(`cap_add`, `devices`, `network_mode`, `privileged: true`, host mounts, ...):

| Setting | Value |
| --- | --- |
| user | `65532:65532` (distroless `nonroot`) |
| privileged | `false` |
| capabilities | `cap_drop: [ALL]`, none added |
| privilege escalation | `no-new-privileges=true` |
| root file system | read-only; `/tmp` is a 16 MiB tmpfs |
| network | the external `dokploy-network` bridge, `expose: 8080` only; nothing published on the host (the dev stack publishes on `127.0.0.1` only) |
| storage | the named volume `meshctl-data` at `/data` |
| limits | 256 MiB memory, 256 pids |

The image is distroless static, built `CGO_ENABLED=0` from pinned base-image digests.

## Agent service (systemd)

`deploy/meshctl-agent.service` is installed unchanged by `agent join`; `TestUnitPosture` pins
its `[Service]` section exactly. The agent owns `wg0`, its routes, two sysctls and (on servers)
firewalld, so it runs as root -- but bounded:

| Setting | Why |
| --- | --- |
| `CapabilityBoundingSet=` / `AmbientCapabilities=CAP_NET_ADMIN CAP_NET_BIND_SERVICE` | the only capabilities: links, addresses, routes, WireGuard, the resolver on :53, and DNS registration with systemd-resolved (which asks for CAP_NET_ADMIN). Root without `CAP_DAC_OVERRIDE`, `CAP_SYS_ADMIN` and the rest. |
| `NoNewPrivileges=yes`, `RestrictSUIDSGID=yes` | nothing it starts can gain privileges |
| `ProtectSystem=strict`, `ReadWritePaths=/var/lib/meshctl` | the file system is read-only except the data dir and the two runtime dirs below |
| `RuntimeDirectory=meshctl wireguard`, `RuntimeDirectoryPreserve=yes` | `/run/meshctl` holds the local API socket, `/run/wireguard` the userspace backend's UAPI socket that `wg show` reads; both are kept across restarts, so another WireGuard daemon's socket in `/run/wireguard` is never removed with the agent's |
| no `ProtectKernelTunables` | deliberate: relay and exit servers need `net.ipv4.ip_forward` written to `/proc/sys` |
| `ProtectHome`, `PrivateTmp`, `ProtectKernelModules`, `ProtectKernelLogs`, `ProtectControlGroups`, `ProtectClock`, `ProtectHostname`, `RestrictNamespaces`, `RestrictRealtime`, `LockPersonality`, `MemoryDenyWriteExecute`, `SystemCallArchitectures=native` | standard hardening; none of it is needed by the agent |
| `DevicePolicy=closed`, `DeviceAllow=/dev/net/tun rw` | only the tun device, for the userspace backend |
| `RestrictAddressFamilies=AF_UNIX AF_INET AF_INET6 AF_NETLINK` | sockets, D-Bus to firewalld, and netlink for links and routes |
| `UMask=0077`, `StateDirectoryMode=0700` | everything it writes is private |

`systemd-analyze security --offline=true deploy/meshctl-agent.service` rated the unit 4.3, "OK",
on EL10 before `ProtectKernelLogs=yes` was added (2026-10-16). What it still counts is what the
agent is for -- root with network capabilities, netlink and IP sockets, the host's network, the
tun device, writable kernel tunables -- plus `SystemCallFilter`, the one large step left. A
syscall the filter missed would kill the agent and drop the tunnel (and `firewall-cmd`, which the
agent runs, inherits the filter), so it waits for a test on a real AlmaLinux node.

With the userspace backend (`agent run --wg-backend userspace`, or `auto` on a host without the
module) the agent itself is the WireGuard implementation, wireguard-go on a tun interface it
creates and marks like a kernel one (alias `meshctl`). Its UAPI socket,
`/var/run/wireguard/wg0.sock`, is created under umask 077, so only root reaches it, and accepts
what a root `wg set` could change through the kernel; it is best effort, and where it cannot be
created only `wg show` loses sight of the interface. The configuration the agent hands the device
never reaches a log: wireguard-go logs no private key, and errors that read the device back
name the field, never its value (`TestParseUAPIRejectsMalformedValues`).

## NAS container

A spoke on a NAS runs the agent from a compose file the server renders for it (`deploy/nas.go`,
amendments A13, A19, A28) from a typed value that yaml.v3 writes, so no value can add a key. It
widens exactly what an agent that owns the host's `wg0` needs; `TestNASComposePosture` pins every
key and value and forbids every other widening token:

| Setting | Value | Why |
| --- | --- | --- |
| user | `0:0` | capabilities do not survive `no-new-privileges` for another user; this root holds only the two below |
| privileged | `false` | |
| capabilities | `cap_drop: [ALL]`, `cap_add: [NET_ADMIN, NET_BIND_SERVICE]` | wg0, its addresses and routes; the resolver on :53 |
| privilege escalation | `no-new-privileges=true` | |
| network | `network_mode: host` | wg0 and its routes belong to the NAS, not to a container network |
| devices | `/dev/net/tun`, in the userspace variant only | wireguard-go's interface; the kernel variant maps none |
| root file system | read-only; `/tmp` is a 16 MiB tmpfs | |
| storage | one bind mount: the chosen host directory at `/data` | keys and the cached config |
| limits | 256 MiB memory, 256 pids | |

Every value written into the file passes `internal/validate` first: the node name, the server's
canonical origin, a release version (a development build gets no file), the backend, and the
host directory as a clean absolute path of `[A-Za-z0-9._/-]`
(`TestNASComposeRefusesHostileValues`). Below the validators the renderer still holds: a
newline, `---`, `: ` or `$` handed to it directly stays inside its value, every `$` is doubled
so compose never interpolates it, and the header comment cannot be broken
(`TestNASRendererHoldsItsStructure`). The file carries no secret: the join token is typed into
the one-time join command.

## Container-mode agent

A container-mode node (amendment A37) runs the agent from a compose file the server renders
(`deploy/container.go`) from a typed value, like the NAS's. The agent lives in its own network
namespace: its capabilities act there only, and the host, its addresses and its LAN are not on
the mesh. `TestContainerComposePosture` pins every key and value and forbids every other
widening token, including the host network:

| Setting | Value | Why |
| --- | --- | --- |
| user | `0:0` | as the NAS: capabilities do not survive `no-new-privileges` for another user |
| privileged | `false` | |
| capabilities | `cap_drop: [ALL]`, `cap_add: [NET_ADMIN, NET_BIND_SERVICE]` | wg0, its routes and the gateway's policy rules, in the agent's namespace; the resolver on :53 |
| privilege escalation | `no-new-privileges=true` | |
| network | its own namespace on two networks: an egress network and the external `meshctl-mesh` | the runtime NATs its WireGuard UDP; containers attach to the mesh network |
| sysctls | `net.ipv4.ip_forward=1`, `net.ipv4.conf.all.proxy_arp=1` | the gateway forwards and answers ARP for mesh addresses; both are per-namespace |
| devices | `/dev/net/tun` unless the kernel backend | wireguard-go when the host has no module |
| ports | one UDP port, on a hub only | the node's endpoint; nothing else is published |
| root file system | read-only; `/tmp` is a 16 MiB tmpfs | |
| storage | a named volume, or the chosen host directory, at `/data` | keys and the cached config |
| runtime socket | the host's Docker (or rootful Podman) API socket at `/var/run/docker.sock`, `:ro`; absent with `--no-names` | the containers' names (amendment A38, below) |
| limits | 256 MiB memory, 256 pids | |

The mesh network is internal (Podman: without its DNS), with the whole mesh on the link and the
node's upper /25 handed out; the bridge address .1 is the host's. What may cross the gateway is
its policy routing, in its own namespace (`internal/agent/containers`): forwarded traffic from
wg0 reaches only the node's subnet and never the bridge address, from the mesh side only the
subnet reaches the mesh, and everything else forwarded is blackholed -- so the host reaches no
mesh address (the gateway does not even answer its ARP), no mesh peer reaches the host or the
gateway's egress network, and a peer that routes a foreign address to the gateway gets nowhere
(the M6 trial's R0-R7, and `TestContainerNetworks` in e2e).

**Host mode with containers.** The host is on the mesh and routes the subnet on the bridge
`meshctl0` the agent owns. A runtime that lets anyone who routes to the host reach its containers
(Docker's `nat-unprotected`, which the mesh needs so its peers reach unpublished ports) must not
open them to the host's LAN: four policy rules let only wg0, the bridge and the host itself reach
the subnet, and blackhole it from every other interface (the trial's G1-G5; e2e's LAN host). One
consequence: a port published on the host for a container on the mesh network is not reachable
through that network from outside; publish it on the container's other network. Every value written into the file passes
`internal/validate` (`TestContainerComposeRefusesHostileValues`), and below the validators the
renderer holds its structure (`TestContainerRendererHoldsItsStructure`).

**The runtime socket** (amendment A38) is the widest token in the file: whoever can talk to it
controls the runtime, and so the host -- `:ro` on the mount does not limit the API, it only says
what the agent does with it. The agent sends GET requests only (`internal/agent/names`): the
containers on `meshctl-mesh`, each one's details, the event stream. Every answer is size-capped
(8 MiB for the list, 1 MiB per container, 64 KiB per event), a container ID must be 64 hex
characters before it becomes part of a path, and every name passes `validate.Hostname` or is left
out. A compromised agent container could use the socket: `meshctl node compose --no-names`
leaves it out, and a socket proxy that allows only GET on `/containers` and `/events` is a
drop-in where that matters.
A host-mode agent already runs as root on the host and opens the socket it finds
(`--runtime-socket off` stops it).

What a node reports is checked by the server before anyone resolves it: a name whose address is
not one of the reporting node's containers' (inside its /24, not .0, .1, .2 or .255) is not kept,
so a node cannot point a name at another node, its gateway or the host; a name in the zone must be
one label (any other name must lie inside an extra domain, below), and a node's name, a custom record's and the host of the server's URL are never given
to a container, so a container cannot take over `vps-1.<zone>` or the server's address. Names
are not secrets: the mesh view carries them to every node, as DNS would.

**Extra domains** (amendment A39) are domains the mesh takes over: the hosts send them to the
mesh resolver, which answers the containers' names in them and forwards every other name in them
to the mesh's upstream resolvers, whoever asks -- only those domains, so the resolver is still
not an open forwarder. A container may then claim any name in such a domain, which is why the
operator chooses them: at most 8, never the zone or around it, no .arpa but home.arpa, and never
the domain of the server's own address (`CheckServerDomains`), which every agent must resolve
without the mesh.

## The host firewall

On servers the agent creates firewalld's `mesh` zone (target DROP) bound to `wg0`, opens the
Swarm ports in it only from the other servers' /32s, DNS (53/udp, 53/tcp) from the mesh CIDR, and
udp/51820 in the public zone. Custom rules from the web UI's Firewall page open one port in the
mesh zone from the source network an operator gives (validated as an IPv4 network; its note is
display-only and never reaches firewall-cmd). It records what it added in `firewall.applied.json`
and only ever removes its own rules. Every 5 minutes, and at every start, it asks firewalld again
(the zone's interface at runtime too, and each recorded rule) and puts back what was changed
underneath it. `--no-firewall` turns this off for hosts whose firewall is managed elsewhere.

Forwarding is opened only where the plan needs it (amendment A4), as two policies the agent owns by
name: `meshctl-relay` (mesh to mesh) on a server that relays between spokes, and `meshctl-exit`
(mesh to public, masqueraded) on an exit-capable server. Nothing else may forward: the mesh zone's
DROP target still applies to every other path, so a spoke reaches another spoke only through its
relay and the internet only through a server an operator allowed with `--can-exit`. A policy
changed by hand is put back as a whole; one the plan no longer asks for is deleted. Their argv is
built from constants only (`TestForwardingPolicies` pins it).

## DNS

Each node's resolver (`internal/agent/dns`) listens on its mesh IP (and its containers' gateway)
only, never on a public or LAN address, and answers the mesh zone from the config the server
signed off. It is not an open resolver: other names are REFUSED, except in exit mode (the node's
own lookups, which then go through the tunnel), from the containers of any container network on
the mesh, and, on a server, from the /32 of a phone whose hub it is (`forward_for`) -- all of
them mesh members.
Servers open 53 only in the mesh zone. Forwarding goes to the operator's upstream resolvers,
validated as IPv4 addresses outside the mesh.

The host side changes only systemd-resolved's settings for wg0, over the system bus, and is
taken back on `dns off` and when the agent stops; `/etc/hosts` and `/etc/resolv.conf` are never
written. resolved accepts it because the agent holds CAP_NET_ADMIN; calls never start resolved
where it is not running. DNSSEC is off for wg0's lookups only: the mesh zone is not signed, and
in exit mode the upstream resolvers (1.1.1.1 and 9.9.9.9 by default) validate. An agent in a
network namespace other than the host's does not register at all, since resolved would apply its
interface index to one of the host's interfaces.

## Changes a node asks for

A node can change only its own exit, keepalive and MTU (`POST /api/v1/agent/prefs`, amendment
A10), signed with its identity key like every agent request, through the same reconciler, plan
check and audit as an operator's change (actor `node:<id>`). It cannot name another node, widen
its AllowedIPs or pick a server that is not exit-capable. On the node, `meshctl exit` reaches the
agent over the local socket (root or group `meshctl`), which validates the request before it
leaves the node. `meshctl dns on|off` stays on the node and changes only its own registration.

## What a node sees of the others

Any joined, enabled node can read the mesh view (`GET /api/v1/agent/mesh`, amendment A31): each
node's name, role, mesh IP, hub, exit, dashboard state, last seen and versions -- most of which
its own config and the DNS zone show already. It carries no key, endpoint, hostname, observed
address or error text, so a compromised laptop learns from it neither another node's public
address nor what failed where. It does carry the containers' published names and mesh addresses
(amendment A38), which every node's resolver answers anyway. The agent validates the view before it caches it in `mesh.json`
(0600) or serves it on the local socket. `meshctl logs` reads the agent's in-memory log over the
same socket (root or group `meshctl`); the log never carries a key or a token, so it shows what
the journal does.

## The local UI (internal/agent/localui)

Each agent serves its UI on a loopback address only (`--ui-listen` refuses any other; amendment
A7, A32). Loopback TCP is open to every local user and, through DNS rebinding, to any web page, so:

- **Sign-in** is a link from `meshctl ui`, which needs the agent socket (root or group
  `meshctl`): whoever can sign in could run `meshctl exit` anyway. The link's token is 32 random
  bytes, kept as its sha256, good for 2 minutes and once; at most 8 are outstanding. Tokens and
  cookies never reach the log.
- **The session** is a cookie of 32 random bytes (sha256 kept), `HttpOnly`, `SameSite=Strict`,
  path `/`, not `Secure` (the UI is plain HTTP on loopback); an idle hour or 12 hours end it, at
  most 8 live in memory, and an agent restart ends them all.
- **Every request** must name the UI's own host (`127.0.0.1:7443` or `localhost:7443`), or it gets
  421 and reaches nothing -- a rebinding page cannot even spend a link.
- **Every change** is a POST whose Origin is the UI's and whose form carries the session's CSRF
  token (compared in constant time), at most 16 KiB; values are validated before they leave the
  page, and exit, keepalive and MTU go to the server through the same signed relay as `meshctl
  exit`, where they are checked again and audited as `node:<id>`.
- **Every answer** carries `default-src 'self'; frame-ancestors 'none'; form-action 'self';
  base-uri 'none'`, `X-Frame-Options: DENY`, `nosniff`, `Referrer-Policy: same-origin` and `no-store`; the pages
  carry no script.

What remains: browsers send a host's cookies to every port of it, so a program another local
user runs on a different loopback port receives this browser's session cookie if the browser
visits it, and could use it until the session ends. The session reaches no more than `meshctl`
on the node does -- exit, keepalive and MTU (checked and audited by the server), split DNS, a
tunnel restart, the mesh view and the log -- for at most 12 hours; sign out when done, and on a
machine other people log in to, run the agent with `--ui-listen off`.

## Windows (M5, in progress)

- **The service** `meshctl-agent` runs `meshctl agent run --data C:\ProgramData\meshctl` as
  LocalSystem, started with Windows and restarted on failure. Installing, changing or removing
  it takes an elevated prompt.
- **The data dir** gets a protected ACL that every file in it inherits: SYSTEM and Administrators
  only (`D:P(A;OICI;FA;;;SY)(A;OICI;FA;;;BA)`), the Windows form of mode 0700, since file modes
  mean nothing there. An unelevated process that opens a data dir (a test, `--no-service`) is
  added too, so it keeps access to what it made. The keys, the cached config and the service's
  `agent.log` live there.
- **The local API** is the named pipe `\\.\pipe\meshctl-agent`, which only Administrators and
  SYSTEM may open (`D:P(A;;GA;;;BA)(A;;GA;;;SY)`, amendment A7): `meshctl status`, `exit` and
  `ui` need an elevated prompt. A second agent cannot take the pipe while one listens.
- **The log file** is capped at 10 MiB plus one older copy and carries what the journal carries
  on Linux: no key, no token.
- **Split DNS** is NRPT rules under the agent's own keys in
  `HKLM\SYSTEM\CurrentControlSet\Services\Dnscache\Parameters\DnsPolicyConfig`: `.mesh`, one per
  extra domain (amendment A39), and in exit mode `.` (every name), all to the mesh IP. Every key
  starts with the agent's prefix (`{6d657368-6374-6c0`, "meshctl" in hex), a domain's derived from
  its name; the agent writes, reads and removes only keys with that prefix, and takes them back on
  `meshctl dns off` and when it stops.
- **Exit mode** routes 0.0.0.0/1 and 128.0.0.0/1 into the tunnel, and ::/1 and 8000::/1 too,
  with no IPv6 peer behind them, so IPv6 cannot leak around the exit (amendment A11). The
  tunnel's own packets are bound to the interface of the host's default route, so they never
  loop into the tunnel; the agent rebinds them when that route moves.
- **wintun.dll** (amendment A34) is the one binary meshctl ships that it does not build. The
  build fetches the wintun 0.14.1 release, refuses it unless the zip and the DLL match their
  SHA-256 pins (`internal/agent/wintun`), and embeds the DLL; it is never committed. At every
  start the agent writes it beside `meshctl.exe` in Program Files (writable by Administrators and
  SYSTEM only, so nobody else can plant a DLL there) unless the file there already matches the
  pin, and Windows loads it from that directory or System32 only. Its licence is
  `docs/third-party/wintun-LICENSE.txt`; meshctl uses it only through its API (wireguard-go's
  `tun` package), as the licence requires.

## Interfaces the agent did not create

The agent marks `wg0` with the link alias `meshctl` and refuses to configure an interface without
it, so a host's existing wg-quick setup is never taken over. Its routes carry protocol 201; it
lists and removes only those (amendment A18). The exit-mode policy rules and the routes in table
51820 carry protocol 201 too (amendment A27), so a host's own policy routing is never touched.
