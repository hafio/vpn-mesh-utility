//go:build e2e && linux

package e2e

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"

	"github.com/hafio/vpn-mesh-utility/internal/server/api"
)

// joinIn adds n to mesh through the CLI, then joins and starts its agent
// in n's namespace with the data dir data: a host in a second mesh runs a
// second agent (amendment A49). Every node here is a server.
func (l *lab) joinIn(mesh, data string, n host, args ...string) {
	l.t.Helper()
	add := append([]string{"node", "add", n.name, "--mesh", mesh}, args...)
	out, err := l.meshctl("", nil, append(add, "--data", l.data("server"), "--json")...)
	if err != nil {
		l.t.Fatalf("node add %s --mesh %s: %v\n%s", n.name, mesh, err, out)
	}
	var ch api.NodeChange
	if err := json.Unmarshal([]byte(out), &ch); err != nil || ch.Enrollment == nil || ch.Node.MeshIP != n.meshIP {
		l.t.Fatalf("node add %s --mesh %s: %v\n%s", n.name, mesh, err, out)
	}
	env := []string{"SSL_CERT_FILE=" + l.ca, "MESHCTL_TOKEN=" + ch.Enrollment.Token}
	join := []string{"agent", "join", "--server", "https://192.0.2.1:8443", "--data", l.data(data), "--no-service",
		"--endpoint", fmt.Sprintf("%s:%d", n.ip, ch.Node.ListenPort)}
	if out, err := l.meshctl(n.ns, env, join...); err != nil {
		l.t.Fatalf("agent join %s in %s: %v\n%s", n.name, mesh, err, out)
	}
	l.startAgentAt(n.ns, data)
}

// Two meshes on one control plane, and one host in both (amendment A49):
// vps-1's namespace runs the default mesh's agent on wg0 and the lab's on
// wg1 side by side -- their own ports, nftables tables, resolvers and local
// UIs. Each mesh's nodes reach theirs and none of the other's, even when a
// node of one routes the other's network into its tunnel: the other
// mesh's peers take no packet from a source that is not their peer's.
func TestTwoMeshes(t *testing.T) {
	l := newLab(t)
	l.startServer()
	vps1, vps2 := nodes[0], nodes[1]
	lab1 := host{vps1.ns, "vps-1", vps1.ip, "10.78.0.1"}
	lab3 := host{nodes[2].ns, "vps-3", nodes[2].ip, "10.78.0.2"}
	// The default mesh first: with two, every node command names its mesh.
	l.enroll(vps1, "--role", "server", "--endpoint", vps1.ip+":51820")
	l.enroll(vps2, "--role", "server", "--endpoint", vps2.ip+":51820")
	if out, err := l.meshctl("", nil, "mesh", "add", "lab", "--data", l.data("server")); err != nil {
		t.Fatalf("mesh add lab: %v\n%s", err, out)
	}
	l.joinIn("lab", "a-lab", lab1, "--role", "server")
	l.joinIn("lab", lab3.ns, lab3, "--role", "server")
	eventually(t, 60*time.Second, "both meshes to pass traffic, vps-1 in both", func() bool {
		for _, d := range []string{"a", "b", "a-lab", "c"} {
			st, err := l.status(d)
			if err != nil || st.AppliedVersion == 0 || len(st.Peers) != 1 {
				return false
			}
		}
		return l.ping("b", vps1.meshIP) && l.ping("a", vps2.meshIP) && l.ping("c", lab1.meshIP) && l.ping("a", lab3.meshIP)
	})

	// vps-1's two agents, side by side.
	if err := inNS("a", func() error {
		c, err := wgctrl.New()
		if err != nil {
			return err
		}
		defer c.Close()
		for name, port := range map[string]int{"wg0": 51820, "wg1": 51821} {
			d, err := c.Device(name)
			if err != nil {
				return err
			}
			if d.ListenPort != port {
				return fmt.Errorf("%s listens on %d, want %d", name, d.ListenPort, port)
			}
		}
		return nil
	}); err != nil {
		t.Fatalf("vps-1's interfaces: %v", err)
	}
	if l.has("c", "link", "show", "wg0") {
		t.Error("the lab's vps-3 made wg0")
	}
	if tables := l.sh("ip", "netns", "exec", ns("a"), "nft", "list", "tables"); !strings.Contains(tables, "table inet meshctl\n") ||
		!strings.Contains(tables, "table inet meshctl1\n") {
		t.Fatalf("vps-1's nftables tables:\n%s", tables)
	}
	for _, c := range []struct{ from, at, name, want string }{
		{"b", vps1.meshIP, "vps-2.mesh.", vps2.meshIP},
		{"c", lab1.meshIP, "vps-3.lab.", lab3.meshIP},
	} {
		if ip, _, err := l.resolve(c.from, c.at, c.name); err != nil || ip != c.want {
			t.Errorf("%s at %s = %s, %v; want %s", c.name, c.at, ip, err, c.want)
		}
	}
	for data, port := range map[string]string{"a": "7443", "a-lab": "7444"} {
		out, err := l.meshctl("", nil, "ui", "--socket", filepath.Join(l.data(data), "agent.sock"))
		if err != nil || !strings.Contains(out, "http://127.0.0.1:"+port+"/login?t=") {
			t.Errorf("the local UI of %s: %v\n%s", data, err, out)
		}
	}

	// vps-2 sends the lab's network into its tunnel to vps-1, which
	// forwards: the lab's vps-3 takes nothing from vps-1's lab side with a
	// default-mesh source.
	l.sh("ip", "netns", "exec", ns("a"), "sysctl", "-qw", "net.ipv4.ip_forward=1")
	_, labNet, _ := net.ParseCIDR("10.78.0.0/16")
	if err := inNS("b", func() error {
		c, err := wgctrl.New()
		if err != nil {
			return err
		}
		defer c.Close()
		d, err := c.Device("wg0")
		if err != nil {
			return err
		}
		if len(d.Peers) != 1 {
			return errors.New("vps-2 has no single peer")
		}
		return c.ConfigureDevice("wg0", wgtypes.Config{Peers: []wgtypes.PeerConfig{
			{PublicKey: d.Peers[0].PublicKey, UpdateOnly: true, AllowedIPs: []net.IPNet{*labNet}}}})
	}); err != nil {
		t.Fatalf("route the lab into vps-2's tunnel: %v", err)
	}
	l.sh("ip", "-n", ns("b"), "route", "add", labNet.String(), "dev", "wg0")
	if l.ping("b", lab3.meshIP) {
		t.Fatal("a node of the default mesh reached a node of the lab")
	}
}
