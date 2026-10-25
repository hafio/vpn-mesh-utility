package store

import (
	"cmp"
	"database/sql"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/hafio/vpn-mesh-utility/internal/validate"
)

// Segmentation (amendment A43): groups of nodes, address and service
// objects, zones and the ordered policies every server enforces -- the mesh
// firewall. The same document is what the admin API takes,
// what an export holds and what a snapshot records; references inside it
// are kind:name strings the planner resolves (and refuses when dangling).
type Segmentation struct {
	Groups    []SegGroup  `json:"groups,omitempty" yaml:"groups,omitempty"`
	Addresses []SegObject `json:"addresses,omitempty" yaml:"addresses,omitempty"`
	Services  []SegObject `json:"services,omitempty" yaml:"services,omitempty"`
	Zones     []SegZone   `json:"zones,omitempty" yaml:"zones,omitempty"`
	// Policies are in order: the first that matches a packet decides.
	Policies []SegPolicy `json:"policies,omitempty" yaml:"policies,omitempty"`
	// TimeZone is the IANA time zone the policies' schedules are read in;
	// empty is the server's (SettingServerTZ).
	TimeZone string `json:"timezone,omitempty" yaml:"timezone,omitempty"`
}

// SegGroup is a set of nodes, by name; a node may be in several.
type SegGroup struct {
	Name        string   `json:"name" yaml:"name"`
	Description string   `json:"description,omitempty" yaml:"description,omitempty"`
	Nodes       []string `json:"nodes,omitempty" yaml:"nodes,omitempty"`
}

// SegObject is a named list of addresses (validate.AddrEntry) or services
// (validate.ServiceEntry).
type SegObject struct {
	Name        string   `json:"name" yaml:"name"`
	Description string   `json:"description,omitempty" yaml:"description,omitempty"`
	Entries     []string `json:"entries" yaml:"entries"`
}

// SegZone is a segment of the mesh: its members are node:, containers:,
// group:, role: and addr: references and literal addresses. Intra is what
// traffic between two of its members gets when no policy matched it first.
type SegZone struct {
	Name        string   `json:"name" yaml:"name"`
	Description string   `json:"description,omitempty" yaml:"description,omitempty"`
	Intra       string   `json:"intra" yaml:"intra"`
	Members     []string `json:"members,omitempty" yaml:"members,omitempty"`
}

// SegPolicy allows or denies traffic From to To for Services. From and To
// take zone: references besides a zone member's kinds; Services takes svc:
// references and literal services. An empty list is any.
type SegPolicy struct {
	Name        string   `json:"name" yaml:"name"`
	Description string   `json:"description,omitempty" yaml:"description,omitempty"`
	Disabled    bool     `json:"disabled,omitempty" yaml:"disabled,omitempty"`
	From        []string `json:"from,omitempty" yaml:"from,omitempty"`
	To          []string `json:"to,omitempty" yaml:"to,omitempty"`
	Services    []string `json:"services,omitempty" yaml:"services,omitempty"`
	Action      string   `json:"action" yaml:"action"`
	// Schedule limits the policy to times of the week; nil is always.
	Schedule *SegSchedule `json:"schedule,omitempty" yaml:"schedule,omitempty"`
}

// SegSchedule is when a policy applies (validate.Schedule), on the wall
// clock of the document's time zone: on Days (every day when empty), from
// Start to End, "HH:MM", an End before Start ending the next day; without
// Start and End, the whole of each day.
type SegSchedule struct {
	Days  []string `json:"days,omitempty" yaml:"days,omitempty"`
	Start string   `json:"start,omitempty" yaml:"start,omitempty"`
	End   string   `json:"end,omitempty" yaml:"end,omitempty"`
}

// String is the schedule as people read it, "mon-fri 09:00-18:00"; one that
// does not parse (Validate says why) reads as it is stored.
func (s SegSchedule) String() string {
	w, err := validate.Schedule("schedule", s.Days, s.Start, s.End)
	if err != nil {
		return strings.TrimSpace(strings.Join(s.Days, ",") + " " + s.Start + "-" + s.End)
	}
	return w.String()
}

// FirewallRule is a custom firewall rule from before segmentation: a port
// opened in the mesh zone of one server (NodeID) or of every server
// (NodeID 0). Snapshots and exports made then still carry them, and
// LegacySegmentation turns them into policies as migration 0006 did.
type FirewallRule struct {
	ID         int64  `json:"id"`
	NodeID     int64  `json:"node_id,omitempty"`
	Proto      string `json:"proto"`
	Port       int    `json:"port"`
	SourceCIDR string `json:"source_cidr"`
	Note       string `json:"note,omitempty"`
}

