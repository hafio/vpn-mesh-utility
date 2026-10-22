package agentfake

import (
	"context"
	"errors"
	"net/netip"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/hafio/vpn-mesh-utility/internal/agent/containers"
	"github.com/hafio/vpn-mesh-utility/internal/agent/dns"
	"github.com/hafio/vpn-mesh-utility/internal/agent/exitmode"
	"github.com/hafio/vpn-mesh-utility/internal/agent/policy"
	"github.com/hafio/vpn-mesh-utility/internal/agent/wg"
	"github.com/hafio/vpn-mesh-utility/internal/proto"
)

// The apply engine's idempotency tests are only as good as this fake's
// fidelity to the kernel, so its semantics are pinned here.

func pfx(ss ...string) []netip.Prefix {
	var out []netip.Prefix
	for _, s := range ss {
		out = append(out, netip.MustParsePrefix(s))
	}
	return out
}

func peer(d wg.Device, k wg.Key) (wg.Peer, bool) {
	i := slices.IndexFunc(d.Peers, func(p wg.Peer) bool { return p.PublicKey == k })
	if i < 0 {
		return wg.Peer{}, false
	}
	return d.Peers[i], true
}

func TestConfigureHasKernelSemantics(t *testing.T) {
	h := NewHost()
	if err := h.Configure("wg0", wg.Change{}); err == nil {
		t.Fatal("Configure on a missing interface succeeded")
	}
	if err := h.CreateLink("wg0"); err != nil {
		t.Fatal(err)
	}
	if err := h.CreateLink("wg0"); err == nil {
		t.Fatal("creating an existing interface succeeded")
	}
	priv, k1, k2 := wg.Key{1}, wg.Key{2}, wg.Key{3}
	port, mark, ka := 51820, proto.FwMark, 25*time.Second
	ep := netip.MustParseAddrPort("203.0.113.11:51820")
	err := h.Configure("wg0", wg.Change{PrivateKey: &priv, ListenPort: &port, FwMark: &mark, Peers: []wg.PeerChange{
		{PublicKey: k1, Endpoint: &ep, Keepalive: &ka, ReplaceAllowedIPs: true, AllowedIPs: pfx("10.77.0.2/32")},
		{PublicKey: k2, AllowedIPs: pfx("10.77.0.3/32")},
	}})
	if err != nil {
		t.Fatal(err)
	}
	d, _ := h.DeviceOf("wg0")
	p1, _ := peer(d, k1)
	if d.PrivateKey != priv || d.ListenPort != port || d.FwMark != mark || len(d.Peers) != 2 ||
		p1.Endpoint != ep || p1.Keepalive != ka || !slices.Equal(p1.AllowedIPs, pfx("10.77.0.2/32")) {
		t.Fatalf("device after the first change: %+v", d)
	}

	// Without ReplaceAllowedIPs the kernel appends; with it, it replaces.
	steps := []struct {
		change wg.PeerChange
		want   []netip.Prefix
	}{
		{wg.PeerChange{PublicKey: k2, AllowedIPs: pfx("10.77.2.1/32")}, pfx("10.77.0.3/32", "10.77.2.1/32")},
		{wg.PeerChange{PublicKey: k2, ReplaceAllowedIPs: true, AllowedIPs: pfx("10.77.1.1/32")}, pfx("10.77.1.1/32")},
	}
	for _, s := range steps {
		if err := h.Configure("wg0", wg.Change{Peers: []wg.PeerChange{s.change}}); err != nil {
			t.Fatal(err)
		}
		d, _ = h.DeviceOf("wg0")
		if p2, _ := peer(d, k2); !slices.Equal(p2.AllowedIPs, s.want) {
			t.Fatalf("after %+v: allowed %v, want %v", s.change, p2.AllowedIPs, s.want)
		}
	}

	// Removal; removing a peer that is not there is a no-op, as in the kernel.
	if err := h.Configure("wg0", wg.Change{Peers: []wg.PeerChange{{PublicKey: k1, Remove: true}, {PublicKey: wg.Key{9}, Remove: true}}}); err != nil {
		t.Fatal(err)
	}
	d, _ = h.DeviceOf("wg0")
	if _, ok := peer(d, k1); ok || len(d.Peers) != 1 {
		t.Fatalf("after removal: %+v", d.Peers)
	}

	// Device hands out copies: a caller cannot change the fake's state.
	got, _, err := h.Device("wg0")
	if err != nil {
		t.Fatal(err)
	}
	got.Peers[0].AllowedIPs[0] = netip.MustParsePrefix("0.0.0.0/0")
	if d, _ := h.DeviceOf("wg0"); d.Peers[0].AllowedIPs[0] == got.Peers[0].AllowedIPs[0] {
		t.Fatal("Device leaked the fake's own slice")
	}
	if h.Mutations() != 5 {
		t.Fatalf("Mutations = %d, want 5 (one create, four configures): %v", h.Mutations(), h.Calls)
	}
}

