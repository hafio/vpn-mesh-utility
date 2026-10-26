package main

import (
	"context"
	"flag"

	"github.com/hafio/vpn-mesh-utility/internal/agent/localapi"
)

func setupLogs(fs *flag.FlagSet, e *env) action {
	_, sock := socketFlags(fs, e)
	follow := fs.Bool("f", false, "keep printing new lines until interrupted")
	return func(ctx context.Context, args []string) error {
		if err := wantArgs(args, 0, ""); err != nil {
			return err
		}
		path, err := sock()
		if err != nil {
			return err
		}
		return localapi.NewClient(path).Logs(ctx, *follow, e.stdout)
	}
}
