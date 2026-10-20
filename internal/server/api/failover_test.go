package api

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/hafio/vpn-mesh-utility/internal/server/store"
	"github.com/hafio/vpn-mesh-utility/internal/testfix"
)

// A hub is down when both its heartbeat and its handshakes are stale, and
// back once its heartbeat has been steady for 2.5 times DownAfter; silence
// before the server started does not count.
func TestJudgeHubs(t *testing.T) {
	now := testfix.Epoch
	f := &Failover{DownAfter: 2 * time.Minute, started: now.Add(-time.Hour)}
	ago := func(d time.Duration) int64 { return now.Add(-d).Unix() }
	for name, c := range map[string]struct {
		seen, handshake int64
		cur, want       hubState
	}{
		"up and heard":            {ago(time.Second), ago(time.Second), hubState{}, hubState{}},
		"silent, tunnels work":    {ago(3 * time.Minute), ago(time.Minute), hubState{}, hubState{}},
		"heard, no handshake":     {ago(time.Second), ago(time.Hour), hubState{}, hubState{}},
		"silent and no handshake": {ago(3 * time.Minute), ago(4 * time.Minute), hubState{}, hubState{down: now.Unix()}},
		"never heard of":          {0, 0, hubState{}, hubState{down: now.Unix()}},
		"down, silent again": {ago(3 * time.Minute), ago(time.Second), hubState{down: ago(time.Hour), healthy: ago(time.Minute)},
			hubState{down: ago(time.Hour)}},
		"down, heard again": {ago(time.Second), ago(time.Hour), hubState{down: ago(time.Hour)},
			hubState{down: ago(time.Hour), healthy: now.Unix()}},
		"down, not healthy long enough": {ago(time.Second), ago(time.Second), hubState{down: ago(time.Hour), healthy: ago(4 * time.Minute)},
			hubState{down: ago(time.Hour), healthy: ago(4 * time.Minute)}},
		"down, healthy long enough": {ago(time.Second), ago(time.Second), hubState{down: ago(time.Hour), healthy: ago(5 * time.Minute)},
			hubState{}},
	} {
		if got := f.judge(now, c.seen, c.handshake, c.cur); got != c.want {
			t.Errorf("%s: judge = %+v, want %+v", name, got, c.want)
		}
	}
	// Just started: a hub not heard of yet is given DownAfter to report.
	f.started = now.Add(-30 * time.Second)
	if got := f.judge(now, 0, 0, hubState{}); got != (hubState{}) {
		t.Errorf("right after the start = %+v", got)
	}
}

func (h *harness) heartbeat(t *testing.T, name string, peers ...store.PeerStat) {
	t.Helper()
	id := h.node(t, name).ID
	if err := h.st.Tx(context.Background(), func(tx *store.Tx) error {
		return tx.RecordStatus(id, store.StatusUpdate{AgentVersion: "v1.0.0", WGBackend: "kernel", Peers: peers})
	}); err != nil {
		t.Fatal(err)
	}
}

// relayOf is the peer that carries name's traffic to the laptop.
func (h *harness) relayOf(t *testing.T, name string) string {
	t.Helper()
	for _, p := range h.config(t, name).Peers {
		for _, a := range p.AllowedIPs {
			if a == testfix.Laptop.MeshIP+"/32" {
				return p.Name
			}
		}
	}
	return ""
}