// DefaultPolicy is the policy every mesh starts with (amendment A43): any
// to any, allowed. It is an ordinary policy, last in the list, to change or
// remove; without it, what no policy allows is dropped.
func DefaultPolicy() SegPolicy {
	return SegPolicy{Name: "allow-all", Action: SegAllow,
		Description: "Every mesh address reaches every other. Change or remove it to segment the mesh."}
}

// LegacySegmentation is what a mesh without segmentation had, as policies:
// each custom rule (by id) an allow from its source to its server (by
// name, names) or to every server, then DefaultPolicy -- what migration
// 0006 makes of the firewall_rules table.
func LegacySegmentation(rules []FirewallRule, names map[int64]string) Segmentation {
	var s Segmentation
	byID := slices.Clone(rules)
	slices.SortFunc(byID, func(a, b FirewallRule) int { return cmp.Compare(a.ID, b.ID) })
	for _, r := range byID {
		to := RefRole + ":server"
		if r.NodeID != 0 {
			to = RefNode + ":" + names[r.NodeID]
		}
		s.Policies = append(s.Policies, SegPolicy{Name: "firewall-" + strconv.FormatInt(r.ID, 10), Description: r.Note,
			From: []string{r.SourceCIDR}, To: []string{to}, Services: []string{r.Proto + "/" + strconv.Itoa(r.Port)}, Action: SegAllow})
	}
	s.Policies = append(s.Policies, DefaultPolicy())
	return s
}

// Segmentation values.
const (
	SegAllow = "allow"
	SegDeny  = "deny"
	// The reference kinds.
	RefNode       = "node"
	RefContainers = "containers"
	RefGroup      = "group"
	RefRole       = "role"
	RefAddr       = "addr"
	RefZone       = "zone"
	RefService    = "svc"
)

// MemberKinds are the references a zone member may be; PolicyKinds those
// of a policy's From and To.
var (
	MemberKinds = []string{RefNode, RefContainers, RefGroup, RefRole, RefAddr}
	PolicyKinds = []string{RefZone, RefNode, RefContainers, RefGroup, RefRole, RefAddr}
)

// Limits on the document, so it stays something a page can show and the
// servers' rulesets stay small.
const (
	MaxSegGroups      = 128
	MaxSegObjects     = 256 // of each kind
	MaxSegEntries     = 64  // per object
	MaxSegZones       = 64
	MaxSegMembers     = 128 // per zone
	MaxSegPolicies    = 256
	MaxSegRefs        = 32 // per From, To or Services
	maxSegDescription = 200
)

// Empty reports a document with nothing in it.
func (s Segmentation) Empty() bool {
	return len(s.Groups)+len(s.Addresses)+len(s.Services)+len(s.Zones)+len(s.Policies) == 0 && s.TimeZone == ""
}

