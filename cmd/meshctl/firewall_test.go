package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hafio/vpn-mesh-utility/internal/server/api"
)

const firewallFile = `segmentation:
  groups:
    - name: laptops
      nodes: [laptop]
  policies:
    - name: smb
      from: [group:laptops]
      to: [node:nas]
      services: [tcp/445]
      action: allow
    - name: no-nas
      to: [node:nas]
      action: deny
    - name: allow-all
      action: allow
`

func TestFirewallCommands(t *testing.T) {
	data, _ := startAdmin(t)
	for _, args := range [][]string{
		{"node", "add", "vps-1", "--role", "server", "--endpoint", "203.0.113.10:51820"},
		{"node", "add", "laptop", "--role", "spoke", "--groups", "laptops,office"},
		{"node", "add", "nas", "--role", "spoke", "--always-on"},
	} {
		if code, out, errOut := runArgs(append(args, "--data", data)...); code != exitOK {
			t.Fatalf("%v: %d\n%s%s", args, code, out, errOut)
		}
	}
	code, out, errOut := runArgs("firewall", "show", "--data", data)
	if code != exitOK || !strings.Contains(out, "1  allow-all  allow   any   any  any") || !strings.Contains(out, "(system: ping)") ||
		!strings.Contains(out, "then drops the rest") {
		t.Fatalf("show: %d\n%s%s", code, out, errOut)
	}

	if code, out, _ = runArgs("node", "show", "laptop", "--data", data); code != exitOK || !strings.Contains(out, "laptops, office") {
		t.Fatalf("show with groups: %d\n%s", code, out)
	}

	dir := t.TempDir()
	file := filepath.Join(dir, "firewall.yaml")
	if err := os.WriteFile(file, []byte(firewallFile), 0o600); err != nil {
		t.Fatal(err)
	}
	// Nothing has joined, so no node's config changes.
	if code, out, errOut = runArgs("firewall", "apply", file, "--data", data); code != exitOK || !strings.Contains(out, "no config changed") {
		t.Fatalf("apply: %d\n%s%s", code, out, errOut)
	}
	if code, out, _ = runArgs("firewall", "show", "--yaml", "--data", data); code != exitOK || !strings.Contains(out, "- name: no-nas") ||
		!strings.Contains(out, "# meshctl mesh firewall") {
		t.Fatalf("show --yaml: %d\n%s", code, out)
	}
	var v api.SegmentView
	if code, out, _ = runArgs("firewall", "show", "--json", "--data", data); code != exitOK || json.Unmarshal([]byte(out), &v) != nil ||
		len(v.Config.Policies) != 3 {
		t.Fatalf("show --json: %d\n%s", code, out)
	}
	for _, c := range []struct {
		args []string
		want string
	}{
		{[]string{"laptop", "nas", "tcp/445"}, "10.77.2.1 -> 10.77.1.1 tcp/445: allowed (rule 2, policy: smb)"},
		{[]string{"laptop", "nas", "tcp/22"}, "dropped (rule 3, policy: no-nas)"},
		{[]string{"laptop", "1.1.1.1", "tcp/443"}, "allowed (not mesh traffic"},
	} {
		if code, out, errOut = runArgs(append(append([]string{"firewall", "check"}, c.args...), "--data", data)...); code != exitOK ||
			!strings.Contains(out, c.want) {
			t.Errorf("check %v: %d\n%s%s", c.args, code, out, errOut)
		}
	}
	if code, _, _ = runArgs("firewall", "check", "laptop", "nas", "--data", data); code != exitUsage {
		t.Errorf("check without a service = %d", code)
	}
	if code, _, errOut = runArgs("firewall", "check", "laptop", "nas", "tcp", "--data", data); code != exitFailure || !strings.Contains(errOut, "check one port") {
		t.Errorf("check every port = %d %s", code, errOut)
	}

	// Schedules: the WHEN column, the zone they are read in, a check at a
	// time inside and outside the window.
	scheduled := filepath.Join(dir, "scheduled.yaml")
	doc := strings.Replace(firewallFile, "      services: [tcp/445]\n", "      services: [tcp/445]\n      schedule:\n        days: [sat, sun]\n", 1) +
		"  timezone: Asia/Singapore\n"
	if err := os.WriteFile(scheduled, []byte(doc), 0o600); err != nil {
		t.Fatal(err)
	}
	if code, out, errOut = runArgs("firewall", "apply", scheduled, "--data", data); code != exitOK {
		t.Fatalf("apply a schedule: %d\n%s%s", code, out, errOut)
	}
	if code, out, _ = runArgs("firewall", "show", "--data", data); code != exitOK || !strings.Contains(out, "tcp/445   sat,sun") ||
		!strings.Contains(out, "schedules are read in Asia/Singapore; now ") || !strings.Contains(out, "tcp/445 sat,sun (policy: smb)") {
		t.Fatalf("show a schedule: %d\n%s", code, out)
	}
	weekend := []string{"firewall", "check", "laptop", "nas", "tcp/445", "--at", "2026-10-17T10:00:00+08:00", "--data", data}
	if code, out, errOut = runArgs(weekend...); code != exitOK || !strings.Contains(out, "allowed (rule 2, policy: smb)") || strings.Contains(out, "not in force") {
		t.Fatalf("check on a Saturday: %d\n%s%s", code, out, errOut)
	}
	weekend[6] = "2026-10-19T10:00:00+08:00" // a Monday
	if code, out, errOut = runArgs(weekend...); code != exitOK || !strings.Contains(out, "dropped (rule 3, policy: no-nas)") ||
		!strings.Contains(out, "at 2026-10-19T10:00:00+08:00, not in force (scheduled): policy: smb") {
		t.Fatalf("check on a Monday: %d\n%s%s", code, out, errOut)
	}
	weekend[6] = "monday"
	if code, _, errOut = runArgs(weekend...); code != exitUsage || !strings.Contains(errOut, "RFC 3339") {
		t.Fatalf("check at a bad time: %d %s", code, errOut)
	}

	// A bad file changes nothing.
	bad := filepath.Join(dir, "bad.yaml")
	if err := os.WriteFile(bad, []byte(strings.Replace(firewallFile, "action: deny", "action: drop", 1)), 0o600); err != nil {
		t.Fatal(err)
	}
	if code, _, errOut = runArgs("firewall", "apply", bad, "--data", data); code != exitFailure || !strings.Contains(errOut, "nothing was changed") {
		t.Fatalf("apply a bad file: %d %s", code, errOut)
	}
	if code, _, errOut = runArgs("firewall", "apply", filepath.Join(dir, "missing.yaml"), "--data", data); code != exitFailure || !strings.Contains(errOut, "read") {
		t.Fatalf("apply a missing file: %d %s", code, errOut)
	}
	dangling := filepath.Join(dir, "dangling.yaml")
	if err := os.WriteFile(dangling, []byte(strings.Replace(firewallFile, "node:nas", "node:ghost", 1)), 0o600); err != nil {
		t.Fatal(err)
	}
	if code, _, errOut = runArgs("firewall", "apply", dangling, "--data", data); code != exitFailure || !strings.Contains(errOut, `there is no node "ghost"`) {
		t.Fatalf("apply a dangling reference: %d %s", code, errOut)
	}

	// apply replaced the whole document, groups included: office is gone.
	if code, out, _ = runArgs("node", "show", "laptop", "--data", data); code != exitOK || !strings.Contains(out, "groups     laptops\n") {
		t.Fatalf("show after apply: %d\n%s", code, out)
	}
	if code, out, errOut = runArgs("node", "set", "laptop", "--groups", "laptops", "--data", data, "--json"); code != exitOK ||
		!strings.Contains(out, `"groups": [`) || strings.Contains(out, "office") {
		t.Fatalf("set --groups: %d\n%s%s", code, out, errOut)
	}
	if code, _, errOut = runArgs("node", "set", "laptop", "--groups", "Laptops", "--data", data); code != exitFailure || !strings.Contains(errOut, "groups[0]") {
		t.Fatalf("set a bad group: %d %s", code, errOut)
	}
}
