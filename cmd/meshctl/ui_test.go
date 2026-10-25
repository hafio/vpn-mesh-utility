package main

import (
	"context"
	"net/http"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/hafio/vpn-mesh-utility/internal/agent/localapi"
	"github.com/hafio/vpn-mesh-utility/internal/proto"
	"github.com/hafio/vpn-mesh-utility/internal/testfix"
)

// `meshctl ui` prints the link alone on stdout and how to use it on
// stderr; the agent's reason for having no link is the error.
func TestUICommand(t *testing.T) {
	link := "http://127.0.0.1:7443/login?t=" + strings.Repeat("a", 43)
	var (
		refuse atomic.Bool
		served atomic.Value
	)
	served.Store(link)
	sock := filepath.Join(shortDir(t), "agent.sock")
	ln, err := localapi.Listen(sock)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	go func() {
		_ = localapi.Serve(ctx, ln, localapi.Handler(localapi.Funcs{UIToken: func(context.Context) (proto.UIToken, error) {
			if refuse.Load() {
				return proto.UIToken{}, &localapi.Refusal{Status: http.StatusConflict, Code: proto.CodeConflict, Message: "the local UI is off on this node"}
			}
			return proto.UIToken{URL: served.Load().(string), ExpiresAt: testfix.Epoch.Unix() + 120}, nil
		}}))
	}()

	if code, out, errOut := runArgs("ui", "--socket", sock); code != exitOK || out != link+"\n" || !strings.Contains(errOut, "within 2m0s; it works once") ||
		!strings.Contains(errOut, "ssh -L 7443:127.0.0.1:7443") {
		t.Fatalf("ui = %d %q %q", code, out, errOut)
	}
	refuse.Store(true)
	if code, out, errOut := runArgs("ui", "--socket", sock); code != exitFailure || out != "" || !strings.Contains(errOut, "is off") {
		t.Fatalf("ui with the UI off = %d %q %q", code, out, errOut)
	}
	// The agent of a mesh in slot 1 (amendment A49) is on its slot's port.
	refuse.Store(false)
	served.Store("http://127.0.0.1:7444/login?t=" + strings.Repeat("a", 43))
	if code, _, errOut := runArgs("ui", "--socket", sock); code != exitOK || !strings.Contains(errOut, "ssh -L 7444:127.0.0.1:7444") {
		t.Fatalf("ui on slot 1's port = %d %q", code, errOut)
	}
	if code, _, _ := runArgs("ui", "now", "--socket", sock); code != exitUsage {
		t.Fatalf("ui now = %d, want a usage error", code)
	}
}

// `agent run --ui-listen` takes a loopback address, auto (the mesh's
// slot's port) or off, and refuses anything else before the agent starts.
func TestAgentRunUIListen(t *testing.T) {
	data := filepath.Join(shortDir(t), "data")
	for _, bad := range []string{"0.0.0.0:7443", "10.77.0.1:7443", "127.0.0.1"} {
		if code, _, errOut := runArgs("agent", "run", "--data", data, "--ui-listen", bad); code != exitUsage || !strings.Contains(errOut, "--ui-listen") {
			t.Errorf("--ui-listen %s = %d %q", bad, code, errOut)
		}
	}
	for _, ok := range []string{"off", "auto"} {
		if code, _, errOut := runArgs("agent", "run", "--data", data, "--ui-listen", ok); code != exitFailure || !strings.Contains(errOut, "meshctl agent join") {
			t.Fatalf("--ui-listen %s before a join = %d %q", ok, code, errOut)
		}
	}
}
