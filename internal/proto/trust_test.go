package proto_test

import (
	"crypto/ed25519"
	"errors"
	"testing"

	"github.com/hafio/vpn-mesh-utility/internal/proto"
	"github.com/hafio/vpn-mesh-utility/internal/testfix"
)

// A signed body verifies for exactly the kind, node, slot, nonce and bytes
// it was signed for, with the pinned key (amendment A50).
func TestSignedBodies(t *testing.T) {
	p := testfix.Pinned(testfix.VPS1.ID)
	b := testfix.SignedConfig(testfix.ServerConfig())
	if err := p.VerifyBody(proto.SignedConfig, b); err != nil {
		t.Fatalf("a signed config = %v", err)
	}
	_, other := testfix.IdentityKey(9)
	changed := b
	changed.Body = append([]byte(nil), b.Body...)
	changed.Body[0] = ' '
	for name, c := range map[string]struct {
		kind string
		p    proto.Pinned
		b    proto.SignedBody
	}{
		"as a mesh view":      {proto.SignedMesh, p, b},
		"for another node":    {proto.SignedConfig, testfix.Pinned(testfix.VPS2.ID), b},
		"for another slot":    {proto.SignedConfig, proto.Pinned{Key: p.Key, NodeID: p.NodeID, Slot: 1}, b},
		"another nonce":       {proto.SignedConfig, p, proto.SignedBody{Body: b.Body, Nonce: "BBBBBBBBBBBBBBBBBBBBBBBBBB", Signature: b.Signature}},
		"changed bytes":       {proto.SignedConfig, p, changed},
		"another key":         {proto.SignedConfig, proto.Pinned{Key: other.Public().(ed25519.PublicKey), NodeID: p.NodeID}, b},
		"no key":              {proto.SignedConfig, proto.Pinned{NodeID: p.NodeID}, b},
		"no signature":        {proto.SignedConfig, p, proto.SignedBody{Body: b.Body, Nonce: b.Nonce}},
		"a signature of junk": {proto.SignedConfig, p, proto.SignedBody{Body: b.Body, Nonce: b.Nonce, Signature: "not base64!"}},
	} {
		if err := c.p.VerifyBody(c.kind, c.b); !errors.Is(err, proto.ErrUntrusted) {
			t.Errorf("%s = %v, want ErrUntrusted", name, err)
		}
	}
}

// A refusal verifies for its node, code, nonce and time only; only
// `replaced` is definitive.
func TestSignedRefusals(t *testing.T) {
	p := testfix.Pinned(testfix.VPS1.ID)
	sig := proto.SignRefusal(testfix.ConfigKey(), p.NodeID, proto.CodeReplaced, testfix.Nonce, 42)
	if err := p.VerifyRefusal(proto.CodeReplaced, testfix.Nonce, 42, sig); err != nil {
		t.Fatalf("a signed refusal = %v", err)
	}
	for name, err := range map[string]error{
		"another code":  p.VerifyRefusal(proto.CodeNodeDisabled, testfix.Nonce, 42, sig),
		"another nonce": p.VerifyRefusal(proto.CodeReplaced, "BBBBBBBBBBBBBBBBBBBBBBBBBB", 42, sig),
		"another time":  p.VerifyRefusal(proto.CodeReplaced, testfix.Nonce, 43, sig),
		"another node":  testfix.Pinned(testfix.VPS2.ID).VerifyRefusal(proto.CodeReplaced, testfix.Nonce, 42, sig),
	} {
		if !errors.Is(err, proto.ErrUntrusted) {
			t.Errorf("%s = %v", name, err)
		}
	}
	if !proto.Definitive(proto.CodeReplaced) || proto.Definitive(proto.CodeNodeDisabled) || proto.Definitive(proto.CodeUnauthorized) {
		t.Error("Definitive is wrong")
	}
}

func TestParseConfigKey(t *testing.T) {
	if k, err := proto.ParseConfigKey(testfix.ConfigKeyB64()); err != nil || !k.Equal(testfix.ConfigKey().Public()) {
		t.Fatalf("ParseConfigKey = %v, %v", k, err)
	}
	for _, bad := range []string{"", "c2hvcnQ=", "not base64!", testfix.ConfigKeyB64() + "AA"} {
		if _, err := proto.ParseConfigKey(bad); err == nil {
			t.Errorf("ParseConfigKey(%q) succeeded", bad)
		}
	}
}
