package deploy

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"go.yaml.in/yaml/v3"

	"github.com/hafio/vpn-mesh-utility/internal/validate"
)

// Container runtimes a container-mode node's compose file is written for
// (amendment A37). They differ only in how the mesh network is created,
// which the file's header shows.
const (
	RuntimeDocker = "docker"
	RuntimePodman = "podman"
	// RuntimeSwarm attaches Swarm services: the agent itself is a plain
	// compose container on the node, the network swarm-scoped.
	RuntimeSwarm = "swarm"
)

// The WireGuard backends of a container-mode agent: auto takes the host's
// module when it is loaded and wireguard-go on the tun device otherwise.
const (
	ContainerAuto      = "auto"
	ContainerKernel    = NASKernel
	ContainerUserspace = NASUserspace
)

// Names of the container-mode compose file. The mesh network's name is
// proto.MeshNetwork and its bridge proto.ContainerBridge; deploy does not
// import proto, so they are repeated here and pinned by a test.
const (
	ContainerService = "meshctl-agent"
	ContainerVolume  = "meshctl-data"
	MeshNetwork      = "meshctl-mesh"
	MeshBridge       = "meshctl0"
)

// ContainerParams are the values in a container-mode node's compose file.
type ContainerParams struct {
	Node    string
	Server  string
	Version string
	// Hub is a server: it publishes its WireGuard port on the host.
	Hub bool
	// ListenPort is the agent's WireGuard port inside its namespace,
	// PublishPort the host's port that the node's endpoint names (hub only).
	ListenPort  int
	PublishPort int
	// Backend is ContainerAuto (empty), ContainerKernel or
	// ContainerUserspace.
	Backend string
	// Runtime is RuntimeDocker (empty), RuntimePodman or RuntimeSwarm.
	Runtime string
	// DataDir is a host directory for /data; empty means the named volume
	// ContainerVolume.
	DataDir string
	// NoNames leaves out the runtime's API socket, which the agent reads
	// the containers' names from (amendment A38).
	NoNames bool
	// The node's container network (proto.Containers) and the mesh's DNS
	// zone, for the network commands and the example in the header.
	Network string
	Subnet  string
	Range   string
	Bridge  string
	Gateway string
	Zone    string
	// Resolvers are the servers' mesh IPs: a Swarm service's resolvers,
	// since its tasks run on several nodes and each node's gateway differs.
	Resolvers []string
	// The containers' IPv6, when it is on the mesh (M8, opt-in): the mesh's
	// IPv6 prefix, the range handed out and the bridge's address; all three
	// or none.
	Network6, Range6, Bridge6 string
}

// The compose file as a typed value, so yaml.v3 does every bit of quoting
// and no value can add a key. Field order is output order.
type containerFile struct {
	Name     string                      `yaml:"name"`
	Services map[string]containerService `yaml:"services"`
	Networks map[string]composeNetwork   `yaml:"networks"`
	Volumes  map[string]struct{}         `yaml:"volumes,omitempty"`
}

type containerService struct {
	Image       string            `yaml:"image"`
	Command     []string          `yaml:"command"`
	User        string            `yaml:"user"`
	Privileged  bool              `yaml:"privileged"`
	CapDrop     []string          `yaml:"cap_drop"`
	CapAdd      []string          `yaml:"cap_add"`
	SecurityOpt []string          `yaml:"security_opt"`
	ReadOnly    bool              `yaml:"read_only"`
	Tmpfs       []string          `yaml:"tmpfs"`
	Devices     []string          `yaml:"devices,omitempty"`
	Sysctls     map[string]string `yaml:"sysctls"`
	Volumes     []string          `yaml:"volumes"`
	Ports       []string          `yaml:"ports,omitempty"`
	Networks    []string          `yaml:"networks"`
	Restart     string            `yaml:"restart"`
	MemLimit    string            `yaml:"mem_limit"`
	PidsLimit   int               `yaml:"pids_limit"`
	Healthcheck nasHealth         `yaml:"healthcheck"`
}

type composeNetwork struct {
	External bool `yaml:"external,omitempty"`
}

