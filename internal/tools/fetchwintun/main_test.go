package main

import (
	"archive/zip"
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
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
	if _, _, err := fetch(func(string) ([]byte, error) { return fake, nil }, wintun.ZipSHA256, pin, dir, "wintun_amd64.dll"); err == nil ||
		!strings.Contains(err.Error(), "not the pinned") {
		t.Fatalf("fetch of an unpinned zip = %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "wintun_amd64.dll")); err == nil {
		t.Fatal("an unpinned download was written")
	}
	if _, _, err := fetch(func(string) ([]byte, error) { return nil, errors.New("no network") }, wintun.ZipSHA256, pin, dir, "wintun_amd64.dll"); err == nil ||
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
	path, fetched, err := fetch(get, wintun.ZipSHA256, pin, dir, "wintun_amd64.dll")
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

// The command: bad flags are exit 2; a download that is not the pinned
// release is exit 1 with the reason, and writes nothing.
func TestRun(t *testing.T) {
	var out, errOut bytes.Buffer
	for _, args := range [][]string{{"-arch", "x86"}, {"extra"}, {"-nope"}} {
		if code := run(args, &out, &errOut, nil); code != 2 {
			t.Errorf("run(%q) = %d, want 2", args, code)
		}
	}
	dir := t.TempDir()
	fake := func(string) ([]byte, error) { return []byte("not the release"), nil }
	if code := run([]string{"-out", dir}, &out, &errOut, fake); code != 1 || !strings.Contains(errOut.String(), "not the pinned") {
		t.Fatalf("run with an unpinned download = %d\n%s", code, errOut.String())
	}
	if entries, _ := os.ReadDir(dir); len(entries) != 0 {
		t.Fatalf("a refused download left %v", entries)
	}
}

// download reads an https answer up to the size limit, and reports a
// failed status, an oversize body and an unreachable server.
func TestDownload(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/ok":
			_, _ = w.Write([]byte("zip bytes"))
		case "/big":
			_, _ = w.Write(bytes.Repeat([]byte("x"), wintun.MaxZip+1))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()
	if b, err := download(srv.URL + "/ok"); err != nil || string(b) != "zip bytes" {
		t.Fatalf("download = %q, %v", b, err)
	}
	for path, want := range map[string]string{"/missing": "404", "/big": "size limit"} {
		if _, err := download(srv.URL + path); err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("download %s = %v, want %q", path, err, want)
		}
	}
	srv.Close()
	if _, err := download(srv.URL + "/ok"); err == nil {
		t.Fatal("download from a server that is gone succeeded")
	}
}

// A zip and DLL that match are written; a DLL that cannot be written is the
// error, after the checks passed.
func TestFetchWritesACheckedDLL(t *testing.T) {
	dll := []byte("MZ a driver")
	pin := wintun.Pin{Path: "wintun/bin/amd64/wintun.dll", SHA256: wintun.Sum(dll)}
	z := zipOf(t, map[string][]byte{pin.Path: dll})
	get := func(string) ([]byte, error) { return z, nil }
	if _, _, err := fetch(get, wintun.Sum(z), pin, filepath.Join(t.TempDir(), "missing"), "wintun_amd64.dll"); err == nil || strings.Contains(err.Error(), "pinned") {
		t.Fatalf("fetch into a missing dir = %v", err)
	}
	dir := t.TempDir()
	if path, fetched, err := fetch(get, wintun.Sum(z), pin, dir, "wintun_amd64.dll"); err != nil || !fetched {
		t.Fatalf("fetch of a matching zip = %s, %v, %v", path, fetched, err)
	}
	if b, err := os.ReadFile(filepath.Join(dir, "wintun_amd64.dll")); err != nil || !bytes.Equal(b, dll) {
		t.Fatalf("the written DLL = %q, %v", b, err)
	}
}
