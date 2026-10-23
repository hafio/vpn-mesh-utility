# Agent protocol

The wire contract between `meshctl server` and its agents, and the admin API behind
`meshctl node`. The Go source of truth is `internal/proto` (shapes, signing, SSE framing) and
`internal/server/api` (routes). `TestRoutesAreDocumented` fails when a route served by the API is
missing here, so this list is complete.

## Routes

Public listener (`--listen`, behind the reverse proxy):

| Route | Auth | Purpose |
| --- | --- | --- |
| `POST /api/v1/agent/join` | one-time token | enroll a node: register its public keys, get its id and mesh IP |
| `GET /api/v1/agent/config` | signature | the node's full desired config; `If-None-Match` answers 304 |
| `GET /api/v1/agent/events` | signature | server-sent events: one `config` event per change, a `mesh` event when the mesh view changed |
| `POST /api/v1/agent/status` | signature | heartbeat every 30 s and after each apply |
| `POST /api/v1/agent/prefs` | signature | the node changes its own exit, keepalive or MTU (amendment A10) |
| `GET /api/v1/agent/mesh` | signature | every node as the dashboard shows it, for the node's local view: no keys, endpoints, hostnames or errors (amendment A31), and the containers' published names (A38) |
| `PUT /api/v1/agent/names` | signature | the names the node's container runtime gives the containers on its mesh network (amendment A38) |
| `GET /healthz` | none | the container healthcheck: serving, and the database answers |
| `GET /install.sh` | none | the Linux installer, with the binaries' SHA-256 baked in |
| `GET /dl/{file}` | none | `meshctl-linux-amd64`, `meshctl-linux-arm64`, `meshctl-windows-amd64.exe`, `SHA256SUMS` |

