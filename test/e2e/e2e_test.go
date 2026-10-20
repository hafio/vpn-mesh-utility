//go:build e2e && linux

// Package e2e runs the real binary against kernel WireGuard in network
// namespaces: three VPS nodes on a bridge, the server on the first one behind
// a TLS proxy (agents only speak https to non-loopback servers), and for the
// spoke tests two spokes, a phone (a namespace that brings its profile up as
// the WireGuard app does) and an "internet" namespace that routes for
// everyone and answers who is asking. The NAS spoke runs the userspace
// backend, as its container does (wireguard-go on /dev/net/tun); while it
// runs, its UAPI socket is the host's /var/run/wireguard/wg0.sock. It needs
// root and the wireguard module; `./scripts/dev.sh e2e` builds everything and
// runs it under sudo.
// Firewall management is off (--no-firewall): firewalld is reached over the
// system bus, which network namespaces do not isolate, so the lab adds the
// exit servers' masquerade with nft itself. For the same reason the agents
// must not register their resolvers with the host's systemd-resolved; they
// see they run in another network namespace and report DNS "unavailable",
// while each resolver still answers on its node's mesh IP.
package e2e

import (
	"bufio"
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"flag"
	"fmt"
	"io"
	"log"
	"math/big"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	mdns "github.com/miekg/dns"
	"github.com/vishvananda/netns"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"

	"github.com/hafio/vpn-mesh-utility/internal/agent/localapi"
	"github.com/hafio/vpn-mesh-utility/internal/proto"
	"github.com/hafio/vpn-mesh-utility/internal/server/api"
)

var meshctl = flag.String("meshctl", "", "path of the meshctl binary under test")

const nsPrefix = "mce2e-"

type host struct{ ns, name, ip, meshIP string }

// The reference servers, one namespace each.
var nodes = []host{
	{"a", "vps-1", "192.0.2.1", "10.77.0.1"},
	{"b", "vps-2", "192.0.2.2", "10.77.0.2"},
	{"c", "vps-3", "192.0.2.3", "10.77.0.3"},
}

// The spokes, a phone, and the "internet": a router for every namespace
// with the echo service behind it.
var (
	laptop   = host{"d", "laptop", "192.0.2.4", "10.77.2.1"}
	nas      = host{"f", "nas", "192.0.2.5", "10.77.1.1"}
	phone    = host{"g", "phone", "192.0.2.6", "10.77.3.1"}
	internet = host{"e", "", "192.0.2.100", ""}
	echoAddr = "198.51.100.1"
)

// spaces are every namespace a lab may create.
var spaces = []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "br"}

func TestMain(m *testing.M) {
	if os.Getenv("E2E_ROLE") == "proxy" {
		runProxy()
		return
	}
	flag.Parse()
	if os.Geteuid() != 0 {
		fmt.Println("e2e: needs root; run it through ./scripts/dev.sh e2e")
		os.Exit(1)
	}
	os.Exit(m.Run())
}

// runProxy is the TLS terminator in front of the server, run as a helper
// process inside the server's namespace.
func runProxy() {
	target, err := url.Parse(os.Getenv("E2E_TARGET"))
	if err != nil {
		log.Fatal(err)
	}
	srv := &http.Server{Addr: os.Getenv("E2E_LISTEN"), Handler: httputil.NewSingleHostReverseProxy(target), ReadHeaderTimeout: 10 * time.Second}
	log.Fatal(srv.ListenAndServeTLS(os.Getenv("E2E_CERT"), os.Getenv("E2E_KEY")))
}

type lab struct {
	t     *testing.T
	dir   string
	bin   string
	ca    string
	procs map[string]*exec.Cmd
	// runtimes are the container runtime sockets agents read names from,
	// by namespace; every other agent reads none (never the host's).
	runtimes map[string]string
	// serverArgs and agentArgs are a test's extra flags for every server
	// and agent it starts.
	serverArgs, agentArgs []string
}

func ns(n string) string { return nsPrefix + n }

func (l *lab) sh(args ...string) string {
	l.t.Helper()
	out, err := exec.Command(args[0], args[1:]...).CombinedOutput()
	if err != nil {
		l.t.Fatalf("%s: %v\n%s", strings.Join(args, " "), err, out)
	}
	return string(out)
}

