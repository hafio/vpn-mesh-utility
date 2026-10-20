// Package proto is the wire contract between the meshctl server and its
// agents: the desired-state document the server computes per node, the
// request-signing scheme, the agent API payloads and the SSE framing.
//
// It is the one package both sides import (with internal/validate), so it
// holds data shapes and pure functions only -- no I/O, no storage, no
// platform code. docs/agent-protocol.md describes the same contract in prose.
package proto

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

// Role is a node's place in the topology (spec S4).
type Role string

const (
	RoleServer Role = "server"
	RoleSpoke  Role = "spoke"
	RolePhone  Role = "phone"
)

// Defaults shared by the planner (which emits them) and the agent (which
// applies them).
const (
	DefaultListenPort = 51820
	DefaultMTU        = 1420
	DefaultKeepalive  = 25
	// FwMark marks the agent's own WireGuard UDP so exit-mode policy routing
	// never sends the tunnel's outer packets into the tunnel.
	FwMark = 51820
	// RouteTable holds the exit-mode default route (spec S7.3).
	RouteTable = 51820
	// DefaultZone is the DNS zone a mesh serves until its settings name
	// another (amendment A36).
	DefaultZone = "mesh"
	// MinMTU and MaxMTU bound the per-node override.
	MinMTU = 1280
	MaxMTU = 1500
)

// NodeConfig is the full desired state of one node (spec S6.3). The server
// computes it, the agent applies it idempotently and caches it on disk. Every
// slice is sorted by the planner so the canonical encoding, and with it Hash,
// is deterministic.
type NodeConfig struct {
	// Version is the node's config_version: the global version at which this
	// node's computed config last changed. It is excluded from Hash.
	Version   int64     `json:"version"`
	Node      Self      `json:"node"`
	Interface Interface `json:"interface"`
	Peers     []Peer    `json:"peers"`
	Exit      *Exit     `json:"exit,omitempty"`
	DNS       *DNS      `json:"dns,omitempty"`
	Firewall  *Firewall `json:"firewall,omitempty"`
	Sysctl    []Sysctl  `json:"sysctl,omitempty"`
	// Containers is set when the node's containers are on the mesh
	// (amendment A37). It is last and omitted otherwise, so a mesh without
	// container networks encodes, and hashes, as before.
	Containers *Containers `json:"containers,omitempty"`
}

// Self identifies the node the config is for.
type Self struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Role Role   `json:"role"`
}

// Interface is the local WireGuard interface.
type Interface struct {
	// Address is the node's mesh IP as a /32 prefix.
	Address string `json:"address"`
	// ListenPort is fixed on servers and null on spokes (they dial out).
	ListenPort *int `json:"listen_port"`
	MTU        int  `json:"mtu"`
	// FwMark is set on spokes, which may switch to exit mode at any time.
	FwMark int `json:"fwmark,omitempty"`
}

// Peer is one WireGuard peer of the node.
type Peer struct {
	Name      string `json:"name"`
	PublicKey string `json:"public_key"`
	// Endpoint is empty for peers that dial in (spokes, phones).
	Endpoint   string   `json:"endpoint,omitempty"`
	AllowedIPs []string `json:"allowed_ips"`
	// Keepalive is PersistentKeepalive in seconds; 0 means none.
	Keepalive int `json:"keepalive,omitempty"`
}

// Exit is a spoke's internet-exit selection.
type Exit struct {
	// Via names the server whose peer carries 0.0.0.0/0; empty means exit off.
	Via string `json:"via,omitempty"`
	// Choices lists the exit-capable servers, for the local exit picker.
	Choices []string `json:"choices,omitempty"`
}

// DNS is the node's resolver configuration (spec S7.4).
type DNS struct {
	Zone     string   `json:"zone"`
	Records  []Record `json:"records"`
	Upstream []string `json:"upstream,omitempty"`
	// Split true answers only the mesh zone and REFUSES everything else.
	Split bool `json:"split"`
	// ForwardFor lists source prefixes (phones that use this node as their
	// resolver) whose non-mesh queries are forwarded even when Split is true.
	ForwardFor []string `json:"forward_for,omitempty"`
	// Domains are the mesh's extra domains (amendment A39), sorted: the
	// containers' names in them are answered, every other name in them is
	// forwarded, and the host sends them to the resolver too. Last and
	// omitted when there are none, so a config without them hashes as
	// before.
	Domains []string `json:"domains,omitempty"`
}

// Record is one A record in the mesh zone.
type Record struct {
	Name string `json:"name"`
	IP   string `json:"ip"`
}

