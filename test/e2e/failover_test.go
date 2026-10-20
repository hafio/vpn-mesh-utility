//go:build e2e && linux

package e2e

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/hafio/vpn-mesh-utility/internal/server/api"
)

// relayFor is the server peer that carries from's traffic to the mesh IP
// to, in from's computed config.
func (l *lab) relayFor(from, to string) string {
	l.t.Helper()
	out, err := l.meshctl("", nil, "node", "show", from, "--data", l.data("server"), "--json")
	var d api.NodeDetail
	if err != nil || json.Unmarshal([]byte(out), &d) != nil || d.Config == nil {
		return ""
	}
	for _, p := range d.Config.Peers {
		for _, a := range p.AllowedIPs {
			if a == to+"/32" {
				return p.Name
			}
		}
	}
	return ""
}

// TestRelayFailover is amendment A40 in the lab, with every agent reporting
// each second and a hub judged down after 3 s. The laptop (hub vps-2) and
// the NAS (hub vps-3) relay through vps-2, the lower id's hub. vps-2's agent
// stops and its wg0 goes: the server judges it down and the pair moves to
// vps-3, still talking. vps-2's agent comes back and, steady for 7.5 s, the
// pair returns to it; once the spokes have new sessions with vps-2, the
// pair talks through it again.
func TestRelayFailover(t *testing.T) {
	l := newLab(t)
	l.serverArgs = []string{"--hub-down-after", "3s"}
	l.agentArgs = []string{"--status-every", "1s"}
	l.addHost(laptop)
	l.addHost(nas)
	l.startServer()
	l.startMesh()
	l.enroll(laptop, "--role", "spoke", "--hub", "vps-2")
	l.enroll(nas, "--role", "spoke", "--hub", "vps-3", "--always-on")
	eventually(t, 60*time.Second, "the spokes to reach each other through vps-2", func() bool {
		return l.relayFor("nas", laptop.meshIP) == "vps-2" && l.ping(laptop.ns, nas.meshIP) && l.ping(nas.ns, laptop.meshIP)
	})

	l.stop("agent-b")
	l.sh("ip", "-n", ns("b"), "link", "del", "wg0")
	eventually(t, 30*time.Second, "the pair to move to vps-3 and talk", func() bool {
		return l.relayFor("nas", laptop.meshIP) == "vps-3" && l.relayFor("laptop", nas.meshIP) == "vps-3" &&
			l.ping(laptop.ns, nas.meshIP) && l.ping(nas.ns, laptop.meshIP)
	})
	out, err := l.meshctl("", nil, "node", "list", "--data", l.data("server"), "--json")
	var list api.NodeList
	if err != nil || json.Unmarshal([]byte(out), &list) != nil {
		t.Fatalf("node list: %v\n%s", err, out)
	}
	for _, n := range list.Nodes {
		if (n.Name == "vps-2") != (n.DownSince > 0) {
			t.Errorf("%s down since %d", n.Name, n.DownSince)
		}
	}

	l.startAgent("b")
	// vps-2's new wg0 has no sessions, and a spoke that only sends
	// keepalives keeps its dead one until it expires (two to three
	// minutes); data that goes unanswered for 15 s makes it handshake again,
	// so both spokes talk to vps-2 itself first. (The restore waits 2.5
	// times --hub-down-after: 5 minutes by default, past that expiry.)
	vps2 := nodes[1].meshIP
	eventually(t, 90*time.Second, "both spokes to reach vps-2 again", func() bool {
		a, b := l.ping(laptop.ns, vps2), l.ping(nas.ns, vps2)
		return a && b
	})
	eventually(t, 60*time.Second, "the pair back on vps-2 and talking", func() bool {
		return l.relayFor("nas", laptop.meshIP) == "vps-2" && l.ping(laptop.ns, nas.meshIP) && l.ping(nas.ns, laptop.meshIP)
	})
}
