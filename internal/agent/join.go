package agent

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/netip"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/hafio/vpn-mesh-utility/internal/agent/client"
	"github.com/hafio/vpn-mesh-utility/internal/agent/state"
	"github.com/hafio/vpn-mesh-utility/internal/agent/wg"
	"github.com/hafio/vpn-mesh-utility/internal/proto"
	"github.com/hafio/vpn-mesh-utility/internal/validate"
)

// Service installs and starts the agent's service (internal/agent/service).
type Service interface {
	Available() (bool, string)
	Install(ctx context.Context) error
}

// JoinOptions configure Join. The detection seams default to the host.
type JoinOptions struct {
	DataDir string
	Server  string
	Token   string
	// Name, when set, must be the name the token was issued for.
	Name string
	// Endpoint is a server's public IPv4:port, when detection would be wrong;
	// Endpoint6 its public [IPv6]:port (M8).
	Endpoint  string
	Endpoint6 string
	Version   string
	// NoService skips installing the system service.
	NoService bool
	Service   Service
	Out       io.Writer
	Log       *slog.Logger
	Now       func() time.Time

	Transport  http.RoundTripper
	Hostname   func() (string, error)
	Backend    func() string
	PublicIPv4 func() (netip.Addr, bool)
	PublicIPv6 func() (netip.Addr, bool)
}

// Join enrolls this node (spec S6.1): it generates the keys locally if they
// do not exist yet, sends only the public halves with the one-time token, and
// records what the server answered. Unless told not to, it then installs and
// starts the service.
func Join(ctx context.Context, o JoinOptions) (proto.JoinResponse, error) {
	if o.Now == nil {
		o.Now = time.Now
	}
	if o.Out == nil {
		o.Out = io.Discard
	}
	if o.Hostname == nil {
		o.Hostname = os.Hostname
	}
	if o.Backend == nil {
		o.Backend = wg.DetectBackend
	}
	if o.PublicIPv4 == nil {
		o.PublicIPv4 = PublicIPv4
	}
	if o.PublicIPv6 == nil {
		o.PublicIPv6 = PublicIPv6
	}
	server, err := validate.ServerURL("--server", o.Server)
	if err != nil {
		return proto.JoinResponse{}, err
	}
	dir, err := state.Open(o.DataDir)
	if err != nil {
		return proto.JoinResponse{}, err
	}
	keys, err := dir.LoadOrCreateKeys()
	if err != nil {
		return proto.JoinResponse{}, err
	}
	wgPub, err := keys.WGPublic()
	if err != nil {
		return proto.JoinResponse{}, err
	}
	host, _ := o.Hostname() // a missing hostname is reported as "unknown"
	req := proto.JoinRequest{
		Token: o.Token, Name: o.Name, Hostname: cleanHostname(host), OS: runtime.GOOS, Arch: runtime.GOARCH,
		WGBackend: o.Backend(), WGPublicKey: wgPub, IdentityPublicKey: keys.IdentityPublic(),
		Endpoint: o.Endpoint, AgentVersion: o.Version, Endpoint6: o.Endpoint6,
	}
	if ip, ok := o.PublicIPv4(); ok {
		req.PublicIP = ip.String()
	}
	if ip, ok := o.PublicIPv6(); ok {
		req.PublicIP6 = ip.String()
	}
	// The server validates too; failing here names the flag and spares a
	// round trip.
	if err := req.Validate(); err != nil {
		return proto.JoinResponse{}, err
	}
	resp, err := client.Join(ctx, client.Options{Server: server, Version: o.Version, Now: o.Now, Transport: o.Transport}, req)
	if err != nil {
		return proto.JoinResponse{}, fmt.Errorf("join %s: %w", server, err)
	}

	prev, err := dir.Enrollment()
	if err == nil && (prev.Server != server || prev.NodeID != resp.NodeID) {
		// Another mesh or another node: the cached config is not ours.
		if err := dir.ClearConfig(); err != nil {
			return resp, err
		}
	}
	if err := dir.SaveEnrollment(state.Enrollment{Server: server, NodeID: resp.NodeID, Name: resp.Name, Role: resp.Role,
		MeshIP: resp.MeshIP, JoinedAt: o.Now().Unix()}); err != nil {
		return resp, err
	}
	fmt.Fprintf(o.Out, "joined %s as %s (%s, mesh %s, node id %d)\n", server, resp.Name, resp.Role, meshAddresses(resp), resp.NodeID)

	if o.NoService || o.Service == nil {
		fmt.Fprintf(o.Out, "start the agent with: meshctl agent run --data %s\n", o.DataDir)
		return resp, nil
	}
	if ok, why := o.Service.Available(); !ok {
		fmt.Fprintf(o.Out, "not installing a service: %s\nstart the agent with: meshctl agent run --data %s\n", why, o.DataDir)
		return resp, nil
	}
	if err := o.Service.Install(ctx); err != nil {
		return resp, fmt.Errorf("joined, but %w", err)
	}
	fmt.Fprintln(o.Out, "installed and started meshctl-agent.service")
	return resp, nil
}

