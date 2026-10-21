package store

import (
	"context"
	"errors"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"
	"testing/fstest"
	"time"

	"github.com/hafio/vpn-mesh-utility/internal/proto"
	"github.com/hafio/vpn-mesh-utility/internal/testfix"
)

func openTest(t *testing.T) (*Store, *testfix.Clock, string) {
	t.Helper()
	clock := testfix.NewClock(testfix.Epoch)
	path := filepath.Join(t.TempDir(), "mesh.db")
	s, err := Open(context.Background(), path, Options{Now: clock.Now})
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s, clock, path
}

func tx(t *testing.T, s *Store, fn func(*Tx) error) {
	t.Helper()
	if err := s.Tx(context.Background(), fn); err != nil {
		t.Fatal(err)
	}
}

// node builds a store row from a reference-topology node.
func node(f testfix.Node) Node {
	return Node{
		Name: f.Name, Role: string(f.Role), MeshIP: f.MeshIP, Endpoint: f.Endpoint,
		WGPubKey: testfix.WGKey(f.Key), IdentityPubKey: testfix.IdentityKeyB64(f.Key),
		ListenPort: 51820, Keepalive: 25, MTU: 1420, Enrolled: true,
	}
}

// insertTopology writes the reference topology: three servers, the NAS (hub
// vps-1) and the laptop (exit via vps-2). It returns rows by name.
func insertTopology(t *testing.T, s *Store) map[string]Node {
	t.Helper()
	out := map[string]Node{}
	tx(t, s, func(t2 *Tx) error {
		for _, f := range []testfix.Node{testfix.VPS1, testfix.VPS2, testfix.VPS3, testfix.NAS, testfix.Laptop} {
			n := node(f)
			if f.Role != "server" {
				n.PrimaryHubID = out["vps-1"].ID
			}
			if f.Name == "laptop" {
				n.ExitVia = out["vps-2"].ID
			}
			if f.Name == "vps-2" {
				n.CanExit = true
			}
			if err := t2.InsertNode(&n); err != nil {
				return err
			}
			out[n.Name] = n
		}
		return nil
	})
	return out
}

func TestOpenAppliesMigrations(t *testing.T) {
	s, _, _ := openTest(t)
	tx(t, s, func(t2 *Tx) error {
		v, err := t2.SchemaVersion()
		if err != nil {
			return err
		}
		if v != 6 {
			t.Errorf("schema version = %d, want 6", v)
		}
		cidr, err := t2.Setting(SettingMeshCIDR)
		if err != nil {
			return err
		}
		zone, err := t2.Setting(SettingDNSZone)
		if err != nil {
			return err
		}
		if cidr != "10.77.0.0/16" || zone != "mesh" {
			t.Errorf("default mesh_cidr = %q, dns_zone = %q", cidr, zone)
		}
		return nil
	})
}

// A node's mode and container subnet (amendment A37): host and none by
// default, kept by insert and update, one node per subnet, and a snapshot
// that names the mode of a container-mode node only, so a mesh without
// container networks snapshots as before.
func TestNodeContainerColumns(t *testing.T) {
	s, _, _ := openTest(t)
	insertTopology(t, s)
	tx(t, s, func(t2 *Tx) error {
		nas, err := t2.NodeByName("nas")
		if err != nil {
			return err
		}
		if nas.Mode != proto.ModeHost || nas.ContainerSubnet != "" {
			t.Errorf("defaults: mode %q, subnet %q", nas.Mode, nas.ContainerSubnet)
		}
		nas.ContainerSubnet = testfix.ContainerSubnet
		if err := t2.UpdateNode(nas); err != nil {
			return err
		}
		box := Node{Name: "box", Role: "spoke", MeshIP: "10.77.2.9", Mode: proto.ModeContainer, ContainerSubnet: "10.77.129.0/24"}
		if err := t2.InsertNode(&box); err != nil {
			return err
		}
		for name, want := range map[string][2]string{"nas": {proto.ModeHost, testfix.ContainerSubnet}, "box": {proto.ModeContainer, "10.77.129.0/24"}} {
			got, err := t2.NodeByName(name)
			if err != nil {
				return err
			}
			if got.Mode != want[0] || got.ContainerSubnet != want[1] {
				t.Errorf("%s: mode %q, subnet %q, want %v", name, got.Mode, got.ContainerSubnet, want)
			}
		}
		dup := Node{Name: "box2", Role: "spoke", MeshIP: "10.77.2.10", ContainerSubnet: testfix.ContainerSubnet}
		if err := t2.InsertNode(&dup); !errors.Is(err, ErrConflict) {
			t.Errorf("a second node with %s: %v, want ErrConflict", testfix.ContainerSubnet, err)
		}
		snap, err := t2.Snapshot()
		if err != nil {
			return err
		}
		b, err := snap.Encode()
		if err != nil {
			return err
		}
		if n := strings.Count(string(b), `"mode":`); n != 1 || !strings.Contains(string(b), `"mode":"container"`) ||
			!strings.Contains(string(b), `"container_subnet":"`+testfix.ContainerSubnet+`"`) {
			t.Errorf("snapshot: %d mode keys, want only box's:\n%s", n, b)
		}
		return nil
	})
}

