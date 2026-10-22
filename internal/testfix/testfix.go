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
// and IdentityKey, so every node has distinct, reproducible keys. MeshIP6
// is its address in MeshCIDR6 for the dual-stack fixtures (M8); only vps-2
// has a public IPv6 endpoint.
type Node struct {
	ID        int64
	Name      string
	Role      proto.Role
	MeshIP    string
	MeshIP6   string
	Endpoint  string
	Endpoint6 string
	Key       byte
}

// The reference topology (spec S1 and S4).
var (
	VPS1   = Node{ID: 1, Name: "vps-1", Role: proto.RoleServer, MeshIP: "10.77.0.1", MeshIP6: "fd10:77::1", Endpoint: "203.0.113.10:51820", Key: 1}
	VPS2   = Node{ID: 2, Name: "vps-2", Role: proto.RoleServer, MeshIP: "10.77.0.2", MeshIP6: "fd10:77::2", Endpoint: "203.0.113.11:51820", Endpoint6: "[2001:db8::11]:51820", Key: 2}
	VPS3   = Node{ID: 3, Name: "vps-3", Role: proto.RoleServer, MeshIP: "10.77.0.3", MeshIP6: "fd10:77::3", Endpoint: "203.0.113.12:51820", Key: 3}
	NAS    = Node{ID: 4, Name: "nas", Role: proto.RoleSpoke, MeshIP: "10.77.1.1", MeshIP6: "fd10:77::1:1", Key: 4}
	Laptop = Node{ID: 5, Name: "laptop", Role: proto.RoleSpoke, MeshIP: "10.77.2.1", MeshIP6: "fd10:77::2:1", Key: 5}
	Phone  = Node{ID: 6, Name: "phone", Role: proto.RolePhone, MeshIP: "10.77.3.1", MeshIP6: "fd10:77::3:1", Key: 6}
)

// Servers is the three-VPS Swarm base.
var Servers = []Node{VPS1, VPS2, VPS3}

// MeshCIDR is the default mesh range; MeshCIDR6 the dual-stack fixtures'
// IPv6 prefix (M8).
const (
	MeshCIDR  = "10.77.0.0/16"
	MeshCIDR6 = "fd10:77::/64"
)

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

// Policy is a server's mesh firewall as the planner compiles it for this
// mesh (amendment A43): the system rules, the laptop allowed SMB to the
// NAS and denied every other server port, a zone of the spokes allowed
// within it. The rest is dropped.
func Policy() *proto.Policy {
	servers := []string{VPS1.MeshIP + "-" + VPS3.MeshIP}
	spokes := []string{NAS.MeshIP, Laptop.MeshIP}
	svc := func(name string, ports ...string) proto.Service { return proto.Service{Proto: name, Ports: ports} }
	return &proto.Policy{Mesh: MeshCIDR, Rules: []proto.PolicyRule{
		{Name: "system: ping", Services: []proto.Service{svc("icmp")}, Action: proto.PolicyAllow},
		{Name: "system: dns", To: servers, Services: []proto.Service{svc("tcp", "53"), svc("udp", "53")}, Action: proto.PolicyAllow},
		{Name: "system: swarm", From: servers, To: servers, Services: []proto.Service{svc("tcp", "2377", "7946"), svc("udp", "4789", "7946")},
			Action: proto.PolicyAllow},
		{Name: "policy: laptop-smb", From: []string{Laptop.MeshIP}, To: []string{NAS.MeshIP}, Services: []proto.Service{svc("tcp", "445")},
			Action: proto.PolicyAllow},
		{Name: "policy: laptop-servers", From: []string{Laptop.MeshIP}, To: servers, Action: proto.PolicyDeny},
		{Name: "zone: spokes", From: spokes, To: spokes, Action: proto.PolicyAllow},
	}}
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

// ContainerSubnet is the first container /24 of the default mesh.
const ContainerSubnet = "10.77.128.0/24"

// Containers is a node's container network at ContainerSubnet (amendment
// A37) in mode, proto.ModeHost or proto.ModeContainer.
func Containers(mode string) *proto.Containers {
	gateway := "10.77.128.1"
	if mode == proto.ModeContainer {
		gateway = "10.77.128.2"
	}
	return &proto.Containers{Mode: mode, Subnet: ContainerSubnet, Network: MeshCIDR, Range: "10.77.128.128/25",
		Bridge: "10.77.128.1", Gateway: gateway}
}

// DualSpokeConfig is SpokeConfig on the dual-stack mesh (M8): both mesh
// addresses, every peer routing both, vps-2 with its IPv6 endpoint as the
// fallback; an exit adds 0.0.0.0/0, and ::/0 when the exit is vps-2 (the
// one with public IPv6).
func DualSpokeConfig(exit string) proto.NodeConfig {
	c := SpokeConfig(exit)
	c.Interface.Address6 = Laptop.MeshIP6 + "/128"
	for i, n := range []Node{VPS1, VPS2, VPS3} {
		p := &c.Peers[i]
		p.AllowedIPs = append(p.AllowedIPs, n.MeshIP6+"/128")
		if n.Name == VPS1.Name {
			p.AllowedIPs = append(p.AllowedIPs, NAS.MeshIP6+"/128")
		}
		if n.Name == exit && n.Endpoint6 != "" {
			p.AllowedIPs = append(p.AllowedIPs, "::/0")
		}
		p.Fallback = n.Endpoint6
	}
	return c
}

// ContainerSubnet6 is the first container /112 of MeshCIDR6 (M8).
const ContainerSubnet6 = "fd10:77::8000:0/112"

// Containers6 is Containers with the IPv6 side (M8).
func Containers6(mode string) *proto.Containers {
	k := Containers(mode)
	k.Subnet6, k.Network6, k.Range6, k.Bridge6 = ContainerSubnet6, MeshCIDR6, "fd10:77::8000:8000/113", "fd10:77::8000:1"
	k.Gateway6 = k.Bridge6
	if mode == proto.ModeContainer {
		k.Gateway6 = "fd10:77::8000:2"
	}
	return k
}

// ContainerSpokeConfig is the laptop's config in container mode: no exit,
// no firewall section (the gateway isolates by policy routing), and the
// container network's forwarding and proxy ARP.
func ContainerSpokeConfig() proto.NodeConfig {
	c := SpokeConfig("")
	c.Exit = nil
	c.Containers = Containers(proto.ModeContainer)
	c.Sysctl = []proto.Sysctl{{Key: proto.SysctlProxyARP, Value: "1"}, {Key: proto.SysctlIPForward, Value: "1"}}
	c.DNS = &proto.DNS{Zone: proto.DefaultZone, Split: true, ForwardFor: []string{ContainerSubnet}}
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

// MeshNames are container names as the server publishes them (amendment
// A38), sorted by name, node and address: the NAS's db, and a web service
// with two containers on the NAS and one on the laptop.
func MeshNames() []proto.MeshName {
	return []proto.MeshName{
		{Name: "db", Node: NAS.Name, IP: "10.77.128.131"},
		{Name: "web", Node: Laptop.Name, IP: "10.77.129.130"},
		{Name: "web", Node: NAS.Name, IP: "10.77.128.129"},
		{Name: "web", Node: NAS.Name, IP: "10.77.128.130"},
	}
}
