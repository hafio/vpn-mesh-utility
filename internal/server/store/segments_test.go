package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/hafio/vpn-mesh-utility/internal/testfix"
)

// testSegmentation is a mesh firewall over the reference topology, in the
// order Segmentation reads it back (by name; policies in their order).
func testSegmentation() Segmentation {
	return Segmentation{
		Groups:    []SegGroup{{Name: "laptops", Nodes: []string{"laptop"}}, {Name: "storage", Description: "the NAS", Nodes: []string{"nas"}}},
		Addresses: []SegObject{{Name: "printers", Entries: []string{"10.77.1.50", "10.77.1.60-10.77.1.70"}}},
		Services:  []SegObject{{Name: "smb", Entries: []string{"tcp/445", "tcp/139"}}},
		Zones:     []SegZone{{Name: "home", Intra: SegAllow, Members: []string{"group:storage", "addr:printers"}}},
		Policies: []SegPolicy{
			{Name: "laptop-smb", From: []string{"group:laptops"}, To: []string{"zone:home"}, Services: []string{"svc:smb"}, Action: SegAllow,
				Schedule: &SegSchedule{Days: []string{"mon", "tue", "wed", "thu", "fri"}, Start: "08:00", End: "19:00"}},
			{Name: "old", Disabled: true, From: []string{"10.77.2.0/24"}, Action: SegDeny, Schedule: &SegSchedule{Days: []string{"sat"}}},
			DefaultPolicy(),
		},
		TimeZone: "Asia/Singapore",
	}
}

func TestSegmentationRoundTrip(t *testing.T) {
	s, _, _ := openTest(t)
	insertTopology(t, s)
	tx(t, s, func(t2 *Tx) error {
		got, err := t2.Segmentation()
		if err != nil {
			return err
		}
		// A new mesh starts with the default policy (migration 0006).
		if want := (Segmentation{Policies: []SegPolicy{DefaultPolicy()}}); !reflect.DeepEqual(got, want) {
			t.Errorf("a new mesh's firewall = %+v, want %+v", got, want)
		}
		if err := t2.SetSegmentation(testSegmentation()); err != nil {
			return err
		}
		if got, err = t2.Segmentation(); err != nil {
			return err
		}
		if !reflect.DeepEqual(got, testSegmentation()) {
			t.Errorf("read back\n%+v\nwant\n%+v", got, testSegmentation())
		}
		groups, err := t2.NodeGroups()
		if err != nil {
			return err
		}
		nas, _ := t2.NodeByName("nas")
		if len(groups[nas.ID]) != 1 || groups[nas.ID][0] != "storage" {
			t.Errorf("NodeGroups = %v", groups)
		}
		if err := t2.SetSegmentation(Segmentation{}); err != nil {
			return err
		}
		if got, err = t2.Segmentation(); err != nil || !got.Empty() {
			t.Errorf("an empty firewall reads %+v, %v", got, err)
		}
		return nil
	})
}

func TestSegScheduleString(t *testing.T) {
	for _, c := range []struct {
		s    SegSchedule
		want string
	}{
		{SegSchedule{Days: []string{"mon", "tue", "wed", "thu", "fri"}, Start: "08:00", End: "19:00"}, "mon-fri 08:00-19:00"},
		{SegSchedule{Days: []string{"sat"}}, "sat"},
		{SegSchedule{Start: "8am", End: "19:00"}, "8am-19:00"},
	} {
		if got := c.s.String(); got != c.want {
			t.Errorf("%+v reads %q, want %q", c.s, got, c.want)
		}
	}
}

// A group naming a node that does not exist is refused, and nothing of the
// change stays.
func TestSetSegmentationRefusesAStranger(t *testing.T) {
	s, _, _ := openTest(t)
	insertTopology(t, s)
	seg := testSegmentation()
	seg.Groups[0].Nodes = []string{"ghost"}
	err := s.Tx(context.Background(), func(t2 *Tx) error { return t2.SetSegmentation(seg) })
	if !errors.Is(err, ErrNotFound) || !strings.Contains(err.Error(), "ghost") {
		t.Fatalf("SetSegmentation with a stranger = %v", err)
	}
	tx(t, s, func(t2 *Tx) error {
		got, err := t2.Segmentation()
		if err == nil && len(got.Policies) != 1 {
			t.Errorf("the refused change stayed: %+v", got)
		}
		return err
	})
}