func TestLinksAddressesAndFailures(t *testing.T) {
	h := NewHost()
	h.AddForeignLink("wg1")
	if l, err := h.Link("wg1"); err != nil || !l.Exists || l.Ours {
		t.Fatalf("foreign link = %+v, %v", l, err)
	}
	if l, err := h.Link("none"); err != nil || l.Exists {
		t.Fatalf("missing link = %+v, %v", l, err)
	}
	if err := h.CreateLink("wg0"); err != nil {
		t.Fatal(err)
	}
	a := netip.MustParsePrefix("10.77.0.1/32")
	for _, step := range []error{h.SetMTU("wg0", 1380), h.SetUp("wg0"), h.AddAddr("wg0", a)} {
		if step != nil {
			t.Fatal(step)
		}
	}
	if l, _ := h.Link("wg0"); !l.Ours || l.MTU != 1380 || !l.Up || !slices.Equal(l.Addrs, []netip.Prefix{a}) {
		t.Fatalf("link = %+v", l)
	}
	if err := h.DelAddr("wg0", a); err != nil {
		t.Fatal(err)
	}
	h.RemoveLink("wg0")
	for name, call := range map[string]func() error{
		"SetMTU":  func() error { return h.SetMTU("wg0", 1420) },
		"SetUp":   func() error { return h.SetUp("wg0") },
		"AddAddr": func() error { return h.AddAddr("wg0", a) },
		"DelAddr": func() error { return h.DelAddr("wg0", a) },
		"Device":  func() error { _, _, err := h.Device("wg0"); return err },
	} {
		if call() == nil {
			t.Errorf("%s on a removed interface succeeded", name)
		}
	}
	h.SetFail("Link")
	if _, err := h.Link("wg1"); err == nil {
		t.Fatal("Fail did not inject an error")
	}
	h.SetFail("")
	if _, err := h.Link("wg1"); err != nil {
		t.Fatalf("Link after SetFail(\"\") = %v", err)
	}
	if _, ok := h.DeviceOf("wg0"); ok || h.Close() != nil {
		t.Fatal("DeviceOf a removed interface, or Close")
	}
}

func TestRoutesSysctlAndFirewall(t *testing.T) {
	r := NewRoutes()
	p := netip.MustParsePrefix("10.77.0.2/32")
	if err := r.Add("wg0", p); err != nil {
		t.Fatal(err)
	}
	if got, _ := r.List("wg0"); !slices.Equal(got, []netip.Prefix{p}) {
		t.Fatalf("List = %v", got)
	}
	if err := r.Del("wg0", p); err != nil || r.Mutations() != 2 {
		t.Fatalf("Del = %v, mutations %d", err, r.Mutations())
	}
	r.Fail = errors.New("netlink down")
	if _, err := r.List("wg0"); err == nil {
		t.Fatal("Fail did not inject an error")
	}

	s := &Sysctl{}
	params := []proto.Sysctl{{Key: proto.SysctlIPForward, Value: "1"}}
	if n, _ := s.Apply(params); n != 1 {
		t.Fatalf("first Apply = %d", n)
	}
	if n, _ := s.Apply(params); n != 0 || s.Writes != 1 {
		t.Fatalf("second Apply = %d, writes %d", n, s.Writes)
	}

	fw := &Firewall{Err: errors.New("firewalld down")}
	if _, err := fw.Apply(context.Background(), nil); err == nil || len(fw.Applied) != 1 {
		t.Fatalf("Firewall.Apply = %v, recorded %d", err, len(fw.Applied))
	}
	if fw.Recheck(); fw.Rechecked() != 1 {
		t.Fatal("Recheck was not counted")
	}
}

