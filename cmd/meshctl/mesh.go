package main

import (
	"cmp"
	"context"
	"flag"
	"fmt"
	"io"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/hafio/vpn-mesh-utility/internal/server/api"
)

// The mesh commands manage the control plane's meshes (amendment A49) over
// <data>/admin.sock, as the node commands do. Every other admin command
// takes --mesh to say which mesh it works on.

func meshCommands() []command {
	return []command{
		{name: "add", args: "<name>", summary: "add a mesh with its own network, DNS zone and IPv6 prefix", setup: setupMeshAdd},
		{name: "list", summary: "list the control plane's meshes", setup: setupMeshList},
		{name: "rm", args: "<name>", summary: "remove a mesh without nodes, with its settings, history and policies", setup: setupMeshRemove},
		{name: "set", args: "<name>", summary: "rename a mesh or change its settings", setup: setupMeshSet},
		{name: "show", args: "<name>", summary: "show a mesh and its settings", setup: setupMeshShow},
	}
}

func setupMeshAdd(fs *flag.FlagSet, e *env) action {
	client, asJSON := socketFlagsOnly(fs)
	cidr := fs.String("mesh-cidr", "", "the mesh's IPv4 network, a /16 to /22 (default: the first free 10.x.0.0/16)")
	zone := fs.String("dns-zone", "", "the mesh's DNS zone (default: the mesh's name)")
	ipv6 := fs.String("ipv6", api.IPv6Auto, "the mesh's IPv6 prefix: a /48 to /64, auto (a random unique local /64) or off")
	return func(ctx context.Context, args []string) error {
		name, err := nodeName(args)
		if err != nil {
			return err
		}
		u, err := client().AddMesh(ctx, api.MeshAddRequest{Name: name, MeshCIDR: *cidr, Zone: *zone, IPv6: *ipv6})
		if err != nil {
			return err
		}
		if *asJSON {
			return printJSON(e.stdout, u)
		}
		m := u.Mesh
		fmt.Fprintf(e.stdout, "added mesh %s in slot %d: %s, %s, zone %s\n", m.Name, m.Slot, m.MeshCIDR, cmp.Or(m.MeshCIDR6, "no IPv6"), m.Zone)
		fmt.Fprintf(e.stdout, "add its nodes with `meshctl node add --mesh %s <name> --role ...`\n", m.Name)
		return nil
	}
}

func setupMeshList(fs *flag.FlagSet, e *env) action {
	client, asJSON := socketFlagsOnly(fs)
	return func(ctx context.Context, args []string) error {
		if err := wantArgs(args, 0, ""); err != nil {
			return err
		}
		l, err := client().Meshes(ctx)
		if err != nil {
			return err
		}
		if *asJSON {
			return printJSON(e.stdout, l)
		}
		tw := tabwriter.NewWriter(e.stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(tw, "NAME\tSLOT\tNODES\tIPV4\tIPV6\tZONE")
		for _, m := range l.Meshes {
			fmt.Fprintf(tw, "%s\t%d\t%d\t%s\t%s\t%s\n", m.Name, m.Slot, m.Nodes, m.MeshCIDR, cmp.Or(m.MeshCIDR6, "-"), m.Zone)
		}
		return tw.Flush()
	}
}

func setupMeshShow(fs *flag.FlagSet, e *env) action {
	client, asJSON := socketFlagsOnly(fs)
	return func(ctx context.Context, args []string) error {
		name, err := nodeName(args)
		if err != nil {
			return err
		}
		m, err := client().ShowMesh(ctx, name)
		if err != nil {
			return err
		}
		if *asJSON {
			return printJSON(e.stdout, m)
		}
		return printMesh(e.stdout, e.now().Location(), m)
	}
}

// printMesh writes a mesh as `mesh show` and `mesh set` print it.
func printMesh(w io.Writer, loc *time.Location, m api.MeshInfo) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	row := func(k, v string) { fmt.Fprintf(tw, "%s\t%s\n", k, v) }
	row("name", m.Name)
	row("uuid", m.UUID)
	row("slot", strconv.Itoa(m.Slot))
	row("created", time.Unix(m.CreatedAt, 0).In(loc).Format(time.RFC3339))
	row("nodes", strconv.Itoa(m.Nodes))
	row("ipv4 network", m.MeshCIDR)
	row("ipv6 prefix", cmp.Or(m.MeshCIDR6, "none"))
	row("dns zone", m.Zone)
	row("extra domains", cmp.Or(strings.Join(m.Domains, ", "), "none"))
	row("upstream dns", strings.Join(m.Upstream, ", "))
	row("default mtu", strconv.Itoa(m.DefaultMTU))
	row("default hub", cmp.Or(m.DefaultHub, "none"))
	return tw.Flush()
}