func TestSetNodeGroups(t *testing.T) {
	s, _, _ := openTest(t)
	nodes := insertTopology(t, s)
	nas := nodes["nas"].ID
	tx(t, s, func(t2 *Tx) error {
		if err := t2.SetNodeGroups(nas, []string{"storage", "backups"}); err != nil {
			return err
		}
		if err := t2.SetNodeGroups(nodes["laptop"].ID, []string{"backups"}); err != nil {
			return err
		}
		groups, err := t2.NodeGroups()
		if err != nil {
			return err
		}
		if !reflect.DeepEqual(groups[nas], []string{"backups", "storage"}) || !reflect.DeepEqual(groups[nodes["laptop"].ID], []string{"backups"}) {
			t.Errorf("NodeGroups = %v", groups)
		}
		seg, err := t2.Segmentation()
		if err != nil {
			return err
		}
		if len(seg.Groups) != 2 || seg.Groups[0].Name != "backups" || !reflect.DeepEqual(seg.Groups[0].Nodes, []string{"laptop", "nas"}) {
			t.Errorf("groups made on the way = %+v", seg.Groups)
		}
		// Out of every group; the groups stay.
		if err := t2.SetNodeGroups(nas, nil); err != nil {
			return err
		}
		groups, err = t2.NodeGroups()
		if err == nil && len(groups[nas]) != 0 {
			t.Errorf("the NAS is still in %v", groups[nas])
		}
		return err
	})
}

func TestSegmentationValidate(t *testing.T) {
	if err := testSegmentation().Validate(); err != nil {
		t.Fatalf("the test firewall is invalid: %v", err)
	}
	for name, c := range map[string]struct {
		edit func(*Segmentation)
		want string
	}{
		"group name":             {func(s *Segmentation) { s.Groups[0].Name = "Laptops" }, "groups[0].name"},
		"group twice":            {func(s *Segmentation) { s.Groups[1].Name = s.Groups[0].Name }, "defined twice"},
		"group node":             {func(s *Segmentation) { s.Groups[0].Nodes = []string{"lap top"} }, "groups[laptops].nodes[0]"},
		"group node twice":       {func(s *Segmentation) { s.Groups[0].Nodes = []string{"laptop", "laptop"} }, "listed twice"},
		"object without entries": {func(s *Segmentation) { s.Addresses[0].Entries = nil }, "at least one entry"},
		"address entry":          {func(s *Segmentation) { s.Addresses[0].Entries = []string{"10.77.1.5/24"} }, "addresses[printers].entries[0]"},
		"service entry":          {func(s *Segmentation) { s.Services[0].Entries = []string{"tcp/http"} }, "services[smb].entries[0]"},
		"zone within":            {func(s *Segmentation) { s.Zones[0].Intra = "maybe" }, "zones[home].intra"},
		"zone in a zone":         {func(s *Segmentation) { s.Zones[0].Members = []string{"zone:home"} }, "zones[home].members[0]"},
		"zone member":            {func(s *Segmentation) { s.Zones[0].Members = []string{"10.77.1.5;reboot"} }, "zones[home].members[0]"},
		"policy action":          {func(s *Segmentation) { s.Policies[0].Action = "accept" }, "policies[laptop-smb].action"},
		"service in from":        {func(s *Segmentation) { s.Policies[0].From = []string{"svc:smb"} }, "policies[laptop-smb].from[0]"},
		"policy service":         {func(s *Segmentation) { s.Policies[0].Services = []string{"addr:printers"} }, "policies[laptop-smb].services[0]"},
		"description newline":    {func(s *Segmentation) { s.Groups[1].Description = "a\nb" }, "groups[storage].description"},
		"schedule day":           {func(s *Segmentation) { s.Policies[0].Schedule.Days = []string{"funday"} }, "policies[laptop-smb].schedule.days[0]"},
		"schedule window":        {func(s *Segmentation) { s.Policies[0].Schedule.Start = "8am" }, "policies[laptop-smb].schedule.start"},
		"time zone":              {func(s *Segmentation) { s.TimeZone = "Mars/Olympus" }, "timezone"},
		"too many refs": {func(s *Segmentation) {
			for len(s.Policies[0].To) <= MaxSegRefs {
				s.Policies[0].To = append(s.Policies[0].To, "zone:home")
			}
		}, "at most 32"},
		"too many policies": {func(s *Segmentation) {
			for i := len(s.Policies); i <= MaxSegPolicies; i++ {
				p := DefaultPolicy()
				p.Name = fmt.Sprintf("p%d", i)
				s.Policies = append(s.Policies, p)
			}
		}, "at most 256"},
	} {
		seg := testSegmentation()
		c.edit(&seg)
		if err := seg.Validate(); err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("%s: Validate = %v, want an error about %q", name, err, c.want)
		}
	}
}