// DNS answers the state of each kind of request, and changes only when the
// request does.
func TestDNS(t *testing.T) {
	d := &DNS{}
	ctx := context.Background()
	on := dns.Want{Iface: "wg0", IP: netip.MustParseAddr("10.77.2.1"), Config: &proto.DNS{Zone: proto.DefaultZone}}
	off := on
	off.Off = true
	for _, c := range []struct {
		w       dns.Want
		state   string
		changes int
	}{{on, proto.DNSOn, 1}, {on, proto.DNSOn, 0}, {off, proto.DNSOff, 1}, {dns.Want{Iface: "wg0"}, "", 1}} {
		if state, n, err := d.Apply(ctx, c.w); err != nil || state != c.state || n != c.changes {
			t.Errorf("Apply(%+v) = %q, %d, %v", c.w, state, n, err)
		}
	}
	d.Unavailable = true
	if state, _, _ := d.Apply(ctx, on); state != proto.DNSUnavailable {
		t.Errorf("unavailable host = %q", state)
	}
	d.SetErr(errors.New("port 53 in use"))
	if _, _, err := d.Apply(ctx, on); err == nil {
		t.Error("SetErr did not inject an error")
	}
	if last, ok := d.Last(); !ok || last != on {
		t.Errorf("Last = %+v, %v", last, ok)
	}
	if d.Close(ctx, "wg0"); !d.IsClosed() {
		t.Error("Close was not recorded")
	}
	names := []proto.MeshName{{Name: "web", Node: "nas", IP: "10.77.128.129"}}
	d.SetNames(names)
	names[0].Name = "changed"
	if got := d.Names(); len(got) != 1 || got[0].Name != "web" {
		t.Errorf("Names = %v, want a copy of what was set", got)
	}
}

// Policy answers like the kernel: an add of what is there is EEXIST, a
// delete of what is not is ENOENT, and neither changes anything.
func TestPolicyHasKernelSemantics(t *testing.T) {
	p := &Policy{}
	rule := exitmode.Rule{Family: exitmode.IPv4, Priority: exitmode.PrioritySuppress, Table: exitmode.MainTable}
	route := exitmode.Route{Family: exitmode.IPv4, Link: "wg0"}
	for _, step := range []error{p.AddRoute(route), p.AddRule(rule), p.DelRule(rule)} {
		if step != nil {
			t.Fatal(step)
		}
	}
	if err := p.AddRoute(route); err == nil || !strings.Contains(err.Error(), "file exists") {
		t.Fatalf("a second AddRoute = %v", err)
	}
	if err := p.DelRule(rule); err == nil || !strings.Contains(err.Error(), "no such entry") {
		t.Fatalf("DelRule of a missing rule = %v", err)
	}
	if err := p.DelRoute(route); err != nil {
		t.Fatal(err)
	}
	want := []string{"add " + route.String(), "add " + rule.String(), "del " + rule.String(), "del " + route.String()}
	if !slices.Equal(p.Calls, want) {
		t.Fatalf("Calls = %q, want %q (refused calls are not recorded)", p.Calls, want)
	}

	p.Install(rule)
	p.InstallRoute(route)
	rules, err1 := p.Rules()
	routes, err2 := p.Routes()
	if err1 != nil || err2 != nil || !slices.Equal(rules, []exitmode.Rule{rule}) || !slices.Equal(routes, []exitmode.Route{route}) {
		t.Fatalf("after Install: %v %v (%v %v)", rules, routes, err1, err2)
	}
	if p.IPv6() != exitmode.V6Policy || (&Policy{NoIPv6: true}).IPv6() != exitmode.V6None ||
		(&Policy{V6Unrouted: true}).IPv6() != exitmode.V6Unrouted {
		t.Fatal("IPv6 must follow NoIPv6 and V6Unrouted")
	}

	// Fail alone fails every mutating call; with FailOn only that method,
	// which may be a read.
	boom := errors.New("operation not permitted")
	p.Fail = boom
	if _, err := p.Rules(); err != nil {
		t.Fatalf("Fail alone failed a read: %v", err)
	}
	if err := p.AddRoute(exitmode.Route{Family: exitmode.IPv6}); !errors.Is(err, boom) {
		t.Fatalf("AddRoute with Fail = %v", err)
	}
	p.FailOn = "Routes"
	if _, err := p.Routes(); !errors.Is(err, boom) {
		t.Fatalf("Routes with FailOn = %v", err)
	}
	if err := p.DelRule(rule); err != nil {
		t.Fatalf("FailOn Routes failed DelRule: %v", err)
	}
}

