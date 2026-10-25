// Command meshctl is the one binary of the mesh: the server, the agent on
// every node, and the CLI (spec S5, S11).
//
// main only wires the process to run(); every subcommand lives in the
// registry in commands.go and receives its dependencies through env, so tests
// drive the whole command tree without touching os.Args, stdio or the clock.
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"runtime"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/hafio/vpn-mesh-utility/internal/agent"
)

// version is stamped at build time with -ldflags "-X main.version=<tag>" by
// scripts/dev.sh and the Dockerfile. It is read once, in main, and handed to
// everything else through env.
var version = "dev"

// Exit codes.
const (
	exitOK      = 0
	exitFailure = 1
	exitUsage   = 2
	// exitRefused is an agent the control plane refused (amendment A50);
	// the systemd unit does not restart it.
	exitRefused = 3
)

// env is what every command gets instead of reaching for globals. goos and
// getenv are seams: docs/commands.md is rendered with goos "linux" and an
// empty environment, so it is the same on every machine.
type env struct {
	version string
	stdin   io.Reader
	stdout  io.Writer
	stderr  io.Writer
	goos    string
	getenv  func(string) string
	now     func() time.Time
	// password reads a password without echo (a terminal only).
	password func(prompt string) (string, error)
}

// usageError marks a mistake in how the command was invoked (exit 2).
type usageError struct{ msg string }

func (u *usageError) Error() string { return u.msg }

func usagef(format string, args ...any) error {
	return &usageError{msg: fmt.Sprintf(format, args...)}
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	e := &env{version: version, stdin: os.Stdin, stdout: os.Stdout, stderr: os.Stderr,
		goos: runtime.GOOS, getenv: os.Getenv, now: time.Now, password: ttyPassword(os.Stdin, os.Stderr)}
	code := run(ctx, e, os.Args[1:])
	stop()
	os.Exit(code)
}

// run dispatches args down the command tree and maps the result to an exit
// code.
func run(ctx context.Context, e *env, args []string) int {
	list := commands()
	if len(args) == 0 || args[0] == "help" || isHelp(args[0]) {
		printUsage(e.stdout, "meshctl", list)
		return exitOK
	}
	if args[0] == "--version" {
		args = append([]string{"version"}, args[1:]...)
	}
	path := "meshctl"
	for {
		c, ok := find(list, args[0])
		if !ok {
			fmt.Fprintf(e.stderr, "%s: unknown command %q\n\n", path, args[0])
			printUsage(e.stderr, path, list)
			return exitUsage
		}
		path += " " + c.name
		args = args[1:]
		if c.setup != nil {
			return runLeaf(ctx, e, path, c, args)
		}
		switch {
		case len(args) == 0:
			fmt.Fprintf(e.stderr, "%s: missing subcommand\n\n", path)
			printUsage(e.stderr, path, c.sub)
			return exitUsage
		case isHelp(args[0]):
			printUsage(e.stdout, path, c.sub)
			return exitOK
		}
		list = c.sub
	}
}

func runLeaf(ctx context.Context, e *env, path string, c command, args []string) int {
	fs := newFlags(path, c, e.stderr)
	act := c.setup(fs, e)
	pos, err := parseArgs(fs, args)
	if err == nil {
		err = act(ctx, pos)
	}
	if err == nil || errors.Is(err, errHelp) {
		return exitOK
	}
	fmt.Fprintf(e.stderr, "%s: %v\n", path, err)
	var u *usageError
	switch {
	case errors.As(err, &u):
		return exitUsage
	case errors.Is(err, agent.ErrRefused):
		return exitRefused
	}
	return exitFailure
}

func isHelp(s string) bool { return s == "-h" || s == "--help" || s == "-help" }

func find(list []command, name string) (command, bool) {
	for _, c := range list {
		if c.name == name {
			return c, true
		}
	}
	return command{}, false
}

func printUsage(w io.Writer, path string, cmds []command) {
	names := make([]string, 0, len(cmds))
	byName := map[string]command{}
	width := 0
	for _, c := range cmds {
		names = append(names, c.name)
		byName[c.name] = c
		width = max(width, len(c.name))
	}
	sort.Strings(names)
	var b strings.Builder
	fmt.Fprintf(&b, "usage: %s <command> [flags]\n\ncommands:\n", path)
	for _, n := range names {
		fmt.Fprintf(&b, "  %-*s  %s\n", width, n, byName[n].summary)
	}
	fmt.Fprintf(&b, "\nRun `%s <command> -h` for the flags of one command.\n", path)
	_, _ = io.WriteString(w, b.String()) // nowhere left to report a failed write
}
