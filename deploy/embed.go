// Package deploy holds the artifacts meshctl installs or serves: the agent's
// systemd unit, the install, reinstall and uninstall script templates for
// Linux (.sh) and Windows (.ps1), the control plane's compose and stack
// files, and the samples of a node's compose and Swarm files the web UI
// fills with a node's values (samples.go, amendment A53; meshctl deploys
// none itself, amendment A48). docs/containers.md shows
// the samples, and samples_test.go holds them to the posture the files here
// state: each one states its security posture explicitly instead of
// inheriting defaults, and the tests pin it (docs/security.md explains every
// widening).
//
// It imports only internal/validate. The scripts install the agent's
// service (amendment A51); the binary installs none.
package deploy

import (
	_ "embed"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/hafio/vpn-mesh-utility/internal/validate"
)

//go:embed meshctl-agent@.service
var systemdUnit string

//go:embed install.sh.tmpl
var installTemplate string

//go:embed reinstall.sh.tmpl
var reinstallTemplate string

//go:embed uninstall.sh.tmpl
var uninstallTemplate string

//go:embed lib.sh.tmpl
var libTemplate string

//go:embed install.ps1.tmpl
var installPS1Template string

//go:embed reinstall.ps1.tmpl
var reinstallPS1Template string

//go:embed uninstall.ps1.tmpl
var uninstallPS1Template string

//go:embed lib.ps1.tmpl
var libPS1Template string

//go:embed compose.control.yml
var composeControl string

// UnitName is the agent's systemd unit, a template: install.sh starts the
// instance meshctl-agent@<mesh> for the mesh a node joins (amendment A51).
const UnitName = "meshctl-agent@.service"

// SystemdUnit returns the agent's unit file. It is static: nothing from
// outside the binary is ever written into it.
func SystemdUnit() []byte { return []byte(systemdUnit) }

// ComposeControl returns the control plane's compose file (spec S12,
// amendments A13, A48).
func ComposeControl() []byte { return []byte(composeControl) }

// The scripts' markers. Each value marker appears exactly once in the shared
// part, lib.sh.tmpl or lib.ps1.tmpl, inside single quotes, and is replaced by
// a validated value that cannot contain a quote.
const (
	// markerLib is a line of its own in each install, reinstall and
	// uninstall script (amendment A55): the part the scripts of one system
	// share, lib.sh.tmpl or lib.ps1.tmpl, which holds the other markers,
	// takes its place.
	markerLib         = "@@MESHCTL_LIB@@\n"
	markerOrigin      = "@@MESHCTL_ORIGIN@@"
	markerVersion     = "@@MESHCTL_VERSION@@"
	markerSHA256Amd64 = "@@MESHCTL_SHA256_AMD64@@"
	markerSHA256Arm64 = "@@MESHCTL_SHA256_ARM64@@"
	// markerUnit is a line of its own inside a quoted here-document; the
	// unit, which is static, takes its place.
	markerUnit = "@@MESHCTL_UNIT@@\n"
	// markerSHA256Windows is the Windows scripts' checksum of
	// meshctl-windows-amd64.exe.
	markerSHA256Windows = "@@MESHCTL_SHA256_WINDOWS@@"
)

var reSHA256 = regexp.MustCompile(`^[0-9a-f]{64}$`)

// InstallParams are the values written into the install, reinstall and
// uninstall scripts.
type InstallParams struct {
	// Origin is the server's public origin; the script downloads from it and
	// joins it.
	Origin string
	// Version is the meshctl version the script installs.
	Version string
	// SHA256Amd64 and SHA256Arm64 are the hex SHA-256 of
	// meshctl-linux-amd64 and meshctl-linux-arm64 (the .sh scripts),
	// SHA256Windows of meshctl-windows-amd64.exe (the .ps1 scripts).
	SHA256Amd64   string
	SHA256Arm64   string
	SHA256Windows string
}

