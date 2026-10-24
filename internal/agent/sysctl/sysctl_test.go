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
// container -- fails with what sets it there: the container's sysctls.
func TestApplyReportsAReadOnlyParameter(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root writes a read-only file")
	}
	w := fakeProc(t, map[string]string{"net/ipv4/conf/all/proxy_arp": "0"})
	if err := os.Chmod(filepath.Join(w.Root, "net", "ipv4", "conf", "all", "proxy_arp"), 0o444); err != nil {
		t.Fatal(err)
	}
	_, err := w.Apply([]proto.Sysctl{{Key: proto.SysctlProxyARP, Value: "1"}})
	if err == nil || !strings.Contains(err.Error(), "container's sysctls must set it") {
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

// IPv6 forwarding (M8) first moves every interface that takes router
// advertisements to accept_ra 2, so it keeps its IPv6 default route.
func TestIPv6ForwardingKeepsRouterAdvertisements(t *testing.T) {
	w := fakeProc(t, map[string]string{"net/ipv6/conf/all/forwarding": "0", "net/ipv6/conf/all/accept_ra": "1",
		"net/ipv6/conf/default/accept_ra": "1", "net/ipv6/conf/lo/accept_ra": "1", "net/ipv6/conf/eth0/accept_ra": "1",
		"net/ipv6/conf/eth1/accept_ra": "0", "net/ipv6/conf/wg0/accept_ra": "2"})
	params := []proto.Sysctl{{Key: proto.SysctlIPv6Forward, Value: "1"}}
	if n, err := w.Apply(params); err != nil || n != 2 {
		t.Fatalf("Apply = %d, %v; want eth0's accept_ra and the forwarding", n, err)
	}
	for rel, want := range map[string]string{"all/forwarding": "1", "eth0/accept_ra": "2", "eth1/accept_ra": "0", "wg0/accept_ra": "2",
		"all/accept_ra": "1", "default/accept_ra": "1", "lo/accept_ra": "1"} {
		if b, _ := os.ReadFile(filepath.Join(w.Root, "net", "ipv6", "conf", filepath.FromSlash(rel))); strings.TrimSpace(string(b)) != want {
			t.Errorf("%s = %q, want %s", rel, b, want)
		}
	}
	if n, err := w.Apply(params); err != nil || n != 0 {
		t.Fatalf("second Apply = %d, %v; want no change", n, err)
	}
	// An interface that comes later is moved at the next apply.
	if err := os.MkdirAll(filepath.Join(w.Root, "net", "ipv6", "conf", "eth2"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(w.Root, "net", "ipv6", "conf", "eth2", "accept_ra"), []byte("1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if n, err := w.Apply(params); err != nil || n != 1 {
		t.Fatalf("Apply with a new uplink = %d, %v", n, err)
	}
	// An interface the agent may not write stops the forwarding, with the
	// key to set by hand.
	if os.Geteuid() != 0 {
		ra := filepath.Join(w.Root, "net", "ipv6", "conf", "eth3", "accept_ra")
		if err := os.MkdirAll(filepath.Dir(ra), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(ra, []byte("1\n"), 0o444); err != nil {
			t.Fatal(err)
		}
		if _, err := w.Apply(params); err == nil || !strings.Contains(err.Error(), "net.ipv6.conf.eth3.accept_ra = 2") {
			t.Fatalf("Apply with an unwritable accept_ra = %v", err)
		}
	}
	// Without the IPv6 tree there is nothing to forward.
	if _, err := fakeProc(t, map[string]string{"net/ipv4/ip_forward": "0"}).Apply(params); err == nil || !strings.Contains(err.Error(), "IPv6 interfaces") {
		t.Fatalf("Apply on a host without IPv6 = %v", err)
	}
}

func TestIPv6(t *testing.T) {
	for name, c := range map[string]struct {
		values map[string]string
		want   bool
	}{
		"enabled":          {map[string]string{"net/ipv6/conf/all/disable_ipv6": "0", "net/ipv6/conf/default/disable_ipv6": "0"}, true},
		"disabled":         {map[string]string{"net/ipv6/conf/all/disable_ipv6": "1", "net/ipv6/conf/default/disable_ipv6": "1"}, false},
		"new ones without": {map[string]string{"net/ipv6/conf/all/disable_ipv6": "0", "net/ipv6/conf/default/disable_ipv6": "1"}, false},
		"no IPv6 stack":    {map[string]string{"net/ipv4/ip_forward": "0"}, false},
	} {
		if got := fakeProc(t, c.values).IPv6(); got != c.want {
			t.Errorf("%s: IPv6() = %v", name, got)
		}
	}
	if !(&Writer{NoProc: true}).IPv6() {
		t.Error("a host without /proc/sys must be taken to have IPv6")
	}
}
