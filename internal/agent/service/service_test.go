package service

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hafio/vpn-mesh-utility/deploy"
	"github.com/hafio/vpn-mesh-utility/internal/agent/execrun"
)

type recorder struct{ calls [][]string }

func (r *recorder) Run(_ context.Context, argv []string) (string, error) {
	if err := execrun.Check(argv); err != nil {
		return "", err
	}
	r.calls = append(r.calls, argv)
	return "", nil
}

func host(t *testing.T, exe string, dirs ...string) (Installer, *recorder) {
	t.Helper()
	root := t.TempDir()
	for _, d := range append(dirs, "/etc/systemd/system") {
		if err := os.MkdirAll(filepath.Join(root, d), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	r := &recorder{}
	return Installer{Run: r, Root: root, Executable: func() (string, error) { return exe, nil }}, r
}

func TestInstallWritesTheUnitAndStartsIt(t *testing.T) {
	i, r := host(t, deploy.BinaryPath, "/run/systemd/system")
	if err := i.Install(context.Background()); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(i.path(deploy.UnitPath))
	if err != nil || string(b) != string(deploy.SystemdUnit()) {
		t.Fatalf("unit = %q, %v", b, err)
	}
	want := "systemctl daemon-reload|systemctl enable meshctl-agent.service|systemctl restart meshctl-agent.service"
	var got []string
	for _, c := range r.calls {
		got = append(got, strings.Join(c, " "))
	}
	if strings.Join(got, "|") != want {
		t.Fatalf("calls = %q", got)
	}
}

func TestInstallRefusesABinaryElsewhere(t *testing.T) {
	i, r := host(t, "/home/me/meshctl", "/run/systemd/system")
	err := i.Install(context.Background())
	if err == nil || !strings.Contains(err.Error(), "--no-service") {
		t.Fatalf("Install = %v", err)
	}
	if len(r.calls) != 0 || exists(i.path(deploy.UnitPath)) {
		t.Fatal("a refused install must not write the unit or run systemctl")
	}
}

func TestAvailable(t *testing.T) {
	bare, _ := host(t, deploy.BinaryPath)
	if ok, _ := bare.Available(); ok {
		t.Error("no /run/systemd/system: must not be available")
	}
	i, _ := host(t, deploy.BinaryPath, "/run/systemd/system")
	if ok, why := i.Available(); !ok && !strings.Contains(why, "Linux-only") {
		t.Errorf("systemd host: Available = false (%s)", why)
	}
	c, _ := host(t, deploy.BinaryPath, "/run/systemd/system")
	if err := os.WriteFile(c.path("/.dockerenv"), nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if ok, why := c.Available(); ok || (!strings.Contains(why, "container") && !strings.Contains(why, "Linux-only")) {
		t.Errorf("container: Available = %v (%s)", ok, why)
	}
}
