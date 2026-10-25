package store

import (
	"context"
	"errors"
	"regexp"
	"testing"

	"github.com/hafio/vpn-mesh-utility/internal/testfix"
)

var reUUID = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

// addMesh creates a mesh with the first mesh's network settings, so a test
// can put the same names and addresses in both.
func addMesh(t *testing.T, s *Store, name string, slot int) Mesh {
	t.Helper()
	m := Mesh{Name: name, Slot: slot}
	tx(t, s, func(t2 *MeshTx) error {
		if err := t2.InsertMesh(&m); err != nil {
			return err
		}
		return t2.In(m.ID).SetSetting(SettingMeshCIDR, "10.77.0.0/16")
	})
	return m
}

// The first mesh is "default" in slot 0; another gets its own uuid, and a
// taken name or slot is a conflict.
func TestMeshes(t *testing.T) {
	s, _, _ := openTest(t)
	lab := addMesh(t, s, "lab", 1)
	if lab.ID == FirstMesh || !reUUID.MatchString(lab.UUID) || lab.CreatedAt != testfix.Epoch.Unix() {
		t.Fatalf("InsertMesh = %+v", lab)
	}
	tx(t, s, func(t2 *MeshTx) error {
		all, err := t2.Meshes()
		if err != nil {
			return err
		}
		if len(all) != 2 || all[0].ID != FirstMesh || all[0].Name != "default" || all[0].Slot != 0 || !reUUID.MatchString(all[0].UUID) ||
			all[1] != lab || all[0].UUID == lab.UUID {
			t.Errorf("Meshes = %+v", all)
		}
		if m, err := t2.MeshByName("lab"); err != nil || m != lab {
			t.Errorf("MeshByName(lab) = %+v, %v", m, err)
		}
		if m, err := t2.MeshByID(lab.ID); err != nil || m != lab {
			t.Errorf("MeshByID = %+v, %v", m, err)
		}
		if _, err := t2.MeshByName("none"); !errors.Is(err, ErrNotFound) {
			t.Errorf("MeshByName(none) = %v", err)
		}
		if _, err := t2.MeshByID(99); !errors.Is(err, ErrNotFound) {
			t.Errorf("MeshByID(99) = %v", err)
		}
		return nil
	})
	for _, m := range []Mesh{{Name: "lab", Slot: 2}, {Name: "home", Slot: 1}} {
		err := s.Tx(context.Background(), func(t2 *Tx) error { return t2.InsertMesh(&m) })
		if !errors.Is(err, ErrConflict) {
			t.Errorf("InsertMesh(%+v) = %v, want a conflict", m, err)
		}
	}
}

