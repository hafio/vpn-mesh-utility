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
	"encoding/binary"
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
// it: IP and Aliases are its address and aliases on that network.
type Container struct {
	ID         string
	Name       string
	Hostname   string
	Domainname string
	Labels     map[string]string
	IP         netip.Addr
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

// hostAddr reports whether ip may be a container's on subnet: not the
// network or broadcast address, the runtime's bridge (.1) or the gateway
// (.2), which is the agent itself in container mode.
func hostAddr(subnet netip.Prefix, ip netip.Addr) bool {
	if !subnet.IsValid() || !subnet.Addr().Is4() || !ip.Is4() || !subnet.Contains(ip) {
		return false
	}
	base, a := subnet.Masked().Addr().As4(), ip.As4()
	host := binary.BigEndian.Uint32(a[:]) - binary.BigEndian.Uint32(base[:])
	return host > 2 && host < 1<<(32-subnet.Bits())-1
}

// Of returns the names of the containers whose address is on subnet --
// valid, lower case, each name and address once, sorted by name and
// address -- and how many were left out over proto.MaxNames.
func Of(cs []Container, subnet netip.Prefix) ([]proto.ContainerName, int) {
	seen := map[proto.ContainerName]bool{}
	var out []proto.ContainerName
	for _, c := range cs {
		if !hostAddr(subnet, c.IP) {
			continue
		}
		for _, n := range c.candidates() {
			name, err := validate.Hostname("name", n.Name)
			if err != nil {
				continue
			}
			key := proto.ContainerName{Name: name, IP: c.IP.String()}
			if seen[key] {
				continue
			}
			seen[key] = true
			out = append(out, proto.ContainerName{Name: name, Kind: n.Kind, IP: key.IP})
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
