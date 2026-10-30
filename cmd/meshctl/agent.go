package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
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
		{name: "join", summary: "enroll this node with a one-time token (install.sh and install.ps1 run it, then start the service)", setup: setupAgentJoin},
		{name: "reset", summary: "take off this host what the stopped agent made: tunnel, routes, rules, firewalld zone, DNS", setup: setupAgentReset},
		{name: "run", summary: "run the agent in the foreground (what the service runs)", setup: setupAgentRun},
	}
}

// joinFlags registers what join takes; the returned func reads them once
// they are parsed.
func joinFlags(fs *flag.FlagSet, e *env) func() (agent.JoinOptions, error) {
	srv := fs.String("server", "", "the control plane's origin, e.g. https://mesh.example.com")
	// The token is never a flag default: -h would print it.
	token := fs.String("token", "", "one-time join token (or set MESHCTL_TOKEN, which keeps it out of the process list)")
	name := fs.String("name", "", "the node name the token must belong to (a safety check)")
	endpoint := fs.String("endpoint", "", "servers only: public IPv4:port, when auto-detection would be wrong")
	endpoint6 := fs.String("endpoint6", "", "servers only: public [IPv6]:port, when auto-detection would be wrong")
	mesh := fs.String("mesh", "", "the token's node's mesh, which names the data dir (default: default, unless --data names one)")
	data := fs.String("data", "", "agent data directory (default: "+meshDir(e, agentRoot(e), "<mesh>")+")")
	machine := machineIDFlag(fs)
	return func() (agent.JoinOptions, error) {
		tok := *token
		if tok == "" {
			tok = e.getenv("MESHCTL_TOKEN")
		}
		if *srv == "" || tok == "" {
			return agent.JoinOptions{}, usagef("--server and --token (or MESHCTL_TOKEN) are required; `meshctl node add` prints both")
		}
		if err := checkMachineIDFile(*machine); err != nil {
			return agent.JoinOptions{}, err
		}
		// The mesh names the data dir (amendment A51); a data dir given
		// outright is joined into for whatever mesh the token's is.
		dir, m := *data, *mesh
		if dir == "" {
			if m == "" {
				m = "default"
			}
			if _, err := validate.Name("--mesh", m); err != nil {
				return agent.JoinOptions{}, usagef("%v", err)
			}
			dir = meshDir(e, agentRoot(e), m)
		}
		return agent.JoinOptions{DataDir: dir, Server: *srv, Token: tok, Name: *name, Mesh: m, Endpoint: *endpoint, Version: e.version,
			Out: e.stdout, Now: e.now, Endpoint6: *endpoint6, MachineIDFile: *machine}, nil
	}
}

// machineIDFlag registers --machine-id-file (amendment A50).
func machineIDFlag(fs *flag.FlagSet) *string {
	return fs.String("machine-id-file", "", "the file holding this host's machine id, which pins the node to it "+
		"(default: /etc/machine-id, /var/lib/dbus/machine-id or the firmware's UUID; MachineGuid on Windows); a container mounts the host's")
}

func checkMachineIDFile(path string) error {
	if path == "" {
		return nil
	}
	if _, err := validate.FilePath("--machine-id-file", path); err != nil {
		return usagef("%v", err)
	}
	return nil
}

func setupAgentJoin(fs *flag.FlagSet, e *env) action {
	opts := joinFlags(fs, e)
	return func(ctx context.Context, args []string) error {
		if err := wantArgs(args, 0, ""); err != nil {
			return err
		}
		o, err := opts()
		if err != nil {
			return err
		}
		_, err = agent.Join(ctx, o)
		return err
	}
}

