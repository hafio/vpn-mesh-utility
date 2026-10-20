package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strconv"

	"github.com/hafio/vpn-mesh-utility/internal/proto"
	"github.com/hafio/vpn-mesh-utility/internal/server/planner"
	"github.com/hafio/vpn-mesh-utility/internal/server/store"
	"github.com/hafio/vpn-mesh-utility/internal/validate"
)

// The custom DNS records and firewall rules (spec S8.3 "DNS" and "Firewall"
// pages). Like every other change they go through the reconciler, so they
// are versioned, audited and pushed to the nodes whose config they change.

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
}

// DNSEntry is one name of the zone.
type DNSEntry struct {
	Name string `json:"name"`
	IP   string `json:"ip"`
	// Node is the node the name follows: the node itself for an automatic
	// record, the target of an alias; empty for a fixed IP.
	Node   string `json:"node,omitempty"`
	Custom bool   `json:"custom,omitempty"`
	// Live is false while the name does not resolve: its node is not
	// enrolled yet, or disabled.
	Live bool `json:"live"`
}

// RecordRequest adds a custom record: an alias of Node, or a fixed IP.
type RecordRequest struct {
	Name string `json:"name"`
	Node string `json:"node,omitempty"`
	IP   string `json:"ip,omitempty"`
}

// FirewallView is what the servers' firewalls hold (spec S7.5): every
// rendered rule with where it comes from, and the custom rules.
type FirewallView struct {
	MeshCIDR string           `json:"mesh_cidr"`
	Servers  []ServerFirewall `json:"servers"`
	Custom   []CustomRule     `json:"custom"`
}

// ServerFirewall is one server's rendered firewall section; Firewall is nil
// while the server gets no config (not enrolled, or disabled).
type ServerFirewall struct {
	Name     string          `json:"name"`
	Firewall *proto.Firewall `json:"firewall,omitempty"`
}

// CustomRule is an operator's opening in the mesh zone.
type CustomRule struct {
	ID int64 `json:"id"`
	// Node is the one server it opens; empty means every server.
	Node   string `json:"node,omitempty"`
	Proto  string `json:"proto"`
	Port   int    `json:"port"`
	Source string `json:"source"`
	Note   string `json:"note,omitempty"`
}

// RuleRequest adds a custom rule; an empty Node means every server.
type RuleRequest struct {
	Node   string `json:"node,omitempty"`
	Proto  string `json:"proto"`
	Port   int    `json:"port"`
	Source string `json:"source"`
	Note   string `json:"note,omitempty"`
}

// maxNote bounds a rule's note; the planner adds a "custom: " prefix and the
// agent accepts 200 bytes.
const maxNote = 120

