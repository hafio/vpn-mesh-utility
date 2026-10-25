package api

import (
	"cmp"
	"fmt"
	"net/http"
	"net/netip"
	"net/url"
	"slices"
	"strings"

	"github.com/hafio/vpn-mesh-utility/internal/proto"
	"github.com/hafio/vpn-mesh-utility/internal/server/planner"
	"github.com/hafio/vpn-mesh-utility/internal/server/store"
	"github.com/hafio/vpn-mesh-utility/internal/validate"
)

// Container names (amendment A38). An agent reports what its runtime calls
// the containers on its mesh network; the server keeps every name whose
// address is one of the node's containers', and publishes in the mesh view
// the ones that cannot shadow anything: a name in the zone is one label,
// <name>.<zone> (every node's containers, round robin) and
// <name>.<node>.<zone> (one node's), and a node, a custom record and the
// server's own address keep theirs. A name inside one of the extra domains
// (amendment A39) is published as it is.

// NameEntry is one reported name as the DNS page shows it.
type NameEntry struct {
	Node string `json:"node"`
	Name string `json:"name"`
	Kind string `json:"kind"`
	IP   string `json:"ip"`
	// Resolves are the names it answers as when published; Dropped says
	// why it is not.
	Resolves []string `json:"resolves,omitempty"`
	Dropped  string   `json:"dropped,omitempty"`
}

// NameSet is every reported name, and what of it is published.
type NameSet struct {
	Published []proto.MeshName
	// Entries are by name, node and address.
	Entries []NameEntry
}

// containerAddr reports whether ip may be one of n's containers': on its
// /24, or its /112 (M8), and not the network address, the runtime's bridge
// (.1, ::1), the gateway (.2, ::2) or IPv4's broadcast address.
func containerAddr(n store.Node, ip netip.Addr) bool {
	subnet, _ := netip.ParsePrefix(n.ContainerSubnet)
	if ip.Is6() {
		subnet, _ = netip.ParsePrefix(n.ContainerSubnet6)
	}
	if !subnet.IsValid() || !ip.IsValid() || subnet.Addr().Is4() != ip.Is4() || !subnet.Contains(ip) {
		return false
	}
	base := subnet.Masked().Addr()
	if ip == base || ip == base.Next() || ip == base.Next().Next() {
		return false
	}
	return ip.Is6() || subnet.Contains(ip.Next())
}

// notContainerAddr says why ip is not one of n's containers'.
func notContainerAddr(n store.Node, ip string) string {
	subnet := n.ContainerSubnet
	if a, err := netip.ParseAddr(ip); err == nil && a.Is6() {
		if subnet = n.ContainerSubnet6; subnet == "" {
			return fmt.Sprintf("%s is IPv6, and the node's containers have no IPv6 on the mesh (`meshctl node set %s --containers6`)", ip, n.Name)
		}
	}
	return fmt.Sprintf("%s is not a container address of the node's network %s", ip, subnet)
}

// PublishNames decides which reported names are published. zone is the
// mesh's zone, domains its extra domains, serverHost the host of the
// server's URL.
func PublishNames(zone string, domains []string, serverHost string, nodes []store.Node, records []store.DNSRecord,
	names []store.ContainerName) NameSet {
	byID := make(map[int64]store.Node, len(nodes))
	taken := map[string]string{}
	for _, n := range nodes {
		byID[n.ID] = n
		taken[n.Name] = "node " + n.Name + " has this name"
	}
	for _, r := range records {
		if _, ok := taken[r.Name]; !ok {
			taken[r.Name] = "a DNS record has this name"
		}
	}
	host := strings.ToLower(strings.TrimSuffix(serverHost, "."))
	set := NameSet{Published: []proto.MeshName{}, Entries: []NameEntry{}}
	seen := map[proto.MeshName]bool{}
	for _, c := range names {
		n, ok := byID[c.NodeID]
		if !ok {
			continue // removed a moment ago; the cascade takes its names
		}
		e := NameEntry{Node: n.Name, Name: c.Name, Kind: c.Kind, IP: c.IP}
		ip, _ := netip.ParseAddr(c.IP)
		label, inZone := strings.CutSuffix(c.Name, "."+zone)
		extra := !inZone && slices.ContainsFunc(domains, func(d string) bool { return strings.HasSuffix(c.Name, "."+d) })
		switch {
		case n.Disabled:
			e.Dropped = "the node is disabled"
		case !containerAddr(n, ip):
			e.Dropped = "the node's container network no longer holds " + c.IP
		case c.Name == zone:
			e.Dropped = "is the zone itself"
		case slices.Contains(domains, c.Name):
			e.Dropped = "is an extra domain itself"
		case extra && c.Name == host:
			e.Dropped = "is the server's address"
		case extra:
		case strings.Contains(c.Name, ".") && (!inZone || strings.Contains(label, ".")):
			e.Dropped = fmt.Sprintf("is neither a single name in the zone %s (<name> or <name>.%s) nor in an extra domain (Settings)", zone, zone)
		case taken[label] != "":
			e.Dropped = taken[label]
		case label+"."+zone == host:
			e.Dropped = "is the server's address"
		}
		if e.Dropped == "" {
			e.Resolves = []string{c.Name}
			if !extra {
				e.Resolves = []string{label + "." + zone, label + "." + n.Name + "." + zone}
			}
		}
		set.Entries = append(set.Entries, e)
		if p := (proto.MeshName{Name: label, Node: n.Name, IP: c.IP}); e.Dropped == "" && !seen[p] {
			seen[p] = true
			set.Published = append(set.Published, p)
		}
	}
	slices.SortFunc(set.Published, func(a, b proto.MeshName) int {
		return cmp.Or(cmp.Compare(a.Name, b.Name), cmp.Compare(a.Node, b.Node), compareAddr(a.IP, b.IP))
	})
	slices.SortFunc(set.Entries, func(a, b NameEntry) int {
		return cmp.Or(cmp.Compare(a.Name, b.Name), cmp.Compare(a.Node, b.Node), compareAddr(a.IP, b.IP), cmp.Compare(a.Kind, b.Kind))
	})
	return set
}

