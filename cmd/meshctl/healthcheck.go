package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/hafio/vpn-mesh-utility/internal/agent/localapi"
	"github.com/hafio/vpn-mesh-utility/internal/proto"
)

// The image is distroless (no curl), so the container healthchecks run the
// binary itself.

func healthCommands() []command {
	return []command{
		{name: "agent", summary: "exit 0 when the local agent answers on its socket", setup: setupHealthAgent},
		{name: "server", summary: "exit 0 when the server answers /healthz", setup: setupHealthServer},
	}
}

func setupHealthServer(fs *flag.FlagSet, _ *env) action {
	addr := fs.String("addr", "127.0.0.1:8080", "host:port the server listens on")
	return func(ctx context.Context, args []string) error {
		if err := wantArgs(args, 0, ""); err != nil {
			return err
		}
		if _, _, err := net.SplitHostPort(*addr); err != nil {
			return usagef("--addr must be host:port: %v", err)
		}
		ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
		defer cancel()
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://"+*addr+proto.PathHealth, nil)
		if err != nil {
			return usagef("--addr: %v", err)
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return fmt.Errorf("server unhealthy: %w", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("server unhealthy: %s answered %s", proto.PathHealth, resp.Status)
		}
		return nil
	}
}

func setupHealthAgent(fs *flag.FlagSet, e *env) action {
	_, sock := socketFlags(fs, e)
	return func(ctx context.Context, args []string) error {
		if err := wantArgs(args, 0, ""); err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
		defer cancel()
		if _, err := localapi.NewClient(sock()).Status(ctx); err != nil {
			return fmt.Errorf("agent unhealthy: %w", err)
		}
		return nil
	}
}
