//go:build windows

package dns

import (
	"context"
	"errors"
	"fmt"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

// nrptBase holds the local NRPT rules the DNS Client reads.
const nrptBase = `SYSTEM\CurrentControlSet\Services\Dnscache\Parameters\DnsPolicyConfig`

// nrptGenericServers is ConfigOptions' flag for "send to GenericDNSServers".
const nrptGenericServers = 0x8

// nrptKeys are every key the agent writes.
var nrptKeys = []string{nrptMeshKey, nrptRootKey}

// nrpt registers the resolver through NRPT rules; the interface name only
// labels errors, since the rules are the host's, not the adapter's.
type nrpt struct{}

// NewRegistrar returns the host's registrar: NRPT.
func NewRegistrar() Registrar { return nrpt{} }

var flushCache = windows.NewLazySystemDLL("dnsapi.dll").NewProc("DnsFlushResolverCache")

// flush drops the DNS Client's cache, so the new rules apply to names
// looked up before. It is best effort: the rules apply to new lookups
// either way.
func flush() { _, _, _ = flushCache.Call() }

func (nrpt) Get(_ context.Context, iface string) (Registration, error) {
	var rules []nrptRule
	for _, key := range nrptKeys {
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

func (nrpt) Set(_ context.Context, iface string, r Registration) error {
	want := rulesFor(r)
	keep := map[string]bool{}
	for _, rule := range want {
		keep[rule.Key] = true
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
	for _, key := range nrptKeys {
		if !keep[key] {
			if err := deleteRule(key); err != nil {
				return fmt.Errorf("remove NRPT rule %s for %s: %w", key, iface, err)
			}
		}
	}
	flush()
	return nil
}

func (nrpt) Revert(_ context.Context, iface string) error {
	for _, key := range nrptKeys {
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
