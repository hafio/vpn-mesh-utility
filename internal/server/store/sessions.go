package store

import (
	"database/sql"
	"errors"
	"fmt"
)

// AgentSession is a node's live session (amendment A50). The hashes are
// sha256 hex of the secrets the agent holds; the store never sees those.
type AgentSession struct {
	NodeID      int64
	SessionHash string
	TicketHash  string
	// PrevHash is the ticket the holder presented last, which a retry
	// after a lost reply presents again.
	PrevHash string
	LastSeen int64
	// Released is a session its holder ended with goodbye.
	Released bool
	Hostname string
	// ChallengeAt is when another copy of the node first said hello while
	// this one held the session, 0 for none; WaitingHost is that copy's
	// hostname, and Answered says the holder proved it is alive since.
	ChallengeAt int64
	Answered    bool
	WaitingHost string
}

const sessionCols = `node_id, session_hash, ticket_hash, prev_hash, last_seen, released, hostname, challenge_at, answered, waiting_host`

// AgentSession returns node's session, or ErrNotFound.
func (t *Tx) AgentSession(node int64) (AgentSession, error) {
	var s AgentSession
	err := t.queryRow(`SELECT `+sessionCols+` FROM agent_sessions WHERE node_id = ?`, node).Scan(&s.NodeID, &s.SessionHash, &s.TicketHash,
		&s.PrevHash, &s.LastSeen, &s.Released, &s.Hostname, &s.ChallengeAt, &s.Answered, &s.WaitingHost)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return AgentSession{}, fmt.Errorf("session of node %d: %w", node, ErrNotFound)
	case err != nil:
		return AgentSession{}, fmt.Errorf("read the session of node %d: %w", node, err)
	}
	return s, nil
}

// PutAgentSession writes node's session, replacing the one it had.
func (t *Tx) PutAgentSession(s AgentSession) error {
	if _, err := t.exec(`INSERT INTO agent_sessions (`+sessionCols+`) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (node_id) DO UPDATE SET session_hash = excluded.session_hash, ticket_hash = excluded.ticket_hash,
		prev_hash = excluded.prev_hash, last_seen = excluded.last_seen, released = excluded.released, hostname = excluded.hostname,
		challenge_at = excluded.challenge_at, answered = excluded.answered, waiting_host = excluded.waiting_host`,
		s.NodeID, s.SessionHash, s.TicketHash, s.PrevHash, s.LastSeen, s.Released, s.Hostname, s.ChallengeAt, s.Answered,
		s.WaitingHost); err != nil {
		return fmt.Errorf("write the session of node %d: %w", s.NodeID, err)
	}
	return nil
}

// DeleteAgentSession ends node's session, if it has one: a new token or a
// join cuts it off.
func (t *Tx) DeleteAgentSession(node int64) error {
	if _, err := t.exec(`DELETE FROM agent_sessions WHERE node_id = ?`, node); err != nil {
		return fmt.Errorf("end the session of node %d: %w", node, err)
	}
	return nil
}

// Machine is the machine a node is pinned to: sha256 of the id it
// presents, and its hostname.
type Machine struct {
	Hash     string
	Hostname string
	PinnedAt int64
}

// Machine returns the machine node is pinned to, or ErrNotFound.
func (t *Tx) Machine(node int64) (Machine, error) {
	var m Machine
	err := t.queryRow(`SELECT machine_hash, hostname, pinned_at FROM node_machines WHERE node_id = ?`, node).Scan(&m.Hash, &m.Hostname, &m.PinnedAt)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return Machine{}, fmt.Errorf("machine of node %d: %w", node, ErrNotFound)
	case err != nil:
		return Machine{}, fmt.Errorf("read the machine of node %d: %w", node, err)
	}
	return m, nil
}

// PinMachine pins node to m.
func (t *Tx) PinMachine(node int64, m Machine) error {
	if _, err := t.exec(`INSERT INTO node_machines (node_id, machine_hash, hostname, pinned_at) VALUES (?, ?, ?, ?)
		ON CONFLICT (node_id) DO UPDATE SET machine_hash = excluded.machine_hash, hostname = excluded.hostname, pinned_at = excluded.pinned_at`,
		node, m.Hash, m.Hostname, m.PinnedAt); err != nil {
		return fmt.Errorf("pin node %d to its machine: %w", node, err)
	}
	return nil
}

// Refusal is the last copy of a node the control plane refused: why
// (proto.CodeDuplicate or proto.CodeForeign), where it ran, and when.
type Refusal struct {
	Code        string
	Hostname    string
	SameMachine bool
	At          int64
}

// MachineMates are the names of the other nodes of the mesh pinned to the
// machine node is pinned to (amendment A53): one machine running two nodes
// of one mesh.
func (t *MeshTx) MachineMates(node int64) ([]string, error) {
	rows, err := t.query(`SELECT n.name FROM node_machines a
		JOIN node_machines b ON b.machine_hash = a.machine_hash AND b.node_id != a.node_id
		JOIN nodes n ON n.id = b.node_id
		WHERE a.node_id = ? AND n.mesh_id = ? ORDER BY n.name`, node, t.mesh)
	if err != nil {
		return nil, fmt.Errorf("list the nodes on node %d's machine: %w", node, err)
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("list the nodes on node %d's machine: %w", node, err)
		}
		out = append(out, name)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list the nodes on node %d's machine: %w", node, err)
	}
	return out, nil
}

// Refusal returns node's last refused copy, or ErrNotFound.
func (t *Tx) Refusal(node int64) (Refusal, error) {
	var r Refusal
	err := t.queryRow(`SELECT code, hostname, same_machine, at FROM agent_refusals WHERE node_id = ?`, node).Scan(&r.Code, &r.Hostname,
		&r.SameMachine, &r.At)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return Refusal{}, fmt.Errorf("refused copy of node %d: %w", node, ErrNotFound)
	case err != nil:
		return Refusal{}, fmt.Errorf("read the refused copy of node %d: %w", node, err)
	}
	return r, nil
}

// RecordRefusal keeps r as node's last refused copy.
func (t *Tx) RecordRefusal(node int64, r Refusal) error {
	if _, err := t.exec(`INSERT INTO agent_refusals (node_id, code, hostname, same_machine, at) VALUES (?, ?, ?, ?, ?)
		ON CONFLICT (node_id) DO UPDATE SET code = excluded.code, hostname = excluded.hostname, same_machine = excluded.same_machine,
		at = excluded.at`, node, r.Code, r.Hostname, r.SameMachine, r.At); err != nil {
		return fmt.Errorf("record the refused copy of node %d: %w", node, err)
	}
	return nil
}
