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
	c, err := containers.Sync(n, &agentfake.Sysctl{}, k, "wg0", nil)
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
	if c, err := containers.Sync(n, &agentfake.Sysctl{}, k, "wg0", nil); err != nil || c != 0 {
		t.Fatalf("second Sync = %d, %v", c, err)
	}
	// Taken away: rules first, then the route; the addresses stay.
	n.Calls = nil
	if c, err := containers.Sync(n, &agentfake.Sysctl{}, nil, "wg0", nil); err != nil || c != 7 {
		t.Fatalf("Sync(nil) = %d, %v", c, err)
	}
	if last := n.Calls[len(n.Calls)-1]; last != "blackhole false" || n.HasBlackhole() {
		t.Errorf("the route must go last: %v", n.Calls)
	}
	if c, err := containers.Sync(n, &agentfake.Sysctl{}, nil, "wg0", nil); err != nil || c != 0 {
		t.Fatalf("Sync(nil) again = %d, %v", c, err)
	}
}

// The agent not on the mesh network is a degraded step that says what to
// do, and changes nothing.
func TestSyncWithoutTheMeshNetwork(t *testing.T) {
	n := agentfake.NewNet()
	n.Attach("eth0", netip.MustParsePrefix("172.20.0.3/16"))
	n.Attach("wg0", netip.MustParsePrefix("10.77.128.9/32")) // the tunnel is skipped
	_, err := containers.Sync(n, &agentfake.Sysctl{}, testfix.Containers(proto.ModeContainer), "wg0", nil)
	if err == nil || !strings.Contains(err.Error(), "not on the meshctl-mesh network") || !strings.Contains(err.Error(), "docs/containers.md") {
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
	c, err := containers.Sync(n, sys, k, "wg0", nil)
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
	if c, err := containers.Sync(n, sys, k, "wg0", nil); err != nil || c != 0 {
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
		_, err := containers.Sync(n, &agentfake.Sysctl{}, testfix.Containers(c.mode), "wg0", nil)
		if err == nil || !strings.HasPrefix(err.Error(), "containers: ") || !strings.Contains(err.Error(), c.want) {
			t.Errorf("%s: Sync = %v, want %q", c.failOn, err, c.want)
		}
	}
	// A stale rule that will not go, and a blackhole route that will not.
	n := agentfake.NewNet()
	n.Attach("eth1", netip.MustParsePrefix("10.77.128.128/16"))
	n.Install(containers.Rule{Priority: 1003, Iif: "wg0", Dst: mesh, Table: containers.MainTable})
	n.Fail, n.FailOn = boom, "DelRule"
	if _, err := containers.Sync(n, &agentfake.Sysctl{}, testfix.Containers(proto.ModeContainer), "wg0", nil); err == nil ||
		!strings.Contains(err.Error(), "operation not permitted") {
		t.Errorf("a stale rule that will not go = %v", err)
	}
	n.FailOn = ""
	n.Fail = nil
	if _, err := containers.Sync(n, &agentfake.Sysctl{}, testfix.Containers(proto.ModeContainer), "wg0", nil); err != nil {
		t.Fatal(err)
	}
	n.Fail, n.FailOn = boom, "SetBlackhole"
	if _, err := containers.Sync(n, &agentfake.Sysctl{}, nil, "wg0", nil); err == nil || !strings.Contains(err.Error(), "operation not permitted") {
		t.Errorf("a blackhole route that will not go = %v", err)
	}
	sysFail := &agentfake.Sysctl{Err: boom}
	if _, err := containers.Sync(agentfake.NewNet(), sysFail, testfix.Containers(proto.ModeHost), "wg0", nil); err == nil ||
		!strings.Contains(err.Error(), "operation not permitted") {
		t.Errorf("a failed sysctl = %v", err)
	}
	k := testfix.Containers(proto.ModeContainer)
	k.Mode = "bridge"
	if _, err := containers.Sync(agentfake.NewNet(), sysFail, k, "wg0", nil); err == nil || !strings.Contains(err.Error(), "unknown mode") {
		t.Errorf("an unknown mode = %v", err)
	}
	k = testfix.Containers(proto.ModeContainer)
	k.Subnet = "10.77.128.0"
	if _, err := containers.Sync(agentfake.NewNet(), sysFail, k, "wg0", nil); err == nil {
		t.Error("a malformed subnet was accepted")
	}
	if _, err := containers.Guard(*k, "wg0"); err == nil {
		t.Error("Guard accepted a malformed subnet")
	}
	if _, err := containers.Gateway(*k, "wg0", "eth1"); err == nil {
		t.Error("Gateway accepted a malformed subnet")
	}
}

var (
	subnet6 = netip.MustParsePrefix(testfix.ContainerSubnet6)
	mesh6   = netip.MustParsePrefix(testfix.MeshCIDR6)
	bridge6 = netip.MustParsePrefix("fd10:77::8000:1/128")
	// proxied is what apply hands Sync: the node's own IPv6 address and the
	// ones it routes over wg0, an address outside the mesh and one of the
	// node's own containers, which Sync leaves out.
	proxied = []netip.Addr{netip.MustParseAddr(testfix.Laptop.MeshIP6), netip.MustParseAddr(testfix.VPS1.MeshIP6),
		netip.MustParseAddr("2001:db8::1"), netip.MustParseAddr("fd10:77::8000:8005")}
	far = []netip.Addr{netip.MustParseAddr(testfix.Laptop.MeshIP6), netip.MustParseAddr(testfix.VPS1.MeshIP6)}
)

// With the containers' IPv6 (M8) the gateway's rules repeat for IPv6,
// with the node's /112 and the mesh's IPv6 prefix (the 8a trial's R0-R7).
func TestGatewayRulesWithIPv6(t *testing.T) {
	got, err := containers.Gateway(*testfix.Containers6(proto.ModeContainer), "wg0", "eth1")
	if err != nil {
		t.Fatal(err)
	}
	m, b := containers.MainTable, containers.BlackholeTable
	want := []containers.Rule{
		{Priority: 1000, IPv6: true, Iif: "lo", Table: m},
		{Priority: 1001, IPv6: true, Iif: "wg0", Dst: bridge6, Table: b},
		{Priority: 1002, IPv6: true, Iif: "wg0", Dst: subnet6, Table: m},
		{Priority: 1004, IPv6: true, Iif: "eth1", Src: bridge6, Table: b},
		{Priority: 1005, IPv6: true, Iif: "eth1", Src: subnet6, Dst: mesh6, Table: m},
		{Priority: 1006, IPv6: true, Table: b},
	}
	if len(got) != 12 || !slices.Equal(got[6:], want) || slices.ContainsFunc(got[:6], func(r containers.Rule) bool { return r.IPv6 }) {
		t.Fatalf("Gateway =\n%v\nwant the IPv4 rules, then\n%v", got, want)
	}
	if s := got[7].String(); s != "ipv6 rule 1001: iif wg0 to fd10:77::8000:1/128 lookup 51821" {
		t.Errorf("String = %q", s)
	}
}

// Host mode with the containers' IPv6: the guard for the /112, and IPv6
// from the bridge goes to the mesh or is refused at once, so an app whose
// IPv6 default route is the bridge falls back to IPv4 (the 8a trial's E6).
func TestGuardRulesWithIPv6(t *testing.T) {
	got, err := containers.Guard(*testfix.Containers6(proto.ModeHost), "wg0")
	if err != nil {
		t.Fatal(err)
	}
	m, br := containers.MainTable, proto.ContainerBridge
	want := []containers.Rule{
		{Priority: 5200, IPv6: true, Iif: "wg0", Dst: subnet6, Table: m},
		{Priority: 5201, IPv6: true, Iif: br, Dst: mesh6, Table: m},
		{Priority: 5202, IPv6: true, Iif: "lo", Dst: subnet6, Table: m},
		{Priority: 5203, IPv6: true, Dst: subnet6, Table: containers.BlackholeTable},
		{Priority: 5204, IPv6: true, Iif: br, Unreachable: true},
	}
	if len(got) != 9 || !slices.Equal(got[4:], want) {
		t.Fatalf("Guard =\n%v\nwant the IPv4 rules, then\n%v", got, want)
	}
	if s := got[8].String(); s != "ipv6 rule 5204: iif meshctl0 unreachable" {
		t.Errorf("String = %q", s)
	}
	for _, r := range got {
		if !containers.Owned(r.Priority) {
			t.Errorf("%s is not at an owned priority", r)
		}
	}
}

// Container mode with IPv6: the gateway claims ::2 beside .2, both
// families' blackhole routes come before their rules, the mesh side
// answers neighbour discovery for the far addresses only, and a second
// Sync changes nothing; a far address that goes loses its entry.
func TestSyncContainerModeWithIPv6(t *testing.T) {
	n := agentfake.NewNet()
	n.Attach("eth1", netip.MustParsePrefix("10.77.128.128/16"))
	n.Attach("wg0", netip.MustParsePrefix("10.77.2.1/32"))
	k := testfix.Containers6(proto.ModeContainer)
	c, err := containers.Sync(n, &agentfake.Sysctl{}, k, "wg0", proxied)
	if err != nil || c != 2+2+12+2 {
		t.Fatalf("Sync = %d, %v; calls %v", c, err, n.Calls)
	}
	if got := n.AddrsOf("eth1"); !slices.Contains(got, netip.MustParsePrefix("fd10:77::8000:2/64")) {
		t.Errorf("eth1 = %v, want the IPv6 gateway fd10:77::8000:2/64 too", got)
	}
	if !n.HasBlackhole() || !n.HasBlackhole6() {
		t.Error("a family's blackhole route is missing")
	}
	if i, j := slices.Index(n.Calls, "ipv6 blackhole true"), slices.IndexFunc(n.Calls, func(s string) bool {
		return strings.HasPrefix(s, "add ipv6 rule")
	}); i < 0 || i > j {
		t.Errorf("the IPv6 blackhole route must come before the IPv6 rules: %v", n.Calls)
	}
	if got := n.ProxiesOf("eth1"); !slices.Equal(got, far) {
		t.Errorf("proxy entries = %v, want %v", got, far)
	}
	if c, err := containers.Sync(n, &agentfake.Sysctl{}, k, "wg0", proxied); err != nil || c != 0 {
		t.Fatalf("second Sync = %d, %v; calls %v", c, err, n.Calls)
	}
	if c, err := containers.Sync(n, &agentfake.Sysctl{}, k, "wg0", far[:1]); err != nil || c != 1 ||
		!slices.Equal(n.ProxiesOf("eth1"), far[:1]) {
		t.Fatalf("Sync without %s = %d, %v; entries %v", far[1], c, err, n.ProxiesOf("eth1"))
	}
	// The containers' IPv6 taken off: its rules and route go, the address
	// and the entries stay with the runtime's network.
	n.Calls = nil
	if c, err := containers.Sync(n, &agentfake.Sysctl{}, testfix.Containers(proto.ModeContainer), "wg0", proxied); err != nil || c != 7 {
		t.Fatalf("Sync without IPv6 = %d, %v; calls %v", c, err, n.Calls)
	}
	if n.HasBlackhole6() || !n.HasBlackhole() || len(n.ProxiesOf("eth1")) != 1 {
		t.Errorf("after IPv6 went: calls %v", n.Calls)
	}
}

// Host mode with IPv6: the bridge gets ::1 beside .1 and proxy NDP, the
// guard covers both families, the bridge answers for the far addresses.
func TestSyncHostModeWithIPv6(t *testing.T) {
	n, sys := agentfake.NewNet(), &agentfake.Sysctl{}
	k := testfix.Containers6(proto.ModeHost)
	c, err := containers.Sync(n, sys, k, "wg0", proxied)
	if err != nil || c != 1+2+4+2+9+2 {
		t.Fatalf("Sync = %d, %v; calls %v", c, err, n.Calls)
	}
	want := []netip.Prefix{netip.MustParsePrefix("10.77.128.1/16"), netip.MustParsePrefix("fd10:77::8000:1/64")}
	if got := n.AddrsOf(proto.ContainerBridge); !slices.Equal(got, want) {
		t.Errorf("%s = %v, want %v", proto.ContainerBridge, got, want)
	}
	if sys.Values[proto.SysctlBridgeProxyNDP] != "1" {
		t.Errorf("proxy NDP is off on the bridge: %v", sys.Values)
	}
	if got := n.ProxiesOf(proto.ContainerBridge); !slices.Equal(got, far) {
		t.Errorf("proxy entries = %v, want %v", got, far)
	}
	if c, err := containers.Sync(n, sys, k, "wg0", proxied); err != nil || c != 0 {
		t.Fatalf("second Sync = %d, %v", c, err)
	}
}

// A kernel without IPv6 policy routing cannot isolate the containers'
// IPv6: the IPv4 side is set up, the IPv6 side left out -- no address, no
// IPv6 rule or route, entries a former kernel left removed -- the /112
// dropped in the main table, and Sync says why and what to do. With policy
// routing the drop goes again.
func TestSyncWithoutIPv6PolicyRouting(t *testing.T) {
	for _, mode := range []string{proto.ModeContainer, proto.ModeHost} {
		n := agentfake.NewNet()
		n.NoIPv6Rules = true
		link := "eth1"
		if mode == proto.ModeHost {
			link = proto.ContainerBridge
		}
		n.Attach(link, netip.MustParsePrefix("10.77.128.128/16"))
		if err := n.AddProxy(link, far[0]); err != nil {
			t.Fatal(err)
		}
		_, err := containers.Sync(n, &agentfake.Sysctl{}, testfix.Containers6(mode), "wg0", proxied)
		if !errors.Is(err, containers.ErrIPv6Unrouted) || !strings.Contains(err.Error(), "--containers6=false") {
			t.Fatalf("%s: Sync = %v", mode, err)
		}
		rules, _ := n.Rules()
		if len(rules) == 0 || slices.ContainsFunc(rules, func(r containers.Rule) bool { return r.IPv6 }) || n.HasBlackhole6() || !n.HasBlackhole() {
			t.Errorf("%s: rules %v, IPv6 blackhole %v", mode, rules, n.HasBlackhole6())
		}
		if slices.ContainsFunc(n.AddrsOf(link), func(p netip.Prefix) bool { return p.Addr().Is6() }) || len(n.ProxiesOf(link)) != 0 {
			t.Errorf("%s: %s = %v, entries %v", mode, link, n.AddrsOf(link), n.ProxiesOf(link))
		}
		if dropped, _ := n.Dropped(); !slices.Equal(dropped, []netip.Prefix{subnet6}) {
			t.Errorf("%s: dropped %v, want %s", mode, dropped, subnet6)
		}
		n.NoIPv6Rules = false
		if _, err := containers.Sync(n, &agentfake.Sysctl{}, testfix.Containers6(mode), "wg0", proxied); err != nil {
			t.Fatalf("%s: Sync with IPv6 policy routing = %v", mode, err)
		}
		if dropped, _ := n.Dropped(); len(dropped) != 0 || !n.HasBlackhole6() {
			t.Errorf("%s: with IPv6 policy routing: dropped %v, IPv6 blackhole %v", mode, dropped, n.HasBlackhole6())
		}
	}
}

// Each IPv6 step's failure names it; nothing after it runs.
func TestSyncFailuresWithIPv6(t *testing.T) {
	boom := errors.New("operation not permitted")
	for _, c := range []struct{ failOn, want string }{
		{"IPv6Rules", "operation not permitted"},
		{"AddAddr", "add fd10:77::8000:2/64 to eth1"},
		{"Proxies", "proxy entries of eth1"},
		{"AddProxy", "answer neighbour discovery for fd10:77::2:1 on eth1"},
	} {
		n := agentfake.NewNet()
		n.Attach("eth1", netip.MustParsePrefix("10.77.128.128/16"))
		n.Attach("eth1", netip.MustParsePrefix("10.77.128.2/16")) // the IPv4 gateway is in place
		n.Fail, n.FailOn = boom, c.failOn
		_, err := containers.Sync(n, &agentfake.Sysctl{}, testfix.Containers6(proto.ModeContainer), "wg0", proxied)
		if err == nil || !strings.HasPrefix(err.Error(), "containers: ") || !strings.Contains(err.Error(), c.want) {
			t.Errorf("%s: Sync = %v, want %q", c.failOn, err, c.want)
		}
	}
	n := agentfake.NewNet()
	n.Attach("eth1", netip.MustParsePrefix("10.77.128.128/16"))
	k := testfix.Containers6(proto.ModeContainer)
	if _, err := containers.Sync(n, &agentfake.Sysctl{}, k, "wg0", proxied); err != nil {
		t.Fatal(err)
	}
	n.Fail, n.FailOn = boom, "DelProxy"
	if _, err := containers.Sync(n, &agentfake.Sysctl{}, k, "wg0", nil); err == nil || !strings.Contains(err.Error(), "stop answering") {
		t.Errorf("an entry that will not go = %v", err)
	}
	for _, c := range []struct{ failOn, want string }{{"Dropped", "operation not permitted"}, {"SetDropped", "drop fd10:77::8000:0/112"}} {
		n := agentfake.NewNet()
		n.NoIPv6Rules = true
		n.Attach("eth1", netip.MustParsePrefix("10.77.128.128/16"))
		n.Fail, n.FailOn = boom, c.failOn
		if _, err := containers.Sync(n, &agentfake.Sysctl{}, k, "wg0", proxied); err == nil || errors.Is(err, containers.ErrIPv6Unrouted) ||
			!strings.Contains(err.Error(), c.want) {
			t.Errorf("%s: Sync = %v, want %q", c.failOn, err, c.want)
		}
	}
	n = agentfake.NewNet()
	n.Attach("eth1", netip.MustParsePrefix("10.77.128.128/16"))
	if err := n.SetDropped(subnet6, true); err != nil {
		t.Fatal(err)
	}
	n.Fail, n.FailOn = boom, "SetDropped"
	if _, err := containers.Sync(n, &agentfake.Sysctl{}, k, "wg0", proxied); err == nil || !strings.Contains(err.Error(), "stop dropping") {
		t.Errorf("a drop that will not go = %v", err)
	}
	k.Gateway6 = "fd10:77::8000:2/64"
	if _, err := containers.Gateway(*k, "wg0", "eth1"); err == nil {
		t.Error("Gateway accepted a malformed IPv6 gateway")
	}
}
