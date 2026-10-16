package phone

import (
	"bytes"
	"encoding/hex"
	"flag"
	"net/netip"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hafio/vpn-mesh-utility/internal/proto"
	"github.com/hafio/vpn-mesh-utility/internal/testfix"
)

var update = flag.Bool("update", false, "rewrite testdata goldens")

// alice is RFC 7748's X25519 test key pair (section 6.1).
const (
	alicePriv = "77076d0a7318a57d3c16c17251b26645df4c2f87ebc0992ab177fba51db92c2a"
	alicePub  = "hSDwCYkwp1R0i33ctD73Wg2/Og0mOBr066SpjqqbTmo="
)

func aliceKey(t *testing.T) string {
	t.Helper()
	seed, err := hex.DecodeString(alicePriv)
	if err != nil {
		t.Fatal(err)
	}
	priv, pub, err := NewKey(bytes.NewReader(seed))
	if err != nil || pub != alicePub {
		t.Fatalf("NewKey = %q, %v; want public key %q", pub, err, alicePub)
	}
	return priv
}

func TestNewKey(t *testing.T) {
	priv := aliceKey(t)
	// Stored clamped, as wg(8) prints a private key.
	if priv != "cAdtCnMYpX08FsFyUbJmRd9ML4frwJkqsXf7pR25LGo=" {
		t.Fatalf("private key = %q", priv)
	}
	if _, _, err := NewKey(bytes.NewReader(make([]byte, 31))); err == nil {
		t.Fatal("NewKey made a key from 31 bytes")
	}
}

func golden(t *testing.T, name, got string) {
	t.Helper()
	path := filepath.Join("testdata", name)
	if *update {
		if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read golden: %v (run ./scripts/dev.sh regen to create it)", err)
	}
	if got != string(want) {
		t.Fatalf("profile differs from %s:\ngot:\n%s\nwant:\n%s", path, got, want)
	}
}

func TestProfiles(t *testing.T) {
	priv := aliceKey(t)
	for _, c := range []struct{ golden, exit string }{
		{"mesh-only.golden.conf", ""},
		{"via-exit.golden.conf", testfix.VPS2.Name},
	} {
		t.Run(c.golden, func(t *testing.T) {
			cfg := testfix.PhoneConfig()
			got, err := Render(Profile{Config: cfg, PrivateKey: priv, DNS: netip.MustParseAddr(testfix.VPS1.MeshIP), Exit: c.exit})
			if err != nil {
				t.Fatal(err)
			}
			golden(t, c.golden, got)
			// The full tunnel is the profile's, never the computed config's.
			for _, p := range cfg.Peers {
				if len(p.AllowedIPs) != 1 {
					t.Fatalf("Render changed the config's peer %s: %v", p.Name, p.AllowedIPs)
				}
			}
		})
	}
}

// A profile is never made from something that is not a valid phone config,
// and no error carries the private key.
func TestRenderRefuses(t *testing.T) {
	priv := aliceKey(t)
	hub := netip.MustParseAddr(testfix.VPS1.MeshIP)
	cases := map[string]func(*Profile){
		"a spoke's config":  func(p *Profile) { p.Config = testfix.SpokeConfig("") },
		"injected endpoint": func(p *Profile) { p.Config.Peers[1].Endpoint = "203.0.113.11:51820\nPostUp = id" },
		"short key":         func(p *Profile) { p.PrivateKey = priv[:20] },
		"key not base64":    func(p *Profile) { p.PrivateKey = strings.Repeat("!", 44) },
		"no dns":            func(p *Profile) { p.DNS = netip.Addr{} },
		"ipv6 dns":          func(p *Profile) { p.DNS = netip.MustParseAddr("fd00::1") },
		"exit not a choice": func(p *Profile) { p.Exit = testfix.VPS1.Name },
		"no exit choices":   func(p *Profile) { p.Config.Exit, p.Exit = nil, testfix.VPS2.Name },
	}
	for name, edit := range cases {
		t.Run(name, func(t *testing.T) {
			p := Profile{Config: testfix.PhoneConfig(), PrivateKey: priv, DNS: hub}
			edit(&p)
			got, err := Render(p)
			if err == nil {
				t.Fatalf("Render accepted it:\n%s", got)
			}
			if strings.Contains(err.Error(), priv) || (len(p.PrivateKey) > 8 && strings.Contains(err.Error(), p.PrivateKey)) {
				t.Fatalf("the error carries the private key: %v", err)
			}
		})
	}
}

func TestHub(t *testing.T) {
	if hub, ok := Hub(testfix.PhoneConfig()); !ok || hub != testfix.VPS1.Name {
		t.Fatalf("Hub = %q, %v", hub, ok)
	}
	if hub, ok := Hub(proto.NodeConfig{Peers: []proto.Peer{{Name: "vps-1", AllowedIPs: []string{"10.77.0.1/32", "0.0.0.0/0"}}}}); ok {
		t.Fatalf("Hub found %q in a config without a mesh route", hub)
	}
}
