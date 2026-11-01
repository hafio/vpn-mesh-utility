package api

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hafio/vpn-mesh-utility/internal/proto"
)

func distHarness(t *testing.T, files map[string]string) http.Handler {
	t.Helper()
	dir := t.TempDir()
	for name, body := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	h := newHarness(t)
	return New(Deps{DB: h.db, Hub: h.hub, Reconciler: h.rec, Now: h.clock.Now, ServerURL: testURL, Version: "v1.2.3", DistDir: dir})
}

var (
	sumAmd64   = strings.Repeat("a", 64)
	sumArm64   = strings.Repeat("b", 64)
	sumWindows = strings.Repeat("c", 64)
	sums       = sumAmd64 + "  meshctl-linux-amd64\n" + sumArm64 + "  meshctl-linux-arm64\n" + sumWindows + "  meshctl-windows-amd64.exe\n"
)

func TestDownloadsServeOnlyAllowlistedNames(t *testing.T) {
	h := distHarness(t, map[string]string{"meshctl-linux-amd64": "ELF", "SHA256SUMS": sums, "secret.key": "nope"})
	w, _ := serve(h, httptest.NewRequest(http.MethodGet, proto.PathDownload+"meshctl-linux-amd64", nil))
	if w.Code != 200 || w.Body.String() != "ELF" || w.Header().Get("Content-Type") != "application/octet-stream" {
		t.Fatalf("binary: %d %q %v", w.Code, w.Body.String(), w.Header())
	}
	for _, name := range []string{"secret.key", "meshctl-linux-amd64.sig", "meshctl-linux-arm64"} {
		w, eb := serve(h, httptest.NewRequest(http.MethodGet, proto.PathDownload+name, nil))
		if w.Code != 404 || eb.Code != proto.CodeNotFound {
			t.Errorf("%s: %d %q", name, w.Code, eb.Code)
		}
	}
}

// install.sh, reinstall.sh and uninstall.sh (amendment A55) carry this
// server's origin, version and checksums, and the unit.
func TestInstallScriptPinsTheChecksums(t *testing.T) {
	h := distHarness(t, map[string]string{"SHA256SUMS": sums})
	for _, path := range []string{proto.PathInstall, proto.PathReinstall, proto.PathUninstall} {
		w, eb := serve(h, httptest.NewRequest(http.MethodGet, path, nil))
		if w.Code != 200 || w.Header().Get("Content-Type") != "text/x-shellscript; charset=utf-8" {
			t.Fatalf("%s: %d %s", path, w.Code, eb.Message)
		}
		body := w.Body.String()
		for _, want := range []string{"SCRIPT=" + strings.TrimPrefix(path, "/"), "ORIGIN='" + testURL + "'", "VERSION='v1.2.3'",
			"SHA256_AMD64='" + sumAmd64 + "'", "SHA256_ARM64='" + sumArm64 + "'"} {
			if !strings.Contains(body, want) {
				t.Errorf("%s lacks %s", path, want)
			}
		}
		if strings.Contains(body, "@@") || !strings.Contains(body, "ExecStart=/usr/local/bin/meshctl agent run --mesh %i") {
			t.Errorf("%s still has a marker, or lacks the unit", path)
		}
	}
}

// install.ps1 (amendment A51), reinstall.ps1 and uninstall.ps1 (amendment
// A55) pin the Windows binary's checksum the same way, and need it.
func TestInstallPS1PinsTheChecksum(t *testing.T) {
	h := distHarness(t, map[string]string{"SHA256SUMS": sums})
	linuxOnly := distHarness(t, map[string]string{"SHA256SUMS": sumAmd64 + "  meshctl-linux-amd64\n" + sumArm64 + "  meshctl-linux-arm64\n"})
	for _, path := range []string{proto.PathInstallPS1, proto.PathReinstallPS1, proto.PathUninstallPS1} {
		w, eb := serve(h, httptest.NewRequest(http.MethodGet, path, nil))
		if w.Code != 200 || w.Header().Get("Content-Type") != "text/plain; charset=utf-8" {
			t.Fatalf("%s: %d %s", path, w.Code, eb.Message)
		}
		body := w.Body.String()
		for _, want := range []string{"$Self = '" + strings.TrimPrefix(path, "/") + "'", "$Origin = '" + testURL + "'", "$Version = 'v1.2.3'",
			"$Sha256 = '" + sumWindows + "'"} {
			if !strings.Contains(body, want) {
				t.Errorf("%s lacks %s", path, want)
			}
		}
		if strings.Contains(body, "@@") {
			t.Errorf("%s still has a marker", path)
		}
		if w, _ := serve(linuxOnly, httptest.NewRequest(http.MethodGet, path, nil)); w.Code != 500 {
			t.Fatalf("%s without a Windows checksum: %d", path, w.Code)
		}
	}
}

func TestInstallScriptFailsWithoutChecksums(t *testing.T) {
	if w, _ := serve(distHarness(t, nil), httptest.NewRequest(http.MethodGet, proto.PathInstall, nil)); w.Code != 500 {
		t.Fatalf("no SHA256SUMS: %d", w.Code)
	}
	partial := distHarness(t, map[string]string{"SHA256SUMS": sumAmd64 + "  meshctl-linux-amd64\n"})
	if w, _ := serve(partial, httptest.NewRequest(http.MethodGet, proto.PathInstall, nil)); w.Code != 500 {
		t.Fatalf("missing arm64 checksum: %d", w.Code)
	}
	h := newHarness(t)
	for _, p := range []string{proto.PathInstall, proto.PathInstallPS1, proto.PathReinstall, proto.PathUninstall, proto.PathReinstallPS1,
		proto.PathUninstallPS1, proto.PathDownload + "SHA256SUMS"} {
		if w, eb := serve(h.pub, httptest.NewRequest(http.MethodGet, p, nil)); w.Code != 404 || !strings.Contains(eb.Message, "MESHCTL_DIST_DIR") {
			t.Errorf("%s without a dist dir: %d %q", p, w.Code, eb.Message)
		}
	}
}
