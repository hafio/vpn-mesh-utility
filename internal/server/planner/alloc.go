package planner

import (
	"encoding/binary"
	"fmt"
	"net/netip"

	"github.com/hafio/vpn-mesh-utility/internal/proto"
)

// Range is one of the address blocks of the mesh (spec S4 address plan).
// Block r is the r-th /24 of the mesh CIDR, so the default 10.77.0.0/16 gives
// servers 10.77.0.x, always-on spokes 10.77.1.x, laptops 10.77.2.x and phones
// 10.77.3.x, and a /22 mesh fits the same four blocks exactly.
type Range int

const (
	RangeServers Range = iota
	RangeAlwaysOn
	RangeLaptops
	RangePhones
)

// RangeFor is the default block for a role. Spokes default to the laptop
// block; always-on spokes (the NAS) are placed in RangeAlwaysOn on request.
func RangeFor(role proto.Role, alwaysOn bool) Range {
	switch role {
	case proto.RoleServer:
		return RangeServers
	case proto.RolePhone:
		return RangePhones
	}
	if alwaysOn {
		return RangeAlwaysOn
	}
	return RangeLaptops
}

// CheckMeshCIDR enforces the supported mesh sizes: an IPv4 /16 to /22, large
// enough for the four /24 blocks.
func CheckMeshCIDR(p netip.Prefix) error {
	if !p.IsValid() || !p.Addr().Is4() || p.Bits() < 16 || p.Bits() > 22 || p.Masked() != p {
		return fmt.Errorf("mesh CIDR %s must be an IPv4 /16 to /22 network (default 10.77.0.0/16)", p)
	}
	return nil
}

// Block returns the /24 of range r inside mesh.
func Block(mesh netip.Prefix, r Range) netip.Prefix {
	a := mesh.Addr().As4()
	v := binary.BigEndian.Uint32(a[:]) + uint32(r)*256
	binary.BigEndian.PutUint32(a[:], v)
	return netip.PrefixFrom(netip.AddrFrom4(a), 24)
}

// Allocate returns the lowest free host (.1 to .254) in range r.
func Allocate(mesh netip.Prefix, r Range, used map[netip.Addr]bool) (netip.Addr, error) {
	if err := CheckMeshCIDR(mesh); err != nil {
		return netip.Addr{}, err
	}
	block := Block(mesh, r)
	a := block.Addr().As4()
	for host := 1; host <= 254; host++ {
		a[3] = byte(host)
		ip := netip.AddrFrom4(a)
		if !used[ip] {
			return ip, nil
		}
	}
	return netip.Addr{}, fmt.Errorf("address block %s is full (254 nodes); pass a fixed IP elsewhere in %s", block, mesh)
}

// AllocateContainers returns the lowest free /24 for a node's containers
// (amendment A37). They come from the upper half of the mesh's /24 blocks,
// so the role blocks below never meet them; a /22 has no half to spare.
// used holds the subnets taken, ips the mesh IPs in use (a fixed IP may sit
// anywhere in the mesh).
func AllocateContainers(mesh netip.Prefix, used map[netip.Prefix]bool, ips map[netip.Addr]bool) (netip.Prefix, error) {
	if err := CheckMeshCIDR(mesh); err != nil {
		return netip.Prefix{}, err
	}
	blocks := 1 << (24 - mesh.Bits())
	if blocks < 8 {
		return netip.Prefix{}, fmt.Errorf("the mesh %s is too small for container networks: they take /24s from its upper half, which needs a /21 or larger", mesh)
	}
next:
	for r := blocks / 2; r < blocks; r++ {
		b := Block(mesh, Range(r))
		if used[b] {
			continue
		}
		for ip := range ips {
			if b.Contains(ip) {
				continue next
			}
		}
		return b, nil
	}
	return netip.Prefix{}, fmt.Errorf("every container /24 of %s is taken (%d of them); remove a node's container network or use a larger mesh", mesh, blocks/2)
}

// CheckRequested validates an operator-chosen mesh IP: inside the mesh, not
// the network or broadcast address of its /24, and not taken.
func CheckRequested(mesh netip.Prefix, ip netip.Addr, used map[netip.Addr]bool) error {
	if !mesh.Contains(ip) {
		return fmt.Errorf("mesh IP %s is outside the mesh %s", ip, mesh)
	}
	if last := ip.As4()[3]; last == 0 || last == 255 {
		return fmt.Errorf("mesh IP %s is the network or broadcast address of its /24; pick .1 to .254", ip)
	}
	if used[ip] {
		return fmt.Errorf("mesh IP %s is already assigned to another node", ip)
	}
	return nil
}

