// Package deploy holds the artifacts meshctl installs or serves: the agent's
// systemd unit, the install.sh template, the server's compose file and the
// spokes' generated one (nas.go). Each one states its security posture
// explicitly instead of inheriting defaults, and the tests pin it
// (docs/security.md explains every widening).
//
// It imports only internal/validate and yaml.v3, so the server (which serves
// install.sh and the spokes' compose files) and the agent (which installs
// the unit) can both use it.
package deploy

import (
	_ "embed"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/hafio/vpn-mesh-utility/internal/validate"
)

//go:embed meshctl-agent.service
var systemdUnit string

//go:embed install.sh.tmpl
var installTemplate string

//go:embed compose.server.yml
var composeServer string

// UnitPath is where `meshctl agent join` installs the unit.
const UnitPath = "/etc/systemd/system/meshctl-agent.service"

// UnitName is the agent's systemd unit.
const UnitName = "meshctl-agent.service"

// BinaryPath is where install.sh puts meshctl and where the unit runs it.
const BinaryPath = "/usr/local/bin/meshctl"

// SystemdUnit returns the agent's unit file. It is static: nothing from
// outside the binary is ever written into it.
func SystemdUnit() []byte { return []byte(systemdUnit) }

// ComposeServer returns the server's compose file (spec S12, amendment A13).
func ComposeServer() []byte { return []byte(composeServer) }

// install.sh markers. Each appears exactly once in the template, inside
// single quotes, and is replaced by a validated value that cannot contain a
// quote.
const (
	markerOrigin      = "@@MESHCTL_ORIGIN@@"
	markerVersion     = "@@MESHCTL_VERSION@@"
	markerSHA256Amd64 = "@@MESHCTL_SHA256_AMD64@@"
	markerSHA256Arm64 = "@@MESHCTL_SHA256_ARM64@@"
)

var reSHA256 = regexp.MustCompile(`^[0-9a-f]{64}$`)

// InstallParams are the values written into install.sh.
type InstallParams struct {
	// Origin is the server's public origin; the script downloads from it and
	// joins it.
	Origin string
	// Version is the meshctl version the script installs.
	Version string
	// SHA256Amd64 and SHA256Arm64 are the hex SHA-256 of
	// meshctl-linux-amd64 and meshctl-linux-arm64.
	SHA256Amd64 string
	SHA256Arm64 string
}

// RenderInstall returns install.sh for p. Every value is checked against its
// grammar first, so a bad value fails here instead of reaching a shell.
func RenderInstall(p InstallParams) ([]byte, error) {
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
	for arch, sum := range map[string]string{"amd64": p.SHA256Amd64, "arm64": p.SHA256Arm64} {
		if !reSHA256.MatchString(sum) {
			errs = append(errs, fmt.Errorf("sha256 of meshctl-linux-%s: must be 64 lowercase hex characters (is SHA256SUMS complete)", arch))
		}
	}
	if err := errors.Join(errs...); err != nil {
		return nil, fmt.Errorf("render install.sh: %w", err)
	}
	r := strings.NewReplacer(
		markerOrigin, p.Origin,
		markerVersion, p.Version,
		markerSHA256Amd64, p.SHA256Amd64,
		markerSHA256Arm64, p.SHA256Arm64,
	)
	return []byte(r.Replace(installTemplate)), nil
}
