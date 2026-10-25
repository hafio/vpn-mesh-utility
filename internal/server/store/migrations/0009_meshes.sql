-- Several meshes on one control plane (amendment A49). Every table of
-- desired state belongs to one mesh, and every name in it is unique only
-- within its mesh; users, sessions and the deployment's own settings
-- (server_url, server_tz) stay the control plane's. The single mesh of an
-- earlier database becomes mesh 1.

-- A mesh's uuid names it to its agents; its slot (0..63) picks the names
-- an agent of this mesh uses on a host (wg<slot>, its routing tables, its
-- firewall names), so agents of two meshes share a host. Mesh 1 keeps
-- slot 0: the names every release used so far.
ALTER TABLE meshes ADD COLUMN uuid TEXT NOT NULL DEFAULT '';
ALTER TABLE meshes ADD COLUMN slot INTEGER NOT NULL DEFAULT 0 CHECK (slot BETWEEN 0 AND 63);
ALTER TABLE meshes ADD COLUMN created_at INTEGER NOT NULL DEFAULT 0;
UPDATE meshes SET
  uuid = lower(hex(randomblob(4)) || '-' || hex(randomblob(2)) || '-4' || substr(hex(randomblob(2)), 2) || '-'
         || substr('89ab', 1 + (random() & 3), 1) || substr(hex(randomblob(2)), 2) || '-' || hex(randomblob(6))),
  created_at = CAST(strftime('%s', 'now') AS INTEGER);
CREATE UNIQUE INDEX meshes_uuid ON meshes (uuid);
CREATE UNIQUE INDEX meshes_slot ON meshes (slot);

-- Node ids are never given out twice, not even after the newest node is
-- removed: a rollback re-creates a removed node with its old id, which
-- another mesh must not have been given in the meantime.
CREATE TABLE id_seq (
  name TEXT    PRIMARY KEY,
  last INTEGER NOT NULL
) STRICT;
INSERT INTO id_seq (name, last) VALUES ('nodes', (SELECT COALESCE(MAX(id), 0) FROM nodes));

CREATE TABLE mesh_settings (
  mesh_id INTEGER NOT NULL REFERENCES meshes(id) ON DELETE CASCADE,
  key     TEXT    NOT NULL,
  value   TEXT    NOT NULL,
  PRIMARY KEY (mesh_id, key)
) STRICT;
INSERT INTO mesh_settings (mesh_id, key, value)
  SELECT 1, key, value FROM settings WHERE key NOT IN ('server_url', 'server_tz');
DELETE FROM settings WHERE key NOT IN ('server_url', 'server_tz');

-- AUTOINCREMENT for the same reason as id_seq: a rollback re-creates a
-- record with its old id.
CREATE TABLE dns_records_new (
  id         INTEGER PRIMARY KEY AUTOINCREMENT,
  mesh_id    INTEGER NOT NULL REFERENCES meshes(id) ON DELETE CASCADE,
  name       TEXT    NOT NULL,
  ip         TEXT,
  ip6        TEXT,
  node_id    INTEGER REFERENCES nodes(id) ON DELETE CASCADE,
  created_at INTEGER NOT NULL,
  UNIQUE (mesh_id, name),
  CHECK ((ip IS NULL) <> (node_id IS NULL))
) STRICT;
INSERT INTO dns_records_new (id, mesh_id, name, ip, ip6, node_id, created_at)
  SELECT id, 1, name, ip, ip6, node_id, created_at FROM dns_records;
DROP TABLE dns_records;
ALTER TABLE dns_records_new RENAME TO dns_records;

-- Each mesh numbers its own versions from 1.
CREATE TABLE config_versions_new (
  mesh_id          INTEGER NOT NULL REFERENCES meshes(id) ON DELETE CASCADE,
  version          INTEGER NOT NULL,
  created_at       INTEGER NOT NULL,
  actor            TEXT    NOT NULL,
  user_id          INTEGER,
  summary          TEXT    NOT NULL,
  snapshot_json    TEXT    NOT NULL,
  rolled_back_from INTEGER,
  PRIMARY KEY (mesh_id, version)
) STRICT;
INSERT INTO config_versions_new (mesh_id, version, created_at, actor, user_id, summary, snapshot_json, rolled_back_from)
  SELECT 1, version, created_at, actor, user_id, summary, snapshot_json, rolled_back_from FROM config_versions;
DROP TABLE config_versions;
ALTER TABLE config_versions_new RENAME TO config_versions;

