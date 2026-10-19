// Package dns is the agent's DNS (spec S7.4, amendment A12): a resolver on
// the node's mesh IP that answers the mesh zone from the synced records, and
// the host's split-DNS registration that sends .mesh lookups to it.
//
// The resolver answers A records of the zone, NODATA for other types of a
// name it has and NXDOMAIN for one it has not. A query for another zone is
// forwarded to the upstream resolvers only in exit mode (every lookup of
// the node goes through the tunnel then) or from a phone that uses this
// node as its resolver (forward_for); otherwise it is REFUSED, so the
// resolver is never an open forwarder.
//
// The registration is systemd-resolved's per-link DNS on Linux: the mesh IP
// as wg0's server and ~mesh as its routing domain, plus ~. and default-route
// in exit mode. On Windows it is two NRPT rules, .mesh and, in exit mode,
// every name (nrpt_windows.go). Nothing else on the host changes:
// /etc/hosts and /etc/resolv.conf are never written. Without a host
// resolver to register with the resolver still serves the mesh and the
// state is "unavailable"; `meshctl dns off` takes the registration back and
// leaves the resolver serving the others.
package dns

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/netip"
	"reflect"
	"slices"
	"strings"
	"sync"
	"sync/atomic"

	mdns "github.com/miekg/dns"

	"github.com/hafio/vpn-mesh-utility/internal/proto"
)

// ErrUnavailable means there is nothing to register with: no
// systemd-resolved, no system bus, or a network namespace the host's
// resolver cannot see.
var ErrUnavailable = errors.New("no host resolver to register with")

// Registration is what the host's resolver holds for the mesh interface.
type Registration struct {
	DNS          []netip.Addr
	Domains      []Domain
	DefaultRoute bool
}

// Domain is one of a link's domains; RouteOnly marks a routing domain
// ("~mesh"), which sends lookups without adding a search suffix.
type Domain struct {
	Name      string
	RouteOnly bool
}

// Registrar registers the resolver with the host.
type Registrar interface {
	// Get returns what the host holds for iface now.
	Get(ctx context.Context, iface string) (Registration, error)
	Set(ctx context.Context, iface string, r Registration) error
	// Revert drops everything set for iface.
	Revert(ctx context.Context, iface string) error
}

// Want is what one apply asks of the DNS step.
type Want struct {
	Iface string
	// IP is the node's mesh IP, which the resolver binds.
	IP netip.Addr
	// Config is the config's DNS section; nil stops the resolver and takes
	// the registration back.
	Config *proto.DNS
	// Exit is exit mode: every lookup of the host goes to the resolver.
	Exit bool
	// Off is the node's own `meshctl dns off`.
	Off bool
}

// Service is the DNS step: the resolver and the registration.
type Service struct {
	// Registrar is nil where there is nothing to register with.
	Registrar Registrar
	// Port is the resolver's port and UpstreamPort the upstreams': New sets
	// both to 53; a zero Port picks a free one (tests).
	Port         uint16
	UpstreamPort int
	Log          *slog.Logger

	mu         sync.Mutex
	res        *resolver
	last       *proto.DNS
	registered bool
	zone       atomic.Pointer[zone]
}

// New returns the service for this host.
func New(r Registrar, log *slog.Logger) *Service {
	return &Service{Registrar: r, Port: 53, UpstreamPort: 53, Log: log}
}

func (s *Service) log() *slog.Logger {
	if s.Log == nil {
		return slog.New(slog.DiscardHandler)
	}
	return s.Log
}

// Addr is where the resolver answers; zero while it does not.
func (s *Service) Addr() netip.AddrPort {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.res == nil {
		return netip.AddrPort{}
	}
	return s.res.addr
}

// Apply makes the resolver and the registration match w and returns the
// state (proto.DNSStates, "" without a DNS section) and how many changes it
// made. Like every apply step it changes only what differs.
func (s *Service) Apply(ctx context.Context, w Want) (string, int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if w.Config == nil {
		return "", s.stopLocked(ctx, w.Iface), nil
	}
	z, err := newZone(w.Config, s.UpstreamPort)
	if err != nil {
		return "", 0, err
	}
	n := 0
	if !reflect.DeepEqual(s.last, w.Config) {
		s.zone.Store(z)
		cfg := *w.Config
		s.last = &cfg
		n++
	}
	want := netip.AddrPortFrom(w.IP, s.Port)
	if s.res == nil || s.res.dead() || s.res.want != want {
		if s.res != nil {
			s.res.close()
			s.res = nil
		}
		r, err := serve(want, s)
		if err != nil {
			return "", n, fmt.Errorf("dns: answer on %s: %w (is another DNS server using port %d on this address?)", want, err, s.Port)
		}
		s.res = r
		n++
		s.log().Info("resolver answering the mesh zone", "addr", r.addr)
	}
	state, c, err := s.register(ctx, w)
	return state, n + c, err
}