Admin socket (`<data>/admin.sock`, mode 0600, so only the server's own user reaches it):

| Route | Purpose |
| --- | --- |
| `GET /admin/v1/nodes` | list nodes and the mesh plan's warnings |
| `POST /admin/v1/nodes` | add a node (`ip_family`, `ip6`, `endpoint6` and `containers6` are the IPv6 side, amendment A46); the answer carries its first join token, shown once (a phone gets none) |
| `GET /admin/v1/nodes/{name}` | one node, its computed config and its last peer report |
| `PATCH /admin/v1/nodes/{name}` | change endpoint, listen port, hub, exit capability, exit, keepalive, MTU, a host-mode node's container network (`containers`), the mesh IP (`mesh_ip`, amendment A41; the answer's `notes` say what the move does not reach) the segmentation groups (`groups`, amendment A43) and the IPv6 side (`ip_family`, which allocates the address a family adds and drops the one it takes away; `mesh_ip6`; `endpoint6`; `containers6`; amendment A46) |
| `DELETE /admin/v1/nodes/{name}` | remove a node |
| `POST /admin/v1/nodes/{name}/disable` | take a node out of every peer list; its stream is closed |
| `POST /admin/v1/nodes/{name}/enable` | put it back |
| `POST /admin/v1/nodes/{name}/token` | issue a new join token, revoking unused ones |
| `GET /admin/v1/nodes/{name}/compose` | the compose file that runs a node's agent in a container: a container-mode node's (`?runtime=docker\|podman\|swarm&backend=auto\|kernel\|userspace&data=<host dir>&names=off`; `names=off` leaves out the runtime's socket) or a host-mode spoke's on the host network (`?backend=userspace\|kernel&data=<host dir>`), all optional (`meshctl node compose`) |
| `GET /admin/v1/nodes/{name}/network` | the commands that create a node's mesh network on its host and the compose lines that attach a service; `?runtime=docker\|podman\|swarm` (`meshctl node network`) |
| `POST /admin/v1/nodes/{name}/phone` | a phone's WireGuard profile with a new key pair; `{"exit": "<server>"}` (optional) sends its internet through that exit; the answer is the only copy of the private key (`meshctl node phone`) |
| `GET /admin/v1/segmentation` | the mesh firewall (amendment A43): the document (groups, address and service objects, zones, policies), the rules it compiles to, each zone's addresses and inbound/outbound rules, and the plan's warnings (`meshctl firewall show`) |
| `PUT /admin/v1/segmentation` | replace the whole document; checked against the mesh like every change (`meshctl firewall apply`) |
| `GET /admin/v1/ipv6` | the mesh's IPv6 prefix (empty for none) and how many nodes have an IPv6 mesh address (`meshctl ipv6`, amendment A46) |
| `PUT /admin/v1/ipv6` | `{"prefix": "<IPv6 /48 to /64>\|auto\|off"}`: a prefix gives every node whose `ip_family` allows IPv6 an address in its role's block; `auto` keeps the prefix there is or draws a random unique local /64; `off` (or empty) takes every IPv6 mesh address and container /112 away, refused while a node is IPv6 only, and is remembered (the setting `mesh_cidr6` holds `off`); a mesh whose IPv6 was never chosen gets an `auto` prefix at the server's start, as `system`; a prefix that holds addresses is not swapped for another (`meshctl ipv6 <prefix>\|auto\|off`) |
| `GET /admin/v1/segmentation/check` | what the policy does with one flow: `?from=<node or address>&to=<node or address>&service=tcp/PORT\|udp/PORT\|icmp[&at=<RFC 3339>]` (default now; two nodes meet over IPv4 when both have it, else IPv6); the answer names the rule that decided and the scheduled rules not in force then (`meshctl firewall check`) |

Every admin write goes through the reconciler (below) with actor `cli` and is audited. Join
tokens appear only in the answer to add and token, never in the audit log or the server log.

## Signed requests

Every agent request except join carries:

```
X-Mesh-Node:      <node id>
X-Mesh-Timestamp: <unix seconds>
X-Mesh-Nonce:     <26 base32 characters, fresh per request>
X-Mesh-Signature: base64(ed25519_sign(identity_key, canonical))

canonical = METHOD "\n" PATH?RAWQUERY "\n" TIMESTAMP "\n" NONCE "\n" hex(sha256(body))
```

The query string is covered so a relay cannot rewrite parameters. The nonce exists because
Ed25519 signatures are deterministic: without it, two identical requests in the same second (a
poll and an SSE-triggered fetch) would carry the same signature, and the second would be refused
as a replay.

The server checks, in this order, and stops at the first failure:

| Check | Refusal |
| --- | --- |
| body at most 256 KiB | 413 `too_large` |
| headers well formed | 401 `unauthorized` |
| timestamp within +-60 s of the server clock | 401 `clock_skew`, with `server_time` |
| signature not seen in the last 120 s | 401 `replay` |
| node exists and has joined | 401 `unauthorized` |
| signature verifies with the node's identity key | 401 `unauthorized` |
| node is not disabled | 403 `node_disabled` |

The first four need no database read. A signature is remembered only after it verifies, and when
the replay cache is full of live entries the server answers 503 `rate_limited` rather than evict
one (an evicted signature would become replayable).

## Errors

Every non-2xx answer is JSON:

```json
{"code": "token_invalid", "message": "join token was already used at 2026-10-12T08:00:00Z; issue a new one with `meshctl node token <name>`"}
```

Agents branch on `code`, never on `message`. Codes: `bad_request`, `unauthorized`,
`clock_skew`, `replay`, `node_disabled`, `token_invalid`, `not_found`, `conflict`,
`rate_limited`, `too_large`, `internal`. Messages say what failed and what to do; they never
contain a token or a key. A 500 says only "internal error"; the detail is in the server log.

## Join

`POST /api/v1/agent/join`, unsigned, rate-limited to 10 attempts per client IP per minute:

```json
{"token": "...", "name": "vps-2", "hostname": "vps-2.example.com", "os": "linux", "arch": "amd64",
 "wg_backend": "kernel", "wg_public_key": "...", "identity_public_key": "...",
 "endpoint": "203.0.113.11:51820", "public_ip": "203.0.113.11", "agent_version": "v1.0.0"}
```

`name` is optional: when given it must be the node the token was issued for. The token is
consumed in the same transaction that registers the keys and recomputes the mesh, so a join that
fails for any reason (a disabled node, a key another node already uses, a plan error) leaves the
token usable. Phones cannot join this way; they enroll by QR code (M4).

A server node's endpoint is `endpoint` (from `--endpoint`), else `public_ip` (the first public
IPv4 on the agent's interfaces) with the node's listen port, else the endpoint set with
`meshctl node add --endpoint`, else the address the request came from if it is public
(amendment A17). Its IPv6 endpoint is chosen the same way from `endpoint6` (`--endpoint6`),
`public_ip6` (the first global IPv6 on its interfaces), `node add --endpoint6` and a global IPv6
source address (amendment A46); an IPv4-only server keeps none, an IPv6-only one no IPv4
endpoint. Spokes ignore all of these.

The answer: `{"node_id", "name", "role", "mesh_ip", "config_version", "server_time"}`, and
`mesh_ip6` when the node has one.

## Config and events

`GET /api/v1/agent/config` returns the node's `NodeConfig` (the shape in `internal/proto/config.go`)
with `ETag: "<hash>"`, where the hash is the SHA-256 of the canonical encoding with `version`
zeroed. An agent sends `If-None-Match` with its applied config's ETag and gets 304 when nothing
changed.

**Dual stack** (M8). A mesh with an IPv6 prefix (the setting `mesh_cidr6`) gives a node an IPv6
mesh address beside its IPv4 one, or instead of it: `interface.address6` (a /128) next to
`interface.address` (a /32), either of them absent. Every peer's `allowed_ips` holds the families
both ends have, a relay is chosen per family, and a server peer's `endpoint` is in a family the
node's ip_family lets it dial, IPv4 first, with the other family's endpoint as `fallback`. An exit
peer carries `::/0` as well as `0.0.0.0/0` when both ends have IPv6 and the exit has an IPv6
endpoint; a server that is such an exit has `firewall.exit6`, the IPv6 prefix it masquerades
(firewalld's policy masquerade is IPv4 only). A record has `ip`, `ip6` or both. Containers whose
IPv6 is on the mesh add `subnet6` (a /112), `network6`, `range6`, `bridge6` and `gateway6` to the
`containers` section, and `sysctl` asks for `net.ipv6.conf.all.forwarding` (and
`net.ipv6.conf.all.proxy_ndp` for a container-mode gateway). Every one of these is absent in an
IPv4-only mesh, whose configs hash as before. The server gives a new mesh a prefix at its first
start; `PUT /admin/v1/ipv6` (above) changes or gives it up, node add and set a node's families
and IPv6 addresses; amendment A46.

Its `dns` section (servers and spokes; spec S7.4, amendment A12) is the whole mesh zone: `records`
holds one name per enrolled, enabled node and the custom records (an alias of a node that is not
on the mesh is left out), sorted by name; `upstream` is the operator's resolvers in their order;
`split` is false only in exit mode, when the node forwards every other name to `upstream`
through the tunnel; `forward_for` lists, on a server, the /32 of each phone whose hub it is,
whose other lookups it forwards too; `domains` (amendment A39, absent when there are none) are
the extra domains, sorted: the resolver answers the containers' names in them, forwards every
other name in them for anyone, and the host registers them with the zone. A server's `firewall` holds udp/51820 in the public zone and,
in the mesh zone, one rule that accepts everything from the mesh CIDR (`proto` `all`, port 0;
amendment A43): what crosses is the `policy`'s to decide (`note` says where a rule comes from; it
never reaches firewall-cmd).

A server's `policy` (amendment A43, the last field of its config; no other node has one) is the
mesh firewall: `mesh` (the mesh CIDR) and `rules` in order. A rule has a `name` (display only:
`system: ping`, `policy: <name>`, `zone: <name>`), `from` and `to` (addresses, networks and ranges
in canonical form, sorted and merged; absent is any), `services` (each a `proto` -- `tcp`, `udp`
or `icmp`, which is ping -- with `ports` such as `22` or `1000-2000`; absent is every port, and
absent services any traffic), `action` (`allow` or `deny`) and, for a scheduled policy
(amendment A44), `schedule`: `days` (`mon`..`sun` in the week's order; absent is every day),
`start` and `end` (`HH:MM`; absent is the whole day, an `end` before `start` ends the next day,
and a window belongs to the day it starts), read on the wall clock of the policy's `tz` (an IANA
name, present only when a rule has a schedule). A scheduled rule takes part only while its window
holds. A packet between two mesh
addresses that the server relays, receives or sends is checked against the rules in order; the
first that matches decides, and one none matches is dropped. Replies and ICMP errors of a flow
that was allowed always pass; traffic with an end outside the mesh (an exit's) is not mesh
traffic. The rules start with the system rules (ping across the mesh, DNS to the servers, Swarm
between them), then the enabled policies, then an allow within each zone that does not deny.

A node whose containers are on the mesh (amendment A37) has a `containers` section, the last
field of the config and absent otherwise, so other configs hash as before: `mode` (`host` or
`container`), `subnet` (its /24), `network` (the whole mesh, on the link of the runtime's
network, so an attached container ARPs for every mesh address), `range` (the subnet's upper /25,
which the runtime hands out), `bridge` (.1) and `gateway` (the agent's .2 in container mode, the
bridge in host mode), and `relay` on a container-mode hub that relays. Every peer routes the
subnet wherever it routes the node's /32, and every node's `forward_for` holds every container
network's subnet (a Swarm service's tasks on several nodes share one resolver list). The node's `sysctl` asks for
`net.ipv4.ip_forward` and, in container mode, `net.ipv4.conf.all.proxy_arp` (both set by the
compose file, so the agent only verifies them). A container-mode node has no `firewall` and no
`exit`; a host-mode server's `firewall` names the containers' `bridge` (`meshctl0`) and relays.

`GET /api/v1/agent/events` is a server-sent event stream with `X-Accel-Buffering: no`. It opens
with `retry: 5000` and the node's current version, then sends

```
event: config
data: {"version":42,"hash":"<hex>"}
```

whenever the node's computed config changes,

```
event: mesh
data: {}
```

to every stream when the containers' names may have changed (amendment A38; the agent fetches
the mesh view), and `: ka` every 20 s. A node has at most one stream: a new one ends the previous
one, and disabling or removing the node ends it. Agents also poll config every 60 s, for proxies
that swallow streams (amendment A16), and fetch the mesh view after every new config.

