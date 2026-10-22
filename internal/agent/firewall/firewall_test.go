package firewall

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/hafio/vpn-mesh-utility/internal/agent/execrun"
	"github.com/hafio/vpn-mesh-utility/internal/proto"
	"github.com/hafio/vpn-mesh-utility/internal/testfix"
)

var update = flag.Bool("update", false, "rewrite testdata goldens")

// fakeRunner answers firewall-cmd queries from a table and records every
// call. It applies the same gate as the real runner. A failing call still
// prints its answer, as `--query-rich-rule` prints "no" with exit status 1.
type fakeRunner struct {
	answers map[string]string
	errs    map[string]error
	calls   [][]string
}

func (f *fakeRunner) Run(_ context.Context, argv []string) (string, error) {
	if err := execrun.Check(argv); err != nil {
		return "", err
	}
	f.calls = append(f.calls, argv)
	key := strings.Join(argv, " ")
	if err, ok := f.errs[key]; ok {
		return f.answers[key], err
	}
	return f.answers[key], nil
}

var queries = []string{"--state", "--version", "--get-zones", "--list-interfaces", "--get-policies", "--list-ports", "--list-rich-rules"}

// mutations are the calls that change firewalld.
func (f *fakeRunner) mutations() [][]string {
	var out [][]string
	for _, c := range f.calls {
		q := strings.HasPrefix(c[len(c)-1], "--info-policy=") || strings.HasPrefix(c[len(c)-1], "--query-") ||
			strings.HasPrefix(c[len(c)-1], "--get-zone-of-interface=")
		for _, s := range queries {
			q = q || c[len(c)-1] == s
		}
		if !q {
			out = append(out, c)
		}
	}
	return out
}

// freshHost is firewalld 1.3 with the stock zones and policies, and no mesh
// zone. At runtime the zone holds wg0 once it exists, as it does after the
// reload that creates it.
func freshHost() *fakeRunner {
	return &fakeRunner{answers: map[string]string{
		"firewall-cmd --state":                                   "running\n",
		"firewall-cmd --version":                                 "1.3.4\n",
		"firewall-cmd --permanent --get-zones":                   "block dmz drop external home internal nm-shared public trusted work\n",
		"firewall-cmd --permanent --zone=mesh --list-interfaces": "\n",
		"firewall-cmd --zone=mesh --list-interfaces":             "wg0\n",
		"firewall-cmd --permanent --get-policies":                "allow-host-ipv6\n",
	}}
}

// serverRich is the one rich rule of testfix.ServerConfig, as firewalld
// lists it.
const serverRich = `rule family="ipv4" source address="10.77.0.2/32" port port="2377" protocol="tcp" accept`

// appliedHost is freshHost after testfix.ServerConfig was applied: the zone
// exists, holds wg0, and holds the rules at runtime.
func appliedHost(r *fakeRunner) {
	r.answers["firewall-cmd --permanent --get-zones"] += " mesh"
	r.answers["firewall-cmd --permanent --zone=mesh --list-interfaces"] = "wg0\n"
	r.answers["firewall-cmd --zone=mesh --list-rich-rules"] = serverRich + "\n"
	r.answers["firewall-cmd --zone=public --list-ports"] = "51820/udp\n"
}

func newFirewall(t *testing.T, r execrun.Runner) *Firewall {
	return &Firewall{Run: r, Dir: t.TempDir(), Iface: "wg0"}
}

func encodeCalls(calls [][]string) []byte {
	var b bytes.Buffer
	for _, c := range calls {
		line, _ := json.Marshal(c)
		b.Write(line)
		b.WriteByte('\n')
	}
	return b.Bytes()
}

