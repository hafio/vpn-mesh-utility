package api

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/hafio/vpn-mesh-utility/internal/proto"
	"github.com/hafio/vpn-mesh-utility/internal/server/store"
)

// Relay failover (amendment A40). A hub is down when its heartbeat is
// DownAfter old AND no node reported a handshake with it for 1.5 times
// that: both, so a control plane that cannot reach it -- or its agent
// alone stopping -- moves no traffic while its tunnels still work. It is
// back once its heartbeat has been steady for 2.5 times DownAfter: an agent
// that reports again has its tunnel up, while handshakes renew only every
// two minutes or so, whatever DownAfter is. Each change is
// one reconciler write by actor "system", audited relay.failover or
// relay.restore, so the pairs it relays move to another hub in a new config
// version and back. Phones cannot move: their profile names one hub.

// Default failover timing.
const DefaultHubDownAfter = 2 * time.Minute

// Failover watches the hubs and moves relays off one that is down.
type Failover struct {
	DB         DB
	Reconciler *Reconciler
	Now        func() time.Time
	Log        *slog.Logger
	// DownAfter is how long a hub's heartbeat may be missing; zero turns
	// failover off.
	DownAfter time.Duration
	// started is when the monitor began: nothing reports to a server that
	// is not running, so no silence before it counts.
	started time.Time
}

func (f *Failover) handshakeAfter() time.Duration { return f.DownAfter * 3 / 2 }
func (f *Failover) restoreAfter() time.Duration   { return f.DownAfter * 5 / 2 }

func (f *Failover) defaults() {
	if f.Log == nil {
		f.Log = slog.New(slog.DiscardHandler)
	}
	if f.Now == nil {
		f.Now = time.Now
	}
}

// Run checks every quarter of DownAfter until ctx ends.
func (f *Failover) Run(ctx context.Context) {
	if f.DownAfter <= 0 {
		return
	}
	f.defaults()
	f.started = f.Now()
	t := time.NewTicker(f.DownAfter / 4)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
		}
		if err := f.Check(ctx); err != nil && !errors.Is(err, context.Canceled) {
			f.Log.Error("relay failover: the hubs were not checked; the next check tries again", "err", err)
		}
	}
}

// hubState is a hub's failover state, as store.Node keeps it.
type hubState struct {
	down, healthy int64
}

// judge is a hub's next state at now, from its last heartbeat and the
// latest handshake any node reported with it (unix seconds).
func (f *Failover) judge(now time.Time, lastSeen, lastHandshake int64, cur hubState) hubState {
	fresh := func(at int64, within time.Duration) bool {
		if s := f.started.Unix(); at < s {
			at = s
		}
		return now.Sub(time.Unix(at, 0)) < within
	}
	heartbeat, handshake := fresh(lastSeen, f.DownAfter), fresh(lastHandshake, f.handshakeAfter())
	switch {
	case cur.down == 0 && !heartbeat && !handshake:
		return hubState{down: now.Unix()}
	case cur.down == 0:
		return cur
	case !heartbeat:
		return hubState{down: cur.down} // not steady yet
	case cur.healthy == 0:
		return hubState{down: cur.down, healthy: now.Unix()}
	case now.Sub(time.Unix(cur.healthy, 0)) >= f.restoreAfter():
		return hubState{}
	}
	return cur
}

// hubMove is one hub whose down state changes.
type hubMove struct {
	node     store.Node
	next     hubState
	lastSeen int64
}

// Check judges every planned hub of every mesh once: a hub that goes down
// or comes back is one reconciler write in its mesh; one that only starts
// or stops looking healthy again is recorded as it is.
func (f *Failover) Check(ctx context.Context) error {
	f.defaults()
	now := f.Now()
	var moves []hubMove
	err := f.DB.Tx(ctx, func(tx *store.Tx) error {
		meshes, err := tx.Meshes()
		if err != nil {
			return err
		}
		var nodes []store.Node
		for _, m := range meshes {
			in, err := tx.In(m.ID).Nodes()
			if err != nil {
				return err
			}
			nodes = append(nodes, in...)
		}
		for _, n := range nodes {
			cur := hubState{down: n.DownSince, healthy: n.HealthySince}
			if n.Role != string(proto.RoleServer) || !n.Enrolled || n.Disabled || n.Endpoint == "" {
				if cur != (hubState{}) {
					// Not a hub any more: it relays nothing, so nothing moves.
					if err := tx.SetHubState(n.ID, 0, 0); err != nil {
						return err
					}
				}
				continue
			}
			hs, err := tx.LastHandshakeWith(n.ID)
			if err != nil {
				return err
			}
			next := f.judge(now, n.LastSeen, hs, cur)
			switch {
			case next == cur:
			case (next.down == 0) != (cur.down == 0):
				moves = append(moves, hubMove{node: n, next: next, lastSeen: n.LastSeen})
			default:
				if err := tx.SetHubState(n.ID, next.down, next.healthy); err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	for _, m := range moves {
		action, detail := "relay.restore", map[string]any{"down_since": m.node.DownSince}
		if m.next.down != 0 {
			action, detail = "relay.failover", map[string]any{"last_seen": m.lastSeen}
		}
		res, err := f.Reconciler.Apply(ctx, Change{Mesh: m.node.MeshID, Actor: "system", Action: action, Target: m.node.Name, Detail: detail},
			func(tx *store.MeshTx, _ *Change) error {
				return tx.SetHubState(m.node.ID, m.next.down, m.next.healthy)
			})
		if err != nil {
			return err
		}
		if m.next.down != 0 {
			f.Log.Warn("relay failover: a hub is down (no heartbeat and no handshake); the pairs it relays go through other hubs",
				"hub", m.node.Name, "version", res.Version)
		} else {
			f.Log.Info("relay failover: a hub is back; its pairs relay through it again", "hub", m.node.Name, "version", res.Version)
		}
	}
	return nil
}
