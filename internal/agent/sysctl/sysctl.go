// Package sysctl sets the kernel parameters a config asks for: ip_forward on
// relay and exit servers (amendment A4), src_valid_mark on exit-mode spokes,
// the container network's forwarding and proxy ARP (amendment A37), and
// IPv6 forwarding and proxy NDP (M8).
// It writes /proc/sys directly -- no sysctl binary, no /etc/sysctl.d -- and
// the agent re-asserts the values on every boot from its cached config.
package sysctl

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/hafio/vpn-mesh-utility/internal/proto"
	"github.com/hafio/vpn-mesh-utility/internal/validate"
)

// Writer applies parameters under Root.
type Writer struct {
	// Root is /proc/sys; tests point it at a directory of their own.
	Root string
	// NoProc is a host without /proc/sys (Windows), where exit mode routes
	// by table, so src_valid_mark has nothing to do, and no node relays.
	NoProc bool
	// Slot is the agent's mesh's slot, whose interfaces' keys are allowed
	// (amendment A49).
	Slot proto.Slot
}

// Apply sets each parameter whose current value differs and returns how many
// it changed. Keys outside the allowlist are refused before any path is
// built from them.
func (w *Writer) Apply(params []proto.Sysctl) (int, error) {
	n := 0
	for _, p := range params {
		if _, err := validate.Enum("sysctl key", p.Key, proto.SysctlKeys(w.Slot)...); err != nil {
			return n, err
		}
		if _, err := validate.Enum("sysctl value", p.Value, "0", "1"); err != nil {
			return n, err
		}
		switch {
		case w.NoProc && p.Key == proto.SysctlSrcValidMark:
			continue
		case w.NoProc:
			return n, fmt.Errorf("set %s = %s: this host has no %s; only a Linux server relays or is an exit", p.Key, p.Value, p.Key)
		}
		if p.Key == proto.SysctlIPv6Forward && p.Value == "1" {
			k, err := w.keepRAs()
			n += k
			if err != nil {
				return n, err
			}
		}
		path := filepath.Join(w.Root, filepath.FromSlash(strings.ReplaceAll(p.Key, ".", "/")))
		cur, err := os.ReadFile(path)
		if err != nil {
			return n, fmt.Errorf("read %s: %w", p.Key, err)
		}
		if strings.TrimSpace(string(cur)) == p.Value {
			continue
		}
		if err := os.WriteFile(path, []byte(p.Value+"\n"), 0o644); err != nil {
			return n, fmt.Errorf("set %s = %s: %w (on a host the agent needs root, and ProtectKernelTunables must stay off; "+
				"in a container /proc/sys is read-only, so the container's sysctls must set it: docs/containers.md lists them)",
				p.Key, p.Value, err)
		}
		n++
	}
	return n, nil
}

// keepRAs lets the interfaces that take router advertisements keep taking
// them once IPv6 forwarding is on (M8): Linux ignores RAs on a forwarding
// host unless accept_ra is 2, so an uplink whose IPv6 default route comes
// from RAs (accept_ra 1) would lose it. Read-only interfaces (a
// container's /proc/sys) are not the agent's to change.
func (w *Writer) keepRAs() (int, error) {
	conf := filepath.Join(w.Root, "net", "ipv6", "conf")
	ifs, err := os.ReadDir(conf)
	if err != nil {
		return 0, fmt.Errorf("list the IPv6 interfaces in %s: %w", conf, err)
	}
	n := 0
	for _, ifc := range ifs {
		name := ifc.Name()
		if name == "all" || name == "default" || name == "lo" {
			continue
		}
		path := filepath.Join(conf, name, "accept_ra")
		if cur, err := os.ReadFile(path); err != nil || strings.TrimSpace(string(cur)) != "1" {
			continue // gone since the listing, or not taking RAs the forwarding would stop
		}
		err := os.WriteFile(path, []byte("2\n"), 0o644)
		switch {
		case errors.Is(err, syscall.EROFS):
			continue
		case err != nil:
			return n, fmt.Errorf("set net.ipv6.conf.%s.accept_ra = 2 before turning IPv6 forwarding on, so %s keeps its router advertisements: %w", name, name, err)
		}
		n++
	}
	return n, nil
}

// IPv6 reports whether the host can use IPv6 (M8): it has the IPv6 stack,
// and IPv6 is disabled neither on all interfaces nor on new ones (where
// the mesh interface would be). A host without /proc/sys (Windows) is
// taken to.
func (w *Writer) IPv6() bool {
	if w.NoProc {
		return true
	}
	for _, scope := range []string{"all", "default"} {
		b, err := os.ReadFile(filepath.Join(w.Root, "net", "ipv6", "conf", scope, "disable_ipv6"))
		if err != nil || strings.TrimSpace(string(b)) != "0" {
			return false
		}
	}
	return true
}
