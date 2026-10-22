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
| `--name` |  | the node name the token must belong to (a safety check) |
| `--server` |  | server origin, e.g. https://mesh.example.com |
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
| `--name` |  | the node name the token must belong to (a safety check) |
| `--no-service` | `false` | do not install and start the system service |
| `--server` |  | server origin, e.g. https://mesh.example.com |
| `--token` |  | one-time join token (or set MESHCTL_TOKEN, which keeps it out of the process list) |

## meshctl agent run

Run the agent in the foreground (what the service runs).

```
meshctl agent run [flags]
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `--data` | `/var/lib/meshctl` | agent data directory |
| `--log-level` | `info` | debug, info, warn or error |
| `--no-firewall` | `false` | leave the host firewall alone (you open udp/51820 and the Swarm ports yourself) |
| `--runtime-socket` | `auto` | the container runtime's API socket the containers' names are read from (GET only): auto (Docker's, then Podman's), off, or a path |
| `--socket` |  | local API socket (default: /run/meshctl/agent.sock for the standard data dir, else &lt;data&gt;/agent.sock; \\.\pipe\meshctl-agent on Windows) |
| `--status-every` | `30s` | how often the agent reports to the server; the server's --hub-down-after must be longer on a server |
| `--ui-listen` | `127.0.0.1:7443` | the local UI's loopback address, or off; meshctl ui prints a sign-in link |
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

## meshctl dns

Show or switch this node's split DNS for the mesh zone's names (stays on the node).

```
meshctl dns [on|off] [flags]
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `--data` | `/var/lib/meshctl` | agent data directory |
| `--socket` |  | local API socket (default: /run/meshctl/agent.sock for the standard data dir, else &lt;data&gt;/agent.sock; \\.\pipe\meshctl-agent on Windows) |

## meshctl exit

Show or switch this spoke's internet exit (through the local agent).

```
meshctl exit [<server>|off] [flags]
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `--data` | `/var/lib/meshctl` | agent data directory |
| `--socket` |  | local API socket (default: /run/meshctl/agent.sock for the standard data dir, else &lt;data&gt;/agent.sock; \\.\pipe\meshctl-agent on Windows) |
| `--wait` | `30s` | how long to wait for the agent to apply the change (0 returns once the server accepted it) |

## meshctl export

Write the mesh's desired state as YAML (reads the database directly).

```
meshctl export [flags]
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `--data` | `/data` | the server's data directory |
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

## meshctl firewall check

Say what the mesh firewall does with one flow, now or --at a time (from and to: nodes or IPv4 addresses; service: tcp/PORT, udp/PORT or icmp).

```
meshctl firewall check <from> <to> <service> [flags]
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `--at` |  | judge the flow at this time, in RFC 3339 form such as 2026-10-23T09:30:00+08:00 (default now) |
| `--data` | `/data` | the server's data directory (holds admin.sock) |
| `--json` | `false` | print the server's answer as JSON |

## meshctl firewall show

Print the mesh firewall: policies, zones and the rules every server enforces.

```
meshctl firewall show [flags]
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `--data` | `/data` | the server's data directory (holds admin.sock) |
| `--json` | `false` | print the server's answer as JSON |
| `--yaml` | `false` | print the document as YAML, to edit and give to `firewall apply` |

## meshctl healthcheck agent

Exit 0 when the local agent answers on its socket.

```
meshctl healthcheck agent [flags]
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `--data` | `/var/lib/meshctl` | agent data directory |
| `--socket` |  | local API socket (default: /run/meshctl/agent.sock for the standard data dir, else &lt;data&gt;/agent.sock; \\.\pipe\meshctl-agent on Windows) |

## meshctl healthcheck server

Exit 0 when the server answers /healthz.

```
meshctl healthcheck server [flags]
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `--addr` | `127.0.0.1:8080` | host:port the server listens on |

## meshctl import

Replace the desired state with an export (only while the server is stopped).