func TestOpenIsIdempotent(t *testing.T) {
	s, clock, path := openTest(t)
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	s2, err := Open(context.Background(), path, Options{Now: clock.Now})
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	defer s2.Close()
	tx(t, s2, func(t2 *Tx) error {
		v, err := t2.SchemaVersion()
		if v != 6 {
			t.Errorf("schema version after reopen = %d, want 6", v)
		}
		return err
	})
}

func TestOpenRefusesNewerSchema(t *testing.T) {
	s, clock, path := openTest(t)
	tx(t, s, func(t2 *Tx) error {
		_, err := t2.exec(`INSERT INTO schema_version (version, applied_at) VALUES (99, 0)`)
		return err
	})
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	_, err := Open(context.Background(), path, Options{Now: clock.Now})
	if err == nil || !strings.Contains(err.Error(), "v99") || !strings.Contains(err.Error(), "upgrade meshctl") {
		t.Fatalf("Open on a newer schema = %v, want an actionable refusal", err)
	}
}

func TestOpenRejectsDSNInjection(t *testing.T) {
	_, err := Open(context.Background(), filepath.Join(t.TempDir(), "x.db?mode=memory"), Options{})
	if err == nil {
		t.Fatal("Open accepted a path that rewrites the DSN")
	}
}

func TestLoadMigrationsValidatesNames(t *testing.T) {
	cases := map[string]fstest.MapFS{
		"gap":      {"migrations/0001_a.sql": {Data: []byte("SELECT 1;")}, "migrations/0003_c.sql": {Data: []byte("SELECT 1;")}},
		"bad name": {"migrations/init.sql": {Data: []byte("SELECT 1;")}},
		"zero":     {"migrations/0000_a.sql": {Data: []byte("SELECT 1;")}},
	}
	for name, fsys := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := loadMigrations(fsys); err == nil {
				t.Fatal("loadMigrations accepted a broken migration set")
			}
		})
	}
	ms, err := loadMigrations(migrationFS)
	if err != nil || len(ms) == 0 || ms[0].version != 1 {
		t.Fatalf("embedded migrations: %v, %+v", err, ms)
	}
}

func TestNodeCRUD(t *testing.T) {
	s, _, _ := openTest(t)
	nodes := insertTopology(t, s)
	tx(t, s, func(t2 *Tx) error {
		got, err := t2.NodeByName("nas")
		if err != nil {
			return err
		}
		if got.PrimaryHubID != nodes["vps-1"].ID || got.MeshIP != "10.77.1.1" || !got.Enrolled {
			t.Errorf("NodeByName(nas) = %+v", got)
		}
		got.MTU = 1380
		got.Disabled = true
		if err := t2.UpdateNode(got); err != nil {
			return err
		}
		again, err := t2.NodeByID(got.ID)
		if err != nil {
			return err
		}
		if again.MTU != 1380 || !again.Disabled {
			t.Errorf("UpdateNode did not persist: %+v", again)
		}
		all, err := t2.Nodes()
		if err != nil {
			return err
		}
		if len(all) != 5 || all[0].Name != "vps-1" {
			t.Errorf("Nodes() = %d rows, first %q", len(all), all[0].Name)
		}
		return nil
	})
}

func TestNodeUniquenessIsAConflict(t *testing.T) {
	s, _, _ := openTest(t)
	insertTopology(t, s)
	cases := map[string]func(*Node){
		"name":    func(n *Node) { n.Name = "vps-1" },
		"mesh ip": func(n *Node) { n.MeshIP = "10.77.0.1" },
		"wg key":  func(n *Node) { n.WGPubKey = testfix.WGKey(1) },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			n := node(testfix.Phone)
			mutate(&n)
			err := s.Tx(context.Background(), func(t2 *Tx) error { return t2.InsertNode(&n) })
			if !errors.Is(err, ErrConflict) {
				t.Fatalf("duplicate %s: err = %v, want ErrConflict", name, err)
			}
		})
	}
}

