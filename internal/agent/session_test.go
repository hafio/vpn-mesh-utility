package agent

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/hafio/vpn-mesh-utility/internal/agent/hostid"
	"github.com/hafio/vpn-mesh-utility/internal/agent/localapi"
	"github.com/hafio/vpn-mesh-utility/internal/agent/state"
	"github.com/hafio/vpn-mesh-utility/internal/proto"
	"github.com/hafio/vpn-mesh-utility/internal/testfix"
)

// endsWith waits for r's run to end and returns why.
func endsWith(t *testing.T, r *rig) error {
	t.Helper()
	select {
	case err := <-r.done:
		r.done <- err // for the cleanup
		return err
	case <-time.After(10 * time.Second):
		t.Fatal("the run did not end")
		return nil
	}
}

// Without a saved session the agent says hello (amendment A50), keeps the
// session it is given and host.json, and runs; a clean stop says goodbye
// and marks the session released.
func TestHelloAndBye(t *testing.T) {
	cfg := testfix.ServerConfig()
	dir, f := joinedTo(t, cfg)
	if err := dir.ClearSession(); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(dir.Path, state.HostFile)); err != nil {
		t.Fatal(err)
	}
	r := start(t, dir, "")
	// host.json is written once the hello is answered, which may be after
	// the config the session let through is applied.
	eventually(t, "the config applied and host.json written after the hello", func() bool {
		h, ok, _ := dir.Host()
		return r.peers() == len(cfg.Peers) && ok && h.Machine == hostid.Local(testMachine)
	})
	if hellos, _, _ := f.sessions(); hellos != 1 {
		t.Fatalf("%d hellos", hellos)
	}
	if s, ok, err := dir.Session(); !ok || err != nil || s.Session != testfix.Session(testfix.VPS1.ID) {
		t.Fatalf("session.json %+v, %v, %v", s, ok, err)
	}
	r.stop(t)
	if _, byes, _ := f.sessions(); byes != 1 {
		t.Fatalf("%d goodbyes", byes)
	}
	if s, _, _ := dir.Session(); !s.Released {
		t.Fatalf("session.json after a clean stop: %+v", s)
	}
	// The next start says hello again, at once.
	r = start(t, dir, "")
	eventually(t, "a second hello", func() bool { h, _, _ := f.sessions(); return h == 2 })
	r.stop(t)
}

// The cached config is applied at boot only where the data dir last ran:
// on another machine the agent waits for the control plane.
func TestCacheOnlyWhereTheDataDirRan(t *testing.T) {
	cfg := testfix.ServerConfig()
	dir, _ := newDir(t)
	enroll(t, dir, "http://127.0.0.1:1") // nothing listens there
	if err := dir.SaveConfig(testfix.SignedConfig(cfg)); err != nil {
		t.Fatal(err)
	}
	if err := dir.SaveHost(state.Host{Machine: hostid.Local("another machine")}); err != nil {
		t.Fatal(err)
	}
	r := start(t, dir, "")
	time.Sleep(100 * time.Millisecond)
	if r.host.Mutations() != 0 {
		t.Fatalf("a data dir from another machine applied its cache: %v", r.host.Calls)
	}
	r.stop(t)
	ranHere(t, dir, testfix.VPS1.ID)
	r = start(t, dir, "")
	eventually(t, "the cache applied here", func() bool { return r.peers() == len(cfg.Peers) })
}

