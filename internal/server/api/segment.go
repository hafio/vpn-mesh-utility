package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/netip"
	"net/url"
	"slices"
	"time"

	"github.com/hafio/vpn-mesh-utility/internal/proto"
	"github.com/hafio/vpn-mesh-utility/internal/server/planner"
	"github.com/hafio/vpn-mesh-utility/internal/server/store"
	"github.com/hafio/vpn-mesh-utility/internal/validate"
)

// The mesh firewall (amendment A43): groups, address and service objects,
// zones and the ordered policies every server enforces. It is one document
// (store.Segmentation); the admin API replaces it whole, the web UI edits
// one item at a time. Every change goes through the reconciler, so it is
// checked against the mesh, versioned, audited and pushed to the servers.

// AdminSegmentation is the admin API's mesh firewall.
const AdminSegmentation = "/admin/v1/segmentation"

// ClockFormat is how the views show a time in the policies' zone.
const ClockFormat = "Mon 2006-01-02 15:04 MST"

// SegmentView is GET /admin/v1/segmentation: the document and what it
// compiles to.
type SegmentView struct {
	Config store.Segmentation `json:"config"`
	// Rules are what every server enforces, in order: the system rules, the
	// enabled policies, then each zone's allow within it; what matches none
	// is dropped. The indexes in Zones point into it.
	Rules []proto.PolicyRule `json:"rules"`
	Zones []ZoneView         `json:"zones"`
	// Nodes are every node's name, for the forms.
	Nodes    []string `json:"nodes"`
	Warnings []string `json:"warnings,omitempty"`
	// TimeZone is the zone the schedules are read in (Config.TimeZone, else
	// the server's) and Now the server's time there (ClockFormat);
	// Inactive are the indexes of the Rules whose schedule does not hold
	// now.
	TimeZone string `json:"timezone"`
	Now      string `json:"now"`
	Inactive []int  `json:"inactive"`
}

// ZoneView is a zone as the policy sees it: its addresses, and the rules
// that reach it from outside (Inbound: their To overlaps it) and that it
// reaches out with (Outbound: their From overlaps it) -- mesh traffic only.
type ZoneView struct {
	Name      string   `json:"name"`
	Addresses []string `json:"addresses"`
	Inbound   []int    `json:"inbound"`
	Outbound  []int    `json:"outbound"`
}

// ReachRequest asks what the policy does with one flow. From and To are
// node names or IPv4 addresses; Service is tcp/PORT, udp/PORT or icmp.
type ReachRequest struct {
	From    string `json:"from"`
	To      string `json:"to"`
	Service string `json:"service"`
	// At is when, in RFC 3339 form; empty is now. Scheduled rules apply
	// only in their window.
	At string `json:"at,omitempty"`
}

// ReachResult is the verdict: allowed or dropped, by which rule.
type ReachResult struct {
	From    string `json:"from"`
	To      string `json:"to"`
	Service string `json:"service"`
	Allowed bool   `json:"allowed"`
	// Rule says what decided it; Index is the rule's place in
	// SegmentView.Rules, -1 for none.
	Rule  string `json:"rule"`
	Index int    `json:"index"`
	// At is the time judged, in the policies' zone (RFC 3339); Inactive
	// names the scheduled rules that did not apply then.
	At       string   `json:"at"`
	Inactive []string `json:"inactive,omitempty"`
}

// Matrix is how every zone reaches every other (Cells[from][to]).
type Matrix struct {
	Zones []string           `json:"zones"`
	Rules []proto.PolicyRule `json:"rules"`
	Cells [][]MatrixCell     `json:"cells"`
	// At is when the cells hold (ClockFormat): scheduled rules count only
	// in their window.
	At string `json:"at"`
}

// MatrixCell is one zone pair: planner.ReachAllow, ReachDeny or
// ReachPartial, and the indexes of the rules that decide it.
type MatrixCell struct {
	Verdict string `json:"verdict"`
	Rules   []int  `json:"rules,omitempty"`
}

// zoneOf is a time zone by name; one that no longer loads (a stored value
// from another build) reads as UTC, and planner.Check says what is wrong.
func zoneOf(name string) *time.Location {
	loc, err := validate.TimeZone("timezone", name)
	if err != nil {
		return time.UTC
	}
	return loc
}

