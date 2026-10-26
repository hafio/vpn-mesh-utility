package agent

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hafio/vpn-mesh-utility/internal/agent/client"
	"github.com/hafio/vpn-mesh-utility/internal/agent/state"
	"github.com/hafio/vpn-mesh-utility/internal/proto"
	"github.com/hafio/vpn-mesh-utility/internal/testfix"
)

// newToken is a second join token, as `meshctl node token` would issue.
var newToken = strings.Repeat("A", 43)

// startDir is a data dir that has joined when joined is set and remembers
// the token recorded ("" for no join.json).
func startDir(t *testing.T, joined bool, recorded string) (string, *state.Dir) {
	t.Helper()
	data := filepath.Join(t.TempDir(), "data")
	dir, err := state.Open(data)
	if err != nil {
		t.Fatal(err)
	}
	if joined {
		if err := dir.SaveEnrollment(state.Enrollment{Server: "https://mesh.example.com", NodeID: 4, Name: "nas", Role: proto.RoleSpoke,
			MeshIP: "10.77.1.1", ConfigKey: testfix.ConfigKeyB64()}); err != nil {
			t.Fatal(err)
		}
	}
	if recorded != "" {
		if err := dir.SaveJoinRecord(state.JoinRecord{Token: tokenFingerprint(recorded)}); err != nil {
			t.Fatal(err)
		}
	}
	return data, dir
}

var (
	spent       = &client.Error{Status: http.StatusUnauthorized, Code: proto.CodeTokenInvalid, Message: "token already used"}
	disabled    = &client.Error{Status: http.StatusForbidden, Code: proto.CodeNodeDisabled, Message: "node is disabled"}
	rateLimited = &client.Error{Status: http.StatusTooManyRequests, Code: proto.CodeRateLimited, Message: "wait a minute"}
	errNoAnswer = errors.New("dial tcp 203.0.113.1:443: connect: connection refused")
)

// Every start a container can see: first, restarted, a new token, no
// token, and each way a join can fail, on a node that has joined and on one
// that has not.
func TestJoinOnStart(t *testing.T) {
	for _, c := range []struct {
		name     string
		joined   bool
		recorded string
		token    string
		joinErr  error
		wantJoin bool
		wantErr  string
		// remembered is the token join.json holds afterwards ("" none).
		remembered string
		wantLog    string
	}{
		{name: "first start", token: testfix.Token, wantJoin: true, remembered: testfix.Token, wantLog: "msg=joined"},
		{name: "restart with the same token", joined: true, recorded: testfix.Token, token: testfix.Token, remembered: testfix.Token},
		{name: "a new token joins again", joined: true, recorded: testfix.Token, token: newToken, wantJoin: true, remembered: newToken,
			wantLog: "msg=joined"},
		{name: "no token after the join", joined: true, recorded: testfix.Token, remembered: testfix.Token},
		{name: "no token, never joined", wantErr: "has no join token"},
		{name: "spent new token keeps the enrollment", joined: true, recorded: testfix.Token, token: newToken, joinErr: spent, wantJoin: true,
			remembered: newToken, wantLog: "earlier enrollment"},
		{name: "spent first token", token: testfix.Token, joinErr: spent, wantJoin: true, wantErr: "new token from `meshctl node token",
			remembered: testfix.Token},
		{name: "spent on an earlier start", recorded: testfix.Token, token: testfix.Token, wantErr: "refused this join token on an earlier start",
			remembered: testfix.Token},
		{name: "joined by hand, old token left set", joined: true, token: testfix.Token, joinErr: spent, wantJoin: true, remembered: testfix.Token,
			wantLog: "earlier enrollment"},
		{name: "server unreachable after the join", joined: true, recorded: testfix.Token, token: newToken, joinErr: errNoAnswer, wantJoin: true,
			remembered: testfix.Token, wantLog: "connection refused"},
		{name: "server unreachable, never joined", token: testfix.Token, joinErr: errNoAnswer, wantJoin: true, wantErr: "the next start tries again"},
		{name: "rate limited", token: testfix.Token, joinErr: rateLimited, wantJoin: true, wantErr: "the next start tries again"},
		{name: "disabled node is asked again", token: testfix.Token, joinErr: disabled, wantJoin: true, wantErr: "node is disabled"},
	} {
		t.Run(c.name, func(t *testing.T) {
			data, dir := startDir(t, c.joined, c.recorded)
			var logs bytes.Buffer
			called := false
			join := func(_ context.Context, o JoinOptions) (proto.JoinResponse, error) {
				called = true
				if o.Token != c.token || !o.NoService || o.Service != nil || o.Out != io.Discard {
					t.Errorf("join options %+v", o)
				}
				if c.joinErr != nil {
					return proto.JoinResponse{}, c.joinErr
				}
				return proto.JoinResponse{NodeID: 4, Name: "nas", Role: proto.RoleSpoke, MeshIP: "10.77.1.1"}, nil
			}
			o := JoinOptions{DataDir: data, Server: "https://mesh.example.com", Token: c.token, Name: "nas",
				Log: slog.New(slog.NewTextHandler(&logs, nil))}
			startErr := joinOnStart(context.Background(), o, join)
			if c.wantErr == "" && startErr != nil || c.wantErr != "" && (startErr == nil || !strings.Contains(startErr.Error(), c.wantErr)) {
				t.Fatalf("err = %v, want %q", startErr, c.wantErr)
			}
			if called != c.wantJoin {
				t.Errorf("joined = %v, want %v", called, c.wantJoin)
			}
			rec, err := dir.JoinRecord()
			want := ""
			if c.remembered != "" {
				want = tokenFingerprint(c.remembered)
			}
			if err != nil || rec.Token != want {
				t.Errorf("join.json = %+v, %v; want the fingerprint of %q", rec, err, c.remembered)
			}
			if !strings.Contains(logs.String(), c.wantLog) {
				t.Errorf("log lacks %q:\n%s", c.wantLog, logs.String())
			}
			for _, tok := range []string{testfix.Token, newToken} {
				if strings.Contains(logs.String(), tok) || startErr != nil && strings.Contains(startErr.Error(), tok) {
					t.Errorf("the token reached the log or the error")
				}
			}
		})
	}
}