// RenderContainerCompose returns the compose file that runs a
// container-mode node's agent (amendment A37): in its own network
// namespace, attached to an egress network and to the mesh network that
// the header's commands create. Every value is checked against its grammar
// first, so a bad one fails here instead of reaching the file.
func RenderContainerCompose(p ContainerParams) ([]byte, error) {
	if p.Backend == "" {
		p.Backend = ContainerAuto
	}
	if p.Runtime == "" {
		p.Runtime = RuntimeDocker
	}
	var errs []error
	keep := func(err error) {
		if err != nil {
			errs = append(errs, err)
		}
	}
	_, err := validate.Name("node", p.Node)
	keep(err)
	origin, err := validate.ServerURL("server", p.Server)
	if err != nil {
		errs = append(errs, err)
	} else if origin != p.Server {
		errs = append(errs, fmt.Errorf("server: must be in canonical form %q (got %q)", origin, p.Server))
	}
	if _, err := validate.SemVer("version", p.Version); err != nil {
		errs = append(errs, fmt.Errorf("%w: only releases have an image; this is a development build", err))
	}
	_, err = validate.Enum("backend", p.Backend, ContainerAuto, ContainerKernel, ContainerUserspace)
	keep(err)
	_, err = validate.Enum("runtime", p.Runtime, RuntimeDocker, RuntimePodman, RuntimeSwarm)
	keep(err)
	if p.DataDir != "" {
		_, err = validate.POSIXPath("data", p.DataDir)
		keep(err)
	}
	_, err = validate.Port("listen_port", p.ListenPort)
	keep(err)
	if p.Hub {
		_, err = validate.Port("publish_port", p.PublishPort)
		keep(err)
	}
	for _, f := range [][2]string{{"network", p.Network}, {"subnet", p.Subnet}, {"range", p.Range}} {
		_, err = validate.Prefix4(f[0], f[1])
		keep(err)
	}
	for _, f := range [][2]string{{"bridge", p.Bridge}, {"gateway", p.Gateway}} {
		_, err = validate.IPv4(f[0], f[1])
		keep(err)
	}
	_, err = validate.Zone("zone", p.Zone)
	keep(err)
	for i, r := range p.Resolvers {
		_, err = validate.IPv4(fmt.Sprintf("resolvers[%d]", i), r)
		keep(err)
	}
	keep(checkIPv6(p.Network6, p.Range6, p.Bridge6))
	if err := errors.Join(errs...); err != nil {
		return nil, fmt.Errorf("render the compose file: %w", err)
	}
	return renderContainer(p)
}

// checkIPv6 checks the IPv6 side of a container network (M8): all of it,
// or none.
func checkIPv6(network6, range6, bridge6 string) error {
	if network6 == "" && range6 == "" && bridge6 == "" {
		return nil
	}
	var errs []error
	for _, f := range [][2]string{{"network6", network6}, {"range6", range6}} {
		if _, err := validate.Prefix6(f[0], f[1]); err != nil {
			errs = append(errs, err)
		}
	}
	if _, err := validate.IPv6("bridge6", bridge6); err != nil {
		errs = append(errs, err)
	}
	return errors.Join(errs...)
}

// RuntimeSocket is where the container-mode agent finds the runtime's API
// socket: Docker's path, whichever runtime's socket is mounted there.
const RuntimeSocket = "/var/run/docker.sock"

// hostSocket is the runtime's API socket on the host: rootful Podman's for
// podman (a rootless one is in the user's runtime directory, which the
// header says), Docker's otherwise.
func hostSocket(runtime string) string {
	if runtime == RuntimePodman {
		return "/run/podman/podman.sock"
	}
	return RuntimeSocket
}

