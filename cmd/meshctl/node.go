package main

import (
	"cmp"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

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
		{name: "disable", args: "<name>", summary: "take a node out of every peer list (its keys stay registered)", setup: nodeAction("disable")},
		{name: "enable", args: "<name>", summary: "put a disabled node back", setup: nodeAction("enable")},
		{name: "list", summary: "list the nodes and the mesh plan's warnings", setup: setupNodeList},
		{name: "phone", args: "<name>", summary: "print a new WireGuard profile for a phone (a new key: its previous profile stops working)", setup: setupNodePhone},
		{name: "rm", args: "<name>", summary: "remove a node; every peer drops it on the next apply", setup: nodeAction("rm")},
		{name: "set", args: "<name>", summary: "change a node's endpoint, hub, exit, keepalive, MTU, container network, mesh IP or groups", setup: setupNodeSet},
		{name: "show", args: "<name>", summary: "show a node, its container network's values, its computed config and its peers", setup: setupNodeShow},
		{name: "token", args: "<name>", summary: "issue a new join token (revokes unused ones; cuts an enrolled node off until it joins with it)",
			setup: nodeAction("token")},
	}
}

// adminFlags registers --data, --json and --mesh and returns the admin
// client constructor, for the mesh --mesh names (amendment A49).
func adminFlags(fs *flag.FlagSet) (func() *api.AdminClient, *bool) {
	client, asJSON := socketFlagsOnly(fs)
	mesh := fs.String("mesh", "", "the mesh to work on; needed when the control plane has more than one (meshctl mesh list)")
	return func() *api.AdminClient { return client().In(*mesh) }, asJSON
}

// socketFlagsOnly registers --data and --json and returns the admin client
// constructor.
func socketFlagsOnly(fs *flag.FlagSet) (func() *api.AdminClient, *bool) {
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
	listenPort := fs.Int("listen-port", 0, "servers: WireGuard UDP port (default 51820, plus the mesh's slot)")
	hub := fs.String("hub", "", "spokes and phones: the server that relays for it (default: the default hub)")
	canExit := fs.Bool("can-exit", false, "servers: may be chosen as an internet exit")
	alwaysOn := fs.Bool("always-on", false, "spokes: allocate from the always-on block (NAS)")
	keepalive := fs.Int("keepalive", 0, "spokes and phones: PersistentKeepalive seconds (default 25)")
	mtu := fs.Int("mtu", 0, "interface MTU (default: the mesh default, 1420)")
	mode := fs.String("mode", "host", "host (WireGuard on the host) or container (in the agent's container: only its containers are on the mesh)")
	containers := fs.Bool("containers", false, "host mode: put the node's containers on the mesh too (container mode always does)")
	groups := fs.String("groups", "", "comma-separated firewall groups to put the node in (a new name makes a new group)")
	family := fs.String("family", "", "dual (default: IPv4 and, once the mesh has an IPv6 prefix, IPv6), ipv4 or ipv6: the node's mesh addresses and the endpoints it dials")
	ip6 := fs.String("ip6", "", "fixed IPv6 mesh address (default: the lowest free one in the role's block, when the mesh has an IPv6 prefix)")
	endpoint6 := fs.String("endpoint6", "", "servers: public [IPv6]:port (spokes dial it when IPv4 does not work)")
	containers6 := fs.Bool("containers6", false, "put the containers' IPv6 on the mesh too (needs --containers or container mode)")
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
			Groups: splitComma(*groups), Family: *family, IP6: *ip6, Endpoint6: *endpoint6, Containers6: *containers6,
		})
		if err != nil {
			return err
		}
		if *asJSON {
			return printJSON(e.stdout, ch)
		}
		fmt.Fprintf(e.stdout, "added %s (%s, mesh %s)\n", ch.Node.Name, ch.Node.Role, addresses(ch.Node))
		if k := joinNonEmpty(", ", ch.Node.ContainerSubnet, ch.Node.ContainerSubnet6); k != "" {
			fmt.Fprintf(e.stdout, "its containers: %s\n", k)
		}
		if ch.Node.Mode == "container" {
			fmt.Fprintf(e.stdout, "\ncontainer mode: run its agent from a compose file of your own (docs/containers.md);\nits network's values:\n  meshctl node show %s\n", ch.Node.Name)
		}
		printEnrollment(e.stdout, e.now().Location(), ch.Enrollment)
		if ch.Node.Role == "phone" {
			fmt.Fprintf(e.stdout, "\nmake its profile for the WireGuard app:\n  meshctl node phone %s [--exit <server>]\nor as a QR code on its page in the web UI\n", ch.Node.Name)
		}
		printWarnings(e.stdout, ch.Warnings)
		return nil
	}
}

