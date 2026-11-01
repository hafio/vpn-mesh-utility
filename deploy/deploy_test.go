package deploy

import (
	"bufio"
	"bytes"
	"flag"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"slices"
	"strings"
	"testing"

	"go.yaml.in/yaml/v3"
)

var update = flag.Bool("update", false, "rewrite testdata goldens")

// The unit's [Service] section, exactly. Adding, removing or widening any
// setting fails here first; docs/security.md explains each one. There is no
// User= (the agent is root, bounded by the capability set) and no
// ProtectKernelTunables (it writes net.ipv4.ip_forward).
var wantService = map[string]string{
	"Type":                     "simple",
	"ExecStart":                "/usr/local/bin/meshctl agent run --mesh %i",
	"Restart":                  "always",
	"RestartSec":               "5",
	"RestartPreventExitStatus": "3",
	"StateDirectory":           "meshctl/%i",
	"StateDirectoryMode":       "0700",
	"RuntimeDirectory":         "meshctl wireguard",
	"RuntimeDirectoryMode":     "0755",
	"RuntimeDirectoryPreserve": "yes",
	"UMask":                    "0077",
	"CapabilityBoundingSet":    "CAP_NET_ADMIN CAP_NET_BIND_SERVICE",
	"AmbientCapabilities":      "CAP_NET_ADMIN CAP_NET_BIND_SERVICE",
	"NoNewPrivileges":          "yes",
	"ProtectSystem":            "strict",
	"ReadWritePaths":           "/var/lib/meshctl",
	"ProtectHome":              "yes",
	"PrivateTmp":               "yes",
	"ProtectKernelModules":     "yes",
	"ProtectKernelLogs":        "yes",
	"ProtectControlGroups":     "yes",
	"ProtectClock":             "yes",
	"ProtectHostname":          "yes",
	"RestrictNamespaces":       "yes",
	"RestrictRealtime":         "yes",
	"RestrictSUIDSGID":         "yes",
	"LockPersonality":          "yes",
	"MemoryDenyWriteExecute":   "yes",
	"SystemCallArchitectures":  "native",
	"DevicePolicy":             "closed",
	"DeviceAllow":              "/dev/net/tun rw",
	"RestrictAddressFamilies":  "AF_UNIX AF_INET AF_INET6 AF_NETLINK",
}

func TestUnitPosture(t *testing.T) {
	got := map[string]string{}
	section := ""
	sc := bufio.NewScanner(bytes.NewReader(SystemdUnit()))
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		switch {
		case line == "" || strings.HasPrefix(line, "#"):
		case strings.HasPrefix(line, "["):
			section = line
		case section == "[Service]":
			k, v, ok := strings.Cut(line, "=")
			if !ok {
				t.Fatalf("malformed line %q", line)
			}
			if _, dup := got[k]; dup {
				t.Fatalf("%s is set twice", k)
			}
			got[k] = v
		}
	}
	if !reflect.DeepEqual(got, wantService) {
		for k, v := range got {
			if wantService[k] != v {
				t.Errorf("%s=%s, want %q", k, v, wantService[k])
			}
		}
		for k := range wantService {
			if _, ok := got[k]; !ok {
				t.Errorf("%s is missing", k)
			}
		}
	}
	if !strings.Contains(string(SystemdUnit()), "WantedBy=multi-user.target") {
		t.Error("the unit must be enabled for multi-user.target")
	}
	// install.sh writes it through a here-document that this line ends.
	if strings.Contains("\n"+string(SystemdUnit()), "\nMESHCTL_UNIT\n") || !strings.HasSuffix(string(SystemdUnit()), "\n") {
		t.Error("the unit would end install.sh's here-document early, or does not end in a newline")
	}
}

func testParams() InstallParams {
	return InstallParams{Origin: "https://mesh.example.com", Version: "v1.2.3",
		SHA256Amd64: strings.Repeat("a", 64), SHA256Arm64: strings.Repeat("b", 64), SHA256Windows: strings.Repeat("c", 64)}
}

// shellScripts are the Linux scripts (amendment A55): each a template that
// takes the shared part, lib.sh.tmpl, and its renderer.
var shellScripts = []struct {
	name   string
	tmpl   string
	render func(InstallParams) ([]byte, error)
}{
	{"install.sh", installTemplate, RenderInstall},
	{"reinstall.sh", reinstallTemplate, RenderReinstall},
	{"uninstall.sh", uninstallTemplate, RenderUninstall},
}