// The container network's fake host answers as the kernel would: an add of
// what is there and a removal of what is not fail, an interface must exist
// to hold an address, LinkHolding skips the named interface, EnsureBridge
// creates once, and Fail/FailOn pick what fails.
func TestNetHasKernelSemantics(t *testing.T) {
	n := NewNet()
	r := containers.Rule{Priority: 1000, Iif: "lo", Table: containers.MainTable}
	if err := n.AddRule(r); err != nil {
		t.Fatal(err)
	}
	if err := n.AddRule(r); err == nil || !strings.Contains(err.Error(), "file exists") {
		t.Fatalf("adding a rule twice = %v", err)
	}
	if err := n.DelRule(r); err != nil {
		t.Fatal(err)
	}
	if err := n.DelRule(r); err == nil || !strings.Contains(err.Error(), "no such rule") {
		t.Fatalf("removing a missing rule = %v", err)
	}
	n.Install(r)
	if rules, err := n.Rules(); err != nil || !slices.Equal(rules, []containers.Rule{r}) {
		t.Fatalf("Rules = %v, %v", rules, err)
	}
	if err := n.SetBlackhole(true); err != nil || !n.HasBlackhole() {
		t.Fatalf("SetBlackhole = %v", err)
	}
	if on, err := n.Blackhole(); err != nil || !on {
		t.Fatalf("Blackhole = %v, %v", on, err)
	}

	if err := n.AddAddr("eth1", netip.MustParsePrefix("10.77.128.2/16")); err == nil {
		t.Fatal("an address went to a missing interface")
	}
	if _, err := n.Addrs("eth1"); err == nil {
		t.Fatal("Addrs of a missing interface succeeded")
	}
	n.Attach("wg0", netip.MustParsePrefix("10.77.128.9/32"))
	n.Attach("eth1", netip.MustParsePrefix("10.77.128.128/16"))
	sub := netip.MustParsePrefix("10.77.128.0/24")
	if link, err := n.LinkHolding(sub, "wg0"); err != nil || link != "eth1" {
		t.Fatalf("LinkHolding skipping wg0 = %q, %v", link, err)
	}
	if link, err := n.LinkHolding(netip.MustParsePrefix("10.88.0.0/24"), ""); err != nil || link != "" {
		t.Fatalf("LinkHolding of a subnet nobody holds = %q, %v", link, err)
	}
	if err := n.AddAddr("eth1", netip.MustParsePrefix("10.77.128.2/16")); err != nil ||
		!slices.Equal(n.AddrsOf("eth1"), pfx("10.77.128.128/16", "10.77.128.2/16")) {
		t.Fatalf("AddAddr = %v; eth1 %v", err, n.AddrsOf("eth1"))
	}
	if created, err := n.EnsureBridge("meshctl0"); err != nil || !created {
		t.Fatalf("EnsureBridge = %v, %v", created, err)
	}
	if created, err := n.EnsureBridge("meshctl0"); err != nil || created {
		t.Fatalf("EnsureBridge again = %v, %v", created, err)
	}
	if want := []string{"add rule 1000: iif lo lookup 254", "remove rule 1000: iif lo lookup 254", "blackhole true",
		"addr eth1 10.77.128.2/16", "bridge meshctl0"}; !slices.Equal(n.Calls, want) {
		t.Fatalf("Calls = %q", n.Calls)
	}

	boom := errors.New("operation not permitted")
	n.Fail = boom
	for name, err := range map[string]error{
		"AddRule":      n.AddRule(r),
		"DelRule":      n.DelRule(r),
		"SetBlackhole": n.SetBlackhole(false),
		"AddAddr":      n.AddAddr("eth1", netip.MustParsePrefix("10.77.128.3/16")),
	} {
		if !errors.Is(err, boom) {
			t.Errorf("%s with Fail = %v", name, err)
		}
	}
	if _, err := n.EnsureBridge("br1"); !errors.Is(err, boom) {
		t.Errorf("EnsureBridge with Fail = %v", err)
	}
	if _, err := n.Rules(); err != nil {
		t.Errorf("Fail alone must leave the queries alone: %v", err)
	}
	for _, method := range []string{"Rules", "Blackhole", "LinkHolding", "Addrs"} {
		n.FailOn = method
		var err error
		switch method {
		case "Rules":
			_, err = n.Rules()
		case "Blackhole":
			_, err = n.Blackhole()
		case "LinkHolding":
			_, err = n.LinkHolding(sub, "")
		case "Addrs":
			_, err = n.Addrs("eth1")
		}
		if !errors.Is(err, boom) {
			t.Errorf("%s with FailOn = %v", method, err)
		}
	}
	s := &Sysctl{Err: boom}
	if _, err := s.Apply([]proto.Sysctl{{Key: proto.SysctlIPForward, Value: "1"}}); !errors.Is(err, boom) {
		t.Errorf("Sysctl with Err = %v", err)
	}
}

