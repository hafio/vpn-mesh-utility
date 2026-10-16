package api

import (
	"net/http"
	"sync"
	"time"

	"github.com/hafio/vpn-mesh-utility/internal/proto"
)

// Replay protection (amendment A8): every accepted signature is remembered
// for replayTTL, longer than the +-60 s window in which its timestamp would
// pass, so a captured request cannot be sent again.
const (
	replayTTL = 2 * proto.MaxSkew
	replayMax = 1 << 16
	// sweepEvery bounds how stale the cache may get before expired entries
	// are dropped.
	sweepEvery = 10 * time.Second
)

var (
	errReplay = errorf(http.StatusUnauthorized, proto.CodeReplay,
		"this signed request was already received; agents sign every request afresh, so it was replayed or retried by a proxy")
	errReplayFull = errorf(http.StatusServiceUnavailable, proto.CodeRateLimited,
		"too many signed requests in the last %s; retry shortly", replayTTL)
)

type replayCache struct {
	mu    sync.Mutex
	seen  map[string]time.Time // signature -> expiry
	swept time.Time
}

func newReplayCache() *replayCache { return &replayCache{seen: map[string]time.Time{}} }

// Seen reports whether sig was accepted within the window. The server checks
// it before any database read; Add makes the final, atomic decision.
func (c *replayCache) Seen(sig string, now time.Time) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	exp, ok := c.seen[sig]
	return ok && now.Before(exp)
}

// Add records sig, failing if it is already present. When the cache is full
// of live entries it refuses rather than evicts: evicting a live signature
// would make it replayable.
func (c *replayCache) Add(sig string, now time.Time) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if now.Sub(c.swept) >= sweepEvery || len(c.seen) >= replayMax {
		for k, exp := range c.seen {
			if !now.Before(exp) {
				delete(c.seen, k)
			}
		}
		c.swept = now
	}
	if exp, ok := c.seen[sig]; ok && now.Before(exp) {
		return errReplay
	}
	if len(c.seen) >= replayMax {
		return errReplayFull
	}
	c.seen[sig] = now.Add(replayTTL)
	return nil
}
