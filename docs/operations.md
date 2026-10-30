# Operations

Running a mesh: bringing up the server, enrolling the VPS nodes, laptops and exits, day-to-day
node changes, phones, DNS and the firewall, the web UI and each node's local UI, backups and the
usual failures. Windows is under way in M5 (spec S15): the service and its local API are in
place, WireGuard there is not yet. Every command and flag is in
[commands.md](commands.md).

## Contents

- [1. Run the control plane (Dokploy)](#1-run-the-control-plane-dokploy)
- [2. Enroll the server nodes](#2-enroll-the-server-nodes)
- [3. Swarm over the mesh](#3-swarm-over-the-mesh)
- [4. Spokes and exits](#4-spokes-and-exits)
  - [Container mode](#container-mode)
- [5. Day to day](#5-day-to-day)
- [6. DNS](#6-dns)
- [7. Firewall](#7-firewall)
- [8. The web UI](#8-the-web-ui)
  - [The mesh firewall](#the-mesh-firewall)
  - [Each node's local UI](#each-nodes-local-ui)
- [9. Users and recovery](#9-users-and-recovery)
- [10. Backups](#10-backups)
- [11. When the control plane is down](#11-when-the-control-plane-is-down)
- [12. Troubleshooting](#12-troubleshooting)
- [13. Upgrading](#13-upgrading)

## 1. Run the control plane (Dokploy)

The control plane (the server: database, web UI, agent API) is one container behind Dokploy's
Traefik; it runs no WireGuard of its own. Create a compose app from
[`deploy/compose.control.yml`](../deploy/compose.control.yml) -- or, on Docker Swarm, a stack app
from [`deploy/stack.control.yml`](../deploy/stack.control.yml), after labelling its node
(`docker node update --label-add meshctl.control=true <node>`) -- and set its environment.
[containers.md](containers.md) has samples for the control plane and every kind of node in a
container: compose, `podman run` and Kubernetes. meshctl generates no container deployment
(amendment A48).

The image runs `meshctl start`, which reads what to run from the container's environment
(amendments A47, A48): the compose file sets `MESHCTL_MODE=control` and takes the rest from the
app's environment (Dokploy's, a `.env` file beside the compose file, or the shell's). A node's
container leaves the mode out (it means `node`): whether a node is a server node or a client node
is its role here, not its container's setting. No file has a command line;
[commands.md](commands.md#meshctl-start) lists every variable `start` reads.

| Variable | Value |
| --- | --- |
| `MESHCTL_VERSION` | the release to run, e.g. `1.0.0` (the image tag, without the `v`) |
| `MESHCTL_URL` | the public origin agents use, e.g. `https://mesh.example.com` (required) |
| `MESHCTL_PORT` | optional: the port the control plane listens on (default `8080`) |
| `MESHCTL_TRUSTED_PROXIES` | optional: the CIDR of `dokploy-network`, so the server sees real client addresses |
| `MESHCTL_HUB_DOWN_AFTER` | optional: relay failover after a hub is silent this long (default `2m`, `0` off) |
| `MESHCTL_LOG_LEVEL` | optional: `debug`, `info` (default), `warn` or `error` |
| `TZ` | optional: the time zone the web UI, the dashboard and the server's log show times in, e.g. `Asia/Singapore` (default `UTC`) |

Point the domain at the server's port (8080, or `MESHCTL_PORT`) in Dokploy. Nothing is published
on the host. The container runs as uid 65532 with no capabilities and a read-only root file
system; the database and the admin socket live in the `meshctl-data` volume at `/data`.

Check it:

```
curl -fsS https://mesh.example.com/healthz          # ok
docker exec <container> /app/meshctl node list      # empty table
```

If a GeoIP allowlist fronts the domain, `/api/v1/agent/*`, `/install.sh`, `/install.ps1` and `/dl/*` must stay
reachable from every node's country.

## 2. Enroll the server nodes

For each VPS, on the Docker host:

```
docker exec <container> /app/meshctl node add vps-1 --role server
```

It prints a join token (valid for an hour, shown once) and the command to run on the VPS, as
root:

```
curl -fsSL https://mesh.example.com/install.sh | sudo sh -s -- --mesh default --token <token>
```

install.sh downloads meshctl from your server, checks its SHA-256 against the value baked into
the script, installs it to `/usr/local/bin`, installs the systemd unit `meshctl-agent@.service`,
joins, and enables and starts the mesh's instance, `meshctl-agent@default` (amendment A51). The
running instance stops for the join; when the join fails (an expired token, the wrong `--mesh`),
install.sh starts it again on its earlier enrollment. The agent's data dir is the mesh's,
`/var/lib/meshctl/default`, and its local API socket
`/run/meshctl/default.sock`; `meshctl status`, `exit`, `dns`, `logs`, `ui` and `agent reset`
find it by themselves while the host has one mesh's agent (`--mesh <mesh>` with several). A join
whose token belongs to another mesh's node is refused before the token is used, saying which.
Without systemd (in a container) install.sh joins and says to run `meshctl agent run --mesh
<mesh>`. The agent creates `wg0` with the node's mesh IP, opens udp/51820 in
firewalld's public zone, the Swarm ports from the other servers and DNS from the mesh in the
`mesh` zone, and answers `.mesh` names on its mesh IP (section 6).

The server's public endpoint is detected from the VPS's own interfaces. When that is wrong (a
NAT'd VPS), pass it: `sudo sh -s -- --mesh default --token <token> --endpoint 203.0.113.10:51820`, or set it on
the server with `meshctl node set vps-1 --endpoint 203.0.113.10:51820`.

Within a minute `meshctl node list` shows the node `in sync` and `wg show` on the other servers
lists it. WireGuard handshakes on first traffic: `ping 10.77.0.2` from vps-1 starts one.

The VPS this server runs on joins the same way: its agent is a normal node.

## 3. Swarm over the mesh

meshctl does not run Docker. Bind Swarm to the mesh addresses (spec S14):

```
docker swarm init --advertise-addr 10.77.0.1 --listen-addr 10.77.0.1 --data-path-addr 10.77.0.1
docker network create -d overlay --opt com.docker.network.driver.mtu=1370 <name>
```

The `mesh` firewalld zone allows 2377/tcp, 7946/tcp+udp and 4789/udp only from the other
servers' /32s.

A Swarm that runs over the mesh never runs the agents: after a reboot it reaches its managers only
through the mesh, so it could not start the agent that brings the mesh up. Its hosts run the agent
from `install.sh` (or compose). Where Swarm runs over the internet or a LAN instead, a node's agent
may be a Swarm service, and one global service can make every labelled Swarm node a client node:
[containers.md](containers.md#choosing-how-to-run-each-node) has the guide and the samples
(amendment A53).

## 4. Spokes and exits

A spoke is a node without a public address: a laptop, the NAS. It dials every server and keeps
the tunnel open with a 25 s keepalive. Traffic between two spokes goes through one server, the
same one in both directions (amendment A3): the hub of the spoke that was added first.

**A Linux laptop.** On the server host:

```
docker exec <container> /app/meshctl node add laptop --role spoke --hub vps-1
```

then, on the laptop as root, the install command it prints. The page the web UI shows after
adding a node (or giving it a new token) also lists the ways to run it in a container -- a compose
file, a Swarm service and, for its containers, the network commands -- filled with the node's
values and token (amendment A53); a server node on its host's network runs from `install.sh` only. The laptop gets an address from
10.77.2.0/24; `--hub` names the server that relays for it (default: the mesh's default hub, else
the first server). A host without the WireGuard kernel module runs WireGuard inside the agent
instead (wireguard-go on `/dev/net/tun`): the agent picks that itself (`agent run --wg-backend
auto`), and `meshctl status` names the backend in use.

**The NAS (a container).** On the control plane's host:

```
docker exec <container> /app/meshctl node add nas --role spoke --always-on
```

`--always-on` gives it an address from 10.77.1.0/24. Its agent runs in a container on the NAS's
host network: write the compose file from the sample in [containers.md](containers.md#a-client-node-on-the-host-network)
(the release image, `MESHCTL_URL`, `MESHCTL_NAME=nas`, the posture docs/security.md describes).
Its WireGuard is the NAS's own module when it has one (`ls /sys/module/wireguard`), else
wireguard-go on the mapped `/dev/net/tun`; the host directory at `/data` keeps its keys and cached
config. Start it with the token `node add` printed; the agent joins on its first start:

```
MESHCTL_TOKEN=<token> docker compose up -d
docker compose exec meshctl /app/meshctl status --data /data
```

**Join on start (every node's container, amendment A47).** A node's container sets
`MESHCTL_URL` (the control plane's origin) and the node's name, and takes the join token from
`MESHCTL_TOKEN` -- exported, as above, or in a `.env` file beside the compose file -- or from the
file `MESHCTL_TOKEN_FILE` names (a compose secret: add `secrets: [meshctl_token]` to the service
and `secrets: {meshctl_token: {file: ./meshctl-token}}` at the top, then set
`MESHCTL_TOKEN_FILE=/run/secrets/meshctl_token`). The token is never written into the file.
At every start the agent decides:

| The node | The token | At start |
| --- | --- | --- |
| has not joined | none | stops: set `MESHCTL_TOKEN` (the container restarts and says so again) |
| has not joined | one it has not used | joins, then runs |
| has joined | none, or the one it joined with | runs; no join |
| has joined | a new one (`meshctl node token <name>`) | joins again with fresh keys, then runs |

The data directory keeps a fingerprint of the token (`join.json`), never the token. A token
the server calls spent (used or expired) is remembered too, so restarts do not ask again; any
other failure (no answer, a rate limit, a disabled node) is tried at the next start. A node
that has joined keeps running on its enrollment when a new token fails, with a warning in its
log: a bad token never takes a working node down. To join again, start it the same way with
the new token -- compose recreates the container when `MESHCTL_TOKEN` changes -- or write the
new token to the secret's file and `docker compose restart`. A token sits in the container's
environment, where `docker inspect` shows it; it is single-use and expires, and the file form
keeps it out.

A userspace interface lives as long as the agent, so the NAS is off the mesh while its container
restarts (a kernel interface stays up). The agent changes no firewall on a spoke: allow the NAS's
app ports from 10.77.0.0/16 in its own firewall (spec S7.5).

**A Windows laptop** (M5: everything below works; it has yet to pass the live test on a real
laptop). Add it as a spoke on the server, which prints, beside install.sh, the command for an
elevated PowerShell (Run as administrator) on the laptop:

```
$env:MESHCTL_TOKEN = '<token>'; $env:MESHCTL_MESH = 'default'
irm https://mesh.example.com/install.ps1 | iex
```

install.ps1 (amendment A51) downloads meshctl for Windows from your server, checks its SHA-256
against the value baked into the script, copies it to `C:\Program Files\meshctl\meshctl.exe`,
joins, and installs and starts the mesh's service `meshctl-agent-default` (a failed join starts
again every service it stopped), which Windows starts at boot and restarts when it fails (`MESHCTL_NAME`, when set, must be the token's node). Its data is
in `C:\ProgramData\meshctl\default` (SYSTEM and Administrators only), its log in `agent.log`
there (10 MiB, plus one older file). `meshctl status`, `exit`, `dns`, `logs` and `ui` reach it over
the named pipe `\\.\pipe\meshctl-agent`, from an elevated prompt only. To remove it:
`Stop-Service meshctl-agent-default; meshctl agent reset --mesh default; sc.exe delete
meshctl-agent-default`, and delete `C:\ProgramData\meshctl\default` for good (the node must then
join again); the binary stays in Program Files.

The agent runs WireGuard itself (wireguard-go) on a wintun adapter named `wg0`, which exists while
the service runs. meshctl carries the wintun driver and writes `wintun.dll` beside
`meshctl.exe`, checked against its pinned hash at every start. Windows Defender Firewall treats
`wg0` as a public network: other nodes reach the laptop's services only where you allow them.
`.mesh` names, and the extra domains, resolve through NRPT rules that send them to the laptop's own
resolver (`Get-DnsClientNrptPolicy` lists them); `meshctl dns off` removes them. `meshctl exit vps-2` works as
on Linux: the laptop routes everything but the mesh through vps-2, IPv6 is blocked meanwhile, and
the tunnel follows the laptop from one network to the next.

A Windows laptop in a second mesh (amendment A49) runs a second agent -- install.ps1 again with
that mesh's token and name -- as the service `meshctl-agent-lab`, with its own data dir
(`C:\ProgramData\meshctl\lab`) on its slot's adapter (`wg1`, a GUID of its own), its local UI
on 7443 plus the slot and its pipe `\\.\pipe\meshctl-agent-<slot>` (the CLI finds it from `--mesh`);
its NRPT rules sit under keys of its own slot. One mesh's exit at a time: a second one stays off
(degraded `exit`), and while an exit is on, the other meshes' tunnels run through it.

**Exits.** Allow a server to be an exit once: `node set vps-2 --can-exit`. Its agent turns on
forwarding and firewalld's `meshctl-exit` policy, which masquerades mesh traffic to the
internet. Then, on the laptop (root, or a member of group `meshctl`):

| Task | Command |
| --- | --- |
| see the exit and the choices | `meshctl exit` |
| send internet traffic through vps-2 | `meshctl exit vps-2` |
| switch to vps-3; connections over the mesh stay up | `meshctl exit vps-3` |
| internet straight from the laptop again | `meshctl exit off` |

The command goes through the local agent to the server, which owns every peer list, so it needs
the server; it returns once the laptop has applied the new config. An operator does the same with
`node set laptop --exit vps-2` (`--exit off`). In exit mode the LAN and the mesh still go direct,
and IPv6 to the internet is blocked rather than sent around the tunnel -- except on a kernel with
IPv6 but no IPv6 policy routing (some NAS kernels), where the node reports `degraded exit`
("carries IPv4 only"): turn IPv6 off there (`sysctl net.ipv6.conf.all.disable_ipv6=1`) before
using it as an exit client.

**Phones.** A phone runs the WireGuard app, not an agent. Add it, then make its profile:

```
docker exec <container> /app/meshctl node add phone --role phone --hub vps-1
docker exec <container> /app/meshctl node phone phone --exit vps-2 > phone.conf
```

or, on the phone's page in the web UI, **Make a profile**, which shows it as a QR code (in the
app: **+**, **Scan from QR code**) and as text. The profile holds the phone's private key: the
server makes the key pair, keeps the public half and shows the profile once, so making another
replaces the key and the previous profile stops working (amendment A6). Two kinds:

- mesh only (no `--exit`): only mesh addresses go through the tunnel;
- via an exit (`--exit vps-2`, an exit-capable server): all of the phone's traffic, IPv6
  included, goes through that server (and out of it when the server has public IPv6; dropped
  there when not).

On a dual-stack mesh the profile gives the phone both mesh addresses and both of its hub's
resolvers (`Address = 10.77.3.1/32, fd...::3:1/128`); a phone added `--family ipv4` gets the IPv4
side alone. A profile made before the mesh had IPv6 has none: make a new one.

Either way the phone sends its DNS lookups to its hub's resolver, which answers `.mesh` names and
forwards the rest upstream. The phone gets an address from 10.77.3.0/24 and reaches spokes
through its hub; two phones reach each other only when they share a hub. A profile is static:
when the servers or the phone's hub change, the phone shows `profile out of date` until you make
a new one. It shows `tunnel on` while a server has seen a handshake in the last 3 minutes,
`idle` otherwise.

**Safe apply.** On a server node every pushed change has to prove the server is still reachable:
if no heartbeat is answered within 60 s of applying it, the agent puts the previous config back.
The node then shows `rolled back` (red on the dashboard; `node show` says why) and stays on the
old config until the server's config changes again. Fix what cut it off, then make the next
change -- or roll back to a good version in the UI.

### Container mode

(Amendment A37.)

A node in container mode runs its WireGuard in the agent's container, in that container's own
network namespace. Only the containers attached to the node's `meshctl-mesh` network are on the
mesh; the host, its addresses and its LAN are not. A server node in container mode is still a hub
with a public endpoint (the host publishes its WireGuard port); a client node only dials out.

```
docker exec <container> /app/meshctl node add box --role spoke --mode container --hub vps-1
docker exec <container> /app/meshctl node add hub4 --role server --mode container --endpoint 198.51.100.7:51820
docker exec <container> /app/meshctl node show box
```

The node gets a mesh IP as any node and a /24 for its containers (10.77.128.0/24, then .129, ...;
the mesh needs to be a /21 or larger). A container-mode server node needs `--endpoint`: inside its
container the agent cannot see the host's public address. `node show` (and the node's page in the
web UI) lists the container network's values -- the subnet on the link, the range the runtime
hands out, the bridge, the agent's address, the resolvers, the port a server node publishes --
which [containers.md](containers.md#container-mode) puts into the commands that create the mesh
network once (Docker, Podman, or Swarm for Swarm services), the agent's compose file and a
service's attachment. Start the agent with the token; it joins on its first start, as above.

An attached service gets an address from the subnet's upper half and reaches every mesh address
-- other nodes, their containers, phones -- through the agent, with no route of its own; its own
internet stays on its other network. Its `dns` (the agent) and `dns_search` (the zone) make
`<name>.<zone>` and short names resolve through the agent.

The agent claims its address (`.2`) on its mesh-network interface itself (a Swarm-scoped network
ignores a fixed address) and isolates by policy routing in its own namespace: the host reaches no
mesh address and no mesh peer reaches the host. `meshctl status` shows `degraded: containers`
when the agent is not on the mesh network -- create it with the node's values and recreate the
agent's container.

A container-mode node is never an exit and uses none; the mesh firewall reaches its containers
by `containers:<node>` (its host is not on the mesh). A Windows laptop is in host mode only. On a
host with both Docker and rootful Podman, containers.md has the step that lets the mesh bridge
through Docker's FORWARD policy.

**A host-mode node's containers.** `node add ... --containers`, or `node set <name>
--containers` later, gives a host-mode node a container subnet too: the host stays on the mesh
(a Swarm VPS keeps Swarm over it) and its containers get mesh addresses. The agent owns the
bridge `meshctl0`: it makes it when it is missing, gives it the subnet's `.1`, turns on its
forwarding and proxy ARP, and adds four rules so that only the mesh, the bridge and the host
reach the subnet -- never the host's LAN. Create the runtime's network on that bridge once, with
the values `node show` lists ([containers.md](containers.md#a-host-mode-nodes-containers) has the
Docker, Podman and Swarm commands): Docker takes the existing bridge without masquerading or
filtering ports (`inhibit_ipv4`, `nat-unprotected`), Podman attaches to it unmanaged, Swarm uses a
config-only network with the same options. A service then attaches as in container mode, with
the bridge's `.1` as its `dns`; a Swarm service, whose tasks run on several nodes, names the
server nodes' mesh IPs instead (`node show` lists them), since every node's resolver answers every
node's containers. On a firewalld server node the agent puts an unzoned `meshctl0` into the mesh
zone (Docker's own bridges stay in its docker zone). `--containers=false` takes the rules away;
the bridge stays until you remove the runtime's network.

**Container names** (amendment A38). Every node resolves the containers on the mesh networks by
name: `web.mesh` answers every node's `web` containers, rotated on each answer (as Swarm does
across tasks), and `web.nas.mesh` only the NAS's. A container's names are its compose service,
its Swarm service (without the stack's prefix), its container name, its aliases on
`meshctl-mesh`, its `hostname:` and, with `domainname:` set to the zone, `hostname.<zone>`;
names in other domains are published only inside an extra domain (DNS, below). A node's name, a custom record and the server's own
address keep theirs: a container called `vps-1` gets no name, and the DNS page in the web UI
lists every reported name with why one is not published. Names follow the containers within a few
seconds (TTL 10 s) and keep resolving from each node's cache while the control plane is down.

The agent reads them from the runtime's API socket, GET requests only. A host-mode agent finds
Docker's or rootful Podman's socket by itself (`meshctl agent run --runtime-socket off`, or a
path, changes that); a container-mode agent reads the socket its container mounts at
`/var/run/docker.sock` ([containers.md](containers.md#container-names) has the paths per
runtime), and with `MESHCTL_RUNTIME_SOCKET=off` and no mount its containers get no names. On a
host whose runtime labels containers for SELinux, the runtime may refuse the agent's container the
socket; turning the socket off is then the way out.

## 5. Day to day

All node commands run against the server's admin socket: `docker exec <container> /app/meshctl
node ...`. Each change is versioned, audited and pushed to the affected agents at once.

| Task | Command |
| --- | --- |
| see the mesh | `node list`, `node show vps-2` |
| a node's public address changed | `node set vps-2 --endpoint 198.51.100.4:51820` |
| take a node out (lost, compromised) | `node disable vps-3` -- every server drops it within seconds |
| put it back | `node enable vps-3` |
| re-install a node from scratch | `node token vps-3`, then the install command again |
| a phone's new profile (a new key) | `node phone <name>`, `--exit vps-2` for all its traffic |
| remove a node | `node rm vps-3`; on its host, `systemctl disable --now meshctl-agent@default` and then `meshctl agent reset` take its tunnel, routes, rules, firewalld zone and DNS registration off the host (the data dir `/var/lib/meshctl/default` stays until you delete it) |
| move a node to another mesh IP | `node set laptop --mesh-ip 10.77.2.9`, or Mesh IP on its page: peers, its DNS name and its agent follow; the answer lists what does not (a server's Swarm advertise address, a phone's profile, a firewall zone, object or policy naming the old address) |
| put a node in firewall groups | `node add laptop --role spoke --groups laptops,office`, `node set laptop --groups laptops`, or Groups on its page |

**IPv6** (amendment A46). A mesh carries both families by default: its server's first start
gives it a random unique local /64 (`meshctl ipv6` shows it, and the Settings page), and every
node whose family allows IPv6 an address in its role's block beside its IPv4 one. `meshctl ipv6
fd12:3456:789a::/64` picks one of your own (any /48 to /64) while no node holds an address in
the current one -- right after the first start, say; `meshctl ipv6 auto` draws a new one after
`off`. A node's family is `dual` (default), `ipv4` or `ipv6`: `node add nas
--role spoke --family ipv4` for a host without usable IPv6, `node set nas --family dual` later;
the address a family adds is allocated, the one it drops is freed. `--ip6` and `--mesh-ip6` pick
an IPv6 address as `--ip` and `--mesh-ip` do; `--endpoint6 [2001:db8::10]:51820` publishes a
server's IPv6 endpoint, which spokes dial when IPv4 does not work (a server's join finds it
itself, or takes `agent join --endpoint6`); `--containers6` puts a node's
containers' IPv6 on the mesh beside their IPv4: its /112,
the upper half handed out; `node show` then lists the IPv6 values too, the mesh network is
created dual stack (`--ipv6 --subnet <prefix> --ip-range <half> --gateway <::1>` after the IPv4
values), and a container-mode agent's container sets `disable_ipv6=0`, IPv6 forwarding and proxy
NDP in its namespace ([containers.md](containers.md)). Recreate the network and the agent's container after turning it on. The agent isolates
the containers' IPv6 as their IPv4 and answers neighbour discovery for the mesh's nodes; their
IPv6 reaches every node, and the named containers of other nodes (by the mesh view's names: with
the runtime socket off a node's containers are reached over IPv4 only). It needs the kernel's IPv6
policy routing: where `ip -6 rule` fails (some NAS kernels), the agent drops the containers'
IPv6, the containers step reports degraded and they keep IPv4 only -- turn `--containers6` off
there. `meshctl ipv6 off` takes every IPv6 address away
again (make IPv6-only nodes dual first, and remove IPv6-only DNS records), and the mesh stays
IPv4 only across restarts until `meshctl ipv6 auto` or a prefix. A prefix that holds
addresses is changed by going off first. When a node's agent reports that its host cannot use
one of its families, the mesh leaves that family out for it (a warning on its page and in `node
show`, which says what the host can use) until the agent reports it works again.

**Several meshes** (amendment A49). One control plane holds up to 64 meshes; the first is
`default`. Each has its own nodes, settings, DNS zone, firewall, history and nightly exports, and
an agent belongs to one (a host in two meshes runs two agents).

| Task | Command |
| --- | --- |
| see the meshes | `mesh list`, `mesh show lab` |
| add one | `mesh add lab` -- the first free 10.x.0.0/16 from 10.77.0.0/16 up, zone `lab`, a random unique local /64; or `--mesh-cidr 10.90.0.0/20 --dns-zone lab.example --ipv6 off` |
| work in it | `--mesh lab` on every `node`, `firewall` and `ipv6` command, and on `export` and `import`: `node add --mesh lab nas --role spoke` |
| rename it, change its settings | `mesh set lab --name office`, `mesh set lab --dns-zone office --default-mtu 1380 --upstream-dns 9.9.9.9` |
| remove it | `node rm --mesh lab <name>` for each node, then `mesh rm lab` |

With more than one mesh, a command without `--mesh` says which there are. Meshes share no network
and no DNS domain: a mesh's IPv4 network and IPv6 prefix may not overlap another's, nor its zone
or extra domains equal, sit inside or hold another's, since agents of two meshes may run on one
host. On the web UI each mesh's pages are under `/m/<mesh>/`; its Meshes page lists, adds,
renames and removes them.

A host in two meshes runs one agent per mesh (amendment A51): install.sh with each mesh's token
and `--mesh` installs `meshctl-agent@<mesh>`, with the data dir `/var/lib/meshctl/<mesh>` and the
socket `/run/meshctl/<mesh>.sock`; the host's commands then take `--mesh` (`meshctl status --mesh
lab`). Every name an agent makes on the host comes from its mesh's slot (`mesh show` prints it), so
the agents never meet:

| Slot | 0 (`default`) | s (e.g. 1) |
| --- | --- | --- |
| WireGuard interface | `wg0` | `wg<s>` (`wg1`) |
| a server's default port | 51820/udp | 51820+s (51821) |
| local UI (`--ui-listen auto`) | 127.0.0.1:7443 | 7443+s (7444) |
| exit mode | rules 32764/32765, table 51820 | 10*s lower (32754/32755), table 51820+2s (51822) |
| container network | rules 1000../5200.., table 51821, bridge `meshctl0`, runtime network `meshctl-mesh` | 10*s higher (1010../5210..), table 51821+2s, `meshctl<s>`, `meshctl-mesh<s>` |
| firewalld | zone `mesh`, policies `meshctl-relay`, `meshctl-exit` | `mesh<s>`, `meshctl<s>-relay`, `meshctl<s>-exit` |
| mesh firewall | `table inet meshctl` | `table inet meshctl<s>` |

The public zone is shared: each server agent opens its own port there. A node keeps the port it
was added with, so a mesh imported from an export keeps its 51820; on a host already holding it,
`node set --mesh <mesh> <name> --listen-port 51821`. One mesh at a time carries a host's internet:
`meshctl exit` on a second mesh leaves its exit off, reported as degraded `exit` with the interface
that holds it, until the first turns its exit off. Every mesh's tunnel bypasses that exit (they
share the fwmark). A host in two meshes is reachable from both, but its tunnels carry no traffic
from one mesh into the other: the other mesh's peers drop any packet whose source is not the
sender's own (WireGuard's allowed IPs).

**Relay failover** (amendment A40). Two spokes with different hubs talk through one of them (the
lower-id spoke's hub). When that hub goes down, the server moves the pair to the other spoke's hub,
else the default hub, else the lowest-id server that is up, as a new config version audited
`relay.failover` by `system`; the dashboard shows the hub "down since ...". A hub counts as down
when its agent has not reported for 2 minutes AND no node has seen a handshake with it for 3: an
agent that merely stopped, or a server that cannot reach the hub, moves nothing while its tunnels
work. It comes back (`relay.restore`) once its agent has reported steadily for 5 minutes, so a
flapping hub does not bounce the pairs. `meshctl control --hub-down-after 2m` (`MESHCTL_HUB_DOWN_AFTER`) sets the first
threshold (the others follow it; `0` turns failover off) and must stay above the agents'
heartbeat (`agent run --status-every`, 30 s). The down hub stays every node's peer, so traffic to
the hub itself resumes on its own. Failover needs the server: run it on a host other than the
hubs it protects, or accept that losing that host stops failover too. Phones do not move: their
profile names one hub.

On a node, `meshctl status` (root, or a member of group `meshctl`) shows the applied version, the
DNS state and each peer's endpoint, last handshake and traffic. `meshctl logs` prints the agent's
last 2000 log lines and `meshctl logs -f` follows them: the journal's lines (`docker logs` for the
NAS container), readable by group `meshctl` too.

## 6. DNS

Every server and spoke answers the `mesh` zone on its mesh IP, port 53 (and on its IPv6 mesh IP,
with AAAA records beside the A ones, on a dual-stack mesh): one name per node
(`vps-1.mesh`, `nas.mesh`) and the custom records of the web UI's DNS page -- an alias that
follows a node's mesh IP, or a fixed mesh IP. The answers come from the node's cached config, so
`nas.mesh` keeps resolving while the server is down. Names of the other zones are REFUSED, so
the host keeps its usual resolver for them.

The zone is `mesh` until Settings names another (amendment A36): `home.arpa` or a name under a
domain you own, never `local` (mDNS) or a public TLD you do not control. Changing it renames every
node's names at once (`nas.mesh` becomes `nas.home.arpa`) and every node picks it up with its next
config; the examples below use the default. Custom records are given without the zone: `files`,
not `files.mesh`.

The host's own lookups of `.mesh` go there through systemd-resolved, on wg0 only (`resolvectl
status wg0` shows the mesh IP and `~mesh`). Nothing else on the host changes; `/etc/hosts` and
`/etc/resolv.conf` are never written. `meshctl dns` on the node, and its page in the web UI, show
the state:

| State | Meaning |
| --- | --- |
| `on` | the host resolves `.mesh` names |
| `off` | turned off on the node with `meshctl dns off` (kept across restarts; `meshctl dns on` undoes it). The resolver still answers the other nodes |
| `unavailable` | no systemd-resolved on the host. The resolver answers the others; the host itself does not use it |

AlmaLinux does not run systemd-resolved by default. To use mesh names on a VPS:

```
dnf install systemd-resolved && systemctl enable --now systemd-resolved
printf '[main]\ndns=systemd-resolved\n' > /etc/NetworkManager/conf.d/dns.conf
ln -sf ../run/systemd/resolve/stub-resolv.conf /etc/resolv.conf
systemctl reload NetworkManager && systemctl restart 'meshctl-agent@*'
```

In exit mode every lookup goes through the tunnel: resolved sends all of them to the node's
resolver (`~.` on wg0), which forwards the other zones to the upstream resolvers (Settings;
1.1.1.1 and 9.9.9.9 by default), so nothing leaks to the local network's resolver. With `dns off`
in exit mode, lookups go to that resolver, outside the tunnel. Phones use their hub's resolver,
which forwards for them.

Containers have their own resolver: one attached to a node's mesh network uses the agent's (above,
Container mode) and resolves every mesh name and the containers' names; any other container does
not see `.mesh` names, so use mesh IPs there.

**Extra domains** (Settings, amendment A39) are domains the mesh takes over besides the zone,
such as `example.lan`: every node answers the containers' names in them -- a container with
`hostname: web` and `domainname: example.lan`, or an alias `web.example.lan` -- and forwards every
other name in them to the upstream resolvers, and each host-mode node's resolver registration
sends the whole domain to the mesh (`~example.lan` in `resolvectl status wg0`, an NRPT rule on
Windows). So a host stops asking its own resolver for those names: choose domains the mesh owns,
not your LAN's or your company's. At most 8; never the zone, inside it or around it, no `.arpa`
but `home.arpa`, and never the domain of the server's own address. The DNS page lists them with
every container name and how it resolves.

## 7. Firewall

On servers the agent owns firewalld's `mesh` zone on wg0 (target DROP, one rule accepting the mesh
CIDR) and what it opened in the public zone (udp/51820); what crosses the mesh is the mesh
firewall's to decide (section 8, "The mesh firewall"), in a table of the agent's own. Rules you
make yourself elsewhere are never touched; to change what the mesh allows, use the Firewall page
or `meshctl firewall`, not firewall-cmd.
Spokes keep their own firewall: on the NAS, allow your apps' ports from the mesh CIDR in its
firewall settings.

Every 5 minutes the agent applies its config again after asking firewalld what it holds: wg0
moved to another zone at runtime (NetworkManager does this), a rule removed by hand, a route or
wg0 itself deleted are put back and logged as `drift check: put back what changed`.

## 8. The web UI

The UI is served at the server's URL (`https://mesh.example.com/`), beside the agent API. Make
the first admin right after the first start, in the browser or on the host (amendment A52):

- **In the browser.** While there is no admin, the control plane logs a setup code at every start
  (Dokploy: the app's logs; or `docker logs <container>`):
  `no admin yet: open the setup page and enter this setup code ... setup_page=https://mesh.example.com/setup setup_code=abcd-efgh-...`.
  Open that page (the sign-in page links it too), enter the latest code, a username and a password:
  the code works once, and only until an admin exists.
- **On the host:** `docker exec -it <container> /app/meshctl user add admin`.

Either way the first sign-in then shows a QR code for an authenticator app (any TOTP app) and asks
for one code to confirm it.

Each mesh's pages are under `/m/<mesh>/` (amendment A49): the bar names the mesh, and with more
than one a row below it switches between them; `/` opens the first mesh.

| Page | What it does |
| --- | --- |
| Meshes | (admins) every mesh (open it, rename it, remove one without nodes); add one with its name and, optionally, its IPv4 network, DNS zone and IPv6 prefix (`auto`, `off` or a prefix) |
| Dashboard | every node with its state (green in sync, amber applying / degraded / late, red silent or rolled back, grey disabled or not joined; a phone green while its tunnel is on, amber when its profile is out of date, grey idle), applied and current version, agent, last heartbeat (a phone's last handshake), handshake age per peer and exit; refreshes every 10 s |
| Nodes | add a node (the join command is shown once), open one to see its computed config, peers and DNS state, edit it, disable / enable, issue a new token, make a phone's profile (a QR code, shown once), remove it |
| DNS | the mesh zone: one name per node and the custom records (add, remove); the upstream resolvers |
| Firewall | the mesh firewall (below): policies in order (add, edit, move, remove; each optionally scheduled) and the time zone of the schedules, zones with their inbound and outbound rules, objects (groups, addresses, services), and Reachability: check one flow, now or at a time, see how every zone reaches every other |
| Versions | every config version with what it changed; roll back to one (a new version; a rollback that adds or removes nodes asks for confirmation) |
| Audit | who changed what, newest first |
| Settings | the mesh's default MTU, default hub, upstream DNS, DNS zone, extra domains and IPv6 prefix; the mesh CIDR only while the mesh is empty; download its export (`<mesh>-YYYY-MM-DD.yaml`) |
| Users | (admins) add users, an admin or not; grant and revoke their meshes; make or stop making them admins; disable them, reset their authenticator |
| Account | change your password (signs you out everywhere) |

Forms are in sections and show only the fields that apply to what is chosen: a server's endpoint
and port for a server, the hub and keepalive for a spoke or phone, the containers' options for a
node that can have them, the IPv6 fields for a family that has IPv6. A hidden field is not sent.
Without JavaScript every field shows, and the control plane checks the form as always.

Changes made in the UI go through the same path as `meshctl node`: versioned, audited under
your user, pushed to the agents at once. The dashboard's 10 s refresh does not keep a session
alive; 12 hours without using the UI signs you out, and 7 days always do.

### The mesh firewall

Every server enforces the mesh firewall (amendment A43) on the mesh traffic it relays, receives and
sends -- and every packet between two nodes crosses a server, so the policies segment the whole
mesh. A mesh starts with one policy, `allow-all`; narrow it from there:

1. Put nodes in **groups**: `node add ... --groups laptops`, `node set nas --groups storage`, or the
   node's page.
2. Make **zones** (Firewall, Zones): their members are `node:nas`, `containers:nas` (its container
   network), `group:laptops`, `role:server|spoke|phone`, `addr:<object>` and addresses
   (`10.77.1.5`, `10.77.1.0/24`, `10.77.1.5-10.77.1.20`). Within a zone traffic is allowed unless
   the zone says deny.
3. Write **policies** (Firewall, Policies), first match wins: `allow from zone:laptops to node:nas
   services tcp/445`, then `deny from zone:laptops to zone:servers`, ... Services are `tcp/22`,
   `udp/1000-2000`, `tcp`, `icmp` or `svc:<object>`; empty is any.
4. Change or remove `allow-all` (last) when the policies above it say everything that should pass:
   what no rule allows is dropped.
5. Optionally give a policy a **schedule** (amendment A44): days (none is every day) and a window,
   `09:00`-`18:00`, or `22:00`-`06:00` across midnight (none is the whole day). Outside it the
   policy is not in force and the next one decides. Schedules are read in one time zone, the
   server's unless the Firewall page (or `timezone:` in the YAML) names another, such as
   `Europe/Berlin`. Each server switches at the edge by its own clock, with the control plane down
   too, and a window that closes cuts the connections it allowed -- so keep the servers on NTP.

Ping across the mesh, DNS to the servers and Swarm between them always pass (system rules). Check
before and after: Firewall, Reachability (one flow, or the zone matrix), or

```
docker exec <container> /app/meshctl firewall check laptop nas tcp/445
docker exec <container> /app/meshctl firewall show
docker exec <container> /app/meshctl firewall show --yaml > firewall.yaml   # edit, then
docker exec -i <container> /app/meshctl firewall apply /dev/stdin < firewall.yaml
```

In the YAML a schedule reads:

```
segmentation:
  policies:
    - name: office-smb
      from: [group:laptops]
      to: [node:nas]
      services: [tcp/445]
      action: allow
      schedule:
        days: [mon, tue, wed, thu, fri]
        start: "08:00"
        end: "19:00"
    - name: allow-all
      action: allow
  timezone: Asia/Singapore
```

`firewall check ... --at 2026-10-24T10:00:00+08:00` answers for that time, and names the
scheduled policies not in force then.

A change applies to new connections at once; the servers re-check existing mesh connections too.
On a server, `nft list table inet meshctl` shows what it enforces (the first rule's comment carries
the policy's fingerprint), and `meshctl status` there reports a policy the kernel refused as the
error of the config it could not apply.

### Each node's local UI

Every agent serves a small UI of its own on `127.0.0.1:7443`: the node's status and peers, the
mesh as the server last showed it (still readable, marked with the server's absence, while the
server is down), the node's own exit, keepalive, MTU and split DNS, a restart of the tunnel
(every peer handshakes anew, nothing else goes down), and the agent's log. It has no accounts: on
the node, as root or a member of group `meshctl`, run `meshctl ui` (on the NAS, `docker exec
<container> /app/meshctl ui`) and open the link it prints within 2 minutes. The link works once
and signs that browser in for up to 12 hours; an idle hour, `Sign out` or an agent restart end it.
From another machine, forward the port first and open the link there:

```
ssh -L 7443:127.0.0.1:7443 vps-1     # then, on vps-1: meshctl ui
```

The agent of a mesh in slot s (amendment A49) listens on 7443+s instead (`--ui-listen auto`, the
default); `meshctl ui` prints the port to forward. `agent run --ui-listen 127.0.0.1:7450` moves
it (loopback addresses only) and `--ui-listen off` turns it off. An exit, keepalive or MTU change asks the server, exactly like `meshctl exit`;
split DNS stays on the node.

## 9. Users and recovery

Accounts are managed on the server host; the commands open the database directly, so they work
while the server is stopped -- the way back in when nobody can sign in. They read passwords from
the terminal only, so use `-it`:

```
docker exec -it <container> /app/meshctl user add admin        # enrolls TOTP at first sign-in
docker exec -it <container> /app/meshctl user passwd admin     # forgotten password
docker exec -it <container> /app/meshctl user reset-totp admin # lost phone: enroll again
```

`passwd` and `reset-totp` end the user's sessions.

Who manages what (amendment A49): an admin manages every mesh, the meshes themselves and the
users; anyone else manages only the meshes granted to them, and sees no other. A mesh needs no
user of its own: the control plane's admins (and this host's commands) manage it. The first user
is always an admin (so `user add` refuses `--mesh` for them), and the control plane keeps at least
one enabled admin: the last one can neither lose the role nor be disabled.

| Task | Command (or the Users page) |
| --- | --- |
| add an admin | `user add boss --admin` |
| let a user manage a mesh | `user add ops --mesh lab`; more meshes with `user grant ops --mesh home` |
| take it away | `user revoke ops --mesh lab` |
| make an admin, or stop being one | `user grant ops --admin`, `user revoke ops --admin` (the last admin stays one) |
| see who manages what | `user list` |
 Five failed sign-ins from one address lock it
out for 15 minutes; twenty for one username within an hour lock the username.

## 10. Backups

The server writes a nightly export of each mesh at 03:00 (container time) to
`/data/exports/<mesh>/mesh-YYYY-MM-DD.yaml`, skipping nights when nothing changed and keeping the
newest 14 of each mesh. For an export on demand:

```
docker exec <container> /app/meshctl export > mesh.yaml
```

(`--mesh lab` for one of several meshes.)

An export is the desired state as YAML -- nodes with their public keys, settings, custom DNS
records and the mesh firewall; never tokens, private keys or users. An export made before the
mesh firewall imports its custom firewall rules as policies. It reads the database directly
and works while the server runs. Commit it to git: an unchanged mesh exports byte-for-byte the
same file.

To restore an export (a new server, or undoing a bad day), stop the server and import it; node
ids are kept, so the agents keep working without re-joining:

```
docker compose stop meshctl
docker compose run --rm meshctl import /data/exports/default/mesh-2026-10-12.yaml
docker compose start meshctl
```

`import` refuses to run while the server holds the data dir, and checks every value of the file
before it changes anything. `import --mesh lab` replaces that mesh, and creates it first when the
control plane has none of that name (with the file's network and zone; it goes again when the
import fails). Node ids belong to the whole control plane: an export whose node ids another mesh
holds is refused. Users are not in an export: for a full restore, including users and
`secret.key` (which seals their TOTP secrets), keep a copy of the `meshctl-data` volume as well
(`docker run --rm -v meshctl-data:/data -v $PWD:/b alpine tar czf /b/meshctl-data.tgz -C /data .`,
with the server stopped). The volume holds `config.key` too, the key every config is signed with
(amendment A50): agents trust no other, so a control plane restored without it signs with a new
one and every node must join again with a new token. A rollback or an import never changes what a
node brought at its join -- its keys, whether it is enrolled, its platform -- so a node cut off by
a token stays cut off; a node the file re-creates gets the keys the file has, and joins again if
it joined since.

## 11. When the control plane is down

Agents keep their last config (`/var/lib/meshctl/<mesh>/config.json`) and re-apply it at boot without
the server, so the mesh survives a server outage and reboots, and every node keeps answering
`.mesh` names -- on the machine the data dir last ran on (`host.json`), never on another
(amendment A50). Each node also keeps the last view of the mesh it fetched (`mesh.json`, at most
5 minutes old when the server went down), which it shows marked with the server's absence. Changes
simply wait, and so does relay failover: a hub that goes down meanwhile keeps its pairs until the
server is back. When it is, agents reconnect within a minute.

## 12. Troubleshooting

`journalctl -u meshctl-agent@<mesh>` or `meshctl logs` on a node, `docker logs <container>` on the control plane.

| Symptom | Cause and fix |
| --- | --- |
| `clock_skew` in the agent log | the node's clock is off by more than 60 s; enable NTP (`timedatectl set-ntp true`) |
| `node_disabled` | an operator disabled the node; `node enable` it |
| `token_invalid ... already used` / `expired` | tokens are single-use and live an hour: `node token <name>` |
| `interface wg0 exists but was not created by meshctl` | a wg-quick setup owns wg0: `wg-quick down wg0` and disable it, then restart the agent |
| `the agent could not take everything it made off this host` | a refused agent's teardown failed at the step the error names (firewalld stopped, say): fix that, then `meshctl agent reset` |
| `another meshctl process uses this data dir ... stop the agent first` | `agent reset` and `agent join` need the agent stopped: `systemctl stop meshctl-agent@<mesh>` (Windows: `Stop-Service meshctl-agent-<mesh>`) |
| `no node has joined a mesh on this host` / `this host runs the agents of several meshes` | the host's commands find the agent by its mesh: pass `--mesh <mesh>` (or `--data`) |
| `this token belongs to a node of the mesh "lab", not "default"` | the install command names the token's mesh: run the one `node add` or `node token` printed (`--mesh lab`, `MESHCTL_MESH=lab`) |
| `this host has no machine id` | a container without the host's `/etc/machine-id` mounted (`/etc/machine-id:/etc/machine-id:ro`), or a host without one (`systemd-machine-id-setup`); `--machine-id-file` (`MESHCTL_MACHINE_ID_FILE`) names another file |
| `another copy of this node holds its session; waiting` | another agent of this node runs on this machine: stop one of them. Within 90 s this one is refused (`duplicate`) if the other answers, or takes over if it is gone |
| `node ... runs on another machine` (`foreign`) | the data dir came from another machine: this copy wiped itself. Issue a token (`meshctl node token <name>`) and join here |
| `another agent of this mesh already runs here` | a second agent of the mesh in this network namespace: a copy of the data dir in another directory, or another node of the mesh; stop it |
| `firewall` listed as degraded | firewalld is missing or stopped; the tunnel works, but Swarm and udp/51820 may be blocked. Install/start firewalld, or open the ports yourself and run the agent with `--no-firewall` |
| `create WireGuard interface wg0: ... is the wireguard kernel module available?` | `modprobe wireguard`, or run WireGuard in the agent: `agent run --wg-backend userspace` (`auto`, the default, picks it when the module is missing) |
| `MESHCTL_MODE=server is gone` (or `=agent`) | a file from before amendment A48: the control plane sets `MESHCTL_MODE=control`, a node leaves it out |
| `MESHCTL_MODE must be control or node` | the container's environment has another value: `control` for the control plane, nothing for a node |
| `... MESHCTL_MODE is not set, so this container runs a node` | a control plane's container without `MESHCTL_MODE=control`: set it |
| `has not joined a mesh and has no join token` | a node container's first start without a token: `MESHCTL_TOKEN=<token> docker compose up -d` (or `MESHCTL_TOKEN` in `.env`) |
| `... MESHCTL_TOKEN_FILE names ..., which does not exist or is empty` | a node that has not joined yet found no token file: write the token there, or set `MESHCTL_TOKEN` (a joined node runs on without the file) |
| `on one machine with box of this mesh` | two nodes of one mesh run on one machine (amendment A53): keep one, and use host mode's `--containers` for the containers |
| `the server refused this join token on an earlier start` | the token was spent (used or expired) before the node joined: issue a new one with `meshctl node token <name>` and start with it |
| `userspace WireGuard needs /dev/net/tun` | the container has no tun device: map it (`devices: ["/dev/net/tun:/dev/net/tun"]`), or set `MESHCTL_WG_BACKEND=kernel` if the host has the module |
| `wg show` does not list a userspace wg0 | the agent could not create `/var/run/wireguard/wg0.sock` (its log says why; a container has no `wg` anyway): use `meshctl status` |
| `another node already uses this WireGuard or identity key` | a data dir was copied between hosts; remove `/var/lib/meshctl/<mesh>` on the copy and join it with its own token |
| a node stays `applying` | it applied an older version and is failing the new one; `meshctl status` on it shows the error |
| `exit "vps-3" is not exit-capable` | allow it first: `node set vps-3 --can-exit` |
| `only spokes choose an exit or a keepalive` | `meshctl exit` ran on a server; exits are for spokes |
| `apply step exit: ... operation not permitted` | the agent lacks root or CAP_NET_ADMIN; run it as the installed service |
| a server shows `rolled back` | safe apply: no heartbeat was answered within 60 s of a change, so the previous config is back. See what the change did (endpoint, MTU, firewall), fix it, and make the next change |
| `the server accepted the change, but the agent has not applied it` | `meshctl status` on the node shows the error the apply hit |
| `the data dir is in use: stop the meshctl control plane first` | `import` ran beside a running control plane; stop it, import, start it |
| `... secret.key is readable by other users` | the key that seals TOTP secrets lost its mode: `chmod 600 /data/secret.key` |
| `this request did not come from https://...` in the UI | the UI was opened under another address than `--url`; open it at that URL |
| `too many failed sign-ins` | wait 15 minutes, or reset the password with `meshctl user passwd` |
| DNS `unavailable` | the host runs no systemd-resolved (AlmaLinux's default): the node answers the others, the host does not use it. See section 6 |
| `dns` listed as degraded, `is another DNS server using port 53 on this address?` | something else listens on port 53 of the mesh IP or of every address (a DNS app on the NAS); bind it to another address |
| `DNS record "nas" has the name of a node` | a custom record cannot take a node's name; that name resolves already |
| `no server has joined with an endpoint yet` for a phone | a phone's profile lists the servers it dials; join a server first |
| `... cannot be this phone's exit` | the exit must be a joined server with an endpoint that can be an exit: `node set vps-2 --can-exit` |
| a phone shows `profile out of date` | the servers or its hub changed under its static profile: make a new profile and import it |
| a phone resolves nothing while its tunnel is on | its profile sends every lookup to its hub's resolver: that server is down, or the phone's hub changed (make a new profile) |
| `fetch the mesh view failed; will retry` | the server is unreachable (the node shows its cached view meanwhile), or older than the agent and has no mesh view yet: upgrade the server |
| `this reader fell behind the log and was dropped` | `meshctl logs -f` was read too slowly (a paused terminal); run it again |
| `the local UI is not running: ... address already in use` | another program has port 7443: `agent run --ui-listen 127.0.0.1:<port>` (and forward that port) |
| `the local UI is off on this node` | the agent runs with `--ui-listen off` |
| `This sign-in link does not work` | links work once, within 2 minutes: run `meshctl ui` again |
| `This is meshctl's local UI; open it at http://127.0.0.1:7443/` | the browser used another name for it (a forward to another local port, the host's name): forward to local port 7443 and open the link as printed |
| `This change did not come from the local UI's own page` | something other than the UI's page posted to it (or a browser extension strips Origin); use the page itself |
| `... is this prompt elevated (Run as administrator)?` on Windows | the local API pipe is for Administrators and SYSTEM: open the prompt with "Run as administrator" |
| `install.ps1: run it from an elevated PowerShell` | install.ps1 installs a service and writes to Program Files: run it as administrator |
| `this meshctl was built without wintun.dll` | a binary from `go build` without `-tags wintun`: use the release binary from `/dl/`, or `./scripts/dev.ps1 build` |
| `the wintun.dll in this build is not the pinned 0.14.1` | the binary was built from a DLL that is not the release: rebuild with `./scripts/dev.ps1 build`, which fetches and checks it |
| the Windows service stops and starts again every few seconds | read `C:\ProgramData\meshctl\<mesh>\agent.log`; the agent exits with what failed and the service manager restarts it |
| `drift check: put back what changed` every few minutes | something keeps changing the host's mesh setup, usually NetworkManager managing wg0: `nmcli device set wg0 managed no` |

## 13. Upgrading

Server: bump `MESHCTL_VERSION` and redeploy. At start the server recomputes every node's config;
if the new release computes differently, the agents receive the change like any other.

Agents: replace `/usr/local/bin/meshctl` with the new release and `systemctl restart
'meshctl-agent@*'` (Windows: `Restart-Service 'meshctl-agent-*'`), or run the install command again with a fresh token (`node token <name>`) -- the
token cuts the node off until the new join (below).

**A new token cuts the node off** (amendment A50). `meshctl node token <name>` (or the node page)
retires the node's current key at once: its peers drop it, and its running agent is told, signed,
that it was replaced -- it takes its tunnel, routes, rules, firewalld zone and DNS registration off
the host and stops (`agent run` exits 3, which the systemd unit does not restart; on Windows the
service stops). When a step of that teardown fails, the log says so and `meshctl agent reset`
(with the agent stopped; `--data` for another mesh's agent) tries again. The node comes back
when it joins with the new token, with fresh keys: every join makes new ones. The node page asks
first when the node was seen in the last 5 minutes; `node token` says it was. A copy of an old
data dir, from any earlier join, is told the same, tears down and stops.

**One node, one running copy** (amendment A50). A node runs on one machine at a time, once:

- A second agent on the same data dir, or on a copy of it in the same network namespace (or, on
  Windows, the same host), is refused before it touches anything.
- A copy elsewhere on the same machine (a second container with the same volume, a cloned VM)
  waits while the control plane asks the running one whether it is alive, and is refused when it
  answers: the copy takes down what it brought up and stops (exit 3), keeping its data dir. If the
  running one does not answer within 90 seconds -- it is gone -- the copy takes over.
- A copy on another machine (the data dir restored onto a new box, a disk moved) is refused, takes
  down what it brought up, wipes its data dir and stops. To move a node to a new machine, issue a
  token (`meshctl node token <name>`) and join there. A container restarted with the old token is
  told it needs a new one.
- A restart, a crash, a reboot or a control plane restart keeps the node's session: it carries on
  at once. A clean stop ends the session, so the next start is let in at once too.

The dashboard and `node show` say when a second copy waited or was refused, and where it ran; the
audit log has `node.session`, `node.duplicate`, `node.foreign` and `node.replaced`. Every node's
container mounts the host's `/etc/machine-id` (docs/containers.md); a host without one makes it
with `systemd-machine-id-setup`, or names another file with `--machine-id-file`.
