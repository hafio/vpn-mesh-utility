// Package export writes the mesh's desired state as YAML (format
// meshctl-export/1) for review and commit to git (spec S8.1). The output is
// deterministic -- no timestamps of its own, slices in id order -- so an
// unchanged mesh exports byte-for-byte the same file.
//
// It holds desired state only: never join tokens, private keys, users,
// sessions or heartbeat state. Node ids are kept because every agent signs
// with its id; references (hub, exit, record and rule targets) are by name
// so the file reads well in a diff.
package export

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"

	"go.yaml.in/yaml/v3"

	"github.com/hafio/vpn-mesh-utility/internal/proto"
	"github.com/hafio/vpn-mesh-utility/internal/server/store"
	"github.com/hafio/vpn-mesh-utility/internal/validate"
)

// Format tags the document so a future change is detected, not misread.
const Format = "meshctl-export/1"

// Document is one export.
type Document struct {
	Format   string            `yaml:"format"`
	Settings map[string]string `yaml:"settings"`
	Nodes    []Node            `yaml:"nodes"`
	DNS      []DNSRecord       `yaml:"dns_records"`
	Firewall []FirewallRule    `yaml:"firewall_rules"`
}

// Node is one node's desired state.
type Node struct {
	ID         int64  `yaml:"id"`
	Name       string `yaml:"name"`
	Role       string `yaml:"role"`
	MeshIP     string `yaml:"mesh_ip"`
	Endpoint   string `yaml:"endpoint,omitempty"`
	ListenPort int    `yaml:"listen_port"`
	Hub        string `yaml:"hub,omitempty"`
	CanExit    bool   `yaml:"can_exit,omitempty"`
	Exit       string `yaml:"exit,omitempty"`
	Keepalive  int    `yaml:"keepalive"`
	MTU        int    `yaml:"mtu"`
	Enrolled   bool   `yaml:"enrolled,omitempty"`
	Disabled   bool   `yaml:"disabled,omitempty"`
	// Public keys only; they are what peers are configured with.
	WGPublicKey       string `yaml:"wg_public_key,omitempty"`
	IdentityPublicKey string `yaml:"identity_public_key,omitempty"`
	OS                string `yaml:"os,omitempty"`
	Arch              string `yaml:"arch,omitempty"`
	Hostname          string `yaml:"hostname,omitempty"`
	WGBackend         string `yaml:"wg_backend,omitempty"`
	CreatedAt         int64  `yaml:"created_at"`
	// Mode is empty for host mode; ContainerSubnet is the node's /24 for
	// containers (amendment A37).
	Mode            string `yaml:"mode,omitempty"`
	ContainerSubnet string `yaml:"container_subnet,omitempty"`
}

// DNSRecord is a custom record pointing at a node (by name) or a fixed IP.
type DNSRecord struct {
	Name string `yaml:"name"`
	Node string `yaml:"node,omitempty"`
	IP   string `yaml:"ip,omitempty"`
}

// FirewallRule opens a mesh-zone port on one server (by name) or on all.
type FirewallRule struct {
	Node   string `yaml:"node,omitempty"`
	Proto  string `yaml:"proto"`
	Port   int    `yaml:"port"`
	Source string `yaml:"source_cidr"`
	Note   string `yaml:"note,omitempty"`
}

// FromSnapshot converts the store's desired-state snapshot.
func FromSnapshot(s store.Snapshot) Document {
	names := map[int64]string{}
	for _, n := range s.Nodes {
		names[n.ID] = n.Name
	}
	d := Document{Format: Format, Settings: s.Settings, Nodes: []Node{}, DNS: []DNSRecord{}, Firewall: []FirewallRule{}}
	for _, n := range s.Nodes {
		d.Nodes = append(d.Nodes, Node{
			ID: n.ID, Name: n.Name, Role: n.Role, MeshIP: n.MeshIP, Endpoint: n.Endpoint, ListenPort: n.ListenPort,
			Hub: names[n.PrimaryHubID], CanExit: n.CanExit, Exit: names[n.ExitVia], Keepalive: n.Keepalive, MTU: n.MTU,
			Enrolled: n.Enrolled, Disabled: n.Disabled, WGPublicKey: n.WGPubKey, IdentityPublicKey: n.IdentityPubKey,
			OS: n.OS, Arch: n.Arch, Hostname: n.Hostname, WGBackend: n.WGBackend, CreatedAt: n.CreatedAt,
			Mode: n.Mode, ContainerSubnet: n.ContainerSubnet,
		})
	}
	sort.Slice(d.Nodes, func(i, j int) bool { return d.Nodes[i].ID < d.Nodes[j].ID })
	for _, r := range s.DNS {
		d.DNS = append(d.DNS, DNSRecord{Name: r.Name, Node: names[r.NodeID], IP: r.IP})
	}
	for _, r := range s.Firewall {
		d.Firewall = append(d.Firewall, FirewallRule{Node: names[r.NodeID], Proto: r.Proto, Port: r.Port, Source: r.SourceCIDR, Note: r.Note})
	}
	return d
}

