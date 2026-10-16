package main

import (
	"context"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/hafio/vpn-mesh-utility/internal/agent/localapi"
	"github.com/hafio/vpn-mesh-utility/internal/proto"
)

// fakeAgent is the local API of a laptop agent: a prefs change becomes the
// applied config once applyNext says so.
type fakeAgent struct {
	mu        sync.Mutex
	st        proto.LocalStatus
	requests  []proto.PrefsRequest
	applyNext bool
	dnsCalls  int
}

func (a *fakeAgent) status(context.Context) (proto.LocalStatus, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.st, nil
}

func (a *fakeAgent) prefs(_ context.Context, req proto.PrefsRequest) (proto.PrefsResponse, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.requests = append(a.requests, req)
	if *req.Exit == "vps-1" {
		return proto.PrefsResponse{}, &localapi.Refusal{Status: http.StatusConflict, Code: proto.CodeConflict, Message: `exit "vps-1" is not exit-capable; enable can-exit on it first`}
	}
	hash := strings.Repeat("b", 64)
	if a.applyNext {
		a.st.AppliedHash, a.st.Exit = hash, *req.Exit
	} else {
		a.st.Errors = []string{"apply step exit: operation not permitted"}
	}
	return proto.PrefsResponse{Version: 12, ConfigVersion: 12, ConfigHash: hash}, nil
}

// dns turns DNS on or off the way an agent on a host with systemd-resolved
// does.
func (a *fakeAgent) dns(_ context.Context, on bool) (proto.LocalStatus, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.st.DNS = proto.DNSOff
	if on {
		a.st.DNS = proto.DNSOn
	}
	a.dnsCalls++
	return a.st, nil
}

func serveAgent(t *testing.T, a *fakeAgent) string {
	t.Helper()
	sock := filepath.Join(shortDir(t), "agent.sock")
	ln, err := localapi.Listen(sock)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	go func() {
		_ = localapi.Serve(ctx, ln, localapi.Handler(localapi.Funcs{Status: a.status, Prefs: a.prefs, DNS: a.dns}))
	}()
	return sock
}

func TestExitCommand(t *testing.T) {
	a := &fakeAgent{applyNext: true, st: proto.LocalStatus{Node: proto.Self{Name: "laptop", Role: proto.RoleSpoke},
		AppliedHash: strings.Repeat("a", 64), ExitChoices: []string{"vps-2", "vps-3"}}}
	sock := serveAgent(t, a)

	if code, out, _ := runArgs("exit", "--socket", sock); code != exitOK || out != "exit off\nchoices: vps-2, vps-3\n" {
		t.Fatalf("exit = %d %q", code, out)
	}
	if code, out, errOut := runArgs("exit", "vps-2", "--socket", sock); code != exitOK || out != "exit via vps-2: config v12 applied\n" {
		t.Fatalf("exit vps-2 = %d %q %q", code, out, errOut)
	}
	if code, out, _ := runArgs("exit", "--socket", sock); code != exitOK || !strings.HasPrefix(out, "exit via vps-2\n") {
		t.Fatalf("exit after the switch = %d %q", code, out)
	}
	if code, _, errOut := runArgs("exit", "vps-1", "--socket", sock); code != exitFailure || !strings.Contains(errOut, "can-exit") {
		t.Fatalf("a refused exit = %d %q", code, errOut)
	}
	if code, out, _ := runArgs("exit", "off", "--socket", sock, "--wait", "0"); code != exitOK || !strings.Contains(out, "exit off: the server made config v12") {
		t.Fatalf("exit off --wait 0 = %d %q", code, out)
	}
	if got := a.requests[len(a.requests)-1]; got.Exit == nil || *got.Exit != "" {
		t.Fatalf("off must send an empty exit, sent %+v", got)
	}

	// Usage errors never reach the agent.
	before := len(a.requests)
	for _, args := range [][]string{{"exit", "vps-2", "vps-3"}, {"exit", "vps-2;reboot"}} {
		if code, _, _ := runArgs(append(args, "--socket", sock)...); code != exitUsage {
			t.Errorf("%q = %d, want a usage error", args, code)
		}
	}
	if len(a.requests) != before {
		t.Fatal("a usage error reached the agent")
	}
}

func TestExitReportsAnAgentThatDoesNotApply(t *testing.T) {
	a := &fakeAgent{st: proto.LocalStatus{AppliedHash: strings.Repeat("a", 64)}}
	sock := serveAgent(t, a)
	code, _, errOut := runArgs("exit", "vps-2", "--socket", sock, "--wait", "300ms")
	if code != exitFailure || !strings.Contains(errOut, "not applied it within 300ms") || !strings.Contains(errOut, "operation not permitted") {
		t.Fatalf("exit = %d %q", code, errOut)
	}
	if code, out, _ := runArgs("exit", "--socket", sock); code != exitOK || !strings.Contains(out, "no server of this mesh can be an exit") {
		t.Fatalf("exit without choices = %d %q", code, out)
	}
}

// `meshctl dns` shows the state, `dns on|off` switches it through the agent;
// anything else is a usage error that never reaches it.
func TestDNSCommand(t *testing.T) {
	a := &fakeAgent{st: proto.LocalStatus{Node: proto.Self{Name: "laptop", Role: proto.RoleSpoke}, MeshIP: "10.77.2.1", DNS: proto.DNSUnavailable}}
	sock := serveAgent(t, a)
	if code, out, _ := runArgs("dns", "--socket", sock); code != exitOK || !strings.HasPrefix(out, "dns unavailable: the resolver answers .mesh on 10.77.2.1") {
		t.Fatalf("dns = %d %q", code, out)
	}
	if code, out, _ := runArgs("dns", "off", "--socket", sock); code != exitOK || !strings.HasPrefix(out, "dns off on this node") {
		t.Fatalf("dns off = %d %q", code, out)
	}
	if code, out, _ := runArgs("dns", "on", "--socket", sock); code != exitOK || out != "dns on: this host sends .mesh lookups to the resolver on 10.77.2.1\n" {
		t.Fatalf("dns on = %d %q", code, out)
	}
	for _, args := range [][]string{{"dns", "yes"}, {"dns", "on", "off"}} {
		if code, _, _ := runArgs(append(args, "--socket", sock)...); code != exitUsage {
			t.Errorf("%q = %d, want a usage error", args, code)
		}
	}
	if a.dnsCalls != 2 {
		t.Fatalf("the agent was asked %d times, want 2", a.dnsCalls)
	}
	a.st.DNS = ""
	if code, out, _ := runArgs("dns", "--socket", sock); code != exitOK || !strings.Contains(out, "no DNS in its config") {
		t.Fatalf("dns before a config = %d %q", code, out)
	}
	if code, _, _ := runArgs("dns", "on", "--socket", filepath.Join(shortDir(t), "none.sock")); code != exitFailure {
		t.Fatalf("dns without an agent = %d", code)
	}
}
