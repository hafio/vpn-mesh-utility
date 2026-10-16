// Package testfix holds the shared test fixtures: deterministic keys, a fake
// clock and the reference topology (three VPS servers, a NAS, a laptop and a
// phone) that server and agent tests build on.
//
// It imports only internal/proto, so tests on both sides of the server/agent
// boundary can use it without either importing the other. Production code
// must never import it (internal/archtest enforces this).
package testfix

import (
	"bytes"
	"crypto/ed25519"
	"encoding/base64"
	"sync"
	"time"

	"github.com/hafio/vpn-mesh-utility/internal/proto"
)

// Epoch is the fixed "now" of every test that does not care about time.
var Epoch = time.Date(2026, 10, 12, 8, 0, 0, 0, time.UTC)

// WGKey is a deterministic WireGuard-shaped public key: base64 of 32 bytes of n.
// It is a valid encoding, not a valid curve point; the server never needs one.
func WGKey(n byte) string {
	return base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{n}, 32))
}

// IdentityKey is a deterministic Ed25519 key pair seeded from n.
func IdentityKey(n byte) (ed25519.PublicKey, ed25519.PrivateKey) {
	priv := ed25519.NewKeyFromSeed(bytes.Repeat([]byte{n}, ed25519.SeedSize))
	return priv.Public().(ed25519.PublicKey), priv
}

// IdentityKeyB64 is the base64 public half of IdentityKey(n), as an agent
// sends it on join.
func IdentityKeyB64(n byte) string {
	pub, _ := IdentityKey(n)
	return base64.StdEncoding.EncodeToString(pub)
}

// Clock is a manually advanced clock, safe for concurrent use.
type Clock struct {
	mu  sync.Mutex
	now time.Time
}

// NewClock returns a clock stopped at t.
func NewClock(t time.Time) *Clock { return &Clock{now: t} }

