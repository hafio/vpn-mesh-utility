// Package firewall renders a server node's firewall section into firewall-cmd
// invocations and applies them (spec S7.5, amendment A5).
//
// The agent owns the firewalld zone "mesh" (bound to the mesh interface,
// target DROP) and the openings it made in the public zone. What it opened is
// recorded in firewall.applied.json in the data dir, so a rule that leaves the
// config is removed on the next apply -- and nothing an operator opened by
// hand is ever touched. Every change is made at runtime and --permanent, so it
// is live at once and survives a reboot or `firewall-cmd --reload`.
//
// Forwarding is two firewalld policies the agent owns by name (amendment A4):
// meshctl-relay (mesh to mesh, on a server that relays between spokes) and
// meshctl-exit (mesh to public with masquerade, on an exit server). They are
// written --permanent and loaded with one --reload, created when the config
// asks for them, put back as a whole when they were changed, and deleted
// when it no longer does. Policy masquerade needs firewalld's nftables
// backend, the default since EL9.
//
// firewalld is asked again on the first apply in a process and after
// Recheck (the agent's drift check): the zone and its interface, at runtime
// too, since NetworkManager may move an interface to another zone; the
// recorded rules at runtime, so one removed by hand is added again; and the
// policies.
package firewall

import (
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/hafio/vpn-mesh-utility/internal/agent/execrun"
	"github.com/hafio/vpn-mesh-utility/internal/fsutil"
	"github.com/hafio/vpn-mesh-utility/internal/proto"
	"github.com/hafio/vpn-mesh-utility/internal/validate"
)

// StateFile records the rules the agent added.
const StateFile = "firewall.applied.json"

const program = "firewall-cmd"

// Firewall applies firewall sections through firewall-cmd.
type Firewall struct {
	Run execrun.Runner
	// Dir is the agent's data dir, which holds StateFile.
	Dir string
	// Iface is the mesh interface the zone is bound to.
	Iface string
	// ready is set once the zone is known to exist in this process, so the
	// checks are not repeated on every apply.
	ready bool
	// synced records the policies last made to match in this process; they
	// are checked again only when the wanted set changes.
	synced   bool
	policies []string
}

type applied struct {
	Rules []proto.FWRule `json:"rules"`
}

// Apply makes firewalld match fw and returns how many changes it made. A nil
// fw (the node is no longer a server) removes the rules the agent added.
func (f *Firewall) Apply(ctx context.Context, fw *proto.Firewall) (n int, err error) {
	var want []proto.FWRule
	if fw != nil {
		// Every value that reaches an argv is checked before the first
		// command runs; execrun checks the argv again.
		if err := check(*fw); err != nil {
			return 0, err
		}
		want = normalize(fw.Rules)
	}
	st, err := f.load()
	if err != nil {
		return 0, err
	}
	if fw == nil && len(st.Rules) == 0 {
		return 0, nil
	}
	before := len(st.Rules)
	if fw != nil && !f.ready {
		c, err := f.ensureZone(ctx)
		if err != nil {
			return c, err
		}
		if fw.Bridge != "" {
			b, err := f.ensureBridge(ctx, fw.Bridge)
			c += b
			if err != nil {
				return c, err
			}
		}
		if st.Rules, err = f.present(ctx, st.Rules); err != nil {
			return c, err
		}
		f.ready, n = true, c
	}

	defer func() {
		if n > 0 || len(st.Rules) != before {
			if serr := f.save(st); serr != nil {
				err = errors.Join(err, serr)
			}
		}
	}()
	for _, r := range slices.Clone(st.Rules) {
		if !slices.Contains(want, r) {
			if err := f.both(ctx, r, "remove"); err != nil {
				return n, err
			}
			st.Rules = slices.DeleteFunc(st.Rules, func(x proto.FWRule) bool { return x == r })
			n++
		}
	}
	for _, r := range want {
		if !slices.Contains(st.Rules, r) {
			if err := f.both(ctx, r, "add"); err != nil {
				return n, err
			}
			st.Rules = append(st.Rules, r)
			n++
		}
	}
	if fw != nil {
		c, err := f.syncPolicies(ctx, wantedPolicies(*fw))
		n += c
		if err != nil {
			return n, err
		}
	}
	return n, nil
}

// Recheck makes the next Apply ask firewalld again instead of trusting what
// this process saw: the agent calls it before its periodic re-apply.
func (f *Firewall) Recheck() { f.ready, f.synced = false, false }

