//go:build linux

package policy

import (
	"fmt"
	"net"
	"net/netip"
	"strings"

	"github.com/google/nftables"
	"github.com/google/nftables/binaryutil"
	"github.com/google/nftables/expr"
	"github.com/google/nftables/userdata"
	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"
)

// NewKernel is the kernel of the agent's own network namespace: the host's
// for a server in host mode, the agent container's for a container-mode
// one.
func NewKernel() Kernel { return nftKernel{} }

type nftKernel struct{}

// markPrefix starts the comment on the table's first rule; the fingerprint
// follows.
const markPrefix = "meshctl policy "

// unmarked is what Installed reports for a table without our mark.
const unmarked = "unmarked"

var nftTable = &nftables.Table{Family: nftables.TableFamilyIPv4, Name: TableName}

func conn() (*nftables.Conn, error) {
	c, err := nftables.New()
	if err != nil {
		return nil, fmt.Errorf("open nftables: %w", err)
	}
	return c, nil
}

func (nftKernel) Installed() (string, error) {
	c, err := conn()
	if err != nil {
		return "", err
	}
	tables, err := c.ListTablesOfFamily(nftables.TableFamilyIPv4)
	if err != nil {
		return "", fmt.Errorf("list the nftables tables: %w (the agent needs CAP_NET_ADMIN)", err)
	}
	found := false
	for _, t := range tables {
		found = found || t.Name == TableName
	}
	if !found {
		return "", nil
	}
	rules, err := c.GetRules(nftTable, &nftables.Chain{Name: chainForward, Table: nftTable})
	if err != nil || len(rules) == 0 {
		return unmarked, nil // its chain is gone: replace it all
	}
	if s, ok := userdata.GetString(rules[0].UserData, userdata.TypeComment); ok {
		if fp, ok := strings.CutPrefix(s, markPrefix); ok {
			return fp, nil
		}
	}
	return unmarked, nil
}

func (nftKernel) Replace(rs *Ruleset, fingerprint string) error {
	c, err := conn()
	if err != nil {
		return err
	}
	b := build(rs, fingerprint)
	c.AddTable(b.table)
	c.FlushTable(b.table)
	for _, ch := range b.chains {
		c.AddChain(ch)
	}
	for _, op := range b.ops {
		if op.set != nil {
			if err := c.AddSet(op.set, op.elems); err != nil {
				return fmt.Errorf("prepare a set of the table %s: %w", TableName, err)
			}
			continue
		}
		c.AddRule(op.rule)
	}
	if err := c.Flush(); err != nil {
		return fmt.Errorf("commit the table %s: %w", TableName, err)
	}
	return nil
}

func (nftKernel) Remove() error {
	c, err := conn()
	if err != nil {
		return err
	}
	c.DelTable(nftTable)
	if err := c.Flush(); err != nil {
		return fmt.Errorf("delete the table %s: %w", TableName, err)
	}
	return nil
}

func (nftKernel) Forget(mesh netip.Prefix) (int, error) {
	ipnet := &net.IPNet{IP: mesh.Masked().Addr().AsSlice(), Mask: net.CIDRMask(mesh.Bits(), 32)}
	f := &netlink.ConntrackFilter{}
	if err := f.AddIPNet(netlink.ConntrackOrigSrcIP, ipnet); err != nil {
		return 0, fmt.Errorf("conntrack filter on %s: %w", mesh, err)
	}
	if err := f.AddIPNet(netlink.ConntrackOrigDstIP, ipnet); err != nil {
		return 0, fmt.Errorf("conntrack filter on %s: %w", mesh, err)
	}
	n, err := netlink.ConntrackDeleteFilters(netlink.ConntrackTable, unix.AF_INET, f)
	if err != nil {
		return int(n), fmt.Errorf("delete the conntrack entries of %s: %w", mesh, err)
	}
	return int(n), nil
}

// nftOp is one message of the batch, in order: an anonymous set before the
// rule that looks it up.
type nftOp struct {
	set   *nftables.Set
	elems []nftables.SetElement
	rule  *nftables.Rule
}

type nftBatch struct {
	table  *nftables.Table
	chains []*nftables.Chain
	ops    []nftOp
}

var nftHooks = map[string]*nftables.ChainHook{
	HookForward: nftables.ChainHookForward,
	HookInput:   nftables.ChainHookInput,
	HookOutput:  nftables.ChainHookOutput,
}

var nftProtos = map[string]byte{"tcp": unix.IPPROTO_TCP, "udp": unix.IPPROTO_UDP, "icmp": unix.IPPROTO_ICMP}

// icmpEchoRequest is ICMP type 8.
const icmpEchoRequest = 8

