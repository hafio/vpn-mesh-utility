package deploy

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"strings"
	"text/template"

	"github.com/hafio/vpn-mesh-utility/internal/validate"
)

// Samples are the ways to run a node's agent, filled with the node's values
// (amendment A53): docker compose and Swarm stack files, and the commands
// that create its containers' mesh network. The web UI shows them after a
// node is added; docs/containers.md shows them with example values
// (TestContainersDocSamples). They are starting points the operator reads and
// runs, not a deployment meshctl makes (amendment A48).

//go:embed samples/*.tmpl
var sampleFS embed.FS

var sampleTemplates = template.Must(template.New("samples").Option("missingkey=error").ParseFS(sampleFS, "samples/*.tmpl"))

// Image is the published image's repository.
const Image = "ghcr.io/hafio/vpn-mesh-utility"

// Sample kinds.
const (
	ComposeHost      = "compose-host"
	StackHost        = "stack-host"
	ComposeContainer = "compose-container"
	StackContainer   = "stack-container"
	StackGlobal      = "stack-global"
	NetworkDocker    = "network-docker"
	NetworkPodman    = "network-podman"
	NetworkSwarm     = "network-swarm"
	// Docker Desktop on Windows, with Linux containers: PowerShell, and the
	// machine id from a file beside the compose file.
	NetworkWindows = "network-windows"
	ComposeWindows = "compose-windows"
)

// sampleFile is the template of a kind that shares another's.
var sampleFile = map[string]string{NetworkWindows: NetworkDocker, ComposeWindows: ComposeContainer}

// SampleParams are the values a node's samples are filled with: the node,
// and for its containers the values `meshctl node show` lists.
type SampleParams struct {
	// URL is the control plane's origin, Version its release (the image's
	// tag; a build that is no release gets the placeholder <release>).
	URL, Version string
	// Mesh and Name are the node's; Name is empty for StackGlobal.
	Mesh, Name string
	// Server is a server node (it publishes its WireGuard port); Container
	// a container-mode node.
	Server, Container bool
	// DataDir is a host directory for a host-network agent's data; empty
	// keeps it in a named volume.
	DataDir string
	// The containers' mesh network: its name and bridge, the mesh on the
	// link, the range the runtime hands out, the runtime's gateway, the
	// agent's address (container mode), and the same for IPv6.
	NetworkName, BridgeName        string
	Network, Range, Gateway, Agent string
	Network6, Range6, Gateway6     string
	ListenPort, PublishPort        int
}

// Sample is one way to run a node, ready to show.
type Sample struct {
	Kind  string
	Title string
	// Lang is the text's language: yaml, sh or powershell.
	Lang string
	Text string
	// Run starts it; <token> stands for the join token. Empty for commands.
	Run string
	// Note says when it fits.
	Note string
}

// view is what the templates read: the params with the derived values.
// Windows is a Docker Desktop sample: Cont continues a command's line
// (PowerShell's backtick) and the machine id is a file beside the compose
// file.
type view struct {
	SampleParams
	Image, Data, Label string
	Windows            bool
	Cont               string
}

// check holds every value to its grammar before any reaches a template.
func (p SampleParams) check(kind string) error {
	var errs []error
	keep := func(_ any, err error) {
		if err != nil {
			errs = append(errs, err)
		}
	}
	if origin, err := validate.ServerURL("url", p.URL); err != nil {
		errs = append(errs, err)
	} else if origin != p.URL {
		errs = append(errs, fmt.Errorf("url: must be in canonical form %q (got %q)", origin, p.URL))
	}
	keep(validate.Name("mesh", p.Mesh))
	if p.Name != "" || kind != StackGlobal {
		keep(validate.Name("name", p.Name))
	}
	if p.DataDir != "" {
		keep(validate.POSIXPath("data_dir", p.DataDir))
	}
	if strings.HasPrefix(kind, "network-") || kind == ComposeContainer || kind == StackContainer || kind == ComposeWindows {
		keep(validate.Name("network_name", p.NetworkName))
		keep(validate.Name("bridge_name", p.BridgeName))
		keep(validate.Prefix4("network", p.Network))
		keep(validate.Prefix4("range", p.Range))
		keep(validate.IPv4("gateway", p.Gateway))
		if p.Network6 != "" {
			keep(validate.Prefix6("network6", p.Network6))
			keep(validate.Prefix6("range6", p.Range6))
			keep(validate.IPv6("gateway6", p.Gateway6))
		}
	}
	if p.Server && p.Container {
		keep(validate.Port("listen_port", p.ListenPort))
		keep(validate.Port("publish_port", p.PublishPort))
	}
	return errors.Join(errs...)
}