func TestNotFound(t *testing.T) {
	s, _, _ := openTest(t)
	err := s.Tx(context.Background(), func(t2 *Tx) error {
		_, err := t2.NodeByName("ghost")
		return err
	})
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("NodeByName(ghost) = %v, want ErrNotFound", err)
	}
	err = s.Tx(context.Background(), func(t2 *Tx) error { return t2.DeleteNode(999) })
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("DeleteNode(999) = %v, want ErrNotFound", err)
	}
}

func TestDeleteNodeCascades(t *testing.T) {
	s, _, _ := openTest(t)
	nodes := insertTopology(t, s)
	vps1, vps2, nas := nodes["vps-1"], nodes["vps-2"], nodes["nas"]
	tx(t, s, func(t2 *Tx) error {
		if _, _, err := t2.IssueToken(vps1.ID, TokenTTL); err != nil {
			return err
		}
		if err := t2.InsertDNSRecord(&DNSRecord{Name: "swarm", NodeID: vps1.ID}); err != nil {
			return err
		}
		if err := t2.SetNodeGroups(vps1.ID, []string{"swarm"}); err != nil {
			return err
		}
		if _, err := t2.exec(`INSERT INTO peer_status (node_id, peer_id, updated_at) VALUES (?, ?, 0), (?, ?, 0)`,
			nas.ID, vps1.ID, vps1.ID, vps2.ID); err != nil {
			return err
		}
		if _, err := t2.SetContainerNames(vps1.ID, []ContainerName{{Name: "web", IP: "10.77.128.129", Kind: "service"}}); err != nil {
			return err
		}
		return t2.DeleteNode(vps1.ID)
	})
	tx(t, s, func(t2 *Tx) error {
		got, err := t2.NodeByID(nas.ID)
		if err != nil {
			return err
		}
		if got.PrimaryHubID != 0 {
			t.Errorf("spoke still points at a deleted hub: %d", got.PrimaryHubID)
		}
		var n int
		for _, q := range []string{
			`SELECT COUNT(*) FROM join_tokens`,
			`SELECT COUNT(*) FROM dns_records`,
			`SELECT COUNT(*) FROM node_groups`,
			`SELECT COUNT(*) FROM peer_status`,
			`SELECT COUNT(*) FROM container_names`,
		} {
			if err := t2.queryRow(q).Scan(&n); err != nil {
				return err
			}
			if n != 0 {
				t.Errorf("%s = %d after deleting the node, want 0", q, n)
			}
		}
		return nil
	})
}

func TestTokenLifecycle(t *testing.T) {
	s, clock, _ := openTest(t)
	nodes := insertTopology(t, s)
	id := nodes["nas"].ID

	var tok string
	tx(t, s, func(t2 *Tx) error {
		var err error
		tok, _, err = t2.IssueToken(id, TokenTTL)
		return err
	})
	if len(tok) != 43 {
		t.Fatalf("token %q is not 43 chars", tok)
	}
	consume := func(token string) (int64, error) {
		var got int64
		err := s.Tx(context.Background(), func(t2 *Tx) error {
			var err error
			got, err = t2.ConsumeToken(token)
			return err
		})
		return got, err
	}
	if got, err := consume(tok); err != nil || got != id {
		t.Fatalf("first consume = %d, %v; want %d", got, err, id)
	}
	if _, err := consume(tok); !errors.Is(err, ErrTokenUsed) {
		t.Fatalf("second consume = %v, want ErrTokenUsed", err)
	}
	if _, err := consume(testfix.Token); !errors.Is(err, ErrTokenUnknown) {
		t.Fatalf("unknown token = %v, want ErrTokenUnknown", err)
	}

	var fresh, newer string
	tx(t, s, func(t2 *Tx) error {
		var err error
		if fresh, _, err = t2.IssueToken(id, TokenTTL); err != nil {
			return err
		}
		newer, _, err = t2.IssueToken(id, TokenTTL)
		return err
	})
	if _, err := consume(fresh); !errors.Is(err, ErrTokenUnknown) {
		t.Fatalf("a reissue must revoke the previous unused token, got %v", err)
	}
	clock.Advance(TokenTTL + time.Second)
	if _, err := consume(newer); !errors.Is(err, ErrTokenExpired) {
		t.Fatalf("expired token = %v, want ErrTokenExpired", err)
	}
	// The stored form is the hash, never the token.
	tx(t, s, func(t2 *Tx) error {
		var n int
		err := t2.queryRow(`SELECT COUNT(*) FROM join_tokens WHERE token_hash = ?`, newer).Scan(&n)
		if n != 0 {
			t.Error("a token was stored in clear")
		}
		return err
	})
}

