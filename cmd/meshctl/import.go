package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/hafio/vpn-mesh-utility/internal/server"
	"github.com/hafio/vpn-mesh-utility/internal/server/export"
	"github.com/hafio/vpn-mesh-utility/internal/validate"
)

func setupImport(fs *flag.FlagSet, e *env) action {
	data := fs.String("data", "/data", "the server's data directory (the server must be stopped)")
	return func(ctx context.Context, args []string) error {
		if err := wantArgs(args, 1, "<file>"); err != nil {
			return err
		}
		file, err := validate.FilePath("file", args[0])
		if err != nil {
			return usagef("%v", err)
		}
		f, err := os.Open(file)
		if err != nil {
			return fmt.Errorf("read %s: %w", file, err)
		}
		doc, err := export.Read(f)
		_ = f.Close() // read-only; the decode error is the one that matters
		if err != nil {
			return err
		}
		res, err := server.Import(ctx, server.ImportConfig{DataDir: *data, Doc: doc, Now: e.now})
		if err != nil {
			return err
		}
		if res.Version == 0 {
			fmt.Fprintf(e.stdout, "imported %d nodes from %s; the mesh already matched it, so no config changed\n", len(doc.Nodes), file)
		} else {
			fmt.Fprintf(e.stdout, "imported %d nodes from %s as config v%d; agents pick it up when the server starts\n", len(doc.Nodes), file, res.Version)
		}
		for _, w := range res.Warnings {
			fmt.Fprintln(e.stdout, w.String())
		}
		return nil
	}
}