// activeAt is which of pol's rules apply at at; nil (every rule) when pol
// does not parse, which planner.Check reports.
func activeAt(pol *proto.Policy, at time.Time) []bool {
	m, err := pol.Matcher()
	if err != nil {
		return nil
	}
	return m.Active(at)
}

// Segment returns the mesh firewall.
func (s *Nodes) Segment(ctx context.Context) (SegmentView, error) {
	v := SegmentView{Rules: []proto.PolicyRule{}, Zones: []ZoneView{}, Nodes: []string{}, Inactive: []int{}}
	err := s.db.Tx(ctx, func(tx *store.Tx) error {
		in, nodes, err := LoadInput(tx)
		if err != nil {
			return err
		}
		if v.Config, err = tx.Segmentation(); err != nil {
			return err
		}
		pol := planner.CompilePolicy(in)
		v.Rules = append(v.Rules, pol.Rules...)
		now := s.now().In(zoneOf(in.Segmentation.TimeZone))
		v.TimeZone, v.Now = in.Segmentation.TimeZone, now.Format(ClockFormat)
		for i, on := range activeAt(pol, now) {
			if !on {
				v.Inactive = append(v.Inactive, i)
			}
		}
		zones := planner.ZoneAddresses(in)
		for _, z := range v.Config.Zones {
			zv := ZoneView{Name: z.Name, Addresses: zones[z.Name], Inbound: []int{}, Outbound: []int{}}
			for i, r := range pol.Rules {
				if planner.Overlaps(r.To, zv.Addresses) {
					zv.Inbound = append(zv.Inbound, i)
				}
				if planner.Overlaps(r.From, zv.Addresses) {
					zv.Outbound = append(zv.Outbound, i)
				}
			}
			v.Zones = append(v.Zones, zv)
		}
		for _, n := range nodes {
			v.Nodes = append(v.Nodes, n.Name)
		}
		slices.Sort(v.Nodes)
		v.Warnings = problemStrings(planner.Check(in))
		return nil
	})
	return v, err
}

// SetSegment replaces the whole document.
func (s *Nodes) SetSegment(ctx context.Context, by Actor, doc store.Segmentation) (MeshChange, error) {
	if err := doc.Validate(); err != nil {
		return MeshChange{}, invalid(err)
	}
	res, err := s.rec.Apply(ctx, Change{Actor: by.Name, UserID: by.UserID, Action: "segmentation.set", Target: "mesh firewall",
		Detail: map[string]any{"groups": len(doc.Groups), "addresses": len(doc.Addresses), "services": len(doc.Services),
			"zones": len(doc.Zones), "policies": len(doc.Policies)}},
		func(tx *store.Tx, _ *Change) error { return setSegmentation(tx, doc) })
	return meshChange(res, err)
}

// EditSegment changes the document with edit in one transaction, so two
// editors never overwrite each other's change; action and target are the
// audit's.
func (s *Nodes) EditSegment(ctx context.Context, by Actor, action, target string, edit func(*store.Segmentation) error) (MeshChange, error) {
	res, err := s.rec.Apply(ctx, Change{Actor: by.Name, UserID: by.UserID, Action: action, Target: target},
		func(tx *store.Tx, _ *Change) error {
			seg, err := tx.Segmentation()
			if err != nil {
				return err
			}
			if err := edit(&seg); err != nil {
				return err
			}
			if err := seg.Validate(); err != nil {
				return invalid(err)
			}
			return setSegmentation(tx, seg)
		})
	return meshChange(res, err)
}

func setSegmentation(tx *store.Tx, seg store.Segmentation) error {
	err := tx.SetSegmentation(seg)
	if errors.Is(err, store.ErrNotFound) {
		return errorf(http.StatusBadRequest, proto.CodeBadRequest, "%v: a group may only list nodes of this mesh", err)
	}
	return err
}

// Segmentation item kinds, for the web UI's edits.
const (
	SegGroup   = "group"
	SegAddress = "address"
	SegService = "service"
	SegZone    = "zone"
	SegPolicy  = "policy"
)

// refKind is the reference kind that names an item of kind, "" for one no
// reference names (a policy).
func refKind(kind string) string {
	return map[string]string{SegGroup: store.RefGroup, SegAddress: store.RefAddr, SegService: store.RefService, SegZone: store.RefZone}[kind]
}

