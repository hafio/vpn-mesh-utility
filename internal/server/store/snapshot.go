package store

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/hafio/vpn-mesh-utility/internal/proto"
)

// SnapshotFormat tags the snapshot encoding so a future change is detected
// instead of mis-restored.
const SnapshotFormat = "meshctl-snapshot/1"

// Snapshot is the desired state of the whole mesh: what a config version
// records and what a rollback restores. It never holds users, sessions,
// tokens, heartbeat state or reconciler bookkeeping.
type Snapshot struct {
	Format   string            `json:"format"`
	Settings map[string]string `json:"settings"`
	Nodes    []SnapNode        `json:"nodes"`
	DNS      []DNSRecord       `json:"dns_records"`
	// Firewall is in snapshots made before segmentation only: their custom
	// firewall rules, which restore as policies (LegacySegmentation).
	Firewall []FirewallRule `json:"firewall_rules,omitempty"`
	// Segmentation is the mesh firewall (amendment A43); nil only in a
	// snapshot made before it existed.
	Segmentation *Segmentation `json:"segmentation"`
}

// EffectiveSegmentation is the segmentation s restores: its own, or, in a
// snapshot made before segmentation, its custom rules as policies.
func (s Snapshot) EffectiveSegmentation() Segmentation {
	if s.Segmentation != nil {
		return *s.Segmentation
	}
	names := map[int64]string{}
	for _, n := range s.Nodes {
		names[n.ID] = n.Name
	}
	return LegacySegmentation(s.Firewall, names)
}

// SnapNode is the desired-state subset of a node.
type SnapNode struct {
	ID             int64  `json:"id"`
	Name           string `json:"name"`
	Role           string `json:"role"`
	MeshIP         string `json:"mesh_ip,omitempty"`
	WGPubKey       string `json:"wg_pubkey,omitempty"`
	IdentityPubKey string `json:"identity_pubkey,omitempty"`
	Endpoint       string `json:"endpoint,omitempty"`
	ListenPort     int    `json:"listen_port"`
	PrimaryHubID   int64  `json:"primary_hub_id,omitempty"`
	CanExit        bool   `json:"can_exit,omitempty"`
	ExitVia        int64  `json:"exit_via,omitempty"`
	Keepalive      int    `json:"keepalive"`
	MTU            int    `json:"mtu"`
	Enrolled       bool   `json:"enrolled,omitempty"`
	Disabled       bool   `json:"disabled,omitempty"`
	OS             string `json:"os,omitempty"`
	Arch           string `json:"arch,omitempty"`
	Hostname       string `json:"hostname,omitempty"`
	WGBackend      string `json:"wg_backend,omitempty"`
	CreatedAt      int64  `json:"created_at"`
	// Mode is empty for proto.ModeHost, so a snapshot of a mesh without
	// container networks encodes as before (amendment A37).
	Mode            string `json:"mode,omitempty"`
	ContainerSubnet string `json:"container_subnet,omitempty"`
	// The IPv6 side (M8); Family is empty for proto.FamilyDual, so an IPv4
	// mesh's snapshot encodes as before.
	MeshIP6          string `json:"mesh_ip6,omitempty"`
	Family           string `json:"ip_family,omitempty"`
	Endpoint6        string `json:"endpoint6,omitempty"`
	ContainerSubnet6 string `json:"container_subnet6,omitempty"`
}

func snapNode(n Node) SnapNode {
	s := SnapNode{
		ID: n.ID, Name: n.Name, Role: n.Role, MeshIP: n.MeshIP, WGPubKey: n.WGPubKey,
		IdentityPubKey: n.IdentityPubKey, Endpoint: n.Endpoint, ListenPort: n.ListenPort,
		PrimaryHubID: n.PrimaryHubID, CanExit: n.CanExit, ExitVia: n.ExitVia, Keepalive: n.Keepalive,
		MTU: n.MTU, Enrolled: n.Enrolled, Disabled: n.Disabled, OS: n.OS, Arch: n.Arch,
		Hostname: n.Hostname, WGBackend: n.WGBackend, CreatedAt: n.CreatedAt, ContainerSubnet: n.ContainerSubnet,
		MeshIP6: n.MeshIP6, Endpoint6: n.Endpoint6, ContainerSubnet6: n.ContainerSubnet6,
	}
	if n.Mode != proto.ModeHost {
		s.Mode = n.Mode
	}
	if n.Family != proto.FamilyDual {
		s.Family = n.Family
	}
	return s
}

// Snapshot reads the current desired state.
func (t *Tx) Snapshot() (Snapshot, error) {
	s := Snapshot{Format: SnapshotFormat, Settings: map[string]string{}, Nodes: []SnapNode{}, DNS: []DNSRecord{}}
	all, err := t.Settings()
	if err != nil {
		return Snapshot{}, err
	}
	for _, k := range DesiredSettings {
		if v, ok := all[k]; ok {
			s.Settings[k] = v
		}
	}
	nodes, err := t.Nodes()
	if err != nil {
		return Snapshot{}, err
	}
	for _, n := range nodes {
		s.Nodes = append(s.Nodes, snapNode(n))
	}
	if s.DNS, err = t.DNSRecords(); err != nil {
		return Snapshot{}, err
	}
	seg, err := t.Segmentation()
	if err != nil {
		return Snapshot{}, err
	}
	s.Segmentation = &seg
	if s.DNS == nil {
		s.DNS = []DNSRecord{}
	}
	return s, nil
}