// What the data directory keeps is a fingerprint: not the token, and not
// the sha256 the server looks tokens up by.
func TestTokenFingerprint(t *testing.T) {
	f := tokenFingerprint(testfix.Token)
	if len(f) != 64 || strings.Contains(f, testfix.Token) || f == tokenFingerprint(newToken) || f != tokenFingerprint(testfix.Token) {
		t.Fatalf("fingerprint %q", f)
	}
	if server := sha256.Sum256([]byte(testfix.Token)); f == hex.EncodeToString(server[:]) {
		t.Fatal("the fingerprint is the server's lookup key (store.HashToken)")
	}
}

// A data directory that cannot say what it holds stops the start before
// the server is asked.
func TestJoinOnStartNeedsAReadableDataDir(t *testing.T) {
	never := func(context.Context, JoinOptions) (proto.JoinResponse, error) {
		t.Error("joined")
		return proto.JoinResponse{}, nil
	}
	data, _ := startDir(t, false, "")
	if err := os.WriteFile(filepath.Join(data, state.EnrollmentFile), []byte("{"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := joinOnStart(context.Background(), JoinOptions{DataDir: data, Token: testfix.Token}, never); err == nil {
		t.Error("a damaged enrollment was ignored")
	}
	data, _ = startDir(t, true, "")
	if err := os.WriteFile(filepath.Join(data, state.JoinFile), []byte("{"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := joinOnStart(context.Background(), JoinOptions{DataDir: data, Token: testfix.Token}, never); err == nil ||
		!strings.Contains(err.Error(), "remove it") {
		t.Errorf("a damaged join.json: %v", err)
	}
	if err := joinOnStart(context.Background(), JoinOptions{DataDir: "", Token: testfix.Token}, never); err == nil {
		t.Error("no data dir")
	}
}

// A join that worked but could not be remembered is an error (the next
// start would try the spent token); a spent token that could not be
// remembered only costs one more question next time.
func TestJoinOnStartCannotRemember(t *testing.T) {
	for _, joinErr := range []error{nil, spent} {
		data, _ := startDir(t, joinErr != nil, "")
		var logs bytes.Buffer
		join := func(context.Context, JoinOptions) (proto.JoinResponse, error) {
			if err := os.Mkdir(filepath.Join(data, state.JoinFile), 0o700); err != nil {
				t.Fatal(err)
			}
			return proto.JoinResponse{NodeID: 4}, joinErr
		}
		err := joinOnStart(context.Background(), JoinOptions{DataDir: data, Token: testfix.Token, Log: slog.New(slog.NewTextHandler(&logs, nil))}, join)
		switch {
		case joinErr == nil && (err == nil || !strings.Contains(err.Error(), "joined, but")):
			t.Errorf("joined: %v", err)
		case joinErr != nil && (err != nil || !strings.Contains(logs.String(), "could not remember")):
			t.Errorf("spent: %v\n%s", err, logs.String())
		}
	}
}

// Through the real client: the server's token_invalid is a spent token.
func TestJoinOnStartAsksTheServer(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = io.WriteString(w, `{"code":"token_invalid","message":"join token already used; issue a new one with meshctl node token <name>"}`)
	}))
	defer srv.Close()
	data, dir := startDir(t, false, "")
	o := JoinOptions{DataDir: data, Server: srv.URL, Token: testfix.Token, Version: "v1",
		Hostname: func() (string, error) { return "nas", nil }, Backend: func() string { return "kernel" },
		PublicIPv4: func() (netip.Addr, bool) { return netip.Addr{}, false }, PublicIPv6: func() (netip.Addr, bool) { return netip.Addr{}, false },
		MachineID: testMachineID}
	if err := JoinOnStart(context.Background(), o); err == nil || !strings.Contains(err.Error(), "already used") {
		t.Fatalf("err = %v", err)
	}
	if rec, _ := dir.JoinRecord(); rec.Token != tokenFingerprint(testfix.Token) {
		t.Fatal("the spent token was not remembered")
	}
}
