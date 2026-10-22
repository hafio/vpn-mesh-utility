package api

import (
	"cmp"
	"context"
	"fmt"
	"log/slog"
	"net/netip"
	"slices"
	"strconv"
	"strings"

	"github.com/hafio/vpn-mesh-utility/internal/proto"
	"github.com/hafio/vpn-mesh-utility/internal/server/planner"
	"github.com/hafio/vpn-mesh-utility/internal/server/store"
)

// Change describes one write, for the version history and the audit log.
type Change struct {
	// Actor is who made the change: "cli", "node:<id>", "system".
	Actor  string
	UserID int64
	// Action is the audit verb ("node.add"); empty writes no audit row.
	Action string
	Target string
	// Detail is stored in the audit row as JSON. Never put a secret here.
	Detail any
	// Summary describes the change in the version history; empty means
	// "<Action> <Target>".
	Summary string
	// RolledBackFrom is the version whose snapshot this change restores
	// (Rollback sets it); 0 for an ordinary change.
	RolledBackFrom int64
}

// Result is what a write did to the computed configs.
type Result struct {
	// Version is the new config version; 0 when no node's config changed.
	Version int64
	// Changed lists the nodes whose config changed, by id.
	Changed []int64
	// Warnings are the plan's non-blocking problems.
	Warnings []planner.Problem
}

// PlanError means the change would make the mesh plan invalid. Nothing was
// written.
type PlanError struct {
	Problems []planner.Problem
}

func (e *PlanError) Error() string {
	lines := make([]string, len(e.Problems))
	for i, p := range e.Problems {
		lines[i] = p.String()
	}
	return "the change was not applied because the mesh plan would have errors:\n  " + strings.Join(lines, "\n  ")
}

// Reconciler is the single write path for desired state. Apply runs a
// mutation and recomputes every node's config in the same transaction,
// writes a config version only when some node's computed config changed
// (amendment A16), audits the change, and after commit pushes the new
// version to exactly the nodes whose config changed.
type Reconciler struct {
	db  DB
	hub *Hub
	log *slog.Logger
}

// NewReconciler wires a reconciler; log may be nil.
func NewReconciler(db DB, hub *Hub, log *slog.Logger) *Reconciler {
	if log == nil {
		log = slog.New(slog.DiscardHandler)
	}
	return &Reconciler{db: db, hub: hub, log: log}
}

// Apply runs mutate (nil recomputes only) and reconciles. mutate may fill in
// c's Target, Detail and Summary once it knows them. A mutation error or a
// plan with errors rolls everything back, so a half-valid plan never lands.
func (r *Reconciler) Apply(ctx context.Context, c Change, mutate func(*store.Tx, *Change) error) (Result, error) {
	var (
		res    Result
		events map[int64]proto.ConfigEvent
	)
	err := r.db.Tx(ctx, func(tx *store.Tx) error {
		res, events = Result{}, map[int64]proto.ConfigEvent{}
		if mutate != nil {
			if err := mutate(tx, &c); err != nil {
				return err
			}
		}
		in, nodes, err := LoadInput(tx)
		if err != nil {
			return err
		}
		configs, problems, err := planner.Compute(in)
		var blocking []planner.Problem
		for _, p := range problems {
			if p.Warning {
				res.Warnings = append(res.Warnings, p)
			} else {
				blocking = append(blocking, p)
			}
		}
		if len(blocking) > 0 {
			return &PlanError{Problems: blocking}
		}
		if err != nil {
			return err
		}

		type changed struct {
			id   int64
			hash string
		}
		var diff []changed
		for _, n := range nodes {
			hash := ""
			if cfg, ok := configs[n.ID]; ok {
				if hash, err = cfg.Hash(); err != nil {
					return err
				}
			}
			if hash != n.ConfigHash {
				diff = append(diff, changed{n.ID, hash})
			}
		}
		if len(diff) > 0 {
			snap, err := tx.Snapshot()
			if err != nil {
				return err
			}
			b, err := snap.Encode()
			if err != nil {
				return err
			}
			summary := c.Summary
			if summary == "" {
				summary = strings.TrimSpace(c.Action + " " + c.Target)
			}
			v, err := tx.InsertVersion(store.Version{Actor: c.Actor, UserID: c.UserID, Summary: summary, Snapshot: b,
				RolledBackFrom: c.RolledBackFrom})
			if err != nil {
				return err
			}
			res.Version = v
			for _, d := range diff {
				ver := v
				if d.hash == "" {
					ver = 0 // the node no longer gets a config (disabled, unenrolled)
				}
				if err := tx.SetConfigState(d.id, d.hash, ver); err != nil {
					return err
				}
				res.Changed = append(res.Changed, d.id)
				if d.hash != "" {
					events[d.id] = proto.ConfigEvent{Version: v, Hash: d.hash}
				}
			}
		}
		if c.Action != "" {
			return tx.Audit(c.Actor, c.UserID, c.Action, c.Target, c.Detail)
		}
		return nil
	})
	if err != nil {
		return Result{}, err
	}
	r.hub.Notify(events)
	if res.Version != 0 {
		r.log.Info("new config version", "version", res.Version, "actor", c.Actor, "action", c.Action,
			"target", c.Target, "nodes_changed", len(res.Changed))
	}
	return res, nil
}

