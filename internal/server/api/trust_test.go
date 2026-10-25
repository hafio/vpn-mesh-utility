package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/hafio/vpn-mesh-utility/internal/proto"
	"github.com/hafio/vpn-mesh-utility/internal/server/store"
	"github.com/hafio/vpn-mesh-utility/internal/testfix"
)

// signedAs is a 200 answer as the agent of f checks it (amendment A50).
func signedAs(t *testing.T, f testfix.Node, kind string, r *http.Request, w interface{ Header() http.Header }, body []byte) {
	t.Helper()
	b := proto.SignedBody{Body: body, Nonce: r.Header.Get(proto.HeaderNonce), Signature: w.Header().Get(proto.HeaderConfigSignature)}
	if err := testfix.Pinned(f.ID).VerifyBody(kind, b); err != nil {
		t.Fatalf("%s answer to %s: %v", kind, f.Name, err)
	}
}

// The control plane signs every config and mesh view for the node and the
// request it answers, gives its key with every join, and answers neither
// without a key (amendment A50).
func TestAnswersAreSigned(t *testing.T) {
	h := newHarness(t)
	h.seed(t, storeNode(testfix.VPS1, true), storeNode(testfix.VPS2, false))
	r := signed(http.MethodGet, proto.PathConfig, nil, testfix.VPS1, testfix.Epoch)
	w, eb := serve(h.pub, r)
	if w.Code != 200 {
		t.Fatalf("config: %d %s", w.Code, eb.Message)
	}
	signedAs(t, testfix.VPS1, proto.SignedConfig, r, w, w.Body.Bytes())
	r = signed(http.MethodGet, proto.PathMesh, nil, testfix.VPS1, testfix.Epoch)
	if w, eb = serve(h.pub, r); w.Code != 200 {
		t.Fatalf("mesh: %d %s", w.Code, eb.Message)
	}
	signedAs(t, testfix.VPS1, proto.SignedMesh, r, w, w.Body.Bytes())

	w, eb = postJoin(h, joinBody(t, h.issue(t, "vps-2"), testfix.VPS2, nil))
	var resp proto.JoinResponse
	if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &resp) != nil || resp.ConfigKey != testfix.ConfigKeyB64() {
		t.Fatalf("join: %d %+v %s", w.Code, resp, eb.Message)
	}

	keyless := New(Deps{DB: h.db, Hub: h.hub, Reconciler: h.rec, Now: h.clock.Now, ServerURL: testURL, Version: "v1.0.0"})
	for _, path := range []string{proto.PathConfig, proto.PathMesh} {
		if w, _ := serve(keyless, signed(http.MethodGet, path, nil, testfix.VPS1, testfix.Epoch)); w.Code != http.StatusInternalServerError {
			t.Errorf("%s without a config key: %d", path, w.Code)
		}
	}
	h.seed(t, storeNode(testfix.VPS3, false))
	body := joinBody(t, h.issue(t, "vps-3"), testfix.VPS3, nil)
	if w, _ := serve(keyless, httptest.NewRequest(http.MethodPost, proto.PathJoin, bytes.NewReader(body))); w.Code != http.StatusInternalServerError {
		t.Errorf("a join without a config key: %d", w.Code)
	}
}

// A new token cuts the node off at once (amendment A50): it is no longer
// enrolled, its peers drop it, its key joins its history, and its agent is
// told so, definitively and signed, on its next request. The join that
// follows needs fresh keys; one with the cut-off key is refused and keeps
// the token.
func TestTokenCutsTheNodeOff(t *testing.T) {
	h := newHarness(t)
	h.seedServers(t)
	// vps-2 reported just now, so the note says it was running.
	if err := h.st.Tx(context.Background(), func(tx *store.Tx) error {
		return tx.RecordStatus(testfix.VPS2.ID, store.StatusUpdate{AgentVersion: "v1.0.0", WGBackend: "kernel", StatusJSON: "{}"})
	}); err != nil {
		t.Fatal(err)
	}
	h.clock.Advance(time.Minute)
	ch, err := h.nodes().Token(context.Background(), cli(), "vps-2")
	if err != nil || ch.Enrollment == nil || len(ch.Notes) != 1 || !strings.Contains(ch.Notes[0], "was seen 1m0s ago and is cut off") {
		t.Fatalf("Token = %+v, %v", ch, err)
	}
	n := h.node(t, "vps-2")
	if n.Enrolled || n.IdentityPubKey != "" || n.WGPubKey != "" {
		t.Fatalf("vps-2 after the token = %+v", n)
	}
	for _, p := range h.config(t, "vps-1").Peers {
		if p.Name == "vps-2" {
			t.Fatal("vps-1 still peers the cut-off vps-2")
		}
	}
	if h.audits(t, "node.token") != 1 {
		t.Fatal("the token is not audited")
	}

	// The old agent's next request.
	r := signed(http.MethodPost, proto.PathStatus, []byte(`{"applied_version":1,"agent_version":"v1.0.0","wg_backend":"kernel","peers":[]}`),
		testfix.VPS2, h.clock.Now())
	w, eb := serve(h.pub, r)
	if w.Code != http.StatusForbidden || eb.Code != proto.CodeReplaced || eb.ServerTime == 0 ||
		testfix.Pinned(testfix.VPS2.ID).VerifyRefusal(eb.Code, r.Header.Get(proto.HeaderNonce), eb.ServerTime, w.Header().Get(proto.HeaderRefusalSignature)) != nil {
		t.Fatalf("the cut-off agent: %d %+v, signature %q", w.Code, eb, w.Header().Get(proto.HeaderRefusalSignature))
	}

	// A join with the cut-off key is refused and the token kept; with
	// fresh keys it enrolls, and the old key is still told it was replaced.
	tok := ch.Enrollment.Token
	if w, eb := postJoin(h, joinBody(t, tok, testfix.VPS2, nil)); w.Code != http.StatusConflict || !strings.Contains(eb.Message, "fresh keys") {
		t.Fatalf("a join with the cut-off key: %d %+v", w.Code, eb)
	}
	fresh := testfix.Node{ID: testfix.VPS2.ID, Name: "vps-2", Role: proto.RoleServer, Key: 60}
	if w, eb := postJoin(h, joinBody(t, tok, fresh, nil)); w.Code != 200 {
		t.Fatalf("a join with fresh keys: %d %+v", w.Code, eb)
	}
	if n := h.node(t, "vps-2"); !n.Enrolled || n.IdentityPubKey != testfix.IdentityKeyB64(60) {
		t.Fatalf("vps-2 after the join = %+v", n)
	}
	r = signed(http.MethodGet, proto.PathConfig, nil, testfix.VPS2, h.clock.Now())
	if w, eb := serve(h.pub, r); w.Code != http.StatusForbidden || eb.Code != proto.CodeReplaced {
		t.Fatalf("the old key after the new join: %d %+v", w.Code, eb)
	}
	r = signed(http.MethodGet, proto.PathConfig, nil, fresh, h.clock.Now())
	if w, eb := serve(h.pub, r); w.Code != 200 {
		t.Fatalf("the new key: %d %+v", w.Code, eb)
	}

	// A node that never joined has nothing to cut off.
	h.seed(t, storeNode(testfix.NAS, false))
	if ch, err := h.nodes().Token(context.Background(), cli(), "nas"); err != nil || len(ch.Notes) != 0 {
		t.Fatalf("Token for a node that never joined = %+v, %v", ch, err)
	}
}
