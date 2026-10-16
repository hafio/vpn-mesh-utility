package server

import (
	"context"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestRunServesAndShutsDown(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the server runs on Linux; Windows only builds it")
	}
	data := t.TempDir()
	ctx, cancel := context.WithCancel(context.Background())
	ready := make(chan net.Addr, 1)
	done := make(chan error, 1)
	go func() {
		done <- Run(ctx, Config{DataDir: data, Listen: "127.0.0.1:0", URL: "http://127.0.0.1:8080", Version: "v1.0.0",
			Ready: func(a net.Addr) { ready <- a }})
	}()
	var addr net.Addr
	select {
	case addr = <-ready:
	case err := <-done:
		t.Fatalf("Run ended early: %v", err)
	case <-time.After(10 * time.Second):
		t.Fatal("server did not start")
	}
	resp, err := http.Get("http://" + addr.String() + "/healthz")
	if err != nil {
		t.Fatal(err)
	}
	b, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != 200 || string(b) != "ok\n" {
		t.Fatalf("healthz: %d %q", resp.StatusCode, b)
	}
	// Every other path is the web UI, with its own security headers.
	resp, err = http.Get("http://" + addr.String() + "/login")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != 200 || !strings.Contains(resp.Header.Get("Content-Security-Policy"), "frame-ancestors 'none'") {
		t.Fatalf("/login: %d %v", resp.StatusCode, resp.Header)
	}
	if _, err := os.Stat(filepath.Join(data, "secret.key")); err != nil {
		t.Fatalf("the TOTP sealing key was not created: %v", err)
	}
	sock := filepath.Join(data, AdminSocket)
	st, err := os.Stat(sock)
	if err != nil || st.Mode().Perm() != 0o600 {
		t.Fatalf("admin socket %v, %v; want mode 0600", st, err)
	}
	if _, err := ListenAdmin(sock); err == nil || !strings.Contains(err.Error(), "already running") {
		t.Fatalf("second server on the same data dir: %v", err)
	}
	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Run = %v", err)
		}
	case <-time.After(15 * time.Second):
		t.Fatal("Run did not shut down")
	}
}

func TestRunRefusesBadConfig(t *testing.T) {
	cases := map[string]Config{
		"http to a public host": {DataDir: t.TempDir(), Listen: "127.0.0.1:0", URL: "http://mesh.example.com"},
		"url with a path":       {DataDir: t.TempDir(), Listen: "127.0.0.1:0", URL: "https://mesh.example.com/x"},
		"data dir with a query": {DataDir: t.TempDir() + "?mode=memory", Listen: "127.0.0.1:0", URL: "https://mesh.example.com"},
		"bad listen address":    {DataDir: t.TempDir(), Listen: "nope", URL: "https://mesh.example.com"},
	}
	for name, c := range cases {
		if err := Run(context.Background(), c); err == nil {
			t.Errorf("%s: Run succeeded", name)
		}
	}
}

func TestListenAdminRefusesALongPath(t *testing.T) {
	if _, err := ListenAdmin("/" + strings.Repeat("d", 120) + "/admin.sock"); err == nil || !strings.Contains(err.Error(), "shorter --data") {
		t.Fatalf("ListenAdmin = %v", err)
	}
}
