package versions

import (
	"reflect"
	"testing"

	"github.com/hafio/vpn-mesh-utility/internal/server/store"
	"github.com/hafio/vpn-mesh-utility/internal/testfix"
)

func snap(nodes ...store.SnapNode) store.Snapshot {
	return store.Snapshot{Format: store.SnapshotFormat, Settings: map[string]string{store.SettingDefaultMTU: "1420"}, Nodes: nodes}
}

func server(f testfix.Node) store.SnapNode {
	return store.SnapNode{ID: f.ID, Name: f.Name, Role: "server", MeshIP: f.MeshIP, Endpoint: f.Endpoint, ListenPort: 51820,
		Keepalive: 25, MTU: 1420, Enrolled: true, WGPubKey: testfix.WGKey(f.Key)}
}

func TestCompareEqualSnapshotsIsEmpty(t *testing.T) {
	a := snap(server(testfix.VPS1), server(testfix.VPS2))
	if d := Compare(a, a); !d.Empty() || d.ChangesMembership() {
		t.Fatalf("Compare(a, a) = %+v", d)
	}
}

// A snapshot from before the mesh firewall compares by what its custom
// rules became: unchanged rules are no change.
func TestCompareSeesThroughTheMigration(t *testing.T) {
	old := snap(server(testfix.VPS1))
	old.Segmentation = nil
	old.Firewall = []store.FirewallRule{{ID: 3, NodeID: testfix.VPS1.ID, Proto: "tcp", Port: 22, SourceCIDR: "10.77.2.0/24"}}
	now := snap(server(testfix.VPS1))
	migrated := old.EffectiveSegmentation()
	now.Segmentation = &migrated
	if d := Compare(old, now); !d.Empty() {
		t.Fatalf("the migration reads as a change: %+v", d)
	}
	now.Segmentation.Policies[0].Disabled = true
	if d := Compare(old, now); len(d.Firewall) != 2 || d.Firewall[0] != "+ policy 1 firewall-3: allow 10.77.2.0/24 -> node:vps-1 tcp/22 (disabled)" {
		t.Fatalf("a disabled policy = %q", d.Firewall)
	}
}

func TestCompareListsEveryKindOfChange(t *testing.T) {
	from := snap(server(testfix.VPS1), server(testfix.VPS2), server(testfix.VPS3))
	from.DNS = []store.DNSRecord{{Name: "files", NodeID: testfix.VPS2.ID}, {Name: "old", IP: "10.77.0.9"}}
	from.Segmentation = &store.Segmentation{Policies: []store.SegPolicy{{Name: "backup", From: []string{"10.77.1.0/24"},
		To: []string{"role:server"}, Services: []string{"tcp/9000"}, Action: store.SegAllow}, store.DefaultPolicy()}}

	moved := server(testfix.VPS2)
	moved.Endpoint, moved.MTU, moved.WGPubKey = "198.51.100.4:51820", 1380, testfix.WGKey(9)
	nas := store.SnapNode{ID: testfix.NAS.ID, Name: "nas", Role: "spoke", MeshIP: testfix.NAS.MeshIP, PrimaryHubID: testfix.VPS1.ID, Keepalive: 25, MTU: 1420}
	to := snap(server(testfix.VPS1), moved, nas)
	to.Settings[store.SettingDefaultMTU] = "1400"
	to.DNS = []store.DNSRecord{{Name: "files", NodeID: testfix.VPS2.ID}}
	to.Segmentation = &store.Segmentation{Groups: []store.SegGroup{{Name: "storage", Nodes: []string{"nas"}}},
		Policies: []store.SegPolicy{{Name: "backup", From: []string{"group:storage"}, To: []string{"node:vps-1"}, Services: []string{"tcp/9000"},
			Action: store.SegAllow, Schedule: &store.SegSchedule{Days: []string{"sat"}}}, store.DefaultPolicy()}, TimeZone: "UTC"}

	d := Compare(from, to)
	want := Diff{
		AddedNodes:   []string{"nas"},
		RemovedNodes: []string{"vps-3"},
		ChangedNodes: []NodeDiff{{Name: "vps-2", Fields: []FieldChange{
			{Field: "endpoint", From: "203.0.113.11:51820", To: "198.51.100.4:51820"},
			{Field: "mtu", From: "1420", To: "1380"},
			{Field: "wg_public_key", From: testfix.WGKey(2)[:8] + "...", To: testfix.WGKey(9)[:8] + "..."},
		}}},
		Settings: []FieldChange{{Field: store.SettingDefaultMTU, From: "1420", To: "1400"}},
		DNS:      []string{"- old -> 10.77.0.9"},
		Firewall: []string{"+ group storage: nas", "+ policy 1 backup: allow group:storage -> node:vps-1 tcp/9000 sat", "+ schedules in UTC",
			"- policy 1 backup: allow 10.77.1.0/24 -> role:server tcp/9000"},
	}
	if !reflect.DeepEqual(d, want) {
		t.Fatalf("Compare =\n%+v\nwant\n%+v", d, want)
	}
	if !d.ChangesMembership() || d.Empty() {
		t.Fatal("membership change not reported")
	}
}

func TestARenameIsAChangeNotAReplacement(t *testing.T) {
	a := server(testfix.VPS1)
	b := a
	b.Name = "vps-one"
	d := Compare(snap(a), snap(b))
	if d.ChangesMembership() || len(d.ChangedNodes) != 1 || d.ChangedNodes[0].Fields[0] != (FieldChange{Field: "name", From: "vps-1", To: "vps-one"}) {
		t.Fatalf("Compare = %+v", d)
	}
}