// vps-1 goes quiet: the NAS <-> laptop pair moves to the laptop's hub in a
// new version audited relay.failover, the node list says vps-1 is down; it
// comes back and, steady long enough, the pair returns (relay.restore).
func TestFailoverMovesRelays(t *testing.T) {
	h := newHarness(t)
	nas, laptop := storeNode(testfix.NAS, true), storeNode(testfix.Laptop, true)
	h.seed(t, storeNode(testfix.VPS1, true), storeNode(testfix.VPS2, true), storeNode(testfix.VPS3, true), nas, laptop)
	vps1, vps2 := h.node(t, "vps-1"), h.node(t, "vps-2")
	for _, n := range []store.Node{h.node(t, "nas"), h.node(t, "laptop")} {
		n.PrimaryHubID = vps1.ID
		if n.Name == "laptop" {
			n.PrimaryHubID = vps2.ID
		}
		if err := h.st.Tx(context.Background(), func(tx *store.Tx) error { return tx.UpdateNode(n) }); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := h.rec.Apply(context.Background(), Change{Actor: "test"}, nil); err != nil {
		t.Fatal(err)
	}
	if got := h.relayOf(t, "nas"); got != "vps-1" {
		t.Fatalf("before: nas reaches the laptop through %q, want vps-1", got)
	}
	f := &Failover{DB: h.db, Reconciler: h.rec, Now: h.clock.Now, DownAfter: 2 * time.Minute, started: testfix.Epoch.Add(-time.Hour)}
	h.clock.Advance(3 * time.Minute)
	h.heartbeat(t, "vps-2")
	h.heartbeat(t, "vps-3")
	before := h.latestVersion(t)
	if err := f.Check(context.Background()); err != nil {
		t.Fatal(err)
	}
	if got := h.relayOf(t, "nas"); got != "vps-2" || h.latestVersion(t) == before || h.audits(t, "relay.failover") != 1 {
		t.Fatalf("vps-1 down: relay %q, version %d (was %d), %d failover audits", got, h.latestVersion(t), before, h.audits(t, "relay.failover"))
	}
	var list NodeList
	if code, _ := adminCall(t, h, "GET", AdminNodes, nil, &list); code != 200 || list.Nodes[0].DownSince == 0 ||
		!strings.Contains(strings.Join(list.Warnings, " "), `node "vps-1": is down`) {
		t.Fatalf("node list: %d %+v", code, list)
	}
	// Nothing changed: no version, no audit.
	v := h.latestVersion(t)
	if err := f.Check(context.Background()); err != nil || h.latestVersion(t) != v || h.audits(t, "relay.failover") != 1 {
		t.Fatalf("a second check: %v, version %d (was %d)", err, h.latestVersion(t), v)
	}

	// Back: healthy since now, then steady for 5 minutes.
	back := func() {
		h.heartbeat(t, "vps-1")
		h.heartbeat(t, "vps-2", store.PeerStat{PeerID: vps1.ID, LastHandshake: h.clock.Now().Unix()})
		h.heartbeat(t, "vps-3")
	}
	back()
	if err := f.Check(context.Background()); err != nil || h.node(t, "vps-1").HealthySince != h.clock.Now().Unix() || h.latestVersion(t) != v {
		t.Fatalf("healthy again: %v, %+v", err, h.node(t, "vps-1"))
	}
	h.clock.Advance(5 * time.Minute)
	back()
	if err := f.Check(context.Background()); err != nil || h.relayOf(t, "nas") != "vps-1" || h.audits(t, "relay.restore") != 1 ||
		h.node(t, "vps-1").DownSince != 0 {
		t.Fatalf("restored: %v, relay %q, %d restore audits", err, h.relayOf(t, "nas"), h.audits(t, "relay.restore"))
	}

	// A server that is no longer a hub loses its state; the plan does not
	// care about it.
	if err := h.st.Tx(context.Background(), func(tx *store.Tx) error { return tx.SetHubState(vps2.ID, 1, 0) }); err != nil {
		t.Fatal(err)
	}
	if _, err := h.nodes().SetDisabled(context.Background(), cli(), "vps-2", true); err != nil {
		t.Fatal(err)
	}
	if err := f.Check(context.Background()); err != nil || h.node(t, "vps-2").DownSince != 0 {
		t.Fatalf("a disabled server keeps its failover state: %v", err)
	}

	// A store that fails is an error that says so.
	h.db.n.Store(0)
	h.db.fail.Store(1)
	if err := f.Check(context.Background()); !errors.Is(err, errDisk) {
		t.Fatalf("a failing store = %v", err)
	}
}

// Run checks on its own until ctx ends; with DownAfter 0 it does nothing.
func TestFailoverRun(t *testing.T) {
	h := newHarness(t)
	h.seedServers(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	(&Failover{DB: h.db, Reconciler: h.rec, Now: h.clock.Now}).Run(ctx) // off: returns at once
	f := &Failover{DB: h.db, Reconciler: h.rec, Now: h.clock.Now, DownAfter: 40 * time.Millisecond}
	done := make(chan struct{})
	go func() {
		defer close(done)
		f.Run(ctx)
	}()
	// The clock moves on while vps-2 and vps-3 keep reporting; whenever Run
	// noted its start, the silent vps-1 is soon long enough silent.
	deadline := time.Now().Add(5 * time.Second)
	for h.audits(t, "relay.failover") == 0 {
		if time.Now().After(deadline) {
			t.Fatal("Run did not judge the silent vps-1 down")
		}
		h.clock.Advance(time.Second)
		h.heartbeat(t, "vps-2")
		h.heartbeat(t, "vps-3")
		time.Sleep(10 * time.Millisecond)
	}
	cancel()
	<-done
	// Check alone takes the wall clock and discards its log.
	if err := (&Failover{DB: h.db, Reconciler: h.rec, DownAfter: time.Hour}).Check(context.Background()); err != nil {
		t.Fatal(err)
	}
}
