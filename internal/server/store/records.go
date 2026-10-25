package store

import (
	"database/sql"
	"fmt"
)

// DNSRecord is a custom record in the mesh zone: a name pointing at either a
// node (NodeID, follows its mesh IP) or a fixed IP.
type DNSRecord struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	IP        string `json:"ip,omitempty"`
	IP6       string `json:"ip6,omitempty"`
	NodeID    int64  `json:"node_id,omitempty"`
	CreatedAt int64  `json:"created_at"`
}

// DNSRecords lists the mesh's custom records by name.
func (t *MeshTx) DNSRecords() ([]DNSRecord, error) {
	rows, err := t.query(`SELECT id, name, ip, ip6, node_id, created_at FROM dns_records WHERE mesh_id = ? ORDER BY name`, t.mesh)
	if err != nil {
		return nil, fmt.Errorf("list dns records: %w", err)
	}
	defer rows.Close()
	var out []DNSRecord
	for rows.Next() {
		var (
			r       DNSRecord
			ip, ip6 sql.NullString
			node    sql.NullInt64
		)
		if err := rows.Scan(&r.ID, &r.Name, &ip, &ip6, &node, &r.CreatedAt); err != nil {
			return nil, fmt.Errorf("list dns records: %w", err)
		}
		r.IP, r.IP6, r.NodeID = ip.String, ip6.String, node.Int64
		out = append(out, r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list dns records: %w", err)
	}
	return out, nil
}

// InsertDNSRecord creates a record of the mesh; r.ID 0 lets the database
// assign one.
func (t *MeshTx) InsertDNSRecord(r *DNSRecord) error {
	if r.CreatedAt == 0 {
		r.CreatedAt = t.now.Unix()
	}
	// A fixed record keeps ip non-NULL for 0001's check, '' when it has
	// IPv6 only (M8).
	var ip any = nullStr(r.IP)
	if r.NodeID == 0 && r.IP == "" && r.IP6 != "" {
		ip = ""
	}
	res, err := t.exec(`INSERT INTO dns_records (id, mesh_id, name, ip, ip6, node_id, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		nullInt(r.ID), t.mesh, r.Name, ip, nullStr(r.IP6), nullInt(r.NodeID), r.CreatedAt)
	if err != nil {
		return fmt.Errorf("insert dns record %q: %w", r.Name, err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("insert dns record %q: %w", r.Name, err)
	}
	r.ID = id
	return nil
}

// DeleteDNSRecord removes one of the mesh's records by id.
func (t *MeshTx) DeleteDNSRecord(id int64) error {
	res, err := t.exec(`DELETE FROM dns_records WHERE id = ? AND mesh_id = ?`, id, t.mesh)
	if err != nil {
		return fmt.Errorf("delete dns record %d: %w", id, err)
	}
	return mustAffect(res, fmt.Sprintf("dns record %d", id))
}