// desired is the registration for w: the resolver for the mesh zone, and in
// exit mode for every name.
func desired(w Want) Registration {
	zone := proto.DefaultZone
	if w.Config != nil {
		zone = w.Config.Zone
	}
	r := Registration{DNS: []netip.Addr{w.IP}, Domains: []Domain{{Name: zone, RouteOnly: true}}}
	if w.Exit {
		r.Domains = append(r.Domains, Domain{Name: ".", RouteOnly: true})
		r.DefaultRoute = true
	}
	return r
}

// domainName is a domain as systemd-resolved may report it ("mesh",
// "mesh." and "MESH" are one domain; "" and "." are the root).
func domainName(s string) string {
	s = strings.ToLower(strings.TrimSuffix(s, "."))
	if s == "" {
		return "."
	}
	return s
}

// same compares registrations as sets; DefaultRoute counts only where it
// was asked for, since resolved derives it for a link nobody configured.
func same(a, b Registration) bool {
	norm := func(r Registration) ([]string, []string) {
		var dns, doms []string
		for _, ip := range r.DNS {
			dns = append(dns, ip.Unmap().String())
		}
		for _, d := range r.Domains {
			doms = append(doms, fmt.Sprintf("%s/%v", domainName(d.Name), d.RouteOnly))
		}
		slices.Sort(dns)
		slices.Sort(doms)
		return dns, doms
	}
	ad, am := norm(a)
	bd, bm := norm(b)
	return slices.Equal(ad, bd) && slices.Equal(am, bm) && a.DefaultRoute == b.DefaultRoute
}

func (s *Service) register(ctx context.Context, w Want) (string, int, error) {
	if s.Registrar == nil {
		if w.Off {
			return proto.DNSOff, 0, nil
		}
		return proto.DNSUnavailable, 0, nil
	}
	cur, err := s.Registrar.Get(ctx, w.Iface)
	switch {
	case errors.Is(err, ErrUnavailable) && w.Off:
		return proto.DNSOff, 0, nil
	case errors.Is(err, ErrUnavailable):
		s.log().Debug("split DNS unavailable; .mesh resolves only through the mesh IP", "reason", err)
		return proto.DNSUnavailable, 0, nil
	case err != nil:
		return "", 0, fmt.Errorf("dns: read %s's DNS from the host's resolver: %w", w.Iface, err)
	}
	if w.Off {
		if len(cur.DNS) == 0 && len(cur.Domains) == 0 {
			s.registered = false
			return proto.DNSOff, 0, nil
		}
		if err := s.Registrar.Revert(ctx, w.Iface); err != nil {
			return "", 0, fmt.Errorf("dns: take %s's DNS back from the host's resolver: %w", w.Iface, err)
		}
		s.registered = false
		return proto.DNSOff, 1, nil
	}
	want := desired(w)
	if same(cur, want) {
		s.registered = true
		return proto.DNSOn, 0, nil
	}
	if err := s.Registrar.Set(ctx, w.Iface, want); err != nil {
		return "", 0, fmt.Errorf("dns: register %s with the host's resolver: %w", w.Iface, err)
	}
	s.registered = true
	return proto.DNSOn, 1, nil
}

// stopLocked stops the resolver and takes the registration back; it
// returns how many of the two it did.
func (s *Service) stopLocked(ctx context.Context, iface string) int {
	n := 0
	if s.res != nil {
		s.res.close()
		s.res, s.last = nil, nil
		s.zone.Store(nil)
		n++
	}
	if s.registered && s.Registrar != nil {
		if err := s.Registrar.Revert(ctx, iface); err != nil {
			s.log().Warn("could not take the DNS registration back; `resolvectl revert "+iface+"` does", "err", err)
		} else {
			n++
		}
		s.registered = false
	}
	return n
}

// Close stops the resolver and takes the registration back, when the
// agent stops: lookups of .mesh then fail at once instead of waiting on a
// resolver that is gone.
func (s *Service) Close(ctx context.Context, iface string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.stopLocked(ctx, iface)
}

// ServeDNS answers one query from the current zone.
func (s *Service) ServeDNS(w mdns.ResponseWriter, req *mdns.Msg) {
	var src netip.Addr
	if ap, err := netip.ParseAddrPort(w.RemoteAddr().String()); err == nil {
		src = ap.Addr().Unmap()
	}
	z := s.zone.Load()
	if z == nil {
		_ = w.WriteMsg(new(mdns.Msg).SetRcode(req, mdns.RcodeServerFailure)) // the client retries
		return
	}
	resp, forward := z.answer(req, src)
	if forward {
		var errs []error
		resp, errs = exchange(req, z.upstream, w.LocalAddr().Network())
		if len(errs) > 0 {
			s.log().Debug("upstream DNS did not answer", "name", req.Question[0].Name, "errs", errors.Join(errs...))
		}
	} else if req.IsEdns0() != nil {
		resp.SetEdns0(1232, false)
	}
	_ = w.WriteMsg(resp) // the client went away; nothing to do
}
