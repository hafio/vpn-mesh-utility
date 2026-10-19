package dns

import (
	"context"
	"errors"
	"net"
	"net/netip"
	"strings"
	"sync"
	"testing"
	"time"

	mdns "github.com/miekg/dns"

	"github.com/hafio/vpn-mesh-utility/internal/proto"
)

func meshDNS(split bool, upstream ...string) *proto.DNS {
	return &proto.DNS{Zone: proto.DefaultZone, Split: split, Upstream: upstream, ForwardFor: []string{"10.77.3.1/32"},
		Records: []proto.Record{{Name: "nas", IP: "10.77.1.1"}, {Name: "printer.office", IP: "10.77.1.200"}, {Name: "vps-1", IP: "10.77.0.1"}}}
}

func query(name string, qtype uint16) *mdns.Msg {
	m := new(mdns.Msg)
	m.SetQuestion(name, qtype)
	return m
}

func TestZoneAnswers(t *testing.T) {
	z, err := newZone(meshDNS(true, "1.1.1.1"), 53)
	if err != nil {
		t.Fatal(err)
	}
	laptop, phone := netip.MustParseAddr("10.77.2.1"), netip.MustParseAddr("10.77.3.1")
	cases := []struct {
		name    string
		q       *mdns.Msg
		src     netip.Addr
		rcode   int
		answer  string
		soa     bool
		forward bool
	}{
		{"a record", query("nas.mesh.", mdns.TypeA), laptop, mdns.RcodeSuccess, "nas.mesh.\t60\tIN\tA\t10.77.1.1", false, false},
		{"case kept", query("NAS.Mesh.", mdns.TypeA), laptop, mdns.RcodeSuccess, "NAS.Mesh.\t60\tIN\tA\t10.77.1.1", false, false},
		{"deep name", query("printer.office.mesh.", mdns.TypeANY), laptop, mdns.RcodeSuccess, "printer.office.mesh.\t60\tIN\tA\t10.77.1.200", false, false},
		{"other type", query("nas.mesh.", mdns.TypeAAAA), laptop, mdns.RcodeSuccess, "", true, false},
		{"empty non-terminal", query("office.mesh.", mdns.TypeA), laptop, mdns.RcodeSuccess, "", true, false},
		{"unknown name", query("ghost.mesh.", mdns.TypeA), laptop, mdns.RcodeNameError, "", true, false},
		{"apex soa", query("mesh.", mdns.TypeSOA), laptop, mdns.RcodeSuccess, "mesh.\t60\tIN\tSOA\tmesh. hostmaster.mesh. 1 3600 600 86400 60", false, false},
		{"another zone, split", query("example.com.", mdns.TypeA), laptop, mdns.RcodeRefused, "", false, false},
		{"another zone, a phone", query("example.com.", mdns.TypeA), phone, 0, "", false, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			resp, fwd := z.answer(c.q, c.src)
			if fwd != c.forward {
				t.Fatalf("forward = %v, want %v", fwd, c.forward)
			}
			if fwd {
				if resp != nil {
					t.Fatal("a forwarded query has an answer of its own")
				}
				return
			}
			got := ""
			if len(resp.Answer) == 1 {
				got = resp.Answer[0].String()
			}
			if resp.Rcode != c.rcode || got != c.answer || (len(resp.Ns) == 1) != c.soa || resp.Id != c.q.Id {
				t.Fatalf("answer = rcode %d, %q, %d authority, id %d", resp.Rcode, got, len(resp.Ns), resp.Id)
			}
			if c.rcode != mdns.RcodeRefused && !resp.Authoritative {
				t.Fatal("an answer from the mesh zone is authoritative")
			}
		})
	}

	// Exit mode forwards everyone's other lookups.
	exit, err := newZone(meshDNS(false), 53)
	if err != nil {
		t.Fatal(err)
	}
	if _, fwd := exit.answer(query("example.com.", mdns.TypeA), laptop); !fwd {
		t.Fatal("exit mode refused another zone")
	}
	// Malformed queries.
	chaos := query("nas.mesh.", mdns.TypeTXT)
	chaos.Question[0].Qclass = mdns.ClassCHAOS
	two := query("nas.mesh.", mdns.TypeA)
	two.Question = append(two.Question, two.Question[0])
	for q, want := range map[*mdns.Msg]int{chaos: mdns.RcodeRefused, two: mdns.RcodeFormatError} {
		if resp, _ := z.answer(q, laptop); resp == nil || resp.Rcode != want {
			t.Errorf("answer(%v) = %v, want rcode %d", q.Question, resp, want)
		}
	}
}

