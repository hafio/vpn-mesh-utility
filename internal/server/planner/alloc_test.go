package planner

import (
	"net/netip"
	"strings"
	"testing"

	"github.com/hafio/vpn-mesh-utility/internal/proto"
)

func TestRangeFor(t *testing.T) {
	cases := []struct {
		role     proto.Role
		alwaysOn bool
		want     Range
	}{
		{proto.RoleServer, false, RangeServers},
		{proto.RoleServer, true, RangeServers},
		{proto.RoleSpoke, false, RangeLaptops},
		{proto.RoleSpoke, true, RangeAlwaysOn},
		{proto.RolePhone, false, RangePhones},
	}
	for _, c := range cases {
		if got := RangeFor(c.role, c.alwaysOn); got != c.want {
			t.Errorf("RangeFor(%s, %v) = %d, want %d", c.role, c.alwaysOn, got, c.want)
		}
	}
}

func TestBlockMatchesTheSpecAddressPlan(t *testing.T) {
	mesh := netip.MustParsePrefix("10.77.0.0/16")
	want := map[Range]string{RangeServers: "10.77.0.0/24", RangeAlwaysOn: "10.77.1.0/24", RangeLaptops: "10.77.2.0/24", RangePhones: "10.77.3.0/24"}
	for r, w := range want {
		if got := Block(mesh, r).String(); got != w {
			t.Errorf("Block(%d) = %s, want %s", r, got, w)
		}
	}
	if got := Block(netip.MustParsePrefix("172.20.4.0/22"), RangePhones).String(); got != "172.20.7.0/24" {
		t.Errorf("a /22 mesh puts phones in %s, want 172.20.7.0/24", got)
	}
}

func TestAllocateLowestFree(t *testing.T) {
	mesh := netip.MustParsePrefix("10.77.0.0/16")
	used := map[netip.Addr]bool{netip.MustParseAddr("10.77.2.1"): true, netip.MustParseAddr("10.77.2.3"): true}
	got, err := Allocate(mesh, RangeLaptops, used)
	if err != nil || got.String() != "10.77.2.2" {
		t.Fatalf("Allocate = %v, %v; want 10.77.2.2", got, err)
	}
	full := map[netip.Addr]bool{}
	for i := 1; i <= 254; i++ {
		full[netip.AddrFrom4([4]byte{10, 77, 3, byte(i)})] = true
	}
	if _, err := Allocate(mesh, RangePhones, full); err == nil || !strings.Contains(err.Error(), "full") {
		t.Fatalf("Allocate in a full block = %v, want a 'full' error", err)
	}
	if _, err := Allocate(netip.MustParsePrefix("10.0.0.0/8"), RangePhones, nil); err == nil {
		t.Fatal("Allocate accepted an unsupported mesh size")
	}
}

// Container /24s come from the upper half of the mesh (amendment A37): the
// lowest free one, past the ones taken and the ones holding a node's IP;
// a /22 has no room.
func TestAllocateContainers(t *testing.T) {
	mesh := netip.MustParsePrefix("10.77.0.0/16")
	got, err := AllocateContainers(mesh, nil, nil)
	if err != nil || got.String() != "10.77.128.0/24" {
		t.Fatalf("first = %v, %v", got, err)
	}
	used := map[netip.Prefix]bool{netip.MustParsePrefix("10.77.128.0/24"): true}
	ips := map[netip.Addr]bool{netip.MustParseAddr("10.77.129.7"): true}
	if got, err = AllocateContainers(mesh, used, ips); err != nil || got.String() != "10.77.130.0/24" {
		t.Errorf("past a used block and a fixed IP = %v, %v", got, err)
	}
	if got, err = AllocateContainers(netip.MustParsePrefix("10.77.0.0/21"), nil, nil); err != nil || got.String() != "10.77.4.0/24" {
		t.Errorf("a /21 = %v, %v", got, err)
	}
	if _, err = AllocateContainers(netip.MustParsePrefix("10.77.0.0/22"), nil, nil); err == nil || !strings.Contains(err.Error(), "/21 or larger") {
		t.Errorf("a /22: %v", err)
	}
	full := map[netip.Prefix]bool{}
	for r := 4; r < 8; r++ {
		full[Block(netip.MustParsePrefix("10.77.0.0/21"), Range(r))] = true
	}
	if _, err = AllocateContainers(netip.MustParsePrefix("10.77.0.0/21"), full, nil); err == nil || !strings.Contains(err.Error(), "is taken") {
		t.Errorf("all taken: %v", err)
	}
	if _, err = AllocateContainers(netip.MustParsePrefix("10.0.0.0/8"), nil, nil); err == nil {
		t.Error("a mesh outside /16-/22 was accepted")
	}
}

