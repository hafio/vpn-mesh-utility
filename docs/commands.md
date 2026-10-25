# Command reference

Every `meshctl` command and flag. This file is generated from the command registry in
`cmd/meshctl` by `TestCommandsDoc`; after changing a command, run
`./scripts/dev.sh regen` and commit the result. Flags may come before or after the
positional arguments, and `-flag` and `--flag` are the same. Defaults are the Linux ones.

## meshctl agent install

Windows: copy meshctl into Program Files, join this node and install the service.

```
meshctl agent install [flags]
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `--data` | `/var/lib/meshctl` | agent data directory |
| `--endpoint` |  | servers only: public IPv4:port, when auto-detection would be wrong |
| `--endpoint6` |  | servers only: public [IPv6]:port, when auto-detection would be wrong |
| `--name` |  | the node name the token must belong to (a safety check) |
| `--server` |  | the control plane's origin, e.g. https://mesh.example.com |
| `--token` |  | one-time join token (or set MESHCTL_TOKEN, which keeps it out of the process list) |

## meshctl agent join

Enroll this node with a one-time token, then install and start the service.

```
meshctl agent join [flags]
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `--data` | `/var/lib/meshctl` | agent data directory |
| `--endpoint` |  | servers only: public IPv4:port, when auto-detection would be wrong |
| `--endpoint6` |  | servers only: public [IPv6]:port, when auto-detection would be wrong |
| `--name` |  | the node name the token must belong to (a safety check) |
| `--no-service` | `false` | do not install and start the system service |
| `--server` |  | the control plane's origin, e.g. https://mesh.example.com |
| `--token` |  | one-time join token (or set MESHCTL_TOKEN, which keeps it out of the process list) |

## meshctl agent run

Run the agent in the foreground (what the service runs).

```
meshctl agent run [flags]
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `--data` | `/var/lib/meshctl` | agent data directory |
| `--endpoint-retry` | `45s` | how long a spoke waits for a first handshake on a server's endpoint before it dials the server's other-family one |
| `--log-level` | `info` | debug, info, warn or error |
| `--no-firewall` | `false` | leave the host firewall alone (you open the WireGuard port and the Swarm ports yourself) |
| `--runtime-socket` | `auto` | the container runtime's API socket the containers' names are read from (GET only): auto (Docker's, then Podman's), off, or a path |
| `--socket` |  | local API socket (default: /run/meshctl/agent.sock for the standard data dir, else &lt;data&gt;/agent.sock; on Windows \\.\pipe\meshctl-agent, with -&lt;slot&gt; for a mesh in another slot) |
| `--status-every` | `30s` | how often the agent reports to the server; the server's --hub-down-after must be longer on a server |
| `--ui-listen` | `auto` | the local UI's loopback address, auto (127.0.0.1:7443, plus the mesh's slot) or off; meshctl ui prints a sign-in link |
| `--wg-backend` | `auto` | WireGuard backend: auto (kernel when the module is there), kernel or userspace (wireguard-go on /dev/net/tun) |

## meshctl agent uninstall

Windows: stop and remove the service (--purge also removes the data dir).

```
meshctl agent uninstall [flags]
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `--data` | `/var/lib/meshctl` | agent data directory |
| `--purge` | `false` | also remove the data directory: this node's keys and cached config (it must join again) |

## meshctl control

Run the control plane: the database, the web UI and the agent API (no WireGuard of its own).

```
meshctl control [flags]
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `--data` | `/data` | data directory: the database and the admin socket |
| `--dist` |  | directory with the binaries and SHA256SUMS served at /dl/ (default $MESHCTL_DIST_DIR) |
| `--hub-down-after` | `2m0s` | relay failover: a hub silent this long, with no handshake for half as long again, has its pairs relayed by another hub; 0 turns it off |
| `--listen` | `:8080` | address for the agent API, health check and downloads |
| `--log-level` | `info` | debug, info, warn or error |
| `--trusted-proxies` |  | comma-separated CIDRs of reverse proxies whose X-Forwarded-For is believed (default $MESHCTL_TRUSTED_PROXIES) |
| `--url` |  | public origin agents and install.sh use, e.g. https://mesh.example.com (default $MESHCTL_URL) |

## meshctl dns

Show or switch this node's split DNS for the mesh zone's names (stays on the node).

```
meshctl dns [on|off] [flags]
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `--data` | `/var/lib/meshctl` | agent data directory |
| `--socket` |  | local API socket (default: /run/meshctl/agent.sock for the standard data dir, else &lt;data&gt;/agent.sock; on Windows \\.\pipe\meshctl-agent, with -&lt;slot&gt; for a mesh in another slot) |

