package api

import (
	"errors"
	"strconv"
	"testing"
	"time"

	"github.com/hafio/vpn-mesh-utility/internal/testfix"
)

func TestReplayCache(t *testing.T) {
	c := newReplayCache()
	now := testfix.Epoch
	if c.Seen("a", now) {
		t.Fatal("empty cache saw a signature")
	}
	if err := c.Add("a", now); err != nil {
		t.Fatal(err)
	}
	if !c.Seen("a", now.Add(replayTTL-time.Second)) || !errors.Is(c.Add("a", now.Add(time.Minute)), errReplay) {
		t.Fatal("a signature inside the window must be a replay")
	}
	later := now.Add(replayTTL + sweepEvery)
	if c.Seen("a", later) || c.Add("a", later) != nil {
		t.Fatal("an expired signature must be forgotten")
	}
}

func TestReplayCacheRefusesWhenFullOfLiveEntries(t *testing.T) {
	c := newReplayCache()
	now := testfix.Epoch
	for i := 0; i < replayMax; i++ {
		c.seen[strconv.Itoa(i)] = now.Add(replayTTL)
	}
	if err := c.Add("new", now); !errors.Is(err, errReplayFull) {
		t.Fatalf("Add on a full cache = %v", err)
	}
	if err := c.Add("new", now.Add(replayTTL)); err != nil {
		t.Fatalf("after expiry the sweep must make room: %v", err)
	}
}