func TestTokenConcurrentConsume(t *testing.T) {
	s, _, _ := openTest(t)
	nodes := insertTopology(t, s)
	var tok string
	tx(t, s, func(t2 *Tx) error {
		var err error
		tok, _, err = t2.IssueToken(nodes["laptop"].ID, TokenTTL)
		return err
	})
	const racers = 8
	var (
		wg   sync.WaitGroup
		mu   sync.Mutex
		wins int
	)
	for i := 0; i < racers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := s.Tx(context.Background(), func(t2 *Tx) error {
				_, err := t2.ConsumeToken(tok)
				return err
			})
			if err == nil {
				mu.Lock()
				wins++
				mu.Unlock()
			} else if !errors.Is(err, ErrTokenUsed) {
				t.Errorf("losing racer got %v, want ErrTokenUsed", err)
			}
		}()
	}
	wg.Wait()
	if wins != 1 {
		t.Fatalf("%d joins succeeded with one token, want exactly 1", wins)
	}
}

func TestVersions(t *testing.T) {
	s, _, _ := openTest(t)
	tx(t, s, func(t2 *Tx) error {
		for i, summary := range []string{"add vps-1", "add vps-2", "disable laptop"} {
			n, err := t2.InsertVersion(Version{Actor: "cli", Summary: summary, Snapshot: []byte(`{"n":` + string(rune('1'+i)) + `}`)})
			if err != nil {
				return err
			}
			if n != int64(i+1) {
				t.Errorf("version %d numbered %d", i+1, n)
			}
		}
		latest, err := t2.LatestVersion()
		if err != nil || latest != 3 {
			t.Errorf("LatestVersion = %d, %v", latest, err)
		}
		list, err := t2.Versions(2, 0)
		if err != nil {
			return err
		}
		if len(list) != 2 || list[0].Version != 3 || list[1].Version != 2 {
			t.Errorf("Versions(2) = %+v", list)
		}
		v, err := t2.VersionByNumber(2)
		if err != nil {
			return err
		}
		if string(v.Snapshot) != `{"n":2}` || v.Summary != "add vps-2" {
			t.Errorf("VersionByNumber(2) = %+v", v)
		}
		if _, err := t2.VersionByNumber(42); !errors.Is(err, ErrNotFound) {
			t.Errorf("VersionByNumber(42) = %v, want ErrNotFound", err)
		}
		return nil
	})
}

func TestAudit(t *testing.T) {
	s, _, _ := openTest(t)
	tx(t, s, func(t2 *Tx) error {
		if err := t2.Audit("cli", 0, "node.add", "nas", map[string]string{"role": "spoke"}); err != nil {
			return err
		}
		if err := t2.Audit("system", 0, "export.nightly", "", nil); err != nil {
			return err
		}
		log, err := t2.AuditLog(10, 0)
		if err != nil {
			return err
		}
		if len(log) != 2 || log[0].Action != "export.nightly" || log[1].Detail != `{"role":"spoke"}` {
			t.Errorf("AuditLog = %+v", log)
		}
		if log[1].At != testfix.Epoch.Unix() {
			t.Errorf("audit timestamp = %d, want the transaction clock", log[1].At)
		}
		return nil
	})
}

