//go:build e2e && linux

package e2e

import (
	"fmt"
	"net"
	"os"
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
// ping still crosses; allowing again lets new connections through.
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
		if out := l.sh("ip", "netns", "exec", ns(n.ns), "nft", "list", "table", "ip", "meshctl"); !strings.Contains(out, `comment "meshctl policy `) {
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
	eventually(t, 20*time.Second, "the open connection to stop and a new one to be refused", func() bool {
		return !s.alive("after") && !l.connects(laptop.ns, nas.meshIP+":7000")
	})
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
}