// Write encodes d with a short header comment.
func Write(w io.Writer, d Document) error {
	var b bytes.Buffer
	b.WriteString("# meshctl export: the mesh's desired state. No secrets: tokens, private keys\n# and users are never exported.\n")
	enc := yaml.NewEncoder(&b)
	enc.SetIndent(2)
	if err := enc.Encode(d); err != nil {
		return fmt.Errorf("encode export: %w", err)
	}
	if err := enc.Close(); err != nil {
		return fmt.Errorf("encode export: %w", err)
	}
	if _, err := w.Write(b.Bytes()); err != nil {
		return fmt.Errorf("write export: %w", err)
	}
	return nil
}

// Read parses an export, refusing unknown fields and formats so a newer
// file is never half-read.
func Read(r io.Reader) (Document, error) {
	var d Document
	dec := yaml.NewDecoder(r)
	dec.KnownFields(true)
	if err := dec.Decode(&d); err != nil {
		return Document{}, fmt.Errorf("read export: %w", err)
	}
	if d.Format != Format {
		return Document{}, fmt.Errorf("read export: format %q is not %q; this meshctl cannot read it", d.Format, Format)
	}
	return d, nil
}

// ToSnapshot checks an export field by field and converts it to the store's
// snapshot, for `meshctl import`. Every value passes internal/validate,
// references by name must resolve, and only the desired settings are taken.
// Cross-node rules (one IP per node, a hub that is a server, ...) are the
// planner's: the reconciler refuses the import if they fail.
func ToSnapshot(d Document) (store.Snapshot, error) {
	var errs []error
	bad := func(err error) {
		if err != nil {
			errs = append(errs, err)
		}
	}
	s := store.Snapshot{Format: store.SnapshotFormat, Settings: map[string]string{}, Nodes: []store.SnapNode{},
		DNS: []store.DNSRecord{}, Firewall: []store.FirewallRule{}}
	ids := map[string]int64{}
	seen := map[int64]bool{}
	for i, n := range d.Nodes {
		f := fmt.Sprintf("nodes[%d]", i)
		if n.ID <= 0 || seen[n.ID] {
			bad(fmt.Errorf("%s.id: must be a positive id used once (got %d)", f, n.ID))
		}
		seen[n.ID] = true
		_, err := validate.Name(f+".name", n.Name)
		bad(err)
		if _, dup := ids[n.Name]; dup {
			bad(fmt.Errorf("%s.name: %q appears twice", f, n.Name))
		}
		ids[n.Name] = n.ID
		_, err = validate.Enum(f+".role", n.Role, string(proto.RoleServer), string(proto.RoleSpoke), string(proto.RolePhone))
		bad(err)
		_, err = validate.IPv4(f+".mesh_ip", n.MeshIP)
		bad(err)
		if n.Endpoint != "" {
			_, err = validate.Endpoint(f+".endpoint", n.Endpoint)
			bad(err)
		}
		_, err = validate.Port(f+".listen_port", n.ListenPort)
		bad(err)
		_, err = validate.IntRange(f+".keepalive", n.Keepalive, 0, 3600)
		bad(err)
		_, err = validate.IntRange(f+".mtu", n.MTU, proto.MinMTU, proto.MaxMTU)
		bad(err)
		for _, k := range [][2]string{{".wg_public_key", n.WGPublicKey}, {".identity_public_key", n.IdentityPublicKey}} {
			if k[1] != "" {
				_, err = validate.Key32(f+k[0], k[1])
				bad(err)
			}
		}
		for _, e := range [][]string{{".os", n.OS, "linux", "windows"}, {".arch", n.Arch, "amd64", "arm64"},
			{".wg_backend", n.WGBackend, "kernel", "userspace"}} {
			if e[1] != "" {
				_, err = validate.Enum(f+e[0], e[1], e[2:]...)
				bad(err)
			}
		}
		if n.Hostname != "" {
			_, err = validate.Hostname(f+".hostname", n.Hostname)
			bad(err)
		}
		if n.Mode != "" {
			_, err = validate.Enum(f+".mode", n.Mode, proto.ModeHost, proto.ModeContainer)
			bad(err)
		}
		if n.ContainerSubnet != "" {
			_, err = validate.Prefix4(f+".container_subnet", n.ContainerSubnet)
			bad(err)
		}
		if n.CreatedAt < 0 {
			bad(fmt.Errorf("%s.created_at: must not be negative", f))
		}
	}
	ref := func(field, name string) int64 {
		if name == "" {
			return 0
		}
		id, ok := ids[name]
		if !ok {
			bad(fmt.Errorf("%s: no node named %q in this export", field, name))
		}
		return id
	}
	for i, n := range d.Nodes {
		f := fmt.Sprintf("nodes[%d]", i)
		s.Nodes = append(s.Nodes, store.SnapNode{
			ID: n.ID, Name: n.Name, Role: n.Role, MeshIP: n.MeshIP, Endpoint: n.Endpoint, ListenPort: n.ListenPort,
			PrimaryHubID: ref(f+".hub", n.Hub), CanExit: n.CanExit, ExitVia: ref(f+".exit", n.Exit), Keepalive: n.Keepalive,
			MTU: n.MTU, Enrolled: n.Enrolled, Disabled: n.Disabled, WGPubKey: n.WGPublicKey, IdentityPubKey: n.IdentityPublicKey,
			OS: n.OS, Arch: n.Arch, Hostname: n.Hostname, WGBackend: n.WGBackend, CreatedAt: n.CreatedAt,
			Mode: n.Mode, ContainerSubnet: n.ContainerSubnet,
		})
	}

	keys := make([]string, 0, len(d.Settings))
	for k := range d.Settings {
		keys = append(keys, k)
	}
	sort.Strings(keys) // errors in a stable order
	for _, k := range keys {
		v := d.Settings[k]
		f := "settings." + k
		switch k {
		case store.SettingMeshCIDR:
			_, err := validate.Prefix4(f, v)
			bad(err)
		case store.SettingUpstreamDNS:
			for _, ip := range strings.Split(v, ",") {
				_, err := validate.IPv4(f, strings.TrimSpace(ip))
				bad(err)
			}
		case store.SettingDNSZone:
			_, err := validate.Zone(f, v)
			bad(err)
		case store.SettingDNSExtraDomains:
			// How they sit with the zone, the plan checks once restored.
			_, err := validate.Domains(f, "", store.SplitDomains(v))
			bad(err)
		case store.SettingDefaultMTU:
			mtu, err := strconv.Atoi(v)
			if err == nil {
				_, err = validate.IntRange(f, mtu, proto.MinMTU, proto.MaxMTU)
			}
			bad(err)
		case store.SettingDefaultHub:
			if v != "" {
				if id, err := strconv.ParseInt(v, 10, 64); err != nil || !seen[id] {
					bad(fmt.Errorf("%s: %q is not the id of a node in this export", f, v))
				}
			}
		default:
			bad(fmt.Errorf("%s: not a setting an export may carry", f))
			continue
		}
		s.Settings[k] = v
	}

	for i, r := range d.DNS {
		f := fmt.Sprintf("dns_records[%d]", i)
		_, err := validate.DNSName(f+".name", r.Name)
		bad(err)
		rec := store.DNSRecord{Name: r.Name}
		switch {
		case (r.Node == "") == (r.IP == ""):
			bad(fmt.Errorf("%s: needs exactly one of node and ip", f))
		case r.Node != "":
			rec.NodeID = ref(f+".node", r.Node)
		default:
			_, err = validate.IPv4(f+".ip", r.IP)
			bad(err)
			rec.IP = r.IP
		}
		s.DNS = append(s.DNS, rec)
	}
	for i, r := range d.Firewall {
		f := fmt.Sprintf("firewall_rules[%d]", i)
		_, err := validate.Enum(f+".proto", r.Proto, "tcp", "udp")
		bad(err)
		_, err = validate.Port(f+".port", r.Port)
		bad(err)
		_, err = validate.Prefix4(f+".source_cidr", r.Source)
		bad(err)
		_, err = validate.Text(f+".note", r.Note, 200)
		bad(err)
		s.Firewall = append(s.Firewall, store.FirewallRule{NodeID: ref(f+".node", r.Node), Proto: r.Proto, Port: r.Port,
			SourceCIDR: r.Source, Note: r.Note})
	}
	if err := errors.Join(errs...); err != nil {
		return store.Snapshot{}, fmt.Errorf("the export is not valid, nothing was imported:\n%w", err)
	}
	return s, nil
}
