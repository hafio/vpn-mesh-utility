package store

import (
	"database/sql"
	"errors"
	"fmt"
)

// Setting keys (spec S8.2).
const (
	SettingMeshCIDR    = "mesh_cidr"
	SettingUpstreamDNS = "upstream_dns"
	SettingDefaultMTU  = "default_mtu"
	SettingDefaultHub  = "default_hub"
	SettingServerURL   = "server_url"
)

// DesiredSettings are the settings that shape computed configs. They are
// part of every snapshot; deployment settings such as server_url are not, so
// a rollback never moves the server.
var DesiredSettings = []string{SettingMeshCIDR, SettingUpstreamDNS, SettingDefaultMTU, SettingDefaultHub}

// Settings returns every setting.
func (t *Tx) Settings() (map[string]string, error) {
	rows, err := t.query(`SELECT key, value FROM settings`)
	if err != nil {
		return nil, fmt.Errorf("read settings: %w", err)
	}
	defer rows.Close()
	out := map[string]string{}
	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			return nil, fmt.Errorf("read settings: %w", err)
		}
		out[k] = v
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read settings: %w", err)
	}
	return out, nil
}

// Setting returns one setting, or ErrNotFound.
func (t *Tx) Setting(key string) (string, error) {
	var v string
	err := t.queryRow(`SELECT value FROM settings WHERE key = ?`, key).Scan(&v)
	if errors.Is(err, sql.ErrNoRows) {
		return "", fmt.Errorf("setting %q: %w", key, ErrNotFound)
	}
	if err != nil {
		return "", fmt.Errorf("read setting %q: %w", key, err)
	}
	return v, nil
}

// SetSetting creates or replaces a setting.
func (t *Tx) SetSetting(key, value string) error {
	if _, err := t.exec(`INSERT INTO settings (key, value) VALUES (?, ?)
		ON CONFLICT (key) DO UPDATE SET value = excluded.value`, key, value); err != nil {
		return fmt.Errorf("write setting %q: %w", key, err)
	}
	return nil
}
