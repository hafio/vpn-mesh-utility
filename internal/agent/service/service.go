// Package service installs the agent as a systemd service (spec S6.1 step 7):
// it writes deploy's static unit file and enables it through systemctl, run
// by the single exec site. Windows services arrive in milestone M5.
package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"

	"github.com/hafio/vpn-mesh-utility/deploy"
	"github.com/hafio/vpn-mesh-utility/internal/agent/execrun"
	"github.com/hafio/vpn-mesh-utility/internal/fsutil"
)

// Installer writes and starts the unit. The fields are seams for tests;
// Systemd fills them for the host.
type Installer struct {
	Run execrun.Runner
	// Root prefixes every host path (tests use a temp dir).
	Root string
	// Executable reports the running binary's path.
	Executable func() (string, error)
}

// Systemd is the installer for this host.
func Systemd() Installer {
	return Installer{Run: execrun.Exec{}, Executable: os.Executable}
}

// Available reports whether a service can be installed here, and why not.
func (i Installer) Available() (bool, string) {
	switch {
	case runtime.GOOS != "linux":
		return false, "service install on " + runtime.GOOS + " arrives in milestone M5"
	case exists(i.path("/.dockerenv")) || exists(i.path("/run/.containerenv")):
		return false, "running in a container; the container's restart policy runs `meshctl agent run`"
	case !exists(i.path("/run/systemd/system")):
		return false, "systemd is not running on this host"
	}
	return true, ""
}

func (i Installer) path(p string) string { return filepath.Join(i.Root, filepath.FromSlash(p)) }

func exists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

// Install writes the unit (if it changed), reloads systemd, enables the unit
// and restarts it so a re-join takes effect.
func (i Installer) Install(ctx context.Context) error {
	exe, err := i.Executable()
	if err != nil {
		return fmt.Errorf("find the meshctl binary: %w", err)
	}
	if resolved, err := filepath.EvalSymlinks(exe); err == nil {
		exe = resolved
	}
	if exe != deploy.BinaryPath {
		return fmt.Errorf("the service runs %s, but this meshctl is %s: install it there (install.sh does), or join with --no-service and run `meshctl agent run` yourself",
			deploy.BinaryPath, exe)
	}
	unit := i.path(deploy.UnitPath)
	cur, err := os.ReadFile(unit)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("read %s: %w", unit, err)
	}
	if !bytes.Equal(cur, deploy.SystemdUnit()) {
		if err := fsutil.WriteAtomic(filepath.Dir(unit), filepath.Base(unit), deploy.SystemdUnit(), 0o644); err != nil {
			return err
		}
	}
	for _, argv := range [][]string{
		{"systemctl", "daemon-reload"},
		{"systemctl", "enable", deploy.UnitName},
		{"systemctl", "restart", deploy.UnitName},
	} {
		if _, err := i.Run.Run(ctx, argv); err != nil {
			return fmt.Errorf("install service: %w", err)
		}
	}
	return nil
}