func newLab(t *testing.T) *lab {
	t.Helper()
	if *meshctl == "" {
		t.Fatal("pass -meshctl=<path of the binary under test>")
	}
	bin, err := filepath.Abs(*meshctl)
	if err != nil {
		t.Fatal(err)
	}
	l := &lab{t: t, dir: t.TempDir(), bin: bin, procs: map[string]*exec.Cmd{}, runtimes: map[string]string{}}
	deleteNamespaces()
	t.Cleanup(l.teardown)

	l.sh("ip", "netns", "add", ns("br"))
	l.sh("ip", "-n", ns("br"), "link", "add", "br0", "type", "bridge")
	l.sh("ip", "-n", ns("br"), "link", "set", "br0", "up")
	for _, n := range nodes {
		l.addHost(n)
	}
	l.ca = l.writeCert("192.0.2.1")
	return l
}

// addHost plugs a namespace into the bridge.
func (l *lab) addHost(n host) {
	l.t.Helper()
	l.sh("ip", "netns", "add", ns(n.ns))
	l.sh("ip", "-n", ns(n.ns), "link", "add", "eth0", "type", "veth", "peer", "name", "p-"+n.ns, "netns", ns("br"))
	l.sh("ip", "-n", ns("br"), "link", "set", "p-"+n.ns, "master", "br0", "up")
	l.sh("ip", "-n", ns(n.ns), "addr", "add", n.ip+"/24", "dev", "eth0")
	l.sh("ip", "-n", ns(n.ns), "link", "set", "eth0", "up")
	l.sh("ip", "-n", ns(n.ns), "link", "set", "lo", "up")
}

func deleteNamespaces() {
	for _, n := range spaces {
		_ = exec.Command("ip", "netns", "del", ns(n)).Run() // absent is fine
	}
}

func (l *lab) teardown() {
	for name := range l.procs {
		l.stop(name)
	}
	deleteNamespaces()
	if l.t.Failed() {
		for _, f := range []string{"server", "proxy", "agent-a", "agent-b", "agent-c", "agent-d", "agent-f", "agent-h", "agent-k"} {
			b, _ := os.ReadFile(filepath.Join(l.dir, f+".log"))
			l.t.Logf("--- %s.log (tail) ---\n%s", f, tail(string(b), 40))
		}
	}
}

func tail(s string, n int) string {
	lines := strings.Split(strings.TrimRight(s, "\n"), "\n")
	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}
	return strings.Join(lines, "\n")
}

// writeCert creates a self-signed certificate for ip; agents trust it
// through SSL_CERT_FILE.
func (l *lab) writeCert(ip string) string {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		l.t.Fatal(err)
	}
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: "meshctl e2e"},
		NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().Add(24 * time.Hour),
		IPAddresses: []net.IP{net.ParseIP(ip)}, IsCA: true, BasicConstraintsValid: true,
		KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		l.t.Fatal(err)
	}
	kb, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		l.t.Fatal(err)
	}
	cert := filepath.Join(l.dir, "cert.pem")
	_ = os.WriteFile(cert, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}), 0o644)
	_ = os.WriteFile(filepath.Join(l.dir, "key.pem"), pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: kb}), 0o600)
	return cert
}

// start runs a long-lived process in a namespace, logging to <name>.log.
func (l *lab) start(name, space string, env []string, prog string, args ...string) {
	l.t.Helper()
	logf, err := os.Create(filepath.Join(l.dir, name+".log"))
	if err != nil {
		l.t.Fatal(err)
	}
	cmd := exec.Command("ip", append([]string{"netns", "exec", ns(space), prog}, args...)...)
	cmd.Env = append(os.Environ(), env...)
	cmd.Stdout, cmd.Stderr = logf, logf
	if err := cmd.Start(); err != nil {
		l.t.Fatal(err)
	}
	l.procs[name] = cmd
}

func (l *lab) stop(name string) {
	cmd := l.procs[name]
	if cmd == nil {
		return
	}
	delete(l.procs, name)
	_ = cmd.Process.Signal(syscall.SIGTERM)
	done := make(chan struct{})
	go func() { _ = cmd.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(10 * time.Second):
		_ = cmd.Process.Kill()
		<-done
	}
}

// meshctl runs the binary to completion, in a namespace when space != "".
func (l *lab) meshctl(space string, env []string, args ...string) (string, error) {
	argv := append([]string{l.bin}, args...)
	if space != "" {
		argv = append([]string{"ip", "netns", "exec", ns(space)}, argv...)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, argv[0], argv[1:]...)
	cmd.Env = append(os.Environ(), env...)
	var out bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &out
	err := cmd.Run()
	return out.String(), err
}

func (l *lab) data(name string) string { return filepath.Join(l.dir, name) }

