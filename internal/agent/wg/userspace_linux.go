//go:build linux

package wg

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"sync"

	"github.com/vishvananda/netlink"
	"golang.zx2c4.com/wireguard/conn"
	"golang.zx2c4.com/wireguard/device"
	"golang.zx2c4.com/wireguard/ipc"
	"golang.zx2c4.com/wireguard/tun"
)

// tunPath is the device the userspace backend needs.
const tunPath = "/dev/net/tun"

// userspace is the Host for wireguard-go: the agent itself runs WireGuard
// behind a tun interface it creates, configured in the cross-platform
// protocol (UAPISet, ParseUAPI). netlink still owns the interface's MTU,
// state and addresses.
//
// The interface lives as long as the agent: a tun device goes away with the
// process that opened it, so the mesh is down while the agent restarts (the
// kernel backend keeps wg0 up).
type userspace struct {
	links
	log *slog.Logger

	mu   sync.Mutex
	name string
	dev  *device.Device
	uapi net.Listener
}

// NewUserspace opens the userspace backend. It fails at once, not at the
// first apply, when the tun device cannot be opened.
func NewUserspace(log *slog.Logger) (Host, error) {
	f, err := os.OpenFile(tunPath, os.O_RDWR, 0)
	if err != nil {
		return nil, fmt.Errorf("userspace WireGuard needs %s: %w (run as root; a container needs the device, which `meshctl node compose` adds)", tunPath, err)
	}
	_ = f.Close() // only a probe; CreateLink opens its own
	return &userspace{links: links{kind: "tuntap"}, log: log}, nil
}

func (u *userspace) Link(name string) (Link, error) {
	l, err := u.links.Link(name)
	if err != nil || !l.Exists {
		return l, err
	}
	u.mu.Lock()
	defer u.mu.Unlock()
	// A tun interface this process did not open is someone else's, alias or
	// not: another agent that is still running.
	l.Ours = l.Ours && u.dev != nil && u.name == name
	return l, nil
}

func (u *userspace) CreateLink(name string) error {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.closeLocked() // a device whose interface was deleted under it
	tdev, err := tun.CreateTUN(name, device.DefaultMTU)
	if err != nil {
		return fmt.Errorf("create userspace WireGuard interface %s: %w", name, err)
	}
	dev := device.NewDevice(tdev, conn.NewDefaultBind(), u.logger(name))
	l, err := lookup(name)
	if err == nil {
		err = netlink.LinkSetAlias(l, Alias)
	}
	if err != nil {
		dev.Close() // closing the tun removes the interface
		return fmt.Errorf("mark interface %s as meshctl's: %w", name, err)
	}
	u.name, u.dev = name, dev
	u.listen(name, dev)
	return nil
}

// logger hands wireguard-go's log to slog: its verbose lines at debug, its
// errors (a peer that does not answer, say) as warnings.
func (u *userspace) logger(name string) *device.Logger {
	log := u.log.With("wireguard", name)
	return &device.Logger{
		Verbosef: func(format string, args ...any) {
			if log.Enabled(context.Background(), slog.LevelDebug) {
				log.Debug(fmt.Sprintf(format, args...))
			}
		},
		Errorf: func(format string, args ...any) { log.Warn(fmt.Sprintf(format, args...)) },
	}
}

// listen serves the device on WireGuard's UAPI socket
// (/var/run/wireguard/<name>.sock, root only), so `wg show` lists it. It is
// best effort: where /var/run is read-only, as in the container, only `wg`
// loses sight of the interface; `meshctl status` still shows its peers.
func (u *userspace) listen(name string, dev *device.Device) {
	f, err := ipc.UAPIOpen(name)
	if err != nil {
		u.log.Info("no UAPI socket: `wg show` will not list the interface (`meshctl status` does)", "interface", name, "err", err)
		return
	}
	ln, err := ipc.UAPIListen(name, f)
	_ = f.Close() // the listener holds its own copy
	if err != nil {
		u.log.Info("no UAPI socket: `wg show` will not list the interface (`meshctl status` does)", "interface", name, "err", err)
		return
	}
	u.uapi = ln
	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return // closed with the device
			}
			go dev.IpcHandle(c)
		}
	}()
}

func (u *userspace) device(name string) (*device.Device, error) {
	u.mu.Lock()
	defer u.mu.Unlock()
	if u.dev == nil || u.name != name {
		return nil, fmt.Errorf("%s is not a WireGuard interface of this agent", name)
	}
	return u.dev, nil
}

func (u *userspace) Device(name string) (Device, []PeerStats, error) {
	dev, err := u.device(name)
	if err != nil {
		return Device{}, nil, err
	}
	s, err := dev.IpcGet()
	if err != nil {
		return Device{}, nil, fmt.Errorf("read WireGuard device %s: %w", name, err)
	}
	return ParseUAPI(s)
}

func (u *userspace) Configure(name string, c Change) error {
	dev, err := u.device(name)
	if err != nil {
		return err
	}
	if err := dev.IpcSet(UAPISet(c)); err != nil {
		return fmt.Errorf("configure WireGuard device %s: %w", name, err)
	}
	return nil
}

func (u *userspace) Close() error {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.closeLocked()
	return nil
}

func (u *userspace) closeLocked() {
	if u.uapi != nil {
		_ = u.uapi.Close() // it only stops `wg show` from reaching a closing device
		u.uapi = nil
	}
	if u.dev != nil {
		u.dev.Close()
		u.dev = nil
	}
	u.name = ""
}
