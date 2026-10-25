package api

import (
	"context"
	"errors"
	"net/http"
	"sort"

	"github.com/hafio/vpn-mesh-utility/internal/proto"
	"github.com/hafio/vpn-mesh-utility/internal/server/store"
	"github.com/hafio/vpn-mesh-utility/internal/validate"
)

// The custom DNS records (spec S8.3 "DNS" page). Like every other change
// they go through the reconciler, so they are versioned, audited and pushed
// to the nodes whose config they change.

// MeshChange is the answer to a change that is not about one node.
type MeshChange struct {
	// Version is the config version the change created; 0 when no node's
	// computed config changed.
	Version  int64    `json:"version"`
	Warnings []string `json:"warnings,omitempty"`
}

// DNSView is the mesh zone (spec S7.4).
type DNSView struct {
	Zone string `json:"zone"`
	// Records are one per node, then the custom ones, each by name.
	Records  []DNSEntry `json:"records"`
	Upstream []string   `json:"upstream"`
	// Domains are the extra domains (amendment A39).
	Domains []string `json:"domains"`
	// Nodes are the names a custom record may follow.
	Nodes []string `json:"nodes"`
	// Names are the containers' reported names (amendment A38).
	Names []NameEntry `json:"names"`
	// IPv6 says the mesh has an IPv6 prefix, so records may carry IPv6
	// addresses (M8).
	IPv6 bool `json:"ipv6"`
}

// DNSEntry is one name of the zone.
type DNSEntry struct {
	Name string `json:"name"`
	IP   string `json:"ip"`
	// IP6 is the IPv6 address (M8); IP is empty for an IPv6-only name.
	IP6 string `json:"ip6,omitempty"`
	// Node is the node the name follows: the node itself for an automatic
	// record, the target of an alias; empty for a fixed IP.
	Node   string `json:"node,omitempty"`
	Custom bool   `json:"custom,omitempty"`
	// Live is false while the name does not resolve: its node is not
	// enrolled yet, or disabled.
	Live bool `json:"live"`
}

// RecordRequest adds a custom record: an alias of Node, or a fixed IP,
// IPv6 address or both.
type RecordRequest struct {
	Name string `json:"name"`
	Node string `json:"node,omitempty"`
	IP   string `json:"ip,omitempty"`
	IP6  string `json:"ip6,omitempty"`
}

// DNS returns the mesh zone.
func (s *Nodes) DNS(ctx context.Context) (DNSView, error) {
	v := DNSView{Records: []DNSEntry{}, Upstream: []string{}, Domains: []string{}, Nodes: []string{}}
	err := s.tx(ctx, func(tx *store.MeshTx) error {
		in, nodes, err := LoadInput(tx)
		if err != nil {
			return err
		}
		v.Zone, v.IPv6 = in.ZoneName(), in.MeshCIDR6.IsValid()
		v.Domains = append(v.Domains, in.Domains...)
		live := func(n store.Node) bool { return n.Enrolled && !n.Disabled && n.WGPubKey != "" }
		byID := map[int64]store.Node{}
		for _, n := range nodes {
			byID[n.ID] = n
			v.Nodes = append(v.Nodes, n.Name)
			v.Records = append(v.Records, DNSEntry{Name: n.Name, IP: n.MeshIP, IP6: n.MeshIP6, Node: n.Name, Live: live(n)})
		}
		records, err := tx.DNSRecords()
		if err != nil {
			return err
		}
		var custom []DNSEntry
		for _, r := range records {
			e := DNSEntry{Name: r.Name, IP: r.IP, IP6: r.IP6, Custom: true, Live: true}
			if n, ok := byID[r.NodeID]; ok {
				e.IP, e.IP6, e.Node, e.Live = n.MeshIP, n.MeshIP6, n.Name, live(n)
			}
			custom = append(custom, e)
		}
		sort.Slice(v.Records, func(i, j int) bool { return v.Records[i].Name < v.Records[j].Name })
		sort.Strings(v.Nodes)
		v.Records = append(v.Records, custom...) // DNSRecords is by name
		for _, u := range in.Upstream {
			v.Upstream = append(v.Upstream, u.String())
		}
		names, err := LoadNames(tx, s.serverURL)
		v.Names = names.Entries
		return err
	})
	return v, err
}

// AddRecord adds a custom record.
func (s *Nodes) AddRecord(ctx context.Context, by Actor, req RecordRequest) (MeshChange, error) {
	name, err := validate.DNSName("name", req.Name)
	if err != nil {
		return MeshChange{}, invalid(err)
	}
	if (req.Node == "") == (req.IP == "" && req.IP6 == "") {
		return MeshChange{}, errorf(http.StatusBadRequest, proto.CodeBadRequest,
			"record %q: give either a node to follow or a fixed IP (IPv4, IPv6 or both), not both or neither", name)
	}
	rec := store.DNSRecord{Name: name}
	detail := map[string]any{}
	if req.Node != "" {
		if _, err := validate.Name("node", req.Node); err != nil {
			return MeshChange{}, invalid(err)
		}
		detail["node"] = req.Node
	}
	if req.IP != "" {
		ip, err := validate.IPv4("ip", req.IP)
		if err != nil {
			return MeshChange{}, invalid(err)
		}
		rec.IP, detail["ip"] = ip.String(), ip.String()
	}
	if req.IP6 != "" {
		ip, err := validate.IPv6("ip6", req.IP6)
		if err != nil {
			return MeshChange{}, invalid(err)
		}
		rec.IP6, detail["ip6"] = ip.String(), ip.String()
	}
	res, err := s.apply(ctx, Change{Actor: by.Name, UserID: by.UserID, Action: "dns.add", Target: name, Detail: detail},
		func(tx *store.MeshTx, _ *Change) error {
			if req.Node != "" {
				n, err := nodeByName(tx, req.Node, "node")
				if err != nil {
					return err
				}
				rec.NodeID = n.ID
			}
			err := tx.InsertDNSRecord(&rec)
			if errors.Is(err, store.ErrConflict) {
				return errorf(http.StatusConflict, proto.CodeConflict, "a DNS record named %q exists already; remove it first", name)
			}
			return err
		})
	return meshChange(res, err)
}

// RemoveRecord removes a custom record by name.
func (s *Nodes) RemoveRecord(ctx context.Context, by Actor, name string) (MeshChange, error) {
	name, err := validate.DNSName("name", name)
	if err != nil {
		return MeshChange{}, invalid(err)
	}
	res, err := s.apply(ctx, Change{Actor: by.Name, UserID: by.UserID, Action: "dns.rm", Target: name},
		func(tx *store.MeshTx, _ *Change) error {
			records, err := tx.DNSRecords()
			if err != nil {
				return err
			}
			for _, r := range records {
				if r.Name == name {
					return tx.DeleteDNSRecord(r.ID)
				}
			}
			return errorf(http.StatusNotFound, proto.CodeNotFound, "there is no custom DNS record %q; node names cannot be removed", name)
		})
	return meshChange(res, err)
}

func meshChange(res Result, err error) (MeshChange, error) {
	if err != nil {
		return MeshChange{}, err
	}
	return MeshChange{Version: res.Version, Warnings: problemStrings(res.Warnings)}, nil
}
