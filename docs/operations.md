# Operations

Running a mesh: bringing up the server, enrolling the VPS nodes, laptops and exits, day-to-day
node changes, phones, DNS and the firewall, the web UI and each node's local UI, backups and the
usual failures. Windows is under way in M5 (spec S15): the service and its local API are in
place, WireGuard there is not yet. Every command and flag is in
[commands.md](commands.md).

## 1. Run the server (Dokploy)

The server is one container behind Dokploy's Traefik. Create a compose app from
[`deploy/compose.server.yml`](../deploy/compose.server.yml) and set its environment:

| Variable | Value |
| --- | --- |
| `MESHCTL_VERSION` | the release to run, e.g. `1.0.0` (the image tag, without the `v`) |
| `MESHCTL_URL` | the public origin agents use, e.g. `https://mesh.example.com` |
| `MESHCTL_TRUSTED_PROXIES` | optional: the CIDR of `dokploy-network`, so the server sees real client addresses |

Point the domain at the service's port 8080 in Dokploy. Nothing is published on the host. The
container runs as uid 65532 with no capabilities and a read-only root file system; the database
and the admin socket live in the `meshctl-data` volume at `/data`.

Check it:

```
curl -fsS https://mesh.example.com/healthz          # ok
docker exec <container> /app/meshctl node list      # empty table
```

If a GeoIP allowlist fronts the domain, `/api/v1/agent/*`, `/install.sh` and `/dl/*` must stay
reachable from every node's country.

## 2. Enroll the servers

For each VPS, on the Docker host:

```
docker exec <container> /app/meshctl node add vps-1 --role server
```

It prints a join token (valid for an hour, shown once) and the command to run on the VPS, as
root:

```
curl -fsSL https://mesh.example.com/install.sh | sudo sh -s -- --token <token>
```

install.sh downloads meshctl from your server, checks its SHA-256 against the value baked into
the script, installs it to `/usr/local/bin`, joins, and installs and starts
`meshctl-agent.service`. The agent creates `wg0` with the node's mesh IP, opens udp/51820 in
firewalld's public zone, the Swarm ports from the other servers and DNS from the mesh in the
`mesh` zone, and answers `.mesh` names on its mesh IP (section 6).

