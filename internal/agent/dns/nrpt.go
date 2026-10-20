package dns

import (
	"crypto/sha256"
	"fmt"
	"net/netip"
	"slices"
	"strings"
)

// The Windows registration is the Name Resolution Policy Table (NRPT): one
// rule per namespace, naming the servers that answer it. The agent owns one
// for the mesh zone (".mesh" by default), one for each extra domain
// (amendment A39) and in exit mode one for "." (every name).
// The mapping lives here, apart from the registry, so every platform's
// tests pin it.

// nrptRule is one NRPT rule: the namespace it matches and the servers that
// answer it.
type nrptRule struct {
	Key     string
	Name    string
	Servers []netip.Addr
}

// The agent's rule keys under DnsPolicyConfig all start with nrptKeyPrefix
// ("meshctl" and a zero in hex), so it finds, and removes, its own rules
// and no other: "." under a fixed key, a domain under one its name gives,
// so a renamed zone or a dropped domain leaves no rule behind.
const (
	nrptKeyPrefix = "{6d657368-6374-6c0"
	nrptRootKey   = "{6d657368-6374-6c00-8a2f-417e0c5b0002}"
)

// nrptKey is the rule key of a domain.
func nrptKey(name string) string {
	h := sha256.Sum256([]byte(name))
	return fmt.Sprintf("{6d657368-6374-6c01-%x-%x}", h[:2], h[2:8])
}

// ownedKey reports whether a key under DnsPolicyConfig is one of the
// agent's (the registry's key names ignore case).
func ownedKey(key string) bool { return strings.HasPrefix(strings.ToLower(key), nrptKeyPrefix) }

// rulesFor is the NRPT form of r: a rule for "." and one for each domain,
// the routing domains the agent sets. Search domains have no NRPT form; the
// agent never sets one.
func rulesFor(r Registration) []nrptRule {
	var out []nrptRule
	for _, d := range r.Domains {
		switch name := domainName(d.Name); {
		case !d.RouteOnly:
		case name == ".":
			out = append(out, nrptRule{Key: nrptRootKey, Name: ".", Servers: r.DNS})
		default:
			out = append(out, nrptRule{Key: nrptKey(name), Name: "." + name, Servers: r.DNS})
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
