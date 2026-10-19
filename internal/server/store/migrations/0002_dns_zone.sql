-- The DNS zone every node serves (amendment A36). It was fixed to mesh
-- before; existing meshes keep it.
INSERT OR IGNORE INTO settings (key, value) VALUES ('dns_zone', 'mesh');
