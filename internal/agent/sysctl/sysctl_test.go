package sysctl

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hafio/vpn-mesh-utility/internal/proto"
)

func fakeProc(t *testing.T, values map[string]string) *Writer {
	t.Helper()
	root := t.TempDir()
	for rel, v := range values {
		p := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(v+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return &Writer{Root: root}
}

func TestApplyChangesOnlyWhatDiffers(t *testing.T) {
	w := fakeProc(t, map[string]string{"net/ipv4/ip_forward": "0", "net/ipv4/conf/all/src_valid_mark": "1"})
	params := []proto.Sysctl{{Key: proto.SysctlIPForward, Value: "1"}, {Key: proto.SysctlSrcValidMark, Value: "1"}}
	n, err := w.Apply(params)
	if err != nil || n != 1 {
		t.Fatalf("first Apply = %d, %v; want 1 change", n, err)
	}
	b, _ := os.ReadFile(filepath.Join(w.Root, "net", "ipv4", "ip_forward"))
	if string(b) != "1\n" {
		t.Fatalf("ip_forward = %q", b)
	}
	if n, err := w.Apply(params); err != nil || n != 0 {
		t.Fatalf("second Apply = %d, %v; want no change", n, err)
	}
}

func TestApplyRefusesUnknownKeysBeforeTouchingAPath(t *testing.T) {
	w := fakeProc(t, map[string]string{"kernel/core_pattern": "core"})
	for _, p := range []proto.Sysctl{
		{Key: "kernel.core_pattern", Value: "1"},
		{Key: "net.ipv4.ip_forward/../../../kernel/core_pattern", Value: "1"},
		{Key: proto.SysctlIPForward, Value: "|/bin/sh"},
	} {
		if _, err := w.Apply([]proto.Sysctl{p}); err == nil {
			t.Errorf("Apply(%+v) was accepted", p)
		}
	}
	b, _ := os.ReadFile(filepath.Join(w.Root, "kernel", "core_pattern"))
	if string(b) != "core\n" {
		t.Fatalf("a refused parameter was written: %q", b)
	}
}

// The container network's keys (amendment A37) map to their own files:
// the bridge's and wg0's forwarding and proxy ARP, the gateway's proxy ARP.
func TestApplyContainerKeys(t *testing.T) {
	w := fakeProc(t, map[string]string{"net/ipv4/conf/meshctl0/forwarding": "0", "net/ipv4/conf/meshctl0/proxy_arp": "0",
		"net/ipv4/conf/wg0/forwarding": "0", "net/ipv4/conf/all/proxy_arp": "0"})
	var params []proto.Sysctl
	for _, k := range []string{proto.SysctlBridgeForwarding, proto.SysctlBridgeProxyARP, proto.SysctlWGForwarding, proto.SysctlProxyARP} {
		params = append(params, proto.Sysctl{Key: k, Value: "1"})
	}
	if n, err := w.Apply(params); err != nil || n != 4 {
		t.Fatalf("Apply = %d, %v; want 4 changes", n, err)
	}
	for _, rel := range []string{"conf/meshctl0/forwarding", "conf/meshctl0/proxy_arp", "conf/wg0/forwarding", "conf/all/proxy_arp"} {
		if b, _ := os.ReadFile(filepath.Join(w.Root, "net", "ipv4", filepath.FromSlash(rel))); string(b) != "1\n" {
			t.Errorf("%s = %q", rel, b)
		}
	}
}

// A parameter the agent may not write -- /proc/sys is read-only in a
// container -- fails with what sets it there: the compose file.
func TestApplyReportsAReadOnlyParameter(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root writes a read-only file")
	}
	w := fakeProc(t, map[string]string{"net/ipv4/conf/all/proxy_arp": "0"})
	if err := os.Chmod(filepath.Join(w.Root, "net", "ipv4", "conf", "all", "proxy_arp"), 0o444); err != nil {
		t.Fatal(err)
	}
	_, err := w.Apply([]proto.Sysctl{{Key: proto.SysctlProxyARP, Value: "1"}})
	if err == nil || !strings.Contains(err.Error(), "compose file's sysctls") {
		t.Fatalf("Apply = %v", err)
	}
}

func TestApplyReportsAMissingParameter(t *testing.T) {
	w := fakeProc(t, nil)
	if _, err := w.Apply([]proto.Sysctl{{Key: proto.SysctlIPForward, Value: "1"}}); err == nil {
		t.Fatal("Apply without /proc/sys/net/ipv4/ip_forward succeeded")
	}
}

// Without /proc/sys (Windows), an exit-mode spoke's src_valid_mark has
// nothing to do; ip_forward is refused, since only a Linux server relays.
func TestApplyWithoutProc(t *testing.T) {
	w := &Writer{NoProc: true}
	if n, err := w.Apply([]proto.Sysctl{{Key: proto.SysctlSrcValidMark, Value: "1"}}); n != 0 || err != nil {
		t.Fatalf("src_valid_mark = %d, %v", n, err)
	}
	if _, err := w.Apply([]proto.Sysctl{{Key: proto.SysctlIPForward, Value: "1"}}); err == nil {
		t.Fatal("ip_forward was accepted on a host without /proc/sys")
	}
	if _, err := w.Apply([]proto.Sysctl{{Key: "kernel.panic", Value: "1"}}); err == nil {
		t.Fatal("an unknown key was accepted")
	}
	if New().Root == "" && !New().NoProc {
		t.Fatal("New has neither /proc/sys nor NoProc")
	}
}
