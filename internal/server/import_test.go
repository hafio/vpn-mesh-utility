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
	if err := st.MeshTx(context.Background(), store.FirstMesh, func(tx *store.MeshTx) error {
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
	if _, err := Import(context.Background(), ImportConfig{DataDir: data, Doc: document()}); !errors.Is(err, ErrLocked) || !strings.Contains(err.Error(), "stop the meshctl control plane") {
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
			if err := st.MeshTx(context.Background(), store.FirstMesh, func(tx *store.MeshTx) error {
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

// import --mesh (amendment A49): a mesh the control plane lacks is made for
// the import with the document's network and zone; an import that fails
// takes it away again, and one whose network another mesh has is refused
// while the mesh is being made.
func TestImportIntoANamedMesh(t *testing.T) {
	linuxOnly(t)
	data, ctx := t.TempDir(), context.Background()
	if _, err := Import(ctx, ImportConfig{DataDir: data, Doc: document()}); err != nil {
		t.Fatal(err)
	}
	lab := document()
	lab.Settings = map[string]string{store.SettingMeshCIDR: "10.78.0.0/16", store.SettingDNSZone: "lab"}
	lab.Firewall = nil
	for i := range lab.Nodes {
		lab.Nodes[i].MeshIP = strings.Replace(lab.Nodes[i].MeshIP, "10.77.", "10.78.", 1)
	}
	// The first mesh's node ids and keys: the lab is made, refused, gone.
	if _, err := Import(ctx, ImportConfig{DataDir: data, Mesh: "lab", Doc: lab}); err == nil {
		t.Fatal("an import onto the first mesh's node ids went through")
	}
	twin := lab
	twin.Settings = map[string]string{store.SettingMeshCIDR: testfix.MeshCIDR, store.SettingDNSZone: "twin"}
	if _, err := Import(ctx, ImportConfig{DataDir: data, Mesh: "twin", Doc: twin}); err == nil ||
		!strings.Contains(err.Error(), "create mesh twin for the import") || !strings.Contains(err.Error(), "overlaps mesh default's") {
		t.Fatalf("an import onto the first mesh's network = %v", err)
	}
	for i := range lab.Nodes {
		key := byte(20 + i)
		lab.Nodes[i].ID += 10
		lab.Nodes[i].WGPublicKey, lab.Nodes[i].IdentityPublicKey = testfix.WGKey(key), testfix.IdentityKeyB64(key)
	}
	res, err := Import(ctx, ImportConfig{DataDir: data, Mesh: "lab", Doc: lab})
	if err != nil || res.Version != 1 || len(res.Changed) != 3 {
		t.Fatalf("Import into the lab = %+v, %v", res, err)
	}
	st, err := store.Open(ctx, filepath.Join(data, DBFile), store.Options{})
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	if err := st.Tx(ctx, func(tx *store.Tx) error {
		meshes, err := tx.Meshes()
		if err != nil {
			return err
		}
		if len(meshes) != 2 || meshes[0].Name != "default" || meshes[1].Name != "lab" {
			t.Errorf("meshes = %+v", meshes)
			return nil
		}
		nodes, err := tx.In(meshes[1].ID).Nodes()
		if err != nil {
			return err
		}
		if len(nodes) != 3 || nodes[0].ID != testfix.VPS1.ID+10 || nodes[0].MeshIP != "10.78.0.1" {
			t.Errorf("the lab's nodes = %+v", nodes)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}
