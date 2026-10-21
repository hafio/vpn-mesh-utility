-- Segmentation (amendment A43): groups of nodes, address and service
-- objects, zones, and the ordered policies every server enforces. It
-- replaces the custom firewall rules: each becomes an allow policy, and the
-- mesh starts with the policy allow-all last. Desired state: part of every
-- snapshot and export. References between them are kind:name strings, checked by the
-- planner on every change, so a dangling one refuses the change that would
-- leave it. List columns are comma separated; their grammar (internal/
-- validate) has no commas.
CREATE TABLE seg_groups (
  id          INTEGER PRIMARY KEY,
  name        TEXT    NOT NULL UNIQUE,
  description TEXT    NOT NULL DEFAULT ''
) STRICT;

CREATE TABLE node_groups (
  node_id  INTEGER NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
  group_id INTEGER NOT NULL REFERENCES seg_groups(id) ON DELETE CASCADE,
  PRIMARY KEY (node_id, group_id)
) STRICT;

CREATE TABLE seg_objects (
  id          INTEGER PRIMARY KEY,
  kind        TEXT    NOT NULL CHECK (kind IN ('address', 'service')),
  name        TEXT    NOT NULL,
  description TEXT    NOT NULL DEFAULT '',
  entries     TEXT    NOT NULL,
  UNIQUE (kind, name)
) STRICT;

CREATE TABLE seg_zones (
  id          INTEGER PRIMARY KEY,
  name        TEXT    NOT NULL UNIQUE,
  description TEXT    NOT NULL DEFAULT '',
  intra       TEXT    NOT NULL DEFAULT 'allow' CHECK (intra IN ('allow', 'deny')),
  members     TEXT    NOT NULL DEFAULT ''
) STRICT;

CREATE TABLE seg_policies (
  id          INTEGER PRIMARY KEY,
  position    INTEGER NOT NULL UNIQUE,
  name        TEXT    NOT NULL UNIQUE,
  description TEXT    NOT NULL DEFAULT '',
  disabled    INTEGER NOT NULL DEFAULT 0 CHECK (disabled IN (0, 1)),
  from_refs   TEXT    NOT NULL DEFAULT '',
  to_refs     TEXT    NOT NULL DEFAULT '',
  services    TEXT    NOT NULL DEFAULT '',
  action      TEXT    NOT NULL CHECK (action IN ('allow', 'deny'))
) STRICT;

-- Each custom rule, by id, becomes an allow from its source to its server
-- (or every server) for its port; store.LegacySegmentation does the same
-- for snapshots and exports made before this migration.
INSERT INTO seg_policies (position, name, description, from_refs, to_refs, services, action)
SELECT ROW_NUMBER() OVER (ORDER BY f.id), 'firewall-' || f.id, f.note, f.source_cidr,
       CASE WHEN f.node_id IS NULL THEN 'role:server' ELSE 'node:' || n.name END,
       f.proto || '/' || f.port, 'allow'
FROM firewall_rules f LEFT JOIN nodes n ON n.id = f.node_id;

-- store.DefaultPolicy.
INSERT INTO seg_policies (position, name, description, action)
VALUES ((SELECT COUNT(*) FROM seg_policies) + 1, 'allow-all',
        'Every mesh address reaches every other. Change or remove it to segment the mesh.', 'allow');

DROP TABLE firewall_rules;