// ps1Scripts are the Windows scripts (amendments A51, A55), each taking the
// shared part lib.ps1.tmpl the same way.
var ps1Scripts = []struct {
	name   string
	tmpl   string
	render func(InstallParams) ([]byte, error)
}{
	{"install.ps1", installPS1Template, RenderInstallPS1},
	{"reinstall.ps1", reinstallPS1Template, RenderReinstallPS1},
	{"uninstall.ps1", uninstallPS1Template, RenderUninstallPS1},
}

func TestInstallGolden(t *testing.T) {
	renders := map[string]func(InstallParams) ([]byte, error){}
	for _, s := range shellScripts {
		renders[strings.TrimSuffix(s.name, ".sh")+".golden.sh"] = s.render
	}
	for _, s := range ps1Scripts {
		renders[strings.TrimSuffix(s.name, ".ps1")+".golden.ps1"] = s.render
	}
	for name, render := range renders {
		got, err := render(testParams())
		if err != nil {
			t.Fatal(err)
		}
		path := filepath.Join("testdata", name)
		if *update {
			if err := os.WriteFile(path, got, 0o644); err != nil {
				t.Fatal(err)
			}
		}
		want, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(got, want) {
			t.Errorf("%s differs from %s; review the template change, then ./scripts/dev.sh regen", name, path)
		}
	}
}

// A rendered script differs from its template, with the shared part in place
// of its marker, only at the value markers, each of which appears exactly
// once in the shared part, inside single quotes.
func TestInstallChangesOnlyTheMarkers(t *testing.T) {
	p := testParams()
	for _, m := range []string{markerOrigin, markerVersion, markerSHA256Amd64, markerSHA256Arm64} {
		if n := strings.Count(libTemplate, "'"+m+"'"); n != 1 || strings.Count(libTemplate, m) != 1 {
			t.Errorf("marker %s must appear exactly once in lib.sh.tmpl, quoted", m)
		}
	}
	if strings.Count(libTemplate, "\n"+markerUnit+"MESHCTL_UNIT\n") != 1 || strings.Count(libTemplate, markerUnit) != 1 {
		t.Error("the unit's marker must be the one line of a here-document ended by MESHCTL_UNIT")
	}
	back := strings.NewReplacer("'"+p.Origin+"'", "'"+markerOrigin+"'", "'"+p.Version+"'", "'"+markerVersion+"'",
		"'"+p.SHA256Amd64+"'", "'"+markerSHA256Amd64+"'", "'"+p.SHA256Arm64+"'", "'"+markerSHA256Arm64+"'", systemdUnit, markerUnit)
	for _, s := range shellScripts {
		if strings.Count(s.tmpl, "\n"+markerLib) != 1 || strings.Count(s.tmpl, "@@") != 2 || strings.Contains(libTemplate, "@@MESHCTL_LIB@@") {
			t.Errorf("%s: the shared part's marker must be its only marker, on a line of its own", s.name)
		}
		got, err := s.render(p)
		if err != nil {
			t.Fatal(err)
		}
		if back.Replace(string(got)) != withLib(s.tmpl, libTemplate) {
			t.Errorf("rendering %s changed more than the markers", s.name)
		}
	}

	// The Windows scripts (amendments A51, A55), the same way.
	for _, m := range []string{markerOrigin, markerVersion, markerSHA256Windows} {
		if n := strings.Count(libPS1Template, "'"+m+"'"); n != 1 || strings.Count(libPS1Template, m) != 1 {
			t.Errorf("marker %s must appear exactly once in lib.ps1.tmpl, quoted", m)
		}
	}
	ps1Back := strings.NewReplacer("'"+p.Origin+"'", "'"+markerOrigin+"'", "'"+p.Version+"'", "'"+markerVersion+"'",
		"'"+p.SHA256Windows+"'", "'"+markerSHA256Windows+"'")
	for _, s := range ps1Scripts {
		if strings.Count(s.tmpl, "\n"+markerLib) != 1 || strings.Count(s.tmpl, "@@") != 2 || strings.Contains(libPS1Template, "@@MESHCTL_LIB@@") {
			t.Errorf("%s: the shared part's marker must be its only marker, on a line of its own", s.name)
		}
		got, err := s.render(p)
		if err != nil {
			t.Fatal(err)
		}
		if ps1Back.Replace(string(got)) != withLib(s.tmpl, libPS1Template) {
			t.Errorf("rendering %s changed more than the markers", s.name)
		}
		for _, b := range got {
			if b > 0x7e || b < 0x20 && b != '\n' && b != '\r' {
				t.Fatalf("%s has the byte %#x: Windows PowerShell reads a script without a BOM as ANSI, so it stays ASCII", s.name, b)
			}
		}
	}
}

