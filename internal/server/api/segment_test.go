package api

import (
	"context"
	"net/http"
	"net/url"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/hafio/vpn-mesh-utility/internal/proto"
	"github.com/hafio/vpn-mesh-utility/internal/server/store"
	"github.com/hafio/vpn-mesh-utility/internal/testfix"
)

// officeFirewall is a mesh firewall for meshHarness: the laptop may reach
// the NAS's SMB and nothing else of the home zone; allow-all is gone.
func officeFirewall() store.Segmentation {
	return store.Segmentation{
		Groups:   []store.SegGroup{{Name: "laptops", Nodes: []string{"laptop"}}},
		Services: []store.SegObject{{Name: "smb", Entries: []string{"tcp/445"}}},
		Zones: []store.SegZone{{Name: "home", Intra: store.SegAllow, Members: []string{"node:nas"}},
			{Name: "office", Intra: store.SegAllow, Members: []string{"group:laptops"}}},
		Policies: []store.SegPolicy{
			{Name: "smb", From: []string{"zone:office"}, To: []string{"zone:home"}, Services: []string{"svc:smb"}, Action: store.SegAllow},
			{Name: "no-home", To: []string{"zone:home"}, Action: store.SegDeny},
		},
	}
}

func TestSegmentView(t *testing.T) {
	h := meshHarness(t)
	s, ctx := h.nodes(), context.Background()
	v, err := s.Segment(ctx)
	if err != nil || len(v.Config.Policies) != 1 || v.Config.Policies[0].Name != "allow-all" || len(v.Rules) != 4 ||
		v.Rules[3].Name != "policy: allow-all" || !slices.Equal(v.Nodes, []string{"laptop", "nas", "vps-1", "vps-2", "vps-3"}) ||
		v.TimeZone != "UTC" || v.Now != "Mon 2026-10-12 08:00 UTC" || len(v.Inactive) != 0 {
		t.Fatalf("a new mesh's firewall = %+v, %v", v, err)
	}
	if _, err := s.SetSegment(ctx, cli(), officeFirewall()); err != nil {
		t.Fatal(err)
	}
	if v, err = s.Segment(ctx); err != nil {
		t.Fatal(err)
	}
	// Rules: ping, dns, swarm, smb, no-home, zone home, zone office.
	home, office := v.Zones[0], v.Zones[1]
	if !slices.Equal(home.Addresses, []string{testfix.NAS.MeshIP}) || !slices.Equal(office.Addresses, []string{testfix.Laptop.MeshIP}) {
		t.Fatalf("zones = %+v", v.Zones)
	}
	// A rule from any reaches out of every zone: no-home (4) is in both
	// zones' outbound.
	if !slices.Equal(home.Inbound, []int{0, 3, 4, 5}) || !slices.Equal(home.Outbound, []int{0, 1, 4, 5}) ||
		!slices.Equal(office.Outbound, []int{0, 1, 3, 4, 6}) {
		t.Errorf("home in %v out %v, office out %v", home.Inbound, home.Outbound, office.Outbound)
	}
}

// The document is replaced whole, checked against the mesh, versioned,
// audited and pushed to the servers.
func TestSetSegment(t *testing.T) {
	h := meshHarness(t)
	s, ctx := h.nodes(), context.Background()
	ch, err := s.SetSegment(ctx, cli(), officeFirewall())
	if err != nil || ch.Version == 0 || h.audits(t, "segmentation.set") != 1 {
		t.Fatalf("SetSegment = %+v, %v", ch, err)
	}
	pol := h.config(t, "vps-1").Policy
	if pol == nil || len(pol.Rules) != 7 || pol.Rules[3].Name != "policy: smb" || !slices.Equal(pol.Rules[3].From, []string{testfix.Laptop.MeshIP}) {
		t.Fatalf("vps-1's policy = %+v", pol)
	}
	if c := h.config(t, "nas"); c.Policy != nil {
		t.Fatal("a spoke got a policy")
	}

	// A bad value is refused before the database.
	h.db.n.Store(0)
	bad := officeFirewall()
	bad.Policies[0].Action = "accept"
	_, err = s.SetSegment(ctx, cli(), bad)
	refused(t, err, http.StatusBadRequest, "policies[smb].action")
	if n := h.db.n.Load(); n != 0 {
		t.Fatalf("a refused document ran %d transactions", n)
	}
	// A dangling reference, or a group of strangers, is refused by the plan.
	v := h.latestVersion(t)
	dangling := officeFirewall()
	dangling.Policies[0].To = []string{"zone:garage"}
	_, err = s.SetSegment(ctx, cli(), dangling)
	refused(t, err, http.StatusConflict, `there is no zone "garage"`)
	stranger := officeFirewall()
	stranger.Groups[0].Nodes = []string{"ghost"}
	_, err = s.SetSegment(ctx, cli(), stranger)
	refused(t, err, http.StatusBadRequest, "a group may only list nodes of this mesh")
	if h.latestVersion(t) != v {
		t.Fatal("a refused document was written")
	}
	// The same document again changes no config.
	if ch, err = s.SetSegment(ctx, cli(), officeFirewall()); err != nil || ch.Version != 0 {
		t.Fatalf("the same document again = %+v, %v", ch, err)
	}
}

