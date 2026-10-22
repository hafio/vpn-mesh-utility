package apply_test

import (
	"context"
	"errors"
	"net/netip"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/hafio/vpn-mesh-utility/internal/agent/agentfake"
	"github.com/hafio/vpn-mesh-utility/internal/agent/apply"
	"github.com/hafio/vpn-mesh-utility/internal/agent/exitmode"
	"github.com/hafio/vpn-mesh-utility/internal/agent/policy"
	"github.com/hafio/vpn-mesh-utility/internal/agent/wg"
	"github.com/hafio/vpn-mesh-utility/internal/proto"
	"github.com/hafio/vpn-mesh-utility/internal/testfix"
)

type rig struct {
	e   *apply.Engine
	h   *agentfake.Host
	r   *agentfake.Routes
	p   *agentfake.Policy
	s   *agentfake.Sysctl
	fw  *agentfake.Firewall
	dns *agentfake.DNS
	k   *agentfake.Kernel
	key wg.Key
}

func newRig(t *testing.T) *rig {
	t.Helper()
	k, err := wg.ParseKey(testfix.WGKey(1))
	if err != nil {
		t.Fatal(err)
	}
	r := &rig{h: agentfake.NewHost(), r: agentfake.NewRoutes(), p: &agentfake.Policy{}, s: &agentfake.Sysctl{}, fw: &agentfake.Firewall{},
		dns: &agentfake.DNS{}, k: &agentfake.Kernel{}, key: k}
	r.e = apply.New(apply.Deps{Host: r.h, Routes: r.r, Exit: r.p, Sysctl: r.s, MeshFirewall: &policy.Enforcer{K: r.k}, Firewall: r.fw,
		DNS: r.dns, Iface: "wg0", Key: k})
	return r
}

func (r *rig) mutations() int {
	return r.h.Mutations() + r.r.Mutations() + len(r.p.Calls) + r.s.Writes + r.k.Replaces
}

// meshDNS is a DNS section for the reference servers.
func meshDNS() *proto.DNS {
	return &proto.DNS{Zone: proto.DefaultZone, Split: true, Upstream: []string{"1.1.1.1"},
		Records: []proto.Record{{Name: "vps-1", IP: testfix.VPS1.MeshIP}, {Name: "vps-2", IP: testfix.VPS2.MeshIP}}}
}

func TestApplyIsIdempotent(t *testing.T) {
	r := newRig(t)
	cfg := testfix.ServerConfig()
	cfg.Sysctl = []proto.Sysctl{{Key: proto.SysctlIPForward, Value: "1"}}
	cfg.DNS = meshDNS()
	res, err := r.e.Apply(context.Background(), cfg)
	if err != nil {
		t.Fatal(err)
	}
	want, _ := cfg.Hash()
	if res.Changes == 0 || res.Hash != want || res.Version != cfg.Version || len(res.Degraded) != 0 {
		t.Fatalf("first apply = %+v", res)
	}
	before := r.mutations()
	res, err = r.e.Apply(context.Background(), cfg)
	if err != nil || res.Changes != 0 || r.mutations() != before {
		t.Fatalf("second apply = %+v, %v; host calls %v, route calls %v", res, err, r.h.Calls, r.r.Calls)
	}
	if len(r.fw.Applied) != 2 || r.fw.Applied[1] != cfg.Firewall {
		t.Fatalf("the firewall step must run on every apply: %d", len(r.fw.Applied))
	}
	// The resolver binds the node's mesh IP, after the address step.
	w, ok := r.dns.Last()
	if !ok || res.DNS != proto.DNSOn || w.IP != netip.MustParseAddr(testfix.VPS1.MeshIP) || w.Config != cfg.DNS || w.Iface != "wg0" || w.Exit || w.Off {
		t.Fatalf("dns step = %+v, state %q", w, res.DNS)
	}
}

