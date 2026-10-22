package apply

import (
	"slices"
	"testing"

	"github.com/hafio/vpn-mesh-utility/internal/proto"
	"github.com/hafio/vpn-mesh-utility/internal/testfix"
)

// forHost leaves every IPv6 part out of a config on a host without IPv6
// (M8), and the caller's config as it was.
func TestForHostWithoutIPv6(t *testing.T) {
	cfg := testfix.ServerConfig()
	cfg.Interface.Address6 = testfix.VPS1.MeshIP6 + "/128"
	cfg.Peers[0].AllowedIPs = append(cfg.Peers[0].AllowedIPs, testfix.VPS2.MeshIP6+"/128")
	cfg.Peers[0].Endpoint, cfg.Peers[0].Fallback = testfix.VPS2.Endpoint6, testfix.VPS2.Endpoint
	cfg.Peers[1].Fallback = "[2001:db8::12]:51820"
	cfg.Sysctl = []proto.Sysctl{{Key: proto.SysctlIPForward, Value: "1"}, {Key: proto.SysctlIPv6Forward, Value: "1"}}
	fw := *cfg.Firewall
	fw.Exit6 = testfix.MeshCIDR6
	cfg.Firewall = &fw
	cfg.Containers = testfix.Containers6(proto.ModeHost)

	if same, err := forHost(cfg, true); err != nil || same.Interface.Address6 == "" || same.Firewall.Exit6 == "" {
		t.Fatalf("a host with IPv6 carries all of it: %v", err)
	}
	got, err := forHost(cfg, false)
	if err != nil {
		t.Fatal(err)
	}
	if got.Interface.Address6 != "" || got.Interface.Address != cfg.Interface.Address {
		t.Errorf("interface %+v", got.Interface)
	}
	if p := got.Peers[0]; p.Endpoint != testfix.VPS2.Endpoint || p.Fallback != "" || slices.Contains(p.AllowedIPs, testfix.VPS2.MeshIP6+"/128") {
		t.Errorf("vps-2's peer %+v", p)
	}
	if p := got.Peers[1]; p.Endpoint != cfg.Peers[1].Endpoint || p.Fallback != "" {
		t.Errorf("vps-3's peer %+v", p)
	}
	if len(got.Sysctl) != 1 || got.Sysctl[0].Key != proto.SysctlIPForward {
		t.Errorf("sysctls %+v", got.Sysctl)
	}
	if got.Firewall.Exit6 != "" || got.Containers.Subnet6 != "" || got.Containers.Gateway6 != "" || got.Containers.Subnet == "" {
		t.Errorf("firewall %+v, containers %+v", got.Firewall, got.Containers)
	}
	if cfg.Interface.Address6 == "" || cfg.Peers[0].Fallback == "" || len(cfg.Sysctl) != 2 || cfg.Firewall.Exit6 == "" || cfg.Containers.Subnet6 == "" {
		t.Error("forHost changed the caller's config")
	}
}
