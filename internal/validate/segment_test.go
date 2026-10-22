package validate

import (
	"net/netip"
	"strings"
	"testing"
)

func TestAddrEntry(t *testing.T) {
	for in, want := range map[string]string{
		"10.77.1.5":               "10.77.1.5",
		"10.77.1.0/24":            "10.77.1.0/24",
		"10.77.1.5/32":            "10.77.1.5",
		"0.0.0.0/0":               "0.0.0.0/0",
		"10.77.1.5-10.77.1.20":    "10.77.1.5-10.77.1.20",
		"10.77.1.0-10.77.1.255":   "10.77.1.0/24", // a range that is a network reads as one
		"10.77.1.7-10.77.1.7":     "10.77.1.7",
		"0.0.0.0-255.255.255.255": "0.0.0.0/0",
		// IPv6 (M8), in canonical form.
		"FD10:77::1:5":                       "fd10:77::1:5",
		"fd10:77::/64":                       "fd10:77::/64",
		"fd10:77::1:5/128":                   "fd10:77::1:5",
		"fd10:77::2:10-fd10:77::2:20":        "fd10:77::2:10-fd10:77::2:20",
		"fd10:77::8000:0-fd10:77::8000:ffff": "fd10:77::8000:0/112",
	} {
		r, err := AddrEntry("e", in)
		if err != nil || r.String() != want {
			t.Errorf("AddrEntry(%q) = %v, %v; want %s", in, r, err, want)
		}
	}
	for in, reason := range map[string]string{
		"":                        "address, network or range",
		"10.77.1.5/24":            "host bits",
		"10.77.1.20-10.77.1.5":    "start at or below",
		"10.77.1.5-":              "two addresses of one family",
		"10.77.1.0/33":            "IPv4 or IPv6 CIDR",
		"10.77.1.5, 10.77.1.6":    "address, network or range",
		"node:nas":                "address, network or range",
		"10.77.1.5-10.77.1.6/24":  "two addresses of one family",
		"10.77.1.5\n10.77.1.6":    "address, network or range",
		"10.77.1.1-10.77.1.2-1.1": "two addresses of one family",
		"10.77.1.5-fd10:77::1":    "two addresses of one family",
		"fd10:77::5-fd10:77::1":   "start at or below",
		"fd10:77::1/64":           "host bits",
		"fe80::1%eth0":            "address, network or range",
		"::ffff:10.77.1.5":        "address, network or range",
	} {
		if _, err := AddrEntry("e", in); err == nil || !strings.Contains(err.Error(), reason) {
			t.Errorf("AddrEntry(%q) = %v, want an error about %q", in, err, reason)
		}
	}
}

func TestAddrRangePrefix(t *testing.T) {
	r, _ := AddrEntry("e", "10.77.0.0-10.77.0.127")
	if p, ok := r.Prefix(); !ok || p.String() != "10.77.0.0/25" {
		t.Fatalf("Prefix = %v, %v", p, ok)
	}
	for _, in := range []string{"10.77.0.1-10.77.0.128", "10.77.0.0-10.77.0.2"} { // unaligned, not a power of two
		r, _ := AddrEntry("e", in)
		if p, ok := r.Prefix(); ok {
			t.Errorf("%s is no network, got %v", in, p)
		}
	}
	if r, _ := AddrEntry("e", "10.77.0.2-10.77.0.3"); r.String() != "10.77.0.2/31" {
		t.Errorf("an aligned pair reads %s", r)
	}
	if r, _ := AddrEntry("e", "::-ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff"); r.String() != "::/0" {
		t.Errorf("the whole of IPv6 reads %s", r)
	}
	if r, _ := AddrEntry("e", "fd10:77::2:0-fd10:77::2:ff"); r.String() != "fd10:77::2:0/120" {
		t.Errorf("an IPv6 range that is a network reads %s", r)
	}
	if r := (AddrRange{First: netip.MustParseAddr("10.77.0.1"), Last: netip.MustParseAddr("fd10:77::1")}); r.String() != "10.77.0.1-fd10:77::1" {
		t.Errorf("a range across families is no network: %s", r)
	}
}