// Dual stack (M8): both addresses and every route of both families but the
// defaults; on a host that cannot use IPv6, the IPv4 side alone, reported.
func TestDualStack(t *testing.T) {
	r := newRig(t)
	cfg := testfix.DualSpokeConfig(testfix.VPS2.Name)
	cfg.Sysctl = []proto.Sysctl{{Key: proto.SysctlSrcValidMark, Value: "1"}, {Key: proto.SysctlIPv6Forward, Value: "1"}}
	want, _ := cfg.Hash()
	res, err := r.e.Apply(context.Background(), cfg)
	if err != nil || strings.Join(res.Families, ",") != "ipv4,ipv6" || res.Hash != want {
		t.Fatalf("dual apply = %+v, %v", res, err)
	}
	addr4, addr6 := netip.MustParsePrefix(testfix.Laptop.MeshIP+"/32"), netip.MustParsePrefix(testfix.Laptop.MeshIP6+"/128")
	if l, _ := r.h.Link("wg0"); !slices.Equal(l.Addrs, []netip.Prefix{addr4, addr6}) {
		t.Errorf("addresses %v", l.Addrs)
	}
	got, _ := r.r.List("wg0")
	if !slices.Contains(got, netip.MustParsePrefix(testfix.NAS.MeshIP6+"/128")) || slices.Contains(got, netip.MustParsePrefix("::/0")) {
		t.Errorf("routes %v: the nas's IPv6, never ::/0", got)
	}
	if _, ok := r.s.Values[proto.SysctlIPv6Forward]; !ok {
		t.Error("the IPv6 sysctl was not set on a host with IPv6")
	}
	// vps-2 has public IPv6, so the exit carries it.
	six := exitmode.Route{Family: exitmode.IPv6, Link: "wg0"}
	if table, _ := r.p.Routes(); !slices.Contains(table, six) {
		t.Errorf("exit table %v: IPv6 goes through vps-2", table)
	}
	before := r.mutations()
	if res, err := r.e.Apply(context.Background(), cfg); err != nil || res.Changes != 0 || r.mutations() != before {
		t.Fatalf("second dual apply = %+v, %v", res, err)
	}

	// The same config on a host that cannot use IPv6; vps-2 is reached on
	// its IPv4 endpoint where its IPv6 one comes first.
	r = newRig(t)
	r.s.NoIPv6 = true
	cfg.Peers[1].Endpoint, cfg.Peers[1].Fallback = testfix.VPS2.Endpoint6, testfix.VPS2.Endpoint
	want, _ = cfg.Hash()
	res, err = r.e.Apply(context.Background(), cfg)
	if err != nil || strings.Join(res.Families, ",") != "ipv4" || res.Hash != want {
		t.Fatalf("apply without IPv6 = %+v, %v", res, err)
	}
	if l, _ := r.h.Link("wg0"); !slices.Equal(l.Addrs, []netip.Prefix{addr4}) {
		t.Errorf("addresses without IPv6: %v", l.Addrs)
	}
	got, _ = r.r.List("wg0")
	dev, _ := r.h.DeviceOf("wg0")
	for _, p := range got {
		if p.Addr().Is6() {
			t.Errorf("an IPv6 route without IPv6: %s", p)
		}
	}
	for _, p := range dev.Peers {
		for _, a := range p.AllowedIPs {
			if a.Addr().Is6() {
				t.Errorf("peer %s carries %s without IPv6", p.PublicKey, a)
			}
		}
		if p.Endpoint.IsValid() && p.Endpoint.Addr().Is6() {
			t.Errorf("peer %s dials %s without IPv6", p.PublicKey, p.Endpoint)
		}
	}
	if k, _ := wg.ParseKey(testfix.WGKey(testfix.VPS2.Key)); !slices.ContainsFunc(dev.Peers, func(p wg.Peer) bool {
		return p.PublicKey == k && p.Endpoint.String() == testfix.VPS2.Endpoint
	}) {
		t.Errorf("vps-2 is not dialled on its IPv4 endpoint: %+v", dev.Peers)
	}
	if _, ok := r.s.Values[proto.SysctlIPv6Forward]; ok || r.s.Values[proto.SysctlSrcValidMark] != "1" {
		t.Errorf("sysctls without IPv6: %v", r.s.Values)
	}
	if table, _ := r.p.Routes(); slices.Contains(table, six) {
		t.Errorf("exit table %v: no IPv6 through the exit without IPv6", table)
	}
	if cfg.Interface.Address6 == "" || !slices.Contains(cfg.Peers[0].AllowedIPs, testfix.VPS1.MeshIP6+"/128") || len(cfg.Sysctl) != 2 {
		t.Error("Apply changed the caller's config")
	}

	// An IPv6-only node cannot run there at all; nothing is touched.
	r = newRig(t)
	r.s.NoIPv6 = true
	cfg.Interface.Address = ""
	if _, err := r.e.Apply(context.Background(), cfg); err == nil || !strings.Contains(err.Error(), "IPv6 only") {
		t.Fatalf("an IPv6-only node without IPv6 = %v", err)
	} else if step, _ := apply.IsStepError(err); step != apply.StepValidate || r.mutations() != 0 {
		t.Fatalf("step %q, %d host changes", step, r.mutations())
	}
}

