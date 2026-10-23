// Package phone makes the WireGuard profile a phone imports (spec S8.3,
// amendment A6). Phones have no agent, so the server generates the key pair
// when the profile is shown, keeps only the public key and hands the private
// key out once, inside the profile. The profile is wg-quick's format, which
// the WireGuard apps import from a file or a QR code.
//
// The private key never appears in an error: a caller may log one.
package phone

import (
	"crypto/ecdh"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/netip"
	"slices"
	"strings"

	"github.com/hafio/vpn-mesh-utility/internal/proto"
)

// Full tunnel: every IPv4 and IPv6 destination. IPv6 goes to the exit too,
// where nothing routes it, so it cannot leak past the tunnel (amendment
// A11's blackhole, the app's way).
var fullTunnel = []string{"0.0.0.0/0", "::/0"}

// NewKey returns a WireGuard key pair, base64 as wg(8) prints it, made from
// 32 bytes of r (crypto/rand.Reader outside tests).
func NewKey(r io.Reader) (priv, pub string, err error) {
	var k [32]byte
	if _, err := io.ReadFull(r, k[:]); err != nil {
		return "", "", fmt.Errorf("read randomness for a phone key: %w", err)
	}
	// Clamped as wg(8) clamps; X25519 would clamp anyway, so the public
	// key is the same either way.
	k[0] &= 248
	k[31] = k[31]&127 | 64
	sk, err := ecdh.X25519().NewPrivateKey(k[:])
	if err != nil {
		return "", "", fmt.Errorf("make a phone key: %w", err)
	}
	enc := base64.StdEncoding
	return enc.EncodeToString(k[:]), enc.EncodeToString(sk.PublicKey().Bytes()), nil
}

// Profile is what Render puts in a phone's profile.
type Profile struct {
	// Config is the phone's computed config: its hub's peer carries the
	// whole mesh.
	Config proto.NodeConfig
	// PrivateKey is NewKey's private half.
	PrivateKey string
	// DNS is the hub's mesh IP: its resolver answers the mesh zone and
	// forwards everything else for its phones (amendment A12) -- its IPv6
	// one for an IPv6-only phone. DNS6 is the hub's IPv6 mesh IP beside it,
	// for a phone with both families (M8); invalid for none.
	DNS, DNS6 netip.Addr
	// Exit is the server the phone's internet goes through, one of the
	// config's exit choices; empty is the mesh-only profile.
	Exit string
}

// Render returns the profile in wg-quick's format.
func Render(p Profile) (string, error) {
	c := p.Config
	if c.Node.Role != proto.RolePhone {
		return "", fmt.Errorf("%s is a %s, not a phone", c.Node.Name, c.Node.Role)
	}
	if err := c.Validate(); err != nil {
		return "", err
	}
	if k, err := base64.StdEncoding.DecodeString(p.PrivateKey); err != nil || len(k) != 32 {
		return "", errors.New("the private key is not a WireGuard key (base64 of 32 bytes)")
	}
	switch {
	case !p.DNS.IsValid():
		return "", errors.New("the profile has no DNS server")
	case p.DNS.Is4() && c.Interface.Address == "", p.DNS.Is6() && c.Interface.Address6 == "":
		return "", fmt.Errorf("the DNS server %s is of a family the phone has no address of", p.DNS)
	case p.DNS6.IsValid() && (!p.DNS6.Is6() || c.Interface.Address6 == ""):
		return "", fmt.Errorf("the second DNS server %s must be an IPv6 address of a phone with IPv6", p.DNS6)
	}
	if p.Exit != "" && (c.Exit == nil || !slices.Contains(c.Exit.Choices, p.Exit)) {
		var choices []string
		if c.Exit != nil {
			choices = c.Exit.Choices
		}
		return "", fmt.Errorf("%s is not an exit this phone can use (choices: %s)", p.Exit, strings.Join(choices, ", "))
	}

	var b strings.Builder
	line := func(key string, value any) { fmt.Fprintf(&b, "%s = %v\n", key, value) }
	b.WriteString("[Interface]\n")
	line("PrivateKey", p.PrivateKey)
	line("Address", joinSet(c.Interface.Address, c.Interface.Address6))
	dns := p.DNS.String()
	if p.DNS6.IsValid() {
		dns += ", " + p.DNS6.String()
	}
	line("DNS", dns)
	line("MTU", c.Interface.MTU)
	for _, peer := range c.Peers {
		allowed := peer.AllowedIPs
		if peer.Name == p.Exit {
			allowed = append(slices.Clone(allowed), fullTunnel...)
		}
		b.WriteString("\n[Peer]\n")
		line("PublicKey", peer.PublicKey)
		if peer.Endpoint != "" {
			line("Endpoint", peer.Endpoint)
		}
		line("AllowedIPs", strings.Join(allowed, ", "))
		if peer.Keepalive > 0 {
			line("PersistentKeepalive", peer.Keepalive)
		}
	}
	return b.String(), nil
}

// joinSet is the non-empty values, comma-separated: wg-quick's list.
func joinSet(values ...string) string {
	return strings.Join(slices.DeleteFunc(values, func(v string) bool { return v == "" }), ", ")
}

// Hub is the peer that carries the whole mesh: the phone's hub, whose
// resolver the profile uses. Its allowed IPs hold the mesh CIDR (a /22 or
// wider) or its IPv6 prefix (a /64 or wider), which no other peer's do: a
// server's container network is a /24 or a /112.
func Hub(c proto.NodeConfig) (string, bool) {
	for _, peer := range c.Peers {
		for _, a := range peer.AllowedIPs {
			pf, err := netip.ParsePrefix(a)
			if err != nil || pf.Bits() == 0 {
				continue
			}
			if (pf.Addr().Is4() && pf.Bits() <= 22) || (pf.Addr().Is6() && pf.Bits() <= 64) {
				return peer.Name, true
			}
		}
	}
	return "", false
}
