package server

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"time"

	"github.com/hafio/vpn-mesh-utility/internal/fsutil"
	"github.com/hafio/vpn-mesh-utility/internal/server/api"
	"github.com/hafio/vpn-mesh-utility/internal/server/export"
	"github.com/hafio/vpn-mesh-utility/internal/server/store"
)

// Nightly exports (spec S8.1).
const (
	// ExportsDir holds them, under the data dir.
	ExportsDir = "exports"
	// exportHour is the local hour they are written at.
	exportHour  = 3
	keepExports = 14
)

// Exporter writes each mesh's desired state to
// <data>/exports/<mesh>/mesh-YYYY-MM-DD.yaml once a night, skipping nights
// when nothing changed and keeping each mesh's newest keepExports files.
// The output is deterministic, so the files diff cleanly in git.
type Exporter struct {
	DB  api.DB
	Dir string
	Now func() time.Time
	Log *slog.Logger
}

// NextExport is the first exportHour:00 strictly after now, in now's zone.
func NextExport(now time.Time) time.Time {
	next := time.Date(now.Year(), now.Month(), now.Day(), exportHour, 0, 0, 0, now.Location())
	if !next.After(now) {
		next = next.AddDate(0, 0, 1)
	}
	return next
}

// Run exports every night until ctx ends. A failed export is logged and
// retried the next night; it never stops the server.
func (e *Exporter) Run(ctx context.Context) {
	for {
		wait := NextExport(e.Now()).Sub(e.Now())
		t := time.NewTimer(wait)
		select {
		case <-ctx.Done():
			t.Stop()
			return
		case <-t.C:
		}
		files, err := e.ExportOnce(ctx)
		for _, f := range files {
			e.Log.Info("nightly export written", "file", f)
		}
		if err != nil {
			e.Log.Error("nightly export failed; it is retried tomorrow", "err", err)
		}
	}
}

// ExportOnce writes today's export of every mesh whose state differs from
// its newest export, then prunes old files. It returns the files it wrote.
func (e *Exporter) ExportOnce(ctx context.Context) ([]string, error) {
	snaps := map[string]store.Snapshot{}
	if err := e.DB.Tx(ctx, func(tx *store.Tx) error {
		meshes, err := tx.Meshes()
		if err != nil {
			return err
		}
		for _, m := range meshes {
			if snaps[m.Name], err = tx.In(m.ID).Snapshot(); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}
	var written []string
	for _, name := range slices.Sorted(maps.Keys(snaps)) {
		file, err := e.exportMesh(filepath.Join(e.Dir, name), snaps[name])
		if err != nil {
			return written, fmt.Errorf("export mesh %s: %w", name, err)
		}
		if file != "" {
			written = append(written, file)
		}
	}
	return written, nil
}

// exportMesh writes one mesh's export into dir unless it equals the newest
// one there, and prunes old files. It returns the file it wrote, or "".
func (e *Exporter) exportMesh(dir string, snap store.Snapshot) (string, error) {
	var b bytes.Buffer
	if err := export.Write(&b, export.FromSnapshot(snap)); err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", fmt.Errorf("create %s: %w", dir, err)
	}
	files, err := filepath.Glob(filepath.Join(dir, "mesh-*.yaml"))
	if err != nil {
		return "", fmt.Errorf("list exports: %w", err)
	}
	sort.Strings(files) // the date in the name sorts chronologically
	if n := len(files); n > 0 {
		if last, err := os.ReadFile(files[n-1]); err == nil && bytes.Equal(last, b.Bytes()) {
			return "", nil
		}
	}
	name := "mesh-" + e.Now().Format("2006-01-02") + ".yaml"
	if err := fsutil.WriteAtomic(dir, name, b.Bytes(), 0o600); err != nil {
		return "", err
	}
	files, err = filepath.Glob(filepath.Join(dir, "mesh-*.yaml"))
	if err != nil {
		return "", fmt.Errorf("list exports: %w", err)
	}
	sort.Strings(files)
	for len(files) > keepExports {
		if err := os.Remove(files[0]); err != nil {
			return "", fmt.Errorf("prune %s: %w", files[0], err)
		}
		files = files[1:]
	}
	return filepath.Join(dir, name), nil
}
