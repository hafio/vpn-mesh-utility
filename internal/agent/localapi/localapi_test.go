package localapi

import (
	"bufio"
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/hafio/vpn-mesh-utility/internal/agent/logbuf"
	"github.com/hafio/vpn-mesh-utility/internal/proto"
	"github.com/hafio/vpn-mesh-utility/internal/testfix"
)

// shortDir is a socket directory short enough for sun_path on every runner.
func shortDir(t *testing.T) string {
	t.Helper()
	d, err := os.MkdirTemp("", "mla")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(d) })
	return d
}

func TestStatusOverTheSocket(t *testing.T) {
	path := filepath.Join(shortDir(t), "agent.sock")
	ln, err := Listen(path)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	want := proto.LocalStatus{Node: proto.Self{ID: 1, Name: "vps-1", Role: proto.RoleServer}, AppliedVersion: 7, Peers: []proto.PeerStatus{{Name: "vps-2"}}}
	go func() {
		done <- Serve(ctx, ln, Handler(Funcs{Status: func(context.Context) (proto.LocalStatus, error) { return want, nil }}))
	}()
	got, err := NewClient(path).Status(context.Background())
	if err != nil || got.Node.Name != "vps-1" || got.AppliedVersion != 7 || len(got.Peers) != 1 {
		t.Fatalf("Status = %+v, %v", got, err)
	}
	if runtime.GOOS != "windows" {
		st, err := os.Stat(path)
		if err != nil || st.Mode().Perm()&0o007 != 0 {
			t.Fatalf("socket mode %v (%v): other users must not reach it", st.Mode(), err)
		}
	}
	if _, err := Listen(path); err == nil || !strings.Contains(err.Error(), "already running") {
		t.Fatalf("second Listen = %v, want a refusal", err)
	}
	cancel()
	if err := <-done; err != nil {
		t.Fatalf("Serve = %v", err)
	}
}

func TestStatusErrorsAreReported(t *testing.T) {
	path := filepath.Join(shortDir(t), "agent.sock")
	ln, err := Listen(path)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		_ = Serve(ctx, ln, Handler(Funcs{Status: func(context.Context) (proto.LocalStatus, error) {
			return proto.LocalStatus{}, errors.New("engine busy")
		}}))
	}()
	if _, err := NewClient(path).Status(context.Background()); err == nil || !strings.Contains(err.Error(), "engine busy") {
		t.Fatalf("Status = %v", err)
	}
}

// serve runs the API on a fresh socket until the test ends.
func serve(t *testing.T, f Funcs) *Client {
	t.Helper()
	path := filepath.Join(shortDir(t), "agent.sock")
	ln, err := Listen(path)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	go func() { _ = Serve(ctx, ln, Handler(f)) }()
	return NewClient(path)
}