// SaveGroup creates a group (old "") or replaces group old, renaming its
// references when the name changes.
func (s *Nodes) SaveGroup(ctx context.Context, by Actor, old string, g store.SegGroup) (MeshChange, error) {
	return s.EditSegment(ctx, by, "segmentation.group", g.Name, func(seg *store.Segmentation) error {
		i, err := place(old, len(seg.Groups), func(i int) string { return seg.Groups[i].Name }, SegGroup)
		if err != nil {
			return err
		}
		seg.Groups = saveAt(seg.Groups, i, g)
		rename(seg, SegGroup, old, g.Name)
		return nil
	})
}

// SaveObject creates or replaces an address or service object.
func (s *Nodes) SaveObject(ctx context.Context, by Actor, kind, old string, o store.SegObject) (MeshChange, error) {
	if kind != SegAddress && kind != SegService {
		return MeshChange{}, errorf(http.StatusBadRequest, proto.CodeBadRequest, "object kind %q: use address or service", kind)
	}
	return s.EditSegment(ctx, by, "segmentation."+kind, o.Name, func(seg *store.Segmentation) error {
		objs := &seg.Addresses
		if kind == SegService {
			objs = &seg.Services
		}
		i, err := place(old, len(*objs), func(i int) string { return (*objs)[i].Name }, kind+" object")
		if err != nil {
			return err
		}
		*objs = saveAt(*objs, i, o)
		rename(seg, kind, old, o.Name)
		return nil
	})
}

// SaveZone creates or replaces a zone.
func (s *Nodes) SaveZone(ctx context.Context, by Actor, old string, z store.SegZone) (MeshChange, error) {
	return s.EditSegment(ctx, by, "segmentation.zone", z.Name, func(seg *store.Segmentation) error {
		i, err := place(old, len(seg.Zones), func(i int) string { return seg.Zones[i].Name }, SegZone)
		if err != nil {
			return err
		}
		seg.Zones = saveAt(seg.Zones, i, z)
		rename(seg, SegZone, old, z.Name)
		return nil
	})
}

// SavePolicy creates or replaces a policy and puts it at position (1 is
// first); 0 keeps a replaced one where it is and appends a new one.
func (s *Nodes) SavePolicy(ctx context.Context, by Actor, old string, p store.SegPolicy, position int) (MeshChange, error) {
	return s.EditSegment(ctx, by, "segmentation.policy", p.Name, func(seg *store.Segmentation) error {
		i, err := place(old, len(seg.Policies), func(i int) string { return seg.Policies[i].Name }, SegPolicy)
		if err != nil {
			return err
		}
		seg.Policies = saveAt(seg.Policies, i, p)
		if position == 0 {
			return nil
		}
		return movePolicy(seg, p.Name, position)
	})
}

// SetTimeZone sets the time zone the policies' schedules are read in; ""
// goes back to the server's.
func (s *Nodes) SetTimeZone(ctx context.Context, by Actor, tz string) (MeshChange, error) {
	target := tz
	if tz == "" {
		target = "the server's"
	}
	return s.EditSegment(ctx, by, "segmentation.timezone", target, func(seg *store.Segmentation) error {
		seg.TimeZone = tz
		return nil
	})
}

// MovePolicy puts a policy at position (1 is first).
func (s *Nodes) MovePolicy(ctx context.Context, by Actor, name string, position int) (MeshChange, error) {
	return s.EditSegment(ctx, by, "segmentation.move", name, func(seg *store.Segmentation) error {
		return movePolicy(seg, name, position)
	})
}

