package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/hafio/vpn-mesh-utility/internal/proto"
	"github.com/hafio/vpn-mesh-utility/internal/server/store"
	"github.com/hafio/vpn-mesh-utility/internal/testfix"
)

// as is a request of node f presenting session and ticket ("" for none),
// at the harness's time.
func (h *harness) as(f testfix.Node, session, ticket, method, path string, body []byte) (*httptest.ResponseRecorder, proto.ErrorBody, *http.Request) {
	r := httptest.NewRequest(method, path, bytes.NewReader(body))
	_, priv := testfix.IdentityKey(f.Key)
	proto.SignSession(priv, f.ID, method, r.URL.RequestURI(), h.clock.Now(), session, ticket, body).Apply(r.Header)
	w, eb := serve(h.pub, r)
	return w, eb, r
}

const report = `{"applied_version":1,"agent_version":"v1.0.0","wg_backend":"kernel","peers":[]}`

// hello says hello as f from machine.
func (h *harness) hello(f testfix.Node, machine, host string) (*httptest.ResponseRecorder, proto.ErrorBody, *http.Request) {
	body, _ := json.Marshal(proto.HelloRequest{MachineID: machine, Hostname: host})
	return h.as(f, "", "", http.MethodPost, proto.PathHello, body)
}

// Every request presents the ticket its previous answer carried and gets
// the next, whatever the handler answers; the previous one is taken once
// more after a lost answer; anything else, or no session, is 409 session
// (amendment A50).
func TestTickets(t *testing.T) {
	h := newHarness(t)
	h.seedServers(t)
	s := testfix.Session(testfix.VPS1.ID)
	step := func(ticket, body string) (string, *httptest.ResponseRecorder, proto.ErrorBody) {
		t.Helper()
		w, eb, _ := h.as(testfix.VPS1, s, ticket, http.MethodPost, proto.PathStatus, []byte(body))
		return w.Header().Get(proto.HeaderNextTicket), w, eb
	}
	n1, w, eb := step(testfix.Ticket(testfix.VPS1.ID), report)
	if w.Code != 200 || !proto.ValidSecret(n1) {
		t.Fatalf("the first ticket: %d %+v, next %q", w.Code, eb, n1)
	}
	n2, w, _ := step(n1, "not json") // a refusal from the handler still moves the ticket on
	if w.Code != http.StatusBadRequest || !proto.ValidSecret(n2) || n2 == n1 {
		t.Fatalf("a bad report: %d, next %q", w.Code, n2)
	}
	n2b, w, _ := step(n1, report) // n2's answer was lost: n1 again
	if w.Code != 200 || !proto.ValidSecret(n2b) || n2b == n2 {
		t.Fatalf("a retry after a lost answer: %d, next %q", w.Code, n2b)
	}
	for name, ticket := range map[string]string{"the replaced ticket": n2, "the first ticket again": testfix.Ticket(testfix.VPS1.ID), "none": ""} {
		if next, w, eb := step(ticket, report); w.Code != http.StatusConflict || eb.Code != proto.CodeSession || next != "" {
			t.Errorf("%s: %d %+v, next %q", name, w.Code, eb, next)
		}
	}
	if w, eb, _ := h.as(testfix.VPS1, testfix.Session(9), n2b, http.MethodPost, proto.PathStatus, []byte(report)); eb.Code != proto.CodeSession {
		t.Fatalf("another session: %d %+v", w.Code, eb)
	}
	// The event stream takes the current or the previous ticket and moves
	// neither.
	r := httptest.NewRequest(http.MethodGet, proto.PathEvents, nil)
	_, priv := testfix.IdentityKey(testfix.VPS1.Key)
	proto.SignSession(priv, testfix.VPS1.ID, http.MethodGet, r.URL.RequestURI(), h.clock.Now(), s, "", nil).Apply(r.Header)
	if w, eb := serve(h.pub, r); eb.Code != proto.CodeSession {
		t.Fatalf("a stream without a ticket: %d %+v", w.Code, eb)
	}
	if next, w, _ := step(n2b, report); w.Code != 200 || next == "" {
		t.Fatalf("the ticket after the stream: %d", w.Code)
	}
}