// present returns the recorded rules firewalld holds at runtime, so the
// apply adds the others again. A rich rule firewalld lists in other words
// than it was added in is confirmed with --query-rich-rule.
func (f *Firewall) present(ctx context.Context, rules []proto.FWRule) ([]proto.FWRule, error) {
	listed := map[string][]string{}
	var out []proto.FWRule
	for _, r := range rules {
		have, ok := listed[r.Zone]
		if !ok {
			for _, list := range []string{"--list-ports", "--list-rich-rules"} {
				o, err := f.Run.Run(ctx, []string{program, "--zone=" + r.Zone, list})
				if err != nil {
					return nil, fmt.Errorf("firewall: %w", err)
				}
				if list == "--list-ports" {
					have = append(have, strings.Fields(o)...)
				} else {
					for _, line := range strings.Split(o, "\n") {
						have = append(have, strings.TrimSpace(line))
					}
				}
			}
			listed[r.Zone] = have
		}
		switch {
		case r.Source == "" && slices.Contains(have, fmt.Sprintf("%d/%s", r.Port, r.Proto)):
		case r.Source != "" && slices.Contains(have, richRule(r)):
		case r.Source != "":
			// Exit status 1 with "no" means absent; any other failure is
			// left to the add, which reports it.
			if o, _ := f.Run.Run(ctx, RuleArgv(r, "query", false)); strings.TrimSpace(o) != "yes" {
				continue
			}
		default:
			continue
		}
		out = append(out, r)
	}
	return out, nil
}

// policy is a forwarding policy the agent owns.
type policy struct {
	name            string
	ingress, egress string
	masquerade      bool
}

var (
	// relayPolicy lets a server forward between mesh peers (spoke to spoke
	// through their relay, amendment A3).
	relayPolicy = policy{name: "meshctl-relay", ingress: proto.ZoneMesh, egress: proto.ZoneMesh}
	// exitPolicy lets an exit server forward mesh traffic to the internet,
	// masqueraded behind its public address.
	exitPolicy = policy{name: "meshctl-exit", ingress: proto.ZoneMesh, egress: proto.ZonePublic, masquerade: true}
	policies   = []policy{relayPolicy, exitPolicy}
)

func wantedPolicies(fw proto.Firewall) []string {
	var out []string
	if fw.Relay {
		out = append(out, relayPolicy.name)
	}
	if fw.ExitGateway {
		out = append(out, exitPolicy.name)
	}
	return out
}

// create is the argv that makes p from nothing.
func (p policy) create() [][]string {
	at := func(arg string) []string {
		return []string{program, "--permanent", "--policy=" + p.name, arg}
	}
	argv := [][]string{
		{program, "--permanent", "--new-policy=" + p.name},
		at("--add-ingress-zone=" + p.ingress),
		at("--add-egress-zone=" + p.egress),
		at("--set-target=ACCEPT"),
	}
	if p.masquerade {
		argv = append(argv, at("--add-masquerade"))
	}
	return argv
}

// matches reports whether `firewall-cmd --info-policy` output shows p as the
// agent made it.
func (p policy) matches(info string) bool {
	kv := map[string]string{}
	for _, line := range strings.Split(info, "\n") {
		if k, v, ok := strings.Cut(strings.TrimSpace(line), ":"); ok {
			kv[k] = strings.TrimSpace(v)
		}
	}
	masq := "no"
	if p.masquerade {
		masq = "yes"
	}
	return kv["target"] == "ACCEPT" && kv["ingress-zones"] == p.ingress && kv["egress-zones"] == p.egress && kv["masquerade"] == masq
}

// syncPolicies makes the agent's policies exactly want and returns how many
// changes it made. It asks firewalld on the first apply in this process and
// whenever want changes.
func (f *Firewall) syncPolicies(ctx context.Context, want []string) (int, error) {
	if f.synced && slices.Equal(f.policies, want) {
		return 0, nil
	}
	out, err := f.Run.Run(ctx, []string{program, "--permanent", "--get-policies"})
	if err != nil {
		return 0, fmt.Errorf("firewall: %w", err)
	}
	have := strings.Fields(out)
	n := 0
	for _, p := range policies {
		var argv [][]string
		wanted, exists := slices.Contains(want, p.name), slices.Contains(have, p.name)
		switch {
		case wanted && exists:
			info, err := f.Run.Run(ctx, []string{program, "--permanent", "--info-policy=" + p.name})
			if err != nil {
				return n, fmt.Errorf("firewall: %w", err)
			}
			if p.matches(info) {
				continue
			}
			// Changed by hand: put it back as a whole.
			argv = append([][]string{{program, "--permanent", "--delete-policy=" + p.name}}, p.create()...)
		case wanted:
			argv = p.create()
		case exists:
			argv = [][]string{{program, "--permanent", "--delete-policy=" + p.name}}
		default:
			continue
		}
		for _, a := range argv {
			if _, err := f.Run.Run(ctx, a); err != nil {
				return n, fmt.Errorf("firewall: policy %s: %w", p.name, err)
			}
		}
		n++
	}
	if n > 0 {
		if _, err := f.Run.Run(ctx, []string{program, "--reload"}); err != nil {
			return n, fmt.Errorf("firewall: %w", err)
		}
		n++
	}
	f.synced, f.policies = true, slices.Clone(want)
	return n, nil
}

