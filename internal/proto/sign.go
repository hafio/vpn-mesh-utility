package proto

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"time"
)

// Request-signing headers (spec S6.2). Agent endpoints never use cookies, so
// a browser session and agent auth cannot be confused.
const (
	HeaderNode      = "X-Mesh-Node"
	HeaderTimestamp = "X-Mesh-Timestamp"
	HeaderNonce     = "X-Mesh-Nonce"
	HeaderSignature = "X-Mesh-Signature"
	// MaxSkew is how far a request timestamp may be from the server clock.
	MaxSkew = 60 * time.Second
)

var (
	// ErrUnsigned means one of the signing headers is missing or malformed.
	ErrUnsigned = errors.New("request is not signed")
	// ErrBadSignature means the signature does not verify with the node key.
	ErrBadSignature = errors.New("signature does not verify")
	// ErrSkew means the timestamp is outside MaxSkew of the server clock.
	ErrSkew = errors.New("request timestamp outside the allowed clock skew")
)

// reNonce is the shape of crypto/rand.Text: 26 base32 characters, 130 bits.
var reNonce = regexp.MustCompile(`^[A-Z2-7]{26}$`)

// Signed is the parsed signing header set of one request.
type Signed struct {
	NodeID    int64
	Timestamp int64
	// Nonce makes every request unique. Ed25519 signatures are deterministic,
	// so without it two identical requests in the same second (a poll and an
	// SSE-triggered fetch) would carry the same signature and the second
	// would be refused as a replay.
	Nonce     string
	Signature []byte
	// Session and Ticket are the agent's session and the one-time ticket
	// this request presents (amendment A50); a hello, a join and a request
	// made without a session have neither.
	Session string
	Ticket  string
}

// CanonicalRequest is the byte string a request signature covers:
//
//	METHOD "\n" PATH?RAWQUERY "\n" UNIX-SECONDS "\n" NONCE "\n" SESSION "\n" TICKET "\n" hex(sha256(body))
//
// The query string is covered (amendment A8) so a relay cannot rewrite
// parameters of a signed request, and so are the session and the ticket
// (amendment A50), empty when the request has none.
func CanonicalRequest(method, requestURI string, ts int64, nonce, session, ticket string, body []byte) []byte {
	sum := sha256.Sum256(body)
	return []byte(method + "\n" + requestURI + "\n" + strconv.FormatInt(ts, 10) + "\n" + nonce + "\n" + session + "\n" + ticket + "\n" +
		hex.EncodeToString(sum[:]))
}

// Sign returns the header values for one request made by node nodeID at time
// now, with a fresh random nonce and no session. requestURI is the path plus
// raw query exactly as sent (http.Request.URL.RequestURI()).
func Sign(priv ed25519.PrivateKey, nodeID int64, method, requestURI string, now time.Time, body []byte) Signed {
	return SignSession(priv, nodeID, method, requestURI, now, "", "", body)
}

// SignSession is Sign for a request that presents session and ticket.
func SignSession(priv ed25519.PrivateKey, nodeID int64, method, requestURI string, now time.Time, session, ticket string, body []byte) Signed {
	ts := now.Unix()
	nonce := rand.Text()
	return Signed{
		NodeID:    nodeID,
		Timestamp: ts,
		Nonce:     nonce,
		Signature: ed25519.Sign(priv, CanonicalRequest(method, requestURI, ts, nonce, session, ticket, body)),
		Session:   session,
		Ticket:    ticket,
	}
}

// Apply writes the signing headers onto h.
func (s Signed) Apply(h http.Header) {
	h.Set(HeaderNode, strconv.FormatInt(s.NodeID, 10))
	h.Set(HeaderTimestamp, strconv.FormatInt(s.Timestamp, 10))
	h.Set(HeaderNonce, s.Nonce)
	h.Set(HeaderSignature, base64.StdEncoding.EncodeToString(s.Signature))
	if s.Session != "" {
		h.Set(HeaderSession, s.Session)
	}
	if s.Ticket != "" {
		h.Set(HeaderTicket, s.Ticket)
	}
}

// ParseSigned reads the signing headers. It checks shape only; Verify checks
// the signature and the clock.
func ParseSigned(h http.Header) (Signed, error) {
	var s Signed
	id, err := strconv.ParseInt(h.Get(HeaderNode), 10, 64)
	if err != nil || id <= 0 {
		return s, fmt.Errorf("%w: %s must be a positive node id", ErrUnsigned, HeaderNode)
	}
	ts, err := strconv.ParseInt(h.Get(HeaderTimestamp), 10, 64)
	if err != nil || ts <= 0 {
		return s, fmt.Errorf("%w: %s must be unix seconds", ErrUnsigned, HeaderTimestamp)
	}
	nonce := h.Get(HeaderNonce)
	if !reNonce.MatchString(nonce) {
		return s, fmt.Errorf("%w: %s must be 26 base32 characters", ErrUnsigned, HeaderNonce)
	}
	sig, err := base64.StdEncoding.Strict().DecodeString(h.Get(HeaderSignature))
	if err != nil || len(sig) != ed25519.SignatureSize {
		return s, fmt.Errorf("%w: %s must be a base64 ed25519 signature", ErrUnsigned, HeaderSignature)
	}
	session, ticket := h.Get(HeaderSession), h.Get(HeaderTicket)
	for name, v := range map[string]string{HeaderSession: session, HeaderTicket: ticket} {
		if v != "" && !ValidSecret(v) {
			return s, fmt.Errorf("%w: %s must be 64 lowercase hex characters", ErrUnsigned, name)
		}
	}
	return Signed{NodeID: id, Timestamp: ts, Nonce: nonce, Signature: sig, Session: session, Ticket: ticket}, nil
}

// Verify checks the clock first (cheap, and the most common real failure),
// then the signature over the canonical request.
func (s Signed) Verify(pub ed25519.PublicKey, method, requestURI string, body []byte, now time.Time) error {
	if err := s.CheckClock(now); err != nil {
		return err
	}
	if len(pub) != ed25519.PublicKeySize {
		return fmt.Errorf("%w: node has no valid identity key", ErrBadSignature)
	}
	if !ed25519.Verify(pub, CanonicalRequest(method, requestURI, s.Timestamp, s.Nonce, s.Session, s.Ticket, body), s.Signature) {
		return ErrBadSignature
	}
	return nil
}

// CheckClock reports ErrSkew when the timestamp is more than MaxSkew from now.
// The server calls it before looking the node up, so a skewed request costs
// no database read.
func (s Signed) CheckClock(now time.Time) error {
	// skew > 0: the node's clock is behind the server's.
	skew := time.Duration(now.Unix()-s.Timestamp) * time.Second
	if skew > MaxSkew || skew < -MaxSkew {
		dir, d := "behind", skew
		if skew < 0 {
			dir, d = "ahead of", -skew
		}
		return fmt.Errorf("%w: node clock is %s %s the server's (limit %s); sync the node with NTP", ErrSkew, d, dir, MaxSkew)
	}
	return nil
}
