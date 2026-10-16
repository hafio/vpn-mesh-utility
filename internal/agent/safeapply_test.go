package agent

import (
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/hafio/vpn-mesh-utility/internal/agent/state"
	"github.com/hafio/vpn-mesh-utility/internal/proto"
	"github.com/hafio/vpn-mesh-utility/internal/testfix"
)

// version is a server config that hashes apart from every other version;
// even versions have one peer, odd ones two.
func version(n int64) proto.NodeConfig {
	c := testfix.ServerConfig()
	c.Version, c.Interface.MTU = n, 1400+int(n)
	if n%2 == 0 {
		c.Peers = c.Peers[:1]
	}
	return c
}

func TestSafeApplyStateMachine(t *testing.T) {
	clock := testfix.NewClock(testfix.Epoch)
	s := &safeApply{window: time.Minute}
	if _, ok := s.confirm(clock.Now()); ok {
		t.Fatal("confirmed with nothing pending")
	}
	s.arm(version(4), "h4", version(3), clock.Now())
	if s.confirmedHash() == "" {
		t.Fatal("a pending apply knows the confirmed config's hash")
	}
	if _, ok := s.confirm(clock.Now().Add(-time.Second)); ok {
		t.Fatal("a report sent before the apply confirmed it")
	}
	clock.Advance(59 * time.Second)
	if _, ok := s.expired(clock.Now()); ok {
		t.Fatal("expired inside the window")
	}
	if p, ok := s.confirm(clock.Now()); !ok || p.cfg.Version != 4 || s.pending != nil {
		t.Fatalf("confirm = %+v, %v", p, ok)
	}

	// A config that replaces a pending one keeps the confirmed one as prev.
	s.arm(version(5), "h5", version(4), clock.Now())
	s.arm(version(6), "h6", version(5), clock.Now())
	clock.Advance(time.Minute)
	p, ok := s.expired(clock.Now())
	if !ok || p.cfg.Version != 6 || p.prev.Version != 4 || s.pending != nil {
		t.Fatalf("expired = v%d over v%d, %v", p.cfg.Version, p.prev.Version, ok)
	}
	if s.confirmedHash() != "" {
		t.Fatal("nothing is pending any more")
	}
	s.arm(version(7), "h7", version(6), clock.Now())
	s.disarm()
	if _, ok := s.expired(clock.Now().Add(time.Hour)); ok {
		t.Fatal("a disarmed apply expired")
	}
}

// safeRig is a server agent on a fake clock that starts at the real time,
// so signatures stay inside the server's skew window, with a 1 s window.
func safeRig(t *testing.T, dir *state.Dir) (*rig, *testfix.Clock) {
	t.Helper()
	clock := testfix.NewClock(time.Now())
	r := startWith(t, dir, "", func(d *Deps) {
		d.Now, d.ConfirmWindow, d.SafeTick = clock.Now, time.Second, 10*time.Millisecond
	})
	return r, clock
}

func pendingExists(dir *state.Dir) bool {
	_, err := os.Stat(filepath.Join(dir.Path, state.PendingConfigFile))
	return err == nil
}

func TestSafeApplyConfirmsWhenTheServerAnswers(t *testing.T) {
	dir, f := joinedTo(t, version(3))
	r, clock := safeRig(t, dir)
	eventually(t, "v3", func() bool { return r.peers() == 2 })
	f.push <- f.set(version(4))
	eventually(t, "v4 confirmed and cached", func() bool {
		c, _ := dir.Config()
		return r.peers() == 1 && c != nil && c.Version == 4 && !pendingExists(dir)
	})
	clock.Advance(2 * time.Second)
	time.Sleep(50 * time.Millisecond) // several safe-apply ticks
	if r.peers() != 1 {
		t.Fatal("a confirmed config was rolled back")
	}
}

func TestSafeApplyRollsBackWithoutAnAnswer(t *testing.T) {
	dir, f := joinedTo(t, version(3))
	r, clock := safeRig(t, dir)
	eventually(t, "v3", func() bool { return r.peers() == 2 })

	f.setStatusDown(true)
	f.push <- f.set(version(4))
	eventually(t, "v4 applied and pending", func() bool { return r.peers() == 1 && pendingExists(dir) })
	if c, _ := dir.Config(); c.Version != 3 {
		t.Fatalf("the cache moved to v%d before confirmation", c.Version)
	}
	clock.Advance(2 * time.Second)
	eventually(t, "the rollback to v3", func() bool { return r.peers() == 2 && !pendingExists(dir) })

	// The server answers again: the rollback is reported, and v4, still the
	// server's config, is not applied again.
	f.setStatusDown(false)
	eventually(t, "the rollback in a report", func() bool {
		rep, ok := f.lastReport()
		return ok && rep.Rollback != nil && rep.Rollback.From == 4 && rep.Rollback.To == 3 && rep.AppliedVersion == 3 &&
			strings.Contains(rep.Rollback.Reason, "no answer from the server")
	})
	time.Sleep(200 * time.Millisecond) // a few polls that fetch v4
	if r.peers() != 2 {
		t.Fatal("the rolled-back v4 was applied again")
	}

	// v5 is a different config: it is applied, confirmed, and the rollback
	// is over.
	f.push <- f.set(version(5))
	eventually(t, "v5 confirmed without the rollback", func() bool {
		rep, ok := f.lastReport()
		c, _ := dir.Config()
		return ok && rep.AppliedVersion == 5 && rep.Rollback == nil && c != nil && c.Version == 5
	})
}