// A spoke's server that does not answer on its endpoint is dialled on its
// other one (M8), and back when that one goes quiet; a server that answers
// stays where it is, and a re-apply keeps what failover chose.
func TestEndpointFailover(t *testing.T) {
	r := newRig(t)
	now := testfix.Epoch
	e := apply.New(apply.Deps{Host: r.h, Routes: r.r, Exit: r.p, Sysctl: r.s, Iface: "wg0", Key: r.key, Now: func() time.Time { return now }})
	ctx := context.Background()
	cfg := testfix.DualSpokeConfig("") // only vps-2 has an IPv6 endpoint
	if _, err := e.Apply(ctx, cfg); err != nil {
		t.Fatal(err)
	}
	vps2, _ := wg.ParseKey(testfix.WGKey(testfix.VPS2.Key))
	endpoint := func() string {
		t.Helper()
		dev, _ := r.h.DeviceOf("wg0")
		for _, p := range dev.Peers {
			if p.PublicKey == vps2 {
				return p.Endpoint.String()
			}
		}
		return ""
	}
	if moved, err := e.Failover(); err != nil || moved != nil {
		t.Fatalf("at once = %v, %v", moved, err)
	}
	now = now.Add(46 * time.Second)
	if moved, err := e.Failover(); err != nil || !slices.Equal(moved, []string{"vps-2 to " + testfix.VPS2.Endpoint6}) || endpoint() != testfix.VPS2.Endpoint6 {
		t.Fatalf("no handshake on IPv4 = %v, %v, endpoint %s", moved, err, endpoint())
	}
	if res, err := e.Apply(ctx, cfg); err != nil || res.Changes != 0 || endpoint() != testfix.VPS2.Endpoint6 {
		t.Fatalf("a re-apply = %+v, %v, endpoint %s", res, err, endpoint())
	}
	// It answers on IPv6, then stops.
	r.h.Stats = []wg.PeerStats{{PublicKey: vps2, LastHandshake: now.Add(time.Second)}}
	now = now.Add(2 * time.Minute)
	if moved, _ := e.Failover(); moved != nil {
		t.Fatalf("an answering server moved: %v", moved)
	}
	now = now.Add(time.Minute + time.Second)
	if moved, _ := e.Failover(); !slices.Equal(moved, []string{"vps-2 to " + testfix.VPS2.Endpoint}) || endpoint() != testfix.VPS2.Endpoint {
		t.Fatalf("stale on IPv6 = %v, endpoint %s", moved, endpoint())
	}
	now = now.Add(time.Minute)
	r.h.Fail = "Configure"
	if _, err := e.Failover(); err == nil || !strings.Contains(err.Error(), "dial vps-2") {
		t.Fatalf("a failed move = %v", err)
	}
	r.h.Fail = ""
	// Without a fallback there is nothing to fail over to.
	cfg.Peers[1].Fallback = ""
	if _, err := e.Apply(ctx, cfg); err != nil {
		t.Fatal(err)
	}
	now = now.Add(time.Hour)
	if moved, err := e.Failover(); err != nil || moved != nil || endpoint() != testfix.VPS2.Endpoint {
		t.Fatalf("without a fallback = %v, %v, endpoint %s", moved, err, endpoint())
	}
}

func TestDNSFailureOnlyDegrades(t *testing.T) {
	r := newRig(t)
	r.dns.Err = errors.New("dns: answer on 10.77.0.1:53: address already in use")
	cfg := testfix.ServerConfig()
	cfg.DNS = meshDNS()
	res, err := r.e.Apply(context.Background(), cfg)
	if err != nil || !slices.Equal(res.Degraded, []string{apply.StepDNS}) || !strings.Contains(res.Errors[0], "already in use") || res.DNS != "" {
		t.Fatalf("Apply = %+v, %v; want the tunnel up and DNS degraded", res, err)
	}
}

// `meshctl dns off` reaches the step on the next apply; exit mode tells it
// to take every lookup.
func TestDNSFollowsDNSOffAndExitMode(t *testing.T) {
	r := newRig(t)
	cfg := testfix.SpokeConfig("vps-2")
	cfg.DNS = meshDNS()
	r.e.SetDNSOff(true)
	res, err := r.e.Apply(context.Background(), cfg)
	w, _ := r.dns.Last()
	if err != nil || res.DNS != proto.DNSOff || !w.Off || !w.Exit {
		t.Fatalf("Apply = %+v, %v; dns step asked %+v", res, err, w)
	}
	r.e.SetDNSOff(false)
	if res, err = r.e.Apply(context.Background(), cfg); err != nil || res.DNS != proto.DNSOn || res.Changes != 1 {
		t.Fatalf("dns on = %+v, %v", res, err)
	}
}