// renderContainer writes p without checking it. It still holds the file's
// structure against any value: yaml.v3 quotes what it must, every '$' is
// doubled so compose does not interpolate it, and the comment lines cannot
// be broken.
func renderContainer(p ContainerParams) ([]byte, error) {
	lit := func(s string) string { return strings.ReplaceAll(s, "$", "$$") }
	data := ContainerVolume
	if p.DataDir != "" {
		data = lit(p.DataDir)
	}
	command := []string{"agent", "run", "--data", "/data", "--wg-backend", lit(p.Backend)}
	volumes := []string{data + ":/data"}
	if p.NoNames {
		command = append(command, "--runtime-socket", "off")
	} else {
		// Read-only says what the agent does with it (GET only); it does
		// not limit the API -- docs/security.md.
		volumes = append(volumes, hostSocket(p.Runtime)+":"+RuntimeSocket+":ro")
	}
	// The gateway forwards between wg0 and the containers and answers ARP
	// for every mesh address it routes -- and, with IPv6 (M8), neighbour
	// discovery; /proc/sys is read-only in a container, so the runtime sets
	// them, in the agent's namespace only.
	svc := containerService{
		Image:       lit(Image + ":" + strings.TrimPrefix(p.Version, "v")),
		Command:     command,
		User:        "0:0",
		Privileged:  false,
		CapDrop:     []string{"ALL"},
		CapAdd:      []string{"NET_ADMIN", "NET_BIND_SERVICE"},
		SecurityOpt: []string{"no-new-privileges=true"},
		ReadOnly:    true,
		Tmpfs:       []string{"/tmp:size=16m,mode=1777"},
		Sysctls:     map[string]string{"net.ipv4.ip_forward": "1", "net.ipv4.conf.all.proxy_arp": "1"},
		Volumes:     volumes,
		Networks:    []string{"egress", MeshNetwork},
		Restart:     "unless-stopped",
		MemLimit:    "256m",
		PidsLimit:   256,
		Healthcheck: nasHealth{
			Test:        []string{"CMD", "/app/meshctl", "healthcheck", "agent", "--data", "/data"},
			Interval:    "30s",
			Timeout:     "5s",
			Retries:     3,
			StartPeriod: "10s",
		},
	}
	if p.Network6 != "" {
		svc.Sysctls["net.ipv6.conf.all.disable_ipv6"] = "0"
		svc.Sysctls["net.ipv6.conf.all.forwarding"] = "1"
		svc.Sysctls["net.ipv6.conf.all.proxy_ndp"] = "1"
	}
	if p.Backend != ContainerKernel {
		svc.Devices = []string{"/dev/net/tun:/dev/net/tun"}
	}
	if p.Hub {
		svc.Ports = []string{strconv.Itoa(p.PublishPort) + ":" + strconv.Itoa(p.ListenPort) + "/udp"}
	}
	f := containerFile{
		Name:     ContainerService,
		Services: map[string]containerService{ContainerService: svc},
		Networks: map[string]composeNetwork{"egress": {}, MeshNetwork: {External: true}},
	}
	if p.DataDir == "" {
		f.Volumes = map[string]struct{}{ContainerVolume: {}}
	}
	var b bytes.Buffer
	for _, line := range containerHeader(p) {
		b.WriteString(strings.TrimRight("# "+comment(line), " ") + "\n")
	}
	enc := yaml.NewEncoder(&b)
	enc.SetIndent(2)
	if err := enc.Encode(f); err != nil {
		return nil, fmt.Errorf("render the compose file: %w", err)
	}
	if err := enc.Close(); err != nil {
		return nil, fmt.Errorf("render the compose file: %w", err)
	}
	return b.Bytes(), nil
}

// NetworkParams are the values of a node's mesh network commands: the
// runtime, whether the node is in host mode, and its container network
// (proto.Containers' network, range, bridge and gateway) and the zone.
type NetworkParams struct {
	Runtime string
	Host    bool
	Network string
	Range   string
	Bridge  string
	Gateway string
	Zone    string
	// Resolvers are the servers' mesh IPs, for a Swarm service.
	Resolvers []string
	// The containers' IPv6 (M8, opt-in), as ContainerParams has it.
	Network6, Range6, Bridge6 string
}

// NetworkSetup is what a node's operator runs and writes to put containers
// on the mesh: the commands that create the mesh network, and the compose
// lines that attach a service to it.
type NetworkSetup struct {
	Commands []string
	Service  []string
}

// MeshNetworkSetup returns the setup for p (amendment A37), every value
// checked against its grammar first.
func MeshNetworkSetup(p NetworkParams) (NetworkSetup, error) {
	if p.Runtime == "" {
		p.Runtime = RuntimeDocker
	}
	var errs []error
	keep := func(err error) {
		if err != nil {
			errs = append(errs, err)
		}
	}
	_, err := validate.Enum("runtime", p.Runtime, RuntimeDocker, RuntimePodman, RuntimeSwarm)
	keep(err)
	for _, f := range [][2]string{{"network", p.Network}, {"range", p.Range}} {
		_, err = validate.Prefix4(f[0], f[1])
		keep(err)
	}
	for _, f := range [][2]string{{"bridge", p.Bridge}, {"gateway", p.Gateway}} {
		_, err = validate.IPv4(f[0], f[1])
		keep(err)
	}
	_, err = validate.Zone("zone", p.Zone)
	keep(err)
	for i, r := range p.Resolvers {
		_, err = validate.IPv4(fmt.Sprintf("resolvers[%d]", i), r)
		keep(err)
	}
	keep(checkIPv6(p.Network6, p.Range6, p.Bridge6))
	if err := errors.Join(errs...); err != nil {
		return NetworkSetup{}, fmt.Errorf("the mesh network's commands: %w", err)
	}
	return NetworkSetup{Commands: networkCommands(p), Service: serviceExample(resolvers(p.Runtime, p.Gateway, p.Resolvers), p.Zone)}, nil
}