The server's public endpoint is detected from the VPS's own interfaces. When that is wrong (a
NAT'd VPS), pass it: `sudo sh -s -- --token <token> --endpoint 203.0.113.10:51820`, or set it on
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

## 4. Spokes and exits

A spoke is a node without a public address: a laptop, the NAS. It dials every server and keeps
the tunnel open with a 25 s keepalive. Traffic between two spokes goes through one server, the
same one in both directions (amendment A3): the hub of the spoke that was added first.

**A Linux laptop.** On the server host:

```
docker exec <container> /app/meshctl node add laptop --role spoke --hub vps-1
```

then, on the laptop as root, the install command it prints. The laptop gets an address from
10.77.2.0/24; `--hub` names the server that relays for it (default: the mesh's default hub, else
the first server). A host without the WireGuard kernel module runs WireGuard inside the agent
instead (wireguard-go on `/dev/net/tun`): the agent picks that itself (`agent run --wg-backend
auto`), and `meshctl status` names the backend in use.

**The NAS (a container).** On the server host:

```
docker exec <container> /app/meshctl node add nas --role spoke --always-on
docker exec <container> /app/meshctl node compose nas > compose.yml
```

`--always-on` gives it an address from 10.77.1.0/24. The compose file -- also a download on the
node's page in the web UI -- runs the agent from the server's own release image, on the host
network, with the posture docs/security.md describes. Choose its WireGuard with `--wg-backend`:
`userspace` (the default) needs only the host's `/dev/net/tun`; `kernel` uses the NAS's own module
and maps no device, if `ls /sys/module/wireguard` finds it. `--host-dir` (default
`/volume1/docker/meshctl`) is where the agent keeps its keys and cached config. Copy the file to
the NAS, join once with the token `node add` printed, then start it:

```
MESHCTL_TOKEN=<token> docker compose run --rm -e MESHCTL_TOKEN meshctl-agent \
  agent join --data /data --server https://mesh.example.com --name nas --no-service
docker compose up -d
docker compose exec meshctl-agent /app/meshctl status --data /data
```

A userspace interface lives as long as the agent, so the NAS is off the mesh while its container
restarts (a kernel interface stays up). The agent changes no firewall on a spoke: allow the NAS's
app ports from 10.77.0.0/16 in its own firewall (spec S7.5).

**A Windows laptop** (M5, in progress: WireGuard and the mesh routes work; exit mode and split DNS
on Windows come next). Add it as a spoke on the server, then in an elevated PowerShell
(Run as administrator) on the laptop, with `meshctl-windows-amd64.exe` downloaded from the
server's `/dl/`:

```
$env:MESHCTL_TOKEN = '<token>'
.\meshctl-windows-amd64.exe agent install --server https://mesh.example.com --name laptop2
```

That copies meshctl to `C:\Program Files\meshctl\meshctl.exe`, joins, and installs and starts
the `meshctl-agent` service, which Windows starts at boot and restarts when it fails. Its data
is in `C:\ProgramData\meshctl` (SYSTEM and Administrators only), its log in `agent.log` there
(10 MiB, plus one older file). `meshctl status`, `exit`, `dns`, `logs` and `ui` reach it over the
named pipe `\\.\pipe\meshctl-agent`, from an elevated prompt only. `meshctl agent uninstall` stops
and removes the service (`--purge` removes the data dir too, so the node must join again); the
binary stays in Program Files.

The agent runs WireGuard itself (wireguard-go) on a wintun adapter named `wg0`, which exists while
the service runs. meshctl carries the wintun driver and writes `wintun.dll` beside
`meshctl.exe`, checked against its pinned hash at every start. Windows Defender Firewall treats
`wg0` as a public network: other nodes reach the laptop's services only where you allow them.

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
and IPv6 to the internet is blocked rather than sent around the tunnel.

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
  included, goes through that server.

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
| remove a node | `node rm vps-3` |

On a node, `meshctl status` (root, or a member of group `meshctl`) shows the applied version, the
DNS state and each peer's endpoint, last handshake and traffic. `meshctl logs` prints the agent's
last 2000 log lines and `meshctl logs -f` follows them: the journal's lines (`docker logs` for the
NAS container), readable by group `meshctl` too.

## 6. DNS

Every server and spoke answers the `mesh` zone on its mesh IP, port 53: one name per node
(`vps-1.mesh`, `nas.mesh`) and the custom records of the web UI's DNS page -- an alias that
follows a node's mesh IP, or a fixed mesh IP. The answers come from the node's cached config, so
`nas.mesh` keeps resolving while the server is down. Names of the other zones are REFUSED, so
the host keeps its usual resolver for them.

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
systemctl reload NetworkManager && systemctl restart meshctl-agent
```

In exit mode every lookup goes through the tunnel: resolved sends all of them to the node's
resolver (`~.` on wg0), which forwards the other zones to the upstream resolvers (Settings;
1.1.1.1 and 9.9.9.9 by default), so nothing leaks to the local network's resolver. With `dns off`
in exit mode, lookups go to that resolver, outside the tunnel. Phones use their hub's resolver,
which forwards for them.

Containers (Docker, the NAS's apps) have their own resolver and do not see `.mesh` names; use mesh
IPs there.

## 7. Firewall

On servers the agent owns firewalld's `mesh` zone on wg0 (target DROP) and what it opened in the
public zone: udp/51820, the Swarm ports from the other servers, DNS from the mesh, and the custom
rules of the web UI's Firewall page -- a port from a source network, on one server or on every
server. That page shows each server's rules with where each comes from. Rules you make yourself
elsewhere are never touched; to change what the mesh allows, use the page, not firewall-cmd.
Spokes keep their own firewall: on the NAS, allow your apps' ports from the mesh CIDR in its
firewall settings.

Every 5 minutes the agent applies its config again after asking firewalld what it holds: wg0
moved to another zone at runtime (NetworkManager does this), a rule removed by hand, a route or
wg0 itself deleted are put back and logged as `drift check: put back what changed`.

## 8. The web UI

The UI is served at the server's URL (`https://mesh.example.com/`), beside the agent API. Create
the first user on the server host, then sign in; at the first sign-in the UI shows a QR code for
an authenticator app (any TOTP app) and asks for one code to confirm it.