// RemoveSegment removes one item. One still named elsewhere stays: the
// plan check refuses the change and says where it is used.
func (s *Nodes) RemoveSegment(ctx context.Context, by Actor, kind, name string) (MeshChange, error) {
	return s.EditSegment(ctx, by, "segmentation.rm", kind+" "+name, func(seg *store.Segmentation) error {
		var ok bool
		switch kind {
		case SegGroup:
			seg.Groups, ok = without(seg.Groups, func(g store.SegGroup) bool { return g.Name == name })
		case SegAddress:
			seg.Addresses, ok = without(seg.Addresses, func(o store.SegObject) bool { return o.Name == name })
		case SegService:
			seg.Services, ok = without(seg.Services, func(o store.SegObject) bool { return o.Name == name })
		case SegZone:
			seg.Zones, ok = without(seg.Zones, func(z store.SegZone) bool { return z.Name == name })
		case SegPolicy:
			seg.Policies, ok = without(seg.Policies, func(p store.SegPolicy) bool { return p.Name == name })
		default:
			return errorf(http.StatusBadRequest, proto.CodeBadRequest, "kind %q: use group, address, service, zone or policy", kind)
		}
		if !ok {
			return errorf(http.StatusNotFound, proto.CodeNotFound, "there is no %s %q", kind, name)
		}
		return nil
	})
}

// place is where an item saved over old goes: old's index, or -1 for a new
// one; old must exist.
func place(old string, n int, name func(int) string, what string) (int, error) {
	if old == "" {
		return -1, nil
	}
	for i := 0; i < n; i++ {
		if name(i) == old {
			return i, nil
		}
	}
	return 0, errorf(http.StatusNotFound, proto.CodeNotFound, "there is no %s %q; it may have been removed meanwhile", what, old)
}

func saveAt[T any](list []T, i int, v T) []T {
	if i < 0 {
		return append(list, v)
	}
	list[i] = v
	return list
}

func without[T any](list []T, match func(T) bool) ([]T, bool) {
	out := slices.DeleteFunc(slices.Clone(list), match)
	return out, len(out) != len(list)
}

func movePolicy(seg *store.Segmentation, name string, position int) error {
	i := slices.IndexFunc(seg.Policies, func(p store.SegPolicy) bool { return p.Name == name })
	if i < 0 {
		return errorf(http.StatusNotFound, proto.CodeNotFound, "there is no policy %q", name)
	}
	if position < 1 || position > len(seg.Policies) {
		return errorf(http.StatusBadRequest, proto.CodeBadRequest, "position %d: there are %d policies, so give 1 to %d", position,
			len(seg.Policies), len(seg.Policies))
	}
	p := seg.Policies[i]
	seg.Policies = slices.Insert(slices.Delete(seg.Policies, i, i+1), position-1, p)
	return nil
}

// rename points the references to a renamed item at its new name.
func rename(seg *store.Segmentation, kind, old, name string) {
	ref := refKind(kind)
	if old == "" || old == name || ref == "" {
		return
	}
	from, to := ref+":"+old, ref+":"+name
	fix := func(vs []string) {
		for i, v := range vs {
			if v == from {
				vs[i] = to
			}
		}
	}
	for _, z := range seg.Zones {
		fix(z.Members)
	}
	for _, p := range seg.Policies {
		fix(p.From)
		fix(p.To)
		fix(p.Services)
	}
}

// Reach says what the policy does with one flow.
func (s *Nodes) Reach(ctx context.Context, req ReachRequest) (ReachResult, error) {
	svc, err := validate.ServiceEntry("service", req.Service)
	if err == nil && svc.Proto != validate.ServiceICMP && (svc.First == 0 || svc.First != svc.Last) {
		err = &validate.Error{Field: "service", Value: req.Service, Reason: "check one port: tcp/PORT or udp/PORT"}
	}
	if err != nil {
		return ReachResult{}, invalid(err)
	}
	var at time.Time
	if req.At != "" {
		if at, err = validate.Timestamp("at", req.At); err != nil {
			return ReachResult{}, invalid(err)
		}
	}
	out := ReachResult{Service: svc.String(), Index: -1}
	err = s.db.Tx(ctx, func(tx *store.Tx) error {
		in, nodes, err := LoadInput(tx)
		if err != nil {
			return err
		}
		src, err := flowEnd("from", req.From, nodes)
		if err != nil {
			return err
		}
		dst, err := flowEnd("to", req.To, nodes)
		if err != nil {
			return err
		}
		out.From, out.To = src.String(), dst.String()
		pol := planner.CompilePolicy(in)
		m, err := pol.Matcher()
		if err != nil {
			return err
		}
		if at.IsZero() {
			at = s.now()
		}
		at = at.In(zoneOf(in.Segmentation.TimeZone))
		out.At = at.Format(time.RFC3339)
		for j, on := range m.Active(at) {
			if !on {
				out.Inactive = append(out.Inactive, pol.Rules[j].Name)
			}
		}
		var i int
		out.Allowed, i = m.Decide(src, dst, svc.Proto, svc.First, at)
		switch i {
		case proto.RuleDefault:
			out.Rule = "no rule matched: dropped"
		case proto.RuleNotMesh:
			out.Rule = "not mesh traffic (an end is outside the mesh): the mesh firewall does not filter it"
		default:
			out.Rule, out.Index = pol.Rules[i].Name, i
		}
		return nil
	})
	return out, err
}

