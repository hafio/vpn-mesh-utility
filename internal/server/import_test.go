package server

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/hafio/vpn-mesh-utility/internal/server/export"
	"github.com/hafio/vpn-mesh-utility/internal/server/store"
	"github.com/hafio/vpn-mesh-utility/internal/testfix"
)

func linuxOnly(t *testing.T) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("the data dir lock is Linux-only")
	}
}

// document is a three-server export as `meshctl export` wrote it before the
// mesh firewall: one custom rule, which imports as a policy.
func document() export.Document {
	d := export.Document{Format: export.Format, Settings: map[string]string{store.SettingMeshCIDR: testfix.MeshCIDR}}
	for _, f := range testfix.Servers {
		d.Nodes = append(d.Nodes, export.Node{ID: f.ID, Name: f.Name, Role: "server", MeshIP: f.MeshIP, Endpoint: f.Endpoint,
			ListenPort: 51820, Keepalive: 25, MTU: 1420, Enrolled: true, WGPublicKey: testfix.WGKey(f.Key),
			IdentityPublicKey: testfix.IdentityKeyB64(f.Key)})
	}
	d.Firewall = []export.FirewallRule{{Node: "vps-1", Proto: "tcp", Port: 9000, Source: "10.77.1.0/24", Note: "backup"}}
	return d
}

func TestImportRestoresAnExport(t *testing.T) {
	linuxOnly(t)
	data := t.TempDir()
	res, err := Import(context.Background(), ImportConfig{DataDir: data, Doc: document()})
	if err != nil || res.Version == 0 || len(res.Changed) != 3 {
		t.Fatalf("Import = %+v, %v", res, err)
	}
	st, err := store.Open(context.Background(), filepath.Join(data, DBFile), store.Options{})
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	var snap store.Snapshot
	var audit []store.AuditEntry
	if err := st.Tx(context.Background(), func(tx *store.Tx) error {
		var err error
		if snap, err = tx.Snapshot(); err != nil {
			return err
		}
		audit, err = tx.AuditLog(1, 0)
		return err
	}); err != nil {
		t.Fatal(err)
	}
	pols := snap.Segmentation.Policies
	if len(snap.Nodes) != 3 || snap.Nodes[1].ID != testfix.VPS2.ID || len(snap.Firewall) != 0 || len(pols) != 2 ||
		pols[0].Name != "firewall-1" || pols[0].To[0] != "node:vps-1" || pols[0].Services[0] != "tcp/9000" || pols[1].Name != "allow-all" {
		t.Fatalf("restored %+v, policies %+v", snap, pols)
	}
	if len(audit) != 1 || audit[0].Action != "mesh.import" || audit[0].Actor != "cli" {
		t.Fatalf("audit %+v", audit)
	}
}

func TestImportRefusesWhileTheServerRuns(t *testing.T) {
	linuxOnly(t)
	data := t.TempDir()
	unlock, err := Lock(data)
	if err != nil {
		t.Fatal(err)
	}
	defer unlock()
	if _, err := Import(context.Background(), ImportConfig{DataDir: data, Doc: document()}); !errors.Is(err, ErrLocked) || !strings.Contains(err.Error(), "stop the meshctl server") {
		t.Fatalf("Import beside a lock holder = %v", err)
	}
	if _, err := os.Stat(filepath.Join(data, DBFile)); err == nil {
		t.Fatal("a refused import created the database")
	}
}

func TestImportRefusesInvalidDocumentsUntouched(t *testing.T) {
	linuxOnly(t)
	for name, edit := range map[string]func(*export.Document){
		"hostile note":         func(d *export.Document) { d.Firewall[0].Note = "x\ny" },
		"two nodes on one ip":  func(d *export.Document) { d.Nodes[1].MeshIP = d.Nodes[0].MeshIP },
		"hub that is no node":  func(d *export.Document) { d.Nodes[2].Hub = "ghost" },
		"mtu the agent refuse": func(d *export.Document) { d.Nodes[0].MTU = 100 },
	} {
		t.Run(name, func(t *testing.T) {
			data := t.TempDir()
			d := document()
			edit(&d)
			if _, err := Import(context.Background(), ImportConfig{DataDir: data, Doc: d}); err == nil {
				t.Fatal("Import accepted it")
			}
			st, err := store.Open(context.Background(), filepath.Join(data, DBFile), store.Options{})
			if err != nil {
				t.Fatal(err)
			}
			defer st.Close()
			if err := st.Tx(context.Background(), func(tx *store.Tx) error {
				nodes, err := tx.Nodes()
				if err == nil && len(nodes) != 0 {
					t.Errorf("a refused import wrote %d nodes", len(nodes))
				}
				return err
			}); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestLockIsExclusive(t *testing.T) {
	linuxOnly(t)
	data := t.TempDir()
	unlock, err := Lock(data)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Lock(data); !errors.Is(err, ErrLocked) {
		t.Fatalf("a second Lock = %v", err)
	}
	unlock()
	again, err := Lock(data)
	if err != nil {
		t.Fatalf("Lock after unlock = %v", err)
	}
	again()
	if _, err := Lock(filepath.Join(data, "missing", "dir")); err == nil {
		t.Fatal("Lock in a missing dir succeeded")
	}
}
