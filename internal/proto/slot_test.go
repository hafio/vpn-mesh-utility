package proto

import (
	"slices"
	"strconv"
	"testing"
)

// Slot 0 keeps the names of a single mesh, and no two slots share a name,
// a port, a table or a rule (amendment A49).
func TestSlotNames(t *testing.T) {
	var zero Slot
	if zero.Interface() != "wg0" || zero.ListenPort() != DefaultListenPort || zero.RouteTable() != RouteTable ||
		zero.ContainersTable() != RouteTable+1 || zero.Bridge() != ContainerBridge || zero.Network() != MeshNetwork ||
		zero.Zone() != ZoneMesh || zero.FirewalldPolicy("relay") != "meshctl-relay" || zero.Table() != "meshctl" ||
		zero.UIPort() != 7443 || zero.RuleShift() != 0 {
		t.Fatal("slot 0's names are not a single mesh's")
	}
	if !slices.Equal(SysctlKeys(0), []string{SysctlIPForward, SysctlSrcValidMark, SysctlProxyARP, SysctlBridgeForwarding,
		SysctlBridgeProxyARP, SysctlWGForwarding, SysctlIPv6Forward, SysctlProxyNDP, SysctlBridgeProxyNDP}) {
		t.Fatalf("slot 0's sysctl keys %v", SysctlKeys(0))
	}

	s := Slot(5)
	got := []string{s.Interface(), s.Bridge(), s.Network(), s.Zone(), s.FirewalldPolicy("exit"), s.Table(),
		strconv.Itoa(s.ListenPort()), strconv.Itoa(s.RouteTable()), strconv.Itoa(s.ContainersTable()), strconv.Itoa(s.UIPort()),
		s.SysctlWGForwarding(), s.SysctlBridgeProxyNDP()}
	want := []string{"wg5", "meshctl5", "meshctl-mesh5", "mesh5", "meshctl5-exit", "meshctl5", "51825", "51830", "51831", "7448",
		"net.ipv4.conf.wg5.forwarding", "net.ipv6.conf.meshctl5.proxy_ndp"}
	if !slices.Equal(got, want) {
		t.Fatalf("slot 5's names %q, want %q", got, want)
	}
	if !slices.Contains(SysctlKeys(5), "net.ipv4.conf.meshctl5.proxy_arp") || slices.Contains(SysctlKeys(5), SysctlWGForwarding) {
		t.Fatalf("slot 5's sysctl keys %v", SysctlKeys(5))
	}

	seen := map[string]Slot{}
	for s := Slot(0); s <= MaxSlot; s++ {
		for _, name := range []string{"if " + s.Interface(), "br " + s.Bridge(), "net " + s.Network(), "zone " + s.Zone(),
			"policy " + s.FirewalldPolicy("relay"), "nft " + s.Table(), "port " + strconv.Itoa(s.ListenPort()),
			"table " + strconv.Itoa(s.RouteTable()), "table " + strconv.Itoa(s.ContainersTable()), "ui " + strconv.Itoa(s.UIPort())} {
			if other, ok := seen[name]; ok {
				t.Fatalf("slots %d and %d share %s", other, s, name)
			}
			seen[name] = s
		}
		if len(s.Interface()) > 15 || len(s.Bridge()) > 15 || len(s.Zone()) > 17 {
			t.Fatalf("slot %d's names are too long for the kernel or firewalld", s)
		}
	}
	if Slot(-1).Valid() || !Slot(MaxSlot).Valid() || Slot(MaxSlot+1).Valid() {
		t.Fatal("Valid is wrong at the edges")
	}
}
