package proto_test

import (
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/hafio/vpn-mesh-utility/internal/proto"
	"github.com/hafio/vpn-mesh-utility/internal/testfix"
)

func TestSignVerifyRoundTrip(t *testing.T) {
	pub, priv := testfix.IdentityKey(1)
	body := []byte(`{"applied_version":3}`)
	s := proto.Sign(priv, 7, http.MethodPost, "/api/v1/agent/status", testfix.Epoch, body)

	h := http.Header{}
	s.Apply(h)
	parsed, err := proto.ParseSigned(h)
	if err != nil {
		t.Fatalf("ParseSigned: %v", err)
	}
	if parsed.NodeID != 7 || parsed.Timestamp != testfix.Epoch.Unix() {
		t.Fatalf("parsed = %+v, want node 7 at %d", parsed, testfix.Epoch.Unix())
	}
	if err := parsed.Verify(pub, http.MethodPost, "/api/v1/agent/status", body, testfix.Epoch); err != nil {
		t.Fatalf("Verify: %v", err)
	}
}

// Every element the canonical string covers must break the signature when
// tampered with -- including the query string (amendment A8) and the nonce.
func TestVerifyRejectsTampering(t *testing.T) {
	pub, priv := testfix.IdentityKey(1)
	otherPub, _ := testfix.IdentityKey(2)
	const uri = "/api/v1/agent/config?since=3"
	body := []byte(`{"x":1}`)
	s := proto.SignSession(priv, 7, http.MethodGet, uri, testfix.Epoch, testfix.Session(7), testfix.Ticket(7), body)

	cases := []struct {
		name   string
		pub    []byte
		method string
		uri    string
		body   []byte
		edit   func(*proto.Signed)
	}{
		{"method", pub, http.MethodPost, uri, body, nil},
		{"path", pub, http.MethodGet, "/api/v1/agent/status?since=3", body, nil},
		{"query", pub, http.MethodGet, "/api/v1/agent/config?since=4", body, nil},
		{"body", pub, http.MethodGet, uri, []byte(`{"x":2}`), nil},
		{"timestamp", pub, http.MethodGet, uri, body, func(s *proto.Signed) { s.Timestamp++ }},
		{"nonce", pub, http.MethodGet, uri, body, func(s *proto.Signed) { s.Nonce = "AAAAAAAAAAAAAAAAAAAAAAAAAA" }},
		{"session", pub, http.MethodGet, uri, body, func(s *proto.Signed) { s.Session = testfix.Session(8) }},
		{"ticket", pub, http.MethodGet, uri, body, func(s *proto.Signed) { s.Ticket = testfix.Ticket(8) }},
		{"no ticket", pub, http.MethodGet, uri, body, func(s *proto.Signed) { s.Ticket = "" }},
		{"other node key", otherPub, http.MethodGet, uri, body, nil},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			tampered := s
			if c.edit != nil {
				c.edit(&tampered)
			}
			err := tampered.Verify(c.pub, c.method, c.uri, c.body, testfix.Epoch)
			if !errors.Is(err, proto.ErrBadSignature) {
				t.Fatalf("Verify after tampering with %s = %v, want ErrBadSignature", c.name, err)
			}
		})
	}
}

// Two identical requests in the same second must still differ, or the
// server's replay cache would refuse the second one.
func TestSignIsUniquePerRequest(t *testing.T) {
	_, priv := testfix.IdentityKey(1)
	a := proto.Sign(priv, 1, http.MethodGet, proto.PathConfig, testfix.Epoch, nil)
	b := proto.Sign(priv, 1, http.MethodGet, proto.PathConfig, testfix.Epoch, nil)
	if a.Nonce == b.Nonce || string(a.Signature) == string(b.Signature) {
		t.Fatalf("identical requests produced the same nonce or signature: %q", a.Nonce)
	}
}

func TestVerifySkewBoundaries(t *testing.T) {
	pub, priv := testfix.IdentityKey(1)
	sign := func(at time.Time) proto.Signed {
		return proto.Sign(priv, 1, http.MethodGet, proto.PathConfig, at, nil)
	}
	cases := []struct {
		name   string
		offset time.Duration
		ok     bool
	}{
		{"exact", 0, true},
		{"60s behind", -proto.MaxSkew, true},
		{"60s ahead", proto.MaxSkew, true},
		{"61s behind", -proto.MaxSkew - time.Second, false},
		{"61s ahead", proto.MaxSkew + time.Second, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := sign(testfix.Epoch.Add(c.offset)).Verify(pub, http.MethodGet, proto.PathConfig, nil, testfix.Epoch)
			if c.ok && err != nil {
				t.Fatalf("Verify = %v, want nil", err)
			}
			if !c.ok {
				if !errors.Is(err, proto.ErrSkew) {
					t.Fatalf("Verify = %v, want ErrSkew", err)
				}
				if !strings.Contains(err.Error(), "NTP") {
					t.Fatalf("skew error must say what to do next: %v", err)
				}
			}
		})
	}
}

