package api

import "github.com/hafio/vpn-mesh-utility/internal/proto"

// Admin API paths, served on <data>/admin.sock (amendment A9).
const (
	AdminNodes = "/admin/v1/nodes"
)

// NodeInfo is one node as the admin API shows it. It never carries a token
// or a private key.
type NodeInfo struct {
	ID             int64  `json:"id"`
	Name           string `json:"name"`
	Role           string `json:"role"`
	MeshIP         string `json:"mesh_ip"`
	Endpoint       string `json:"endpoint,omitempty"`
	ListenPort     int    `json:"listen_port"`
	Hub            string `json:"hub,omitempty"`
	CanExit        bool   `json:"can_exit,omitempty"`
	Exit           string `json:"exit,omitempty"`
	Keepalive      int    `json:"keepalive"`
	MTU            int    `json:"mtu"`
	Enrolled       bool   `json:"enrolled"`
	Disabled       bool   `json:"disabled"`
	OS             string `json:"os,omitempty"`
	Arch           string `json:"arch,omitempty"`
	Hostname       string `json:"hostname,omitempty"`
	WGBackend      string `json:"wg_backend,omitempty"`
	AgentVersion   string `json:"agent_version,omitempty"`
	ConfigVersion  int64  `json:"config_version"`
	AppliedVersion int64  `json:"applied_version"`
	// InSync means the node reported applying exactly its current config;
	// for a phone, that its last profile is still current.
	InSync   bool  `json:"in_sync"`
	LastSeen int64 `json:"last_seen,omitempty"`
	// LastHandshake is, for a phone, the latest handshake a server
	// reported with it.
	LastHandshake int64 `json:"last_handshake,omitempty"`
	// Degraded, Errors, Rollback and DNS are from the node's last
	// heartbeat.
	Degraded []string        `json:"degraded,omitempty"`
	Errors   []string        `json:"errors,omitempty"`
	Rollback *proto.Rollback `json:"rollback,omitempty"`
	DNS      string          `json:"dns,omitempty"`
	// Mode is where the node's WireGuard lives; ContainerSubnet is its /24
	// for containers, empty while they are not on the mesh (amendment A37).
	Mode            string `json:"mode"`
	ContainerSubnet string `json:"container_subnet,omitempty"`
	// DownSince is when relay failover judged this server down (amendment
	// A40); 0 while it is up.
	DownSince int64 `json:"down_since,omitempty"`
	// Groups are the segmentation groups the node is in (amendment A43).
	Groups []string `json:"groups,omitempty"`
	// The IPv6 side (M8): the node's IPv6 mesh address, its address
	// families (dual, ipv4 or ipv6), a server's IPv6 endpoint and its
	// containers' IPv6 /112.
	MeshIP6          string `json:"mesh_ip6,omitempty"`
	Family           string `json:"ip_family"`
	Endpoint6        string `json:"endpoint6,omitempty"`
	ContainerSubnet6 string `json:"container_subnet6,omitempty"`
	// HostFamilies are the families its agent reported the host can use;
	// empty when it has not said.
	HostFamilies []string `json:"host_families,omitempty"`
	// Copy is another copy of the node the control plane saw (amendment
	// A50): one waiting while this one is asked whether it runs, else the
	// last one refused.
	Copy *CopyInfo `json:"copy,omitempty"`
}

// CopyInfo is a second copy of a node: Waiting since At while the holder
// is challenged, or refused at At as Code (proto.CodeDuplicate or
// proto.CodeForeign); Hostname is where it runs.
type CopyInfo struct {
	Waiting     bool   `json:"waiting,omitempty"`
	Code        string `json:"code,omitempty"`
	Hostname    string `json:"hostname"`
	SameMachine bool   `json:"same_machine,omitempty"`
	At          int64  `json:"at"`
}

// NodeList is GET /admin/v1/nodes.
type NodeList struct {
	Nodes    []NodeInfo `json:"nodes"`
	Warnings []string   `json:"warnings,omitempty"`
}

// NodeAddRequest is POST /admin/v1/nodes. Zero values mean "default".
type NodeAddRequest struct {
	Name string `json:"name"`
	Role string `json:"role"`
	// IP is a fixed mesh IP; empty allocates the lowest free one in the
	// role's block (spec S4).
	IP       string `json:"ip,omitempty"`
	Endpoint string `json:"endpoint,omitempty"`
	// ListenPort is a server's WireGuard port; 0 means the mesh's default,
	// 51820 plus its slot (amendment A49).
	ListenPort int    `json:"listen_port,omitempty"`
	Hub        string `json:"hub,omitempty"`
	CanExit    bool   `json:"can_exit,omitempty"`
	// AlwaysOn places a spoke in the always-on block (the NAS).
	AlwaysOn  bool `json:"always_on,omitempty"`
	Keepalive int  `json:"keepalive,omitempty"`
	MTU       int  `json:"mtu,omitempty"`
	// Mode is proto.ModeHost (the default) or proto.ModeContainer; a
	// container-mode node always has a container subnet, a host-mode one
	// when Containers is set (amendment A37).
	Mode       string `json:"mode,omitempty"`
	Containers bool   `json:"containers,omitempty"`
	// Groups puts the node in these segmentation groups, creating one that
	// does not exist yet (amendment A43).
	Groups []string `json:"groups,omitempty"`
	// The IPv6 side (M8): Family is dual (the default), ipv4 or ipv6; IP6 a
	// fixed IPv6 mesh address, empty to allocate one when the mesh has an
	// IPv6 prefix and the family is not ipv4; Endpoint6 a server's public
	// [IPv6]:port; Containers6 puts the containers' IPv6 on the mesh too.
	Family      string `json:"ip_family,omitempty"`
	IP6         string `json:"ip6,omitempty"`
	Endpoint6   string `json:"endpoint6,omitempty"`
	Containers6 bool   `json:"containers6,omitempty"`
}