func setupMeshSet(fs *flag.FlagSet, e *env) action {
	client, asJSON := socketFlagsOnly(fs)
	rename := fs.String("name", "", "a new name for the mesh (its uuid and slot stay)")
	cidr := fs.String("mesh-cidr", "", "the mesh's IPv4 network, a /16 to /22; only while the mesh has no nodes")
	zone := fs.String("dns-zone", "", "the mesh's DNS zone")
	domains := fs.String("dns-extra-domains", "", "extra DNS domains every node resolves through the mesh, comma separated (\"\" for none)")
	upstream := fs.String("upstream-dns", "", "the resolvers the mesh's resolvers forward to, one to four IPv4 or IPv6 addresses, comma separated")
	mtu := fs.Int("default-mtu", 0, "the MTU of nodes that set none")
	hub := fs.String("default-hub", "", "the server spokes and phones without a hub relay through (\"\" for none)")
	return func(ctx context.Context, args []string) error {
		name, err := nodeName(args)
		if err != nil {
			return err
		}
		// Only flags given on the command line change anything.
		var req api.MeshSetRequest
		set := 0
		fs.Visit(func(f *flag.Flag) {
			set++
			switch f.Name {
			case "name":
				req.Name = rename
			case "mesh-cidr":
				req.MeshCIDR = cidr
			case "dns-zone":
				req.Zone = zone
			case "dns-extra-domains":
				d := splitComma(*domains)
				req.Domains = &d
			case "upstream-dns":
				u := splitComma(*upstream)
				req.Upstream = &u
			case "default-mtu":
				req.DefaultMTU = mtu
			case "default-hub":
				req.DefaultHub = hub
			default:
				set-- // --data, --json
			}
		})
		if set == 0 {
			return usagef("nothing to change; give at least one of --name, --mesh-cidr, --dns-zone, --dns-extra-domains, --upstream-dns, --default-mtu, --default-hub")
		}
		u, err := client().SetMesh(ctx, name, req)
		if err != nil {
			return err
		}
		if *asJSON {
			return printJSON(e.stdout, u)
		}
		if u.Version != 0 {
			fmt.Fprintf(e.stdout, "updated mesh %s; config v%d is on its way to the agents\n", u.Mesh.Name, u.Version)
		} else {
			fmt.Fprintf(e.stdout, "updated mesh %s; no node's config changed\n", u.Mesh.Name)
		}
		printWarnings(e.stdout, u.Warnings)
		return nil
	}
}

func setupMeshRemove(fs *flag.FlagSet, e *env) action {
	client, asJSON := socketFlagsOnly(fs)
	return func(ctx context.Context, args []string) error {
		name, err := nodeName(args)
		if err != nil {
			return err
		}
		if err := client().RemoveMesh(ctx, name); err != nil {
			return err
		}
		if *asJSON {
			return printJSON(e.stdout, map[string]string{"removed": name})
		}
		fmt.Fprintf(e.stdout, "removed mesh %s with its settings, history and policies; its nightly exports stay on disk\n", name)
		return nil
	}
}
