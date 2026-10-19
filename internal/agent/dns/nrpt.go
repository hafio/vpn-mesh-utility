package dns

import (
	"net/netip"
	"slices"
	"strings"
)

// The Windows registration is the Name Resolution Policy Table (NRPT): one
// rule per namespace, naming the servers that answer it. The agent owns
// two, each under a fixed key: the mesh zone (".mesh" by default), and in
// exit mode "." (every name).
// The mapping lives here, apart from the registry, so every platform's
// tests pin it.

// nrptRule is one NRPT rule: the namespace it matches and the servers that
// answer it.
type nrptRule struct {
	Key     string
	Name    string
	Servers []netip.Addr
}

// The agent's rule keys under DnsPolicyConfig.
const (
	nrptMeshKey = "{6d657368-6374-6c00-8a2f-417e0c5b0001}"
	nrptRootKey = "{6d657368-6374-6c00-8a2f-417e0c5b0002}"
)

// rulesFor is the NRPT form of r: a rule for "." and one for the zone, the
// routing domains the agent sets. Search domains have no NRPT form; the
// agent never sets one.
func rulesFor(r Registration) []nrptRule {
	var out []nrptRule
	for _, d := range r.Domains {
		switch name := domainName(d.Name); {
		case !d.RouteOnly:
		case name == ".":
			out = append(out, nrptRule{Key: nrptRootKey, Name: ".", Servers: r.DNS})
		default:
			out = append(out, nrptRule{Key: nrptMeshKey, Name: "." + name, Servers: r.DNS})
		}
	}
	return out
}

// fromRules is the registration the agent's rules stand for; a "." rule is
// exit mode, so it carries the default route.
func fromRules(rules []nrptRule) Registration {
	var r Registration
	for _, rule := range rules {
		name := strings.TrimPrefix(domainName(rule.Name), ".")
		if name == "" {
			name = "."
		}
		r.Domains = append(r.Domains, Domain{Name: name, RouteOnly: true})
		r.DefaultRoute = r.DefaultRoute || name == "."
		for _, ip := range rule.Servers {
			if !slices.Contains(r.DNS, ip) {
				r.DNS = append(r.DNS, ip)
			}
		}
	}
	return r
}

// serverList is GenericDNSServers: the servers, ";"-separated.
func serverList(ips []netip.Addr) string {
	var s []string
	for _, ip := range ips {
		s = append(s, ip.String())
	}
	return strings.Join(s, ";")
}

// parseServers reads GenericDNSServers, skipping what is not an address.
func parseServers(s string) []netip.Addr {
	var out []netip.Addr
	for _, f := range strings.Split(s, ";") {
		if ip, err := netip.ParseAddr(strings.TrimSpace(f)); err == nil {
			out = append(out, ip)
		}
	}
	return out
}
