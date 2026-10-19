// Package wg is the agent's WireGuard interface (spec S7.1): the device
// types the apply engine works with, the pure diff between what the kernel
// has and what the config asks for, and the steps that apply it through a
// Host -- netlink and wgctrl on Linux (kernel_linux.go), a fake in tests.
//
// Peers are changed one by one with ReplacePeers off, so an apply that only
// touches one peer leaves every other handshake alone.
package wg

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net/netip"
	"slices"
	"time"

	"github.com/hafio/vpn-mesh-utility/internal/proto"
)

// Alias marks interfaces the agent created (amendment A18). The agent never
// adopts an interface without it, so a host's own wg-quick setup is safe.
const Alias = "meshctl"

// Key is a WireGuard key.
type Key [32]byte

// ParseKey decodes a base64 key.
func ParseKey(s string) (Key, error) {
	var k Key
	b, err := base64.StdEncoding.Strict().DecodeString(s)
	if err != nil || len(b) != len(k) {
		return k, fmt.Errorf("%q is not a base64 32-byte WireGuard key", s)
	}
	copy(k[:], b)
	return k, nil
}

func (k Key) String() string { return base64.StdEncoding.EncodeToString(k[:]) }

// Peer is one peer as configured or observed.
type Peer struct {
	PublicKey Key
	// Endpoint is the zero value for peers that dial in (spokes, phones).
	Endpoint   netip.AddrPort
	Keepalive  time.Duration
	AllowedIPs []netip.Prefix
}

// Device is the WireGuard side of the interface.
type Device struct {
	PrivateKey Key
	// ListenPort 0 lets the kernel pick (spokes dial out).
	ListenPort int
	FwMark     int
	Peers      []Peer
}

// PeerStats is the runtime state of one peer.
type PeerStats struct {
	PublicKey     Key
	Endpoint      netip.AddrPort
	LastHandshake time.Time
	RxBytes       int64
	TxBytes       int64
}

// Change turns an observed device into the desired one. Nil fields stay as
// they are.
type Change struct {
	PrivateKey *Key
	ListenPort *int
	FwMark     *int
	Peers      []PeerChange
}

// PeerChange adds, updates or removes one peer.
type PeerChange struct {
	PublicKey Key
	Remove    bool
	Endpoint  *netip.AddrPort
	Keepalive *time.Duration
	// ReplaceAllowedIPs sets the peer's list to AllowedIPs exactly.
	ReplaceAllowedIPs bool
	AllowedIPs        []netip.Prefix
}

// Count is the number of separate changes, for the engine's tally.
func (c Change) Count() int {
	n := len(c.Peers)
	for _, set := range []bool{c.PrivateKey != nil, c.ListenPort != nil, c.FwMark != nil} {
		if set {
			n++
		}
	}
	return n
}

// Desired is the device a config asks for.
func Desired(cfg proto.NodeConfig, priv Key) (Device, error) {
	d := Device{PrivateKey: priv, FwMark: cfg.Interface.FwMark}
	if cfg.Interface.ListenPort != nil {
		d.ListenPort = *cfg.Interface.ListenPort
	}
	var errs []error
	for _, p := range cfg.Peers {
		key, err := ParseKey(p.PublicKey)
		if err != nil {
			errs = append(errs, fmt.Errorf("peer %q: %w", p.Name, err))
			continue
		}
		peer := Peer{PublicKey: key, Keepalive: time.Duration(p.Keepalive) * time.Second}
		if p.Endpoint != "" {
			if peer.Endpoint, err = netip.ParseAddrPort(p.Endpoint); err != nil {
				errs = append(errs, fmt.Errorf("peer %q: endpoint: %w", p.Name, err))
			}
		}
		for _, s := range p.AllowedIPs {
			pf, err := netip.ParsePrefix(s)
			if err != nil {
				errs = append(errs, fmt.Errorf("peer %q: allowed ip: %w", p.Name, err))
				continue
			}
			peer.AllowedIPs = append(peer.AllowedIPs, pf)
		}
		SortPrefixes(peer.AllowedIPs)
		d.Peers = append(d.Peers, peer)
	}
	return d, errors.Join(errs...)
}