// Windows PowerShell parses the Windows scripts (on Windows runners).
func TestInstallPS1Parses(t *testing.T) {
	ps, err := exec.LookPath("powershell")
	if err != nil {
		t.Skip("no Windows PowerShell on this runner")
	}
	for _, s := range ps1Scripts {
		got, err := s.render(testParams())
		if err != nil {
			t.Fatal(err)
		}
		path := filepath.Join(t.TempDir(), s.name)
		if err := os.WriteFile(path, got, 0o600); err != nil {
			t.Fatal(err)
		}
		check := "$e = $null; [void][System.Management.Automation.Language.Parser]::ParseFile($env:SCRIPT, [ref]$null, [ref]$e); " +
			"if ($e) { $e | ForEach-Object { $_.Message }; exit 1 }"
		cmd := exec.Command(ps, "-NoProfile", "-NonInteractive", "-Command", check)
		cmd.Env = append(os.Environ(), "SCRIPT="+path)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Errorf("%s does not parse: %v\n%s", s.name, err, out)
		}
	}
}

func TestInstallParses(t *testing.T) {
	sh, err := exec.LookPath("sh")
	if err != nil {
		t.Skip("no sh on this runner")
	}
	for _, s := range shellScripts {
		got, err := s.render(testParams())
		if err != nil {
			t.Fatal(err)
		}
		cmd := exec.Command(sh, "-n")
		cmd.Stdin = bytes.NewReader(got)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Errorf("%s: sh -n: %v\n%s", s.name, err, out)
		}
	}
}

// A value that would change a script's meaning is refused by every script
// that uses it, before anything is rendered.
func TestInstallRefusesHostileValues(t *testing.T) {
	for _, c := range []struct {
		name    string
		edit    func(*InstallParams)
		sh, ps1 bool // refused by the Linux scripts, by the Windows ones
	}{
		{"quote in origin", func(p *InstallParams) { p.Origin = "https://mesh.example.com'; id; '" }, true, true},
		{"origin with a path", func(p *InstallParams) { p.Origin = "https://mesh.example.com/x" }, true, true},
		{"non-canonical origin", func(p *InstallParams) { p.Origin = "https://Mesh.Example.com" }, true, true},
		{"plain http", func(p *InstallParams) { p.Origin = "http://mesh.example.com" }, true, true},
		{"substitution in version", func(p *InstallParams) { p.Version = "v1$(id)" }, true, true},
		{"newline in version", func(p *InstallParams) { p.Version = "v1\nid" }, true, true},
		{"short checksum", func(p *InstallParams) { p.SHA256Amd64 = "abc" }, true, false},
		{"quote in checksum", func(p *InstallParams) { p.SHA256Arm64 = strings.Repeat("a", 63) + "'" }, true, false},
		{"no Windows checksum", func(p *InstallParams) { p.SHA256Windows = "" }, false, true},
	} {
		p := testParams()
		c.edit(&p)
		for _, s := range shellScripts {
			if out, err := s.render(p); c.sh != (err != nil) || c.sh && out != nil {
				t.Errorf("%s: %s = %v", c.name, s.name, err)
			}
		}
		for _, s := range ps1Scripts {
			if out, err := s.render(p); c.ps1 != (err != nil) || c.ps1 && out != nil {
				t.Errorf("%s: %s = %v", c.name, s.name, err)
			}
		}
	}
}