func TestNewZoneRejects(t *testing.T) {
	for name, edit := range map[string]func(*proto.DNS){
		"zone":        func(d *proto.DNS) { d.Zone = "office.local" },
		"record ip":   func(d *proto.DNS) { d.Records[0].IP = "2001:db8::1" },
		"upstream":    func(d *proto.DNS) { d.Upstream = []string{"one.one.one.one"} },
		"forward_for": func(d *proto.DNS) { d.ForwardFor = []string{"10.77.3.1"} },
	} {
		d := meshDNS(true, "1.1.1.1")
		edit(d)
		if _, err := newZone(d, 53); err == nil {
			t.Errorf("%s: accepted", name)
		}
	}
}

// The zone is the config's: another one answers its own names and leaves
// the default's to forwarding, and the registration names it.
func TestZoneFollowsTheConfig(t *testing.T) {
	d := meshDNS(true, "1.1.1.1")
	d.Zone = "home.arpa"
	z, err := newZone(d, 53)
	if err != nil {
		t.Fatal(err)
	}
	laptop := netip.MustParseAddr("10.77.2.1")
	if resp, fwd := z.answer(query("nas.home.arpa.", mdns.TypeA), laptop); fwd || len(resp.Answer) != 1 {
		t.Fatalf("nas.home.arpa = %v, forward %v", resp, fwd)
	}
	if resp, fwd := z.answer(query("nas.mesh.", mdns.TypeA), laptop); fwd || resp.Rcode != mdns.RcodeRefused {
		t.Fatalf("nas.mesh in the home.arpa zone = %v, forward %v", resp, fwd)
	}
	if r := desired(Want{IP: laptop, Config: d}); len(r.Domains) != 1 || r.Domains[0].Name != "home.arpa" {
		t.Fatalf("registration = %+v", r)
	}
	if rules := rulesFor(desired(Want{IP: laptop, Config: d, Exit: true})); len(rules) != 2 || rules[0].Name != ".home.arpa" {
		t.Fatalf("NRPT rules = %+v", rules)
	}
}

// fakeRegistrar is systemd-resolved: it keeps what was set per interface.
type fakeRegistrar struct {
	mu    sync.Mutex
	links map[string]Registration
	fail  map[string]error // by method
	calls []string
}

func newRegistrar() *fakeRegistrar {
	return &fakeRegistrar{links: map[string]Registration{}, fail: map[string]error{}}
}

func (f *fakeRegistrar) do(method string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls = append(f.calls, method)
	return f.fail[method]
}

