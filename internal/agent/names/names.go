// Package names reads what the container runtime calls the containers on
// the node's mesh network (amendment A38), so the agent can report them and
// every node's resolver answers them.
//
// The runtime is asked over its Docker-compatible API on a unix socket --
// Docker's, or Podman's, which speaks the same -- with GET requests only:
// the containers on the mesh network, each one's details, and the event
// stream that says when to look again. Every answer is size-capped, every
// container ID is checked before it becomes part of a path, and every name
// passes internal/validate; a name that does not is left out, never
// mended.
package names

import (
	"cmp"
	"net/netip"
	"slices"
	"strings"

	"github.com/hafio/vpn-mesh-utility/internal/proto"
	"github.com/hafio/vpn-mesh-utility/internal/validate"
)

// Labels the runtimes set.
const (
	LabelComposeService = "com.docker.compose.service"
	LabelSwarmService   = "com.docker.swarm.service.name"
	LabelStack          = "com.docker.stack.namespace"
)

// Container is one container on the mesh network, as the runtime reports
// it: IP, IP6 (M8; invalid when the network has no IPv6) and Aliases are
// its addresses and aliases on that network.
type Container struct {
	ID         string
	Name       string
	Hostname   string
	Domainname string
	Labels     map[string]string
	IP, IP6    netip.Addr
	Aliases    []string
}

// candidates are c's names, most telling first.
func (c Container) candidates() []proto.ContainerName {
	var out []proto.ContainerName
	add := func(kind, name string) {
		if name != "" {
			out = append(out, proto.ContainerName{Name: name, Kind: kind})
		}
	}
	add(proto.NameService, c.Labels[LabelComposeService])
	if svc := c.Labels[LabelSwarmService]; svc != "" {
		if stack := c.Labels[LabelStack]; stack != "" {
			svc = strings.TrimPrefix(svc, stack+"_")
		}
		add(proto.NameSwarm, svc)
	}
	add(proto.NameContainer, strings.TrimPrefix(c.Name, "/"))
	for _, a := range c.Aliases {
		if !c.isID(a) {
			add(proto.NameAlias, a)
		}
	}
	if h := c.Hostname; h != "" && !c.isID(h) {
		add(proto.NameHostname, h)
		if c.Domainname != "" {
			add(proto.NameHostname, h+"."+c.Domainname)
		}
	}
	return out
}

// isID reports whether s is the container's ID or its short form, which
// runtimes use as the default hostname and alias.
func (c Container) isID(s string) bool {
	return len(s) >= 12 && strings.HasPrefix(c.ID, strings.ToLower(s))
}

// hostAddr reports whether ip may be a container's on subnet, of its
// family: not the network address, the runtime's bridge (.1, ::1), the
// gateway (.2, ::2), which is the agent itself in container mode, or
// IPv4's broadcast address.
func hostAddr(subnet netip.Prefix, ip netip.Addr) bool {
	if !subnet.IsValid() || !ip.IsValid() || subnet.Addr().Is4() != ip.Is4() || !subnet.Contains(ip) {
		return false
	}
	base := subnet.Masked().Addr()
	if ip == base || ip == base.Next() || ip == base.Next().Next() {
		return false
	}
	return ip.Is6() || subnet.Contains(ip.Next())
}

// Of returns the names of the containers whose address is on subnet or
// subnet6 (M8; invalid when the containers have no IPv6) -- valid, lower
// case, each name and address once, sorted by name and address -- and how
// many were left out over proto.MaxNames.
func Of(cs []Container, subnet, subnet6 netip.Prefix) ([]proto.ContainerName, int) {
	seen := map[proto.ContainerName]bool{}
	var out []proto.ContainerName
	for _, c := range cs {
		var ips []netip.Addr
		if hostAddr(subnet, c.IP) {
			ips = append(ips, c.IP)
		}
		if hostAddr(subnet6, c.IP6) {
			ips = append(ips, c.IP6)
		}
		for _, n := range c.candidates() {
			name, err := validate.Hostname("name", n.Name)
			if err != nil {
				continue
			}
			for _, ip := range ips {
				key := proto.ContainerName{Name: name, IP: ip.String()}
				if seen[key] {
					continue
				}
				seen[key] = true
				out = append(out, proto.ContainerName{Name: name, Kind: n.Kind, IP: key.IP})
			}
		}
	}
	slices.SortFunc(out, func(a, b proto.ContainerName) int {
		return cmp.Or(cmp.Compare(a.Name, b.Name), netip.MustParseAddr(a.IP).Compare(netip.MustParseAddr(b.IP)))
	})
	if over := len(out) - proto.MaxNames; over > 0 {
		return out[:proto.MaxNames], over
	}
	return out, 0
}