// Rollback restores the desired state recorded in version n, as a new
// version: history is never rewritten (spec S8.3). Rolling back to a version
// equal to the current state writes nothing (Result.Version 0).
func (r *Reconciler) Rollback(ctx context.Context, c Change, n int64) (Result, error) {
	c.RolledBackFrom = n
	if c.Action == "" {
		c.Action = "version.rollback"
	}
	c.Target = fmt.Sprintf("v%d", n)
	if c.Summary == "" {
		c.Summary = fmt.Sprintf("rollback to v%d", n)
	}
	return r.Apply(ctx, c, func(tx *store.Tx, _ *Change) error {
		v, err := tx.VersionByNumber(n)
		if err != nil {
			return err
		}
		snap, err := store.DecodeSnapshot(v.Snapshot)
		if err != nil {
			return err
		}
		return tx.Restore(snap)
	})
}

// LoadInput reads the desired state as the planner's input. Stored values
// were validated on the way in; one that still does not parse becomes a zero
// value, which planner.Check reports against its node.
func LoadInput(tx *store.Tx) (planner.Input, []store.Node, error) {
	settings, err := tx.Settings()
	if err != nil {
		return planner.Input{}, nil, err
	}
	mesh, err := netip.ParsePrefix(settings[store.SettingMeshCIDR])
	if err != nil {
		return planner.Input{}, nil, fmt.Errorf("setting %s = %q is not a CIDR: %w", store.SettingMeshCIDR, settings[store.SettingMeshCIDR], err)
	}
	in := planner.Input{MeshCIDR: mesh, Zone: settings[store.SettingDNSZone], Domains: store.SplitDomains(settings[store.SettingDNSExtraDomains])}
	if s := settings[store.SettingMeshCIDR6]; s != "" {
		if in.MeshCIDR6, err = netip.ParsePrefix(s); err != nil {
			return planner.Input{}, nil, fmt.Errorf("setting %s = %q is not a CIDR: %w", store.SettingMeshCIDR6, s, err)
		}
	}
	if s := settings[store.SettingDefaultHub]; s != "" {
		if in.DefaultHub, err = strconv.ParseInt(s, 10, 64); err != nil {
			return planner.Input{}, nil, fmt.Errorf("setting %s = %q is not a node id: %w", store.SettingDefaultHub, s, err)
		}
	}
	for _, f := range strings.Split(settings[store.SettingUpstreamDNS], ",") {
		if f = strings.TrimSpace(f); f != "" {
			a, _ := netip.ParseAddr(f) // the zero value is reported by Check
			in.Upstream = append(in.Upstream, a)
		}
	}
	nodes, err := tx.Nodes()
	if err != nil {
		return planner.Input{}, nil, err
	}
	for _, n := range nodes {
		in.Nodes = append(in.Nodes, plannerNode(n))
	}
	records, err := tx.DNSRecords()
	if err != nil {
		return planner.Input{}, nil, err
	}
	for _, r := range records {
		pr := planner.Record{Name: r.Name, Node: r.NodeID}
		if r.IP != "" {
			pr.IP, _ = netip.ParseAddr(r.IP)
		}
		if r.IP6 != "" {
			pr.IP6, _ = netip.ParseAddr(r.IP6)
		}
		in.Records = append(in.Records, pr)
	}
	seg, err := tx.Segmentation()
	if err != nil {
		return planner.Input{}, nil, err
	}
	in.Segmentation = plannerSegmentation(seg)
	in.Segmentation.TimeZone = PolicyZone(seg, settings)
	return in, nodes, nil
}

