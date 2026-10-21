package proto

import (
	"errors"
	"fmt"
	"net/netip"
	"slices"

	"github.com/hafio/vpn-mesh-utility/internal/validate"
)

// Validate is the agent's first apply step: it checks a config pushed by the
// server (or read back from the on-disk cache) against the same grammar the
// server validated its inputs with, before anything on the host is touched.
// Every problem is reported, not just the first.
func (c *NodeConfig) Validate() error {
	var errs []error
	add := func(err error) {
		if err != nil {
			errs = append(errs, err)
		}
	}

	if c.Version < 0 {
		add(fmt.Errorf("version: must not be negative (got %d)", c.Version))
	}
	_, err := validate.Name("node.name", c.Node.Name)
	add(err)
	_, err = validate.Enum("node.role", string(c.Node.Role), string(RoleServer), string(RoleSpoke), string(RolePhone))
	add(err)

	add(c.Interface.validate())

	names := map[string]bool{}
	keys := map[string]bool{}
	owner := map[netip.Prefix]string{}
	for i, p := range c.Peers {
		field := fmt.Sprintf("peers[%d]", i)
		if _, err := validate.Name(field+".name", p.Name); err != nil {
			add(err)
		} else if names[p.Name] {
			add(fmt.Errorf("%s.name: duplicate peer %q", field, p.Name))
		}
		names[p.Name] = true
		if _, err := validate.Key32(field+".public_key", p.PublicKey); err != nil {
			add(err)
		} else if keys[p.PublicKey] {
			add(fmt.Errorf("%s.public_key: the same key appears on two peers", field))
		}
		keys[p.PublicKey] = true
		if p.Endpoint != "" {
			_, err := validate.Endpoint(field+".endpoint", p.Endpoint)
			add(err)
		}
		if len(p.AllowedIPs) == 0 {
			add(fmt.Errorf("%s.allowed_ips: must not be empty", field))
		}
		for j, s := range p.AllowedIPs {
			pf, err := validate.Prefix4(fmt.Sprintf("%s.allowed_ips[%d]", field, j), s)
			if err != nil {
				add(err)
				continue
			}
			// WireGuard keeps a prefix on one peer only; a second assignment
			// silently moves it, so the server must never emit one.
			if prev, dup := owner[pf]; dup {
				add(fmt.Errorf("%s.allowed_ips: %s is already routed to peer %q", field, pf, prev))
			}
			owner[pf] = p.Name
		}
		if _, err := validate.IntRange(field+".keepalive", p.Keepalive, 0, 3600); err != nil {
			add(err)
		}
	}

	if c.Exit != nil {
		add(c.Exit.validate(c.Peers))
	}
	if c.DNS != nil {
		add(c.DNS.validate())
	}
	if c.Firewall != nil {
		add(c.Firewall.validate())
	}
	if c.Containers != nil {
		add(c.Containers.validate(c.Interface.Address))
	}
	if c.Policy != nil {
		if c.Node.Role != RoleServer {
			add(fmt.Errorf("policy: only a server enforces the mesh policy (this node is a %s)", c.Node.Role))
		}
		_, err := c.Policy.Matcher()
		add(err)
	}
	for i, s := range c.Sysctl {
		field := fmt.Sprintf("sysctl[%d]", i)
		_, err := validate.Enum(field+".key", s.Key, SysctlKeys()...)
		add(err)
		_, err = validate.Enum(field+".value", s.Value, "0", "1")
		add(err)
	}
	if len(errs) > 0 {
		return fmt.Errorf("config v%d for %q is invalid: %w", c.Version, c.Node.Name, errors.Join(errs...))
	}
	return nil
}

func (i Interface) validate() error {
	var errs []error
	p, err := validate.Prefix4("interface.address", i.Address)
	switch {
	case err != nil:
		// A /32 host address has no host bits, so Prefix4 accepts it; anything
		// else that fails here is malformed.
		errs = append(errs, err)
	case p.Bits() != 32:
		errs = append(errs, fmt.Errorf("interface.address: must be a /32 (got %s)", i.Address))
	}
	if i.ListenPort != nil {
		if _, err := validate.Port("interface.listen_port", *i.ListenPort); err != nil {
			errs = append(errs, err)
		}
	}
	if _, err := validate.IntRange("interface.mtu", i.MTU, MinMTU, MaxMTU); err != nil {
		errs = append(errs, err)
	}
	if _, err := validate.IntRange("interface.fwmark", i.FwMark, 0, 1<<31-1); err != nil {
		errs = append(errs, err)
	}
	return errors.Join(errs...)
}

func (e Exit) validate(peers []Peer) error {
	var errs []error
	for i, n := range e.Choices {
		if _, err := validate.Name(fmt.Sprintf("exit.choices[%d]", i), n); err != nil {
			errs = append(errs, err)
		}
	}
	if e.Via == "" {
		return errors.Join(errs...)
	}
	for _, p := range peers {
		if p.Name != e.Via {
			continue
		}
		for _, a := range p.AllowedIPs {
			if a == "0.0.0.0/0" {
				return errors.Join(errs...)
			}
		}
		return errors.Join(append(errs, fmt.Errorf("exit.via: peer %q does not carry 0.0.0.0/0", e.Via))...)
	}
	return errors.Join(append(errs, fmt.Errorf("exit.via: %q is not a peer of this node", e.Via))...)
}

