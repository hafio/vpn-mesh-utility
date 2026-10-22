//go:build windows

package wg

import (
	"errors"
	"fmt"
	"log/slog"
	"math"
	"net/netip"
	"slices"

	"golang.org/x/sys/windows"
	"golang.zx2c4.com/wireguard/conn"
	"golang.zx2c4.com/wireguard/device"
	"golang.zx2c4.com/wireguard/tun"
	"golang.zx2c4.com/wireguard/windows/tunnel/winipcfg"

	"github.com/hafio/vpn-mesh-utility/internal/agent/routes"
	"github.com/hafio/vpn-mesh-utility/internal/agent/wintun"
)

// adapterGUID is the wintun adapter's GUID. It is fixed, so Windows sees
// the same network after every restart of the agent instead of a new one
// each time (amendment A34).
var adapterGUID = windows.GUID{Data1: 0x6d657368, Data2: 0x6374, Data3: 0x6c00, Data4: [8]byte{0x8a, 0x2f, 0x41, 0x7e, 0x0c, 0x5b, 0x93, 0xd1}}

// userspace is the Host on Windows: wireguard-go (wgdev) on a wintun
// adapter whose MTU and addresses go through the IP helper API (winipcfg).
// As with a Linux tun, the adapter lives as long as the agent.
type userspace struct {
	wgdev
	tun  *tun.NativeTun
	luid winipcfg.LUID
	bind conn.Bind
	// watch follows the host's default routes while exit mode is on;
	// exitIf4 and exitIf6 are the interfaces the tunnel's own packets are
	// bound to (0: none).
	watch            *winipcfg.RouteChangeCallback
	exitIf4, exitIf6 uint32
}

// NewUserspace opens the Windows backend: wintun.dll goes beside
// meshctl.exe first, checked against its pin, so the first apply can load
// it.
func NewUserspace(log *slog.Logger) (Host, error) {
	path, wrote, err := wintun.Ensure()
	if err != nil {
		return nil, err
	}
	if wrote {
		log.Info("wintun.dll written beside meshctl", "path", path, "version", wintun.Version)
	}
	return &userspace{wgdev: wgdev{log: log}}, nil
}

// Link reports the adapter this agent opened. Another program's adapter is
// never this agent's: it is not looked for, and the adapter this agent
// creates takes the name.
func (u *userspace) Link(name string) (Link, error) {
	u.mu.Lock()
	defer u.mu.Unlock()
	if u.dev == nil || u.name != name {
		return Link{}, nil
	}
	ipif, err := u.luid.IPInterface(windows.AF_INET)
	if err != nil {
		return Link{}, fmt.Errorf("read adapter %s: %w", name, err)
	}
	rows, err := winipcfg.GetUnicastIPAddressTable(windows.AF_UNSPEC)
	if err != nil {
		return Link{}, fmt.Errorf("read the addresses of adapter %s: %w", name, err)
	}
	// A wintun adapter is up while its session is open, that is, always.
	l := Link{Exists: true, Ours: true, Up: true, MTU: int(ipif.NLMTU)}
	for i := range rows {
		// Windows gives the adapter an IPv6 link-local address of its own.
		if a := rows[i].Address.Addr(); rows[i].InterfaceLUID == u.luid && !a.IsLinkLocalUnicast() {
			l.Addrs = append(l.Addrs, netip.PrefixFrom(a, int(rows[i].OnLinkPrefixLength)))
		}
	}
	return l, nil
}

func (u *userspace) CreateLink(name string) error {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.closeLocked()
	tdev, err := tun.CreateTUNWithRequestedGUID(name, &adapterGUID, device.DefaultMTU)
	if err != nil {
		return fmt.Errorf("create the WireGuard adapter %s: %w (it takes the service, or an elevated prompt)", name, err)
	}
	nt, ok := tdev.(*tun.NativeTun)
	if !ok {
		_ = tdev.Close() // not the wintun device this host drives
		return fmt.Errorf("create the WireGuard adapter %s: not a wintun device", name)
	}
	u.name, u.tun, u.luid, u.bind = name, nt, winipcfg.LUID(nt.LUID()), conn.NewDefaultBind()
	u.dev = device.NewDevice(tdev, u.bind, u.logger(name))
	// wintun never reports the adapter up, so nothing brings the device up
	// for us (wireguard-go's own Windows main calls Up itself); a device
	// left down sends nothing, not even a handshake.
	if err := u.dev.Up(); err != nil {
		u.closeLocked()
		return fmt.Errorf("bring up the WireGuard device on adapter %s: %w", name, err)
	}
	return nil
}

