package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/hafio/vpn-mesh-utility/internal/agent"
	"github.com/hafio/vpn-mesh-utility/internal/validate"
)

// `meshctl start` is the image's command (amendment A47): the container's
// environment says what runs, so a compose file needs no command line. The
// control plane and a node are told apart by MESHCTL_MODE alone (amendment
// A48); whether a node is a server or a client is its role on the control
// plane, never the container's setting.

// The modes MESHCTL_MODE picks; unset is modeNode, so only the control
// plane's container sets it.
const (
	modeControl = "control"
	modeNode    = "node"
)

// envVar is an environment variable `meshctl start` reads. One with a flag
// is handed on as that flag of the mode's command (`control` or `agent
// run`) when the command has it, so its value meets the same checks as on
// the command line, and docs/commands.md takes its default and meaning from
// there unless it has its own; start reads the others itself. Mode "" is
// both.
type envVar struct {
	name, mode, flag, def, meaning string
}

func startVars() []envVar {
	return []envVar{
		{name: "MESHCTL_MODE", def: modeNode,
			meaning: "what the container runs: control, the control plane (database, web UI, agent API), or node, an agent -- a server or a client as its role on the control plane says"},
		{name: "MESHCTL_URL", flag: "url",
			meaning: "the control plane's public origin, e.g. https://mesh.example.com: the control plane's own (required), the one a node joins (needed with a join token)"},
		{name: "MESHCTL_DATA", def: "/data", meaning: "the data directory, where the volume is mounted"},
		{name: "MESHCTL_LOG_LEVEL", flag: "log-level"},
		{name: "MESHCTL_PORT", mode: modeControl, def: "8080", meaning: "the port the control plane listens on, on every address"},
		{name: "MESHCTL_TRUSTED_PROXIES", mode: modeControl, flag: "trusted-proxies"},
		{name: "MESHCTL_HUB_DOWN_AFTER", mode: modeControl, flag: "hub-down-after"},
		{name: "MESHCTL_NAME", mode: modeNode, meaning: "the node's name: a token issued for another node is refused"},
		{name: "MESHCTL_TOKEN", mode: modeNode,
			meaning: "the node's join token: start joins with a token it has not used before, then runs; with the same token or none, a node that has joined just runs"},
		{name: "MESHCTL_TOKEN_FILE", mode: modeNode, meaning: "an absolute path to a file holding the join token, in place of MESHCTL_TOKEN (a compose secret)"},
		{name: "MESHCTL_ENDPOINT", mode: modeNode, meaning: "server nodes only: the public IPv4:port to join with, when auto-detection would be wrong"},
		{name: "MESHCTL_ENDPOINT6", mode: modeNode, meaning: "server nodes only: the public [IPv6]:port to join with, when auto-detection would be wrong"},
		{name: "MESHCTL_WG_BACKEND", mode: modeNode, flag: "wg-backend"},
		{name: "MESHCTL_RUNTIME_SOCKET", mode: modeNode, flag: "runtime-socket"},
		{name: "MESHCTL_UI_LISTEN", mode: modeNode, flag: "ui-listen"},
		{name: "MESHCTL_MACHINE_ID_FILE", mode: modeNode, flag: "machine-id-file"},
	}
}

func setupStart(_ *flag.FlagSet, e *env) action {
	return func(ctx context.Context, args []string) error {
		if err := wantArgs(args, 0, ""); err != nil {
			return err
		}
		mode, err := startMode(e)
		if err != nil {
			return err
		}
		data, err := startData(e)
		if err != nil {
			return err
		}
		if mode == modeControl {
			port, err := startPort(e)
			if err != nil {
				return err
			}
			return runAs(ctx, e, mode, setupControl, []string{"--data", data, "--listen", ":" + strconv.Itoa(port)})
		}
		o, err := startJoin(e, data)
		if err != nil {
			return err
		}
		if err := agent.JoinOnStart(ctx, o); err != nil {
			if errors.Is(err, agent.ErrNoToken) && e.getenv("MESHCTL_MODE") == "" {
				// The likeliest container without a token and a mode: a
				// control plane whose file left MESHCTL_MODE out.
				err = fmt.Errorf("%w; MESHCTL_MODE is not set, so this container runs a node -- the control plane's sets MESHCTL_MODE=control", err)
			}
			return err
		}
		return runAs(ctx, e, mode, setupAgentRun, []string{"--data", data})
	}
}

// startMode is MESHCTL_MODE, else modeNode. The names before amendment A48
// are refused, never taken as aliases: "server" named the control plane,
// which a server node is not.
func startMode(e *env) (string, error) {
	switch m := e.getenv("MESHCTL_MODE"); m {
	case "":
		return modeNode, nil
	case modeControl, modeNode:
		return m, nil
	case "server", "agent":
		return "", usagef("MESHCTL_MODE=%s is gone: set control for the control plane, or leave it out (node) for a node, server and client alike", m)
	default:
		return "", usagef("MESHCTL_MODE must be control or node (got %q): leave it out for a node", m)
	}
}

// startData is MESHCTL_DATA, else /data.
func startData(e *env) (string, error) {
	d := e.getenv("MESHCTL_DATA")
	if d == "" {
		return "/data", nil
	}
	if _, err := validate.FilePath("MESHCTL_DATA", d); err != nil {
		return "", usagef("%v", err)
	}
	return d, nil
}