func TestSafeApplyRecoversAPendingConfigAtBoot(t *testing.T) {
	dir, f := joinedTo(t, version(4))
	if err := dir.SaveConfig(version(3)); err != nil {
		t.Fatal(err)
	}
	if err := dir.SavePending(version(4)); err != nil { // the agent stopped inside the window
		t.Fatal(err)
	}
	r, _ := safeRig(t, dir)
	eventually(t, "v3 back with the rollback reported", func() bool {
		rep, ok := f.lastReport()
		return r.peers() == 2 && ok && rep.Rollback != nil && rep.Rollback.From == 4 && rep.Rollback.To == 3 &&
			strings.Contains(rep.Rollback.Reason, "stopped before")
	})
	if pendingExists(dir) {
		t.Fatal("the pending config survived the recovery")
	}
	time.Sleep(200 * time.Millisecond)
	if r.peers() != 2 {
		t.Fatal("the never-confirmed v4 was applied again")
	}
}

// The server going back to the confirmed config while a change is pending
// ends safe apply: nothing is rolled back, and the confirmed config stays
// the cached one.
func TestSafeApplyDisarmsWhenTheServerGoesBack(t *testing.T) {
	dir, f := joinedTo(t, version(3))
	clock := testfix.NewClock(time.Now())
	var logs syncBuffer
	r := startWith(t, dir, "", func(d *Deps) {
		d.Now, d.ConfirmWindow, d.SafeTick = clock.Now, time.Second, 10*time.Millisecond
		d.Log = slog.New(slog.NewTextHandler(&logs, nil))
	})
	eventually(t, "v3", func() bool { return r.peers() == 2 })

	f.setStatusDown(true)
	f.push <- f.set(version(4))
	eventually(t, "v4 applied and pending", func() bool { return r.peers() == 1 && pendingExists(dir) })
	f.push <- f.set(version(3))
	eventually(t, "v3 back and nothing pending", func() bool { return r.peers() == 2 && !pendingExists(dir) })
	clock.Advance(2 * time.Second)
	time.Sleep(50 * time.Millisecond) // several safe-apply ticks
	if strings.Contains(logs.String(), "rolled back") {
		t.Fatalf("a disarmed apply was rolled back:\n%s", logs.String())
	}
	if c, _ := dir.Config(); c == nil || c.Version != 3 {
		t.Fatalf("cached config %+v, want v3", c)
	}
}

// A rollback that fails too is reported, as an error and as the rollback.
func TestSafeApplyReportsAFailedRollback(t *testing.T) {
	dir, f := joinedTo(t, version(3))
	r, clock := safeRig(t, dir)
	eventually(t, "v3", func() bool { return r.peers() == 2 })

	f.setStatusDown(true)
	f.push <- f.set(version(4))
	eventually(t, "v4 applied and pending", func() bool { return r.peers() == 1 && pendingExists(dir) })
	r.host.SetFail("Configure")
	clock.Advance(2 * time.Second)
	eventually(t, "the rollback attempt", func() bool { return !pendingExists(dir) })
	f.setStatusDown(false)
	eventually(t, "the failed rollback in a report", func() bool {
		rep, ok := f.lastReport()
		return ok && rep.Rollback != nil && rep.Rollback.From == 4 && rep.Rollback.To == 3 && len(rep.Errors) > 0 &&
			strings.Contains(rep.Errors[0], "rolling back to v3 failed")
	})
}

// Spokes have no safe apply: a pushed config is cached at once.
func TestSpokesDoNotArmSafeApply(t *testing.T) {
	dir, pub := newDir(t)
	f, srv := newFakeServer(t, pub, testfix.SpokeConfig(""), nil)
	enrollAs(t, dir, srv.URL, testfix.Laptop)
	f.setStatusDown(true) // no confirmation would ever come
	r, clock := safeRig(t, dir)
	eventually(t, "the first config", func() bool { return r.peers() == 3 })
	next := testfix.SpokeConfig("")
	next.Version = 6
	next.Peers = next.Peers[:2]
	f.push <- f.set(next)
	eventually(t, "v6 cached", func() bool {
		c, _ := dir.Config()
		return r.peers() == 2 && c != nil && c.Version == 6 && !pendingExists(dir)
	})
	clock.Advance(2 * time.Second) // past the window, inside the signature skew
	time.Sleep(50 * time.Millisecond)
	if r.peers() != 2 {
		t.Fatal("a spoke rolled back")
	}
}
