package server

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
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

// Exporter writes the desired state to <data>/exports/mesh-YYYY-MM-DD.yaml
// once a night, skipping nights when nothing changed and keeping the newest
// keepExports files. The output is deterministic, so the files diff cleanly
// in git.
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
		if file, err := e.ExportOnce(ctx); err != nil {
			e.Log.Error("nightly export failed; it is retried tomorrow", "err", err)
		} else if file != "" {
			e.Log.Info("nightly export written", "file", file)
		}
	}
}

// ExportOnce writes today's export unless it equals the newest one, then
// prunes old files. It returns the file it wrote, or "".
func (e *Exporter) ExportOnce(ctx context.Context) (string, error) {
	var snap store.Snapshot
	if err := e.DB.Tx(ctx, func(tx *store.Tx) error {
		var err error
		snap, err = tx.Snapshot()
		return err
	}); err != nil {
		return "", err
	}
	var b bytes.Buffer
	if err := export.Write(&b, export.FromSnapshot(snap)); err != nil {
		return "", err
	}
	if err := os.MkdirAll(e.Dir, 0o700); err != nil {
		return "", fmt.Errorf("create %s: %w", e.Dir, err)
	}
	files, err := filepath.Glob(filepath.Join(e.Dir, "mesh-*.yaml"))
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
	if err := fsutil.WriteAtomic(e.Dir, name, b.Bytes(), 0o600); err != nil {
		return "", err
	}
	files, err = filepath.Glob(filepath.Join(e.Dir, "mesh-*.yaml"))
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
	return filepath.Join(e.Dir, name), nil
}