// Two meshes hold the same names and addresses without seeing each
// other's: nodes, settings, records, versions, segmentation, container
// names and audit entries.
func TestMeshesAreApart(t *testing.T) {
	s, _, _ := openTest(t)
	nodes := insertTopology(t, s)
	lab := addMesh(t, s, "lab", 1)
	var labNAS Node
	err := s.MeshTx(context.Background(), lab.ID, func(t2 *MeshTx) error {
		labNAS = node(testfix.NAS)
		labNAS.WGPubKey, labNAS.IdentityPubKey = "", "" // keys are unique across meshes
		if err := t2.InsertNode(&labNAS); err != nil {
			return err
		}
		if err := t2.SetSetting(SettingDNSZone, "lab"); err != nil {
			return err
		}
		if err := t2.InsertDNSRecord(&DNSRecord{Name: "printer", IP: "10.77.1.50"}); err != nil {
			return err
		}
		if _, err := t2.InsertVersion(Version{Actor: "cli", Summary: "add nas", Snapshot: []byte(`{}`)}); err != nil {
			return err
		}
		seg := Segmentation{Groups: []SegGroup{{Name: "storage", Nodes: []string{"nas"}}}, Policies: []SegPolicy{DefaultPolicy()}}
		if err := t2.SetSegmentation(seg); err != nil {
			return err
		}
		if _, err := t2.SetContainerNames(labNAS.ID, []ContainerName{{Name: "web", IP: "10.77.128.129", Kind: "service"}}); err != nil {
			return err
		}
		return t2.Audit("cli", 0, "node.add", "nas", nil)
	})
	if err != nil {
		t.Fatal(err)
	}
	if labNAS.MeshID != lab.ID || labNAS.ID == nodes["nas"].ID {
		t.Fatalf("the lab's nas = %+v", labNAS)
	}
	tx(t, s, func(t2 *MeshTx) error {
		if err := t2.InsertDNSRecord(&DNSRecord{Name: "printer", IP: "10.77.1.51"}); err != nil {
			return err
		}
		if err := t2.Tx.Audit("cli", 0, "user.add", "alice", nil); err != nil {
			return err
		}
		all, err := t2.Nodes()
		if err != nil {
			return err
		}
		if len(all) != 5 || all[3].ID != nodes["nas"].ID || all[3].MeshID != FirstMesh {
			t.Errorf("the first mesh's nodes = %+v", all)
		}
		if zone, err := t2.Setting(SettingDNSZone); err != nil || zone != "mesh" {
			t.Errorf("the first mesh's zone = %q, %v", zone, err)
		}
		if v, err := t2.LatestVersion(); err != nil || v != 0 {
			t.Errorf("the first mesh's latest version = %d, %v", v, err)
		}
		if names, err := t2.ContainerNames(); err != nil || len(names) != 0 {
			t.Errorf("the first mesh's container names = %+v, %v", names, err)
		}
		seg, err := t2.Segmentation()
		if err != nil {
			return err
		}
		if len(seg.Groups) != 0 {
			t.Errorf("the first mesh's groups = %+v", seg.Groups)
		}
		log, err := t2.AuditLog(10, 0)
		if err != nil {
			return err
		}
		if len(log) != 1 || log[0].Action != "user.add" || log[0].MeshID != 0 {
			t.Errorf("the first mesh's audit = %+v", log)
		}
		return nil
	})
	err = s.MeshTx(context.Background(), lab.ID, func(t2 *MeshTx) error {
		if n, err := t2.NodeByName("nas"); err != nil || n.ID != labNAS.ID {
			t.Errorf("the lab's NodeByName(nas) = %+v, %v", n, err)
		}
		if _, err := t2.NodeByName("vps-1"); !errors.Is(err, ErrNotFound) {
			t.Errorf("the lab sees vps-1: %v", err)
		}
		if recs, err := t2.DNSRecords(); err != nil || len(recs) != 1 || recs[0].IP != "10.77.1.50" {
			t.Errorf("the lab's records = %+v, %v", recs, err)
		}
		if v, err := t2.LatestVersion(); err != nil || v != 1 {
			t.Errorf("the lab's latest version = %d, %v", v, err)
		}
		if groups, err := t2.NodeGroups(); err != nil || len(groups[labNAS.ID]) != 1 {
			t.Errorf("the lab's node groups = %+v, %v", groups, err)
		}
		if names, err := t2.ContainerNames(); err != nil || len(names) != 1 {
			t.Errorf("the lab's container names = %+v, %v", names, err)
		}
		log, err := t2.AuditLog(10, 0)
		if err != nil {
			return err
		}
		if len(log) != 2 || log[0].Action != "user.add" || log[1].Action != "node.add" || log[1].MeshID != lab.ID {
			t.Errorf("the lab's audit = %+v", log)
		}
		all, err := t2.Tx.AuditLog(10, 0)
		if err != nil || len(all) != 2 {
			t.Errorf("the whole audit = %+v, %v", all, err)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	// A mesh with nodes stays; without them it goes with everything in it.
	err = s.Tx(context.Background(), func(t2 *Tx) error { return t2.DeleteMesh(lab.ID) })
	if !errors.Is(err, ErrConflict) {
		t.Fatalf("DeleteMesh with a node = %v, want a conflict", err)
	}
	tx(t, s, func(t2 *MeshTx) error {
		if err := t2.DeleteNode(labNAS.ID); err != nil {
			return err
		}
		if err := t2.DeleteMesh(lab.ID); err != nil {
			return err
		}
		for _, table := range []string{"mesh_settings", "dns_records", "config_versions", "seg_groups", "seg_policies", "audit_log"} {
			var n int
			if err := t2.queryRow(`SELECT COUNT(*) FROM `+table+` WHERE mesh_id = ?`, lab.ID).Scan(&n); err != nil {
				return err
			}
			if n != 0 {
				t.Errorf("%d rows of %s outlived their mesh", n, table)
			}
		}
		if err := t2.DeleteMesh(lab.ID); !errors.Is(err, ErrNotFound) {
			t.Errorf("DeleteMesh twice = %v", err)
		}
		return nil
	})
}

// A removed node's id is never given out again, so restoring a snapshot
// that has it cannot collide with a node another mesh added since.
func TestNodeIDsAreNeverReused(t *testing.T) {
	s, _, _ := openTest(t)
	nodes := insertTopology(t, s)
	laptop := nodes["laptop"]
	var snap Snapshot
	tx(t, s, func(t2 *MeshTx) error {
		var err error
		if snap, err = t2.Snapshot(); err != nil {
			return err
		}
		return t2.DeleteNode(laptop.ID)
	})
	lab := addMesh(t, s, "lab", 1)
	var other Node
	err := s.MeshTx(context.Background(), lab.ID, func(t2 *MeshTx) error {
		other = node(testfix.Phone)
		return t2.InsertNode(&other)
	})
	if err != nil {
		t.Fatal(err)
	}
	if other.ID <= laptop.ID {
		t.Fatalf("a new node got id %d after id %d was removed", other.ID, laptop.ID)
	}
	tx(t, s, func(t2 *MeshTx) error {
		if err := t2.Restore(snap); err != nil {
			return err
		}
		n, err := t2.NodeByName("laptop")
		if err != nil || n.ID != laptop.ID || n.MeshID != FirstMesh {
			t.Errorf("restored laptop = %+v, %v", n, err)
		}
		return nil
	})
}

// A join token leads to its node's mesh before it is consumed.
func TestTokenMesh(t *testing.T) {
	s, _, _ := openTest(t)
	nodes := insertTopology(t, s)
	tx(t, s, func(t2 *MeshTx) error {
		tok, _, err := t2.IssueToken(nodes["nas"].ID, 0)
		if err != nil {
			return err
		}
		if mesh, err := t2.TokenMesh(tok); err != nil || mesh != FirstMesh {
			t.Errorf("TokenMesh = %d, %v", mesh, err)
		}
		if _, err := t2.TokenMesh("not-a-token"); !errors.Is(err, ErrTokenUnknown) {
			t.Errorf("TokenMesh of an unknown token = %v", err)
		}
		return nil
	})
}

// The control plane's own settings are apart from every mesh's.
func TestServerSettings(t *testing.T) {
	s, _, _ := openTest(t)
	tx(t, s, func(t2 *MeshTx) error {
		if v, err := t2.ServerSetting(SettingServerTZ); err != nil || v != "" {
			t.Errorf("an unset server setting = %q, %v", v, err)
		}
		for _, v := range []string{"UTC", "Asia/Singapore"} {
			if err := t2.SetServerSetting(SettingServerTZ, v); err != nil {
				return err
			}
		}
		if v, err := t2.ServerSetting(SettingServerTZ); err != nil || v != "Asia/Singapore" {
			t.Errorf("ServerSetting = %q, %v", v, err)
		}
		if _, err := t2.Setting(SettingServerTZ); !errors.Is(err, ErrNotFound) {
			t.Errorf("the server's zone is a mesh setting too: %v", err)
		}
		return nil
	})
}

// Migration 0009 makes an earlier database's single mesh mesh 1: its
// settings, records, versions, groups, policies and audit entries carry
// over, the deployment's own settings stay the control plane's, and a new
// node's id follows the existing ones.
func TestMigrationMakesMeshOne(t *testing.T) {
	s := openAt(t, 8)
	ids := insertAt5(t, s)
	tx(t, s, func(t2 *MeshTx) error {
		for _, q := range []struct {
			sql  string
			args []any
		}{
			{`INSERT INTO settings (key, value) VALUES ('server_tz', 'Europe/Berlin')`, nil},
			{`INSERT INTO dns_records (name, ip, created_at) VALUES ('printer', '10.77.1.50', 1)`, nil},
			{`INSERT INTO config_versions (created_at, actor, summary, snapshot_json) VALUES (1, 'cli', 'add nas', '{}')`, nil},
			{`INSERT INTO seg_groups (id, name) VALUES (7, 'storage')`, nil},
			{`INSERT INTO node_groups (node_id, group_id) VALUES (?, 7)`, []any{ids["nas"]}},
			{`INSERT INTO audit_log (at, actor, action) VALUES (1, 'cli', 'node.add')`, nil},
			{`INSERT INTO audit_log (at, actor, action) VALUES (2, 'cli', 'user.add')`, nil},
		} {
			if _, err := t2.exec(q.sql, q.args...); err != nil {
				return err
			}
		}
		return nil
	})
	if err := s.migrate(context.Background()); err != nil {
		t.Fatal(err)
	}
	tx(t, s, func(t2 *MeshTx) error {
		meshes, err := t2.Meshes()
		if err != nil {
			return err
		}
		if len(meshes) != 1 || meshes[0].Name != "default" || meshes[0].Slot != 0 || !reUUID.MatchString(meshes[0].UUID) {
			t.Errorf("meshes = %+v", meshes)
		}
		if cidr, err := t2.Setting(SettingMeshCIDR); err != nil || cidr != "10.77.0.0/16" {
			t.Errorf("mesh_cidr = %q, %v", cidr, err)
		}
		if tz, err := t2.ServerSetting(SettingServerTZ); err != nil || tz != "Europe/Berlin" {
			t.Errorf("server_tz = %q, %v", tz, err)
		}
		if recs, err := t2.DNSRecords(); err != nil || len(recs) != 1 || recs[0].Name != "printer" {
			t.Errorf("records = %+v, %v", recs, err)
		}
		if v, err := t2.VersionByNumber(1); err != nil || v.Summary != "add nas" {
			t.Errorf("v1 = %+v, %v", v, err)
		}
		seg, err := t2.Segmentation()
		if err != nil {
			return err
		}
		if len(seg.Groups) != 1 || len(seg.Groups[0].Nodes) != 1 || seg.Groups[0].Nodes[0] != "nas" || len(seg.Policies) != 1 ||
			seg.Policies[0].Name != "allow-all" {
			t.Errorf("segmentation = %+v", seg)
		}
		log, err := t2.Tx.AuditLog(10, 0)
		if err != nil {
			return err
		}
		if len(log) != 2 || log[0].MeshID != 0 || log[1].MeshID != FirstMesh {
			t.Errorf("audit = %+v", log)
		}
		n := Node{Name: "pc", Role: "spoke", MeshIP: "10.77.2.20", ListenPort: 51820, Keepalive: 25, MTU: 1420}
		if err := t2.InsertNode(&n); err != nil {
			return err
		}
		if n.ID <= ids["laptop"] {
			t.Errorf("a new node got id %d; the newest was %d", n.ID, ids["laptop"])
		}
		return nil
	})
}

// A record is removed only from its own mesh: another mesh's id is not
// found there, and the record stays.
func TestRecordsStayInTheirMesh(t *testing.T) {
	s, _, _ := openTest(t)
	lab := addMesh(t, s, "lab", 1)
	rec := DNSRecord{Name: "printer", IP: "10.77.1.50"}
	err := s.MeshTx(context.Background(), lab.ID, func(t2 *MeshTx) error {
		if t2.MeshID() != lab.ID {
			t.Errorf("MeshID = %d, want %d", t2.MeshID(), lab.ID)
		}
		return t2.InsertDNSRecord(&rec)
	})
	if err != nil {
		t.Fatal(err)
	}
	tx(t, s, func(t2 *MeshTx) error {
		if err := t2.DeleteDNSRecord(rec.ID); !errors.Is(err, ErrNotFound) {
			t.Errorf("the first mesh removed the lab's record: %v", err)
		}
		return nil
	})
	err = s.MeshTx(context.Background(), lab.ID, func(t2 *MeshTx) error {
		if err := t2.DeleteDNSRecord(rec.ID); err != nil {
			return err
		}
		recs, err := t2.DNSRecords()
		if len(recs) != 0 {
			t.Errorf("records left: %+v", recs)
		}
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
}
