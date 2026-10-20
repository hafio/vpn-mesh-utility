package containers_test

import (
	"errors"
	"net/netip"
	"slices"
	"strings"
	"testing"

	"github.com/hafio/vpn-mesh-utility/internal/agent/agentfake"
	"github.com/hafio/vpn-mesh-utility/internal/agent/containers"
	"github.com/hafio/vpn-mesh-utility/internal/proto"
	"github.com/hafio/vpn-mesh-utility/internal/testfix"
)

var (
	subnet = netip.MustParsePrefix(testfix.ContainerSubnet)
	mesh   = netip.MustParsePrefix(testfix.MeshCIDR)
	bridge = netip.MustParsePrefix("10.77.128.1/32")
)

// The gateway's rules, exactly as the M6 trial proved them (R0-R7), with
// the relay rule on a hub that relays only.
func TestGatewayRules(t *testing.T) {
	k := *testfix.Containers(proto.ModeContainer)
	got, err := containers.Gateway(k, "wg0", "eth1")
	if err != nil {
		t.Fatal(err)
	}
	m, b := containers.MainTable, containers.BlackholeTable
	want := []containers.Rule{
		{Priority: 1000, Iif: "lo", Table: m},
		{Priority: 1001, Iif: "wg0", Dst: bridge, Table: b},
		{Priority: 1002, Iif: "wg0", Dst: subnet, Table: m},
		{Priority: 1004, Iif: "eth1", Src: bridge, Table: b},
		{Priority: 1005, Iif: "eth1", Src: subnet, Dst: mesh, Table: m},
		{Priority: 1006, Table: b},
	}
	if !slices.Equal(got, want) {
		t.Fatalf("Gateway =\n%v\nwant\n%v", got, want)
	}
	k.Relay = true
	got, _ = containers.Gateway(k, "wg0", "eth1")
	if relay := (containers.Rule{Priority: 1003, Iif: "wg0", Dst: mesh, Table: m}); len(got) != 7 || got[3] != relay {
		t.Fatalf("a relaying hub's rules = %v", got)
	}
	for _, r := range got {
		if !containers.Owned(r.Priority) {
			t.Errorf("%s is not at an owned priority", r)
		}
	}
	if containers.Owned(32765) || containers.Owned(999) || containers.Owned(5207) {
		t.Error("Owned claims a priority that is not the package's")
	}
}

// The host-mode guard: the subnet is reached from wg0, the bridge and the
// host, and blackholed from anywhere else (the trial's G1-G5).
func TestGuardRules(t *testing.T) {
	got, err := containers.Guard(*testfix.Containers(proto.ModeHost), "wg0")
	if err != nil {
		t.Fatal(err)
	}
	m := containers.MainTable
	want := []containers.Rule{
		{Priority: 5200, Iif: "wg0", Dst: subnet, Table: m},
		{Priority: 5201, Iif: proto.ContainerBridge, Table: m},
		{Priority: 5202, Iif: "lo", Dst: subnet, Table: m},
		{Priority: 5203, Dst: subnet, Table: containers.BlackholeTable},
	}
	if !slices.Equal(got, want) {
		t.Fatalf("Guard =\n%v\nwant\n%v", got, want)
	}
	if s := got[3].String(); s != "rule 5203: to 10.77.128.0/24 lookup 51821" {
		t.Errorf("String = %q", s)
	}
}