// DNS returns the mesh zone.
func (s *Nodes) DNS(ctx context.Context) (DNSView, error) {
	v := DNSView{Records: []DNSEntry{}, Upstream: []string{}, Domains: []string{}, Nodes: []string{}}
	err := s.db.Tx(ctx, func(tx *store.Tx) error {
		in, nodes, err := LoadInput(tx)
		if err != nil {
			return err
		}
		v.Zone = in.ZoneName()
		v.Domains = append(v.Domains, in.Domains...)
		live := func(n store.Node) bool { return n.Enrolled && !n.Disabled && n.WGPubKey != "" }
		byID := map[int64]store.Node{}
		for _, n := range nodes {
			byID[n.ID] = n
			v.Nodes = append(v.Nodes, n.Name)
			v.Records = append(v.Records, DNSEntry{Name: n.Name, IP: n.MeshIP, Node: n.Name, Live: live(n)})
		}
		records, err := tx.DNSRecords()
		if err != nil {
			return err
		}
		var custom []DNSEntry
		for _, r := range records {
			e := DNSEntry{Name: r.Name, IP: r.IP, Custom: true, Live: true}
			if n, ok := byID[r.NodeID]; ok {
				e.IP, e.Node, e.Live = n.MeshIP, n.Name, live(n)
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
	if (req.Node == "") == (req.IP == "") {
		return MeshChange{}, errorf(http.StatusBadRequest, proto.CodeBadRequest,
			"record %q: give either a node to follow or a fixed IP, not both or neither", name)
	}
	rec := store.DNSRecord{Name: name}
	detail := map[string]any{}
	if req.Node != "" {
		if _, err := validate.Name("node", req.Node); err != nil {
			return MeshChange{}, invalid(err)
		}
		detail["node"] = req.Node
	} else {
		ip, err := validate.IPv4("ip", req.IP)
		if err != nil {
			return MeshChange{}, invalid(err)
		}
		rec.IP, detail["ip"] = ip.String(), ip.String()
	}
	res, err := s.rec.Apply(ctx, Change{Actor: by.Name, UserID: by.UserID, Action: "dns.add", Target: name, Detail: detail},
		func(tx *store.Tx, _ *Change) error {
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
	res, err := s.rec.Apply(ctx, Change{Actor: by.Name, UserID: by.UserID, Action: "dns.rm", Target: name},
		func(tx *store.Tx, _ *Change) error {
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

// Firewall returns every server's rendered firewall and the custom rules.
func (s *Nodes) Firewall(ctx context.Context) (FirewallView, error) {
	v := FirewallView{Servers: []ServerFirewall{}, Custom: []CustomRule{}}
	err := s.db.Tx(ctx, func(tx *store.Tx) error {
		in, nodes, err := LoadInput(tx)
		if err != nil {
			return err
		}
		configs, _, err := planner.Compute(in)
		if err != nil {
			return err
		}
		v.MeshCIDR = in.MeshCIDR.String()
		names := map[int64]string{}
		for _, n := range nodes {
			names[n.ID] = n.Name
			if n.Role != string(proto.RoleServer) {
				continue
			}
			sf := ServerFirewall{Name: n.Name}
			if c, ok := configs[n.ID]; ok {
				sf.Firewall = c.Firewall
			}
			v.Servers = append(v.Servers, sf)
		}
		sort.Slice(v.Servers, func(i, j int) bool { return v.Servers[i].Name < v.Servers[j].Name })
		rules, err := tx.FirewallRules()
		if err != nil {
			return err
		}
		for _, r := range rules {
			v.Custom = append(v.Custom, CustomRule{ID: r.ID, Node: names[r.NodeID], Proto: r.Proto, Port: r.Port, Source: r.SourceCIDR, Note: r.Note})
		}
		return nil
	})
	return v, err
}

// AddRule adds a custom rule.
func (s *Nodes) AddRule(ctx context.Context, by Actor, req RuleRequest) (MeshChange, error) {
	var errs []error
	check := func(err error) {
		if err != nil {
			errs = append(errs, err)
		}
	}
	if req.Node != "" {
		_, err := validate.Name("server", req.Node)
		check(err)
	}
	_, err := validate.Enum("proto", req.Proto, "tcp", "udp")
	check(err)
	_, err = validate.Port("port", req.Port)
	check(err)
	src, err := validate.Prefix4("source", req.Source)
	check(err)
	_, err = validate.Text("note", req.Note, maxNote)
	check(err)
	if err := errors.Join(errs...); err != nil {
		return MeshChange{}, invalid(err)
	}
	rule := store.FirewallRule{Proto: req.Proto, Port: req.Port, SourceCIDR: src.String(), Note: req.Note}
	where := "every server"
	if req.Node != "" {
		where = req.Node
	}
	target := fmt.Sprintf("%s/%d from %s on %s", req.Proto, req.Port, rule.SourceCIDR, where)
	res, err := s.rec.Apply(ctx, Change{Actor: by.Name, UserID: by.UserID, Action: "firewall.add", Target: target,
		Detail: map[string]any{"note": req.Note}},
		func(tx *store.Tx, _ *Change) error {
			if req.Node != "" {
				n, err := nodeByName(tx, req.Node, "server")
				if err != nil {
					return err
				}
				if n.Role != string(proto.RoleServer) {
					return errorf(http.StatusBadRequest, proto.CodeBadRequest,
						"%q is a %s; only servers have a firewall meshctl manages (spokes keep their own)", n.Name, n.Role)
				}
				rule.NodeID = n.ID
			}
			return tx.InsertFirewallRule(&rule)
		})
	return meshChange(res, err)
}

// RemoveRule removes a custom rule by id.
func (s *Nodes) RemoveRule(ctx context.Context, by Actor, id int64) (MeshChange, error) {
	res, err := s.rec.Apply(ctx, Change{Actor: by.Name, UserID: by.UserID, Action: "firewall.rm", Target: "rule " + strconv.FormatInt(id, 10)},
		func(tx *store.Tx, c *Change) error {
			rules, err := tx.FirewallRules()
			if err != nil {
				return err
			}
			for _, r := range rules {
				if r.ID == id {
					c.Detail = map[string]any{"proto": r.Proto, "port": r.Port, "source": r.SourceCIDR, "note": r.Note}
					return tx.DeleteFirewallRule(id)
				}
			}
			return errorf(http.StatusNotFound, proto.CodeNotFound, "there is no custom firewall rule %d; the firewall page lists them", id)
		})
	return meshChange(res, err)
}

func meshChange(res Result, err error) (MeshChange, error) {
	if err != nil {
		return MeshChange{}, err
	}
	return MeshChange{Version: res.Version, Warnings: problemStrings(res.Warnings)}, nil
}