// A copy that is told to wait touches nothing, asks again, and on a signed
// duplicate stops (exit 3) with its enrollment and keys, its session gone
// and why it was refused in refused.json.
func TestWaitThenDuplicate(t *testing.T) {
	dir, f := joinedTo(t, testfix.ServerConfig())
	if err := dir.ClearSession(); err != nil {
		t.Fatal(err)
	}
	answers := 0
	f.hello = func() (*proto.HelloResponse, *proto.ErrorBody) {
		answers++
		if answers < 3 {
			return nil, &proto.ErrorBody{Code: proto.CodeWait, Message: "another copy holds the session"}
		}
		return nil, &proto.ErrorBody{Code: proto.CodeDuplicate, Message: "node vps-1 is running on vps-1.example.com"}
	}
	r := start(t, dir, "")
	if err := endsWith(t, r); !errors.Is(err, ErrRefused) {
		t.Fatalf("Run = %v", err)
	}
	if hellos, _, _ := f.sessions(); hellos != 3 || len(f.reportsCopy()) != 0 {
		t.Fatalf("%d hellos, %d reports", hellos, len(f.reportsCopy()))
	}
	if slices.ContainsFunc(r.host.Calls, func(c string) bool { return !strings.HasPrefix(c, "DeleteLink") }) {
		t.Fatalf("a waiting copy changed the host: %v", r.host.Calls)
	}
	rec, ok, _ := dir.Refused()
	if _, sok, _ := dir.Session(); !ok || rec.Code != proto.CodeDuplicate || sok {
		t.Fatalf("refused.json %+v, %v; session kept %v", rec, ok, sok)
	}
	if _, err := dir.Enrollment(); err != nil {
		t.Fatalf("a duplicate lost its enrollment: %v", err)
	}
	if _, err := dir.LoadKeys(); err != nil {
		t.Fatalf("a duplicate lost its keys: %v", err)
	}
}

// A copy on another machine than the node's wipes its data dir, keeping
// only why (with the token it joined with), so a container restarted with
// that token is told it needs a new one without asking the server.
func TestForeignWipes(t *testing.T) {
	dir, f := joinedTo(t, testfix.ServerConfig())
	if err := dir.ClearSession(); err != nil {
		t.Fatal(err)
	}
	if err := dir.SaveJoinRecord(state.JoinRecord{Token: tokenFingerprint(testfix.Token)}); err != nil {
		t.Fatal(err)
	}
	f.hello = func() (*proto.HelloResponse, *proto.ErrorBody) {
		return nil, &proto.ErrorBody{Code: proto.CodeForeign, Message: "node vps-1 runs on another machine"}
	}
	r := start(t, dir, "")
	if err := endsWith(t, r); !errors.Is(err, ErrRefused) {
		t.Fatalf("Run = %v", err)
	}
	entries, _ := os.ReadDir(dir.Path)
	var left []string
	for _, e := range entries {
		left = append(left, e.Name())
	}
	slices.Sort(left)
	if !slices.Equal(left, []string{state.LockFile, state.RefusedFile}) {
		t.Fatalf("left in the data dir: %v", left)
	}
	var paused time.Duration
	err := JoinOnStart(context.Background(), JoinOptions{DataDir: dir.Path, Server: "https://mesh.example.com", Token: testfix.Token,
		Pause: func(d time.Duration) { paused = d }})
	if err == nil || !strings.Contains(err.Error(), "new token") || paused != refusedPause {
		t.Fatalf("a start with the spent token = %v (paused %s)", err, paused)
	}
	if err := Run(context.Background(), Deps{DataDir: dir.Path}); !errors.Is(err, state.ErrNotJoined) || !strings.Contains(err.Error(), proto.CodeForeign) {
		t.Fatalf("agent run after the wipe = %v", err)
	}
}

// A session the control plane no longer knows sends the agent back to
// hello; the node keeps running throughout.
func TestSessionLostSaysHello(t *testing.T) {
	cfg := testfix.ServerConfig()
	dir, f := joinedTo(t, cfg)
	r := start(t, dir, "")
	eventually(t, "the config applied", func() bool { return r.peers() == len(cfg.Peers) })
	f.mu.Lock()
	f.lose = true
	f.mu.Unlock()
	eventually(t, "a hello after the loss", func() bool { h, _, _ := f.sessions(); return h == 1 })
	before := len(f.reportsCopy())
	eventually(t, "heartbeats after the hello", func() bool { return len(f.reportsCopy()) > before })
	if r.peers() != len(cfg.Peers) {
		t.Fatal("the tunnel went down while the agent said hello")
	}
}

