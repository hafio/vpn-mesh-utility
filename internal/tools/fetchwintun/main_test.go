package main

import (
	"archive/zip"
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hafio/vpn-mesh-utility/internal/agent/wintun"
)

func zipOf(t *testing.T, files map[string][]byte) []byte {
	t.Helper()
	var b bytes.Buffer
	w := zip.NewWriter(&b)
	for name, data := range files {
		f, err := w.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := f.Write(data); err != nil {
			t.Fatal(err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	return b.Bytes()
}

// The real release cannot be fetched in a test, so this pins what a download
// that is not it gets: refused before anything is written.
func TestFetchRefusesAnythingButThePinnedRelease(t *testing.T) {
	dir := t.TempDir()
	pin := wintun.DLLs["amd64"]
	fake := zipOf(t, map[string][]byte{pin.Path: []byte("MZ not wintun")})
	if _, _, err := fetch(func(string) ([]byte, error) { return fake, nil }, pin, dir, "wintun_amd64.dll"); err == nil ||
		!strings.Contains(err.Error(), "not the pinned") {
		t.Fatalf("fetch of an unpinned zip = %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "wintun_amd64.dll")); err == nil {
		t.Fatal("an unpinned download was written")
	}
	if _, _, err := fetch(func(string) ([]byte, error) { return nil, errors.New("no network") }, pin, dir, "wintun_amd64.dll"); err == nil ||
		!strings.Contains(err.Error(), "no network") {
		t.Fatalf("fetch offline = %v", err)
	}
}

// A DLL already in place that matches its pin is kept without a download.
func TestFetchKeepsAMatchingDLL(t *testing.T) {
	dir := t.TempDir()
	dll := []byte("MZ a driver")
	pin := wintun.Pin{Path: "wintun/bin/amd64/wintun.dll", SHA256: wintun.Sum(dll)}
	if err := os.WriteFile(filepath.Join(dir, "wintun_amd64.dll"), dll, 0o644); err != nil {
		t.Fatal(err)
	}
	called := false
	get := func(string) ([]byte, error) {
		called = true
		return nil, errors.New("no download expected")
	}
	path, fetched, err := fetch(get, pin, dir, "wintun_amd64.dll")
	if err != nil || fetched || called || path != filepath.Join(dir, "wintun_amd64.dll") {
		t.Fatalf("fetch with the DLL in place = %s, %v, %v (download called: %v)", path, fetched, err, called)
	}
}

// Inside a zip that matches its pin, the DLL must match its own: one
// missing, too large or different is refused.
func TestExtractChecksTheDLL(t *testing.T) {
	dll := []byte("MZ a driver")
	pin := wintun.Pin{Path: "wintun/bin/amd64/wintun.dll", SHA256: wintun.Sum(dll)}
	good := zipOf(t, map[string][]byte{pin.Path: dll, "wintun/LICENSE.txt": []byte("license")})
	if got, err := extract(good, wintun.Sum(good), pin); err != nil || !bytes.Equal(got, dll) {
		t.Fatalf("extract = %q, %v", got, err)
	}
	if _, err := extract(good, wintun.ZipSHA256, pin); err == nil || !strings.Contains(err.Error(), "not the pinned") {
		t.Fatalf("a zip that misses its pin = %v", err)
	}
	for name, z := range map[string][]byte{
		"no DLL":      zipOf(t, map[string][]byte{"wintun/README.md": []byte("x")}),
		"another DLL": zipOf(t, map[string][]byte{pin.Path: []byte("MZ planted")}),
		"too large":   zipOf(t, map[string][]byte{pin.Path: bytes.Repeat([]byte("x"), wintun.MaxDLL+1)}),
		"not a zip":   []byte("not a zip"),
	} {
		if _, err := extract(z, wintun.Sum(z), pin); err == nil {
			t.Errorf("%s: extract accepted it", name)
		}
	}
}