func TestSnapshotRestoreRoundTrip(t *testing.T) {
	s, _, _ := openTest(t)
	nodes := insertTopology(t, s)
	var before []byte
	tx(t, s, func(t2 *Tx) error {
		if err := t2.InsertDNSRecord(&DNSRecord{Name: "immich", NodeID: nodes["nas"].ID}); err != nil {
			return err
		}
		if err := t2.InsertDNSRecord(&DNSRecord{Name: "printer", IP: "10.77.1.50"}); err != nil {
			return err
		}
		if err := t2.SetSegmentation(testSegmentation()); err != nil {
			return err
		}
		box := Node{Name: "box", Role: "spoke", MeshIP: "10.77.2.9", Mode: proto.ModeContainer, ContainerSubnet: "10.77.129.0/24"}
		if err := t2.InsertNode(&box); err != nil {
			return err
		}
		snap, err := t2.Snapshot()
		if err != nil {
			return err
		}
		before, err = snap.Encode()
		return err
	})

	// Drift: remove two nodes (one with its record, one in container
	// mode), add one, edit one, change a setting and the mesh firewall.
	tx(t, s, func(t2 *Tx) error {
		if err := t2.DeleteNode(nodes["laptop"].ID); err != nil {
			return err
		}
		box, err := t2.NodeByName("box")
		if err != nil {
			return err
		}
		if err := t2.DeleteNode(box.ID); err != nil {
			return err
		}
		extra := node(testfix.Phone)
		if err := t2.InsertNode(&extra); err != nil {
			return err
		}
		nas, err := t2.NodeByName("nas")
		if err != nil {
			return err
		}
		nas.MTU = 1380
		if err := t2.UpdateNode(nas); err != nil {
			return err
		}
		if err := t2.SetSegmentation(Segmentation{Policies: []SegPolicy{DefaultPolicy()}}); err != nil {
			return err
		}
		return t2.SetSetting(SettingDefaultMTU, "1380")
	})

	tx(t, s, func(t2 *Tx) error {
		snap, err := DecodeSnapshot(before)
		if err != nil {
			return err
		}
		if err := t2.Restore(snap); err != nil {
			return err
		}
		after, err := t2.Snapshot()
		if err != nil {
			return err
		}
		b, err := after.Encode()
		if err != nil {
			return err
		}
		if string(b) != string(before) {
			t.Errorf("restore is not exact:\nbefore %s\nafter  %s", before, b)
		}
		laptop, err := t2.NodeByName("laptop")
		if err != nil {
			return err
		}
		if laptop.ID != nodes["laptop"].ID || laptop.ExitVia != nodes["vps-2"].ID {
			t.Errorf("restored laptop lost its id or exit: %+v", laptop)
		}
		return nil
	})
}

// A snapshot taken before a setting existed puts the setting's default
// back: rolling back past the DNS zone setting serves mesh again.
func TestRestoreOfAnOlderSnapshot(t *testing.T) {
	s, _, _ := openTest(t)
	tx(t, s, func(t2 *Tx) error {
		snap, err := t2.Snapshot()
		if err != nil {
			return err
		}
		delete(snap.Settings, SettingDNSZone)
		if err := t2.SetSetting(SettingDNSZone, "home.arpa"); err != nil {
			return err
		}
		if err := t2.Restore(snap); err != nil {
			return err
		}
		zone, err := t2.Setting(SettingDNSZone)
		if err == nil && zone != "mesh" {
			t.Errorf("dns_zone after restoring an older snapshot = %q, want mesh", zone)
		}
		return err
	})
	// The extra domains (amendment A39) came later still: none.
	tx(t, s, func(t2 *Tx) error {
		snap, err := t2.Snapshot()
		if err != nil {
			return err
		}
		delete(snap.Settings, SettingDNSExtraDomains)
		if err := t2.SetSetting(SettingDNSExtraDomains, "example.lan"); err != nil {
			return err
		}
		if err := t2.Restore(snap); err != nil {
			return err
		}
		if v, err := t2.Setting(SettingDNSExtraDomains); !errors.Is(err, ErrNotFound) {
			t.Errorf("dns_extra_domains after restoring an older snapshot = %q, %v; want none", v, err)
		}
		return nil
	})
}

func TestSplitDomains(t *testing.T) {
	if got := SplitDomains(" example.lan, ,home.arpa,"); !slices.Equal(got, []string{"example.lan", "home.arpa"}) {
		t.Errorf("SplitDomains = %v", got)
	}
	if got := SplitDomains(""); got != nil {
		t.Errorf("SplitDomains of nothing = %v", got)
	}
}

func TestDecodeSnapshotRefusesUnknownShapes(t *testing.T) {
	for name, in := range map[string]string{
		"unknown field": `{"format":"meshctl-snapshot/1","settings":{},"nodes":[],"dns_records":[],"firewall_rules":[],"users":[]}`,
		"other format":  `{"format":"meshctl-snapshot/9","settings":{},"nodes":[],"dns_records":[],"firewall_rules":[]}`,
		"not json":      `nope`,
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := DecodeSnapshot([]byte(in)); err == nil {
				t.Fatal("DecodeSnapshot accepted it")
			}
		})
	}
}

