//go:build !windows

package main

import (
	"context"

	"github.com/hafio/vpn-mesh-utility/internal/agent"
	"github.com/hafio/vpn-mesh-utility/internal/agent/service"
)

// hostService is the service join installs here: systemd.
func hostService(string) agent.Service { return service.Systemd() }

func installAgent(context.Context, *env, agent.JoinOptions) error {
	return usagef("`agent install` is for Windows; on Linux, run the install.sh command `meshctl node add` printed (or `meshctl agent join`)")
}

func uninstallAgent(context.Context, *env, string, bool) error {
	return usagef("`agent uninstall` is for Windows; on Linux: systemctl disable --now meshctl-agent")
}
