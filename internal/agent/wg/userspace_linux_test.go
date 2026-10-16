//go:build linux

package wg

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
)

// The userspace host before it has a device: what the agent sees when the
// interface is gone or belongs to another process. Creating the tun needs
// root, so that half is the e2e suite's (TestSpokesAndExit's NAS).
func TestUserspaceWithoutADevice(t *testing.T) {
	var logs bytes.Buffer
	u := &userspace{links: links{kind: "tuntap"}, log: slog.New(slog.NewTextHandler(&logs, nil))}
	if l, err := u.Link("mctest-none"); err != nil || l.Exists {
		t.Fatalf("Link of a missing interface = %+v, %v", l, err)
	}
	if l, err := u.Link("lo"); err != nil || !l.Exists || l.Ours {
		t.Fatalf("Link(lo) = %+v, %v: an interface this process did not open is never ours", l, err)
	}
	if _, _, err := u.Device("wg0"); err == nil || !strings.Contains(err.Error(), "not a WireGuard interface of this agent") {
		t.Fatalf("Device = %v", err)
	}
	if err := u.Configure("wg0", Change{}); err == nil {
		t.Fatal("Configure without a device succeeded")
	}
	if err := u.Close(); err != nil {
		t.Fatal(err)
	}

	// wireguard-go's log: verbose lines only at debug, errors as warnings.
	log := u.logger("wg0")
	log.Verbosef("peer(%s) - Sending keepalive packet", "AAAA")
	log.Errorf("peer(%s) - Failed to send handshake initiation: %v", "AAAA", "network is unreachable")
	got := logs.String()
	if strings.Contains(got, "keepalive") || !strings.Contains(got, "level=WARN") || !strings.Contains(got, "network is unreachable") ||
		!strings.Contains(got, "wireguard=wg0") {
		t.Fatalf("log:\n%s", got)
	}
}
