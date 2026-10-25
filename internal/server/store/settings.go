package store

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

// Setting keys (spec S8.2). Each mesh has its own settings (amendment
// A49); server_url and server_tz are the control plane's (ServerSetting).
const (
	SettingMeshCIDR    = "mesh_cidr"
	SettingUpstreamDNS = "upstream_dns"
	SettingDefaultMTU  = "default_mtu"
	SettingDefaultHub  = "default_hub"
	SettingServerURL   = "server_url"
	// SettingDNSZone is the zone every node serves (amendment A36).
	SettingDNSZone = "dns_zone"
	// SettingDNSExtraDomains are the extra domains every node resolves
	// (amendment A39), comma separated.
	SettingDNSExtraDomains = "dns_extra_domains"
	// SettingMeshCIDR6 is the mesh's IPv6 prefix (M8): MeshCIDR6Off when
	// IPv6 was turned off, "" (or unset) until the server picks one at its
	// next start.
	SettingMeshCIDR6 = "mesh_cidr6"
	// MeshCIDR6Off is SettingMeshCIDR6 for a mesh whose IPv6 was turned
	// off (M8): the server leaves it IPv4 only.
	MeshCIDR6Off = "off"
	// SettingPolicyTZ is the time zone the mesh firewall's schedules are
	// read in (amendment A43): Segmentation.TimeZone, so part of the
	// segmentation document, not of DesiredSettings. Unset is
	// SettingServerTZ.
	SettingPolicyTZ = "policy_tz"
	// SettingServerTZ is the server's own time zone, written at every
	// start: a deployment setting, never in a snapshot.
	SettingServerTZ = "server_tz"
)

// DesiredSettings are the settings that shape computed configs. They are
// part of every snapshot; deployment settings such as server_url are not, so
// a rollback never moves the server.
var DesiredSettings = []string{SettingMeshCIDR, SettingUpstreamDNS, SettingDefaultMTU, SettingDefaultHub, SettingDNSZone,
	SettingDNSExtraDomains, SettingMeshCIDR6}

// settingDefaults are what a setting held before it existed: restoring a
// snapshot taken then puts the default back. "mesh" is proto.DefaultZone.
var settingDefaults = map[string]string{SettingDNSZone: "mesh", SettingDNSExtraDomains: "", SettingMeshCIDR6: ""}

// SplitDomains reads SettingDNSExtraDomains: the comma-separated domains,
// none when it is empty.
func SplitDomains(s string) []string {
	var out []string
	for _, d := range strings.Split(s, ",") {
		if d = strings.TrimSpace(d); d != "" {
			out = append(out, d)
		}
	}
	return out
}

// Settings returns every setting of the mesh.
func (t *MeshTx) Settings() (map[string]string, error) {
	rows, err := t.query(`SELECT key, value FROM mesh_settings WHERE mesh_id = ?`, t.mesh)
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

// Setting returns one setting of the mesh, or ErrNotFound.
func (t *MeshTx) Setting(key string) (string, error) {
	var v string
	err := t.queryRow(`SELECT value FROM mesh_settings WHERE mesh_id = ? AND key = ?`, t.mesh, key).Scan(&v)
	if errors.Is(err, sql.ErrNoRows) {
		return "", fmt.Errorf("setting %q: %w", key, ErrNotFound)
	}
	if err != nil {
		return "", fmt.Errorf("read setting %q: %w", key, err)
	}
	return v, nil
}

// DeleteSetting removes a setting of the mesh, so it reads as never set.
func (t *MeshTx) DeleteSetting(key string) error {
	if _, err := t.exec(`DELETE FROM mesh_settings WHERE mesh_id = ? AND key = ?`, t.mesh, key); err != nil {
		return fmt.Errorf("remove setting %q: %w", key, err)
	}
	return nil
}

// SetSetting creates or replaces a setting of the mesh.
func (t *MeshTx) SetSetting(key, value string) error {
	if _, err := t.exec(`INSERT INTO mesh_settings (mesh_id, key, value) VALUES (?, ?, ?)
		ON CONFLICT (mesh_id, key) DO UPDATE SET value = excluded.value`, t.mesh, key, value); err != nil {
		return fmt.Errorf("write setting %q: %w", key, err)
	}
	return nil
}

// ServerSetting returns one of the control plane's own settings
// (SettingServerURL, SettingServerTZ), "" when it is not set.
func (t *Tx) ServerSetting(key string) (string, error) {
	var v string
	err := t.queryRow(`SELECT value FROM settings WHERE key = ?`, key).Scan(&v)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("read setting %q: %w", key, err)
	}
	return v, nil
}

// SetServerSetting creates or replaces one of the control plane's own
// settings.
func (t *Tx) SetServerSetting(key, value string) error {
	if _, err := t.exec(`INSERT INTO settings (key, value) VALUES (?, ?)
		ON CONFLICT (key) DO UPDATE SET value = excluded.value`, key, value); err != nil {
		return fmt.Errorf("write setting %q: %w", key, err)
	}
	return nil
}
