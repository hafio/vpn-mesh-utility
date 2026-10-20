-- Container networks (amendment A37): where a node's WireGuard lives (host,
-- or the agent container's own network namespace) and the node's /24 for
-- containers, empty while its containers are not on the mesh.
ALTER TABLE nodes ADD COLUMN mode TEXT NOT NULL DEFAULT 'host' CHECK (mode IN ('host', 'container'));
ALTER TABLE nodes ADD COLUMN container_subnet TEXT NOT NULL DEFAULT '';
CREATE UNIQUE INDEX nodes_container_subnet ON nodes (mesh_id, container_subnet) WHERE container_subnet != '';
