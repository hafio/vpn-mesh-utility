package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"strconv"
	"text/tabwriter"
	"time"

	"github.com/hafio/vpn-mesh-utility/deploy"
	"github.com/hafio/vpn-mesh-utility/internal/server"
	"github.com/hafio/vpn-mesh-utility/internal/server/api"
	"github.com/hafio/vpn-mesh-utility/internal/validate"
)

// The node commands talk to the running server over <data>/admin.sock
// (amendment A9), so every change goes through its reconciler: versioned,
// audited and pushed to the agents at once. In the container:
// `docker exec <container> /app/meshctl node list`.

func nodeCommands() []command {
	return []command{
		{name: "add", args: "<name>", summary: "add a node and print its one-time join command", setup: setupNodeAdd},
		{name: "compose", args: "<name>", summary: "print the compose file that runs a node's agent in a container (container mode, or a NAS)", setup: setupNodeCompose},
		{name: "disable", args: "<name>", summary: "take a node out of every peer list (its keys stay registered)", setup: nodeAction("disable")},
		{name: "enable", args: "<name>", summary: "put a disabled node back", setup: nodeAction("enable")},
		{name: "list", summary: "list the nodes and the mesh plan's warnings", setup: setupNodeList},
		{name: "network", args: "<name>", summary: "print the commands that put a node's containers on the mesh (its mesh network, then a service)", setup: setupNodeNetwork},
		{name: "phone", args: "<name>", summary: "print a new WireGuard profile for a phone (a new key: its previous profile stops working)", setup: setupNodePhone},
		{name: "rm", args: "<name>", summary: "remove a node; every peer drops it on the next apply", setup: nodeAction("rm")},
		{name: "set", args: "<name>", summary: "change a node's endpoint, hub, exit, keepalive, MTU or container network", setup: setupNodeSet},
		{name: "show", args: "<name>", summary: "show a node, its computed config and its peers", setup: setupNodeShow},
		{name: "token", args: "<name>", summary: "issue a new join token (revokes unused ones; re-keys an enrolled node)", setup: nodeAction("token")},
	}
}

// adminFlags registers --data and --json and returns the admin client
// constructor.
func adminFlags(fs *flag.FlagSet) (func() *api.AdminClient, *bool) {
	data := fs.String("data", "/data", "the server's data directory (holds admin.sock)")
	asJSON := fs.Bool("json", false, "print the server's answer as JSON")
	return func() *api.AdminClient { return api.NewAdminClient(filepath.Join(*data, server.AdminSocket)) }, asJSON
}

// nodeName takes and validates the one positional argument.
func nodeName(args []string) (string, error) {
	if err := wantArgs(args, 1, "<name>"); err != nil {
		return "", err
	}
	name, err := validate.Name("name", args[0])
	if err != nil {
		return "", usagef("%v", err)
	}
	return name, nil
}

func printJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func setupNodeAdd(fs *flag.FlagSet, e *env) action {
	client, asJSON := adminFlags(fs)
	role := fs.String("role", "", "server, spoke or phone (required)")
	ip := fs.String("ip", "", "fixed mesh IP (default: the lowest free one in the role's block)")
	endpoint := fs.String("endpoint", "", "servers: public IPv4:port (default: what the agent detects at join)")
	listenPort := fs.Int("listen-port", 0, "servers: WireGuard UDP port (default 51820)")
	hub := fs.String("hub", "", "spokes and phones: the server that relays for it (default: the default hub)")
	canExit := fs.Bool("can-exit", false, "servers: may be chosen as an internet exit")
	alwaysOn := fs.Bool("always-on", false, "spokes: allocate from the always-on block (NAS)")
	keepalive := fs.Int("keepalive", 0, "spokes and phones: PersistentKeepalive seconds (default 25)")
	mtu := fs.Int("mtu", 0, "interface MTU (default: the mesh default, 1420)")
	mode := fs.String("mode", "host", "host (WireGuard on the host) or container (in the agent's container: only its containers are on the mesh)")
	containers := fs.Bool("containers", false, "host mode: put the node's containers on the mesh too (container mode always does)")
	return func(ctx context.Context, args []string) error {
		name, err := nodeName(args)
		if err != nil {
			return err
		}
		if *role == "" {
			return usagef("--role is required (server, spoke or phone)")
		}
		ch, err := client().Add(ctx, api.NodeAddRequest{
			Name: name, Role: *role, IP: *ip, Endpoint: *endpoint, ListenPort: *listenPort, Hub: *hub,
			CanExit: *canExit, AlwaysOn: *alwaysOn, Keepalive: *keepalive, MTU: *mtu, Mode: *mode, Containers: *containers,
		})
		if err != nil {
			return err
		}
		if *asJSON {
			return printJSON(e.stdout, ch)
		}
		fmt.Fprintf(e.stdout, "added %s (%s, mesh IP %s)\n", ch.Node.Name, ch.Node.Role, ch.Node.MeshIP)
		if ch.Node.ContainerSubnet != "" {
			fmt.Fprintf(e.stdout, "its containers: %s\n", ch.Node.ContainerSubnet)
		}
		if ch.Node.Mode == "container" {
			fmt.Fprintf(e.stdout, "\ncontainer mode: run its agent from the compose file\n  meshctl node compose %s [--runtime docker|podman|swarm] > compose.yml\n", ch.Node.Name)
		}
		printEnrollment(e.stdout, e.now().Location(), ch.Enrollment)
		if ch.Node.Role == "phone" {
			fmt.Fprintf(e.stdout, "\nmake its profile for the WireGuard app:\n  meshctl node phone %s [--exit <server>]\nor as a QR code on its page in the web UI\n", ch.Node.Name)
		}
		printWarnings(e.stdout, ch.Warnings)
		return nil
	}
}

// printEnrollment shows a join token and how to use it; times are in loc,
// the local time zone (TZ).
func printEnrollment(w io.Writer, loc *time.Location, en *api.Enrollment) {
	if en == nil {
		return
	}
	fmt.Fprintf(w, "\njoin token (shown once, valid until %s):\n  %s\n", time.Unix(en.ExpiresAt, 0).In(loc).Format(time.RFC3339), en.Token)
	fmt.Fprintf(w, "\non the node, as root:\n  %s\n", en.Install)
	fmt.Fprintf(w, "or, where meshctl is already installed:\n  %s\n", en.Join)
}

func printWarnings(w io.Writer, ws []string) {
	if len(ws) == 0 {
		return
	}
	fmt.Fprintln(w)
	for _, s := range ws {
		fmt.Fprintln(w, s)
	}
}

func setupNodeList(fs *flag.FlagSet, e *env) action {
	client, asJSON := adminFlags(fs)
	return func(ctx context.Context, args []string) error {
		if err := wantArgs(args, 0, ""); err != nil {
			return err
		}
		l, err := client().List(ctx)
		if err != nil {
			return err
		}
		if *asJSON {
			return printJSON(e.stdout, l)
		}
		tw := tabwriter.NewWriter(e.stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(tw, "NAME\tROLE\tMESH IP\tENDPOINT\tSTATE\tVERSION\tLAST SEEN")
		for _, n := range l.Nodes {
			ep := n.Endpoint
			if ep == "" {
				ep = "-"
			}
			fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%d/%d\t%s\n", n.Name, n.Role, n.MeshIP, ep, nodeState(n),
				n.AppliedVersion, n.ConfigVersion, ago(heard(n), e.now()))
		}
		if err := tw.Flush(); err != nil {
			return err
		}
		printWarnings(e.stdout, l.Warnings)
		return nil
	}
}

