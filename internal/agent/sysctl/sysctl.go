// Package sysctl sets the kernel parameters a config asks for: ip_forward on
// relay and exit servers (amendment A4), src_valid_mark on exit-mode spokes.
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
}

// New returns the host's writer.
func New() *Writer { return &Writer{Root: "/proc/sys"} }

// Apply sets each parameter whose current value differs and returns how many
// it changed. Keys outside the allowlist are refused before any path is
// built from them.
func (w *Writer) Apply(params []proto.Sysctl) (int, error) {
	n := 0
	for _, p := range params {
		if _, err := validate.Enum("sysctl key", p.Key, proto.SysctlIPForward, proto.SysctlSrcValidMark); err != nil {
			return n, err
		}
		if _, err := validate.Enum("sysctl value", p.Value, "0", "1"); err != nil {
			return n, err
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
			return n, fmt.Errorf("set %s = %s: %w (the agent needs root, and ProtectKernelTunables must stay off)", p.Key, p.Value, err)
		}
		n++
	}
	return n, nil
}