func (f *fakeRegistrar) Get(_ context.Context, iface string) (Registration, error) {
	if err := f.do("get"); err != nil {
		return Registration{}, err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.links[iface], nil
}

func (f *fakeRegistrar) Set(_ context.Context, iface string, r Registration) error {
	if err := f.do("set"); err != nil {
		return err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.links[iface] = r
	return nil
}

func (f *fakeRegistrar) Revert(_ context.Context, iface string) error {
	if err := f.do("revert"); err != nil {
		return err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.links, iface)
	return nil
}

func (f *fakeRegistrar) link(iface string) Registration {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.links[iface]
}

// upstream is a resolver on loopback that knows one name.
func upstream(t *testing.T) (port int, stop func()) {
	t.Helper()
	h := mdns.HandlerFunc(func(w mdns.ResponseWriter, req *mdns.Msg) {
		resp := new(mdns.Msg).SetReply(req)
		if req.Question[0].Name == "example.com." {
			rr, _ := mdns.NewRR("example.com. 300 IN A 192.0.2.7")
			resp.Answer = []mdns.RR{rr}
		} else {
			resp.Rcode = mdns.RcodeNameError
		}
		_ = w.WriteMsg(resp)
	})
	r, err := serve(netip.MustParseAddrPort("127.0.0.1:0"), h)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(r.close)
	return int(r.addr.Port()), r.close
}

// ask queries the resolver at at. The client outwaits the resolver's own
// forwardTimeout: on Windows a dead UDP upstream is noticed only then (Go
// ignores the ICMP port unreachable there), not at once as on Linux.
func ask(t *testing.T, network string, at netip.AddrPort, name string) *mdns.Msg {
	t.Helper()
	c := &mdns.Client{Net: network, Timeout: forwardTimeout + 3*time.Second}
	resp, _, err := c.Exchange(query(name, mdns.TypeA), at.String())
	if err != nil {
		t.Fatalf("%s query for %s: %v", network, name, err)
	}
	return resp
}

func answerOf(m *mdns.Msg) string {
	if len(m.Answer) != 1 {
		return mdns.RcodeToString[m.Rcode]
	}
	return m.Answer[0].(*mdns.A).A.String()
}

// The resolver on a real socket: the zone over UDP and TCP, refusal and
// forwarding by mode, the registration following exit mode and `dns off`,
// and every repeat of the same apply changing nothing.
func TestServiceAnswersAndRegisters(t *testing.T) {
	up, stopUp := upstream(t)
	reg := newRegistrar()
	s := &Service{Registrar: reg, UpstreamPort: up}
	t.Cleanup(func() { s.Close(context.Background(), "wg0") })
	ctx, lo := context.Background(), netip.MustParseAddr("127.0.0.1")
	apply := func(w Want, wantState string, wantChanges int) {
		t.Helper()
		state, n, err := s.Apply(ctx, w)
		if err != nil || state != wantState || n != wantChanges {
			t.Fatalf("Apply(%+v) = %q, %d, %v; want %q, %d", w, state, n, err, wantState, wantChanges)
		}
	}
	split := Want{Iface: "wg0", IP: lo, Config: meshDNS(true, "127.0.0.1")}
	apply(split, proto.DNSOn, 3) // zone, resolver, registration
	apply(split, proto.DNSOn, 0)
	if r := reg.link("wg0"); !same(r, Registration{DNS: []netip.Addr{lo}, Domains: []Domain{{Name: "mesh", RouteOnly: true}}}) {
		t.Fatalf("registered %+v", r)
	}
	at := s.Addr()
	for _, network := range []string{"udp", "tcp"} {
		if got := answerOf(ask(t, network, at, "nas.mesh.")); got != "10.77.1.1" {
			t.Errorf("%s nas.mesh = %s", network, got)
		}
	}
	if got := answerOf(ask(t, "udp", at, "example.com.")); got != "REFUSED" {
		t.Errorf("split mode forwarded: %s", got)
	}

	exit := Want{Iface: "wg0", IP: lo, Config: meshDNS(false, "127.0.0.1"), Exit: true}
	apply(exit, proto.DNSOn, 2) // the zone's mode, the registration
	if r := reg.link("wg0"); !r.DefaultRoute || len(r.Domains) != 2 {
		t.Fatalf("exit mode registered %+v", r)
	}
	if got := answerOf(ask(t, "tcp", at, "example.com.")); got != "192.0.2.7" {
		t.Errorf("exit mode, forwarded = %s", got)
	}
	stopUp()
	if got := answerOf(ask(t, "udp", at, "example.com.")); got != "SERVFAIL" {
		t.Errorf("upstream gone = %s", got)
	}

	off := exit
	off.Off = true
	apply(off, proto.DNSOff, 1)
	apply(off, proto.DNSOff, 0)
	if _, ok := reg.links["wg0"]; ok {
		t.Fatal("dns off left the registration")
	}
	if got := answerOf(ask(t, "udp", at, "nas.mesh.")); got != "10.77.1.1" {
		t.Errorf("dns off stopped the resolver others use: %s", got)
	}
	apply(exit, proto.DNSOn, 1)

	// No DNS section: the resolver stops and the registration goes.
	apply(Want{Iface: "wg0", IP: lo}, "", 2)
	if s.Addr().IsValid() || len(reg.link("wg0").DNS) != 0 {
		t.Fatal("the resolver or the registration outlived the DNS section")
	}
	apply(Want{Iface: "wg0", IP: lo}, "", 0)
}

func TestServiceWithoutAResolverToRegisterWith(t *testing.T) {
	ctx, lo := context.Background(), netip.MustParseAddr("127.0.0.1")
	w := Want{Iface: "wg0", IP: lo, Config: meshDNS(true, "1.1.1.1")}
	unavailable := newRegistrar()
	unavailable.fail["get"] = errors.Join(ErrUnavailable, errors.New("systemd-resolved is not running"))
	for name, r := range map[string]Registrar{"none": nil, "resolved not running": unavailable} {
		s := &Service{Registrar: r}
		if state, _, err := s.Apply(ctx, w); err != nil || state != proto.DNSUnavailable {
			t.Errorf("%s: %q, %v", name, state, err)
		}
		off := w
		off.Off = true
		if state, _, err := s.Apply(ctx, off); err != nil || state != proto.DNSOff {
			t.Errorf("%s, dns off: %q, %v", name, state, err)
		}
		s.Close(ctx, "wg0")
	}
	if strings.Join(unavailable.calls, ",") != "get,get" {
		t.Fatalf("an unavailable resolver was written to: %v", unavailable.calls)
	}
}

func TestServiceRegistrationFailuresAreReported(t *testing.T) {
	ctx, lo := context.Background(), netip.MustParseAddr("127.0.0.1")
	for method, want := range map[string]string{"get": "read wg0's DNS", "set": "register wg0", "revert": "take wg0's DNS back"} {
		reg := newRegistrar()
		s := &Service{Registrar: reg}
		w := Want{Iface: "wg0", IP: lo, Config: meshDNS(true, "1.1.1.1")}
		if method == "revert" {
			if _, _, err := s.Apply(ctx, w); err != nil {
				t.Fatal(err)
			}
			w.Off = true
		}
		reg.fail[method] = errors.New("access denied")
		state, _, err := s.Apply(ctx, w)
		if err == nil || state != "" || !strings.Contains(err.Error(), want) || !strings.Contains(err.Error(), "access denied") {
			t.Errorf("%s failing: %q, %v", method, state, err)
		}
		// The resolver serves regardless.
		if !s.Addr().IsValid() {
			t.Errorf("%s failing stopped the resolver", method)
		}
		delete(reg.fail, method)
		s.Close(ctx, "wg0")
	}
}

// A port in use, or an address the host does not have, is reported with
// what to check; a resolver that died is started again on the next apply.
func TestServiceBinding(t *testing.T) {
	ctx := context.Background()
	taken, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer taken.Close()
	port := uint16(taken.LocalAddr().(*net.UDPAddr).Port)
	reg := newRegistrar()
	s := &Service{Registrar: reg, Port: port}
	if _, _, err := s.Apply(ctx, Want{Iface: "wg0", IP: netip.MustParseAddr("127.0.0.1"), Config: meshDNS(true)}); err == nil ||
		!strings.Contains(err.Error(), "another DNS server") {
		t.Fatalf("a taken port = %v", err)
	}
	s.Port = 0
	if _, _, err := s.Apply(ctx, Want{Iface: "wg0", IP: netip.MustParseAddr("192.0.2.1"), Config: meshDNS(true)}); err == nil {
		t.Fatal("bound an address this host does not have")
	}
	if len(reg.calls) != 0 {
		t.Fatalf("registered a resolver that does not answer: %v", reg.calls)
	}

	w := Want{Iface: "wg0", IP: netip.MustParseAddr("127.0.0.1"), Config: meshDNS(true)}
	if _, _, err := s.Apply(ctx, w); err != nil {
		t.Fatal(err)
	}
	s.mu.Lock()
	s.res.close() // as if its socket failed
	s.mu.Unlock()
	<-s.res.stopped
	if _, n, err := s.Apply(ctx, w); err != nil || n != 1 || answerOf(ask(t, "udp", s.Addr(), "vps-1.mesh.")) != "10.77.0.1" {
		t.Fatalf("restart after the resolver died: %d, %v", n, err)
	}
	s.Close(ctx, "wg0")
}

func TestSameRegistration(t *testing.T) {
	ip := []netip.Addr{netip.MustParseAddr("10.77.2.1")}
	a := Registration{DNS: ip, Domains: []Domain{{Name: ".", RouteOnly: true}, {Name: "mesh", RouteOnly: true}}, DefaultRoute: true}
	b := Registration{DNS: ip, Domains: []Domain{{Name: "MESH.", RouteOnly: true}, {Name: "", RouteOnly: true}}, DefaultRoute: true}
	if !same(a, b) {
		t.Fatal("the same domains, spelled and ordered differently, differ")
	}
	for _, c := range []Registration{
		{DNS: ip, Domains: a.Domains},
		{DNS: ip, Domains: []Domain{{Name: "mesh", RouteOnly: false}, {Name: ".", RouteOnly: true}}, DefaultRoute: true},
		{Domains: a.Domains, DefaultRoute: true},
	} {
		if same(a, c) {
			t.Errorf("%+v is not %+v", c, a)
		}
	}
}