## Status

`POST /api/v1/agent/status` carries the applied version and hash, the agent version, the
WireGuard backend in use (`kernel` or `userspace`), per-peer `last_handshake`, `rx_bytes`,
`tx_bytes` and `endpoint`, the last errors, the degraded steps, any rollback and `dns`: `on` (the
host sends .mesh lookups to the node's resolver), `off` (`meshctl dns off` on the node) or
`unavailable` (the resolver answers on the mesh IP, but there is no systemd-resolved to register
with), and `families`, the address families the host can use (`ipv4`, `ipv6`; absent when the
agent does not say). The server keeps it as operational state (never part of a config version)
and audits a rollback once: the agent repeats it in every heartbeat until a newer config is
confirmed, and only a different rollback writes another `node.rollback` row. A changed `families`
goes through the reconciler as the node (`node.families`, amendment A46): the plan leaves a
family the host cannot use out for that node -- its address, endpoint and containers in that
family -- so no peer routes it there and DNS does not name it there, unless the node cannot do
without it (no other family, or a server or a container network, which need IPv4); either way the
plan warns, and it is never refused for it. The answer carries the
node's current version and hash, so an agent that missed an event catches up here; on a server
node it is also what confirms a safe apply (below).

## Prefs

`POST /api/v1/agent/prefs` takes `{"exit": "vps-2"}` (`""` turns exit off), `{"keepalive": 15}`
or `{"mtu": 1380}`, any combination, and nothing else. The server applies it through the same
write path as `meshctl node set`, audited as `node.set` with actor `node:<id>`, so it is a new
config version pushed to every node it changes. Only a spoke chooses an exit or a keepalive (400
otherwise); an exit that is not exit-capable is a plan error (409); an unknown one is 404. The
answer is `{"version", "config_version", "config_hash"}`: `version` is 0 when nothing changed,
and the agent has applied the change once its applied hash equals `config_hash`. `meshctl exit`
on the node is the usual caller, through the local API.

## Mesh view

`GET /api/v1/agent/mesh` answers `{"version", "server_time", "nodes"}`, the mesh's latest
version and every node as the web dashboard shows it: `name`, `role`, `mesh_ip`, `hub`, `exit`,
`can_exit`, `health` (the dashboard's colour: `ok`, `warn`, `bad` or `off`), `state` (its label,
such as `in sync` or `silent for 6m0s`), `last_seen` (for a phone, the latest handshake a server
reported) and `config_version` and `applied_version`. Any joined, enabled node may ask. Keys,
endpoints, hostnames, observed addresses and error text are never in it (amendment A31). The agent
validates the view (`proto.MeshView.Validate`) before it caches or shows it.

`names` (amendment A38, absent when there are none) are the containers' published names, sorted:
`{"name": "web", "node": "nas", "ip": "10.77.128.129"}`; a dotted name is one inside an extra
domain (amendment A39), answered as it is. Every node's resolver answers
`<name>.<zone>` with every node's addresses of that name, rotated on each answer, and
`<name>.<node>.<zone>` with that node's, both with a TTL of 10 s; a name of the config (a node,
a custom record) wins. They are in the view, not in `NodeConfig`, so a container that starts or
stops is not a config version; mesh.json keeps them, so they resolve while the server is down.