// Validate checks every value against its grammar and the limits: names,
// descriptions, entries, references, actions, schedules and the time zone,
// each name once per kind.
// Whether a reference names something that exists, and whether an address
// is inside the mesh, the planner checks with the rest of the mesh.
func (s Segmentation) Validate() error {
	var errs []error
	bad := func(err error) {
		if err != nil {
			errs = append(errs, err)
		}
	}
	limit := func(field string, n, max int) {
		if n > max {
			bad(fmt.Errorf("%s: %d of them, at most %d", field, n, max))
		}
	}
	names := func(field string, n int, name func(int) string) {
		seen := map[string]bool{}
		for i := 0; i < n; i++ {
			v := name(i)
			if _, err := validate.Name(fmt.Sprintf("%s[%d].name", field, i), v); err != nil {
				bad(err)
				continue
			}
			if seen[v] {
				bad(fmt.Errorf("%s: %q is defined twice", field, v))
			}
			seen[v] = true
		}
	}
	text := func(field, v string) {
		_, err := validate.Text(field, v, maxSegDescription)
		bad(err)
	}
	addr := func(field, v string) {
		_, err := validate.AddrEntry(field, v)
		bad(err)
	}
	refs := func(field string, vs []string, kinds []string) {
		limit(field, len(vs), MaxSegRefs)
		for i, v := range vs {
			f := fmt.Sprintf("%s[%d]", field, i)
			if _, _, ok, err := validate.Ref(f, v, kinds...); ok {
				bad(err)
				continue
			}
			addr(f, v)
		}
	}

	limit("groups", len(s.Groups), MaxSegGroups)
	names("groups", len(s.Groups), func(i int) string { return s.Groups[i].Name })
	for _, g := range s.Groups {
		f := fmt.Sprintf("groups[%s]", g.Name)
		text(f+".description", g.Description)
		seen := map[string]bool{}
		for j, n := range g.Nodes {
			if _, err := validate.Name(fmt.Sprintf("%s.nodes[%d]", f, j), n); err != nil {
				bad(err)
				continue
			}
			if seen[n] {
				bad(fmt.Errorf("%s.nodes: %q is listed twice", f, n))
			}
			seen[n] = true
		}
	}
	for _, kind := range []struct {
		field string
		objs  []SegObject
		entry func(field, v string)
	}{
		{"addresses", s.Addresses, addr},
		{"services", s.Services, func(field, v string) {
			_, err := validate.ServiceEntry(field, v)
			bad(err)
		}},
	} {
		limit(kind.field, len(kind.objs), MaxSegObjects)
		names(kind.field, len(kind.objs), func(i int) string { return kind.objs[i].Name })
		for _, o := range kind.objs {
			f := fmt.Sprintf("%s[%s]", kind.field, o.Name)
			text(f+".description", o.Description)
			if len(o.Entries) == 0 {
				bad(fmt.Errorf("%s.entries: an object needs at least one entry", f))
			}
			limit(f+".entries", len(o.Entries), MaxSegEntries)
			for j, e := range o.Entries {
				kind.entry(fmt.Sprintf("%s.entries[%d]", f, j), e)
			}
		}
	}
	limit("zones", len(s.Zones), MaxSegZones)
	names("zones", len(s.Zones), func(i int) string { return s.Zones[i].Name })
	for _, z := range s.Zones {
		f := fmt.Sprintf("zones[%s]", z.Name)
		text(f+".description", z.Description)
		_, err := validate.Enum(f+".intra", z.Intra, SegAllow, SegDeny)
		bad(err)
		limit(f+".members", len(z.Members), MaxSegMembers)
		for j, m := range z.Members {
			mf := fmt.Sprintf("%s.members[%d]", f, j)
			if _, _, ok, err := validate.Ref(mf, m, MemberKinds...); ok {
				bad(err)
				continue
			}
			addr(mf, m)
		}
	}
	limit("policies", len(s.Policies), MaxSegPolicies)
	names("policies", len(s.Policies), func(i int) string { return s.Policies[i].Name })
	for _, p := range s.Policies {
		f := fmt.Sprintf("policies[%s]", p.Name)
		text(f+".description", p.Description)
		_, err := validate.Enum(f+".action", p.Action, SegAllow, SegDeny)
		bad(err)
		refs(f+".from", p.From, PolicyKinds)
		refs(f+".to", p.To, PolicyKinds)
		limit(f+".services", len(p.Services), MaxSegRefs)
		for j, v := range p.Services {
			sf := fmt.Sprintf("%s.services[%d]", f, j)
			if _, _, ok, err := validate.Ref(sf, v, RefService); ok {
				bad(err)
				continue
			}
			_, err := validate.ServiceEntry(sf, v)
			bad(err)
		}
		if w := p.Schedule; w != nil {
			_, err := validate.Schedule(f+".schedule", w.Days, w.Start, w.End)
			bad(err)
		}
	}
	if s.TimeZone != "" {
		_, err := validate.TimeZone("timezone", s.TimeZone)
		bad(err)
	}
	return errors.Join(errs...)
}