func (l *lab) startServer() {
	args := []string{"server", "--data", l.data("server"), "--listen", "127.0.0.1:8080",
		"--url", "https://192.0.2.1:8443", "--trusted-proxies", "127.0.0.1/32", "--log-level", "debug"}
	l.start("server", "a", nil, l.bin, append(args, l.serverArgs...)...)
	self, err := os.Executable()
	if err != nil {
		l.t.Fatal(err)
	}
	l.start("proxy", "a", []string{"E2E_ROLE=proxy", "E2E_LISTEN=192.0.2.1:8443", "E2E_TARGET=http://127.0.0.1:8080",
		"E2E_CERT=" + l.ca, "E2E_KEY=" + filepath.Join(l.dir, "key.pem")}, self)
	eventually(l.t, 20*time.Second, "the server's admin socket", func() bool {
		_, err := l.meshctl("", nil, "node", "list", "--data", l.data("server"))
		return err == nil
	})
}

// startAgent runs a node's agent: the NAS on the userspace backend,
// everyone else on the kernel's; a fake container runtime where the test
// gave one.
func (l *lab) startAgent(space string) {
	backend := "kernel"
	if space == nas.ns {
		backend = "userspace"
	}
	sock := "off"
	if s, ok := l.runtimes[space]; ok {
		sock = s
	}
	args := []string{"agent", "run", "--data", l.data(space), "--socket", filepath.Join(l.data(space), "agent.sock"), "--no-firewall",
		"--wg-backend", backend, "--log-level", "debug", "--runtime-socket", sock}
	l.start("agent-"+space, space, []string{"SSL_CERT_FILE=" + l.ca}, l.bin, append(args, l.agentArgs...)...)
}

// uapiGet asks a userspace device for its state over WireGuard's UAPI
// socket, as `wg show` does.
func uapiGet(iface string) (string, error) {
	c, err := net.DialTimeout("unix", "/var/run/wireguard/"+iface+".sock", 2*time.Second)
	if err != nil {
		return "", err
	}
	defer c.Close()
	_ = c.SetDeadline(time.Now().Add(5 * time.Second))
	if _, err := io.WriteString(c, "get=1\n\n"); err != nil {
		return "", err
	}
	var b strings.Builder
	sc := bufio.NewScanner(c)
	for sc.Scan() {
		if sc.Text() == "" {
			return b.String(), nil
		}
		b.WriteString(sc.Text() + "\n")
	}
	return b.String(), fmt.Errorf("the answer ended early: %v", sc.Err())
}

// checkBackends proves the NAS runs wireguard-go and says so: its wg0 is
// meshctl's tun, `wg show` can reach it, and both agents and the server
// report the backend each spoke uses.
func (l *lab) checkBackends() {
	l.t.Helper()
	if out := l.sh("ip", "-d", "-n", ns(nas.ns), "link", "show", "wg0"); !strings.Contains(out, "tun") || !strings.Contains(out, "alias meshctl") {
		l.t.Fatalf("the NAS's wg0 is not meshctl's tun:\n%s", out)
	}
	if got, err := uapiGet("wg0"); err != nil || strings.Count(got, "public_key=") != 3 || !strings.Contains(got, "errno=0") {
		l.t.Fatalf("UAPI get on the NAS's wg0 = %v\n%s", err, got)
	}
	for _, s := range []host{laptop, nas} {
		want := "kernel"
		if s == nas {
			want = "userspace"
		}
		if st, err := l.status(s.ns); err != nil || st.WGBackend != want {
			l.t.Fatalf("%s reports backend %q (%v), want %s", s.name, st.WGBackend, err, want)
		}
		eventually(l.t, 40*time.Second, "the server to record "+s.name+"'s backend", func() bool {
			out, err := l.meshctl("", nil, "node", "show", s.name, "--data", l.data("server"), "--json")
			var d api.NodeDetail
			return err == nil && json.Unmarshal([]byte(out), &d) == nil && d.Node.WGBackend == want
		})
	}
}

func (l *lab) status(space string) (proto.LocalStatus, error) {
	out, err := l.meshctl("", nil, "status", "--socket", filepath.Join(l.data(space), "agent.sock"), "--json")
	if err != nil {
		return proto.LocalStatus{}, fmt.Errorf("%v: %s", err, out)
	}
	var st proto.LocalStatus
	return st, json.Unmarshal([]byte(out), &st)
}

// meshView asks a node's agent for its view of the mesh.
func (l *lab) meshView(space string) proto.LocalMesh {
	l.t.Helper()
	m, err := localapi.NewClient(filepath.Join(l.data(space), "agent.sock")).Mesh(context.Background())
	if err != nil {
		l.t.Fatalf("mesh view on %s: %v", space, err)
	}
	return m
}