// The IPv6 address plan (M8). The mesh's IPv6 prefix is a /48 to /64, and
// only the low 32 bits of an address are planned: block r is the /112
// <prefix>::r:0, so the roles keep their IPv4 block numbers (servers
// <prefix>::1, always-on spokes <prefix>::1:1, laptops <prefix>::2:1,
// phones <prefix>::3:1) and container networks take the blocks from
// 0x8000 on, the upper half, as they do in IPv4. A node's IPv6 address is
// allocated on its own; it does not follow its IPv4 one.

// containerBlock6 is the first IPv6 block of the container networks.
const containerBlock6 = 0x8000

// CheckMeshCIDR6 enforces the supported IPv6 prefixes: a /48 to /64,
// without host bits; a unique local (fd00::/8) prefix is the usual choice.
func CheckMeshCIDR6(p netip.Prefix) error {
	if !p.IsValid() || !p.Addr().Is6() || p.Addr().Is4In6() || p.Bits() < 48 || p.Bits() > 64 || p.Masked() != p {
		return fmt.Errorf("IPv6 mesh prefix %s must be an IPv6 /48 to /64 network, such as a unique local fdXX:XXXX:XXXX::/64", p)
	}
	return nil
}

// Block6 returns IPv6 block r (a /112) of mesh6.
func Block6(mesh6 netip.Prefix, r int) netip.Prefix {
	a := mesh6.Addr().As16()
	binary.BigEndian.PutUint16(a[12:14], uint16(r))
	a[14], a[15] = 0, 0
	return netip.PrefixFrom(netip.AddrFrom16(a), 112)
}

// containers6 is the part of mesh6 the container networks take.
func containers6(mesh6 netip.Prefix) netip.Prefix {
	return netip.PrefixFrom(Block6(mesh6, containerBlock6).Addr(), 97)
}

// Allocate6 returns the lowest free IPv6 address (host 1 to 65534) in block
// r of mesh6.
func Allocate6(mesh6 netip.Prefix, r Range, used map[netip.Addr]bool) (netip.Addr, error) {
	if err := CheckMeshCIDR6(mesh6); err != nil {
		return netip.Addr{}, err
	}
	block := Block6(mesh6, int(r))
	a := block.Addr().As16()
	for host := 1; host <= 65534; host++ {
		binary.BigEndian.PutUint16(a[14:16], uint16(host))
		ip := netip.AddrFrom16(a)
		if !used[ip] {
			return ip, nil
		}
	}
	return netip.Addr{}, fmt.Errorf("IPv6 address block %s is full; pass a fixed IPv6 address elsewhere in %s", block, mesh6)
}

// AllocateContainers6 returns the lowest free /112 for a node's containers'
// IPv6 (M8) from the container blocks of mesh6.
func AllocateContainers6(mesh6 netip.Prefix, used map[netip.Prefix]bool) (netip.Prefix, error) {
	if err := CheckMeshCIDR6(mesh6); err != nil {
		return netip.Prefix{}, err
	}
	for r := containerBlock6; r <= 0xffff; r++ {
		if b := Block6(mesh6, r); !used[b] {
			return b, nil
		}
	}
	return netip.Prefix{}, fmt.Errorf("every container /112 of %s is taken", mesh6)
}

// CheckRequested6 validates an operator-chosen IPv6 mesh address: inside
// mesh6, not the prefix's own (subnet-router) address, outside the
// container networks' part, and not taken.
func CheckRequested6(mesh6 netip.Prefix, ip netip.Addr, used map[netip.Addr]bool) error {
	switch {
	case !mesh6.Contains(ip):
		return fmt.Errorf("IPv6 mesh address %s is outside the mesh %s", ip, mesh6)
	case ip == mesh6.Addr():
		return fmt.Errorf("IPv6 mesh address %s is the prefix's own address; pick a host in it", ip)
	case containers6(mesh6).Contains(ip):
		return fmt.Errorf("IPv6 mesh address %s is in %s, where the container networks are", ip, containers6(mesh6))
	case used[ip]:
		return fmt.Errorf("IPv6 mesh address %s is already assigned to another node", ip)
	}
	return nil
}
