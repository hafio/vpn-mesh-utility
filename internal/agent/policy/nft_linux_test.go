//go:build linux

package policy

import (
	"net/netip"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/nftables"
	"github.com/google/nftables/expr"
	"github.com/google/nftables/userdata"

	"github.com/hafio/vpn-mesh-utility/internal/testfix"
	"github.com/hafio/vpn-mesh-utility/internal/validate"
)

// The batch: the table, the six chains (the base ones accepting by
// default), every rule in order with its anonymous sets before it, and the
// fingerprint on the first rule only.
func TestBuildBatch(t *testing.T) {
	rs, err := Compile(*testfix.Policy(), Local{Self: netip.MustParseAddr(testfix.VPS1.MeshIP)}, time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	b := build(rs, "0123456789abcdef")
	if b.table.Name != TableName || b.table.Family != nftables.TableFamilyINet || len(b.chains) != len(rs.Chains) {
		t.Fatalf("table %+v, %d chains", b.table, len(b.chains))
	}
	for i, ch := range b.chains {
		hooked := rs.Chains[i].Hook != ""
		if ch.Name != rs.Chains[i].Name || (ch.Hooknum != nil) != hooked || (hooked && (*ch.Policy != nftables.ChainPolicyAccept ||
			ch.Type != nftables.ChainTypeFilter || *ch.Priority != *nftables.ChainPriorityFilter)) {
			t.Errorf("chain %d = %+v", i, ch)
		}
	}
	sets := map[uint32]*nftables.Set{}
	rules := 0
	for _, op := range b.ops {
		if op.set != nil {
			if !op.set.Anonymous || !op.set.Constant || !op.set.Interval || sets[op.set.ID] != nil {
				t.Fatalf("set %+v", op.set)
			}
			sets[op.set.ID] = op.set
			continue
		}
		for _, x := range op.rule.Exprs {
			if l, ok := x.(*expr.Lookup); ok && sets[l.SetID] == nil {
				t.Fatalf("a rule looks up set %d before it is added", l.SetID)
			}
		}
		comment, ok := userdata.GetString(op.rule.UserData, userdata.TypeComment)
		if (rules == 0) != ok || (ok && comment != markPrefix+"0123456789abcdef") {
			t.Fatalf("rule %d's comment = %q, %v", rules, comment, ok)
		}
		rules++
	}
	if rules != rs.Count() {
		t.Fatalf("%d rules in the batch, %d in the ruleset", rules, rs.Count())
	}
}

func TestIntervalElems(t *testing.T) {
	type el struct {
		key string
		end bool
	}
	read := func(es []nftables.SetElement) []el {
		var out []el
		for _, e := range es {
			var parts []string
			for _, b := range e.Key {
				parts = append(parts, strconv.Itoa(int(b)))
			}
			out = append(out, el{strings.Join(parts, "."), e.IntervalEnd})
		}
		return out
	}
	for name, c := range map[string]struct {
		rs   []Range
		size int
		want []el
	}{
		"addresses": {[]Range{{0x0a4d0001, 0x0a4d0003}, {0x0a4d0105, 0x0a4d0105}}, 4,
			[]el{{"0.0.0.0", true}, {"10.77.0.1", false}, {"10.77.0.4", true}, {"10.77.1.5", false}, {"10.77.1.6", true}}},
		"from zero":  {[]Range{{0, 9}}, 4, []el{{"0.0.0.0", false}, {"0.0.0.10", true}}},
		"to the top": {[]Range{{0xffffff00, 0xffffffff}}, 4, []el{{"0.0.0.0", true}, {"255.255.255.0", false}}},
		"ports":      {[]Range{{22, 22}, {1000, 65535}}, 2, []el{{"0.0", true}, {"0.22", false}, {"0.23", true}, {"3.232", false}}},
	} {
		if got := read(intervalElems(c.rs, c.size)); !reflect.DeepEqual(got, c.want) {
			t.Errorf("%s: %v, want %v", name, got, c.want)
		}
	}
}

// Address sets of both families (M8): 4- or 16-byte keys, an end past each
// range but the family's last address.
func TestIntervalAddrElems(t *testing.T) {
	r := func(first, last string) validate.AddrRange {
		return validate.AddrRange{First: netip.MustParseAddr(first), Last: netip.MustParseAddr(last)}
	}
	text := func(es []nftables.SetElement) string {
		var out []string
		for _, e := range es {
			a, _ := netip.AddrFromSlice(e.Key)
			out = append(out, a.String()+map[bool]string{true: "]"}[e.IntervalEnd])
		}
		return strings.Join(out, " ")
	}
	for name, c := range map[string]struct {
		rs   []validate.AddrRange
		want string
	}{
		"IPv4":            {[]validate.AddrRange{r("10.77.0.1", "10.77.0.3"), r("10.77.1.5", "10.77.1.5")}, "0.0.0.0] 10.77.0.1 10.77.0.4] 10.77.1.5 10.77.1.6]"},
		"IPv6":            {[]validate.AddrRange{r("fd10:77::1", "fd10:77::3")}, "::] fd10:77::1 fd10:77::4]"},
		"IPv6 to the top": {[]validate.AddrRange{r("ffff::", "ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff")}, "::] ffff::"},
		"from zero":       {[]validate.AddrRange{r("::", "::9")}, ":: ::a]"},
	} {
		if got := text(intervalAddrElems(c.rs)); got != c.want {
			t.Errorf("%s: %s, want %s", name, got, c.want)
		}
		for _, e := range intervalAddrElems(c.rs) {
			if len(e.Key) != len(c.rs[0].First.AsSlice()) {
				t.Errorf("%s: a %d-byte key", name, len(e.Key))
			}
		}
	}
	if intervalAddrElems(nil) != nil {
		t.Error("no ranges, no elements")
	}
}

// A rule's expressions, field by field.
func TestExprs(t *testing.T) {
	var added []*nftables.Set
	set := func(key nftables.SetDatatype, elems []nftables.SetElement) *nftables.Set {
		s := &nftables.Set{ID: uint32(len(added) + 1), Name: "__set%d", KeyType: key}
		added = append(added, s)
		return s
	}
	one := func(s string) []validate.AddrRange {
		a := netip.MustParseAddr(s)
		return []validate.AddrRange{{First: a, Last: a}}
	}
	got := exprs(Rule{IIF: "wg0", State: StateEstablishedRelated, Family: IPv4, Src: one("10.77.0.1"), Proto: "tcp", Ports: []Range{{22, 22}},
		Verdict: Jump, Target: "x"}, set)
	want := []expr.Any{
		&expr.Meta{Key: expr.MetaKeyIIFNAME, Register: 1}, &expr.Cmp{Op: expr.CmpOpEq, Register: 1, Data: ifname("wg0")},
		&expr.Ct{Register: 1, Key: expr.CtKeySTATE}, got[3], &expr.Cmp{Op: expr.CmpOpNeq, Register: 1, Data: []byte{0, 0, 0, 0}},
		&expr.Meta{Key: expr.MetaKeyNFPROTO, Register: 1}, &expr.Cmp{Op: expr.CmpOpEq, Register: 1, Data: []byte{2}},
		&expr.Payload{DestRegister: 1, Base: expr.PayloadBaseNetworkHeader, Offset: 12, Len: 4},
		&expr.Lookup{SourceRegister: 1, SetName: "__set%d", SetID: 1},
		&expr.Meta{Key: expr.MetaKeyL4PROTO, Register: 1}, &expr.Cmp{Op: expr.CmpOpEq, Register: 1, Data: []byte{6}},
		&expr.Payload{DestRegister: 1, Base: expr.PayloadBaseTransportHeader, Offset: 2, Len: 2},
		&expr.Lookup{SourceRegister: 1, SetName: "__set%d", SetID: 2},
		&expr.Verdict{Kind: expr.VerdictJump, Chain: "x"},
	}
	if !reflect.DeepEqual(got, want) || len(added) != 2 || added[0].KeyType.Name != "ipv4_addr" || added[1].KeyType.Name != "inet_service" {
		t.Fatalf("exprs =\n%#v\nwant\n%#v", got, want)
	}
	if bw, ok := got[3].(*expr.Bitwise); !ok || bw.Len != 4 {
		t.Fatalf("the state mask = %#v", got[3])
	}
	for _, c := range []struct {
		r    Rule
		last expr.Any
	}{
		{Rule{OIF: "lo", Verdict: Accept}, &expr.Verdict{Kind: expr.VerdictAccept}},
		{Rule{State: StateInvalid, Verdict: Drop}, &expr.Verdict{Kind: expr.VerdictDrop}},
		{Rule{Proto: "icmp", Echo: true, Verdict: Accept}, &expr.Verdict{Kind: expr.VerdictAccept}},
	} {
		x := exprs(c.r, set)
		if !reflect.DeepEqual(x[len(x)-1], c.last) {
			t.Errorf("%+v ends %#v", c.r, x[len(x)-1])
		}
	}
	if echo := exprs(Rule{Proto: "icmp", Echo: true, Verdict: Accept}, set); !reflect.DeepEqual(echo[3], &expr.Cmp{Op: expr.CmpOpEq,
		Register: 1, Data: []byte{icmpEchoRequest}}) {
		t.Errorf("an echo request = %#v", echo)
	}
	// IPv6 (M8): its family, the IPv6 header's destination, ICMPv6's ping.
	added = nil
	six := exprs(Rule{Family: IPv6, Dst: one("fd10:77::1"), Proto: ICMPv6, Echo: true, Verdict: Accept}, set)
	want = []expr.Any{
		&expr.Meta{Key: expr.MetaKeyNFPROTO, Register: 1}, &expr.Cmp{Op: expr.CmpOpEq, Register: 1, Data: []byte{10}},
		&expr.Payload{DestRegister: 1, Base: expr.PayloadBaseNetworkHeader, Offset: 24, Len: 16},
		&expr.Lookup{SourceRegister: 1, SetName: "__set%d", SetID: 1},
		&expr.Meta{Key: expr.MetaKeyL4PROTO, Register: 1}, &expr.Cmp{Op: expr.CmpOpEq, Register: 1, Data: []byte{58}},
		&expr.Payload{DestRegister: 1, Base: expr.PayloadBaseTransportHeader, Offset: 0, Len: 1},
		&expr.Cmp{Op: expr.CmpOpEq, Register: 1, Data: []byte{icmpv6EchoRequest}},
		&expr.Verdict{Kind: expr.VerdictAccept},
	}
	if !reflect.DeepEqual(six, want) || len(added) != 1 || added[0].KeyType.Name != "ipv6_addr" {
		t.Fatalf("IPv6 exprs =\n%#v\nwant\n%#v", six, want)
	}
}

// Without CAP_NET_ADMIN every kernel call is refused, and says so; as root
// the test would change the host's tables, so it does not run.
func TestKernelNeedsPrivilege(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("as root this would change the host's nftables (the e2e tests cover it)")
	}
	k := NewKernel()
	rs, err := Compile(*testfix.Policy(), Local{Self: netip.MustParseAddr(testfix.VPS1.MeshIP)}, time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := k.Installed(); err == nil || !strings.Contains(err.Error(), "CAP_NET_ADMIN") {
		t.Errorf("Installed = %v", err)
	}
	if err := k.Replace(rs, rs.Fingerprint()); err == nil || !strings.Contains(err.Error(), "commit the table meshctl") {
		t.Errorf("Replace = %v", err)
	}
	if err := k.Remove(); err == nil || !strings.Contains(err.Error(), "delete the table meshctl") {
		t.Errorf("Remove = %v", err)
	}
	if _, err := k.Forget(netip.MustParsePrefix(testfix.MeshCIDR)); err == nil || !strings.Contains(err.Error(), "conntrack") {
		t.Errorf("Forget = %v", err)
	}
}
