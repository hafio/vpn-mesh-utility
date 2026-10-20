// Package validate is the single input gate for meshctl.
//
// Every value that arrives from outside the process -- a CLI flag, an HTTP body,
// an agent report, a config pushed by the server -- passes through one of these
// functions before it is stored, rendered into an artifact (compose file,
// systemd unit, install.sh, firewalld rule) or handed to an exec'd command.
// Each function accepts a positive grammar rather than scrubbing a denylist,
// returns the normalised value, and fails with an *Error that names the field,
// the offending value (masked for secrets) and the rule it broke.
//
// Arg is the second, shell-layer half of the gate: the exec runner re-checks
// every argv element with it right before the command runs, so a value that
// bypassed the boundary validation still cannot smuggle a metacharacter in.
//
// The package is stdlib-only and imports nothing else from this module, so both
// the server and the agent use it without crossing the module boundary.
package validate

import (
	"encoding/base64"
	"fmt"
	"net/netip"
	"net/url"
	"path"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Error reports one rejected value. Value is already masked when the field
// carries a secret, so the message is always safe to log or return to a client.
type Error struct {
	Field  string
	Value  string
	Reason string
}

func (e *Error) Error() string {
	if e.Value == "" {
		return fmt.Sprintf("%s: %s", e.Field, e.Reason)
	}
	return fmt.Sprintf("%s: %s (got %q)", e.Field, e.Reason, e.Value)
}

func fail(field, value, reason string) error {
	// Never echo a huge or binary value back: clip and escape it.
	if len(value) > 64 {
		value = value[:64] + "..."
	}
	return &Error{Field: field, Value: value, Reason: reason}
}

// mask hides all but a short prefix of a secret so an error stays actionable
// ("which token did I paste?") without leaking it.
func mask(s string) string {
	if len(s) <= 4 {
		return "****"
	}
	return s[:4] + "****"
}

var (
	reName     = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]{0,30}[a-z0-9])?$`)
	reLabel    = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?$`)
	reUser     = regexp.MustCompile(`^[a-z][a-z0-9_.-]{0,31}$`)
	reToken    = regexp.MustCompile(`^[A-Za-z0-9_-]{43}$`)
	rePipe     = regexp.MustCompile(`^\\\\\.\\pipe\\[A-Za-z0-9._-]{1,200}$`)
	reVersion  = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._+-]{0,63}$`)
	reSemVer   = regexp.MustCompile(`^v?(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(-[0-9A-Za-z.-]{1,32})?$`)
	rePosix    = regexp.MustCompile(`^/[A-Za-z0-9._/-]*$`)
	reArgChars = regexp.MustCompile(`^[A-Za-z0-9 ._/:=,"-]+$`)
)

// Name validates a node name: a lowercase DNS label of 1-32 characters. Node
// names become records in the mesh's DNS zone, compose service names and
// file names, so the grammar is the strictest of all of those sinks.
func Name(field, s string) (string, error) {
	if !reName.MatchString(s) {
		return "", fail(field, s, "must be 1-32 chars of a-z, 0-9 and '-', not starting or ending with '-'")
	}
	return s, nil
}

// DNSName validates a record name inside the mesh zone: one or more labels
// separated by dots, given without the zone, at most 200 characters.
func DNSName(field, s string) (string, error) {
	if s == "" || len(s) > 200 {
		return "", fail(field, s, "must be 1-200 chars")
	}
	for _, l := range strings.Split(s, ".") {
		if !reLabel.MatchString(l) {
			return "", fail(field, s, "each dot-separated label must be 1-63 chars of a-z, 0-9 and '-', not starting or ending with '-'")
		}
	}
	return s, nil
}

// Zone validates the mesh's DNS zone (amendment A36): one to four labels
// as in DNSName, at most 100 characters, without a leading or trailing dot.
// local and localhost are refused: hosts resolve those themselves (mDNS,
// loopback).
func Zone(field, s string) (string, error) {
	labels := strings.Split(s, ".")
	if s == "" || len(s) > 100 || len(labels) > 4 {
		return "", fail(field, s, "must be 1-4 dot-separated labels, at most 100 chars, such as mesh or home.arpa")
	}
	for _, l := range labels {
		if !reLabel.MatchString(l) {
			return "", fail(field, s, "each dot-separated label must be 1-63 chars of a-z, 0-9 and '-', not starting or ending with '-'")
		}
	}
	if last := labels[len(labels)-1]; last == "local" || last == "localhost" {
		return "", fail(field, s, "must not end in local or localhost, which hosts resolve themselves (mDNS, loopback)")
	}
	return s, nil
}

// MaxDomains caps the mesh's extra DNS domains.
const MaxDomains = 8

// Domains validates the mesh's extra DNS domains (amendment A39), which
// every node resolves through the mesh and the hosts send to it: each as a
// Zone, lower case without a trailing dot, at most MaxDomains, none under
// another or listed twice, no .arpa domain but home.arpa and under it, and
// none equal to, under or around zone ("" skips that check). It returns them
// normalised and sorted.
func Domains(field, zone string, ds []string) ([]string, error) {
	if len(ds) > MaxDomains {
		return nil, fail(field, strings.Join(ds, ","), fmt.Sprintf("at most %d domains", MaxDomains))
	}
	out := make([]string, 0, len(ds))
	for _, d := range ds {
		d = strings.ToLower(strings.TrimSuffix(strings.TrimSpace(d), "."))
		if _, err := Zone(field, d); err != nil {
			return nil, err
		}
		if d == "arpa" || (strings.HasSuffix(d, ".arpa") && d != "home.arpa" && !strings.HasSuffix(d, ".home.arpa")) {
			return nil, fail(field, d, "must not be an .arpa domain other than home.arpa (reverse lookups are the host's)")
		}
		if zone != "" && (d == zone || strings.HasSuffix(d, "."+zone) || strings.HasSuffix(zone, "."+d)) {
			return nil, fail(field, d, "must not be the mesh's zone "+zone+", inside it or around it")
		}
		out = append(out, d)
	}
	sort.Strings(out)
	for i := range out {
		for j := range out {
			if i != j && (out[i] == out[j] || strings.HasSuffix(out[i], "."+out[j])) {
				return nil, fail(field, out[i], "is listed twice or inside "+out[j]+"; list each domain once")
			}
		}
	}
	return out, nil
}

// Hostname validates a hostname an agent reports about itself (RFC 1123,
// case-folded to lowercase).
func Hostname(field, s string) (string, error) {
	h := strings.ToLower(strings.TrimSuffix(s, "."))
	if h == "" || len(h) > 253 {
		return "", fail(field, s, "must be 1-253 chars")
	}
	for _, l := range strings.Split(h, ".") {
		if !reLabel.MatchString(l) {
			return "", fail(field, s, "must be dot-separated labels of a-z, 0-9 and '-'")
		}
	}
	return h, nil
}

// Username validates a UI login name.
func Username(field, s string) (string, error) {
	if !reUser.MatchString(s) {
		return "", fail(field, s, "must be 1-32 chars of a-z, 0-9, '_', '.', '-', starting with a letter")
	}
	return s, nil
}

// Enum validates s against a fixed set of allowed values.
func Enum(field, s string, allowed ...string) (string, error) {
	for _, a := range allowed {
		if s == a {
			return s, nil
		}
	}
	return "", fail(field, s, "must be one of "+strings.Join(allowed, ", "))
}

// IPv4 validates a bare IPv4 address (no zone, no IPv4-mapped IPv6 form).
func IPv4(field, s string) (netip.Addr, error) {
	a, err := netip.ParseAddr(s)
	if err != nil || !a.Is4() {
		return netip.Addr{}, fail(field, s, "must be an IPv4 address such as 10.77.0.1")
	}
	return a, nil
}

// Prefix4 validates an IPv4 CIDR. Host bits must be zero: "10.77.0.5/24" is
// rejected rather than silently widened, because the operator meant something.
func Prefix4(field, s string) (netip.Prefix, error) {
	p, err := netip.ParsePrefix(s)
	if err != nil || !p.Addr().Is4() {
		return netip.Prefix{}, fail(field, s, "must be an IPv4 CIDR such as 10.77.0.0/16")
	}
	if p.Masked() != p {
		return netip.Prefix{}, fail(field, s, "has host bits set; did you mean "+p.Masked().String()+"?")
	}
	return p, nil
}

// Port validates a TCP/UDP port number.
func Port(field string, n int) (int, error) {
	if n < 1 || n > 65535 {
		return 0, fail(field, strconv.Itoa(n), "must be a port between 1 and 65535")
	}
	return n, nil
}

// IntRange validates n within [lo, hi].
func IntRange(field string, n, lo, hi int) (int, error) {
	if n < lo || n > hi {
		return 0, fail(field, strconv.Itoa(n), fmt.Sprintf("must be between %d and %d", lo, hi))
	}
	return n, nil
}

// Endpoint validates a WireGuard endpoint: IPv4:port. Hostnames are not
// accepted in v1 -- every server has a static public IP, and resolving names
// would make the computed config depend on DNS at apply time.
func Endpoint(field, s string) (netip.AddrPort, error) {
	ap, err := netip.ParseAddrPort(s)
	if err != nil || !ap.Addr().Is4() || ap.Port() == 0 {
		return netip.AddrPort{}, fail(field, s, "must be IPv4:port such as 203.0.113.10:51820")
	}
	return ap, nil
}

// Key32 validates a base64 (standard alphabet, padded) 32-byte public key, the
// encoding WireGuard and this project use for both WireGuard and Ed25519 keys.
func Key32(field, s string) (string, error) {
	if len(s) != 44 {
		return "", fail(field, mask(s), "must be a base64-encoded 32-byte key (44 chars)")
	}
	b, err := base64.StdEncoding.Strict().DecodeString(s)
	if err != nil || len(b) != 32 {
		return "", fail(field, mask(s), "must be a base64-encoded 32-byte key (44 chars)")
	}
	return s, nil
}

// Token validates a join token: 32 random bytes, base64url without padding.
// The value is masked in the error.
func Token(field, s string) (string, error) {
	if !reToken.MatchString(s) {
		return "", fail(field, mask(s), "must be the 43-character join token shown by `meshctl node add`")
	}
	return s, nil
}

// LocalToken validates a local UI sign-in token or session cookie (amendment
// A7): 32 random bytes, base64url without padding, like a join token. The
// value is masked in the error.
func LocalToken(field, s string) (string, error) {
	if !reToken.MatchString(s) {
		return "", fail(field, mask(s), "must be the 43-character token of a link `meshctl ui` printed")
	}
	return s, nil
}

// PipePath validates the Windows named pipe of the agent's local API:
// \\.\pipe\ and a name of letters, digits, dots, dashes and underscores.
func PipePath(field, s string) (string, error) {
	if !rePipe.MatchString(s) {
		return "", fail(field, s, `must be a named pipe such as \\.\pipe\meshctl-agent`)
	}
	return s, nil
}

// Version validates a version string an agent reports or a build stamps:
// a tag (v1.2.3), git describe output (v1.2.3-4-gabcdef0-dirty), a bare
// commit hash, or "dev".
func Version(field, s string) (string, error) {
	if !reVersion.MatchString(s) {
		return "", fail(field, s, "must be 1-64 chars of A-Z, a-z, 0-9, '.', '_', '+', '-'")
	}
	return s, nil
}

// SemVer validates a release version (v1.2.3 or 1.2.3, optional pre-release).
func SemVer(field, s string) (string, error) {
	if !reSemVer.MatchString(s) {
		return "", fail(field, s, "must be a release version such as v1.2.3")
	}
	return s, nil
}

// ServerURL validates the control-plane origin agents and install.sh talk to.
// https only; plain http is accepted for loopback hosts (local dev, e2e tests).
// The result is the origin with no path, query, fragment or credentials.
func ServerURL(field, s string) (string, error) {
	u, err := url.Parse(s)
	if err != nil || u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" ||
		(u.Path != "" && u.Path != "/") || u.Opaque != "" {
		return "", fail(field, s, "must be an origin such as https://mesh.example.com (no path, query or credentials)")
	}
	host := u.Hostname()
	if _, err := Hostname(field, host); err != nil {
		if a, perr := netip.ParseAddr(host); perr != nil || !a.Is4() {
			return "", fail(field, s, "host must be a DNS name or an IPv4 address")
		}
	}
	if p := u.Port(); p != "" {
		n, err := strconv.Atoi(p)
		if err != nil || n < 1 || n > 65535 {
			return "", fail(field, s, "port must be between 1 and 65535")
		}
	}
	switch u.Scheme {
	case "https":
	case "http":
		if !isLoopbackHost(host) {
			return "", fail(field, s, "must use https (plain http is only allowed for localhost)")
		}
	default:
		return "", fail(field, s, "must use https")
	}
	return u.Scheme + "://" + strings.ToLower(u.Host), nil
}

func isLoopbackHost(h string) bool {
	if h == "localhost" {
		return true
	}
	a, err := netip.ParseAddr(h)
	return err == nil && a.IsLoopback()
}

// POSIXPath validates an absolute POSIX path that is written into an artifact
// as-is (a compose bind-mount source). Only [A-Za-z0-9._/-] is allowed, and no
// "." or ".." segments.
func POSIXPath(field, s string) (string, error) {
	if len(s) > 255 || !rePosix.MatchString(s) {
		return "", fail(field, s, "must be an absolute path of A-Z, a-z, 0-9, '.', '_', '/', '-'")
	}
	for _, seg := range strings.Split(s, "/") {
		if seg == "." || seg == ".." {
			return "", fail(field, s, "must not contain '.' or '..' segments")
		}
	}
	if c := path.Clean(s); c != s && c+"/" != s {
		return "", fail(field, s, "must be a clean path; did you mean "+c+"?")
	}
	return s, nil
}

// FilePath validates a local file-system path given on the command line (the
// data dir, an export file). It is never written into an artifact or a shell,
// but it does reach a SQLite DSN, so the URI-significant characters ?, # and %
// and every control character are refused.
func FilePath(field, s string) (string, error) {
	if s == "" || len(s) > 1024 {
		return "", fail(field, s, "must be a non-empty path")
	}
	if strings.ContainsAny(s, "?#%") {
		return "", fail(field, s, "must not contain '?', '#' or '%'")
	}
	for _, r := range s {
		if unicode.IsControl(r) {
			return "", fail(field, strconv.Quote(s), "must not contain control characters")
		}
	}
	return s, nil
}

// Text validates free text that is stored and displayed but never reaches a
// shell: an audit summary, a firewall rule note. It must be valid UTF-8 without
// control characters (so it cannot add a line to a log or a YAML key), at most
// max bytes.
func Text(field, s string, max int) (string, error) {
	if len(s) > max {
		return "", fail(field, s, fmt.Sprintf("must be at most %d bytes", max))
	}
	if !utf8.ValidString(s) {
		return "", fail(field, strconv.Quote(s), "must be valid UTF-8")
	}
	for _, r := range s {
		if unicode.IsControl(r) {
			return "", fail(field, strconv.Quote(s), "must not contain control characters (newline, tab, ...)")
		}
	}
	return s, nil
}

// Arg is the shell-layer check the exec runner applies to every argv element
// right before a command runs. Commands are never run through a shell, so this
// is defence in depth: an element may only contain [A-Za-z0-9 ._/:=,"-], which
// is enough for firewall-cmd options and rich rules
// (rule family="ipv4" source address="10.77.0.2/32" port port="2377" protocol="tcp" accept)
// and excludes every shell metacharacter, quote other than '"', and control
// character.
func Arg(s string) error {
	if s == "" || len(s) > 512 || !reArgChars.MatchString(s) {
		return fail("argv", strconv.Quote(s), `must be 1-512 chars of A-Z, a-z, 0-9, space and ._/:=,"-`)
	}
	return nil
}