// A challenge on an answer is answered at once with its nonce.
func TestChallengeIsAnswered(t *testing.T) {
	dir, f := joinedTo(t, testfix.ServerConfig())
	f.mu.Lock()
	f.challenge = "BBBBBBBBBBBBBBBBBBBBBBBBBB"
	f.mu.Unlock()
	start(t, dir, "")
	eventually(t, "the challenge answered", func() bool {
		_, _, alive := f.sessions()
		return slices.Contains(alive, "BBBBBBBBBBBBBBBBBBBBBBBBBB")
	})
}

// A second agent of the mesh in this network namespace is refused before
// it touches anything (amendment A50); another mesh's runs beside it.
func TestOneAgentPerMeshHere(t *testing.T) {
	cfg := testfix.ServerConfig()
	dir, _ := joinedTo(t, cfg)
	r := start(t, dir, "")
	eventually(t, "the first agent", func() bool { return r.peers() == len(cfg.Peers) })
	enr, _ := dir.Enrollment()
	unlock, err := state.LockMesh(meshKey(enr))
	switch {
	case errors.Is(err, state.ErrMeshLocked):
	case err == nil:
		unlock()
		t.Skip("no mesh lock on this platform")
	default:
		t.Fatal(err)
	}
	other := enr
	other.MeshUUID = ""
	if meshKey(enr) == meshKey(other) || meshKey(other) == meshKey(state.Enrollment{Server: "https://b"}) {
		t.Fatal("mesh keys collide")
	}
}

// reportsCopy is the heartbeats so far.
func (f *fakeServer) reportsCopy() []proto.StatusReport {
	f.mu.Lock()
	defer f.mu.Unlock()
	return slices.Clone(f.reports)
}

// Damaged session and host files, and a refusal from before, are named in
// the log; the agent says hello and runs. A heartbeat the hello answers
// shorter than --status-every is the one it keeps. While it has no
// session, a change of its own settings is refused at once.
func TestStartWithLeftovers(t *testing.T) {
	cfg := testfix.ServerConfig()
	dir, f := joinedTo(t, cfg)
	for name, body := range map[string]string{state.SessionFile: "{", state.HostFile: "{"} {
		if err := os.WriteFile(filepath.Join(dir.Path, name), []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if err := dir.SaveRefused(state.Refused{Code: proto.CodeDuplicate, At: 1, Message: "a second copy"}); err != nil {
		t.Fatal(err)
	}
	f.hello = func() (*proto.HelloResponse, *proto.ErrorBody) {
		return &proto.HelloResponse{Session: testfix.Session(1), Ticket: testfix.Ticket(1), Heartbeat: 1}, nil
	}
	var logs syncBuffer
	r := startWith(t, dir, "", func(d *Deps) {
		d.Log = slog.New(slog.NewTextHandler(&logs, nil))
		d.StatusEvery = time.Hour
	})
	eventually(t, "heartbeats every second", func() bool { return len(f.reportsCopy()) >= 2 })
	for _, want := range []string{"ignoring the saved session", "ignoring host.json", "refused this node here before"} {
		if !strings.Contains(logs.String(), want) {
			t.Errorf("the log does not say %q:\n%s", want, logs.String())
		}
	}
	if _, ok, _ := dir.Refused(); ok {
		t.Error("refused.json outlived a granted hello")
	}
	r.stop(t)

	// Told to wait, the agent refuses a change of its own exit at once.
	dir2, f2 := joinedTo(t, testfix.SpokeConfig(""))
	if err := dir2.ClearSession(); err != nil {
		t.Fatal(err)
	}
	f2.hello = func() (*proto.HelloResponse, *proto.ErrorBody) {
		return nil, &proto.ErrorBody{Code: proto.CodeWait, Message: "another copy"}
	}
	sock := filepath.Join(shortTemp(t), "agent.sock")
	start(t, dir2, sock)
	via := "vps-2"
	var rf *localapi.Refusal
	eventually(t, "a refused exit while waiting", func() bool {
		_, err := localapi.NewClient(sock).Prefs(context.Background(), proto.PrefsRequest{Exit: &via})
		return errors.As(err, &rf) && rf.Code == proto.CodeSession
	})
}
