package auth

import (
	"net/http"
	"net/netip"
	"strings"
	"sync"
	"time"
)

// ClientIP is the address that sent r: the TCP peer, or -- when the peer is a
// trusted proxy -- the right-most address in X-Forwarded-For that is not
// itself a trusted proxy. Left-most entries are client-supplied and never
// believed. The zero Addr means the peer address could not be parsed.
func ClientIP(r *http.Request, trusted []netip.Prefix) netip.Addr {
	ap, err := netip.ParseAddrPort(r.RemoteAddr)
	if err != nil {
		return netip.Addr{}
	}
	ip := ap.Addr().Unmap()
	if !inAny(ip, trusted) {
		return ip
	}
	var hops []string
	for _, h := range r.Header.Values("X-Forwarded-For") {
		hops = append(hops, strings.Split(h, ",")...)
	}
	for i := len(hops) - 1; i >= 0; i-- {
		a, err := netip.ParseAddr(strings.TrimSpace(hops[i]))
		if err != nil {
			return ip // a malformed hop: trust nothing beyond the last proxy
		}
		a = a.Unmap()
		if !inAny(a, trusted) {
			return a
		}
		ip = a
	}
	return ip
}

func inAny(a netip.Addr, ps []netip.Prefix) bool {
	for _, p := range ps {
		if p.Contains(a) {
			return true
		}
	}
	return false
}

// RateLimiter is a fixed-window counter per key (a client IP).
type RateLimiter struct {
	mu     sync.Mutex
	limit  int
	window time.Duration
	hits   map[string]*hitWindow
}

type hitWindow struct {
	start time.Time
	n     int
}

// maxLimiterKeys bounds the limiter's memory; on overflow expired windows are
// dropped, and if every window is live the newest key is refused.
const maxLimiterKeys = 4096

// NewRateLimiter allows limit attempts per key per window.
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{limit: limit, window: window, hits: map[string]*hitWindow{}}
}

// Allow counts one attempt for key and reports whether it is within limit.
func (l *RateLimiter) Allow(key string, now time.Time) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	w, ok := l.hits[key]
	if ok && now.Sub(w.start) < l.window {
		if w.n >= l.limit {
			return false
		}
		w.n++
		return true
	}
	if !ok && len(l.hits) >= maxLimiterKeys {
		for k, old := range l.hits {
			if now.Sub(old.start) >= l.window {
				delete(l.hits, k)
			}
		}
		if len(l.hits) >= maxLimiterKeys {
			return false
		}
	}
	l.hits[key] = &hitWindow{start: now, n: 1}
	return true
}