// A hello with no live holder is granted: a session, its first ticket and
// the heartbeat, which the requests that follow use; the machine is pinned
// and the hello audited.
func TestHelloIsGranted(t *testing.T) {
	h := newHarness(t)
	h.seedServers(t)
	h.forget(t, testfix.VPS1.ID)
	w, eb, _ := h.hello(testfix.VPS1, testfix.MachineID, "vps-1.example.com")
	var resp proto.HelloResponse
	if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &resp) != nil || resp.Validate() != nil || resp.Heartbeat != proto.Heartbeat {
		t.Fatalf("hello: %d %+v %s", w.Code, eb, w.Body.String())
	}
	if w, eb, _ := h.as(testfix.VPS1, resp.Session, resp.Ticket, http.MethodPost, proto.PathStatus, []byte(report)); w.Code != 200 {
		t.Fatalf("a request in the new session: %d %+v", w.Code, eb)
	}
	if h.audits(t, "node.session") != 1 {
		t.Fatal("the hello is not audited")
	}
	for name, body := range map[string]string{"no machine": `{"hostname":"vps-1"}`, "junk": `{"machine_id":1}`} {
		if w, eb, _ := h.as(testfix.VPS1, "", "", http.MethodPost, proto.PathHello, []byte(body)); w.Code != http.StatusBadRequest {
			t.Errorf("%s: %d %+v", name, w.Code, eb)
		}
	}
}

// forget removes node's session, as a node that never said hello has none.
func (h *harness) forget(t *testing.T, node int64) {
	t.Helper()
	if err := h.st.Tx(context.Background(), func(tx *store.Tx) error { return tx.DeleteAgentSession(node) }); err != nil {
		t.Fatal(err)
	}
}

// A second copy that says hello while the holder is alive waits; the holder
// is challenged on its stream and on its next answer, proves it is alive,
// and the copy's next hello is a signed duplicate, recorded and audited.
func TestDuplicateIsRefused(t *testing.T) {
	h := newHarness(t)
	h.seedServers(t)
	_, _, challenges, _, cancel := h.hub.Subscribe(testfix.VPS1.ID, store.FirstMesh)
	defer cancel()
	w, eb, _ := h.hello(testfix.VPS1, testfix.MachineID, "copy.example.com")
	if w.Code != http.StatusConflict || eb.Code != proto.CodeWait || w.Header().Get("Retry-After") != "5" ||
		!strings.Contains(eb.Message, "waits") || w.Header().Get(proto.HeaderRefusalSignature) != "" {
		t.Fatalf("the copy's hello: %d %+v", w.Code, eb)
	}
	var nonce string
	select {
	case nonce = <-challenges:
	default:
		t.Fatal("the holder's stream got no challenge")
	}
	if info := h.show(t, "vps-1"); info.Copy == nil || !info.Copy.Waiting || info.Copy.Hostname != "copy.example.com" ||
		!strings.Contains(CopyLine(*info.Copy, time.UTC), "has waited since") {
		t.Fatalf("node show while the copy waits: %+v", info.Copy)
	}
	// Asking again changes nothing.
	if _, eb, _ := h.hello(testfix.VPS1, testfix.MachineID, "copy.example.com"); eb.Code != proto.CodeWait {
		t.Fatalf("the copy's second hello: %+v", eb)
	}
	if again := <-challenges; again != nonce {
		t.Fatalf("a second challenge %q, want the first %q", again, nonce)
	}
	s, tk := testfix.Session(testfix.VPS1.ID), testfix.Ticket(testfix.VPS1.ID)
	w, _, _ = h.as(testfix.VPS1, s, tk, http.MethodPost, proto.PathStatus, []byte(report))
	if w.Header().Get(proto.HeaderChallenge) != nonce {
		t.Fatalf("the holder's answer carries challenge %q, want %q", w.Header().Get(proto.HeaderChallenge), nonce)
	}
	next := w.Header().Get(proto.HeaderNextTicket)
	// A wrong nonce is ignored.
	body, _ := json.Marshal(proto.AliveRequest{Nonce: "CCCCCCCCCCCCCCCCCCCCCCCCCC"})
	w, _, _ = h.as(testfix.VPS1, s, next, http.MethodPost, proto.PathAlive, body)
	if w.Code != http.StatusNoContent || w.Header().Get(proto.HeaderChallenge) != nonce {
		t.Fatalf("a wrong nonce: %d, challenge %q", w.Code, w.Header().Get(proto.HeaderChallenge))
	}
	next = w.Header().Get(proto.HeaderNextTicket)
	body, _ = json.Marshal(proto.AliveRequest{Nonce: nonce})
	if w, eb, _ := h.as(testfix.VPS1, s, next, http.MethodPost, proto.PathAlive, body); w.Code != http.StatusNoContent ||
		w.Header().Get(proto.HeaderChallenge) != "" {
		t.Fatalf("alive: %d %+v", w.Code, eb)
	}
	w, eb, r := h.hello(testfix.VPS1, testfix.MachineID, "copy.example.com")
	if w.Code != http.StatusForbidden || eb.Code != proto.CodeDuplicate ||
		testfix.Pinned(testfix.VPS1.ID).VerifyRefusal(eb.Code, r.Header.Get(proto.HeaderNonce), eb.ServerTime, w.Header().Get(proto.HeaderRefusalSignature)) != nil {
		t.Fatalf("the copy after the holder answered: %d %+v", w.Code, eb)
	}
	if h.audits(t, "node.duplicate") != 1 {
		t.Fatal("the duplicate is not audited")
	}
	info := h.show(t, "vps-1")
	if c := info.Copy; c == nil || c.Waiting || c.Code != proto.CodeDuplicate || !c.SameMachine || c.Hostname != "copy.example.com" {
		t.Fatalf("node show after the refusal: %+v", c)
	}
	if _, label := Health(info, h.clock.Now()); !strings.Contains(label, "was refused") {
		t.Fatalf("the dashboard says %q", label)
	}
}

