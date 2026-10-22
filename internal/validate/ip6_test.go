package validate

import (
	"strings"
	"testing"
)

func TestIPv6Grammar(t *testing.T) {
	for _, c := range []struct {
		name string
		fn   func(string) (string, error)
		ok   map[string]string // input -> canonical form
		bad  []string
	}{
		{"IPv6", func(s string) (string, error) { a, err := IPv6("f", s); return a.String(), err },
			map[string]string{"fd00:10:77::1": "fd00:10:77::1", "FD00:0010:0077::0001": "fd00:10:77::1", "2001:db8::10": "2001:db8::10"},
			[]string{"", "10.77.0.1", "::ffff:10.77.0.1", "fe80::1%eth0", "fd00::1/64", "fd00::1\n", "fd00::g"}},
		{"Prefix6", func(s string) (string, error) { p, err := Prefix6("f", s); return p.String(), err },
			map[string]string{"fd00:10:77::/64": "fd00:10:77::/64", "fd00:10:77:0::/48": "fd00:10:77::/48"},
			[]string{"", "10.77.0.0/16", "fd00:10:77::1/64", "fd00::/129", "::ffff:10.0.0.0/104"}},
		{"IP", func(s string) (string, error) { a, err := IP("f", s); return a.String(), err },
			map[string]string{"10.77.0.1": "10.77.0.1", "fd00:10:77::1": "fd00:10:77::1"},
			[]string{"", "::ffff:10.77.0.1", "10.77.0.1/32", "host.example"}},
		{"Prefix", func(s string) (string, error) { p, err := Prefix("f", s); return p.String(), err },
			map[string]string{"10.77.0.0/16": "10.77.0.0/16", "fd00:10:77::/64": "fd00:10:77::/64", "::/0": "::/0"},
			[]string{"", "10.77.0.5/16", "fd00::1/64", "10.77.0.1"}},
		{"PeerEndpoint", func(s string) (string, error) { ap, err := PeerEndpoint("f", s); return ap.String(), err },
			map[string]string{"203.0.113.10:51820": "203.0.113.10:51820", "[2001:db8::10]:51820": "[2001:db8::10]:51820"},
			[]string{"", "2001:db8::10:51820", "[2001:db8::10]:0", "[::ffff:1.2.3.4]:1", "[fe80::1%eth0]:51820", "vps.example:51820"}},
	} {
		for in, want := range c.ok {
			if got, err := c.fn(in); err != nil || got != want {
				t.Errorf("%s(%q) = %q, %v; want %q", c.name, in, got, err, want)
			}
		}
		for _, in := range c.bad {
			if _, err := c.fn(in); err == nil {
				t.Errorf("%s(%q) accepted", c.name, in)
			}
		}
	}
	if _, err := Endpoint6("endpoint6", "203.0.113.10:51820"); err == nil || !strings.Contains(err.Error(), "[IPv6]:port") {
		t.Errorf("Endpoint6 of an IPv4 endpoint = %v", err)
	}
	if _, err := Endpoint("endpoint", "[2001:db8::10]:51820"); err == nil {
		t.Error("Endpoint, the IPv4 one, accepted an IPv6 endpoint")
	}
}