// Container mode: the gateway claims .2 on the interface the runtime gave
// an address of the subnet, the blackhole route comes before the rules, a
// stale rule at an owned priority goes, and a second Sync changes nothing.
func TestSyncContainerMode(t *testing.T) {
	n := agentfake.NewNet()
	n.Attach("eth0", netip.MustParsePrefix("172.20.0.3/16"))
	n.Attach("eth1", netip.MustParsePrefix("10.77.128.128/16"))
	n.Attach("wg0", netip.MustParsePrefix("10.77.2.1/32"))
	stale := containers.Rule{Priority: 1003, Iif: "wg0", Dst: mesh, Table: containers.MainTable}
	n.Install(stale)
	k := testfix.Containers(proto.ModeContainer)
	c, err := containers.Sync(n, &agentfake.Sysctl{}, k, "wg0")
	if err != nil || c != 1+1+6+1 {
		t.Fatalf("Sync = %d, %v; calls %v", c, err, n.Calls)
	}
	if !slices.Contains(n.AddrsOf("eth1"), netip.MustParsePrefix("10.77.128.2/16")) {
		t.Errorf("eth1 = %v, want the gateway 10.77.128.2/16 too", n.AddrsOf("eth1"))
	}
	if n.Calls[1] != "blackhole true" {
		t.Errorf("the blackhole route must come before the rules: %v", n.Calls)
	}
	if rules, _ := n.Rules(); slices.Contains(rules, stale) || len(rules) != 6 {
		t.Errorf("rules after Sync = %v", rules)
	}
	if c, err := containers.Sync(n, &agentfake.Sysctl{}, k, "wg0"); err != nil || c != 0 {
		t.Fatalf("second Sync = %d, %v", c, err)
	}
	// Taken away: rules first, then the route; the addresses stay.
	n.Calls = nil
	if c, err := containers.Sync(n, &agentfake.Sysctl{}, nil, "wg0"); err != nil || c != 7 {
		t.Fatalf("Sync(nil) = %d, %v", c, err)
	}
	if last := n.Calls[len(n.Calls)-1]; last != "blackhole false" || n.HasBlackhole() {
		t.Errorf("the route must go last: %v", n.Calls)
	}
	if c, err := containers.Sync(n, &agentfake.Sysctl{}, nil, "wg0"); err != nil || c != 0 {
		t.Fatalf("Sync(nil) again = %d, %v", c, err)
	}
}

// The agent not on the mesh network is a degraded step that says what to
// do, and changes nothing.
func TestSyncWithoutTheMeshNetwork(t *testing.T) {
	n := agentfake.NewNet()
	n.Attach("eth0", netip.MustParsePrefix("172.20.0.3/16"))
	n.Attach("wg0", netip.MustParsePrefix("10.77.128.9/32")) // the tunnel is skipped
	_, err := containers.Sync(n, &agentfake.Sysctl{}, testfix.Containers(proto.ModeContainer), "wg0")
	if err == nil || !strings.Contains(err.Error(), "not on the meshctl-mesh network") || !strings.Contains(err.Error(), "meshctl node compose") {
		t.Fatalf("Sync = %v", err)
	}
	if len(n.Calls) != 0 {
		t.Fatalf("calls %v", n.Calls)
	}
}

// Host mode: the agent makes the bridge, gives it .1, turns on its
// forwarding and proxy ARP and wg0's forwarding, then the guard.
func TestSyncHostMode(t *testing.T) {
	n, sys := agentfake.NewNet(), &agentfake.Sysctl{}
	k := testfix.Containers(proto.ModeHost)
	c, err := containers.Sync(n, sys, k, "wg0")
	if err != nil || c != 1+1+3+1+4 {
		t.Fatalf("Sync = %d, %v; calls %v", c, err, n.Calls)
	}
	if got := n.AddrsOf(proto.ContainerBridge); !slices.Equal(got, []netip.Prefix{netip.MustParsePrefix("10.77.128.1/16")}) {
		t.Errorf("%s = %v", proto.ContainerBridge, got)
	}
	for _, key := range []string{proto.SysctlBridgeForwarding, proto.SysctlBridgeProxyARP, proto.SysctlWGForwarding} {
		if sys.Values[key] != "1" {
			t.Errorf("sysctl %s was not set: %v", key, sys.Values)
		}
	}
	if c, err := containers.Sync(n, sys, k, "wg0"); err != nil || c != 0 {
		t.Fatalf("second Sync = %d, %v", c, err)
	}
}

