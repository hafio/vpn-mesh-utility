package store

import "fmt"

// StatusUpdate is one agent heartbeat as the server keeps it. It is
// operational state: never part of a snapshot, never a config version.
type StatusUpdate struct {
	AppliedVersion int64
	AppliedHash    string
	AgentVersion   string
	WGBackend      string
	ObservedIP     string
	// StatusJSON holds the display-only parts (errors, degraded steps, the
	// last rollback) as the API received them, already validated.
	StatusJSON string
	Peers      []PeerStat
}

// PeerStat is the runtime state of one peer as a node reported it.
type PeerStat struct {
	PeerID        int64
	LastHandshake int64
	RxBytes       int64
	TxBytes       int64
	Endpoint      string
	UpdatedAt     int64
}

// RecordStatus stores a heartbeat: the node's applied state and last-seen
// time, and replaces its peer rows.
func (t *Tx) RecordStatus(nodeID int64, u StatusUpdate) error {
	res, err := t.exec(`UPDATE nodes SET applied_version = ?, applied_hash = ?, agent_version = ?,
			wg_backend = ?, observed_ip = ?, status_json = ?, last_seen = ?
		WHERE id = ?`,
		u.AppliedVersion, u.AppliedHash, u.AgentVersion, u.WGBackend, u.ObservedIP, u.StatusJSON, t.now.Unix(), nodeID)
	if err != nil {
		return fmt.Errorf("record status of node %d: %w", nodeID, err)
	}
	if err := mustAffect(res, fmt.Sprintf("node %d", nodeID)); err != nil {
		return err
	}
	if _, err := t.exec(`DELETE FROM peer_status WHERE node_id = ?`, nodeID); err != nil {
		return fmt.Errorf("record status of node %d: %w", nodeID, err)
	}
	for _, p := range u.Peers {
		if _, err := t.exec(`INSERT INTO peer_status (node_id, peer_id, last_handshake, rx_bytes, tx_bytes, endpoint, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?)`,
			nodeID, p.PeerID, p.LastHandshake, p.RxBytes, p.TxBytes, p.Endpoint, t.now.Unix()); err != nil {
			return fmt.Errorf("record peer %d of node %d: %w", p.PeerID, nodeID, err)
		}
	}
	return nil
}

// SetApplied records the config a node runs when no agent reports it: a
// phone holds the profile it was last given.
func (t *Tx) SetApplied(nodeID int64, hash string, version int64) error {
	res, err := t.exec(`UPDATE nodes SET applied_hash = ?, applied_version = ? WHERE id = ?`, hash, version, nodeID)
	if err != nil {
		return fmt.Errorf("record the applied config of node %d: %w", nodeID, err)
	}
	return mustAffect(res, fmt.Sprintf("node %d", nodeID))
}

// LastHandshakeWith is the latest handshake any node reported with peerID
// (unix seconds, 0 for none): what the servers know of a phone.
func (t *Tx) LastHandshakeWith(peerID int64) (int64, error) {
	var at int64
	err := t.queryRow(`SELECT COALESCE(MAX(last_handshake), 0) FROM peer_status WHERE peer_id = ?`, peerID).Scan(&at)
	if err != nil {
		return 0, fmt.Errorf("read the handshakes with node %d: %w", peerID, err)
	}
	return at, nil
}

// PeerStats returns what a node last reported about its peers, by peer id.
func (t *Tx) PeerStats(nodeID int64) ([]PeerStat, error) {
	rows, err := t.query(`SELECT peer_id, last_handshake, rx_bytes, tx_bytes, endpoint, updated_at
		FROM peer_status WHERE node_id = ? ORDER BY peer_id`, nodeID)
	if err != nil {
		return nil, fmt.Errorf("read peer status of node %d: %w", nodeID, err)
	}
	defer rows.Close()
	var out []PeerStat
	for rows.Next() {
		var p PeerStat
		if err := rows.Scan(&p.PeerID, &p.LastHandshake, &p.RxBytes, &p.TxBytes, &p.Endpoint, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("read peer status of node %d: %w", nodeID, err)
		}
		out = append(out, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read peer status of node %d: %w", nodeID, err)
	}
	return out, nil
}
