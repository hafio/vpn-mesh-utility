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
