package proto

import "strconv"

// Slot is a mesh's number on its control plane, 0..MaxSlot (amendment A49).
// An agent derives from it every name it uses on its host, so the agents of
// several meshes share one host without touching each other's interface,
// routing tables and rules, firewall zones, policies and tables, container
// bridge and network, ports or keys. Slot 0 keeps the names of a single
// mesh: the constants above (ContainerBridge, MeshNetwork, ZoneMesh,
// RouteTable, DefaultListenPort and the sysctl keys) are slot 0's.
//
// What slots share on purpose: FwMark (every mesh's tunnel packets bypass
// whichever mesh's exit carries the host's internet), the public zone, and
// the host-wide sysctl keys.
type Slot int

// MaxSlot is the highest slot, so a control plane holds at most 64 meshes.
const MaxSlot = 63

// Valid reports whether s is a slot.
func (s Slot) Valid() bool { return s >= 0 && s <= MaxSlot }

// n is the slot's number as text, "" for slot 0 where a name has none.
func (s Slot) n() string {
	if s == 0 {
		return ""
	}
	return strconv.Itoa(int(s))
}

// Interface is the WireGuard interface: wg0, wg1, ...
func (s Slot) Interface() string { return "wg" + strconv.Itoa(int(s)) }

// ListenPort is a server's default WireGuard port: 51820, 51821, ...
func (s Slot) ListenPort() int { return DefaultListenPort + int(s) }

// RouteTable holds exit mode's default routes: 51820, 51822, ...
func (s Slot) RouteTable() int { return RouteTable + 2*int(s) }

// ContainersTable holds the container network's blackhole defaults:
// 51821, 51823, ...
func (s Slot) ContainersTable() int { return RouteTable + 1 + 2*int(s) }

// RuleShift is how far the slot's policy rules move from slot 0's: exit
// mode's 10 lower per slot, the container network's 10 higher.
func (s Slot) RuleShift() int { return 10 * int(s) }

// Bridge is a host-mode node's container bridge: meshctl0, meshctl1, ...
func (s Slot) Bridge() string { return "meshctl" + strconv.Itoa(int(s)) }

// Network is the runtime network the node's containers attach to:
// meshctl-mesh, meshctl-mesh1, ...
func (s Slot) Network() string { return MeshNetwork + s.n() }

// Zone is the firewalld zone of the mesh interface: mesh, mesh1, ...
func (s Slot) Zone() string { return ZoneMesh + s.n() }

// FirewalldPolicy is the name of one of the agent's firewalld policies
// (kind "relay" or "exit"): meshctl-relay, meshctl1-relay, ...
func (s Slot) FirewalldPolicy(kind string) string { return "meshctl" + s.n() + "-" + kind }

// Table is the nftables table of a server's mesh firewall (amendment A43):
// meshctl, meshctl1, ...
func (s Slot) Table() string { return "meshctl" + s.n() }

// UIPort is the local UI's default port: 7443, 7444, ...
func (s Slot) UIPort() int { return 7443 + int(s) }

// SysctlWGForwarding is the tunnel's own IPv4 forwarding, which the
// containers step of a host-mode node sets.
func (s Slot) SysctlWGForwarding() string { return "net.ipv4.conf." + s.Interface() + ".forwarding" }

// SysctlBridgeForwarding is the host-mode container bridge's IPv4
// forwarding.
func (s Slot) SysctlBridgeForwarding() string { return "net.ipv4.conf." + s.Bridge() + ".forwarding" }

// SysctlBridgeProxyARP is the host-mode container bridge's proxy ARP.
func (s Slot) SysctlBridgeProxyARP() string { return "net.ipv4.conf." + s.Bridge() + ".proxy_arp" }

// SysctlBridgeProxyNDP is the host-mode container bridge's proxy NDP.
func (s Slot) SysctlBridgeProxyNDP() string { return "net.ipv6.conf." + s.Bridge() + ".proxy_ndp" }
