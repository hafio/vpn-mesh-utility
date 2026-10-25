package main

import (
	"context"
	"flag"
	"fmt"
	"net/url"
	"time"

	"github.com/hafio/vpn-mesh-utility/internal/agent/localapi"
)

// setupUI is `meshctl ui`: the link goes alone to stdout, the note to
// stderr, so the link can be piped on.
func setupUI(fs *flag.FlagSet, e *env) action {
	_, sock := socketFlags(fs, e)
	return func(ctx context.Context, args []string) error {
		if err := wantArgs(args, 0, ""); err != nil {
			return err
		}
		tok, err := localapi.NewClient(sock()).UIToken(ctx)
		if err != nil {
			return err
		}
		fmt.Fprintln(e.stdout, tok.URL)
		left := time.Unix(tok.ExpiresAt, 0).Sub(e.now()).Round(time.Second)
		fmt.Fprintf(e.stderr, "Open this link in a browser on this machine within %s; it works once.\n", left)
		port := "7443"
		if u, err := url.Parse(tok.URL); err == nil && u.Port() != "" {
			port = u.Port() // the mesh's slot's (amendment A49), or --ui-listen's
		}
		fmt.Fprintf(e.stderr, "From another machine: ssh -L %s:127.0.0.1:%s to this node, then open the link there.\n", port, port)
		return nil
	}
}
