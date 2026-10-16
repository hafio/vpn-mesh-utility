package netconv

import (
	"net"
	"net/netip"
	"testing"
)

func TestPrefixRoundTrip(t *testing.T) {
	for _, s := range []string{"10.77.0.1/32", "10.77.0.0/16", "0.0.0.0/0"} {
		p := netip.MustParsePrefix(s)
		got, ok := Prefix(IPNet(p))
		if !ok || got != p {
			t.Errorf("Prefix(IPNet(%s)) = %s, %v", p, got, ok)
		}
	}
}

func TestPrefixUnmapsIPv4(t *testing.T) {
	// net.ParseCIDR returns a 16-byte IP for IPv4 with a 4-byte mask.
	_, n, err := net.ParseCIDR("10.77.1.0/24")
	if err != nil {
		t.Fatal(err)
	}
	n.IP = n.IP.To16()
	if got, ok := Prefix(*n); !ok || got != netip.MustParsePrefix("10.77.1.0/24") {
		t.Errorf("16-byte IPv4 with 4-byte mask = %s, %v", got, ok)
	}
	mapped := net.IPNet{IP: net.ParseIP("10.77.1.0"), Mask: net.CIDRMask(120, 128)}
	if got, ok := Prefix(mapped); !ok || got != netip.MustParsePrefix("10.77.1.0/24") {
		t.Errorf("v4-mapped /120 = %s, %v", got, ok)
	}
	if _, ok := Prefix(net.IPNet{IP: net.IPv4(10, 0, 0, 1), Mask: net.IPMask{255, 0, 255, 0}}); ok {
		t.Error("a non-canonical mask must be refused")
	}
}

func TestAddrPort(t *testing.T) {
	if got := AddrPort(nil); got.IsValid() {
		t.Errorf("AddrPort(nil) = %s", got)
	}
	a := &net.UDPAddr{IP: net.ParseIP("203.0.113.10"), Port: 51820}
	if got := AddrPort(a); got != netip.MustParseAddrPort("203.0.113.10:51820") {
		t.Errorf("AddrPort = %s", got)
	}
}
