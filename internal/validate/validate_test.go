package validate

import (
	"errors"
	"strings"
	"testing"
)

// hostile is the shared set of injection payloads every string validator must
// refuse: shell metacharacters, quoting, command substitution, newlines that
// would add a YAML key or a systemd directive, NUL, and path traversal.
var hostile = []string{
	"a;rm -rf /",
	"a|b",
	"a&b",
	"$(id)",
	"`id`",
	"a\nprivileged: true",
	"a\r\nExecStartPre=/bin/sh",
	"a\x00b",
	"../etc/passwd",
	"a b",
	"a'b",
	"a>b",
	"${X}",
	"--- ",
}

func TestName(t *testing.T) {
	ok := []string{"a", "vps-1", "nas", "laptop-01", strings.Repeat("a", 32)}
	for _, s := range ok {
		t.Run("accepts "+s, func(t *testing.T) {
			got, err := Name("name", s)
			if err != nil || got != s {
				t.Fatalf("Name(%q) = %q, %v; want %q, nil", s, got, err, s)
			}
		})
	}
	bad := append([]string{"", "-a", "a-", "Vps1", "a_b", "a.b", strings.Repeat("a", 33)}, hostile...)
	for _, s := range bad {
		t.Run("rejects "+s, func(t *testing.T) {
			if _, err := Name("name", s); err == nil {
				t.Fatalf("Name(%q) accepted a bad name", s)
			}
		})
	}
}

func TestDNSName(t *testing.T) {
	cases := []struct {
		in, want string
		ok       bool
	}{
		{"immich", "immich", true},
		{"photos.nas", "photos.nas", true},
		{"photos.nas.mesh", "photos.nas", true},
		{"", "", false},
		{"a..b", "", false},
		{"-a.b", "", false},
		{"A.b", "", false},
		{"a b", "", false},
	}
	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			got, err := DNSName("record", c.in)
			if (err == nil) != c.ok || got != c.want {
				t.Fatalf("DNSName(%q) = %q, %v; want %q ok=%v", c.in, got, err, c.want, c.ok)
			}
		})
	}
}

func TestHostnameFoldsCase(t *testing.T) {
	got, err := Hostname("hostname", "NAS-01.Local.")
	if err != nil || got != "nas-01.local" {
		t.Fatalf("Hostname = %q, %v; want nas-01.local", got, err)
	}
	for _, s := range hostile {
		if _, err := Hostname("hostname", s); err == nil {
			t.Errorf("Hostname(%q) accepted a hostile value", s)
		}
	}
}

func TestUsername(t *testing.T) {
	if _, err := Username("user", "admin"); err != nil {
		t.Fatalf("Username(admin): %v", err)
	}
	for _, s := range append([]string{"", "1admin", "Admin"}, hostile...) {
		if _, err := Username("user", s); err == nil {
			t.Errorf("Username(%q) accepted a bad name", s)
		}
	}
}

func TestEnum(t *testing.T) {
	if got, err := Enum("role", "spoke", "server", "spoke", "phone"); err != nil || got != "spoke" {
		t.Fatalf("Enum = %q, %v", got, err)
	}
	_, err := Enum("role", "admin", "server", "spoke", "phone")
	if err == nil || !strings.Contains(err.Error(), "server, spoke, phone") {
		t.Fatalf("Enum error should list the allowed values, got %v", err)
	}
}

func TestIPv4(t *testing.T) {
	if a, err := IPv4("ip", "10.77.0.1"); err != nil || a.String() != "10.77.0.1" {
		t.Fatalf("IPv4 = %v, %v", a, err)
	}
	for _, s := range []string{"", "10.77.0", "::1", "::ffff:10.0.0.1", "10.77.0.1/32", "fe80::1%eth0"} {
		if _, err := IPv4("ip", s); err == nil {
			t.Errorf("IPv4(%q) accepted a non-IPv4 value", s)
		}
	}
}

func TestPrefix4(t *testing.T) {
	if p, err := Prefix4("cidr", "10.77.0.0/16"); err != nil || p.String() != "10.77.0.0/16" {
		t.Fatalf("Prefix4 = %v, %v", p, err)
	}
	_, err := Prefix4("cidr", "10.77.0.5/16")
	if err == nil || !strings.Contains(err.Error(), "10.77.0.0/16") {
		t.Fatalf("host bits must be rejected with a suggestion, got %v", err)
	}
	for _, s := range []string{"", "10.77.0.0", "::/0", "10.77.0.0/33"} {
		if _, err := Prefix4("cidr", s); err == nil {
			t.Errorf("Prefix4(%q) accepted a bad prefix", s)
		}
	}
}