// both applies one rule at runtime and permanently.
func (f *Firewall) both(ctx context.Context, r proto.FWRule, op string) error {
	for _, permanent := range []bool{false, true} {
		if _, err := f.Run.Run(ctx, RuleArgv(r, op, permanent)); err != nil {
			return fmt.Errorf("firewall: %s %s/%d in zone %s: %w", op, r.Proto, r.Port, r.Zone, err)
		}
	}
	return nil
}

// RuleArgv renders one rule. A rule without a source opens the port; one with
// a source is a rich rule that accepts only that prefix.
func RuleArgv(r proto.FWRule, op string, permanent bool) []string {
	argv := []string{program}
	if permanent {
		argv = append(argv, "--permanent")
	}
	argv = append(argv, "--zone="+r.Zone)
	if r.Source == "" {
		return append(argv, fmt.Sprintf("--%s-port=%d/%s", op, r.Port, r.Proto))
	}
	return append(argv, fmt.Sprintf("--%s-rich-rule=%s", op, richRule(r)))
}

// richRule is a rule with a source, in firewalld's own spelling (what
// --list-rich-rules prints).
func richRule(r proto.FWRule) string {
	return fmt.Sprintf(`rule family="ipv4" source address="%s" port port="%d" protocol="%s" accept`, r.Source, r.Port, r.Proto)
}

// ensureZone checks firewalld and creates the mesh zone once: created
// --permanent, then one --reload.
func (f *Firewall) ensureZone(ctx context.Context) (int, error) {
	out, err := f.Run.Run(ctx, []string{program, "--state"})
	switch {
	case errors.Is(err, execrun.ErrNotInstalled):
		return 0, errors.New("firewall: firewalld is not installed; install it (`dnf install firewalld`), or open udp/51820 yourself and run the agent with --no-firewall")
	case err != nil || strings.TrimSpace(out) != "running":
		return 0, fmt.Errorf("firewall: firewalld is not running (`systemctl enable --now firewalld`): %v", err)
	}
	out, err = f.Run.Run(ctx, []string{program, "--version"})
	if err != nil {
		return 0, fmt.Errorf("firewall: %w", err)
	}
	if !atLeast(out, 0, 9) {
		return 0, fmt.Errorf("firewall: firewalld %s is older than 0.9, which meshctl needs; upgrade it", strings.TrimSpace(out))
	}
	n, reload := 0, false
	out, err = f.Run.Run(ctx, []string{program, "--permanent", "--get-zones"})
	if err != nil {
		return 0, fmt.Errorf("firewall: %w", err)
	}
	if !slices.Contains(strings.Fields(out), proto.ZoneMesh) {
		for _, argv := range [][]string{
			{program, "--permanent", "--new-zone=" + proto.ZoneMesh},
			{program, "--permanent", "--zone=" + proto.ZoneMesh, "--set-target=DROP"},
		} {
			if _, err := f.Run.Run(ctx, argv); err != nil {
				return n, fmt.Errorf("firewall: create zone %s: %w", proto.ZoneMesh, err)
			}
			n++
		}
		reload = true
	}
	out, err = f.Run.Run(ctx, []string{program, "--permanent", "--zone=" + proto.ZoneMesh, "--list-interfaces"})
	if err != nil {
		return n, fmt.Errorf("firewall: %w", err)
	}
	if !slices.Contains(strings.Fields(out), f.Iface) {
		if _, err := f.Run.Run(ctx, []string{program, "--permanent", "--zone=" + proto.ZoneMesh, "--change-interface=" + f.Iface}); err != nil {
			return n, fmt.Errorf("firewall: bind %s to zone %s: %w", f.Iface, proto.ZoneMesh, err)
		}
		n++
		reload = true
	}
	if reload {
		if _, err := f.Run.Run(ctx, []string{program, "--reload"}); err != nil {
			return n, fmt.Errorf("firewall: %w", err)
		}
		n++
	}
	// At runtime too: NetworkManager may have moved the interface to
	// another zone, which neither the permanent binding nor a reload undoes.
	out, err = f.Run.Run(ctx, []string{program, "--zone=" + proto.ZoneMesh, "--list-interfaces"})
	if err != nil {
		return n, fmt.Errorf("firewall: %w", err)
	}
	if !slices.Contains(strings.Fields(out), f.Iface) {
		if _, err := f.Run.Run(ctx, []string{program, "--zone=" + proto.ZoneMesh, "--change-interface=" + f.Iface}); err != nil {
			return n, fmt.Errorf("firewall: bind %s to zone %s: %w", f.Iface, proto.ZoneMesh, err)
		}
		n++
	}
	return n, nil
}

