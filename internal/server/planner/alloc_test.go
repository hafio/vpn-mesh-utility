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
