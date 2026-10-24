package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/hafio/vpn-mesh-utility/internal/agent"
	"github.com/hafio/vpn-mesh-utility/internal/agent/localapi"
	"github.com/hafio/vpn-mesh-utility/internal/proto"
)

// The image is distroless (no curl), so the container healthchecks run the
// binary itself.

func healthCommands() []command {
	return []command{
		{name: "agent", summary: "exit 0 when the local agent answers on its socket", setup: setupHealthAgent},
		{name: "control", summary: "exit 0 when the control plane answers /healthz", setup: setupHealthControl},
		{name: "start", summary: "exit 0 when what `meshctl start` runs answers: the image's health check, from MESHCTL_MODE and MESHCTL_PORT or MESHCTL_DATA", setup: setupHealthStart},
	}
}

func setupHealthControl(fs *flag.FlagSet, _ *env) action {
	addr := fs.String("addr", "127.0.0.1:8080", "host:port the control plane listens on")
	return func(ctx context.Context, args []string) error {
		if err := wantArgs(args, 0, ""); err != nil {
			return err
		}
		if _, _, err := net.SplitHostPort(*addr); err != nil {
			return usagef("--addr must be host:port: %v", err)
		}
		return checkControl(ctx, *addr)
	}
}

func setupHealthAgent(fs *flag.FlagSet, e *env) action {
	_, sock := socketFlags(fs, e)
	return func(ctx context.Context, args []string) error {
		if err := wantArgs(args, 0, ""); err != nil {
			return err
		}
		return checkAgent(ctx, sock())
	}
}

// setupHealthStart checks what `meshctl start` runs, found the way start
// finds it: MESHCTL_MODE (a node when unset), then MESHCTL_PORT or
// MESHCTL_DATA (amendments A47, A48).
func setupHealthStart(_ *flag.FlagSet, e *env) action {
	return func(ctx context.Context, args []string) error {
		if err := wantArgs(args, 0, ""); err != nil {
			return err
		}
		mode, err := startMode(e)
		if err != nil {
			return err
		}
		if mode == modeControl {
			port, err := startPort(e)
			if err != nil {
				return err
			}
			return checkControl(ctx, net.JoinHostPort("127.0.0.1", strconv.Itoa(port)))
		}
		data, err := startData(e)
		if err != nil {
			return err
		}
		return checkAgent(ctx, agent.DefaultSocket(e.goos, data))
	}
}

// checkControl asks the control plane at addr for /healthz.
func checkControl(ctx context.Context, addr string) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://"+addr+proto.PathHealth, nil)
	if err != nil {
		return usagef("--addr: %v", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("control plane unhealthy: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("control plane unhealthy: %s answered %s", proto.PathHealth, resp.Status)
	}
	return nil
}

// checkAgent asks the agent on its local API socket for its status.
func checkAgent(ctx context.Context, sock string) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	if _, err := localapi.NewClient(sock).Status(ctx); err != nil {
		return fmt.Errorf("agent unhealthy: %w", err)
	}
	return nil
}
