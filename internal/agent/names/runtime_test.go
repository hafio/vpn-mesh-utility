package names

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/netip"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"

	"github.com/hafio/vpn-mesh-utility/internal/proto"
)

// fakeRuntime serves a Docker-compatible API on a unix socket and records
// every request.
type fakeRuntime struct {
	mu   sync.Mutex
	reqs []string
}

func (f *fakeRuntime) seen() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]string(nil), f.reqs...)
}

func serveRuntime(t *testing.T, h http.HandlerFunc) (*Runtime, *fakeRuntime) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "rt.sock")
	ln, err := net.Listen("unix", path)
	if err != nil {
		t.Skipf("no unix sockets here: %v", err)
	}
	f := &fakeRuntime{}
	srv := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f.mu.Lock()
		f.reqs = append(f.reqs, r.Method+" "+r.URL.RequestURI())
		f.mu.Unlock()
		h(w, r)
	})}
	go func() { _ = srv.Serve(ln) }()
	t.Cleanup(func() { _ = srv.Close() })
	return New(path), f
}

func inspectOf(id, ip string, aliases ...string) map[string]any {
	return map[string]any{"Id": id, "Name": "/app-" + id[:4],
		"Config":          map[string]any{"Hostname": id[:12], "Labels": map[string]string{LabelComposeService: "web"}},
		"NetworkSettings": map[string]any{"Networks": map[string]any{proto.MeshNetwork: map[string]any{"IPAddress": ip, "Aliases": aliases}}}}
}

// withIPv6 gives an inspected container an IPv6 address on the mesh
// network (M8).
func withIPv6(c map[string]any, ip6 string) map[string]any {
	c["NetworkSettings"].(map[string]any)["Networks"].(map[string]any)[proto.MeshNetwork].(map[string]any)["GlobalIPv6Address"] = ip6
	return c
}

// Only GETs; a listed ID that is not one never reaches a path; a container
// gone since the listing, or not on the mesh network, is skipped; its IPv6
// address there comes with it.
func TestList(t *testing.T) {
	rt, f := serveRuntime(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/containers/json":
			_ = json.NewEncoder(w).Encode([]map[string]string{{"Id": id1}, {"Id": "../../images"}, {"Id": id2}, {"Id": id3}})
		case "/containers/" + id1 + "/json":
			_ = json.NewEncoder(w).Encode(withIPv6(inspectOf(id1, "10.77.128.129", "shop"), "fd10:77::8000:8001"))
		case "/containers/" + id2 + "/json":
			http.Error(w, `{"message":"No such container"}`, http.StatusNotFound)
		case "/containers/" + id3 + "/json":
			c := inspectOf(id3, "10.77.128.131")
			c["NetworkSettings"] = map[string]any{"Networks": map[string]any{"bridge": map[string]any{"IPAddress": "172.17.0.3"}}}
			_ = json.NewEncoder(w).Encode(c)
		default:
			http.NotFound(w, r)
		}
	})
	got, err := rt.List(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].ID != id1 || got[0].IP != netip.MustParseAddr("10.77.128.129") || got[0].Aliases[0] != "shop" ||
		got[0].IP6 != netip.MustParseAddr("fd10:77::8000:8001") ||
		got[0].Labels[LabelComposeService] != "web" || got[0].Name != "/app-aaaa" {
		t.Fatalf("List = %+v", got)
	}
	reqs := f.seen()
	if len(reqs) != 4 || reqs[0] != `GET /containers/json?filters=%7B%22network%22%3A%5B%22meshctl-mesh%22%5D%7D` {
		t.Fatalf("requests %v", reqs)
	}
	for _, r := range reqs {
		if !strings.HasPrefix(r, "GET ") || strings.Contains(r, "..") || strings.Contains(r, "images") {
			t.Errorf("request %q", r)
		}
	}
}

// A refusal, an answer too large or not JSON, and a dead socket are errors
// that say what was asked.
func TestListFailures(t *testing.T) {
	for name, c := range map[string]struct {
		h    http.HandlerFunc
		want string
	}{
		"refused": {func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "permission denied\nwhile trying", http.StatusForbidden)
		}, "GET /containers/json answered 403: permission denied while trying"},
		"too large": {func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte("[" + strings.Repeat(" ", maxList) + "]"))
		}, "more than"},
		"not JSON": {func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte("<html>")) }, "decode GET /containers/json"},
		"inspect fails": {func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/containers/json" {
				_ = json.NewEncoder(w).Encode([]map[string]string{{"Id": id1}})
				return
			}
			http.Error(w, "boom", http.StatusInternalServerError)
		}, "answered 500: boom"},
	} {
		t.Run(name, func(t *testing.T) {
			rt, _ := serveRuntime(t, c.h)
			if _, err := rt.List(context.Background()); err == nil || !strings.Contains(err.Error(), c.want) {
				t.Fatalf("List = %v, want %q", err, c.want)
			}
		})
	}
	dead := New(filepath.Join(t.TempDir(), "none.sock"))
	if _, err := dead.List(context.Background()); err == nil || !strings.Contains(err.Error(), "none.sock") {
		t.Fatalf("a dead socket = %v", err)
	}
	if err := dead.Watch(context.Background(), func() {}); err == nil {
		t.Fatal("Watch on a dead socket succeeded")
	}
}

// Every event line is a change; the stream's end is an error.
func TestWatch(t *testing.T) {
	rt, f := serveRuntime(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"Type":"container","Action":"start"}` + "\n\n" + `{"Type":"network","Action":"connect"}` + "\n"))
	})
	n := 0
	err := rt.Watch(context.Background(), func() { n++ })
	if err == nil || !strings.Contains(err.Error(), "ended") || n != 2 {
		t.Fatalf("Watch = %v after %d changes", err, n)
	}
	if reqs := f.seen(); len(reqs) != 1 || !strings.HasPrefix(reqs[0], "GET /events?filters=") || !strings.Contains(reqs[0], "network") {
		t.Fatalf("requests %v", reqs)
	}
	long, _ := serveRuntime(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(strings.Repeat("x", maxEventLine+1) + "\n"))
	})
	if err := long.Watch(context.Background(), func() {}); err == nil || !strings.Contains(err.Error(), "event stream:") {
		t.Fatalf("an overlong event = %v", err)
	}
	var se *StatusError
	refused, _ := serveRuntime(t, func(w http.ResponseWriter, r *http.Request) { http.Error(w, "no", http.StatusForbidden) })
	if err := refused.Watch(context.Background(), func() {}); !errors.As(err, &se) || se.Status != 403 {
		t.Fatalf("a refused stream = %v", err)
	}
}

// Detect takes the first path that is a socket.
func TestDetect(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("no container runtime socket on Windows; the agent finds none there")
	}
	dir := t.TempDir()
	file := filepath.Join(dir, "docker.sock")
	if err := os.WriteFile(file, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	rt, _ := serveRuntime(t, func(http.ResponseWriter, *http.Request) {})
	if got := Detect([]string{filepath.Join(dir, "none.sock"), file, rt.Socket()}); got != rt.Socket() {
		t.Fatalf("Detect = %q, want %q", got, rt.Socket())
	}
	if got := Detect([]string{file}); got != "" {
		t.Fatalf("Detect of a plain file = %q", got)
	}
	if len(Sockets()) != 2 || Sockets()[0] != "/var/run/docker.sock" {
		t.Fatalf("Sockets = %v", Sockets())
	}
}