// show is node name as the admin API shows it.
func (h *harness) show(t *testing.T, name string) NodeInfo {
	t.Helper()
	d, err := h.nodes().Show(context.Background(), name)
	if err != nil {
		t.Fatal(err)
	}
	return d.Node
}

// A holder that does not answer within the window loses the session to
// the waiting copy; when it comes back it is sent to hello. A holder that
// said goodbye, or went silent, leaves it at once.
func TestSessionLapses(t *testing.T) {
	h := newHarness(t)
	h.seedServers(t)
	if _, eb, _ := h.hello(testfix.VPS1, testfix.MachineID, "copy"); eb.Code != proto.CodeWait {
		t.Fatalf("the copy's hello: %+v", eb)
	}
	h.clock.Advance(proto.SessionWindow * time.Second)
	w, eb, _ := h.hello(testfix.VPS1, testfix.MachineID, "copy")
	var resp proto.HelloResponse
	if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &resp) != nil {
		t.Fatalf("the copy after the window: %d %+v", w.Code, eb)
	}
	s, tk := testfix.Session(testfix.VPS1.ID), testfix.Ticket(testfix.VPS1.ID)
	if _, eb, _ := h.as(testfix.VPS1, s, tk, http.MethodPost, proto.PathStatus, []byte(report)); eb.Code != proto.CodeSession {
		t.Fatalf("the old holder after the lapse: %+v", eb)
	}
	// Goodbye: the next hello is granted at once.
	w, _, _ = h.as(testfix.VPS1, resp.Session, resp.Ticket, http.MethodPost, proto.PathBye, nil)
	if w.Code != http.StatusNoContent {
		t.Fatalf("bye: %d", w.Code)
	}
	if w, eb, _ := h.hello(testfix.VPS1, testfix.MachineID, "vps-1"); w.Code != 200 {
		t.Fatalf("a hello after goodbye: %d %+v", w.Code, eb)
	}
	// A holder silent for the window: granted at once too.
	h.clock.Advance(proto.SessionWindow * time.Second)
	if w, eb, _ := h.hello(testfix.VPS1, testfix.MachineID, "vps-1"); w.Code != 200 {
		t.Fatalf("a hello after a silent holder: %d %+v", w.Code, eb)
	}
	if n := h.audits(t, "node.session"); n != 3 {
		t.Fatalf("%d sessions audited", n)
	}
}

// An answered challenge that is over (older than the window) does not
// refuse the next copy: the holder is asked again.
func TestStaleAnswerAsksAgain(t *testing.T) {
	h := newHarness(t)
	h.seedServers(t)
	h.hello(testfix.VPS1, testfix.MachineID, "copy")
	s, tk := testfix.Session(testfix.VPS1.ID), testfix.Ticket(testfix.VPS1.ID)
	w, _, _ := h.as(testfix.VPS1, s, tk, http.MethodPost, proto.PathStatus, []byte(report))
	nonce := w.Header().Get(proto.HeaderChallenge)
	body, _ := json.Marshal(proto.AliveRequest{Nonce: nonce})
	h.as(testfix.VPS1, s, tk, http.MethodPost, proto.PathAlive, body)
	h.clock.Advance(proto.SessionWindow * time.Second)
	h.as(testfix.VPS1, s, tk, http.MethodPost, proto.PathStatus, []byte(report)) // the holder is alive
	if _, eb, _ := h.hello(testfix.VPS1, testfix.MachineID, "copy"); eb.Code != proto.CodeWait {
		t.Fatalf("a copy after a stale answer: %+v", eb)
	}
}

