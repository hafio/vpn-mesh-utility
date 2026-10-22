//go:build linux

package dns

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/netip"
	"os"
	"sync"
	"time"

	"github.com/godbus/dbus/v5"
)

// systemd-resolved's D-Bus API (org.freedesktop.resolve1(5)).
const (
	resolvedName = "org.freedesktop.resolve1"
	resolvedPath = dbus.ObjectPath("/org/freedesktop/resolve1")
	managerIface = "org.freedesktop.resolve1.Manager"
	linkIface    = "org.freedesktop.resolve1.Link"
	afInet       = 2  // AF_INET
	afInet6      = 10 // AF_INET6
	// busTimeout bounds each registrar call, so a hung resolved cannot
	// hold up the apply.
	busTimeout = 5 * time.Second
)

// resolved registers with systemd-resolved over the system bus. Calls do
// not start resolved where it is installed but not running: a host whose
// DNS is managed another way is left as it is.
type resolved struct {
	mu   sync.Mutex
	conn *dbus.Conn
}

// NewRegistrar returns the host's registrar: systemd-resolved.
func NewRegistrar() Registrar { return &resolved{} }

// linkDNS and linkDomain are resolved's a(iay) and a(sb).
type linkDNS struct {
	Family  int32
	Address []byte
}

type linkDomain struct {
	Name      string
	RouteOnly bool
}

// foreignNetns reports whether the agent runs in another network namespace
// than pid 1, the host's: resolved would read the interface index as one
// of the host's own interfaces. Where pid 1's namespace cannot be read (the
// unit has no CAP_SYS_PTRACE) the agent is the host's service.
func foreignNetns() bool {
	self, err1 := os.Readlink("/proc/self/ns/net")
	host, err2 := os.Readlink("/proc/1/ns/net")
	return err1 == nil && err2 == nil && self != host
}

func (r *resolved) bus() (*dbus.Conn, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.conn != nil && r.conn.Connected() {
		return r.conn, nil
	}
	if foreignNetns() {
		return nil, fmt.Errorf("%w: the agent runs in a network namespace systemd-resolved does not see", ErrUnavailable)
	}
	c, err := dbus.ConnectSystemBus()
	if err != nil {
		return nil, fmt.Errorf("%w: no system bus (%v)", ErrUnavailable, err)
	}
	r.conn = c
	return c, nil
}

// classify names the failures that mean "resolved is not there".
func classify(err error) error {
	name := ""
	var ve dbus.Error
	var pe *dbus.Error
	switch {
	case errors.As(err, &ve):
		name = ve.Name
	case errors.As(err, &pe):
		name = pe.Name
	}
	switch name {
	case "org.freedesktop.DBus.Error.ServiceUnknown", "org.freedesktop.DBus.Error.NameHasNoOwner":
		return fmt.Errorf("%w: systemd-resolved is not running", ErrUnavailable)
	case "org.freedesktop.DBus.Error.AccessDenied", "org.freedesktop.DBus.Error.InteractiveAuthorizationRequired":
		return fmt.Errorf("systemd-resolved refused the agent (%s); it needs CAP_NET_ADMIN, which its unit grants: %w", name, err)
	}
	return err
}

// link returns the bus and resolved's index of iface.
func (r *resolved) link(iface string) (*dbus.Conn, int32, error) {
	c, err := r.bus()
	if err != nil {
		return nil, 0, err
	}
	ifc, err := net.InterfaceByName(iface)
	if err != nil {
		return nil, 0, fmt.Errorf("look up %s: %w", iface, err)
	}
	return c, int32(ifc.Index), nil
}

func (r *resolved) call(ctx context.Context, c *dbus.Conn, method string, args ...any) *dbus.Call {
	return c.Object(resolvedName, resolvedPath).CallWithContext(ctx, managerIface+"."+method, dbus.FlagNoAutoStart, args...)
}

func (r *resolved) Get(ctx context.Context, iface string) (Registration, error) {
	ctx, cancel := context.WithTimeout(ctx, busTimeout)
	defer cancel()
	c, idx, err := r.link(iface)
	if err != nil {
		return Registration{}, err
	}
	var path dbus.ObjectPath
	if err := r.call(ctx, c, "GetLink", idx).Store(&path); err != nil {
		return Registration{}, classify(err)
	}
	link := c.Object(resolvedName, path)
	var (
		addrs   []linkDNS
		domains []linkDomain
		reg     Registration
	)
	for _, p := range []struct {
		name string
		v    any
	}{{"DNS", &addrs}, {"Domains", &domains}, {"DefaultRoute", &reg.DefaultRoute}} {
		if err := link.CallWithContext(ctx, "org.freedesktop.DBus.Properties.Get", dbus.FlagNoAutoStart, linkIface, p.name).Store(p.v); err != nil {
			return Registration{}, fmt.Errorf("read %s of %s: %w", p.name, iface, classify(err))
		}
	}
	for _, a := range addrs {
		if ip, ok := netip.AddrFromSlice(a.Address); ok {
			reg.DNS = append(reg.DNS, ip.Unmap())
		}
	}
	for _, d := range domains {
		reg.Domains = append(reg.Domains, Domain(d))
	}
	return reg, nil
}

func (r *resolved) Set(ctx context.Context, iface string, reg Registration) error {
	ctx, cancel := context.WithTimeout(ctx, busTimeout)
	defer cancel()
	c, idx, err := r.link(iface)
	if err != nil {
		return err
	}
	addrs := []linkDNS{}
	for _, a := range reg.DNS {
		family := afInet
		if a.Is6() {
			family = afInet6 // an IPv6-only node's resolver (M8)
		}
		addrs = append(addrs, linkDNS{Family: int32(family), Address: a.AsSlice()})
	}
	domains := []linkDomain{}
	for _, d := range reg.Domains {
		domains = append(domains, linkDomain(d))
	}
	for _, m := range []struct {
		method string
		arg    any
	}{{"SetLinkDNS", addrs}, {"SetLinkDomains", domains}, {"SetLinkDefaultRoute", reg.DefaultRoute}} {
		if err := r.call(ctx, c, m.method, idx, m.arg).Err; err != nil {
			return fmt.Errorf("%s: %w", m.method, classify(err))
		}
	}
	// Best effort, as older versions lack some: the mesh zone is not
	// signed, and the mesh needs no multicast names.
	for _, m := range []string{"SetLinkDNSSEC", "SetLinkLLMNR", "SetLinkMulticastDNS"} {
		_ = r.call(ctx, c, m, idx, "no").Err
	}
	return nil
}

func (r *resolved) Revert(ctx context.Context, iface string) error {
	ctx, cancel := context.WithTimeout(ctx, busTimeout)
	defer cancel()
	c, idx, err := r.link(iface)
	if err != nil {
		return err
	}
	return classify(r.call(ctx, c, "RevertLink", idx).Err)
}