// checkLocalUI signs in to a node's local UI with the link `meshctl ui`
// prints, as a browser on that node would, and finds the mesh there; a
// change from another page is refused (amendment A7).
func (l *lab) checkLocalUI(space string) {
	t := l.t
	out, err := l.meshctl("", nil, "ui", "--socket", filepath.Join(l.data(space), "agent.sock"))
	link, _, _ := strings.Cut(out, "\n")
	if err != nil || !strings.HasPrefix(link, "http://127.0.0.1:7443/login?t=") {
		t.Fatalf("meshctl ui on %s: %v\n%s", space, err, out)
	}
	jar, _ := cookiejar.New(nil)
	browser := &http.Client{Jar: jar, Timeout: 5 * time.Second, Transport: &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			var c net.Conn
			err := inNS(space, func() error {
				var err error
				c, err = (&net.Dialer{Timeout: 2 * time.Second}).DialContext(ctx, network, addr)
				return err
			})
			return c, err
		}}}
	resp, err := browser.Get(link)
	if err != nil {
		t.Fatalf("local UI on %s: %v", space, err)
	}
	page, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	for _, want := range []string{"<h1>laptop", "Restart the tunnel", ">nas<", ">phone<", ">vps-3<"} {
		if resp.StatusCode != http.StatusOK || !strings.Contains(string(page), want) {
			t.Fatalf("local UI dashboard on %s = %d, no %q:\n%s", space, resp.StatusCode, want, page)
		}
	}
	resp, err = browser.PostForm("http://127.0.0.1:7443/restart", url.Values{"csrf": {"x"}})
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("a change without the UI's Origin = %d, want 403", resp.StatusCode)
	}
}

// meshNames are the nodes a mesh view lists, sorted.
func meshNames(m proto.LocalMesh) string {
	var names []string
	for _, n := range m.View.Nodes {
		names = append(names, n.Name)
	}
	slices.Sort(names)
	return strings.Join(names, ",")
}

// resolve asks the resolver at server:53, from inside space, for name's A
// record.
func (l *lab) resolve(space, server, name string) (ip string, rcode int, err error) {
	err = inNS(space, func() error {
		c := &mdns.Client{Timeout: 2 * time.Second}
		m := new(mdns.Msg)
		m.SetQuestion(name, mdns.TypeA)
		resp, _, err := c.Exchange(m, net.JoinHostPort(server, "53"))
		if err != nil {
			return err
		}
		rcode = resp.Rcode
		if len(resp.Answer) == 1 {
			if a, ok := resp.Answer[0].(*mdns.A); ok {
				ip = a.A.String()
			}
		}
		return nil
	})
	return ip, rcode, err
}

// checkDNS proves every node answers the mesh zone on its mesh IP -- its
// own resolver for a spoke, a server's from another node -- refuses to
// forward other names in split mode, and left the host's resolver alone.
func (l *lab) checkDNS() {
	t := l.t
	for _, c := range []struct{ from, at, name, want string }{
		{laptop.ns, laptop.meshIP, "nas.mesh.", nas.meshIP},
		{laptop.ns, laptop.meshIP, "vps-3.mesh.", nodes[2].meshIP},
		{nas.ns, nodes[0].meshIP, "laptop.mesh.", laptop.meshIP},
	} {
		eventually(t, 10*time.Second, c.name+" from "+c.from+" at "+c.at, func() bool {
			ip, _, err := l.resolve(c.from, c.at, c.name)
			return err == nil && ip == c.want
		})
	}
	if _, rcode, err := l.resolve(laptop.ns, laptop.meshIP, "example.com."); err != nil || rcode != mdns.RcodeRefused {
		t.Fatalf("split mode: example.com = rcode %d, %v; want REFUSED", rcode, err)
	}
	for _, s := range []host{laptop, nas} {
		if st, err := l.status(s.ns); err != nil || st.DNS != proto.DNSUnavailable {
			t.Fatalf("%s reports DNS %q (%v); an agent in a lab namespace must not register with the host's resolver", s.name, st.DNS, err)
		}
	}
}

func (l *lab) ping(space, ip string) bool {
	return exec.Command("ip", "netns", "exec", ns(space), "ping", "-c", "1", "-W", "1", ip).Run() == nil
}

// eventually waits for cond, logging a wait over 2 s so e2e.log shows where
// a slow run spent its time.
func eventually(t *testing.T, within time.Duration, what string, cond func() bool) {
	t.Helper()
	start := time.Now()
	deadline := start.Add(within)
	for time.Now().Before(deadline) {
		if cond() {
			if d := time.Since(start); d > 2*time.Second {
				t.Logf("%s took %s", what, d.Round(100*time.Millisecond))
			}
			return
		}
		time.Sleep(250 * time.Millisecond)
	}
	t.Fatalf("not within %s: %s", within, what)
}

