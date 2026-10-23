//go:build e2e && linux

package e2e

import (
	"net"
	"strings"
	"testing"
	"time"

	mdns "github.com/miekg/dns"
)

// TestDualStack (M8): a mesh with an IPv6 prefix carries IPv6 between its
// nodes -- inside tunnels with IPv4 endpoints, and to an IPv6-only spoke
// that dials a server's IPv6 endpoint, relayed for a dual spoke -- through
// the servers' inet table; a host without IPv6 is planned on IPv4 alone;
// and a spoke whose IPv4 path to a server is broken fails over to the
// server's IPv6 endpoint.
func TestDualStack(t *testing.T) {
	l := newLab(t)
	l.agentArgs = []string{"--status-every", "1s"} // the failover is judged at each heartbeat
	pad := host{"h", "pad", "192.0.2.7", ""}
	for _, h := range []host{laptop, nas, pad} {
		l.addHost(h)
	}
	for space, a := range map[string]string{"a": "2001:db8::1", "b": "2001:db8::2", "c": "2001:db8::3", "d": "2001:db8::4", "h": "2001:db8::7"} {
		l.sh("ip", "-n", ns(space), "addr", "add", a+"/64", "dev", "eth0", "nodad")
	}
	// The NAS's host has no IPv6, as some NAS kernels do not.
	l.sh("ip", "netns", "exec", ns(nas.ns), "sysctl", "-qw", "net.ipv6.conf.all.disable_ipv6=1", "net.ipv6.conf.default.disable_ipv6=1")
	// The laptop's network cannot reach vps-2 over IPv4.
	l.sh("ip", "-n", ns(laptop.ns), "route", "add", "blackhole", "192.0.2.2/32")

	l.startServer()
	if out, err := l.meshctl("", nil, "ipv6", "fd10:77::/64", "--data", l.data("server")); err != nil {
		t.Fatalf("ipv6: %v\n%s", err, out)
	}
	l.startMesh()
	if out, err := l.meshctl("", nil, "node", "set", "vps-2", "--endpoint6", "[2001:db8::2]:51820", "--data", l.data("server")); err != nil {
		t.Fatalf("node set --endpoint6: %v\n%s", err, out)
	}
	l.enroll(laptop, "--role", "spoke", "--hub", "vps-1")
	l.enroll(nas, "--role", "spoke", "--hub", "vps-3", "--always-on")
	l.enroll(pad, "--role", "spoke", "--hub", "vps-2", "--family", "ipv6")

	// IPv6 inside tunnels whose endpoints are IPv4, through the servers'
	// table.
	eventually(t, 60*time.Second, "the laptop to reach vps-1 and vps-3 over IPv6", func() bool {
		return l.ping(laptop.ns, "fd10:77::1") && l.ping(laptop.ns, "fd10:77::3") && l.ping("c", "fd10:77::2:1")
	})
	if out := l.sh("ip", "netns", "exec", ns("a"), "nft", "list", "table", "inet", "meshctl"); !strings.Contains(out, "meta nfproto ipv6") {
		t.Fatalf("vps-1's table has no IPv6 rules:\n%s", out)
	}
	// vps-1's resolver answers on its IPv6 mesh address, with AAAA records.
	eventually(t, 30*time.Second, "vps-1 to answer laptop.mesh's AAAA over IPv6", func() bool {
		ip, err := l.resolveAAAA(laptop.ns, "fd10:77::1", "laptop.mesh.")
		return err == nil && ip == "fd10:77::2:1"
	})

	// The NAS reports a host without IPv6: planned on IPv4 alone.
	eventually(t, 30*time.Second, "the server to plan the NAS without IPv6", func() bool {
		out, err := l.meshctl("", nil, "node", "show", "nas", "--data", l.data("server"))
		return err == nil && strings.Contains(out, "ipv4 (its agent's report)") && strings.Contains(out, "cannot use IPv6")
	})
	eventually(t, 30*time.Second, "the laptop and the NAS to reach each other over IPv4", func() bool {
		return l.ping(laptop.ns, nas.meshIP) && l.ping(nas.ns, laptop.meshIP)
	})
	eventually(t, 30*time.Second, "nas.mesh to lose its AAAA: a host without IPv6 has none", func() bool {
		ip, err := l.resolveAAAA(laptop.ns, "10.77.0.1", "nas.mesh.")
		return err == nil && ip == ""
	})

	// The laptop gives up vps-2's IPv4 endpoint and dials its IPv6 one.
	eventually(t, 90*time.Second, "the laptop to fail over to vps-2's IPv6 endpoint", func() bool {
		st, err := l.status(laptop.ns)
		if err != nil {
			return false
		}
		for _, p := range st.Peers {
			if p.Name == "vps-2" && p.Endpoint == "[2001:db8::2]:51820" && p.LastHandshake != 0 {
				return l.ping(laptop.ns, "fd10:77::2")
			}
		}
		return false
	})

	// The IPv6-only spoke has no IPv4 mesh address, dials vps-2's IPv6
	// endpoint, and reaches the laptop through it.
	eventually(t, 60*time.Second, "the IPv6-only spoke and the laptop to reach each other", func() bool {
		st, err := l.status(pad.ns)
		return err == nil && st.MeshIP == "fd10:77::2:2" && l.ping(pad.ns, "fd10:77::2") &&
			l.ping(pad.ns, "fd10:77::2:1") && l.ping(laptop.ns, "fd10:77::2:2")
	})
}

// resolveAAAA asks server, from inside space, for name's IPv6 address; ""
// when it has none.
func (l *lab) resolveAAAA(space, server, name string) (string, error) {
	var ip string
	err := inNS(space, func() error {
		m := new(mdns.Msg)
		m.SetQuestion(name, mdns.TypeAAAA)
		resp, _, err := (&mdns.Client{Timeout: 2 * time.Second}).Exchange(m, net.JoinHostPort(server, "53"))
		if err != nil {
			return err
		}
		if len(resp.Answer) == 1 {
			if a, ok := resp.Answer[0].(*mdns.AAAA); ok {
				ip = a.AAAA.String()
			}
		}
		return nil
	})
	return ip, err
}
