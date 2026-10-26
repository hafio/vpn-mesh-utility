package main

import (
	"context"
	"maps"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hafio/vpn-mesh-utility/internal/agent/localapi"
	"github.com/hafio/vpn-mesh-utility/internal/agent/state"
	"github.com/hafio/vpn-mesh-utility/internal/proto"
	"github.com/hafio/vpn-mesh-utility/internal/testfix"
)

// runEnv runs args with vars as the environment, as in the Linux image.
func runEnv(vars map[string]string, args ...string) (code int, stderr string) {
	e, _, errb := testEnv(vars)
	e.goos = "linux"
	code = run(context.Background(), e, args)
	return code, errb.String()
}

func with(base, more map[string]string) map[string]string {
	m := maps.Clone(base)
	maps.Copy(m, more)
	return m
}

// MESHCTL_MODE is control or node, a node when unset (amendment A48); the
// names before it are refused, not taken as aliases.
func TestStartMode(t *testing.T) {
	for m, want := range map[string]string{
		"client": "MESHCTL_MODE must be control or node",
		"server": "MESHCTL_MODE=server is gone: set control for the control plane",
		"agent":  "MESHCTL_MODE=agent is gone",
	} {
		if code, errOut := runEnv(map[string]string{"MESHCTL_MODE": m}, "start"); code != exitUsage || !strings.Contains(errOut, want) {
			t.Errorf("mode %q: code %d, %s", m, code, errOut)
		}
	}
	// Unset is a node: one that never joined says so, and that a control
	// plane sets the mode.
	code, errOut := runEnv(map[string]string{"MESHCTL_DATA": t.TempDir()}, "start")
	if code != exitFailure || !strings.Contains(errOut, "has no join token") || !strings.Contains(errOut, "MESHCTL_MODE=control") {
		t.Errorf("no mode: code %d, %s", code, errOut)
	}
	if code, _ := runEnv(map[string]string{"MESHCTL_MODE": modeControl}, "start", "extra"); code != exitUsage {
		t.Errorf("a stray argument: code %d", code)
	}
}

// The control plane's variables reach its flags, and an error names the
// variable it came from.
func TestStartControl(t *testing.T) {
	base := map[string]string{"MESHCTL_MODE": modeControl, "MESHCTL_DATA": t.TempDir(), "MESHCTL_URL": "https://mesh.example.com"}
	for _, c := range []struct {
		vars map[string]string
		want string
	}{
		{map[string]string{"MESHCTL_URL": ""}, "--url (or MESHCTL_URL) is required"},
		{map[string]string{"MESHCTL_PORT": "70000"}, "MESHCTL_PORT must be a port"},
		{map[string]string{"MESHCTL_PORT": "http"}, "MESHCTL_PORT must be a port"},
		{map[string]string{"MESHCTL_DATA": "/data?mode=memory"}, "MESHCTL_DATA"},
		{map[string]string{"MESHCTL_TRUSTED_PROXIES": "10.0.0.1/8"}, "(start sets --trusted-proxies from MESHCTL_TRUSTED_PROXIES)"},
		{map[string]string{"MESHCTL_HUB_DOWN_AFTER": "soon"}, "(start sets --hub-down-after from MESHCTL_HUB_DOWN_AFTER)"},
		{map[string]string{"MESHCTL_HUB_DOWN_AFTER": "2h"}, "must be 0 (off) or between 1s and 1h"},
		{map[string]string{"MESHCTL_LOG_LEVEL": "loud"}, "(start sets --log-level from MESHCTL_LOG_LEVEL)"},
	} {
		if code, errOut := runEnv(with(base, c.vars), "start"); code != exitUsage || !strings.Contains(errOut, c.want) {
			t.Errorf("%v: code %d, %s; want %q", c.vars, code, errOut, c.want)
		}
	}
}