// openAt is a store migrated up to version only.
func openAt(t *testing.T, version int) *Store {
	t.Helper()
	db, err := sql.Open("sqlite", filepath.Join(t.TempDir(), "mesh.db")+dsnParams)
	if err != nil {
		t.Fatal(err)
	}
	db.SetMaxOpenConns(1)
	s := &Store{db: db, now: testfix.NewClock(testfix.Epoch).Now, log: slog.New(slog.DiscardHandler)}
	t.Cleanup(func() { _ = s.Close() })
	ms, err := loadMigrations(migrationFS)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.migrateTo(context.Background(), ms[:version]); err != nil {
		t.Fatal(err)
	}
	return s
}

// Migration 0006 turns the custom firewall rules into policies exactly as
// LegacySegmentation does for older snapshots and exports, and drops them.
func TestMigrationTurnsRulesIntoPolicies(t *testing.T) {
	s := openAt(t, 5)
	nodes := insertTopology(t, s)
	rules := []FirewallRule{
		{ID: 1, NodeID: nodes["vps-1"].ID, Proto: "tcp", Port: 9000, SourceCIDR: "10.77.1.0/24", Note: "backup"},
		{ID: 2, Proto: "udp", Port: 161, SourceCIDR: "0.0.0.0/0"},
	}
	tx(t, s, func(t2 *Tx) error {
		for _, r := range rules {
			if _, err := t2.exec(`INSERT INTO firewall_rules (id, node_id, proto, port, source_cidr, note) VALUES (?, ?, ?, ?, ?, ?)`,
				r.ID, nullInt(r.NodeID), r.Proto, r.Port, r.SourceCIDR, r.Note); err != nil {
				return err
			}
		}
		return nil
	})
	if err := s.migrate(context.Background()); err != nil {
		t.Fatal(err)
	}
	names := map[int64]string{nodes["vps-1"].ID: "vps-1"}
	want := LegacySegmentation(rules, names)
	if want.Policies[0].To[0] != "node:vps-1" || want.Policies[1].To[0] != "role:server" || want.Policies[2].Name != "allow-all" {
		t.Fatalf("LegacySegmentation = %+v", want)
	}
	tx(t, s, func(t2 *Tx) error {
		got, err := t2.Segmentation()
		if err != nil {
			return err
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("after migration 0006\n%+v\nwant\n%+v", got, want)
		}
		var n int
		err = t2.queryRow(`SELECT COUNT(*) FROM sqlite_master WHERE name = 'firewall_rules'`).Scan(&n)
		if err == nil && n != 0 {
			t.Error("the firewall_rules table survived the migration")
		}
		return err
	})
}

// A snapshot made before the mesh firewall restores its custom rules as
// policies.
func TestRestoreOfASnapshotWithFirewallRules(t *testing.T) {
	s, _, _ := openTest(t)
	nodes := insertTopology(t, s)
	tx(t, s, func(t2 *Tx) error {
		snap, err := t2.Snapshot()
		if err != nil {
			return err
		}
		// What one looked like then: custom rules, no segmentation key.
		snap.Segmentation = nil
		snap.Firewall = []FirewallRule{{ID: 7, NodeID: nodes["vps-1"].ID, Proto: "tcp", Port: 22, SourceCIDR: "10.77.2.0/24"}}
		b, err := snap.Encode()
		if err != nil {
			return err
		}
		old := strings.Replace(string(b), `,"segmentation":null`, "", 1)
		if old == string(b) {
			t.Fatalf("no segmentation key to take out of %s", b)
		}
		legacy, err := DecodeSnapshot([]byte(old))
		if err != nil {
			return err
		}
		if legacy.Segmentation != nil || len(legacy.Firewall) != 1 {
			t.Fatalf("the old snapshot reads %+v", legacy)
		}
		if err := t2.Restore(legacy); err != nil {
			return err
		}
		got, err := t2.Segmentation()
		if err != nil {
			return err
		}
		if len(got.Policies) != 2 || got.Policies[0].Name != "firewall-7" || got.Policies[0].To[0] != "node:"+nodes["vps-1"].Name ||
			got.Policies[0].Services[0] != "tcp/22" || got.Policies[1].Name != "allow-all" {
			t.Errorf("restored firewall = %+v", got)
		}
		return nil
	})
}
