package names

import (
	"fmt"
	"net/netip"
	"slices"
	"testing"

	"github.com/hafio/vpn-mesh-utility/internal/proto"
	"github.com/hafio/vpn-mesh-utility/internal/testfix"
)

const (
	id1 = "aaaaaaaaaaaa1111111111111111111111111111111111111111111111111111"
	id2 = "bbbbbbbbbbbb2222222222222222222222222222222222222222222222222222"
	id3 = "cccccccccccc3333333333333333333333333333333333333333333333333333"
)

var subnet = netip.MustParsePrefix(testfix.ContainerSubnet)

// Every kind of name a runtime gives, each once per address; the ID
// defaults, invalid names and the addresses that are not a container's
// are left out.
func TestOf(t *testing.T) {
	cs := []Container{
		{ID: id1, Name: "/shop-web-1", Hostname: id1[:12], IP: netip.MustParseAddr("10.77.128.129"),
			Labels: map[string]string{LabelComposeService: "web"}, Aliases: []string{"web", id1[:12], "Shop-Front"}},
		{ID: id2, Name: "/shop_api.1.x7f", Hostname: "api", Domainname: "example.home.arpa", IP: netip.MustParseAddr("10.77.128.130"),
			Labels: map[string]string{LabelSwarmService: "shop_api", LabelStack: "shop"}},
		{ID: id3, Name: "/meshctl-agent", IP: netip.MustParseAddr("10.77.128.2")}, // the gateway itself
		{ID: id3, Name: "/elsewhere", IP: netip.MustParseAddr("10.77.129.130")},
	}
	got, over := Of(cs, subnet)
	want := []proto.ContainerName{
		{Name: "api", Kind: proto.NameSwarm, IP: "10.77.128.130"},
		{Name: "api.example.home.arpa", Kind: proto.NameHostname, IP: "10.77.128.130"},
		{Name: "shop-front", Kind: proto.NameAlias, IP: "10.77.128.129"},
		{Name: "shop-web-1", Kind: proto.NameContainer, IP: "10.77.128.129"},
		{Name: "web", Kind: proto.NameService, IP: "10.77.128.129"},
	}
	if over != 0 || !slices.Equal(got, want) {
		t.Fatalf("Of =\n%v (over %d)\nwant\n%v", got, over, want)
	}
	if got, _ := Of(cs, netip.Prefix{}); len(got) != 0 {
		t.Fatalf("without a container network = %v", got)
	}
}

// A swarm service outside a stack keeps its name.
func TestOfSwarmWithoutAStack(t *testing.T) {
	got, _ := Of([]Container{{ID: id1, IP: netip.MustParseAddr("10.77.128.129"), Labels: map[string]string{LabelSwarmService: "cache"}}}, subnet)
	if len(got) != 1 || got[0].Name != "cache" || got[0].Kind != proto.NameSwarm {
		t.Fatalf("Of = %v", got)
	}
}

// More names than a report may carry: the first MaxNames by name, and how
// many were left out.
func TestOfCapsTheNames(t *testing.T) {
	c := Container{ID: id1, IP: netip.MustParseAddr("10.77.128.129")}
	for i := range proto.MaxNames + 3 {
		c.Aliases = append(c.Aliases, fmt.Sprintf("n%04d", i))
	}
	got, over := Of([]Container{c}, subnet)
	if len(got) != proto.MaxNames || over != 3 || got[0].Name != "n0000" {
		t.Fatalf("Of = %d names (first %v), over %d", len(got), got[0], over)
	}
}

func TestHostAddr(t *testing.T) {
	for ip, want := range map[string]bool{
		"10.77.128.0": false, "10.77.128.1": false, "10.77.128.2": false, "10.77.128.255": false,
		"10.77.128.3": true, "10.77.128.254": true, "10.77.129.3": false,
	} {
		if got := hostAddr(subnet, netip.MustParseAddr(ip)); got != want {
			t.Errorf("hostAddr(%s) = %v, want %v", ip, got, want)
		}
	}
	if hostAddr(subnet, netip.Addr{}) || hostAddr(netip.Prefix{}, netip.MustParseAddr("10.77.128.3")) {
		t.Error("an invalid address or network holds no container")
	}
}
