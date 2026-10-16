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
	events, _, cancel := h.Subscribe(1)
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
	_, done1, cancel1 := h.Subscribe(1)
	_, done2, cancel2 := h.Subscribe(1)
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
	_, done3, cancel3 := h.Subscribe(3)
	defer cancel3()
	h.Close()
	_, done4, cancel4 := h.Subscribe(4)
	defer cancel4()
	if !closed(done3) || !closed(done4) || h.Subscribers() != 0 {
		t.Fatal("Close must end every stream and refuse new ones")
	}
}
