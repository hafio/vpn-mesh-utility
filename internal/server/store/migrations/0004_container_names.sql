-- Container names (amendment A38): what a node's agent reported its
-- runtime calls the containers on its mesh network. Operational state, like
-- peer_status: replaced by every report, never part of a snapshot or an
-- export.
CREATE TABLE container_names (
  node_id INTEGER NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
  name    TEXT    NOT NULL,
  ip      TEXT    NOT NULL,
  kind    TEXT    NOT NULL,
  PRIMARY KEY (node_id, name, ip)
) STRICT;
