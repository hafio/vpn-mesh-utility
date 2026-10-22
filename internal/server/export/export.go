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
	// Firewall is in exports made before segmentation only: their custom
	// firewall rules, which import as policies (store.LegacySegmentation).
	Firewall []FirewallRule `yaml:"firewall_rules,omitempty"`
	// Segmentation is the mesh firewall (amendment A43): groups (nodes by
	// name), objects, zones and policies, in order.
	Segmentation *store.Segmentation `yaml:"segmentation,omitempty"`
}

// Node is one node's desired state.
type Node struct {
	ID         int64  `yaml:"id"`
	Name       string `yaml:"name"`
	Role       string `yaml:"role"`
	MeshIP     string `yaml:"mesh_ip,omitempty"`
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
	// The IPv6 side (M8): Family is empty for dual stack; an IPv6-only node
	// has no mesh_ip.
	MeshIP6          string `yaml:"mesh_ip6,omitempty"`
	Family           string `yaml:"ip_family,omitempty"`
	Endpoint6        string `yaml:"endpoint6,omitempty"`
	ContainerSubnet6 string `yaml:"container_subnet6,omitempty"`
}

// DNSRecord is a custom record pointing at a node (by name) or a fixed IP
// (IPv4, IPv6 or both).
type DNSRecord struct {
	Name string `yaml:"name"`
	Node string `yaml:"node,omitempty"`
	IP   string `yaml:"ip,omitempty"`
	IP6  string `yaml:"ip6,omitempty"`
}