// The mesh firewall's kernel holds one table by fingerprint, replaces and
// removes it, counts deleted flows, and fails each call on request.
func TestKernelHoldsOneTable(t *testing.T) {
	k := &Kernel{Forgotten: 2}
	if mark, err := k.Installed(); mark != "" || err != nil {
		t.Fatalf("an empty kernel = %q, %v", mark, err)
	}
	rs := &policy.Ruleset{Mesh: netip.MustParsePrefix("10.77.0.0/16")}
	if err := k.Replace(rs, "abc"); err != nil {
		t.Fatal(err)
	}
	if table, mark := k.Holds(); table != rs || mark != "abc" || k.Replaces != 1 {
		t.Fatalf("after Replace: %v %q %d", table, mark, k.Replaces)
	}
	if n, err := k.Forget(rs.Mesh); n != 2 || err != nil || k.Forgets != 1 {
		t.Fatalf("Forget = %d, %v", n, err)
	}
	if err := k.Remove(); err != nil {
		t.Fatal(err)
	}
	if mark, _ := k.Installed(); mark != "" {
		t.Fatalf("after Remove the kernel holds %q", mark)
	}
	boom := errors.New("boom")
	k.ReadErr, k.ReplaceErr, k.RemoveErr, k.ForgetErr = boom, boom, boom, boom
	if _, err := k.Installed(); !errors.Is(err, boom) {
		t.Fatal("ReadErr ignored")
	}
	if err := k.Replace(rs, "def"); !errors.Is(err, boom) || k.Replaces != 1 {
		t.Fatal("ReplaceErr ignored, or it replaced anyway")
	}
	if err := k.Remove(); !errors.Is(err, boom) {
		t.Fatal("RemoveErr ignored")
	}
	if _, err := k.Forget(rs.Mesh); !errors.Is(err, boom) {
		t.Fatal("ForgetErr ignored")
	}
}