// networkCommands are the commands that create the node's mesh network for
// a runtime (amendment A37, settled by the M6 trial): the whole mesh on the
// link, the node's upper half handed out, the bridge at .1.
//
// In container mode Docker and Podman make it internal (Podman also
// without its DNS, whose listener on an internal network answers NXDOMAIN
// for every outside name), and the agent in its own namespace is the
// gateway. In host mode the agent owns the bridge: Docker takes it as it is
// (inhibit_ipv4), masquerades nothing and filters no port the mesh may
// reach (nat-unprotected; the agent's guard keeps the subnet to the mesh),
// Podman attaches to it unmanaged. Swarm needs a config-only network on
// each node and one swarm-scoped network from it. With the containers'
// IPv6 (M8, the 8a trial) the network is dual stack: the mesh's IPv6 prefix
// on the link too, the node's /112 upper half handed out, the bridge at
// ::1 -- each IPv6 value after its IPv4 one, as Podman pairs them -- and a
// host-mode Docker bridge leaves IPv6 unfiltered too (the guard holds it).
func networkCommands(p NetworkParams) []string {
	ipam := fmt.Sprintf("--subnet %s --ip-range %s --gateway %s", p.Network, p.Range, p.Bridge)
	if p.Network6 != "" {
		ipam += fmt.Sprintf(" --ipv6 --subnet %s --ip-range %s --gateway %s", p.Network6, p.Range6, p.Bridge6)
	}
	docker := []string{"-o com.docker.network.bridge.name=" + MeshBridge}
	if p.Host {
		docker = []string{
			"-o com.docker.network.bridge.name=" + MeshBridge + " -o com.docker.network.bridge.inhibit_ipv4=true",
			"-o com.docker.network.bridge.enable_ip_masquerade=false -o com.docker.network.bridge.gateway_mode_ipv4=nat-unprotected",
		}
		if p.Network6 != "" {
			docker[1] += " -o com.docker.network.bridge.gateway_mode_ipv6=nat-unprotected"
		}
	}
	switch {
	case p.Runtime == RuntimePodman && p.Host:
		return []string{
			fmt.Sprintf("  podman network create --disable-dns %s \\", ipam),
			fmt.Sprintf("    --interface-name %s -o mode=unmanaged -o no_default_route=true %s", MeshBridge, MeshNetwork),
		}
	case p.Runtime == RuntimePodman:
		return []string{
			fmt.Sprintf("  podman network create --internal --disable-dns %s \\", ipam),
			fmt.Sprintf("    --interface-name %s %s", MeshBridge, MeshNetwork),
		}
	case p.Runtime == RuntimeSwarm && p.Host:
		return []string{
			"on this node:",
			fmt.Sprintf("  docker network create --config-only %s \\", ipam),
			"    " + docker[0] + " \\",
			"    " + docker[1] + " \\",
			"    " + MeshNetwork + "-config",
			"once, on a manager (every node with mesh services has its own -config):",
			fmt.Sprintf("  docker network create -d bridge --scope swarm --attachable --config-from %s-config %s",
				MeshNetwork, MeshNetwork),
		}
	case p.Runtime == RuntimeSwarm:
		return []string{
			"on this node:",
			fmt.Sprintf("  docker network create --config-only %s %s-config", ipam, MeshNetwork),
			"once, on a manager (every node with mesh services has its own -config):",
			fmt.Sprintf("  docker network create -d bridge --scope swarm --attachable --config-from %s-config %s",
				MeshNetwork, MeshNetwork),
		}
	case p.Host:
		return []string{
			fmt.Sprintf("  docker network create %s \\", ipam),
			"    " + docker[0] + " \\",
			"    " + docker[1] + " \\",
			"    " + MeshNetwork,
		}
	default:
		return []string{
			fmt.Sprintf("  docker network create --internal %s \\", ipam),
			fmt.Sprintf("    %s %s", docker[0], MeshNetwork),
		}
	}
}