// startPort is MESHCTL_PORT, else 8080.
func startPort(e *env) (int, error) {
	p := e.getenv("MESHCTL_PORT")
	if p == "" {
		return 8080, nil
	}
	n, err := strconv.Atoi(p)
	if err == nil {
		_, err = validate.Port("MESHCTL_PORT", n)
	}
	if err != nil {
		return 0, usagef("MESHCTL_PORT must be a port between 1 and 65535 (got %q)", p)
	}
	return n, nil
}

// startJoin is what JoinOnStart needs, from the environment.
func startJoin(e *env, data string) (agent.JoinOptions, error) {
	token, err := startToken(e)
	if err != nil {
		return agent.JoinOptions{}, err
	}
	srv, name := e.getenv("MESHCTL_URL"), e.getenv("MESHCTL_NAME")
	if token != "" {
		if _, err := validate.ServerURL("MESHCTL_URL", srv); err != nil {
			return agent.JoinOptions{}, usagef("%v; it is needed with a join token", err)
		}
	}
	if name != "" {
		if _, err := validate.Name("MESHCTL_NAME", name); err != nil {
			return agent.JoinOptions{}, usagef("%v", err)
		}
	}
	log, err := newLogger(e.stderr, logLevel(e))
	if err != nil {
		return agent.JoinOptions{}, fmt.Errorf("%w (MESHCTL_LOG_LEVEL)", err)
	}
	machine := e.getenv("MESHCTL_MACHINE_ID_FILE")
	if err := checkMachineIDFile(machine); err != nil {
		return agent.JoinOptions{}, fmt.Errorf("%w (MESHCTL_MACHINE_ID_FILE)", err)
	}
	return agent.JoinOptions{DataDir: data, Server: srv, Token: token, Name: name, Endpoint: e.getenv("MESHCTL_ENDPOINT"),
		Endpoint6: e.getenv("MESHCTL_ENDPOINT6"), Version: e.version, Now: e.now, Log: log, MachineIDFile: machine}, nil
}

func logLevel(e *env) string {
	if l := e.getenv("MESHCTL_LOG_LEVEL"); l != "" {
		return l
	}
	return "info"
}

// maxTokenFile caps what is read of MESHCTL_TOKEN_FILE; a token is 43
// characters.
const maxTokenFile = 1024

// startToken is the join token in MESHCTL_TOKEN, or in the file
// MESHCTL_TOKEN_FILE names; "" when neither holds one. Errors never show it.
func startToken(e *env) (string, error) {
	tok, file := strings.TrimSpace(e.getenv("MESHCTL_TOKEN")), e.getenv("MESHCTL_TOKEN_FILE")
	field := "MESHCTL_TOKEN"
	switch {
	case tok != "" && file != "":
		return "", usagef("set MESHCTL_TOKEN or MESHCTL_TOKEN_FILE, not both")
	case file != "":
		if _, err := validate.FilePath("MESHCTL_TOKEN_FILE", file); err != nil || !filepath.IsAbs(file) {
			return "", usagef("MESHCTL_TOKEN_FILE must be an absolute path, such as /run/secrets/meshctl_token (got %q)", file)
		}
		b, err := readCapped(file, maxTokenFile)
		if err != nil {
			return "", fmt.Errorf("read the join token from MESHCTL_TOKEN_FILE: %w", err)
		}
		tok, field = strings.TrimSpace(string(b)), "the token in MESHCTL_TOKEN_FILE"
	}
	if tok == "" {
		return "", nil
	}
	if _, err := validate.Token(field, tok); err != nil {
		return "", usagef("%v", err)
	}
	return tok, nil
}

func readCapped(path string, max int64) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	b, err := io.ReadAll(io.LimitReader(f, max+1))
	if err != nil {
		return nil, err
	}
	if int64(len(b)) > max {
		return nil, fmt.Errorf("%s holds more than %d bytes; it must hold the join token only", path, max)
	}
	return b, nil
}

// runAs runs the leaf command setup as its own command line would, with args
// plus a flag for each of the mode's variables that is set and that the
// command has.
func runAs(ctx context.Context, e *env, mode string, setup func(*flag.FlagSet, *env) action, args []string) error {
	fs := flag.NewFlagSet("meshctl start", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	act := setup(fs, e)
	var used []envVar
	for _, v := range startVars() {
		if val := e.getenv(v.name); v.flag != "" && val != "" && (v.mode == "" || v.mode == mode) && fs.Lookup(v.flag) != nil {
			args = append(args, "--"+v.flag, val)
			used = append(used, v)
		}
	}
	err := fs.Parse(args)
	if err != nil {
		err = usagef("%v", err)
	} else {
		err = act(ctx, fs.Args())
	}
	return fromVar(err, used)
}

// fromVar names the variable behind a flag an error is about.
func fromVar(err error, used []envVar) error {
	if err == nil {
		return nil
	}
	for _, v := range used {
		if strings.Contains(err.Error(), "-"+v.flag) {
			return fmt.Errorf("%w (start sets --%s from %s)", err, v.flag, v.name)
		}
	}
	return err
}
