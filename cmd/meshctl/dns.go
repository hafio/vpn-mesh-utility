package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/hafio/vpn-mesh-utility/internal/agent/localapi"
	"github.com/hafio/vpn-mesh-utility/internal/proto"
)

// `meshctl dns` shows or switches this node's split DNS (spec S7.4): whether
// the host sends its .mesh lookups to the node's resolver. The choice is the
// node's own and never reaches the server (amendment A10); the agent keeps
// it across restarts and applies it at once.
func setupDNS(fs *flag.FlagSet, e *env) action {
	_, sock := socketFlags(fs, e)
	return func(ctx context.Context, args []string) error {
		c := localapi.NewClient(sock())
		var (
			st  proto.LocalStatus
			err error
		)
		switch {
		case len(args) == 0:
			st, err = c.Status(ctx)
		case len(args) == 1 && (args[0] == "on" || args[0] == "off"):
			st, err = c.DNS(ctx, args[0] == "on")
		default:
			return usagef("expects on or off (got %q)", args)
		}
		if err != nil {
			return err
		}
		fmt.Fprintln(e.stdout, "dns "+dnsLine(st))
		return nil
	}
}

// dnsLine says what the node's DNS state means on this host.
func dnsLine(st proto.LocalStatus) string {
	zone := st.Zone
	if zone == "" {
		zone = proto.DefaultZone
	}
	switch st.DNS {
	case proto.DNSOn:
		return "on: this host sends ." + zone + " lookups to the resolver on " + st.MeshIP
	case proto.DNSOff:
		return "off on this node; the resolver on " + st.MeshIP + " still answers the others (`meshctl dns on` turns it back on)"
	case proto.DNSUnavailable:
		return "unavailable: the resolver answers ." + zone + " on " + st.MeshIP + ", but this host has no systemd-resolved to send its own lookups there"
	}
	return "not set up yet: the node has no DNS in its config"
}
