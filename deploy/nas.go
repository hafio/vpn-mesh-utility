package deploy

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	"go.yaml.in/yaml/v3"

	"github.com/hafio/vpn-mesh-utility/internal/validate"
)

// Image is the published image (spec S12). Release tags carry no "v": the
// image of meshctl v1.2.3 is Image:1.2.3.
const Image = "ghcr.io/hafio/vpn-mesh-utility"

// The two ways a spoke's container runs WireGuard (amendment A19).
const (
	// NASUserspace runs wireguard-go inside the agent and needs only the
	// host's /dev/net/tun. It is the default until the NAS is known to ship
	// the module (spec S16).
	NASUserspace = "userspace"
	// NASKernel uses the host's WireGuard module; no device is mapped.
	NASKernel = "kernel"
)

// DefaultNASData is the host directory that holds the agent's keys and
// cached config (spec S12).
const DefaultNASData = "/volume1/docker/meshctl"

// NASParams are the values in a spoke's compose file.
type NASParams struct {
	// Node is the spoke's name; it appears in the comments and the join
	// command only.
	Node string
	// Server is the server's public origin, for the join command.
	Server string
	// Version is the meshctl release the container runs (v1.2.3 or 1.2.3).
	Version string
	// Backend is NASUserspace or NASKernel; empty means NASUserspace.
	Backend string
	// DataDir is the host directory mounted at /data; empty means
	// DefaultNASData.
	DataDir string
}

// The compose file as a typed value, so yaml.v3 does every bit of quoting
// and no value can add a key. Field order is output order.
type nasFile struct {
	Name     string                `yaml:"name"`
	Services map[string]nasService `yaml:"services"`
}

type nasService struct {
	Image       string    `yaml:"image"`
	Command     []string  `yaml:"command"`
	User        string    `yaml:"user"`
	Privileged  bool      `yaml:"privileged"`
	NetworkMode string    `yaml:"network_mode"`
	CapDrop     []string  `yaml:"cap_drop"`
	CapAdd      []string  `yaml:"cap_add"`
	SecurityOpt []string  `yaml:"security_opt"`
	ReadOnly    bool      `yaml:"read_only"`
	Tmpfs       []string  `yaml:"tmpfs"`
	Devices     []string  `yaml:"devices,omitempty"`
	Volumes     []string  `yaml:"volumes"`
	Restart     string    `yaml:"restart"`
	MemLimit    string    `yaml:"mem_limit"`
	PidsLimit   int       `yaml:"pids_limit"`
	Healthcheck nasHealth `yaml:"healthcheck"`
}

type nasHealth struct {
	Test        []string `yaml:"test"`
	Interval    string   `yaml:"interval"`
	Timeout     string   `yaml:"timeout"`
	Retries     int      `yaml:"retries"`
	StartPeriod string   `yaml:"start_period"`
}

// NASService is the compose service name, for `docker compose run`.
const NASService = "meshctl-agent"

// RenderNASCompose returns the compose file that runs a spoke's agent as a
// container (spec S12, amendments A13 and A19). Every value is checked
// against its grammar first, so a bad one fails here instead of reaching
// the file.
func RenderNASCompose(p NASParams) ([]byte, error) {
	if p.Backend == "" {
		p.Backend = NASUserspace
	}
	if p.DataDir == "" {
		p.DataDir = DefaultNASData
	}
	var errs []error
	if _, err := validate.Name("node", p.Node); err != nil {
		errs = append(errs, err)
	}
	origin, err := validate.ServerURL("server", p.Server)
	if err != nil {
		errs = append(errs, err)
	} else if origin != p.Server {
		errs = append(errs, fmt.Errorf("server: must be in canonical form %q (got %q)", origin, p.Server))
	}
	if _, err := validate.SemVer("version", p.Version); err != nil {
		errs = append(errs, fmt.Errorf("%w: only releases have an image; this is a development build", err))
	}
	if _, err := validate.Enum("backend", p.Backend, NASUserspace, NASKernel); err != nil {
		errs = append(errs, err)
	}
	if _, err := validate.POSIXPath("data", p.DataDir); err != nil {
		errs = append(errs, err)
	}
	if err := errors.Join(errs...); err != nil {
		return nil, fmt.Errorf("render the compose file: %w", err)
	}
	return renderNAS(p)
}

