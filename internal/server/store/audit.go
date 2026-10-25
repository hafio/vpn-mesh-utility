package store

import (
	"encoding/json"
	"fmt"
)

// AuditEntry is one row of audit_log (spec S8.2: who changed what, when).
type AuditEntry struct {
	ID int64
	// MeshID is the mesh the action was in; 0 for the control plane's own
	// (users, meshes).
	MeshID int64
	At     int64
	Actor  string
	UserID int64
	Action string
	Target string
	Detail string
}

// Audit records an action of the control plane's own (users, meshes).
// detail is marshalled to JSON; callers must not put secrets in it (tokens,
// passwords, keys' private halves).
func (t *Tx) Audit(actor string, userID int64, action, target string, detail any) error {
	return t.audit(0, actor, userID, action, target, detail)
}

// Audit records an action in the mesh; see Tx.Audit.
func (t *MeshTx) Audit(actor string, userID int64, action, target string, detail any) error {
	return t.audit(t.mesh, actor, userID, action, target, detail)
}

func (t *Tx) audit(mesh int64, actor string, userID int64, action, target string, detail any) error {
	d := ""
	if detail != nil {
		b, err := json.Marshal(detail)
		if err != nil {
			return fmt.Errorf("audit %s: encode detail: %w", action, err)
		}
		d = string(b)
	}
	if _, err := t.exec(`INSERT INTO audit_log (at, actor, user_id, action, target, detail, mesh_id) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		t.now.Unix(), actor, nullInt(userID), action, target, d, nullInt(mesh)); err != nil {
		return fmt.Errorf("audit %s: %w", action, err)
	}
	return nil
}

// AuditLog lists every entry, newest first. before = 0 starts at the
// newest.
func (t *Tx) AuditLog(limit int, before int64) ([]AuditEntry, error) {
	return t.auditLog(`id < ?`, before, limit)
}

// AuditLog lists the mesh's entries and the control plane's own, newest
// first; see Tx.AuditLog.
func (t *MeshTx) AuditLog(limit int, before int64) ([]AuditEntry, error) {
	return t.auditLog(`id < ? AND (mesh_id = ? OR mesh_id IS NULL)`, before, t.mesh, limit)
}

// OwnAuditLog lists the mesh's entries alone, newest first: what a user who
// manages only this mesh sees (amendment A49).
func (t *MeshTx) OwnAuditLog(limit int, before int64) ([]AuditEntry, error) {
	return t.auditLog(`id < ? AND mesh_id = ?`, before, t.mesh, limit)
}

// auditLog reads entries matching where, whose first argument is before;
// the last of args is the limit.
func (t *Tx) auditLog(where string, before int64, args ...any) ([]AuditEntry, error) {
	if before <= 0 {
		before = 1<<63 - 1
	}
	rows, err := t.query(`SELECT id, COALESCE(mesh_id, 0), at, actor, COALESCE(user_id, 0), action, target, detail
		FROM audit_log WHERE `+where+` ORDER BY id DESC LIMIT ?`, append([]any{before}, args...)...)
	if err != nil {
		return nil, fmt.Errorf("read audit log: %w", err)
	}
	defer rows.Close()
	var out []AuditEntry
	for rows.Next() {
		var e AuditEntry
		if err := rows.Scan(&e.ID, &e.MeshID, &e.At, &e.Actor, &e.UserID, &e.Action, &e.Target, &e.Detail); err != nil {
			return nil, fmt.Errorf("read audit log: %w", err)
		}
		out = append(out, e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read audit log: %w", err)
	}
	return out, nil
}