// A prefs change is checked before it leaves the node, and the server's
// refusal comes back as it was sent.
func TestPrefsAreRelayed(t *testing.T) {
	var got []proto.PrefsRequest
	answer := func(_ context.Context, req proto.PrefsRequest) (proto.PrefsResponse, error) {
		got = append(got, req)
		switch {
		case req.Exit != nil && *req.Exit == "vps-3":
			return proto.PrefsResponse{}, &Refusal{Status: 409, Code: proto.CodeConflict, Message: `exit "vps-3" is not exit-capable`}
		case req.Exit != nil && *req.Exit == "vps-9":
			return proto.PrefsResponse{}, errors.New("POST /api/v1/agent/prefs: connection refused")
		}
		return proto.PrefsResponse{Version: 12, ConfigVersion: 12, ConfigHash: "abc"}, nil
	}
	c := serve(t, Funcs{Prefs: answer})
	exit := func(s string) proto.PrefsRequest { return proto.PrefsRequest{Exit: &s} }

	resp, err := c.Prefs(context.Background(), exit("vps-2"))
	if err != nil || resp.Version != 12 || resp.ConfigHash != "abc" {
		t.Fatalf("Prefs = %+v, %v", resp, err)
	}
	var rf *Refusal
	if _, err := c.Prefs(context.Background(), exit("vps-3")); !errors.As(err, &rf) || rf.Status != 409 || !strings.Contains(rf.Message, "not exit-capable") {
		t.Fatalf("a server refusal = %v", err)
	}
	if _, err := c.Prefs(context.Background(), exit("vps-9")); !errors.As(err, &rf) || rf.Status != 502 || !strings.Contains(err.Error(), "connection refused") {
		t.Fatalf("an unreachable server = %v", err)
	}
	if _, err := c.Prefs(context.Background(), exit("vps-2;reboot")); !errors.As(err, &rf) || rf.Status != 400 {
		t.Fatalf("a hostile exit = %v", err)
	}
	if _, err := c.Prefs(context.Background(), proto.PrefsRequest{}); !errors.As(err, &rf) || rf.Status != 400 {
		t.Fatalf("an empty change = %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("Prefs reached the agent %d times, want 3 (refused requests stop at the socket)", len(got))
	}
}

// `meshctl dns on|off` stays on the node; a body that is not a DNS request
// never reaches the agent, and an agent without the route says so.
func TestDNSIsLocal(t *testing.T) {
	var got []bool
	c := serve(t, Funcs{DNS: func(_ context.Context, on bool) (proto.LocalStatus, error) {
		got = append(got, on)
		if on {
			return proto.LocalStatus{DNS: proto.DNSOn}, nil
		}
		return proto.LocalStatus{DNS: proto.DNSOff}, nil
	}})
	ctx := context.Background()
	if st, err := c.DNS(ctx, false); err != nil || st.DNS != proto.DNSOff {
		t.Fatalf("dns off = %+v, %v", st, err)
	}
	if st, err := c.DNS(ctx, true); err != nil || st.DNS != proto.DNSOn {
		t.Fatalf("dns on = %+v, %v", st, err)
	}
	var rf *Refusal
	var st proto.LocalStatus
	if err := c.call(ctx, http.MethodPost, proto.LocalPathDNS, map[string]any{"on": true, "server": "x"}, &st); !errors.As(err, &rf) ||
		rf.Status != 400 || !strings.Contains(rf.Message, "not a DNS request") {
		t.Fatalf("an unknown field = %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("the agent was asked %d times, want 2", len(got))
	}
	if _, err := serve(t, Funcs{}).DNS(ctx, true); err == nil {
		t.Fatal("an agent without the DNS route answered")
	}
}

// The mesh view passes through as the agent has it, the reason it is stale
// included.
func TestMeshView(t *testing.T) {
	want := proto.LocalMesh{View: testfix.MeshView(), FetchedAt: testfix.Epoch.Unix(), Error: "the server did not answer"}
	c := serve(t, Funcs{Mesh: func(context.Context) (proto.LocalMesh, error) { return want, nil }})
	got, err := c.Mesh(context.Background())
	if err != nil || got.FetchedAt != want.FetchedAt || got.Error != want.Error || len(got.View.Nodes) != 6 || got.View.Nodes[4] != want.View.Nodes[4] {
		t.Fatalf("Mesh = %+v, %v", got, err)
	}
	if _, err := serve(t, Funcs{}).Mesh(context.Background()); err == nil || !strings.Contains(err.Error(), "answered 404") {
		t.Fatalf("an agent without the route = %v", err)
	}
}

// `meshctl ui` gets the agent's link, or its reason for having none.
func TestUIToken(t *testing.T) {
	want := proto.UIToken{URL: "http://127.0.0.1:7443/login?t=" + strings.Repeat("a", 43), ExpiresAt: testfix.Epoch.Unix()}
	c := serve(t, Funcs{UIToken: func(context.Context) (proto.UIToken, error) { return want, nil }})
	if got, err := c.UIToken(context.Background()); err != nil || got != want {
		t.Fatalf("UIToken = %+v, %v", got, err)
	}
	off := serve(t, Funcs{UIToken: func(context.Context) (proto.UIToken, error) {
		return proto.UIToken{}, &Refusal{Status: http.StatusConflict, Code: proto.CodeConflict, Message: "the local UI is off on this node"}
	}})
	var rf *Refusal
	if _, err := off.UIToken(context.Background()); !errors.As(err, &rf) || rf.Status != http.StatusConflict || !strings.Contains(rf.Message, "is off") {
		t.Fatalf("UIToken with the UI off = %v", err)
	}
	if _, err := serve(t, Funcs{}).UIToken(context.Background()); err == nil || !strings.Contains(err.Error(), "answered 404") {
		t.Fatalf("an agent without the route = %v", err)
	}
}

// `meshctl logs` prints what the ring holds; -f follows it until the reader
// leaves, and says so when the agent goes away under it.
func TestLogs(t *testing.T) {
	ring := logbuf.New(10)
	_, _ = io.WriteString(ring, "first\nsecond\n")
	c := serve(t, Funcs{Logs: ring})
	var out strings.Builder
	if err := c.Logs(context.Background(), false, &out); err != nil || out.String() != "first\nsecond\n" {
		t.Fatalf("Logs = %q, %v", out.String(), err)
	}
	var rf *Refusal
	if err := c.call(context.Background(), http.MethodGet, proto.LocalPathLogs+"?follow=yes", nil, nil); !errors.As(err, &rf) ||
		rf.Status != http.StatusBadRequest || !strings.Contains(rf.Message, "follow") {
		t.Fatalf("follow=yes = %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	pr, pw := io.Pipe()
	done := make(chan error, 1)
	go func() {
		done <- c.Logs(ctx, true, pw)
		_ = pw.Close()
	}()
	read := func(rd *bufio.Reader) string {
		t.Helper()
		l, err := rd.ReadString('\n')
		if err != nil {
			t.Fatalf("follow: %v", err)
		}
		return l
	}
	rd := bufio.NewReader(pr)
	if a, b := read(rd), read(rd); a != "first\n" || b != "second\n" {
		t.Fatalf("held lines %q %q", a, b)
	}
	_, _ = io.WriteString(ring, "third\n")
	if l := read(rd); l != "third\n" {
		t.Fatalf("followed %q", l)
	}
	cancel()
	go func() { _, _ = io.Copy(io.Discard, pr) }()
	if err := <-done; err != nil {
		t.Fatalf("an interrupted follow = %v", err)
	}

	// The agent stops: the stream ends, and the reader is told why.
	path := filepath.Join(shortDir(t), "agent.sock")
	ln, err := Listen(path)
	if err != nil {
		t.Fatal(err)
	}
	sctx, stop := context.WithCancel(context.Background())
	defer stop()
	go func() { _ = Serve(sctx, ln, Handler(Funcs{Logs: ring})) }()
	pr2, pw2 := io.Pipe()
	go func() {
		done <- NewClient(path).Logs(context.Background(), true, pw2)
		_ = pw2.Close()
	}()
	rd2 := bufio.NewReader(pr2)
	if got := read(rd2) + read(rd2) + read(rd2); got != "first\nsecond\nthird\n" {
		t.Fatalf("held lines %q", got)
	}
	stop()
	go func() { _, _ = io.Copy(io.Discard, pr2) }()
	if err := <-done; err == nil || !strings.Contains(err.Error(), "stopped") {
		t.Fatalf("a follow when the agent stops = %v", err)
	}

	if err := NewClient(filepath.Join(shortDir(t), "none.sock")).Logs(context.Background(), false, &out); err == nil || !strings.Contains(err.Error(), "agent run") {
		t.Fatalf("logs without an agent = %v", err)
	}
	if err := serve(t, Funcs{}).Logs(context.Background(), false, &out); err == nil || !strings.Contains(err.Error(), "answered 404") {
		t.Fatalf("logs from an agent without the route = %v", err)
	}

	// A reader the ring dropped is told to start again.
	out.Reset()
	if err := serve(t, Funcs{Logs: droppedLogs{}}).Logs(context.Background(), true, &out); err == nil ||
		!strings.Contains(out.String(), "held\nmeshctl: this reader fell behind") {
		t.Fatalf("a dropped reader got %q, %v", out.String(), err)
	}
}

// droppedLogs is a log whose follower has fallen behind at once.
type droppedLogs struct{}

func (droppedLogs) Lines() []string { return nil }

func (droppedLogs) Follow(context.Context) ([]string, <-chan string) {
	ch := make(chan string)
	close(ch)
	return []string{"held"}, ch
}

func TestClientExplainsAMissingAgent(t *testing.T) {
	_, err := NewClient(filepath.Join(shortDir(t), "none.sock")).Status(context.Background())
	if err == nil || !strings.Contains(err.Error(), "agent run") {
		t.Fatalf("Status = %v", err)
	}
}

func TestListenRemovesAStaleSocket(t *testing.T) {
	path := filepath.Join(shortDir(t), "agent.sock")
	if err := os.WriteFile(path, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	ln, err := Listen(path)
	if err != nil {
		t.Fatalf("Listen over a stale file: %v", err)
	}
	_ = ln.Close()
	if _, err := Listen(strings.Repeat("x", 120)); err == nil {
		t.Fatal("an over-long path was accepted")
	}
}

// Listen says which path it could not use: a directory it cannot create, or
// a directory where the socket should be.
func TestListenExplainsAnUnusablePath(t *testing.T) {
	d := shortDir(t)
	file := filepath.Join(d, "file")
	if err := os.WriteFile(file, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Listen(filepath.Join(file, "agent.sock")); err == nil || !strings.Contains(err.Error(), "create "+file) {
		t.Errorf("a socket under a file: %v", err)
	}
	busy := filepath.Join(d, "busy")
	if err := os.MkdirAll(filepath.Join(busy, "x"), 0o700); err != nil {
		t.Fatal(err)
	}
	if _, err := Listen(busy); err == nil || !strings.Contains(err.Error(), "remove stale socket") {
		t.Errorf("a directory in the socket's place: %v", err)
	}
}

func TestServeReportsAClosedListener(t *testing.T) {
	ln, err := Listen(filepath.Join(shortDir(t), "agent.sock"))
	if err != nil {
		t.Fatal(err)
	}
	_ = ln.Close()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := Serve(ctx, ln, Handler(Funcs{})); err == nil {
		t.Fatal("Serve on a closed listener succeeded")
	}
}

// The client explains answers that are not the agent's: a request body the
// API does not take, a route a status-only agent does not have, and
// something else answering on the socket.
func TestClientExplainsUnexpectedAnswers(t *testing.T) {
	ok := func(context.Context) (proto.LocalStatus, error) { return proto.LocalStatus{}, nil }
	c := serve(t, Funcs{Status: ok, Prefs: func(context.Context, proto.PrefsRequest) (proto.PrefsResponse, error) {
		t.Error("a malformed request reached Prefs")
		return proto.PrefsResponse{}, nil
	}})
	var rf *Refusal
	err := c.call(context.Background(), http.MethodPost, proto.LocalPathPrefs, map[string]bool{"can_exit": true}, &proto.PrefsResponse{})
	if !errors.As(err, &rf) || rf.Status != http.StatusBadRequest || !strings.Contains(rf.Message, "not a prefs request") {
		t.Errorf("an unknown field = %v", err)
	}

	statusOnly := serve(t, Funcs{Status: ok})
	if _, err := statusOnly.Prefs(context.Background(), proto.PrefsRequest{}); err == nil || !strings.Contains(err.Error(), "answered 404") {
		t.Errorf("prefs on a status-only agent = %v", err)
	}

	path := filepath.Join(shortDir(t), "other.sock")
	ln, err := net.Listen("unix", path)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	go func() {
		_ = Serve(ctx, ln, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte("hello")) }))
	}()
	if _, err := NewClient(path).Status(context.Background()); err == nil || !strings.Contains(err.Error(), "decode the agent's answer") {
		t.Errorf("another program on the socket = %v", err)
	}
}