## meshctl exit

Show or switch this spoke's internet exit (through the local agent).

```
meshctl exit [<server>|off] [flags]
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `--data` | `/var/lib/meshctl` | agent data directory |
| `--socket` |  | local API socket (default: /run/meshctl/agent.sock for the standard data dir, else &lt;data&gt;/agent.sock; on Windows \\.\pipe\meshctl-agent, with -&lt;slot&gt; for a mesh in another slot) |
| `--wait` | `30s` | how long to wait for the agent to apply the change (0 returns once the server accepted it) |

## meshctl export

Write a mesh's desired state as YAML (reads the database directly).

```
meshctl export [flags]
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `--data` | `/data` | the server's data directory |
| `--mesh` |  | the mesh to export; needed when the control plane has more than one |
| `--out` |  | write to this file (mode 0600) instead of stdout |

## meshctl firewall apply

Replace the mesh firewall with a YAML file, as `firewall show --yaml` writes it.

```
meshctl firewall apply <file> [flags]
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `--data` | `/data` | the server's data directory (holds admin.sock) |
| `--json` | `false` | print the server's answer as JSON |
| `--mesh` |  | the mesh to work on; needed when the control plane has more than one (meshctl mesh list) |

## meshctl firewall check

Say what the mesh firewall does with one flow, now or --at a time (from and to: nodes or addresses, IPv4 or IPv6; service: tcp/PORT, udp/PORT or icmp).

```
meshctl firewall check <from> <to> <service> [flags]
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `--at` |  | judge the flow at this time, in RFC 3339 form such as 2026-10-23T09:30:00+08:00 (default now) |
| `--data` | `/data` | the server's data directory (holds admin.sock) |
| `--json` | `false` | print the server's answer as JSON |
| `--mesh` |  | the mesh to work on; needed when the control plane has more than one (meshctl mesh list) |

## meshctl firewall show

Print the mesh firewall: policies, zones and the rules every server enforces.

```
meshctl firewall show [flags]
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `--data` | `/data` | the server's data directory (holds admin.sock) |
| `--json` | `false` | print the server's answer as JSON |
| `--mesh` |  | the mesh to work on; needed when the control plane has more than one (meshctl mesh list) |
| `--yaml` | `false` | print the document as YAML, to edit and give to `firewall apply` |

## meshctl healthcheck agent

Exit 0 when the local agent answers on its socket.

```
meshctl healthcheck agent [flags]
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `--data` | `/var/lib/meshctl` | agent data directory |
| `--socket` |  | local API socket (default: /run/meshctl/agent.sock for the standard data dir, else &lt;data&gt;/agent.sock; on Windows \\.\pipe\meshctl-agent, with -&lt;slot&gt; for a mesh in another slot) |

## meshctl healthcheck control

Exit 0 when the control plane answers /healthz.

```
meshctl healthcheck control [flags]
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `--addr` | `127.0.0.1:8080` | host:port the control plane listens on |

## meshctl healthcheck start

Exit 0 when what `meshctl start` runs answers: the image's health check, from MESHCTL_MODE and MESHCTL_PORT or MESHCTL_DATA.

```
meshctl healthcheck start [flags]
```

## meshctl import

Replace a mesh's desired state with an export (only while the control plane is stopped).

```
meshctl import <file> [flags]
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `--data` | `/data` | the server's data directory (the server must be stopped) |
| `--mesh` |  | the mesh to replace, created when it does not exist; needed when the control plane has more than one |

## meshctl ipv6

Show or set the mesh's IPv6 prefix on the running control plane: a /48 to /64, auto (a random unique local /64) or off.

```
meshctl ipv6 [<prefix>|auto|off] [flags]
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `--data` | `/data` | the server's data directory (holds admin.sock) |
| `--json` | `false` | print the server's answer as JSON |
| `--mesh` |  | the mesh to work on; needed when the control plane has more than one (meshctl mesh list) |

## meshctl logs

Print this node's recent agent log; -f keeps following it.

```
meshctl logs [flags]
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `--data` | `/var/lib/meshctl` | agent data directory |
| `-f` | `false` | keep printing new lines until interrupted |
| `--socket` |  | local API socket (default: /run/meshctl/agent.sock for the standard data dir, else &lt;data&gt;/agent.sock; on Windows \\.\pipe\meshctl-agent, with -&lt;slot&gt; for a mesh in another slot) |

