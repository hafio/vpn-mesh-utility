package server

import (
	"context"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/hafio/vpn-mesh-utility/internal/server/api"
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

// The first start gives the mesh an IPv6 prefix (M8: both families by
// default); given up, it stays given up across a restart.
func TestRunGivesTheMeshIPv6(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the server runs on Linux; Windows only builds it")
	}
	data := t.TempDir()
	run := func(check func(c *api.AdminClient)) {
		t.Helper()
		ctx, cancel := context.WithCancel(context.Background())
		ready, done := make(chan net.Addr, 1), make(chan error, 1)
		go func() {
			done <- Run(ctx, Config{DataDir: data, Listen: "127.0.0.1:0", URL: "http://127.0.0.1:8080", Version: "v1.0.0",
				Ready: func(a net.Addr) { ready <- a }})
		}()
		select {
		case <-ready:
		case err := <-done:
			t.Fatalf("Run ended early: %v", err)
		case <-time.After(10 * time.Second):
			t.Fatal("server did not start")
		}
		check(api.NewAdminClient(filepath.Join(data, AdminSocket)))
		cancel()
		if err := <-done; err != nil {
			t.Fatalf("Run = %v", err)
		}
	}
	run(func(c *api.AdminClient) {
		v, err := c.IPv6(context.Background())
		if err != nil || !strings.HasPrefix(v.Prefix, "fd") || !strings.HasSuffix(v.Prefix, "::/64") {
			t.Fatalf("after the first start: %+v, %v", v, err)
		}
		if _, err := c.SetIPv6(context.Background(), api.IPv6Request{Prefix: api.IPv6Off}); err != nil {
			t.Fatal(err)
		}
	})
	run(func(c *api.AdminClient) {
		if v, err := c.IPv6(context.Background()); err != nil || v.Prefix != "" {
			t.Fatalf("after a restart: %+v, %v", v, err)
		}
	})
}

// With no admin, every start logs a setup code for the setup page (amendment
// A52), and the logged code is the one the page takes; once an admin exists
// a start logs none.
func TestRunLogsTheSetupCode(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the server runs on Linux; Windows only builds it")
	}
	data := t.TempDir()
	run := func(check func(addr string, log *syncBuffer)) {
		t.Helper()
		log := &syncBuffer{}
		ctx, cancel := context.WithCancel(context.Background())
		ready, done := make(chan net.Addr, 1), make(chan error, 1)
		go func() {
			done <- Run(ctx, Config{DataDir: data, Listen: "127.0.0.1:0", URL: "http://127.0.0.1:8080", Version: "v1.0.0",
				Log: slog.New(slog.NewTextHandler(log, nil)), Ready: func(a net.Addr) { ready <- a }})
		}()
		var addr net.Addr
		select {
		case addr = <-ready:
		case err := <-done:
			t.Fatalf("Run ended early: %v", err)
		case <-time.After(10 * time.Second):
			t.Fatal("server did not start")
		}
		check(addr.String(), log)
		cancel()
		if err := <-done; err != nil {
			t.Fatalf("Run = %v", err)
		}
	}
	run(func(addr string, log *syncBuffer) {
		re := regexp.MustCompile(`setup_page=http://127.0.0.1:8080/setup setup_code=((?:[a-z2-7]{4}-){7}[a-z2-7]{4})`)
		m := re.FindStringSubmatch(log.String())
		if m == nil {
			t.Fatalf("no setup code in the log:\n%s", log.String())
		}
		form := url.Values{"code": {m[1]}, "username": {"boss"}, "password": {"correct horse battery"}, "password2": {"correct horse battery"}}
		req, err := http.NewRequest(http.MethodPost, "http://"+addr+"/setup", strings.NewReader(form.Encode()))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("Origin", "http://127.0.0.1:8080")
		client := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusSeeOther || resp.Header.Get("Location") != "/login/enroll" {
			t.Fatalf("POST /setup with the logged code = %d %v", resp.StatusCode, resp.Header)
		}
	})
	run(func(_ string, log *syncBuffer) {
		if strings.Contains(log.String(), "setup_code") {
			t.Fatalf("a start with an admin logged a setup code:\n%s", log.String())
		}
	})
}
