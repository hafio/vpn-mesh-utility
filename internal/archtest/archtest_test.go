package archtest

import (
	"os"
	"strings"
	"testing"
)

// rules is the boundary policy of this repository. CLAUDE.md names this test;
// change the two together.
var rules = []Rule{
	{Dir: "internal/server", Deny: []string{"./internal/agent"},
		Why: "server and agent talk only through internal/proto"},
	{Dir: "internal/agent", Deny: []string{"./internal/server"},
		Why: "server and agent talk only through internal/proto"},
	{Dir: "internal/proto", Deny: []string{"./internal/server", "./internal/agent"},
		Why: "the wire contract must not depend on either side"},
	{Dir: "internal/validate", Deny: []string{"./internal", "./cmd"},
		Why: "the validation gate is stdlib-only so every package can use it"},
	{Dir: "internal/fsutil", Deny: []string{"./internal", "./cmd"},
		Why: "the shared file helpers are stdlib-only so both sides can use them"},
	{Dir: "internal/testfix", Deny: []string{"./internal/server", "./internal/agent"},
		Why: "shared fixtures must be importable from both sides"},
	{Dir: "internal", Deny: []string{"./internal/testfix", "./internal/agent/agentfake"}, ProdOnly: true,
		Why: "fixtures and fakes are for tests only"},
	{Dir: "cmd", Deny: []string{"./internal/testfix", "./internal/agent/agentfake"}, ProdOnly: true,
		Why: "fixtures and fakes are for tests only"},
	{Dir: "deploy", Deny: []string{"./internal/server", "./internal/agent", "./cmd"},
		Why: "deploy artifacts are shared by the server (install.sh) and the agent (the unit)"},
	// The single exec site (CLAUDE.md security invariants).
	{Dir: "internal", Deny: []string{"os/exec"}, Except: []string{"internal/agent/execrun", "internal/tools"}, ProdOnly: true,
		Why: "programs run only through internal/agent/execrun, which re-validates every argv element"},
	{Dir: "cmd", Deny: []string{"os/exec"}, ProdOnly: true,
		Why: "programs run only through internal/agent/execrun, which re-validates every argv element"},
	{Dir: "internal/server", Deny: []string{
		"golang.zx2c4.com/wireguard", // includes wgctrl
		"github.com/vishvananda/netlink",
		"github.com/miekg/dns",
		"github.com/godbus/dbus",
		"golang.org/x/sys/windows",
		"github.com/Microsoft/go-winio",
	}, Why: "the server only computes config; data-plane libraries belong to the agent"},
	{Dir: "internal/agent", Deny: []string{
		"modernc.org/sqlite",
		"golang.org/x/crypto/argon2",
		"github.com/pquerna/otp",
		"github.com/boombuler/barcode",
		"github.com/skip2/go-qrcode",
	}, Why: "storage and UI auth belong to the server"},
}

func TestModuleBoundaries(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	root, module, err := ModuleRoot(wd)
	if err != nil {
		t.Fatal(err)
	}
	if module != "github.com/hafio/vpn-mesh-utility" {
		t.Fatalf("module = %q; the rules below assume github.com/hafio/vpn-mesh-utility", module)
	}
	vs, err := Check(root, module, rules)
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range vs {
		t.Errorf("boundary violation: %s", v)
	}
}

// TestCheckerCatchesViolations is the negative control: a checker that
// silently matched nothing would pass TestModuleBoundaries forever.
func TestCheckerCatchesViolations(t *testing.T) {
	vs, err := Check("testdata/violation", "example.com/fake", rules)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		"example.com/fake/internal/agent imports example.com/fake/internal/testfix (GOOS=linux)",
		"example.com/fake/internal/agent imports os/exec (GOOS=linux)",
		"example.com/fake/internal/server imports example.com/fake/internal/agent (GOOS=linux)",
		"example.com/fake/internal/server imports golang.zx2c4.com/wireguard/wgctrl (GOOS=linux)",
	}
	got := make([]string, 0, len(vs))
	for _, v := range vs {
		got = append(got, v.String())
	}
	joined := strings.Join(got, "\n")
	for _, w := range want {
		if !strings.Contains(joined, w) {
			t.Errorf("checker missed %q; got:\n%s", w, joined)
		}
	}
	// The sanctioned exec site is exempt.
	if strings.Contains(joined, "fake/internal/agent/execrun imports os/exec") {
		t.Errorf("the Except list was ignored:\n%s", joined)
	}
	// A test-only import of the fixtures is allowed (ProdOnly rule).
	if strings.Contains(joined, "fake/internal/proto imports example.com/fake/internal/testfix") {
		t.Errorf("a test file importing testfix must not count as a production import:\n%s", joined)
	}
}
