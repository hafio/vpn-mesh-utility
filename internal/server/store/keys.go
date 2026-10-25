package store

import (
	"fmt"
)

// RetireKey moves identity key pub into node id's key history (amendment
// A50): a new token cuts it off, or the node joined again with a fresh one.
// A key retired before keeps its first time; an empty key is no key.
func (t *Tx) RetireKey(id int64, pub string, at int64) error {
	if pub == "" {
		return nil
	}
	if _, err := t.exec(`INSERT INTO node_keys (node_id, identity_pubkey, replaced_at) VALUES (?, ?, ?)
		ON CONFLICT DO NOTHING`, id, pub, at); err != nil {
		return fmt.Errorf("retire an identity key of node %d: %w", id, err)
	}
	return nil
}

// RetiredKeys are node id's earlier identity keys, oldest first.
func (t *Tx) RetiredKeys(id int64) ([]string, error) {
	rows, err := t.query(`SELECT identity_pubkey FROM node_keys WHERE node_id = ? ORDER BY replaced_at, identity_pubkey`, id)
	if err != nil {
		return nil, fmt.Errorf("list the retired keys of node %d: %w", id, err)
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var k string
		if err := rows.Scan(&k); err != nil {
			return nil, fmt.Errorf("list the retired keys of node %d: %w", id, err)
		}
		out = append(out, k)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list the retired keys of node %d: %w", id, err)
	}
	return out, nil
}

// KeyRetired reports whether any node ever had identity key pub and lost
// it.
func (t *Tx) KeyRetired(pub string) (bool, error) {
	var n int
	if err := t.queryRow(`SELECT COUNT(*) FROM node_keys WHERE identity_pubkey = ?`, pub).Scan(&n); err != nil {
		return false, fmt.Errorf("look up an identity key: %w", err)
	}
	return n > 0, nil
}
