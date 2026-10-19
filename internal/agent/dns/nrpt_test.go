package dns

import (
	"net/netip"
	"slices"
	"testing"
)

// The NRPT form of a registration: .mesh always, "." (every name) in exit
// mode, each under its fixed key; read back it is the same registration.
func TestNRPTRules(t *testing.T) {
	ip := netip.MustParseAddr("10.77.2.1")
	split := desired(Want{IP: ip})
	exit := desired(Want{IP: ip, Exit: true})
	if got := rulesFor(split); len(got) != 1 || got[0].Key != nrptMeshKey || got[0].Name != ".mesh" || !slices.Equal(got[0].Servers, []netip.Addr{ip}) {
		t.Fatalf("split rules = %+v", got)
	}
	got := rulesFor(exit)
	if len(got) != 2 || got[1].Key != nrptRootKey || got[1].Name != "." {
		t.Fatalf("exit rules = %+v", got)
	}
	for _, r := range []Registration{split, exit} {
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