func TestRecordStatus(t *testing.T) {
	s, clock, _ := openTest(t)
	nodes := insertTopology(t, s)
	vps1, vps2, nas := nodes["vps-1"], nodes["vps-2"], nodes["nas"]
	update := StatusUpdate{
		AppliedVersion: 4, AppliedHash: strings.Repeat("a", 64), AgentVersion: "v1.0.0", WGBackend: "kernel",
		ObservedIP: "203.0.113.10", StatusJSON: `{"degraded":["firewall"]}`,
		Peers: []PeerStat{{PeerID: vps2.ID, LastHandshake: 100, RxBytes: 5, TxBytes: 6, Endpoint: "203.0.113.11:51820"}, {PeerID: nas.ID}},
	}
	tx(t, s, func(t2 *Tx) error { return t2.RecordStatus(vps1.ID, update) })
	clock.Advance(time.Minute)
	update.Peers = update.Peers[:1]
	tx(t, s, func(t2 *Tx) error { return t2.RecordStatus(vps1.ID, update) })

	tx(t, s, func(t2 *Tx) error {
		n, err := t2.NodeByID(vps1.ID)
		if err != nil {
			return err
		}
		if n.AppliedVersion != 4 || n.WGBackend != "kernel" || n.ObservedIP != "203.0.113.10" ||
			n.StatusJSON != update.StatusJSON || n.LastSeen != clock.Now().Unix() {
			t.Errorf("node after heartbeat = %+v", n)
		}
		peers, err := t2.PeerStats(vps1.ID)
		if err != nil {
			return err
		}
		if len(peers) != 1 || peers[0].PeerID != vps2.ID || peers[0].RxBytes != 5 || peers[0].UpdatedAt != clock.Now().Unix() {
			t.Errorf("peer rows are replaced, not merged: %+v", peers)
		}
		if err := t2.RecordStatus(999, update); !errors.Is(err, ErrNotFound) {
			t.Errorf("RecordStatus for a missing node = %v, want ErrNotFound", err)
		}
		return nil
	})
}

// A report replaces the node's names and says whether anything changed;
// the list is ordered by name, node and address.
func TestContainerNames(t *testing.T) {
	s, _, _ := openTest(t)
	nodes := insertTopology(t, s)
	nas, laptop := nodes["nas"].ID, nodes["laptop"].ID
	web := func(ip string) ContainerName { return ContainerName{Name: "web", IP: ip, Kind: "service"} }
	tx(t, s, func(t2 *Tx) error {
		for _, c := range []struct {
			node    int64
			names   []ContainerName
			changed bool
		}{
			{nas, []ContainerName{web("10.77.128.130"), {Name: "db", IP: "10.77.128.131", Kind: "container"}, web("10.77.128.129")}, true},
			{nas, []ContainerName{web("10.77.128.129"), {Name: "db", IP: "10.77.128.131", Kind: "container"}, web("10.77.128.130")}, false},
			{laptop, []ContainerName{web("10.77.129.130")}, true},
			{nas, []ContainerName{web("10.77.128.129"), {Name: "db", IP: "10.77.128.131", Kind: "alias"}, web("10.77.128.130")}, true},
		} {
			changed, err := t2.SetContainerNames(c.node, c.names)
			if err != nil {
				return err
			}
			if changed != c.changed {
				t.Errorf("SetContainerNames(%d, %v) changed = %v, want %v", c.node, c.names, changed, c.changed)
			}
		}
		got, err := t2.ContainerNames()
		if err != nil {
			return err
		}
		want := []ContainerName{{NodeID: nas, Name: "db", IP: "10.77.128.131", Kind: "alias"},
			{NodeID: nas, Name: "web", IP: "10.77.128.129", Kind: "service"}, {NodeID: nas, Name: "web", IP: "10.77.128.130", Kind: "service"},
			{NodeID: laptop, Name: "web", IP: "10.77.129.130", Kind: "service"}}
		if nas > laptop {
			t.Fatalf("the fixture inserts the NAS before the laptop (ids %d, %d)", nas, laptop)
		}
		if !slices.Equal(got, want) {
			t.Errorf("ContainerNames =\n%v\nwant\n%v", got, want)
		}
		if changed, err := t2.SetContainerNames(nas, nil); err != nil || !changed {
			t.Errorf("clearing the NAS's names = %v, %v", changed, err)
		}
		if got, err := t2.ContainerNames(); err != nil || len(got) != 1 || got[0].NodeID != laptop {
			t.Errorf("after clearing the NAS = %v, %v", got, err)
		}
		return nil
	})
}

