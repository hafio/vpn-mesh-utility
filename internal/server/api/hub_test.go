package api

import (
	"testing"

	"github.com/hafio/vpn-mesh-utility/internal/proto"
)

func closed(ch <-chan struct{}) bool {
	select {
	case <-ch:
		return true
	default:
		return false
	}
}

func TestHubLatestWins(t *testing.T) {
	h := NewHub()
	events, _, _, _, cancel := h.Subscribe(1, 1)
	defer cancel()
	h.Notify(map[int64]proto.ConfigEvent{1: {Version: 1}, 2: {Version: 1}})
	h.Notify(map[int64]proto.ConfigEvent{1: {Version: 2}})
	if ev := <-events; ev.Version != 2 {
		t.Fatalf("got v%d, want only the newest", ev.Version)
	}
	select {
	case ev := <-events:
		t.Fatalf("stale event v%d left behind", ev.Version)
	default:
	}
}

func TestHubReplacesKicksAndCloses(t *testing.T) {
	h := NewHub()
	_, _, _, done1, cancel1 := h.Subscribe(1, 1)
	_, _, _, done2, cancel2 := h.Subscribe(1, 1)
	if !closed(done1) || closed(done2) || h.Subscribers() != 1 {
		t.Fatal("a new subscription must end the old one")
	}
	cancel1() // the old stream's cleanup must not remove the new one
	if h.Subscribers() != 1 {
		t.Fatal("cancel of a replaced stream removed its successor")
	}
	h.Kick(1)
	if !closed(done2) || h.Subscribers() != 0 {
		t.Fatal("Kick must end the stream")
	}
	cancel2()
	_, _, _, done3, cancel3 := h.Subscribe(3, 1)
	defer cancel3()
	h.Close()
	_, _, _, done4, cancel4 := h.Subscribe(4, 1)
	defer cancel4()
	if !closed(done3) || !closed(done4) || h.Subscribers() != 0 {
		t.Fatal("Close must end every stream and refuse new ones")
	}
}

// A mesh nudge reaches every stream of that mesh and no other, and nudges
// that pile up collapse into one.
func TestHubNudgesEveryStream(t *testing.T) {
	h := NewHub()
	_, mesh1, _, _, cancel1 := h.Subscribe(1, 1)
	defer cancel1()
	_, mesh2, _, _, cancel2 := h.Subscribe(2, 1)
	defer cancel2()
	_, other, _, _, cancel3 := h.Subscribe(3, 2)
	defer cancel3()
	h.NotifyMesh(1)
	h.NotifyMesh(1)
	if closed(other) {
		t.Fatal("another mesh's stream got the nudge")
	}
	for i, ch := range []<-chan struct{}{mesh1, mesh2} {
		if !closed(ch) {
			t.Fatalf("stream %d got no nudge", i+1)
		}
		if closed(ch) {
			t.Fatalf("stream %d got two nudges; they must collapse", i+1)
		}
	}
}

// A challenge reaches the node's own stream, latest wins, and a node
// without one is skipped (amendment A50).
func TestHubDeliversChallenges(t *testing.T) {
	h := NewHub()
	_, _, challenge, _, cancel := h.Subscribe(1, 1)
	defer cancel()
	h.Challenge(1, "first")
	h.Challenge(1, "second")
	h.Challenge(2, "nobody")
	if got := <-challenge; got != "second" {
		t.Fatalf("challenge %q", got)
	}
	select {
	case got := <-challenge:
		t.Fatalf("a second challenge %q", got)
	default:
	}
}
