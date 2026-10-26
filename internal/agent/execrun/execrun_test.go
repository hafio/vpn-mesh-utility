package execrun

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// fakeProgram puts an executable shell script named name first and alone on
// PATH, so Exec runs it instead of the real tool.
func fakeProgram(t *testing.T, name, body string) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("the fake programs are shell scripts")
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, name), []byte("#!/bin/sh\n"+body+"\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir)
}

func TestCheckAllowsFirewallArgv(t *testing.T) {
	for _, argv := range [][]string{
		{"firewall-cmd", "--permanent", "--zone=mesh", "--add-port=51820/udp"},
		{"firewall-cmd", "--zone=mesh", `--add-rich-rule=rule family="ipv4" source address="10.77.0.2/32" port port="2377" protocol="tcp" accept`},
	} {
		if err := Check(argv); err != nil {
			t.Errorf("Check(%q) = %v", argv, err)
		}
	}
}

func TestCheckRefuses(t *testing.T) {
	cases := map[string][]string{
		"empty":           nil,
		"unlisted":        {"sh", "-c", "id"},
		"systemctl":       {"systemctl", "enable", "meshctl-agent@default.service"},
		"path to program": {"/usr/bin/firewall-cmd", "--state"},
		"semicolon":       {"firewall-cmd", "--zone=mesh; rm -rf /"},
		"substitution":    {"firewall-cmd", "--zone=$(id)"},
		"newline":         {"firewall-cmd", "--zone=mesh\n--reload"},
		"single quote":    {"firewall-cmd", "--zone='mesh'"},
		"empty argument":  {"firewall-cmd", ""},
	}
	for name, argv := range cases {
		if err := Check(argv); err == nil {
			t.Errorf("%s: Check(%q) accepted", name, argv)
		}
	}
}

// A refused argv must never reach exec: Run fails before the PATH lookup,
// so the error is the gate's, not "not installed".
func TestRunRefusesBeforeExecuting(t *testing.T) {
	_, err := Exec{}.Run(context.Background(), []string{"firewall-cmd", "--zone=`id`"})
	if err == nil || errors.Is(err, ErrNotInstalled) || !strings.Contains(err.Error(), "refusing to run") {
		t.Fatalf("Run = %v, want the gate's refusal", err)
	}
}

func TestCappedKeepsTheHead(t *testing.T) {
	var c capped
	big := strings.Repeat("x", maxOutput+10)
	if n, err := c.Write([]byte(big)); n != len(big) || err != nil {
		t.Fatalf("Write = %d, %v", n, err)
	}
	c.Write([]byte("more"))
	if len(c.String()) != maxOutput {
		t.Fatalf("kept %d bytes, want %d", len(c.String()), maxOutput)
	}
}

func TestRunCapturesOutputWithAMinimalEnvironment(t *testing.T) {
	fakeProgram(t, "firewall-cmd", `echo "running $1 LC_ALL=$LC_ALL HOME=${HOME:-unset}"`)
	t.Setenv("HOME", "/root")
	out, err := Exec{}.Run(context.Background(), []string{"firewall-cmd", "--state"})
	if err != nil {
		t.Fatal(err)
	}
	if out != "running --state LC_ALL=C HOME=unset\n" {
		t.Fatalf("stdout = %q; the child must get only PATH and LC_ALL", out)
	}
}

func TestRunFoldsStderrIntoTheError(t *testing.T) {
	fakeProgram(t, "firewall-cmd", `echo "Error: INVALID_ZONE: mesh9" >&2; exit 5`)
	_, err := Exec{}.Run(context.Background(), []string{"firewall-cmd", "--zone=mesh9", "--list-ports"})
	if err == nil || !strings.Contains(err.Error(), "firewall-cmd --zone=mesh9 --list-ports") ||
		!strings.Contains(err.Error(), "exit status 5") || !strings.Contains(err.Error(), "INVALID_ZONE") {
		t.Fatalf("Run = %v; want the argv, the exit status and stderr", err)
	}
}

func TestRunTimesOut(t *testing.T) {
	fakeProgram(t, "firewall-cmd", "exec sleep 10")
	start := time.Now()
	_, err := Exec{Timeout: 100 * time.Millisecond}.Run(context.Background(), []string{"firewall-cmd", "--reload"})
	if err == nil || time.Since(start) > 5*time.Second {
		t.Fatalf("Run = %v after %s; want a timeout", err, time.Since(start))
	}
}

func TestRunReportsAMissingProgram(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	if _, err := (Exec{}).Run(context.Background(), []string{"firewall-cmd", "--state"}); !errors.Is(err, ErrNotInstalled) {
		t.Fatalf("Run = %v, want ErrNotInstalled", err)
	}
}