// resolvers are what a service on the mesh network names as its DNS: the
// node's gateway, or for Swarm, whose tasks run on several nodes with
// gateways of their own, the servers (every resolver forwards for every
// container network).
func resolvers(runtime, gateway string, servers []string) []string {
	if runtime == RuntimeSwarm && len(servers) > 0 {
		return servers
	}
	return []string{gateway}
}

// serviceExample is how a compose service joins the mesh network and uses
// the resolvers dns (ndots:1: Docker writes ndots:0, which keeps a short
// name from trying the zone).
func serviceExample(dns []string, zone string) []string {
	return []string{
		"  services:",
		"    web:",
		"      networks: [default, " + MeshNetwork + "]",
		"      dns: [" + strings.Join(dns, ", ") + "]",
		"      dns_search: [" + zone + "]",
		"      dns_opt: [\"ndots:1\"]",
		"  networks:",
		"    " + MeshNetwork + ":",
		"      external: true",
	}
}

func containerHeader(p ContainerParams) []string {
	compose := "docker compose"
	if p.Runtime == RuntimePodman {
		compose = "podman compose"
	}
	role := "client: it dials the servers' public endpoints"
	if p.Hub {
		role = fmt.Sprintf("server: the host publishes %d/udp, which the node's endpoint names", p.PublishPort)
	}
	backend := "WireGuard: the host's module when it is loaded, else wireguard-go on /dev/net/tun."
	switch p.Backend {
	case ContainerKernel:
		backend = "WireGuard: the host's module (ls /sys/module/wireguard)."
	case ContainerUserspace:
		backend = "WireGuard: wireguard-go on the host's /dev/net/tun."
	}
	lines := []string{
		fmt.Sprintf("meshctl agent for %q in container mode (amendment A37), a %s.", p.Node, role),
		"Only containers on the " + MeshNetwork + " network are on the mesh; the host is not.",
		backend,
	}
	if p.Network6 != "" {
		lines = append(lines, "IPv6: the containers get addresses from "+p.Range6+" too; the agent answers neighbour discovery for the mesh.")
	}
	lines = append(lines,
		"",
		"1. Create the mesh network ("+p.Runtime+"):",
	)
	lines = append(lines, networkCommands(NetworkParams{Runtime: p.Runtime, Network: p.Network, Range: p.Range, Bridge: p.Bridge,
		Network6: p.Network6, Range6: p.Range6, Bridge6: p.Bridge6})...)
	lines = append(lines,
		"2. Join once, with the node's join token (web UI, `meshctl node add` or `node token`):",
		fmt.Sprintf("  MESHCTL_TOKEN=<token> %s run --rm -e MESHCTL_TOKEN %s \\", compose, ContainerService),
		fmt.Sprintf("    agent join --data /data --server %s --name %s --no-service", p.Server, p.Node),
		"3. Start it: "+compose+" up -d",
		"",
		"Put a service on the mesh: attach the network and use the agent as its resolver",
		"(ndots:1 lets a short name try the zone; Docker writes ndots:0):",
	)
	lines = append(lines, serviceExample(resolvers(p.Runtime, p.Gateway, p.Resolvers), p.Zone)...)
	if p.Runtime == RuntimeSwarm && len(p.Resolvers) > 0 {
		lines = append(lines, "(A Swarm service's tasks run on several nodes, so its dns is the servers: every",
			"node's resolver answers every node's containers.)")
	}
	lines = append(lines, "")
	if p.NoNames {
		lines = append(lines, "Names: none; the containers resolve by address only (--no-names).")
	} else {
		lines = append(lines,
			"Names: <service>."+p.Zone+" (every node's, round robin) and <service>.<node>."+p.Zone+",",
			"from compose services, container names, aliases and hostnames, which the agent",
			"reads from the runtime's socket (GET only). The socket gives whoever holds it",
			"the runtime; `meshctl node compose --no-names` leaves it out.",
		)
		if p.Runtime == RuntimePodman {
			lines = append(lines,
				"Podman: systemctl enable --now podman.socket; rootless, mount",
				"/run/user/<uid>/podman/podman.sock instead (systemctl --user enable --now podman.socket).",
			)
		}
	}
	lines = append(lines,
		"",
		"Posture (docs/security.md): not privileged; every capability dropped except",
		"NET_ADMIN and NET_BIND_SERVICE, which act in the agent's own network namespace",
		"only; no privilege escalation; read-only root. It runs as uid 0 because",
		"capabilities do not survive no-new-privileges for another user.",
	)
	return lines
}