// A hello from another machine than the node's is a signed foreign,
// recorded and audited; join pins the machine and ends the session, and so
// does a token.
func TestForeignAndPinning(t *testing.T) {
	h := newHarness(t)
	h.seed(t, storeNode(testfix.VPS1, true), storeNode(testfix.VPS2, false))
	w, eb, r := h.hello(testfix.VPS1, testfix.OtherMachine, "elsewhere")
	if w.Code != http.StatusForbidden || eb.Code != proto.CodeForeign || !strings.Contains(eb.Message, "node token") ||
		testfix.Pinned(testfix.VPS1.ID).VerifyRefusal(eb.Code, r.Header.Get(proto.HeaderNonce), eb.ServerTime, w.Header().Get(proto.HeaderRefusalSignature)) != nil {
		t.Fatalf("a hello from another machine: %d %+v", w.Code, eb)
	}
	if c := h.show(t, "vps-1").Copy; c == nil || c.Code != proto.CodeForeign || c.SameMachine || h.audits(t, "node.foreign") != 1 {
		t.Fatalf("the foreign copy: %+v", c)
	}
	if !strings.Contains(CopyLine(*h.show(t, "vps-1").Copy, time.UTC), "another machine (elsewhere)") {
		t.Fatal(CopyLine(*h.show(t, "vps-1").Copy, time.UTC))
	}
	// A join pins the machine it came from and leaves no session.
	tok := h.issue(t, "vps-2")
	body := joinBody(t, tok, testfix.VPS2, func(r *proto.JoinRequest) { r.MachineID = testfix.OtherMachine })
	if w, eb := serve(h.pub, httptest.NewRequest(http.MethodPost, proto.PathJoin, bytes.NewReader(body))); w.Code != 200 {
		t.Fatalf("join: %d %+v", w.Code, eb)
	}
	if err := h.st.Tx(context.Background(), func(tx *store.Tx) error {
		m, err := tx.Machine(testfix.VPS2.ID)
		if err != nil || m.Hash != secretHash(testfix.OtherMachine) {
			t.Errorf("the joined machine %+v, %v", m, err)
		}
		if _, err := tx.AgentSession(testfix.VPS2.ID); !errors.Is(err, store.ErrNotFound) {
			t.Errorf("a join left a session: %v", err)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if w, eb, _ := h.hello(testfix.VPS2, testfix.MachineID, "vps-2"); eb.Code != proto.CodeForeign {
		t.Fatalf("vps-2 from the first machine: %d %+v", w.Code, eb)
	}
	if w, _, _ := h.hello(testfix.VPS2, testfix.OtherMachine, "vps-2"); w.Code != 200 {
		t.Fatalf("vps-2 from its own machine: %d", w.Code)
	}
	if _, err := h.nodes().Token(context.Background(), cli(), "vps-1"); err != nil {
		t.Fatal(err)
	}
	if err := h.st.Tx(context.Background(), func(tx *store.Tx) error {
		if _, err := tx.AgentSession(testfix.VPS1.ID); !errors.Is(err, store.ErrNotFound) {
			t.Errorf("a token left the session: %v", err)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}

// The challenge's nonce is the same at every delivery, and another
// session's, or another challenge's, differs.
func TestChallengeNonce(t *testing.T) {
	a := newAPI(Deps{ConfigKey: testfix.ConfigKey()})
	s := store.AgentSession{SessionHash: "s", ChallengeAt: 1}
	n := a.challengeNonce(1, s)
	if !proto.ValidChallenge(n) || a.challengeNonce(1, s) != n {
		t.Fatalf("nonce %q", n)
	}
	other := s
	other.ChallengeAt = 2
	if a.challengeNonce(1, other) == n || a.challengeNonce(2, s) == n {
		t.Fatal("nonces collide")
	}
}

// The event stream opens with a challenge that is already open, and
// carries one a later hello sends again.
func TestStreamCarriesChallenges(t *testing.T) {
	h := newHarness(t)
	h.seedServers(t)
	if _, eb, _ := h.hello(testfix.VPS1, testfix.MachineID, "copy"); eb.Code != proto.CodeWait {
		t.Fatalf("the copy's hello: %+v", eb)
	}
	srv := httptest.NewServer(h.pub)
	defer srv.Close()
	r, err := http.NewRequest(http.MethodGet, srv.URL+proto.PathEvents, nil)
	if err != nil {
		t.Fatal(err)
	}
	_, priv := testfix.IdentityKey(testfix.VPS1.Key)
	proto.SignSession(priv, testfix.VPS1.ID, http.MethodGet, r.URL.RequestURI(), h.clock.Now(), testfix.Session(testfix.VPS1.ID),
		testfix.Ticket(testfix.VPS1.ID), nil).Apply(r.Header)
	resp, err := srv.Client().Do(r)
	if err != nil || resp.StatusCode != 200 {
		t.Fatalf("the stream: %v %v", resp, err)
	}
	defer resp.Body.Close()
	rd := proto.NewReader(resp.Body)
	next := func(name string) string {
		t.Helper()
		for {
			ev, err := rd.Next()
			if err != nil {
				t.Fatalf("waiting for %s: %v", name, err)
			}
			if ev.Name == name {
				return string(ev.Data)
			}
		}
	}
	first := next(proto.EventChallenge)
	h.hello(testfix.VPS1, testfix.MachineID, "copy") // asks again
	if again := next(proto.EventChallenge); again != first || !strings.Contains(first, `"nonce"`) {
		t.Fatalf("challenges %s, then %s", first, again)
	}
}

// Requests without a session, and the session calls' own failures: the
// signature, the body, the store -- each answered without touching the
// session, the store's cause only in the log.
func TestSessionCallFailures(t *testing.T) {
	h := newHarness(t)
	h.seedServers(t)
	h.forget(t, testfix.VPS2.ID)
	if _, eb, _ := h.as(testfix.VPS2, testfix.Session(testfix.VPS2.ID), testfix.Ticket(testfix.VPS2.ID), http.MethodPost, proto.PathStatus,
		[]byte(report)); eb.Code != proto.CodeSession {
		t.Fatalf("a request without a session: %+v", eb)
	}
	r := httptest.NewRequest(http.MethodGet, proto.PathEvents, nil)
	_, priv := testfix.IdentityKey(testfix.VPS2.Key)
	proto.SignSession(priv, testfix.VPS2.ID, http.MethodGet, r.URL.RequestURI(), h.clock.Now(), testfix.Session(9), testfix.Ticket(9), nil).Apply(r.Header)
	if _, eb := serve(h.pub, r); eb.Code != proto.CodeSession {
		t.Fatalf("a stream without a session: %+v", eb)
	}
	r = httptest.NewRequest(http.MethodGet, proto.PathEvents, nil)
	_, priv = testfix.IdentityKey(testfix.VPS1.Key)
	proto.SignSession(priv, testfix.VPS1.ID, http.MethodGet, r.URL.RequestURI(), h.clock.Now(), testfix.Session(9), testfix.Ticket(9), nil).Apply(r.Header)
	if _, eb := serve(h.pub, r); eb.Code != proto.CodeSession {
		t.Fatalf("a stream with another session: %+v", eb)
	}
	for _, path := range []string{proto.PathHello, proto.PathAlive, proto.PathBye} {
		if w, _ := serve(h.pub, httptest.NewRequest(http.MethodPost, path, nil)); w.Code != http.StatusUnauthorized {
			t.Errorf("an unsigned %s: %d", path, w.Code)
		}
	}
	s, tk := testfix.Session(testfix.VPS1.ID), testfix.Ticket(testfix.VPS1.ID)
	for body, want := range map[string]int{`{"nonce":1}`: http.StatusBadRequest, `{"nonce":"short"}`: http.StatusBadRequest} {
		if w, _, _ := h.as(testfix.VPS1, s, tk, http.MethodPost, proto.PathAlive, []byte(body)); w.Code != want {
			t.Errorf("alive %s: %d", body, w.Code)
		}
	}
	keyless := New(Deps{DB: h.db, Hub: h.hub, Reconciler: h.rec, Now: h.clock.Now, ServerURL: testURL, Version: "v1.0.0"})
	body, _ := json.Marshal(proto.HelloRequest{MachineID: testfix.MachineID, Hostname: "vps-3"})
	r = httptest.NewRequest(http.MethodPost, proto.PathHello, bytes.NewReader(body))
	_, priv = testfix.IdentityKey(testfix.VPS3.Key)
	proto.Sign(priv, testfix.VPS3.ID, http.MethodPost, r.URL.RequestURI(), h.clock.Now(), body).Apply(r.Header)
	if w, _ := serve(keyless, r); w.Code != http.StatusInternalServerError {
		t.Fatalf("a hello without a config key: %d", w.Code)
	}
	alive, _ := json.Marshal(proto.AliveRequest{Nonce: testfix.Nonce})
	hello3 := func() *httptest.ResponseRecorder {
		w, _, _ := h.hello(testfix.VPS3, testfix.MachineID, "vps-3")
		return w
	}
	post := func(path string, body []byte) func() *httptest.ResponseRecorder {
		return func() *httptest.ResponseRecorder {
			w, _, _ := h.as(testfix.VPS1, s, tk, http.MethodPost, path, body)
			return w
		}
	}
	stream := func() *httptest.ResponseRecorder {
		r := httptest.NewRequest(http.MethodGet, proto.PathEvents, nil)
		_, priv := testfix.IdentityKey(testfix.VPS1.Key)
		proto.SignSession(priv, testfix.VPS1.ID, http.MethodGet, r.URL.RequestURI(), h.clock.Now(), s, tk, nil).Apply(r.Header)
		w, _ := serve(h.pub, r)
		return w
	}
	for _, c := range []struct {
		name string
		call func() *httptest.ResponseRecorder
		tx   int64
	}{{"hello", hello3, 2}, {"alive", post(proto.PathAlive, alive), 3}, {"bye", post(proto.PathBye, nil), 3}, {"events", stream, 3}} {
		h.db.n.Store(0)
		h.db.fail.Store(c.tx)
		if w := c.call(); w.Code != http.StatusInternalServerError || strings.Contains(w.Body.String(), "disk") {
			t.Errorf("%s with the store failing: %d %s", c.name, w.Code, w.Body.String())
		}
	}
	h.db.fail.Store(0)
	// An audit that fails does not change the answer.
	if _, err := h.nodes().Token(context.Background(), cli(), "vps-2"); err != nil {
		t.Fatal(err)
	}
	h.db.n.Store(0)
	h.db.fail.Store(2)
	if _, eb, _ := h.as(testfix.VPS2, "", "", http.MethodPost, proto.PathStatus, []byte(report)); eb.Code != proto.CodeReplaced {
		t.Fatalf("a replaced key whose audit fails: %+v", eb)
	}
	h.db.fail.Store(0)
}

// A holder that keeps sending but does not answer loses the session when
// the window from the challenge is over (a lapse, audited as one), with a
// window of the server's own.
func TestUnansweredHolderLapses(t *testing.T) {
	h := newHarness(t)
	h.seedServers(t)
	a := newAPI(Deps{DB: h.db, Hub: h.hub, Reconciler: h.rec, Now: h.clock.Now, ServerURL: testURL, ConfigKey: testfix.ConfigKey(),
		SessionWindow: 30 * time.Second})
	if a.window() != 30 {
		t.Fatalf("window %d", a.window())
	}
	h.hello(testfix.VPS1, testfix.MachineID, "copy")
	s, tk := testfix.Session(testfix.VPS1.ID), testfix.Ticket(testfix.VPS1.ID)
	h.clock.Advance(60 * time.Second)
	h.as(testfix.VPS1, s, tk, http.MethodPost, proto.PathStatus, []byte(report)) // alive, but no answer
	h.clock.Advance(30 * time.Second)
	if w, eb, _ := h.hello(testfix.VPS1, testfix.MachineID, "copy"); w.Code != 200 {
		t.Fatalf("the copy after an unanswered window: %d %+v", w.Code, eb)
	}
	var lapsed bool
	if err := h.st.MeshTx(context.Background(), store.FirstMesh, func(tx *store.MeshTx) error {
		log, err := tx.AuditLog(10, 0)
		for _, e := range log {
			lapsed = lapsed || e.Action == "node.session" && strings.Contains(e.Detail, `"lapsed":true`)
		}
		return err
	}); err != nil || !lapsed {
		t.Fatalf("no lapse audited: %v", err)
	}
}
