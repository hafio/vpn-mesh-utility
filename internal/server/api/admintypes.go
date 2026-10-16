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
	// ListenPort is a server's WireGuard port; 0 means 51820.
	ListenPort int    `json:"listen_port,omitempty"`
	Hub        string `json:"hub,omitempty"`
	CanExit    bool   `json:"can_exit,omitempty"`
	// AlwaysOn places a spoke in the always-on block (the NAS).
	AlwaysOn  bool `json:"always_on,omitempty"`
	Keepalive int  `json:"keepalive,omitempty"`
	MTU       int  `json:"mtu,omitempty"`
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
}

// Enrollment is a freshly issued join token and the commands that use it.
// The token is shown once; the server keeps only its hash.
type Enrollment struct {
	Token     string `json:"token"`
	ExpiresAt int64  `json:"expires_at"`
	// Install is the one-line install for a Linux host.
	Install string `json:"install"`
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

// ComposeFile is GET /admin/v1/nodes/{name}/compose?backend=&data=: a
// spoke's container compose file and the choices it was rendered with.
type ComposeFile struct {
	Backend string `json:"backend"`
	DataDir string `json:"data_dir"`
	YAML    string `json:"yaml"`
}

// NodeDetail is GET /admin/v1/nodes/{name}: the node, what the server
// computed for it, and what it last reported about its peers.
type NodeDetail struct {
	Node     NodeInfo           `json:"node"`
	Config   *proto.NodeConfig  `json:"config,omitempty"`
	Peers    []proto.PeerStatus `json:"peers,omitempty"`
	Warnings []string           `json:"warnings,omitempty"`
}
