package api

import (
	"testing"
	"time"

	"github.com/hafio/vpn-mesh-utility/internal/proto"
	"github.com/hafio/vpn-mesh-utility/internal/testfix"
)

func TestDashboardHealth(t *testing.T) {
	now := testfix.Epoch
	seen := func(d time.Duration) int64 { return now.Add(-d).Unix() }
	for _, c := range []struct {
		n    NodeInfo
		want string
	}{
		{NodeInfo{Disabled: true}, proto.HealthOff},
		{NodeInfo{}, proto.HealthOff},
		{NodeInfo{Enrolled: true}, proto.HealthBad},
		{NodeInfo{Enrolled: true, LastSeen: seen(10 * time.Minute)}, proto.HealthBad},
		{NodeInfo{Enrolled: true, LastSeen: seen(time.Second), Degraded: []string{"firewall"}}, proto.HealthWarn},
		{NodeInfo{Enrolled: true, LastSeen: seen(time.Second), InSync: true, Rollback: &proto.Rollback{From: 5, To: 4}}, proto.HealthBad},
		{NodeInfo{Enrolled: true, LastSeen: seen(time.Second)}, proto.HealthWarn},
		{NodeInfo{Enrolled: true, LastSeen: seen(2 * time.Minute), InSync: true}, proto.HealthWarn},
		{NodeInfo{Enrolled: true, LastSeen: seen(time.Second), InSync: true}, proto.HealthOK},
		{NodeInfo{Role: "phone"}, proto.HealthOff},
		{NodeInfo{Role: "phone", Enrolled: true}, proto.HealthWarn},
		{NodeInfo{Role: "phone", Enrolled: true, InSync: true}, proto.HealthOff},
		{NodeInfo{Role: "phone", Enrolled: true, InSync: true, LastHandshake: seen(10 * time.Minute)}, proto.HealthOff},
		{NodeInfo{Role: "phone", Enrolled: true, InSync: true, LastHandshake: seen(time.Minute)}, proto.HealthOK},
		{NodeInfo{Role: "phone", Disabled: true, Enrolled: true, InSync: true, LastHandshake: seen(time.Minute)}, proto.HealthOff},
	} {
		if got, label := Health(c.n, now); got != c.want || label == "" {
			t.Errorf("Health(%+v) = %s %q, want %s", c.n, got, label, c.want)
		}
	}
	// A phone is heard of through its servers' handshakes, anything else by
	// its heartbeat.
	if (NodeInfo{Role: "phone", LastSeen: 1, LastHandshake: 2}).Seen() != 2 || (NodeInfo{Role: "spoke", LastSeen: 1, LastHandshake: 2}).Seen() != 1 {
		t.Fatal("Seen")
	}
}
