//go:build windows

package wg

import (
	"fmt"
	"log/slog"
	"net/netip"

	"golang.org/x/sys/windows"
	"golang.zx2c4.com/wireguard/conn"
	"golang.zx2c4.com/wireguard/device"
	"golang.zx2c4.com/wireguard/tun"
	"golang.zx2c4.com/wireguard/windows/tunnel/winipcfg"

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
	rows, err := winipcfg.GetUnicastIPAddressTable(windows.AF_INET)
	if err != nil {
		return Link{}, fmt.Errorf("read the addresses of adapter %s: %w", name, err)
	}
	// A wintun adapter is up while its session is open, that is, always.
	l := Link{Exists: true, Ours: true, Up: true, MTU: int(ipif.NLMTU)}
	for i := range rows {
		if rows[i].InterfaceLUID == u.luid {
			l.Addrs = append(l.Addrs, netip.PrefixFrom(rows[i].Address.Addr(), int(rows[i].OnLinkPrefixLength)))
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
	u.name, u.tun, u.luid = name, nt, winipcfg.LUID(nt.LUID())
	u.dev = device.NewDevice(tdev, conn.NewDefaultBind(), u.logger(name))
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

// SetMTU sets the adapter's IPv4 MTU and the size wireguard-go packs to.
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

// closeLocked closes the device, which closes the adapter. The caller
// holds mu.
func (u *userspace) closeLocked() {
	if u.dev != nil {
		u.dev.Close()
		u.dev = nil
	}
	u.name, u.tun, u.luid = "", nil, 0
}
