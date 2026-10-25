package store

import (
	"database/sql"
	"errors"
	"fmt"
)

// Version is one row of config_versions. Each mesh numbers its own from 1
// (amendment A49).
type Version struct {
	Version        int64
	CreatedAt      int64
	Actor          string
	UserID         int64
	Summary        string
	Snapshot       []byte
	RolledBackFrom int64
}

// InsertVersion appends a version of the mesh and returns its number.
func (t *MeshTx) InsertVersion(v Version) (int64, error) {
	n, err := t.LatestVersion()
	if err != nil {
		return 0, err
	}
	n++
	if _, err := t.exec(`INSERT INTO config_versions (mesh_id, version, created_at, actor, user_id, summary, snapshot_json, rolled_back_from)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		t.mesh, n, t.now.Unix(), v.Actor, nullInt(v.UserID), v.Summary, string(v.Snapshot), nullInt(v.RolledBackFrom)); err != nil {
		return 0, fmt.Errorf("insert config version: %w", err)
	}
	return n, nil
}

// LatestVersion returns the mesh's newest version number, 0 when there is
// none.
func (t *MeshTx) LatestVersion() (int64, error) {
	var v int64
	if err := t.queryRow(`SELECT COALESCE(MAX(version), 0) FROM config_versions WHERE mesh_id = ?`, t.mesh).Scan(&v); err != nil {
		return 0, fmt.Errorf("read latest config version: %w", err)
	}
	return v, nil
}

// VersionByNumber returns one of the mesh's versions with its snapshot, or
// ErrNotFound.
func (t *MeshTx) VersionByNumber(n int64) (Version, error) {
	var (
		v        Version
		user, rb sql.NullInt64
		snap     string
	)
	err := t.queryRow(`SELECT version, created_at, actor, user_id, summary, snapshot_json, rolled_back_from
		FROM config_versions WHERE mesh_id = ? AND version = ?`, t.mesh, n).Scan(&v.Version, &v.CreatedAt, &v.Actor, &user, &v.Summary, &snap, &rb)
	if errors.Is(err, sql.ErrNoRows) {
		return Version{}, fmt.Errorf("config v%d: %w", n, ErrNotFound)
	}
	if err != nil {
		return Version{}, fmt.Errorf("read config v%d: %w", n, err)
	}
	v.UserID, v.RolledBackFrom, v.Snapshot = user.Int64, rb.Int64, []byte(snap)
	return v, nil
}

// Versions lists the mesh's versions newest first, without snapshots.
// before = 0 starts at the newest.
func (t *MeshTx) Versions(limit int, before int64) ([]Version, error) {
	if before <= 0 {
		before = 1<<63 - 1
	}
	rows, err := t.query(`SELECT version, created_at, actor, user_id, summary, rolled_back_from
		FROM config_versions WHERE mesh_id = ? AND version < ? ORDER BY version DESC LIMIT ?`, t.mesh, before, limit)
	if err != nil {
		return nil, fmt.Errorf("list config versions: %w", err)
	}
	defer rows.Close()
	var out []Version
	for rows.Next() {
		var (
			v        Version
			user, rb sql.NullInt64
		)
		if err := rows.Scan(&v.Version, &v.CreatedAt, &v.Actor, &user, &v.Summary, &rb); err != nil {
			return nil, fmt.Errorf("list config versions: %w", err)
		}
		v.UserID, v.RolledBackFrom = user.Int64, rb.Int64
		out = append(out, v)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list config versions: %w", err)
	}
	return out, nil
}