## meshctl mesh add

Add a mesh with its own network, DNS zone and IPv6 prefix.

```
meshctl mesh add <name> [flags]
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `--data` | `/data` | the server's data directory (holds admin.sock) |
| `--dns-zone` |  | the mesh's DNS zone (default: the mesh's name) |
| `--ipv6` | `auto` | the mesh's IPv6 prefix: a /48 to /64, auto (a random unique local /64) or off |
| `--json` | `false` | print the server's answer as JSON |
| `--mesh-cidr` |  | the mesh's IPv4 network, a /16 to /22 (default: the first free 10.x.0.0/16) |

## meshctl mesh list

List the control plane's meshes.

```
meshctl mesh list [flags]
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `--data` | `/data` | the server's data directory (holds admin.sock) |
| `--json` | `false` | print the server's answer as JSON |

## meshctl mesh rm

Remove a mesh without nodes, with its settings, history and policies.

```
meshctl mesh rm <name> [flags]
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `--data` | `/data` | the server's data directory (holds admin.sock) |
| `--json` | `false` | print the server's answer as JSON |

## meshctl mesh set

Rename a mesh or change its settings.

```
meshctl mesh set <name> [flags]
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `--data` | `/data` | the server's data directory (holds admin.sock) |
| `--default-hub` |  | the server spokes and phones without a hub relay through ("" for none) |
| `--default-mtu` | `0` | the MTU of nodes that set none |
| `--dns-extra-domains` |  | extra DNS domains every node resolves through the mesh, comma separated ("" for none) |
| `--dns-zone` |  | the mesh's DNS zone |
| `--json` | `false` | print the server's answer as JSON |
| `--mesh-cidr` |  | the mesh's IPv4 network, a /16 to /22; only while the mesh has no nodes |
| `--name` |  | a new name for the mesh (its uuid and slot stay) |
| `--upstream-dns` |  | the resolvers the mesh's resolvers forward to, one to four IPv4 or IPv6 addresses, comma separated |

## meshctl mesh show

Show a mesh and its settings.