func TestCheckClockNeedsNoKey(t *testing.T) {
	_, priv := testfix.IdentityKey(1)
	s := proto.Sign(priv, 1, http.MethodGet, proto.PathConfig, testfix.Epoch, nil)
	if err := s.CheckClock(testfix.Epoch.Add(proto.MaxSkew)); err != nil {
		t.Fatalf("CheckClock at the limit = %v", err)
	}
	if err := s.CheckClock(testfix.Epoch.Add(-proto.MaxSkew - time.Second)); !errors.Is(err, proto.ErrSkew) || !strings.Contains(err.Error(), "ahead of") {
		t.Fatalf("CheckClock with the node ahead = %v, want ErrSkew naming the direction", err)
	}
}

func TestVerifyRejectsMissingKey(t *testing.T) {
	_, priv := testfix.IdentityKey(1)
	s := proto.Sign(priv, 1, http.MethodGet, proto.PathConfig, testfix.Epoch, nil)
	if err := s.Verify(nil, http.MethodGet, proto.PathConfig, nil, testfix.Epoch); !errors.Is(err, proto.ErrBadSignature) {
		t.Fatalf("Verify with no key = %v, want ErrBadSignature", err)
	}
}

func TestParseSignedRejectsMalformedHeaders(t *testing.T) {
	_, priv := testfix.IdentityKey(1)
	good := http.Header{}
	proto.Sign(priv, 1, http.MethodGet, proto.PathConfig, testfix.Epoch, nil).Apply(good)

	cases := []struct {
		name, header, value string
	}{
		{"missing node", proto.HeaderNode, ""},
		{"zero node", proto.HeaderNode, "0"},
		{"negative node", proto.HeaderNode, "-4"},
		{"text node", proto.HeaderNode, "vps-1"},
		{"missing timestamp", proto.HeaderTimestamp, ""},
		{"text timestamp", proto.HeaderTimestamp, "yesterday"},
		{"missing nonce", proto.HeaderNonce, ""},
		{"short nonce", proto.HeaderNonce, "ABC"},
		{"lowercase nonce", proto.HeaderNonce, "aaaaaaaaaaaaaaaaaaaaaaaaaa"},
		{"missing signature", proto.HeaderSignature, ""},
		{"short signature", proto.HeaderSignature, "AAAA"},
		{"non-base64 signature", proto.HeaderSignature, "!!!"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			h := good.Clone()
			h.Set(c.header, c.value)
			if _, err := proto.ParseSigned(h); !errors.Is(err, proto.ErrUnsigned) {
				t.Fatalf("ParseSigned = %v, want ErrUnsigned", err)
			}
		})
	}
}

// The session and the ticket travel in their own headers (amendment A50),
// parse back as sent, and are refused when malformed; a request without
// them has neither header.
func TestSessionHeaders(t *testing.T) {
	_, priv := testfix.IdentityKey(1)
	h := http.Header{}
	proto.SignSession(priv, 1, http.MethodGet, proto.PathConfig, testfix.Epoch, testfix.Session(1), testfix.Ticket(1), nil).Apply(h)
	s, err := proto.ParseSigned(h)
	if err != nil || s.Session != testfix.Session(1) || s.Ticket != testfix.Ticket(1) {
		t.Fatalf("ParseSigned = %+v, %v", s, err)
	}
	for _, name := range []string{proto.HeaderSession, proto.HeaderTicket} {
		bad := h.Clone()
		bad.Set(name, strings.ToUpper(testfix.Session(1)))
		if _, err := proto.ParseSigned(bad); !errors.Is(err, proto.ErrUnsigned) || !strings.Contains(err.Error(), name) {
			t.Errorf("a malformed %s: %v", name, err)
		}
	}
	plain := http.Header{}
	proto.Sign(priv, 1, http.MethodGet, proto.PathConfig, testfix.Epoch, nil).Apply(plain)
	if plain.Get(proto.HeaderSession) != "" || plain.Get(proto.HeaderTicket) != "" {
		t.Fatalf("a request without a session sent %v", plain)
	}
}
