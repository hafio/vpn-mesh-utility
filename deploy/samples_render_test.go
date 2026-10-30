package deploy

import (
	"errors"
	"io"
	"os"
	"regexp"
	"slices"
	"strings"
	"testing"

	"go.yaml.in/yaml/v3"
)

// docSamples are the example nodes docs/containers.md fills its samples
// with: a NAS on its host's network, a container-mode client, a host-mode
// server node with containers, and the global Swarm service.
var docSamples = map[string]SampleParams{
	"nas": {URL: "https://mesh.example.com", Version: "v1.0.0", Mesh: "default", Name: "nas", DataDir: "/volume1/docker/meshctl"},
	"box": {URL: "https://mesh.example.com", Version: "v1.0.0", Mesh: "default", Name: "box", Container: true,
		NetworkName: "meshctl-mesh", BridgeName: "meshctl0", Network: "10.77.0.0/16", Range: "10.77.128.128/25",
		Gateway: "10.77.128.1", Agent: "10.77.128.2"},
	"vps": {URL: "https://mesh.example.com", Version: "v1.0.0", Mesh: "default", Name: "vps-1", Server: true,
		NetworkName: "meshctl-mesh", BridgeName: "meshctl0", Network: "10.77.0.0/16", Range: "10.77.129.128/25",
		Gateway: "10.77.129.1"},
	"global": {URL: "https://mesh.example.com", Version: "v1.0.0", Mesh: "default"},
}

var reSampleMarker = regexp.MustCompile(`^<!-- sample: ([a-z-]+) ([a-z0-9-]+) -->$`)

// docWithSamples is docs/containers.md with each block after a sample marker
// rendered from its template, and the kinds it shows.
func docWithSamples(t *testing.T, doc string) (string, map[string]bool) {
	t.Helper()
	lines := strings.Split(doc, "\n")
	shown := map[string]bool{}
	var out []string
	for i := 0; i < len(lines); i++ {
		out = append(out, lines[i])
		m := reSampleMarker.FindStringSubmatch(lines[i])
		if m == nil {
			continue
		}
		end := i + 2
		for end < len(lines) && lines[end] != "```" {
			end++
		}
		if i+1 >= len(lines) || !strings.HasPrefix(lines[i+1], "```") || end >= len(lines) {
			t.Fatalf("line %d: a sample marker must stand right before a fenced block", i+1)
		}
		p, ok := docSamples[m[2]]
		if !ok {
			t.Fatalf("line %d: no example node %q", i+1, m[2])
		}
		s, err := RenderSample(m[1], p)
		if err != nil {
			t.Fatalf("line %d: %v", i+1, err)
		}
		out = append(out, "```"+s.Lang)
		out = append(out, strings.Split(strings.TrimSuffix(s.Text, "\n"), "\n")...)
		out = append(out, "```")
		shown[m[1]] = true
		i = end
	}
	return strings.Join(out, "\n"), shown
}

