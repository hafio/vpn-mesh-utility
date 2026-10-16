package auth

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"

	"github.com/hafio/vpn-mesh-utility/internal/testfix"
)

func code(t *testing.T, secret string, at time.Time) string {
	t.Helper()
	c, err := totp.GenerateCodeCustom(secret, at, totp.ValidateOpts{Period: 30, Digits: otp.DigitsSix, Algorithm: otp.AlgorithmSHA1})
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestMatchStepAllowsOneStepOfSkew(t *testing.T) {
	k, err := newTOTPKey("meshctl", "admin", "")
	if err != nil {
		t.Fatal(err)
	}
	now := testfix.Epoch
	step := now.Unix() / 30
	for _, c := range []struct {
		at   time.Time
		want int64
	}{{now, step}, {now.Add(-30 * time.Second), step - 1}, {now.Add(30 * time.Second), step + 1}} {
		if got, err := matchStep(k.Secret(), code(t, k.Secret(), c.at), now); err != nil || got != c.want {
			t.Errorf("code from %s: step %d, %v; want %d", c.at, got, err, c.want)
		}
	}
	for _, bad := range []string{code(t, k.Secret(), now.Add(-90*time.Second)), "12345", "abcdef", "1234567"} {
		if _, err := matchStep(k.Secret(), bad, now); !errors.Is(err, ErrBadCode) {
			t.Errorf("matchStep(%q) = %v, want ErrBadCode", bad, err)
		}
	}
}

func TestSealerBindsTheSecretToItsUser(t *testing.T) {
	s, err := NewSealer(bytes.Repeat([]byte{7}, 32))
	if err != nil {
		t.Fatal(err)
	}
	sealed := s.Seal(1, "JBSWY3DPEHPK3PXP")
	if bytes.Contains(sealed, []byte("JBSWY3DPEHPK3PXP")) {
		t.Fatal("the secret is stored in the clear")
	}
	if got, err := s.Open(1, sealed); err != nil || got != "JBSWY3DPEHPK3PXP" {
		t.Fatalf("Open = %q, %v", got, err)
	}
	if _, err := s.Open(2, sealed); err == nil {
		t.Fatal("a secret sealed for user 1 opened for user 2")
	}
	other, _ := NewSealer(bytes.Repeat([]byte{8}, 32))
	if _, err := other.Open(1, sealed); err == nil || !strings.Contains(err.Error(), "reset-totp") {
		t.Fatalf("Open with another key = %v", err)
	}
	tampered := append([]byte(nil), sealed...)
	tampered[len(tampered)-1] ^= 1
	if _, err := s.Open(1, tampered); err == nil {
		t.Fatal("a tampered secret opened")
	}
	if _, err := s.Open(1, []byte{1, 2}); err == nil {
		t.Fatal("a truncated secret opened")
	}
	if _, err := NewSealer([]byte("short")); err == nil {
		t.Fatal("a short key was accepted")
	}
}

func TestLoadOrCreateKey(t *testing.T) {
	path := filepath.Join(t.TempDir(), KeyFile)
	k1, err := LoadOrCreateKey(path)
	if err != nil || len(k1) != 32 {
		t.Fatalf("create = %d bytes, %v", len(k1), err)
	}
	k2, err := LoadOrCreateKey(path)
	if err != nil || !bytes.Equal(k1, k2) {
		t.Fatal("the key changed on reload")
	}
	if runtime.GOOS != "windows" {
		if st, _ := os.Stat(path); st.Mode().Perm() != 0o600 {
			t.Fatalf("mode %v, want 0600", st.Mode().Perm())
		}
		if err := os.Chmod(path, 0o644); err != nil {
			t.Fatal(err)
		}
		if _, err := LoadOrCreateKey(path); err == nil || !strings.Contains(err.Error(), "chmod 600") {
			t.Fatalf("a world-readable key = %v", err)
		}
		if err := os.Chmod(path, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(path, []byte("short"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadOrCreateKey(path); err == nil || !strings.Contains(err.Error(), "32-byte") {
		t.Fatalf("a short key = %v", err)
	}
}

func TestNewTOTPKeyRebuildsAStoredSecret(t *testing.T) {
	k, err := newTOTPKey("meshctl", "admin", "")
	if err != nil {
		t.Fatal(err)
	}
	again, err := newTOTPKey("meshctl", "admin", k.Secret())
	if err != nil || again.Secret() != k.Secret() || again.URL() != k.URL() {
		t.Fatalf("rebuilt key %v, %v; want %v", again, err, k)
	}
	if _, err := newTOTPKey("meshctl", "admin", "not base32!"); err == nil {
		t.Fatal("a damaged secret was accepted")
	}
}