// The relay failover monitor's state (amendment A40) is operational: it
// survives an update of the node's desired fields and stays out of the
// snapshot.
func TestHubState(t *testing.T) {
	s, _, _ := openTest(t)
	vps1 := insertTopology(t, s)["vps-1"]
	tx(t, s, func(t2 *Tx) error {
		if err := t2.SetHubState(vps1.ID, 100, 160); err != nil {
			return err
		}
		n, err := t2.NodeByID(vps1.ID)
		if err != nil {
			return err
		}
		n.MTU = 1380
		if err := t2.UpdateNode(n); err != nil {
			return err
		}
		if n, err = t2.NodeByID(vps1.ID); err != nil || n.DownSince != 100 || n.HealthySince != 160 {
			t.Errorf("after an update = %+v, %v", n, err)
		}
		snap, err := t2.Snapshot()
		if err != nil {
			return err
		}
		if b, _ := snap.Encode(); strings.Contains(string(b), "down_since") || strings.Contains(string(b), "healthy") {
			t.Errorf("the snapshot carries the failover state: %s", b)
		}
		if err := t2.SetHubState(999, 1, 0); !errors.Is(err, ErrNotFound) {
			t.Errorf("SetHubState for a missing node = %v", err)
		}
		return t2.SetHubState(vps1.ID, 0, 0)
	})
}

// A phone has no agent: the server records the profile it handed out, and
// what the servers report of their handshakes with it.
func TestPhoneState(t *testing.T) {
	s, _, _ := openTest(t)
	nodes := insertTopology(t, s)
	vps1, vps2, nas := nodes["vps-1"], nodes["vps-2"], nodes["nas"]
	beat := func(from Node, at int64) {
		tx(t, s, func(t2 *Tx) error {
			return t2.RecordStatus(from.ID, StatusUpdate{Peers: []PeerStat{{PeerID: nas.ID, LastHandshake: at}}})
		})
	}
	tx(t, s, func(t2 *Tx) error {
		if at, err := t2.LastHandshakeWith(nas.ID); err != nil || at != 0 {
			t.Errorf("before any report: %d, %v", at, err)
		}
		return nil
	})
	beat(vps1, 300)
	beat(vps2, 200)
	tx(t, s, func(t2 *Tx) error {
		if at, err := t2.LastHandshakeWith(nas.ID); err != nil || at != 300 {
			t.Errorf("LastHandshakeWith = %d, %v; want the latest report, 300", at, err)
		}
		if err := t2.SetApplied(nas.ID, strings.Repeat("b", 64), 7); err != nil {
			return err
		}
		n, err := t2.NodeByID(nas.ID)
		if err != nil {
			return err
		}
		if n.AppliedHash != strings.Repeat("b", 64) || n.AppliedVersion != 7 || n.LastSeen != 0 {
			t.Errorf("after SetApplied: %+v (it is not a heartbeat)", n)
		}
		if err := t2.SetApplied(999, "", 0); !errors.Is(err, ErrNotFound) {
			t.Errorf("SetApplied for a missing node = %v, want ErrNotFound", err)
		}
		return nil
	})
}

func TestOpenReadOnly(t *testing.T) {
	s, clock, path := openTest(t)
	insertTopology(t, s)
	ro, err := Open(context.Background(), path, Options{Now: clock.Now, ReadOnly: true})
	if err != nil {
		t.Fatalf("Open read-only: %v", err)
	}
	defer ro.Close()
	err = ro.Tx(context.Background(), func(t2 *Tx) error {
		nodes, err := t2.Nodes()
		if err != nil || len(nodes) != 5 {
			t.Errorf("read-only Nodes = %d, %v", len(nodes), err)
		}
		return t2.SetSetting(SettingMeshCIDR, "10.0.0.0/16")
	})
	if err == nil {
		t.Fatal("a read-only store accepted a write")
	}

	if _, err := Open(context.Background(), filepath.Join(t.TempDir(), "absent.db"), Options{ReadOnly: true}); err == nil ||
		!strings.Contains(err.Error(), "--data") {
		t.Fatalf("read-only Open of a missing file = %v, want a pointer at --data", err)
	}

	tx(t, s, func(t2 *Tx) error {
		_, err := t2.exec(`INSERT INTO schema_version (version, applied_at) VALUES (99, 0)`)
		return err
	})
	if _, err := Open(context.Background(), path, Options{ReadOnly: true}); err == nil || !strings.Contains(err.Error(), "own meshctl binary") {
		t.Fatalf("read-only Open of another schema = %v, want a refusal", err)
	}
}