// Segmentation reads the mesh's document: groups (with their nodes),
// objects and zones by name, policies in order.
func (t *MeshTx) Segmentation() (Segmentation, error) {
	var s Segmentation
	members, err := t.groupMembers()
	if err != nil {
		return Segmentation{}, err
	}
	rows, err := t.query(`SELECT id, name, description FROM seg_groups WHERE mesh_id = ? ORDER BY name`, t.mesh)
	if err != nil {
		return Segmentation{}, fmt.Errorf("read groups: %w", err)
	}
	if err := scanAll(rows, "groups", func(r *sql.Rows) error {
		var (
			g  SegGroup
			id int64
		)
		if err := r.Scan(&id, &g.Name, &g.Description); err != nil {
			return err
		}
		g.Nodes = members[id]
		s.Groups = append(s.Groups, g)
		return nil
	}); err != nil {
		return Segmentation{}, err
	}
	rows, err = t.query(`SELECT kind, name, description, entries FROM seg_objects WHERE mesh_id = ? ORDER BY kind, name`, t.mesh)
	if err != nil {
		return Segmentation{}, fmt.Errorf("read objects: %w", err)
	}
	if err := scanAll(rows, "objects", func(r *sql.Rows) error {
		var (
			o             SegObject
			kind, entries string
		)
		if err := r.Scan(&kind, &o.Name, &o.Description, &entries); err != nil {
			return err
		}
		o.Entries = splitList(entries)
		if kind == "service" {
			s.Services = append(s.Services, o)
		} else {
			s.Addresses = append(s.Addresses, o)
		}
		return nil
	}); err != nil {
		return Segmentation{}, err
	}
	rows, err = t.query(`SELECT name, description, intra, members FROM seg_zones WHERE mesh_id = ? ORDER BY name`, t.mesh)
	if err != nil {
		return Segmentation{}, fmt.Errorf("read zones: %w", err)
	}
	if err := scanAll(rows, "zones", func(r *sql.Rows) error {
		var (
			z       SegZone
			members string
		)
		if err := r.Scan(&z.Name, &z.Description, &z.Intra, &members); err != nil {
			return err
		}
		z.Members = splitList(members)
		s.Zones = append(s.Zones, z)
		return nil
	}); err != nil {
		return Segmentation{}, err
	}
	rows, err = t.query(`SELECT name, description, disabled, from_refs, to_refs, services, action, days, start_time, end_time
		FROM seg_policies WHERE mesh_id = ? ORDER BY position`, t.mesh)
	if err != nil {
		return Segmentation{}, fmt.Errorf("read policies: %w", err)
	}
	if err := scanAll(rows, "policies", func(r *sql.Rows) error {
		var (
			p              SegPolicy
			disabled       int
			from, to, svcs string
			days, startAt  string
			endAt          string
		)
		if err := r.Scan(&p.Name, &p.Description, &disabled, &from, &to, &svcs, &p.Action, &days, &startAt, &endAt); err != nil {
			return err
		}
		p.Disabled = disabled == 1
		p.From, p.To, p.Services = splitList(from), splitList(to), splitList(svcs)
		if days != "" || startAt != "" || endAt != "" {
			p.Schedule = &SegSchedule{Days: splitList(days), Start: startAt, End: endAt}
		}
		s.Policies = append(s.Policies, p)
		return nil
	}); err != nil {
		return Segmentation{}, err
	}
	tz, err := t.Setting(SettingPolicyTZ)
	if err != nil && !errors.Is(err, ErrNotFound) {
		return Segmentation{}, err
	}
	s.TimeZone = tz
	return s, nil
}

// groupMembers is each of the mesh's groups' nodes, by name.
func (t *MeshTx) groupMembers() (map[int64][]string, error) {
	rows, err := t.query(`SELECT ng.group_id, n.name FROM node_groups ng JOIN nodes n ON n.id = ng.node_id
		WHERE n.mesh_id = ? ORDER BY n.name`, t.mesh)
	if err != nil {
		return nil, fmt.Errorf("read group members: %w", err)
	}
	out := map[int64][]string{}
	err = scanAll(rows, "group members", func(r *sql.Rows) error {
		var (
			id   int64
			name string
		)
		if err := r.Scan(&id, &name); err != nil {
			return err
		}
		out[id] = append(out[id], name)
		return nil
	})
	return out, err
}

// NodeGroups is the groups of every node of the mesh in one, by node id,
// each sorted.
func (t *MeshTx) NodeGroups() (map[int64][]string, error) {
	rows, err := t.query(`SELECT ng.node_id, g.name FROM node_groups ng JOIN seg_groups g ON g.id = ng.group_id
		WHERE g.mesh_id = ? ORDER BY g.name`, t.mesh)
	if err != nil {
		return nil, fmt.Errorf("read node groups: %w", err)
	}
	out := map[int64][]string{}
	err = scanAll(rows, "node groups", func(r *sql.Rows) error {
		var (
			id   int64
			name string
		)
		if err := r.Scan(&id, &name); err != nil {
			return err
		}
		out[id] = append(out[id], name)
		return nil
	})
	return out, err
}