// Every failure names the step and what failed; nothing after it runs.
func TestSyncFailures(t *testing.T) {
	boom := errors.New("operation not permitted")
	for _, c := range []struct {
		failOn string
		mode   string
		want   string
	}{
		{"EnsureBridge", proto.ModeHost, "bridge meshctl0"},
		{"AddAddr", proto.ModeHost, "add 10.77.128.1/16 to meshctl0"},
		{"SetBlackhole", proto.ModeHost, "operation not permitted"},
		{"AddRule", proto.ModeContainer, "operation not permitted"},
		{"LinkHolding", proto.ModeContainer, "operation not permitted"},
		{"Rules", proto.ModeContainer, "operation not permitted"},
		{"Blackhole", proto.ModeContainer, "operation not permitted"},
		{"Addrs", proto.ModeContainer, "addresses of eth1"},
		{"AddAddr", proto.ModeContainer, "add 10.77.128.2/16 to eth1"},
	} {
		n := agentfake.NewNet()
		n.Attach("eth1", netip.MustParsePrefix("10.77.128.128/16"))
		n.Fail, n.FailOn = boom, c.failOn
		_, err := containers.Sync(n, &agentfake.Sysctl{}, testfix.Containers(c.mode), "wg0")
		if err == nil || !strings.HasPrefix(err.Error(), "containers: ") || !strings.Contains(err.Error(), c.want) {
			t.Errorf("%s: Sync = %v, want %q", c.failOn, err, c.want)
		}
	}
	// A stale rule that will not go, and a blackhole route that will not.
	n := agentfake.NewNet()
	n.Attach("eth1", netip.MustParsePrefix("10.77.128.128/16"))
	n.Install(containers.Rule{Priority: 1003, Iif: "wg0", Dst: mesh, Table: containers.MainTable})
	n.Fail, n.FailOn = boom, "DelRule"
	if _, err := containers.Sync(n, &agentfake.Sysctl{}, testfix.Containers(proto.ModeContainer), "wg0"); err == nil ||
		!strings.Contains(err.Error(), "operation not permitted") {
		t.Errorf("a stale rule that will not go = %v", err)
	}
	n.FailOn = ""
	n.Fail = nil
	if _, err := containers.Sync(n, &agentfake.Sysctl{}, testfix.Containers(proto.ModeContainer), "wg0"); err != nil {
		t.Fatal(err)
	}
	n.Fail, n.FailOn = boom, "SetBlackhole"
	if _, err := containers.Sync(n, &agentfake.Sysctl{}, nil, "wg0"); err == nil || !strings.Contains(err.Error(), "operation not permitted") {
		t.Errorf("a blackhole route that will not go = %v", err)
	}
	sysFail := &agentfake.Sysctl{Err: boom}
	if _, err := containers.Sync(agentfake.NewNet(), sysFail, testfix.Containers(proto.ModeHost), "wg0"); err == nil ||
		!strings.Contains(err.Error(), "operation not permitted") {
		t.Errorf("a failed sysctl = %v", err)
	}
	k := testfix.Containers(proto.ModeContainer)
	k.Mode = "bridge"
	if _, err := containers.Sync(agentfake.NewNet(), sysFail, k, "wg0"); err == nil || !strings.Contains(err.Error(), "unknown mode") {
		t.Errorf("an unknown mode = %v", err)
	}
	k = testfix.Containers(proto.ModeContainer)
	k.Subnet = "10.77.128.0"
	if _, err := containers.Sync(agentfake.NewNet(), sysFail, k, "wg0"); err == nil {
		t.Error("a malformed subnet was accepted")
	}
	if _, err := containers.Guard(*k, "wg0"); err == nil {
		t.Error("Guard accepted a malformed subnet")
	}
	if _, err := containers.Gateway(*k, "wg0", "eth1"); err == nil {
		t.Error("Gateway accepted a malformed subnet")
	}
}