// nodeState summarises a node for the list. A phone's state is its
// profile's: it has no agent to report anything else.
func nodeState(n api.NodeInfo) string {
	switch {
	case n.Disabled:
		return "disabled"
	case n.Role == "phone" && !n.Enrolled:
		return "no profile yet"
	case n.Role == "phone" && !n.InSync:
		return "profile out of date"
	case n.Role == "phone":
		return "profile current"
	case !n.Enrolled:
		return "not joined"
	case n.Rollback != nil:
		return "rolled back"
	case n.InSync:
		return "in sync"
	case n.LastSeen == 0:
		return "joined, no heartbeat"
	default:
		return "applying"
	}
}

// heard is when a node was last heard of: its heartbeat, or for a phone the
// latest handshake a server reported.
func heard(n api.NodeInfo) int64 {
	if n.Role == "phone" {
		return n.LastHandshake
	}
	return n.LastSeen
}

// setupNodePhone prints a phone's new profile, which the WireGuard app
// imports as a file (the web UI shows it as a QR code too). It holds the
// phone's private key, which the server does not keep; the notes go to
// stderr, so `> phone.conf` saves the profile alone.
func setupNodePhone(fs *flag.FlagSet, e *env) action {
	client, asJSON := adminFlags(fs)
	exit := fs.String("exit", "", "send all of the phone's traffic through this exit server (default: only mesh traffic uses the tunnel)")
	return func(ctx context.Context, args []string) error {
		name, err := nodeName(args)
		if err != nil {
			return err
		}
		p, err := client().Phone(ctx, name, *exit)
		if err != nil {
			return err
		}
		if *asJSON {
			return printJSON(e.stdout, p)
		}
		fmt.Fprintf(e.stderr, "new profile for %s: shown once, and the one before it no longer works\n", p.Node.Name)
		printWarnings(e.stderr, p.Warnings)
		_, err = io.WriteString(e.stdout, p.Profile)
		return err
	}
}

// setupNodeCompose prints the compose file that runs a node's agent in a
// container; the server checks the choices and renders it, so the CLI and
// the web UI give the same file.
func setupNodeCompose(fs *flag.FlagSet, e *env) action {
	client, asJSON := adminFlags(fs)
	backend := fs.String("wg-backend", "", "WireGuard in the container: auto (container mode's default: the host's module, else /dev/net/tun), userspace (the NAS default; needs /dev/net/tun) or kernel (the host has the wireguard module)")
	hostDir := fs.String("host-dir", "", "host directory for the agent's keys and cached config, mounted at /data (default: "+deploy.DefaultNASData+" for a NAS, a named volume in container mode)")
	runtimeName := fs.String("runtime", "", "container mode: docker (default), podman or swarm (Swarm services attach)")
	noNames := fs.Bool("no-names", false, "container mode: leave out the runtime's socket, so the node's containers get no names on the mesh")
	return func(ctx context.Context, args []string) error {
		name, err := nodeName(args)
		if err != nil {
			return err
		}
		f, err := client().Compose(ctx, name, api.ComposeOptions{Backend: *backend, DataDir: *hostDir, Runtime: *runtimeName, NoNames: *noNames})
		if err != nil {
			return err
		}
		if *asJSON {
			return printJSON(e.stdout, f)
		}
		_, err = io.WriteString(e.stdout, f.YAML)
		return err
	}
}

// setupNodeNetwork prints the commands that create a node's mesh network on
// its host and the compose lines that attach a service (amendment A37).
func setupNodeNetwork(fs *flag.FlagSet, e *env) action {
	client, asJSON := adminFlags(fs)
	runtimeName := fs.String("runtime", "", "docker (default), podman or swarm (Swarm services attach)")
	return func(ctx context.Context, args []string) error {
		name, err := nodeName(args)
		if err != nil {
			return err
		}
		n, err := client().Network(ctx, name, *runtimeName)
		if err != nil {
			return err
		}
		if *asJSON {
			return printJSON(e.stdout, n)
		}
		fmt.Fprintf(e.stdout, "the containers of %s (%s, %s mode), with %s -- create the mesh network on its host once:\n", name, n.Subnet, n.Mode, n.Runtime)
		for _, l := range n.Commands {
			fmt.Fprintln(e.stdout, l)
		}
		fmt.Fprintln(e.stdout, "\nthen attach a service to it:")
		for _, l := range n.Service {
			fmt.Fprintln(e.stdout, l)
		}
		return nil
	}
}