// Reassert makes the steps that remember the host look again, then applies:
// an interface deleted underneath the agent comes back.
func TestReassertPutsBackWhatDrifted(t *testing.T) {
	r := newRig(t)
	cfg := testfix.ServerConfig()
	if _, err := r.e.Apply(context.Background(), cfg); err != nil {
		t.Fatal(err)
	}
	res, err := r.e.Reassert(context.Background(), cfg)
	if err != nil || res.Changes != 0 || r.fw.Rechecks != 1 {
		t.Fatalf("reassert of an unchanged host = %+v, %v, rechecks %d", res, err, r.fw.Rechecks)
	}
	r.h.RemoveLink("wg0")
	if res, err = r.e.Reassert(context.Background(), cfg); err != nil || res.Changes == 0 {
		t.Fatalf("reassert after wg0 went = %+v, %v", res, err)
	}
	if dev, ok := r.h.DeviceOf("wg0"); !ok || len(dev.Peers) != 2 {
		t.Fatal("wg0 did not come back with its peers")
	}
	r.e.Close(context.Background())
	if !r.dns.IsClosed() {
		t.Fatal("Close did not stop the DNS step")
	}
}

// Restart drops every peer and adds it back, so each tunnel handshakes
// anew, without taking the interface or its routes down; a missing wg0 is
// rebuilt, a foreign one refused, a failed drop reported.
func TestRestartRenewsEveryPeer(t *testing.T) {
	r := newRig(t)
	cfg := testfix.ServerConfig()
	if _, err := r.e.Apply(context.Background(), cfg); err != nil {
		t.Fatal(err)
	}
	routes := r.r.Mutations()
	r.h.Calls = nil
	res, err := r.e.Restart(context.Background(), cfg)
	if err != nil || res.Changes != 4 || r.fw.Rechecks != 1 {
		t.Fatalf("Restart = %+v, %v (rechecks %d); want 2 peers dropped and 2 added back", res, err, r.fw.Rechecks)
	}
	if !slices.Equal(r.h.Calls, []string{"Configure wg0 (2 changes)", "Configure wg0 (2 changes)"}) || r.r.Mutations() != routes {
		t.Fatalf("host calls %q, route changes %d -> %d; want the peers dropped and added back, the routes left", r.h.Calls, routes, r.r.Mutations())
	}
	if dev, ok := r.h.DeviceOf("wg0"); !ok || len(dev.Peers) != 2 {
		t.Fatalf("after the restart wg0 has %+v", dev)
	}

	r.h.RemoveLink("wg0")
	if res, err := r.e.Restart(context.Background(), cfg); err != nil || res.Changes == 0 {
		t.Fatalf("Restart without wg0 = %+v, %v", res, err)
	}
	r.h.SetFail("Configure")
	var se *apply.StepError
	if _, err := r.e.Restart(context.Background(), cfg); !errors.As(err, &se) || se.Step != apply.StepPeers || !strings.Contains(err.Error(), "restart the tunnel") {
		t.Fatalf("a failed drop = %v", err)
	}

	f := newRig(t)
	f.h.AddForeignLink("wg0")
	if _, err := f.e.Restart(context.Background(), cfg); !errors.Is(err, wg.ErrNotOurs) || f.h.Mutations() != 0 {
		t.Fatalf("Restart over a foreign wg0 = %v, %d host calls", err, f.h.Mutations())
	}
}

func TestApplyValidatesBeforeTouchingTheHost(t *testing.T) {
	r := newRig(t)
	cfg := testfix.ServerConfig()
	cfg.Interface.Address = "10.77.0.1/24"
	cfg.Sysctl = []proto.Sysctl{{Key: "kernel.core_pattern", Value: "1"}}
	_, err := r.e.Apply(context.Background(), cfg)
	if step, ok := apply.IsStepError(err); !ok || step != apply.StepValidate {
		t.Fatalf("Apply = %v, want a validate failure", err)
	}
	if r.mutations() != 0 || len(r.fw.Applied) != 0 {
		t.Fatalf("an invalid config touched the host: %v %v", r.h.Calls, r.r.Calls)
	}
}

func TestApplyStopsAtTheFailingStep(t *testing.T) {
	r := newRig(t)
	r.h.Fail = "Configure"
	_, err := r.e.Apply(context.Background(), testfix.ServerConfig())
	if step, ok := apply.IsStepError(err); !ok || step != apply.StepPeers || !strings.Contains(err.Error(), "apply step peers") {
		t.Fatalf("Apply = %v", err)
	}
	if r.r.Mutations() != 0 || len(r.fw.Applied) != 0 {
		t.Fatal("steps after the failing one ran")
	}
	var se *apply.StepError
	if !errors.As(err, &se) || errors.Unwrap(se) == nil {
		t.Fatal("the step error must keep its cause")
	}
}