// docs/containers.md shows every sample kind as its template renders it
// (amendment A53); regen rewrites a stale one.
func TestContainersDocSamples(t *testing.T) {
	b, err := os.ReadFile(containersDoc)
	if err != nil {
		t.Fatal(err)
	}
	got, shown := docWithSamples(t, string(b))
	if *update && got != string(b) {
		if err := os.WriteFile(containersDoc, []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
	} else if got != string(b) {
		t.Error("docs/containers.md's samples differ from deploy/samples; run ./scripts/dev.sh regen")
	}
	for kind := range sampleInfo {
		if !shown[kind] {
			t.Errorf("docs/containers.md shows no %s sample", kind)
		}
	}
}

// parseSample decodes a rendered yaml sample and holds it to the posture
// checks of docs/containers.md's samples.
func parseSample(t *testing.T, s Sample) {
	t.Helper()
	if s.Lang != "yaml" {
		return
	}
	dec := yaml.NewDecoder(strings.NewReader(s.Text))
	for {
		var doc map[string]any
		err := dec.Decode(&doc)
		if errors.Is(err, io.EOF) {
			return
		}
		if err != nil {
			t.Fatalf("%s does not parse: %v\n%s", s.Kind, err, s.Text)
		}
		if checkComposeSample(t, doc)+checkStackSample(t, doc) != 1 {
			t.Errorf("%s holds no service of meshctl's image:\n%s", s.Kind, s.Text)
		}
	}
}

// A node gets the samples that fit it: a server node on its host's network
// only its containers' network (it runs from install.sh), a client node on
// its host's network compose and Swarm, a container-mode node its network,
// compose and Swarm; a server node's port and the containers' IPv6 appear
// where they apply, and every file keeps its posture.
func TestNodeSamples(t *testing.T) {
	base := docSamples["box"]
	server := base
	server.Server, server.ListenPort, server.PublishPort = true, 51821, 51830
	dual := base
	dual.Network6, dual.Range6, dual.Gateway6 = "fd10:77::/64", "fd10:77::8000:8000/113", "fd10:77::8000:1"
	hostSpoke := docSamples["nas"]
	hostSpokeContainers := hostSpoke
	hostSpokeContainers.NetworkName, hostSpokeContainers.BridgeName = "meshctl-mesh1", "meshctl1"
	hostSpokeContainers.Network, hostSpokeContainers.Range, hostSpokeContainers.Gateway = "10.78.0.0/16", "10.78.130.128/25", "10.78.130.1"
	hostServer := docSamples["vps"]
	hostServer.NetworkName = ""
	hostServer.Network = ""
	container := []string{NetworkDocker, NetworkPodman, NetworkSwarm, ComposeContainer, StackContainer, NetworkWindows, ComposeWindows}
	for _, c := range []struct {
		name  string
		p     SampleParams
		kinds []string
		want  []string
	}{
		{"host spoke", hostSpoke, []string{ComposeHost, StackHost}, []string{"network_mode: host", "name: host", "node.labels.meshctl.default.nas"}},
		{"host spoke with containers", hostSpokeContainers, []string{ComposeHost, StackHost, NetworkDocker, NetworkPodman, NetworkSwarm},
			[]string{"inhibit_ipv4=true", "mode=unmanaged", "--config-from meshctl-mesh1-config meshctl-mesh1"}},
		{"host server", hostServer, nil, nil},
		{"host server with containers", docSamples["vps"], []string{NetworkDocker, NetworkPodman, NetworkSwarm}, []string{"nat-unprotected"}},
		{"container spoke", base, container, []string{"--internal", "node.labels.meshctl.default.box", "./machine-id:/etc/machine-id:ro"}},
		{"container server", server, container,
			[]string{`ports: ["51830:51821/udp"]`, "target: 51821", "published: 51830", "mode: host"}},
		{"container dual stack", dual, container, []string{"--ipv6 --subnet fd10:77::/64", "net.ipv6.conf.all.proxy_ndp"}},
	} {
		got, err := NodeSamples(c.p)
		if err != nil {
			t.Fatalf("%s: %v", c.name, err)
		}
		var kinds []string
		all := ""
		for _, s := range got {
			kinds = append(kinds, s.Kind)
			all += s.Text + s.Run + "\n"
			parseSample(t, s)
		}
		if !slices.Equal(kinds, c.kinds) {
			t.Errorf("%s: samples %v, want %v", c.name, kinds, c.kinds)
		}
		for _, w := range c.want {
			if !strings.Contains(all, w) {
				t.Errorf("%s: no %q in\n%s", c.name, w, all)
			}
		}
		if strings.Contains(all, "proxy_ndp") != (c.p.Network6 != "") {
			t.Errorf("%s: the IPv6 sysctls follow the containers' IPv6", c.name)
		}
	}
	g, err := RenderSample(StackGlobal, docSamples["global"])
	if err != nil {
		t.Fatal(err)
	}
	parseSample(t, g)
	if !strings.Contains(g.Text, "mode: global") || !strings.Contains(g.Run, "meshctl.default=true") {
		t.Errorf("the global service:\n%s\n%s", g.Text, g.Run)
	}
}

// Docker Desktop on Windows gets PowerShell: the network command continues
// its lines with a backtick, and the compose file mounts the PC's machine
// id from a file beside it, which the first command writes.
func TestWindowsSamples(t *testing.T) {
	for kind, cont := range map[string]string{NetworkDocker: " \\", NetworkWindows: " `"} {
		s, err := RenderSample(kind, docSamples["box"])
		if err != nil {
			t.Fatal(err)
		}
		lines := strings.Split(strings.TrimSuffix(s.Text, "\n"), "\n")
		for i, l := range lines {
			if (i < len(lines)-1) != strings.HasSuffix(l, cont) || strings.HasSuffix(l, " \\") && cont != " \\" {
				t.Errorf("%s line %d: %q (lines but the last end in %q)", kind, i+1, l, cont)
			}
		}
	}
	s, err := RenderSample(ComposeWindows, docSamples["box"])
	if err != nil {
		t.Fatal(err)
	}
	parseSample(t, s)
	if !strings.Contains(s.Text, "- ./machine-id:/etc/machine-id:ro\n") || strings.Contains(s.Text, "- /etc/machine-id:") ||
		!strings.Contains(s.Run, "MachineGuid | Set-Content -NoNewline -Encoding ascii machine-id\n") ||
		!strings.HasSuffix(s.Run, "$env:MESHCTL_TOKEN = '<token>'; docker compose up -d") {
		t.Errorf("compose-windows:\n%s\n%s", s.Text, s.Run)
	}
}

// The image is the control plane's release; a build that is none says so.
func TestSampleImage(t *testing.T) {
	for version, want := range map[string]string{"v1.0.0-rc.6": Image + ":1.0.0-rc.6", "1.2.3": Image + ":1.2.3", "dev": Image + ":<release>"} {
		p := docSamples["nas"]
		p.Version = version
		s, err := RenderSample(ComposeHost, p)
		if err != nil || !strings.Contains(s.Text, "image: "+want+"\n") {
			t.Errorf("version %q: %v\n%s", version, err, s.Text)
		}
	}
	if _, err := RenderSample("compose-k8s", docSamples["nas"]); err == nil {
		t.Error("an unknown sample rendered")
	}
}

// A value that is not its grammar never reaches a sample: the render fails
// and returns no text.
func TestSamplesRefuseHostileValues(t *testing.T) {
	for _, c := range []struct {
		field string
		kind  string
		set   func(*SampleParams)
	}{
		{"url", ComposeHost, func(p *SampleParams) { p.URL = "https://mesh.example.com/\nprivileged: true" }},
		{"url", ComposeHost, func(p *SampleParams) { p.URL = "https://Mesh.Example.com" }},
		{"name", ComposeHost, func(p *SampleParams) { p.Name = "nas\n    privileged: true" }},
		{"name", StackHost, func(p *SampleParams) { p.Name = "" }},
		{"mesh", StackGlobal, func(p *SampleParams) { p.Mesh = "default == true]\n" }},
		{"data_dir", ComposeHost, func(p *SampleParams) { p.DataDir = `/data", "/:/host` }},
		{"network", ComposeContainer, func(p *SampleParams) { p.Network = "10.77.0.0/16 --privileged" }},
		{"bridge_name", NetworkDocker, func(p *SampleParams) { p.BridgeName = "meshctl0; reboot" }},
		{"network_name", NetworkWindows, func(p *SampleParams) { p.NetworkName = "meshctl-mesh`; rm" }},
		{"range", ComposeWindows, func(p *SampleParams) { p.Range = "$(whoami)" }},
		{"gateway6", NetworkPodman, func(p *SampleParams) {
			p.Network6, p.Range6, p.Gateway6 = "fd10:77::/64", "fd10:77::8000:8000/113", "x"
		}},
		{"publish_port", StackContainer, func(p *SampleParams) { p.Server, p.ListenPort, p.PublishPort = true, 51820, 70000 }},
	} {
		p := docSamples["box"]
		if c.kind == ComposeHost || c.kind == StackHost {
			p = docSamples["nas"]
		}
		if c.kind == StackGlobal {
			p = docSamples["global"]
		}
		c.set(&p)
		s, err := RenderSample(c.kind, p)
		if err == nil || !strings.Contains(err.Error(), c.field) || s.Text != "" {
			t.Errorf("%s in %s: %v\n%s", c.field, c.kind, err, s.Text)
		}
	}
}
