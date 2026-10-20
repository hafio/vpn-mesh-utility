// Package sysctl sets the kernel parameters a config asks for: ip_forward on
// relay and exit servers (amendment A4), src_valid_mark on exit-mode spokes,
// and the container network's forwarding and proxy ARP (amendment A37).
// It writes /proc/sys directly -- no sysctl binary, no /etc/sysctl.d -- and
// the agent re-asserts the values on every boot from its cached config.
package sysctl

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
}

// Apply sets each parameter whose current value differs and returns how many
// it changed. Keys outside the allowlist are refused before any path is
// built from them.
func (w *Writer) Apply(params []proto.Sysctl) (int, error) {
	n := 0
	for _, p := range params {
		if _, err := validate.Enum("sysctl key", p.Key, proto.SysctlKeys()...); err != nil {
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
				"in a container /proc/sys is read-only, so the compose file's sysctls must set it: render it again with `meshctl node compose`)",
				p.Key, p.Value, err)
		}
		n++
	}
	return n, nil
}
