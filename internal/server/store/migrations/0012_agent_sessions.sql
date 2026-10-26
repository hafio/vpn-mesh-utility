-- One live connection per node (amendment A50). Operational: never in
-- snapshots or exports. Secrets are kept as sha256 hex.
--
-- A node's session: the ticket its next request presents and the one
-- before (a lost reply presents that again), when the holder was last
-- seen, whether it said goodbye, and a challenge that is open because
-- another copy of the node said hello (when, the waiting copy's hostname,
-- and whether the holder answered).
CREATE TABLE agent_sessions (
  node_id      INTEGER PRIMARY KEY REFERENCES nodes(id) ON DELETE CASCADE,
  session_hash TEXT    NOT NULL,
  ticket_hash  TEXT    NOT NULL,
  prev_hash    TEXT    NOT NULL DEFAULT '',
  last_seen    INTEGER NOT NULL,
  released     INTEGER NOT NULL DEFAULT 0,
  hostname     TEXT    NOT NULL DEFAULT '',
  challenge_at INTEGER NOT NULL DEFAULT 0,
  answered     INTEGER NOT NULL DEFAULT 0,
  waiting_host TEXT    NOT NULL DEFAULT ''
) STRICT;

-- The machine a node joined or last ran on: sha256 of the id it presents,
-- and its hostname for display.
CREATE TABLE node_machines (
  node_id      INTEGER PRIMARY KEY REFERENCES nodes(id) ON DELETE CASCADE,
  machine_hash TEXT    NOT NULL,
  hostname     TEXT    NOT NULL DEFAULT '',
  pinned_at    INTEGER NOT NULL
) STRICT;

-- The last copy of a node the control plane refused, as the dashboard
-- shows it.
CREATE TABLE agent_refusals (
  node_id      INTEGER PRIMARY KEY REFERENCES nodes(id) ON DELETE CASCADE,
  code         TEXT    NOT NULL,
  hostname     TEXT    NOT NULL DEFAULT '',
  same_machine INTEGER NOT NULL,
  at           INTEGER NOT NULL
) STRICT;
