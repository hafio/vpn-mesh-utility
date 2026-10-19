package wintun

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The driver is written once, beside the binary, only when it matches its
// pin; one already there that matches is left alone, one that does not is
// replaced.
func TestInstall(t *testing.T) {
	dir := t.TempDir()
	dll := []byte("MZ a driver")
	want := Sum(dll)
	path, wrote, err := install(dir, dll, want)
	if err != nil || !wrote || path != filepath.Join(dir, FileName) {
		t.Fatalf("install = %s, %v, %v", path, wrote, err)
	}
	if _, wrote, err := install(dir, dll, want); err != nil || wrote {
		t.Fatalf("install again = %v, %v; want nothing written", wrote, err)
	}
	if err := os.WriteFile(path, []byte("planted"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, wrote, err := install(dir, dll, want); err != nil || !wrote {
		t.Fatalf("install over a changed file = %v, %v; want it replaced", wrote, err)
	}
	if b, _ := os.ReadFile(path); string(b) != string(dll) {
		t.Fatalf("the file holds %q", b)
	}

	if _, _, err := install(dir, dll, Sum([]byte("other"))); err == nil || !strings.Contains(err.Error(), "not the pinned") {
		t.Fatalf("a bundle that misses its pin = %v", err)
	}
	if _, _, err := install(dir, nil, want); !errors.Is(err, ErrNotBundled) {
		t.Fatalf("no bundle = %v", err)
	}
	if _, _, err := install(filepath.Join(dir, "missing"), dll, want); err == nil || !strings.Contains(err.Error(), "write") {
		t.Fatalf("install into a missing dir = %v", err)
	}
}

// Off the Windows release build nothing is bundled, so Ensure says how to
// get a binary that has the driver.
func TestEnsureWithoutABundle(t *testing.T) {
	if _, _, err := Ensure(); !errors.Is(err, ErrNotBundled) {
		t.Fatalf("Ensure = %v", err)
	}
	for arch, p := range DLLs {
		if len(p.SHA256) != 64 || !strings.HasPrefix(p.Path, "wintun/bin/"+arch+"/") {
			t.Errorf("pin %s = %+v", arch, p)
		}
	}
}
