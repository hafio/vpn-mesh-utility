package store

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/hafio/vpn-mesh-utility/internal/testfix"
)

// A node's earlier identity keys stay in its history (amendment A50):
// retiring one twice keeps its first time, an empty key is no key, and the
// history goes with the node.
func TestKeyHistory(t *testing.T) {
	s, _, _ := openTest(t)
	nodes := insertTopology(t, s)
	vps1 := nodes["vps-1"]
	old, older := testfix.IdentityKeyB64(90), testfix.IdentityKeyB64(91)
	tx(t, s, func(t2 *MeshTx) error {
		for _, k := range []struct {
			key string
			at  int64
		}{{older, 100}, {old, 200}, {older, 300}, {"", 400}} {
			if err := t2.RetireKey(vps1.ID, k.key, k.at); err != nil {
				return err
			}
		}
		got, err := t2.RetiredKeys(vps1.ID)
		if err != nil || !slices.Equal(got, []string{older, old}) {
			t.Errorf("RetiredKeys = %v, %v", got, err)
		}
		for key, want := range map[string]bool{old: true, older: true, vps1.IdentityPubKey: false} {
			if used, err := t2.KeyRetired(key); err != nil || used != want {
				t.Errorf("KeyRetired(%s) = %v, %v", key, used, err)
			}
		}
		if err := t2.DeleteNode(vps1.ID); err != nil {
			return err
		}
		if used, err := t2.KeyRetired(old); err != nil || used {
			t.Errorf("a removed node's history = %v, %v", used, err)
		}
		return nil
	})
}

// A failing query is an error from each accessor, naming what it did.
func TestKeyHistoryErrors(t *testing.T) {
	s, _, _ := openTest(t)
	errBack := errors.New("roll back")
	err := s.MeshTx(context.Background(), FirstMesh, func(t2 *MeshTx) error {
		if _, err := t2.exec(`DROP TABLE node_keys`); err != nil {
			return err
		}
		if err := t2.RetireKey(1, testfix.IdentityKeyB64(90), 1); err == nil || !strings.Contains(err.Error(), "retire") {
			t.Errorf("RetireKey = %v", err)
		}
		if _, err := t2.RetiredKeys(1); err == nil || !strings.Contains(err.Error(), "retired keys") {
			t.Errorf("RetiredKeys = %v", err)
		}
		if _, err := t2.KeyRetired(testfix.IdentityKeyB64(90)); err == nil || !strings.Contains(err.Error(), "identity key") {
			t.Errorf("KeyRetired = %v", err)
		}
		return errBack
	})
	if !errors.Is(err, errBack) {
		t.Fatal(err)
	}
}

// A rollback or an import leaves what an existing node brought at its join
// as it is now -- its keys, enrollment and platform -- so a key a token cut
// off never comes back; a node the snapshot re-creates gets its own back
// (amendment A50).
func TestRestoreKeepsWhatTheNodeBrought(t *testing.T) {
	s, _, _ := openTest(t)
	nodes := insertTopology(t, s)
	var snap Snapshot
	tx(t, s, func(t2 *MeshTx) error {
		var err error
		snap, err = t2.Snapshot()
		return err
	})
	tx(t, s, func(t2 *MeshTx) error {
		// After the snapshot: vps-1 joined again elsewhere, the laptop was
		// cut off by a token, the NAS was removed.
		vps1 := nodes["vps-1"]
		vps1.IdentityPubKey, vps1.WGPubKey, vps1.Hostname, vps1.MTU = testfix.IdentityKeyB64(80), testfix.WGKey(80), "new-host", 1380
		laptop := nodes["laptop"]
		laptop.Enrolled, laptop.IdentityPubKey, laptop.WGPubKey = false, "", ""
		for _, n := range []Node{vps1, laptop} {
			if err := t2.UpdateNode(n); err != nil {
				return err
			}
		}
		if err := t2.DeleteNode(nodes["nas"].ID); err != nil {
			return err
		}
		if err := t2.Restore(snap); err != nil {
			return err
		}
		vps1, err := t2.NodeByName("vps-1")
		if err != nil || vps1.IdentityPubKey != testfix.IdentityKeyB64(80) || vps1.WGPubKey != testfix.WGKey(80) ||
			vps1.Hostname != "new-host" || vps1.MTU != 1420 {
			t.Errorf("vps-1 after the restore = %+v, %v", vps1, err)
		}
		laptop, err = t2.NodeByName("laptop")
		if err != nil || laptop.Enrolled || laptop.IdentityPubKey != "" {
			t.Errorf("the cut-off laptop after the restore = %+v, %v", laptop, err)
		}
		nas, err := t2.NodeByName("nas")
		if err != nil || !nas.Enrolled || nas.IdentityPubKey != testfix.IdentityKeyB64(testfix.NAS.Key) {
			t.Errorf("the re-created nas = %+v, %v", nas, err)
		}
		return nil
	})
}