func peerNames(st proto.LocalStatus) []string {
	var out []string
	for _, p := range st.Peers {
		out = append(out, p.Name)
	}
	return out
}

// enroll adds a node through the CLI, as an operator would, joins its agent
// with the one-time token and starts it.
func (l *lab) enroll(n host, args ...string) {
	l.t.Helper()
	add := append([]string{"node", "add", n.name}, args...)
	out, err := l.meshctl("", nil, append(add, "--data", l.data("server"), "--json")...)
	if err != nil {
		l.t.Fatalf("node add %s: %v\n%s", n.name, err, out)
	}
	var ch api.NodeChange
	if err := json.Unmarshal([]byte(out), &ch); err != nil || ch.Enrollment == nil || ch.Node.MeshIP != n.meshIP {
		l.t.Fatalf("node add %s: %v\n%s", n.name, err, out)
	}
	env := []string{"SSL_CERT_FILE=" + l.ca, "MESHCTL_TOKEN=" + ch.Enrollment.Token}
	join := []string{"agent", "join", "--server", "https://192.0.2.1:8443", "--data", l.data(n.ns), "--no-service"}
	if ch.Node.Role == string(proto.RoleServer) {
		join = append(join, "--endpoint", n.ip+":51820")
	}
	if out, err := l.meshctl(n.ns, env, join...); err != nil {
		l.t.Fatalf("agent join %s: %v\n%s", n.name, err, out)
	}
	l.startAgent(n.ns)
}

// startMesh enrolls the three servers and waits until they pass traffic:
// join -> handshake within 60 s (spec S10 acceptance).
func (l *lab) startMesh() {
	l.t.Helper()
	for _, n := range nodes {
		l.enroll(n, "--role", "server", "--endpoint", n.ip+":51820")
	}
	eventually(l.t, 60*time.Second, "every server to apply a config with two peers and pass traffic", func() bool {
		for _, n := range nodes {
			st, err := l.status(n.ns)
			if err != nil || st.AppliedVersion == 0 || len(st.Peers) != 2 {
				return false
			}
		}
		return l.ping("b", "10.77.0.1") && l.ping("c", "10.77.0.2") && l.ping("a", "10.77.0.3")
	})
}

func TestServerMesh(t *testing.T) {
	l := newLab(t)
	l.startServer()
	l.startMesh()
	st, _ := l.status("a")
	for _, p := range st.Peers {
		if p.LastHandshake == 0 {
			t.Fatalf("vps-1 has no handshake with %s: %+v", p.Name, st.Peers)
		}
	}

	// Disable -> gone from every peer within 30 s.
	if out, err := l.meshctl("", nil, "node", "disable", "vps-3", "--data", l.data("server")); err != nil {
		t.Fatalf("node disable: %v\n%s", err, out)
	}
	eventually(t, 30*time.Second, "vps-3 to leave vps-1 and vps-2", func() bool {
		a, errA := l.status("a")
		b, errB := l.status("b")
		return errA == nil && errB == nil && strings.Join(peerNames(a), ",") == "vps-2" && strings.Join(peerNames(b), ",") == "vps-1"
	})
	if l.ping("a", "10.77.0.3") {
		t.Fatal("the disabled node is still reachable")
	}

	// Server down + agent restart -> the cached config comes back.
	before, err := l.status("b")
	if err != nil {
		t.Fatal(err)
	}
	l.stop("proxy")
	l.stop("server")
	l.stop("agent-b")
	l.sh("ip", "-n", ns("b"), "link", "del", "wg0")
	l.startAgent("b")
	eventually(t, 30*time.Second, "vps-2 to rebuild wg0 from its cache and reach vps-1", func() bool {
		st, err := l.status("b")
		return err == nil && st.AppliedVersion == before.AppliedVersion && len(st.Peers) == 1 && l.ping("b", "10.77.0.1")
	})
}

// inNS runs fn on an OS thread switched into a lab namespace: the sockets fn
// opens belong to that namespace for their whole life. A thread that cannot
// be switched back is left locked, so the runtime retires it.
func inNS(space string, fn func() error) error {
	errc := make(chan error, 1)
	go func() {
		runtime.LockOSThread()
		orig, err := netns.Get()
		if err != nil {
			errc <- err
			return
		}
		defer orig.Close()
		target, err := netns.GetFromName(ns(space))
		if err != nil {
			errc <- err
			return
		}
		defer target.Close()
		if err := netns.Set(target); err != nil {
			errc <- err
			return
		}
		ferr := fn()
		if err := netns.Set(orig); err != nil {
			errc <- fmt.Errorf("switch back from %s: %w", space, err)
			return
		}
		runtime.UnlockOSThread()
		errc <- ferr
	}()
	return <-errc
}

