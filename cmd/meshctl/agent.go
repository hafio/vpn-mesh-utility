package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"time"

	"github.com/hafio/vpn-mesh-utility/internal/agent"
	"github.com/hafio/vpn-mesh-utility/internal/agent/localui"
	"github.com/hafio/vpn-mesh-utility/internal/agent/logbuf"
	"github.com/hafio/vpn-mesh-utility/internal/agent/service"
	"github.com/hafio/vpn-mesh-utility/internal/agent/state"
	"github.com/hafio/vpn-mesh-utility/internal/agent/wg"
	"github.com/hafio/vpn-mesh-utility/internal/validate"
)

func agentCommands() []command {
	return []command{
		{name: "install", summary: "Windows: copy meshctl into Program Files, join this node and install the service", setup: setupAgentInstall},
		{name: "join", summary: "enroll this node with a one-time token, then install and start the service", setup: setupAgentJoin},
		{name: "run", summary: "run the agent in the foreground (what the service runs)", setup: setupAgentRun},
		{name: "uninstall", summary: "Windows: stop and remove the service (--purge also removes the data dir)", setup: setupAgentUninstall},
	}
}

// joinFlags registers what join and install take; the returned func reads
// them once they are parsed.
func joinFlags(fs *flag.FlagSet, e *env) func() (agent.JoinOptions, error) {
	srv := fs.String("server", "", "the control plane's origin, e.g. https://mesh.example.com")
	// The token is never a flag default: -h would print it.
	token := fs.String("token", "", "one-time join token (or set MESHCTL_TOKEN, which keeps it out of the process list)")
	name := fs.String("name", "", "the node name the token must belong to (a safety check)")
	endpoint := fs.String("endpoint", "", "servers only: public IPv4:port, when auto-detection would be wrong")
	endpoint6 := fs.String("endpoint6", "", "servers only: public [IPv6]:port, when auto-detection would be wrong")
	data := fs.String("data", defaultAgentData(e), "agent data directory")
	return func() (agent.JoinOptions, error) {
		tok := *token
		if tok == "" {
			tok = e.getenv("MESHCTL_TOKEN")
		}
		if *srv == "" || tok == "" {
			return agent.JoinOptions{}, usagef("--server and --token (or MESHCTL_TOKEN) are required; `meshctl node add` prints both")
		}
		return agent.JoinOptions{DataDir: *data, Server: *srv, Token: tok, Name: *name, Endpoint: *endpoint, Version: e.version,
			Out: e.stdout, Now: e.now, Endpoint6: *endpoint6}, nil
	}
}

func setupAgentJoin(fs *flag.FlagSet, e *env) action {
	opts := joinFlags(fs, e)
	noService := fs.Bool("no-service", false, "do not install and start the system service")
	return func(ctx context.Context, args []string) error {
		if err := wantArgs(args, 0, ""); err != nil {
			return err
		}
		o, err := opts()
		if err != nil {
			return err
		}
		o.NoService, o.Service = *noService, hostService(o.DataDir)
		_, err = agent.Join(ctx, o)
		return err
	}
}

func setupAgentInstall(fs *flag.FlagSet, e *env) action {
	opts := joinFlags(fs, e)
	return func(ctx context.Context, args []string) error {
		if err := wantArgs(args, 0, ""); err != nil {
			return err
		}
		o, err := opts()
		if err != nil {
			return err
		}
		return installAgent(ctx, e, o)
	}
}

func setupAgentUninstall(fs *flag.FlagSet, e *env) action {
	data := fs.String("data", defaultAgentData(e), "agent data directory")
	purge := fs.Bool("purge", false, "also remove the data directory: this node's keys and cached config (it must join again)")
	return func(ctx context.Context, args []string) error {
		if err := wantArgs(args, 0, ""); err != nil {
			return err
		}
		return uninstallAgent(ctx, e, *data, *purge)
	}
}