func TestMergeRanges(t *testing.T) {
	var in []AddrRange
	for _, s := range []string{"fd10:77::5", "10.77.0.4-10.77.0.9", "10.77.0.1-10.77.0.3", "fd10:77::1-fd10:77::4", "10.77.0.20",
		"255.255.255.255", "::", "10.77.0.2"} {
		r, err := AddrEntry("e", s)
		if err != nil {
			t.Fatal(err)
		}
		in = append(in, r)
	}
	var got []string
	for _, r := range MergeRanges(in) {
		got = append(got, r.String())
	}
	// Touching ranges join; IPv4's last address and IPv6's first never do.
	if want := "10.77.0.1-10.77.0.9 10.77.0.20 255.255.255.255 :: fd10:77::1-fd10:77::5"; strings.Join(got, " ") != want {
		t.Fatalf("MergeRanges = %v, want %s", got, want)
	}
	if MergeRanges(nil) != nil {
		t.Fatal("nothing merged is nil")
	}
	merged := MergeRanges(in)
	for addr, want := range map[string]bool{"10.77.0.5": true, "10.77.0.10": false, "fd10:77::3": true, "fd10:77::6": false, "::": true} {
		if RangesContain(merged, netip.MustParseAddr(addr)) != want {
			t.Errorf("RangesContain(%s) != %v", addr, want)
		}
	}
}

func TestServiceEntry(t *testing.T) {
	for in, want := range map[string]string{
		"tcp":           "tcp",
		"udp":           "udp",
		"tcp/22":        "tcp/22",
		"udp/53-53":     "udp/53",
		"tcp/8000-8100": "tcp/8000-8100",
		"icmp":          "icmp",
		"tcp/65535":     "tcp/65535",
	} {
		s, err := ServiceEntry("s", in)
		if err != nil || s.String() != want {
			t.Errorf("ServiceEntry(%q) = %v, %v; want %s", in, s, err, want)
		}
	}
	if s, _ := ServiceEntry("s", "tcp/8000-8100"); s.Ports() != "8000-8100" {
		t.Errorf("Ports = %q", s.Ports())
	}
	if s, _ := ServiceEntry("s", "icmp"); s.Ports() != "" {
		t.Errorf("icmp Ports = %q", s.Ports())
	}
	for in, reason := range map[string]string{
		"":            "must be tcp, udp",
		"sctp/1":      "must be tcp, udp",
		"icmp/8":      "no ports",
		"tcp/":        "between 1 and 65535",
		"tcp/0":       "between 1 and 65535",
		"tcp/65536":   "between 1 and 65535",
		"tcp/022":     "between 1 and 65535",
		"tcp/+22":     "between 1 and 65535",
		"tcp/30-20":   "start at or below",
		"tcp/22,80":   "between 1 and 65535",
		"TCP/22":      "must be tcp, udp",
		"tcp/22 ":     "between 1 and 65535",
		"tcp/1-2-3":   "between 1 and 65535",
		"tcp/22;ls":   "between 1 and 65535",
		"svc:web":     "must be tcp, udp",
		"udp/1-65536": "between 1 and 65535",
	} {
		if _, err := ServiceEntry("s", in); err == nil || !strings.Contains(err.Error(), reason) {
			t.Errorf("ServiceEntry(%q) = %v, want an error about %q", in, err, reason)
		}
	}
}

func TestRef(t *testing.T) {
	kind, name, ok, err := Ref("r", "group:laptops", "group", "node")
	if kind != "group" || name != "laptops" || !ok || err != nil {
		t.Fatalf("Ref = %q %q %v %v", kind, name, ok, err)
	}
	for _, literal := range []string{"10.77.1.5", "fd10:77::1", "fd10:77::/64", "::1-::5", "fd10::zz"} {
		if _, _, ok, err := Ref("r", literal, "group"); ok || err != nil {
			t.Fatalf("Ref(%q): a literal is not a reference: %v %v", literal, ok, err)
		}
	}
	for in, reason := range map[string]string{
		"zone:office":   "kind one of group, node", // a kind this field does not take
		"group:":        "the name after group:",
		"group:Laptops": "the name after group:",
		"group:a b":     "the name after group:",
		"group:a:b":     "the name after group:",
		":laptops":      "kind one of",
	} {
		if _, _, ok, err := Ref("r", in, "group", "node"); !ok || err == nil || !strings.Contains(err.Error(), reason) {
			t.Errorf("Ref(%q) = %v %v, want a refusal about %q", in, ok, err, reason)
		}
	}
}