func TestPortAndRange(t *testing.T) {
	for _, n := range []int{1, 51820, 65535} {
		if _, err := Port("port", n); err != nil {
			t.Errorf("Port(%d): %v", n, err)
		}
	}
	for _, n := range []int{0, -1, 65536} {
		if _, err := Port("port", n); err == nil {
			t.Errorf("Port(%d) accepted an out-of-range port", n)
		}
	}
	if _, err := IntRange("mtu", 1279, 1280, 1500); err == nil {
		t.Error("IntRange accepted a value below the range")
	}
	if _, err := IntRange("mtu", 1420, 1280, 1500); err != nil {
		t.Errorf("IntRange(1420): %v", err)
	}
}

func TestEndpoint(t *testing.T) {
	if ap, err := Endpoint("endpoint", "203.0.113.10:51820"); err != nil || ap.String() != "203.0.113.10:51820" {
		t.Fatalf("Endpoint = %v, %v", ap, err)
	}
	for _, s := range []string{"", "203.0.113.10", "vps1.example.com:51820", "[::1]:51820", "203.0.113.10:0", "203.0.113.10:99999"} {
		if _, err := Endpoint("endpoint", s); err == nil {
			t.Errorf("Endpoint(%q) accepted a bad endpoint", s)
		}
	}
}

func TestKey32(t *testing.T) {
	good := "AQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQE="
	if _, err := Key32("key", good); err != nil {
		t.Fatalf("Key32(good): %v", err)
	}
	for _, s := range []string{"", "AQEB", good[:43] + "!", strings.Repeat("A", 44)} {
		_, err := Key32("key", s)
		if s == strings.Repeat("A", 44) {
			// 44 x 'A' is not canonical padded base64 of 32 bytes.
			if err == nil {
				t.Errorf("Key32 accepted unpadded 44-char input")
			}
			continue
		}
		if err == nil {
			t.Errorf("Key32(%q) accepted a bad key", s)
		}
	}
}

func TestTokenMasksTheValue(t *testing.T) {
	good := strings.Repeat("a", 43)
	if _, err := Token("token", good); err != nil {
		t.Fatalf("Token(good): %v", err)
	}
	secret := "supersecretvalue-that-is-not-a-valid-token!!"
	_, err := Token("token", secret)
	if err == nil {
		t.Fatal("Token accepted a malformed token")
	}
	if strings.Contains(err.Error(), "secretvalue") {
		t.Fatalf("Token error leaks the value: %v", err)
	}
	var ve *Error
	if !errors.As(err, &ve) || ve.Field != "token" {
		t.Fatalf("want *Error for field token, got %#v", err)
	}
}

func TestLocalTokenMasksTheValue(t *testing.T) {
	if _, err := LocalToken("t", strings.Repeat("Z", 43)); err != nil {
		t.Fatalf("LocalToken(good): %v", err)
	}
	for _, s := range append([]string{"", strings.Repeat("a", 42), strings.Repeat("a", 44)}, hostile...) {
		if _, err := LocalToken("t", s); err == nil || (len(s) > 8 && strings.Contains(err.Error(), s)) {
			t.Errorf("LocalToken(%q) = %v", s, err)
		}
	}
}

