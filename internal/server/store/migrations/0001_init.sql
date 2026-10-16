-- 0001_init: the v1 schema (spec S8.2 plus the amendments in docs/spec.md).
-- Applied once, inside one transaction, by store.migrate. Never edit a
-- shipped migration; add NNNN_<name>.sql instead.

-- Multiple meshes per server are P2: the column exists now, with one row.
CREATE TABLE meshes (
  id   INTEGER PRIMARY KEY,
  name TEXT    NOT NULL UNIQUE
) STRICT;
INSERT INTO meshes (id, name) VALUES (1, 'default');

CREATE TABLE nodes (
  id              INTEGER PRIMARY KEY,
  mesh_id         INTEGER NOT NULL DEFAULT 1 REFERENCES meshes(id),
  name            TEXT    NOT NULL,
  role            TEXT    NOT NULL CHECK (role IN ('server', 'spoke', 'phone')),
  -- Reported by the agent at join.
  os              TEXT    NOT NULL DEFAULT '',
  arch            TEXT    NOT NULL DEFAULT '',
  hostname        TEXT    NOT NULL DEFAULT '',
  wg_backend      TEXT    NOT NULL DEFAULT '',
  agent_version   TEXT    NOT NULL DEFAULT '',
  mesh_ip         TEXT    NOT NULL,
  mesh_ip6        TEXT,                                -- reserved (P2: IPv6)
  wg_pubkey       TEXT,
  identity_pubkey TEXT,
  endpoint        TEXT    NOT NULL DEFAULT '',         -- servers: public IPv4:port
  listen_port     INTEGER NOT NULL DEFAULT 51820,
  primary_hub_id  INTEGER REFERENCES nodes(id) ON DELETE SET NULL,
  can_exit        INTEGER NOT NULL DEFAULT 0 CHECK (can_exit IN (0, 1)),
  exit_via        INTEGER REFERENCES nodes(id) ON DELETE SET NULL,
  keepalive       INTEGER NOT NULL DEFAULT 25,
  mtu             INTEGER NOT NULL DEFAULT 1420,
  enrolled        INTEGER NOT NULL DEFAULT 0 CHECK (enrolled IN (0, 1)),
  disabled        INTEGER NOT NULL DEFAULT 0 CHECK (disabled IN (0, 1)),
  -- Desired state bookkeeping, written by the reconciler only.
  config_hash     TEXT    NOT NULL DEFAULT '',
  config_version  INTEGER NOT NULL DEFAULT 0,
  -- Operational state from the last heartbeat; never part of a snapshot.
  applied_version INTEGER NOT NULL DEFAULT 0,
  applied_hash    TEXT    NOT NULL DEFAULT '',
  last_seen       INTEGER,
  observed_ip     TEXT    NOT NULL DEFAULT '',
  status_json     TEXT    NOT NULL DEFAULT '{}',
  created_at      INTEGER NOT NULL,
  UNIQUE (mesh_id, name),
  UNIQUE (mesh_id, mesh_ip)
) STRICT;
CREATE UNIQUE INDEX nodes_wg_pubkey ON nodes (wg_pubkey) WHERE wg_pubkey IS NOT NULL;
CREATE UNIQUE INDEX nodes_identity_pubkey ON nodes (identity_pubkey) WHERE identity_pubkey IS NOT NULL;

-- Join tokens are stored as sha256(token) hex; the token itself is shown once.
CREATE TABLE join_tokens (
  token_hash TEXT    PRIMARY KEY,
  node_id    INTEGER NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
  created_at INTEGER NOT NULL,
  expires_at INTEGER NOT NULL,
  used_at    INTEGER
) STRICT;
CREATE INDEX join_tokens_node ON join_tokens (node_id);