// RenderInstall returns install.sh for p. Every value is checked against its
// grammar first, so a bad value fails here instead of reaching a shell.
func RenderInstall(p InstallParams) ([]byte, error) {
	return renderShell("install.sh", installTemplate, p)
}

// RenderReinstall returns reinstall.sh for p (amendment A55), checked as
// RenderInstall checks it.
func RenderReinstall(p InstallParams) ([]byte, error) {
	return renderShell("reinstall.sh", reinstallTemplate, p)
}

// RenderUninstall returns uninstall.sh for p (amendment A55), checked as
// RenderInstall checks it.
func RenderUninstall(p InstallParams) ([]byte, error) {
	return renderShell("uninstall.sh", uninstallTemplate, p)
}

// renderShell writes the shared part into the Linux script tmpl, then p's
// values into the markers.
func renderShell(name, tmpl string, p InstallParams) ([]byte, error) {
	errs := p.check(map[string]string{"meshctl-linux-amd64": p.SHA256Amd64, "meshctl-linux-arm64": p.SHA256Arm64})
	if err := errors.Join(errs...); err != nil {
		return nil, fmt.Errorf("render %s: %w", name, err)
	}
	r := strings.NewReplacer(
		markerOrigin, p.Origin,
		markerVersion, p.Version,
		markerSHA256Amd64, p.SHA256Amd64,
		markerSHA256Arm64, p.SHA256Arm64,
		markerUnit, systemdUnit,
	)
	return []byte(r.Replace(withLib(tmpl, libTemplate))), nil
}

// withLib is the script tmpl with the shared part lib in its marker's place.
func withLib(tmpl, lib string) string { return strings.Replace(tmpl, markerLib, lib, 1) }

// RenderInstallPS1 returns install.ps1 for p, its values checked as
// RenderInstall checks them.
func RenderInstallPS1(p InstallParams) ([]byte, error) {
	return renderPS1("install.ps1", installPS1Template, p)
}

// RenderReinstallPS1 returns reinstall.ps1 for p (amendment A55), checked as
// RenderInstall checks it.
func RenderReinstallPS1(p InstallParams) ([]byte, error) {
	return renderPS1("reinstall.ps1", reinstallPS1Template, p)
}

// RenderUninstallPS1 returns uninstall.ps1 for p (amendment A55), checked as
// RenderInstall checks it.
func RenderUninstallPS1(p InstallParams) ([]byte, error) {
	return renderPS1("uninstall.ps1", uninstallPS1Template, p)
}

// renderPS1 writes the shared part into the Windows script tmpl, then p's
// values into the markers.
func renderPS1(name, tmpl string, p InstallParams) ([]byte, error) {
	errs := p.check(map[string]string{"meshctl-windows-amd64.exe": p.SHA256Windows})
	if err := errors.Join(errs...); err != nil {
		return nil, fmt.Errorf("render %s: %w", name, err)
	}
	r := strings.NewReplacer(
		markerOrigin, p.Origin,
		markerVersion, p.Version,
		markerSHA256Windows, p.SHA256Windows,
	)
	return []byte(r.Replace(withLib(tmpl, libPS1Template))), nil
}

// check holds the origin, the version and the checksums of sums (by
// file) to their grammars.
func (p InstallParams) check(sums map[string]string) []error {
	var errs []error
	origin, err := validate.ServerURL("origin", p.Origin)
	if err != nil {
		errs = append(errs, err)
	} else if origin != p.Origin {
		errs = append(errs, fmt.Errorf("origin: must be in canonical form %q (got %q)", origin, p.Origin))
	}
	if _, err := validate.Version("version", p.Version); err != nil {
		errs = append(errs, err)
	}
	for file, sum := range sums {
		if !reSHA256.MatchString(sum) {
			errs = append(errs, fmt.Errorf("sha256 of %s: must be 64 lowercase hex characters (is SHA256SUMS complete)", file))
		}
	}
	return errs
}
