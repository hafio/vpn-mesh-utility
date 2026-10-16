//go:build linux

package wg

import (
	"errors"
	"fmt"
	"net"
	"os"

	"github.com/vishvananda/netlink"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"

	"github.com/hafio/vpn-mesh-utility/internal/agent/netconv"
)

// kernel is the Host for in-kernel WireGuard: netlink creates the link and
// its addresses, wgctrl configures keys and peers.
type kernel struct {
	links
	wg *wgctrl.Client
}

// NewKernel opens the kernel backend.
func NewKernel() (Host, error) {
	c, err := wgctrl.New()
	if err != nil {
		return nil, fmt.Errorf("open WireGuard control: %w", err)
	}
	return &kernel{links: links{kind: "wireguard"}, wg: c}, nil
}

// DetectBackend reports the kernel backend when the WireGuard module is
// loaded or can be auto-loaded by creating a link, else userspace (spec
// S7.1).
func DetectBackend() string {
	if _, err := os.Stat("/sys/module/wireguard"); err == nil {
		return BackendKernel
	}
	la := netlink.NewLinkAttrs()
	la.Name = "meshctl-probe"
	probe := &netlink.Wireguard{LinkAttrs: la}
	if err := netlink.LinkAdd(probe); err != nil {
		return BackendUserspace
	}
	_ = netlink.LinkDel(probe) // a leftover probe link is harmless and removed on the next probe
	return BackendKernel
}

func (k *kernel) Close() error { return k.wg.Close() }

func (k *kernel) CreateLink(name string) error {
	la := netlink.NewLinkAttrs()
	la.Name = name
	l := &netlink.Wireguard{LinkAttrs: la}
	if err := netlink.LinkAdd(l); err != nil {
		return fmt.Errorf("create WireGuard interface %s: %w (is the wireguard kernel module available? try `modprobe wireguard`, or run the agent with --wg-backend userspace)", name, err)
	}
	if err := netlink.LinkSetAlias(l, Alias); err != nil {
		_ = netlink.LinkDel(l) // without the alias the agent would refuse it next time
		return fmt.Errorf("mark interface %s as meshctl's: %w", name, err)
	}
	return nil
}

func (k *kernel) Device(name string) (Device, []PeerStats, error) {
	d, err := k.wg.Device(name)
	if errors.Is(err, os.ErrNotExist) {
		return Device{}, nil, fmt.Errorf("%s is not a WireGuard interface", name)
	}
	if err != nil {
		return Device{}, nil, fmt.Errorf("read WireGuard device %s: %w", name, err)
	}
	out := Device{PrivateKey: Key(d.PrivateKey), ListenPort: d.ListenPort, FwMark: d.FirewallMark}
	stats := make([]PeerStats, 0, len(d.Peers))
	for _, p := range d.Peers {
		peer := Peer{PublicKey: Key(p.PublicKey), Endpoint: netconv.AddrPort(p.Endpoint), Keepalive: p.PersistentKeepaliveInterval}
		for _, a := range p.AllowedIPs {
			if pf, ok := netconv.Prefix(a); ok {
				peer.AllowedIPs = append(peer.AllowedIPs, pf)
			}
		}
		SortPrefixes(peer.AllowedIPs)
		out.Peers = append(out.Peers, peer)
		stats = append(stats, PeerStats{PublicKey: peer.PublicKey, Endpoint: peer.Endpoint,
			LastHandshake: p.LastHandshakeTime, RxBytes: p.ReceiveBytes, TxBytes: p.TransmitBytes})
	}
	return out, stats, nil
}

func (k *kernel) Configure(name string, c Change) error {
	cfg := wgtypes.Config{ListenPort: c.ListenPort, FirewallMark: c.FwMark}
	if c.PrivateKey != nil {
		key := wgtypes.Key(*c.PrivateKey)
		cfg.PrivateKey = &key
	}
	for _, p := range c.Peers {
		pc := wgtypes.PeerConfig{PublicKey: wgtypes.Key(p.PublicKey), Remove: p.Remove, ReplaceAllowedIPs: p.ReplaceAllowedIPs}
		if p.Endpoint != nil {
			pc.Endpoint = net.UDPAddrFromAddrPort(*p.Endpoint)
		}
		if p.Keepalive != nil {
			ka := *p.Keepalive
			pc.PersistentKeepaliveInterval = &ka
		}
		for _, a := range p.AllowedIPs {
			pc.AllowedIPs = append(pc.AllowedIPs, netconv.IPNet(a))
		}
		cfg.Peers = append(cfg.Peers, pc)
	}
	if err := k.wg.ConfigureDevice(name, cfg); err != nil {
		return fmt.Errorf("configure WireGuard device %s: %w", name, err)
	}
	return nil
}