func TestFirewallFailureOnlyDegrades(t *testing.T) {
	r := newRig(t)
	r.fw.Err = errors.New("firewalld is not running")
	res, err := r.e.Apply(context.Background(), testfix.ServerConfig())
	if err != nil || !slices.Equal(res.Degraded, []string{apply.StepFirewall}) || len(res.Errors) != 1 || !strings.Contains(res.Errors[0], "firewalld") {
		t.Fatalf("Apply = %+v, %v; want the tunnel up and the firewall degraded", res, err)
	}
	if dev, ok := r.h.DeviceOf("wg0"); !ok || len(dev.Peers) != 2 {
		t.Fatal("the peers must be in place despite the firewall")
	}
}

// watchingFirewall is a firewall step that notes whether the mesh
// firewall's table was in the kernel when it ran.
type watchingFirewall struct {
	k     *agentfake.Kernel
	saw   []bool
	fails bool
}

func (w *watchingFirewall) Apply(context.Context, *proto.Firewall) (int, error) {
	table, _ := w.k.Holds()
	w.saw = append(w.saw, table != nil)
	if w.fails {
		return 0, errors.New("firewalld is not running")
	}
	return 0, nil
}

// The mesh firewall (amendment A43): in the kernel before the firewall step
// lets the mesh through the mesh zone, left alone when unchanged, fatal when
// it fails; a config without a policy takes the table away once the
// firewall step closed the zone, and keeps it when that failed.
func TestMeshFirewallStep(t *testing.T) {
	r := newRig(t)
	w := &watchingFirewall{k: r.k}
	e := apply.New(apply.Deps{Host: r.h, Routes: r.r, Exit: r.p, Sysctl: r.s, MeshFirewall: &policy.Enforcer{K: r.k}, Firewall: w,
		Iface: "wg0", Key: r.key})
	ctx := context.Background()
	cfg := testfix.ServerConfig()
	cfg.Policy = testfix.Policy()
	if res, err := e.Apply(ctx, cfg); err != nil || res.Changes == 0 || r.k.Replaces != 1 || !slices.Equal(w.saw, []bool{true}) {
		t.Fatalf("first apply = %+v, %v; %d replaces, the firewall saw a table: %v", res, err, r.k.Replaces, w.saw)
	}
	if res, err := e.Apply(ctx, cfg); err != nil || res.Changes != 0 || r.k.Replaces != 1 {
		t.Fatalf("the same config again = %+v, %v; %d replaces", res, err, r.k.Replaces)
	}
	changed := testfix.ServerConfig()
	changed.Policy = testfix.Policy()
	changed.Policy.Rules = changed.Policy.Rules[:3]
	if _, err := e.Apply(ctx, changed); err != nil || r.k.Replaces != 2 || r.k.Forgets != 1 {
		t.Fatalf("a new policy: %v; %d replaces, %d forgets", err, r.k.Replaces, r.k.Forgets)
	}

	// Without a policy: removed after the firewall step, which still saw it.
	res, err := e.Apply(ctx, testfix.ServerConfig())
	if table, _ := r.k.Holds(); err != nil || table != nil || res.Changes == 0 || !w.saw[len(w.saw)-1] {
		t.Fatalf("no policy = %+v, %v; the table %v", res, err, table)
	}
	// ... but kept while the firewall step fails, which degrades alone.
	r.k.Mark, w.fails = "kept", true
	if res, err = e.Apply(ctx, testfix.ServerConfig()); err != nil || !slices.Equal(res.Degraded, []string{apply.StepFirewall}) || r.k.Mark != "kept" {
		t.Fatalf("firewall down = %+v, %v; mark %q", res, err, r.k.Mark)
	}
	w.fails, r.k.RemoveErr = false, errors.New("device busy")
	if res, err = e.Apply(ctx, testfix.ServerConfig()); err != nil || !slices.Equal(res.Degraded, []string{apply.StepPolicy}) ||
		!strings.Contains(res.Errors[0], "remove the table meshctl") {
		t.Fatalf("a failed removal = %+v, %v", res, err)
	}

	// A table the kernel refuses stops the apply before the firewall step.
	runs := len(w.saw)
	r.k.ReplaceErr = errors.New("no buffer space available")
	if _, err := e.Apply(ctx, cfg); !isStep(err, apply.StepPolicy) || !strings.Contains(err.Error(), "install the table meshctl") || len(w.saw) != runs {
		t.Fatalf("a refused table = %v; the firewall ran %d more times", err, len(w.saw)-runs)
	}
	// A node that enforces nothing refuses a policy; one outside its mesh
	// is not a config to apply.
	spoke := apply.New(apply.Deps{Host: agentfake.NewHost(), Routes: agentfake.NewRoutes(), Exit: &agentfake.Policy{}, Sysctl: &agentfake.Sysctl{},
		Iface: "wg0", Key: r.key})
	if _, err := spoke.Apply(ctx, cfg); !isStep(err, apply.StepPolicy) || !strings.Contains(err.Error(), "only servers") {
		t.Fatalf("a policy without a mesh firewall = %v", err)
	}
	elsewhere := testfix.ServerConfig()
	elsewhere.Policy = &proto.Policy{Mesh: "10.99.0.0/16"}
	if _, err := e.Apply(ctx, elsewhere); !isStep(err, apply.StepValidate) || !strings.Contains(err.Error(), "outside the mesh") {
		t.Fatalf("a policy for another mesh = %v", err)
	}
}

