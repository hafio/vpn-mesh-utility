# Compose and stack files

Every compose and stack file meshctl uses, in one place. The server's are the files in
[deploy/](../deploy); an agent's are what the server renders for a node, shown here for example
nodes of the mesh 10.77.0.0/16 (zone `mesh`, server https://mesh.example.com, release v1.0.0). A node's own file has its
name, its container subnet and the server's release in place of these: get it with
`meshctl node compose <name>` (with `--runtime docker|podman|swarm` in container mode) or from
the node's page in the web UI. Generated from the renderers by `./scripts/dev.sh regen`
(`TestComposeDoc`); do not edit.

| Node | Runs as | File |
| --- | --- | --- |
| the server, Docker or Dokploy compose app | a container on Traefik's network | [Server: Docker Compose](#server-docker-compose) |
| the server, Docker Swarm or Dokploy stack | one Swarm task, pinned to a labelled node | [Server: Docker Swarm stack](#server-docker-swarm-stack) |
| a Linux host (VPS, laptop) | the meshctl binary under systemd | no compose: `curl -fsSL https://mesh.example.com/install.sh \| sudo sh -s -- --token <token>` |
| a NAS or any spoke on the host network | a container with the host's network | [Client: host network](#client-host-network-nas) |
| a host whose containers only are on the mesh | the agent's container, their gateway | [Container mode](#container-mode) |
| a host-mode node's containers (Swarm services on the VPSes) | the host's own agent | [A host-mode node's containers](#a-host-mode-nodes-containers) |

## Server: Docker Compose

`deploy/compose.server.yml`: a Dokploy compose app, or `docker compose up -d` with the variables in
its header set.

```yaml
# meshctl server as a Dokploy compose app (spec S12, amendment A13).
# Traefik, on the external dokploy-network, routes the server's domain to
# port 8080; nothing is published on the host. Set in the app's environment:
#   MESHCTL_VERSION  the release to run, e.g. 1.0.0
#   MESHCTL_URL      the public origin, e.g. https://mesh.example.com
#   MESHCTL_TRUSTED_PROXIES  optional: Traefik's network CIDR, so the server
#                    believes X-Forwarded-For from it
#   TZ               optional: the time zone the web UI's times are shown in,
#                    e.g. Asia/Singapore (default UTC)
# Posture: non-root, no capabilities, no privilege escalation, read-only
# root file system; deploy_test.go forbids every widening token.
name: meshctl
services:
  meshctl:
    image: ghcr.io/hafio/vpn-mesh-utility:${MESHCTL_VERSION:?set MESHCTL_VERSION to a release such as 1.0.0}
    command:
      - server
      - --data
      - /data
      - --listen
      - ":8080"
      - --url
      - ${MESHCTL_URL:?set MESHCTL_URL to the public https origin}
      - --trusted-proxies
      - ${MESHCTL_TRUSTED_PROXIES:-}
    environment: ["TZ=${TZ:-UTC}"]
    user: "65532:65532"
    privileged: false
    cap_drop: [ALL]
    security_opt: ["no-new-privileges=true"]
    read_only: true
    tmpfs: ["/tmp:size=16m,mode=1777"]
    volumes: ["meshctl-data:/data"]
    expose: ["8080"]
    networks: [dokploy-network]
    restart: unless-stopped
    mem_limit: 256m
    pids_limit: 256
    healthcheck:
      test: ["CMD", "/app/meshctl", "healthcheck", "server"]
      interval: 30s
      timeout: 5s
      retries: 3
      start_period: 10s
volumes:
  meshctl-data: {}
networks:
  dokploy-network:
    external: true
```

## Server: Docker Swarm stack

`deploy/stack.server.yml`: `docker stack deploy` or a Dokploy stack app, one task on the node labelled
`meshctl.server=true`.

```yaml
# meshctl server as a Docker Swarm stack (spec S12): `docker stack deploy`, or a
# Dokploy "Stack" app. Traefik, on the external dokploy-network, routes the
# server's domain to port 8080; nothing is published on the host. Set:
#   MESHCTL_VERSION  the release to run, e.g. 1.0.0
#   MESHCTL_URL      the public origin, e.g. https://mesh.example.com
#   MESHCTL_TRUSTED_PROXIES  optional: Traefik's network CIDR, so the server
#                    believes X-Forwarded-For from it (sign-in lockout and join
#                    limits are per client); empty trusts nobody
#   TZ               optional: the time zone the web UI's times are shown in,
#                    e.g. Asia/Singapore (default UTC)
#
#   docker node update --label-add meshctl.server=true <node>   # once
#   docker stack config -c stack.server.yml                      # checks it
#   docker stack deploy -c stack.server.yml meshctl
#
# What Swarm changes against compose.server.yml:
#   - exactly one task, stopped before a new one starts: SQLite has one writer
#     and the server holds a lock on /data
#   - pinned by a node label, not just to a manager: the data volume is local
#     to one node, and another node would start the server on an empty one
#   - no privileged, security_opt or pids limit, which a stack cannot set: the
#     server runs as 65532 with every capability dropped in a distroless image
#     with no setuid binaries, so there is nothing to escalate through
#   - no container_name: the task's container is <stack>_meshctl.1.<id>
# Without Dokploy, route the domain with Traefik's labels under deploy.labels:
#   traefik.enable=true
#   traefik.http.routers.meshctl.rule=Host(`mesh.example.com`)
#   traefik.http.services.meshctl.loadbalancer.server.port=8080
# Posture: non-root, no capabilities, read-only root file system;
# deploy_test.go forbids every widening token.
services:
  meshctl:
    image: ghcr.io/hafio/vpn-mesh-utility:${MESHCTL_VERSION:?set MESHCTL_VERSION to a release such as 1.0.0}
    command:
      - server
      - --data
      - /data
      - --listen
      - ":8080"
      - --url
      - ${MESHCTL_URL:?set MESHCTL_URL to the public https origin}
      - --trusted-proxies
      - ${MESHCTL_TRUSTED_PROXIES:-}
    environment: ["TZ=${TZ:-UTC}"]
    user: "65532:65532"
    cap_drop: [ALL]
    read_only: true
    volumes:
      - meshctl-data:/data
      - type: tmpfs
        target: /tmp
        tmpfs:
          size: 16777216
    networks: [dokploy-network]
    stop_grace_period: 15s
    healthcheck:
      test: ["CMD", "/app/meshctl", "healthcheck", "server"]
      interval: 30s
      timeout: 5s
      retries: 3
      start_period: 10s
    deploy:
      mode: replicated
      replicas: 1
      placement:
        constraints: [node.labels.meshctl.server == true]
      update_config:
        parallelism: 1
        order: stop-first
        failure_action: rollback
      rollback_config:
        parallelism: 1
        order: stop-first
      restart_policy:
        condition: any
        delay: 5s
      resources:
        limits:
          memory: 256M
volumes:
  meshctl-data: {}
networks:
  dokploy-network:
    external: true
```

## Client: host network (NAS)

A spoke whose agent runs in a container on the host's network (`meshctl node compose nas`): the
whole host is on the mesh. Userspace WireGuard (the default; needs /dev/net/tun):

```yaml
# meshctl agent for the spoke "nas", as a container (spec S12; amendments A13, A19).
# The userspace backend runs WireGuard in the agent, on the host's /dev/net/tun.
#
# Join once, with the node's join token (web UI, `meshctl node add` or `node token`):
#   MESHCTL_TOKEN=<token> docker compose run --rm -e MESHCTL_TOKEN meshctl-agent \
#     agent join --data /data --server https://mesh.example.com --name nas --no-service
# then start it: docker compose up -d
#
# Posture (docs/security.md): not privileged; every capability dropped except
# NET_ADMIN (wg0, its routes) and NET_BIND_SERVICE (DNS on :53); no privilege
# escalation; read-only root. It runs as uid 0 because capabilities do not
# survive no-new-privileges for another user, and on the host network because
# wg0 belongs to the host.
name: meshctl-agent
services:
  meshctl-agent:
    image: ghcr.io/hafio/vpn-mesh-utility:1.0.0
    command:
      - agent
      - run
      - --data
      - /data
      - --wg-backend
      - userspace
    user: "0:0"
    privileged: false
    network_mode: host
    cap_drop:
      - ALL
    cap_add:
      - NET_ADMIN
      - NET_BIND_SERVICE
    security_opt:
      - no-new-privileges=true
    read_only: true
    tmpfs:
      - /tmp:size=16m,mode=1777
    devices:
      - /dev/net/tun:/dev/net/tun
    volumes:
      - /volume1/docker/meshctl:/data
    restart: unless-stopped
    mem_limit: 256m
    pids_limit: 256
    healthcheck:
      test:
        - CMD
        - /app/meshctl
        - healthcheck
        - agent
        - --data
        - /data
      interval: 30s
      timeout: 5s
      retries: 3
      start_period: 10s
```

The kernel's WireGuard, where the host has the module (`--wg-backend kernel`):

```yaml
# meshctl agent for the spoke "nas", as a container (spec S12; amendments A13, A19).
# The kernel backend uses the host's WireGuard module (ls /sys/module/wireguard).
#
# Join once, with the node's join token (web UI, `meshctl node add` or `node token`):
#   MESHCTL_TOKEN=<token> docker compose run --rm -e MESHCTL_TOKEN meshctl-agent \
#     agent join --data /data --server https://mesh.example.com --name nas --no-service
# then start it: docker compose up -d
#
# Posture (docs/security.md): not privileged; every capability dropped except
# NET_ADMIN (wg0, its routes) and NET_BIND_SERVICE (DNS on :53); no privilege
# escalation; read-only root. It runs as uid 0 because capabilities do not
# survive no-new-privileges for another user, and on the host network because
# wg0 belongs to the host.
name: meshctl-agent
services:
  meshctl-agent:
    image: ghcr.io/hafio/vpn-mesh-utility:1.0.0
    command:
      - agent
      - run
      - --data
      - /data
      - --wg-backend
      - kernel
    user: "0:0"
    privileged: false
    network_mode: host
    cap_drop:
      - ALL
    cap_add:
      - NET_ADMIN
      - NET_BIND_SERVICE
    security_opt:
      - no-new-privileges=true
    read_only: true
    tmpfs:
      - /tmp:size=16m,mode=1777
    volumes:
      - /volume1/docker/meshctl:/data
    restart: unless-stopped
    mem_limit: 256m
    pids_limit: 256
    healthcheck:
      test:
        - CMD
        - /app/meshctl
        - healthcheck
        - agent
        - --data
        - /data
      interval: 30s
      timeout: 5s
      retries: 3
      start_period: 10s
```

## Container mode

A node added with `--mode container`: only the containers on its `meshctl-mesh` network are on the
mesh, and the agent's container is their gateway. Each file's header has the commands that create
that network once, the join, and how a service attaches.

### Container mode: a client on Docker

```yaml
# meshctl agent for "box" in container mode (amendment A37), a client: it dials the servers' public endpoints.
# Only containers on the meshctl-mesh network are on the mesh; the host is not.
# WireGuard: the host's module when it is loaded, else wireguard-go on /dev/net/tun.
#
# 1. Create the mesh network (docker):
#   docker network create --internal --subnet 10.77.0.0/16 --ip-range 10.77.128.128/25 --gateway 10.77.128.1 \
#     -o com.docker.network.bridge.name=meshctl0 meshctl-mesh
# 2. Join once, with the node's join token (web UI, `meshctl node add` or `node token`):
#   MESHCTL_TOKEN=<token> docker compose run --rm -e MESHCTL_TOKEN meshctl-agent \
#     agent join --data /data --server https://mesh.example.com --name box --no-service
# 3. Start it: docker compose up -d
#
# Put a service on the mesh: attach the network and use the agent as its resolver
# (ndots:1 lets a short name try the zone; Docker writes ndots:0):
#   services:
#     web:
#       networks: [default, meshctl-mesh]
#       dns: [10.77.128.2]
#       dns_search: [mesh]
#       dns_opt: ["ndots:1"]
#   networks:
#     meshctl-mesh:
#       external: true
#
# Names: <service>.mesh (every node's, round robin) and <service>.<node>.mesh,
# from compose services, container names, aliases and hostnames, which the agent
# reads from the runtime's socket (GET only). The socket gives whoever holds it
# the runtime; `meshctl node compose --no-names` leaves it out.
#
# Posture (docs/security.md): not privileged; every capability dropped except
# NET_ADMIN and NET_BIND_SERVICE, which act in the agent's own network namespace
# only; no privilege escalation; read-only root. It runs as uid 0 because
# capabilities do not survive no-new-privileges for another user.
name: meshctl-agent
services:
  meshctl-agent:
    image: ghcr.io/hafio/vpn-mesh-utility:1.0.0
    command:
      - agent
      - run
      - --data
      - /data
      - --wg-backend
      - auto
    user: "0:0"
    privileged: false
    cap_drop:
      - ALL
    cap_add:
      - NET_ADMIN
      - NET_BIND_SERVICE
    security_opt:
      - no-new-privileges=true
    read_only: true
    tmpfs:
      - /tmp:size=16m,mode=1777
    devices:
      - /dev/net/tun:/dev/net/tun
    sysctls:
      net.ipv4.conf.all.proxy_arp: "1"
      net.ipv4.ip_forward: "1"
    volumes:
      - meshctl-data:/data
      - /var/run/docker.sock:/var/run/docker.sock:ro
    networks:
      - egress
      - meshctl-mesh
    restart: unless-stopped
    mem_limit: 256m
    pids_limit: 256
    healthcheck:
      test:
        - CMD
        - /app/meshctl
        - healthcheck
        - agent
        - --data
        - /data
      interval: 30s
      timeout: 5s
      retries: 3
      start_period: 10s
networks:
  egress: {}
  meshctl-mesh:
    external: true
volumes:
  meshctl-data: {}
```

### Container mode: a server on Docker

A server publishes its WireGuard port on the host, which its endpoint names (`node add hub --role server --mode container --endpoint <ip>:51820`):

```yaml
# meshctl agent for "hub" in container mode (amendment A37), a server: the host publishes 51820/udp, which the node's endpoint names.
# Only containers on the meshctl-mesh network are on the mesh; the host is not.
# WireGuard: the host's module when it is loaded, else wireguard-go on /dev/net/tun.
#
# 1. Create the mesh network (docker):
#   docker network create --internal --subnet 10.77.0.0/16 --ip-range 10.77.129.128/25 --gateway 10.77.129.1 \
#     -o com.docker.network.bridge.name=meshctl0 meshctl-mesh
# 2. Join once, with the node's join token (web UI, `meshctl node add` or `node token`):
#   MESHCTL_TOKEN=<token> docker compose run --rm -e MESHCTL_TOKEN meshctl-agent \
#     agent join --data /data --server https://mesh.example.com --name hub --no-service
# 3. Start it: docker compose up -d
#
# Put a service on the mesh: attach the network and use the agent as its resolver
# (ndots:1 lets a short name try the zone; Docker writes ndots:0):
#   services:
#     web:
#       networks: [default, meshctl-mesh]
#       dns: [10.77.129.2]
#       dns_search: [mesh]
#       dns_opt: ["ndots:1"]
#   networks:
#     meshctl-mesh:
#       external: true
#
# Names: <service>.mesh (every node's, round robin) and <service>.<node>.mesh,
# from compose services, container names, aliases and hostnames, which the agent
# reads from the runtime's socket (GET only). The socket gives whoever holds it
# the runtime; `meshctl node compose --no-names` leaves it out.
#
# Posture (docs/security.md): not privileged; every capability dropped except
# NET_ADMIN and NET_BIND_SERVICE, which act in the agent's own network namespace
# only; no privilege escalation; read-only root. It runs as uid 0 because
# capabilities do not survive no-new-privileges for another user.
name: meshctl-agent
services:
  meshctl-agent:
    image: ghcr.io/hafio/vpn-mesh-utility:1.0.0
    command:
      - agent
      - run
      - --data
      - /data
      - --wg-backend
      - auto
    user: "0:0"
    privileged: false
    cap_drop:
      - ALL
    cap_add:
      - NET_ADMIN
      - NET_BIND_SERVICE
    security_opt:
      - no-new-privileges=true
    read_only: true
    tmpfs:
      - /tmp:size=16m,mode=1777
    devices:
      - /dev/net/tun:/dev/net/tun
    sysctls:
      net.ipv4.conf.all.proxy_arp: "1"
      net.ipv4.ip_forward: "1"
    volumes:
      - meshctl-data:/data
      - /var/run/docker.sock:/var/run/docker.sock:ro
    ports:
      - 51820:51820/udp
    networks:
      - egress
      - meshctl-mesh
    restart: unless-stopped
    mem_limit: 256m
    pids_limit: 256
    healthcheck:
      test:
        - CMD
        - /app/meshctl
        - healthcheck
        - agent
        - --data
        - /data
      interval: 30s
      timeout: 5s
      retries: 3
      start_period: 10s
networks:
  egress: {}
  meshctl-mesh:
    external: true
volumes:
  meshctl-data: {}
```

### Container mode: Podman

Rootful or rootless Podman (`--runtime podman`); a rootless one can only be a client:

```yaml
# meshctl agent for "pbox" in container mode (amendment A37), a client: it dials the servers' public endpoints.
# Only containers on the meshctl-mesh network are on the mesh; the host is not.
# WireGuard: the host's module when it is loaded, else wireguard-go on /dev/net/tun.
#
# 1. Create the mesh network (podman):
#   podman network create --internal --disable-dns --subnet 10.77.0.0/16 --ip-range 10.77.130.128/25 --gateway 10.77.130.1 \
#     --interface-name meshctl0 meshctl-mesh
# 2. Join once, with the node's join token (web UI, `meshctl node add` or `node token`):
#   MESHCTL_TOKEN=<token> podman compose run --rm -e MESHCTL_TOKEN meshctl-agent \
#     agent join --data /data --server https://mesh.example.com --name pbox --no-service
# 3. Start it: podman compose up -d
#
# Put a service on the mesh: attach the network and use the agent as its resolver
# (ndots:1 lets a short name try the zone; Docker writes ndots:0):
#   services:
#     web:
#       networks: [default, meshctl-mesh]
#       dns: [10.77.130.2]
#       dns_search: [mesh]
#       dns_opt: ["ndots:1"]
#   networks:
#     meshctl-mesh:
#       external: true
#
# Names: <service>.mesh (every node's, round robin) and <service>.<node>.mesh,
# from compose services, container names, aliases and hostnames, which the agent
# reads from the runtime's socket (GET only). The socket gives whoever holds it
# the runtime; `meshctl node compose --no-names` leaves it out.
# Podman: systemctl enable --now podman.socket; rootless, mount
# /run/user/<uid>/podman/podman.sock instead (systemctl --user enable --now podman.socket).
#
# Posture (docs/security.md): not privileged; every capability dropped except
# NET_ADMIN and NET_BIND_SERVICE, which act in the agent's own network namespace
# only; no privilege escalation; read-only root. It runs as uid 0 because
# capabilities do not survive no-new-privileges for another user.
name: meshctl-agent
services:
  meshctl-agent:
    image: ghcr.io/hafio/vpn-mesh-utility:1.0.0
    command:
      - agent
      - run
      - --data
      - /data
      - --wg-backend
      - auto
    user: "0:0"
    privileged: false
    cap_drop:
      - ALL
    cap_add:
      - NET_ADMIN
      - NET_BIND_SERVICE
    security_opt:
      - no-new-privileges=true
    read_only: true
    tmpfs:
      - /tmp:size=16m,mode=1777
    devices:
      - /dev/net/tun:/dev/net/tun
    sysctls:
      net.ipv4.conf.all.proxy_arp: "1"
      net.ipv4.ip_forward: "1"
    volumes:
      - meshctl-data:/data
      - /run/podman/podman.sock:/var/run/docker.sock:ro
    networks:
      - egress
      - meshctl-mesh
    restart: unless-stopped
    mem_limit: 256m
    pids_limit: 256
    healthcheck:
      test:
        - CMD
        - /app/meshctl
        - healthcheck
        - agent
        - --data
        - /data
      interval: 30s
      timeout: 5s
      retries: 3
      start_period: 10s
networks:
  egress: {}
  meshctl-mesh:
    external: true
volumes:
  meshctl-data: {}
```

### Container mode: a Docker Swarm node

The agent itself is a plain compose container on each node (a Swarm service cannot have
/dev/net/tun or the capabilities it needs); Swarm services attach to a swarm-scoped network
built from each node's config-only one, and resolve through the servers (`--runtime swarm`):

```yaml
# meshctl agent for "sbox" in container mode (amendment A37), a client: it dials the servers' public endpoints.
# Only containers on the meshctl-mesh network are on the mesh; the host is not.
# WireGuard: the host's module when it is loaded, else wireguard-go on /dev/net/tun.
#
# 1. Create the mesh network (swarm):
# on this node:
#   docker network create --config-only --subnet 10.77.0.0/16 --ip-range 10.77.131.128/25 --gateway 10.77.131.1 meshctl-mesh-config
# once, on a manager (every node with mesh services has its own -config):
#   docker network create -d bridge --scope swarm --attachable --config-from meshctl-mesh-config meshctl-mesh
# 2. Join once, with the node's join token (web UI, `meshctl node add` or `node token`):
#   MESHCTL_TOKEN=<token> docker compose run --rm -e MESHCTL_TOKEN meshctl-agent \
#     agent join --data /data --server https://mesh.example.com --name sbox --no-service
# 3. Start it: docker compose up -d
#
# Put a service on the mesh: attach the network and use the agent as its resolver
# (ndots:1 lets a short name try the zone; Docker writes ndots:0):
#   services:
#     web:
#       networks: [default, meshctl-mesh]
#       dns: [10.77.0.1, 10.77.0.2, 10.77.0.3]
#       dns_search: [mesh]
#       dns_opt: ["ndots:1"]
#   networks:
#     meshctl-mesh:
#       external: true
# (A Swarm service's tasks run on several nodes, so its dns is the servers: every
# node's resolver answers every node's containers.)
#
# Names: <service>.mesh (every node's, round robin) and <service>.<node>.mesh,
# from compose services, container names, aliases and hostnames, which the agent
# reads from the runtime's socket (GET only). The socket gives whoever holds it
# the runtime; `meshctl node compose --no-names` leaves it out.
#
# Posture (docs/security.md): not privileged; every capability dropped except
# NET_ADMIN and NET_BIND_SERVICE, which act in the agent's own network namespace
# only; no privilege escalation; read-only root. It runs as uid 0 because
# capabilities do not survive no-new-privileges for another user.
name: meshctl-agent
services:
  meshctl-agent:
    image: ghcr.io/hafio/vpn-mesh-utility:1.0.0
    command:
      - agent
      - run
      - --data
      - /data
      - --wg-backend
      - auto
    user: "0:0"
    privileged: false
    cap_drop:
      - ALL
    cap_add:
      - NET_ADMIN
      - NET_BIND_SERVICE
    security_opt:
      - no-new-privileges=true
    read_only: true
    tmpfs:
      - /tmp:size=16m,mode=1777
    devices:
      - /dev/net/tun:/dev/net/tun
    sysctls:
      net.ipv4.conf.all.proxy_arp: "1"
      net.ipv4.ip_forward: "1"
    volumes:
      - meshctl-data:/data
      - /var/run/docker.sock:/var/run/docker.sock:ro
    networks:
      - egress
      - meshctl-mesh
    restart: unless-stopped
    mem_limit: 256m
    pids_limit: 256
    healthcheck:
      test:
        - CMD
        - /app/meshctl
        - healthcheck
        - agent
        - --data
        - /data
      interval: 30s
      timeout: 5s
      retries: 3
      start_period: 10s
networks:
  egress: {}
  meshctl-mesh:
    external: true
volumes:
  meshctl-data: {}
```

## A host-mode node's containers

A host-mode node with `--containers` keeps the host on the mesh and gives its containers mesh
addresses too: the host's own agent routes them on the bridge `meshctl0`. Create the network once
on the host, then attach services (`meshctl node network <name> --runtime ...`). For a server at
10.77.132.0/24:

Docker:

```sh
  docker network create --subnet 10.77.0.0/16 --ip-range 10.77.132.128/25 --gateway 10.77.132.1 \
    -o com.docker.network.bridge.name=meshctl0 -o com.docker.network.bridge.inhibit_ipv4=true \
    -o com.docker.network.bridge.enable_ip_masquerade=false -o com.docker.network.bridge.gateway_mode_ipv4=nat-unprotected \
    meshctl-mesh
```

```yaml
  services:
    web:
      networks: [default, meshctl-mesh]
      dns: [10.77.132.1]
      dns_search: [mesh]
      dns_opt: ["ndots:1"]
  networks:
    meshctl-mesh:
      external: true
```

Swarm services:

```sh
on this node:
  docker network create --config-only --subnet 10.77.0.0/16 --ip-range 10.77.132.128/25 --gateway 10.77.132.1 \
    -o com.docker.network.bridge.name=meshctl0 -o com.docker.network.bridge.inhibit_ipv4=true \
    -o com.docker.network.bridge.enable_ip_masquerade=false -o com.docker.network.bridge.gateway_mode_ipv4=nat-unprotected \
    meshctl-mesh-config
once, on a manager (every node with mesh services has its own -config):
  docker network create -d bridge --scope swarm --attachable --config-from meshctl-mesh-config meshctl-mesh
```

```yaml
  services:
    web:
      networks: [default, meshctl-mesh]
      dns: [10.77.0.1, 10.77.0.2, 10.77.0.3]
      dns_search: [mesh]
      dns_opt: ["ndots:1"]
  networks:
    meshctl-mesh:
      external: true
```
