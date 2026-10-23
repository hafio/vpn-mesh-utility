//go:build e2e && linux

package e2e

import (
	"encoding/json"
	"net"
	"net/http"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// port plugs namespace space into the bridge br of namespace brNS with the
// interface ifname holding addr, as a runtime attaches a container.
func (l *lab) port(brNS, br, space, ifname, addr string) {
	l.t.Helper()
	peer := "p-" + space + ifname
	l.sh("ip", "-n", ns(space), "link", "add", ifname, "type", "veth", "peer", "name", peer, "netns", ns(brNS))
	l.sh("ip", "-n", ns(brNS), "link", "set", peer, "master", br, "up")
	l.sh("ip", "-n", ns(space), "addr", "add", addr, "dev", ifname)
	l.sh("ip", "-n", ns(space), "link", "set", ifname, "up")
	l.sh("ip", "-n", ns(space), "link", "set", "lo", "up")
}

func (l *lab) has(space string, args ...string) bool {
	out, err := exec.Command("ip", append([]string{"-n", ns(space)}, args...)...).CombinedOutput()
	return err == nil && len(out) > 0
}

func (l *lab) out(space string, args ...string) string {
	out, _ := exec.Command("ip", append([]string{"-n", ns(space)}, args...)...).CombinedOutput()
	return string(out)
}

// fakeRuntime serves, for the agent in space, a Docker-compatible API on a
// unix socket with one container, the compose service service, on the mesh
// network at ip and ip6 (amendment A38; M8), and returns the socket's path.
// Its event stream stays open and silent.
func (l *lab) fakeRuntime(space, service, ip, ip6 string) string {
	l.t.Helper()
	const id = "5eb5eb5eb5eb0000000000000000000000000000000000000000000000000000"
	path := filepath.Join(l.dir, "runtime-"+space+".sock")
	ln, err := net.Listen("unix", path)
	if err != nil {
		l.t.Fatal(err)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /containers/json", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode([]map[string]string{{"Id": id}})
	})
	mux.HandleFunc("GET /containers/"+id+"/json", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"Id": id, "Name": "/shop-" + service + "-1",
			"Config": map[string]any{"Hostname": id[:12], "Labels": map[string]string{"com.docker.compose.service": service}},
			"NetworkSettings": map[string]any{"Networks": map[string]any{
				"meshctl-mesh": map[string]any{"IPAddress": ip, "GlobalIPv6Address": ip6}}}})
	})
	mux.HandleFunc("GET /events", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.(http.Flusher).Flush()
		<-r.Context().Done()
	})
	srv := &http.Server{Handler: mux}
	go func() { _ = srv.Serve(ln) }()
	l.t.Cleanup(func() { _ = srv.Close() })
	return path
}

