package wg_test

import (
	"errors"
	"net/netip"
	"strings"
	"testing"

	"github.com/hafio/vpn-mesh-utility/internal/agent/agentfake"
	"github.com/hafio/vpn-mesh-utility/internal/agent/wg"
	"github.com/hafio/vpn-mesh-utility/internal/proto"
	"github.com/hafio/vpn-mesh-utility/internal/testfix"
)

func key(n byte) wg.Key {
	k, err := wg.ParseKey(testfix.WGKey(n))
	if err != nil {
		panic(err)
	}
	return k
}

func TestParseKey(t *testing.T) {
	if k := key(7); k.String() != testfix.WGKey(7) {
		t.Fatalf("round trip = %s", k)
	}
	for _, s := range []string{"", "abc", testfix.WGKey(1)[:43], "!" + testfix.WGKey(1)[1:]} {
		if _, err := wg.ParseKey(s); err == nil {
			t.Errorf("ParseKey(%q) accepted", s)
		}
	}
}

func TestDesiredFromServerConfig(t *testing.T) {
	cfg := testfix.ServerConfig()
	d, err := wg.Desired(cfg, key(1))
	if err != nil {
		t.Fatal(err)
	}
	if d.ListenPort != 51820 || d.FwMark != 0 || len(d.Peers) != 2 || d.Peers[0].PublicKey != key(2) ||
		d.Peers[0].Endpoint != netip.MustParseAddrPort(testfix.VPS2.Endpoint) || d.Peers[0].Keepalive != 0 {
		t.Fatalf("Desired = %+v", d)
	}
	cfg.Peers[1].PublicKey = "bad"
	if _, err := wg.Desired(cfg, key(1)); err == nil {
		t.Fatal("a bad peer key was accepted")
	}
}

func TestDiffChangesOnlyWhatDiffers(t *testing.T) {
	desired, err := wg.Desired(testfix.ServerConfig(), key(1))
	if err != nil {
		t.Fatal(err)
	}
	if c := wg.Diff(desired, desired); c.Count() != 0 {
		t.Fatalf("Diff of equal devices = %+v", c)
	}

	observed := desired
	observed.Peers = []wg.Peer{
		{PublicKey: key(2), Endpoint: netip.MustParseAddrPort("198.51.100.1:51820"), AllowedIPs: desired.Peers[0].AllowedIPs},
		{PublicKey: key(9), AllowedIPs: []netip.Prefix{netip.MustParsePrefix("10.77.9.9/32")}},
	}
	c := wg.Diff(observed, desired)
	if c.PrivateKey != nil || c.ListenPort != nil || c.FwMark != nil || len(c.Peers) != 3 {
		t.Fatalf("Diff = %+v", c)
	}
	byKey := map[wg.Key]wg.PeerChange{}
	for _, p := range c.Peers {
		byKey[p.PublicKey] = p
	}
	if p := byKey[key(2)]; p.Endpoint == nil || p.Keepalive != nil || p.ReplaceAllowedIPs || p.Remove {
		t.Errorf("vps-2 must only get its endpoint back: %+v", p)
	}
	if p := byKey[key(3)]; !p.ReplaceAllowedIPs || p.Endpoint == nil || p.Keepalive == nil {
		t.Errorf("vps-3 is new and must be fully configured: %+v", p)
	}
	if p := byKey[key(9)]; !p.Remove {
		t.Errorf("an unknown peer must be removed: %+v", p)
	}
}

func TestDiffLeavesARoamingDialInPeerAlone(t *testing.T) {
	spoke := wg.Peer{PublicKey: key(4), AllowedIPs: []netip.Prefix{netip.MustParsePrefix("10.77.1.1/32")}}
	desired := wg.Device{PrivateKey: key(1), ListenPort: 51820, Peers: []wg.Peer{spoke}}
	observed := desired
	roamed := spoke
	roamed.Endpoint = netip.MustParseAddrPort("100.64.3.4:40000")
	observed.Peers = []wg.Peer{roamed}
	if c := wg.Diff(observed, desired); c.Count() != 0 {
		t.Fatalf("a spoke's learned endpoint must be kept: %+v", c)
	}
	// A spoke's random port is not reset either.
	spokeDev := wg.Device{PrivateKey: key(4), FwMark: proto.FwMark}
	if c := wg.Diff(wg.Device{PrivateKey: key(4), FwMark: proto.FwMark, ListenPort: 38211}, spokeDev); c.Count() != 0 {
		t.Fatalf("Diff = %+v", c)
	}
}

func TestSyncIsIdempotent(t *testing.T) {
	h := agentfake.NewHost()
	desired, err := wg.Desired(testfix.ServerConfig(), key(1))
	if err != nil {
		t.Fatal(err)
	}
	addr := netip.MustParsePrefix("10.77.0.1/32")
	steps := func() int {
		n1, err1 := wg.SyncLink(h, "wg0", 1420)
		n2, err2 := wg.SyncAddr(h, "wg0", addr)
		n3, err3 := wg.SyncDevice(h, "wg0", desired)
		if err := errors.Join(err1, err2, err3); err != nil {
			t.Fatal(err)
		}
		return n1 + n2 + n3
	}
	if n := steps(); n == 0 {
		t.Fatal("first sync changed nothing")
	}
	before := h.Mutations()
	if n := steps(); n != 0 || h.Mutations() != before {
		t.Fatalf("second sync: %d changes, calls %v", n, h.Calls[before:])
	}
	dev, _ := h.DeviceOf("wg0")
	if c := wg.Diff(dev, desired); c.Count() != 0 {
		t.Fatalf("the host does not hold the desired device: %+v", c)
	}
	// A second address (left by hand) is removed; the MTU follows the config.
	if err := h.AddAddr("wg0", netip.MustParsePrefix("10.99.0.1/32")); err != nil {
		t.Fatal(err)
	}
	if n, err := wg.SyncAddr(h, "wg0", addr); err != nil || n != 1 {
		t.Fatalf("SyncAddr = %d, %v", n, err)
	}
	if n, err := wg.SyncLink(h, "wg0", 1380); err != nil || n != 1 {
		t.Fatalf("SyncLink with a new MTU = %d, %v", n, err)
	}
}

func TestSyncLinkRefusesAForeignInterface(t *testing.T) {
	h := agentfake.NewHost()
	h.AddForeignLink("wg0")
	_, err := wg.SyncLink(h, "wg0", 1420)
	if !errors.Is(err, wg.ErrNotOurs) || !strings.Contains(err.Error(), "wg-quick down wg0") {
		t.Fatalf("SyncLink = %v", err)
	}
	if h.Mutations() != 0 {
		t.Fatalf("a foreign interface was changed: %v", h.Calls)
	}
}

func TestSyncSurfacesHostFailures(t *testing.T) {
	for _, method := range []string{"CreateLink", "SetUp", "Configure"} {
		h := agentfake.NewHost()
		h.Fail = method
		desired, _ := wg.Desired(testfix.ServerConfig(), key(1))
		_, err1 := wg.SyncLink(h, "wg0", 1420)
		var err2 error
		if err1 == nil {
			_, err2 = wg.SyncDevice(h, "wg0", desired)
		}
		if errors.Join(err1, err2) == nil {
			t.Errorf("%s failure was swallowed", method)
		}
	}
}
