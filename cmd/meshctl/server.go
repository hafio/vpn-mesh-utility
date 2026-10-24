package main

import (
	"context"
	"flag"
	"time"

	"github.com/hafio/vpn-mesh-utility/internal/server"
	"github.com/hafio/vpn-mesh-utility/internal/server/api"
)

// setupControl runs the control plane (`meshctl control`; `meshctl server`
// before amendment A48).
func setupControl(fs *flag.FlagSet, e *env) action {
	data := fs.String("data", "/data", "data directory: the database and the admin socket")
	listen := fs.String("listen", ":8080", "address for the agent API, health check and downloads")
	url := fs.String("url", e.getenv("MESHCTL_URL"), "public origin agents and install.sh use, e.g. https://mesh.example.com (default $MESHCTL_URL)")
	proxies := fs.String("trusted-proxies", e.getenv("MESHCTL_TRUSTED_PROXIES"),
		"comma-separated CIDRs of reverse proxies whose X-Forwarded-For is believed (default $MESHCTL_TRUSTED_PROXIES)")
	dist := fs.String("dist", e.getenv("MESHCTL_DIST_DIR"), "directory with the binaries and SHA256SUMS served at /dl/ (default $MESHCTL_DIST_DIR)")
	level := fs.String("log-level", "info", "debug, info, warn or error")
	downAfter := fs.Duration("hub-down-after", api.DefaultHubDownAfter,
		"relay failover: a hub silent this long, with no handshake for half as long again, has its pairs relayed by another hub; 0 turns it off")
	return func(ctx context.Context, args []string) error {
		if err := wantArgs(args, 0, ""); err != nil {
			return err
		}
		if *downAfter != 0 && (*downAfter < time.Second || *downAfter > time.Hour) {
			return usagef("--hub-down-after must be 0 (off) or between 1s and 1h (got %s)", *downAfter)
		}
		if *url == "" {
			return usagef("--url (or MESHCTL_URL) is required: the public origin agents use, e.g. https://mesh.example.com")
		}
		log, err := newLogger(e.stderr, *level)
		if err != nil {
			return err
		}
		tp, err := parsePrefixes("--trusted-proxies", *proxies)
		if err != nil {
			return err
		}
		return server.Run(ctx, server.Config{
			DataDir: *data, Listen: *listen, URL: *url, DistDir: *dist, TrustedProxies: tp,
			Version: e.version, Log: log, Now: e.now, HubDownAfter: *downAfter,
		})
	}
}
