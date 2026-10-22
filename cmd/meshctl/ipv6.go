package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/hafio/vpn-mesh-utility/internal/server/api"
)

// `meshctl ipv6` shows or sets the mesh's IPv6 prefix (M8) on the running
// server, over its admin socket: setting it gives every node that may have
// IPv6 an address, off takes them away.

func setupIPv6(fs *flag.FlagSet, e *env) action {
	client, asJSON := adminFlags(fs)
	return func(ctx context.Context, args []string) error {
		if len(args) > 1 {
			return usagef("give at most one of <prefix>, auto or off")
		}
		var (
			v   api.IPv6View
			err error
		)
		if len(args) == 0 {
			v, err = client().IPv6(ctx)
		} else {
			v, err = client().SetIPv6(ctx, api.IPv6Request{Prefix: args[0]})
		}
		switch {
		case err != nil:
			return err
		case *asJSON:
			return printJSON(e.stdout, v)
		case v.Prefix == "":
			fmt.Fprintln(e.stdout, "the mesh has no IPv6 prefix: it runs IPv4 only (`meshctl ipv6 auto` gives it one)")
		default:
			fmt.Fprintf(e.stdout, "IPv6 prefix %s; %d node(s) with an IPv6 mesh address\n", v.Prefix, v.Addressed)
		}
		if v.Version != 0 {
			fmt.Fprintf(e.stdout, "config v%d is on its way to the agents\n", v.Version)
		}
		printWarnings(e.stdout, v.Warnings)
		return nil
	}
}