// addInternet makes namespace e the router of every namespace, with the echo
// service on echoAddr answering the address each request came from.
func (l *lab) addInternet(hosts ...host) {
	l.t.Helper()
	l.addHost(internet)
	l.sh("ip", "-n", ns(internet.ns), "addr", "add", echoAddr+"/32", "dev", "lo")
	for _, h := range hosts {
		l.sh("ip", "-n", ns(h.ns), "route", "add", "default", "via", internet.ip)
	}
	var ln net.Listener
	if err := inNS(internet.ns, func() error {
		var err error
		ln, err = net.Listen("tcp", echoAddr+":80")
		return err
	}); err != nil {
		l.t.Fatal(err)
	}
	srv := &http.Server{ReadHeaderTimeout: 5 * time.Second, Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, _ := net.SplitHostPort(r.RemoteAddr)
		fmt.Fprintln(w, ip)
	})}
	go func() { _ = srv.Serve(ln) }()
	l.t.Cleanup(func() { _ = srv.Close() })
}

// masquerade does on an exit server what firewalld's meshctl-exit policy
// does when the agent manages the firewall.
func (l *lab) masquerade(space string) {
	l.t.Helper()
	rules := filepath.Join(l.dir, "nat-"+space+".nft")
	if err := os.WriteFile(rules, []byte(`table ip e2e {
	chain post {
		type nat hook postrouting priority 100;
		oifname "eth0" ip saddr 10.77.0.0/16 masquerade
	}
}
`), 0o600); err != nil {
		l.t.Fatal(err)
	}
	l.sh("ip", "netns", "exec", ns(space), "nft", "-f", rules)
}

// whoami asks the echo service, from inside space, which address it sees.
func (l *lab) whoami(space string) string {
	tr := &http.Transport{DisableKeepAlives: true, DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
		var c net.Conn
		err := inNS(space, func() error {
			var err error
			c, err = (&net.Dialer{Timeout: 2 * time.Second}).DialContext(ctx, network, addr)
			return err
		})
		return c, err
	}}
	resp, err := (&http.Client{Transport: tr, Timeout: 3 * time.Second}).Get("http://" + echoAddr + "/")
	if err != nil {
		return "error: " + err.Error()
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(io.LimitReader(resp.Body, 256))
	return strings.TrimSpace(string(b))
}

// stream is a TCP connection from a spoke to an echo listener on a server's
// mesh address; check proves it still carries data.
type stream struct {
	c  net.Conn
	rd *bufio.Reader
}

func (l *lab) openStream(from, serverNS, meshIP string) *stream {
	l.t.Helper()
	var ln net.Listener
	if err := inNS(serverNS, func() error {
		var err error
		ln, err = net.Listen("tcp", meshIP+":7000")
		return err
	}); err != nil {
		l.t.Fatal(err)
	}
	l.t.Cleanup(func() { _ = ln.Close() })
	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			go func() { _, _ = io.Copy(c, c); _ = c.Close() }()
		}
	}()
	var c net.Conn
	if err := inNS(from, func() error {
		var err error
		c, err = net.DialTimeout("tcp", meshIP+":7000", 5*time.Second)
		return err
	}); err != nil {
		l.t.Fatal(err)
	}
	l.t.Cleanup(func() { _ = c.Close() })
	return &stream{c: c, rd: bufio.NewReader(c)}
}

func (s *stream) check(t *testing.T, msg string) {
	t.Helper()
	_ = s.c.SetDeadline(time.Now().Add(5 * time.Second))
	if _, err := fmt.Fprintln(s.c, msg); err != nil {
		t.Fatalf("the stream broke (write %q): %v", msg, err)
	}
	line, err := s.rd.ReadString('\n')
	if err != nil || strings.TrimSpace(line) != msg {
		t.Fatalf("the stream broke (read %q: %q, %v)", msg, line, err)
	}
}

