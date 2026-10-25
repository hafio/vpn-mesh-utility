package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/hafio/vpn-mesh-utility/internal/server/api"
	"github.com/hafio/vpn-mesh-utility/internal/server/export"
	"github.com/hafio/vpn-mesh-utility/internal/server/store"
	"github.com/hafio/vpn-mesh-utility/internal/validate"
)

// LockFile is the data dir's lock (see Lock).
const LockFile = "server.lock"

// ErrLocked means another process -- a running server, or another import --
// holds the data dir.
var ErrLocked = errors.New("the data dir is in use: stop the meshctl control plane first (import runs only while it is stopped)")

// ImportConfig is what Import needs.
type ImportConfig struct {
	DataDir string
	// Mesh names the mesh to replace; "" is the sole mesh. A named mesh
	// that does not exist is created first, as `meshctl mesh add` would
	// with the document's network and zone, and removed again when the
	// import fails.
	Mesh string
	Doc  export.Document
	Log  *slog.Logger
	Now  func() time.Time
}

// Import replaces a mesh's desired state with an export (`meshctl import`).
// It takes the data dir's lock, so it refuses to run beside a server; checks
// every value of the document; and applies it through the reconciler, as an
// audited config version that keeps every node id (agents sign with theirs).
// Node ids belong to the whole control plane, so an export whose ids
// another mesh holds is refused.
func Import(ctx context.Context, c ImportConfig) (api.Result, error) {
	if c.Log == nil {
		c.Log = slog.New(slog.DiscardHandler)
	}
	if _, err := validate.FilePath("--data", c.DataDir); err != nil {
		return api.Result{}, err
	}
	snap, err := export.ToSnapshot(c.Doc)
	if err != nil {
		return api.Result{}, err
	}
	if err := os.MkdirAll(c.DataDir, 0o700); err != nil {
		return api.Result{}, fmt.Errorf("create data dir %s: %w", c.DataDir, err)
	}
	unlock, err := Lock(c.DataDir)
	if err != nil {
		return api.Result{}, err
	}
	defer unlock()
	st, err := store.Open(ctx, filepath.Join(c.DataDir, DBFile), store.Options{Now: c.Now, Log: c.Log})
	if err != nil {
		return api.Result{}, err
	}
	defer st.Close()
	hub := api.NewHub()
	rec := api.NewReconciler(st, hub, c.Log)
	meshes := api.NewNodes(api.Deps{DB: st, Hub: hub, Reconciler: rec, Now: c.Now})
	by := api.Actor{Name: "cli"}
	mesh, err := pickMesh(ctx, st, c.Mesh)
	created := false
	if status, _ := api.ErrorMessage(err); c.Mesh != "" && status == http.StatusNotFound {
		req := api.MeshAddRequest{Name: c.Mesh, MeshCIDR: snap.Settings[store.SettingMeshCIDR], Zone: snap.Settings[store.SettingDNSZone],
			IPv6: api.IPv6Off}
		if _, err = meshes.AddMesh(ctx, by, req); err != nil {
			return api.Result{}, fmt.Errorf("create mesh %s for the import: %w", c.Mesh, err)
		}
		created = true
		mesh, err = pickMesh(ctx, st, c.Mesh)
	}
	if err != nil {
		return api.Result{}, err
	}
	res, err := rec.Apply(ctx, api.Change{Mesh: mesh.ID, Actor: by.Name, Action: "mesh.import", Summary: "import",
		Detail: map[string]any{"nodes": len(snap.Nodes), "dns_records": len(snap.DNS), "policies": len(snap.EffectiveSegmentation().Policies)}},
		func(tx *store.MeshTx, _ *api.Change) error { return tx.Restore(snap) })
	if err != nil && created {
		if rerr := meshes.RemoveMesh(ctx, by, c.Mesh); rerr != nil {
			return api.Result{}, fmt.Errorf("%w (and the mesh %s made for it stays: %w)", err, c.Mesh, rerr)
		}
	}
	return res, err
}

// pickMesh reads the mesh name names, or the sole mesh.
func pickMesh(ctx context.Context, st *store.Store, name string) (store.Mesh, error) {
	var m store.Mesh
	err := st.Tx(ctx, func(tx *store.Tx) error {
		var err error
		m, err = api.PickMesh(tx, name)
		return err
	})
	return m, err
}
