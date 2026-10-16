package auth

import (
	"errors"
	"strings"
	"testing"
)

// cheap keeps the tests fast; one test checks the real parameters.
var cheap = Params{Time: 1, MemoryKiB: 64, Threads: 1}

func TestHashUsesTheSpecParameters(t *testing.T) {
	h := HashPassword("correct horse battery", DefaultParams())
	if !strings.HasPrefix(h, "$argon2id$v=19$m=65536,t=3,p=4$") {
		t.Fatalf("hash %q does not carry argon2id t=3 m=64MiB p=4", h)
	}
	if ok, rehash, err := VerifyPassword(h, "correct horse battery", DefaultParams()); !ok || rehash || err != nil {
		t.Fatalf("Verify = %v, %v, %v", ok, rehash, err)
	}
}

func TestVerifyPassword(t *testing.T) {
	h := HashPassword("correct horse battery", cheap)
	if h == HashPassword("correct horse battery", cheap) {
		t.Fatal("two hashes of one password must differ (random salt)")
	}
	if ok, rehash, err := VerifyPassword(h, "correct horse battery", cheap); !ok || rehash || err != nil {
		t.Fatalf("right password = %v, %v, %v", ok, rehash, err)
	}
	if ok, _, err := VerifyPassword(h, "correct horse batterY", cheap); ok || err != nil {
		t.Fatalf("wrong password = %v, %v", ok, err)
	}
	if ok, rehash, _ := VerifyPassword(h, "correct horse battery", DefaultParams()); !ok || !rehash {
		t.Fatal("a hash under other parameters must ask for a rehash")
	}
}

func TestVerifyRefusesMalformedHashes(t *testing.T) {
	good := HashPassword("correct horse battery", cheap)
	parts := strings.Split(good, "$")
	for name, h := range map[string]string{
		"empty":            "",
		"bcrypt":           "$2y$10$abcdefghijklmnopqrstuuvwxyzABCDEFGHIJKLMNOPQRSTUVWX",
		"argon2i":          strings.Replace(good, "argon2id", "argon2i", 1),
		"other version":    strings.Replace(good, "v=19", "v=16", 1),
		"huge memory":      strings.Join([]string{"", "argon2id", "v=19", "m=4194304,t=1,p=1", parts[4], parts[5]}, "$"),
		"zero time":        strings.Join([]string{"", "argon2id", "v=19", "m=64,t=0,p=1", parts[4], parts[5]}, "$"),
		"salt not base64":  strings.Join([]string{"", "argon2id", "v=19", parts[3], "!!!", parts[5]}, "$"),
		"truncated digest": strings.Join([]string{"", "argon2id", "v=19", parts[3], parts[4], "AAAA"}, "$"),
	} {
		if ok, _, err := VerifyPassword(h, "correct horse battery", cheap); ok || !errors.Is(err, ErrBadHash) {
			t.Errorf("%s: Verify = %v, %v", name, ok, err)
		}
	}
}

func TestCheckPasswordNeverEchoesThePassword(t *testing.T) {
	for _, pw := range []string{"short", strings.Repeat("a", 1025), "twelve chars\x00!", "bad utf8 \xff\xfe xx"} {
		err := CheckPassword(pw)
		if err == nil {
			t.Errorf("CheckPassword(%q) accepted", pw)
			continue
		}
		if strings.Contains(err.Error(), pw) {
			t.Errorf("the error echoes the password: %v", err)
		}
	}
	if err := CheckPassword("correct horse battery"); err != nil {
		t.Fatal(err)
	}
}