// Now returns the current fake time.
func (c *Clock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

// Advance moves the clock forward by d.
func (c *Clock) Advance(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.now = c.now.Add(d)
}

// Node is one node of the reference topology. Key is the n passed to WGKey
// and IdentityKey, so every node has distinct, reproducible keys.
type Node struct {
	ID       int64
	Name     string
	Role     proto.Role
	MeshIP   string
	Endpoint string
	Key      byte
}

// The reference topology (spec S1 and S4).
var (
	VPS1   = Node{ID: 1, Name: "vps-1", Role: proto.RoleServer, MeshIP: "10.77.0.1", Endpoint: "203.0.113.10:51820", Key: 1}
	VPS2   = Node{ID: 2, Name: "vps-2", Role: proto.RoleServer, MeshIP: "10.77.0.2", Endpoint: "203.0.113.11:51820", Key: 2}
	VPS3   = Node{ID: 3, Name: "vps-3", Role: proto.RoleServer, MeshIP: "10.77.0.3", Endpoint: "203.0.113.12:51820", Key: 3}
	NAS    = Node{ID: 4, Name: "nas", Role: proto.RoleSpoke, MeshIP: "10.77.1.1", Key: 4}
	Laptop = Node{ID: 5, Name: "laptop", Role: proto.RoleSpoke, MeshIP: "10.77.2.1", Key: 5}
	Phone  = Node{ID: 6, Name: "phone", Role: proto.RolePhone, MeshIP: "10.77.3.1", Key: 6}
)

// Servers is the three-VPS Swarm base.
var Servers = []Node{VPS1, VPS2, VPS3}

// MeshCIDR is the default mesh range.
const MeshCIDR = "10.77.0.0/16"

// Token is a well-formed join token (43 base64url chars).
const Token = "dGVzdC10b2tlbi10ZXN0LXRva2VuLXRlc3QtdG9rZW4"

// ServerConfig is a valid desired state for vps-1 in the three-server mesh:
// the shape the planner emits and the agent applies. Callers may mutate the
// returned value freely.
func ServerConfig() proto.NodeConfig {
	port := proto.DefaultListenPort
	peer := func(n Node) proto.Peer {
		return proto.Peer{
			Name:       n.Name,
			PublicKey:  WGKey(n.Key),
			Endpoint:   n.Endpoint,
			AllowedIPs: []string{n.MeshIP + "/32"},
		}
	}
	return proto.NodeConfig{
		Version: 3,
		Node:    proto.Self{ID: VPS1.ID, Name: VPS1.Name, Role: proto.RoleServer},
		Interface: proto.Interface{
			Address:    VPS1.MeshIP + "/32",
			ListenPort: &port,
			MTU:        proto.DefaultMTU,
		},
		Peers: []proto.Peer{peer(VPS2), peer(VPS3)},
		Firewall: &proto.Firewall{
			Backend: "firewalld",
			Rules: []proto.FWRule{
				{Zone: proto.ZoneMesh, Proto: "tcp", Port: 2377, Source: VPS2.MeshIP + "/32", Note: "swarm: vps-2"},
				{Zone: proto.ZonePublic, Proto: "udp", Port: proto.DefaultListenPort, Note: "wireguard"},
			},
		},
	}
}

// PhoneConfig is the phone's computed config: a static profile whose hub,
// vps-1, carries the whole mesh, with vps-2 and vps-3 exit-capable.
func PhoneConfig() proto.NodeConfig {
	peer := func(n Node, allowed string) proto.Peer {
		return proto.Peer{Name: n.Name, PublicKey: WGKey(n.Key), Endpoint: n.Endpoint, AllowedIPs: []string{allowed}, Keepalive: proto.DefaultKeepalive}
	}
	return proto.NodeConfig{
		Version:   7,
		Node:      proto.Self{ID: Phone.ID, Name: Phone.Name, Role: proto.RolePhone},
		Interface: proto.Interface{Address: Phone.MeshIP + "/32", MTU: proto.DefaultMTU},
		Peers:     []proto.Peer{peer(VPS1, MeshCIDR), peer(VPS2, VPS2.MeshIP+"/32"), peer(VPS3, VPS3.MeshIP+"/32")},
		Exit:      &proto.Exit{Choices: []string{VPS2.Name, VPS3.Name}},
	}
}

// SpokeConfig is a valid desired state for the laptop: it peers with the
// three servers, vps-1 is its hub (and relays to the NAS), and vps-2 and
// vps-3 are exit-capable. A non-empty exit puts 0.0.0.0/0 on that server's
// peer and turns exit mode on.
func SpokeConfig(exit string) proto.NodeConfig {
	peer := func(n Node, extra ...string) proto.Peer {
		allowed := append([]string{n.MeshIP + "/32"}, extra...)
		if n.Name == exit {
			allowed = append(allowed, "0.0.0.0/0")
		}
		return proto.Peer{Name: n.Name, PublicKey: WGKey(n.Key), Endpoint: n.Endpoint, AllowedIPs: allowed, Keepalive: proto.DefaultKeepalive}
	}
	c := proto.NodeConfig{
		Version:   5,
		Node:      proto.Self{ID: Laptop.ID, Name: Laptop.Name, Role: proto.RoleSpoke},
		Interface: proto.Interface{Address: Laptop.MeshIP + "/32", MTU: proto.DefaultMTU, FwMark: proto.FwMark},
		Peers:     []proto.Peer{peer(VPS1, NAS.MeshIP+"/32"), peer(VPS2), peer(VPS3)},
		Exit:      &proto.Exit{Choices: []string{VPS2.Name, VPS3.Name}},
	}
	if exit != "" {
		c.Exit.Via = exit
		c.Sysctl = []proto.Sysctl{{Key: proto.SysctlSrcValidMark, Value: "1"}}
	}
	return c
}

// MeshView is the server's mesh view of the reference topology at Epoch:
// the servers in sync, the NAS applying v7, the laptop (exit vps-2) silent
// and the phone idle. Callers may mutate the returned value freely.
func MeshView() proto.MeshView {
	now := Epoch.Unix()
	node := func(n Node, health, state string, seen int64) proto.MeshNode {
		return proto.MeshNode{Name: n.Name, Role: n.Role, MeshIP: n.MeshIP, Health: health, State: state, LastSeen: seen,
			ConfigVersion: 7, AppliedVersion: 7}
	}
	v := proto.MeshView{Version: 7, ServerTime: now, Nodes: []proto.MeshNode{
		node(VPS1, proto.HealthOK, "in sync", now-10),
		node(VPS2, proto.HealthOK, "in sync", now-12),
		node(VPS3, proto.HealthOK, "in sync", now-14),
		node(NAS, proto.HealthWarn, "applying v7", now-20),
		node(Laptop, proto.HealthBad, "silent for 10m0s", now-600),
		node(Phone, proto.HealthOff, "idle", now-3600),
	}}
	v.Nodes[1].CanExit, v.Nodes[2].CanExit = true, true
	v.Nodes[3].AppliedVersion = 6
	for i := 3; i < 6; i++ {
		v.Nodes[i].Hub = VPS1.Name
	}
	v.Nodes[4].Exit = VPS2.Name
	return v
}