func (d DNS) validate() error {
	var errs []error
	if _, err := validate.Zone("dns.zone", d.Zone); err != nil {
		errs = append(errs, err)
	}
	for i, r := range d.Records {
		field := fmt.Sprintf("dns.records[%d]", i)
		if _, err := validate.DNSName(field+".name", r.Name); err != nil {
			errs = append(errs, err)
		}
		if _, err := validate.IPv4(field+".ip", r.IP); err != nil {
			errs = append(errs, err)
		}
	}
	for i, u := range d.Upstream {
		if _, err := validate.IPv4(fmt.Sprintf("dns.upstream[%d]", i), u); err != nil {
			errs = append(errs, err)
		}
	}
	for i, f := range d.ForwardFor {
		if _, err := validate.Prefix4(fmt.Sprintf("dns.forward_for[%d]", i), f); err != nil {
			errs = append(errs, err)
		}
	}
	if got, err := validate.Domains("dns.domains", d.Zone, d.Domains); err != nil {
		errs = append(errs, err)
	} else if !slices.Equal(got, d.Domains) {
		errs = append(errs, fmt.Errorf("dns.domains: must be lower case and sorted (got %v)", d.Domains))
	}
	return errors.Join(errs...)
}

func (f Firewall) validate() error {
	var errs []error
	if _, err := validate.Enum("firewall.backend", f.Backend, FirewallFirewalld); err != nil {
		errs = append(errs, err)
	}
	if f.Bridge != "" {
		if _, err := validate.Enum("firewall.bridge", f.Bridge, ContainerBridge); err != nil {
			errs = append(errs, err)
		}
	}
	for i, r := range f.Rules {
		field := fmt.Sprintf("firewall.rules[%d]", i)
		if _, err := validate.Enum(field+".zone", r.Zone, ZonePublic, ZoneMesh); err != nil {
			errs = append(errs, err)
		}
		if _, err := validate.Enum(field+".proto", r.Proto, "tcp", "udp", ProtoICMP, ProtoAll); err != nil {
			errs = append(errs, err)
		}
		switch {
		case r.Proto == ProtoICMP && (r.Port != 0 || r.Source == ""):
			errs = append(errs, fmt.Errorf("%s: a ping rule has no port and a source (got port %d, source %q)", field, r.Port, r.Source))
		case r.Proto == ProtoAll && (r.Port != 0 || r.Source == "" || r.Zone != ZoneMesh):
			errs = append(errs, fmt.Errorf("%s: an accept-all rule is in the mesh zone, from a source, with no port (got zone %q, port %d, source %q)",
				field, r.Zone, r.Port, r.Source))
		case r.Proto != ProtoICMP && r.Proto != ProtoAll:
			if _, err := validate.Port(field+".port", r.Port); err != nil {
				errs = append(errs, err)
			}
		}
		if r.Source != "" {
			if _, err := validate.Prefix4(field+".source", r.Source); err != nil {
				errs = append(errs, err)
			}
		}
		if _, err := validate.Text(field+".note", r.Note, 200); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

// validate checks a container network against its own shape (amendment
// A37): a /24 inside the on-link network, the runtime's range in its upper
// half, the bridge and the gateway in the subnet but outside the range (so
// the runtime never hands them out), the gateway the bridge in host mode
// and the agent's own address in container mode, and the node's mesh IP
// outside the subnet.
func (k Containers) validate(address string) error {
	var errs []error
	bad := func(err error) {
		if err != nil {
			errs = append(errs, err)
		}
	}
	_, err := validate.Enum("containers.mode", k.Mode, ModeHost, ModeContainer)
	bad(err)
	subnet, err := validate.Prefix4("containers.subnet", k.Subnet)
	bad(err)
	network, err := validate.Prefix4("containers.network", k.Network)
	bad(err)
	rng, err := validate.Prefix4("containers.range", k.Range)
	bad(err)
	bridge, err := validate.IPv4("containers.bridge", k.Bridge)
	bad(err)
	gateway, err := validate.IPv4("containers.gateway", k.Gateway)
	bad(err)
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	if subnet.Bits() != 24 {
		bad(fmt.Errorf("containers.subnet: must be a /24 (got %s)", subnet))
	}
	if network.Bits() > subnet.Bits() || !network.Contains(subnet.Addr()) {
		bad(fmt.Errorf("containers.network: %s does not contain the subnet %s", network, subnet))
	}
	if rng.Bits() <= subnet.Bits() || !subnet.Contains(rng.Addr()) {
		bad(fmt.Errorf("containers.range: %s is not a part of the subnet %s", rng, subnet))
	}
	for _, a := range []struct {
		field string
		ip    netip.Addr
	}{{"containers.bridge", bridge}, {"containers.gateway", gateway}} {
		if !subnet.Contains(a.ip) || rng.Contains(a.ip) {
			bad(fmt.Errorf("%s: %s must be in %s and outside the range %s", a.field, a.ip, subnet, rng))
		}
	}
	switch {
	case k.Mode == ModeHost && gateway != bridge:
		bad(fmt.Errorf("containers.gateway: in host mode the gateway is the bridge %s (got %s)", bridge, gateway))
	case k.Mode == ModeContainer && gateway == bridge:
		bad(fmt.Errorf("containers.gateway: in container mode the gateway is the agent, not the bridge %s", bridge))
	}
	if self, err := netip.ParsePrefix(address); err == nil && subnet.Contains(self.Addr()) {
		bad(fmt.Errorf("containers.subnet: %s holds the node's own mesh IP %s", subnet, self.Addr()))
	}
	return errors.Join(errs...)
}