// flowEnd is a node's mesh IP, or an IPv4 address.
func flowEnd(field, v string, nodes []store.Node) (netip.Addr, error) {
	for _, n := range nodes {
		if n.Name == v {
			a, err := netip.ParseAddr(n.MeshIP)
			if err != nil {
				return netip.Addr{}, fmt.Errorf("node %q has no mesh IP: %w", v, err)
			}
			return a, nil
		}
	}
	a, err := validate.IPv4(field, v)
	if err != nil {
		return netip.Addr{}, invalid(&validate.Error{Field: field, Value: v, Reason: "must be a node of this mesh or an IPv4 address"})
	}
	return a, nil
}

// Matrix is how every zone reaches every other.
func (s *Nodes) Matrix(ctx context.Context) (Matrix, error) {
	out := Matrix{Zones: []string{}, Cells: [][]MatrixCell{}}
	err := s.db.Tx(ctx, func(tx *store.Tx) error {
		in, _, err := LoadInput(tx)
		if err != nil {
			return err
		}
		pol := planner.CompilePolicy(in)
		out.Rules = pol.Rules
		now := s.now().In(zoneOf(in.Segmentation.TimeZone))
		out.At = now.Format(ClockFormat)
		active := activeAt(pol, now)
		zones := planner.ZoneAddresses(in)
		for _, z := range in.Segmentation.Zones {
			out.Zones = append(out.Zones, z.Name)
		}
		slices.Sort(out.Zones)
		for _, a := range out.Zones {
			row := make([]MatrixCell, 0, len(out.Zones))
			for _, b := range out.Zones {
				v, rules := planner.Summarize(pol, active, zones[a], zones[b])
				row = append(row, MatrixCell{Verdict: v, Rules: rules})
			}
			out.Cells = append(out.Cells, row)
		}
		return nil
	})
	return out, err
}

func (a *api) adminSegment(w http.ResponseWriter, r *http.Request) {
	out, err := a.nodes.Segment(r.Context())
	a.reply(w, r, out, err)
}

func (a *api) adminSetSegment(w http.ResponseWriter, r *http.Request) {
	var doc store.Segmentation
	if err := a.readJSON(w, r, &doc); err != nil {
		a.fail(w, r, err)
		return
	}
	out, err := a.nodes.SetSegment(r.Context(), cli(), doc)
	a.reply(w, r, out, err)
}

func (a *api) adminReach(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	out, err := a.nodes.Reach(r.Context(), ReachRequest{From: q.Get("from"), To: q.Get("to"), Service: q.Get("service"), At: q.Get("at")})
	a.reply(w, r, out, err)
}

// Segment returns the mesh firewall.
func (c *AdminClient) Segment(ctx context.Context) (SegmentView, error) {
	var out SegmentView
	err := c.do(ctx, http.MethodGet, AdminSegmentation, nil, &out)
	return out, err
}

// SetSegment replaces the mesh firewall's document.
func (c *AdminClient) SetSegment(ctx context.Context, doc store.Segmentation) (MeshChange, error) {
	var out MeshChange
	err := c.do(ctx, http.MethodPut, AdminSegmentation, doc, &out)
	return out, err
}

// Reach asks what the mesh firewall does with one flow.
func (c *AdminClient) Reach(ctx context.Context, req ReachRequest) (ReachResult, error) {
	q := url.Values{"from": {req.From}, "to": {req.To}, "service": {req.Service}}
	if req.At != "" {
		q.Set("at", req.At)
	}
	var out ReachResult
	err := c.do(ctx, http.MethodGet, AdminSegmentation+"/check?"+q.Encode(), nil, &out)
	return out, err
}
