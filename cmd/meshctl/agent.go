package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
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
	mesh := fs.String("mesh", "", "the token's node's mesh, which names the data dir (default: default, unless --data or MESHCTL_DATA names one)")
	data := fs.String("data", "", dataHelp(e))
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
		// The mesh names the data dir (amendments A51, A56); a data dir given
		// outright is joined into for whatever mesh the token's is, unless
		// its node joined another mesh than --mesh.
		m := *mesh
		if m != "" {
			if _, err := validate.Name("--mesh", m); err != nil {
				return agent.JoinOptions{}, usagef("%v", err)
			}
		}
		dir, from, err := givenDir(e, *data)
		switch {
		case err != nil:
			return agent.JoinOptions{}, err
		case dir != "":
			if _, err := dirMesh(e, dir, from, m); err != nil {
				return agent.JoinOptions{}, err
			}
		default:
			if m == "" {
				m = "default"
			}
			if dir, err = installedDir(e, m); err != nil {
				return agent.JoinOptions{}, err
			}
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
		dir, _, err := data()
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
// --socket, else on Linux /run/meshctl/<mesh>.sock for a mesh installed on
// this host (the unit's, wherever the data is), else the default for the
// data dir.
func socketFlags(fs *flag.FlagSet, e *env) (data, sock func() (string, error)) {
	where := meshFlags(fs, e)
	socket := fs.String("socket", "", `local API socket (default: /run/meshctl/<mesh>.sock for a mesh installed on this host, `+
		`else <data>/agent.sock; on Windows \\.\pipe\meshctl-agent, with -<slot> for a mesh in another slot)`)
	data = func() (string, error) {
		dir, _, err := where()
		return dir, err
	}
	return data, func() (string, error) {
		if *socket != "" {
			return *socket, nil
		}
		dir, mesh, err := where()
		switch {
		case mesh != "" && e.goos != "windows":
			return agent.MeshSocket(mesh), nil
		case err != nil:
			return "", err
		}
		return agent.DefaultSocket(e.goos, dir), nil
	}
}

// meshFlags registers --mesh and --data. The returned func is, once flags
// are parsed, the agent's data dir and, when that is the dir of a mesh
// installed on this host, the mesh (amendments A51, A56): --data, else
// MESHCTL_DATA, else for --mesh the dir install.sh --data recorded in
// /etc/meshctl/<mesh>.env, else the mesh's under the agent root; without
// --mesh, the one mesh that has joined on this host. A dir given outright
// whose node joined another mesh than --mesh is refused. For --mesh the
// mesh comes back even when its record cannot be read, with that error.
func meshFlags(fs *flag.FlagSet, e *env) func() (dir, mesh string, err error) {
	mesh := fs.String("mesh", "", "the mesh whose agent this is (default: the only one on this host)")
	data := fs.String("data", "", dataHelp(e))
	return func() (string, string, error) {
		if *mesh != "" {
			if _, err := validate.Name("--mesh", *mesh); err != nil {
				return "", "", usagef("%v", err)
			}
		}
		dir, from, err := givenDir(e, *data)
		switch {
		case err != nil:
			return "", "", err
		case dir != "":
			m, err := dirMesh(e, dir, from, *mesh)
			return dir, m, err
		case *mesh != "":
			d, err := installedDir(e, *mesh)
			return d, *mesh, err
		}
		return onlyMesh(e, agentRoot(e))
	}
}

// dataHelp is what --data says about its default on e's platform.
func dataHelp(e *env) string {
	recorded := ""
	if e.goos != "windows" {
		recorded = "the dir " + recordsDir + "/<mesh>.env records, else "
	}
	return "the agent's data directory (default: MESHCTL_DATA, else " + recorded + meshDir(e, agentRoot(e), "<mesh>") + ")"
}

// givenDir is the data dir given outright, and what gave it: --data
// (given), else MESHCTL_DATA; "" when neither did.
func givenDir(e *env, given string) (dir, from string, err error) {
	dir, from = given, "--data"
	if dir == "" {
		dir, from = e.getenv("MESHCTL_DATA"), "MESHCTL_DATA"
	}
	if dir == "" {
		return "", "", nil
	}
	if _, err := validate.FilePath(from, dir); err != nil {
		return "", "", usagef("%v", err)
	}
	return dir, from, nil
}

// dirMesh checks a data dir given outright (from says how) against --mesh
// (want): a dir whose node joined another mesh is refused, so a
// MESHCTL_DATA left set for one mesh never acts on another. It returns the
// dir's mesh when the dir is that mesh's installed on this host, else "". A
// dir without an enrollment this process can read has nothing to check.
func dirMesh(e *env, dir, from, want string) (string, error) {
	en, err := (&state.Dir{Path: dir}).Enrollment()
	if err != nil {
		return "", nil
	}
	if _, err := validate.Name("mesh", en.Mesh); err != nil {
		return "", nil
	}
	if want != "" && en.Mesh != want {
		return "", usagef("%s %s is the data dir of the mesh %s, not of %s: name %s's own, or leave %s out", from, dir, en.Mesh, want, want, from)
	}
	if home, err := installedDir(e, en.Mesh); err != nil || filepath.Clean(home) != filepath.Clean(dir) {
		return "", nil
	}
	return en.Mesh, nil
}

// recordsDir is where install.sh --data records, per mesh, where the
// agent keeps its data (amendment A56): <mesh>.env, a line
// MESHCTL_DATA=<dir>.
const recordsDir = "/etc/meshctl"

// recordMax caps a record: install.sh writes a few lines.
const recordMax = 4 << 10

// installedDir is the data dir of mesh's agent as installed on this host:
// the one install.sh --data recorded (amendment A56), else the mesh's under
// the agent root.
func installedDir(e *env, mesh string) (string, error) {
	dir, err := recorded(e, mesh)
	if err != nil || dir != "" {
		return dir, err
	}
	return meshDir(e, agentRoot(e), mesh), nil
}

// recorded is the data dir e.records/<mesh>.env records for mesh's agent;
// "" when there is no record, or no records on this platform. A record that
// cannot be read, or that records no path install.sh's --data takes
// (validate.DataDir), is the error.
func recorded(e *env, mesh string) (string, error) {
	if e.records == "" {
		return "", nil
	}
	file := filepath.Join(e.records, mesh+".env")
	b, err := os.ReadFile(file)
	switch {
	case errors.Is(err, os.ErrNotExist):
		return "", nil
	case err != nil:
		return "", fmt.Errorf("read where the agent of %s keeps its data: %w (or pass --data)", mesh, err)
	case len(b) > recordMax:
		return "", fmt.Errorf("%s is over %d bytes, which install.sh never writes: correct or remove it", file, recordMax)
	}
	for _, line := range strings.Split(string(b), "\n") {
		if v, ok := strings.CutPrefix(strings.TrimSpace(line), "MESHCTL_DATA="); ok {
			if _, err := validate.DataDir(file, v); err != nil {
				return "", fmt.Errorf("%w: correct or remove that file", err)
			}
			return v, nil
		}
	}
	return "", fmt.Errorf("%s has no MESHCTL_DATA=<dir> line: correct or remove it", file)
}

// meshDir is the data dir of mesh's agent under root.
func meshDir(e *env, root, mesh string) string {
	if e.goos == "windows" {
		return root + `\` + mesh
	}
	return root + "/" + mesh
}

// onlyMesh is the data dir of the one mesh that has joined on this host,
// and that mesh: a mesh's dir under root that holds an enrollment, or the
// dir its record names (amendment A56), which wins over the one under root.
func onlyMesh(e *env, root string) (string, string, error) {
	found := map[string]string{}
	entries, err := os.ReadDir(root)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return "", "", fmt.Errorf("list the meshes under %s: %w", root, err)
	}
	for _, en := range entries {
		if _, err := validate.Name("mesh", en.Name()); err == nil && joinedAt(e, meshDir(e, root, en.Name())) {
			found[en.Name()] = meshDir(e, root, en.Name())
		}
	}
	if e.records != "" {
		recs, err := os.ReadDir(e.records)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return "", "", fmt.Errorf("list the meshes recorded in %s: %w (or pass --mesh)", e.records, err)
		}
		for _, r := range recs {
			m, ok := strings.CutSuffix(r.Name(), ".env")
			if _, err := validate.Name("mesh", m); !ok || err != nil {
				continue
			}
			dir, err := recorded(e, m)
			if err != nil {
				return "", "", err
			}
			delete(found, m)
			if joinedAt(e, dir) {
				found[m] = dir
			}
		}
	}
	meshes := make([]string, 0, len(found))
	for m := range found {
		meshes = append(meshes, m)
	}
	sort.Strings(meshes)
	switch len(meshes) {
	case 0:
		return "", "", fmt.Errorf("no node has joined a mesh on this host (%s holds none): join with `meshctl agent join --mesh <mesh>`, "+
			"or pass --data", root)
	case 1:
		return found[meshes[0]], meshes[0], nil
	}
	return "", "", usagef("this host runs the agents of several meshes (%s): pass --mesh", strings.Join(meshes, ", "))
}

// joinedAt is true when the data dir holds an enrollment.
func joinedAt(e *env, dir string) bool {
	_, err := os.Stat(dir + sep(e) + state.EnrollmentFile)
	return err == nil
}

// sep is the path separator of e's platform.
func sep(e *env) string {
	if e.goos == "windows" {
		return `\`
	}
	return "/"
}
