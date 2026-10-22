-- Dual stack (M8): a node's IPv6 mesh address (mesh_ip6, reserved since
-- 0001), its address families (dual, ipv4 or ipv6: which mesh addresses it
-- has and which endpoints it dials), a server's public [IPv6]:port, and its
-- containers' IPv6 /112, empty while their IPv6 is not on the mesh; a
-- custom DNS record's IPv6 target (an IPv6-only fixed record keeps ip ''
-- for 0001's check). The mesh's IPv6 prefix is the setting mesh_cidr6.
-- host_families is operational, like down_since: the families the node's
-- agent last reported its host can use ('ipv4', 'ipv6', 'ipv4,ipv6'; ''
-- when it has not said).
ALTER TABLE nodes ADD COLUMN ip_family TEXT NOT NULL DEFAULT 'dual' CHECK (ip_family IN ('dual', 'ipv4', 'ipv6'));
ALTER TABLE nodes ADD COLUMN endpoint6 TEXT NOT NULL DEFAULT '';
ALTER TABLE nodes ADD COLUMN container_subnet6 TEXT NOT NULL DEFAULT '';
ALTER TABLE nodes ADD COLUMN host_families TEXT NOT NULL DEFAULT '';
CREATE UNIQUE INDEX nodes_mesh_ip6 ON nodes (mesh_id, mesh_ip6) WHERE mesh_ip6 IS NOT NULL;
CREATE UNIQUE INDEX nodes_container_subnet6 ON nodes (mesh_id, container_subnet6) WHERE container_subnet6 != '';
ALTER TABLE dns_records ADD COLUMN ip6 TEXT;
