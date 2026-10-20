package dns

import (
	"net/netip"
	"slices"
	"strings"
	"testing"

	"github.com/hafio/vpn-mesh-utility/internal/proto"
)

// The NRPT form of a registration: .mesh always, each extra domain, "."
// (every name) in exit mode, each under a key of the agent's (the zone's
// and a domain's from its name); read back it is the same registration.
func TestNRPTRules(t *testing.T) {
	ip := netip.MustParseAddr("10.77.2.1")
	split := desired(Want{IP: ip})
	exit := desired(Want{IP: ip, Exit: true})
	extra := desired(Want{IP: ip, Config: &proto.DNS{Zone: "home.arpa", Domains: []string{"example.lan"}}})
	if got := rulesFor(split); len(got) != 1 || got[0].Key != nrptKey("mesh") || got[0].Name != ".mesh" || !slices.Equal(got[0].Servers, []netip.Addr{ip}) {
		t.Fatalf("split rules = %+v", got)
	}
	got := rulesFor(exit)
	if len(got) != 2 || got[1].Key != nrptRootKey || got[1].Name != "." {
		t.Fatalf("exit rules = %+v", got)
	}
	got = rulesFor(extra)
	if len(got) != 2 || got[0].Name != ".home.arpa" || got[1].Name != ".example.lan" || got[1].Key != nrptKey("example.lan") ||
		got[0].Key == got[1].Key {
		t.Fatalf("rules with an extra domain = %+v", got)
	}
	for _, r := range got {
		if !ownedKey(r.Key) || !ownedKey(strings.ToUpper(r.Key)) || len(r.Key) != len(nrptRootKey) {
			t.Errorf("key %s is not a GUID the agent owns", r.Key)
		}
	}
	// The key the zone had before A39 is the agent's too, so an upgrade
	// removes it; another program's rule is not.
	if !ownedKey(nrptRootKey) || !ownedKey("{6d657368-6374-6c00-8a2f-417e0c5b0001}") || ownedKey("{2b5c0e59-0000-4000-8000-000000000001}") {
		t.Error("ownedKey claims the wrong keys")
	}
	for _, r := range []Registration{split, exit, extra} {
		if back := fromRules(rulesFor(r)); !same(back, r) {
			t.Errorf("%+v reads back as %+v", r, back)
		}
	}
	if got := rulesFor(Registration{DNS: []netip.Addr{ip}, Domains: []Domain{{Name: "corp.example"}}}); len(got) != 0 {
		t.Fatalf("a search domain became rules %+v", got)
	}
	if r := fromRules(nil); len(r.DNS) != 0 || len(r.Domains) != 0 || r.DefaultRoute {
		t.Fatalf("no rules = %+v", r)
	}
}

func TestNRPTServers(t *testing.T) {
	ips := []netip.Addr{netip.MustParseAddr("10.77.2.1"), netip.MustParseAddr("10.77.0.1")}
	if s := serverList(ips); s != "10.77.2.1;10.77.0.1" || !slices.Equal(parseServers(s), ips) {
		t.Fatalf("serverList = %q", s)
	}
	if got := parseServers(" 10.77.2.1 ;junk;;"); !slices.Equal(got, ips[:1]) {
		t.Fatalf("parseServers kept %v", got)
	}
}