// One item at a time, as the web UI edits: a rename follows into the
// references, policies move, an item still named elsewhere stays.
func TestEditSegment(t *testing.T) {
	h := meshHarness(t)
	s, ctx, by := h.nodes(), context.Background(), Actor{Name: "user:ops", UserID: 1}
	steps := []func() (MeshChange, error){
		func() (MeshChange, error) {
			return s.SaveGroup(ctx, by, "", store.SegGroup{Name: "laptops", Nodes: []string{"laptop"}})
		},
		func() (MeshChange, error) {
			return s.SaveObject(ctx, by, SegService, "", store.SegObject{Name: "smb", Entries: []string{"tcp/445"}})
		},
		func() (MeshChange, error) {
			return s.SaveObject(ctx, by, SegAddress, "", store.SegObject{Name: "printer", Entries: []string{"10.77.1.200"}})
		},
		func() (MeshChange, error) {
			return s.SaveZone(ctx, by, "", store.SegZone{Name: "home", Intra: store.SegAllow, Members: []string{"node:nas", "addr:printer"}})
		},
		func() (MeshChange, error) {
			return s.SavePolicy(ctx, by, "", store.SegPolicy{Name: "smb", From: []string{"group:laptops"}, To: []string{"zone:home"},
				Services: []string{"svc:smb"}, Action: store.SegAllow}, 1)
		},
	}
	for i, step := range steps {
		if _, err := step(); err != nil {
			t.Fatalf("step %d: %v", i, err)
		}
	}
	seg := h.segmentation(t)
	if names := policyNames(seg); !slices.Equal(names, []string{"smb", "allow-all"}) {
		t.Fatalf("policies = %v", names)
	}

	// Renames: the zone, the group and the objects, each followed.
	for _, rename := range []func() (MeshChange, error){
		func() (MeshChange, error) {
			return s.SaveZone(ctx, by, "home", store.SegZone{Name: "house", Intra: store.SegAllow, Members: []string{"node:nas", "addr:printer"}})
		},
		func() (MeshChange, error) {
			return s.SaveGroup(ctx, by, "laptops", store.SegGroup{Name: "notebooks", Nodes: []string{"laptop"}})
		},
		func() (MeshChange, error) {
			return s.SaveObject(ctx, by, SegService, "smb", store.SegObject{Name: "cifs", Entries: []string{"tcp/445"}})
		},
		func() (MeshChange, error) {
			return s.SaveObject(ctx, by, SegAddress, "printer", store.SegObject{Name: "lp", Entries: []string{"10.77.1.200"}})
		},
	} {
		if _, err := rename(); err != nil {
			t.Fatalf("rename: %v", err)
		}
	}
	seg = h.segmentation(t)
	p := seg.Policies[0]
	if !reflect.DeepEqual(p.From, []string{"group:notebooks"}) || !reflect.DeepEqual(p.To, []string{"zone:house"}) ||
		!reflect.DeepEqual(p.Services, []string{"svc:cifs"}) || !reflect.DeepEqual(seg.Zones[0].Members, []string{"node:nas", "addr:lp"}) {
		t.Fatalf("after the renames: policy %+v, zone %+v", p, seg.Zones[0])
	}

	// Order: down, then to the top again; a bad position is refused.
	if _, err := s.MovePolicy(ctx, by, "smb", 2); err != nil {
		t.Fatal(err)
	}
	if names := policyNames(h.segmentation(t)); !slices.Equal(names, []string{"allow-all", "smb"}) {
		t.Fatalf("after the move: %v", names)
	}
	_, err := s.MovePolicy(ctx, by, "smb", 3)
	refused(t, err, http.StatusBadRequest, "give 1 to 2")
	_, err = s.MovePolicy(ctx, by, "ghost", 1)
	refused(t, err, http.StatusNotFound, `no policy "ghost"`)
	if _, err := s.SavePolicy(ctx, by, "smb", p, 1); err != nil {
		t.Fatal(err)
	}
	if names := policyNames(h.segmentation(t)); !slices.Equal(names, []string{"smb", "allow-all"}) {
		t.Fatalf("after saving at 1: %v", names)
	}

	// The zone a policy names cannot go first; the policy, then the zone can.
	_, err = s.RemoveSegment(ctx, by, SegZone, "house")
	refused(t, err, http.StatusConflict, `there is no zone "house"`)
	for _, item := range [][2]string{{SegPolicy, "smb"}, {SegZone, "house"}, {SegAddress, "lp"}, {SegService, "cifs"}, {SegGroup, "notebooks"}} {
		if _, err := s.RemoveSegment(ctx, by, item[0], item[1]); err != nil {
			t.Fatalf("remove %s %s: %v", item[0], item[1], err)
		}
	}
	if seg := h.segmentation(t); len(seg.Policies) != 1 || len(seg.Zones)+len(seg.Groups)+len(seg.Addresses)+len(seg.Services) != 0 {
		t.Fatalf("left after the removals: %+v", seg)
	}
	_, err = s.RemoveSegment(ctx, by, SegZone, "house")
	refused(t, err, http.StatusNotFound, `there is no zone "house"`)
	_, err = s.RemoveSegment(ctx, by, "rule", "x")
	refused(t, err, http.StatusBadRequest, "use group, address, service, zone or policy")
	_, err = s.SaveObject(ctx, by, "rule", "", store.SegObject{Name: "x", Entries: []string{"tcp"}})
	refused(t, err, http.StatusBadRequest, "use address or service")
	_, err = s.SaveZone(ctx, by, "ghost", store.SegZone{Name: "x", Intra: store.SegAllow})
	refused(t, err, http.StatusNotFound, `no zone "ghost"`)
	_, err = s.SaveGroup(ctx, by, "", store.SegGroup{Name: "Bad Name"})
	refused(t, err, http.StatusBadRequest, "groups[0].name")
	if h.audits(t, "segmentation.rm") != 5 || h.audits(t, "segmentation.move") != 1 {
		t.Fatal("the edits were not audited")
	}
}

