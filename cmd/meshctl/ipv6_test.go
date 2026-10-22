package main

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/hafio/vpn-mesh-utility/internal/server/api"
	"github.com/hafio/vpn-mesh-utility/internal/testfix"
)

func TestIPv6Command(t *testing.T) {
	data, _ := startAdmin(t)
	if code, out, errOut := runArgs("node", "add", "vps-1", "--role", "server", "--endpoint", testfix.VPS1.Endpoint, "--data", data); code != exitOK {
		t.Fatalf("add vps-1: %d %s%s", code, out, errOut)
	}
	if code, out, _ := runArgs("ipv6", "--data", data); code != exitOK || !strings.Contains(out, "runs IPv4 only") {
		t.Fatalf("ipv6 without a prefix: %d\n%s", code, out)
	}
	if code, _, errOut := runArgs("node", "add", "pad", "--role", "spoke", "--family", "ipv6", "--data", data); code != exitFailure ||
		!strings.Contains(errOut, "meshctl ipv6 auto") {
		t.Fatalf("an ipv6 spoke before a prefix: %d %s", code, errOut)
	}
	code, out, _ := runArgs("ipv6", testfix.MeshCIDR6, "--data", data)
	if code != exitOK || !strings.Contains(out, "IPv6 prefix "+testfix.MeshCIDR6+"; 1 node(s)") {
		t.Fatalf("ipv6 <prefix>: %d\n%s", code, out)
	}
	code, out, _ = runArgs("ipv6", "--data", data, "--json")
	var v api.IPv6View
	if code != exitOK || json.Unmarshal([]byte(out), &v) != nil || v.Prefix != testfix.MeshCIDR6 || v.Addressed != 1 {
		t.Fatalf("ipv6 --json: %d %s", code, out)
	}
	if code, _, errOut := runArgs("ipv6", "fd20::/64", "--data", data); code != exitFailure || !strings.Contains(errOut, "give IPv6 up first") {
		t.Fatalf("ipv6 swap: %d %s", code, errOut)
	}
	if code, _, errOut := runArgs("ipv6", "fd20::/64", "off", "--data", data); code != exitUsage || !strings.Contains(errOut, "at most one") {
		t.Fatalf("ipv6 with two arguments: %d %s", code, errOut)
	}

	// The node flags of the IPv6 side.
	code, out, errOut := runArgs("node", "add", "pad", "--role", "spoke", "--family", "ipv6", "--data", data)
	if code != exitOK || !strings.Contains(out, "added pad (spoke, mesh IPv6 fd10:77::2:1)") {
		t.Fatalf("add --family ipv6: %d\n%s%s", code, out, errOut)
	}
	code, out, _ = runArgs("node", "add", "nas", "--role", "spoke", "--always-on", "--ip6", "fd10:77::1:9", "--containers", "--containers6", "--data", data)
	if code != exitOK || !strings.Contains(out, "added nas (spoke, mesh IP 10.77.1.1 and fd10:77::1:9)") ||
		!strings.Contains(out, "its containers: 10.77.128.0/24, "+testfix.ContainerSubnet6) {
		t.Fatalf("add --ip6 --containers6: %d\n%s", code, out)
	}
	if code, out, _ = runArgs("node", "add", "vps-2", "--role", "server", "--endpoint6", "[2001:db8::11]:51820", "--data", data); code != exitOK {
		t.Fatalf("add --endpoint6: %d\n%s", code, out)
	}
	if code, out, _ = runArgs("node", "list", "--data", data); code != exitOK || !strings.Contains(out, "fd10:77::2:1") || !strings.Contains(out, "[2001:db8::11]:51820") {
		t.Fatalf("list shows the IPv6-only node and the IPv6 endpoint: %d\n%s", code, out)
	}
	if code, out, _ = runArgs("node", "show", "pad", "--data", data); code != exitOK || !strings.Contains(out, "fd10:77::2:1") || !strings.Contains(out, "ipv6 only") {
		t.Fatalf("show pad: %d\n%s", code, out)
	}
	if code, out, _ = runArgs("node", "show", "vps-2", "--data", data); code != exitOK || !strings.Contains(out, "[2001:db8::11]:51820") {
		t.Fatalf("show vps-2: %d\n%s", code, out)
	}
	if code, out, _ = runArgs("node", "set", "pad", "--family", "dual", "--data", data, "--json"); code != exitOK || !strings.Contains(out, `"mesh_ip": "10.77.2.1"`) {
		t.Fatalf("set --family dual: %d\n%s", code, out)
	}
	if code, out, _ = runArgs("node", "set", "pad", "--mesh-ip6", "fd10:77::2:50", "--data", data); code != exitOK ||
		!strings.Contains(out, "pad moves from fd10:77::2:1 to fd10:77::2:50") {
		t.Fatalf("set --mesh-ip6: %d\n%s", code, out)
	}
	if code, out, _ = runArgs("node", "set", "vps-2", "--endpoint6", "none", "--data", data, "--json"); code != exitOK || strings.Contains(out, `"endpoint6":`) {
		t.Fatalf("set --endpoint6 none: %d\n%s", code, out)
	}
	if code, out, _ = runArgs("node", "set", "nas", "--containers6=false", "--data", data, "--json"); code != exitOK || strings.Contains(out, `"container_subnet6":`) {
		t.Fatalf("set --containers6=false: %d\n%s", code, out)
	}
	if code, _, errOut = runArgs("node", "set", "nas", "--family", "both", "--data", data); code != exitFailure || !strings.Contains(errOut, "ip_family") {
		t.Fatalf("set --family both: %d %s", code, errOut)
	}

	if code, _, errOut = runArgs("ipv6", "off", "--data", data); code != exitOK {
		t.Fatalf("ipv6 off: %d %s", code, errOut)
	}
	if code, out, _ = runArgs("ipv6", "--data", data); code != exitOK || !strings.Contains(out, "runs IPv4 only") {
		t.Fatalf("ipv6 after off: %d\n%s", code, out)
	}
}
