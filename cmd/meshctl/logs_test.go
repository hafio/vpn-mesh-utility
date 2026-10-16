package main

import (
	"context"
	"io"
	"path/filepath"
	"testing"

	"github.com/hafio/vpn-mesh-utility/internal/agent/localapi"
	"github.com/hafio/vpn-mesh-utility/internal/agent/logbuf"
)

// `meshctl logs` prints what the agent's ring holds, as it holds it.
func TestLogsCommand(t *testing.T) {
	ring := logbuf.New(10)
	_, _ = io.WriteString(ring, "level=INFO msg=\"meshctl agent running\"\nlevel=WARN msg=\"event stream failed; will retry\"\n")
	sock := filepath.Join(shortDir(t), "agent.sock")
	ln, err := localapi.Listen(sock)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	go func() { _ = localapi.Serve(ctx, ln, localapi.Handler(localapi.Funcs{Logs: ring})) }()

	if code, out, _ := runArgs("logs", "--socket", sock); code != exitOK || out != "level=INFO msg=\"meshctl agent running\"\nlevel=WARN msg=\"event stream failed; will retry\"\n" {
		t.Fatalf("logs = %d %q", code, out)
	}
	if code, _, _ := runArgs("logs", "now", "--socket", sock); code != exitUsage {
		t.Fatalf("logs now = %d, want a usage error", code)
	}
	if code, _, errOut := runArgs("logs", "--socket", filepath.Join(shortDir(t), "none.sock")); code != exitFailure || errOut == "" {
		t.Fatalf("logs without an agent = %d %q", code, errOut)
	}
}