func TestPipePath(t *testing.T) {
	for _, s := range []string{`\\.\pipe\meshctl-agent`, `\\.\pipe\meshctl-test_1.a`} {
		if _, err := PipePath("socket", s); err != nil {
			t.Errorf("PipePath(%q) = %v", s, err)
		}
	}
	for _, s := range append([]string{`\\.\pipe\`, `\\server\pipe\x`, `C:\data\agent.sock`, `\\.\pipe\a\b`, "/run/meshctl/agent.sock"}, hostile...) {
		if _, err := PipePath("socket", s); err == nil {
			t.Errorf("PipePath(%q) accepted it", s)
		}
	}
}

func TestVersionAndSemVer(t *testing.T) {
	for _, s := range []string{"dev", "v1.2.3", "v1.2.3-4-gabcdef0-dirty", "abcdef0"} {
		if _, err := Version("version", s); err != nil {
			t.Errorf("Version(%q): %v", s, err)
		}
	}
	for _, s := range append([]string{"", "-v1"}, hostile...) {
		if _, err := Version("version", s); err == nil {
			t.Errorf("Version(%q) accepted a bad version", s)
		}
	}
	for _, s := range []string{"v1.2.3", "1.2.3", "v1.0.0-rc.1"} {
		if _, err := SemVer("version", s); err != nil {
			t.Errorf("SemVer(%q): %v", s, err)
		}
	}
	for _, s := range []string{"dev", "v1.2", "v01.2.3", "v1.2.3-4-gabcdef0-dirty+x"} {
		if _, err := SemVer("version", s); err == nil {
			t.Errorf("SemVer(%q) accepted a non-release version", s)
		}
	}
}

func TestServerURL(t *testing.T) {
	cases := []struct {
		in, want string
		ok       bool
	}{
		{"https://mesh.example.com", "https://mesh.example.com", true},
		{"https://Mesh.Example.com/", "https://mesh.example.com", true},
		{"https://mesh.example.com:8443", "https://mesh.example.com:8443", true},
		{"http://127.0.0.1:8080", "http://127.0.0.1:8080", true},
		{"http://localhost:8080", "http://localhost:8080", true},
		{"http://mesh.example.com", "", false},
		{"https://user:pw@mesh.example.com", "", false},
		{"https://mesh.example.com/x", "", false},
		{"https://mesh.example.com?a=1", "", false},
		{"ftp://mesh.example.com", "", false},
		{"https://x';rm -rf /;'", "", false},
		{"mesh.example.com", "", false},
	}
	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			got, err := ServerURL("server", c.in)
			if (err == nil) != c.ok || got != c.want {
				t.Fatalf("ServerURL(%q) = %q, %v; want %q ok=%v", c.in, got, err, c.want, c.ok)
			}
		})
	}
}

func TestPOSIXPath(t *testing.T) {
	for _, s := range []string{"/volume1/docker/meshctl", "/data", "/"} {
		if _, err := POSIXPath("path", s); err != nil {
			t.Errorf("POSIXPath(%q): %v", s, err)
		}
	}
	for _, s := range append([]string{"", "data", "/a/../b", "/a/./b", "//a", "/a b"}, hostile...) {
		if _, err := POSIXPath("path", s); err == nil {
			t.Errorf("POSIXPath(%q) accepted a bad path", s)
		}
	}
}

func TestFilePath(t *testing.T) {
	for _, s := range []string{"/data", `C:\ProgramData\meshctl`, "rel/dir"} {
		if _, err := FilePath("data", s); err != nil {
			t.Errorf("FilePath(%q): %v", s, err)
		}
	}
	for _, s := range []string{"", "/data?mode=memory", "/data#x", "/data%00", "/da\nta"} {
		if _, err := FilePath("data", s); err == nil {
			t.Errorf("FilePath(%q) accepted a bad path", s)
		}
	}
}

func TestText(t *testing.T) {
	if _, err := Text("note", "Immich (photos) -- port 2283", 200); err != nil {
		t.Fatalf("Text should allow punctuation that never reaches a shell: %v", err)
	}
	for _, s := range []string{"a\nb", "a\tb", "a\x00b", "\xff\xfe", strings.Repeat("x", 201)} {
		if _, err := Text("note", s, 200); err == nil {
			t.Errorf("Text(%q) accepted a bad value", s)
		}
	}
}

func TestArg(t *testing.T) {
	ok := []string{
		"--permanent",
		"--zone=mesh",
		"--add-port=51820/udp",
		`rule family="ipv4" source address="10.77.0.2/32" port port="2377" protocol="tcp" accept`,
	}
	for _, s := range ok {
		if err := Arg(s); err != nil {
			t.Errorf("Arg(%q): %v", s, err)
		}
	}
	for _, s := range append([]string{"", strings.Repeat("a", 513)}, hostile[:13]...) {
		if s == "a b" || s == "../etc/passwd" {
			continue // space and dots are legal in an argv element; no shell sees it
		}
		if err := Arg(s); err == nil {
			t.Errorf("Arg(%q) accepted a hostile argv element", s)
		}
	}
}

// FuzzArg pins the invariant the exec runner relies on: nothing Arg accepts
// contains a byte outside its documented charset.
func FuzzArg(f *testing.F) {
	for _, s := range hostile {
		f.Add(s)
	}
	f.Add(`rule family="ipv4" accept`)
	f.Fuzz(func(t *testing.T, s string) {
		if Arg(s) != nil {
			return
		}
		for i := 0; i < len(s); i++ {
			c := s[i]
			if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || strings.IndexByte(` ._/:=,"-`, c) >= 0) {
				t.Fatalf("Arg accepted %q with byte %q", s, c)
			}
		}
	})
}

// FuzzName pins that an accepted node name is always safe as a DNS label,
// a compose service name and a file name.
func FuzzName(f *testing.F) {
	for _, s := range hostile {
		f.Add(s)
	}
	f.Add("vps-1")
	f.Fuzz(func(t *testing.T, s string) {
		got, err := Name("name", s)
		if err != nil {
			return
		}
		if len(got) == 0 || len(got) > 32 || strings.HasPrefix(got, "-") || strings.HasSuffix(got, "-") {
			t.Fatalf("Name accepted %q", s)
		}
		for i := 0; i < len(got); i++ {
			c := got[i]
			if !(c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '-') {
				t.Fatalf("Name accepted %q with byte %q", s, c)
			}
		}
	})
}
