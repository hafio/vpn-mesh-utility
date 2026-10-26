package client

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/hafio/vpn-mesh-utility/internal/proto"
	"github.com/hafio/vpn-mesh-utility/internal/testfix"
)

// ticketServer is a control plane's session side for the client tests: it
// checks that every request but a hello presents the ticket it issued last
// and hands out the next one, unless told otherwise.
type ticketServer struct {
	t  *testing.T
	mu sync.Mutex
	// ticket is the one a request must present; n counts tickets issued.
	ticket string
	n      int
	// seen lists the paths asked, with the ticket each presented.
	seen []string
	// fail answers the next request at path with this code (and no next
	// ticket); challenge puts a challenge on the next answer.
	fail      map[string]string
	challenge string
	alive     []string
}

func newTicketServer(t *testing.T) *ticketServer {
	return &ticketServer{t: t, ticket: testfix.Ticket(testfix.VPS1.ID), fail: map[string]string{}}
}

func (s *ticketServer) next() string {
	s.n++
	return strings.Repeat("0", 62) + string(rune('a'+s.n/16%6)) + string("0123456789abcdef"[s.n%16])
}

func (s *ticketServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()
	body, _ := io.ReadAll(r.Body)
	signed, err := proto.ParseSigned(r.Header)
	pub, _ := testfix.IdentityKey(testfix.VPS1.Key)
	if err == nil {
		err = signed.Verify(pub, r.Method, r.URL.RequestURI(), body, testfix.Epoch)
	}
	if err != nil {
		s.t.Errorf("%s %s: %v", r.Method, r.URL.Path, err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	s.seen = append(s.seen, r.URL.Path+" "+signed.Ticket)
	if r.URL.Path == proto.PathHello {
		s.ticket = s.next()
		writeJSONT(w, http.StatusOK, proto.HelloResponse{Session: testfix.Session(9), Ticket: s.ticket, Heartbeat: 10})
		return
	}
	if code := s.fail[r.URL.Path]; code != "" {
		delete(s.fail, r.URL.Path)
		writeJSONT(w, http.StatusConflict, proto.ErrorBody{Code: code, Message: "no"})
		return
	}
	if signed.Ticket != s.ticket && r.URL.Path != proto.PathEvents {
		s.t.Errorf("%s presented %q, want %q", r.URL.Path, signed.Ticket, s.ticket)
	}
	if r.URL.Path != proto.PathEvents {
		s.ticket = s.next()
		w.Header().Set(proto.HeaderNextTicket, s.ticket)
	}
	if s.challenge != "" && r.URL.Path != proto.PathAlive {
		w.Header().Set(proto.HeaderChallenge, s.challenge)
	}
	switch r.URL.Path {
	case proto.PathAlive:
		var a proto.AliveRequest
		_ = json.Unmarshal(body, &a)
		s.alive = append(s.alive, a.Nonce)
		s.challenge = ""
		w.WriteHeader(http.StatusNoContent)
	case proto.PathBye:
		w.WriteHeader(http.StatusNoContent)
	case proto.PathEvents:
		w.Header().Set("Content-Type", "text/event-stream")
		data, _ := json.Marshal(proto.ChallengeEvent{Nonce: testfix.Nonce})
		_ = proto.WriteEvent(w, proto.Event{Name: proto.EventChallenge, Data: data})
	default:
		writeJSONT(w, http.StatusOK, proto.StatusResponse{})
	}
}

func writeJSONT(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// sessionClient is a client of s with the session given, keeping every
// change to it.
func sessionClient(t *testing.T, s *ticketServer, sess *Session) (*Client, *[]Session) {
	t.Helper()
	srv := httptest.NewServer(s)
	t.Cleanup(srv.Close)
	_, priv := testfix.IdentityKey(testfix.VPS1.Key)
	var kept []Session
	c, err := New(Options{Server: srv.URL, NodeID: testfix.VPS1.ID, Key: priv, ConfigKey: testfix.Pinned(testfix.VPS1.ID).Key, Version: "v1.0.0",
		Now: func() time.Time { return testfix.Epoch }, Session: sess, Keep: func(s Session) { kept = append(kept, s) }})
	if err != nil {
		t.Fatal(err)
	}
	return c, &kept
}

// Every request presents the ticket the previous answer carried, and the
// new one is kept; an answer without one leaves the ticket where it was.
func TestTicketsRoll(t *testing.T) {
	s := newTicketServer(t)
	c, kept := sessionClient(t, s, testSession())
	for range 3 {
		if _, err := c.Status(context.Background(), proto.StatusReport{}); err != nil {
			t.Fatal(err)
		}
	}
	if got := c.Session().Ticket; got != s.ticket || len(*kept) != 3 || (*kept)[2].Ticket != got {
		t.Fatalf("ticket %q, server's %q; kept %v", got, s.ticket, *kept)
	}
	// A refusal with no next ticket: the same ticket again.
	s.fail[proto.PathStatus] = proto.CodeClockSkew
	before := c.Session().Ticket
	if _, err := c.Status(context.Background(), proto.StatusReport{}); err == nil {
		t.Fatal("a refusal was not an error")
	}
	if c.Session().Ticket != before || !c.HasSession() {
		t.Fatal("a refusal moved the ticket or dropped the session")
	}
	if _, err := c.Status(context.Background(), proto.StatusReport{}); err != nil {
		t.Fatalf("after a refusal: %v", err)
	}
}

// 409 session drops the session: requests wait, Lost is closed, and the
// data dir is told it is released; a hello takes a new one and the waiting
// request goes on.
func TestSessionLostAndHello(t *testing.T) {
	s := newTicketServer(t)
	c, kept := sessionClient(t, s, testSession())
	select {
	case <-c.Lost():
		t.Fatal("Lost is closed while there is a session")
	default:
	}
	s.fail[proto.PathStatus] = proto.CodeSession
	if _, err := c.Status(context.Background(), proto.StatusReport{}); err == nil {
		t.Fatal("409 session was not an error")
	}
	select {
	case <-c.Lost():
	default:
		t.Fatal("Lost is open after 409 session")
	}
	if c.HasSession() || !(*kept)[len(*kept)-1].Released {
		t.Fatalf("the session is still usable: kept %v", *kept)
	}
	if _, err := c.Prefs(context.Background(), proto.PrefsRequest{}); !errors.Is(err, ErrNoSession) {
		t.Fatalf("Prefs without a session = %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	if _, err := c.Status(ctx, proto.StatusReport{}); !errors.Is(err, ErrNoSession) {
		t.Fatalf("Status waiting for a session = %v", err)
	}
	done := make(chan error, 1)
	go func() {
		_, err := c.Status(context.Background(), proto.StatusReport{})
		done <- err
	}()
	resp, err := c.Hello(context.Background(), proto.HelloRequest{MachineID: testfix.MachineID, Hostname: "vps-1"})
	if err != nil || resp.Session != testfix.Session(9) || c.Session().Heartbeat != 10 {
		t.Fatalf("Hello = %+v, %v", resp, err)
	}
	if err := <-done; err != nil {
		t.Fatalf("the waiting request after the hello = %v", err)
	}
	if !strings.HasPrefix(s.seen[len(s.seen)-1], proto.PathStatus+" ") || strings.Contains(s.seen[len(s.seen)-2], testfix.Ticket(testfix.VPS1.ID)) {
		t.Fatalf("requests %v", s.seen)
	}
}

// A challenge on an answer is answered before the next request; one on the
// event stream at once; neither twice.
func TestChallengesAreAnswered(t *testing.T) {
	s := newTicketServer(t)
	c, _ := sessionClient(t, s, testSession())
	s.challenge = "BBBBBBBBBBBBBBBBBBBBBBBBBB"
	if _, err := c.Status(context.Background(), proto.StatusReport{}); err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(s.alive, []string{"BBBBBBBBBBBBBBBBBBBBBBBBBB"}) || !strings.HasPrefix(s.seen[len(s.seen)-1], proto.PathAlive) {
		t.Fatalf("alive %v after %v", s.alive, s.seen)
	}
	if err := c.Events(context.Background(), func(proto.ConfigEvent) {}, func() {}); err == nil {
		t.Fatal("a stream that ended is an error")
	}
	if !slices.Equal(s.alive, []string{"BBBBBBBBBBBBBBBBBBBBBBBBBB", testfix.Nonce}) {
		t.Fatalf("alive after the stream's challenge: %v", s.alive)
	}
	if err := c.Alive(context.Background(), testfix.Nonce); err != nil || len(s.alive) != 2 {
		t.Fatalf("a challenge answered twice: %v, %v", err, s.alive)
	}
}

// Goodbye releases the session when the server heard it; one it did not
// hear keeps it, so the next start carries on with it.
func TestBye(t *testing.T) {
	s := newTicketServer(t)
	c, kept := sessionClient(t, s, testSession())
	s.fail[proto.PathBye] = proto.CodeInternal
	if err := c.Bye(context.Background()); err == nil || !c.HasSession() {
		t.Fatalf("a failed goodbye = %v; session %v", err, c.HasSession())
	}
	if err := c.Bye(context.Background()); err != nil || c.HasSession() || !(*kept)[len(*kept)-1].Released {
		t.Fatalf("Bye = %v; kept %v", err, *kept)
	}
	if err := c.Bye(context.Background()); err != nil {
		t.Fatalf("Bye without a session = %v", err)
	}
	// A released session from the data dir waits for a hello.
	r, _ := sessionClient(t, s, &Session{ID: testfix.Session(1), Ticket: testfix.Ticket(1), Released: true})
	if r.HasSession() {
		t.Fatal("a released session was carried on with")
	}
}

// A hello answer that is not a session is refused.
func TestHelloAnswerIsChecked(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSONT(w, http.StatusOK, proto.HelloResponse{Session: "short", Ticket: testfix.Ticket(1), Heartbeat: 30})
	}))
	defer srv.Close()
	_, priv := testfix.IdentityKey(testfix.VPS1.Key)
	c, err := New(Options{Server: srv.URL, NodeID: testfix.VPS1.ID, Key: priv, Version: "v1.0.0"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c.Hello(context.Background(), proto.HelloRequest{}); err == nil || c.HasSession() {
		t.Fatalf("a malformed hello answer = %v", err)
	}
}

// A hello the server refuses is the refusal; Alive and Events wait for no
// session that is not there.
func TestWithoutASession(t *testing.T) {
	s := newTicketServer(t)
	c, _ := sessionClient(t, s, nil)
	if err := c.Alive(context.Background(), testfix.Nonce); !errors.Is(err, ErrNoSession) {
		t.Fatalf("Alive without a session = %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := c.Events(ctx, func(proto.ConfigEvent) {}, func() {}); !errors.Is(err, ErrNoSession) {
		t.Fatalf("Events without a session = %v", err)
	}
	refusing := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSONT(w, http.StatusConflict, proto.ErrorBody{Code: proto.CodeWait, Message: "another copy"})
	}))
	defer refusing.Close()
	_, priv := testfix.IdentityKey(testfix.VPS1.Key)
	w, err := New(Options{Server: refusing.URL, NodeID: testfix.VPS1.ID, Key: priv, Version: "v1.0.0"})
	if err != nil {
		t.Fatal(err)
	}
	var ce *Error
	if _, err := w.Hello(context.Background(), proto.HelloRequest{}); !errors.As(err, &ce) || ce.Code != proto.CodeWait || w.HasSession() {
		t.Fatalf("a hello told to wait = %v", err)
	}
}
