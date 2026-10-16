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
	// Zone is the DNS zone every node serves.
	Zone = "mesh"
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
}

// Record is one A record in the mesh zone.
type Record struct {
	Name string `json:"name"`
	IP   string `json:"ip"`
}

// Firewall is the rendered firewall state for a server node (spec S7.5).
type Firewall struct {
	// Backend is "firewalld" on server nodes.
	Backend string   `json:"backend"`
	Rules   []FWRule `json:"rules"`
	// Relay permits mesh->mesh forwarding (this server relays spoke traffic).
	Relay bool `json:"relay,omitempty"`
	// ExitGateway permits mesh->public forwarding with masquerade.
	ExitGateway bool `json:"exit_gateway,omitempty"`
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
)

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