// A scheduled rule is in the table for the time of the apply, and the
// result says when to apply again; applying then installs what applies
// then. A config without a schedule has no next time.
func TestScheduledPolicy(t *testing.T) {
	r := newRig(t)
	now := time.Date(2026, 10, 19, 10, 0, 0, 0, time.UTC) // a Monday
	e := apply.New(apply.Deps{Host: r.h, Routes: r.r, Exit: r.p, Sysctl: r.s, MeshFirewall: &policy.Enforcer{K: r.k}, Iface: "wg0", Key: r.key,
		Now: func() time.Time { return now }})
	ctx := context.Background()
	cfg := testfix.ServerConfig()
	cfg.Policy = testfix.Policy()
	cfg.Policy.TZ = "UTC"
	cfg.Policy.Rules[3].Schedule = &proto.Schedule{Start: "09:00", End: "18:00"}
	res, err := e.Apply(ctx, cfg)
	table, _ := r.k.Holds()
	if err != nil || table == nil || !strings.Contains(table.Text(), "policy: laptop-smb") || !res.PolicyNext.Equal(now.Add(8*time.Hour)) {
		t.Fatalf("in the window = %+v, %v", res, err)
	}
	now = res.PolicyNext
	res, err = e.Apply(ctx, cfg)
	table, _ = r.k.Holds()
	if err != nil || res.Changes == 0 || strings.Contains(table.Text(), "policy: laptop-smb") ||
		!res.PolicyNext.Equal(time.Date(2026, 10, 20, 9, 0, 0, 0, time.UTC)) {
		t.Fatalf("after the window = %+v, %v", res, err)
	}
	if res, err := e.Apply(ctx, testfix.ServerConfig()); err != nil || !res.PolicyNext.IsZero() {
		t.Fatalf("no policy = %+v, %v", res, err)
	}
}

func isStep(err error, step string) bool {
	s, ok := apply.IsStepError(err)
	return ok && s == step
}

// Exit on installs the policy rules; switching exits only moves 0.0.0.0/0
// between peers, leaving the rules (and every connection) alone; exit off
// removes them. 0.0.0.0/0 never reaches the main table.
// The container network (amendment A37) is step 9: before DNS, which then
// answers the containers on their gateway too; a failure only degrades,
// and an agent without a container host says so.
func TestContainersStep(t *testing.T) {
	cfg := testfix.ContainerSpokeConfig()
	cfg.DNS = meshDNS()
	r := newRig(t)
	n := agentfake.NewNet()
	n.Attach("eth1", netip.MustParsePrefix("10.77.128.128/16"))
	r.e = apply.New(apply.Deps{Host: r.h, Routes: r.r, Exit: r.p, Sysctl: r.s, DNS: r.dns, Containers: n, Iface: "wg0", Key: r.key})
	res, err := r.e.Apply(context.Background(), cfg)
	if err != nil || len(res.Degraded) != 0 || !n.HasBlackhole() {
		t.Fatalf("Apply = %+v, %v; calls %v", res, err, n.Calls)
	}
	if w, _ := r.dns.Last(); w.Also != netip.MustParseAddr(cfg.Containers.Gateway) {
		t.Fatalf("the resolver must also answer on the gateway: %+v", w)
	}
	if res, err = r.e.Apply(context.Background(), cfg); err != nil || res.Changes != 0 {
		t.Fatalf("second Apply = %+v, %v", res, err)
	}

	// Off the mesh network: degraded, and DNS answers on the mesh IP only.
	off := agentfake.NewNet()
	r.e = apply.New(apply.Deps{Host: r.h, Routes: r.r, Exit: r.p, Sysctl: r.s, DNS: r.dns, Containers: off, Iface: "wg0", Key: r.key})
	res, err = r.e.Apply(context.Background(), cfg)
	if err != nil || !slices.Equal(res.Degraded, []string{apply.StepContainers}) || !strings.Contains(strings.Join(res.Errors, " "), "meshctl-mesh") {
		t.Fatalf("off the mesh network = %+v, %v", res, err)
	}
	if w, _ := r.dns.Last(); w.Also.IsValid() {
		t.Fatalf("the gateway is not up, yet DNS was asked to answer on it: %+v", w)
	}

	// No container host at all.
	r.e = apply.New(apply.Deps{Host: r.h, Routes: r.r, Exit: r.p, Sysctl: r.s, DNS: r.dns, Iface: "wg0", Key: r.key})
	if res, err = r.e.Apply(context.Background(), cfg); err != nil || !slices.Equal(res.Degraded, []string{apply.StepContainers}) {
		t.Fatalf("without a container host = %+v, %v", res, err)
	}
	if !slices.Contains(proto.DegradableSteps, apply.StepContainers) {
		t.Fatal("the server would refuse a status report naming the containers step")
	}
}