-- Custom records only; one automatic record per node is derived at plan time.
-- A record targets either a node (an alias that follows its mesh IP) or a
-- fixed IP, never both.
CREATE TABLE dns_records (
  id         INTEGER PRIMARY KEY,
  name       TEXT    NOT NULL UNIQUE,
  ip         TEXT,
  node_id    INTEGER REFERENCES nodes(id) ON DELETE CASCADE,
  created_at INTEGER NOT NULL,
  CHECK ((ip IS NULL) <> (node_id IS NULL))
) STRICT;

-- Extra mesh-zone openings on server nodes. node_id NULL means every server.
CREATE TABLE firewall_rules (
  id          INTEGER PRIMARY KEY,
  node_id     INTEGER REFERENCES nodes(id) ON DELETE CASCADE,
  proto       TEXT    NOT NULL CHECK (proto IN ('tcp', 'udp')),
  port        INTEGER NOT NULL CHECK (port BETWEEN 1 AND 65535),
  source_cidr TEXT    NOT NULL,
  note        TEXT    NOT NULL DEFAULT ''
) STRICT;

CREATE TABLE peer_status (
  node_id        INTEGER NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
  peer_id        INTEGER NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
  last_handshake INTEGER NOT NULL DEFAULT 0,
  rx_bytes       INTEGER NOT NULL DEFAULT 0,
  tx_bytes       INTEGER NOT NULL DEFAULT 0,
  endpoint       TEXT    NOT NULL DEFAULT '',
  updated_at     INTEGER NOT NULL,
  PRIMARY KEY (node_id, peer_id)
) STRICT;

-- One row per change to any node's computed config. snapshot_json holds the
-- desired-state tables only (store.Snapshot), never users or status.
CREATE TABLE config_versions (
  version          INTEGER PRIMARY KEY AUTOINCREMENT,
  created_at       INTEGER NOT NULL,
  actor            TEXT    NOT NULL,
  user_id          INTEGER,
  summary          TEXT    NOT NULL,
  snapshot_json    TEXT    NOT NULL,
  rolled_back_from INTEGER
) STRICT;

CREATE TABLE users (
  id             INTEGER PRIMARY KEY,
  username       TEXT    NOT NULL UNIQUE,
  password_hash  TEXT    NOT NULL,                     -- argon2id PHC string
  totp_secret    BLOB,                                 -- AES-256-GCM sealed
  totp_enrolled  INTEGER NOT NULL DEFAULT 0 CHECK (totp_enrolled IN (0, 1)),
  totp_last_step INTEGER NOT NULL DEFAULT 0,
  disabled       INTEGER NOT NULL DEFAULT 0 CHECK (disabled IN (0, 1)),
  created_at     INTEGER NOT NULL
) STRICT;

-- id_hash is sha256(cookie value): a database read cannot hijack a session.
CREATE TABLE sessions (
  id_hash    TEXT    PRIMARY KEY,
  user_id    INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  created_at INTEGER NOT NULL,
  last_seen  INTEGER NOT NULL,
  expires_at INTEGER NOT NULL,
  csrf       TEXT    NOT NULL,
  totp_ok    INTEGER NOT NULL DEFAULT 0 CHECK (totp_ok IN (0, 1))
) STRICT;
CREATE INDEX sessions_user ON sessions (user_id);

CREATE TABLE audit_log (
  id      INTEGER PRIMARY KEY,
  at      INTEGER NOT NULL,
  actor   TEXT    NOT NULL,                            -- user:<name> | cli | node:<id> | system
  user_id INTEGER,
  action  TEXT    NOT NULL,
  target  TEXT    NOT NULL DEFAULT '',
  detail  TEXT    NOT NULL DEFAULT ''
) STRICT;
CREATE INDEX audit_log_at ON audit_log (at);

CREATE TABLE settings (
  key   TEXT PRIMARY KEY,
  value TEXT NOT NULL
) STRICT;
INSERT INTO settings (key, value) VALUES
  ('mesh_cidr',    '10.77.0.0/16'),
  ('upstream_dns', '1.1.1.1,9.9.9.9'),
  ('default_mtu',  '1420'),
  ('default_hub',  '');