// addresses is a node's mesh addresses as the CLI shows them: "IP
// 10.77.2.1", "IP 10.77.2.1 and fd10:77::2:1", "IPv6 fd10:77::2:1".
func addresses(n api.NodeInfo) string {
	switch {
	case n.MeshIP != "" && n.MeshIP6 != "":
		return "IP " + n.MeshIP + " and " + n.MeshIP6
	case n.MeshIP6 != "":
		return "IPv6 " + n.MeshIP6
	}
	return "IP " + n.MeshIP
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
			ep, ip := cmp.Or(n.Endpoint, n.Endpoint6, "-"), cmp.Or(n.MeshIP, n.MeshIP6)
			fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\tv%d/v%d\t%s\n", n.Name, n.Role, ip, ep, nodeState(n),
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
		row("mesh ipv6", n.MeshIP6)
		if n.Family != "" && n.Family != "dual" {
			row("family", n.Family+" only")
		}
		if len(n.HostFamilies) > 0 {
			row("host can use", strings.Join(n.HostFamilies, " and ")+" (its agent's report)")
		}
		if n.Mode == "container" {
			row("mode", "container: only its containers are on the mesh")
		}
		row("containers", joinNonEmpty(", ", n.ContainerSubnet, n.ContainerSubnet6))
		row("groups", strings.Join(n.Groups, ", "))
		row("endpoint", n.Endpoint)
		row("endpoint6", n.Endpoint6)
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
		row("version", fmt.Sprintf("applied v%d, current v%d", n.AppliedVersion, n.ConfigVersion))
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
		if err := printContainers(e.stdout, d.Containers); err != nil {
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

// printContainers shows a node's container network as its host's runtime
// creates it (amendment A48): the values docs/containers.md's samples take.
func printContainers(w io.Writer, k *api.ContainerNetwork) error {
	if k == nil {
		return nil
	}
	fmt.Fprintln(w, "\ncontainers on the mesh (docs/containers.md has the samples):")
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	row := func(key, v string) {
		if v != "" {
			fmt.Fprintf(tw, "  %s\t%s\n", key, v)
		}
	}
	network := k.Name
	if k.Mode != "container" {
		network += " (the agent owns its bridge " + k.BridgeName + ")"
	}
	row("network", network)
	row("subnet", joinNonEmpty(", ", k.Network, k.Network6))
	row("ip range", joinNonEmpty(", ", k.Range, k.Range6))
	row("gateway", joinNonEmpty(", ", k.Bridge, k.Bridge6))
	if k.Mode == "container" {
		row("agent", joinNonEmpty(", ", k.Gateway, k.Gateway6))
	}
	row("dns", k.Gateway+", search "+k.Zone+", ndots:1")
	row("swarm dns", strings.Join(k.Resolvers, ", "))
	switch {
	case k.ListenPort != 0 && k.PublishPort != 0:
		row("published", fmt.Sprintf("%d/udp on the host, to %d/udp", k.PublishPort, k.ListenPort))
	case k.ListenPort != 0:
		row("published", "none known: set the node's endpoint")
	}
	return tw.Flush()
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
	groups := fs.String("groups", "", "the node's firewall groups, comma separated, replacing the ones it is in (\"\" takes it out of all)")
	family := fs.String("family", "", "dual, ipv4 or ipv6: allocates the address a family adds, drops the one it takes away")
	meshIP6 := fs.String("mesh-ip6", "", "move the node to another IPv6 mesh address in the mesh's IPv6 prefix")
	endpoint6 := fs.String("endpoint6", "", "servers: public [IPv6]:port, or "+clearEndpoint+" to clear it")
	containers6 := fs.Bool("containers6", false, "put the containers' IPv6 on the mesh (--containers6=false takes it off)")
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
			case "groups":
				g := splitComma(*groups)
				req.Groups = &g
			case "family":
				req.Family = family
			case "mesh-ip6":
				req.MeshIP6 = meshIP6
			case "endpoint6":
				req.Endpoint6 = cleared(*endpoint6, clearEndpoint)
			case "containers6":
				req.Containers6 = containers6
			default:
				set-- // --data, --json, --mesh
			}
		})
		if set == 0 {
			return usagef("nothing to change; give at least one of --endpoint, --listen-port, --hub, --can-exit, --exit, --keepalive, --mtu, --containers, --mesh-ip, --groups, --family, --mesh-ip6, --endpoint6, --containers6")
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
		fmt.Fprintf(w, "%s %s; config v%d is on its way to the agents\n", done, ch.Node.Name, ch.Version)
	} else {
		fmt.Fprintf(w, "%s %s; no node's config changed\n", done, ch.Node.Name)
	}
	printEnrollment(w, loc, ch.Enrollment)
	printWarnings(w, ch.Notes)
	printWarnings(w, ch.Warnings)
	return nil
}

// splitComma is a comma-separated flag's values, blanks dropped.
func splitComma(s string) []string {
	out := []string{}
	for _, v := range strings.Split(s, ",") {
		if v = strings.TrimSpace(v); v != "" {
			out = append(out, v)
		}
	}
	return out
}