```
meshctl mesh show <name> [flags]
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `--data` | `/data` | the server's data directory (holds admin.sock) |
| `--json` | `false` | print the server's answer as JSON |

## meshctl node add

Add a node and print its one-time join command.

```
meshctl node add <name> [flags]
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `--always-on` | `false` | spokes: allocate from the always-on block (NAS) |
| `--can-exit` | `false` | servers: may be chosen as an internet exit |
| `--containers` | `false` | host mode: put the node's containers on the mesh too (container mode always does) |
| `--containers6` | `false` | put the containers' IPv6 on the mesh too (needs --containers or container mode) |
| `--data` | `/data` | the server's data directory (holds admin.sock) |
| `--endpoint` |  | servers: public IPv4:port (default: what the agent detects at join) |
| `--endpoint6` |  | servers: public [IPv6]:port (spokes dial it when IPv4 does not work) |
| `--family` |  | dual (default: IPv4 and, once the mesh has an IPv6 prefix, IPv6), ipv4 or ipv6: the node's mesh addresses and the endpoints it dials |
| `--groups` |  | comma-separated firewall groups to put the node in (a new name makes a new group) |
| `--hub` |  | spokes and phones: the server that relays for it (default: the default hub) |
| `--ip` |  | fixed mesh IP (default: the lowest free one in the role's block) |
| `--ip6` |  | fixed IPv6 mesh address (default: the lowest free one in the role's block, when the mesh has an IPv6 prefix) |
| `--json` | `false` | print the server's answer as JSON |
| `--keepalive` | `0` | spokes and phones: PersistentKeepalive seconds (default 25) |
| `--listen-port` | `0` | servers: WireGuard UDP port (default 51820, plus the mesh's slot) |
| `--mesh` |  | the mesh to work on; needed when the control plane has more than one (meshctl mesh list) |
| `--mode` | `host` | host (WireGuard on the host) or container (in the agent's container: only its containers are on the mesh) |
| `--mtu` | `0` | interface MTU (default: the mesh default, 1420) |
| `--role` |  | server, spoke or phone (required) |

## meshctl node disable

Take a node out of every peer list (its keys stay registered).

```
meshctl node disable <name> [flags]
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `--data` | `/data` | the server's data directory (holds admin.sock) |
| `--json` | `false` | print the server's answer as JSON |
| `--mesh` |  | the mesh to work on; needed when the control plane has more than one (meshctl mesh list) |

## meshctl node enable

Put a disabled node back.

```
meshctl node enable <name> [flags]
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `--data` | `/data` | the server's data directory (holds admin.sock) |
| `--json` | `false` | print the server's answer as JSON |
| `--mesh` |  | the mesh to work on; needed when the control plane has more than one (meshctl mesh list) |

## meshctl node list

List the nodes and the mesh plan's warnings.

```
meshctl node list [flags]
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `--data` | `/data` | the server's data directory (holds admin.sock) |
| `--json` | `false` | print the server's answer as JSON |
| `--mesh` |  | the mesh to work on; needed when the control plane has more than one (meshctl mesh list) |

## meshctl node phone

Print a new WireGuard profile for a phone (a new key: its previous profile stops working).

```
meshctl node phone <name> [flags]
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `--data` | `/data` | the server's data directory (holds admin.sock) |
| `--exit` |  | send all of the phone's traffic through this exit server (default: only mesh traffic uses the tunnel) |
| `--json` | `false` | print the server's answer as JSON |
| `--mesh` |  | the mesh to work on; needed when the control plane has more than one (meshctl mesh list) |

## meshctl node rm

Remove a node; every peer drops it on the next apply.

```
meshctl node rm <name> [flags]
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `--data` | `/data` | the server's data directory (holds admin.sock) |
| `--json` | `false` | print the server's answer as JSON |
| `--mesh` |  | the mesh to work on; needed when the control plane has more than one (meshctl mesh list) |

## meshctl node set

Change a node's endpoint, hub, exit, keepalive, MTU, container network, mesh IP or groups.

```
meshctl node set <name> [flags]
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `--can-exit` | `false` | servers: may be chosen as an internet exit (--can-exit=false to revoke) |
| `--containers` | `false` | host mode: put the node's containers on the mesh (--containers=false takes them off) |
| `--containers6` | `false` | put the containers' IPv6 on the mesh (--containers6=false takes it off) |
| `--data` | `/data` | the server's data directory (holds admin.sock) |
| `--endpoint` |  | servers: public IPv4:port, or none to clear it |
| `--endpoint6` |  | servers: public [IPv6]:port, or none to clear it |
| `--exit` |  | spokes: exit server, or off |
| `--family` |  | dual, ipv4 or ipv6: allocates the address a family adds, drops the one it takes away |
| `--groups` |  | the node's firewall groups, comma separated, replacing the ones it is in ("" takes it out of all) |
| `--hub` |  | spokes and phones: relay server, or default for the default hub |
| `--json` | `false` | print the server's answer as JSON |
| `--keepalive` | `0` | PersistentKeepalive seconds (0 turns it off) |
| `--listen-port` | `0` | servers: WireGuard UDP port |
| `--mesh` |  | the mesh to work on; needed when the control plane has more than one (meshctl mesh list) |
| `--mesh-ip` |  | move the node to another mesh IP in the mesh (its peers, DNS name and agent follow) |
| `--mesh-ip6` |  | move the node to another IPv6 mesh address in the mesh's IPv6 prefix |
| `--mtu` | `0` | interface MTU |

## meshctl node show

Show a node, its container network's values, its computed config and its peers.

```
meshctl node show <name> [flags]
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `--data` | `/data` | the server's data directory (holds admin.sock) |
| `--json` | `false` | print the server's answer as JSON |
| `--mesh` |  | the mesh to work on; needed when the control plane has more than one (meshctl mesh list) |

## meshctl node token

Issue a new join token (revokes unused ones; cuts an enrolled node off until it joins with it).

```
meshctl node token <name> [flags]
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `--data` | `/data` | the server's data directory (holds admin.sock) |
| `--json` | `false` | print the server's answer as JSON |
| `--mesh` |  | the mesh to work on; needed when the control plane has more than one (meshctl mesh list) |

## meshctl start

The image's command: run the control plane or a node as MESHCTL_MODE says (a node when unset), a node joining first with a join token it has not used yet.

```
meshctl start [flags]
```

| Variable | Mode | Default | Meaning |
| --- | --- | --- | --- |
| `MESHCTL_MODE` | both | `node` | what the container runs: control, the control plane (database, web UI, agent API), or node, an agent -- a server or a client as its role on the control plane says |
| `MESHCTL_URL` | both |  | the control plane's public origin, e.g. https://mesh.example.com: the control plane's own (required), the one a node joins (needed with a join token) |
| `MESHCTL_DATA` | both | `/data` | the data directory, where the volume is mounted |
| `MESHCTL_LOG_LEVEL` | both | `info` | `meshctl control --log-level`: debug, info, warn or error |
| `MESHCTL_PORT` | control | `8080` | the port the control plane listens on, on every address |
| `MESHCTL_TRUSTED_PROXIES` | control |  | `meshctl control --trusted-proxies`: comma-separated CIDRs of reverse proxies whose X-Forwarded-For is believed (default $MESHCTL_TRUSTED_PROXIES) |
| `MESHCTL_HUB_DOWN_AFTER` | control | `2m0s` | `meshctl control --hub-down-after`: relay failover: a hub silent this long, with no handshake for half as long again, has its pairs relayed by another hub; 0 turns it off |
| `MESHCTL_NAME` | node |  | the node's name: a token issued for another node is refused |
| `MESHCTL_TOKEN` | node |  | the node's join token: start joins with a token it has not used before, then runs; with the same token or none, a node that has joined just runs |
| `MESHCTL_TOKEN_FILE` | node |  | an absolute path to a file holding the join token, in place of MESHCTL_TOKEN (a compose secret) |
| `MESHCTL_ENDPOINT` | node |  | server nodes only: the public IPv4:port to join with, when auto-detection would be wrong |
| `MESHCTL_ENDPOINT6` | node |  | server nodes only: the public [IPv6]:port to join with, when auto-detection would be wrong |
| `MESHCTL_WG_BACKEND` | node | `auto` | `meshctl agent run --wg-backend`: WireGuard backend: auto (kernel when the module is there), kernel or userspace (wireguard-go on /dev/net/tun) |
| `MESHCTL_RUNTIME_SOCKET` | node | `auto` | `meshctl agent run --runtime-socket`: the container runtime's API socket the containers' names are read from (GET only): auto (Docker's, then Podman's), off, or a path |
| `MESHCTL_UI_LISTEN` | node | `auto` | `meshctl agent run --ui-listen`: the local UI's loopback address, auto (127.0.0.1:7443, plus the mesh's slot) or off; meshctl ui prints a sign-in link |

## meshctl status

Show this node's agent: applied config and peers.

```
meshctl status [flags]
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `--data` | `/var/lib/meshctl` | agent data directory |
| `--json` | `false` | print the status as JSON |
| `--socket` |  | local API socket (default: /run/meshctl/agent.sock for the standard data dir, else &lt;data&gt;/agent.sock; on Windows \\.\pipe\meshctl-agent, with -&lt;slot&gt; for a mesh in another slot) |

## meshctl ui

Print a one-time sign-in link to this node's local UI.

```
meshctl ui [flags]
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `--data` | `/var/lib/meshctl` | agent data directory |
| `--socket` |  | local API socket (default: /run/meshctl/agent.sock for the standard data dir, else &lt;data&gt;/agent.sock; on Windows \\.\pipe\meshctl-agent, with -&lt;slot&gt; for a mesh in another slot) |

## meshctl user add

Create a web UI user; they enroll an authenticator at their first sign-in.

```
meshctl user add <name> [flags]
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `--admin` | `false` | an admin manages every mesh, the meshes themselves and the users (the first user always is one) |
| `--data` | `/data` | the server's data directory |
| `--mesh` |  | the mesh the user manages, alone (not for the first user, an admin; meshctl user grant adds more) |

## meshctl user grant

Let a user manage a mesh (--mesh), or make them an admin (--admin).

```
meshctl user grant <name> [flags]
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `--admin` | `false` | the admin role instead of a mesh |
| `--data` | `/data` | the server's data directory |
| `--mesh` |  | the mesh to grant or take away |

## meshctl user list

List the web UI users, the admins and the meshes granted to each.

```
meshctl user list [flags]
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `--data` | `/data` | the server's data directory |

## meshctl user passwd

Set a user's password and end their sessions.

```
meshctl user passwd <name> [flags]
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `--data` | `/data` | the server's data directory |

## meshctl user reset-totp

Remove a user's authenticator (they enroll a new one at their next sign-in) and end their sessions.

```
meshctl user reset-totp <name> [flags]
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `--data` | `/data` | the server's data directory |

## meshctl user revoke

Take a mesh away from a user (--mesh), or their admin role (--admin).

```
meshctl user revoke <name> [flags]
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `--admin` | `false` | the admin role instead of a mesh |
| `--data` | `/data` | the server's data directory |
| `--mesh` |  | the mesh to grant or take away |

## meshctl version

Print the meshctl version.

```
meshctl version [flags]
```
