-- Who manages what (amendment A49). An admin manages the control plane:
-- every mesh, the meshes themselves and the users. Any other user manages
-- the meshes granted to them, fully, and nothing else. Every user so far
-- could do everything, so each becomes an admin.
ALTER TABLE users ADD COLUMN is_admin INTEGER NOT NULL DEFAULT 0 CHECK (is_admin IN (0, 1));
UPDATE users SET is_admin = 1;

CREATE TABLE user_meshes (
  user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  mesh_id INTEGER NOT NULL REFERENCES meshes(id) ON DELETE CASCADE,
  PRIMARY KEY (user_id, mesh_id)
) STRICT;