func TestExitModeOnSwitchOff(t *testing.T) {
	r := newRig(t)
	allowed := func(peer string) []netip.Prefix {
		dev, _ := r.h.DeviceOf("wg0")
		k, _ := wg.ParseKey(testfix.WGKey(map[string]byte{"vps-1": 1, "vps-2": 2, "vps-3": 3}[peer]))
		for _, p := range dev.Peers {
			if p.PublicKey == k {
				return p.AllowedIPs
			}
		}
		return nil
	}
	def := netip.MustParsePrefix("0.0.0.0/0")

	if _, err := r.e.Apply(context.Background(), testfix.SpokeConfig("vps-2")); err != nil {
		t.Fatal(err)
	}
	rules, _ := r.p.Rules()
	if len(rules) != 4 || !slices.Contains(allowed("vps-2"), def) || r.s.Values[proto.SysctlSrcValidMark] != "1" {
		t.Fatalf("exit on: rules %v, vps-2 %v, sysctls %v", rules, allowed("vps-2"), r.s.Values)
	}
	if routes, _ := r.r.List("wg0"); slices.Contains(routes, def) {
		t.Fatal("0.0.0.0/0 was routed in the main table")
	}
	before := len(r.p.Calls)
	if res, err := r.e.Apply(context.Background(), testfix.SpokeConfig("vps-3")); err != nil || res.Changes == 0 {
		t.Fatalf("switch = %+v, %v", res, err)
	}
	if len(r.p.Calls) != before || slices.Contains(allowed("vps-2"), def) || !slices.Contains(allowed("vps-3"), def) {
		t.Fatalf("switch: policy calls %v, vps-2 %v, vps-3 %v", r.p.Calls[before:], allowed("vps-2"), allowed("vps-3"))
	}
	if _, err := r.e.Apply(context.Background(), testfix.SpokeConfig("")); err != nil {
		t.Fatal(err)
	}
	if rules, _ := r.p.Rules(); len(rules) != 0 || slices.Contains(allowed("vps-3"), def) {
		t.Fatalf("exit off: rules %v, vps-3 %v", rules, allowed("vps-3"))
	}
}

func TestExitModeFailureIsFatal(t *testing.T) {
	r := newRig(t)
	r.p.Fail = errors.New("operation not permitted")
	_, err := r.e.Apply(context.Background(), testfix.SpokeConfig("vps-2"))
	if step, ok := apply.IsStepError(err); !ok || step != apply.StepExit || !strings.Contains(err.Error(), "not permitted") {
		t.Fatalf("Apply = %v", err)
	}
	if r.s.Writes != 0 || len(r.fw.Applied) != 0 {
		t.Fatal("steps after exit ran")
	}

	// An engine without policy routing refuses exit on, and applies exit off.
	k, _ := wg.ParseKey(testfix.WGKey(5))
	e := apply.New(apply.Deps{Host: agentfake.NewHost(), Routes: agentfake.NewRoutes(), Sysctl: &agentfake.Sysctl{}, Iface: "wg0", Key: k})
	if _, err := e.Apply(context.Background(), testfix.SpokeConfig("vps-2")); err == nil || !strings.Contains(err.Error(), "meshctl exit off") {
		t.Fatalf("Apply without a policy = %v", err)
	}
	if _, err := e.Apply(context.Background(), testfix.SpokeConfig("")); err != nil {
		t.Fatalf("exit off without a policy = %v", err)
	}
}

