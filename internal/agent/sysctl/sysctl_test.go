package sysctl

import (
	"os"
	"path/filepath"
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
