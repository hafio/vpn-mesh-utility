//go:build windows

package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hafio/vpn-mesh-utility/internal/agent"
	"github.com/hafio/vpn-mesh-utility/internal/agent/service"
	"github.com/hafio/vpn-mesh-utility/internal/fsutil"
)

// hostService is the service join installs here: the Windows service,
// running this binary.
func hostService(data string) agent.Service { return service.Windows(data) }

// installAgent is `meshctl agent install`: meshctl copied into Program
// Files, then join, with the service set to run that copy.
func installAgent(ctx context.Context, e *env, o agent.JoinOptions) error {
	pf := e.getenv("ProgramFiles")
	if pf == "" {
		pf = `C:\Program Files`
	}
	dest := filepath.Join(pf, "meshctl", "meshctl.exe")
	scm := service.SCM{DataDir: o.DataDir, Executable: func() (string, error) { return dest, nil }}
	if ok, why := scm.Available(); !ok {
		return errors.New(why)
	}
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("find the meshctl binary: %w", err)
	}
	if !strings.EqualFold(filepath.Clean(exe), filepath.Clean(dest)) {
		// The service runs the copy, which must not be in use while it is
		// replaced.
		if err := scm.Stop(ctx); err != nil {
			return err
		}
		b, err := os.ReadFile(exe)
		if err != nil {
			return fmt.Errorf("read %s: %w", exe, err)
		}
		if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			return fmt.Errorf("create %s: %w", filepath.Dir(dest), err)
		}
		if err := fsutil.WriteAtomic(filepath.Dir(dest), filepath.Base(dest), b, 0o755); err != nil {
			return err
		}
		fmt.Fprintf(e.stdout, "installed %s\n", dest)
	}
	o.Service = scm
	_, err = agent.Join(ctx, o)
	return err
}

// uninstallAgent is `meshctl agent uninstall`: the service stopped and
// removed, and with purge the data dir too. The binary stays: a running
// program cannot remove itself on Windows.
func uninstallAgent(ctx context.Context, e *env, data string, purge bool) error {
	scm := service.Windows(data)
	if ok, why := scm.Available(); !ok {
		return errors.New(why)
	}
	if err := scm.Uninstall(ctx); err != nil {
		return err
	}
	fmt.Fprintf(e.stdout, "removed the %s service\n", service.Name)
	if purge {
		if err := os.RemoveAll(data); err != nil {
			return fmt.Errorf("remove %s: %w", data, err)
		}
		fmt.Fprintf(e.stdout, "removed %s\n", data)
	}
	return nil
}