// FirewallRule is a custom rule of an export made before segmentation: a
// mesh-zone port on one server (by name) or on all.
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
	seg := s.EffectiveSegmentation()
	d := Document{Format: Format, Settings: s.Settings, Nodes: []Node{}, DNS: []DNSRecord{}, Segmentation: &seg}
	for _, n := range s.Nodes {
		d.Nodes = append(d.Nodes, Node{
			ID: n.ID, Name: n.Name, Role: n.Role, MeshIP: n.MeshIP, Endpoint: n.Endpoint, ListenPort: n.ListenPort,
			Hub: names[n.PrimaryHubID], CanExit: n.CanExit, Exit: names[n.ExitVia], Keepalive: n.Keepalive, MTU: n.MTU,
			Enrolled: n.Enrolled, Disabled: n.Disabled, WGPublicKey: n.WGPubKey, IdentityPublicKey: n.IdentityPubKey,
			OS: n.OS, Arch: n.Arch, Hostname: n.Hostname, WGBackend: n.WGBackend, CreatedAt: n.CreatedAt,
			Mode: n.Mode, ContainerSubnet: n.ContainerSubnet,
			MeshIP6: n.MeshIP6, Family: n.Family, Endpoint6: n.Endpoint6, ContainerSubnet6: n.ContainerSubnet6,
		})
	}
	sort.Slice(d.Nodes, func(i, j int) bool { return d.Nodes[i].ID < d.Nodes[j].ID })
	for _, r := range s.DNS {
		d.DNS = append(d.DNS, DNSRecord{Name: r.Name, Node: names[r.NodeID], IP: r.IP, IP6: r.IP6})
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
		DNS: []store.DNSRecord{}}
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
		if n.MeshIP == "" && n.MeshIP6 == "" {
			bad(fmt.Errorf("%s: needs a mesh_ip, a mesh_ip6 or both", f))
		}
		if n.MeshIP != "" {
			_, err = validate.IPv4(f+".mesh_ip", n.MeshIP)
			bad(err)
		}
		if n.MeshIP6 != "" {
			_, err = validate.IPv6(f+".mesh_ip6", n.MeshIP6)
			bad(err)
		}
		if n.Family != "" {
			_, err = validate.Enum(f+".ip_family", n.Family, proto.FamilyDual, proto.FamilyIPv4, proto.FamilyIPv6)
			bad(err)
		}
		if n.Endpoint != "" {
			_, err = validate.Endpoint(f+".endpoint", n.Endpoint)
			bad(err)
		}
		if n.Endpoint6 != "" {
			_, err = validate.Endpoint6(f+".endpoint6", n.Endpoint6)
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
		if n.ContainerSubnet6 != "" {
			_, err = validate.Prefix6(f+".container_subnet6", n.ContainerSubnet6)
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
			MeshIP6: n.MeshIP6, Family: n.Family, Endpoint6: n.Endpoint6, ContainerSubnet6: n.ContainerSubnet6,
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
		case store.SettingMeshCIDR6:
			// Empty is not chosen yet, off given up; the size the plan checks.
			if v != "" && v != store.MeshCIDR6Off {
				_, err := validate.Prefix6(f, v)
				bad(err)
			}
		case store.SettingUpstreamDNS:
			for _, ip := range strings.Split(v, ",") {
				_, err := validate.IP(f, strings.TrimSpace(ip))
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
		case (r.Node == "") == (r.IP == "" && r.IP6 == ""):
			bad(fmt.Errorf("%s: needs either a node or a fixed ip/ip6, not both or neither", f))
		case r.Node != "":
			rec.NodeID = ref(f+".node", r.Node)
		default:
			if r.IP != "" {
				_, err = validate.IPv4(f+".ip", r.IP)
				bad(err)
			}
			if r.IP6 != "" {
				_, err = validate.IPv6(f+".ip6", r.IP6)
				bad(err)
			}
			rec.IP, rec.IP6 = r.IP, r.IP6
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
		s.Firewall = append(s.Firewall, store.FirewallRule{ID: int64(i + 1), NodeID: ref(f+".node", r.Node), Proto: r.Proto,
			Port: r.Port, SourceCIDR: r.Source, Note: r.Note})
	}
	if seg := d.Segmentation; seg != nil {
		if len(d.Firewall) > 0 {
			bad(errors.New("firewall_rules: an export made before segmentation has them, a newer one a segmentation section; keep one"))
		}
		if err := seg.Validate(); err != nil {
			bad(fmt.Errorf("segmentation: %w", err))
		}
		for _, g := range seg.Groups {
			for _, n := range g.Nodes {
				ref(fmt.Sprintf("segmentation.groups[%s].nodes", g.Name), n)
			}
		}
		s.Segmentation = seg
	}
	if err := errors.Join(errs...); err != nil {
		return store.Snapshot{}, fmt.Errorf("the export is not valid, nothing was imported:\n%w", err)
	}
	return s, nil
}

// segDoc is the file `meshctl firewall show --yaml` writes and `meshctl
// firewall apply` reads: the segmentation section on its own.
type segDoc struct {
	Segmentation store.Segmentation `yaml:"segmentation"`
}

// WriteSegmentation encodes the mesh firewall's document as YAML.
func WriteSegmentation(w io.Writer, s store.Segmentation) error {
	var b bytes.Buffer
	b.WriteString("# meshctl mesh firewall: groups, objects, zones and policies (first match wins).\n" +
		"# Edit it and give it back to `meshctl firewall apply <file>`.\n")
	enc := yaml.NewEncoder(&b)
	enc.SetIndent(2)
	if err := enc.Encode(segDoc{Segmentation: s}); err != nil {
		return fmt.Errorf("encode the firewall: %w", err)
	}
	if err := enc.Close(); err != nil {
		return fmt.Errorf("encode the firewall: %w", err)
	}
	if _, err := w.Write(b.Bytes()); err != nil {
		return fmt.Errorf("write the firewall: %w", err)
	}
	return nil
}

// ReadSegmentation parses what WriteSegmentation wrote, refusing unknown
// fields, and checks it against its grammar.
func ReadSegmentation(r io.Reader) (store.Segmentation, error) {
	var d segDoc
	dec := yaml.NewDecoder(r)
	dec.KnownFields(true)
	if err := dec.Decode(&d); err != nil {
		return store.Segmentation{}, fmt.Errorf("read the firewall file: %w", err)
	}
	if err := d.Segmentation.Validate(); err != nil {
		return store.Segmentation{}, fmt.Errorf("the firewall file is not valid, nothing was changed:\n%w", err)
	}
	return d.Segmentation, nil
}
