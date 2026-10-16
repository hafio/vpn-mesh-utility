package auth

import (
	"net/http/httptest"
	"net/netip"
	"testing"
	"time"

	"github.com/hafio/vpn-mesh-utility/internal/testfix"
)

func TestRateLimiter(t *testing.T) {
	l := NewRateLimiter(2, time.Minute)
	now := testfix.Epoch
	first, second, third := l.Allow("a", now), l.Allow("a", now), l.Allow("a", now)
	if !first || !second || third {
		t.Fatal("limit 2 must allow two and refuse the third")
	}
	if !l.Allow("b", now) {
		t.Fatal("keys are independent")
	}
	if !l.Allow("a", now.Add(time.Minute)) {
		t.Fatal("a new window starts over")
	}
}

func TestClientIP(t *testing.T) {
	trusted := []netip.Prefix{netip.MustParsePrefix("10.0.0.0/8")}
	cases := []struct {
		name   string
		remote string
		xff    []string
		want   string
	}{
		{"direct", "203.0.113.5:4000", nil, "203.0.113.5"},
		{"untrusted peer's XFF is ignored", "203.0.113.5:4000", []string{"198.51.100.1"}, "203.0.113.5"},
		{"trusted proxy", "10.0.1.2:4000", []string{"198.51.100.1"}, "198.51.100.1"},
		{"spoofed left-most entry", "10.0.1.2:4000", []string{"6.6.6.6, 198.51.100.1"}, "198.51.100.1"},
		{"proxy chain", "10.0.1.2:4000", []string{"198.51.100.1", "10.0.9.9"}, "198.51.100.1"},
		{"malformed hop", "10.0.1.2:4000", []string{"198.51.100.1, junk"}, "10.0.1.2"},
		{"only proxies", "10.0.1.2:4000", []string{"10.0.9.9"}, "10.0.9.9"},
	}
	for _, c := range cases {
		r := httptest.NewRequest("GET", "/", nil)
		r.RemoteAddr = c.remote
		for _, v := range c.xff {
			r.Header.Add("X-Forwarded-For", v)
		}
		if got := ClientIP(r, trusted).String(); got != c.want {
			t.Errorf("%s: %s, want %s", c.name, got, c.want)
		}
	}
	r := httptest.NewRequest("GET", "/", nil)
	r.RemoteAddr = "@"
	if ClientIP(r, nil).IsValid() {
		t.Error("an unparseable peer must give the zero Addr")
	}
}
