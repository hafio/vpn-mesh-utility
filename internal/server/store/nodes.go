package store

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/hafio/vpn-mesh-utility/internal/proto"
)

// Node is one row of the nodes table. Optional references and keys use their
// zero value for SQL NULL (0 for an id, "" for a key).
type Node struct {
	ID   int64
	Name string
	Role string

	// Reported by the agent at join.
	OS           string
	Arch         string
	Hostname     string
	WGBackend    string
	AgentVersion string

	MeshIP         string
	WGPubKey       string
	IdentityPubKey string
	Endpoint       string
	ListenPort     int
	PrimaryHubID   int64
	CanExit        bool
	ExitVia        int64
	Keepalive      int
	MTU            int
	Enrolled       bool
	Disabled       bool
	// Mode is proto.ModeHost or proto.ModeContainer; ContainerSubnet is the
	// node's /24 for containers, empty while they are not on the mesh
	// (amendment A37).
	Mode            string
	ContainerSubnet string

	// Written by the reconciler.
	ConfigHash    string
	ConfigVersion int64

	// From the last heartbeat.
	AppliedVersion int64
	AppliedHash    string
	LastSeen       int64
	ObservedIP     string
	StatusJSON     string

	// From the relay failover monitor (amendment A40): when it judged this
	// server down (0: up), and since when it has looked healthy again.
	DownSince    int64
	HealthySince int64

	CreatedAt int64
}

const nodeCols = `id, name, role, os, arch, hostname, wg_backend, agent_version,
	mesh_ip, wg_pubkey, identity_pubkey, endpoint, listen_port, primary_hub_id,
	can_exit, exit_via, keepalive, mtu, enrolled, disabled, config_hash,
	config_version, applied_version, applied_hash, last_seen, observed_ip,
	status_json, created_at, mode, container_subnet, down_since, healthy_since`

type scanner interface {
	Scan(dest ...any) error
}

func scanNode(r scanner) (Node, error) {
	var (
		n                       Node
		wgKey, idKey            sql.NullString
		hub, exitVia, lastSeen  sql.NullInt64
		canExit, enrolled, disb int
	)
	err := r.Scan(&n.ID, &n.Name, &n.Role, &n.OS, &n.Arch, &n.Hostname, &n.WGBackend, &n.AgentVersion,
		&n.MeshIP, &wgKey, &idKey, &n.Endpoint, &n.ListenPort, &hub,
		&canExit, &exitVia, &n.Keepalive, &n.MTU, &enrolled, &disb, &n.ConfigHash,
		&n.ConfigVersion, &n.AppliedVersion, &n.AppliedHash, &lastSeen, &n.ObservedIP,
		&n.StatusJSON, &n.CreatedAt, &n.Mode, &n.ContainerSubnet, &n.DownSince, &n.HealthySince)
	if err != nil {
		return Node{}, err
	}
	n.WGPubKey, n.IdentityPubKey = wgKey.String, idKey.String
	n.PrimaryHubID, n.ExitVia, n.LastSeen = hub.Int64, exitVia.Int64, lastSeen.Int64
	n.CanExit, n.Enrolled, n.Disabled = canExit == 1, enrolled == 1, disb == 1
	return n, nil
}

