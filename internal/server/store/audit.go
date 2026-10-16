package store

import (
	"encoding/json"
	"fmt"
)

// AuditEntry is one row of audit_log (spec S8.2: who changed what, when).
type AuditEntry struct {
	ID     int64
	At     int64
	Actor  string
	UserID int64
	Action string
	Target string
	Detail string
}

// Audit records an action. detail is marshalled to JSON; callers must not put
// secrets in it (tokens, passwords, keys' private halves).
func (t *Tx) Audit(actor string, userID int64, action, target string, detail any) error {
	d := ""
	if detail != nil {
		b, err := json.Marshal(detail)
		if err != nil {
			return fmt.Errorf("audit %s: encode detail: %w", action, err)
		}
		d = string(b)
	}
	if _, err := t.exec(`INSERT INTO audit_log (at, actor, user_id, action, target, detail) VALUES (?, ?, ?, ?, ?, ?)`,
		t.now.Unix(), actor, nullInt(userID), action, target, d); err != nil {
		return fmt.Errorf("audit %s: %w", action, err)
	}
	return nil
}

// AuditLog lists entries newest first. before = 0 starts at the newest.
func (t *Tx) AuditLog(limit int, before int64) ([]AuditEntry, error) {
	if before <= 0 {
		before = 1<<63 - 1
	}
	rows, err := t.query(`SELECT id, at, actor, COALESCE(user_id, 0), action, target, detail
		FROM audit_log WHERE id < ? ORDER BY id DESC LIMIT ?`, before, limit)
	if err != nil {
		return nil, fmt.Errorf("read audit log: %w", err)
	}
	defer rows.Close()
	var out []AuditEntry
	for rows.Next() {
		var e AuditEntry
		if err := rows.Scan(&e.ID, &e.At, &e.Actor, &e.UserID, &e.Action, &e.Target, &e.Detail); err != nil {
			return nil, fmt.Errorf("read audit log: %w", err)
		}
		out = append(out, e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read audit log: %w", err)
	}
	return out, nil
}
