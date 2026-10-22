package ui

import (
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"unicode"

	"github.com/hafio/vpn-mesh-utility/internal/proto"
	"github.com/hafio/vpn-mesh-utility/internal/server/api"
	"github.com/hafio/vpn-mesh-utility/internal/server/auth"
	"github.com/hafio/vpn-mesh-utility/internal/server/planner"
	"github.com/hafio/vpn-mesh-utility/internal/server/store"
	"github.com/hafio/vpn-mesh-utility/internal/validate"
)

// The Firewall pages (amendment A43): the mesh firewall every server
// enforces -- policies in order, zones, objects (groups, addresses,
// services) and reachability. Each form changes one item through api.Nodes,
// so every change is checked against the mesh, versioned and audited.

type fwTab struct{ Tab, Path, Label string }

// fwTabs are the Firewall pages' tabs.
var fwTabs = []fwTab{
	{"policies", "/firewall", "Policies"},
	{"zones", "/firewall/zones", "Zones"},
	{"objects", "/firewall/objects", "Objects"},
	{"reach", "/firewall/reach", "Reachability"},
}

// fwPage is what every Firewall template gets: its tab, and its own data.
type fwPage struct {
	Tab      string
	Tabs     []fwTab
	Warnings []string
	View     any
}

func (u *ui) showFirewall(w http.ResponseWriter, r *http.Request, s auth.Session, status int, name, tab, title string, p page,
	warnings []string, view any) {
	p.Data = fwPage{Tab: tab, Tabs: fwTabs, Warnings: warnings, View: view}
	u.show(w, r, s, status, name, "firewall", title, p)
}

// listOf reads a list field: entries separated by commas or spaces.
func listOf(s string) []string {
	return strings.FieldsFunc(s, func(c rune) bool { return c == ',' || unicode.IsSpace(c) })
}

func listText(vs []string) string { return strings.Join(vs, ", ") }

// anyText shows an empty list as "any".
func anyText(vs []string) string {
	if len(vs) == 0 {
		return "any"
	}
	return listText(vs)
}