// runScript runs the Linux script name, rendered, as `sh -s -- args...`, with
// env in place of the caller's MESHCTL_ variables.
func runScript(t *testing.T, name string, env []string, args ...string) (string, error) {
	t.Helper()
	sh, err := exec.LookPath("sh")
	if err != nil {
		t.Skip("no sh on this runner")
	}
	var script []byte
	for _, s := range shellScripts {
		if s.name == name {
			if script, err = s.render(testParams()); err != nil {
				t.Fatal(err)
			}
		}
	}
	if script == nil {
		t.Fatalf("no script %s", name)
	}
	cmd := exec.Command(sh, append([]string{"-s", "--"}, args...)...)
	cmd.Stdin = bytes.NewReader(script)
	for _, kv := range os.Environ() {
		if !strings.HasPrefix(kv, "MESHCTL_") {
			cmd.Env = append(cmd.Env, kv)
		}
	}
	cmd.Env = append(cmd.Env, env...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// --data (amendment A54) is checked before anything else happens, by every
// Linux script that takes it: a value that is not a plain absolute path, or
// that lies where the agent's unit cannot keep data, stops the script before
// its root check, so before it downloads, links, writes or deletes anything.
// A plain path gets as far as the root check.
func TestInstallChecksTheDataDir(t *testing.T) {
	base := map[string][]string{"install.sh": {"--token", "t0ken"}, "reinstall.sh": nil, "uninstall.sh": nil}
	for _, s := range shellScripts {
		for _, c := range []struct{ data, want string }{
			{"relative/dir", "absolute path"},
			{"-x", "absolute path"},
			{"/srv/$(id)", "only letters"},
			{"/srv/a b", "only letters"},
			{"/srv/x;id", "only letters"},
			{"/srv/x%i", "only letters"},
			{"/srv/x'y", "only letters"},
			{"/srv/x\nid", "only letters"},
			{"/srv/x\nExecStartPre=/bin/id", "only letters"},
			{"/srv/../etc", "'..' parts"},
			{"/srv/./x", "'..' parts"},
			{"/srv//x", "'..' parts"},
			{"/srv/x/", "'..' parts"},
			{"/", "'..' parts"},
			{"/home/me/meshctl", "cannot be under"},
			{"/root", "cannot be under"},
			{"/tmp/meshctl", "cannot be under"},
			{"/run/meshctl", "cannot be under"},
			{"/var/lib/meshctl", "cannot be under"},
			{"/var/lib/meshctl/default", "cannot be under"},
		} {
			// On Windows the MSYS sh parses its own command line again (quotes,
			// line breaks), so the value reaches the script as given only
			// through the environment there.
			if runtime.GOOS != "windows" {
				out, err := runScript(t, s.name, nil, append(base[s.name], "--data", c.data)...)
				if err == nil || !strings.Contains(out, "--data") || !strings.Contains(out, c.want) || strings.Contains(out, "must run as root") {
					t.Errorf("%s --data %q = %v\n%s", s.name, c.data, err, out)
				}
			}
			out, err := runScript(t, s.name, []string{"MESHCTL_DATA=" + c.data}, base[s.name]...)
			if err == nil || !strings.Contains(out, "--data") || !strings.Contains(out, c.want) || strings.Contains(out, "must run as root") {
				t.Errorf("%s MESHCTL_DATA=%q = %v\n%s", s.name, c.data, err, out)
			}
		}
		if os.Geteuid() == 0 {
			continue // past the checks a root run would download from the test origin, or uninstall
		}
		for _, data := range []string{"/volume1/meshctl/default", "/srv/meshctl-lab_1.0", "/var/lib/meshctl-nas"} {
			if out, err := runScript(t, s.name, nil, append(base[s.name], "--data", data)...); err == nil || !strings.Contains(out, "must run as root") {
				t.Errorf("%s --data %q stopped before the root check: %v\n%s", s.name, data, err, out)
			}
		}
	}
}

// Each Linux script refuses what it does not take, or a bad mesh, before its
// root check (amendment A55); what it takes gets as far as the root check.
func TestScriptsCheckTheirArguments(t *testing.T) {
	for _, c := range []struct {
		script string
		args   []string
		want   string // "": as far as the root check
	}{
		{"install.sh", nil, "no join token"},
		{"install.sh", []string{"--token"}, "--token needs a value"},
		{"install.sh", []string{"--token", "t0ken", "--mesh", "Lab"}, "--mesh must be a mesh name"},
		{"install.sh", []string{"--token", "t0ken", "--mesh", strings.Repeat("a", 33)}, "at most 32"},
		{"install.sh", []string{"--token", "t0ken", "--action", "rejoin"}, "unknown argument: --action"},
		{"install.sh", []string{"--token", "t0ken", "--mesh", "lab"}, ""},
		{"reinstall.sh", []string{"--action", "upgrade"}, "--action must be reinstall, rejoin or fresh"},
		{"reinstall.sh", []string{"--mesh", "-x"}, "--mesh must be a mesh name"},
		{"reinstall.sh", []string{"--yes"}, "unknown argument: --yes"},
		{"reinstall.sh", nil, ""},
		{"reinstall.sh", []string{"--mesh", "lab", "--action", "fresh", "--token", "t0ken", "--data", "/srv/meshctl"}, ""},
		{"uninstall.sh", []string{"--mesh", "a_b"}, "--mesh must be a mesh name"},
		{"uninstall.sh", []string{"--token", "t0ken"}, "unknown argument: --token"},
		{"uninstall.sh", []string{"--mesh", "lab", "--yes"}, ""},
	} {
		if c.want == "" && os.Geteuid() == 0 {
			continue // a root run would act on this host
		}
		out, err := runScript(t, c.script, nil, c.args...)
		switch {
		case err == nil:
			t.Errorf("%s %q succeeded:\n%s", c.script, c.args, out)
		case c.want == "" && !strings.Contains(out, "must run as root"):
			t.Errorf("%s %q stopped before the root check:\n%s", c.script, c.args, out)
		case c.want != "" && (!strings.Contains(out, c.want) || strings.Contains(out, "must run as root")):
			t.Errorf("%s %q = %q, want %q before the root check", c.script, c.args, out, c.want)
		}
	}
}

// The scripts arrive on stdin (curl ... | sh -s), so a question read from
// stdin would read the script itself: every read takes the terminal.
func TestScriptsReadTheTerminal(t *testing.T) {
	for _, s := range shellScripts {
		for i, line := range strings.Split(withLib(s.tmpl, libTemplate), "\n") {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "#") {
				continue
			}
			for _, f := range strings.Fields(line) {
				if f == "read" && !strings.Contains(line, "</dev/tty") {
					t.Errorf("%s line %d reads stdin: %s", s.name, i+1, line)
				}
			}
		}
	}
}

// No message a script prints holds the join token (CLAUDE.md: secrets never
// reach logs), wherever it came from: --token, MESHCTL_TOKEN or the prompt.
func TestScriptsNeverPrintTheToken(t *testing.T) {
	for _, s := range shellScripts {
		for i, line := range strings.Split(withLib(s.tmpl, libTemplate), "\n") {
			if !strings.Contains(line, "$token") && !strings.Contains(line, "${token") {
				continue
			}
			for _, cmd := range []string{"die \"", "echo \"", "printf "} {
				at := strings.Index(line, cmd)
				if at < 0 {
					continue
				}
				said := line[at+len(cmd):]
				if cmd != "printf " {
					said, _, _ = strings.Cut(said, "\"")
				}
				if strings.Contains(said, "$token") || strings.Contains(said, "${token") {
					t.Errorf("%s line %d prints the token: %s", s.name, i+1, line)
				}
			}
		}
	}
}

// The Windows scripts print no string that holds the join token either --
// no double-quoted string in them names $token -- and never put it on
// meshctl's command line, where other processes could read it: it reaches
// the join in MESHCTL_TOKEN only.
func TestPS1ScriptsNeverShowTheToken(t *testing.T) {
	for _, s := range ps1Scripts {
		for i, line := range strings.Split(withLib(s.tmpl, libPS1Template), "\n") {
			if strings.HasPrefix(strings.TrimSpace(line), "#") {
				continue
			}
			if strings.Contains(line, "--token") {
				t.Errorf("%s line %d puts a token on a command line: %s", s.name, i+1, line)
			}
			for j, part := range strings.Split(strings.ToLower(line), "\"") {
				if j%2 == 1 && (strings.Contains(part, "$token") || strings.Contains(part, "${token")) {
					t.Errorf("%s line %d shows the token: %s", s.name, i+1, line)
				}
			}
		}
	}
}

// The Windows scripts run in the PowerShell session that pipes them into
// iex: each is one script block, so its functions, variables and preferences
// ($ErrorActionPreference) end with it, and nothing in it exits, which would
// close that session.
func TestPS1ScriptsKeepToTheirBlock(t *testing.T) {
	for _, s := range ps1Scripts {
		var code []string
		for _, line := range strings.Split(withLib(s.tmpl, libPS1Template), "\n") {
			if l := strings.TrimSpace(line); l != "" && !strings.HasPrefix(l, "#") {
				code = append(code, line)
			}
		}
		if len(code) < 2 || code[0] != "& {" || code[len(code)-1] != "}" {
			t.Errorf("%s is not one script block, & { ... }", s.name)
		}
		for _, line := range code {
			l := strings.ToLower(line)
			if slices.Contains(strings.Fields(l), "exit") || strings.Contains(l, "$global:") || strings.Contains(l, "$script:") {
				t.Errorf("%s reaches out of its block: %s", s.name, line)
			}
		}
	}
}

// Each Windows script refuses a bad MESHCTL_MESH, and a value outside the
// fixed set a variable takes, before its elevation check (amendment A55), so
// before it downloads, joins or deletes anything. It runs as the docs say,
// piped into iex.
func TestPS1ScriptsCheckTheirVariables(t *testing.T) {
	ps, err := exec.LookPath("powershell")
	if err != nil {
		t.Skip("no Windows PowerShell on this runner")
	}
	for _, c := range []struct {
		script string
		env    []string
		want   string
	}{
		{"install.ps1", nil, "no join token"},
		{"install.ps1", []string{"MESHCTL_TOKEN=t0ken", "MESHCTL_MESH=Lab"}, "MESHCTL_MESH must be a mesh name"},
		{"install.ps1", []string{"MESHCTL_TOKEN=t0ken", "MESHCTL_MESH=" + strings.Repeat("a", 33)}, "MESHCTL_MESH must be a mesh name"},
		{"reinstall.ps1", []string{"MESHCTL_ACTION=upgrade"}, "MESHCTL_ACTION must be reinstall, rejoin or fresh"},
		{"reinstall.ps1", []string{"MESHCTL_ACTION=Fresh"}, "MESHCTL_ACTION must be reinstall, rejoin or fresh"},
		{"reinstall.ps1", []string{"MESHCTL_MESH=-x"}, "MESHCTL_MESH must be a mesh name"},
		{"uninstall.ps1", []string{"MESHCTL_MESH=a_b"}, "MESHCTL_MESH must be a mesh name"},
		{"uninstall.ps1", []string{"MESHCTL_YES=y"}, "MESHCTL_YES must be 'yes' or unset"},
	} {
		var script []byte
		for _, s := range ps1Scripts {
			if s.name == c.script {
				if script, err = s.render(testParams()); err != nil {
					t.Fatal(err)
				}
			}
		}
		path := filepath.Join(t.TempDir(), c.script)
		if err := os.WriteFile(path, script, 0o600); err != nil {
			t.Fatal(err)
		}
		run := "try { [IO.File]::ReadAllText($env:SCRIPT) | Invoke-Expression } catch { [Console]::Out.WriteLine($_.Exception.Message); exit 1 }"
		cmd := exec.Command(ps, "-NoProfile", "-NonInteractive", "-Command", run)
		for _, kv := range os.Environ() {
			if !strings.HasPrefix(strings.ToUpper(kv), "MESHCTL_") {
				cmd.Env = append(cmd.Env, kv)
			}
		}
		cmd.Env = append(append(cmd.Env, "SCRIPT="+path), c.env...)
		out, err := cmd.CombinedOutput()
		if err == nil || !strings.Contains(string(out), c.script+": "+c.want) || strings.Contains(string(out), "elevated") {
			t.Errorf("%s with %q = %v, want %q before the elevation check\n%s", c.script, c.env, err, c.want, out)
		}
	}
}

// The drop-in the scripts write for --data (amendments A54, A55) waits for
// the data dir's file system, drops the instance's own state directory (the
// link stands in for it), lets it write in the data dir and runs the unit's
// command on it, with the socket local commands find through the link
// (/run/meshctl/<mesh>.sock, agent.DefaultSocket for /var/lib/meshctl/<mesh>):
// nothing else, and no path but the checked one.
func TestInstallDataDropIn(t *testing.T) {
	start := strings.Index(libTemplate, "cat <<MESHCTL_DROPIN\n")
	end := strings.Index(libTemplate, "\nMESHCTL_DROPIN\n")
	if start < 0 || end < start {
		t.Fatal("lib.sh.tmpl has no MESHCTL_DROPIN here-document")
	}
	var got []string
	section := ""
	for _, line := range strings.Split(libTemplate[start+len("cat <<MESHCTL_DROPIN\n"):end], "\n") {
		switch line = strings.TrimSpace(line); {
		case line == "" || strings.HasPrefix(line, "#"):
		case strings.HasPrefix(line, "["):
			section = line
		default:
			got = append(got, section+" "+line)
		}
	}
	run := strings.TrimSuffix(wantService["ExecStart"], " --mesh %i")
	want := []string{"[Unit] RequiresMountsFor=$1", "[Service] StateDirectory=", "[Service] ReadWritePaths=$1", "[Service] ExecStart=",
		"[Service] ExecStart=" + run + " --data $1 --socket /run/meshctl/%i.sock"}
	if !reflect.DeepEqual(got, want) || run == wantService["ExecStart"] {
		t.Fatalf("drop-in %q, want exactly %q", got, want)
	}
}

// Every token that would widen a container's privileges.
var widening = []string{"cap_add", "devices", "network_mode", "pid:", "ipc:", "userns_mode", "privileged: true",
	"docker.sock", "seccomp", "apparmor", "SYS_ADMIN", "/dev/", "host:"}

func service(t *testing.T, file string) (map[string]any, map[string]any) {
	t.Helper()
	b, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	for _, w := range widening {
		if strings.Contains(string(b), w) {
			t.Errorf("%s contains the widening token %q", file, w)
		}
	}
	var doc map[string]any
	if err := yaml.Unmarshal(b, &doc); err != nil {
		t.Fatal(err)
	}
	svcs, _ := doc["services"].(map[string]any)
	svc, _ := svcs["meshctl"].(map[string]any)
	if len(svcs) != 1 || svc == nil {
		t.Fatalf("%s must define exactly the meshctl service", file)
	}
	return doc, svc
}

func keys(m map[string]any) []string {
	var out []string
	for k := range m {
		out = append(out, k)
	}
	return out
}

func checkPosture(t *testing.T, file string, svc map[string]any, want []string) {
	t.Helper()
	got := keys(svc)
	if !sameSet(got, want) {
		t.Errorf("%s service keys = %v, want exactly %v", file, got, want)
	}
	exact := map[string]any{
		"user":         "65532:65532",
		"privileged":   false,
		"cap_drop":     []any{"ALL"},
		"security_opt": []any{"no-new-privileges=true"},
		"read_only":    true,
	}
	for k, v := range exact {
		if !reflect.DeepEqual(svc[k], v) {
			t.Errorf("%s: %s = %#v, want %#v", file, k, svc[k], v)
		}
	}
	vols, _ := svc["volumes"].([]any)
	for _, v := range vols {
		if s, _ := v.(string); strings.HasPrefix(s, "/") || strings.HasPrefix(s, ".") {
			t.Errorf("%s mounts a host path %q; only named volumes are allowed", file, s)
		}
	}
}

func sameSet(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	seen := map[string]bool{}
	for _, s := range a {
		seen[s] = true
	}
	for _, s := range b {
		if !seen[s] {
			return false
		}
	}
	return true
}

func TestControlComposePosture(t *testing.T) {
	if !bytes.Equal(ComposeControl(), mustRead(t, "compose.control.yml")) {
		t.Fatal("the embedded compose file is stale")
	}
	doc, svc := service(t, "compose.control.yml")
	checkPosture(t, "compose.control.yml", svc, []string{"image", "environment", "user", "privileged", "cap_drop", "security_opt",
		"read_only", "tmpfs", "volumes", "expose", "networks", "restart", "mem_limit", "pids_limit", "healthcheck"})
	if !reflect.DeepEqual(svc["environment"], controlEnv()) {
		t.Errorf("the control plane's environment = %#v, want exactly %#v", svc["environment"], controlEnv())
	}
	if !reflect.DeepEqual(svc["expose"], []any{"${MESHCTL_PORT:-8080}"}) {
		t.Errorf("expose = %#v: the control plane's port only", svc["expose"])
	}
	if img, _ := svc["image"].(string); !strings.HasPrefix(img, "ghcr.io/hafio/vpn-mesh-utility:${MESHCTL_VERSION:?") {
		t.Errorf("image %q must be the pinned release, required from the environment", img)
	}
	nets, _ := doc["networks"].(map[string]any)
	dok, _ := nets["dokploy-network"].(map[string]any)
	if dok["external"] != true {
		t.Error("dokploy-network must be external (Dokploy owns it)")
	}
}

// controlEnv is the control plane's environment in its compose and stack
// files (amendments A47, A48): `meshctl start` in control mode, every value
// from the app's environment, the origin required.
func controlEnv() map[string]any {
	return map[string]any{
		"MESHCTL_MODE":            "control",
		"MESHCTL_PORT":            "${MESHCTL_PORT:-8080}",
		"MESHCTL_URL":             "${MESHCTL_URL:?set MESHCTL_URL to the public https origin}",
		"MESHCTL_TRUSTED_PROXIES": "${MESHCTL_TRUSTED_PROXIES:-}",
		"MESHCTL_HUB_DOWN_AFTER":  "${MESHCTL_HUB_DOWN_AFTER:-}",
		"MESHCTL_LOG_LEVEL":       "${MESHCTL_LOG_LEVEL:-}",
		"TZ":                      "${TZ:-UTC}",
	}
}

// startHealth is the image's own health check, which follows MESHCTL_MODE.
func startHealth() []any { return []any{"CMD", "/app/meshctl", "healthcheck", "start"} }

func TestDevComposePosture(t *testing.T) {
	_, svc := service(t, "compose.dev.yml")
	checkPosture(t, "compose.dev.yml", svc, []string{"image", "environment", "user", "privileged", "cap_drop", "security_opt",
		"read_only", "tmpfs", "volumes", "ports", "restart", "mem_limit", "pids_limit", "healthcheck"})
	if env := map[string]any{"MESHCTL_MODE": "control", "MESHCTL_URL": "http://127.0.0.1:8080"}; !reflect.DeepEqual(svc["environment"], env) {
		t.Errorf("the dev control plane's environment = %#v, want %#v", svc["environment"], env)
	}
	if hc, _ := svc["healthcheck"].(map[string]any); !reflect.DeepEqual(hc["test"], startHealth()) {
		t.Errorf("healthcheck = %#v", svc["healthcheck"])
	}
	if !reflect.DeepEqual(svc["ports"], []any{"127.0.0.1:8080:8080"}) {
		t.Errorf("the dev stack publishes on loopback only: %#v", svc["ports"])
	}
}

// The Swarm stack states the posture a stack can: non-root, no
// capabilities, read-only root, one task pinned to a labelled node and
// stopped before the next starts; nothing that widens it.
func TestControlStackPosture(t *testing.T) {
	doc, svc := service(t, "stack.control.yml")
	if got, want := keys(svc), []string{"image", "environment", "user", "cap_drop", "read_only", "volumes", "networks",
		"stop_grace_period", "healthcheck", "deploy"}; !sameSet(got, want) {
		t.Errorf("stack service keys = %v, want exactly %v", got, want)
	}
	for k, v := range map[string]any{"user": "65532:65532", "cap_drop": []any{"ALL"}, "read_only": true,
		"environment": controlEnv()} {
		if !reflect.DeepEqual(svc[k], v) {
			t.Errorf("stack: %s = %#v, want %#v", k, svc[k], v)
		}
	}
	if img, _ := svc["image"].(string); !strings.HasPrefix(img, "ghcr.io/hafio/vpn-mesh-utility:${MESHCTL_VERSION:?") {
		t.Errorf("image %q must be the pinned release, required from the environment", img)
	}
	vols, _ := svc["volumes"].([]any)
	if len(vols) != 2 || vols[0] != "meshctl-data:/data" {
		t.Errorf("stack volumes = %#v: the named data volume and a tmpfs only", vols)
	}
	if tmp, _ := vols[1].(map[string]any); tmp["type"] != "tmpfs" || tmp["target"] != "/tmp" {
		t.Errorf("stack tmpfs = %#v", vols[1])
	}
	dep, _ := svc["deploy"].(map[string]any)
	place, _ := dep["placement"].(map[string]any)
	update, _ := dep["update_config"].(map[string]any)
	if dep["mode"] != "replicated" || dep["replicas"] != 1 || !reflect.DeepEqual(place["constraints"], []any{"node.labels.meshctl.control == true"}) ||
		update["order"] != "stop-first" {
		t.Errorf("stack deploy = %#v", dep)
	}
	nets, _ := doc["networks"].(map[string]any)
	if dok, _ := nets["dokploy-network"].(map[string]any); dok["external"] != true {
		t.Error("dokploy-network must be external (Dokploy owns it)")
	}
	// The same environment and health check as the compose file: only how
	// it is run differs.
	_, compose := service(t, "compose.control.yml")
	if !reflect.DeepEqual(svc["environment"], compose["environment"]) || !reflect.DeepEqual(svc["healthcheck"], compose["healthcheck"]) {
		t.Errorf("the stack runs with %#v, the compose file with %#v", svc["environment"], compose["environment"])
	}
	if hc, _ := svc["healthcheck"].(map[string]any); !reflect.DeepEqual(hc["test"], startHealth()) {
		t.Errorf("healthcheck = %#v", svc["healthcheck"])
	}
}

func mustRead(t *testing.T, name string) []byte {
	t.Helper()
	b, err := os.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	return b
}
