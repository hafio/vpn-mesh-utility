package validate

import (
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
	} {
		r, err := AddrEntry("e", in)
		if err != nil || r.String() != want {
			t.Errorf("AddrEntry(%q) = %v, %v; want %s", in, r, err, want)
		}
	}
	for in, reason := range map[string]string{
		"":                        "IPv4 address, network or range",
		"10.77.1.5/24":            "host bits",
		"10.77.1.20-10.77.1.5":    "start at or below",
		"10.77.1.5-":              "two IPv4 addresses",
		"::1":                     "IPv4 address, network or range",
		"10.77.1.0/33":            "IPv4 CIDR",
		"10.77.1.5, 10.77.1.6":    "IPv4 address, network or range",
		"node:nas":                "IPv4 address, network or range",
		"10.77.1.5-10.77.1.6/24":  "two IPv4 addresses",
		"10.77.1.5\n10.77.1.6":    "IPv4 address, network or range",
		"10.77.1.1-10.77.1.2-1.1": "two IPv4 addresses",
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
	if U32(FromU32(0x0a4d0105)) != 0x0a4d0105 || FromU32(0x0a4d0105).String() != "10.77.1.5" {
		t.Fatal("U32 and FromU32 do not round trip")
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
	if _, _, ok, err := Ref("r", "10.77.1.5", "group"); ok || err != nil {
		t.Fatalf("a literal is not a reference: %v %v", ok, err)
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
