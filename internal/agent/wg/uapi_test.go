package wg_test

import (
	"bytes"
	"encoding/hex"
	"flag"
	"log/slog"
	"net/netip"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/hafio/vpn-mesh-utility/internal/agent/wg"
)

var update = flag.Bool("update", false, "rewrite testdata goldens")

func hexOf(k wg.Key) string { return hex.EncodeToString(k[:]) }

// A change of every kind: device fields, a new server peer, a removed peer,
// a dial-in spoke and a roamed endpoint.
func uapiChange() wg.Change {
	priv, port, mark := key(1), 51820, 51820
	ep2 := netip.MustParseAddrPort("203.0.113.12:51820")
	ep5 := netip.MustParseAddrPort("[2001:db8::5]:51820")
	ka25, ka0 := 25*time.Second, time.Duration(0)
	return wg.Change{PrivateKey: &priv, ListenPort: &port, FwMark: &mark, Peers: []wg.PeerChange{
		{PublicKey: key(2), Endpoint: &ep2, Keepalive: &ka0, ReplaceAllowedIPs: true, AllowedIPs: pfxs("10.77.0.2/32", "10.77.2.0/24")},
		{PublicKey: key(3), Remove: true},
		{PublicKey: key(4), Keepalive: &ka25, ReplaceAllowedIPs: true, AllowedIPs: pfxs("10.77.1.1/32")},
		{PublicKey: key(5), Endpoint: &ep5},
	}}
}

func pfxs(ss ...string) []netip.Prefix {
	var out []netip.Prefix
	for _, s := range ss {
		out = append(out, netip.MustParsePrefix(s))
	}
	return out
}

func TestUAPISetGolden(t *testing.T) {
	got := []byte(wg.UAPISet(uapiChange()))
	path := filepath.Join("testdata", "uapi-set.golden.txt")
	if *update {
		if err := os.WriteFile(path, got, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("UAPI set differs from %s; review the change, then ./scripts/dev.sh regen\n%s", path, got)
	}
	if wg.UAPISet(wg.Change{}) != "" {
		t.Fatal("an empty change must render nothing: a bare newline would end the operation")
	}
}

// getAnswer is what wireguard-go answers to "get" for the device the
// golden's change produces, peers in its map order.
func getAnswer() string {
	return strings.Join([]string{
		"private_key=" + hexOf(key(1)),
		"listen_port=51820",
		"fwmark=51820",
		"public_key=" + hexOf(key(4)),
		"preshared_key=" + strings.Repeat("0", 64),
		"protocol_version=1",
		"last_handshake_time_sec=0",
		"last_handshake_time_nsec=0",
		"tx_bytes=0",
		"rx_bytes=0",
		"persistent_keepalive_interval=25",
		"allowed_ip=10.77.1.1/32",
		"public_key=" + hexOf(key(2)),
		"preshared_key=" + strings.Repeat("0", 64),
		"protocol_version=1",
		"endpoint=203.0.113.12:51820",
		"last_handshake_time_sec=1760000000",
		"last_handshake_time_nsec=5",
		"tx_bytes=100",
		"rx_bytes=200",
		"persistent_keepalive_interval=0",
		"allowed_ip=10.77.2.0/24",
		"allowed_ip=10.77.0.2/32",
	}, "\n") + "\n"
}

func TestParseUAPI(t *testing.T) {
	d, stats, err := wg.ParseUAPI(getAnswer())
	if err != nil {
		t.Fatal(err)
	}
	if d.PrivateKey != key(1) || d.ListenPort != 51820 || d.FwMark != 51820 || len(d.Peers) != 2 || len(stats) != 2 {
		t.Fatalf("device = %+v, stats %+v", d, stats)
	}
	p2, p4 := d.Peers[0], d.Peers[1]
	if p2.PublicKey != key(2) || p4.PublicKey != key(4) {
		t.Fatalf("peers are not sorted by key: %v, %v", p2.PublicKey, p4.PublicKey)
	}
	if p2.Endpoint != netip.MustParseAddrPort("203.0.113.12:51820") || p2.Keepalive != 0 ||
		!slices.Equal(p2.AllowedIPs, pfxs("10.77.0.2/32", "10.77.2.0/24")) {
		t.Fatalf("server peer = %+v", p2)
	}
	if p4.Endpoint.IsValid() || p4.Keepalive != 25*time.Second || !slices.Equal(p4.AllowedIPs, pfxs("10.77.1.1/32")) {
		t.Fatalf("dial-in peer = %+v", p4)
	}
	s2, s4 := stats[0], stats[1]
	if s2.PublicKey != key(2) || !s2.LastHandshake.Equal(time.Unix(1760000000, 5)) || s2.RxBytes != 200 || s2.TxBytes != 100 ||
		s2.Endpoint != p2.Endpoint {
		t.Fatalf("server peer stats = %+v", s2)
	}
	if !s4.LastHandshake.IsZero() {
		t.Fatalf("a peer that never shook hands has handshake %v", s4.LastHandshake)
	}

	// What the device reports is what Diff compares: the same desired
	// device again is no change, so a re-apply on userspace is a no-op.
	desired := wg.Device{PrivateKey: key(1), ListenPort: 51820, FwMark: 51820, Peers: []wg.Peer{
		{PublicKey: key(2), Endpoint: p2.Endpoint, AllowedIPs: pfxs("10.77.0.2/32", "10.77.2.0/24")},
		{PublicKey: key(4), Keepalive: 25 * time.Second, AllowedIPs: pfxs("10.77.1.1/32")},
	}}
	if c := wg.Diff(d, desired); c.Count() != 0 {
		t.Fatalf("Diff after a faithful read = %+v", c)
	}
	if empty, none, err := wg.ParseUAPI(""); err != nil || len(empty.Peers) != 0 || none != nil {
		t.Fatalf("an empty device = %+v, %v, %v", empty, none, err)
	}
}

func TestParseUAPIRejectsMalformedValues(t *testing.T) {
	secret := strings.Repeat("ab", 31) + "zz"
	peer := "public_key=" + hexOf(key(2)) + "\n"
	cases := []struct{ in, want string }{
		{"private_key=" + secret, "line 1, private_key"},
		{"listen_port=70000", "line 1, listen_port"},
		{"fwmark=-1", "line 1, fwmark"},
		{"no equals sign", "line 1 has no '='"},
		{"public_key=abc", "line 1, public_key"},
		{peer + "endpoint=nowhere", "line 2, endpoint"},
		{peer + "allowed_ip=10.77.0.2", "line 2, allowed_ip"},
		{peer + "rx_bytes=many", "line 2, rx_bytes"},
		{peer + "persistent_keepalive_interval=99999", "line 2, persistent_keepalive_interval"},
	}
	for _, c := range cases {
		_, _, err := wg.ParseUAPI(c.in)
		if err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("ParseUAPI(%q) = %v, want %q", c.in, err, c.want)
			continue
		}
		if strings.Contains(err.Error(), secret) {
			t.Errorf("the error quotes a key: %v", err)
		}
	}
}

func TestOpenRefusesUnknownBackends(t *testing.T) {
	for _, b := range []string{"", "Kernel", "wireguard-go", "kernel;reboot"} {
		h, got, err := wg.Open(b, slog.New(slog.DiscardHandler))
		if err == nil || h != nil || got != "" || !strings.Contains(err.Error(), "--wg-backend") {
			t.Errorf("Open(%q) = %v, %q, %v", b, h, got, err)
		}
	}
}
