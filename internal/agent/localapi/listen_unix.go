//go:build !windows

package localapi

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/user"
	"strconv"
)

// Group may open the socket in addition to root, when it exists.
const Group = "meshctl"

// reachHint is what to check when the agent does not answer.
const reachHint = "is `meshctl agent run` running, and are you root or in group " + Group + "?"

// Listen opens the socket at path: mode 0660, group meshctl when that group
// exists (else root only), replacing a stale socket but refusing to start
// when another agent answers on it.
func Listen(path string) (net.Listener, error) {
	ln, err := listenSocket(path)
	if err != nil {
		return nil, err
	}
	mode := os.FileMode(0o600)
	if g, err := user.LookupGroup(Group); err == nil {
		if gid, err := strconv.Atoi(g.Gid); err == nil && os.Chown(path, -1, gid) == nil {
			mode = 0o660
		}
	}
	if err := os.Chmod(path, mode); err != nil {
		_ = ln.Close()
		return nil, fmt.Errorf("restrict %s: %w", path, err)
	}
	return ln, nil
}

func dial(ctx context.Context, path string) (net.Conn, error) {
	var d net.Dialer
	return d.DialContext(ctx, "unix", path)
}