// adapter is the LUID of this agent's adapter called name.
func (u *userspace) adapter(name string) (winipcfg.LUID, *tun.NativeTun, error) {
	u.mu.Lock()
	defer u.mu.Unlock()
	if u.dev == nil || u.name != name {
		return 0, nil, fmt.Errorf("%s is not a WireGuard interface of this agent", name)
	}
	return u.luid, u.tun, nil
}

// SetMTU sets the adapter's MTU, IPv4 and IPv6 (where the host has IPv6),
// and the size wireguard-go packs to.
func (u *userspace) SetMTU(name string, mtu int) error {
	luid, nt, err := u.adapter(name)
	if err != nil {
		return err
	}
	ipif, err := luid.IPInterface(windows.AF_INET)
	if err == nil {
		ipif.NLMTU = uint32(mtu)
		err = ipif.Set()
	}
	if err != nil {
		return fmt.Errorf("set the MTU of adapter %s to %d: %w", name, mtu, err)
	}
	if ipif6, err := luid.IPInterface(windows.AF_INET6); err == nil {
		ipif6.NLMTU = uint32(mtu)
		if err := ipif6.Set(); err != nil {
			return fmt.Errorf("set the IPv6 MTU of adapter %s to %d: %w", name, mtu, err)
		}
	}
	nt.ForceMTU(mtu)
	return nil
}

// SetUp has nothing to do: Link reports the adapter up while it exists.
func (u *userspace) SetUp(string) error { return nil }

func (u *userspace) AddAddr(name string, p netip.Prefix) error {
	luid, _, err := u.adapter(name)
	if err == nil {
		err = luid.AddIPAddress(p)
	}
	if err != nil {
		return fmt.Errorf("add %s to adapter %s: %w", p, name, err)
	}
	return nil
}

func (u *userspace) DelAddr(name string, p netip.Prefix) error {
	luid, _, err := u.adapter(name)
	if err == nil {
		err = luid.DeleteIPAddress(p)
	}
	if err != nil {
		return fmt.Errorf("remove %s from adapter %s: %w", p, name, err)
	}
	return nil
}

func (u *userspace) Close() error {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.closeLocked()
	return nil
}

// closeLocked closes the device, which closes the adapter and its routes.
// The caller holds mu.
func (u *userspace) closeLocked() {
	if u.watch != nil {
		_ = u.watch.Unregister() // the adapter it guarded is going
		u.watch = nil
	}
	if u.dev != nil {
		u.dev.Close()
		u.dev = nil
	}
	u.name, u.tun, u.luid, u.bind, u.exitIf4, u.exitIf6 = "", nil, 0, nil, 0, 0
}

// SetExit turns exit mode on or off (wg.ExitRouter): the split default
// routes on the adapter, and the tunnel's own packets bound to the
// physical interface so they never loop into the tunnel (amendment A11).
func (u *userspace) SetExit(name string, on bool) (int, error) {
	luid, _, err := u.adapter(name)
	if err != nil {
		return 0, err
	}
	have, err := exitRoutes(luid)
	if err != nil {
		return 0, err
	}
	n := 0
	for _, p := range routes.ExitPrefixes {
		switch {
		case on && !have[p]:
			if err := luid.AddRoute(p, unspecified(p), 0); err != nil {
				return n, fmt.Errorf("exit mode: add route %s via %s: %w", p, name, err)
			}
			n++
		case !on && have[p]:
			if err := luid.DeleteRoute(p, unspecified(p)); err != nil {
				return n, fmt.Errorf("exit mode: remove route %s via %s: %w", p, name, err)
			}
			n++
		}
	}
	c, err := u.bindExit(on)
	return n + c, err
}

func unspecified(p netip.Prefix) netip.Addr {
	if p.Addr().Is4() {
		return netip.IPv4Unspecified()
	}
	return netip.IPv6Unspecified()
}

// exitRoutes are the split default routes on the adapter now.
func exitRoutes(luid winipcfg.LUID) (map[netip.Prefix]bool, error) {
	have := map[netip.Prefix]bool{}
	for _, family := range []winipcfg.AddressFamily{windows.AF_INET, windows.AF_INET6} {
		rows, err := winipcfg.GetIPForwardTable2(family)
		if err != nil {
			return nil, fmt.Errorf("exit mode: list routes: %w", err)
		}
		for i := range rows {
			r := &rows[i]
			if p := r.DestinationPrefix.Prefix(); r.InterfaceLUID == luid && slices.Contains(routes.ExitPrefixes, p) {
				have[p] = true
			}
		}
	}
	return have, nil
}