// SetSegmentation replaces the mesh's whole document. A group's node that
// is not in the mesh is ErrNotFound; s is validated by the caller.
func (t *MeshTx) SetSegmentation(s Segmentation) error {
	// node_groups goes with its groups (ON DELETE CASCADE).
	for _, table := range []string{"seg_groups", "seg_objects", "seg_zones", "seg_policies"} {
		if _, err := t.exec(`DELETE FROM `+table+` WHERE mesh_id = ?`, t.mesh); err != nil {
			return fmt.Errorf("replace segmentation: clear %s: %w", table, err)
		}
	}
	for _, g := range s.Groups {
		id, err := t.insertGroup(g.Name, g.Description)
		if err != nil {
			return err
		}
		for _, name := range g.Nodes {
			n, err := t.NodeByName(name)
			if err != nil {
				return fmt.Errorf("group %q: %w", g.Name, err)
			}
			if _, err := t.exec(`INSERT INTO node_groups (node_id, group_id) VALUES (?, ?)`, n.ID, id); err != nil {
				return fmt.Errorf("group %q: add node %q: %w", g.Name, name, err)
			}
		}
	}
	for _, kind := range []struct {
		kind string
		objs []SegObject
	}{{"address", s.Addresses}, {"service", s.Services}} {
		for _, o := range kind.objs {
			if _, err := t.exec(`INSERT INTO seg_objects (mesh_id, kind, name, description, entries) VALUES (?, ?, ?, ?, ?)`,
				t.mesh, kind.kind, o.Name, o.Description, joinList(o.Entries)); err != nil {
				return fmt.Errorf("insert %s object %q: %w", kind.kind, o.Name, err)
			}
		}
	}
	for _, z := range s.Zones {
		if _, err := t.exec(`INSERT INTO seg_zones (mesh_id, name, description, intra, members) VALUES (?, ?, ?, ?, ?)`,
			t.mesh, z.Name, z.Description, z.Intra, joinList(z.Members)); err != nil {
			return fmt.Errorf("insert zone %q: %w", z.Name, err)
		}
	}
	for i, p := range s.Policies {
		var when SegSchedule
		if p.Schedule != nil {
			when = *p.Schedule
		}
		if _, err := t.exec(`INSERT INTO seg_policies (mesh_id, position, name, description, disabled, from_refs, to_refs, services,
			action, days, start_time, end_time) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, t.mesh, i+1, p.Name, p.Description, boolInt(p.Disabled),
			joinList(p.From), joinList(p.To), joinList(p.Services), p.Action, joinList(when.Days), when.Start, when.End); err != nil {
			return fmt.Errorf("insert policy %q: %w", p.Name, err)
		}
	}
	if s.TimeZone == "" {
		return t.DeleteSetting(SettingPolicyTZ)
	}
	return t.SetSetting(SettingPolicyTZ, s.TimeZone)
}

func (t *MeshTx) insertGroup(name, description string) (int64, error) {
	res, err := t.exec(`INSERT INTO seg_groups (mesh_id, name, description) VALUES (?, ?, ?)`, t.mesh, name, description)
	if err != nil {
		return 0, fmt.Errorf("insert group %q: %w", name, err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("insert group %q: %w", name, err)
	}
	return id, nil
}

// SetNodeGroups makes groups exactly the groups node id (of the mesh) is
// in, creating a group that does not exist yet; names are validated by the
// caller.
func (t *MeshTx) SetNodeGroups(id int64, groups []string) error {
	if _, err := t.exec(`DELETE FROM node_groups WHERE node_id = ?`, id); err != nil {
		return fmt.Errorf("set the groups of node %d: %w", id, err)
	}
	for _, name := range groups {
		var gid int64
		err := t.queryRow(`SELECT id FROM seg_groups WHERE mesh_id = ? AND name = ?`, t.mesh, name).Scan(&gid)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			if gid, err = t.insertGroup(name, ""); err != nil {
				return err
			}
		case err != nil:
			return fmt.Errorf("read group %q: %w", name, err)
		}
		if _, err := t.exec(`INSERT INTO node_groups (node_id, group_id) VALUES (?, ?) ON CONFLICT DO NOTHING`, id, gid); err != nil {
			return fmt.Errorf("add node %d to group %q: %w", id, name, err)
		}
	}
	return nil
}

// scanAll calls scan for every row and closes rows.
func scanAll(rows *sql.Rows, what string, scan func(*sql.Rows) error) error {
	defer rows.Close()
	for rows.Next() {
		if err := scan(rows); err != nil {
			return fmt.Errorf("read %s: %w", what, err)
		}
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("read %s: %w", what, err)
	}
	return nil
}

func splitList(s string) []string {
	if s == "" {
		return nil
	}
	return strings.Split(s, ",")
}

func joinList(vs []string) string { return strings.Join(vs, ",") }