func setupNodeShow(fs *flag.FlagSet, e *env) action {
	client, asJSON := adminFlags(fs)
	return func(ctx context.Context, args []string) error {
		name, err := nodeName(args)
		if err != nil {
			return err
		}
		d, err := client().Show(ctx, name)
		if err != nil {
			return err
		}
		if *asJSON {
			return printJSON(e.stdout, d)
		}
		n := d.Node
		tw := tabwriter.NewWriter(e.stdout, 0, 0, 2, ' ', 0)
		row := func(k, v string) {
			if v != "" {
				fmt.Fprintf(tw, "%s\t%s\n", k, v)
			}
		}
		row("name", fmt.Sprintf("%s (id %d)", n.Name, n.ID))
		row("role", n.Role)
		row("state", nodeState(n))
		row("mesh ip", n.MeshIP)
		if n.Mode == "container" {
			row("mode", "container: only its containers are on the mesh")
		}
		row("containers", n.ContainerSubnet)
		row("endpoint", n.Endpoint)
		if n.Role == "server" {
			row("listen port", strconv.Itoa(n.ListenPort))
			row("can exit", strconv.FormatBool(n.CanExit))
		}
		row("hub", n.Hub)
		row("exit", n.Exit)
		row("keepalive", strconv.Itoa(n.Keepalive))
		row("mtu", strconv.Itoa(n.MTU))
		row("host", n.Hostname)
		row("platform", joinNonEmpty("/", n.OS, n.Arch))
		row("agent", joinNonEmpty(", ", n.AgentVersion, n.WGBackend))
		row("version", fmt.Sprintf("applied %d, current %d", n.AppliedVersion, n.ConfigVersion))
		row("dns", n.DNS)
		if rb := n.Rollback; rb != nil {
			row("rollback", fmt.Sprintf("v%d -> v%d: %s", rb.From, rb.To, rb.Reason))
		}
		if n.Role == "phone" {
			row("last handshake", ago(n.LastHandshake, e.now()))
		} else {
			row("last seen", ago(n.LastSeen, e.now()))
		}
		if err := tw.Flush(); err != nil {
			return err
		}
		if len(d.Peers) > 0 {
			fmt.Fprintln(e.stdout, "\npeers as last reported:")
			tw = tabwriter.NewWriter(e.stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(tw, "PEER\tENDPOINT\tHANDSHAKE\tRX\tTX")
			for _, p := range d.Peers {
				fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n", p.Name, p.Endpoint, ago(p.LastHandshake, e.now()), bytesize(p.RxBytes), bytesize(p.TxBytes))
			}
			if err := tw.Flush(); err != nil {
				return err
			}
		}
		if d.Config != nil {
			fmt.Fprintln(e.stdout, "\ncomputed config:")
			if err := printJSON(e.stdout, d.Config); err != nil {
				return err
			}
		}
		printWarnings(e.stdout, d.Warnings)
		return nil
	}
}

func joinNonEmpty(sep string, parts ...string) string {
	out := ""
	for _, p := range parts {
		if p == "" {
			continue
		}
		if out != "" {
			out += sep
		}
		out += p
	}
	return out
}

// Keywords that clear a reference in `node set`.
const (
	clearEndpoint = "none"
	clearHub      = "default"
	clearExit     = "off"
)

func setupNodeSet(fs *flag.FlagSet, e *env) action {
	client, asJSON := adminFlags(fs)
	endpoint := fs.String("endpoint", "", "servers: public IPv4:port, or "+clearEndpoint+" to clear it")
	listenPort := fs.Int("listen-port", 0, "servers: WireGuard UDP port")
	hub := fs.String("hub", "", "spokes and phones: relay server, or "+clearHub+" for the default hub")
	canExit := fs.Bool("can-exit", false, "servers: may be chosen as an internet exit (--can-exit=false to revoke)")
	exit := fs.String("exit", "", "spokes: exit server, or "+clearExit)
	keepalive := fs.Int("keepalive", 0, "PersistentKeepalive seconds (0 turns it off)")
	mtu := fs.Int("mtu", 0, "interface MTU")
	containers := fs.Bool("containers", false, "host mode: put the node's containers on the mesh (--containers=false takes them off)")
	meshIP := fs.String("mesh-ip", "", "move the node to another mesh IP in the mesh (its peers, DNS name and agent follow)")
	return func(ctx context.Context, args []string) error {
		name, err := nodeName(args)
		if err != nil {
			return err
		}
		// Only flags given on the command line change anything.
		var req api.NodeSetRequest
		set := 0
		fs.Visit(func(f *flag.Flag) {
			set++
			switch f.Name {
			case "endpoint":
				req.Endpoint = cleared(*endpoint, clearEndpoint)
			case "listen-port":
				req.ListenPort = listenPort
			case "hub":
				req.Hub = cleared(*hub, clearHub)
			case "can-exit":
				req.CanExit = canExit
			case "exit":
				req.Exit = cleared(*exit, clearExit)
			case "keepalive":
				req.Keepalive = keepalive
			case "mtu":
				req.MTU = mtu
			case "containers":
				req.Containers = containers
			case "mesh-ip":
				req.MeshIP = meshIP
			default:
				set-- // --data, --json
			}
		})
		if set == 0 {
			return usagef("nothing to change; give at least one of --endpoint, --listen-port, --hub, --can-exit, --exit, --keepalive, --mtu, --containers, --mesh-ip")
		}
		return printChange(e.stdout, e.now().Location(), *asJSON, "updated", func() (api.NodeChange, error) { return client().Set(ctx, name, req) })
	}
}

func cleared(v, keyword string) *string {
	if v == keyword {
		v = ""
	}
	return &v
}

// nodeAction is the shared shape of disable, enable, rm and token.
func nodeAction(verb string) func(fs *flag.FlagSet, e *env) action {
	return func(fs *flag.FlagSet, e *env) action {
		client, asJSON := adminFlags(fs)
		return func(ctx context.Context, args []string) error {
			name, err := nodeName(args)
			if err != nil {
				return err
			}
			c := client()
			call := map[string]func() (api.NodeChange, error){
				"disable": func() (api.NodeChange, error) { return c.Disable(ctx, name) },
				"enable":  func() (api.NodeChange, error) { return c.Enable(ctx, name) },
				"rm":      func() (api.NodeChange, error) { return c.Remove(ctx, name) },
				"token":   func() (api.NodeChange, error) { return c.Token(ctx, name) },
			}[verb]
			done := map[string]string{"disable": "disabled", "enable": "enabled", "rm": "removed", "token": "issued a token for"}[verb]
			return printChange(e.stdout, e.now().Location(), *asJSON, done, call)
		}
	}
}

func printChange(w io.Writer, loc *time.Location, asJSON bool, done string, call func() (api.NodeChange, error)) error {
	ch, err := call()
	if err != nil {
		return err
	}
	if asJSON {
		return printJSON(w, ch)
	}
	if ch.Version != 0 {
		fmt.Fprintf(w, "%s %s; config version %d is on its way to the agents\n", done, ch.Node.Name, ch.Version)
	} else {
		fmt.Fprintf(w, "%s %s; no node's config changed\n", done, ch.Node.Name)
	}
	printEnrollment(w, loc, ch.Enrollment)
	printWarnings(w, ch.Notes)
	printWarnings(w, ch.Warnings)
	return nil
}
