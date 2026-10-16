package store

import (
	"database/sql"
	"errors"
	"fmt"
)

// Version is one row of config_versions.
type Version struct {
	Version        int64
	CreatedAt      int64
	Actor          string
	UserID         int64
	Summary        string
	Snapshot       []byte
	RolledBackFrom int64
}

// InsertVersion appends a version and returns its number.
func (t *Tx) InsertVersion(v Version) (int64, error) {
	res, err := t.exec(`INSERT INTO config_versions (created_at, actor, user_id, summary, snapshot_json, rolled_back_from)
		VALUES (?, ?, ?, ?, ?, ?)`,
		t.now.Unix(), v.Actor, nullInt(v.UserID), v.Summary, string(v.Snapshot), nullInt(v.RolledBackFrom))
	if err != nil {
		return 0, fmt.Errorf("insert config version: %w", err)
	}
	n, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("insert config version: %w", err)
	}
	return n, nil
}

// LatestVersion returns the newest version number, 0 when there is none.
func (t *Tx) LatestVersion() (int64, error) {
	var v int64
	if err := t.queryRow(`SELECT COALESCE(MAX(version), 0) FROM config_versions`).Scan(&v); err != nil {
		return 0, fmt.Errorf("read latest config version: %w", err)
	}
	return v, nil
}

// VersionByNumber returns one version with its snapshot, or ErrNotFound.
func (t *Tx) VersionByNumber(n int64) (Version, error) {
	var (
		v        Version
		user, rb sql.NullInt64
		snap     string
	)
	err := t.queryRow(`SELECT version, created_at, actor, user_id, summary, snapshot_json, rolled_back_from
		FROM config_versions WHERE version = ?`, n).Scan(&v.Version, &v.CreatedAt, &v.Actor, &user, &v.Summary, &snap, &rb)
	if errors.Is(err, sql.ErrNoRows) {
		return Version{}, fmt.Errorf("config version %d: %w", n, ErrNotFound)
	}
	if err != nil {
		return Version{}, fmt.Errorf("read config version %d: %w", n, err)
	}
	v.UserID, v.RolledBackFrom, v.Snapshot = user.Int64, rb.Int64, []byte(snap)
	return v, nil
}

// Versions lists versions newest first, without snapshots. before = 0 starts
// at the newest.
func (t *Tx) Versions(limit int, before int64) ([]Version, error) {
	if before <= 0 {
		before = 1<<63 - 1
	}
	rows, err := t.query(`SELECT version, created_at, actor, user_id, summary, rolled_back_from
		FROM config_versions WHERE version < ? ORDER BY version DESC LIMIT ?`, before, limit)
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