-- node_groups is rebuilt with seg_groups, so dropping the old groups
-- cascades into nothing.
CREATE TABLE seg_groups_new (
  id          INTEGER PRIMARY KEY,
  mesh_id     INTEGER NOT NULL REFERENCES meshes(id) ON DELETE CASCADE,
  name        TEXT    NOT NULL,
  description TEXT    NOT NULL DEFAULT '',
  UNIQUE (mesh_id, name)
) STRICT;
INSERT INTO seg_groups_new (id, mesh_id, name, description) SELECT id, 1, name, description FROM seg_groups;
CREATE TABLE node_groups_new (
  node_id  INTEGER NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
  group_id INTEGER NOT NULL REFERENCES seg_groups_new(id) ON DELETE CASCADE,
  PRIMARY KEY (node_id, group_id)
) STRICT;
INSERT INTO node_groups_new (node_id, group_id) SELECT node_id, group_id FROM node_groups;
DROP TABLE node_groups;
DROP TABLE seg_groups;
ALTER TABLE seg_groups_new RENAME TO seg_groups;
ALTER TABLE node_groups_new RENAME TO node_groups;

CREATE TABLE seg_objects_new (
  id          INTEGER PRIMARY KEY,
  mesh_id     INTEGER NOT NULL REFERENCES meshes(id) ON DELETE CASCADE,
  kind        TEXT    NOT NULL CHECK (kind IN ('address', 'service')),
  name        TEXT    NOT NULL,
  description TEXT    NOT NULL DEFAULT '',
  entries     TEXT    NOT NULL,
  UNIQUE (mesh_id, kind, name)
) STRICT;
INSERT INTO seg_objects_new (id, mesh_id, kind, name, description, entries)
  SELECT id, 1, kind, name, description, entries FROM seg_objects;
DROP TABLE seg_objects;
ALTER TABLE seg_objects_new RENAME TO seg_objects;

CREATE TABLE seg_zones_new (
  id          INTEGER PRIMARY KEY,
  mesh_id     INTEGER NOT NULL REFERENCES meshes(id) ON DELETE CASCADE,
  name        TEXT    NOT NULL,
  description TEXT    NOT NULL DEFAULT '',
  intra       TEXT    NOT NULL DEFAULT 'allow' CHECK (intra IN ('allow', 'deny')),
  members     TEXT    NOT NULL DEFAULT '',
  UNIQUE (mesh_id, name)
) STRICT;
INSERT INTO seg_zones_new (id, mesh_id, name, description, intra, members)
  SELECT id, 1, name, description, intra, members FROM seg_zones;
DROP TABLE seg_zones;
ALTER TABLE seg_zones_new RENAME TO seg_zones;

CREATE TABLE seg_policies_new (
  id          INTEGER PRIMARY KEY,
  mesh_id     INTEGER NOT NULL REFERENCES meshes(id) ON DELETE CASCADE,
  position    INTEGER NOT NULL,
  name        TEXT    NOT NULL,
  description TEXT    NOT NULL DEFAULT '',
  disabled    INTEGER NOT NULL DEFAULT 0 CHECK (disabled IN (0, 1)),
  from_refs   TEXT    NOT NULL DEFAULT '',
  to_refs     TEXT    NOT NULL DEFAULT '',
  services    TEXT    NOT NULL DEFAULT '',
  action      TEXT    NOT NULL CHECK (action IN ('allow', 'deny')),
  days        TEXT    NOT NULL DEFAULT '',
  start_time  TEXT    NOT NULL DEFAULT '',
  end_time    TEXT    NOT NULL DEFAULT '',
  UNIQUE (mesh_id, position),
  UNIQUE (mesh_id, name)
) STRICT;
INSERT INTO seg_policies_new (id, mesh_id, position, name, description, disabled, from_refs, to_refs, services, action,
    days, start_time, end_time)
  SELECT id, 1, position, name, description, disabled, from_refs, to_refs, services, action, days, start_time, end_time
  FROM seg_policies;
DROP TABLE seg_policies;
ALTER TABLE seg_policies_new RENAME TO seg_policies;

-- A mesh's audit entries go with it; the control plane's own (users, and
-- the meshes themselves) have none.
ALTER TABLE audit_log ADD COLUMN mesh_id INTEGER REFERENCES meshes(id) ON DELETE CASCADE;
UPDATE audit_log SET mesh_id = 1 WHERE action NOT LIKE 'user.%';
CREATE INDEX audit_log_mesh ON audit_log (mesh_id, id);