// renderNAS writes p without checking it. It still holds the file's
// structure against any value: yaml.v3 quotes what it must, every '$' is
// doubled so compose does not interpolate it, and the comment lines cannot
// be broken.
func renderNAS(p NASParams) ([]byte, error) {
	lit := func(s string) string { return strings.ReplaceAll(s, "$", "$$") }
	svc := nasService{
		Image:       lit(Image + ":" + strings.TrimPrefix(p.Version, "v")),
		Command:     []string{"agent", "run", "--data", "/data", "--wg-backend", lit(p.Backend)},
		User:        "0:0",
		Privileged:  false,
		NetworkMode: "host",
		CapDrop:     []string{"ALL"},
		CapAdd:      []string{"NET_ADMIN", "NET_BIND_SERVICE"},
		SecurityOpt: []string{"no-new-privileges=true"},
		ReadOnly:    true,
		Tmpfs:       []string{"/tmp:size=16m,mode=1777"},
		Volumes:     []string{lit(p.DataDir) + ":/data"},
		Restart:     "unless-stopped",
		MemLimit:    "256m",
		PidsLimit:   256,
		Healthcheck: nasHealth{
			Test:        []string{"CMD", "/app/meshctl", "healthcheck", "agent", "--data", "/data"},
			Interval:    "30s",
			Timeout:     "5s",
			Retries:     3,
			StartPeriod: "10s",
		},
	}
	if p.Backend == NASUserspace {
		svc.Devices = []string{"/dev/net/tun:/dev/net/tun"}
	}
	var b bytes.Buffer
	for _, line := range nasHeader(p) {
		b.WriteString(strings.TrimRight("# "+comment(line), " ") + "\n")
	}
	enc := yaml.NewEncoder(&b)
	enc.SetIndent(2)
	if err := enc.Encode(nasFile{Name: NASService, Services: map[string]nasService{NASService: svc}}); err != nil {
		return nil, fmt.Errorf("render the compose file: %w", err)
	}
	if err := enc.Close(); err != nil {
		return nil, fmt.Errorf("render the compose file: %w", err)
	}
	return b.Bytes(), nil
}

// comment keeps a value on its comment line: control characters become '?'.
func comment(s string) string {
	return strings.Map(func(r rune) rune {
		if r < 0x20 || r == 0x7f {
			return '?'
		}
		return r
	}, s)
}

func nasHeader(p NASParams) []string {
	backend := "The userspace backend runs WireGuard in the agent, on the host's /dev/net/tun."
	if p.Backend == NASKernel {
		backend = "The kernel backend uses the host's WireGuard module (ls /sys/module/wireguard)."
	}
	return []string{
		fmt.Sprintf("meshctl agent for the spoke %q, as a container (spec S12; amendments A13, A19).", p.Node),
		backend,
		"",
		"Join once, with the node's join token (web UI, `meshctl node add` or `node token`):",
		fmt.Sprintf("  MESHCTL_TOKEN=<token> docker compose run --rm -e MESHCTL_TOKEN %s \\", NASService),
		fmt.Sprintf("    agent join --data /data --server %s --name %s --no-service", p.Server, p.Node),
		"then start it: docker compose up -d",
		"",
		"Posture (docs/security.md): not privileged; every capability dropped except",
		"NET_ADMIN (wg0, its routes) and NET_BIND_SERVICE (DNS on :53); no privilege",
		"escalation; read-only root. It runs as uid 0 because capabilities do not",
		"survive no-new-privileges for another user, and on the host network because",
		"wg0 belongs to the host.",
	}
}