func compareAddr(a, b string) int {
	x, _ := netip.ParseAddr(a)
	y, _ := netip.ParseAddr(b)
	return x.Compare(y)
}

// LoadNames reads a mesh's reported names and decides what is published;
// serverURL is the server's validated origin.
func LoadNames(tx *store.MeshTx, serverURL string) (NameSet, error) {
	settings, err := tx.Settings()
	if err != nil {
		return NameSet{}, err
	}
	nodes, err := tx.Nodes()
	if err != nil {
		return NameSet{}, err
	}
	records, err := tx.DNSRecords()
	if err != nil {
		return NameSet{}, err
	}
	names, err := tx.ContainerNames()
	if err != nil {
		return NameSet{}, err
	}
	var host string
	if u, err := url.Parse(serverURL); err == nil {
		host = u.Hostname()
	}
	zone := planner.Input{Zone: settings[store.SettingDNSZone]}.ZoneName()
	return PublishNames(zone, store.SplitDomains(settings[store.SettingDNSExtraDomains]), host, nodes, records, names), nil
}

func (a *api) publishedNames(tx *store.MeshTx) (NameSet, error) { return LoadNames(tx, a.ServerURL) }

// CheckServerDomains refuses extra domains (amendment A39) that hold the
// server's own address: every host would send it to the mesh's resolvers,
// which an agent needs the server to set up.
func CheckServerDomains(serverURL string, domains []string) error {
	u, err := url.Parse(serverURL)
	if err != nil {
		return nil // not a URL: no host to protect
	}
	host := strings.ToLower(u.Hostname())
	for _, d := range domains {
		if host == d || strings.HasSuffix(host, "."+d) {
			return &validate.Error{Field: "dns_extra_domains", Value: d,
				Reason: "holds the server's address " + host + ", which every node must resolve without the mesh"}
		}
	}
	return nil
}

// names takes a node's container names (amendment A38): they replace what
// the server held, a name whose address is not one of the node's
// containers' is dropped, and every agent is nudged to fetch the mesh view
// when the published names may have changed. The answer says what is not
// published and why, so the agent can log it.
func (a *api) names(w http.ResponseWriter, r *http.Request) {
	node, body, err := a.authenticate(w, r)
	if err != nil {
		a.fail(w, r, err)
		return
	}
	var req proto.NamesRequest
	if err := decodeJSON(body, &req); err != nil {
		a.fail(w, r, err)
		return
	}
	if err := req.Validate(); err != nil {
		a.fail(w, r, invalid(err))
		return
	}
	if node.ContainerSubnet == "" {
		a.fail(w, r, errorf(http.StatusConflict, proto.CodeConflict,
			"node %q has no container network, so its containers have no names on the mesh; `meshctl node set %s --containers` gives it one",
			node.Name, node.Name))
		return
	}
	resp := proto.NamesResponse{Dropped: []proto.DroppedName{}}
	changed := false
	err = a.DB.Tx(r.Context(), func(raw *store.Tx) error {
		tx := raw.In(node.MeshID)
		var keep []store.ContainerName
		for _, n := range req.Names {
			if ip, _ := netip.ParseAddr(n.IP); !containerAddr(node, ip) {
				resp.Dropped = append(resp.Dropped, proto.DroppedName{Name: n.Name, IP: n.IP, Reason: notContainerAddr(node, n.IP)})
				continue
			}
			keep = append(keep, store.ContainerName{Name: n.Name, IP: n.IP, Kind: n.Kind})
		}
		var err error
		if changed, err = tx.SetContainerNames(node.ID, keep); err != nil {
			return err
		}
		resp.Stored = len(keep)
		set, err := a.publishedNames(tx)
		for _, e := range set.Entries {
			if e.Node == node.Name && e.Dropped != "" {
				resp.Dropped = append(resp.Dropped, proto.DroppedName{Name: e.Name, IP: e.IP, Reason: e.Dropped})
			}
		}
		return err
	})
	if err != nil {
		a.fail(w, r, err)
		return
	}
	if changed {
		a.Hub.NotifyMesh(node.MeshID)
	}
	writeJSON(w, http.StatusOK, resp)
}
