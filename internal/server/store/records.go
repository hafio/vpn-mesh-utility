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
	NodeID    int64  `json:"node_id,omitempty"`
	CreatedAt int64  `json:"created_at"`
}

// FirewallRule opens a port in the mesh zone of one server (NodeID) or of
// every server (NodeID 0).
type FirewallRule struct {
	ID         int64  `json:"id"`
	NodeID     int64  `json:"node_id,omitempty"`
	Proto      string `json:"proto"`
	Port       int    `json:"port"`
	SourceCIDR string `json:"source_cidr"`
	Note       string `json:"note,omitempty"`
}

// DNSRecords lists custom records by name.
func (t *Tx) DNSRecords() ([]DNSRecord, error) {
	rows, err := t.query(`SELECT id, name, ip, node_id, created_at FROM dns_records ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("list dns records: %w", err)
	}
	defer rows.Close()
	var out []DNSRecord
	for rows.Next() {
		var (
			r    DNSRecord
			ip   sql.NullString
			node sql.NullInt64
		)
		if err := rows.Scan(&r.ID, &r.Name, &ip, &node, &r.CreatedAt); err != nil {
			return nil, fmt.Errorf("list dns records: %w", err)
		}
		r.IP, r.NodeID = ip.String, node.Int64
		out = append(out, r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list dns records: %w", err)
	}
	return out, nil
}

// InsertDNSRecord creates a record; r.ID 0 lets the database assign one.
func (t *Tx) InsertDNSRecord(r *DNSRecord) error {
	if r.CreatedAt == 0 {
		r.CreatedAt = t.now.Unix()
	}
	res, err := t.exec(`INSERT INTO dns_records (id, name, ip, node_id, created_at) VALUES (?, ?, ?, ?, ?)`,
		nullInt(r.ID), r.Name, nullStr(r.IP), nullInt(r.NodeID), r.CreatedAt)
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

// DeleteDNSRecord removes a record by id.
func (t *Tx) DeleteDNSRecord(id int64) error {
	res, err := t.exec(`DELETE FROM dns_records WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete dns record %d: %w", id, err)
	}
	return mustAffect(res, fmt.Sprintf("dns record %d", id))
}

// FirewallRules lists custom rules by id.
func (t *Tx) FirewallRules() ([]FirewallRule, error) {
	rows, err := t.query(`SELECT id, node_id, proto, port, source_cidr, note FROM firewall_rules ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("list firewall rules: %w", err)
	}
	defer rows.Close()
	var out []FirewallRule
	for rows.Next() {
		var (
			r    FirewallRule
			node sql.NullInt64
		)
		if err := rows.Scan(&r.ID, &node, &r.Proto, &r.Port, &r.SourceCIDR, &r.Note); err != nil {
			return nil, fmt.Errorf("list firewall rules: %w", err)
		}
		r.NodeID = node.Int64
		out = append(out, r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list firewall rules: %w", err)
	}
	return out, nil
}

// InsertFirewallRule creates a rule; r.ID 0 lets the database assign one.
func (t *Tx) InsertFirewallRule(r *FirewallRule) error {
	res, err := t.exec(`INSERT INTO firewall_rules (id, node_id, proto, port, source_cidr, note) VALUES (?, ?, ?, ?, ?, ?)`,
		nullInt(r.ID), nullInt(r.NodeID), r.Proto, r.Port, r.SourceCIDR, r.Note)
	if err != nil {
		return fmt.Errorf("insert firewall rule %s/%d: %w", r.Proto, r.Port, err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("insert firewall rule %s/%d: %w", r.Proto, r.Port, err)
	}
	r.ID = id
	return nil
}

// DeleteFirewallRule removes a rule by id.
func (t *Tx) DeleteFirewallRule(id int64) error {
	res, err := t.exec(`DELETE FROM firewall_rules WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete firewall rule %d: %w", id, err)
	}
	return mustAffect(res, fmt.Sprintf("firewall rule %d", id))
}