// PolicyZone is the time zone the policies' schedules are read in: the
// document's, else the server's (written at every start), else UTC.
func PolicyZone(seg store.Segmentation, settings map[string]string) string {
	return cmp.Or(seg.TimeZone, settings[store.SettingServerTZ], "UTC")
}

// plannerSegmentation is the stored segmentation as the planner takes it.
func plannerSegmentation(s store.Segmentation) planner.Segmentation {
	ps := planner.Segmentation{Groups: map[string][]string{}, Addresses: map[string][]string{}, Services: map[string][]string{}}
	for _, g := range s.Groups {
		ps.Groups[g.Name] = g.Nodes
	}
	for _, o := range s.Addresses {
		ps.Addresses[o.Name] = o.Entries
	}
	for _, o := range s.Services {
		ps.Services[o.Name] = o.Entries
	}
	for _, z := range s.Zones {
		ps.Zones = append(ps.Zones, planner.SegZone{Name: z.Name, Intra: z.Intra, Members: z.Members})
	}
	for _, p := range s.Policies {
		sp := planner.SegPolicy{Name: p.Name, Disabled: p.Disabled, From: p.From, To: p.To, Services: p.Services, Action: p.Action}
		if w := p.Schedule; w != nil {
			sp.Schedule = &proto.Schedule{Days: w.Days, Start: w.Start, End: w.End}
		}
		ps.Policies = append(ps.Policies, sp)
	}
	return ps
}

func plannerNode(n store.Node) planner.Node {
	p := planner.Node{
		ID: n.ID, Name: n.Name, Role: proto.Role(n.Role), WGPubKey: n.WGPubKey, ListenPort: n.ListenPort,
		PrimaryHub: n.PrimaryHubID, CanExit: n.CanExit, ExitVia: n.ExitVia, Keepalive: n.Keepalive,
		MTU: n.MTU, Enrolled: n.Enrolled, Disabled: n.Disabled,
	}
	if ip, err := netip.ParseAddr(n.MeshIP); err == nil {
		p.MeshIP = ip
	}
	if ap, err := netip.ParseAddrPort(n.Endpoint); err == nil {
		p.Endpoint = ap
	}
	p.Mode = n.Mode
	if k, err := netip.ParsePrefix(n.ContainerSubnet); err == nil {
		p.Containers = k
	}
	// The IPv6 side (M8); a value that does not parse is the zero value,
	// which Check reports or treats as none.
	p.Family = n.Family
	p.MeshIP6, _ = netip.ParseAddr(n.MeshIP6)
	p.Endpoint6, _ = netip.ParseAddrPort(n.Endpoint6)
	p.Containers6, _ = netip.ParsePrefix(n.ContainerSubnet6)
	p.Down = n.Role == string(proto.RoleServer) && n.DownSince > 0
	// A family the agent reported its host cannot use; "" is not said.
	if n.HostFamilies != "" {
		fams := strings.Split(n.HostFamilies, ",")
		p.NoIPv4, p.NoIPv6 = !slices.Contains(fams, proto.FamilyIPv4), !slices.Contains(fams, proto.FamilyIPv6)
	}
	return p
}

// configFor computes one node's current config, stamped with its stored
// config version. cfg is nil when the node gets no config (not enrolled,
// disabled).
func configFor(tx *store.Tx, id int64) (store.Node, *proto.NodeConfig, error) {
	in, nodes, err := LoadInput(tx)
	if err != nil {
		return store.Node{}, nil, err
	}
	var self store.Node
	found := false
	for _, n := range nodes {
		if n.ID == id {
			self, found = n, true
		}
	}
	if !found {
		return store.Node{}, nil, fmt.Errorf("node %d: %w", id, store.ErrNotFound)
	}
	configs, _, err := planner.Compute(in)
	if err != nil {
		return self, nil, err
	}
	cfg, ok := configs[id]
	if !ok {
		return self, nil, nil
	}
	cfg.Version = self.ConfigVersion
	return self, &cfg, nil
}
