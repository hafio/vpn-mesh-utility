// Package deploy holds the artifacts meshctl installs or serves: the agent's
// systemd unit, the install.sh and install.ps1 templates, the control
// plane's compose and stack files, and the samples of a node's compose and
// Swarm files the web UI fills with a node's values (samples.go, amendment
// A53; meshctl deploys none itself, amendment A48). docs/containers.md shows
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

//go:embed install.ps1.tmpl
var installPS1Template string

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

// install.sh markers. Each appears exactly once in the template, inside
// single quotes, and is replaced by a validated value that cannot contain a
// quote.
const (
	markerOrigin      = "@@MESHCTL_ORIGIN@@"
	markerVersion     = "@@MESHCTL_VERSION@@"
	markerSHA256Amd64 = "@@MESHCTL_SHA256_AMD64@@"
	markerSHA256Arm64 = "@@MESHCTL_SHA256_ARM64@@"
	// markerUnit is a line of its own inside a quoted here-document; the
	// unit, which is static, takes its place.
	markerUnit = "@@MESHCTL_UNIT@@\n"
	// markerSHA256Windows is install.ps1's checksum of
	// meshctl-windows-amd64.exe.
	markerSHA256Windows = "@@MESHCTL_SHA256_WINDOWS@@"
)

var reSHA256 = regexp.MustCompile(`^[0-9a-f]{64}$`)

// InstallParams are the values written into install.sh and install.ps1.
type InstallParams struct {
	// Origin is the server's public origin; the script downloads from it and
	// joins it.
	Origin string
	// Version is the meshctl version the script installs.
	Version string
	// SHA256Amd64 and SHA256Arm64 are the hex SHA-256 of
	// meshctl-linux-amd64 and meshctl-linux-arm64 (install.sh),
	// SHA256Windows of meshctl-windows-amd64.exe (install.ps1).
	SHA256Amd64   string
	SHA256Arm64   string
	SHA256Windows string
}

// RenderInstall returns install.sh for p. Every value is checked against its
// grammar first, so a bad value fails here instead of reaching a shell.
func RenderInstall(p InstallParams) ([]byte, error) {
	errs := p.check(map[string]string{"meshctl-linux-amd64": p.SHA256Amd64, "meshctl-linux-arm64": p.SHA256Arm64})
	if err := errors.Join(errs...); err != nil {
		return nil, fmt.Errorf("render install.sh: %w", err)
	}
	r := strings.NewReplacer(
		markerOrigin, p.Origin,
		markerVersion, p.Version,
		markerSHA256Amd64, p.SHA256Amd64,
		markerSHA256Arm64, p.SHA256Arm64,
		markerUnit, systemdUnit,
	)
	return []byte(r.Replace(installTemplate)), nil
}

// RenderInstallPS1 returns install.ps1 for p, its values checked as
// RenderInstall checks them.
func RenderInstallPS1(p InstallParams) ([]byte, error) {
	errs := p.check(map[string]string{"meshctl-windows-amd64.exe": p.SHA256Windows})
	if err := errors.Join(errs...); err != nil {
		return nil, fmt.Errorf("render install.ps1: %w", err)
	}
	r := strings.NewReplacer(
		markerOrigin, p.Origin,
		markerVersion, p.Version,
		markerSHA256Windows, p.SHA256Windows,
	)
	return []byte(r.Replace(installPS1Template)), nil
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
