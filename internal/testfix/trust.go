package testfix

import (
	"bytes"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"

	"github.com/hafio/vpn-mesh-utility/internal/proto"
)

// ConfigKey is the control plane's signing key of the tests (amendment
// A50), deterministic like IdentityKey.
func ConfigKey() ed25519.PrivateKey {
	return ed25519.NewKeyFromSeed(bytes.Repeat([]byte{0xc0}, ed25519.SeedSize))
}

// ConfigKeyB64 is ConfigKey's public half as the join answer carries it.
func ConfigKeyB64() string {
	return base64.StdEncoding.EncodeToString(ConfigKey().Public().(ed25519.PublicKey))
}

// Pinned is what an agent of node in slot 0 pins at a test join.
func Pinned(node int64) proto.Pinned {
	return proto.Pinned{Key: ConfigKey().Public().(ed25519.PublicKey), NodeID: node}
}

// Nonce is a request nonce of the shape proto.Sign makes.
const Nonce = "AAAAAAAAAAAAAAAAAAAAAAAAAA"

// Signed is v as the control plane serves it -- JSON and a newline -- and
// signs it as kind for its node in slot 0 (a config's own slot when v is
// one), answering the request with Nonce.
func Signed(kind string, node int64, v any) proto.SignedBody {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err) // the fixtures always encode
	}
	b = append(b, '\n')
	var slot proto.Slot
	if c, ok := v.(proto.NodeConfig); ok {
		slot = c.Node.Slot
	}
	return proto.SignedBody{Body: b, Nonce: Nonce, Signature: proto.SignBody(ConfigKey(), kind, node, slot, Nonce, b)}
}

// SignedConfig is c signed for its own node.
func SignedConfig(c proto.NodeConfig) proto.SignedBody {
	return Signed(proto.SignedConfig, c.Node.ID, c)
}