// ensureBridge puts the containers' bridge in the mesh zone when no zone
// holds it (amendment A37), at runtime and permanently, so the relay policy
// forwards between it and wg0. Docker puts its own bridges in its docker
// zone, whose policy forwards from any zone; an operator's zone is left
// alone too. firewalld takes an interface by name before it exists, so the
// first apply may bind the bridge the containers step creates after it.
func (f *Firewall) ensureBridge(ctx context.Context, bridge string) (int, error) {
	// --get-zone-of-interface fails with "no zone" when none holds it.
	if out, err := f.Run.Run(ctx, []string{program, "--get-zone-of-interface=" + bridge}); err == nil && strings.TrimSpace(out) != "" {
		return 0, nil
	}
	n := 0
	for _, argv := range [][]string{
		{program, "--zone=" + proto.ZoneMesh, "--add-interface=" + bridge},
		{program, "--permanent", "--zone=" + proto.ZoneMesh, "--add-interface=" + bridge},
	} {
		if _, err := f.Run.Run(ctx, argv); err != nil {
			return n, fmt.Errorf("firewall: put the containers' bridge %s in zone %s: %w", bridge, proto.ZoneMesh, err)
		}
		n++
	}
	return n, nil
}

// atLeast compares a firewalld version string ("1.3.4") with major.minor.
func atLeast(v string, major, minor int) bool {
	parts := strings.Split(strings.TrimSpace(v), ".")
	if len(parts) < 2 {
		return false
	}
	ma, err1 := strconv.Atoi(parts[0])
	mi, err2 := strconv.Atoi(parts[1])
	if err1 != nil || err2 != nil {
		return false
	}
	return ma > major || (ma == major && mi >= minor)
}

func check(fw proto.Firewall) error {
	var errs []error
	add := func(err error) {
		if err != nil {
			errs = append(errs, err)
		}
	}
	_, err := validate.Enum("firewall.backend", fw.Backend, proto.FirewallFirewalld)
	add(err)
	if fw.Bridge != "" {
		_, err = validate.Enum("firewall.bridge", fw.Bridge, proto.ContainerBridge)
		add(err)
	}
	for i, r := range fw.Rules {
		field := fmt.Sprintf("firewall.rules[%d]", i)
		_, err = validate.Enum(field+".zone", r.Zone, proto.ZonePublic, proto.ZoneMesh)
		add(err)
		_, err = validate.Enum(field+".proto", r.Proto, "tcp", "udp")
		add(err)
		_, err = validate.Port(field+".port", r.Port)
		add(err)
		if r.Source != "" {
			_, err = validate.Prefix4(field+".source", r.Source)
			add(err)
		}
	}
	if err := errors.Join(errs...); err != nil {
		return fmt.Errorf("firewall: refusing the config: %w", err)
	}
	return nil
}

// normalize drops the display-only note, deduplicates and sorts, so the
// recorded state compares by what firewalld actually holds.
func normalize(rules []proto.FWRule) []proto.FWRule {
	out := make([]proto.FWRule, 0, len(rules))
	for _, r := range rules {
		r.Note = ""
		if !slices.Contains(out, r) {
			out = append(out, r)
		}
	}
	slices.SortFunc(out, func(a, b proto.FWRule) int {
		return cmp.Or(cmp.Compare(a.Zone, b.Zone), cmp.Compare(a.Port, b.Port), cmp.Compare(a.Proto, b.Proto), cmp.Compare(a.Source, b.Source))
	})
	return out
}

func (f *Firewall) load() (applied, error) {
	b, err := os.ReadFile(filepath.Join(f.Dir, StateFile))
	if errors.Is(err, fs.ErrNotExist) {
		return applied{}, nil
	}
	if err != nil {
		return applied{}, fmt.Errorf("firewall: read %s: %w", StateFile, err)
	}
	var st applied
	if err := json.Unmarshal(b, &st); err != nil {
		return applied{}, fmt.Errorf("firewall: %s is damaged (%w); remove it and the agent re-adds every rule", filepath.Join(f.Dir, StateFile), err)
	}
	st.Rules = normalize(st.Rules)
	return st, nil
}

func (f *Firewall) save(st applied) error {
	st.Rules = normalize(st.Rules)
	b, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return fmt.Errorf("firewall: encode state: %w", err)
	}
	return fsutil.WriteAtomic(f.Dir, StateFile, append(b, '\n'), 0o600)
}