```
meshctl import <file> [flags]
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `--data` | `/data` | the server's data directory (the server must be stopped) |

## meshctl logs

Print this node's recent agent log; -f keeps following it.

```
meshctl logs [flags]
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `--data` | `/var/lib/meshctl` | agent data directory |
| `-f` | `false` | keep printing new lines until interrupted |
| `--socket` |  | local API socket (default: /run/meshctl/agent.sock for the standard data dir, else &lt;data&gt;/agent.sock; \\.\pipe\meshctl-agent on Windows) |

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
| `--data` | `/data` | the server's data directory (holds admin.sock) |
| `--endpoint` |  | servers: public IPv4:port (default: what the agent detects at join) |
| `--groups` |  | comma-separated firewall groups to put the node in (a new name makes a new group) |
| `--hub` |  | spokes and phones: the server that relays for it (default: the default hub) |
| `--ip` |  | fixed mesh IP (default: the lowest free one in the role's block) |
| `--json` | `false` | print the server's answer as JSON |
| `--keepalive` | `0` | spokes and phones: PersistentKeepalive seconds (default 25) |
| `--listen-port` | `0` | servers: WireGuard UDP port (default 51820) |
| `--mode` | `host` | host (WireGuard on the host) or container (in the agent's container: only its containers are on the mesh) |
| `--mtu` | `0` | interface MTU (default: the mesh default, 1420) |
| `--role` |  | server, spoke or phone (required) |

## meshctl node compose

Print the compose file that runs a node's agent in a container (container mode, or a NAS).

```
meshctl node compose <name> [flags]
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `--data` | `/data` | the server's data directory (holds admin.sock) |
| `--host-dir` |  | host directory for the agent's keys and cached config, mounted at /data (default: /volume1/docker/meshctl for a NAS, a named volume in container mode) |
| `--json` | `false` | print the server's answer as JSON |
| `--no-names` | `false` | container mode: leave out the runtime's socket, so the node's containers get no names on the mesh |
| `--runtime` |  | container mode: docker (default), podman or swarm (Swarm services attach) |
| `--wg-backend` |  | WireGuard in the container: auto (container mode's default: the host's module, else /dev/net/tun), userspace (the NAS default; needs /dev/net/tun) or kernel (the host has the wireguard module) |

## meshctl node disable

Take a node out of every peer list (its keys stay registered).

```
meshctl node disable <name> [flags]
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `--data` | `/data` | the server's data directory (holds admin.sock) |
| `--json` | `false` | print the server's answer as JSON |

## meshctl node enable

Put a disabled node back.

```
meshctl node enable <name> [flags]
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `--data` | `/data` | the server's data directory (holds admin.sock) |
| `--json` | `false` | print the server's answer as JSON |

## meshctl node list

List the nodes and the mesh plan's warnings.

```
meshctl node list [flags]
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `--data` | `/data` | the server's data directory (holds admin.sock) |
| `--json` | `false` | print the server's answer as JSON |

## meshctl node network

Print the commands that put a node's containers on the mesh (its mesh network, then a service).

```
meshctl node network <name> [flags]
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `--data` | `/data` | the server's data directory (holds admin.sock) |
| `--json` | `false` | print the server's answer as JSON |
| `--runtime` |  | docker (default), podman or swarm (Swarm services attach) |

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

## meshctl node rm

Remove a node; every peer drops it on the next apply.

```
meshctl node rm <name> [flags]
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `--data` | `/data` | the server's data directory (holds admin.sock) |
| `--json` | `false` | print the server's answer as JSON |

## meshctl node set

Change a node's endpoint, hub, exit, keepalive, MTU, container network, mesh IP or groups.

```
meshctl node set <name> [flags]
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `--can-exit` | `false` | servers: may be chosen as an internet exit (--can-exit=false to revoke) |
| `--containers` | `false` | host mode: put the node's containers on the mesh (--containers=false takes them off) |
| `--data` | `/data` | the server's data directory (holds admin.sock) |
| `--endpoint` |  | servers: public IPv4:port, or none to clear it |
| `--exit` |  | spokes: exit server, or off |
| `--groups` |  | the node's firewall groups, comma separated, replacing the ones it is in ("" takes it out of all) |
| `--hub` |  | spokes and phones: relay server, or default for the default hub |
| `--json` | `false` | print the server's answer as JSON |
| `--keepalive` | `0` | PersistentKeepalive seconds (0 turns it off) |
| `--listen-port` | `0` | servers: WireGuard UDP port |
| `--mesh-ip` |  | move the node to another mesh IP in the mesh (its peers, DNS name and agent follow) |
| `--mtu` | `0` | interface MTU |

## meshctl node show

Show a node, its computed config and its peers.

```
meshctl node show <name> [flags]
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `--data` | `/data` | the server's data directory (holds admin.sock) |
| `--json` | `false` | print the server's answer as JSON |

## meshctl node token

Issue a new join token (revokes unused ones; re-keys an enrolled node).

```
meshctl node token <name> [flags]
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `--data` | `/data` | the server's data directory (holds admin.sock) |
| `--json` | `false` | print the server's answer as JSON |

## meshctl server

Run the control-plane server.

```
meshctl server [flags]
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

## meshctl status

Show this node's agent: applied config and peers.

```
meshctl status [flags]
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `--data` | `/var/lib/meshctl` | agent data directory |
| `--json` | `false` | print the status as JSON |
| `--socket` |  | local API socket (default: /run/meshctl/agent.sock for the standard data dir, else &lt;data&gt;/agent.sock; \\.\pipe\meshctl-agent on Windows) |

## meshctl ui

Print a one-time sign-in link to this node's local UI.

```
meshctl ui [flags]
```

| Flag | Default | Meaning |
| --- | --- | --- |
| `--data` | `/var/lib/meshctl` | agent data directory |
| `--socket` |  | local API socket (default: /run/meshctl/agent.sock for the standard data dir, else &lt;data&gt;/agent.sock; \\.\pipe\meshctl-agent on Windows) |

## meshctl user add

Create a web UI user; they enroll an authenticator at their first sign-in.

```
meshctl user add <name> [flags]
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

## meshctl version

Print the meshctl version.

```
meshctl version [flags]
```