// Diff returns the change from observed to desired: only what differs.
// A dial-in peer (no desired endpoint) keeps whatever endpoint it roamed to,
// and a spoke's random listen port is left alone.
func Diff(observed, desired Device) Change {
	var c Change
	if observed.PrivateKey != desired.PrivateKey {
		k := desired.PrivateKey
		c.PrivateKey = &k
	}
	if desired.ListenPort != 0 && observed.ListenPort != desired.ListenPort {
		p := desired.ListenPort
		c.ListenPort = &p
	}
	if observed.FwMark != desired.FwMark {
		m := desired.FwMark
		c.FwMark = &m
	}
	have := map[Key]Peer{}
	for _, p := range observed.Peers {
		have[p.PublicKey] = p
	}
	want := map[Key]bool{}
	for _, d := range desired.Peers {
		want[d.PublicKey] = true
		o, ok := have[d.PublicKey]
		pc := PeerChange{PublicKey: d.PublicKey}
		changed := !ok
		if d.Endpoint.IsValid() && (!ok || o.Endpoint != d.Endpoint) {
			ep := d.Endpoint
			pc.Endpoint, changed = &ep, true
		}
		if !ok || o.Keepalive != d.Keepalive {
			ka := d.Keepalive
			pc.Keepalive, changed = &ka, true
		}
		if !ok || !samePrefixes(o.AllowedIPs, d.AllowedIPs) {
			pc.ReplaceAllowedIPs, pc.AllowedIPs, changed = true, d.AllowedIPs, true
		}
		if changed {
			c.Peers = append(c.Peers, pc)
		}
	}
	for _, o := range observed.Peers {
		if !want[o.PublicKey] {
			c.Peers = append(c.Peers, PeerChange{PublicKey: o.PublicKey, Remove: true})
		}
	}
	slices.SortFunc(c.Peers, func(a, b PeerChange) int { return compareKeys(a.PublicKey, b.PublicKey) })
	return c
}

func compareKeys(a, b Key) int {
	for i := range a {
		if a[i] != b[i] {
			return int(a[i]) - int(b[i])
		}
	}
	return 0
}

func samePrefixes(a, b []netip.Prefix) bool {
	a, b = slices.Clone(a), slices.Clone(b)
	SortPrefixes(a)
	SortPrefixes(b)
	return slices.Equal(a, b)
}

// SortPrefixes orders prefixes by address, then length.
func SortPrefixes(ps []netip.Prefix) {
	slices.SortFunc(ps, func(a, b netip.Prefix) int {
		if c := a.Addr().Compare(b.Addr()); c != 0 {
			return c
		}
		return a.Bits() - b.Bits()
	})
}

// Link is the observed state of the network interface.
type Link struct {
	Exists bool
	// Ours means the interface carries Alias.
	Ours  bool
	MTU   int
	Up    bool
	Addrs []netip.Prefix
}

// Host is what the apply engine needs from the operating system.
type Host interface {
	Link(name string) (Link, error)
	// CreateLink creates a WireGuard interface marked with Alias.
	CreateLink(name string) error
	SetMTU(name string, mtu int) error
	SetUp(name string) error
	AddAddr(name string, p netip.Prefix) error
	DelAddr(name string, p netip.Prefix) error
	// Device reads the WireGuard configuration and peer statistics.
	Device(name string) (Device, []PeerStats, error)
	Configure(name string, c Change) error
	Close() error
}

// ExitRouter is a Host that does exit mode itself, where the platform has
// no policy routing for it (Windows): on, everything but the mesh's own
// routes goes into the tunnel; off, that stops. It says how many changes it
// made, so a repeat is a no-op.
type ExitRouter interface {
	SetExit(name string, on bool) (int, error)
}

// ErrNotOurs means an interface of that name exists but the agent did not
// create it.
var ErrNotOurs = errors.New("interface exists but was not created by meshctl")

// SyncLink makes the interface exist, be ours, have mtu and be up. It returns
// how many changes it made.
func SyncLink(h Host, name string, mtu int) (int, error) {
	l, err := h.Link(name)
	if err != nil {
		return 0, err
	}
	n := 0
	switch {
	case !l.Exists:
		if err := h.CreateLink(name); err != nil {
			return 0, err
		}
		n++
		if l, err = h.Link(name); err != nil {
			return n, err
		}
	case !l.Ours:
		return 0, fmt.Errorf("%s: %w (no %q alias); if it is a wg-quick interface, stop and disable it (`wg-quick down %s`) first",
			name, ErrNotOurs, Alias, name)
	}
	if l.MTU != mtu {
		if err := h.SetMTU(name, mtu); err != nil {
			return n, err
		}
		n++
	}
	if !l.Up {
		if err := h.SetUp(name); err != nil {
			return n, err
		}
		n++
	}
	return n, nil
}

// SyncAddr makes p the interface's only IPv4 address.
func SyncAddr(h Host, name string, p netip.Prefix) (int, error) {
	l, err := h.Link(name)
	if err != nil {
		return 0, err
	}
	n, has := 0, false
	for _, a := range l.Addrs {
		if a == p {
			has = true
			continue
		}
		if err := h.DelAddr(name, a); err != nil {
			return n, err
		}
		n++
	}
	if !has {
		if err := h.AddAddr(name, p); err != nil {
			return n, err
		}
		n++
	}
	return n, nil
}

// SyncDevice applies the difference between the device and desired.
func SyncDevice(h Host, name string, desired Device) (int, error) {
	observed, _, err := h.Device(name)
	if err != nil {
		return 0, err
	}
	c := Diff(observed, desired)
	if c.Count() == 0 {
		return 0, nil
	}
	if err := h.Configure(name, c); err != nil {
		return 0, err
	}
	return c.Count(), nil
}
