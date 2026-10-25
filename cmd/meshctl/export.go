package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hafio/vpn-mesh-utility/internal/server"
	"github.com/hafio/vpn-mesh-utility/internal/server/api"
	"github.com/hafio/vpn-mesh-utility/internal/server/export"
	"github.com/hafio/vpn-mesh-utility/internal/server/store"
	"github.com/hafio/vpn-mesh-utility/internal/validate"
)

func setupExport(fs *flag.FlagSet, e *env) action {
	data := fs.String("data", "/data", "the server's data directory")
	out := fs.String("out", "", "write to this file (mode 0600) instead of stdout")
	return func(ctx context.Context, args []string) error {
		if err := wantArgs(args, 0, ""); err != nil {
			return err
		}
		if *out != "" {
			if _, err := validate.FilePath("--out", *out); err != nil {
				return usagef("%v", err)
			}
		}
		// Read-only: works beside a running server and never migrates its
		// database.
		st, err := store.Open(ctx, filepath.Join(*data, server.DBFile), store.Options{ReadOnly: true})
		if err != nil {
			return err
		}
		defer st.Close()
		var snap store.Snapshot
		if err := st.Tx(ctx, func(tx *store.Tx) error {
			m, err := api.SoleMesh(tx)
			if err != nil {
				return err
			}
			snap, err = tx.In(m.ID).Snapshot()
			return err
		}); err != nil {
			return err
		}
		var b bytes.Buffer
		if err := export.Write(&b, export.FromSnapshot(snap)); err != nil {
			return err
		}
		if *out == "" {
			_, err := e.stdout.Write(b.Bytes())
			return err
		}
		if err := os.WriteFile(*out, b.Bytes(), 0o600); err != nil {
			return fmt.Errorf("write %s: %w", *out, err)
		}
		// WriteFile keeps the mode of an existing file.
		if err := os.Chmod(*out, 0o600); err != nil {
			return fmt.Errorf("restrict %s: %w", *out, err)
		}
		return nil
	}
}
