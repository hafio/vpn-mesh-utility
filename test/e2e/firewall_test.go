//go:build e2e && linux

package e2e

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// alive reports whether the stream still echoes, without failing the test.
func (s *stream) alive(msg string) bool {
	_ = s.c.SetDeadline(time.Now().Add(2 * time.Second))
	if _, err := fmt.Fprintln(s.c, msg); err != nil {
		return false
	}
	line, err := s.rd.ReadString('\n')
	return err == nil && strings.TrimSpace(line) == msg
}

// connects reports whether a new TCP connection from space reaches addr.
func (l *lab) connects(space, addr string) bool {
	err := inNS(space, func() error {
		c, err := net.DialTimeout("tcp", addr, 2*time.Second)
		if err == nil {
			_ = c.Close()
		}
		return err
	})
	return err == nil
}

// waitOrDump waits for cond like eventually, and on a timeout fails with
// what the relay's kernel holds: the mesh firewall's table and the rest.
func (l *lab) waitOrDump(within time.Duration, what, space string, cond func() bool) {
	l.t.Helper()
	for deadline := time.Now().Add(within); time.Now().Before(deadline); time.Sleep(250 * time.Millisecond) {
		if cond() {
			return
		}
	}
	table, _ := exec.Command("ip", "netns", "exec", ns(space), "nft", "-a", "list", "table", "inet", "meshctl").CombinedOutput()
	tables, _ := exec.Command("ip", "netns", "exec", ns(space), "nft", "list", "tables").CombinedOutput()
	l.t.Fatalf("not within %s: %s\n--- %s: nft list table inet meshctl ---\n%s\n--- nft list tables ---\n%s", within, what, space, table, tables)
}

func (l *lab) applyFirewall(doc string) {
	l.t.Helper()
	file := filepath.Join(l.dir, "firewall.yaml")
	if err := os.WriteFile(file, []byte(doc), 0o600); err != nil {
		l.t.Fatal(err)
	}
	if out, err := l.meshctl("", nil, "firewall", "apply", file, "--data", l.data("server")); err != nil {
		l.t.Fatalf("firewall apply: %v\n%s", err, out)
	}
}

// The mesh firewall (amendment A43) in the kernel: every server holds the
// table with allow-all; a policy denying the laptop TCP to the NAS stops a
// connection that is open (the relay judges it again) and new ones, while
// ping still crosses; allowing again lets new connections through. The
// same deny with a schedule refuses inside its window only.
func TestMeshFirewall(t *testing.T) {
	l := newLab(t)
	l.addHost(laptop)
	l.addHost(nas)
	l.startServer()
	l.startMesh()
	l.enroll(laptop, "--role", "spoke", "--hub", "vps-1")
	l.enroll(nas, "--role", "spoke", "--hub", "vps-3", "--always-on")
	eventually(t, 60*time.Second, "both spokes to reach each other across their hubs", func() bool {
		return l.ping(laptop.ns, nas.meshIP) && l.ping(nas.ns, laptop.meshIP)
	})
	for _, n := range nodes {
		if out := l.sh("ip", "netns", "exec", ns(n.ns), "nft", "list", "table", "inet", "meshctl"); !strings.Contains(out, `comment "meshctl policy `) {
			t.Fatalf("%s holds no mesh firewall table:\n%s", n.name, out)
		}
	}
	s := l.openStream(laptop.ns, nas.ns, nas.meshIP)
	s.check(t, "before")

	l.applyFirewall(`segmentation:
  policies:
    - name: no-nas-tcp
      from: [node:laptop]
      to: [node:nas]
      services: [tcp]
      action: deny
    - name: allow-all
      action: allow
`)
	l.waitOrDump(20*time.Second, "a new laptop-NAS connection to be refused", "a", func() bool { return !l.connects(laptop.ns, nas.meshIP+":7000") })
	l.waitOrDump(10*time.Second, "the open laptop-NAS connection to stop", "a", func() bool { return !s.alive("after") })
	if !l.ping(laptop.ns, nas.meshIP) {
		t.Fatal("ping no longer crosses the mesh (a system rule)")
	}
	if out, err := l.meshctl("", nil, "firewall", "check", "laptop", "nas", "tcp/7000", "--data", l.data("server")); err != nil ||
		!strings.Contains(out, "dropped (rule 4, policy: no-nas-tcp)") {
		t.Fatalf("firewall check: %v\n%s", err, out)
	}
	if st, err := l.status("a"); err != nil || len(st.Errors) != 0 {
		t.Fatalf("vps-1 after the new policy: %+v, %v", st, err)
	}

	l.applyFirewall(`segmentation:
  policies:
    - name: allow-all
      action: allow
`)
	eventually(t, 20*time.Second, "new connections to the NAS again", func() bool { return l.connects(laptop.ns, nas.meshIP+":7000") })

	// A schedule (amendment A44): the deny is in force inside its window
	// and not outside it, read in UTC whatever the hosts' zones.
	now := time.Now().UTC()
	scheduled := func(from, to time.Time) string {
		return fmt.Sprintf(`segmentation:
  policies:
    - name: no-nas-tcp
      from: [node:laptop]
      to: [node:nas]
      services: [tcp]
      action: deny
      schedule:
        start: "%s"
        end: "%s"
    - name: allow-all
      action: allow
  timezone: UTC
`, from.Format("15:04"), to.Format("15:04"))
	}
	l.applyFirewall(scheduled(now.Add(-time.Hour), now.Add(time.Hour)))
	l.waitOrDump(20*time.Second, "a new laptop-NAS connection to be refused inside the window", "a", func() bool {
		return !l.connects(laptop.ns, nas.meshIP+":7000")
	})
	l.applyFirewall(scheduled(now.Add(time.Hour), now.Add(2*time.Hour)))
	eventually(t, 20*time.Second, "new connections to the NAS outside the window", func() bool { return l.connects(laptop.ns, nas.meshIP+":7000") })
}