func setupAgentRun(fs *flag.FlagSet, e *env) action {
	data, sock := socketFlags(fs, e)
	noFirewall := fs.Bool("no-firewall", false, "leave the host firewall alone (you open the WireGuard port and the Swarm ports yourself)")
	backend := fs.String("wg-backend", wg.BackendAuto, "WireGuard backend: auto (kernel when the module is there), kernel or userspace (wireguard-go on /dev/net/tun)")
	level := fs.String("log-level", "info", "debug, info, warn or error")
	ui := fs.String("ui-listen", agent.UIAuto,
		"the local UI's loopback address, auto (127.0.0.1:7443, plus the mesh's slot) or off; meshctl ui prints a sign-in link")
	runtime := fs.String("runtime-socket", agent.RuntimeAuto,
		"the container runtime's API socket the containers' names are read from (GET only): auto (Docker's, then Podman's), off, or a path")
	statusEvery := fs.Duration("status-every", 30*time.Second,
		"how often the agent reports to the server; the server's --hub-down-after must be longer on a server")
	endpointRetry := fs.Duration("endpoint-retry", 45*time.Second,
		"how long a spoke waits for a first handshake on a server's endpoint before it dials the server's other-family one")
	return func(ctx context.Context, args []string) error {
		if err := wantArgs(args, 0, ""); err != nil {
			return err
		}
		if *statusEvery < time.Second || *statusEvery > 5*time.Minute {
			return usagef("--status-every must be between 1s and 5m (got %s)", *statusEvery)
		}
		if *endpointRetry < 5*time.Second || *endpointRetry > 5*time.Minute {
			return usagef("--endpoint-retry must be between 5s and 5m (got %s)", *endpointRetry)
		}
		if *runtime != agent.RuntimeAuto && *runtime != agent.RuntimeOff {
			if _, err := validate.POSIXPath("--runtime-socket", *runtime); err != nil {
				return usagef("%v", err)
			}
		}
		uiListen := ""
		if *ui != "off" {
			if _, err := localui.ParseAddr(*ui); err != nil && *ui != agent.UIAuto {
				return usagef("%v", err)
			}
			uiListen = *ui
		}
		asService, err := service.IsWindowsService()
		if err != nil {
			return fmt.Errorf("find out whether the Windows service manager started meshctl: %w", err)
		}
		out, closeLog, err := agentLog(e.stderr, *data, asService)
		if err != nil {
			return err
		}
		defer closeLog()
		ring := logbuf.New(logbuf.DefaultLines)
		log, err := newLogger(io.MultiWriter(out, ring), *level)
		if err != nil {
			return err
		}
		d := agent.Deps{
			DataDir: *data, Socket: sock(), UIListen: uiListen, Version: e.version,
			NoFirewall: *noFirewall, Backend: *backend, Log: log, Logs: ring, Now: e.now, RuntimeSocket: *runtime,
			StatusEvery: *statusEvery, EndpointRetry: *endpointRetry,
		}
		if asService {
			return service.RunWindowsService(func(ctx context.Context) error {
				err := agent.Run(ctx, d)
				if errors.Is(err, agent.ErrRefused) {
					// A clean stop: the service manager restarts only a
					// failed run (amendment A50).
					log.Error("the agent stopped and the service stays stopped until an operator acts", "err", err)
					return nil
				}
				return err
			})
		}
		return agent.Run(ctx, d)
	}
}

// serviceLogMax caps the Windows service's log file (and its one rotated
// copy).
const serviceLogMax = 10 << 20

// agentLog is where the agent's text log goes besides the ring `meshctl
// logs` reads: stderr (the journal), or for the Windows service, whose
// stderr nobody reads, <data>/agent.log in the data dir (protected like the
// keys beside it).
func agentLog(stderr io.Writer, data string, asService bool) (io.Writer, func(), error) {
	if !asService {
		return stderr, func() {}, nil
	}
	if _, err := state.Open(data); err != nil {
		return nil, nil, err
	}
	f, err := logbuf.OpenFile(filepath.Join(data, "agent.log"), serviceLogMax)
	if err != nil {
		return nil, nil, err
	}
	return f, func() { _ = f.Close() }, nil // the agent has stopped; nothing to report to
}

// socketFlags registers --data and --socket. The returned func resolves the
// socket once flags are parsed: --socket, else the default for the data dir.
func socketFlags(fs *flag.FlagSet, e *env) (*string, func() string) {
	data := fs.String("data", defaultAgentData(e), "agent data directory")
	socket := fs.String("socket", "", `local API socket (default: /run/meshctl/agent.sock for the standard data dir, else <data>/agent.sock; `+
		`on Windows \\.\pipe\meshctl-agent, with -<slot> for a mesh in another slot)`)
	return data, func() string {
		if *socket != "" {
			return *socket
		}
		return agent.DefaultSocket(e.goos, *data)
	}
}
