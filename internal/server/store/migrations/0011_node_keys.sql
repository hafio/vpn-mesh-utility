-- Every identity key a node had (amendment A50). A join makes fresh keys,
-- and a new token cuts the node's current key off at once; the old keys
-- stay here, so a copy of an earlier data dir is told it was replaced
-- instead of retrying a bare refusal forever. Operational: never in
-- snapshots or exports.
CREATE TABLE node_keys (
  node_id         INTEGER NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
  identity_pubkey TEXT    NOT NULL,
  replaced_at     INTEGER NOT NULL,
  PRIMARY KEY (node_id, identity_pubkey)
) STRICT;

CREATE INDEX node_keys_identity_pubkey ON node_keys (identity_pubkey);
