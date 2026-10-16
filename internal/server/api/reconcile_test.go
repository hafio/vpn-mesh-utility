package api

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/hafio/vpn-mesh-utility/internal/proto"
	"github.com/hafio/vpn-mesh-utility/internal/server/store"
	"github.com/hafio/vpn-mesh-utility/internal/testfix"
)

// A version is written only when some node's computed config changes
// (amendment A16), and only the changed nodes are told.
func TestReconcilerVersionsOnlyRealChanges(t *testing.T) {
	h := newHarness(t)
	h.seedServers(t)
	ctx := context.Background()
	v := h.latestVersion(t)
	if res, err := h.rec.Apply(ctx, Change{Actor: "test"}, nil); err != nil || res.Version != 0 || h.latestVersion(t) != v {
		t.Fatalf("no-op reconcile: %+v, %v", res, err)
	}

	subs := map[int64]<-chan proto.ConfigEvent{}
	for _, f := range testfix.Servers {
		ev, _, cancel := h.hub.Subscribe(f.ID)
		defer cancel()
		subs[f.ID] = ev
	}
	// vps-3's endpoint is in the other servers' configs, not in its own.
	res, err := h.rec.Apply(ctx, Change{Actor: "test", Action: "node.set", Target: "vps-3"}, func(tx *store.Tx, _ *Change) error {
		n, err := tx.NodeByName("vps-3")
		if err != nil {
			return err
		}
		n.Endpoint = "203.0.113.99:51820"
		return tx.UpdateNode(n)
	})
	if err != nil || res.Version != v+1 || len(res.Changed) != 2 {
		t.Fatalf("Apply = %+v, %v", res, err)
	}
	for _, f := range testfix.Servers {
		select {
		case ev := <-subs[f.ID]:
			if f.Name == "vps-3" || ev.Version != res.Version {
				t.Errorf("%s got %+v", f.Name, ev)
			}
		default:
			if f.Name != "vps-3" {
				t.Errorf("%s was not told", f.Name)
			}
		}
	}
	var ver store.Version
	if err := h.st.Tx(ctx, func(tx *store.Tx) error {
		var err error
		ver, err = tx.VersionByNumber(res.Version)
		return err
	}); err != nil {
		t.Fatal(err)
	}
	snap, err := store.DecodeSnapshot(ver.Snapshot)
	if err != nil || ver.Summary != "node.set vps-3" || len(snap.Nodes) != 3 {
		t.Fatalf("version %+v: %v", ver, err)
	}
}

func TestReconcilerRollsBackAFailedMutation(t *testing.T) {
	h := newHarness(t)
	h.seedServers(t)
	boom := errors.New("boom")
	_, err := h.rec.Apply(context.Background(), Change{Actor: "test", Action: "node.set"}, func(tx *store.Tx, _ *Change) error {
		n, err := tx.NodeByName("vps-1")
		if err != nil {
			return err
		}
		n.MTU = 1300
		if err := tx.UpdateNode(n); err != nil {
			return err
		}
		return boom
	})
	if !errors.Is(err, boom) || h.node(t, "vps-1").MTU != 1420 {
		t.Fatalf("Apply = %v; the mutation must be rolled back", err)
	}
}

func TestPlanErrorListsTheProblems(t *testing.T) {
	h := newHarness(t)
	h.seedServers(t)
	_, err := h.rec.Apply(context.Background(), Change{Actor: "test"}, func(tx *store.Tx, _ *Change) error {
		n, err := tx.NodeByName("vps-2")
		if err != nil {
			return err
		}
		n.MTU = 100
		return tx.UpdateNode(n)
	})
	var pe *PlanError
	if !errors.As(err, &pe) || len(pe.Problems) == 0 || pe.Problems[0].Node != "vps-2" {
		t.Fatalf("Apply = %v", err)
	}
	if ae := toAPIError(err); ae.status != 409 || ae.code != proto.CodeConflict {
		t.Fatalf("toAPIError = %+v", ae)
	}
}

func TestLoadInputRefusesBrokenSettings(t *testing.T) {
	h := newHarness(t)
	for key, value := range map[string]string{store.SettingMeshCIDR: "not-a-cidr", store.SettingDefaultHub: "vps-1"} {
		err := h.st.Tx(context.Background(), func(tx *store.Tx) error {
			if err := tx.SetSetting(key, value); err != nil {
				return err
			}
			_, _, err := LoadInput(tx)
			if err == nil {
				t.Errorf("%s = %q was accepted", key, value)
			}
			return errors.New("roll back")
		})
		if err == nil {
			t.Fatal("expected the rollback error")
		}
	}
}

func TestRollbackIsANewVersion(t *testing.T) {
	h := newHarness(t)
	h.seedServers(t)
	ctx := context.Background()
	v1 := h.latestVersion(t)
	move := func(tx *store.Tx, _ *Change) error {
		n, err := tx.NodeByName("vps-3")
		if err != nil {
			return err
		}
		n.Endpoint = "203.0.113.99:51820"
		return tx.UpdateNode(n)
	}
	if _, err := h.rec.Apply(ctx, Change{Actor: "test", Action: "node.set"}, move); err != nil {
		t.Fatal(err)
	}
	events, _, cancel := h.hub.Subscribe(testfix.VPS1.ID)
	defer cancel()

	res, err := h.rec.Rollback(ctx, Change{Actor: "user:admin", UserID: 1}, v1)
	if err != nil || res.Version != v1+2 {
		t.Fatalf("Rollback = %+v, %v; want version %d", res, err, v1+2)
	}
	if h.node(t, "vps-3").Endpoint != testfix.VPS3.Endpoint {
		t.Fatal("the endpoint was not restored")
	}
	select {
	case ev := <-events:
		if ev.Version != res.Version {
			t.Fatalf("event %+v", ev)
		}
	default:
		t.Fatal("vps-1 was not told about the rollback")
	}
	var ver store.Version
	if err := h.st.Tx(ctx, func(tx *store.Tx) error {
		var err error
		ver, err = tx.VersionByNumber(res.Version)
		return err
	}); err != nil {
		t.Fatal(err)
	}
	if ver.RolledBackFrom != v1 || ver.Summary != fmt.Sprintf("rollback to v%d", v1) || ver.Actor != "user:admin" {
		t.Fatalf("version row %+v", ver)
	}

	// Rolling back to a state equal to the current one writes nothing.
	if res, err := h.rec.Rollback(ctx, Change{Actor: "test"}, v1); err != nil || res.Version != 0 {
		t.Fatalf("no-op rollback = %+v, %v", res, err)
	}
	if _, err := h.rec.Rollback(ctx, Change{Actor: "test"}, 999); toAPIError(err).status != 404 {
		t.Fatalf("rollback to a missing version = %v", err)
	}
}
