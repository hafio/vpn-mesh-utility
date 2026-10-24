package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
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
	Doc     export.Document
	Log     *slog.Logger
	Now     func() time.Time
}

// Import replaces the desired state with an export (`meshctl import`). It
// takes the data dir's lock, so it refuses to run beside a server; checks
// every value of the document; and applies it through the reconciler, as an
// audited config version that keeps every node id (agents sign with theirs).
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
	rec := api.NewReconciler(st, api.NewHub(), c.Log)
	return rec.Apply(ctx, api.Change{Actor: "cli", Action: "mesh.import", Summary: "import",
		Detail: map[string]any{"nodes": len(snap.Nodes), "dns_records": len(snap.DNS), "policies": len(snap.EffectiveSegmentation().Policies)}},
		func(tx *store.Tx, _ *api.Change) error { return tx.Restore(snap) })
}