// RenderSample fills the sample kind with p.
func RenderSample(kind string, p SampleParams) (Sample, error) {
	s, ok := sampleInfo[kind]
	if !ok {
		return Sample{}, fmt.Errorf("render sample: no sample %q", kind)
	}
	if err := p.check(kind); err != nil {
		return Sample{}, fmt.Errorf("render sample %s: %w", kind, err)
	}
	v := view{SampleParams: p, Image: Image + ":<release>", Data: "meshctl-data", Label: "meshctl." + p.Mesh, Cont: `\`}
	if kind == NetworkWindows || kind == ComposeWindows {
		v.Windows, v.Cont = true, "`"
	}
	if _, err := validate.SemVer("version", p.Version); err == nil {
		v.Image = Image + ":" + strings.TrimPrefix(p.Version, "v")
	}
	if p.DataDir != "" {
		v.Data = p.DataDir
	}
	if p.Name != "" {
		v.Label += "." + p.Name
	}
	var b bytes.Buffer
	file := kind
	if f, ok := sampleFile[kind]; ok {
		file = f
	}
	if err := sampleTemplates.ExecuteTemplate(&b, file+".tmpl", v); err != nil {
		return Sample{}, fmt.Errorf("render sample %s: %w", kind, err)
	}
	s.Kind, s.Text = kind, b.String()
	r := strings.NewReplacer("<label>", v.Label, "<stack>", "meshctl-"+p.Mesh+stackSuffix(p.Name), "<mesh>", p.Mesh)
	s.Run, s.Note = r.Replace(s.Run), r.Replace(s.Note)
	return s, nil
}

func stackSuffix(name string) string {
	if name == "" {
		return ""
	}
	return "-" + name
}

// NodeSamples are the samples that fit a node: a server node on its host's
// network runs from install.sh (it needs firewalld on the host), so it gets
// only its containers' network; a client node on the host network gets a
// compose file and a Swarm service; a container-mode node its network, a
// compose file and a Swarm service, then both for Docker Desktop on Windows.
func NodeSamples(p SampleParams) ([]Sample, error) {
	var kinds []string
	switch {
	case p.Container:
		kinds = []string{NetworkDocker, NetworkPodman, NetworkSwarm, ComposeContainer, StackContainer, NetworkWindows, ComposeWindows}
	case !p.Server:
		kinds = []string{ComposeHost, StackHost}
		fallthrough
	default:
		if p.Network != "" {
			kinds = append(kinds, NetworkDocker, NetworkPodman, NetworkSwarm)
		}
	}
	out := make([]Sample, 0, len(kinds))
	for _, k := range kinds {
		s, err := RenderSample(k, p)
		if err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, nil
}

// swarmNote is when a node's agent may be a Swarm service.
const swarmNote = "Only where Swarm does not run over this mesh: after a reboot, a Swarm that needs the mesh " +
	"cannot start the agent that brings the mesh up. The host needs the WireGuard kernel module (Swarm " +
	"cannot map /dev/net/tun), and Docker 20.10 or later. Pin it to its host with the label."

var sampleInfo = map[string]Sample{
	ComposeHost: {Title: "Docker compose, on the host's network", Lang: "yaml",
		Run:  "MESHCTL_TOKEN=<token> docker compose up -d",
		Note: "The whole host joins the mesh. With the WireGuard module on the host (ls /sys/module/wireguard), set MESHCTL_WG_BACKEND: kernel and leave devices out."},
	StackHost: {Title: "Docker Swarm service, on the host's network", Lang: "yaml",
		Run:  "docker node update --label-add <label>=true <swarm node>\nMESHCTL_TOKEN=<token> docker stack deploy -c <stack>.yml <stack>",
		Note: swarmNote},
	ComposeContainer: {Title: "Docker compose, in its own network namespace", Lang: "yaml",
		Run:  "MESHCTL_TOKEN=<token> docker compose up -d",
		Note: "Create the mesh network first. Only the containers on it are on the mesh; attach a service with networks: [the mesh network] and its dns set to the agent."},
	StackContainer: {Title: "Docker Swarm service, in its own network namespace", Lang: "yaml",
		Run:  "docker node update --label-add <label>=true <swarm node>\nMESHCTL_TOKEN=<token> docker stack deploy -c <stack>.yml <stack>",
		Note: swarmNote + " Create the mesh network for Swarm first."},
	StackGlobal: {Title: "Every labelled Swarm node a client node", Lang: "yaml",
		Run:  "docker node update --label-add <label>=true <swarm node>   # each, with its token in /etc/meshctl/<mesh>.token\ndocker stack deploy -c <stack>.yml <stack>",
		Note: swarmNote + " Each node joins with the token in its own /etc/meshctl/<mesh>.token (mode 0600), which can go once it joined."},
	ComposeWindows: {Title: "Docker Desktop on Windows, in its own network namespace", Lang: "yaml",
		Run: "(Get-ItemProperty HKLM:\\SOFTWARE\\Microsoft\\Cryptography).MachineGuid | Set-Content -NoNewline -Encoding ascii machine-id\n" +
			"$env:MESHCTL_TOKEN = '<token>'; docker compose up -d",
		Note: "Docker Desktop with Linux containers (WSL 2): only the containers on the mesh network are on the mesh, " +
			"not Windows itself (install.ps1 puts Windows on it). Create the mesh network first, then run this in " +
			"PowerShell in the compose file's folder: the first line writes this PC's machine id beside it, which pins " +
			"the node to this PC."},
	NetworkDocker:  {Title: "The mesh network, with Docker (once per host)", Lang: "sh"},
	NetworkPodman:  {Title: "The mesh network, with Podman (once per host)", Lang: "sh"},
	NetworkSwarm:   {Title: "The mesh network, for Swarm services (the first command on each node, the second once on a manager)", Lang: "sh"},
	NetworkWindows: {Title: "The mesh network, with Docker Desktop on Windows (PowerShell, once per PC)", Lang: "powershell"},
}