## Container names

`PUT /api/v1/agent/names` takes `{"names": [{"name", "kind", "ip"}]}`: every name the node's
container runtime gives the containers on its `meshctl-mesh` network, at most 512, lower case,
each name once per address. `kind` says where it came from: `service` (the compose service),
`swarm` (the Swarm service without its stack's prefix), `container` (the container's name),
`alias` (a network alias) or `hostname` (the hostname, or hostname.domainname). The list replaces
what the server held; an empty one clears it. A node without a container network is refused
(409). An address is of either family: a container with IPv6 on the mesh (M8) has each name once
per address, so 512 entries are 256 such names. The server keeps a name only when its address is one of the node's containers' (in its
/24, not .0, .1, .2 or .255; in its /112, not ::0, ::1 or ::2), and publishes it only when it cannot shadow anything: a name in the
zone is one label (`web`, or `web.<zone>`, which is the same), any other name lies inside an
extra domain (not the domain itself), a node's name, a custom record's and the server's own
address stay theirs, and a disabled node's names are not published. Names
are operational state, like status: never in a config version, a snapshot or an export. The
answer is `{"stored", "dropped"}`: how many it keeps, and each name it does not publish with
`reason`, which the agent logs. A report that changed anything sends every stream a `mesh` event.
The agent sends one a second after its runtime reports a change (or a new config moves its
network), and lists the containers every 2 minutes in case an event was missed; it sends nothing
when the names did not change.

