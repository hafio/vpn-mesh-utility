//go:build e2e && linux

package e2e

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"
)

// agentLog is what the agent of data dir data logged.
func (l *lab) agentLog(data string) string {
	b, _ := os.ReadFile(filepath.Join(l.dir, "agent-"+data+".log"))
	return string(b)
}

// copyData copies data dir from into to, as a backup restored or a volume
// cloned would.
func (l *lab) copyData(from, to string) {
	l.t.Helper()
	l.sh("cp", "-a", l.data(from), l.data(to))
	_ = os.Remove(filepath.Join(l.data(to), "agent.sock")) // the original's socket file
}

// One node, one running copy (amendment A50), with real agents: a crash
// restart carries on; a copy on the same machine is refused while the
// original runs; a copy that came from another machine wipes itself; a new
// token cuts a running node off, which takes its tunnel down.
func TestOneRunningCopy(t *testing.T) {
	l := newLab(t)
	l.agentArgs = []string{"--status-every", "2s"}
	l.startServer()
	l.startMesh()

	// A crash, and the agent started again: it carries on with its
	// session, without waiting.
	l.kill("agent-b")
	l.startAgent("b")
	eventually(t, 20*time.Second, "vps-2 back after a crash", func() bool {
		st, err := l.status("b")
		return err == nil && st.AppliedVersion > 0 && len(st.Peers) == 2 && st.LastContact > 0
	})
	if log := l.agentLog("b"); strings.Contains(log, "another copy of this node holds its session") || strings.Contains(log, "let this node in") {
		t.Fatalf("a crash restart said hello or waited:\n%s", tail(log, 20))
	}

	// A copy of vps-2's data dir in another namespace of this machine,
	// whose ticket fell behind the original's: it is sent to hello, waits,
	// and is refused when the original answers the challenge. Its own
	// WireGuard cannot reach the mesh, so peers never roam to it.
	l.copyData("b", "b-copy")
	staleTicket(t, l.data("b-copy"))
	copyHost := host{"h", "vps-2", "192.0.2.12", ""}
	l.addHost(copyHost)
	for _, cmd := range []string{"add table inet t", "add chain inet t o { type filter hook output priority 0; }", "add rule inet t o udp dport 51820 drop"} {
		l.sh("ip", "netns", "exec", ns("h"), "nft", cmd)
	}
	l.startAgentAt("h", "b-copy")
	if code := l.exitCode("agent-b-copy", 60*time.Second); code != 3 {
		t.Fatalf("the copy exited %d, want 3:\n%s", code, tail(l.agentLog("b-copy"), 20))
	}
	if log := l.agentLog("b-copy"); !strings.Contains(log, "duplicate") {
		t.Fatalf("the copy's log:\n%s", tail(log, 20))
	}
	if _, err := os.Stat(filepath.Join(l.data("b-copy"), "state.json")); err != nil {
		t.Fatalf("the refused copy lost its enrollment: %v", err)
	}
	if l.has("h", "link", "show", "wg0") {
		t.Error("the refused copy left its interface")
	}
	if out, err := l.meshctl("", nil, "node", "show", "vps-2", "--data", l.data("server")); err != nil || !strings.Contains(out, "the same machine, was refused") {
		t.Errorf("node show vps-2: %v\n%s", err, out)
	}
	eventually(t, 20*time.Second, "the original vps-2 still in the mesh", func() bool {
		st, err := l.status("b")
		return err == nil && len(st.Peers) == 2 && l.ping("a", "10.77.0.2")
	})

	// A copy of vps-3's data dir that presents another machine: refused as
	// foreign, it wipes its data dir but why.
	l.copyData("c", "c-foreign")
	fake := filepath.Join(l.dir, "other-machine-id")
	if err := os.WriteFile(fake, []byte("0123456789abcdef0123456789abcdef\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	l.addHost(host{"i", "vps-3", "192.0.2.13", ""})
	args := []string{"agent", "run", "--data", l.data("c-foreign"), "--socket", filepath.Join(l.data("c-foreign"), "agent.sock"), "--no-firewall",
		"--log-level", "debug", "--runtime-socket", "off", "--machine-id-file", fake}
	l.start("agent-c-foreign", "i", []string{"SSL_CERT_FILE=" + l.ca}, l.bin, args...)
	if code := l.exitCode("agent-c-foreign", 30*time.Second); code != 3 {
		t.Fatalf("the foreign copy exited %d, want 3:\n%s", code, tail(l.agentLog("c-foreign"), 20))
	}
	entries, _ := os.ReadDir(l.data("c-foreign"))
	var left []string
	for _, e := range entries {
		left = append(left, e.Name())
	}
	slices.Sort(left)
	if !slices.Equal(left, []string{"agent.lock", "refused.json"}) {
		t.Fatalf("left in the foreign copy's data dir: %v", left)
	}
	if l.has("i", "link", "show", "wg0") {
		t.Error("the foreign copy brought up an interface")
	}

	// A new token for vps-3 while it runs: it is told it was replaced, takes
	// its tunnel down and stops.
	if out, err := l.meshctl("", nil, "node", "token", "vps-3", "--data", l.data("server")); err != nil {
		t.Fatalf("node token vps-3: %v\n%s", err, out)
	}
	if code := l.exitCode("agent-c", 30*time.Second); code != 3 {
		t.Fatalf("vps-3's agent exited %d after the token, want 3:\n%s", code, tail(l.agentLog("c"), 20))
	}
	if l.has("c", "link", "show", "wg0") {
		t.Error("the cut-off vps-3 left its interface")
	}
}

// staleTicket moves a data dir's saved ticket away from the control
// plane's, as the original's requests after a copy was taken do.
func staleTicket(t *testing.T, data string) {
	t.Helper()
	path := filepath.Join(data, "session.json")
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var s map[string]any
	if err := json.Unmarshal(b, &s); err != nil {
		t.Fatal(err)
	}
	s["ticket"] = strings.Repeat("0", 64)
	if b, err = json.Marshal(s); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, b, 0o600); err != nil {
		t.Fatal(err)
	}
}