```
docker exec -it <container> /app/meshctl user add admin
```

| Page | What it does |
| --- | --- |
| Dashboard | every node with its state (green in sync, amber applying / degraded / late, red silent or rolled back, grey disabled or not joined; a phone green while its tunnel is on, amber when its profile is out of date, grey idle), applied and current version, agent, last heartbeat (a phone's last handshake), handshake age per peer and exit; refreshes every 10 s |
| Nodes | add a node (the join command is shown once), open one to see its computed config, peers and DNS state, edit it, disable / enable, issue a new token, make a phone's profile (a QR code, shown once), remove it |
| DNS | the mesh zone: one name per node and the custom records (add, remove); the upstream resolvers |
| Firewall | each server's rules with where each comes from; add and remove custom rules |
| Versions | every config version with what it changed; roll back to one (a new version; a rollback that adds or removes nodes asks for confirmation) |
| Audit | who changed what, newest first |
| Settings | default MTU, default hub, upstream DNS; the mesh CIDR only while the mesh is empty; download the export |
| Users | add users, disable them, reset their authenticator |
| Account | change your password (signs you out everywhere) |

Changes made in the UI go through the same path as `meshctl node`: versioned, audited under
your user, pushed to the agents at once. The dashboard's 10 s refresh does not keep a session
alive; 12 hours without using the UI signs you out, and 7 days always do.

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

`agent run --ui-listen 127.0.0.1:7444` moves it (loopback addresses only) and `--ui-listen off`
turns it off. An exit, keepalive or MTU change asks the server, exactly like `meshctl exit`;
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

`passwd` and `reset-totp` end the user's sessions. Five failed sign-ins from one address lock it
out for 15 minutes; twenty for one username within an hour lock the username.

## 10. Backups

The server writes a nightly export at 03:00 (container time) to `/data/exports/mesh-YYYY-MM-DD.yaml`,
skipping nights when nothing changed and keeping the newest 14. For an export on demand:

```
docker exec <container> /app/meshctl export > mesh.yaml
```

An export is the desired state as YAML -- nodes with their public keys, settings, custom DNS
records and firewall rules; never tokens, private keys or users. It reads the database directly
and works while the server runs. Commit it to git: an unchanged mesh exports byte-for-byte the
same file.

To restore an export (a new server, or undoing a bad day), stop the server and import it; node
ids are kept, so the agents keep working without re-joining:

```
docker compose stop meshctl
docker compose run --rm meshctl import /data/exports/mesh-2026-10-12.yaml
docker compose start meshctl
```

`import` refuses to run while the server holds the data dir, and checks every value of the file
before it changes anything. Users are not in an export: for a full restore, including users and
`secret.key` (which seals their TOTP secrets), keep a copy of the `meshctl-data` volume as well
(`docker run --rm -v meshctl-data:/data -v $PWD:/b alpine tar czf /b/meshctl-data.tgz -C /data .`,
with the server stopped).

## 11. When the server is down

Agents keep their last config (`/var/lib/meshctl/config.json`) and re-apply it at boot without
the server, so the mesh survives a server outage and reboots, and every node keeps answering
`.mesh` names. Each node also keeps the last view of the mesh it fetched (`mesh.json`, at most
5 minutes old when the server went down), which it shows marked with the server's absence. Changes
simply wait. When the server is back, agents reconnect within a minute.

## 12. Troubleshooting

`journalctl -u meshctl-agent` or `meshctl logs` on a node, `docker logs <container>` on the server.

| Symptom | Cause and fix |
| --- | --- |
| `clock_skew` in the agent log | the node's clock is off by more than 60 s; enable NTP (`timedatectl set-ntp true`) |
| `node_disabled` | an operator disabled the node; `node enable` it |
| `token_invalid ... already used` / `expired` | tokens are single-use and live an hour: `node token <name>` |
| `interface wg0 exists but was not created by meshctl` | a wg-quick setup owns wg0: `wg-quick down wg0` and disable it, then restart the agent |
| `firewall` listed as degraded | firewalld is missing or stopped; the tunnel works, but Swarm and udp/51820 may be blocked. Install/start firewalld, or open the ports yourself and run the agent with `--no-firewall` |
| `create WireGuard interface wg0: ... is the wireguard kernel module available?` | `modprobe wireguard`, or run WireGuard in the agent: `agent run --wg-backend userspace` (`auto`, the default, picks it when the module is missing) |
| `userspace WireGuard needs /dev/net/tun` | the container has no tun device: use the compose file's userspace variant, which maps it, or the kernel variant if the NAS has the module |
| `wg show` does not list a userspace wg0 | the agent could not create `/var/run/wireguard/wg0.sock` (its log says why; a container has no `wg` anyway): use `meshctl status` |
| `only spokes run in a container` | `node compose` names a server; servers install with the command `node add` prints |
| `... only releases have an image; this is a development build` | `node compose` on a server built from source: run a release image, whose version is a tag the registry has |
| `another node already uses this WireGuard or identity key` | a data dir was copied between hosts; remove `/var/lib/meshctl` on the copy and join it with its own token |
| a node stays `applying` | it applied an older version and is failing the new one; `meshctl status` on it shows the error |
| `exit "vps-3" is not exit-capable` | allow it first: `node set vps-3 --can-exit` |
| `only spokes choose an exit or a keepalive` | `meshctl exit` ran on a server; exits are for spokes |
| `apply step exit: ... operation not permitted` | the agent lacks root or CAP_NET_ADMIN; run it as the installed service |
| a server shows `rolled back` | safe apply: no heartbeat was answered within 60 s of a change, so the previous config is back. See what the change did (endpoint, MTU, firewall), fix it, and make the next change |
| `the server accepted the change, but the agent has not applied it` | `meshctl status` on the node shows the error the apply hit |
| `the data dir is in use: stop the meshctl server first` | `import` ran beside a running server; stop it, import, start it |
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
| `installing the Windows service needs an elevated prompt` | `agent install`, `agent join` and `agent uninstall` change services: run them elevated |
| `this meshctl was built without wintun.dll` | a binary from `go build` without `-tags wintun`: use the release binary from `/dl/`, or `./scripts/dev.ps1 build` |
| `the wintun.dll in this build is not the pinned 0.14.1` | the binary was built from a DLL that is not the release: rebuild with `./scripts/dev.ps1 build`, which fetches and checks it |
| the Windows service stops and starts again every few seconds | read `C:\ProgramData\meshctl\agent.log`; the agent exits with what failed and the service manager restarts it |
| `drift check: put back what changed` every few minutes | something keeps changing the host's mesh setup, usually NetworkManager managing wg0: `nmcli device set wg0 managed no` |

## 13. Upgrading

Server: bump `MESHCTL_VERSION` and redeploy. At start the server recomputes every node's config;
if the new release computes differently, the agents receive the change like any other.

Agents: run the install command again with a fresh token (`node token <name>`), or replace
`/usr/local/bin/meshctl` with the new release and `systemctl restart meshctl-agent`.