func TestUsers(t *testing.T) {
	s, _, _ := openTest(t)
	tx(t, s, func(t2 *Tx) error {
		u := User{Username: "admin", PasswordHash: "$argon2id$x"}
		if err := t2.InsertUser(&u); err != nil {
			return err
		}
		if err := t2.InsertUser(&User{Username: "admin", PasswordHash: "y"}); !errors.Is(err, ErrConflict) {
			t.Errorf("duplicate username = %v, want ErrConflict", err)
		}
		if err := t2.SetPassword(u.ID, "$argon2id$z"); err != nil {
			return err
		}
		if err := t2.SetTOTP(u.ID, []byte("sealed"), true); err != nil {
			return err
		}
		if err := t2.UseTOTPStep(u.ID, 100); err != nil {
			return err
		}
		for _, step := range []int64{100, 99} {
			if err := t2.UseTOTPStep(u.ID, step); !errors.Is(err, ErrTOTPReplay) {
				t.Errorf("UseTOTPStep(%d) after 100 = %v, want ErrTOTPReplay", step, err)
			}
		}
		if err := t2.SetUserDisabled(u.ID, true); err != nil {
			return err
		}
		got, err := t2.UserByName("admin")
		if err != nil {
			return err
		}
		if got.PasswordHash != "$argon2id$z" || string(got.TOTPSecret) != "sealed" || !got.TOTPEnrolled || got.TOTPLastStep != 100 || !got.Disabled {
			t.Errorf("user %+v", got)
		}
		// A new secret starts the replay counter over.
		if err := t2.SetTOTP(u.ID, nil, false); err != nil {
			return err
		}
		if got, _ := t2.UserByID(u.ID); got.TOTPSecret != nil || got.TOTPLastStep != 0 {
			t.Errorf("after reset %+v", got)
		}
		if all, err := t2.Users(); err != nil || len(all) != 1 {
			t.Errorf("Users = %v, %v", all, err)
		}
		for _, err := range []error{t2.SetPassword(99, "x"), t2.SetTOTP(99, nil, false), t2.SetUserDisabled(99, true)} {
			if !errors.Is(err, ErrNotFound) {
				t.Errorf("write to a missing user = %v", err)
			}
		}
		if _, err := t2.UserByName("ghost"); !errors.Is(err, ErrNotFound) {
			t.Errorf("UserByName(ghost) = %v", err)
		}
		if _, err := t2.UserByID(99); !errors.Is(err, ErrNotFound) {
			t.Errorf("UserByID(99) = %v", err)
		}
		return nil
	})
}

func TestSessions(t *testing.T) {
	s, clock, _ := openTest(t)
	now := clock.Now().Unix()
	tx(t, s, func(t2 *Tx) error {
		u := User{Username: "admin", PasswordHash: "x"}
		if err := t2.InsertUser(&u); err != nil {
			return err
		}
		live := Session{IDHash: "aa", UserID: u.ID, CreatedAt: now, LastSeen: now, ExpiresAt: now + 60, CSRF: "c"}
		old := Session{IDHash: "bb", UserID: u.ID, CreatedAt: now - 100, LastSeen: now - 100, ExpiresAt: now - 1, CSRF: "d"}
		for _, x := range []Session{live, old} {
			if err := t2.InsertSession(x); err != nil {
				return err
			}
		}
		live.TOTPOK, live.ExpiresAt = true, now+600
		if err := t2.UpdateSession(live); err != nil {
			return err
		}
		if got, err := t2.SessionByHash("aa"); err != nil || got != live {
			t.Errorf("SessionByHash = %+v, %v", got, err)
		}
		if n, err := t2.PurgeSessions(now); err != nil || n != 1 {
			t.Errorf("PurgeSessions = %d, %v; want the expired one", n, err)
		}
		if err := t2.DeleteSession("aa"); err != nil {
			return err
		}
		if _, err := t2.SessionByHash("aa"); !errors.Is(err, ErrNotFound) {
			t.Errorf("deleted session = %v", err)
		}
		if err := t2.UpdateSession(live); !errors.Is(err, ErrNotFound) {
			t.Errorf("UpdateSession of a deleted session = %v", err)
		}
		if err := t2.InsertSession(live); err != nil {
			return err
		}
		if err := t2.DeleteUserSessions(u.ID); err != nil {
			return err
		}
		if _, err := t2.SessionByHash("aa"); !errors.Is(err, ErrNotFound) {
			t.Errorf("DeleteUserSessions left a session: %v", err)
		}
		return nil
	})
}
