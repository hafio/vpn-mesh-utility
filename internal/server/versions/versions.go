// Package versions explains config versions to an operator: Compare lists
// what changed between two snapshots of the desired state, by node name and
// field, for the versions page and for the confirmation a rollback that adds
// or removes nodes needs (spec S8.3).
package versions

import (
	"fmt"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/hafio/vpn-mesh-utility/internal/server/store"
)

// FieldChange is one changed value.
type FieldChange struct {
	Field string `json:"field"`
	From  string `json:"from"`
	To    string `json:"to"`
}

// NodeDiff lists the changed fields of a node present in both snapshots.
type NodeDiff struct {
	Name   string        `json:"name"`
	Fields []FieldChange `json:"fields"`
}

// Diff is what turns one snapshot into another.
type Diff struct {
	AddedNodes   []string      `json:"added_nodes,omitempty"`
	RemovedNodes []string      `json:"removed_nodes,omitempty"`
	ChangedNodes []NodeDiff    `json:"changed_nodes,omitempty"`
	Settings     []FieldChange `json:"settings,omitempty"`
	// DNS and Firewall list added ("+ ...") and removed ("- ...") entries;
	// Firewall is the segmentation (amendment A43), one entry per group,
	// object, zone and policy, a policy with its place in the order.
	DNS      []string `json:"dns,omitempty"`
	Firewall []string `json:"firewall,omitempty"`
}

// Empty reports that the snapshots describe the same desired state.
func (d Diff) Empty() bool {
	return len(d.AddedNodes)+len(d.RemovedNodes)+len(d.ChangedNodes)+len(d.Settings)+len(d.DNS)+len(d.Firewall) == 0
}

// ChangesMembership reports that nodes appear or disappear: a rollback
// across it needs the operator's confirmation.
func (d Diff) ChangesMembership() bool { return len(d.AddedNodes)+len(d.RemovedNodes) > 0 }

// fingerprint shortens a public key for display.
func fingerprint(k string) string {
	if len(k) <= 8 {
		return k
	}
	return k[:8] + "..."
}

// fields is the display order of a node's desired fields.
func fields(n store.SnapNode, names map[int64]string) [][2]string {
	return [][2]string{
		{"name", n.Name},
		{"role", n.Role},
		{"mesh_ip", n.MeshIP},
		{"mesh_ip6", n.MeshIP6},
		{"ip_family", n.Family},
		{"endpoint", n.Endpoint},
		{"endpoint6", n.Endpoint6},
		{"listen_port", strconv.Itoa(n.ListenPort)},
		{"hub", names[n.PrimaryHubID]},
		{"can_exit", strconv.FormatBool(n.CanExit)},
		{"exit", names[n.ExitVia]},
		{"keepalive", strconv.Itoa(n.Keepalive)},
		{"mtu", strconv.Itoa(n.MTU)},
		{"enrolled", strconv.FormatBool(n.Enrolled)},
		{"disabled", strconv.FormatBool(n.Disabled)},
		{"wg_public_key", fingerprint(n.WGPubKey)},
		{"identity_public_key", fingerprint(n.IdentityPubKey)},
		{"container_subnet6", n.ContainerSubnet6},
	}
}

// Compare returns the changes from one snapshot to the other. Nodes are
// matched by id, so a renamed node is a change, not a removal and an
// addition.
func Compare(from, to store.Snapshot) Diff {
	var d Diff
	fromNames, toNames := names(from), names(to)
	fromByID := map[int64]store.SnapNode{}
	for _, n := range from.Nodes {
		fromByID[n.ID] = n
	}
	seen := map[int64]bool{}
	for _, n := range to.Nodes {
		seen[n.ID] = true
		old, ok := fromByID[n.ID]
		if !ok {
			d.AddedNodes = append(d.AddedNodes, n.Name)
			continue
		}
		var changed []FieldChange
		a, b := fields(old, fromNames), fields(n, toNames)
		for i := range a {
			if a[i][1] != b[i][1] {
				changed = append(changed, FieldChange{Field: a[i][0], From: a[i][1], To: b[i][1]})
			}
		}
		if len(changed) > 0 {
			d.ChangedNodes = append(d.ChangedNodes, NodeDiff{Name: n.Name, Fields: changed})
		}
	}
	for _, n := range from.Nodes {
		if !seen[n.ID] {
			d.RemovedNodes = append(d.RemovedNodes, n.Name)
		}
	}
	sort.Strings(d.AddedNodes)
	sort.Strings(d.RemovedNodes)
	slices.SortFunc(d.ChangedNodes, func(a, b NodeDiff) int { return strings.Compare(a.Name, b.Name) })

	keys := map[string]bool{}
	for k := range from.Settings {
		keys[k] = true
	}
	for k := range to.Settings {
		keys[k] = true
	}
	for _, k := range sortedKeys(keys) {
		if from.Settings[k] != to.Settings[k] {
			d.Settings = append(d.Settings, FieldChange{Field: k, From: from.Settings[k], To: to.Settings[k]})
		}
	}
	d.DNS = setDiff(describeDNS(from, fromNames), describeDNS(to, toNames))
	d.Firewall = setDiff(describeSegmentation(from.EffectiveSegmentation()), describeSegmentation(to.EffectiveSegmentation()))
	return d
}

func names(s store.Snapshot) map[int64]string {
	out := map[int64]string{}
	for _, n := range s.Nodes {
		out[n.ID] = n.Name
	}
	return out
}

func sortedKeys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func describeDNS(s store.Snapshot, names map[int64]string) []string {
	var out []string
	for _, r := range s.DNS {
		target := r.IP
		if r.IP6 != "" {
			target = strings.TrimPrefix(target+", "+r.IP6, ", ")
		}
		if r.NodeID != 0 {
			target = "node " + names[r.NodeID]
		}
		out = append(out, fmt.Sprintf("%s -> %s", r.Name, target))
	}
	return out
}

func describeSegmentation(s store.Segmentation) []string {
	list := func(vs []string) string {
		if len(vs) == 0 {
			return "any"
		}
		return strings.Join(vs, ",")
	}
	var out []string
	for _, g := range s.Groups {
		out = append(out, fmt.Sprintf("group %s: %s", g.Name, strings.Join(g.Nodes, ",")))
	}
	for _, o := range s.Addresses {
		out = append(out, fmt.Sprintf("address %s: %s", o.Name, strings.Join(o.Entries, ",")))
	}
	for _, o := range s.Services {
		out = append(out, fmt.Sprintf("service %s: %s", o.Name, strings.Join(o.Entries, ",")))
	}
	for _, z := range s.Zones {
		out = append(out, fmt.Sprintf("zone %s (within: %s): %s", z.Name, z.Intra, strings.Join(z.Members, ",")))
	}
	for i, p := range s.Policies {
		line := fmt.Sprintf("policy %d %s: %s %s -> %s %s", i+1, p.Name, p.Action, list(p.From), list(p.To), list(p.Services))
		if p.Schedule != nil {
			line += " " + p.Schedule.String()
		}
		if p.Disabled {
			line += " (disabled)"
		}
		out = append(out, line)
	}
	if s.TimeZone != "" {
		out = append(out, "schedules in "+s.TimeZone)
	}
	return out
}

// setDiff lists what to has that from lacks ("+ ") and the reverse ("- "),
// sorted.
func setDiff(from, to []string) []string {
	var out []string
	for _, s := range to {
		if !slices.Contains(from, s) {
			out = append(out, "+ "+s)
		}
	}
	for _, s := range from {
		if !slices.Contains(to, s) {
			out = append(out, "- "+s)
		}
	}
	sort.Strings(out)
	return out
}
