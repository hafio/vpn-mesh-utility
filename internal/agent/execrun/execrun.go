// Package execrun is the agent's only way to run a program (CLAUDE.md
// security invariants). Commands never go through a shell; only allowlisted
// programs run; and every argv element is re-checked with validate.Arg right
// before the call, so a value that slipped past boundary validation still
// cannot smuggle in a metacharacter. TestModuleBoundaries fails if any other
// package under internal/ or cmd/ imports os/exec.
package execrun

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/hafio/vpn-mesh-utility/internal/validate"
)

// Runner runs one command and returns its standard output.
type Runner interface {
	Run(ctx context.Context, argv []string) (string, error)
}

// ErrNotInstalled means the program is not on PATH.
var ErrNotInstalled = errors.New("program is not installed")

// allowed is the closed set of programs the agent may run: firewall-cmd
// only (the install scripts set up the service, amendment A51).
func allowed(name string) bool { return name == "firewall-cmd" }

// Check is the shell-layer gate: an allowlisted program, and validate.Arg on
// every element.
func Check(argv []string) error {
	if len(argv) == 0 {
		return errors.New("execrun: empty command")
	}
	if !allowed(argv[0]) {
		return fmt.Errorf("execrun: %q is not a program the agent may run", argv[0])
	}
	for _, a := range argv[1:] {
		if err := validate.Arg(a); err != nil {
			return fmt.Errorf("execrun: refusing to run %s: %w", argv[0], err)
		}
	}
	return nil
}

// Exec runs commands on the host.
type Exec struct {
	// Timeout bounds each command; zero means 30 s.
	Timeout time.Duration
}

// maxOutput caps what is kept of a command's output.
const maxOutput = 64 << 10

// Run checks argv and runs it with a minimal environment.
func (e Exec) Run(ctx context.Context, argv []string) (string, error) {
	if err := Check(argv); err != nil {
		return "", err
	}
	path, err := exec.LookPath(argv[0])
	if err != nil {
		return "", fmt.Errorf("%s: %w", argv[0], ErrNotInstalled)
	}
	timeout := e.Timeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, path, argv[1:]...)
	cmd.Env = []string{"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin", "LC_ALL=C"}
	var stdout, stderr capped
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); err != nil {
		return stdout.String(), fmt.Errorf("%s: %w: %s", strings.Join(argv, " "), err, strings.TrimSpace(stderr.String()))
	}
	return stdout.String(), nil
}

// capped keeps the first maxOutput bytes and discards the rest.
type capped struct{ b bytes.Buffer }

func (c *capped) Write(p []byte) (int, error) {
	if room := maxOutput - c.b.Len(); room > 0 {
		if len(p) > room {
			c.b.Write(p[:room])
		} else {
			c.b.Write(p)
		}
	}
	return len(p), nil
}

func (c *capped) String() string { return c.b.String() }