func (h *harness) segmentation(t *testing.T) store.Segmentation {
	t.Helper()
	var seg store.Segmentation
	if err := h.st.Tx(context.Background(), func(tx *store.Tx) error {
		var err error
		seg, err = tx.Segmentation()
		return err
	}); err != nil {
		t.Fatal(err)
	}
	return seg
}

func policyNames(s store.Segmentation) []string {
	var out []string
	for _, p := range s.Policies {
		out = append(out, p.Name)
	}
	return out
}

func TestReach(t *testing.T) {
	h := meshHarness(t)
	s, ctx := h.nodes(), context.Background()
	if _, err := s.SetSegment(ctx, cli(), officeFirewall()); err != nil {
		t.Fatal(err)
	}
	for _, c := range []struct {
		from, to, service string
		allowed           bool
		rule              string
	}{
		{"laptop", "nas", "tcp/445", true, "policy: smb"},
		{"laptop", "nas", "tcp/22", false, "policy: no-home"},
		{"vps-1", "nas", "icmp", true, "system: ping"},
		{"nas", "vps-2", "udp/53", true, "system: dns"},
		{"nas", "laptop", "tcp/80", false, "no rule matched: dropped"},
		{"10.77.2.1", testfix.NAS.MeshIP, "tcp/445", true, "policy: smb"},
		{"laptop", "1.1.1.1", "tcp/443", true, "not mesh traffic"},
	} {
		r, err := s.Reach(ctx, ReachRequest{From: c.from, To: c.to, Service: c.service})
		if err != nil || r.Allowed != c.allowed || !strings.HasPrefix(r.Rule, c.rule) {
			t.Errorf("%s -> %s %s = %+v, %v; want allowed=%v by %q", c.from, c.to, c.service, r, err, c.allowed, c.rule)
		}
	}
	h.db.n.Store(0)
	for _, req := range []ReachRequest{
		{From: "laptop", To: "nas", Service: "tcp"},
		{From: "laptop", To: "nas", Service: "tcp/1-2"},
		{From: "laptop", To: "nas", Service: "sctp/1"},
	} {
		_, err := s.Reach(ctx, req)
		refused(t, err, http.StatusBadRequest, "service")
	}
	if n := h.db.n.Load(); n != 0 {
		t.Fatalf("a refused check ran %d transactions", n)
	}
	_, err := s.Reach(ctx, ReachRequest{From: "ghost", To: "nas", Service: "icmp"})
	refused(t, err, http.StatusBadRequest, "a node of this mesh or an address")
}