// Nodes returns every node, ordered by id.
func (t *Tx) Nodes() ([]Node, error) {
	rows, err := t.query(`SELECT ` + nodeCols + ` FROM nodes WHERE mesh_id = 1 ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("list nodes: %w", err)
	}
	defer rows.Close()
	var out []Node
	for rows.Next() {
		n, err := scanNode(rows)
		if err != nil {
			return nil, fmt.Errorf("list nodes: %w", err)
		}
		out = append(out, n)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list nodes: %w", err)
	}
	return out, nil
}

// NodeByID returns one node or ErrNotFound.
func (t *Tx) NodeByID(id int64) (Node, error) {
	n, err := scanNode(t.queryRow(`SELECT `+nodeCols+` FROM nodes WHERE id = ?`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return Node{}, fmt.Errorf("node %d: %w", id, ErrNotFound)
	}
	if err != nil {
		return Node{}, fmt.Errorf("read node %d: %w", id, err)
	}
	return n, nil
}

// NodeByName returns one node or ErrNotFound.
func (t *Tx) NodeByName(name string) (Node, error) {
	n, err := scanNode(t.queryRow(`SELECT `+nodeCols+` FROM nodes WHERE mesh_id = 1 AND name = ?`, name))
	if errors.Is(err, sql.ErrNoRows) {
		return Node{}, fmt.Errorf("node %q: %w", name, ErrNotFound)
	}
	if err != nil {
		return Node{}, fmt.Errorf("read node %q: %w", name, err)
	}
	return n, nil
}

// InsertNode creates a node from its desired fields and sets n.ID and
// n.CreatedAt. Operational fields start at their defaults.
func (t *Tx) InsertNode(n *Node) error {
	n.CreatedAt = t.now.Unix()
	if n.Mode == "" {
		n.Mode = proto.ModeHost
	}
	res, err := t.exec(`INSERT INTO nodes (name, role, mesh_ip, wg_pubkey, identity_pubkey, endpoint,
			listen_port, primary_hub_id, can_exit, exit_via, keepalive, mtu, enrolled, disabled,
			os, arch, hostname, wg_backend, agent_version, created_at, mode, container_subnet)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		n.Name, n.Role, n.MeshIP, nullStr(n.WGPubKey), nullStr(n.IdentityPubKey), n.Endpoint,
		n.ListenPort, nullInt(n.PrimaryHubID), boolInt(n.CanExit), nullInt(n.ExitVia), n.Keepalive, n.MTU,
		boolInt(n.Enrolled), boolInt(n.Disabled),
		n.OS, n.Arch, n.Hostname, n.WGBackend, n.AgentVersion, n.CreatedAt, n.Mode, n.ContainerSubnet)
	if err != nil {
		return fmt.Errorf("insert node %q: %w", n.Name, err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("insert node %q: %w", n.Name, err)
	}
	n.ID = id
	return nil
}

// UpdateNode writes the desired fields of n (everything an operator or a join
// may change). Reconciler bookkeeping and heartbeat fields are untouched.
func (t *Tx) UpdateNode(n Node) error {
	res, err := t.exec(`UPDATE nodes SET name = ?, mesh_ip = ?, wg_pubkey = ?, identity_pubkey = ?,
			endpoint = ?, listen_port = ?, primary_hub_id = ?, can_exit = ?, exit_via = ?,
			keepalive = ?, mtu = ?, enrolled = ?, disabled = ?, os = ?, arch = ?, hostname = ?,
			wg_backend = ?, agent_version = ?, container_subnet = ?
		WHERE id = ?`,
		n.Name, n.MeshIP, nullStr(n.WGPubKey), nullStr(n.IdentityPubKey),
		n.Endpoint, n.ListenPort, nullInt(n.PrimaryHubID), boolInt(n.CanExit), nullInt(n.ExitVia),
		n.Keepalive, n.MTU, boolInt(n.Enrolled), boolInt(n.Disabled), n.OS, n.Arch, n.Hostname,
		n.WGBackend, n.AgentVersion, n.ContainerSubnet, n.ID)
	if err != nil {
		return fmt.Errorf("update node %q: %w", n.Name, err)
	}
	return mustAffect(res, fmt.Sprintf("node %d", n.ID))
}

// SetConfigState records the reconciler's view of a node's computed config.
func (t *Tx) SetConfigState(id int64, hash string, version int64) error {
	res, err := t.exec(`UPDATE nodes SET config_hash = ?, config_version = ? WHERE id = ?`, hash, version, id)
	if err != nil {
		return fmt.Errorf("set config state of node %d: %w", id, err)
	}
	return mustAffect(res, fmt.Sprintf("node %d", id))
}

// SetHubState records the relay failover monitor's view of a server
// (amendment A40): down since (0: up) and healthy again since (0: not).
// It is operational state, never a config version by itself.
func (t *Tx) SetHubState(id, downSince, healthySince int64) error {
	res, err := t.exec(`UPDATE nodes SET down_since = ?, healthy_since = ? WHERE id = ?`, downSince, healthySince, id)
	if err != nil {
		return fmt.Errorf("record the failover state of node %d: %w", id, err)
	}
	return mustAffect(res, fmt.Sprintf("node %d", id))
}

// DeleteNode removes a node. Its tokens, peer status and node-bound DNS
// records go with it; spokes that used it as hub or exit fall back (the
// references are SET NULL).
func (t *Tx) DeleteNode(id int64) error {
	res, err := t.exec(`DELETE FROM nodes WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete node %d: %w", id, err)
	}
	return mustAffect(res, fmt.Sprintf("node %d", id))
}

func mustAffect(res sql.Result, what string) error {
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: %w", what, err)
	}
	if n == 0 {
		return fmt.Errorf("%s: %w", what, ErrNotFound)
	}
	return nil
}