func setupAgentReset(fs *flag.FlagSet, e *env) action {
	data := meshFlags(fs, e)
	noFirewall := fs.Bool("no-firewall", false, "leave firewalld alone, as an agent run with --no-firewall did")
	return func(ctx context.Context, args []string) error {
		if err := wantArgs(args, 0, ""); err != nil {
			return err
		}
		dir, err := data()
		if err != nil {
			return err
		}
		log, err := newLogger(e.stderr, "info")
		if err != nil {
			return err
		}
		n, err := agent.Reset(ctx, agent.Deps{DataDir: dir, NoFirewall: *noFirewall, Log: log, Version: e.version, Now: e.now})
		switch {
		case err != nil && n > 0:
			return fmt.Errorf("took %d changes off this host, then: %w", n, err)
		case err != nil:
			return err
		case n == 0:
			fmt.Fprintln(e.stdout, "nothing this agent made is left on this host")
		default:
			fmt.Fprintf(e.stdout, "took %d changes off this host; the data dir is as it was\n", n)
		}
		return nil
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
	machine := machineIDFlag(fs)
	return func(ctx context.Context, args []string) error {
		if err := wantArgs(args, 0, ""); err != nil {
			return err
		}
		if err := checkMachineIDFile(*machine); err != nil {
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
		dir, err := data()
		if err != nil {
			return err
		}
		socket, err := sock()
		if err != nil {
			return err
		}
		asService, err := service.IsWindowsService()
		if err != nil {
			return fmt.Errorf("find out whether the Windows service manager started meshctl: %w", err)
		}
		out, closeLog, err := agentLog(e.stderr, dir, asService)
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
			DataDir: dir, Socket: socket, UIListen: uiListen, Version: e.version,
			NoFirewall: *noFirewall, Backend: *backend, Log: log, Logs: ring, Now: e.now, RuntimeSocket: *runtime,
			StatusEvery: *statusEvery, EndpointRetry: *endpointRetry, MachineIDFile: *machine,
		}
		if asService {
			return service.RunWindowsService(service.Name(filepath.Base(dir)), func(ctx context.Context) error {
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

// socketFlags registers --mesh, --data and --socket. The returned funcs
// resolve the data dir (meshFlags) and the socket once flags are parsed:
// --socket, else the default for the data dir.
func socketFlags(fs *flag.FlagSet, e *env) (data, sock func() (string, error)) {
	data = meshFlags(fs, e)
	socket := fs.String("socket", "", `local API socket (default: /run/meshctl/<mesh>.sock for a mesh's data dir under /var/lib/meshctl, `+
		`else <data>/agent.sock; on Windows \\.\pipe\meshctl-agent, with -<slot> for a mesh in another slot)`)
	return data, func() (string, error) {
		if *socket != "" {
			return *socket, nil
		}
		dir, err := data()
		if err != nil {
			return "", err
		}
		return agent.DefaultSocket(e.goos, dir), nil
	}
}

// meshFlags registers --mesh and --data. The returned func is the agent's
// data dir once flags are parsed (amendment A51): --data, else the mesh's
// under the agent root, else the one mesh that has joined on this host.
func meshFlags(fs *flag.FlagSet, e *env) func() (string, error) {
	root := agentRoot(e)
	mesh := fs.String("mesh", "", "the mesh whose agent this is (default: the only one on this host)")
	data := fs.String("data", "", "the agent's data directory (default: "+meshDir(e, root, "<mesh>")+")")
	return func() (string, error) {
		switch {
		case *data != "":
			return *data, nil
		case *mesh != "":
			if _, err := validate.Name("--mesh", *mesh); err != nil {
				return "", usagef("%v", err)
			}
			return meshDir(e, root, *mesh), nil
		}
		return onlyMesh(e, root)
	}
}

// meshDir is the data dir of mesh's agent under root.
func meshDir(e *env, root, mesh string) string {
	if e.goos == "windows" {
		return root + `\` + mesh
	}
	return root + "/" + mesh
}

// onlyMesh is the data dir of the one mesh that has joined under root. A
// mesh's entry may be a link to its data dir elsewhere (install.sh --data,
// amendment A54).
func onlyMesh(e *env, root string) (string, error) {
	entries, err := os.ReadDir(root)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return "", fmt.Errorf("list the meshes under %s: %w", root, err)
	}
	var meshes []string
	for _, en := range entries {
		if _, err := os.Stat(meshDir(e, root, en.Name()) + sep(e) + state.EnrollmentFile); err == nil {
			meshes = append(meshes, en.Name())
		}
	}
	switch len(meshes) {
	case 0:
		return "", fmt.Errorf("no node has joined a mesh on this host (%s holds none): join with `meshctl agent join --mesh <mesh>`, "+
			"or pass --data", root)
	case 1:
		return meshDir(e, root, meshes[0]), nil
	}
	return "", usagef("this host runs the agents of several meshes (%s): pass --mesh", strings.Join(meshes, ", "))
}

// sep is the path separator of e's platform.
func sep(e *env) string {
	if e.goos == "windows" {
		return `\`
	}
	return "/"
}