// wgQuick brings a WireGuard profile up in space the way the phone's app
// does: the interface, its address, and a route for every allowed IPv4
// prefix (the lab has no IPv6). It returns the profile's DNS server.
func (l *lab) wgQuick(space, profile string) string {
	t := l.t
	t.Helper()
	var (
		key       wgtypes.Key
		addr, dns string
		peers     []wgtypes.PeerConfig
		routes    []string
		err       error
	)
	mtu := "1420"
	for _, line := range strings.Split(profile, "\n") {
		k, v, _ := strings.Cut(line, " = ")
		var p *wgtypes.PeerConfig
		if len(peers) > 0 {
			p = &peers[len(peers)-1]
		}
		switch k {
		case "[Peer]":
			peers = append(peers, wgtypes.PeerConfig{ReplaceAllowedIPs: true})
		case "PrivateKey":
			key, err = wgtypes.ParseKey(v)
		case "Address":
			addr = v
		case "DNS":
			dns = v
		case "MTU":
			mtu = v
		case "PublicKey":
			p.PublicKey, err = wgtypes.ParseKey(v)
		case "Endpoint":
			p.Endpoint, err = net.ResolveUDPAddr("udp", v)
		case "AllowedIPs":
			for _, a := range strings.Split(v, ", ") {
				_, n, perr := net.ParseCIDR(a)
				if perr != nil {
					t.Fatalf("AllowedIPs %q: %v", a, perr)
				}
				p.AllowedIPs = append(p.AllowedIPs, *n)
				if n.IP.To4() != nil {
					routes = append(routes, a)
				}
			}
		case "PersistentKeepalive":
			var s int
			s, err = strconv.Atoi(v)
			d := time.Duration(s) * time.Second
			p.PersistentKeepaliveInterval = &d
		}
		if err != nil {
			t.Fatalf("profile line %q: %v", line, err)
		}
	}
	l.sh("ip", "-n", ns(space), "link", "add", "wg0", "type", "wireguard")
	if err := inNS(space, func() error {
		c, err := wgctrl.New()
		if err != nil {
			return err
		}
		defer c.Close()
		return c.ConfigureDevice("wg0", wgtypes.Config{PrivateKey: &key, ReplacePeers: true, Peers: peers})
	}); err != nil {
		t.Fatal(err)
	}
	l.sh("ip", "-n", ns(space), "addr", "add", addr, "dev", "wg0")
	l.sh("ip", "-n", ns(space), "link", "set", "wg0", "mtu", mtu, "up")
	for _, r := range routes {
		if r == "0.0.0.0/0" {
			r = "default"
		}
		l.sh("ip", "-n", ns(space), "route", "add", r, "dev", "wg0")
	}
	return dns
}

// checkPhone adds a phone whose hub and exit is vps-2 and imports its
// via-exit profile as the WireGuard app would. The phone reaches the NAS
// through its hub (amendment A3), resolves mesh names at the hub's resolver
// and sees the internet through vps-2 (spec S10: phone profiles).
func (l *lab) checkPhone() {
	t := l.t
	l.addHost(phone)
	if out, err := l.meshctl("", nil, "node", "add", phone.name, "--role", "phone", "--hub", "vps-2", "--data", l.data("server")); err != nil {
		t.Fatalf("node add phone: %v\n%s", err, out)
	}
	out, err := l.meshctl("", nil, "node", "phone", phone.name, "--exit", "vps-2", "--json", "--data", l.data("server"))
	var p api.PhoneProfile
	if err != nil || json.Unmarshal([]byte(out), &p) != nil {
		t.Fatalf("node phone: %v\n%s", err, out)
	}
	if dns := l.wgQuick(phone.ns, p.Profile); dns != nodes[1].meshIP {
		t.Fatalf("the profile's DNS is %q, want its hub's resolver %s", dns, nodes[1].meshIP)
	}
	eventually(t, 30*time.Second, "the phone to reach the NAS through its hub", func() bool { return l.ping(phone.ns, nas.meshIP) })
	if ip, _, err := l.resolve(phone.ns, nodes[1].meshIP, "nas.mesh."); err != nil || ip != nas.meshIP {
		t.Fatalf("nas.mesh from the phone at its hub = %q, %v", ip, err)
	}
	if got := l.whoami(phone.ns); got != nodes[1].ip {
		t.Fatalf("the phone's internet is seen as %q, want its exit vps-2 (%s)", got, nodes[1].ip)
	}
}

// phoneSeen waits for a server's heartbeat to report the phone's handshake,
// which is all the server ever learns of a phone.
func (l *lab) phoneSeen() {
	eventually(l.t, 45*time.Second, "a server to report the phone's handshake", func() bool {
		out, err := l.meshctl("", nil, "node", "list", "--json", "--data", l.data("server"))
		var list api.NodeList
		if err != nil || json.Unmarshal([]byte(out), &list) != nil {
			return false
		}
		for _, n := range list.Nodes {
			if n.Name == phone.name {
				return n.LastHandshake > 0 && n.InSync
			}
		}
		return false
	})
}

// exit runs `meshctl exit` against a spoke's agent.
func (l *lab) exit(space, via string) {
	l.t.Helper()
	out, err := l.meshctl("", nil, "exit", via, "--socket", filepath.Join(l.data(space), "agent.sock"))
	if err != nil {
		l.t.Fatalf("meshctl exit %s: %v\n%s", via, err, out)
	}
}

