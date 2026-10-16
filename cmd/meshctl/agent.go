package main

import (
	"context"
	"flag"
	"io"

	"github.com/hafio/vpn-mesh-utility/internal/agent"
	"github.com/hafio/vpn-mesh-utility/internal/agent/localui"
	"github.com/hafio/vpn-mesh-utility/internal/agent/logbuf"
	"github.com/hafio/vpn-mesh-utility/internal/agent/service"
	"github.com/hafio/vpn-mesh-utility/internal/agent/wg"
)

func agentCommands() []command {
	return []command{
		{name: "join", summary: "enroll this node with a one-time token, then install and start the service", setup: setupAgentJoin},
		{name: "run", summary: "run the agent in the foreground (what the service runs)", setup: setupAgentRun},
	}
}

func setupAgentJoin(fs *flag.FlagSet, e *env) action {
	srv := fs.String("server", "", "server origin, e.g. https://mesh.example.com")
	// The token is never a flag default: -h would print it.
	token := fs.String("token", "", "one-time join token (or set MESHCTL_TOKEN, which keeps it out of the process list)")
	name := fs.String("name", "", "the node name the token must belong to (a safety check)")
	endpoint := fs.String("endpoint", "", "servers only: public IPv4:port, when auto-detection would be wrong")
	data := fs.String("data", defaultAgentData(e), "agent data directory")
	noService := fs.Bool("no-service", false, "do not install and start the system service")
	return func(ctx context.Context, args []string) error {
		if err := wantArgs(args, 0, ""); err != nil {
			return err
		}
		tok := *token
		if tok == "" {
			tok = e.getenv("MESHCTL_TOKEN")
		}
		if *srv == "" || tok == "" {
			return usagef("--server and --token (or MESHCTL_TOKEN) are required; `meshctl node add` prints both")
		}
		_, err := agent.Join(ctx, agent.JoinOptions{
			DataDir: *data, Server: *srv, Token: tok, Name: *name, Endpoint: *endpoint, Version: e.version,
			NoService: *noService, Service: service.Systemd(), Out: e.stdout, Now: e.now,
		})
		return err
	}
}

func setupAgentRun(fs *flag.FlagSet, e *env) action {
	data, sock := socketFlags(fs, e)
	noFirewall := fs.Bool("no-firewall", false, "leave the host firewall alone (you open udp/51820 and the Swarm ports yourself)")
	backend := fs.String("wg-backend", wg.BackendAuto, "WireGuard backend: auto (kernel when the module is there), kernel or userspace (wireguard-go on /dev/net/tun)")
	level := fs.String("log-level", "info", "debug, info, warn or error")
	ui := fs.String("ui-listen", "127.0.0.1:7443", "the local UI's loopback address, or off; meshctl ui prints a sign-in link")
	return func(ctx context.Context, args []string) error {
		if err := wantArgs(args, 0, ""); err != nil {
			return err
		}
		uiListen := ""
		if *ui != "off" {
			if _, err := localui.ParseAddr(*ui); err != nil {
				return usagef("%v", err)
			}
			uiListen = *ui
		}
		// The log goes to stderr (the journal) and to the ring `meshctl
		// logs` reads.
		ring := logbuf.New(logbuf.DefaultLines)
		log, err := newLogger(io.MultiWriter(e.stderr, ring), *level)
		if err != nil {
			return err
		}
		return agent.Run(ctx, agent.Deps{
			DataDir: *data, Socket: sock(), UIListen: uiListen, Version: e.version,
			NoFirewall: *noFirewall, Backend: *backend, Log: log, Logs: ring, Now: e.now,
		})
	}
}

// socketFlags registers --data and --socket. The returned func resolves the
// socket once flags are parsed: --socket, else the default for the data dir.
func socketFlags(fs *flag.FlagSet, e *env) (*string, func() string) {
	data := fs.String("data", defaultAgentData(e), "agent data directory")
	socket := fs.String("socket", "", "local API socket (default: /run/meshctl/agent.sock for the standard data dir, else <data>/agent.sock)")
	return data, func() string {
		if *socket != "" {
			return *socket
		}
		return agent.DefaultSocket(*data)
	}
}