// A node's join comes from the environment: a token or a token file,
// checked before anything is asked, and never shown back.
func TestStartNodeJoin(t *testing.T) {
	dir := t.TempDir()
	file := func(name, body string) string {
		p := filepath.Join(dir, name)
		if err := os.WriteFile(p, []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
		return p
	}
	good, big, bad := file("token", testfix.Token+"\n"), file("big", strings.Repeat("x", maxTokenFile+1)), file("bad", "not-a-token-secret")
	down := "http://127.0.0.1:1" // nothing listens: the join cannot be asked
	base := map[string]string{"MESHCTL_MODE": modeNode, "MESHCTL_DATA": t.TempDir()}
	for _, c := range []struct {
		vars map[string]string
		code int
		want string
	}{
		{nil, exitFailure, "has no join token"},
		{map[string]string{"MESHCTL_TOKEN": testfix.Token}, exitUsage, "MESHCTL_URL"},
		{map[string]string{"MESHCTL_TOKEN": testfix.Token, "MESHCTL_TOKEN_FILE": good}, exitUsage, "not both"},
		{map[string]string{"MESHCTL_TOKEN_FILE": "token"}, exitUsage, "must be an absolute path"},
		{map[string]string{"MESHCTL_TOKEN_FILE": filepath.Join(dir, "none")}, exitFailure, "read the join token from MESHCTL_TOKEN_FILE"},
		{map[string]string{"MESHCTL_TOKEN_FILE": big}, exitFailure, "more than 1024 bytes"},
		{map[string]string{"MESHCTL_TOKEN_FILE": bad}, exitUsage, "the token in MESHCTL_TOKEN_FILE"},
		{map[string]string{"MESHCTL_TOKEN": "short-secret"}, exitUsage, "MESHCTL_TOKEN"},
		{map[string]string{"MESHCTL_TOKEN": testfix.Token, "MESHCTL_URL": down, "MESHCTL_NAME": "NAS 1"}, exitUsage, "MESHCTL_NAME"},
		{map[string]string{"MESHCTL_LOG_LEVEL": "loud"}, exitUsage, "(MESHCTL_LOG_LEVEL)"},
		{map[string]string{"MESHCTL_MACHINE_ID_FILE": "/etc/machine-id?x"}, exitUsage, "(MESHCTL_MACHINE_ID_FILE)"},
		{map[string]string{"MESHCTL_TOKEN": testfix.Token, "MESHCTL_URL": down, "MESHCTL_NAME": "nas"}, exitFailure, "the next start tries again"},
		{map[string]string{"MESHCTL_TOKEN_FILE": good, "MESHCTL_URL": down}, exitFailure, "the next start tries again"},
	} {
		code, errOut := runEnv(with(base, c.vars), "start")
		if code != c.code || !strings.Contains(errOut, c.want) {
			t.Errorf("%v: code %d, %s; want %d %q", c.vars, code, errOut, c.code, c.want)
		}
		for _, secret := range []string{testfix.Token, "not-a-token-secret", "short-secret"} {
			if strings.Contains(errOut, secret) {
				t.Errorf("%v: the token was shown: %s", c.vars, errOut)
			}
		}
	}
}

// A node that has joined runs the agent with no token, the agent's
// variables reaching `agent run`'s flags.
func TestStartNodeRuns(t *testing.T) {
	data := t.TempDir()
	dir, err := state.Open(data)
	if err != nil {
		t.Fatal(err)
	}
	if err := dir.SaveEnrollment(state.Enrollment{Server: "https://mesh.example.com", NodeID: 4, Name: "nas", Role: proto.RoleSpoke,
		ConfigKey: testfix.ConfigKeyB64()}); err != nil {
		t.Fatal(err)
	}
	// MESHCTL_URL is the control plane's flag only: a node that has joined
	// runs without it.
	base := map[string]string{"MESHCTL_MODE": modeNode, "MESHCTL_DATA": data, "MESHCTL_URL": "https://mesh.example.com"}
	for name, val := range map[string]string{
		"MESHCTL_UI_LISTEN":      "0.0.0.0:7443",
		"MESHCTL_RUNTIME_SOCKET": "docker.sock",
	} {
		code, errOut := runEnv(with(base, map[string]string{name: val}), "start")
		if code != exitUsage || !strings.Contains(errOut, "from "+name+")") {
			t.Errorf("%s: code %d, %s", name, code, errOut)
		}
	}
}

// The image's health check finds the control plane's port and the agent's
// socket as start does, a node when the mode is unset.
func TestHealthcheckStart(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != proto.PathHealth {
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()
	_, port, _ := net.SplitHostPort(strings.TrimPrefix(srv.URL, "http://"))
	data := shortDir(t)
	ln, err := localapi.Listen(filepath.Join(data, "agent.sock"))
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		_ = localapi.Serve(ctx, ln, localapi.Handler(localapi.Funcs{Status: func(context.Context) (proto.LocalStatus, error) {
			return proto.LocalStatus{}, nil
		}}))
	}()
	for _, c := range []struct {
		vars map[string]string
		code int
	}{
		{map[string]string{"MESHCTL_MODE": modeControl, "MESHCTL_PORT": port}, exitOK},
		{map[string]string{"MESHCTL_MODE": modeControl, "MESHCTL_PORT": "1"}, exitFailure},
		{map[string]string{"MESHCTL_MODE": modeControl, "MESHCTL_PORT": "x"}, exitUsage},
		{map[string]string{"MESHCTL_MODE": modeNode, "MESHCTL_DATA": data}, exitOK},
		{map[string]string{"MESHCTL_DATA": data}, exitOK},
		{map[string]string{"MESHCTL_MODE": modeNode, "MESHCTL_DATA": shortDir(t)}, exitFailure},
		{map[string]string{"MESHCTL_MODE": modeNode, "MESHCTL_DATA": "/data?x"}, exitUsage},
		{map[string]string{"MESHCTL_MODE": "server"}, exitUsage},
	} {
		if code, errOut := runEnv(c.vars, "healthcheck", "start"); code != c.code {
			t.Errorf("%v: code %d, want %d: %s", c.vars, code, c.code, errOut)
		}
	}
	if code, _ := runEnv(nil, "healthcheck", "start", "extra"); code != exitUsage {
		t.Errorf("a stray argument: code %d", code)
	}
}

// Every variable start hands on names a flag of one of its modes'
// commands, so docs/commands.md can take the default and meaning from it.
func TestStartVarsNameRealFlags(t *testing.T) {
	if fromVar(nil, startVars()) != nil {
		t.Error("no error became one")
	}
	e, _, _ := testEnv(nil)
	seen := map[string]bool{}
	for _, v := range startVars() {
		if seen[v.name] || !strings.HasPrefix(v.name, "MESHCTL_") {
			t.Errorf("%s: twice, or not MESHCTL_", v.name)
		}
		seen[v.name] = true
		if v.flag == "" && v.meaning == "" {
			t.Errorf("%s has no meaning", v.name)
		}
		if f, _ := startFlag(e, v); v.flag != "" && f == nil {
			t.Errorf("%s names --%s, which none of its modes' commands has", v.name, v.flag)
		}
	}
}
