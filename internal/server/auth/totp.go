package auth

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base32"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

// TOTP parameters: RFC 6238 defaults, which every authenticator app accepts.
const (
	totpPeriod = 30
	// totpSkew is how many steps either side of now a code may come from.
	totpSkew = 1
)

// KeyFile holds the key that seals TOTP secrets at rest, in the data dir.
const KeyFile = "secret.key"

var reCode = regexp.MustCompile(`^[0-9]{6}$`)

// ErrBadCode means the TOTP code is wrong (or not six digits).
var ErrBadCode = errors.New("the code is wrong; check the authenticator app's clock and try the next code")

// LoadOrCreateKey reads the 32-byte sealing key at path, creating it (mode
// 0600) on first start. A key readable by other users is refused.
func LoadOrCreateKey(path string) ([]byte, error) {
	st, err := os.Stat(path)
	if errors.Is(err, fs.ErrNotExist) {
		key := make([]byte, 32)
		_, _ = rand.Read(key) // crypto/rand.Read never fails (Go 1.24+)
		f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
		if err != nil {
			return nil, fmt.Errorf("create %s: %w", path, err)
		}
		_, werr := f.Write(key)
		if werr == nil {
			werr = f.Sync()
		}
		if cerr := f.Close(); werr == nil {
			werr = cerr
		}
		if werr != nil {
			return nil, fmt.Errorf("write %s: %w", path, werr)
		}
		return key, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	if runtime.GOOS != "windows" && st.Mode().Perm()&0o077 != 0 {
		return nil, fmt.Errorf("%s is readable by other users (mode %04o); run `chmod 600 %s`", path, st.Mode().Perm(), path)
	}
	key, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	if len(key) != 32 {
		return nil, fmt.Errorf("%s holds %d bytes, not a 32-byte key; restore it from the volume backup, or delete it and reset every user's TOTP", path, len(key))
	}
	return key, nil
}

// Sealer encrypts TOTP secrets at rest with AES-256-GCM. Each secret is
// bound to its user id, so a sealed secret copied to another row does not
// open.
type Sealer struct {
	aead cipher.AEAD
}

// NewSealer returns a sealer for a 32-byte key.
func NewSealer(key []byte) (*Sealer, error) {
	block, err := aes.NewCipher(key)
	if err != nil || len(key) != 32 {
		return nil, errors.New("the TOTP sealing key must be 32 bytes")
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("TOTP sealing key: %w", err)
	}
	return &Sealer{aead: aead}, nil
}

func aad(userID int64) []byte { return []byte("meshctl-totp:user:" + strconv.FormatInt(userID, 10)) }

// Seal returns nonce || ciphertext.
func (s *Sealer) Seal(userID int64, secret string) []byte {
	nonce := make([]byte, s.aead.NonceSize())
	_, _ = rand.Read(nonce) // crypto/rand.Read never fails (Go 1.24+)
	return s.aead.Seal(nonce, nonce, []byte(secret), aad(userID))
}

// Open recovers a secret sealed for userID.
func (s *Sealer) Open(userID int64, sealed []byte) (string, error) {
	n := s.aead.NonceSize()
	if len(sealed) < n {
		return "", errors.New("the stored TOTP secret is damaged; reset it with `meshctl user reset-totp`")
	}
	plain, err := s.aead.Open(nil, sealed[:n], sealed[n:], aad(userID))
	if err != nil {
		return "", errors.New("the stored TOTP secret does not decrypt (was secret.key replaced?); reset it with `meshctl user reset-totp`")
	}
	return string(plain), nil
}

// newTOTPKey returns the otpauth:// key for an authenticator app: for a new
// random secret when secret is "", else for that base32 secret.
func newTOTPKey(issuer, account, secret string) (*otp.Key, error) {
	opts := totp.GenerateOpts{Issuer: issuer, AccountName: account, Period: totpPeriod, SecretSize: 20}
	if secret != "" {
		raw, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(secret)
		if err != nil {
			return nil, errors.New("the stored TOTP secret is damaged; reset it with `meshctl user reset-totp`")
		}
		opts.Secret = raw
	}
	k, err := totp.Generate(opts)
	if err != nil {
		return nil, fmt.Errorf("generate TOTP key: %w", err)
	}
	return k, nil
}

// matchStep returns the time step whose code equals code, looking totpSkew
// steps either side of now. The caller refuses a step at or before the last
// one it accepted (replay).
func matchStep(secret, code string, now time.Time) (int64, error) {
	if !reCode.MatchString(code) {
		return 0, ErrBadCode
	}
	cur := now.Unix() / totpPeriod
	for d := -totpSkew; d <= totpSkew; d++ {
		step := cur + int64(d)
		want, err := totp.GenerateCodeCustom(secret, time.Unix(step*totpPeriod, 0),
			totp.ValidateOpts{Period: totpPeriod, Digits: otp.DigitsSix, Algorithm: otp.AlgorithmSHA1})
		if err != nil {
			return 0, fmt.Errorf("compute TOTP code: %w", err)
		}
		if subtle.ConstantTimeCompare([]byte(want), []byte(code)) == 1 {
			return step, nil
		}
	}
	return 0, ErrBadCode
}