// bindExit binds the tunnel's sockets to the interfaces of the host's
// default routes, and binds them again whenever those routes move
// (another Wi-Fi network); off, it lets the sockets go back to routing.
func (u *userspace) bindExit(on bool) (int, error) {
	u.mu.Lock()
	defer u.mu.Unlock()
	b, ok := u.bind.(conn.BindSocketToInterface)
	switch {
	case !ok && on:
		return 0, errors.New("exit mode: this WireGuard socket cannot be bound to an interface")
	case !ok, on == (u.watch != nil):
		return 0, nil // nothing to change: the callback keeps an "on" bound
	case !on:
		_ = u.watch.Unregister() // it only rebinds a socket that is being released
		u.watch, u.exitIf4, u.exitIf6 = nil, 0, 0
		if err := errors.Join(b.BindSocketToInterface4(0, false), b.BindSocketToInterface6(0, false)); err != nil {
			return 1, fmt.Errorf("exit mode: release the tunnel's socket: %w", err)
		}
		return 1, nil
	}
	if err := u.bindLocked(b); err != nil {
		return 0, err
	}
	w, err := winipcfg.RegisterRouteChangeCallback(func(_ winipcfg.MibNotificationType, r *winipcfg.MibIPforwardRow2) {
		if r != nil && r.DestinationPrefix.PrefixLength == 0 {
			go u.rebind()
		}
	})
	if err != nil {
		// Back to routing; exit mode is not on.
		_ = errors.Join(b.BindSocketToInterface4(0, false), b.BindSocketToInterface6(0, false))
		return 0, fmt.Errorf("exit mode: watch the default route: %w", err)
	}
	u.watch = w
	return 1, nil
}

// bindLocked binds the tunnel's sockets to the interfaces of the host's
// default routes: the IPv4 one, which exit mode needs, and the IPv6 one
// (M8), whose tunnel packets are dropped while the host has no IPv6
// default route rather than sent into the tunnel itself. A host whose
// IPv6 socket cannot be bound (IPv6 off) keeps exit mode on IPv4. The
// caller holds mu.
func (u *userspace) bindLocked(b conn.BindSocketToInterface) error {
	idx4, err := defaultInterface(u.luid, windows.AF_INET)
	if err != nil {
		return err
	}
	if err := b.BindSocketToInterface4(idx4, false); err != nil {
		return fmt.Errorf("exit mode: bind the tunnel's socket to interface %d: %w", idx4, err)
	}
	idx6, _ := defaultInterface(u.luid, windows.AF_INET6) // none: 0, dropped
	if err := b.BindSocketToInterface6(idx6, idx6 == 0); err != nil {
		u.log.Warn("exit mode: the tunnel's IPv6 socket stays unbound; IPv6 endpoints may not work in exit mode", "interface", idx6, "err", err)
		idx6 = 0
	}
	u.exitIf4, u.exitIf6 = idx4, idx6
	return nil
}

// rebind follows the default routes to their new interfaces.
func (u *userspace) rebind() {
	u.mu.Lock()
	defer u.mu.Unlock()
	b, ok := u.bind.(conn.BindSocketToInterface)
	if u.watch == nil || !ok {
		return
	}
	idx4, err := defaultInterface(u.luid, windows.AF_INET)
	if err != nil {
		u.log.Warn("exit mode: the host has no default route now; the tunnel waits for one", "err", err)
		return
	}
	if idx6, _ := defaultInterface(u.luid, windows.AF_INET6); idx4 == u.exitIf4 && idx6 == u.exitIf6 {
		return
	}
	if err := u.bindLocked(b); err != nil {
		u.log.Error("exit mode: could not follow the default route to its new interface", "err", err)
		return
	}
	u.log.Info("exit mode: the host's default route moved; the tunnel follows it", "interface", u.exitIf4, "interface6", u.exitIf6)
}

// defaultInterface is the interface of the host's default route of family,
// leaving out the tunnel's own adapter: the lowest route plus interface
// metric, as Windows picks it.
func defaultInterface(tunnel winipcfg.LUID, family winipcfg.AddressFamily) (uint32, error) {
	name := "IPv4"
	if family == windows.AF_INET6 {
		name = "IPv6"
	}
	rows, err := winipcfg.GetIPForwardTable2(family)
	if err != nil {
		return 0, fmt.Errorf("exit mode: list %s routes: %w", name, err)
	}
	best, bestMetric := uint32(0), uint32(math.MaxUint32)
	for i := range rows {
		r := &rows[i]
		if r.InterfaceLUID == tunnel || r.DestinationPrefix.PrefixLength != 0 {
			continue
		}
		ipif, err := r.InterfaceLUID.IPInterface(family)
		if err != nil || !ipif.Connected {
			continue
		}
		if m := r.Metric + ipif.Metric; m < bestMetric {
			best, bestMetric = r.InterfaceIndex, m
		}
	}
	if best == 0 {
		return 0, fmt.Errorf("exit mode: this host has no %s default route to carry the tunnel", name)
	}
	return best, nil
}
