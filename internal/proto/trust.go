package proto

import (
	"crypto/ed25519"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
)

// The control plane signs what an agent applies or acts on (amendment A50):
// every config and mesh view, and every definitive refusal. Its key is
// pinned at join (JoinResponse.ConfigKey), so a TLS terminator or proxy in
// the path can read these answers but neither change, replay nor forge them.
// Each signature covers the node it is for, its mesh's slot and the nonce of
// the request it answers, so an answer signed for another node, mesh or
// request never verifies.
const (
	// HeaderConfigSignature carries the signature of a config or mesh view
	// body.
	HeaderConfigSignature = "X-Mesh-Config-Signature"
	// HeaderRefusalSignature carries the signature of a definitive refusal.
	HeaderRefusalSignature = "X-Mesh-Refusal-Signature"
)

// The kinds of signed bodies.
const (
	SignedConfig = "meshctl-config"
	SignedMesh   = "meshctl-mesh"
)

// CodeReplaced is the definitive refusal of a request signed with a key the
// node no longer has: a token was issued for it, or it joined again
// elsewhere. The agent stops (amendment A50).
const CodeReplaced = "replaced"

// Definitive reports whether a refusal with code ends the agent's run --
// when, and only when, its signature verifies: replaced, duplicate and
// foreign.
func Definitive(code string) bool {
	return code == CodeReplaced || code == CodeDuplicate || code == CodeForeign
}

// SignedBody is an answer as the control plane signed it: the exact body,
// the nonce of the request it answers and the signature.
type SignedBody struct {
	Body      []byte
	Nonce     string
	Signature string
}

// bodyMessage is what a body signature covers.
func bodyMessage(kind string, node int64, slot Slot, nonce string, body []byte) []byte {
	head := kind + "\n" + strconv.FormatInt(node, 10) + "\n" + strconv.Itoa(int(slot)) + "\n" + nonce + "\n"
	return append([]byte(head), body...)
}

// refusalMessage is what a refusal signature covers.
func refusalMessage(node int64, code, nonce string, serverTime int64) []byte {
	return []byte("meshctl-refusal\n" + strconv.FormatInt(node, 10) + "\n" + code + "\n" + nonce + "\n" + strconv.FormatInt(serverTime, 10))
}

// SignBody is the control plane's signature of body, of kind, for node in
// slot, answering the request with nonce.
func SignBody(key ed25519.PrivateKey, kind string, node int64, slot Slot, nonce string, body []byte) string {
	return base64.StdEncoding.EncodeToString(ed25519.Sign(key, bodyMessage(kind, node, slot, nonce, body)))
}

// SignRefusal is the control plane's signature of a definitive refusal
// with code, for node, answering the request with nonce at serverTime.
func SignRefusal(key ed25519.PrivateKey, node int64, code, nonce string, serverTime int64) string {
	return base64.StdEncoding.EncodeToString(ed25519.Sign(key, refusalMessage(node, code, nonce, serverTime)))
}

// ErrUntrusted means an answer's signature does not verify with the pinned
// key: it was changed on the way, replayed from another request, meant for
// another node or mesh, or signed by another control plane.
var ErrUntrusted = errors.New("the answer is not signed by this node's control plane")

// Pinned is what an agent checks the control plane's answers with: the key
// it pinned at join, and the node and slot the answers must be for.
type Pinned struct {
	Key    ed25519.PublicKey
	NodeID int64
	Slot   Slot
}

// VerifyBody checks a signed body of kind.
func (p Pinned) VerifyBody(kind string, b SignedBody) error {
	if err := p.verify(b.Signature, bodyMessage(kind, p.NodeID, p.Slot, b.Nonce, b.Body)); err != nil {
		return fmt.Errorf("%s: %w", kind, err)
	}
	return nil
}

// VerifyRefusal checks a refusal with code answering the request with
// nonce.
func (p Pinned) VerifyRefusal(code, nonce string, serverTime int64, sig string) error {
	return p.verify(sig, refusalMessage(p.NodeID, code, nonce, serverTime))
}

func (p Pinned) verify(sig string, msg []byte) error {
	raw, err := base64.StdEncoding.Strict().DecodeString(sig)
	if len(p.Key) != ed25519.PublicKeySize || err != nil || !ed25519.Verify(p.Key, msg, raw) {
		return ErrUntrusted
	}
	return nil
}

// ParseConfigKey reads a control plane's public key as the join answer
// carries it.
func ParseConfigKey(s string) (ed25519.PublicKey, error) {
	raw, err := base64.StdEncoding.Strict().DecodeString(s)
	if err != nil || len(raw) != ed25519.PublicKeySize {
		return nil, errors.New("the control plane's config key is not a base64 Ed25519 public key; is it an older meshctl? Upgrade it")
	}
	return ed25519.PublicKey(raw), nil
}
