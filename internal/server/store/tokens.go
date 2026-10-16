package store

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"time"
)

// TokenTTL is how long a join token stays valid (spec S6.1).
const TokenTTL = time.Hour

var (
	// ErrTokenUnknown means no such token was ever issued (or its node is gone).
	ErrTokenUnknown = errors.New("join token is not valid")
	// ErrTokenUsed means the token was already consumed by a join.
	ErrTokenUsed = errors.New("join token was already used")
	// ErrTokenExpired means the token is past its expiry.
	ErrTokenExpired = errors.New("join token has expired")
)

// NewToken returns 32 random bytes, base64url without padding (43 chars).
func NewToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate join token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// HashToken is the stored form of a token.
func HashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// IssueToken creates a fresh single-use token for a node and revokes the
// node's unused ones, so only the newest join command works.
func (t *Tx) IssueToken(nodeID int64, ttl time.Duration) (token string, expires time.Time, err error) {
	token, err = NewToken()
	if err != nil {
		return "", time.Time{}, err
	}
	if _, err := t.exec(`DELETE FROM join_tokens WHERE node_id = ? AND used_at IS NULL`, nodeID); err != nil {
		return "", time.Time{}, fmt.Errorf("revoke old tokens of node %d: %w", nodeID, err)
	}
	expires = t.now.Add(ttl)
	if _, err := t.exec(`INSERT INTO join_tokens (token_hash, node_id, created_at, expires_at) VALUES (?, ?, ?, ?)`,
		HashToken(token), nodeID, t.now.Unix(), expires.Unix()); err != nil {
		return "", time.Time{}, fmt.Errorf("store token for node %d: %w", nodeID, err)
	}
	return token, expires, nil
}

// ConsumeToken marks a token used and returns its node. The check and the
// write are one UPDATE, so two concurrent joins with the same token cannot
// both succeed.
func (t *Tx) ConsumeToken(token string) (int64, error) {
	h := HashToken(token)
	var nodeID int64
	err := t.queryRow(`UPDATE join_tokens SET used_at = ?
		WHERE token_hash = ? AND used_at IS NULL AND expires_at > ?
		RETURNING node_id`, t.now.Unix(), h, t.now.Unix()).Scan(&nodeID)
	if err == nil {
		return nodeID, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return 0, fmt.Errorf("consume join token: %w", err)
	}
	// Explain why, so the operator knows whether to reissue.
	var used sql.NullInt64
	var expires int64
	err = t.queryRow(`SELECT used_at, expires_at FROM join_tokens WHERE token_hash = ?`, h).Scan(&used, &expires)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return 0, ErrTokenUnknown
	case err != nil:
		return 0, fmt.Errorf("consume join token: %w", err)
	case used.Valid:
		return 0, fmt.Errorf("%w at %s", ErrTokenUsed, time.Unix(used.Int64, 0).UTC().Format(time.RFC3339))
	default:
		return 0, fmt.Errorf("%w at %s", ErrTokenExpired, time.Unix(expires, 0).UTC().Format(time.RFC3339))
	}
}