// TestContainerNetworks is amendment A37 in the lab. box is in container
// mode: its agent runs in namespace h (the agent's container), whose eth1
// is a port of the runtime's bridge dk0 in namespace i (the host, which
// holds .1), with an attached container in j. pc is in host mode with its
// containers on the mesh: its agent owns the bridge meshctl0 in k, where
// the container in l is attached. A LAN host in m routes to pc's
// containers through k, which the guard must refuse. box's runtime calls
// its container web, which every node resolves (amendment A38). Both
// nodes' containers have IPv6 on the mesh too (M8); pc's runtime calls its
// container db, and the two reach each other over IPv6 by their names'
// addresses.
func TestContainerNetworks(t *testing.T) {
	l := newLab(t)
	l.startServer()
	if out, err := l.meshctl("", nil, "ipv6", "fd10:77::/64", "--data", l.data("server")); err != nil {
		t.Fatalf("ipv6: %v\n%s", err, out)
	}
	l.startMesh()
	box := host{"h", "box", "192.0.2.7", "10.77.2.1"}
	pc := host{"k", "pc", "192.0.2.8", "10.77.2.2"}
	l.addHost(box)
	l.addHost(pc)

	l.sh("ip", "netns", "add", ns("i"))
	l.sh("ip", "-n", ns("i"), "link", "add", "dk0", "type", "bridge")
	l.sh("ip", "-n", ns("i"), "addr", "add", "10.77.128.1/16", "dev", "dk0")
	l.sh("ip", "-n", ns("i"), "addr", "add", "fd10:77::8000:1/64", "dev", "dk0", "nodad")
	l.sh("ip", "-n", ns("i"), "link", "set", "dk0", "up")
	l.port("i", "dk0", "h", "eth1", "10.77.128.128/16") // the runtime's dynamic addresses
	l.sh("ip", "-n", ns("h"), "addr", "add", "fd10:77::8000:8000/64", "dev", "eth1", "nodad")
	l.sh("ip", "netns", "add", ns("j"))
	l.port("i", "dk0", "j", "eth0", "10.77.128.130/16")
	l.sh("ip", "-n", ns("j"), "addr", "add", "fd10:77::8000:8002/64", "dev", "eth0", "nodad")
	// What the compose file's sysctls do in the agent's namespace.
	l.sh("ip", "netns", "exec", ns("h"), "sysctl", "-qw", "net.ipv4.ip_forward=1", "net.ipv4.conf.all.proxy_arp=1",
		"net.ipv6.conf.all.forwarding=1", "net.ipv6.conf.all.proxy_ndp=1")

	l.runtimes["h"] = l.fakeRuntime("h", "web", "10.77.128.130", "fd10:77::8000:8002")
	l.runtimes["k"] = l.fakeRuntime("k", "db", "10.77.129.130", "fd10:77::8001:8002")
	l.enroll(box, "--role", "spoke", "--mode", "container", "--containers6", "--hub", "vps-1")
	l.enroll(pc, "--role", "spoke", "--containers", "--containers6", "--hub", "vps-2")

	// Container mode: the gateway claims .2 and installs its rules.
	eventually(t, 60*time.Second, "box's gateway to claim 10.77.128.2 and its rules", func() bool {
		return strings.Contains(l.out("h", "-4", "addr", "show", "dev", "eth1"), "10.77.128.2/16") &&
			strings.Contains(l.out("h", "rule", "show"), "1006:")
	})
	eventually(t, 60*time.Second, "box's container to reach vps-1 through the gateway", func() bool { return l.ping("j", "10.77.0.1") })
	// A container's first packet to a new mesh address waits for the
	// gateway's proxy ARP, which the kernel delays by up to proxy_delay
	// (0.8 s): longer than one ping's wait, so these retry.
	eventually(t, 10*time.Second, "vps-2 to reach box's container", func() bool { return l.ping("b", "10.77.128.130") })
	if ip, _, err := l.resolve("j", "10.77.128.2", "vps-1.mesh."); err != nil || ip != "10.77.0.1" {
		t.Errorf("box's container resolves vps-1.mesh through the gateway = %q, %v", ip, err)
	}
	if l.ping("i", "10.77.0.1") {
		t.Error("box's host reaches the mesh through the gateway")
	}
	if l.ping("a", "10.77.128.1") {
		t.Error("the mesh reaches box's host through the gateway")
	}

	// The containers' IPv6: the gateway claims ::2 and its IPv6 rules, and
	// answers neighbour discovery for the mesh addresses it routes (the
	// first packet to each waits for the proxy delay, as ARP does).
	eventually(t, 30*time.Second, "box's gateway to claim fd10:77::8000:2, its IPv6 rules and proxy entries", func() bool {
		return strings.Contains(l.out("h", "-6", "addr", "show", "dev", "eth1"), "fd10:77::8000:2/64") &&
			strings.Contains(l.out("h", "-6", "rule", "show"), "1006:") &&
			strings.Contains(l.out("h", "-6", "neigh", "show", "proxy"), "fd10:77::1 ")
	})
	eventually(t, 10*time.Second, "box's container to reach vps-1 over IPv6", func() bool { return l.ping("j", "fd10:77::1") })
	eventually(t, 10*time.Second, "vps-2 to reach box's container over IPv6", func() bool { return l.ping("b", "fd10:77::8000:8002") })
	if ip, err := l.resolveAAAA("j", "10.77.128.2", "vps-1.mesh."); err != nil || ip != "fd10:77::1" {
		t.Errorf("box's container resolves vps-1.mesh's AAAA through the gateway = %q, %v", ip, err)
	}
	if l.ping("i", "fd10:77::1") {
		t.Error("box's host reaches the mesh over IPv6 through the gateway")
	}
	if l.ping("a", "fd10:77::8000:1") {
		t.Error("the mesh reaches box's host over IPv6 through the gateway")
	}

	// Container names: box reports web, the server publishes it, every
	// resolver answers it -- bare and with the node.
	for _, c := range []struct{ from, at, name string }{
		{"b", "10.77.0.2", "web.mesh."}, {"b", "10.77.0.2", "web.box.mesh."}, {"j", "10.77.128.2", "web.mesh."},
	} {
		eventually(t, 30*time.Second, c.name+" from "+c.from, func() bool {
			ip, _, err := l.resolve(c.from, c.at, c.name)
			return err == nil && ip == "10.77.128.130"
		})
	}

	// Host mode: the agent makes meshctl0 with .1; the container attaches.
	eventually(t, 60*time.Second, "pc's agent to make meshctl0 with 10.77.129.1 and fd10:77::8001:1", func() bool {
		return strings.Contains(l.out("k", "-4", "addr", "show", "dev", "meshctl0"), "10.77.129.1/16") &&
			strings.Contains(l.out("k", "-6", "addr", "show", "dev", "meshctl0"), "fd10:77::8001:1/64")
	})
	l.sh("ip", "netns", "add", ns("l"))
	l.port("k", "meshctl0", "l", "eth0", "10.77.129.130/16")
	l.sh("ip", "-n", ns("l"), "addr", "add", "fd10:77::8001:8002/64", "dev", "eth0", "nodad")
	l.sh("ip", "-n", ns("l"), "-6", "route", "add", "default", "via", "fd10:77::8001:1") // as Docker makes it
	eventually(t, 30*time.Second, "vps-1 to reach pc's container", func() bool { return l.ping("a", "10.77.129.130") })
	eventually(t, 10*time.Second, "box's container to reach pc's (relayed by vps-1)", func() bool { return l.ping("j", "10.77.129.130") })
	if !l.ping("k", "10.77.129.130") {
		t.Error("pc's host cannot reach its own container")
	}
	eventually(t, 10*time.Second, "vps-1 to reach pc's container over IPv6", func() bool { return l.ping("a", "fd10:77::8001:8002") })
	// Containers on two nodes over IPv6: each side answers neighbour
	// discovery for the other's container, whose address the mesh view's
	// names carry (the AAAA records are the same names).
	eventually(t, 30*time.Second, "web.mesh's AAAA at vps-2", func() bool {
		ip, err := l.resolveAAAA("b", "10.77.0.2", "web.mesh.")
		return err == nil && ip == "fd10:77::8000:8002"
	})
	eventually(t, 30*time.Second, "box's and pc's containers to reach each other over IPv6", func() bool {
		return l.ping("j", "fd10:77::8001:8002") && l.ping("l", "fd10:77::8000:8002")
	})

	// pc's host has IPv6 to the internet, its containers do not: one whose
	// IPv6 default route is the bridge is refused at once, so an app falls
	// back to IPv4 instead of waiting.
	l.sh("ip", "-n", ns("k"), "addr", "add", "2001:db8::8/64", "dev", "eth0", "nodad")
	l.sh("ip", "-n", ns("k"), "-6", "route", "add", "default", "via", "2001:db8::ff")
	start := time.Now()
	out, _ := exec.Command("ip", "netns", "exec", ns("l"), "ping", "-c", "1", "-W", "3", "2001:db8::99").CombinedOutput()
	if d := time.Since(start); d > 2*time.Second || !strings.Contains(strings.ToLower(string(out)), "unreachable") {
		t.Errorf("an IPv6 internet address from pc's container took %s: %s; want an unreachable at once", d, out)
	}

	// The guard: a LAN host that routes the subnet to pc gets nowhere.
	l.sh("ip", "netns", "add", ns("m"))
	l.sh("ip", "-n", ns("m"), "link", "add", "eth0", "type", "veth", "peer", "name", "p-mlan", "netns", ns("k"))
	l.sh("ip", "-n", ns("k"), "addr", "add", "10.197.0.1/30", "dev", "p-mlan")
	l.sh("ip", "-n", ns("k"), "link", "set", "p-mlan", "up")
	l.sh("ip", "-n", ns("m"), "addr", "add", "10.197.0.2/30", "dev", "eth0")
	l.sh("ip", "-n", ns("m"), "link", "set", "eth0", "up")
	l.sh("ip", "-n", ns("m"), "route", "add", "10.77.0.0/16", "via", "10.197.0.1")
	l.sh("ip", "-n", ns("l"), "route", "add", "10.197.0.0/30", "via", "10.77.129.1")
	if l.ping("m", "10.77.129.130") {
		t.Error("a LAN host reaches pc's container: the guard did not hold")
	}
	l.sh("ip", "-n", ns("k"), "addr", "add", "fd97::1/64", "dev", "p-mlan", "nodad")
	l.sh("ip", "-n", ns("m"), "addr", "add", "fd97::2/64", "dev", "eth0", "nodad")
	l.sh("ip", "-n", ns("m"), "-6", "route", "add", "fd10:77::/64", "via", "fd97::1")
	if l.ping("m", "fd10:77::8001:8002") {
		t.Error("a LAN host reaches pc's container over IPv6: the guard did not hold")
	}

	// Containers off: the agent takes its rules away.
	if out, err := l.meshctl("", nil, "node", "set", "pc", "--containers=false", "--data", l.data("server")); err != nil {
		t.Fatalf("node set pc --containers=false: %v\n%s", err, out)
	}
	eventually(t, 60*time.Second, "pc's guard to go", func() bool {
		return !strings.Contains(l.out("k", "rule", "show"), "5203:") && !strings.Contains(l.out("k", "-6", "rule", "show"), "5204:")
	})
	if !l.has("k", "link", "show", "meshctl0") {
		t.Error("the bridge went with the guard; a runtime may still use it")
	}
}