## The reconciler

Every change to desired state -- a join, every admin write, and a recompute at server start --
runs in one transaction: apply the change, recompute every node's config, compare each node's
hash with the stored one. If any changed, a config version is written with a snapshot of the
desired state, the changed nodes' hashes and versions are updated, and after commit exactly those
nodes get a `config` event. A change that would make the plan invalid is refused as a whole (409
`conflict`, listing the problems) and nothing is written.

Relay failover (amendment A40) writes through the reconciler too, as actor `system`:
`relay.failover` when a hub's heartbeat is older than `--hub-down-after` and no node reported a
handshake with it for 1.5 times that, `relay.restore` once its heartbeat has been steady for 2.5
times that. The down mark (`nodes.down_since`) is operational state, never in a snapshot; the
planner then relays each pair of spokes that used the hub through the other spoke's hub, else the
default hub, else the lowest-id server that is up. A phone keeps its hub.

## Apply (agent)

The agent applies a config in fixed steps, each comparing desired with observed state:

| # | Step | On failure |
| --- | --- | --- |
| 1 | validate the whole config | fatal; nothing touched |
| 2 | interface: create `wg0` with alias `meshctl` (never adopt one without it), MTU, up | fatal |
| 3 | address: the node's /32 and /128 (either may be absent, M8) are the only addresses; the host's IPv6 link-local ones are left alone, and the IPv6 one skips duplicate address detection | fatal |
| 4 | WireGuard: key, port, fwmark, then per-peer changes (no full replace) | fatal |
| 5 | routes: every AllowedIP of both families except 0.0.0.0/0 and ::/0, protocol 201 | fatal |
| 6 | exit mode: policy rules and table 51820 (below), or their removal; on Windows split default routes and a bound socket (amendment A35) | fatal |
| 7 | sysctls (relay and exit servers, container networks); before IPv6 forwarding, every interface taking router advertisements (`accept_ra` 1) moves to `accept_ra` 2, so it keeps its IPv6 default route | fatal |
| 8 | the mesh firewall (servers, amendment A43, below): the `policy` as the table `inet meshctl`, before step 9 lets the mesh through the mesh zone | fatal; a config without a policy removes the table after step 9, and keeps it while step 9 fails (degraded `policy` when the removal fails) |
| 9 | firewall (servers, firewalld; an unzoned containers' bridge goes into the mesh zone) | degraded; the tunnel stays up |
| 10 | containers (amendment A37, below) | degraded; the tunnel stays up |
| 11 | DNS: the resolver on the mesh IP and the containers' gateway, then the host's registration (systemd-resolved; NRPT on Windows) | degraded; the tunnel stays up |

**A host that cannot use IPv6** (M8: `net.ipv6.conf.all` or `.default.disable_ipv6` is 1, or the
kernel has no IPv6) applies a dual-stack config without its IPv6 side -- the IPv6 address, routes
and endpoints (a server whose IPv6 endpoint comes first is dialled on its IPv4 fallback), the
IPv6 sysctls, the exit's IPv6 masquerade, the containers' IPv6 -- and reports `families: ["ipv4"]`;
the server then plans the node without IPv6 (Status, above). An IPv6-only node cannot run on such a
host: step 1 fails and says so. The hash it reports is of the config as sent.

A config is cached (`config.json`) only after steps 1-8 succeed, and re-applied at boot, once the
agent holds its local API socket and before the server is contacted. A config that failed is not retried until a newer one arrives or five
minutes pass.

**The mesh firewall** (step 8, `internal/agent/policy`). A server compiles its `policy` into one
nftables table of its own, `table inet meshctl` (both families, M8; an agent's older `table ip
meshctl` is deleted in the same transaction), in its own network namespace (the host's, or a
container-mode server's): base chains on forward, input and output accept established and related
traffic, the node's own traffic (loopback, its containers' subnet) and anything with an end
outside the mesh, and jump to `mesh_forward` (matched on both addresses), `mesh_input` (on the
source; the server is the destination, whatever address the packet names) or `mesh_output` (on the
destination): the policy's rules in order, each service its own rule (`icmp` is an echo request),
then drop. Addresses and ports are anonymous interval sets; the only string in the table is the
first rule's comment, `meshctl policy <fingerprint>`, by which the agent sees whether the kernel
holds the table it installed. A different table is replaced whole in one transaction, and the
connection tracking of flows between two mesh addresses is then deleted, so open connections
are judged by the new policy at their next packet. The step runs at every apply and every drift
check; an unchanged table costs one read. `nft list table inet meshctl` shows it. With the mesh's IPv6 prefix
(`policy.mesh6`), a policy rule becomes one rule per family it names -- `meta nfproto ipv4` with `ip
saddr`/`ip daddr`, `meta nfproto ipv6` with `ip6 saddr`/`ip6 daddr` -- and ping is ICMP echo or
ICMPv6 echo; a rule that names only one family's addresses has no rule in the other, and the
connection tracking deleted on a change is both families'. A policy with
schedules (amendment A44) is compiled for the time of the apply, leaving out the rules whose
window does not hold, and the agent applies the config again when the next window starts or ends:
a timer on that edge, with the drift check as the net when the clock is set or a
daylight-saving gap swallows an edge. A window that closes replaces the table like any change, so
the connections it allowed are judged again and cut. The server's clock decides: keep servers on
NTP.

**The container network** (step 10, `internal/agent/containers`). In container mode the agent is
its containers' gateway in its own namespace: it adds `gateway` (.2) to the interface the runtime
gave an address of the subnet, and isolates by policy routing -- forwarded traffic from wg0 may
reach the subnet (never the bridge address), from the mesh side only the subnet may reach the
mesh, a relaying hub forwards wg0 to wg0, and anything else forwarded meets a blackhole table
(51821); the agent's own traffic and local delivery are untouched. In host mode the agent owns the
bridge `meshctl0` (it makes it when it is missing, holds `bridge` on it, and turns on its
forwarding and proxy ARP and wg0's forwarding), and four rules let only wg0, the bridge and the
host itself reach the subnet. Every rule carries protocol 201 at a priority of its own (1000-1006,
5200-5204), so exit mode's rules and the host's are never touched; without a `containers`
section the rules and the blackhole route go, the addresses and the bridge stay. With the
containers' IPv6 (`subnet6`, M8) the same rules hold for IPv6 (`ip -6 rule`, an IPv6 blackhole
default in table 51821) with the node's /112 and the mesh's IPv6 prefix: the gateway claims
`gateway6` (::2) beside .2, a host-mode bridge gets `bridge6` (::1) and proxy NDP, and IPv6 from
the bridge may go only to the mesh -- anything else is refused at once (rule 5204, unreachable),
since Docker makes the mesh bridge its containers' IPv6 default route and an app would otherwise
wait seconds before falling back to IPv4. A container sees the whole IPv6 prefix on its link,
and IPv6 has no proxy ARP, so the mesh side answers neighbour discovery by one proxy entry per
mesh address: the node's own IPv6 address and every single IPv6 address in its peers' allowed
IPs, and every IPv6 address of the mesh view's container names (other nodes' containers; a
change to those applies the config again). A kernel without
IPv6 policy routing cannot isolate IPv6: the step sets up IPv4, leaves the IPv6 side out, drops
the /112 by a blackhole route of its own in the main table (device-less, so the routes step,
which keeps wg0's, never touches it) and is degraded, saying to take the containers' IPv6 off
(`--containers6=false`); the IPv6 address and entries of a network whose IPv6 went stay, as its
addresses do.

**The drift check** applies the config on the interface again every 5 minutes, after the firewall
step forgets what it saw: zone, interface binding (at runtime too), the recorded rules at
runtime, the policies. Whatever changed underneath the agent -- wg0 moved to another zone, a rule
removed by hand, a route or the interface deleted -- is put back and logged; an unchanged host
costs only the checks.

**DNS** (step 11): the resolver binds UDP and TCP 53 on the node's mesh IP -- and on its IPv6 one
too when it has both (M8) -- and answers the zone authoritatively: A and AAAA records (both for
ANY), NODATA for another type of a name it has (an AAAA of a name with no IPv6), NXDOMAIN for one it
has not.
A name outside the zone is forwarded to `upstream` (over the transport the query came on) when
`split` is false or the source is in `forward_for`, and REFUSED otherwise. On Linux it then
registers with systemd-resolved over the system bus, on wg0 only: the mesh IP (and the IPv6 one)
as the link's DNS servers and `~mesh` as its routing domain, plus `~.` and default-route in exit mode, with DNSSEC,
LLMNR and mDNS off for the link. It compares what resolved holds first, so a repeat changes
nothing, and never starts resolved where it is not running. The registration is taken back on
`dns off` and when the agent stops. An agent in another network namespace than the host's (the
e2e lab, a container without the host's network) does not register: the host's resolver would read
its interface index as one of the host's own interfaces.

**Exit mode** (spokes, spec S7.3) is wg-quick's policy routing. The planner puts `0.0.0.0/0` on
the exit server's peer; step 6 adds, for IPv4 and IPv6:

| Priority | Rule |
| --- | --- |
| 32764 | `from all lookup main suppress_prefixlength 0` |
| 32765 | `not from all fwmark 0xca6c lookup 51820` |

and in table 51820 `default dev wg0` (IPv4) and a blackhole default (IPv6, so nothing leaks
around the tunnel; amendment A11) -- or, when the exit carries IPv6 (M8: `::/0` on its peer, an
exit with public IPv6 for a node with IPv6), `default dev wg0` for IPv6 as well; the route is
replaced in place when the exit changes, so table 51820 is never without one. The exit
masquerades the mesh's IPv6 prefix (`firewall.exit6`) with the rich rule `rule family="ipv6"
source address="<prefix>" masquerade` in its `meshctl-exit` policy, since firewalld's policy
masquerade is IPv4 only. On Windows the split default routes `::/1` and `8000::/1` go into the
tunnel and are carried when the exit's peer has `::/0`, dropped when not; the tunnel's IPv6
socket is bound to the interface of the host's IPv6 default route (its packets dropped while
there is none, never sent into the tunnel). Every specific route of the main table still wins; WireGuard's
own UDP carries fwmark 51820 and takes the normal path. Switching exits moves `0.0.0.0/0` between
peers and leaves the rules alone, so no connection drops; exit off removes them. The rules and
routes carry protocol 201, so only the agent's own are ever listed or removed (amendment A27).
A kernel with IPv6 but without IPv6 policy routing (it refuses IPv6 rule dumps with EAFNOSUPPORT,
as some NAS kernels built without IPV6_MULTIPLE_TABLES do) has no IPv6 rules to manage: step 6
handles IPv4 alone, and with an exit on it is degraded (`exit`), not fatal, since IPv6 can then
neither go through the exit nor be held back (amendment A45).

**Endpoint failover** (M8). A server peer may come with a `fallback` endpoint in the other
family. At every heartbeat the agent looks at each peer it keeps alive (a spoke's servers) that
has one: when the endpoint it dials has had no handshake for 3 minutes, or none at all 45 seconds
after it was set, the agent dials the other one instead, and back the same way; a re-apply keeps
the choice until a config gives the peer other endpoints. A laptop on an IPv4-only network thus
reaches a server over IPv4, one behind IPv6-only access over IPv6, without a setting. Server to
server tunnels are idle between packets, so they are not judged; the planner gives them the
endpoint of the first family both servers have.

**Safe apply** (server nodes, spec S6.4, amendment A15): a config the server pushes over a
previous one is applied at once and written to `config.pending.json`; `config.json` stays the last
confirmed config. A successful answer to a status report sent after the apply confirms it (the
pending config becomes `config.json`). Without one within 60 s the agent puts the confirmed config
back, reports `rollback {from, to, reason}` in every heartbeat, and does not apply the rejected
config again until the server's config differs. An agent that stops while a config is pending
boots from the confirmed one and reports the rollback the same way. A config read from the cache
at boot, and a node's first config, never arm it.

## Local API

`meshctl status` reads `GET /v1/status` from the agent's Unix socket (`/run/meshctl/agent.sock`
for the standard data dir, else `<data>/agent.sock`; mode 0660 with group `meshctl` when that
group exists, else 0600). On Windows it is the named pipe `\\.\pipe\meshctl-agent`, which only
Administrators and SYSTEM may open. It never listens on the network. The agent takes it before it
applies the cached config or contacts the server: when another agent already answers on it, or the
path is unusable, the agent stops there, leaving the running agent's interface and safe apply alone.

`meshctl exit` posts a prefs request to `POST /v1/prefs`. The agent validates it, relays it to the
server as a signed prefs request, fetches the new config at once, and answers the server's answer
-- or the server's refusal with its status and message, or 502 when the server cannot be reached.

`meshctl dns on|off` posts `{"on": true|false}` to `POST /v1/dns`. The choice is the node's own
(amendment A10): the agent keeps it in `<data>/local.json`, applies it to the config on the
interface at once, and answers its status after; nothing goes to the server.

`GET /v1/mesh` answers the mesh view as the node has it: `{"view", "fetched_at", "error"}`. The
agent asks the server at start, every 5 minutes, and when the view is asked for and its last try
is over 10 s old or a new config arrived since; each view it gets is cached in `<data>/mesh.json`.
When the server does not answer, the answer is the view held -- after a restart, the cached one --
with `error` saying why, so the node's view of the mesh stays readable while the server is down.

`meshctl ui` posts to `POST /v1/ui-token` and gets `{"url", "expires_at"}`: a sign-in link to
the node's local UI (amendment A32) that works once, within 2 minutes. When the UI is off or could
not listen, the answer is 409 with the reason.

`meshctl logs` reads `GET /v1/logs`: the agent's last 2000 log lines as text, a line cut at
8 KiB. With `?follow=true` the answer stays open and carries each new line until the reader leaves
or the agent stops; a reader 256 lines behind is dropped, with a last line saying so. The lines
are the ones the agent writes to the journal, kept in memory only.

Errors are `{"code", "message"}` like the agent API's.
