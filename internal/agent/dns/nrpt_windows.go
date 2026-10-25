//go:build windows

package dns

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"

	"github.com/hafio/vpn-mesh-utility/internal/proto"
)

// nrptBase holds the local NRPT rules the DNS Client reads.
const nrptBase = `SYSTEM\CurrentControlSet\Services\Dnscache\Parameters\DnsPolicyConfig`

// nrptGenericServers is ConfigOptions' flag for "send to GenericDNSServers".
const nrptGenericServers = 0x8

// nrpt registers the resolver through NRPT rules under slot's keys; the
// interface name only labels errors, since the rules are the host's, not
// the adapter's.
type nrpt struct{ slot proto.Slot }

// NewRegistrar returns the host's registrar for the agent of slot s: NRPT.
func NewRegistrar(s proto.Slot) Registrar { return nrpt{slot: s} }

var flushCache = windows.NewLazySystemDLL("dnsapi.dll").NewProc("DnsFlushResolverCache")

// flush drops the DNS Client's cache, so the new rules apply to names
// looked up before. It is best effort: the rules apply to new lookups
// either way.
func flush() { _, _, _ = flushCache.Call() }

// ownedKeys lists slot s's rules under nrptBase.
func ownedKeys(s proto.Slot) ([]string, error) {
	base, err := registry.OpenKey(registry.LOCAL_MACHINE, nrptBase, registry.ENUMERATE_SUB_KEYS)
	if errors.Is(err, registry.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer base.Close() // read-only
	names, err := base.ReadSubKeyNames(-1)
	if err != nil {
		return nil, err
	}
	var out []string
	for _, n := range names {
		if ownedKey(n, s) {
			out = append(out, n)
		}
	}
	return out, nil
}

func (n nrpt) Get(_ context.Context, iface string) (Registration, error) {
	keys, err := ownedKeys(n.slot)
	if err != nil {
		return Registration{}, fmt.Errorf("list the NRPT rules for %s: %w", iface, err)
	}
	var rules []nrptRule
	for _, key := range keys {
		k, err := registry.OpenKey(registry.LOCAL_MACHINE, nrptBase+`\`+key, registry.QUERY_VALUE)
		if errors.Is(err, registry.ErrNotExist) {
			continue
		}
		if err != nil {
			return Registration{}, fmt.Errorf("read NRPT rule %s for %s: %w", key, iface, err)
		}
		names, _, nerr := k.GetStringsValue("Name")
		servers, _, serr := k.GetStringValue("GenericDNSServers")
		_ = k.Close() // read-only
		if nerr != nil || serr != nil || len(names) != 1 {
			continue // not the shape the agent writes: it is replaced on Set
		}
		rules = append(rules, nrptRule{Key: key, Name: names[0], Servers: parseServers(servers)})
	}
	return fromRules(rules), nil
}

func (n nrpt) Set(_ context.Context, iface string, r Registration) error {
	want := rulesFor(r, n.slot)
	keep := map[string]bool{}
	for _, rule := range want {
		keep[strings.ToLower(rule.Key)] = true
		k, _, err := registry.CreateKey(registry.LOCAL_MACHINE, nrptBase+`\`+rule.Key, registry.SET_VALUE)
		if err != nil {
			return fmt.Errorf("write NRPT rule %s for %s: %w (run from an elevated prompt or the service)", rule.Name, iface, err)
		}
		err = errors.Join(
			k.SetDWordValue("Version", 2),
			k.SetStringsValue("Name", []string{rule.Name}),
			k.SetStringValue("GenericDNSServers", serverList(rule.Servers)),
			k.SetDWordValue("ConfigOptions", nrptGenericServers),
			k.SetStringValue("IPSECCARestriction", ""),
		)
		_ = k.Close() // the values are written or err says why
		if err != nil {
			return fmt.Errorf("write NRPT rule %s for %s: %w", rule.Name, iface, err)
		}
	}
	keys, err := ownedKeys(n.slot)
	if err != nil {
		return fmt.Errorf("list the NRPT rules for %s: %w", iface, err)
	}
	for _, key := range keys {
		if !keep[strings.ToLower(key)] {
			if err := deleteRule(key); err != nil {
				return fmt.Errorf("remove NRPT rule %s for %s: %w", key, iface, err)
			}
		}
	}
	flush()
	return nil
}

func (n nrpt) Revert(_ context.Context, iface string) error {
	keys, err := ownedKeys(n.slot)
	if err != nil {
		return fmt.Errorf("list the NRPT rules for %s: %w", iface, err)
	}
	for _, key := range keys {
		if err := deleteRule(key); err != nil {
			return fmt.Errorf("remove NRPT rule %s for %s: %w", key, iface, err)
		}
	}
	flush()
	return nil
}

func deleteRule(key string) error {
	err := registry.DeleteKey(registry.LOCAL_MACHINE, nrptBase+`\`+key)
	if errors.Is(err, registry.ErrNotExist) {
		return nil
	}
	return err
}