// Node modes (amendment A37): where a node's WireGuard lives.
const (
	// ModeHost is WireGuard on the host (the binary, or the NAS agent on the
	// host network): the whole machine is on the mesh.
	ModeHost = "host"
	// ModeContainer is WireGuard in the agent container's own network
	// namespace: only the containers on the node's mesh network are on the
	// mesh, and the agent is their gateway.
	ModeContainer = "container"
)

// The runtime network containers attach to, and its bridge on the host.
const (
	MeshNetwork     = "meshctl-mesh"
	ContainerBridge = "meshctl0"
)

// Containers is a node's mesh network for containers (amendment A37).
// Subnet is the node's /24 of the mesh. The runtime network carries the
// whole mesh as its on-link Network, so an attached container ARPs for
// every mesh address and the gateway answers by proxy ARP; the runtime
// hands out Range, and Bridge is its bridge address. Gateway is where the
// containers' mesh traffic and DNS go: the agent at .2 in container mode,
// the host's bridge (Bridge) in host mode. Relay lets a container-mode hub
// forward between its peers, as a host-mode one's firewall does.
type Containers struct {
	Mode    string `json:"mode"`
	Subnet  string `json:"subnet"`
	Network string `json:"network"`
	Range   string `json:"range"`
	Bridge  string `json:"bridge"`
	Gateway string `json:"gateway"`
	Relay   bool   `json:"relay,omitempty"`
}

// FirewallFirewalld is the firewall backend of a host-mode server. A
// container-mode node has no firewall section: its gateway's isolation is
// policy routing in its own namespace, derived from Containers.
const FirewallFirewalld = "firewalld"

// Firewall is the rendered firewall state for a server node (spec S7.5).
type Firewall struct {
	// Backend is FirewallFirewalld.
	Backend string   `json:"backend"`
	Rules   []FWRule `json:"rules"`
	// Relay permits mesh->mesh forwarding (this server relays spoke traffic,
	// or between wg0 and its containers' bridge).
	Relay bool `json:"relay,omitempty"`
	// ExitGateway permits mesh->public forwarding with masquerade.
	ExitGateway bool `json:"exit_gateway,omitempty"`
	// Bridge is the containers' bridge of a host-mode server with a
	// container network (amendment A37): if no zone holds it, the agent
	// puts it in the mesh zone.
	Bridge string `json:"bridge,omitempty"`
}

// Firewall zones.
const (
	ZonePublic = "public"
	ZoneMesh   = "mesh"
)

// FWRule opens one port in a zone, optionally from one source prefix only.
type FWRule struct {
	Zone   string `json:"zone"`
	Proto  string `json:"proto"`
	Port   int    `json:"port"`
	Source string `json:"source,omitempty"`
	// Note says where the rule comes from ("swarm: vps-2"). Display only; it
	// never reaches firewall-cmd.
	Note string `json:"note,omitempty"`
}

// Sysctl is one kernel parameter the node needs.
type Sysctl struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// Sysctl keys the server may ask for. The agent refuses any other key.
const (
	SysctlIPForward    = "net.ipv4.ip_forward"
	SysctlSrcValidMark = "net.ipv4.conf.all.src_valid_mark"
	// SysctlProxyARP lets a container-mode gateway answer ARP for every
	// mesh address it routes; the compose file sets it.
	SysctlProxyARP = "net.ipv4.conf.all.proxy_arp"
	// The host-mode container bridge's own forwarding and proxy ARP, and
	// wg0's forwarding: a new interface can come up with forwarding off
	// although ip_forward is on. The agent's containers step sets them, once
	// the bridge exists; the server never asks for them.
	SysctlBridgeForwarding = "net.ipv4.conf." + ContainerBridge + ".forwarding"
	SysctlBridgeProxyARP   = "net.ipv4.conf." + ContainerBridge + ".proxy_arp"
	SysctlWGForwarding     = "net.ipv4.conf.wg0.forwarding"
)

// SysctlKeys is every key the server may ask for, for the allowlists.
func SysctlKeys() []string {
	return []string{SysctlIPForward, SysctlSrcValidMark, SysctlProxyARP, SysctlBridgeForwarding, SysctlBridgeProxyARP,
		SysctlWGForwarding}
}

// Hash returns the hex SHA-256 of the canonical encoding with Version zeroed.
// Two configs with the same Hash apply to the same state.
func (c NodeConfig) Hash() (string, error) {
	c.Version = 0
	b, err := json.Marshal(c)
	if err != nil {
		return "", fmt.Errorf("hash node config %q: %w", c.Node.Name, err)
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:]), nil
}

// ConfigEvent is the payload of an SSE `config` event.
type ConfigEvent struct {
	Version int64  `json:"version"`
	Hash    string `json:"hash"`
}