func TestCheckRequested(t *testing.T) {
	mesh := netip.MustParsePrefix("10.77.0.0/16")
	used := map[netip.Addr]bool{netip.MustParseAddr("10.77.1.1"): true}
	cases := []struct {
		ip   string
		want string
	}{
		{"10.77.1.2", ""},
		{"10.78.0.1", "outside"},
		{"10.77.1.0", "network or broadcast"},
		{"10.77.1.255", "network or broadcast"},
		{"10.77.1.1", "already assigned"},
	}
	for _, c := range cases {
		err := CheckRequested(mesh, netip.MustParseAddr(c.ip), used)
		if c.want == "" && err != nil {
			t.Errorf("CheckRequested(%s) = %v", c.ip, err)
		}
		if c.want != "" && (err == nil || !strings.Contains(err.Error(), c.want)) {
			t.Errorf("CheckRequested(%s) = %v, want %q", c.ip, err, c.want)
		}
	}
}

func TestCheckMeshCIDR(t *testing.T) {
	for _, s := range []string{"10.77.0.0/16", "10.77.0.0/22", "192.168.64.0/18"} {
		if err := CheckMeshCIDR(netip.MustParsePrefix(s)); err != nil {
			t.Errorf("CheckMeshCIDR(%s): %v", s, err)
		}
	}
	for _, s := range []string{"10.0.0.0/8", "10.77.0.0/24", "fd00::/48", "10.77.0.1/16"} {
		if err := CheckMeshCIDR(netip.MustParsePrefix(s)); err == nil {
			t.Errorf("CheckMeshCIDR(%s) accepted it", s)
		}
	}
}

// The IPv6 plan (M8): the role blocks keep their IPv4 numbers in the low
// 32 bits, container networks take the blocks from 0x8000.
func TestIPv6AddressPlan(t *testing.T) {
	mesh6 := netip.MustParsePrefix("fd10:77::/64")
	for r, want := range map[Range]string{RangeServers: "fd10:77::/112", RangeAlwaysOn: "fd10:77::1:0/112",
		RangeLaptops: "fd10:77::2:0/112", RangePhones: "fd10:77::3:0/112"} {
		if got := Block6(mesh6, int(r)).String(); got != want {
			t.Errorf("Block6(%d) = %s, want %s", r, got, want)
		}
	}
	used := map[netip.Addr]bool{netip.MustParseAddr("fd10:77::1"): true}
	if ip, err := Allocate6(mesh6, RangeServers, used); err != nil || ip.String() != "fd10:77::2" {
		t.Errorf("Allocate6 = %v, %v", ip, err)
	}
	if ip, err := Allocate6(mesh6, RangePhones, nil); err != nil || ip.String() != "fd10:77::3:1" {
		t.Errorf("Allocate6 phones = %v, %v", ip, err)
	}
	k, err := AllocateContainers6(mesh6, map[netip.Prefix]bool{netip.MustParsePrefix("fd10:77::8000:0/112"): true})
	if err != nil || k.String() != "fd10:77::8001:0/112" {
		t.Errorf("AllocateContainers6 = %v, %v", k, err)
	}
	for _, c := range []struct {
		ip, want string
	}{
		{"fd10:77::2:9", ""},
		{"fd10:77:0:0:1234:5678:9abc:def0", ""},
		{"fd10:78::1", "outside the mesh"},
		{"fd10:77::", "prefix's own address"},
		{"fd10:77::8000:5", "where the container networks are"},
		{"fd10:77::ffff:1", "where the container networks are"},
		{"fd10:77::1", "already assigned"},
	} {
		err := CheckRequested6(mesh6, netip.MustParseAddr(c.ip), used)
		if (c.want == "") != (err == nil) || (err != nil && !strings.Contains(err.Error(), c.want)) {
			t.Errorf("CheckRequested6(%s) = %v, want %q", c.ip, err, c.want)
		}
	}
	for _, s := range []string{"fd10:77::/48", "fd10:77::/64", "2001:db8:1::/56"} {
		if err := CheckMeshCIDR6(netip.MustParsePrefix(s)); err != nil {
			t.Errorf("CheckMeshCIDR6(%s) = %v", s, err)
		}
	}
	for _, s := range []string{"fd10::/32", "fd10:77::/80", "fd10:77::1/64", "10.77.0.0/16", "::ffff:10.0.0.0/112"} {
		if err := CheckMeshCIDR6(netip.MustParsePrefix(s)); err == nil {
			t.Errorf("CheckMeshCIDR6(%s) accepted", s)
		}
	}
	if _, err := Allocate6(netip.MustParsePrefix("fd10::/32"), RangeServers, nil); err == nil {
		t.Error("Allocate6 accepted an unsupported prefix")
	}
	if _, err := AllocateContainers6(netip.Prefix{}, nil); err == nil {
		t.Error("AllocateContainers6 accepted no prefix")
	}
}
