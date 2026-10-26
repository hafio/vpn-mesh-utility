//go:build linux

package wg

import (
	"errors"
	"log/slog"
	"os"
	"strings"
	"testing"
)

// Open opens the backend --wg-backend names and says which; auto settles on
// what DetectBackend finds. Opening takes no privilege, so this runs as any
// user; a host without /dev/net/tun has no userspace backend and says so.
func TestOpen(t *testing.T) {
	log := slog.New(slog.DiscardHandler)
	for _, b := range []string{BackendKernel, BackendUserspace, BackendAuto} {
		want := b
		if b == BackendAuto {
			want = DetectBackend()
		}
		h, got, err := Open(b, log)
		if err != nil {
			if want != BackendUserspace || h != nil || got != "" || !strings.Contains(err.Error(), tunPath) {
				t.Fatalf("Open(%q) = %v, %q, %v", b, h, got, err)
			}
			continue
		}
		if got != want {
			t.Errorf("Open(%q) opened %q, want %q", b, got, want)
		}
		_, isKernel := h.(*kernel)
		if _, isUser := h.(*userspace); isKernel != (got == BackendKernel) || isUser != (got == BackendUserspace) {
			t.Errorf("Open(%q) reports %q but opened a %T", b, got, h)
		}
		if err := h.Close(); err != nil {
			t.Errorf("Open(%q): Close: %v", b, err)
		}
	}
}

// Creating the interface takes CAP_NET_ADMIN. Without it either backend
// fails with the interface's name, a hint, and nothing left behind.
func TestCreateLinkNeedsPrivilege(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("as root this creates real interfaces; the e2e suite covers that")
	}
	k := &kernel{links: links{kind: "wireguard"}}
	if err := k.CreateLink("mctest-wg"); err == nil || !strings.Contains(err.Error(), "create WireGuard interface mctest-wg") ||
		!strings.Contains(err.Error(), "--wg-backend userspace") {
		t.Fatalf("kernel CreateLink without privilege = %v", err)
	}
	u := &userspace{links: links{kind: "tuntap"}, log: slog.New(slog.DiscardHandler)}
	if err := u.CreateLink("mctest-tun"); err == nil || !strings.Contains(err.Error(), "create userspace WireGuard interface mctest-tun") {
		t.Fatalf("userspace CreateLink without privilege = %v", err)
	}
	for _, name := range []string{"mctest-wg", "mctest-tun"} {
		if l, err := k.Link(name); err != nil || l.Exists {
			t.Errorf("%s was left behind: %+v, %v", name, l, err)
		}
	}
	if _, _, err := u.Device("mctest-tun"); err == nil {
		t.Error("a failed CreateLink left a device")
	}
}

// Removing an interface that is not there is nothing to do, and one the
// agent did not make (lo) is left alone, by either backend; neither takes
// privilege.
func TestDeleteLink(t *testing.T) {
	k := &kernel{links: links{kind: "wireguard"}}
	u := &userspace{links: links{kind: "tuntap"}, log: slog.New(slog.DiscardHandler)}
	for name, h := range map[string]Host{"kernel": k, "userspace": u} {
		if err := h.DeleteLink("mctest-none"); err != nil {
			t.Errorf("%s: DeleteLink of a missing interface = %v", name, err)
		}
		if err := h.DeleteLink("lo"); !errors.Is(err, ErrNotOurs) {
			t.Errorf("%s: DeleteLink(lo) = %v", name, err)
		}
	}
}