// Spokes behind different hubs reach each other through one relay
// (amendment A3), the NAS on the userspace backend; every node answers the
// mesh zone (spec S7.4); a laptop switches its internet exit between servers
// without dropping a connection over the mesh, then turns it off (spec
// S7.3, S10 acceptance); a phone's via-exit profile works and its servers
// report its handshake; the laptop's local view of the mesh lists every
// node, in its local UI too; with the server stopped, a restarted laptop still resolves nas.mesh
// (spec S10 acceptance), shows its cached mesh view and serves its log.
func TestSpokesAndExit(t *testing.T) {
	l := newLab(t)
	l.addHost(laptop)
	l.addHost(nas)
	l.addInternet(append(slices.Clone(nodes), laptop, nas)...)
	l.startServer()
	l.startMesh()
	for _, n := range nodes[1:] {
		if out, err := l.meshctl("", nil, "node", "set", n.name, "--can-exit", "--data", l.data("server")); err != nil {
			t.Fatalf("node set %s --can-exit: %v\n%s", n.name, err, out)
		}
		l.masquerade(n.ns)
	}
	l.enroll(laptop, "--role", "spoke", "--hub", "vps-1")
	l.enroll(nas, "--role", "spoke", "--hub", "vps-3", "--always-on")

	eventually(t, 60*time.Second, "both spokes to reach each other across their hubs", func() bool {
		for _, s := range []host{laptop, nas} {
			if st, err := l.status(s.ns); err != nil || st.AppliedVersion == 0 || len(st.Peers) != 3 {
				return false
			}
		}
		return l.ping(laptop.ns, nas.meshIP) && l.ping(nas.ns, laptop.meshIP)
	})
	l.checkBackends()
	l.checkDNS()
	l.checkPhone()

	if got := l.whoami(laptop.ns); got != laptop.ip {
		t.Fatalf("with exit off the internet sees %q, want the laptop's own %s", got, laptop.ip)
	}
	s := l.openStream(laptop.ns, "a", "10.77.0.1")
	s.check(t, "before")

	for i, n := range nodes[1:] {
		l.exit(laptop.ns, n.name)
		eventually(t, 5*time.Second, "the internet to see "+n.name, func() bool { return l.whoami(laptop.ns) == n.ip })
		s.check(t, fmt.Sprintf("after exit %d", i))
	}
	st, err := l.status(laptop.ns)
	if err != nil || st.Exit != "vps-3" || strings.Join(st.ExitChoices, ",") != "vps-2,vps-3" {
		t.Fatalf("status after the switch: %+v, %v", st, err)
	}

	l.exit(laptop.ns, "off")
	eventually(t, 5*time.Second, "the laptop's own address again", func() bool { return l.whoami(laptop.ns) == laptop.ip })
	s.check(t, "after exit off")
	for _, family := range []string{"-4", "-6"} {
		if out := l.sh("ip", family, "-n", ns(laptop.ns), "rule"); strings.Contains(out, "51820") {
			t.Fatalf("exit off left policy rules (%s):\n%s", family, out)
		}
	}
	l.phoneSeen()
	// The laptop sees the mesh as the server's dashboard does (spec S7.6).
	if m := l.meshView(laptop.ns); m.Error != "" || meshNames(m) != "laptop,nas,phone,vps-1,vps-2,vps-3" {
		t.Fatalf("the laptop's mesh view: %+v", m)
	}
	l.checkLocalUI(laptop.ns)

	l.stop("proxy")
	l.stop("server")
	l.stop("agent-" + laptop.ns)
	l.sh("ip", "-n", ns(laptop.ns), "link", "del", "wg0")
	l.startAgent(laptop.ns)
	eventually(t, 30*time.Second, "nas.mesh from the restarted laptop, the server down", func() bool {
		ip, _, err := l.resolve(laptop.ns, laptop.meshIP, "nas.mesh.")
		return err == nil && ip == nas.meshIP && l.ping(laptop.ns, nas.meshIP)
	})
	// Its mesh view is the cached one, and says why.
	if m := l.meshView(laptop.ns); m.Error == "" || m.FetchedAt == 0 || meshNames(m) != "laptop,nas,phone,vps-1,vps-2,vps-3" {
		t.Fatalf("the restarted laptop's mesh view, the server down: %+v", m)
	}
	if out, err := l.meshctl("", nil, "logs", "--socket", filepath.Join(l.data(laptop.ns), "agent.sock")); err != nil ||
		!strings.Contains(out, "meshctl agent running") {
		t.Fatalf("meshctl logs on the laptop: %v\n%s", err, out)
	}
}