func golden(t *testing.T, name string, got []byte) {
	t.Helper()
	path := filepath.Join("testdata", name)
	if *update {
		if err := os.WriteFile(path, got, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read golden: %v (run ./scripts/dev.sh regen to create it)", err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("argv differs from %s:\ngot:\n%s\nwant:\n%s", path, got, want)
	}
}

func TestFreshHostArgvGolden(t *testing.T) {
	r := freshHost()
	f := newFirewall(t, r)
	cfg := testfix.ServerConfig()
	n, err := f.Apply(context.Background(), cfg.Firewall)
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	// new zone, set target, bind, reload, then one change per rule.
	if n != 6 {
		t.Errorf("Apply reported %d changes, want 6", n)
	}
	golden(t, "fresh.golden.jsonl", encodeCalls(r.calls))
}

func TestApplyIsIdempotent(t *testing.T) {
	r := freshHost()
	f := newFirewall(t, r)
	cfg := testfix.ServerConfig()
	if _, err := f.Apply(context.Background(), cfg.Firewall); err != nil {
		t.Fatal(err)
	}
	r.calls = nil
	n, err := f.Apply(context.Background(), cfg.Firewall)
	if err != nil || n != 0 || len(r.calls) != 0 {
		t.Fatalf("second Apply = %d changes, %v, calls %q; want nothing", n, err, r.calls)
	}

	// A restarted agent re-checks the zone and the rules but changes nothing.
	appliedHost(r)
	restarted := &Firewall{Run: r, Dir: f.Dir, Iface: "wg0"}
	if n, err := restarted.Apply(context.Background(), cfg.Firewall); err != nil || n != 0 || len(r.mutations()) != 0 {
		t.Fatalf("after restart: %d changes, %v, mutations %q", n, err, r.mutations())
	}
}

func TestRemovedRuleIsRemovedBothWays(t *testing.T) {
	r := freshHost()
	f := newFirewall(t, r)
	cfg := testfix.ServerConfig()
	if _, err := f.Apply(context.Background(), cfg.Firewall); err != nil {
		t.Fatal(err)
	}
	r.calls = nil
	cfg.Firewall.Rules = cfg.Firewall.Rules[1:] // vps-2 left the mesh
	n, err := f.Apply(context.Background(), cfg.Firewall)
	if err != nil || n != 1 {
		t.Fatalf("Apply = %d, %v", n, err)
	}
	rich := `rule family="ipv4" source address="10.77.0.2/32" port port="2377" protocol="tcp" accept`
	want := [][]string{
		{"firewall-cmd", "--zone=mesh", "--remove-rich-rule=" + rich},
		{"firewall-cmd", "--permanent", "--zone=mesh", "--remove-rich-rule=" + rich},
	}
	if got := encodeCalls(r.calls); !bytes.Equal(got, encodeCalls(want)) {
		t.Fatalf("calls:\n%s\nwant:\n%s", got, encodeCalls(want))
	}
	b, _ := os.ReadFile(filepath.Join(f.Dir, StateFile))
	if strings.Contains(string(b), "2377") {
		t.Fatalf("state still records the removed rule:\n%s", b)
	}
}

func TestNilFirewallRemovesOnlyOurRules(t *testing.T) {
	r := freshHost()
	f := newFirewall(t, r)
	cfg := testfix.ServerConfig()
	if _, err := f.Apply(context.Background(), cfg.Firewall); err != nil {
		t.Fatal(err)
	}
	r.calls = nil
	if n, err := f.Apply(context.Background(), nil); err != nil || n != 2 || len(r.calls) != 4 {
		t.Fatalf("Apply(nil) = %d, %v, calls %q", n, err, r.calls)
	}
	r.calls = nil
	if n, err := f.Apply(context.Background(), nil); err != nil || n != 0 || len(r.calls) != 0 {
		t.Fatalf("second Apply(nil) = %d, %v, calls %q", n, err, r.calls)
	}
}

// Injection: a value that would change the rich rule's meaning is refused
// before the first command runs.
func TestHostileRulesNeverReachFirewallCmd(t *testing.T) {
	cases := map[string]proto.FWRule{
		"quote in source":   {Zone: "mesh", Proto: "tcp", Port: 2377, Source: `10.0.0.0/8" accept rule family="ipv4`},
		"zone with option":  {Zone: "mesh --reload", Proto: "tcp", Port: 2377},
		"unknown zone":      {Zone: "trusted", Proto: "tcp", Port: 22},
		"proto with option": {Zone: "mesh", Proto: "tcp --permanent", Port: 2377},
		"port out of range": {Zone: "public", Proto: "udp", Port: 70000},
		"source is a host":  {Zone: "mesh", Proto: "tcp", Port: 2377, Source: "10.77.0.2/24"},
		"ping with a port":  {Zone: "mesh", Proto: proto.ProtoICMP, Port: 22, Source: "10.77.0.0/16"},
		"ping from anyone":  {Zone: "mesh", Proto: proto.ProtoICMP},
		"all from anyone":   {Zone: "mesh", Proto: proto.ProtoAll},
		"all with a port":   {Zone: "mesh", Proto: proto.ProtoAll, Port: 22, Source: "10.77.0.0/16"},
		"all in public":     {Zone: "public", Proto: proto.ProtoAll, Source: "10.77.0.0/16"},
	}
	for name, rule := range cases {
		t.Run(name, func(t *testing.T) {
			r := freshHost()
			f := newFirewall(t, r)
			fw := &proto.Firewall{Backend: "firewalld", Rules: []proto.FWRule{rule}}
			if _, err := f.Apply(context.Background(), fw); err == nil {
				t.Fatal("Apply accepted a hostile rule")
			}
			if len(r.calls) != 0 {
				t.Fatalf("firewall-cmd ran %d times: %q", len(r.calls), r.calls)
			}
		})
	}
}

// A ping rule is a rich rule accepting ICMP echo requests from its source,
// both at runtime and permanently, and passes the argv gate.
func TestPingRule(t *testing.T) {
	rule := proto.FWRule{Zone: proto.ZoneMesh, Proto: proto.ProtoICMP, Source: "10.77.0.0/16", Note: "ping"}
	want := []string{"firewall-cmd", "--permanent", "--zone=mesh",
		`--add-rich-rule=rule family="ipv4" source address="10.77.0.0/16" icmp-type name="echo-request" accept`}
	if got := RuleArgv(rule, "add", true); !slices.Equal(got, want) {
		t.Fatalf("RuleArgv = %q", got)
	}
	r := freshHost()
	f := newFirewall(t, r)
	if _, err := f.Apply(context.Background(), &proto.Firewall{Backend: "firewalld", Rules: []proto.FWRule{rule}}); err != nil {
		t.Fatal(err)
	}
	added := 0
	for _, c := range r.calls {
		if strings.Contains(strings.Join(c, " "), `icmp-type name="echo-request"`) {
			added++
		}
	}
	if added != 2 {
		t.Fatalf("the ping rule was added %d times (want runtime and permanent): %q", added, r.calls)
	}
}

// The mesh firewall's zone rule (amendment A43) is a rich rule accepting
// everything from the mesh, both at runtime and permanently.
func TestAcceptAllRule(t *testing.T) {
	rule := proto.FWRule{Zone: proto.ZoneMesh, Proto: proto.ProtoAll, Source: "10.77.0.0/16", Note: "segmentation: the policy decides"}
	want := []string{"firewall-cmd", "--zone=mesh", `--add-rich-rule=rule family="ipv4" source address="10.77.0.0/16" accept`}
	if got := RuleArgv(rule, "add", false); !slices.Equal(got, want) {
		t.Fatalf("RuleArgv = %q", got)
	}
	r := freshHost()
	if _, err := newFirewall(t, r).Apply(context.Background(), &proto.Firewall{Backend: "firewalld", Rules: []proto.FWRule{rule}}); err != nil {
		t.Fatal(err)
	}
	added := 0
	for _, c := range r.calls {
		if strings.HasSuffix(strings.Join(c, " "), `source address="10.77.0.0/16" accept`) && strings.Contains(strings.Join(c, " "), "--add-rich-rule") {
			added++
		}
	}
	if added != 2 {
		t.Fatalf("the accept-all rule was added %d times (want runtime and permanent): %q", added, r.calls)
	}
}

// A host-mode server's containers' bridge (amendment A37) goes into the
// mesh zone when no zone holds it -- a podman bridge the agent made -- and
// is left where Docker (its docker zone) or an operator put it.
func TestContainersBridgeZone(t *testing.T) {
	ctx := context.Background()
	query := "firewall-cmd --get-zone-of-interface=" + proto.ContainerBridge
	add := []string{program, "--zone=" + proto.ZoneMesh, "--add-interface=" + proto.ContainerBridge}
	addPermanent := []string{program, "--permanent", "--zone=" + proto.ZoneMesh, "--add-interface=" + proto.ContainerBridge}
	fw := testfix.ServerConfig().Firewall
	fw.Bridge, fw.Relay = proto.ContainerBridge, true

	r := freshHost()
	r.answers[query] = "no zone\n"
	r.errs = map[string]error{query: errors.New("exit status 2")}
	if _, err := newFirewall(t, r).Apply(ctx, fw); err != nil {
		t.Fatal(err)
	}
	if !slices.ContainsFunc(r.calls, func(c []string) bool { return slices.Equal(c, add) }) ||
		!slices.ContainsFunc(r.calls, func(c []string) bool { return slices.Equal(c, addPermanent) }) {
		t.Fatalf("an unzoned bridge was not put in the mesh zone: %q", r.calls)
	}

	r = freshHost()
	r.answers[query] = "docker\n"
	if _, err := newFirewall(t, r).Apply(ctx, fw); err != nil {
		t.Fatal(err)
	}
	if slices.ContainsFunc(r.calls, func(c []string) bool { return slices.Contains(c, "--add-interface="+proto.ContainerBridge) }) {
		t.Fatalf("a bridge in docker's zone was moved: %q", r.calls)
	}

	r = freshHost()
	r.errs = map[string]error{query: errors.New("exit status 2"), strings.Join(add, " "): errors.New("INVALID_INTERFACE")}
	if _, err := newFirewall(t, r).Apply(ctx, fw); err == nil || !strings.Contains(err.Error(), "put the containers' bridge meshctl0 in zone mesh") {
		t.Fatalf("a failed binding = %v", err)
	}

	r = freshHost()
	bad := *fw
	bad.Bridge = "docker0 --reload"
	if _, err := newFirewall(t, r).Apply(ctx, &bad); err == nil || len(r.calls) != 0 {
		t.Fatalf("a hostile bridge = %v, %q", err, r.calls)
	}
}

func TestMissingFirewalldIsActionable(t *testing.T) {
	r := freshHost()
	r.errs = map[string]error{"firewall-cmd --state": execrun.ErrNotInstalled}
	f := newFirewall(t, r)
	cfg := testfix.ServerConfig()
	_, err := f.Apply(context.Background(), cfg.Firewall)
	if err == nil || !strings.Contains(err.Error(), "--no-firewall") || len(r.mutations()) != 0 {
		t.Fatalf("Apply = %v, mutations %q", err, r.mutations())
	}
}

func TestOldFirewalldIsRefused(t *testing.T) {
	r := freshHost()
	r.answers["firewall-cmd --version"] = "0.8.2\n"
	f := newFirewall(t, r)
	cfg := testfix.ServerConfig()
	if _, err := f.Apply(context.Background(), cfg.Firewall); err == nil || !strings.Contains(err.Error(), "0.9") || len(r.mutations()) != 0 {
		t.Fatalf("Apply = %v, mutations %q", err, r.mutations())
	}
}

// policyInfo is `firewall-cmd --permanent --info-policy` for a policy as the
// agent makes it.
func policyInfo(name, egress, masquerade string) string {
	return name + "\n  priority: -1\n  target: ACCEPT\n  ingress-zones: mesh\n  egress-zones: " + egress +
		"\n  services: \n  ports: \n  protocols: \n  masquerade: " + masquerade + "\n  forward-ports: \n  rich rules: \n"
}

// A relaying exit server gets both forwarding policies, created --permanent
// and loaded with one reload; they are checked again only when the config
// changes them, repaired when edited by hand, and deleted when no longer
// wanted.
func TestForwardingPolicies(t *testing.T) {
	r := freshHost()
	f := newFirewall(t, r)
	cfg := testfix.ServerConfig()
	cfg.Firewall.Relay, cfg.Firewall.ExitGateway = true, true
	n, err := f.Apply(context.Background(), cfg.Firewall)
	if err != nil || n != 9 {
		t.Fatalf("Apply = %d, %v (zone 4, rules 2, policies 2, reload 1)", n, err)
	}
	golden(t, "policies.golden.jsonl", encodeCalls(r.mutations()[8:]))

	r.calls = nil
	if n, err := f.Apply(context.Background(), cfg.Firewall); err != nil || n != 0 || len(r.calls) != 0 {
		t.Fatalf("an unchanged config = %d, %v, calls %q", n, err, r.calls)
	}

	// A restarted agent finds both in place and changes nothing.
	r.answers["firewall-cmd --permanent --get-policies"] = "allow-host-ipv6 meshctl-exit meshctl-relay\n"
	r.answers["firewall-cmd --permanent --info-policy=meshctl-relay"] = policyInfo("meshctl-relay", "mesh", "no")
	r.answers["firewall-cmd --permanent --info-policy=meshctl-exit"] = policyInfo("meshctl-exit", "public", "yes")
	restarted := &Firewall{Run: r, Dir: f.Dir, Iface: "wg0", ready: true}
	if n, err := restarted.Apply(context.Background(), cfg.Firewall); err != nil || n != 0 || len(r.mutations()) != 0 {
		t.Fatalf("after restart: %d, %v, mutations %q", n, err, r.mutations())
	}

	// Someone turned masquerade off: the policy is put back as a whole.
	r.answers["firewall-cmd --permanent --info-policy=meshctl-exit"] = policyInfo("meshctl-exit", "public", "no")
	restarted = &Firewall{Run: r, Dir: f.Dir, Iface: "wg0", ready: true}
	r.calls = nil
	if n, err := restarted.Apply(context.Background(), cfg.Firewall); err != nil || n != 2 {
		t.Fatalf("repair = %d, %v", n, err)
	}
	if m := r.mutations(); len(m) != 7 || m[0][2] != "--delete-policy=meshctl-exit" || m[len(m)-1][1] != "--reload" {
		t.Fatalf("repair mutations %q", m)
	}

	// The node stops relaying: only the relay policy goes.
	r.answers["firewall-cmd --permanent --info-policy=meshctl-exit"] = policyInfo("meshctl-exit", "public", "yes")
	cfg.Firewall.Relay = false
	r.calls = nil
	if n, err := restarted.Apply(context.Background(), cfg.Firewall); err != nil || n != 2 {
		t.Fatalf("relay off = %d, %v", n, err)
	}
	want := [][]string{{"firewall-cmd", "--permanent", "--delete-policy=meshctl-relay"}, {"firewall-cmd", "--reload"}}
	if got := encodeCalls(r.mutations()); !bytes.Equal(got, encodeCalls(want)) {
		t.Fatalf("relay off:\n%s\nwant:\n%s", got, encodeCalls(want))
	}
}

// An exit with IPv6 (M8) masquerades the mesh's IPv6 prefix by a rich rule
// in its policy: firewalld's policy masquerade is IPv4 only.
func TestExitPolicyMasqueradesIPv6(t *testing.T) {
	r := freshHost()
	f := newFirewall(t, r)
	cfg := testfix.ServerConfig()
	cfg.Firewall.ExitGateway, cfg.Firewall.Exit6 = true, testfix.MeshCIDR6
	if _, err := f.Apply(context.Background(), cfg.Firewall); err != nil {
		t.Fatal(err)
	}
	rule := `rule family="ipv6" source address="` + testfix.MeshCIDR6 + `" masquerade`
	add := []string{"firewall-cmd", "--permanent", "--policy=meshctl-exit", "--add-rich-rule=" + rule}
	if !slices.ContainsFunc(r.mutations(), func(m []string) bool { return slices.Equal(m, add) }) {
		t.Fatalf("no IPv6 masquerade in %q", r.mutations())
	}

	// A restarted agent finds it in place; without the rule it is put back.
	info := policyInfo("meshctl-exit", "public", "yes") + "\t" + rule + "\n"
	r.answers["firewall-cmd --permanent --get-policies"] = "meshctl-exit\n"
	r.answers["firewall-cmd --permanent --info-policy=meshctl-exit"] = info
	r.calls = nil
	restarted := &Firewall{Run: r, Dir: f.Dir, Iface: "wg0", ready: true}
	if n, err := restarted.Apply(context.Background(), cfg.Firewall); err != nil || n != 0 {
		t.Fatalf("after restart: %d, %v, mutations %q", n, err, r.mutations())
	}
	r.answers["firewall-cmd --permanent --info-policy=meshctl-exit"] = policyInfo("meshctl-exit", "public", "yes")
	restarted = &Firewall{Run: r, Dir: f.Dir, Iface: "wg0", ready: true}
	if n, err := restarted.Apply(context.Background(), cfg.Firewall); err != nil || n != 2 {
		t.Fatalf("repair = %d, %v", n, err)
	}
	// The exit loses its IPv6: the policy goes back to IPv4 only.
	r.answers["firewall-cmd --permanent --info-policy=meshctl-exit"] = info
	cfg.Firewall.Exit6 = ""
	r.calls = nil
	if n, err := restarted.Apply(context.Background(), cfg.Firewall); err != nil || n != 2 ||
		slices.ContainsFunc(r.mutations(), func(m []string) bool { return strings.Contains(strings.Join(m, " "), "rich-rule") }) {
		t.Fatalf("without IPv6 = %d, %v, mutations %q", n, err, r.mutations())
	}
	// A hostile prefix never reaches firewall-cmd.
	r.calls = nil
	cfg.Firewall.Exit6 = testfix.MeshCIDR6 + `" accept`
	if _, err := newFirewall(t, r).Apply(context.Background(), cfg.Firewall); err == nil || !strings.Contains(err.Error(), "firewall.exit6") || len(r.calls) != 0 {
		t.Fatalf("hostile exit6 = %v, calls %q", err, r.calls)
	}
}

func TestPolicyFailureIsReported(t *testing.T) {
	r := freshHost()
	r.errs = map[string]error{"firewall-cmd --permanent --policy=meshctl-exit --add-masquerade": errors.New("INVALID_ZONE: masquerade cannot be used")}
	f := newFirewall(t, r)
	cfg := testfix.ServerConfig()
	cfg.Firewall.ExitGateway = true
	if _, err := f.Apply(context.Background(), cfg.Firewall); err == nil || !strings.Contains(err.Error(), "policy meshctl-exit") {
		t.Fatalf("Apply = %v", err)
	}
	// Not recorded as in place: the next apply tries again.
	r.calls = nil
	_, _ = f.Apply(context.Background(), cfg.Firewall) // fails the same way; the calls are what matter
	if len(r.calls) == 0 {
		t.Fatal("a failed policy sync was cached")
	}
}

// The drift check: after Recheck the next Apply asks firewalld again and
// puts back what changed underneath the agent -- wg0 moved to another zone
// at runtime, a rule removed by hand -- and nothing else.
func TestRecheckPutsBackWhatDrifted(t *testing.T) {
	r := freshHost()
	f := newFirewall(t, r)
	cfg := testfix.ServerConfig()
	if _, err := f.Apply(context.Background(), cfg.Firewall); err != nil {
		t.Fatal(err)
	}
	appliedHost(r)
	r.calls = nil
	f.Recheck()
	if n, err := f.Apply(context.Background(), cfg.Firewall); err != nil || n != 0 || len(r.mutations()) != 0 {
		t.Fatalf("a recheck of an unchanged host = %d, %v, mutations %q", n, err, r.mutations())
	}

	r.answers["firewall-cmd --zone=mesh --list-interfaces"] = "\n" // NetworkManager moved it
	r.answers["firewall-cmd --zone=mesh --list-rich-rules"] = "\n" // removed by hand
	key := "firewall-cmd --zone=mesh --query-rich-rule=" + serverRich
	r.answers[key], r.errs = "no\n", map[string]error{key: errors.New("exit status 1")}
	r.calls = nil
	f.Recheck()
	n, err := f.Apply(context.Background(), cfg.Firewall)
	want := [][]string{
		{"firewall-cmd", "--zone=mesh", "--change-interface=wg0"},
		{"firewall-cmd", "--zone=mesh", "--add-rich-rule=" + serverRich},
		{"firewall-cmd", "--permanent", "--zone=mesh", "--add-rich-rule=" + serverRich},
	}
	if err != nil || n != 2 || !bytes.Equal(encodeCalls(r.mutations()), encodeCalls(want)) {
		t.Fatalf("drift = %d, %v, mutations:\n%s\nwant:\n%s", n, err, encodeCalls(r.mutations()), encodeCalls(want))
	}
	// Without a Recheck the same process trusts what it did.
	r.calls = nil
	if n, err := f.Apply(context.Background(), cfg.Firewall); err != nil || n != 0 || len(r.calls) != 0 {
		t.Fatalf("no recheck = %d, %v, calls %q", n, err, r.calls)
	}
}

// firewalld may list a rich rule in other words than it was added in; the
// agent asks before adding it twice.
func TestRecheckAsksAboutARuleSpelledDifferently(t *testing.T) {
	r := freshHost()
	f := newFirewall(t, r)
	cfg := testfix.ServerConfig()
	if _, err := f.Apply(context.Background(), cfg.Firewall); err != nil {
		t.Fatal(err)
	}
	appliedHost(r)
	r.answers["firewall-cmd --zone=mesh --list-rich-rules"] = strings.Replace(serverRich, "/32", "", 1) + "\n"
	r.answers["firewall-cmd --zone=mesh --query-rich-rule="+serverRich] = "yes\n"
	r.calls = nil
	f.Recheck()
	if n, err := f.Apply(context.Background(), cfg.Firewall); err != nil || n != 0 || len(r.mutations()) != 0 {
		t.Fatalf("= %d, %v, mutations %q", n, err, r.mutations())
	}
	if !slices.ContainsFunc(r.calls, func(c []string) bool { return strings.HasPrefix(c[len(c)-1], "--query-rich-rule=") }) {
		t.Fatal("the rule was not confirmed with firewalld")
	}
}

func TestAtLeast(t *testing.T) {
	for v, want := range map[string]bool{"0.9.0": true, "1.3.4\n": true, "0.8.2": false, "2": false, "x.y": false} {
		if got := atLeast(v, 0, 9); got != want {
			t.Errorf("atLeast(%q) = %v", v, got)
		}
	}
}

func TestDamagedStateIsActionable(t *testing.T) {
	f := newFirewall(t, freshHost())
	if err := os.WriteFile(filepath.Join(f.Dir, StateFile), []byte("{"), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg := testfix.ServerConfig()
	if _, err := f.Apply(context.Background(), cfg.Firewall); err == nil || !strings.Contains(err.Error(), "remove it") {
		t.Fatalf("Apply = %v", err)
	}
}

// Every firewall-cmd call can fail. The apply stops at the failing call and
// returns its error, so nothing runs on top of a half-made change.
func TestEachFailingCallStopsTheApply(t *testing.T) {
	host := func() *fakeRunner {
		r := freshHost()
		// A relay policy edited by hand: checked, deleted and made again.
		r.answers["firewall-cmd --permanent --get-policies"] = "allow-host-ipv6 meshctl-relay\n"
		r.answers["firewall-cmd --permanent --info-policy=meshctl-relay"] = policyInfo("meshctl-relay", "public", "no")
		return r
	}
	cfg := testfix.ServerConfig()
	cfg.Firewall.Relay = true
	clean := host()
	if _, err := newFirewall(t, clean).Apply(context.Background(), cfg.Firewall); err != nil {
		t.Fatal(err)
	}
	for i, call := range clean.calls {
		if slices.IndexFunc(clean.calls, func(c []string) bool { return slices.Equal(c, call) }) != i {
			continue // a second --reload: failing it fails the first
		}
		r := host()
		r.errs = map[string]error{strings.Join(call, " "): errors.New("DBUS_EXCEPTION: firewalld is reloading")}
		_, err := newFirewall(t, r).Apply(context.Background(), cfg.Firewall)
		switch {
		case err == nil || !strings.Contains(err.Error(), "DBUS_EXCEPTION"):
			t.Errorf("failing %q: Apply = %v", call, err)
		case len(r.calls) != i+1:
			t.Errorf("failing %q: %d calls ran, want %d (none after the failure)", call, len(r.calls), i+1)
		}
	}
}