// Dual stack (M8): two nodes meet over IPv4 when both have it, else over
// IPv6; an address of either family is an end; nodes without a family in
// common cannot meet.
func TestReachDualStack(t *testing.T) {
	h := meshHarness(t)
	s, ctx := h.nodes(), context.Background()
	if _, err := s.SetIPv6(ctx, cli(), IPv6Request{Prefix: testfix.MeshCIDR6}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Set(ctx, cli(), "laptop", NodeSetRequest{Family: ptr(proto.FamilyIPv6)}); err != nil {
		t.Fatal(err)
	}
	for _, c := range []struct{ from, to, wantFrom, wantTo string }{
		{"nas", "vps-1", testfix.NAS.MeshIP, testfix.VPS1.MeshIP},
		{"laptop", "vps-1", testfix.Laptop.MeshIP6, testfix.VPS1.MeshIP6},
		{testfix.NAS.MeshIP6, "vps-2", testfix.NAS.MeshIP6, testfix.VPS2.MeshIP6},
	} {
		r, err := s.Reach(ctx, ReachRequest{From: c.from, To: c.to, Service: "icmp"})
		if err != nil || r.From != c.wantFrom || r.To != c.wantTo || !r.Allowed || r.Rule != "system: ping" {
			t.Errorf("%s -> %s = %+v, %v; want %s -> %s by system: ping", c.from, c.to, r, err, c.wantFrom, c.wantTo)
		}
	}
	if _, err := s.Set(ctx, cli(), "nas", NodeSetRequest{Family: ptr(proto.FamilyIPv4)}); err != nil {
		t.Fatal(err)
	}
	_, err := s.Reach(ctx, ReachRequest{From: "laptop", To: "nas", Service: "icmp"})
	refused(t, err, http.StatusBadRequest, "no address family in common")
}

func TestMatrix(t *testing.T) {
	h := meshHarness(t)
	s, ctx := h.nodes(), context.Background()
	if _, err := s.SetSegment(ctx, cli(), officeFirewall()); err != nil {
		t.Fatal(err)
	}
	m, err := s.Matrix(ctx)
	if err != nil || !slices.Equal(m.Zones, []string{"home", "office"}) {
		t.Fatalf("Matrix = %+v, %v", m, err)
	}
	// Ping crosses everywhere; no-home denies the rest into home, even from
	// home (it comes before the zone's own allow); office allows within
	// itself; home reaches office for ping only.
	for i, row := range [][]string{{"partial", "partial"}, {"partial", "allow"}} {
		for j, want := range row {
			if got := m.Cells[i][j].Verdict; got != want {
				t.Errorf("%s -> %s = %s, want %s", m.Zones[i], m.Zones[j], got, want)
			}
		}
	}
}

// Scheduled policies (the harness's clock reads Monday 08:00 UTC, 16:00 in
// Singapore): every server gets the schedule and the zone; the views, the
// check and the matrix judge the time asked about, and say which rules
// sleep then. The zone falls back to the server's, then UTC.
func TestScheduledPolicies(t *testing.T) {
	h := meshHarness(t)
	s, ctx, by := h.nodes(), context.Background(), Actor{Name: "user:ops", UserID: 1}
	seg := officeFirewall()
	seg.Policies[0].Schedule = &store.SegSchedule{Days: []string{"fri", "mon", "tue", "wed", "thu"}, Start: "09:00", End: "17:00"}
	seg.TimeZone = "Asia/Singapore"
	if _, err := s.SetSegment(ctx, cli(), seg); err != nil {
		t.Fatal(err)
	}
	pol := h.config(t, "vps-1").Policy
	want := &proto.Schedule{Days: []string{"mon", "tue", "wed", "thu", "fri"}, Start: "09:00", End: "17:00"}
	if pol.TZ != "Asia/Singapore" || !reflect.DeepEqual(pol.Rules[3].Schedule, want) {
		t.Fatalf("vps-1's policy = %+v", pol)
	}
	smb := ReachRequest{From: "laptop", To: "nas", Service: "tcp/445"}
	if r, err := s.Reach(ctx, smb); err != nil || !r.Allowed || r.Rule != "policy: smb" || r.At != "2026-10-12T16:00:00+08:00" || r.Inactive != nil {
		t.Fatalf("SMB now = %+v, %v", r, err)
	}
	smb.At = "2026-10-12T18:00:00+08:00"
	if r, err := s.Reach(ctx, smb); err != nil || r.Allowed || r.Rule != "policy: no-home" || !slices.Equal(r.Inactive, []string{"policy: smb"}) {
		t.Fatalf("SMB after hours = %+v, %v", r, err)
	}

	h.clock.Advance(2 * time.Hour) // 18:00 in Singapore
	v, err := s.Segment(ctx)
	if err != nil || v.TimeZone != "Asia/Singapore" || v.Now != "Mon 2026-10-12 18:00 +08" || !slices.Equal(v.Inactive, []int{3}) {
		t.Fatalf("the view after hours = %+v, %v", v, err)
	}
	m, err := s.Matrix(ctx)
	if err != nil || m.At != v.Now || !slices.Equal(m.Cells[1][0].Rules, []int{0, 4}) {
		t.Fatalf("the matrix after hours = %+v, %v", m, err)
	}

	// A bad time is refused before the database; so is a bad zone.
	h.db.n.Store(0)
	_, err = s.Reach(ctx, ReachRequest{From: "laptop", To: "nas", Service: "tcp/445", At: "18:00"})
	refused(t, err, http.StatusBadRequest, "RFC 3339")
	if n := h.db.n.Load(); n != 0 {
		t.Fatalf("a refused check ran %d transactions", n)
	}
	version := h.latestVersion(t)
	_, err = s.SetTimeZone(ctx, by, "Mars/Olympus")
	refused(t, err, http.StatusBadRequest, "timezone")
	if h.latestVersion(t) != version {
		t.Fatal("a refused time zone was written")
	}
	// The server's zone, then UTC.
	if _, err := s.SetTimeZone(ctx, by, ""); err != nil || h.audits(t, "segmentation.timezone") != 1 {
		t.Fatalf("SetTimeZone = %v", err)
	}
	if v, err = s.Segment(ctx); err != nil || v.TimeZone != "UTC" || h.config(t, "vps-1").Policy.TZ != "UTC" {
		t.Fatalf("without a zone: %q, %v", v.TimeZone, err)
	}
	if err := h.st.Tx(ctx, func(tx *store.Tx) error { return tx.SetSetting(store.SettingServerTZ, "Europe/Berlin") }); err != nil {
		t.Fatal(err)
	}
	if v, err = s.Segment(ctx); err != nil || v.TimeZone != "Europe/Berlin" || v.Config.TimeZone != "" {
		t.Fatalf("the server's zone: %q, %v", v.TimeZone, err)
	}
}

func TestNodeGroups(t *testing.T) {
	h := meshHarness(t)
	s, ctx := h.nodes(), context.Background()
	ch, err := s.Add(ctx, cli(), NodeAddRequest{Name: "box", Role: "spoke", Groups: []string{"office", "lab"}})
	if err != nil || !slices.Equal(ch.Node.Groups, []string{"lab", "office"}) {
		t.Fatalf("add with groups = %+v, %v", ch.Node, err)
	}
	none := []string{}
	if ch, err = s.Set(ctx, cli(), "box", NodeSetRequest{Groups: &none}); err != nil || len(ch.Node.Groups) != 0 {
		t.Fatalf("set no groups = %+v, %v", ch.Node, err)
	}
	office := []string{"office"}
	if _, err = s.Set(ctx, cli(), "box", NodeSetRequest{Groups: &office}); err != nil {
		t.Fatal(err)
	}
	l, err := s.List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, n := range l.Nodes {
		if (n.Name == "box") != slices.Equal(n.Groups, office) {
			t.Errorf("%s is in %v", n.Name, n.Groups)
		}
	}
	h.db.n.Store(0)
	_, err = s.Add(ctx, cli(), NodeAddRequest{Name: "box2", Role: "spoke", Groups: []string{"Office"}})
	refused(t, err, http.StatusBadRequest, "groups[0]")
	twice := []string{"lab", "lab"}
	_, err = s.Set(ctx, cli(), "box", NodeSetRequest{Groups: &twice})
	refused(t, err, http.StatusBadRequest, "listed twice")
	if n := h.db.n.Load(); n != 0 {
		t.Fatalf("a refused group ran %d transactions", n)
	}

	// A node a zone names cannot be removed; out of the zone, it can.
	zone := store.SegZone{Name: "lab", Intra: store.SegAllow, Members: []string{"node:box"}}
	if _, err := s.SaveZone(ctx, cli(), "", zone); err != nil {
		t.Fatal(err)
	}
	_, err = s.Remove(ctx, cli(), "box")
	refused(t, err, http.StatusConflict, `there is no node "box"`)
	if _, err := s.RemoveSegment(ctx, cli(), SegZone, "lab"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Remove(ctx, cli(), "box"); err != nil {
		t.Fatalf("remove once out of the zone: %v", err)
	}
}

// A mesh IP move names the firewall entries that held the old address.
func TestMoveMeshIPNotesTheFirewall(t *testing.T) {
	h := meshHarness(t)
	s, ctx := h.nodes(), context.Background()
	doc := officeFirewall()
	doc.Zones[0].Members = append(doc.Zones[0].Members, "10.77.1.0/30")
	if _, err := s.SetSegment(ctx, cli(), doc); err != nil {
		t.Fatal(err)
	}
	ip := "10.77.1.9"
	ch, err := s.Set(ctx, cli(), "nas", NodeSetRequest{MeshIP: &ip})
	if err != nil || !slices.ContainsFunc(ch.Notes, func(n string) bool {
		return n == `zone "home" names 10.77.1.0/30 by address, which held the old one; node:nas follows the node`
	}) {
		t.Fatalf("notes = %q, %v", ch.Notes, err)
	}
}

func TestSegmentAdminRoutes(t *testing.T) {
	h := meshHarness(t)
	if code, eb := adminCall(t, h, http.MethodPut, AdminSegmentation, officeFirewall(), nil); code != http.StatusOK {
		t.Fatalf("PUT = %d %+v", code, eb)
	}
	var v SegmentView
	if code, _ := adminCall(t, h, http.MethodGet, AdminSegmentation, nil, &v); code != http.StatusOK || len(v.Config.Policies) != 2 {
		t.Fatalf("GET = %d %+v", code, v)
	}
	q := url.Values{"from": {"laptop"}, "to": {"nas"}, "service": {"tcp/445"}}
	var r ReachResult
	if code, _ := adminCall(t, h, http.MethodGet, AdminSegmentation+"/check?"+q.Encode(), nil, &r); code != http.StatusOK || !r.Allowed ||
		r.Rule != "policy: smb" || r.Index != 3 {
		t.Fatalf("check = %d %+v", code, r)
	}
	q.Set("at", "2026-10-12T09:00:00Z")
	if code, _ := adminCall(t, h, http.MethodGet, AdminSegmentation+"/check?"+q.Encode(), nil, &r); code != http.StatusOK ||
		r.At != "2026-10-12T09:00:00Z" {
		t.Fatalf("check at = %d %+v", code, r)
	}
	bad := officeFirewall()
	bad.Zones[0].Intra = "maybe"
	if code, eb := adminCall(t, h, http.MethodPut, AdminSegmentation, bad, nil); code != http.StatusBadRequest || eb.Code != proto.CodeBadRequest {
		t.Fatalf("a bad PUT = %d %+v", code, eb)
	}
}
