# Containers

meshctl ships one image, `ghcr.io/hafio/vpn-mesh-utility:<release>` (the tag has no `v`), and
writes no container deployment of its own (amendment A48): the samples below are starting points
for your own compose files, `podman run` lines or Kubernetes manifests. Each one states its
posture -- what it is allowed beyond a locked-down container, and nothing else -- and
`TestContainerSamples` (deploy/samples_test.go) parses every sample here and fails on a widening
it does not expect; [security.md](security.md) explains each one.

## One image, two modes

The image's command is `meshctl start`, which runs what the container's environment says
(amendments A47, A48):

| `MESHCTL_MODE` | Runs | Needs |
| --- | --- | --- |
| `control` | the control plane: the database, the web UI and the agent API on `MESHCTL_PORT` (8080) | a volume at `/data`, a reverse proxy for https; no WireGuard, no UDP, no capability |
| unset, or `node` | a node's agent | `NET_ADMIN`, its own WireGuard, a volume at `/data` |

Only the control plane sets the mode. Whether a node is a **server node** (a hub with a public
UDP endpoint) or a **client node** (a spoke that dials out) is its role on the control plane
(`meshctl node add <name> --role server|spoke`), never the container's setting. A node's
container takes:

| Variable | Value |
| --- | --- |
| `MESHCTL_URL` | the control plane's public origin, e.g. `https://mesh.example.com` (the same value the control plane's own `MESHCTL_URL` has) |
| `MESHCTL_NAME` | the node's name; a token issued for another node is refused |
| `MESHCTL_TOKEN` | the node's join token, from the web UI or `meshctl node token <name>` |
| `MESHCTL_TOKEN_FILE` | in place of `MESHCTL_TOKEN`: an absolute path to a file holding it (a compose or Kubernetes secret) |
| `MESHCTL_WG_BACKEND` | `auto` (default: the host's WireGuard module, else wireguard-go on `/dev/net/tun`), `kernel` or `userspace` |
| `MESHCTL_RUNTIME_SOCKET` | `auto` (default), `off`, or the runtime's socket inside the container: where the agent reads its containers' names |

[commands.md](commands.md#meshctl-start) lists every variable. The agent joins on its first start
and keeps only the token's fingerprint, so a restart does not join again; a new token joins again
([operations.md](operations.md) has the table). Never write the token into a file you commit:
compose takes `${MESHCTL_TOKEN}` from the shell or a `.env` file beside the compose file, `podman
run -e MESHCTL_TOKEN` (no `=`) takes it from the shell, and a secret mounted as a file keeps it out
of the container's environment altogether.

A node's network values -- its container subnet, the range its runtime hands out, the bridge, the
agent's address, the resolvers -- come from the control plane: `meshctl node show <name>` and the
node's page in the web UI list them under "containers on the mesh". The samples use these:

| Value | `node show` | Example |
| --- | --- | --- |
| the mesh on the link | subnet | `10.77.0.0/16` (IPv6 `fd10:77::/64`) |
| what the runtime hands out | ip range | `10.77.128.128/25` (`fd10:77::8000:8000/113`) |
| the runtime's side of the bridge | gateway | `10.77.128.1` (`fd10:77::8000:1`) |
| the agent, in container mode | agent | `10.77.128.2` (`fd10:77::8000:2`) |
| the network and the bridge | network | `meshctl-mesh`, `meshctl0` (a mesh in slot s: `meshctl-mesh<s>`, `meshctl<s>`) |

## The control plane

On Dokploy, [`deploy/compose.control.yml`](../deploy/compose.control.yml) is a compose app and
[`deploy/stack.control.yml`](../deploy/stack.control.yml) a Swarm stack (their headers list the
variables; `deploy_test.go` pins both). Anywhere else, behind your own reverse proxy:

```yaml
services:
  meshctl:
    image: ghcr.io/hafio/vpn-mesh-utility:1.0.0
    environment:
      MESHCTL_MODE: control
      MESHCTL_URL: https://mesh.example.com
      MESHCTL_TRUSTED_PROXIES: 172.18.0.0/16
      TZ: UTC
    user: "65532:65532"
    privileged: false
    cap_drop: [ALL]
    security_opt: ["no-new-privileges=true"]
    read_only: true
    tmpfs: ["/tmp:size=16m,mode=1777"]
    volumes: ["meshctl-data:/data"]
    ports: ["127.0.0.1:8080:8080"]
    restart: unless-stopped
    mem_limit: 256m
    pids_limit: 256
volumes:
  meshctl-data: {}
```

The proxy terminates https and forwards to port 8080; `MESHCTL_TRUSTED_PROXIES` is the CIDR it
connects from, so sign-in lockouts and join limits see real client addresses. Exactly one control
plane runs per data volume (it holds a lock on `/data`). Manage it with `docker exec <container>
/app/meshctl node ...`.

With Podman:

```sh
podman volume create meshctl-data
podman run -d --name meshctl --user=65532:65532 --cap-drop=ALL --security-opt=no-new-privileges \
  --read-only --tmpfs=/tmp:size=16m,mode=1777 -v meshctl-data:/data -p 127.0.0.1:8080:8080 \
  -e MESHCTL_MODE=control -e MESHCTL_URL=https://mesh.example.com \
  ghcr.io/hafio/vpn-mesh-utility:1.0.0
```

On Kubernetes, one replica on a persistent volume, behind your ingress:

```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: meshctl-data
spec:
  accessModes: [ReadWriteOnce]
  resources:
    requests:
      storage: 1Gi
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: meshctl-control
spec:
  replicas: 1
  strategy:
    type: Recreate
  selector:
    matchLabels:
      app: meshctl-control
  template:
    metadata:
      labels:
        app: meshctl-control
    spec:
      securityContext:
        runAsUser: 65532
        runAsGroup: 65532
        fsGroup: 65532
        runAsNonRoot: true
        seccompProfile:
          type: RuntimeDefault
      containers:
        - name: meshctl
          image: ghcr.io/hafio/vpn-mesh-utility:1.0.0
          env:
            - name: MESHCTL_MODE
              value: control
            - name: MESHCTL_URL
              value: https://mesh.example.com
          ports:
            - containerPort: 8080
          readinessProbe:
            httpGet:
              path: /healthz
              port: 8080
          securityContext:
            privileged: false
            allowPrivilegeEscalation: false
            readOnlyRootFilesystem: true
            capabilities:
              drop: [ALL]
          volumeMounts:
            - name: data
              mountPath: /data
            - name: tmp
              mountPath: /tmp
      volumes:
        - name: data
          persistentVolumeClaim:
            claimName: meshctl-data
        - name: tmp
          emptyDir:
            sizeLimit: 16Mi
---
apiVersion: v1
kind: Service
metadata:
  name: meshctl-control
spec:
  selector:
    app: meshctl-control
  ports:
    - port: 8080
      targetPort: 8080
```

`Recreate` stops the old pod before the new one starts: the database has one writer. Manage it
with `kubectl exec deploy/meshctl-control -- /app/meshctl node list`.

## A client node on the host network

A NAS, or any Linux host with a container runtime, can run its agent in a container on the host's
network: the whole host joins the mesh, as with `install.sh`. Add it as a spoke (`node add nas
--role spoke --always-on`), then:

```yaml
services:
  meshctl:
    image: ghcr.io/hafio/vpn-mesh-utility:1.0.0
    environment:
      MESHCTL_URL: https://mesh.example.com
      MESHCTL_NAME: nas
      MESHCTL_TOKEN: ${MESHCTL_TOKEN:-}
    user: "0:0"
    privileged: false
    network_mode: host
    cap_drop: [ALL]
    cap_add: [NET_ADMIN, NET_BIND_SERVICE]
    security_opt: ["no-new-privileges=true"]
    read_only: true
    tmpfs: ["/tmp:size=16m,mode=1777"]
    devices: ["/dev/net/tun:/dev/net/tun"]
    volumes: ["/volume1/docker/meshctl:/data"]
    restart: unless-stopped
    mem_limit: 256m
    pids_limit: 256
```

```sh
MESHCTL_TOKEN=<token> docker compose up -d
docker compose exec meshctl /app/meshctl status --data /data
```

It runs as uid 0 because capabilities do not survive `no-new-privileges` for another user, with
`NET_ADMIN` for its interface and routes and `NET_BIND_SERVICE` for DNS on port 53. The tun device
is for wireguard-go; where the host has the WireGuard module (`ls /sys/module/wireguard`), set
`MESHCTL_WG_BACKEND: kernel` and leave `devices` out -- a kernel interface also stays up while the
container restarts. The data directory holds the node's keys: keep it on the host.

With Podman:

```sh
podman run -d --name meshctl --network=host --user=0:0 --cap-drop=ALL --cap-add=NET_ADMIN \
  --cap-add=NET_BIND_SERVICE --security-opt=no-new-privileges --read-only \
  --tmpfs=/tmp:size=16m,mode=1777 --device=/dev/net/tun -v /var/lib/meshctl:/data \
  -e MESHCTL_URL=https://mesh.example.com -e MESHCTL_NAME=nas -e MESHCTL_TOKEN \
  ghcr.io/hafio/vpn-mesh-utility:1.0.0
```

On Kubernetes, a pod on one cluster node's host network joins that cluster node to the mesh
(one Deployment per cluster node and mesh node; never a DaemonSet, whose pods would all be the
same node). The kernel backend keeps the device out; the token comes from a secret as a file:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: meshctl-k8s1
spec:
  replicas: 1
  strategy:
    type: Recreate
  selector:
    matchLabels:
      app: meshctl-k8s1
  template:
    metadata:
      labels:
        app: meshctl-k8s1
    spec:
      hostNetwork: true
      dnsPolicy: ClusterFirstWithHostNet
      nodeSelector:
        kubernetes.io/hostname: k8s1
      containers:
        - name: meshctl
          image: ghcr.io/hafio/vpn-mesh-utility:1.0.0
          env:
            - name: MESHCTL_URL
              value: https://mesh.example.com
            - name: MESHCTL_NAME
              value: k8s1
            - name: MESHCTL_TOKEN_FILE
              value: /run/secrets/meshctl/token
            - name: MESHCTL_WG_BACKEND
              value: kernel
            - name: MESHCTL_RUNTIME_SOCKET
              value: "off"
          securityContext:
            runAsUser: 0
            privileged: false
            allowPrivilegeEscalation: false
            readOnlyRootFilesystem: true
            capabilities:
              drop: [ALL]
              add: [NET_ADMIN, NET_BIND_SERVICE]
          volumeMounts:
            - name: data
              mountPath: /data
            - name: tmp
              mountPath: /tmp
            - name: token
              mountPath: /run/secrets/meshctl
              readOnly: true
      volumes:
        - name: data
          hostPath:
            path: /var/lib/meshctl
            type: DirectoryOrCreate
        - name: tmp
          emptyDir:
            sizeLimit: 16Mi
        - name: token
          secret:
            secretName: meshctl-token-k8s1
            optional: true
```

`kubectl create secret generic meshctl-token-k8s1 --from-file=token=<file>` holds the token; once
the node has joined, the secret can go (`optional: true`). These manifests follow the posture of
the compose files; they have not been through the project's live tests.

**Server nodes do not run this way.** A server node on the host network needs firewalld on the
host for its public port, the relay and the exit (`firewall-cmd`, which the image does not carry):
install it with `install.sh` on the host, or run it in container mode below, which publishes its
UDP port and needs no host firewall changes.

## Container mode

A node added with `--mode container` runs WireGuard in its agent's own network namespace. Only the
containers attached to the node's mesh network are on the mesh, and the agent is their gateway;
the host, its addresses and its LAN are not (amendment A37). A container-mode **client node**
only dials out; a container-mode **server node** publishes its WireGuard port on the host, which
its endpoint names:

```sh
meshctl node add box --role spoke --mode container
meshctl node add hub4 --role server --mode container --endpoint 198.51.100.7:51820
meshctl node show box
```

### 1. The mesh network, once per host

With Docker (the values from `node show`):

```sh
docker network create --internal --subnet 10.77.0.0/16 --ip-range 10.77.128.128/25 \
  --gateway 10.77.128.1 -o com.docker.network.bridge.name=meshctl0 meshctl-mesh
```

With rootful or rootless Podman (`--disable-dns`: Podman's resolver on an internal network
answers every outside name with NXDOMAIN):

```sh
podman network create --internal --disable-dns --subnet 10.77.0.0/16 --ip-range 10.77.128.128/25 \
  --gateway 10.77.128.1 --interface-name meshctl0 meshctl-mesh
```

For Swarm services, a config-only network on each node with mesh services, and one swarm-scoped
network from it, once, on a manager:

```sh
docker network create --config-only --subnet 10.77.0.0/16 --ip-range 10.77.128.128/25 \
  --gateway 10.77.128.1 meshctl-mesh-config
docker network create -d bridge --scope swarm --attachable --config-from meshctl-mesh-config meshctl-mesh
```

With the containers' IPv6 on the mesh (`--containers6`), add the IPv6 values after the IPv4 ones:
`--ipv6 --subnet fd10:77::/64 --ip-range fd10:77::8000:8000/113 --gateway fd10:77::8000:1`.

### 2. The agent

```yaml
services:
  meshctl:
    image: ghcr.io/hafio/vpn-mesh-utility:1.0.0
    environment:
      MESHCTL_URL: https://mesh.example.com
      MESHCTL_NAME: box
      MESHCTL_TOKEN: ${MESHCTL_TOKEN:-}
    user: "0:0"
    privileged: false
    cap_drop: [ALL]
    cap_add: [NET_ADMIN, NET_BIND_SERVICE]
    security_opt: ["no-new-privileges=true"]
    read_only: true
    tmpfs: ["/tmp:size=16m,mode=1777"]
    devices: ["/dev/net/tun:/dev/net/tun"]
    sysctls:
      net.ipv4.ip_forward: "1"
      net.ipv4.conf.all.proxy_arp: "1"
    volumes:
      - meshctl-data:/data
      - /var/run/docker.sock:/var/run/docker.sock:ro
    networks: [egress, meshctl-mesh]
    restart: unless-stopped
    mem_limit: 256m
    pids_limit: 256
networks:
  egress: {}
  meshctl-mesh:
    external: true
volumes:
  meshctl-data: {}
```

`egress` is the agent's way to the internet (the servers' endpoints); the mesh network is
internal. The agent claims the agent address (`.2`) on the mesh network itself. The sysctls act in
the agent's own namespace only: it forwards between WireGuard and the containers and answers ARP
for every mesh address it routes. With the containers' IPv6, add `net.ipv6.conf.all.disable_ipv6:
"0"`, `net.ipv6.conf.all.forwarding: "1"` and `net.ipv6.conf.all.proxy_ndp: "1"`. A **server
node** adds the port its endpoint names, to the agent's listen port:

```yaml
services:
  meshctl:
    ports: ["51820:51820/udp"]
```

The capabilities act in the agent's namespace only; the host's network is untouched. Swarm
services attach to the swarm-scoped network, but the agent itself is this plain compose service on
each node (a Swarm service cannot have the tun device or the capabilities).

With Podman (rootful; rootless for a client node, with its own socket below):

```sh
podman network create meshctl-egress
podman run -d --name meshctl --user=0:0 --cap-drop=ALL --cap-add=NET_ADMIN --cap-add=NET_BIND_SERVICE \
  --security-opt=no-new-privileges --read-only --tmpfs=/tmp:size=16m,mode=1777 --device=/dev/net/tun \
  --sysctl=net.ipv4.ip_forward=1 --sysctl=net.ipv4.conf.all.proxy_arp=1 \
  -v meshctl-data:/data -v /run/podman/podman.sock:/var/run/docker.sock:ro \
  --network=meshctl-egress --network=meshctl-mesh \
  -e MESHCTL_URL=https://mesh.example.com -e MESHCTL_NAME=box -e MESHCTL_TOKEN \
  ghcr.io/hafio/vpn-mesh-utility:1.0.0
```

Container mode is not for Kubernetes: a pod network is not a runtime bridge the agent can be the
gateway of. Put a cluster node on the mesh on the host network (above) instead.

### 3. A service on the mesh

Attach the network and make the agent its resolver (`ndots:1`, because Docker writes `ndots:0`,
which keeps a short name from trying the zone):

```yaml
services:
  web:
    image: nginx:1.27
    networks: [default, meshctl-mesh]
    dns: [10.77.128.2]
    dns_search: [mesh]
    dns_opt: ["ndots:1"]
networks:
  meshctl-mesh:
    external: true
```

It gets an address from the range and reaches every mesh address -- other nodes, their
containers, phones -- through the agent, with no route of its own; its internet stays on its other
network. A Swarm service, whose tasks run on several nodes, names the server nodes' mesh IPs as its
`dns` instead (`node show` lists them as swarm dns): every node's resolver answers every node's
containers.

**Rootful Podman next to Docker.** Docker loads `br_netfilter` and sets the kernel's FORWARD
policy to DROP, which drops traffic between containers on a Podman bridge. On such a host, let the
mesh bridge through both Docker's chain and firewalld:

```sh
iptables -I DOCKER-USER -i meshctl0 -j ACCEPT
iptables -I DOCKER-USER -o meshctl0 -j ACCEPT
firewall-cmd --permanent --zone=trusted --add-interface=meshctl0 && firewall-cmd --reload
```

## A host-mode node's containers

A host-mode node with `--containers` (any role) keeps the host on the mesh and gives its
containers mesh addresses too: the host's agent owns the bridge `meshctl0` and routes them
(amendment A37). Its agent runs on the host (`install.sh`); create the runtime's network on its
bridge once, with the node's values (here a server node at 10.77.129.0/24, whose gateway is the
bridge):

```sh
docker network create --subnet 10.77.0.0/16 --ip-range 10.77.129.128/25 --gateway 10.77.129.1 \
  -o com.docker.network.bridge.name=meshctl0 -o com.docker.network.bridge.inhibit_ipv4=true \
  -o com.docker.network.bridge.enable_ip_masquerade=false \
  -o com.docker.network.bridge.gateway_mode_ipv4=nat-unprotected meshctl-mesh
```

Docker takes the agent's bridge as it is (`inhibit_ipv4`), masquerades nothing and filters no
port the mesh may reach (`nat-unprotected`; the agent's rules keep the subnet to the mesh). With
the containers' IPv6, add the IPv6 values and `-o
com.docker.network.bridge.gateway_mode_ipv6=nat-unprotected`. Podman attaches unmanaged:

```sh
podman network create --disable-dns --subnet 10.77.0.0/16 --ip-range 10.77.129.128/25 \
  --gateway 10.77.129.1 --interface-name meshctl0 -o mode=unmanaged -o no_default_route=true meshctl-mesh
```

Swarm builds the swarm-scoped network from a config-only one with Docker's options, as in container
mode. A service attaches as above, with the bridge (`10.77.129.1`) as its `dns`.

## Container names

Every node resolves the containers on the mesh networks by name (amendment A38): `web.mesh`
answers every node's `web` containers, rotated, and `web.box.mesh` only box's. The agent reads the
names from the runtime's API socket, GET requests only. A host-mode agent finds Docker's or rootful
Podman's socket itself; a container-mode agent finds it where the samples mount it,
`/var/run/docker.sock` inside its container: Docker's, or rootful Podman's
`/run/podman/podman.sock` (`systemctl enable --now podman.socket`), or for rootless Podman
`/run/user/<uid>/podman/podman.sock` (`systemctl --user enable --now podman.socket`). Whoever holds
that socket holds the runtime, read-only mount or not: leave the mount out and set
`MESHCTL_RUNTIME_SOCKET: "off"` to give the node's containers no names. A runtime that labels
containers for SELinux may refuse the agent the socket; the same setting is the way out.

## Posture

| Sample | Beyond a locked-down container | Why |
| --- | --- | --- |
| control plane | nothing (uid 65532, no capability, read-only root) | it only serves https behind a proxy |
| client node on the host network | uid 0, `network_mode: host` / `hostNetwork`, `NET_ADMIN`, `NET_BIND_SERVICE`, `/dev/net/tun`, a host data directory | its interface, routes and DNS are the host's |
| container-mode agent | uid 0, `NET_ADMIN`, `NET_BIND_SERVICE`, `/dev/net/tun`, the two (five with IPv6) sysctls, the runtime's socket, a published UDP port on a server node | all in its own namespace, except the socket |

Every sample keeps `privileged: false`, drops every capability first, sets `no-new-privileges` and
a read-only root, and passes no secret on a command line.
