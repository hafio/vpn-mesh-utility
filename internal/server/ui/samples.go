package ui

import (
	"cmp"
	"net/http"
	"strings"

	"github.com/hafio/vpn-mesh-utility/deploy"
	"github.com/hafio/vpn-mesh-utility/internal/proto"
	"github.com/hafio/vpn-mesh-utility/internal/server/api"
)

// hostServerNote says why a server node on its host's network has no
// container samples (amendment A53).
const hostServerNote = "A server node on its host's network runs from install.sh (above): it needs firewalld on the host " +
	"for its public port, the relay and the exit, which the image does not carry, so it has no compose file. To run it in " +
	"Docker, remove it and add it again in container mode (a node's mode is chosen when it is added): only its containers " +
	"are then on the mesh, and this page shows its compose files."

// nodeSamples are the compose, Swarm and network samples that fit node n,
// filled with its values and token (amendment A53), and a note when a kind
// of node runs from install.sh only. A phone has none.
func (u *ui) nodeSamples(r *http.Request, n api.NodeInfo, token string) ([]deploy.Sample, string) {
	if n.Role == string(proto.RolePhone) {
		return nil, ""
	}
	p := deploy.SampleParams{URL: u.ServerURL, Version: u.Version, Mesh: u.mesh.Name, Name: n.Name,
		Server: n.Role == string(proto.RoleServer), Container: n.Mode == proto.ModeContainer}
	note := ""
	if p.Server && !p.Container {
		note = hostServerNote
	}
	d, err := u.Nodes.Show(r.Context(), n.Name)
	if err != nil {
		u.Log.Error("read the node for its samples", "node", n.Name, "err", err)
		return nil, note
	}
	if k := d.Containers; k != nil {
		p.NetworkName, p.BridgeName = k.Name, k.BridgeName
		p.Network, p.Range, p.Gateway, p.Agent = k.Network, k.Range, k.Bridge, k.Gateway
		p.Network6, p.Range6, p.Gateway6 = k.Network6, k.Range6, k.Bridge6
		p.ListenPort, p.PublishPort = k.ListenPort, cmp.Or(k.PublishPort, k.ListenPort)
	}
	samples, err := deploy.NodeSamples(p)
	if err != nil {
		u.Log.Error("render the node's samples", "node", n.Name, "err", err)
		return nil, note
	}
	for i := range samples {
		samples[i].Run = strings.ReplaceAll(samples[i].Run, "<token>", token)
	}
	return samples, note
}