// build turns rs into the batch Replace sends: the table, its chains (the
// base ones accepting what no rule decides), then each rule with its
// anonymous sets; the first rule carries the fingerprint.
func build(rs *Ruleset, fingerprint string) nftBatch {
	b := nftBatch{table: nftTable}
	chains := map[string]*nftables.Chain{}
	for _, c := range rs.Chains {
		ch := &nftables.Chain{Name: c.Name, Table: nftTable}
		if c.Hook != "" {
			accept := nftables.ChainPolicyAccept
			ch.Type, ch.Hooknum, ch.Priority, ch.Policy = nftables.ChainTypeFilter, nftHooks[c.Hook], nftables.ChainPriorityFilter, &accept
		}
		chains[c.Name] = ch
		b.chains = append(b.chains, ch)
	}
	var setID uint32
	set := func(key nftables.SetDatatype, elems []nftables.SetElement) *nftables.Set {
		setID++ // unique in the batch, where the kernel resolves it
		s := &nftables.Set{Table: nftTable, ID: setID, Name: "__set%d", Anonymous: true, Constant: true, Interval: true, KeyType: key}
		b.ops = append(b.ops, nftOp{set: s, elems: elems})
		return s
	}
	first := true
	for _, c := range rs.Chains {
		for _, r := range c.Rules {
			rule := &nftables.Rule{Table: nftTable, Chain: chains[c.Name], Exprs: exprs(r, set)}
			if first {
				rule.UserData = userdata.AppendString(nil, userdata.TypeComment, markPrefix+fingerprint)
				first = false
			}
			b.ops = append(b.ops, nftOp{rule: rule})
		}
	}
	return b
}

// ifname is an interface name as the kernel compares it: 16 bytes, NUL
// padded.
func ifname(n string) []byte {
	b := make([]byte, unix.IFNAMSIZ)
	copy(b, n)
	return b
}

func exprs(r Rule, set func(nftables.SetDatatype, []nftables.SetElement) *nftables.Set) []expr.Any {
	var x []expr.Any
	if r.IIF != "" {
		x = append(x, &expr.Meta{Key: expr.MetaKeyIIFNAME, Register: 1}, &expr.Cmp{Op: expr.CmpOpEq, Register: 1, Data: ifname(r.IIF)})
	}
	if r.OIF != "" {
		x = append(x, &expr.Meta{Key: expr.MetaKeyOIFNAME, Register: 1}, &expr.Cmp{Op: expr.CmpOpEq, Register: 1, Data: ifname(r.OIF)})
	}
	if r.State != "" {
		bits := uint32(expr.CtStateBitESTABLISHED | expr.CtStateBitRELATED)
		if r.State == StateInvalid {
			bits = uint32(expr.CtStateBitINVALID)
		}
		x = append(x, &expr.Ct{Register: 1, Key: expr.CtKeySTATE},
			&expr.Bitwise{SourceRegister: 1, DestRegister: 1, Len: 4, Mask: binaryutil.NativeEndian.PutUint32(bits),
				Xor: binaryutil.NativeEndian.PutUint32(0)},
			&expr.Cmp{Op: expr.CmpOpNeq, Register: 1, Data: []byte{0, 0, 0, 0}})
	}
	for _, side := range []struct {
		offset uint32
		ranges []Range
	}{{12, r.Src}, {16, r.Dst}} { // the IPv4 header's source and destination
		if side.ranges == nil {
			continue
		}
		s := set(nftables.TypeIPAddr, intervalElems(side.ranges, 4))
		x = append(x, &expr.Payload{DestRegister: 1, Base: expr.PayloadBaseNetworkHeader, Offset: side.offset, Len: 4},
			&expr.Lookup{SourceRegister: 1, SetName: s.Name, SetID: s.ID})
	}
	if r.Proto != "" {
		x = append(x, &expr.Meta{Key: expr.MetaKeyL4PROTO, Register: 1}, &expr.Cmp{Op: expr.CmpOpEq, Register: 1, Data: []byte{nftProtos[r.Proto]}})
	}
	if r.Ports != nil {
		s := set(nftables.TypeInetService, intervalElems(r.Ports, 2))
		x = append(x, &expr.Payload{DestRegister: 1, Base: expr.PayloadBaseTransportHeader, Offset: 2, Len: 2},
			&expr.Lookup{SourceRegister: 1, SetName: s.Name, SetID: s.ID})
	}
	if r.Echo {
		x = append(x, &expr.Payload{DestRegister: 1, Base: expr.PayloadBaseTransportHeader, Offset: 0, Len: 1},
			&expr.Cmp{Op: expr.CmpOpEq, Register: 1, Data: []byte{icmpEchoRequest}})
	}
	switch r.Verdict {
	case Accept:
		x = append(x, &expr.Verdict{Kind: expr.VerdictAccept})
	case Drop:
		x = append(x, &expr.Verdict{Kind: expr.VerdictDrop})
	case Jump:
		x = append(x, &expr.Verdict{Kind: expr.VerdictJump, Chain: r.Target})
	}
	return x
}

// intervalElems are merged ranges as an interval set's elements, as nft
// sends them: each range's start, and the first value past its end marked
// as an interval end; an interval end at 0 first when no range starts
// there; none past the largest key.
func intervalElems(rs []Range, size int) []nftables.SetElement {
	key := func(v uint64) []byte {
		b := make([]byte, size)
		for i := size - 1; i >= 0; i-- {
			b[i] = byte(v)
			v >>= 8
		}
		return b
	}
	top := uint64(1)<<(8*size) - 1
	var out []nftables.SetElement
	if len(rs) > 0 && rs[0].First != 0 {
		out = append(out, nftables.SetElement{Key: key(0), IntervalEnd: true})
	}
	for _, r := range rs {
		out = append(out, nftables.SetElement{Key: key(uint64(r.First))})
		if uint64(r.Last) < top {
			out = append(out, nftables.SetElement{Key: key(uint64(r.Last) + 1), IntervalEnd: true})
		}
	}
	return out
}
