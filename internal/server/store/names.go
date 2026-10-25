package store

import (
	"database/sql"
	"fmt"
	"slices"
)

// ContainerName is one name an agent reported for a container on its mesh
// network (amendment A38). Operational state: never in a snapshot.
type ContainerName struct {
	NodeID int64
	Name   string
	IP     string
	Kind   string
}

// ContainerNames lists the reported names of every node of the mesh by
// name, node and address.
func (t *MeshTx) ContainerNames() ([]ContainerName, error) {
	rows, err := t.query(`SELECT c.node_id, c.name, c.ip, c.kind FROM container_names c JOIN nodes n ON n.id = c.node_id
		WHERE n.mesh_id = ? ORDER BY c.name, c.node_id, c.ip`, t.mesh)
	if err != nil {
		return nil, fmt.Errorf("list container names: %w", err)
	}
	defer rows.Close()
	var out []ContainerName
	for rows.Next() {
		var c ContainerName
		if err := rows.Scan(&c.NodeID, &c.Name, &c.IP, &c.Kind); err != nil {
			return nil, fmt.Errorf("list container names: %w", err)
		}
		out = append(out, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list container names: %w", err)
	}
	return out, nil
}

// SetContainerNames replaces a node's names with names (their NodeID is
// ignored) and reports whether they changed.
func (t *Tx) SetContainerNames(nodeID int64, names []ContainerName) (bool, error) {
	rows, err := t.query(`SELECT name, ip, kind FROM container_names WHERE node_id = ?`, nodeID)
	if err != nil {
		return false, fmt.Errorf("read the container names of node %d: %w", nodeID, err)
	}
	key := func(c ContainerName) string { return c.Name + "\x00" + c.IP + "\x00" + c.Kind }
	var have, want []string
	if err := scanAll(rows, "container names", func(r *sql.Rows) error {
		var c ContainerName
		if err := r.Scan(&c.Name, &c.IP, &c.Kind); err != nil {
			return err
		}
		have = append(have, key(c))
		return nil
	}); err != nil {
		return false, err
	}
	for _, c := range names {
		want = append(want, key(c))
	}
	slices.Sort(have)
	slices.Sort(want)
	if slices.Equal(have, want) {
		return false, nil
	}
	if _, err := t.exec(`DELETE FROM container_names WHERE node_id = ?`, nodeID); err != nil {
		return false, fmt.Errorf("replace the container names of node %d: %w", nodeID, err)
	}
	for _, c := range names {
		if _, err := t.exec(`INSERT INTO container_names (node_id, name, ip, kind) VALUES (?, ?, ?, ?)`,
			nodeID, c.Name, c.IP, c.Kind); err != nil {
			return false, fmt.Errorf("store container name %q of node %d: %w", c.Name, nodeID, err)
		}
	}
	return true, nil
}
