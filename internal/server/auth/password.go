package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/crypto/argon2"

	"github.com/hafio/vpn-mesh-utility/internal/validate"
)

// Params are argon2id's cost parameters.
type Params struct {
	Time      uint32
	MemoryKiB uint32
	Threads   uint8
}

// DefaultParams is the cost every new hash uses (spec S8.4): t=3, m=64 MiB,
// p=4. A login with a hash made under other parameters re-hashes it.
func DefaultParams() Params { return Params{Time: 3, MemoryKiB: 64 << 10, Threads: 4} }

const (
	saltLen = 16
	keyLen  = 32
	// Bounds on parameters read back from a stored hash, so a tampered row
	// cannot make a login allocate gigabytes.
	maxMemoryKiB = 1 << 20
	maxTime      = 16
	// Password policy.
	minPassword = 12
	maxPassword = 1024
)

var b64 = base64.RawStdEncoding

// ErrBadHash means a stored hash is not an argon2id PHC string this binary
// can verify.
var ErrBadHash = errors.New("stored password hash is malformed; reset it with `meshctl user passwd`")

// HashPassword returns the argon2id PHC string of password:
// $argon2id$v=19$m=<KiB>,t=<time>,p=<threads>$<salt>$<key>.
func HashPassword(password string, p Params) string {
	salt := make([]byte, saltLen)
	_, _ = rand.Read(salt) // crypto/rand.Read never fails (Go 1.24+)
	key := argon2.IDKey([]byte(password), salt, p.Time, p.MemoryKiB, p.Threads, keyLen)
	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s", argon2.Version, p.MemoryKiB, p.Time, p.Threads,
		b64.EncodeToString(salt), b64.EncodeToString(key))
}

// VerifyPassword checks password against a PHC string in constant time.
// rehash reports that the hash was made under parameters other than want and
// should be replaced now that the password is known.
func VerifyPassword(encoded, password string, want Params) (ok, rehash bool, err error) {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[0] != "" || parts[1] != "argon2id" || parts[2] != fmt.Sprintf("v=%d", argon2.Version) {
		return false, false, ErrBadHash
	}
	var p Params
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &p.MemoryKiB, &p.Time, &p.Threads); err != nil ||
		p.MemoryKiB == 0 || p.MemoryKiB > maxMemoryKiB || p.Time == 0 || p.Time > maxTime || p.Threads == 0 {
		return false, false, ErrBadHash
	}
	salt, err1 := b64.DecodeString(parts[4])
	key, err2 := b64.DecodeString(parts[5])
	if err1 != nil || err2 != nil || len(salt) < 8 || len(key) < 16 {
		return false, false, ErrBadHash
	}
	got := argon2.IDKey([]byte(password), salt, p.Time, p.MemoryKiB, p.Threads, uint32(len(key)))
	if subtle.ConstantTimeCompare(got, key) != 1 {
		return false, false, nil
	}
	return true, p != want || len(key) != keyLen, nil
}

// CheckPassword enforces the password policy: 12 to 1024 bytes of valid
// UTF-8 without control characters. Its errors are *validate.Error without a
// value, so they never contain the password.
func CheckPassword(password string) error {
	reason := ""
	switch {
	case len(password) < minPassword:
		reason = fmt.Sprintf("must be at least %d characters", minPassword)
	case len(password) > maxPassword:
		reason = fmt.Sprintf("must be at most %d bytes", maxPassword)
	case !utf8.ValidString(password):
		reason = "must be valid UTF-8"
	}
	for _, r := range password {
		if reason == "" && unicode.IsControl(r) {
			reason = "must not contain control characters"
		}
	}
	if reason != "" {
		return &validate.Error{Field: "password", Reason: reason}
	}
	return nil
}
