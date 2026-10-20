-- Relay failover (amendment A40): when the server judged a hub down (0: it
-- is up), and since when that hub has looked healthy again (0: it has not).
-- Operational state, like last_seen: never part of a snapshot or an export.
ALTER TABLE nodes ADD COLUMN down_since INTEGER NOT NULL DEFAULT 0;
ALTER TABLE nodes ADD COLUMN healthy_since INTEGER NOT NULL DEFAULT 0;