func title(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// meshChanged describes what a change did to the computed configs.
func meshChanged(what string, ch api.MeshChange) string {
	return changed(what, api.NodeChange{Version: ch.Version})
}

// removeItem removes one item and shows its list again.
func (u *ui) removeItem(w http.ResponseWriter, r *http.Request, s auth.Session, kind string, render func(int, page)) {
	name := r.PathValue("name")
	ch, err := u.Nodes.RemoveSegment(r.Context(), actor(s), kind, name)
	if err != nil {
		status, msg := u.refusal(r, err)
		render(status, page{Error: msg})
		return
	}
	render(http.StatusOK, page{Flash: meshChanged(fmt.Sprintf("%s %s removed", title(kind), name), ch)})
}

// --- policies -------------------------------------------------------------

type policyRow struct {
	Pos, Up, Down                                 int
	Name, Action, From, To, Services, Description string
	Disabled, First, Last                         bool
	// Schedule is when it applies, "" for always; Inactive says its
	// schedule does not hold now.
	Schedule string
	Inactive bool
}

type ruleRow struct {
	N        int
	Text     string
	System   bool
	Inactive bool
}

type policyForm struct {
	Old, Name, Action, From, To, Services, Description, Position string
	Disabled                                                     bool
	// Days, Start and End are the schedule; Zone is the time zone it is
	// read in, for the hint.
	Days             []string
	Start, End, Zone string
}

// dayBox is one day's checkbox of the schedule.
type dayBox struct {
	Value, Label string
	Checked      bool
}

// DayBoxes are the week's checkboxes, Monday first, the form's days ticked.
func (f policyForm) DayBoxes() []dayBox {
	out := make([]dayBox, 0, len(validate.Weekdays))
	for _, d := range validate.Weekdays {
		out = append(out, dayBox{Value: d, Label: title(d), Checked: slices.Contains(f.Days, d)})
	}
	return out
}

type policiesView struct {
	Policies []policyRow
	Rules    []ruleRow
	Form     policyForm
	// TimeZone is the document's zone ("" is the server's), Zone the one
	// in use and Now the time there.
	TimeZone, Zone, Now string
}

// ruleRows lists what every server enforces; inactive are the indexes of
// the rules whose schedule does not hold now.
func ruleRows(rules []proto.PolicyRule, inactive []int) []ruleRow {
	out := make([]ruleRow, 0, len(rules))
	for i, r := range rules {
		out = append(out, ruleRow{N: i + 1, Text: r.String(), System: strings.HasPrefix(r.Name, "system: "), Inactive: slices.Contains(inactive, i)})
	}
	return out
}

// scheduleText is a policy's schedule as people read it, "" for none.
func scheduleText(s *store.SegSchedule) string {
	if s == nil {
		return ""
	}
	return s.String()
}

// newPolicyForm goes just above the last policy (allow-all, at first), so
// a new policy is not shadowed by it.
func newPolicyForm(n int) policyForm {
	return policyForm{Action: store.SegAllow, Position: strconv.Itoa(max(n, 1))}
}

func (u *ui) firewallPage(w http.ResponseWriter, r *http.Request, s auth.Session) {
	u.renderPolicies(w, r, s, http.StatusOK, page{}, nil)
}

func (u *ui) renderPolicies(w http.ResponseWriter, r *http.Request, s auth.Session, status int, p page, f *policyForm) {
	v, err := u.Nodes.Segment(r.Context())
	if err != nil {
		u.fail(w, r, s, err)
		return
	}
	pv := policiesView{Rules: ruleRows(v.Rules, v.Inactive), Form: newPolicyForm(len(v.Config.Policies)), TimeZone: v.Config.TimeZone,
		Zone: v.TimeZone, Now: v.Now}
	inactive := map[string]bool{}
	for _, i := range v.Inactive {
		inactive[v.Rules[i].Name] = true
	}
	for i, pol := range v.Config.Policies {
		pv.Policies = append(pv.Policies, policyRow{Pos: i + 1, Up: i, Down: i + 2, Name: pol.Name, Action: pol.Action, From: anyText(pol.From),
			To: anyText(pol.To), Services: anyText(pol.Services), Description: pol.Description, Disabled: pol.Disabled,
			First: i == 0, Last: i == len(v.Config.Policies)-1, Schedule: scheduleText(pol.Schedule), Inactive: inactive["policy: "+pol.Name]})
	}
	if f != nil {
		pv.Form = *f
	}
	pv.Form.Zone = v.TimeZone
	u.showFirewall(w, r, s, status, "firewall", "policies", "Firewall", p, v.Warnings, pv)
}

func readPolicyForm(r *http.Request) policyForm {
	f := policyForm{Old: r.PathValue("name"), Name: strings.TrimSpace(r.PostFormValue("name")), Action: r.PostFormValue("action"),
		From: r.PostFormValue("from"), To: r.PostFormValue("to"), Services: r.PostFormValue("services"),
		Description: strings.TrimSpace(r.PostFormValue("description")), Position: strings.TrimSpace(r.PostFormValue("position")),
		Disabled: r.PostFormValue("disabled") != "", Start: strings.TrimSpace(r.PostFormValue("start")),
		End: strings.TrimSpace(r.PostFormValue("end"))}
	f.Days = r.PostForm["days"] // parsed by the PostFormValue calls above
	return f
}

// policy is the form as a policy: no days and no times is no schedule.
func (f policyForm) policy() store.SegPolicy {
	p := store.SegPolicy{Name: f.Name, Description: f.Description, Disabled: f.Disabled, From: listOf(f.From), To: listOf(f.To),
		Services: listOf(f.Services), Action: f.Action}
	if len(f.Days) > 0 || f.Start != "" || f.End != "" {
		p.Schedule = &store.SegSchedule{Days: f.Days, Start: f.Start, End: f.End}
	}
	return p
}

// savePolicy adds a policy (POST /firewall/policies) or changes one (POST
// /firewall/policies/{name}).
func (u *ui) savePolicy(w http.ResponseWriter, r *http.Request, s auth.Session) {
	f := readPolicyForm(r)
	pos := 0
	var err error
	if f.Position != "" {
		pos, err = atoi("position", f.Position)
	}
	var ch api.MeshChange
	if err == nil {
		ch, err = u.Nodes.SavePolicy(r.Context(), actor(s), f.Old, f.policy(), pos)
	}
	if err != nil {
		status, msg := u.refusal(r, err)
		if f.Old != "" {
			u.renderPolicy(w, r, s, status, page{Error: msg}, &f)
			return
		}
		u.renderPolicies(w, r, s, status, page{Error: msg}, &f)
		return
	}
	u.renderPolicies(w, r, s, http.StatusOK, page{Flash: meshChanged("Policy "+f.Name+" saved", ch)}, nil)
}

type policyView struct {
	Form  policyForm
	Count int
}

func (u *ui) policyPage(w http.ResponseWriter, r *http.Request, s auth.Session) {
	u.renderPolicy(w, r, s, http.StatusOK, page{}, nil)
}

func (u *ui) renderPolicy(w http.ResponseWriter, r *http.Request, s auth.Session, status int, p page, f *policyForm) {
	name := r.PathValue("name")
	v, err := u.Nodes.Segment(r.Context())
	if err != nil {
		u.fail(w, r, s, err)
		return
	}
	i := slices.IndexFunc(v.Config.Policies, func(p store.SegPolicy) bool { return p.Name == name })
	if i < 0 {
		u.renderPolicies(w, r, s, http.StatusNotFound, page{Error: fmt.Sprintf("there is no policy %q; the list below has them all", name)}, nil)
		return
	}
	pol := v.Config.Policies[i]
	pv := policyView{Count: len(v.Config.Policies), Form: policyForm{Old: name, Name: pol.Name, Action: pol.Action, From: listText(pol.From),
		To: listText(pol.To), Services: listText(pol.Services), Description: pol.Description, Disabled: pol.Disabled,
		Position: strconv.Itoa(i + 1)}}
	if pol.Schedule != nil {
		pv.Form.Days, pv.Form.Start, pv.Form.End = pol.Schedule.Days, pol.Schedule.Start, pol.Schedule.End
	}
	if f != nil {
		pv.Form = *f
	}
	pv.Form.Zone = v.TimeZone
	u.showFirewall(w, r, s, status, "policy", "policies", "Policy "+name, p, nil, pv)
}

func (u *ui) movePolicy(w http.ResponseWriter, r *http.Request, s auth.Session) {
	name := r.PathValue("name")
	pos, err := atoi("position", r.PostFormValue("position"))
	var ch api.MeshChange
	if err == nil {
		ch, err = u.Nodes.MovePolicy(r.Context(), actor(s), name, pos)
	}
	if err != nil {
		status, msg := u.refusal(r, err)
		u.renderPolicies(w, r, s, status, page{Error: msg}, nil)
		return
	}
	u.renderPolicies(w, r, s, http.StatusOK, page{Flash: meshChanged(fmt.Sprintf("Policy %s moved to %d", name, pos), ch)}, nil)
}

func (u *ui) removePolicy(w http.ResponseWriter, r *http.Request, s auth.Session) {
	u.removeItem(w, r, s, api.SegPolicy, func(status int, p page) { u.renderPolicies(w, r, s, status, p, nil) })
}

// saveTimeZone sets the time zone the schedules are read in (POST
// /firewall/timezone); empty is the server's.
func (u *ui) saveTimeZone(w http.ResponseWriter, r *http.Request, s auth.Session) {
	tz := strings.TrimSpace(r.PostFormValue("timezone"))
	ch, err := u.Nodes.SetTimeZone(r.Context(), actor(s), tz)
	if err != nil {
		status, msg := u.refusal(r, err)
		u.renderPolicies(w, r, s, status, page{Error: msg}, nil)
		return
	}
	if tz == "" {
		tz = "the server's"
	}
	u.renderPolicies(w, r, s, http.StatusOK, page{Flash: meshChanged("Schedules now follow "+tz+" time", ch)}, nil)
}

// --- zones -----------------------------------------------------------------

type zoneRow struct {
	Name, Intra, Members, Addresses, Description string
	Inbound, Outbound                            int
}

type zoneForm struct {
	Old, Name, Intra, Members, Description string
}

type zonesView struct {
	Zones []zoneRow
	Form  zoneForm
}

func (u *ui) zonesPage(w http.ResponseWriter, r *http.Request, s auth.Session) {
	u.renderZones(w, r, s, http.StatusOK, page{}, nil)
}

func (u *ui) renderZones(w http.ResponseWriter, r *http.Request, s auth.Session, status int, p page, f *zoneForm) {
	v, err := u.Nodes.Segment(r.Context())
	if err != nil {
		u.fail(w, r, s, err)
		return
	}
	zv := zonesView{Form: zoneForm{Intra: store.SegAllow}}
	for i, z := range v.Config.Zones {
		zv.Zones = append(zv.Zones, zoneRow{Name: z.Name, Intra: z.Intra, Members: listText(z.Members),
			Addresses: listText(v.Zones[i].Addresses), Description: z.Description,
			Inbound: len(v.Zones[i].Inbound), Outbound: len(v.Zones[i].Outbound)})
	}
	if f != nil {
		zv.Form = *f
	}
	u.showFirewall(w, r, s, status, "zones", "zones", "Firewall zones", p, v.Warnings, zv)
}

func readZoneForm(r *http.Request) zoneForm {
	return zoneForm{Old: r.PathValue("name"), Name: strings.TrimSpace(r.PostFormValue("name")), Intra: r.PostFormValue("intra"),
		Members: r.PostFormValue("members"), Description: strings.TrimSpace(r.PostFormValue("description"))}
}

// saveZone adds a zone (POST /firewall/zones) or changes one (POST
// /firewall/zones/{name}).
func (u *ui) saveZone(w http.ResponseWriter, r *http.Request, s auth.Session) {
	f := readZoneForm(r)
	ch, err := u.Nodes.SaveZone(r.Context(), actor(s), f.Old, store.SegZone{Name: f.Name, Description: f.Description, Intra: f.Intra,
		Members: listOf(f.Members)})
	if err != nil {
		status, msg := u.refusal(r, err)
		if f.Old != "" {
			u.renderZone(w, r, s, status, page{Error: msg}, &f)
			return
		}
		u.renderZones(w, r, s, status, page{Error: msg}, &f)
		return
	}
	u.renderZones(w, r, s, http.StatusOK, page{Flash: meshChanged("Zone "+f.Name+" saved", ch)}, nil)
}

type zoneView struct {
	Form      zoneForm
	Addresses string
	Inbound   []ruleRow
	Outbound  []ruleRow
}

func (u *ui) zonePage(w http.ResponseWriter, r *http.Request, s auth.Session) {
	u.renderZone(w, r, s, http.StatusOK, page{}, nil)
}

func (u *ui) renderZone(w http.ResponseWriter, r *http.Request, s auth.Session, status int, p page, f *zoneForm) {
	name := r.PathValue("name")
	v, err := u.Nodes.Segment(r.Context())
	if err != nil {
		u.fail(w, r, s, err)
		return
	}
	i := slices.IndexFunc(v.Config.Zones, func(z store.SegZone) bool { return z.Name == name })
	if i < 0 {
		u.renderZones(w, r, s, http.StatusNotFound, page{Error: fmt.Sprintf("there is no zone %q; the list below has them all", name)}, nil)
		return
	}
	z, rules := v.Config.Zones[i], ruleRows(v.Rules, v.Inactive)
	zv := zoneView{Form: zoneForm{Old: name, Name: z.Name, Intra: z.Intra, Members: listText(z.Members), Description: z.Description},
		Addresses: listText(v.Zones[i].Addresses)}
	for _, n := range v.Zones[i].Inbound {
		zv.Inbound = append(zv.Inbound, rules[n])
	}
	for _, n := range v.Zones[i].Outbound {
		zv.Outbound = append(zv.Outbound, rules[n])
	}
	if f != nil {
		zv.Form = *f
	}
	u.showFirewall(w, r, s, status, "zone", "zones", "Zone "+name, p, nil, zv)
}

func (u *ui) removeZone(w http.ResponseWriter, r *http.Request, s auth.Session) {
	u.removeItem(w, r, s, api.SegZone, func(status int, p page) { u.renderZones(w, r, s, status, p, nil) })
}

// --- objects ---------------------------------------------------------------

type objectRow struct {
	Name, Entries, Description string
}

type objectForm struct {
	Kind, Name, Entries, Description string
}

type objectsView struct {
	Groups, Addresses, Services []objectRow
	Nodes                       string
	// Form is the form the operator last used, kept on a refusal.
	Form objectForm
}

func objectRows(objs []store.SegObject) []objectRow {
	var out []objectRow
	for _, o := range objs {
		out = append(out, objectRow{Name: o.Name, Entries: listText(o.Entries), Description: o.Description})
	}
	return out
}

func (u *ui) objectsPage(w http.ResponseWriter, r *http.Request, s auth.Session) {
	u.renderObjects(w, r, s, http.StatusOK, page{}, objectForm{})
}

func (u *ui) renderObjects(w http.ResponseWriter, r *http.Request, s auth.Session, status int, p page, f objectForm) {
	v, err := u.Nodes.Segment(r.Context())
	if err != nil {
		u.fail(w, r, s, err)
		return
	}
	ov := objectsView{Addresses: objectRows(v.Config.Addresses), Services: objectRows(v.Config.Services), Nodes: listText(v.Nodes), Form: f}
	for _, g := range v.Config.Groups {
		ov.Groups = append(ov.Groups, objectRow{Name: g.Name, Entries: listText(g.Nodes), Description: g.Description})
	}
	u.showFirewall(w, r, s, status, "objects", "objects", "Firewall objects", p, v.Warnings, ov)
}

// saveObject creates a group, an address object or a service object, or
// replaces the one of that name (POST /firewall/objects/{kind}).
func (u *ui) saveObject(w http.ResponseWriter, r *http.Request, s auth.Session) {
	kind := r.PathValue("kind")
	f := objectForm{Kind: kind, Name: strings.TrimSpace(r.PostFormValue("name")), Entries: r.PostFormValue("entries"),
		Description: strings.TrimSpace(r.PostFormValue("description"))}
	ch, err := u.saveObjectKind(r, s, f)
	if err != nil {
		status, msg := u.refusal(r, err)
		u.renderObjects(w, r, s, status, page{Error: msg}, f)
		return
	}
	u.renderObjects(w, r, s, http.StatusOK, page{Flash: meshChanged(title(kind)+" "+f.Name+" saved", ch)}, objectForm{})
}

func (u *ui) saveObjectKind(r *http.Request, s auth.Session, f objectForm) (api.MeshChange, error) {
	v, err := u.Nodes.Segment(r.Context())
	if err != nil {
		return api.MeshChange{}, err
	}
	// Saving under a name that exists replaces it: one form adds and changes.
	old := func(has bool) string {
		if has {
			return f.Name
		}
		return ""
	}
	switch f.Kind {
	case api.SegGroup:
		has := slices.ContainsFunc(v.Config.Groups, func(g store.SegGroup) bool { return g.Name == f.Name })
		return u.Nodes.SaveGroup(r.Context(), actor(s), old(has), store.SegGroup{Name: f.Name, Description: f.Description,
			Nodes: listOf(f.Entries)})
	case api.SegAddress, api.SegService:
		objs := v.Config.Addresses
		if f.Kind == api.SegService {
			objs = v.Config.Services
		}
		has := slices.ContainsFunc(objs, func(o store.SegObject) bool { return o.Name == f.Name })
		return u.Nodes.SaveObject(r.Context(), actor(s), f.Kind, old(has), store.SegObject{Name: f.Name, Description: f.Description,
			Entries: listOf(f.Entries)})
	}
	return api.MeshChange{}, &validate.Error{Field: "kind", Value: f.Kind, Reason: "must be group, address or service"}
}

func (u *ui) removeObject(w http.ResponseWriter, r *http.Request, s auth.Session) {
	u.removeItem(w, r, s, r.PathValue("kind"), func(status int, p page) { u.renderObjects(w, r, s, status, p, objectForm{}) })
}

// --- reachability ----------------------------------------------------------

type matrixRow struct {
	Zone  string
	Cells []matrixCell
}

type matrixCell struct {
	Class, Verdict, Rules string
}

type reachView struct {
	From, To, Service, At string
	Nodes                 []string
	Result                *api.ReachResult
	// RuleN is the deciding rule's number, 0 for none.
	RuleN int
	Zones []string
	Rows  []matrixRow
	// MatrixAt is when the matrix holds.
	MatrixAt string
}

// reachPage shows how every zone reaches every other, and the verdict for
// one flow when the check form was filled in (GET
// /firewall/reach?from=&to=&service=).
func (u *ui) reachPage(w http.ResponseWriter, r *http.Request, s auth.Session) {
	q := r.URL.Query()
	rv := reachView{From: strings.TrimSpace(q.Get("from")), To: strings.TrimSpace(q.Get("to")), Service: strings.TrimSpace(q.Get("service")),
		At: strings.TrimSpace(q.Get("at"))}
	p, status := page{}, http.StatusOK
	if rv.From != "" || rv.To != "" || rv.Service != "" {
		res, err := u.Nodes.Reach(r.Context(), api.ReachRequest{From: rv.From, To: rv.To, Service: rv.Service, At: rv.At})
		if err != nil {
			status, p.Error = u.refusal(r, err)
		} else {
			rv.Result, rv.RuleN = &res, res.Index+1
		}
	}
	m, err := u.Nodes.Matrix(r.Context())
	if err != nil {
		u.fail(w, r, s, err)
		return
	}
	v, err := u.Nodes.Segment(r.Context())
	if err != nil {
		u.fail(w, r, s, err)
		return
	}
	rv.Nodes, rv.Zones, rv.MatrixAt = v.Nodes, m.Zones, m.At
	class := map[string]string{planner.ReachAllow: "ok", planner.ReachDeny: "bad", planner.ReachPartial: "warn"}
	for i, z := range m.Zones {
		row := matrixRow{Zone: z}
		for _, c := range m.Cells[i] {
			var nums []string
			for _, n := range c.Rules {
				nums = append(nums, strconv.Itoa(n+1))
			}
			row.Cells = append(row.Cells, matrixCell{Class: class[c.Verdict], Verdict: c.Verdict, Rules: strings.Join(nums, ", ")})
		}
		rv.Rows = append(rv.Rows, row)
	}
	u.showFirewall(w, r, s, status, "reach", "reach", "Firewall reachability", p, nil, rv)
}