// A kernel with IPv6 but no IPv6 policy routing (amendment A45): a spoke
// without an exit applies as usual; with one, exit mode carries IPv4 and the
// apply is degraded, not refused.
func TestExitModeWithoutIPv6PolicyRouting(t *testing.T) {
	r := newRig(t)
	r.p.V6Unrouted = true
	if res, err := r.e.Apply(context.Background(), testfix.SpokeConfig("")); err != nil || len(res.Degraded) != 0 {
		t.Fatalf("no exit = %+v, %v", res, err)
	}
	res, err := r.e.Apply(context.Background(), testfix.SpokeConfig("vps-2"))
	rules, _ := r.p.Rules()
	if err != nil || !slices.Equal(res.Degraded, []string{apply.StepExit}) || !strings.Contains(res.Errors[0], "IPv4 only") || len(rules) != 2 {
		t.Fatalf("exit on = %+v, %v; rules %v", res, err, rules)
	}
}

// exitHost is a host that does exit mode itself, as the Windows one does
// (wg.ExitRouter): it records each call and changes only when asked to
// change.
type exitHost struct {
	*agentfake.Host
	on    bool
	calls []bool
	err   error
}

func (h *exitHost) SetExit(_ string, on bool) (int, error) {
	h.calls = append(h.calls, on)
	if h.err != nil {
		return 0, h.err
	}
	if h.on == on {
		return 0, nil
	}
	h.on = on
	return 1, nil
}

// Without policy routing, step 6 hands exit mode to the host when it does
// it: on with an exit, off without, nothing on a repeat, and a failure is
// fatal.
func TestExitModeThroughTheHost(t *testing.T) {
	k, err := wg.ParseKey(testfix.WGKey(5))
	if err != nil {
		t.Fatal(err)
	}
	h := &exitHost{Host: agentfake.NewHost()}
	e := apply.New(apply.Deps{Host: h, Routes: agentfake.NewRoutes(), Sysctl: &agentfake.Sysctl{}, Iface: "wg0", Key: k})
	if _, err := e.Apply(context.Background(), testfix.SpokeConfig("vps-2")); err != nil || !h.on {
		t.Fatalf("exit on = %v, host on %v", err, h.on)
	}
	res, err := e.Apply(context.Background(), testfix.SpokeConfig("vps-2"))
	if err != nil || res.Changes != 0 || !slices.Equal(h.calls, []bool{true, true}) {
		t.Fatalf("a repeat = %+v, %v, calls %v", res, err, h.calls)
	}
	if _, err := e.Apply(context.Background(), testfix.SpokeConfig("")); err != nil || h.on {
		t.Fatalf("exit off = %v, host on %v", err, h.on)
	}
	h.err = errors.New("exit mode: this host has no IPv4 default route to carry the tunnel")
	var se *apply.StepError
	if _, err := e.Apply(context.Background(), testfix.SpokeConfig("vps-3")); !errors.As(err, &se) || se.Step != apply.StepExit {
		t.Fatalf("a failing host = %v", err)
	}
}

func TestRemovedPeerLeavesDeviceAndRoutes(t *testing.T) {
	r := newRig(t)
	cfg := testfix.ServerConfig()
	if _, err := r.e.Apply(context.Background(), cfg); err != nil {
		t.Fatal(err)
	}
	cfg.Peers = cfg.Peers[:1]
	cfg.Firewall.Rules = cfg.Firewall.Rules[1:]
	if _, err := r.e.Apply(context.Background(), cfg); err != nil {
		t.Fatal(err)
	}
	dev, _ := r.h.DeviceOf("wg0")
	routes, _ := r.r.List("wg0")
	if len(dev.Peers) != 1 || !slices.Equal(routes, []netip.Prefix{netip.MustParsePrefix("10.77.0.2/32")}) {
		t.Fatalf("after removing vps-3: peers %+v, routes %v", dev.Peers, routes)
	}
}

func TestStatsNamesKnownPeers(t *testing.T) {
	r := newRig(t)
	cfg := testfix.ServerConfig()
	if _, err := r.e.Apply(context.Background(), cfg); err != nil {
		t.Fatal(err)
	}
	k2, _ := wg.ParseKey(testfix.WGKey(2))
	k9, _ := wg.ParseKey(testfix.WGKey(9))
	r.h.Stats = []wg.PeerStats{
		{PublicKey: k2, Endpoint: netip.MustParseAddrPort(testfix.VPS2.Endpoint), LastHandshake: testfix.Epoch, RxBytes: 5, TxBytes: 6},
		{PublicKey: k9, LastHandshake: time.Time{}},
	}
	got, err := r.e.Stats(cfg)
	if err != nil || len(got) != 1 {
		t.Fatalf("Stats = %+v, %v", got, err)
	}
	if p := got[0]; p.Name != "vps-2" || p.LastHandshake != testfix.Epoch.Unix() || p.Endpoint != testfix.VPS2.Endpoint || p.RxBytes != 5 {
		t.Fatalf("Stats = %+v", p)
	}
}