// meshAddresses is how join names the node's mesh addresses: "IP
// 10.77.0.1", "IP 10.77.0.1 and fd10:77::1" or "IPv6 fd10:77::2:1".
func meshAddresses(r proto.JoinResponse) string {
	switch {
	case r.MeshIP != "" && r.MeshIP6 != "":
		return "IP " + r.MeshIP + " and " + r.MeshIP6
	case r.MeshIP6 != "":
		return "IPv6 " + r.MeshIP6
	}
	return "IP " + r.MeshIP
}

// cleanHostname folds a host name into the grammar the server accepts
// (lowercase DNS labels); anything else becomes "unknown". The name is
// informational only.
func cleanHostname(h string) string {
	h = strings.ToLower(strings.TrimSuffix(h, "."))
	h = strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '-', r == '.':
			return r
		}
		return '-'
	}, h)
	if s, err := validate.Hostname("hostname", strings.Trim(h, "-.")); err == nil {
		return s
	}
	return "unknown"
}

var cgnat = netip.MustParsePrefix("100.64.0.0/10")

// PublicIPv4 returns the first public IPv4 address on the host's interfaces:
// the address a VPS receives WireGuard traffic on (amendment A17).
func PublicIPv4() (netip.Addr, bool) { return firstAddr(isPublic) }

// PublicIPv6 returns the first global IPv6 address on the host's interfaces
// (M8), a server's IPv6 endpoint.
func PublicIPv6() (netip.Addr, bool) { return firstAddr(isPublic6) }

// firstAddr returns the first address of an up, non-loopback interface
// that accept takes.
func firstAddr(accept func(netip.Addr) bool) (netip.Addr, bool) {
	ifs, err := net.Interfaces()
	if err != nil {
		return netip.Addr{}, false
	}
	for _, ifc := range ifs {
		if ifc.Flags&net.FlagUp == 0 || ifc.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := ifc.Addrs()
		if err != nil {
			continue
		}
		for _, a := range addrs {
			n, ok := a.(*net.IPNet)
			if !ok {
				continue
			}
			ip, parsed := netip.AddrFromSlice(n.IP)
			if ip = ip.Unmap(); parsed && accept(ip) {
				return ip, true
			}
		}
	}
	return netip.Addr{}, false
}

func isPublic(ip netip.Addr) bool {
	return ip.Is4() && ip.IsGlobalUnicast() && !ip.IsPrivate() && !cgnat.Contains(ip)
}

// isPublic6 is a global IPv6 address: not unique local (fc00::/7), link
// local or IPv4-mapped.
func isPublic6(ip netip.Addr) bool {
	return ip.Is6() && !ip.Is4In6() && ip.IsGlobalUnicast() && !ip.IsPrivate()
}
