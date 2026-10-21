package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/netip"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// command is one node of the command tree: a group (sub set, setup nil) or a
// leaf whose setup registers its flags and returns what runs once they are
// parsed. docs/commands.md is rendered from this tree (TestCommandsDoc).
type command struct {
	name string
	// args is the synopsis of the positional arguments ("<name>").
	args    string
	summary string
	sub     []command
	setup   func(fs *flag.FlagSet, e *env) action
}

// action runs a leaf command with its positional arguments.
type action func(ctx context.Context, args []string) error

// commands is the registry. Adding a command is one entry here (or in a
// group's list) plus its file.
func commands() []command {
	return []command{
		{name: "agent", summary: "join this node to a mesh, or run the agent", sub: agentCommands()},
		{name: "dns", args: "[on|off]", summary: "show or switch this node's split DNS for the mesh zone's names (stays on the node)", setup: setupDNS},
		{name: "exit", args: "[<server>|off]", summary: "show or switch this spoke's internet exit (through the local agent)", setup: setupExit},
		{name: "export", summary: "write the mesh's desired state as YAML (reads the database directly)", setup: setupExport},
		{name: "healthcheck", summary: "exit 0 when the local server or agent answers", sub: healthCommands()},
		{name: "import", args: "<file>", summary: "replace the desired state with an export (only while the server is stopped)", setup: setupImport},
		{name: "logs", summary: "print this node's recent agent log; -f keeps following it", setup: setupLogs},
		{name: "node", summary: "add, change and remove nodes on the running server", sub: nodeCommands()},
		{name: "server", summary: "run the control-plane server", setup: setupServer},
		{name: "status", summary: "show this node's agent: applied config and peers", setup: setupStatus},
		{name: "ui", summary: "print a one-time sign-in link to this node's local UI", setup: setupUI},
		{name: "user", summary: "manage web UI users (reads the database directly)", sub: userCommands()},
		{name: "version", summary: "print the meshctl version", setup: setupVersion},
	}
}

// newFlags returns a flag set that reports errors instead of exiting.
func newFlags(path string, c command, stderr io.Writer) *flag.FlagSet {
	fs := flag.NewFlagSet(path, flag.ContinueOnError)
	fs.SetOutput(stderr)
	fs.Usage = func() {
		fmt.Fprintf(stderr, "usage: %s%s [flags]\n\n%s\n", path, argSynopsis(c.args), c.summary)
		n := 0
		fs.VisitAll(func(*flag.Flag) { n++ })
		if n > 0 {
			fmt.Fprint(stderr, "\nflags:\n")
			printFlags(stderr, fs)
		}
	}
	return fs
}

// flagName is how help and docs/commands.md spell a flag: -f for a
// one-letter flag, --name otherwise. The flag package takes either.
func flagName(name string) string {
	if len(name) == 1 {
		return "-" + name
	}
	return "--" + name
}

// printFlags is flag.PrintDefaults with flagName's spelling: the flag, its
// value's kind, then its meaning and default on an indented line.
func printFlags(w io.Writer, fs *flag.FlagSet) {
	fs.VisitAll(func(f *flag.Flag) {
		kind, usage := flag.UnquoteUsage(f)
		line := "  " + flagName(f.Name)
		if kind != "" {
			line += " " + kind
		}
		line += "\n    \t" + strings.ReplaceAll(usage, "\n", "\n    \t")
		if def := shownDefault(f); def != "" {
			line += " (default " + def + ")"
		}
		fmt.Fprintln(w, line)
	})
}

// shownDefault is a flag's default as help shows it, as PrintDefaults does:
// text quoted, nothing for a zero value.
func shownDefault(f *flag.Flag) string {
	g, ok := f.Value.(flag.Getter)
	if !ok {
		return f.DefValue
	}
	switch g.Get().(type) {
	case string:
		if f.DefValue == "" {
			return ""
		}
		return strconv.Quote(f.DefValue)
	case bool:
		if f.DefValue == "false" {
			return ""
		}
	case time.Duration:
		if f.DefValue == "0s" {
			return ""
		}
	case int, int64, uint, uint64, float64:
		if f.DefValue == "0" {
			return ""
		}
	}
	return f.DefValue
}

func argSynopsis(args string) string {
	if args == "" {
		return ""
	}
	return " " + args
}

// errHelp means -h was given and usage was printed; run exits 0.
var errHelp = errors.New("help shown")

// parseFlags parses args, mapping -h to errHelp and any other flag error to a
// usage error.
func parseFlags(fs *flag.FlagSet, args []string) error {
	err := fs.Parse(args)
	switch {
	case err == nil:
		return nil
	case errors.Is(err, flag.ErrHelp):
		return errHelp
	default:
		return usagef("%v", err)
	}
}

// parseArgs allows flags before and after positional arguments
// (`meshctl node add vps-1 --role server`), which the flag package alone
// does not.
func parseArgs(fs *flag.FlagSet, args []string) ([]string, error) {
	var pos []string
	for {
		if err := parseFlags(fs, args); err != nil {
			return nil, err
		}
		args = fs.Args()
		if len(args) == 0 {
			return pos, nil
		}
		pos, args = append(pos, args[0]), args[1:]
	}
}

// wantArgs checks the positional argument count.
func wantArgs(args []string, n int, synopsis string) error {
	if len(args) != n {
		if n == 0 {
			return usagef("takes no arguments (got %q)", args)
		}
		return usagef("expects %s (got %d arguments)", synopsis, len(args))
	}
	return nil
}

func setupVersion(_ *flag.FlagSet, e *env) action {
	return func(_ context.Context, args []string) error {
		if err := wantArgs(args, 0, ""); err != nil {
			return err
		}
		fmt.Fprintf(e.stdout, "meshctl %s %s %s/%s\n", e.version, runtime.Version(), runtime.GOOS, runtime.GOARCH)
		return nil
	}
}

// newLogger is the process logger: text to stderr at the given level.
func newLogger(w io.Writer, level string) (*slog.Logger, error) {
	var l slog.Level
	if err := l.UnmarshalText([]byte(level)); err != nil {
		return nil, usagef("--log-level must be debug, info, warn or error (got %q)", level)
	}
	return slog.New(slog.NewTextHandler(w, &slog.HandlerOptions{Level: l})), nil
}

// parsePrefixes reads a comma-separated CIDR list; "" is none.
func parsePrefixes(flagName, s string) ([]netip.Prefix, error) {
	var out []netip.Prefix
	for _, f := range strings.Split(s, ",") {
		if f = strings.TrimSpace(f); f == "" {
			continue
		}
		p, err := netip.ParsePrefix(f)
		if err != nil || p.Masked() != p {
			return nil, usagef("%s: %q is not a network in CIDR form such as 10.0.0.0/8", flagName, f)
		}
		out = append(out, p)
	}
	return out, nil
}

// defaultAgentData is the agent's data dir for goos.
func defaultAgentData(e *env) string {
	if e.goos == "windows" {
		pd := e.getenv("ProgramData")
		if pd == "" {
			pd = `C:\ProgramData`
		}
		return pd + `\meshctl`
	}
	return "/var/lib/meshctl"
}
