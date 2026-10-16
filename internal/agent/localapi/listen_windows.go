//go:build windows

package localapi

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/Microsoft/go-winio"

	"github.com/hafio/vpn-mesh-utility/internal/validate"
)

// PipeSDDL lets only Administrators and SYSTEM open the local API pipe
// (amendment A7): a protected DACL with nothing inherited.
const PipeSDDL = "D:P(A;;GA;;;BA)(A;;GA;;;SY)"

// reachHint is what to check when the agent does not answer.
const reachHint = "is the meshctl-agent service running, and is this prompt elevated (Run as administrator)?"

// Listen opens the local API at path: a named pipe (\\.\pipe\...) only
// Administrators and SYSTEM can open, refusing to start when another agent
// answers on it. A file path is a Unix socket, which tests use.
func Listen(path string) (net.Listener, error) {
	if !isPipe(path) {
		return listenSocket(path)
	}
	return listenPipe(path, PipeSDDL)
}

func listenPipe(path, sddl string) (net.Listener, error) {
	if _, err := validate.PipePath("--socket", path); err != nil {
		return nil, err
	}
	wait := time.Second
	if c, err := winio.DialPipe(path, &wait); err == nil {
		_ = c.Close()
		return nil, fmt.Errorf("another meshctl agent is already running (it answers on %s)", path)
	}
	ln, err := winio.ListenPipe(path, &winio.PipeConfig{SecurityDescriptor: sddl})
	if err != nil {
		return nil, fmt.Errorf("listen on %s: %w", path, err)
	}
	return ln, nil
}

func isPipe(path string) bool { return strings.HasPrefix(path, `\\.\pipe\`) }

func dial(ctx context.Context, path string) (net.Conn, error) {
	if isPipe(path) {
		return winio.DialPipeContext(ctx, path)
	}
	var d net.Dialer
	return d.DialContext(ctx, "unix", path)
}