// NodeSetRequest is PATCH /admin/v1/nodes/{name}: only non-nil fields
// change. An empty Endpoint, Hub or Exit clears it.
type NodeSetRequest struct {
	Endpoint   *string `json:"endpoint,omitempty"`
	ListenPort *int    `json:"listen_port,omitempty"`
	Hub        *string `json:"hub,omitempty"`
	CanExit    *bool   `json:"can_exit,omitempty"`
	Exit       *string `json:"exit,omitempty"`
	Keepalive  *int    `json:"keepalive,omitempty"`
	MTU        *int    `json:"mtu,omitempty"`
	// Containers gives a host-mode node a container subnet, or takes it
	// away; a node's mode is chosen when it is added.
	Containers *bool `json:"containers,omitempty"`
	// MeshIP moves the node to another mesh IP (amendment A41): inside the
	// mesh and not taken. Its peers, its DNS name and its agent follow.
	MeshIP *string `json:"mesh_ip,omitempty"`
	// Groups replaces the node's segmentation groups; empty takes it out
	// of every group (amendment A43).
	Groups *[]string `json:"groups,omitempty"`
	// The IPv6 side (M8): a new Family allocates the address it adds and
	// drops the one it takes away; MeshIP6 moves the IPv6 address;
	// Endpoint6 sets a server's IPv6 endpoint ("" clears it); Containers6
	// puts the containers' IPv6 on the mesh or takes it off.
	Family      *string `json:"ip_family,omitempty"`
	MeshIP6     *string `json:"mesh_ip6,omitempty"`
	Endpoint6   *string `json:"endpoint6,omitempty"`
	Containers6 *bool   `json:"containers6,omitempty"`
}

// Enrollment is a freshly issued join token and the commands that use it.
// The token is shown once; the server keeps only its hash.
type Enrollment struct {
	Token     string `json:"token"`
	ExpiresAt int64  `json:"expires_at"`
	// Install is the one-line install for a Linux host, InstallWindows for
	// an elevated PowerShell (amendment A51).
	Install        string `json:"install"`
	InstallWindows string `json:"install_windows"`
	// Join is the agent command for a host that already has meshctl.
	Join string `json:"join"`
}

// NodeChange is the answer to every admin write.
type NodeChange struct {
	Node NodeInfo `json:"node"`
	// Version is the config version the change created; 0 when no node's
	// computed config changed.
	Version    int64       `json:"version"`
	Enrollment *Enrollment `json:"enrollment,omitempty"`
	Warnings   []string    `json:"warnings,omitempty"`
	// Notes are what the operator must do by hand after the change: what
	// a new mesh IP does not reach (Swarm, a phone's profile, a rule).
	Notes []string `json:"notes,omitempty"`
}

// PhoneRequest is POST /admin/v1/nodes/{name}/phone. An empty Exit is the
// mesh-only profile; a server name sends the phone's internet through it.
type PhoneRequest struct {
	Exit string `json:"exit,omitempty"`
}

// PhoneProfile is a phone's new WireGuard profile (amendment A6). Profile
// holds the phone's private key: it is in this answer only, the server
// keeps the public key, and the previous profile stops working.
type PhoneProfile struct {
	Node    NodeInfo `json:"node"`
	Version int64    `json:"version"`
	Exit    string   `json:"exit,omitempty"`
	// Profile is wg-quick's format, as the WireGuard app imports it.
	Profile  string   `json:"profile"`
	Warnings []string `json:"warnings,omitempty"`
}

// ContainerNetwork is a node's container network as its host's runtime
// creates it (amendments A37, A48): the values docs/containers.md's samples
// take, since meshctl writes no compose file. Name is the runtime network,
// BridgeName the bridge a host-mode node's agent owns. Network is the whole
// mesh on the link, Range the node's addresses the runtime hands out,
// Bridge the runtime's side (its --gateway) and Gateway the agent: its own
// static address in container mode, the bridge in host mode. A service
// names Gateway as its resolver; a Swarm service, whose tasks run on
// several nodes, names Resolvers, the server nodes' mesh IPs, with Zone as
// its search domain either way. A container-mode server publishes
// PublishPort/udp on its host to ListenPort, the agent's.
type ContainerNetwork struct {
	Name        string   `json:"name"`
	BridgeName  string   `json:"bridge_name"`
	Mode        string   `json:"mode"`
	Subnet      string   `json:"subnet"`
	Network     string   `json:"network"`
	Range       string   `json:"range"`
	Bridge      string   `json:"bridge"`
	Gateway     string   `json:"gateway"`
	Zone        string   `json:"zone"`
	Resolvers   []string `json:"resolvers,omitempty"`
	Subnet6     string   `json:"subnet6,omitempty"`
	Network6    string   `json:"network6,omitempty"`
	Range6      string   `json:"range6,omitempty"`
	Bridge6     string   `json:"bridge6,omitempty"`
	Gateway6    string   `json:"gateway6,omitempty"`
	PublishPort int      `json:"publish_port,omitempty"`
	ListenPort  int      `json:"listen_port,omitempty"`
}

// NodeDetail is GET /admin/v1/nodes/{name}: the node, what the server
// computed for it, and what it last reported about its peers.
type NodeDetail struct {
	Node       NodeInfo           `json:"node"`
	Containers *ContainerNetwork  `json:"containers,omitempty"`
	Config     *proto.NodeConfig  `json:"config,omitempty"`
	Peers      []proto.PeerStatus `json:"peers,omitempty"`
	Warnings   []string           `json:"warnings,omitempty"`
}