// Encode returns the canonical JSON of the snapshot (map keys are sorted by
// encoding/json; slices are in id order).
func (s Snapshot) Encode() ([]byte, error) {
	b, err := json.Marshal(s)
	if err != nil {
		return nil, fmt.Errorf("encode snapshot: %w", err)
	}
	return b, nil
}

// DecodeSnapshot parses a stored snapshot, refusing unknown fields and
// formats so a newer encoding is never half-restored.
func DecodeSnapshot(b []byte) (Snapshot, error) {
	var s Snapshot
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&s); err != nil {
		return Snapshot{}, fmt.Errorf("decode snapshot: %w", err)
	}
	if s.Format != SnapshotFormat {
		return Snapshot{}, fmt.Errorf("decode snapshot: format %q is not %q; this meshctl cannot restore it", s.Format, SnapshotFormat)
	}
	return s, nil
}

// Restore replaces the desired state with s: nodes absent from s are
// deleted, nodes in s are re-created or updated (keeping their ids, so
// references stay valid), and custom DNS records, segmentation and desired
// settings are replaced. Heartbeat and reconciler fields of surviving nodes
// are left alone; the caller recomputes configs afterwards.
func (t *Tx) Restore(s Snapshot) error {
	if s.Format != SnapshotFormat {
		return fmt.Errorf("restore: snapshot format %q is not %q", s.Format, SnapshotFormat)
	}
	current, err := t.Nodes()
	if err != nil {
		return err
	}
	keep := map[int64]bool{}
	for _, n := range s.Nodes {
		keep[n.ID] = true
	}
	exists := map[int64]bool{}
	for _, n := range current {
		if !keep[n.ID] {
			if err := t.DeleteNode(n.ID); err != nil {
				return fmt.Errorf("restore: %w", err)
			}
			continue
		}
		exists[n.ID] = true
	}
	// Pass 1: rows without cross-references (a node may point at one that is
	// restored later in the list).
	for _, n := range s.Nodes {
		mode, family := n.Mode, n.Family
		if mode == "" {
			mode = proto.ModeHost
		}
		if family == "" {
			family = proto.FamilyDual
		}
		args := []any{n.Name, n.Role, nullStr(n.MeshIP), nullStr(n.WGPubKey), nullStr(n.IdentityPubKey), n.Endpoint,
			n.ListenPort, boolInt(n.CanExit), n.Keepalive, n.MTU, boolInt(n.Enrolled), boolInt(n.Disabled),
			n.OS, n.Arch, n.Hostname, n.WGBackend, mode, n.ContainerSubnet,
			nullStr(n.MeshIP6), family, n.Endpoint6, n.ContainerSubnet6}
		if exists[n.ID] {
			_, err = t.exec(`UPDATE nodes SET name = ?, role = ?, mesh_ip = ?, wg_pubkey = ?, identity_pubkey = ?,
					endpoint = ?, listen_port = ?, can_exit = ?, keepalive = ?, mtu = ?, enrolled = ?, disabled = ?,
					os = ?, arch = ?, hostname = ?, wg_backend = ?, mode = ?, container_subnet = ?,
					mesh_ip6 = ?, ip_family = ?, endpoint6 = ?, container_subnet6 = ?,
					primary_hub_id = NULL, exit_via = NULL
				WHERE id = ?`, append(args, n.ID)...)
		} else {
			_, err = t.exec(`INSERT INTO nodes (name, role, mesh_ip, wg_pubkey, identity_pubkey, endpoint,
					listen_port, can_exit, keepalive, mtu, enrolled, disabled, os, arch, hostname, wg_backend,
					mode, container_subnet, mesh_ip6, ip_family, endpoint6, container_subnet6, id, created_at)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, append(args, n.ID, n.CreatedAt)...)
		}
		if err != nil {
			return fmt.Errorf("restore node %q: %w", n.Name, err)
		}
	}
	// Pass 2: hub and exit references.
	for _, n := range s.Nodes {
		if n.PrimaryHubID == 0 && n.ExitVia == 0 {
			continue
		}
		if _, err := t.exec(`UPDATE nodes SET primary_hub_id = ?, exit_via = ? WHERE id = ?`,
			nullInt(n.PrimaryHubID), nullInt(n.ExitVia), n.ID); err != nil {
			return fmt.Errorf("restore references of node %q: %w", n.Name, err)
		}
	}
	if _, err := t.exec(`DELETE FROM dns_records`); err != nil {
		return fmt.Errorf("restore dns records: %w", err)
	}
	for _, r := range s.DNS {
		if err := t.InsertDNSRecord(&r); err != nil {
			return fmt.Errorf("restore: %w", err)
		}
	}
	if err := t.SetSegmentation(s.EffectiveSegmentation()); err != nil {
		return fmt.Errorf("restore: %w", err)
	}
	for _, k := range DesiredSettings {
		v, ok := s.Settings[k]
		if !ok {
			// A snapshot older than the setting had its default; an empty
			// one is no row at all, as before the setting existed.
			if v, ok = settingDefaults[k]; !ok {
				continue
			}
			if v == "" {
				if err := t.DeleteSetting(k); err != nil {
					return fmt.Errorf("restore: %w", err)
				}
				continue
			}
		}
		if err := t.SetSetting(k, v); err != nil {
			return fmt.Errorf("restore: %w", err)
		}
	}
	return nil
}
