package routes_test

import (
	"errors"
	"net/netip"
	"slices"
	"testing"

	"github.com/hafio/vpn-mesh-utility/internal/agent/agentfake"
	"github.com/hafio/vpn-mesh-utility/internal/agent/routes"
	"github.com/hafio/vpn-mesh-utility/internal/testfix"
)

func prefixes(ss ...string) []netip.Prefix {
	var out []netip.Prefix
	for _, s := range ss {
		out = append(out, netip.MustParsePrefix(s))
	}
	return out
}

func TestDesiredSkipsTheDefaultRouteAndDuplicates(t *testing.T) {
	cfg := testfix.ServerConfig()
	cfg.Peers[0].AllowedIPs = append(cfg.Peers[0].AllowedIPs, "0.0.0.0/0", "10.77.2.1/32")
	cfg.Peers[1].AllowedIPs = append(cfg.Peers[1].AllowedIPs, "10.77.2.1/32")
	got, err := routes.Desired(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if want := prefixes("10.77.0.2/32", "10.77.0.3/32", "10.77.2.1/32"); !slices.Equal(got, want) {
		t.Fatalf("Desired = %v, want %v", got, want)
	}
	cfg.Peers[0].AllowedIPs = []string{"bogus"}
	if _, err := routes.Desired(cfg); err == nil {
		t.Fatal("a malformed prefix was accepted")
	}
}

func TestSyncAddsThenRemovesAndIsIdempotent(t *testing.T) {
	tbl := agentfake.NewRoutes()
	if n, err := routes.Sync(tbl, "wg0", prefixes("10.77.0.2/32", "10.77.0.3/32")); err != nil || n != 2 {
		t.Fatalf("first Sync = %d, %v", n, err)
	}
	tbl.Calls = nil
	if n, err := routes.Sync(tbl, "wg0", prefixes("10.77.0.2/32", "10.77.0.3/32")); err != nil || n != 0 || tbl.Mutations() != 0 {
		t.Fatalf("second Sync = %d, %v, calls %v", n, err, tbl.Calls)
	}
	if n, err := routes.Sync(tbl, "wg0", prefixes("10.77.0.3/32", "10.77.1.1/32")); err != nil || n != 2 {
		t.Fatalf("change = %d, %v", n, err)
	}
	if want := []string{"add 10.77.1.1/32", "del 10.77.0.2/32"}; !slices.Equal(tbl.Calls, want) {
		t.Fatalf("calls %v, want add before del: %v", tbl.Calls, want)
	}
	tbl.Fail = errors.New("netlink down")
	if _, err := routes.Sync(tbl, "wg0", nil); err == nil {
		t.Fatal("a List failure must surface")
	}
}
