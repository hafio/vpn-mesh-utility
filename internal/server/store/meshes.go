package store

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
)

// Mesh is one row of meshes (amendment A49): a WireGuard mesh the control
// plane manages. Everything that shapes its configs belongs to one mesh.
type Mesh struct {
	ID   int64
	Name string
	// UUID names the mesh to its agents.
	UUID string
	// Slot (0..MaxSlot) picks the names an agent of this mesh uses on a
	// host, so agents of several meshes can share one.
	Slot      int
	CreatedAt int64
}

// FirstMesh is the mesh the first migration creates, named "default".
const FirstMesh int64 = 1

// MaxSlot is the highest slot a mesh may have.
const MaxSlot = 63

// MeshTx is a transaction scoped to one mesh. Its methods read and write
// that mesh's nodes, DNS records, settings, versions, segmentation and
// audit entries; the embedded Tx has what belongs to the control plane
// (users, sessions, tokens, nodes by id, heartbeats). A mesh-scoped
// accessor cannot be called without a mesh, so no query reaches another
// mesh's rows by accident.
type MeshTx struct {
	*Tx
	mesh int64
}

// In scopes t to a mesh.
func (t *Tx) In(mesh int64) *MeshTx { return &MeshTx{Tx: t, mesh: mesh} }

// MeshID is the mesh t is scoped to.
func (t *MeshTx) MeshID() int64 { return t.mesh }

// MeshTx runs fn in a transaction scoped to mesh.
func (s *Store) MeshTx(ctx context.Context, mesh int64, fn func(*MeshTx) error) error {
	return s.Tx(ctx, func(t *Tx) error { return fn(t.In(mesh)) })
}

const meshCols = `id, name, uuid, slot, created_at`

func scanMesh(r scanner) (Mesh, error) {
	var m Mesh
	err := r.Scan(&m.ID, &m.Name, &m.UUID, &m.Slot, &m.CreatedAt)
	return m, err
}

// Meshes returns every mesh, ordered by id.
func (t *Tx) Meshes() ([]Mesh, error) {
	rows, err := t.query(`SELECT ` + meshCols + ` FROM meshes ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("list meshes: %w", err)
	}
	var out []Mesh
	err = scanAll(rows, "meshes", func(r *sql.Rows) error {
		m, err := scanMesh(r)
		if err != nil {
			return err
		}
		out = append(out, m)
		return nil
	})
	return out, err
}

// MeshByID returns one mesh or ErrNotFound.
func (t *Tx) MeshByID(id int64) (Mesh, error) {
	m, err := scanMesh(t.queryRow(`SELECT `+meshCols+` FROM meshes WHERE id = ?`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return Mesh{}, fmt.Errorf("mesh %d: %w", id, ErrNotFound)
	}
	if err != nil {
		return Mesh{}, fmt.Errorf("read mesh %d: %w", id, err)
	}
	return m, nil
}

// MeshByName returns one mesh or ErrNotFound.
func (t *Tx) MeshByName(name string) (Mesh, error) {
	m, err := scanMesh(t.queryRow(`SELECT `+meshCols+` FROM meshes WHERE name = ?`, name))
	if errors.Is(err, sql.ErrNoRows) {
		return Mesh{}, fmt.Errorf("mesh %q: %w", name, ErrNotFound)
	}
	if err != nil {
		return Mesh{}, fmt.Errorf("read mesh %q: %w", name, err)
	}
	return m, nil
}

// InsertMesh creates a mesh with m's name and slot (validated by the
// caller) and sets m.ID, m.UUID and m.CreatedAt. A taken name or slot is
// ErrConflict.
func (t *Tx) InsertMesh(m *Mesh) error {
	id, err := newUUID()
	if err != nil {
		return err
	}
	m.UUID, m.CreatedAt = id, t.now.Unix()
	res, err := t.exec(`INSERT INTO meshes (name, uuid, slot, created_at) VALUES (?, ?, ?, ?)`, m.Name, m.UUID, m.Slot, m.CreatedAt)
	if err != nil {
		return fmt.Errorf("insert mesh %q: %w", m.Name, err)
	}
	if m.ID, err = res.LastInsertId(); err != nil {
		return fmt.Errorf("insert mesh %q: %w", m.Name, err)
	}
	return nil
}

// DeleteMesh removes a mesh that has no nodes, with its settings, records,
// versions, segmentation and audit entries. One with nodes is ErrConflict.
func (t *Tx) DeleteMesh(id int64) error {
	var n int
	if err := t.queryRow(`SELECT COUNT(*) FROM nodes WHERE mesh_id = ?`, id).Scan(&n); err != nil {
		return fmt.Errorf("count the nodes of mesh %d: %w", id, err)
	}
	if n > 0 {
		return fmt.Errorf("mesh %d still has %d node(s): %w", id, n, ErrConflict)
	}
	res, err := t.exec(`DELETE FROM meshes WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete mesh %d: %w", id, err)
	}
	return mustAffect(res, fmt.Sprintf("mesh %d", id))
}

// newUUID is a random (version 4) UUID.
func newUUID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", fmt.Errorf("generate a mesh uuid: %w", err)
	}
	b[6] = b[6]&0x0f | 0x40
	b[8] = b[8]&0x3f | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16]), nil
}
